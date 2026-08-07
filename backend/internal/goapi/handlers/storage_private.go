package handlers

import (
	"encoding/base64"
	"net/http"
	"os"
	"strings"

	msmodels "smlcloudplatform/pkg/microservice/models"

	"github.com/labstack/echo/v4"
)

const storageAllowPresignedURLEnv = "STORAGE_ALLOW_PRESIGNED_URL"

func storageAllowsDirectPresignedURL() bool {
	return strings.EqualFold(strings.TrimSpace(os.Getenv(storageAllowPresignedURLEnv)), "true")
}

func storageProxyURL(objectKey string) string {
	return "/s3/file/" + storageNormalizeObjectKey(objectKey)
}

func storageNormalizeObjectKey(objectKey string) string {
	normalized := strings.ReplaceAll(strings.TrimSpace(objectKey), "\\", "/")
	return strings.Trim(normalized, "/")
}

func storageObjectHoldingCode(objectKey string) string {
	normalized := storageNormalizeObjectKey(objectKey)
	if normalized == "" || strings.Contains(normalized, "..") {
		return ""
	}
	parts := strings.SplitN(normalized, "/", 2)
	return strings.TrimSpace(parts[0])
}

func storageObjectBelongsToShop(objectKey string, holdingCode string) bool {
	holdingCode = strings.TrimSpace(holdingCode)
	if holdingCode == "" {
		return false
	}
	return storageObjectHoldingCode(objectKey) == holdingCode
}

func storageObjectBusinessCode(objectKey string) string {
	normalized := storageNormalizeObjectKey(objectKey)
	if normalized == "" || strings.Contains(normalized, "..") {
		return ""
	}
	parts := strings.Split(normalized, "/")
	if len(parts) < 3 || parts[1] != "companies" {
		return ""
	}
	return storageBusinessCodeFromPathSegment(parts[2])
}

func storageBusinessPathSegment(businessCode string) string {
	businessCode = strings.TrimSpace(businessCode)
	if businessCode == "" {
		return ""
	}
	return "~" + base64.RawURLEncoding.EncodeToString([]byte(businessCode))
}

func storageBusinessCodeFromPathSegment(segment string) string {
	segment = strings.TrimSpace(segment)
	if !strings.HasPrefix(segment, "~") {
		return segment
	}
	decoded, err := base64.RawURLEncoding.DecodeString(strings.TrimPrefix(segment, "~"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(decoded))
}

func storageObjectBelongsToContext(objectKey, holdingCode, businessCode string) bool {
	if !storageObjectBelongsToShop(objectKey, holdingCode) {
		return false
	}
	parts := strings.Split(storageNormalizeObjectKey(objectKey), "/")
	if len(parts) < 2 || parts[1] != "companies" {
		return true
	}
	objectBusinessCode := storageObjectBusinessCode(objectKey)
	return objectBusinessCode != "" && objectBusinessCode == strings.TrimSpace(businessCode)
}

func storageContextHoldingCode(c echo.Context) string {
	userInfo, ok := c.Get("UserInfo").(msmodels.UserInfo)
	if !ok {
		return ""
	}
	return strings.TrimSpace(userInfo.HoldingCode)
}

func storageContextBusinessCode(c echo.Context) string {
	userInfo, ok := c.Get("UserInfo").(msmodels.UserInfo)
	if !ok {
		return ""
	}
	businessCode := strings.TrimSpace(userInfo.BusinessCode)
	if businessCode == "" {
		return ""
	}
	return businessCode
}

func storageAuthorizedBusinessCode(c echo.Context) (string, int) {
	businessCode := storageContextBusinessCode(c)
	if businessCode == "" {
		return "", http.StatusBadRequest
	}
	return businessCode, http.StatusOK
}

func storageSanitizeCategory(value string) string {
	parts := strings.FieldsFunc(strings.TrimSpace(value), func(r rune) bool { return r == '/' || r == '\\' })
	clean := make([]string, 0, len(parts))
	for _, part := range parts {
		var segment strings.Builder
		for _, r := range part {
			if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '~' {
				segment.WriteRune(r)
			}
		}
		if segment.Len() > 0 {
			clean = append(clean, segment.String())
		}
	}
	return strings.Join(clean, "/")
}

func storageAuthorizedHoldingCode(c echo.Context, requestedHoldingCode string) (string, int) {
	tokenHoldingCode := storageContextHoldingCode(c)
	if tokenHoldingCode == "" {
		return "", http.StatusUnauthorized
	}
	requestedHoldingCode = strings.TrimSpace(requestedHoldingCode)
	if requestedHoldingCode != "" && requestedHoldingCode != tokenHoldingCode {
		return "", http.StatusForbidden
	}
	return tokenHoldingCode, http.StatusOK
}
