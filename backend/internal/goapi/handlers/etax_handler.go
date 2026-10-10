package handlers

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v4"
	"smlcloudplatform/internal/etax"
)

type ETaxGenerateXMLRequest struct {
	Document etax.Document `json:"document"`
}

type ETaxGenerateXMLResponse struct {
	Success  bool   `json:"success"`
	XML      string `json:"xml,omitempty"`
	FileName string `json:"filename,omitempty"`
	DocNo    string `json:"doc_no,omitempty"`
	Message  string `json:"message,omitempty"`
}

// ETaxGenerateXMLHandler รับข้อมูลใบกำกับภาษี/ใบลดหนี้/ใบเพิ่มหนี้ แล้วสร้าง XML มธพ. 3-2560
func ETaxGenerateXMLHandler(c echo.Context) error {
	var req ETaxGenerateXMLRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, ETaxGenerateXMLResponse{
			Success: false,
			Message: fmt.Sprintf("invalid payload: %v", err),
		})
	}

	xmlBytes, err := etax.GenerateXML(req.Document)
	if err != nil {
		return c.JSON(http.StatusBadRequest, ETaxGenerateXMLResponse{
			Success: false,
			Message: err.Error(),
		})
	}

	fileName := fmt.Sprintf("%s_ETAX.xml", req.Document.DocNo)

	return c.JSON(http.StatusOK, ETaxGenerateXMLResponse{
		Success:  true,
		XML:      string(xmlBytes),
		FileName: fileName,
		DocNo:    req.Document.DocNo,
	})
}
