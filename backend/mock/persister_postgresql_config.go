package mock

import (
	"fmt"
	"os"
	"regexp"
	"smlcloudplatform/internal/config"

	"github.com/joho/godotenv"
)

type PersisterPostgresqlConfig struct{}

func NewPersisterPostgresqlConfig() config.IPersisterConfig {

	env := os.Getenv("MODE")
	if env == "" {
		os.Setenv("MODE", "development")
		env = "development"
	}

	re := regexp.MustCompile(`^(.*` + projectDirName + `)`)
	cwd, _ := os.Getwd()
	rootPath := re.Find([]byte(cwd))
	basePath := string(rootPath)

	godotenv.Load(basePath + "/.env." + env + ".test.local")
	err := godotenv.Load(basePath + `/mock/.env`)
	if err != nil {
		fmt.Println("Load Env Failed ")
	}

	return &PersisterPostgresqlConfig{}
}

func (cfg *PersisterPostgresqlConfig) Host() string {
	return os.Getenv("POSTGRES_HOST")
}

func (cfg *PersisterPostgresqlConfig) Port() string {
	return os.Getenv("POSTGRES_PORT")
}

func (cfg *PersisterPostgresqlConfig) DB() string {
	return os.Getenv("POSTGRES_DB_NAME")
}

func (cfg *PersisterPostgresqlConfig) Username() string {
	return os.Getenv("POSTGRES_USERNAME")
}

func (cfg *PersisterPostgresqlConfig) Password() string {
	return os.Getenv("POSTGRES_PASSWORD")
}

func (cfg *PersisterPostgresqlConfig) SSLMode() string {
	sslMode := os.Getenv("POSTGRES_SSL_MODE")
	if sslMode != "" {
		sslMode = "disable"
	}
	return sslMode
}

func (cfg *PersisterPostgresqlConfig) TimeZone() string {
	return "Asia/Bangkok"
}

func (cfg *PersisterPostgresqlConfig) LoggerLevel() string {
	return "debug"
}
