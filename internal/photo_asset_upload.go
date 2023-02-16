package internal

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func (r *PhotoService) Upload(filename string, file io.Reader) (bool, error) {
	webServiceURL, err := r.icloud.getWebServiceURL("uploadimagews")
	if err != nil {
		return false, err
	}

	resp := new(uploadPhotoResp)
	body, err := r.icloud.request(&rawReq{
		Method:  http.MethodPost,
		URL:     webServiceURL + "/upload",
		Headers: r.icloud.getCommonHeaders(map[string]string{"Content-Type": "text/plain"}),
		Querys:  map[string]string{"filename": filename},
		Body:    file,
	})
	if err != nil {
		return false, fmt.Errorf("upload %s failed: %w", filename, err)
	}
	if err := json.Unmarshal([]byte(body), resp); err != nil {
		return false, fmt.Errorf("upload %s unmarshal failed: %w", filename, err)
	}
	return resp.IsDuplicate, nil
}

type uploadPhotoResp struct {
	IsDuplicate bool `json:"isDuplicate"`
}
