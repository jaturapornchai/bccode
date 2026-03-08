package main

import (
	"smlcloudplatform/internal/config"
	"smlcloudplatform/pkg/microservice"
)

func main() {
	cfg := config.NewConfig()
	ms, err := microservice.NewMicroservice(cfg)
	if err != nil {
		panic(err)
	}

	ms.Start()
}
