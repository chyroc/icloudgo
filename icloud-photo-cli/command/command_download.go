package command

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/glebarez/sqlite"
	"github.com/urfave/cli/v2"
	"gorm.io/gorm"

	"github.com/chyroc/icloudgo"
	"github.com/chyroc/icloudgo/internal"
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

	client   *icloudgo.Client
	photoCli *icloudgo.PhotoService
	db       *gorm.DB
	lock     *sync.Mutex
	exit     chan struct{}
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
		lock:            &sync.Mutex{},
		exit:            make(chan struct{}),
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

	return cmd, nil
}

func (r *downloadCommand) saveMeta() error {
	album, err := r.photoCli.GetAlbum(r.AlbumName)
	if err != nil {
		return err
	}

	for {
		downloadOffset := r.getDownloadOffset()
		fmt.Printf("[icloudgo] [meta] album: %s, total: %d, offset: %d, target: %s, thread-num: %d, stop-num: %d\n", album.Name, album.Size(), downloadOffset, r.Output, r.ThreadNum, r.StopNum)
		err = album.WalkPhotos(downloadOffset, func(offset int, assets []*internal.PhotoAsset) error {
			if err := r.dalAddAssets(assets); err != nil {
				return err
			}
			if err := r.saveDownloadOffset(offset, true); err != nil {
				return err
			}
			fmt.Printf("[icloudgo] [meta] update download offst to %d\n", offset)
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

func (r *downloadCommand) download() error {
	if err := mkdirAll(r.Output); err != nil {
		return err
	}
	if err := mkdirAll(filepath.Join(r.Output, ".tmp")); err != nil {
		return err
	}

	for {
		if err := r.downloadFromDatabase(); err != nil {
			fmt.Printf("[icloudgo] [download] download err: %s", err)
			time.Sleep(time.Minute)
		} else {
			time.Sleep(time.Hour)
		}
	}
}

func (r *downloadCommand) downloadFromDatabase() error {
	assets, err := r.dalGetUnDownloadAssets()
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
	var finalErr error
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

				assetPO := <-assetPOChan
				photoAsset := r.photoCli.NewPhotoAssetFromBytes([]byte(assetPO.Data))

				if isDownloaded, err := r.downloadPhotoAsset(photoAsset, threadIndex); err != nil {
					if errors.Is(err, internal.ErrResourceGone) {
						continue
					}
					atomic.AddInt32(&errCount, 1)
					finalErr = err
					continue
				} else if isDownloaded {
					_ = r.dalSetDownloaded(photoAsset.ID())
					atomic.AddInt32(&foundDownloadedNum, 1)
					if r.StopNum > 0 && foundDownloadedNum >= int32(r.StopNum) {
						return
					}
				} else {
					_ = r.dalSetDownloaded(assetPO.ID)
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
	outputDir := photo.OutputDir(r.Output, r.FolderStructure)
	tmpPath := photo.LocalPath(filepath.Join(r.Output, ".tmp"), icloudgo.PhotoVersionOriginal)
	path := photo.LocalPath(outputDir, icloudgo.PhotoVersionOriginal)
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		fmt.Printf("[icloudgo] [download] mkdir '%s' output dir: '%s' failed: %s\n", photo.Filename(), outputDir, err)
		return false, err
	}
	fmt.Printf("[icloudgo] [download] %v, %v, %v, thread=%d\n", photo.ID(), filename, photo.FormatSize(), threadIndex)

	if f, _ := os.Stat(path); f != nil {
		if photo.Size() != int(f.Size()) {
			return false, r.downloadTo(photo, tmpPath, path)
		} else {
			fmt.Printf("[icloudgo] [download] '%s' exist, skip.\n", path)
			return true, nil
		}
	} else {
		return false, r.downloadTo(photo, tmpPath, path)
	}
}

func (r *downloadCommand) downloadTo(photo *icloudgo.PhotoAsset, tmpPath, realPath string) error {
	if err := photo.DownloadTo(icloudgo.PhotoVersionOriginal, tmpPath); err != nil {
		return err
	}

	if err := os.Rename(tmpPath, realPath); err != nil {
		return fmt.Errorf("rename '%s' to '%s' failed: %w", tmpPath, realPath, err)
	}

	return nil
}

func (r *downloadCommand) autoDeletePhoto() error {
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
				path := photoAsset.LocalPath(photoAsset.OutputDir(r.Output, r.FolderStructure), icloudgo.PhotoVersionOriginal)
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
