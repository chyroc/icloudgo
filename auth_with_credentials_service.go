package icloudgo

import (
	"fmt"
	"net/http"
)

func (r *Client) authWithCredentialsService(service string) error {
	_, err := r.Request(&reqParam{
		Method:  http.MethodPost,
		URL:     r.setupEndpoint + "/accountLogin",
		Headers: r.getCommonHeaders(map[string]string{}),
		Body: map[string]string{
			"appName":  service,
			"apple_id": r.User.AccountName,
			"password": r.User.Password,
		},
	})
	if err != nil {
		return fmt.Errorf("authWithCredentialsService failed, err: %w", err)
	}

	return r.validateToken()
}
