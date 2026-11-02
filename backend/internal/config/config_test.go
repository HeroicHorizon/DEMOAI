package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func env(m map[string]string) func(string) string {
	return func(k string) string { return m[k] }
}

func TestLoad_MissingRequired(t *testing.T) {
	_, err := Load(env(map[string]string{}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "MONGODB_URI")
	assert.Contains(t, err.Error(), "MONGODB_DB")
	assert.Contains(t, err.Error(), "CORS_ALLOWED_ORIGIN")
}

func TestLoad_Defaults(t *testing.T) {
	cfg, err := Load(env(map[string]string{
		"MONGODB_URI": "mongodb://x", "MONGODB_DB": "d", "CORS_ALLOWED_ORIGIN": "http://o",
	}))
	require.NoError(t, err)
	assert.Equal(t, "8080", cfg.Port)
	assert.EqualValues(t, 1<<20, cfg.MaxBodySize)
}

func TestLoad_InvalidMaxBody(t *testing.T) {
	_, err := Load(env(map[string]string{
		"MONGODB_URI": "mongodb://x", "MONGODB_DB": "d", "CORS_ALLOWED_ORIGIN": "http://o", "MAX_BODY_BYTES": "abc",
	}))
	assert.Error(t, err)
}
