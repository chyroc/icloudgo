package command

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chyroc/icloudgo"
	"github.com/chyroc/icloudgo/internal"
	"github.com/dgraph-io/badger/v3"
	"github.com/urfave/cli/v2"
)

func NewDownloadFlag() []cli.Flag {
	var res []cli.Flag
	res = append(res, commonFlag...)
	res = append(res,
		&cli.StringFlag{
			Name:     "output",
			Usage:    "output dir",
			Required: false,
			Value:    "./iCloudPhotos",
			Aliases:  []string{"o"},
			EnvVars:  []string{"ICLOUD_OUTPUT"},
		},
		&cli.StringFlag{
			Name:     "album",
			Usage:    "album name, if not set, download all albums",
			Required: false,
			Aliases:  []string{"a"},
			EnvVars:  []string{"ICLOUD_ALBUM"},
		},
		&cli.StringFlag{
			Name:     "folder-structure",
			Usage:    "support: `2006`(year), `01`(month), `02`(day), `15`(24-hour), `03`(12-hour), `04`(minute), `05`(second), example: `2006/01/02`, default is `/`",
			Required: false,
			Value:    "/",
			Aliases:  []string{"fs"},
			EnvVars:  []string{"ICLOUD_FOLDER_STRUCTURE"},
		},
		&cli.StringFlag{
			Name:     "file-structure",
			Usage:    "support: id(unique file id), name(file human readable name)",
			Required: false,
			Value:    "id",
			EnvVars:  []string{"ICLOUD_FILE_STRUCTURE"},
		},
		&cli.IntFlag{
			Name:     "stop-found-num",
			Usage:    "stop download when found `stop-found-num` photos have been downloaded",
			Required: false,
			Value:    0,
			Aliases:  []string{"s"},
			EnvVars:  []string{"ICLOUD_STOP_FOUND_NUM"},
		},
		&cli.IntFlag{
			Name:     "thread-num",
			Usage:    "thread num, if not set, means 1",
			Required: false,
			Aliases:  []string{"t"},
			Value:    1,
			EnvVars:  []string{"ICLOUD_THREAD_NUM"},
		},
		&cli.BoolFlag{
			Name:     "auto-delete",
			Usage:    "Automatically delete photos from local but recently deleted folders",
			Required: false,
			Value:    true,
			Aliases:  []string{"ad"},
			EnvVars:  []string{"ICLOUD_AUTO_DELETE"},
		},
	)
	return res
}

func Download(c *cli.Context) error {
	cmd, err := newDownloadCommand(c)
	if err != nil {
		return err
	}
	defer cmd.client.Close()

	go cmd.saveMeta()
	go cmd.download()
	go cmd.autoDeletePhoto()

	// hold
	<-cmd.exit

	cmd.Close()

	return nil
}

type downloadCommand struct {
	Username        string
	Password        string
	CookieDir       string
	Domain          string
	Output          string
	StopNum         int
	AlbumName       string
	ThreadNum       int
	AutoDelete      bool
	FolderStructure string
	FileStructure   string

	client        *icloudgo.Client
	photoCli      *icloudgo.PhotoService
	db            *badger.DB
	lock          *sync.Mutex
	exit          chan struct{}
	startDownload chan struct{}
}

func newDownloadCommand(c *cli.Context) (*downloadCommand, error) {
	cmd := &downloadCommand{
		Username:        c.String("username"),
		Password:        c.String("password"),
		CookieDir:       c.String("cookie-dir"),
		Domain:          c.String("domain"),
		Output:          c.String("output"),
		StopNum:         c.Int("stop-found-num"),
		AlbumName:       c.String("album"),
		ThreadNum:       c.Int("thread-num"),
		AutoDelete:      c.Bool("auto-delete"),
		FolderStructure: c.String("folder-structure"),
		FileStructure:   c.String("file-structure"),
		lock:            &sync.Mutex{},
		exit:            make(chan struct{}),
		startDownload:   make(chan struct{}),
	}
	if cmd.AlbumName == "" {
		cmd.AlbumName = icloudgo.AlbumNameAll
	}

	cli, err := icloudgo.New(&icloudgo.ClientOption{
		AppID:           cmd.Username,
		CookieDir:       cmd.CookieDir,
		PasswordGetter:  getTextInput("apple id password", cmd.Password),
		TwoFACodeGetter: getTextInput("2fa code", ""),
		Domain:          cmd.Domain,
	})
	if err != nil {
		return nil, err
	}
	if err := cli.Authenticate(false, nil); err != nil {
		return nil, err
	}
	photoCli, err := cli.PhotoCli()
	if err != nil {
		return nil, err
	}

	dbPath := cli.ConfigPath("badger.db")
	db, err := badger.Open(badger.DefaultOptions(dbPath))
	if err != nil {
		return nil, err
	}

	cmd.client = cli
	cmd.photoCli = photoCli
	cmd.db = db

	return cmd, nil
}

func (r *downloadCommand) saveMeta() (err error) {
	defer func() {
		if err != nil {
			fmt.Printf("[icloudgo] [meta] final err:%s\n", err.Error())
		}
	}()
	album, err := r.photoCli.GetAlbum(r.AlbumName)
	if err != nil {
		return err
	}

	for {
		dbOffset := r.dalGetDownloadOffset(album.Size())
		fmt.Printf("[icloudgo] [meta] album: %s, total: %d, db_offset: %d, target: %s, thread-num: %d, stop-num: %d\n", album.Name, album.Size(), dbOffset, r.Output, r.ThreadNum, r.StopNum)
		err = album.WalkPhotos(dbOffset, func(offset int, assets []*internal.PhotoAsset) error {
			if err := r.dalAddAssets(assets); err != nil {
				return err
			}
			if err := r.saveDownloadOffset(nil, offset, true); err != nil {
				return err
			}
			fmt.Printf("[icloudgo] [meta] update download offst to %d\n", offset)
			r.setStartDownload()
			return nil
		})
		if err != nil {
			fmt.Printf("[icloudgo] [meta] walk photos err: %s\n", err)
			time.Sleep(time.Minute)
		} else {
			time.Sleep(time.Hour)
		}
	}
}

func (r *downloadCommand) setStartDownload() {
	select {
	case r.startDownload <- struct{}{}:
		return
	case <-time.After(time.Second / 10):
		return
	}
}

func (r *downloadCommand) download() (err error) {
	defer func() {
		if err != nil {
			fmt.Printf("[icloudgo] [download] final err:%s\n", err.Error())
		}
	}()
	if err := mkdirAll(r.Output); err != nil {
		return err
	}
	if err := mkdirAll(filepath.Join(r.Output, ".tmp")); err != nil {
		return err
	}

	fmt.Printf("[icloudgo] [download] start\n")
	short := time.Minute
	long := time.Hour
	timer := time.NewTimer(time.Second / 10) // 立刻开始
	download := func() {
		fmt.Printf("[icloudgo] [download] start run %s\n", time.Now())
		if err := r.downloadFromDatabase(); err != nil {
			fmt.Printf("[icloudgo] [download] download err: %s, sleep %s", err, short)
			timer.Reset(short)
		} else {
			fmt.Printf("[icloudgo] [download] download success, sleep %s", long)
			timer.Reset(long)
		}
	}
	for {
		select {
		case <-r.startDownload:
			download()
		case <-timer.C:
			download()
		}
	}
}

func (r *downloadCommand) downloadFromDatabase() error {
	assetQueue, err := r.getUnDownloadAssets()
	if err != nil {
		return fmt.Errorf("get undownload assets err: %w", err)
	} else if assetQueue.empty() {
		fmt.Printf("[icloudgo] [download] no undownload assets\n")
		return nil
	}
	fmt.Printf("[icloudgo] [download] found %d undownload assets\n", assetQueue.len())

	wait := new(sync.WaitGroup)
	foundDownloadedNum := int32(0)
	var downloaded int32
	var errCount int32
	var finalErr error
	addError := func(msg string, err error) {
		if err == nil {
			return
		}
		atomic.AddInt32(&errCount, 1)
		finalErr = err
		fmt.Printf("[icloudgo] [download] %s failed: %s\n", msg, err.Error())
	}
	for threadIndex := 0; threadIndex < r.ThreadNum; threadIndex++ {
		wait.Add(1)
		go func(threadIndex int) {
			defer wait.Done()
			for {
				if atomic.LoadInt32(&errCount) > 20 {
					fmt.Printf("[icloudgo] [download] too many errors, stop download, last error: %s\n", finalErr.Error())
					os.Exit(1)
					return
				}

				if r.StopNum > 0 && atomic.LoadInt32(&foundDownloadedNum) >= int32(r.StopNum) {
					return
				}

				photoAsset, pickReason := assetQueue.pick(float32(threadIndex) / float32(r.ThreadNum))
				if photoAsset == nil {
					return
				}

				if isDownloaded, err := r.downloadPhotoAsset(photoAsset, pickReason); err != nil {
					if errors.Is(err, internal.ErrResourceGone) || strings.Contains(err.Error(), "no such host") {
						// delete db
						if err := r.dalDeleteAsset(photoAsset.ID()); err != nil {
							fmt.Printf("[icloudgo] [download] remove gone resource failed: %s\n", err)
						}
						continue
					}
					addError("downloadPhotoAsset", err)
					continue
				} else if isDownloaded {
					if err = r.dalSetDownloaded(photoAsset.ID()); err != nil {
						addError("dalSetDownloaded[downloaded]", err)
						continue
					}
					atomic.AddInt32(&foundDownloadedNum, 1)
					if r.StopNum > 0 && foundDownloadedNum >= int32(r.StopNum) {
						return
					}
				} else {
					if err = r.dalSetDownloaded(photoAsset.ID()); err != nil {
						addError("dalSetDownloaded[download]", err)
						continue
					}
					atomic.AddInt32(&downloaded, 1)
				}
			}
		}(threadIndex)
	}
	wait.Wait()
	return nil
}

func (r *downloadCommand) downloadPhotoAsset(photo *icloudgo.PhotoAsset, pickReason string) (bool, error) {
	outputDir := photo.OutputDir(r.Output, r.FolderStructure)
	tmpPath := photo.LocalPath(filepath.Join(r.Output, ".tmp"), icloudgo.PhotoVersionOriginal, r.FileStructure)
	path := photo.LocalPath(outputDir, icloudgo.PhotoVersionOriginal, r.FileStructure)
	name := path[len(r.Output):]
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		fmt.Printf("[icloudgo] [download] [%s] mkdir '%s' output dir: '%s' failed: %s\n", pickReason, photo.Filename(), outputDir, err)
		return false, err
	}

	if f, _ := os.Stat(path); f != nil {
		if photo.Size() != int(f.Size()) {
			return false, r.downloadTo(pickReason, photo, tmpPath, path, name)
		} else {
			// fmt.Printf("[icloudgo] [download] '%s' exist, skip.\n", path)
			return true, nil
		}
	} else {
		return false, r.downloadTo(pickReason, photo, tmpPath, path, name)
	}
}

func (r *downloadCommand) downloadTo(pickReason string, photo *icloudgo.PhotoAsset, tmpPath, realPath, saveName string) (err error) {
	start := time.Now()
	fmt.Printf("[icloudgo] [download] [%s] start %v, %v, %v\n", pickReason, saveName, photo.Filename(), photo.FormatSize())
	defer func() {
		diff := time.Now().Sub(start)
		speed := float64(photo.Size()) / 1024 / diff.Seconds()
		if err != nil && !errors.Is(err, internal.ErrResourceGone) && !strings.Contains(err.Error(), "no such host") {
			fmt.Printf("[icloudgo] [download] fail %v, %v, %v/%v %.2fKB/s err=%s\n", saveName, photo.Filename(), photo.FormatSize(), diff, speed, err)
		} else {
			fmt.Printf("[icloudgo] [download] [%s] succ %v, %v, %v/%v %.2fKB/s\n", pickReason, saveName, photo.Filename(), photo.FormatSize(), diff, speed)
		}
	}()
	retry := 5
	for i := 0; i < retry; i++ {
		if err := photo.DownloadTo(icloudgo.PhotoVersionOriginal, tmpPath); err != nil {
			if strings.Contains(err.Error(), "i/o timeout") && i < retry-1 {
				continue
			}
			return err
		}
	}

	if err := os.Rename(tmpPath, realPath); err != nil {
		return fmt.Errorf("rename '%s' to '%s' failed: %w", tmpPath, realPath, err)
	}

	return nil
}

func (r *downloadCommand) autoDeletePhoto() (err error) {
	defer func() {
		if err != nil {
			fmt.Printf("[icloudgo] [auto_delete] final err:%s\n", err.Error())
		}
	}()
	if !r.AutoDelete {
		return nil
	}

	for {
		album, err := r.photoCli.GetAlbum(icloudgo.AlbumNameRecentlyDeleted)
		if err != nil {
			time.Sleep(time.Minute)
			continue
		}

		fmt.Printf("[icloudgo] [auto_delete] auto delete album total: %d\n", album.Size())
		if err = album.WalkPhotos(0, func(offset int, assets []*internal.PhotoAsset) error {
			for _, photoAsset := range assets {
				if err := r.dalDeleteAsset(photoAsset.ID()); err != nil {
					return err
				}
				path := photoAsset.LocalPath(photoAsset.OutputDir(r.Output, r.FolderStructure), icloudgo.PhotoVersionOriginal, r.FileStructure)
				if err := os.Remove(path); err != nil {
					if errors.Is(err, os.ErrNotExist) {
						continue
					}
					return err
				}
				fmt.Printf("[icloudgo] [auto_delete] delete %v, %v, %v\n", photoAsset.ID(), photoAsset.Filename(), photoAsset.FormatSize())
			}
			return nil
		}); err != nil {
			time.Sleep(time.Minute)
			continue
		}
		time.Sleep(time.Hour)
	}
}

func (r *downloadCommand) Close() {
	if r.db != nil {
		r.db.Close()
	}
}

func (r *downloadCommand) getUnDownloadAssets() (*assertQueue, error) {
	assets, err := r.dalGetUnDownloadAssets(&[]int{0}[0])
	if err != nil {
		return nil, err
	} else if len(assets) == 0 {
		return newAssertQueue(nil), nil
	}
	fmt.Printf("[icloudgo] [download] found %d undownload assets\n", len(assets))

	var photoAssetList []*icloudgo.PhotoAsset
	for _, po := range assets {
		photoAssetList = append(photoAssetList, r.photoCli.NewPhotoAssetFromBytes([]byte(po.Data)))
	}
	sort.SliceStable(photoAssetList, func(i, j int) bool {
		return photoAssetList[i].Size() < photoAssetList[j].Size()
	})

	return newAssertQueue(photoAssetList), nil
}

type assertQueue struct {
	recentAssets []*icloudgo.PhotoAsset
	recentIndex  int

	oldAssets []*icloudgo.PhotoAsset
	lowIndex  int
	highIndex int
	lock      *sync.Mutex
}

func newAssertQueue(data []*icloudgo.PhotoAsset) *assertQueue {
	// 2天前的时间
	now := time.Now()
	twoDaysAge := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location()).Add(-time.Hour * 24 * 2)
	// 区分热数据, 老数据
	recentAssets := []*icloudgo.PhotoAsset{}
	oldAssets := []*icloudgo.PhotoAsset{}
	for _, v := range data {
		if v.Created().Before(twoDaysAge) {
			oldAssets = append(oldAssets, v)
		} else {
			recentAssets = append(recentAssets, v)
		}
	}
	return &assertQueue{
		recentAssets: recentAssets,
		recentIndex:  -1,
		oldAssets:    oldAssets,
		lowIndex:     -1,
		highIndex:    len(oldAssets),
		lock:         new(sync.Mutex),
	}
}

func (r *assertQueue) pick(percent float32) (*icloudgo.PhotoAsset, string) {
	r.lock.Lock()
	defer r.lock.Unlock()

	// 30% 的概率从 [热数据] 中选取
	if percent <= 0.3 {
		r.recentIndex++
		if r.recentIndex < len(r.recentAssets) {
			return r.recentAssets[r.recentIndex], "recent"
		}
	}

	// 20% ~ 50% 的概率从 [小数据] 中选取
	if percent <= 0.5 {
		r.lowIndex++
		if r.lowIndex < r.highIndex {
			return r.oldAssets[r.lowIndex], "small"
		}
		return nil, ""
	}

	// 50% ~ 80% 的概率从 [大数据] 中选取
	r.highIndex--
	if r.highIndex > r.lowIndex {
		return r.oldAssets[r.highIndex], "big"
	}
	return nil, ""
}

func (r *assertQueue) empty() bool {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.highIndex-1 <= r.lowIndex && r.recentIndex >= len(r.recentAssets)-1
}

func (r *assertQueue) len() int {
	r.lock.Lock()
	defer r.lock.Unlock()
	return (r.highIndex - 1 - r.lowIndex) + 1 + (len(r.recentAssets) - 1 - r.recentIndex)
}
