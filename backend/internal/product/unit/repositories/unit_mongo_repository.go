package repositories

import (
	"context"
	"errors"
	"os"
	"smlcloudplatform/internal/product/unit/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IUnitRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.UnitDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.UnitDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.UnitDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.UnitInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.UnitDoc, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]models.UnitDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.UnitItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.UnitDoc, error)

	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.UnitInfo, mongopagination.PaginationData, error)
	FindPageFilterNoShopid(ctx context.Context, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.UnitInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, selectFields map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.UnitInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.UnitDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.UnitActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.UnitDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.UnitActivity, error)
	FindMasterInCodes(ctx context.Context, codes []string) ([]models.UnitInfo, error)

	FindByUnitCodes(ctx context.Context, shopID string, unitCodes []string) ([]models.UnitInfo, error)

	Transaction(ctx context.Context, callback func(ctx context.Context) error) error
}

type UnitRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.UnitDoc]
	repositories.SearchRepository[models.UnitInfo]
	repositories.GuidRepository[models.UnitItemGuid]
	repositories.ActivityRepository[models.UnitActivity, models.UnitDeleteActivity]
}

func NewUnitRepository(pst microservice.IPersisterMongo) *UnitRepository {

	insRepo := &UnitRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.UnitDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.UnitInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.UnitItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.UnitActivity, models.UnitDeleteActivity](pst)

	return insRepo
}

func (repo UnitRepository) FindMasterInCodes(ctx context.Context, codes []string) ([]models.UnitInfo, error) {

	masterShopID := os.Getenv("MASTER_SHOP_ID")

	if len(masterShopID) == 0 {
		return []models.UnitInfo{}, errors.New("master shop id is empty")
	}

	docList := []models.UnitInfo{}

	filters := bson.M{
		"shopid": masterShopID,
		"unitcode": bson.M{
			"$in": codes,
		},
	}

	err := repo.pst.Find(ctx, []models.UnitInfo{}, filters, &docList)

	if err != nil {
		return nil, err
	}

	return docList, nil
}

func (repo UnitRepository) FindByUnitCodes(ctx context.Context, shopID string, unitCodes []string) ([]models.UnitInfo, error) {

	filters := bson.M{
		"shopid":    shopID,
		"deletedat": bson.M{"$exists": false},
		"unitcode":  bson.M{"$in": unitCodes},
	}

	var results []models.UnitInfo
	err := repo.pst.Find(ctx, models.UnitInfo{}, filters, &results)

	if err != nil {
		return nil, err
	}

	return results, nil
}

func (repo UnitRepository) Transaction(ctx context.Context, callback func(ctx context.Context) error) error {
	return repo.pst.Transaction(ctx, callback)
}
