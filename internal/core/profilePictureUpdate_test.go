package core_test

import (
	"bytes"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/a-novel-kit/golib/picture"

	"github.com/a-novel/service-profile/internal/core"
	coremocks "github.com/a-novel/service-profile/internal/core/mocks"
	"github.com/a-novel/service-profile/internal/dao"
)

func TestProfilePictureUpdate(t *testing.T) {
	t.Parallel()

	errFoo := errors.New("foo")

	squareJPEG := newJPEG(64, 64)

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
			name: "Error/NotAnImage",

			request: &core.ProfilePictureUpdateRequest{UserID: userID, Picture: []byte("not an image")},

			expectErr: picture.ErrFormat,
		},
		{
			name: "Error/NotSquare",

			request: &core.ProfilePictureUpdateRequest{UserID: userID, Picture: newJPEG(64, 32)},

			expectErr: picture.ErrRatio,
		},
		{
			name: "Error/TooLarge",

			request: &core.ProfilePictureUpdateRequest{UserID: userID, Picture: newJPEG(513, 513)},

			expectErr: picture.ErrDimensions,
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
