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

//go:embed pg.profileUsernameUpsert.sql
var profileUsernameUpsertQuery string

// ProfileUsernameUpsertRequest is the input to [PgProfileUsernameUpsert.Exec].
type ProfileUsernameUpsertRequest struct {
	// ID is the ID a new profile takes. An existing profile keeps its own.
	ID uuid.UUID
	// UserID is the user who owns the profile.
	UserID uuid.UUID

	// Username is the name to store. Nil clears it.
	Username *string

	// Now is the instant stored as updated_at, and as created_at for a new profile.
	Now time.Time
}

// PgProfileUsernameUpsert sets the username of a user's profile, and creates the profile on the
// user's first write.
type PgProfileUsernameUpsert struct{}

func NewPgProfileUsernameUpsert() *PgProfileUsernameUpsert {
	return &PgProfileUsernameUpsert{}
}

func (dao *PgProfileUsernameUpsert) Exec(ctx context.Context, request *ProfileUsernameUpsertRequest) (*Profile, error) {
	ctx, span := otel.Tracer().Start(ctx, "dao.PgProfileUsernameUpsert")
	defer span.End()

	db, err := postgres.GetContext(ctx)
	if err != nil {
		return nil, otel.ReportError(span, fmt.Errorf("get postgres context: %w", err))
	}

	var entity Profile

	err = db.NewRaw(
		profileUsernameUpsertQuery, request.ID, request.UserID, request.Username, request.Now,
	).Scan(ctx, &entity)
	if err != nil {
		return nil, otel.ReportError(span, fmt.Errorf("execute query: %w", err))
	}

	return &entity, nil
}
