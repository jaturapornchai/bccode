package repositories

import (
	"context"
	"smlcloudplatform/internal/channel/salechannel/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type ISaleChannelRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.SaleChannelDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.SaleChannelDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.SaleChannelDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.SaleChannelInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.SaleChannelDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.SaleChannelDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.SaleChannelItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.SaleChannelDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.SaleChannelInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.SaleChannelInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SaleChannelDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SaleChannelActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SaleChannelDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SaleChannelActivity, error)
}

type SaleChannelRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.SaleChannelDoc]
	repositories.SearchRepository[models.SaleChannelInfo]
	repositories.GuidRepository[models.SaleChannelItemGuid]
	repositories.ActivityRepository[models.SaleChannelActivity, models.SaleChannelDeleteActivity]
}

func NewSaleChannelRepository(pst microservice.IPersisterMongo) *SaleChannelRepository {

	insRepo := &SaleChannelRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.SaleChannelDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.SaleChannelInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.SaleChannelItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.SaleChannelActivity, models.SaleChannelDeleteActivity](pst)

	return insRepo
}
