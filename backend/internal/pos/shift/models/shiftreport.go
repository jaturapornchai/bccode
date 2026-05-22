package models

import (
	"smlcloudplatform/internal/models"
	saleinvoicemodels "smlcloudplatform/internal/transaction/saleinvoice/models"
	"time"
)

type ShiftReport struct {
	Shifts []Shift                             `json:"shifts"`
	SaleInvoices []saleinvoicemodels.SaleInvoiceInfo `json:"saleinvoices"`
	Summary ShiftReportSummary                  `json:"summary"`
	CashDrawer ShiftReportCashDrawerSummary        `json:"cashdrawer"`
	Movements []ShiftReportMovement               `json:"movements"`
}

type ShiftReportSummary struct {
	TotalAmount float64 `json:"total_amount"`
	Cash float64 `json:"cash"`
	Credit float64 `json:"credit"`
	Transfer float64 `json:"transfer"`
	Cheque float64 `json:"cheque"`
	Coupon float64 `json:"coupon"`
	QRCode float64 `json:"qrcode"`
}

type ShiftReportCashDrawerSummary struct {
	Open float64 `json:"open"`
	Add float64 `json:"add"`
	Withdraw float64 `json:"withdraw"`
	Close float64 `json:"close"`
	Expected float64 `json:"expected"`
	Actual float64 `json:"actual"`
	Diff float64 `json:"diff"`
}

type ShiftReportMovement struct {
	MovementType string          `json:"movementtype"`
	DocNo string          `json:"docno"`
	DocDatetime time.Time       `json:"docdatetime"`
	DocType int8            `json:"doc_type,omitempty"`
	UserCode string          `json:"usercode,omitempty"`
	Username string          `json:"username,omitempty"`
	Remark string          `json:"remark,omitempty"`
	Amount float64         `json:"amount,omitempty"`
	CustCode string          `json:"custcode,omitempty"`
	CustNames *[]models.NameX `json:"custnames,omitempty"`
	SaleCode string          `json:"salecode,omitempty"`
	SaleName string          `json:"salename,omitempty"`
	MemberCode string          `json:"membercode,omitempty"`
	IsCancel bool            `json:"iscancel,omitempty"`
	GetPoint float64         `json:"getpoint,omitempty"`
	UsePoint float64         `json:"usepoint,omitempty"`
	PaymentDetailRaw string          `json:"paymentdetailraw,omitempty"`
	PayCashAmount float64         `json:"paycashamount,omitempty"`
	PayCashChange float64         `json:"paycashchange,omitempty"`
	SumQRCode float64         `json:"sumqrcode,omitempty"`
	SumCreditCard float64         `json:"sumcreditcard,omitempty"`
	SumMoneyTransfer float64         `json:"summoneytransfer,omitempty"`
	SumCheque float64         `json:"sumcheque,omitempty"`
	SumCoupon float64         `json:"sumcoupon,omitempty"`
	DetailTotalDiscount float64         `json:"detailtotaldiscount,omitempty"`
	DetailDiscountFormula string          `json:"detaildiscountformula,omitempty"`
	RoundAmount float64         `json:"roundamount,omitempty"`
	TotalQty float64         `json:"totalqty,omitempty"`
	TotalAmount float64         `json:"total_amount,omitempty"`
}
