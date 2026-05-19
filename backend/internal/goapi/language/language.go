package language

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
)

var (
	cache         map[string]map[string]string
	cacheLock     sync.RWMutex
	loaded        bool
	loadedPath    string
	loadedModTime time.Time
	loadedSize    int64
)

var aliases = map[string]string{
	"zh":    "cn",
	"zh-cn": "cn",
	"zh-tw": "cn",
	"jp":    "ja",
	"kr":    "ko",
	"tl":    "fil",
}

var supportedLanguages = map[string]struct{}{
	"th":  {},
	"en":  {},
	"cn":  {},
	"ja":  {},
	"ko":  {},
	"lo":  {},
	"my":  {},
	"km":  {},
	"vi":  {},
	"ms":  {},
	"id":  {},
	"fil": {},
}

// Normalize converts frontend, report, and legacy language codes to the TSV column code.
func Normalize(code string) string {
	value := strings.ToLower(strings.TrimSpace(code))
	value = strings.TrimPrefix(value, "ai:")
	value = strings.ReplaceAll(value, "_", "-")
	if value == "" {
		return "th"
	}
	if mapped, ok := aliases[value]; ok {
		return mapped
	}
	base := strings.Split(value, "-")[0]
	if mapped, ok := aliases[base]; ok {
		return mapped
	}
	if _, ok := supportedLanguages[base]; ok {
		return base
	}
	return "th"
}

// Text returns a translated value using language/languages.tsv with key fallback.
func Text(key string, lang string) string {
	dict, err := Dictionary(lang)
	if err != nil {
		return key
	}
	if text := dict[key]; text != "" {
		return text
	}
	return key
}

// Dictionary returns all keys for one language. Missing language cells fall back to the key itself.
func Dictionary(lang string) (map[string]string, error) {
	if err := Load(); err != nil {
		return nil, err
	}
	normalized := Normalize(lang)

	cacheLock.RLock()
	defer cacheLock.RUnlock()

	result := make(map[string]string, len(cache))
	for key, translations := range cache {
		if text := translations[normalized]; text != "" {
			result[key] = text
			continue
		}
		result[key] = key
	}
	return result, nil
}

// Load reads the shared language TSV. Docker Desktop development can enable mtime reload.
func Load() error {
	cacheLock.Lock()
	defer cacheLock.Unlock()

	path, err := findLanguageFile()
	if err != nil {
		return err
	}

	info, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("stat language file: %w", err)
	}
	if info.IsDir() {
		return fmt.Errorf("language file path is a directory: %s", path)
	}

	if loaded {
		if !languageReloadEnabled() {
			return nil
		}
		if loadedPath == path && loadedSize == info.Size() && loadedModTime.Equal(info.ModTime()) {
			return nil
		}
	}

	nextCache, err := readLanguageFile(path)
	if err != nil {
		return err
	}

	cache = nextCache
	loaded = true
	loadedPath = path
	loadedModTime = info.ModTime()
	loadedSize = info.Size()
	return nil
}

func readLanguageFile(path string) (map[string]map[string]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open language file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	scanner.Buffer(make([]byte, 0, 1024*1024), 4*1024*1024)

	if !scanner.Scan() {
		if err := scanner.Err(); err != nil {
			return nil, fmt.Errorf("read language header: %w", err)
		}
		return nil, fmt.Errorf("language file is empty")
	}

	headers := strings.Split(scanner.Text(), "\t")
	if len(headers) < 2 || headers[0] != "key" {
		return nil, fmt.Errorf("invalid language TSV header")
	}

	nextCache := make(map[string]map[string]string, 5000)
	for scanner.Scan() {
		cols := strings.Split(scanner.Text(), "\t")
		if len(cols) < 2 {
			continue
		}
		key := strings.TrimSpace(cols[0])
		if key == "" || strings.HasPrefix(key, "#") {
			continue
		}

		translations := make(map[string]string, len(headers)-1)
		for index := 1; index < len(headers) && index < len(cols); index++ {
			if cols[index] == "" {
				continue
			}
			translations[Normalize(headers[index])] = cols[index]
		}
		nextCache[key] = translations
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scan language file: %w", err)
	}
	return nextCache, nil
}

func languageReloadEnabled() bool {
	value := strings.ToLower(strings.TrimSpace(os.Getenv("LANGUAGE_RELOAD_ENABLED")))
	return value == "1" || value == "true" || value == "yes" || value == "on"
}

func findLanguageFile() (string, error) {
	if configured := strings.TrimSpace(os.Getenv("LANGUAGE_FILE_PATH")); configured != "" {
		path := filepath.Clean(configured)
		if info, err := os.Stat(path); err == nil && !info.IsDir() {
			return path, nil
		}
		return "", fmt.Errorf("LANGUAGE_FILE_PATH %q not found", configured)
	}

	candidates := []string{
		filepath.Join("language", "languages.tsv"),
		filepath.Join("assets", "language", "languages.tsv"),
	}

	if _, filename, _, ok := runtime.Caller(0); ok {
		sourceRoot := filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", ".."))
		candidates = append(candidates, filepath.Join(sourceRoot, "assets", "language", "languages.tsv"))
	}

	for _, candidate := range candidates {
		if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("language/languages.tsv not found")
}
