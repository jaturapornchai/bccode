package accountadmin

import (
	"context"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type IAccountAdminRepository interface {
	ListAllUser(context.Context) ([]Account, error)
}

type AccountAdminRepository struct {
	pst microservice.IPersisterMongo
}

func NewAccountAdminRepository(pst microservice.IPersisterMongo) IAccountAdminRepository {
	return &AccountAdminRepository{
		pst: pst,
	}
}

func (r *AccountAdminRepository) ListAllUser(ctx context.Context) ([]Account, error) {

	userList := []Account{}

	err := r.pst.Find(context.TODO(), Account{}, bson.M{}, &userList)

	if err != nil {
		return []Account{}, err
	}

	return userList, nil
}
