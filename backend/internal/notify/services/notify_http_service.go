package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	"smlcloudplatform/internal/notify/models"
	"smlcloudplatform/internal/notify/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type INotifyHttpService interface {
	CreateNotify(holdingCode string, authUsername string, doc models.NotifyRequest) (string, error)
	UpdateNotify(holdingCode string, guid string, authUsername string, doc models.NotifyRequest) error
	DeleteNotify(holdingCode string, guid string, authUsername string) error
	DeleteNotifyByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoNotify(holdingCode string, guid string) (models.NotifyInfo, error)
	InfoNotifyByCode(holdingCode string, code string) (models.NotifyInfo, error)
	SearchNotify(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.Notify, mongopagination.PaginationData, error)
	SearchNotifyInfo(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.NotifyInfo, mongopagination.PaginationData, error)
	SearchNotifyStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.NotifyInfo, int, error)

	GetModuleName() string
}

type NotifyHttpService struct {
	repo repositories.INotifyRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.NotifyActivity, models.NotifyDeleteActivity]
	contextTimeout time.Duration
}

func NewNotifyHttpService(
	repo repositories.INotifyRepository,
	syncCacheRepo mastersync.IMasterSyncCacheRepository,

	contextTimeout time.Duration,
) *NotifyHttpService {

	insSvc := &NotifyHttpService{
		repo:          repo,
		syncCacheRepo: syncCacheRepo,

		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.NotifyActivity, models.NotifyDeleteActivity](repo)

	return insSvc
}

func (svc NotifyHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc NotifyHttpService) CreateNotify(holdingCode string, authUsername string, doc models.NotifyRequest) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	newGuidFixed := utils.NewGUID()

	dataDoc := models.NotifyDoc{}
	dataDoc.HoldingCode = holdingCode
	dataDoc.GuidFixed = newGuidFixed
	dataDoc.Token = doc.Token
	dataDoc.Notify = doc.Notify

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

func (svc NotifyHttpService) UpdateNotify(holdingCode string, guid string, authUsername string, doc models.NotifyRequest) error {

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
	dataDoc.Token = doc.Token
	dataDoc.Notify = doc.Notify

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

func (svc NotifyHttpService) DeleteNotify(holdingCode string, guid string, authUsername string) error {

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

func (svc NotifyHttpService) DeleteNotifyByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

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

func (svc NotifyHttpService) InfoNotify(holdingCode string, guid string) (models.NotifyInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.NotifyInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.NotifyInfo{}, errors.New("document not found")
	}

	return findDoc.NotifyInfo, nil
}

func (svc NotifyHttpService) InfoNotifyByCode(holdingCode string, code string) (models.NotifyInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "uid", code)

	if err != nil {
		return models.NotifyInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.NotifyInfo{}, errors.New("document not found")
	}

	return findDoc.NotifyInfo, nil
}

func (svc NotifyHttpService) SearchNotify(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.Notify, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"name",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.Notify{}, pagination, err
	}

	tempResults := make([]models.Notify, len(docList))
	for _, doc := range docList {
		tempResults = append(tempResults, doc.Notify)
	}

	return tempResults, pagination, nil
}

func (svc NotifyHttpService) SearchNotifyInfo(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.NotifyInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"name",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.NotifyInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc NotifyHttpService) SearchNotifyStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.NotifyInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"name",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.NotifyInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc NotifyHttpService) saveMasterSync(holdingCode string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(holdingCode, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc NotifyHttpService) GetModuleName() string {
	return "notify"
}
