package internal

import (
	"encoding/json"
	"fmt"
	"net/http"

	uuid "github.com/satori/go.uuid"
)

func (r *DriveService) CreateFolder(parentDriveID, name string) (*DriveFolder, error) {
	return r.createDriveFolder(parentDriveID, name)
}

func (r *DriveService) createDriveFolder(parentDriveID, name string) (*DriveFolder, error) {
	clientID := uuid.NewV4().String()
	text, err := r.icloud.request(&rawReq{
		Method:  http.MethodPost,
		URL:     r.serviceEndpoint + "/createFolders",
		Headers: r.icloud.getCommonHeaders(map[string]string{"Content-type": "text/plain"}),
		Body:    fmt.Sprintf(`{"destinationDrivewsId":"%s","folders":[{"clientId":"FOLDER::%s::%s","name":"%s"}]}`, parentDriveID, clientID, clientID, name),
	})
	if err != nil {
		return nil, fmt.Errorf("createDriveFolder failed, err: %w", err)
	}

	var res = new(createDriveFolderResp)
	if err = json.Unmarshal([]byte(text), res); err != nil {
		return nil, fmt.Errorf("createDriveFolder unmarshal failed, err: %w, text: %s", err, text)
	}
	if len(res.Folders) == 0 {
		return nil, fmt.Errorf("createDriveFolder failed, no folder response, text: %s", text)
	}
	return res.Folders[0], nil
}

type createDriveFolderResp struct {
	DestinationDrivewsId string         `json:"destinationDrivewsId"`
	Folders              []*DriveFolder `json:"folders"`
}
