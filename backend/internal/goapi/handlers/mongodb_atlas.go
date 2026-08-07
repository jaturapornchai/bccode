package handlers

import (
	"context"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	"smlcloudplatform/internal/goapi/config"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/myglobal"
	"smlcloudplatform/internal/utils"
	msmodels "smlcloudplatform/pkg/microservice/models"

	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	atlasClient *mongo.Client
	atlasDB     *mongo.Database
)

// InitMongoAtlas เชื่อมต่อ MongoDB โดยใช้ myglobal.MongoConnect()
func InitMongoAtlas() error {
	client, err := myglobal.MongoConnect()
	if err != nil {
		logger.Error("Failed to connect to MongoDB: %v", err)
		return err
	}

	atlasClient = client

	dbName := config.NewServiceConfig().MongodbDatabaseName()
	atlasDB = client.Database(dbName)

	logger.Success("✅ MongoDB connected for handlers (DB: %s)", dbName)
	return nil
}

// DisconnectMongoAtlas ปิดการเชื่อมต่อ MongoDB
// หมายเหตุ: ใช้ myglobal.DisconnectMongo() แทน เพราะ connection เป็น singleton
func DisconnectMongoAtlas() {
	// ไม่ต้อง disconnect ที่นี่ เพราะ myglobal.DisconnectMongo() จะจัดการให้
	logger.Info("MongoDB handlers cleanup completed")
}

// getDatabase - เลือก database ตาม request หรือใช้ default
func getDatabase(dbName string) *mongo.Database {
	if dbName != "" && atlasClient != nil {
		return atlasClient.Database(dbName)
	}
	return atlasDB
}

// GetAtlasConnection returns the shared MongoDB client and database for use by other packages
func GetAtlasConnection() (*mongo.Client, *mongo.Database) {
	return atlasClient, atlasDB
}

func normalizedAtlasTenantID(holdingCode string) string {
	if strings.TrimSpace(holdingCode) != "" {
		return strings.TrimSpace(holdingCode)
	}
	return ""
}

type atlasTenantIDs struct {
	authTenant string
	dataTenant string
	filterIDs  []string
}

func resolveAtlasTenantIDs(holdingCode string) atlasTenantIDs {
	holdingCode = strings.TrimSpace(holdingCode)
	authTenant := holdingCode
	dataTenant := holdingCode

	filterIDs := []string{dataTenant}

	return atlasTenantIDs{authTenant: authTenant, dataTenant: dataTenant, filterIDs: filterIDs}
}

func atlasTenantFilter(tenantID string) bson.M {
	return bson.M{
		"$or": []bson.M{
			{"holdingcode": tenantID},
			{"holdingcode": tenantID},
		},
	}
}

func atlasTenantFilterAny(tenantIDs ...string) bson.M {
	clauses := make([]bson.M, 0, len(tenantIDs)*2)
	seen := map[string]bool{}
	for _, tenantID := range tenantIDs {
		tenantID = strings.TrimSpace(tenantID)
		if tenantID == "" || seen[tenantID] {
			continue
		}
		clauses = append(clauses, bson.M{"holdingcode": tenantID}, bson.M{"holdingcode": tenantID})
		seen[tenantID] = true
	}
	if len(clauses) == 0 {
		return atlasTenantFilter("")
	}
	return bson.M{"$or": clauses}
}

func atlasIdentityFilter(guidFixed string, email string, cartID string, userUID string) bson.M {
	guidFixed = strings.TrimSpace(guidFixed)
	email = strings.TrimSpace(email)
	cartID = strings.TrimSpace(cartID)
	userUID = strings.TrimSpace(userUID)
	identityFilters := make([]bson.M, 0, 3)
	if guidFixed != "" {
		identityFilters = append(identityFilters,
			bson.M{"guidfixed": guidFixed},
			bson.M{"email": guidFixed, "cartid": guidFixed},
		)
	}
	if email != "" && cartID != "" && (email != guidFixed || cartID != guidFixed) {
		identityFilters = append(identityFilters, bson.M{"email": email, "cartid": cartID})
	}
	if userUID != "" {
		identityFilters = append(identityFilters, bson.M{"useruid": userUID})
	}
	if len(identityFilters) == 0 {
		return bson.M{}
	}
	if len(identityFilters) == 1 {
		return identityFilters[0]
	}
	return bson.M{"$or": identityFilters}
}

func andAtlasFilters(filters ...bson.M) bson.M {
	active := make([]bson.M, 0, len(filters))
	for _, filter := range filters {
		if len(filter) > 0 {
			active = append(active, filter)
		}
	}
	if len(active) == 1 {
		return active[0]
	}
	return bson.M{"$and": active}
}

func validateAtlasTenant(c echo.Context, requestHoldingCode string) error {
	userInfo, ok := c.Get("UserInfo").(msmodels.UserInfo)
	if !ok || userInfo.HoldingCode == "" {
		logger.Warn("MongoDB tenant check failed: missing authenticated user info")
		return c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"status":  "error",
			"code":    401,
			"message": "Unauthorized",
		})
	}
	if strings.TrimSpace(requestHoldingCode) == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "holdingcode is required",
		})
	}
	if requestHoldingCode != userInfo.HoldingCode {
		logger.Warn("MongoDB tenant mismatch: token_shop=%s request_shop=%s", userInfo.HoldingCode, requestHoldingCode)
		return c.JSON(http.StatusForbidden, map[string]interface{}{
			"status":  "error",
			"code":    403,
			"message": "Forbidden",
		})
	}
	return nil
}

// atlasBusinessCodeFields maps an atlas collection to its user-facing business code field.
// On save the value is normalized to uppercase + no whitespace, duplicate-guarded within the
// holding, and backed by a partial-unique index. This mirrors the frontend `businessCode`
// flag so the rule holds even when a client (MCP/AI/mobile) calls /atlas/update directly,
// bypassing the Next proxy. Email/username code fields (e.g. employeepermissions.employeecode)
// are intentionally excluded — emails are exempt from uppercasing.
var atlasBusinessCodeFields = map[string]string{
	"permissiondefinitions":   "permissioncode",
	"permissiongroups":        "groupcode",
	"approvalsettings":        "approvalcode",
	"productcolors":           "code",
	"productsizes":            "code",
	"productvariantmatrices":  "code",
	"productchannelprices":    "pricecode",
	"productserialregistries": "serialno",
}

func atlasBusinessCodeField(collection string) string {
	return atlasBusinessCodeFields[strings.TrimSpace(collection)]
}

// atlasIdentityCodeFields maps a collection to an identity code that must be unique per
// holding but is NOT a business code (e.g. employeepermissions.employeecode = email/username).
// It is trimmed only (NOT uppercased — emails are exempt) and case-insensitive duplicate-guarded.
var atlasIdentityCodeFields = map[string]string{
	"employeepermissions": "employeecode",
}

func atlasIdentityCodeField(collection string) string {
	return atlasIdentityCodeFields[strings.TrimSpace(collection)]
}

// normalizeAtlasBusinessCode uppercases and removes ALL whitespace (no spaces inside a code),
// e.g. "qa test 99" -> "QATEST99".
func normalizeAtlasBusinessCode(value interface{}) string {
	s, _ := value.(string)
	return strings.ToUpper(strings.Join(strings.Fields(s), ""))
}

var atlasCodeIndexEnsured sync.Map

// ensureAtlasCodeIndex creates a partial-unique index on (holdingcode, codeField) once per
// collection. Best-effort: if it fails (legacy duplicates or perms), log and continue — the
// in-handler duplicate check still guards new writes.
func ensureAtlasCodeIndex(ctx context.Context, db *mongo.Database, collection, codeField string) {
	key := db.Name() + "." + collection + "." + codeField
	if _, done := atlasCodeIndexEnsured.LoadOrStore(key, true); done {
		return
	}
	_, err := db.Collection(collection).Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "holdingcode", Value: 1}, {Key: codeField, Value: 1}},
		Options: options.Index().
			SetUnique(true).
			SetName("uniq_holdingcode_" + codeField).
			SetPartialFilterExpression(bson.M{codeField: bson.M{"$gt": ""}}),
	})
	if err != nil {
		atlasCodeIndexEnsured.Delete(key) // allow a retry on a later request
		logger.Warn("atlas unique index for %s.%s not created: %v", collection, codeField, err)
	}
}

// MongoAtlasUpdateHandler - Update/Insert (Upsert) document
func MongoAtlasUpdateHandler(c echo.Context) error {
	// Check if MongoDB is connected
	if atlasClient == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "error",
			"code":    503,
			"message": "MongoDB is not connected. Please set the environment-specific MongoDB URI.",
		})
	}

	var reqBody struct {
		Database    string                 `json:"database"` // optional - ถ้าไม่ระบุจะใช้ default
		Collection  string                 `json:"collection"`
		HoldingCode string                 `json:"holdingcode"`
		GuidFixed   string                 `json:"guidfixed"`
		Email       string                 `json:"email"`
		CartId      string                 `json:"cartid"`
		UserUID     string                 `json:"useruid"`
		Data        map[string]interface{} `json:"data"`
		Upsert      bool                   `json:"upsert"`
	}

	if err := c.Bind(&reqBody); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate required fields
	if reqBody.Collection == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Collection name is required",
		})
	}

	tenantIDs := resolveAtlasTenantIDs(reqBody.HoldingCode)
	if err := validateAtlasTenant(c, tenantIDs.authTenant); err != nil {
		return err
	}

	// Validate identifiers - ต้องไม่ว่าง
	if tenantIDs.dataTenant == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "holdingcode is required and cannot be empty",
		})
	}
	reqBody.GuidFixed = strings.TrimSpace(reqBody.GuidFixed)
	usesLegacyIdentity := reqBody.GuidFixed == "" && strings.TrimSpace(reqBody.Email) != "" && strings.TrimSpace(reqBody.CartId) != ""
	if reqBody.GuidFixed == "" && reqBody.Email == "" && reqBody.UserUID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "guidfixed is required and cannot be empty",
		})
	}
	if reqBody.GuidFixed == "" && reqBody.UserUID == "" && reqBody.CartId == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "cartid is required and cannot be empty",
		})
	}

	filter := andAtlasFilters(
		atlasTenantFilterAny(tenantIDs.filterIDs...),
		atlasIdentityFilter(reqBody.GuidFixed, reqBody.Email, reqBody.CartId, reqBody.UserUID),
	)

	// ใช้ data ทั้งหมดสำหรับ update
	if reqBody.Data == nil {
		reqBody.Data = make(map[string]interface{})
	}

	if strings.TrimSpace(reqBody.GuidFixed) == "" {
		reqBody.GuidFixed = utils.NewGUID()
	}

	// เพิ่ม identifiers ลงใน data ตาม model ใหม่ และเก็บ legacy key เฉพาะ request เก่าที่ไม่มี guidfixed
	reqBody.Data["holdingcode"] = tenantIDs.dataTenant
	reqBody.Data["guidfixed"] = reqBody.GuidFixed
	if usesLegacyIdentity {
		reqBody.Data["holdingcode"] = tenantIDs.authTenant
		reqBody.Data["email"] = reqBody.Email
		reqBody.Data["cartid"] = reqBody.CartId
	}

	// Business-code enforcement (uppercase + no whitespace) for code-bearing collections.
	codeField := atlasBusinessCodeField(reqBody.Collection)
	if codeField != "" {
		if _, ok := reqBody.Data[codeField]; ok {
			reqBody.Data[codeField] = normalizeAtlasBusinessCode(reqBody.Data[codeField])
		}
	}
	// Identity-code: trim only (NOT uppercased — e.g. employeecode = email/username).
	idCodeField := atlasIdentityCodeField(reqBody.Collection)
	if idCodeField != "" {
		if v, ok := reqBody.Data[idCodeField].(string); ok {
			reqBody.Data[idCodeField] = strings.TrimSpace(v)
		}
	}

	// เพิ่ม timestamp
	reqBody.Data["updatedat"] = time.Now()

	// Build update document
	update := bson.M{
		"$set": reqBody.Data,
	}

	// Perform upsert/update
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := getDatabase(reqBody.Database)
	collection := db.Collection(reqBody.Collection)
	opts := options.Update().SetUpsert(reqBody.Upsert)

	// Duplicate-code guard + unique index (defense for direct/MCP/mobile callers that bypass
	// the Next proxy). Rejects a code already used by a different record in the same holding.
	if codeField != "" {
		ensureAtlasCodeIndex(ctx, db, reqBody.Collection, codeField)
		if codeValue, _ := reqBody.Data[codeField].(string); codeValue != "" {
			dupFilter := andAtlasFilters(
				atlasTenantFilterAny(tenantIDs.filterIDs...),
				bson.M{codeField: codeValue, "guidfixed": bson.M{"$ne": reqBody.GuidFixed}},
			)
			if cnt, dupErr := collection.CountDocuments(ctx, dupFilter); dupErr == nil && cnt > 0 {
				return c.JSON(http.StatusConflict, map[string]interface{}{
					"status":  "error",
					"code":    409,
					"message": "duplicate code: " + codeValue,
				})
			}
		}
	}

	// Identity-code duplicate guard (case-insensitive, not uppercased — e.g. employeecode).
	if idCodeField != "" {
		ensureAtlasCodeIndex(ctx, db, reqBody.Collection, idCodeField)
		if idValue, _ := reqBody.Data[idCodeField].(string); idValue != "" {
			dupFilter := andAtlasFilters(
				atlasTenantFilterAny(tenantIDs.filterIDs...),
				bson.M{
					idCodeField: bson.M{"$regex": "^" + regexp.QuoteMeta(idValue) + "$", "$options": "i"},
					"guidfixed": bson.M{"$ne": reqBody.GuidFixed},
				},
			)
			if cnt, dupErr := collection.CountDocuments(ctx, dupFilter); dupErr == nil && cnt > 0 {
				return c.JSON(http.StatusConflict, map[string]interface{}{
					"status":  "error",
					"code":    409,
					"message": "duplicate code: " + idValue,
				})
			}
		}
	}

	result, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		logger.Error("MongoDB update error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to update document",
			"error":   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":         "success",
		"code":           200,
		"matched_count":  result.MatchedCount,
		"modified_count": result.ModifiedCount,
		"upserted_id":    result.UpsertedID,
	})
}

// MongoAtlasDeleteHandler - Delete document(s)
func MongoAtlasDeleteHandler(c echo.Context) error {
	// Check if MongoDB is connected
	if atlasClient == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "error",
			"code":    503,
			"message": "MongoDB is not connected. Please set the environment-specific MongoDB URI.",
		})
	}

	var reqBody struct {
		Database    string `json:"database"` // optional - ถ้าไม่ระบุจะใช้ default
		Collection  string `json:"collection"`
		HoldingCode string `json:"holdingcode"`
		GuidFixed   string `json:"guidfixed"`
		Email       string `json:"email"`
		CartId      string `json:"cartid"`
		UserUID     string `json:"useruid"`
		DeleteMany  bool   `json:"deletemany"` // true = deleteMany, false = deleteOne
	}

	if err := c.Bind(&reqBody); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate required fields
	if reqBody.Collection == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Collection name is required",
		})
	}

	tenantIDs := resolveAtlasTenantIDs(reqBody.HoldingCode)
	if err := validateAtlasTenant(c, tenantIDs.authTenant); err != nil {
		return err
	}

	// Validate identifiers - ต้องไม่ว่าง
	if tenantIDs.dataTenant == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "holdingcode is required and cannot be empty",
		})
	}
	reqBody.GuidFixed = strings.TrimSpace(reqBody.GuidFixed)
	if reqBody.GuidFixed == "" && reqBody.Email == "" && reqBody.UserUID == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "guidfixed is required and cannot be empty",
		})
	}
	if reqBody.GuidFixed == "" && reqBody.UserUID == "" && reqBody.CartId == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "cartid is required and cannot be empty",
		})
	}

	filter := andAtlasFilters(
		atlasTenantFilterAny(tenantIDs.filterIDs...),
		atlasIdentityFilter(reqBody.GuidFixed, reqBody.Email, reqBody.CartId, reqBody.UserUID),
	)

	// Perform delete
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := getDatabase(reqBody.Database)
	collection := db.Collection(reqBody.Collection)
	var deletedCount int64
	var deleteErr error

	if reqBody.DeleteMany {
		result, err := collection.DeleteMany(ctx, filter)
		if err != nil {
			deleteErr = err
		} else {
			deletedCount = result.DeletedCount
		}
	} else {
		result, err := collection.DeleteOne(ctx, filter)
		if err != nil {
			deleteErr = err
		} else {
			deletedCount = result.DeletedCount
		}
	}

	if deleteErr != nil {
		logger.Error("MongoDB delete error: %v", deleteErr)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to delete document(s)",
			"error":   deleteErr.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status":        "success",
		"code":          200,
		"deleted_count": deletedCount,
	})
}

// MongoAtlasGetHandler - Get document(s)
func MongoAtlasGetHandler(c echo.Context) error {
	// Check if MongoDB is connected
	if atlasClient == nil {
		return c.JSON(http.StatusServiceUnavailable, map[string]interface{}{
			"status":  "error",
			"code":    503,
			"message": "MongoDB is not connected. Please set the environment-specific MongoDB URI.",
		})
	}

	var reqBody struct {
		Database    string `json:"database"` // optional - ถ้าไม่ระบุจะใช้ default
		Collection  string `json:"collection"`
		HoldingCode string `json:"holdingcode"`
		GuidFixed   string `json:"guidfixed"`
		Email       string `json:"email"`
		CartId      string `json:"cartid"`
		UserUID     string `json:"useruid"`
		Limit       int64  `json:"limit"`
		Skip        int64  `json:"skip"`
	}

	if err := c.Bind(&reqBody); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Invalid request body",
			"error":   err.Error(),
		})
	}

	// Validate required fields
	if reqBody.Collection == "" {
		return c.JSON(http.StatusBadRequest, map[string]interface{}{
			"status":  "error",
			"code":    400,
			"message": "Collection name is required",
		})
	}

	tenantIDs := resolveAtlasTenantIDs(reqBody.HoldingCode)
	if err := validateAtlasTenant(c, tenantIDs.authTenant); err != nil {
		return err
	}

	filter := atlasTenantFilterAny(tenantIDs.filterIDs...)
	if reqBody.GuidFixed != "" || reqBody.Email != "" || reqBody.CartId != "" || reqBody.UserUID != "" {
		filter = andAtlasFilters(
			filter,
			atlasIdentityFilter(reqBody.GuidFixed, reqBody.Email, reqBody.CartId, reqBody.UserUID),
		)
	}

	// Query options
	opts := options.Find()
	if reqBody.Limit > 0 {
		opts.SetLimit(reqBody.Limit)
	}
	if reqBody.Skip > 0 {
		opts.SetSkip(reqBody.Skip)
	}

	// Perform query
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db := getDatabase(reqBody.Database)
	collection := db.Collection(reqBody.Collection)
	cursor, err := collection.Find(ctx, filter, opts)
	if err != nil {
		logger.Error("MongoDB find error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to query documents",
			"error":   err.Error(),
		})
	}
	defer cursor.Close(ctx)

	// Decode results
	var results []bson.M
	if err := cursor.All(ctx, &results); err != nil {
		logger.Error("MongoDB decode error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"status":  "error",
			"code":    500,
			"message": "Failed to decode documents",
			"error":   err.Error(),
		})
	}

	return c.JSON(http.StatusOK, map[string]interface{}{
		"status": "success",
		"code":   200,
		"count":  len(results),
		"data":   results,
	})
}
