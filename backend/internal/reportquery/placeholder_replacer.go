package reportquery

import (
	"errors"
	"fmt"

	"go.mongodb.org/mongo-driver/bson"
)

const maxNestingDepth = 3

var ErrMaxNestingDepthExceeded = errors.New("maximum nesting depth exceeded")

type Param struct {
	Name  string
	Type  string
	Value interface{}
}

func ReplacePlaceholdersInMap(m bson.M, placeholderReplacements *map[string]interface{}) (bson.M, error) {
	return replacePlaceholdersInMap(m, placeholderReplacements, 0)
}

func replacePlaceholdersInMap(m bson.M, placeholderReplacements *map[string]interface{}, depth int) (bson.M, error) {
	if depth > maxNestingDepth {
		return nil, ErrMaxNestingDepthExceeded
	}

	updatedMap := make(bson.M)
	for key, value := range m {
		updatedValue, err := replacePlaceholders(value, placeholderReplacements, depth+1)
		if err != nil {
			return nil, err
		}
		updatedMap[key] = updatedValue
	}

	return updatedMap, nil
}

func replacePlaceholders(value interface{}, placeholderReplacements *map[string]interface{}, depth int) (interface{}, error) {
	switch v := value.(type) {
	case bson.M:
		return replacePlaceholdersInMap(v, placeholderReplacements, depth)
	case []interface{}:
		return replacePlaceholdersInArray(v, placeholderReplacements, depth)
	case string:
		if newValue, ok := (*placeholderReplacements)[v]; ok {
			return newValue, nil
		}
		return v, nil
	default:
		return value, nil
	}
}

func replacePlaceholdersInArray(arr []interface{}, placeholderReplacements *map[string]interface{}, depth int) ([]interface{}, error) {
	if depth > maxNestingDepth {
		return nil, ErrMaxNestingDepthExceeded
	}

	updatedArray := make([]interface{}, len(arr))
	for i, value := range arr {
		updatedValue, err := replacePlaceholders(value, placeholderReplacements, depth+1)
		if err != nil {
			return nil, err
		}
		updatedArray[i] = updatedValue
	}

	return updatedArray, nil
}

func XReplacePlaceholdersRecursively(value interface{}, placeholderReplacements map[string]interface{}) interface{} {
	switch v := value.(type) {
	case bson.M:
		updatedMap := make(bson.M)
		for key, val := range v {
			updatedMap[key] = XReplacePlaceholdersRecursively(val, placeholderReplacements)
		}
		return updatedMap
	case []interface{}:
		updatedArray := make([]interface{}, len(v))
		for i, val := range v {
			updatedArray[i] = XReplacePlaceholdersRecursively(val, placeholderReplacements)
		}
		return updatedArray
	case string:
		if newValue, ok := placeholderReplacements[v]; ok {
			return newValue
		}
		return v
	default:
		return value
	}
}

func main() {
	m := bson.M{
		"name": "@name@",
		"age":  "@age@",
	}
	placeholderReplacements := map[string]interface{}{
		"@name@": "xxxx",
		"@age@":  21,
	}
	updatedMap := XReplacePlaceholdersRecursively(m, placeholderReplacements)
	fmt.Printf("Original m: %v\n", m)
	fmt.Printf("Updated map: %v\n", updatedMap)
}
