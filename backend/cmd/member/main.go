package main

import (
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/member"
	"smlcloudplatform/internal/shop"
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
	authService := microservice.NewAuthServicePrefix("linemember:", "linememberrefresh:", cacher, 24*3*time.Hour, 24*30*time.Hour)

	publicPath := []string{
		"/shop/*",
		"/member/line",
		"/healthz",
	}

	exceptShopPath := []string{}
	ms.HttpPreRemoveTrailingSlash()

	ms.HttpMiddleware(authService.MWFuncWithRedisMixShop(cacher, exceptShopPath, publicPath...))

	ms.RegisterLivenessProbeEndpoint("/healthz")

	member.NewMemberHttp(ms, cfg).RegisterLineHttp()
	shop.NewShopHttp(ms, cfg).RegisterHttpMember()

	ms.Start()
}
