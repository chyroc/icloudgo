package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type PhotoVersion string

const (
	PhotoVersionOriginal PhotoVersion = "original"
	PhotoVersionMedium   PhotoVersion = "medium"
	PhotoVersionThumb    PhotoVersion = "thumb"
)

func (r *PhotoAsset) DownloadTo(version PhotoVersion, livePhoto bool, target string) error {
	body, err := r.Download(version, livePhoto)
	if body != nil {
		defer body.Close()
	}
	if err != nil {
		return err
	}

	f, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE, 0o644)
	if f != nil {
		defer f.Close()
	}
	if err != nil {
		return fmt.Errorf("open file error: %v", err)
	}

	_, err = io.Copy(f, body)
	if err != nil {
		return fmt.Errorf("copy file error: %v", err)
	}

	// 1676381385791 to time.time
	created := r.Created()
	if err := os.Chtimes(target, created, created); err != nil {
		return fmt.Errorf("change file time error: %v", err)
	}

	return nil
}

func (r *PhotoAsset) Download(version PhotoVersion, livePhoto bool) (io.ReadCloser, error) {
	versionDetail, ok := r.getVersions(livePhoto)[version]
	if !ok {
		var keys []string
		for k := range r.getVersions(livePhoto) {
			keys = append(keys, string(k))
		}
		return nil, fmt.Errorf("version %s not found, valid: %s", version, strings.Join(keys, ","))
	}

	timeout := time.Minute * 10 // 10分钟
	if versionDetail.Size > 0 {
		slowSecond := time.Duration(versionDetail.Size/1024/100) * time.Second // 100 KB/s 秒
		if slowSecond > timeout {
			timeout = slowSecond
		}
	}

	body, err := r.service.icloud.requestStream(&rawReq{
		Method:       http.MethodGet,
		URL:          versionDetail.URL,
		Headers:      r.service.icloud.getCommonHeaders(map[string]string{}),
		ExpectStatus: newSet[int](http.StatusOK),
		Timeout:      timeout,
	})
	if err != nil {
		return body, fmt.Errorf("download %s(timeout: %s) failed: %w", r.Filename(livePhoto), timeout, err)
	}
	return body, nil
}

func (r *PhotoAsset) IsLivePhoto() bool {
	f := r._masterRecord.Fields
	return f.ResOriginalVidComplRes.Value.DownloadURL != "" &&
		f.ResOriginalRes.Value.DownloadURL != ""
}

func (r *PhotoAsset) getVersions(livePhoto bool) map[PhotoVersion]*photoVersionDetail {
	r.lock.Lock()
	defer r.lock.Unlock()

	if len(r.normalPhotos) == 0 {
		r.normalPhotos, r.livePhotoVideos = r.packVersion()
	}
	if livePhoto {
		return r.livePhotoVideos
	}

	return r.normalPhotos
}

func (r *PhotoAsset) packVersion() (map[PhotoVersion]*photoVersionDetail, map[PhotoVersion]*photoVersionDetail) {
	fields := r._masterRecord.Fields

	normal := map[PhotoVersion]*photoVersionDetail{
		PhotoVersionOriginal: {
			Filename: r.Filename(false),
			Width:    fields.ResOriginalWidth.Value,
			Height:   fields.ResOriginalHeight.Value,
			Size:     fields.ResOriginalRes.Value.Size,
			URL:      fields.ResOriginalRes.Value.DownloadURL,
			Type:     fields.ResOriginalFileType.Value,
		},
		PhotoVersionMedium: {
			Filename: r.Filename(false),
			Width:    fields.ResJPEGMedWidth.Value,
			Height:   fields.ResJPEGMedHeight.Value,
			Size:     fields.ResJPEGMedRes.Value.Size,
			URL:      fields.ResJPEGMedRes.Value.DownloadURL,
			Type:     fields.ResJPEGMedFileType.Value,
		},
		PhotoVersionThumb: {
			Filename: r.Filename(false),
			Width:    fields.ResJPEGThumbWidth.Value,
			Height:   fields.ResJPEGThumbHeight.Value,
			Size:     fields.ResJPEGThumbRes.Value.Size,
			URL:      fields.ResJPEGThumbRes.Value.DownloadURL,
			Type:     fields.ResJPEGThumbFileType.Value,
		},
	}
	livePhotoVideo := map[PhotoVersion]*photoVersionDetail{
		PhotoVersionOriginal: {
			Filename: r.Filename(true),
			Width:    fields.ResOriginalVidComplWidth.Value,
			Height:   fields.ResOriginalVidComplHeight.Value,
			Size:     fields.ResOriginalVidComplRes.Value.Size,
			URL:      fields.ResOriginalVidComplRes.Value.DownloadURL,
			Type:     fields.ResOriginalVidComplFileType.Value,
		},
		PhotoVersionMedium: {
			Filename: r.Filename(true),
			Width:    fields.ResVidMedWidth.Value,
			Height:   fields.ResVidMedHeight.Value,
			Size:     fields.ResVidMedRes.Value.Size,
			URL:      fields.ResVidMedRes.Value.DownloadURL,
			Type:     fields.ResVidMedFileType.Value,
		},
		PhotoVersionThumb: {
			Filename: r.Filename(true),
			Width:    fields.ResVidSmallWidth.Value,
			Height:   fields.ResVidSmallHeight.Value,
			Size:     fields.ResVidSmallRes.Value.Size,
			URL:      fields.ResVidSmallRes.Value.DownloadURL,
			Type:     fields.ResVidSmallFileType.Value,
		},
	}

	return normal, livePhotoVideo
}

type photoVersionDetail struct {
	Filename string `json:"filename"`
	Width    int64  `json:"width"`
	Height   int64  `json:"height"`
	Size     int    `json:"size"`
	URL      string `json:"url"`
	Type     string `json:"type"`
}
