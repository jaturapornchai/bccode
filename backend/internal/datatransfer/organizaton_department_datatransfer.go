package datatransfer

import (
	"context"
	"smlcloudplatform/internal/organization/department/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	organizationDepartmentRepository "smlcloudplatform/internal/organization/department/repositories"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type OrganizationDepartmentDataTransfer struct {
	transferConnection IDataTransferConnection
}

type IOrganizationDepartmentDataTransferRepository interface {
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable msModels.Pageable) ([]models.DepartmentDoc, mongopagination.PaginationData, error)
}

type OrganizationDepartmentDataTransferRepository struct {
	pst microservice.IPersisterMongo
	repositories.SearchRepository[models.DepartmentDoc]
}

func NewOrganizationDepartmentDataTransferRepository(mongodbPersister microservice.IPersisterMongo) IOrganizationDepartmentDataTransferRepository {
	repo := &OrganizationDepartmentDataTransferRepository{
		pst: mongodbPersister,
	}

	repo.SearchRepository = repositories.NewSearchRepository[models.DepartmentDoc](mongodbPersister)
	return repo
}

func (repo OrganizationDepartmentDataTransferRepository) FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable msModels.Pageable) ([]models.DepartmentDoc, mongopagination.PaginationData, error) {

	results, pagination, err := repo.SearchRepository.FindPage(ctx, holdingCode, searchInFields, pageable)

	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return results, pagination, nil
}

func NewOrganizationDepartmentDataTransfer(transferConnection IDataTransferConnection) IDataTransfer {
	return &OrganizationDepartmentDataTransfer{
		transferConnection: transferConnection,
	}
}

func (dt *OrganizationDepartmentDataTransfer) StartTransfer(ctx context.Context, holdingCode string, targetHoldingCode string) error {

	sourceRepository := NewOrganizationDepartmentDataTransferRepository(dt.transferConnection.GetSourceConnection())
	targetRepository := organizationDepartmentRepository.NewDepartmentRepository(dt.transferConnection.GetTargetConnection())

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
