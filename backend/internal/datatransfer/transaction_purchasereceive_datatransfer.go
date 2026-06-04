package datatransfer

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/transaction/purchasepartial/models"
	purchasePartialRepository "smlcloudplatform/internal/transaction/purchasepartial/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PurchaseReceiveDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IPurchaseReceiveDataTransferRepository interface {
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable msModels.Pageable) ([]models.PurchasepartialDoc, mongopagination.PaginationData, error)
}

type PurchaseReceiveDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.PurchasepartialDoc]
}

func NewPurchaseReceiveDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IPurchaseReceiveDataTransferRepository {
	repo := &PurchaseReceiveDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.PurchasepartialDoc](mongodbPersister)
	return repo
}

func (repo PurchaseReceiveDataTransferRepository) FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable msModels.Pageable) ([]models.PurchasepartialDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, holdingCode, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewPurchaseReceiveDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &PurchaseReceiveDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *PurchaseReceiveDataTransfer) StartTransfer(ctx context.Context, holdingCode string, targetHoldingCode string) error {

	sourceRepository := NewPurchaseReceiveDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := purchasePartialRepository.NewPurchasepartialRepository(pdt.transferConnection.GetTargetConnection())

	pageRequest := msModels.Pageable{
		Limit: 100,
		Page:  1,
	}

	for {

		fmt.Println("Transfering page: ", pageRequest.Page)
		docs, pages, err := sourceRepository.FindPage(ctx, holdingCode, nil, pageRequest)
		if err != nil {
			return err
		}

		if len(docs) > 0 {

			if targetHoldingCode != "" {
				for i := range docs {
					docs[i].HoldingCode = targetHoldingCode
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
