package configtest

import (
	"github.com/a-novel/service-profile/internal/config"
)

// PostgresPreset is the PostgreSQL configuration the integration tests connect with. It aliases the
// production preset, so tests read the same POSTGRES_* settings.
var PostgresPreset = config.PostgresPresetDefault
