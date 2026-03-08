package debtoradmin

import (
	"context"
	debtorModels "smlcloudplatform/internal/debtaccount/debtor/models"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type IDebtorAdminMongoRepository interface {
	FindDebtorByShopId(ctx context.Context, shopID string) ([]debtorModels.DebtorDoc, error)
}

type DebtorAdminMongoRepository struct {
	pst microservice.IPersisterMongo
}

func NewDebtorAdminMongoRepository(pst microservice.IPersisterMongo) IDebtorAdminMongoRepository {
	return &DebtorAdminMongoRepository{
		pst: pst,
	}
}

func (r DebtorAdminMongoRepository) FindDebtorByShopId(ctx context.Context, shopID string) ([]debtorModels.DebtorDoc, error) {

	docList := []debtorModels.DebtorDoc{}
	err := r.pst.Find(ctx, &debtorModels.DebtorDoc{}, bson.M{"shopid": shopID}, &docList)
	if err != nil {
		return nil, err
	}

	return docList, nil
}
