package chartofaccountadmin

import (
	"context"
	chartOfAccountModels "smlcloudplatform/internal/vfgl/chartofaccount/models"
	"smlcloudplatform/pkg/microservice"
	msModels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IChartOfAccountAdminRepository interface {
	FindChartOfAccountDocByShopID(ctx context.Context, shopID string, isDeleted bool, pageable msModels.Pageable) ([]chartOfAccountModels.ChartOfAccountDoc, mongopagination.PaginationData, error)
}

type ChartOfAccountAdminRepository struct {
	pst microservice.IPersisterMongo
}

func NewChartOfAccountAdminRepository(pst microservice.IPersisterMongo) IChartOfAccountAdminRepository {
	return &ChartOfAccountAdminRepository{
		pst: pst,
	}
}

func (r ChartOfAccountAdminRepository) FindChartOfAccountDocByShopID(ctx context.Context, shopID string, isDeleted bool, pageable msModels.Pageable) ([]chartOfAccountModels.ChartOfAccountDoc, mongopagination.PaginationData, error) {

	docList := []chartOfAccountModels.ChartOfAccountDoc{}

	// err := r.pst.Find(ctx, &chartOfAccountModels.ChartOfAccountDoc{}, bson.M{"shopid": shopID}, &docList)
	// if err != nil {
	// 	return nil, err
	// }

	queryFilters := bson.M{
		"shopid":    shopID,
		"deletedat": bson.M{"$exists": isDeleted},
	}

	pagination, err := r.pst.FindPage(ctx, &chartOfAccountModels.ChartOfAccountDoc{}, queryFilters, pageable, &docList)
	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}

	return docList, pagination, nil
}
