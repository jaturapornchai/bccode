package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/restaurant/notifier/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
)

type INotifierRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.NotifierDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.NotifierDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.NotifierDoc) error
	DeleteByGuidfixed(sctx context.Context, hopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.NotifierInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.NotifierDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.NotifierDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.NotifierItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.NotifierDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.NotifierInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.NotifierInfo, int, error)
}

type NotifierRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.NotifierDoc]
	repositories.SearchRepository[models.NotifierInfo]
	repositories.GuidRepository[models.NotifierItemGuid]
}

func NewNotifierRepository(pst microservice.IPersisterMongo) *NotifierRepository {

	insRepo := &NotifierRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.NotifierDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.NotifierInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.NotifierItemGuid](pst)

	return insRepo
}
