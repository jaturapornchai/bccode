package main

import (
	"context"
	"crypto/subtle"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"smlcloudplatform/internal/authentication"
	"smlcloudplatform/internal/config"
	fahttp "smlcloudplatform/internal/fixedasset/httpapi"
	glhttp "smlcloudplatform/internal/generalledger/httpapi"
	goapi "smlcloudplatform/internal/goapi"
	goapi_handlers "smlcloudplatform/internal/goapi/handlers"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/internal/logger"
	"smlcloudplatform/internal/mcptoken"
	"smlcloudplatform/internal/media"
	"smlcloudplatform/internal/organization/branch"
	"smlcloudplatform/internal/organization/businesstype"
	"smlcloudplatform/internal/organization/company"
	"smlcloudplatform/internal/organization/rolepermission"
	"smlcloudplatform/internal/setupconfig"
	"smlcloudplatform/internal/shop"
	"smlcloudplatform/internal/shop/employee"
	"smlcloudplatform/pkg/microservice"

	"github.com/labstack/echo/v4"
)

func init() {
	time.Local = time.UTC
}

func validReloadConfigSecret(expectedSecret, suppliedSecret string) bool {
	if expectedSecret == "" || suppliedSecret == "" || len(expectedSecret) != len(suppliedSecret) {
		return false
	}
	return subtle.ConstantTimeCompare([]byte(expectedSecret), []byte(suppliedSecret)) == 1
}

// BC Ai Account API — PostgreSQL เท่านั้น (ถอด MongoDB / Kafka / Redis / ClickHouse ออกแล้ว 2026-09-23)
func main() {
	setupconfig.LoadBootstrapConfig()

	if host := os.Getenv("HOST_API"); host != "" {
		fmt.Printf("Host: %v\n", host)
	}

	cfg := config.NewConfig()
	ms, err := microservice.NewMicroservice(cfg)
	if err != nil {
		panic(err)
	}

	// session/token เก็บใน PostgreSQL ฐานควบคุมกลาง (ตาราง cache_entries) — ใช้ร่วมกันทั้ง mainapi และ goapi
	controlDB, err := mypg.PgSqlFastConnect(mcptoken.ControlDatabase)
	if err != nil {
		panic(fmt.Errorf("connect control database: %w", err))
	}
	cacher, err := microservice.NewCacher(controlDB)
	if err != nil {
		panic(err)
	}
	ms.SetCacher(cacher)
	ms.RegisterBackgroundWorker(func(ctx context.Context) {
		ticker := time.NewTicker(10 * time.Minute)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if _, err := cacher.PurgeExpired(ctx); err != nil {
					logger.GetLogger().Errorf("purge expired sessions: %v", err)
				}
			}
		}
	})

	ms.HttpUsePrometheus()
	logger.GetLogger().Info("Starting API...")

	// ตรวจสิทธิ์สด (ผู้ใช้ถูกปิด/ถูกถอดจากกลุ่มกิจการ/บริษัท-สาขาถูกปิด) จากฐานควบคุมกลางทุกคำขอ
	authService := microservice.NewAuthService(ms.Cacher(), 24*3*time.Hour, 24*30*time.Hour, controlDB)
	publicPath := []string{
		"/mcp/gl",              // Uses its own scoped MCP bearer token, never session bypass.
		"/integration/gl/v2/*", // Separate API-token authentication on each registered route.

		"/googlelogin",
		"/login",
		"/dev-login",
		"/demo-login",
		"/refresh",

		"/healthz",
		"/metrics",
		"/reload-config",

		"/goapi/*",        // GoAPI routes — bypass auth (goapi มี auth ของตัวเอง)
		"/api/language/*", // Language — public (pre-login language loading)
	}

	exceptShopPath := []string{
		"/holding",
		"/shop",
		"/logout",
		"/verify-token",
		"/profile",
		"/profile/password",
		"/profile/disable-user",
		"/sessions/active-count",
		"/list-holding",
		"/list-shop",
		"/select-holding",
		"/select-shop",
		"/create-holding",
		"/create-shop",
		"/favorite-holding",
		"/favorite-shop",
		// Holding admin management — works from the holding-selection screen (no shop selected yet);
		// the handlers resolve the caller's role per-holding from the request holdingcode.
		"/holding-member/list",
	}

	// Reload config endpoint — goapi เรียกหลัง save config เพื่อให้ mainapi ใช้ config ใหม่
	ms.Echo().POST("/reload-config", func(c echo.Context) error {
		secret := c.Request().Header.Get("X-Reload-Secret")
		if !validReloadConfigSecret(os.Getenv("RELOAD_CONFIG_SECRET"), secret) {
			return c.JSON(http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		}
		log.Println("[ReloadConfig] ได้รับ request reload config จาก goapi")
		go setupconfig.ReloadConfig()
		return c.JSON(http.StatusOK, map[string]string{"message": "กำลัง reload config..."})
	})

	ms.HttpMiddleware(authService.MWFuncMixShop(ms.Cacher(), exceptShopPath, publicPath...))
	ms.RegisterLivenessProbeEndpoint("/healthz")
	ms.HttpUseCors()
	ms.HttpPreRemoveTrailingSlash()

	serviceStartHttp(ms,
		authentication.NewAuthenticationHttp(ms, cfg),
		shop.NewShopHttp(ms, cfg),
		shop.NewShopMemberHttp(ms, cfg),
		employee.NewEmployeeHttp(ms, cfg),
		fahttp.NewHttp(ms, cfg),
		glhttp.NewHttp(ms, cfg),
		mcptoken.NewHttp(ms),
		company.NewCompanyHttp(ms, cfg),
		branch.NewBranchHttp(ms, cfg),
		businesstype.NewBusinessTypeHttp(ms, cfg),
		rolepermission.NewRolePermissionHttp(ms, cfg),
	)
	ms.RegisterHttp(media.InitMediaUploadHttp(ms, cfg))

	goapiServer := goapi.New()
	if err := goapiServer.Init(); err != nil {
		log.Printf("GoAPI init failed: %v (goapi routes disabled)", err)
	} else {
		goapiGroup := ms.Echo().Group("/goapi")
		goapiServer.RegisterMiddleware(goapiGroup)
		goapiServer.RegisterRoutes(goapiGroup, "/goapi", ms.Cacher(), controlDB)
		defer goapiServer.Shutdown()
		log.Println("GoAPI routes registered under /goapi/*")

		ms.Echo().GET("/api/language/:lang", goapi_handlers.GetLanguageHandler)
	}

	ms.Start()
}

type HttpRegister interface {
	RegisterHttp()
}

func serviceStartHttp(ms *microservice.Microservice, services ...HttpRegister) {
	for _, service := range services {
		ms.RegisterHttp(service)
	}
}
