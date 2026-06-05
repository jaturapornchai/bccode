package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	"smlcloudplatform/internal/purchasetype/models"
	"smlcloudplatform/internal/purchasetype/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IPurchaseTypeHttpService interface {
	CreatePurchaseType(holdingCode string, authUsername string, doc models.PurchaseType) (string, error)
	UpdatePurchaseType(holdingCode string, guid string, authUsername string, doc models.PurchaseType) error
	DeletePurchaseType(holdingCode string, guid string, authUsername string) error
	DeletePurchaseTypeByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoPurchaseType(holdingCode string, guid string) (models.PurchaseTypeInfo, error)
	InfoPurchaseTypeByCode(holdingCode string, code string) (models.PurchaseTypeInfo, error)
	SearchPurchaseType(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PurchaseTypeInfo, mongopagination.PaginationData, error)
	SearchPurchaseTypeStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PurchaseTypeInfo, int, error)

	GetModuleName() string
}

type PurchaseTypeHttpService struct {
	repo          repositories.IPurchaseTypeRepository
	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.PurchaseTypeActivity, models.PurchaseTypeDeleteActivity]
	contextTimeout time.Duration
}

func NewPurchaseTypeHttpService(
	repo repositories.IPurchaseTypeRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,
	contextTimeout time.Duration,
) *PurchaseTypeHttpService {

	insSvc := &PurchaseTypeHttpService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.PurchaseTypeActivity, models.PurchaseTypeDeleteActivity](repo)

	return insSvc
}

func (svc PurchaseTypeHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc PurchaseTypeHttpService) CreatePurchaseType(holdingCode string, authUsername string, doc models.PurchaseType) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// Check if code already exists
	findDoc, _ := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", doc.Code)
	if len(findDoc.GuidFixed) > 0 {
		return "", errors.New("code already exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.PurchaseTypeDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.PurchaseType = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err := svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	go func() {
		svc.saveMasterSync(holdingCode)
	}()

	return newGuidFixed, nil
}

func (svc PurchaseTypeHttpService) UpdatePurchaseType(holdingCode string, guid string, authUsername string, doc models.PurchaseType) error {

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
	dataDoc.PurchaseType = doc

	dataDoc.Code = findDoc.Code // Keep original code
	dataDoc.UpdatedBy = authUsername
	dataDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, dataDoc)

	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc PurchaseTypeHttpService) DeletePurchaseType(holdingCode string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, holdingCode, guid, authUsername)
	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc PurchaseTypeHttpService) DeletePurchaseTypeByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, holdingCode, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc PurchaseTypeHttpService) InfoPurchaseType(holdingCode string, guid string) (models.PurchaseTypeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.PurchaseTypeInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.PurchaseTypeInfo{}, errors.New("document not found")
	}

	return findDoc.PurchaseTypeInfo, nil
}

func (svc PurchaseTypeHttpService) InfoPurchaseTypeByCode(holdingCode string, code string) (models.PurchaseTypeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", code)

	if err != nil {
		return models.PurchaseTypeInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.PurchaseTypeInfo{}, errors.New("document not found")
	}

	return findDoc.PurchaseTypeInfo, nil
}

func (svc PurchaseTypeHttpService) SearchPurchaseType(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.PurchaseTypeInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.PurchaseTypeInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc PurchaseTypeHttpService) SearchPurchaseTypeStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.PurchaseTypeInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.PurchaseTypeInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc PurchaseTypeHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc PurchaseTypeHttpService) GetModuleName() string {
	return "purchaseType"
}
