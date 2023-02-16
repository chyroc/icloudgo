package icloudgo

import (
	"github.com/chyroc/icloudgo/internal"
)

func New(option *ClientOption) (*Client, error) {
	return internal.NewClient(option)
}

type (
	TextGetter   func(appleID string) (string, error)
	Client       = internal.Client
	ClientOption = internal.ClientOption
	Error        = internal.Error
	PhotoAlbum   = internal.PhotoAlbum
	PhotoAsset   = internal.PhotoAsset
)

var (
	ErrValidateCodeWrong = internal.ErrValidateCodeWrong
	ErrPhotosIterateEnd  = internal.ErrPhotosIterateEnd
)

func NewError(code string, message string) *Error {
	return internal.NewError(code, message)
}

func IsErrorCode(err error, code string) bool {
	return internal.IsErrorCode(err, code)
}
