package main

import (
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/shop"
	"smlcloudplatform/pkg/microservice"
	"time"
)

func main() {
	// err := godotenv.Load()
	// if err != nil {
	// 	log.Fatal("Error loading .env file")
	// }

	cfg := config.NewConfig()
	ms, err := microservice.NewMicroservice(cfg)
	if err != nil {
		panic(err)
	}

	cacher := ms.Cacher(cfg.CacherConfig())
	authService := microservice.NewAuthService(cacher, 24*3*time.Hour, 24*30*time.Hour)

	ms.HttpMiddleware(authService.MWFuncWithShop(cacher))

	svc := shop.NewShopHttp(ms, cfg)

	svc.RegisterHttp()

	//ms.Echo().GET("/swagger/*", echoSwagger.WrapHandler)

	ms.Start()
}
