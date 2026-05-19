package dataimport

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/joho/godotenv"
)

// TestXLSXProductCompare - ทดสอบการ compare Excel กับ MongoDB
func TestXLSXProductCompare(t *testing.T) {
	if os.Getenv("BC_RUN_XLSX_PRODUCT_COMPARE") != "1" {
		t.Skip("set BC_RUN_XLSX_PRODUCT_COMPARE=1 to run this integration test with local .env and upload fixture")
	}

	// โหลด .env จาก parent directory (เพราะ working dir คือ dataimport/)
	wd, _ := os.Getwd()
	envPath := filepath.Join(wd, "..", ".env.development")
	if err := godotenv.Load(envPath); err != nil {
		t.Fatalf("Failed to load .env: %v", err)
	}

	fmt.Println("\n=================================")
	fmt.Println("Testing XLSX Product Comparison")
	fmt.Println("=================================")
	fmt.Println()

	// กำหนดค่า
	shopID := "33UYr4vEDECjXsql4x3Lb4FdWfX"
	fileName := "761a0b3a-a2ce-47b8-8711-f3525158b484.xlsx"

	// หา file path
	tempDir := os.TempDir()
	filePath := filepath.Join(tempDir, "uploads", fileName)

	fmt.Printf("Shop ID: %s\n", shopID)
	fmt.Printf("File: %s\n", fileName)
	fmt.Printf("Path: %s\n", filePath)
	fmt.Println()

	// ตรวจสอบว่าไฟล์มีอยู่จริง
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		t.Fatalf("File not found at %s\nPlease upload the file first", filePath)
	}

	// สร้าง session
	session := &ProductPrepareSession{
		ShopID:        shopID,
		FileName:      fileName,
		FilePath:      filePath,
		Status:        StatusNotStarted,
		ProcessedRows: 0,
		SuccessCount:  0,
		ErrorCount:    0,
		StartTime:     time.Now(),
		Result:        []ProductPrepareResult{},
	}

	// เริ่ม process
	fmt.Println("Starting Excel processing...")
	fmt.Println()
	processProductExcel(session)

	// แสดงผลลัพธ์
	fmt.Println("\n=== PROCESSING SUMMARY ===")
	fmt.Printf("Status: %d\n", session.Status)
	fmt.Printf("Total Rows: %d\n", session.TotalRows)
	fmt.Printf("Processed Rows: %d\n", session.ProcessedRows)
	fmt.Printf("Success Count: %d\n", session.SuccessCount)
	fmt.Printf("Error Count: %d\n", session.ErrorCount)
	fmt.Printf("Progress: %.1f%%\n", session.Progress)

	if session.EndTime != nil {
		duration := session.EndTime.Sub(session.StartTime)
		fmt.Printf("Duration: %s\n", duration)
	}

	if session.ErrorMessage != "" {
		fmt.Printf("Error Message: %s\n", session.ErrorMessage)
	}

	// แสดง Duplicate Barcodes JSON
	if len(session.DuplicateBarcodes) > 0 {
		fmt.Printf("\n=== DUPLICATE BARCODES (JSON) ===\n")
		duplicateJSON, _ := json.MarshalIndent(session.DuplicateBarcodes, "", "  ")
		fmt.Printf("%s\n", string(duplicateJSON))
	}

	// แสดงผลการเปรียบเทียบ
	if session.ComparisonResult != nil {
		comp := session.ComparisonResult
		fmt.Println("\n=== COMPARISON RESULTS ===")
		fmt.Printf("Total Excel Rows: %d\n", comp.TotalExcelRows)
		fmt.Printf("Total MongoDB Products: %d\n", comp.TotalMongoProducts)
		fmt.Printf("INSERT (action=1): %d\n", comp.NewProductCount)
		fmt.Printf("UPDATE (action=2): %d\n", comp.UpdatedProductCount)
		fmt.Printf("MATCH (action=0): %d\n", comp.UnchangedCount)
		fmt.Printf("Total Products: %d\n", len(comp.Products))

		// แสดง sample products
		fmt.Println("\n=== SAMPLE PRODUCTS (First 5) ===")
		maxSample := 5
		if len(comp.Products) < maxSample {
			maxSample = len(comp.Products)
		}

		for i := 0; i < maxSample; i++ {
			p := comp.Products[i]
			fmt.Printf("\n[%d] Barcode: %s\n", i+1, p.Barcode)

			if p.Mongo != nil {
				fmt.Printf("  MongoDB:\n")
				fmt.Printf("    Code: %s\n", p.Mongo.Code)
				fmt.Printf("    Name: %s\n", p.Mongo.Name)
				fmt.Printf("    DivideValue: %.2f\n", p.Mongo.DivideValue)
				fmt.Printf("    StandValue: %.2f\n", p.Mongo.StandValue)
			}

			if p.Excel != nil {
				fmt.Printf("  Excel:\n")
				fmt.Printf("    Code: %s\n", p.Excel.Code)
				fmt.Printf("    Name: %s\n", p.Excel.Name)
				fmt.Printf("    DivideValue: %.2f\n", p.Excel.DivideValue)
				fmt.Printf("    StandValue: %.2f\n", p.Excel.StandValue)
			} else {
				fmt.Printf("  Excel: (not found)\n")
			}

			actionText := ""
			switch p.Action {
			case 0:
				actionText = "MATCH (ข้อมูลตรงกัน)"
			case 1:
				actionText = "INSERT (ต้อง insert)"
			case 2:
				actionText = "UPDATE (ข้อมูลต่างกัน)"
			}
			fmt.Printf("  Action: %d - %s\n", p.Action, actionText)
		}

		// นับ action distribution
		actionCounts := make(map[int]int)
		for _, p := range comp.Products {
			actionCounts[p.Action]++
		}

		fmt.Println("\n=== ACTION DISTRIBUTION ===")
		for action, count := range actionCounts {
			actionText := ""
			switch action {
			case 0:
				actionText = "MATCH"
			case 1:
				actionText = "INSERT"
			case 2:
				actionText = "UPDATE"
			}
			fmt.Printf("Action %d (%s): %d products\n", action, actionText, count)
		}

		// บันทึกเป็น JSON file
		outputFile := fmt.Sprintf("test_result_%s.json", time.Now().Format("20060102_150405"))
		jsonData, err := json.MarshalIndent(comp, "", "  ")
		if err != nil {
			fmt.Printf("\nError marshaling JSON: %v\n", err)
		} else {
			err = os.WriteFile(outputFile, jsonData, 0644)
			if err != nil {
				fmt.Printf("\nError writing file: %v\n", err)
			} else {
				fmt.Printf("\nFull comparison result saved to: %s\n", outputFile)
			}
		}
	} else {
		fmt.Println("\nNo comparison results available")
	}

	// ตรวจสอบว่ามีข้อผิดพลาดหรือไม่
	if session.Status == StatusError {
		t.Errorf("Processing failed: %s", session.ErrorMessage)
	}

	fmt.Println()
	fmt.Println("=== TEST COMPLETED ===")
	fmt.Println()
}
