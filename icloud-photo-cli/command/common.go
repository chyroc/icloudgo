package command

import (
	"fmt"

	"github.com/urfave/cli/v2"
)

var commonFlag = []cli.Flag{
	&cli.StringFlag{
		Name:     "username",
		Usage:    "apple id username",
		Required: true,
		Aliases:  []string{"u"},
		EnvVars:  []string{"ICLOUD_USERNAME"},
	},
	&cli.StringFlag{
		Name:     "password",
		Usage:    "apple id password",
		Required: false,
		Aliases:  []string{"p"},
		EnvVars:  []string{"ICLOUD_PASSWORD"},
	},
	&cli.StringFlag{
		Name:     "cookie-dir",
		Usage:    "cookie dir",
		Required: false,
		Aliases:  []string{"c"},
		EnvVars:  []string{"ICLOUD_COOKIE_DIR"},
	},
	&cli.StringFlag{
		Name:        "domain",
		Usage:       "icloud domain(com,cn)",
		Required:    false,
		DefaultText: "com",
		Aliases:     []string{"d"},
		EnvVars:     []string{"ICLOUD_DOMAIN"},
		Action: func(context *cli.Context, s string) error {
			if s != "com" && s != "cn" && s != "" {
				return fmt.Errorf("domain must be com or cn")
			}
			return nil
		},
	},
}
