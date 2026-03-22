package handlers

import (
	"bufio"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

// languageCache - cache สำหรับเก็บ languages.tsv
var (
	allLanguagesCache map[string]map[string]string // key → lang → text
	languageCacheLock sync.RWMutex
	languageLoaded    bool
)

// loadLanguages - โหลด languages.tsv เข้า cache
// TSV format: key\tth\ten\tcn\tja\tkm\tko\tlo\tmy\tvi
func loadLanguages() error {
	languageCacheLock.Lock()
	defer languageCacheLock.Unlock()

	if languageLoaded {
		return nil
	}

	filePath := filepath.Join("language", "languages.tsv")
	f, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024) // 1MB buffer

	// อ่าน header เพื่อหา column index ของแต่ละภาษา
	if !scanner.Scan() {
		return scanner.Err()
	}
	headers := strings.Split(scanner.Text(), "\t")
	// headers[0] = "key", headers[1..] = lang codes

	allLanguagesCache = make(map[string]map[string]string, 5000)

	for scanner.Scan() {
		cols := strings.Split(scanner.Text(), "\t")
		if len(cols) < 2 {
			continue
		}
		key := cols[0]
		langs := make(map[string]string, len(headers)-1)
		for i := 1; i < len(headers) && i < len(cols); i++ {
			if cols[i] != "" {
				langs[headers[i]] = cols[i]
			}
		}
		allLanguagesCache[key] = langs
	}

	if err := scanner.Err(); err != nil {
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
