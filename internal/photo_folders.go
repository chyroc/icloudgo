package internal

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (r *PhotoService) getFolders() ([]*folderRecord, error) {
	text, err := r.icloud.request(&rawReq{
		Method:  http.MethodPost,
		URL:     r.serviceEndpoint + "/records/query",
		Headers: r.icloud.getCommonHeaders(map[string]string{"Content-type": "text/plain"}),
		Body:    `{"query":{"recordType":"CPLAlbumByPositionLive"},"zoneID":{"zoneName":"PrimarySync"}}`,
	})
	if err != nil {
		return nil, fmt.Errorf("getFolders failed, err: %w", err)
	}

	res := new(getFoldersResp)
	if err = json.Unmarshal([]byte(text), res); err != nil {
		return nil, fmt.Errorf("getFolders unmarshal failed, err: %w, text: %s", err, text)
	}

	return res.Records, nil
}

type getFoldersResp struct {
	Records []*folderRecord `json:"records"`
}

type folderRecord struct {
	RecordName string `json:"recordName"`
	RecordType string `json:"recordType"`
	Fields     struct {
		RecordModificationDate struct {
			Value int64  `json:"value"`
			Type  string `json:"type"`
		} `json:"recordModificationDate"`
		SortAscending struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"sortAscending"`
		SortType struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"sortType"`
		AlbumType struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"albumType"`
		Position struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"position"`
		SortTypeExt struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"sortTypeExt"`
		ImportedByBundleIdentifierEnc struct {
			Value string `json:"value"`
			Type  string `json:"type"`
		} `json:"importedByBundleIdentifierEnc,omitempty"`
		AlbumNameEnc *folderTypeValue `json:"albumNameEnc,omitempty"`
		IsDeleted    *folderTypeValue `json:"isDeleted,omitempty"`
	} `json:"fields"`
	PluginFields    struct{} `json:"pluginFields"`
	RecordChangeTag string   `json:"recordChangeTag"`
	Created         struct {
		Timestamp      int64  `json:"timestamp"`
		UserRecordName string `json:"userRecordName"`
		DeviceID       string `json:"deviceID"`
	} `json:"created"`
	Modified struct {
		Timestamp      int64  `json:"timestamp"`
		UserRecordName string `json:"userRecordName"`
		DeviceID       string `json:"deviceID"`
	} `json:"modified"`
	Deleted bool `json:"deleted"`
	ZoneID  struct {
		ZoneName        string `json:"zoneName"`
		OwnerRecordName string `json:"ownerRecordName"`
		ZoneType        string `json:"zoneType"`
	} `json:"zoneID"`
}

type folderTypeValue struct {
	Value any    `json:"value"`
	Type  string `json:"type"`
}
