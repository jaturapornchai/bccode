package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/task/models"
	"smlcloudplatform/internal/utils/mogoutil"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type ITaskRepository interface {
	FindTaskChild(ctx context.Context, holdingCode string, rejectFromTaskGUID string) (models.TaskChild, error)
	FindLastTaskByCode(ctx context.Context, holdingCode string, codeFormat string) (models.TaskDoc, error)
	FindOneTaskByCode(ctx context.Context, holdingCode string, taskCode string) (models.TaskInfo, error)
	Count(ctx context.Context, holdingCode string) (int, error)
	CountTaskParent(ctx context.Context, holdingCode string, taskGUID string) (int, error)
	Create(ctx context.Context, doc models.TaskDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.TaskDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.TaskDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	Delete(ctx context.Context, holdingCode string, username string, filters map[string]interface{}) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.TaskInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.TaskDoc, error)

	UpdateTotalDocumentImageGroup(ctx context.Context, holdingCode string, taskGUID string, total int, countStatus []models.TotalStatus, billCount float64, referenceCount float64, referenceBalance float64) error
	UpdateTotalRejectDocumentImageGroup(ctx context.Context, holdingCode string, taskGUID string, total int) error

	FindPageByTaskReject(ctx context.Context, holdingCode string, module string, taskGUID string) ([]models.TaskInfo, error)
	FindInItemGuid(ctx context.Context, holdingCode string, columnName string, itemGuidList []string) ([]models.TaskItemGuid, error)
	FindByDocIndentityGuid(ctx context.Context, holdingCode string, indentityField string, indentityValue interface{}) (models.TaskDoc, error)
	FindStep(ctx context.Context, holdingCode string, filters map[string]interface{}, searchInFields []string, projects map[string]interface{}, pageableLimit micromodels.PageableStep) ([]models.TaskInfo, int, error)

	FindDeletedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.TaskDeleteActivity, mongopagination.PaginationData, error)
	FindCreatedOrUpdatedPage(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.TaskActivity, mongopagination.PaginationData, error)
	FindDeletedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.TaskDeleteActivity, error)
	FindCreatedOrUpdatedStep(ctx context.Context, holdingCode string, lastUpdatedDate time.Time, extraFilters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.TaskActivity, error)

	FindPageTask(ctx context.Context, holdingCode string, module string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.TaskInfo, mongopagination.PaginationData, error)
}

type TaskRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.TaskDoc]
	repositories.SearchRepository[models.TaskInfo]
	repositories.GuidRepository[models.TaskItemGuid]
	repositories.ActivityRepository[models.TaskActivity, models.TaskDeleteActivity]
}

func NewTaskRepository(pst microservice.IPersisterMongo) *TaskRepository {

	insRepo := TaskRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.TaskDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.TaskInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.TaskItemGuid](pst)
	insRepo.ActivityRepository = repositories.NewActivityRepository[models.TaskActivity, models.TaskDeleteActivity](pst)

	return &insRepo
}

func (repo *TaskRepository) FindTaskChild(ctx context.Context, holdingCode string, rejectFromTaskGUID string) (models.TaskChild, error) {

	queryFilters := bson.M{
		"holdingcode":        holdingCode,
		"deletedat":          bson.M{"$exists": false},
		"rejectfromtaskguid": rejectFromTaskGUID,
	}

	opts := &options.FindOneOptions{}

	opts.SetSort(bson.M{"code": 1})

	findDoc := models.TaskChild{}
	err := repo.pst.FindOne(ctx, models.TaskChild{}, queryFilters, &findDoc, opts)

	if err != nil {
		return models.TaskChild{}, err
	}

	return findDoc, nil
}

func (repo *TaskRepository) FindLastTaskByCode(ctx context.Context, holdingCode string, codeFormat string) (models.TaskDoc, error) {

	queryFilters := bson.M{
		"holdingcode": holdingCode,
		"deletedat":   bson.M{"$exists": false},
		"code": bson.M{"$regex": primitive.Regex{
			Pattern: codeFormat + ".*",
			Options: "i",
		}},
	}

	opts := &options.FindOneOptions{}

	opts.SetSort(bson.M{"code": -1})

	findDoc := new(models.TaskDoc)
	err := repo.pst.FindOne(ctx, models.TaskDoc{}, queryFilters, &findDoc, opts)
	// err := repo.pst.FindOne(models.TaskDoc{}, bson.M{"holdingcode": holdingCode}, &findDoc)

	if err != nil {
		return models.TaskDoc{}, err
	}

	return *findDoc, nil
}

func (repo *TaskRepository) CountTaskParent(ctx context.Context, holdingCode string, taskGUID string) (int, error) {

	queryFilters := bson.M{
		"holdingcode": holdingCode,
		"deletedat":   bson.M{"$exists": false},
	}

	queryFilters["parentguidfixed"] = taskGUID

	count, err := repo.pst.Count(ctx, models.TaskInfo{}, queryFilters)

	if err != nil {
		return 0, err
	}

	return count, nil
}

func (repo *TaskRepository) UpdateTotalDocumentImageGroup(ctx context.Context, holdingCode string, taskGUID string, totalDoc int, totalDocStatus []models.TotalStatus, billCount float64, referenceCount float64, referenceBalance float64) error {

	queryFilters := bson.M{
		"holdingcode": holdingCode,
		"guidfixed":   taskGUID,
		"deletedat":   bson.M{"$exists": false},
	}

	queryFilters["guidfixed"] = taskGUID

	err := repo.pst.UpdateOne(ctx, models.TaskDocumentTotal{}, queryFilters, models.TaskDocumentTotal{TotalDocument: totalDoc, TotalDocumentStatus: &totalDocStatus, BillCount: billCount, ReferenceCount: referenceCount, ReferenceBalance: referenceBalance})

	if err != nil {
		return err
	}

	return nil
}

func (repo *TaskRepository) UpdateTotalRejectDocumentImageGroup(ctx context.Context, holdingCode string, taskGUID string, total int) error {

	queryFilters := bson.M{
		"holdingcode": holdingCode,
		"guidfixed":   taskGUID,
		"deletedat":   bson.M{"$exists": false},
	}

	queryFilters["guidfixed"] = taskGUID

	err := repo.pst.UpdateOne(ctx, models.TaskTotalReject{}, queryFilters, models.TaskTotalReject{ToTalReject: total})

	if err != nil {
		return err
	}

	return nil
}

func (repo *TaskRepository) FindPageByTaskReject(ctx context.Context, holdingCode string, module string, taskGUID string) ([]models.TaskInfo, error) {

	queryFilters := bson.M{
		"holdingcode": holdingCode,
		"deletedat":   bson.M{"$exists": false},
	}

	if len(module) > 0 {
		queryFilters["module"] = module
	}

	queryFilters["parentguidfixed"] = taskGUID

	docList := []models.TaskInfo{}

	err := repo.pst.Find(ctx, models.TaskInfo{}, queryFilters, &docList)

	if err != nil {
		return []models.TaskInfo{}, err
	}

	return docList, nil
}

func (repo *TaskRepository) FindOneTaskByCode(ctx context.Context, holdingCode string, taskCode string) (models.TaskInfo, error) {

	queryFilters := bson.M{
		"holdingcode": holdingCode,
		"deletedat":   bson.M{"$exists": false},
		"code":        taskCode,
	}

	findDoc := models.TaskInfo{}

	err := repo.pst.FindOne(ctx, models.TaskInfo{}, queryFilters, &findDoc)

	if err != nil {
		return models.TaskInfo{}, err
	}

	return findDoc, nil
}

func (repo *TaskRepository) FindPageTask(ctx context.Context, holdingCode string, module string, filters map[string]interface{}, searchInFields []string, pageable micromodels.Pageable) ([]models.TaskInfo, mongopagination.PaginationData, error) {

	matchFilterList := []interface{}{}

	for key, value := range filters {
		matchFilterList = append(matchFilterList, bson.M{key: value})
	}

	searchFilterList := []interface{}{}

	for _, colName := range searchInFields {
		searchFilterList = append(searchFilterList, bson.M{colName: bson.M{"$regex": primitive.Regex{
			Pattern: ".*" + pageable.Query + ".*",
			Options: "",
		}}})
	}

	queryFilters := bson.M{
		"holdingcode": holdingCode,
		"deletedat":   bson.M{"$exists": false},
	}

	if len(module) > 0 {
		queryFilters["module"] = module
	}

	if len(searchFilterList) > 0 {
		queryFilters["$or"] = searchFilterList
	}

	if len(matchFilterList) > 0 {
		queryFilters["$and"] = matchFilterList
	}

	pageable.Sorts = append(pageable.Sorts, micromodels.KeyInt{Key: "guidfixed", Value: 1})

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

	aggData, err := repo.pst.AggregatePage(ctx, models.TaskInfo{}, pageable, matchQuery, sortQuery)

	if err != nil {
		return []models.TaskInfo{}, mongopagination.PaginationData{}, err
	}

	docList, err := mogoutil.AggregatePageDecode[models.TaskInfo](aggData)

	if err != nil {
		return []models.TaskInfo{}, mongopagination.PaginationData{}, err
	}

	if err != nil {
		return []models.TaskInfo{}, mongopagination.PaginationData{}, err
	}

	if docList == nil {
		docList = []models.TaskInfo{}
	}

	return docList, aggData.Pagination, nil
}
