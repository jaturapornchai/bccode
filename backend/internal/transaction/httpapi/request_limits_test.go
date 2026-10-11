package httpapi

import (
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"gopkg.in/DATA-DOG/go-sqlmock.v1"

	gl "smlcloudplatform/internal/generalledger"
	glhttp "smlcloudplatform/internal/generalledger/httpapi"
	"smlcloudplatform/pkg/apperr"
	"smlcloudplatform/pkg/microservice"
)

// writeHandler: Http ที่ผ่าน authorize ได้ (สิทธิ์ "*", สร้างตารางเอกสาร+ตารางบัญชีอย่างละหนึ่งธุรกรรม) แต่ไม่มี Begin ที่สาม —
// ถ้าคำขอหลุดไปถึง SaveDoc การเปิดธุรกรรมจะล้มเป็น 500
func writeHandler(t *testing.T) *Http {
	t.Helper()
	db, mock, _ := newRecordedMock(t)
	mock.MatchExpectationsInOrder(false)
	for i := 0; i < 2; i++ {
		mock.ExpectBegin()
		mock.ExpectCommit()
	}
	for i := 0; i < 16; i++ {
		mock.ExpectExec(".").WillReturnResult(sqlmock.NewResult(0, 0))
	}
	return &Http{
		store: NewStore(),
		session: func(context.Context, microservice.IContext) (glhttp.SessionAccess, error) {
			return glhttp.NewSessionAccess(testScope(), "*"), nil
		},
		connect: func(string) (*sql.DB, error) { return db, nil },
	}
}

// Rule 10: decimal แปลงตัวเลขยาว n หลักใช้เวลา O(n^2) (ล้านหลัก ≈ 3 วินาทีต่อช่อง ทั้งแบบข้อความและตัวเลข JSON)
// คำขอบันทึก/แก้เอกสารจึงต้องตัดตัวเลขที่ยาวเกินก่อนแปลง ตอบ 400 พร้อมชื่อช่องทันที และปฏิเสธ body ที่ใหญ่เกินกำหนด
func TestWriteHandlersRejectOversizedInputFast(t *testing.T) {
	digits := "1" + strings.Repeat("7", 1_000_000)
	cases := []struct {
		name, method, body, field, code string
	}{
		{"create-qty-string", http.MethodPost, `{"details":[{"qty":"` + digits + `","price":"1"}]}`, "details[0].qty", ""},
		{"create-qty-number", http.MethodPost, `{"details":[{"qty":1},{"qty":` + digits + `}]}`, "details[1].qty", ""},
		{"update-total-string", http.MethodPut, `{"totalamount":"` + digits + `"}`, "totalamount", ""},
		{"update-vatrate-number", http.MethodPut, `{"vatrate":0.` + digits + `}`, "vatrate", ""},
		{"create-body-over-limit", http.MethodPost, `{"description":"` + strings.Repeat("ก", maxDocBodyBytes/3+1) + `"}`, "", "erp_err_invalid_payload"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := writeHandler(t)
			request := newPermContext("sale-invoice", "?lang=en")
			request.req = httptest.NewRequest(tc.method, "/transaction/sale-invoice?lang=en", strings.NewReader(tc.body))
			handle := h.create
			if tc.method == http.MethodPut {
				request.params["id"] = "IV6910-0001"
				handle = h.update
			}
			started := time.Now()
			if err := handle(request); err != nil {
				t.Fatal(err)
			}
			if elapsed := time.Since(started); elapsed > 200*time.Millisecond {
				t.Errorf("took %v, want < 200ms (a quadratic parse took seconds)", elapsed)
			}
			response, ok := request.body.(apperr.Response)
			if request.status != http.StatusBadRequest || !ok {
				t.Fatalf("response = %d %#v, want 400", request.status, request.body)
			}
			if response.Field != tc.field || (tc.code != "" && response.Code != tc.code) {
				t.Errorf("field/code = %q/%q, want %q/%q", response.Field, response.Code, tc.field, tc.code)
			}
		})
	}
}

// การตรวจความยาวต้องไม่เปลี่ยนสัญญาเดิม: ตัวเลขรับได้ทั้งตัวเลข JSON และข้อความ, null/ไม่มีช่อง = ศูนย์, ช่องอื่นอ่านครบ,
// ข้อความที่ไม่ใช่ตัวเลข = error ธรรมดา (handler ตอบ erp_err_invalid_payload)
func TestDocJSONDecodeKeepsContract(t *testing.T) {
	var doc ErpTransactionDoc
	body := `{"docno":"IV6910-0001","totalamount":107.5,"vatrate":"7","totalvatvalue":null,
		"details":[{"itemcode":"A01","qty":"0.1","price":0.2,"sumamount":null,"whcode":"WH1"},{"itemcode":"A02","qty":"1e2"}]}`
	if err := json.Unmarshal([]byte(body), &doc); err != nil {
		t.Fatal(err)
	}
	if doc.DocNo != "IV6910-0001" || doc.TotalAmount.String() != "107.5" || doc.VatRate.String() != "7" || !doc.TotalVatValue.IsZero() {
		t.Errorf("header = %+v", doc)
	}
	if len(doc.Details) != 2 || doc.Details[0].ItemCode != "A01" || doc.Details[0].WhCode != "WH1" ||
		doc.Details[0].Qty.String() != "0.1" || doc.Details[0].Price.String() != "0.2" || !doc.Details[0].SumAmount.IsZero() ||
		doc.Details[1].Qty.String() != "100" {
		t.Errorf("details = %+v", doc.Details)
	}
	if !doc.Details[0].Qty.Add(doc.Details[0].Price).Equal(amt("0.3")) {
		t.Error("0.1 + 0.2 must be exactly 0.3")
	}
	var empty ErpTransactionDoc
	if err := json.Unmarshal([]byte(`{"docno":"X"}`), &empty); err != nil || empty.Details != nil {
		t.Errorf("missing details must stay nil: %v %+v", err, empty.Details)
	}
	var bad ErpTransactionDoc
	err := json.Unmarshal([]byte(`{"details":[{"qty":"abc"}]}`), &bad)
	if _, isUser := gl.AsUserError(err); err == nil || isUser {
		t.Errorf("non-numeric qty must be a plain decode error, got %v", err)
	}
}
