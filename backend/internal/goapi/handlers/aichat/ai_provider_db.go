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

const aiProviderCollection = "aiproviderconfigs"

// AIProviderConfig — config ของ AI provider ต่อ shop
type AIProviderConfig struct {
	HoldingCode   string     `bson:"holdingcode" json:"holdingcode"`
	ProviderName  string     `bson:"providername" json:"providername"`
	APIKey        string     `bson:"apikey" json:"apikey"`
	BaseURL       string     `bson:"baseurl" json:"baseurl"`
	Model         string     `bson:"model" json:"model"`
	Capabilities  []string   `bson:"capabilities" json:"capabilities"` // ["tools","vision","thinking"]
	IsActive      bool       `bson:"isactive" json:"isactive"`
	Priority      int        `bson:"priority" json:"priority"`
	LastError     string     `bson:"lasterror" json:"lasterror"`
	LastErrorAt   *time.Time `bson:"lasterrorat" json:"lasterrorat"`
	CooldownUntil *time.Time `bson:"cooldownuntil" json:"cooldownuntil"`
	CreatedAt     time.Time  `bson:"createdat" json:"createdat"`
	UpdatedAt     time.Time  `bson:"updatedat" json:"updatedat"`
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

// ensureAIProviderIndex สร้าง unique index (holdingcode + providername)
func ensureAIProviderIndex(col *mongo.Collection) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := col.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{
			{Key: "holdingcode", Value: 1},
			{Key: "providername", Value: 1},
		},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		logger.Warn("[AIProviderDB] ensureIndex warning: %v", err)
	}
}

// getAIProviderConfigs ดึง config ทั้งหมดของ shop เรียงตาม priority
func getAIProviderConfigs(holdingCode string) ([]AIProviderConfig, error) {
	col, err := getAIProviderCollection()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"holdingcode": holdingCode}
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
func upsertAIProviderConfig(holdingCode string, cfg AIProviderConfig) error {
	col, err := getAIProviderCollection()
	if err != nil {
		return err
	}

	ensureAIProviderIndex(col)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	now := time.Now()
	cfg.HoldingCode = holdingCode
	cfg.UpdatedAt = now

	filter := bson.M{"holdingcode": holdingCode, "providername": cfg.ProviderName}
	setFields := bson.M{
		"model":         cfg.Model,
		"baseurl":       cfg.BaseURL,
		"capabilities":  cfg.Capabilities,
		"isactive":      cfg.IsActive,
		"priority":      cfg.Priority,
		"updatedat":     now,
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
			"holdingcode":  holdingCode,
			"providername": cfg.ProviderName,
			"createdat":    now,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err = col.UpdateOne(ctx, filter, update, opts)
	if err != nil {
		return fmt.Errorf("upsertAIProviderConfig: %w", err)
	}
	// Invalidate provider cache — ให้ admin เห็นการเปลี่ยนแปลงทันที
	aiprovider.InvalidateShopProviderCache(holdingCode)
	logger.Info("[AIProviderDB] upsert shop=%s provider=%s", holdingCode, cfg.ProviderName)
	return nil
}

// deleteAIProviderConfig ลบ provider config
func deleteAIProviderConfig(holdingCode, providerName string) error {
	col, err := getAIProviderCollection()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	result, err := col.DeleteOne(ctx, bson.M{"holdingcode": holdingCode, "providername": providerName})
	if err != nil {
		return fmt.Errorf("deleteAIProviderConfig: %w", err)
	}
	if result.DeletedCount == 0 {
		return fmt.Errorf("provider '%s' ไม่พบในฐานข้อมูล", providerName)
	}
	aiprovider.InvalidateShopProviderCache(holdingCode)
	logger.Info("[AIProviderDB] deleted shop=%s provider=%s", holdingCode, providerName)
	return nil
}

// updateAIProviderCooldown บันทึก error + cooldown
func updateAIProviderCooldown(holdingCode, providerName string, errMsg string, cooldownUntil time.Time) error {
	col, err := getAIProviderCollection()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"holdingcode": holdingCode, "providername": providerName}
	update := bson.M{
		"$set": bson.M{
			"lasterror":     errMsg,
			"lasterrorat":   time.Now(),
			"cooldownuntil": cooldownUntil,
			"updatedat":     time.Now(),
		},
	}
	_, err = col.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("updateAIProviderCooldown: %w", err)
	}
	aiprovider.InvalidateShopProviderCache(holdingCode)
	return nil
}

// clearAIProviderCooldown ล้าง cooldown
func clearAIProviderCooldown(holdingCode, providerName string) error {
	col, err := getAIProviderCollection()
	if err != nil {
		return err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"holdingcode": holdingCode, "providername": providerName}
	update := bson.M{
		"$set": bson.M{
			"lasterror":     "",
			"cooldownuntil": nil,
			"updatedat":     time.Now(),
		},
		"$unset": bson.M{
			"cooldownuntil": "",
		},
	}
	_, err = col.UpdateOne(ctx, filter, update)
	if err != nil {
		return fmt.Errorf("clearAIProviderCooldown: %w", err)
	}
	aiprovider.InvalidateShopProviderCache(holdingCode)
	return nil
}
