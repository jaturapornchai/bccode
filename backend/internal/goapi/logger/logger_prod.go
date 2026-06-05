//go:build production

package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

// LogLevel types
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	SUCCESS
	WARN
	ERROR
	FATAL
)

// Production Log Level Configuration
var (
	// Production default: แสดงเฉพาะ INFO, SUCCESS, WARN, ERROR, FATAL
	CurrentLogLevel = INFO

	// Map string to LogLevel
	logLevelMap = map[string]LogLevel{
		"DEBUG":   DEBUG,
		"INFO":    INFO,
		"SUCCESS": SUCCESS,
		"WARN":    WARN,
		"ERROR":   ERROR,
		"FATAL":   FATAL,
	}
)

func init() {
	// ตั้งค่า Production mode
	log.SetFlags(log.Ldate | log.Ltime)
	log.SetOutput(os.Stdout)

	// อ่าน LOG_LEVEL จาก environment (default: INFO)
	logLevelStr := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	if level, exists := logLevelMap[logLevelStr]; exists {
		CurrentLogLevel = level
	}

	// ไม่แสดง log ข้อมูลการเริ่มต้นใน production
	// log.Println("[LOGGER] Production mode - selective logging enabled")
}

// getCallerInfo - ดึงข้อมูล file และ line ของ caller
func getCallerInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown:0"
	}
	file = filepath.ToSlash(file)
	return fmt.Sprintf("%s:%d", file, line)
}

// shouldLog - ตรวจสอบว่าควรแสดง log หรือไม่
func shouldLog(level LogLevel) bool {
	return level >= CurrentLogLevel
}

// Debug - แสดงเฉพาะใน dev mode หรือ level >= DEBUG
func Debug(format string, args ...interface{}) {
	if !shouldLog(DEBUG) {
		return
	}
	caller := getCallerInfo(2)
	log.Printf("%s: [DEBUG] "+format, append([]interface{}{caller}, args...)...)
}

// Info - แสดงในทุก mode (default production level)
func Info(format string, args ...interface{}) {
	if !shouldLog(INFO) {
		return
	}
	caller := getCallerInfo(2)
	log.Printf("%s: [INFO] "+format, append([]interface{}{caller}, args...)...)
}

// Success - แสดงในทุก mode
func Success(format string, args ...interface{}) {
	if !shouldLog(SUCCESS) {
		return
	}
	caller := getCallerInfo(2)
	log.Printf("%s: ✓ "+format, append([]interface{}{caller}, args...)...)
}

// Error - แสดงในทุก mode
func Error(format string, args ...interface{}) {
	if !shouldLog(ERROR) {
		return
	}
	caller := getCallerInfo(2)
	log.Printf("%s: ❌ "+format, append([]interface{}{caller}, args...)...)
}

// Warn - แสดงในทุก mode
func Warn(format string, args ...interface{}) {
	if !shouldLog(WARN) {
		return
	}
	caller := getCallerInfo(2)
	log.Printf("%s: ⚠️  "+format, append([]interface{}{caller}, args...)...)
}

// Fatal - แสดงและหยุดโปรแกรมในทุก mode
func Fatal(format string, args ...interface{}) {
	caller := getCallerInfo(2)
	log.Fatalf("%s: 💥 "+format, append([]interface{}{caller}, args...)...)
}

// Printf - สำหรับ backward compatibility
func Printf(format string, args ...interface{}) {
	// ใน production ไม่แสดง Printf โดย default
	if !shouldLog(INFO) {
		return
	}
	caller := getCallerInfo(2)
	log.Printf("%s: "+format, append([]interface{}{caller}, args...)...)
}

// SetLogLevel - เปลี่ยน log level ใน runtime
func SetLogLevel(levelStr string) {
	if level, exists := logLevelMap[strings.ToUpper(levelStr)]; exists {
		CurrentLogLevel = level
	} else {
		// Invalid level, keep current
	}
}

// GetLogLevel - ดึง log level ปัจจุบัน
func GetLogLevel() string {
	for levelStr, level := range logLevelMap {
		if level == CurrentLogLevel {
			return levelStr
		}
	}
	return "UNKNOWN"
}

// InitProductionMode - เริ่มต้น production mode
func InitProductionMode() {
	// ตั้งค่า environment
	if os.Getenv("ENVIRONMENT") == "production" {
		os.Setenv("LOG_LEVEL", "INFO") // Production: INFO level
	}

	// อ่านค่า log level
	logLevelStr := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	if level, exists := logLevelMap[logLevelStr]; exists {
		CurrentLogLevel = level
	}
}
