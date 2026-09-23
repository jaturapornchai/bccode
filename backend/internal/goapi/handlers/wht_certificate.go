package handlers

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/language"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/whtcert"

	"github.com/labstack/echo/v4"
)

// whtCompanyHeader - จุดสลับให้เทสระดับ handler ไม่ต้องต่อฐานข้อมูลควบคุมกลาง
var whtCompanyHeader = loadCompanyHeader

// WhtCertificateRequest - ข้อมูลใบ 50 ทวิ ที่หน้าจอเตรียมจากรายงานภาษีหัก ณ ที่จ่าย (นักบัญชีตรวจ/แก้ได้ก่อนสั่งพิมพ์)
type WhtCertificateRequest struct {
	HoldingCode  string              `json:"holdingcode"`
	BusinessCode string              `json:"businesscode"`
	Certificate  whtcert.Certificate `json:"certificate"`
}

// WhtCertificateHandler - POST /api/report/tax/wht/certificate → application/pdf
// backend สร้าง PDF บนแบบฟอร์มของกรมสรรพากรเอง (ตรวจเลข/ยอด คำนวณยอดรวมและตัวอักษร) หน้าจอแค่แสดงผล
// ชื่อและเลขผู้เสียภาษีของผู้หักภาษียึดทะเบียนบริษัทเสมอ เมื่อทะเบียนมีค่า (กันพิมพ์ในนามบริษัทอื่น)
func WhtCertificateHandler(c echo.Context) error {
	lang := language.Normalize(c.Request().Header.Get("Accept-Language"))
	fail := func(status int, key, field string) error {
		return c.JSON(status, map[string]any{"success": false, "code": key, "field": field, "message": language.Text(key, lang)})
	}

	var req WhtCertificateRequest
	if err := c.Bind(&req); err != nil {
		return fail(http.StatusBadRequest, "wht_cert_payload_invalid", "")
	}
	holdingCode, businessCode, scopeErr := authenticatedCompanyContext(c, req.HoldingCode, req.BusinessCode)
	if scopeErr != nil {
		return c.JSON(scopeErr.Status, map[string]any{"success": false, "code": scopeErr.Code, "message": scopeErr.Message})
	}

	ctx, cancel := context.WithTimeout(c.Request().Context(), 15*time.Second)
	defer cancel()
	company, err := whtCompanyHeader(ctx, holdingCode, businessCode)
	if err != nil {
		logger.Error("WhtCertificate: company header: %v", err)
		return fail(http.StatusInternalServerError, "wht_cert_render_failed", "")
	}
	cert := req.Certificate
	if name := strings.TrimSpace(company.Name); name != "" {
		cert.Payer.Name = name
	}
	if taxID := strings.TrimSpace(company.TaxID); taxID != "" {
		cert.Payer.TaxID = taxID
	}

	pdf, err := whtcert.Render(cert)
	var invalid *whtcert.ValidationError
	if errors.As(err, &invalid) {
		return fail(http.StatusBadRequest, invalid.Key, invalid.Field)
	}
	if err != nil {
		logger.Error("WhtCertificate: render: %v", err)
		return fail(http.StatusInternalServerError, "wht_cert_render_failed", "")
	}

	c.Response().Header().Set("Content-Disposition", fmt.Sprintf(`inline; filename="50tawi-%s.pdf"`, safeFileToken(cert.RunNo)))
	c.Response().Header().Set("Cache-Control", "no-store")
	return c.Blob(http.StatusOK, "application/pdf", pdf)
}

// safeFileToken - เลขที่เอกสารสำหรับชื่อไฟล์ (ตัด / และอักขระพิเศษ กัน header injection)
func safeFileToken(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= '0' && r <= '9', r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r == '-':
			b.WriteRune(r)
		case r == '/' || r == '_' || r == ' ':
			b.WriteByte('-')
		}
	}
	if b.Len() == 0 {
		return "certificate"
	}
	return b.String()
}
