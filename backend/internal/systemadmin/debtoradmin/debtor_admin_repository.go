package debtoradmin

import (
	"context"
	debtorModels "smlcloudplatform/internal/debtaccount/debtor/models"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type IDebtorAdminMongoRepository interface {
	FindDebtorByHoldingCode(ctx context.Context, holdingCode string) ([]debtorModels.DebtorDoc, error)
}

type DebtorAdminMongoRepository struct {
	pst microservice.IPersisterMongo
}

func NewDebtorAdminMongoRepository(pst microservice.IPersisterMongo) IDebtorAdminMongoRepository {
	return &DebtorAdminMongoRepository{
		pst: pst,
	}
}

func (r DebtorAdminMongoRepository) FindDebtorByHoldingCode(ctx context.Context, holdingCode string) ([]debtorModels.DebtorDoc, error) {

	docList := []debtorModels.DebtorDoc{}
	err := r.pst.Find(ctx, &debtorModels.DebtorDoc{}, bson.M{"holding_code": holdingCode}, &docList)
	if err != nil {
		return nil, err
	}

	return docList, nil
}
