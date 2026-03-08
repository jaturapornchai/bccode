package services

import (
	"context"
	"errors"
	"smlcloudplatform/internal/restaurant/notifier/models"
	"smlcloudplatform/internal/restaurant/notifier/repositories"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type INotifierHttpService interface {
	CreateNotifier(shopID string, authUsername string, doc models.Notifier) (string, error)
	UpdateNotifier(shopID string, guid string, authUsername string, doc models.Notifier) error
	DeleteNotifier(shopID string, guid string, authUsername string) error
	DeleteNotifierByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoNotifier(shopID string, guid string) (models.NotifierInfo, error)
	InfoNotifierByCode(shopID string, code string) (models.NotifierInfo, error)
	SearchNotifier(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.NotifierInfo, mongopagination.PaginationData, error)
}

type NotifierHttpService struct {
	repo            repositories.INotifierRepository
	contextTimeout  time.Duration
	generateRefCode func(int) string
	generateNumber  func(int) string
}

func NewNotifierHttpService(
	repo repositories.INotifierRepository,
	generateRefCode func(int) string,
	generateNumber func(int) string,
	contextTimeout time.Duration,
) *NotifierHttpService {

	insSvc := &NotifierHttpService{
		repo:            repo,
		generateRefCode: generateRefCode,
		generateNumber:  generateNumber,
		contextTimeout:  contextTimeout,
	}

	return insSvc
}

func (svc NotifierHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc NotifierHttpService) CreateNotifier(shopID string, authUsername string, doc models.Notifier) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if len(findDoc.GuidFixed) > 0 {
		return "", errors.New("code is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.NotifierDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.Notifier = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	return newGuidFixed, nil
}

func (svc NotifierHttpService) UpdateNotifier(shopID string, guid string, authUsername string, doc models.Notifier) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	findDoc.Notifier = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, findDoc)

	if err != nil {
		return err
	}

	return nil
}

func (svc NotifierHttpService) DeleteNotifier(shopID string, guid string, authUsername string) error {

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

func (svc NotifierHttpService) DeleteNotifierByGUIDs(shopID string, authUsername string, GUIDs []string) error {

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

func (svc NotifierHttpService) InfoNotifier(shopID string, guid string) (models.NotifierInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.NotifierInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.NotifierInfo{}, errors.New("document not found")
	}

	return findDoc.NotifierInfo, nil
}

func (svc NotifierHttpService) InfoNotifierByCode(shopID string, code string) (models.NotifierInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "usercode", code)

	if err != nil {
		return models.NotifierInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.NotifierInfo{}, errors.New("document not found")
	}

	return findDoc.NotifierInfo, nil
}

func (svc NotifierHttpService) SearchNotifier(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.NotifierInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"usercode",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.NotifierInfo{}, pagination, err
	}

	return docList, pagination, nil
}
