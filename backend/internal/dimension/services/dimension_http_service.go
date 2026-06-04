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
	CreateDimension(holdingCode string, authUsername string, doc models.Dimension) (string, error)
	UpdateDimension(holdingCode string, guid string, authUsername string, doc models.Dimension) error
	DeleteDimension(holdingCode string, guid string, authUsername string) error
	DeleteDimensionByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoDimension(holdingCode string, guid string) (models.DimensionInfo, error)
	SearchDimension(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DimensionInfo, mongopagination.PaginationData, error)
	SearchDimensionStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DimensionInfo, int, error)

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

func (svc DimensionHttpService) CreateDimension(holdingCode string, authUsername string, doc models.Dimension) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	newGuidFixed := utils.NewGUID()

	dataDoc := models.DimensionDoc{}
	dataDoc.HoldingCode = holdingCode
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
		svc.saveMasterSync(holdingCode)
	}()

	return newGuidFixed, nil
}

func (svc DimensionHttpService) UpdateDimension(holdingCode string, guid string, authUsername string, doc models.Dimension) error {

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

	err = svc.repo.Update(ctx, holdingCode, guid, dataDoc)

	if err != nil {
		return err
	}

	go func() {
		svc.saveMasterSync(holdingCode)
	}()

	return nil
}

func (svc DimensionHttpService) DeleteDimension(holdingCode string, guid string, authUsername string) error {

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

func (svc DimensionHttpService) DeleteDimensionByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guid_fixed": bson.M{"$in": GUIDs},
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

func (svc DimensionHttpService) InfoDimension(holdingCode string, guid string) (models.DimensionInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.DimensionInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.DimensionInfo{}, errors.New("document not found")
	}

	return findDoc.DimensionInfo, nil
}

func (svc DimensionHttpService) SearchDimension(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DimensionInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"guid_fixed",
		"names",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.DimensionInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc DimensionHttpService) SearchDimensionStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.DimensionInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"guid_fixed",
		"names",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.DimensionInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc DimensionHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc DimensionHttpService) GetModuleName() string {
	return "dimension"
}
