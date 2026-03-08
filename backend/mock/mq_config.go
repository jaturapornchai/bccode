package mock

import (
	"fmt"
	"os"
	"regexp"
	"smlcloudplatform/internal/config"

	"github.com/joho/godotenv"
)

type MqConfig struct{}

func (MqConfig) URI() string {
	return os.Getenv("KAFKA_SERVER_URL")
}

func (MqConfig) SecurityProtocol() string {
	return os.Getenv("KAFKA_SECURITY_PROTOCOL")
}

func (MqConfig) SSLKeyFile() string {
	return os.Getenv("KAFKA_SSL_KEY_FILE")
}

func (MqConfig) SSLCAFile() string {
	return os.Getenv("KAFKA_SSL_CA_FILE")
}

func (MqConfig) SSLCertFile() string {
	return os.Getenv("KAFKA_SSL_CERT_FILE")
}

func NewMqConfig() config.IMQConfig {
	re := regexp.MustCompile(`^(.*` + projectDirName + `)`)
	cwd, _ := os.Getwd()
	rootPath := re.Find([]byte(cwd))

	err := godotenv.Load(string(rootPath) + `/mock/.env`)

	if err != nil {
		fmt.Println("Load Env Failed ")
	}

	return &MqConfig{}
}
