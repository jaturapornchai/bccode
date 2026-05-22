package aichat

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/goapi/aiprovider"
	"smlcloudplatform/internal/goapi/logger"
	"strings"
	"time"

	serviceConfig "smlcloudplatform/internal/goapi/config"
	myGlobal "smlcloudplatform/internal/goapi/myglobal"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const aiProviderCollection = "aiProviderConfigs"

// AIProviderConfig — config ของ AI provider ต่อ shop
type AIProviderConfig struct {
	ShopID string     `bson:"shopid" json:"shop_id"`
	ProviderName string     `bson:"provider_name" json:"provider_name"`
	APIKey string     `bson:"apikey" json:"api_key"`
	BaseURL string     `bson:"baseurl" json:"base_url"`
	Model string     `bson:"model" json:"model"`
	Capabilities []string   `bson:"capabilities" json:"capabilities"` // ["tools","vision","thinking"]
	IsActive bool       `bson:"isactive" json:"is_active"`
	Priority int        `bson:"priority" json:"priority"`
	LastError string     `bson:"lasterror" json:"last_error"`
	LastErrorAt *time.Time `bson:"last_error_at" json:"last_error_at"`
	CooldownUntil *time.Time `bson:"cooldownuntil" json:"cooldown_until"`
	CreatedAt time.Time  `bson:"created_at" json:"created_at"`
	UpdatedAt time.Time  `bson:"updated_at" json:"updated_at"`
}

// getAIProviderCollection คืน MongoDB collection
func getAIProviderCollection() (*mongo.Collection, error) {
	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("MongoDB client not available")
	}
	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	return mongoClient.Database(dbName).Collection(aiProviderCollection), nil
}

// ensureAIProviderIndex สร้าง unique index (shopid + providername)
func ensureAIProviderIndex(col *mongo.Collection) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "shopid", Value: 1},
			{Key: "provider_name", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		logger.Warn("[AIProviderDB] ensureIndex warning: %v", err)
	}
}

// getAIProviderConfigs ดึง config ทั้งหมดของ shop เรียงตาม priority
func getAIProviderConfigs(shopID string) ([]AIProviderConfig, error) {
	col, err := getAIProviderCollection()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"shopid": shopID}
	opts := options.Find().SetSort(bson.D{{Key: "priority", Value: 1}})

	cur, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("getAIProviderConfigs: %w", err)
	}
	defer cur.Close(ctx)

	var configs []AIProviderConfig
	if err := cur.All(ctx, &configs); err != nil {
		return nil, fmt.Errorf("getAIProviderConfigs decode: %w", err)
	}
	return configs, nil
}

// upsertAIProviderConfig สร้างหรืออัปเดต provider config
func upsertAIProviderConfig(shopID string, cfg AIProviderConfig) error {
	col, err := getAIProviderCollection()
	if err != nil {
		return err
	}

	ensureAIProviderIndex(col)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	cfg.ShopID = shopID
	cfg.UpdatedAt = now

	filter := bson.M{"shopid": shopID, "provider_name": cfg.ProviderName}
	setFields := bson.M{
		"model":         cfg.Model,
		"baseurl":       cfg.BaseURL,
		"capabilities":  cfg.Capabilities,
		"isactive":      cfg.IsActive,
		"priority":      cfg.Priority,
		"updated_at":     now,
		"lasterror":     "",
		"cooldownuntil": nil,
	}
	// อัปเดต apikey เฉพาะเมื่อส่งค่าจริง (ไม่ใช่ masked value)
	if cfg.APIKey != "" && !strings.Contains(cfg.APIKey, "****") {
		setFields["apikey"] = cfg.APIKey
	}
	update := bson.M{
		"$set": setFields,
		"$setOnInsert": bson.M{
			"shopid":       shopID,
			"provider_name": cfg.ProviderName,
			"created_at":    now,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err = col.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("upsertAIProviderConfig: %w", err)
	}
	// Invalidate provider cache — ให้ admin เห็นการเปลี่ยนแปลงทันที
	aiprovider.InvalidateShopProviderCache(shopID)
	logger.Info("[AIProviderDB] upsert shop=%s provider=%s", shopID, cfg.ProviderName)
	return nil
}

// deleteAIProviderConfig ลบ provider config
func deleteAIProviderConfig(shopID, providerName string) error {
	col, err := getAIProviderCollection()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := col.DeleteOne(ctx, bson.M{"shopid": shopID, "provider_name": providerName})
	if err != nil {
		return fmt.Errorf("deleteAIProviderConfig: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("provider '%s' ไม่พบในฐานข้อมูล", providerName)
	}
	aiprovider.InvalidateShopProviderCache(shopID)
	logger.Info("[AIProviderDB] deleted shop=%s provider=%s", shopID, providerName)
	return nil
}

// updateAIProviderCooldown บันทึก error + cooldown
func updateAIProviderCooldown(shopID, providerName string, errMsg string, cooldownUntil time.Time) error {
	col, err := getAIProviderCollection()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"shopid": shopID, "provider_name": providerName}
	update := bson.M{
		"$set": bson.M{
			"lasterror":     errMsg,
			"last_error_at":   time.Now(),
			"cooldownuntil": cooldownUntil,
			"updated_at":     time.Now(),
		},
	}
	_, err = col.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("updateAIProviderCooldown: %w", err)
	}
	aiprovider.InvalidateShopProviderCache(shopID)
	return nil
}

// clearAIProviderCooldown ล้าง cooldown
func clearAIProviderCooldown(shopID, providerName string) error {
	col, err := getAIProviderCollection()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"shopid": shopID, "provider_name": providerName}
	update := bson.M{
		"$set": bson.M{
			"lasterror":     "",
			"cooldownuntil": nil,
			"updated_at":     time.Now(),
		},
		"$unset": bson.M{
			"cooldownuntil": "",
		},
	}
	_, err = col.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("clearAIProviderCooldown: %w", err)
	}
	aiprovider.InvalidateShopProviderCache(shopID)
	return nil
}
