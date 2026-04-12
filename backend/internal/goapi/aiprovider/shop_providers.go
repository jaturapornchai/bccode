package aiprovider

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/goapi/logger"
	"strings"
	"sync"
	"time"

	serviceConfig "smlcloudplatform/internal/goapi/config"
	myGlobal "smlcloudplatform/internal/goapi/myglobal"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const shopAIProviderCollection = "aiProviderConfigs"

// shopProviderDoc — document ใน MongoDB collection aiProviderConfigs
type shopProviderDoc struct {
	ShopID        string     `bson:"shopid"`
	ProviderName  string     `bson:"providername"`
	APIKey        string     `bson:"apikey"`
	BaseURL       string     `bson:"baseurl"`
	Model         string     `bson:"model"`
	IsActive      bool       `bson:"isactive"`
	Priority      int        `bson:"priority"`
	CooldownUntil *time.Time `bson:"cooldownuntil"`
}

// ====== In-memory cache สำหรับ shop provider docs ======
//
// เหตุผล: ทุก chatbot request เรียก GetShopProviders → loadShopProviderDocs → MongoDB round-trip (~5-50ms)
// สำหรับ OpenClaw gateway ที่เรียก compactor + agent = 2 round-trips ต่อ request
// TTL 2 นาที — สั้นพอที่ admin เปลี่ยน config แล้วเห็นผลเร็ว + บันทึก cooldown invalidate ทันที
//
// Safety: ถ้า MongoDB fail / doc เปลี่ยน → cache miss ยิง query ใหม่ตาม TTL
// Invalidation: MarkProviderFailed + SaveAIProvider จะ invalidate cache ของ shop นั้น
const shopProviderCacheTTL = 2 * time.Minute

type cachedShopProviderDocs struct {
	docs      []shopProviderDoc
	loadedAt  time.Time
}

var (
	shopProviderCache   = map[string]*cachedShopProviderDocs{}
	shopProviderCacheMu sync.RWMutex
)

// getCachedShopProviderDocs — คืน docs จาก cache ถ้ายังไม่หมดอายุ
func getCachedShopProviderDocs(shopID string) ([]shopProviderDoc, bool) {
	shopProviderCacheMu.RLock()
	defer shopProviderCacheMu.RUnlock()
	entry, ok := shopProviderCache[shopID]
	if !ok {
		return nil, false
	}
	if time.Since(entry.loadedAt) > shopProviderCacheTTL {
		return nil, false
	}
	// return copy เพื่อป้องกัน caller mutate
	out := make([]shopProviderDoc, len(entry.docs))
	copy(out, entry.docs)
	return out, true
}

// putCachedShopProviderDocs — เก็บ docs ลง cache
func putCachedShopProviderDocs(shopID string, docs []shopProviderDoc) {
	shopProviderCacheMu.Lock()
	defer shopProviderCacheMu.Unlock()
	stored := make([]shopProviderDoc, len(docs))
	copy(stored, docs)
	shopProviderCache[shopID] = &cachedShopProviderDocs{
		docs:     stored,
		loadedAt: time.Now(),
	}
	// opportunistic cleanup — ป้องกัน memory bloat
	if len(shopProviderCache) > 1000 {
		cutoff := time.Now().Add(-shopProviderCacheTTL * 2)
		for k, v := range shopProviderCache {
			if v.loadedAt.Before(cutoff) {
				delete(shopProviderCache, k)
			}
		}
	}
}

// InvalidateShopProviderCache — เรียกเมื่อ config เปลี่ยน (save/delete/cooldown)
func InvalidateShopProviderCache(shopID string) {
	if shopID == "" {
		return
	}
	shopProviderCacheMu.Lock()
	defer shopProviderCacheMu.Unlock()
	delete(shopProviderCache, shopID)
}

// shopProviderBaseURLs — base URL ต่อ provider (OpenAI-compatible)
var shopProviderBaseURLs = map[string]string{
	"openrouter": "https://openrouter.ai/api/v1/chat/completions",
	"groq":       "https://api.groq.com/openai/v1/chat/completions",
	"deepseek":   "https://api.deepseek.com/v1/chat/completions",
	"ollama":     "http://host.docker.internal:11434/v1/chat/completions",
}

// shopProviderDefaultModels — default model ต่อ provider
var shopProviderDefaultModels = map[string]string{
	"openrouter": "google/gemini-2.0-flash-exp:free",
	"groq":       "llama-3.3-70b-versatile",
	"deepseek":   "deepseek-chat",
	"gemini":     "gemini-2.0-flash",
	"ollama":     "gemma4:e4b-max",
	"custom":     "bcproxy/tools", // bcproxyai auto-route ไป model ที่รองรับ tool calling + vision
}

// loadShopProviderDocs ดึง active provider docs ของ shop จาก MongoDB
// มี in-memory cache (TTL 2 นาที) เพื่อลด MongoDB round-trip ใน chatbot hot path
func loadShopProviderDocs(shopID string) ([]shopProviderDoc, error) {
	// Cache hit — คืนทันทีไม่ยิง MongoDB
	if cached, ok := getCachedShopProviderDocs(shopID); ok {
		return cached, nil
	}

	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		return nil, fmt.Errorf("MongoDB client not available")
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	col := mongoClient.Database(dbName).Collection(shopAIProviderCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	filter := bson.M{"shopid": shopID, "isactive": true}
	opts := options.Find().SetSort(bson.D{{Key: "priority", Value: 1}})

	cur, err := col.Find(ctx, filter, opts)
	if err != nil {
		return nil, fmt.Errorf("loadShopProviderDocs: %w", err)
	}
	defer cur.Close(ctx)

	var docs []shopProviderDoc
	if err := cur.All(ctx, &docs); err != nil {
		return nil, fmt.Errorf("loadShopProviderDocs decode: %w", err)
	}

	// เก็บ cache ไว้เฉพาะกรณี query สำเร็จ (อาจเป็น empty list ก็ cache ได้ — บอกว่า shop นี้ไม่มี config)
	putCachedShopProviderDocs(shopID, docs)
	return docs, nil
}

// updateShopProviderCooldownDB บันทึก cooldown กลับ MongoDB
func updateShopProviderCooldownDB(shopID, providerName, errMsg string, cooldownUntil time.Time) {
	mongoClient := myGlobal.SafeMongoConnectFast()
	if mongoClient == nil {
		logger.Error("[ShopProvider] MongoDB not available for cooldown update")
		return
	}

	svcConfig := serviceConfig.NewServiceConfig()
	dbName := svcConfig.MongodbDatabaseName()
	col := mongoClient.Database(dbName).Collection(shopAIProviderCollection)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	filter := bson.M{"shopid": shopID, "providername": providerName}
	update := bson.M{
		"$set": bson.M{
			"lasterror":     errMsg,
			"lasterrorat":   time.Now(),
			"cooldownuntil": cooldownUntil,
			"updatedat":     time.Now(),
		},
	}

	if _, err := col.UpdateOne(ctx, filter, update); err != nil {
		logger.Error("[ShopProvider] updateCooldown error: %v", err)
	}

	// Invalidate cache — ให้ request ถัดไปเห็น cooldown ใหม่ทันที
	InvalidateShopProviderCache(shopID)
}

// GetShopProviders คืน providers สำหรับ shop จาก MongoDB — fallback to env vars ถ้าไม่มี config
func GetShopProviders(shopID string) []providerEntry {
	if shopID == "" {
		return getAvailableProviders()
	}

	docs, err := loadShopProviderDocs(shopID)
	if err != nil {
		logger.Warn("[ShopProvider] loadDocs failed for shop=%s, falling back to env: %v", shopID, err)
		return getAvailableProviders()
	}

	if len(docs) == 0 {
		logger.Info("[ShopProvider] no DB config for shop=%s, using env fallback", shopID)
		return getAvailableProviders()
	}

	now := time.Now()
	singleProvider := len(docs) == 1 // ถ้ามี provider เดียว → ไม่ cooldown (ไม่มี fallback)
	var providers []providerEntry

	for _, doc := range docs {
		// skip cooldown — แต่ถ้ามี provider เดียว ข้าม cooldown ไป (ไม่งั้นใช้ไม่ได้เลย)
		if !singleProvider && doc.CooldownUntil != nil && doc.CooldownUntil.After(now) {
			logger.Info("[ShopProvider] %s is in cooldown until %v (shop=%s)", doc.ProviderName, doc.CooldownUntil, shopID)
			continue
		}

		model := doc.Model
		if model == "" {
			model = shopProviderDefaultModels[doc.ProviderName]
			// custom provider fallback → bcproxy/tools
			if model == "" && strings.HasPrefix(doc.ProviderName, "custom") {
				model = shopProviderDefaultModels["custom"]
			}
		}

		if doc.ProviderName == "gemini" {
			providers = append(providers, providerEntry{
				name:     "gemini",
				provider: newGeminiProvider(),
			})
			continue
		}

		// Custom provider — ใช้ base_url จาก config
		if strings.HasPrefix(doc.ProviderName, "custom") {
			if doc.BaseURL == "" {
				logger.Warn("[ShopProvider] custom provider '%s' has no base_url, skipped (shop=%s)", doc.ProviderName, shopID)
				continue
			}
			// Custom provider — ต่อ /chat/completions ตามมาตรฐาน OpenAI
			chatURL := doc.BaseURL
			if !strings.HasSuffix(chatURL, "/chat/completions") {
				chatURL = strings.TrimRight(chatURL, "/") + "/chat/completions"
			}
			providers = append(providers, providerEntry{
				name:     doc.ProviderName,
				provider: newOpenAICompatProvider(doc.ProviderName, chatURL, doc.APIKey, model),
			})
			continue
		}

		baseURL, ok := shopProviderBaseURLs[doc.ProviderName]
		if !ok {
			logger.Warn("[ShopProvider] unknown provider '%s' skipped (shop=%s)", doc.ProviderName, shopID)
			continue
		}

		// Ollama: ถ้า user ตั้ง base_url เอง (เครื่องอื่น) → ใช้ค่านั้นแทน default
		if doc.ProviderName == "ollama" && doc.BaseURL != "" {
			chatURL := strings.TrimRight(doc.BaseURL, "/") + "/chat/completions"
			baseURL = chatURL
		}

		providers = append(providers, providerEntry{
			name:     doc.ProviderName,
			provider: newOpenAICompatProvider(doc.ProviderName, baseURL, doc.APIKey, model),
		})
	}

	if len(providers) == 0 {
		logger.Warn("[ShopProvider] all DB providers cooled down for shop=%s, falling back to env", shopID)
		return getAvailableProviders()
	}

	logger.Info("[ShopProvider] shop=%s loaded %d provider(s) from DB", shopID, len(providers))
	return providers
}

// MarkProviderFailed บันทึก cooldown เมื่อ provider fail — ทั้ง in-memory + MongoDB
func MarkProviderFailed(shopID, providerName string, err error) {
	cd := defaultCooldown
	if isRateLimitError(err) {
		cd = rateLimitCooldown
	}

	// in-memory cooldown (สำหรับ env-var fallback path)
	markFailed(providerName, cd)

	// MongoDB cooldown (สำหรับ shop-specific path)
	if shopID != "" {
		cooldownUntil := time.Now().Add(cd)
		updateShopProviderCooldownDB(shopID, providerName, err.Error(), cooldownUntil)
	}
}

// GetShopToolCallingProviders คืน tool-calling providers สำหรับ shop
func GetShopToolCallingProviders(shopID string) []ToolCallingProvider {
	entries := GetShopProviders(shopID)
	var result []ToolCallingProvider
	for _, p := range entries {
		if tc, ok := p.provider.(ToolCallingProvider); ok {
			result = append(result, tc)
		}
	}
	return result
}

// GetShopAIProviders คืน AIProvider list (text generation) สำหรับ shop
// ใช้สำหรับงาน text-only เช่น summarization, classification ที่ไม่ต้อง tool calling
func GetShopAIProviders(shopID string) []AIProvider {
	entries := GetShopProviders(shopID)
	result := make([]AIProvider, 0, len(entries))
	for _, p := range entries {
		result = append(result, p.provider)
	}
	return result
}
