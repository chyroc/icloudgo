package internal

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type PhotoVersion string

const (
	PhotoVersionOriginal PhotoVersion = "original"
	PhotoVersionMedium   PhotoVersion = "medium"
	PhotoVersionThumb    PhotoVersion = "thumb"
)

func (r *PhotoAsset) DownloadTo(version PhotoVersion, target string) error {
	body, err := r.Download(version)
	if body != nil {
		defer body.Close()
	}
	if err != nil {
		return err
	}

	f, err := os.OpenFile(target, os.O_RDWR|os.O_CREATE, 0o644)
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

func (r *PhotoAsset) Download(version PhotoVersion) (io.ReadCloser, error) {
	versionDetail, ok := r.getVersions()[version]
	if !ok {
		var keys []string
		for k := range r.getVersions() {
			keys = append(keys, string(k))
		}
		return nil, fmt.Errorf("version %s not found, valid: %s", version, strings.Join(keys, ","))
	}

	body, err := r.service.icloud.requestStream(&rawReq{
		Method:       http.MethodGet,
		URL:          versionDetail.URL,
		Headers:      r.service.icloud.getCommonHeaders(map[string]string{}),
		ExpectStatus: newSet[int](http.StatusOK),
	})
	if err != nil {
		return body, fmt.Errorf("download %s failed: %w", r.Filename(), err)
	}
	return body, nil
}

func (r *PhotoAsset) getVersions() map[PhotoVersion]*photoVersionDetail {
	r.lock.Lock()
	defer r.lock.Unlock()

	if len(r._versions) == 0 {
		r._versions = r.packVersion()
	}

	return r._versions
}

func (r *PhotoAsset) packVersion() map[PhotoVersion]*photoVersionDetail {
	fields := r._masterRecord.Fields

	if fields.ResVidSmallRes.Type != "" || fields.ResVidSmallRes.Value.Size != 0 {
		return map[PhotoVersion]*photoVersionDetail{
			PhotoVersionOriginal: {
				Filename: r.Filename(),
				Width:    fields.ResOriginalWidth.Value,
				Height:   fields.ResOriginalHeight.Value,
				Size:     fields.ResOriginalRes.Value.Size,
				URL:      fields.ResOriginalRes.Value.DownloadURL,
				Type:     fields.ResOriginalFileType.Value,
			},
			PhotoVersionMedium: {
				Filename: r.Filename(),
				Width:    fields.ResJPEGMedWidth.Value,
				Height:   fields.ResJPEGMedHeight.Value,
				Size:     fields.ResJPEGMedRes.Value.Size,
				URL:      fields.ResJPEGMedRes.Value.DownloadURL,
				Type:     fields.ResJPEGMedFileType.Value,
			},
			PhotoVersionThumb: {
				Filename: r.Filename(),
				Width:    fields.ResJPEGThumbWidth.Value,
				Height:   fields.ResJPEGThumbHeight.Value,
				Size:     fields.ResJPEGThumbRes.Value.Size,
				URL:      fields.ResJPEGThumbRes.Value.DownloadURL,
				Type:     fields.ResJPEGThumbFileType.Value,
			},
		}
	} else {
		return map[PhotoVersion]*photoVersionDetail{
			PhotoVersionOriginal: {
				Filename: r.Filename(),
				Width:    fields.ResOriginalWidth.Value,
				Height:   fields.ResOriginalHeight.Value,
				Size:     fields.ResOriginalRes.Value.Size,
				URL:      fields.ResOriginalRes.Value.DownloadURL,
				Type:     fields.ResOriginalFileType.Value,
			},
			PhotoVersionMedium: {
				Filename: r.Filename(),
				Width:    fields.ResVidMedWidth.Value,
				Height:   fields.ResVidMedHeight.Value,
				Size:     fields.ResVidMedRes.Value.Size,
				URL:      fields.ResVidMedRes.Value.DownloadURL,
				Type:     fields.ResVidMedFileType.Value,
			},
			PhotoVersionThumb: {
				Filename: r.Filename(),
				Width:    fields.ResVidSmallWidth.Value,
				Height:   fields.ResVidSmallHeight.Value,
				Size:     fields.ResVidSmallRes.Value.Size,
				URL:      fields.ResVidSmallRes.Value.DownloadURL,
				Type:     fields.ResVidSmallFileType.Value,
			},
		}
	}
}

type photoVersionDetail struct {
	Filename string `json:"filename"`
	Width    int    `json:"width"`
	Height   int    `json:"height"`
	Size     int    `json:"size"`
	URL      string `json:"url"`
	Type     string `json:"type"`
}
