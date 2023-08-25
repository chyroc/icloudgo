package main

import (
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
				Flags:       command.NewDownloadFlag(),
				Action:      command.Download,
			},
			{
				Name:        "upload",
				Aliases:     []string{"u"},
				Description: "upload photos",
				Flags:       command.NewUploadFlag(),
				Action:      command.Upload,
			},
			{
				Name:        "list-db",
				Aliases:     []string{"ld"},
				Description: "list database datas",
				Flags:       command.NewListDBFlag(),
				Action:      command.ListDB,
			},
		},
	}
	if err := app.Run(os.Args); err != nil {
		log.Fatalln(err)
	}
}
