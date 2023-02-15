package icloudgo

import (
	"fmt"
	"net/http"
)

func (r *Client) signIn(password string) error {
	body := map[string]any{
		"accountName": r.appleID,
		"password":    password,
		"rememberMe":  true,
		"trustTokens": []string{},
	}
	if r.sessionData.TrustToken != "" {
		body["trustTokens"] = []string{r.sessionData.TrustToken}
	}

	headers := r.getAuthHeaders(map[string]string{})
	headers = setIfNotEmpty(headers, "scnt", r.sessionData.Scnt)
	headers = setIfNotEmpty(headers, "X-Apple-ID-Session-Id", r.sessionData.SessionID)

	_, err := r.request(&rawReq{
		Method:       http.MethodPost,
		URL:          r.authEndpoint + "/signin",
		Headers:      headers,
		Querys:       map[string]string{"isRememberMeEnabled": "true"},
		Body:         body,
		ExpectStatus: newSet[int](http.StatusOK),
	})
	if err != nil {
		return fmt.Errorf("signin failed: %w", err)
	}

	return r.authWithToken()
}
