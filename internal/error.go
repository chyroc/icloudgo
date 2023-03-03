package internal

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
)

var (
	ErrValidateCodeWrong = NewError("-21669", "validate code wrong")
	ErrPhotosIterateEnd  = NewError("photos_iterate_end", "photos iterate end")
	ErrResourceGone      = NewHttpError(410, "resource gone")
)

type Error struct {
	HttpStatus int
	Code       string
	Message    string
}

func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func NewHttpError(httpStatus int, body string) *Error {
	return &Error{
		HttpStatus: httpStatus,
		Code:       fmt.Sprintf("http_%d", httpStatus),
		Message:    body,
	}
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

func mayErr(respText []byte) error {
	for _, errResp := range []interface{ err() error }{new(errResp1), new(errResp2), new(errResp3), new(errResp4)} {
		if err := json.Unmarshal(respText, errResp); err == nil && errResp.err() != nil {
			return errResp.err()
		}
	}
	return nil
}

// {"service_errors":[{"code":"-21669","title":"Incorrect verification code.","message":"Please try again."}],"hasError":true}
type errResp1 struct {
	ServiceErrors []struct {
		Code    string `json:"code"`
		Title   string `json:"title"`
		Message string `json:"message"`
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

type errResp2 struct {
	Reason string `json:"reason"`
	Error  string `json:"error"`
}

func (r errResp2) err() error {
	if r.Error != "" {
		if r.Reason == "" {
			return NewError("-2", r.Error)
		}
		return NewError("-2", r.Error+" "+r.Reason)
	}
	return nil
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

// {"errors":[{"errorCode":"CLOUD_DB_FAILURE"}],"requestUUID":"fb28547f-3785-4a4f-903c-13b51aa236a9"}
type errResp4 struct {
	Errors []struct {
		ErrorCode string `json:"errorCode"`
	} `json:"errors"`
}

func (r errResp4) err() error {
	if len(r.Errors) == 0 {
		return nil
	}
	return NewError("-2", r.Errors[0].ErrorCode)
}
