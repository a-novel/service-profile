package dao_test

import (
	"context"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/a-novel-kit/golib/postgres/postgrestest"

	"github.com/a-novel/service-profile/internal/config/configtest"
	"github.com/a-novel/service-profile/internal/dao"
	"github.com/a-novel/service-profile/internal/models/migrations"
)

func TestPgProfileSelectByUser(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		request *dao.ProfileSelectByUserRequest

		expect    *dao.Profile
		expectErr error
	}{
		{
			name: "Success",

			request: &dao.ProfileSelectByUserRequest{UserID: userID},

			expect: &dao.Profile{
				ID:               profileID,
				Username:         lo.ToPtr("Jean Paul"),
				PictureUpdatedAt: &updatedAt,
			},
		},
		{
			name: "Error/NotFound",

			request: &dao.ProfileSelectByUserRequest{UserID: newUserID},

			expectErr: dao.ErrProfileSelectByUserNotFound,
		},
	}

	daoProfileSelectByUser := dao.NewPgProfileSelectByUser()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			postgrestest.RunDBTest(t, configtest.PostgresPreset, migrations.Migrations, func(ctx context.Context, t *testing.T) {
				t.Helper()

				insertProfiles(ctx, t)

				result, err := daoProfileSelectByUser.Exec(ctx, testCase.request)
				require.ErrorIs(t, err, testCase.expectErr)
				require.Equal(t, testCase.expect, result)
			})
		})
	}
}
