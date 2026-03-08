package journalbook

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/vfgl/journalbook/models"
	"smlcloudplatform/internal/vfgl/journalbook/repositories"
	"smlcloudplatform/internal/vfgl/journalbook/services"
	"smlcloudplatform/pkg/microservice"

	common "smlcloudplatform/internal/models"
)

type JournalBookHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IJournalBookHttpService
}

func NewJournalBookHttp(ms *microservice.Microservice, cfg config.IConfig) JournalBookHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	prod := ms.Producer(cfg.MQConfig())

	mongoRepo := repositories.NewJournalBookMongoRepository(pst)
	mqRepo := repositories.NewJournalBookMqRepository(prod)
	svc := services.NewJournalBookHttpService(mongoRepo, mqRepo)

	return JournalBookHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h JournalBookHttp) RegisterHttp() {

	h.ms.POST("/gl/journalbook/bulk", h.SaveBulk)

	h.ms.GET("/gl/journalbook", h.SearchJournalBook)
	h.ms.POST("/gl/journalbook", h.CreateJournalBook)
	h.ms.GET("/gl/journalbook/:id", h.InfoJournalBook)
	h.ms.PUT("/gl/journalbook/:id", h.UpdateJournalBook)
	h.ms.DELETE("/gl/journalbook/:id", h.DeleteJournalBook)
}

// Create Journal Book godoc
// @Summary		สร้างสมุดรายวัน
// @Description สร้างสมุดรายวัน
// @Tags		GL
// @Param		JournalBook  body      models.JournalBook  true  "สมุดรายวัน"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/journalbook [post]
func (h JournalBookHttp) CreateJournalBook(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.JournalBook{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.Create(shopID, authUsername, *docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      idx,
	})
	return nil
}

// Update Journal Book godoc
// @Summary		แก้ไขสมุดรายวัน
// @Description แก้ไขสมุดรายวัน
// @Tags		GL
// @Param		id  path      string  true  "ID"
// @Param		Journal  body      models.JournalBook  true  "สมุดรายวัน"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/journalbook/{id} [put]
func (h JournalBookHttp) UpdateJournalBook(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.JournalBook{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.Update(id, shopID, authUsername, *docReq)

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

// Delete Journal Book godoc
// @Summary		ลบสมุดรายวัน
// @Description ลบสมุดรายวัน
// @Tags		GL
// @Param		id  path      string  true  "ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/journalbook/{id} [delete]
func (h JournalBookHttp) DeleteJournalBook(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.Delete(id, shopID, authUsername)

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

// Get Journal Book Infomation godoc
// @Summary		แสดงรายละเอียดสมุดรายวัน
// @Description แสดงรายละเอียดสมุดรายวัน
// @Tags		GL
// @Param		id  path      string  true  "Id"
// @Accept 		json
// @Success		200	{object}	models.JournalBookInfoResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/journalbook/{id} [get]
func (h JournalBookHttp) InfoJournalBook(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get doc %v", id)
	doc, err := h.svc.Info(id, shopID)

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

// List Journal Book godoc
// @Summary		แสดงรายการสมุดรายวัน
// @Description แสดงรายการสมุดรายวัน
// @Tags		GL
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Size"
// @Accept 		json
// @Success		200	{object}	models.JournalBookPageResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/journalbook [get]
func (h JournalBookHttp) SearchJournalBook(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)
	docList, pagination, err := h.svc.Search(shopID, pageable)

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

// Create Journal Book Bulk godoc
// @Summary		นำเข้าสมุดรายวัน
// @Description นำเข้าสมุดรายวัน
// @Tags		GL
// @Param		JournalBook  body      []models.JournalBook  true  "สมุดรายวัน"
// @Accept 		json
// @Success		201	{object}	common.BulkInsertResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/journalbook/bulk [post]
func (h JournalBookHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.JournalBook{}
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
