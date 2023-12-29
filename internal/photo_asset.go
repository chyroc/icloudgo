package internal

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type PhotoAsset struct {
	service         *PhotoService
	normalPhotos    map[PhotoVersion]*photoVersionDetail
	livePhotoVideos map[PhotoVersion]*photoVersionDetail
	_masterRecord   *photoRecord
	_assetRecord    *photoRecord
	lock            *sync.Mutex
}

func (r *PhotoAsset) Bytes() []byte {
	bs, _ := json.Marshal(photoAssetData{
		MasterRecord: r._masterRecord,
		AssetRecord:  r._assetRecord,
	})
	return bs
}

func (r *PhotoService) NewPhotoAssetFromBytes(bs []byte) *PhotoAsset {
	data := &photoAssetData{}
	_ = json.Unmarshal(bs, data)

	return r.newPhotoAsset(data.MasterRecord, data.AssetRecord)
}

type photoAssetData struct {
	MasterRecord *photoRecord `json:"master_record"`
	AssetRecord  *photoRecord `json:"asset_record"`
}

func (r *PhotoService) newPhotoAsset(masterRecord, assetRecords *photoRecord) *PhotoAsset {
	return &PhotoAsset{
		service:         r,
		normalPhotos:    nil,
		livePhotoVideos: nil,
		_masterRecord:   masterRecord,
		_assetRecord:    assetRecords,
		lock:            new(sync.Mutex),
	}
}

func (r *PhotoAsset) Filename(livePhoto bool) string {
	name := r.filename()
	if !livePhoto {
		return name
	}
	l := strings.SplitN(name, ".", 2)
	if len(l) == 2 {
		return l[0] + ".MOV"
	}
	return name + ".MOV"
}

func (r *PhotoAsset) filename() string {
	if v := r._masterRecord.Fields.FilenameEnc.Value; v != "" {
		bs, _ := base64.StdEncoding.DecodeString(v)
		if len(bs) > 0 {
			return cleanFilename(string(bs))
		}
	}

	return cleanFilename(r.ID())
}

func (r *PhotoAsset) LocalPath(outputDir string, size PhotoVersion, fileStructure string, livePhoto bool) string {
	filename := r.Filename(livePhoto)
	ext := filepath.Ext(filename)
	name := ""
	switch fileStructure {
	case "name":
		name = cleanFilename(filename)
	default:
		name = cleanFilename(r.ID())
	}

	if size == PhotoVersionOriginal || size == "" {
		return filepath.Join(outputDir, name+ext)
	}

	return filepath.Join(outputDir, name+"_"+string(size)+ext)
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

func (r *PhotoAsset) AddDate() time.Time {
	return time.UnixMilli(r._masterRecord.Created.Timestamp)
}

func (r *PhotoAsset) AssetDate() time.Time {
	return time.UnixMilli(r._assetRecord.Fields.AssetDate.Value)
}

func (r *PhotoAsset) OutputDir(output, folderStructure string) string {
	if folderStructure == "" || folderStructure == "/" {
		return output
	}

	assetDate := r.AssetDate().Format(folderStructure)
	return filepath.Join(output, assetDate)
}

// 仅为兼容性
func (r *PhotoAsset) OldOutputDir(output, folderStructure string) string {
	if folderStructure == "" || folderStructure == "/" {
		return output
	}

	assetDate := r.AddDate().Format(folderStructure)
	return filepath.Join(output, assetDate)
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
