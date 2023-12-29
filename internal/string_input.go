package internal

import "fmt"

type StdinTextGetter struct {
	Tip          string
	DefaultValue string
}

func (r *StdinTextGetter) GetText(tip string) (string, error) {
	if r.DefaultValue != "" {
		return r.DefaultValue, nil
	}
	if tip == "" {
		tip = r.Tip
	}
	fmt.Println("Please input", tip)
	var s string
	_, err := fmt.Scanln(&s)
	return s, err
}
