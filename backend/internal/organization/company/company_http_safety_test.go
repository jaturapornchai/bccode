package company

import (
	"encoding/json"
	"errors"
	"os"
	common "smlcloudplatform/internal/models"
	orgaccess "smlcloudplatform/internal/organization"
	companyModels "smlcloudplatform/internal/organization/company/models"
	"smlcloudplatform/internal/taxaddress"
	"strings"
	"testing"
	"time"
)

func TestCreateCompanyDoesNotSilentlyCreateDefaultBranch(t *testing.T) {
	source, err := os.ReadFile("company_http.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"ensureDefaultHeadOfficeBranch", "defaultbranchid"} {
		if strings.Contains(string(source), forbidden) {
			t.Fatalf("CreateCompany still contains implicit Branch behavior: %s", forbidden)
		}
	}
}

func TestPrepareCompanyCreateOverridesClientIdentityAndStatus(t *testing.T) {
	code, name := "th", "บริษัททดสอบ"
	req := companyModels.CompanyDoc{
		HoldingUID: "client-holding", CompanyUID: "client-company", GuidFixed: "client-guid",
		IsDeleted: true, Version: 99,
	}
	req.Code = " cmp-001 "
	req.Names = common.JSONB{{Code: &code, Name: &name}}
	now := time.Date(2026, 8, 14, 1, 2, 3, 0, time.FixedZone("test", 7*60*60))
	if err := prepareCompanyCreate(&req, "holding-code", "owner@example.com", now); err != nil {
		t.Fatal(err)
	}
	if req.HoldingUID != "holding-code" || req.CompanyUID != "CMP-001" || req.GuidFixed != "CMP-001" || req.IsDeleted || req.Version != 0 || !req.IsActive {
		t.Fatalf("unsafe saved identity/status: %#v", req)
	}
	if req.ID != "CMP-001" || req.Code != "CMP-001" {
		t.Fatalf("saved identity/code is incomplete: %#v", req)
	}
	if req.CreatedAt.Location() != time.UTC || req.UpdatedAt.Location() != time.UTC {
		t.Fatalf("timestamps are not UTC: %v %v", req.CreatedAt, req.UpdatedAt)
	}
}

func TestCompanyCodeCannotChangeThroughOrdinaryUpdate(t *testing.T) {
	if err := requireUnchangedCompanyCode(" cmp-001 ", "CMP-001"); err != nil {
		t.Fatal(err)
	}
	if err := requireUnchangedCompanyCode("CMP-001", "CMP-002"); !errors.Is(err, errCompanyCodeChange) {
		t.Fatalf("error = %v, want %v", err, errCompanyCodeChange)
	}
}

func TestCompanyCreateResponseExposesSavedEntity(t *testing.T) {
	payload, err := json.Marshal(orgaccess.OrganizationCreateResponse{Entity: companyModels.CompanyDoc{CompanyUID: "company-uid", HoldingUID: "holding-uid"}})
	if err != nil {
		t.Fatal(err)
	}
	text := string(payload)
	for _, required := range []string{`"entity"`, `"companyuid":"company-uid"`, `"holdinguid":"holding-uid"`} {
		if !strings.Contains(text, required) {
			t.Fatalf("response %s missing %s", text, required)
		}
	}
}

func TestValidateCompanyNamesRequiresAtLeastOneNonBlankLocalizedName(t *testing.T) {
	code, name := "th", "บริษัททดสอบ"
	blank := " "

	tests := []struct {
		name    string
		names   common.JSONB
		wantErr bool
	}{
		{name: "missing", wantErr: true},
		{name: "blank name", names: common.JSONB{{Code: &code, Name: &blank}}, wantErr: true},
		{name: "blank language", names: common.JSONB{{Code: &blank, Name: &name}}, wantErr: true},
		{name: "valid", names: common.JSONB{{Code: &code, Name: &name}}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := validateCompanyNames(test.names)
			if test.wantErr != (err != nil) {
				t.Fatalf("error = %v, wantErr %v", err, test.wantErr)
			}
		})
	}
}

// ที่อยู่สำหรับภาษี: ตัดช่องว่าง + ตรวจก่อนบันทึก; ช่องที่ไม่ได้ส่ง (nil) ไม่ตรวจ
func TestPrepareCompanyTaxAddressNormalizesAndValidates(t *testing.T) {
	phone := " 02-123-4567 "
	req := companyModels.CompanyDoc{}
	req.Address = &taxaddress.Address{No: " 88/12 ", Province: " ปทุมธานี ", Postcode: "12120"}
	req.Phone = &phone
	if err := prepareCompanyTaxAddress(&req); err != nil {
		t.Fatal(err)
	}
	if req.Address.No != "88/12" || req.Address.Province != "ปทุมธานี" || *req.Phone != "02-123-4567" {
		t.Fatalf("not normalized: %+v phone %q", req.Address, *req.Phone)
	}
	if err := prepareCompanyTaxAddress(&companyModels.CompanyDoc{}); err != nil {
		t.Fatalf("nil address/phone must pass: %v", err)
	}
	bad := companyModels.CompanyDoc{}
	bad.Address = &taxaddress.Address{Postcode: "1212"}
	fe := taxaddress.AsFieldError(prepareCompanyTaxAddress(&bad))
	if fe == nil || fe.Field != "addr_postcode" || fe.Reason != taxaddress.ReasonPostcode {
		t.Fatalf("bad postcode: %+v", fe)
	}
}

// PUT แทนทั้งเอกสาร แต่ไม่ส่งที่อยู่/โทรศัพท์ (client เก่า) ต้องคงค่าเดิม; ส่งมา = แทนทั้งชุด 13 ช่อง; POST ไม่ส่ง = ว่าง
func TestMergeCompanyTaxAddressKeepsExistingWhenNotSent(t *testing.T) {
	oldPhone := "02-123-4567"
	existing := companyModels.CompanyDoc{}
	existing.Address = &taxaddress.Address{No: "88/12", Road: "ลาดหลุมแก้ว", Province: "ปทุมธานี", Postcode: "12140"}
	existing.Phone = &oldPhone

	address, phone := mergeCompanyTaxAddress(existing, companyModels.CompanyDoc{})
	if address != *existing.Address || phone != oldPhone {
		t.Fatalf("nil must keep existing: %+v %q", address, phone)
	}

	newPhone := ""
	req := companyModels.CompanyDoc{}
	req.Address = &taxaddress.Address{No: "175", Province: "กรุงเทพมหานคร"}
	req.Phone = &newPhone
	address, phone = mergeCompanyTaxAddress(existing, req)
	if address != (taxaddress.Address{No: "175", Province: "กรุงเทพมหานคร"}) || phone != "" {
		t.Fatalf("sent address must replace all keys: %+v %q", address, phone)
	}

	address, phone = mergeCompanyTaxAddress(companyModels.CompanyDoc{}, companyModels.CompanyDoc{})
	if !address.IsBlank() || phone != "" {
		t.Fatalf("create without address must be blank: %+v %q", address, phone)
	}
}

func TestCompanyTaxAddressSQLShape(t *testing.T) {
	if len(companyTaxAddressColumns) != 14 || companyTaxAddressColumns[13] != "phone" {
		t.Fatalf("columns %v", companyTaxAddressColumns)
	}
	if got := companyPlaceholders(9, 3); got != "$9, $10, $11" {
		t.Fatalf("placeholders %q", got)
	}
	assign := companyTaxAddressAssignments(11)
	if !strings.HasPrefix(assign, "addr_building = $11, ") || !strings.HasSuffix(assign, "phone = $24") {
		t.Fatalf("assignments %q", assign)
	}
	if args := companyTaxAddressArgs(taxaddress.Address{Building: "A", Postcode: "10120"}, "02"); len(args) != 14 || args[0] != "A" || args[12] != "10120" || args[13] != "02" {
		t.Fatalf("args %v", args)
	}
	if !strings.HasSuffix(companyColumns, "addr_postcode, phone") {
		t.Fatalf("select columns %q", companyColumns)
	}
}

// ข้อความผิดพลาดมาจาก languages.tsv พร้อมชื่อช่อง + field ให้จอพาไปที่ช่อง (ไม่ใช่ข้อความอังกฤษดิบ)
func TestCompanyTaxAddressErrorUsesLanguageKeys(t *testing.T) {
	appErr := companyTaxAddressError(&taxaddress.FieldError{Field: "addr_road", Reason: taxaddress.ReasonTooLong, Max: 100}, "en")
	if appErr.Code != "VALIDATION_FAILED" || appErr.Field != "address.addr_road" || appErr.HTTPStatus != 400 {
		t.Fatalf("app error %+v", appErr)
	}
	if appErr.ThaiMsg != "ถนน ยาวเกิน 100 ตัวอักษร — กรุณาย่อให้สั้นลง" {
		t.Fatalf("thai message %q", appErr.ThaiMsg)
	}
	if appErr.Message != "Road is longer than 100 characters — please shorten it" {
		t.Fatalf("english message %q", appErr.Message)
	}
	phone := companyTaxAddressError(&taxaddress.FieldError{Field: "phone", Reason: taxaddress.ReasonControl}, "th")
	if phone.Field != "phone" || !strings.Contains(phone.ThaiMsg, "หมายเลขโทรศัพท์") || strings.Contains(phone.ThaiMsg, "{field}") {
		t.Fatalf("phone error %+v", phone)
	}
	postcode := companyTaxAddressError(&taxaddress.FieldError{Field: "addr_postcode", Reason: taxaddress.ReasonPostcode}, "th")
	if postcode.ThaiMsg != "รหัสไปรษณีย์ ต้องเป็นตัวเลข 5 หลัก หรือเว้นว่าง" {
		t.Fatalf("postcode %q", postcode.ThaiMsg)
	}
	for field := range companyTaxAddressLabels {
		if field != taxaddress.PhoneKey && !containsKey(taxaddress.Keys, field) {
			t.Fatalf("label for unknown field %s", field)
		}
	}
	if len(companyTaxAddressLabels) != 14 {
		t.Fatalf("labels %d", len(companyTaxAddressLabels))
	}
}

// GET/list คืน address (object ตาม key หัวแบบ) และ phone เสมอ แม้ยังว่าง — จอแยกไม่ออกระหว่าง "ว่าง" กับ "backend เก่า"
func TestCompanyJSONAlwaysCarriesTaxAddress(t *testing.T) {
	address, phone := taxaddress.Address{Province: "ปทุมธานี"}, ""
	doc := companyModels.CompanyDoc{}
	doc.Address, doc.Phone = &address, &phone
	payload, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"address":{"addr_building":""`, `"addr_province":"ปทุมธานี"`, `"phone":""`} {
		if !strings.Contains(string(payload), want) {
			t.Fatalf("payload %s missing %s", payload, want)
		}
	}
	var req companyModels.CompanyDoc
	if err := json.Unmarshal([]byte(`{"code":"01","names":[]}`), &req); err != nil || req.Address != nil || req.Phone != nil {
		t.Fatalf("absent address must decode as nil: %+v %v", req.Address, err)
	}
}

func containsKey(keys []string, key string) bool {
	for _, k := range keys {
		if k == key {
			return true
		}
	}
	return false
}
