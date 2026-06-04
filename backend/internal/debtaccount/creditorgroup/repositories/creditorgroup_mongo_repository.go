package repositories

import (
	"context"
	"smlcloudplatform/internal/debtaccount/creditorgroup/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type ICreditorGroupRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.CreditorGroupDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.CreditorGroupDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.CreditorGroupDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.CreditorGroupInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.CreditorGroupDoc, error)
	FindByGuids(ctx context.Context, holdingCode string, guids []string) ([]models.CreditorGroupDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.CreditorGroupItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.CreditorGroupDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.CreditorGroupInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.CreditorGroupInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CreditorGroupDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.CreditorGroupActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.CreditorGroupDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.CreditorGroupActivity, error)
}

type CreditorGroupRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.CreditorGroupDoc]
	repositories.SearchRepository[models.CreditorGroupInfo]
	repositories.GuidRepository[models.CreditorGroupItemGuid]
	repositories.ActivityRepository[models.CreditorGroupActivity, models.CreditorGroupDeleteActivity]
}

func NewCreditorGroupRepository(pst microservice.IPersisterMongo) *CreditorGroupRepository {

	insRepo := &CreditorGroupRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.CreditorGroupDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.CreditorGroupInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.CreditorGroupItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.CreditorGroupActivity, models.CreditorGroupDeleteActivity](pst)

	return insRepo
}
