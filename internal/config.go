package internal

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func (r *Client) ConfigPath(configName string) string {
	return filepath.Join(r.cookieDir, configName)
}

func (r *Client) LoadConfig(configName string) ([]byte, error) {
	bs, err := os.ReadFile(r.ConfigPath(configName))
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, nil
		}
		return nil, err
	}
	return bs, nil
}

func (r *Client) SaveConfig(configName string, data []byte) error {
	return os.WriteFile(r.ConfigPath(configName), data, 0o644)
}

func (r *Client) initCookieDir(defaultValue string) (string, error) {
	if defaultValue == "" {
		defaultValue = filepath.Join(os.TempDir(), "icloudgo")
	}

	if f, _ := os.Stat(defaultValue); f == nil {
		if err := os.MkdirAll(defaultValue, 0o700); err != nil {
			return "", fmt.Errorf("create cookie dir failed, err: %w", err)
		}
	}
	return defaultValue, nil
}
