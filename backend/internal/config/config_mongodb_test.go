package config_test

import (
	"fmt"
	"smlcloudplatform/internal/config"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMongodbConfigWithSSL(t *testing.T) {

	giveProtocal := "mongodb"
	giveServer := "demo-mongo-server"
	givePort := "00000"
	giveUser := "mongo-user"
	givePassword := "mongo-password"
	giveDBName := "mongodb-db"
	giveSSLMode := "true"
	giveCAFile := "/cert/ca.cert"

	t.Setenv("MONGODB_PROTOCAL", giveProtocal)
	t.Setenv("MONGODB_SERVER", giveServer)
	t.Setenv("MONGODB_PORT", givePort)
	t.Setenv("MONGODB_USERNAME", giveUser)
	t.Setenv("MONGODB_PASSWORD", givePassword)
	t.Setenv("MONGODB_DB", giveDBName)
	t.Setenv("MONGODB_SSL", giveSSLMode)
	t.Setenv("MONGODB_TLS_CA_FILE", giveCAFile)

	mongoConfig := &config.MongoPersisterConfig{}

	want := fmt.Sprintf(
		"%s://%s:%s@%s:%s/?ssl=%s&tlsCAFile=%s",
		giveProtocal,
		giveUser,
		givePassword,
		giveServer,
		givePort,
		giveSSLMode,
		giveCAFile,
	)
	assert.Equal(t, giveProtocal, mongoConfig.MongodbProtocal())
	assert.Equal(t, want, mongoConfig.MongodbURI())
}

func TestMongodbSVCConfigWithSSL(t *testing.T) {

	giveProtocal := "mongodb"
	giveServer := "demo-mongo-server"
	giveUser := "mongo-user"
	givePassword := "mongo-password"
	giveDBName := "mongodb-db"
	giveSSLMode := "true"
	giveCAFile := "/cert/ca.cert"

	t.Setenv("MONGODB_PROTOCAL", giveProtocal)
	t.Setenv("MONGODB_SERVER", giveServer)
	t.Setenv("MONGODB_USERNAME", giveUser)
	t.Setenv("MONGODB_PASSWORD", givePassword)
	t.Setenv("MONGODB_DB", giveDBName)
	t.Setenv("MONGODB_SSL", giveSSLMode)
	t.Setenv("MONGODB_TLS_CA_FILE", giveCAFile)

	mongoConfig := &config.MongoPersisterConfig{}

	want := fmt.Sprintf(
		"%s://%s:%s@%s/?ssl=%s&tlsCAFile=%s",
		giveProtocal,
		giveUser,
		givePassword,
		giveServer,
		giveSSLMode,
		giveCAFile,
	)
	assert.Equal(t, giveProtocal, mongoConfig.MongodbProtocal())
	assert.Equal(t, want, mongoConfig.MongodbURI())
}
