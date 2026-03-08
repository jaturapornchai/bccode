package repositories

import (
	"context"
	"smlcloudplatform/internal/productsection/sectiondepartment/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type ISectionDepartmentRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.SectionDepartmentDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.SectionDepartmentDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.SectionDepartmentDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.SectionDepartmentInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.SectionDepartmentDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.SectionDepartmentItemGuid, error)
	FindOneFilter(ctx context.Context, shopID string, filters map[string]interface{}) (models.SectionDepartmentDoc, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.SectionDepartmentDoc, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.SectionDepartmentInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.SectionDepartmentInfo, int, error)

	FindDeletedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SectionDepartmentDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SectionDepartmentActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SectionDepartmentDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, shopID string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SectionDepartmentActivity, error)

	FindOneByCode(ctx context.Context, shopID, branchCode, departmentCode string) (models.SectionDepartmentDoc, error)
}

type SectionDepartmentRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.SectionDepartmentDoc]
	repositories.SearchRepository[models.SectionDepartmentInfo]
	repositories.GuidRepository[models.SectionDepartmentItemGuid]
	repositories.ActivityRepository[models.SectionDepartmentActivity, models.SectionDepartmentDeleteActivity]
}

func NewSectionDepartmentRepository(pst microservice.IPersisterMongo) *SectionDepartmentRepository {

	insRepo := &SectionDepartmentRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.SectionDepartmentDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.SectionDepartmentInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.SectionDepartmentItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.SectionDepartmentActivity, models.SectionDepartmentDeleteActivity](pst)

	return insRepo
}

func (repo SectionDepartmentRepository) FindOneByCode(ctx context.Context, shopID, branchCode, departmentCode string) (models.SectionDepartmentDoc, error) {
	doc := models.SectionDepartmentDoc{}
	err := repo.pst.FindOne(
		ctx,
		models.SectionDepartmentDoc{}, bson.M{
			"shopid":         shopID,
			"deletedat":      bson.M{"$exists": false},
			"branchcode":     branchCode,
			"departmentcode": departmentCode,
		},
		&doc,
	)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
