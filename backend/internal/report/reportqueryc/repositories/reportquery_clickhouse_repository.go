package repositories

import (
	"context"
	"fmt"
	"math"
	"reflect"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/report/reportqueryc/models"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type IReportQueryClickHouseRepository interface {
	Execute(queryParm models.Query, pageable micromodels.Pageable) ([]map[string]interface{}, common.Pagination, error)
	Playground(queryParm models.Query) ([]map[string]interface{}, error)
}

type ReportQueryClickHouseRepository struct {
	pst microservice.IPersisterClickHouse
}

func NewReportQueryClickHouseRepository(pst microservice.IPersisterClickHouse) *ReportQueryClickHouseRepository {

	insRepo := &ReportQueryClickHouseRepository{
		pst: pst,
	}

	return insRepo
}

func (repo *ReportQueryClickHouseRepository) Execute(queryParm models.Query, pageable micromodels.Pageable) ([]map[string]interface{}, common.Pagination, error) {
	conn := repo.pst.Conn()

	params, err := prepareParams(queryParm.Params)
	if err != nil {
		return nil, common.Pagination{}, err
	}

	queryCount := fmt.Sprintf("SELECT COUNT(*) as count FROM (%s)", queryParm.SQL)

	var count uint64
	err = conn.QueryRow(context.Background(), queryCount, params...).Scan(&count)

	if err != nil {
		return nil, common.Pagination{}, err
	}

	totalPage := math.Ceil(float64(count) / float64(pageable.Limit))

	pagination := common.Pagination{
		Total:     int(count),
		Page:      pageable.Page,
		PerPage:   pageable.Limit,
		TotalPage: int(totalPage),
	}
	offset := pageable.GetOffest()

	query := fmt.Sprintf("%s LIMIT %d OFFSET %d", queryParm.SQL, pageable.Limit, offset)

	result, err := executeQuery(conn, query, params)

	return result, pagination, err
}

////

func (repo *ReportQueryClickHouseRepository) Playground(queryParm models.Query) ([]map[string]interface{}, error) {
	query := fmt.Sprintf("%s LIMIT 10", queryParm.SQL)
	conn := repo.pst.Conn()

	params, err := prepareParams(queryParm.Params)
	if err != nil {
		return nil, err
	}

	return executeQuery(conn, query, params)
}

func prepareParams(queryParams []models.QueryParam) ([]interface{}, error) {
	params := []interface{}{}

	for _, param := range queryParams {
		convertedParam, err := convertParam(param)
		if err != nil {
			return nil, err
		}
		params = append(params, convertedParam)
	}

	return params, nil
}

func convertParam(param models.QueryParam) (interface{}, error) {
	switch param.Type {
	case "string":
		if value, ok := param.Value.(string); ok {
			return clickhouse.Named(param.Name, value), nil
		}
	case "int", "int32":
		if value, ok := param.Value.(int32); ok {
			return clickhouse.Named(param.Name, value), nil
		}
	case "int64":
		if value, ok := param.Value.(int64); ok {
			return clickhouse.Named(param.Name, value), nil
		}
	case "numeric":
		if value, ok := param.Value.(float64); ok {
			return clickhouse.Named(param.Name, value), nil
		}
	case "bool":
		if value, ok := param.Value.(bool); ok {
			return clickhouse.Named(param.Name, value), nil
		}
	case "date", "datetime", "datetime64":
		if value, ok := param.Value.(time.Time); ok {
			return clickhouse.Named(param.Name, value), nil
		}
	case "int8":
		if value, ok := param.Value.(int8); ok {
			return clickhouse.Named(param.Name, value), nil
		}
	case "int16":
		if value, ok := param.Value.(int16); ok {
			return clickhouse.Named(param.Name, value), nil
		}
	case "float32":
		if value, ok := param.Value.(float32); ok {
			return clickhouse.Named(param.Name, value), nil
		}
	case "uint8":
		if value, ok := param.Value.(uint8); ok {
			return clickhouse.Named(param.Name, value), nil
		}
	case "uint16":
		if value, ok := param.Value.(uint16); ok {
			return clickhouse.Named(param.Name, value), nil
		}
	case "uint32":
		if value, ok := param.Value.(uint32); ok {
			return clickhouse.Named(param.Name, value), nil
		}
	case "uint64":
		if value, ok := param.Value.(uint64); ok {
			return clickhouse.Named(param.Name, value), nil
		}
	case "enum8", "enum16":
		if value, ok := param.Value.(string); ok {
			return clickhouse.Named(param.Name, value), nil
		}
	default:
		return nil, fmt.Errorf("unsupported type: %s", param.Type)
	}

	return nil, fmt.Errorf("invalid value type: %s for parameter: %s", reflect.TypeOf(param.Value), param.Name)
}

func executeQuery(conn driver.Conn, query string, params []interface{}) ([]map[string]interface{}, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 21*time.Second)
	defer cancel()
	rows, err := conn.Query(ctx, query, params...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	colTypes := rows.ColumnTypes()

	var results []map[string]interface{}
	for rows.Next() {
		vals := scanValues(colTypes)

		if err = rows.Scan(vals...); err != nil {
			return nil, err
		}
		m := make(map[string]interface{})
		for i, col := range colTypes {
			m[col.Name()] = vals[i]
		}

		results = append(results, m)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return results, nil
}

func scanValues(colTypes []driver.ColumnType) []interface{} {
	vals := make([]interface{}, len(colTypes))

	for i, colType := range colTypes {
		val := newValueForType(colType.DatabaseTypeName())
		vals[i] = val
	}

	return vals
}

func newValueForType(typeName string) interface{} {
	switch typeName {
	case "String":
		return new(string)
	case "Int32":
		return new(int32)
	case "Int64":
		return new(int64)
	case "Float64":
		return new(float64)
	case "Bool":
		return new(bool)
	case "Date", "DateTime", "DateTime64":
		return new(time.Time)
	case "Int8":
		return new(int8)
	case "Int16":
		return new(int16)
	case "Float32":
		return new(float32)
	case "UInt8":
		return new(uint8)
	case "UInt16":
		return new(uint16)
	case "UInt32":
		return new(uint32)
	case "UInt64":
		return new(uint64)
	case "Enum8", "Enum16":
		return new(string)
	case "Array(String)":
		return new([]string)
	case "Array(Int32)":
		return new([]int32)
	default:
		return new(interface{})
		// default:
		// 	return nil, fmt.Errorf("unsupported type: %s", typeName)
	}
}
