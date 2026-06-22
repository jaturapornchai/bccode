package services

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strconv"
	"strings"
	"time"

	common "smlcloudplatform/internal/models"
	branch_repo "smlcloudplatform/internal/organization/branch/repositories"
	businesstype_repo "smlcloudplatform/internal/organization/businesstype/repositories"
	product_models "smlcloudplatform/internal/product/productbarcode/models"
	productbarcode_repo "smlcloudplatform/internal/product/productbarcode/repositories"
	product_services "smlcloudplatform/internal/product/productbarcode/services"
	productunit_models "smlcloudplatform/internal/product/unit/models"
	productunit_repo "smlcloudplatform/internal/product/unit/repositories"
	"smlcloudplatform/internal/productimport/models"
	"smlcloudplatform/internal/productimport/repositories"
	brandproduct_models "smlcloudplatform/internal/smlaiproduct/brandproduct/models"
	brandproduct_repo "smlcloudplatform/internal/smlaiproduct/brandproduct/repositories"
	categoryproduct_models "smlcloudplatform/internal/smlaiproduct/categoryproduct/models"
	categoryproduct_repo "smlcloudplatform/internal/smlaiproduct/categoryproduct/repositories"
	classproduct_models "smlcloudplatform/internal/smlaiproduct/classproduct/models"
	classproduct_repo "smlcloudplatform/internal/smlaiproduct/classproduct/repositories"
	designproduct_models "smlcloudplatform/internal/smlaiproduct/designproduct/models"
	designproduct_repo "smlcloudplatform/internal/smlaiproduct/designproduct/repositories"
	gradeproduct_models "smlcloudplatform/internal/smlaiproduct/gradeproduct/models"
	gradeproduct_repo "smlcloudplatform/internal/smlaiproduct/gradeproduct/repositories"
	groupproduct_models "smlcloudplatform/internal/smlaiproduct/groupproduct/models"
	groupproduct_repo "smlcloudplatform/internal/smlaiproduct/groupproduct/repositories"
	groupsuboneproduct_models "smlcloudplatform/internal/smlaiproduct/groupsuboneproduct/models"
	groupsuboneproduct_repo "smlcloudplatform/internal/smlaiproduct/groupsuboneproduct/repositories"
	groupsubtwoproduct_models "smlcloudplatform/internal/smlaiproduct/groupsubtwoproduct/models"
	groupsubtwoproduct_repo "smlcloudplatform/internal/smlaiproduct/groupsubtwoproduct/repositories"
	modelproduct_models "smlcloudplatform/internal/smlaiproduct/modelproduct/models"
	modelproduct_repo "smlcloudplatform/internal/smlaiproduct/modelproduct/repositories"
	patternproduct_models "smlcloudplatform/internal/smlaiproduct/patternproduct/models"
	patternproduct_repo "smlcloudplatform/internal/smlaiproduct/patternproduct/repositories"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/xuri/excelize/v2"
	"go.mongodb.org/mongo-driver/bson"
)

type IProductImportService interface {
	List(holdingCode string, taskID string, pageable micromodels.Pageable) ([]models.ProductImportInfo, models.PaginationData, error)
	Create(holdingCode string, authUsername string, req *models.ProductImport) error
	Update(holdingCode string, guid string, doc models.ProductImportRaw) error
	Delete(holdingCode string, guid string) error
	DeleteTask(holdingCode string, taskID string) error
	ImportFromFile(holdingCode string, authUsername string, fileUpload io.Reader) (string, error)
	SaveTask(holdingCode string, authUsername string, taskID string, docHeader models.ProductImportHeader) error
	Verify(holdingCode string, taskID string) error
	GetTaskStatus(holdingCode string, taskID string) (models.TaskStatusModel, error)

	// Enhanced methods สำหรับ compare และ insert/update functionality
	SaveTaskWithMode(holdingCode string, authUsername string, taskID string, docHeader models.ProductImportHeader, importMode string, forceUpdate bool) error
	PreviewSave(holdingCode string, taskID string, docHeader models.ProductImportHeader, importMode string) (*models.CompareResult, error)
	CompareWithExisting(holdingCode string, taskID string) (*models.CompareResult, error)
	GetCompareResult(holdingCode string, taskID string) (*models.CompareResult, error)
	PreviewChanges(holdingCode string, taskID string, importMode string) (*models.PreviewResult, error)
	ApplyChanges(holdingCode string, authUsername string, taskID string, importMode string, forceUpdate bool) error
	GetImportSummary(holdingCode string, taskID string) (*models.ImportSummary, error)

	// 🚀 เพิ่ม methods ใหม่สำหรับ pagination และ performance
	GetCompareResultPaginated(holdingCode string, taskID string, pageable micromodels.Pageable) (*models.CompareResult, models.PaginationData, error)
	GetCompareResultSummary(holdingCode string, taskID string) (*models.CompareSummary, error)

	// 🔧 เพิ่ม methods สำหรับ task status management และ progress tracking
	CreateTaskStatus(holdingCode string, taskID string, status string, progress int, message string) error
	UpdateTaskStatus(holdingCode string, taskID string, status string, progress int, message string) error
	ApplyChangesWithProgress(holdingCode string, authUsername string, taskID string, importMode string, forceUpdate bool, batchSize ...int) error
}

type ProductImportService struct {
	deafultPartSize        int
	sizeID                 int
	cacheExpire            time.Duration
	chRepo                 repositories.IProductImportClickHouseRepository
	taskStatusRepo         repositories.ITaskStatusRepository
	productBarcodeRepo     productbarcode_repo.IProductBarcodeRepository
	productBarcodeService  product_services.IProductBarcodeHttpService
	productUnitRepo        productunit_repo.IUnitRepository
	groupProductRepo       groupproduct_repo.IGroupProductRepository
	groupsuboneProductRepo groupsuboneproduct_repo.IGroupsuboneProductRepository
	groupsubtwoproductRepo groupsubtwoproduct_repo.IGroupsubtwoProductRepository
	brandProductRepo       brandproduct_repo.IBrandProductRepository
	designProductRepo      designproduct_repo.IDesignProductRepository
	modelProductRepo       modelproduct_repo.IModelProductRepository
	patternProductRepo     patternproduct_repo.IPatternProductRepository
	gradeProductRepo       gradeproduct_repo.IGradeProductRepository
	categoryProductRepo    categoryproduct_repo.ICategoryProductRepository
	classProductRepo       classproduct_repo.IClassProductRepository
	branchRepo             branch_repo.IBranchRepository
	businessTypeRepo       businesstype_repo.IBusinessTypeRepository
	generateID             func(int) string
	generateGUID           func() string
	timeNow                func() time.Time
}

func NewProductImportService(
	chRepo repositories.IProductImportClickHouseRepository,
	taskStatusRepo repositories.ITaskStatusRepository,
	productBarcodeRepo productbarcode_repo.IProductBarcodeRepository,
	productBarcodeService product_services.IProductBarcodeHttpService,
	productUnitRepo productunit_repo.IUnitRepository,
	groupProductRepo groupproduct_repo.IGroupProductRepository,
	groupsuboneProductRepo groupsuboneproduct_repo.IGroupsuboneProductRepository,
	groupsubtwoproductRepo groupsubtwoproduct_repo.IGroupsubtwoProductRepository,
	brandProductRepo brandproduct_repo.IBrandProductRepository,
	designProductRepo designproduct_repo.IDesignProductRepository,
	modelProductRepo modelproduct_repo.IModelProductRepository,
	patternProductRepo patternproduct_repo.IPatternProductRepository,
	gradeProductRepo gradeproduct_repo.IGradeProductRepository,
	categoryProductRepo categoryproduct_repo.ICategoryProductRepository,
	classProductRepo classproduct_repo.IClassProductRepository,
	branchRepo branch_repo.IBranchRepository,
	businessTypeRepo businesstype_repo.IBusinessTypeRepository,
	generateID func(int) string,
	generateGUID func() string,
	timeNow func() time.Time,
) *ProductImportService {
	return &ProductImportService{
		deafultPartSize:        100,
		sizeID:                 12,
		cacheExpire:            time.Minute * 60,
		chRepo:                 chRepo,
		taskStatusRepo:         taskStatusRepo,
		productBarcodeRepo:     productBarcodeRepo,
		productBarcodeService:  productBarcodeService,
		productUnitRepo:        productUnitRepo,
		groupProductRepo:       groupProductRepo,
		groupsuboneProductRepo: groupsuboneProductRepo,
		groupsubtwoproductRepo: groupsubtwoproductRepo,
		brandProductRepo:       brandProductRepo,
		designProductRepo:      designProductRepo,
		modelProductRepo:       modelProductRepo,
		patternProductRepo:     patternProductRepo,
		gradeProductRepo:       gradeProductRepo,
		categoryProductRepo:    categoryProductRepo,
		classProductRepo:       classProductRepo,
		branchRepo:             branchRepo,
		businessTypeRepo:       businessTypeRepo,
		generateID:             generateID,
		generateGUID:           generateGUID,
		timeNow:                timeNow,
	}
}

func (svc *ProductImportService) List(holdingCode string, taskID string, pageable micromodels.Pageable) ([]models.ProductImportInfo, models.PaginationData, error) {
	findDocs, pagination, err := svc.chRepo.List(context.Background(), holdingCode, taskID, pageable)

	if err != nil {
		return []models.ProductImportInfo{}, models.PaginationData{}, err
	}

	results := []models.ProductImportInfo{}

	for _, doc := range findDocs {
		results = append(results, doc.ProductImportInfo)
	}

	return results, pagination, nil

}

func (svc *ProductImportService) Create(holdingCode string, authUsername string, doc *models.ProductImport) error {
	docData := models.ProductImportDoc{}
	docData.HoldingCode = holdingCode
	docData.GUIDFixed = svc.generateGUID()
	docData.ProductImport = *doc

	result, err := svc.chRepo.FindOne(context.Background(), holdingCode, doc.TaskID, []micromodels.KeyInt{
		{
			Key:   "rownumber",
			Value: -1,
		},
	})

	if err != nil {
		return err
	}

	if doc.RowNumber == 0 {
		docData.RowNumber = result.RowNumber + 1
	}

	docData.CreatedAt = svc.timeNow()
	docData.CreatedBy = authUsername

	return svc.chRepo.Create(context.Background(), docData)
}

func (svc *ProductImportService) ImportFromFile(holdingCode string, authUsername string, fileUpload io.Reader) (string, error) {

	f, err := excelize.OpenReader(fileUpload)
	if err != nil {
		return "", err
	}

	if len(f.GetSheetList()) == 0 {
		return "", errors.New("sheet not found")
	}

	sheetName := f.GetSheetList()[0]

	rows, err := f.GetRows(sheetName)
	if err != nil {
		return "", err
	}

	if len(rows) <= 1 {
		return "", errors.New("sheet is empty")
	}

	colIdxs := map[string]int{}

	cols := rows[0]
	for i, col := range cols {
		colIdxs[col] = i
	}

	// บังคับให้มี column เหล่านี้เท่านั้น
	requiredColumns := []string{
		"Barcode",
		"Name",
		"Unit Code",
		"Price",
		"Price Member",
	}

	// ตรวจสอบ required columns เท่านั้น
	isNotfoundColumn := false
	columnNotFound := []string{}
	for _, colName := range requiredColumns {
		if _, ok := colIdxs[colName]; !ok {
			isNotfoundColumn = true
			columnNotFound = append(columnNotFound, colName)
		}
	}

	if isNotfoundColumn {
		return "", fmt.Errorf("required column not found: %v", strings.Join(columnNotFound, ", "))
	}

	prepareDataDoc := []models.ProductImportDoc{}

	taskID := svc.generateID(svc.sizeID)

	for i, doc := range rows {

		if i == 0 {
			continue
		}

		tempDataList, err := svc.prepareData(holdingCode, taskID, float64(i), colIdxs, doc)
		if err != nil {
			return "", err
		}

		prepareDataDoc = append(prepareDataDoc, tempDataList...)
	}

	createdAt := time.Now()
	createdBy := authUsername

	for i := range prepareDataDoc {
		prepareDataDoc[i].CreatedAt = createdAt
		prepareDataDoc[i].CreatedBy = createdBy
	}

	err = svc.chRepo.CreateInBatch(context.Background(), prepareDataDoc)

	if err != nil {
		return "", err
	}
	// 🔧 เพิ่ม: สร้าง initial task status
	now := time.Now().UTC()
	log.Printf("Creating initial task status with timestamp: %v", now)

	status := models.TaskStatusModel{
		TaskID:      taskID,
		HoldingCode: holdingCode,
		Status:      "UPLOADED", // เปลี่ยนเป็น string แทน constant
		Progress:    int32(0),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := svc.taskStatusRepo.Create(context.Background(), status); err != nil {
		// Log error แต่ไม่ return เพราะ upload สำเร็จแล้ว
		log.Printf("Warning: Failed to create initial task status: %v", err)
	} else {
		log.Printf("Success: Created initial task status for task: %s", taskID)
	}

	return taskID, nil
}

func (svc *ProductImportService) prepareData(holdingCode string, taskID string, rowNumber float64, colIdx map[string]int, doc []string) ([]models.ProductImportDoc, error) {

	// Helper function เพื่อรับค่าจาก column อย่างปลอดภัย
	getColumnValue := func(columnName string) string {
		if idx, ok := colIdx[columnName]; ok && idx < len(doc) {
			return doc[idx]
		}
		return ""
	}

	getColumnValueNumber := func(columnName string) float64 {
		if idx, ok := colIdx[columnName]; ok && idx < len(doc) {
			val, _ := strconv.ParseFloat(doc[idx], 64)
			return val
		}
		return 0
	}

	getColumnValueBool := func(columnName string) bool {
		if idx, ok := colIdx[columnName]; ok && idx < len(doc) {
			value := strings.TrimSpace(strings.ToLower(doc[idx]))

			// รองรับหลายรูปแบบ
			switch value {
			case "true", "1", "yes", "y", "ใช่", "✓", "x", "v":
				return true
			case "false", "0", "no", "n", "ไม่ใช่", "":
				return false
			default:
				// พยายามแปลงเป็นตัวเลข
				if num, err := strconv.Atoi(value); err == nil {
					return num != 0
				}
				return false
			}
		}
		return false // default value
	}

	// Required fields - ต้องมีค่า
	priceStr := getColumnValue("Price")
	if priceStr == "" {
		return []models.ProductImportDoc{}, fmt.Errorf("price in row %d is empty", int(rowNumber))
	}
	price, err := strconv.ParseFloat(priceStr, 64)
	if err != nil {
		return []models.ProductImportDoc{}, fmt.Errorf("price in row %d invalid", int(rowNumber))
	}

	priceMemberStr := getColumnValue("Price Member")
	if priceMemberStr == "" {
		return []models.ProductImportDoc{}, fmt.Errorf("price member in row %d is empty", int(rowNumber))
	}
	priceMember, err := strconv.ParseFloat(priceMemberStr, 64)
	if err != nil {
		return []models.ProductImportDoc{}, fmt.Errorf("price member in row %d invalid", int(rowNumber))
	}

	barcodeRaw := getColumnValue("Barcode")
	if barcodeRaw == "" {
		return []models.ProductImportDoc{}, fmt.Errorf("barcode in row %d is empty", int(rowNumber))
	}

	name := getColumnValue("Name")
	if name == "" {
		return []models.ProductImportDoc{}, fmt.Errorf("name in row %d is empty", int(rowNumber))
	}

	unitCode := getColumnValue("Unit Code")
	if unitCode == "" {
		return []models.ProductImportDoc{}, fmt.Errorf("unit code in row %d is empty", int(rowNumber))
	}

	// แยก barcode ถ้ามี / (สำหรับ barcode หลายตัวที่ใช้ข้อมูลเดียวกัน)
	barcodes := strings.Split(barcodeRaw, "/")

	// ล้างค่าว่างออก
	var cleanBarcodes []string
	for _, barcode := range barcodes {
		barcode = strings.TrimSpace(barcode)
		if barcode != "" {
			cleanBarcodes = append(cleanBarcodes, barcode)
		}
	}

	if len(cleanBarcodes) == 0 {
		return []models.ProductImportDoc{}, fmt.Errorf("no valid barcode found in row %d", int(rowNumber))
	}

	var resultDocs []models.ProductImportDoc

	// สร้าง ProductImportDoc สำหรับแต่ละ barcode
	for i, barcode := range cleanBarcodes {
		newGUID := svc.generateGUID()

		dataDoc := models.ProductImportDoc{}
		dataDoc.GUIDFixed = newGUID
		dataDoc.HoldingCode = holdingCode
		dataDoc.TaskID = taskID
		dataDoc.RowNumber = rowNumber + float64(i)*0.001 // ใช้ decimal เพื่อแยก barcode ในแถวเดียวกัน
		dataDoc.Barcode = barcode
		dataDoc.Name = name
		dataDoc.UnitCode = unitCode
		dataDoc.BarcodeRef = getColumnValue("BarcodeRef")
		dataDoc.StandValue = getColumnValueNumber("StandValue")
		dataDoc.DivideValue = getColumnValueNumber("DivideValue")

		dataDoc.Price = price
		dataDoc.PriceMember = priceMember
		dataDoc.PriceDelivery = getColumnValueNumber("Price Delivery")

		//price 1-9
		dataDoc.PriceOne = getColumnValueNumber("PriceOne")
		dataDoc.PriceTwo = getColumnValueNumber("PriceTwo")
		dataDoc.PriceThree = getColumnValueNumber("PriceThree")
		dataDoc.PriceFour = getColumnValueNumber("PriceFour")
		dataDoc.PriceFive = getColumnValueNumber("PriceFive")
		dataDoc.PriceSix = getColumnValueNumber("PriceSix")
		dataDoc.PriceSeven = getColumnValueNumber("PriceSeven")
		dataDoc.PriceEight = getColumnValueNumber("PriceEight")
		dataDoc.PriceNine = getColumnValueNumber("PriceNine")

		// Optional fields - ถ้าไม่มีหรือว่างจะใช้ค่าว่างเป็น default
		dataDoc.Code = getColumnValue("Code")
		dataDoc.GroupCode = getColumnValue("GroupCode")
		dataDoc.GroupsuboneCode = getColumnValue("GroupsuboneCode")
		dataDoc.GroupsubtwoCode = getColumnValue("GroupsubtwoCode")
		dataDoc.BrandCode = getColumnValue("BrandCode")
		dataDoc.DesignCode = getColumnValue("DesignCode")
		dataDoc.ModelCode = getColumnValue("ModelCode")
		dataDoc.PatternCode = getColumnValue("PatternCode")
		dataDoc.GradeCode = getColumnValue("GradeCode")
		dataDoc.CategoryCode = getColumnValue("CategoryCode")
		dataDoc.ClassCode = getColumnValue("ClassCode")

		dataDoc.IsSumPoint = getColumnValueBool("IsSumPoint")

		resultDocs = append(resultDocs, dataDoc)
	}

	return resultDocs, nil
}

func (svc *ProductImportService) Update(holdingCode string, guid string, doc models.ProductImportRaw) error {
	return svc.chRepo.Update(context.Background(), holdingCode, guid, doc)
}

func (svc *ProductImportService) Delete(holdingCode string, guid string) error {
	return svc.chRepo.DeleteByGUID(context.Background(), holdingCode, guid)
}

func (svc *ProductImportService) DeleteTask(holdingCode string, taskID string) error {
	return svc.chRepo.DeleteByTaskID(context.Background(), holdingCode, taskID)
}

func (svc ProductImportService) Verify(holdingCode string, taskID string) error {
	docs, err := svc.chRepo.All(context.Background(), holdingCode, taskID)

	if err != nil {
		return err
	}

	previousDuplicate := map[string]struct{}{}
	previousExist := map[string]struct{}{}
	previousUnitNotExist := map[string]struct{}{}

	itemDulpicated := map[string]struct{}{}
	itemExist := map[string]struct{}{}
	itemUnitNotExist := map[string]struct{}{}

	tempBarcodes := []string{}
	tempUnitCodes := []string{}

	itemDict := map[string]struct{}{}
	for i, doc := range docs {

		if doc.IsExist {
			previousExist[doc.Barcode] = struct{}{}
		}

		if _, ok := itemDict[doc.Barcode]; ok {
			itemDulpicated[doc.Barcode] = struct{}{}
		} else {
			itemDict[doc.Barcode] = struct{}{}
			tempBarcodes = append(tempBarcodes, doc.Barcode)

			if doc.IsDuplicate {
				previousDuplicate[doc.Barcode] = struct{}{}
			}
		}

		if doc.IsUnitNotExist {
			previousUnitNotExist[doc.UnitCode] = struct{}{}
		}

		tempUnitCodes = append(tempUnitCodes, doc.UnitCode)

		if (i > 1 && i%5000 == 0) || i == len(docs)-1 {
			productList, err := svc.productBarcodeRepo.FindByBarcodes(context.Background(), holdingCode, tempBarcodes)
			if err != nil {
				return err
			}

			for _, product := range productList {
				itemExist[product.Barcode] = struct{}{}
				delete(itemDulpicated, product.Barcode)
			}

			unitList, err := svc.productUnitRepo.FindByUnitCodes(context.Background(), holdingCode, tempUnitCodes)
			if err != nil {
				return err
			}

			tempDictUnitCodes := map[string]struct{}{}
			for _, unit := range unitList {
				tempDictUnitCodes[unit.UnitCode] = struct{}{}
			}

			for _, unitCode := range tempUnitCodes {
				if _, ok := tempDictUnitCodes[unitCode]; !ok {
					itemUnitNotExist[unitCode] = struct{}{}
				}
			}

			//Clear previous duplicate
			if err := svc.updateDuplicate(holdingCode, taskID, false, previousDuplicate); err != nil {
				return err
			}

			//Update duplicate
			if err := svc.updateDuplicate(holdingCode, taskID, true, itemDulpicated); err != nil {
				return err
			}

			//Clear previous exist
			if err := svc.updateExist(holdingCode, taskID, false, previousExist); err != nil {
				return err
			}

			// Update exist
			if err := svc.updateExist(holdingCode, taskID, true, itemExist); err != nil {
				return err
			}

			// clear previous unit not exist
			if err := svc.updateUnitExist(holdingCode, taskID, false, previousUnitNotExist); err != nil {
				return err
			}

			// Update unit not exist
			if err := svc.updateUnitExist(holdingCode, taskID, true, itemUnitNotExist); err != nil {
				return err
			}

			previousDuplicate = map[string]struct{}{}
			previousExist = map[string]struct{}{}
			itemDulpicated = map[string]struct{}{}
			itemExist = map[string]struct{}{}
			tempBarcodes = []string{}
		}
	}

	return nil

}

func (svc ProductImportService) updateDuplicate(holdingCode string, taskID string, isDuplicate bool, barcodes map[string]struct{}) error {

	if len(barcodes) == 0 {
		return nil
	}

	tempPreviousDuplicate := []string{}
	for barcode := range barcodes {
		tempPreviousDuplicate = append(tempPreviousDuplicate, barcode)
	}
	err := svc.chRepo.UpdateDuplicate(context.Background(), holdingCode, taskID, isDuplicate, tempPreviousDuplicate)

	if err != nil {
		return err
	}

	return nil
}

func (svc ProductImportService) updateExist(holdingCode string, taskID string, isExist bool, barcodes map[string]struct{}) error {
	if len(barcodes) == 0 {
		return nil
	}

	tempPreviousExist := []string{}
	for barcode := range barcodes {
		tempPreviousExist = append(tempPreviousExist, barcode)
	}
	err := svc.chRepo.UpdateExist(context.Background(), holdingCode, taskID, isExist, tempPreviousExist)

	if err != nil {
		return err
	}

	return nil
}

func (svc ProductImportService) updateUnitExist(holdingCode string, taskID string, isExist bool, unitCodes map[string]struct{}) error {
	if len(unitCodes) == 0 {
		return nil
	}

	tempPreviousExist := []string{}
	for barcode := range unitCodes {
		tempPreviousExist = append(tempPreviousExist, barcode)
	}
	err := svc.chRepo.UpdateUnitExist(context.Background(), holdingCode, taskID, isExist, tempPreviousExist)

	if err != nil {
		return err
	}

	return nil
}

func (svc ProductImportService) SaveTask(holdingCode string, authUsername string, taskID string, docHeader models.ProductImportHeader) error {
	// สร้าง status เป็น processing
	status := models.TaskStatusModel{
		TaskID:      taskID,
		HoldingCode: holdingCode,
		Status:      models.TaskStatusModelProcessing,
		Progress:    int32(0),
		CreatedAt:   svc.timeNow(),
		UpdatedAt:   svc.timeNow(),
	}

	if err := svc.taskStatusRepo.Create(context.Background(), status); err != nil {
		return err
	}

	// Update progress 10%
	status.Progress = int32(10)
	status.UpdatedAt = svc.timeNow()
	if updateErr := svc.taskStatusRepo.Update(context.Background(), taskID, status); updateErr != nil {
		fmt.Printf("Failed to update task status progress: %v\n", updateErr)
	}

	err := svc.Verify(holdingCode, taskID)
	if err != nil {
		// Update status เป็น failed
		status.Status = models.TaskStatusModelFailed
		status.ErrorMsg = err.Error()
		status.UpdatedAt = svc.timeNow()
		if updateErr := svc.taskStatusRepo.Update(context.Background(), taskID, status); updateErr != nil {
			fmt.Printf("Failed to update task status to failed: %v\n", updateErr)
		}
		return err
	}

	// Update progress 30%
	status.Progress = int32(30)
	status.UpdatedAt = svc.timeNow()
	if updateErr := svc.taskStatusRepo.Update(context.Background(), taskID, status); updateErr != nil {
		fmt.Printf("Failed to update task status progress: %v\n", updateErr)
	}

	countDuplicate, err := svc.chRepo.CountDuplicate(context.Background(), holdingCode, taskID, true)
	if err != nil {
		status.Status = models.TaskStatusModelFailed
		status.ErrorMsg = "counting duplicate failed"
		status.UpdatedAt = svc.timeNow()
		if updateErr := svc.taskStatusRepo.Update(context.Background(), taskID, status); updateErr != nil {
			fmt.Printf("Failed to update task status to failed: %v\n", updateErr)
		}
		return errors.New("counting duplicate failed")
	}

	if countDuplicate > 0 {
		status.Status = models.TaskStatusModelFailed
		status.ErrorMsg = "items barcode duplicate"
		status.UpdatedAt = svc.timeNow()
		if updateErr := svc.taskStatusRepo.Update(context.Background(), taskID, status); updateErr != nil {
			fmt.Printf("Failed to update task status to failed: %v\n", updateErr)
		}
		return errors.New("items barcode duplicate")
	}

	// Update progress 40%
	status.Progress = int32(40)
	status.UpdatedAt = svc.timeNow()
	if updateErr := svc.taskStatusRepo.Update(context.Background(), taskID, status); updateErr != nil {
		fmt.Printf("Failed to update task status progress: %v\n", updateErr)
	}

	countExist, err := svc.chRepo.CountExist(context.Background(), holdingCode, taskID, true)
	if err != nil {
		status.Status = models.TaskStatusModelFailed
		status.ErrorMsg = "counting exist failed"
		status.UpdatedAt = svc.timeNow()
		if updateErr := svc.taskStatusRepo.Update(context.Background(), taskID, status); updateErr != nil {
			fmt.Printf("Failed to update task status to failed: %v\n", updateErr)
		}
		return errors.New("counting exist failed")
	}

	if countExist > 0 {
		status.Status = models.TaskStatusModelFailed
		status.ErrorMsg = "items barcode exist"
		status.UpdatedAt = svc.timeNow()
		if updateErr := svc.taskStatusRepo.Update(context.Background(), taskID, status); updateErr != nil {
			// Log update error but still return the original error
			fmt.Printf("Failed to update task status to failed: %v\n", updateErr)
		}
		return errors.New("items barcode exist")
	}

	// Update progress 50%
	status.Progress = int32(50)
	status.UpdatedAt = svc.timeNow()
	if updateErr := svc.taskStatusRepo.Update(context.Background(), taskID, status); updateErr != nil {
		fmt.Printf("Failed to update task status progress: %v\n", updateErr)
	}

	countUnitNotExist, err := svc.chRepo.CountUnitExist(context.Background(), holdingCode, taskID, true)
	if err != nil {
		status.Status = models.TaskStatusModelFailed
		status.ErrorMsg = "counting unit not exist failed"
		status.UpdatedAt = svc.timeNow()
		if updateErr := svc.taskStatusRepo.Update(context.Background(), taskID, status); updateErr != nil {
			fmt.Printf("Failed to update task status to failed: %v\n", updateErr)
		}
		return errors.New("counting unit not exist failed")
	}

	if countUnitNotExist > 0 {
		status.Status = models.TaskStatusModelFailed
		status.ErrorMsg = "items unit not exist"
		status.UpdatedAt = svc.timeNow()
		if updateErr := svc.taskStatusRepo.Update(context.Background(), taskID, status); updateErr != nil {
			fmt.Printf("Failed to update task status to failed: %v\n", updateErr)
		}
		return errors.New("items unit not exist")
	}

	// Update progress 60%
	status.Progress = int32(60)
	status.UpdatedAt = svc.timeNow()
	if updateErr := svc.taskStatusRepo.Update(context.Background(), taskID, status); updateErr != nil {
		fmt.Printf("Failed to update task status progress: %v\n", updateErr)
	}

	docs, err := svc.chRepo.All(context.Background(), holdingCode, taskID)
	if err != nil {
		status.Status = models.TaskStatusModelFailed
		status.ErrorMsg = err.Error()
		status.UpdatedAt = svc.timeNow()
		if updateErr := svc.taskStatusRepo.Update(context.Background(), taskID, status); updateErr != nil {
			fmt.Printf("Failed to update task status to failed: %v\n", updateErr)
		}
		return err
	}

	// Update progress 70%
	status.Progress = int32(70)
	status.UpdatedAt = svc.timeNow()
	if updateErr := svc.taskStatusRepo.Update(context.Background(), taskID, status); updateErr != nil {
		fmt.Printf("Failed to update task status progress: %v\n", updateErr)
	}

	dataDocs := svc.PrepareProductBarcodes(docHeader.LanguangeCode, docs)

	tempUnitCodes := []string{}
	for _, doc := range docs {
		tempUnitCodes = append(tempUnitCodes, doc.UnitCode)
	}

	unitDocs, err := svc.productUnitRepo.FindByUnitCodes(context.Background(), holdingCode, tempUnitCodes)
	if err != nil {
		status.Status = models.TaskStatusModelFailed
		status.ErrorMsg = err.Error()
		status.UpdatedAt = svc.timeNow()
		if updateErr := svc.taskStatusRepo.Update(context.Background(), taskID, status); updateErr != nil {
			fmt.Printf("Failed to update task status to failed: %v\n", updateErr)
		}
		return err
	}

	dataDocs = svc.PrepareProductUnit(unitDocs, dataDocs)

	// Update progress 80%
	status.Progress = int32(80)
	status.UpdatedAt = svc.timeNow()
	if updateErr := svc.taskStatusRepo.Update(context.Background(), taskID, status); updateErr != nil {
		fmt.Printf("Failed to update task status progress: %v\n", updateErr)
	}

	// Prepare master data
	dataDocs, err = svc.prepareMasterData(holdingCode, dataDocs)
	if err != nil {
		status.Status = models.TaskStatusModelFailed
		status.ErrorMsg = err.Error()
		status.UpdatedAt = svc.timeNow()
		if updateErr := svc.taskStatusRepo.Update(context.Background(), taskID, status); updateErr != nil {
			fmt.Printf("Failed to update task status to failed: %v\n", updateErr)
		}
		return err
	}

	// Update progress 90%
	status.Progress = int32(90)
	status.UpdatedAt = svc.timeNow()
	if updateErr := svc.taskStatusRepo.Update(context.Background(), taskID, status); updateErr != nil {
		fmt.Printf("Failed to update task status progress: %v\n", updateErr)
	}

	_, err = svc.productBarcodeService.SaveInBatch(holdingCode, authUsername, dataDocs)
	if err != nil {
		status.Status = models.TaskStatusModelFailed
		status.ErrorMsg = err.Error()
		status.UpdatedAt = svc.timeNow()
		if updateErr := svc.taskStatusRepo.Update(context.Background(), taskID, status); updateErr != nil {
			fmt.Printf("Failed to update task status to failed: %v\n", updateErr)
		}
		return err
	}

	// Update progress 95%
	status.Progress = int32(95)
	status.UpdatedAt = svc.timeNow()
	if updateErr := svc.taskStatusRepo.Update(context.Background(), taskID, status); updateErr != nil {
		fmt.Printf("Failed to update task status progress: %v\n", updateErr)
	}

	err = svc.DeleteTask(holdingCode, taskID)
	if err != nil {
		status.Status = models.TaskStatusModelFailed
		status.ErrorMsg = err.Error()
		status.UpdatedAt = svc.timeNow()
		if updateErr := svc.taskStatusRepo.Update(context.Background(), taskID, status); updateErr != nil {
			fmt.Printf("Failed to update task status to failed: %v\n", updateErr)
		}
		return err
	}

	// เช็คและอัพเดท refbarcodes ถ้ามี BarcodeRef
	refUpdateCount, err := svc.processRefBarcodeUpdates(context.Background(), holdingCode, authUsername, docs)
	if err != nil {
		fmt.Printf("Warning: Failed to update some refbarcodes: %v\n", err)
	} else if refUpdateCount > 0 {
		fmt.Printf("Updated refbarcodes for %d products\n", refUpdateCount)
	}

	// เมื่อสำเร็จ
	now := svc.timeNow()
	status.Status = models.TaskStatusModelCompleted
	status.Progress = int32(100)
	status.UpdatedAt = now
	status.CompletedAt = &now
	if updateErr := svc.taskStatusRepo.Update(context.Background(), taskID, status); updateErr != nil {
		fmt.Printf("Failed to update task status to completed: %v\n", updateErr)
	}

	// ลบ task status หลังจากสำเร็จ (อาจเก็บไว้สักพักหรือลบทันที)
	if deleteErr := svc.taskStatusRepo.Delete(context.Background(), holdingCode, taskID); deleteErr != nil {
		fmt.Printf("Failed to delete task status: %v\n", deleteErr)
	}

	return nil
}

func (svc ProductImportService) PrepareProductUnit(unitList []productunit_models.UnitInfo, docs []product_models.ProductBarcode) []product_models.ProductBarcode {

	unitDict := map[string]productunit_models.UnitInfo{}
	for _, unit := range unitList {
		unitDict[unit.UnitCode] = unit
	}

	for i := range docs {

		doc := docs[i]

		tempUnit, ok := unitDict[doc.ItemUnitCode]

		if !ok {
			continue
		}

		docs[i].ItemUnitNames = tempUnit.Names
	}

	return docs

}

// 🚀 Helper function สำหรับ pre-fetch master data แบบ batch เพื่อเพิ่มความเร็ว
type MasterDataCache struct {
	Groups       map[string]groupproduct_models.GroupProductInfo
	Groupsubones map[string]groupsuboneproduct_models.GroupsuboneProductInfo
	Groupsubtwos map[string]groupsubtwoproduct_models.GroupsubtwoProductInfo
	Brands       map[string]brandproduct_models.BrandProductInfo
	Designs      map[string]designproduct_models.DesignProductInfo
	Models       map[string]modelproduct_models.ModelProductInfo
	Patterns     map[string]patternproduct_models.PatternProductInfo
	Grades       map[string]gradeproduct_models.GradeProductInfo
	Categories   map[string]categoryproduct_models.CategoryProductInfo
	Classes      map[string]classproduct_models.ClassProductInfo
	Units        map[string]productunit_models.UnitInfo
}

func (svc ProductImportService) fetchMasterDataCache(ctx context.Context, holdingCode string, importDataList []models.ProductImportRaw) (*MasterDataCache, error) {
	cache := &MasterDataCache{
		Groups:       make(map[string]groupproduct_models.GroupProductInfo),
		Groupsubones: make(map[string]groupsuboneproduct_models.GroupsuboneProductInfo),
		Groupsubtwos: make(map[string]groupsubtwoproduct_models.GroupsubtwoProductInfo),
		Brands:       make(map[string]brandproduct_models.BrandProductInfo),
		Designs:      make(map[string]designproduct_models.DesignProductInfo),
		Models:       make(map[string]modelproduct_models.ModelProductInfo),
		Patterns:     make(map[string]patternproduct_models.PatternProductInfo),
		Grades:       make(map[string]gradeproduct_models.GradeProductInfo),
		Categories:   make(map[string]categoryproduct_models.CategoryProductInfo),
		Classes:      make(map[string]classproduct_models.ClassProductInfo),
		Units:        make(map[string]productunit_models.UnitInfo),
	}

	// 🔧 รวบรวม unique codes จาก batch
	groupCodes := make(map[string]struct{})
	groupsuboneCodes := make(map[string]struct{})
	groupsubtwoCodes := make(map[string]struct{})
	brandCodes := make(map[string]struct{})
	designCodes := make(map[string]struct{})
	modelCodes := make(map[string]struct{})
	patternCodes := make(map[string]struct{})
	gradeCodes := make(map[string]struct{})
	categoryCodes := make(map[string]struct{})
	classCodes := make(map[string]struct{})
	unitCodes := make(map[string]struct{})

	for _, data := range importDataList {
		if data.GroupCode != "" {
			groupCodes[data.GroupCode] = struct{}{}
		}
		if data.GroupsuboneCode != "" {
			groupsuboneCodes[data.GroupsuboneCode] = struct{}{}
		}
		if data.GroupsubtwoCode != "" {
			groupsubtwoCodes[data.GroupsubtwoCode] = struct{}{}
		}
		if data.BrandCode != "" {
			brandCodes[data.BrandCode] = struct{}{}
		}
		if data.DesignCode != "" {
			designCodes[data.DesignCode] = struct{}{}
		}
		if data.ModelCode != "" {
			modelCodes[data.ModelCode] = struct{}{}
		}
		if data.PatternCode != "" {
			patternCodes[data.PatternCode] = struct{}{}
		}
		if data.GradeCode != "" {
			gradeCodes[data.GradeCode] = struct{}{}
		}
		if data.CategoryCode != "" {
			categoryCodes[data.CategoryCode] = struct{}{}
		}
		if data.ClassCode != "" {
			classCodes[data.ClassCode] = struct{}{}
		}
		if data.UnitCode != "" {
			unitCodes[data.UnitCode] = struct{}{}
		}
	}

	// 🚀 Fetch ทั้งหมดพร้อมกัน (10 queries แทน N*10 queries)

	// Groups
	if len(groupCodes) > 0 {
		codes := make([]string, 0, len(groupCodes))
		for code := range groupCodes {
			codes = append(codes, code)
		}
		if groups, err := svc.groupProductRepo.FindByCodes(ctx, holdingCode, codes); err == nil {
			for _, g := range groups {
				cache.Groups[g.Code] = g
			}
		}
	}

	// Groupsubones
	if len(groupsuboneCodes) > 0 {
		codes := make([]string, 0, len(groupsuboneCodes))
		for code := range groupsuboneCodes {
			codes = append(codes, code)
		}
		if groupsubones, err := svc.groupsuboneProductRepo.FindByCodes(ctx, holdingCode, codes); err == nil {
			for _, g := range groupsubones {
				cache.Groupsubones[g.Code] = g
			}
		}
	}

	// Groupsubtwos
	if len(groupsubtwoCodes) > 0 {
		codes := make([]string, 0, len(groupsubtwoCodes))
		for code := range groupsubtwoCodes {
			codes = append(codes, code)
		}
		if groupsubtwos, err := svc.groupsubtwoproductRepo.FindByCodes(ctx, holdingCode, codes); err == nil {
			for _, g := range groupsubtwos {
				cache.Groupsubtwos[g.Code] = g
			}
		}
	}

	// Brands
	if len(brandCodes) > 0 {
		codes := make([]string, 0, len(brandCodes))
		for code := range brandCodes {
			codes = append(codes, code)
		}
		if brands, err := svc.brandProductRepo.FindByCodes(ctx, holdingCode, codes); err == nil {
			for _, b := range brands {
				cache.Brands[b.Code] = b
			}
		}
	}

	// Designs
	if len(designCodes) > 0 {
		codes := make([]string, 0, len(designCodes))
		for code := range designCodes {
			codes = append(codes, code)
		}
		if designs, err := svc.designProductRepo.FindByCodes(ctx, holdingCode, codes); err == nil {
			for _, d := range designs {
				cache.Designs[d.Code] = d
			}
		}
	}

	// Models
	if len(modelCodes) > 0 {
		codes := make([]string, 0, len(modelCodes))
		for code := range modelCodes {
			codes = append(codes, code)
		}
		if models, err := svc.modelProductRepo.FindByCodes(ctx, holdingCode, codes); err == nil {
			for _, m := range models {
				cache.Models[m.Code] = m
			}
		}
	}

	// Patterns
	if len(patternCodes) > 0 {
		codes := make([]string, 0, len(patternCodes))
		for code := range patternCodes {
			codes = append(codes, code)
		}
		if patterns, err := svc.patternProductRepo.FindByCodes(ctx, holdingCode, codes); err == nil {
			for _, p := range patterns {
				cache.Patterns[p.Code] = p
			}
		}
	}

	// Grades
	if len(gradeCodes) > 0 {
		codes := make([]string, 0, len(gradeCodes))
		for code := range gradeCodes {
			codes = append(codes, code)
		}
		if grades, err := svc.gradeProductRepo.FindByCodes(ctx, holdingCode, codes); err == nil {
			for _, g := range grades {
				cache.Grades[g.Code] = g
			}
		}
	}

	// Categories
	if len(categoryCodes) > 0 {
		codes := make([]string, 0, len(categoryCodes))
		for code := range categoryCodes {
			codes = append(codes, code)
		}
		if categories, err := svc.categoryProductRepo.FindByCodes(ctx, holdingCode, codes); err == nil {
			for _, c := range categories {
				cache.Categories[c.Code] = c
			}
		}
	}

	// Classes
	if len(classCodes) > 0 {
		codes := make([]string, 0, len(classCodes))
		for code := range classCodes {
			codes = append(codes, code)
		}
		if classes, err := svc.classProductRepo.FindByCodes(ctx, holdingCode, codes); err == nil {
			for _, c := range classes {
				cache.Classes[c.Code] = c
			}
		}
	}

	// Units
	if len(unitCodes) > 0 {
		codes := make([]string, 0, len(unitCodes))
		for code := range unitCodes {
			codes = append(codes, code)
		}
		if units, err := svc.productUnitRepo.FindByUnitCodes(ctx, holdingCode, codes); err == nil {
			for _, u := range units {
				cache.Units[u.UnitCode] = u
			}
		}
	}

	return cache, nil
}

// 🚀 Helper function สำหรับ apply master data จาก cache
func (svc ProductImportService) applyMasterDataFromCache(productBase *product_models.ProductBarcodeBase, importData models.ProductImportRaw, cache *MasterDataCache) error {
	var missingMasters []string // เก็บ master data ที่หายไป
	// Group
	if importData.GroupCode != "" {
		if group, ok := cache.Groups[importData.GroupCode]; ok {
			productBase.GroupGuid = group.GuidFixed
			productBase.GroupNames = group.Names
		} else {
			missingMasters = append(missingMasters, fmt.Sprintf("Group code '%s' not found", importData.GroupCode))
		}
	}
	// Groupsubone
	if importData.GroupsuboneCode != "" {
		if groupsubone, ok := cache.Groupsubones[importData.GroupsuboneCode]; ok {
			productBase.GroupsuboneGuid = groupsubone.GuidFixed
			productBase.GroupsuboneNames = groupsubone.Names
		} else {
			missingMasters = append(missingMasters, fmt.Sprintf("Groupsubone code '%s' not found", importData.GroupsuboneCode))
		}
	}

	// Groupsubtwo
	if importData.GroupsubtwoCode != "" {
		if groupsubtwo, ok := cache.Groupsubtwos[importData.GroupsubtwoCode]; ok {
			productBase.GroupsubtwoGuid = groupsubtwo.GuidFixed
			productBase.GroupsubtwoNames = groupsubtwo.Names
		} else {
			missingMasters = append(missingMasters, fmt.Sprintf("Groupsubtwo code '%s' not found", importData.GroupsubtwoCode))
		}
	}

	// Brand
	if importData.BrandCode != "" {
		if brand, ok := cache.Brands[importData.BrandCode]; ok {
			productBase.BrandGuid = brand.GuidFixed
			productBase.BrandNames = brand.Names
		} else {
			missingMasters = append(missingMasters, fmt.Sprintf("Brand code '%s' not found", importData.BrandCode))
		}
	}

	// Design
	if importData.DesignCode != "" {
		if design, ok := cache.Designs[importData.DesignCode]; ok {
			productBase.DesignGuid = design.GuidFixed
			productBase.DesignNames = design.Names
		} else {
			missingMasters = append(missingMasters, fmt.Sprintf("Design code '%s' not found", importData.DesignCode))
		}
	}

	// Model
	if importData.ModelCode != "" {
		if model, ok := cache.Models[importData.ModelCode]; ok {
			productBase.ModelGuid = model.GuidFixed
			productBase.ModelNames = model.Names
		} else {
			missingMasters = append(missingMasters, fmt.Sprintf("Model code '%s' not found", importData.ModelCode))
		}
	}

	// Pattern
	if importData.PatternCode != "" {
		if pattern, ok := cache.Patterns[importData.PatternCode]; ok {
			productBase.PatternGuid = pattern.GuidFixed
			productBase.PatternNames = pattern.Names
		} else {
			missingMasters = append(missingMasters, fmt.Sprintf("Pattern code '%s' not found", importData.PatternCode))
		}
	}

	// Grade
	if importData.GradeCode != "" {
		if grade, ok := cache.Grades[importData.GradeCode]; ok {
			productBase.GradeGuid = grade.GuidFixed
			productBase.GradeNames = grade.Names
		} else {
			missingMasters = append(missingMasters, fmt.Sprintf("Grade code '%s' not found", importData.GradeCode))
		}
	}

	// Category
	if importData.CategoryCode != "" {
		if category, ok := cache.Categories[importData.CategoryCode]; ok {
			productBase.CategoryGuid = category.GuidFixed
			productBase.CategoryNames = category.Names
		} else {
			missingMasters = append(missingMasters, fmt.Sprintf("Category code '%s' not found", importData.CategoryCode))
		}
	}

	// Class
	if importData.ClassCode != "" {
		if class, ok := cache.Classes[importData.ClassCode]; ok {
			productBase.ClassGuid = class.GuidFixed
			productBase.ClassNames = class.Names
		} else {
			missingMasters = append(missingMasters, fmt.Sprintf("Class code '%s' not found", importData.ClassCode))
		}
	}

	// Unit
	if importData.UnitCode != "" {
		if unit, ok := cache.Units[importData.UnitCode]; ok {
			productBase.ItemUnitGuid = unit.GuidFixed
			productBase.ItemUnitNames = unit.Names
		} else {
			missingMasters = append(missingMasters, fmt.Sprintf("Unit code '%s' not found", importData.UnitCode))
		}
	}

	// Return error ถ้ามี master data หายไป
	if len(missingMasters) > 0 {
		return fmt.Errorf("missing master data: %s", strings.Join(missingMasters, ", "))
	}

	return nil
}

// 🚀 Helper function สำหรับ apply master data updates จาก cache (สำหรับ updateExistingProduct)
func (svc ProductImportService) applyMasterDataUpdatesFromCache(updateData bson.M, importData models.ProductImportRaw, changes []models.FieldChange, cache *MasterDataCache) error {
	for _, change := range changes {
		switch change.Field {
		case "groupcode":
			if importData.GroupCode != "" {
				if group, ok := cache.Groups[importData.GroupCode]; ok {
					updateData["groupguid"] = group.GuidFixed
					updateData["groupnames"] = group.Names
				} else {
					return fmt.Errorf("group code '%s' not found in master data", importData.GroupCode)
				}
			} else {
				updateData["groupguid"] = ""
				updateData["groupnames"] = nil
			}

		case "groupsubonecode":
			if importData.GroupsuboneCode != "" {
				if groupsubone, ok := cache.Groupsubones[importData.GroupsuboneCode]; ok {
					updateData["groupsuboneguid"] = groupsubone.GuidFixed
					updateData["groupsubonenames"] = groupsubone.Names
				} else {
					return fmt.Errorf("groupsubone code '%s' not found in master data", importData.GroupsuboneCode)
				}
			} else {
				updateData["groupsuboneguid"] = ""
				updateData["groupsubonenames"] = nil
			}

		case "groupsubtwocode":
			if importData.GroupsubtwoCode != "" {
				if groupsubtwo, ok := cache.Groupsubtwos[importData.GroupsubtwoCode]; ok {
					updateData["groupsubtwoguid"] = groupsubtwo.GuidFixed
					updateData["groupsubtwonames"] = groupsubtwo.Names
				} else {
					return fmt.Errorf("groupsubtwo code '%s' not found in master data", importData.GroupsubtwoCode)
				}
			} else {
				updateData["groupsubtwoguid"] = ""
				updateData["groupsubtwonames"] = nil
			}

		case "brand_code":
			if importData.BrandCode != "" {
				if brand, ok := cache.Brands[importData.BrandCode]; ok {
					updateData["brandguid"] = brand.GuidFixed
					updateData["brandnames"] = brand.Names
				} else {
					return fmt.Errorf("brand code '%s' not found in master data", importData.BrandCode)
				}
			} else {
				updateData["brandguid"] = ""
				updateData["brandnames"] = nil
			}

		case "designcode":
			if importData.DesignCode != "" {
				if design, ok := cache.Designs[importData.DesignCode]; ok {
					updateData["designguid"] = design.GuidFixed
					updateData["designnames"] = design.Names
				} else {
					return fmt.Errorf("design code '%s' not found in master data", importData.DesignCode)
				}
			} else {
				updateData["designguid"] = ""
				updateData["designnames"] = nil
			}

		case "modelcode":
			if importData.ModelCode != "" {
				if model, ok := cache.Models[importData.ModelCode]; ok {
					updateData["modelguid"] = model.GuidFixed
					updateData["modelnames"] = model.Names
				} else {
					return fmt.Errorf("model code '%s' not found in master data", importData.ModelCode)
				}
			} else {
				updateData["modelguid"] = ""
				updateData["modelnames"] = nil
			}

		case "patterncode":
			if importData.PatternCode != "" {
				if pattern, ok := cache.Patterns[importData.PatternCode]; ok {
					updateData["patternguid"] = pattern.GuidFixed
					updateData["patternnames"] = pattern.Names
				} else {
					return fmt.Errorf("pattern code '%s' not found in master data", importData.PatternCode)
				}
			} else {
				updateData["patternguid"] = ""
				updateData["patternnames"] = nil
			}

		case "gradecode":
			if importData.GradeCode != "" {
				if grade, ok := cache.Grades[importData.GradeCode]; ok {
					updateData["gradeguid"] = grade.GuidFixed
					updateData["gradenames"] = grade.Names
				} else {
					return fmt.Errorf("grade code '%s' not found in master data", importData.GradeCode)
				}
			} else {
				updateData["gradeguid"] = ""
				updateData["gradenames"] = nil
			}

		case "categorycode":
			if importData.CategoryCode != "" {
				if category, ok := cache.Categories[importData.CategoryCode]; ok {
					updateData["categoryguid"] = category.GuidFixed
					updateData["categorynames"] = category.Names
				} else {
					return fmt.Errorf("category code '%s' not found in master data", importData.CategoryCode)
				}
			} else {
				updateData["categoryguid"] = ""
				updateData["categorynames"] = nil
			}

		case "classcode":
			if importData.ClassCode != "" {
				if class, ok := cache.Classes[importData.ClassCode]; ok {
					updateData["classguid"] = class.GuidFixed
					updateData["classnames"] = class.Names
				} else {
					return fmt.Errorf("class code '%s' not found in master data", importData.ClassCode)
				}
			} else {
				updateData["classguid"] = ""
				updateData["classnames"] = nil
			}

		case "unitcode":
			if importData.UnitCode != "" {
				if unit, ok := cache.Units[importData.UnitCode]; ok {
					updateData["itemunitguid"] = unit.GuidFixed
					updateData["itemunitnames"] = unit.Names
				} else {
					return fmt.Errorf("unit code '%s' not found in master data", importData.UnitCode)
				}
			} else {
				updateData["itemunitguid"] = ""
				updateData["itemunitnames"] = nil
			}
		}
	}
	return nil
}

func (svc ProductImportService) prepareMasterData(holdingCode string, docs []product_models.ProductBarcode) ([]product_models.ProductBarcode, error) {
	// Fetch business types from branch code "00000" once for all products
	branchBusinessTypes := []product_models.ProductBarcodeBusinessType{}
	branchDoc, err := svc.branchRepo.FindByDocIndentityGuid(context.Background(), holdingCode, "code", "00000")
	if err == nil && len(branchDoc.GuidFixed) > 0 {
		if branchDoc.Branch.BusinessTypes != nil {
			for _, businessTypeGUID := range *branchDoc.Branch.BusinessTypes {
				businessTypeDoc, err := svc.businessTypeRepo.FindByGuid(context.Background(), holdingCode, businessTypeGUID)
				if err == nil && len(businessTypeDoc.GuidFixed) > 0 {
					branchBusinessTypes = append(branchBusinessTypes, product_models.ProductBarcodeBusinessType{
						DocIdentity: businessTypeDoc.DocIdentity,
						Code:        businessTypeDoc.BusinessType.Code,
						Names:       businessTypeDoc.BusinessType.Names,
						IsIgnore:    false,
					})
				}
			}
		}
	}

	// Collect all unique codes
	groupCodes := make(map[string]struct{})
	groupsuboneCodes := make(map[string]struct{})
	groupsubtwoCodes := make(map[string]struct{})
	brandCodes := make(map[string]struct{})
	designCodes := make(map[string]struct{})
	modelCodes := make(map[string]struct{})
	patternCodes := make(map[string]struct{})
	gradeCodes := make(map[string]struct{})
	categoryCodes := make(map[string]struct{})
	classCodes := make(map[string]struct{})

	for _, doc := range docs {

		if doc.GroupCode != "" {
			groupCodes[doc.GroupCode] = struct{}{}
		}
		if doc.GroupsuboneCode != "" {
			groupsuboneCodes[doc.GroupsuboneCode] = struct{}{}
		}
		if doc.GroupsubtwoCode != "" {
			groupsubtwoCodes[doc.GroupsubtwoCode] = struct{}{}
		}
		if doc.BrandCode != "" {
			brandCodes[doc.BrandCode] = struct{}{}
		}
		if doc.DesignCode != "" {
			designCodes[doc.DesignCode] = struct{}{}
		}
		if doc.ModelCode != "" {
			modelCodes[doc.ModelCode] = struct{}{}
		}
		if doc.PatternCode != "" {
			patternCodes[doc.PatternCode] = struct{}{}
		}
		if doc.GradeCode != "" {
			gradeCodes[doc.GradeCode] = struct{}{}
		}
		if doc.CategoryCode != "" {
			categoryCodes[doc.CategoryCode] = struct{}{}
		}
		if doc.ClassCode != "" {
			classCodes[doc.ClassCode] = struct{}{}
		}
	}

	// Convert maps to slices
	groupCodesList := make([]string, 0, len(groupCodes))
	for code := range groupCodes {
		groupCodesList = append(groupCodesList, code)
	}

	groupsuboneCodesList := make([]string, 0, len(groupsuboneCodes))
	for code := range groupsuboneCodes {
		groupsuboneCodesList = append(groupsuboneCodesList, code)
	}

	groupsubtwoCodesList := make([]string, 0, len(groupsubtwoCodes))
	for code := range groupsubtwoCodes {
		groupsubtwoCodesList = append(groupsubtwoCodesList, code)
	}

	brandCodesList := make([]string, 0, len(brandCodes))
	for code := range brandCodes {
		brandCodesList = append(brandCodesList, code)
	}

	designCodesList := make([]string, 0, len(designCodes))
	for code := range designCodes {
		designCodesList = append(designCodesList, code)
	}

	modelCodesList := make([]string, 0, len(modelCodes))
	for code := range modelCodes {
		modelCodesList = append(modelCodesList, code)
	}

	patternCodesList := make([]string, 0, len(patternCodes))
	for code := range patternCodes {
		patternCodesList = append(patternCodesList, code)
	}

	gradeCodesList := make([]string, 0, len(gradeCodes))
	for code := range gradeCodes {
		gradeCodesList = append(gradeCodesList, code)
	}

	categoryCodesList := make([]string, 0, len(categoryCodes))
	for code := range categoryCodes {
		categoryCodesList = append(categoryCodesList, code)
	}

	classCodesList := make([]string, 0, len(classCodes))
	for code := range classCodes {
		classCodesList = append(classCodesList, code)
	}

	// Fetch master data
	groupDict := make(map[string]groupproduct_models.GroupProductInfo)
	if len(groupCodesList) > 0 {
		groupDocs, err := svc.groupProductRepo.FindByCodes(context.Background(), holdingCode, groupCodesList)
		if err == nil {
			for _, group := range groupDocs {
				groupDict[group.Code] = group
			}
		}
	}

	groupsuboneDict := make(map[string]groupsuboneproduct_models.GroupsuboneProductInfo)
	if len(groupsuboneCodesList) > 0 {
		groupsuboneDocs, err := svc.groupsuboneProductRepo.FindByCodes(context.Background(), holdingCode, groupsuboneCodesList)
		if err == nil {
			for _, groupsubone := range groupsuboneDocs {
				groupsuboneDict[groupsubone.Code] = groupsubone
			}
		}
	}

	groupsubtwoDict := make(map[string]groupsubtwoproduct_models.GroupsubtwoProductInfo)
	if len(groupsubtwoCodesList) > 0 {
		groupsubtwoDocs, err := svc.groupsubtwoproductRepo.FindByCodes(context.Background(), holdingCode, groupsubtwoCodesList)
		if err == nil {
			for _, groupsubtwo := range groupsubtwoDocs {
				groupsubtwoDict[groupsubtwo.Code] = groupsubtwo
			}
		}
	}

	brandDict := make(map[string]brandproduct_models.BrandProductInfo)
	if len(brandCodesList) > 0 {
		brandDocs, err := svc.brandProductRepo.FindByCodes(context.Background(), holdingCode, brandCodesList)
		if err == nil {
			for _, brand := range brandDocs {
				brandDict[brand.Code] = brand
			}
		}
	}

	designDict := make(map[string]designproduct_models.DesignProductInfo)
	if len(designCodesList) > 0 {
		designDocs, err := svc.designProductRepo.FindByCodes(context.Background(), holdingCode, designCodesList)
		if err == nil {
			for _, design := range designDocs {
				designDict[design.Code] = design
			}
		}
	}

	modelDict := make(map[string]modelproduct_models.ModelProductInfo)
	if len(modelCodesList) > 0 {
		modelDocs, err := svc.modelProductRepo.FindByCodes(context.Background(), holdingCode, modelCodesList)
		if err == nil {
			for _, model := range modelDocs {
				modelDict[model.Code] = model
			}
		}
	}

	patternDict := make(map[string]patternproduct_models.PatternProductInfo)
	if len(patternCodesList) > 0 {
		patternDocs, err := svc.patternProductRepo.FindByCodes(context.Background(), holdingCode, patternCodesList)
		if err == nil {
			for _, pattern := range patternDocs {
				patternDict[pattern.Code] = pattern
			}
		}
	}

	gradeDict := make(map[string]gradeproduct_models.GradeProductInfo)
	if len(gradeCodesList) > 0 {
		gradeDocs, err := svc.gradeProductRepo.FindByCodes(context.Background(), holdingCode, gradeCodesList)
		if err == nil {
			for _, grade := range gradeDocs {
				gradeDict[grade.Code] = grade
			}
		}
	}

	categoryDict := make(map[string]categoryproduct_models.CategoryProductInfo)
	if len(categoryCodesList) > 0 {
		categoryDocs, err := svc.categoryProductRepo.FindByCodes(context.Background(), holdingCode, categoryCodesList)
		if err == nil {
			for _, category := range categoryDocs {
				categoryDict[category.Code] = category
			}
		}
	}

	classDict := make(map[string]classproduct_models.ClassProductInfo)
	if len(classCodesList) > 0 {
		classDocs, err := svc.classProductRepo.FindByCodes(context.Background(), holdingCode, classCodesList)
		if err == nil {
			for _, class := range classDocs {
				classDict[class.Code] = class
			}
		}
	}

	// Apply master data to docs
	for i := range docs {
		// Set Group data
		if docs[i].GroupCode != "" {
			if group, exists := groupDict[docs[i].GroupCode]; exists {
				docs[i].GroupGuid = group.GuidFixed
				docs[i].GroupNames = group.Names
			}
		}

		// Set Groupsubone data
		if docs[i].GroupsuboneCode != "" {
			if groupsubone, exists := groupsuboneDict[docs[i].GroupsuboneCode]; exists {
				docs[i].GroupsuboneGuid = groupsubone.GuidFixed
				docs[i].GroupsuboneNames = groupsubone.Names
			}
		}

		// Set Groupsubtwo data
		if docs[i].GroupsubtwoCode != "" {
			if groupsubtwo, exists := groupsubtwoDict[docs[i].GroupsubtwoCode]; exists {
				docs[i].GroupsubtwoGuid = groupsubtwo.GuidFixed
				docs[i].GroupsubtwoNames = groupsubtwo.Names
			}
		}

		// Set Brand data
		if docs[i].BrandCode != "" {
			if brand, exists := brandDict[docs[i].BrandCode]; exists {
				docs[i].BrandGuid = brand.GuidFixed
				docs[i].BrandNames = brand.Names
			}
		}

		// Set Design data
		if docs[i].DesignCode != "" {
			if design, exists := designDict[docs[i].DesignCode]; exists {
				docs[i].DesignGuid = design.GuidFixed
				docs[i].DesignNames = design.Names
			}
		}

		// Set Model data
		if docs[i].ModelCode != "" {
			if model, exists := modelDict[docs[i].ModelCode]; exists {
				docs[i].ModelGuid = model.GuidFixed
				docs[i].ModelNames = model.Names
			}
		}

		// Set Pattern data
		if docs[i].PatternCode != "" {
			if pattern, exists := patternDict[docs[i].PatternCode]; exists {
				docs[i].PatternGuid = pattern.GuidFixed
				docs[i].PatternNames = pattern.Names
			}
		}

		// Set Grade data
		if docs[i].GradeCode != "" {
			if grade, exists := gradeDict[docs[i].GradeCode]; exists {
				docs[i].GradeGuid = grade.GuidFixed
				docs[i].GradeNames = grade.Names
			}
		}

		// Set Category data
		if docs[i].CategoryCode != "" {
			if category, exists := categoryDict[docs[i].CategoryCode]; exists {
				docs[i].CategoryGuid = category.GuidFixed
				docs[i].CategoryNames = category.Names
			}
		}

		// Set Class data
		if docs[i].ClassCode != "" {
			if class, exists := classDict[docs[i].ClassCode]; exists {
				docs[i].ClassGuid = class.GuidFixed
				docs[i].ClassNames = class.Names
			}
		}

		docs[i].IsMainBarcode = true

		docs[i].IgnoreBranches = &[]product_models.ProductBarcodeBranch{}
		docs[i].IsALaCarte = true
		//BusinessTypes
		docs[i].BusinessTypes = &branchBusinessTypes

	}

	return docs, nil
}

func (svc ProductImportService) PrepareProductBarcodes(langCode string, docs []models.ProductImportDoc) []product_models.ProductBarcode {

	dataDocs := []product_models.ProductBarcode{}

	for i := range docs {

		temp := product_models.ProductBarcode{}

		temp.Barcode = docs[i].Barcode
		temp.ItemCode = docs[i].Code
		temp.ItemUnitCode = docs[i].UnitCode

		// Set master data codes from import
		temp.GroupCode = docs[i].GroupCode
		temp.GroupsuboneCode = docs[i].GroupsuboneCode
		temp.GroupsubtwoCode = docs[i].GroupsubtwoCode
		temp.BrandCode = docs[i].BrandCode
		temp.DesignCode = docs[i].DesignCode
		temp.ModelCode = docs[i].ModelCode
		temp.PatternCode = docs[i].PatternCode
		temp.GradeCode = docs[i].GradeCode
		temp.CategoryCode = docs[i].CategoryCode
		temp.ClassCode = docs[i].ClassCode
		temp.IsSumPoint = docs[i].IsSumPoint

		productPrices := []product_models.ProductPrice{}

		productPrices = append(productPrices, product_models.ProductPrice{
			KeyNumber: 1,
			Price:     docs[i].Price,
		})

		productPrices = append(productPrices, product_models.ProductPrice{
			KeyNumber: 2,
			Price:     docs[i].PriceMember,
		})

		productPrices = append(productPrices, product_models.ProductPrice{
			KeyNumber: 3,
			Price:     docs[i].PriceDelivery,
		})

		//price 1-9
		productPrices = append(productPrices, product_models.ProductPrice{
			KeyNumber: 9,
			Price:     docs[i].PriceOne,
		})
		productPrices = append(productPrices, product_models.ProductPrice{
			KeyNumber: 10,
			Price:     docs[i].PriceTwo,
		})
		productPrices = append(productPrices, product_models.ProductPrice{
			KeyNumber: 11,
			Price:     docs[i].PriceThree,
		})

		//price 4-6
		productPrices = append(productPrices, product_models.ProductPrice{
			KeyNumber: 12,
			Price:     docs[i].PriceFour,
		})
		productPrices = append(productPrices, product_models.ProductPrice{
			KeyNumber: 13,
			Price:     docs[i].PriceFive,
		})
		productPrices = append(productPrices, product_models.ProductPrice{
			KeyNumber: 14,
			Price:     docs[i].PriceSix,
		})
		//price 7-9
		productPrices = append(productPrices, product_models.ProductPrice{
			KeyNumber: 15,
			Price:     docs[i].PriceSeven,
		})
		productPrices = append(productPrices, product_models.ProductPrice{
			KeyNumber: 16,
			Price:     docs[i].PriceEight,
		})
		productPrices = append(productPrices, product_models.ProductPrice{
			KeyNumber: 17,
			Price:     docs[i].PriceNine,
		})

		temp.Prices = &productPrices

		productNames := []common.NameX{}

		productNames = append(productNames, common.NameX{
			Code: &langCode,
			Name: &docs[i].Name,
		})

		temp.Names = &productNames

		dataDocs = append(dataDocs, temp)
	}

	return dataDocs
}

func (svc *ProductImportService) GetTaskStatus(holdingCode string, taskID string) (models.TaskStatusModel, error) {
	log.Printf("GetTaskStatus called for holdingCode: %s, taskID: %s", holdingCode, taskID)
	status, err := svc.taskStatusRepo.FindByTaskID(context.Background(), holdingCode, taskID)
	if err != nil {
		log.Printf("GetTaskStatus error: %v", err)
	} else {
		log.Printf("GetTaskStatus found: status=%s, progress=%d", status.Status, status.Progress)
	}
	return status, err
}

// processRefBarcodeUpdates processes refbarcode updates for products that have BarcodeRef
func (svc *ProductImportService) processRefBarcodeUpdates(ctx context.Context, holdingCode, authUsername string, docs []models.ProductImportDoc) (int, error) {
	// Extract products that need refbarcode processing
	var refRequests []string
	refMap := make(map[string]models.ProductImportDoc)
	codeToLookupMap := make(map[string]bool)                  // 🔧 ใช้ map เพื่อเก็บ unique ItemCode
	codeDocsMap := make(map[string][]models.ProductImportDoc) // 🔧 เก็บ docs ทั้งหมดที่มี ItemCode เดียวกัน

	for _, doc := range docs {
		// เงื่อนไขเดิม: StandValue > 0 และมี BarcodeRef
		if doc.StandValue > 0 && doc.BarcodeRef != "" {
			refRequests = append(refRequests, doc.Barcode, doc.BarcodeRef)
			refMap[doc.Barcode] = doc
		} else if doc.StandValue > 0 && doc.BarcodeRef == "" && doc.Code != "" {
			// เงื่อนไขใหม่: StandValue > 0, ไม่มี BarcodeRef แต่มี Code
			codeToLookupMap[doc.Code] = true
			codeDocsMap[doc.Code] = append(codeDocsMap[doc.Code], doc)
		}
	}

	// หาก barcode จาก ItemCode สำหรับกรณีที่ไม่มี BarcodeRef
	if len(codeToLookupMap) > 0 {
		// หา products จาก ItemCode แต่ละตัว (unique codes only)
		for itemCode := range codeToLookupMap {
			// ดึง docs ที่เกี่ยวข้องกับ ItemCode นี้
			docsForCode := codeDocsMap[itemCode]

			// 🔍 หา doc ที่เป็นตัวหลัก (StandValue = 1 และ DivideValue = 1)
			var mainDoc *models.ProductImportDoc
			for i := range docsForCode {
				if docsForCode[i].StandValue == 1 && docsForCode[i].DivideValue == 1 {
					mainDoc = &docsForCode[i]
					fmt.Printf("Found main barcode %s for ItemCode %s (StandValue=1, DivideValue=1)\n", mainDoc.Barcode, itemCode)
					break
				}
			}

			// ถ้าไม่พบตัวหลักใน import data ให้ลอง query จาก DB
			if mainDoc == nil {
				products, err := svc.productBarcodeRepo.FindByItemCode(ctx, holdingCode, itemCode)
				if err != nil {
					fmt.Printf("Warning: Failed to find product by ItemCode %s: %v\n", itemCode, err)
					continue
				}

				if len(products) > 0 {
					// ใช้ตัวแรกที่เจอจาก DB เป็น BarcodeRef
					eligibleProduct := products[0]
					fmt.Printf("Using existing product barcode %s for ItemCode %s (from DB)\n", eligibleProduct.Barcode, itemCode)

					// กำหนด BarcodeRef สำหรับ docs ที่มี StandValue > 0 และไม่ใช่ eligibleProduct
					for i := range docsForCode {
						if docsForCode[i].StandValue > 0 && docsForCode[i].Barcode != eligibleProduct.Barcode {
							docsForCode[i].BarcodeRef = eligibleProduct.Barcode
							refRequests = append(refRequests, docsForCode[i].Barcode, eligibleProduct.Barcode)
							refMap[docsForCode[i].Barcode] = docsForCode[i]
							fmt.Printf("Set BarcodeRef for %s -> %s (StandValue=%f, from DB)\n",
								docsForCode[i].Barcode, eligibleProduct.Barcode, docsForCode[i].StandValue)
						}
					}
				} else {
					fmt.Printf("Warning: No main product found for ItemCode %s (neither in import nor in DB)\n", itemCode)
				}
				continue
			}

			// ถ้าพบตัวหลักใน import data ให้ใช้ barcode ของมันเป็น BarcodeRef สำหรับตัวอื่นๆ
			mainBarcode := mainDoc.Barcode
			for i := range docsForCode {
				// กำหนด BarcodeRef สำหรับทุกตัวที่มี StandValue > 0 และไม่ใช่ตัวหลัก
				if docsForCode[i].StandValue > 0 && docsForCode[i].Barcode != mainBarcode {
					docsForCode[i].BarcodeRef = mainBarcode
					refRequests = append(refRequests, docsForCode[i].Barcode, mainBarcode)
					refMap[docsForCode[i].Barcode] = docsForCode[i]
					fmt.Printf("Set BarcodeRef for %s -> %s (StandValue=%f)\n",
						docsForCode[i].Barcode, mainBarcode, docsForCode[i].StandValue)
				}
			}
		}
	}

	if len(refRequests) == 0 {
		return 0, nil // No refbarcodes to update
	}

	// Find all relevant products by barcodes
	products, err := svc.productBarcodeRepo.FindByBarcodesMap(holdingCode, refRequests)
	if err != nil {
		return 0, fmt.Errorf("failed to find products: %v", err)
	}

	updateCount := 0

	// Process each refbarcode update
	for barcode, docData := range refMap {
		// Find target product
		targetProduct, targetExists := products[barcode]
		if !targetExists {
			fmt.Printf("Warning: Target barcode %s not found\n", barcode)
			continue
		}

		// Find reference product
		refProduct, refExists := products[docData.BarcodeRef]
		if !refExists {
			fmt.Printf("Warning: Reference barcode %s not found\n", docData.BarcodeRef)
			continue
		}

		// Determine condition based on StandValue and DivideValue
		condition := false
		if (docData.StandValue == 1 && docData.DivideValue == 1) || (docData.StandValue >= docData.DivideValue) {
			condition = false
		} else if docData.StandValue < docData.DivideValue {
			condition = true
		}

		// Create RefProductBarcode structure
		refBarcode := product_models.RefProductBarcode{
			GuidFixed:     refProduct.GuidFixed,
			Names:         refProduct.Names,
			ItemUnitCode:  refProduct.ItemUnitCode,
			ItemUnitNames: refProduct.ItemUnitNames,
			Barcode:       docData.BarcodeRef,
			Condition:     condition,
			DivideValue:   docData.DivideValue,
			StandValue:    docData.StandValue,
			Qty:           0,
		}

		// Update target product
		updateData := bson.M{
			"$set": bson.M{
				"refbarcodes":      []product_models.RefProductBarcode{refBarcode},
				"ismainbarcode":    false,
				"isusesubbarcodes": true,
				"updatedby":        authUsername,
				"updatedat":        time.Now(),
			},
		}

		err := svc.productBarcodeRepo.UpdateByID(targetProduct.GuidFixed, updateData)
		if err != nil {
			fmt.Printf("Warning: Failed to update refbarcode for %s: %v\n", barcode, err)
			continue
		}

		updateCount++
		fmt.Printf("Updated refbarcodes for %s -> %s (standvalue: %f, dividevalue: %f, condition: %t)\n",
			barcode, docData.BarcodeRef, docData.StandValue, docData.DivideValue, condition)
	}

	return updateCount, nil
}

// Enhanced methods สำหรับ compare และ insert/update functionality

func (svc ProductImportService) SaveTaskWithMode(holdingCode string, authUsername string, taskID string, docHeader models.ProductImportHeader, importMode string, forceUpdate bool) error {
	// First compare to get the analysis
	compareResult, err := svc.CompareWithExisting(holdingCode, taskID)
	if err != nil {
		return fmt.Errorf("failed to compare with existing data: %w", err)
	}

	ctx := context.Background()

	// Get all import data
	importDocList, err := svc.chRepo.All(ctx, holdingCode, taskID)
	if err != nil {
		return fmt.Errorf("failed to get import data: %w", err)
	}

	// Convert ProductImportDoc to ProductImportRaw
	var importDataList []models.ProductImportRaw
	for _, doc := range importDocList {
		importDataList = append(importDataList, doc.ProductImportRaw)
	}

	successInsert := 0
	successUpdate := 0
	failed := 0
	skipped := 0

	// startTime := time.Now()

	for _, importData := range importDataList {
		// Find corresponding compare result
		var compareItem *models.ProductBarcodeCompare
		for _, item := range compareResult.Items {
			if item.ImportData.Barcode == importData.Barcode {
				compareItem = &item
				break
			}
		}

		if compareItem == nil {
			continue
		}

		switch importMode {
		case "INSERT_ONLY":
			if compareItem.Status == "NEW" {
				err := svc.insertNewProduct(holdingCode, authUsername, importData, docHeader)
				if err != nil {
					failed++
					// Log error but continue processing
					log.Printf("Error inserting product %s: %v", importData.Barcode, err)
				} else {
					successInsert++
				}
			} else {
				skipped++
			}

		case "UPDATE_ONLY":
			if compareItem.Status == "UPDATE" || (compareItem.Status == "EXISTING" && forceUpdate) {
				err := svc.updateExistingProduct(holdingCode, authUsername, importData, compareItem, docHeader)
				if err != nil {
					failed++
					// Log error but continue processing
					log.Printf("Error updating product %s: %v", importData.Barcode, err)
				} else {
					successUpdate++
				}
			} else {
				skipped++
			}

		case "BOTH", "AUTO":
			if compareItem.Status == "NEW" {
				err := svc.insertNewProduct(holdingCode, authUsername, importData, docHeader)
				if err != nil {
					failed++
					// Log error but continue processing
					log.Printf("Error inserting product %s: %v", importData.Barcode, err)
				} else {
					successInsert++
				}
			} else if compareItem.Status == "UPDATE" || (compareItem.Status == "EXISTING" && forceUpdate) {
				err := svc.updateExistingProduct(holdingCode, authUsername, importData, compareItem, docHeader)
				if err != nil {
					failed++
					// Log error but continue processing
					log.Printf("Error updating product %s: %v", importData.Barcode, err)
				} else {
					successUpdate++
				}
			} else {
				skipped++
			}
		}
	}

	// endTime := time.Now()

	// TODO: Save import summary to database or cache for later retrieval
	// summary := models.ImportSummary{
	//     TaskID:        taskID,
	//     ImportMode:    importMode,
	//     StartTime:     startTime,
	//     EndTime:       endTime,
	//     Duration:      endTime.Sub(startTime).String(),
	//     TotalRecords:  len(importDataList),
	//     SuccessInsert: successInsert,
	//     SuccessUpdate: successUpdate,
	//     Failed:        failed,
	//     Skipped:       skipped,
	//     Errors:        errors,
	//     CreatedBy:     authUsername,
	//     Status:        determineImportStatus(failed, len(importDataList)),
	// }

	if failed > 0 && failed == len(importDataList) {
		return fmt.Errorf("all records failed to import")
	}

	// เช็คและอัพเดท refbarcodes ถ้ามี BarcodeRef หลังจากทำการ import เสร็จ
	refUpdateCount, err := svc.processRefBarcodeUpdates(context.Background(), holdingCode, authUsername, importDocList)
	if err != nil {
		log.Printf("Warning: Failed to update some refbarcodes: %v\n", err)
	} else if refUpdateCount > 0 {
		log.Printf("Updated refbarcodes for %d products\n", refUpdateCount)
	}

	// ลบ task หลังจากทำการ import เสร็จ (เหมือน SaveTask เดิม)
	err = svc.DeleteTask(holdingCode, taskID)
	if err != nil {
		log.Printf("Warning: Failed to delete task %s: %v\n", taskID, err)
		// ไม่ return error เพราะการ import สำเร็จแล้ว
	}

	return nil
}

func (svc ProductImportService) PreviewSave(holdingCode string, taskID string, docHeader models.ProductImportHeader, importMode string) (*models.CompareResult, error) {
	// This method returns a preview of what would happen without actually saving
	compareResult, err := svc.CompareWithExisting(holdingCode, taskID)
	if err != nil {
		return nil, err
	}

	// Adjust the compare result based on import mode
	switch importMode {
	case "INSERT_ONLY":
		// Filter to show only new records
		filteredItems := []models.ProductBarcodeCompare{}
		for _, item := range compareResult.Items {
			if item.Status == "NEW" {
				filteredItems = append(filteredItems, item)
			}
		}
		compareResult.Items = filteredItems
		compareResult.NewRecords = len(filteredItems)
		compareResult.ExistingRecords = 0
		compareResult.UpdatedRecords = 0

	case "UPDATE_ONLY":
		// Filter to show only existing/update records
		filteredItems := []models.ProductBarcodeCompare{}
		for _, item := range compareResult.Items {
			if item.Status == "UPDATE" || item.Status == "EXISTING" {
				filteredItems = append(filteredItems, item)
			}
		}
		compareResult.Items = filteredItems
		compareResult.NewRecords = 0
		compareResult.UpdatedRecords = len(filteredItems)
	}

	return compareResult, nil
}

func (svc ProductImportService) CompareWithExisting(holdingCode string, taskID string) (*models.CompareResult, error) {
	ctx := context.Background()

	// Get all import data for this task
	importDocList, err := svc.chRepo.All(ctx, holdingCode, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get import data: %w", err)
	}

	// Convert ProductImportDoc to ProductImportRaw
	var importDataList []models.ProductImportRaw
	var barcodes []string
	importDataMap := make(map[string]models.ProductImportRaw)

	for _, doc := range importDocList {
		importDataList = append(importDataList, doc.ProductImportRaw)
		barcodes = append(barcodes, doc.Barcode)
		importDataMap[doc.Barcode] = doc.ProductImportRaw
	}

	result := &models.CompareResult{
		TaskID:       taskID,
		TotalRecords: len(importDataList),
		Items:        []models.ProductBarcodeCompare{},
		CreatedAt:    time.Now(),
	}

	newRecords := 0
	existingRecords := 0
	updatedRecords := 0

	// 🚀 เปลี่ยนจาก FindByBarcode ทีละตัว เป็น FindByBarcodesMap ครั้งเดียว
	existingProductsMap, err := svc.productBarcodeRepo.FindByBarcodesMap(holdingCode, barcodes)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing products: %w", err)
	}

	// 🚀 ประมวลผลแบบ batch
	for _, importData := range importDataList {
		existingProduct, exists := existingProductsMap[importData.Barcode]

		compareItem := models.ProductBarcodeCompare{
			ImportData: importData,
			Changes:    []models.FieldChange{},
			Conflicts:  []models.FieldConflict{},
			CanUpdate:  true,
		}

		if !exists || existingProduct.GuidFixed == "" {
			// New product
			compareItem.Status = "NEW"
			compareItem.ExistingData = nil
			newRecords++
		} else {
			// Existing product - ใช้ fast comparison
			existingData := &models.ProductBarcodeExisting{
				Barcode:       existingProduct.Barcode,
				Code:          existingProduct.ItemCode,
				Name:          getLocalizedName(existingProduct.Names),
				UnitCode:      existingProduct.ItemUnitCode,
				Price:         getPrice(existingProduct.Prices, 1), // Primary price
				PriceMember:   getPrice(existingProduct.Prices, 2), // Member price
				PriceDelivery: getPrice(existingProduct.Prices, 3), // Delivery price

				// Extended price fields
				PriceOne:   getPrice(existingProduct.Prices, 9),  // PriceOne
				PriceTwo:   getPrice(existingProduct.Prices, 10), // PriceTwo
				PriceThree: getPrice(existingProduct.Prices, 11), // PriceThree
				PriceFour:  getPrice(existingProduct.Prices, 12), // PriceFour
				PriceFive:  getPrice(existingProduct.Prices, 13), // PriceFive
				PriceSix:   getPrice(existingProduct.Prices, 14), // PriceSix
				PriceSeven: getPrice(existingProduct.Prices, 15), // PriceSeven
				PriceEight: getPrice(existingProduct.Prices, 16), // PriceEight
				PriceNine:  getPrice(existingProduct.Prices, 17), // PriceNine

				// Master data codes
				GroupCode:       existingProduct.GroupCode,
				GroupsuboneCode: existingProduct.GroupsuboneCode,
				GroupsubtwoCode: existingProduct.GroupsubtwoCode,
				BrandCode:       existingProduct.BrandCode,
				DesignCode:      existingProduct.DesignCode,
				ModelCode:       existingProduct.ModelCode,
				PatternCode:     existingProduct.PatternCode,
				GradeCode:       existingProduct.GradeCode,
				CategoryCode:    existingProduct.CategoryCode,
				ClassCode:       existingProduct.ClassCode,

				// Reference barcode information - need to get from RefBarcodes
				BarcodeRef:  getRefBarcode(existingProduct.RefBarcodes),
				StandValue:  getRefStandValue(existingProduct.RefBarcodes),
				DivideValue: getRefDivideValue(existingProduct.RefBarcodes),

				IsSumPoint: existingProduct.IsSumPoint, // 🆕 เพิ่มบรรทัดนี้

				LastModified: time.Now(), // ใช้ current time
				ModifiedBy:   "system",   // ใช้ default value
			}
			compareItem.ExistingData = existingData

			// 🚀 Fast comparison - เช็คเฉพาะ field สำคัญก่อน
			hasChanges := svc.fastCompareProduct(importData, existingData)

			if hasChanges {
				// 🔧 ถ้ามีการเปลี่ยนแปลง ค่อยทำ detailed comparison
				compareItem.Changes = svc.getDetailedChanges(importData, existingData)
				compareItem.Status = "UPDATE"
				compareItem.UpdateReason = fmt.Sprintf("Found %d field changes", len(compareItem.Changes))
				compareItem.CanUpdate = true
				updatedRecords++
			} else {
				compareItem.Status = "EXISTING"
				existingRecords++
			}
		}

		result.Items = append(result.Items, compareItem)
	}

	result.NewRecords = newRecords
	result.ExistingRecords = existingRecords
	result.UpdatedRecords = updatedRecords
	result.ConflictRecords = 0 // ไม่มี conflicts

	// Create summary
	result.Summary = svc.createCompareSummary(result.Items, updatedRecords)

	return result, nil
}

// 🚀 เพิ่มฟังก์ชัน fast comparison
func (svc ProductImportService) fastCompareProduct(importData models.ProductImportRaw, existingData *models.ProductBarcodeExisting) bool {
	// เช็คทุกฟิลด์เพื่อให้แน่ใจว่าไม่พลาดการเปลี่ยนแปลงใดๆ

	// Basic fields
	if existingData.Name != importData.Name {
		return true
	}
	if existingData.Code != importData.Code {
		return true
	}
	if existingData.UnitCode != importData.UnitCode {
		return true
	}

	// All price fields
	if existingData.Price != importData.Price {
		return true
	}
	if existingData.PriceMember != importData.PriceMember {
		return true
	}
	if existingData.PriceDelivery != importData.PriceDelivery {
		return true
	}
	if existingData.PriceOne != importData.PriceOne {
		return true
	}
	if existingData.PriceTwo != importData.PriceTwo {
		return true
	}
	if existingData.PriceThree != importData.PriceThree {
		return true
	}
	if existingData.PriceFour != importData.PriceFour {
		return true
	}
	if existingData.PriceFive != importData.PriceFive {
		return true
	}
	if existingData.PriceSix != importData.PriceSix {
		return true
	}
	if existingData.PriceSeven != importData.PriceSeven {
		return true
	}
	if existingData.PriceEight != importData.PriceEight {
		return true
	}
	if existingData.PriceNine != importData.PriceNine {
		return true
	}
	if existingData.IsSumPoint != importData.IsSumPoint {
		return true
	}

	// Master data codes
	if existingData.GroupCode != importData.GroupCode {
		return true
	}
	if existingData.GroupsuboneCode != importData.GroupsuboneCode {
		return true
	}
	if existingData.GroupsubtwoCode != importData.GroupsubtwoCode {
		return true
	}
	if existingData.BrandCode != importData.BrandCode {
		return true
	}
	if existingData.DesignCode != importData.DesignCode {
		return true
	}
	if existingData.ModelCode != importData.ModelCode {
		return true
	}
	if existingData.PatternCode != importData.PatternCode {
		return true
	}
	if existingData.GradeCode != importData.GradeCode {
		return true
	}
	if existingData.CategoryCode != importData.CategoryCode {
		return true
	}
	if existingData.ClassCode != importData.ClassCode {
		return true
	}

	return false
}

// 🔧 เพิ่มฟังก์ชัน detailed comparison สำหรับกรณีที่มีการเปลี่ยนแปลง
func (svc ProductImportService) getDetailedChanges(importData models.ProductImportRaw, existingData *models.ProductBarcodeExisting) []models.FieldChange {
	var changes []models.FieldChange

	// Compare each field systematically
	fieldComparisons := []struct {
		fieldName string
		oldValue  interface{}
		newValue  interface{}
		reason    string
	}{
		{"name", existingData.Name, importData.Name, "Product name update"},
		{"code", existingData.Code, importData.Code, "Product code update"},
		{"unitcode", existingData.UnitCode, importData.UnitCode, "Unit code update"},
		{"price", existingData.Price, importData.Price, "Primary price update"},
		{"pricemember", existingData.PriceMember, importData.PriceMember, "Member price update"},
		{"pricedelivery", existingData.PriceDelivery, importData.PriceDelivery, "Delivery price update"},
		{"priceone", existingData.PriceOne, importData.PriceOne, "Price 1 update"},
		{"pricetwo", existingData.PriceTwo, importData.PriceTwo, "Price 2 update"},
		{"pricethree", existingData.PriceThree, importData.PriceThree, "Price 3 update"},
		{"pricefour", existingData.PriceFour, importData.PriceFour, "Price 4 update"},
		{"pricefive", existingData.PriceFive, importData.PriceFive, "Price 5 update"},
		{"pricesix", existingData.PriceSix, importData.PriceSix, "Price 6 update"},
		{"priceseven", existingData.PriceSeven, importData.PriceSeven, "Price 7 update"},
		{"priceeight", existingData.PriceEight, importData.PriceEight, "Price 8 update"},
		{"pricenine", existingData.PriceNine, importData.PriceNine, "Price 9 update"},
		{"groupcode", existingData.GroupCode, importData.GroupCode, "Group code update"},
		{"groupsubonecode", existingData.GroupsuboneCode, importData.GroupsuboneCode, "Group sub one code update"},
		{"groupsubtwocode", existingData.GroupsubtwoCode, importData.GroupsubtwoCode, "Group sub two code update"},
		{"brand_code", existingData.BrandCode, importData.BrandCode, "Brand code update"},
		{"designcode", existingData.DesignCode, importData.DesignCode, "Design code update"},
		{"modelcode", existingData.ModelCode, importData.ModelCode, "Model code update"},
		{"patterncode", existingData.PatternCode, importData.PatternCode, "Pattern code update"},
		{"gradecode", existingData.GradeCode, importData.GradeCode, "Grade code update"},
		{"categorycode", existingData.CategoryCode, importData.CategoryCode, "Category code update"},
		{"classcode", existingData.ClassCode, importData.ClassCode, "Class code update"},
		{"issumpoint", existingData.IsSumPoint, importData.IsSumPoint, "Sum point flag update"},
	}

	for _, comp := range fieldComparisons {
		if !isEqualValue(comp.oldValue, comp.newValue) {
			changes = append(changes, models.FieldChange{
				Field:    comp.fieldName,
				OldValue: comp.oldValue,
				NewValue: comp.newValue,
				Reason:   comp.reason,
			})
		}
	}

	return changes
}

// 🔧 เพิ่มฟังก์ชัน create summary
func (svc ProductImportService) createCompareSummary(items []models.ProductBarcodeCompare, updatedRecords int) *models.CompareSummary {
	priceChanges := 0
	nameChanges := 0
	unitChanges := 0
	masterDataChanges := 0
	refBarcodeChanges := 0

	for _, item := range items {
		for _, change := range item.Changes {
			switch change.Field {
			case "price", "pricemember", "pricedelivery", "priceone", "pricetwo", "pricethree", "pricefour", "pricefive", "pricesix", "priceseven", "priceeight", "pricenine":
				priceChanges++
			case "name":
				nameChanges++
			case "unitcode":
				unitChanges++
			case "groupcode", "groupsubonecode", "groupsubtwocode", "brand_code", "designcode", "modelcode", "patterncode", "gradecode", "categorycode", "classcode":
				masterDataChanges++
			case "barcoderef":
				refBarcodeChanges++
			}
		}
	}

	return &models.CompareSummary{
		ImportMode:        "AUTO",
		TotalChanges:      updatedRecords,
		PriceChanges:      priceChanges,
		NameChanges:       nameChanges,
		UnitChanges:       unitChanges,
		MasterDataChanges: masterDataChanges,
		RefBarcodeChanges: refBarcodeChanges,
		TotalConflicts:    0, // ไม่มี conflicts
		HighConflicts:     0,
		MediumConflicts:   0,
		LowConflicts:      0,
		EstimatedTime:     estimateImportTime(len(items)),
		CreatedAt:         time.Now(),
	}
}

func (svc ProductImportService) GetCompareResult(holdingCode string, taskID string) (*models.CompareResult, error) {
	// 🔧 เพิ่ม simple caching mechanism
	// cacheKey := fmt.Sprintf("compare_result:%s:%s", holdingCode, taskID)

	// TODO: ถ้ามี cache service สามารถเพิ่มได้
	// if cached := svc.cache.Get(cacheKey); cached != nil {
	//     return cached.(*models.CompareResult), nil
	// }

	result, err := svc.CompareWithExisting(holdingCode, taskID)
	if err != nil {
		return nil, err
	}

	// TODO: cache result for 5 minutes
	// svc.cache.Set(cacheKey, result, 5*time.Minute)

	return result, nil
}

// 🚀 เพิ่มฟังก์ชันใหม่สำหรับ pagination ที่แท้จริง
func (svc ProductImportService) GetCompareResultPaginated(holdingCode string, taskID string, pageable micromodels.Pageable) (*models.CompareResult, models.PaginationData, error) {
	ctx := context.Background()

	// 📋 Log input parameters
	log.Printf("[GetCompareResultPaginated] START - HoldingCode: %s, TaskID: %s, Page: %d, Limit: %d",
		holdingCode, taskID, pageable.Page, pageable.Limit)

	// Get paginated import data for this task
	importDocList, paginationData, err := svc.chRepo.List(ctx, holdingCode, taskID, pageable)
	if err != nil {
		log.Printf("[GetCompareResultPaginated] ERROR - Failed to get import data: %v", err)
		return nil, models.PaginationData{}, fmt.Errorf("failed to get import data: %w", err)
	}

	// 📊 Log pagination data
	log.Printf("[GetCompareResultPaginated] Retrieved %d records from page %d (Total: %d)",
		len(importDocList), paginationData.Page, paginationData.Total)

	// Convert ProductImportDoc to ProductImportRaw
	var importDataList []models.ProductImportRaw
	var barcodes []string
	importDataMap := make(map[string]models.ProductImportRaw)

	for _, doc := range importDocList {
		log.Printf("doc.ProductImportRaw: %v", doc.ProductImportRaw)
		importDataList = append(importDataList, doc.ProductImportRaw)
		barcodes = append(barcodes, doc.Barcode)
		importDataMap[doc.Barcode] = doc.ProductImportRaw
	}

	// 📋 Log barcodes being processed
	log.Printf("[GetCompareResultPaginated] Processing %d barcodes: %v", len(barcodes), barcodes)

	result := &models.CompareResult{
		TaskID:       taskID,
		TotalRecords: int(paginationData.Total),
		Items:        []models.ProductBarcodeCompare{},
		CreatedAt:    time.Now(),
	}

	newRecords := 0
	existingRecords := 0
	updatedRecords := 0

	// 🚀 เปลี่ยนจาก FindByBarcode ทีละตัว เป็น FindByBarcodesMap ครั้งเดียว
	existingProductsMap, err := svc.productBarcodeRepo.FindByBarcodesMap(holdingCode, barcodes)
	if err != nil {
		log.Printf("[GetCompareResultPaginated] ERROR - Failed to find existing products: %v", err)
		return nil, models.PaginationData{}, fmt.Errorf("failed to find existing products: %w", err)
	}

	// 📊 Log existing products found
	log.Printf("[GetCompareResultPaginated] Found %d existing products in database", len(existingProductsMap))

	// 🚀 ประมวลผลแบบ batch
	for _, importData := range importDataList {
		existingProduct, exists := existingProductsMap[importData.Barcode]

		// 🔍 Log each barcode processing
		log.Printf("[GetCompareResultPaginated] Processing barcode: %s, exists: %v", importData.Barcode, exists)

		compareItem := models.ProductBarcodeCompare{
			ImportData: importData,
			Changes:    []models.FieldChange{},
			Conflicts:  []models.FieldConflict{},
			CanUpdate:  true,
		}

		if !exists || existingProduct.GuidFixed == "" {
			// New product
			compareItem.Status = "NEW"
			compareItem.ExistingData = nil
			newRecords++
			log.Printf("[GetCompareResultPaginated] Barcode %s marked as NEW", importData.Barcode)
		} else {
			// Existing product - ใช้ fast comparison
			existingData := &models.ProductBarcodeExisting{
				Barcode:       existingProduct.Barcode,
				Code:          existingProduct.ItemCode,
				Name:          getLocalizedName(existingProduct.Names),
				UnitCode:      existingProduct.ItemUnitCode,
				Price:         getPrice(existingProduct.Prices, 1), // Primary price
				PriceMember:   getPrice(existingProduct.Prices, 2), // Member price
				PriceDelivery: getPrice(existingProduct.Prices, 3), // Delivery price

				// Extended price fields
				PriceOne:   getPrice(existingProduct.Prices, 9),  // PriceOne
				PriceTwo:   getPrice(existingProduct.Prices, 10), // PriceTwo
				PriceThree: getPrice(existingProduct.Prices, 11), // PriceThree
				PriceFour:  getPrice(existingProduct.Prices, 12), // PriceFour
				PriceFive:  getPrice(existingProduct.Prices, 13), // PriceFive
				PriceSix:   getPrice(existingProduct.Prices, 14), // PriceSix
				PriceSeven: getPrice(existingProduct.Prices, 15), // PriceSeven
				PriceEight: getPrice(existingProduct.Prices, 16), // PriceEight
				PriceNine:  getPrice(existingProduct.Prices, 17), // PriceNine

				// Master data codes
				GroupCode:       existingProduct.GroupCode,
				GroupsuboneCode: existingProduct.GroupsuboneCode,
				GroupsubtwoCode: existingProduct.GroupsubtwoCode,
				BrandCode:       existingProduct.BrandCode,
				DesignCode:      existingProduct.DesignCode,
				ModelCode:       existingProduct.ModelCode,
				PatternCode:     existingProduct.PatternCode,
				GradeCode:       existingProduct.GradeCode,
				CategoryCode:    existingProduct.CategoryCode,
				ClassCode:       existingProduct.ClassCode,
				IsSumPoint:      existingProduct.IsSumPoint, // 🆕 เพิ่มบรรทัดนี้

				// Reference barcode information - need to get from RefBarcodes
				BarcodeRef:  getRefBarcode(existingProduct.RefBarcodes),
				StandValue:  getRefStandValue(existingProduct.RefBarcodes),
				DivideValue: getRefDivideValue(existingProduct.RefBarcodes),

				LastModified: time.Now(), // ใช้ current time
				ModifiedBy:   "system",   // ใช้ default value
			}
			compareItem.ExistingData = existingData

			// 🚀 Fast comparison - เช็คเฉพาะ field สำคัญก่อน
			hasChanges := svc.fastCompareProduct(importData, existingData)

			// 🔍 Log comparison result
			log.Printf("[GetCompareResultPaginated] Barcode %s comparison: hasChanges=%v", importData.Barcode, hasChanges)

			if hasChanges {
				// 🔧 ถ้ามีการเปลี่ยนแปลง ค่อยทำ detailed comparison
				compareItem.Changes = svc.getDetailedChanges(importData, existingData)
				compareItem.Status = "UPDATE"
				compareItem.UpdateReason = fmt.Sprintf("Found %d field changes", len(compareItem.Changes))
				compareItem.CanUpdate = true
				updatedRecords++
				log.Printf("[GetCompareResultPaginated] Barcode %s marked as UPDATE with %d changes",
					importData.Barcode, len(compareItem.Changes))
			} else {
				compareItem.Status = "EXISTING"
				existingRecords++
				log.Printf("[GetCompareResultPaginated] Barcode %s marked as EXISTING (no changes)", importData.Barcode)
			}
		}

		result.Items = append(result.Items, compareItem)
	}

	result.NewRecords = newRecords
	result.ExistingRecords = existingRecords
	result.UpdatedRecords = updatedRecords
	result.ConflictRecords = 0 // ไม่มี conflicts

	// 📊 Log final summary
	log.Printf("[GetCompareResultPaginated] SUMMARY - Total: %d, NEW: %d, EXISTING: %d, UPDATE: %d",
		len(result.Items), newRecords, existingRecords, updatedRecords)

	// Create summary for current page
	result.Summary = svc.createCompareSummary(result.Items, updatedRecords)

	log.Printf("[GetCompareResultPaginated] END - Returning %d items", len(result.Items))
	return result, paginationData, nil
}

// 🚀 เพิ่มฟังก์ชันสำหรับดึง summary เฉพาะ (เร็วมาก)
func (svc ProductImportService) GetCompareResultSummary(holdingCode string, taskID string) (*models.CompareSummary, error) {
	ctx := context.Background()

	// Get count only - very fast
	importDocList, err := svc.chRepo.All(ctx, holdingCode, taskID)
	if err != nil {
		return nil, fmt.Errorf("failed to get import data: %w", err)
	}

	totalRecords := len(importDocList)
	var barcodes []string

	for _, doc := range importDocList {
		barcodes = append(barcodes, doc.Barcode)
	}

	// 🚀 Quick check for existing products
	existingProductsMap, err := svc.productBarcodeRepo.FindByBarcodesMap(holdingCode, barcodes)
	if err != nil {
		return nil, fmt.Errorf("failed to find existing products: %w", err)
	}

	newRecords := 0
	existingRecords := 0
	for _, doc := range importDocList {
		if _, exists := existingProductsMap[doc.Barcode]; exists {
			existingRecords++
		} else {
			newRecords++
		}
	}

	// สำหรับ summary mode ไม่ต้องทำ detailed comparison
	// แค่ประมาณการว่าจะมี update กี่ record
	estimatedUpdates := existingRecords / 2 // ประมาณการว่า 50% จะมีการเปลี่ยนแปลง

	summary := &models.CompareSummary{
		ImportMode:        "AUTO",
		TotalChanges:      estimatedUpdates,
		PriceChanges:      estimatedUpdates / 4,  // ประมาณการ 25% เป็น price changes
		NameChanges:       estimatedUpdates / 10, // ประมาณการ 10% เป็น name changes
		UnitChanges:       estimatedUpdates / 20, // ประมาณการ 5% เป็น unit changes
		MasterDataChanges: estimatedUpdates / 5,  // ประมาณการ 20% เป็น master data changes
		RefBarcodeChanges: estimatedUpdates / 50, // ประมาณการ 2% เป็น refbarcode changes
		TotalConflicts:    0,
		HighConflicts:     0,
		MediumConflicts:   0,
		LowConflicts:      0,
		EstimatedTime:     estimateImportTime(totalRecords),
		CreatedAt:         time.Now(),
	}

	return summary, nil
}

func (svc ProductImportService) PreviewChanges(holdingCode string, taskID string, importMode string) (*models.PreviewResult, error) {
	compareResult, err := svc.CompareWithExisting(holdingCode, taskID)
	if err != nil {
		return nil, err
	}

	preview := &models.PreviewResult{
		TaskID:          taskID,
		ImportMode:      importMode,
		WillInsert:      []models.ProductImportRaw{},
		WillUpdate:      []models.ProductBarcodeCompare{},
		WillSkip:        []models.ProductBarcodeCompare{},
		Warnings:        []models.PreviewWarning{},
		RequiredActions: []string{},
		CreatedAt:       time.Now(),
	}

	for _, item := range compareResult.Items {
		switch importMode {
		case "INSERT_ONLY":
			if item.Status == "NEW" {
				preview.WillInsert = append(preview.WillInsert, item.ImportData)
			} else {
				preview.WillSkip = append(preview.WillSkip, item)
			}

		case "UPDATE_ONLY":
			if item.Status == "UPDATE" || item.Status == "EXISTING" {
				preview.WillUpdate = append(preview.WillUpdate, item)
			} else {
				preview.WillSkip = append(preview.WillSkip, item)
			}

		case "BOTH", "AUTO":
			if item.Status == "NEW" {
				preview.WillInsert = append(preview.WillInsert, item.ImportData)
			} else if item.Status == "UPDATE" {
				preview.WillUpdate = append(preview.WillUpdate, item)
			} else {
				// ไม่มี CONFLICT - ทุกสิ่งที่ไม่ใช่ NEW หรือ UPDATE จะถูก skip
				preview.WillSkip = append(preview.WillSkip, item)
			}
		}
	}

	// Add required actions
	if len(preview.WillSkip) > 0 {
		preview.RequiredActions = append(preview.RequiredActions, "Review skipped items")
	}
	// ไม่มี conflicts warnings แล้ว

	preview.EstimatedTime = estimateImportTime(len(preview.WillInsert) + len(preview.WillUpdate))

	return preview, nil
}

func (svc ProductImportService) ApplyChanges(holdingCode string, authUsername string, taskID string, importMode string, forceUpdate bool) error {
	// 🔧 เปลี่ยนให้เรียก ApplyChangesWithProgress แทน SaveTaskWithMode
	log.Printf("ApplyChanges called for task: %s, mode: %s", taskID, importMode)
	return svc.ApplyChangesWithProgress(holdingCode, authUsername, taskID, importMode, forceUpdate)
}

func (svc ProductImportService) GetImportSummary(holdingCode string, taskID string) (*models.ImportSummary, error) {
	// TODO: Implement retrieval from database/cache where summary was stored during import
	// For now, return a placeholder
	return &models.ImportSummary{
		TaskID:    taskID,
		Status:    "NOT_FOUND",
		CreatedBy: "",
	}, nil
}

// Helper functions

func (svc ProductImportService) insertNewProduct(holdingCode string, authUsername string, importData models.ProductImportRaw, docHeader models.ProductImportHeader, masterCache ...*MasterDataCache) error {
	// สร้าง ProductBarcodeRequest โดยใช้ field จาก ProductBarcodeBase
	productBase := product_models.ProductBarcodeBase{
		ItemCode:         importData.Code,
		Barcode:          importData.Barcode,
		GroupCode:        importData.GroupCode,
		GroupsuboneCode:  importData.GroupsuboneCode,
		GroupsubtwoCode:  importData.GroupsubtwoCode,
		BrandCode:        importData.BrandCode,
		DesignCode:       importData.DesignCode,
		ModelCode:        importData.ModelCode,
		PatternCode:      importData.PatternCode,
		GradeCode:        importData.GradeCode,
		CategoryCode:     importData.CategoryCode,
		ClassCode:        importData.ClassCode,
		ItemUnitCode:     importData.UnitCode,
		IsMainBarcode:    true,
		IsSumPoint:       importData.IsSumPoint,
		IsALaCarte:       true,
		IsUseSubBarcodes: false,
	}

	// � ใช้ cache ถ้ามี ไม่เช่นนั้นค่อย fetch แบบเดิม
	if len(masterCache) > 0 && masterCache[0] != nil {
		// ⚡ ใช้ cache (เร็วมาก - O(1) lookup)
		err := svc.applyMasterDataFromCache(&productBase, importData, masterCache[0])
		if err != nil {
			// ✅ Return error โดยไม่ซ้ำ barcode (เพราะ caller จะใส่เอง)
			return err
		}
	} else {
		// 🐌 Fallback แบบเดิม (ช้า - ต้อง query DB 10 ครั้ง)
		ctx := context.Background()

		// Group
		if importData.GroupCode != "" {
			if groups, err := svc.groupProductRepo.FindByCodes(ctx, holdingCode, []string{importData.GroupCode}); err == nil && len(groups) > 0 {
				productBase.GroupGuid = groups[0].GuidFixed
				productBase.GroupNames = groups[0].Names
			}
		}

		// Groupsubone
		if importData.GroupsuboneCode != "" {
			if groupsubones, err := svc.groupsuboneProductRepo.FindByCodes(ctx, holdingCode, []string{importData.GroupsuboneCode}); err == nil && len(groupsubones) > 0 {
				productBase.GroupsuboneGuid = groupsubones[0].GuidFixed
				productBase.GroupsuboneNames = groupsubones[0].Names
			}
		}

		// Groupsubtwo
		if importData.GroupsubtwoCode != "" {
			if groupsubtwos, err := svc.groupsubtwoproductRepo.FindByCodes(ctx, holdingCode, []string{importData.GroupsubtwoCode}); err == nil && len(groupsubtwos) > 0 {
				productBase.GroupsubtwoGuid = groupsubtwos[0].GuidFixed
				productBase.GroupsubtwoNames = groupsubtwos[0].Names
			}
		}

		// Brand
		if importData.BrandCode != "" {
			if brands, err := svc.brandProductRepo.FindByCodes(ctx, holdingCode, []string{importData.BrandCode}); err == nil && len(brands) > 0 {
				productBase.BrandGuid = brands[0].GuidFixed
				productBase.BrandNames = brands[0].Names
			}
		}

		// Design
		if importData.DesignCode != "" {
			if designs, err := svc.designProductRepo.FindByCodes(ctx, holdingCode, []string{importData.DesignCode}); err == nil && len(designs) > 0 {
				productBase.DesignGuid = designs[0].GuidFixed
				productBase.DesignNames = designs[0].Names
			}
		}

		// Model
		if importData.ModelCode != "" {
			if models, err := svc.modelProductRepo.FindByCodes(ctx, holdingCode, []string{importData.ModelCode}); err == nil && len(models) > 0 {
				productBase.ModelGuid = models[0].GuidFixed
				productBase.ModelNames = models[0].Names
			}
		}

		// Pattern
		if importData.PatternCode != "" {
			if patterns, err := svc.patternProductRepo.FindByCodes(ctx, holdingCode, []string{importData.PatternCode}); err == nil && len(patterns) > 0 {
				productBase.PatternGuid = patterns[0].GuidFixed
				productBase.PatternNames = patterns[0].Names
			}
		}

		// Grade
		if importData.GradeCode != "" {
			if grades, err := svc.gradeProductRepo.FindByCodes(ctx, holdingCode, []string{importData.GradeCode}); err == nil && len(grades) > 0 {
				productBase.GradeGuid = grades[0].GuidFixed
				productBase.GradeNames = grades[0].Names
			}
		}

		// Category
		if importData.CategoryCode != "" {
			if categories, err := svc.categoryProductRepo.FindByCodes(ctx, holdingCode, []string{importData.CategoryCode}); err == nil && len(categories) > 0 {
				productBase.CategoryGuid = categories[0].GuidFixed
				productBase.CategoryNames = categories[0].Names
			}
		}

		// Class
		if importData.ClassCode != "" {
			if classes, err := svc.classProductRepo.FindByCodes(ctx, holdingCode, []string{importData.ClassCode}); err == nil && len(classes) > 0 {
				productBase.ClassGuid = classes[0].GuidFixed
				productBase.ClassNames = classes[0].Names
			}
		}
	}

	// Set names with default language code "th"
	languageCode := "th"
	names := []common.NameX{
		{
			Code: &languageCode,
			Name: &importData.Name,
		},
	}
	productBase.Names = &names

	// Set all prices including extended price fields
	prices := []product_models.ProductPrice{
		{KeyNumber: 1, Price: importData.Price},         // Primary price
		{KeyNumber: 2, Price: importData.PriceMember},   // Member price
		{KeyNumber: 3, Price: importData.PriceDelivery}, // Delivery price
		{KeyNumber: 9, Price: importData.PriceOne},      // Extended prices
		{KeyNumber: 10, Price: importData.PriceTwo},
		{KeyNumber: 11, Price: importData.PriceThree},
		{KeyNumber: 12, Price: importData.PriceFour},
		{KeyNumber: 13, Price: importData.PriceFive},
		{KeyNumber: 14, Price: importData.PriceSix},
		{KeyNumber: 15, Price: importData.PriceSeven},
		{KeyNumber: 16, Price: importData.PriceEight},
		{KeyNumber: 17, Price: importData.PriceNine},
	}
	productBase.Prices = &prices

	// Handle reference barcode if exists
	var refBarcodes []product_models.BarcodeRequest
	if importData.BarcodeRef != "" && importData.StandValue > 0 {
		refBarcodes = append(refBarcodes, product_models.BarcodeRequest{
			Barcode:     importData.BarcodeRef,
			StandValue:  importData.StandValue,
			DivideValue: importData.DivideValue,
		})
		productBase.IsUseSubBarcodes = true
		productBase.IsMainBarcode = false
	}

	productReq := product_models.ProductBarcodeRequest{
		ProductBarcodeBase: productBase,
		RefBarcodes:        refBarcodes,
		BOM:                []product_models.BOMRequest{},
		IgnoreBranches:     []product_models.ProductBarcodeBranch{},
		BusinessTypes:      []product_models.ProductBarcodeBusinessType{},
	}

	// เรียกใช้ CreateProductBarcode service
	_, err := svc.productBarcodeService.CreateProductBarcode(holdingCode, authUsername, productReq)
	return err
}

func (svc ProductImportService) updateExistingProduct(holdingCode string, authUsername string, importData models.ProductImportRaw, compareItem *models.ProductBarcodeCompare, docHeader models.ProductImportHeader, masterCache ...*MasterDataCache) error {
	// Update existing product with new data
	existingProduct, err := svc.productBarcodeRepo.FindByBarcode(context.Background(), holdingCode, importData.Barcode)
	if err != nil {
		return err
	}

	// สร้างข้อมูลที่จะ update
	updateData := bson.M{
		"updatedby": authUsername,
		"updatedat": time.Now(),
	}

	// จัดการ prices array แยกต่างหาก
	var newPrices []product_models.ProductPrice
	if existingProduct.Prices != nil {
		newPrices = *existingProduct.Prices
	}

	// 🚀 เช็คว่ามี cache หรือไม่
	hasCache := len(masterCache) > 0 && masterCache[0] != nil

	// อัปเดตข้อมูลตามการเปลี่ยนแปลง
	for _, change := range compareItem.Changes {
		switch change.Field {
		case "name":
			// Update localized names with default language code "th"
			languageCode := "th"
			updateData["names"] = createLocalizedNames(importData.Name, languageCode)
		case "issumpoint": // 🆕 เพิ่ม case สำหรับ IsSumPoint
			updateData["issumpoint"] = importData.IsSumPoint

		case "code":
			updateData["itemcode"] = importData.Code

		case "unitcode":
			updateData["itemunitcode"] = importData.UnitCode
			// อาจต้องอัปเดต unit names ด้วย
			if importData.UnitCode != "" {
				unitDocs, err := svc.productUnitRepo.FindByUnitCodes(context.Background(), holdingCode, []string{importData.UnitCode})
				if err == nil && len(unitDocs) > 0 {
					updateData["itemunitnames"] = unitDocs[0].Names
				}
			}

		case "price":
			// Update primary price (KeyNumber = 1)
			updated := false
			for i, price := range newPrices {
				if price.KeyNumber == 1 {
					newPrices[i].Price = importData.Price
					updated = true
					break
				}
			}
			if !updated {
				newPrices = append(newPrices, product_models.ProductPrice{
					KeyNumber: 1,
					Price:     importData.Price,
				})
			}

		case "pricemember":
			// Update member price (KeyNumber = 2)
			updated := false
			for i, price := range newPrices {
				if price.KeyNumber == 2 {
					newPrices[i].Price = importData.PriceMember
					updated = true
					break
				}
			}
			if !updated {
				newPrices = append(newPrices, product_models.ProductPrice{
					KeyNumber: 2,
					Price:     importData.PriceMember,
				})
			}

		case "pricedelivery":
			// Update delivery price (KeyNumber = 3)
			updated := false
			for i, price := range newPrices {
				if price.KeyNumber == 3 {
					newPrices[i].Price = importData.PriceDelivery
					updated = true
					break
				}
			}
			if !updated {
				newPrices = append(newPrices, product_models.ProductPrice{
					KeyNumber: 3,
					Price:     importData.PriceDelivery,
				})
			}

		// Extended price fields (PriceOne to PriceNine)
		case "priceone":
			updatePriceInArray(&newPrices, 9, importData.PriceOne)
		case "pricetwo":
			updatePriceInArray(&newPrices, 10, importData.PriceTwo)
		case "pricethree":
			updatePriceInArray(&newPrices, 11, importData.PriceThree)
		case "pricefour":
			updatePriceInArray(&newPrices, 12, importData.PriceFour)
		case "pricefive":
			updatePriceInArray(&newPrices, 13, importData.PriceFive)
		case "pricesix":
			updatePriceInArray(&newPrices, 14, importData.PriceSix)
		case "priceseven":
			updatePriceInArray(&newPrices, 15, importData.PriceSeven)
		case "priceeight":
			updatePriceInArray(&newPrices, 16, importData.PriceEight)
		case "pricenine":
			updatePriceInArray(&newPrices, 17, importData.PriceNine)

		// Master data codes
		case "groupcode":
			updateData["groupcode"] = importData.GroupCode
		case "groupsubonecode":
			updateData["groupsubonecode"] = importData.GroupsuboneCode
		case "groupsubtwocode":
			updateData["groupsubtwocode"] = importData.GroupsubtwoCode
		case "brand_code":
			updateData["brandcode"] = importData.BrandCode
		case "designcode":
			updateData["designcode"] = importData.DesignCode
		case "modelcode":
			updateData["modelcode"] = importData.ModelCode
		case "patterncode":
			updateData["patterncode"] = importData.PatternCode
		case "gradecode":
			updateData["gradecode"] = importData.GradeCode
		case "categorycode":
			updateData["categorycode"] = importData.CategoryCode
		case "classcode":
			updateData["classcode"] = importData.ClassCode

		// Reference barcode fields
		case "barcoderef":
			// Handle RefBarcodes update - always use import data values for standvalue and dividevalue
			if importData.BarcodeRef != "" && importData.StandValue > 0 {
				// Create or update RefBarcodes
				refBarcodes := []product_models.RefProductBarcode{
					{
						Barcode:     importData.BarcodeRef,
						StandValue:  importData.StandValue,
						DivideValue: importData.DivideValue,
						Condition:   importData.StandValue < importData.DivideValue,
						Qty:         0,
					},
				}
				updateData["refbarcodes"] = refBarcodes
				updateData["isusesubbarcodes"] = true
				updateData["ismainbarcode"] = false
			} else {
				// Clear RefBarcodes if no reference data
				updateData["refbarcodes"] = []product_models.RefProductBarcode{}
				updateData["isusesubbarcodes"] = false
				updateData["ismainbarcode"] = true
			}
		}
	}

	// 🚀 Apply master data from cache (ถ้ามี)
	if hasCache {
		err = svc.applyMasterDataUpdatesFromCache(updateData, importData, compareItem.Changes, masterCache[0])
		if err != nil {
			// ✅ Return error โดยไม่ซ้ำ barcode (เพราะ caller จะใส่เอง)
			return err
		}
	} else {
		// � Fallback แบบเดิม (ช้า - query DB สำหรับแต่ละ master data code)
		ctx := context.Background()
		for _, change := range compareItem.Changes {
			switch change.Field {
			case "groupcode":
				if importData.GroupCode != "" {
					if groups, err := svc.groupProductRepo.FindByCodes(ctx, holdingCode, []string{importData.GroupCode}); err == nil && len(groups) > 0 {
						updateData["groupguid"] = groups[0].GuidFixed
						updateData["groupnames"] = groups[0].Names
					}
				} else {
					updateData["groupguid"] = ""
					updateData["groupnames"] = nil
				}
			case "groupsubonecode":
				if importData.GroupsuboneCode != "" {
					if groupsubones, err := svc.groupsuboneProductRepo.FindByCodes(ctx, holdingCode, []string{importData.GroupsuboneCode}); err == nil && len(groupsubones) > 0 {
						updateData["groupsuboneguid"] = groupsubones[0].GuidFixed
						updateData["groupsubonenames"] = groupsubones[0].Names
					}
				} else {
					updateData["groupsuboneguid"] = ""
					updateData["groupsubonenames"] = nil
				}
			case "groupsubtwocode":
				if importData.GroupsubtwoCode != "" {
					if groupsubtwos, err := svc.groupsubtwoproductRepo.FindByCodes(ctx, holdingCode, []string{importData.GroupsubtwoCode}); err == nil && len(groupsubtwos) > 0 {
						updateData["groupsubtwoguid"] = groupsubtwos[0].GuidFixed
						updateData["groupsubtwonames"] = groupsubtwos[0].Names
					}
				} else {
					updateData["groupsubtwoguid"] = ""
					updateData["groupsubtwonames"] = nil
				}
			case "brand_code":
				if importData.BrandCode != "" {
					if brands, err := svc.brandProductRepo.FindByCodes(ctx, holdingCode, []string{importData.BrandCode}); err == nil && len(brands) > 0 {
						updateData["brandguid"] = brands[0].GuidFixed
						updateData["brandnames"] = brands[0].Names
					}
				} else {
					updateData["brandguid"] = ""
					updateData["brandnames"] = nil
				}
			case "designcode":
				if importData.DesignCode != "" {
					if designs, err := svc.designProductRepo.FindByCodes(ctx, holdingCode, []string{importData.DesignCode}); err == nil && len(designs) > 0 {
						updateData["designguid"] = designs[0].GuidFixed
						updateData["designnames"] = designs[0].Names
					}
				} else {
					updateData["designguid"] = ""
					updateData["designnames"] = nil
				}
			case "modelcode":
				if importData.ModelCode != "" {
					if models, err := svc.modelProductRepo.FindByCodes(ctx, holdingCode, []string{importData.ModelCode}); err == nil && len(models) > 0 {
						updateData["modelguid"] = models[0].GuidFixed
						updateData["modelnames"] = models[0].Names
					}
				} else {
					updateData["modelguid"] = ""
					updateData["modelnames"] = nil
				}
			case "patterncode":
				if importData.PatternCode != "" {
					if patterns, err := svc.patternProductRepo.FindByCodes(ctx, holdingCode, []string{importData.PatternCode}); err == nil && len(patterns) > 0 {
						updateData["patternguid"] = patterns[0].GuidFixed
						updateData["patternnames"] = patterns[0].Names
					}
				} else {
					updateData["patternguid"] = ""
					updateData["patternnames"] = nil
				}
			case "gradecode":
				if importData.GradeCode != "" {
					if grades, err := svc.gradeProductRepo.FindByCodes(ctx, holdingCode, []string{importData.GradeCode}); err == nil && len(grades) > 0 {
						updateData["gradeguid"] = grades[0].GuidFixed
						updateData["gradenames"] = grades[0].Names
					}
				} else {
					updateData["gradeguid"] = ""
					updateData["gradenames"] = nil
				}
			case "categorycode":
				if importData.CategoryCode != "" {
					if categories, err := svc.categoryProductRepo.FindByCodes(ctx, holdingCode, []string{importData.CategoryCode}); err == nil && len(categories) > 0 {
						updateData["categoryguid"] = categories[0].GuidFixed
						updateData["categorynames"] = categories[0].Names
					}
				} else {
					updateData["categoryguid"] = ""
					updateData["categorynames"] = nil
				}
			case "classcode":
				if importData.ClassCode != "" {
					if classes, err := svc.classProductRepo.FindByCodes(ctx, holdingCode, []string{importData.ClassCode}); err == nil && len(classes) > 0 {
						updateData["classguid"] = classes[0].GuidFixed
						updateData["classnames"] = classes[0].Names
					}
				} else {
					updateData["classguid"] = ""
					updateData["classnames"] = nil
				}
			}
		}
	}

	// เพิ่ม prices array ใน updateData
	updateData["prices"] = newPrices

	return svc.productBarcodeRepo.UpdateByID(existingProduct.GuidFixed, bson.M{"$set": updateData})
}

func getLocalizedName(names *[]common.NameX) string {
	if names == nil || len(*names) == 0 {
		return ""
	}
	if (*names)[0].Name != nil {
		return *(*names)[0].Name
	}
	return ""
}

func getPrice(prices *[]product_models.ProductPrice, keyNumber int) float64 {
	if prices == nil || len(*prices) == 0 {
		return 0
	}
	for _, price := range *prices {
		if price.KeyNumber == keyNumber {
			return price.Price
		}
	}
	return 0
}

func createLocalizedNames(name string, languageCode string) *[]common.NameX {
	names := []common.NameX{
		{
			Code: &languageCode,
			Name: &name,
		},
	}
	return &names
}

func estimateImportTime(recordCount int) string {
	// Rough estimation: 100 records per second
	seconds := recordCount / 100
	if seconds < 1 {
		return "< 1 second"
	} else if seconds < 60 {
		return fmt.Sprintf("%d seconds", seconds)
	} else {
		minutes := seconds / 60
		return fmt.Sprintf("%d minutes", minutes)
	}
}

// Helper functions for RefBarcode data extraction
func getRefBarcode(refBarcodes *[]product_models.RefProductBarcode) string {
	if refBarcodes == nil || len(*refBarcodes) == 0 {
		return ""
	}
	return (*refBarcodes)[0].Barcode
}

func getRefStandValue(refBarcodes *[]product_models.RefProductBarcode) float64 {
	if refBarcodes == nil || len(*refBarcodes) == 0 {
		return 0
	}
	return (*refBarcodes)[0].StandValue
}

func getRefDivideValue(refBarcodes *[]product_models.RefProductBarcode) float64 {
	if refBarcodes == nil || len(*refBarcodes) == 0 {
		return 0
	}
	return (*refBarcodes)[0].DivideValue
}

// Helper function to compare values with proper type handling
func isEqualValue(oldValue, newValue interface{}) bool {
	// Handle string comparison
	if oldStr, ok := oldValue.(string); ok {
		if newStr, ok := newValue.(string); ok {
			return oldStr == newStr
		}
	}

	// Handle float64 comparison with tolerance for floating point precision
	if oldFloat, ok := oldValue.(float64); ok {
		if newFloat, ok := newValue.(float64); ok {
			return fmt.Sprintf("%.6f", oldFloat) == fmt.Sprintf("%.6f", newFloat)
		}
	}

	// Handle int comparison
	if oldInt, ok := oldValue.(int); ok {
		if newInt, ok := newValue.(int); ok {
			return oldInt == newInt
		}
	}

	// Default comparison
	return oldValue == newValue
}

// Helper function to update price in prices array
func updatePriceInArray(prices *[]product_models.ProductPrice, keyNumber int, newPrice float64) {
	updated := false
	for i, price := range *prices {
		if price.KeyNumber == keyNumber {
			(*prices)[i].Price = newPrice
			updated = true
			break
		}
	}
	if !updated {
		*prices = append(*prices, product_models.ProductPrice{
			KeyNumber: keyNumber,
			Price:     newPrice,
		})
	}
}

// 🔧 เพิ่ม methods สำหรับ task status management และ progress tracking

func (svc ProductImportService) CreateTaskStatus(holdingCode string, taskID string, status string, progress int, message string) error {
	// 🔧 ใช้ time.Now().UTC() แทน svc.timeNow() เพื่อป้องกัน datetime overflow
	now := time.Now().UTC()

	taskStatus := models.TaskStatusModel{
		TaskID:      taskID,
		HoldingCode: holdingCode,
		Status:      status,
		Progress:    int32(progress),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	err := svc.taskStatusRepo.Create(context.Background(), taskStatus)
	if err != nil {
		log.Printf("CreateTaskStatus error: %v", err)
	} else {
		log.Printf("CreateTaskStatus success: taskID=%s", taskID)
	}
	return err
}

func (svc ProductImportService) UpdateTaskStatus(holdingCode string, taskID string, status string, progress int, message string) error {
	// 🔧 ใช้ time.Now().UTC() แทน svc.timeNow() เพื่อป้องกัน datetime overflow
	now := time.Now().UTC()

	taskStatus := models.TaskStatusModel{
		TaskID:      taskID,
		HoldingCode: holdingCode,
		Status:      status,
		Progress:    int32(progress),
		UpdatedAt:   now,
	}

	// ✅ เซ็ต ErrorMsg ถ้าเป็น error status หรือมี message
	if status == "ERROR" || status == "FAILED" || status == "COMPLETED_WITH_ERRORS" {
		taskStatus.ErrorMsg = message
	}

	err := svc.taskStatusRepo.Update(context.Background(), taskID, taskStatus)
	if err != nil {
		log.Printf("UpdateTaskStatus error: %v", err)
	} else {
		log.Printf("UpdateTaskStatus success: taskID=%s", taskID)
	}
	return err
}

func (svc ProductImportService) ApplyChangesWithProgress(holdingCode string, authUsername string, taskID string, importMode string, forceUpdate bool, batchSize ...int) error {
	ctx := context.Background()

	// 🔧 กำหนด batch size (default = 100)
	bSize := 100
	if len(batchSize) > 0 && batchSize[0] > 0 && batchSize[0] <= 1000 {
		bSize = batchSize[0]
	}
	log.Printf("[ApplyChangesWithProgress] Using batch size: %d", bSize)

	// 🔧 เช็คว่ามี task status อยู่แล้วหรือไม่
	_, err := svc.GetTaskStatus(holdingCode, taskID)
	if err != nil {
		// ไม่มี task status ให้สร้างใหม่
		log.Printf("[ApplyChangesWithProgress] No existing task status found, creating new one for task: %s", taskID)
		err = svc.CreateTaskStatus(holdingCode, taskID, "APPLYING", 0, "Starting import process...")
		if err != nil {
			log.Printf("[ApplyChangesWithProgress] Failed to create task status: %v", err)
			return fmt.Errorf("failed to create task status: %w", err)
		}
	} else {
		// มี task status แล้ว ให้ update
		log.Printf("[ApplyChangesWithProgress] Found existing task status, updating for task: %s", taskID)
		_ = svc.UpdateTaskStatus(holdingCode, taskID, "APPLYING", 0, "Starting import process...")
	}

	// อัพเดต status เป็น processing
	_ = svc.UpdateTaskStatus(holdingCode, taskID, "PROCESSING", 5, "Getting import data...")
	log.Printf("[ApplyChangesWithProgress] Updated task status to PROCESSING for task: %s", taskID)

	// Get all import data for this task
	importDocList, err := svc.chRepo.All(ctx, holdingCode, taskID)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to get import data: %v", err)
		_ = svc.UpdateTaskStatus(holdingCode, taskID, "ERROR", 5, errorMsg)
		log.Printf("[ApplyChangesWithProgress] %s", errorMsg)
		return fmt.Errorf("failed to get import data: %w", err)
	}

	totalRecords := len(importDocList)
	log.Printf("[ApplyChangesWithProgress] Retrieved %d total records", totalRecords)

	if totalRecords == 0 {
		_ = svc.UpdateTaskStatus(holdingCode, taskID, "COMPLETED", 100, "No records to process")
		log.Printf("[ApplyChangesWithProgress] No records to process, completed")
		return nil
	}

	_ = svc.UpdateTaskStatus(holdingCode, taskID, "PROCESSING", 10, fmt.Sprintf("Processing %d records with batch size %d...", totalRecords, bSize))

	// Convert to raw data and get barcodes
	var importDataList []models.ProductImportRaw
	var barcodes []string

	for _, doc := range importDocList {
		importDataList = append(importDataList, doc.ProductImportRaw)
		barcodes = append(barcodes, doc.Barcode)
	}

	_ = svc.UpdateTaskStatus(holdingCode, taskID, "PROCESSING", 20, "Checking existing products...")

	// Check existing products
	existingProductsMap, err := svc.productBarcodeRepo.FindByBarcodesMap(holdingCode, barcodes)
	if err != nil {
		errorMsg := fmt.Sprintf("Failed to check existing products: %v", err)
		_ = svc.UpdateTaskStatus(holdingCode, taskID, "ERROR", 20, errorMsg)
		log.Printf("[ApplyChangesWithProgress] %s", errorMsg)
		return fmt.Errorf("failed to find existing products: %w", err)
	}

	// Categorize records
	var newRecords []models.ProductImportRaw
	var updateRecords []models.ProductImportRaw

	for _, importData := range importDataList {
		if _, exists := existingProductsMap[importData.Barcode]; exists {
			updateRecords = append(updateRecords, importData)
		} else {
			newRecords = append(newRecords, importData)
		}
	}

	_ = svc.UpdateTaskStatus(holdingCode, taskID, "PROCESSING", 30, fmt.Sprintf("Found %d new, %d existing products", len(newRecords), len(updateRecords)))
	log.Printf("[ApplyChangesWithProgress] Categorized: %d new, %d existing products", len(newRecords), len(updateRecords))

	processedCount := 0
	errorCount := 0
	var errorMessages []string // 🆕 เก็บ error messages
	batchCount := 0

	// Process based on import mode
	switch importMode {
	case "INSERT_ONLY":
		totalInsert := len(newRecords)
		log.Printf("[ApplyChangesWithProgress] Processing INSERT_ONLY mode - %d records to insert with batch size %d", totalInsert, bSize)

		// Process in batches
		for i := 0; i < totalInsert; i += bSize {
			batchCount++
			end := i + bSize
			if end > totalInsert {
				end = totalInsert
			}
			batch := newRecords[i:end]
			batchSize := len(batch)

			log.Printf("[ApplyChangesWithProgress] Processing batch #%d: records %d-%d (%d records)", batchCount, i+1, end, batchSize)

			// 🚀 Pre-fetch master data cache for this batch (เพิ่ม performance)
			masterCache, err := svc.fetchMasterDataCache(ctx, holdingCode, batch)
			if err != nil {
				log.Printf("[ApplyChangesWithProgress] Warning: Failed to fetch master data cache for batch #%d: %v", batchCount, err)
			} else {
				log.Printf("[ApplyChangesWithProgress] ✅ Master data cache fetched for batch #%d", batchCount)
			}

			for _, importData := range batch {
				err := svc.insertNewProduct(holdingCode, authUsername, importData, models.ProductImportHeader{}, masterCache)
				if err != nil {
					errMsg := fmt.Sprintf("Barcode %s: %v", importData.Barcode, err)
					log.Printf("[ApplyChangesWithProgress] ERROR inserting %s", errMsg)
					errorMessages = append(errorMessages, errMsg)
					errorCount++
				}
				processedCount++
			}

			// อัพเดต progress หลังแต่ละ batch
			progress := 30 + (processedCount*50)/totalInsert
			_ = svc.UpdateTaskStatus(holdingCode, taskID, "PROCESSING", progress,
				fmt.Sprintf("Inserted %d/%d records (batch #%d, errors: %d)", processedCount, totalInsert, batchCount, errorCount))
			log.Printf("[ApplyChangesWithProgress] Batch #%d completed: %d/%d records processed, %d errors", batchCount, processedCount, totalInsert, errorCount)
		}

	case "UPDATE_ONLY":
		totalUpdate := len(updateRecords)
		log.Printf("[ApplyChangesWithProgress] Processing UPDATE_ONLY mode - %d records to update with batch size %d", totalUpdate, bSize)
		batchCount = 0 // reset counter

		// Process in batches
		for i := 0; i < totalUpdate; i += bSize {
			batchCount++
			end := i + bSize
			if end > totalUpdate {
				end = totalUpdate
			}
			batch := updateRecords[i:end]
			batchSize := len(batch)

			log.Printf("[ApplyChangesWithProgress] Processing UPDATE batch #%d: records %d-%d (%d records)", batchCount, i+1, end, batchSize)

			// 🚀 Pre-fetch master data cache for this batch (เพิ่ม performance)
			masterCache, err := svc.fetchMasterDataCache(ctx, holdingCode, batch)
			if err != nil {
				log.Printf("[ApplyChangesWithProgress] Warning: Failed to fetch master data cache for batch #%d: %v", batchCount, err)
			} else {
				log.Printf("[ApplyChangesWithProgress] ✅ Master data cache fetched for batch #%d", batchCount)
			}

			for _, importData := range batch {
				existingProduct := existingProductsMap[importData.Barcode]

				// 🔧 สร้าง existingData ที่สมบูรณ์
				existingData := &models.ProductBarcodeExisting{
					Barcode:         existingProduct.Barcode,
					Code:            existingProduct.ItemCode,
					Name:            getLocalizedName(existingProduct.Names),
					UnitCode:        existingProduct.ItemUnitCode,
					Price:           getPrice(existingProduct.Prices, 1),
					PriceMember:     getPrice(existingProduct.Prices, 2),
					PriceDelivery:   getPrice(existingProduct.Prices, 3),
					PriceOne:        getPrice(existingProduct.Prices, 9),
					PriceTwo:        getPrice(existingProduct.Prices, 10),
					PriceThree:      getPrice(existingProduct.Prices, 11),
					PriceFour:       getPrice(existingProduct.Prices, 12),
					PriceFive:       getPrice(existingProduct.Prices, 13),
					PriceSix:        getPrice(existingProduct.Prices, 14),
					PriceSeven:      getPrice(existingProduct.Prices, 15),
					PriceEight:      getPrice(existingProduct.Prices, 16),
					PriceNine:       getPrice(existingProduct.Prices, 17),
					GroupCode:       existingProduct.GroupCode,
					GroupsuboneCode: existingProduct.GroupsuboneCode,
					GroupsubtwoCode: existingProduct.GroupsubtwoCode,
					BrandCode:       existingProduct.BrandCode,
					DesignCode:      existingProduct.DesignCode,
					ModelCode:       existingProduct.ModelCode,
					PatternCode:     existingProduct.PatternCode,
					GradeCode:       existingProduct.GradeCode,
					CategoryCode:    existingProduct.CategoryCode,
					ClassCode:       existingProduct.ClassCode,
					BarcodeRef:      getRefBarcode(existingProduct.RefBarcodes),
					StandValue:      getRefStandValue(existingProduct.RefBarcodes),
					DivideValue:     getRefDivideValue(existingProduct.RefBarcodes),
					IsSumPoint:      existingProduct.IsSumPoint,
				}

				// 🔧 คำนวณ changes ก่อนการ update
				changes := svc.getDetailedChanges(importData, existingData)

				// สร้าง ProductBarcodeCompare สำหรับการ update
				compareItem := &models.ProductBarcodeCompare{
					ImportData:   importData,
					ExistingData: existingData,
					Changes:      changes,
					Status:       "UPDATE",
				}

				if len(changes) > 0 {
					err := svc.updateExistingProduct(holdingCode, authUsername, importData, compareItem, models.ProductImportHeader{}, masterCache)
					if err != nil {
						errMsg := fmt.Sprintf("Barcode %s: %v", importData.Barcode, err)
						log.Printf("[ApplyChangesWithProgress] ERROR updating %s", errMsg)
						errorMessages = append(errorMessages, errMsg)
						errorCount++
					}
				}
				processedCount++
			}

			// อัพเดต progress หลังแต่ละ batch
			progress := 30 + (processedCount*50)/totalUpdate
			_ = svc.UpdateTaskStatus(holdingCode, taskID, "PROCESSING", progress,
				fmt.Sprintf("Updated %d/%d records (batch #%d, errors: %d)", processedCount, totalUpdate, batchCount, errorCount))
			log.Printf("[ApplyChangesWithProgress] UPDATE batch #%d completed: %d/%d records processed, %d errors", batchCount, processedCount, totalUpdate, errorCount)
		}

	case "BOTH", "AUTO":
		totalToProcess := len(newRecords) + len(updateRecords)
		log.Printf("[ApplyChangesWithProgress] Processing BOTH/AUTO mode - %d new, %d update (total: %d) with batch size %d", len(newRecords), len(updateRecords), totalToProcess, bSize)
		batchCount = 0 // reset counter

		// Insert new records in batches
		totalInsert := len(newRecords)
		if totalInsert > 0 {
			log.Printf("[ApplyChangesWithProgress] Starting INSERT phase: %d records", totalInsert)
			for i := 0; i < totalInsert; i += bSize {
				batchCount++
				end := i + bSize
				if end > totalInsert {
					end = totalInsert
				}
				batch := newRecords[i:end]
				batchSize := len(batch)

				log.Printf("[ApplyChangesWithProgress] Processing INSERT batch #%d: records %d-%d (%d records)", batchCount, i+1, end, batchSize)

				// 🚀 Pre-fetch master data cache for this batch (เพิ่ม performance)
				masterCache, err := svc.fetchMasterDataCache(ctx, holdingCode, batch)
				if err != nil {
					log.Printf("[ApplyChangesWithProgress] Warning: Failed to fetch master data cache for INSERT batch #%d: %v", batchCount, err)
				} else {
					log.Printf("[ApplyChangesWithProgress] ✅ Master data cache fetched for INSERT batch #%d", batchCount)
				}

				for _, importData := range batch {
					err := svc.insertNewProduct(holdingCode, authUsername, importData, models.ProductImportHeader{}, masterCache)
					if err != nil {
						errMsg := fmt.Sprintf("Barcode %s: %v", importData.Barcode, err)
						log.Printf("[ApplyChangesWithProgress] ERROR inserting %s", errMsg)
						errorMessages = append(errorMessages, errMsg)
						errorCount++
					}
					processedCount++
				}

				progress := 30 + (processedCount*50)/totalToProcess
				_ = svc.UpdateTaskStatus(holdingCode, taskID, "PROCESSING", progress,
					fmt.Sprintf("Processing %d/%d records (batch #%d, errors: %d)", processedCount, totalToProcess, batchCount, errorCount))
				log.Printf("[ApplyChangesWithProgress] INSERT batch #%d completed: %d/%d total processed", batchCount, processedCount, totalToProcess)
			}
		}

		// Update existing records in batches
		totalUpdate := len(updateRecords)
		if totalUpdate > 0 {
			log.Printf("[ApplyChangesWithProgress] Starting UPDATE phase: %d records", totalUpdate)
			updateBatchCount := 0
			for i := 0; i < totalUpdate; i += bSize {
				updateBatchCount++
				end := i + bSize
				if end > totalUpdate {
					end = totalUpdate
				}
				batch := updateRecords[i:end]
				batchSize := len(batch)

				log.Printf("[ApplyChangesWithProgress] Processing UPDATE batch #%d: records %d-%d (%d records)", updateBatchCount, i+1, end, batchSize)

				// 🚀 Pre-fetch master data cache for this batch (เพิ่ม performance)
				masterCache, err := svc.fetchMasterDataCache(ctx, holdingCode, batch)
				if err != nil {
					log.Printf("[ApplyChangesWithProgress] Warning: Failed to fetch master data cache for UPDATE batch #%d: %v", updateBatchCount, err)
				} else {
					log.Printf("[ApplyChangesWithProgress] ✅ Master data cache fetched for UPDATE batch #%d", updateBatchCount)
				}

				for _, importData := range batch {
					existingProduct := existingProductsMap[importData.Barcode]

					// 🔧 สร้าง existingData ที่สมบูรณ์
					existingData := &models.ProductBarcodeExisting{
						Barcode:         existingProduct.Barcode,
						Code:            existingProduct.ItemCode,
						Name:            getLocalizedName(existingProduct.Names),
						UnitCode:        existingProduct.ItemUnitCode,
						Price:           getPrice(existingProduct.Prices, 1),
						PriceMember:     getPrice(existingProduct.Prices, 2),
						PriceDelivery:   getPrice(existingProduct.Prices, 3),
						PriceOne:        getPrice(existingProduct.Prices, 9),
						PriceTwo:        getPrice(existingProduct.Prices, 10),
						PriceThree:      getPrice(existingProduct.Prices, 11),
						PriceFour:       getPrice(existingProduct.Prices, 12),
						PriceFive:       getPrice(existingProduct.Prices, 13),
						PriceSix:        getPrice(existingProduct.Prices, 14),
						PriceSeven:      getPrice(existingProduct.Prices, 15),
						PriceEight:      getPrice(existingProduct.Prices, 16),
						PriceNine:       getPrice(existingProduct.Prices, 17),
						GroupCode:       existingProduct.GroupCode,
						GroupsuboneCode: existingProduct.GroupsuboneCode,
						GroupsubtwoCode: existingProduct.GroupsubtwoCode,
						BrandCode:       existingProduct.BrandCode,
						DesignCode:      existingProduct.DesignCode,
						ModelCode:       existingProduct.ModelCode,
						PatternCode:     existingProduct.PatternCode,
						GradeCode:       existingProduct.GradeCode,
						CategoryCode:    existingProduct.CategoryCode,
						ClassCode:       existingProduct.ClassCode,
						BarcodeRef:      getRefBarcode(existingProduct.RefBarcodes),
						StandValue:      getRefStandValue(existingProduct.RefBarcodes),
						DivideValue:     getRefDivideValue(existingProduct.RefBarcodes),
						IsSumPoint:      existingProduct.IsSumPoint, // 🆕 เพิ่มบรรทัดนี้

					}

					// 🔧 คำนวณ changes ก่อนการ update
					changes := svc.getDetailedChanges(importData, existingData)

					// สร้าง ProductBarcodeCompare สำหรับการ update
					compareItem := &models.ProductBarcodeCompare{
						ImportData:   importData,
						ExistingData: existingData,
						Changes:      changes,
						Status:       "UPDATE",
					}

					if len(changes) > 0 {
						err := svc.updateExistingProduct(holdingCode, authUsername, importData, compareItem, models.ProductImportHeader{}, masterCache)
						if err != nil {
							errMsg := fmt.Sprintf("Barcode %s: %v", importData.Barcode, err)
							log.Printf("[ApplyChangesWithProgress] ERROR updating %s", errMsg)
							errorMessages = append(errorMessages, errMsg)
							errorCount++
						}
					}
					processedCount++
				}

				progress := 30 + (processedCount*50)/totalToProcess
				_ = svc.UpdateTaskStatus(holdingCode, taskID, "PROCESSING", progress,
					fmt.Sprintf("Processing %d/%d records (batch #%d, errors: %d)", processedCount, totalToProcess, batchCount+updateBatchCount, errorCount))
				log.Printf("[ApplyChangesWithProgress] UPDATE batch #%d completed: %d/%d total processed", updateBatchCount, processedCount, totalToProcess)
			}
		}
	}

	// 🔧 Process RefBarcode updates หลังจากบันทึกข้อมูลเสร็จแล้ว
	log.Printf("[ApplyChangesWithProgress] Starting RefBarcode processing...")
	_ = svc.UpdateTaskStatus(holdingCode, taskID, "PROCESSING", 85, "Processing RefBarcodes...")

	refUpdateCount, refErr := svc.processRefBarcodeUpdates(ctx, holdingCode, authUsername, importDocList)
	if refErr != nil {
		log.Printf("[ApplyChangesWithProgress] Warning: RefBarcode processing failed: %v", refErr)
		// ไม่ return error เพราะ main process สำเร็จแล้ว
	} else {
		log.Printf("[ApplyChangesWithProgress] RefBarcode processing completed: %d products updated", refUpdateCount)
	}

	// Final status
	log.Printf("[ApplyChangesWithProgress] Processing completed: %d processed, %d errors", processedCount, errorCount)

	if errorCount > 0 {
		// 🆕 สร้าง error response เป็น JSON format
		type ErrorDetail struct {
			Summary      string   `json:"summary"`
			TotalErrors  int      `json:"totalerrors"`
			TotalRecords int      `json:"totalrecords"`
			Errors       []string `json:"errors"`
		}

		errorDetail := ErrorDetail{
			Summary:      fmt.Sprintf("Completed with %d errors out of %d records (RefBarcodes: %d updated)", errorCount, processedCount, refUpdateCount),
			TotalErrors:  errorCount,
			TotalRecords: processedCount,
			Errors:       errorMessages,
		}

		// Convert to JSON string
		errorJSON, err := json.Marshal(errorDetail)
		if err != nil {
			log.Printf("[ApplyChangesWithProgress] Warning: Failed to marshal error details: %v", err)
			// Fallback to simple message
			_ = svc.UpdateTaskStatus(holdingCode, taskID, "COMPLETED_WITH_ERRORS", 100,
				fmt.Sprintf("Completed with %d errors out of %d records", errorCount, processedCount))
		} else {
			_ = svc.UpdateTaskStatus(holdingCode, taskID, "COMPLETED_WITH_ERRORS", 100, string(errorJSON))
		}
		log.Printf("[ApplyChangesWithProgress] Task completed with %d errors", errorCount)
	} else {
		_ = svc.UpdateTaskStatus(holdingCode, taskID, "COMPLETED", 100,
			fmt.Sprintf("Successfully processed %d records (RefBarcodes: %d updated)", processedCount, refUpdateCount))
		log.Printf("[ApplyChangesWithProgress] Task completed successfully")
	}

	// ✅ ไม่ return error เมื่อมี validation errors
	// เพราะ status ถูกตั้งเป็น "COMPLETED_WITH_ERRORS" แล้ว และมี error details ใน errormessage
	// Return error เฉพาะ critical errors (database, network, etc.) เท่านั้น
	return nil
}

// Helper function สำหรับหา max value
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
