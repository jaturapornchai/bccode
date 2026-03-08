package repositories

import (
	"context"
	"smlcloudplatform/internal/pos/media/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IMediaRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.MediaDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.MediaDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.MediaDoc) error
	DeleteByGuidfixed(sctx context.Context, hopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.MediaInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.MediaDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.MediaDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.MediaItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.MediaDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.MediaInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.MediaInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.MediaDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(sctx context.Context, hopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.MediaActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.MediaDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.MediaActivity, error)
}

type MediaRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.MediaDoc]
	repositories.SearchRepository[models.MediaInfo]
	repositories.GuidRepository[models.MediaItemGuid]
	repositories.ActivityRepository[models.MediaActivity, models.MediaDeleteActivity]
}

func NewMediaRepository(pst microservice.IPersisterMongo) *MediaRepository {

	insRepo := &MediaRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.MediaDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.MediaInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.MediaItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.MediaActivity, models.MediaDeleteActivity](pst)

	return insRepo
}
