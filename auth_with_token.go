package icloudgo

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// auth using session token
func (r *Client) authWithToken() error {
	text, err := r.Request(&reqParam{
		Method:  http.MethodPost,
		URL:     r.setupEndpoint + "/accountLogin",
		Headers: r.getCommonHeaders(map[string]string{}),
		Body: map[string]any{
			"accountCountryCode": r.SessionData.AccountCountry,
			"dsWebAuthToken":     r.SessionData.SessionToken,
			"extended_login":     true,
			"trustToken":         r.SessionData.TrustToken,
		},
		ExpectStatus: newSet[int](200),
	})
	if err != nil {
		return fmt.Errorf("authWithToken failed, err: %w", err)
	}

	data := new(ValidateData)
	if err = json.Unmarshal([]byte(text), data); err != nil {
		return fmt.Errorf("authWithToken unmarshal failed, text: %s", text)
	}
	r.Data = data
	return nil
}
