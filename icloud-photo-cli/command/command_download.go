package command

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

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
		&cli.StringFlag{
			Name:     "duplicate",
			Usage:    "duplicate policy, if not set, means skip",
			Required: false,
			Aliases:  []string{"dup"},
			EnvVars:  []string{"ICLOUD_DUPLICATE"},
			Action: func(context *cli.Context, s string) error {
				if s != "skip" && s != "rename" && s != "overwrite" && s != "" {
					return fmt.Errorf("invalid duplicate policy: %s, should be skip, rename, overwrite", s)
				}
				return nil
			},
		},
		&cli.IntFlag{
			Name:     "thread-num",
			Usage:    "thread num, if not set, means 1",
			Required: false,
			Aliases:  []string{"t"},
			Value:    1,
			EnvVars:  []string{"ICLOUD_THREAD_NUM"},
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
	duplicate := c.String("duplicate")
	threadNum := c.Int("thread-num")

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

	return downloadPhoto(cli, output, album, int(recent), duplicate, threadNum)
}

const (
	downloadPhotoDuplicatePolicySkip      = "skip"
	downloadPhotoDuplicatePolicyOverwrite = "overwrite"
	downloadPhotoDuplicatePolicyRename    = "rename"
)

func downloadPhoto(cli *icloudgo.Client, output, albumName string, recent int, duplicate string, threadNum int) error {
	if f, _ := os.Stat(output); f == nil {
		if err := os.MkdirAll(output, os.ModePerm); err != nil {
			return err
		}
	}

	photoCli, err := cli.PhotoCli()
	if err != nil {
		return err
	}

	albums, err := photoCli.Albums()
	if err != nil {
		return err
	}

	album := albums[icloudgo.AlbumNameAll]
	if albumName != "" {
		var ok bool
		album, ok = albums[albumName]
		if !ok {
			return fmt.Errorf("album %s not found", albumName)
		}
	}

	if err = downloadPhotoAlbum(album, output, recent, duplicate, threadNum); err != nil {
		return err
	}

	return nil
}

func downloadPhotoAlbum(album *icloudgo.PhotoAlbum, outputDir string, count int, duplicatePolicy string, threadNum int) error {
	fmt.Printf("album: %s, total: %d, target: %s, dup policy: %v, thread-num: %d\n", album.Name, album.Size(), outputDir, duplicatePolicy, threadNum)
	var err error
	if count == 0 {
		count, err = album.GetSize()
		if err != nil {
			return err
		}
	}

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
					if finalErr != nil {
						finalErr = err
					}
					return
				}

				if err := downloadPhotoAsset(photoAsset, outputDir, duplicatePolicy, threadIndex); err != nil {
					if finalErr != nil {
						finalErr = err
					}
					return
				}
			}
		}(threadIndex)
	}
	wait.Wait()

	return finalErr
}

func downloadPhotoAsset(photo *icloudgo.PhotoAsset, outputDir string, duplicatePolicy string, threadIndex int) error {
	filename := photo.Filename()
	fmt.Printf("start %v, %v, %v, thread=%d\n", photo.ID(), filename, photo.FormatSize(), threadIndex)
	ext := filepath.Ext(filename)
	filename = strings.ReplaceAll(filename, "/", "-")
	filename = filename[:len(filename)-len(ext)]
	path := filepath.Join(outputDir, filename+ext)

	f, _ := os.Stat(path)
	isFileDup := f != nil && photo.Size() != int(f.Size())
	if isFileDup && duplicatePolicy == downloadPhotoDuplicatePolicyRename {
		for i := 2; i < 10000; i++ {
			path = filepath.Join(outputDir, fmt.Sprintf("%s(%d)%s", filename, i, ext))
			if f, _ := os.Stat(path); f == nil {
				break
			}
		}
	}

	if f, _ := os.Stat(path); f != nil {
		if photo.Size() != int(f.Size()) {
			switch duplicatePolicy {
			case downloadPhotoDuplicatePolicySkip:
				fmt.Printf("file '%s' exist, skip.\n", path)
			case downloadPhotoDuplicatePolicyOverwrite:
				fmt.Printf("file '%s' exist, overwrite.\n", path)
				return downloadPhotoAssetData(photo, path)
			case downloadPhotoDuplicatePolicyRename:
				fmt.Printf("file '%s' exist, rename.\n", path)
				return downloadPhotoAssetData(photo, path)
			default:
				return fmt.Errorf("unknown duplicate policy")
			}
		} else {
			fmt.Printf("file '%s' exist, skip.\n", path)
		}
	} else {
		return downloadPhotoAssetData(photo, path)
	}
	//        if auto_delete:
	//            photo.delete()
	return nil
}

func downloadPhotoAssetData(photo *icloudgo.PhotoAsset, target string) error {
	body, err := photo.Download(icloudgo.PhotoVersionOriginal)
	if err != nil {
		return err
	}
	defer body.Close()

	f, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE, os.ModePerm)
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
