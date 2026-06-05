package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/mypg"
	reportstock "smlcloudplatform/internal/goapi/process/report-stock"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/labstack/echo/v4"
)

// Report Get Handler - handles report generation and file serving
func ReportGetHandler(c echo.Context) error {
	commandId := c.QueryParam("command_id")
	holdingCode := c.QueryParam("holdingcode")
	guid := c.QueryParam("guid")

	// ดึงข้อมูลจาก PostgreSQL เพื่อแสดงผล
	db, err := mypg.PgSqlFastConnect(holdingCode)
	if err != nil {
		logger.Info("Database connection error: %v", err)
		return c.JSON(http.StatusInternalServerError, map[string]string{
			"error": "Database connection failed",
			"code":  "DB_CONNECTION_ERROR",
		})
	}

	if commandId == "data_bin" {
		filePath := c.QueryParam("path")

		var fileContent []byte
		var fileName string

		// ตรวจสอบว่าเป็น S3 object key หรือ local path
		if strings.HasPrefix(filePath, "reports/") {
			// ─── S3 path: download จาก Cloudflare R2 ───
			logger.Info("Downloading report from S3: %s", filePath)
			client, err := GetR2Client()
			if err != nil {
				logger.Error("S3 client not available: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Storage service not available",
					"code":  "S3_CLIENT_ERROR",
				})
			}

			ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
			defer cancel()

			output, err := client.GetObject(ctx, &s3.GetObjectInput{
				Bucket: aws.String(r2BucketName),
				Key:    aws.String(filePath),
			})
			if err != nil {
				logger.Error("Failed to get report from S3: %v", err)
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "Report file not found",
					"code":  "FILE_NOT_FOUND",
				})
			}
			defer output.Body.Close()

			fileContent, err = io.ReadAll(output.Body)
			if err != nil {
				logger.Error("Failed to read S3 object: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Error reading report file",
					"code":  "FILE_READ_ERROR",
				})
			}

			// ใช้ชื่อไฟล์จาก S3 key
			parts := strings.Split(filePath, "/")
			fileName = parts[len(parts)-1]
		} else {
			// ─── Local path (backward compatible) ───
			fileInfo, err := os.Stat(filePath)
			if err != nil {
				logger.Info("File not found: %s", filePath)
				return c.JSON(http.StatusNotFound, map[string]string{
					"error": "File not found",
					"code":  "FILE_NOT_FOUND",
				})
			}

			const maxFileSize = 50 * 1024 * 1024
			if fileInfo.Size() > maxFileSize {
				logger.Info("File too large: %d bytes", fileInfo.Size())
				return c.JSON(http.StatusBadRequest, map[string]string{
					"error": "File too large",
					"code":  "FILE_TOO_LARGE",
				})
			}

			file, err := os.Open(filePath)
			if err != nil {
				logger.Error("opening file: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Error opening report file",
					"code":  "FILE_OPEN_ERROR",
				})
			}
			defer file.Close()

			fileContent, err = io.ReadAll(io.LimitReader(file, maxFileSize))
			if err != nil {
				logger.Error("reading file: %v", err)
				return c.JSON(http.StatusInternalServerError, map[string]string{
					"error": "Error reading report file",
					"code":  "FILE_READ_ERROR",
				})
			}
			fileName = filePath
		}

		// Set the response header to indicate a file download
		c.Response().Header().Set(echo.HeaderContentDisposition, "attachment; filename="+fileName)
		c.Response().Header().Set(echo.HeaderContentType, "application/pdf")
		c.Response().Header().Set(echo.HeaderContentLength, fmt.Sprintf("%d", len(fileContent)))
		c.Response().WriteHeader(http.StatusOK)
		_, err = c.Response().Write(fileContent)
		if err != nil {
			logger.Error("writing file content to response: %v", err)
			return c.String(http.StatusInternalServerError, "Error writing report file")
		}
		return nil
	}

	if commandId == "data_json" {
		// offset string to int
		offset, err := strconv.Atoi(c.QueryParam("offset"))
		if err != nil {
			offset = 0
		}
		// limit string to int
		limit, err := strconv.Atoi(c.QueryParam("limit"))
		if err != nil {
			limit = 100
		}

		// Validate input parameters to prevent injection
		if limit < 1 || limit > 10000 {
			limit = 100
		}
		if offset < 0 {
			offset = 0
		}

		query := "SELECT datajson FROM result WHERE guid = $1 ORDER BY linenumber LIMIT $2 OFFSET $3"
		rows, err := db.Query(query, guid, limit, offset)
		if err != nil {
			logger.Error("Database query error: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Database query failed",
				"code":  "DB_QUERY_ERROR",
			})
		}
		defer rows.Close()

		var datajsonList []map[string]any // เปลี่ยนจาก []string เป็น []map[string]any
		for rows.Next() {
			var datajson []byte // JSONB returns as []byte
			err := rows.Scan(&datajson)
			if err != nil {
				logger.Error("Row scan error: %v", err)
				continue
			}

			// ตรวจสอบว่าข้อมูลไม่ว่าง
			if len(datajson) == 0 {
				logger.Error("Empty JSON data for guid: %s", guid)
				continue
			}

			// Unmarshal JSON เป็น map
			var jsonObj map[string]any
			if err := json.Unmarshal(datajson, &jsonObj); err != nil {
				logger.Error("Failed to unmarshal JSON: %v, data: %s", err, string(datajson))
				continue
			}

			datajsonList = append(datajsonList, jsonObj)
		}

		// ตรวจสอบ rows.Err() หลังจาก loop
		if err = rows.Err(); err != nil {
			logger.Error("Error iterating rows: %v", err)
			return c.JSON(http.StatusInternalServerError, map[string]string{
				"error": "Error reading data",
				"code":  "DB_READ_ERROR",
			})
		}

		// ถ้าไม่มีข้อมูลเลย
		if len(datajsonList) == 0 {
			logger.Info("No data found for guid: %s", guid)
			return c.JSON(http.StatusNotFound, map[string]any{
				"message": "No data found",
				"status":  "error",
				"code":    404,
				"guid":    guid,
			})
		}

		jsonData := map[string]any{
			"message": commandId,
			"status":  "success",
			"code":    200,
			"guid":    guid,
			"count":   len(datajsonList),
			"data":    datajsonList, // ตอนนี้เป็น array of JSON objects แล้ว
		}

		return c.JSON(http.StatusOK, jsonData)
	}

	// เพิ่มคำสั่ง return สำหรับกรณีที่ไม่เข้าเงื่อนไขใดๆ
	return c.JSON(http.StatusOK, map[string]any{
		"message": "Invalid command_id",
		"status":  "error",
		"code":    400,
	})
}

// Report Stock Balance by Warehouse and Barcode
func ReportProductBalanceByWareHouseBarcode(holdingCode string, guid string, finalDate string, timezoneCode string, languageCode string) string {
	return reportstock.ReportProductBalanceByWareHouseAndItem(holdingCode, guid, finalDate, timezoneCode, languageCode)
}

// Report Stock Balance by Location and Barcode
func ReportProductBalanceByLocationBarcode(holdingCode string, guid string, finalDate string, timezoneCode string, languageCode string) string {
	return reportstock.ReportProductBalanceByLocationAndItem(holdingCode, guid, finalDate, timezoneCode, languageCode)
}

// Report Stock Balance by Barcode, Warehouse Code and Location Code
func ReportProductBalanceByBarcodeWhCodeLocationCode(holdingCode string, guid string, condition int, finalDate string, timezoneCode string, languageCode string) string {
	return reportstock.ReportProductBalanceByItemAndWareHouseAndLocation(holdingCode, guid, condition, finalDate, timezoneCode, languageCode)
}

// Report Product Stock Movement
func ReportProductStockMovement(holdingCode string, guid string, timezoneCode string, languageCode string) {
	reportstock.ReportProductStockMovement(holdingCode, guid, timezoneCode, languageCode)
}
