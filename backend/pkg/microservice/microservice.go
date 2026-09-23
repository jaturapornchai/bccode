package microservice

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/logger"
	"smlcloudplatform/internal/middlewares"
	msValidator "smlcloudplatform/pkg/validator"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo-contrib/jaegertracing"
	"github.com/labstack/echo-contrib/prometheus"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type IMicroservice interface {
	Start() error
	Cleanup() error
	Log(tag string, message string)

	// HTTP Services
	HttpMiddleware(middleware ...echo.MiddlewareFunc)
	GET(path string, h ServiceHandleFunc, m ...echo.MiddlewareFunc)
	POST(path string, h ServiceHandleFunc, m ...echo.MiddlewareFunc)
	PUT(path string, h ServiceHandleFunc, m ...echo.MiddlewareFunc)
	PATCH(path string, h ServiceHandleFunc, m ...echo.MiddlewareFunc)
	DELETE(path string, h ServiceHandleFunc, m ...echo.MiddlewareFunc)

	TimeNow() func() time.Time

	// CRUD(cfg IConfig, pathName string, modelx GenCrud)
	ECHO() *echo.Echo
}

type Microservice struct {
	backgroundWorkers         []func(context.Context)
	echo                      *echo.Echo
	exitChannel               chan bool
	cacher                    ICacher
	persisters                map[string]IPersister
	persistersMutex           sync.Mutex
	websocketPool             *WebsocketPool
	pathPrefix                string
	config                    config.IConfig
	jaegerCloser              io.Closer
	Logger                    logger.ILogger
	Mode                      string
	middlewareManager         middlewares.IMiddlewareManager
}

type ServiceHandleFunc func(context IContext) error

func NewMicroservice(config config.IConfig) (*Microservice, error) {

	logger := logger.NewAppLogger(config.LoggerConfig())
	logger.InitLogger()

	e := echo.New()
	e.HideBanner = true
	e.HidePort = true
	// e.Validator = &msValidator.CustomValidator{Validator: validator.New()}
	e.Validator = msValidator.NewCustomValidator()

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"*"},
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept, echo.HeaderAuthorization},
	}))

	websocketPool := WebsocketPool{
		Handler: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		Connections: map[string]*websocket.Conn{},
	}

	m := &Microservice{
		echo:                 e,
		persisters:           map[string]IPersister{},
		pathPrefix:           config.PathPrefix(),
		config:               config,
		Logger:               logger,
		Mode:                 os.Getenv("MODE"),
		websocketPool:        &websocketPool,
	}

	// Init logger
	m.middlewareManager = middlewares.NewMiddlewareManager(logger, config, m.getHttpMetricsCb())

	m.Logger.Info("Initial Microservice.")
	err := m.CheckReadyToStart()
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (ms *Microservice) CheckReadyToStart() error {
	// check resources availability

	// postgresql
	persisterConfig := ms.config.PersisterConfig()
	if persisterConfig.Host() != "" {
		// TEST Connetion PostgreSQL
		ms.Logger.Debug("[PostgreSQL]Test Connection.")
		postgresqlPst := ms.Persister(persisterConfig)
		err := postgresqlPst.TestConnect()
		if err != nil {
			ms.Logger.Error("[PostgreSQL]Connection Failed.", err)
			return err
		}
	}

	return nil
}

func maskConnectionString(raw string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		if strings.Contains(raw, "@") {
			return "***"
		}
		return raw
	}
	if parsed.User == nil {
		return raw
	}
	username := parsed.User.Username()
	if username == "" {
		parsed.User = url.User("***")
		return parsed.String()
	}
	if _, hasPassword := parsed.User.Password(); hasPassword {
		parsed.User = url.UserPassword(username, "***")
	} else {
		parsed.User = url.User("***")
	}
	return parsed.String()
}

// RegisterBackgroundWorker registers a worker before Start. Workers must stop
// when ctx is cancelled; resources remain open until all workers have stopped.
func (ms *Microservice) RegisterBackgroundWorker(worker func(context.Context)) {
	ms.backgroundWorkers = append(ms.backgroundWorkers, worker)
}

// Start start all registered services
func (ms *Microservice) Start() error {
	workerCtx, cancelWorkers := context.WithCancel(context.Background())
	var workers sync.WaitGroup
	for _, worker := range ms.backgroundWorkers {
		workers.Add(1)
		go func(run func(context.Context)) {
			defer workers.Done()
			run(workerCtx)
		}(worker)
	}
	defer func() {
		cancelWorkers()
		workers.Wait()
		ms.Cleanup()
	}()

	ms.Logger.Debugf("Start App: %s Mode: %s", ms.config.ApplicationName(), ms.Mode)

	// if ms.Mode == "development" {
	// 	// register swagger api spec
	// 	ms.echo.Static("/swagger/doc.json", "./../../api/swagger/swagger.json")
	// }
	httpN := len(ms.echo.Routes())
	var exitHTTP chan bool
	if httpN > 0 {
		exitHTTP = make(chan bool, 1)
		go func() {
			ms.startHTTP(exitHTTP)
		}()

	}

	// There are 2 ways to exit from Microservices
	// 1. The SigTerm can be send from outside program such as from k8s
	// 2. Send true to ms.exitChannel
	osQuit := make(chan os.Signal, 1)
	ms.exitChannel = make(chan bool, 1)
	signal.Notify(osQuit, syscall.SIGTERM, syscall.SIGINT)
	exit := false
	for {
		if exit {
			break
		}
		select {
		case <-osQuit:
			// Exit from HTTP as well
			if exitHTTP != nil {
				exitHTTP <- true
			}
			exit = true
		case <-ms.exitChannel:
			// Exit from HTTP as well
			if exitHTTP != nil {
				exitHTTP <- true
			}
			exit = true
		}
	}

	return nil
}

// Stop stop the services
func (ms *Microservice) Stop() {
	if ms.exitChannel == nil {
		return
	}
	ms.exitChannel <- true
}

// Cleanup clean resources up from every registered services before exit
func (ms *Microservice) Cleanup() error {
	ms.Logger.Info("Stop Service Cleanup System.")
	if ms.cacher != nil {
		ms.cacher.Close()
	}

	if ms.jaegerCloser != nil {
		ms.jaegerCloser.Close()
	}

	return nil
}

func (ms *Microservice) TimeNow() time.Time {
	return time.Now()
}

// Log log message to console
func (ms *Microservice) Log(tag string, message string) {
	_, fn, line, _ := runtime.Caller(1)
	fns := strings.Split(fn, "/")
	fmt.Println(tag+":", fns[len(fns)-1], line, message)

}

type tenantPersisterConfig struct {
	config.IPersisterConfig
	dbName string
}

func (t *tenantPersisterConfig) DB() string {
	return t.dbName
}

var (
	tenantModels []interface{}
	DBCheckHook  func(dbName string) error
)

func RegisterTenantModel(models ...interface{}) {
	tenantModels = append(tenantModels, models...)
}

func (ms *Microservice) PersisterTenant(cfg config.IPersisterConfig, dbName string) IPersister {
	if dbName == "" {
		return ms.Persister(cfg)
	}
	key := fmt.Sprintf("%s/%s", cfg.Host(), dbName)
	ms.persistersMutex.Lock()
	defer ms.persistersMutex.Unlock()
	pst, ok := ms.persisters[key]
	if !ok {
		if DBCheckHook != nil {
			if err := DBCheckHook(dbName); err != nil {
				ms.Logger.Errorf("PersisterTenant: DBCheckHook failed for db=%s: %v", dbName, err)
			}
		}
		tenantCfg := &tenantPersisterConfig{
			IPersisterConfig: cfg,
			dbName:           dbName,
		}
		pst = NewPersister(tenantCfg)
		if len(tenantModels) > 0 {
			if err := pst.AutoMigrate(tenantModels...); err != nil {
				ms.Logger.Errorf("PersisterTenant: AutoMigrate failed for db=%s: %v", dbName, err)
			} else {
				ms.Logger.Infof("PersisterTenant: AutoMigrate completed for db=%s", dbName)
			}
		}
		ms.persisters[key] = pst
	}
	return pst
}

func (ms *Microservice) Persister(cfg config.IPersisterConfig) IPersister {
	pst, ok := ms.persisters[cfg.Host()]
	if !ok {
		pst = NewPersister(cfg)
		ms.persistersMutex.Lock()
		ms.persisters[cfg.Host()] = pst
		ms.persistersMutex.Unlock()
	}
	return pst
}

// SetCacher - กำหนดที่เก็บ session (PostgreSQL) ที่ทุก module ใช้ร่วมกัน
func (ms *Microservice) SetCacher(cacher ICacher) {
	ms.cacher = cacher
}

// Cacher - ที่เก็บ session กลาง (ต้องเรียก SetCacher ก่อน register module)
func (ms *Microservice) Cacher() ICacher {
	if ms.cacher == nil {
		panic("microservice: cacher is not configured; call SetCacher before registering modules")
	}
	return ms.cacher
}

func (ms *Microservice) Websocket(id string, response http.ResponseWriter, request *http.Request) (*websocket.Conn, error) {
	ws, err := ms.websocketPool.Handler.Upgrade(response, request, nil)

	if err != nil {
		return nil, err
	}

	ms.websocketPool.Lock()
	ms.websocketPool.Connections[id] = ws
	ms.websocketPool.Unlock()

	return ws, nil
}

func (ms *Microservice) WebsocketClose(id string) {
	ms.websocketPool.Lock()
	ms.websocketPool.Connections[id].Close()
	delete(ms.websocketPool.Connections, id)
	ms.websocketPool.Unlock()
}

func (ms *Microservice) WebsocketCount() int {
	ms.websocketPool.Lock()

	defer ms.websocketPool.Unlock()

	return len(ms.websocketPool.Connections)
}

func (ms *Microservice) HttpMiddleware(middleware ...echo.MiddlewareFunc) {
	ms.echo.Use(middleware...)
}

func (ms *Microservice) HttpPreRemoveTrailingSlash() {
	ms.echo.Pre(middleware.RemoveTrailingSlash())
	ms.Logger.Info("Use remove trailing")
}

func (ms *Microservice) HttpUsePrometheus() {
	ms.Logger.Info("Start Prometheus.")
	p := prometheus.NewPrometheus("smlcloudplatform", nil)
	p.Use(ms.echo)
}

func (ms *Microservice) HttpUseJaeger() {
	ms.Logger.Info("Start Jaeger.")
	c := jaegertracing.New(ms.echo, nil)
	ms.jaegerCloser = c
}

func (ms *Microservice) HttpUseCors() {

	ms.echo.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: ms.config.HttpCORS(),
		AllowHeaders: []string{echo.HeaderOrigin, echo.HeaderContentType, echo.HeaderAccept},
	}))
}

func (ms *Microservice) Echo() *echo.Echo {
	return ms.echo
}
