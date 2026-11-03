package core_test

import (
	"errors"
	"reflect"
	"testing"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/a-novel/service-profile/internal/core"
	coremocks "github.com/a-novel/service-profile/internal/core/mocks"
	"github.com/a-novel/service-profile/internal/dao"
)

func TestProfileUsernameUpdate(t *testing.T) {
	t.Parallel()

	errFoo := errors.New("foo")

	type daoMock struct {
		username *string

		resp *dao.Profile
		err  error
	}

	testCases := []struct {
		name string

		request *core.ProfileUsernameUpdateRequest

		daoMock *daoMock

		expect    *core.Profile
		expectErr error
	}{
		{
			name: "Success",

			request: &core.ProfileUsernameUpdateRequest{UserID: userID, Username: "  Jean Paul  "},

			daoMock: &daoMock{username: lo.ToPtr("Jean Paul"), resp: daoProfile},

			expect: coreProfile,
		},
		{
			// A name of spaces trims to nothing, which clears the username.
			name: "Success/Clear",

			request: &core.ProfileUsernameUpdateRequest{UserID: userID, Username: "   "},

			daoMock: &daoMock{resp: daoProfile},

			expect: coreProfile,
		},
		{
			name: "Error/Username",

			request: &core.ProfileUsernameUpdateRequest{UserID: userID, Username: "Jean\tPaul"},

			expectErr: core.ErrInvalidRequest,
		},
		{
			name: "Error/NoUser",

			request: &core.ProfileUsernameUpdateRequest{UserID: uuid.Nil, Username: "Jean Paul"},

			expectErr: core.ErrInvalidRequest,
		},
		{
			name: "Error/Dao",

			request: &core.ProfileUsernameUpdateRequest{UserID: userID, Username: "Jean Paul"},

			daoMock: &daoMock{username: lo.ToPtr("Jean Paul"), err: errFoo},

			expectErr: errFoo,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			daoUpsert := coremocks.NewMockProfileUsernameUpdateDao(t)

			if testCase.daoMock != nil {
				daoUpsert.EXPECT().
					Exec(mock.Anything, mock.MatchedBy(func(request *dao.ProfileUsernameUpsertRequest) bool {
						return request.ID != uuid.Nil &&
							request.UserID == testCase.request.UserID &&
							reflect.DeepEqual(request.Username, testCase.daoMock.username) &&
							!request.Now.IsZero()
					})).
					Return(testCase.daoMock.resp, testCase.daoMock.err)
			}

			result, err := core.NewProfileUsernameUpdate(daoUpsert).Exec(t.Context(), testCase.request)
			require.ErrorIs(t, err, testCase.expectErr)
			require.Equal(t, testCase.expect, result)

			daoUpsert.AssertExpectations(t)
		})
	}
}
