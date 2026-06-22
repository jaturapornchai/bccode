package lineoa

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var (
	atlasClient *mongo.Client
	atlasDB     *mongo.Database
)

// Init initializes the Line OA handlers with MongoDB connection
func Init(client *mongo.Client, db *mongo.Database) {
	atlasClient = client
	atlasDB = db
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

// checkConnection returns error response if not connected
func checkConnection(c echo.Context) error {
	if !IsConnected() {
		return c.JSON(http.StatusServiceUnavailable, map[string]string{
			"error": "MongoDB is not connected",
			"code":  "MONGO_NOT_CONNECTED",
		})
	}
	return nil
}

// GetConfigsHandler - Get all Line OA configs for a shop
func GetConfigsHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req ConfigRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	if req.HoldingCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "holdingcode is required",
			"code":  "MISSING_HOLDING_CODE",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := getCollection(ConfigCollection)
	filter := bson.M{"holdingcode": req.HoldingCode}

	cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.M{"lineoatype": 1}))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Query execution failed",
			"code":  "QUERY_ERROR",
		})
	}
	defer cursor.Close(ctx)

	var configs []ConfigDoc
	if err := cursor.All(ctx, &configs); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to decode results",
			"code":  "DECODE_ERROR",
		})
	}

	if configs == nil {
		configs = []ConfigDoc{}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"data":   configs,
	})
}

// GetConfigHandler - Get single Line OA config by type
func GetConfigHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req ConfigRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	if req.HoldingCode == "" || req.LineOAType == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "holdingcode and lineoa_type are required",
			"code":  "MISSING_REQUIRED_FIELDS",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := getCollection(ConfigCollection)
	filter := bson.M{"holdingcode": req.HoldingCode, "lineoatype": req.LineOAType}

	var config ConfigDoc
	err := collection.FindOne(ctx, filter).Decode(&config)
	if err == mongo.ErrNoDocuments {
		return c.JSON(http.StatusNotFound, map[string]string{
			"error": "Config not found",
			"code":  "NOT_FOUND",
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Query execution failed",
			"code":  "QUERY_ERROR",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"data":   config,
	})
}

// SaveConfigHandler - Save Line OA config
func SaveConfigHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req SaveConfigRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	if req.HoldingCode == "" || req.LineOAType == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "holdingcode and lineoa_type are required",
			"code":  "MISSING_REQUIRED_FIELDS",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := getCollection(ConfigCollection)

	now := time.Now()
	var resultGUID string

	if req.GUID == "" {
		// Insert new or upsert
		resultGUID = uuid.New().String()
		filter := bson.M{"holdingcode": req.HoldingCode, "lineoatype": req.LineOAType}
		update := bson.M{
			"$set": bson.M{
				"guid":          resultGUID,
				"holdingcode":   req.HoldingCode,
				"lineoatype":    req.LineOAType,
				"channelid":     req.ChannelID,
				"channelsecret": req.ChannelSecret,
				"accesstoken":   req.AccessToken,
				"liffid":        req.LiffID,
				"isactive":      req.IsActive,
				"updatedat":     now,
			},
			"$setOnInsert": bson.M{
				"createdat": now,
			},
		}
		opts := options.Update().SetUpsert(true)
		result, err := collection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to save config",
				"code":  "SAVE_ERROR",
			})
		}
		if result.UpsertedID == nil {
			// Document was updated, get existing GUID
			var existing ConfigDoc
			if err := collection.FindOne(ctx, filter).Decode(&existing); err == nil {
				resultGUID = existing.GUID
			}
		}
	} else {
		// Update existing
		resultGUID = req.GUID
		filter := bson.M{"guid": req.GUID, "holdingcode": req.HoldingCode}
		update := bson.M{
			"$set": bson.M{
				"channelid":     req.ChannelID,
				"channelsecret": req.ChannelSecret,
				"accesstoken":   req.AccessToken,
				"liffid":        req.LiffID,
				"isactive":      req.IsActive,
				"updatedat":     now,
			},
		}
		_, err := collection.UpdateOne(ctx, filter, update)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to update config",
				"code":  "UPDATE_ERROR",
			})
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"guid":   resultGUID,
	})
}

// TestHandler - Test Line OA connection
func TestHandler(c echo.Context) error {
	var req TestRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Invalid request payload",
		})
	}

	if req.AccessToken == "" {
		return c.JSON(http.StatusBadRequest, map[string]any{
			"success": false,
			"message": "Access token is required",
		})
	}

	// Call LINE API to verify bot info
	client := &http.Client{Timeout: 10 * time.Second}
	apiReq, err := http.NewRequest("GET", "https://api.line.me/v2/bot/info", nil)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]any{
			"success": false,
			"message": "Failed to create request",
		})
	}

	apiReq.Header.Set("Authorization", "Bearer "+req.AccessToken)

	resp, err := client.Do(apiReq)
	if err != nil {
		return c.JSON(http.StatusOK, map[string]any{
			"success": false,
			"message": fmt.Sprintf("Connection failed: %v", err),
		})
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return c.JSON(http.StatusOK, map[string]any{
			"success": false,
			"message": fmt.Sprintf("LINE API returned status: %d", resp.StatusCode),
		})
	}

	var botInfo map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&botInfo); err != nil {
		return c.JSON(http.StatusOK, map[string]any{
			"success": false,
			"message": "Failed to parse LINE API response",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"success":  true,
		"message":  fmt.Sprintf("Connected to bot: %s", botInfo["displayName"]),
		"bot_info": botInfo,
	})
}

// GetEmployeesHandler - Get employees for Line OA config
func GetEmployeesHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req EmployeeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	if req.HoldingCode == "" || req.LineOAConfigGUID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "holdingcode and lineoa_config_guid are required",
			"code":  "MISSING_REQUIRED_FIELDS",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := getCollection(EmployeeCollection)
	filter := bson.M{
		"lineoaconfigguid": req.LineOAConfigGUID,
		"isactive":         true,
	}

	cursor, err := collection.Find(ctx, filter, options.Find().SetSort(bson.M{"employeename": 1}))
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Query execution failed",
			"code":  "QUERY_ERROR",
		})
	}
	defer cursor.Close(ctx)

	var employees []EmployeeDoc
	if err := cursor.All(ctx, &employees); err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to decode results",
			"code":  "DECODE_ERROR",
		})
	}

	if employees == nil {
		employees = []EmployeeDoc{}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"data":   employees,
	})
}

// AddEmployeeHandler - Add employee to Line OA
func AddEmployeeHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req AddEmployeeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	if req.HoldingCode == "" || req.LineOAConfigGUID == "" || req.EmployeeCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "holdingcode, lineoa_config_guid, and employee_code are required",
			"code":  "MISSING_REQUIRED_FIELDS",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := getCollection(EmployeeCollection)
	newGUID := uuid.New().String()

	filter := bson.M{
		"lineoaconfigguid": req.LineOAConfigGUID,
		"employeecode":     req.EmployeeCode,
	}
	update := bson.M{
		"$set": bson.M{
			"guid":             newGUID,
			"lineoaconfigguid": req.LineOAConfigGUID,
			"holdingcode":      req.HoldingCode,
			"employeecode":     req.EmployeeCode,
			"employeename":     req.EmployeeName,
			"isactive":         true,
		},
		"$setOnInsert": bson.M{
			"lineuserid": "",
			"createdat":  time.Now(),
		},
	}

	opts := options.Update().SetUpsert(true)
	result, err := collection.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to add employee",
			"code":  "INSERT_ERROR",
		})
	}

	resultGUID := newGUID
	if result.UpsertedID == nil {
		// Get existing GUID
		var existing EmployeeDoc
		if err := collection.FindOne(ctx, filter).Decode(&existing); err == nil {
			resultGUID = existing.GUID
		}
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"guid":   resultGUID,
	})
}

// RemoveEmployeeHandler - Remove employee from Line OA
func RemoveEmployeeHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req RemoveEmployeeRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	if req.HoldingCode == "" || req.GUID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "holdingcode and guid are required",
			"code":  "MISSING_REQUIRED_FIELDS",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := getCollection(EmployeeCollection)
	filter := bson.M{"guid": req.GUID}
	update := bson.M{"$set": bson.M{"isactive": false}}

	_, err := collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to remove employee",
			"code":  "DELETE_ERROR",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status":  "success",
		"message": "Employee removed",
	})
}

// GenerateLinkHandler - Generate LIFF link for employee
func GenerateLinkHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req GenerateLinkRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	if req.HoldingCode == "" || req.LineOAConfigGUID == "" || req.EmployeeCode == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "holdingcode, lineoaconfigguid, and employeecode are required",
			"code":  "MISSING_REQUIRED_FIELDS",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get LIFF ID from config
	configCollection := getCollection(ConfigCollection)
	var config ConfigDoc
	err := configCollection.FindOne(ctx, bson.M{"guid": req.LineOAConfigGUID}).Decode(&config)
	if err != nil || config.LiffID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "LIFF ID not configured",
			"code":  "LIFF_NOT_CONFIGURED",
		})
	}

	// Generate link token
	linkToken := uuid.New().String()

	// Save link token
	tokenCollection := getCollection(LinkTokenCollection)
	tokenDoc := LinkTokenDoc{
		Token:            linkToken,
		HoldingCode:      req.HoldingCode,
		LineOAConfigGUID: req.LineOAConfigGUID,
		EmployeeCode:     req.EmployeeCode,
		TokenType:        "employee",
		ExpiresAt:        time.Now().Add(24 * time.Hour),
		CreatedAt:        time.Now(),
	}

	_, err = tokenCollection.InsertOne(ctx, tokenDoc)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate link token",
			"code":  "TOKEN_ERROR",
		})
	}

	// Generate LIFF URL
	liffURL := fmt.Sprintf("https://liff.line.me/%s?token=%s&holdingcode=%s", config.LiffID, linkToken, req.HoldingCode)

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"link":   liffURL,
	})
}

// UserLinkHandler - Generate LIFF link for user to link Line OA
func UserLinkHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req UserLinkRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	if req.HoldingCode == "" || req.Username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "holdingcode and username are required",
			"code":  "MISSING_REQUIRED_FIELDS",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Get LIFF ID from lineoa_config (use 'manager' type as default for user linking)
	configCollection := getCollection(ConfigCollection)
	var config ConfigDoc

	// Try manager type first
	err := configCollection.FindOne(ctx, bson.M{
		"holdingcode": req.HoldingCode,
		"lineoatype":  "manager",
		"isactive":    true,
	}).Decode(&config)

	if err != nil || config.LiffID == "" {
		// Try any active config with LIFF ID
		err = configCollection.FindOne(ctx, bson.M{
			"holdingcode": req.HoldingCode,
			"isactive":    true,
			"liffid":      bson.M{"$ne": ""},
		}).Decode(&config)
		if err != nil || config.LiffID == "" {
			return c.JSON(http.StatusOK, map[string]any{
				"status":  "error",
				"message": "กรุณาตั้งค่า Line OA และ LIFF ID ก่อน",
			})
		}
	}

	// Generate link token
	linkToken := uuid.New().String()

	// Save link token for user
	tokenCollection := getCollection(LinkTokenCollection)
	tokenDoc := LinkTokenDoc{
		Token:       linkToken,
		HoldingCode: req.HoldingCode,
		Username:    req.Username,
		TokenType:   "user",
		ExpiresAt:   time.Now().Add(24 * time.Hour),
		CreatedAt:   time.Now(),
	}

	_, err = tokenCollection.InsertOne(ctx, tokenDoc)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Failed to generate link token",
			"code":  "TOKEN_ERROR",
		})
	}

	// Generate LIFF URL with type=user to differentiate from employee linking
	liffURL := fmt.Sprintf("https://liff.line.me/%s?token=%s&type=user&holdingcode=%s", config.LiffID, linkToken, req.HoldingCode)

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"link":   liffURL,
	})
}

// CallbackHandler - Handle callback from LIFF
func CallbackHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req CallbackRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	if req.HoldingCode == "" || req.Token == "" || req.LineUserID == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "holdingcode, token and lineuserid are required",
			"code":  "MISSING_REQUIRED_FIELDS",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Find and validate token
	tokenCollection := getCollection(LinkTokenCollection)
	var tokenDoc LinkTokenDoc
	err := tokenCollection.FindOne(ctx, bson.M{
		"token":       req.Token,
		"holdingcode": req.HoldingCode,
		"expiresat":   bson.M{"$gt": time.Now()},
		"usedat":      nil,
	}).Decode(&tokenDoc)

	if err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid or expired token",
			"code":  "INVALID_TOKEN",
		})
	}

	now := time.Now()

	// Handle based on token type
	if tokenDoc.TokenType == "user" {
		// Update user Line info - store in MongoDB
		userCollection := getCollection(UserProfileCollection)
		filter := bson.M{"holdingcode": req.HoldingCode, "username": tokenDoc.Username}
		update := bson.M{
			"$set": bson.M{
				"holdingcode":     req.HoldingCode,
				"username":        tokenDoc.Username,
				"lineuserid":      req.LineUserID,
				"linedisplayname": req.DisplayName,
				"linepictureurl":  req.PictureURL,
				"linelinkedat":    now,
				"updatedat":       now,
			},
			"$setOnInsert": bson.M{
				"createdat": now,
			},
		}
		opts := options.Update().SetUpsert(true)
		_, err = userCollection.UpdateOne(ctx, filter, update, opts)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to update user",
				"code":  "UPDATE_ERROR",
			})
		}
	} else if tokenDoc.TokenType == "employee" {
		// Update employee Line info
		empCollection := getCollection(EmployeeCollection)
		filter := bson.M{
			"lineoaconfigguid": tokenDoc.LineOAConfigGUID,
			"employeecode":     tokenDoc.EmployeeCode,
		}
		update := bson.M{
			"$set": bson.M{
				"lineuserid":  req.LineUserID,
				"displayname": req.DisplayName,
				"pictureurl":  req.PictureURL,
				"linkedat":    now,
			},
		}
		_, err = empCollection.UpdateOne(ctx, filter, update)
		if err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Failed to update employee",
				"code":  "UPDATE_ERROR",
			})
		}
	}

	// Mark token as used
	tokenCollection.UpdateOne(ctx, bson.M{"token": req.Token}, bson.M{"$set": bson.M{"usedat": now}})

	return c.JSON(http.StatusOK, map[string]any{
		"status":  "success",
		"message": "เชื่อมต่อ Line OA สำเร็จ",
	})
}

// GetUserProfileHandler - Get user's LINE profile from MongoDB
func GetUserProfileHandler(c echo.Context) error {
	if err := checkConnection(c); err != nil {
		return err
	}

	var req UserProfileRequest
	if err := c.Bind(&req); err != nil {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "Invalid request payload",
			"code":  "INVALID_PAYLOAD",
		})
	}

	if req.HoldingCode == "" || req.Username == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{
			"error": "holdingcode and username are required",
			"code":  "MISSING_REQUIRED_FIELDS",
		})
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	collection := getCollection(UserProfileCollection)
	var profile bson.M
	err := collection.FindOne(ctx, bson.M{
		"holdingcode": req.HoldingCode,
		"username":    req.Username,
	}).Decode(&profile)

	if err == mongo.ErrNoDocuments {
		return c.JSON(http.StatusOK, map[string]any{
			"status": "success",
			"data":   nil,
			"linked": false,
		})
	}
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Query failed",
			"code":  "QUERY_ERROR",
		})
	}

	return c.JSON(http.StatusOK, map[string]any{
		"status": "success",
		"data":   profile,
		"linked": true,
	})
}
