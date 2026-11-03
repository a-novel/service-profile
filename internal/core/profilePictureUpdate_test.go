package core_test

import (
	"bytes"
	_ "embed"
	"errors"
	"image/gif"
	"image/png"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/a-novel/service-profile/internal/core"
	coremocks "github.com/a-novel/service-profile/internal/core/mocks"
	"github.com/a-novel/service-profile/internal/dao"
)

// The WebP fixtures are 16 px squares that libwebp 1.6 encoded: still.webp with cwebp -q 90, alpha.webp
// with cwebp -q 90 -exact from a translucent picture, and animated.webp with img2webp -lossy -q 90 from
// two frames.
var (
	//go:embed testdata/still.webp
	stillWebP []byte
	//go:embed testdata/alpha.webp
	alphaWebP []byte
	//go:embed testdata/animated.webp
	animatedWebP []byte
)

func TestProfilePictureUpdate(t *testing.T) {
	t.Parallel()

	errFoo := errors.New("foo")

	squareJPEG := newJPEG(64, 64)

	var pngPicture, gifPicture bytes.Buffer

	require.NoError(t, png.Encode(&pngPicture, newImage(64, 64)))
	require.NoError(t, gif.Encode(&gifPicture, newImage(64, 64), nil))

	type daoMock struct {
		pictureType string

		resp *dao.Profile
		err  error
	}

	testCases := []struct {
		name string

		request *core.ProfilePictureUpdateRequest

		daoMock *daoMock

		expect    *core.Profile
		expectErr error
	}{
		{
			name: "Success/WebP",

			request: &core.ProfilePictureUpdateRequest{UserID: userID, Picture: stillWebP},

			daoMock: &daoMock{pictureType: "image/webp", resp: daoProfile},

			expect: coreProfile,
		},
		{
			name: "Success/WebPAlpha",

			request: &core.ProfilePictureUpdateRequest{UserID: userID, Picture: alphaWebP},

			daoMock: &daoMock{pictureType: "image/webp", resp: daoProfile},

			expect: coreProfile,
		},
		{
			name: "Success/JPEG",

			request: &core.ProfilePictureUpdateRequest{UserID: userID, Picture: squareJPEG},

			daoMock: &daoMock{pictureType: "image/jpeg", resp: daoProfile},

			expect: coreProfile,
		},
		{
			name: "Success/SideMax",

			request: &core.ProfilePictureUpdateRequest{UserID: userID, Picture: newJPEG(512, 512)},

			daoMock: &daoMock{pictureType: "image/jpeg", resp: daoProfile},

			expect: coreProfile,
		},
		{
			name: "Success/OnePixel",

			request: &core.ProfilePictureUpdateRequest{UserID: userID, Picture: newJPEG(1, 1)},

			daoMock: &daoMock{pictureType: "image/jpeg", resp: daoProfile},

			expect: coreProfile,
		},
		{
			name: "Error/PNG",

			request: &core.ProfilePictureUpdateRequest{UserID: userID, Picture: pngPicture.Bytes()},

			expectErr: core.ErrPictureFormat,
		},
		{
			name: "Error/GIF",

			request: &core.ProfilePictureUpdateRequest{UserID: userID, Picture: gifPicture.Bytes()},

			expectErr: core.ErrPictureFormat,
		},
		{
			name: "Error/AnimatedWebP",

			request: &core.ProfilePictureUpdateRequest{UserID: userID, Picture: animatedWebP},

			expectErr: core.ErrPictureFormat,
		},
		{
			name: "Error/NotAnImage",

			request: &core.ProfilePictureUpdateRequest{UserID: userID, Picture: []byte("not an image")},

			expectErr: core.ErrPictureFormat,
		},
		{
			name: "Error/TruncatedJPEG",

			request: &core.ProfilePictureUpdateRequest{UserID: userID, Picture: squareJPEG[:len(squareJPEG)*3/4]},

			expectErr: core.ErrPictureFormat,
		},
		{
			name: "Error/TruncatedWebP",

			request: &core.ProfilePictureUpdateRequest{UserID: userID, Picture: stillWebP[:len(stillWebP)-8]},

			expectErr: core.ErrPictureFormat,
		},
		{
			name: "Error/NotSquare",

			request: &core.ProfilePictureUpdateRequest{UserID: userID, Picture: newJPEG(64, 32)},

			expectErr: core.ErrPictureRatio,
		},
		{
			name: "Error/TooLarge",

			request: &core.ProfilePictureUpdateRequest{UserID: userID, Picture: newJPEG(513, 513)},

			expectErr: core.ErrPictureDimensions,
		},
		{
			name: "Error/Empty",

			request: &core.ProfilePictureUpdateRequest{UserID: userID},

			expectErr: core.ErrInvalidRequest,
		},
		{
			name: "Error/NoUser",

			request: &core.ProfilePictureUpdateRequest{Picture: stillWebP},

			expectErr: core.ErrInvalidRequest,
		},
		{
			name: "Error/Dao",

			request: &core.ProfilePictureUpdateRequest{UserID: userID, Picture: stillWebP},

			daoMock: &daoMock{pictureType: "image/webp", err: errFoo},

			expectErr: errFoo,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			daoUpsert := coremocks.NewMockProfilePictureUpdateDao(t)

			if testCase.daoMock != nil {
				daoUpsert.EXPECT().
					Exec(mock.Anything, mock.MatchedBy(func(request *dao.ProfilePictureUpsertRequest) bool {
						return request.ID != uuid.Nil &&
							request.UserID == testCase.request.UserID &&
							bytes.Equal(request.Picture, testCase.request.Picture) &&
							request.Type == testCase.daoMock.pictureType &&
							!request.Now.IsZero()
					})).
					Return(testCase.daoMock.resp, testCase.daoMock.err)
			}

			result, err := core.NewProfilePictureUpdate(daoUpsert).Exec(t.Context(), testCase.request)
			require.ErrorIs(t, err, testCase.expectErr)
			require.Equal(t, testCase.expect, result)

			daoUpsert.AssertExpectations(t)
		})
	}
}
