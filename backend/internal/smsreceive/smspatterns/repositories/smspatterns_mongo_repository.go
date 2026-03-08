package repositories

import (
	"context"
	"smlcloudplatform/internal/smsreceive/smspatterns/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ISmsPatternsRepository interface {
	Create(ctx context.Context, doc models.SmsPatternsDoc) (string, error)
	UpdateByGuid(ctx context.Context, guid string, doc models.SmsPatternsDoc) error
	DeleteByGuid(ctx context.Context, guid string) error
	FindByCode(ctx context.Context, code string) (models.SmsPatternsDoc, error)
	FindByGuid(ctx context.Context, guid string) (models.SmsPatternsDoc, error)
	FindPage(ctx context.Context, searchInFields []string, pageable micromodels.Pageable) ([]models.SmsPatternsInfo, mongopagination.PaginationData, error)
}

type SmsPatternsRepository struct {
	pst microservice.IPersisterMongo
}

func NewSmsPatternsRepository(pst microservice.IPersisterMongo) *SmsPatternsRepository {

	insRepo := &SmsPatternsRepository{
		pst: pst,
	}

	return insRepo
}

func (repo SmsPatternsRepository) Create(ctx context.Context, doc models.SmsPatternsDoc) (string, error) {

	idx, err := repo.pst.Create(ctx, models.SmsPatternsDoc{}, doc)

	if err != nil {
		return "", err
	}

	return idx.Hex(), nil
}

func (repo SmsPatternsRepository) UpdateByGuid(ctx context.Context, guid string, doc models.SmsPatternsDoc) error {
	filterDoc := map[string]interface{}{
		"guidfixed": guid,
	}

	err := repo.pst.UpdateOne(ctx, models.SmsPatternsDoc{}, filterDoc, doc)

	if err != nil {
		return err
	}

	return nil
}

func (repo SmsPatternsRepository) UpdateByCode(ctx context.Context, code string, doc models.SmsPatternsDoc) error {
	filterDoc := map[string]interface{}{
		"code": code,
	}

	err := repo.pst.UpdateOne(ctx, models.SmsPatternsDoc{}, filterDoc, doc)

	if err != nil {
		return err
	}

	return nil
}

func (repo SmsPatternsRepository) DeleteByGuid(ctx context.Context, guid string) error {
	filterDoc := map[string]interface{}{
		"guidfixed": guid,
	}

	err := repo.pst.Delete(ctx, models.SmsPatternsDoc{}, filterDoc)

	if err != nil {
		return err
	}

	return nil
}

func (repo SmsPatternsRepository) FindByCode(ctx context.Context, code string) (models.SmsPatternsDoc, error) {

	doc := models.SmsPatternsDoc{}

	filters := bson.M{
		"code": code,
	}

	err := repo.pst.FindOne(ctx, models.SmsPatternsDoc{}, filters, &doc)

	if err != nil {
		return models.SmsPatternsDoc{}, err
	}

	return doc, nil
}

func (repo SmsPatternsRepository) FindByGuid(ctx context.Context, guidFixed string) (models.SmsPatternsDoc, error) {

	doc := models.SmsPatternsDoc{}

	filters := bson.M{
		"guidfixed": guidFixed,
	}

	err := repo.pst.FindOne(ctx, models.SmsPatternsDoc{}, filters, &doc)

	if err != nil {
		return models.SmsPatternsDoc{}, err
	}

	return doc, nil
}

func (repo SmsPatternsRepository) FindPage(ctx context.Context, searchInFields []string, pageable micromodels.Pageable) ([]models.SmsPatternsInfo, mongopagination.PaginationData, error) {

	searchFilterList := []interface{}{}

	for _, colName := range searchInFields {
		searchFilterList = append(searchFilterList, bson.M{colName: bson.M{"$regex": primitive.Regex{
			Pattern: ".*" + pageable.Query + ".*",
			Options: "",
		}}})
	}

	filterQueries := bson.M{
		"$or": searchFilterList,
	}

	docList := []models.SmsPatternsInfo{}
	pagination, err := repo.pst.FindPage(ctx, models.SmsPatternsInfo{}, filterQueries, pageable, &docList)

	if err != nil {
		return []models.SmsPatternsInfo{}, mongopagination.PaginationData{}, err
	}

	return docList, pagination, nil
}
