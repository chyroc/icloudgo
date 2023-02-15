package icloudgo

import (
	"fmt"
	"strings"
)

func (r *Client) Authenticate(forceRefresh bool, service *string) error {
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

	password, err := getPassword(r.appleID, r.passwordGetter)
	if err != nil {
		return err
	}

	if service != nil {
		if r.Data != nil && len(r.Data.Apps) > 0 && r.Data.Apps[*service] != nil && r.Data.Apps[*service].CanLaunchWithOneFactor {

			fmt.Printf("Authenticating as %s for %s\n", r.appleID, *service)
			if err := r.authWithCredentialsService(*service, password); err != nil {
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
		err := r.signIn(password)
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

func getPassword(appleID string, passwordGetter TextGetter) (string, error) {
	// if password != "" {
	// 	return password, nil
	// }
	if passwordGetter == nil {
		return "", fmt.Errorf("password getter is empty")
	}
	password, err := passwordGetter(appleID)
	if err != nil {
		return "", fmt.Errorf("password get failed, err: %w", err)
	}
	return password, nil
}
