package repositories

import (
	"context"
	"smlcloudplatform/internal/report/reportqueryc/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IReportQueryRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.ReportQueryDoc) (string, error)
	Update(ctx context.Context, holdingCode string, guid string, doc models.ReportQueryDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.ReportQueryInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.ReportQueryDoc, error)
	FindOne(ctx context.Context, holdingCode string, filters interface{}) (models.ReportQueryDoc, error)
	FindOneByCode(ctx context.Context, reportCode string) (models.ReportQueryDoc, error)

	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.ReportQueryDoc, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.ReportQueryInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.ReportQueryInfo, int, error)
}

type ReportQueryRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ReportQueryDoc]
	repositories.SearchRepository[models.ReportQueryInfo]
	repositories.GuidRepository[models.ReportQueryItemGuid]
	repositories.ActivityRepository[models.ReportQueryActivity, models.ReportQueryDeleteActivity]
}

func NewReportQueryRepository(pst microservice.IPersisterMongo) *ReportQueryRepository {

	insRepo := &ReportQueryRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ReportQueryDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ReportQueryInfo](pst)
	return insRepo
}

func (repo ReportQueryRepository) FindOneByCode(ctx context.Context, reportCode string) (models.ReportQueryDoc, error) {
	findDoc := models.ReportQueryDoc{}
	err := repo.pst.FindOne(ctx, models.ReportQueryDoc{}, bson.M{
		"code":       reportCode,
		"isapproved": true,
		"isactived":  true,
	}, &findDoc)
	if err != nil {
		return models.ReportQueryDoc{}, err
	}

	return findDoc, nil
}
