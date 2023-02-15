package command

import "fmt"

func getTextInput(tip string) func(string) (string, error) {
	return func(string2 string) (string, error) {
		fmt.Println("Please input", tip)
		var s string
		_, err := fmt.Scanln(&s)
		return s, err
	}
}
