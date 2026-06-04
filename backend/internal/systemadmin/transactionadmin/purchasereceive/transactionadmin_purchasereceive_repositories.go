package purchasereceive

import (
	"context"
	"smlcloudplatform/internal/repositories"
	purchasePartialModels "smlcloudplatform/internal/transaction/purchasepartial/models"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IPurchaseReceiveTransactionAdminRepositories interface {
	FindPurchaseReceiveDocByHoldingCode(ctx context.Context, holdingCode string) ([]purchasePartialModels.PurchasepartialDoc, error)
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable msModels.Pageable) ([]purchasePartialModels.PurchasepartialDoc, mongopagination.PaginationData, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable msModels.Pageable) ([]purchasePartialModels.PurchasepartialDoc, mongopagination.PaginationData, error)
	FindPurchaseReceiveDocDeleteByHoldingCode(ctx context.Context, holdingCode string) ([]purchasePartialModels.PurchasepartialDoc, error)
}

type PurchaseReceiveTransactionAdminRepositories struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[purchasePartialModels.PurchasepartialDoc]
}

func NewPurchaseReceiveTransactionAdminRepositories(pst microservice.IPersisterMongo) IPurchaseReceiveTransactionAdminRepositories {
	return &PurchaseReceiveTransactionAdminRepositories{
		pst:              pst,
		SearchRepository: repositories.NewSearchRepository[purchasePartialModels.PurchasepartialDoc](pst),
	}
}

func (r PurchaseReceiveTransactionAdminRepositories) FindPurchaseReceiveDocByHoldingCode(ctx context.Context, holdingCode string) ([]purchasePartialModels.PurchasepartialDoc, error) {

	docList := []purchasePartialModels.PurchasepartialDoc{}

	err := r.pst.Find(ctx, &purchasePartialModels.PurchasepartialDoc{},
		bson.M{
			"holding_code": holdingCode,
			"deleted_at":   bson.M{"$exists": false},
		},
		&docList)
	if err != nil {
		return nil, err
	}

	return docList, nil
}

func (r PurchaseReceiveTransactionAdminRepositories) FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable msModels.Pageable) ([]purchasePartialModels.PurchasepartialDoc, mongopagination.PaginationData, error) {

	results, pagination, err := r.SearchRepository.FindPage(ctx, holdingCode, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func (r PurchaseReceiveTransactionAdminRepositories) FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable msModels.Pageable) ([]purchasePartialModels.PurchasepartialDoc, mongopagination.PaginationData, error) {

	results, pagination, err := r.SearchRepository.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func (r PurchaseReceiveTransactionAdminRepositories) FindPurchaseReceiveDocDeleteByHoldingCode(ctx context.Context, holdingCode string) ([]purchasePartialModels.PurchasepartialDoc, error) {
	docList := []purchasePartialModels.PurchasepartialDoc{}

	err := r.pst.Find(ctx, &purchasePartialModels.PurchasepartialDoc{},
		bson.M{
			"holding_code": holdingCode,
			"deleted_at":   bson.M{"$exists": true},
		},
		&docList)
	if err != nil {
		return nil, err
	}

	return docList, nil
}
