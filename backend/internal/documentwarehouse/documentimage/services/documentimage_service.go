package services

import (
	"context"
	"errors"
	"fmt"
	"mime/multipart"
	"smlcloudplatform/internal/documentwarehouse/documentimage/models"
	"smlcloudplatform/internal/documentwarehouse/documentimage/repositories"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"strings"
	"time"

	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type IDocumentImageService interface {
	CreateDocumentImage(holdingCode string, authUsername string, doc models.DocumentImageRequest) (string, string, error)
	BulkCreateDocumentImage(holdingCode string, authUsername string, docs []models.DocumentImageRequest) error
	InfoDocumentImage(holdingCode string, guid string) (models.DocumentImageInfo, error)
	SearchDocumentImage(holdingCode string, matchFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.DocumentImageInfo, mongopagination.PaginationData, error)
	UploadDocumentImage(holdingCode string, authUsername string, fh *multipart.FileHeader) (*models.DocumentImageInfo, error)
	CreateImageEdit(holdingCode string, authUsername string, docImageGUID string, docRequest models.ImageEditRequest) error
	CreateImageComment(holdingCode string, authUsername string, docImageGUID string, docRequest models.CommentRequest) error

	CreateDocumentImageGroup(holdingCode string, authUsername string, docImageGroup models.DocumentImageGroup) (string, error)
	GetDocumentImageDocRefGroup(holdingCode string, docImageGroupGUID string) (models.DocumentImageGroupInfo, error)
	GetDocumentImageGroupByDocRef(holdingCode string, docRef string) (models.DocumentImageGroupInfo, error)
	UpdateDocumentImageGroup(holdingCode string, authUsername string, groupGUID string, docImageGroup models.DocumentImageGroup) error
	UpdateImageReferenceByDocumentImageGroup(holdingCode string, authUsername string, groupGUID string, docImages []models.ImageReferenceBody) error
	UpdateReferenceByDocumentImageGroup(holdingCode string, authUsername string, groupGUID string, docRef models.Reference) error
	UpdateTagsInDocumentImageGroup(holdingCode string, authUsername string, groupGUID string, tags []string) error
	UpdateStatusDocumentImageGroup(holdingCode string, authUsername string, groupGUID string, status int8) error
	UnGroupDocumentImageGroup(holdingCode string, authUsername string, groupGUID string) ([]string, error)
	ListDocumentImageGroup(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DocumentImageGroupInfo, mongopagination.PaginationData, error)
	DeleteReferenceByDocumentImageGroup(holdingCode string, authUsername string, groupGUID string, docRef models.Reference) error
	DeleteDocumentImageGroupByGuid(holdingCode string, authUsername string, DocumentImageGroupGuidFixed string) error
	DeleteDocumentImageGroupByGuids(holdingCode string, authUsername string, documentImageGroupGuidFixeds []string) error
	XSortsUpdate(ctx context.Context, holdingCode string, authUsername string, taskGUID string, xsorts []models.XSortDocumentImageGroupRequest) error

	ReCountStatusDocumentImageGroupByGUID(holdingCode string, authUsername string, documentImageGroupGUID string) error
	UpdateDocumentImageReferenceGroup() error
	UpdateStatusDocumentImageGroupByTask(holdingCode string, authUsername string, taskGUID string, status int8) error
	ReCountStatusDocumentImageGroupByTask(holdingCode string, authUsername string, taskGUID string) error
	UpdateDocNoInReferences(holdingCode string, module string, oldDocNo string, newDocNo string) error
}

type DocumentImageService struct {
	repoImageGroup               repositories.DocumentImageGroupRepository
	repoImage                    repositories.IDocumentImageRepository
	repoMessagequeue             repositories.DocumentImageMessageQueueRepository
	FilePersister                microservice.IPersisterFile
	maxImageReferences           int
	contextTimeout               time.Duration
	timeNowFnc                   func() time.Time
	newDocumentImageGUIDFnc      func() string
	newDocumentImageGroupGUIDFnc func() string
}

func NewDocumentImageService(repo repositories.IDocumentImageRepository, repoImageGroup repositories.DocumentImageGroupRepository, repoMessagequeue repositories.DocumentImageMessageQueueRepository, filePersister microservice.IPersisterFile) DocumentImageService {

	contextTimeout := time.Duration(15) * time.Second

	return DocumentImageService{
		maxImageReferences: 100,
		repoImageGroup:     repoImageGroup,
		repoImage:          repo,
		repoMessagequeue:   repoMessagequeue,
		FilePersister:      filePersister,
		contextTimeout:     contextTimeout,
		timeNowFnc: func() time.Time {
			return time.Now()
		},
		newDocumentImageGUIDFnc:      utils.NewGUID,
		newDocumentImageGroupGUIDFnc: utils.NewGUID,
	}
}

func (svc DocumentImageService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc DocumentImageService) CreateDocumentImage(holdingCode string, authUsername string, docRequest models.DocumentImageRequest) (string, string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDocImgGroup := models.DocumentImageGroupDoc{}
	if len(docRequest.DocumentImageGroupGUID) > 0 {
		_, err := svc.repoImageGroup.FindByGuid(ctx, holdingCode, docRequest.DocumentImageGroupGUID)

		if err != nil {
			return "", "", err
		}
	}

	// do upload first

	createdAt := svc.timeNowFnc()

	documentImageGUID := svc.newDocumentImageGUIDFnc()

	docData := models.DocumentImageDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = documentImageGUID
	docData.DocumentImage = docRequest.DocumentImage

	docData.Edits = []models.ImageEdit{}
	docData.References = []models.Reference{}
	docData.MetaFileAt = docRequest.MetaFileAt

	docData.CreatedBy = authUsername
	docData.CreatedAt = createdAt

	docData.UploadedBy = authUsername
	docData.UploadedAt = createdAt

	tags := []string{}
	if docRequest.Tags != nil {
		tags = *docRequest.Tags
	}

	// image group
	imageGroupGUID := svc.newDocumentImageGroupGUIDFnc()

	if len(findDocImgGroup.GuidFixed) > 0 {
		imageGroupGUID = findDocImgGroup.GuidFixed
	}

	docImageRef := svc.documentImageToImageReference(documentImageGUID, docRequest.DocumentImage, authUsername, createdAt)
	docDataImageGroup := svc.createImageGroupByDocumentImage(holdingCode, authUsername, imageGroupGUID, docImageRef, docRequest.ImageURI, tags, docRequest.TaskGUID, docRequest.PathTask, createdAt, 0)

	newXOrderDocImgGroup, _ := svc.newXOrderDocumentImageGroup(ctx, holdingCode, docRequest.TaskGUID)
	docDataImageGroup.XOrder = newXOrderDocImgGroup

	err := svc.repoImageGroup.Transaction(ctx, func(ctx context.Context) error {

		_, err := svc.repoImage.Create(ctx, docData)

		if err != nil {
			return err
		}

		_, err = svc.repoImageGroup.Create(ctx, docDataImageGroup)

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", "", err
	}

	_, err = svc.messageQueueReCountDocumentImageGroup(ctx, holdingCode, docRequest.TaskGUID)
	if err != nil {
		return "", "", err
	}

	return documentImageGUID, imageGroupGUID, nil
}

func (svc DocumentImageService) CreateDocumentImageWithTask(holdingCode string, authUsername string, docRequest models.DocumentImageRequest) (string, string, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// do upload first

	createdAt := svc.timeNowFnc()

	documentImageGUID := svc.newDocumentImageGUIDFnc()

	docData := models.DocumentImageDoc{}
	docData.HoldingCode = holdingCode
	docData.GuidFixed = documentImageGUID
	docData.DocumentImage = docRequest.DocumentImage

	docData.Edits = []models.ImageEdit{}
	docData.References = []models.Reference{}
	docData.MetaFileAt = docRequest.MetaFileAt

	docData.CreatedBy = authUsername
	docData.CreatedAt = createdAt

	docData.UploadedBy = authUsername
	docData.UploadedAt = createdAt

	tags := []string{}
	if docRequest.Tags != nil {
		tags = *docRequest.Tags
	}

	// image group
	imageGroupGUID := svc.newDocumentImageGroupGUIDFnc()
	docImageRef := svc.documentImageToImageReference(documentImageGUID, docRequest.DocumentImage, authUsername, createdAt)
	docDataImageGroup := svc.createImageGroupByDocumentImage(holdingCode, authUsername, imageGroupGUID, docImageRef, docRequest.ImageURI, tags, docRequest.TaskGUID, docRequest.PathTask, createdAt, 0)

	newXOrderDocImgGroup, _ := svc.newXOrderDocumentImageGroup(ctx, holdingCode, docRequest.TaskGUID)

	docDataImageGroup.XOrder = newXOrderDocImgGroup

	err := svc.repoImageGroup.Transaction(ctx, func(ctx context.Context) error {

		_, err := svc.repoImage.Create(ctx, docData)

		if err != nil {
			return err
		}

		_, err = svc.repoImageGroup.Create(ctx, docDataImageGroup)

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return "", "", err
	}

	_, err = svc.messageQueueReCountDocumentImageGroup(ctx, holdingCode, docRequest.TaskGUID)
	if err != nil {
		return "", "", err
	}

	return documentImageGUID, imageGroupGUID, nil
}

func (svc DocumentImageService) CreateImageEdit(holdingCode string, authUsername string, docImageGUID string, docRequest models.ImageEditRequest) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repoImage.FindByGuid(ctx, holdingCode, docImageGUID)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) == 0 {
		return errors.New("document image not found")
	}

	tempDoc := findDoc

	if tempDoc.Edits == nil {
		tempDoc.Edits = []models.ImageEdit{}
	}

	imageEdit := models.ImageEdit{}

	imageEdit.ImageURI = docRequest.ImageURI
	imageEdit.EditedBy = authUsername
	imageEdit.EditedAt = svc.timeNowFnc()

	tempDoc.Edits = append(tempDoc.Edits, imageEdit)

	err = svc.repoImage.Update(ctx, holdingCode, docImageGUID, tempDoc)
	if err != nil {
		return err
	}

	return nil
}

func (svc DocumentImageService) CreateImageComment(holdingCode string, authUsername string, docImageGUID string, docRequest models.CommentRequest) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repoImage.FindByGuid(ctx, holdingCode, docImageGUID)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) == 0 {
		return errors.New("document image not found")
	}

	tempDoc := findDoc

	if tempDoc.Comments == nil {
		tempDoc.Comments = []models.Comment{}
	}

	comment := models.Comment{}

	comment.Comment = docRequest.Comment
	comment.CommentedBy = authUsername
	comment.CommentedAt = svc.timeNowFnc()

	tempDoc.Comments = append(tempDoc.Comments, comment)

	err = svc.repoImage.Update(ctx, holdingCode, docImageGUID, tempDoc)
	if err != nil {
		return err
	}

	return nil
}

func (svc DocumentImageService) BulkCreateDocumentImage(holdingCode string, authUsername string, docs []models.DocumentImageRequest) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// do upload first

	createdAt := svc.timeNowFnc()
	docDataList := []models.DocumentImageDoc{}
	docDataImageGroupList := []models.DocumentImageGroupDoc{}

	taskLastXOrder := map[string]int{}

	for _, doc := range docs {
		if len(doc.TaskGUID) < 1 {
			return fmt.Errorf("job is empty")
		}

		_, ok := taskLastXOrder[doc.TaskGUID]
		if !ok {
			newXOrderDocImgGroup, _ := svc.newXOrderDocumentImageGroup(ctx, holdingCode, doc.TaskGUID)
			taskLastXOrder[doc.TaskGUID] = newXOrderDocImgGroup
		} else {
			taskLastXOrder[doc.TaskGUID]++
		}

		documentImageGUID := svc.newDocumentImageGUIDFnc()

		docData := models.DocumentImageDoc{}
		docData.HoldingCode = holdingCode
		docData.GuidFixed = documentImageGUID
		docData.DocumentImage = doc.DocumentImage

		docData.Edits = []models.ImageEdit{}
		docData.References = []models.Reference{}
		docData.MetaFileAt = doc.MetaFileAt

		docData.CreatedBy = authUsername
		docData.CreatedAt = createdAt

		// docData.UpdatedBy = authUsername
		// docData.UpdatedAt = createdAt

		docData.UploadedBy = authUsername
		// docData.UploadedAt = createdAt

		// image group
		imageGroupGUID := svc.newDocumentImageGroupGUIDFnc()
		docImageRef := svc.documentImageToImageReference(documentImageGUID, doc.DocumentImage, authUsername, createdAt)

		tags := []string{}
		if doc.Tags != nil {
			tags = *doc.Tags
		}

		docDataImageGroup := svc.createImageGroupByDocumentImage(holdingCode, authUsername, imageGroupGUID, docImageRef, doc.ImageURI, tags, doc.TaskGUID, doc.PathTask, createdAt, 0)

		docDataImageGroup.XOrder = taskLastXOrder[doc.TaskGUID]

		docDataList = append(docDataList, docData)
		docDataImageGroupList = append(docDataImageGroupList, docDataImageGroup)
	}

	err := svc.repoImageGroup.Transaction(ctx, func(ctx context.Context) error {

		err := svc.repoImage.CreateInBatch(ctx, docDataList)

		if err != nil {
			return err
		}

		err = svc.repoImageGroup.CreateInBatch(ctx, docDataImageGroupList)

		if err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	taskGUIDsChanged := map[string]struct{}{}
	for _, tempDocGroup := range docDataImageGroupList {
		taskGUIDsChanged[tempDocGroup.TaskGUID] = struct{}{}
	}

	for taskGUID := range taskGUIDsChanged {
		_, err = svc.messageQueueReCountDocumentImageGroup(ctx, holdingCode, taskGUID)
		if err != nil {
			fmt.Println(err)
		}
	}

	return nil
}

func (svc DocumentImageService) InfoDocumentImage(holdingCode string, guid string) (models.DocumentImageInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repoImage.FindByGuid(ctx, holdingCode, guid)

	if err != nil {
		return models.DocumentImageInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.DocumentImageInfo{}, errors.New("document not found")
	}

	return findDoc.DocumentImageInfo, nil
}

func (svc DocumentImageService) SearchDocumentImage(holdingCode string, matchFilters map[string]interface{}, pageable micromodels.Pageable) ([]models.DocumentImageInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{"guid_fixed", "documentref", "module"}
	docList, pagination, err := svc.repoImage.FindPageFilter(ctx, holdingCode, matchFilters, searchInFields, pageable)

	if err != nil {
		return []models.DocumentImageInfo{}, pagination, err
	}

	return docList, pagination, nil
}

func (svc DocumentImageService) UploadDocumentImage(holdingCode string, authUsername string, fh *multipart.FileHeader) (*models.DocumentImageInfo, error) {

	if fh.Filename == "" {
		return nil, errors.New("image file name not found")
	}
	// try upload
	fileUploadMetadataSlice := strings.Split(fh.Filename, ".")
	fileName := svc.newDocumentImageGUIDFnc() //fileUploadMetadataSlice[0]
	fileExtension := fileUploadMetadataSlice[1]

	fileNameWithShop := fmt.Sprintf("%s/%s", holdingCode, fileName)

	imageUri, err := svc.FilePersister.Save(fh, fileNameWithShop, fileExtension)
	if err != nil {
		return nil, err
	}

	// create document image
	doc := new(models.DocumentImageDoc)
	doc.GuidFixed = svc.newDocumentImageGUIDFnc()
	doc.ImageURI = imageUri
	doc.HoldingCode = holdingCode
	doc.UploadedBy = authUsername
	doc.UploadedAt = svc.timeNowFnc()
	doc.CreatedBy = authUsername
	doc.CreatedAt = doc.UploadedAt

	docRequest := models.DocumentImageRequest{
		DocumentImage: doc.DocumentImage,
		Tags:          &[]string{},
		TaskGUID:      "",
		PathTask:      "",
	}
	_, _, err = svc.CreateDocumentImage(holdingCode, authUsername, docRequest)

	if err != nil {
		return nil, err
	}

	return &doc.DocumentImageInfo, err
}

func (svc DocumentImageService) UpdateDocumentImageReferenceGroup() error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDocList, err := svc.repoImage.FindAll(ctx)

	if err != nil {
		return err
	}

	for _, findDoc := range findDocList {
		findGroupDoc, err := svc.repoImageGroup.FindOneByDocumentImageGUIDAll(ctx, findDoc.GuidFixed)

		if err != nil {
			return err
		}

		refGroups := []models.ReferenceGroup{}
		if findGroupDoc.ID == primitive.NilObjectID {
			refGroup := models.ReferenceGroup{}
			refGroup.GroupType = ""
			refGroup.ParentGUID = ""
			refGroup.XOrder = 1
			refGroup.XType = 0

			refGroups = append(refGroups, refGroup)

		}

		findDoc.ReferenceGroups = refGroups

		err = svc.repoImage.UpdateAll(ctx, findDoc)

		if err != nil {
			return err
		}
	}

	return nil
}
