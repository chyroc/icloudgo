package internal

import (
	"fmt"
	"net/http"
)

func (r *PhotoAsset) Delete() error {
	body := fmt.Sprintf(`{"operations":[{"operationType":"update","record":{"recordName":"%s","recordType":"%s","recordChangeTag":"%s","fields":{"isDeleted":{"value":1}}}}],"zoneID":{"zoneName":"PrimarySync"},"atomic":true}`,
		r._assetRecord.RecordName,
		r._assetRecord.RecordType,
		r._masterRecord.RecordChangeTag,
	)
	_, err := r.service.icloud.request(&rawReq{
		Method:  http.MethodPost,
		URL:     fmt.Sprintf("%s/records/modify", r.service.serviceEndpoint),
		Querys:  r.service.querys,
		Headers: r.service.icloud.getCommonHeaders(map[string]string{}),
		Body:    body,
	})
	if err != nil {
		return fmt.Errorf("delete %s failed: %w", r.Filename(), err)
	}
	return nil
}
