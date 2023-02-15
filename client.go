package icloudgo

import (
	"github.com/chyroc/icloudgo/internal"
)

func New(option *ClientOption) (*Client, error) {
	return internal.NewClient(option)
}

type TextGetter func(appleID string) (string, error)
type Client = internal.Client
type ClientOption = internal.ClientOption
type Error = internal.Error
type PhotoAlbum = internal.PhotoAlbum
type PhotoAsset = internal.PhotoAsset

var ErrValidateCodeWrong = internal.ErrValidateCodeWrong
var ErrPhotosIterateEnd = internal.ErrPhotosIterateEnd

func NewError(code string, message string) *Error {
	return internal.NewError(code, message)
}

func IsErrorCode(err error, code string) bool {
	return internal.IsErrorCode(err, code)
}
