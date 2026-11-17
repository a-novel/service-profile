package config_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestAppPresetDefault(t *testing.T) {
	t.Parallel()

	// The preset resolves at package initialization, so each case reads it in a process of its own,
	// from one build of the app.
	app := filepath.Join(t.TempDir(), "app")

	//nolint:gosec // The output path is the test's own temporary directory.
	build := exec.CommandContext(t.Context(), "go", "build", "-o", app, "./testdata/app")

	build.Env = append(os.Environ(), "GOWORK=off")

	output, err := build.CombinedOutput()
	require.NoError(t, err, string(output))

	const local = `&grpcf.LocalCredentialsProvider{}`

	testCases := []struct {
		name string

		env []string

		expect string
	}{
		{
			// The unprefixed variable would fail to parse: reading the prefixed one proves the prefix
			// applies.
			name:   "Success",
			env:    []string{"TEST_REST_TIMEOUT_READ=2s"},
			expect: "2s " + local,
		},
		{
			name:   "Success/ServiceToken",
			env:    []string{"TEST_SERVICE_JSON_KEYS_TOKEN_PATH=/tokens/json-keys"},
			expect: `15s &grpcf.ServiceTokenCredentialsProvider{ServiceTokenPath:"/tokens/json-keys", TLS:false}`,
		},
		{
			name: "Success/ServiceTokenTLS",
			env: []string{
				"TEST_SERVICE_JSON_KEYS_TOKEN_PATH=/tokens/json-keys",
				"TEST_SERVICE_JSON_KEYS_TLS=true",
			},
			expect: `15s &grpcf.ServiceTokenCredentialsProvider{ServiceTokenPath:"/tokens/json-keys", TLS:true}`,
		},
		{
			// A local stack has no service token, so it keeps a plaintext connection.
			name:   "Success/TLSWithoutServiceToken",
			env:    []string{"TEST_SERVICE_JSON_KEYS_TLS=true"},
			expect: "15s " + local,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()

			command := exec.CommandContext(t.Context(), app)

			command.Env = append(os.Environ(), "SERVICE_PROFILE_ENV_PREFIX=TEST_", "REST_TIMEOUT_READ=invalid")
			command.Env = append(command.Env, testCase.env...)

			output, err := command.CombinedOutput()
			require.NoError(t, err, string(output))
			require.Equal(t, testCase.expect, strings.TrimSpace(string(output)))
		})
	}
}
