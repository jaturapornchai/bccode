package config

type IPersisterClickHouseConfig interface {
	ServerAddress() []string
	DatabaseName() string
	Username() string
	Password() string
}

type PersisterClickHouseConfig struct{}

func NewPersisterClickHouseConfig() *PersisterClickHouseConfig {
	return &PersisterClickHouseConfig{}
}

func (p PersisterClickHouseConfig) ServerAddress() []string {
	addr := getEnv("CH_SERVER_ADDRESS", "")
	return []string{addr}
}

func (p PersisterClickHouseConfig) DatabaseName() string {
	return getEnv("CH_DATABASE_NAME", "")
}

func (p PersisterClickHouseConfig) Username() string {
	return getEnv("CH_USERNAME", "")
}

func (p PersisterClickHouseConfig) Password() string {
	return getEnv("CH_PASSWORD", "")
}
