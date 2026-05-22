package myglobal

import (
	"compress/gzip"
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"reflect"
	coreconfig "smlcloudplatform/internal/config"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readconcern"
	"go.mongodb.org/mongo-driver/mongo/readpref"
	"go.mongodb.org/mongo-driver/mongo/writeconcern"

	"smlcloudplatform/internal/goapi/config"
)

// Global MongoDB client instance for connection pooling
var (
	mongoClient *mongo.Client
	mongoOnce   sync.Once
)

func CalculateMD5(input string) string {
	hash := md5.Sum([]byte(input))
	return hex.EncodeToString(hash[:]) // ผลลัพธ์ยาว 32 ตัวอักษร
}

// GetTransactionMultiplier returns calc multiplier based on transaction flag
func GetTransactionMultiplier(transFlag int) float64 {
	switch transFlag {
	case 12, 48, 60, 58, 66, 310: // ซื้อสินค้า, รับสินค้า, รับสำเร็จรูป, รับคืนจากการเบิก, ปรับสต็อก (เพิ่ม), รับสินค้า (พาเชียล)
		return 1.0
	case 16, 44, 56, 68: // ส่งคืนสินค้า, ขายสินค้า, เบิกสินค้า, ปรับสต็อก (ลด)
		return -1.0
	case 54: // Stock Balance - ปรับยอดคงเหลือ
		return 1.0
	case 72: // Transfer - โอนย้าย
		return 1.0
	case 6: // ใบสั่งซื้อ (PO) - ไม่กระทบสต็อก
		return 0
	default:
		return 1.0
	}
}

// MongoConnect returns a singleton MongoDB client with optimized connection pool
func MongoConnect() (*mongo.Client, error) {
	var err error
	mongoOnce.Do(func() {
		mongoClient, err = createMongoClient()
	})
	return mongoClient, err
}

func createMongoProductionClient() (*mongo.Client, error) {
	config := config.NewServiceConfig()
	uri := config.MongodbURI()
	if uri == "" {
		return nil, fmt.Errorf("MongoDB URI not found for %s environment", coreconfig.CurrentDataEnvironment())
	}

	// Optimized connection options for balanced performance
	clientOptions := options.Client().
		ApplyURI(uri).
		SetConnectTimeout(5 * time.Second).              // Faster connection timeout
		SetSocketTimeout(30 * time.Second).              // Socket timeout for operations
		SetServerSelectionTimeout(5 * time.Second).      // Fast server selection
		SetHeartbeatInterval(10 * time.Second).          // Connection health check
		SetMaxPoolSize(75).                              // OPTIMIZED: reduced from 100 to 75 for better resource balance
		SetMinPoolSize(10).                              // Keep minimum connections ready
		SetMaxConnIdleTime(20 * time.Minute).            // Extended from 15 to 20 minutes
		SetRetryWrites(true).                            // Auto retry failed writes
		SetRetryReads(true).                             // Auto retry failed reads
		SetCompressors([]string{"zstd", "snappy"}).      // Enable compression for faster data transfer
		SetReadPreference(readpref.SecondaryPreferred()) // Use secondary for reads when possible

	mongoClient, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("error connecting to MongoDB: %v", err)
	}

	// Quick ping test with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error ping to MongoDB: %v", err)
	}

	logger.Success("เชื่อมต่อ MongoDB สำเร็จ (pool: max=%d, min=%d, idle_time=20m)", 75, 10)
	return mongoClient, nil
}

func createMongoClient() (*mongo.Client, error) {
	config := config.NewServiceConfig()
	dataEnv := coreconfig.CurrentDataEnvironment()

	uri := config.MongodbURI()
	if uri == "" {
		return nil, fmt.Errorf("no MongoDB URI configured for %s environment", dataEnv)
	}
	serverDescription := fmt.Sprintf("%s MongoDB configured by environment-specific URI", strings.ToUpper(dataEnv))
	logger.Info("MongoDB connection target: %s", serverDescription)

	logger.Debug("MongoDB URI: %s", maskPassword(uri))

	// ⚡ OPTIMIZED: ลด pool size ตามคำแนะนำ (จาก 200 → 50)
	// - MaxPoolSize: 50 (เพียงพอสำหรับ operations ส่วนใหญ่)
	// - MinPoolSize: 10 (ลดจาก 20 เพื่อประหยัด resources)
	clientOptions := options.Client().
		ApplyURI(uri).
		SetConnectTimeout(10 * time.Second).                 // 10 วินาที
		SetSocketTimeout(60 * time.Second).                  // 60 วินาทีสำหรับ operations ใหญ่
		SetServerSelectionTimeout(10 * time.Second).         // 10 วินาที
		SetHeartbeatInterval(10 * time.Second).              // Connection health check
		SetMaxPoolSize(50).                                  // ลดจาก 200 → 50 (ตามคำแนะนำ)
		SetMinPoolSize(10).                                  // ลดจาก 20 → 10 (ตามคำแนะนำ)
		SetMaxConnIdleTime(30 * time.Minute).                // 30 นาที
		SetRetryWrites(true).                                // Auto retry failed writes
		SetRetryReads(true).                                 // Auto retry failed reads
		SetCompressors([]string{"zstd", "snappy", "zlib"}).  // Compression
		SetReadPreference(readpref.SecondaryPreferred()).    // Use secondary for reads when possible
		SetDirect(false).                                    // Use replica set routing
		SetReadConcern(readconcern.Local()).                 // Local read concern for better performance
		SetWriteConcern(writeconcern.New(writeconcern.W(1))) // W:1 for faster writes

	mongoClient, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		return nil, fmt.Errorf("error connecting to MongoDB: %v", err)
	}

	// Quick ping test with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second) // เพิ่มเป็น 5 วินาที
	defer cancel()

	err = mongoClient.Ping(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("error ping to MongoDB: %v", err)
	}

	logger.Success("เชื่อมต่อ %s สำเร็จ (pool: max=%d, min=%d)", serverDescription, 50, 10)
	return mongoClient, nil
}

// maskPassword masks sensitive data in logs
func maskPassword(uri string) string {
	// Pattern: mongodb://user:pass@host or mongodb+srv://user:pass@host
	if strings.Contains(uri, "://") && strings.Contains(uri, "@") {
		parts := strings.SplitN(uri, "://", 2)
		if len(parts) == 2 {
			afterProtocol := parts[1]
			atIndex := strings.Index(afterProtocol, "@")
			if atIndex > 0 {
				credentials := afterProtocol[:atIndex]
				rest := afterProtocol[atIndex:]

				// Hide password
				if strings.Contains(credentials, ":") {
					credParts := strings.SplitN(credentials, ":", 2)
					return parts[0] + "://" + credParts[0] + ":****" + rest
				}
			}
		}
	}

	return uri
}

// MongoConnectFast returns fast MongoDB client optimized for polling
func MongoConnectFast() (*mongo.Client, error) {
	return MongoConnect() // Use the singleton instance
}

// MongoFindFast performs optimized MongoDB find with polling support
func MongoFindFast(client *mongo.Client, database, collection string, filter any, opts ...*options.FindOptions) (*mongo.Cursor, error) {
	// Add timeout and optimize options
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Set optimized find options
	findOpts := options.Find()
	if len(opts) > 0 {
		findOpts = opts[0]
	}

	// Enable fast cursor options
	findOpts.SetNoCursorTimeout(false)    // Prevent cursor timeout
	findOpts.SetAllowPartialResults(true) // Allow partial results for speed

	coll := client.Database(database).Collection(collection)
	return coll.Find(ctx, filter, findOpts)
}

// MongoAggreggateFast performs optimized aggregation with polling
func MongoAggreggateFast(client *mongo.Client, database, collection string, pipeline any, opts ...*options.AggregateOptions) (*mongo.Cursor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	// Set optimized aggregate options
	aggOpts := options.Aggregate()
	if len(opts) > 0 {
		aggOpts = opts[0]
	}

	// Enable optimizations
	aggOpts.SetAllowDiskUse(true)        // Use disk for large datasets
	aggOpts.SetMaxTime(15 * time.Second) // Max execution time

	coll := client.Database(database).Collection(collection)
	return coll.Aggregate(ctx, pipeline, aggOpts)
}

// GetMongoClient returns the MongoDB client singleton
func GetMongoClient() *mongo.Client {
	return mongoClient
}

// GetMongoDatabase returns the MongoDB database instance
func GetMongoDatabase() *mongo.Database {
	if mongoClient == nil {
		return nil
	}
	config := config.NewServiceConfig()
	dbName := config.MongodbDatabaseName()
	if dbName == "" {
		dbName = config.MongoDBName()
	}
	return mongoClient.Database(dbName)
}

// DisconnectMongo closes MongoDB connection gracefully
func DisconnectMongo() {
	if mongoClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := mongoClient.Disconnect(ctx); err != nil {
			logger.Error("disconnecting from MongoDB: %v", err)
		} else {
			logger.Success("disconnected from MongoDB")
		}
	}
}

// MongoPolling performs fast polling of MongoDB collections with change streams
func MongoPolling(client *mongo.Client, database, collection string, filter any, callback func(any)) error {
	ctx := context.Background()

	coll := client.Database(database).Collection(collection)

	// Create change stream for real-time monitoring
	changeStreamOpts := options.ChangeStream().SetFullDocument(options.UpdateLookup)

	changeStream, err := coll.Watch(ctx, mongo.Pipeline{}, changeStreamOpts)
	if err != nil {
		return fmt.Errorf("failed to create change stream: %w", err)
	}
	defer changeStream.Close(ctx)

	logger.Info("Started MongoDB polling on %s.%s", database, collection)

	// Process change events
	for changeStream.Next(ctx) {
		var changeEvent map[string]any
		if err := changeStream.Decode(&changeEvent); err != nil {
			logger.Error("decoding change event: %v", err)
			continue
		}

		// Call the callback function with the change event
		callback(changeEvent)
	}

	if err := changeStream.Err(); err != nil {
		return fmt.Errorf("change stream error: %w", err)
	}

	return nil
}

// MongoBulkRead performs optimized bulk reading with cursor
func MongoBulkRead(client *mongo.Client, database, collection string, filter any, batchSize int, callback func([]any)) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Optimized find options for bulk reading
	findOpts := options.Find().
		SetBatchSize(int32(batchSize)). // Process in batches
		SetNoCursorTimeout(true).       // Prevent timeout for large datasets
		SetAllowPartialResults(true)    // Allow partial results for speed

	coll := client.Database(database).Collection(collection)
	cursor, err := coll.Find(ctx, filter, findOpts)
	if err != nil {
		return fmt.Errorf("failed to create cursor: %w", err)
	}
	defer cursor.Close(ctx)

	// Process documents in batches
	batch := make([]any, 0, batchSize)

	for cursor.Next(ctx) {
		var doc any
		if err := cursor.Decode(&doc); err != nil {
			logger.Error("decoding document: %v", err)
			continue
		}

		batch = append(batch, doc)

		// Process batch when full
		if len(batch) >= batchSize {
			callback(batch)
			batch = batch[:0] // Reset batch
		}
	}

	// Process remaining documents
	if len(batch) > 0 {
		callback(batch)
	}

	if err := cursor.Err(); err != nil {
		return fmt.Errorf("cursor error: %w", err)
	}

	return nil
}

func FixMongoData(data any) any {
	switch value := data.(type) {
	case map[string]any:
		if len(value) == 1 {
			if numberInt, ok := value["$numberInt"]; ok {
				switch numberInt.(type) {
				case string:
					var result int
					fmt.Sscanf(numberInt.(string), "%d", &result)
					return result
				default:
					return int(numberInt.(float64))
				}
			}
			if numberDouble, ok := value["$numberDouble"]; ok {
				switch numberDouble.(type) {
				case string:
					var result float64
					fmt.Sscanf(numberDouble.(string), "%f", &result)
					return result
				default:
					return numberDouble.(float64)
				}
			}
			if numberLong, ok := value["$numberLong"]; ok {
				switch numberLong.(type) {
				case string:
					ms, _ := time.ParseDuration(fmt.Sprintf("%sms", numberLong.(string)))
					return time.Unix(0, ms.Nanoseconds())
				default:
					ms, _ := time.ParseDuration(fmt.Sprintf("%sms", numberLong.(string)))
					return time.Unix(0, ms.Nanoseconds())
				}
			}
			if date, ok := value["$date"]; ok {
				switch date.(type) {
				case string:
					t, _ := time.Parse(time.RFC3339, date.(string))
					return t
				case map[string]any:
					if dateVal, ok := date.(map[string]any)["$numberLong"]; ok {
						switch dateVal.(type) {
						case string:
							ms, _ := time.ParseDuration(fmt.Sprintf("%sms", dateVal.(string)))
							return time.Unix(0, ms.Nanoseconds())
						}
					}
				}
			}
		}
		for k, v := range value {
			value[k] = FixMongoData(v)
		}
	case []any:
		for i, v := range value {
			value[i] = FixMongoData(v)
		}
	}
	return data
}

func CustomDecodeHookFunc(f reflect.Type, t reflect.Type, data any) (any, error) {
	if f.Kind() == reflect.Map {
		mapData := data.(map[string]any)
		if val, ok := mapData["$numberInt"]; ok && t.Kind() == reflect.Int {
			return int(val.(float64)), nil
		}
		if val, ok := mapData["$numberDouble"]; ok && t.Kind() == reflect.Float64 {
			switch val.(type) {
			case string:
				return fmt.Sscanf(val.(string), "%f")
			default:
				return val.(float64), nil
			}
		}
		if val, ok := mapData["$numberLong"]; ok && t == reflect.TypeOf(time.Time{}) {
			ms, _ := time.ParseDuration(fmt.Sprintf("%sms", val.(string)))
			return time.Unix(0, ms.Nanoseconds()), nil
		}
		if val, ok := mapData["$date"]; ok && t == reflect.TypeOf(time.Time{}) {
			switch val.(type) {
			case string:
				return time.Parse(time.RFC3339, val.(string))
			case map[string]any:
				if dateVal, ok := val.(map[string]any)["$numberLong"]; ok {
					switch dateVal.(type) {
					case string:
						ms, _ := time.ParseDuration(fmt.Sprintf("%sms", dateVal.(string)))
						return time.Unix(0, ms.Nanoseconds()), nil
					}
				}
			}
		}
	}

	if f.Kind() == reflect.String && t == reflect.TypeOf(time.Time{}) {
		return time.Parse(time.RFC3339, data.(string))
	}

	return data, nil
}

func GenUUID() string {
	// create guid
	return uuid.New().String()
}

func RoundFloat64(value float64, precision int) float64 {
	multiplier := math.Pow10(precision)
	return math.Round(value*multiplier) / multiplier
}

func FormatDateFullThai(date time.Time) string {
	// แปลงเป็นปีไทย พ.ศ. เช่น 1 มกราคม 2564
	thaiYear := date.Year() + 543

	// กำหนดชื่อเดือนภาษาไทย
	thaiMonths := []string{
		"มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน",
		"กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม",
	}

	// จัดรูปแบบวันที่เป็น "วันที่ เดือน พ.ศ."
	formattedDate := fmt.Sprintf("%d %s %d", date.Day(), thaiMonths[date.Month()-1], thaiYear)

	return formattedDate
}

func FormatDateFull(date time.Time) string {
	// แปลงเป็นปีไทย พ.ศ. เช่น 1 มกราคม 2564
	thaiYear := (date.Year() + 543) - 2500

	// จัดรูปแบบวันที่เป็น "วันที่ เดือน พ.ศ."
	formattedDate := fmt.Sprintf("%02d/%02d/%02d", date.Day(), date.Month(), thaiYear)

	return formattedDate
}

func FormatNumber(number float64, point int) string {
	if number == 0 || math.IsNaN(number) {
		return ""
	}

	// แยกส่วนจำนวนเต็มกับทศนิยม
	format := fmt.Sprintf("%%.%df", point)
	numStr := fmt.Sprintf(format, number)

	// แยกส่วนจำนวนเต็มออกมา
	parts := strings.Split(numStr, ".")
	integerPart := parts[0]

	// เพิ่มคอมมาในส่วนจำนวนเต็ม
	var result string
	for i, digit := range integerPart {
		if i > 0 && (len(integerPart)-i)%3 == 0 && digit != '-' {
			result += ","
		}
		result += string(digit)
	}

	// ถ้ามีทศนิยม ให้ต่อส่วนทศนิยมกลับเข้าไป
	if len(parts) > 1 {
		result += "." + parts[1]
	}

	return result
}

func FormatNumberWithOptionalDecimal(number float64, point int) string {
	if number == 0 || math.IsNaN(number) {
		return ""
	}

	// แยกส่วนจำนวนเต็มกับทศนิยม
	format := fmt.Sprintf("%%.%df", point)
	numStr := fmt.Sprintf(format, number)

	// แยกส่วนจำนวนเต็มออกมา
	parts := strings.Split(numStr, ".")
	integerPart := parts[0]

	// เพิ่มคอมมาในส่วนจำนวนเต็ม
	var result string
	for i, digit := range integerPart {
		if i > 0 && (len(integerPart)-i)%3 == 0 && digit != '-' {
			result += ","
		}
		result += string(digit)
	}

	// ถ้ามีทศนิยม ให้ตรวจสอบว่าเป็นจำนวนเต็มหรือไม่
	if len(parts) > 1 {
		decimalPart := parts[1]
		// ตรวจสอบว่าทศนิยมเป็น 0 ทั้งหมดหรือไม่
		isZeroDecimal := true
		for _, digit := range decimalPart {
			if digit != '0' {
				isZeroDecimal = false
				break
			}
		}
		// ถ้าทศนิยมไม่ใช่ 0 ให้ต่อส่วนทศนิยมกลับเข้าไป
		if !isZeroDecimal {
			result += "." + decimalPart
		}
	}

	return result
}

func TransFlagName(transFlag int, qty float64) string {
	name := ""
	switch transFlag {
	case 12:
		name = "ซื้อ"
	case 16:
		name = "ส่งคืน"
	case 44:
		name = "ขาย"
	case 48:
		name = "รับคืน"
	case 56:
		name = "เบิก"
	case 60:
		name = "รับ"
	case 54:
		name = "ยกมา"
	case 58:
		name = "รับคืนจากเบิก"
	case 66:
		name = "ปรับปรุงเพิ่ม"
	case 68:
		name = "ปรับปรุงลด"
	case 72:
		if qty > 0 {
			name = "โอนเข้า"
		} else {
			name = "โอนออก"
		}
	case 310:
		name = "รับ (พาเชียล)"
	case 866:
		name = "ปรับปรุงต้นทุน (เพิ่ม)"
	case 868:
		name = "ปรับปรุงต้นทุน (ลด)"
	default:
		logger.Info("Unknown transFlag: %d", transFlag)
		return "Unknown (transFlag: " + strconv.Itoa(transFlag) + ")"
	}
	return name + " " + FormatNumberWithOptionalDecimal(qty, 2)
}

func ConvertPDFToGzip(inputPath, outputPath string) error {
	// เปิดไฟล์ input
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("ไม่สามารถเปิดไฟล์ input: %v", err)
	}
	defer inputFile.Close()

	// สร้างไฟล์ output
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("ไม่สามารถสร้างไฟล์ output: %v", err)
	}
	defer outputFile.Close()

	// สร้าง gzip writer
	gzipWriter := gzip.NewWriter(outputFile)
	defer gzipWriter.Close()

	// คัดลอกข้อมูลจาก input ไปยัง gzip writer
	_, err = io.Copy(gzipWriter, inputFile)
	if err != nil {
		return fmt.Errorf("ไม่สามารถคัดลอกข้อมูล: %v", err)
	}

	// ปิด gzip writer เพื่อให้แน่ใจว่าข้อมูลทั้งหมดถูกเขียน
	err = gzipWriter.Close()
	if err != nil {
		return fmt.Errorf("ไม่สามารถปิด gzip writer: %v", err)
	}

	return nil
}

func DeleteFilesInFolder(folderPath string, beginName string) error {
	// สร้าง pattern เพื่อหาทุกไฟล์ใน folderPath
	pattern := filepath.Join(folderPath, "*")
	logger.Info("ค้นหาไฟล์ทั้งหมดด้วย pattern: %s", pattern)

	// ค้นหาทุกไฟล์ใน folder
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	// ตรวจสอบว่าพบไฟล์หรือไม่
	if len(files) == 0 {
		logger.Info("ไม่พบไฟล์ใน %s", folderPath)
		return nil // ไม่ถือเป็น error ถ้าไม่พบไฟล์
	}

	currentTime := time.Now()
	deletedCount := 0

	// ลบไฟล์แต่ละไฟล์
	for _, file := range files {
		// ตรวจสอบว่าเป็นไฟล์จริงๆ ไม่ใช่ directory
		fileInfo, err := os.Stat(file)
		if err != nil {
			logger.Info("เกิดข้อผิดพลาดในการอ่านข้อมูลไฟล์ %s: %v", file, err)
			continue // ข้ามไปถ้ามี error
		}

		// ข้ามถ้าเป็นโฟลเดอร์
		if fileInfo.IsDir() {
			continue
		}

		// ตรวจสอบว่าไฟล์มีคำนำหน้าตามที่กำหนดหรือไม่
		fileName := filepath.Base(file)
		if beginName != "" && !strings.HasPrefix(fileName, beginName) {
			continue
		}

		// คำนวณอายุของไฟล์
		fileAge := currentTime.Sub(fileInfo.ModTime())
		fileAgeHours := fileAge.Hours()

		// ลบไฟล์ที่มีอายุเกินกำหนด
		if fileAgeHours > 24 {
			logger.Info("กำลังลบไฟล์ที่มีอายุ %.2f ชั่วโมง: %s", fileAgeHours, file)
			err = os.Remove(file)
			if err != nil {
				logger.Info("ล้มเหลวในการลบไฟล์ %s: %v", file, err)
				return err
			}
			deletedCount++
		}
	}

	logger.Info("ลบไฟล์เก่าทั้งหมด %d ไฟล์เรียบร้อยแล้ว", deletedCount)
	return nil
}

func ConvertStringToInt(str string) (int, error) {
	// แปลงสตริงเป็นจำนวนเต็ม
	result, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("ไม่สามารถแปลงสตริงเป็นจำนวนเต็มได้: %v", err)
	}
	return result, nil
}

func CalcStockQtyWord(qty float64, barcodePacking []models.ProductBarcodePackingStruct) string {
	calcBalanceWord := ""
	calcBalanceQty := qty
	// ถ้าติดลบ
	if calcBalanceQty < 0 {
		calcBalanceQty = calcBalanceQty * -1
		calcBalanceWord = "(ลบ) "
	}
	for index, packing := range barcodePacking {
		// logger.Info("UnitName: %s, BarcodeRefUnitStand: %f, BarcodeRefUnitDivide: %f", packing.UnitName, packing.BarcodeRefUnitStand, packing.BarcodeRefUnitDivide)
		// หาจำนวนเต็ม
		if packing.BarcodeRefUnitDivide == 0 {
			calcBalanceQty = 0
			calcBalanceWord = "Error"
			break
		}
		// หาจำนวนเต็ม
		calcQty := 0.0
		if packing.BarcodeRefUnitDivide > packing.BarcodeRefUnitStand {
			calcQty = calcBalanceQty * packing.BarcodeRefUnitDivide
		} else {
			calcQty = calcBalanceQty / packing.BarcodeRefUnitStand
		}
		if index != len(barcodePacking)-1 {
			// ปัดเศษลง
			calcQty = math.Floor(calcQty)
		}
		// logger.Info("calcQty: %f %f %f", calcQty, packing.BarcodeRefUnitStand, packing.BarcodeRefUnitDivide)
		if calcQty > 0 {
			if calcBalanceWord != "" {
				calcBalanceWord = calcBalanceWord + "/"
			}
			calcBalanceWord = calcBalanceWord + fmt.Sprintf("%s %s", CalcStockQtyWordCut(calcQty), packing.UnitName)
		}
		// หาเศษ
		calcBalanceQty = calcBalanceQty - (calcQty * packing.BarcodeRefUnitStand)
		// logger.Info("calcBalanceQty: %f", calcBalanceQty)
	}
	return calcBalanceWord
}

func CalcStockQtyWordCut(calcQty float64) string {
	// หาทศนิยมกี่ตำแหน่ง
	qtyWord := fmt.Sprintf("%f", calcQty)
	// ลบ 0 ท้ายสุดมาเรื่อยๆ จนกว่าจะเจอ . และลบ . แต่ถ้าเจอตัวเลข 1-9 ให้หยุด
	for {
		if qtyWord[len(qtyWord)-1:] == "0" {
			qtyWord = qtyWord[:len(qtyWord)-1]
		} else if qtyWord[len(qtyWord)-1:] == "." {
			qtyWord = qtyWord[:len(qtyWord)-1]
			break
		} else {
			break
		}
	}
	return qtyWord
}

func SafeFloatConversion(value any) float64 {
	if value == nil {
		return 0.0
	}

	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0.0
}

// ParseFloat64 แปลงค่าจาก any เป็น float64
func ParseFloat64(val any, defaultValue float64) float64 {
	if val == nil {
		return defaultValue
	}

	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case int32:
		return float64(v)
	case int16:
		return float64(v)
	case int8:
		return float64(v)
	case uint:
		return float64(v)
	case uint64:
		return float64(v)
	case uint32:
		return float64(v)
	case uint16:
		return float64(v)
	case uint8:
		return float64(v)
	case string:
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			return parsed
		}
	case []byte:
		if parsed, err := strconv.ParseFloat(string(v), 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// ParseInt แปลงค่าจาก any เป็น int
func ParseInt(val any, defaultValue int) int {
	if val == nil {
		return defaultValue
	}

	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case int32:
		return int(v)
	case int16:
		return int(v)
	case int8:
		return int(v)
	case uint:
		return int(v)
	case uint64:
		return int(v)
	case uint32:
		return int(v)
	case uint16:
		return int(v)
	case uint8:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	case string:
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	case []byte:
		if parsed, err := strconv.Atoi(string(v)); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// ParseString แปลงค่าจาก any เป็น string
func ParseString(val any, defaultValue string) string {
	if val == nil {
		return defaultValue
	}

	switch v := val.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8:
		return fmt.Sprintf("%d", v)
	case float64, float32:
		return fmt.Sprintf("%g", v)
	case bool:
		return strconv.FormatBool(v)
	}
	return fmt.Sprintf("%v", val)
}

func ParseNumericValue(value any, defaultValue float64) float64 {
	if value == nil {
		return defaultValue
	}

	switch val := value.(type) {
	case float64:
		return val
	case string:
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			return parsed
		}
	case int64:
		return float64(val)
	case int:
		return float64(val)
	}
	return defaultValue
}

func StockTransactionList() []models.StockTransactionStruct {
	result := []models.StockTransactionStruct{
		{Name: "transactionPurchaseOrder", Flags: []int{0, 6}},     // สั่งซื้อ
		{Name: "transactionPurchasepartial", Flags: []int{0, 310}}, // รับบางส่วน (พาเชียล)
		{Name: "transactionPurchase", Flags: []int{0, 12}},         // ซื้อ (ตั้งหนี้/พร้อมสินค้า)

		{Name: "transactionSaleOrder", Flags: []int{0, 36}},         // สั่งขาย
		{Name: "transactionSaleInvoice", Flags: []int{0, 44}},       // ขาย
		{Name: "transactionSaleInvoiceReturn", Flags: []int{0, 48}}, // รับคืน
		{Name: "transactionPurchaseReturn", Flags: []int{0, 16}},    // ส่งคืน

		{Name: "transactionStockTransfer", Flags: []int{72}},       // โอน
		{Name: "transactionStockReceiveProduct", Flags: []int{60}}, // รับเข้า

		{Name: "transactionStockBalance", Flags: []int{54}},

		{Name: "transactionStockPickupProduct", Flags: []int{56}}, // เบิกออก
		{Name: "transactionStockReturnProduct", Flags: []int{58}}, // รับคืนจากเบิก

		{Name: "transactionStockAdjustment", Flags: []int{66, 68, 866, 868}}, // ปรับปรุง
	}

	return result
}
