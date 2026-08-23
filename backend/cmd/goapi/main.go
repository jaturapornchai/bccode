package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"smlcloudplatform/internal/config"
	goapi "smlcloudplatform/internal/goapi"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/pkg/microservice"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	logger.Info("เริ่มต้นแอปพลิเคชัน GoAPI (standalone)...")

	// Initialize all goapi services
	s := goapi.New()
	if err := s.Init(); err != nil {
		logger.Error("GoAPI init failed: %v", err)
		os.Exit(1)
	}
	defer s.Shutdown()

	// Create Echo instance
	e := echo.New()

	// Standalone-specific middleware (CORS, Request Logging)
	corsOrigins := os.Getenv("CORS_ALLOWED_ORIGINS")
	allowedOrigins := []string{"https://bcaicloud.com", "https://*.bcaicloud.com"}
	if corsOrigins != "" {
		allowedOrigins = nil
		for _, origin := range strings.Split(corsOrigins, ",") {
			origin = strings.TrimSpace(origin)
			if origin != "" {
				allowedOrigins = append(allowedOrigins, origin)
			}
		}
	}
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     allowedOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete},
		AllowHeaders:     []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
		AllowCredentials: true,
		MaxAge:           86400,
	}))
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}, latency=${latency_human}, " +
			"remote_ip=${remote_ip}, user_agent=${user_agent}\n",
		CustomTimeFormat: "2006-01-02 15:04:05",
	}))

	// Register goapi middleware + routes (no prefix in standalone mode)
	root := e.Group("")
	s.RegisterMiddleware(root)
	cfg := config.NewConfig()
	s.RegisterRoutes(root, "", microservice.NewPersisterMongo(cfg.MongoPersisterConfig()))

	logger.Success("GoAPI routes registered (standalone mode)")

	// Start HTTP server with graceful shutdown
	port := os.Getenv("PORT")
	if port == "" {
		port = "8888"
	}

	exitHTTP := make(chan bool, 1)
	go func() {
		<-exitHTTP
		e.Shutdown(context.Background())
	}()

	go func() {
		srv := &http.Server{
			Addr:         ":" + port,
			ReadTimeout:  60 * time.Second,
			WriteTimeout: 300 * time.Second,
			IdleTimeout:  120 * time.Second,
		}
		logger.Info("http server เริ่มทำงานที่ [::]:%s", port)
		if err := e.StartServer(srv); err != nil && err != http.ErrServerClosed {
			logger.Error("ล้มเหลวในการเริ่มต้น server: %v", err)
		}
	}()

	// Wait for OS signal
	osQuit := make(chan os.Signal, 1)
	signal.Notify(osQuit, syscall.SIGTERM, syscall.SIGINT)
	<-osQuit

	logger.Info("ปิดระบบอย่างสุภาพ...")
	exitHTTP <- true

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		logger.Error("Server ถูกบังคับให้ปิด: %v", err)
	}

	logger.Success("GoAPI standalone ปิดระบบเรียบร้อย")
}
