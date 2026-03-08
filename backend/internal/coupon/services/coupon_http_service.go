package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/coupon/models"
	"smlcloudplatform/internal/coupon/repositories"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	productbarcode_repositories "smlcloudplatform/internal/product/productbarcode/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ICouponHttpService interface {
	CreateCoupon(shopID string, authUsername string, doc models.Coupon) (string, error)
	UpdateCoupon(guid string, shopID string, authUsername string, doc models.Coupon) error
	DeleteCoupon(guid string, shopID string, authUsername string) error
	InfoCoupon(guid string, shopID string) (models.CouponInfo, error)
	InfoCouponByCode(couponCode string, shopID string) (models.CouponInfo, error)
	SearchCoupon(shopID string, q string) ([]models.CouponInfo, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.Coupon) (common.BulkImport, error)
	PreviewBulkImport(shopID string, dataList []models.Coupon) (common.BulkPreviewData, error)

	// เพิ่มฟังก์ชันใหม่สำหรับการจัดการการจองคูปอง
	CheckCouponAvailability(couponCode, shopID, customerID string) (*models.CouponAvailabilityResponse, error)
	CheckCouponAvailabilityAdvanced(couponCode, shopID string, req *models.CouponAvailabilityCheckRequest) (*models.CouponAvailabilityAdvancedResponse, error)
	ReserveCoupon(couponCode, shopID string, req models.ReserveCouponRequest) (*models.CouponReservationResponse, error)
	CancelReserveCoupon(couponCode, shopID string, req models.CancelReserveCouponRequest) error
	UseCoupon(couponCode, shopID, authUsername string, req models.UseCouponRequest) (*models.UseCouponResponse, error)
	FindActiveReservationsByCustomerAndCoupon(shopID, customerID, couponID string) ([]models.CouponReservationDoc, error)

	// ฟังก์ชันสำหรับ cleanup expired reservations
	CleanupExpiredReservations(shopID string) error

	// ฟังก์ชันสำหรับคำนวนคูปอง
	CalculateCoupons(shopID string, req models.CalculateCouponRequest) (*models.CalculateCouponResponse, error)

	// ฟังก์ชันสำหรับเช็คสถานะการจอง
	CheckReservationByTransactionID(transactionID, shopID string) (*models.ReservationStatusResponse, error)
	LookupReservationDetails(shopID string, req models.ReservationLookupRequest) (*models.ReservationLookupResponse, error)

	// ฟังก์ชันสำหรับประวัติการใช้คูปอง
	CreateUsageHistory(shopID, authUsername string, req models.CreateUsageHistoryRequest) (string, error)
	GetUsageHistory(shopID string, req models.CouponUsageHistoryRequest) (*models.CouponUsageHistoryResponse, error)
	GetUsageHistoryByCoupon(shopID, couponID string, page, pageSize int) (*models.CouponUsageHistoryResponse, error)
	GetUsageHistoryByCustomer(shopID, customerID string, page, pageSize int) (*models.CouponUsageHistoryResponse, error)
	GetUsageHistoryBySaleInvoice(shopID, saleInvoiceID string) ([]models.CouponUsageHistoryItem, error)
	GetUsageHistoryByTransactionID(shopID, transactionID string) ([]models.CouponUsageHistoryItem, error)
}

type CouponHttpService struct {
	repo                repositories.CouponRepository
	reservationRepo     repositories.CouponReservationRepository
	usageHistoryRepo    repositories.CouponUsageHistoryRepository
	productBarcodeRepo  productbarcode_repositories.IProductBarcodeRepository
	masterSyncCacheRepo mastersync.IMasterSyncCacheRepository
	contextTimeout      time.Duration
}

func NewCouponHttpService(
	repo repositories.CouponRepository,
	reservationRepo repositories.CouponReservationRepository,
	usageHistoryRepo repositories.CouponUsageHistoryRepository,
	productBarcodeRepo productbarcode_repositories.IProductBarcodeRepository,
	masterSyncCacheRepo mastersync.IMasterSyncCacheRepository,
) CouponHttpService {

	contextTimeout := time.Duration(15) * time.Second

	return CouponHttpService{
		repo:                repo,
		reservationRepo:     reservationRepo,
		usageHistoryRepo:    usageHistoryRepo,
		productBarcodeRepo:  productBarcodeRepo,
		masterSyncCacheRepo: masterSyncCacheRepo,
		contextTimeout:      contextTimeout,
	}
}

func (svc CouponHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

// ตรวจสอบเงื่อนไขสินค้า
func (svc CouponHttpService) checkProductCondition(ctx context.Context, shopID, barcode string, condition *models.CouponProductCondition) (bool, string, error) {
	// หาข้อมูลสินค้าจาก barcode
	productInfo, err := svc.productBarcodeRepo.FindByBarcode(ctx, shopID, barcode)
	if err != nil {
		return false, "", fmt.Errorf("error finding product by barcode: %v", err)
	}

	// ตรวจสอบ Product Codes
	if len(condition.ProductCodes) > 0 {
		for _, productCode := range condition.ProductCodes {
			if productInfo.ProductBarcode.ItemCode == productCode {
				return true, "product_code", nil
			}
		}
	}

	// ตรวจสอบ Group Codes
	if len(condition.GroupCodes) > 0 {
		for _, groupCode := range condition.GroupCodes {
			if productInfo.ProductBarcode.GroupCode == groupCode {
				return true, "group_code", nil
			}
		}
	}

	// ตรวจสอบ GroupSubOne Codes
	if len(condition.GroupSubOneCodes) > 0 {
		for _, groupSubOneCode := range condition.GroupSubOneCodes {
			if productInfo.ProductBarcode.GroupsuboneCode == groupSubOneCode {
				return true, "group_subone_code", nil
			}
		}
	}

	// ตรวจสอบ GroupSubTwo Codes
	if len(condition.GroupSubTwoCodes) > 0 {
		for _, groupSubTwoCode := range condition.GroupSubTwoCodes {
			if productInfo.ProductBarcode.GroupsubtwoCode == groupSubTwoCode {
				return true, "group_subtwo_code", nil
			}
		}
	}

	// ตรวจสอบ Brand Codes
	if len(condition.BrandCodes) > 0 {
		for _, brandCode := range condition.BrandCodes {
			if productInfo.ProductBarcode.BrandCode == brandCode {
				return true, "brand_code", nil
			}
		}
	}

	// ตรวจสอบ Design Codes
	if len(condition.DesignCodes) > 0 {
		for _, designCode := range condition.DesignCodes {
			if productInfo.ProductBarcode.DesignCode == designCode {
				return true, "design_code", nil
			}
		}
	}

	// ตรวจสอบ Model Codes
	if len(condition.ModelCodes) > 0 {
		for _, modelCode := range condition.ModelCodes {
			if productInfo.ProductBarcode.ModelCode == modelCode {
				return true, "model_code", nil
			}
		}
	}

	// ตรวจสอบ Pattern Codes
	if len(condition.PatternCodes) > 0 {
		for _, patternCode := range condition.PatternCodes {
			if productInfo.ProductBarcode.PatternCode == patternCode {
				return true, "pattern_code", nil
			}
		}
	}

	// ตรวจสอบ Grade Codes
	if len(condition.GradeCodes) > 0 {
		for _, gradeCode := range condition.GradeCodes {
			if productInfo.ProductBarcode.GradeCode == gradeCode {
				return true, "grade_code", nil
			}
		}
	}

	// ตรวจสอบ Category Codes
	if len(condition.CategoryCodes) > 0 {
		for _, categoryCode := range condition.CategoryCodes {
			if productInfo.ProductBarcode.CategoryCode == categoryCode {
				return true, "category_code", nil
			}
		}
	}

	// ตรวจสอบ Class Codes
	if len(condition.ClassCodes) > 0 {
		for _, classCode := range condition.ClassCodes {
			if productInfo.ProductBarcode.ClassCode == classCode {
				return true, "class_code", nil
			}
		}
	}

	// ถ้าไม่เจอเงื่อนไขไหนเลย
	return false, "", nil
}

func (svc CouponHttpService) getDocIDKey(doc models.Coupon) string {
	return doc.CouponCode
}

func (svc CouponHttpService) CreateCoupon(shopID string, authUsername string, doc models.Coupon) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "couponcode", doc.CouponCode)

	if err != nil {
		return "", err
	}

	if findDoc.CouponCode != "" {
		return "", errors.New("CouponCode is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.CouponDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.Coupon = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now().UTC() // Use UTC time consistently

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	return newGuidFixed, nil
}

func (svc CouponHttpService) UpdateCoupon(guid string, shopID string, authUsername string, doc models.Coupon) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.Coupon = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now().UTC() // Use UTC time consistently

	err = svc.repo.Update(ctx, shopID, guid, findDoc)

	if err != nil {
		return err
	}

	return nil
}

func (svc CouponHttpService) DeleteCoupon(guid string, shopID string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, shopID, guid, authUsername)
	if err != nil {
		return err
	}

	return nil
}

func (svc CouponHttpService) InfoCoupon(guid string, shopID string) (models.CouponInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.CouponInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.CouponInfo{}, errors.New("document not found")
	}

	return findDoc.CouponInfo, nil

}

func (svc CouponHttpService) InfoCouponByCode(couponCode string, shopID string) (models.CouponInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "couponcode", couponCode)

	if err != nil {
		return models.CouponInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.CouponInfo{}, errors.New("coupon not found")
	}

	return findDoc.CouponInfo, nil

}

func (svc CouponHttpService) SearchCoupon(shopID string, q string) ([]models.CouponInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"guidfixed",
		"couponcode",
		"customercode",
	}

	docList, err := svc.repo.Find(ctx, shopID, searchInFields, q)

	if err != nil {
		return []models.CouponInfo{}, err
	}

	return docList, nil
}

func (svc CouponHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.Coupon) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.Coupon](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.CouponCode)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, shopID, "couponcode", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.CouponCode)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.Coupon, models.CouponDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.Coupon) models.CouponDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.CouponDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.Coupon = doc

			currentTime := time.Now().UTC() // Use UTC time consistently
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.Coupon, models.CouponDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.CouponDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "couponcode", guid)
		},
		func(doc models.CouponDoc) bool {
			return doc.CouponCode != ""
		},
		func(shopID string, authUsername string, data models.Coupon, doc models.CouponDoc) error {

			doc.Coupon = data
			doc.UpdatedBy = authUsername
			doc.UpdatedAt = time.Now().UTC() // Use UTC time consistently

			err = svc.repo.Update(ctx, shopID, doc.GuidFixed, doc)
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
		createDataKey = append(createDataKey, doc.CouponCode)
	}

	payloadDuplicateDataKey := []string{}
	for _, doc := range payloadDuplicateList {
		payloadDuplicateDataKey = append(payloadDuplicateDataKey, doc.CouponCode)
	}

	updateDataKey := []string{}
	for _, doc := range updateSuccessDataList {
		updateDataKey = append(updateDataKey, doc.CouponCode)
	}

	updateFailDataKey := []string{}
	for _, doc := range updateFailDataList {
		updateFailDataKey = append(updateFailDataKey, svc.getDocIDKey(doc))
	}

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

// Preview Bulk Import - ตรวจสอบข้อมูลก่อนนำเข้าจริง
func (svc CouponHttpService) PreviewBulkImport(shopID string, dataList []models.Coupon) (common.BulkPreviewData, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// กรองข้อมูลซ้ำใน payload
	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.Coupon](dataList, svc.getDocIDKey)

	// สร้างรายการ CouponCode สำหรับตรวจสอบ
	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.CouponCode)
	}

	// ค้นหา CouponCode ที่มีอยู่ในระบบ
	findItemGuid, err := svc.repo.FindInItemGuid(ctx, shopID, "couponcode", itemCodeGuidList)
	if err != nil {
		return common.BulkPreviewData{}, err
	}

	// สร้างรายการ CouponCode ที่พบในระบบ
	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.CouponCode)
	}

	// แยกข้อมูลเป็น Create/Update
	willCreate := []string{}
	willUpdate := []string{}

	foundItemGuidMap := make(map[string]bool)
	for _, guid := range foundItemGuidList {
		foundItemGuidMap[guid] = true
	}

	for _, doc := range payloadList {
		couponCode := svc.getDocIDKey(doc)
		if foundItemGuidMap[couponCode] {
			willUpdate = append(willUpdate, couponCode)
		} else {
			willCreate = append(willCreate, couponCode)
		}
	}

	// สร้างรายการ PayloadDuplicate
	payloadDuplicateDataKey := []string{}
	for _, doc := range payloadDuplicateList {
		payloadDuplicateDataKey = append(payloadDuplicateDataKey, doc.CouponCode)
	}

	return common.BulkPreviewData{
		WillCreate:       willCreate,
		WillUpdate:       willUpdate,
		PayloadDuplicate: payloadDuplicateDataKey,
		ValidationErrors: []string{}, // จะถูกเพิ่มใน controller
	}, nil
}

// ตรวจสอบความพร้อมใช้งานของคูปอง
func (svc CouponHttpService) CheckCouponAvailability(couponCode, shopID, customerID string) (*models.CouponAvailabilityResponse, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// หาข้อมูลคูปองจาก CouponCode
	coupon, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "couponcode", couponCode)
	if err != nil {
		return nil, err
	}

	if coupon.CouponCode == "" {
		return nil, errors.New("coupon not found")
	}

	// ตรวจสอบสิทธิ์ของลูกค้า: ใช้ IsCustomerEligible แทน
	if !coupon.Coupon.IsCustomerEligible(customerID) {
		response := &models.CouponAvailabilityResponse{
			Available:      false,
			UsageCount:     0,
			MaxUsageCount:  0,
			RemainingUsage: 0,
			Status:         int8(coupon.Coupon.Status),
			IsExpired:      coupon.Coupon.IsExpired(),
			Message:        "คุณไม่มีสิทธิ์ใช้คูปองนี้",
		}
		return response, nil
	}

	// Dynamic Usage Count Logic
	var totalUsage int
	var maxUsage int
	var activeReservations int

	if coupon.Coupon.IsGlobalUsageMode() {
		// กรณีที่ 1: customers = [] (ไม่กำหนดลูกค้า)
		// ใช้ MaxUsageCount = คูปองใช้ได้รวมทั้งหมด X ครั้ง
		// ใครก็ใช้ได้ จนกว่าจะครบจำนวนรวม

		// นับจำนวนครั้งที่ใช้ไปแล้วทั้งหมด
		searchReq := models.CouponUsageHistoryRequest{
			CouponCode: coupon.Coupon.CouponCode,
			Page:       1,
			PageSize:   10000, // เอาทั้งหมดเพื่อนับ
		}
		_, total, err := svc.usageHistoryRepo.SearchUsageHistory(ctx, shopID, searchReq)
		if err != nil {
			return nil, err
		}
		totalUsage = total
		maxUsage = coupon.Coupon.GetMaxUsageLimit()

		// นับจำนวน active reservations ทั้งหมด
		activeReservations, err = svc.reservationRepo.CountActiveReservations(ctx, shopID, coupon.GuidFixed)
		if err != nil {
			return nil, err
		}

	} else {
		// กรณีที่ 2: customers = ["A", "B", "C"] (กำหนดลูกค้า)
		// ใช้ MaxUsageCountPerCustomer = แต่ละลูกค้าใช้ได้ Y ครั้ง
		// เฉพาะลูกค้าที่กำหนดเท่านั้น และแต่ละคนใช้ได้ไม่เกิน Y ครั้ง

		// นับจำนวนครั้งที่ลูกค้าคนนี้ใช้ไปแล้ว
		searchReq := models.CouponUsageHistoryRequest{
			CouponCode: coupon.Coupon.CouponCode,
			CustomerID: customerID,
			Page:       1,
			PageSize:   10000, // เอาทั้งหมดเพื่อนับ
		}
		_, customerTotal, err := svc.usageHistoryRepo.SearchUsageHistory(ctx, shopID, searchReq)
		if err != nil {
			return nil, err
		}
		totalUsage = customerTotal
		maxUsage = coupon.Coupon.GetMaxUsageLimit()

		// นับจำนวน active reservations ของลูกค้าคนนี้
		activeReservations, err = svc.reservationRepo.CountActiveReservationsByCustomer(ctx, shopID, coupon.GuidFixed, customerID)
		if err != nil {
			return nil, err
		}
	}

	// คำนวณจำนวนครั้งที่ใช้ได้อีก (ต้องพิจารณา active reservations ด้วย)
	remainingUsage := maxUsage - totalUsage - activeReservations
	if remainingUsage < 0 {
		remainingUsage = 0
	}

	// Debug: ตรวจสอบวันที่
	now := time.Now().UTC()
	expiryDate := coupon.Coupon.ExpiryDate.UTC()
	isExpired := coupon.Coupon.IsExpired()
	fmt.Printf("DEBUG: Current time: %s\n", now.Format("2006-01-02 15:04:05 UTC"))
	fmt.Printf("DEBUG: Expiry date: %s\n", expiryDate.Format("2006-01-02 15:04:05 UTC"))
	fmt.Printf("DEBUG: IsExpired: %t\n", isExpired)
	fmt.Printf("DEBUG: After comparison: %t\n", now.After(expiryDate))

	response := &models.CouponAvailabilityResponse{
		Available:      coupon.Coupon.IsActive() && !coupon.Coupon.IsExpired(),
		UsageCount:     totalUsage,
		MaxUsageCount:  maxUsage,
		RemainingUsage: remainingUsage,
		Status:         int8(coupon.Coupon.Status),
		IsExpired:      coupon.Coupon.IsExpired(),
	}

	// ตรวจสอบการใช้งานคูปอง - สำหรับคูปองที่ใช้ได้ครั้งเดียว
	if coupon.Coupon.IsOneTimeUse && (totalUsage > 0 || activeReservations > 0) {
		response.Available = false
		if totalUsage > 0 {
			response.Message = "คูปองนี้ใช้ได้เพียงครั้งเดียวและถูกใช้ไปแล้ว"
		} else {
			response.Message = "คูปองนี้ใช้ได้เพียงครั้งเดียวและถูกจองแล้ว"
		}
		return response, nil
	}

	// ตรวจสอบจำนวนครั้งสูงสุด (รวม active reservations)
	if maxUsage > 0 && (totalUsage+activeReservations) >= maxUsage {
		response.Available = false
		if coupon.Coupon.IsGlobalUsageMode() {
			response.Message = "คูปองถูกใช้หรือจองครบจำนวนครั้งสูงสุดแล้ว (รวมทั้งหมด)"
		} else {
			response.Message = "คุณใช้หรือจองคูปองครบจำนวนครั้งสูงสุดต่อลูกค้าแล้ว"
		}
		return response, nil
	}

	if !response.Available {
		if coupon.Coupon.IsExpired() {
			response.Message = "คูปองหมดอายุแล้ว"
		} else if coupon.Coupon.Status == models.CouponStatusCanceled {
			response.Message = "คูปองถูกยกเลิกแล้ว"
		} else {
			response.Message = "คูปองไม่สามารถใช้งานได้"
		}
	}

	return response, nil
}

// ตรวจสอบความพร้อมใช้งานของคูปองแบบขั้นสูง (พร้อมตรวจสอบสินค้าและสาขา)
func (svc CouponHttpService) CheckCouponAvailabilityAdvanced(couponCode, shopID string, req *models.CouponAvailabilityCheckRequest) (*models.CouponAvailabilityAdvancedResponse, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// 1. ตรวจสอบคูปองพื้นฐาน
	basicAvailability, err := svc.CheckCouponAvailability(couponCode, shopID, req.CustomerID)
	if err != nil {
		return nil, err
	}

	// 2. หาข้อมูลคูปองจาก CouponCode
	coupon, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "couponcode", couponCode)
	if err != nil {
		return nil, err
	}

	if coupon.CouponCode == "" {
		return nil, errors.New("coupon not found")
	}

	// 3. ตรวจสอบสาขา (IgnoreBranchCode)
	branchAllowed := true
	if len(coupon.Coupon.IgnoreBranchCode) > 0 {
		for _, ignoredBranch := range coupon.Coupon.IgnoreBranchCode {
			if ignoredBranch == req.BranchCode {
				branchAllowed = false
				break
			}
		}
	}

	// 4. คำนวณยอดรวมทั้งหมด
	totalAmount := float64(0)
	for _, item := range req.Items {
		totalAmount += item.SumAmount
	}

	// 5. ตรวจสอบสินค้าแต่ละรายการ
	itemResults := make([]models.CouponItemCheckResult, len(req.Items))
	eligibleAmount := float64(0)
	eligibleItemCount := 0

	for i, item := range req.Items {
		itemResult := models.CouponItemCheckResult{
			Barcode:   item.Barcode,
			Qty:       item.Qty,
			Price:     item.Price,
			SumAmount: item.SumAmount,
		}

		// ตรวจสอบเงื่อนไขสินค้า
		if coupon.Coupon.ProductCondition != nil {
			isEligible, matchedCategory, err := svc.checkProductCondition(ctx, shopID, item.Barcode, coupon.Coupon.ProductCondition)
			if err != nil {
				itemResult.Message = fmt.Sprintf("Error checking product condition: %v", err)
			} else {
				itemResult.IsEligible = isEligible
				itemResult.MatchedCategory = matchedCategory

				if isEligible {
					eligibleAmount += item.SumAmount
					eligibleItemCount++
				}
			}
		} else {
			// ไม่มีเงื่อนไขสินค้า = ทุกสินค้าเข้าเงื่อนไข
			itemResult.IsEligible = true
			eligibleAmount += item.SumAmount
			eligibleItemCount++
		}

		itemResults[i] = itemResult
	}

	// 5.1. ตรวจสอบมูลค่าขั้นต่ำ (Minimum Amount)
	minimumAmountMet := true
	if coupon.Coupon.ProductCondition != nil && coupon.Coupon.ProductCondition.MinimumAmount > 0 {
		if eligibleAmount < coupon.Coupon.ProductCondition.MinimumAmount {
			minimumAmountMet = false
		}
	}

	// 6. คำนวณส่วนลด
	totalDiscount := float64(0)
	if basicAvailability.Available && branchAllowed && eligibleAmount > 0 && minimumAmountMet {
		switch coupon.Coupon.CouponType {
		case models.CouponTypeValueDiscount:
			// ลดตามมูลค่า
			totalDiscount = coupon.Coupon.CouponValue
		case models.CouponTypePercentDiscount:
			// ลดตามเปอร์เซ็นต์ (คำนวณจากยอดที่เข้าเงื่อนไขเท่านั้น)
			totalDiscount = eligibleAmount * (coupon.Coupon.CouponValue / 100)
		case models.CouponTypeCashVoucher:
			// คูปองแทนเงินสด
			totalDiscount = coupon.Coupon.CouponValue
		}

		// แจกส่วนลดให้กับสินค้าที่เข้าเงื่อนไข
		if eligibleItemCount > 0 {
			for i := range itemResults {
				if itemResults[i].IsEligible {
					// แจกส่วนลดตามสัดส่วนของยอดเงิน
					discountRatio := itemResults[i].SumAmount / eligibleAmount
					itemResults[i].DiscountAmount = totalDiscount * discountRatio
				}
			}
		}
	}

	// 7. สร้าง response (ไม่ใช้ basicAvailability.Available เพราะมี minimum amount check)
	isBasicallyAvailable := coupon.Coupon.IsActive() && !coupon.Coupon.IsExpired()
	response := &models.CouponAvailabilityAdvancedResponse{
		Available:         isBasicallyAvailable && branchAllowed && minimumAmountMet,
		UsageCount:        basicAvailability.UsageCount,
		MaxUsageCount:     basicAvailability.MaxUsageCount,
		RemainingUsage:    basicAvailability.RemainingUsage,
		Status:            basicAvailability.Status,
		IsExpired:         basicAvailability.IsExpired,
		BranchAllowed:     branchAllowed,
		TotalAmount:       totalAmount,
		EligibleAmount:    eligibleAmount,
		TotalDiscount:     totalDiscount,
		EligibleItemCount: eligibleItemCount,
		ItemResults:       itemResults,
	}

	// 8. กำหนดข้อความ
	if !isBasicallyAvailable {
		if coupon.Coupon.IsExpired() {
			response.Message = "คูปองหมดอายุแล้ว"
		} else if coupon.Coupon.Status == models.CouponStatusCanceled {
			response.Message = "คูปองถูกยกเลิกแล้ว"
		} else {
			response.Message = "คูปองไม่สามารถใช้งานได้"
		}
	} else if !branchAllowed {
		response.Message = "ไม่สามารถใช้คูปองในสาขานี้ได้"
	} else if !minimumAmountMet {
		response.Message = fmt.Sprintf("ยอดสินค้าที่เข้าเงื่อนไข %.2f บาท ต้องไม่น้อยกว่า %.2f บาท", eligibleAmount, coupon.Coupon.ProductCondition.MinimumAmount)
	} else if eligibleItemCount == 0 && coupon.Coupon.ProductCondition != nil {
		response.Message = "ไม่มีสินค้าที่เข้าเงื่อนไขการใช้คูปอง"
	} else if response.Available {
		response.Message = "คูปองสามารถใช้งานได้"
	}

	return response, nil
}

// จองคูปอง
func (svc CouponHttpService) ReserveCoupon(couponCode, shopID string, req models.ReserveCouponRequest) (*models.CouponReservationResponse, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// 1. ตรวจสอบคูปองจาก CouponCode
	coupon, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "couponcode", couponCode)
	if err != nil {
		return nil, err
	}

	if coupon.CouponCode == "" {
		return nil, errors.New("coupon not found")
	}

	// 2. ตรวจสอบว่าสามารถใช้งานได้หรือไม่
	canUse, message := coupon.Coupon.CanUse()
	if !canUse {
		return nil, errors.New("ไม่สามารถจองคูปองได้: " + message)
	}

	// 3. ตรวจสอบจำนวนครั้งการใช้งานและการจองที่ยังใช้งานได้
	var totalUsage int
	var activeReservations int

	if coupon.Coupon.IsGlobalUsageMode() {
		// นับจำนวนครั้งที่ใช้ไปแล้วทั้งหมด
		searchReq := models.CouponUsageHistoryRequest{
			CouponCode: coupon.Coupon.CouponCode,
			Page:       1,
			PageSize:   10000,
		}
		_, total, err := svc.usageHistoryRepo.SearchUsageHistory(ctx, shopID, searchReq)
		if err != nil {
			return nil, err
		}
		totalUsage = total

		// นับจำนวน active reservations ทั้งหมด
		activeReservations, err = svc.reservationRepo.CountActiveReservations(ctx, shopID, coupon.GuidFixed)
		if err != nil {
			return nil, err
		}
	} else {
		// นับจำนวนครั้งที่ลูกค้าคนนี้ใช้ไปแล้ว
		searchReq := models.CouponUsageHistoryRequest{
			CouponCode: coupon.Coupon.CouponCode,
			CustomerID: req.CustomerID,
			Page:       1,
			PageSize:   10000,
		}
		_, customerTotal, err := svc.usageHistoryRepo.SearchUsageHistory(ctx, shopID, searchReq)
		if err != nil {
			return nil, err
		}
		totalUsage = customerTotal

		// นับจำนวน active reservations ของลูกค้าคนนี้
		activeReservations, err = svc.reservationRepo.CountActiveReservationsByCustomer(ctx, shopID, coupon.GuidFixed, req.CustomerID)
		if err != nil {
			return nil, err
		}
	}

	maxUsage := coupon.Coupon.GetMaxUsageLimit()

	// ตรวจสอบคูปองใช้ได้ครั้งเดียว
	if coupon.Coupon.IsOneTimeUse && (totalUsage > 0 || activeReservations > 0) {
		if totalUsage > 0 {
			return nil, errors.New("ไม่สามารถจองได้ เนื่องจากคูปองนี้ใช้ได้เพียงครั้งเดียวและถูกใช้แล้ว")
		} else {
			return nil, errors.New("ไม่สามารถจองได้ เนื่องจากคูปองนี้ใช้ได้เพียงครั้งเดียวและถูกจองแล้ว")
		}
	}

	// ตรวจสอบจำนวนครั้งสูงสุด (รวม active reservations)
	if maxUsage > 0 && (totalUsage+activeReservations) >= maxUsage {
		if coupon.Coupon.IsGlobalUsageMode() {
			return nil, errors.New("ไม่สามารถจองได้ เนื่องจากคูปองถูกใช้หรือจองครบจำนวนครั้งสูงสุดแล้ว (รวมทั้งหมด)")
		} else {
			return nil, errors.New("ไม่สามารถจองได้ เนื่องจากคุณใช้หรือจองคูปองครบจำนวนครั้งสูงสุดต่อลูกค้าแล้ว")
		}
	}

	// 6. สร้างการจองใหม่
	reservation := models.NewCouponReservation(coupon.GuidFixed, req.CustomerID, req.TransactionID)

	reservationDoc := models.CouponReservationDoc{
		CouponReservationData: models.CouponReservationData{
			CouponReservation: *reservation,
		},
	}

	// 7. บันทึกการจองลงฐานข้อมูล
	reservationID, err := svc.reservationRepo.CreateReservation(ctx, shopID, reservationDoc)
	if err != nil {
		return nil, err
	}

	return &models.CouponReservationResponse{
		ReservationID: reservationID,
		CouponID:      coupon.GuidFixed,
		CouponCode:    coupon.Coupon.CouponCode,
		TransactionID: req.TransactionID,
		ExpiresAt:     reservation.ExpiresAt,
		Reserved:      true,
		Message:       "จองคูปองสำเร็จ จะหมดอายุใน 15 นาที",
	}, nil
}

// ยกเลิกการจองคูปอง
func (svc CouponHttpService) CancelReserveCoupon(couponCode, shopID string, req models.CancelReserveCouponRequest) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// ยกเลิกการจองในฐานข้อมูล (ไม่ต้องตรวจสอบคูปองเพราะใช้ ReservationID)
	err := svc.reservationRepo.CancelReservation(ctx, shopID, req.ReservationID, req.CustomerID)
	if err != nil {
		return err
	}

	return nil
}

// ใช้คูปอง
func (svc CouponHttpService) UseCoupon(couponCode, shopID, authUsername string, req models.UseCouponRequest) (*models.UseCouponResponse, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// 1. ตรวจสอบการจอง
	reservation, err := svc.reservationRepo.FindReservationByID(ctx, shopID, req.ReservationID)
	if err != nil {
		return nil, err
	}

	if reservation.CustomerID != req.CustomerID {
		return nil, errors.New("การจองไม่ตรงกับลูกค้า")
	}

	if !reservation.CouponReservation.IsActive() {
		return nil, errors.New("การจองหมดอายุหรือถูกยกเลิกแล้ว")
	}

	// 2. ตรวจสอบคูปองจาก CouponCode
	coupon, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "couponcode", couponCode)
	if err != nil {
		return nil, err
	}

	if coupon.CouponCode == "" {
		return nil, errors.New("coupon not found")
	}

	// 3. ตรวจสอบว่าสามารถใช้งานได้หรือไม่
	canUse, message := coupon.Coupon.CanUse()
	if !canUse {
		return nil, errors.New("ไม่สามารถใช้คูปองได้: " + message)
	}

	// 4. ตรวจสอบสิทธิ์ลูกค้า
	if !coupon.Coupon.IsCustomerEligible(req.CustomerID) {
		return nil, errors.New("คุณไม่มีสิทธิ์ใช้คูปองนี้")
	}

	// 5. ตรวจสอบสิทธิ์ลูกค้า
	if !coupon.Coupon.IsCustomerEligible(req.CustomerID) {
		return nil, errors.New("คุณไม่มีสิทธิ์ใช้คูปองนี้")
	}

	// 6. หมายเหตุ: การตรวจสอบ Product Condition, Branch Exclusion และ Minimum Amount
	// ควรได้รับการตรวจสอบแล้วจาก CheckCouponAvailabilityAdvanced ก่อนการจอง
	// UseCoupon จึงเน้นการตรวจสอบเฉพาะสิทธิ์และการจอง

	// // 4. ตรวจสอบว่าจำนวนที่ใช้ไม่เกินที่จองไว้
	// if req.UseAmount > reservation.ReservedAmount {
	// 	return nil, errors.New("จำนวนที่ใช้เกินจำนวนที่จองไว้")
	// }

	// 5. คำนวณจำนวนส่วนลด
	var discountAmount float64

	if coupon.Coupon.CouponType == models.CouponTypeValueDiscount {
		// ส่วนลดตามมูลค่า
		discountAmount = req.UseAmount
	} else if coupon.Coupon.CouponType == models.CouponTypePercentDiscount {
		// ส่วนลดตามเปอร์เซ็นต์
		discountAmount = req.UseAmount
		// discountAmount = req.UseAmount * (coupon.Coupon.CouponValue / 100)
		// if discountAmount > coupon.Coupon.RemainingValue {
		// 	discountAmount = coupon.Coupon.RemainingValue
		// }
	}

	// 6. ตรวจสอบจำนวนครั้งการใช้งาน
	var totalUsage int

	if coupon.Coupon.IsGlobalUsageMode() {
		// นับจำนวนครั้งที่ใช้ไปแล้วทั้งหมด
		searchReq := models.CouponUsageHistoryRequest{
			CouponCode: coupon.Coupon.CouponCode,
			Page:       1,
			PageSize:   10000,
		}
		_, total, err := svc.usageHistoryRepo.SearchUsageHistory(ctx, shopID, searchReq)
		if err != nil {
			return nil, err
		}
		totalUsage = total
	} else {
		// นับจำนวนครั้งที่ลูกค้าคนนี้ใช้ไปแล้ว
		searchReq := models.CouponUsageHistoryRequest{
			CouponCode: coupon.Coupon.CouponCode,
			CustomerID: req.CustomerID,
			Page:       1,
			PageSize:   10000,
		}
		_, customerTotal, err := svc.usageHistoryRepo.SearchUsageHistory(ctx, shopID, searchReq)
		if err != nil {
			return nil, err
		}
		totalUsage = customerTotal
	}

	maxUsage := coupon.Coupon.GetMaxUsageLimit()

	// ตรวจสอบจำนวนครั้งสูงสุด
	if maxUsage > 0 && totalUsage >= maxUsage {
		if coupon.Coupon.IsGlobalUsageMode() {
			return nil, errors.New("คูปองถูกใช้ครบจำนวนครั้งสูงสุดแล้ว (รวมทั้งหมด)")
		} else {
			return nil, errors.New("คุณใช้คูปองครบจำนวนครั้งสูงสุดต่อลูกค้าแล้ว")
		}
	}

	// 7. ไม่ต้องอัพเดท RemainingValue อีกต่อไป

	// 8. อัพเดทสถานะการจอง
	err = svc.reservationRepo.UpdateReservationStatus(ctx, shopID, req.ReservationID, models.ReservationStatusUsed)
	if err != nil {
		return nil, err
	}

	// 9. สร้างประวัติการใช้คูปอง
	var cashVoucherAmount float64
	if coupon.Coupon.CouponType == models.CouponTypeCashVoucher {
		cashVoucherAmount = req.UseAmount
	}

	historyDoc := models.CouponUsageHistoryDoc{
		CouponUsageHistoryData: models.CouponUsageHistoryData{
			CouponUsageHistory: models.CouponUsageHistory{
				CouponID:          coupon.GuidFixed,
				CouponCode:        coupon.Coupon.CouponCode,
				CustomerID:        req.CustomerID,
				CustomerCode:      req.CustomerCode,
				CustomerName:      req.CustomerName,
				SaleInvoiceID:     req.SaleInvoiceID,
				SaleInvoiceNumber: req.SaleInvoiceNumber,
				TransactionID:     req.TransactionID,
				ReservationID:     req.ReservationID,
				UsedAmount:        req.UseAmount,
				DiscountAmount:    discountAmount,
				CashVoucherAmount: cashVoucherAmount,
				OrderAmount:       req.OrderAmount,
				CouponType:        coupon.Coupon.CouponType,
				CouponValue:       coupon.Coupon.CouponValue,
				UsedAt:            time.Now().UTC(),
				UsedBy:            authUsername,
				Remark:            req.Remark,
			},
		},
	}

	// บันทึกประวัติการใช้ (ไม่ให้ error หยุดการทำงาน)
	if _, historyErr := svc.usageHistoryRepo.CreateUsageHistory(ctx, shopID, historyDoc); historyErr != nil {
		// Log error but don't fail the transaction
		// TODO: Add proper logging
		_ = historyErr // Acknowledge we're intentionally ignoring this error
	}

	return &models.UseCouponResponse{
		Used:           true,
		DiscountAmount: discountAmount,
		TransactionID:  req.TransactionID,
		UsageCount:     totalUsage + 1,
		RemainingUsage: maxUsage - (totalUsage + 1),
		Message:        "ใช้คูปองสำเร็จ",
	}, nil
}

// FindActiveReservationsByCustomerAndCoupon ค้นหาการจองที่ยังใช้งานได้ของลูกค้าและคูปองเฉพาะ
func (svc CouponHttpService) FindActiveReservationsByCustomerAndCoupon(shopID, customerID, couponID string) ([]models.CouponReservationDoc, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	reservations, err := svc.reservationRepo.FindActiveReservationsByCustomerAndCoupon(ctx, shopID, customerID, couponID)
	if err != nil {
		return nil, err
	}

	return reservations, nil
}

// ล้างการจองที่หมดอายุ
func (svc CouponHttpService) CleanupExpiredReservations(shopID string) error {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.reservationRepo.CleanupExpiredReservations(ctx, shopID)
	if err != nil {
		return err
	}

	return nil
}

// คำนวนส่วนลดจากคูปอง
func (svc CouponHttpService) CalculateCoupons(shopID string, req models.CalculateCouponRequest) (*models.CalculateCouponResponse, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	var totalDiscount float64 = 0
	var totalCashVoucher float64 = 0
	var couponResults []models.CalculateCouponItemResponse
	var errors []models.CouponCalculationError

	// ใช้ cascading calculation - แต่ละคูปองคำนวนจากยอดที่เหลือหลังจากส่วนลดก่อนหน้า
	currentOrderAmount := req.OrderAmount
	currentItems := make([]models.CouponCheckItem, len(req.Items))
	copy(currentItems, req.Items)

	// วนลูปคำนวนแต่ละคูปอง
	for _, couponReq := range req.Coupons {
		result, err := svc.calculateSingleCouponAdvanced(ctx, shopID, currentOrderAmount, currentItems, req.BranchCode, couponReq, req.CustomerID)
		if err != nil {
			// เก็บ error แต่ยังคำนวนคูปองอื่นต่อ
			errors = append(errors, models.CouponCalculationError{
				CouponCode: couponReq.CouponCode,
				Error:      err.Error(),
				Code:       "CALCULATION_ERROR",
			})
			// เพิ่ม result ที่ไม่สามารถใช้ได้
			couponResults = append(couponResults, models.CalculateCouponItemResponse{
				CouponCode:        couponReq.CouponCode,
				CouponType:        0,
				CouponTypeName:    "ไม่ทราบ",
				DiscountAmount:    0,
				CashVoucherAmount: 0,
				UsedAmount:        0,
				UsageCount:        0,
				RemainingUsage:    0,
				Applied:           false,
				BranchAllowed:     false,
				EligibleAmount:    0,
				EligibleItemCount: 0,
				MinimumAmount:     0,
				Message:           err.Error(),
			})
			continue
		}

		couponResults = append(couponResults, *result)

		// สะสมยอดส่วนลดและเงินสด
		if result.Applied {
			totalDiscount += result.DiscountAmount
			totalCashVoucher += result.CashVoucherAmount

			// อัปเดตยอดคำสั่งซื้อปัจจุบันสำหรับคูปองตัวถัดไป (cascading)
			// เฉพาะคูปองลดราคาเท่านั้นที่ลดยอดรวม คูปองแทนเงินสดไม่ลด
			if result.CouponType != models.CouponTypeCashVoucher {
				currentOrderAmount = currentOrderAmount - result.DiscountAmount
				// ตรวจสอบไม่ให้ยอดติดลบ
				if currentOrderAmount < 0 {
					currentOrderAmount = 0
				}
			}

			// อัปเดตยอดสินค้าที่เหลือหลังจากใช้คูปอง (สำหรับ cascading calculation)
			svc.updateItemAmountsAfterCouponApplied(currentItems, result)
		}
	}

	// คำนวนยอดสุทธิ
	finalAmount := req.OrderAmount - totalDiscount

	return &models.CalculateCouponResponse{
		Success:          len(errors) == 0,
		TotalDiscount:    totalDiscount,
		TotalCashVoucher: totalCashVoucher,
		FinalAmount:      finalAmount,
		CouponResults:    couponResults,
		Errors:           errors,
	}, nil
}

// คำนวนคูปองแต่ละตัว
func (svc CouponHttpService) calculateSingleCoupon(ctx context.Context, shopID string, orderAmount float64, couponReq models.CalculateCouponItemRequest, customerID string) (*models.CalculateCouponItemResponse, error) {
	// ค้นหาคูปอง
	coupon, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "couponcode", couponReq.CouponCode)
	if err != nil {
		return nil, err
	}

	if coupon.CouponCode == "" {
		return nil, errors.New("coupon not found")
	}

	// ตรวจสอบสถานะคูปอง
	canUse, message := coupon.Coupon.CanUse()
	if !canUse {
		return &models.CalculateCouponItemResponse{
			CouponCode:        couponReq.CouponCode,
			CouponType:        coupon.Coupon.CouponType,
			CouponTypeName:    coupon.Coupon.CouponType.String(),
			DiscountAmount:    0,
			CashVoucherAmount: 0,
			UsedAmount:        0,
			UsageCount:        0,
			RemainingUsage:    0,
			Applied:           false,
			Message:           message,
		}, nil
	}

	// ตรวจสอบสิทธิ์ลูกค้า
	if !coupon.Coupon.IsCustomerEligible(customerID) {
		return &models.CalculateCouponItemResponse{
			CouponCode:        couponReq.CouponCode,
			CouponType:        coupon.Coupon.CouponType,
			CouponTypeName:    coupon.Coupon.CouponType.String(),
			DiscountAmount:    0,
			CashVoucherAmount: 0,
			UsedAmount:        0,
			UsageCount:        0,
			RemainingUsage:    0,
			Applied:           false,
			Message:           "คุณไม่มีสิทธิ์ใช้คูปองนี้",
		}, nil
	}

	// คำนวนจำนวนที่ใช้จริง
	useAmount := couponReq.UseAmount

	// คำนวนส่วนลดตามประเภทคูปอง
	var discountAmount float64 = 0
	var cashVoucherAmount float64 = 0

	switch coupon.Coupon.CouponType {
	case models.CouponTypeValueDiscount:
		// ลดตามมูลค่า - ใช้ useAmount ที่ส่งมา หรือ CouponValue ถ้าไม่ระบุ
		if useAmount <= 0 {
			useAmount = coupon.Coupon.CouponValue
		}
		// ไม่ต้องเช็ค RemainingValue อีกต่อไป
		discountAmount = useAmount
		if discountAmount > orderAmount {
			discountAmount = orderAmount
		}

	case models.CouponTypePercentDiscount:
		// ลดตามเปอร์เซ็นต์
		// ใช้ UseAmount เป็นเปอร์เซ็นต์ที่ต้องการใช้
		percentageToUse := couponReq.UseAmount
		if percentageToUse <= 0 || percentageToUse > coupon.Coupon.CouponValue {
			percentageToUse = coupon.Coupon.CouponValue
		}
		discountAmount = (orderAmount * percentageToUse) / 100
		// ไม่จำกัดด้วย useAmount สำหรับคูปองเปอร์เซ็นต์

	case models.CouponTypeCashVoucher:
		// คูปองแทนเงินสด - ใช้ useAmount ที่ส่งมา หรือ CouponValue ถ้าไม่ระบุ
		if useAmount <= 0 {
			useAmount = coupon.Coupon.CouponValue
		}
		// ไม่ต้องเช็ค RemainingValue อีกต่อไป
		cashVoucherAmount = useAmount

		// ตรวจสอบเงื่อนไขขั้นต่ำจากคูปอง (ถ้ามี ProductCondition)
		// หากไม่มี MinimumAmount ให้ใช้ค่าเริ่มต้น 0 (ไม่จำกัด)
		minOrderAmount := float64(0)
		if coupon.Coupon.ProductCondition != nil && coupon.Coupon.ProductCondition.MinimumAmount > 0 {
			minOrderAmount = coupon.Coupon.ProductCondition.MinimumAmount
		}

		// ตรวจสอบเงื่อนไขขั้นต่ำ (ถ้ามีกำหนด)
		if minOrderAmount > 0 && orderAmount < minOrderAmount {
			return &models.CalculateCouponItemResponse{
				CouponCode:        couponReq.CouponCode,
				CouponType:        coupon.Coupon.CouponType,
				CouponTypeName:    coupon.Coupon.CouponType.String(),
				DiscountAmount:    0,
				CashVoucherAmount: 0,
				UsedAmount:        0,
				UsageCount:        0,
				RemainingUsage:    0,
				Applied:           false,
				Message:           fmt.Sprintf("ยอดสินค้าต้องมีอย่างน้อย %.2f บาท", minOrderAmount),
			}, nil
		}

	default:
		return nil, errors.New("unsupported coupon type")
	}

	// Dynamic Usage Count Logic - คำนวนจำนวนครั้งการใช้งาน
	var totalUsage int
	var maxUsage int

	if coupon.Coupon.IsGlobalUsageMode() {
		// กรณีที่ 1: customers = [] (ไม่กำหนดลูกค้า)
		// ใช้ MaxUsageCount = คูปองใช้ได้รวมทั้งหมด X ครั้ง
		// ใครก็ใช้ได้ จนกว่าจะครบจำนวนรวม

		// นับจำนวนครั้งที่ใช้ไปแล้วทั้งหมด
		searchReq := models.CouponUsageHistoryRequest{
			CouponCode: coupon.Coupon.CouponCode,
			Page:       1,
			PageSize:   10000,
		}
		_, total, err := svc.usageHistoryRepo.SearchUsageHistory(ctx, shopID, searchReq)
		if err != nil {
			return nil, err
		}
		totalUsage = total
		maxUsage = coupon.Coupon.GetMaxUsageLimit()

	} else {
		// กรณีที่ 2: customers = ["A", "B", "C"] (กำหนดลูกค้า)
		// ใช้ MaxUsageCountPerCustomer = แต่ละลูกค้าใช้ได้ Y ครั้ง
		// เฉพาะลูกค้าที่กำหนดเท่านั้น และแต่ละคนใช้ได้ไม่เกิน Y ครั้ง

		// นับจำนวนครั้งที่ลูกค้าคนนี้ใช้ไปแล้ว
		searchReq := models.CouponUsageHistoryRequest{
			CouponCode: coupon.Coupon.CouponCode,
			CustomerID: customerID,
			Page:       1,
			PageSize:   10000,
		}
		_, customerTotal, err := svc.usageHistoryRepo.SearchUsageHistory(ctx, shopID, searchReq)
		if err != nil {
			return nil, err
		}
		totalUsage = customerTotal
		maxUsage = coupon.Coupon.GetMaxUsageLimit()
	}

	remainingUsage := maxUsage - totalUsage
	if remainingUsage < 0 {
		remainingUsage = 0
	}

	return &models.CalculateCouponItemResponse{
		CouponCode:        couponReq.CouponCode,
		CouponType:        coupon.Coupon.CouponType,
		CouponTypeName:    coupon.Coupon.CouponType.String(),
		DiscountAmount:    discountAmount,
		CashVoucherAmount: cashVoucherAmount,
		UsedAmount:        useAmount,
		UsageCount:        totalUsage,
		RemainingUsage:    remainingUsage,
		Applied:           true,
		Message:           "คำนวนสำเร็จ",
	}, nil
}

// คำนวนคูปองแต่ละตัวแบบขั้นสูง (รวม Product Condition และ Branch Exclusion)
func (svc CouponHttpService) calculateSingleCouponAdvanced(ctx context.Context, shopID string, orderAmount float64, items []models.CouponCheckItem, branchCode string, couponReq models.CalculateCouponItemRequest, customerID string) (*models.CalculateCouponItemResponse, error) {
	// ค้นหาคูปอง
	coupon, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "couponcode", couponReq.CouponCode)
	if err != nil {
		return nil, err
	}

	if coupon.CouponCode == "" {
		return nil, errors.New("coupon not found")
	}

	// ตรวจสอบสถานะคูปอง
	canUse, message := coupon.Coupon.CanUse()
	if !canUse {
		return &models.CalculateCouponItemResponse{
			CouponCode:        couponReq.CouponCode,
			CouponType:        coupon.Coupon.CouponType,
			CouponTypeName:    coupon.Coupon.CouponType.String(),
			DiscountAmount:    0,
			CashVoucherAmount: 0,
			UsedAmount:        0,
			UsageCount:        0,
			RemainingUsage:    0,
			Applied:           false,
			BranchAllowed:     true, // default ถือว่าสาขาอนุญาต เมื่อปัญหาไม่ได้มาจากสาขา
			EligibleAmount:    0,
			EligibleItemCount: 0,
			MinimumAmount:     0,
			Message:           message,
		}, nil
	}

	// ตรวจสอบสิทธิ์ลูกค้า
	if !coupon.Coupon.IsCustomerEligible(customerID) {
		return &models.CalculateCouponItemResponse{
			CouponCode:        couponReq.CouponCode,
			CouponType:        coupon.Coupon.CouponType,
			CouponTypeName:    coupon.Coupon.CouponType.String(),
			DiscountAmount:    0,
			CashVoucherAmount: 0,
			UsedAmount:        0,
			UsageCount:        0,
			RemainingUsage:    0,
			Applied:           false,
			BranchAllowed:     true,
			EligibleAmount:    0,
			EligibleItemCount: 0,
			MinimumAmount:     0,
			Message:           "คุณไม่มีสิทธิ์ใช้คูปองนี้",
		}, nil
	}

	// 1. ตรวจสอบสาขา (IgnoreBranchCode)
	branchAllowed := true
	if len(coupon.Coupon.IgnoreBranchCode) > 0 {
		for _, ignoredBranch := range coupon.Coupon.IgnoreBranchCode {
			if ignoredBranch == branchCode {
				branchAllowed = false
				break
			}
		}
	}

	if !branchAllowed {
		return &models.CalculateCouponItemResponse{
			CouponCode:        couponReq.CouponCode,
			CouponType:        coupon.Coupon.CouponType,
			CouponTypeName:    coupon.Coupon.CouponType.String(),
			DiscountAmount:    0,
			CashVoucherAmount: 0,
			UsedAmount:        0,
			UsageCount:        0,
			RemainingUsage:    0,
			Applied:           false,
			BranchAllowed:     false,
			EligibleAmount:    0,
			EligibleItemCount: 0,
			MinimumAmount:     0,
			Message:           "คูปองไม่สามารถใช้ได้ที่สาขานี้",
		}, nil
	}

	// 2. ตรวจสอบเงื่อนไขสินค้าและคำนวนส่วนลด
	var eligibleAmount float64 = 0
	var eligibleItemCount int = 0
	var itemResults []models.CouponItemCheckResult

	// ตรวจสอบว่าคูปองมีเงื่อนไขสินค้าหรือไม่
	hasProductCondition := coupon.Coupon.HasProductCondition()

	if hasProductCondition {
		// มีเงื่อนไขสินค้า - ต้องตรวจสอบแต่ละรายการ
		for _, item := range items {
			// ตรวจสอบสินค้าแต่ละรายการ
			isEligible, category, err := svc.checkProductCondition(ctx, shopID, item.Barcode, coupon.Coupon.ProductCondition)
			if err != nil {
				// ถ้าเกิดข้อผิดพลาดในการตรวจสอบ ถือว่าไม่เข้าเงื่อนไข
				isEligible = false
				category = "error: " + err.Error()
			}

			discountForThisItem := float64(0)
			if isEligible {
				eligibleAmount += item.SumAmount
				eligibleItemCount++
				// คำนวนส่วนลดสำหรับสินค้านี้ (จะคำนวนจริงทีหลัง)
				discountForThisItem = item.SumAmount // placeholder
			}

			itemResults = append(itemResults, models.CouponItemCheckResult{
				Barcode:         item.Barcode,
				Qty:             item.Qty,
				Price:           item.Price,
				SumAmount:       item.SumAmount,
				IsEligible:      isEligible,
				DiscountAmount:  discountForThisItem,
				MatchedCategory: category,
				Message:         "",
			})
		}

		// ตรวจสอบยอดขั้นต่ำ (MinimumAmount)
		if coupon.Coupon.ProductCondition.MinimumAmount > 0 && eligibleAmount < coupon.Coupon.ProductCondition.MinimumAmount {
			return &models.CalculateCouponItemResponse{
				CouponCode:        couponReq.CouponCode,
				CouponType:        coupon.Coupon.CouponType,
				CouponTypeName:    coupon.Coupon.CouponType.String(),
				DiscountAmount:    0,
				CashVoucherAmount: 0,
				UsedAmount:        0,
				UsageCount:        0,
				RemainingUsage:    0,
				Applied:           false,
				BranchAllowed:     true,
				EligibleAmount:    eligibleAmount,
				EligibleItemCount: eligibleItemCount,
				MinimumAmount:     coupon.Coupon.ProductCondition.MinimumAmount,
				ItemResults:       itemResults,
				Message:           fmt.Sprintf("ยอดรวมสินค้าที่เข้าเงื่อนไข %.2f บาท ต้องมีอย่างน้อย %.2f บาท", eligibleAmount, coupon.Coupon.ProductCondition.MinimumAmount),
			}, nil
		}

		// ถ้าไม่มีสินค้าเข้าเงื่อนไข
		if eligibleItemCount == 0 {
			return &models.CalculateCouponItemResponse{
				CouponCode:        couponReq.CouponCode,
				CouponType:        coupon.Coupon.CouponType,
				CouponTypeName:    coupon.Coupon.CouponType.String(),
				DiscountAmount:    0,
				CashVoucherAmount: 0,
				UsedAmount:        0,
				UsageCount:        0,
				RemainingUsage:    0,
				Applied:           false,
				BranchAllowed:     true,
				EligibleAmount:    0,
				EligibleItemCount: 0,
				MinimumAmount:     0,
				ItemResults:       itemResults,
				Message:           "ไม่มีสินค้าในตะกร้าที่เข้าเงื่อนไขคูปอง",
			}, nil
		}
	} else {
		// ไม่มีเงื่อนไขสินค้า - ใช้ได้กับสินค้าทั้งหมด
		for _, item := range items {
			eligibleAmount += item.SumAmount
			eligibleItemCount++

			itemResults = append(itemResults, models.CouponItemCheckResult{
				Barcode:         item.Barcode,
				Qty:             item.Qty,
				Price:           item.Price,
				SumAmount:       item.SumAmount,
				IsEligible:      true,
				DiscountAmount:  item.SumAmount, // placeholder
				MatchedCategory: "ทั้งหมด",
				Message:         "",
			})
		}
	}

	// 3. คำนวนส่วนลดตามประเภทคูปอง
	var discountAmount float64 = 0
	var cashVoucherAmount float64 = 0
	useAmount := couponReq.UseAmount

	// ใช้ยอดที่เข้าเงื่อนไขในการคำนวน (แทน orderAmount)
	calculationBase := eligibleAmount

	switch coupon.Coupon.CouponType {
	case models.CouponTypeValueDiscount:
		// ลดตามมูลค่า - ใช้ useAmount ที่ส่งมา หรือ CouponValue ถ้าไม่ระบุ
		if useAmount <= 0 {
			useAmount = coupon.Coupon.CouponValue
		}
		discountAmount = useAmount
		if discountAmount > calculationBase {
			discountAmount = calculationBase
		}

	case models.CouponTypePercentDiscount:
		// ลดตามเปอร์เซ็นต์
		percentageToUse := couponReq.UseAmount
		if percentageToUse <= 0 || percentageToUse > coupon.Coupon.CouponValue {
			percentageToUse = coupon.Coupon.CouponValue
		}
		discountAmount = (calculationBase * percentageToUse) / 100

	case models.CouponTypeCashVoucher:
		// คูปองแทนเงินสด - ใช้ useAmount ที่ส่งมา หรือ CouponValue ถ้าไม่ระบุ
		if useAmount <= 0 {
			useAmount = coupon.Coupon.CouponValue
		}
		cashVoucherAmount = useAmount

		// ตรวจสอบเงื่อนไขขั้นต่ำจากคูปอง (ถ้ามี ProductCondition)
		// หากไม่มี MinimumAmount ให้ใช้ค่าเริ่มต้น 0 (ไม่จำกัด)
		minOrderAmount := float64(0)
		if coupon.Coupon.ProductCondition != nil && coupon.Coupon.ProductCondition.MinimumAmount > 0 {
			minOrderAmount = coupon.Coupon.ProductCondition.MinimumAmount
		}

		// ตรวจสอบเงื่อนไขขั้นต่ำ (ถ้ามีกำหนด)
		if minOrderAmount > 0 && calculationBase < minOrderAmount {
			return &models.CalculateCouponItemResponse{
				CouponCode:        couponReq.CouponCode,
				CouponType:        coupon.Coupon.CouponType,
				CouponTypeName:    coupon.Coupon.CouponType.String(),
				DiscountAmount:    0,
				CashVoucherAmount: 0,
				UsedAmount:        0,
				UsageCount:        0,
				RemainingUsage:    0,
				Applied:           false,
				BranchAllowed:     true,
				EligibleAmount:    eligibleAmount,
				EligibleItemCount: eligibleItemCount,
				MinimumAmount:     minOrderAmount,
				ItemResults:       itemResults,
				Message:           fmt.Sprintf("ยอดสินค้าที่เข้าเงื่อนไข %.2f บาท ต้องมีอย่างน้อย %.2f บาท", calculationBase, minOrderAmount),
			}, nil
		}

	default:
		return nil, errors.New("unsupported coupon type")
	}

	// 4. อัปเดตส่วนลดสำหรับแต่ละสินค้า
	svc.updateItemDiscountResults(itemResults, discountAmount, eligibleAmount)

	// 5. Dynamic Usage Count Logic - คำนวนจำนวนครั้งการใช้งาน
	var totalUsage int
	var maxUsage int

	if coupon.Coupon.IsGlobalUsageMode() {
		// Global mode - นับจำนวนครั้งที่ใช้ไปแล้วทั้งหมด
		searchReq := models.CouponUsageHistoryRequest{
			CouponCode: coupon.Coupon.CouponCode,
			Page:       1,
			PageSize:   10000,
		}
		_, total, err := svc.usageHistoryRepo.SearchUsageHistory(ctx, shopID, searchReq)
		if err != nil {
			return nil, err
		}
		totalUsage = total
		maxUsage = coupon.Coupon.GetMaxUsageLimit()
	} else {
		// Per-customer mode - นับจำนวนครั้งที่ลูกค้าคนนี้ใช้ไปแล้ว
		searchReq := models.CouponUsageHistoryRequest{
			CouponCode: coupon.Coupon.CouponCode,
			CustomerID: customerID,
			Page:       1,
			PageSize:   10000,
		}
		_, customerTotal, err := svc.usageHistoryRepo.SearchUsageHistory(ctx, shopID, searchReq)
		if err != nil {
			return nil, err
		}
		totalUsage = customerTotal
		maxUsage = coupon.Coupon.GetMaxUsageLimit()
	}

	remainingUsage := maxUsage - totalUsage
	if remainingUsage < 0 {
		remainingUsage = 0
	}

	// ดึง MinimumAmount จากคูปอง
	minimumAmount := float64(0)
	if coupon.Coupon.ProductCondition != nil {
		minimumAmount = coupon.Coupon.ProductCondition.MinimumAmount
	}

	return &models.CalculateCouponItemResponse{
		CouponCode:        couponReq.CouponCode,
		CouponType:        coupon.Coupon.CouponType,
		CouponTypeName:    coupon.Coupon.CouponType.String(),
		DiscountAmount:    discountAmount,
		CashVoucherAmount: cashVoucherAmount,
		UsedAmount:        useAmount,
		UsageCount:        totalUsage,
		RemainingUsage:    remainingUsage,
		Applied:           true,
		BranchAllowed:     true,
		EligibleAmount:    eligibleAmount,
		EligibleItemCount: eligibleItemCount,
		MinimumAmount:     minimumAmount,
		ItemResults:       itemResults,
		Message:           "คำนวนสำเร็จ",
	}, nil
}

// อัปเดตส่วนลดสำหรับแต่ละสินค้าตามสัดส่วน
func (svc CouponHttpService) updateItemDiscountResults(itemResults []models.CouponItemCheckResult, totalDiscount, eligibleAmount float64) {
	if eligibleAmount <= 0 {
		return
	}

	for i := range itemResults {
		if itemResults[i].IsEligible {
			if totalDiscount > 0 {
				// คำนวนส่วนลดตามสัดส่วน (สำหรับคูปองลดราคา)
				proportion := itemResults[i].SumAmount / eligibleAmount
				itemResults[i].DiscountAmount = totalDiscount * proportion
			} else {
				// สำหรับคูปองแทนเงินสด - ไม่มีการลดราคาสินค้า
				itemResults[i].DiscountAmount = 0
			}
		} else {
			itemResults[i].DiscountAmount = 0
		}
	}
}

// อัปเดตยอดสินค้าหลังจากใช้คูปอง (สำหรับ cascading calculation)
func (svc CouponHttpService) updateItemAmountsAfterCouponApplied(items []models.CouponCheckItem, result *models.CalculateCouponItemResponse) {
	if !result.Applied || len(result.ItemResults) == 0 {
		return
	}

	if result.CouponType == models.CouponTypeCashVoucher {
		// สำหรับคูปองแทนเงินสด - ต้อง "จอง" ยอดสินค้าตาม MinimumAmount
		// เพื่อไม่ให้คูปองตัวถัดไปใช้ยอดเดียวกันซ้ำ
		if result.EligibleAmount > 0 && result.MinimumAmount > 0 {
			// ยอดที่ต้องจอง = MinimumAmount (ไม่ใช่ EligibleAmount ทั้งหมด)
			// แต่ไม่เกินยอดที่เข้าเงื่อนไขจริง
			reservedAmount := result.MinimumAmount
			if reservedAmount > result.EligibleAmount {
				reservedAmount = result.EligibleAmount
			}

			// ลดยอดสินค้าที่เข้าเงื่อนไขตามสัดส่วน
			for i, item := range items {
				for _, itemResult := range result.ItemResults {
					if item.Barcode == itemResult.Barcode && itemResult.IsEligible {
						// คำนวนสัดส่วนการลด
						proportion := itemResult.SumAmount / result.EligibleAmount
						reductionAmount := reservedAmount * proportion

						newAmount := item.SumAmount - reductionAmount
						if newAmount < 0 {
							newAmount = 0
						}
						items[i].SumAmount = newAmount
						break
					}
				}
			}
		}
		return
	} // สำหรับคูปองลดราคา - ลดยอดสินค้าตามส่วนลดที่ได้รับ
	for i, item := range items {
		for _, itemResult := range result.ItemResults {
			if item.Barcode == itemResult.Barcode && itemResult.IsEligible && itemResult.DiscountAmount > 0 {
				// ลดยอดสินค้าตามส่วนลดที่ได้รับ
				newAmount := item.SumAmount - itemResult.DiscountAmount
				if newAmount < 0 {
					newAmount = 0
				}
				items[i].SumAmount = newAmount
				break
			}
		}
	}
}

// CheckReservationByTransactionID เช็คสถานะการจองด้วย Transaction ID
func (svc CouponHttpService) CheckReservationByTransactionID(transactionID, shopID string) (*models.ReservationStatusResponse, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// ค้นหาการจองด้วย transaction ID
	reservation, err := svc.reservationRepo.FindReservationByTransactionID(ctx, shopID, transactionID)
	if err != nil {
		return &models.ReservationStatusResponse{
			TransactionID:     transactionID,
			Reservations:      []models.ReservationInfo{},
			ReservationStatus: "not_found",
			Count:             0,
		}, nil
	}

	// ตรวจสอบว่าการจองหมดอายุหรือไม่
	isExpired := reservation.CouponReservation.IsExpired()

	status := reservation.CouponReservation.Status
	statusName := "unknown"
	switch status {
	case models.ReservationStatusActive:
		statusName = "active"
	case models.ReservationStatusUsed:
		statusName = "used"
	case models.ReservationStatusExpired:
		statusName = "expired"
	case models.ReservationStatusCanceled:
		statusName = "canceled"
	}

	if isExpired && status == models.ReservationStatusActive {
		statusName = "expired"
	}

	// ค้นหาข้อมูลคูปองเพื่อแสดงรายละเอียด
	coupon, err := svc.repo.FindByGuid(ctx, shopID, reservation.CouponID)
	couponCode := ""
	if err == nil {
		couponCode = coupon.Coupon.CouponCode
	}

	reservationInfo := models.ReservationInfo{
		ID:         reservation.ID.Hex(),
		CouponID:   reservation.CouponID,
		CouponCode: couponCode,
		CustomerID: reservation.CouponReservation.CustomerID,
		Status:     status,
		StatusName: statusName,
		ReservedAt: reservation.CouponReservation.ReservedAt,
		ExpiresAt:  reservation.CouponReservation.ExpiresAt,
		UsedAt:     reservation.CouponReservation.UsedAt,
		CanceledAt: reservation.CouponReservation.CanceledAt,
		IsExpired:  isExpired,
	}

	return &models.ReservationStatusResponse{
		TransactionID:     transactionID,
		Reservations:      []models.ReservationInfo{reservationInfo},
		ReservationStatus: statusName,
		Count:             1,
	}, nil
}

// LookupReservationDetails ค้นหารายละเอียดการจองแบบละเอียด
func (svc CouponHttpService) LookupReservationDetails(shopID string, req models.ReservationLookupRequest) (*models.ReservationLookupResponse, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// ค้นหาการจองด้วย transaction ID
	reservation, err := svc.reservationRepo.FindReservationByTransactionID(ctx, shopID, req.TransactionID)
	if err != nil {
		return &models.ReservationLookupResponse{
			TransactionID: req.TransactionID,
			Found:         false,
			Reservations:  []models.ReservationInfo{},
			Summary: struct {
				TotalReservations int `json:"total_reservations"`
				ActiveCount       int `json:"active_count"`
				UsedCount         int `json:"used_count"`
				ExpiredCount      int `json:"expired_count"`
				CancelledCount    int `json:"cancelled_count"`
			}{
				TotalReservations: 0,
				ActiveCount:       0,
				UsedCount:         0,
				ExpiredCount:      0,
				CancelledCount:    0,
			},
		}, nil
	}

	// ตรวจสอบว่าจะแสดงการจองนี้หรือไม่ตามเงื่อนไข
	status := reservation.CouponReservation.Status
	isExpired := reservation.CouponReservation.IsExpired()

	shouldInclude := true
	if status == models.ReservationStatusExpired && !req.IncludeExpired {
		shouldInclude = false
	}
	if status == models.ReservationStatusCanceled && !req.IncludeCancelled {
		shouldInclude = false
	}
	if status == models.ReservationStatusUsed && !req.IncludeUsed {
		shouldInclude = false
	}
	if isExpired && status == models.ReservationStatusActive && !req.IncludeExpired {
		shouldInclude = false
	}

	var reservations []models.ReservationInfo
	var summary struct {
		TotalReservations int `json:"total_reservations"`
		ActiveCount       int `json:"active_count"`
		UsedCount         int `json:"used_count"`
		ExpiredCount      int `json:"expired_count"`
		CancelledCount    int `json:"cancelled_count"`
	}

	if shouldInclude {
		// ค้นหาข้อมูลคูปอง
		coupon, err := svc.repo.FindByGuid(ctx, shopID, reservation.CouponID)
		couponCode := ""
		if err == nil {
			couponCode = coupon.Coupon.CouponCode
		}

		statusName := "unknown"
		switch status {
		case models.ReservationStatusActive:
			statusName = "active"
		case models.ReservationStatusUsed:
			statusName = "used"
		case models.ReservationStatusExpired:
			statusName = "expired"
		case models.ReservationStatusCanceled:
			statusName = "canceled"
		}

		if isExpired && status == models.ReservationStatusActive {
			statusName = "expired"
		}

		reservationInfo := models.ReservationInfo{
			ID:         reservation.ID.Hex(),
			CouponID:   reservation.CouponID,
			CouponCode: couponCode,
			CustomerID: reservation.CouponReservation.CustomerID,
			Status:     status,
			StatusName: statusName,
			ReservedAt: reservation.CouponReservation.ReservedAt,
			ExpiresAt:  reservation.CouponReservation.ExpiresAt,
			UsedAt:     reservation.CouponReservation.UsedAt,
			CanceledAt: reservation.CouponReservation.CanceledAt,
			IsExpired:  isExpired,
		}
		reservations = append(reservations, reservationInfo)

		// คำนวน summary
		summary.TotalReservations = 1
		if status == models.ReservationStatusActive && !isExpired {
			summary.ActiveCount = 1
		} else if status == models.ReservationStatusUsed {
			summary.UsedCount = 1
		} else if status == models.ReservationStatusCanceled {
			summary.CancelledCount = 1
		} else if isExpired {
			summary.ExpiredCount = 1
		}
	}

	return &models.ReservationLookupResponse{
		TransactionID: req.TransactionID,
		Found:         len(reservations) > 0,
		Reservations:  reservations,
		Summary:       summary,
	}, nil
}

// Usage History Service Methods

// CreateUsageHistory สร้างประวัติการใช้คูปอง
func (svc CouponHttpService) CreateUsageHistory(shopID, authUsername string, req models.CreateUsageHistoryRequest) (string, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// สร้าง usage history doc
	historyDoc := models.CouponUsageHistoryDoc{
		CouponUsageHistoryData: models.CouponUsageHistoryData{
			CouponUsageHistory: models.CouponUsageHistory{
				CouponID:          req.CouponID,
				CouponCode:        req.CouponCode,
				CustomerID:        req.CustomerID,
				CustomerCode:      req.CustomerCode,
				CustomerName:      req.CustomerName,
				SaleInvoiceID:     req.SaleInvoiceID,
				SaleInvoiceNumber: req.SaleInvoiceNumber,
				TransactionID:     req.TransactionID,
				ReservationID:     req.ReservationID,
				UsedAmount:        req.UsedAmount,
				DiscountAmount:    req.DiscountAmount,
				CashVoucherAmount: req.CashVoucherAmount,
				OrderAmount:       req.OrderAmount,
				CouponType:        req.CouponType,
				CouponValue:       req.CouponValue,
				UsedAt:            time.Now().UTC(),
				UsedBy:            authUsername,
				Remark:            req.Remark,
			},
		},
	}

	// สร้างประวัติการใช้
	historyID, err := svc.usageHistoryRepo.CreateUsageHistory(ctx, shopID, historyDoc)
	if err != nil {
		return "", err
	}

	return historyID, nil
}

// GetUsageHistory ดึงประวัติการใช้คูปองตามเงื่อนไข
func (svc CouponHttpService) GetUsageHistory(shopID string, req models.CouponUsageHistoryRequest) (*models.CouponUsageHistoryResponse, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	response, err := svc.usageHistoryRepo.GetUsageHistorySummary(ctx, shopID, req)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetUsageHistoryByCoupon ดึงประวัติการใช้คูปองตามรหัสคูปอง
func (svc CouponHttpService) GetUsageHistoryByCoupon(shopID, couponID string, page, pageSize int) (*models.CouponUsageHistoryResponse, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	req := models.CouponUsageHistoryRequest{
		CouponID: couponID,
		Page:     page,
		PageSize: pageSize,
	}

	response, err := svc.usageHistoryRepo.GetUsageHistorySummary(ctx, shopID, req)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetUsageHistoryByCustomer ดึงประวัติการใช้คูปองตามรหัสลูกค้า
func (svc CouponHttpService) GetUsageHistoryByCustomer(shopID, customerID string, page, pageSize int) (*models.CouponUsageHistoryResponse, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	req := models.CouponUsageHistoryRequest{
		CustomerID: customerID,
		Page:       page,
		PageSize:   pageSize,
	}

	response, err := svc.usageHistoryRepo.GetUsageHistorySummary(ctx, shopID, req)
	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GetUsageHistoryBySaleInvoice ดึงประวัติการใช้คูปองตามรหัสใบกำกับสินค้า
func (svc CouponHttpService) GetUsageHistoryBySaleInvoice(shopID, saleInvoiceID string) ([]models.CouponUsageHistoryItem, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	histories, err := svc.usageHistoryRepo.FindUsageHistoryBySaleInvoice(ctx, shopID, saleInvoiceID)
	if err != nil {
		return nil, err
	}

	// Convert to response items
	var items []models.CouponUsageHistoryItem
	for _, history := range histories {
		item := models.CouponUsageHistoryItem{
			ID:                history.ID.Hex(),
			CouponID:          history.CouponUsageHistory.CouponID,
			CouponCode:        history.CouponUsageHistory.CouponCode,
			CouponType:        history.CouponUsageHistory.CouponType,
			CouponTypeName:    history.CouponUsageHistory.CouponType.String(),
			CouponValue:       history.CouponUsageHistory.CouponValue,
			CustomerID:        history.CouponUsageHistory.CustomerID,
			CustomerCode:      history.CouponUsageHistory.CustomerCode,
			CustomerName:      history.CouponUsageHistory.CustomerName,
			SaleInvoiceID:     history.CouponUsageHistory.SaleInvoiceID,
			SaleInvoiceNumber: history.CouponUsageHistory.SaleInvoiceNumber,
			TransactionID:     history.CouponUsageHistory.TransactionID,
			ReservationID:     history.CouponUsageHistory.ReservationID,
			UsedAmount:        history.CouponUsageHistory.UsedAmount,
			DiscountAmount:    history.CouponUsageHistory.DiscountAmount,
			CashVoucherAmount: history.CouponUsageHistory.CashVoucherAmount,
			OrderAmount:       history.CouponUsageHistory.OrderAmount,
			UsedAt:            history.CouponUsageHistory.UsedAt,
			UsedBy:            history.CouponUsageHistory.UsedBy,
			Remark:            history.CouponUsageHistory.Remark,
		}
		items = append(items, item)
	}

	return items, nil
}

// GetUsageHistoryByTransactionID ดึงประวัติการใช้คูปองตาม Transaction ID
func (svc CouponHttpService) GetUsageHistoryByTransactionID(shopID, transactionID string) ([]models.CouponUsageHistoryItem, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	histories, err := svc.usageHistoryRepo.FindUsageHistoryByTransactionID(ctx, shopID, transactionID)
	if err != nil {
		return nil, err
	}

	// Convert to response items
	var items []models.CouponUsageHistoryItem
	for _, history := range histories {
		item := models.CouponUsageHistoryItem{
			ID:                history.ID.Hex(),
			CouponID:          history.CouponUsageHistory.CouponID,
			CouponCode:        history.CouponUsageHistory.CouponCode,
			CouponType:        history.CouponUsageHistory.CouponType,
			CouponTypeName:    history.CouponUsageHistory.CouponType.String(),
			CouponValue:       history.CouponUsageHistory.CouponValue,
			CustomerID:        history.CouponUsageHistory.CustomerID,
			CustomerCode:      history.CouponUsageHistory.CustomerCode,
			CustomerName:      history.CouponUsageHistory.CustomerName,
			SaleInvoiceID:     history.CouponUsageHistory.SaleInvoiceID,
			SaleInvoiceNumber: history.CouponUsageHistory.SaleInvoiceNumber,
			TransactionID:     history.CouponUsageHistory.TransactionID,
			ReservationID:     history.CouponUsageHistory.ReservationID,
			UsedAmount:        history.CouponUsageHistory.UsedAmount,
			DiscountAmount:    history.CouponUsageHistory.DiscountAmount,
			CashVoucherAmount: history.CouponUsageHistory.CashVoucherAmount,
			OrderAmount:       history.CouponUsageHistory.OrderAmount,
			UsedAt:            history.CouponUsageHistory.UsedAt,
			UsedBy:            history.CouponUsageHistory.UsedBy,
			Remark:            history.CouponUsageHistory.Remark,
		}
		items = append(items, item)
	}

	return items, nil
}
