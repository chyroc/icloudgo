package command

import (
	"os"
)

func mkdirAll(path string) error {
	if f, _ := os.Stat(path); f == nil {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}
