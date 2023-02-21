package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

const (
	DriveRootID  = "FOLDER::com.apple.CloudDocs::root"
	DriveTrashID = "TRASH_ROOT"
)

func (r *DriveService) Folders(driveID string) (int, []*DriveFolder, error) {
	return r.getDriveFolders(driveID)
}

func (r *DriveService) getDriveFolders(driveID string) (int, []*DriveFolder, error) {
	text, err := r.icloud.request(&rawReq{
		Method:  http.MethodPost,
		URL:     r.serviceEndpoint + "/retrieveItemDetailsInFolders",
		Headers: r.icloud.getCommonHeaders(map[string]string{"Content-type": "text/plain"}),
		Body:    fmt.Sprintf(`[{"drivewsid":"%s","partialData":false,"includeHierarchy":true}]`, driveID),
	})
	if err != nil {
		return 0, nil, fmt.Errorf("getDriveFolders failed, err: %w", err)
	}

	var res []*getDriveFoldersResp
	if err = json.Unmarshal([]byte(text), &res); err != nil {
		return 0, nil, fmt.Errorf("getDriveFolders unmarshal failed, err: %w, text: %s", err, text)
	}
	if len(res) == 0 {
		return 0, nil, nil
	}
	return res[0].NumberOfItems, res[0].Items, nil
}

type DriveFolder struct {
	DateCreated         time.Time `json:"dateCreated"`
	Drivewsid           string    `json:"drivewsid"`
	Docwsid             string    `json:"docwsid"`
	Zone                string    `json:"zone"`
	Name                string    `json:"name"`
	ParentId            string    `json:"parentId"`
	Etag                string    `json:"etag"`
	Type                string    `json:"type"`
	AssetQuota          int       `json:"assetQuota,omitempty"`
	FileCount           int       `json:"fileCount,omitempty"`
	ShareCount          int       `json:"shareCount,omitempty"`
	ShareAliasCount     int       `json:"shareAliasCount,omitempty"`
	DirectChildrenCount int       `json:"directChildrenCount,omitempty"`
	MaxDepth            string    `json:"maxDepth,omitempty"`
	Icons               []struct {
		Url  string `json:"url"`
		Type string `json:"type"`
		Size int    `json:"size"`
	} `json:"icons,omitempty"`
	SupportedExtensions []string  `json:"supportedExtensions,omitempty"`
	SupportedTypes      []string  `json:"supportedTypes,omitempty"`
	Extension           string    `json:"extension,omitempty"`
	DateModified        time.Time `json:"dateModified,omitempty"`
	DateChanged         time.Time `json:"dateChanged,omitempty"`
	Size                int       `json:"size,omitempty"`
}

type getDriveFoldersResp struct {
	NumberOfItems int            `json:"numberOfItems"`
	Items         []*DriveFolder `json:"items"`
}
