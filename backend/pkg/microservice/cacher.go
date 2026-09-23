package microservice

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/lib/pq"
)

// ICacher - ที่เก็บ session/ข้อมูลชั่วคราวแบบ key/hash พร้อมวันหมดอายุ
// เก็บใน PostgreSQL ตาราง cache_entries (แทน Redis ที่ถอดออกแล้ว 2026-09-23)
// ความหมายของแต่ละคำสั่งคงเดิมตาม Redis เพื่อให้ auth/session ทำงานเหมือนเดิม
type ICacher interface {
	HMSet(key string, fieldValues map[string]interface{}) error
	HGet(key string, field string) (string, error)
	HGetAll(key string) (map[string]string, error)
	HMGet(key string, fields []string) ([]interface{}, error)
	// ConsumeHash อ่าน field ที่ต้องการ บันทึก marker แล้วลบ hash ทิ้งในคำสั่งเดียว (ใช้กับ refresh token ใช้ครั้งเดียว)
	// ถ้า hash ถูกใช้ไปแล้ว คืน marker ที่ผู้ใช้คนแรกบันทึกไว้ เพื่อให้ผู้เรียกเพิกถอนทั้งตระกูล token
	ConsumeHash(key string, markerKey string, fields []string, markerTTL time.Duration) ([]interface{}, string, bool, error)

	Set(key string, value interface{}, expire time.Duration) error
	SetS(key string, value string, expire time.Duration) error
	SetNoExpire(key string, value interface{}) error
	Incr(key string) (int, error)
	Get(key string) (string, error)
	// Expire ค่าติดลบหรือศูนย์ = ลบ key ทันที (เหมือน Redis)
	Expire(key string, expire time.Duration) error
	Expires(keys []string, expire time.Duration) error
	Del(keys ...string) error
	Exists(key string) (bool, error)
	// Keys รองรับ glob แบบ * เท่านั้น
	Keys(pattern string) ([]string, error)

	Close() error
	Healthcheck() error
}

// field ว่าง = ค่าแบบ string ธรรมดา (ไม่ใช่ hash)
const cacheSchemaSQL = `
CREATE TABLE IF NOT EXISTS cache_entries (
	cache_key  text        NOT NULL,
	field      text        NOT NULL DEFAULT '',
	value      text        NOT NULL,
	expires_at timestamptz,
	PRIMARY KEY (cache_key, field)
);
CREATE INDEX IF NOT EXISTS idx_cache_entries_expires_at ON cache_entries (expires_at) WHERE expires_at IS NOT NULL;`

const cacheLiveSQL = `(expires_at IS NULL OR expires_at > now())`

type Cacher struct {
	db *sql.DB
}

// NewCacher - สร้างตาราง cache_entries (ถ้ายังไม่มี) ในฐานข้อมูลควบคุมกลาง แล้วคืนตัวเก็บ session
func NewCacher(db *sql.DB) (*Cacher, error) {
	if db == nil {
		return nil, fmt.Errorf("cacher database is required")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if _, err := db.ExecContext(ctx, cacheSchemaSQL); err != nil {
		return nil, fmt.Errorf("create cache_entries: %w", err)
	}
	return &Cacher{db: db}, nil
}

func cacheCtx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), 5*time.Second)
}

// cacheValueText - แปลงค่าเป็น string ในรูปแบบเดียวกับที่ระบบ session เดิมเขียนลง hash (bool → 1/0, ตัวเลข → ข้อความ)
func cacheValueText(value interface{}) (string, error) {
	switch v := value.(type) {
	case nil:
		return "", nil
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case bool:
		if v {
			return "1", nil
		}
		return "0", nil
	case int:
		return strconv.Itoa(v), nil
	case int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return fmt.Sprintf("%d", v), nil
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32), nil
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64), nil
	case time.Time:
		return v.Format(time.RFC3339Nano), nil
	case fmt.Stringer:
		return v.String(), nil
	default:
		return "", fmt.Errorf("cacher: unsupported hash value type %T", value)
	}
}

func expiresAt(expire time.Duration) interface{} {
	if expire <= 0 {
		return nil
	}
	return time.Now().Add(expire)
}

// HMSet - เขียน field ของ hash โดยคงวันหมดอายุเดิมของ key (เหมือน Redis HSET)
func (c *Cacher) HMSet(key string, fieldValues map[string]interface{}) error {
	if len(fieldValues) == 0 {
		return nil
	}
	ctx, cancel := cacheCtx()
	defer cancel()
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var expiry sql.NullTime
	err = tx.QueryRowContext(ctx, `SELECT expires_at FROM cache_entries WHERE cache_key = $1 AND field <> '' AND `+cacheLiveSQL+` LIMIT 1`, key).Scan(&expiry)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	// hash ที่หมดอายุแล้วถือว่าไม่มี — ล้างทิ้งก่อนเขียนใหม่
	if _, err := tx.ExecContext(ctx, `DELETE FROM cache_entries WHERE cache_key = $1 AND NOT `+cacheLiveSQL, key); err != nil {
		return err
	}
	var expiryArg interface{}
	if expiry.Valid {
		expiryArg = expiry.Time
	}
	for field, value := range fieldValues {
		text, err := cacheValueText(value)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `INSERT INTO cache_entries (cache_key, field, value, expires_at) VALUES ($1, $2, $3, $4)
ON CONFLICT (cache_key, field) DO UPDATE SET value = EXCLUDED.value`, key, field, text, expiryArg); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (c *Cacher) HGet(key string, field string) (string, error) {
	ctx, cancel := cacheCtx()
	defer cancel()
	var value string
	err := c.db.QueryRowContext(ctx, `SELECT value FROM cache_entries WHERE cache_key = $1 AND field = $2 AND field <> '' AND `+cacheLiveSQL, key, field).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (c *Cacher) HGetAll(key string) (map[string]string, error) {
	ctx, cancel := cacheCtx()
	defer cancel()
	rows, err := c.db.QueryContext(ctx, `SELECT field, value FROM cache_entries WHERE cache_key = $1 AND field <> '' AND `+cacheLiveSQL, key)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := map[string]string{}
	for rows.Next() {
		var field, value string
		if err := rows.Scan(&field, &value); err != nil {
			return nil, err
		}
		result[field] = value
	}
	return result, rows.Err()
}

// HMGet - คืนค่าตามลำดับ fields; field ที่ไม่มีเป็น nil (เหมือน Redis)
func (c *Cacher) HMGet(key string, fields []string) ([]interface{}, error) {
	all, err := c.HGetAll(key)
	if err != nil {
		return nil, err
	}
	return pickFields(all, fields), nil
}

func pickFields(all map[string]string, fields []string) []interface{} {
	values := make([]interface{}, len(fields))
	for i, field := range fields {
		if value, ok := all[field]; ok {
			values[i] = value
		}
	}
	return values
}

func (c *Cacher) ConsumeHash(key string, markerKey string, fields []string, markerTTL time.Duration) ([]interface{}, string, bool, error) {
	if len(fields) == 0 {
		return nil, "", false, fmt.Errorf("consume hash fields are required")
	}
	if markerTTL < time.Millisecond {
		markerTTL = time.Millisecond
	}
	ctx, cancel := cacheCtx()
	defer cancel()
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, "", false, err
	}
	defer tx.Rollback()

	// ลบแบบ RETURNING ในคำสั่งเดียว — ถ้าสอง request ใช้ token เดียวกันพร้อมกัน จะมีแค่ตัวแรกที่ได้แถวคืน
	rows, err := tx.QueryContext(ctx, `DELETE FROM cache_entries WHERE cache_key = $1 AND field <> '' AND `+cacheLiveSQL+` RETURNING field, value`, key)
	if err != nil {
		return nil, "", false, err
	}
	all := map[string]string{}
	for rows.Next() {
		var field, value string
		if err := rows.Scan(&field, &value); err != nil {
			rows.Close()
			return nil, "", false, err
		}
		all[field] = value
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return nil, "", false, err
	}

	marker, found := all[fields[0]]
	if !found {
		var existing string
		err := tx.QueryRowContext(ctx, `SELECT value FROM cache_entries WHERE cache_key = $1 AND field = '' AND `+cacheLiveSQL, markerKey).Scan(&existing)
		if err != nil && err != sql.ErrNoRows {
			return nil, "", false, err
		}
		return nil, existing, false, tx.Commit()
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO cache_entries (cache_key, field, value, expires_at) VALUES ($1, '', $2, $3)
ON CONFLICT (cache_key, field) DO UPDATE SET value = EXCLUDED.value, expires_at = EXCLUDED.expires_at`, markerKey, marker, time.Now().Add(markerTTL)); err != nil {
		return nil, "", false, err
	}
	if err := tx.Commit(); err != nil {
		return nil, "", false, err
	}
	return pickFields(all, fields), marker, true, nil
}

func (c *Cacher) setText(key string, value string, expire time.Duration) error {
	ctx, cancel := cacheCtx()
	defer cancel()
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	// SET แทนที่ key เดิมทั้งหมด (รวม hash field) เหมือน Redis
	if _, err := tx.ExecContext(ctx, `DELETE FROM cache_entries WHERE cache_key = $1`, key); err != nil {
		return err
	}
	if _, err := tx.ExecContext(ctx, `INSERT INTO cache_entries (cache_key, field, value, expires_at) VALUES ($1, '', $2, $3)`, key, value, expiresAt(expire)); err != nil {
		return err
	}
	return tx.Commit()
}

// Set - เก็บค่าเป็น JSON
func (c *Cacher) Set(key string, value interface{}, expire time.Duration) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return c.setText(key, string(raw), expire)
}

func (c *Cacher) SetS(key string, value string, expire time.Duration) error {
	return c.setText(key, value, expire)
}

func (c *Cacher) SetNoExpire(key string, value interface{}) error {
	return c.Set(key, value, 0)
}

func (c *Cacher) Incr(key string) (int, error) {
	ctx, cancel := cacheCtx()
	defer cancel()
	tx, err := c.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	if _, err := tx.ExecContext(ctx, `DELETE FROM cache_entries WHERE cache_key = $1 AND field = '' AND NOT `+cacheLiveSQL, key); err != nil {
		return 0, err
	}
	var text string
	err = tx.QueryRowContext(ctx, `INSERT INTO cache_entries (cache_key, field, value) VALUES ($1, '', '1')
ON CONFLICT (cache_key, field) DO UPDATE SET value = (cache_entries.value::bigint + 1)::text
RETURNING value`, key).Scan(&text)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return strconv.Atoi(text)
}

// Get - คืน "" เมื่อไม่มี key
func (c *Cacher) Get(key string) (string, error) {
	ctx, cancel := cacheCtx()
	defer cancel()
	var value string
	err := c.db.QueryRowContext(ctx, `SELECT value FROM cache_entries WHERE cache_key = $1 AND field = '' AND `+cacheLiveSQL, key).Scan(&value)
	if err == sql.ErrNoRows {
		return "", nil
	}
	return value, err
}

func (c *Cacher) Expire(key string, expire time.Duration) error {
	return c.Expires([]string{key}, expire)
}

func (c *Cacher) Expires(keys []string, expire time.Duration) error {
	if len(keys) == 0 {
		return nil
	}
	if expire <= 0 {
		return c.Del(keys...)
	}
	ctx, cancel := cacheCtx()
	defer cancel()
	_, err := c.db.ExecContext(ctx, `UPDATE cache_entries SET expires_at = $2 WHERE cache_key = ANY($1) AND `+cacheLiveSQL, pq.Array(keys), time.Now().Add(expire))
	return err
}

func (c *Cacher) Del(keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	ctx, cancel := cacheCtx()
	defer cancel()
	_, err := c.db.ExecContext(ctx, `DELETE FROM cache_entries WHERE cache_key = ANY($1)`, pq.Array(keys))
	return err
}

func (c *Cacher) Exists(key string) (bool, error) {
	ctx, cancel := cacheCtx()
	defer cancel()
	var exists bool
	err := c.db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM cache_entries WHERE cache_key = $1 AND `+cacheLiveSQL+`)`, key).Scan(&exists)
	return exists, err
}

func (c *Cacher) Keys(pattern string) ([]string, error) {
	escaped := strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(pattern)
	like := strings.ReplaceAll(escaped, "*", "%")
	ctx, cancel := cacheCtx()
	defer cancel()
	rows, err := c.db.QueryContext(ctx, `SELECT DISTINCT cache_key FROM cache_entries WHERE cache_key LIKE $1 AND `+cacheLiveSQL, like)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	keys := []string{}
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	return keys, rows.Err()
}

// PurgeExpired - ลบแถวที่หมดอายุ (เรียกเป็นระยะจาก background worker)
func (c *Cacher) PurgeExpired(ctx context.Context) (int64, error) {
	result, err := c.db.ExecContext(ctx, `DELETE FROM cache_entries WHERE expires_at IS NOT NULL AND expires_at <= now()`)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// Close - ฐานข้อมูลเป็นของผู้สร้าง (pool กลาง) จึงไม่ปิดที่นี่
func (c *Cacher) Close() error {
	return nil
}

func (c *Cacher) Healthcheck() error {
	ctx, cancel := cacheCtx()
	defer cancel()
	return c.db.PingContext(ctx)
}
