package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/shop/branch/models"
	"smlcloudplatform/internal/shop/branch/repositories"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IBranchHttpService interface {
	CreateBranch(shopID string, authUsername string, doc models.Branch) (string, error)
	UpdateBranch(shopID string, guid string, authUsername string, doc models.Branch) error
	DeleteBranch(shopID string, guid string, authUsername string) error
	DeleteBranchByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoBranch(shopID string, guid string) (models.BranchInfo, error)
	SearchBranch(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.BranchInfo, mongopagination.PaginationData, error)
	SearchBranchStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.BranchInfo, int, error)

	GetModuleName() string
}

type BranchHttpService struct {
	repo repositories.IBranchRepository

	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.BranchActivity, models.BranchDeleteActivity]
	contextTimeout time.Duration
}

func NewBranchHttpService(repo repositories.IBranchRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository) *BranchHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &BranchHttpService{
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.BranchActivity, models.BranchDeleteActivity](repo)

	return insSvc
}

func (svc BranchHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc BranchHttpService) CreateBranch(shopID string, authUsername string, doc models.Branch) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if len(findDoc.GuidFixed) > 0 {
		return "", errors.New("Code is exists")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.BranchDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.Branch = doc

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(shopID)

	return newGuidFixed, nil
}

func (svc BranchHttpService) UpdateBranch(shopID string, guid string, authUsername string, doc models.Branch) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	findDoc.Branch = doc

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, findDoc)

	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	return nil
}

func (svc BranchHttpService) DeleteBranch(shopID string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, shopID, guid, authUsername)
	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	return nil
}

func (svc BranchHttpService) DeleteBranchByGUIDs(shopID string, authUsername string, GUIDs []string) error {

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

func (svc BranchHttpService) InfoBranch(shopID string, guid string) (models.BranchInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.BranchInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.BranchInfo{}, errors.New("document not found")
	}

	return findDoc.BranchInfo, nil

}

func (svc BranchHttpService) SearchBranch(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.BranchInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.BranchInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc BranchHttpService) SearchBranchStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.BranchInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
	}

	selectFields := map[string]interface{}{
		"guidfixed": 1,
		"code":      1,
	}

	if langCode != "" {
		selectFields["names"] = bson.M{"$elemMatch": bson.M{"code": langCode}}
	} else {
		selectFields["names"] = 1
	}

	docList, total, err := svc.repo.FindStep(ctx, shopID, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.BranchInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc BranchHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc BranchHttpService) GetModuleName() string {
	return "branch"
}
