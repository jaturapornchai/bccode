package reportquery_test

import (
	"fmt"
	"smlcloudplatform/internal/reportquery"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestTraverse(t *testing.T) {
	bsonMap := bson.M{
		"name": "@name@",
		// "name":  "John Doe",
		// "age":   30,
		// "email": "johndoe@example.com",
		// "address": bson.M{
		// 	"street": "123 Main St",
		// 	"city":   "New York",
		// 	"state":  "NY",
		// 	"country": bson.M{
		// 		"name": "United States",
		// 		"code": "US",
		// 	},
		// },
		// "friends": []interface{}{
		// 	bson.M{"name": "Alice", "age": 28},
		// 	bson.M{"name": "Bob", "age": 32},
		// },
	}

	fmt.Printf("bsonMap: %v\n", bsonMap)
	err := reportquery.TraverseMap(bsonMap)
	fmt.Printf("bsonMap: %v\n", bsonMap)

	assert.Nil(t, err)

}
