package internal

import (
	"fmt"
	"net/http"
)

func (r *Client) authWithCredentialsService(service, password string) error {
	_, err := r.request(&rawReq{
		Method:  http.MethodPost,
		URL:     r.setupEndpoint + "/accountLogin",
		Headers: r.getCommonHeaders(map[string]string{}),
		Body: map[string]string{
			"appName":  service,
			"apple_id": r.appleID,
			"password": password,
		},
	})
	if err != nil {
		return fmt.Errorf("authWithCredentialsService failed, err: %w", err)
	}

	return r.validateToken()
}
