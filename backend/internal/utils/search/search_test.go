package search_test

import (
	"smlcloudplatform/internal/utils/search"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestGenerateFieldFiltersMatchesAllTermsAcrossAnyField(t *testing.T) {
	filters := search.GenerateFieldFilters([]string{"username", "name"}, []string{"กิต", "001"})

	if len(filters) != 1 {
		t.Fatalf("expected one root filter, got %d", len(filters))
	}

	root, ok := filters[0].(bson.M)
	if !ok {
		t.Fatalf("expected bson.M root filter, got %T", filters[0])
	}

	andFilters, ok := root["$and"].([]interface{})
	if !ok {
		t.Fatalf("expected $and filter list, got %T", root["$and"])
	}
	if len(andFilters) != 2 {
		t.Fatalf("expected one $and entry per search term, got %d", len(andFilters))
	}

	for index, termFilter := range andFilters {
		termMap, ok := termFilter.(bson.M)
		if !ok {
			t.Fatalf("expected term filter %d to be bson.M, got %T", index, termFilter)
		}
		orFilters, ok := termMap["$or"].([]interface{})
		if !ok {
			t.Fatalf("expected term filter %d to have $or list, got %T", index, termMap["$or"])
		}
		if len(orFilters) != 2 {
			t.Fatalf("expected term filter %d to search both fields, got %d", index, len(orFilters))
		}
	}
}

func TestGenerateFieldFiltersEscapesRegexInput(t *testing.T) {
	filters := search.GenerateFieldFilters([]string{"name"}, []string{"a+b"})
	root := filters[0].(bson.M)
	andFilters := root["$and"].([]interface{})
	termMap := andFilters[0].(bson.M)
	orFilters := termMap["$or"].([]interface{})
	fieldMap := orFilters[0].(bson.M)
	regex := fieldMap["name"].(primitive.Regex)

	if regex.Pattern != `a\+b` {
		t.Fatalf("expected escaped regex pattern, got %q", regex.Pattern)
	}
	if regex.Options != "i" {
		t.Fatalf("expected case-insensitive regex option, got %q", regex.Options)
	}
}
