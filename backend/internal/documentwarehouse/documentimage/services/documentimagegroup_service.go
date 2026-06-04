package services

import (
	"context"
	"errors"
	"fmt"
	"smlcloudplatform/internal/documentwarehouse/documentimage/models"
	micromodels "smlcloudplatform/pkg/microservice/models"
	"sort"
	"time"

	"github.com/samber/lo"
	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Group

func (svc DocumentImageService) getDocumentImageNotReferencedInGroup(holdingCode string, currentGroupGUID string, docImageRefs []models.ImageReferenceBody) ([]models.ImageReference, []string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	docImageGUIDs := []string{}

	for _, imageRef := range docImageRefs {
		docImageGUIDs = append(docImageGUIDs, imageRef.DocumentImageGUID)
	}

	passDocImagesRef := []models.ImageReference{}

	findGroups, err := svc.repoImageGroup.FindWithoutGUIDByDocumentImageGUIDs(ctx, holdingCode, currentGroupGUID, docImageGUIDs)

	if err != nil {
		return []models.ImageReference{}, []string{}, err
	}

	for _, imageGroup := range findGroups {
		for _, imageRef := range *imageGroup.ImageReferences {
			foundImageRef, isFound := lo.Find[models.ImageReferenceBody](docImageRefs, func(tempImageRef models.ImageReferenceBody) bool {
				return imageRef.DocumentImageGUID == tempImageRef.DocumentImageGUID
			})

			if isFound && (imageGroup.References == nil || len(imageGroup.References) < 1) {
				imageRef.XOrder = foundImageRef.XOrder
				passDocImagesRef = append(passDocImagesRef, imageRef)
			} else if imageGroup.References != nil && len(imageGroup.References) > 0 {
				return []models.ImageReference{}, []string{}, fmt.Errorf("document image guid %s has referenced in %s", imageRef.DocumentImageGUID, imageGroup.GuidFixed)
			} else {
				return []models.ImageReference{}, []string{}, fmt.Errorf("document image guid \"%s\" has referenced in %s", imageRef.DocumentImageGUID, imageGroup.GuidFixed)
			}
		}
	}

	tempDocumentImageGroupGUIDs := []string{}
	for _, imageGroup := range findGroups {
		tempDocumentImageGroupGUIDs = append(tempDocumentImageGroupGUIDs, imageGroup.GuidFixed)
	}

	return passDocImagesRef, tempDocumentImageGroupGUIDs, nil
}

func (svc DocumentImageService) CreateDocumentImageGroup(holdingCode string, authUsername string, docImageGroup models.DocumentImageGroup) (string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	if docImageGroup.ImageReferences == nil || len(*docImageGroup.ImageReferences) < 1 {
		return "", errors.New("document image is size 0")
	}

	if docImageGroup.ImageReferences == nil || len(*docImageGroup.ImageReferences) > svc.maxImageReferences {
		return "", fmt.Errorf("document images is over size %d", svc.maxImageReferences)
	}

	docImageGroupData := models.DocumentImageGroupDoc{}

	tempImageRefs := lo.Map[models.ImageReference, models.ImageReferenceBody](
		*docImageGroup.ImageReferences,
		func(temp models.ImageReference, index int) models.ImageReferenceBody {
			return temp.ImageReferenceBody
		})

	passDocImagesRef, docImageGroupGUIDs, err := svc.getDocumentImageNotReferencedInGroup(holdingCode, "", tempImageRefs)
	if err != nil {
		return "", err
	}

	if len(passDocImagesRef) < 1 {
		return "", fmt.Errorf("document images invalid")
	}

	createdAt := svc.timeNowFnc()
	docImageGroupGUIDFixed := svc.newDocumentImageGroupGUIDFnc()

	docImageGroupData.HoldingCode = holdingCode
	docImageGroupData.DocumentImageGroup = docImageGroup
	docImageGroupData.GuidFixed = docImageGroupGUIDFixed
	docImageGroupData.Status = models.IMAGE_PENDING

	docImageGroupData.References = []models.Reference{}

	newXOrder, _ := svc.newXOrderDocumentImageGroup(ctx, holdingCode, docImageGroup.TaskGUID)
	docImageGroupData.XOrder = newXOrder

	docImageGroupData.CreatedBy = authUsername
	docImageGroupData.CreatedAt = createdAt

	sort.Slice(passDocImagesRef, func(i, j int) bool {
		return passDocImagesRef[i].XOrder < passDocImagesRef[j].XOrder
	})

	// if len(passDocImagesRef) > 0 {
	// 	docImageGroupData.UploadedBy = passDocImagesRef[0].UploadedBy
	// 	docImageGroupData.UploadedAt = passDocImagesRef[0].UploadedAt
	// }

	docImageGroupData.UploadedBy = authUsername
	docImageGroupData.UploadedAt = docImageGroup.UploadedAt

	docImageGroupData.ImageReferences = &passDocImagesRef

	tempGUIDDocumentImages := []string{}
	for _, imageRef := range passDocImagesRef {
		tempGUIDDocumentImages = append(tempGUIDDocumentImages, imageRef.DocumentImageGUID)

	}

	err = svc.clearCreateDocumentImageGroupByDocumentGUIDs(holdingCode, docImageGroupGUIDs, tempGUIDDocumentImages)
	if err != nil {
		return "", err
	}

	_, err = svc.repoImageGroup.Create(ctx, docImageGroupData)

	if err != nil {
		return "", err
	}

	return docImageGroupGUIDFixed, nil
}

func (svc DocumentImageService) UpdateStatusDocumentImageGroup(holdingCode string, authUsername string, groupGUID string, status int8) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repoImageGroup.FindByGuid(ctx, holdingCode, groupGUID)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	if findDoc.Status == status {
		return nil
	}

	if status < models.IMAGE_PENDING || status > models.IMAGE_GL_COMPLETED {
		return errors.New("status out of range")
	}

	updateDoc := findDoc

	lastStatusHistory := models.StatusHistory{
		Status:    findDoc.Status,
		ChangedBy: authUsername,
		ChangedAt: svc.timeNowFnc(),
	}

	updateDoc.StatusHistories = append(updateDoc.StatusHistories, lastStatusHistory)
	updateDoc.Status = status
	svc.repoImageGroup.Update(ctx, holdingCode, groupGUID, updateDoc)

	_, err = svc.messageQueueReCountDocumentImageGroup(ctx, holdingCode, findDoc.TaskGUID)
	if err != nil {
		fmt.Println(err.Error())
	}

	return nil
}

// ReCountStatusDocumentImageGroupByGUID recount status by document image group guid without update
func (svc DocumentImageService) ReCountStatusDocumentImageGroupByGUID(holdingCode string, authUsername string, groupGUID string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repoImageGroup.FindByGuid(ctx, holdingCode, groupGUID)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	_, err = svc.messageQueueReCountDocumentImageGroup(ctx, holdingCode, findDoc.TaskGUID)
	if err != nil {
		return err
	}

	return nil
}

func (svc DocumentImageService) UpdateStatusDocumentImageGroupByTask(holdingCode string, authUsername string, taskGUID string, status int8) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	if status != models.IMAGE_PENDING && status != models.IMAGE_CHECKED {
		return errors.New("status out of range")
	}

	err := svc.repoImageGroup.UpdateStatusByTask(ctx, holdingCode, taskGUID, status)

	if err != nil {
		return err
	}

	_, err = svc.messageQueueReCountDocumentImageGroup(ctx, holdingCode, taskGUID)
	if err != nil {
		fmt.Println(err.Error())
	}

	return nil
}

func (svc DocumentImageService) ReCountStatusDocumentImageGroupByTask(holdingCode string, authUsername string, taskGUID string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	_, err := svc.messageQueueReCountDocumentImageGroup(ctx, holdingCode, taskGUID)
	if err != nil {
		fmt.Println(err.Error())
	}

	return nil
}

func (svc DocumentImageService) UpdateDocumentImageGroup(holdingCode string, authUsername string, groupGUID string, docImageGroup models.DocumentImageGroup) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	if docImageGroup.ImageReferences == nil || len(*docImageGroup.ImageReferences) > svc.maxImageReferences {
		return fmt.Errorf("document image is over size %d", svc.maxImageReferences)
	}

	findDoc, err := svc.repoImageGroup.FindByGuid(ctx, holdingCode, groupGUID)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	docImages := *docImageGroup.ImageReferences

	tempImageRefs := lo.Map[models.ImageReference, models.ImageReferenceBody](*docImageGroup.ImageReferences, func(temp models.ImageReference, index int) models.ImageReferenceBody {
		return temp.ImageReferenceBody
	})

	passDocImagesRef, docImageGroupGUIDs, err := svc.getDocumentImageNotReferencedInGroup(holdingCode, groupGUID, tempImageRefs)
	if err != nil {
		return err
	}

	updateDocImagesRef := map[string]models.ImageReference{}
	for _, imageRef := range passDocImagesRef {
		updateDocImagesRef[imageRef.DocumentImageGUID] = imageRef
	}

	tempRemoveDocImageFromGroup := []models.ImageReference{}
	// check exist from current document image group
	if findDoc.ImageReferences != nil {
		for _, imageRef := range *findDoc.ImageReferences {
			tempImageRef, isFound := lo.Find[models.ImageReference](docImages, func(tempImageRef models.ImageReference) bool {
				return imageRef.DocumentImageGUID == tempImageRef.DocumentImageGUID
			})

			if isFound {
				updateDocImagesRef[tempImageRef.DocumentImageGUID] = tempImageRef
			} else {
				tempRemoveDocImageFromGroup = append(tempRemoveDocImageFromGroup, imageRef)
			}
		}
	}

	tempDocImageGUIDs := []string{}
	tempDocImageRef := map[string]models.ImageReference{}
	for _, docImageRef := range updateDocImagesRef {
		tempDocImageGUIDs = append(tempDocImageGUIDs, docImageRef.DocumentImageGUID)
		tempDocImageRef[docImageRef.DocumentImageGUID] = docImageRef
	}

	findDocImages, err := svc.repoImage.FindInGUIDs(ctx, holdingCode, tempDocImageGUIDs)
	if err != nil {
		return err
	}

	docImgRefs := []models.ImageReference{}
	for _, docRefImage := range findDocImages {

		tempDocImgRef := models.ImageReference{}

		tempDocImgRef.DocumentImageGUID = docRefImage.GuidFixed
		tempDocImgRef.ImageURI = docRefImage.ImageURI
		tempDocImgRef.CloneImageFrom = docRefImage.GuidFixed
		tempDocImgRef.Name = docRefImage.Name
		tempDocImgRef.MetaFileAt = docRefImage.MetaFileAt
		tempDocImgRef.UploadedAt = docRefImage.UploadedAt
		tempDocImgRef.UploadedBy = docRefImage.UploadedBy

		if temp, ok := tempDocImageRef[docRefImage.GuidFixed]; ok {
			tempDocImgRef.XOrder = temp.XOrder
		}

		docImgRefs = append(docImgRefs, tempDocImgRef)
	}

	sort.Slice(docImgRefs, func(i, j int) bool {
		return docImgRefs[i].XOrder < docImgRefs[j].XOrder
	})

	if len(tempDocImageRef) > 0 {
		findDoc.UploadedBy = docImgRefs[0].UploadedBy
		findDoc.UploadedAt = docImgRefs[0].UploadedAt
	}

	updateDoc := findDoc
	timeAt := svc.timeNowFnc()

	updateDoc.DocumentImageGroup = docImageGroup
	updateDoc.ImageReferences = &docImgRefs
	updateDoc.References = findDoc.References

	updateDoc.UpdatedAt = timeAt
	updateDoc.UpdatedBy = authUsername

	updateDoc.Status = findDoc.Status
	updateDoc.StatusChangedBy = findDoc.StatusChangedBy
	updateDoc.StatusChangedAt = findDoc.StatusChangedAt
	updateDoc.StatusHistories = findDoc.StatusHistories

	if err = svc.repoImageGroup.Update(ctx, holdingCode, groupGUID, updateDoc); err != nil {
		return err
	}

	if err = svc.clearUpdateDocumentImageGroupByDocumentGUIDs(holdingCode, groupGUID, docImageGroupGUIDs, tempDocImageGUIDs); err != nil {
		return err
	}

	for _, imageRef := range tempRemoveDocImageFromGroup {
		imageGroupGUID := svc.newDocumentImageGroupGUIDFnc()
		tempTags := []string{}
		if findDoc.Tags != nil {
			tempTags = *findDoc.Tags
		}

		docImageGroup := svc.createImageGroupByDocumentImage(holdingCode, authUsername, imageGroupGUID, imageRef, imageRef.ImageURI, tempTags, findDoc.TaskGUID, findDoc.PathTask, timeAt, docImageGroup.BillCount)

		newXOrderDocImgGroup, _ := svc.newXOrderDocumentImageGroup(ctx, holdingCode, findDoc.TaskGUID)

		docImageGroup.XOrder = newXOrderDocImgGroup

		_, err = svc.repoImageGroup.Create(ctx, docImageGroup)

		if err != nil {
			return err
		}
	}

	return nil
}

func (svc DocumentImageService) UpdateImageReferenceByDocumentImageGroup(holdingCode string, authUsername string, groupGUID string, docImages []models.ImageReferenceBody) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repoImageGroup.FindByGuid(ctx, holdingCode, groupGUID)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	tempDocImageGUIDs := []string{}

	for _, imageRef := range docImages {
		tempDocImageGUIDs = append(tempDocImageGUIDs, imageRef.DocumentImageGUID)
	}

	findDocImages, err := svc.repoImage.FindInGUIDs(ctx, holdingCode, tempDocImageGUIDs)

	if err != nil {
		return err
	}

	maxSizeInvalid := 50
	if len(docImages) > (len(findDocImages) + maxSizeInvalid) {
		return errors.New("document image invalid")
	}

	passDocImagesRef, docImageGroupGUIDs, err := svc.getDocumentImageNotReferencedInGroup(holdingCode, groupGUID, docImages)
	if err != nil {
		return err
	}

	updateDocImagesRef := map[string]models.ImageReference{}
	for _, imageRef := range passDocImagesRef {
		updateDocImagesRef[imageRef.DocumentImageGUID] = imageRef
	}

	tempRemoveDocImageFromGroup := []models.ImageReference{}
	// check exist from current document image group
	if findDoc.ImageReferences != nil {
		for _, imageRef := range *findDoc.ImageReferences {
			tempImageRef, isFound := lo.Find[models.ImageReferenceBody](docImages, func(tempImageRef models.ImageReferenceBody) bool {
				return imageRef.DocumentImageGUID == tempImageRef.DocumentImageGUID
			})

			if isFound {
				imageRef.XOrder = tempImageRef.XOrder
				updateDocImagesRef[tempImageRef.DocumentImageGUID] = imageRef
			} else {
				tempRemoveDocImageFromGroup = append(tempRemoveDocImageFromGroup, imageRef)
			}
		}
	}

	tempGUIDDocumentImages := []string{}
	tempDocImageRef := []models.ImageReference{}
	for _, docImageRef := range updateDocImagesRef {
		tempGUIDDocumentImages = append(tempGUIDDocumentImages, docImageRef.DocumentImageGUID)
		tempDocImageRef = append(tempDocImageRef, docImageRef)

	}

	sort.Slice(tempDocImageRef, func(i, j int) bool {
		return tempDocImageRef[i].XOrder < tempDocImageRef[j].XOrder
	})

	if len(tempDocImageRef) > 0 {
		findDoc.UploadedBy = tempDocImageRef[0].UploadedBy
		findDoc.UploadedAt = tempDocImageRef[0].UploadedAt
	}

	timeAt := svc.timeNowFnc()

	findDoc.ImageReferences = &tempDocImageRef

	findDoc.UpdatedAt = timeAt
	findDoc.UpdatedBy = authUsername

	if err = svc.repoImageGroup.Update(ctx, holdingCode, groupGUID, findDoc); err != nil {
		return err
	}

	if err = svc.clearUpdateDocumentImageGroupByDocumentGUIDs(holdingCode, groupGUID, docImageGroupGUIDs, tempGUIDDocumentImages); err != nil {
		return err
	}

	for _, imageRef := range tempRemoveDocImageFromGroup {
		imageGroupGUID := svc.newDocumentImageGroupGUIDFnc()

		tempTags := []string{}
		if findDoc.Tags != nil {
			tempTags = *findDoc.Tags
		}

		docImageGroup := svc.createImageGroupByDocumentImage(holdingCode, authUsername, imageGroupGUID, imageRef, imageRef.ImageURI, tempTags, findDoc.TaskGUID, findDoc.PathTask, timeAt, 0)
		newXOrderDocImgGroup, _ := svc.newXOrderDocumentImageGroup(ctx, holdingCode, findDoc.TaskGUID)

		docImageGroup.XOrder = newXOrderDocImgGroup

		_, err = svc.repoImageGroup.Create(ctx, docImageGroup)

		if err != nil {
			return err
		}
	}

	return nil
}

func (svc DocumentImageService) UpdateReferenceByDocumentImageGroup(holdingCode string, authUsername string, groupGUID string, docRef models.Reference) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repoImageGroup.FindByGuid(ctx, holdingCode, groupGUID)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document image group not found")
	}

	// Check if exact same Module+DocNo already exists (prevent exact duplicate only)
	if findDoc.References != nil {
		_, isExactDuplicate := lo.Find[models.Reference](findDoc.References, func(tempDoc models.Reference) bool {
			return tempDoc.Module == docRef.Module && tempDoc.DocNo == docRef.DocNo
		})

		if isExactDuplicate {
			return errors.New("document has the same reference (Module + DocNo) already")
		}
	}

	// Clear references
	docImageClearRefs, err := svc.repoImage.FindByReference(ctx, holdingCode, docRef)
	if err != nil {
		return err
	}

	docImageGroupClearRefs, err := svc.repoImageGroup.FindByReference(ctx, holdingCode, docRef)
	if err != nil {
		return err
	}

	for _, docImage := range docImageClearRefs {
		tempDocRefs := lo.Filter[models.Reference](docImage.References, func(tempDocRef models.Reference, idx int) bool {
			return docRef.DocNo != tempDocRef.DocNo
		})

		docImage.References = tempDocRefs

		svc.repoImage.Update(ctx, holdingCode, docImage.GuidFixed, docImage)
	}

	for _, docImageGroup := range docImageGroupClearRefs {
		tempDocRefs := lo.Filter[models.Reference](docImageGroup.References, func(tempDocRef models.Reference, idx int) bool {
			return docRef.DocNo != tempDocRef.DocNo
		})

		docImageGroup.References = tempDocRefs

		svc.repoImageGroup.Update(ctx, holdingCode, docImageGroup.GuidFixed, docImageGroup)
	}

	tempDocImageGUIDs := []string{}

	for _, imageRef := range *findDoc.ImageReferences {
		tempDocImageGUIDs = append(tempDocImageGUIDs, imageRef.DocumentImageGUID)
	}

	findDocImages, err := svc.repoImage.FindInGUIDs(ctx, holdingCode, tempDocImageGUIDs)

	if err != nil {
		return err
	}

	findDoc.References = append(findDoc.References, docRef)

	timeAt := svc.timeNowFnc()

	findDoc.UpdatedAt = timeAt
	findDoc.UpdatedBy = authUsername

	if err = svc.repoImageGroup.Update(ctx, holdingCode, groupGUID, findDoc); err != nil {
		return err
	}

	for _, docImage := range findDocImages {

		docImage.UpdatedAt = timeAt
		docImage.UpdatedBy = authUsername

		docImage.References = append(docImage.References, docRef)

		if err = svc.repoImage.Update(ctx, holdingCode, docImage.GuidFixed, docImage); err != nil {
			return err
		}
	}

	return nil
}

func (svc DocumentImageService) UpdateTagsInDocumentImageGroup(holdingCode string, authUsername string, groupGUID string, tags []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repoImageGroup.FindByGuid(ctx, holdingCode, groupGUID)

	if err != nil {
		return err
	}

	if len(findDoc.GuidFixed) < 1 {
		return errors.New("document not found")
	}

	findDoc.Tags = &tags

	findDoc.UpdatedAt = svc.timeNowFnc()
	findDoc.UpdatedBy = authUsername

	if err = svc.repoImageGroup.Update(ctx, holdingCode, groupGUID, findDoc); err != nil {
		return err
	}

	return nil
}

func (svc DocumentImageService) DeleteReferenceByDocumentImageGroup(holdingCode string, authUsername string, groupGUID string, docRef models.Reference) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repoImageGroup.FindByGuid(ctx, holdingCode, groupGUID)

	if err != nil {
		return err
	}

	if findDoc.ID == primitive.NilObjectID {
		return errors.New("document image group not found")
	}

	// Clear references
	docImageClearRefs, err := svc.repoImage.FindByReference(ctx, holdingCode, docRef)
	if err != nil {
		return err
	}

	docImageGroupClearRefs, err := svc.repoImageGroup.FindByReference(ctx, holdingCode, docRef)
	if err != nil {
		return err
	}

	for _, docImage := range docImageClearRefs {
		tempDocRefs := lo.Filter[models.Reference](docImage.References, func(tempDocRef models.Reference, idx int) bool {
			return docRef.Module != tempDocRef.Module && docRef.DocNo != tempDocRef.DocNo
		})

		docImage.References = tempDocRefs

		svc.repoImage.Update(ctx, holdingCode, docImage.GuidFixed, docImage)
	}

	for _, docImageGroup := range docImageGroupClearRefs {
		tempDocRefs := lo.Filter[models.Reference](docImageGroup.References, func(tempDocRef models.Reference, idx int) bool {
			return docRef.Module != tempDocRef.Module && docRef.DocNo != tempDocRef.DocNo
		})

		docImageGroup.References = tempDocRefs

		svc.repoImageGroup.Update(ctx, holdingCode, docImageGroup.GuidFixed, docImageGroup)
	}

	_, err = svc.messageQueueReCountDocumentImageGroup(ctx, holdingCode, findDoc.TaskGUID)
	if err != nil {
		fmt.Println(err.Error())
	}

	return nil
}

func (svc DocumentImageService) DeleteDocumentImageGroupByGuid(holdingCode string, authUsername string, documentImageGroupGuidFixed string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDocGroup, err := svc.repoImageGroup.FindByGuid(ctx, holdingCode, documentImageGroupGuidFixed)

	if err != nil {
		return err
	}

	if len(findDocGroup.GuidFixed) < 1 {
		return nil
	}

	// Warning: Document has references but will be deleted anyway
	if svc.isDocumentImageGroupHasReferenced(findDocGroup) {
		fmt.Printf("Warning: Deleting document group %s which has %d reference(s)\n", documentImageGroupGuidFixed, len(findDocGroup.References))
	}

	err = svc.repoImageGroup.Transaction(ctx, func(ctx context.Context) error {

		for _, docImage := range *findDocGroup.ImageReferences {
			err = svc.repoImage.DeleteByGuidfixed(ctx, holdingCode, docImage.DocumentImageGUID, authUsername)

			if err != nil {
				return err
			}
		}

		if err = svc.repoImageGroup.DeleteByGuidfixed(ctx, holdingCode, findDocGroup.GuidFixed); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return err
	}

	_, err = svc.messageQueueReCountDocumentImageGroup(ctx, holdingCode, findDocGroup.TaskGUID)
	if err != nil {
		fmt.Println(err.Error())
	}

	return nil
}

func (svc DocumentImageService) DeleteDocumentImageGroupByGuids(holdingCode string, authUsername string, documentImageGroupGuidFixeds []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.repoImageGroup.Transaction(ctx, func(ctx context.Context) error {
		for _, DocumentImageGroupGuidFixed := range documentImageGroupGuidFixeds {
			findDocGroup, err := svc.repoImageGroup.FindByGuid(ctx, holdingCode, DocumentImageGroupGuidFixed)

			if err != nil {
				return err
			}

			// Warning: Document has references but will be deleted anyway
			if svc.isDocumentImageGroupHasReferenced(findDocGroup) {
				fmt.Printf("Warning: Deleting document group %s which has %d reference(s)\n", DocumentImageGroupGuidFixed, len(findDocGroup.References))
			}

			for _, docImage := range *findDocGroup.ImageReferences {
				err = svc.repoImage.DeleteByGuidfixed(ctx, holdingCode, docImage.DocumentImageGUID, authUsername)

				if err != nil {
					return err
				}
			}

			if err = svc.repoImageGroup.DeleteByGuidfixed(ctx, holdingCode, findDocGroup.GuidFixed); err != nil {
				return err
			}

			_, err = svc.messageQueueReCountDocumentImageGroup(ctx, holdingCode, findDocGroup.TaskGUID)
			if err != nil {
				fmt.Println(err.Error())
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	return nil
}

func (svc DocumentImageService) UnGroupDocumentImageGroup(holdingCode string, authUsername string, groupGUID string) ([]string, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDocGroup, err := svc.repoImageGroup.FindByGuid(ctx, holdingCode, groupGUID)

	if err != nil {
		return []string{}, err
	}

	if len(findDocGroup.GuidFixed) < 1 {
		return []string{}, nil
	}

	updatedAt := svc.timeNowFnc()
	newImageGroupGUIDs := []string{}
	for _, imageRef := range *findDocGroup.ImageReferences {
		imageGroupGUID := svc.newDocumentImageGroupGUIDFnc()

		tags := []string{}
		if findDocGroup.Tags != nil {
			tags = *findDocGroup.Tags
		}

		docImageGroup := svc.createImageGroupByDocumentImage(holdingCode, authUsername, imageGroupGUID, imageRef, imageRef.ImageURI, tags, findDocGroup.TaskGUID, findDocGroup.PathTask, updatedAt, 0)

		newXOrderDocImgGroup, _ := svc.newXOrderDocumentImageGroup(ctx, holdingCode, docImageGroup.TaskGUID)

		docImageGroup.XOrder = newXOrderDocImgGroup
		_, err = svc.repoImageGroup.Create(ctx, docImageGroup)

		if err != nil {
			return []string{}, err
		}

		newImageGroupGUIDs = append(newImageGroupGUIDs, imageGroupGUID)
	}

	svc.repoImageGroup.DeleteByGuidfixed(ctx, holdingCode, groupGUID)

	return newImageGroupGUIDs, nil
}

func (svc DocumentImageService) ListDocumentImageGroup(holdingCode string, filters map[string]interface{}, pageable micromodels.Pageable) ([]models.DocumentImageGroupInfo, mongopagination.PaginationData, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	searchInFields := []string{"title"}
	docList, pagination, err := svc.repoImageGroup.FindPageImageGroup(ctx, holdingCode, filters, searchInFields, pageable)

	return docList, pagination, err
}

func (svc DocumentImageService) GetDocumentImageDocRefGroup(holdingCode string, docImageGroupGUID string) (models.DocumentImageGroupInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	doc, err := svc.repoImageGroup.FindByGuid(ctx, holdingCode, docImageGroupGUID)

	if err != nil {
		return models.DocumentImageGroupInfo{}, err
	}

	if doc.ID.IsZero() {
		return models.DocumentImageGroupInfo{}, errors.New("document not found")
	}

	return doc.DocumentImageGroupInfo, nil
}

func (svc DocumentImageService) GetDocumentImageGroupByDocRef(holdingCode string, docRef string) (models.DocumentImageGroupInfo, error) {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	findDoc, err := svc.repoImageGroup.FindOne(ctx, holdingCode, bson.M{"references.docno": docRef})

	if err != nil {
		return models.DocumentImageGroupInfo{}, err
	}

	if findDoc.ID == primitive.NilObjectID {
		return models.DocumentImageGroupInfo{}, errors.New("document not found")
	}

	return findDoc.DocumentImageGroupInfo, nil

}

func (svc DocumentImageService) XSortsUpdate(ctx context.Context, holdingCode string, authUsername string, taskGUID string, xsorts []models.XSortDocumentImageGroupRequest) error {
	for _, xsort := range xsorts {
		if len(xsort.GUIDFixed) < 1 {
			continue
		}

		err := svc.repoImageGroup.UpdateXOrder(ctx, holdingCode, taskGUID, xsort.GUIDFixed, xsort.XOrder)

		if err != nil {
			return err
		}
	}

	return nil
}

func (svc DocumentImageService) isDocumentImageGroupHasReferenced(doc models.DocumentImageGroupDoc) bool {
	return doc.References != nil && len(doc.References) > 0
}

func (svc DocumentImageService) isDocumentImageHasReferenced(doc models.DocumentImageDoc) bool {
	return doc.References != nil && len(doc.References) > 0
}

func (svc DocumentImageService) documentImageToImageReference(documentImageGUID string, documentImage models.DocumentImage, authUsername string, createdAt time.Time) models.ImageReference {
	return models.ImageReference{
		ImageReferenceBody: models.ImageReferenceBody{
			XOrder:            1,
			DocumentImageGUID: documentImageGUID,
		},
		BillCount:      documentImage.BillCount,
		ImageURI:       documentImage.ImageURI,
		CloneImageFrom: documentImage.CloneImageFrom,
		Name:           documentImage.Name,
		UploadedBy:     authUsername,
		UploadedAt:     createdAt,
		MetaFileAt:     documentImage.MetaFileAt,
	}
}

func (svc DocumentImageService) createImageGroupByDocumentImage(holdingCode string, authUsername string, imageGroupGUID string, documentImageRef models.ImageReference, imageURI string, tags []string, fileFolderGUID string, pathTask string, createdAt time.Time, billCount float64) models.DocumentImageGroupDoc {
	docDataImageGroup := models.DocumentImageGroupDoc{}
	docDataImageGroup.HoldingCode = holdingCode
	docDataImageGroup.GuidFixed = imageGroupGUID
	docDataImageGroup.Title = documentImageRef.Name
	docDataImageGroup.References = []models.Reference{}
	docDataImageGroup.Tags = &tags
	docDataImageGroup.ImageReferences = &[]models.ImageReference{
		documentImageRef,
	}
	if billCount > 0 {
		docDataImageGroup.BillCount = billCount
	} else {
		docDataImageGroup.BillCount = documentImageRef.BillCount
	}

	docDataImageGroup.TaskGUID = fileFolderGUID
	docDataImageGroup.PathTask = pathTask
	docDataImageGroup.Status = models.IMAGE_PENDING

	docDataImageGroup.CreatedBy = authUsername
	docDataImageGroup.CreatedAt = createdAt

	docDataImageGroup.UploadedBy = documentImageRef.UploadedBy
	docDataImageGroup.UploadedAt = documentImageRef.UploadedAt

	return docDataImageGroup
}

func (svc DocumentImageService) clearCreateDocumentImageGroupByDocumentGUIDs(holdingCode string, docImageGroupGUIDs []string, docImageGUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.repoImageGroup.RemoveDocumentImageByDocumentImageGUIDs(ctx, holdingCode, docImageGUIDs)
	if err != nil {
		return err
	}

	err = svc.repoImageGroup.DeleteByGUIDsIsDocumentImageEmpty(ctx, holdingCode, docImageGroupGUIDs)
	if err != nil {
		return err
	}

	return nil
}

func (svc DocumentImageService) clearUpdateDocumentImageGroupByDocumentGUIDs(holdingCode string, docGroupGUID string, clearDocImageGUIDs []string, docImageGUIDs []string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	err := svc.repoImageGroup.RemoveDocumentImageByDocumentImageGUIDsWithoutDocumentImageGroupGUID(ctx, holdingCode, docGroupGUID, docImageGUIDs)
	if err != nil {
		return err
	}

	err = svc.repoImageGroup.DeleteByGUIDsIsDocumentImageEmptyWithoutDocumentImageGroupGUID(ctx, holdingCode, docGroupGUID, clearDocImageGUIDs)
	if err != nil {
		return err
	}

	err = svc.repoImageGroup.DeleteByGUIDIsDocumentImageEmpty(ctx, holdingCode, docGroupGUID)
	if err != nil {
		return err
	}

	return nil
}

func (svc DocumentImageService) messageQueueReCountDocumentImageGroup(ctx context.Context, holdingCode string, taskGUID string) (int, error) {

	docList, err := svc.repoImageGroup.FindStatusByDocumentImageGroupTask(ctx, holdingCode, taskGUID)

	if err != nil {
		return 0, err
	}

	countDoc := len(docList)
	if countDoc < 1 {
		return 0, nil
	}

	totalStatus := map[int8]int{}
	BillCount := 0.0
	ReferenceCount := 0.0
	for _, doc := range docList {
		BillCount = BillCount + doc.BillCount
		ReferenceCount = ReferenceCount + float64(len(doc.References))
	}
	for i := 0; i <= models.IMAGE_FROM_REJECT; i++ {
		totalStatus[int8(i)] = 0
	}

	for _, doc := range docList {
		if _, ok := totalStatus[doc.Status]; !ok {
			totalStatus[doc.Status] = 0
		}
		totalStatus[doc.Status] = totalStatus[doc.Status] + 1
	}

	countStatus := []models.CountStatus{}
	for status, count := range totalStatus {
		countStatus = append(countStatus, models.CountStatus{
			Status: status,
			Count:  count,
		})
	}

	taskMsg := models.DocumentImageTaskChangeMessage{
		HoldingCode:      holdingCode,
		TaskGUID:         taskGUID,
		Count:            countDoc,
		CountStatus:      countStatus,
		BillCount:        BillCount,
		ReferenceCount:   ReferenceCount,
		ReferenceBalance: BillCount - ReferenceCount,
	}

	err = svc.repoMessagequeue.TaskChange(taskMsg)
	if err != nil {
		return 0, err
	}

	return countDoc, nil
}

func (svc DocumentImageService) newXOrderDocumentImageGroup(ctx context.Context, holdingCode string, taskGUID string) (int, error) {

	findDoc, err := svc.repoImageGroup.FindLastOneByTask(ctx, holdingCode, taskGUID)

	if err != nil {
		return 0, err
	}

	if len(findDoc.GuidFixed) < 1 {
		return 0, nil
	}

	return findDoc.XOrder + 1, nil
}

// UpdateDocNoInReferences updates all references from oldDocNo to newDocNo
// in both documentImageGroups and documentImages collections
func (svc DocumentImageService) UpdateDocNoInReferences(holdingCode string, module string, oldDocNo string, newDocNo string) error {

	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	// 1. Find all documentImageGroups that have the old reference
	oldRef := models.Reference{
		Module: module,
		DocNo:  oldDocNo,
	}

	groups, err := svc.repoImageGroup.FindByReference(ctx, holdingCode, oldRef)
	if err != nil {
		return fmt.Errorf("failed to find document image groups: %w", err)
	}

	// 2. Update each group's references
	for _, group := range groups {
		updated := false

		// Update references array
		if group.References != nil {
			for i := range group.References {
				if group.References[i].Module == module && group.References[i].DocNo == oldDocNo {
					group.References[i].DocNo = newDocNo
					updated = true
				}
			}
		}

		if updated {
			// Save the updated group
			err = svc.repoImageGroup.Update(ctx, holdingCode, group.GuidFixed, group)
			if err != nil {
				return fmt.Errorf("failed to update document image group %s: %w", group.GuidFixed, err)
			}
		}
	}

	// 3. Find and update all documentImages that have the old reference
	images, err := svc.repoImage.FindByReference(ctx, holdingCode, oldRef)
	if err != nil {
		return fmt.Errorf("failed to find document images: %w", err)
	}

	// 4. Update each image's references
	for _, image := range images {
		updated := false

		// Update references array
		if image.References != nil {
			for i := range image.References {
				if image.References[i].Module == module && image.References[i].DocNo == oldDocNo {
					image.References[i].DocNo = newDocNo
					updated = true
				}
			}
		}

		if updated {
			// Save the updated image
			err = svc.repoImage.Update(ctx, holdingCode, image.GuidFixed, image)
			if err != nil {
				return fmt.Errorf("failed to update document image %s: %w", image.GuidFixed, err)
			}
		}
	}

	return nil
}
