package internal

import (
	"sync"
)

type PhotosIterNext interface {
	Next() (*PhotoAsset, error)
}

type photosIterNextImpl struct {
	album  *PhotoAlbum
	lock   *sync.Mutex
	offset int
	assets []*PhotoAsset
	index  int
	end    bool
}

func (r *photosIterNextImpl) Next() (*PhotoAsset, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r.index < len(r.assets) {
		r.index++
		return r.assets[r.index-1], nil
	}

	if r.end {
		return nil, ErrPhotosIterateEnd
	}

	assets, err := r.album.GetPhotosByOffset(r.offset, 200)
	if err != nil {
		return nil, err
	}
	r.index = 1
	r.assets = assets
	r.offset = r.album.calOffset(r.offset, len(assets))
	r.end = len(assets) == 0

	return r.assets[r.index-1], nil
}
