package stockengine

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

// DirtyItem คืองานคำนวณหนึ่งชิ้นที่ถูกจองไว้แล้ว
type DirtyItem struct {
	BusinessCode string
	ItemCode     string
	FromDate     time.Time
	Reason       string
	Attempts     int
}

// MaxAttempts จำนวนครั้งที่ยอมให้ลองใหม่ก่อนย้ายเข้าคิวงานเสีย
const MaxAttempts = 5

// leaseDuration อายุการจองงาน ถ้าตัวที่จองตายไปเฉย ๆ งานจะกลับมาให้ตัวอื่นทำเมื่อครบเวลานี้
const leaseDuration = 5 * time.Minute

// notifyChannel ชื่อช่องสัญญาณที่ worker เฝ้าฟัง แทนการวนถามฐานข้อมูลทุกวินาที
const notifyChannel = "stock_dirty"

// MarkDirty บันทึกว่าสินค้าตัวนี้ต้องคำนวณต้นทุนใหม่ตั้งแต่วันที่ที่ระบุ
//
// ถ้ามีงานค้างอยู่แล้ว จะยุบรวมเป็นแถวเดียวและเลื่อนจุดเริ่มคิดไปยังวันที่เก่าที่สุดที่ถูกกระทบ
// เอกสารหลายร้อยใบที่แตะสินค้าเดียวกันจึงกลายเป็นงานคำนวณชิ้นเดียว
func MarkDirty(ctx context.Context, db *sql.DB, businessCode, itemCode string, from time.Time, reason string) error {
	if itemCode == "" {
		return fmt.Errorf("stockengine: item code is required")
	}
	if reason == "" {
		reason = "doc"
	}

	if _, err := db.ExecContext(ctx, `
		INSERT INTO stock_dirty (businesscode, itemcode, fromdate, reason)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (businesscode, itemcode)
		DO UPDATE SET fromdate = LEAST(stock_dirty.fromdate, EXCLUDED.fromdate),
		              reason = EXCLUDED.reason,
		              enqueuedat = now()`,
		businessCode, itemCode, from, reason); err != nil {
		return fmt.Errorf("mark dirty %s/%s: %w", businessCode, itemCode, err)
	}

	// ปลุก worker ทันที ถ้าไม่มีใครฟังอยู่ก็ไม่เป็นไร งานยังค้างอยู่ในตารางให้รอบถัดไปเก็บ
	if _, err := db.ExecContext(ctx, `SELECT pg_notify($1, $2)`, notifyChannel, businessCode+"|"+itemCode); err != nil {
		return fmt.Errorf("notify %s: %w", notifyChannel, err)
	}
	return nil
}

// ClaimDirty จองงานที่ยังไม่มีใครทำ สูงสุด limit ชิ้น
// งานที่ถูกจองค้างไว้เกินอายุการจองจะถูกหยิบกลับมาใหม่อัตโนมัติ
func ClaimDirty(ctx context.Context, db *sql.DB, owner string, limit int) ([]DirtyItem, error) {
	if limit <= 0 {
		limit = 1
	}

	rows, err := db.QueryContext(ctx, `
		UPDATE stock_dirty
		SET leaseowner = $1,
		    leaseuntil = now() + $2::interval,
		    attempts = attempts + 1
		WHERE (businesscode, itemcode) IN (
			SELECT businesscode, itemcode
			FROM stock_dirty
			WHERE (leaseuntil IS NULL OR leaseuntil < now())
			  AND attempts < $3
			ORDER BY enqueuedat
			LIMIT $4
			FOR UPDATE SKIP LOCKED
		)
		RETURNING businesscode, itemcode, fromdate, reason, attempts`,
		owner, fmt.Sprintf("%d seconds", int(leaseDuration.Seconds())), MaxAttempts, limit)
	if err != nil {
		return nil, fmt.Errorf("claim dirty items: %w", err)
	}
	defer rows.Close()

	items := make([]DirtyItem, 0, limit)
	for rows.Next() {
		var item DirtyItem
		if err := rows.Scan(&item.BusinessCode, &item.ItemCode, &item.FromDate, &item.Reason, &item.Attempts); err != nil {
			return nil, fmt.Errorf("scan dirty item: %w", err)
		}
		items = append(items, item)
	}
	return items, rows.Err()
}

// ReleaseDirty ปล่อยงานกลับเข้าคิวพร้อมบันทึกสาเหตุที่ทำไม่สำเร็จ
// ไม่ลดจำนวนครั้งที่พยายาม เพื่อให้งานที่พังซ้ำ ๆ ไปจบที่คิวงานเสียในที่สุด
func ReleaseDirty(ctx context.Context, db *sql.DB, item DirtyItem, cause error) error {
	message := ""
	if cause != nil {
		message = cause.Error()
	}
	if _, err := db.ExecContext(ctx, `
		UPDATE stock_dirty
		SET leaseowner = NULL, leaseuntil = NULL, lasterror = $3
		WHERE businesscode = $1 AND itemcode = $2`,
		item.BusinessCode, item.ItemCode, message); err != nil {
		return fmt.Errorf("release dirty item: %w", err)
	}
	return nil
}

// MoveToDeadLetter ย้ายงานที่ลองจนครบจำนวนครั้งแล้วยังไม่สำเร็จออกจากคิวหลัก
func MoveToDeadLetter(ctx context.Context, db *sql.DB, item DirtyItem, cause error) error {
	message := "unknown error"
	if cause != nil {
		message = cause.Error()
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin dead letter transaction: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO deadletterqueue (holdingcode, docno, transflag, retrycount, createdat, errormessage)
		VALUES ($1, $2, '0', $3, now(), $4)`,
		item.BusinessCode, item.ItemCode, item.Attempts, message); err != nil {
		return fmt.Errorf("insert dead letter: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `
		DELETE FROM stock_dirty WHERE businesscode = $1 AND itemcode = $2`,
		item.BusinessCode, item.ItemCode); err != nil {
		return fmt.Errorf("remove dead letter source: %w", err)
	}
	return tx.Commit()
}

// QueueStatus สรุปสภาพคิวสำหรับแสดงให้ผู้ใช้เห็นว่ามีงานค้างจริงเท่าไร
type QueueStatus struct {
	Pending    int       `json:"pending"`
	Processing int       `json:"processing"`
	Failing    int       `json:"failing"`
	OldestWait time.Time `json:"oldestwait,omitempty"`
}

// LoadQueueStatus อ่านสภาพคิวปัจจุบัน
func LoadQueueStatus(ctx context.Context, db *sql.DB) (QueueStatus, error) {
	var status QueueStatus
	var oldest sql.NullTime

	err := db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE leaseuntil IS NULL OR leaseuntil < now()),
			COUNT(*) FILTER (WHERE leaseuntil >= now()),
			COUNT(*) FILTER (WHERE attempts > 0 AND lasterror <> ''),
			MIN(enqueuedat)
		FROM stock_dirty`).Scan(&status.Pending, &status.Processing, &status.Failing, &oldest)
	if err != nil {
		return status, fmt.Errorf("load queue status: %w", err)
	}
	if oldest.Valid {
		status.OldestWait = oldest.Time
	}
	return status, nil
}
