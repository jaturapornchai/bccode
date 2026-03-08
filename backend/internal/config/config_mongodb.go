package config

import (
	"fmt"
	"strings"
)

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

	uri := getEnv("MONGODB_URI", "") // mongodb://root:rootx@localhost:27017/
	if uri != "" {
		return uri
	}

	if cfg.MongodbServer() == "" {
		return ""
	}

	userNamePassword := fmt.Sprintf("%s:%s@", cfg.MongodbUserName(), cfg.MongodbPassWord())
	if userNamePassword == ":@" {
		userNamePassword = ""
	}

	var connectionOptions []string

	ssl := cfg.MongoConnectionSSL()
	if ssl != "" {
		connectionOptions = append(connectionOptions, ssl)
	}
	tlsCaFile := cfg.MongoTlsCaFile()
	if ssl != "" {
		connectionOptions = append(connectionOptions, tlsCaFile)
	}

	connetionOptional := ""
	joinConnectionOption := strings.Join(connectionOptions[:], "&")
	if joinConnectionOption != "" {
		connetionOptional = "?" + joinConnectionOption
	}

	connectionUri := fmt.Sprintf("%s://%s%s%s/%s",
		cfg.MongodbProtocal(),
		userNamePassword, cfg.MongodbServer(),
		cfg.MongodbPort(),
		connetionOptional,
	)
	return connectionUri
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
	return getEnv("MONGODB_DB", "bcaiclouddb")
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
