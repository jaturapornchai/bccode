package datatransfer

import (
	"context"
	"smlcloudplatform/internal/product/unit/models"
	productUnitRepository "smlcloudplatform/internal/product/unit/repositories"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductUnitDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IProductUnitDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.UnitDoc, mongopagination.PaginationData, error)
}

type ProductUnitDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.UnitDoc]
}

func NewProductUnitDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IProductUnitDataTransferRepository {
	repo := &ProductUnitDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.UnitDoc](mongodbPersister)
	return repo
}

func (repo ProductUnitDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.UnitDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewProductUnitDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &ProductUnitDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *ProductUnitDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewProductUnitDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := productUnitRepository.NewUnitRepository(pdt.transferConnection.GetTargetConnection())

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
