package repositories

import (
	"context"
	"smlcloudplatform/internal/organization/jobproject/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IJobProjectRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.JobProjectDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.JobProjectDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.JobProjectDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.JobProjectInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.JobProjectDoc, error)

	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.JobProjectItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.JobProjectDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.JobProjectInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.JobProjectInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.JobProjectDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.JobProjectActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.JobProjectDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.JobProjectActivity, error)

	FindOneByCode(ctx context.Context, holdingCode, branchCode, jobProjectCode string) (models.JobProjectDoc, error)
}

type JobProjectRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.JobProjectDoc]
	repositories.SearchRepository[models.JobProjectInfo]
	repositories.GuidRepository[models.JobProjectItemGuid]
	repositories.ActivityRepository[models.JobProjectActivity, models.JobProjectDeleteActivity]
}

func NewJobProjectRepository(pst microservice.IPersisterMongo) *JobProjectRepository {

	insRepo := &JobProjectRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.JobProjectDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.JobProjectInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.JobProjectItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.JobProjectActivity, models.JobProjectDeleteActivity](pst)

	return insRepo
}

func (repo JobProjectRepository) FindOneByCode(ctx context.Context, holdingCode string, branchCode, jobProjectCode string) (models.JobProjectDoc, error) {
	doc := models.JobProjectDoc{}
	err := repo.pst.FindOne(ctx,
		models.JobProjectDoc{},
		bson.M{
			"holdingcode":    holdingCode,
			"deletedat":      bson.M{"$exists": false},
			"branchcode":     branchCode,
			"jobprojectcode": jobProjectCode,
		}, &doc)

	if err != nil {
		return doc, err
	}

	return doc, nil
}
