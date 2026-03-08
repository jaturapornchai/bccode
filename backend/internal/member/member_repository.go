package member

import (
	"context"
	"smlcloudplatform/internal/member/models"
	"smlcloudplatform/internal/utils/search"
	"smlcloudplatform/pkg/microservice"

	micro_models "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IMemberRepository interface {
	Create(ctx context.Context, doc models.MemberDoc) (primitive.ObjectID, error)
	Update(ctx context.Context, guid string, doc models.MemberDoc) error
	FindByGuid(ctx context.Context, shopID string, guid string) (models.MemberDoc, error)

	FindByLineUID(ctx context.Context, lineUID string) (models.MemberDoc, error)
	FindPageFilter(ctx context.Context, shopID string, searchInFields []string, pageable micro_models.Pageable) ([]models.MemberInfo, mongopagination.PaginationData, error)
	FindStep(ctx context.Context, shopID string, searchInFields []string, projects map[string]interface{}, pageableStep micro_models.PageableStep) ([]models.MemberInfo, int, error)
}

type MemberRepository struct {
	pst microservice.IPersisterMongo
}

func NewMemberRepository(pst microservice.IPersisterMongo) *MemberRepository {

	insRepo := &MemberRepository{
		pst: pst,
	}

	return insRepo
}

func (repo MemberRepository) Create(ctx context.Context, doc models.MemberDoc) (primitive.ObjectID, error) {
	idx, err := repo.pst.Create(ctx, &models.MemberDoc{}, doc)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return idx, nil
}

func (repo MemberRepository) Update(ctx context.Context, guid string, doc models.MemberDoc) error {
	filterDoc := map[string]interface{}{
		"guidfixed": guid,
	}
	err := repo.pst.UpdateOne(ctx, &models.MemberDoc{}, filterDoc, doc)
	if err != nil {
		return err
	}
	return nil
}

func (repo MemberRepository) FindByGuid(ctx context.Context, shopID string, guid string) (models.MemberDoc, error) {
	doc := &models.MemberDoc{}
	err := repo.pst.FindOne(ctx, &models.MemberDoc{}, bson.M{"guidfixed": guid, "shops": shopID, "deletedat": bson.M{"$exists": false}}, doc)
	if err != nil {
		return *doc, err
	}
	return *doc, nil
}

func (repo MemberRepository) FindByLineUID(ctx context.Context, lineUID string) (models.MemberDoc, error) {
	doc := &models.MemberDoc{}
	err := repo.pst.FindOne(ctx, &models.MemberDoc{}, bson.M{"lineuid": lineUID, "deletedat": bson.M{"$exists": false}}, doc)
	if err != nil {
		return *doc, err
	}
	return *doc, nil
}

func (repo MemberRepository) FindPageFilter(ctx context.Context, shopID string, searchInFields []string, pageable micro_models.Pageable) ([]models.MemberInfo, mongopagination.PaginationData, error) {

	matchFilterList := []interface{}{}

	searchFilterQuery := search.CreateTextFilter(searchInFields, pageable.Query)

	queryFilters := bson.M{
		"shops":     shopID,
		"deletedat": bson.M{"$exists": false},
	}

	if len(matchFilterList) > 0 {
		queryFilters["$and"] = matchFilterList
	}

	if len(pageable.Sorts) < 1 {
		pageable.Sorts = append(pageable.Sorts, micro_models.KeyInt{Key: "createdat", Value: 1})
	}

	if len(searchFilterQuery) > 0 {
		if queryFilters["$or"] == nil {
			queryFilters["$or"] = searchFilterQuery
		} else {
			queryFilters["$or"] = append(queryFilters["$or"].([]interface{}), searchFilterQuery...)
		}
	}

	docList := []models.MemberInfo{}
	pagination, err := repo.pst.FindPage(ctx, new(models.MemberInfo), queryFilters, pageable, &docList)

	if err != nil {
		return []models.MemberInfo{}, mongopagination.PaginationData{}, err
	}

	return docList, pagination, nil
}

func (repo MemberRepository) FindStep(ctx context.Context, shopID string, searchInFields []string, projects map[string]interface{}, pageableStep micro_models.PageableStep) ([]models.MemberInfo, int, error) {

	filterQuery := bson.M{
		"shops":     shopID,
		"deletedat": bson.M{"$exists": false},
	}

	matchFilterList := []interface{}{}

	if len(matchFilterList) > 0 {
		filterQuery["$and"] = matchFilterList
	}

	searchFilterQuery := search.CreateTextFilter(searchInFields, pageableStep.Query)

	if len(searchFilterQuery) > 0 {
		if filterQuery["$or"] == nil {
			filterQuery["$or"] = searchFilterQuery
		} else {
			filterQuery["$or"] = append(filterQuery["$or"].([]interface{}), searchFilterQuery...)
		}
	}

	tempSkip := int64(pageableStep.Skip)
	tempLimit := int64(pageableStep.Limit)

	tempOptions := &options.FindOptions{}
	tempOptions.SetSkip(tempSkip)
	tempOptions.SetLimit(tempLimit)

	projectOptions := bson.M{}

	for key, val := range projects {
		projectOptions[key] = val
	}

	tempOptions.SetProjection(projectOptions)

	for _, pageSort := range pageableStep.Sorts {
		tempOptions.SetSort(bson.M{pageSort.Key: pageSort.Value})
	}

	if len(pageableStep.Sorts) < 1 {
		tempOptions.SetSort(bson.M{"createdat": 1})
	}

	docList := []models.MemberInfo{}
	err := repo.pst.Find(ctx, new(models.MemberInfo), filterQuery, &docList, tempOptions)

	if err != nil {
		return []models.MemberInfo{}, 0, err
	}

	count, err := repo.pst.Count(ctx, new(models.MemberInfo), filterQuery)

	if err != nil {
		return []models.MemberInfo{}, 0, err
	}

	return docList, count, nil
}
