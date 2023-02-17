package command

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/urfave/cli/v2"

	"github.com/chyroc/icloudgo"
)

func NewUploadFlag() []cli.Flag {
	var res []cli.Flag
	res = append(res, commonFlag...)
	res = append(res,
		&cli.StringFlag{
			Name:     "file",
			Usage:    "file path",
			Required: true,
			Aliases:  []string{"f"},
			EnvVars:  []string{"ICLOUD_FILE"},
		},
	)
	return res
}

func Upload(c *cli.Context) error {
	username := c.String("username")
	password := c.String("password")
	cookieDir := c.String("cookie-dir")
	domain := c.String("domain")
	file := c.String("file")

	cli, err := icloudgo.New(&icloudgo.ClientOption{
		AppID:           username,
		CookieDir:       cookieDir,
		PasswordGetter:  getTextInput("apple id password", password),
		TwoFACodeGetter: getTextInput("2fa code", ""),
		Domain:          domain,
	})
	if err != nil {
		return err
	}

	defer cli.Close()

	if err := cli.Authenticate(false, nil); err != nil {
		return err
	}

	photoCli, err := cli.PhotoCli()
	if err != nil {
		return err
	}

	basename := filepath.Base(file)
	f, err := os.Open(file)
	if err != nil {
		return err
	}
	defer f.Close()

	isDuplicate, err := photoCli.Upload(basename, f)
	if err != nil {
		return err
	}
	if isDuplicate {
		fmt.Printf("file %s is duplicate\n", basename)
	}
	return nil
}
