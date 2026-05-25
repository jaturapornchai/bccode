package microservice

import (
	"context"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

type IMicroserviceHTTP interface {
	RegisterHttp()
}

// GET register service endpoint for HTTP GET
func (ms *Microservice) GET(path string, h ServiceHandleFunc, m ...echo.MiddlewareFunc) {
	trimmedPath := strings.TrimSpace(path)
	if !strings.HasPrefix(trimmedPath, "/") {
		trimmedPath = "/" + trimmedPath
	}

	legacyPath := ms.pathPrefix + trimmedPath
	ms.Logger.Debugf("Register HTTP Handler GET \"%s\".", legacyPath)
	ms.echo.GET(legacyPath, func(c echo.Context) error {
		return h(NewHTTPContext(ms, c))
	}, m...)

	if !strings.HasPrefix(trimmedPath, "/v1/") && trimmedPath != "/v1" &&
		!strings.HasPrefix(trimmedPath, "/api/v1/") && trimmedPath != "/api/v1" {
		v1Path := ms.pathPrefix + "/v1" + trimmedPath
		if v1Path != legacyPath {
			ms.Logger.Debugf("Register HTTP Handler GET (V1 Alias) \"%s\".", v1Path)
			ms.echo.GET(v1Path, func(c echo.Context) error {
				return h(NewHTTPContext(ms, c))
			}, m...)
		}
	}
}

// POST register service endpoint for HTTP POST
func (ms *Microservice) POST(path string, h ServiceHandleFunc, m ...echo.MiddlewareFunc) {
	trimmedPath := strings.TrimSpace(path)
	if !strings.HasPrefix(trimmedPath, "/") {
		trimmedPath = "/" + trimmedPath
	}

	legacyPath := ms.pathPrefix + trimmedPath
	ms.Logger.Debugf("Register HTTP Handler POST \"%s\".", legacyPath)
	ms.echo.POST(legacyPath, func(c echo.Context) error {
		return h(NewHTTPContext(ms, c))
	}, m...)

	if !strings.HasPrefix(trimmedPath, "/v1/") && trimmedPath != "/v1" &&
		!strings.HasPrefix(trimmedPath, "/api/v1/") && trimmedPath != "/api/v1" {
		v1Path := ms.pathPrefix + "/v1" + trimmedPath
		if v1Path != legacyPath {
			ms.Logger.Debugf("Register HTTP Handler POST (V1 Alias) \"%s\".", v1Path)
			ms.echo.POST(v1Path, func(c echo.Context) error {
				return h(NewHTTPContext(ms, c))
			}, m...)
		}
	}
}

// PUT register service endpoint for HTTP PUT
func (ms *Microservice) PUT(path string, h ServiceHandleFunc, m ...echo.MiddlewareFunc) {
	trimmedPath := strings.TrimSpace(path)
	if !strings.HasPrefix(trimmedPath, "/") {
		trimmedPath = "/" + trimmedPath
	}

	legacyPath := ms.pathPrefix + trimmedPath
	ms.Logger.Debugf("Register HTTP Handler PUT \"%s\".", legacyPath)
	ms.echo.PUT(legacyPath, func(c echo.Context) error {
		return h(NewHTTPContext(ms, c))
	}, m...)

	if !strings.HasPrefix(trimmedPath, "/v1/") && trimmedPath != "/v1" &&
		!strings.HasPrefix(trimmedPath, "/api/v1/") && trimmedPath != "/api/v1" {
		v1Path := ms.pathPrefix + "/v1" + trimmedPath
		if v1Path != legacyPath {
			ms.Logger.Debugf("Register HTTP Handler PUT (V1 Alias) \"%s\".", v1Path)
			ms.echo.PUT(v1Path, func(c echo.Context) error {
				return h(NewHTTPContext(ms, c))
			}, m...)
		}
	}
}

// PATCH register service endpoint for HTTP PATCH
func (ms *Microservice) PATCH(path string, h ServiceHandleFunc, m ...echo.MiddlewareFunc) {
	trimmedPath := strings.TrimSpace(path)
	if !strings.HasPrefix(trimmedPath, "/") {
		trimmedPath = "/" + trimmedPath
	}

	legacyPath := ms.pathPrefix + trimmedPath
	ms.Logger.Debugf("Register HTTP Handler PATCH \"%s\".", legacyPath)
	ms.echo.PATCH(legacyPath, func(c echo.Context) error {
		return h(NewHTTPContext(ms, c))
	}, m...)

	if !strings.HasPrefix(trimmedPath, "/v1/") && trimmedPath != "/v1" &&
		!strings.HasPrefix(trimmedPath, "/api/v1/") && trimmedPath != "/api/v1" {
		v1Path := ms.pathPrefix + "/v1" + trimmedPath
		if v1Path != legacyPath {
			ms.Logger.Debugf("Register HTTP Handler PATCH (V1 Alias) \"%s\".", v1Path)
			ms.echo.PATCH(v1Path, func(c echo.Context) error {
				return h(NewHTTPContext(ms, c))
			}, m...)
		}
	}
}

// DELETE register service endpoint for HTTP DELETE
func (ms *Microservice) DELETE(path string, h ServiceHandleFunc, m ...echo.MiddlewareFunc) {
	trimmedPath := strings.TrimSpace(path)
	if !strings.HasPrefix(trimmedPath, "/") {
		trimmedPath = "/" + trimmedPath
	}

	legacyPath := ms.pathPrefix + trimmedPath
	ms.Logger.Debugf("Register HTTP Handler DELETE \"%s\".", legacyPath)
	ms.echo.DELETE(legacyPath, func(c echo.Context) error {
		return h(NewHTTPContext(ms, c))
	}, m...)

	if !strings.HasPrefix(trimmedPath, "/v1/") && trimmedPath != "/v1" &&
		!strings.HasPrefix(trimmedPath, "/api/v1/") && trimmedPath != "/api/v1" {
		v1Path := ms.pathPrefix + "/v1" + trimmedPath
		if v1Path != legacyPath {
			ms.Logger.Debugf("Register HTTP Handler DELETE (V1 Alias) \"%s\".", v1Path)
			ms.echo.DELETE(v1Path, func(c echo.Context) error {
				return h(NewHTTPContext(ms, c))
			}, m...)
		}
	}
}

// startHTTP will start HTTP service, this function will block thread
func (ms *Microservice) startHTTP(exitChannel chan bool) error {

	ms.echo.Use(ms.middlewareManager.RequestLoggerMiddleware)

	port := ms.config.HttpConfig().Port()
	// Caller can exit by sending value to exitChannel
	go func() {
		<-exitChannel
		ms.stopHTTP()
	}()

	ms.Logger.Infof("Listening: %v Entrypoint: %v ", port, ms.pathPrefix)

	err := ms.echo.Start("0.0.0.0:" + port)
	if err == nil {
		ms.Logger.Error("Failed After Start", err)
	}

	return err
}

func (ms *Microservice) stopHTTP() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ms.echo.Shutdown(ctx)
}

func (ms *Microservice) RegisterHttp(http IMicroserviceHTTP) {
	http.RegisterHttp()
}
