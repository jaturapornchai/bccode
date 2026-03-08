package microservice

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/config"
	"sync"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type IPersisterClickHouse interface {
	Conn() driver.Conn
	Count(ctx context.Context, model interface{}, expr string, args ...interface{}) (int, error)
	Select(ctx context.Context, dest interface{}, sql string, args ...interface{}) error
	Exec(ctx context.Context, sql string, args ...interface{}) error
	Create(ctx context.Context, model any) error
	CreateInBatch(ctx context.Context, models []interface{}) error
}

type ChModel interface {
	TableName() string
}

type PersisterClickHouse struct {
	cfg config.IPersisterClickHouseConfig

	dbMutex sync.Mutex
	db      driver.Conn
}

func NewPersisterClickHouse(cfg config.IPersisterClickHouseConfig) *PersisterClickHouse {

	pst := PersisterClickHouse{
		cfg: cfg,
	}

	_, err := pst.getClient()
	if err != nil {
		panic(err)
	}

	return &pst
}

func (pst *PersisterClickHouse) getClient() (driver.Conn, error) {
	if pst.db != nil {
		return pst.db, nil
	}

	connection, err := clickhouse.Open(&clickhouse.Options{
		Addr: pst.cfg.ServerAddress(),
		Auth: clickhouse.Auth{
			Database: pst.cfg.DatabaseName(),
			Username: pst.cfg.Username(),
			Password: pst.cfg.Password(),
		},
		Debug: true,
	})

	if err != nil {
		return nil, err
	}
	pst.dbMutex.Lock()
	pst.db = connection
	pst.dbMutex.Unlock()

	return pst.db, nil
}

func (pst *PersisterClickHouse) GetTableName(model interface{}) (string, error) {

	chmodel, ok := model.(ChModel)

	if ok {
		return chmodel.TableName(), nil
	}
	return "", fmt.Errorf("struct is not implement Clickhouse Model")
}

func (pst *PersisterClickHouse) Conn() driver.Conn {
	return pst.db
}

func (pst *PersisterClickHouse) Count(ctx context.Context, model interface{}, expr string, args ...interface{}) (int, error) {
	tableName, err := pst.GetTableName(model)
	if err != nil {
		return 0, err
	}

	conn := pst.Conn()

	whereExpr := ""

	if expr != "" {
		whereExpr = fmt.Sprintf("WHERE %s", expr)
	}

	query := fmt.Sprintf("SELECT COUNT(*) as count FROM %s %s ", tableName, whereExpr)
	var results []struct {
		Count uint64 `ch:"count"`
	}
	err = conn.Select(ctx, &results, query, args...)

	if err != nil {
		return 0, err
	}

	return int(results[0].Count), nil
}

func (pst *PersisterClickHouse) Select(ctx context.Context, dest interface{}, sql string, args ...interface{}) error {
	conn := pst.Conn()
	return conn.Select(ctx, dest, sql, args...)
}

func (pst *PersisterClickHouse) Exec(ctx context.Context, sql string, args ...interface{}) error {
	conn := pst.Conn()

	return conn.Exec(ctx, sql, args...)
}

func (pst *PersisterClickHouse) Create(ctx context.Context, model interface{}) error {
	tableName, err := pst.GetTableName(model)
	if err != nil {
		return err
	}

	conn := pst.Conn()

	prepareBatch, err := conn.PrepareBatch(context.Background(), fmt.Sprintf("INSERT INTO %s", tableName))

	if err != nil {
		return err
	}

	err = prepareBatch.AppendStruct(model)
	if err != nil {
		return err
	}

	err = prepareBatch.Send()
	if err != nil {
		return err
	}

	return nil
}

func (pst *PersisterClickHouse) CreateInBatch(ctx context.Context, models []interface{}) error {
	if len(models) == 0 {
		return nil
	}

	tableName, err := pst.GetTableName(models[0])
	if err != nil {
		return err
	}

	conn := pst.Conn()

	prepareBatch, err := conn.PrepareBatch(ctx, fmt.Sprintf("INSERT INTO %s", tableName))

	if err != nil {
		return err
	}

	for _, model := range models {
		err = prepareBatch.AppendStruct(model)
		if err != nil {
			return err
		}
	}

	err = prepareBatch.Send()
	if err != nil {
		return err
	}

	return nil
}
