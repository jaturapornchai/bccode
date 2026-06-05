package repositories

import (
	"context"
	"smlcloudplatform/internal/repositories"
	"smlcloudplatform/internal/vfgl/journal/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type IJournalRepository interface {
	FindAll(ctx context.Context) ([]models.JournalDoc, error)
	Count(ctx context.Context, holdingCode string) (int, error)
	Create(ctx context.Context, doc models.JournalDoc) (string, error)
	CreateInBatch(ctx context.Context, docList []models.JournalDoc) error
	Update(ctx context.Context, holdingCode string, guid string, doc models.JournalDoc) error
	DeleteByGuidfixed(ctx context.Context, holdingCode string, guid string, username string) error
	FindPage(ctx context.Context, holdingCode string, searchInFields []string, pageable micromodels.Pageable) ([]models.JournalInfo, mongopagination.PaginationData, error)
	FindByGuid(ctx context.Context, holdingCode string, guid string) (models.JournalDoc, error)
	FindOne(ctx context.Context, holdingCode string, filters interface{}) (models.JournalDoc, error)
	FindFilter(ctx context.Context, holdingCode string, filters map[string]interface{}) ([]models.JournalDoc, error)
	IsAccountCodeUsed(ctx context.Context, holdingCode string, accountCode string) (bool, error)
	FindGUIDEmptyAll() []models.JournalDoc
	UpdateGuidEmpty(ctx context.Context, id string, guidfixed string) error
	GetDuplicateDocNos(ctx context.Context, holdingCode string) ([]models.DuplicateDocNo, error)
	CheckVatDocNoExists(ctx context.Context, holdingCode string, debtType int, code string, vatDocNo string) (bool, error)
	CheckTaxDocNoExists(ctx context.Context, holdingCode string, debtType int, code string, taxDocNo string) (bool, error)
	// FindLastDocno(holdingCode string, docFormat string) (string, error)
}

type JournalRepository struct {
	pst microservice.IPersisterMongo
	repositories.CrudRepository[models.JournalDoc]
	repositories.SearchRepository[models.JournalInfo]
	repositories.GuidRepository[models.JournalItemGuid]
}

func NewJournalRepository(pst microservice.IPersisterMongo) JournalRepository {

	insRepo := JournalRepository{
		pst: pst,
	}

	insRepo.CrudRepository = repositories.NewCrudRepository[models.JournalDoc](pst)
	insRepo.SearchRepository = repositories.NewSearchRepository[models.JournalInfo](pst)
	insRepo.GuidRepository = repositories.NewGuidRepository[models.JournalItemGuid](pst)

	return insRepo
}

func (repo *JournalRepository) IsAccountCodeUsed(ctx context.Context, holdingCode string, accountCode string) (bool, error) {

	findDoc := models.JournalDoc{}

	filters := bson.M{
		"holdingcode":               holdingCode,
		"journaldetail.accountcode": accountCode,
		"deletedat":                 bson.M{"$exists": false},
	}

	err := repo.pst.FindOne(ctx, models.JournalDoc{}, filters, &findDoc)

	if err != nil {
		return true, nil
	}

	return findDoc.ID != primitive.NilObjectID, nil

}

func (repo *JournalRepository) FindLastDocno(ctx context.Context, holdingCode string, docFormat string) (string, error) {

	findDocList := []models.JournalDoc{}

	filters := bson.M{
		"holdingcode": holdingCode,
		"deletedat":   bson.M{"$exists": false},
	}

	if len(docFormat) < 1 {
		filters["$or"] = []interface{}{
			bson.M{"docformat": ""},
			bson.M{"docformat": bson.M{"$exists": false}},
		}
	} else {
		filters["docformat"] = docFormat
	}

	findOptions := options.Find()

	findOptions.SetSort(bson.M{"docformat": -1})
	findOptions.SetLimit(1)

	err := repo.pst.Find(ctx, models.JournalDoc{}, filters, &findDocList, findOptions)

	if err != nil {
		return "", nil
	}

	if len(findDocList) < 1 {
		return "", nil
	}

	return findDocList[0].DocNo, nil

}

func (repo *JournalRepository) FindGUIDEmptyAll() ([]models.JournalDoc, error) {
	findDocList := []models.JournalDoc{}

	filters := bson.M{
		"guidfixed": "",
	}

	err := repo.pst.Find(context.Background(), models.JournalDoc{}, filters, &findDocList)

	if err != nil {
		return []models.JournalDoc{}, nil
	}

	return findDocList, nil
}

func (repo *JournalRepository) UpdateGuidEmpty(ctx context.Context, id primitive.ObjectID, guidfixed string) error {

	err := repo.pst.UpdateOne(ctx, models.JournalDoc{}, bson.M{"_id": id}, bson.M{"guidfixed": guidfixed})

	if err != nil {
		return err
	}

	return nil
}

// GetDuplicateDocNos returns all DocNos that appear more than once in the system for a given shop
func (repo *JournalRepository) GetDuplicateDocNos(ctx context.Context, holdingCode string) ([]models.DuplicateDocNo, error) {

	pipeline := []interface{}{
		// Match documents for this shop that are not deleted
		bson.M{"$match": bson.M{
			"holdingcode": holdingCode,
			"deletedat":   bson.M{"$exists": false},
		}},
		// Group by docno and count occurrences
		bson.M{"$group": bson.M{
			"_id":   "$docno",
			"count": bson.M{"$sum": 1},
		}},
		// Filter only duplicates (count > 1)
		bson.M{"$match": bson.M{
			"count": bson.M{"$gt": 1},
		}},
		// Project to match our model structure
		bson.M{"$project": bson.M{
			"docno": "$_id",
			"count": 1,
			"_id":   0,
		}},
		// Sort by count descending, then by docno
		bson.M{"$sort": bson.M{
			"count": -1,
			"docno": 1,
		}},
	}

	results := []models.DuplicateDocNo{}
	err := repo.pst.Aggregate(ctx, &models.DuplicateDocNo{}, pipeline, &results)

	if err != nil {
		return []models.DuplicateDocNo{}, err
	}

	return results, nil
}

// CheckVatDocNoExists checks if a VAT document number already exists for a specific creditor/debtor
func (repo *JournalRepository) CheckVatDocNoExists(ctx context.Context, holdingCode string, debtType int, code string, vatDocNo string) (bool, error) {
	codeField := "debtor.code"
	if debtType == 1 {
		codeField = "creditor.code"
	}

	filter := bson.M{
		"holdingcode": holdingCode,
		"deletedat":   bson.M{"$exists": false},
		codeField:     code,
		"vats": bson.M{
			"$elemMatch": bson.M{
				"vatdocno": vatDocNo,
			},
		},
	}

	count, err := repo.pst.Count(ctx, models.JournalInfo{}, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// CheckTaxDocNoExists checks if a Tax document number already exists for a specific creditor/debtor
func (repo *JournalRepository) CheckTaxDocNoExists(ctx context.Context, holdingCode string, debtType int, code string, taxDocNo string) (bool, error) {
	codeField := "debtor.code"
	if debtType == 1 {
		codeField = "creditor.code"
	}

	filter := bson.M{
		"holdingcode": holdingCode,
		"deletedat":   bson.M{"$exists": false},
		codeField:     code,
		"taxes": bson.M{
			"$elemMatch": bson.M{
				"taxdocno": taxDocNo,
			},
		},
	}

	count, err := repo.pst.Count(ctx, models.JournalInfo{}, filter)
	if err != nil {
		return false, err
	}

	return count > 0, nil
}
