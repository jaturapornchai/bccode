package services

import (
	"context"
	"errors"
	"smlcloudplatform/internal/restaurant/notifierdevice/models"
	"smlcloudplatform/internal/restaurant/notifierdevice/repositories"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type INotifierDeviceHttpService interface {
	CreateAuthCode(holdingCode string, authUsername string) (models.NotifierDeviceAuth, error)
	ConfirmAuthCode(payload models.NotifierDeviceConfirmAuthPayload) (bool, error)

	UpdateNotifierDevice(holdingCode string, guid string, authUsername string, doc models.NotifierDevice) error
	DeleteNotifierDevice(holdingCode string, guid string, authUsername string) error
	DeleteNotifierDeviceByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoNotifierDevice(holdingCode string, guid string) (models.NotifierDeviceInfo, error)
	SearchNotifierDevice(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.NotifierDeviceInfo, mongopagination.PaginationData, error)
}

type NotifierDeviceHttpService struct {
	repo            repositories.INotifierDeviceRepository
	cacheRepo       repositories.INotifierDeviceCacheRepository
	contextTimeout  time.Duration
	generateRefCode func(int) string
	generateNumber  func(int) string
}

func NewNotifierDeviceHttpService(
	repo repositories.INotifierDeviceRepository,
	cacheRepo repositories.INotifierDeviceCacheRepository,
	generateRefCode func(int) string,
	generateNumber func(int) string,
	contextTimeout time.Duration,
) *NotifierDeviceHttpService {

	insSvc := &NotifierDeviceHttpService{
		repo:            repo,
		cacheRepo:       cacheRepo,
		generateRefCode: generateRefCode,
		generateNumber:  generateNumber,
		contextTimeout:  contextTimeout,
	}

	return insSvc
}

func (svc NotifierDeviceHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc NotifierDeviceHttpService) CreateAuthCode(holdingCode string, authUsername string) (models.NotifierDeviceAuth, error) {

	refCode := svc.generateRefCode(8)

	notifierAuth := models.NotifierDeviceAuth{
		HoldingCode: holdingCode,
		UserAddedBy: authUsername,
		RefCode:     refCode,
	}

	err := svc.cacheRepo.Save(refCode, notifierAuth, 5*time.Minute)

	if err != nil {
		return models.NotifierDeviceAuth{}, err
	}

	return notifierAuth, nil
}

func (svc NotifierDeviceHttpService) ConfirmAuthCode(payload models.NotifierDeviceConfirmAuthPayload) (bool, error) {

	notifierAuth, err := svc.cacheRepo.Get(payload.RefCode)

	if err != nil {
		return false, err
	}

	if notifierAuth.RefCode == "" {
		return false, errors.New("refcode invalid")
	}

	_, err = svc.CreateNotifierDevice(notifierAuth.HoldingCode, notifierAuth.UserAddedBy, models.NotifierDevice{
		FCMToken:   payload.FCMToken,
		DeviceID:   payload.DeviceID,
		DeviceName: payload.DeviceName,
	})

	if err != nil {
		return false, err
	}

	return true, nil
}

func (svc NotifierDeviceHttpService) CreateNotifierDevice(holdingCode string, authUsername string, doc models.NotifierDevice) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "fcmtoken", doc.FCMToken)

	if err != nil {
		return "", err
	}

	if len(findDoc.GuidFixed) > 0 {
		return "", errors.New("fcm token is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.NotifierDeviceDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.NotifierDevice = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	return newGuidFixed, nil
}

func (svc NotifierDeviceHttpService) UpdateNotifierDevice(holdingCode string, guid string, authUsername string, doc models.NotifierDevice) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	findDoc.NotifierDevice = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, findDoc)

	if err != nil {
		return err
	}

	return nil
}

func (svc NotifierDeviceHttpService) DeleteNotifierDevice(holdingCode string, guid string, authUsername string) error {

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

	return nil
}

func (svc NotifierDeviceHttpService) DeleteNotifierDeviceByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guidfixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, holdingCode, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	return nil
}

func (svc NotifierDeviceHttpService) InfoNotifierDevice(holdingCode string, guid string) (models.NotifierDeviceInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.NotifierDeviceInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.NotifierDeviceInfo{}, errors.New("document not found")
	}

	return findDoc.NotifierDeviceInfo, nil
}

func (svc NotifierDeviceHttpService) SearchNotifierDevice(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.NotifierDeviceInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"usercode",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.NotifierDeviceInfo{}, pagination, err
	}

	return docList, pagination, nil
}
