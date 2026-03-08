package reportquery_test

import (
	"reflect"
	"smlcloudplatform/internal/reportquery"
	"testing"

	"go.mongodb.org/mongo-driver/bson"
)

func TestReplacePlaceholdersInMap(t *testing.T) {
	tests := []struct {
		name                    string
		input                   bson.M
		placeholderReplacements map[string]interface{}
		expected                bson.M
	}{
		{
			name: "Test simple replacements",
			input: bson.M{
				"name": "@name@",
				"age":  "@age@",
			},
			placeholderReplacements: map[string]interface{}{
				"@name@": "John",
				"@age@":  30,
			},
			expected: bson.M{
				"name": "John",
				"age":  30,
			},
		},
		{
			name: "Test nested replacements",
			input: bson.M{
				"user": bson.M{
					"name": "@name@",
					"age":  "@age@",
				},
			},
			placeholderReplacements: map[string]interface{}{
				"@name@": "Jane",
				"@age@":  25,
			},
			expected: bson.M{
				"user": bson.M{
					"name": "Jane",
					"age":  25,
				},
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			updatedMap, err := reportquery.ReplacePlaceholdersInMap(test.input, &test.placeholderReplacements)
			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
			if !reflect.DeepEqual(updatedMap, test.expected) {
				t.Errorf("Expected: %v, got: %v", test.expected, updatedMap)
			}
		})
	}
}
