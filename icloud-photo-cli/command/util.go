package command

import (
	"fmt"
	"os"
)

func getTextInput(tip, defaultValue string) func(string) (string, error) {
	return func(string2 string) (string, error) {
		if defaultValue != "" {
			return defaultValue, nil
		}
		fmt.Println("Please input", tip)
		var s string
		_, err := fmt.Scanln(&s)
		return s, err
	}
}

func mkdirAll(path string) error {
	if f, _ := os.Stat(path); f == nil {
		if err := os.MkdirAll(path, os.ModePerm); err != nil {
			return err
		}
	}
	return nil
}
