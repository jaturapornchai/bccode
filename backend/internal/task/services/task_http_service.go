package services

import (
	"context"
	"crypto/md5"
	"errors"
	"fmt"
	documentImageModel "smlcloudplatform/internal/documentwarehouse/documentimage/models"
	repositoriesDocumentImage "smlcloudplatform/internal/documentwarehouse/documentimage/repositories"
	servicesDocumentImage "smlcloudplatform/internal/documentwarehouse/documentimage/services"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/services"
	"smlcloudplatform/internal/task/models"
	"smlcloudplatform/internal/task/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/importdata"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strconv"
	"strings"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type ITaskHttpService interface {
	GenerateTaskID(holdingCode string, authUsername string) (string, error)
	CreateTask(holdingCode string, authUsername string, doc models.Task) (string, error)
	UpdateTask(holdingCode string, guid string, authUsername string, doc models.Task) error
	UpdateTaskStatus(holdingCode string, guid string, authUsername string, jobStatus int8) error
	DeleteTask(holdingCode string, guid string, authUsername string) error
	DeleteTaskByGUIDs(holdingCode string, authUsername string, GUIDs []string) error
	InfoTask(holdingCode string, guid string) (models.TaskInfo, error)
	SearchTask(holdingCode string, module string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.TaskInfo, mongopagination.PaginationData, error)
	SearchTaskStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.TaskInfo, int, error)
	SaveInBatch(holdingCode string, authUsername string, dataList []models.Task) (common.BulkImport, error)

	TaskStatusReference() map[int]string
	GetTaskReject(holdingCode string, module string, taskGUID string) ([]models.TaskInfo, error)

	UpdateTaskTotalImage(docReq documentImageModel.DocumentImageTaskChangeMessage) error
	UpdateTaskTotalRejectImage(docReq documentImageModel.DocumentImageTaskRejectMessage) error
}

type TaskHttpService struct {
	repo              repositories.ITaskRepository
	repoDocImageGroup repositoriesDocumentImage.IDocumentImageGroupRepository
	serviceDocImage   servicesDocumentImage.IDocumentImageService
	services.ActivityService[models.TaskActivity, models.TaskDeleteActivity]
	contextTimeout time.Duration
}

func NewTaskHttpService(repo repositories.ITaskRepository, repoDocImageGroup repositoriesDocumentImage.IDocumentImageGroupRepository, serviceDocImage servicesDocumentImage.IDocumentImageService) *TaskHttpService {

	contextTimeout := time.Duration(15) * time.Second

	insSvc := &TaskHttpService{
		repo:              repo,
		repoDocImageGroup: repoDocImageGroup,
		serviceDocImage:   serviceDocImage,
		contextTimeout:    contextTimeout,
	}

	insSvc.ActivityService = services.NewActivityService[models.TaskActivity, models.TaskDeleteActivity](repo)

	return insSvc
}

func (svc TaskHttpService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc TaskHttpService) TaskStatusReference() map[int]string {
	return map[int]string{
		models.TaskPending:   "Pending",
		models.TaskUplaoded:  "Uploaded",
		models.TaskChecking:  "Checking",
		models.TaskCompleted: "Completed",
	}
}

func (svc TaskHttpService) GenerateTaskID(holdingCode string, authUsername string) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	timeNow := time.Now()

	userHash := fmt.Sprintf("%x", md5.Sum([]byte(authUsername)))
	userHashLength := len(userHash)
	userFmt := userHash[(userHashLength - 4):userHashLength]
	timeFmt := timeNow.Format("060102")

	codeFmt := fmt.Sprintf("%s%s-", userFmt, timeFmt)

	findDoc, err := svc.repo.FindLastTaskByCode(ctx, holdingCode, codeFmt)
	if err != nil {
		return "", err
	}

	if len(findDoc.Code) == 0 {
		return codeFmt + "001", nil
	}

	lastTaskCode := findDoc.Code

	splitStrTaskCode := strings.Split(lastTaskCode, "-")

	lastNumberStr := splitStrTaskCode[len(splitStrTaskCode)-1]

	lastNumber, err := strconv.Atoi(lastNumberStr)

	if err != nil {
		return "", err
	}

	lastNumber++

	return codeFmt + fmt.Sprintf("%03d", lastNumber), nil
}

func (svc TaskHttpService) PaddingNumber(number int) string {
	return fmt.Sprintf("%03d", number)
}

func (svc TaskHttpService) CreateTask(holdingCode string, authUsername string, doc models.Task) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "name", doc.Name)

	if err != nil {
		return "", err
	}

	if findDoc.Name != "" {
		return "", errors.New("name is duplicated")
	}

	findDocCode, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", doc.Code)

	if err != nil {
		return "", err
	}

	if findDocCode.Code == doc.Code {
		return "", errors.New("code is duplicated")
	}

	newGuidFixed := utils.NewGUID()

	docData := models.TaskDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = newGuidFixed
	docData.Task = doc
	docData.Status = models.TaskPending

	timeNow := time.Now()

	docData.OwnerBy = authUsername
	docData.OwnerAt = timeNow

	docData.CreatedBy = authUsername
	docData.CreatedAt = timeNow

	_, err = svc.repo.Create(ctx, docData)

	if err != nil {
		return "", err
	}

	return newGuidFixed, nil
}

func (svc TaskHttpService) UpdateTask(holdingCode string, guid string, authUsername string, doc models.Task) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	updateDoc := findDoc
	updateDoc.Task = doc
	updateDoc.Status = findDoc.Status
	updateDoc.Code = findDoc.Code

	updateDoc.TotalDocument = findDoc.TotalDocument
	updateDoc.TotalDocumentStatus = findDoc.TotalDocumentStatus

	updateDoc.RejectFromTaskGUID = findDoc.RejectFromTaskGUID
	updateDoc.RejectedAt = findDoc.RejectedAt
	updateDoc.RejectedBy = findDoc.RejectedBy

	updateDoc.OwnerBy = findDoc.OwnerBy
	updateDoc.OwnerAt = findDoc.OwnerAt

	updateDoc.UpdatedBy = authUsername
	updateDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, guid, updateDoc)

	if err != nil {
		return err
	}

	return nil
}

func (svc TaskHttpService) UpdateTaskStatus(holdingCode string, taskGUID string, authUsername string, jobStatus int8) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, taskGUID)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	if findDoc.Status >= jobStatus {
		return errors.New("task status invalid")
	}

	if jobStatus < models.TaskPending || jobStatus > models.TaskCanceled {
		return errors.New("task status out of range")
	}

	totalImageGroup := 0
	totalRejectImageGroup := 0
	if jobStatus == models.TaskCompleted {
		findDocImageGroups, err := svc.repoDocImageGroup.FindByTaskGUID(ctx, holdingCode, taskGUID)

		if err != nil {
			return err
		}

		for _, docImageGroup := range findDocImageGroups {
			if docImageGroup.Status == documentImageModel.IMAGE_REJECT {
				totalRejectImageGroup += 1
			}

			totalImageGroup += 1
		}

		err = svc.repoDocImageGroup.UpdateTaskIsCompletedByTaskGUID(ctx, holdingCode, findDoc.GuidFixed, true)
		if err != nil {
			return err
		}
	}

	totalImageGroup = 0
	totalRejectKeyingImageGroup := 0
	if jobStatus == models.TaskGlCompleted {
		findDocImageGroups, err := svc.repoDocImageGroup.FindByTaskGUID(ctx, holdingCode, taskGUID)

		if err != nil {
			return err
		}

		for _, docImageGroup := range findDocImageGroups {
			if docImageGroup.Status == documentImageModel.IMAGE_REJECT_KEYING {
				totalRejectKeyingImageGroup += 1
			}

			totalImageGroup += 1
		}
	}

	if totalRejectImageGroup > 0 || totalRejectKeyingImageGroup > 0 {
		newTaskGuidFixed := utils.NewGUID()

		docData := models.TaskDoc{}
		docData.HoldingCode = holdingCode
		docData.GuidFixed = newTaskGuidFixed
		docData.Module = findDoc.Module

		parentGUID := findDoc.GuidFixed
		if len(findDoc.ParentGUIDFixed) > 0 {
			parentGUID = findDoc.ParentGUIDFixed
		}

		if len(findDoc.ParentGUIDFixed) > 0 {
			parentGUID = findDoc.ParentGUIDFixed
			docData.Path = findDoc.Path
		} else {
			docData.Path = fmt.Sprintf("%s/%s", findDoc.Path, findDoc.GuidFixed)
		}

		taskCount, err := svc.repo.CountTaskParent(ctx, holdingCode, parentGUID)

		if err != nil {
			taskCount = 0
		}

		taskCount += 1
		rejectTaskName := fmt.Sprintf("%s - [%d]", findDoc.Task.Name, taskCount)

		newTaskCode := ""
		for i := 0; i < 5; i++ {
			newTaskCode, err = svc.GenerateTaskID(holdingCode, authUsername)
			if err != nil {
				return err
			}

			findDocCode, err := svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "code", newTaskCode)

			if err != nil {
				return err
			}

			if findDocCode.Code != newTaskCode {
				break
			}
		}

		docData.Code = newTaskCode

		docData.Name = rejectTaskName
		docData.ParentGUIDFixed = parentGUID
		docData.RejectFromTaskGUID = findDoc.GuidFixed

		docData.Status = models.TaskPending

		timeNow := time.Now()

		docData.OwnerBy = findDoc.OwnerBy
		docData.OwnerAt = timeNow

		docData.RejectedBy = authUsername
		docData.RejectedAt = timeNow

		docData.CreatedBy = authUsername
		docData.CreatedAt = timeNow

		_, err = svc.repo.Create(ctx, docData)

		if err != nil {
			return err
		}

		if jobStatus == models.TaskCompleted {
			findDocImageGroups, err := svc.repoDocImageGroup.FindByTaskGUID(ctx, holdingCode, taskGUID)

			if err != nil {
				return err
			}

			for _, docImageGroup := range findDocImageGroups {

				if docImageGroup.Status == documentImageModel.IMAGE_REJECT {
					newDocImgGroupGUID := utils.NewGUID()
					for _, docImage := range *docImageGroup.ImageReferences {
						docImgReq := documentImageModel.DocumentImageRequest{}

						docImgReq.DocumentImageGroupGUID = newDocImgGroupGUID
						docImgReq.TaskGUID = newTaskGuidFixed
						docImgReq.ImageURI = docImage.ImageURI
						docImgReq.Name = docImage.Name // + " -- REJECT"

						_, _, err := svc.serviceDocImage.CreateDocumentImage(holdingCode, authUsername, docImgReq)

						if err != nil {
							fmt.Println(err)
						}
					}
				}
			}
		}
	}

	updateDoc := findDoc
	updateDoc.Status = jobStatus
	// findDoc.ToTal = totalImageGroup
	// findDoc.ToTalReject = totalRejectImageGroup

	updateDoc.UpdatedBy = authUsername
	updateDoc.UpdatedAt = time.Now()

	err = svc.repo.Update(ctx, holdingCode, taskGUID, updateDoc)

	if err != nil {
		return err
	}

	return nil
}

func (svc TaskHttpService) DeleteTask(holdingCode string, guid string, authUsername string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document not found")
	}

	err = svc.repo.DeleteByGuidfixed(ctx, holdingCode, guid, authUsername)
	if err != nil {
		return err
	}

	return nil
}

func (svc TaskHttpService) DeleteTaskByGUIDs(holdingCode string, authUsername string, GUIDs []string) error {

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

func (svc TaskHttpService) InfoTask(holdingCode string, guid string) (models.TaskInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repo.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.TaskInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.TaskInfo{}, errors.New("document not found")
	}

	findDocChild, err := svc.repo.FindTaskChild(ctx, holdingCode, findDoc.GuidFixed)

	if err != nil {
		return models.TaskInfo{}, err
	}

	findDoc.TaskChild = findDocChild

	return findDoc.TaskInfo, nil

}

func (svc TaskHttpService) GetTaskReject(holdingCode string, module string, taskGUID string) ([]models.TaskInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docList, err := svc.repo.FindPageByTaskReject(ctx, holdingCode, module, taskGUID)

	if err != nil {
		return []models.TaskInfo{}, err
	}

	return docList, nil
}

func (svc TaskHttpService) SearchTask(holdingCode string, module string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.TaskInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"name",
	}

	docList, pagination, err := svc.repo.FindPageTask(ctx, holdingCode, module, filters, searchInFields, pageable)

	if err != nil {
		return []models.TaskInfo{}, pagination, err
	}

	for i, doc := range docList {
		findDocChild, _ := svc.repo.FindTaskChild(ctx, holdingCode, doc.GuidFixed)
		docList[i].TaskChild = findDocChild
	}

	return docList, pagination, nil
}

func (svc TaskHttpService) SearchTaskStep(holdingCode string, langCode string, pageableStep micromodels.PageableStep) ([]models.TaskInfo, int, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{
		"name",
	}

	selectFields := map[string]interface{}{}

	docList, total, err := svc.repo.FindStep(ctx, holdingCode, map[string]interface{}{}, searchInFields, selectFields, pageableStep)

	if err != nil {
		return []models.TaskInfo{}, 0, err
	}

	for i, doc := range docList {
		findDocChild, _ := svc.repo.FindTaskChild(ctx, holdingCode, doc.GuidFixed)

		docList[i].TaskChild = findDocChild
	}

	return docList, total, nil
}

func (svc TaskHttpService) SaveInBatch(holdingCode string, authUsername string, dataList []models.Task) (common.BulkImport, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	payloadList, payloadDuplicateList := importdata.FilterDuplicate[models.Task](dataList, svc.getDocIDKey)

	itemCodeGuidList := []string{}
	for _, doc := range payloadList {
		itemCodeGuidList = append(itemCodeGuidList, doc.Name)
	}

	findItemGuid, err := svc.repo.FindInItemGuid(ctx, holdingCode, "name", itemCodeGuidList)

	if err != nil {
		return common.BulkImport{}, err
	}

	foundItemGuidList := []string{}
	for _, doc := range findItemGuid {
		foundItemGuidList = append(foundItemGuidList, doc.Name)
	}

	duplicateDataList, createDataList := importdata.PreparePayloadData[models.Task, models.TaskDoc](
		holdingCode,
		authUsername,
		foundItemGuidList,
		payloadList,
		svc.getDocIDKey,
		func(holdingCode string, authUsername string, doc models.Task) models.TaskDoc {
			newGuid := utils.NewGUID()

			dataDoc := models.TaskDoc{}

			dataDoc.GuidFixed = newGuid
			dataDoc.HoldingCode = holdingCode
			dataDoc.Task = doc

			currentTime := time.Now()
			dataDoc.CreatedBy = authUsername
			dataDoc.CreatedAt = currentTime
			return dataDoc
		},
	)

	updateSuccessDataList, updateFailDataList := importdata.UpdateOnDuplicate[models.Task, models.TaskDoc](
		holdingCode,
		authUsername,
		duplicateDataList,
		svc.getDocIDKey,
		func(holdingCode string, guid string) (models.TaskDoc, error) {
			return svc.repo.FindByDocIndentityGuid(ctx, holdingCode, "name", guid)
		},
		func(doc models.TaskDoc) bool {
			return doc.Name != ""
		},
		func(holdingCode string, authUsername string, data models.Task, doc models.TaskDoc) error {

			doc.Task = data
			doc.UpdatedBy = authUsername
			doc.UpdatedAt = time.Now()

			err = svc.repo.Update(ctx, holdingCode, doc.GuidFixed, doc)
			if err != nil {
				return nil
			}
			return nil
		},
	)

	if len(createDataList) > 0 {
		err = svc.repo.CreateInBatch(ctx, createDataList)

		if err != nil {
			return common.BulkImport{}, err
		}

	}

	createDataKey := []string{}

	for _, doc := range createDataList {
		createDataKey = append(createDataKey, doc.Name)
	}

	payloadDuplicateDataKey := []string{}
	for _, doc := range payloadDuplicateList {
		payloadDuplicateDataKey = append(payloadDuplicateDataKey, doc.Name)
	}

	updateDataKey := []string{}
	for _, doc := range updateSuccessDataList {

		updateDataKey = append(updateDataKey, doc.Name)
	}

	updateFailDataKey := []string{}
	for _, doc := range updateFailDataList {
		updateFailDataKey = append(updateFailDataKey, svc.getDocIDKey(doc))
	}

	return common.BulkImport{
		Created:          createDataKey,
		Updated:          updateDataKey,
		UpdateFailed:     updateFailDataKey,
		PayloadDuplicate: payloadDuplicateDataKey,
	}, nil
}

func (svc TaskHttpService) getDocIDKey(doc models.Task) string {
	return doc.Name
}

func (svc TaskHttpService) UpdateTaskTotalImage(docReq documentImageModel.DocumentImageTaskChangeMessage) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	totalDocStatus := []models.TotalStatus{}

	for _, doc := range docReq.CountStatus {
		totalDocStatus = append(totalDocStatus, models.TotalStatus{
			Status: doc.Status,
			Total:  doc.Count,
		})
	}

	err := svc.repo.UpdateTotalDocumentImageGroup(ctx, docReq.HoldingCode, docReq.TaskGUID, docReq.Count, totalDocStatus, docReq.BillCount, docReq.ReferenceCount, docReq.ReferenceBalance)

	if err != nil {
		return err
	}

	return nil
}

func (svc TaskHttpService) UpdateTaskTotalRejectImage(docReq documentImageModel.DocumentImageTaskRejectMessage) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// findDoc, err := svc.repo.FindByGuid(docReq.HoldingCode, docReq.TaskGUID)

	// if err != nil {
	// 	return err
	// }

	// if findDoc.ID == primitive.NilObjectID {
	// 	return errors.New("document not found")
	// }

	// currentTotal := findDoc.ToTalReject
	// if docReq.Event == documentImageModel.TaskRejectPlus {
	// 	findDoc.ToTalReject = currentTotal + docReq.Count
	// } else if docReq.Event == documentImageModel.TaskRejectMinus && currentTotal > 0 {
	// 	findDoc.ToTalReject = currentTotal - docReq.Count
	// }

	err := svc.repo.UpdateTotalRejectDocumentImageGroup(ctx, docReq.HoldingCode, docReq.TaskGUID, docReq.Count)

	if err != nil {
		return err
	}

	return nil
}
