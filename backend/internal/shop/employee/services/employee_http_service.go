package services

import (
	"context"
	"errors"
	"fmt"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/shop/employee/models"
	"smlcloudplatform/internal/shop/employee/repositories"
	"smlcloudplatform/internal/utils"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IEmployeeHttpService interface {
	CreateEmployee(shopID string, authUsername string, doc models.EmployeeRequestRegister) (string, error)
	UpdateEmployee(shopID string, guid string, authUsername string, doc models.EmployeeRequestUpdate) error
	UpdatePassword(shopID string, authUsername string, emp models.EmployeeRequestPassword) error
	DeleteEmployee(shopID string, guid string, authUsername string) error
	DeleteEmployeeByGUIDs(shopID string, authUsername string, GUIDs []string) error
	InfoEmployee(shopID string, guid string) (models.EmployeeInfo, error)
	InfoEmployeeByCode(shopID string, code string) (models.EmployeeInfo, error)
	InfoEmployeeByEmail(shopID string, email string) (models.EmployeeInfo, error)
	SearchEmployee(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.EmployeeInfo, mongopagination.PaginationData, error)
	SearchEmployeeStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.EmployeeInfo, int, error)

	GetModuleName() string
}

type EmployeeHttpService struct {
	hashPassword  func(string) (string, error)
	repo          repositories.IEmployeeRepository
	syncCacheRepo mastersync.IMasterSyncCacheRepository
	services.ActivityService[models.EmployeeActivity, models.EmployeeDeleteActivity]
	contextTimeout time.Duration
}

func NewEmployeeHttpService(repo repositories.IEmployeeRepository, syncCacheRepo mastersync.IMasterSyncCacheRepository, hashPasswordFunc func(string) (string, error)) *EmployeeHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &EmployeeHttpService{
		hashPassword:   hashPasswordFunc,
		repo:           repo,
		syncCacheRepo:  syncCacheRepo,
		contextTimeout: contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.EmployeeActivity, models.EmployeeDeleteActivity](repo)

	return insSvc
}

func (svc EmployeeHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc EmployeeHttpService) CreateEmployee(shopID string, authUsername string, doc models.EmployeeRequestRegister) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if findDoc.Code != "" {
		return "", errors.New("code is exists")
	}

	hashedPassword, err := utils.HashPassword(doc.Password)

	if err != nil {
		return "", err
	}

	newGuidFixed := utils.NewGUID()

	docData := models.EmployeeDoc{}
	docData.ShopID = shopID
	docData.GuidFixed = newGuidFixed
	docData.Employee = doc.Employee
	docData.Password = hashedPassword

	docData.CreatedBy = authUsername
	docData.CreatedAt = time.Now()

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	svc.saveMasterSync(shopID)

	return newGuidFixed, nil
}

func (svc EmployeeHttpService) UpdateEmployee(shopID string, guid string, authUsername string, doc models.EmployeeRequestUpdate) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		findDoc, err = svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", guid)
		if err != nil {
			return err
		}

		if findDoc.ID == primitive.NilObjectID {
			return errors.New("document not found")
		}
	}

	docData := findDoc
	docData.Employee = doc.Employee

	docData.Code = doc.Code

	findDoc.UpdatedBy = authUsername
	findDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, guid, docData)

	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	return nil
}

func (svc EmployeeHttpService) UpdatePassword(shopID string, authUsername string, emp models.EmployeeRequestPassword) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	userFind, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", emp.Code)
	if err != nil && err.Error() != "mongo: no documents in result" {
		return err
	}

	if len(userFind.Code) < 1 {
		return errors.New("user or password invalid")
	}

	isCurrentPasswordInvalid := utils.CheckHashPassword(emp.CurrentPassword, userFind.Password)

	if !isCurrentPasswordInvalid {
		return errors.New("user or password invalid")
	}

	hashPassword, err := utils.HashPassword(emp.NewPassword)

	if err != nil {
		return err
	}

	userFind.Password = hashPassword

	userFind.UpdatedBy = authUsername
	userFind.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, shopID, userFind.GuidFixed, userFind)

	if err != nil {
		return err
	}

	svc.saveMasterSync(shopID)

	return nil
}

func (svc EmployeeHttpService) DeleteEmployee(shopID string, guid string, authUsername string) error {

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

func (svc EmployeeHttpService) DeleteEmployeeByGUIDs(shopID string, authUsername string, GUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	deleteFilterQuery := map[string]interface{}{
		"guid_fixed": bson.M{"$in": GUIDs},
	}

	err := svc.repo.Delete(ctx, shopID, authUsername, deleteFilterQuery)
	if err != nil {
		return err
	}

	return nil
}

func (svc EmployeeHttpService) InfoEmployeeByCode(shopID string, code string) (models.EmployeeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "code", code)

	if err != nil {
		return models.EmployeeInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.EmployeeInfo{}, errors.New("document not found")
	}

	return findDoc.EmployeeInfo, nil

}

func (svc EmployeeHttpService) InfoEmployeeByEmail(shopID string, email string) (models.EmployeeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, shopID, "email", email)

	if err != nil {
		return models.EmployeeInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.EmployeeInfo{}, errors.New("document not found")
	}

	return findDoc.EmployeeInfo, nil

}

func (svc EmployeeHttpService) InfoEmployee(shopID string, guid string) (models.EmployeeInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, shopID, guid)

	if err != nil {
		return models.EmployeeInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.EmployeeInfo{}, errors.New("document not found")
	}

	return findDoc.EmployeeInfo, nil
}

func (svc EmployeeHttpService) SearchEmployee(shopID string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.EmployeeInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"name",
		"contact.phonenumber",
	}

	docList, pagination, err := svc.repo.FindPageFilter(ctx, shopID, filters, searchInFields, pageable)

	if err != nil {
		return []models.EmployeeInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc EmployeeHttpService) SearchEmployeeStep(shopID string, langCode string, pageableStep micromodels.PageableStep) ([]models.EmployeeInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"code",
		"name",
		"contact.phonenumber",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, shopID, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.EmployeeInfo{}, 0, err
	}

	return docList, total, nil
}

func (svc EmployeeHttpService) saveMasterSync(shopID string) {
	if svc.syncCacheRepo != nil {
		err := svc.syncCacheRepo.Save(shopID, svc.GetModuleName())

		if err != nil {
			fmt.Printf("save %s cache error :: %s", svc.GetModuleName(), err.Error())
		}
	}
}

func (svc EmployeeHttpService) GetModuleName() string {
	return "employee"
}
