package internal

import (
	"fmt"
	"os"
)

func (r *Client) verify2Fa() error {
	if r.Data == nil || r.Data.DsInfo == nil {
		return fmt.Errorf("not authenticated validate data")
	}

	if r.isRequires2FA() {
		code, err := r.twoFACodeGetter(r.appleID)
		if err != nil {
			return fmt.Errorf("get 2fa code failed, err: %w", err)
		}
		if err := r.validate2FACode(code); err != nil {
			return err
		}

		if !r.Data.HsaTrustedBrowser {
			if err := r.trustSession(); err != nil {
				return err
			}
		}
	} else if r.isRequires2SA() {
		fmt.Printf("Two-step authentication required. Your trusted devices are:\n")
		devices, err := r.trustedDevices()
		if err != nil {
			return err
		}
		for i, device := range devices {
			fmt.Printf("  %d: %s\n", i, device.GetName())
		}

		fmt.Printf("not impl")
		os.Exit(1)
	}
	return nil
}

func (r *Client) isRequires2FA() bool {
	return r.Data.DsInfo.HsaVersion == 2 && (r.Data.HsaChallengeRequired || !r.Data.HsaTrustedBrowser)
}

func (r *Client) isRequires2SA() bool {
	return r.Data.DsInfo.HsaVersion >= 1 && (r.Data.HsaChallengeRequired || !r.Data.HsaTrustedBrowser)
}
