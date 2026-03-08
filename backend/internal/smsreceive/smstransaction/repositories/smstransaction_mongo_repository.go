package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/smsreceive/smstransaction/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type ISmsTransactionRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.SmsTransactionDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.SmsTransactionDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.SmsTransactionDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.SmsTransactionInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.SmsTransactionDoc, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.SmsTransactionDoc, error)

	FindFilterSms(ctx context.Context, shopID string, storefrontGUID string, address string, startTime time.Time, endTime time.Time) ([]models.SmsTransactionInfo, error)
}

type SmsTransactionRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.SmsTransactionDoc]
	repositories.SearchRepository[models.SmsTransactionInfo]
	repositories.GuidRepository[models.SmsTransactionItemGuid]
	contextTimeout time.Duration
}

func NewSmsTransactionRepository(pst microservice.IPersisterMongo) SmsTransactionRepository {

	contextTimeout := time.Duration(15) * time.Second

	insRepo := SmsTransactionRepository{
		pst:            pst,
		contextTimeout: contextTimeout,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.SmsTransactionDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.SmsTransactionInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.SmsTransactionItemGuid](pst)

	return insRepo
}

func (repo SmsTransactionRepository) FindFilterSms(ctx context.Context, shopID string, storefrontGUID string, address string, startTime time.Time, endTime time.Time) ([]models.SmsTransactionInfo, error) {

	filters := bson.M{
		"shopid":         shopID,
		"storefrontguid": storefrontGUID,
		"deletedat":      bson.M{"$exists": false},
		"address":        address,
		"status":         0,
		"createdat": bson.M{
			"$gte": startTime,
			"$lte": endTime,
		},
	}

	docList := []models.SmsTransactionInfo{}
	err := repo.pst.Find(ctx, models.SmsTransactionInfo{}, filters, &docList)

	if err != nil {
		return []models.SmsTransactionInfo{}, nil
	}

	return docList, nil
}
