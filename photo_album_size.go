package icloudgo

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (r *PhotoAlbum) Size() int {
	size, _ := r.GetSize()
	return size
}

func (r *PhotoAlbum) GetSize() (int, error) {
	r.lock.Lock()
	defer r.lock.Unlock()

	if r._length != nil {
		return *r._length, nil
	}

	size, err := r.getSize()
	if err != nil {
		return 0, err
	}

	r._length = &size
	return size, nil
}

func (r *PhotoAlbum) getSize() (int, error) {
	text, err := r.service.icloud.Request(&reqParam{
		Method:  http.MethodPost,
		URL:     fmt.Sprintf("%s/internal/records/query/batch", r.service.serviceEndpoint),
		Querys:  r.service.querys,
		Headers: r.service.icloud.getCommonHeaders(map[string]string{}),
		Body: map[string]any{
			"batch": []any{
				map[string]any{
					"resultsLimit": 1,
					"query": map[string]any{
						"filterBy": []any{
							map[string]any{
								"fieldName":  "indexCountID",
								"fieldValue": map[string]any{"type": "STRING_LIST", "value": []string{r.ObjType}},
								"comparator": "IN",
							},
						},
						"recordType": "HyperionIndexCountLookup",
					},
					"zoneWide": true,
					"zoneID":   map[string]string{"zoneName": "PrimarySync"},
				},
			},
		},
	})
	if err != nil {
		return 0, fmt.Errorf("get album size failed, err: %w", err)
	}
	res := new(getAlbumSizeResp)
	if err = json.Unmarshal([]byte(text), res); err != nil {
		return 0, fmt.Errorf("get album size unmarshal failed, err: %w, text: %s", err, text)
	} else if len(res.Batch) == 0 {
		return 0, fmt.Errorf("get album size failed, err: no batch response")
	} else if len(res.Batch[0].Records) == 0 {
		return 0, fmt.Errorf("get album size failed, err: no batch records response")
	}

	return res.Batch[0].Records[0].Fields.ItemCount.Value, nil
}

type getAlbumSizeResp struct {
	Batch []struct {
		Records []struct {
			RecordName string `json:"recordName"`
			RecordType string `json:"recordType"`
			Fields     struct {
				ItemCount struct {
					Value int    `json:"value"`
					Type  string `json:"type"`
				} `json:"itemCount"`
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
		} `json:"records"`
	} `json:"batch"`
}
