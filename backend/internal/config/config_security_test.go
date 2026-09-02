package config

import (
	"os"
	"testing"
)

func TestJwtSecretKeyHasNoSourceFallback(t *testing.T) {
	t.Setenv("JWT_SECRET_KEY", "")
	if got := NewConfig().JwtSecretKey(); got != "" {
		t.Fatalf("JwtSecretKey() = %q, want empty when deployment secret is absent", got)
	}

	t.Setenv("JWT_SECRET_KEY", "deployment-secret")
	if got := NewConfig().JwtSecretKey(); got != "deployment-secret" {
		t.Fatalf("JwtSecretKey() = %q, want deployment value", got)
	}
}

func TestGoogleClientIdUsesDeploymentOverride(t *testing.T) {
	previous, existed := os.LookupEnv("GOOGLE_CLIENT_ID")
	if err := os.Unsetenv("GOOGLE_CLIENT_ID"); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if existed {
			_ = os.Setenv("GOOGLE_CLIENT_ID", previous)
		} else {
			_ = os.Unsetenv("GOOGLE_CLIENT_ID")
		}
	})

	if got := NewConfig().GoogleClientId(); got != "212036599086-c7aqvm005jiv2kqi4duju8spd9b3jb94.apps.googleusercontent.com" {
		t.Fatalf("GoogleClientId() = %q, want current web client ID", got)
	}

	t.Setenv("GOOGLE_CLIENT_ID", "deployment-client-id")
	if got := NewConfig().GoogleClientId(); got != "deployment-client-id" {
		t.Fatalf("GoogleClientId() = %q, want deployment override", got)
	}
}
