package handlers

import (
	"net/http"
	"strings"

	"smlcloudplatform/internal/utils"
	msmodels "smlcloudplatform/pkg/microservice/models"

	"github.com/labstack/echo/v4"
)

type companyContextError struct {
	Status  int
	Code    string
	Message string
}

func authenticatedCompanyContext(c echo.Context, requestedHolding, requestedBusiness string) (string, string, *companyContextError) {
	userInfo, ok := c.Get("UserInfo").(msmodels.UserInfo)
	holdingCode := strings.TrimSpace(userInfo.HoldingCode)
	if !ok || holdingCode == "" {
		return "", "", &companyContextError{http.StatusUnauthorized, "UNAUTHORIZED", "Unauthorized"}
	}
	if requested := strings.TrimSpace(requestedHolding); requested != "" && !strings.EqualFold(requested, holdingCode) {
		return "", "", &companyContextError{http.StatusForbidden, "FORBIDDEN", "Forbidden"}
	}
	businessCode := utils.NormalizeBusinessCode(userInfo.BusinessCode)
	if businessCode == "" {
		return "", "", &companyContextError{http.StatusConflict, "COMPANY_REQUIRED", "กรุณาเลือกบริษัทก่อนใช้งาน"}
	}
	if requested := utils.NormalizeBusinessCode(requestedBusiness); requested != "" && requested != businessCode {
		return "", "", &companyContextError{http.StatusForbidden, "FORBIDDEN", "Forbidden"}
	}
	return holdingCode, businessCode, nil
}
