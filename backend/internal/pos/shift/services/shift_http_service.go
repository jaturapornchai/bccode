package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/logger"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/pos/shift/models"
	"smlcloudplatform/internal/pos/shift/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"sort"
	"time"

	saleinvoiceservices "smlcloudplatform/internal/transaction/saleinvoice/services"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IShiftHttpService interface {
	CreateShift(shopID string, authUsername string, doc models.Shift) (string, error)
	UpdateShift(shopID string, guid string, authUsername string, doc models.Shift) error
	DeleteShift(shopID string, guid string, authUsername string) error
	DeleteShiftByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoShift(shopID string, guid string) (models.ShiftInfo, error)
	ReportShift(shopID string, docno string) (models.ShiftInfo, error)
	InfoShiftByCode(shopID string, code string) (models.ShiftInfo, error)
	SearchShift(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ShiftInfo, mongopagination.PaginationData, error)
	SearchShiftStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ShiftInfo, int, error)
	SaveInBatch(shopID string, authUsername string, dataList []models.Shift) (common.BulkImport, error)

	GetModuleName() string
}

type ShiftHttpService struct {
	repo          repositories.IShiftRepository
	repoMq        repositories.IShiftMessageQueueRepository
	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.ShiftActivity, models.ShiftDeleteActivity]
	contextTimeout time.Duration

	SaleInvoiceService saleinvoiceservices.ISaleInvoiceService // exported for DI
}

func NewShiftHttpService(
	repo repositories.IShiftRepository,
	repoMq repositories.IShiftMessageQueueRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,

	contextTimeout time.Duration,
) *ShiftHttpService {

	insSvc := &ShiftHttpService{
		repo:          repo,
		repoMq:        repoMq,
		syncCacheRepo: syncCacheRepo,

		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.ShiftActivity, models.ShiftDeleteActivity](repo)

	return insSvc
}

// Add a constructor that accepts SaleInvoiceService
func NewShiftHttpServiceWithSaleInvoice(
	repo repositories.IShiftRepository,
	repoMq repositories.IShiftMessageQueueRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
	contextTimeout time.Duration,
	saleInvoiceService saleinvoiceservices.ISaleInvoiceService,
) *ShiftHttpService {
	insSvc := &ShiftHttpService{
		repo:               repo,
		repoMq:             repoMq,
		syncCacheRepo:      syncCacheRepo,
		contextTimeout:     contextTimeout,
		SaleInvoiceService: saleInvoiceService,
	}
	insSvc.ActivityService = services.NewActivityService[models.ShiftActivity, models.ShiftDeleteActivity](repo)
	return insSvc
}

func (svc ShiftHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc ShiftHttpService) CreateShift(shopID string, authUsername string, doc models.Shift) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	newGuidFixed := utils.NewGUID()

	docData := models.ShiftDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.Shift = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err := svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	go func() {
		svc.saveMasterSync(shopID)
		err = svc.repoMq.Create(docData)
		if err != nil {
			logger.GetLogger().Errorf("Create shift message queue error :: %s", err.Error())
		}
	}()

	return newGuidFixed, nil
}

func (svc ShiftHttpService) UpdateShift(shopID string, guid string, authUsername string, doc models.Shift) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	findDoc.Shift = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, findDoc)

	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(shopID)
		err = svc.repoMq.Update(findDoc)
		if err != nil {
			logger.GetLogger().Errorf("Update shift message queue error :: %s", err.Error())
		}
	}()

	return nil
}

func (svc ShiftHttpService) DeleteShift(shopID string, guid string, authUsername string) error {

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

	go func() {
		svc.saveMasterSync(shopID)
		err = svc.repoMq.Delete(findDoc)
		if err != nil {
			logger.GetLogger().Errorf("Delete creditor message queue error :: %s", err.Error())
		}
	}()

	return nil
}

func (svc ShiftHttpService) DeleteShiftByGUIDs(shopID string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guid_fixed": bson.M{"$in": GUIDs},
	}
	findDocs, err := svc.repo.FindByGuids(ctx, shopID, GUIDs)

	if err != nil {
		return err
	}

	err = svc.repo.Delete(ctx, shopID, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}
	go func() {
		svc.saveMasterSync(shopID)
		err = svc.repoMq.DeleteInBatch(findDocs)
		if err != nil {
			logger.GetLogger().Errorf("Delete creditor message queue error :: %s", err.Error())
		}
	}()

	return nil
}

func (svc ShiftHttpService) InfoShift(shopID string, guid string) (models.ShiftInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.ShiftInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.ShiftInfo{}, errors.New("document not found")
	}

	return findDoc.ShiftInfo, nil
}

// Satisfy old interface for backward compatibility
func (svc *ShiftHttpService) ReportShift(shopID string, docno string) (models.ShiftInfo, error) {
	report, err := svc.ReportShiftReport(shopID, docno)
	if err != nil {
		return models.ShiftInfo{}, err
	}
	if len(report.Shifts) > 0 {
		return models.ShiftInfo{Shift: report.Shifts[0]}, nil // minimal stub for interface
	}
	return models.ShiftInfo{}, errors.New("no shift found in report")
}

func (svc *ShiftHttpService) ReportShiftReport(shopID string, docno string) (models.ShiftReport, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	shiftDocs, err := svc.repo.FindByDocNo(ctx, shopID, docno)
	if err != nil {
		return models.ShiftReport{}, err
	}
	if len(shiftDocs) == 0 {
		return models.ShiftReport{}, errors.New("document not found")
	}

	sort.Slice(shiftDocs, func(i, j int) bool {
		return shiftDocs[i].DocDate.Before(shiftDocs[j].DocDate)
	})

	shifts := make([]models.Shift, len(shiftDocs))
	for i, doc := range shiftDocs {
		shifts[i] = doc.Shift
	}

	cashDrawer := models.ShiftReportCashDrawerSummary{}
	for _, doc := range shiftDocs {
		switch doc.DocType {
		case 1:
			cashDrawer.Open += doc.Amount
		case 3:
			cashDrawer.Add += doc.Amount
		case 4:
			cashDrawer.Withdraw += doc.Amount
		case 2:
			cashDrawer.Close += doc.Amount
		}
	}

	filters := map[string]interface{}{"shiftdocno": docno}
	invoices, _, err := svc.SaleInvoiceService.SearchSaleInvoice(shopID, filters, micromodels.Pageable{Limit: 1000000, Page: 1})
	if err != nil {
		return models.ShiftReport{}, err
	}

	sort.Slice(invoices, func(i, j int) bool {
		return invoices[i].DocDatetime.Before(invoices[j].DocDatetime)
	})

	summary := models.ShiftReportSummary{}
	for _, inv := range invoices {
		summary.TotalAmount += inv.TotalAmount
		summary.Cash += inv.PayCashAmount - inv.PayCashChange
		summary.Credit += inv.SumCreditCard
		summary.Transfer += inv.SumMoneyTransfer
		summary.Cheque += inv.SumCheque
		summary.Coupon += inv.SumCoupon
		summary.QRCode += inv.SumQRCode
	}

	cashDrawer.Expected = cashDrawer.Open + cashDrawer.Add - cashDrawer.Withdraw + summary.Cash
	cashDrawer.Actual = cashDrawer.Close
	cashDrawer.Diff = cashDrawer.Actual - cashDrawer.Expected

	var movements []models.ShiftReportMovement
	for _, doc := range shiftDocs {
		movements = append(movements, models.ShiftReportMovement{
			MovementType: "shift",
			DocNo:        doc.DocNo,
			DocDatetime:  doc.DocDate,
			DocType:      doc.DocType,
			UserCode:     doc.UserCode,
			Username:     doc.Username,
			Remark:       doc.Remark,
			Amount:       doc.Amount,
		})
	}
	for _, inv := range invoices {
		movements = append(movements, models.ShiftReportMovement{
			MovementType:          "saleinvoice",
			DocNo:                 inv.DocNo,
			DocDatetime:           inv.DocDatetime,
			CustCode:              inv.CustCode,
			CustNames:             inv.CustNames,
			SaleCode:              inv.SaleCode,
			SaleName:              inv.SaleName,
			MemberCode:            inv.MemberCode,
			IsCancel:              inv.IsCancel,
			GetPoint:              inv.GetPoint,
			UsePoint:              inv.UsePoint,
			PaymentDetailRaw:      inv.PaymentDetailRaw,
			PayCashAmount:         inv.PayCashAmount,
			PayCashChange:         inv.PayCashChange,
			SumQRCode:             inv.SumQRCode,
			SumCreditCard:         inv.SumCreditCard,
			SumMoneyTransfer:      inv.SumMoneyTransfer,
			SumCheque:             inv.SumCheque,
			SumCoupon:             inv.SumCoupon,
			DetailTotalDiscount:   inv.DetailTotalDiscount,
			DetailDiscountFormula: inv.DetailDiscountFormula,
			RoundAmount:           inv.RoundAmount,
			TotalQty:              inv.TotalQty,
			TotalAmount:           inv.TotalAmount,
		})
	}
	sort.Slice(movements, func(i, j int) bool {
		return movements[i].DocDatetime.Before(movements[j].DocDatetime)
	})

	return models.ShiftReport{
		Shifts:       shifts,
		SaleInvoices: invoices,
		Summary:      summary,
		CashDrawer:   cashDrawer,
		Movements:    movements,
	}, nil
}

func (svc ShiftHttpService) InfoShiftByCode(shopID string, code string) (models.ShiftInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", code)

	if err != nil {
		return models.ShiftInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.ShiftInfo{}, errors.New("document not found")
	}

	return findDoc.ShiftInfo, nil
}

func (svc ShiftHttpService) SearchShift(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.ShiftInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
		"username",
		"remark",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.ShiftInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc ShiftHttpService) SearchShiftStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.ShiftInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
		"username",
		"remark",
	}

	selectFields := map[string]interface{}{}

	/*
		if langCode != "" {
			selectFields["names"] = bson.M{"$elemMatch": bson.M{"code": langCode}}
		} else {
			selectFields["names"] = 1
		}
	*/

	docList, total, err := svc.repo.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.ShiftInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc ShiftHttpService) SaveInBatch(shopID string, authUsername string, dataList []models.Shift) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.Shift](dataList, svc.getDocIDKey)

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

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.Shift, models.ShiftDoc](
		shopID,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(shopID string, authUsername string, doc models.Shift) models.ShiftDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.ShiftDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.ShopID = shopID
			dataDoc.Shift = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.Shift, models.ShiftDoc](
		shopID,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(shopID string, guid string) (models.ShiftDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, shopID, "docno", guid)
		},
		func(doc models.ShiftDoc) bool {
			return doc.DocNo != ""
		},
		func(shopID string, authUsername string, data models.Shift, doc models.ShiftDoc) error {

			doc.Shift = data
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

	go func() {
		svc.saveMasterSync(shopID)
		err = svc.repoMq.CreateInBatch(createDataList)
		if err != nil {
			logger.GetLogger().Errorf("Create shift message queue error :: %s", err.Error())
		}
		svc.repoMq.UpdateInBatch(updateSuccessDataList)

		if err != nil {
			logger.GetLogger().Errorf("Update shift message queue error :: %s", err.Error())
		}
	}()

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

func (svc ShiftHttpService) getDocIDKey(doc models.Shift) string {
	return doc.DocNo
}

func (svc ShiftHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc ShiftHttpService) GetModuleName() string {
	return "shift"
}
