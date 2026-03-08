package datatransfer

import (
	"context"
	"smlcloudplatform/internal/product/bom/models"
	productbarcodeBOMRepository "smlcloudplatform/internal/product/bom/repositories"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductbarcodeBOMDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IProductbarcodeBOMDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.ProductBarcodeBOMViewDoc, mongopagination.PaginationData, error)
}

type ProductbarcodeBOMDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.ProductBarcodeBOMViewDoc]
}

func NewProductbarcodeBOMDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IProductbarcodeBOMDataTransferRepository {
	repo := &ProductbarcodeBOMDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.ProductBarcodeBOMViewDoc](mongodbPersister)
	return repo
}

func (repo ProductbarcodeBOMDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.ProductBarcodeBOMViewDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewProductbarcodeBOMDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &ProductbarcodeBOMDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *ProductbarcodeBOMDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewProductbarcodeBOMDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := productbarcodeBOMRepository.NewBomRepository(pdt.transferConnection.GetTargetConnection())

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
