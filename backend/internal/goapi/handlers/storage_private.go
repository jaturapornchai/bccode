package handlers

import (
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

func storageContextHoldingCode(c echo.Context) string {
	userInfo, ok := c.Get("UserInfo").(msmodels.UserInfo)
	if !ok {
		return ""
	}
	return strings.TrimSpace(userInfo.HoldingCode)
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
