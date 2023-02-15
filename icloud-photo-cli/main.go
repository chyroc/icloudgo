package main

import (
	"fmt"
	"log"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/chyroc/icloudgo/icloud-photo-cli/command"
)

func main() {
	app := &cli.App{
		Name:  "icloud-photo-cli",
		Usage: "icloud photo cli",
		Commands: []*cli.Command{
			{
				Name:        "download",
				Aliases:     []string{"d"},
				Description: "download photos",
				Flags: []cli.Flag{
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
					&cli.StringFlag{
						Name:        "output",
						Usage:       "output dir",
						Required:    false,
						DefaultText: "./iCloudPhotos",
						Aliases:     []string{"o"},
						EnvVars:     []string{"ICLOUD_OUTPUT"},
					},
					&cli.StringFlag{
						Name:     "album",
						Usage:    "album name, if not set, download all albums",
						Required: false,
						Aliases:  []string{"a"},
						EnvVars:  []string{"ICLOUD_ALBUM"},
					},
					&cli.Int64Flag{
						Name:     "recent",
						Usage:    "download recent photos, if not set, means all",
						Required: false,
						Aliases:  []string{"r"},
						EnvVars:  []string{"ICLOUD_RECENT"},
					},
					&cli.StringFlag{
						Name:     "duplicate",
						Usage:    "duplicate policy, if not set, means skip",
						Required: false,
						Aliases:  []string{"dup"},
						EnvVars:  []string{"ICLOUD_DUPLICATE"},
						Action: func(context *cli.Context, s string) error {
							if s != "skip" && s != "rename" && s != "overwrite" && s != "" {
								return fmt.Errorf("invalid duplicate policy: %s, should be skip, rename, overwrite", s)
							}
							return nil
						},
					},
				},
				Action: command.Download,
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}
