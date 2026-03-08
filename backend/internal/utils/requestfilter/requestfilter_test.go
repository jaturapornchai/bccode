package requestfilter_test

import (
	"smlcloudplatform/internal/utils/requestfilter"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson"
)

func TestGenerateFilters(t *testing.T) {
	filterFields := []requestfilter.FilterRequest{
		{Param: "strparam", Field: "strfield", Type: requestfilter.FieldTypeString},
		{Param: "intparam", Field: "intfield", Type: requestfilter.FieldTypeInt},
		{Param: "floatparam", Field: "floatfield", Type: requestfilter.FieldTypeFloat64},
		{Param: "boolparam", Field: "boolfield", Type: requestfilter.FieldTypeBoolean},
		{Param: "dateparam", Field: "datefield", Type: requestfilter.FieldTypeRangeDate},
	}

	getParamFunc := func(param string) string {
		switch param {
		case "strparam":
			return "strvalue"
		case "intparam":
			return "42"
		case "floatparam":
			return "42.42"
		case "boolparam":
			return "true"
		case "dateparam":
			return "2023-07-05"
		case "fromdate":
			return "2023-07-05"
		case "todate":
			return "2023-07-06"
		default:
			return ""
		}
	}

	filters := requestfilter.GenerateFilters(getParamFunc, filterFields)

	if filters["strfield"] != "strvalue" {
		t.Errorf("Unexpected value for strfield: got %v", filters["strfield"])
	}

	if len(filters["strfieldarr"].(bson.M)["$in"].([]string)) != 3 {
		t.Errorf("Unexpected value for strfieldarr: got %v", filters["strfieldarr"])
	}

	if filters["intfield"] != 42 {
		t.Errorf("Unexpected value for intfield: got %v", filters["intfield"])
	}

	if filters["floatfield"] != 42.42 {
		t.Errorf("Unexpected value for floatfield: got %v", filters["floatfield"])
	}

	if filters["boolfield"] != true {
		t.Errorf("Unexpected value for boolfield: got %v", filters["boolfield"])
	}

	fromDate, _ := time.Parse(requestfilter.DateFormate, "2023-07-05")
	toDate, _ := time.Parse(requestfilter.DateFormate, "2023-07-06")

	expectedDateRange := bson.M{
		"$gte": fromDate,
		"$lt":  toDate.AddDate(0, 0, 1),
	}

	if dateRange, ok := filters["datefield"].(bson.M); ok {
		if !dateRange["$gte"].(time.Time).Equal(expectedDateRange["$gte"].(time.Time)) || !dateRange["$lt"].(time.Time).Equal(expectedDateRange["$lt"].(time.Time)) {
			t.Errorf("Unexpected value for datefield: got %v", filters["datefield"])
		}
	} else {
		t.Errorf("Unexpected type for datefield: got %T", filters["datefield"])
	}
}

func TestGenerateFiltersComma(t *testing.T) {
	getParamFunc := func(param string) string {
		switch param {
		case "strparam":
			return "value1,value2"
		case "intparam":
			return "42,43"
		case "floatparam":
			return "42.42,43.43"
		case "boolparam":
			return "true,false"
		default:
			return ""
		}
	}

	t.Run("string array", func(t *testing.T) {
		filterField := []requestfilter.FilterRequest{
			{Param: "strparam", Field: "strfield", Type: requestfilter.FieldTypeString},
		}

		stringField := requestfilter.GenerateFilters(getParamFunc, filterField)

		fieldValue := stringField["strfield"].(bson.M)["$in"].([]string)

		assert.Equal(t, []string{"value1", "value2"}, fieldValue)
	})

	t.Run("int array", func(t *testing.T) {
		filterField := []requestfilter.FilterRequest{
			{Param: "intparam", Field: "intfield", Type: requestfilter.FieldTypeInt},
		}

		intField := requestfilter.GenerateFilters(getParamFunc, filterField)

		fieldValue := intField["intfield"].(bson.M)["$in"].([]int)

		assert.Equal(t, []int{42, 43}, fieldValue)
	})

	t.Run("float array", func(t *testing.T) {
		filterField := []requestfilter.FilterRequest{
			{Param: "floatparam", Field: "floatfield", Type: requestfilter.FieldTypeFloat64},
		}

		floatField := requestfilter.GenerateFilters(getParamFunc, filterField)

		fieldValue := floatField["floatfield"].(bson.M)["$in"].([]float64)

		assert.Equal(t, []float64{42.42, 43.43}, fieldValue)
	})

	t.Run("bool array", func(t *testing.T) {
		filterField := []requestfilter.FilterRequest{
			{Param: "boolparam", Field: "boolfield", Type: requestfilter.FieldTypeBoolean},
		}

		boolField := requestfilter.GenerateFilters(getParamFunc, filterField)

		fieldValue := boolField["boolfield"].(bson.M)["$in"].([]bool)

		assert.Equal(t, []bool{true, false}, fieldValue)
	})
}
