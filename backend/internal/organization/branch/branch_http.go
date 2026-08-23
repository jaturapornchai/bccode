package branch

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"regexp"
	authModels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	orgaccess "smlcloudplatform/internal/organization"
	orgpolicy "smlcloudplatform/internal/organization/access"
	branchModels "smlcloudplatform/internal/organization/branch/models"
	companyModels "smlcloudplatform/internal/organization/company/models"
	orgEvents "smlcloudplatform/internal/organization/events"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/apperr"
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
	if authErr := orgaccess.RequireHoldingAdmin(pst, ctx.UserInfo()); authErr != nil {
		return apperr.Respond(ctx, authErr)
	}
	membership, err := orgpolicy.FindActiveHoldingManager(mongoCtx, pst, ctx.UserInfo(), time.Now())
	if err != nil {
		return respondBranchMembershipError(ctx, err)
	}
	holding, err := orgpolicy.FindActiveHolding(mongoCtx, pst, holdingCode)
	if err != nil {
		return respondBranchMembershipError(ctx, err)
	}
	holdingUID := strings.TrimSpace(holding.HoldingUID)
	if holdingUID == "" || strings.TrimSpace(holding.GuidFixed) != holdingUID || strings.TrimSpace(membership.HoldingUID) != holdingUID {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("Holding lineage does not match Membership").WithThaiMessage("Membership ไม่ตรงกับ Holding ที่เลือก"))
	}

	now := time.Now().UTC()
	if err := prepareBranchCreate(&req, holdingCode, holdingUID, authUsername, utils.NewGUID(), now); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	parent, err := findActiveBranchCompany(mongoCtx, pst, req.HoldingCode, req.HoldingUID, req.CompanyUID)
	if err != nil {
		responseBranchWriteError(ctx, err)
		return err
	}
	req.CompanyUID = parent.CompanyUID
	req.CompanyGuid = parent.CompanyUID
	req.BusinessCode = parent.Code
	if err := orgaccess.EnsureOrganizationCreateIndexes(mongoCtx, pst,
		orgaccess.UniqueIndexSpec{Model: branchModels.BranchOrgDoc{}, Name: "uniq_organizationbranches_branchuid", Keys: bson.D{{Key: "branchuid", Value: 1}}, PartialFilter: orgaccess.CanonicalStringFieldsFilter("branchuid")},
		orgaccess.UniqueIndexSpec{Model: branchModels.BranchOrgDoc{}, Name: "uniq_organizationbranches_company_code", Keys: bson.D{{Key: "companyuid", Value: 1}, {Key: "branchcode", Value: 1}}, PartialFilter: orgaccess.CanonicalStringFieldsFilter("companyuid", "branchcode")},
	); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	actorUID := strings.TrimSpace(ctx.UserInfo().UID)
	auditUID := primitive.NewObjectID().Hex()
	err = orgaccess.ApplyOrganizationCreate(mongoCtx, pst, orgaccess.OrganizationCreate{
		TargetModel: branchModels.BranchOrgDoc{},
		Target:      req,
		CodeClaim: orgaccess.OrganizationCodeClaim{
			EntityType: "branch", ScopeUID: req.CompanyUID, NormalizedCode: req.BranchCode,
			EntityUID: req.BranchUID, ClaimedAt: now, ClaimedBy: actorUID,
		},
		Audit: orgaccess.OrganizationAudit{
			AuditUID: auditUID, ActorUID: actorUID, Action: "branch.created",
			TargetType: "branch", TargetUID: req.BranchUID, HoldingUID: req.HoldingUID,
			CompanyUID: req.CompanyUID, BranchUID: req.BranchUID,
			After:      bson.M{"holdinguid": req.HoldingUID, "companyuid": req.CompanyUID, "branchuid": req.BranchUID, "branchcode": req.BranchCode, "isactive": true},
			OccurredAt: now,
		},
		Outbox: orgaccess.OrganizationOutboxEvent{
			EventUID: primitive.NewObjectID().Hex(), AggregateType: "branch", AggregateUID: req.BranchUID,
			Version: 0, EventType: "branch.created",
			Payload: bson.M{"audituid": auditUID, "holdinguid": req.HoldingUID, "companyuid": req.CompanyUID, "branchuid": req.BranchUID, "branchcode": req.BranchCode},
			Status:  "PENDING", Attempts: 0, OccurredAt: now,
		},
	})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return apperr.Respond(ctx, apperr.DuplicateCode("code", req.Code).
				WithMessage("branch code is exists").WithThaiMessage("รหัสสาขานี้มีอยู่แล้วในบริษัทนี้").WithWrap(err))
		}
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      req.BranchUID,
		Data:    orgaccess.OrganizationCreateResponse{Entity: req, KafkaSync: "outbox"},
	})
	return nil
}

func (h BranchHttp) SearchBranch(ctx microservice.IContext) error {
	if strings.EqualFold(strings.TrimSpace(ctx.QueryParam("management")), "true") {
		return h.SearchBranchManagement(ctx)
	}
	holdingCode := ctx.UserInfo().HoldingCode
	companyGuid := ctx.QueryParam("companyguid")

	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())
	membership, err := orgpolicy.FindActiveMembership(mongoCtx, pst, ctx.UserInfo(), time.Now())
	if err != nil {
		return respondBranchMembershipError(ctx, err)
	}
	companies, err := loadScopedBranchCompanies(mongoCtx, pst, holdingCode, membership.AccessScopes)
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
	}
	if len(companies) == 0 {
		ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Data: []branchModels.BranchOrgDoc{}})
		return nil
	}

	filter := scopedBranchFilter(holdingCode, companies, membership.AccessScopes)
	if companyGuid != "" {
		companyGuid = strings.TrimSpace(companyGuid)
		if _, ok := companies[companyGuid]; !ok {
			return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("company access denied").WithThaiMessage("ไม่มีสิทธิ์ใช้งานบริษัทนี้"))
		}
		filter["companyuid"] = companyGuid
	}

	return h.respondBranchList(ctx, mongoCtx, pst, filter)
}

// SearchBranchManagement returns organization structure for Holding managers.
// Workspace reads stay on SearchBranch and continue to enforce transaction scopes.
func (h BranchHttp) SearchBranchManagement(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	companyGuid := strings.TrimSpace(ctx.QueryParam("companyguid"))

	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())
	if _, err := orgpolicy.FindActiveHoldingManager(mongoCtx, pst, ctx.UserInfo(), time.Now()); err != nil {
		return respondBranchMembershipError(ctx, err)
	}
	if err := orgpolicy.RequireActiveHolding(mongoCtx, pst, holdingCode); err != nil {
		return respondBranchMembershipError(ctx, err)
	}

	filter := visibleBranchFilter(holdingCode)
	if companyGuid != "" {
		filter["$or"] = bson.A{bson.M{"companyuid": companyGuid}, bson.M{"companyguid": companyGuid}}
	}
	return h.respondBranchList(ctx, mongoCtx, pst, filter)
}

func (h BranchHttp) respondBranchList(ctx microservice.IContext, mongoCtx context.Context, pst microservice.IPersisterMongo, filter bson.M) error {
	var list []branchModels.BranchOrgDoc
	opts := options.Find().SetSort(bson.D{{Key: "createdat", Value: 1}, {Key: "code", Value: 1}})
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
	membership, err := orgpolicy.FindActiveMembership(mongoCtx, pst, ctx.UserInfo(), time.Now())
	if err != nil {
		return respondBranchMembershipError(ctx, err)
	}
	companies, err := loadScopedBranchCompanies(mongoCtx, pst, holdingCode, membership.AccessScopes)
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
	}
	if len(companies) == 0 {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("branch access denied").WithThaiMessage("ไม่มีสิทธิ์ใช้งานสาขานี้"))
	}

	var data branchModels.BranchOrgDoc
	if err := pst.FindOne(mongoCtx, branchModels.BranchOrgDoc{}, branchIdentityFilter(holdingCode, id), &data); err != nil {
		ctx.ResponseError(http.StatusNotFound, "Branch not found")
		return err
	}
	branchUID := stableOrLegacyBranchUID(data)
	if branchUID == "" {
		ctx.ResponseError(http.StatusNotFound, "Branch not found")
		return errors.New("Branch not found")
	}
	companyUID := stableOrLegacyBranchCompanyUID(data)
	if _, companyAllowed := companies[companyUID]; !companyAllowed || !orgpolicy.AllowsBranch(membership.AccessScopes, companyUID, branchUID) {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("branch access denied").WithThaiMessage("ไม่มีสิทธิ์ใช้งานสาขานี้"))
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
	membership, err := orgpolicy.FindActiveHoldingManager(mongoCtx, pst, ctx.UserInfo(), time.Now())
	if err != nil {
		return respondBranchMembershipError(ctx, err)
	}
	holding, err := orgpolicy.FindActiveHolding(mongoCtx, pst, holdingCode)
	if err != nil {
		return respondBranchMembershipError(ctx, err)
	}
	canonicalHoldingUID := strings.TrimSpace(holding.HoldingUID)
	if canonicalHoldingUID == "" || strings.TrimSpace(holding.GuidFixed) != canonicalHoldingUID || strings.TrimSpace(membership.HoldingUID) != canonicalHoldingUID {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("Holding lineage does not match Membership").WithThaiMessage("Membership ไม่ตรงกับ Holding ที่เลือก"))
	}

	var existing branchModels.BranchOrgDoc
	identityFilter := branchIdentityFilter(holdingCode, id)
	if err := pst.FindOne(mongoCtx, branchModels.BranchOrgDoc{}, identityFilter, &existing); err != nil {
		ctx.ResponseError(http.StatusNotFound, "Branch not found")
		return err
	}
	branchUID := stableOrLegacyBranchUID(existing)
	if branchUID == "" {
		ctx.ResponseError(http.StatusNotFound, "Branch not found")
		return errors.New("Branch not found")
	}

	var req branchModels.BranchOrgDoc
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}

	if err := preserveBranchParent(&req, existing); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	if err := prepareBranchUpdate(&req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	if strings.TrimSpace(existing.HoldingUID) != canonicalHoldingUID {
		return apperr.Respond(ctx, apperr.ErrConflict.WithMessage("stable Branch Holding identity is required").WithThaiMessage("สาขานี้ยังไม่มีรหัส Holding ถาวร"))
	}
	if err := requireUnchangedBranchCode(existing, req.BranchCode); err != nil {
		ctx.ResponseError(http.StatusConflict, err.Error())
		return err
	}
	if _, err := findActiveBranchCompany(mongoCtx, pst, holdingCode, existing.HoldingUID, existing.CompanyUID); err != nil {
		responseBranchWriteError(ctx, err)
		return err
	}
	existing.Names = req.Names
	existing.Code = req.Code
	existing.BranchCode = req.BranchCode
	existing.BusinessTypes = req.BusinessTypes
	existing.LogoURI = req.LogoURI
	existing.Timezone = req.Timezone
	existing.TimezoneLabel = req.TimezoneLabel
	existing.TimezoneOffset = req.TimezoneOffset
	existing.DateFormat = req.DateFormat
	existing.YearType = req.YearType
	existing.BaseCurrency = req.BaseCurrency
	existing.Language = req.Language
	existing.BranchType = req.BranchType
	existing.IsVatRegistered = req.IsVatRegistered
	existing.CompanyRegistrationNo = req.CompanyRegistrationNo
	existing.Email = req.Email
	existing.ManagerName = req.ManagerName
	existing.Addresses = req.Addresses
	existing.CountryCode = req.CountryCode
	existing.ProvinceCode = req.ProvinceCode
	existing.DistrictCode = req.DistrictCode
	existing.SubDistrictCode = req.SubDistrictCode
	existing.ZipCode = req.ZipCode
	existing.FiscalStartMonth = req.FiscalStartMonth
	existing.DocumentFormats = req.DocumentFormats
	existing.ETaxEnabled = req.ETaxEnabled
	requestedStatus, statusChanged, err := orgaccess.ResolveRequestedActiveStatus(input, existing.IsActive)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	existing.IsActive = requestedStatus
	existing.UpdatedAt = time.Now()
	existing.UpdatedBy = authUsername

	set := bson.M{
		"names":                 existing.Names,
		"code":                  existing.Code,
		"branchcode":            existing.BranchCode,
		"businesstypes":         existing.BusinessTypes,
		"logouri":               existing.LogoURI,
		"timezone":              existing.Timezone,
		"timezonelabel":         existing.TimezoneLabel,
		"timezoneoffset":        existing.TimezoneOffset,
		"dateformat":            existing.DateFormat,
		"yeartype":              existing.YearType,
		"basecurrency":          existing.BaseCurrency,
		"language":              existing.Language,
		"branchtype":            existing.BranchType,
		"isvatregistered":       existing.IsVatRegistered,
		"companyregistrationno": existing.CompanyRegistrationNo,
		"email":                 existing.Email,
		"managername":           existing.ManagerName,
		"addresses":             existing.Addresses,
		"countrycode":           existing.CountryCode,
		"provincecode":          existing.ProvinceCode,
		"districtcode":          existing.DistrictCode,
		"subdistrictcode":       existing.SubDistrictCode,
		"zipcode":               existing.ZipCode,
		"fiscalstartmonth":      existing.FiscalStartMonth,
		"documentformats":       existing.DocumentFormats,
		"etaxenabled":           existing.ETaxEnabled,
		"updatedat":             existing.UpdatedAt,
		"updatedby":             existing.UpdatedBy,
	}
	kafkaSync := ""
	if statusChanged {
		reason, err := orgaccess.StatusChangeReason(input)
		if err != nil {
			ctx.ResponseError(http.StatusBadRequest, err.Error())
			return err
		}
		holdingUID := strings.TrimSpace(membership.HoldingUID)
		if holdingUID == "" {
			ctx.ResponseError(http.StatusConflict, "stable Holding identity is required")
			return errors.New("stable Holding identity is required")
		}
		err = orgaccess.ApplyStatusChange(mongoCtx, pst, orgaccess.StatusChange{
			TargetModel:      branchModels.BranchOrgDoc{},
			TargetFilter:     identityFilter,
			MembershipFilter: orgaccess.BranchMembershipStatusFilter(holdingCode, existing.CompanyUID, branchUID),
			TargetType:       "branch",
			TargetUID:        branchUID,
			HoldingUID:       holdingUID,
			CompanyUID:       existing.CompanyUID,
			BranchUID:        branchUID,
			ActorUID:         ctx.UserInfo().UID,
			Reason:           reason,
			Before:           !requestedStatus,
			After:            requestedStatus,
			Version:          existing.Version,
			Set:              set,
			OccurredAt:       existing.UpdatedAt,
		})
		kafkaSync = "outbox"
	} else {
		err = orgaccess.ApplyMetadataUpdate(
			mongoCtx,
			pst,
			branchModels.BranchOrgDoc{},
			identityFilter,
			existing.Version,
			existing.IsActive,
			set,
		)
		if err == nil {
			existing.Version++
			kafkaSync, err = orgEvents.PublishOrOutbox(mongoCtx, pst, h.ms.Producer(h.cfg.MQConfig()), holdingCode, "organizationbranch", "updated", branchTopicUpdated, branchUID, existing)
		}
	}
	if err != nil {
		if errors.Is(err, orgaccess.ErrStatusChangeConflict) {
			ctx.ResponseError(http.StatusConflict, err.Error())
			return err
		}
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
	if _, err := orgpolicy.FindActiveHoldingManager(mongoCtx, pst, ctx.UserInfo(), time.Now()); err != nil {
		return respondBranchMembershipError(ctx, err)
	}
	if err := orgpolicy.RequireActiveHolding(mongoCtx, pst, holdingCode); err != nil {
		return respondBranchMembershipError(ctx, err)
	}

	branchFilter := bson.M{"holdingcode": holdingCode, "branchuid": id, "isdeleted": bson.M{"$ne": true}, "deletedat": bson.M{"$exists": false}}
	var data branchModels.BranchOrgDoc
	if err := pst.FindOne(mongoCtx, branchModels.BranchOrgDoc{}, branchFilter, &data); err != nil {
		ctx.ResponseError(http.StatusNotFound, "Branch not found")
		return err
	}
	if strings.TrimSpace(data.BranchUID) == "" {
		ctx.ResponseError(http.StatusNotFound, "Branch not found")
		return errors.New("Branch not found")
	}
	if err := ensureBranchCompanyExists(mongoCtx, pst, holdingCode, data.HoldingUID, data.CompanyUID); err != nil {
		responseBranchWriteError(ctx, err)
		return err
	}

	if branchModels.IsThaiHeadOfficeBranchCode(data.Code) {
		ctx.ResponseError(http.StatusBadRequest, "head office branch cannot be deleted")
		return errors.New("head office branch cannot be deleted")
	}

	total, err := countCompanyBranches(mongoCtx, pst, holdingCode, data.CompanyUID)
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

	kafkaSync, err := orgEvents.PublishOrOutbox(mongoCtx, pst, h.ms.Producer(h.cfg.MQConfig()), holdingCode, "organizationbranch", "deleted", branchTopicDeleted, id, data)
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
	membership, err := orgpolicy.FindActiveMembership(mongoCtx, pst, ctx.UserInfo(), time.Now())
	if err != nil {
		return respondBranchMembershipError(ctx, err)
	}
	companies, err := loadScopedBranchCompanies(mongoCtx, pst, holdingCode, membership.AccessScopes)
	if err != nil {
		return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
	}
	if len(companies) == 0 {
		ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Data: []branchModels.BranchOrgDoc{}, Total: 0})
		return nil
	}

	filter := scopedBranchFilter(holdingCode, companies, membership.AccessScopes)
	if q != "" {
		re := primitive.Regex{Pattern: regexp.QuoteMeta(q), Options: "i"}
		filter["$and"] = bson.A{bson.M{"$or": bson.A{bson.M{"code": re}, bson.M{"names.name": re}}}}
	}

	total, err := pst.Count(mongoCtx, branchModels.BranchOrgDoc{}, filter)
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	var list []branchModels.BranchOrgDoc
	opts := options.Find().SetSkip(int64(offset)).SetLimit(int64(limit)).SetSort(bson.D{{Key: "createdat", Value: 1}, {Key: "code", Value: 1}})
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

var errBranchParentChange = errors.New("branch parent cannot be changed")
var errBranchCodeChange = errors.New("branch code cannot be changed by ordinary update")

func prepareBranchCreate(req *branchModels.BranchOrgDoc, holdingCode, holdingUID, authUsername, branchUID string, now time.Time) error {
	if err := prepareBranchUpdate(req); err != nil {
		return err
	}
	holdingCode = strings.TrimSpace(holdingCode)
	holdingUID = strings.TrimSpace(holdingUID)
	branchUID = strings.TrimSpace(branchUID)
	if holdingCode == "" || holdingUID == "" || branchUID == "" {
		return errors.New("stable Branch identity is required")
	}
	now = now.UTC()
	req.ID = primitive.NewObjectID()
	req.HoldingCode = holdingCode
	req.HoldingUID = holdingUID
	req.GuidFixed = branchUID
	req.BranchUID = branchUID
	req.BranchCode = req.Code
	if req.BusinessTypes == nil {
		req.BusinessTypes = []string{}
	}
	req.IsDeleted = false
	req.Version = 0
	req.CreatedAt = now
	req.UpdatedAt = now
	req.IsActive = true
	req.DeletedAt = nil
	req.CreatedBy = strings.TrimSpace(authUsername)
	req.UpdatedBy = ""
	req.DeletedBy = ""
	return nil
}

func prepareBranchUpdate(req *branchModels.BranchOrgDoc) error {
	companyUID := strings.TrimSpace(req.CompanyUID)
	companyGuid := strings.TrimSpace(req.CompanyGuid)
	if companyUID != "" && companyGuid != "" && companyUID != companyGuid {
		return errBranchParentChange
	}
	if companyUID == "" {
		companyUID = companyGuid
	}
	if companyUID == "" {
		return branchModels.ErrBranchCompanyGuidRequired
	}
	req.CompanyUID = companyUID
	req.CompanyGuid = companyUID

	inputCode := strings.TrimSpace(req.Code)
	inputBranchCode := strings.TrimSpace(req.BranchCode)
	if inputCode != "" && inputBranchCode != "" && inputCode != inputBranchCode {
		return errors.New("branch code aliases do not match")
	}
	if inputCode == "" {
		inputCode = inputBranchCode
	}
	normalizedCode, err := branchModels.NormalizeThaiTaxBranchCode(inputCode)
	if err != nil {
		return err
	}
	req.Code = normalizedCode
	req.BranchCode = normalizedCode
	if req.BusinessTypes == nil {
		req.BusinessTypes = []string{}
	}
	normalizedTimezone, err := branchModels.NormalizeIANATimezone(req.Timezone)
	if err != nil {
		return err
	}
	req.Timezone = normalizedTimezone
	req.Addresses = sanitizeBranchAddresses(req.Addresses)
	req.CountryCode = strings.TrimSpace(req.CountryCode)
	req.ProvinceCode = strings.TrimSpace(req.ProvinceCode)
	req.DistrictCode = strings.TrimSpace(req.DistrictCode)
	req.SubDistrictCode = strings.TrimSpace(req.SubDistrictCode)
	req.ZipCode = strings.TrimSpace(req.ZipCode)
	if err := normalizeAndValidateDocFormats(req); err != nil {
		return err
	}
	return validateBranchNames(req.Names)
}

func preserveBranchParent(req *branchModels.BranchOrgDoc, existing branchModels.BranchOrgDoc) error {
	existingHoldingUID := strings.TrimSpace(existing.HoldingUID)
	existingUID := strings.TrimSpace(existing.CompanyUID)
	existingBranchUID := strings.TrimSpace(existing.BranchUID)
	if existingHoldingUID == "" || existingUID == "" || existingBranchUID == "" {
		return errors.New("stable Branch parent identity is required")
	}
	for _, identity := range []struct{ requested, existing string }{
		{req.HoldingUID, existingHoldingUID},
		{req.HoldingCode, strings.TrimSpace(existing.HoldingCode)},
		{req.BranchUID, existingBranchUID},
		{req.GuidFixed, existingBranchUID},
	} {
		if value := strings.TrimSpace(identity.requested); value != "" && value != identity.existing {
			return errBranchParentChange
		}
	}
	requestedUID := strings.TrimSpace(req.CompanyUID)
	requestedGuid := strings.TrimSpace(req.CompanyGuid)
	if requestedUID != "" && requestedGuid != "" && requestedUID != requestedGuid {
		return errBranchParentChange
	}
	if requestedUID == "" {
		requestedUID = requestedGuid
	}
	if requestedUID != "" && requestedUID != existingUID {
		return errBranchParentChange
	}
	req.CompanyUID = existingUID
	req.CompanyGuid = existingUID
	req.HoldingUID = existingHoldingUID
	req.HoldingCode = strings.TrimSpace(existing.HoldingCode)
	req.BranchUID = existingBranchUID
	req.GuidFixed = existingBranchUID
	return nil
}

func validateBranchNames(names common.JSONB) error {
	for _, name := range names {
		if name.Code != nil && name.Name != nil && strings.TrimSpace(*name.Code) != "" && strings.TrimSpace(*name.Name) != "" {
			return nil
		}
	}
	return errors.New("branch name is required")
}

func storedBranchCode(branch branchModels.BranchOrgDoc) string {
	value := strings.TrimSpace(branch.BranchCode)
	if value == "" {
		value = strings.TrimSpace(branch.Code)
	}
	normalized, err := branchModels.NormalizeThaiTaxBranchCode(value)
	if err != nil {
		return value
	}
	return normalized
}

func requireUnchangedBranchCode(existing branchModels.BranchOrgDoc, requestedCode string) error {
	normalized, err := branchModels.NormalizeThaiTaxBranchCode(requestedCode)
	if err != nil {
		return err
	}
	if storedBranchCode(existing) != normalized {
		return errBranchCodeChange
	}
	return nil
}

var docFormatPrefixAllowed = regexp.MustCompile(`[^A-Z0-9]+`)

// normalizeAndValidateDocFormats uppercases/trims each document-number format's
// DocType and Prefix (Prefix keeps only A-Z0-9), then validates that every
// entry has a non-empty DocType and Prefix and that no Prefix repeats anywhere
// in the branch (across all doctypes/formats). Validation runs on every entry
// sent, regardless of Enabled.
func normalizeAndValidateDocFormats(req *branchModels.BranchOrgDoc) error {
	seen := make(map[string]struct{}, len(req.DocumentFormats))
	for i := range req.DocumentFormats {
		f := &req.DocumentFormats[i]
		f.DocType = strings.ToUpper(strings.TrimSpace(f.DocType))
		f.Prefix = strings.TrimSpace(docFormatPrefixAllowed.ReplaceAllString(strings.ToUpper(f.Prefix), ""))
		if f.DocType == "" {
			return fmt.Errorf("ประเภทเอกสาร (doctype) ของรูปแบบเลขที่เอกสารห้ามว่าง")
		}
		if f.Prefix == "" {
			return fmt.Errorf("คำนำหน้าเลขที่เอกสารห้ามว่าง (doctype: %s)", f.DocType)
		}
		if _, dup := seen[f.Prefix]; dup {
			return fmt.Errorf("คำนำหน้าเลขที่เอกสารซ้ำ: %s", f.Prefix)
		}
		seen[f.Prefix] = struct{}{}
	}
	return nil
}

// sanitizeBranchAddresses trims each per-language address, caps its length, and
// drops entries with an empty code or address. Branch handlers don't run
// ctx.Validate, so this is the defense-in-depth guard against oversized/garbage
// direct-API payloads. Length is capped by rune count to avoid splitting a
// multibyte (Thai) character mid-address.
func sanitizeBranchAddresses(in []branchModels.BranchAddress) []branchModels.BranchAddress {
	const maxAddressRunes = 1000
	out := make([]branchModels.BranchAddress, 0, len(in))
	for _, a := range in {
		code := strings.TrimSpace(a.Code)
		addr := strings.TrimSpace(a.Address)
		if code == "" || addr == "" {
			continue
		}
		if r := []rune(addr); len(r) > maxAddressRunes {
			addr = string(r[:maxAddressRunes])
		}
		out = append(out, branchModels.BranchAddress{Code: code, Address: addr})
	}
	return out
}

func visibleBranchFilter(holdingCode string) bson.M {
	return bson.M{"holdingcode": holdingCode, "isdeleted": bson.M{"$ne": true}, "deletedat": bson.M{"$exists": false}}
}

func loadScopedBranchCompanies(ctx context.Context, pst microservice.IPersisterMongo, holdingCode string, scopes []authModels.AccessScope) (map[string]string, error) {
	companyUIDs := orgpolicy.AllowedCompanyUIDs(scopes)
	if len(companyUIDs) == 0 {
		return map[string]string{}, nil
	}

	var companies []companyModels.CompanyDoc
	filter := bson.M{
		"holdingcode": holdingCode,
		"companyuid":  bson.M{"$in": companyUIDs},
		"isdeleted":   bson.M{"$ne": true},
		"deletedat":   bson.M{"$exists": false},
	}
	if err := pst.Find(ctx, companyModels.CompanyDoc{}, filter, &companies, options.Find()); err != nil {
		return nil, err
	}

	codeByGuid := make(map[string]string, len(companies))
	for _, company := range companies {
		if companyUID := strings.TrimSpace(company.CompanyUID); companyUID != "" {
			codeByGuid[companyUID] = companyUID
		}
	}
	return codeByGuid, nil
}

func scopedBranchFilter(holdingCode string, companies map[string]string, scopes []authModels.AccessScope) bson.M {
	filter := visibleBranchFilter(holdingCode)
	allowed := bson.A{}
	for companyUID := range companies {
		if orgpolicy.AllowsAllBranches(scopes, companyUID) {
			allowed = append(allowed, bson.M{"companyuid": companyUID})
			continue
		}
		if branchUIDs := orgpolicy.AllowedBranchUIDs(scopes, companyUID); len(branchUIDs) > 0 {
			allowed = append(allowed, bson.M{"companyuid": companyUID, "branchuid": bson.M{"$in": branchUIDs}})
		}
	}
	if len(allowed) == 0 {
		filter["branchuid"] = bson.M{"$in": []string{}}
	} else {
		filter["$or"] = allowed
	}
	return filter
}

func respondBranchMembershipError(ctx microservice.IContext, err error) error {
	if errors.Is(err, orgpolicy.ErrActiveMembershipRequired) || errors.Is(err, orgpolicy.ErrHoldingManagerRequired) {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("active Holding membership is required").WithThaiMessage("ไม่มี Membership ที่ใช้งานได้ใน Holding นี้"))
	}
	if errors.Is(err, orgpolicy.ErrActiveHoldingRequired) {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("active Holding is required").WithThaiMessage("Holding นี้ไม่ได้เปิดใช้งาน"))
	}
	return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
}

func findActiveBranchCompany(ctx context.Context, pst orgpolicy.MembershipFinder, holdingCode, holdingUID, companyUID string) (companyModels.CompanyDoc, error) {
	holdingCode = strings.TrimSpace(holdingCode)
	holdingUID = strings.TrimSpace(holdingUID)
	companyUID = strings.TrimSpace(companyUID)
	if holdingCode == "" || holdingUID == "" || companyUID == "" {
		return companyModels.CompanyDoc{}, branchModels.ErrBranchCompanyNotFound
	}

	var company companyModels.CompanyDoc
	err := pst.FindOne(ctx, companyModels.CompanyDoc{}, bson.M{
		"holdingcode": holdingCode,
		"holdinguid":  holdingUID,
		"companyuid":  companyUID,
		"isactive":    true,
		"isdeleted":   bson.M{"$ne": true},
		"deletedat":   bson.M{"$exists": false},
	}, &company)
	if errors.Is(err, mongo.ErrNoDocuments) {
		return companyModels.CompanyDoc{}, branchModels.ErrBranchCompanyNotFound
	}
	if err != nil {
		return companyModels.CompanyDoc{}, err
	}
	if strings.TrimSpace(company.CompanyUID) != companyUID || strings.TrimSpace(company.HoldingUID) != holdingUID || company.IsDeleted || !company.IsActive || company.DeletedAt != nil {
		return companyModels.CompanyDoc{}, branchModels.ErrBranchCompanyNotFound
	}
	return company, nil
}

func ensureBranchCompanyExists(ctx context.Context, pst orgpolicy.MembershipFinder, holdingCode, holdingUID, companyUID string) error {
	_, err := findActiveBranchCompany(ctx, pst, holdingCode, holdingUID, companyUID)
	return err
}

func ensureBranchCodeAvailableMongo(ctx context.Context, pst microservice.IPersisterMongo, holdingCode string, companyUID string, code string, excludeUID string) error {
	filter := visibleBranchFilter(holdingCode)
	filter["companyuid"] = companyUID
	filter["branchcode"] = code
	if excludeUID != "" {
		filter["branchuid"] = bson.M{"$ne": excludeUID}
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

func countCompanyBranches(ctx context.Context, pst microservice.IPersisterMongo, holdingCode string, companyUID string) (int, error) {
	filter := visibleBranchFilter(holdingCode)
	filter["companyuid"] = companyUID
	return pst.Count(ctx, branchModels.BranchOrgDoc{}, filter)
}

func stableOrLegacyBranchUID(branch branchModels.BranchOrgDoc) string {
	if uid := strings.TrimSpace(branch.BranchUID); uid != "" {
		return uid
	}
	return strings.TrimSpace(branch.GuidFixed)
}

func stableOrLegacyBranchCompanyUID(branch branchModels.BranchOrgDoc) string {
	if uid := strings.TrimSpace(branch.CompanyUID); uid != "" {
		return uid
	}
	return strings.TrimSpace(branch.CompanyGuid)
}

func branchIdentityFilter(holdingCode, id string) bson.M {
	id = strings.TrimSpace(id)
	return bson.M{
		"holdingcode": strings.TrimSpace(holdingCode),
		"$or": bson.A{
			bson.M{"branchuid": id},
			bson.M{"guidfixed": id},
		},
		"isdeleted": bson.M{"$ne": true},
		"deletedat": bson.M{"$exists": false},
	}
}
