package microservice

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/config"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	internalLogger "smlcloudplatform/internal/logger"
)

// IPersister is interface for persister
type IPersister interface {
	WhereSP(model interface{}, sortexpr string, pageLimit int, page int, expr string, args ...interface{}) ( /*result*/ interface{}, error)
	WhereS(model interface{}, sortexpr string, expr string, args ...interface{}) ( /*result*/ interface{}, error)
	WhereP(model interface{}, pageLimit int, page int, expr string, args ...interface{}) ( /*result*/ interface{}, error)
	Where(model interface{}, expr string, args ...interface{}) ( /*result*/ interface{}, error)
	FindOne(model interface{}, idColumn string, id string) ( /*result*/ interface{}, error)
	First(model interface{}, query interface{}, args ...interface{}) ( /*result*/ interface{}, error)
	Create(model interface{}) error
	Update(model interface{}, where map[string]interface{}) error
	Delete(model interface{}, where map[string]interface{}) error
	CreateInBatch(models interface{}, bulkSize int) error
	CreateInBatchOnConflict(models interface{}, bulkSize int) error
	CreateInBatchClauses(models interface{}, bulkSize int, expression ...clause.Expression) error
	Exec(sql string, args ...interface{}) error
	TableExists(model interface{}) (bool, error)
	Count(model interface{}, expr string, args ...interface{}) (int64, error)
	DropTable(table ...interface{}) error
	SetupJoinTable(model interface{}, field string, joinTable interface{}) error
	AutoMigrate(dst ...interface{}) error
	TestConnect() error
	Transaction(funcTransaction func(*Persister) error) error
	Raw(queryStr string, where map[string]interface{}, model interface{}) ( /*result*/ interface{}, error)
	DBClient() *gorm.DB
}

// Persister is persister
type Persister struct {
	config  config.IPersisterConfig
	db      *gorm.DB
	dbMutex sync.Mutex
}

// NewPersister return new persister
func NewPersister(config config.IPersisterConfig) *Persister {
	pst := &Persister{
		config: config,
	}

	_, err := pst.getClient()
	if err != nil {
		panic(err)
	}

	return pst
}

func NewPersisterWithDB(db *gorm.DB) *Persister {
	pst := &Persister{
		db: db,
	}
	return pst
}

func (pst *Persister) getConnectionString() (string, error) {
	cfg := pst.config

	return fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s TimeZone=%s",
		cfg.Host(),
		cfg.Username(),
		cfg.Password(),
		cfg.DB(),
		cfg.Port(),
		cfg.SSLMode(),
		cfg.TimeZone(),
	), nil
}

func (pst *Persister) GetLeggerLevel() logger.LogLevel {

	logLevel := pst.config.LoggerLevel()
	if logLevel == "debug" || logLevel == "DEBUG" {
		return logger.Info
	}
	if logLevel == "error" || logLevel == "ERROR" {
		return logger.Error
	}
	if logLevel == "warn" || logLevel == "WARN" {
		return logger.Warn
	}
	return logger.Silent
}

func (pst *Persister) getClient() (*gorm.DB, error) {
	if pst.db != nil {
		return pst.db, nil
	}

	pst.dbMutex.Lock()
	defer pst.dbMutex.Unlock()

	connection, err := pst.getConnectionString()
	if err != nil {
		return nil, err
	}
	loggerLevel := pst.GetLeggerLevel()
	pgLogger := NewPersisterPostgresLogger(loggerLevel)
	db, err := gorm.Open(postgres.Open(connection), &gorm.Config{
		SkipDefaultTransaction: true,
		Logger:                 pgLogger, // logger.Default.LogMode(loggerLevel), //
		PrepareStmt:            true,
	})
	if err != nil {
		return nil, err
	}

	pst.db = db

	return db, nil
}

// TableExists check if table exists
func (pst *Persister) TableExists(model interface{}) (bool, error) {
	db := pst.db

	has := db.Migrator().HasTable(model)

	return has, nil
}

// Exec execute sql
func (pst *Persister) Exec(sql string, args ...interface{}) error {
	db := pst.db

	if err := db.Exec(sql, args...).Error; err != nil {
		return err
	}
	return nil
}

func (pst *Persister) calcOffset(page int, pageLimit int) int {
	offset := 0
	if pageLimit > 0 {
		if page < 1 {
			page = 1
		}
		offset = (page - 1) * pageLimit
	}
	return offset
}

// WhereSP find objects by expressions and sorting with paging
func (pst *Persister) WhereSP(model interface{}, sortexpr string, pageLimit int, page int, expr string, args ...interface{}) ( /*result*/ interface{}, error) {
	db := pst.db

	offset := pst.calcOffset(page, pageLimit)

	if len(sortexpr) > 0 && pageLimit > 0 {
		// Sorting and paging
		if err := db.Offset(offset).Limit(pageLimit).Order(sortexpr).Where(expr, args...).Find(model).Error; err != nil {
			return nil, err
		}
	} else if len(sortexpr) > 0 {
		// Sorting
		if err := db.Order(sortexpr).Where(expr, args...).Find(model).Error; err != nil {
			return nil, err
		}
	} else if pageLimit > 0 {
		// Paging
		if err := db.Offset(offset).Limit(pageLimit).Where(expr, args...).Find(model).Error; err != nil {
			return nil, err
		}
	} else {
		// No Sorting, No Paging
		if err := db.Where(expr, args...).Find(model).Error; err != nil {
			return nil, err
		}
	}
	return model, nil
}

// WhereS find objects by expressions and sorting
func (pst *Persister) WhereS(model interface{}, sortexpr string, expr string, args ...interface{}) ( /*result*/ interface{}, error) {
	return pst.WhereSP(model, sortexpr, -1, -1, expr, args...)
}

// WhereP find objects by expression and paging
func (pst *Persister) WhereP(model interface{}, pageLimit int, page int, expr string, args ...interface{}) ( /*result*/ interface{}, error) {
	return pst.WhereSP(model, "", pageLimit, page, expr, args...)
}

// Where find objects by expressions
func (pst *Persister) Where(model interface{}, expr string, args ...interface{}) ( /*result*/ interface{}, error) {
	return pst.WhereSP(model, "", -1, -1, expr, args...)
}

// Count return count by expression
func (pst *Persister) Count(model interface{}, expr string, args ...interface{}) (int64, error) {
	db := pst.db

	count := new(int64)
	if err := db.Model(model).Where(expr, args...).Count(count).Error; err != nil {
		return 0, err
	}

	return *count, nil
}

// FindOne find object by id
func (pst *Persister) FindOne(model interface{}, idColumn string, id string) ( /*result*/ interface{}, error) {
	db := pst.db

	where := fmt.Sprintf("%s = ?", idColumn)
	if err := db.Where(where, id).First(model).Error; err != nil {
		return nil, err
	}
	return model, nil
}

func (pst *Persister) First(model interface{}, query interface{}, args ...interface{}) ( /*result*/ interface{}, error) {

	db := pst.db
	if err := db.Where(query, args...).First(model).Error; err != nil {
		return nil, err
	}
	return model, nil
}

// Create create the object
func (pst *Persister) Create(model interface{}) error {
	db := pst.db

	err := db.Create(model).Error
	if err != nil {
		return err
	}

	return nil
}

// Update update the object
func (pst *Persister) Update(model interface{}, where map[string]interface{}) error {

	db := pst.db

	whereConditions := ""
	whereValues := []interface{}{}

	for columnName, value := range where {
		if len(whereConditions) != 0 {
			whereConditions = whereConditions + " AND "
		}
		whereConditions = whereConditions + fmt.Sprintf(`"%s" = ? `, columnName)
		whereValues = append(whereValues, value)
	}

	err := db.Model(model).Where(whereConditions, whereValues...).Updates(model).Error
	if err != nil {
		return err
	}

	return nil
}

// Delete update the object
func (pst *Persister) Delete(model interface{}, where map[string]interface{}) error {

	db := pst.db

	whereConditions := ""
	whereValues := []interface{}{}

	for columnName, value := range where {
		if len(whereConditions) != 0 {
			whereConditions = whereConditions + " AND "
		}
		whereConditions = whereConditions + fmt.Sprintf(`"%s" = ? `, columnName)
		whereValues = append(whereValues, value)
	}

	err := db.Where(whereConditions, whereValues...).Delete(model).Error
	if err != nil {
		return err
	}

	return nil
}

// CreateInBatch create the objects in batch
func (pst *Persister) CreateInBatch(models interface{}, bulkSize int) error {
	db := pst.db

	db.CreateInBatches(models, bulkSize)

	return nil
}

// CreateInBatch create the objects in batch
func (pst *Persister) CreateInBatchOnConflict(models interface{}, bulkSize int) error {
	db := pst.db
	db.Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(models, bulkSize)
	return nil
}

func (pst *Persister) CreateInBatchClauses(models interface{}, bulkSize int, expression ...clause.Expression) error {
	db := pst.db
	db.Clauses(expression...).CreateInBatches(models, bulkSize)
	return nil
}

func (pst *Persister) DropTable(table ...interface{}) error {
	db := pst.db

	return db.Migrator().DropTable(table...)
}

func (pst *Persister) SetupJoinTable(model interface{}, field string, joinTable interface{}) error {
	db := pst.db

	return db.SetupJoinTable(model, field, joinTable)
}

func (pst *Persister) AutoMigrate(dst ...interface{}) error {
	db := pst.db
	return db.AutoMigrate(dst...)
}

func (pst *Persister) TestConnect() error {
	db := pst.db

	var success int
	err := db.Raw("SELECT 1").Scan(&success).Error
	return err
}

func (pst *Persister) Transaction(funcTransaction func(*Persister) error) error {
	pst.db.Transaction(func(tx *gorm.DB) error {
		pst := &Persister{
			config: nil,
			db:     tx,
		}
		return funcTransaction(pst)
	})

	return nil
}

func (pst *Persister) Raw(queryStr string, where map[string]interface{}, model interface{}) ( /*result*/ interface{}, error) {
	err := pst.db.Raw(queryStr, where).Find(model).Error
	return model, err
}

func (pst *Persister) DBClient() *gorm.DB {
	return pst.db
}

type ContextFn func(ctx context.Context) []zapcore.Field

type PersisterPostgresLogger struct {
	LogLevel                  logger.LogLevel
	SlowThreshold             time.Duration
	Context                   ContextFn
	SkipCallerLookup          bool
	IgnoreRecordNotFoundError bool
}

func (l *PersisterPostgresLogger) SetAsDefault() {
	logger.Default = l
}

func NewPersisterPostgresLogger(level logger.LogLevel) *PersisterPostgresLogger {
	return &PersisterPostgresLogger{
		LogLevel:                  level,
		SlowThreshold:             100 * time.Millisecond,
		SkipCallerLookup:          false,
		IgnoreRecordNotFoundError: false,
	}
}

func (l *PersisterPostgresLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	newLogger.LogLevel = level
	return &newLogger
}

func (l PersisterPostgresLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		internalLogger.GetLogger().Info(msg)
	}
}

func (l PersisterPostgresLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		internalLogger.GetLogger().Warn(msg)
	}
}

func (l PersisterPostgresLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		internalLogger.GetLogger().Error(msg)
	}
}

func (l PersisterPostgresLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {

	if l.LogLevel <= 0 {
		return
	}

	elapsed := time.Since(begin)

	switch {
	case err != nil && l.LogLevel >= logger.Error && (!l.IgnoreRecordNotFoundError || !errors.Is(err, gorm.ErrRecordNotFound)):
		sql, rows := fc()
		internalLogger.GetLogger().Error("trace", zap.Error(err), zap.Duration("elapsed", elapsed), zap.Int64("rows", rows), zap.String("sql", sql))
	case l.SlowThreshold != 0 && elapsed > l.SlowThreshold && l.LogLevel >= logger.Warn:
		sql, rows := fc()
		internalLogger.GetLogger().Warn("trace", zap.Duration("elapsed", elapsed), zap.Int64("rows", rows), zap.String("sql", sql))
	case l.LogLevel >= logger.Info:
		sql, rows := fc()
		internalLogger.GetLogger().Debug("trace", zap.Duration("elapsed", elapsed), zap.Int64("rows", rows), zap.String("sql", sql))
	}
}
