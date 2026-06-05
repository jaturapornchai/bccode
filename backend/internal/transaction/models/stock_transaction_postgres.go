package models

import (
	"smlcloudplatform/internal/models"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type StockTransaction struct {
	//TaxDocNo            string                    `json:"taxdocno" gorm:"column:taxdocno"`
	//TaxDocDate          time.Time                 `json:"taxdocdate" gorm:"column:taxdocno"`
	//SaleCode		string                    `json:"salecode" gorm:"column:salecode"`
	models.HoldingCodeentity `bson:"inline"`
	models.PartitionIdentity `gorm:"embedded;"`
	GuidFixed                string                    `json:"guidfixed" gorm:"column:guidfixed"`
	DocNo                    string                    `json:"docno" gorm:"column:docno;primaryKey"`
	DocDate                  time.Time                 `json:"docdate" gorm:"column:docdate"`
	GuidRef                  string                    `json:"guidref" gorm:"column:guidref"`
	DocRefType               int8                      `json:"docreftype" gorm:"column:docreftype"`
	DocRefNo                 string                    `json:"docrefno" gorm:"column:docrefno"`
	DocRefDate               time.Time                 `json:"docrefdate" gorm:"column:docrefdate"`
	BranchCode               string                    `json:"branchcode" gorm:"column:branchcode"`
	BranchNames              models.JSONB              `json:"branchnames" gorm:"column:branchnames;type:jsonb"`
	InquiryType              int                       `json:"inquirytype" gorm:"column:inquirytype"`
	TransFlag                int16                     `json:"transflag" gorm:"column:transflag" `
	VatType                  int8                      `json:"vattype" gorm:"column:vattype" `
	VatRate                  float64                   `json:"vatrate" gorm:"column:vatrate"`
	Details                  *[]StockTransactionDetail `json:"details" gorm:"details;foreignKey:holdingcode,docno"`
	Description              string                    `json:"description" gorm:"column:description"`
	TotalValue               float64                   `json:"totalvalue" gorm:"column:totalvalue"`
	DiscountWord             string                    `json:"discountword" gorm:"column:discountword"`
	TotalDiscount            float64                   `json:"totaldiscount" gorm:"column:totaldiscount"`
	TotalBeforeVat           float64                   `json:"totalbeforevat" gorm:"column:totalbeforevat"`
	TotalVatValue            float64                   `json:"totalvatvalue" gorm:"column:totalvatvalue"`
	TotalExceptVat           float64                   `json:"totalexceptvat" gorm:"column:totalexceptvat"`
	TotalAfterVat            float64                   `json:"totalaftervat" gorm:"column:totalaftervat"`
	TotalAmount              float64                   `json:"totalamount" gorm:"column:totalamount"`
	TotalCost                float64                   `json:"totalcost" gorm:"column:totalcost"`
	Status                   int8                      `json:"status" gorm:"column:status"`
	IsCancel                 bool                      `json:"iscancel" gorm:"column:iscancel"`
	PosID                    string                    `json:"posid" gorm:"column:posid"`
}

func (StockTransaction) TableName() string {
	return "stocktransaction"
}

type StockTransactionDetail struct {
	ID                       uint   `gorm:"primarykey"`
	HoldingCode              string `json:"holdingcode" gorm:"column:holdingcode"`
	models.PartitionIdentity `gorm:"embedded;"`
	DocNo                    string  `json:"docno" gorm:"column:docno"`
	Barcode                  string  `json:"barcode" gorm:"column:barcode"`
	RefBarcode               string  `json:"refbarcode" gorm:"column:refbarcode"`
	UnitCode                 string  `json:"unitcode" gorm:"column:unitcode"`
	Qty                      float64 `json:"qty" gorm:"column:qty"`
	Price                    float64 `json:"price" gorm:"column:price"`
	Discount                 string  `json:"discount" gorm:"column:discount"`
	DiscountAmount           float64 `json:"discountamount" gorm:"column:discountamount"`
	SumAmount                float64 `json:"sumamount" gorm:"column:sumamount"`
	StandValue               float64 `json:"standvalue" gorm:"column:standvalue"`
	DivideValue              float64 `json:"dividevalue" gorm:"column:dividevalue"`
	CalcFlag                 int8    `json:"calcflag" gorm:"column:calcflag"`
	LineNumber               int8    `json:"linenumber" gorm:"column:linenumber"`
	CostPerUnit              float64 `json:"costperunit" gorm:"column:costperunit"` // ทุนต่อหน่วย
	TotalCost                float64 `json:"totalcost" gorm:"column:totalcost"`     // ต้นทุนรวม
	WhCode                   string  `json:"whcode" gorm:"whcode"`
	LocationCode             string  `json:"locationcode" gorm:"locationcode"`
	SumAmountExcludeVat      float64 `json:"sumamountexcludevat" gorm:"column:sumamountexcludevat"`
	TotalValueVat            float64 `json:"totalvaluevat" gorm:"column:totalvaluevat"`
	ItemGuid                 string  `json:"itemguid" gorm:"column:itemguid"`
	VatType                  int8    `json:"vattype" gorm:"column:vattype"`
	TaxType                  int8    `json:"taxtype" gorm:"column:taxtype"`
	PriceExcludeVat          float64 `json:"priceexcludevat" gorm:"column:priceexcludevat"`
	ItemType                 int8    `json:"itemtype" gorm:"column:itemtype"`
	DocRef                   string  `json:"docref" gorm:"column:docref"`
	BalanceQty               float64 `json:"balanceqty" gorm:"column:balanceqty"`         // ยอดคงเหลือ
	BalanceAmount            float64 `json:"balanceamount" gorm:"column:balanceamount"`   // มูลค่าคงเหลือ
	BalanceAverage           float64 `json:"balanceaverage" gorm:"column:balanceaverage"` // ต้นทุนเฉลี่ยคงเหลือ
	// SumOfCost                float64 `json:"sumofcost" gorm:"column:sumofcost"`
	// AverageCost              float64 `json:"averagecost" gorm:"column:averagecost"`
}

func (StockTransactionDetail) TableName() string {
	return "stocktransactiondetail"
}

func (j *StockTransaction) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]StockTransactionDetail
	tx.Model(&StockTransactionDetail{}).Where(" holdingcode=? AND docno=?", j.HoldingCode, j.DocNo).Find(&details)

	// delete un use data
	for _, tmp := range *details {
		var foundUpdate bool = false
		for _, data := range *j.Details {
			if data.ID == tmp.ID {
				foundUpdate = true
			}
		}
		if foundUpdate == false {
			// mark delete
			tx.Delete(&StockTransactionDetail{}, tmp.ID)
		}
	}

	return nil
}

func (jd *StockTransactionDetail) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *StockTransaction) CompareTo(other *StockTransaction) bool {

	diff := cmp.Diff(s, other,
		cmpopts.IgnoreFields(StockTransaction{}, "TotalCost"),
		cmpopts.IgnoreFields(StockTransactionDetail{}, "ID", "TotalCost", "CostPerUnit"),
	)

	if diff == "" {
		return true
	}

	return false
}
