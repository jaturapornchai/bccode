package chartofaccount

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/vfgl/chartofaccount/models"
	"smlcloudplatform/internal/vfgl/chartofaccount/repositories"
	"smlcloudplatform/internal/vfgl/chartofaccount/services"
	journalRepo "smlcloudplatform/internal/vfgl/journal/repositories"
	"smlcloudplatform/pkg/microservice"
	"strings"
)

type IChartOfAccountHttp interface{}

type ChartOfAccountHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.IChartOfAccountHttpService
}

func NewChartOfAccountHttp(ms *microservice.Microservice, cfg config.IConfig) ChartOfAccountHttp {

	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	prod := ms.Producer(cfg.MQConfig())

	repo := repositories.NewChartOfAccountRepository(pst)
	mqRepo := repositories.NewChartOfAccountMQRepository(prod)

	repoJournal := journalRepo.NewJournalRepository(pst)

	svc := services.NewChartOfAccountHttpService(repo, repoJournal, mqRepo)
	return ChartOfAccountHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h ChartOfAccountHttp) RegisterHttp() {
	h.ms.GET("/gl/chartofaccount", h.Search)
	h.ms.POST("/gl/chartofaccount", h.Create)
	h.ms.GET("/gl/chartofaccount/:id", h.Info)
	h.ms.PUT("/gl/chartofaccount/:id", h.Update)
	h.ms.DELETE("/gl/chartofaccount/:id", h.Delete)
	h.ms.POST("/gl/chartofaccount/bulk", h.SaveBulk)
}

// List Chart Of Account godoc
// @Summary		แสดงรายการผังบัญชี
// @Description แสดงรายการผังบัญชี
// @Tags		GL
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Size"
// @Accept 		json
// @Success		200	{object}	models.ChartOfAccountPageResponse
// @Failure		401 {object}	models.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/chartofaccount [get]
func (h ChartOfAccountHttp) Search(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	accountCodeRangeRaw := strings.TrimSpace(ctx.QueryParam("accountcode"))

	accRanges := []models.AccountCodeRange{}
	if len(accountCodeRangeRaw) > 0 {
		accSplitRaw := strings.Split(accountCodeRangeRaw, ",")
		for _, accRaw := range accSplitRaw {
			if len(accRaw) > 0 {
				accRangeRaw := strings.Split(accRaw, ":")
				if len(accRangeRaw) == 2 {

					accRanges = append(accRanges, models.AccountCodeRange{
						Start: accRangeRaw[0],
						End:   accRangeRaw[1],
					})
				}
			}

		}

	}

	docList, pagination, err := h.svc.Search(shopID, accRanges, pageable)

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

// Create Chart Of Account godoc
// @Summary		สร้างผังบัญชี
// @Description สร้างผังบัญชี
// @Tags		GL
// @Param		ChartOfAccount  body      models.ChartOfAccount  true  "ChartOfAccount"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/chartofaccount [post]
func (h ChartOfAccountHttp) Create(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.ChartOfAccount{}
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

// Get Chart Of Account Infomation godoc
// @Summary		แสดงรายละเอียดผังบัญชี
// @Description แสดงรายละเอียดผังบัญชี
// @Tags		GL
// @Param		id  path      string  true  "Id"
// @Accept 		json
// @Success		200	{object}	models.ChartOfAccountInfoResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/chartofaccount/{id} [get]
func (h ChartOfAccountHttp) Info(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get Journal %v", id)
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

// Update Chart Of Account godoc
// @Summary		แก้ไขผังบัญชี
// @Description แก้ไขผังบัญชี
// @Tags		GL
// @Param		id  path      string  true  "ID"
// @Param		ChartOfAccount  body      models.ChartOfAccount  true  "ChartOfAccount"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/chartofaccount/{id} [put]
func (h ChartOfAccountHttp) Update(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.ChartOfAccount{}
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

// Delete Chart Of Account godoc
// @Summary		ลบผังบัญชี
// @Description ลบผังบัญชี
// @Tags		GL
// @Param		id  path      string  true  "ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/chartofaccount/{id} [delete]
func (h ChartOfAccountHttp) Delete(ctx microservice.IContext) error {
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

// Create Chart Of Account Bulk godoc
// @Summary		นำเข้าผังบัญชี
// @Description นำเข้าผังบัญชี
// @Tags		GL
// @Param		ChartOfAccount  body      []models.ChartOfAccount  true  "ChartOfAccount"
// @Accept 		json
// @Success		201	{object}	common.BulkInsertResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /gl/chartofaccount/bulk [post]
func (h ChartOfAccountHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.ChartOfAccount{}
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
