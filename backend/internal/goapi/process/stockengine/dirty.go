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

// Execer คือสิ่งที่สั่งคำสั่งฐานข้อมูลได้ ทั้งการเชื่อมต่อปกติและทรานแซกชัน
// มีไว้ให้ส่วนอื่นทำเครื่องหมายงานค้างในทรานแซกชันเดียวกับที่แก้เอกสารได้
type Execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
}

// MarkDirty บันทึกว่าสินค้าตัวนี้ต้องคำนวณต้นทุนใหม่ตั้งแต่วันที่ที่ระบุ
//
// ถ้ามีงานค้างอยู่แล้ว จะยุบรวมเป็นแถวเดียวและเลื่อนจุดเริ่มคิดไปยังวันที่เก่าที่สุดที่ถูกกระทบ
// เอกสารหลายร้อยใบที่แตะสินค้าเดียวกันจึงกลายเป็นงานคำนวณชิ้นเดียว
//
// enqueuedat เก็บเวลาที่เข้าคิวครั้งแรกไว้เสมอ (ไม่เลื่อนตามการแตะครั้งหลัง)
// สินค้าที่ถูกแก้บ่อยจึงไม่ถูกดันไปท้ายคิวไม่รู้จบ ส่วน markedat คือเวลาแตะล่าสุด
// ใช้ตอนจบงานเพื่อดูว่ามีเอกสารเข้ามาใหม่ระหว่างที่กำลังคำนวณหรือไม่
func MarkDirty(ctx context.Context, db Execer, businessCode, itemCode string, from time.Time, reason string) error {
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
		              markedat = now()`,
		businessCode, itemCode, from, reason); err != nil {
		return fmt.Errorf("mark dirty %s/%s: %w", businessCode, itemCode, err)
	}

	return notifyWorkers(ctx, db, businessCode+"|"+itemCode)
}

// MarkDocumentDirty ทำเครื่องหมายให้สินค้าทุกตัวในเอกสารใบนี้ถูกคิดต้นทุนใหม่
// โดยใช้วันที่ของเอกสารที่ยังอยู่ในฐานข้อมูลจริง
//
// ต้องเรียก "ก่อน" ลบหรือแก้เอกสาร มิฉะนั้นจะไม่มีวันที่เดิมให้อ้างอิง
// เอกสารที่ย้ายวันจากมกราคมไปกุมภาพันธ์ ถ้าคิดใหม่แค่เดือนกุมภาพันธ์
// รายการเดิมในเดือนมกราคมจะค้างอยู่ในสมุดสต็อกและทำให้ยอดเบิ้ล
func MarkDocumentDirty(ctx context.Context, exec Execer, businessCode, docNo string, transFlag int, reason string) error {
	if businessCode == "" || docNo == "" {
		return nil
	}
	if reason == "" {
		reason = "docchange"
	}

	if _, err := exec.ExecContext(ctx, `
		INSERT INTO stock_dirty (businesscode, itemcode, fromdate, reason)
		SELECT businesscode, itemcode, MIN(docdatetime), $4
		FROM docdetail
		WHERE businesscode = $1 AND docno = $2 AND transflag = $3 AND itemcode <> ''
		GROUP BY businesscode, itemcode
		ON CONFLICT (businesscode, itemcode)
		DO UPDATE SET fromdate = LEAST(stock_dirty.fromdate, EXCLUDED.fromdate),
		              reason = EXCLUDED.reason,
		              markedat = now()`,
		businessCode, docNo, transFlag, reason); err != nil {
		return fmt.Errorf("mark document dirty %s/%s: %w", businessCode, docNo, err)
	}

	return notifyWorkers(ctx, exec, businessCode+"|"+docNo)
}

// notifyWorkers ปลุก worker ทันที ถ้าไม่มีใครฟังอยู่ก็ไม่เป็นไร งานยังค้างอยู่ในตารางให้รอบถัดไปเก็บ
func notifyWorkers(ctx context.Context, exec Execer, payload string) error {
	if _, err := exec.ExecContext(ctx, `SELECT pg_notify($1, $2)`, notifyChannel, payload); err != nil {
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

// retryBackoff คือเวลาพักก่อนหยิบงานที่เพิ่งล้มเหลวมาลองใหม่ คูณตามจำนวนครั้งที่ลองไปแล้ว
const retryBackoff = 30 * time.Second

// ReleaseDirty ปล่อยงานกลับเข้าคิวพร้อมบันทึกสาเหตุที่ทำไม่สำเร็จ
// ไม่ลดจำนวนครั้งที่พยายาม เพื่อให้งานที่พังซ้ำ ๆ ไปจบที่คิวงานเสียในที่สุด
//
// งานที่ล้มเหลวถูกพักไว้ก่อนตามจำนวนครั้งที่ลอง ไม่ปล่อยให้หยิบกลับมาทันที
// ฐานข้อมูลสะดุดหนึ่งวินาทีจะได้ไม่ทำให้ลองครบห้าครั้งภายในไม่กี่มิลลิวินาทีแล้วตกไปคิวงานเสีย
// งานที่ถูกแย่งไว้โดยตัวอื่น (cause เป็น nil) ปล่อยคืนทันทีไม่ต้องพัก
func ReleaseDirty(ctx context.Context, db Execer, item DirtyItem, cause error) error {
	if cause == nil {
		if _, err := db.ExecContext(ctx, `
			UPDATE stock_dirty
			SET leaseowner = NULL, leaseuntil = NULL
			WHERE businesscode = $1 AND itemcode = $2`,
			item.BusinessCode, item.ItemCode); err != nil {
			return fmt.Errorf("release dirty item: %w", err)
		}
		return nil
	}

	attempts := item.Attempts
	if attempts < 1 {
		attempts = 1
	}
	wait := fmt.Sprintf("%d seconds", int(retryBackoff.Seconds())*attempts)

	if _, err := db.ExecContext(ctx, `
		UPDATE stock_dirty
		SET leaseowner = NULL, leaseuntil = now() + $3::interval, lasterror = $4
		WHERE businesscode = $1 AND itemcode = $2`,
		item.BusinessCode, item.ItemCode, wait, cause.Error()); err != nil {
		return fmt.Errorf("release dirty item: %w", err)
	}
	return nil
}

// MoveToDeadLetter ย้ายงานที่ลองจนครบจำนวนครั้งแล้วยังไม่สำเร็จออกจากคิวหลัก
//
// เก็บในตารางของเครื่องคิดต้นทุนเอง ไม่ปนกับคิวงานเสียของเอกสาร
// เพราะงานที่นี่เป็นระดับสินค้า ไม่มีเลขที่เอกสารและประเภทเอกสารให้บันทึก
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
		INSERT INTO stock_dead_letter (businesscode, itemcode, fromdate, attempts, lasterror)
		VALUES ($1, $2, $3, $4, $5)`,
		item.BusinessCode, item.ItemCode, item.FromDate, item.Attempts, message); err != nil {
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

// LoadQueueStatus อ่านสภาพคิวปัจจุบันของบริษัทหนึ่ง
// ฐานข้อมูลหนึ่งฐานเก็บงานของหลายบริษัท ถ้าไม่กรองผู้ใช้จะเห็นงานของบริษัทอื่นปนมาด้วย
func LoadQueueStatus(ctx context.Context, db *sql.DB, businessCode string) (QueueStatus, error) {
	var status QueueStatus
	var oldest sql.NullTime

	// งานที่เก่าที่สุดต้องนับเฉพาะงานที่ยังรอคิว ไม่รวมงานที่มีคนทำอยู่แล้ว
	// ไม่อย่างนั้นจอจะรายงานว่ามีงานค้างนานทั้งที่กำลังคำนวณอยู่
	err := db.QueryRowContext(ctx, `
		SELECT
			COUNT(*) FILTER (WHERE leaseuntil IS NULL OR leaseuntil < now()),
			COUNT(*) FILTER (WHERE leaseuntil >= now()),
			COUNT(*) FILTER (WHERE attempts > 0 AND lasterror <> ''),
			MIN(enqueuedat) FILTER (WHERE leaseuntil IS NULL OR leaseuntil < now())
		FROM stock_dirty
		WHERE businesscode = $1`, businessCode).Scan(&status.Pending, &status.Processing, &status.Failing, &oldest)
	if err != nil {
		return status, fmt.Errorf("load queue status: %w", err)
	}
	if oldest.Valid {
		status.OldestWait = oldest.Time
	}
	return status, nil
}
