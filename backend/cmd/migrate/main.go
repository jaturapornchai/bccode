package main

import (
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/migration"
	"smlcloudplatform/pkg/microservice"
)

func main() {
	cfg := config.NewConfig()
	ms, err := microservice.NewMicroservice(cfg)
	if err != nil {
		panic(err)
	}

	migration.StartMigrateModel(ms, cfg)
}
