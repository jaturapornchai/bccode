package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
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

// Development Log Configuration
var (
	// Development default: แสดงทุก level
	CurrentLogLevel = DEBUG
	
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

// devLogFile ไฟล์ log สำหรับ dev mode ให้ AI อ่านได้
var devLogFile *os.File

func init() {
	// Development mode: แสดง timestamp
	log.SetFlags(log.Ldate | log.Ltime)

	// Dev mode: เขียน log ทั้ง console และ file
	devLogFile = setupDevLogFile()
	if devLogFile != nil {
		log.SetOutput(io.MultiWriter(os.Stdout, devLogFile))
	} else {
		log.SetOutput(os.Stdout)
	}

	// อ่าน LOG_LEVEL จาก environment (default: DEBUG for dev)
	logLevelStr := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	if level, exists := logLevelMap[logLevelStr]; exists {
		CurrentLogLevel = level
	} else {
		CurrentLogLevel = DEBUG // Dev default: แสดงทุกอย่าง
	}

	log.Println("[LOGGER] Development mode - logging ENABLED with", GetLogLevel(), "level")
}

// getCallerInfo - ดึงข้อมูล file และ line ของ caller
func getCallerInfo(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown:0"
	}
	// แปลง path ให้เป็น forward slash และแสดง full path
	file = filepath.ToSlash(file)
	return fmt.Sprintf("%s:%d", file, line)
}

// shouldLog - ตรวจสอบว่าควรแสดง log หรือไม่
func shouldLog(level LogLevel) bool {
	return level >= CurrentLogLevel
}

// Debug - log เฉพาะตอน level >= DEBUG
func Debug(format string, args ...interface{}) {
	if !shouldLog(DEBUG) {
		return
	}
	caller := getCallerInfo(2)
	log.Printf("%s: [DEBUG] "+format, append([]interface{}{caller}, args...)...)
}

// Info - log ข้อมูลทั่วไป
func Info(format string, args ...interface{}) {
	if !shouldLog(INFO) {
		return
	}
	caller := getCallerInfo(2)
	log.Printf("%s: [INFO] "+format, append([]interface{}{caller}, args...)...)
}

// Success - log เมื่อสำเร็จ
func Success(format string, args ...interface{}) {
	if !shouldLog(SUCCESS) {
		return
	}
	caller := getCallerInfo(2)
	log.Printf("%s: ✓ "+format, append([]interface{}{caller}, args...)...)
}

// Error - log error
func Error(format string, args ...interface{}) {
	if !shouldLog(ERROR) {
		return
	}
	caller := getCallerInfo(2)
	log.Printf("%s: ❌ "+format, append([]interface{}{caller}, args...)...)
}

// Warn - log warning
func Warn(format string, args ...interface{}) {
	if !shouldLog(WARN) {
		return
	}
	caller := getCallerInfo(2)
	log.Printf("%s: ⚠️  "+format, append([]interface{}{caller}, args...)...)
}

// Fatal - log error และหยุดโปรแกรม
func Fatal(format string, args ...interface{}) {
	caller := getCallerInfo(2)
	log.Fatalf("%s: 💥 "+format, append([]interface{}{caller}, args...)...)
}

// Printf - backward compatibility
func Printf(format string, args ...interface{}) {
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
		log.Println("[LOGGER] Log level set to", levelStr)
	} else {
		log.Println("[LOGGER] Invalid log level:", levelStr, "keeping current level")
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

// setupDevLogFile สร้างไฟล์ log สำหรับ dev mode ให้ AI อ่านได้
func setupDevLogFile() *os.File {
	logDir := "logs"
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil
	}

	logPath := filepath.Join(logDir, "goapi_debug.log")

	// ล้าง log file ทุกครั้งที่ server start
	_ = os.WriteFile(logPath, []byte(fmt.Sprintf("--- Log started: %s ---\n", time.Now().Format("2006-01-02 15:04:05"))), 0644)

	f, err := os.OpenFile(logPath, os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil
	}
	return f
}

// ClearLogFile ล้าง log file (เรียกจากภายนอก)
func ClearLogFile() {
	logDir := "logs"
	logPath := filepath.Join(logDir, "goapi_debug.log")

	if devLogFile != nil {
		devLogFile.Close()
	}

	_ = os.WriteFile(logPath, []byte(fmt.Sprintf("--- Log cleared: %s ---\n", time.Now().Format("2006-01-02 15:04:05"))), 0644)

	devLogFile = setupDevLogFile()
	if devLogFile != nil {
		log.SetOutput(io.MultiWriter(os.Stdout, devLogFile))
	}
}

// InitDebugMode - reload DEBUG_MODE setting (legacy support)
func InitDebugMode() {
	logLevelStr := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	if logLevelStr == "" {
		// Legacy compatibility: DEBUG_MODE=true -> DEBUG level, false -> INFO level
		if os.Getenv("DEBUG_MODE") == "true" {
			CurrentLogLevel = DEBUG
		} else {
			CurrentLogLevel = INFO // Default to INFO if DEBUG_MODE=false or not set
		}
	} else {
		// Use new LOG_LEVEL if provided
		if level, exists := logLevelMap[logLevelStr]; exists {
			CurrentLogLevel = level
		}
	}

	if CurrentLogLevel <= INFO {
		log.Println("[LOGGER] ⚙️  Debug mode ENABLED with", GetLogLevel(), "level")
	} else {
		log.Println("[LOGGER] ⚙️  Debug mode DISABLED (minimal logging) -", GetLogLevel(), "level")
	}
}
