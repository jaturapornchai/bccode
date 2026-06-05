package models

import (
	pkgModels "smlcloudplatform/internal/models"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// ตั้งหนี้จากการรับสินค้า
type APPurchaseReceivePG struct {
	TransactionPG `bson:"inline"`
	CreditorCode  string                       `json:"creditorcode" gorm:"column:creditorcode"`
	CreditorNames pkgModels.JSONB              `json:"creditornames" gorm:"column:creditornames;type:jsonb"`
	Items         *[]APPurchaseReceiveDetailPG `json:"items" gorm:"items;foreignKey:holdingcode,docno"`
}

// รายละเอียด ตั้งหนี้จากการรับสินค้า
type APPurchaseReceiveDetailPG struct {
	TransactionDetailPG `gorm:"embedded;"`
}

func (APPurchaseReceivePG) TableName() string {
	return "ap_purchasereceive_transaction"
}

func (APPurchaseReceiveDetailPG) TableName() string {
	return "ap_purchasereceive_transaction_detail"
}

func (m *APPurchaseReceivePG) BeforeUpdate(tx *gorm.DB) (err error) {

	// find old data
	var details *[]APPurchaseReceiveDetailPG
	tx.Model(&APPurchaseReceiveDetailPG{}).Where(" holdingcode=? AND docno=?", m.HoldingCode, m.DocNo).Find(&details)

	// delete un use data
	for _, tmp := range *details {
		var foundUpdate bool = false
		for _, data := range *m.Items {
			if data.ID == tmp.ID {
				foundUpdate = true
			}
		}
		if !foundUpdate {
			// mark delete
			tx.Delete(&APPurchaseReceiveDetailPG{}, tmp.ID)
		}
	}

	return nil
}

func (jd *APPurchaseReceivePG) BeforeCreate(tx *gorm.DB) error {

	tx.Statement.AddClause(clause.OnConflict{
		UpdateAll: true,
	})
	return nil
}

func (s *APPurchaseReceivePG) CompareTo(other *APPurchaseReceivePG) bool {

	diff := cmp.Diff(s, other,
		// cmpopts.IgnoreFields(PurchaseTransactionPG{}, "TotalCost"),
		cmpopts.IgnoreFields(APPurchaseReceiveDetailPG{}, "ID"),
	)

	return diff == ""
}
