package config

import (
	"os"
	"strconv"
)

// Query Limits Configuration (จาก Python config.py)
const (
	DefaultMaxQueryRows        = 10000
	DefaultMaxResultRowsPerPage = 1000
	DefaultPageSize            = 100
	DefaultQueryTimeoutSeconds = 60
)

// PDF Configuration (จาก Python config.py)
const (
	DefaultPDFFontPath       = "./fonts/THSarabunNew.ttf"
	DefaultPDFPageSize       = "A4"
	DefaultPDFOrientation    = "L" // L = Landscape, P = Portrait
)

// GetMaxQueryRows returns max rows for query results
func GetMaxQueryRows() int {
	return getEnvAsInt("MAX_QUERY_ROWS", DefaultMaxQueryRows)
}

// GetMaxResultRowsPerPage returns max rows per page
func GetMaxResultRowsPerPage() int {
	return getEnvAsInt("MAX_RESULT_ROWS_PER_PAGE", DefaultMaxResultRowsPerPage)
}

// GetDefaultPageSize returns default page size
func GetDefaultPageSize() int {
	return getEnvAsInt("DEFAULT_PAGE_SIZE", DefaultPageSize)
}

// GetQueryTimeoutSeconds returns query timeout in seconds
func GetQueryTimeoutSeconds() int {
	return getEnvAsInt("QUERY_TIMEOUT_SECONDS", DefaultQueryTimeoutSeconds)
}

// GetPDFFontPath returns PDF font path
func GetPDFFontPath() string {
	return getEnv("PDF_FONT_PATH", DefaultPDFFontPath)
}

// GetPDFDefaultPageSize returns default PDF page size
func GetPDFDefaultPageSize() string {
	return getEnv("PDF_DEFAULT_PAGE_SIZE", DefaultPDFPageSize)
}

// GetPDFDefaultOrientation returns default PDF orientation (L=Landscape, P=Portrait)
func GetPDFDefaultOrientation() string {
	return getEnv("PDF_DEFAULT_ORIENTATION", DefaultPDFOrientation)
}

// getEnvAsInt helper function to get env as int
func getEnvAsInt(key string, fallback int) int {
	value, exists := os.LookupEnv(key)
	if !exists {
		return fallback
	}
	intVal, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return intVal
}
