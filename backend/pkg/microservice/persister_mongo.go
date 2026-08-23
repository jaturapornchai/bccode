package microservice

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/event"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"smlcloudplatform/internal/config"
	"smlcloudplatform/pkg/microservice/models"
)

// IPersister is interface for persister
type IPersisterMongo interface {
	Aggregate(ctx context.Context, model interface{}, pipeline interface{}, decode interface{}) error
	AggregatePage(ctx context.Context, model interface{}, pageable models.Pageable, criteria ...interface{}) (*mongopagination.PaginatedData, error)
	Find(ctx context.Context, model interface{}, filter interface{}, decode interface{}, opts ...*options.FindOptions) error
	FindPage(ctx context.Context, model interface{}, filter interface{}, pageable models.Pageable, decode interface{}) (mongopagination.PaginationData, error)
	FindSelectPage(ctx context.Context, model interface{}, selectFields interface{}, filter interface{}, pageable models.Pageable, decode interface{}) (mongopagination.PaginationData, error)
	// FindPage(ctx context.Context, model interface{}, limit int, page int, filter interface{}, decode interface{}) (mongopagination.PaginationData, error)
	// FindPageSort(ctx context.Context, model interface{}, limit int, page int, filter interface{}, sorts map[string]int, decode interface{}) (mongopagination.PaginationData, error)
	FindOne(ctx context.Context, model interface{}, filter interface{}, decode interface{}, opts ...*options.FindOneOptions) error
	FindByID(ctx context.Context, model interface{}, keyName string, id interface{}, decode interface{}) error
	Create(ctx context.Context, model interface{}, data interface{}) (primitive.ObjectID, error)
	UpdateOne(ctx context.Context, model interface{}, filterConditions map[string]interface{}, data interface{}) error
	Update(ctx context.Context, model interface{}, filter interface{}, data interface{}, opts ...*options.UpdateOptions) error
	CreateInBatch(ctx context.Context, model interface{}, data []interface{}) error
	Count(ctx context.Context, model interface{}, filter interface{}) (int, error)
	Exec(ctx context.Context, model interface{}) (*mongo.Collection, error)
	Delete(ctx context.Context, model interface{}, filter interface{}) error
	DeleteByID(ctx context.Context, model interface{}, id string) error
	SoftDelete(ctx context.Context, model interface{}, username string, filter interface{}) error
	SoftDeleteLastUpdate(ctx context.Context, model interface{}, username string, filter interface{}) error
	SoftBatchDeleteByID(ctx context.Context, model interface{}, username string, ids []string) error
	SoftDeleteByID(ctx context.Context, model interface{}, id string, username string) error
	Transaction(ctx context.Context, queryFunc func(ctx context.Context) error) error
	Cleanup(ctx context.Context) error
	TestConnect(ctx context.Context) error
	Healthcheck(ctx context.Context) error
	CreateIndex(ctx context.Context, model interface{}, indexName string, keys interface{}) (string, error)
	CreatePartialUniqueIndex(ctx context.Context, model interface{}, indexName string, keys interface{}, filter interface{}) (string, error)
}

type MongoModel interface {
	CollectionName() string
}

type PersisterMongo struct {
	config  config.IPersisterMongoConfig
	db      *mongo.Database
	dbMutex sync.Mutex
	client  *mongo.Client
	// ctx       context.Context
	ctxCancel context.CancelFunc
}

func NewPersisterMongo(config config.IPersisterMongoConfig) *PersisterMongo {
	// ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// ctx := context.Background()
	// defer cancel()

	return &PersisterMongo{
		config: config,
		// ctx:       ctx,
		ctxCancel: nil,
	}
}

func NewPersisterMongoWithDBContext(db *mongo.Database) *PersisterMongo {
	// ctx := context.Background()
	return &PersisterMongo{
		db: db,
		// ctx: ctx,
	}
}

func (pst *PersisterMongo) getConnectionString() (string, error) {
	cfg := pst.config

	return cfg.MongodbURI(), nil
}

func (pst *PersisterMongo) TestConnect(ctx context.Context) error {
	_, err := pst.getClient(ctx)

	if err != nil {
		return err
	}

	return nil
}

func (pst *PersisterMongo) getClient(ctx context.Context) (*mongo.Database, error) {
	if pst.db != nil {
		return pst.db, nil
	}

	pst.dbMutex.Lock()
	defer pst.dbMutex.Unlock()

	connectionStr, err := pst.getConnectionString()

	if err != nil {
		return nil, err
	}

	if pst.config.Debug() {
		cmdMonitor := &event.CommandMonitor{
			Started: func(_ context.Context, evt *event.CommandStartedEvent) {
				log.Print(evt.Command)
			},
		}

		pst.client, err = mongo.NewClient(options.Client().ApplyURI(connectionStr).SetMonitor(cmdMonitor))
		if err != nil {
			return nil, err
		}

	} else {
		pst.client, err = mongo.NewClient(options.Client().ApplyURI(connectionStr))
		if err != nil {
			return nil, err
		}
	}

	err = pst.client.Connect(ctx)
	if err != nil {
		return nil, err
	}

	// check connection
	err = pst.client.Ping(context.TODO(), nil)
	if err != nil {
		return nil, err
	}
	// defer client.Disconnect(ctx)

	// databases, err := client.ListDatabaseNames(ctx, bson.M{})
	// if err != nil {
	// 	fmt.Println(err)
	// }
	// fmt.Println(databases)

	db := pst.client.Database(pst.config.DB())

	pst.db = db

	return db, nil
}

func (pst *PersisterMongo) getCollectionName(model interface{}) (string, error) {

	mongoModel, ok := model.(MongoModel)

	if ok {
		return mongoModel.CollectionName(), nil
	}
	return "", fmt.Errorf("struct is not implement MongoModel")
}

func (pst *PersisterMongo) toDoc(v interface{}) (doc *bson.D, err error) {
	data, err := bson.Marshal(v)
	if err != nil {
		return
	}

	err = bson.Unmarshal(data, &doc)
	return
}

func (pst *PersisterMongo) Count(ctx context.Context, model interface{}, filter interface{}) (int, error) {

	collectionName, err := pst.getCollectionName(model)
	if err != nil {
		return 0, err
	}
	return pst.PersisterCount(ctx, collectionName, filter)
}

func (pst *PersisterMongo) PersisterCount(ctx context.Context, collectionName string, filter interface{}) (int, error) {

	db, err := pst.getClient(ctx)
	if err != nil {
		return 0, err
	}

	count, err := db.Collection(collectionName).CountDocuments(ctx, filter)
	if err != nil {
		return 0, err
	}
	return int(count), nil
}

func (pst *PersisterMongo) FindPage(ctx context.Context, model interface{}, filter interface{}, pageable models.Pageable, decode interface{}) (mongopagination.PaginationData, error) {
	db, err := pst.getClient(ctx)

	emptyPage := mongopagination.PaginationData{}

	if err != nil {
		return emptyPage, err
	}

	collectionName, err := pst.getCollectionName(model)
	if err != nil {
		return emptyPage, err
	}

	var limit64 int64 = int64(pageable.Limit)
	var page64 int64 = int64(pageable.Page)

	pagingQuery := mongopagination.New(db.Collection(collectionName)).Context(ctx).Limit(limit64).Page(page64).Filter(filter)

	for _, tempSort := range pageable.Sorts {
		tempSortVal := 1
		if tempSort.Value < 1 {
			tempSortVal = -1
		}
		pagingQuery = pagingQuery.Sort(tempSort.Key, tempSortVal)
	}

	paginatedData, err := pagingQuery.Decode(decode).Find()
	if err != nil {
		return emptyPage, err
	}

	return paginatedData.Pagination, nil
}

func (pst *PersisterMongo) FindSelectPage(ctx context.Context, model interface{}, selectFields interface{}, filter interface{}, pageable models.Pageable, decode interface{}) (mongopagination.PaginationData, error) {
	db, err := pst.getClient(ctx)

	emptyPage := mongopagination.PaginationData{}

	if err != nil {
		return emptyPage, err
	}

	collectionName, err := pst.getCollectionName(model)
	if err != nil {
		return emptyPage, err
	}

	var limit64 int64 = int64(pageable.Limit)
	var page64 int64 = int64(pageable.Page)

	pagingQuery := mongopagination.New(db.Collection(collectionName)).Context(ctx).Select(selectFields).Limit(limit64).Page(page64).Filter(filter)

	for _, tempSort := range pageable.Sorts {
		tempSortVal := 1
		if tempSort.Value < 1 {
			tempSortVal = -1
		}
		pagingQuery = pagingQuery.Sort(tempSort.Key, tempSortVal)
	}

	paginatedData, err := pagingQuery.Decode(decode).Find()
	if err != nil {
		return emptyPage, err
	}

	return paginatedData.Pagination, nil
}

func (pst *PersisterMongo) Find(ctx context.Context, model interface{}, filter interface{}, decode interface{}, opts ...*options.FindOptions) error {
	db, err := pst.getClient(ctx)
	if err != nil {
		return err
	}

	collectionName, err := pst.getCollectionName(model)
	if err != nil {
		return err
	}

	filterCursor, err := db.Collection(collectionName).Find(ctx, filter, opts...)
	if err != nil {
		return err
	}

	if err = filterCursor.All(ctx, decode); err != nil {
		return err
	}

	return nil
}

func (pst *PersisterMongo) FindOne(ctx context.Context, model interface{}, filter interface{}, decode interface{}, opts ...*options.FindOneOptions) error {
	db, err := pst.getClient(ctx)
	if err != nil {
		return err
	}

	collectionName, err := pst.getCollectionName(model)
	if err != nil {
		return err
	}

	result := db.Collection(collectionName).FindOne(context.TODO(), filter, opts...)

	err = result.Decode(decode)
	if err != nil && err.Error() != "mongo: no documents in result" {
		return err
	}

	return nil
}

func (pst *PersisterMongo) FindByID(ctx context.Context, model interface{}, keyName string, id interface{}, decode interface{}) error {
	db, err := pst.getClient(ctx)
	if err != nil {
		return err
	}

	collectionName, err := pst.getCollectionName(model)
	if err != nil {
		return err
	}

	err = db.Collection(collectionName).FindOne(ctx, bson.D{{Key: keyName, Value: id}}).Decode(decode)
	if err != nil {
		return err
	}

	return nil
}

func (pst *PersisterMongo) Create(ctx context.Context, model interface{}, data interface{}) (primitive.ObjectID, error) {

	collectionName, err := pst.getCollectionName(model)
	if err != nil {
		return primitive.NilObjectID, err
	}
	return pst.PersisterCreate(ctx, collectionName, data)
}

func (pst *PersisterMongo) PersisterCreate(ctx context.Context, collectionName string, data interface{}) (primitive.ObjectID, error) {

	db, err := pst.getClient(ctx)
	if err != nil {
		return primitive.NilObjectID, err
	}

	result, err := db.Collection(collectionName).InsertOne(ctx, &data)
	if err != nil {
		return primitive.NilObjectID, err
	}

	return result.InsertedID.(primitive.ObjectID), nil
}

func (pst *PersisterMongo) CreateInBatch(ctx context.Context, model interface{}, data []interface{}) error {
	db, err := pst.getClient(ctx)
	if err != nil {
		return err
	}

	collectionName, err := pst.getCollectionName(model)
	if err != nil {
		return err
	}

	_, err = db.Collection(collectionName).InsertMany(ctx, data)

	if err != nil {
		return err
	}

	return nil
}

func (pst *PersisterMongo) Update(ctx context.Context, model interface{}, filter interface{}, data interface{}, opts ...*options.UpdateOptions) error {
	db, err := pst.getClient(ctx)
	if err != nil {
		return err
	}

	collectionName, err := pst.getCollectionName(model)
	if err != nil {
		return err
	}

	_, err = db.Collection(collectionName).UpdateMany(
		ctx,
		filter,
		data,
		opts...,
	)

	if err != nil {
		return err
	}

	return nil
}

func (pst *PersisterMongo) UpdateOne(ctx context.Context, model interface{}, filterConditions map[string]interface{}, data interface{}) error {
	db, err := pst.getClient(ctx)
	if err != nil {
		return err
	}

	collectionName, err := pst.getCollectionName(model)
	if err != nil {
		return err
	}

	updateDoc, err := pst.toDoc(data)
	if err != nil {
		return err
	}

	filterDoc := bson.M{}

	for key, val := range filterConditions {
		filterDoc[key] = val
	}

	_, err = db.Collection(collectionName).UpdateOne(
		ctx,
		filterDoc,
		bson.D{
			{Key: "$set", Value: updateDoc},
		},
	)

	if err != nil {
		return err
	}

	return nil
}

func (pst *PersisterMongo) SoftDeleteByID(ctx context.Context, model interface{}, id string, username string) error {
	db, err := pst.getClient(ctx)
	if err != nil {
		return err
	}

	collectionName, err := pst.getCollectionName(model)
	if err != nil {
		return err
	}

	deletedAt := time.Now()
	// _, err := pst.UpdateOne(model, "guidfixed", id, map[string]interface{}{"deletedat": deletedAt})

	_, err = db.Collection(collectionName).UpdateOne(
		ctx,
		bson.D{{
			Key:   "guidfixed",
			Value: id,
		}},
		bson.D{
			{Key: "$set",
				Value: bson.D{
					{Key: "deletedby", Value: username},
					{Key: "deletedat", Value: deletedAt},
				},
			},
		},
	)
	if err != nil {
		return err
	}

	return nil
}

func (pst *PersisterMongo) SoftDelete(ctx context.Context, model interface{}, username string, filter interface{}) error {

	db, err := pst.getClient(ctx)
	if err != nil {
		return err
	}

	collectionName, err := pst.getCollectionName(model)
	if err != nil {
		return err
	}

	deletedAt := time.Now()

	_, err = db.Collection(collectionName).UpdateMany(ctx, filter, bson.D{
		{Key: "$set", Value: bson.M{"deletedat": deletedAt, "deletedby": username}},
	})

	if err != nil {
		return err
	}

	return nil
}

func (pst *PersisterMongo) SoftDeleteLastUpdate(ctx context.Context, model interface{}, username string, filter interface{}) error {

	db, err := pst.getClient(ctx)
	if err != nil {
		return err
	}

	collectionName, err := pst.getCollectionName(model)
	if err != nil {
		return err
	}

	deletedAt := time.Now()

	_, err = db.Collection(collectionName).UpdateMany(ctx, filter, bson.D{
		{Key: "$set", Value: bson.M{"deletedat": deletedAt, "deletedby": username, "lastupdatedat": deletedAt}},
	})

	if err != nil {
		return err
	}

	return nil
}

func (pst *PersisterMongo) SoftBatchDeleteByID(ctx context.Context, model interface{}, username string, ids []string) error {
	db, err := pst.getClient(ctx)
	if err != nil {
		return err
	}

	collectionName, err := pst.getCollectionName(model)
	if err != nil {
		return err
	}

	objIDs := []primitive.ObjectID{}

	for _, id := range ids {
		idx, err := primitive.ObjectIDFromHex(id)
		if err != nil {
			return err
		}
		objIDs = append(objIDs, idx)
	}

	deletedAt := time.Now()

	_, err = db.Collection(collectionName).UpdateMany(ctx,
		bson.M{"_id": bson.M{"$in": objIDs}},
		bson.D{
			{Key: "$set", Value: bson.M{"deletedat": deletedAt, "deletedby": username}},
		})

	if err != nil {
		return err
	}

	return nil
}

func (pst *PersisterMongo) Delete(ctx context.Context, model interface{}, filter interface{}) error {
	db, err := pst.getClient(ctx)
	if err != nil {
		return err
	}

	collectionName, err := pst.getCollectionName(model)
	if err != nil {
		return err
	}

	_, err = db.Collection(collectionName).DeleteMany(ctx, filter)
	if err != nil {
		return err
	}

	return nil
}

func (pst *PersisterMongo) DeleteByID(ctx context.Context, model interface{}, id string) error {
	db, err := pst.getClient(ctx)
	if err != nil {
		return err
	}

	collectionName, err := pst.getCollectionName(model)
	if err != nil {
		return err
	}

	idx, _ := primitive.ObjectIDFromHex(id)
	_, err = db.Collection(collectionName).DeleteOne(ctx, bson.M{"_id": idx})
	if err != nil {
		return err
	}

	return nil
}

func (pst *PersisterMongo) Exec(ctx context.Context, model interface{}) (*mongo.Collection, error) {
	db, err := pst.getClient(ctx)
	if err != nil {
		return nil, err
	}

	collectionName, err := pst.getCollectionName(model)
	if err != nil {
		return nil, err
	}

	mongoCollection := db.Collection(collectionName)

	return mongoCollection, nil
}

func (pst *PersisterMongo) Aggregate(ctx context.Context, model interface{}, pipeline interface{}, decode interface{}) error {
	db, err := pst.getClient(ctx)
	if err != nil {
		return err
	}

	collectionName, err := pst.getCollectionName(model)
	if err != nil {
		return err
	}

	// var aggregationFilter []bson.M
	// for _, filter := range filters {
	// 	aggregationFilter = append(aggregationFilter, filter.(bson.M))
	// }

	filterCursor, err := db.Collection(collectionName).Aggregate(ctx, pipeline)

	if err != nil {
		return err
	}

	if err = filterCursor.All(ctx, decode); err != nil {
		return err
	}

	return nil
}

func (pst *PersisterMongo) AggregatePage(ctx context.Context, model interface{}, pageable models.Pageable, criteria ...interface{}) (*mongopagination.PaginatedData, error) {
	db, err := pst.getClient(ctx)

	emptyPage := &mongopagination.PaginatedData{}

	if err != nil {
		return emptyPage, err
	}

	collectionName, err := pst.getCollectionName(model)
	if err != nil {
		return emptyPage, err
	}

	var page64 int64 = int64(pageable.Page)
	var limit64 int64 = int64(pageable.Limit)

	paginatedData, err := mongopagination.New(db.Collection(collectionName)).Context(ctx).Limit(limit64).Page(page64).Aggregate(criteria...)
	if err != nil {
		return emptyPage, err
	}

	return paginatedData, nil
}

func (pst *PersisterMongo) Cleanup(ctx context.Context) error {
	err := pst.client.Disconnect(ctx)
	if err != nil {
		return err
	}

	if pst != nil && pst.ctxCancel != nil {
		pst.ctxCancel()
	}

	return nil
}

func (pst *PersisterMongo) Healthcheck(ctx context.Context) error {
	retry := 5
	// We will try to getClient 5 times
	for {
		if retry <= 0 {
			return fmt.Errorf("mongodb healthcheck failed")
		}
		retry--

		_, err := pst.getClient(ctx)
		if err != nil {
			// Healthcheck failed, wait 250ms then try again
			time.Sleep(250 * time.Millisecond)
			continue
		}
		return nil
	}
}

func (pst *PersisterMongo) Transaction(ctx context.Context, queryFunc func(context.Context) error) error {
	db, err := pst.getClient(ctx)
	if err != nil {
		return err
	}
	client := db.Client()
	if client == nil {
		return fmt.Errorf("mongodb client is not available")
	}

	err = client.UseSession(ctx, func(sessionContext mongo.SessionContext) error {
		err := sessionContext.StartTransaction()
		if err != nil {
			return err
		}

		err = queryFunc(sessionContext)

		if err != nil {
			if err := sessionContext.AbortTransaction(sessionContext); err != nil {
				return err
			}

			return err
		}

		err = sessionContext.CommitTransaction(sessionContext)
		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (pst *PersisterMongo) CreateIndex(ctx context.Context, model interface{}, indexName string, keys interface{}) (string, error) {
	db, err := pst.getClient(ctx)
	if err != nil {
		return "", err
	}

	collectionName, err := pst.getCollectionName(model)
	if err != nil {
		return "", err
	}

	indexModel := mongo.IndexModel{
		Keys:    keys,
		Options: options.Index().SetUnique(true).SetName(indexName),
	}

	mongoCollection := db.Collection(collectionName)

	resultIndexName, err := mongoCollection.Indexes().CreateOne(ctx, indexModel)

	if err != nil {
		return "", err
	}

	return resultIndexName, nil
}

func (pst *PersisterMongo) CreatePartialUniqueIndex(
	ctx context.Context,
	model interface{},
	indexName string,
	keys interface{},
	filter interface{},
) (string, error) {
	db, err := pst.getClient(ctx)
	if err != nil {
		return "", err
	}

	collectionName, err := pst.getCollectionName(model)
	if err != nil {
		return "", err
	}

	return db.Collection(collectionName).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: keys,
		Options: options.Index().
			SetUnique(true).
			SetName(indexName).
			SetPartialFilterExpression(filter),
	})
}
