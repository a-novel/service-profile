package core_test

import (
	"bytes"
	_ "embed"
	"image"
	"image/color"
	"image/jpeg"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"

	"github.com/a-novel/service-profile/internal/core"
	"github.com/a-novel/service-profile/internal/dao"
)

var (
	userID    = uuid.MustParse("00000000-0000-0000-0000-0000000000a1")
	profileID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

	pictureVersion = time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC)

	// daoProfile is the profile the DAO mocks return.
	daoProfile = &dao.Profile{ID: profileID, Username: lo.ToPtr("Jean Paul"), PictureUpdatedAt: &pictureVersion}
	// coreProfile is daoProfile as the core returns it.
	coreProfile = &core.Profile{ID: profileID, Username: lo.ToPtr("Jean Paul"), PictureVersion: &pictureVersion}
)

// stillWebP is a 16 px square that libwebp 1.6 encoded with cwebp -q 90.
//
//go:embed testdata/still.webp
var stillWebP []byte

// newImage returns an image of the given size, shaded so its encoding holds more than a header.
func newImage(width, height int) image.Image {
	img := image.NewGray(image.Rect(0, 0, width, height))

	for x := range width {
		for y := range height {
			img.SetGray(x, y, color.Gray{Y: uint8((x*7 + y*13) % 256)})
		}
	}

	return img
}

// newJPEG encodes an image of the given size as a JPEG.
func newJPEG(width, height int) []byte {
	var buf bytes.Buffer

	err := jpeg.Encode(&buf, newImage(width, height), nil)
	if err != nil {
		panic(err)
	}

	return buf.Bytes()
}
