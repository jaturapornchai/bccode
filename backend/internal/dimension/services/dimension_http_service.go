package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/dimension/models"
	"smlcloudplatform/internal/dimension/repositories"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IDimensionHttpService interface {
	CreateDimension(shopID string, authUsername string, doc models.Dimension) (string, error)
	UpdateDimension(shopID string, guid string, authUsername string, doc models.Dimension) error
	DeleteDimension(shopID string, guid string, authUsername string) error
	DeleteDimensionByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoDimension(shopID string, guid string) (models.DimensionInfo, error)
	SearchDimension(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DimensionInfo, mongopagination.PaginationData, error)
	SearchDimensionStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DimensionInfo, int, error)

	GetModuleName() string
}

type DimensionHttpService struct {
	repo repositories.IDimensionRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.DimensionActivity, models.DimensionDeleteActivity]
	contextTimeout time.Duration
}

func NewDimensionHttpService(
	repo repositories.IDimensionRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,

	contextTimeout time.Duration,
) *DimensionHttpService {

	insSvc := &DimensionHttpService{
		repo:          repo,
		syncCacheRepo: syncCacheRepo,

		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.DimensionActivity, models.DimensionDeleteActivity](repo)

	return insSvc
}

func (svc DimensionHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc DimensionHttpService) CreateDimension(shopID string, authUsername string, doc models.Dimension) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	newGuidFixed := utils.NewGUID()

	dataDoc := models.DimensionDoc{}
	dataDoc.ShopID = shopID
	dataDoc.GuidFixed = newGuidFixed
	dataDoc.Dimension = doc

	for i := 0; i < len(dataDoc.Items); i++ {
		dataDoc.Items[i].GuidFixed = utils.NewGUID()
	}

	dataDoc.CreatedBy = authUsername
	dataDoc.CreatedAt = time.Now()

	_, err := svc.repo.Create(ctx, dataDoc)

	if err != nil {
		return "", err
	}

	go func() {
		svc.saveMasterSync(shopID)
	}()

	return newGuidFixed, nil
}

func (svc DimensionHttpService) UpdateDimension(shopID string, guid string, authUsername string, doc models.Dimension) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	dataDoc := findDoc

	itemDict := map[string]struct{}{}
	for i := 0; i < len(findDoc.Items); i++ {
		tempItem := dataDoc.Items[i]
		if tempItem.GuidFixed != "" {
			itemDict[tempItem.GuidFixed] = struct{}{}
		}
	}

	for i := 0; i < len(doc.Items); i++ {
		tempItem := doc.Items[i]

		if _, ok := itemDict[doc.Items[i].GuidFixed]; !ok {
			doc.Items[i].GuidFixed = utils.NewGUID()
		}

		if tempItem.GuidFixed == "" {
			doc.Items[i].GuidFixed = utils.NewGUID()
		}
	}

	dataDoc.Dimension = doc

	dataDoc.GuidFixed = findDoc.GuidFixed
	dataDoc.UpdatedBy = authUsername
	dataDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, dataDoc)

	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc DimensionHttpService) DeleteDimension(shopID string, guid string, authUsername string) error {

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
	}()

	return nil
}

func (svc DimensionHttpService) DeleteDimensionByGUIDs(shopID string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, shopID, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(shopID)
	}()

	return nil
}

func (svc DimensionHttpService) InfoDimension(shopID string, guid string) (models.DimensionInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.DimensionInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.DimensionInfo{}, errors.New("document not found")
	}

	return findDoc.DimensionInfo, nil
}

func (svc DimensionHttpService) SearchDimension(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DimensionInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"guidfixed",
		"names",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.DimensionInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc DimensionHttpService) SearchDimensionStep(shopID string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DimensionInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"guidfixed",
		"names",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.DimensionInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc DimensionHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc DimensionHttpService) GetModuleName() string {
	return "dimension"
}
