package internal

import (
	"fmt"
	"strings"
)

func (r *Client) Authenticate(forceRefresh bool, service *string) (finalErr error) {
	defer func() {
		if finalErr == nil {
			r.flush()
		}
	}()

	var errs []string
	if r.sessionData.SessionToken != "" && !forceRefresh {
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
			fmt.Printf("Authenticating as %s for %s\n", r.appleID, *service)
			if err := r.authWithCredentialsService(*service, r.password); err != nil {
				errs = append(errs, err.Error())
				fmt.Printf("Could not log into service. Attempting brand new login.\n")
			} else {
				return nil
			}
		}
	}

	// default, login to icloud.com[.cn]
	{
		fmt.Printf("Authenticating as %s\n", r.appleID)
		err := r.signIn(r.password)
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
