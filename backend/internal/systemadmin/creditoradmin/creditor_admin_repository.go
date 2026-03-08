package creditoradmin

import (
	"context"
	creditorModels "smlcloudplatform/internal/debtaccount/creditor/models"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type ICreditorAdminMongoRepository interface {
	FindCreditorByShopId(ctx context.Context, shopID string) ([]creditorModels.CreditorDoc, error)
}

type CreditorAdminMongoRepository struct {
	pst microservice.IPersisterMongo
}

func NewCreditorAdminMongoRepository(pst microservice.IPersisterMongo) ICreditorAdminMongoRepository {
	return &CreditorAdminMongoRepository{
		pst: pst,
	}
}

func (r CreditorAdminMongoRepository) FindCreditorByShopId(ctx context.Context, shopID string) ([]creditorModels.CreditorDoc, error) {

	docList := []creditorModels.CreditorDoc{}
	err := r.pst.Find(ctx, &creditorModels.CreditorDoc{}, bson.M{"shopid": shopID}, &docList)
	if err != nil {
		return nil, err
	}

	return docList, nil
}
