package icloudgo

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/chyroc/gorequests"
	uuid "github.com/satori/go.uuid"
)

type TextGetter func() (string, error)

type Client struct {
	// storage
	cookieDir       string
	cookiePath      string
	clientIDPath    string
	sessionDataPath string

	// user data
	User        *User
	ClientID    string
	WithFamily  bool
	SessionData *SessionData
	Data        *ValidateData

	// server
	httpCli       *gorequests.Session
	setupEndpoint string
	homeEndpoint  string
	authEndpoint  string

	// iface
	TwoFACodeGetter TextGetter
	PasswordGetter  TextGetter

	// service
	photo *PhotoService
}

type ClientOption struct {
	AppID           string
	CookieDir       string
	PasswordGetter  TextGetter
	TwoFACodeGetter TextGetter
	Domain          string // com,cn
}

func New(option *ClientOption) (*Client, error) {
	return newClient(option)
}

func newClient(option *ClientOption) (*Client, error) {
	cli := &Client{
		TwoFACodeGetter: option.TwoFACodeGetter,
		PasswordGetter:  option.PasswordGetter,
	}

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
		// cookie dir
		if option.CookieDir == "" {
			option.CookieDir = filepath.Join(os.TempDir(), "icloudgo")
		}

		cli.cookieDir = option.CookieDir
		if f, _ := os.Stat(cli.cookieDir); f == nil {
			if err := os.MkdirAll(cli.cookieDir, 0700); err != nil {
				return nil, fmt.Errorf("create cookie dir failed, err: %w", err)
			}
		}

		cli.cookiePath = filepath.Join(cli.cookieDir, "cookie.json")
		cli.clientIDPath = filepath.Join(cli.cookieDir, "client_id.txt")
		cli.sessionDataPath = filepath.Join(cli.cookieDir, "session_data.json")
	}

	// load from file
	{
		// client id
		if clientIDCached := readFile(cli.clientIDPath); len(clientIDCached) > 0 {
			cli.ClientID = string(clientIDCached)
		} else {
			cli.ClientID = "auth-" + uuid.NewV1().String()
		}

		// session data
		cli.SessionData = new(SessionData)
		if sessionDataCached := readFile(cli.sessionDataPath); len(sessionDataCached) > 0 {
			_ = json.Unmarshal(sessionDataCached, cli.SessionData)
		}

		// data
		cli.Data = new(ValidateData)
	}

	cli.User = &User{AccountName: option.AppID, Password: ""}

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
