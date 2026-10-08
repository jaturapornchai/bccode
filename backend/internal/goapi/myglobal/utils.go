package myglobal

import (
	"compress/gzip"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"smlcloudplatform/internal/goapi/logger"
	"smlcloudplatform/internal/goapi/models"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

func GenUUID() string {
	// create guid
	return uuid.New().String()
}

func RoundFloat64(value float64, precision int) float64 {
	multiplier := math.Pow10(precision)
	return math.Round(value*multiplier) / multiplier
}

func FormatDateFullThai(date time.Time) string {
	// แปลงเป็นปีไทย พ.ศ. เช่น 1 มกราคม 2564
	thaiYear := date.Year() + 543

	// กำหนดชื่อเดือนภาษาไทย
	thaiMonths := []string{
		"มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน",
		"กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม",
	}

	// จัดรูปแบบวันที่เป็น "วันที่ เดือน พ.ศ."
	formattedDate := fmt.Sprintf("%d %s %d", date.Day(), thaiMonths[date.Month()-1], thaiYear)

	return formattedDate
}

func FormatDateFull(date time.Time) string {
	// แปลงเป็นปีไทย พ.ศ. เช่น 1 มกราคม 2564
	thaiYear := (date.Year() + 543) - 2500

	// จัดรูปแบบวันที่เป็น "วันที่ เดือน พ.ศ."
	formattedDate := fmt.Sprintf("%02d/%02d/%02d", date.Day(), date.Month(), thaiYear)

	return formattedDate
}

func FormatNumber(number float64, point int) string {
	if number == 0 || math.IsNaN(number) {
		return ""
	}

	// แยกส่วนจำนวนเต็มกับทศนิยม
	format := fmt.Sprintf("%%.%df", point)
	numStr := fmt.Sprintf(format, number)

	// แยกส่วนจำนวนเต็มออกมา
	parts := strings.Split(numStr, ".")
	integerPart := parts[0]

	// เพิ่มคอมมาในส่วนจำนวนเต็ม
	var result string
	for i, digit := range integerPart {
		if i > 0 && (len(integerPart)-i)%3 == 0 && digit != '-' {
			result += ","
		}
		result += string(digit)
	}

	// ถ้ามีทศนิยม ให้ต่อส่วนทศนิยมกลับเข้าไป
	if len(parts) > 1 {
		result += "." + parts[1]
	}

	return result
}

func ConvertPDFToGzip(inputPath, outputPath string) error {
	// เปิดไฟล์ input
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("ไม่สามารถเปิดไฟล์ input: %v", err)
	}
	defer inputFile.Close()

	// สร้างไฟล์ output
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("ไม่สามารถสร้างไฟล์ output: %v", err)
	}
	defer outputFile.Close()

	// สร้าง gzip writer
	gzipWriter := gzip.NewWriter(outputFile)
	defer gzipWriter.Close()

	// คัดลอกข้อมูลจาก input ไปยัง gzip writer
	_, err = io.Copy(gzipWriter, inputFile)
	if err != nil {
		return fmt.Errorf("ไม่สามารถคัดลอกข้อมูล: %v", err)
	}

	// ปิด gzip writer เพื่อให้แน่ใจว่าข้อมูลทั้งหมดถูกเขียน
	err = gzipWriter.Close()
	if err != nil {
		return fmt.Errorf("ไม่สามารถปิด gzip writer: %v", err)
	}

	return nil
}

func DeleteFilesInFolder(folderPath string, beginName string) error {
	// สร้าง pattern เพื่อหาทุกไฟล์ใน folderPath
	pattern := filepath.Join(folderPath, "*")
	logger.Info("ค้นหาไฟล์ทั้งหมดด้วย pattern: %s", pattern)

	// ค้นหาทุกไฟล์ใน folder
	files, err := filepath.Glob(pattern)
	if err != nil {
		return err
	}

	// ตรวจสอบว่าพบไฟล์หรือไม่
	if len(files) == 0 {
		logger.Info("ไม่พบไฟล์ใน %s", folderPath)
		return nil // ไม่ถือเป็น error ถ้าไม่พบไฟล์
	}

	currentTime := time.Now()
	deletedCount := 0

	// ลบไฟล์แต่ละไฟล์
	for _, file := range files {
		// ตรวจสอบว่าเป็นไฟล์จริงๆ ไม่ใช่ directory
		fileInfo, err := os.Stat(file)
		if err != nil {
			logger.Info("เกิดข้อผิดพลาดในการอ่านข้อมูลไฟล์ %s: %v", file, err)
			continue // ข้ามไปถ้ามี error
		}

		// ข้ามถ้าเป็นโฟลเดอร์
		if fileInfo.IsDir() {
			continue
		}

		// ตรวจสอบว่าไฟล์มีคำนำหน้าตามที่กำหนดหรือไม่
		fileName := filepath.Base(file)
		if beginName != "" && !strings.HasPrefix(fileName, beginName) {
			continue
		}

		// คำนวณอายุของไฟล์
		fileAge := currentTime.Sub(fileInfo.ModTime())
		fileAgeHours := fileAge.Hours()

		// ลบไฟล์ที่มีอายุเกินกำหนด
		if fileAgeHours > 24 {
			logger.Info("กำลังลบไฟล์ที่มีอายุ %.2f ชั่วโมง: %s", fileAgeHours, file)
			err = os.Remove(file)
			if err != nil {
				logger.Info("ล้มเหลวในการลบไฟล์ %s: %v", file, err)
				return err
			}
			deletedCount++
		}
	}

	logger.Info("ลบไฟล์เก่าทั้งหมด %d ไฟล์เรียบร้อยแล้ว", deletedCount)
	return nil
}

func ConvertStringToInt(str string) (int, error) {
	// แปลงสตริงเป็นจำนวนเต็ม
	result, err := strconv.Atoi(str)
	if err != nil {
		return 0, fmt.Errorf("ไม่สามารถแปลงสตริงเป็นจำนวนเต็มได้: %v", err)
	}
	return result, nil
}

func CalcStockQtyWord(qty float64, barcodePacking []models.ProductBarcodePackingStruct) string {
	calcBalanceWord := ""
	calcBalanceQty := qty
	// ถ้าติดลบ
	if calcBalanceQty < 0 {
		calcBalanceQty = calcBalanceQty * -1
		calcBalanceWord = "(ลบ) "
	}
	for index, packing := range barcodePacking {
		// logger.Info("UnitName: %s, BarcodeRefUnitStand: %f, BarcodeRefUnitDivide: %f", packing.UnitName, packing.BarcodeRefUnitStand, packing.BarcodeRefUnitDivide)
		// หาจำนวนเต็ม
		if packing.BarcodeRefUnitDivide == 0 {
			calcBalanceQty = 0
			calcBalanceWord = "Error"
			break
		}
		// หาจำนวนเต็ม
		calcQty := 0.0
		if packing.BarcodeRefUnitDivide > packing.BarcodeRefUnitStand {
			calcQty = calcBalanceQty * packing.BarcodeRefUnitDivide
		} else {
			calcQty = calcBalanceQty / packing.BarcodeRefUnitStand
		}
		if index != len(barcodePacking)-1 {
			// ปัดเศษลง
			calcQty = math.Floor(calcQty)
		}
		// logger.Info("calcQty: %f %f %f", calcQty, packing.BarcodeRefUnitStand, packing.BarcodeRefUnitDivide)
		if calcQty > 0 {
			if calcBalanceWord != "" {
				calcBalanceWord = calcBalanceWord + "/"
			}
			calcBalanceWord = calcBalanceWord + fmt.Sprintf("%s %s", CalcStockQtyWordCut(calcQty), packing.UnitName)
		}
		// หาเศษ
		calcBalanceQty = calcBalanceQty - (calcQty * packing.BarcodeRefUnitStand)
		// logger.Info("calcBalanceQty: %f", calcBalanceQty)
	}
	return calcBalanceWord
}

func CalcStockQtyWordCut(calcQty float64) string {
	// หาทศนิยมกี่ตำแหน่ง
	qtyWord := fmt.Sprintf("%f", calcQty)
	// ลบ 0 ท้ายสุดมาเรื่อยๆ จนกว่าจะเจอ . และลบ . แต่ถ้าเจอตัวเลข 1-9 ให้หยุด
	for {
		if qtyWord[len(qtyWord)-1:] == "0" {
			qtyWord = qtyWord[:len(qtyWord)-1]
		} else if qtyWord[len(qtyWord)-1:] == "." {
			qtyWord = qtyWord[:len(qtyWord)-1]
			break
		} else {
			break
		}
	}
	return qtyWord
}

func SafeFloatConversion(value any) float64 {
	if value == nil {
		return 0.0
	}

	switch v := value.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case string:
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return 0.0
}

// ParseFloat64 แปลงค่าจาก any เป็น float64
func ParseFloat64(val any, defaultValue float64) float64 {
	if val == nil {
		return defaultValue
	}

	switch v := val.(type) {
	case float64:
		return v
	case float32:
		return float64(v)
	case int:
		return float64(v)
	case int64:
		return float64(v)
	case int32:
		return float64(v)
	case int16:
		return float64(v)
	case int8:
		return float64(v)
	case uint:
		return float64(v)
	case uint64:
		return float64(v)
	case uint32:
		return float64(v)
	case uint16:
		return float64(v)
	case uint8:
		return float64(v)
	case string:
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			return parsed
		}
	case []byte:
		if parsed, err := strconv.ParseFloat(string(v), 64); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// ParseInt แปลงค่าจาก any เป็น int
func ParseInt(val any, defaultValue int) int {
	if val == nil {
		return defaultValue
	}

	switch v := val.(type) {
	case int:
		return v
	case int64:
		return int(v)
	case int32:
		return int(v)
	case int16:
		return int(v)
	case int8:
		return int(v)
	case uint:
		return int(v)
	case uint64:
		return int(v)
	case uint32:
		return int(v)
	case uint16:
		return int(v)
	case uint8:
		return int(v)
	case float64:
		return int(v)
	case float32:
		return int(v)
	case string:
		if parsed, err := strconv.Atoi(v); err == nil {
			return parsed
		}
	case []byte:
		if parsed, err := strconv.Atoi(string(v)); err == nil {
			return parsed
		}
	}
	return defaultValue
}

// ParseString แปลงค่าจาก any เป็น string
func ParseString(val any, defaultValue string) string {
	if val == nil {
		return defaultValue
	}

	switch v := val.(type) {
	case string:
		return v
	case []byte:
		return string(v)
	case int, int64, int32, int16, int8, uint, uint64, uint32, uint16, uint8:
		return fmt.Sprintf("%d", v)
	case float64, float32:
		return fmt.Sprintf("%g", v)
	case bool:
		return strconv.FormatBool(v)
	}
	return fmt.Sprintf("%v", val)
}

func ParseNumericValue(value any, defaultValue float64) float64 {
	if value == nil {
		return defaultValue
	}

	switch val := value.(type) {
	case float64:
		return val
	case string:
		if parsed, err := strconv.ParseFloat(val, 64); err == nil {
			return parsed
		}
	case int64:
		return float64(val)
	case int:
		return float64(val)
	}
	return defaultValue
}
