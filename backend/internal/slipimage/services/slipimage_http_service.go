package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/logger"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/slipimage/models"
	"smlcloudplatform/internal/slipimage/repositories"
	saleInvoiceServices "smlcloudplatform/internal/transaction/saleinvoice/services"
	saleInvoiceReturnServices "smlcloudplatform/internal/transaction/saleinvoicereturn/services"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strings"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type ISlipImageHttpService interface {
	CreateSlipImage(shopID string, authUsername string, doc models.SlipImageRequest) (models.SlipImageInfo, error)
	DeleteSlipImage(shopID string, guid string, authUsername string) error
	DeleteSlipImageByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoSlipImage(shopID string, guid string) (models.SlipImageInfo, error)
	InfoSlipImageByDocno(shopID string, mode uint8, docNo string) ([]models.SlipImageInfo, error)
	SearchSlipImage(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SlipImageInfo, mongopagination.PaginationData, error)
	SearchSlipImageStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SlipImageInfo, int, error)
}

type SlipImageHttpService struct {
	svcSaleInvoice       saleInvoiceServices.ISaleInvoiceService
	svcSaleInvoiceReturn saleInvoiceReturnServices.ISaleInvoiceReturnService
	repo                 repositories.ISlipImageMongoRepository
	repoStorageImage     repositories.ISlipImageStorageImageRepository
	syncCacheRepo        mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.SlipImageActivity, models.SlipImageDeleteActivity]
	contextTimeout time.Duration
}

func NewSlipImageHttpService(
	repo repositories.ISlipImageMongoRepository,
	repoStorageImage repositories.ISlipImageStorageImageRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
	svcSaleInvoice saleInvoiceServices.ISaleInvoiceService,
	svcSaleInvoiceReturn saleInvoiceReturnServices.ISaleInvoiceReturnService,
	contextTimeout time.Duration,
) *SlipImageHttpService {

	insSvc := &SlipImageHttpService{
		repo:                 repo,
		repoStorageImage:     repoStorageImage,
		syncCacheRepo:        syncCacheRepo,
		svcSaleInvoice:       svcSaleInvoice,
		svcSaleInvoiceReturn: svcSaleInvoiceReturn,
		contextTimeout:       contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.SlipImageActivity, models.SlipImageDeleteActivity](repo)

	return insSvc
}

func (svc SlipImageHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc SlipImageHttpService) CreateSlipImage(shopID string, authUsername string, payload models.SlipImageRequest) (models.SlipImageInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	fileSize := payload.File.Size

	newGuidFixed := utils.NewGUID()

	// layout := "2006-01-02"
	// docDateStr := payload.DocDate.Format(layout)

	tempSplitFileName := strings.Split(payload.File.Filename, ".")
	fileExt := tempSplitFileName[len(tempSplitFileName)-1]

	// fileName := fmt.Sprintf("%s/slip/%s/%s/%s", shopID, payload.PosID, docDateStr, payload.DocNo)
	fileNameUUId := fmt.Sprintf("%s-%s", utils.NewUUID(), utils.NewUUID())
	fileName := fmt.Sprintf("slip/%s", fileNameUUId)

	uploadUri, err := svc.repoStorageImage.Upload(payload.File, fileName, fileExt)

	if err != nil {
		return models.SlipImageInfo{}, fmt.Errorf("upload file failed")
	}

	mode := uint8(0)

	if payload.Mode == 1 {
		mode = 1
	}

	docData := models.SlipImageDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.SlipImage = models.SlipImage{
		Mode:            mode,
		URI:             uploadUri,
		Size:            fileSize,
		DocNo:           payload.DocNo,
		DocDate:         payload.DocDate,
		PosID:           payload.PosID,
		MachineCode:     payload.MachineCode,
		BranchCode:      payload.BranchCode,
		ZoneGroupNumber: payload.ZoneGroupNumber,
	}

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return models.SlipImageInfo{}, err
	}

	err = svc.svcSaleInvoice.UpdateSlip(shopID, authUsername, payload.DocNo, mode, payload.MachineCode, payload.ZoneGroupNumber, uploadUri)

	if err != nil {
		logger.GetLogger().Error(err)
	}

	err = svc.svcSaleInvoiceReturn.UpdateSlip(shopID, authUsername, payload.DocNo, mode, payload.MachineCode, payload.ZoneGroupNumber, uploadUri)

	if err != nil {
		logger.GetLogger().Error(err)
	}

	return docData.SlipImageInfo, nil
}

func (svc SlipImageHttpService) DeleteSlipImage(shopID string, guid string, authUsername string) error {

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

	return nil
}

func (svc SlipImageHttpService) DeleteSlipImageByGUIDs(shopID string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, shopID, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	return nil
}

func (svc SlipImageHttpService) InfoSlipImage(shopID string, guid string) (models.SlipImageInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.SlipImageInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.SlipImageInfo{}, errors.New("document not found")
	}

	return findDoc.SlipImageInfo, nil
}

func (svc SlipImageHttpService) InfoSlipImageByDocno(shopID string, mode uint8, docNo string) ([]models.SlipImageInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDocs, err := svc.repo.FindByDocNo(ctx, shopID, mode, docNo)

	if err != nil {
		return []models.SlipImageInfo{}, err
	}

	if len(findDocs) < 1 {
		return []models.SlipImageInfo{}, errors.New("document not found")
	}

	return findDocs, nil
}

func (svc SlipImageHttpService) SearchSlipImage(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.SlipImageInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.SlipImageInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc SlipImageHttpService) SearchSlipImageStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.SlipImageInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"docno",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.SlipImageInfo{}, 0, err
	}

	return docList, total, nil
}
