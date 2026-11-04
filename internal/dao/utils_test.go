package dao_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/samber/lo"
	"github.com/stretchr/testify/require"
	"github.com/uptrace/bun"
	"github.com/uptrace/bun/driver/pgdriver"

	"github.com/a-novel-kit/golib/postgres"
)

var (
	// userID owns profileID, which holds a username and a picture.
	userID = uuid.MustParse("00000000-0000-0000-0000-0000000000a1")
	// otherUserID owns otherProfileID, which holds a username and no picture.
	otherUserID = uuid.MustParse("00000000-0000-0000-0000-0000000000a2")
	// newUserID holds no profile.
	newUserID = uuid.MustParse("00000000-0000-0000-0000-0000000000a3")

	profileID      = uuid.MustParse("00000000-0000-0000-0000-000000000001")
	otherProfileID = uuid.MustParse("00000000-0000-0000-0000-000000000002")
	// newProfileID is the ID a write offers a new profile.
	newProfileID = uuid.MustParse("00000000-0000-0000-0000-000000000003")
	// absentID names no profile.
	absentID = uuid.MustParse("00000000-0000-0000-0000-0000000000ff")

	createdAt = time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC)
	updatedAt = time.Date(2021, 1, 2, 0, 0, 0, 0, time.UTC)
	now       = time.Date(2021, 1, 3, 0, 0, 0, 0, time.UTC)

	// picture stands for an encoded image: the DAO stores bytes without reading them.
	picture = []byte("picture")
)

// profileRow is a whole row of the profiles table, which no DAO read returns.
type profileRow struct {
	bun.BaseModel `bun:"table:profiles,alias:profiles"`

	ID     uuid.UUID `bun:"id,pk,type:uuid"`
	UserID uuid.UUID `bun:"user_id,type:uuid"`

	Username *string `bun:"username"`

	Picture          []byte     `bun:"picture"`
	PictureType      *string    `bun:"picture_type"`
	PictureUpdatedAt *time.Time `bun:"picture_updated_at"`

	CreatedAt time.Time `bun:"created_at"`
	UpdatedAt time.Time `bun:"updated_at"`
}

// profileFixtures returns the profiles every DAO test seeds: one with a picture, one without. The
// insert scans the returned rows back into them, so they point at copies, never at the variables
// parallel tests share.
func profileFixtures() []*profileRow {
	return []*profileRow{
		{
			ID:               profileID,
			UserID:           userID,
			Username:         lo.ToPtr("Jean Paul"),
			Picture:          picture,
			PictureType:      lo.ToPtr("image/webp"),
			PictureUpdatedAt: lo.ToPtr(updatedAt),
			CreatedAt:        createdAt,
			UpdatedAt:        updatedAt,
		},
		{
			ID:        otherProfileID,
			UserID:    otherUserID,
			Username:  lo.ToPtr("Marie"),
			CreatedAt: createdAt,
			UpdatedAt: updatedAt,
		},
	}
}

// insertProfiles seeds the database of ctx with [profileFixtures].
func insertProfiles(ctx context.Context, t *testing.T) {
	t.Helper()

	db, err := postgres.GetContext(ctx)
	require.NoError(t, err)

	fixtures := profileFixtures()

	_, err = db.NewInsert().Model(&fixtures).Exec(ctx)
	require.NoError(t, err)
}

// selectProfileRow reads the whole row of a user's profile, picture included.
func selectProfileRow(ctx context.Context, t *testing.T, owner uuid.UUID) *profileRow {
	t.Helper()

	db, err := postgres.GetContext(ctx)
	require.NoError(t, err)

	var row profileRow

	err = db.NewSelect().Model(&row).Where("user_id = ?", owner).Scan(ctx)
	require.NoError(t, err)

	return &row
}

// requireViolation checks that err comes from a write that broke the named constraint.
func requireViolation(t *testing.T, err error, constraint string) {
	t.Helper()

	var pgErr pgdriver.Error

	require.ErrorAs(t, err, &pgErr)
	require.Equal(t, constraint, pgErr.Field('n'))
}
