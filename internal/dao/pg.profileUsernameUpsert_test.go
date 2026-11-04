package dao_test

import (
	"context"
	"strings"
	"testing"

	"github.com/samber/lo"
	"github.com/stretchr/testify/require"

	"github.com/a-novel-kit/golib/postgres/postgrestest"

	"github.com/a-novel/service-profile/internal/config/configtest"
	"github.com/a-novel/service-profile/internal/dao"
	"github.com/a-novel/service-profile/internal/models/migrations"
)

func TestPgProfileUsernameUpsert(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string

		request *dao.ProfileUsernameUpsertRequest

		expect          *dao.Profile
		expectRow       *profileRow
		expectViolation string
	}{
		{
			name: "Success/Create",

			request: &dao.ProfileUsernameUpsertRequest{
				ID:       newProfileID,
				UserID:   newUserID,
				Username: lo.ToPtr("Jean Paul"),
				Now:      now,
			},

			expect: &dao.Profile{
				ID:       newProfileID,
				Username: lo.ToPtr("Jean Paul"),
			},
			expectRow: &profileRow{
				ID:        newProfileID,
				UserID:    newUserID,
				Username:  lo.ToPtr("Jean Paul"),
				CreatedAt: now,
				UpdatedAt: now,
			},
		},
		{
			// The profile keeps its ID and its picture.
			name: "Success/Update",

			request: &dao.ProfileUsernameUpsertRequest{
				ID:       newProfileID,
				UserID:   userID,
				Username: lo.ToPtr("Marie Curie"),
				Now:      now,
			},

			expect: &dao.Profile{
				ID:               profileID,
				Username:         lo.ToPtr("Marie Curie"),
				PictureUpdatedAt: &updatedAt,
			},
			expectRow: &profileRow{
				ID:               profileID,
				UserID:           userID,
				Username:         lo.ToPtr("Marie Curie"),
				Picture:          picture,
				PictureType:      lo.ToPtr("image/webp"),
				PictureUpdatedAt: &updatedAt,
				CreatedAt:        createdAt,
				UpdatedAt:        now,
			},
		},
		{
			// The profile keeps its ID and its picture.
			name: "Success/Clear",

			request: &dao.ProfileUsernameUpsertRequest{ID: newProfileID, UserID: userID, Now: now},

			expect: &dao.Profile{
				ID:               profileID,
				PictureUpdatedAt: &updatedAt,
			},
			expectRow: &profileRow{
				ID:               profileID,
				UserID:           userID,
				Picture:          picture,
				PictureType:      lo.ToPtr("image/webp"),
				PictureUpdatedAt: &updatedAt,
				CreatedAt:        createdAt,
				UpdatedAt:        now,
			},
		},
		{
			name: "Error/Empty",

			request: &dao.ProfileUsernameUpsertRequest{
				ID:       newProfileID,
				UserID:   newUserID,
				Username: lo.ToPtr(""),
				Now:      now,
			},

			expectViolation: "profiles_username_check",
		},
		{
			name: "Error/TooLong",

			request: &dao.ProfileUsernameUpsertRequest{
				ID:       newProfileID,
				UserID:   newUserID,
				Username: lo.ToPtr(strings.Repeat("a", 2049)),
				Now:      now,
			},

			expectViolation: "profiles_username_check",
		},
	}

	daoProfileUsernameUpsert := dao.NewPgProfileUsernameUpsert()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			postgrestest.RunDBTest(t, configtest.PostgresPreset, migrations.Migrations, func(ctx context.Context, t *testing.T) {
				t.Helper()

				insertProfiles(ctx, t)

				result, err := daoProfileUsernameUpsert.Exec(ctx, testCase.request)
				if testCase.expectViolation != "" {
					requireViolation(t, err, testCase.expectViolation)

					return
				}

				require.NoError(t, err)
				require.Equal(t, testCase.expect, result)
				require.Equal(t, testCase.expectRow, selectProfileRow(ctx, t, testCase.request.UserID))
			})
		})
	}
}
