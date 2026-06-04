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
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.DepartmentDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.DepartmentDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.DepartmentDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.DepartmentInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.DepartmentDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.DepartmentItemGuid, error)
	FindOneFilter(ctx context.Context, holdingCode string, filters map[string]interface{}) (models.DepartmentDoc, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.DepartmentDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.DepartmentInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.DepartmentInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DepartmentDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DepartmentActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DepartmentDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DepartmentActivity, error)

	FindOneByCode(ctx context.Context, holdingCode, branchCode, departmentCode string) (models.DepartmentDoc, error)
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

func (repo DepartmentRepository) FindOneByCode(ctx context.Context, holdingCode string, branchCode, departmentCode string) (models.DepartmentDoc, error) {
	doc := models.DepartmentDoc{}
	filter := bson.M{
		"holding_code": holdingCode,
		"deleted_at":   bson.M{"$exists": false},
		"code":         departmentCode,
	}
	if branchCode != "" {
		filter["branchcode"] = branchCode
	}
	err := repo.pst.FindOne(ctx,
		models.DepartmentDoc{},
		filter,
		&doc)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
