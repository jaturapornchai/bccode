package services

import (
	"context"
	"errors"
	"fmt"
	"math"
	currencyrepo "smlcloudplatform/internal/currency/repositories"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/transaction/purchaseorder/models"
	"smlcloudplatform/internal/transaction/purchaseorder/repositories"
	"smlcloudplatform/internal/transaction/purchaseorder/validators"
	trancache "smlcloudplatform/internal/transaction/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strconv"
	"strings"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IPurchaseOrderHttpService interface {
	CreatePurchaseOrder(shopID string, authUsername string, doc models.PurchaseOrder) (string, string, *validators.ValidationResult, error)
	UpdatePurchaseOrder(shopID string, guid string, authUsername string, doc models.PurchaseOrder) (*validators.ValidationResult, error)
	DeletePurchaseOrder(shopID string, guid string, authUsername string) error
	DeletePurchaseOrderByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoPurchaseOrder(shopID string, guid string) (models.PurchaseOrderInfo, error)
	InfoPurchaseOrderByCode(shopID string, code string) (models.PurchaseOrderInfo, error)
	SearchPurchaseOrder(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PurchaseOrderInfo, mongopagination.PaginationData, error)
	SearchPurchaseOrderStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PurchaseOrderInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.PurchaseOrder) (common.BulkImport, error)

	GetModuleName() string
}

const (
	MODULE_NAME = "PO"
)

type PurchaseOrderHttpService struct {
	repoMq           repositories.IPurchaseOrderMessageQueueRepository
	repo             repositories.IPurchaseOrderRepository
	repoCache        trancache.ICacheRepository
	currencyRepo     currencyrepo.ICurrencyRepository
	cacheExpireDocNo time.Duration
	syncCacheRepo    mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.PurchaseOrderActivity, models.PurchaseOrderDeleteActivity]
	contextTimeout time.Duration
}

func NewPurchaseOrderHttpService(
	repo repositories.IPurchaseOrderRepository,
	repoCache trancache.ICacheRepository,
	repoMq repositories.IPurchaseOrderMessageQueueRepository,
	currencyRepo currencyrepo.ICurrencyRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
) *PurchaseOrderHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &PurchaseOrderHttpService{
		repo:             repo,
		repoMq:           repoMq,
		repoCache:        repoCache,
		currencyRepo:     currencyRepo,
		syncCacheRepo:    syncCacheRepo,
		cacheExpireDocNo: time.Hour * 24,
		contextTimeout:   contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.PurchaseOrderActivity, models.PurchaseOrderDeleteActivity](repo)

	return insSvc
}

func (svc PurchaseOrderHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

// getDefaultCurrencySymbol - คืนค่า default currency symbol สำหรับ common currencies
func getDefaultCurrencySymbol(currencyCode string) string {
	symbols := map[string]string{
		"THB": "฿",
		"USD": "$",
		"EUR": "€",
		"GBP": "£",
		"JPY": "¥",
		"CNY": "¥",
		"KRW": "₩",
		"SGD": "S$",
		"MYR": "RM",
		"IDR": "Rp",
		"VND": "₫",
		"PHP": "₱",
		"AUD": "A$",
		"NZD": "NZ$",
		"HKD": "HK$",
		"TWD": "NT$",
		"INR": "₹",
		"CHF": "Fr",
		"CAD": "C$",
		"SEK": "kr",
		"NOK": "kr",
		"DKK": "kr",
		"RUB": "₽",
		"BRL": "R$",
		"ZAR": "R",
		"AED": "د.إ",
		"SAR": "﷼",
		"LAK": "₭",
		"KHR": "៛",
		"MMK": "K",
	}

	code := strings.ToUpper(currencyCode)
	if symbol, ok := symbols[code]; ok {
		return symbol
	}
	// ถ้าไม่มีใน map ใช้ currency code เป็น symbol
	return code
}

// populateCurrencySymbol - ดึง currency symbol จาก currency code (มี fallback กรณี database ไม่มีข้อมูล)
func (svc PurchaseOrderHttpService) populateCurrencySymbol(ctx context.Context, shopID string, doc *models.PurchaseOrder) {
	fmt.Printf("[CURRENCY-SYMBOL] Start - Currency='%s'\n", doc.Currency)

	// ถ้าไม่มี currency ให้ default เป็น THB
	if doc.Currency == "" {
		doc.Currency = "THB"
		doc.CurrencySymbol = "฿"
		fmt.Println("[CURRENCY-SYMBOL] Empty currency, defaulting to THB (฿)")
		return
	}

	currencyCode := strings.ToUpper(doc.Currency)

	// พยายามดึงข้อมูล currency จาก database ก่อน
	fmt.Printf("[CURRENCY-SYMBOL] Looking up currency code: %s for shop: %s\n", currencyCode, shopID)
	currencyDoc, err := svc.currencyRepo.FindByCode(ctx, shopID, currencyCode)

	if err == nil && len(currencyDoc.GuidFixed) > 0 && currencyDoc.Symbol != "" {
		// พบข้อมูลใน database และมี symbol
		doc.CurrencySymbol = currencyDoc.Symbol
		fmt.Printf("[CURRENCY-SYMBOL] Success (from DB) - Set symbol='%s' for currency '%s'\n", doc.CurrencySymbol, doc.Currency)
		return
	}

	// ถ้า database ไม่มีข้อมูล หรือ query ไม่สำเร็จ ใช้ fallback default symbols
	if err != nil {
		fmt.Printf("[CURRENCY-SYMBOL] Database lookup failed: %v, using fallback\n", err)
	} else {
		fmt.Printf("[CURRENCY-SYMBOL] Currency %s not found in database, using fallback\n", currencyCode)
	}

	doc.CurrencySymbol = getDefaultCurrencySymbol(currencyCode)
	fmt.Printf("[CURRENCY-SYMBOL] Fallback - Set symbol='%s' for currency '%s'\n", doc.CurrencySymbol, doc.Currency)
}

// sanitizeExchangeRate - ตรวจสอบและแก้ไข exchange rate ที่ไม่ถูกต้อง (<=0, NaN, Inf) → default 1.0
func sanitizeExchangeRate(doc *models.PurchaseOrder) {
	rate := doc.ExchangeRate
	if rate <= 0 || math.IsNaN(rate) || math.IsInf(rate, 0) {
		fmt.Printf("[PO] Exchange rate ไม่ถูกต้อง: %v → ใช้ 1.0 แทน\n", rate)
		doc.ExchangeRate = 1.0
	}
}

// getDocNoPrefix สร้าง prefix สำหรับ docno จาก docDateLocal ที่ frontend ส่งมา
// docDateLocalStr ต้องมีค่าเสมอ (เช่น "20260107" หรือ "2026-01-07")
func (svc PurchaseOrderHttpService) getDocNoPrefix(docDateLocalStr string) string {
	// รองรับทั้ง format "20260107" และ "2026-01-07"
	docDateStr := strings.ReplaceAll(docDateLocalStr, "-", "")
	if len(docDateStr) >= 8 {
		docDateStr = docDateStr[:8] // ตัดเอาเฉพาะ YYYYMMDD
	}

	return fmt.Sprintf("%s%s", MODULE_NAME, docDateStr)
}

func (svc PurchaseOrderHttpService) generateNewDocNo(ctx context.Context, shopID, prefixDocNo string, docNumber int) (string, int, error) {
	prevoiusDocNumber, err := svc.repoCache.Get(shopID, prefixDocNo)

	if prevoiusDocNumber == 0 || err != nil {
		lastDoc, err := svc.repo.FindLastDocNo(ctx, shopID, prefixDocNo)

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

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", newDocNo)

	if err != nil {
		return "", 0, err
	}

	if len(findDoc.GuidFixed) > 0 {
		return "", 0, errors.New("DocNo is exists")
	}

	return newDocNo, newDocNumber, nil
}

func (svc PurchaseOrderHttpService) CreatePurchaseOrder(shopID string, authUsername string, doc models.PurchaseOrder) (string, string, *validators.ValidationResult, error) {

	// Layer 2: Business validation — ตรวจสอบข้อมูล PO ก่อนบันทึก
	validationResult := validators.ValidatePurchaseOrder(&doc)
	if !validationResult.IsValid() {
		return "", "", validationResult, nil
	}

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	prefixDocNo := svc.getDocNoPrefix(doc.DocDateLocal)

	newDocNo, newDocNumber, err := svc.generateNewDocNo(ctx, shopID, prefixDocNo, 1)

	if err != nil {
		return "", "", nil, err
	}

	newGuidFixed := utils.NewGUID()

	docData := models.PurchaseOrderDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.PurchaseOrder = doc

	docData.DocNo = newDocNo
	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	// ตั้งค่า creatorcode, creatorname และ createdat สำหรับส่งไป backend ผ่าน Kafka
	// CreatorCode ใช้ authUsername จาก JWT (เชื่อถือได้)
	// CreatorName ใช้ค่าจาก request (ชื่อจริงที่ Flutter ส่งมา) ถ้าว่างค่อย fallback เป็น authUsername
	docData.PurchaseOrder.CreatorCode = authUsername
	if docData.PurchaseOrder.CreatorName == "" {
		docData.PurchaseOrder.CreatorName = authUsername
	}
	docData.PurchaseOrder.CreatedAt = time.Now()

	// ตรวจสอบ exchange rate และดึง currency symbol
	sanitizeExchangeRate(&docData.PurchaseOrder)
	svc.populateCurrencySymbol(ctx, shopID, &docData.PurchaseOrder)

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", "", nil, err
	}

	go svc.repoCache.Save(shopID, prefixDocNo, newDocNumber, svc.cacheExpireDocNo)

	go func() {
		err := svc.repoMq.Create(docData)
		if err != nil {
			fmt.Printf("[KAFKA-ERROR] Failed to publish PO create message: DocNo=%s, Error=%v\n", docData.DocNo, err)
		} else {
			fmt.Printf("[KAFKA-OK] Published PO create message: DocNo=%s\n", docData.DocNo)
		}
		svc.repoCache.Save(shopID, prefixDocNo, newDocNumber, svc.cacheExpireDocNo)
		svc.saveMasterSync(shopID)
	}()

	return newGuidFixed, newDocNo, nil, nil
}

func (svc PurchaseOrderHttpService) UpdatePurchaseOrder(shopID string, guid string, authUsername string, doc models.PurchaseOrder) (*validators.ValidationResult, error) {

	// Layer 2: Business validation — ตรวจสอบข้อมูล PO ก่อนบันทึก
	validationResult := validators.ValidatePurchaseOrder(&doc)
	if !validationResult.IsValid() {
		return validationResult, nil
	}

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return nil, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return nil, errors.New("document not found")
	}

	// เก็บค่า creator เดิมก่อนที่จะถูก overwrite
	originalCreatorCode := findDoc.PurchaseOrder.CreatorCode
	originalCreatorName := findDoc.PurchaseOrder.CreatorName
	originalCreatedAt := findDoc.PurchaseOrder.CreatedAt

	docData := findDoc
	docData.PurchaseOrder = doc

	docData.DocNo = findDoc.DocNo
	docData.UpdatedBy = authUsername
	docData.UpdatedAt = time.Now()

	// คืนค่า creator เดิม (ไม่ให้ถูก overwrite จาก request)
	docData.PurchaseOrder.CreatorCode = originalCreatorCode
	docData.PurchaseOrder.CreatorName = originalCreatorName
	docData.PurchaseOrder.CreatedAt = originalCreatedAt

	// ตั้งค่า updatercode และ updatername สำหรับส่งไป backend ผ่าน Kafka
	docData.PurchaseOrder.UpdaterCode = authUsername
	docData.PurchaseOrder.UpdaterName = authUsername
	docData.PurchaseOrder.UpdatedAt = time.Now()

	// ตรวจสอบ exchange rate และดึง currency symbol
	sanitizeExchangeRate(&docData.PurchaseOrder)
	svc.populateCurrencySymbol(ctx, shopID, &docData.PurchaseOrder)

	err = svc.repo.Update(ctx, shopID, guid, docData)

	if err != nil {
		return nil, err
	}

	func() {
		err := svc.repoMq.Update(docData)
		if err != nil {
			fmt.Printf("[KAFKA-ERROR] Failed to publish PO update message: DocNo=%s, Error=%v\n", docData.DocNo, err)
		} else {
			fmt.Printf("[KAFKA-OK] Published PO update message: DocNo=%s\n", docData.DocNo)
		}
		svc.saveMasterSync(shopID)
	}()

	return nil, nil
}

func (svc PurchaseOrderHttpService) DeletePurchaseOrder(shopID string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, shopID, guid, authUsername)
	if err != nil {
		return err
	}

	func() {
		err := svc.repoMq.Delete(findDoc)
		if err != nil {
			fmt.Printf("[KAFKA-ERROR] Failed to publish PO delete message: DocNo=%s, Error=%v\n", findDoc.DocNo, err)
		} else {
			fmt.Printf("[KAFKA-OK] Published PO delete message: DocNo=%s\n", findDoc.DocNo)
		}
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc PurchaseOrderHttpService) DeletePurchaseOrderByGUIDs(shopID string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, shopID, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	func() {
		docs, _ := svc.repo.FindByGuids(ctx, shopID, GUIDs)
		svc.repoMq.DeleteInBatch(docs)
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc PurchaseOrderHttpService) InfoPurchaseOrder(shopID string, guid string) (models.PurchaseOrderInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.PurchaseOrderInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.PurchaseOrderInfo{}, errors.New("document not found")
	}

	return findDoc.PurchaseOrderInfo, nil
}

func (svc PurchaseOrderHttpService) InfoPurchaseOrderByCode(shopID string, code string) (models.PurchaseOrderInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", code)

	if err != nil {
		return models.PurchaseOrderInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.PurchaseOrderInfo{}, errors.New("document not found")
	}

	return findDoc.PurchaseOrderInfo, nil
}

func (svc PurchaseOrderHttpService) SearchPurchaseOrder(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PurchaseOrderInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.PurchaseOrderInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc PurchaseOrderHttpService) SearchPurchaseOrderStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PurchaseOrderInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.PurchaseOrderInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc PurchaseOrderHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.PurchaseOrder) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.PurchaseOrder](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.DocNo)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, shopID, "docno", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.DocNo)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.PurchaseOrder, models.PurchaseOrderDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.PurchaseOrder) models.PurchaseOrderDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.PurchaseOrderDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.PurchaseOrder = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.PurchaseOrder, models.PurchaseOrderDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.PurchaseOrderDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", guid)
		},
		func(doc models.PurchaseOrderDoc) bool {
			return doc.DocNo != ""
		},
		func(shopID string, authUsername string, data models.PurchaseOrder, doc models.PurchaseOrderDoc) error {

			doc.PurchaseOrder = data
			doc.UpdatedBy = authUsername
			doc.UpdatedAt = time.Now()

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

	svc.saveMasterSync(shopID)

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

func (svc PurchaseOrderHttpService) getDocIDKey(doc models.PurchaseOrder) string {
	return doc.DocNo
}

func (svc PurchaseOrderHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc PurchaseOrderHttpService) GetModuleName() string {
	return "purchaseorder"
}
