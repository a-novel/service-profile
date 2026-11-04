package dao

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/a-novel-kit/golib/otel"
	"github.com/a-novel-kit/golib/postgres"
)

//go:embed pg.profilePictureDelete.sql
var profilePictureDeleteQuery string

// ErrProfilePictureDeleteNotFound is returned by [PgProfilePictureDelete.Exec] when the user holds no
// profile.
var ErrProfilePictureDeleteNotFound = errors.New("profile not found")

// ProfilePictureDeleteRequest is the input to [PgProfilePictureDelete.Exec].
type ProfilePictureDeleteRequest struct {
	// UserID is the user who owns the profile.
	UserID uuid.UUID
	// Now is the instant stored as updated_at.
	Now time.Time
}

// PgProfilePictureDelete removes the picture of a user's profile. A profile without a picture keeps
// none.
type PgProfilePictureDelete struct{}

func NewPgProfilePictureDelete() *PgProfilePictureDelete {
	return &PgProfilePictureDelete{}
}

func (dao *PgProfilePictureDelete) Exec(ctx context.Context, request *ProfilePictureDeleteRequest) (*Profile, error) {
	ctx, span := otel.Tracer().Start(ctx, "dao.PgProfilePictureDelete")
	defer span.End()

	db, err := postgres.GetContext(ctx)
	if err != nil {
		return nil, otel.ReportError(span, fmt.Errorf("get postgres context: %w", err))
	}

	var entity Profile

	err = db.NewRaw(profilePictureDeleteQuery, request.UserID, request.Now).Scan(ctx, &entity)
	if errors.Is(err, sql.ErrNoRows) {
		err = errors.Join(err, ErrProfilePictureDeleteNotFound)
	}

	if err != nil {
		return nil, otel.ReportError(span, fmt.Errorf("execute query: %w", err))
	}

	return &entity, nil
}
