package repositories

import (
	"context"
	"smlcloudplatform/internal/documentwarehouse/documentimage/models"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/utils/mogoutil"
	"smlcloudplatform/pkg/microservice"

	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IDocumentImageGroupRepository interface {
	CountByTask(ctx context.Context, holdingCode string, taskGUID string) (int, error)
	CountRejectByTask(ctx context.Context, holdingCode string, taskGUID string) (int, error)
	Create(ctx context.Context, doc models.DocumentImageGroupDoc) (string, error)
	CreateInBatch(ctx context.Context, doc []models.DocumentImageGroupDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.DocumentImageGroupDoc) error
	UpdateXOrder(ctx context.Context, holdingCode string, taskGUID string, GUID string, xorder uint) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string) error
	DeleteByDocumentImageGUIDsHasOne(ctx context.Context, holdingCode string, imageGUIDs []string) error
	DeleteByGUIDsIsDocumentImageEmpty(ctx context.Context, holdingCode string, GUIDs []string) error
	RemoveDocumentImageByDocumentImageGUIDs(ctx context.Context, holdingCode string, imageGUIDs []string) error
	DeleteByDocumentImageGUIDsHasOneWithoutDocumentImageGroupGUID(ctx context.Context, holdingCode string, withoutGUID string, imageGUIDs []string) error
	DeleteByGUIDsIsDocumentImageEmptyWithoutDocumentImageGroupGUID(ctx context.Context, holdingCode string, withoutGUID string, GUIDs []string) error
	DeleteByGUIDIsDocumentImageEmpty(ctx context.Context, holdingCode string, imageGroupGUID string) error
	RemoveDocumentImageByDocumentImageGUIDsWithoutDocumentImageGroupGUID(ctx context.Context, holdingCode string, withoutGUID string, imageGUIDs []string) error
	FindOne(ctx context.Context, holdingCode string, filters interface{}) (models.DocumentImageGroupDoc, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.DocumentImageGroupDoc, error)

	FindStatusByDocumentImageGroupTask(ctx context.Context, holdingCode string, taskGUID string) ([]models.DocumentImageGroupStatus, error)
	FindLastOneByTask(ctx context.Context, holdingCode string, taskGUID string) (models.DocumentImageGroupDoc, error)
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.DocumentImageGroupInfo, mongopagination.PaginationData, error)
	FindPageFilter(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.DocumentImageGroupInfo, mongopagination.PaginationData, error)
	FindByTaskGUID(ctx context.Context, holdingCode string, taskGUID string) ([]models.DocumentImageGroupDoc, error)

	UpdateTaskIsCompletedByTaskGUID(ctx context.Context, holdingCode string, taskGUID string, isCompleted bool) error
	FindOneByReference(ctx context.Context, holdingCode string, reference models.Reference) (models.DocumentImageGroupDoc, error)
	FindOneByDocumentImageGUID(ctx context.Context, holdingCode string, documentImageGUID string) (models.DocumentImageGroupDoc, error)
	FindByDocumentImageGUIDs(ctx context.Context, holdingCode string, documentImageGUIDs []string) ([]models.DocumentImageGroupInfo, error)
	FindByReference(ctx context.Context, holdingCode string, reference models.Reference) ([]models.DocumentImageGroupDoc, error)
	FindByReferenceDocNo(ctx context.Context, holdingCode string, docNo string) ([]models.DocumentImageGroupDoc, error)
	FindPageImageGroup(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.DocumentImageGroupInfo, mongopagination.PaginationData, error)
	Transaction(ctx context.Context, fnc func(ctx context.Context) error) error

	FindOneByDocumentImageGUIDAll(ctx context.Context, documentImageGUID string) (models.DocumentImageGroupDoc, error)
	UpdateStatusByTask(ctx context.Context, holdingCode string, taskGUID string, status int8) error
}

type DocumentImageGroupRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.DocumentImageGroupDoc]
	repositories.SearchRepository[models.DocumentImageGroupInfo]
}

func NewDocumentImageGroupRepository(pst microservice.IPersisterMongo) DocumentImageGroupRepository {
	insRepo := DocumentImageGroupRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.DocumentImageGroupDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.DocumentImageGroupInfo](pst)

	return insRepo
}

func (repo DocumentImageGroupRepository) Transaction(ctx context.Context, fnc func(ctx context.Context) error) error {
	return repo.pst.Transaction(ctx, fnc)
}

func (repo DocumentImageGroupRepository) UpdateStatusByTask(ctx context.Context, holdingCode string, taskGUID string, status int8) error {
	return repo.pst.Update(ctx, models.DocumentImageGroupDoc{}, bson.M{"holdingcode": holdingCode, "task_guid": taskGUID}, bson.M{"$set": bson.M{"status": status}})
}

func (repo DocumentImageGroupRepository) FindStatusByDocumentImageGroupTask(ctx context.Context, holdingCode string, taskGUID string) ([]models.DocumentImageGroupStatus, error) {

	filters := bson.M{
		"holdingcode": holdingCode,
		"task_guid":   taskGUID,
		"deletedat":   bson.M{"$exists": false},
	}
	docList := []models.DocumentImageGroupStatus{}
	err := repo.pst.Find(ctx, models.DocumentImageGroupDoc{}, filters, &docList)

	if err != nil {
		return []models.DocumentImageGroupStatus{}, err
	}

	return docList, nil
}

func (repo DocumentImageGroupRepository) CountByTask(ctx context.Context, holdingCode string, taskGUID string) (int, error) {

	filters := bson.M{
		"holdingcode": holdingCode,
		"task_guid":   taskGUID,
		"deletedat":   bson.M{"$exists": false},
	}

	return repo.pst.Count(ctx, models.DocumentImageGroupDoc{}, filters)
}

func (repo DocumentImageGroupRepository) CountRejectByTask(ctx context.Context, holdingCode string, taskGUID string) (int, error) {

	filters := bson.M{
		"holdingcode": holdingCode,
		"task_guid":   taskGUID,
		"$or": []interface{}{
			bson.M{"status": models.IMAGE_REJECT},
			bson.M{"status": models.IMAGE_REJECT_KEYING},
		},
		"deletedat": bson.M{"$exists": false},
	}

	return repo.pst.Count(ctx, models.DocumentImageGroupDoc{}, filters)
}

func (repo DocumentImageGroupRepository) UpdateTaskIsCompletedByTaskGUID(ctx context.Context, holdingCode string, taskGUID string, isCompleted bool) error {

	filters := bson.M{
		"holdingcode": holdingCode,
		"task_guid":   taskGUID,
		"deletedat":   bson.M{"$exists": false},
	}

	err := repo.pst.Update(ctx, models.DocumentImageGroupDoc{}, filters, bson.M{"$set": bson.M{"iscompleted": isCompleted}})

	if err != nil {
		return err
	}

	return nil
}

func (repo DocumentImageGroupRepository) UpdateXOrder(ctx context.Context, holdingCode string, taskGUID string, GUID string, xorder uint) error {

	filters := bson.M{
		"holdingcode": holdingCode,
		"task_guid":   taskGUID,
		"guidfixed":   GUID,
		"deletedat":   bson.M{"$exists": false},
	}

	err := repo.pst.UpdateOne(ctx, models.DocumentImageGroupDoc{}, filters, bson.M{"xorder": xorder})

	if err != nil {
		return err
	}

	return nil
}

func (repo DocumentImageGroupRepository) FindLastOneByTask(ctx context.Context, holdingCode string, taskGUID string) (models.DocumentImageGroupDoc, error) {

	results := []models.DocumentImageGroupDoc{}
	err := repo.pst.Aggregate(ctx, models.DocumentImageGroupDoc{}, []interface{}{
		bson.M{"$match": bson.M{
			"holdingcode": holdingCode,
			"task_guid":   taskGUID,
			"deletedat":   bson.M{"$exists": false},
		}},
		bson.M{"$sort": bson.M{"xorder": -1}},
		bson.M{"$limit": 1},
	}, &results)

	if err != nil {
		return models.DocumentImageGroupDoc{}, err
	}

	if len(results) < 1 {
		return models.DocumentImageGroupDoc{}, nil
	}

	return results[0], nil

}

func (repo DocumentImageGroupRepository) FindOneByReference(ctx context.Context, holdingCode string, reference models.Reference) (models.DocumentImageGroupDoc, error) {

	results := []models.DocumentImageGroupDoc{}
	err := repo.pst.Aggregate(ctx, models.DocumentImageGroupDoc{}, []interface{}{
		bson.M{"$match": bson.M{
			"holdingcode":       holdingCode,
			"references.module": reference.Module,
			"references.docno":  reference.DocNo,
			"deletedat":         bson.M{"$exists": false},
		}},
		bson.M{"$limit": 1},
	}, &results)

	if err != nil {
		return models.DocumentImageGroupDoc{}, err
	}

	if len(results) < 1 {
		return models.DocumentImageGroupDoc{}, nil
	}

	return results[0], nil

}

func (repo DocumentImageGroupRepository) FindOneByDocumentImageGUIDAll(ctx context.Context, documentImageGUID string) (models.DocumentImageGroupDoc, error) {

	matchQuery := bson.M{"$match": bson.M{
		"imagereferences.documentimageguid": documentImageGUID,
	}}

	results := []models.DocumentImageGroupDoc{}
	err := repo.pst.Aggregate(ctx, models.DocumentImageGroupDoc{}, []interface{}{
		matchQuery,
		bson.M{"$limit": 1},
	}, &results)

	if err != nil {
		return models.DocumentImageGroupDoc{}, err
	}

	if len(results) < 1 {
		return models.DocumentImageGroupDoc{}, nil
	}

	return results[0], nil
}

func (repo DocumentImageGroupRepository) FindOneByDocumentImageGUID(ctx context.Context, holdingCode string, documentImageGUID string) (models.DocumentImageGroupDoc, error) {

	matchQuery := bson.M{"$match": bson.M{
		"holdingcode":                       holdingCode,
		"imagereferences":                   bson.M{"$exists": true},
		"imagereferences.documentimageguid": documentImageGUID,
		"deletedat":                         bson.M{"$exists": false},
	}}

	results := []models.DocumentImageGroupDoc{}
	err := repo.pst.Aggregate(ctx, models.DocumentImageGroupDoc{}, []interface{}{
		matchQuery,
		bson.M{"$limit": 1},
	}, &results)

	if err != nil {
		return models.DocumentImageGroupDoc{}, err
	}

	if len(results) < 1 {
		return models.DocumentImageGroupDoc{}, nil
	}

	return results[0], nil
}

func (repo DocumentImageGroupRepository) FindWithoutGUIDByDocumentImageGUIDs(ctx context.Context, holdingCode string, documentImageGroupGUID string, documentImageGUIDs []string) ([]models.DocumentImageGroupInfo, error) {

	matchQuery := bson.M{"$match": bson.M{
		"holdingcode":                       holdingCode,
		"guidfixed":                         bson.M{"$ne": documentImageGroupGUID},
		"imagereferences":                   bson.M{"$exists": true},
		"imagereferences.documentimageguid": bson.M{"$in": documentImageGUIDs},
		"deletedat":                         bson.M{"$exists": false},
	}}

	results := []models.DocumentImageGroupInfo{}
	err := repo.pst.Aggregate(ctx, models.DocumentImageGroupDoc{}, []interface{}{
		matchQuery,
	}, &results)

	if err != nil {
		return []models.DocumentImageGroupInfo{}, err
	}

	return results, nil
}

func (repo DocumentImageGroupRepository) FindByDocumentImageGUIDs(ctx context.Context, holdingCode string, documentImageGUIDs []string) ([]models.DocumentImageGroupInfo, error) {

	matchQuery := bson.M{"$match": bson.M{
		"holdingcode":                       holdingCode,
		"imagereferences":                   bson.M{"$exists": true},
		"imagereferences.documentimageguid": bson.M{"$in": documentImageGUIDs},
		"deletedat":                         bson.M{"$exists": false},
	}}

	results := []models.DocumentImageGroupInfo{}
	err := repo.pst.Aggregate(ctx, models.DocumentImageGroupDoc{}, []interface{}{
		matchQuery,
	}, &results)

	if err != nil {
		return []models.DocumentImageGroupInfo{}, err
	}

	return results, nil
}

func (repo DocumentImageGroupRepository) FindByReferenceDocNo(ctx context.Context, holdingCode string, docNo string) ([]models.DocumentImageGroupDoc, error) {
	docList := []models.DocumentImageGroupDoc{}
	err := repo.pst.Find(ctx, models.DocumentImageGroupDoc{}, bson.M{
		"references.docno": docNo,
		"deletedat":        bson.M{"$exists": false},
	}, &docList)

	if err != nil {
		return []models.DocumentImageGroupDoc{}, err
	}

	return docList, nil
}

func (repo DocumentImageGroupRepository) FindByTaskGUID(ctx context.Context, holdingCode string, taskGUID string) ([]models.DocumentImageGroupDoc, error) {
	docList := []models.DocumentImageGroupDoc{}
	err := repo.pst.Find(ctx, models.DocumentImageGroupDoc{}, bson.M{
		"task_guid": taskGUID,
		"deletedat": bson.M{"$exists": false},
	}, &docList)

	if err != nil {
		return []models.DocumentImageGroupDoc{}, err
	}

	return docList, nil
}

func (repo DocumentImageGroupRepository) FindByReference(ctx context.Context, holdingCode string, reference models.Reference) ([]models.DocumentImageGroupDoc, error) {
	docList := []models.DocumentImageGroupDoc{}
	err := repo.pst.Find(ctx, models.DocumentImageGroupDoc{}, bson.M{
		"references.module": reference.Module,
		"references.docno":  reference.DocNo,
		"deletedat":         bson.M{"$exists": false},
	}, &docList)

	if err != nil {
		return []models.DocumentImageGroupDoc{}, err
	}

	return docList, nil
}

func (repo DocumentImageGroupRepository) DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string) error {
	return repo.pst.Delete(ctx, models.DocumentImageGroupDoc{}, bson.M{
		"holdingcode": holdingCode,
		"guidfixed":   guid,
	})
}

func (repo DocumentImageGroupRepository) DeleteByGUIDsIsDocumentImageEmpty(ctx context.Context, holdingCode string, GUIDs []string) error {
	return repo.pst.Delete(ctx, models.DocumentImageGroupDoc{}, bson.M{
		"holdingcode":     holdingCode,
		"imagereferences": bson.M{"$exists": true, "$size": 0},
		"$or": []interface{}{
			bson.M{"references": bson.M{"$exists": false}},
			bson.M{"references": bson.M{"$size": 0}},
			bson.M{"references": nil},
		},
	})
}

func (repo DocumentImageGroupRepository) DeleteByDocumentImageGUIDsHasOne(ctx context.Context, holdingCode string, imageGUIDs []string) error {
	return repo.pst.Delete(ctx, models.DocumentImageGroupDoc{}, bson.M{
		"holdingcode":                       holdingCode,
		"imagereferences":                   bson.M{"$exists": true, "$size": 1},
		"imagereferences.documentimageguid": bson.M{"$in": imageGUIDs},
		"$or": []interface{}{
			bson.M{"references": bson.M{"$exists": false}},
			bson.M{"references": bson.M{"$size": 0}},
		},
	})
}

func (repo DocumentImageGroupRepository) RemoveDocumentImageByDocumentImageGUIDs(ctx context.Context, holdingCode string, imageGUIDs []string) error {

	filterQuery := bson.M{
		"holdingcode":                       holdingCode,
		"imagereferences.documentimageguid": bson.M{"$in": imageGUIDs},
		"imagereferences":                   bson.M{"$exists": true},
		"$or": []interface{}{
			bson.M{"references": bson.M{"$exists": false}},
			bson.M{"references": bson.M{"$size": 0}},
			bson.M{"references": nil},
		},
	}

	removeQuery := bson.M{
		"$pull": bson.M{"imagereferences": bson.M{"document_image_guid": bson.M{"$in": imageGUIDs}}},
	}

	return repo.pst.Update(ctx, models.DocumentImageGroupDoc{}, filterQuery, removeQuery)
}

func (repo DocumentImageGroupRepository) DeleteByGUIDsIsDocumentImageEmptyWithoutDocumentImageGroupGUID(ctx context.Context, holdingCode string, withoutGUID string, GUIDs []string) error {

	filterQuery := bson.D{
		{Key: "holdingcode", Value: holdingCode},
		{Key: "guidfixed", Value: bson.M{"$ne": withoutGUID, "$in": GUIDs}},
		// {Key: "guidfixed", Value: bson.M{"$in": GUIDs}},
		{Key: "$or", Value: bson.A{
			bson.D{{Key: "imagereferences", Value: bson.D{{Key: "$exists", Value: false}}}},
			bson.D{{Key: "imagereferences", Value: bson.D{{Key: "$size", Value: 0}}}},
		}},
		{Key: "$or", Value: bson.A{
			bson.D{{Key: "references", Value: bson.D{{Key: "$exists", Value: false}}}},
			bson.D{{Key: "references", Value: bson.D{{Key: "$size", Value: 0}}}},
		}},
	}

	return repo.pst.Delete(ctx, models.DocumentImageGroupDoc{}, filterQuery)
}

func (repo DocumentImageGroupRepository) DeleteByGUIDIsDocumentImageEmpty(ctx context.Context, holdingCode string, imageGroupGUID string) error {

	filterQuery := bson.D{
		{Key: "holdingcode", Value: holdingCode},
		{Key: "guidfixed", Value: imageGroupGUID},
		{Key: "$or", Value: bson.A{
			bson.D{{Key: "imagereferences", Value: bson.D{{Key: "$exists", Value: false}}}},
			bson.D{{Key: "imagereferences", Value: bson.D{{Key: "$size", Value: 0}}}},
		}},
		{Key: "$or", Value: bson.A{
			bson.D{{Key: "references", Value: bson.D{{Key: "$exists", Value: false}}}},
			bson.D{{Key: "references", Value: bson.D{{Key: "$size", Value: 0}}}},
		}},
	}

	return repo.pst.Delete(ctx, models.DocumentImageGroupDoc{}, filterQuery)
}

func (repo DocumentImageGroupRepository) DeleteByDocumentImageGUIDsHasOneWithoutDocumentImageGroupGUID(ctx context.Context, holdingCode string, withoutGUID string, imageGUIDs []string) error {
	return repo.pst.Delete(ctx, models.DocumentImageGroupDoc{}, bson.M{
		"holdingcode":                       holdingCode,
		"guidfixed":                         bson.M{"$ne": withoutGUID},
		"imagereferences":                   bson.M{"$exists": true, "$size": 1},
		"imagereferences.documentimageguid": bson.M{"$in": imageGUIDs},
		"$or": []interface{}{
			bson.M{"references": bson.M{"$exists": false}},
			bson.M{"references": bson.M{"$size": 0}},
		},
	})
}

func (repo DocumentImageGroupRepository) RemoveDocumentImageByDocumentImageGUIDsWithoutDocumentImageGroupGUID(ctx context.Context, holdingCode string, withoutGUID string, imageGUIDs []string) error {

	filterQuery := bson.M{
		"holdingcode":                       holdingCode,
		"guidfixed":                         bson.M{"$ne": withoutGUID},
		"imagereferences.documentimageguid": bson.M{"$in": imageGUIDs},
		"imagereferences":                   bson.M{"$exists": true},
		"$or": []interface{}{
			bson.M{"references": bson.M{"$exists": false}},
			bson.M{"references": bson.M{"$size": 0}},
		},
	}

	removeQuery := bson.M{
		"$pull": bson.M{"imagereferences": bson.M{"document_image_guid": bson.M{"$in": imageGUIDs}}},
	}

	return repo.pst.Update(ctx, models.DocumentImageGroupDoc{}, filterQuery, removeQuery)
}

func (repo DocumentImageGroupRepository) FindPageImageGroup(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.DocumentImageGroupInfo, mongopagination.PaginationData, error) {

	matchFilterList := []interface{}{}

	for key, value := range filters {
		matchFilterList = append(matchFilterList, bson.M{key: value})
	}

	searchFilterList := []interface{}{}

	for _, colName := range searchInFields {
		searchFilterList = append(searchFilterList, bson.M{colName: bson.M{"$regex": primitive.Regex{
			Pattern: ".*" + pageable.Query + ".*",
			Options: "i",
		}}})
	}

	queryFilters := bson.M{
		"holdingcode": holdingCode,
		"deletedat":   bson.M{"$exists": false},
	}

	if len(searchFilterList) > 0 {
		queryFilters["$or"] = searchFilterList
	}

	if len(matchFilterList) > 0 {
		queryFilters["$and"] = matchFilterList
	}

	// check if not have sort by guidfixed, add default sort by guidfixed asc
	hasSortByGuidfixed := false
	for _, sortItem := range pageable.Sorts {
		if sortItem.Key == "guidfixed" {
			hasSortByGuidfixed = true
			break
		}
	}

	if !hasSortByGuidfixed {
		pageable.Sorts = append(pageable.Sorts, micromodels.KeyInt{Key: "guidfixed", Value: 1})
	}

	matchQuery := bson.M{
		"$match": queryFilters,
	}

	sortFields := bson.D{}
	for _, sortTemp := range pageable.Sorts {
		tempSortVal := 1
		if sortTemp.Value < 1 {
			tempSortVal = -1
		}
		sortFields = append(sortFields, bson.E{Key: sortTemp.Key, Value: tempSortVal})
	}

	sortQuery := bson.M{
		"$sort": sortFields,
	}

	aggData, err := repo.pst.AggregatePage(ctx, models.DocumentImageGroupInfo{}, pageable, matchQuery, sortQuery)

	if err != nil {
		return []models.DocumentImageGroupInfo{}, mongopagination.PaginationData{}, err
	}

	docList, err := mogoutil.AggregatePageDecode[models.DocumentImageGroupInfo](aggData)

	if err != nil {
		return []models.DocumentImageGroupInfo{}, mongopagination.PaginationData{}, err
	}

	if err != nil {
		return []models.DocumentImageGroupInfo{}, mongopagination.PaginationData{}, err
	}

	return docList, aggData.Pagination, nil
}
