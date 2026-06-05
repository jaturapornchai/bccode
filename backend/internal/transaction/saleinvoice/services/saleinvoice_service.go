package services

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	couponModels "smlcloudplatform/internal/coupon/models"
	couponServices "smlcloudplatform/internal/coupon/services"
	custModels "smlcloudplatform/internal/debtaccount/debtor/models"
	repoCust "smlcloudplatform/internal/debtaccount/debtor/repositories"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_models "smlcloudplatform/internal/product/productbarcode/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/services"
	trans_models "smlcloudplatform/internal/transaction/models"
	trans_cache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/transaction/saleinvoice/models"
	"smlcloudplatform/internal/transaction/saleinvoice/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strconv"
	"strings"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type ISaleInvoiceService interface {
	CreateSaleInvoice(holdingCode string, authUsername string, doc models.SaleInvoice) (string, string, error)
	CreateSaleInvoiceWithAutoCoupon(holdingCode string, authUsername string, doc models.SaleInvoice) (string, string, error)
	UpdateSaleInvoice(holdingCode string, guid string, authUsername string, doc models.SaleInvoice) error
	UpdateSlip(holdingCode string, authUsername string, docNo string, mode uint8, machineCode string, zoneGroupNumber string, imageUrl string) error
	RecalPoint(holdingCode string, code string, authUsername string) error
	DeleteSaleInvoice(holdingCode string, guid string, authUsername string) error
	DeleteSaleInvoiceByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoSaleInvoice(holdingCode string, guid string) (models.SaleInvoiceInfo, error)
	InfoSaleInvoiceByCode(holdingCode string, code string) (models.SaleInvoiceInfo, error)
	InfoSaleInvoiceByGuidPos(holdingCode string, code string) (models.SaleInvoiceInfo, error)
	SearchSaleInvoice(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SaleInvoiceInfo, mongopagination.PaginationData, error)
	SearchSaleInvoiceStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SaleInvoiceInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.SaleInvoice) (common.BulkImport, error)
	GetLastPOSDocNo(holdingCode, posID, maxDocNo string) (string, error)
	Export(languageCode string, holdingCode string, languageHeader map[string]string) ([][]string, error)

	GetModuleName() string
}

const (
	MODULE_NAME = "SI"
	TRANS_FLAG  = 44

	// Point transaction limits and configuration
	MaxPointAmountPerTransaction = 1000000.0 // Maximum points per transaction
	MaxPointsCodeLength          = 50        // Maximum length for points code
	MinPointAmount               = 0.01      // Minimum point amount
)

type ISaleInvocieParser interface {
	ParseProductBarcode(detail trans_models.Detail, productBarcodeInfo productbarcode_models.ProductBarcodeInfo) trans_models.Detail
}

type ISaleInvoiceExport interface {
	ParseCSV(languageCode string, data models.SaleInvoiceInfo) [][]string
}

type SaleInvoiceService struct {
	repoMq               repositories.ISaleInvoiceMessageQueueRepository
	repoCust             repoCust.IDebtorRepository
	repo                 repositories.ISaleInvoiceRepository
	repoCache            trans_cache.ICacheRepository
	productbarcodeRepo   productbarcode_repositories.IProductBarcodeRepository
	pointTransactionRepo repoCust.IPointTransactionRepository
	cacheExpireDocNo     time.Duration
	syncCacheRepo        mastersync.IMasterSyncCacheRepository
	couponService        couponServices.ICouponHttpService // Add coupon service
	services.ActivityService[models.SaleInvoiceActivity, models.SaleInvoiceDeleteActivity]
	parser         ISaleInvocieParser
	exporter       ISaleInvoiceExport
	contextTimeout time.Duration
	pointConfig    PointTransactionConfig // Add point configuration
	logger         *slog.Logger           // Add structured logger
}

func NewSaleInvoiceService(
	repo repositories.ISaleInvoiceRepository,
	repoCust repoCust.IDebtorRepository,
	repoCache trans_cache.ICacheRepository,
	productbarcodeRepo productbarcode_repositories.IProductBarcodeRepository,
	pointTransactionRepo repoCust.IPointTransactionRepository,
	repoMq repositories.ISaleInvoiceMessageQueueRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
	couponService couponServices.ICouponHttpService, // Add coupon service parameter
	parser ISaleInvocieParser,
	exporter ISaleInvoiceExport,
) *SaleInvoiceService {

	contextTimeout := time.Duration(15) * time.Second

	// Initialize structured logger
	logger := slog.Default().With(
		slog.String("service", "SaleInvoiceService"),
		slog.String("module", MODULE_NAME),
	)

	insSvc := &SaleInvoiceService{
		repo:                 repo,
		repoCust:             repoCust,
		repoMq:               repoMq,
		pointTransactionRepo: pointTransactionRepo,
		repoCache:            repoCache,
		productbarcodeRepo:   productbarcodeRepo,
		syncCacheRepo:        syncCacheRepo,
		couponService:        couponService, // Add coupon service assignment
		parser:               parser,
		exporter:             exporter,
		cacheExpireDocNo:     time.Hour * 24,
		contextTimeout:       contextTimeout,
		pointConfig:          DefaultPointTransactionConfig(),
		logger:               logger,
	}

	insSvc.ActivityService = services.NewActivityService[models.SaleInvoiceActivity, models.SaleInvoiceDeleteActivity](repo)

	return insSvc
}

func (svc SaleInvoiceService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc SaleInvoiceService) getDocNoPrefix(docDate time.Time) string {
	docDateStr := docDate.Format("20060102")
	return fmt.Sprintf("%s%s", MODULE_NAME, docDateStr)
}

func (svc SaleInvoiceService) generateNewDocNo(ctx context.Context, holdingCode, prefixDocNo string, docNumber int) (string, int, error) {
	prevoiusDocNumber, err := svc.repoCache.Get(holdingCode, prefixDocNo)

	if prevoiusDocNumber == 0 || err != nil {
		lastDoc, err := svc.repo.FindLastDocNo(ctx, holdingCode, prefixDocNo)

		if err != nil {
			return "", 0, err
		}

		if len(lastDoc.DocNo) > 0 {
			rawNumber := strings.Replace(lastDoc.DocNo, prefixDocNo, "", -1)
			prevoiusDocNumber, err = strconv.Atoi(rawNumber)

			if err != nil {
				prevoiusDocNumber = 0
			}
		}

	}

	newDocNumber := prevoiusDocNumber + 1
	newDocNo := fmt.Sprintf("%s%05d", prefixDocNo, newDocNumber)

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "docno", newDocNo)

	if err != nil {
		return "", 0, err
	}

	if len(findDoc.GuidFixed) > 0 {
		return "", 0, errors.New("DocNo is exists")
	}

	return newDocNo, newDocNumber, nil
}

func (svc SaleInvoiceService) processPointTransactions(ctx context.Context, holdingCode string, authUsername string, custCode string, pointsCode string, docNo string, docDate time.Time, getPoint float64, usePoint float64) error {
	svc.logger.Debug("Starting point transaction processing",
		slog.String("holdingCode", holdingCode),
		slog.String("custCode", custCode),
		slog.String("pointsCode", pointsCode),
		slog.String("docNo", docNo),
		slog.Float64("getPoint", getPoint),
		slog.Float64("usePoint", usePoint),
	)

	// Skip processing if no points involved
	if getPoint == 0 && usePoint == 0 {
		svc.logger.Debug("Skipping point transaction - no points involved")
		return nil
	}

	// Skip processing if pointsCode is empty (required for new point system)
	if pointsCode == "" {
		svc.logger.Debug("Skipping point transaction - pointsCode is empty")
		return nil
	}

	// Validate point transaction parameters
	if err := svc.validatePointTransaction(custCode, pointsCode, getPoint, usePoint); err != nil {
		svc.logger.Error("Point transaction validation failed",
			slog.String("error", err.Error()),
			slog.String("custCode", custCode),
			slog.String("pointsCode", pointsCode),
		)
		return err
	}

	// Create extended timeout context for point operations
	extendedCtx, cancel := svc.getPointTransactionContext()
	defer cancel()

	// Handle UsePoint - ใช้ custCode (เหมือนเดิม)
	if usePoint > 0 && custCode != "" {
		svc.logger.Info("Processing UsePoint transaction",
			slog.String("custCode", custCode),
			slog.Float64("usePoint", usePoint),
		)

		err := svc.processUsePointTransaction(extendedCtx, holdingCode, authUsername, custCode, docNo, docDate, usePoint)
		if err != nil {
			svc.logger.Error("UsePoint transaction failed",
				slog.String("error", err.Error()),
				slog.String("custCode", custCode),
				slog.Float64("usePoint", usePoint),
			)
			return err
		}
	}

	// Handle GetPoint - ใช้ pointsCode (ระบบใหม่)
	if getPoint > 0 && pointsCode != "" {
		svc.logger.Info("Processing GetPoint transaction",
			slog.String("pointsCode", pointsCode),
			slog.Float64("getPoint", getPoint),
		)

		err := svc.processGetPointTransaction(extendedCtx, holdingCode, authUsername, custCode, pointsCode, docNo, docDate, getPoint)
		if err != nil {
			svc.logger.Error("GetPoint transaction failed",
				slog.String("error", err.Error()),
				slog.String("pointsCode", pointsCode),
				slog.Float64("getPoint", getPoint),
			)
			return err
		}
	}

	svc.logger.Info("Point transaction processing completed successfully",
		slog.String("docNo", docNo),
		slog.Float64("getPoint", getPoint),
		slog.Float64("usePoint", usePoint),
	)

	return nil
}

func (svc SaleInvoiceService) RecalPoint(holdingCode string, code string, authUsername string) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	if code == "" {
		return errors.New("code is required")
	}

	updateCust := &custModels.DebtorDoc{}

	findCust, err := svc.repoCust.FindByDocIndentityGuid(ctx, holdingCode, "code", code)

	if err != nil {
		return err
	}

	updateCust = &findCust
	updateCust.GuidFixed = findCust.GuidFixed
	updateCust.UpdatedBy = authUsername
	updateCust.UpdatedAt = time.Now()

	searchInFields := []string{
		"custcode",
	}

	findDoc, err := svc.repo.Find(context.Background(), holdingCode, searchInFields, code)

	if err != nil {
		return err
	}

	CalGetPoint := 0.0
	CalUsePoint := 0.0
	if len(findDoc) > 1 {
		//CalPoint += findDoc.GetPoint
		for _, doc := range findDoc {
			CalGetPoint += doc.GetPoint
			CalUsePoint += doc.UsePoint
		}
	}
	updateCust.PointBalance = CalGetPoint - CalUsePoint
	err = svc.repoCust.Update(ctx, holdingCode, code, *updateCust)
	if err != nil {
		return err
	}

	return nil
}

func (svc SaleInvoiceService) CreateSaleInvoice(holdingCode string, authUsername string, doc models.SaleInvoice) (string, string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	isGenerateDocNo := !doc.IsPOS

	prefixDocNo, docNo, newDocNumber := "", "", 0

	if isGenerateDocNo {
		docDate := doc.DocDatetime
		prefixDocNo = svc.getDocNoPrefix(docDate)

		tempNewDocNo, tempNewDocNumber, err := svc.generateNewDocNo(ctx, holdingCode, prefixDocNo, 1)

		if err != nil {
			return "", "", err
		}

		docNo = tempNewDocNo
		newDocNumber = tempNewDocNumber
	} else {
		if doc.DocNo == "" {
			return "", "", errors.New("docno is required")
		}

		docNo = doc.DocNo
	}

	if isGenerateDocNo {
		findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "docno", docNo)

		if err != nil {
			return "", "", err
		}

		if len(findDoc.GuidFixed) > 0 {
			return "", "", errors.New("DocNo is exists")
		}
	}

	newGuidFixed := utils.NewGUID()

	dataDoc := models.SaleInvoiceDoc{}
	dataDoc.HoldingCode = holdingCode
	dataDoc.GuidFixed = newGuidFixed
	dataDoc.SaleInvoice = doc
	dataDoc.IsClose = false

	dataDoc.TransFlag = TRANS_FLAG
	dataDoc.DocNo = docNo
	if isGenerateDocNo && doc.TaxDocNo == "" {
		dataDoc.TaxDocNo = docNo
	}

	productBarcodes, err := svc.GetDetailProductBarcodes(ctx, holdingCode, *doc.Details)
	if err != nil {
		return "", "", err
	}

	details := svc.PrepareDetail(*doc.Details, productBarcodes)
	dataDoc.Details = &details

	if dataDoc.PointsCode != "" {
		// Handle point transactions (earning and redeeming)
		err := svc.processPointTransactions(ctx, holdingCode, authUsername, dataDoc.CustCode, dataDoc.PointsCode, dataDoc.DocNo, dataDoc.DocDatetime, dataDoc.GetPoint, dataDoc.UsePoint)
		if err != nil {
			return "", "", err
		}
	}

	dataDoc.CreatedBy = authUsername
	dataDoc.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, dataDoc)

	if err != nil {
		return "", "", err
	}

	go func() {
		err := svc.repoMq.Create(dataDoc)
		if err != nil {
			fmt.Printf("create mq error :: %s", err.Error())
		}
		svc.saveMasterSync(holdingCode)

		if isGenerateDocNo {
			svc.repoCache.Save(holdingCode, prefixDocNo, newDocNumber, svc.cacheExpireDocNo)
		}
	}()

	return newGuidFixed, docNo, nil
}

// CreateSaleInvoiceWithAutoCoupon creates a sale invoice and processes coupons asynchronously
// This method provides automatic coupon usage with error isolation between invoice creation and coupon processing
func (svc SaleInvoiceService) CreateSaleInvoiceWithAutoCoupon(holdingCode string, authUsername string, doc models.SaleInvoice) (string, string, error) {
	svc.logger.Info("Starting CreateSaleInvoiceWithAutoCoupon",
		slog.String("holdingCode", holdingCode),
		slog.String("authUsername", authUsername),
		slog.String("docNo", doc.DocNo),
		slog.Int("couponCount", len(doc.Coupons)),
	)

	// Step 1: Create the sale invoice normally (synchronous)
	invoiceID, docNo, err := svc.CreateSaleInvoice(holdingCode, authUsername, doc)
	if err != nil {
		svc.logger.Error("Failed to create sale invoice",
			slog.String("error", err.Error()),
			slog.String("docNo", doc.DocNo),
		)
		return "", "", err
	}

	svc.logger.Info("Sale invoice created successfully",
		slog.String("invoiceID", invoiceID),
		slog.String("docNo", docNo),
	)

	// Step 2: Process coupons asynchronously in background (non-blocking)
	if len(doc.Coupons) > 0 {
		go svc.processAutoCoupons(holdingCode, authUsername, invoiceID, docNo, doc.Coupons, doc.CustCode)
	}

	return invoiceID, docNo, nil
}

// processAutoCoupons handles the asynchronous coupon processing
func (svc SaleInvoiceService) processAutoCoupons(holdingCode, authUsername, invoiceID, docNo string, coupons []trans_models.SaleInvoiceCoupon, customerID string) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	svc.logger.Info("Starting async coupon processing",
		slog.String("holdingCode", holdingCode),
		slog.String("invoiceID", invoiceID),
		slog.String("docNo", docNo),
		slog.Int("couponCount", len(coupons)),
		slog.String("customerID", customerID),
	)

	var successCount, failureCount int

	for _, coupon := range coupons {
		err := svc.processSingleCoupon(ctx, holdingCode, authUsername, invoiceID, docNo, coupon, customerID)
		if err != nil {
			failureCount++
			svc.logger.Error("Failed to process coupon",
				slog.String("couponNo", coupon.CouponNo),
				slog.String("invoiceID", invoiceID),
				slog.String("error", err.Error()),
			)
		} else {
			successCount++
			svc.logger.Info("Coupon processed successfully",
				slog.String("couponNo", coupon.CouponNo),
				slog.String("invoiceID", invoiceID),
				slog.Float64("amount", coupon.CouponAmount),
			)
		}
	}

	svc.logger.Info("Completed async coupon processing",
		slog.String("invoiceID", invoiceID),
		slog.String("docNo", docNo),
		slog.Int("successCount", successCount),
		slog.Int("failureCount", failureCount),
	)
}

// processSingleCoupon processes a single coupon usage
func (svc SaleInvoiceService) processSingleCoupon(ctx context.Context, holdingCode, authUsername, invoiceID, docNo string, coupon trans_models.SaleInvoiceCoupon, customerID string) error {
	// Find the coupon by code to get its ID
	coupons, err := svc.couponService.SearchCoupon(holdingCode, coupon.CouponNo)
	if err != nil {
		return fmt.Errorf("failed to search coupon %s: %w", coupon.CouponNo, err)
	}

	if len(coupons) == 0 {
		return fmt.Errorf("coupon not found: %s", coupon.CouponNo)
	}

	couponID := coupons[0].GuidFixed

	// Use the existing coupon service to process the coupon
	// This assumes the coupon is already reserved and we're now using it
	useCouponReq := couponModels.UseCouponRequest{
		CustomerID:    customerID,
		TransactionID: fmt.Sprintf("%s-%s", docNo, invoiceID), // Unique transaction ID
		UseAmount:     coupon.CouponAmount,
		// Note: ReservationID would need to be passed from the frontend or stored during reservation
		// For now, we'll implement a simple version that works with pre-reserved coupons
	}

	// Use the coupon
	_, err = svc.couponService.UseCoupon(couponID, holdingCode, authUsername, useCouponReq)
	if err != nil {
		return fmt.Errorf("failed to use coupon %s: %w", coupon.CouponNo, err)
	}

	return nil
}

func (svc SaleInvoiceService) GetDetailProductBarcodes(ctx context.Context, holdingCode string, details []trans_models.Detail) ([]productbarcode_models.ProductBarcodeInfo, error) {
	var tempBarcodes []string
	for _, doc := range details {
		tempBarcodes = append(tempBarcodes, doc.Barcode)
	}
	return svc.productbarcodeRepo.FindByBarcodes(ctx, holdingCode, tempBarcodes)
}

func (svc SaleInvoiceService) PrepareDetail(details []trans_models.Detail, productBarcodes []productbarcode_models.ProductBarcodeInfo) []trans_models.Detail {

	productBarcodeDict := map[string]productbarcode_models.ProductBarcodeInfo{}
	for _, doc := range productBarcodes {
		productBarcodeDict[doc.Barcode] = doc
	}

	for i := 0; i < len(details); i++ {
		tempDetail := (details)[i]
		tempProduct := productBarcodeDict[tempDetail.Barcode]
		if _, ok := productBarcodeDict[tempDetail.Barcode]; ok {
			tempDetail = svc.parser.ParseProductBarcode(tempDetail, tempProduct)
		}
		tempDetail.Discount = (details)[i].Discount
		tempDetail.StandValue = (details)[i].StandValue
		tempDetail.DivideValue = (details)[i].DivideValue

		(details)[i] = tempDetail
	}

	return details
}

func (svc SaleInvoiceService) UpdateSaleInvoice(holdingCode string, guid string, authUsername string, doc models.SaleInvoice) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	dataDoc := findDoc
	dataDoc.SaleInvoice = doc

	productBarcodes, err := svc.GetDetailProductBarcodes(ctx, holdingCode, *doc.Details)
	if err != nil {
		return err
	}

	details := svc.PrepareDetail(*doc.Details, productBarcodes)
	dataDoc.Details = &details

	dataDoc.SlipQrUrl = findDoc.SlipQrUrl
	dataDoc.SlipQrUrlHistories = findDoc.SlipQrUrlHistories
	dataDoc.SlipUrl = findDoc.SlipUrl
	dataDoc.SlipUrlHistories = findDoc.SlipUrlHistories

	dataDoc.DocNo = findDoc.DocNo
	dataDoc.TransFlag = TRANS_FLAG
	dataDoc.UpdatedBy = authUsername
	dataDoc.UpdatedAt = time.Now()

	// Handle point transactions when updating
	// Step 1: Delete old point transactions for this document
	if findDoc.PointsCode != "" && (findDoc.GetPoint > 0 || findDoc.UsePoint > 0) {
		err := svc.pointTransactionRepo.DeletePointTransactionsByDocNo(ctx, holdingCode, findDoc.DocNo, authUsername)
		if err != nil {
			return fmt.Errorf("failed to delete old point transactions: %w", err)
		}
	}

	// Step 2: Handle point refund when cancelling transaction (iscancel = true)
	if doc.IsCancel && !findDoc.IsCancel && dataDoc.PointsCode != "" {
		// Process point refund when cancelling a transaction
		err := svc.processPointCancelTransactions(ctx, holdingCode, authUsername, dataDoc.CustCode, dataDoc.PointsCode, dataDoc.DocNo, dataDoc.DocDatetime, dataDoc.GetPoint, dataDoc.UsePoint)
		if err != nil {
			return err
		}
	} else if !doc.IsCancel && dataDoc.PointsCode != "" {
		// Step 3: Create new point transactions if not cancelled
		err := svc.processPointTransactions(ctx, holdingCode, authUsername, dataDoc.CustCode, dataDoc.PointsCode, dataDoc.DocNo, dataDoc.DocDatetime, dataDoc.GetPoint, dataDoc.UsePoint)
		if err != nil {
			return fmt.Errorf("failed to process point transactions: %w", err)
		}
	}

	// Step 4: Recalculate point balance
	if dataDoc.PointsCode != "" {
		err := svc.pointTransactionRepo.RecalculatePointBalanceByPointsCode(ctx, holdingCode, dataDoc.PointsCode)
		if err != nil {
			return fmt.Errorf("failed to recalculate point balance: %w", err)
		}
	}

	err = svc.repo.Update(ctx, holdingCode, guid, dataDoc)

	if err != nil {
		return err
	}

	go func() {
		err := svc.repoMq.Update(dataDoc)
		if err != nil {
			fmt.Printf("create mq error :: %s", err.Error())
		}
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc SaleInvoiceService) UpdateSlip(holdingCode string, authUsername string, docNo string, mode uint8, machineCode string, zoneGroupNumber string, imageUrl string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "docno", docNo)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	dataDoc := findDoc

	if mode == 1 {
		if findDoc.SlipQrUrl != "" {
			dataDoc.SlipQrUrlHistories = append(dataDoc.SlipQrUrlHistories, findDoc.SlipQrUrl)
		}

		dataDoc.SlipQrUrl = imageUrl
	} else {
		if findDoc.SlipUrl != "" {
			dataDoc.SlipUrlHistories = append(dataDoc.SlipUrlHistories, findDoc.SlipUrl)
		}

		dataDoc.SlipUrl = imageUrl
	}

	dataDoc.MachineCode = machineCode
	dataDoc.ZoneGroupNumber = zoneGroupNumber

	dataDoc.UpdatedBy = authUsername
	dataDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, findDoc.GuidFixed, dataDoc)

	if err != nil {
		return err
	}

	go func() {
		err := svc.repoMq.Update(dataDoc)
		if err != nil {
			fmt.Printf("create mq error :: %s", err.Error())
		}
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc SaleInvoiceService) DeleteSaleInvoice(holdingCode string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	// Step 1: Refund points before deleting if document has points
	if findDoc.PointsCode != "" && (findDoc.GetPoint > 0 || findDoc.UsePoint > 0) {
		// Process point refund (reverse both GetPoint and UsePoint)
		err := svc.processPointCancelTransactions(ctx, holdingCode, authUsername, findDoc.CustCode, findDoc.PointsCode, findDoc.DocNo, findDoc.DocDatetime, findDoc.GetPoint, findDoc.UsePoint)
		if err != nil {
			return fmt.Errorf("failed to refund points: %w", err)
		}
	}

	// Step 2: Delete all point transactions for this document
	if findDoc.DocNo != "" {
		err := svc.pointTransactionRepo.DeletePointTransactionsByDocNo(ctx, holdingCode, findDoc.DocNo, authUsername)
		if err != nil {
			return fmt.Errorf("failed to delete point transactions: %w", err)
		}
	}

	// Step 3: Recalculate point balance
	if findDoc.PointsCode != "" {
		err := svc.pointTransactionRepo.RecalculatePointBalanceByPointsCode(ctx, holdingCode, findDoc.PointsCode)
		if err != nil {
			return fmt.Errorf("failed to recalculate point balance: %w", err)
		}
	}

	// Step 4: Delete the sale invoice document
	err = svc.repo.DeleteByGuidfixed(ctx, holdingCode, guid, authUsername)
	if err != nil {
		return err
	}

	go func() {
		svc.repoMq.Delete(findDoc)
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc SaleInvoiceService) DeleteSaleInvoiceByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// Step 1: Find all documents first before deleting
	docs, err := svc.repo.FindByGuids(ctx, holdingCode, GUIDs)
	if err != nil {
		return fmt.Errorf("failed to find documents: %w", err)
	}

	// Step 2: Refund points for each document
	for _, doc := range docs {
		if doc.PointsCode != "" && (doc.GetPoint > 0 || doc.UsePoint > 0) {
			// Process point refund (reverse both GetPoint and UsePoint)
			err := svc.processPointCancelTransactions(ctx, holdingCode, authUsername, doc.CustCode, doc.PointsCode, doc.DocNo, doc.DocDatetime, doc.GetPoint, doc.UsePoint)
			if err != nil {
				return fmt.Errorf("failed to refund points for doc %s: %w", doc.DocNo, err)
			}
		}

		// Delete point transactions for this document
		if doc.DocNo != "" {
			err := svc.pointTransactionRepo.DeletePointTransactionsByDocNo(ctx, holdingCode, doc.DocNo, authUsername)
			if err != nil {
				return fmt.Errorf("failed to delete point transactions for doc %s: %w", doc.DocNo, err)
			}
		}
	}

	// Step 3: Recalculate point balance for all affected customers
	pointsCodeSet := make(map[string]bool)
	for _, doc := range docs {
		if doc.PointsCode != "" {
			pointsCodeSet[doc.PointsCode] = true
		}
	}

	for pointsCode := range pointsCodeSet {
		err := svc.pointTransactionRepo.RecalculatePointBalanceByPointsCode(ctx, holdingCode, pointsCode)
		if err != nil {
			return fmt.Errorf("failed to recalculate point balance for %s: %w", pointsCode, err)
		}
	}

	// Step 4: Delete all documents
	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err = svc.repo.Delete(ctx, holdingCode, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	func() {
		svc.repoMq.DeleteInBatch(docs)
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc SaleInvoiceService) GetLastPOSDocNo(holdingCode, posID, maxDocNo string) (string, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	lastDocNo, err := svc.repo.FindLastPOSDocNo(ctx, holdingCode, posID, maxDocNo)

	if err != nil {
		return "", err
	}

	return lastDocNo, nil
}

func (svc SaleInvoiceService) InfoSaleInvoice(holdingCode string, guid string) (models.SaleInvoiceInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.SaleInvoiceInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.SaleInvoiceInfo{}, errors.New("document not found")
	}

	return findDoc.SaleInvoiceInfo, nil
}

func (svc SaleInvoiceService) InfoSaleInvoiceByCode(holdingCode string, code string) (models.SaleInvoiceInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "docno", code)

	if err != nil {
		return models.SaleInvoiceInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.SaleInvoiceInfo{}, errors.New("document not found")
	}

	return findDoc.SaleInvoiceInfo, nil
}

func (svc SaleInvoiceService) InfoSaleInvoiceByGuidPos(holdingCode string, code string) (models.SaleInvoiceInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "guidpos", code)

	if err != nil {
		return models.SaleInvoiceInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.SaleInvoiceInfo{}, errors.New("document not found")
	}

	return findDoc.SaleInvoiceInfo, nil
}

func (svc SaleInvoiceService) SearchSaleInvoice(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SaleInvoiceInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	if _, ok := filters["createdatetimeafter"]; ok {

		// phase date from YYYY-MM-DD HH:mm:ss to time.Time
		datePhaseFormst := "2006-01-02 15:04:05"
		createDateTimeAfter, err := time.Parse(datePhaseFormst, filters["createdatetimeafter"].(string))
		if err != nil {
			return []models.SaleInvoiceInfo{}, mongopagination.PaginationData{}, err
		}

		// append new filter
		filters["createdat"] = bson.M{"$gt": createDateTimeAfter}

		// remove createdatetimeafter from filters
		delete(filters, "createdatetimeafter")

	}

	if _, ok := filters["updateafterdatetime"]; ok {

		// phase date from YYYY-MM-DD HH:mm:ss to time.Time
		datePhaseFormst := "2006-01-02 15:04:05"
		createDateTimeAfter, err := time.Parse(datePhaseFormst, filters["updateafterdatetime"].(string))
		if err != nil {
			return []models.SaleInvoiceInfo{}, mongopagination.PaginationData{}, err
		}

		// append new filter
		filters["updatedat"] = bson.M{"$gt": createDateTimeAfter}

		// remove createdatetimeafter from filters
		delete(filters, "updateafterdatetime")

	}

	if _, ok := filters["createupdateafterdatetime"]; ok {

		// phase date from YYYY-MM-DD HH:mm:ss to time.Time
		datePhaseFormst := "2006-01-02 15:04:05"
		createDateTimeAfter, err := time.Parse(datePhaseFormst, filters["createupdateafterdatetime"].(string))
		if err != nil {
			return []models.SaleInvoiceInfo{}, mongopagination.PaginationData{}, err
		}

		// append new filter
		filters["$or"] = []bson.M{
			{"createdat": bson.M{"$gt": createDateTimeAfter}},
			{"updatedat": bson.M{"$gt": createDateTimeAfter}},
		}

		// remove createdatetimeafter from filters
		delete(filters, "createupdateafterdatetime")

	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.SaleInvoiceInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc SaleInvoiceService) SearchSaleInvoiceStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SaleInvoiceInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.SaleInvoiceInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc SaleInvoiceService) SaveInBatch(holdingCode string, authUsername string, dataList []models.SaleInvoice) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.SaleInvoice](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.DocNo)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, holdingCode, "docno", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.DocNo)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.SaleInvoice, models.SaleInvoiceDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.SaleInvoice) models.SaleInvoiceDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.SaleInvoiceDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.SaleInvoice = doc

			dataDoc.TransFlag = TRANS_FLAG
			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.SaleInvoice, models.SaleInvoiceDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.SaleInvoiceDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "docno", guid)
		},
		func(doc models.SaleInvoiceDoc) bool {
			return doc.DocNo != ""
		},
		func(holdingCode string, authUsername string, data models.SaleInvoice, doc models.SaleInvoiceDoc) error {

			doc.SaleInvoice = data
			doc.TransFlag = TRANS_FLAG
			doc.UpdatedBy = authUsername
			doc.UpdatedAt = time.Now()

			err = svc.repo.Update(ctx, holdingCode, doc.GuidFixed, doc)
			if err != nil {
				return nil
			}
			return nil
		},
	)

	if len(createDataList) > 0 {
		err = svc.repo.CreateInBatch(ctx, createDataList)

		if err != nil {
			return common.BulkImport{}, err
		}

	}

	createDataKey := []string{}

	for _, doc := range createDataList {
		createDataKey = append(createDataKey, doc.DocNo)
	}

	payloadDuplicateDataKey := []string{}
	for _, doc := range payloadDuplicateList {
		payloadDuplicateDataKey = append(payloadDuplicateDataKey, doc.DocNo)
	}

	updateDataKey := []string{}
	for _, doc := range updateSuccessDataList {

		updateDataKey = append(updateDataKey, doc.DocNo)
	}

	updateFailDataKey := []string{}
	for _, doc := range updateFailDataList {
		updateFailDataKey = append(updateFailDataKey, svc.getDocIDKey(doc))
	}

	svc.saveMasterSync(holdingCode)

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

func (svc SaleInvoiceService) getDocIDKey(doc models.SaleInvoice) string {
	return doc.DocNo
}

func (svc SaleInvoiceService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc SaleInvoiceService) GetModuleName() string {
	return "saleInvoice"
}

func (svc SaleInvoiceService) Export(holdingCode string, languageCode string, languageHeader map[string]string) ([][]string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docs, err := svc.repo.Find(ctx, holdingCode, []string{}, "")

	if err != nil {
		return [][]string{}, err
	}

	keyCols := []string{
		"docdate",        //"วันที่",
		"docno",          //เลขที่เอกสาร",
		"barcode",        //บาร์โค้ด",
		"productname",    //"ชื่อสินค้า",
		"unitcode",       //"หน่วยนับ",
		"unitname",       //"ชื่อหน่วยนับ",
		"qty",            //"จำนวน",
		"price",          //ราคา",
		"discountamount", // "มูลค่าส่วนลด",
		"sumamount",      //"มูลค่าสินค้า",
	}

	headerRow := []string{}
	for _, keyCol := range keyCols {
		tempVal := keyCol
		if val, ok := languageHeader[keyCol]; ok && val != "" {
			tempVal = val
		}
		headerRow = append(headerRow, tempVal)
	}

	results := [][]string{}

	results = append(results, headerRow)

	for _, doc := range docs {
		tempResults := svc.exporter.ParseCSV(languageCode, doc)
		results = append(results, tempResults...)
	}

	return results, nil
}

// processPointCancelTransactions handles point refund when cancelling a sale invoice
func (svc SaleInvoiceService) processPointCancelTransactions(ctx context.Context, holdingCode string, authUsername string, custCode string, pointsCode string, docNo string, docDate time.Time, getPoint float64, usePoint float64) error {
	// Skip processing if no points involved
	if getPoint == 0 && usePoint == 0 {
		return nil
	}

	// Skip processing if pointsCode is empty (required for new point system)
	if pointsCode == "" {
		return nil
	}

	// When cancelling a sale invoice:
	// 1. If points were earned (GetPoint), we need to deduct them back from pointsCode
	// 2. If points were used (UsePoint), we need to refund them back to custCode

	// Handle refund for earned points (deduct points that were earned from original sale)
	// GetPoint cancellation - ใช้ pointsCode
	if getPoint > 0 && pointsCode != "" {
		// Find customer by pointsCode for GetPoint cancellation (allow non-members)
		findCustForGet, err := svc.repoCust.FindByDocIndentityGuid(ctx, holdingCode, "points_code", pointsCode)
		if err != nil {
			return err
		}

		var currentBalanceForGet float64 = 0
		var newBalanceForGet float64 = 0
		customerExists := findCustForGet.Code != ""

		if customerExists {
			currentBalanceForGet = findCustForGet.PointBalance
			newBalanceForGet = currentBalanceForGet - getPoint
		}

		// Create point transaction for cancelling earned points (works for both members and non-members)
		pointTransaction := custModels.PointTransactionDoc{
			PointTransactionData: custModels.PointTransactionData{
				HoldingCodeentity: common.HoldingCodeentity{HoldingCode: holdingCode},
				PointTransactionInfo: custModels.PointTransactionInfo{
					DocIdentity: common.DocIdentity{GuidFixed: utils.NewGUID()},
					PointTransaction: custModels.PointTransaction{
						TransactionDocNo: docNo,
						TransactionDate:  docDate,
						DebtorCode:       custCode,   // Customer who made the transaction
						PointsCode:       pointsCode, // Customer whose points get deducted (GetPoint cancellation)
						TransactionType:  5,          // 5 = cancel (reverse earning)
						PointAmount:      -getPoint,
						BalanceBefore:    currentBalanceForGet,
						BalanceAfter:     newBalanceForGet,
						Description:      fmt.Sprintf("Point cancelled from sale cancellation %s", docNo),
					},
				},
			},
			ActivityDoc: common.ActivityDoc{
				CreatedBy: authUsername,
				CreatedAt: time.Now(),
			},
		}

		_, err = svc.pointTransactionRepo.Create(ctx, pointTransaction)
		if err != nil {
			return err
		}

		// Update balance only if customer exists in debtor table
		if customerExists {
			err = svc.pointTransactionRepo.UpdateDebtorPointBalanceByPointsCode(ctx, holdingCode, pointsCode, -getPoint)
			if err != nil {
				return err
			}
		}
		// Note: For non-members, cancellation transaction is recorded in point_transaction table
		// for proper audit trail even when debtor record doesn't exist
	}

	// Handle refund for used points (refund points that were used in original sale)
	// UsePoint refund - ใช้ custCode (การคืนแต้มจากการใช้ ใช้ custCode)
	if usePoint > 0 && custCode != "" {
		// Find customer by custCode for UsePoint refund
		findCustForUse, err := svc.repoCust.FindByDocIndentityGuid(ctx, holdingCode, "code", custCode)
		if err != nil {
			return err
		}
		if findCustForUse.Code == "" {
			return fmt.Errorf("customer not found with code: %s", custCode)
		}

		currentBalanceForUse := findCustForUse.PointBalance

		// Create point transaction for refunding used points (adding back to balance)
		pointTransaction := custModels.PointTransactionDoc{
			PointTransactionData: custModels.PointTransactionData{
				HoldingCodeentity: common.HoldingCodeentity{HoldingCode: holdingCode},
				PointTransactionInfo: custModels.PointTransactionInfo{
					DocIdentity: common.DocIdentity{GuidFixed: utils.NewGUID()},
					PointTransaction: custModels.PointTransaction{
						TransactionDocNo: docNo,
						TransactionDate:  docDate,
						DebtorCode:       custCode, // Customer who made the transaction
						PointsCode:       custCode, // UsePoint refund ใช้ custCode (การคืนแต้มจากการใช้)
						TransactionType:  6,        // 6 = cancel refund (reverse redemption)
						PointAmount:      usePoint,
						BalanceBefore:    currentBalanceForUse,
						BalanceAfter:     currentBalanceForUse + usePoint,
						Description:      fmt.Sprintf("Point refunded from sale cancellation %s", docNo),
					},
				},
			},
			ActivityDoc: common.ActivityDoc{
				CreatedBy: authUsername,
				CreatedAt: time.Now(),
			},
		}

		_, err = svc.pointTransactionRepo.Create(ctx, pointTransaction)
		if err != nil {
			return err
		}

		// Update balance using custCode for UsePoint refund (การคืนแต้มจากการใช้)
		err = svc.pointTransactionRepo.UpdateDebtorPointBalanceByCode(ctx, holdingCode, custCode, usePoint)
		if err != nil {
			return err
		}
	}

	return nil
}

// PointTransactionConfig holds configuration for point transactions
type PointTransactionConfig struct {
	MaxPointAmountPerTransaction float64
	MaxPointsCodeLength          int
	MinPointAmount               float64
	ContextTimeoutMultiplier     float64 // Multiplier for context timeout in point operations
}

// DefaultPointTransactionConfig returns default configuration for point transactions
func DefaultPointTransactionConfig() PointTransactionConfig {
	return PointTransactionConfig{
		MaxPointAmountPerTransaction: MaxPointAmountPerTransaction,
		MaxPointsCodeLength:          MaxPointsCodeLength,
		MinPointAmount:               MinPointAmount,
		ContextTimeoutMultiplier:     1.5, // 1.5x normal timeout for point operations
	}
}

// validatePointTransaction validates point transaction parameters
func (svc SaleInvoiceService) validatePointTransaction(custCode, pointsCode string, getPoint, usePoint float64) *models.InvalidPointAmountError {
	// Validate point amounts
	if getPoint < 0 {
		return &models.InvalidPointAmountError{
			Amount:    getPoint,
			Operation: "GetPoint",
			Reason:    "negative amounts not allowed",
		}
	}

	if usePoint < 0 {
		return &models.InvalidPointAmountError{
			Amount:    usePoint,
			Operation: "UsePoint",
			Reason:    "negative amounts not allowed",
		}
	}

	// Validate maximum amounts
	if getPoint > svc.pointConfig.MaxPointAmountPerTransaction {
		return &models.InvalidPointAmountError{
			Amount:    getPoint,
			Operation: "GetPoint",
			Reason:    fmt.Sprintf("exceeds maximum allowed amount of %.2f", svc.pointConfig.MaxPointAmountPerTransaction),
		}
	}

	if usePoint > svc.pointConfig.MaxPointAmountPerTransaction {
		return &models.InvalidPointAmountError{
			Amount:    usePoint,
			Operation: "UsePoint",
			Reason:    fmt.Sprintf("exceeds maximum allowed amount of %.2f", svc.pointConfig.MaxPointAmountPerTransaction),
		}
	}

	// Validate minimum amounts (only if > 0)
	if getPoint > 0 && getPoint < svc.pointConfig.MinPointAmount {
		return &models.InvalidPointAmountError{
			Amount:    getPoint,
			Operation: "GetPoint",
			Reason:    fmt.Sprintf("below minimum allowed amount of %.2f", svc.pointConfig.MinPointAmount),
		}
	}

	if usePoint > 0 && usePoint < svc.pointConfig.MinPointAmount {
		return &models.InvalidPointAmountError{
			Amount:    usePoint,
			Operation: "UsePoint",
			Reason:    fmt.Sprintf("below minimum allowed amount of %.2f", svc.pointConfig.MinPointAmount),
		}
	}

	// Validate points code length
	if len(pointsCode) > svc.pointConfig.MaxPointsCodeLength {
		return &models.InvalidPointAmountError{
			Amount:    float64(len(pointsCode)),
			Operation: "PointsCode",
			Reason:    fmt.Sprintf("exceeds maximum length of %d characters", svc.pointConfig.MaxPointsCodeLength),
		}
	}

	return nil
}

// getPointTransactionContext creates a context with extended timeout for point operations
func (svc SaleInvoiceService) getPointTransactionContext() (context.Context, context.CancelFunc) {
	timeout := time.Duration(float64(svc.contextTimeout) * svc.pointConfig.ContextTimeoutMultiplier)
	return context.WithTimeout(context.Background(), timeout)
}

// processUsePointTransaction handles the UsePoint (redemption) operation
func (svc SaleInvoiceService) processUsePointTransaction(ctx context.Context, holdingCode string, authUsername string, custCode string, docNo string, docDate time.Time, usePoint float64) error {
	svc.logger.Debug("Processing UsePoint transaction",
		slog.String("custCode", custCode),
		slog.Float64("usePoint", usePoint),
	)

	// Find customer by custCode for UsePoint (ใช้ custCode เหมือนเดิม)
	findCustForUse, err := svc.repoCust.FindByDocIndentityGuid(ctx, holdingCode, "code", custCode)
	if err != nil {
		return &models.PointTransactionError{
			Operation: "UsePoint",
			CustCode:  custCode,
			Reason:    "failed to find customer",
			Err:       err,
		}
	}

	if findCustForUse.Code == "" {
		return &models.PointTransactionError{
			Operation: "UsePoint",
			CustCode:  custCode,
			Reason:    "customer not found",
			Err:       nil,
		}
	}

	currentBalanceForUse := findCustForUse.PointBalance
	if currentBalanceForUse < usePoint {
		return &models.InsufficientPointsError{
			Current:  currentBalanceForUse,
			Required: usePoint,
			CustCode: custCode,
		}
	}

	// Create point transaction for redemption using custCode (ใช้ custCode เหมือนเดิม)
	pointTransaction := custModels.PointTransactionDoc{
		PointTransactionData: custModels.PointTransactionData{
			HoldingCodeentity: common.HoldingCodeentity{HoldingCode: holdingCode},
			PointTransactionInfo: custModels.PointTransactionInfo{
				DocIdentity: common.DocIdentity{GuidFixed: utils.NewGUID()},
				PointTransaction: custModels.PointTransaction{
					TransactionDocNo: docNo,
					TransactionDate:  docDate,
					DebtorCode:       custCode,                     // Customer who made the transaction
					PointsCode:       custCode,                     // UsePoint ใช้ custCode เหมือนเดิม
					TransactionType:  models.TransactionTypeRedeem, // Use constant instead of hardcoded value
					PointAmount:      -usePoint,
					BalanceBefore:    currentBalanceForUse,
					BalanceAfter:     currentBalanceForUse - usePoint,
					Description:      fmt.Sprintf("Point redeemed from sale %s", docNo),
				},
			},
		},
		ActivityDoc: common.ActivityDoc{
			CreatedBy: authUsername,
			CreatedAt: time.Now(),
		},
	}

	_, err = svc.pointTransactionRepo.Create(ctx, pointTransaction)
	if err != nil {
		return &models.PointTransactionError{
			Operation: "UsePoint",
			CustCode:  custCode,
			Reason:    "failed to create point transaction",
			Err:       err,
		}
	}

	// Update balance using custCode for UsePoint (ใช้ custCode เหมือนเดิม)
	err = svc.pointTransactionRepo.UpdateDebtorPointBalanceByCode(ctx, holdingCode, custCode, -usePoint)
	if err != nil {
		return &models.PointTransactionError{
			Operation: "UsePoint",
			CustCode:  custCode,
			Reason:    "failed to update point balance",
			Err:       err,
		}
	}

	svc.logger.Info("UsePoint transaction completed successfully",
		slog.String("custCode", custCode),
		slog.Float64("usePoint", usePoint),
		slog.Float64("newBalance", currentBalanceForUse-usePoint),
	)

	return nil
}

// processGetPointTransaction handles the GetPoint (earning) operation
func (svc SaleInvoiceService) processGetPointTransaction(ctx context.Context, holdingCode string, authUsername string, custCode string, pointsCode string, docNo string, docDate time.Time, getPoint float64) error {
	svc.logger.Debug("Processing GetPoint transaction",
		slog.String("custCode", custCode),
		slog.String("pointsCode", pointsCode),
		slog.Float64("getPoint", getPoint),
	)

	// Find customer by pointsCode for GetPoint (allow non-members)
	findCustForGet, err := svc.repoCust.FindByDocIndentityGuid(ctx, holdingCode, "points_code", pointsCode)
	if err != nil {
		return &models.PointTransactionError{
			Operation: "GetPoint",
			CustCode:  pointsCode,
			Reason:    "failed to find customer",
			Err:       err,
		}
	}

	var currentBalanceForGet float64 = 0
	var newBalanceForGet float64 = getPoint
	customerExists := findCustForGet.Code != ""

	if customerExists {
		currentBalanceForGet = findCustForGet.PointBalance
		newBalanceForGet = currentBalanceForGet + getPoint
		svc.logger.Debug("Customer exists for GetPoint",
			slog.String("pointsCode", pointsCode),
			slog.Float64("currentBalance", currentBalanceForGet),
		)
	} else {
		svc.logger.Debug("Non-member GetPoint transaction",
			slog.String("pointsCode", pointsCode),
		)
	}

	// Create point transaction for earning using pointsCode (works for both members and non-members)
	pointTransaction := custModels.PointTransactionDoc{
		PointTransactionData: custModels.PointTransactionData{
			HoldingCodeentity: common.HoldingCodeentity{HoldingCode: holdingCode},
			PointTransactionInfo: custModels.PointTransactionInfo{
				DocIdentity: common.DocIdentity{GuidFixed: utils.NewGUID()},
				PointTransaction: custModels.PointTransaction{
					TransactionDocNo: docNo,
					TransactionDate:  docDate,
					DebtorCode:       custCode,                   // Customer who made the transaction
					PointsCode:       pointsCode,                 // GetPoint ใช้ pointsCode (ระบบใหม่)
					TransactionType:  models.TransactionTypeEarn, // Use constant instead of hardcoded value
					PointAmount:      getPoint,
					BalanceBefore:    currentBalanceForGet,
					BalanceAfter:     newBalanceForGet,
					Description:      fmt.Sprintf("Point earned from sale %s", docNo),
				},
			},
		},
		ActivityDoc: common.ActivityDoc{
			CreatedBy: authUsername,
			CreatedAt: time.Now(),
		},
	}

	_, err = svc.pointTransactionRepo.Create(ctx, pointTransaction)
	if err != nil {
		return &models.PointTransactionError{
			Operation: "GetPoint",
			CustCode:  pointsCode,
			Reason:    "failed to create point transaction",
			Err:       err,
		}
	}

	// Update balance only if customer exists in debtor table
	if customerExists {
		err = svc.pointTransactionRepo.UpdateDebtorPointBalanceByPointsCode(ctx, holdingCode, pointsCode, getPoint)
		if err != nil {
			return &models.PointTransactionError{
				Operation: "GetPoint",
				CustCode:  pointsCode,
				Reason:    "failed to update point balance",
				Err:       err,
			}
		}

		svc.logger.Info("GetPoint transaction completed with balance update",
			slog.String("pointsCode", pointsCode),
			slog.Float64("getPoint", getPoint),
			slog.Float64("newBalance", newBalanceForGet),
		)
	} else {
		svc.logger.Info("GetPoint transaction completed for non-member",
			slog.String("pointsCode", pointsCode),
			slog.Float64("getPoint", getPoint),
		)
	}

	return nil
}
