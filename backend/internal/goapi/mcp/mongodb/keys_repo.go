package mongodb

import (
	"context"
	"fmt"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mcp/tools"
	"smlcloudplatform/internal/goapi/myglobal"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	CollectionAPIKeys  = "mcp_api_keys"
	CollectionAuditLog = "mcp_audit_log"
)

// APIKey represents an MCP API Key
type APIKey struct {
	ID                 primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	APIKey             string             `bson:"api_key" json:"api_key"`
	ShopID             string             `bson:"shop_id" json:"shop_id"`
	Name               string             `bson:"name" json:"name"`
	Description        string             `bson:"description" json:"description"`
	IsActive           bool               `bson:"is_active" json:"is_active"`
	AllowedTools       []string           `bson:"allowed_tools" json:"allowed_tools"`
	RateLimitPerMinute int                `bson:"rate_limit_per_minute" json:"rate_limit_per_minute"`
	CreatedAt          time.Time          `bson:"created_at" json:"created_at"`
	ExpiresAt          *time.Time         `bson:"expires_at,omitempty" json:"expires_at,omitempty"`
	LastUsedAt         *time.Time         `bson:"last_used_at,omitempty" json:"last_used_at,omitempty"`
	CreatedBy          string             `bson:"created_by" json:"created_by"`
}

// AuditLog represents an MCP audit log entry
type AuditLog struct {
	ID              primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	APIKeyID        primitive.ObjectID `bson:"api_key_id" json:"api_key_id"`
	ShopID          string             `bson:"shop_id" json:"shop_id"`
	ToolName        string             `bson:"tool_name" json:"tool_name"`
	RequestParams   bson.M             `bson:"request_params" json:"request_params"`
	ResponseStatus  string             `bson:"response_status" json:"response_status"`
	ErrorMessage    string             `bson:"error_message,omitempty" json:"error_message,omitempty"`
	ExecutionTimeMs int64              `bson:"execution_time_ms" json:"execution_time_ms"`
	CreatedAt       time.Time          `bson:"created_at" json:"created_at"`
}

// KeysRepository handles API key operations
type KeysRepository struct {
	db *mongo.Database
}

// NewKeysRepository creates a new keys repository
func NewKeysRepository() *KeysRepository {
	return &KeysRepository{
		db: myglobal.GetMongoDatabase(),
	}
}

// CreateAPIKey creates a new API key
func (r *KeysRepository) CreateAPIKey(ctx context.Context, apiKey *APIKey) (*APIKey, error) {
	collection := r.db.Collection(CollectionAPIKeys)

	apiKey.ID = primitive.NewObjectID()
	apiKey.CreatedAt = time.Now()
	apiKey.IsActive = true

	if apiKey.RateLimitPerMinute == 0 {
		apiKey.RateLimitPerMinute = 60 // Default 60 requests per minute
	}

	_, err := collection.InsertOne(ctx, apiKey)
	if err != nil {
		logger.Error("Failed to create API key: %v", err)
		return nil, fmt.Errorf("failed to create API key: %w", err)
	}

	return apiKey, nil
}

// GetAPIKeyByKey retrieves an API key by its key string
func (r *KeysRepository) GetAPIKeyByKey(ctx context.Context, key string) (*APIKey, error) {
	collection := r.db.Collection(CollectionAPIKeys)

	var apiKey APIKey
	err := collection.FindOne(ctx, bson.M{
		"api_key":   key,
		"is_active": true,
	}).Decode(&apiKey)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		logger.Error("Failed to get API key: %v", err)
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	// Check if expired
	if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
		return nil, nil
	}

	return &apiKey, nil
}

// GetAPIKeyByID retrieves an API key by its ID
func (r *KeysRepository) GetAPIKeyByID(ctx context.Context, id string) (*APIKey, error) {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return nil, fmt.Errorf("invalid API key ID: %w", err)
	}

	collection := r.db.Collection(CollectionAPIKeys)

	var apiKey APIKey
	err = collection.FindOne(ctx, bson.M{"_id": objectID}).Decode(&apiKey)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		logger.Error("Failed to get API key by ID: %v", err)
		return nil, fmt.Errorf("failed to get API key: %w", err)
	}

	return &apiKey, nil
}

// GetAPIKeysByShop retrieves all API keys for a shop
func (r *KeysRepository) GetAPIKeysByShop(ctx context.Context, shopID string) ([]APIKey, error) {
	collection := r.db.Collection(CollectionAPIKeys)

	cursor, err := collection.Find(ctx, bson.M{
		"shop_id": shopID,
	}, options.Find().SetSort(bson.D{{Key: "created_at", Value: -1}}))

	if err != nil {
		logger.Error("Failed to get API keys by shop: %v", err)
		return nil, fmt.Errorf("failed to get API keys: %w", err)
	}
	defer cursor.Close(ctx)

	var apiKeys []APIKey
	if err := cursor.All(ctx, &apiKeys); err != nil {
		logger.Error("Failed to decode API keys: %v", err)
		return nil, fmt.Errorf("failed to decode API keys: %w", err)
	}

	return apiKeys, nil
}

// UpdateAPIKey updates an API key
func (r *KeysRepository) UpdateAPIKey(ctx context.Context, id string, updates bson.M) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid API key ID: %w", err)
	}

	collection := r.db.Collection(CollectionAPIKeys)

	updates["updated_at"] = time.Now()

	_, err = collection.UpdateOne(ctx,
		bson.M{"_id": objectID},
		bson.M{"$set": updates},
	)

	if err != nil {
		logger.Error("Failed to update API key: %v", err)
		return fmt.Errorf("failed to update API key: %w", err)
	}

	return nil
}

// DeleteAPIKey soft deletes an API key by setting is_active to false
func (r *KeysRepository) DeleteAPIKey(ctx context.Context, id string) error {
	objectID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		return fmt.Errorf("invalid API key ID: %w", err)
	}

	collection := r.db.Collection(CollectionAPIKeys)

	_, err = collection.UpdateOne(ctx,
		bson.M{"_id": objectID},
		bson.M{
			"$set": bson.M{
				"is_active":  false,
				"updated_at": time.Now(),
			},
		},
	)

	if err != nil {
		logger.Error("Failed to delete API key: %v", err)
		return fmt.Errorf("failed to delete API key: %w", err)
	}

	return nil
}

// UpdateLastUsedAt updates the last used timestamp for an API key
func (r *KeysRepository) UpdateLastUsedAt(ctx context.Context, key string) error {
	collection := r.db.Collection(CollectionAPIKeys)
	now := time.Now()

	_, err := collection.UpdateOne(ctx,
		bson.M{"api_key": key},
		bson.M{"$set": bson.M{"last_used_at": now}},
	)

	if err != nil {
		logger.Error("Failed to update last used at: %v", err)
		return fmt.Errorf("failed to update last used at: %w", err)
	}

	return nil
}

// CreateAuditLog creates an audit log entry
func (r *KeysRepository) CreateAuditLog(ctx context.Context, log *AuditLog) error {
	collection := r.db.Collection(CollectionAuditLog)

	log.ID = primitive.NewObjectID()
	log.CreatedAt = time.Now()

	_, err := collection.InsertOne(ctx, log)
	if err != nil {
		logger.Error("Failed to create audit log: %v", err)
		return fmt.Errorf("failed to create audit log: %w", err)
	}

	return nil
}

// GetAuditLogsByShop retrieves audit logs for a shop
func (r *KeysRepository) GetAuditLogsByShop(ctx context.Context, shopID string, limit int64, skip int64) ([]AuditLog, error) {
	collection := r.db.Collection(CollectionAuditLog)

	opts := options.Find().
		SetSort(bson.D{{Key: "created_at", Value: -1}}).
		SetLimit(limit).
		SetSkip(skip)

	cursor, err := collection.Find(ctx, bson.M{"shop_id": shopID}, opts)
	if err != nil {
		logger.Error("Failed to get audit logs: %v", err)
		return nil, fmt.Errorf("failed to get audit logs: %w", err)
	}
	defer cursor.Close(ctx)

	var logs []AuditLog
	if err := cursor.All(ctx, &logs); err != nil {
		logger.Error("Failed to decode audit logs: %v", err)
		return nil, fmt.Errorf("failed to decode audit logs: %w", err)
	}

	return logs, nil
}

// IsToolAllowed checks if a tool is in the allowed tools list
func (k *APIKey) IsToolAllowed(toolName string) bool {
	return tools.IsToolAllowedByList(toolName, k.AllowedTools)
}

// EnsureIndexes creates necessary indexes for the collections
func (r *KeysRepository) EnsureIndexes(ctx context.Context) error {
	// API Keys indexes
	keysCollection := r.db.Collection(CollectionAPIKeys)

	indexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "api_key", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "shop_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "is_active", Value: 1}},
		},
		{
			Keys: bson.D{
				{Key: "api_key", Value: 1},
				{Key: "is_active", Value: 1},
			},
		},
	}

	_, err := keysCollection.Indexes().CreateMany(ctx, indexes)
	if err != nil {
		logger.Error("Failed to create API keys indexes: %v", err)
		return fmt.Errorf("failed to create indexes: %w", err)
	}

	// Audit Log indexes
	auditCollection := r.db.Collection(CollectionAuditLog)

	auditIndexes := []mongo.IndexModel{
		{
			Keys: bson.D{{Key: "api_key_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "shop_id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "created_at", Value: -1}},
		},
		{
			Keys: bson.D{
				{Key: "shop_id", Value: 1},
				{Key: "created_at", Value: -1},
			},
		},
	}

	_, err = auditCollection.Indexes().CreateMany(ctx, auditIndexes)
	if err != nil {
		logger.Error("Failed to create audit log indexes: %v", err)
		return fmt.Errorf("failed to create audit indexes: %w", err)
	}

	logger.Success("MCP indexes created successfully")
	return nil
}
