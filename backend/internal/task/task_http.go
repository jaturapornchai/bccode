package task

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	repositoriesDocumentImage "smlcloudplatform/internal/documentwarehouse/documentimage/repositories"
	servicesDocumentImage "smlcloudplatform/internal/documentwarehouse/documentimage/services"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/internal/task/models"
	"smlcloudplatform/internal/task/repositories"
	"smlcloudplatform/internal/task/services"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"strconv"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

type ITaskHttp interface{}

type TaskHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
	svc services.ITaskHttpService
}

func NewTaskHttp(ms *microservice.Microservice, cfg config.IConfig) TaskHttp {
	pst := ms.MongoPersister(cfg.MongoPersisterConfig())
	prod := ms.Producer(cfg.MQConfig())

	repo := repositories.NewTaskRepository(pst)
	repoImage := repositoriesDocumentImage.NewDocumentImageRepository(pst)
	repoImageGroup := repositoriesDocumentImage.NewDocumentImageGroupRepository(pst)
	azureblob := microservice.NewFilePersister()

	repoImageMessagequeue := repositoriesDocumentImage.NewDocumentImageMessageQueueRepository(prod)

	svcImage := servicesDocumentImage.NewDocumentImageService(repoImage, repoImageGroup, repoImageMessagequeue, azureblob)
	svc := services.NewTaskHttpService(repo, repoImageGroup, svcImage)

	return TaskHttp{
		ms:  ms,
		cfg: cfg,
		svc: svc,
	}
}

func (h TaskHttp) RegisterHttp() {

	h.ms.POST("/task/bulk", h.SaveBulk)

	h.ms.GET("/task", h.SearchTaskPage)
	h.ms.GET("/task/list", h.SearchTaskLimit)
	h.ms.POST("/task", h.CreateTask)
	h.ms.GET("/task/generate-code", h.GenerateTaskCode)
	h.ms.GET("/task/:id", h.InfoTask)
	h.ms.GET("/task/reject/:guid", h.GetTaskReject)
	h.ms.PUT("/task/:id", h.UpdateTask)
	h.ms.PUT("/task/:id/status", h.UpdateTaskStatus)
	h.ms.DELETE("/task/:id", h.DeleteTask)
	h.ms.DELETE("/task", h.DeleteTaskByGUIDs)
}

// Create Task godoc
// @Description Create Task
// @Tags		Task
// @Param		Task  body      models.Task  true  "Task"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /task [post]
func (h TaskHttp) CreateTask(ctx microservice.IContext) error {
	authUsername := ctx.UserInfo().Username
	shopID := ctx.UserInfo().ShopID
	input := ctx.ReadInput()

	docReq := &models.Task{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	idx, err := h.svc.CreateTask(shopID, authUsername, *docReq)

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

// Update Task godoc
// @Description Update Task
// @Tags		Task
// @Param		id  path      string  true  "Task ID"
// @Param		Task  body      models.Task  true  "Task"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /task/{id} [put]
func (h TaskHttp) UpdateTask(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.Task{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateTask(shopID, id, authUsername, *docReq)

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

// Update Task Status godoc
// @Description Update Task Status
// @Tags		Task
// @Param		id  path      string  true  "Task ID"
// @Param		TaskStatus  body      models.TaskStatus  true  "Task Status"
// @Accept 		json
// @Success		201	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /task/{id}/status [put]
func (h TaskHttp) UpdateTaskStatus(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	id := ctx.Param("id")
	input := ctx.ReadInput()

	docReq := &models.TaskStatus{}
	err := json.Unmarshal([]byte(input), &docReq)

	if err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	if err = ctx.Validate(docReq); err != nil {
		ctx.ResponseError(400, err.Error())
		return err
	}

	err = h.svc.UpdateTaskStatus(shopID, id, authUsername, docReq.Status)

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

// Delete Task godoc
// @Description Delete Task
// @Tags		Task
// @Param		id  path      string  true  "Task ID"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /task/{id} [delete]
func (h TaskHttp) DeleteTask(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	id := ctx.Param("id")

	err := h.svc.DeleteTask(shopID, id, authUsername)

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

// Delete Task godoc
// @Description Delete Task
// @Tags		Task
// @Param		Task  body      []string  true  "Task GUIDs"
// @Accept 		json
// @Success		200	{object}	common.ResponseSuccessWithID
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /task [delete]
func (h TaskHttp) DeleteTaskByGUIDs(ctx microservice.IContext) error {
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

	err = h.svc.DeleteTaskByGUIDs(shopID, authUsername, docReq)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
	})

	return nil
}

// Generate Task Code godoc
// @Description generate new Task Code
// @Tags		Task
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /task/generate-code [get]
func (h TaskHttp) GenerateTaskCode(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID
	authUsername := userInfo.Username

	newTaskCode, err := h.svc.GenerateTaskID(shopID, authUsername)

	if err != nil {
		h.ms.Logger.Errorf("Error generate task code: %v", err)
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    newTaskCode,
	})
	return nil
}

// Get Task godoc
// @Description get struct array by ID
// @Tags		Task
// @Param		id  path      string  true  "Task ID"
// @Accept 		json
// @Success		200	{object}	common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /task/{id} [get]
func (h TaskHttp) InfoTask(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	id := ctx.Param("id")

	h.ms.Logger.Debugf("Get Task %v", id)
	doc, err := h.svc.InfoTask(shopID, id)

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

// Get Task Reject List godoc
// @Description get Task Reject by task guid
// @Tags		Task
// @Param		module	query	integer		false  "Module"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /task/reject/{guid} [get]
func (h TaskHttp) GetTaskReject(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	taskGUID := ctx.Param("guid")

	module := ctx.QueryParam("module")

	docList, err := h.svc.GetTaskReject(shopID, module, taskGUID)

	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    docList,
	})
	return nil
}

// List Task godoc
// @Description get struct array by ID
// @Tags		Task
// @Param		q		query	string		false  "Search Value"
// @Param		page	query	integer		false  "Page"
// @Param		limit	query	integer		false  "Limit"
// @Param		module	query	integer		false  "Module"
// @Param		jobfilter	query	integer		false  "Job Filter"
// @Param		status	query	integer		false  "ex. status=0 status=1,2,3"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /task [get]
func (h TaskHttp) SearchTaskPage(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageable := utils.GetPageable(ctx.QueryParam)

	module := ctx.QueryParam("module")

	statusReq := ctx.QueryParam("status")
	jobfilter := ctx.QueryParam("jobfilter")

	filters := map[string]interface{}{}

	if len(statusReq) > 0 {
		tempStatusStrArr := strings.Split(statusReq, ",")

		tempStatus := []int{}
		for _, temp := range tempStatusStrArr {
			status, err := strconv.Atoi(temp)
			if err == nil {
				tempStatus = append(tempStatus, status)
			}
		}
		filters["status"] = bson.M{"$in": tempStatus}

		// Check if jobfilter = 1 and status 3 (TaskCompleted) is in the request
		if jobfilter == "1" {
			containsTaskCompleted := false
			for _, status := range tempStatus {
				if status == 3 { // TaskCompleted
					containsTaskCompleted = true
					break
				}
			}

			// If TaskCompleted is requested, exclude tasks with TotalDocumentStatus
			// containing {status:1, total:0}
			if containsTaskCompleted {
				filters["totaldocumentstatus"] = bson.M{
					"$not": bson.M{
						"$elemMatch": bson.M{
							"status": int8(1), // TaskUplaoded
							"total":  0,
						},
					},
				}
			}
		}
	}

	docList, pagination, err := h.svc.SearchTask(shopID, module, filters, pageable)

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

// List Task godoc
// @Description search limit offset
// @Tags		Task
// @Param		q		query	string		false  "Search Value"
// @Param		offset	query	integer		false  "offset"
// @Param		limit	query	integer		false  "limit"
// @Param		lang	query	string		false  "lang"
// @Accept 		json
// @Success		200	{array}		common.ApiResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /task/list [get]
func (h TaskHttp) SearchTaskLimit(ctx microservice.IContext) error {
	userInfo := ctx.UserInfo()
	shopID := userInfo.ShopID

	pageableStep := utils.GetPageableStep(ctx.QueryParam)

	lang := ctx.QueryParam("lang")

	docList, total, err := h.svc.SearchTaskStep(shopID, lang, pageableStep)

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

// Create Task Bulk godoc
// @Description Create Task
// @Tags		Task
// @Param		Task  body      []models.Task  true  "Task"
// @Accept 		json
// @Success		201	{object}	common.BulkResponse
// @Failure		401 {object}	common.AuthResponseFailed
// @Security     AccessToken
// @Router /task/bulk [post]
func (h TaskHttp) SaveBulk(ctx microservice.IContext) error {

	userInfo := ctx.UserInfo()
	authUsername := userInfo.Username
	shopID := userInfo.ShopID

	input := ctx.ReadInput()

	dataReq := []models.Task{}
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
