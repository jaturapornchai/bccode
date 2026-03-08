package reportquery

import (
	"errors"

	"go.mongodb.org/mongo-driver/bson"
)

const maxNestingLevel = 5

func TraverseMap(m bson.M) error {
	return traverseMap(m, 0)
}

func traverseMap(m bson.M, nestingLevel int) error {
	if nestingLevel > maxNestingLevel {
		return errors.New("maximum nesting level exceeded")
	}

	for key, value := range m {
		if err := processValue(m, key, &value, nestingLevel+1); err != nil {
			return err
		}
		m[key] = value
	}

	return nil
}

func processValue(m bson.M, key string, value *interface{}, nestingLevel int) error {
	switch v := (*value).(type) {
	case bson.M:
		return traverseMap(v, nestingLevel)
	case []interface{}:
		return traverseArray(v, nestingLevel)
	default:
		// Check if value is equal to "@name@" and replace it with "xxxxx"
		if str, ok := (*value).(string); ok && str == "@name@" {
			*value = "xxxxx"
		}
		return nil
	}
}

func traverseArray(arr []interface{}, nestingLevel int) error {
	if nestingLevel > maxNestingLevel {
		return errors.New("maximum nesting level exceeded")
	}

	for i, value := range arr {
		if err := processValue(nil, "", &value, nestingLevel+1); err != nil {
			return err
		}
		arr[i] = value
	}

	return nil
}

// func traverseMap(m bson.M, nestingLevel int) error {
// 	if nestingLevel > maxNestingLevel {
// 		return errors.New("maximum nesting level exceeded")
// 	}

// 	for key, value := range m {
// 		fmt.Printf("Key: %v, Value: %v\n", key, value)
// 		if err := processValue(value, nestingLevel+1); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }

// func processValue(value interface{}, nestingLevel int) error {
// 	switch v := value.(type) {
// 	case bson.M:
// 		return traverseMap(v, nestingLevel)
// 	case []interface{}:
// 		return traverseArray(v, nestingLevel)
// 	default:
// 		// Do nothing, value has already been printed in traverseMap
// 		return nil
// 	}
// }

// func traverseArray(arr []interface{}, nestingLevel int) error {
// 	if nestingLevel > maxNestingLevel {
// 		return errors.New("maximum nesting level exceeded")
// 	}

// 	for _, value := range arr {
// 		fmt.Printf("Array Value: %v\n", value)
// 		if err := processValue(value, nestingLevel+1); err != nil {
// 			return err
// 		}
// 	}

// 	return nil
// }
