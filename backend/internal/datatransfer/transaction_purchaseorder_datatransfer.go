package datatransfer

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/purchaseorder/models"
	purchaseOrderRepository "smlcloudplatform/internal/transaction/purchaseorder/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PurchaseOrderDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IPurchaseOrderDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.PurchaseOrderDoc, mongopagination.PaginationData, error)
}

type PurchaseOrderDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.PurchaseOrderDoc]
}

func NewPurchaseOrderDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IPurchaseOrderDataTransferRepository {
	repo := &PurchaseOrderDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.PurchaseOrderDoc](mongodbPersister)
	return repo
}

func (repo PurchaseOrderDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.PurchaseOrderDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewPurchaseOrderDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &PurchaseOrderDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *PurchaseOrderDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewPurchaseOrderDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := purchaseOrderRepository.NewPurchaseOrderRepository(pdt.transferConnection.GetTargetConnection())

	pageRequest := msModels.Pageable{
		Limit: 100,
		Page:  1,
	}

	for {

		fmt.Println("Transfering page: ", pageRequest.Page)
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
