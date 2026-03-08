package datatransfer

import (
	"context"
	"smlcloudplatform/internal/pos/setting/models"
	posSettingRepository "smlcloudplatform/internal/pos/setting/repositories"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type PosSettingDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IPosSettingDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.SettingDoc, mongopagination.PaginationData, error)
}

type PosSettingDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.SettingDoc]
}

func NewPosSettingDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IPosSettingDataTransferRepository {
	repo := &PosSettingDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.SettingDoc](mongodbPersister)
	return repo
}

func (repo PosSettingDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.SettingDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewPosSettingDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &PosSettingDataTransfer{
		transferConnection: transferConnection,
	}
}

func (pdt *PosSettingDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewPosSettingDataTransferRepository(pdt.transferConnection.GetSourceConnection())
	targetRepository := posSettingRepository.NewSettingRepository(pdt.transferConnection.GetTargetConnection())

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
