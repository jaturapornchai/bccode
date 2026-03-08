package repositories

import (
	"context"
	"smlcloudplatform/pkg/microservice"
)

type ITableDesign interface {
	Create(ctx context.Context, zoneGuidfixed string, tableIndex int, doc TableDesign) error
}

type TableDesign struct {
}

type TableDesignRepository struct {
	pst microservice.IPersisterMongo
}

func NewTableDesignRepository(pst microservice.IPersisterMongo) ZoneDesignRepository {
	insRepo := ZoneDesignRepository{
		pst: pst,
	}
	return insRepo

}

func (repo TableDesignRepository) Create(ctx context.Context, zoneGuidfixed string, tableIndex int, doc TableDesign) error {
	return repo.pst.UpdateOne(
		ctx,
		doc,
		map[string]interface{}{
			"guidfixed":    zoneGuidfixed,
			"tables.index": tableIndex,
		}, doc)
}
