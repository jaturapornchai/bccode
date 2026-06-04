package handlers

import (
	"context"
	"os"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/myclickhouse"
	"smlcloudplatform/internal/goapi/mypg"
	"strings"
)

func CloneClickHouseDatabase() {
	// ตรวจสอบว่าการตั้งค่า ENABLE_CLONE_CLICKHOUSE เป็น true หรือไม่
	if os.Getenv("ENABLE_CLONE_CLICKHOUSE") != "true" {
		logger.Info("Clone ClickHouse is disabled. Skipping operation.")
		return
	}
	// connect postgresql database name : bcaiconfig table shopgroup
	logger.Info("Starting ClickHouse database cloning process...")
	myPgConn, err := mypg.PgSqlFastConnect("bcaiconfig")
	if err != nil {
		logger.Error("Failed to connect to PostgreSQL: %v", err)
		return
	}

	// ดึงข้อมูล shopgroup
	rows, err := myPgConn.Query("SELECT groupname, holding_codelist FROM shopgroup")
	if err != nil {
		logger.Error("Failed to query shopgroup: %v", err)
		return
	}
	defer rows.Close()
	type ShopGroup struct {
		GroupName       string
		HoldingCodeList string
	}
	var shopGroups []ShopGroup
	for rows.Next() {
		var sg ShopGroup
		if err := rows.Scan(&sg.GroupName, &sg.HoldingCodeList); err != nil {
			logger.Error("Failed to scan shopgroup row: %v", err)
			return
		}
		shopGroups = append(shopGroups, sg)
	}
	logger.Info("Fetched %d shop groups", len(shopGroups))

	// ดำเนินการ clone database ที่นี่
	for _, sg := range shopGroups {
		logger.Info("Cloning ClickHouse database for group: %s with shops: %s", sg.GroupName, sg.HoldingCodeList)
		err := CloneClickHouseDatabaseStart(sg.GroupName, sg.HoldingCodeList)
		if err != nil {
			logger.Error("Failed to clone ClickHouse database: %v", err)
			return
		}
	}
}

// formatHoldingCodeList แปลง "x001,x002,x003" เป็น "'x001','x002','x003'"
func formatHoldingCodeList(holdingCodeList string) string {
	// ตัดช่องว่างออก
	holdingCodeList = strings.ReplaceAll(holdingCodeList, " ", "")
	holdingCodeList = strings.ReplaceAll(holdingCodeList, "'", "")

	// แยกด้วย comma
	shops := strings.Split(holdingCodeList, ",")

	// ใส่ quotes รอบแต่ละ shop
	quotedShops := make([]string, len(shops))
	for i, shop := range shops {
		if shop != "" {
			quotedShops[i] = "'" + shop + "'"
		}
	}

	// รวมกลับด้วย comma
	return strings.Join(quotedShops, ",")
}

func CloneClickHouseDatabaseStart(databaseName string, holdingCodeList string) error {
	// แปลง holdingCodeList จาก "x001,x002,x003" เป็น "'x001','x002','x003'"
	formattedHoldingCodeList := formatHoldingCodeList(holdingCodeList)
	logger.Info("📋 Original holdingCodeList: %s", holdingCodeList)
	logger.Info("📋 Formatted holdingCodeList: %s", formattedHoldingCodeList)

	// เชื่อมต่อ ClickHouse
	chClient, err := myclickhouse.CreateClickHouseConnection()
	if err != nil || chClient == nil {
		logger.Error("Failed to connect to ClickHouse: %v", err)
		return err
	}
	defer chClient.Close()

	// ⚠️ ห้าม CREATE DATABASE ใน ClickHouse - database ถูกสร้างและจัดการโดย DBA แล้ว
	// Database ต้องมีอยู่แล้วก่อนใช้งาน function นี้
	ctx := context.Background()
	logger.Info("Using existing ClickHouse database: %s", databaseName)

	// หา table docdetail ใน database นั้นก่อน ถ้าไม่มีให้สร้างใหม่
	// ถ้ามีให้ลบข้อมูลเก่าออกก่อนให้หมด
	query := "SELECT name FROM system.tables WHERE database = ? AND name = 'docdetail'"
	rows, err := chClient.Query(ctx, query, []any{databaseName}...)
	if err != nil {
		logger.Error("Failed to query tables in database %s: %v", databaseName, err)
		return err
	}
	defer rows.Close()
	tableExists := false
	for rows.Next() {
		var tableName string
		if err := rows.Scan(&tableName); err != nil {
			logger.Error("Failed to scan table name: %v", err)
			return err
		}
		if tableName == "docdetail" {
			tableExists = true
			break
		}
	}
	if err := rows.Err(); err != nil {
		logger.Error("Error iterating table rows: %v", err)
		return err
	}
	if tableExists {
		// ลบข้อมูลเก่าออก
		delQuery := "ALTER TABLE " + databaseName + ".docdetail DELETE WHERE holding_code IN (" + formattedHoldingCodeList + ")"
		err = chClient.Exec(ctx, delQuery)
		if err != nil {
			logger.Error("Failed to delete old data from docdetail in database %s: %v", databaseName, err)
			return err
		}
		logger.Info("Deleted old data from docdetail in database %s", databaseName)
	} else {
		// ⚠️ Table ไม่มี - ข้ามการสร้าง table (ต้องสร้างด้วย manual)
		logger.Warn("⚠️ Table docdetail does not exist in database %s. Skipping table creation (must be created manually).", databaseName)
	}

	// สุดท้าย: Clone ข้อมูลจาก source database
	sourceDB := myclickhouse.GetDatabaseName()
	logger.Info("🔄 Starting data clone from %s.docdetail to %s.docdetail", sourceDB, databaseName)

	insertQuery := `INSERT INTO ` + databaseName + `.docdetail
	SELECT *
	FROM ` + sourceDB + `.docdetail
	WHERE holding_code IN (` + formattedHoldingCodeList + `)`

	err = chClient.Exec(ctx, insertQuery)
	if err != nil {
		logger.Error("Failed to clone data to docdetail in database %s: %v", databaseName, err)
		return err
	}

	logger.Info("✅ Successfully cloned docdetail data to database %s for shops: %s", databaseName, formattedHoldingCodeList)

	return nil
}
