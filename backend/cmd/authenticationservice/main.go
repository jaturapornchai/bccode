package main

import (
	"smlcloudplatform/internal/authentication"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/pkg/microservice"
	"time"
)

func main() {

	cfg := config.NewConfig()
	ms, err := microservice.NewMicroservice(cfg)
	if err != nil {
		panic(err)
	}
	svc := authentication.NewAuthenticationHttp(ms, cfg)

	cacher := ms.Cacher(cfg.CacherConfig())
	// jwtService := microservice.NewJwtService(cacher, cfg.JwtSecretKey(), 24*3)
	authService := microservice.NewAuthService(cacher, 24*3*time.Hour, 24*30*time.Hour)

	publicPath := []string{
		"/login",
		"/poslogin",
		"/register",
		"/list-shop",
		"/select-shop",
		"/create-shop",
		"/healthz",
	}

	ms.HttpMiddleware(authService.MWFuncWithShop(cacher, publicPath...))
	svc.RegisterHttp()

	// ms.Echo().GET("/swagger/*", echoSwagger.WrapHandler)

	ms.Start()
}
