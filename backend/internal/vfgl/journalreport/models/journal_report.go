package models

import (
	chartofaccountModel "smlcloudplatform/internal/vfgl/chartofaccount/models"
	"time"
)

type TrialBalanceSheetReport struct {
	ReportDate             time.Time                         `json:"reportdate"`             // วันที่ทำรายการ
	StartDate              time.Time                         `json:"startdate"`              // วันที่เริ่มต้น
	EndDate                time.Time                         `json:"enddate"`                // วันที่สิ้นสุด
	AccountGroup           string                            `json:"accountgroup"`           // เล่มบัญชี
	AccountDetails         *[]TrialBalanceSheetAccountDetail `json:"accountdetails"`         // รายละเอียดบัญชี
	TotalBalanceDebit      float64                           `json:"totalbalancedebit"`      // รวมยอดยกมาเดบิต
	TotalBalanceCredit     float64                           `json:"totalbalancecredit"`     // รวมยอดยกมาเครดิต
	TotalAmountDebit       float64                           `json:"totalamountdebit"`       // รวมเดบิต
	TotalAmountCredit      float64                           `json:"totalamountcredit"`      // รวมเครดิต
	TotalNextBalanceDebit  float64                           `json:"totalnextbalancedebit"`  // รวมยอดสะสมเดบิต
	TotalNextBalanceCredit float64                           `json:"totalnextbalancecredit"` // รวมยอดสะสมเครดิต
}

type TrialBalanceSheetAccountDetail struct {
	chartofaccountModel.ChartOfAccountPG
	Amount                  float64 `json:"amount" gorm:"column:amount"`                                // ยอดคงเหลือ(ประจำงวด)
	BalanceAmount           float64 `json:"balanceamount" gorm:"column:balanceamount"`                  // ยอดคงเหลือยกมา
	NextBalanceAmount       float64 `json:"nextbalanceamount" gorm:"column:nextbalanceamount"`          // ยอดคงเหลือสะสม
	DebitAmount             float64 `json:"debitamount" gorm:"-"`                                       // ยอดเดบิต
	CreditAmount            float64 `json:"creditamount" gorm:"-"`                                      // ยอดเครดิต
	SumDebit                float64 `json:"sumdebit" gorm:"column:debitamount"`                         // ยอดเครดิต
	SumCredit               float64 `json:"sumcredit" gorm:"column:creditamount"`                       // ยอดเครดิต
	BalanceDebitAmount      float64 `json:"balancedebitamount" gorm:"-"`                                // ยอดยกมาเดบิต
	BalanceCreditAmount     float64 `json:"balancecreditamount" gorm:"-"`                               // ยอดยกมาเครดิต
	SumBalanceDebit         float64 `json:"sumbalancedebit" gorm:"column:balancedebitamount"`           //
	SumBalanceCredit        float64 `json:"sumbalancecredit" gorm:"column:balancecreditamount"`         //
	NextBalanceDebitAmount  float64 `json:"nextbalancedebitamount" gorm:"-"`                            // ยอดสะสมเดบิต
	NextBalanceCreditAmount float64 `json:"nextbalancecreditamount" gorm:"-"`                           // ยอดสะสมเครดิต
	SumNextBalanceDebit     float64 `json:"sumnextbalancedebit" gorm:"column:nextbalancedebitamount"`   //
	SumNextBalanceCredit    float64 `json:"sumnextbalancecredit" gorm:"column:nextbalancecreditamount"` //
}

type BalanceSheetReport struct {
	// วันที่ทำรายการ
	ReportDate time.Time `json:"reportdate"`
	// วันที่สิ้นสุด
	EndDate time.Time `json:"enddate"`
	// เล่มบัญชี
	AccountGroup string `json:"accountgroup"`
	// สินทรัพย์
	Assets *[]BalanceSheetAccountDetail `json:"assets"`
	// หนี้สิน
	Liabilities *[]BalanceSheetAccountDetail `json:"liabilities"`
	// ทุนและส่วนของเจ้าของ
	OwnesEquities *[]BalanceSheetAccountDetail `json:"ownesequities"`
	// รวมสินทรัพย์
	TotalAssetAmount float64 `json:"totalassetamount"`
	// รวมหนี้สิน
	TotalLiabilityAmount float64 `json:"totalliabilityamount"`
	// รวมทุนและส่วนของเจ้าของ
	TotalOwnersEquityAmount float64 `json:"totalownersequityamount"`
	// รวมหนี้สิน ทุน และส่วนของเจ้าของ
	TotalLiabilityAndOwnersEquityAmount float64 `json:"totalliabilityandownersequityamount"`
}

type BalanceSheetAccountDetail struct {
	chartofaccountModel.ChartOfAccountPG
	// มูลค่า
	Amount float64 `json:"amount" gorm:"column:amount"`
}

type ProfitAndLossSheetReport struct {
	// วันที่ทำรายการ
	ReportDate time.Time `json:"reportdate"`
	// วันที่เริ่มต้น
	StartDate time.Time `json:"startdate"`
	// วันที่สิ้นสุด
	EndDate time.Time `json:"enddate"`
	// เล่มบัญชี
	AccountGroup string `json:"accountgroup"`
	// รายการรายได้
	Incomes *[]ProfitAndLossSheetAccountDetail `json:"incomes"`
	// รายการค่าใช้จ่าย
	Expenses *[]ProfitAndLossSheetAccountDetail `json:"expenses"`
	// รวมรายได้
	TotalIncomeAmount float64 `json:"totalincomeamount"`
	// รวมค่าใช้จ่าย
	TotalExpenseAmount float64 `json:"totalexpenseamount"`
	// กำไรขาดทุน
	ProfitAndLossAmount float64 `json:"profitandlossamount"`
}

type ProfitAndLossSheetAccountDetail struct {
	chartofaccountModel.ChartOfAccountPG
	// มูลค่า
	Amount float64 `json:"amount"`
}

type TrialBalanceSheetReportResponse struct {
	Success bool                    `json:"success"`
	Data    TrialBalanceSheetReport `json:"data,omitempty"`
}

type BalanceSheetReportResponse struct {
	Success bool               `json:"success"`
	Data    BalanceSheetReport `json:"data,omitempty"`
}

type LostAndProfitSheetReportResponse struct {
	Success bool                     `json:"success"`
	Data    ProfitAndLossSheetReport `json:"data,omitempty"`
}

type LedgerAccountRaw struct {
	RowMode                int8      `json:"rowmode" gorm:"column:rowmode"`
	DocDate                time.Time `json:"docdate" gorm:"column:docdate"`
	DocNo                  string    `json:"docno" gorm:"column:docno"`
	AccountCode            string    `json:"accountcode" gorm:"column:accountcode"`
	AccountName            string    `json:"accountname" gorm:"column:accountname"`
	AccountDescription     string    `json:"accountdescription" gorm:"column:accountdescription"`
	AccountGroup           string    `json:"accountgroup" gorm:"column:accountgroup"`
	ConsolidateAccountCode string    `json:"consolidateaccountcode" gorm:"column:consolidateaccountcode"`
	DebitAmount            float64   `json:"debitamount" gorm:"column:debitamount"`
	CreditAmount           float64   `json:"creditamount" gorm:"column:creditamount"`
	Amount                 float64   `json:"amount" gorm:"column:amount"`
}

type LedgerAccount struct {
	AccountCode            string                 `json:"accountcode" gorm:"column:accountcode"`
	AccountName            string                 `json:"accountname" gorm:"column:accountname"`
	AccountGroup           string                 `json:"accountgroup" gorm:"column:accountgroup"`
	ConsolidateAccountCode string                 `json:"consolidateaccountcode" gorm:"column:consolidateaccountcode"`
	Balance                float64                `json:"balance" gorm:"column:balance"`
	NextBalance            float64                `json:"nextbalance" gorm:"column:nextbalance"`
	Details                *[]LedgerAccountDetail `json:"details" gorm:"column:details"`
}

type LedgerAccountDetail struct {
	DocNo              string    `json:"docno" gorm:"column:docno"`
	DocDate            time.Time `json:"docdate" gorm:"column:docdate"`
	AccountDescription string    `json:"accountdescription" gorm:"column:accountdescription"`
	Debit              float64   `json:"debit" gorm:"column:debit"`
	Credit             float64   `json:"credit" gorm:"column:credit"`
	Amount             float64   `json:"amount" gorm:"column:amount"`
	CountVat           int       `json:"countvat"`
	CountTax           int       `json:"counttax"`
	CountImage         int       `json:"countimage"`
}

type LedgerAccountCodeRange struct {
	Start string
	End   string
}

type JournalSummary struct {
	DocNo    string `json:"docno" bson:"docno"`
	CountVat int    `json:"countvat" bson:"countvat"`
	CountTax int    `json:"counttax" bson:"counttax"`
}

func (JournalSummary) CollectionName() string {
	return "journals"
}

type JournalImageSummary struct {
	DocNo      string `json:"docno" bson:"docno"`
	CountImage int    `json:"countimage" bson:"countimage"`
}

func (JournalImageSummary) CollectionName() string {
	return "documentimagegroups"
}

type JournalDocNoByShop struct {
	HoldingCode string `json:"holdingcode" bson:"holdingcode"`
	DocNo       string `json:"docno" bson:"docno"`
}

func (JournalDocNoByShop) CollectionName() string {
	return "journals"
}

// Multi-Shop Dashboard Response
type MultiShopDashboardResponse struct {
	Success bool                   `json:"success"`
	Data    []ShopDashboardSummary `json:"data"`
	Period  PeriodInfo             `json:"period"`
}

// Summary per shop
type ShopDashboardSummary struct {
	HoldingCode    string  `json:"holdingcode"`
	ShopName       string  `json:"shopname"`
	DailyAverage   float64 `json:"dailyaverage"`
	MonthlyAverage float64 `json:"monthlyaverage"`
	YearlyAverage  float64 `json:"yearlyaverage"`
	TotalRevenue   float64 `json:"totalrevenue"`
	TotalProfit    float64 `json:"totalprofit"`
	ImageCount     int     `json:"imagecount"`
}

// Period metadata
type PeriodInfo struct {
	StartDate time.Time `json:"startdate"`
	EndDate   time.Time `json:"enddate"`
	TotalDays int       `json:"totaldays"`
}

// PostgreSQL raw result (internal use)
type MultiShopRevenueRaw struct {
	HoldingCode     string  `gorm:"column:holdingcode"`
	AccountCategory int16   `gorm:"column:accountcategory"`
	TotalAmount     float64 `gorm:"column:totalamount"`
}

// MongoDB image count result
type ShopImageCount struct {
	HoldingCode string `json:"holdingcode" bson:"holdingcode"`
	ImageCount  int    `json:"imagecount" bson:"imagecount"`
}

func (ShopImageCount) CollectionName() string {
	return "documentimagegroups"
}
