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

func TestProfilePictureDelete(t *testing.T) {
	t.Parallel()

	errFoo := errors.New("foo")

	type daoMock struct {
		resp *dao.Profile
		err  error
	}

	testCases := []struct {
		name string

		request *core.ProfilePictureDeleteRequest

		daoMock *daoMock

		expect    *core.Profile
		expectErr error
	}{
		{
			name: "Success",

			request: &core.ProfilePictureDeleteRequest{UserID: userID},

			daoMock: &daoMock{resp: daoProfile},

			expect: coreProfile,
		},
		{
			name: "Error/NotFound",

			request: &core.ProfilePictureDeleteRequest{UserID: userID},

			daoMock: &daoMock{err: dao.ErrProfilePictureDeleteNotFound},

			expectErr: core.ErrProfilePictureDeleteNotFound,
		},
		{
			name: "Error/Dao",

			request: &core.ProfilePictureDeleteRequest{UserID: userID},

			daoMock: &daoMock{err: errFoo},

			expectErr: errFoo,
		},
		{
			name: "Error/NoUser",

			request: &core.ProfilePictureDeleteRequest{UserID: uuid.Nil},

			expectErr: core.ErrInvalidRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			daoDelete := coremocks.NewMockProfilePictureDeleteDao(t)

			if testCase.daoMock != nil {
				daoDelete.EXPECT().
					Exec(mock.Anything, mock.MatchedBy(func(request *dao.ProfilePictureDeleteRequest) bool {
						return request.UserID == testCase.request.UserID && !request.Now.IsZero()
					})).
					Return(testCase.daoMock.resp, testCase.daoMock.err)
			}

			result, err := core.NewProfilePictureDelete(daoDelete).Exec(t.Context(), testCase.request)
			require.ErrorIs(t, err, testCase.expectErr)
			require.Equal(t, testCase.expect, result)

			daoDelete.AssertExpectations(t)
		})
	}
}
