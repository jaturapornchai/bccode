package handlers

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync"

	"github.com/labstack/echo/v4"
)

// languageCache - cache สำหรับเก็บ languages.json
var (
	allLanguagesCache map[string]map[string]string
	languageCacheLock sync.RWMutex
	languageLoaded    bool
)

// loadLanguages - โหลด languages.json เข้า cache
func loadLanguages() error {
	languageCacheLock.Lock()
	defer languageCacheLock.Unlock()

	if languageLoaded {
		return nil
	}

	filePath := filepath.Join("language", "languages.json")
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(data, &allLanguagesCache); err != nil {
		return err
	}

	languageLoaded = true
	return nil
}

// GetLanguageHandler - GET /api/language/:lang
// ดึงข้อมูลภาษาเฉพาะที่ต้องการ
// เช่น /api/language/th จะได้ {"code1": "ข้อความ1", "code2": "ข้อความ2", ...}
func GetLanguageHandler(c echo.Context) error {
	lang := c.Param("lang")
	if lang == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "กรุณาระบุภาษา เช่น /api/language/th",
		})
	}

	// โหลด cache ถ้ายังไม่เคยโหลด
	if err := loadLanguages(); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "ไม่สามารถโหลดไฟล์ภาษาได้: " + err.Error(),
		})
	}

	languageCacheLock.RLock()
	defer languageCacheLock.RUnlock()

	// แปลงเป็น map[code]text สำหรับภาษาที่ต้องการ
	result := make(map[string]string)
	for code, translations := range allLanguagesCache {
		text := translations[lang]
		if text != "" {
			result[code] = text
		}
	}

	return c.JSON(http.StatusOK, result)
}
