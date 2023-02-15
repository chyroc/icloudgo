package icloudgo

import (
	"fmt"
	"net/http"
)

// session trust to avoid user log in going forward
func (r *Client) trustSession() error {
	headers := r.getAuthHeaders(map[string]string{})
	headers = setIfNotEmpty(headers, "scnt", r.sessionData.Scnt)
	headers = setIfNotEmpty(headers, "X-Apple-ID-Session-Id", r.sessionData.SessionID)

	_, err := r.request(&rawReq{
		Method:       http.MethodGet,
		URL:          r.authEndpoint + "/2sv/trust",
		Headers:      headers,
		ExpectStatus: newSet[int](http.StatusNoContent),
	})
	if err != nil {
		return fmt.Errorf("trustSession failed: %w", err)
	}

	return r.authWithToken()
}
