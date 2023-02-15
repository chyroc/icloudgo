package icloudgo

import (
	"fmt"
	"strings"
)

func (r *Client) Authenticate(forceRefresh bool, service *string) error {
	var errs []string
	if r.SessionData.SessionToken != "" && !forceRefresh {
		fmt.Printf("Checking session token validity")
		if err := r.validateToken(); err == nil {
			return nil
		} else {
			errs = append(errs, err.Error())
			fmt.Printf("Invalid session token. Attempting brand new login.\n")
		}
	}

	if service != nil {
		if r.Data != nil && len(r.Data.Apps) > 0 && r.Data.Apps[*service] != nil && r.Data.Apps[*service].CanLaunchWithOneFactor {
			fmt.Printf("Authenticating as %s for %s\n", r.User.AccountName, *service)
			err := r.authWithCredentialsService(*service)
			if err == nil {
				return nil
			}
			errs = append(errs, err.Error())
			fmt.Printf("Could not log into service. Attempting brand new login.\n")
		}
	}

	// default, login to icloud.com[.cn]
	{
		fmt.Printf("Authenticating as %s\n", r.User.AccountName)
		err := r.signIn()
		if err == nil {
			err = r.verify2Fa()
			if err == nil {
				return nil
			}
		}
		// self._webservices = self.data["webservices"]
		errs = append(errs, err.Error())
		fmt.Printf("Login failed\n")
	}

	return fmt.Errorf("login failed: %s", strings.Join(errs, "; "))
}
