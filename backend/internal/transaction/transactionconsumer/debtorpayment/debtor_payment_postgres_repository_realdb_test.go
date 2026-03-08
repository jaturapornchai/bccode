package debtorpayment_test

import (
	"smlcloudplatform/internal/config"
	pkgModels "smlcloudplatform/internal/models"
	models "smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/debtorpayment"
	"smlcloudplatform/pkg/microservice"
	"testing"

	"github.com/stretchr/testify/assert"
)

var repo debtorpayment.IDebtorPaymentTransactionPGRepository
var pst microservice.IPersister

func init() {
	config := config.NewConfig()
	pst = microservice.NewPersister(config.PersisterConfig())
	repo = debtorpayment.NewDebtorPaymentTransactionPGRepository(pst)
}

func TestMigrationDB(t *testing.T) {

	pst.AutoMigrate(
		models.DebtorPaymentTransactionPG{},
		models.DebtorPaymentTransactionDetailPG{},
	)
}

func TestInsertData(t *testing.T) {

	giveDoc := wantDataDebtorPayment()
	err := repo.Create(*giveDoc)
	assert.Nil(t, err)

	gotDoc, err := repo.Get(giveDoc.ShopID, giveDoc.DocNo)
	assert.Nil(t, err)
	assert.Equal(t, giveDoc.DocNo, gotDoc.DocNo)

	giveDoc.TotalAmount = 99999

	err = repo.Update(giveDoc.ShopID, giveDoc.DocNo, *giveDoc)
	assert.Nil(t, err)

}

func TestDeleteDoc(t *testing.T) {
	giveDoc := wantDataDebtorPayment()
	err := repo.Delete(giveDoc.ShopID, giveDoc.DocNo, models.DebtorPaymentTransactionPG{
		ShopIdentity: pkgModels.ShopIdentity{
			ShopID: giveDoc.ShopID,
		},
		DocNo: giveDoc.DocNo,
	})
	assert.Nil(t, err)
}
