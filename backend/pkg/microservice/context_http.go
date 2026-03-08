package microservice

import (
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/pkg/microservice/models"

	"github.com/labstack/echo/v4"
)

// HTTPContext implement IContext it is context for HTTP
type HTTPContext struct {
	ms *Microservice
	c  echo.Context
}

// NewHTTPContext is the constructor function for HTTPContext
func NewHTTPContext(ms *Microservice, c echo.Context) *HTTPContext {

	return &HTTPContext{
		ms: ms,
		c:  c,
	}
}

// Log will log a message
func (ctx *HTTPContext) Log(message string) {
	fmt.Println("HTTP: ", message)
}

// Param return parameter by name
func (ctx *HTTPContext) Param(name string) string {
	return ctx.c.Param(name)
}

// QueryParam return query param
func (ctx *HTTPContext) QueryParam(name string) string {
	return ctx.c.QueryParam(name)
}

// ReadInput read the request body and return it as string
func (ctx *HTTPContext) ReadInput() string {
	body, err := io.ReadAll(ctx.c.Request().Body)
	if err != nil {
		return ""
	}
	return string(body)
}

// Header return header value by key
func (ctx *HTTPContext) Header(attribute string) string {
	return ctx.c.Request().Header.Get(attribute)
}

func (ctx *HTTPContext) RealIp() string {
	return ctx.c.RealIP()
}

func (ctx *HTTPContext) FormFile(attribute string) (*multipart.FileHeader, error) {
	return ctx.c.FormFile(attribute)
}

func (ctx *HTTPContext) FormValue(attribute string) string {
	return ctx.c.FormValue(attribute)
}

func (ctx *HTTPContext) UserInfo() models.UserInfo {
	userInfo := ctx.c.Get("UserInfo")
	if userInfo == nil {
		return models.UserInfo{}
	}
	return userInfo.(models.UserInfo)

}

func (ctx *HTTPContext) Validate(model interface{}) error {
	if err := ctx.c.Validate(model); err != nil {
		return err
	}
	return nil
}

func (ctx *HTTPContext) Response(responseCode int, responseData interface{}) {
	ctx.c.JSON(responseCode, responseData)
}

func (ctx *HTTPContext) ResponseS(responseCode int, responseData string) {
	ctx.c.String(responseCode, responseData)
}

func (ctx *HTTPContext) ResponseError(responseCode int, errorMessage string) {
	ctx.c.JSON(responseCode, map[string]interface{}{
		"success": false,
		"message": errorMessage,
	})
}

// Persister return perister
func (ctx *HTTPContext) Persister(cfg config.IPersisterConfig) IPersister {
	return ctx.ms.Persister(cfg)
}

// Cacher return cacher
func (ctx *HTTPContext) Cacher(cacheConfig config.ICacherConfig) ICacher {
	return ctx.ms.Cacher(cacheConfig)
}

// Producer return producer
func (ctx *HTTPContext) Producer(mqConfig config.IMQConfig) IProducer {
	return ctx.ms.Producer(mqConfig)
}

// MQ return MQ
func (ctx *HTTPContext) MQ(mqConfig config.IMQConfig) IMQ {
	return NewMQ(mqConfig, ctx.ms.Logger)
}

func (ctx *HTTPContext) ResponseWriter() http.ResponseWriter {
	return ctx.c.Response()
}

func (ctx *HTTPContext) Request() *http.Request {
	return ctx.c.Request()
}

func (ctx *HTTPContext) Bind(obj any) {
	ctx.c.Bind(obj)
}

func (ctx *HTTPContext) EchoContext() echo.Context {
	return ctx.c
}
