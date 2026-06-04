package handlers

import (
	"context"
	"fmt"
	"net/http"
	"smlcloudplatform/internal/goapi/logger"
	mypg "smlcloudplatform/internal/goapi/mypg"
	"strconv"
	"strings"
	"time"

	serviceConfig "smlcloudplatform/internal/goapi/config"
	myGlobal "smlcloudplatform/internal/goapi/myglobal"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
)

func MongoGetDataHandler(c echo.Context) error {
	logger.Info("MongoSelectHandler called")
	var payLoad struct {
		HoldingCode string `json:"holding_code"`
		Collection  string `json:"collection"`
		GuidFixed   string `json:"guid_fixed"`
	}

	if err := c.Bind(&payLoad); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON payload",
			"code":  "INVALID_JSON",
		})
	}

	// Debug: Log incoming payload
	logger.Info("[MongoGetDataHandler] Payload: holding_code=%s, collection=%s, guidfixed=%s", payLoad.HoldingCode, payLoad.Collection, payLoad.GuidFixed)

	// ⭐ แก้ไข: เพิ่มการตรวจสอบ Collection
	if payLoad.HoldingCode == "" || payLoad.GuidFixed == "" || payLoad.Collection == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Missing holding_code, collection, or guidfixed",
			"code":  "MISSING_PARAMETERS",
		})
	}

	// query from mongodb
	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		logger.Info("MongoConnect failed")
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "MongoDB connection failed",
			"code":  "MONGO_CONNECTION_ERROR",
		})
	}

	svcConfig := serviceConfig.NewServiceConfig()
	MongodbDatabaseName := svcConfig.MongodbDatabaseName()
	collection := mongoClient.Database(MongodbDatabaseName).Collection(payLoad.Collection)

	filter := bson.M{"holding_code": payLoad.HoldingCode, "guid_fixed": payLoad.GuidFixed}

	// ⭐ แก้ไข: Log ก่อน Find และใช้ collection name ที่ถูกต้อง
	logger.Info("Finding documents in MongoDB collection '%s' with filter: %+v", payLoad.Collection, filter)

	cur, err := collection.Find(context.Background(), filter)
	if err != nil {
		logger.Error("finding documents: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database query error",
			"code":  "DB_QUERY_ERROR",
		})
	}
	defer cur.Close(context.Background())

	var result map[string]any
	if cur.Next(context.Background()) {
		if err := cur.Decode(&result); err != nil {
			logger.Error("decoding document: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Error decoding document",
				"code":  "DECODE_ERROR",
			})
		}
	} else {
		// ไม่พบข้อมูล
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Document not found",
			"code":  "NOT_FOUND",
		})
	}

	// ⭐ แก้ไข: เพิ่ม cursor error checking
	if err := cur.Err(); err != nil {
		logger.Info("Cursor error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database cursor error",
			"code":  "DB_CURSOR_ERROR",
		})
	}

	logger.Info("Found document in collection '%s'", payLoad.Collection)

	// Return single result
	response := map[string]any{
		"status": 200,
		"result": result, // เปลี่ยนจาก "results" เป็น "result" (รายการเดียว)
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return c.JSON(http.StatusOK, response)
}

// PostgreSQL Get Document Handler - supports multiple GET queries
func PgGetDocHandler(c echo.Context) error {
	logger.Info("PgGetDoc called")
	// รับ JSON payLoad จาก request body
	var payLoad struct {
		HoldingCode string   `json:"holding_code"`
		System      string   `json:"system"`
		OffSet      int      `json:"offset"`
		Limit       int      `json:"limit"`
		Search      string   `json:"search"`
		CustCode    string   `json:"custcode"`
		DateOrder   int      `json:"dateorder"` // 0=asc, 1=desc
		FromDate    string   `json:"fromdate"`  // วันที่เริ่มต้น format: "2026-02-01"
		ToDate      string   `json:"todate"`    // วันที่สิ้นสุด format: "2026-02-28"
		MinAmount   *float64 `json:"minamount"` // ยอดเงินต่ำสุด (nil = ไม่กรอง)
		MaxAmount   *float64 `json:"maxamount"` // ยอดเงินสูงสุด (nil = ไม่กรอง)
		CustCodes   []string `json:"custcodes"` // รายการเจ้าหนี้ที่เลือก (multi-select)
	}

	if err := c.Bind(&payLoad); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid JSON payload",
			"code":  "INVALID_JSON",
		})
	}

	// Debug: Log all incoming filter params
	logger.Info("[PgGetDoc] Incoming payload - System: %s, FromDate: '%s', ToDate: '%s', MinAmount: %v, MaxAmount: %v, CustCodes: %v",
		payLoad.System, payLoad.FromDate, payLoad.ToDate, payLoad.MinAmount, payLoad.MaxAmount, payLoad.CustCodes)

	var transflags []int
	switch strings.ToLower(payLoad.System) {
	case "purchase-order":
		transflags = []int{6}
	case "purchase-requisition":
		transflags = []int{21}
	case "rfq":
		transflags = []int{22}
	default:
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Unsupported system type",
			"code":  "INVALID_SYSTEM",
		})
	}

	// แปลง []int เป็น string สำหรับ IN clause
	transflagStrings := make([]string, len(transflags))
	for i, flag := range transflags {
		transflagStrings[i] = strconv.Itoa(flag)
	}
	transflagList := strings.Join(transflagStrings, ",")
	// query from postgresql

	// open postgresql connection
	db, err := mypg.PgSqlFastConnect(payLoad.HoldingCode)
	if err != nil {
		logger.Error("connecting to database: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database connection error",
			"code":  "DB_CONNECTION_ERROR",
		})
	}

	// query based on transflag (แสดงรายการทั้งหมดรวมที่ลบแล้ว)
	whereClause := fmt.Sprintf("transflag in (%s)", transflagList)
	// เพิ่มเงื่อนไขค้นหา (ถ้ามี)
	if payLoad.Search != "" {
		whereClause += fmt.Sprintf(" AND (docno ILIKE '%%%s%%' OR custcode ILIKE '%%%s%%')", payLoad.Search, payLoad.Search)
	}
	// กรองตามช่วงวันที่
	if payLoad.FromDate != "" {
		logger.Info("[PgGetDoc] Filtering fromDate: %s", payLoad.FromDate)
		whereClause += fmt.Sprintf(" AND docdatetime >= '%s'", payLoad.FromDate)
	}
	if payLoad.ToDate != "" {
		// เพิ่ม 1 วันเพื่อให้รวมวันสุดท้ายด้วย
		whereClause += fmt.Sprintf(" AND docdatetime < '%s'::date + interval '1 day'", payLoad.ToDate)
	}
	// กรองตามช่วงจำนวนเงิน
	if payLoad.MinAmount != nil {
		whereClause += fmt.Sprintf(" AND totalamount >= %f", *payLoad.MinAmount)
	}
	if payLoad.MaxAmount != nil {
		whereClause += fmt.Sprintf(" AND totalamount <= %f", *payLoad.MaxAmount)
	}
	// กรองตามเจ้าหนี้ (multi-select)
	if len(payLoad.CustCodes) > 0 {
		quotedCodes := make([]string, len(payLoad.CustCodes))
		for i, code := range payLoad.CustCodes {
			quotedCodes[i] = fmt.Sprintf("'%s'", code)
		}
		whereClause += fmt.Sprintf(" AND custcode IN (%s)", strings.Join(quotedCodes, ","))
	}

	// Debug: Log the where clause to verify filtering
	logger.Info("[PgGetDoc] whereClause: %s", whereClause)

	queryCount := fmt.Sprintf("SELECT count(*) FROM doc WHERE %s", whereClause)
	var totalCount int
	err = db.QueryRow(queryCount).Scan(&totalCount)
	if err != nil {
		logger.Error("Query %s", queryCount)
		logger.Error("executing count query: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database query error",
			"code":  "DB_QUERY_ERROR",
		})
	}

	// docdatetime : วันที่/เวลา UTC+0
	// docno : เลขที่เอกสาร
	// custcode : รหัสผู้จำหน่าย/เจ้าหนี้
	// custname : ชื่อผู้จำหน่าย/เจ้าหนี้
	// totalamount : ยอดเงินรวม
	// detailcount : จำนวนรายการ (บรรทัด)
	// transflag : ประเภทเอกสาร
	// isref : มีการอ้างอืงเอกสารแล้วหรือไม่ (true/false) true=มีการอ้างอิงแล้ว
	// islocked : เอกสารถูกล็อคหรือไม่ (true/false) true=ถูกล็อค
	// isclosed : เอกสารถูกปิดหรือไม่ (true/false) true=ถูกปิดรับครบแล้ว หรือรับเกิน
	// ใช้ DISTINCT ON เพื่อป้องกันรายการซ้ำ (กรณีมีข้อมูลซ้ำใน database)
	// ใช้ subquery เพื่อให้ DISTINCT ON ทำงานก่อน แล้วค่อยเรียงลำดับทีหลัง
	// ใช้ COALESCE เพื่อ handle NULL values สำหรับ creator fields (เอกสารเก่าไม่มี fields เหล่านี้)
	distinctQuery := fmt.Sprintf("SELECT DISTINCT ON (docno) guidfixed, docdatetime, docno, custcode, coalesce((select name0 from creditor where creditor.code = doc.custcode), 'X') as custname, totalamount, (select count(*) from docdetail where docdetail.docno = doc.docno and docdetail.transflag = doc.transflag) as detailcount, transflag, isref, islocked, isclosed, COALESCE(creator_code, '') as creator_code, COALESCE(creator_name, '') as creator_name, COALESCE(created_at, docdatetime) as created_at, COALESCE(doc_currency, '') as doc_currency, COALESCE(doc_currency_symbol, '') as doc_currency_symbol, COALESCE(exchange_rate, 0) as exchange_rate, COALESCE(totalamount_doc, 0) as totalamount_doc, COALESCE(iscancel, false) as iscancel, COALESCE(isdelete, false) as isdelete, COALESCE(iscomparedsuccess, 0) as iscomparedsuccess, COALESCE(isclosedmanual, false) as isclosedmanual, COALESCE(closedmanual_by_code, '') as closedmanual_by_code, COALESCE(closedmanual_by_name, '') as closedmanual_by_name, COALESCE(closedmanual_at, '1970-01-01') as closedmanual_at, COALESCE(closedmanual_reason, '') as closedmanual_reason FROM doc WHERE %s ORDER BY docno, docdatetime DESC", whereClause)
	query := fmt.Sprintf("SELECT * FROM (%s) AS unique_docs ORDER BY docdatetime", distinctQuery)
	// เรียงลำดับวันที่/เวลา (ใช้ docdatetime เต็มรวมเวลาด้วย)
	if payLoad.DateOrder == 1 {
		query += " DESC"
	}
	// เรียงลำดับเลขที่เอกสารเป็นลำดับรอง
	query += " , docno DESC"

	if payLoad.OffSet > 0 {
		query += fmt.Sprintf(" OFFSET %d", payLoad.OffSet)
	}
	if payLoad.Limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", payLoad.Limit)
	}

	rows, err := db.Query(query)
	if err != nil {
		logger.Error("executing query: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database query error",
			"code":  "DB_QUERY_ERROR",
		})
	}
	defer rows.Close()

	results := []map[string]any{}
	for rows.Next() {
		var guidfixed string
		var docdatetime time.Time
		var docno string
		var custcode string
		var custname string
		var totalamount float64
		var detailcount int
		var transflag int
		var isref bool
		var islocked bool
		var isclosed bool
		var creatorCode, creatorName string
		var createdAt time.Time
		var docCurrency, docCurrencySymbol string
		var exchangeRate, totalamountDoc float64
		var iscancel, isdelete bool
		var iscomparedsuccess int
		var isclosedmanual bool
		var closedmanualByCode, closedmanualByName, closedmanualReason string
		var closedmanualAt time.Time
		if err := rows.Scan(&guidfixed, &docdatetime, &docno, &custcode, &custname, &totalamount, &detailcount, &transflag, &isref, &islocked, &isclosed, &creatorCode, &creatorName, &createdAt, &docCurrency, &docCurrencySymbol, &exchangeRate, &totalamountDoc, &iscancel, &isdelete, &iscomparedsuccess, &isclosedmanual, &closedmanualByCode, &closedmanualByName, &closedmanualAt, &closedmanualReason); err != nil {
			logger.Error("scanning row: %v", err)
			continue
		}

		results = append(results, map[string]any{
			"guid_fixed":           guidfixed,
			"docdatetime":          docdatetime.UTC().Format("2006-01-02T15:04:05.000Z"),
			"docno":                docno,
			"custcode":             custcode,
			"cust_name":            custname,
			"total_amount":         totalamount,
			"detailcount":          detailcount,
			"transflag":            transflag,
			"isref":                isref,
			"islocked":             islocked,
			"isclosed":             isclosed,
			"creator_code":         creatorCode,
			"creator_name":         creatorName,
			"created_at":           createdAt.UTC().Format("2006-01-02T15:04:05.000Z"),
			"doc_currency":         docCurrency,
			"doc_currencysymbol":   docCurrencySymbol,
			"exchange_rate":        exchangeRate,
			"totalamount_doc":      totalamountDoc,
			"iscancel":             iscancel,
			"isdelete":             isdelete,
			"iscomparedsuccess":    iscomparedsuccess,
			"isclosedmanual":       isclosedmanual,
			"closedmanual_by_code": closedmanualByCode,
			"closedmanual_by_name": closedmanualByName,
			"closedmanual_at":      closedmanualAt.UTC().Format("2006-01-02T15:04:05.000Z"),
			"closedmanual_reason":  closedmanualReason,
		})
	}

	// Return results
	response := map[string]any{
		"status":  200,
		"results": results,
		"total":   totalCount,
	}

	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	return c.JSON(http.StatusOK, response)
}
