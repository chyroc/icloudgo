package command

import (
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/chyroc/icloudgo"
	"github.com/dgraph-io/badger/v3"
)

type PhotoAssetModel struct {
	ID     string `gorm:"column:id; index:uniq_id,unique"`
	Name   string `gorm:"column:name"`
	Data   string `gorm:"column:data"`
	Status int    `gorm:"column:status"`
}

func (r PhotoAssetModel) bytes() []byte {
	val, _ := json.Marshal(r)
	return val
}

func valToPhotoAssetModel(val []byte) (*PhotoAssetModel, error) {
	res := new(PhotoAssetModel)
	return res, json.Unmarshal(val, res)
}

func (r *downloadCommand) dalAddAssets(assets []*icloudgo.PhotoAsset) error {
	r.lock.Lock()
	defer r.lock.Unlock()
	return r.db.Update(func(txn *badger.Txn) error {
		for _, v := range assets {
			po := &PhotoAssetModel{
				ID:     v.ID(),
				Data:   string(v.Bytes()),
				Status: 0,
			}
			if err := txn.Set(r.keyAssert(v.ID()), po.bytes()); err != nil {
				return err
			}
		}
		return nil
	})
}

func (r *downloadCommand) dalDeleteAsset(id string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.db.Update(func(txn *badger.Txn) error {
		return txn.Delete(r.keyAssert(id))
	})
}

func (r *downloadCommand) dalGetUnDownloadAssets(status *int) ([]*PhotoAssetModel, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	pos := []*PhotoAssetModel{}
	err := r.db.Update(func(txn *badger.Txn) error {
		it := txn.NewIterator(badger.DefaultIteratorOptions)
		defer it.Close()
		for it.Seek(r.keyAssertPrefix()); it.ValidForPrefix(r.keyAssertPrefix()); it.Next() {
			val, err := it.Item().ValueCopy(nil)
			if err != nil {
				return err
			}
			po, err := valToPhotoAssetModel(val)
			if err != nil {
				return err
			}
			if status == nil {
				pos = append(pos, po)
			} else if po.Status == *status {
				pos = append(pos, po)
			}
		}
		return nil
	})

	return pos, err
}

func (r *downloadCommand) dalSetDownloaded(id string) error {
	r.lock.Lock()
	defer r.lock.Unlock()

	return r.db.Update(func(txn *badger.Txn) error {
		item, err := txn.Get(r.keyAssert(id))
		if err != nil {
			return err
		}
		val, err := item.ValueCopy(nil)
		if err != nil {
			return err
		}
		po, err := valToPhotoAssetModel(val)
		if err != nil {
			return err
		}
		po.Status = 1
		return txn.Set(r.keyAssert(id), po.bytes())
	})
}

func (r *downloadCommand) keyAssertPrefix() []byte {
	return []byte("assert_")
}

func (r *downloadCommand) keyAssert(id string) []byte {
	return []byte("assert_" + id)
}

func (r *downloadCommand) dalGetDownloadOffset(albumSize int64) int64 {
	r.lock.Lock()
	defer r.lock.Unlock()

	var result int64
	_ = r.db.Update(func(txn *badger.Txn) error {
		offset, err := r.getDownloadOffset(txn, false)
		if err != nil {
			if errors.Is(err, badger.ErrKeyNotFound) {
				return nil
			}
			fmt.Printf("[icloudgo] [offset] get db offset err: %s, reset to 0\n", err)
			return nil
		}
		fmt.Printf("[icloudgo] [offset] get db offset: %d\n", offset)
		if offset > albumSize {
			result = 0
			if err = r.saveDownloadOffset(txn, 0, false); err != nil {
				fmt.Printf("[icloudgo] [offset] db offset=%d, album_size=%d, reset to 0, and save_db failed: %s\n", offset, albumSize, err)
			} else {
				fmt.Printf("[icloudgo] [offset] db offset=%d, album_size=%d, reset to 0\n", offset, albumSize)
			}
		}
		result = offset
		return nil
	})
	return result
}

func (r *downloadCommand) getDownloadOffset(txn *badger.Txn, needLock bool) (int64, error) {
	if needLock {
		r.lock.Lock()
		defer r.lock.Unlock()
	}
	item, err := txn.Get(r.keyOffset())
	if err != nil {
		return 0, err
	} else if item.IsDeletedOrExpired() {
		return 0, badger.ErrKeyNotFound
	}
	val, err := item.ValueCopy(nil)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(string(val), 10, 64)
}

func (r *downloadCommand) saveDownloadOffset(txn *badger.Txn, offset int64, needLock bool) error {
	if needLock {
		r.lock.Lock()
		defer r.lock.Unlock()
	}
	if txn == nil {
		return r.db.Update(func(txn *badger.Txn) error {
			e := badger.NewEntry(r.keyOffset(), []byte(strconv.FormatInt(offset, 10)))
			e.ExpiresAt = uint64(time.Now().Add(time.Hour * 12).Unix())
			return txn.SetEntry(e)
		})
	}
	e := badger.NewEntry(r.keyOffset(), []byte(strconv.FormatInt(offset, 10)))
	e.ExpiresAt = uint64(time.Now().Add(time.Hour * 12).Unix())
	return txn.SetEntry(e)
}

func (r *downloadCommand) keyOffset() []byte {
	return []byte("download_offset_" + r.AlbumName)
}
