package handlers

import (
	"fmt"
	"net"
	"time"

	"smlcloudplatform/internal/goapi/logger"
)

// min function สำหรับ Go version ที่ไม่มี built-in min
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// GetLocalIPAddress - gets the local IP address of the machine
func GetLocalIPAddress() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() {
			if ipnet.IP.To4() != nil {
				return ipnet.IP.String(), nil
			}
		}
	}
	return "", fmt.Errorf("ไม่พบ IP address ในระบบ local")
}

// Helper function to validate holding Code format
func ValidateHoldingCode(holdingCode string) bool {
	if len(holdingCode) < 10 || len(holdingCode) > 50 {
		return false
	}
	// Add more validation logic as needed
	return true
}

// Helper function to validate GUID format
func ValidateGUID(guid string) bool {
	if len(guid) < 10 || len(guid) > 100 {
		return false
	}
	// Add more validation logic as needed
	return true
}

// ParseRebuildDates - parses rebuild dates from string format
func ParseRebuildDates(startDateStr, endDateStr string) (time.Time, time.Time, bool) {
	layout := "2006-01-02"
	var startDate, endDate time.Time
	useAllData := false

	if startDateStr == "" && endDateStr == "" {
		useAllData = true
		return startDate, endDate, useAllData
	}

	now := time.Now().UTC()
	startDate = now.AddDate(-100, 0, 0) // ย้อนหลัง 100 ปี เป็นค่าเริ่มต้น
	endDate = now

	if startDateStr != "" {
		if parsedDate, err := time.Parse(layout, startDateStr); err == nil {
			startDate = parsedDate.UTC()
		} else {
			logger.Info("Invalid start date format. Using default: %s", startDate.Format(layout))
		}
	}

	if endDateStr != "" {
		if parsedDate, err := time.Parse(layout, endDateStr); err == nil {
			endDate = parsedDate.UTC().Add(time.Hour*23 + time.Minute*59 + time.Second*59)
		} else {
			logger.Info("Invalid end date format. Using default: %s", endDate.Format(layout))
		}
	}

	return startDate, endDate, useAllData
}
