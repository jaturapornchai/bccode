package services

import (
	"context"
	"errors"
	"fmt"
	"io"

	productbarocde_models "smlcloudplatform/internal/product/productbarcode/models"
	productbarcode_repo "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/stockbalanceimport/models"
	"smlcloudplatform/internal/stockbalanceimport/repositories"
	transactionModels "smlcloudplatform/internal/transaction/models"
	stockBalanceModels "smlcloudplatform/internal/transaction/stockbalance/models"
	stockbalance_models "smlcloudplatform/internal/transaction/stockbalance/models"
	stockbalance_services "smlcloudplatform/internal/transaction/stockbalance/services"
	stockbalancedetail_models "smlcloudplatform/internal/transaction/stockbalancedetail/models"
	stockbalancedetail_services "smlcloudplatform/internal/transaction/stockbalancedetail/services"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strconv"
	"time"

	"github.com/xuri/excelize/v2"
)

type IStockBalanceImportService interface {
	List(shopID string, taskID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockBalanceImportInfo, models.PaginationData, error)
	Create(shopID string, authUsername string, req *models.StockBalanceImport) error
	Update(shopID string, guid string, doc models.StockBalanceImportRaw) error
	Delete(shopID string, guid string) error
	DeleteTask(shopID string, taskID string) error
	ImportFromFile(shopID string, authUsername string, fileUpload io.Reader) (string, error)
	SaveTask(shopID string, authUsername string, taskID string, headerDoc stockbalance_models.StockBalanceHeader) (string, error)
	Meta(shopID string, taskID string) (models.StockBalanceImportMeta, error)
	Verify(shopID string, taskID string) error
}

type StockBalanceImportService struct {
	deafultPartSize           int
	sizeID                    int
	cacheExpire               time.Duration
	chRepo                    repositories.IStockBalanceImportClickHouseRepository
	productBarcodeRepo        productbarcode_repo.IProductBarcodeRepository
	stockBalanceService       stockbalance_services.IStockBalanceHttpService
	stockBalanceDetailService stockbalancedetail_services.IStockBalanceDetailService
	generateID                func(int) string
	generateGUID              func() string
	timeNow                   func() time.Time
}

func NewStockBalanceImportService(
	chRepo repositories.IStockBalanceImportClickHouseRepository,
	productBarcodeRepo productbarcode_repo.IProductBarcodeRepository,
	stockBalanceService stockbalance_services.IStockBalanceHttpService,
	stockBalanceDetailService stockbalancedetail_services.IStockBalanceDetailService,
	generateID func(int) string,
	generateGUID func() string,
	timeNow func() time.Time,
) *StockBalanceImportService {
	return &StockBalanceImportService{
		deafultPartSize:           100,
		sizeID:                    12,
		cacheExpire:               time.Minute * 60,
		chRepo:                    chRepo,
		productBarcodeRepo:        productBarcodeRepo,
		stockBalanceService:       stockBalanceService,
		stockBalanceDetailService: stockBalanceDetailService,
		generateID:                generateID,
		generateGUID:              generateGUID,
		timeNow:                   timeNow,
	}
}

func (svc StockBalanceImportService) List(shopID string, taskID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.StockBalanceImportInfo, models.PaginationData, error) {
	findDocs, patination, err := svc.chRepo.List(context.Background(), shopID, taskID, filters, pageable)

	if err != nil {
		return []models.StockBalanceImportInfo{}, models.PaginationData{}, err
	}

	results := []models.StockBalanceImportInfo{}

	for _, doc := range findDocs {
		results = append(results, doc.StockBalanceImportInfo)
	}

	return results, patination, nil
}

func (svc StockBalanceImportService) Create(shopID string, authUsername string, doc *models.StockBalanceImport) error {
	docData := models.StockBalanceImportDoc{}
	docData.ShopID = shopID
	docData.GUIDFixed = svc.generateGUID()
	docData.StockBalanceImport = *doc

	result, err := svc.chRepo.FindOne(context.Background(), shopID, doc.TaskID, []micromodels.KeyInt{
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

func (svc StockBalanceImportService) ImportFromFile(shopID string, authUsername string, fileUpload io.Reader) (string, error) {

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

	expectColName := []string{
		"Barcode",
		"Name",
		"Unit Code",
		"Warehouse Code",
		"Shelf Code",
		"Qty",
		"Amount",
	}

	for _, colName := range expectColName {
		if _, ok := colIdxs[colName]; !ok {
			return "", fmt.Errorf("column %s not found", colName)
		}
	}

	prepareDataDoc := []models.StockBalanceImportDoc{}

	taskID := svc.generateID(svc.sizeID)

	for i, doc := range rows {

		if i == 0 {
			continue
		}

		tempData, err := svc.prepareData(shopID, taskID, float64(i), colIdxs, doc)
		if err != nil {
			return "", err
		}

		prepareDataDoc = append(prepareDataDoc, tempData)
	}

	createdAt := svc.timeNow()
	createdBy := authUsername

	for i := range prepareDataDoc {
		prepareDataDoc[i].CreatedAt = createdAt
		prepareDataDoc[i].CreatedBy = createdBy
	}

	err = svc.chRepo.CreateInBatch(context.Background(), prepareDataDoc)

	if err != nil {
		return "", err
	}

	return taskID, nil
}
func (svc StockBalanceImportService) prepareData(shopID string, taskID string, rowNumber float64, colIdx map[string]int, doc []string) (models.StockBalanceImportDoc, error) {

	qty, err := strconv.ParseFloat(doc[colIdx["Qty"]], 64)

	if err != nil {
		return models.StockBalanceImportDoc{}, fmt.Errorf("qty in row %d invalid", int(rowNumber))
	}

	amount, err := strconv.ParseFloat(doc[colIdx["Amount"]], 64)

	if err != nil {
		return models.StockBalanceImportDoc{}, fmt.Errorf("amount in row %d invalid", int(rowNumber))
	}

	price := float64(0)

	if qty > 0 && amount > 0 {
		price = amount / qty
	}

	newGUID := svc.generateGUID()

	dataDoc := models.StockBalanceImportDoc{}

	dataDoc.GUIDFixed = newGUID
	dataDoc.ShopID = shopID
	dataDoc.TaskID = taskID
	dataDoc.RowNumber = rowNumber
	dataDoc.Barcode = doc[colIdx["Barcode"]]
	dataDoc.Name = doc[colIdx["Name"]]
	dataDoc.UnitCode = doc[colIdx["Unit Code"]]
	dataDoc.WarehouseCode = doc[colIdx["Warehouse Code"]]
	dataDoc.ShelfCode = doc[colIdx["Shelf Code"]]
	dataDoc.Qty = qty
	dataDoc.Price = price
	dataDoc.SumAmount = amount

	return dataDoc, nil
}

func (svc StockBalanceImportService) Meta(shopID string, taskID string) (models.StockBalanceImportMeta, error) {
	return svc.chRepo.Meta(context.Background(), shopID, taskID)
}

func (svc StockBalanceImportService) Update(shopID string, guid string, doc models.StockBalanceImportRaw) error {
	return svc.chRepo.Update(context.Background(), shopID, guid, doc)
}

func (svc StockBalanceImportService) Delete(shopID string, guid string) error {
	return svc.chRepo.DeleteByGUID(context.Background(), shopID, guid)
}

func (svc StockBalanceImportService) DeleteTask(shopID string, taskID string) error {
	return svc.chRepo.DeleteByTaskID(context.Background(), shopID, taskID)
}

func (svc StockBalanceImportService) SaveTask(shopID string, authUsername string, taskID string, headerDoc stockbalance_models.StockBalanceHeader) (string, error) {

	err := svc.Verify(shopID, taskID)
	if err != nil {
		return "", err
	}

	countNotExist, err := svc.chRepo.CountExist(context.Background(), shopID, taskID, false)

	if err != nil {
		return "", err
	}

	if countNotExist > 0 {
		return "", errors.New("have barcode not found in product")
	}

	docs, err := svc.chRepo.All(context.Background(), shopID, taskID)

	if err != nil {
		return "", err
	}

	tempDetails := []stockbalancedetail_models.StockBalanceDetail{}
	tempDocDetails := []transactionModels.Detail{}

	barcodes := []string{}
	tempBarcodes := map[string]models.StockBalanceImportDoc{}
	for i, doc := range docs {
		barcodes = append(barcodes, doc.Barcode)
		tempBarcodes[docs[i].Barcode] = docs[i]
		if (i > 1 && i%5000 == 0) || i == len(docs)-1 {
			productList, err := svc.productBarcodeRepo.FindByBarcodes(context.Background(), shopID, barcodes)
			if err != nil {
				return "", err
			}

			productBarcodeDict := map[string]productbarocde_models.ProductBarcodeInfo{}
			for _, product := range productList {
				productBarcodeDict[product.Barcode] = product
			}

			for _, barcode := range barcodes {
				if _, ok := productBarcodeDict[barcode]; !ok {
					return "", fmt.Errorf("barcode %s not found", barcode)
				}
			}

			for _, product := range productList {
				stockbalanceDetail := stockbalancedetail_models.StockBalanceDetail{}
				temp := tempBarcodes[product.Barcode]

				stockbalanceDetail.ItemCode = product.ItemCode
				stockbalanceDetail.Barcode = product.Barcode
				stockbalanceDetail.ItemNames = product.Names
				stockbalanceDetail.ItemType = product.ItemType
				stockbalanceDetail.TaxType = product.TaxType
				stockbalanceDetail.VatType = product.VatType
				stockbalanceDetail.DivideValue = product.DivideValue
				stockbalanceDetail.StandValue = product.StandValue
				stockbalanceDetail.VatCal = product.VatCal
				stockbalanceDetail.UnitCode = product.ItemUnitCode
				stockbalanceDetail.UnitNames = product.ItemUnitNames

				stockbalanceDetail.Qty = temp.Qty
				stockbalanceDetail.Price = temp.Price
				stockbalanceDetail.SumAmount = temp.SumAmount

				tempDetails = append(tempDetails, stockbalanceDetail)

				// create transaction detail for produce in kafka
				tempDocDetail := transactionModels.Detail{}
				tempDocDetail.DocRef = headerDoc.DocNo
				tempDocDetail.DocRefDatetime = headerDoc.DocDatetime
				tempDocDetail.Barcode = product.Barcode
				tempDocDetail.ItemNames = product.Names
				tempDocDetail.ItemType = product.ItemType
				tempDocDetail.TaxType = product.TaxType
				tempDocDetail.VatType = product.VatType
				tempDocDetail.StandValue = product.StandValue
				tempDocDetail.DivideValue = product.DivideValue
				tempDocDetail.VatCal = product.VatCal
				tempDocDetail.UnitCode = product.ItemUnitCode
				tempDocDetail.WhCode = temp.WarehouseCode
				tempDocDetail.LocationCode = temp.ShelfCode
				tempDocDetail.Qty = temp.Qty
				tempDocDetail.Price = temp.Price
				tempDocDetail.PriceExcludeVat = temp.Price
				tempDocDetail.TotalValueVat = temp.SumAmount
				tempDocDetail.SumAmount = temp.SumAmount
				tempDocDetail.SumAmountExcludeVat = temp.SumAmount

				tempDocDetails = append(tempDocDetails, tempDocDetail)
			}

			barcodes = []string{}
			tempBarcodes = map[string]models.StockBalanceImportDoc{}

		}
	}

	tempTransaction := stockbalance_models.StockBalance{}

	tempTransaction.StockBalanceHeader = headerDoc

	docCreate, docGUIDFixed, docNo, err := svc.stockBalanceService.CreateStockBalance(shopID, authUsername, tempTransaction)
	if err != nil {
		return "", err
	}

	for i := range tempDetails {
		tempDetails[i].DocNo = docNo
	}

	err = svc.stockBalanceDetailService.CreateStockBalanceDetail(shopID, authUsername, tempDetails)
	if err != nil {
		err = svc.stockBalanceService.DeleteStockBalance(shopID, docGUIDFixed, authUsername)

		if err != nil {
			return "", err
		}

		return "", err
	}

	stockBalanceMessage := stockBalanceModels.StockBalanceMessage{}
	stockBalanceMessage.StockBalanceHeader = headerDoc
	stockBalanceMessage.DocNo = docCreate.DocNo
	stockBalanceMessage.ShopID = docCreate.ShopID
	stockBalanceMessage.CreatedBy = docCreate.CreatedBy
	stockBalanceMessage.CreatedAt = docCreate.CreatedAt
	stockBalanceMessage.Details = &tempDocDetails

	err = svc.stockBalanceService.ProduceCreateStockBalance(shopID, stockBalanceMessage)
	if err != nil {
		return "", err
	}

	err = svc.DeleteTask(shopID, taskID)

	if err != nil {
		return "", err
	}

	return docNo, nil
}

func (svc StockBalanceImportService) Verify(shopID string, taskID string) error {
	docs, err := svc.chRepo.All(context.Background(), shopID, taskID)

	if err != nil {
		return err
	}

	previousNotExist := map[string]struct{}{}

	itemNotExist := map[string]struct{}{}

	tempBarcodes := []string{}

	tempItemDict := map[string]models.StockBalanceImportDoc{}
	for i, doc := range docs {

		if doc.IsNotExist {
			previousNotExist[doc.Barcode] = struct{}{}
		}

		tempItemDict[doc.Barcode] = doc
		tempBarcodes = append(tempBarcodes, doc.Barcode)

		if (i > 1 && i%5000 == 0) || i == len(docs)-1 {
			productList, err := svc.productBarcodeRepo.FindByBarcodes(context.Background(), shopID, tempBarcodes)
			if err != nil {
				return err
			}

			for _, barcode := range tempBarcodes {
				itemNotExist[barcode] = struct{}{}
			}

			for _, product := range productList {
				if _, ok := tempItemDict[product.Barcode]; ok {
					delete(itemNotExist, product.Barcode)
				}
			}

			//Clear previous exist
			if err := svc.updateExist(shopID, taskID, true, previousNotExist); err != nil {
				return err
			}

			// Update exist
			if err := svc.updateExist(shopID, taskID, false, itemNotExist); err != nil {
				return err
			}

			previousNotExist = map[string]struct{}{}
			itemNotExist = map[string]struct{}{}
			tempBarcodes = []string{}
			tempItemDict = map[string]models.StockBalanceImportDoc{}
		}
	}

	return nil

}

func (svc StockBalanceImportService) updateExist(shopID string, taskID string, isExist bool, barcodes map[string]struct{}) error {
	if len(barcodes) == 0 {
		return nil
	}

	tempBarcodes := []string{}
	for barcode := range barcodes {
		tempBarcodes = append(tempBarcodes, barcode)
	}
	err := svc.chRepo.UpdateExist(context.Background(), shopID, taskID, isExist, tempBarcodes)

	if err != nil {
		return err
	}

	return nil
}
