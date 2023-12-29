package internal

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/chyroc/gorequests"
	uuid "github.com/satori/go.uuid"
)

type TextGetter interface {
	GetText(tip string) (string, error)
}

type Client struct {
	// param
	appleID         string
	password        string
	twoFACodeGetter TextGetter

	// storage
	cookieDir       string
	cookiePath      string
	clientIDPath    string
	sessionDataPath string

	// user data
	clientID    string
	sessionData *SessionData
	Data        *ValidateData
	httpCli     *gorequests.Session

	// server
	setupEndpoint string
	homeEndpoint  string
	authEndpoint  string

	// service
	photo *PhotoService
	drive *DriveService
}

type ClientOption struct {
	AppID           string
	Password        string
	CookieDir       string
	TwoFACodeGetter TextGetter
	Domain          string // com,cn
}

func NewClient(option *ClientOption) (*Client, error) {
	return newClient(option)
}

func newClient(option *ClientOption) (*Client, error) {
	cli := &Client{
		twoFACodeGetter: option.TwoFACodeGetter,
	}
	var err error

	// domain
	if option.Domain == "cn" {
		cli.setupEndpoint = "https://setup.icloud.com.cn/setup/ws/1"
		cli.homeEndpoint = "https://www.icloud.com.cn"
		cli.authEndpoint = "https://idmsa.apple.com/appleauth/auth"
	} else if option.Domain == "com" {
		cli.setupEndpoint = "https://setup.icloud.com/setup/ws/1"
		cli.homeEndpoint = "https://www.icloud.com"
		cli.authEndpoint = "https://idmsa.apple.com/appleauth/auth"
	} else {
		return nil, fmt.Errorf("invalid domain: %s", option.Domain)
	}

	// storage
	{
		cli.cookieDir, err = cli.initCookieDir(option.CookieDir)
		if err != nil {
			return nil, err
		}
		cli.cookiePath = cli.ConfigPath("cookie.json")
		cli.clientIDPath = cli.ConfigPath("client_id.txt")
		cli.sessionDataPath = cli.ConfigPath("session_data.json")
	}

	// load from file
	{
		// client id
		if clientIDCached := readFile(cli.clientIDPath); len(clientIDCached) > 0 {
			cli.clientID = string(clientIDCached)
		} else {
			cli.clientID = "auth-" + uuid.NewV1().String()
		}

		// session data
		cli.sessionData = new(SessionData)
		if sessionDataCached := readFile(cli.sessionDataPath); len(sessionDataCached) > 0 {
			_ = json.Unmarshal(sessionDataCached, cli.sessionData)
		}

		// data
		cli.Data = new(ValidateData)
	}

	cli.appleID = option.AppID

	cli.httpCli = gorequests.NewSession(
		fmt.Sprintf("%s/session.json", cli.cookieDir),
		gorequests.WithLogger(gorequests.NewDiscardLogger()),
	)

	return cli, nil
}

func readFile(path string) []byte {
	bs, _ := os.ReadFile(path)
	return bs
}

const (
	serviceDrive       = "drivews"
	serviceDatabase    = "ckdatabasews"
	serviceUploadImage = "uploadimagews"
)

func (r *Client) getWebServiceURL(key string) (string, error) {
	if _, ok := r.Data.Webservices[key]; !ok {
		return "", fmt.Errorf("webservice not available: %s", key)
	}
	return r.Data.Webservices[key].URL, nil
}
