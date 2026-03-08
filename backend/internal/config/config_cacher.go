package config

import "time"

type CacherConfig struct{}

// ICacherConfig is cacher configuration interface
type ICacherConfig interface {
	Endpoint() string
	Password() string
	UserName() string
	TLS() bool
	DB() int
	ConnectionSettings() ICacherConnectionSettings
}

// ICacherConnectionSettings is connection settings for cacher
type ICacherConnectionSettings interface {
	PoolSize() int
	MinIdleConns() int
	MaxRetries() int
	MinRetryBackoff() time.Duration
	MaxRetryBackoff() time.Duration
	IdleTimeout() time.Duration
	IdleCheckFrequency() time.Duration
	PoolTimeout() time.Duration
	ReadTimeout() time.Duration
	WriteTimeout() time.Duration
}

func NewCacherConfig() *CacherConfig {
	return &CacherConfig{}
}

func (cfg *CacherConfig) Endpoint() string {
	return getEnv("REDIS_CACHE_URI", "")
}

func (cfg *CacherConfig) Password() string {
	return getEnv("REDIS_CACHE_PASSWORD", "")
}

func (cfg *CacherConfig) UserName() string {
	return getEnv("REDIS_CACHE_USERNAME", "")
}

func (cfg *CacherConfig) TLS() bool {
	tlsEnable := getEnv("REDIS_CACHE_TLS_ENABLE", "")

	if tlsEnable != "" && tlsEnable == "true" {
		return true
	}
	return false
}

func (cfg *CacherConfig) DB() int {
	return 0
}

func (cfg *CacherConfig) ConnectionSettings() ICacherConnectionSettings {
	return NewDefaultCacherConnectionSettings()
}

// DefaultCacherConnectionSettings contains default connection settings, this intend to use as embed struct
type DefaultCacherConnectionSettings struct{}

func NewDefaultCacherConnectionSettings() ICacherConnectionSettings {
	return &DefaultCacherConnectionSettings{}
}

func (setting *DefaultCacherConnectionSettings) PoolSize() int {
	return 50
}

func (setting *DefaultCacherConnectionSettings) MinIdleConns() int {
	return 5
}

func (setting *DefaultCacherConnectionSettings) MaxRetries() int {
	return 3
}

func (setting *DefaultCacherConnectionSettings) MinRetryBackoff() time.Duration {
	return 10 * time.Millisecond
}

func (setting *DefaultCacherConnectionSettings) MaxRetryBackoff() time.Duration {
	return 500 * time.Millisecond
}

func (setting *DefaultCacherConnectionSettings) IdleTimeout() time.Duration {
	return 30 * time.Minute
}

func (setting *DefaultCacherConnectionSettings) IdleCheckFrequency() time.Duration {
	return time.Minute
}

func (setting *DefaultCacherConnectionSettings) PoolTimeout() time.Duration {
	return time.Minute
}

func (setting *DefaultCacherConnectionSettings) ReadTimeout() time.Duration {
	return time.Minute
}

func (setting *DefaultCacherConnectionSettings) WriteTimeout() time.Duration {
	return time.Minute
}
