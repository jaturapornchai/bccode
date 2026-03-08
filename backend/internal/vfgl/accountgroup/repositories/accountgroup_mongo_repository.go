package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/vfgl/accountgroup/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
)

type IAccountGroupMongoRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.AccountGroupDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.AccountGroupDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.AccountGroupDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	FindOne(ctx context.Context, shopID string, filters interface{}) (models.AccountGroupDoc, error)
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.AccountGroupInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.AccountGroupDoc, error)
}

type AccountGroupMongoRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.AccountGroupDoc]
	repositories.SearchRepository[models.AccountGroupInfo]
	repositories.GuidRepository[models.AccountGroupIdentifier]
}

func NewAccountGroupMongoRepository(pst microservice.IPersisterMongo) AccountGroupMongoRepository {

	insRepo := AccountGroupMongoRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.AccountGroupDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.AccountGroupInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.AccountGroupIdentifier](pst)

	return insRepo
}
