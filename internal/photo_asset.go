package internal

import (
	"encoding/base64"
	"fmt"
	"path/filepath"
	"sync"
)

type PhotoAsset struct {
	service       *PhotoService
	_versions     map[PhotoVersion]*photoVersionDetail
	_masterRecord *photoRecord
	_assetRecord  *photoRecord
	lock          *sync.Mutex
}

func (r *PhotoService) newPhotoAsset(masterRecord, assetRecords *photoRecord) *PhotoAsset {
	return &PhotoAsset{
		service:       r,
		_masterRecord: masterRecord,
		_assetRecord:  assetRecords,
		_versions:     nil,
		lock:          new(sync.Mutex),
	}
}

func (r *PhotoAsset) Filename() string {
	if v := r._masterRecord.Fields.FilenameEnc.Value; v != "" {
		bs, _ := base64.StdEncoding.DecodeString(v)
		if len(bs) > 0 {
			return cleanFilename(string(bs))
		}
	}

	return cleanFilename(r.ID())
}

func (r *PhotoAsset) LocalPath(outputDir string, size PhotoVersion) string {
	filename := r.Filename()
	ext := filepath.Ext(filename)
	filename = filename[:len(filename)-len(ext)]

	if size == PhotoVersionOriginal || size == "" {
		return filepath.Join(outputDir, filename+ext)
	}

	return filepath.Join(outputDir, filename+"_"+string(size)+ext)
}

func (r *PhotoAsset) ID() string {
	return r._masterRecord.RecordName
}

func (r *PhotoAsset) Size() int {
	return r._masterRecord.Fields.ResOriginalRes.Value.Size
}

func (r *PhotoAsset) FormatSize() string {
	return formatSize(r.Size())
}

func formatSize(size int) string {
	if size < 1024 {
		return fmt.Sprintf("%dB", size)
	} else if size < 1024*1024 {
		return fmt.Sprintf("%.2fKB", float64(size)/1024)
	} else if size < 1024*1024*1024 {
		return fmt.Sprintf("%.2fMB", float64(size)/1024/1024)
	} else {
		return fmt.Sprintf("%.2fGB", float64(size)/1024/1024/1024)
	}
}
