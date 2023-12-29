package internal

import (
	"sync"
)

type DriveService struct {
	icloud          *Client
	serviceRoot     string
	serviceEndpoint string

	lock *sync.Mutex
}

func (r *Client) DriveCli() (*DriveService, error) {
	if r.drive == nil {
		driveWS, err := r.getWebServiceURL(serviceDrive)
		if err != nil {
			return nil, err
		}
		r.drive, err = newDriveService(r, driveWS)
		if err != nil {
			return nil, err
		}
	}
	return r.drive, nil
}

func newDriveService(icloud *Client, serviceRoot string) (*DriveService, error) {
	photoCli := &DriveService{
		icloud:          icloud,
		serviceRoot:     serviceRoot,
		serviceEndpoint: serviceRoot,

		lock: new(sync.Mutex),
	}

	return photoCli, nil
}
