package repositories

import (
	"reflect"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestVerifiedEmailUserFilterUsesTrimmedCaseInsensitiveExactMatch(t *testing.T) {
	want := bson.M{
		"isdeleted": bson.M{"$ne": true},
		"$expr": bson.M{"$eq": bson.A{
			bson.M{"$toLower": bson.M{"$trim": bson.M{
				"input": bson.M{"$ifNull": bson.A{"$email", ""}},
			}}},
			"user+tag@example.com",
		}},
	}

	got := verifiedEmailUserFilter("  User+Tag@Example.COM ")
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("filter = %#v, want %#v", got, want)
	}
}
