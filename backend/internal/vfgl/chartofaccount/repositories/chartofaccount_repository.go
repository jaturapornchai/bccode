package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/vfgl/chartofaccount/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
)

type IChartOfAccountRepository interface {
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, category models.ChartOfAccountDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ChartOfAccountDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.ChartOfAccountDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	FindOne(ctx context.Context, holdingCode string, filters interface{}) (models.ChartOfAccountDoc, error)
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.ChartOfAccountInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.ChartOfAccountDoc, error)
}

type ChartOfAccountRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ChartOfAccountDoc]
	repositories.SearchRepository[models.ChartOfAccountInfo]
	repositories.GuidRepository[models.ChartOfAccountIndentityId]
}

func NewChartOfAccountRepository(pst microservice.IPersisterMongo) ChartOfAccountRepository {
	repo := ChartOfAccountRepository{
		pst: pst,
	}

	repo.CrudRepository = repositories.NewCrudRepository[models.ChartOfAccountDoc](pst)
	repo.SearchRepository = repositories.NewSearchRepository[models.ChartOfAccountInfo](pst)
	repo.GuidRepository = repositories.NewGuidRepository[models.ChartOfAccountIndentityId](pst)
	return repo
}
