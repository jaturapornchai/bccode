package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	msmodels "smlcloudplatform/pkg/microservice/models"

	"github.com/labstack/echo/v4"
)

type productV2ErrorBody struct {
	Success bool                  `json:"success"`
	Code    string                `json:"code"`
	Message string                `json:"message"`
	Fields  []productV2FieldError `json:"fields"`
}

func stubProductV2Permission(t *testing.T, allowed bool, err error) {
	t.Helper()
	original := productV2CheckPermission
	productV2CheckPermission = func(context.Context, msmodels.UserInfo, string) (bool, error) { return allowed, err }
	t.Cleanup(func() { productV2CheckPermission = original })
}

func callProductV2(t *testing.T, handler echo.HandlerFunc, body string, user *msmodels.UserInfo) (*httptest.ResponseRecorder, productV2ErrorBody) {
	t.Helper()
	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(body))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)
	if user != nil {
		ctx.Set("UserInfo", *user)
	}
	if err := handler(ctx); err != nil {
		t.Fatalf("handler error: %v", err)
	}
	var parsed productV2ErrorBody
	if err := json.Unmarshal(rec.Body.Bytes(), &parsed); err != nil {
		t.Fatalf("decode response %q: %v", rec.Body.String(), err)
	}
	return rec, parsed
}

var productV2TestUser = msmodels.UserInfo{Username: "tester", HoldingCode: "h1", BusinessCode: "b1", UID: "u1"}

func TestScanProductV2Payload_EveryAccountingFieldRejected(t *testing.T) {
	if len(productAccountingFields) < 60 {
		t.Fatalf("expected reflection to collect the accounting keys, got %d", len(productAccountingFields))
	}
	for _, must := range []string{"code", "names", "guidfixed", "unitcode", "unitnames", "unitconversions", "standvalue", "dividevalue", "condition", "itemtype", "materialtype", "taxtype", "vattype", "groupcode", "categorycode", "brandcode", "bom", "isusesubbarcodes", "refbarcodes", "orderpoint", "minpoint", "maxpoint", "isactive", "createdat", "updatedby", "deletedat", "isdeleted", "_id", "id"} {
		if _, ok := productAccountingFields[must]; !ok {
			t.Errorf("accounting set missing %q", must)
		}
	}
	for key := range productAccountingFields {
		accounting, unknown := scanProductV2Payload(map[string]any{"itemcode": "X", "__v": 1, key: "v"})
		if len(accounting) != 1 || accounting[0] != key || len(unknown) != 0 {
			t.Errorf("key %q: accounting=%v unknown=%v", key, accounting, unknown)
		}
	}
	for key := range productV2WritableTop {
		if _, bad := productAccountingFields[key]; bad {
			t.Errorf("writable key %q must not be in accounting set", key)
		}
	}
}

func TestScanProductV2Payload_UnknownAndPackageTopLevel(t *testing.T) {
	accounting, unknown := scanProductV2Payload(map[string]any{"itemcode": "X", "foo": 1, "packageweight": 1.5})
	if len(accounting) != 0 {
		t.Fatalf("unexpected accounting %v", accounting)
	}
	if strings.Join(unknown, ",") != "foo,packageweight" {
		t.Fatalf("unknown = %v", unknown)
	}
}

func TestUpdateListing_AccountingFieldReadonly422(t *testing.T) {
	stubProductV2Permission(t, true, nil)
	rec, body := callProductV2(t, ProductV2ItemUpdateListingHandler, `{"itemcode":"SHIRT-001","__v":0,"unitcode":"PCS","names":[],"condition":true}`, &productV2TestUser)
	if rec.Code != http.StatusUnprocessableEntity || body.Code != "ACCOUNTING_FIELD_READONLY" {
		t.Fatalf("status=%d body=%+v", rec.Code, body)
	}
	got := make([]string, 0, len(body.Fields))
	for _, f := range body.Fields {
		got = append(got, f.Field)
	}
	if strings.Join(got, ",") != "condition,names,unitcode" {
		t.Fatalf("fields = %v", got)
	}
	for _, label := range []string{"หน่วยนับ", "ชื่อสินค้า"} {
		if !strings.Contains(body.Message, label) {
			t.Errorf("message %q lacks %q", body.Message, label)
		}
	}
}

func TestUpdateListing_UnknownField422(t *testing.T) {
	stubProductV2Permission(t, true, nil)
	cases := map[string]struct{ payload, field, code string }{
		"top-level":     {`{"itemcode":"X","__v":0,"foo":1}`, "foo", "UNKNOWN_FIELD"},
		"nested":        {`{"itemcode":"X","__v":0,"listing":{"foo":1}}`, "foo", "UNKNOWN_FIELD"},
		"packageweight": {`{"itemcode":"X","__v":0,"packageweight":1}`, "packageweight", "INVALID_FIELD"},
		"listing.tiers": {`{"itemcode":"X","__v":0,"listing":{"tiers":[]}}`, "listing.tiers", "READONLY"},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			rec, body := callProductV2(t, ProductV2ItemUpdateListingHandler, tc.payload, &productV2TestUser)
			if rec.Code != http.StatusUnprocessableEntity || body.Code != "VALIDATION_FAILED" || len(body.Fields) != 1 || body.Fields[0].Code != tc.code || body.Fields[0].Field != tc.field {
				t.Fatalf("status=%d body=%+v", rec.Code, body)
			}
		})
	}
}

func TestUpdateListing_BodyTooLarge413(t *testing.T) {
	stubProductV2Permission(t, true, nil)
	payload := `{"itemcode":"X","__v":0,"description":"` + strings.Repeat("a", productV2MaxBodyBytes) + `"}`
	rec, body := callProductV2(t, ProductV2ItemUpdateListingHandler, payload, &productV2TestUser)
	if rec.Code != http.StatusRequestEntityTooLarge || body.Code != "PAYLOAD_TOO_LARGE" {
		t.Fatalf("status=%d body=%+v", rec.Code, body)
	}
}

func TestUpdateListing_RequiredKeys(t *testing.T) {
	stubProductV2Permission(t, true, nil)
	rec, body := callProductV2(t, ProductV2ItemUpdateListingHandler, `{"itemcode":"X"}`, &productV2TestUser)
	if rec.Code != http.StatusUnprocessableEntity || body.Fields[0].Field != "__v" {
		t.Fatalf("status=%d body=%+v", rec.Code, body)
	}
	rec, body = callProductV2(t, ProductV2ItemUpdateListingHandler, `{"__v":1}`, &productV2TestUser)
	if rec.Code != http.StatusUnprocessableEntity || body.Fields[0].Field != "itemcode" {
		t.Fatalf("status=%d body=%+v", rec.Code, body)
	}
	rec, body = callProductV2(t, ProductV2ItemUpdateListingHandler, `not json`, &productV2TestUser)
	if rec.Code != http.StatusBadRequest || body.Code != "INVALID_JSON" {
		t.Fatalf("status=%d body=%+v", rec.Code, body)
	}
}

func TestUpdateListing_CompanyScope(t *testing.T) {
	rec, body := callProductV2(t, ProductV2ItemUpdateListingHandler, `{"itemcode":"X","__v":0}`, nil)
	if rec.Code != http.StatusUnauthorized || body.Message != productV2MsgUnauthorized {
		t.Fatalf("status=%d body=%+v", rec.Code, body)
	}
	rec, body = callProductV2(t, ProductV2ItemUpdateListingHandler, `{"itemcode":"X","__v":0,"holdingcode":"other"}`, &productV2TestUser)
	if rec.Code != http.StatusForbidden || body.Code != "FORBIDDEN" {
		t.Fatalf("status=%d body=%+v", rec.Code, body)
	}
	noCompany := msmodels.UserInfo{HoldingCode: "h1"}
	rec, body = callProductV2(t, ProductV2ItemUpdateListingHandler, `{"itemcode":"X","__v":0}`, &noCompany)
	if rec.Code != http.StatusConflict || body.Code != "COMPANY_REQUIRED" {
		t.Fatalf("status=%d body=%+v", rec.Code, body)
	}
}
