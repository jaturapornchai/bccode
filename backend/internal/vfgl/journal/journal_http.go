package journal

import (
	"encoding/json"
	"fmt"
	"net/http"
	msConfig "smlcloudplatform/internal/config"
	documentImageModel "smlcloudplatform/internal/documentwarehouse/documentimage/models"
	repoDocumentimage "smlcloudplatform/internal/documentwarehouse/documentimage/repositories"
	serviceDocumentimage "smlcloudplatform/internal/documentwarehouse/documentimage/services"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/vfgl/journal/models"
	"smlcloudplatform/internal/vfgl/journal/repositories"
	"smlcloudplatform/internal/vfgl/journal/services"
	"smlcloudplatform/pkg/microservice"
	"strconv"
	"strings"
	"time"
)

type IJournalHttp interface{}

type JournalHttp struct {
	ms          *microservice.Microservice
	cfg         msConfig.IConfig
	svc         services.IJournalHttpService
	svcDocImage serviceDocumentimage.DocumentImageService
	Module      string
	// svcWebsocket services.IJournalWebsocketService
}

func NewJournalHttp(ms *microservice.Microservice, cfg msConfig.IConfig) JournalHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	prod := ms.Producer(cfg.MQConfig())
	// cache := ms.Cacher(cfg.CacherConfig())

	repo := repositories.NewJournalRepository(pst)
	mqRepo := repositories.NewJournalMqRepository(prod)
	svc := services.NewJournalHttpService(repo, mqRepo)

	// cacheRepo := repositories.NewJournalCacheRepository(cache)
	// svcWebsocket := services.NewJournalWebsocketService(repo, cacheRepo, time.Duration(30)*time.Minute)

	repoDocImage := repoDocumentimage.NewDocumentImageRepository(pst)
	repoDocImageGroup := repoDocumentimage.NewDocumentImageGroupRepository(pst)
	repoDocImageGroupMessagequeue := repoDocumentimage.NewDocumentImageMessageQueueRepository(prod)

	svcDocImage := serviceDocumentimage.NewDocumentImageService(repoDocImage, repoDocImageGroup, repoDocImageGroupMessagequeue, nil)

	return JournalHttp{
		Module:      "GL",
		ms:          ms,
		cfg:         cfg,
		svc:         svc,
		svcDocImage: svcDocImage,
		// svcWebsocket: svcWebsocket,
	}
}

func (h JournalHttp) RegisterHttp() {

	h.ms.GET("/gl/journal/re-gen-guid-empty", h.ReUpdateGuidJournalEmpty)
	h.ms.POST("/gl/journal/bulk", h.SaveBulk)

	h.ms.GET("/gl/journal", h.SearchJournal)
	h.ms.POST("/gl/journal", h.CreateJournal)
	h.ms.GET("/gl/journal/last-docno", h.GetLastDocNo)
	h.ms.GET("/gl/journal/duplicate-docnos", h.GetDuplicateDocNos)
	h.ms.GET("/gl/journal/check-vatdocno", h.CheckVatDocNo)
	h.ms.GET("/gl/journal/check-taxdocno", h.CheckTaxDocNo)
	h.ms.GET("/gl/journal/:id", h.InfoJournal)
	h.ms.GET("/gl/journal/rebuild/:shopid", h.RebuildPgJournal)
	h.ms.GET("/gl/journal/docno/:docno", h.InfoJournalByDocno)
	h.ms.GET("/gl/journal/docref/:doc", h.InfoJournalByDocumentRef)
	h.ms.PUT("/gl/journal/:id", h.UpdateJournal)
	h.ms.DELETE("/gl/journal/:id", h.DeleteJournal)
	h.ms.DELETE("/gl/journal", h.DeleteJournalByGUIDs)
	h.ms.DELETE("/gl/journal/batchid/:batchid", h.DeleteJournalByBatchID)

}

func (h JournalHttp) ReUpdateGuidJournalEmpty(ctx microservice.IContext) error {

	passCode := ctx.QueryParam("passcode")

	if len(passCode) < 1 {
		ctx.ResponseError(http.StatusNotFound, "not found")
		return nil
	}

	if passCode != "smldev" {
		ctx.ResponseError(http.StatusBadRequest, "passcode invalid")
		return nil
	}

	err := h.svc.ReGenerateGuidEmpty()

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Create Journal godoc
// @Summary		บันทึกข้อมูลรายวัน
// @Description บันทึกข้อมูลรายวัน
// @Tags		GL
// @Param		Journal  body      models.Journal  true  "ข้อมูลรายวัน"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/journal [post]
func (h JournalHttp) CreateJournal(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.Journal{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if len(docReq.DocumentRef) > 0 {
		docImageGroup, err := h.svcDocImage.GetDocumentImageDocRefGroup(shopID, docReq.DocumentRef)
		if err != nil {
			messageError := ""
			if err.Error() == "document not found" {
				messageError = "document image group not found"
			} else {
				messageError = err.Error()
			}

			ctx.ResponseError(400, messageError)
			return err
		}

		if len(docImageGroup.GuidFixed) < 1 {
			if err != nil {
				ctx.ResponseError(400, "document image group not found")
				return err
			}
		}
	}

	idx, err := h.svc.CreateJournal(shopID, authUsername, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	// Update reference after creating journal to include GuidFixed
	if len(docReq.DocumentRef) > 0 {
		err = h.svcDocImage.UpdateReferenceByDocumentImageGroup(shopID, authUsername, docReq.DocumentRef, documentImageModel.Reference{
			GuidFixed: idx,
			Module:    h.Module,
			DocNo:     docReq.DocNo,
		})

		if err != nil {
			ctx.ResponseError(http.StatusBadRequest, err.Error())
			return err
		}
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      idx,
	})
	return nil
}

// Rebuild Journal godoc
// @Summary		บันทึกข้อมูลรายวัน
// @Description บันทึกข้อมูลรายวัน
// @Tags		GL
// @Accept 		json
// @Param		shopid  path      string  true  "Journal Shop ID"
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/journal/rebuild/{shopid}  [get]
func (h JournalHttp) RebuildPgJournal(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	// shopID := ctx.UserInfo().ShopID
	shopID := ctx.Param("shopid")

	idx, err := h.svc.RebuildPgJournal(shopID, authUsername)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		Data:    idx,
	})
	return nil
}

// Update Journal godoc
// @Summary		แก้ไขข้อมูลรายวัน
// @Description แก้ไขข้อมูลรายวัน
// @Tags		GL
// @Param		id  path      string  true  "Journal ID"
// @Param		Journal  body      models.Journal  true  "ข้อมูลรายวัน"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/journal/{id} [put]
func (h JournalHttp) UpdateJournal(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.Journal{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if len(docReq.DocumentRef) > 0 {
		docImageGroup, err := h.svcDocImage.GetDocumentImageDocRefGroup(shopID, docReq.DocumentRef)
		if err != nil {
			messageError := ""
			if err.Error() == "document not found" {
				messageError = "document image group not found"
			} else {
				messageError = err.Error()
			}

			ctx.ResponseError(400, messageError)
			return err
		}

		if len(docImageGroup.GuidFixed) < 1 {
			if err != nil {
				ctx.ResponseError(400, "document image group not found")
				return err
			}
		}
	}

	journalInfo, err := h.svc.InfoJournal(shopID, id)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	oldDocNo, newDocNo, err := h.svc.UpdateJournal(id, shopID, authUsername, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	// Sync references when DocumentRef changes
	if journalInfo.DocumentRef != docReq.DocumentRef {

		if len(docReq.DocumentRef) > 0 {
			err = h.svcDocImage.UpdateReferenceByDocumentImageGroup(shopID, authUsername, docReq.DocumentRef, documentImageModel.Reference{
				GuidFixed: id,
				Module: h.Module,
				DocNo:  docReq.DocNo,
			})
			if err != nil {
				ctx.ResponseError(http.StatusBadRequest, err.Error())
				return err
			}
		}
	}

	// Sync references when DocNo changes
	if oldDocNo != newDocNo {
		err = h.svcDocImage.UpdateDocNoInReferences(shopID, h.Module, oldDocNo, newDocNo)
		if err != nil {
			ctx.ResponseError(http.StatusBadRequest, fmt.Sprintf("failed to update references: %s", err.Error()))
			return err
		}
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      id,
	})

	return nil
}

// Delete Journal godoc
// @Summary		ลบข้อมูลรายวัน
// @Description ลบข้อมูลรายวัน
// @Tags		GL
// @Param		id  path      string  true  "Journal ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/journal/{id} [delete]
func (h JournalHttp) DeleteJournal(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	journal, err := h.svc.InfoJournal(shopID, id)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	err = h.svc.DeleteJournal(id, shopID, authUsername)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	if len(journal.DocumentRef) > 0 {
		imageRef := documentImageModel.Reference{
			Module: h.Module,
			DocNo:  journal.DocNo,
		}
		h.svcDocImage.DeleteReferenceByDocumentImageGroup(shopID, authUsername, journal.DocumentRef, imageRef)
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
	})

	return nil
}

// Delete Journal godoc
// @Description Delete Journal
// @Tags		GL
// @Param		GL  body      []string  true  "GL GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/journal/ [delete]
func (h JournalHttp) DeleteJournalByGUIDs(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	input := ctx.ReadInput()

	docReq := []string{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.DeleteJournalByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Delete Journal By Batch ID godoc
// @Summary		ลบข้อมูลรายวัน By Batch ID
// @Description ลบข้อมูลรายวัน By Batch ID
// @Tags		GL
// @Param		batchid  path      string  true  "Journal Batch ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/journal/batchid/{batchid} [delete]
func (h JournalHttp) DeleteJournalByBatchID(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	batchID := ctx.Param("batchid")

	err := h.svc.DeleteJournalByBatchID(shopID, authUsername, batchID)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Get Journal Infomation godoc
// @Summary		แสดงรายละเอียดข้อมูลรายวัน
// @Description แสดงรายละเอียดข้อมูลรายวัน
// @Tags		GL
// @Param		id  path      string  true  "Journal Id"
// @Accept 		json
// @Success		200	{object}	models.JournalInfoResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/journal/{id} [get]
func (h JournalHttp) InfoJournal(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get Journal %v", id)
	doc, err := h.svc.InfoJournal(shopID, id)

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

// Get Journal Infomation By DocNo godoc
// @Summary		แสดงรายละเอียดข้อมูลรายวัน
// @Description แสดงรายละเอียดข้อมูลรายวัน
// @Tags		GL
// @Param		docno  path      string  true  "Journal DocNo"
// @Accept 		json
// @Success		200	{object}	models.JournalInfoResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/journal/docno/{docno} [get]
func (h JournalHttp) InfoJournalByDocno(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	docNo := ctx.Param("docno")

	h.ms.Logger.Debugf("Get Journal %v", docNo)
	doc, err := h.svc.InfoJournalByDocNo(shopID, docNo)

	if err != nil {
		h.ms.Logger.Errorf("Error getting document %v: %v", docNo, err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}

// Get Journal Infomation By document ref godoc
// @Summary		แสดงรายละเอียดข้อมูลรายวัน ตามเอกสารอ้างอิง
// @Description แสดงรายละเอียดข้อมูลรายวัน ตามเอกสารอ้างอิง
// @Tags		GL
// @Param		doc  path      string  true  "Document Ref"
// @Accept 		json
// @Success		200	{object}	models.JournalInfoResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/journal/docref/{doc} [get]
func (h JournalHttp) InfoJournalByDocumentRef(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	docRef := ctx.Param("doc")

	h.ms.Logger.Debugf("Get Journal Document ref %v:", docRef)
	doc, err := h.svc.InfoJournalByDocumentRef(shopID, docRef)

	if err != nil {
		h.ms.Logger.Errorf("Error getting document %v: %v", doc, err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}

// List Journal godoc
// @Summary		แสดงรายการข้อมูลรายวัน
// @Description แสดงรายการข้อมูลรายวัน
// @Tags		GL
// @Param		q		query	string		false  "Search Value"
// @Param		docno		query	string		false  "DocNo"
// @Param		appname		query	string		false  "AppName"
// @Param		docdate		query	string		false  "DocDate ex. 2020-01-01"
// @Param		debtorname		query	string		false  "Debtor Name"
// @Param		accountyear		query	int		false  "Account Year"
// @Param		accountperiod		query	int		false  "Account Period"
// @Param		accountdescription		query	int		false  "Account Description"
// @Param		amount		query	int		false  "Amount"
// @Param		createdby		query	string		false  "Created By"
// @Param		createdat		query	string		false  "Create Date ex. 2020-01-01"
// @Param		startdate  query	string		false  "Start Date ex 2020-01-01"
// @Param		enddate  query	string		false  "End Date ex 2020-01-01"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Size"
// @Accept 		json
// @Success		200	{object}	models.JournalPageResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/journal [get]
func (h JournalHttp) SearchJournal(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	accountGroup := ctx.QueryParam("accountgroup")

	startDateText := strings.TrimSpace(ctx.QueryParam("startdate"))
	endDateText := strings.TrimSpace(ctx.QueryParam("enddate"))

	startDate := time.Time{}
	endDate := time.Time{}

	var err error

	if len(startDateText) > 0 {
		startDate, err = time.Parse("2006-01-02", startDateText)

		if err == nil {
			startDate = startDate.Add(time.Duration(-24) * time.Hour)
		}
	}

	if len(endDateText) > 0 {
		endDate, err = time.Parse("2006-01-02", endDateText)

		if err == nil {
			endDate = endDate.Add(time.Duration(24) * time.Hour)
		}
	}

	// filterFields := []string{"docno", "docdate", "accountyear", "amount"}
	filterFields := []common.SearchFilter{
		{
			Field: "docno",
			Type:  "string",
		},
		{
			Field: "appname",
			Type:  "string",
		},
		{
			Field: "docdate",
			Type:  "time.Time",
		},
		{
			Field: "accountyear",
			Type:  "int16",
		},
		{
			Field: "accountperiod",
			Type:  "int16",
		},
		{
			Field: "accountdescription",
			Type:  "string",
		},
		{
			Field: "amount",
			Type:  "float64",
		},
		{
			Field: "createdby",
			Type:  "string",
		},
		{
			Field: "createdat",
			Type:  "time.Time",
		},
		{
			Field: "debtorname",
			Type:  "string",
		},
		{
			Field: "creditorname",
			Type:  "string",
		},
	}
	searchFilters := map[string]interface{}{}

	for _, searchFilter := range filterFields {
		qVal := strings.TrimSpace(ctx.QueryParam(searchFilter.Field))

		if len(qVal) > 0 {

			switch searchFilter.Type {
			case "string":
				searchFilters[searchFilter.Field] = qVal
			case "int":
				intVal, err := strconv.Atoi(qVal)
				if err == nil {
					searchFilters[searchFilter.Field] = intVal
				}
			case "int16":
				intVal, err := strconv.Atoi(qVal)
				if err == nil {
					searchFilters[searchFilter.Field] = int16(intVal)
				}
			case "float64":
				floatVal, err := strconv.ParseFloat(qVal, 64)
				if err == nil {
					searchFilters[searchFilter.Field] = floatVal
				}
			case "time.Time":
				dateVal, err := time.Parse("2006-01-02", qVal)
				if err == nil {
					searchFilters[searchFilter.Field] = dateVal
				}
			case "bool":
				boolVal, err := strconv.ParseBool(qVal)
				if err == nil {
					searchFilters[searchFilter.Field] = boolVal
				}
			}
		}
	}

	docList, pagination, err := h.svc.SearchJournal(shopID, pageable, searchFilters, startDate, endDate, accountGroup)

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

// Create Journal Bulk godoc
// @Summary		นำเข้าข้อมูลรายวัน
// @Description นำเข้าข้อมูลรายวัน
// @Tags		GL
// @Param		Journal  body      []models.Journal  true  "ข้อมูลรายวัน"
// @Accept 		json
// @Success		201	{object}	common.BulkInsertResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/journal/bulk [post]
func (h JournalHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.Journal{}
	err := json.Unmarshal([]byte(input), &dataReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	bulkResponse, err := h.svc.SaveInBatch(shopID, authUsername, dataReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	ctx.Response(
		http.StatusCreated,
		common.BulkResponse{
			Success:    true,
			BulkImport: bulkResponse,
		},
	)

	return nil
}

// Get Journal Last DocNo godoc
// @Summary		แสดงรายละเอียดข้อมูลรายวัน
// @Description แสดงรายละเอียดข้อมูลรายวัน
// @Tags		GL
// @Accept 		json
// @Success		200	{object}	models.JournalInfoResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/journal/last-docno [get]
func (h JournalHttp) GetLastDocNo(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	docFormat := ctx.QueryParam("docformat")

	doc, err := h.svc.FindLastDocnoFromFormat(shopID, docFormat)

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

// Get Duplicate DocNos godoc
// @Summary		ดึงรายการเลขที่เอกสารที่ซ้ำกัน
// @Description ดึงรายการเลขที่เอกสารที่ซ้ำกันในระบบ พร้อมจำนวนครั้งที่ซ้ำ
// @Tags		GL
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/journal/duplicate-docnos [get]
func (h JournalHttp) GetDuplicateDocNos(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	duplicates, err := h.svc.GetDuplicateDocNos(shopID)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    duplicates,
	})
	return nil
}

// Check VAT DocNo godoc
// @Summary		เช็คเลขที่ภาษีซ้ำ
// @Description เช็คว่าเลขที่ภาษีถูกใช้งานแล้วหรือไม่
// @Tags		GL
// @Param		debttype	query	int		true	"0=เจ้าหนี้, 1=ลูกหนี้"
// @Param		code		query	string	true	"รหัสเจ้าหนี้/ลูกหนี้"
// @Param		docnumber	query	string	true	"เลขที่เอกสารภาษี"
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/journal/check-vatdocno [get]
func (h JournalHttp) CheckVatDocNo(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	debtTypeStr := ctx.QueryParam("debttype")
	code := ctx.QueryParam("code")
	docNumber := ctx.QueryParam("docnumber")

	// Validate parameters
	if len(code) == 0 || len(docNumber) == 0 {
		ctx.ResponseError(http.StatusBadRequest, "code and docnumber are required")
		return nil
	}

	debtType, err := strconv.Atoi(debtTypeStr)
	if err != nil || (debtType != 0 && debtType != 1) {
		ctx.ResponseError(http.StatusBadRequest, "debttype must be 0 or 1")
		return nil
	}

	result, err := h.svc.CheckVatDocNoExists(shopID, debtType, code, docNumber)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    result,
	})
	return nil
}

// Check Tax DocNo godoc
// @Summary		เช็คเลขที่ภาษีหัก ณ ที่จ่ายซ้ำ
// @Description เช็คว่าเลขที่ภาษีหัก ณ ที่จ่ายถูกใช้งานแล้วหรือไม่
// @Tags		GL
// @Param		debttype	query	int		true	"0=เจ้าหนี้, 1=ลูกหนี้"
// @Param		code		query	string	true	"รหัสเจ้าหนี้/ลูกหนี้"
// @Param		docnumber	query	string	true	"เลขที่เอกสารภาษีหัก ณ ที่จ่าย"
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/journal/check-taxdocno [get]
func (h JournalHttp) CheckTaxDocNo(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	debtTypeStr := ctx.QueryParam("debttype")
	code := ctx.QueryParam("code")
	docNumber := ctx.QueryParam("docnumber")

	// Validate parameters
	if len(code) == 0 || len(docNumber) == 0 {
		ctx.ResponseError(http.StatusBadRequest, "code and docnumber are required")
		return nil
	}

	debtType, err := strconv.Atoi(debtTypeStr)
	if err != nil || (debtType != 0 && debtType != 1) {
		ctx.ResponseError(http.StatusBadRequest, "debttype must be 0 or 1")
		return nil
	}

	result, err := h.svc.CheckTaxDocNoExists(shopID, debtType, code, docNumber)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    result,
	})
	return nil
}
