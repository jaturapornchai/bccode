package repositories

import (
	"context"
	"smlcloudplatform/internal/documentwarehouse/documentimage/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IDocumentImageRepository interface {
	Create(ctx context.Context, doc models.DocumentImageDoc) (string, error)
	CreateInBatch(ctx context.Context, doc []models.DocumentImageDoc) error
	Update(ctx context.Context, shopID string, guid string, doc models.DocumentImageDoc) error
	DeleteByGuidfixed(ctx context.Context, shopID string, guid string, username string) error
	FindOne(ctx context.Context, shopID string, filters interface{}) (models.DocumentImageDoc, error)
	FindByReferenceDocNo(ctx context.Context, shopID string, docNo string) ([]models.DocumentImageDoc, error)
	FindByReference(ctx context.Context, shopID string, reference models.Reference) ([]models.DocumentImageDoc, error)
	FindByGuid(ctx context.Context, shopID string, guid string) (models.DocumentImageDoc, error)

	FindPage(ctx context.Context, shopID string, searchInFields []string, pageable micromodels.Pageable) ([]models.DocumentImageInfo, mongopagination.PaginationData, error)
	FindPageFilter(ctx context.Context, shopID string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.DocumentImageInfo, mongopagination.PaginationData, error)

	FindInItemGuid(ctx context.Context, shopID string, columnName string, itemGuidList []string) ([]models.DocumentImageItemGuid, error)
	FindInGUIDs(ctx context.Context, shopID string, docImageGUIDs []string) ([]models.DocumentImageDoc, error)

	FindAll(ctx context.Context) ([]models.DocumentImageDoc, error)
	UpdateAll(ctx context.Context, doc models.DocumentImageDoc) error

	// UpdateDocumentImageStatus(shopID string, guid string, docnoGUIDRef string, status int8) error
	// UpdateDocumentImageStatusByDocumentRef(shopID string, docRef string, docnoGUIDRef string, status int8) error
	// SaveDocumentImageDocRefGroup(shopID string, docRef string, docImages []models.DocumentImageGroup) error
	// ListDocumentImageGroup(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DocumentImageGroup, mongopagination.PaginationData, error)
	// GetDocumentImageGroup(shopID string, docRef string) (models.DocumentImageGroup, error)
}

type DocumentImageRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.DocumentImageDoc]
	repositories.SearchRepository[models.DocumentImageInfo]
	repositories.GuidRepository[models.DocumentImageItemGuid]
}

func NewDocumentImageRepository(pst microservice.IPersisterMongo) DocumentImageRepository {
	insRepo := DocumentImageRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.DocumentImageDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.DocumentImageInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.DocumentImageItemGuid](pst)

	return insRepo
}

func (repo DocumentImageRepository) FindInGUIDs(ctx context.Context, shopID string, docImageGUIDs []string) ([]models.DocumentImageDoc, error) {
	docList := []models.DocumentImageDoc{}
	err := repo.pst.Find(
		ctx,
		models.DocumentImageDoc{},
		bson.M{
			"guidfixed": bson.M{
				"$in": docImageGUIDs,
			},
			"deletedat": bson.M{"$exists": false},
		}, &docList)

	if err != nil {
		return []models.DocumentImageDoc{}, err
	}

	return docList, nil
}

func (repo DocumentImageRepository) FindByReferenceDocNo(ctx context.Context, shopID string, docNo string) ([]models.DocumentImageDoc, error) {
	docList := []models.DocumentImageDoc{}
	err := repo.pst.Find(
		ctx,
		models.DocumentImageDoc{},
		bson.M{
			"references.docno": docNo,
			"deletedat":        bson.M{"$exists": false},
		},
		&docList,
	)

	if err != nil {
		return []models.DocumentImageDoc{}, err
	}

	return docList, nil
}

func (repo DocumentImageRepository) FindByReference(ctx context.Context, shopID string, reference models.Reference) ([]models.DocumentImageDoc, error) {
	docList := []models.DocumentImageDoc{}
	err := repo.pst.Find(
		ctx,
		models.DocumentImageDoc{},
		bson.M{
			"references.module": reference.Module,
			"references.docno":  reference.DocNo,
			"deletedat":         bson.M{"$exists": false},
		},
		&docList,
	)

	if err != nil {
		return []models.DocumentImageDoc{}, err
	}

	return docList, nil
}

func (repo DocumentImageRepository) UpdateReject(ctx context.Context, shopID string, authUsername string, updatedAt time.Time, docImageGUID string, isReject bool) error {
	fillter := bson.M{
		"shopid":    shopID,
		"guidfixed": docImageGUID,
	}

	data := bson.M{
		"$set": bson.M{"isreject": isReject, "updatedby": authUsername, "updatedat": updatedAt},
	}

	return repo.pst.Update(ctx,
		models.DocumentImageDoc{},
		fillter,
		data,
	)
}

func (repo DocumentImageRepository) FindAll(ctx context.Context) ([]models.DocumentImageDoc, error) {
	docList := []models.DocumentImageDoc{}
	err := repo.pst.Find(ctx, models.DocumentImageDoc{}, bson.M{}, &docList)

	if err != nil {
		return []models.DocumentImageDoc{}, err
	}

	return docList, nil
}

func (repo DocumentImageRepository) UpdateAll(ctx context.Context, doc models.DocumentImageDoc) error {
	filterDoc := map[string]interface{}{
		"guidfixed": doc.GuidFixed,
	}

	err := repo.pst.UpdateOne(ctx, models.DocumentImageDoc{}, filterDoc, doc)

	if err != nil {
		return err
	}

	return nil
}
