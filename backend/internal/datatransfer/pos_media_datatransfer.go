package datatransfer

import (
	"context"
	"smlcloudplatform/internal/pos/media/models"
	posMediaRepository "smlcloudplatform/internal/pos/media/repositories"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PosMediaDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IPosMediaDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.MediaDoc, mongopagination.PaginationData, error)
}

type PosMediaDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.MediaDoc]
}

func NewPosMediaDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IPosMediaDataTransferRepository {
	repo := &PosMediaDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.MediaDoc](mongodbPersister)
	return repo
}

func (repo PosMediaDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.MediaDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewPosMediaDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &PosMediaDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *PosMediaDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewPosMediaDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := posMediaRepository.NewMediaRepository(pdt.transferConnection.GetTargetConnection())

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
