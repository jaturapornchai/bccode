package config

import "fmt"

// IPersisterConfig is interface for persister
type IPersisterMongoConfig interface {
	MongodbURI() string
	DB() string
	Debug() bool
}

type MongoPersisterConfig struct{}

func NewMongoPersisterConfig() *MongoPersisterConfig {
	return &MongoPersisterConfig{}
}

func (cfg *MongoPersisterConfig) MongodbURI() string {
	return MongoURIForCurrentEnvironment()
}

func (cfg *MongoPersisterConfig) MongodbProtocal() string {
	return getEnv("MONGODB_PROTOCAL", "")
}

func (cfg *MongoPersisterConfig) MongodbServer() string {
	return getEnv("MONGODB_SERVER", "")
}

func (cfg *MongoPersisterConfig) MongodbPort() string {
	port := getEnv("MONGODB_PORT", "")
	if port != "" {
		return ":" + port
	}
	return port
}

func (cfg *MongoPersisterConfig) DB() string {
	return MongoDatabaseForCurrentEnvironment("bcaiclouddb")
}

func (cfg *MongoPersisterConfig) MongodbUserName() string {
	return getEnv("MONGODB_USERNAME", "")
}

func (cfg *MongoPersisterConfig) MongodbPassWord() string {
	return getEnv("MONGODB_PASSWORD", "")
}

func (cfg *MongoPersisterConfig) MongoConnectionSSL() string {
	sslMode := getEnv("MONGODB_SSL", "")
	if sslMode != "" {
		return fmt.Sprintf("ssl=%s", sslMode)
	}
	return sslMode
}

func (cfg *MongoPersisterConfig) MongoTlsCaFile() string {
	tlsCaFile := getEnv("MONGODB_TLS_CA_FILE", "")
	if tlsCaFile != "" {
		return fmt.Sprintf("tlsCAFile=%s", tlsCaFile)
	}
	return tlsCaFile
}

func (cfg *MongoPersisterConfig) Debug() bool {
	return getEnv("MONGODB_DEBUG", "false") == "true"
}
