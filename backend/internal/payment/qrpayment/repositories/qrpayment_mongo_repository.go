package repositories

import (
	"context"
	"smlcloudplatform/internal/payment/qrpayment/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
)

type IQrPaymentRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.QrPaymentDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.QrPaymentDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.QrPaymentDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.QrPaymentInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.QrPaymentDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.QrPaymentItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.QrPaymentDoc, error)

	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, selectFields map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.QrPaymentInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.QrPaymentDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.QrPaymentActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.QrPaymentDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.QrPaymentActivity, error)
}

type QrPaymentRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.QrPaymentDoc]
	repositories.SearchRepository[models.QrPaymentInfo]
	repositories.GuidRepository[models.QrPaymentItemGuid]
	repositories.ActivityRepository[models.QrPaymentActivity, models.QrPaymentDeleteActivity]
}

func NewQrPaymentRepository(pst microservice.IPersisterMongo) *QrPaymentRepository {

	insRepo := &QrPaymentRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.QrPaymentDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.QrPaymentInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.QrPaymentItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.QrPaymentActivity, models.QrPaymentDeleteActivity](pst)

	return insRepo
}
