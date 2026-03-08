package repositories

import (
	"context"
	"smlcloudplatform/internal/product/optionpattern/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
)

type IOptionPatternRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.OptionPatternDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.OptionPatternDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.OptionPatternDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.OptionPatternInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.OptionPatternDoc, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.OptionPatternItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (models.OptionPatternDoc, error)
}

type OptionPatternRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.OptionPatternDoc]
	repositories.SearchRepository[models.OptionPatternInfo]
	repositories.GuidRepository[models.OptionPatternItemGuid]
}

func NewOptionPatternRepository(pst microservice.IPersisterMongo) *OptionPatternRepository {

	insRepo := &OptionPatternRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.OptionPatternDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.OptionPatternInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.OptionPatternItemGuid](pst)

	return insRepo
}
