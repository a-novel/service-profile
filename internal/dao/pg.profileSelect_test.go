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

func TestPgProfileSelect(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		request *dao.ProfileSelectRequest

		expect    *dao.Profile
		expectErr error
	}{
		{
			name: "Success",

			request: &dao.ProfileSelectRequest{ID: profileID},

			expect: &dao.Profile{
				ID:               profileID,
				Username:         lo.ToPtr("Jean Paul"),
				PictureUpdatedAt: &updatedAt,
			},
		},
		{
			name: "Success/NoPicture",

			request: &dao.ProfileSelectRequest{ID: otherProfileID},

			expect: &dao.Profile{
				ID:       otherProfileID,
				Username: lo.ToPtr("Marie"),
			},
		},
		{
			name: "Error/NotFound",

			request: &dao.ProfileSelectRequest{ID: absentID},

			expectErr: dao.ErrProfileSelectNotFound,
		},
	}

	daoProfileSelect := dao.NewPgProfileSelect()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			postgrestest.RunDBTest(t, configtest.PostgresPreset, migrations.Migrations, func(ctx context.Context, t *testing.T) {
				t.Helper()

				insertProfiles(ctx, t)

				result, err := daoProfileSelect.Exec(ctx, testCase.request)
				require.ErrorIs(t, err, testCase.expectErr)
				require.Equal(t, testCase.expect, result)
			})
		})
	}
}
