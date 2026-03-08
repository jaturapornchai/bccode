package models

import (
	"smlcloudplatform/internal/models"
	pkgModels "smlcloudplatform/internal/models"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DebtorPaymentTransactionPG struct {
	GuidFixed                string `json:"guidfixed" gorm:"column:guidfixed"`
	models.ShopIdentity      `bson:"inline"`
	models.PartitionIdentity `gorm:"embedded;"`
	DocNo                    string                              `json:"docno" gorm:"column:docno;primaryKey"`
	DocDate                  time.Time                           `json:"docdate" gorm:"column:docdate"`
	BranchCode               string                              `json:"branchcode" gorm:"column:branchcode"`
	BranchNames              models.JSONB                        `json:"branchnames" gorm:"column:branchnames;type:jsonb"`
	DebtorCode               string                              `json:"creditorcode" gorm:"column:creditorcode"`
	DebtorNames              pkgModels.JSONB                     `json:"creditornames" gorm:"column:creditornames;type:jsonb"`
	TotalAmount              float64                             `json:"totalamount" gorm:"column:totalamount"`
	TotalPayCash             float64                             `json:"totalpaycash" gorm:"column:totalpaycash"`
	TotalPayTransfer         float64                             `json:"totalpaytransfer" gorm:"column:totalpaytransfer"`
	TotalPayCredit           float64                             `json:"totalpaycredit" gorm:"column:totalpaycredit"`
	Details                  *[]DebtorPaymentTransactionDetailPG `json:"details" gorm:"details;foreignKey:shopid,docno"`
}

func (DebtorPaymentTransactionPG) TableName() string {
	return "debtor_payment_transaction"
}

type DebtorPaymentTransactionDetailPG struct {
	ID                       uint   `gorm:"primarykey"`
	ShopID                   string `json:"shopid" gorm:"column:shopid"`
	models.PartitionIdentity `gorm:"embedded;"`
	DocNo                    string  `json:"docno" gorm:"column:docno"`
	LineNumber               int8    `json:"linenumber" gorm:"column:linenumber"`
	BillingNo                string  `json:"billingno" gorm:"column:billingno"`
	BillType                 int8    `json:"billtype" gorm:"column:billtype"`
	BillAmount               float64 `json:"billamount" gorm:"column:billamount"`
	BalanceAmount            float64 `json:"balanceamount" gorm:"column:balanceamount"`
	PayAmount                float64 `json:"payamount" gorm:"column:payamount"`
}

func (DebtorPaymentTransactionDetailPG) TableName() string {
	return "debtor_payment_transaction_detail"
}

func (j *DebtorPaymentTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]DebtorPaymentTransactionDetailPG
	tx.Model(&DebtorPaymentTransactionDetailPG{}).Where(" shopid=? AND docno=?", j.ShopID, j.DocNo).Find(&details)

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
			tx.Delete(&DebtorPaymentTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (jd *DebtorPaymentTransactionDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (j *DebtorPaymentTransactionPG) BeforeDelete(tx *gorm.DB) (err error) {

	// find old data
	var details *[]DebtorPaymentTransactionDetailPG
	tx.Model(&DebtorPaymentTransactionDetailPG{}).Where(" shopid=? AND docno=?", j.ShopID, j.DocNo).Find(&details)

	// delete unuse data
	for _, tmp := range *details {
		tx.Delete(&DebtorPaymentTransactionDetailPG{}, tmp.ID)
	}

	return nil
}

func (s *DebtorPaymentTransactionPG) CompareTo(other *DebtorPaymentTransactionPG) bool {

	diff := cmp.Diff(s, other, //cmpopts.IgnoreFields(StockReturnProductTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(DebtorPaymentTransactionDetailPG{}, "ID"),
	)

	if diff == "" {
		return true
	}

	return false
}
