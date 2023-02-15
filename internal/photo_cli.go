package internal

import (
	"fmt"
	"sync"
)

type PhotoService struct {
	icloud          *Client
	serviceRoot     string
	serviceEndpoint string
	querys          map[string]string

	_albums map[string]*PhotoAlbum
	lock    *sync.Mutex
}

func (r *Client) PhotoCli() (*PhotoService, error) {
	if r.photo == nil {
		ckDatabaseWS, err := r.getWebServiceURL("ckdatabasews")
		if err != nil {
			return nil, err
		}
		r.photo, err = newPhotoService(r, ckDatabaseWS)
		if err != nil {
			return nil, err
		}
	}
	return r.photo, nil
}

func newPhotoService(icloud *Client, serviceRoot string) (*PhotoService, error) {
	photoCli := &PhotoService{
		icloud:          icloud,
		serviceRoot:     serviceRoot,
		serviceEndpoint: fmt.Sprintf("%s/database/1/com.apple.photos.cloud/production/private", serviceRoot),
		querys:          map[string]string{"remapEnums": "true", "getCurrentSyncToken": "true"},

		_albums: map[string]*PhotoAlbum{},
		lock:    new(sync.Mutex),
	}

	if err := photoCli.checkPhotoServiceState(); err != nil {
		return nil, err
	}

	return photoCli, nil
}
