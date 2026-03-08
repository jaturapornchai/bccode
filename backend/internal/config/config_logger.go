package config

type ILoggerConfig interface {
	LogLevel() string
	DevMode() bool
	Encoder() string
}

type LoggerConfig struct{}

func NewLoggerConfig() ILoggerConfig {
	return &LoggerConfig{}
}

func (*LoggerConfig) LogLevel() string {
	return getEnv("LOG_LEVEL", "info")
}

func (*LoggerConfig) Encoder() string {
	return getEnv("LOG_ENCODER", "")
}

func (*LoggerConfig) DevMode() bool {
	env := getEnv("MODE", "development")
	return env == "development"
}
