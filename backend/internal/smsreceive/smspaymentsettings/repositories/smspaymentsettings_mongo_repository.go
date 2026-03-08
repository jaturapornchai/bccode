package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/smsreceive/smspaymentsettings/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
)

type ISmsPaymentSettingsRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.SmsPaymentSettingsDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.SmsPaymentSettingsDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.SmsPaymentSettingsDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	FindOne(ctx context.Context, shopID string, filters interface{}) (models.SmsPaymentSettingsDoc, error)
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.SmsPaymentSettingsInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.SmsPaymentSettingsDoc, error)
}

type SmsPaymentSettingsRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.SmsPaymentSettingsDoc]
	repositories.SearchRepository[models.SmsPaymentSettingsInfo]
	repositories.GuidRepository[models.SmsPaymentSettingsItemGuid]
}

func NewSmsPaymentSettingsRepository(pst microservice.IPersisterMongo) SmsPaymentSettingsRepository {

	insRepo := SmsPaymentSettingsRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.SmsPaymentSettingsDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.SmsPaymentSettingsInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.SmsPaymentSettingsItemGuid](pst)

	return insRepo
}
