package gentranspdf

import (
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// TestFormatDateWithFormat_BranchTimezoneConversion ตรวจ Timezone Iron Rule:
// DB เก็บ UTC+0 — ค่า 2026-06-22T18:30:00Z คือ 23 มิ.ย. 2026 เวลา 01:30 ตามเวลาไทย (Asia/Bangkok, UTC+7)
// PDF ต้องพิมพ์วันที่ 23 (ไม่ใช่ 22) เพื่อกัน off-by-one ของเอกสารที่ทำรายการช่วง 00:00–06:59 เวลาไทย
func TestFormatDateWithFormat_BranchTimezoneConversion(t *testing.T) {
	utcInstant := time.Date(2026, time.June, 22, 18, 30, 0, 0, time.UTC)

	cases := []struct {
		name     string
		format   string
		timezone string
		want     string
	}{
		{"iso_ad_bangkok", "YYYY-MM-DD", "Asia/Bangkok", "2026-06-23"},
		{"iso_ad_empty_defaults_bangkok", "YYYY-MM-DD", "", "2026-06-23"},
		{"default_ddmmyyyy_bangkok", "DD/MM/YYYY", "Asia/Bangkok", "23/06/2026"},
		{"buddhist_year_bangkok", "DD/MM/BBBB", "Asia/Bangkok", "23/06/2569"}, // 2026 + 543 = 2569
		{"buddhist_full_month_bangkok", "DD MMMM BBBB", "Asia/Bangkok", "23 มิถุนายน 2569"},
		{"english_short_month_bangkok", "DD MMM YYYY", "Asia/Bangkok", "23 Jun 2026"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// time.Time path
			if got := FormatDateWithFormat(utcInstant, tc.format, tc.timezone); got != tc.want {
				t.Fatalf("time.Time: FormatDateWithFormat(%s, %q, %q) = %q, want %q",
					utcInstant.Format(time.RFC3339), tc.format, tc.timezone, got, tc.want)
			}
			// primitive.DateTime path (ค่าจริงจาก doc["docdatetime"] ของ MongoDB)
			pdt := primitive.NewDateTimeFromTime(utcInstant)
			if got := FormatDateWithFormat(pdt, tc.format, tc.timezone); got != tc.want {
				t.Fatalf("primitive.DateTime: FormatDateWithFormat(%s, %q, %q) = %q, want %q",
					utcInstant.Format(time.RFC3339), tc.format, tc.timezone, got, tc.want)
			}
		})
	}
}

// TestFormatDateWithFormat_SameDayUTC ตรวจว่ากรณีที่ยังไม่ข้ามวัน วันที่ยังถูกต้อง (ไม่ over-correct)
func TestFormatDateWithFormat_SameDayUTC(t *testing.T) {
	// 2026-06-22T05:00:00Z = 12:00 เวลาไทย ยังเป็นวันที่ 22
	utcInstant := time.Date(2026, time.June, 22, 5, 0, 0, 0, time.UTC)
	if got := FormatDateWithFormat(utcInstant, "YYYY-MM-DD", "Asia/Bangkok"); got != "2026-06-22" {
		t.Fatalf("FormatDateWithFormat(same-day UTC) = %q, want 2026-06-22", got)
	}
}
