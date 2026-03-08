package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"smlcloudplatform/internal/config"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	HTTP   = "http"
	METHOD = "method"
	URI    = "uri"
	STATUS = "status"
	SIZE   = "size"
	TIME   = "time"

	KafkaHeaders = "kafkaHeaders"
	MessageSize  = "MessageSize"
	Topic        = "topic"
	Partition    = "partition"
	Message      = "message"
	WorkerID     = "workerID"
	Headers      = "headers"
	Offset       = "offset"
	TimeStamp    = "timestamp"
)

type ILogger interface {
	InitLogger()
	Sync() error
	Debug(args ...interface{})
	Debugf(template string, args ...interface{})
	Info(args ...interface{})
	Infof(template string, args ...interface{})
	Warn(args ...interface{})
	Warnf(template string, args ...interface{})
	WarnErrMsg(msg string, err error)
	Error(args ...interface{})
	Errorf(template string, args ...interface{})
	Err(msg string, err error)
	DPanic(args ...interface{})
	DPanicf(template string, args ...interface{})
	Fatal(args ...interface{})
	Fatalf(template string, args ...interface{})
	Printf(template string, args ...interface{})

	HttpMiddlewareAccessLogger(method string, uri string, status int, size int64, time time.Duration)
	KafkaProcessMessage(topic string, partition int, message []byte, workerID int, offset int64, time time.Time)
	KafkaProcessMessageWithHeaders(topic string, partition int, message []byte, workerID int, offset int64, time time.Time, headers map[string]interface{})
}

// For mapping config logger to email_service logger levels
var loggerLevelMap = map[string]zapcore.Level{
	"debug":  zapcore.DebugLevel,
	"info":   zapcore.InfoLevel,
	"warn":   zapcore.WarnLevel,
	"error":  zapcore.ErrorLevel,
	"dpanic": zapcore.DPanicLevel,
	"panic":  zapcore.PanicLevel,
	"fatal":  zapcore.FatalLevel,
}

type AppLogger struct {
	level       string
	devMode     bool
	encoder     string
	sugarLogger *zap.SugaredLogger
	logger      *zap.Logger
}

var (
	instance *AppLogger
	syncOnce sync.Once
)

func GetLogger() ILogger {
	return instance
}

func NewAppLogger(cfg config.ILoggerConfig) *AppLogger {
	syncOnce.Do(func() {
		instance = &AppLogger{
			level:   cfg.LogLevel(),
			devMode: cfg.DevMode(),
			encoder: cfg.Encoder(),
		}

		instance.InitLogger()
	})

	return instance
}

func (l *AppLogger) getLoggerLevel() zapcore.Level {
	level, exist := loggerLevelMap[l.level]
	if !exist {
		return zapcore.DebugLevel
	}

	return level
}

func (l *AppLogger) InitLogger() {
	logLevel := l.getLoggerLevel()

	// Dev mode: เขียน log ทั้ง console และ file ให้ AI อ่านได้
	var logWriter zapcore.WriteSyncer
	if l.devMode {
		fileWriter := setupMainapiLogFile()
		if fileWriter != nil {
			logWriter = zapcore.NewMultiWriteSyncer(
				zapcore.AddSync(os.Stdout),
				zapcore.AddSync(fileWriter),
			)
		} else {
			logWriter = zapcore.AddSync(os.Stdout)
		}
	} else {
		logWriter = zapcore.AddSync(os.Stdout)
	}

	var encoderCfg zapcore.EncoderConfig
	if l.devMode {
		encoderCfg = zap.NewDevelopmentEncoderConfig()
	} else {
		encoderCfg = zap.NewProductionEncoderConfig()
	}

	var encoder zapcore.Encoder
	encoderCfg.NameKey = "service"
	encoderCfg.TimeKey = "timestamp"
	encoderCfg.LevelKey = "level"
	encoderCfg.CallerKey = "line"
	encoderCfg.MessageKey = "message"
	encoderCfg.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderCfg.EncodeLevel = zapcore.CapitalLevelEncoder
	encoderCfg.EncodeCaller = zapcore.ShortCallerEncoder
	encoderCfg.EncodeDuration = zapcore.StringDurationEncoder

	if l.encoder == "console" {
		encoderCfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		encoderCfg.EncodeCaller = zapcore.FullCallerEncoder
		encoderCfg.ConsoleSeparator = " | "
		encoder = zapcore.NewConsoleEncoder(encoderCfg)
	} else {
		encoderCfg.FunctionKey = "caller"
		encoderCfg.EncodeName = zapcore.FullNameEncoder
		encoder = zapcore.NewJSONEncoder(encoderCfg)
	}

	core := zapcore.NewCore(encoder, logWriter, zap.NewAtomicLevelAt(logLevel))
	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))

	l.logger = logger
	l.sugarLogger = logger.Sugar()
}

// Sync flushes any buffered log entries
func (l *AppLogger) Sync() error {
	go l.logger.Sync() // nolint: errcheck
	return l.sugarLogger.Sync()
}

// Debug uses fmt.Sprint to construct and log a message.
func (l *AppLogger) Debug(args ...interface{}) {
	l.sugarLogger.Debug(args...)
}

// Debugf uses fmt.Sprintf to log a templated message
func (l *AppLogger) Debugf(template string, args ...interface{}) {
	l.sugarLogger.Debugf(template, args...)
}

// Info uses fmt.Sprint to construct and log a message
func (l *AppLogger) Info(args ...interface{}) {
	l.sugarLogger.Info(args...)
}

// Infof uses fmt.Sprintf to log a templated message.
func (l *AppLogger) Infof(template string, args ...interface{}) {
	l.sugarLogger.Infof(template, args...)
}

// Printf uses fmt.Sprintf to log a templated message
func (l *AppLogger) Printf(template string, args ...interface{}) {
	l.sugarLogger.Infof(template, args...)
}

// Warn uses fmt.Sprint to construct and log a message.
func (l *AppLogger) Warn(args ...interface{}) {
	l.sugarLogger.Warn(args...)
}

// WarnErrMsg log error message with warn level.
func (l *AppLogger) WarnErrMsg(msg string, err error) {
	l.logger.Warn(msg, zap.String("error", err.Error()))
}

// Warnf uses fmt.Sprintf to log a templated message.
func (l *AppLogger) Warnf(template string, args ...interface{}) {
	l.sugarLogger.Warnf(template, args...)
}

// Error uses fmt.Sprint to construct and log a message.
func (l *AppLogger) Error(args ...interface{}) {
	l.sugarLogger.Error(args...)
}

// Errorf uses fmt.Sprintf to log a templated message.
func (l *AppLogger) Errorf(template string, args ...interface{}) {
	l.sugarLogger.Errorf(template, args...)
}

// Err uses error to log a message.
func (l *AppLogger) Err(msg string, err error) {
	l.logger.Error(msg, zap.Error(err))
}

// DPanic uses fmt.Sprint to construct and log a message. In development, the logger then panics. (See DPanicLevel for details.)
func (l *AppLogger) DPanic(args ...interface{}) {
	l.sugarLogger.DPanic(args...)
}

// DPanicf uses fmt.Sprintf to log a templated message. In development, the logger then panics. (See DPanicLevel for details.)
func (l *AppLogger) DPanicf(template string, args ...interface{}) {
	l.sugarLogger.DPanicf(template, args...)
}

// Panic uses fmt.Sprint to construct and log a message, then panics.
func (l *AppLogger) Panic(args ...interface{}) {
	l.sugarLogger.Panic(args...)
}

// Panicf uses fmt.Sprintf to log a templated message, then panics
func (l *AppLogger) Panicf(template string, args ...interface{}) {
	l.sugarLogger.Panicf(template, args...)
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func (l *AppLogger) Fatal(args ...interface{}) {
	l.sugarLogger.Fatal(args...)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func (l *AppLogger) Fatalf(template string, args ...interface{}) {
	l.sugarLogger.Fatalf(template, args...)
}

func (l *AppLogger) HttpMiddlewareAccessLogger(method, uri string, status int, size int64, time time.Duration) {
	l.logger.Info(
		HTTP,
		zap.String(METHOD, method),
		zap.String(URI, uri),
		zap.Int(STATUS, status),
		zap.Int64(SIZE, size),
		zap.Duration(TIME, time),
	)
}

func (l *AppLogger) KafkaProcessMessage(topic string, partition int, message []byte, workerID int, offset int64, time time.Time) {
	l.logger.Debug(
		"(Processing Kafka message)",
		zap.String(Topic, topic),
		zap.Int(Partition, partition),
		zap.Int(MessageSize, len(message)),
		zap.Int(WorkerID, workerID),
		zap.Int64(Offset, offset),
		zap.Time(TIME, time),
	)
}

func (l *AppLogger) KafkaProcessMessageWithHeaders(topic string, partition int, message []byte, workerID int, offset int64, time time.Time, headers map[string]interface{}) {
	l.logger.Debug(
		"(Processing Kafka message)",
		zap.String(Topic, topic),
		zap.Int(Partition, partition),
		zap.Int(MessageSize, len(message)),
		zap.Int(WorkerID, workerID),
		zap.Int64(Offset, offset),
		zap.Time(TIME, time),
		zap.Any(KafkaHeaders, headers),
	)
}

// setupMainapiLogFile สร้างไฟล์ log สำหรับ dev mode ให้ AI อ่านได้
func setupMainapiLogFile() *os.File {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil
	}

	logPath := filepath.Join(logDir, "mainapi_debug.log")

	// ล้าง log file ทุกครั้งที่ server start
	_ = os.WriteFile(logPath, []byte(fmt.Sprintf("--- Log started: %s ---\n", time.Now().Format("2006-01-02 15:04:05"))), 0644)

	f, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil
	}
	return f
}
