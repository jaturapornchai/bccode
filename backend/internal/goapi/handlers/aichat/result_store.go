package aichat

// Result Store — เก็บ tool result เต็มไว้ใน memory เพื่อให้ frontend ดึงไปแสดงครบ
//
// ปัญหา: tool บางตัวคืนข้อมูลใหญ่ (เช่น 100 rows) → ส่งเข้า LLM context ทั้งหมดไม่ไหว
// ต้องตัด → AI ตอบไม่ครบ + ผู้ใช้เห็นแค่บางส่วน
//
// แนวคิด:
//  1. Wrap layer: ถ้า raw result ใหญ่กว่า threshold → store full → ได้ result_id
//  2. ส่งให้ LLM แค่ summary { rowCount, preview, result_id, see_full_url }
//  3. Frontend ดึง full data ผ่าน GET /goapi/api/aichat/result/:id
//  4. AI ใส่ลิงก์ markdown `[ดูทั้งหมด N รายการ](.../result/xxx)` ในคำตอบ
//
// Storage: in-memory, TTL 10 นาที, max 200 entries (LRU eviction)

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

const (
	resultStoreTTL          = 10 * time.Minute
	resultStoreMaxEntries   = 200
	resultStoreLargeThreshold = 6000 // bytes ของ JSON เต็ม — ถ้าเกินนี้ → store + ส่ง preview
	resultStorePreviewItems   = 10   // จำนวน rows ที่ส่งเป็น preview ให้ LLM
)

type resultEntry struct {
	id        string
	tool      string
	data      any
	rowCount  int
	storedAt  time.Time
}

var (
	resultStore   = map[string]*resultEntry{}
	resultStoreMu sync.RWMutex
)

// generateResultID — สร้าง random hex id (16 chars)
func generateResultID() string {
	buf := make([]byte, 8)
	if _, err := rand.Read(buf); err != nil {
		// fallback ใช้ timestamp
		return "ts" + time.Now().Format("150405.000")
	}
	return hex.EncodeToString(buf)
}

// PutResult เก็บ full result + คืน id
func PutResult(tool string, data any, rowCount int) string {
	id := generateResultID()
	resultStoreMu.Lock()
	defer resultStoreMu.Unlock()

	resultStore[id] = &resultEntry{
		id:       id,
		tool:     tool,
		data:     data,
		rowCount: rowCount,
		storedAt: time.Now(),
	}

	// Cleanup ถ้าเกิน max
	if len(resultStore) > resultStoreMaxEntries {
		cleanupResultStore()
	}

	return id
}

// GetResult ดึง full result ตาม id (คืน nil ถ้าไม่เจอหรือหมดอายุ)
func GetResult(id string) (any, bool) {
	resultStoreMu.RLock()
	defer resultStoreMu.RUnlock()

	entry, ok := resultStore[id]
	if !ok {
		return nil, false
	}
	if time.Since(entry.storedAt) > resultStoreTTL {
		return nil, false
	}
	return entry.data, true
}

// GetResultMeta ดึง metadata + data
func GetResultMeta(id string) (*resultEntry, bool) {
	resultStoreMu.RLock()
	defer resultStoreMu.RUnlock()
	entry, ok := resultStore[id]
	if !ok {
		return nil, false
	}
	if time.Since(entry.storedAt) > resultStoreTTL {
		return nil, false
	}
	return entry, true
}

// cleanupResultStore — ลบรายการเก่าที่หมดอายุ + ถ้ายังเกิน max ลบที่เก่าที่สุด
// must hold resultStoreMu.Lock
func cleanupResultStore() {
	cutoff := time.Now().Add(-resultStoreTTL)
	for k, v := range resultStore {
		if v.storedAt.Before(cutoff) {
			delete(resultStore, k)
		}
	}
	// ถ้ายังเกิน → ลบที่เก่าสุดออกจนเหลือ max
	for len(resultStore) > resultStoreMaxEntries {
		var oldestKey string
		var oldestTime time.Time
		first := true
		for k, v := range resultStore {
			if first || v.storedAt.Before(oldestTime) {
				oldestKey = k
				oldestTime = v.storedAt
				first = false
			}
		}
		delete(resultStore, oldestKey)
	}
}

// countRows — เดาจำนวน row จาก result (รองรับ slice/array/map ที่มี field rows/data/items/results)
func countRows(data any) int {
	if data == nil {
		return 0
	}
	switch v := data.(type) {
	case []any:
		return len(v)
	case []map[string]any:
		return len(v)
	case map[string]any:
		// ลอง field ที่นิยม
		for _, key := range []string{"rows", "data", "items", "results", "records"} {
			if inner, ok := v[key]; ok {
				if arr, ok := inner.([]any); ok {
					return len(arr)
				}
			}
		}
	}
	return 0
}

// extractPreview — สกัด preview N rows แรกจาก result
// คืน { rows: [...], total: N, truncated: true } ถ้ามีหลาย rows
func extractPreview(data any, n int) any {
	if data == nil {
		return nil
	}
	switch v := data.(type) {
	case []any:
		if len(v) <= n {
			return data
		}
		return v[:n]
	case []map[string]any:
		if len(v) <= n {
			return data
		}
		return v[:n]
	case map[string]any:
		// shallow copy + ตัด field array ที่ใหญ่
		out := make(map[string]any, len(v))
		for k, val := range v {
			if arr, ok := val.([]any); ok && len(arr) > n {
				out[k] = arr[:n]
			} else {
				out[k] = val
			}
		}
		return out
	}
	return data
}
