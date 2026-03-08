package datatransfer

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/shop/employee/models"
	shopEmployeeRepositories "smlcloudplatform/internal/shop/employee/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ShopEmployeeDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IShopEmployeeDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.EmployeeDoc, mongopagination.PaginationData, error)
}

type ShopEmployeeDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.EmployeeDoc]
}

func NewShopEmployeeDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IShopEmployeeDataTransferRepository {
	repo := &ShopEmployeeDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.EmployeeDoc](mongodbPersister)
	return repo
}

func (repo ShopEmployeeDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.EmployeeDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewShopEmployeeDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {

	return &ShopEmployeeDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *ShopEmployeeDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewShopEmployeeDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := shopEmployeeRepositories.NewEmployeeRepository(pdt.transferConnection.GetTargetConnection())

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
