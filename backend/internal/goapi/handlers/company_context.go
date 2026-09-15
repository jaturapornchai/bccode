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

// respond ส่งข้อผิดพลาดขอบเขตบริษัทกลับในรูปแบบเดียวกันทุกที่
func (e *companyContextError) respond(c echo.Context) error {
	return c.JSON(e.Status, map[string]any{
		"success": false,
		"code":    e.Code,
		"message": e.Message,
	})
}

// reportCompanyScope คืนรหัสบริษัทของผู้ใช้ที่ล็อกอินอยู่ สำหรับรายงานที่ต้องกรองข้อมูลตามบริษัท
//
// รายงานสต็อกอ่านจากตารางที่มีหลายบริษัทอยู่ในฐานข้อมูลเดียวกัน
// ถ้าไม่กรองบริษัท ยอดของอีกบริษัทในกลุ่มเดียวกันจะปนเข้ามาในรายงาน
func reportCompanyScope(c echo.Context, requestedHolding string) (string, *companyContextError) {
	_, businessCode, err := authenticatedCompanyContext(c, requestedHolding, "")
	if err != nil {
		return "", err
	}
	return businessCode, nil
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
