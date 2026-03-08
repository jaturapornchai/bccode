package datatransfer

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/purchasereturn/models"
	purchaseReturnRepository "smlcloudplatform/internal/transaction/purchasereturn/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PurchaseReturnDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IPurchaseReturnDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.PurchaseReturnDoc, mongopagination.PaginationData, error)
}

type PurchaseReturnDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.PurchaseReturnDoc]
}

func NewPurchaseReturnDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IPurchaseReturnDataTransferRepository {
	repo := &PurchaseReturnDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.PurchaseReturnDoc](mongodbPersister)
	return repo
}

func (repo PurchaseReturnDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.PurchaseReturnDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewPurchaseReturnDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &PurchaseReturnDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *PurchaseReturnDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewPurchaseReturnDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := purchaseReturnRepository.NewPurchaseReturnRepository(pdt.transferConnection.GetTargetConnection())

	pageRequest := msModels.Pageable{
		Limit: 100,
		Page:  1,
	}

	for {

		docs, pages, err := sourceRepository.FindPage(ctx, shopID, nil, pageRequest)
		if err != nil {
			return err
		}

		if len(docs) > 0 {

			if targetShopID != "" {
				for i := range docs {
					docs[i].ShopID = targetShopID
					docs[i].ID = primitive.NewObjectID()
				}
			}

			err = targetRepository.CreateInBatch(ctx, docs)
			if err != nil {
				return err
			}
		}

		if pages.TotalPage > int64(pageRequest.Page) {
			pageRequest.Page++
		} else {
			break
		}
	}

	return nil
}
