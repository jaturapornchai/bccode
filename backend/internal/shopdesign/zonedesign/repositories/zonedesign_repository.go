package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/shopdesign/zonedesign/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
)

type IZoneDesignRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.ZoneDesignDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.ZoneDesignDoc) error
	Update(ctx context.Context, shopID string, guid string, category models.ZoneDesignDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	FindOne(ctx context.Context, shopID string, filters interface{}) (models.ZoneDesignDoc, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.ZoneDesignDoc, error)
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.ZoneDesignInfo, mongopagination.PaginationData, error)
}

type ZoneDesignRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.ZoneDesignDoc]
	repositories.SearchRepository[models.ZoneDesignInfo]
}

func NewZoneDesignRepository(pst microservice.IPersisterMongo) ZoneDesignRepository {
	insRepo := ZoneDesignRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.ZoneDesignDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.ZoneDesignInfo](pst)

	return insRepo

}
