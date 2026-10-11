-- ผลตรวจของห้องบัญชี แยกจาก journal_entries.status ที่ใช้ควบคุมการผ่านรายการ
-- รันหลัง gltrans.sql; ไม่ใช่ลำดับอนุมัติหลายแผนก ผู้ทำบัญชีคนเดียวตรวจได้ตามสิทธิ์
CREATE TABLE IF NOT EXISTS journal_review_events (
    company_code VARCHAR(20) NOT NULL,
    doc_no VARCHAR(30) NOT NULL,
    event_no BIGINT NOT NULL CHECK (event_no > 0), -- service ออกลำดับขณะล็อกหัวใบสำคัญ
    journal_revision INT NOT NULL CHECK (journal_revision > 0),
    review_status SMALLINT NOT NULL CHECK (review_status IN (1,2,3)),
    -- 1=รอตรวจ/เปิดตรวจใหม่, 2=พบข้อแตกต่าง, 3=ตรวจแล้ว; ไม่มี event ของ revision ปัจจุบัน=รอตรวจ
    note TEXT,
    reviewed_by VARCHAR(50) NOT NULL CHECK (btrim(reviewed_by) <> ''),
    reviewed_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    PRIMARY KEY(company_code,doc_no,event_no),
    FOREIGN KEY(company_code,doc_no) REFERENCES journal_entries(company_code,doc_no) ON DELETE RESTRICT,
    CHECK (review_status <> 2 OR (note IS NOT NULL AND btrim(note) <> ''))
);
CREATE INDEX IF NOT EXISTS idx_journal_review_revision
    ON journal_review_events(company_code,doc_no,journal_revision,event_no DESC);
-- service ต้องตรวจ revision ปัจจุบันและ append-only; DDL นี้ยังไม่ห้าม UPDATE/DELETE โดยตรง
