package branch

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"regexp"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	branchModels "smlcloudplatform/internal/organization/branch/models"
	companyModels "smlcloudplatform/internal/organization/company/models"
	orgEvents "smlcloudplatform/internal/organization/events"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	branchTopicCreated = "when-organization-branch-created"
	branchTopicUpdated = "when-organization-branch-updated"
	branchTopicDeleted = "when-organization-branch-deleted"
)

type BranchHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
}

func NewBranchHttp(ms *microservice.Microservice, cfg config.IConfig) BranchHttp {
	return BranchHttp{
		ms:  ms,
		cfg: cfg,
	}
}

func (h BranchHttp) RegisterHttp() {
	h.ms.POST("/organization/branch", h.CreateBranch)
	h.ms.GET("/organization/branch", h.SearchBranch)
	h.ms.GET("/organization/branch/list", h.SearchBranchStep)
	h.ms.GET("/organization/branch/:id", h.InfoBranch)
	h.ms.PUT("/organization/branch/:id", h.UpdateBranch)
	h.ms.DELETE("/organization/branch/:id", h.DeleteBranch)
}

func (h BranchHttp) CreateBranch(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	authUsername := ctx.UserInfo().Username
	input := ctx.ReadInput()

	var req branchModels.BranchOrgDoc
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())

	if err := prepareBranchCreate(&req, holdingCode, authUsername); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	if err := ensureBranchCompanyExists(mongoCtx, pst, holdingCode, req.CompanyGuid); err != nil {
		responseBranchWriteError(ctx, err)
		return err
	}
	if err := ensureBranchCodeAvailableMongo(mongoCtx, pst, holdingCode, req.CompanyGuid, req.Code, ""); err != nil {
		responseBranchWriteError(ctx, err)
		return err
	}
	if _, err := pst.Create(mongoCtx, branchModels.BranchOrgDoc{}, req); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	kafkaSync, err := orgEvents.PublishOrOutbox(mongoCtx, pst, h.ms.Producer(h.cfg.MQConfig()), holdingCode, "organization_branch", "created", branchTopicCreated, req.GuidFixed, req)
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      req.GuidFixed,
		Data:    map[string]string{"kafka_sync": kafkaSync},
	})
	return nil
}

func (h BranchHttp) SearchBranch(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	companyGuid := ctx.QueryParam("company_guid")

	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())

	filter := visibleBranchFilter(holdingCode)
	if companyGuid != "" {
		filter["company_guid"] = strings.TrimSpace(companyGuid)
	}

	var list []branchModels.BranchOrgDoc
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}, {Key: "code", Value: 1}})
	if err := pst.Find(mongoCtx, branchModels.BranchOrgDoc{}, filter, &list, opts); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    list,
	})
	return nil
}

func (h BranchHttp) InfoBranch(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	id := ctx.Param("id")

	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())

	var data branchModels.BranchOrgDoc
	if err := pst.FindOne(mongoCtx, branchModels.BranchOrgDoc{}, bson.M{"holding_code": holdingCode, "guid_fixed": id, "deleted_at": bson.M{"$exists": false}}, &data); err != nil {
		ctx.ResponseError(http.StatusNotFound, "Branch not found")
		return err
	}
	if strings.TrimSpace(data.GuidFixed) == "" {
		ctx.ResponseError(http.StatusNotFound, "Branch not found")
		return errors.New("Branch not found")
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    data,
	})
	return nil
}

func (h BranchHttp) UpdateBranch(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	authUsername := ctx.UserInfo().Username
	id := ctx.Param("id")
	input := ctx.ReadInput()

	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())

	var existing branchModels.BranchOrgDoc
	if err := pst.FindOne(mongoCtx, branchModels.BranchOrgDoc{}, bson.M{"holding_code": holdingCode, "guid_fixed": id, "deleted_at": bson.M{"$exists": false}}, &existing); err != nil {
		ctx.ResponseError(http.StatusNotFound, "Branch not found")
		return err
	}
	if strings.TrimSpace(existing.GuidFixed) == "" {
		ctx.ResponseError(http.StatusNotFound, "Branch not found")
		return errors.New("Branch not found")
	}

	var req branchModels.BranchOrgDoc
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	if err := prepareBranchUpdate(&req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	if err := ensureBranchCompanyExists(mongoCtx, pst, holdingCode, req.CompanyGuid); err != nil {
		responseBranchWriteError(ctx, err)
		return err
	}
	if err := ensureBranchCodeAvailableMongo(mongoCtx, pst, holdingCode, req.CompanyGuid, req.Code, id); err != nil {
		responseBranchWriteError(ctx, err)
		return err
	}
	existing.Names = req.Names
	existing.Code = req.Code
	existing.CompanyGuid = req.CompanyGuid
	existing.IsActive = req.IsActive
	existing.UpdatedAt = time.Now()
	existing.UpdatedBy = authUsername

	if err := pst.Update(mongoCtx, branchModels.BranchOrgDoc{}, bson.M{"holding_code": holdingCode, "guid_fixed": id, "deleted_at": bson.M{"$exists": false}}, bson.M{"$set": bson.M{
		"names":        existing.Names,
		"code":         existing.Code,
		"company_guid": existing.CompanyGuid,
		"is_active":    existing.IsActive,
		"updated_at":   existing.UpdatedAt,
		"updatedby":    existing.UpdatedBy,
	}}); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	kafkaSync, err := orgEvents.PublishOrOutbox(mongoCtx, pst, h.ms.Producer(h.cfg.MQConfig()), holdingCode, "organization_branch", "updated", branchTopicUpdated, id, existing)
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
		Data:    map[string]string{"kafka_sync": kafkaSync},
	})
	return nil
}

func (h BranchHttp) DeleteBranch(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	id := ctx.Param("id")

	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())

	branchFilter := bson.M{"holding_code": holdingCode, "guid_fixed": id, "deleted_at": bson.M{"$exists": false}}
	var data branchModels.BranchOrgDoc
	if err := pst.FindOne(mongoCtx, branchModels.BranchOrgDoc{}, branchFilter, &data); err != nil {
		ctx.ResponseError(http.StatusNotFound, "Branch not found")
		return err
	}
	if strings.TrimSpace(data.GuidFixed) == "" {
		ctx.ResponseError(http.StatusNotFound, "Branch not found")
		return errors.New("Branch not found")
	}

	if branchModels.IsThaiHeadOfficeBranchCode(data.Code) {
		ctx.ResponseError(http.StatusBadRequest, "head office branch cannot be deleted")
		return errors.New("head office branch cannot be deleted")
	}

	total, err := countCompanyBranches(mongoCtx, pst, holdingCode, data.CompanyGuid)
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}
	if total <= 1 {
		ctx.ResponseError(http.StatusBadRequest, "company must have at least one branch")
		return errors.New("company must have at least one branch")
	}

	if err := pst.Delete(mongoCtx, branchModels.BranchOrgDoc{}, branchFilter); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	kafkaSync, err := orgEvents.PublishOrOutbox(mongoCtx, pst, h.ms.Producer(h.cfg.MQConfig()), holdingCode, "organization_branch", "deleted", branchTopicDeleted, id, data)
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
		Data:    map[string]string{"kafka_sync": kafkaSync},
	})
	return nil
}

func responseBranchWriteError(ctx microservice.IContext, err error) {
	switch {
	case errors.Is(err, branchModels.ErrBranchCompanyGuidRequired):
		ctx.ResponseError(http.StatusBadRequest, err.Error())
	case errors.Is(err, branchModels.ErrBranchCompanyNotFound):
		ctx.ResponseError(http.StatusBadRequest, err.Error())
	case errors.Is(err, branchModels.ErrBranchCodeExists):
		ctx.ResponseError(http.StatusConflict, err.Error())
	default:
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
	}
}

func (h BranchHttp) SearchBranchStep(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	q := ctx.QueryParam("q")
	offsetStr := ctx.QueryParam("offset")
	limitStr := ctx.QueryParam("limit")

	offset := 0
	limit := 100
	if offsetStr != "" {
		if val, err := strconv.Atoi(offsetStr); err == nil {
			offset = val
		}
	}
	if limitStr != "" {
		if val, err := strconv.Atoi(limitStr); err == nil {
			limit = val
		}
	}

	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())

	filter := visibleBranchFilter(holdingCode)
	if q != "" {
		re := primitive.Regex{Pattern: regexp.QuoteMeta(q), Options: "i"}
		filter["$or"] = []bson.M{{"code": re}, {"names.name": re}}
	}

	total, err := pst.Count(mongoCtx, branchModels.BranchOrgDoc{}, filter)
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	var list []branchModels.BranchOrgDoc
	opts := options.Find().SetSkip(int64(offset)).SetLimit(int64(limit)).SetSort(bson.D{{Key: "created_at", Value: 1}, {Key: "code", Value: 1}})
	if err := pst.Find(mongoCtx, branchModels.BranchOrgDoc{}, filter, &list, opts); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    list,
		Total:   total,
	})
	return nil
}

func prepareBranchCreate(req *branchModels.BranchOrgDoc, holdingCode string, authUsername string) error {
	if err := prepareBranchUpdate(req); err != nil {
		return err
	}
	now := time.Now()
	req.HoldingCode = holdingCode
	if req.GuidFixed == "" {
		req.GuidFixed = utils.NewGUID()
	}
	req.CreatedAt = now
	req.UpdatedAt = now
	req.IsActive = true
	req.CreatedBy = authUsername
	return nil
}

func prepareBranchUpdate(req *branchModels.BranchOrgDoc) error {
	req.CompanyGuid = strings.TrimSpace(req.CompanyGuid)
	if req.CompanyGuid == "" {
		return branchModels.ErrBranchCompanyGuidRequired
	}

	normalizedCode, err := branchModels.NormalizeThaiTaxBranchCode(req.Code)
	if err != nil {
		return err
	}
	req.Code = normalizedCode
	return nil
}

func visibleBranchFilter(holdingCode string) bson.M {
	return bson.M{"holding_code": holdingCode, "deleted_at": bson.M{"$exists": false}}
}

func ensureBranchCompanyExists(ctx context.Context, pst microservice.IPersisterMongo, holdingCode string, companyGuid string) error {
	var company companyModels.CompanyDoc
	err := pst.FindOne(ctx, companyModels.CompanyDoc{}, bson.M{"holding_code": holdingCode, "guid_fixed": companyGuid, "deleted_at": bson.M{"$exists": false}}, &company)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return branchModels.ErrBranchCompanyNotFound
	}
	if err != nil {
		return err
	}
	if strings.TrimSpace(company.GuidFixed) == "" {
		return branchModels.ErrBranchCompanyNotFound
	}
	return nil
}

func ensureBranchCodeAvailableMongo(ctx context.Context, pst microservice.IPersisterMongo, holdingCode string, companyGuid string, code string, excludeGuid string) error {
	filter := visibleBranchFilter(holdingCode)
	filter["company_guid"] = companyGuid
	filter["code"] = code
	if excludeGuid != "" {
		filter["guid_fixed"] = bson.M{"$ne": excludeGuid}
	}
	count, err := pst.Count(ctx, branchModels.BranchOrgDoc{}, filter)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}
	if count > 0 {
		return branchModels.ErrBranchCodeExists
	}
	return nil
}

func countCompanyBranches(ctx context.Context, pst microservice.IPersisterMongo, holdingCode string, companyGuid string) (int, error) {
	filter := visibleBranchFilter(holdingCode)
	filter["company_guid"] = companyGuid
	return pst.Count(ctx, branchModels.BranchOrgDoc{}, filter)
}
