package microservice

import (
	"errors"
	"testing"

	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/logger"
)

type startupPersisterConfig struct{ config.IPersisterConfig }

func (startupPersisterConfig) Host() string { return "postgres" }

type startupLoggerConfig struct{}

func (startupLoggerConfig) LogLevel() string { return "warn" }
func (startupLoggerConfig) DevMode() bool    { return false }
func (startupLoggerConfig) Encoder() string  { return "console" }

type startupConfig struct{ config.IConfig }

func (startupConfig) PersisterConfig() config.IPersisterConfig { return startupPersisterConfig{} }

// flakyPersister fails TestConnect a fixed number of times, like postgres DNS right after a host reboot.
type flakyPersister struct {
	IPersister
	failures int
	calls    int
}

func (p *flakyPersister) TestConnect() error {
	p.calls++
	if p.calls <= p.failures {
		return errors.New("hostname resolving error")
	}
	return nil
}

func TestCheckReadyToStartRetriesUntilPostgresIsReady(t *testing.T) {
	pst := &flakyPersister{failures: 2}
	ms := &Microservice{
		config:     startupConfig{},
		Logger:     startupLogger(),
		persisters: map[string]IPersister{"postgres": pst},
	}

	if err := ms.CheckReadyToStart(); err != nil {
		t.Fatalf("expected success after retries, got %v", err)
	}
	if pst.calls != 3 {
		t.Fatalf("expected 3 connection attempts, got %d", pst.calls)
	}
}

func startupLogger() logger.ILogger {
	appLogger := logger.NewAppLogger(startupLoggerConfig{})
	appLogger.InitLogger()
	return appLogger
}
