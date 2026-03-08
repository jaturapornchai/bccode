package printer

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/restaurant/printer/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IPrinterRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(category models.PrinterDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.PrinterDoc) error
	Update(ctx context.Context, shopID string, guid string, category models.PrinterDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.PrinterInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.PrinterDoc, error)
	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.PrinterItemGuid, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, pageable micromodels.Pageable) ([]models.PrinterDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, pageable micromodels.Pageable) ([]models.PrinterActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, pageableStep micromodels.PageableStep) ([]models.PrinterDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, pageableStep micromodels.PageableStep) ([]models.PrinterActivity, error)
	FindStep(ctx context.Context, shopID string, searchInFields []string, q string, skip int, limit int, sorts map[string]int, projects map[string]interface{}) ([]models.PrinterInfo, int, error)
}

type PrinterRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.PrinterDoc]
	repositories.SearchRepository[models.PrinterInfo]
	repositories.GuidRepository[models.PrinterItemGuid]
	repositories.ActivityRepository[models.PrinterActivity, models.PrinterDeleteActivity]
}

func NewPrinterRepository(pst microservice.IPersisterMongo) PrinterRepository {

	insRepo := PrinterRepository{
		pst: pst,
	}
	insRepo.CrudRepository = repositories.NewCrudRepository[models.PrinterDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.PrinterInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.PrinterItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.PrinterActivity, models.PrinterDeleteActivity](pst)

	return insRepo
}
