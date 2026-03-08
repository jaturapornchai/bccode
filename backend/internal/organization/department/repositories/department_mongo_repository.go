package repositories

import (
	"context"
	"smlcloudplatform/internal/organization/department/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IDepartmentRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.DepartmentDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.DepartmentDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.DepartmentDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.DepartmentInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.DepartmentDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.DepartmentItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.DepartmentDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.DepartmentInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.DepartmentInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DepartmentDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DepartmentActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DepartmentDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DepartmentActivity, error)

	FindOneByCode(ctx context.Context, shopID, branchCode, departmentCode string) (models.DepartmentDoc, error)
}

type DepartmentRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.DepartmentDoc]
	repositories.SearchRepository[models.DepartmentInfo]
	repositories.GuidRepository[models.DepartmentItemGuid]
	repositories.ActivityRepository[models.DepartmentActivity, models.DepartmentDeleteActivity]
}

func NewDepartmentRepository(pst microservice.IPersisterMongo) *DepartmentRepository {

	insRepo := &DepartmentRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.DepartmentDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.DepartmentInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.DepartmentItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.DepartmentActivity, models.DepartmentDeleteActivity](pst)

	return insRepo
}

func (repo DepartmentRepository) FindOneByCode(ctx context.Context, shopID string, branchCode, departmentCode string) (models.DepartmentDoc, error) {
	doc := models.DepartmentDoc{}
	err := repo.pst.FindOne(ctx,
		models.DepartmentDoc{},
		bson.M{
			"shopid":         shopID,
			"deletedat":      bson.M{"$exists": false},
			"branchcode":     branchCode,
			"departmentcode": departmentCode,
		}, &doc)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
