package datatransfer

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	models "smlcloudplatform/internal/organization/branch/models"
	organizationBranchRepository "smlcloudplatform/internal/organization/branch/repositories"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrganizationBranchDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IOrganizationBranchDataTransferRepository interface {
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.BranchDoc, mongopagination.PaginationData, error)
}

type OrganizationBranchDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.BranchDoc]
}

func NewOrganizationBranchDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IOrganizationBranchDataTransferRepository {
	repo := &OrganizationBranchDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.BranchDoc](mongodbPersister)
	return repo
}

func (repo OrganizationBranchDataTransferRepository) FindPage(ctx context.Context, shopID string, searchInFields []string, pageable msModels.Pageable) ([]models.BranchDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, shopID, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewOrganizationBranchDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &OrganizationBranchDataTransfer{
		transferConnection: transferConnection,
	}
}

func (dt *OrganizationBranchDataTransfer) StartTransfer(ctx context.Context, shopID string, targetShopID string) error {

	sourceRepository := NewOrganizationBranchDataTransferRepository(dt.transferConnection.GetSourceConnection())
	targetRepository := organizationBranchRepository.NewBranchRepository(dt.transferConnection.GetTargetConnection())

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
