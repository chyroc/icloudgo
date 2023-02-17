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
	PhotoService = internal.PhotoService
)

var (
	ErrValidateCodeWrong = internal.ErrValidateCodeWrong
	ErrPhotosIterateEnd  = internal.ErrPhotosIterateEnd
)

const (
	AlbumNameAll             = internal.AlbumNameAll
	AlbumNameTimeLapse       = internal.AlbumNameTimeLapse
	AlbumNameVideos          = internal.AlbumNameVideos
	AlbumNameSloMo           = internal.AlbumNameSloMo
	AlbumNameBursts          = internal.AlbumNameBursts
	AlbumNameFavorites       = internal.AlbumNameFavorites
	AlbumNamePanoramas       = internal.AlbumNamePanoramas
	AlbumNameScreenshots     = internal.AlbumNameScreenshots
	AlbumNameLive            = internal.AlbumNameLive
	AlbumNameRecentlyDeleted = internal.AlbumNameRecentlyDeleted
	AlbumNameHidden          = internal.AlbumNameHidden
)

type PhotoVersion = internal.PhotoVersion

const (
	PhotoVersionOriginal = internal.PhotoVersionOriginal
	PhotoVersionMedium   = internal.PhotoVersionMedium
	PhotoVersionThumb    = internal.PhotoVersionThumb
)
