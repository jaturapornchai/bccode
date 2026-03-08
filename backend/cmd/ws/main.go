package main

import (
	"log"
	"net/http"
	_ "net/http/pprof"
	"smlcloudplatform/internal/authentication"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/vfgl/journal"
	"smlcloudplatform/pkg/microservice"
	"time"
)

func main() {

	cfg := config.NewConfig()
	ms, err := microservice.NewMicroservice(cfg)
	if err != nil {
		panic(err)
	}

	go func() {
		log.Println(http.ListenAndServe(":6060", nil))
	}()

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
		"/employee/login",
		"/healthz",
		"/metrics",
	}

	ms.HttpPreRemoveTrailingSlash()
	ms.HttpUsePrometheus()
	ms.HttpUseJaeger()

	ms.HttpMiddleware(authService.MWFuncWithRedis(cacher, publicPath...))

	ms.RegisterLivenessProbeEndpoint("/healthz")

	authHttp := authentication.NewAuthenticationHttp(ms, cfg)
	authHttp.RegisterHttp()

	journalWs := journal.NewJournalWs(ms, cfg)
	journalWs.RegisterHttp()

	ms.Start()
}
