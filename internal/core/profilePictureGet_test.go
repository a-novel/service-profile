package core_test

import (
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/a-novel/service-profile/internal/core"
	coremocks "github.com/a-novel/service-profile/internal/core/mocks"
	"github.com/a-novel/service-profile/internal/dao"
)

func TestProfilePictureGet(t *testing.T) {
	t.Parallel()

	errFoo := errors.New("foo")

	type daoMock struct {
		resp *dao.ProfilePicture
		err  error
	}

	testCases := []struct {
		name string

		request *core.ProfilePictureGetRequest

		daoMock *daoMock

		expect    *core.ProfilePicture
		expectErr error
	}{
		{
			name: "Success",

			request: &core.ProfilePictureGetRequest{ID: profileID},

			daoMock: &daoMock{
				resp: &dao.ProfilePicture{Picture: stillWebP, Type: "image/webp", UpdatedAt: pictureVersion},
			},

			expect: &core.ProfilePicture{Type: "image/webp", Picture: stillWebP, Version: pictureVersion},
		},
		{
			name: "Error/NotFound",

			request: &core.ProfilePictureGetRequest{ID: profileID},

			daoMock: &daoMock{err: dao.ErrProfilePictureSelectNotFound},

			expectErr: core.ErrProfilePictureSelectNotFound,
		},
		{
			name: "Error/Dao",

			request: &core.ProfilePictureGetRequest{ID: profileID},

			daoMock: &daoMock{err: errFoo},

			expectErr: errFoo,
		},
		{
			name: "Error/NoID",

			request: &core.ProfilePictureGetRequest{ID: uuid.Nil},

			expectErr: core.ErrInvalidRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			daoSelect := coremocks.NewMockProfilePictureGetDao(t)

			if testCase.daoMock != nil {
				daoSelect.EXPECT().
					Exec(mock.Anything, &dao.ProfilePictureSelectRequest{ID: testCase.request.ID}).
					Return(testCase.daoMock.resp, testCase.daoMock.err)
			}

			result, err := core.NewProfilePictureGet(daoSelect).Exec(t.Context(), testCase.request)
			require.ErrorIs(t, err, testCase.expectErr)
			require.Equal(t, testCase.expect, result)

			daoSelect.AssertExpectations(t)
		})
	}
}
