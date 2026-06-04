package datatransfer

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/slipimage/models"
	slipImageRepository "smlcloudplatform/internal/slipimage/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SlipImageDataTransfer struct {
	transferConnection IDataTransferConnection
}

type ISlipImageDataTransferRepository interface {
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable msModels.Pageable) ([]models.SlipImageDoc, mongopagination.PaginationData, error)
}

type SlipImageDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.SlipImageDoc]
}

func NewSlipImageDataTransferRepository(mongodbPersister microservice.IPersisterMongo) ISlipImageDataTransferRepository {
	repo := &SlipImageDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.SlipImageDoc](mongodbPersister)
	return repo
}

func (repo SlipImageDataTransferRepository) FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable msModels.Pageable) ([]models.SlipImageDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, holdingCode, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewSlipImageDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &SlipImageDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *SlipImageDataTransfer) StartTransfer(ctx context.Context, holdingCode string, targetHoldingCode string) error {

	sourceRepository := NewSlipImageDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := slipImageRepository.NewSlipImageMongoRepository(pdt.transferConnection.GetTargetConnection())

	pageRequest := msModels.Pageable{
		Limit: 100,
		Page:  1,
	}

	for {

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
