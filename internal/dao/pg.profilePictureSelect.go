package dao

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/a-novel-kit/golib/otel"
	"github.com/a-novel-kit/golib/postgres"
)

//go:embed pg.profilePictureSelect.sql
var profilePictureSelectQuery string

// ErrProfilePictureSelectNotFound is returned by [PgProfilePictureSelect.Exec] when no profile has the
// requested ID, or the profile has no picture.
var ErrProfilePictureSelectNotFound = errors.New("profile picture not found")

// ProfilePictureSelectRequest is the input to [PgProfilePictureSelect.Exec].
type ProfilePictureSelectRequest struct {
	// ID is the ID of the profile.
	ID uuid.UUID
}

// PgProfilePictureSelect reads the picture of a profile.
type PgProfilePictureSelect struct{}

func NewPgProfilePictureSelect() *PgProfilePictureSelect {
	return &PgProfilePictureSelect{}
}

func (dao *PgProfilePictureSelect) Exec(
	ctx context.Context, request *ProfilePictureSelectRequest,
) (*ProfilePicture, error) {
	ctx, span := otel.Tracer().Start(ctx, "dao.PgProfilePictureSelect")
	defer span.End()

	db, err := postgres.GetContext(ctx)
	if err != nil {
		return nil, otel.ReportError(span, fmt.Errorf("get postgres context: %w", err))
	}

	var entity ProfilePicture

	err = db.NewRaw(profilePictureSelectQuery, request.ID).Scan(ctx, &entity)
	if errors.Is(err, sql.ErrNoRows) {
		err = errors.Join(err, ErrProfilePictureSelectNotFound)
	}

	if err != nil {
		return nil, otel.ReportError(span, fmt.Errorf("execute query: %w", err))
	}

	return &entity, nil
}
