package dao_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/a-novel-kit/golib/postgres/postgrestest"

	"github.com/a-novel/service-profile/internal/config/configtest"
	"github.com/a-novel/service-profile/internal/dao"
	"github.com/a-novel/service-profile/internal/models/migrations"
)

func TestPgProfilePictureSelect(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		request *dao.ProfilePictureSelectRequest

		expect    *dao.ProfilePicture
		expectErr error
	}{
		{
			name: "Success",

			request: &dao.ProfilePictureSelectRequest{ID: profileID},

			expect: &dao.ProfilePicture{Picture: picture, Type: "image/webp", UpdatedAt: updatedAt},
		},
		{
			name: "Error/NoPicture",

			request: &dao.ProfilePictureSelectRequest{ID: otherProfileID},

			expectErr: dao.ErrProfilePictureSelectNotFound,
		},
		{
			name: "Error/NotFound",

			request: &dao.ProfilePictureSelectRequest{ID: absentID},

			expectErr: dao.ErrProfilePictureSelectNotFound,
		},
	}

	daoProfilePictureSelect := dao.NewPgProfilePictureSelect()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			postgrestest.RunDBTest(t, configtest.PostgresPreset, migrations.Migrations, func(ctx context.Context, t *testing.T) {
				t.Helper()

				insertProfiles(ctx, t)

				result, err := daoProfilePictureSelect.Exec(ctx, testCase.request)
				require.ErrorIs(t, err, testCase.expectErr)
				require.Equal(t, testCase.expect, result)
			})
		})
	}
}
