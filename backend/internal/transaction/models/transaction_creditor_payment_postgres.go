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

type CreditorPaymentTransactionPG struct {
	GuidFixed                string `json:"guidfixed" gorm:"column:guidfixed"`
	models.ShopIdentity      `bson:"inline"`
	models.PartitionIdentity `gorm:"embedded;"`
	DocNo                    string                                `json:"docno" gorm:"column:docno;primaryKey"`
	DocDate                  time.Time                             `json:"docdate" gorm:"column:docdate"`
	BranchCode               string                                `json:"branchcode" gorm:"column:branchcode"`
	BranchNames              models.JSONB                          `json:"branchnames" gorm:"column:branchnames;type:jsonb"`
	CreditorCode             string                                `json:"creditorcode" gorm:"column:creditorcode"`
	CreditorNames            pkgModels.JSONB                       `json:"creditornames" gorm:"column:creditornames;type:jsonb"`
	TotalAmount              float64                               `json:"totalamount" gorm:"column:totalamount"`
	TotalPayCash             float64                               `json:"totalpaycash" gorm:"column:totalpaycash"`
	TotalPayTransfer         float64                               `json:"totalpaytransfer" gorm:"column:totalpaytransfer"`
	TotalPayCredit           float64                               `json:"totalpaycredit" gorm:"column:totalpaycredit"`
	Details                  *[]CreditorPaymentTransactionDetailPG `json:"details" gorm:"details;foreignKey:shopid,docno"`
}

func (CreditorPaymentTransactionPG) TableName() string {
	return "creditor_payment_transaction"
}

type CreditorPaymentTransactionDetailPG struct {
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

func (CreditorPaymentTransactionDetailPG) TableName() string {
	return "creditor_payment_transaction_detail"
}

func (j *CreditorPaymentTransactionPG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]CreditorPaymentTransactionDetailPG
	tx.Model(&CreditorPaymentTransactionDetailPG{}).Where(" shopid=? AND docno=?", j.ShopID, j.DocNo).Find(&details)

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
			tx.Delete(&CreditorPaymentTransactionDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (jd *CreditorPaymentTransactionDetailPG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (j *CreditorPaymentTransactionPG) BeforeDelete(tx *gorm.DB) (err error) {

	// find old data
	var details *[]CreditorPaymentTransactionDetailPG
	tx.Model(&CreditorPaymentTransactionDetailPG{}).Where(" shopid=? AND docno=?", j.ShopID, j.DocNo).Find(&details)

	// delete unuse data
	for _, tmp := range *details {
		tx.Delete(&CreditorPaymentTransactionDetailPG{}, tmp.ID)
	}

	return nil
}

func (s *CreditorPaymentTransactionPG) CompareTo(other *CreditorPaymentTransactionPG) bool {

	diff := cmp.Diff(s, other,
		//cmpopts.IgnoreFields(StockReturnProductTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(CreditorPaymentTransactionDetailPG{}, "ID"),
	)

	if diff == "" {
		return true
	}

	return false
}
