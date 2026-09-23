package config

import (
	"os"
	"strings"
)

const (
	DataEnvironmentDev = "dev"
	DataEnvironmentUAT = "uat"
	DataEnvironmentPRO = "pro"
)

func ConfiguredDataEnvironment() (string, bool) {
	for _, key := range []string{"BC_ENV", "APP_ENV", "RUN_ENV", "ENVIRONMENT", "MODE"} {
		if value := strings.TrimSpace(os.Getenv(key)); value != "" {
			return NormalizeDataEnvironment(value), true
		}
	}
	return "", false
}

func CurrentDataEnvironment() string {
	if value, configured := ConfiguredDataEnvironment(); configured {
		return value
	}
	return DataEnvironmentDev
}

func NormalizeDataEnvironment(value string) string {
	switch strings.ToLower(strings.TrimSpace(value)) {
	case "uat", "staging", "stage":
		return DataEnvironmentUAT
	case "pro", "prod", "production":
		return DataEnvironmentPRO
	default:
		return DataEnvironmentDev
	}
}
