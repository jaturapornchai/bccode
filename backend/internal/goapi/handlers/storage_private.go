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

func storageObjectShopID(objectKey string) string {
	normalized := storageNormalizeObjectKey(objectKey)
	if normalized == "" || strings.Contains(normalized, "..") {
		return ""
	}
	parts := strings.SplitN(normalized, "/", 2)
	return strings.TrimSpace(parts[0])
}

func storageObjectBelongsToShop(objectKey string, shopID string) bool {
	shopID = strings.TrimSpace(shopID)
	if shopID == "" {
		return false
	}
	return storageObjectShopID(objectKey) == shopID
}

func storageContextShopID(c echo.Context) string {
	userInfo, ok := c.Get("UserInfo").(msmodels.UserInfo)
	if !ok {
		return ""
	}
	return strings.TrimSpace(userInfo.ShopID)
}

func storageAuthorizedShopID(c echo.Context, requestedShopID string) (string, int) {
	tokenShopID := storageContextShopID(c)
	if tokenShopID == "" {
		return "", http.StatusUnauthorized
	}
	requestedShopID = strings.TrimSpace(requestedShopID)
	if requestedShopID != "" && requestedShopID != tokenShopID {
		return "", http.StatusForbidden
	}
	return tokenShopID, http.StatusOK
}
