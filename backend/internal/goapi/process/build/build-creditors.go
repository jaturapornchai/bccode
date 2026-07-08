package build

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"smlcloudplatform/internal/goapi/myclickhouse"
	"smlcloudplatform/internal/goapi/myglobal"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"

	"smlcloudplatform/internal/goapi/config"
	"smlcloudplatform/internal/goapi/mypg"

	"github.com/mitchellh/mapstructure"
)

func processInsertCreditorList(db *sql.DB, data *[]models.ProcessMongoCreditorModel) error {
	if len(*data) == 0 {
		return nil
	}

	columns := []string{
		"holding_code",
		"guid_fixed",
		"code",
		"names",
		"tax_id",
		"personal_type",
		"customer_type",
		"branch_number",
		"fund_code",
		"creditday",
		"email",
		"phone_primary",
		"phone_secondary",
		"addressforbilling",
	}

	rows := make([][]any, 0, len(*data))

	for i, item := range *data {
		if item.Code == "" {
			logger.Warn("Creditor at index %d has empty code, skipping", i)
			continue
		}

		// names → JSON string สำหรับ JSONB column
		namesJSON := "[]"
		if len(item.Names) > 0 {
			if b, err := json.Marshal(item.Names); err == nil {
				namesJSON = string(b)
			}
		}

		// addressforbilling → JSON string สำหรับ JSONB column
		addrJSON := "null"
		phonePrimary := ""
		phoneSecondary := ""
		if item.AddressBilling != nil {
			if b, err := json.Marshal(item.AddressBilling); err == nil {
				addrJSON = string(b)
			}
			if v, ok := item.AddressBilling["phone_primary"].(string); ok {
				phonePrimary = v
			}
			if v, ok := item.AddressBilling["phone_secondary"].(string); ok {
				phoneSecondary = v
			}
		}

		rows = append(rows, []any{
			item.HoldingCode,
			item.GuidFixed,
			item.Code,
			namesJSON,
			item.TaxID,
			item.PersonalType,
			item.CustomerType,
			item.BranchNumber,
			item.FundCode,
			item.CreditDay,
			item.Email,
			phonePrimary,
			phoneSecondary,
			addrJSON,
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	return mypg.BulkInsertWithCopy(ctx, db, "creditor", columns, rows)
}

func processCreditorDecode(jsonData string) models.ProcessMongoCreditorModel {
	jsonDecode := map[string]any{}
	err := json.Unmarshal([]byte(jsonData), &jsonDecode)
	if err != nil {
		logger.Error("Error unmarshal: %v", err)
		return models.ProcessMongoCreditorModel{}
	}

	fixedData := myglobal.FixMongoData(jsonDecode)

	var creditor models.ProcessMongoCreditorModel
	decoderConfig := mapstructure.DecoderConfig{
		DecodeHook: myglobal.CustomDecodeHookFunc,
		Result:     &creditor,
		TagName:    "json",
	}
	decoder, err := mapstructure.NewDecoder(&decoderConfig)
	if err != nil {
		logger.Error("creating decoder: %v", err)
		return models.ProcessMongoCreditorModel{}
	}

	err = decoder.Decode(fixedData)
	if err != nil {
		logger.Error("Error decoding map to struct: %v", err)
	}
	return creditor
}

func ProcessCreditorRebuildAll(holdingCode string) int {
	mongoClient := myglobal.SafeMongoConnectFast() // Use optimized connection
	if mongoClient == nil {
		logger.Error("MongoConnect failed")
		return 0
	}

	timeStart := time.Now()
	logger.Info("Starting CreditorRebuildAll for shop %s", holdingCode)

	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Info("Failed to connect to Postgres: %v", err)
		return 0
	}

	// delete existing rows for this shop
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_, err = db.ExecContext(ctx, "TRUNCATE TABLE creditor")
	if err != nil {
		logger.Error("deleting existing creditor rows: %v", err)
		return 0
	}
	logger.Info("Cleared creditor rows for holdingCode %s", holdingCode)

	// read from MongoDB
	svcConfig := config.NewServiceConfig()
	MongodbDatabaseName := svcConfig.MongodbDatabaseName()
	collection := mongoClient.Database(MongodbDatabaseName).Collection("creditors")
	cur, err := collection.Find(context.Background(), bson.M{"holdingcode": holdingCode, "deletedat": bson.M{"$exists": false}})
	logger.Info("Finding documents in MongoDB collection %s", MongodbDatabaseName)
	if err != nil {
		logger.Error("finding documents: %v", err)
		return 0
	}
	defer cur.Close(context.Background())

	var creditors []models.ProcessMongoCreditorModel

	for cur.Next(context.Background()) {
		creditors = append(creditors, processCreditorDecode(cur.Current.String()))
	}
	if err := cur.Err(); err != nil {
		logger.Info("Cursor error: %v", err)
		return 0
	}
	logger.Info("[Creditor] MongoDB → %d records for shop %s", len(creditors), holdingCode)

	// insert into Postgres
	pgStart := time.Now()
	err = processInsertCreditorList(db, &creditors)
	if err != nil {
		logger.Error("[Creditor] PG insert failed: %v", err)
		return 0
	}
	logger.Info("[Creditor] PG ← %d records (14 fields, pg_trgm indexes) [%v]", len(creditors), time.Since(pgStart))

	// sync to ClickHouse
	chStart := time.Now()
	err = RebuildClickHouseCreditor(holdingCode, &creditors)
	if err != nil {
		logger.Error("[Creditor] ClickHouse sync failed: %v", err)
	} else {
		logger.Info("[Creditor] ClickHouse ← %d records (15 fields, bloom_filter) [%v]", len(creditors), time.Since(chStart))
	}

	logger.Info("[Creditor] Completed: MongoDB(%d) → PG(%d) + ClickHouse(%d) for shop %s [total %v]",
		len(creditors), len(creditors), len(creditors), holdingCode, time.Since(timeStart))
	return len(creditors)
}

// RebuildClickHouseCreditor — sync creditors จาก MongoDB data ไป ClickHouse
func RebuildClickHouseCreditor(holdingCode string, creditors *[]models.ProcessMongoCreditorModel) error {
	ctx := context.Background()

	clickHouseDB, err := myclickhouse.ClickHouseFastConnect()
	if err != nil {
		return fmt.Errorf("failed to connect to ClickHouse: %v", err)
	}

	tableName := myclickhouse.TableName("creditors")

	// เพิ่ม column ที่ยังไม่มี (กรณี table เก่ามีแค่ 5 columns)
	newColumns := []string{
		"taxid String DEFAULT ''",
		"email String DEFAULT ''",
		"phoneprimary String DEFAULT ''",
		"phonesecondary String DEFAULT ''",
		"branchnumber String DEFAULT ''",
		"fundcode String DEFAULT ''",
		"personaltype Int8 DEFAULT 0",
		"customertype Int32 DEFAULT 0",
		"creditday Int32 DEFAULT 0",
		"address String DEFAULT ''",
	}
	for _, col := range newColumns {
		myclickhouse.ExecuteCommand(ctx, clickHouseDB,
			fmt.Sprintf("ALTER TABLE %s ADD COLUMN IF NOT EXISTS %s", tableName, col))
	}

	// ลบข้อมูลเก่าของ shop นี้
	deleteQuery := fmt.Sprintf("ALTER TABLE %s DELETE WHERE holding_code = '%s'", tableName, holdingCode)
	err = myclickhouse.ExecuteCommand(ctx, clickHouseDB, deleteQuery)
	if err != nil {
		logger.Error("Failed to delete old ClickHouse creditors: %v", err)
	}

	if len(*creditors) == 0 {
		return nil
	}

	// batch insert
	batchSize := 1000
	values := make([]string, 0, batchSize)
	count := 0

	for _, item := range *creditors {
		if item.Code == "" {
			continue
		}

		name1, name2 := "", ""
		if len(item.Names) > 0 {
			name1 = item.Names[0].Name
		}
		if len(item.Names) > 1 {
			name2 = item.Names[1].Name
		}

		// ดึง phone + address จาก addressforbilling
		phonePrimary, phoneSecondary, address := "", "", ""
		if item.AddressBilling != nil {
			if v, ok := item.AddressBilling["phone_primary"].(string); ok {
				phonePrimary = v
			}
			if v, ok := item.AddressBilling["phone_secondary"].(string); ok {
				phoneSecondary = v
			}
			if addrList, ok := item.AddressBilling["address"].([]interface{}); ok && len(addrList) > 0 {
				if v, ok := addrList[0].(string); ok {
					address = v
				}
			}
		}

		valueStr := fmt.Sprintf("('%s','%s','%s','%s','%s','%s','%s','%s','%s','%s','%s',%d,%d,%d,'%s')",
			escapeCH(holdingCode), escapeCH(item.GuidFixed), escapeCH(item.Code),
			escapeCH(name1), escapeCH(name2),
			escapeCH(item.TaxID), escapeCH(item.Email),
			escapeCH(phonePrimary), escapeCH(phoneSecondary),
			escapeCH(item.BranchNumber), escapeCH(item.FundCode),
			item.PersonalType, item.CustomerType, item.CreditDay,
			escapeCH(address))
		values = append(values, valueStr)
		count++

		if len(values) >= batchSize {
			insertQuery := fmt.Sprintf(
				"INSERT INTO %s (holding_code,guidfixed,code,name1,name2,taxid,email,phoneprimary,phonesecondary,branchnumber,fundcode,personaltype,customertype,creditday,address) VALUES %s",
				tableName, strings.Join(values, ","))
			err = myclickhouse.ExecuteCommand(ctx, clickHouseDB, insertQuery)
			if err != nil {
				logger.Error("Failed to insert creditor batch to ClickHouse: %v", err)
			}
			values = make([]string, 0, batchSize)
		}
	}

	// batch สุดท้าย
	if len(values) > 0 {
		insertQuery := fmt.Sprintf(
			"INSERT INTO %s (holding_code,guidfixed,code,name1,name2,taxid,email,phoneprimary,phonesecondary,branchnumber,fundcode,personaltype,customertype,creditday,address) VALUES %s",
			tableName, strings.Join(values, ","))
		err = myclickhouse.ExecuteCommand(ctx, clickHouseDB, insertQuery)
		if err != nil {
			logger.Error("Failed to insert final creditor batch to ClickHouse: %v", err)
		}
	}

	logger.Info("Inserted %d creditors to ClickHouse for shop %s", count, holdingCode)
	return nil
}

// escapeCH — escape single quotes สำหรับ ClickHouse SQL
func escapeCH(s string) string {
	return strings.ReplaceAll(s, "'", "\\'")
}
