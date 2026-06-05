package repositories

import (
	"context"
	"smlcloudplatform/pkg/microservice"

	"go.mongodb.org/mongo-driver/bson"
)

type GuidRepository[T any] struct {
	pst microservice.IPersisterMongo
}

func NewGuidRepository[T any](pst microservice.IPersisterMongo) GuidRepository[T] {
	return GuidRepository[T]{
		pst: pst,
	}
}

func (repo GuidRepository[T]) FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]T, error) {

	findDoc := []T{}
	err := repo.pst.Find(ctx, new(T), bson.M{"holdingcode": holdingCode, columnName: bson.M{"$in": itemGuidList}, "deletedat": bson.M{"$exists": false}}, &findDoc)

	if err != nil {
		return []T{}, err
	}
	return findDoc, nil
}

func (repo GuidRepository[T]) FindInItemGuids(ctx context.Context, holdingCode string, columnName string, itemGuidList []interface{}) ([]T, error) {

	findDoc := []T{}
	err := repo.pst.Find(ctx, new(T), bson.M{"holdingcode": holdingCode, columnName: bson.M{"$in": itemGuidList}, "deletedat": bson.M{"$exists": false}}, &findDoc)

	if err != nil {
		return []T{}, err
	}
	return findDoc, nil
}
