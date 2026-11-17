package config_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	postgrespresets "github.com/a-novel-kit/golib/postgres/presets"

	"github.com/a-novel/service-profile/internal/config"
)

// TestNewPostgresPreset checks the pool limits the preset sets. golib's presets tests cover the
// connection options.
func TestNewPostgresPreset(t *testing.T) {
	t.Parallel()

	connection := postgrespresets.Connection{
		Host:     "postgres.internal",
		User:     "profile",
		Password: "password",
		Database: "profile",
	}

	preset := config.NewPostgresPreset(connection, 12, 8)

	require.Equal(t, 12, preset.MaxOpenConns)
	require.Equal(t, 8, preset.MaxIdleConns)
}
