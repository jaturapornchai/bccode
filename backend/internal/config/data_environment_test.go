package config_test

import (
	"smlcloudplatform/internal/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestConfiguredDataEnvironmentRequiresExplicitValue(t *testing.T) {
	clearDataEnvironment(t)

	dataEnvironment, configured := config.ConfiguredDataEnvironment()

	assert.False(t, configured)
	assert.Empty(t, dataEnvironment)
	assert.Equal(t, config.DataEnvironmentDev, config.CurrentDataEnvironment())
}

func TestConfiguredDataEnvironmentNormalizesExplicitValue(t *testing.T) {
	clearDataEnvironment(t)
	t.Setenv("BC_ENV", " production ")

	dataEnvironment, configured := config.ConfiguredDataEnvironment()

	assert.True(t, configured)
	assert.Equal(t, config.DataEnvironmentPRO, dataEnvironment)
}

func clearDataEnvironment(t *testing.T) {
	t.Helper()
	for _, key := range []string{"BC_ENV", "APP_ENV", "RUN_ENV", "ENVIRONMENT", "MODE"} {
		t.Setenv(key, "")
	}
}
