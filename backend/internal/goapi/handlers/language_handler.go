package handlers

import (
	"net/http"
	"smlcloudplatform/internal/goapi/language"
	"strings"

	"github.com/labstack/echo/v4"
)

// GetLanguageHandler - GET /api/language/:lang
// ดึงข้อมูลภาษาเฉพาะที่ต้องการ
// เช่น /api/language/th จะได้ {"code1": "ข้อความ1", "code2": "ข้อความ2", ...}
func GetLanguageHandler(c echo.Context) error {
	rawLang := c.Param("lang")
	if strings.TrimSpace(rawLang) == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "กรุณาระบุภาษา เช่น /api/language/th",
		})
	}
	lang := language.Normalize(rawLang)

	result, err := language.Dictionary(lang)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "ไม่สามารถโหลดไฟล์ภาษาได้: " + err.Error(),
		})
	}

	return c.JSON(http.StatusOK, result)
}
