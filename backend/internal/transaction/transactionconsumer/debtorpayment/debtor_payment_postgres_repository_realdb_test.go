package debtorpayment_test

import (
	"os"
	"smlcloudplatform/internal/config"
	pkgModels "smlcloudplatform/internal/models"
	models "smlcloudplatform/internal/transaction/models"
	"smlcloudplatform/internal/transaction/transactionconsumer/debtorpayment"
	"smlcloudplatform/pkg/microservice"
	"testing"

	"github.com/stretchr/testify/assert"
)

func newRealDBRepository(t *testing.T) (debtorpayment.IDebtorPaymentTransactionPGRepository, microservice.IPersister) {
	t.Helper()
	if os.Getenv("BC_REAL_DB_TESTS") != "1" {
		t.Skip("set BC_REAL_DB_TESTS=1 to run PostgreSQL integration tests")
	}

	config := config.NewConfig()
	pst := microservice.NewPersister(config.PersisterConfig())
	return debtorpayment.NewDebtorPaymentTransactionPGRepository(pst), pst
}

func TestMigrationDB(t *testing.T) {
	_, pst := newRealDBRepository(t)

	pst.AutoMigrate(
		models.DebtorPaymentTransactionPG{},
		models.DebtorPaymentTransactionDetailPG{},
	)
}

func TestInsertData(t *testing.T) {
	repo, _ := newRealDBRepository(t)

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
	repo, _ := newRealDBRepository(t)

	giveDoc := wantDataDebtorPayment()
	err := repo.Delete(giveDoc.ShopID, giveDoc.DocNo, models.DebtorPaymentTransactionPG{
		ShopIdentity: pkgModels.ShopIdentity{
			ShopID: giveDoc.ShopID,
		},
		DocNo: giveDoc.DocNo,
	})
	assert.Nil(t, err)
}
