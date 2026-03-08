package coupon

import (
	"bytes"
	"fmt"
	"smlcloudplatform/internal/coupon/models"
	common "smlcloudplatform/internal/models"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/xuri/excelize/v2"
)

// Global batch storage - ใช้ store เดียว
var (
	batchStatusStore = make(map[string]*models.CouponBatchImportStatus)
	batchMutex       = sync.RWMutex{}
)

// Helper functions สำหรับ batch management
func (h CouponHttp) generateBatchID() string {
	return fmt.Sprintf("BATCH_%d_%08x", time.Now().Unix(), time.Now().UnixNano()%100000000)
}

func (h CouponHttp) setBatchStatus(batchID string, status *models.CouponBatchImportStatus) {
	batchMutex.Lock()
	defer batchMutex.Unlock()
	batchStatusStore[batchID] = status
}

func (h CouponHttp) getBatchStatus(batchID string) *models.CouponBatchImportStatus {
	batchMutex.RLock()
	defer batchMutex.RUnlock()
	return batchStatusStore[batchID]
}

// อัพเดทความคืบหน้า
func (h CouponHttp) updateBatchProgress(batchID string, progress, totalRows, processedRows int, message string) {
	batchMutex.Lock()
	defer batchMutex.Unlock()

	if status, exists := batchStatusStore[batchID]; exists {
		status.Progress = progress
		status.TotalRows = totalRows
		status.ProcessedRows = processedRows
		status.SuccessCount = processedRows
		status.Message = message
		status.UpdatedAt = time.Now()
	}
}

// Helper method สำหรับประมวลผล Excel Import
func (h CouponHttp) processExcelImportSync(shopID, authUsername string, req models.CouponExcelImportRequest) (*models.CouponExcelImportResponse, error) {
	// สร้าง batch ID
	batchID := h.generateBatchID()

	// เริ่มต้น batch status
	h.setBatchStatus(batchID, &models.CouponBatchImportStatus{
		BatchID:       batchID,
		Status:        "processing",
		Progress:      0,
		TotalRows:     0,
		ProcessedRows: 0,
		SuccessCount:  0,
		ErrorCount:    0,
		Message:       "เริ่มต้นประมวลผล",
		UpdatedAt:     time.Now(),
	})

	// Parse Excel file
	f, err := excelize.OpenReader(bytes.NewReader(req.FileData))
	if err != nil {
		h.setBatchStatus(batchID, &models.CouponBatchImportStatus{
			BatchID:   batchID,
			Status:    "failed",
			Progress:  0,
			Message:   "ไม่สามารถเปิดไฟล์ Excel ได้: " + err.Error(),
			UpdatedAt: time.Now(),
		})
		return nil, fmt.Errorf("ไม่สามารถเปิดไฟล์ Excel ได้: %v", err)
	}
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Printf("ปิดไฟล์ไม่สำเร็จ: %v\n", err)
		}
	}()

	// กำหนด sheet name
	sheetName := req.SheetName
	if sheetName == "" {
		sheetName = f.GetSheetName(0) // ใช้ sheet แรก
	}

	// อ่านข้อมูลจาก Excel
	rows, err := f.GetRows(sheetName)
	if err != nil {
		h.setBatchStatus(batchID, &models.CouponBatchImportStatus{
			BatchID:  batchID,
			Status:   "failed",
			Progress: 0,
			Message:  "ไม่สามารถอ่านข้อมูลจาก sheet ได้: " + err.Error(),
		})
		return nil, fmt.Errorf("ไม่สามารถอ่านข้อมูลจาก sheet '%s' ได้: %v", sheetName, err)
	}

	totalDataRows := len(rows) - req.StartRow + 1
	if totalDataRows <= 0 {
		totalDataRows = 0
	}

	// อัพเดท total rows
	h.updateBatchProgress(batchID, 0, totalDataRows, 0, "เริ่มต้นประมวลผลข้อมูล")

	if len(rows) < req.StartRow {
		response := &models.CouponExcelImportResponse{
			BatchID:    batchID,
			Success:    false,
			TotalRows:  len(rows),
			ErrorCount: 1,
			Message:    "ไฟล์ไม่มีข้อมูลเพียงพอ",
			Errors: []models.CouponImportError{{
				Row:   0,
				Error: "ไฟล์ต้องมีข้อมูลอย่างน้อย " + strconv.Itoa(req.StartRow) + " แถว",
			}},
		}

		h.setBatchStatus(batchID, &models.CouponBatchImportStatus{
			BatchID:   batchID,
			Status:    "failed",
			Progress:  100,
			TotalRows: totalDataRows,
			Message:   "ไฟล์ไม่มีข้อมูลเพียงพอ",
		})

		return response, nil
	}

	// ตรวจสอบ header
	if len(rows) == 0 || len(rows[0]) < 12 {
		return &models.CouponExcelImportResponse{
			Success:    false,
			TotalRows:  len(rows),
			ErrorCount: 1,
			Message:    "รูปแบบไฟล์ไม่ถูกต้อง",
			Errors: []models.CouponImportError{{
				Row:   1,
				Error: "Header ไม่ครบหรือรูปแบบไม่ถูกต้อง ต้องมีอย่างน้อย 12 คอลัมน์",
			}},
		}, nil
	}

	// เก็บรหัสคูปองเพื่อเช็คความซ้ำ
	couponCodes := make(map[string]int) // couponcode -> row number
	var errors []models.CouponImportError
	importedCount := 0

	// ประมวลผลแต่ละแถว
	for i := req.StartRow - 1; i < len(rows); i++ {
		row := rows[i]
		rowNum := i + 1
		processedRows := i - req.StartRow + 2

		// คำนวณความคืบหน้า
		progress := int(float64(processedRows) / float64(totalDataRows) * 100)
		h.updateBatchProgress(batchID, progress, totalDataRows, processedRows-1, fmt.Sprintf("กำลังประมวลผลแถวที่ %d", rowNum))

		if len(row) < 12 {
			errors = append(errors, models.CouponImportError{
				Row:   rowNum,
				Error: "ข้อมูลในแถวไม่ครบ ต้องมีอย่างน้อย 12 คอลัมน์",
			})
			continue
		}

		couponCode := strings.TrimSpace(row[0])
		if couponCode == "" {
			errors = append(errors, models.CouponImportError{
				Row:        rowNum,
				CouponCode: couponCode,
				Field:      "couponcode",
				Error:      "รหัสคูปองไม่สามารถเป็นค่าว่างได้",
			})
			continue
		}

		// เช็ครหัสคูปองซ้ำในไฟล์
		if existingRow, exists := couponCodes[couponCode]; exists {
			errors = append(errors, models.CouponImportError{
				Row:        rowNum,
				CouponCode: couponCode,
				Field:      "couponcode",
				Error:      fmt.Sprintf("รหัสคูปองซ้ำกับแถวที่ %d", existingRow),
			})
			continue
		}
		couponCodes[couponCode] = rowNum

		// เช็ครหัสคูปองซ้ำในฐานข้อมูล
		exists, err := h.checkCouponCodeExists(shopID, couponCode)
		if err != nil {
			errors = append(errors, models.CouponImportError{
				Row:        rowNum,
				CouponCode: couponCode,
				Field:      "couponcode",
				Error:      "ไม่สามารถตรวจสอบรหัสคูปองในฐานข้อมูลได้: " + err.Error(),
			})
			continue
		}

		if exists && !req.OverwriteExisting {
			errors = append(errors, models.CouponImportError{
				Row:        rowNum,
				CouponCode: couponCode,
				Field:      "couponcode",
				Error:      "มีรหัสคูปองนี้อยู่ในระบบแล้ว",
			})
			continue
		}

		// ถ้าเป็น validate only ไม่ต้องบันทึก
		if !req.ValidateOnly {
			// บันทึกข้อมูลคูปอง
			err = h.saveCouponFromRow(shopID, authUsername, row, rowNum)
			if err != nil {
				errors = append(errors, models.CouponImportError{
					Row:        rowNum,
					CouponCode: couponCode,
					Error:      "ไม่สามารถบันทึกข้อมูลได้: " + err.Error(),
				})
				continue
			}
		}

		importedCount++

		// อัพเดท progress หลังจากบันทึกสำเร็จ
		finalProgress := int(float64(processedRows) / float64(totalDataRows) * 100)
		h.updateBatchProgress(batchID, finalProgress, totalDataRows, importedCount, fmt.Sprintf("บันทึกสำเร็จแถวที่ %d", rowNum))
	}

	// ถ้ามี error และไม่ได้เป็น validate only ให้ rollback
	if len(errors) > 0 && !req.ValidateOnly {
		h.setBatchStatus(batchID, &models.CouponBatchImportStatus{
			BatchID:       batchID,
			Status:        "failed",
			Progress:      100,
			TotalRows:     len(rows) - req.StartRow + 1,
			ProcessedRows: 0,
			SuccessCount:  0,
			ErrorCount:    len(errors),
			Message:       "พบข้อผิดพลาด ไม่บันทึกข้อมูลทั้งไฟล์",
			UpdatedAt:     time.Now(),
		})

		return &models.CouponExcelImportResponse{
			BatchID:       batchID,
			Success:       false,
			ImportedCount: 0,
			ErrorCount:    len(errors),
			TotalRows:     len(rows) - req.StartRow + 1,
			Message:       "พบข้อผิดพลาด ไม่บันทึกข้อมูลทั้งไฟล์",
			Errors:        errors,
		}, nil
	}

	message := "ประมวลผลสำเร็จ"
	if req.ValidateOnly {
		message = "ตรวจสอบข้อมูลเสร็จสิ้น"
	}

	// อัพเดท batch status เป็น completed
	finalStatus := "completed"
	if len(errors) > 0 {
		finalStatus = "completed_with_errors"
	}

	h.setBatchStatus(batchID, &models.CouponBatchImportStatus{
		BatchID:       batchID,
		Status:        finalStatus,
		Progress:      100,
		TotalRows:     len(rows) - req.StartRow + 1,
		ProcessedRows: importedCount,
		SuccessCount:  importedCount,
		ErrorCount:    len(errors),
		Message:       message,
		UpdatedAt:     time.Now(),
	})

	return &models.CouponExcelImportResponse{
		BatchID:       batchID,
		Success:       len(errors) == 0,
		ImportedCount: importedCount,
		ErrorCount:    len(errors),
		TotalRows:     len(rows) - req.StartRow + 1,
		Message:       message,
		Errors:        errors,
	}, nil
}

// เช็ครหัสคูปองซ้ำในฐานข้อมูล
func (h CouponHttp) checkCouponCodeExists(shopID, couponCode string) (bool, error) {
	// เรียก service เพื่อเช็คในฐานข้อมูล
	// ใช้ SearchCoupon แล้วเช็คว่ามีรหัสคูปองนี้หรือไม่
	searchResult, err := h.svc.SearchCoupon(shopID, couponCode)
	if err != nil {
		return false, err
	}

	// ตรวจสอบว่ามีคูปองที่มีรหัสตรงกันหรือไม่
	if len(searchResult) > 0 {
		for _, coupon := range searchResult {
			if coupon.CouponCode == couponCode {
				return true, nil
			}
		}
	}

	return false, nil
}

// บันทึกข้อมูลคูปองจากแถว Excel
func (h CouponHttp) saveCouponFromRow(shopID, authUsername string, row []string, rowNum int) error {
	// Parse ข้อมูลจาก Excel row
	couponData := models.Coupon{}

	// Basic fields (ต้องมี)
	couponData.CouponCode = strings.TrimSpace(row[0])
	name := strings.TrimSpace(row[1])
	code := "th"
	couponData.Names = &[]common.NameX{{Code: &code, Name: &name}}

	// Parse CouponValue
	if couponValue, err := strconv.ParseFloat(strings.TrimSpace(row[2]), 64); err == nil {
		couponData.CouponValue = couponValue
	} else {
		return fmt.Errorf("invalid coupon value at row %d", rowNum)
	}

	// Parse dates
	if issuedDate, err := time.Parse("2006-01-02", strings.TrimSpace(row[3])); err == nil {
		// Set to UTC at start of day
		couponData.IssuedDate = time.Date(issuedDate.Year(), issuedDate.Month(), issuedDate.Day(), 0, 0, 0, 0, time.UTC)
	} else {
		return fmt.Errorf("invalid issued date format at row %d, expected YYYY-MM-DD", rowNum)
	}

	if expiryDate, err := time.Parse("2006-01-02", strings.TrimSpace(row[4])); err == nil {
		// Set to UTC at end of day (23:59:59) so coupon is valid until end of expiry date
		couponData.ExpiryDate = time.Date(expiryDate.Year(), expiryDate.Month(), expiryDate.Day(), 23, 59, 59, 999999999, time.UTC)
	} else {
		return fmt.Errorf("invalid expiry date format at row %d, expected YYYY-MM-DD", rowNum)
	}

	// Parse CouponType
	if couponType, err := strconv.ParseInt(strings.TrimSpace(row[5]), 10, 8); err == nil {
		couponData.CouponType = models.CouponType(couponType)
	} else {
		return fmt.Errorf("invalid coupon type at row %d", rowNum)
	}

	// Parse CustomerCodes (optional)
	if customerCodes := strings.TrimSpace(row[6]); customerCodes != "" {
		couponData.CustomerCodes = strings.Split(customerCodes, ";")
	}

	// Remark (optional)
	couponData.Remark = strings.TrimSpace(row[7])

	// Parse Status
	if status, err := strconv.ParseInt(strings.TrimSpace(row[8]), 10, 8); err == nil {
		couponData.Status = models.CouponStatus(status)
	} else {
		couponData.Status = models.CouponStatusActive // default
	}

	// Parse IsOneTimeUse
	if oneTimeUse := strings.TrimSpace(row[9]); oneTimeUse == "true" || oneTimeUse == "TRUE" {
		couponData.IsOneTimeUse = true
	}

	// Parse MaxUsageCount
	if maxUsage, err := strconv.Atoi(strings.TrimSpace(row[10])); err == nil {
		couponData.MaxUsageCount = maxUsage
	}

	// Parse MaxUsageCountPerCustomer
	if maxUsagePerCustomer, err := strconv.Atoi(strings.TrimSpace(row[11])); err == nil {
		couponData.MaxUsageCountPerCustomer = maxUsagePerCustomer
	}

	// Parse Product Condition - เพิ่ม debug logging
	fmt.Printf("DEBUG: Row %d - Processing product condition, row length: %d\n", rowNum, len(row))
	for i, cell := range row {
		if i >= 12 && i <= 23 {
			fmt.Printf("DEBUG: Column %d (index %d): '%s'\n", i+1, i, cell)
		}
	}

	// แก้ไขเงื่อนไข: ตรวจสอบตั้งแต่คอลัมน์ 13 ขึ้นไป (index 12+)
	if len(row) >= 13 {
		condition := &models.CouponProductCondition{}
		hasCondition := false

		// Parse แต่ละฟิลด์และเช็คว่ามีข้อมูลหรือไม่
		if len(row) > 12 && strings.TrimSpace(row[12]) != "" {
			condition.ProductCodes = strings.Split(strings.TrimSpace(row[12]), ";")
			hasCondition = true
			fmt.Printf("DEBUG: Found ProductCodes: %v\n", condition.ProductCodes)
		}
		if len(row) > 13 && strings.TrimSpace(row[13]) != "" {
			condition.GroupCodes = strings.Split(strings.TrimSpace(row[13]), ";")
			hasCondition = true
			fmt.Printf("DEBUG: Found GroupCodes: %v\n", condition.GroupCodes)
		}
		if len(row) > 14 && strings.TrimSpace(row[14]) != "" {
			condition.GroupSubOneCodes = strings.Split(strings.TrimSpace(row[14]), ";")
			hasCondition = true
			fmt.Printf("DEBUG: Found GroupSubOneCodes: %v\n", condition.GroupSubOneCodes)
		}
		if len(row) > 15 && strings.TrimSpace(row[15]) != "" {
			condition.GroupSubTwoCodes = strings.Split(strings.TrimSpace(row[15]), ";")
			hasCondition = true
			fmt.Printf("DEBUG: Found GroupSubTwoCodes: %v\n", condition.GroupSubTwoCodes)
		}
		if len(row) > 16 && strings.TrimSpace(row[16]) != "" {
			condition.BrandCodes = strings.Split(strings.TrimSpace(row[16]), ";")
			hasCondition = true
			fmt.Printf("DEBUG: Found BrandCodes: %v (from row[16]='%s')\n", condition.BrandCodes, row[16])
		}
		if len(row) > 17 && strings.TrimSpace(row[17]) != "" {
			condition.DesignCodes = strings.Split(strings.TrimSpace(row[17]), ";")
			hasCondition = true
			fmt.Printf("DEBUG: Found DesignCodes: %v\n", condition.DesignCodes)
		}
		if len(row) > 18 && strings.TrimSpace(row[18]) != "" {
			condition.ModelCodes = strings.Split(strings.TrimSpace(row[18]), ";")
			hasCondition = true
			fmt.Printf("DEBUG: Found ModelCodes: %v\n", condition.ModelCodes)
		}
		if len(row) > 19 && strings.TrimSpace(row[19]) != "" {
			condition.PatternCodes = strings.Split(strings.TrimSpace(row[19]), ";")
			hasCondition = true
			fmt.Printf("DEBUG: Found PatternCodes: %v\n", condition.PatternCodes)
		}
		if len(row) > 20 && strings.TrimSpace(row[20]) != "" {
			condition.GradeCodes = strings.Split(strings.TrimSpace(row[20]), ";")
			hasCondition = true
			fmt.Printf("DEBUG: Found GradeCodes: %v\n", condition.GradeCodes)
		}
		if len(row) > 21 && strings.TrimSpace(row[21]) != "" {
			condition.CategoryCodes = strings.Split(strings.TrimSpace(row[21]), ";")
			hasCondition = true
			fmt.Printf("DEBUG: Found CategoryCodes: %v\n", condition.CategoryCodes)
		}
		if len(row) > 22 && strings.TrimSpace(row[22]) != "" {
			condition.ClassCodes = strings.Split(strings.TrimSpace(row[22]), ";")
			hasCondition = true
			fmt.Printf("DEBUG: Found ClassCodes: %v\n", condition.ClassCodes)
		}
		if len(row) > 23 && strings.TrimSpace(row[23]) != "" {
			if minimumAmount, err := strconv.ParseFloat(strings.TrimSpace(row[23]), 64); err == nil {
				condition.MinimumAmount = minimumAmount
				hasCondition = true
				fmt.Printf("DEBUG: Found MinimumAmount: %.2f\n", condition.MinimumAmount)
			}
		}

		fmt.Printf("DEBUG: hasCondition: %t\n", hasCondition)

		// ถ้ามีเงื่อนไขอย่างน้อย 1 อย่าง ถึงจะเซ็ต ProductCondition
		if hasCondition {
			couponData.ProductCondition = condition
			fmt.Printf("DEBUG: Setting ProductCondition: %+v\n", condition)
		} else {
			couponData.ProductCondition = nil
			fmt.Printf("DEBUG: No product condition found, setting to nil\n")
		}
	} else {
		couponData.ProductCondition = nil
		fmt.Printf("DEBUG: Row length %d < 13, no product condition\n", len(row))
	}

	// Parse IgnoreBranchCode (ถ้ามีข้อมูลครบ 25 คอลัมน์)
	if len(row) >= 25 {
		if ignoreBranchCodes := strings.TrimSpace(row[24]); ignoreBranchCodes != "" {
			couponData.IgnoreBranchCode = strings.Split(ignoreBranchCodes, ";")
		}
	}

	// Apply validation logic
	if couponData.IsOneTimeUse {
		couponData.MaxUsageCount = 1
		couponData.MaxUsageCountPerCustomer = 1
	} else {
		if len(couponData.CustomerCodes) == 0 {
			// Global mode: ต้องมี MaxUsageCount
			if couponData.MaxUsageCount <= 0 {
				return fmt.Errorf("MaxUsageCount is required when CustomerCodes is empty (global mode) at row %d", rowNum)
			}
		} else {
			// Per-Customer mode: ต้องมี MaxUsageCountPerCustomer
			if couponData.MaxUsageCountPerCustomer <= 0 {
				return fmt.Errorf("MaxUsageCountPerCustomer is required when CustomerCodes is specified at row %d", rowNum)
			}
		}
	}

	// เรียก service เพื่อบันทึกข้อมูล
	_, err := h.svc.CreateCoupon(shopID, authUsername, couponData)
	return err
}
