package command

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/urfave/cli/v2"

	"github.com/chyroc/icloudgo"
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
		&cli.Int64Flag{
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
		&cli.Int64Flag{
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
	params := getDownloadParam(c)

	cli, err := icloudgo.New(&icloudgo.ClientOption{
		AppID:           params.Username,
		CookieDir:       params.CookieDir,
		PasswordGetter:  getTextInput("apple id password", params.Password),
		TwoFACodeGetter: getTextInput("2fa code", ""),
		Domain:          params.Domain,
	})
	if err != nil {
		return err
	}

	if params.Offset == -1 {
		params.Offset = getDownloadOffset(cli)
	}

	defer cli.Close()

	if err := cli.Authenticate(false, nil); err != nil {
		return err
	}

	photoCli, err := cli.PhotoCli()
	if err != nil {
		return err
	}

	if err := downloadPhoto(cli, photoCli, params.Output, params.Album, int(params.Recent), params.Offset, params.StopNum, params.ThreadNum); err != nil {
		return err
	}

	if params.AutoDelete {
		return autoDeletePhoto(photoCli, params.Output, params.ThreadNum)
	}

	return nil
}

type downloadParam struct {
	Username   string
	Password   string
	CookieDir  string
	Domain     string
	Output     string
	Recent     int64
	Offset     int
	StopNum    int64
	Album      string
	ThreadNum  int
	AutoDelete bool
}

func getDownloadParam(c *cli.Context) *downloadParam {
	return &downloadParam{
		Username:   c.String("username"),
		Password:   c.String("password"),
		CookieDir:  c.String("cookie-dir"),
		Domain:     c.String("domain"),
		Output:     c.String("output"),
		Offset:     c.Int("offset"),
		Recent:     c.Int64("recent"),
		StopNum:    c.Int64("stop-found-num"),
		Album:      c.String("album"),
		ThreadNum:  c.Int("thread-num"),
		AutoDelete: c.Bool("auto-delete"),
	}
}

func downloadPhoto(cli *icloudgo.Client, photoCli *icloudgo.PhotoService, outputDir, albumName string, recent, downloadOffset int, stopNum int64, threadNum int) error {
	if f, _ := os.Stat(outputDir); f == nil {
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			return err
		}
	}

	album, err := photoCli.GetAlbum(albumName)
	if err != nil {
		return err
	}

	defer saveDownloadOffset(cli, downloadOffset)
	fmt.Printf("album: %s, total: %d, offset: %d, target: %s, thread-num: %d\n", album.Name, album.Size(), downloadOffset, outputDir, threadNum)

	if recent == 0 {
		recent, err = album.GetSize()
		if err != nil {
			return err
		}
	}

	photoIter := album.PhotosIter(downloadOffset)
	wait := new(sync.WaitGroup)
	foundDownloadedNum := int64(0)
	var downloaded int32
	var finalErr error
	for threadIndex := 0; threadIndex < threadNum; threadIndex++ {
		wait.Add(1)
		go func(threadIndex int) {
			defer wait.Done()

			for {
				if atomic.LoadInt32(&downloaded) >= int32(recent) {
					return
				}
				if stopNum > 0 && atomic.LoadInt64(&foundDownloadedNum) >= stopNum {
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
					saveDownloadOffset(cli, offset)
					downloadOffset = offset
				}

				if isDownloaded, err := downloadPhotoAsset(photoAsset, outputDir, threadIndex); err != nil {
					if finalErr != nil {
						finalErr = err
					}
					return
				} else if isDownloaded {
					atomic.AddInt64(&foundDownloadedNum, 1)
					if stopNum > 0 && foundDownloadedNum >= stopNum {
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

func downloadPhotoAsset(photo *icloudgo.PhotoAsset, outputDir string, threadIndex int) (bool, error) {
	filename := photo.Filename()
	path := photo.LocalPath(outputDir, icloudgo.PhotoVersionOriginal)
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

func autoDeletePhoto(photoCli *icloudgo.PhotoService, outputDir string, threadNum int) error {
	album, err := photoCli.GetAlbum(icloudgo.AlbumNameRecentlyDeleted)
	if err != nil {
		return err
	}

	fmt.Printf("auto delete album: %s, total: %d\n", album.Name, album.Size())

	photoIter := album.PhotosIter(0)
	wait := new(sync.WaitGroup)
	var finalErr error
	for threadIndex := 0; threadIndex < threadNum; threadIndex++ {
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

				path := photoAsset.LocalPath(outputDir, icloudgo.PhotoVersionOriginal)

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

func saveDownloadOffset(cli *icloudgo.Client, i int) {
	err := cli.SaveConfig(configDownloadOffset, []byte(strconv.Itoa(i)))
	if err != nil {
		fmt.Printf("save download_offset config failed: %s", err)
	}
}
