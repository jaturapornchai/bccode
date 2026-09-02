package main

import (
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

	cacher := ms.Cacher(cfg.CacherConfig())
	// jwtService := microservice.NewJwtService(cacher, cfg.JwtSecretKey(), 24*3)
	authService := microservice.NewAuthService(cacher, 24*3*time.Hour, 24*30*time.Hour, ms.MongoPersister(cfg.MongoPersisterConfig()))

	publicPath := []string{
		"/login",
		"/dev-login",
		"/demo-login",
		"/googlelogin",
		"/refresh",
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
