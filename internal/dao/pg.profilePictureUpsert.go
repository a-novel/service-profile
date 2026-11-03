package dao

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/a-novel-kit/golib/otel"
	"github.com/a-novel-kit/golib/postgres"
)

//go:embed pg.profilePictureUpsert.sql
var profilePictureUpsertQuery string

// ProfilePictureUpsertRequest is the input to [PgProfilePictureUpsert.Exec].
type ProfilePictureUpsertRequest struct {
	// ID is the ID a new profile takes. An existing profile keeps its own.
	ID uuid.UUID
	// UserID is the user who owns the profile.
	UserID uuid.UUID

	// Picture holds the encoded image.
	Picture []byte
	// Type is the media type the picture is encoded in.
	Type string

	// Now is the instant stored as the picture's version and updated_at, and as created_at for a new
	// profile.
	Now time.Time
}

// PgProfilePictureUpsert replaces the picture of a user's profile, and creates the profile on the
// user's first write.
type PgProfilePictureUpsert struct{}

func NewPgProfilePictureUpsert() *PgProfilePictureUpsert {
	return &PgProfilePictureUpsert{}
}

func (dao *PgProfilePictureUpsert) Exec(ctx context.Context, request *ProfilePictureUpsertRequest) (*Profile, error) {
	ctx, span := otel.Tracer().Start(ctx, "dao.PgProfilePictureUpsert")
	defer span.End()

	db, err := postgres.GetContext(ctx)
	if err != nil {
		return nil, otel.ReportError(span, fmt.Errorf("get postgres context: %w", err))
	}

	var entity Profile

	err = db.NewRaw(
		profilePictureUpsertQuery, request.ID, request.UserID, request.Picture, request.Type, request.Now,
	).Scan(ctx, &entity)
	if err != nil {
		return nil, otel.ReportError(span, fmt.Errorf("execute query: %w", err))
	}

	return &entity, nil
}
