package command

import (
	"errors"
	"fmt"
	"io"
	"os"
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
	username := c.String("username")
	password := c.String("password")
	cookieDir := c.String("cookie-dir")
	domain := c.String("domain")
	output := c.String("output")
	recent := c.Int64("recent")
	album := c.String("album")
	threadNum := c.Int("thread-num")
	autoDelete := c.Bool("auto-delete")

	cli, err := icloudgo.New(&icloudgo.ClientOption{
		AppID:           username,
		CookieDir:       cookieDir,
		PasswordGetter:  getTextInput("apple id password", password),
		TwoFACodeGetter: getTextInput("2fa code", ""),
		Domain:          domain,
	})
	if err != nil {
		return err
	}

	defer cli.Close()

	if err := cli.Authenticate(false, nil); err != nil {
		return err
	}

	photoCli, err := cli.PhotoCli()
	if err != nil {
		return err
	}

	if err := downloadPhoto(photoCli, output, album, int(recent), threadNum); err != nil {
		return err
	}

	if autoDelete {
		if err := autoDeletePhoto(photoCli, output, threadNum); err != nil {
			return err
		}
	}

	return nil
}

func downloadPhoto(photoCli *icloudgo.PhotoService, outputDir, albumName string, recent int, threadNum int) error {
	if f, _ := os.Stat(outputDir); f == nil {
		if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
			return err
		}
	}

	album, err := photoCli.GetAlbum(albumName)
	if err != nil {
		return err
	}

	fmt.Printf("album: %s, total: %d, target: %s, thread-num: %d\n", album.Name, album.Size(), outputDir, threadNum)

	if recent == 0 {
		recent, err = album.GetSize()
		if err != nil {
			return err
		}
	}

	photoIter := album.PhotosIter()
	wait := new(sync.WaitGroup)
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

				if err := downloadPhotoAsset(photoAsset, outputDir, threadIndex); err != nil {
					if finalErr != nil {
						finalErr = err
					}
					return
				}

				atomic.AddInt32(&downloaded, 1)
			}
		}(threadIndex)
	}
	wait.Wait()

	return finalErr
}

func downloadPhotoAsset(photo *icloudgo.PhotoAsset, outputDir string, threadIndex int) error {
	filename := photo.Filename()
	path := photo.LocalPath(outputDir, icloudgo.PhotoVersionOriginal)
	fmt.Printf("start %v, %v, %v, thread=%d\n", photo.ID(), filename, photo.FormatSize(), threadIndex)

	if f, _ := os.Stat(path); f != nil {
		if photo.Size() != int(f.Size()) {
			return downloadPhotoAssetData(photo, path)
		} else {
			fmt.Printf("file '%s' exist, skip.\n", path)
		}
	} else {
		return downloadPhotoAssetData(photo, path)
	}
	return nil
}

func downloadPhotoAssetData(photo *icloudgo.PhotoAsset, target string) error {
	body, err := photo.Download(icloudgo.PhotoVersionOriginal)
	if err != nil {
		return err
	}
	defer body.Close()

	f, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		return fmt.Errorf("open file error: %v", err)
	}

	_, err = io.Copy(f, body)
	if err != nil {
		return fmt.Errorf("copy file error: %v", err)
	}

	// modify_create_date

	return nil
}

func autoDeletePhoto(photoCli *icloudgo.PhotoService, outputDir string, threadNum int) error {
	album, err := photoCli.GetAlbum(icloudgo.AlbumNameRecentlyDeleted)
	if err != nil {
		return err
	}

	fmt.Printf("auto delete album: %s, total: %d\n", album.Name, album.Size())

	photoIter := album.PhotosIter()
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
