package command

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chyroc/icloudgo"
	"github.com/chyroc/icloudgo/internal"
	"github.com/glebarez/sqlite"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"
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
		&cli.IntFlag{
			Name:     "offset",
			Usage:    "download offset, if not set, means 0, or re-stored from cookie dir",
			Required: false,
			Value:    -1,
			EnvVars:  []string{"ICLOUD_OFFSET"},
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
			Usage:    "auto delete photos after download",
			Required: false,
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

	go cmd.savePhotoMeta(cmd.Offset)
	cmd.download()

	if cmd.AutoDelete {
		return cmd.autoDeletePhoto()
	}

	return nil
}

type downloadCommand struct {
	Username   string
	Password   string
	CookieDir  string
	Domain     string
	Output     string
	Offset     int
	StopNum    int
	AlbumName  string
	ThreadNum  int
	AutoDelete bool

	client   *icloudgo.Client
	photoCli *icloudgo.PhotoService
	db       *gorm.DB
	lock     *sync.Mutex
	exit     chan struct{}
}

func newDownloadCommand(c *cli.Context) (*downloadCommand, error) {
	cmd := &downloadCommand{
		Username:   c.String("username"),
		Password:   c.String("password"),
		CookieDir:  c.String("cookie-dir"),
		Domain:     c.String("domain"),
		Output:     c.String("output"),
		Offset:     c.Int("offset"),
		StopNum:    c.Int("stop-found-num"),
		AlbumName:  c.String("album"),
		ThreadNum:  c.Int("thread-num"),
		AutoDelete: c.Bool("auto-delete"),
		lock:       &sync.Mutex{},
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

	dbPath := cli.ConfigPath("download.db")
	db, err := gorm.Open(sqlite.Open(dbPath), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.Migrator().AutoMigrate(&PhotoAssetModel{}); err != nil {
		return nil, err
	}
	if err := db.Migrator().AutoMigrate(&DownloadOffsetModel{}); err != nil {
		return nil, err
	}

	cmd.client = cli
	cmd.photoCli = photoCli
	cmd.db = db
	if cmd.Offset == -1 {
		cmd.Offset = cmd.getDownloadOffset()
	}

	return cmd, nil
}

func (r *downloadCommand) savePhotoMeta(downloadOffset int) error {
	album, err := r.photoCli.GetAlbum(r.AlbumName)
	if err != nil {
		return err
	}

	fmt.Printf("[icloudgo] [meta] album: %s, total: %d, offset: %d, target: %s, thread-num: %d, stop-num: %d\n", album.Name, album.Size(), downloadOffset, r.Output, r.ThreadNum, r.StopNum)
	fmt.Printf("[icloudgo] [meta] start download photo meta\n")
	err = album.WalkPhotos(downloadOffset, func(offset int, assets []*internal.PhotoAsset) error {
		if err := r.insertAssets(assets); err != nil {
			return err
		}
		if err := r.saveDownloadOffset(offset); err != nil {
			return err
		}
		fmt.Printf("[icloudgo] [meta] update download offst to %d\n", offset)
		return nil
	})
	if err != nil {
		fmt.Printf("[icloudgo] [meta] walk photos err: %s\n", err)
	}
	return nil
}

func (r *downloadCommand) download() error {
	if f, _ := os.Stat(r.Output); f == nil {
		if err := os.MkdirAll(r.Output, os.ModePerm); err != nil {
			return err
		}
	}

	if err := r.downloadFromDatabase(); err != nil {
		fmt.Printf("[icloudgo] [download] download err: %s", err)
	}

	ticker := time.NewTicker(time.Minute)
	for {
		select {
		case <-ticker.C:
			if err := r.downloadFromDatabase(); err != nil {
				fmt.Printf("[icloudgo] [download] download err: %s", err)
			}
		}
	}
}

func (r *downloadCommand) downloadFromDatabase() error {
	assets, err := r.getUnDownloadAssets()
	if err != nil {
		return fmt.Errorf("get undownload assets err: %w", err)
	} else if len(assets) == 0 {
		fmt.Printf("[icloudgo] [download] no undownload assets\n")
		return nil
	}
	fmt.Printf("[icloudgo] [download] found %d undownload assets\n", len(assets))

	assetPOChan := make(chan *PhotoAssetModel, len(assets))
	for _, asset := range assets {
		assetPOChan <- asset
	}

	wait := new(sync.WaitGroup)
	foundDownloadedNum := int32(0)
	var downloaded int32
	var errCount int32
	for threadIndex := 0; threadIndex < r.ThreadNum; threadIndex++ {
		wait.Add(1)
		go func(threadIndex int) {
			defer wait.Done()
			for {
				if atomic.LoadInt32(&errCount) > 20 {
					fmt.Printf("[icloudgo] [download] too many errors, stop download\n")
					os.Exit(1)
					return
				}

				if r.StopNum > 0 && atomic.LoadInt32(&foundDownloadedNum) >= int32(r.StopNum) {
					return
				}

				assetPO := <-assetPOChan
				photoAsset := r.photoCli.NewPhotoAssetFromBytes([]byte(assetPO.Data))

				if isDownloaded, err := r.downloadPhotoAsset(photoAsset, threadIndex); err != nil {
					atomic.AddInt32(&errCount, 1)
					continue
				} else if isDownloaded {
					_ = r.setDownloaded(photoAsset.ID())
					atomic.AddInt32(&foundDownloadedNum, 1)
					if r.StopNum > 0 && foundDownloadedNum >= int32(r.StopNum) {
						return
					}
				} else {
					_ = r.setDownloaded(assetPO.ID)
					atomic.AddInt32(&downloaded, 1)
				}
			}
		}(threadIndex)
	}
	wait.Wait()
	return nil
}

func (r *downloadCommand) downloadPhotoAsset(photo *icloudgo.PhotoAsset, threadIndex int) (bool, error) {
	filename := photo.Filename()
	path := photo.LocalPath(r.Output, icloudgo.PhotoVersionOriginal)
	fmt.Printf("[icloudgo] [download] %v, %v, %v, thread=%d\n", photo.ID(), filename, photo.FormatSize(), threadIndex)

	if f, _ := os.Stat(path); f != nil {
		if photo.Size() != int(f.Size()) {
			return false, photo.DownloadTo(icloudgo.PhotoVersionOriginal, path)
		} else {
			fmt.Printf("[icloudgo] [download] '%s' exist, skip.\n", path)
			return true, nil
		}
	} else {
		return false, photo.DownloadTo(icloudgo.PhotoVersionOriginal, path)
	}
}

func (r *downloadCommand) autoDeletePhoto() error {
	album, err := r.photoCli.GetAlbum(icloudgo.AlbumNameRecentlyDeleted)
	if err != nil {
		return err
	}

	fmt.Printf("auto delete album: %s, total: %d\n", album.Name, album.Size())

	photoIter := album.PhotosIter(0)
	wait := new(sync.WaitGroup)
	var finalErr error
	for threadIndex := 0; threadIndex < r.ThreadNum; threadIndex++ {
		wait.Add(1)
		go func(threadIndex int) {
			defer wait.Done()

			for {
				photoAsset, err := photoIter.Next()
				if err != nil {
					if errors.Is(err, icloudgo.ErrPhotosIterateEnd) {
						return
					}
					if finalErr == nil {
						finalErr = err
					}
					return
				}

				path := photoAsset.LocalPath(r.Output, icloudgo.PhotoVersionOriginal)

				if err := os.Remove(path); err != nil {
					if errors.Is(err, os.ErrNotExist) {
						continue
					}
					if finalErr != nil {
						finalErr = err
					}
					return
				} else {
					fmt.Printf("delete %v, %v, %v, thread=%d\n", photoAsset.ID(), photoAsset.Filename(), photoAsset.FormatSize(), threadIndex)
				}
			}
		}(threadIndex)
	}
	wait.Wait()

	return finalErr
}
