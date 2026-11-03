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

func TestPgProfilePictureUpsert(t *testing.T) {
	t.Parallel()

	newPicture := []byte("new picture")

	testCases := []struct {
		name string

		request *dao.ProfilePictureUpsertRequest

		expect          *dao.Profile
		expectRow       *profileRow
		expectViolation string
	}{
		{
			name: "Success/Create",

			request: &dao.ProfilePictureUpsertRequest{
				ID:      newProfileID,
				UserID:  newUserID,
				Picture: newPicture,
				Type:    "image/jpeg",
				Now:     now,
			},

			expect: &dao.Profile{
				ID:               newProfileID,
				PictureUpdatedAt: &now,
			},
			expectRow: &profileRow{
				ID:               newProfileID,
				UserID:           newUserID,
				Picture:          newPicture,
				PictureType:      lo.ToPtr("image/jpeg"),
				PictureUpdatedAt: &now,
				CreatedAt:        now,
				UpdatedAt:        now,
			},
		},
		{
			// The profile keeps its ID and its username.
			name: "Success/Update",

			request: &dao.ProfilePictureUpsertRequest{
				ID:      newProfileID,
				UserID:  otherUserID,
				Picture: newPicture,
				Type:    "image/webp",
				Now:     now,
			},

			expect: &dao.Profile{
				ID:               otherProfileID,
				Username:         lo.ToPtr("Marie"),
				PictureUpdatedAt: &now,
			},
			expectRow: &profileRow{
				ID:               otherProfileID,
				UserID:           otherUserID,
				Username:         lo.ToPtr("Marie"),
				Picture:          newPicture,
				PictureType:      lo.ToPtr("image/webp"),
				PictureUpdatedAt: &now,
				CreatedAt:        createdAt,
				UpdatedAt:        now,
			},
		},
		{
			name: "Error/Type",

			request: &dao.ProfilePictureUpsertRequest{
				ID:      newProfileID,
				UserID:  newUserID,
				Picture: newPicture,
				Type:    "image/png",
				Now:     now,
			},

			expectViolation: "profiles_picture_type_check",
		},
		{
			name: "Error/TooLarge",

			request: &dao.ProfilePictureUpsertRequest{
				ID:      newProfileID,
				UserID:  newUserID,
				Picture: make([]byte, 262145),
				Type:    "image/jpeg",
				Now:     now,
			},

			expectViolation: "profiles_picture_check",
		},
		{
			name: "Error/NoPicture",

			request: &dao.ProfilePictureUpsertRequest{
				ID:     newProfileID,
				UserID: newUserID,
				Type:   "image/jpeg",
				Now:    now,
			},

			expectViolation: "profiles_picture_set_check",
		},
	}

	daoProfilePictureUpsert := dao.NewPgProfilePictureUpsert()

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			postgrestest.RunDBTest(t, configtest.PostgresPreset, migrations.Migrations, func(ctx context.Context, t *testing.T) {
				t.Helper()

				insertProfiles(ctx, t)

				result, err := daoProfilePictureUpsert.Exec(ctx, testCase.request)
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
