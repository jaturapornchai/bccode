package datatransfer

import (
	"context"
	"smlcloudplatform/internal/product/productgroup/models"
	productGroupRepository "smlcloudplatform/internal/product/productgroup/repositories"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductGroupDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IProductGroupDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.ProductGroupDoc, mongopagination.PaginationData, error)
}

type ProductGroupDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.ProductGroupDoc]
}

func NewProductGroupDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IProductGroupDataTransferRepository {
	repo := &ProductGroupDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.ProductGroupDoc](mongodbPersister)
	return repo
}

func (repo ProductGroupDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.ProductGroupDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewProductGroupDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &ProductGroupDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *ProductGroupDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewProductGroupDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := productGroupRepository.NewProductGroupRepository(pdt.transferConnection.GetTargetConnection())

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
