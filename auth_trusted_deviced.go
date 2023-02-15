package icloudgo

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// Returns devices trusted for two-step authentication.
func (r *Client) trustedDevices() ([]*device, error) {
	text, err := r.Request(&reqParam{
		Method:  http.MethodPost,
		URL:     r.setupEndpoint + "/listDevices",
		Headers: r.getCommonHeaders(map[string]string{}),
	})
	if err != nil {
		return nil, fmt.Errorf("listDevices failed, err: %w", err)
	}
	res := new(trustedDevicesResp)
	if err = json.Unmarshal([]byte(text), res); err != nil {
		return nil, fmt.Errorf("listDevices unmarshal failed, text: %s", text)
	}
	return res.Devices, nil
}

type trustedDevicesResp struct {
	Devices []*device `json:"devices"`
}

type device struct {
	DeviceName  string `json:"deviceName"`
	PhoneNumber string `json:"phoneNumber"`
}

func (r *device) GetName() string {
	if r.DeviceName != "" {
		return r.DeviceName
	}
	return fmt.Sprintf("SMS to %s", r.PhoneNumber)
}
