package repositories

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type ICRUDRepository[T any] interface {
	Count(ctx context.Context, shopID string) (int, error)
	CountByKey(shopID string, keyName string, keyValue []string) (int, error)
	CountByInKeys(shopID string, keyName string, keyValues []string) (int, error)

	Create(ctx context.Context, doc T) (string, error)
	CreateInBatch(ctx context.Context, docList []T) error
	Update(ctx context.Context, shopID string, guid string, doc T) error
	Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	FindOne(ctx context.Context, shopID string, filters interface{}) (T, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (T, error)
	FindByGuids(ctx context.Context, shopID string, guids []string) ([]T, error)
	FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (T, error)
	FindByDocIndentityGuids(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) ([]T, error)
	FindOneFilter(ctx context.Context, shopID string, filters map[string]interface{}) (T, error)
}

type CrudRepository[T any] struct {
	pst microservice.IPersisterMongo
}

func NewCrudRepository[T any](pst microservice.IPersisterMongo) CrudRepository[T] {
	return CrudRepository[T]{
		pst: pst,
	}
}

func (repo CrudRepository[T]) Count(ctx context.Context, shopID string) (int, error) {

	count, err := repo.pst.Count(ctx, new(T), bson.M{"shopid": shopID})

	if err != nil {
		return 0, err
	}
	return count, nil
}

func (repo CrudRepository[T]) CountByKey(ctx context.Context, shopID string, keyName string, keyValue string) (int, error) {

	filters := bson.M{
		"shopid":    shopID,
		"deletedat": bson.M{"$exists": false},
		keyName:     keyValue,
	}

	return repo.pst.Count(ctx, new(T), filters)
}

func (repo CrudRepository[T]) CountByInKeys(ctx context.Context, shopID string, keyName string, keyValues []string) (int, error) {

	filters := bson.M{
		"shopid":    shopID,
		"deletedat": bson.M{"$exists": false},
		keyName:     bson.M{"$in": keyValues},
	}

	return repo.pst.Count(ctx, new(T), filters)
}

func (repo CrudRepository[T]) Create(ctx context.Context, doc T) (string, error) {
	idx, err := repo.pst.Create(ctx, new(T), doc)

	if err != nil {
		return "", err
	}

	return idx.Hex(), nil
}

func (repo CrudRepository[T]) CreateInBatch(ctx context.Context, docList []T) error {
	var tempList []interface{}

	for _, inv := range docList {
		tempList = append(tempList, inv)
	}

	err := repo.pst.CreateInBatch(ctx, new(T), tempList)

	if err != nil {
		return err
	}
	return nil
}

func (repo CrudRepository[T]) Update(ctx context.Context, shopID string, guid string, doc T) error {
	filterDoc := map[string]interface{}{
		"shopid":    shopID,
		"guidfixed": guid,
	}

	err := repo.pst.UpdateOne(ctx, new(T), filterDoc, doc)

	if err != nil {
		return err
	}

	return nil
}

func (repo CrudRepository[T]) Delete(ctx context.Context, shopID string, username string, filters map[string]interface{}) error {

	filterQuery := bson.M{}

	for col, val := range filters {
		filterQuery[col] = val
	}

	filterQuery["shopid"] = shopID

	err := repo.pst.SoftDelete(ctx, new(T), username, filterQuery)

	if err != nil {
		return err
	}

	return nil
}

func (repo CrudRepository[T]) DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error {
	err := repo.pst.SoftDelete(ctx, new(T), username, bson.M{"guidfixed": guid, "shopid": shopID})

	if err != nil {
		return err
	}

	return nil
}

func (repo CrudRepository[T]) FindOne(ctx context.Context, shopID string, filters interface{}) (T, error) {

	var filterQuery interface{}

	switch filters.(type) {
	case bson.M:
		tempFilterQuery := filters.(bson.M)
		tempFilterQuery["shopid"] = shopID
		tempFilterQuery["deletedat"] = bson.M{"$exists": false}
		filterQuery = tempFilterQuery
	case bson.D:
		tempFilterQuery := filters.(bson.D)
		tempFilterQuery = append(tempFilterQuery, bson.E{"shopid", shopID})
		tempFilterQuery = append(tempFilterQuery, bson.E{"deletedat", bson.D{{"$exists", false}}})

		filterQuery = tempFilterQuery
	default:
		return *new(T), errors.New("invalid query filter type")
	}

	doc := new(T)

	err := repo.pst.FindOne(ctx, new(T), filterQuery, doc)

	if err != nil {
		return *new(T), err
	}

	return *doc, nil
}

func (repo CrudRepository[T]) FindByGuid(ctx context.Context, shopID string, guid string) (T, error) {

	doc := new(T)

	err := repo.pst.FindOne(
		ctx,
		new(T),
		bson.M{"guidfixed": guid, "shopid": shopID, "deletedat": bson.M{"$exists": false}},
		doc,
	)

	if err != nil {
		return *new(T), err
	}

	return *doc, nil
}

func (repo CrudRepository[T]) FindByGuids(ctx context.Context, shopID string, guids []string) ([]T, error) {

	doc := new([]T)

	err := repo.pst.Find(
		ctx,
		new(T),
		bson.M{"guidfixed": bson.M{"$in": guids}, "shopid": shopID, "deletedat": bson.M{"$exists": false}},
		doc,
	)

	if err != nil {
		return *new([]T), err
	}

	return *doc, nil
}

func (repo CrudRepository[T]) FindByDocIndentityGuid(ctx context.Context, shopID string, indentityField string, indentityValue interface{}) (T, error) {

	doc := new(T)

	err := repo.pst.FindOne(ctx, new(T), bson.M{"shopid": shopID, "deletedat": bson.M{"$exists": false}, indentityField: indentityValue}, doc)

	if err != nil {
		return *new(T), err
	}

	return *doc, nil
}

func (repo CrudRepository[T]) FindByDocIndentityGuids(ctx context.Context, shopID string, indentityField string, indentityValues interface{}) ([]T, error) {

	var values interface{}
	switch v := indentityValues.(type) {
	case []int, []string:
		values = indentityValues
	case int, string:
		values = []interface{}{indentityValues}
	default:
		return nil, fmt.Errorf("unsupported input type: %T", v)
	}

	doc := new([]T)

	err := repo.pst.Find(ctx, new(T), bson.M{"shopid": shopID, "deletedat": bson.M{"$exists": false}, indentityField: bson.M{"$in": values}}, doc)

	if err != nil {
		return *new([]T), err
	}

	return *doc, nil
}

func (repo CrudRepository[T]) FindOneFilter(ctx context.Context, shopID string, filters map[string]interface{}) (T, error) {

	doc := new(T)

	findFilters := bson.M{}

	for col, val := range filters {
		findFilters[col] = val
	}

	findFilters["shopid"] = shopID
	findFilters["deletedat"] = bson.M{"$exists": false}

	err := repo.pst.FindOne(ctx, new(T), findFilters, doc)

	if err != nil {
		return *new(T), err
	}

	return *doc, nil
}

func (repo CrudRepository[T]) FindFilter(ctx context.Context, shopID string, filters map[string]interface{}) ([]T, error) {

	doc := new([]T)

	findFilters := bson.M{}

	for col, val := range filters {
		findFilters[col] = val
	}

	findFilters["shopid"] = shopID
	findFilters["deletedat"] = bson.M{"$exists": false}

	err := repo.pst.Find(ctx, new(T), findFilters, doc)

	if err != nil {
		return *new([]T), err
	}

	return *doc, nil
}
