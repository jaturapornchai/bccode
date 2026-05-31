package tools

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/myollama"
	"smlcloudplatform/internal/goapi/mypg"
)

// ==================== Rebuild Embeddings ====================
// สร้าง vector embeddings สำหรับ semantic search
// รองรับ: product, debtor, creditor, customer

const embeddingBatchSize = 50

// RebuildEmbeddingsResponse ผลลัพธ์
type RebuildEmbeddingsResponse struct {
	Success     bool      `json:"success"`
	Message     string    `json:"message"`
	ShopID      string    `json:"shop_id"`
	EntityType  string    `json:"entity_type"`
	Total       int       `json:"total"`
	Updated     int       `json:"updated"`
	Skipped     int       `json:"skipped"`
	Errors      int       `json:"errors"`
	Duration    string    `json:"duration"`
	GeneratedAt time.Time `json:"generated_at"`
}

// embedRow — generic row สำหรับ embedding (key เป็น string เพราะ debtor/creditor ใช้ guidfixed)
type embedRow struct {
	Key  string // id (as string) for product/customer, guidfixed for debtor/creditor
	Text string
}

// RebuildEmbeddings builds pgvector embeddings on PostgreSQL projection tables.
// MongoDB remains the authoritative operational source for entity data.
func RebuildEmbeddings(ctx context.Context, shopID string, forceAll bool, entityType string) (*RebuildEmbeddingsResponse, error) {
	if shopID == "" {
		return nil, fmt.Errorf("shop_id is required")
	}
	if entityType == "" {
		entityType = "product"
	}

	start := time.Now()
	logger.Info("[Embeddings] เริ่มสร้าง embeddings สำหรับ %s shop=%s (forceAll=%v)", entityType, shopID, forceAll)

	db, err := mypg.PgSqlFastConnect(shopID)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	if err := ensureVectorExtension(db); err != nil {
		return nil, fmt.Errorf("pgvector not available: %w", err)
	}

	cfg, err := getEntityConfig(entityType)
	if err != nil {
		return nil, err
	}

	ensureEmbeddingColumn(db, cfg.tableName)

	rows, err := fetchRowsForEmbedding(db, cfg, forceAll)
	if err != nil {
		return nil, fmt.Errorf("fetch %s failed: %w", entityType, err)
	}

	total := len(rows)
	if total == 0 {
		return &RebuildEmbeddingsResponse{
			Success:     true,
			Message:     fmt.Sprintf("ไม่มี %s ที่ต้องสร้าง embedding", entityType),
			ShopID:      shopID,
			EntityType:  entityType,
			Total:       0,
			GeneratedAt: time.Now(),
		}, nil
	}

	logger.Info("[Embeddings] พบ %d %s ที่ต้องสร้าง embedding", total, entityType)

	updated, skipped, errors := 0, 0, 0

	for i := 0; i < total; i += embeddingBatchSize {
		if ctx.Err() != nil {
			return nil, fmt.Errorf("cancelled: %w", ctx.Err())
		}

		end := i + embeddingBatchSize
		if end > total {
			end = total
		}
		batch := rows[i:end]

		texts := make([]string, len(batch))
		for j, row := range batch {
			texts[j] = row.Text
		}

		embeddings, err := myollama.GenerateEmbeddings(texts)
		if err != nil {
			logger.Error("[Embeddings] Ollama error batch %d-%d: %v", i, end, err)
			errors += len(batch)
			continue
		}

		for j, emb := range embeddings {
			if len(emb) == 0 {
				skipped++
				continue
			}
			if err := updateEmbedding(db, cfg, batch[j].Key, emb); err != nil {
				logger.Error("[Embeddings] update key=%s error: %v", batch[j].Key, err)
				errors++
				continue
			}
			updated++
		}

		logger.Info("[Embeddings] %s batch %d-%d/%d done (updated=%d)", entityType, i, end, total, updated)
	}

	duration := time.Since(start).Round(time.Millisecond)
	msg := fmt.Sprintf("สร้าง %s embedding สำเร็จ: %d/%d (skip=%d, errors=%d, %s)",
		entityType, updated, total, skipped, errors, duration)
	logger.Info("[Embeddings] %s", msg)

	return &RebuildEmbeddingsResponse{
		Success:     true,
		Message:     msg,
		ShopID:      shopID,
		EntityType:  entityType,
		Total:       total,
		Updated:     updated,
		Skipped:     skipped,
		Errors:      errors,
		Duration:    duration.String(),
		GeneratedAt: time.Now(),
	}, nil
}

// ==================== Entity Config ====================

type entityConfig struct {
	tableName string
	keyColumn string // column ที่ใช้เป็น key: "id" for product/customer, "guid_fixed" for debtor/creditor
	query     string // SQL: SELECT key, text ... WHERE name_embedding IS NULL
	queryAll  string // SQL: SELECT key, text ... (force all)
}

func getEntityConfig(entityType string) (*entityConfig, error) {
	switch entityType {
	case "product":
		return &entityConfig{
			tableName: "productbarcode",
			keyColumn: "id",
			query: `SELECT id::text, CONCAT_WS(' | ', name0, NULLIF(brandnames,''), NULLIF(categorynames,''), NULLIF(groupnames,''))
				FROM (SELECT DISTINCT ON (itemcode) id, itemcode, name0, brandnames, categorynames, groupnames, name_embedding
					FROM productbarcode ORDER BY itemcode, id) sub
				WHERE name0 != '' AND name_embedding IS NULL ORDER BY itemcode`,
			queryAll: `SELECT id::text, CONCAT_WS(' | ', name0, NULLIF(brandnames,''), NULLIF(categorynames,''), NULLIF(groupnames,''))
				FROM (SELECT DISTINCT ON (itemcode) id, itemcode, name0, brandnames, categorynames, groupnames
					FROM productbarcode ORDER BY itemcode, id) sub
				WHERE name0 != '' ORDER BY itemcode`,
		}, nil

	case "debtor":
		return &entityConfig{
			tableName: "debtor",
			keyColumn: "guid_fixed",
			query: `SELECT d.guidfixed, CONCAT_WS(' | ', d.code, string_agg(n.elem->>'name', ' / '))
				FROM debtor d, jsonb_array_elements(d.names) AS n(elem)
				WHERE d.names IS NOT NULL AND jsonb_array_length(d.names) > 0 AND d.name_embedding IS NULL
				GROUP BY d.guidfixed, d.code`,
			queryAll: `SELECT d.guidfixed, CONCAT_WS(' | ', d.code, string_agg(n.elem->>'name', ' / '))
				FROM debtor d, jsonb_array_elements(d.names) AS n(elem)
				WHERE d.names IS NOT NULL AND jsonb_array_length(d.names) > 0
				GROUP BY d.guidfixed, d.code`,
		}, nil

	case "creditor":
		return &entityConfig{
			tableName: "creditor",
			keyColumn: "guid_fixed",
			query: `SELECT d.guidfixed, CONCAT_WS(' | ', d.code, string_agg(n.elem->>'name', ' / '))
				FROM creditor d, jsonb_array_elements(d.names) AS n(elem)
				WHERE d.names IS NOT NULL AND jsonb_array_length(d.names) > 0 AND d.name_embedding IS NULL
				GROUP BY d.guidfixed, d.code`,
			queryAll: `SELECT d.guidfixed, CONCAT_WS(' | ', d.code, string_agg(n.elem->>'name', ' / '))
				FROM creditor d, jsonb_array_elements(d.names) AS n(elem)
				WHERE d.names IS NOT NULL AND jsonb_array_length(d.names) > 0
				GROUP BY d.guidfixed, d.code`,
		}, nil

	case "customer":
		return &entityConfig{
			tableName: "customer",
			keyColumn: "id",
			query: `SELECT id::text, CONCAT_WS(' | ', code, name0)
				FROM customer
				WHERE name0 IS NOT NULL AND name0 != '' AND name_embedding IS NULL`,
			queryAll: `SELECT id::text, CONCAT_WS(' | ', code, name0)
				FROM customer
				WHERE name0 IS NOT NULL AND name0 != ''`,
		}, nil

	default:
		return nil, fmt.Errorf("unsupported entity_type: %s (ใช้ได้: product, debtor, creditor, customer)", entityType)
	}
}

// ==================== Helpers ====================

func ensureVectorExtension(db *sql.DB) error {
	var exists bool
	err := db.QueryRow(`SELECT EXISTS(SELECT 1 FROM pg_extension WHERE extname = 'vector')`).Scan(&exists)
	if err != nil {
		return fmt.Errorf("check pgvector: %w", err)
	}
	if !exists {
		if _, err := db.Exec(`CREATE EXTENSION IF NOT EXISTS vector`); err != nil {
			return fmt.Errorf("pgvector extension ไม่สามารถสร้างได้ — ต้องติดตั้ง pgvector binary ก่อน: %w", err)
		}
	}
	return nil
}

func ensureEmbeddingColumn(db *sql.DB, tableName string) {
	var colExists bool
	db.QueryRow(`SELECT EXISTS(SELECT 1 FROM information_schema.columns WHERE table_name = $1 AND column_name = 'name_embedding')`, tableName).Scan(&colExists)
	if !colExists {
		db.Exec(fmt.Sprintf(`ALTER TABLE %s ADD COLUMN IF NOT EXISTS name_embedding vector(768)`, tableName))
		db.Exec(fmt.Sprintf(`CREATE INDEX IF NOT EXISTS idx_%s_embedding_hnsw ON %s USING hnsw (name_embedding vector_cosine_ops) WITH (m = 16, ef_construction = 64)`, tableName, tableName))
	}
}

func fetchRowsForEmbedding(db *sql.DB, cfg *entityConfig, forceAll bool) ([]embedRow, error) {
	query := cfg.query
	if forceAll {
		query = cfg.queryAll
	}

	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []embedRow
	for rows.Next() {
		var r embedRow
		if err := rows.Scan(&r.Key, &r.Text); err != nil {
			continue
		}
		if r.Text != "" {
			result = append(result, r)
		}
	}
	return result, rows.Err()
}

func updateEmbedding(db *sql.DB, cfg *entityConfig, key string, embedding []float32) error {
	vecStr := float32SliceToVectorString(embedding)
	query := fmt.Sprintf(`UPDATE %s SET name_embedding = $1::vector WHERE %s = $2`, cfg.tableName, cfg.keyColumn)
	_, err := db.Exec(query, vecStr, key)
	return err
}

// float32SliceToVectorString แปลง []float32 → "[0.1,0.2,...]" สำหรับ pgvector
func float32SliceToVectorString(v []float32) string {
	var sb strings.Builder
	sb.WriteByte('[')
	for i, f := range v {
		if i > 0 {
			sb.WriteByte(',')
		}
		fmt.Fprintf(&sb, "%g", f)
	}
	sb.WriteByte(']')
	return sb.String()
}

// RebuildAllEmbeddings สร้าง embeddings ทุก entity type ในรอบเดียว
func RebuildAllEmbeddings(ctx context.Context, shopID string, forceAll bool) ([]RebuildEmbeddingsResponse, error) {
	types := []string{"product", "debtor", "creditor", "customer"}
	var results []RebuildEmbeddingsResponse

	for _, t := range types {
		resp, err := RebuildEmbeddings(ctx, shopID, forceAll, t)
		if err != nil {
			logger.Warn("[Embeddings] %s failed: %v", t, err)
			continue
		}
		results = append(results, *resp)
	}

	return results, nil
}

// RebuildAllEmbeddingsJSON — wrapper ที่ return JSON string
func RebuildAllEmbeddingsJSON(ctx context.Context, shopID string, forceAll bool) (string, error) {
	results, err := RebuildAllEmbeddings(ctx, shopID, forceAll)
	if err != nil {
		return "", err
	}
	b, _ := json.Marshal(results)
	return string(b), nil
}
