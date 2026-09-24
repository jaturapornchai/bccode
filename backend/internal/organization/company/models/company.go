package models

import (
	"strings"
	"time"

	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/taxaddress"
)

type Company struct {
	Code     string       `json:"code"`
	Names    common.JSONB `json:"names"`
	TaxID    string       `json:"taxid"`
	LogoURI  string       `json:"logouri"`
	IsActive bool         `json:"isactive"`
	// ที่อยู่สำหรับภาษี (สำนักงานใหญ่) + โทรศัพท์ สำหรับหัวแบบกรมสรรพากรและ 50 ทวิ: GET คืนเสมอ,
	// POST ไม่ส่ง = ว่าง, PUT ไม่ส่ง (nil) = คงค่าเดิม / ส่งมา = แทนทั้งชุด (company_http.go)
	Address *taxaddress.Address `json:"address,omitempty"`
	Phone   *string             `json:"phone,omitempty"`
}

type CompanyDoc struct {
	ID          string `json:"id"`
	Version     int64  `json:"__v"`
	HoldingCode string `json:"holdingcode"`
	HoldingUID  string `json:"holdinguid"`
	GuidFixed   string `json:"guidfixed"`
	CompanyUID  string `json:"companyuid"`
	IsDeleted   bool   `json:"isdeleted"`
	Company
	CreatedAt time.Time  `json:"createdat"`
	UpdatedAt time.Time  `json:"updatedat"`
	DeletedAt *time.Time `json:"deletedat,omitempty"`
	CreatedBy string     `json:"createdby,omitempty"`
	UpdatedBy string     `json:"updatedby,omitempty"`
	DeletedBy string     `json:"deletedby,omitempty"`
}

func NormalizeCompanyCode(value string) string {
	return strings.ToUpper(strings.TrimSpace(value))
}
