package microservice

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"runtime"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/pkg/microservice/models"
	"strings"
	"time"

	"github.com/labstack/echo/v4"
)

// ConsumerContext implement IContext it is context for Consumer
type ConsumerContext struct {
	ms      *Microservice
	message string
}

// NewConsumerContext is the constructor function for ConsumerContext
func NewConsumerContext(ms *Microservice, message string) *ConsumerContext {
	return &ConsumerContext{
		ms:      ms,
		message: message,
	}
}

// Log will log a message
func (ctx *ConsumerContext) Log(message string) {
	_, fn, line, _ := runtime.Caller(1)
	fns := strings.Split(fn, "/")
	fmt.Println("Consumer:", fns[len(fns)-1], line, message)
}

// Param return parameter by name (empty in case of Consumer)
func (ctx *ConsumerContext) Param(name string) string {
	return ""
}

// QueryParam return empty in consumer
func (ctx *ConsumerContext) QueryParam(name string) string {
	return ""
}

// ReadInput return message
func (ctx *ConsumerContext) ReadInput() string {
	return ctx.message
}

// ReadInputs return nil in case Consumer
func (ctx *ConsumerContext) ReadInputs() []string {
	return nil
}

func (ctx *ConsumerContext) Validate(model interface{}) error {
	return nil
}

// Response return response to client
func (ctx *ConsumerContext) Response(responseCode int, responseData interface{}) {

}

func (ctx *ConsumerContext) ResponseS(responseCode int, responseData string) {

}

func (ctx *ConsumerContext) ResponseError(responseCode int, errorMessage string) {

}

// Header return header value by key
func (ctx *ConsumerContext) Header(attribute string) string {
	return ""
}

func (ctx *ConsumerContext) RealIp() string {
	return ""
}

func (ctx *ConsumerContext) FormFile(attribute string) (*multipart.FileHeader, error) {
	return nil, nil
}

func (ctx *ConsumerContext) FormValue(attribute string) string {
	return ""
}

func (ctx *ConsumerContext) UserInfo() models.UserInfo {
	return models.UserInfo{}

}

// Persister return perister
func (ctx *ConsumerContext) Persister(cfg config.IPersisterConfig) IPersister {
	return ctx.ms.Persister(cfg)
}

// Now return now
func (ctx *ConsumerContext) Now() time.Time {
	return time.Now()
}

// Cacher return cacher
func (ctx *ConsumerContext) Cacher(cacheConfig config.ICacherConfig) ICacher {
	return ctx.ms.Cacher(cacheConfig)
}

// Producer return producer
func (ctx *ConsumerContext) Producer(mqConfig config.IMQConfig) IProducer {
	return ctx.ms.Producer(mqConfig)
}

// MQ return MQ
func (ctx *ConsumerContext) MQ(mqConfig config.IMQConfig) IMQ {
	return NewMQ(mqConfig, ctx.ms.Logger)
}

func (ctx *ConsumerContext) ResponseWriter() http.ResponseWriter {
	return nil
}

func (ctx *ConsumerContext) Request() *http.Request {
	return nil
}

func (ctx *ConsumerContext) EchoContext() echo.Context {
	return nil
}
