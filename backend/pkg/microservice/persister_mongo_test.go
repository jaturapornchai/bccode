package microservice_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"reflect"
	"smlcloudplatform/mock"
	"smlcloudplatform/pkg/microservice"
	"smlcloudplatform/pkg/microservice/models"
	"testing"

	"github.com/tj/assert"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/bsoncodec"
	"go.mongodb.org/mongo-driver/bson/bsonrw"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

// type ConfigDBTest struct{}

// func (c *ConfigDBTest) MongodbURI() string {
// 	return "mongodb://192.168.2.209:27017/"
// }

// func (c *ConfigDBTest) DB() string {
// 	return "micro_test"
// }

type Product struct {
	ID          primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ProductCode string             `json:"product_code" bson:"product_code"`
	ProductName string             `json:"product_name" bson:"product_name"`
}

func (pdt *Product) CollectionName() string {
	return "product"
}

var (
	productCode  = "pdt-01"
	productName  = "product name 01"
	noClientOpts = mtest.NewOptions().CreateClient(false)
)

type negateCodec struct {
	ID int64 `bson:"_id"`
}

func (e *negateCodec) EncodeValue(ectx bsoncodec.EncodeContext, vw bsonrw.ValueWriter, val reflect.Value) error {
	return vw.WriteInt64(val.Int())
}

// DecodeValue negates the value of ID when reading
func (e *negateCodec) DecodeValue(ectx bsoncodec.DecodeContext, vr bsonrw.ValueReader, val reflect.Value) error {
	i, err := vr.ReadInt64()
	if err != nil {
		return err
	}

	val.SetInt(i * -1)
	return nil
}

// init

func initCollection(mt *mtest.T, coll *mongo.Collection) {
	mt.Helper()

	var docs []interface{}
	for i := 1; i <= 5; i++ {
		docs = append(docs, bson.D{{"product_code", int32(i)}})
	}

	_, err := coll.InsertMany(context.Background(), docs)
	assert.Nil(mt, err, "InsertMany error for initial data: %v", err)
}

func TestMongoPersisterCountWithmtest(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	mt.Run("Success", func(mt *mtest.T) {

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"n", 1},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)

		mt.AddMockResponses(first, killCursors)

		pst := microservice.NewPersisterMongoWithDBContext(mt.DB)
		count, err := pst.Count(context.TODO(), &Product{}, bson.M{"product_code": "0001"})
		if err != nil {
			t.Error(err.Error())
			return
		}

		fmt.Printf("Count :: %d \n", count)

		if count < 0 {
			t.Error("can't get Count ")
			return
		}
	})
}

func TestMongodbCreate(t *testing.T) {

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	mt.Run("Success", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		pst := microservice.NewPersisterMongoWithDBContext(mt.DB)
		objID, err := pst.Create(context.TODO(), &Product{}, &Product{
			ProductCode: productCode,
			ProductName: productName,
		})

		t.Log(objID)

		if err != nil {
			t.Error(err.Error())
		}
	})
}

func TestMongodbCreateInBatchWithMtest(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	mt.Run("Success", func(mt *mtest.T) {
		mt.AddMockResponses(mtest.CreateSuccessResponse())

		pst := microservice.NewPersisterMongoWithDBContext(mt.DB)
		products := make([]interface{}, 0)

		for i := 1; i <= 25; i++ {
			products = append(products, Product{
				ProductCode: fmt.Sprintf("pdt-%02d", i),
				ProductName: fmt.Sprintf("pdt name %d", i),
			})
		}

		err := pst.CreateInBatch(context.TODO(), &Product{}, products)

		if err != nil {
			t.Error(err.Error())
		}
	})
}

func TestMongoDBFindMockData(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()
	mt.Run("Success", func(mt *mtest.T) {

		id1 := primitive.NewObjectID()
		id2 := primitive.NewObjectID()

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"_id", id1},
			{"product_code", "0001"},
			{"product_name", "name 0001"},
		})

		second := mtest.CreateCursorResponse(1, "foo.bar", mtest.NextBatch, bson.D{
			{"_id", id2},
			{"product_code", "0002"},
			{"product_name", "name 0002"},
		})

		killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)

		mt.AddMockResponses(first, second, killCursors)

		pst := microservice.NewPersisterMongoWithDBContext(mt.DB)
		products := []Product{}
		err := pst.Find(context.TODO(), &Product{}, bson.M{}, &products)

		if err != nil {
			t.Error(err.Error())
			return
		}

		if len(products) < 1 {
			t.Error("Find not found item")
		}

		t.Log("product all :: ", len(products))
	})
}

func TestMongodbFindPage(t *testing.T) {

	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	defer mt.Close()

	// // registryOpts := options.Client().
	// // 	SetRegistry(bson.NewRegistryBuilder().RegisterCodec(reflect.TypeOf(int64(0)), &negateCodec{}).Build())

	// mt.RunOpts("registry passed to cursors", mtest.NewOptions().CreateClient(false), func(mt *mtest.T) {
	// 	_, err := mt.Coll.InsertOne(context.Background(), negateCodec{ID: 10})
	// 	assert.Nil(mt, err, "InsertOne error: %v", err)
	// 	var got negateCodec
	// 	err = mt.Coll.FindOne(context.Background(), bson.D{}).Decode(&got)
	// 	assert.Nil(mt, err, "Find error: %v", err)

	// 	assert.Equal(mt, int64(-10), got.ID, "expected ID -10, got %v", got.ID)
	// })

	// mt.RunOpts("try find", noClientOpts, func(mt *mtest.T) {
	if os.Getenv("SERVERLESS") == "serverless" {
		mt.Skip("skipping as serverless forbids capped collections")
	}

	mt.Run("Success", func(mt *mtest.T) {

		//initCollection(mt, mt.Coll)
		id1 := primitive.NewObjectID()
		id2 := primitive.NewObjectID()

		first := mtest.CreateCursorResponse(1, "foo.bar", mtest.FirstBatch, bson.D{
			{"_id", id1},
			{"product_code", "0001"},
			{"product_name", "name 0001"},
		})

		second := mtest.CreateCursorResponse(1, "foo.bar", mtest.NextBatch, bson.D{
			{"_id", id2},
			{"product_code", "0002"},
			{"product_name", "name 0002"},
		})

		//killCursors := mtest.CreateCursorResponse(0, "foo.bar", mtest.NextBatch)

		mt.AddMockResponses(first, second)
		pst := microservice.NewPersisterMongoWithDBContext(mt.DB)

		products := []Product{}
		pagination, err := pst.FindPage(context.TODO(), &Product{}, bson.M{}, models.Pageable{Page: 1, Limit: 10}, &products)

		if err != nil {
			t.Error(err.Error())
			return
		}

		fmt.Println(pagination)
		fmt.Println(products)

		if len(products) < 1 {
			t.Error("Find not found item")
		}
	})
	// })
}

type Model1 struct {
	Name string `bson:"name"`
}

func (m Model1) CollectionName() string {
	return "model1"
}

type Model2 struct {
	Name string `bson:"name"`
}

// func (m Model2) CollectionName() string {
// 	return "model2"
// }

func TestMongoTransaction(t *testing.T) {

	pst := microservice.NewPersisterMongo(mock.NewPersisterMongoConfig())

	ctx := context.Background()
	err := pst.Transaction(ctx, func(sc context.Context) error {

		idx, err := pst.Create(sc, &Model1{}, &Model1{Name: "xxx"})

		if err != nil {
			return err
		}

		fmt.Println(idx)

		return errors.New("test")

		// err = pst.Update(ctx, &Model1{}, bson.M{"name": "name3"}, bson.M{"$set": bson.M{"name": "name1"}})

		// if err != nil {
		// 	return err
		// }

		return nil
	})

	t.Log(err)
}
