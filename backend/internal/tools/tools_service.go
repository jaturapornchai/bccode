package tools

import (
	"os"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type ToolsService struct {
	ms  *microservice.Microservice
	cfg config.IConfig
}

func NewToolsService(ms *microservice.Microservice, cfg config.IConfig) *ToolsService {

	return &ToolsService{
		ms:  ms,
		cfg: cfg,
	}
}

func (svc *ToolsService) RegisterHttp() {
	svc.ms.GET("tool/mongo", svc.CheckMongodbConnect)
	svc.ms.GET("tool/env", svc.AllEnv)
}

func (svc *ToolsService) CheckMongodbConnect(ctx microservice.IContext) error {
	mongoCfg := svc.cfg.MongoPersisterConfig()
	repStr := mongoCfg.MongodbURI()

	repStr = repStr + " \n " + mongoCfg.DB()

	ctx.Response(200, repStr)

	return nil
}

func (svc *ToolsService) AllEnv(ctx microservice.IContext) error {

	ctx.Response(200, os.Environ())

	return nil
}

func (svc *ToolsService) MockAuth(ctx microservice.IContext) error {
	cacher := svc.ms.Cacher(svc.cfg.CacherConfig())

	profiles := []struct {
		AuthKey  string
		Username string
		Name     string
		ShopID   string
		Role     int8
	}{
		{
			AuthKey:  "57b740fdb4b3e64b0da0ab5f0f677df02b44d0bf2468f9ad65e2590aa82f9142",
			Username: "error404",
			Name:     "Error Shop",
			ShopID:   "2Gf5cN6DP1kX7TYq3EJ1m4DKsJC",
			Role:     2,
		},
	}

	curProfile := profiles[0]

	cacheKey := "auth-" + curProfile.AuthKey
	cacher.HMSet(cacheKey, map[string]interface{}{
		"username": curProfile.Username,
		"name":     curProfile.Name,
		"shopid":   curProfile.ShopID,
		"role":     curProfile.Role,
	})

	cacher.Expire(cacheKey, time.Hour*168) // 7 days
	return nil
}
