package icloudgo

import (
	"encoding/json"
	"fmt"
	"net/http"
)

func (r *Client) validateToken() error {
	fmt.Printf("Checking session token validity\n")

	text, err := r.Request(&reqParam{
		Method:  http.MethodPost,
		URL:     r.setupEndpoint + "/validate",
		Headers: r.getCommonHeaders(map[string]string{}),
	})
	if err != nil {
		return fmt.Errorf("validateToken failed, err: %w", err)
	}

	res := new(ValidateData)
	if err = json.Unmarshal([]byte(text), res); err != nil {
		return fmt.Errorf("validateToken unmarshal failed, err: %w, text: %s", err, text)
	}
	r.Data = res

	return nil
}
