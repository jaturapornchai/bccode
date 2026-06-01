package approval

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"smlcloudplatform/internal/goapi/logger"
	"time"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Collections
// NOTE: purchase_types ถูกย้ายไป mainapi แล้ว - ใช้ purchaseTypes collection ใน mainapi แทน
const (
	POApprovalSettingsCollection = "po_approval_settings"
)

var (
	atlasClient *mongo.Client
	atlasDB     *mongo.Database
	// MongoDB สำหรับ approval tokens (ใช้ร่วมกับ lineoa-liff)
	tokenAtlasClient *mongo.Client
	tokenAtlasDB     *mongo.Database
)

// Init initializes the approval handlers with MongoDB connection
func Init(client *mongo.Client, db *mongo.Database) {
	atlasClient = client
	atlasDB = db

	// เชื่อมต่อ MongoDB สำหรับ tokens (ใช้ร่วมกับ lineoa-liff)
	initTokenAtlas()
}

// initTokenAtlas เชื่อมต่อ MongoDB สำหรับ approval tokens
func initTokenAtlas() {
	tokenURI := firstNonEmptyEnv("MONGODB_TOKEN_URI", "MONGODB_TOKEN_ATLAS_URI")
	if tokenURI == "" {
		logger.Warn("[TokenAtlas] MONGODB_TOKEN_URI not set, tokens will use main MongoDB connection")
		tokenAtlasClient = atlasClient
		tokenAtlasDB = atlasDB
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	clientOptions := options.Client().ApplyURI(tokenURI)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		logger.Error("[TokenAtlas] Failed to connect: %v", err)
		// Fallback to main MongoDB connection.
		tokenAtlasClient = atlasClient
		tokenAtlasDB = atlasDB
		return
	}

	// Ping to verify connection
	if err := client.Ping(ctx, nil); err != nil {
		logger.Error("[TokenAtlas] Failed to ping: %v", err)
		tokenAtlasClient = atlasClient
		tokenAtlasDB = atlasDB
		return
	}

	tokenAtlasClient = client

	// Database name from env or default
	dbName := firstNonEmptyEnv("MONGODB_TOKEN_DB", "MONGODB_TOKEN_DBNAME", "MONGODB_TOKEN_ATLAS_DBNAME")
	if dbName == "" {
		dbName = "bcai_documents"
	}
	tokenAtlasDB = client.Database(dbName)

	logger.Success("[TokenAtlas] ✅ Connected to MongoDB for tokens (DB: %s)", dbName)
}

func firstNonEmptyEnv(keys ...string) string {
	for _, key := range keys {
		if value := os.Getenv(key); value != "" {
			return value
		}
	}
	return ""
}

// getTokenCollection returns a MongoDB collection for approval tokens.
func getTokenCollection(name string) *mongo.Collection {
	if tokenAtlasDB == nil {
		return nil
	}
	return tokenAtlasDB.Collection(name)
}

// IsConnected returns true if MongoDB is connected
func IsConnected() bool {
	return atlasClient != nil && atlasDB != nil
}

// getCollection returns a MongoDB collection
func getCollection(name string) *mongo.Collection {
	if atlasDB == nil {
		return nil
	}
	return atlasDB.Collection(name)
}

// getUserUID ค้นหา uid จาก users ด้วย username
func getUserUID(ctx context.Context, username string) string {
	if atlasDB == nil || username == "" {
		return ""
	}
	var doc struct {
		UID string `bson:"uid"`
	}
	err := atlasDB.Collection("users").FindOne(ctx, bson.M{"username": username}).Decode(&doc)
	if err != nil {
		return ""
	}
	return doc.UID
}

// checkConnection returns error response if not connected
func checkConnection(c echo.Context) error {
	if !IsConnected() {
		return c.JSON(http.StatusServiceUnavailable, map[string]any{
			"success": false,
			"message": "MongoDB is not connected",
		})
	}
	return nil
}

// NOTE: Purchase Type Models ถูกย้ายไป mainapi แล้ว
// ใช้ purchaseTypes collection และ /purchase-type endpoints ใน mainapi แทน

// =====================================================
// PO Approval Setting Models
// =====================================================

// ApproverInfo represents approver information in each rule
type ApproverInfo struct {
	UserCode        string `bson:"user_code" json:"user_code"`
	UserName        string `bson:"user_name" json:"user_name"`
	Email           string `bson:"email,omitempty" json:"email,omitempty"`
	LineUserID      string `bson:"line_user_id,omitempty" json:"line_user_id,omitempty"`
	LineDisplayName string `bson:"line_display_name,omitempty" json:"line_display_name,omitempty"`
	Position        string `bson:"position,omitempty" json:"position,omitempty"`
	Department      string `bson:"department,omitempty" json:"department,omitempty"`
}

// ApprovalRule represents an approval rule
type ApprovalRule struct {
	MinAmount         float64        `bson:"min_amount" json:"min_amount"`
	MaxAmount         float64        `bson:"max_amount" json:"max_amount"`
	ApprovalLevel     int            `bson:"approval_level" json:"approval_level"`
	ApprovalLevelName string         `bson:"approval_level_name" json:"approval_level_name"`
	Approvers         []ApproverInfo `bson:"approvers" json:"approvers"` // รายการผู้มีสิทธิ์อนุมัติ
}

// POApprovalSetting represents a PO approval setting document
type POApprovalSetting struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"guid,omitempty"`
	ShopID           string             `bson:"shop_id" json:"shop_id"`
	PurchaseTypeCode string             `bson:"purchase_type_code" json:"purchase_type_code"`
	PurchaseTypeName string             `bson:"purchase_type_name" json:"purchase_type_name"`
	Rules            []ApprovalRule     `bson:"rules" json:"rules"`
	IsActive         bool               `bson:"is_active" json:"is_active"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	UpdatedAt        time.Time          `bson:"updated_at" json:"updated_at"`
}

// POApprovalSettingRequest represents a request for PO approval setting operations
type POApprovalSettingRequest struct {
	ShopID           string         `json:"shop_id"`
	PurchaseTypeCode string         `json:"purchase_type_code"`
	PurchaseTypeName string         `json:"purchase_type_name"`
	Rules            []ApprovalRule `json:"rules"`
	IsActive         *bool          `json:"is_active"`
	Action           string         `json:"action"` // list, get, save, delete
}

// NOTE: Purchase Type Handlers ถูกย้ายไป mainapi แล้ว
// ใช้ /purchase-type endpoints ใน mainapi แทน

// =====================================================
// PO Approval Setting Handlers
// =====================================================

// GetPOApprovalSettingsHandler - Get all PO approval settings for a shop
func GetPOApprovalSettingsHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req POApprovalSettingRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.ShopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "shop_id is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := getCollection(POApprovalSettingsCollection)
	filter := bson.M{"shop_id": req.ShopID}

	cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.M{"purchase_type_code": 1}))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Query execution failed",
		})
	}
	defer cursor.Close(ctx)

	var items []POApprovalSetting
	if err := cursor.All(ctx, &items); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Failed to decode results",
		})
	}

	if items == nil {
		items = []POApprovalSetting{}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"data":    items,
	})
}

// GetPOApprovalSettingHandler - Get single PO approval setting by purchase type code
func GetPOApprovalSettingHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req POApprovalSettingRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.ShopID == "" || req.PurchaseTypeCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "shop_id and purchase_type_code are required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := getCollection(POApprovalSettingsCollection)
	filter := bson.M{"shop_id": req.ShopID, "purchase_type_code": req.PurchaseTypeCode}

	var setting POApprovalSetting
	err := collection.FindOne(ctx, filter).Decode(&setting)
	if err == mongo.ErrNoDocuments {
		return c.JSON(http.StatusOK, map[string]any{
			"success": true,
			"found":   false,
			"message": "PO approval setting not found",
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Query execution failed",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"found":   true,
		"data":    setting,
	})
}

// SavePOApprovalSettingHandler - Save or update a PO approval setting
func SavePOApprovalSettingHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req POApprovalSettingRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.ShopID == "" || req.PurchaseTypeCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "shop_id and purchase_type_code are required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := getCollection(POApprovalSettingsCollection)
	now := time.Now().UTC() // เก็บเวลาเป็น UTC+0

	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	rules := req.Rules
	if rules == nil {
		rules = []ApprovalRule{}
	}

	filter := bson.M{"shop_id": req.ShopID, "purchase_type_code": req.PurchaseTypeCode}
	update := bson.M{
		"$set": bson.M{
			"purchase_type_name": req.PurchaseTypeName,
			"rules":              rules,
			"is_active":          isActive,
			"updated_at":         now,
		},
		"$setOnInsert": bson.M{
			"shop_id":            req.ShopID,
			"purchase_type_code": req.PurchaseTypeCode,
			"created_at":         now,
		},
	}

	opts := options.Update().SetUpsert(true)
	result, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Failed to save PO approval setting",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"message": "บันทึกการตั้งค่าอนุมัติสำเร็จ",
		"data": map[string]any{
			"matched":  result.MatchedCount,
			"modified": result.ModifiedCount,
			"upserted": result.UpsertedCount,
		},
	})
}

// DeletePOApprovalSettingHandler - Delete a PO approval setting
func DeletePOApprovalSettingHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req POApprovalSettingRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.ShopID == "" || req.PurchaseTypeCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "shop_id and purchase_type_code are required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := getCollection(POApprovalSettingsCollection)
	filter := bson.M{"shop_id": req.ShopID, "purchase_type_code": req.PurchaseTypeCode}

	result, err := collection.DeleteOne(ctx, filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Failed to delete PO approval setting",
		})
	}

	if result.DeletedCount == 0 {
		return c.JSON(http.StatusNotFound, map[string]any{
			"success": false,
			"message": "ไม่พบข้อมูลที่ต้องการลบ",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"message": "ลบข้อมูลสำเร็จ",
	})
}

// =====================================================
// PO Approval Status (สถานะการอนุมัติ PO)
// =====================================================
// NOTE: purchase_type_code และ purchase_type_names ย้ายไปเก็บใน TransactionHeader แล้ว
// (ไม่ใช้ po_document_types collection อีกต่อไป)

// Collection สำหรับเก็บสถานะการอนุมัติ PO
const POApprovalStatusCollection = "po_approval_status"

// ApprovalHistory ประวัติการอนุมัติ
type ApprovalHistory struct {
	Action          string    `bson:"action" json:"action"`                       // submit, approve, reject
	ActionBy        string    `bson:"action_by" json:"action_by"`                 // รหัสผู้ดำเนินการ
	ActionByName    string    `bson:"action_by_name" json:"action_by_name"`       // ชื่อผู้ดำเนินการ
	ApproverUserUID string    `bson:"approver_user_uid,omitempty" json:"approver_user_uid,omitempty"` // UID ของผู้อนุมัติ
	Level           int       `bson:"level" json:"level"`                         // ระดับการอนุมัติ
	Comment         string    `bson:"comment,omitempty" json:"comment,omitempty"` // หมายเหตุ
	ActionAt        time.Time `bson:"action_at" json:"action_at"`                 // เวลาดำเนินการ (UTC)
	Source          string    `bson:"source,omitempty" json:"source,omitempty"`   // ช่องทาง: line, email, app
}

// POApprovalStatusItem รายการสินค้าใน PO สำหรับแสดงใน LIFF
type POApprovalStatusItem struct {
	LineNumber int     `bson:"line_number" json:"line_number"`
	ItemCode   string  `bson:"itemcode" json:"itemcode"`
	ItemName   string  `bson:"item_name" json:"item_name"`
	Qty        float64 `bson:"qty" json:"qty"`
	UnitName   string  `bson:"unit_name" json:"unit_name"`
	Price      float64 `bson:"price" json:"price"`
	SumAmount  float64 `bson:"sum_amount" json:"sum_amount"`
}

// POApprovalStatus สถานะการอนุมัติ PO
type POApprovalStatus struct {
	ID                   primitive.ObjectID `bson:"_id,omitempty" json:"guid,omitempty"`
	ShopID               string             `bson:"shop_id" json:"shop_id"`
	DocNo                string             `bson:"docno" json:"docno"`
	GuidFixed            string             `bson:"guid_fixed" json:"guid_fixed"`
	SourceDocNo          string             `bson:"source_docno,omitempty" json:"source_docno,omitempty"`         // เลขที่เอกสารต้นแบบ
	SourceGuidFixed      string             `bson:"source_guidfixed,omitempty" json:"source_guidfixed,omitempty"` // GUID ต้นแบบ
	PurchaseTypeCode     string             `bson:"purchase_type_code" json:"purchase_type_code"`
	PurchaseTypeName     string             `bson:"purchase_type_name" json:"purchase_type_name"`
	TotalAmount          float64            `bson:"total_amount" json:"total_amount"`
	RequiredLevel        int                `bson:"required_level" json:"required_level"`
	RequiredLevelName    string             `bson:"required_level_name" json:"required_level_name"`
	CurrentApprovedLevel int                `bson:"current_approved_level" json:"current_approved_level"`
	Status               string             `bson:"status" json:"status"` // draft, auto_approved, pending, approved, rejected
	CreatedBy            string             `bson:"created_by" json:"created_by"`
	CreatedByName        string             `bson:"created_by_name" json:"created_by_name"`
	History              []ApprovalHistory  `bson:"history" json:"history"`
	LastComment          string             `bson:"last_comment,omitempty" json:"last_comment,omitempty"`
	// Transaction details for LIFF display (copy from transactiondb)
	DocDatetime string                 `bson:"docdatetime,omitempty" json:"docdatetime,omitempty"`
	CustCode    string                 `bson:"custcode,omitempty" json:"custcode,omitempty"`
	CustName    string                 `bson:"cust_name,omitempty" json:"cust_name,omitempty"`
	Items       []POApprovalStatusItem `bson:"items,omitempty" json:"items,omitempty"`
	CreatedAt   time.Time              `bson:"created_at" json:"created_at"`
	UpdatedAt   time.Time              `bson:"updated_at" json:"updated_at"`
}

// POApprovalStatusRequestItem รายการสินค้าใน request
type POApprovalStatusRequestItem struct {
	LineNumber int     `json:"line_number"`
	ItemCode   string  `json:"itemcode"`
	ItemName   string  `json:"item_name"`
	Qty        float64 `json:"qty"`
	UnitName   string  `json:"unit_name"`
	Price      float64 `json:"price"`
	SumAmount  float64 `json:"sum_amount"`
}

// POApprovalStatusRequest request สำหรับจัดการสถานะการอนุมัติ
type POApprovalStatusRequest struct {
	ShopID           string  `json:"shop_id"`
	DocNo            string  `json:"docno"`
	GuidFixed        string  `json:"guid_fixed"`
	SourceDocNo      string  `json:"source_docno"`     // เลขที่เอกสารต้นแบบ
	SourceGuidFixed  string  `json:"source_guidfixed"` // GUID ต้นแบบ
	PurchaseTypeCode string  `json:"purchase_type_code"`
	PurchaseTypeName string  `json:"purchase_type_name"`
	TotalAmount      float64 `json:"total_amount"`
	ActionBy         string  `json:"action_by"`
	ActionByName     string  `json:"action_by_name"`
	ApprovalLevel    int     `json:"approval_level"` // ระดับสิทธิ์ของผู้กระทำ
	Comment          string  `json:"comment"`        // หมายเหตุ
	IsModified       bool    `json:"is_modified"`    // เป็นการแก้ไขเอกสารหรือไม่
	Source           string  `json:"source"`         // ช่องทาง: line, email, app
	// Transaction details for LIFF display
	DocDatetime string                        `json:"docdatetime,omitempty"`
	CustCode    string                        `json:"custcode,omitempty"`
	CustName    string                        `json:"cust_name,omitempty"`
	Items       []POApprovalStatusRequestItem `json:"items,omitempty"`
}

// GetPOApprovalStatusHandler - ดึงสถานะการอนุมัติ PO
func GetPOApprovalStatusHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req POApprovalStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.ShopID == "" || req.DocNo == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "shop_id and docno are required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := getTokenCollection(POApprovalStatusCollection)
	filter := bson.M{"shop_id": req.ShopID, "docno": req.DocNo}

	var status POApprovalStatus
	err := collection.FindOne(ctx, filter).Decode(&status)
	if err == mongo.ErrNoDocuments {
		return c.JSON(http.StatusOK, map[string]any{
			"success": true,
			"found":   false,
			"message": "ไม่พบสถานะการอนุมัติของเอกสารนี้",
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Query execution failed",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"found":   true,
		"data":    status,
	})
}

// GetBatchPOApprovalStatusRequest request สำหรับดึงสถานะหลายเอกสาร
type GetBatchPOApprovalStatusRequest struct {
	ShopID string   `json:"shop_id"`
	DocNos []string `json:"docnos"`
}

// GetBatchPOApprovalStatusHandler - ดึงสถานะการอนุมัติ PO หลายเอกสารพร้อมกัน
func GetBatchPOApprovalStatusHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req GetBatchPOApprovalStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.ShopID == "" || len(req.DocNos) == 0 {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "shop_id and docnos are required",
		})
	}

	// จำกัดจำนวนเอกสารสูงสุด 100 รายการ
	if len(req.DocNos) > 100 {
		req.DocNos = req.DocNos[:100]
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := getTokenCollection(POApprovalStatusCollection)
	filter := bson.M{
		"shop_id": req.ShopID,
		"docno":   bson.M{"$in": req.DocNos},
	}

	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Query execution failed",
		})
	}
	defer cursor.Close(ctx)

	// สร้าง map docno -> status
	statusMap := make(map[string]map[string]any)
	docNoList := []string{} // เก็บ docno ที่มี status
	for cursor.Next(ctx) {
		var status POApprovalStatus
		if err := cursor.Decode(&status); err != nil {
			continue
		}
		// หา latest event จาก history
		var latestAction string
		var latestActionText string
		var latestActionAt time.Time
		var latestActionBy string
		if len(status.History) > 0 {
			lastHistory := status.History[len(status.History)-1]
			latestAction = lastHistory.Action
			latestActionAt = lastHistory.ActionAt
			latestActionBy = lastHistory.ActionByName
			// แปลง action เป็นข้อความภาษาไทย
			switch lastHistory.Action {
			case "submit":
				latestActionText = "ส่งขออนุมัติ"
			case "approve":
				latestActionText = "อนุมัติแล้ว"
			case "reject":
				latestActionText = "ไม่อนุมัติ"
			case "resubmit":
				latestActionText = "ส่งใหม่"
			default:
				latestActionText = lastHistory.Action
			}
		}
		statusMap[status.DocNo] = map[string]any{
			"status":                 status.Status,
			"required_level":         status.RequiredLevel,
			"required_level_name":    status.RequiredLevelName,
			"current_approved_level": status.CurrentApprovedLevel,
			"latest_action":          latestAction,
			"latest_action_text":     latestActionText,
			"latest_action_at":       latestActionAt,
			"latest_action_by":       latestActionBy,
		}
		docNoList = append(docNoList, status.DocNo)
	}

	// ดึงข้อมูล notification log ล่าสุดสำหรับแต่ละเอกสาร
	if len(docNoList) > 0 {
		notificationCollection := getTokenCollection(POApprovalNotificationLogCollection)
		notificationFilter := bson.M{
			"shop_id": req.ShopID,
			"docno":   bson.M{"$in": docNoList},
		}
		notificationCursor, err := notificationCollection.Find(ctx, notificationFilter, options.Find().SetSort(bson.D{{Key: "sent_at", Value: -1}}))
		if err == nil {
			defer notificationCursor.Close(ctx)
			// เก็บ notification ล่าสุดสำหรับแต่ละ docno
			notificationMap := make(map[string]NotificationLog)
			for notificationCursor.Next(ctx) {
				var notif NotificationLog
				if err := notificationCursor.Decode(&notif); err != nil {
					continue
				}
				// เก็บเฉพาะ notification ล่าสุดของแต่ละ docno
				if _, exists := notificationMap[notif.DocNo]; !exists {
					notificationMap[notif.DocNo] = notif
				}
			}
			// เพิ่มข้อมูล notification ลง statusMap
			for docNo, notif := range notificationMap {
				if statusData, exists := statusMap[docNo]; exists {
					statusData["notification_sent_at"] = notif.SentAt
					statusData["notification_opened_at"] = notif.OpenedAt
					statusData["notification_type"] = notif.NotificationType
					statusData["notification_approver"] = notif.ApproverName
					// ถ้ามี notification ส่งแล้ว ให้เพิ่ม latest_action_text
					if !notif.SentAt.IsZero() {
						var notifText string
						if notif.OpenedAt != nil {
							notifText = "เปิดอ่านแล้ว"
							statusData["latest_action_text"] = notifText
							statusData["latest_action_at"] = *notif.OpenedAt
							statusData["latest_action_by"] = notif.ApproverName
						} else {
							notifText = "ส่งการแจ้งเตือนแล้ว"
							// อัพเดท latest_action_text ถ้า notification ใหม่กว่า history
							if currentLatest, ok := statusData["latest_action_at"].(time.Time); !ok || notif.SentAt.After(currentLatest) {
								statusData["latest_action_text"] = notifText
								statusData["latest_action_at"] = notif.SentAt
								statusData["latest_action_by"] = notif.ApproverName
							}
						}
					}
				}
			}
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"data":    statusMap,
	})
}

// SubmitPOApprovalHandler - ส่ง PO เพื่อขออนุมัติ
func SubmitPOApprovalHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req POApprovalStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	// Log เพื่อ debug ค่าที่รับมาจาก Frontend
	logger.Info("[SubmitPO] Request received - docNo: %s, shopID: %s, is_modified: %v, totalAmount: %.2f",
		req.DocNo, req.ShopID, req.IsModified, req.TotalAmount)
	logger.Info("[SubmitPO] PurchaseType - code: '%s', name: '%s'", req.PurchaseTypeCode, req.PurchaseTypeName)
	logger.Info("[SubmitPO] Transaction details - custCode: '%s', custName: '%s', items: %d",
		req.CustCode, req.CustName, len(req.Items))
	logger.Info("[SubmitPO] Comment (หมายเหตุ): '%s'", req.Comment)

	if req.ShopID == "" || req.DocNo == "" || req.PurchaseTypeCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "shop_id, docno, and purchase_type_code are required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	now := time.Now().UTC() // เก็บเวลาเป็น UTC+0

	// ดึงกฎการอนุมัติตามประเภทการจัดซื้อ
	settingsCollection := getCollection(POApprovalSettingsCollection)
	var setting POApprovalSetting
	err := settingsCollection.FindOne(ctx, bson.M{
		"shop_id":            req.ShopID,
		"purchase_type_code": req.PurchaseTypeCode,
	}).Decode(&setting)

	requiredLevel := 0
	requiredLevelName := ""

	// หาระดับการอนุมัติที่ต้องการตามวงเงิน
	if err == nil && setting.Rules != nil {
		for _, rule := range setting.Rules {
			// ตรวจสอบว่าวงเงินอยู่ในช่วงหรือไม่
			if req.TotalAmount >= rule.MinAmount {
				if rule.MaxAmount == 0 || req.TotalAmount <= rule.MaxAmount {
					requiredLevel = rule.ApprovalLevel
					requiredLevelName = rule.ApprovalLevelName
					break
				}
			}
		}
	}

	// กำหนดสถานะเริ่มต้น
	status := "pending"
	currentApprovedLevel := 0

	// ถ้าไม่มีกฎอนุมัติ หรือ ระดับ 0 = อนุมัติอัตโนมัติโดยผู้บันทึก
	if requiredLevel == 0 {
		status = "auto_approved"
		requiredLevelName = "อนุมัติอัตโนมัติ"
	} else if req.ApprovalLevel >= requiredLevel {
		// ผู้บันทึกมีสิทธิ์เพียงพอ = อนุมัติอัตโนมัติ
		status = "auto_approved"
		currentApprovedLevel = req.ApprovalLevel
	}

	// สร้างประวัติการส่งอนุมัติ
	actionType := "submit"
	if req.IsModified {
		actionType = "modify"
	}

	newHistoryEntry := ApprovalHistory{
		Action:          actionType,
		ActionBy:        req.ActionBy,
		ActionByName:    req.ActionByName,
		ApproverUserUID: getUserUID(ctx, req.ActionBy),
		Level:           req.ApprovalLevel,
		Comment:         req.Comment,
		ActionAt:        now,
	}

	collection := getTokenCollection(POApprovalStatusCollection)

	// ดึง history เดิม (ถ้ามี) เพื่อ append ไม่ใช่ overwrite
	var existingStatus POApprovalStatus
	existingFilter := bson.M{"shop_id": req.ShopID, "docno": req.DocNo}
	existingErr := collection.FindOne(ctx, existingFilter).Decode(&existingStatus)

	var history []ApprovalHistory
	if existingErr == nil && len(existingStatus.History) > 0 {
		// มีประวัติเดิม - append เข้าไป
		history = append(existingStatus.History, newHistoryEntry)
	} else {
		// ไม่มีประวัติ - สร้างใหม่
		history = []ApprovalHistory{newHistoryEntry}
	}

	// ถ้าอนุมัติอัตโนมัติ ให้เพิ่มประวัติการอนุมัติด้วย
	if status == "auto_approved" {
		history = append(history, ApprovalHistory{
			Action:          "auto_approve",
			ActionBy:        req.ActionBy,
			ActionByName:    req.ActionByName,
			ApproverUserUID: getUserUID(ctx, req.ActionBy),
			Level:           req.ApprovalLevel,
			Comment:         "อนุมัติอัตโนมัติ (ไม่มีกฎอนุมัติหรือผู้บันทึกมีสิทธิ์เพียงพอ)",
			ActionAt:        now,
		})
	}

	// Convert items from request to status items
	var items []POApprovalStatusItem
	for _, item := range req.Items {
		items = append(items, POApprovalStatusItem{
			LineNumber: item.LineNumber,
			ItemCode:   item.ItemCode,
			ItemName:   item.ItemName,
			Qty:        item.Qty,
			UnitName:   item.UnitName,
			Price:      item.Price,
			SumAmount:  item.SumAmount,
		})
	}

	// Upsert - ถ้ามีแล้วให้อัปเดต ถ้าไม่มีให้สร้างใหม่
	filter := bson.M{"shop_id": req.ShopID, "docno": req.DocNo}
	update := bson.M{
		"$set": bson.M{
			"guid_fixed":             req.GuidFixed,
			"source_docno":           req.SourceDocNo,
			"source_guidfixed":       req.SourceGuidFixed,
			"purchase_type_code":     req.PurchaseTypeCode,
			"purchase_type_name":     req.PurchaseTypeName,
			"total_amount":           req.TotalAmount,
			"required_level":         requiredLevel,
			"required_level_name":    requiredLevelName,
			"current_approved_level": currentApprovedLevel,
			"status":                 status,
			"created_by":             req.ActionBy,
			"created_by_name":        req.ActionByName,
			"history":                history,
			"last_comment":           req.Comment, // หมายเหตุจากใบสั่งซื้อ
			"updated_at":             now,
			// Transaction details for LIFF display
			"docdatetime": req.DocDatetime,
			"custcode":    req.CustCode,
			"cust_name":   req.CustName,
			"items":       items,
		},
		"$setOnInsert": bson.M{
			"shop_id":    req.ShopID,
			"docno":      req.DocNo,
			"created_at": now,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err = collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Failed to submit PO for approval",
		})
	}

	// ดึงข้อมูลที่บันทึกไว้
	var savedStatus POApprovalStatus
	collection.FindOne(ctx, filter).Decode(&savedStatus)

	// ส่งแจ้งเตือนอัตโนมัติถ้าสถานะเป็น pending หรือมีการแก้ไข
	// หมายเหตุ: ถ้า isModified=true จะส่งแจ้งเตือนใหม่เสมอ (รวมถึง auto_approved)
	notificationSent := 0
	shouldSendNotification := status == "pending" || req.IsModified
	logger.Info("[SubmitPO] Check notification - status: %s, isModified: %v, shouldSend: %v", status, req.IsModified, shouldSendNotification)

	if shouldSendNotification {
		logger.Info("[SubmitPO] Calling sendAutoNotifications with isModified: %v", req.IsModified)
		notificationSent = sendAutoNotifications(ctx, req.ShopID, req.DocNo, savedStatus, setting, req.IsModified)
		logger.Info("[SubmitPO] Notifications sent: %d", notificationSent)
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"message": func() string {
			if status == "auto_approved" {
				return "อนุมัติอัตโนมัติสำเร็จ"
			}
			return "ส่งขออนุมัติสำเร็จ"
		}(),
		"data":               savedStatus,
		"notifications_sent": notificationSent,
	})
}

// sendAutoNotifications ส่งแจ้งเตือนอัตโนมัติไปยังผู้อนุมัติ
// isModified = true จะส่งแจ้งเตือนใหม่แม้ว่าจะเคยส่งไปแล้ว และแสดงว่ามีการแก้ไข
func sendAutoNotifications(ctx context.Context, shopID, docNo string, poStatus POApprovalStatus, setting POApprovalSetting, isModified bool) int {
	logger.Info("[sendAutoNotifications] START - docNo: %s, shopID: %s, isModified: %v, requiredLevel: %d",
		docNo, shopID, isModified, poStatus.RequiredLevel)

	// หาผู้อนุมัติที่เกี่ยวข้องตามวงเงิน
	var approvers []ApproverInfo
	var matchedLevel int

	for _, rule := range setting.Rules {
		if poStatus.TotalAmount >= rule.MinAmount {
			if rule.MaxAmount == 0 || poStatus.TotalAmount <= rule.MaxAmount {
				// ถ้า isModified และเป็น auto_approved (requiredLevel=0) ให้หาผู้อนุมัติตามวงเงินแทน
				// มิฉะนั้นใช้ logic เดิม คือ match ตาม requiredLevel
				shouldMatch := false
				if isModified && poStatus.RequiredLevel == 0 {
					// กรณี auto_approved + แก้ไข: ส่งให้ผู้อนุมัติที่มี level สูงสุดตามวงเงิน
					if len(rule.Approvers) > 0 && rule.ApprovalLevel > matchedLevel {
						shouldMatch = true
					}
				} else if rule.ApprovalLevel == poStatus.RequiredLevel {
					// กรณีปกติ: match ตาม requiredLevel
					shouldMatch = true
				}

				if shouldMatch && len(rule.Approvers) > 0 {
					approvers = rule.Approvers
					matchedLevel = rule.ApprovalLevel
					logger.Info("[sendAutoNotifications] Matched rule - level: %d, minAmount: %.2f, maxAmount: %.2f, approverCount: %d",
						rule.ApprovalLevel, rule.MinAmount, rule.MaxAmount, len(rule.Approvers))
					if !isModified || poStatus.RequiredLevel != 0 {
						break // ถ้าไม่ใช่ auto_approved + แก้ไข ให้หยุดที่ match แรก
					}
				}
			}
		}
	}

	if len(approvers) == 0 {
		logger.Info("[sendAutoNotifications] No approvers found for requiredLevel: %d, totalAmount: %.2f",
			poStatus.RequiredLevel, poStatus.TotalAmount)
		return 0
	}

	logger.Info("[sendAutoNotifications] Found %d approvers for level %d (matchedLevel: %d)", len(approvers), poStatus.RequiredLevel, matchedLevel)

	// Log รายละเอียดผู้อนุมัติแต่ละคน
	for i, approver := range approvers {
		logger.Info("[sendAutoNotifications] Approver[%d]: code=%s, name=%s, email=%s, lineUserID=%s, lineDisplayName=%s",
			i, approver.UserCode, approver.UserName, approver.Email, approver.LineUserID, approver.LineDisplayName)
	}

	// ดึง LIFF ID สำหรับ LINE
	liffID, liffErr := getLiffID(shopID)
	if liffErr != nil {
		logger.Warn("[sendAutoNotifications] Failed to get LIFF ID for shop %s: %v", shopID, liffErr)
	}
	logger.Info("[sendAutoNotifications] LIFF ID for shop %s: '%s' (empty=%v)", shopID, liffID, liffID == "")

	// Base URL
	baseURL := os.Getenv("APP_BASE_URL")
	if baseURL == "" {
		baseURL = "https://erp.bcai.cloud"
	}

	// ใช้ getTokenCollection เพื่อเขียนไป bcai_documents ที่ LINE-LIFF อ่าน
	notificationCollection := getTokenCollection(POApprovalNotificationLogCollection)
	now := time.Now().UTC() // เก็บเวลาเป็น UTC+0
	totalAmount := fmt.Sprintf("%.2f", poStatus.TotalAmount)
	sentCount := 0

	for _, approver := range approvers {
		// ถ้าไม่ใช่การแก้ไข ให้ตรวจสอบว่าเคยส่งสำเร็จแล้วหรือยัง
		if !isModified {
			existingFilter := bson.M{
				"shop_id":       shopID,
				"docno":         docNo,
				"approver_code": approver.UserCode,
				"status":        "sent",
			}
			existingCount, _ := notificationCollection.CountDocuments(ctx, existingFilter)
			if existingCount > 0 {
				continue
			}
		}

		// ส่ง Email ถ้ามี (ใช้ Brevo API)
		if approver.Email != "" && isBrevoConfigured() {
			approveToken, _ := generateApprovalToken(shopID, docNo, poStatus.GuidFixed, approver.UserCode, approver.UserName, "approve")
			rejectToken, _ := generateApprovalToken(shopID, docNo, poStatus.GuidFixed, approver.UserCode, approver.UserName, "reject")

			approveURL := fmt.Sprintf("%s/api/approval/action?token=%s&action=approve", baseURL, approveToken)
			rejectURL := fmt.Sprintf("%s/api/approval/action?token=%s&action=reject", baseURL, rejectToken)
			pdfURL := fmt.Sprintf("%s/api/po/pdf?shop_id=%s&docno=%s", baseURL, shopID, docNo)

			emailErr := SendApprovalEmail(
				approver.Email,
				approver.UserName,
				shopID,
				docNo,
				totalAmount,
				poStatus.PurchaseTypeName,
				poStatus.CreatedByName,
				approveURL,
				rejectURL,
				pdfURL,
				isModified,
			)

			status := "sent"
			errMsg := ""
			if emailErr != nil {
				status = "failed"
				errMsg = emailErr.Error()
			} else {
				sentCount++
			}

			emailLog := NotificationLog{
				ShopID:           shopID,
				DocNo:            docNo,
				GuidFixed:        poStatus.GuidFixed,
				ApproverCode:     approver.UserCode,
				ApproverName:     approver.UserName,
				NotificationType: "email",
				RecipientEmail:   approver.Email,
				Status:           status,
				ErrorMessage:     errMsg,
				SentAt:           now,
				CreatedAt:        now,
			}
			notificationCollection.InsertOne(ctx, emailLog)
		}

		// ส่ง LINE Push ถ้ามี
		if approver.LineUserID != "" && liffID != "" {
			approveToken, _ := generateApprovalToken(shopID, docNo, poStatus.GuidFixed, approver.UserCode, approver.UserName, "approve")
			liffURL := fmt.Sprintf("https://liff.line.me/%s/approve?token=%s&shop_id=%s&docno=%s", liffID, approveToken, shopID, docNo)

			logger.Info("[sendAutoNotifications] Sending LINE Push to %s (lineUserID: %s) with isModified: %v",
				approver.UserName, approver.LineUserID, isModified)

			lineErr := SendLinePushApprovalV2(ApprovalNotificationParams{
				LineUserID:       approver.LineUserID,
				ShopID:           shopID,
				DocNo:            docNo,
				DocDatetime:      poStatus.DocDatetime,
				TotalAmount:      totalAmount,
				PurchaseTypeName: poStatus.PurchaseTypeName,
				CreatedByName:    poStatus.CreatedByName,
				CustName:         poStatus.CustName,
				LiffApproveURL:   liffURL,
				IsModified:       isModified,
				POComment:        poStatus.LastComment,
			})

			status := "sent"
			errMsg := ""
			if lineErr != nil {
				status = "failed"
				errMsg = lineErr.Error()
			} else {
				sentCount++
			}

			lineLog := NotificationLog{
				ShopID:           shopID,
				DocNo:            docNo,
				GuidFixed:        poStatus.GuidFixed,
				ApproverCode:     approver.UserCode,
				ApproverName:     approver.UserName,
				NotificationType: "line",
				RecipientLineID:  approver.LineUserID,
				Status:           status,
				ErrorMessage:     errMsg,
				SentAt:           now,
				CreatedAt:        now,
			}
			notificationCollection.InsertOne(ctx, lineLog)
		} else {
			// Log เหตุผลที่ไม่ส่ง LINE Push
			if approver.LineUserID == "" {
				logger.Warn("[sendAutoNotifications] SKIPPED LINE Push for %s: no line_user_id in approver data", approver.UserCode)
			}
			if liffID == "" {
				logger.Warn("[sendAutoNotifications] SKIPPED LINE Push for %s: no liff_id configured for shop %s", approver.UserCode, shopID)
			}
		}
	}

	return sentCount
}

// ApprovePOHandler - อนุมัติ PO (ใช้ FindOneAndUpdate atomic ป้องกัน race condition)
func ApprovePOHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req POApprovalStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.ShopID == "" || req.DocNo == "" || req.ActionBy == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "shop_id, docno, and action_by are required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := getTokenCollection(POApprovalStatusCollection)
	now := time.Now().UTC()

	// กำหนด source
	source := req.Source
	if source == "" {
		source = "app"
	}

	newHistory := ApprovalHistory{
		Action:          "approve",
		ActionBy:        req.ActionBy,
		ActionByName:    req.ActionByName,
		ApproverUserUID: getUserUID(ctx, req.ActionBy),
		Level:           req.ApprovalLevel,
		Comment:         req.Comment,
		ActionAt:        now,
		Source:          source,
	}

	// Atomic: FindOneAndUpdate กับ filter status=pending → ป้องกัน 2 คน approve พร้อมกัน
	atomicFilter := bson.M{
		"shop_id": req.ShopID,
		"docno":   req.DocNo,
		"status":  "pending",
	}
	update := bson.M{
		"$set": bson.M{
			"status":                 "approved",
			"current_approved_level": req.ApprovalLevel,
			"updated_at":             now,
		},
		"$push": bson.M{
			"history": newHistory,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedStatus POApprovalStatus
	err := collection.FindOneAndUpdate(ctx, atomicFilter, update, opts).Decode(&updatedStatus)

	if err == mongo.ErrNoDocuments {
		// ตรวจสอบว่าเอกสารมีอยู่จริงหรือไม่ และสถานะปัจจุบันคืออะไร
		var existing POApprovalStatus
		findErr := collection.FindOne(ctx, bson.M{"shop_id": req.ShopID, "docno": req.DocNo}).Decode(&existing)
		if findErr == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]any{
				"success": false,
				"message": "ไม่พบเอกสารนี้ในระบบอนุมัติ",
			})
		}
		// ตรวจสอบสิทธิ์ (กรณี status ถูกต้องแต่ไม่ match filter อื่น)
		if existing.Status != "pending" {
			return c.JSON(http.StatusBadRequest, map[string]any{
				"success": false,
				"message": "เอกสารนี้ไม่ได้อยู่ในสถานะรออนุมัติ (สถานะปัจจุบัน: " + existing.Status + ")",
			})
		}
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Failed to approve PO",
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Failed to approve PO",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"message": "อนุมัติเอกสารสำเร็จ",
		"data":    updatedStatus,
	})
}

// RejectPOHandler - ปฏิเสธ PO (ใช้ FindOneAndUpdate atomic ป้องกัน race condition)
func RejectPOHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req POApprovalStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.ShopID == "" || req.DocNo == "" || req.ActionBy == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "shop_id, docno, and action_by are required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := getTokenCollection(POApprovalStatusCollection)
	now := time.Now().UTC()

	source := req.Source
	if source == "" {
		source = "app"
	}

	newHistory := ApprovalHistory{
		Action:          "reject",
		ActionBy:        req.ActionBy,
		ActionByName:    req.ActionByName,
		ApproverUserUID: getUserUID(ctx, req.ActionBy),
		Level:           req.ApprovalLevel,
		Comment:         req.Comment,
		ActionAt:        now,
		Source:          source,
	}

	// Atomic: FindOneAndUpdate กับ filter status=pending
	atomicFilter := bson.M{
		"shop_id": req.ShopID,
		"docno":   req.DocNo,
		"status":  "pending",
	}
	update := bson.M{
		"$set": bson.M{
			"status":       "rejected",
			"last_comment": req.Comment,
			"updated_at":   now,
		},
		"$push": bson.M{
			"history": newHistory,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedStatus POApprovalStatus
	err := collection.FindOneAndUpdate(ctx, atomicFilter, update, opts).Decode(&updatedStatus)

	if err == mongo.ErrNoDocuments {
		var existing POApprovalStatus
		findErr := collection.FindOne(ctx, bson.M{"shop_id": req.ShopID, "docno": req.DocNo}).Decode(&existing)
		if findErr == mongo.ErrNoDocuments {
			return c.JSON(http.StatusNotFound, map[string]any{
				"success": false,
				"message": "ไม่พบเอกสารนี้ในระบบอนุมัติ",
			})
		}
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "เอกสารนี้ไม่ได้อยู่ในสถานะรออนุมัติ",
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Failed to reject PO",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"message": "ปฏิเสธเอกสารสำเร็จ",
		"data":    updatedStatus,
	})
}

// WithdrawPOHandler - ถอนการส่งอนุมัติ PO (เปลี่ยนสถานะจาก pending → rejected)
// ผู้สร้างเอกสารหรือผู้มีสิทธิ์อนุมัติสามารถถอนได้
func WithdrawPOHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req POApprovalStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.ShopID == "" || req.DocNo == "" || req.ActionBy == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "shop_id, docno, and action_by are required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := getTokenCollection(POApprovalStatusCollection)
	filter := bson.M{"shop_id": req.ShopID, "docno": req.DocNo}

	// ดึงสถานะปัจจุบัน
	var currentStatus POApprovalStatus
	err := collection.FindOne(ctx, filter).Decode(&currentStatus)
	if err == mongo.ErrNoDocuments {
		return c.JSON(http.StatusNotFound, map[string]any{
			"success": false,
			"message": "ไม่พบเอกสารนี้ในระบบอนุมัติ",
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Query execution failed",
		})
	}

	// ตรวจสอบว่าสถานะเป็น pending เท่านั้นจึงจะถอนได้
	if currentStatus.Status != "pending" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "เฉพาะเอกสารที่อยู่ในสถานะรออนุมัติเท่านั้นที่ถอนได้",
		})
	}

	// ตรวจสอบสิทธิ์: ผู้สร้างเอกสาร หรือ ผู้มีสิทธิ์อนุมัติ
	isCreator := req.ActionBy == currentStatus.CreatedBy
	isApprover := false

	if !isCreator {
		// ดึง approval settings เพื่อตรวจสอบว่าผู้ถอนอยู่ในรายชื่อผู้อนุมัติหรือไม่
		settingsCollection := getCollection(POApprovalSettingsCollection)
		var setting POApprovalSetting
		settingErr := settingsCollection.FindOne(ctx, bson.M{
			"shop_id":            req.ShopID,
			"purchase_type_code": currentStatus.PurchaseTypeCode,
		}).Decode(&setting)

		if settingErr == nil && setting.Rules != nil {
			for _, rule := range setting.Rules {
				for _, approver := range rule.Approvers {
					if approver.UserCode == req.ActionBy {
						isApprover = true
						break
					}
				}
				if isApprover {
					break
				}
			}
		}
	}

	if !isCreator && !isApprover {
		return c.JSON(http.StatusForbidden, map[string]any{
			"success": false,
			"message": "ไม่มีสิทธิ์ถอนการอนุมัติ — เฉพาะผู้สร้างเอกสารหรือผู้มีสิทธิ์อนุมัติเท่านั้น",
		})
	}

	now := time.Now().UTC()

	// กำหนด source
	source := req.Source
	if source == "" {
		source = "app"
	}

	// กำหนด comment ถ้าไม่ได้ส่งมา
	comment := req.Comment
	if comment == "" {
		comment = "ถอนการส่งอนุมัติโดย " + req.ActionByName
	}

	newHistory := ApprovalHistory{
		Action:          "withdraw",
		ActionBy:        req.ActionBy,
		ActionByName:    req.ActionByName,
		ApproverUserUID: getUserUID(ctx, req.ActionBy),
		Level:           req.ApprovalLevel,
		Comment:         comment,
		ActionAt:        now,
		Source:          source,
	}

	// Atomic: FindOneAndUpdate กับ filter status=pending → ป้องกัน race condition
	atomicFilter := bson.M{
		"shop_id": req.ShopID,
		"docno":   req.DocNo,
		"status":  "pending",
	}
	update := bson.M{
		"$set": bson.M{
			"status":       "rejected",
			"last_comment": comment,
			"updated_at":   now,
		},
		"$push": bson.M{
			"history": newHistory,
		},
	}

	opts := options.FindOneAndUpdate().SetReturnDocument(options.After)
	var updatedStatus POApprovalStatus
	err = collection.FindOneAndUpdate(ctx, atomicFilter, update, opts).Decode(&updatedStatus)

	if err == mongo.ErrNoDocuments {
		// สถานะเปลี่ยนไประหว่างตรวจสอบสิทธิ์กับการอัปเดต
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "เอกสารนี้ไม่ได้อยู่ในสถานะรออนุมัติแล้ว (อาจมีคนอนุมัติ/ปฏิเสธไปก่อน)",
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "ถอนการอนุมัติไม่สำเร็จ",
		})
	}

	logger.Info("[WithdrawPO] สำเร็จ - docNo: %s, ถอนโดย: %s (%s)",
		req.DocNo, req.ActionByName, req.ActionBy)

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"message": "ถอนการส่งอนุมัติสำเร็จ",
		"data":    updatedStatus,
	})
}

// GetPendingApprovalsHandler - ดึงรายการ PO ที่รออนุมัติสำหรับผู้ใช้
func GetPendingApprovalsHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req POApprovalStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.ShopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "shop_id is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := getTokenCollection(POApprovalStatusCollection)

	// ดึง PO ที่รออนุมัติ และ ผู้ใช้มีสิทธิ์อนุมัติได้
	filter := bson.M{
		"shop_id": req.ShopID,
		"status":  "pending",
	}

	// ถ้าระบุระดับสิทธิ์ ให้กรองเฉพาะที่ผู้ใช้อนุมัติได้
	if req.ApprovalLevel > 0 {
		filter["required_level"] = bson.M{"$lte": req.ApprovalLevel}
	}

	cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.M{"created_at": -1}))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Query execution failed",
		})
	}
	defer cursor.Close(ctx)

	var items []POApprovalStatus
	if err := cursor.All(ctx, &items); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Failed to decode results",
		})
	}

	if items == nil {
		items = []POApprovalStatus{}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"data":    items,
		"count":   len(items),
	})
}

// GetRejectedPOListHandler - ดึงรายการ PO ที่ถูกปฏิเสธ (สำหรับเลือกเป็นต้นแบบ)
func GetRejectedPOListHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req POApprovalStatusRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.ShopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "shop_id is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := getTokenCollection(POApprovalStatusCollection)

	// ดึง PO ที่ถูกปฏิเสธ
	filter := bson.M{
		"shop_id": req.ShopID,
		"status":  "rejected",
	}

	cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.M{"updated_at": -1}).SetLimit(50))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Query execution failed",
		})
	}
	defer cursor.Close(ctx)

	var items []POApprovalStatus
	if err := cursor.All(ctx, &items); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Failed to decode results",
		})
	}

	if items == nil {
		items = []POApprovalStatus{}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"data":    items,
		"count":   len(items),
	})
}

// =====================================================
// PO Approval Notification System (ระบบแจ้งเตือน)
// =====================================================

// Collection สำหรับเก็บ log การแจ้งเตือน
const POApprovalNotificationLogCollection = "po_approval_notification_log"

// NotificationLog บันทึกการส่งการแจ้งเตือน
type NotificationLog struct {
	ID               primitive.ObjectID `bson:"_id,omitempty" json:"guid,omitempty"`
	ShopID           string             `bson:"shop_id" json:"shop_id"`
	DocNo            string             `bson:"docno" json:"docno"`
	GuidFixed        string             `bson:"guid_fixed" json:"guid_fixed"`
	ApproverCode     string             `bson:"approver_code" json:"approver_code"`
	ApproverName     string             `bson:"approver_name" json:"approver_name"`
	NotificationType string             `bson:"notification_type" json:"notification_type"` // email, line
	RecipientEmail   string             `bson:"recipient_email,omitempty" json:"recipient_email,omitempty"`
	RecipientLineID  string             `bson:"recipient_line_id,omitempty" json:"recipient_line_id,omitempty"`
	Status           string             `bson:"status" json:"status"` // sent, failed, pending
	ErrorMessage     string             `bson:"error_message,omitempty" json:"error_message,omitempty"`
	SentAt           time.Time          `bson:"sent_at" json:"sent_at"`
	CreatedAt        time.Time          `bson:"created_at" json:"created_at"`
	// Tracking "opened/read" notification
	OpenedAt   *time.Time `bson:"opened_at,omitempty" json:"opened_at,omitempty"`     // เวลาที่เปิดอ่าน
	OpenedFrom string     `bson:"opened_from,omitempty" json:"opened_from,omitempty"` // แหล่งที่เปิด: email, line, liff
}

// NotificationRequest request สำหรับการแจ้งเตือน
type NotificationRequest struct {
	ShopID       string `json:"shop_id"`
	DocNo        string `json:"docno"`
	GuidFixed    string `json:"guid_fixed"`
	ApproverCode string `json:"approver_code"`
	CheckOnly    bool   `json:"check_only"` // true = ตรวจสอบอย่างเดียว ไม่ส่งจริง
}

// CheckNotificationSentHandler - ตรวจสอบว่าเคยส่งแจ้งเตือนแล้วหรือยัง
func CheckNotificationSentHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req NotificationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.ShopID == "" || req.DocNo == "" || req.ApproverCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "shop_id, docno, and approver_code are required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// ใช้ getTokenCollection เพื่ออ่านจาก bcai_documents ที่ LINE-LIFF ใช้
	collection := getTokenCollection(POApprovalNotificationLogCollection)

	// ตรวจสอบว่าเคยส่งแจ้งเตือนสำเร็จแล้วหรือยัง
	filter := bson.M{
		"shop_id":       req.ShopID,
		"docno":         req.DocNo,
		"approver_code": req.ApproverCode,
		"status":        "sent",
	}

	count, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Query execution failed",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success":    true,
		"sent":       count > 0,
		"sent_count": count,
	})
}

// SendApprovalNotificationHandler - ส่งการแจ้งเตือนไปยังผู้อนุมัติ (ไม่ส่งซ้ำถ้าส่งแล้ว)
func SendApprovalNotificationHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req NotificationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.ShopID == "" || req.DocNo == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "shop_id and docno are required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// 1. ดึงสถานะการอนุมัติ PO
	statusCollection := getTokenCollection(POApprovalStatusCollection)
	var poStatus POApprovalStatus
	err := statusCollection.FindOne(ctx, bson.M{
		"shop_id": req.ShopID,
		"docno":   req.DocNo,
		"status":  "pending",
	}).Decode(&poStatus)

	if err == mongo.ErrNoDocuments {
		return c.JSON(http.StatusOK, map[string]any{
			"success": true,
			"message": "ไม่พบ PO ที่รออนุมัติ หรืออนุมัติแล้ว",
			"sent":    0,
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Query execution failed",
		})
	}

	// 2. ดึงกฎการอนุมัติเพื่อหาผู้อนุมัติ
	settingsCollection := getCollection(POApprovalSettingsCollection)
	var setting POApprovalSetting
	err = settingsCollection.FindOne(ctx, bson.M{
		"shop_id":            req.ShopID,
		"purchase_type_code": poStatus.PurchaseTypeCode,
	}).Decode(&setting)

	if err == mongo.ErrNoDocuments {
		return c.JSON(http.StatusOK, map[string]any{
			"success": true,
			"message": "ไม่พบกฎการอนุมัติ",
			"sent":    0,
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Query execution failed",
		})
	}

	// 3. หาผู้อนุมัติที่เกี่ยวข้องตามวงเงิน
	var approvers []ApproverInfo
	for _, rule := range setting.Rules {
		if poStatus.TotalAmount >= rule.MinAmount {
			if rule.MaxAmount == 0 || poStatus.TotalAmount <= rule.MaxAmount {
				if rule.ApprovalLevel == poStatus.RequiredLevel && len(rule.Approvers) > 0 {
					approvers = rule.Approvers
					break
				}
			}
		}
	}

	if len(approvers) == 0 {
		return c.JSON(http.StatusOK, map[string]any{
			"success": true,
			"message": "ไม่พบรายชื่อผู้อนุมัติในกฎ",
			"sent":    0,
		})
	}

	// 4. ส่งการแจ้งเตือนไปยังผู้อนุมัติแต่ละคน (ตรวจสอบไม่ส่งซ้ำ)
	// ใช้ getTokenCollection เพื่อเขียนไป bcai_documents ที่ LINE-LIFF ใช้
	notificationCollection := getTokenCollection(POApprovalNotificationLogCollection)
	now := time.Now().UTC() // เก็บเวลาเป็น UTC+0
	sentCount := 0
	skippedCount := 0
	results := []map[string]any{}

	for _, approver := range approvers {
		// ตรวจสอบว่าเคยส่งสำเร็จแล้วหรือยัง
		existingFilter := bson.M{
			"shop_id":       req.ShopID,
			"docno":         req.DocNo,
			"approver_code": approver.UserCode,
			"status":        "sent",
		}
		existingCount, _ := notificationCollection.CountDocuments(ctx, existingFilter)
		if existingCount > 0 {
			skippedCount++
			results = append(results, map[string]any{
				"approver": approver.UserCode,
				"status":   "skipped",
				"reason":   "already_sent",
			})
			continue
		}

		// ส่ง Email ถ้ามี
		if approver.Email != "" {
			emailLog := NotificationLog{
				ShopID:           req.ShopID,
				DocNo:            req.DocNo,
				GuidFixed:        poStatus.GuidFixed,
				ApproverCode:     approver.UserCode,
				ApproverName:     approver.UserName,
				NotificationType: "email",
				RecipientEmail:   approver.Email,
				Status:           "sent", // TODO: เปลี่ยนเป็น pending เมื่อเรียก email service จริง
				SentAt:           now,
				CreatedAt:        now,
			}
			notificationCollection.InsertOne(ctx, emailLog)
			sentCount++
			results = append(results, map[string]any{
				"approver": approver.UserCode,
				"type":     "email",
				"status":   "sent",
				"to":       approver.Email,
			})
		}

		// ส่ง LINE Push ถ้ามี
		if approver.LineUserID != "" {
			lineLog := NotificationLog{
				ShopID:           req.ShopID,
				DocNo:            req.DocNo,
				GuidFixed:        poStatus.GuidFixed,
				ApproverCode:     approver.UserCode,
				ApproverName:     approver.UserName,
				NotificationType: "line",
				RecipientLineID:  approver.LineUserID,
				Status:           "sent", // TODO: เปลี่ยนเป็น pending เมื่อเรียก LINE service จริง
				SentAt:           now,
				CreatedAt:        now,
			}
			notificationCollection.InsertOne(ctx, lineLog)
			sentCount++
			results = append(results, map[string]any{
				"approver": approver.UserCode,
				"type":     "line",
				"status":   "sent",
				"to":       approver.LineDisplayName,
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success":       true,
		"message":       "ประมวลผลการแจ้งเตือนเสร็จสิ้น",
		"sent_count":    sentCount,
		"skipped_count": skippedCount,
		"results":       results,
	})
}

// ProcessPendingNotificationsHandler - ตรวจสอบและส่งแจ้งเตือน PO ที่รออนุมัติทั้งหมด
func ProcessPendingNotificationsHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	type ProcessRequest struct {
		ShopID string `json:"shop_id"`
	}

	var req ProcessRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.ShopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "shop_id is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	// ดึง PO ทั้งหมดที่รออนุมัติ
	statusCollection := getTokenCollection(POApprovalStatusCollection)
	cursor, err := statusCollection.Find(ctx, bson.M{
		"shop_id": req.ShopID,
		"status":  "pending",
	})
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Query execution failed",
		})
	}
	defer cursor.Close(ctx)

	var pendingPOs []POApprovalStatus
	if err := cursor.All(ctx, &pendingPOs); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Failed to decode results",
		})
	}

	processedCount := 0
	notificationsSent := 0

	for _, po := range pendingPOs {
		processedCount++

		// ดึงกฎการอนุมัติ
		settingsCollection := getCollection(POApprovalSettingsCollection)
		var setting POApprovalSetting
		err := settingsCollection.FindOne(ctx, bson.M{
			"shop_id":            req.ShopID,
			"purchase_type_code": po.PurchaseTypeCode,
		}).Decode(&setting)

		if err != nil {
			continue
		}

		// หาผู้อนุมัติ
		var approvers []ApproverInfo
		for _, rule := range setting.Rules {
			if po.TotalAmount >= rule.MinAmount {
				if rule.MaxAmount == 0 || po.TotalAmount <= rule.MaxAmount {
					if rule.ApprovalLevel == po.RequiredLevel && len(rule.Approvers) > 0 {
						approvers = rule.Approvers
						break
					}
				}
			}
		}

		// ส่งแจ้งเตือน (ตรวจสอบไม่ซ้ำ)
		// ใช้ getTokenCollection เพื่อเขียนไป bcai_documents ที่ LINE-LIFF ใช้
		notificationCollection := getTokenCollection(POApprovalNotificationLogCollection)
		now := time.Now().UTC() // เก็บเวลาเป็น UTC+0

		for _, approver := range approvers {
			existingFilter := bson.M{
				"shop_id":       req.ShopID,
				"docno":         po.DocNo,
				"approver_code": approver.UserCode,
				"status":        "sent",
			}
			existingCount, _ := notificationCollection.CountDocuments(ctx, existingFilter)
			if existingCount > 0 {
				continue
			}

			if approver.Email != "" {
				emailLog := NotificationLog{
					ShopID:           req.ShopID,
					DocNo:            po.DocNo,
					GuidFixed:        po.GuidFixed,
					ApproverCode:     approver.UserCode,
					ApproverName:     approver.UserName,
					NotificationType: "email",
					RecipientEmail:   approver.Email,
					Status:           "sent",
					SentAt:           now,
					CreatedAt:        now,
				}
				notificationCollection.InsertOne(ctx, emailLog)
				notificationsSent++
			}

			if approver.LineUserID != "" {
				lineLog := NotificationLog{
					ShopID:           req.ShopID,
					DocNo:            po.DocNo,
					GuidFixed:        po.GuidFixed,
					ApproverCode:     approver.UserCode,
					ApproverName:     approver.UserName,
					NotificationType: "line",
					RecipientLineID:  approver.LineUserID,
					Status:           "sent",
					SentAt:           now,
					CreatedAt:        now,
				}
				notificationCollection.InsertOne(ctx, lineLog)
				notificationsSent++
			}
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success":            true,
		"message":            "ประมวลผลการแจ้งเตือนเสร็จสิ้น",
		"processed_po_count": processedCount,
		"notifications_sent": notificationsSent,
	})
}

// GetNotificationLogsHandler - ดึงประวัติการแจ้งเตือน
func GetNotificationLogsHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req NotificationRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.ShopID == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "shop_id is required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// ใช้ getTokenCollection เพื่ออ่านจาก bcai_documents ที่ LINE-LIFF ใช้
	collection := getTokenCollection(POApprovalNotificationLogCollection)

	filter := bson.M{"shop_id": req.ShopID}
	if req.DocNo != "" {
		filter["docno"] = req.DocNo
	}
	if req.ApproverCode != "" {
		filter["approver_code"] = req.ApproverCode
	}

	cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.M{"created_at": -1}).SetLimit(100))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Query execution failed",
		})
	}
	defer cursor.Close(ctx)

	var logs []NotificationLog
	if err := cursor.All(ctx, &logs); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Failed to decode results",
		})
	}

	if logs == nil {
		logs = []NotificationLog{}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success": true,
		"data":    logs,
		"count":   len(logs),
	})
}

// =====================================================
// Mark Notification Opened (บันทึกการเปิดอ่าน)
// =====================================================

// MarkOpenedRequest request สำหรับบันทึกการเปิดอ่าน
type MarkOpenedRequest struct {
	ShopID       string `json:"shop_id"`
	DocNo        string `json:"docno"`
	GuidFixed    string `json:"guid_fixed"`
	ApproverCode string `json:"approver_code"`
	OpenedFrom   string `json:"opened_from"` // email, line, liff
}

// MarkNotificationOpenedHandler - บันทึกว่าผู้อนุมัติเปิดอ่านแจ้งเตือนแล้ว
// พร้อมส่ง LINE notification แจ้งผู้สร้างเอกสารว่ามีคนเปิดอ่านแล้ว
func MarkNotificationOpenedHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req MarkOpenedRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.ShopID == "" || req.DocNo == "" || req.ApproverCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "shop_id, docno, and approver_code are required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// ใช้ getTokenCollection เพื่อเขียนไป bcai_documents ที่ LINE-LIFF ใช้
	collection := getTokenCollection(POApprovalNotificationLogCollection)
	now := time.Now().UTC() // เก็บเวลาเป็น UTC+0

	// อัปเดต notification ที่ยังไม่ได้เปิดอ่าน
	filter := bson.M{
		"shop_id":       req.ShopID,
		"docno":         req.DocNo,
		"approver_code": req.ApproverCode,
		"opened_at":     bson.M{"$exists": false}, // ยังไม่เคยเปิด
	}

	update := bson.M{
		"$set": bson.M{
			"opened_at":   now,
			"opened_from": req.OpenedFrom,
		},
	}

	result, err := collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]any{
			"success": false,
			"message": "Failed to mark notification as opened",
		})
	}

	// ถ้ามี notification ถูกอัปเดต ให้ส่ง LINE แจ้งผู้สร้างเอกสาร
	if result.ModifiedCount > 0 {
		// ดึงข้อมูล PO จาก po_approval_status
		statusCollection := getTokenCollection(POApprovalStatusCollection)
		var poStatus POApprovalStatus
		err := statusCollection.FindOne(ctx, bson.M{
			"shop_id": req.ShopID,
			"docno":   req.DocNo,
		}).Decode(&poStatus)

		if err == nil {
			// ดึงชื่อผู้อนุมัติจาก notification log
			var notifLog NotificationLog
			_ = collection.FindOne(ctx, bson.M{
				"shop_id":       req.ShopID,
				"docno":         req.DocNo,
				"approver_code": req.ApproverCode,
			}).Decode(&notifLog)

			// แปลง DocDatetime string เป็น time.Time
			docDatetime := time.Now().UTC()
			if poStatus.DocDatetime != "" {
				if parsedTime, parseErr := time.Parse("2006-01-02T15:04:05", poStatus.DocDatetime); parseErr == nil {
					docDatetime = parsedTime
				} else if parsedTime, parseErr := time.Parse("2006-01-02 15:04:05", poStatus.DocDatetime); parseErr == nil {
					docDatetime = parsedTime
				}
			}

			// ใช้ชื่อผู้อนุมัติจาก notification log ถ้ามี ไม่งั้นใช้ approver code
			openerName := req.ApproverCode
			if notifLog.ApproverName != "" {
				openerName = notifLog.ApproverName
			}

			// สร้าง parameters สำหรับส่ง LINE notification
			params := OpenedNotificationParams{
				ShopID:           req.ShopID,
				DocNo:            req.DocNo,
				DocDatetime:      docDatetime,
				CreatorCode:      poStatus.CreatedBy,
				CreatorName:      poStatus.CreatedByName,
				OpenerCode:       req.ApproverCode,
				OpenerName:       openerName,
				OpenedFrom:       req.OpenedFrom,
				PurchaseTypeName: poStatus.PurchaseTypeName,
				TotalAmount:      poStatus.TotalAmount,
				CreditorName:     poStatus.CustName,
			}

			// ส่ง LINE notification แจ้งผู้สร้างเอกสาร (ทำแบบ async ไม่ block response)
			go func() {
				if sendErr := SendLineNotifyCreatorOpened(params); sendErr != nil {
					// Log error but don't fail the response
					logger.Warn("[MarkOpened] Failed to send LINE notification to creator: %v", sendErr)
				} else {
					logger.Info("[MarkOpened] LINE notification sent to creator %s for PO %s", params.CreatorCode, params.DocNo)
				}
			}()
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success":       true,
		"message":       "บันทึกการเปิดอ่านสำเร็จ",
		"updated_count": result.ModifiedCount,
		"opened_at":     now.Format("2006-01-02 15:04:05"),
		"opened_from":   req.OpenedFrom,
	})
}

// =====================================================
// Approval Timeline (ประวัติรวมทั้งหมด)
// =====================================================

// TimelineItem รายการใน timeline
type TimelineItem struct {
	Action       string    `json:"action"`                   // submit, modify, approve, reject, notification_sent, notification_opened
	ActionBy     string    `json:"action_by,omitempty"`      // รหัสผู้ดำเนินการ
	ActionByName string    `json:"action_by_name,omitempty"` // ชื่อผู้ดำเนินการ
	ActionAt     time.Time `json:"action_at"`                // เวลาดำเนินการ
	Detail       string    `json:"detail,omitempty"`         // รายละเอียดเพิ่มเติม
	Comment      string    `json:"comment,omitempty"`        // หมายเหตุ
	Level        int       `json:"level,omitempty"`          // ระดับการอนุมัติ
	Type         string    `json:"type,omitempty"`           // ประเภท notification: email, line
	Recipient    string    `json:"recipient,omitempty"`      // ผู้รับ notification
}

// TimelineRequest request สำหรับดึง timeline
type TimelineRequest struct {
	ShopID    string `json:"shop_id"`
	DocNo     string `json:"docno"`
	GuidFixed string `json:"guid_fixed"`
}

// GetApprovalTimelineHandler - ดึงประวัติรวมทั้งหมด (submit, approve, reject, notification sent/opened)
func GetApprovalTimelineHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req TimelineRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.ShopID == "" || req.DocNo == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "shop_id and docno are required",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	var timeline []TimelineItem

	// 1. ดึงประวัติการอนุมัติจาก POApprovalStatus
	statusCollection := getTokenCollection(POApprovalStatusCollection)
	var poStatus POApprovalStatus
	err := statusCollection.FindOne(ctx, bson.M{
		"shop_id": req.ShopID,
		"docno":   req.DocNo,
	}).Decode(&poStatus)

	if err == nil {
		// เพิ่มประวัติจาก ApprovalHistory
		for _, h := range poStatus.History {
			item := TimelineItem{
				Action:       h.Action,
				ActionBy:     h.ActionBy,
				ActionByName: h.ActionByName,
				ActionAt:     h.ActionAt,
				Comment:      h.Comment,
				Level:        h.Level,
				Type:         h.Source, // ช่องทาง: line, email, app
			}
			// เพิ่ม detail ตาม action (รวม source ถ้ามี)
			sourceText := ""
			if h.Source != "" {
				switch h.Source {
				case "line":
					sourceText = " (ผ่าน LINE)"
				case "email":
					sourceText = " (ผ่าน Email)"
				case "app":
					sourceText = " (ผ่าน App)"
				}
			}
			switch h.Action {
			case "submit":
				item.Detail = "ส่งเอกสารขออนุมัติ"
			case "modify":
				item.Detail = "แก้ไขเอกสาร"
			case "approve":
				item.Detail = "อนุมัติเอกสาร" + sourceText
			case "reject":
				item.Detail = "ไม่อนุมัติเอกสาร" + sourceText
			case "auto_approve":
				item.Detail = "อนุมัติอัตโนมัติ"
			}
			timeline = append(timeline, item)
		}
	}

	// 2. ดึงประวัติการแจ้งเตือนจาก NotificationLog
	// ใช้ getTokenCollection เพื่ออ่านจาก bcai_documents ที่ LINE-LIFF ใช้
	notificationCollection := getTokenCollection(POApprovalNotificationLogCollection)
	cursor, err := notificationCollection.Find(ctx, bson.M{
		"shop_id": req.ShopID,
		"docno":   req.DocNo,
	})
	if err == nil {
		defer cursor.Close(ctx)
		for cursor.Next(ctx) {
			var log NotificationLog
			if err := cursor.Decode(&log); err != nil {
				continue
			}

			// เพิ่ม "sent" event
			if log.Status == "sent" {
				recipient := log.RecipientEmail
				if log.NotificationType == "line" {
					recipient = log.ApproverName
				}
				timeline = append(timeline, TimelineItem{
					Action:       "notification_sent",
					ActionBy:     log.ApproverCode,
					ActionByName: log.ApproverName,
					ActionAt:     log.SentAt,
					Detail:       fmt.Sprintf("ส่งแจ้งเตือนทาง %s", log.NotificationType),
					Type:         log.NotificationType,
					Recipient:    recipient,
				})
			}

			// เพิ่ม "opened" event (ถ้ามี)
			if log.OpenedAt != nil {
				timeline = append(timeline, TimelineItem{
					Action:       "notification_opened",
					ActionBy:     log.ApproverCode,
					ActionByName: log.ApproverName,
					ActionAt:     *log.OpenedAt,
					Detail:       fmt.Sprintf("เปิดอ่านจาก %s", log.OpenedFrom),
					Type:         log.OpenedFrom,
				})
			}
		}
	}

	// 3. เรียงลำดับตามเวลา (ล่าสุดก่อน)
	for i := 0; i < len(timeline)-1; i++ {
		for j := i + 1; j < len(timeline); j++ {
			if timeline[i].ActionAt.Before(timeline[j].ActionAt) {
				timeline[i], timeline[j] = timeline[j], timeline[i]
			}
		}
	}

	// เพิ่มข้อมูล PO status ถ้ามี
	var statusInfo map[string]any
	if err == nil || poStatus.DocNo != "" {
		statusInfo = map[string]any{
			"docno":                  poStatus.DocNo,
			"guid_fixed":             poStatus.GuidFixed,
			"status":                 poStatus.Status,
			"total_amount":           poStatus.TotalAmount,
			"purchase_type_name":     poStatus.PurchaseTypeName,
			"required_level":         poStatus.RequiredLevel,
			"required_level_name":    poStatus.RequiredLevelName,
			"current_approved_level": poStatus.CurrentApprovedLevel,
			"created_by":             poStatus.CreatedBy,
			"created_by_name":        poStatus.CreatedByName,
			"last_comment":           poStatus.LastComment,
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success":  true,
		"status":   statusInfo,
		"timeline": timeline,
		"count":    len(timeline),
	})
}

// =====================================================
// Update PO Approval Status to Cancelled
// =====================================================

// UpdatePOApprovalStatusToCancelled - อัปเดตสถานะ PO approval เป็น cancelled เมื่อ PO ถูกยกเลิก
// ฟังก์ชันนี้ถูกเรียกจาก Kafka consumer เมื่อมีการยกเลิก PO
func UpdatePOApprovalStatusToCancelled(shopID, docNo, cancelReason, cancelUserCode, cancelUserName string) error {
	if tokenAtlasDB == nil {
		return fmt.Errorf("MongoDB not connected")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	statusCollection := tokenAtlasDB.Collection("po_approval_status")

	// หา PO status ที่มีอยู่
	filter := bson.M{
		"shop_id": shopID,
		"docno":   docNo,
	}

	var poStatus POApprovalStatus
	err := statusCollection.FindOne(ctx, filter).Decode(&poStatus)
	if err == mongo.ErrNoDocuments {
		// ไม่มี approval status = ไม่ต้องทำอะไร
		logger.Info("[CancelPO] No approval status found for %s - skipping", docNo)
		return nil
	}
	if err != nil {
		return fmt.Errorf("failed to find PO status: %w", err)
	}

	// ถ้าสถานะเป็น cancelled อยู่แล้ว ไม่ต้องทำอะไร
	if poStatus.Status == "cancelled" {
		logger.Info("[CancelPO] PO %s already cancelled - skipping", docNo)
		return nil
	}

	// เพิ่มประวัติการยกเลิก
	history := poStatus.History
	if history == nil {
		history = []ApprovalHistory{}
	}
	history = append(history, ApprovalHistory{
		Action:          "cancel",
		ActionBy:        cancelUserCode,
		ActionByName:    cancelUserName,
		ApproverUserUID: getUserUID(ctx, cancelUserCode),
		Comment:         cancelReason,
		ActionAt:        time.Now(),
		Source:          "system",
	})

	// อัปเดตสถานะเป็น cancelled
	update := bson.M{
		"$set": bson.M{
			"status":        "cancelled",
			"history":       history,
			"last_comment":  cancelReason,
			"cancelled_at":  time.Now(),
			"cancelled_by":  cancelUserCode,
			"cancel_reason": cancelReason,
		},
	}

	_, err = statusCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("failed to update PO status to cancelled: %w", err)
	}

	logger.Success("[CancelPO] ✅ PO %s status updated to cancelled - reason: %s", docNo, cancelReason)
	return nil
}

// GetApprovalStatusMapByShop — ดึง approval status ทั้งหมดของ shop (สำหรับ rebuild)
// return map[docno]status เช่น {"PO20260107001": "approved", "PO20260108002": "pending"}
func GetApprovalStatusMapByShop(shopID string) (map[string]string, error) {
	collection := getTokenCollection(POApprovalStatusCollection)
	if collection == nil {
		return nil, fmt.Errorf("approval collection not initialized")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	filter := bson.M{"shop_id": shopID}
	opts := options.Find().SetProjection(bson.M{"docno": 1, "status": 1})

	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("query po_approval_status: %w", err)
	}
	defer cursor.Close(ctx)

	result := make(map[string]string)
	for cursor.Next(ctx) {
		var doc struct {
			DocNo  string `bson:"docno"`
			Status string `bson:"status"`
		}
		if err := cursor.Decode(&doc); err == nil && doc.DocNo != "" {
			result[doc.DocNo] = doc.Status
		}
	}

	logger.Info("[Approval] ดึง approval status สำหรับ shop %s ได้ %d รายการ", shopID, len(result))
	return result, nil
}
