package cheque

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
)

// requestIDPattern: requestid ที่หน้าจอสร้างต่อการกดบันทึกหนึ่งครั้ง (เช่น UUID) — ตัวอักษรที่ปลอดภัยเท่านั้น
var requestIDPattern = regexp.MustCompile(`^[A-Za-z0-9._:-]{16,80}$`)

// claimRequest จองเลข requestid ใน tx ของคำสั่งเขียน: คำขอแรกได้ replayed=false แล้วต้องเรียก completeRequest ก่อน commit;
// คำขอซ้ำ (กดบันทึกซ้ำ/ส่งซ้ำหลัง timeout) รอจนคำขอแรกจบ แล้วได้ผลเดิมใน out โดยไม่เขียนอะไรเพิ่ม
func claimRequest(ctx context.Context, tx *sql.Tx, scope Scope, requestID, operation string, out any) (bool, error) {
	if !requestIDPattern.MatchString(requestID) {
		return false, badRequest("cheque_err_request_id_invalid", "requestid")
	}
	res, err := tx.ExecContext(ctx, `INSERT INTO cheque_requests (company, request_id, operation, created_by) VALUES ($1, $2, $3, $4)
		ON CONFLICT (company, request_id) DO NOTHING`, scope.Company, requestID, operation, scope.Actor)
	if err != nil {
		return false, err
	}
	if n, err := res.RowsAffected(); err != nil || n == 1 {
		return false, err
	}
	var (
		storedOp string
		result   []byte
	)
	if err = tx.QueryRowContext(ctx, `SELECT operation, result FROM cheque_requests WHERE company = $1 AND request_id = $2`,
		scope.Company, requestID).Scan(&storedOp, &result); err != nil {
		return false, err
	}
	if storedOp != operation || len(result) == 0 {
		return false, newError(http.StatusConflict, "cheque_err_request_id_reused", "requestid")
	}
	if err = json.Unmarshal(result, out); err != nil {
		return false, fmt.Errorf("replay request %s: %w", requestID, err)
	}
	return true, nil
}

// completeRequest เก็บผลของคำสั่งไว้ตอบคำขอซ้ำ (อยู่ใน tx เดียวกับการเขียนจริง จึงมีผลก็ต่อเมื่อ commit สำเร็จ)
func completeRequest(ctx context.Context, tx *sql.Tx, scope Scope, requestID string, result any) error {
	payload, err := json.Marshal(result)
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `UPDATE cheque_requests SET result = $3 WHERE company = $1 AND request_id = $2`, scope.Company, requestID, payload)
	return err
}
