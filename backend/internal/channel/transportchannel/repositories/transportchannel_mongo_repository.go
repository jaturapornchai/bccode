package repositories

import (
	"context"
	"smlcloudplatform/internal/channel/transportchannel/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type ITransportChannelRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.TransportChannelDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.TransportChannelDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.TransportChannelDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.TransportChannelInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.TransportChannelDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.TransportChannelItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.TransportChannelDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.TransportChannelInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.TransportChannelInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.TransportChannelDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.TransportChannelActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.TransportChannelDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.TransportChannelActivity, error)
}

type TransportChannelRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.TransportChannelDoc]
	repositories.SearchRepository[models.TransportChannelInfo]
	repositories.GuidRepository[models.TransportChannelItemGuid]
	repositories.ActivityRepository[models.TransportChannelActivity, models.TransportChannelDeleteActivity]
}

func NewTransportChannelRepository(pst microservice.IPersisterMongo) *TransportChannelRepository {

	insRepo := &TransportChannelRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.TransportChannelDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.TransportChannelInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.TransportChannelItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.TransportChannelActivity, models.TransportChannelDeleteActivity](pst)

	return insRepo
}
