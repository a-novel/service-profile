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

func TestPgProfilePictureDelete(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		request *dao.ProfilePictureDeleteRequest

		expect    *dao.Profile
		expectRow *profileRow
		expectErr error
	}{
		{
			name: "Success",

			request: &dao.ProfilePictureDeleteRequest{UserID: userID, Now: now},

			expect: &dao.Profile{
				ID:       profileID,
				Username: lo.ToPtr("Jean Paul"),
			},
			expectRow: &profileRow{
				ID:        profileID,
				UserID:    userID,
				Username:  lo.ToPtr("Jean Paul"),
				CreatedAt: createdAt,
				UpdatedAt: now,
			},
		},
		{
			name: "Success/NoPicture",

			request: &dao.ProfilePictureDeleteRequest{UserID: otherUserID, Now: now},

			expect: &dao.Profile{
				ID:       otherProfileID,
				Username: lo.ToPtr("Marie"),
			},
			expectRow: &profileRow{
				ID:        otherProfileID,
				UserID:    otherUserID,
				Username:  lo.ToPtr("Marie"),
				CreatedAt: createdAt,
				UpdatedAt: now,
			},
		},
		{
			name: "Error/NotFound",

			request: &dao.ProfilePictureDeleteRequest{UserID: newUserID, Now: now},

			expectErr: dao.ErrProfilePictureDeleteNotFound,
		},
	}

	daoProfilePictureDelete := dao.NewPgProfilePictureDelete()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			postgrestest.RunDBTest(t, configtest.PostgresPreset, migrations.Migrations, func(ctx context.Context, t *testing.T) {
				t.Helper()

				insertProfiles(ctx, t)

				result, err := daoProfilePictureDelete.Exec(ctx, testCase.request)
				require.ErrorIs(t, err, testCase.expectErr)
				require.Equal(t, testCase.expect, result)

				if testCase.expectRow != nil {
					require.Equal(t, testCase.expectRow, selectProfileRow(ctx, t, testCase.request.UserID))
				}
			})
		})
	}
}
