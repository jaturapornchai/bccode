package logger

import (
	"os"
	"testing"
)

func TestLogLevelConfiguration(t *testing.T) {
	// Test environment variables แต่ละแบบ
	tests := []struct {
		name     string
		envVars  map[string]string
		expected LogLevel
	}{
		{
			name:     "LOG_LEVEL=INFO",
			envVars:  map[string]string{"LOG_LEVEL": "INFO"},
			expected: INFO,
		},
		{
			name:     "LOG_LEVEL=WARN",
			envVars:  map[string]string{"LOG_LEVEL": "WARN"},
			expected: WARN,
		},
		{
			name:     "LOG_LEVEL=error (case insensitive)",
			envVars:  map[string]string{"LOG_LEVEL": "error"},
			expected: ERROR,
		},
		{
			name:     "DEBUG_MODE=true",
			envVars:  map[string]string{"DEBUG_MODE": "true"},
			expected: DEBUG, // Legacy DEBUG_MODE should set to DEBUG in development
		},
		{
			name:     "DEBUG_MODE=false",
			envVars:  map[string]string{"DEBUG_MODE": "false"},
			expected: INFO, // Should default to INFO
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Save current environment
			originalEnv := make(map[string]string)
			for key := range tt.envVars {
				originalEnv[key] = os.Getenv(key)
			}

			// Clear environment first
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("DEBUG_MODE")

			// Set test environment
			for key, value := range tt.envVars {
				os.Setenv(key, value)
			}

			// Use InitDebugMode to test environment variable handling
			InitDebugMode()
			
			if CurrentLogLevel != tt.expected {
				t.Errorf("Expected log level %v, got %v", tt.expected, CurrentLogLevel)
			}

			// Restore environment
			os.Unsetenv("LOG_LEVEL")
			os.Unsetenv("DEBUG_MODE")
			for key, value := range originalEnv {
				if value != "" {
					os.Setenv(key, value)
				}
			}
		})
	}
}

func TestShouldLog(t *testing.T) {
	tests := []struct {
		level     LogLevel
		current   LogLevel
		shouldLog bool
	}{
		{DEBUG, DEBUG, true},
		{DEBUG, INFO, false},
		{INFO, DEBUG, true},
		{INFO, INFO, true},
		{INFO, WARN, false},
		{WARN, INFO, true},
		{WARN, WARN, true},
		{WARN, ERROR, false},
		{ERROR, WARN, true},
		{ERROR, ERROR, true},
		{ERROR, FATAL, false},
		{FATAL, ERROR, true},
		{FATAL, FATAL, true},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			original := CurrentLogLevel
			CurrentLogLevel = tt.current
			defer func() { CurrentLogLevel = original }()

			result := shouldLog(tt.level)
			if result != tt.shouldLog {
				t.Errorf("shouldLog(%v) with current level %v should be %v, got %v",
					tt.level, tt.current, tt.shouldLog, result)
			}
		})
	}
}

func TestSetLogLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected LogLevel
	}{
		{"DEBUG", DEBUG},
		{"INFO", INFO},
		{"SUCCESS", SUCCESS},
		{"WARN", WARN},
		{"ERROR", ERROR},
		{"FATAL", FATAL},
		{"debug", DEBUG},    // case insensitive
		{"Info", INFO},      // case insensitive
	}

	for _, tt := range tests {
		t.Run("Set to "+tt.input, func(t *testing.T) {
			original := CurrentLogLevel
			SetLogLevel(tt.input)
			defer func() { CurrentLogLevel = original }()

			if CurrentLogLevel != tt.expected {
				t.Errorf("SetLogLevel(%q) should set level to %v, got %v", tt.input, tt.expected, CurrentLogLevel)
			}
		})
	}
}

// Test separate test for invalid level
func TestSetLogLevelInvalid(t *testing.T) {
	original := CurrentLogLevel
	defer func() { CurrentLogLevel = original }()

	// Set to known level first
	CurrentLogLevel = INFO
	
	// Try to set invalid level
	SetLogLevel("unknown")
	
	// Should keep current level
	if CurrentLogLevel != INFO {
		t.Errorf("SetLogLevel('unknown') should keep current level INFO, got %v", CurrentLogLevel)
	}
}

func TestGetLogLevel(t *testing.T) {
	tests := []struct {
		level    LogLevel
		expected string
	}{
		{DEBUG, "DEBUG"},
		{INFO, "INFO"},
		{SUCCESS, "SUCCESS"},
		{WARN, "WARN"},
		{ERROR, "ERROR"},
		{FATAL, "FATAL"},
	}

	for _, tt := range tests {
		t.Run("Get "+tt.expected, func(t *testing.T) {
			original := CurrentLogLevel
			CurrentLogLevel = tt.level
			defer func() { CurrentLogLevel = original }()

			result := GetLogLevel()
			if result != tt.expected {
				t.Errorf("GetLogLevel() should return %q, got %q", tt.expected, result)
			}
		})
	}
}

// Benchmark test for performance comparison
func BenchmarkLoggerDebug(b *testing.B) {
	b.Run("WithLevelCheck", func(b *testing.B) {
		CurrentLogLevel = INFO // Lower than DEBUG
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			Debug("Test message %d", i)
		}
	})

	b.Run("WithLevelMatch", func(b *testing.B) {
		CurrentLogLevel = DEBUG // Same level
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			Debug("Test message %d", i)
		}
	})
}

func BenchmarkLoggerInfo(b *testing.B) {
	b.Run("WithLevelCheck", func(b *testing.B) {
		CurrentLogLevel = WARN // Higher than INFO
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			Info("Test message %d", i)
		}
	})

	b.Run("WithLevelMatch", func(b *testing.B) {
		CurrentLogLevel = INFO // Same level
		b.ResetTimer()
		
		for i := 0; i < b.N; i++ {
			Info("Test message %d", i)
		}
	})
}