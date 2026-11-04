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

//go:embed pg.profileSelect.sql
var profileSelectQuery string

// ErrProfileSelectNotFound is returned by [PgProfileSelect.Exec] when no profile has the requested ID.
var ErrProfileSelectNotFound = errors.New("profile not found")

// ProfileSelectRequest is the input to [PgProfileSelect.Exec].
type ProfileSelectRequest struct {
	ID uuid.UUID
}

// PgProfileSelect reads a profile by its ID.
type PgProfileSelect struct{}

func NewPgProfileSelect() *PgProfileSelect {
	return &PgProfileSelect{}
}

func (dao *PgProfileSelect) Exec(ctx context.Context, request *ProfileSelectRequest) (*Profile, error) {
	ctx, span := otel.Tracer().Start(ctx, "dao.PgProfileSelect")
	defer span.End()

	db, err := postgres.GetContext(ctx)
	if err != nil {
		return nil, otel.ReportError(span, fmt.Errorf("get postgres context: %w", err))
	}

	var entity Profile

	err = db.NewRaw(profileSelectQuery, request.ID).Scan(ctx, &entity)
	if errors.Is(err, sql.ErrNoRows) {
		err = errors.Join(err, ErrProfileSelectNotFound)
	}

	if err != nil {
		return nil, otel.ReportError(span, fmt.Errorf("execute query: %w", err))
	}

	return &entity, nil
}
