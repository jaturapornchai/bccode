package datatransfer

import (
	"context"
	"smlcloudplatform/internal/product/productcategory/models"
	productcategoryrepository "smlcloudplatform/internal/product/productcategory/repositories"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ProductCategoryDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IProductCategoryDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.ProductCategoryDoc, mongopagination.PaginationData, error)
}

type ProductCategoryDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.ProductCategoryDoc]
}

func NewProductCategoryDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IProductCategoryDataTransferRepository {

	repo := &ProductCategoryDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.ProductCategoryDoc](mongodbPersister)
	return repo
}

func (repo ProductCategoryDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.ProductCategoryDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewProductCategoryDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &ProductCategoryDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pbd *ProductCategoryDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceProductCategoryRepository := NewProductCategoryDataTransferRepository(pbd.transferConnection.GetSourceConnection())
	targetProductCategoryRepository := productcategoryrepository.NewProductCategoryRepository(pbd.transferConnection.GetTargetConnection())

	pageRequest := msModels.Pageable{
		Limit: 100,
		Page:  1,
	}

	for {
		docs, pages, err := sourceProductCategoryRepository.FindPage(ctx, shopID, nil, pageRequest)
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

			err = targetProductCategoryRepository.CreateInBatch(ctx, docs)
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
