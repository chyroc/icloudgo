package command

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/chyroc/icloudgo"
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
			Name:     "recent",
			Usage:    "download recent photos, if not set, means all",
			Required: false,
			Aliases:  []string{"r"},
			EnvVars:  []string{"ICLOUD_RECENT"},
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
			Value:    50,
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

	if err := cmd.downloadPhoto(cmd.Recent, cmd.Offset); err != nil {
		return err
	}

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
	Recent     int
	Offset     int
	StopNum    int
	Album      string
	ThreadNum  int
	AutoDelete bool

	client   *icloudgo.Client
	photoCli *icloudgo.PhotoService
	db       *gorm.DB
}

func newDownloadCommand(c *cli.Context) (*downloadCommand, error) {
	cmd := &downloadCommand{
		Username:   c.String("username"),
		Password:   c.String("password"),
		CookieDir:  c.String("cookie-dir"),
		Domain:     c.String("domain"),
		Output:     c.String("output"),
		Offset:     c.Int("offset"),
		Recent:     c.Int("recent"),
		StopNum:    c.Int("stop-found-num"),
		Album:      c.String("album"),
		ThreadNum:  c.Int("thread-num"),
		AutoDelete: c.Bool("auto-delete"),
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

	db, err := gorm.Open(sqlite.Open(cli.ConfigPath("download.db")), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	cmd.client = cli
	cmd.photoCli = photoCli
	cmd.db = db
	if cmd.Offset == -1 {
		cmd.Offset = getDownloadOffset(cli)
	}

	return cmd, nil
}

func (r *downloadCommand) downloadPhoto(recent, downloadOffset int) error {
	if f, _ := os.Stat(r.Output); f == nil {
		if err := os.MkdirAll(r.Output, os.ModePerm); err != nil {
			return err
		}
	}

	album, err := r.photoCli.GetAlbum(r.Album)
	if err != nil {
		return err
	}

	defer r.saveDownloadOffset(r.Offset)
	fmt.Printf("album: %s, total: %d, offset: %d, target: %s, thread-num: %d\n", album.Name, album.Size(), downloadOffset, r.Output, r.ThreadNum)

	if recent == 0 {
		recent, err = album.GetSize()
		if err != nil {
			return err
		}
	}

	photoIter := album.PhotosIter(downloadOffset)
	wait := new(sync.WaitGroup)
	foundDownloadedNum := int32(0)
	var downloaded int32
	var finalErr error
	for threadIndex := 0; threadIndex < r.ThreadNum; threadIndex++ {
		wait.Add(1)
		go func(threadIndex int) {
			defer wait.Done()

			for {
				if recent > 0 && atomic.LoadInt32(&downloaded) >= int32(recent) {
					return
				}
				if r.StopNum > 0 && atomic.LoadInt32(&foundDownloadedNum) >= int32(r.StopNum) {
					return
				}

				photoAsset, err := photoIter.Next()
				if err != nil {
					if errors.Is(err, icloudgo.ErrPhotosIterateEnd) {
						return
					}
					if finalErr != nil {
						finalErr = err
					}
					return
				}

				if offset := photoIter.Offset(); offset != downloadOffset {
					r.saveDownloadOffset(offset)
					downloadOffset = offset
				}

				if isDownloaded, err := r.downloadPhotoAsset(photoAsset, threadIndex); err != nil {
					if finalErr != nil {
						finalErr = err
					}
					return
				} else if isDownloaded {
					atomic.AddInt32(&foundDownloadedNum, 1)
					if r.StopNum > 0 && foundDownloadedNum >= int32(r.StopNum) {
						return
					}
				} else {
					atomic.AddInt32(&downloaded, 1)
				}
			}
		}(threadIndex)
	}
	wait.Wait()

	return finalErr
}

func (r *downloadCommand) downloadPhotoAsset(photo *icloudgo.PhotoAsset, threadIndex int) (bool, error) {
	filename := photo.Filename()
	path := photo.LocalPath(r.Output, icloudgo.PhotoVersionOriginal)
	fmt.Printf("start %v, %v, %v, thread=%d\n", photo.ID(), filename, photo.FormatSize(), threadIndex)

	if f, _ := os.Stat(path); f != nil {
		if photo.Size() != int(f.Size()) {
			return false, photo.DownloadTo(icloudgo.PhotoVersionOriginal, path)
		} else {
			fmt.Printf("file '%s' exist, skip.\n", path)
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

const configDownloadOffset = "download_offset.txt"

func getDownloadOffset(cli *icloudgo.Client) int {
	content, err := cli.LoadConfig(configDownloadOffset)
	if err != nil {
		fmt.Printf("load download_offset config failed: %s, reset to 0\n", err)
		return 0
	} else if len(content) == 0 {
		return 0
	}

	i, err := strconv.Atoi(string(content))
	if err != nil {
		fmt.Printf("load download_offset config, strconv failed: %s, reset to 0\n", err)
		_ = cli.SaveConfig("download_offset", []byte("0"))
		return 0
	}

	if i < 0 {
		fmt.Printf("load download_offset config, invalid data: %d, reset to 0\n", i)
		_ = cli.SaveConfig("download_offset", []byte("0"))
		return 0
	}

	return i
}

func (r *downloadCommand) saveDownloadOffset(i int) {
	err := r.client.SaveConfig(configDownloadOffset, []byte(strconv.Itoa(i)))
	if err != nil {
		fmt.Printf("save download_offset config failed: %s", err)
	}
}
