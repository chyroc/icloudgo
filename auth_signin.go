package icloudgo

import (
	"fmt"
	"net/http"
)

func (r *Client) signIn() error {
	body := map[string]any{
		"accountName": r.User.AccountName,
		"password":    r.User.Password,
		"rememberMe":  true,
		"trustTokens": []string{},
	}
	if r.SessionData.TrustToken != "" {
		body["trustTokens"] = []string{r.SessionData.TrustToken}
	}

	headers := r.getAuthHeaders(map[string]string{})
	headers = setIfNotEmpty(headers, "scnt", r.SessionData.Scnt)
	headers = setIfNotEmpty(headers, "X-Apple-ID-Session-Id", r.SessionData.SessionID)

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
