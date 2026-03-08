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
	FindPurchaseReceiveDocByShopID(ctx context.Context, shopID string) ([]purchasePartialModels.PurchasepartialDoc, error)
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]purchasePartialModels.PurchasepartialDoc, mongopagination.PaginationData, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable msModels.Pageable) ([]purchasePartialModels.PurchasepartialDoc, mongopagination.PaginationData, error)
	FindPurchaseReceiveDocDeleteByShopID(ctx context.Context, shopID string) ([]purchasePartialModels.PurchasepartialDoc, error)
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

func (r PurchaseReceiveTransactionAdminRepositories) FindPurchaseReceiveDocByShopID(ctx context.Context, shopID string) ([]purchasePartialModels.PurchasepartialDoc, error) {

	docList := []purchasePartialModels.PurchasepartialDoc{}

	err := r.pst.Find(ctx, &purchasePartialModels.PurchasepartialDoc{},
		bson.M{
			"shopid":    shopID,
			"deletedat": bson.M{"$exists": false},
		},
		&docList)
	if err != nil {
		return nil, err
	}

	return docList, nil
}

func (r PurchaseReceiveTransactionAdminRepositories) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]purchasePartialModels.PurchasepartialDoc, mongopagination.PaginationData, error) {

	results, pagination, err := r.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func (r PurchaseReceiveTransactionAdminRepositories) FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable msModels.Pageable) ([]purchasePartialModels.PurchasepartialDoc, mongopagination.PaginationData, error) {

	results, pagination, err := r.SearchRepository.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func (r PurchaseReceiveTransactionAdminRepositories) FindPurchaseReceiveDocDeleteByShopID(ctx context.Context, shopID string) ([]purchasePartialModels.PurchasepartialDoc, error) {
	docList := []purchasePartialModels.PurchasepartialDoc{}

	err := r.pst.Find(ctx, &purchasePartialModels.PurchasepartialDoc{},
		bson.M{
			"shopid":    shopID,
			"deletedat": bson.M{"$exists": true},
		},
		&docList)
	if err != nil {
		return nil, err
	}

	return docList, nil
}
