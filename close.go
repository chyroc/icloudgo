package icloudgo

import (
	"encoding/json"
	"os"
)

func (r *Client) Close() error {
	if err := os.WriteFile(r.clientIDPath, []byte(r.clientID), 0o644); err != nil {
		return err
	}

	if bs, _ := json.Marshal(r.sessionData); len(bs) > 0 {
		if err := os.WriteFile(r.sessionDataPath, bs, 0o644); err != nil {
			return err
		}
	}

	return nil
}
