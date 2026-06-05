package creditoradmin

import (
	"context"
	creditorModels "smlcloudplatform/internal/debtaccount/creditor/models"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type ICreditorAdminMongoRepository interface {
	FindCreditorByHoldingCode(ctx context.Context, holdingCode string) ([]creditorModels.CreditorDoc, error)
}

type CreditorAdminMongoRepository struct {
	pst microservice.IPersisterMongo
}

func NewCreditorAdminMongoRepository(pst microservice.IPersisterMongo) ICreditorAdminMongoRepository {
	return &CreditorAdminMongoRepository{
		pst: pst,
	}
}

func (r CreditorAdminMongoRepository) FindCreditorByHoldingCode(ctx context.Context, holdingCode string) ([]creditorModels.CreditorDoc, error) {

	docList := []creditorModels.CreditorDoc{}
	err := r.pst.Find(ctx, &creditorModels.CreditorDoc{}, bson.M{"holdingcode": holdingCode}, &docList)
	if err != nil {
		return nil, err
	}

	return docList, nil
}
