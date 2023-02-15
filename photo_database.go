package icloudgo

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (r *PhotoService) checkPhotoServiceState() error {
	text, err := r.icloud.Request(&reqParam{
		Method:  http.MethodPost,
		URL:     fmt.Sprintf("%s/records/query", r.serviceEndpoint),
		Body:    `{"query":{"recordType":"CheckIndexingState"},"zoneID":{"zoneName":"PrimarySync"}}`,
		Headers: r.icloud.getCommonHeaders(map[string]string{"Content-type": "text/plain"}),
		Querys:  r.querys,
	})
	if err != nil {
		return fmt.Errorf("checkPhotoServiceState failed, err: %w", err)
	}
	res := new(getPhotoDatabaseResp)
	if err = json.Unmarshal([]byte(text), res); err != nil {
		return fmt.Errorf("checkPhotoServiceState unmarshal failed, err: %w, text: %s", err, text)
	}

	if len(res.Records) > 0 {
		if res.Records[0].Fields.State.Value != "FINISHED" {
			return fmt.Errorf("iCloud Photo Library not finished indexing. Please try again in a few minutes.")
		}
	}

	return nil
}

type getPhotoDatabaseResp struct {
	Records   []*photoDatabaseRecord `json:"records"`
	SyncToken string                 `json:"syncToken"`
}

type photoDatabaseRecord struct {
	RecordName string `json:"recordName"`
	RecordType string `json:"recordType"`
	Fields     struct {
		Progress struct {
			Value int    `json:"value"`
			Type  string `json:"type"`
		} `json:"progress"`
		State struct {
			Value string `json:"value"`
			Type  string `json:"type"`
		} `json:"state"`
	} `json:"fields"`
	PluginFields struct {
	} `json:"pluginFields"`
	RecordChangeTag string `json:"recordChangeTag"`
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
