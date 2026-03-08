package requestfilter

import (
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

// DateFormate represents the layout of date string.
const DateFormate = "2006-01-02"

// FilterType represents the type of filter.
type FilterType string

const (
	FieldTypeString    = "string"
	FieldTypeInt       = "int"
	FieldTypeFloat64   = "float64"
	FieldTypeBoolean   = "boolean"
	FieldTypeDate      = "date"
	FieldTypeRangeDate = "rangeDate"
)

// FilterRequest is a struct that contains the parameters of a request filter.
type FilterRequest struct {
	Param string
	Field string
	Type  FilterType
}

// DataTypeHandler is a function type that handles the conversion of filter parameters based on their type.
type DataTypeHandler func(field FilterRequest, filterValue string, getParamFunc func(string) string) interface{}

// getStringHandler handles string type filter requests.
func getStringHandler(field FilterRequest, filterValue string, getParamFunc func(string) string) interface{} {
	values := strings.Split(filterValue, ",")
	if len(values) == 1 {
		return filterValue
	}
	return bson.M{"$in": values}
}

// getIntHandler handles integer type filter requests.
func getIntHandler(field FilterRequest, filterValue string, getParamFunc func(string) string) interface{} {
	values := strings.Split(filterValue, ",")
	if len(values) == 1 {
		val, err := strconv.Atoi(filterValue)
		if err == nil {
			return val
		}
		return nil
	}
	intValues := parseToInts(values)
	return bson.M{"$in": intValues}
}

// getFloatHandler handles float type filter requests.
func getFloatHandler(field FilterRequest, filterValue string, getParamFunc func(string) string) interface{} {
	values := strings.Split(filterValue, ",")
	if len(values) == 1 {
		val, err := strconv.ParseFloat(filterValue, 64)
		if err == nil {
			return val
		}
		return nil
	}
	floatValues := parseToFloats(values)
	return bson.M{"$in": floatValues}
}

// getBooleanHandler handles boolean type filter requests.
func getBooleanHandler(field FilterRequest, filterValue string, getParamFunc func(string) string) interface{} {
	values := strings.Split(filterValue, ",")
	if len(values) == 1 {
		val, err := strconv.ParseBool(filterValue)
		if err == nil {
			return val
		}
		return nil
	}
	boolValues := parseToBools(values)
	return bson.M{"$in": boolValues}
}

// getRangeDateHandler handles date range filter requests.
func getRangeDateHandler(field FilterRequest, filterValue string, getParamFunc func(string) string) interface{} {
	fromDateStr := strings.TrimSpace(getParamFunc("fromdate"))
	toDateStr := strings.TrimSpace(getParamFunc("todate"))

	if len(fromDateStr) > 0 && len(toDateStr) > 0 {
		fromDate, err1 := time.Parse(DateFormate, fromDateStr)
		toDate, err2 := time.Parse(DateFormate, toDateStr)

		if err1 == nil && err2 == nil {
			return bson.M{
				"$gte": fromDate,
				"$lt":  toDate.AddDate(0, 0, 1),
			}
		}
	}
	return nil
}

func getDateHandler(field FilterRequest, filterValue string, getParamFunc func(string) string) interface{} {

	if len(filterValue) > 0 {
		fromDate, err1 := time.Parse(DateFormate, filterValue)
		toDate, err2 := time.Parse(DateFormate, filterValue)

		if err1 == nil && err2 == nil {
			return bson.M{
				"$gte": fromDate,
				"$lt":  toDate.AddDate(0, 0, 1),
			}
		}
	}
	return nil
}

// parseToInts converts an array of strings into an array of integers.
func parseToInts(values []string) []int {
	ints := make([]int, 0, len(values))
	for _, val := range values {
		if intVal, err := strconv.Atoi(val); err == nil {
			ints = append(ints, intVal)
		}
	}
	return ints
}

// parseToFloats converts an array of strings into an array of floats.
func parseToFloats(values []string) []float64 {
	floats := make([]float64, 0, len(values))
	for _, val := range values {
		if floatVal, err := strconv.ParseFloat(val, 64); err == nil {
			floats = append(floats, floatVal)
		}
	}
	return floats
}

// parseToBools converts an array of strings into an array of booleans.
func parseToBools(values []string) []bool {
	bools := make([]bool, 0, len(values))
	for _, val := range values {
		if boolVal, err := strconv.ParseBool(val); err == nil {
			bools = append(bools, boolVal)
		}
	}
	return bools
}

// GenerateFilters generates a map of filters for the given filter fields using the provided function to get the value of parameters.
// If the type of the filter field is not recognized, the raw string value is used.
func GenerateFilters(getParamFunc func(string) string, filterFields []FilterRequest) map[string]interface{} {
	filters := make(map[string]interface{})
	handlers := map[FilterType]DataTypeHandler{
		FieldTypeString:    getStringHandler,
		FieldTypeInt:       getIntHandler,
		FieldTypeFloat64:   getFloatHandler,
		FieldTypeBoolean:   getBooleanHandler,
		FieldTypeDate:      getDateHandler,
		FieldTypeRangeDate: getRangeDateHandler,
	}

	for _, field := range filterFields {
		// if field.Param == "" || field.Param == "-" {
		if field.Param == "" {
			continue
		}

		filterValue := ""

		if field.Param != "-" {
			filterValue = getParamFunc(field.Param)
			if filterValue == "" {
				continue
			}
		}

		if field.Field == "" {
			field.Field = field.Param
		}

		if handler, ok := handlers[field.Type]; ok {
			result := handler(field, filterValue, getParamFunc)
			if result != nil {
				filters[field.Field] = result
			}
		} else {
			filters[field.Field] = filterValue
		}
	}

	return filters
}
