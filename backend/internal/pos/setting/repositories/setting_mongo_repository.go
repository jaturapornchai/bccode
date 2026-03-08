package repositories

import (
	"context"
	"smlcloudplatform/internal/pos/setting/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type ISettingRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.SettingDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.SettingDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.SettingDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.SettingInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.SettingDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.SettingItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.SettingDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.SettingInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.SettingInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SettingDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SettingActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SettingDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SettingActivity, error)
}

type SettingRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.SettingDoc]
	repositories.SearchRepository[models.SettingInfo]
	repositories.GuidRepository[models.SettingItemGuid]
	repositories.ActivityRepository[models.SettingActivity, models.SettingDeleteActivity]
}

func NewSettingRepository(pst microservice.IPersisterMongo) *SettingRepository {

	insRepo := &SettingRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.SettingDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.SettingInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.SettingItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.SettingActivity, models.SettingDeleteActivity](pst)

	return insRepo
}
