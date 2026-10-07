package httpapi

import (
	"strings"
	"testing"

	gl "smlcloudplatform/internal/generalledger"
)

// The exact JSON the purchase VAT adjustment screen sends (frontend/src/app/gl/gl-vat-adjust.tsx vatAdjustRowsCommand / vatAdjustInvoiceDateCommand).
const frontendVatAdjustRows = `{
  "resource": "vat-adjustments",
  "action": "update",
  "requestid": "0f8c1c0a-4a3b-4f2e-9d1a-7c6b5e4d3c2b",
  "reason": "ย้ายงวดตามใบกำกับภาษีที่ได้รับ",
  "vatadjust": {
    "mode": "rows",
    "items": [{"journalid": "j1", "version": 3, "vatid": "v1", "taxperiodyear": 2026, "taxperiodmonth": 2, "claimstatus": 1}]
  }
}`

const frontendVatAdjustInvoiceDate = `{
  "resource": "vat-adjustments",
  "action": "update",
  "requestid": "0f8c1c0a-4a3b-4f2e-9d1a-7c6b5e4d3c2b",
  "reason": "ตั้งงวดตามวันที่ใบกำกับภาษี",
  "vatadjust": {
    "mode": "invoice_date",
    "filter": {"invdatefrom": "2026-01-01", "invdateto": "2026-01-31", "partnerfrom": "AP001"},
    "expectedcount": 12
  }
}`

func TestVatAdjustBodiesDecodeWithCurrentSource(t *testing.T) {
	cmd, err := decodeCommand(strings.NewReader(frontendVatAdjustRows))
	if err != nil || cmd.VatAdjust == nil || cmd.VatAdjust.Mode != gl.VatAdjustModeRows || len(cmd.VatAdjust.Items) != 1 || cmd.VatAdjust.Items[0].Version != 3 || cmd.VatAdjust.Items[0].TaxPeriodMonth != 2 {
		t.Fatalf("rows body: %+v err %v", cmd.VatAdjust, err)
	}
	cmd, err = decodeCommand(strings.NewReader(frontendVatAdjustInvoiceDate))
	if err != nil || cmd.VatAdjust == nil || cmd.VatAdjust.Filter.InvDateTo != "2026-01-31" || cmd.VatAdjust.Filter.PartnerFrom != "AP001" || cmd.VatAdjust.ExpectedCount != 12 {
		t.Fatalf("invoice date body: %+v err %v", cmd.VatAdjust, err)
	}
	if _, err = decodeCommand(strings.NewReader(`{"resource":"vat-adjustments","action":"update","vatadjust":{"mode":"rows","items":[{"journalid":"j","amount":1}]}}`)); err == nil {
		t.Fatal("an unknown item field must be rejected")
	}
}

// Reading the list needs the screen, saving needs its update right; the demo owner ("*") has both.
func TestVatAdjustPermissions(t *testing.T) {
	if resourceScreens[gl.VatAdjustKind] != "purchase-vat-adjust" || vatAdjustScreen != "purchase-vat-adjust" {
		t.Fatalf("screen = %q", resourceScreens[gl.VatAdjustKind])
	}
	screen := resourceScreens[gl.VatAdjustKind]
	if !allowed(map[string]bool{"*": true}, screen, "") || !allowed(map[string]bool{"*": true}, screen, "update") {
		t.Fatal("owner denied")
	}
	readOnly := map[string]bool{screen: true}
	if !allowed(readOnly, screen, "") || allowed(readOnly, screen, "update") {
		t.Fatal("screen entry must not grant saving")
	}
	if allowed(map[string]bool{"gl-journals": true, "gl-journals:update": true}, screen, "update") {
		t.Fatal("journal rights must not open the VAT adjustment screen")
	}
}
