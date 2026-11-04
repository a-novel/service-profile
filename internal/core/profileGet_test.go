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

func TestProfileGet(t *testing.T) {
	t.Parallel()

	errFoo := errors.New("foo")

	type daoMock struct {
		resp *dao.Profile
		err  error
	}

	testCases := []struct {
		name string

		request *core.ProfileGetRequest

		daoMock *daoMock

		expect    *core.Profile
		expectErr error
	}{
		{
			name: "Success",

			request: &core.ProfileGetRequest{ID: profileID},

			daoMock: &daoMock{resp: daoProfile},

			expect: coreProfile,
		},
		{
			name: "Error/NotFound",

			request: &core.ProfileGetRequest{ID: profileID},

			daoMock: &daoMock{err: dao.ErrProfileSelectNotFound},

			expectErr: core.ErrProfileSelectNotFound,
		},
		{
			name: "Error/Dao",

			request: &core.ProfileGetRequest{ID: profileID},

			daoMock: &daoMock{err: errFoo},

			expectErr: errFoo,
		},
		{
			name: "Error/NoID",

			request: &core.ProfileGetRequest{ID: uuid.Nil},

			expectErr: core.ErrInvalidRequest,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			daoSelect := coremocks.NewMockProfileGetDao(t)

			if testCase.daoMock != nil {
				daoSelect.EXPECT().
					Exec(mock.Anything, &dao.ProfileSelectRequest{ID: testCase.request.ID}).
					Return(testCase.daoMock.resp, testCase.daoMock.err)
			}

			result, err := core.NewProfileGet(daoSelect).Exec(t.Context(), testCase.request)
			require.ErrorIs(t, err, testCase.expectErr)
			require.Equal(t, testCase.expect, result)

			daoSelect.AssertExpectations(t)
		})
	}
}
