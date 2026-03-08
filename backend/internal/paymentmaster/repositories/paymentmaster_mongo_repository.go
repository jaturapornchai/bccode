package repositories

import (
	"context"
	"smlcloudplatform/internal/paymentmaster/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
)

type IPaymentMasterRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(category models.PaymentMasterDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.PaymentMasterDoc) error
	Update(ctx context.Context, shopID string, guid string, category models.PaymentMasterDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Find(shopID string, searchInFields []string, q string) ([]models.PaymentMasterInfo, error)
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.PaymentMasterInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.PaymentMasterDoc, error)
}

type PaymentMasterRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.PaymentMasterDoc]
	repositories.SearchRepository[models.PaymentMasterInfo]
	repositories.GuidRepository[models.PaymentMasterItemGuid]
}

func NewPaymentMasterRepository(pst microservice.IPersisterMongo) PaymentMasterRepository {

	insRepo := PaymentMasterRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.PaymentMasterDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.PaymentMasterInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.PaymentMasterItemGuid](pst)

	return insRepo
}
