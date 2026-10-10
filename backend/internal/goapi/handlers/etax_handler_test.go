package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shopspring/decimal"
	"smlcloudplatform/internal/etax"
)

func TestETaxGenerateXMLHandler_Success(t *testing.T) {
	e := echo.New()
	reqBody := ETaxGenerateXMLRequest{
		Document: etax.Document{
			DocNo:         "INV6910-0012",
			TypeCode:      etax.TypeInvoice,
			TypeName:      "ใบกำกับภาษี",
			IssueDateTime: time.Now(),
			CurrencyCode:  "THB",
			Seller: etax.Party{
				Name:     "บริษัท รุ่งเรืองค้าวัสดุก่อสร้าง จำกัด",
				TaxID:    "0105560012345",
				BranchID: "00000",
				Address: etax.Address{
					Postcode:    "10110",
					LineOne:     "ถนนสุขุมวิท กรุงเทพฯ",
					CountryCode: "TH",
				},
			},
			Buyer: etax.Party{
				Name:     "บริษัท สยามพัฒนาการก่อสร้าง จำกัด (มหาชน)",
				TaxID:    "0107537000123",
				BranchID: "00000",
				Address: etax.Address{
					Postcode:    "10330",
					LineOne:     "ถนนพระราม 9 กรุงเทพฯ",
					CountryCode: "TH",
				},
			},
			Items: []etax.LineItem{
				{
					LineID:    1,
					ItemCode:  "P001",
					ItemName:  "ปูนซีเมนต์",
					Quantity:  decimal.NewFromInt(10),
					UnitPrice: decimal.NewFromFloat(150.00),
					Amount:    decimal.NewFromFloat(1500.00),
					NetAmount: decimal.NewFromFloat(1500.00),
					TaxRate:   decimal.NewFromFloat(7.00),
					TaxAmount: decimal.NewFromFloat(105.00),
				},
			},
			LineTotalAmount: decimal.NewFromFloat(1500.00),
			TaxBasisTotal:   decimal.NewFromFloat(1500.00),
			TaxRate:         decimal.NewFromFloat(7.00),
			TaxTotal:        decimal.NewFromFloat(105.00),
			GrandTotal:      decimal.NewFromFloat(1605.00),
		},
	}

	payload, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/report/tax/etax/generate-xml", bytes.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := ETaxGenerateXMLHandler(c)
	if err != nil {
		t.Fatalf("unexpected handler error: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d: %s", rec.Code, rec.Body.String())
	}

	var res ETaxGenerateXMLResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &res); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if !res.Success {
		t.Fatalf("expected success true, got false")
	}
	if !strings.Contains(res.XML, "ER3-2560") {
		t.Errorf("expected XML to contain ER3-2560")
	}
	if res.DocNo != "INV6910-0012" {
		t.Errorf("expected doc_no INV6910-0012, got %s", res.DocNo)
	}
	if res.FileName != "INV6910-0012_ETAX.xml" {
		t.Errorf("expected filename INV6910-0012_ETAX.xml, got %s", res.FileName)
	}
}

func TestETaxGenerateXMLHandler_ValidationError(t *testing.T) {
	e := echo.New()
	reqBody := ETaxGenerateXMLRequest{
		Document: etax.Document{
			DocNo: "INV001",
			// ขาด TaxID และ Items
		},
	}

	payload, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/report/tax/etax/generate-xml", bytes.NewReader(payload))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	err := ETaxGenerateXMLHandler(c)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}
