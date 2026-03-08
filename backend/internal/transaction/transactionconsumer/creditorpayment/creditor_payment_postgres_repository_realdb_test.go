package creditorpayment_test

import (
	"smlcloudplatform/internal/config"
	pkgModels "smlcloudplatform/internal/models"
	models "smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/creditorpayment"
	"smlcloudplatform/pkg/microservice"
	"testing"

	"github.com/stretchr/testify/assert"
)

var repo creditorpayment.ICreditorPaymentTransactionPGRepository

func init() {
	config := config.NewConfig()
	pst := microservice.NewPersister(config.PersisterConfig())
	repo = creditorpayment.NewCreditorPaymentTransactionPGRepository(pst)
}

func TestMigrationDB(t *testing.T) {

	err := repo.MigrationDatabase()
	assert.Nil(t, err)
}

func TestInsertData(t *testing.T) {
	giveDoc := wantDataCreditPayment()
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
	giveDoc := wantDataCreditPayment()
	err := repo.Delete(giveDoc.ShopID, giveDoc.DocNo, models.CreditorPaymentTransactionPG{
		ShopIdentity: pkgModels.ShopIdentity{
			ShopID: giveDoc.ShopID,
		},
		DocNo: giveDoc.DocNo,
	})
	assert.Nil(t, err)
}
