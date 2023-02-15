package icloudgo

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var ErrValidateCodeWrong = NewError("-21669", "validate code wrong")
var ErrPhotosIterateEnd = NewError("photos_iterate_end", "photos iterate end")

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewError(code string, message string) *Error {
	return &Error{
		Code:    code,
		Message: message,
	}
}

func IsErrorCode(err error, code string) bool {
	if err == nil {
		return false
	}
	if e, ok := err.(*Error); ok {
		return e.Code == code
	}
	return IsErrorCode(errors.Unwrap(err), code)
}

type errResp1 struct {
	ServiceErrors []struct {
		Code              string `json:"code"`
		Title             string `json:"title"`
		Message           string `json:"message"`
		SuppressDismissal bool   `json:"suppressDismissal"`
	} `json:"service_errors"`
	HasError bool `json:"hasError"`
}

func (r errResp1) err() error {
	for _, v := range r.ServiceErrors {
		if v.Code == "" || v.Code == "0" {
			continue
		}
		text1 := strings.Trim(v.Title, ".")
		text2 := strings.Trim(v.Message, ".")

		if strings.ToLower(text2) != strings.ToLower(text1) {
			text1 = text1 + ", " + text2
		}

		return NewError(v.Code, text1)
	}
	if r.HasError {
		bs, _ := json.Marshal(r)
		return NewError("1", "unknown error: "+string(bs))
	}
	return nil
}

func mayErr1(respText []byte) error {
	err := new(errResp1)
	_ = json.Unmarshal(respText, err)
	return err.err()
}

type errResp2 struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
}

func (r errResp2) err() error {
	if r.Success || r.Error == "" {
		return nil
	}
	return NewError("-2", r.Error)
}

func mayErr2(respText []byte) error {
	err := new(errResp2)
	_ = json.Unmarshal(respText, err)
	return err.err()
}

type errResp3 struct {
	Reason string `json:"reason"`
	Error  int    `json:"error"`
}

func (r errResp3) err() error {
	if r.Reason == "" {
		return nil
	}
	return NewError(fmt.Sprintf("%d", r.Error), r.Reason)
}

func mayErr3(respText []byte) error {
	err := new(errResp3)
	_ = json.Unmarshal(respText, err)
	return err.err()
}

type errResp4 struct {
	Uuid            string `json:"uuid"`
	ServerErrorCode string `json:"serverErrorCode"`
	Reason          string `json:"reason"`
	ErrorClass      string `json:"errorClass"`
	Error           string `json:"error"`
}

func (r errResp4) err() error {
	if r.Reason == "" {
		return nil
	}
	return NewError("-2", r.Error+" "+r.Reason)
}

func mayErr4(respText []byte) error {
	err := new(errResp4)
	_ = json.Unmarshal(respText, err)
	return err.err()
}
