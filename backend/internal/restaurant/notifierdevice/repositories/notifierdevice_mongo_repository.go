package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/restaurant/notifierdevice/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
)

type INotifierDeviceRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.NotifierDeviceDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.NotifierDeviceDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.NotifierDeviceDoc) error
	DeleteByGuidfixed(sctx context.Context, hopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.NotifierDeviceInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.NotifierDeviceDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.NotifierDeviceDoc, error)

	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.NotifierDeviceDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.NotifierDeviceInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.NotifierDeviceInfo, int, error)
}

type NotifierDeviceRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.NotifierDeviceDoc]
	repositories.SearchRepository[models.NotifierDeviceInfo]
}

func NewNotifierDeviceRepository(pst microservice.IPersisterMongo) *NotifierDeviceRepository {

	insRepo := &NotifierDeviceRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.NotifierDeviceDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.NotifierDeviceInfo](pst)

	return insRepo
}
