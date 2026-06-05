package documentimage

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"time"

	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/documentwarehouse/documentimage/models"
	"smlcloudplatform/internal/documentwarehouse/documentimage/repositories"
	"smlcloudplatform/internal/documentwarehouse/documentimage/services"
	"smlcloudplatform/pkg/microservice"

	journalRepo "smlcloudplatform/internal/vfgl/journal/repositories"
	journalSvc "smlcloudplatform/internal/vfgl/journal/services"

	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"

	"go.mongodb.org/mongo-driver/bson"
)

type IDocumentImageHttp interface {
	RegisterHttp()
	SearchDocumentImage(ctx microservice.IContext) error
	GetDocumentImageInfo(ctx microservice.IContext) error
	UploadDocumentImage(ctx microservice.IContext) error
	UpdateDocumentImage(ctx microservice.IContext) error
	DeleteDocumentImage(ctx microservice.IContext) error
}

type DocumentImageHttp struct {
	Module       string
	ms           *microservice.Microservice
	cfg          config.IConfig
	service      services.IDocumentImageService
	svcWsJournal journalSvc.IJournalWebsocketService
}

func NewDocumentImageHttp(ms *microservice.Microservice, cfg config.IConfig) *DocumentImageHttp {

	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())
	prod := ms.Producer(cfg.MQConfig())

	repo := repositories.NewDocumentImageRepository(pst)
	repoImageGroup := repositories.NewDocumentImageGroupRepository(pst)
	repoMessagequeue := repositories.NewDocumentImageMessageQueueRepository(prod)

	azureblob := microservice.NewFilePersister()
	svc := services.NewDocumentImageService(repo, repoImageGroup, repoMessagequeue, azureblob)

	docImageRepo := repositories.NewDocumentImageRepository(pst)

	cacheRepo := journalRepo.NewJournalCacheRepository(cache)
	svcWsJournal := journalSvc.NewJournalWebsocketService(docImageRepo, cacheRepo, time.Duration(30)*time.Minute)

	return &DocumentImageHttp{
		Module:       "GL",
		ms:           ms,
		cfg:          cfg,
		service:      svc,
		svcWsJournal: svcWsJournal,
	}
}

func (h DocumentImageHttp) RegisterHttp() {
	h.ms.GET("/documentimage", h.SearchDocumentImage)
	// h.ms.GET("/documentimage/special", h.DocumentImageSpecial)
	h.ms.GET("/documentimage/:guid", h.GetDocumentImageInfo)
	h.ms.POST("/documentimage/upload", h.UploadDocumentImage)
	h.ms.POST("/documentimage", h.CreateDocumentImage)
	h.ms.POST("/documentimage/bulk", h.BulkCreateDocumentImage)
	h.ms.PUT("/documentimage/:guid/imageedit", h.CreateDocumentImageEdit)
	h.ms.PUT("/documentimage/:guid/comment", h.CreateDocumentImageComment)
	// h.ms.PUT("/documentimage/status/:id", h.UpdateDocumentImageStatus)
	// h.ms.DELETE("/documentimage/:guid", h.DeleteDocumentImage)
	// h.ms.PUT("/documentimage/documentref/status/:docref", h.UpdateDocumentImageStatusByDocumentRef)

	h.ms.GET("/documentimagegroup", h.ListDocumentImageGroup)
	h.ms.GET("/documentimagegroup/:guid", h.GetDocumentImageGroup)
	h.ms.POST("/documentimagegroup", h.CreateDocumentImageGroup)
	h.ms.PUT("/documentimagegroup/:guid", h.UpdateDocumentImageGroup)
	h.ms.PUT("/documentimagegroup/xsort/:taskguid", h.UpdateXSort)
	h.ms.PUT("/documentimagegroup/:guid/documentimages", h.UpdateImageReferenceByDocumentImageGroup)
	h.ms.PUT("/documentimagegroup/:guid/reference", h.UpdateReferenceByDocumentImageGroup)
	h.ms.PUT("/documentimagegroup/:guid/ungroup", h.UngroupDocumentImageGroup)
	h.ms.PUT("/documentimagegroup/:guid/images", h.UpdateDocumentImageGroup)
	h.ms.PUT("/documentimagegroup/:guid/status", h.UpdateStatusDocumentImageGroup)
	h.ms.PUT("/documentimagegroup/:guid/tags", h.UpdateTagsDocumentImageGroup)
	h.ms.PUT("/documentimagegroup/task/:taskguid/status", h.UpdateStatusDocumentImageGroupByTask)
	h.ms.PUT("/documentimagegroup/task/:taskguid/recount", h.ReCountStatusDocumentImageGroupByTask)
	h.ms.DELETE("/documentimagegroup/:guid", h.DeleteDocumentImageGroup)
	h.ms.DELETE("/documentimagegroup", h.DeleteDocumentImageGroups)

	h.ms.GET("/documentimagegroup/docref/:docref", h.GetDocumentImageGroupByDocRefInfo)
}

func (h DocumentImageHttp) DocumentImageSpecial(ctx microservice.IContext) error {

	err := h.service.UpdateDocumentImageReferenceGroup()
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})
	return nil
}

func (h DocumentImageHttp) Info(ctx microservice.IContext) error {

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// List Document Image
// @Summary		List Document Image
// @Description List Document Image
// @Tags		DocumentImage
// @Param		q		query	string		false  "Search Value"
// @Param		status		query	integer		false  "Status Value"
// @Param		module		query	string		false  "Module Value"
// @Param		docguidref		query	string		false  "Doc GUID Ref Value"
// @Param		documentref		query	string		false  "Document Ref Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Size"
// @Param		sort	query	string		false  "Sort"
// @Accept 		json
// @Success		200	{object}	models.DocumentImagePageResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimage [get]
func (h DocumentImageHttp) SearchDocumentImage(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageable := utils.GetPageable(ctx.QueryParam)

	matchFilters := map[string]interface{}{}

	statusRaw := strings.TrimSpace(ctx.QueryParam("status"))

	statusFilter := []int{}
	if len(statusRaw) > 0 {

		statusRawArray := strings.Split(statusRaw, ",")

		for _, status := range statusRawArray {
			tempStatus, err := strconv.Atoi(status)
			if err == nil {
				statusFilter = append(statusFilter, tempStatus)
			}
		}
	}

	lenStatus := len(statusFilter)
	if lenStatus > 0 {

		if lenStatus == 1 {
			matchFilters["status"] = statusFilter[0]
		} else {
			matchFilters["status"] = bson.M{"$in": statusFilter}
		}
	}

	module := strings.TrimSpace(ctx.QueryParam("module"))

	if len(module) > 0 {
		matchFilters["references.module"] = module
	}

	docGuidRef := strings.TrimSpace(ctx.QueryParam("doc_guid_ref"))

	if len(docGuidRef) > 0 {
		matchFilters["doc_guid_ref"] = docGuidRef
	}

	documentRef := strings.TrimSpace(ctx.QueryParam("documentref"))

	docRefReserve := strings.TrimSpace(ctx.QueryParam("docref-reserve"))

	if len(docRefReserve) > 0 && docRefReserve != "0" {
		docRefPoolList, err := h.svcWsJournal.GetAllDocRefPool(holdingCode)

		if err == nil {
			docRefList := []string{}
			for docRef := range docRefPoolList {
				docRefList = append(docRefList, docRef)
			}
			matchFilters["documentref"] = bson.M{"$eq": documentRef, "$nin": docRefList}
		}

	} else if len(documentRef) > 0 {
		matchFilters["documentref"] = documentRef
	}

	docList, pagination, err := h.service.SearchDocumentImage(holdingCode, matchFilters, pageable)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success:    true,
		Data:       docList,
		Pagination: pagination,
	})

	return nil
}

// Get Document Image Infomation godoc
// @Summary		Get Document Image Infomation
// @Description Get Document Image Infomation
// @Tags		DocumentImage
// @Param		guid  path      string  true  "document image guid"
// @Accept 		json
// @Success		200	{object}	models.DocumentImageInfoResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimage/{guid} [get]
func (h DocumentImageHttp) GetDocumentImageInfo(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("guid")
	doc, err := h.service.InfoDocumentImage(holdingCode, id)
	if err != nil {
		h.ms.Logger.Errorf("Error getting document %v: %v", id, err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}

// Create Document Image godoc
// @Summary		Create Document Image
// @Description Create Document Image
// @Tags		DocumentImage
// @Param		DocumentImage  body      models.DocumentImage  true  "DocumentImage"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimage [post]
func (h DocumentImageHttp) CreateDocumentImage(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &models.DocumentImageRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, imageGroupGUID, err := h.service.CreateDocumentImage(holdingCode, authUsername, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, map[string]interface{}{
		"success": true,
		"id":      idx,
		"groupid": imageGroupGUID,
	})
	return nil
}

// Create Document Image Edit godoc
// @Summary		Create Document Image Edit
// @Description Create Document Image Edit
// @Tags		DocumentImage
// @Param		guid  path      string  true  "document image guid"
// @Param		ImageEditRequest  body      models.ImageEditRequest  true  "ImageEditRequest"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimage/{guid}/imageedit [put]
func (h DocumentImageHttp) CreateDocumentImageEdit(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docImageGUID := ctx.Param("guid")

	if len(docImageGUID) == 0 {
		ctx.ResponseError(http.StatusBadRequest, "document image guid is required")
		return nil
	}

	docReq := &models.ImageEditRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.service.CreateImageEdit(holdingCode, authUsername, docImageGUID, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, map[string]interface{}{
		"success": true,
	})
	return nil
}

// Create Document Image Comment godoc
// @Summary		Create Document Image Comment
// @Description Create Document Image Comment
// @Tags		DocumentImage
// @Param		guid  path      string  true  "document image guid"
// @Param		CommentRequest  body      models.CommentRequest  true  "CommentRequest"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimage/{guid}/comment [put]
func (h DocumentImageHttp) CreateDocumentImageComment(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docImageGUID := ctx.Param("guid")

	if len(docImageGUID) == 0 {
		ctx.ResponseError(http.StatusBadRequest, "document image guid is required")
		return nil
	}

	docReq := &models.CommentRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.service.CreateImageComment(holdingCode, authUsername, docImageGUID, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, map[string]interface{}{
		"success": true,
	})
	return nil
}

// Bulk Create Document Image godoc
// @Summary		Create Document Image
// @Description Create Document Image
// @Tags		DocumentImage
// @Param		DocumentImage  body      []models.DocumentImage  true  "DocumentImage"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimage/bulk [post]
func (h DocumentImageHttp) BulkCreateDocumentImage(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	holdingCode := ctx.UserInfo().HoldingCode
	input := ctx.ReadInput()

	docReq := &[]models.DocumentImageRequest{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.service.BulkCreateDocumentImage(holdingCode, authUsername, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
	})
	return nil
}

/*
// Update Document Image Status godoc
// @Summary		Update Document Image Status
// @Description Update Document Image Status
// @Tags		DocumentImageStatus
// @Param		id  path      string  true  "ID"
// @Param		DocumentImageStatus  body      models.DocumentImageStatus  true  "DocumentImageStatus"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimage/status/{id} [put]
func (h DocumentImageHttp) UpdateDocumentImageStatus(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.DocumentImageStatus{}
	err := json.Unmarshal([]byte(input), &docReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.service.UpdateDocumentImageStatus(holdingCode, id, docReq.DocGUIDRef, docReq.Status)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      id,
	})
	return nil
}
*/

/*
// Update Document Image Status By Document Ref godoc
// @Summary		Update Document Image Status By Document Ref
// @Description Update Document Image Status By Document Ref
// @Tags		DocumentImageStatusByDocumentRef
// @Param		docref  path      string  true  "Document Ref"
// @Param		DocumentImageStatus  body      models.DocumentImageStatus  true  "DocumentImageStatus"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimage/documentref/status/{docref} [put]
func (h DocumentImageHttp) UpdateDocumentImageStatusByDocumentRef(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	docref := ctx.Param("docref")
	input := ctx.ReadInput()

	docReq := &models.DocumentImageStatus{}
	err := json.Unmarshal([]byte(input), &docReq)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.service.UpdateDocumentImageStatusByDocumentRef(holdingCode, docref, docReq.DocGUIDRef, docReq.Status)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      docref,
	})
	return nil
}
*/

// Upload Document Image
// @Description Upload Document Image
// @Tags		DocumentImage
// @Accept 		json
// @Param		module  query      string  true  "Module"
// @Param		file  formData      file  true  "Image"
// @Success		200	{array}	models.DocumentImageInfoResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Failure		400	{object}	common.AuthResponseFailed
// @Failure		500	{object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimage/upload [post]
func (h DocumentImageHttp) UploadDocumentImage(ctx microservice.IContext) error {

	// moduleName := ctx.QueryParam("module")
	// if moduleName == "" {
	// 	ctx.ResponseError(400, "No Module Special")
	// 	return errors.New("Upload Image Without Module")
	// }

	fileHeader, err := ctx.FormFile("file")
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	holdingCode := userInfo.HoldingCode

	idx, err := h.service.UploadDocumentImage(holdingCode, authUsername, fileHeader)

	if err != nil {
		ctx.ResponseError(400, err.Error())
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		Data:    idx,
	})
	return nil
}

// List Document Image Group
// @Description Get Document Image Group
// @Tags		DocumentImageGroup
// @Accept 		json
// @Tags		Restaurant
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Size"
// @Param		pathtask	query	string		false  "Filter Path Job"
// @Param		taskguid	query	string		false  "Filter by Job guidfixed"
// @Param		reserve	query	integer		false  "เอกสารที่มีการจอง,0 not filter, 1 filter"
// @Param		reserveby	query	integer		false  "เอกสารที่มีการจองทั้งหมดหรือตามผู้ใช้,0 all, 1 user"
// @Param		status	query	integer		false  "ex. status=0 status=1,2,3"
// @Param		ref	query	integer		false  "document reference: empty not filter, 1 not reference, 2 referenced"
// @Param		fromdate		query	string		false  "From Date (YYYY-MM-DD)"
// @Param		todate		query	string		false  "To Date (YYYY-MM-DD)"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimagegroup [get]
func (h DocumentImageHttp) ListDocumentImageGroup(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	pageable := utils.GetPageable(ctx.QueryParam)

	matchFilters := map[string]interface{}{}

	docRefReserve := strings.TrimSpace(ctx.QueryParam("reserve"))
	reserveBy := strings.TrimSpace(ctx.QueryParam("reserveby"))
	statusReq := strings.TrimSpace(ctx.QueryParam("status"))
	isref := strings.TrimSpace(ctx.QueryParam("ref"))
	path := strings.TrimSpace(ctx.QueryParam("pathtask"))
	folder := strings.TrimSpace(ctx.QueryParam("task_guid"))

	fromDateStr := strings.TrimSpace(ctx.QueryParam("fromdate"))
	toDateStr := strings.TrimSpace(ctx.QueryParam("todate"))

	if len(statusReq) > 0 {
		tempStatusStrArr := strings.Split(statusReq, ",")

		tempStatus := []int{}
		for _, temp := range tempStatusStrArr {
			status, err := strconv.Atoi(temp)
			if err == nil {
				tempStatus = append(tempStatus, status)
			}
		}
		matchFilters["status"] = bson.M{"$in": tempStatus}
	}

	if len(isref) > 0 && isref != "0" {
		if isref == "1" {
			matchFilters["references.module"] = bson.M{"$ne": h.Module}
		} else if isref == "2" {
			matchFilters["references.module"] = h.Module
		}
	}

	documentImageGUID := strings.TrimSpace(ctx.QueryParam("document_image_guid"))

	if len(documentImageGUID) > 0 {
		matchFilters["imagereferences.documentimageguid"] = documentImageGUID
	}

	if len(docRefReserve) > 0 && docRefReserve != "0" {

		if len(reserveBy) == 0 {
			reserveBy = "0"
		}
		if reserveBy == "1" {
			// ✅ Filter เฉพาะของ user ที่ request (reserveby=1)
			docRef, err := h.svcWsJournal.GetDocRefUserPool(holdingCode, userInfo.Username)

			if err == nil && len(docRef) > 0 {
				// User นี้ select เอกสารไว้ -> แสดงเฉพาะเอกสารนั้น
				matchFilters["guidfixed"] = docRef
			} else {
				// ✅ User นี้ยังไม่ได้ select -> ไม่แสดงอะไรเลย (empty result)
				matchFilters["guidfixed"] = "" // force empty result
			}
		} else {
			// ✅ Filter ทั้ง shop (reserveby=0 หรือไม่ส่ง) - แสดงเฉพาะที่ยังไม่มีใคร select
			docRefPoolList, err := h.svcWsJournal.GetAllDocRefPool(holdingCode)

			if err == nil {
				docRefList := []string{}
				for docRef := range docRefPoolList {
					docRefList = append(docRefList, docRef)
				}
				matchFilters["guidfixed"] = bson.M{"$nin": docRefList}
			}
		}

	}

	if len(fromDateStr) > 0 && len(toDateStr) > 0 {
		fromDate, err1 := time.Parse("2006-01-02", fromDateStr)
		toDate, err2 := time.Parse("2006-01-02", toDateStr)

		if err1 == nil && err2 == nil {
			matchFilters["uploaded_at"] = bson.M{
				"$gte": fromDate,
				"$lt":  toDate.AddDate(0, 0, 1),
			}
		}
	}

	if len(path) > 0 {
		matchFilters["pathtask"] = path
	}

	if len(folder) > 0 {
		matchFilters["task_guid"] = folder
	}

	docList, pagination, err := h.service.ListDocumentImageGroup(holdingCode, matchFilters, pageable)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success:    true,
		Data:       docList,
		Pagination: pagination,
	})

	return nil
}

// Get Document Image Group
// @Description Get Document Image Group
// @Tags		DocumentImageGroup
// @Accept 		json
// @Param		guid  path      string  true  "document image group guid"
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimagegroup/{guid} [get]
func (h DocumentImageHttp) GetDocumentImageGroup(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	docImageGroupGUID := ctx.Param("guid")

	doc, err := h.service.GetDocumentImageDocRefGroup(holdingCode, docImageGroupGUID)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})

	return nil
}

// Save Document Image Group
// @Description Save Document Image Group
// @Tags		DocumentImageGroup
// @Param		DocumentImageGroup  body      models.DocumentImageGroup  true  "Document Image Group"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimagegroup [post]
func (h DocumentImageHttp) CreateDocumentImageGroup(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	input := ctx.ReadInput()

	docImageGroup := &models.DocumentImageGroup{}

	err := json.Unmarshal([]byte(input), &docImageGroup)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.service.CreateDocumentImageGroup(holdingCode, authUsername, *docImageGroup)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      idx,
	})

	return nil
}

// Update Document Image Group
// @Description Update Document Image Group
// @Tags		DocumentImageGroup
// @Accept 		json
// @Param		guid  path      string  true  "document image group guid"
// @Param		DocumentImageGroup  body      models.DocumentImageGroup  true  "Document Image Group"
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimagegroup/{guid} [put]
func (h DocumentImageHttp) UpdateDocumentImageGroup(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()

	docImageGroupGUID := ctx.Param("guid")

	input := ctx.ReadInput()

	docImageGroup := &models.DocumentImageGroup{}

	err := json.Unmarshal([]byte(input), &docImageGroup)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.service.UpdateDocumentImageGroup(userInfo.HoldingCode, userInfo.Username, docImageGroupGUID, *docImageGroup)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Update Document Image Group
// @Description Update Document Image Group
// @Tags		DocumentImageGroup
// @Accept 		json
// @Param		guid  path      string  true  "document image group guid"
// @Param		ImageReferenceBody  body      models.ImageReferenceBody  true  "Image Reference"
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimagegroup/{guid}/documentimages [put]
func (h DocumentImageHttp) UpdateImageReferenceByDocumentImageGroup(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()

	docImageGroupGUID := ctx.Param("guid")

	input := ctx.ReadInput()

	docImages := &[]models.ImageReferenceBody{}

	err := json.Unmarshal([]byte(input), &docImages)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.service.UpdateImageReferenceByDocumentImageGroup(userInfo.HoldingCode, userInfo.Username, docImageGroupGUID, *docImages)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Update Reference In Document Image Group
// @Description Update Reference In Document Image Group
// @Tags		DocumentImageGroup
// @Accept 		json
// @Param		guid  path      string  true  "document image group guid"
// @Param		Reference  body      models.Reference  true  "Reference"
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimagegroup/{guid}/reference [put]
func (h DocumentImageHttp) UpdateReferenceByDocumentImageGroup(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()

	docImageGroupGUID := ctx.Param("guid")

	input := ctx.ReadInput()

	docImages := &models.Reference{}

	err := json.Unmarshal([]byte(input), &docImages)
	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.service.UpdateReferenceByDocumentImageGroup(userInfo.HoldingCode, userInfo.Username, docImageGroupGUID, *docImages)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Ungroup Document Image Group
// @Description Ungroup Document Image Group
// @Tags		DocumentImageGroup
// @Accept 		json
// @Param		guid  path      string  true  "document image group guid"
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimagegroup/{guid}/ungroup [put]
func (h DocumentImageHttp) UngroupDocumentImageGroup(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()

	docImageGroupGUID := ctx.Param("guid")

	guids, err := h.service.UnGroupDocumentImageGroup(userInfo.HoldingCode, userInfo.Username, docImageGroupGUID)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    guids,
	})

	return nil
}

// Update Status Document Image Group
// @Description Update Status Document Image Group
// @Tags		DocumentImageGroup
// @Accept 		json
// @Param		guid  path      string  true  "document image group guid"
// @Param		models.Status  body      models.Status  true  "status"
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimagegroup/{guid}/status [put]
func (h DocumentImageHttp) UpdateStatusDocumentImageGroup(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	docImageGroupGUID := ctx.Param("guid")

	input := ctx.ReadInput()

	docReq := models.Status{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.service.UpdateStatusDocumentImageGroup(holdingCode, authUsername, docImageGroupGUID, docReq.Status)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// ReCount Status Document Image Group By GUID
// @Description ReCount Status Document Image Group By GUID
// @Tags		DocumentImageGroup
// @Accept 		json
// @Param		guid  path      string  true  "document image group guid"
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimagegroup/{guid}/recount [put]
func (h DocumentImageHttp) ReCountDocumentImageGroupByGUID(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	guid := ctx.Param("guid")

	err := h.service.ReCountStatusDocumentImageGroupByGUID(holdingCode, authUsername, guid)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Message: "Recount successfully",
	})

	return nil
}

// Update Status Document Image Group By Task
// @Description Update Status Document Image Group By Task
// @Tags		DocumentImageGroup
// @Accept 		json
// @Param		taskguid  path      string  true  "task guid"
// @Param		models.Status  body      models.Status  true  "status"
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimagegroup/task/{taskguid}/status [put]
func (h DocumentImageHttp) UpdateStatusDocumentImageGroupByTask(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	taskGUID := ctx.Param("task_guid")

	input := ctx.ReadInput()

	docReq := models.Status{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.service.UpdateStatusDocumentImageGroupByTask(holdingCode, authUsername, taskGUID, docReq.Status)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// ReCount Status Document Image Group By Task
// @Description ReCount Status Document Image Group By Task
// @Tags		DocumentImageGroup
// @Accept 		json
// @Param		taskguid  path      string  true  "task guid"
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimagegroup/task/{taskguid}/recount [put]
func (h DocumentImageHttp) ReCountStatusDocumentImageGroupByTask(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	docImageGroupGUID := ctx.Param("task_guid")

	err := h.service.ReCountStatusDocumentImageGroupByTask(holdingCode, authUsername, docImageGroupGUID)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Update Status Document Image Group
// @Description Update Status Document Image Group
// @Tags		DocumentImageGroup
// @Accept 		json
// @Param		guid  path      string  true  "document image group guid"
// @Param		[]string  body      []string  true  "tags"
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimagegroup/{guid}/tags [put]
func (h DocumentImageHttp) UpdateTagsDocumentImageGroup(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	docImageGroupGUID := ctx.Param("guid")

	input := ctx.ReadInput()

	if len(input) < 2 {
		ctx.ResponseError(400, "payload invalid")
		return nil
	}

	tags := []string{}
	err := json.Unmarshal([]byte(input), &tags)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.service.UpdateTagsInDocumentImageGroup(holdingCode, authUsername, docImageGroupGUID, tags)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get Document Image Group By document reference Infomation godoc
// @Summary		Get Document Image Group By document reference Infomation
// @Description Get Document Image Group By document reference Infomation
// @Tags		DocumentImageGroup
// @Param		docref  path      string  true  "document reference"
// @Accept 		json
// @Success		200	{object}	models.DocumentImageInfoResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimagegroup/docref/{docref} [get]
func (h DocumentImageHttp) GetDocumentImageGroupByDocRefInfo(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode

	docRef := ctx.Param("docref")
	doc, err := h.service.GetDocumentImageGroupByDocRef(holdingCode, docRef)
	if err != nil {
		h.ms.Logger.Errorf("Error getting document %v: %v", docRef, err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}

// Delete Document Image Group godoc
// @Summary		Delete Document Image Group
// @Description Delete Document Image Group
// @Tags		DocumentImageGroup
// @Param		guid  path      string  true  "document image group guid"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimagegroup/{guid} [delete]
func (h DocumentImageHttp) DeleteDocumentImageGroup(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	id := ctx.Param("guid")

	err := h.service.DeleteDocumentImageGroupByGuid(holdingCode, authUsername, id)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
	})
	return nil
}

// Delete Many Document Image Group godoc
// @Summary		Delete Many Document Image Group
// @Description Delete Many Document Image Group
// @Tags		DocumentImageGroup
// @Param		guids  body      []string  true  "document image group guids"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimagegroup [delete]
func (h DocumentImageHttp) DeleteDocumentImageGroups(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	input := ctx.ReadInput()

	guids := []string{}
	err := json.Unmarshal([]byte(input), &guids)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	err = h.service.DeleteDocumentImageGroupByGuids(holdingCode, authUsername, guids)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    guids,
	})
	return nil
}

// Update XSort Document Image Group godoc
// @Description Update XSort Document Image Group
// @Tags		DocumentImageGroup
// @Param		XSort  body      []models.XSortDocumentImageGroupRequest  true  "XSort"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /documentimagegroup/xsort/{taskguid} [put]
func (h DocumentImageHttp) UpdateXSort(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	holdingCode := userInfo.HoldingCode
	authUsername := userInfo.Username

	input := ctx.ReadInput()

	taskGUID := ctx.Param("task_guid")

	if len(taskGUID) < 1 {
		ctx.ResponseError(http.StatusBadRequest, "taskguid is required")
		return nil
	}

	reqBody := []models.XSortDocumentImageGroupRequest{}
	err := json.Unmarshal([]byte(input), &reqBody)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	err = h.service.XSortsUpdate(context.Background(), holdingCode, authUsername, taskGUID, reqBody)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})
	return nil
}
