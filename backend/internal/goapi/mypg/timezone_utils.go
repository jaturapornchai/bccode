package mypg

import (
	"fmt"
	"strings"
	"time"
)

// TimezoneConfig เก็บข้อมูลการตั้งค่า timezone
type TimezoneConfig struct {
	TimezoneCode string
	TimezoneName string
	UTCOffset    int // offset ในหน่วยชั่วโมง
	DisplayName  string
}

// GetSupportedTimezones คืนค่ารายการ timezone ที่รองรับ
func GetSupportedTimezones() map[string]TimezoneConfig {
	return map[string]TimezoneConfig{
		"TH":     {TimezoneCode: "TH", TimezoneName: "Asia/Bangkok", UTCOffset: 7, DisplayName: "Thailand (UTC+7)"},
		"US_EST": {TimezoneCode: "US_EST", TimezoneName: "America/New_York", UTCOffset: -5, DisplayName: "USA Eastern (UTC-5)"},
		"US_PST": {TimezoneCode: "US_PST", TimezoneName: "America/Los_Angeles", UTCOffset: -8, DisplayName: "USA Pacific (UTC-8)"},
		"US_CST": {TimezoneCode: "US_CST", TimezoneName: "America/Chicago", UTCOffset: -6, DisplayName: "USA Central (UTC-6)"},
		"UK":     {TimezoneCode: "UK", TimezoneName: "Europe/London", UTCOffset: 0, DisplayName: "United Kingdom (UTC+0)"},
		"JP":     {TimezoneCode: "JP", TimezoneName: "Asia/Tokyo", UTCOffset: 9, DisplayName: "Japan (UTC+9)"},
		"CN":     {TimezoneCode: "CN", TimezoneName: "Asia/Shanghai", UTCOffset: 8, DisplayName: "China (UTC+8)"},
		"IN":     {TimezoneCode: "IN", TimezoneName: "Asia/Kolkata", UTCOffset: 5, DisplayName: "India (UTC+5:30)"},
		"AU_SYD": {TimezoneCode: "AU_SYD", TimezoneName: "Australia/Sydney", UTCOffset: 10, DisplayName: "Australia Sydney (UTC+10)"},
		"AU_MEL": {TimezoneCode: "AU_MEL", TimezoneName: "Australia/Melbourne", UTCOffset: 10, DisplayName: "Australia Melbourne (UTC+10)"},
		"SG":     {TimezoneCode: "SG", TimezoneName: "Asia/Singapore", UTCOffset: 8, DisplayName: "Singapore (UTC+8)"},
		"MY":     {TimezoneCode: "MY", TimezoneName: "Asia/Kuala_Lumpur", UTCOffset: 8, DisplayName: "Malaysia (UTC+8)"},
		"VN":     {TimezoneCode: "VN", TimezoneName: "Asia/Ho_Chi_Minh", UTCOffset: 7, DisplayName: "Vietnam (UTC+7)"},
		"ID_JKT": {TimezoneCode: "ID_JKT", TimezoneName: "Asia/Jakarta", UTCOffset: 7, DisplayName: "Indonesia Jakarta (UTC+7)"},
		"ID_BLI": {TimezoneCode: "ID_BLI", TimezoneName: "Asia/Makassar", UTCOffset: 8, DisplayName: "Indonesia Bali (UTC+8)"},
		"PH":     {TimezoneCode: "PH", TimezoneName: "Asia/Manila", UTCOffset: 8, DisplayName: "Philippines (UTC+8)"},
		"KR":     {TimezoneCode: "KR", TimezoneName: "Asia/Seoul", UTCOffset: 9, DisplayName: "South Korea (UTC+9)"},
		"DE":     {TimezoneCode: "DE", TimezoneName: "Europe/Berlin", UTCOffset: 1, DisplayName: "Germany (UTC+1)"},
		"FR":     {TimezoneCode: "FR", TimezoneName: "Europe/Paris", UTCOffset: 1, DisplayName: "France (UTC+1)"},
		"IT":     {TimezoneCode: "IT", TimezoneName: "Europe/Rome", UTCOffset: 1, DisplayName: "Italy (UTC+1)"},
		"ES":     {TimezoneCode: "ES", TimezoneName: "Europe/Madrid", UTCOffset: 1, DisplayName: "Spain (UTC+1)"},
		"RU_MSK": {TimezoneCode: "RU_MSK", TimezoneName: "Europe/Moscow", UTCOffset: 3, DisplayName: "Russia Moscow (UTC+3)"},
		"BR":     {TimezoneCode: "BR", TimezoneName: "America/Sao_Paulo", UTCOffset: -3, DisplayName: "Brazil (UTC-3)"},
		"MX":     {TimezoneCode: "MX", TimezoneName: "America/Mexico_City", UTCOffset: -6, DisplayName: "Mexico (UTC-6)"},
		"UTC":    {TimezoneCode: "UTC", TimezoneName: "UTC", UTCOffset: 0, DisplayName: "UTC (UTC+0)"},
	}
}

// BuildDateConditionSQLWithTimeZone สร้าง SQL condition สำหรับการ filter ตามวันที่ของ timezone ที่กำหนด
// โดยใช้ AT TIME ZONE ใน PostgreSQL
func BuildDateConditionSQLWithTimeZone(fieldName, dateStr, timezoneCode string) (condition string, args []interface{}, err error) {
	timezones := GetSupportedTimezones()
	config, exists := timezones[strings.ToUpper(timezoneCode)]
	if !exists {
		return "", nil, fmt.Errorf("unsupported timezone code: %s", timezoneCode)
	}

	// Parse วันที่
	_, err = time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", nil, fmt.Errorf("invalid date format: %v", err)
	}

	// สร้าง condition SQL ที่แปลง timestamp ไปเป็นเวลาท้องถิ่นก่อนเปรียบเทียบ
	condition = fmt.Sprintf("(%s AT TIME ZONE '%s')::date <= $1::date", fieldName, config.TimezoneName)
	args = []interface{}{dateStr}

	return condition, args, nil
}

// GetDateRangeForQuery สร้าง query condition สำหรับการค้นหาข้อมูลตามช่วงวันที่ของ timezone ที่กำหนด
func GetDateRangeForQuery(dateStr string, timezoneCode string) (startUTC, endUTC string, err error) {
	timezones := GetSupportedTimezones()
	config, exists := timezones[strings.ToUpper(timezoneCode)]
	if !exists {
		return "", "", fmt.Errorf("unsupported timezone code: %s", timezoneCode)
	}

	// Parse วันที่
	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", "", fmt.Errorf("invalid date format: %v", err)
	}

	// โหลด timezone ของประเทศ
	location, err := time.LoadLocation(config.TimezoneName)
	if err != nil {
		return "", "", fmt.Errorf("failed to load timezone %s: %v", config.TimezoneName, err)
	}

	// เริ่มต้นวัน (00:00:00 เวลาท้องถิ่น)
	startOfDay := time.Date(
		parsedDate.Year(),
		parsedDate.Month(),
		parsedDate.Day(),
		0, 0, 0, 0,
		location,
	)

	// สิ้นสุดวัน (23:59:59.999999999 เวลาท้องถิ่น)
	endOfDay := time.Date(
		parsedDate.Year(),
		parsedDate.Month(),
		parsedDate.Day(),
		23, 59, 59, 999999999,
		location,
	)

	// แปลงเป็น UTC
	utcStart := startOfDay.UTC()
	utcEnd := endOfDay.UTC()

	return utcStart.Format("2006-01-02 15:04:05"), utcEnd.Format("2006-01-02 15:04:05"), nil
}

// BuildDateConditionSQL สร้าง SQL condition สำหรับการ filter ตามวันที่ของ timezone ที่กำหนด
func BuildDateConditionSQL(fieldName, dateStr, timezoneCode string) (condition string, args []interface{}, err error) {
	startUTC, endUTC, err := GetDateRangeForQuery(dateStr, timezoneCode)
	if err != nil {
		return "", nil, err
	}

	// สร้าง condition SQL
	condition = fmt.Sprintf("%s >= $1 AND %s <= $2", fieldName, fieldName)
	args = []interface{}{startUTC, endUTC}

	return condition, args, nil
}

// GetCurrentTimeInTimezone ดึงเวลาปัจจุบันในเวลาของ timezone ที่กำหนด
func GetCurrentTimeInTimezone(timezoneCode string) (time.Time, error) {
	timezones := GetSupportedTimezones()
	config, exists := timezones[strings.ToUpper(timezoneCode)]
	if !exists {
		return time.Time{}, fmt.Errorf("unsupported timezone code: %s", timezoneCode)
	}

	location, err := time.LoadLocation(config.TimezoneName)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to load timezone %s: %v", config.TimezoneName, err)
	}

	return time.Now().In(location), nil
}

// ConvertUTCToLocalTime แปลงเวลา UTC เป็นเวลาท้องถิ่นของ timezone ที่กำหนด
func ConvertUTCToLocalTime(utcTime time.Time, timezoneCode string) (time.Time, error) {
	timezones := GetSupportedTimezones()
	config, exists := timezones[strings.ToUpper(timezoneCode)]
	if !exists {
		return time.Time{}, fmt.Errorf("unsupported timezone code: %s", timezoneCode)
	}

	location, err := time.LoadLocation(config.TimezoneName)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to load timezone %s: %v", config.TimezoneName, err)
	}

	return utcTime.In(location), nil
}

// รองรับเฉพาะประเทศไทย (backward compatibility)
func ConvertDateToThaiTimezone(dateStr string) (string, error) {
	timezones := GetSupportedTimezones()
	config := timezones["TH"]

	parsedDate, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		return "", fmt.Errorf("invalid date format: %v", err)
	}

	location, err := time.LoadLocation(config.TimezoneName)
	if err != nil {
		return "", fmt.Errorf("failed to load timezone %s: %v", config.TimezoneName, err)
	}

	endOfDay := time.Date(
		parsedDate.Year(),
		parsedDate.Month(),
		parsedDate.Day(),
		23, 59, 59, 999999999,
		location,
	)

	utcEndOfDay := endOfDay.UTC()
	return utcEndOfDay.Format("2006-01-02 15:04:05"), nil
}

func GetThaiDateRangeForQuery(dateStr string) (startUTC, endUTC string, err error) {
	return GetDateRangeForQuery(dateStr, "TH")
}
