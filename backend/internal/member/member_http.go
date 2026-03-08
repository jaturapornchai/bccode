package member

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	mastersync "smlcloudplatform/internal/mastersync/repositories"
	"smlcloudplatform/internal/member/models"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/shop"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/internal/utils/requestfilter"
	"smlcloudplatform/pkg/microservice"
	"time"
)

type IMemberHttp interface {
	RegisterHttp()
	CreateMember(ctx microservice.IContext) error
	UpdateMember(ctx microservice.IContext) error
	DeleteMember(ctx microservice.IContext) error
	InfoMember(ctx microservice.IContext) error
	SearchMember(ctx microservice.IContext) error
}

type MemberHttp struct {
	ms      *microservice.Microservice
	cfg     config.IConfig
	service IMemberService
}

func NewMemberHttp(ms *microservice.Microservice, cfg config.IConfig) MemberHttp {

	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	pstPg := ms.Persister(cfg.PersisterConfig())
	cache := ms.Cacher(cfg.CacherConfig())

	memberRepo := NewMemberRepository(pst)
	memberPgRepo := NewMemberPGRepository(pstPg)
	masterSyncCacheRepo := mastersync.NewMasterSyncCacheRepository(cache)

	shopRepo := shop.NewShopRepository(pst)
	shopUserRepo := shop.NewShopUserRepository(pst)
	shopService := shop.NewShopService(shopRepo, shopUserRepo, utils.NewGUID, ms.TimeNow)

	authService := microservice.NewAuthServicePrefix("linemember:", "linememberrefresh:", ms.Cacher(cfg.CacherConfig()), 24*3*time.Hour, 24*30*time.Hour)

	service := NewMemberService(memberRepo, memberPgRepo, shopService, authService, masterSyncCacheRepo)

	return MemberHttp{
		ms:      ms,
		cfg:     cfg,
		service: service,
	}
}

func (h MemberHttp) RegisterLineHttp() {
	h.ms.POST("/member/line", h.MemberAuthLine)
	h.ms.GET("/member/profile", h.LineProfileInfo)
	h.ms.PUT("/member/profile", h.UpdateMemberProfileWithLine)
}

func (h MemberHttp) RegisterHttp() {

	h.ms.GET("/member/:id", h.InfoMember)
	h.ms.POST("/member", h.CreateMember)
	h.ms.PUT("/member/:id", h.UpdateMember)

	h.ms.GET("/member", h.SearchMemberPage)
	h.ms.GET("/member/list", h.SearchMemberStep)
}

// Auth Line Member godoc
// @Description Auth Line Member
// @Tags		MemberLine
// @Param		LineAuthRequest  body      models.LineAuthRequest  true  "Line Auth Request"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /member/line [post]
func (h MemberHttp) MemberAuthLine(ctx microservice.IContext) error {

	input := ctx.ReadInput()

	payload := models.LineAuthRequest{}
	err := json.Unmarshal([]byte(input), &payload)

	if err != nil {
		ctx.ResponseError(400, "payload invalid")
		return err
	}

	idx, err := h.service.AuthWithLine(payload)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      idx,
	})

	return nil
}

// Update Line Member godoc
// @Description Updat Line Member
// @Tags		MemberLine
// @Param		Member  body      models.Member  true  "Member"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /member/profile [put]
func (h MemberHttp) UpdateMemberProfileWithLine(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	lineUID := userInfo.Username

	input := ctx.ReadInput()

	docReq := models.Member{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.service.UpdateProfileWithLine(shopID, lineUID, docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})
	return nil
}

// Get Member Line Profile Infomation godoc
// @Description Get Member Line Profile
// @Tags		MemberLine
// @Accept 		json
// @Success		200	{object}	models.MemberInfoResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /member/profile [get]
func (h MemberHttp) LineProfileInfo(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	lineUID := userInfo.Username

	doc, err := h.service.LineProfileInfo(shopID, lineUID)

	if err != nil && err.Error() != "mongo: no documents in result" {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if len(doc.GuidFixed) == 0 {
		ctx.Response(http.StatusNotFound, common.ApiResponse{
			Success: false,
			Message: "document not found",
		})
		return nil
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}

// Create Member godoc
// @Description Create Member
// @Tags		Member
// @Param		Member  body      models.Member  true  "Member"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /member [post]
func (h MemberHttp) CreateMember(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	doc := models.Member{}
	err := json.Unmarshal([]byte(input), &doc)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.service.Create(shopID, authUsername, doc)

	if err != nil {
		ctx.ResponseError(400, err.Error())
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      idx,
	})

	return nil
}

// Update Member godoc
// @Description Update Member
// @Tags		Member
// @Param		id  path      string  true  "Member ID"
// @Param		Member  body      models.Member  true  "Member"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /member/{id} [put]
func (h MemberHttp) UpdateMember(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	guid := ctx.Param("id")

	input := ctx.ReadInput()

	docReq := &models.Member{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.service.Update(shopID, authUsername, guid, *docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})
	return nil
}

// Get Member Infomation godoc
// @Description Get Member
// @Tags		Member
// @Param		id  path      string  true  "Member Id"
// @Accept 		json
// @Success		200	{object}	models.MemberInfoResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /member/{id} [get]
func (h MemberHttp) InfoMember(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	doc, err := h.service.Info(shopID, id)

	if err != nil && err.Error() != "mongo: no documents in result" {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if len(doc.GuidFixed) == 0 {
		ctx.Response(http.StatusNotFound, common.ApiResponse{
			Success: false,
			Message: "document not found",
		})
		return nil
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    doc,
	})
	return nil
}

// List Member step godoc
// @Description get list step
// @Tags		Member
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /member [get]
func (h MemberHttp) SearchMemberPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{})

	docList, pagination, err := h.service.SearchMemberInfo(shopID, filters, pageable)

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

// List Member godoc
// @Description search limit offset
// @Tags		Member
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /member/list [get]
func (h MemberHttp) SearchMemberStep(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	filters := requestfilter.GenerateFilters(ctx.QueryParam, []requestfilter.FilterRequest{})

	docList, total, err := h.service.SearchMemberStep(shopID, lang, filters, pageableStep)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    docList,
		Total:   total,
	})
	return nil
}
