package company

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
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
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	companyTopicUpdated       = "when-organization-company-updated"
	companyTopicDeleted       = "when-organization-company-deleted"
	companyBranchTopicDeleted = "when-organization-branch-deleted"
)

type CompanyHttp struct {
	ms  *microservice.Microservice
	cfg config.IConfig
}

func NewCompanyHttp(ms *microservice.Microservice, cfg config.IConfig) CompanyHttp {
	return CompanyHttp{
		ms:  ms,
		cfg: cfg,
	}
}

func (h CompanyHttp) RegisterHttp() {
	h.ms.POST("/organization/company", h.CreateCompany)
	h.ms.GET("/organization/company", h.SearchCompany)
	h.ms.GET("/organization/company/:id", h.InfoCompany)
	h.ms.PUT("/organization/company/:id", h.UpdateCompany)
}

func (h CompanyHttp) CreateCompany(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	authUsername := ctx.UserInfo().Username
	input := ctx.ReadInput()

	var req companyModels.CompanyDoc
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
		return respondCompanyMembershipError(ctx, err)
	}
	holding, err := orgpolicy.FindActiveHolding(mongoCtx, pst, holdingCode)
	if err != nil {
		return respondCompanyMembershipError(ctx, err)
	}
	holdingUID := strings.TrimSpace(holding.HoldingUID)
	if holdingUID == "" || strings.TrimSpace(holding.GuidFixed) != holdingUID || strings.TrimSpace(membership.HoldingUID) != holdingUID {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("Holding lineage does not match Membership").WithThaiMessage("Membership ไม่ตรงกับ Holding ที่เลือก"))
	}

	actorUID := strings.TrimSpace(ctx.UserInfo().UID)
	now := time.Now().UTC()
	if err := prepareCompanyCreate(&req, holdingCode, holdingUID, authUsername, utils.NewGUID(), now); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	if err := orgaccess.EnsureOrganizationCreateIndexes(mongoCtx, pst,
		orgaccess.UniqueIndexSpec{Model: companyModels.CompanyDoc{}, Name: "uniq_organizationcompanies_companyuid", Keys: bson.D{{Key: "companyuid", Value: 1}}, PartialFilter: orgaccess.CanonicalStringFieldsFilter("companyuid")},
		orgaccess.UniqueIndexSpec{Model: companyModels.CompanyDoc{}, Name: "uniq_organizationcompanies_holding_code", Keys: bson.D{{Key: "holdinguid", Value: 1}, {Key: "code", Value: 1}}, PartialFilter: orgaccess.CanonicalStringFieldsFilter("holdinguid", "code")},
	); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	auditUID := primitive.NewObjectID().Hex()
	err = orgaccess.ApplyOrganizationCreate(mongoCtx, pst, orgaccess.OrganizationCreate{
		TargetModel: companyModels.CompanyDoc{},
		Target:      req,
		CodeClaim: orgaccess.OrganizationCodeClaim{
			EntityType: "company", ScopeUID: req.HoldingUID, NormalizedCode: req.Code,
			EntityUID: req.CompanyUID, ClaimedAt: now, ClaimedBy: actorUID,
		},
		Audit: orgaccess.OrganizationAudit{
			AuditUID: auditUID, ActorUID: actorUID, Action: "company.created",
			TargetType: "company", TargetUID: req.CompanyUID, HoldingUID: req.HoldingUID, CompanyUID: req.CompanyUID,
			After:      bson.M{"holdinguid": req.HoldingUID, "companyuid": req.CompanyUID, "code": req.Code, "isactive": true},
			OccurredAt: now,
		},
		Outbox: orgaccess.OrganizationOutboxEvent{
			EventUID: primitive.NewObjectID().Hex(), AggregateType: "company", AggregateUID: req.CompanyUID,
			Version: 0, EventType: "company.created",
			Payload: bson.M{"audituid": auditUID, "holdinguid": req.HoldingUID, "companyuid": req.CompanyUID, "code": req.Code},
			Status:  "PENDING", Attempts: 0, OccurredAt: now,
		},
	})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return apperr.Respond(ctx, apperr.DuplicateCode("code", req.Code).
				WithMessage("company code is exists").WithThaiMessage("รหัสบริษัทนี้มีอยู่แล้วใน Holding").WithWrap(err))
		}
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}
	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      req.CompanyUID,
		Data:    orgaccess.OrganizationCreateResponse{Entity: req, KafkaSync: "outbox"},
	})
	return nil
}

func visibleCompanyBranchFilter(holdingCode string, companyGuid string) bson.M {
	return bson.M{
		"holdingcode": holdingCode,
		"companyguid": companyGuid,
		"isdeleted":   bson.M{"$ne": true},
		"deletedat":   bson.M{"$exists": false},
	}
}

func (h CompanyHttp) SearchCompany(ctx microservice.IContext) error {
	if strings.EqualFold(strings.TrimSpace(ctx.QueryParam("management")), "true") {
		return h.SearchCompanyManagement(ctx)
	}
	holdingCode := ctx.UserInfo().HoldingCode

	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())
	membership, err := orgpolicy.FindActiveMembership(mongoCtx, pst, ctx.UserInfo(), time.Now())
	if err != nil {
		return respondCompanyMembershipError(ctx, err)
	}
	allowedUIDs := orgpolicy.AllowedCompanyUIDs(membership.AccessScopes)
	if len(allowedUIDs) == 0 {
		ctx.Response(http.StatusOK, common.ApiResponse{Success: true, Data: []companyModels.CompanyDoc{}})
		return nil
	}

	filter := visibleCompanyFilter(holdingCode)
	filter["companyuid"] = bson.M{"$in": allowedUIDs}
	return h.respondCompanyList(ctx, mongoCtx, pst, filter)
}

// SearchCompanyManagement returns organization structure for Holding managers.
// It is intentionally separate from SearchCompany, whose result remains limited
// by transaction scopes used by the workspace selector.
func (h CompanyHttp) SearchCompanyManagement(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode

	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())
	if _, err := orgpolicy.FindActiveHoldingManager(mongoCtx, pst, ctx.UserInfo(), time.Now()); err != nil {
		return respondCompanyMembershipError(ctx, err)
	}
	if err := orgpolicy.RequireActiveHolding(mongoCtx, pst, holdingCode); err != nil {
		return respondCompanyMembershipError(ctx, err)
	}

	return h.respondCompanyList(ctx, mongoCtx, pst, visibleCompanyFilter(holdingCode))
}

func (h CompanyHttp) respondCompanyList(ctx microservice.IContext, mongoCtx context.Context, pst microservice.IPersisterMongo, filter bson.M) error {
	var list []companyModels.CompanyDoc
	opts := options.Find().SetSort(bson.D{{Key: "createdat", Value: 1}, {Key: "code", Value: 1}})
	if err := pst.Find(mongoCtx, companyModels.CompanyDoc{}, filter, &list, opts); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    list,
	})
	return nil
}

func (h CompanyHttp) InfoCompany(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	id := ctx.Param("id")

	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())
	membership, err := orgpolicy.FindActiveMembership(mongoCtx, pst, ctx.UserInfo(), time.Now())
	if err != nil {
		return respondCompanyMembershipError(ctx, err)
	}
	if len(orgpolicy.AllowedCompanyUIDs(membership.AccessScopes)) == 0 {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("company access denied").WithThaiMessage("ไม่มีสิทธิ์ใช้งานบริษัทนี้"))
	}

	var data companyModels.CompanyDoc
	if err := pst.FindOne(mongoCtx, companyModels.CompanyDoc{}, companyIdentityFilter(holdingCode, id), &data); err != nil {
		ctx.ResponseError(http.StatusNotFound, "Company not found")
		return err
	}
	companyUID := stableOrLegacyCompanyUID(data)
	if companyUID == "" {
		ctx.ResponseError(http.StatusNotFound, "Company not found")
		return errors.New("Company not found")
	}
	if !orgpolicy.AllowsCompany(membership.AccessScopes, companyUID) {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("company access denied").WithThaiMessage("ไม่มีสิทธิ์ใช้งานบริษัทนี้"))
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    data,
	})
	return nil
}

func respondCompanyMembershipError(ctx microservice.IContext, err error) error {
	if errors.Is(err, orgpolicy.ErrActiveMembershipRequired) || errors.Is(err, orgpolicy.ErrHoldingManagerRequired) {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("active Holding membership is required").WithThaiMessage("ไม่มี Membership ที่ใช้งานได้ใน Holding นี้"))
	}
	if errors.Is(err, orgpolicy.ErrActiveHoldingRequired) {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("active Holding is required").WithThaiMessage("Holding นี้ไม่ได้เปิดใช้งาน"))
	}
	return apperr.Respond(ctx, apperr.ErrInternal.WithWrap(err))
}

func (h CompanyHttp) UpdateCompany(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	authUsername := ctx.UserInfo().Username
	id := ctx.Param("id")
	input := ctx.ReadInput()

	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())
	membership, err := orgpolicy.FindActiveHoldingManager(mongoCtx, pst, ctx.UserInfo(), time.Now())
	if err != nil {
		return respondCompanyMembershipError(ctx, err)
	}
	holding, err := orgpolicy.FindActiveHolding(mongoCtx, pst, holdingCode)
	if err != nil {
		return respondCompanyMembershipError(ctx, err)
	}
	canonicalHoldingUID := strings.TrimSpace(holding.HoldingUID)
	if canonicalHoldingUID == "" || strings.TrimSpace(holding.GuidFixed) != canonicalHoldingUID || strings.TrimSpace(membership.HoldingUID) != canonicalHoldingUID {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("Holding lineage does not match Membership").WithThaiMessage("Membership ไม่ตรงกับ Holding ที่เลือก"))
	}

	var existing companyModels.CompanyDoc
	identityFilter := companyIdentityFilter(holdingCode, id)
	if err := pst.FindOne(mongoCtx, companyModels.CompanyDoc{}, identityFilter, &existing); err != nil {
		ctx.ResponseError(http.StatusNotFound, "Company not found")
		return err
	}
	companyUID := stableOrLegacyCompanyUID(existing)
	if companyUID == "" {
		ctx.ResponseError(http.StatusNotFound, "Company not found")
		return errors.New("Company not found")
	}

	var req companyModels.CompanyDoc
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	req.Code = companyModels.NormalizeCompanyCode(req.Code)
	if req.Code == "" {
		ctx.ResponseError(http.StatusBadRequest, "company code is required")
		return errors.New("company code is required")
	}
	if err := validateCompanyNames(req.Names); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	if err := requireUnchangedCompanyCode(existing.Code, req.Code); err != nil {
		ctx.ResponseError(http.StatusConflict, err.Error())
		return err
	}
	if existing.HoldingUID != "" && strings.TrimSpace(existing.HoldingUID) != canonicalHoldingUID {
		return apperr.Respond(ctx, apperr.ErrForbidden.WithMessage("Company lineage does not match Holding").WithThaiMessage("บริษัทไม่อยู่ใน Holding ที่เลือก"))
	}
	existing.Names = req.Names
	existing.TaxID = req.TaxID
	existing.Code = req.Code
	existing.LogoURI = req.LogoURI
	requestedStatus, statusChanged, err := orgaccess.ResolveRequestedActiveStatus(input, existing.IsActive)
	if err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	existing.IsActive = requestedStatus
	existing.UpdatedAt = time.Now()
	existing.UpdatedBy = authUsername

	set := bson.M{
		"names":     existing.Names,
		"taxid":     existing.TaxID,
		"code":      existing.Code,
		"logouri":   existing.LogoURI,
		"updatedat": existing.UpdatedAt,
		"updatedby": existing.UpdatedBy,
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
			TargetModel:      companyModels.CompanyDoc{},
			TargetFilter:     identityFilter,
			MembershipFilter: orgaccess.CompanyMembershipStatusFilter(holdingCode, companyUID),
			TargetType:       "company",
			TargetUID:        companyUID,
			HoldingUID:       holdingUID,
			CompanyUID:       companyUID,
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
			companyModels.CompanyDoc{},
			identityFilter,
			existing.Version,
			existing.IsActive,
			set,
		)
		if err == nil {
			existing.Version++
			kafkaSync, err = orgEvents.PublishOrOutbox(mongoCtx, pst, h.ms.Producer(h.cfg.MQConfig()), holdingCode, "organizationcompany", "updated", companyTopicUpdated, companyUID, existing)
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

func (h CompanyHttp) DeleteCompany(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	id := ctx.Param("id")

	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())
	if _, err := orgpolicy.FindActiveHoldingManager(mongoCtx, pst, ctx.UserInfo(), time.Now()); err != nil {
		return respondCompanyMembershipError(ctx, err)
	}
	if err := orgpolicy.RequireActiveHolding(mongoCtx, pst, holdingCode); err != nil {
		return respondCompanyMembershipError(ctx, err)
	}

	companyFilter := companyIdentityFilter(holdingCode, id)
	var data companyModels.CompanyDoc
	if err := pst.FindOne(mongoCtx, companyModels.CompanyDoc{}, companyFilter, &data); err != nil {
		ctx.ResponseError(http.StatusNotFound, "Company not found")
		return err
	}
	if strings.TrimSpace(data.CompanyUID) == "" {
		ctx.ResponseError(http.StatusNotFound, "Company not found")
		return errors.New("Company not found")
	}

	branchFilter := visibleCompanyBranchFilter(holdingCode, id)
	var branches []branchModels.BranchOrgDoc
	if err := pst.Find(mongoCtx, branchModels.BranchOrgDoc{}, branchFilter, &branches); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}
	if err := pst.Delete(mongoCtx, branchModels.BranchOrgDoc{}, branchFilter); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}
	if err := pst.Delete(mongoCtx, companyModels.CompanyDoc{}, companyFilter); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	lastBranchKafkaSync := ""
	for _, branch := range branches {
		branchKafkaSync, err := orgEvents.PublishOrOutbox(mongoCtx, pst, h.ms.Producer(h.cfg.MQConfig()), holdingCode, "organizationbranch", "deleted", companyBranchTopicDeleted, branch.GuidFixed, branch)
		if err != nil {
			ctx.ResponseError(http.StatusInternalServerError, err.Error())
			return err
		}
		lastBranchKafkaSync = branchKafkaSync
	}

	kafkaSync, err := orgEvents.PublishOrOutbox(mongoCtx, pst, h.ms.Producer(h.cfg.MQConfig()), holdingCode, "organizationcompany", "deleted", companyTopicDeleted, id, data)
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		ID:      id,
		Data: map[string]string{
			"kafka_sync":        kafkaSync,
			"branch_kafka_sync": lastBranchKafkaSync,
		},
	})
	return nil
}

func visibleCompanyFilter(holdingCode string) bson.M {
	return bson.M{"holdingcode": holdingCode, "isdeleted": bson.M{"$ne": true}, "deletedat": bson.M{"$exists": false}}
}

func prepareCompanyCreate(req *companyModels.CompanyDoc, holdingCode, holdingUID, actor, companyUID string, now time.Time) error {
	req.Code = companyModels.NormalizeCompanyCode(req.Code)
	if req.Code == "" {
		return errors.New("company code is required")
	}
	if err := validateCompanyNames(req.Names); err != nil {
		return err
	}
	holdingCode = strings.TrimSpace(holdingCode)
	holdingUID = strings.TrimSpace(holdingUID)
	companyUID = strings.TrimSpace(companyUID)
	if holdingCode == "" || holdingUID == "" || companyUID == "" {
		return errors.New("stable Company parent identity is required")
	}
	now = now.UTC()
	req.ID = primitive.NewObjectID()
	req.HoldingCode = holdingCode
	req.HoldingUID = holdingUID
	req.GuidFixed = companyUID
	req.CompanyUID = companyUID
	req.IsDeleted = false
	req.Version = 0
	req.IsActive = true
	req.CreatedAt = now
	req.UpdatedAt = now
	req.DeletedAt = nil
	req.CreatedBy = strings.TrimSpace(actor)
	req.UpdatedBy = ""
	req.DeletedBy = ""
	return nil
}

func validateCompanyNames(names common.JSONB) error {
	for _, name := range names {
		if name.Code != nil && name.Name != nil && strings.TrimSpace(*name.Code) != "" && strings.TrimSpace(*name.Name) != "" {
			return nil
		}
	}
	return errors.New("company name is required")
}

var errCompanyCodeChange = errors.New("company code cannot be changed by ordinary update")

func requireUnchangedCompanyCode(existingCode, requestedCode string) error {
	if companyModels.NormalizeCompanyCode(existingCode) != companyModels.NormalizeCompanyCode(requestedCode) {
		return errCompanyCodeChange
	}
	return nil
}

func stableOrLegacyCompanyUID(company companyModels.CompanyDoc) string {
	if uid := strings.TrimSpace(company.CompanyUID); uid != "" {
		return uid
	}
	return strings.TrimSpace(company.GuidFixed)
}

func companyIdentityFilter(holdingCode, id string) bson.M {
	id = strings.TrimSpace(id)
	return bson.M{
		"holdingcode": strings.TrimSpace(holdingCode),
		"$or": bson.A{
			bson.M{"companyuid": id},
			bson.M{"guidfixed": id},
		},
		"isdeleted": bson.M{"$ne": true},
		"deletedat": bson.M{"$exists": false},
	}
}

func ensureCompanyCodeAvailable(ctx context.Context, pst microservice.IPersisterMongo, holdingCode string, code string, excludeGuid string) error {
	filter := visibleCompanyFilter(holdingCode)
	filter["code"] = code
	if excludeGuid != "" {
		filter["companyuid"] = bson.M{"$ne": excludeGuid}
	}
	count, err := pst.Count(ctx, companyModels.CompanyDoc{}, filter)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return err
	}
	if count > 0 {
		return errors.New("company code is exists")
	}
	return nil
}
