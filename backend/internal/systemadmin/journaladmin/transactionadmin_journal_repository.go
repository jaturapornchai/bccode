package journal

import (
	"context"
	journalModels "smlcloudplatform/internal/vfgl/journal/models"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IJournalTransactionAdminRepository interface {
	FindJournalTransactionDocByShopID(ctx context.Context, shopID string, isDeleted bool, pageable msModels.Pageable) ([]journalModels.JournalDoc, mongopagination.PaginationData, error)
}

type JournalTransactionAdminRepository struct {
	pst microservice.IPersisterMongo
}

func NewJournalTransactionAdminRepository(pst microservice.IPersisterMongo) IJournalTransactionAdminRepository {
	return &JournalTransactionAdminRepository{
		pst: pst,
	}
}

func (r JournalTransactionAdminRepository) FindJournalTransactionDocByShopID(ctx context.Context, shopID string, isDeleted bool, pageable msModels.Pageable) ([]journalModels.JournalDoc, mongopagination.PaginationData, error) {

	docList := []journalModels.JournalDoc{}

	queryFilters := bson.M{
		"shopid":    shopID,
		"deletedat": bson.M{"$exists": isDeleted},
	}

	pagination, err := r.pst.FindPage(ctx, &journalModels.JournalDoc{}, queryFilters, pageable, &docList)
	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return docList, pagination, nil
}
