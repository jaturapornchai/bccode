package datainfo

import (
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"html"
)

func DataInfoHtml(shopId string, function string, mode int) string {
	htmlContent := ""
	switch function {
	case "main_menu":
		// ดำเนินการสำหรับ main_menu และ menu-trans
		htmlContent = TransactionInfoHtml(
			shopId, mode,
		)
	default:
		// กรณีที่ function ไม่ตรงกับที่กำหนด
		htmlContent = fmt.Sprintf("<html><body><h1>Function '%s' not recognized</h1></body></html>", html.EscapeString(function))
	}

	logger.Debug("Generated simple HTML content for shopId: %s", shopId)
	return "<!DOCTYPE html>" + htmlContent
}
