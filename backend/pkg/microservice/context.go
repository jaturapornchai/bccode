package microservice

import (
	"mime/multipart"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/pkg/microservice/models"

	"github.com/labstack/echo/v4"
)

type IContext interface {
	Log(message string)
	UserInfo() models.UserInfo
	Header(attribute string) string
	RealIp() string
	Param(name string) string
	QueryParam(name string) string
	ReadInput() string
	Response(responseCode int, responseData interface{})
	ResponseError(responseCode int, errorMessage string)
	Validate(model interface{}) error
	FormFile(field string) (*multipart.FileHeader, error)
	FormValue(field string) string
	ResponseWriter() http.ResponseWriter
	Request() *http.Request

	Persister(cfg config.IPersisterConfig) IPersister
	Cacher(cacherConfig config.ICacherConfig) ICacher
	Producer(servers config.IMQConfig) IProducer
	MQ(servers config.IMQConfig) IMQ

	EchoContext() echo.Context
}
