package command

import (
	"errors"
	"fmt"
	"time"

	"github.com/chyroc/icloudgo"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PhotoAssetModel struct {
	ID     string `gorm:"column:id; index:uniq_id,unique"`
	Name   string `gorm:"column:name"`
	Data   string `gorm:"column:data"`
	Status int    `gorm:"column:status"`
}

func (PhotoAssetModel) TableName() string {
	return "photo_asset"
}

func (r *downloadCommand) dalAddAssets(assets []*icloudgo.PhotoAsset) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	pos := []*PhotoAssetModel{}
	for _, v := range assets {
		pos = append(pos, &PhotoAssetModel{
			ID:     v.ID(),
			Data:   string(v.Bytes()),
			Status: 0,
		})
	}
	return r.db.Clauses(clause.Insert{Modifier: "OR IGNORE"}).Create(pos).Error
}

func (r *downloadCommand) dalDeleteAsset(id string) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.db.Delete(&PhotoAssetModel{}, "id = ?", id).Error
}

func (r *downloadCommand) dalGetUnDownloadAssets() ([]*PhotoAssetModel, error) {
	r.lock.Lock()
	defer r.lock.Unlock()
	var pos []*PhotoAssetModel
	err := r.db.Model(&PhotoAssetModel{}).Where("status = ?", 0).Find(&pos).Error
	return pos, err
}

func (r *downloadCommand) dalSetDownloaded(id string) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.db.Model(&PhotoAssetModel{}).Where("id = ?", id).Update("status", 1).Error
}

type DownloadOffsetModel struct {
	AlbumName string    `gorm:"column:album_name; index:uniq_album_name,unique"`
	Offset    int       `gorm:"column:offset"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (DownloadOffsetModel) TableName() string {
	return "download_offset"
}

func (r *downloadCommand) getDownloadOffset() int {
	r.lock.Lock()
	defer r.lock.Unlock()
	var offset DownloadOffsetModel
	err := r.db.Model(&DownloadOffsetModel{}).Where("album_name = ?", r.AlbumName).First(&offset).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0
		}
		fmt.Printf("[icloudgo] [meta] get download offset err: %s, reset to 0\n", err)
		return 0
	}
	if time.Now().Sub(offset.UpdatedAt) > time.Hour*24 {
		fmt.Printf("[icloudgo] [meta] download offset is expired, reset to 0\n")
		_ = r.saveDownloadOffset(0, false)
		return 0
	}
	return offset.Offset
}

func (r *downloadCommand) saveDownloadOffset(offset int, needLock bool) error {
	if needLock {
		r.lock.Lock()
		defer r.lock.Unlock()
	}
	return r.db.Clauses(clause.Insert{Modifier: "OR REPLACE"}).Create(&DownloadOffsetModel{
		AlbumName: r.AlbumName,
		Offset:    offset,
		UpdatedAt: time.Now(),
	}).Error
}
