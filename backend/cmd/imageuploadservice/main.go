package main

import (
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/images"
	"smlcloudplatform/pkg/microservice"
	"time"
)

func main() {

	cfg := config.NewConfig()
	ms, err := microservice.NewMicroservice(cfg)
	if err != nil {
		panic(err)
	}

	publicPath := []string{
		"/images*",
		"/productimage",
	}

	cacher := ms.Cacher(cfg.CacherConfig())
	authService := microservice.NewAuthService(cacher, 24*3*time.Hour, 24*30*time.Hour, ms.MongoPersister(cfg.MongoPersisterConfig()))
	ms.HttpMiddleware(authService.MWFuncWithRedis(cacher, publicPath...))

	filePersister := microservice.NewFilePersister()
	imagePersister := microservice.NewPersisterImage(filePersister)

	imageHttp := images.NewImagesHttp(ms, cfg, imagePersister)
	imageHttp.RegisterHttp()

	ms.Start()
}
