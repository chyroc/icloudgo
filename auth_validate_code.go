package icloudgo

import (
	"fmt"
	"net/http"
)

func (r *Client) validate2FACode(code string) error {
	body := map[string]interface{}{"securityCode": map[string]string{"code": code}}

	headers := r.getAuthHeaders(map[string]string{"Accept": "application/json"})
	headers = setIfNotEmpty(headers, "scnt", r.SessionData.Scnt)
	headers = setIfNotEmpty(headers, "X-Apple-ID-Session-Id", r.SessionData.SessionID)

	if _, err := r.Request(&reqParam{
		Method:       http.MethodPost,
		URL:          r.authEndpoint + "/verify/trusteddevice/securitycode",
		Headers:      headers,
		Body:         body,
		ExpectStatus: newSet[int](http.StatusNoContent),
	}); err != nil {
		if IsErrorCode(err, ErrValidateCodeWrong.Code) {
			return ErrValidateCodeWrong
		}
		return fmt.Errorf("validate2FACode failed: %w", err)
	}

	if err := r.trustSession(); err != nil {
		return err
	}

	if r.isRequires2FA() {
		return fmt.Errorf("2FA is still required after validate2FACode")
	}

	return nil
}
