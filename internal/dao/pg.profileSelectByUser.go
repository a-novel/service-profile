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

//go:embed pg.profileSelectByUser.sql
var profileSelectByUserQuery string

// ErrProfileSelectByUserNotFound is returned by [PgProfileSelectByUser.Exec] when the user holds no
// profile.
var ErrProfileSelectByUserNotFound = errors.New("profile not found")

// ProfileSelectByUserRequest is the input to [PgProfileSelectByUser.Exec].
type ProfileSelectByUserRequest struct {
	// UserID is the user who owns the profile.
	UserID uuid.UUID
}

// PgProfileSelectByUser reads the profile of a user.
type PgProfileSelectByUser struct{}

func NewPgProfileSelectByUser() *PgProfileSelectByUser {
	return &PgProfileSelectByUser{}
}

func (dao *PgProfileSelectByUser) Exec(ctx context.Context, request *ProfileSelectByUserRequest) (*Profile, error) {
	ctx, span := otel.Tracer().Start(ctx, "dao.PgProfileSelectByUser")
	defer span.End()

	db, err := postgres.GetContext(ctx)
	if err != nil {
		return nil, otel.ReportError(span, fmt.Errorf("get postgres context: %w", err))
	}

	var entity Profile

	err = db.NewRaw(profileSelectByUserQuery, request.UserID).Scan(ctx, &entity)
	if errors.Is(err, sql.ErrNoRows) {
		err = errors.Join(err, ErrProfileSelectByUserNotFound)
	}

	if err != nil {
		return nil, otel.ReportError(span, fmt.Errorf("execute query: %w", err))
	}

	return &entity, nil
}
