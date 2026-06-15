package models

import (
	"smlcloudplatform/internal/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const chartOfAccountCollectionName = "chartofaccounts"

type ChartOfAccount struct {
	// รหัสผังบัญชี
	AccountCode string `json:"accountcode" bson:"accountcode"`
	// ชื่อบัญชี
	AccountName string `json:"accountname" bson:"accountname"`

	// ผลกระทบต่องบการเงิน 1=ทั่วไป, 2=ต้นทุนขาย
	FinancialStatements int16 `json:"financialstatements" bson:"financialstatements"`

	// หมวดบัญชี 1=สินทรัพย์, 2=หนี้สิน, 3=ทุน, 4=รายได้, 5=ค่าใช้จ่าย
	AccountCategory int16 `json:"accountcategory" bson:"accountcategory"`
	// ด้านบัญชี 1=เดบิต,2=เครดิต
	AccountBalanceType int16 `json:"accountbalancetype" bson:"accountbalancetype"`
	// กลุ่มบัญชี
	AccountGroup string `json:"accountgroup" bson:"accountgroup"`
	// ระดับบัญชี 0=บัญชีย่อย, มากกว่า 0 คือแต่ละระดับ
	AccountLevel int16 `json:"accountlevel" bson:"accountlevel"`
	// รหัสผังบัญชีกลาง
	ConsolidateAccountCode string `json:"consolidateaccountcode" bson:"consolidateaccountcode"`
}

type ChartOfAccountIndentityId struct {
	AccountCode string `json:"accountcode" bson:"accountcode" gorm:"accountcode"`
}

func (ChartOfAccountIndentityId) CollectionName() string {
	return chartOfAccountCollectionName
}

type ChartOfAccountInfo struct {
	models.DocIdentity `bson:"inline" gorm:"embedded;"`
	ChartOfAccount     `bson:"inline" gorm:"embedded;"`
}

func (ChartOfAccountInfo) CollectionName() string {
	return chartOfAccountCollectionName
}

type ChartOfAccountData struct {
	models.HoldingCodeentity `bson:"inline" gorm:"embedded;"`
	ChartOfAccountInfo       `bson:"inline" gorm:"embedded;"`
}

type ChartOfAccountDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ChartOfAccountData `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
	models.LastUpdate  `bson:"inline"`
}

func (ChartOfAccountDoc) CollectionName() string {
	return chartOfAccountCollectionName
}

type ChartOfAccountActivity struct {
	ChartOfAccountData `bson:"inline"`
	CreatedAt          *time.Time `json:"createdat,omitempty" bson:"createdat,omitempty"`
	UpdatedAt          *time.Time `json:"updatedat,omitempty" bson:"updatedat,omitempty"`
	DeletedAt          *time.Time `json:"deletedat,omitempty" bson:"deletedat,omitempty"`
}

func (ChartOfAccountActivity) CollectionName() string {
	return chartOfAccountCollectionName
}

type ChartOfAccountPageResponse struct {
	Success    bool                          `json:"success"`
	Data       []ChartOfAccountInfo          `json:"data,omitempty"`
	Pagination models.PaginationDataResponse `json:"pagination,omitempty"`
}

type ChartOfAccountInfoResponse struct {
	Success bool               `json:"success"`
	Data    ChartOfAccountInfo `json:"data,omitempty"`
}

type ChartOfAccountPG struct {
	models.HoldingCodeentity `gorm:"embedded;"`
	models.PartitionIdentity `gorm:"embedded;"`
	// รหัสผังบัญชี
	AccountCode string `json:"accountcode" gorm:"column:accountcode;primaryKey"`
	// ชื่อบัญชี
	AccountName string `json:"accountname" gorm:"column:accountname"`
	// ผลกระทบต่องบการเงิน 1=ทั่วไป, 2=ต้นทุนขาย
	FinancialStatements int16 `json:"financialstatements" bson:"financialstatements"`
	// หมวดบัญชี 1=สินทรัพย์, 2=หนี้สิน, 3=ทุน, 4=รายได้, 5=ค่าใช้จ่าย
	AccountCategory int16 `json:"accountcategory" gorm:"column:accountcategory"`
	// ด้านบัญชี 1=เดบิต,2=เครดิต
	AccountBalanceType int16 `json:"accountbalancetype" gorm:"column:accountbalancetype"`
	// กลุ่มบัญชี
	AccountGroup string `json:"accountgroup" gorm:"column:accountgroup"`
	// ระดับบัญชี 0=บัญชีย่อย, มากกว่า 0 คือแต่ละระดับ
	AccountLevel int16 `json:"accountlevel" gorm:"column:accountlevel"`
	// รหัสผังบัญชีกลาง
	ConsolidateAccountCode string `json:"consolidateaccountcode" gorm:"column:consolidateaccountcode"`
}

func (ChartOfAccountPG) TableName() string {
	return "chartofaccounts"
}

type AccountCodeRange struct {
	Start string
	End   string
}
