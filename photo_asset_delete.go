package icloudgo

import (
	"fmt"
	"net/http"
)

func (r *PhotoAsset) delete() {
	body := fmt.Sprintf(`{"operations":[{"operationType":"update","record":{"recordName":"%s","recordType":"%s","recordChangeTag":"%s","fields":{"isDeleted":{"value":1}}}}],"zoneID":{"zoneName":"PrimarySync"},"atomic":true}`,
		r._assetRecord.RecordName,
		r._assetRecord.RecordType,
		r._masterRecord.RecordChangeTag,
	)
	text, err := r.service.icloud.request(&rawReq{
		Method:  http.MethodPost,
		URL:     fmt.Sprintf("%s/records/modify", r.service.serviceEndpoint),
		Querys:  r.service.querys,
		Headers: r.service.icloud.getCommonHeaders(map[string]string{}),
		Body:    body,
	})
	fmt.Println(text, err)
}
