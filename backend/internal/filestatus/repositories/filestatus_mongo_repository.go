package repositories

import (
	"context"
	"smlcloudplatform/internal/filestatus/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IFileStatusRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.FileStatusDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.FileStatusDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.FileStatusDoc) error
	DeleteByGuidfixed(sctx context.Context, hopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.FileStatusInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.FileStatusDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.FileStatusDoc, error)

	FindOne(ctx context.Context, shopID string, filters interface{}) (models.FileStatusDoc, error)
	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.FileStatusItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.FileStatusDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.FileStatusInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.FileStatusInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.FileStatusDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(sctx context.Context, hopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.FileStatusActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.FileStatusDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.FileStatusActivity, error)
}

type FileStatusRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.FileStatusDoc]
	repositories.SearchRepository[models.FileStatusInfo]
	repositories.GuidRepository[models.FileStatusItemGuid]
	repositories.ActivityRepository[models.FileStatusActivity, models.FileStatusDeleteActivity]
}

func NewFileStatusRepository(pst microservice.IPersisterMongo) *FileStatusRepository {

	insRepo := &FileStatusRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.FileStatusDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.FileStatusInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.FileStatusItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.FileStatusActivity, models.FileStatusDeleteActivity](pst)

	return insRepo
}
