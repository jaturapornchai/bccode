package repositories

import (
	"context"
	"smlcloudplatform/internal/product/color/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
)

type IColorRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.ColorDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ColorDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.ColorDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.ColorInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.ColorDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.ColorItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.ColorDoc, error)

	FindStep(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, selectFields map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ColorInfo, int, error)
}

type ColorRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ColorDoc]
	repositories.SearchRepository[models.ColorInfo]
	repositories.GuidRepository[models.ColorItemGuid]
}

func NewColorRepository(pst microservice.IPersisterMongo) *ColorRepository {

	insRepo := &ColorRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ColorDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ColorInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.ColorItemGuid](pst)

	return insRepo
}
