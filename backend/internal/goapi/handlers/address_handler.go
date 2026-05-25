package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/labstack/echo/v4"
)

const thailandAddressContentType = "application/json; charset=utf-8"

var (
	thailandAddressOnce sync.Once
	thailandAddressData []byte
	thailandAddressErr  error
)

// GetThailandAddressHandler - GET /api/address/thailand
// Serves the licensed Thailand province/district/subdistrict/postal-code dataset
// from backend assets so the frontend build does not carry the static JSON file.
func GetThailandAddressHandler(c echo.Context) error {
	data, err := loadThailandAddressData()
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "ไม่สามารถโหลดข้อมูลที่อยู่ไทยได้: " + err.Error(),
		})
	}

	c.Response().Header().Set(echo.HeaderCacheControl, "public, max-age=86400")
	return c.Blob(http.StatusOK, thailandAddressContentType, data)
}

func loadThailandAddressData() ([]byte, error) {
	thailandAddressOnce.Do(func() {
		for _, candidate := range thailandAddressDataPathCandidates() {
			data, err := os.ReadFile(candidate)
			if err == nil {
				thailandAddressData = data
				thailandAddressErr = nil
				return
			}
		}
		thailandAddressErr = fmt.Errorf("file not found in configured asset paths")
	})
	return thailandAddressData, thailandAddressErr
}

func thailandAddressDataPathCandidates() []string {
	candidates := make([]string, 0, 6)
	if configured := strings.TrimSpace(os.Getenv("THAILAND_ADDRESS_DATA_PATH")); configured != "" {
		candidates = append(candidates, configured)
	}
	candidates = append(candidates,
		filepath.FromSlash("/app/address/thailand-addresses.json"),
		filepath.FromSlash("address/thailand-addresses.json"),
		filepath.FromSlash("assets/address/thailand-addresses.json"),
		filepath.FromSlash("backend/assets/address/thailand-addresses.json"),
		filepath.FromSlash("../../../assets/address/thailand-addresses.json"),
	)
	return candidates
}
