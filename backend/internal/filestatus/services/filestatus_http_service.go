package services

import (
	"context"
	"errors"
	"smlcloudplatform/internal/filestatus/models"
	"smlcloudplatform/internal/filestatus/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
)

type IFileStatusHttpService interface {
	CreateFileStatus(holdingCode string, authUsername string, doc models.FileStatus) (string, error)
	UpdateFileStatus(holdingCode string, guid string, authUsername string, doc models.FileStatus) error
	DeleteFileStatus(holdingCode string, guid string, authUsername string) error
	DeleteFileStatusByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	DeleteFileStatusByMenu(holdingCode string, authUsername string, menu string) error
	InfoFileStatus(holdingCode string, guid string) (models.FileStatusInfo, error)
	SearchFileStatus(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.FileStatusInfo, mongopagination.PaginationData, error)
	SearchFileStatusStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.FileStatusInfo, int, error)
}

type FileStatusHttpService struct {
	repo repositories.IFileStatusRepository
	services.ActivityService[models.FileStatusActivity, models.FileStatusDeleteActivity]
	contextTimeout time.Duration
}

func NewFileStatusHttpService(
	repo repositories.IFileStatusRepository,
	contextTimeout time.Duration,
) *FileStatusHttpService {

	insSvc := &FileStatusHttpService{
		repo: repo,

		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.FileStatusActivity, models.FileStatusDeleteActivity](repo)

	return insSvc
}

func (svc FileStatusHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc FileStatusHttpService) CreateFileStatus(holdingCode string, authUsername string, doc models.FileStatus) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindOne(ctx, holdingCode, bson.M{"menu": doc.Menu, "username": authUsername, "job_id": doc.JobID})

	if err != nil {
		return "", err
	}

	if len(findDoc.GuidFixed) > 0 {
		return "", errors.New("code is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.FileStatusDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.FileStatus = doc

	docData.Username = authUsername
	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	return newGuidFixed, nil
}

func (svc FileStatusHttpService) UpdateFileStatus(holdingCode string, guid string, authUsername string, doc models.FileStatus) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindOne(ctx, holdingCode, bson.M{"guid_fixed": guid})

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	dataDoc := findDoc
	dataDoc.FileStatus = doc

	dataDoc.Menu = findDoc.Menu
	dataDoc.Username = findDoc.Username
	dataDoc.UpdatedBy = authUsername
	dataDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, dataDoc)

	if err != nil {
		return err
	}

	return nil
}

func (svc FileStatusHttpService) DeleteFileStatusByMenu(holdingCode string, authUsername string, menu string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.repo.Delete(ctx, holdingCode, authUsername, bson.M{"menu": menu})
	if err != nil {
		return err
	}

	return nil
}

func (svc FileStatusHttpService) DeleteFileStatus(holdingCode string, guid string, authUsername string) error {

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

func (svc FileStatusHttpService) DeleteFileStatusByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guid_fixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, holdingCode, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	return nil
}

func (svc FileStatusHttpService) InfoFileStatus(holdingCode string, guid string) (models.FileStatusInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.FileStatusInfo{}, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return models.FileStatusInfo{}, errors.New("document not found")
	}

	return findDoc.FileStatusInfo, nil
}

func (svc FileStatusHttpService) SearchFileStatus(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.FileStatusInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, holdingCode, filters, searchInFields, pageable)

	if err != nil {
		return []models.FileStatusInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc FileStatusHttpService) SearchFileStatusStep(holdingCode string, langCode string, filters map[string]interface{}, pageableStep micromodels.PageableStep) ([]models.FileStatusInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"path",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, filters, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.FileStatusInfo{}, 0, err
	}

	return docList, total, nil
}
