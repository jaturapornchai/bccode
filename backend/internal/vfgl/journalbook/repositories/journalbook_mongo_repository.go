package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/vfgl/journalbook/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
)

type IJournalBookMongoRepository interface {
	Count(ctx context.Context, shopID string) (int, error)
	Create(ctx context.Context, doc models.JournalBookDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.JournalBookDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.JournalBookDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	FindOne(ctx context.Context, shopID string, filters interface{}) (models.JournalBookDoc, error)
	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.JournalBookInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.JournalBookDoc, error)
}

type JournalBookMongoRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.JournalBookDoc]
	repositories.SearchRepository[models.JournalBookInfo]
	repositories.GuidRepository[models.JournalBookIdentifier]
}

func NewJournalBookMongoRepository(pst microservice.IPersisterMongo) JournalBookMongoRepository {

	insRepo := JournalBookMongoRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.JournalBookDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.JournalBookInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.JournalBookIdentifier](pst)

	return insRepo
}
