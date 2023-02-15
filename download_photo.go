package icloudgo

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

type DownloadPhotoReq struct {
	OutputDir       string
	Album           string
	Count           int
	DuplicatePolicy DuplicatePolicy // 0: skip, 1: overwrite, 2: rename
}

type DuplicatePolicy int

const (
	SkipDuplicatePolicy      DuplicatePolicy = 0
	OverwriteDuplicatePolicy DuplicatePolicy = 1
	RenameDuplicatePolicy    DuplicatePolicy = 2
)

func (r *PhotoService) DownloadPhoto(req *DownloadPhotoReq) error {
	if f, _ := os.Stat(req.OutputDir); f == nil {
		if err := os.MkdirAll(req.OutputDir, os.ModePerm); err != nil {
			return err
		}
	}

	albums, err := r.Albums()
	if err != nil {
		return err
	}

	album := albums[AlbumNameAll]
	if req.Album != "" {
		var ok bool
		album, ok = albums[req.Album]
		if !ok {
			return fmt.Errorf("album %s not found", req.Album)
		}
	}

	if err = r.downloadPhotoAlbum(album, req.Count, req.OutputDir, req.DuplicatePolicy); err != nil {
		return err
	}

	return nil
}

func (r *PhotoService) downloadPhotoAlbum(album *PhotoAlbum, count int, outputDir string, duplicatePolicy DuplicatePolicy) error {
	fmt.Printf("album: %s, total: %d, target: %s, dyp policy: %v\n", album.Name, album.Size(), outputDir, duplicatePolicy)
	var err error
	if count == 0 {
		count, err = album.getSize()
		if err != nil {
			return err
		}
	}

	photos, err := album.Photos(count)
	if err != nil {
		return err
	}

	for _, photoAsset := range photos {
		if err := r.downloadPhotoAsset(photoAsset, outputDir, duplicatePolicy); err != nil {
			return err
		}
	}

	return nil
}

func (r *PhotoService) downloadPhotoAsset(photo *PhotoAsset, outputDir string, duplicatePolicy DuplicatePolicy) error {
	fmt.Printf("start %v, %v, %v\n", photo.ID(), photo.Filename(), photo.Size())
	ext := filepath.Ext(photo.Filename())
	pureFilename := strings.ReplaceAll(photo.ID(), "/", "-")
	path := filepath.Join(outputDir, pureFilename+ext)

	f, _ := os.Stat(path)
	isFileDup := f != nil && photo.Size() != int(f.Size())
	if isFileDup && duplicatePolicy == RenameDuplicatePolicy {
		for i := 2; i < 10000; i++ {
			path = filepath.Join(outputDir, fmt.Sprintf("%s(%d)%s", pureFilename, i, ext))
			if f, _ := os.Stat(path); f == nil {
				break
			}
		}
	}

	if f, _ := os.Stat(path); f != nil {
		if photo.Size() != int(f.Size()) {
			switch duplicatePolicy {
			case SkipDuplicatePolicy:
				fmt.Printf("file '%s' exist, skip.\n", path)
			case OverwriteDuplicatePolicy:
				fmt.Printf("file '%s' exist, overwrite.\n", path)
				return r.download(photo, path)
			case RenameDuplicatePolicy:
				fmt.Printf("file '%s' exist, rename.\n", path)
				return r.download(photo, path)
			default:
				return fmt.Errorf("unknown duplicate policy")
			}
		} else {
			fmt.Printf("file '%s' exist, skip.\n", path)
		}
	} else {
		return r.download(photo, path)
	}
	//        if auto_delete:
	//            photo.delete()
	return nil
}

func (r *PhotoService) download(photo *PhotoAsset, target string) error {
	body, err := photo.Download(PhotoVersionOriginal)
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
