package company

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
	branchModels "smlcloudplatform/internal/organization/branch/models"
	companyModels "smlcloudplatform/internal/organization/company/models"
	orgEvents "smlcloudplatform/internal/organization/events"
	"smlcloudplatform/internal/utils"
	"smlcloudplatform/pkg/microservice"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const (
	companyTopicCreated              = "when-organization-company-created"
	companyTopicUpdated              = "when-organization-company-updated"
	companyTopicDeleted              = "when-organization-company-deleted"
	companyDefaultBranchTopicCreated = "when-organization-branch-created"
	companyBranchTopicDeleted        = "when-organization-branch-deleted"
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
	h.ms.DELETE("/organization/company/:id", h.DeleteCompany)
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

	req.Code = companyModels.NormalizeCompanyCode(req.Code)
	if req.Code == "" {
		ctx.ResponseError(http.StatusBadRequest, "company code is required")
		return errors.New("company code is required")
	}
	if err := ensureCompanyCodeAvailable(mongoCtx, pst, holdingCode, req.Code, ""); err != nil {
		ctx.ResponseError(http.StatusConflict, err.Error())
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

	if _, err := pst.Create(mongoCtx, companyModels.CompanyDoc{}, req); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	kafkaSync, err := orgEvents.PublishOrOutbox(mongoCtx, pst, h.ms.Producer(h.cfg.MQConfig()), holdingCode, "organizationcompany", "created", companyTopicCreated, req.GuidFixed, req)
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}
	branchID, branchKafkaSync, err := h.ensureDefaultHeadOfficeBranch(mongoCtx, pst, holdingCode, authUsername, req.GuidFixed, req.Names)
	if err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	ctx.Response(http.StatusCreated, common.ApiResponse{
		Success: true,
		ID:      req.GuidFixed,
		Data: map[string]string{
			"kafka_sync":        kafkaSync,
			"defaultbranchid":   branchID,
			"branch_kafka_sync": branchKafkaSync,
		},
	})
	return nil
}

func (h CompanyHttp) ensureDefaultHeadOfficeBranch(
	ctx context.Context,
	pst microservice.IPersisterMongo,
	holdingCode string,
	authUsername string,
	companyGuid string,
	companyNames common.JSONB,
) (string, string, error) {
	filter := visibleCompanyBranchFilter(holdingCode, companyGuid)
	count, err := pst.Count(ctx, branchModels.BranchOrgDoc{}, filter)
	if err != nil && !errors.Is(err, mongo.ErrNoDocuments) {
		return "", "", err
	}
	if count > 0 {
		return "", "", nil
	}

	now := time.Now()
	branch := branchModels.BranchOrgDoc{
		HoldingCode: holdingCode,
		GuidFixed:   utils.NewGUID(),
		CompanyGuid: companyGuid,
		Code:        branchModels.ThaiHeadOfficeBranchCode,
		Names:       defaultHeadOfficeNames(companyNames),
		IsActive:    true,
		// Default locale for Thai SMEs: Bangkok timezone, Buddhist calendar,
		// ISO date format. The branch owner can change these on the branch screen.
		Timezone:       "Asia/Bangkok",
		TimezoneLabel:  "(GMT+07:00) กรุงเทพฯ",
		TimezoneOffset: "+07:00",
		DateFormat:     "dd/MM/yyyy",
		YearType:       "buddhist",
		Language:       "th",
		BranchType:     "head",
		CreatedAt:      now,
		UpdatedAt:      now,
		CreatedBy:      authUsername,
	}
	if _, err := pst.Create(ctx, branchModels.BranchOrgDoc{}, branch); err != nil {
		return "", "", err
	}

	kafkaSync, err := orgEvents.PublishOrOutbox(ctx, pst, h.ms.Producer(h.cfg.MQConfig()), holdingCode, "organizationbranch", "created", companyDefaultBranchTopicCreated, branch.GuidFixed, branch)
	if err != nil {
		return "", "", err
	}
	return branch.GuidFixed, kafkaSync, nil
}

func visibleCompanyBranchFilter(holdingCode string, companyGuid string) bson.M {
	return bson.M{
		"holdingcode": holdingCode,
		"companyguid": companyGuid,
		"deletedat":   bson.M{"$exists": false},
	}
}

func defaultHeadOfficeNames(companyNames common.JSONB) common.JSONB {
	names := common.JSONB{}
	seen := map[string]bool{}
	for _, item := range companyNames {
		if item.Code == nil {
			continue
		}
		code := strings.ToLower(strings.TrimSpace(*item.Code))
		if code == "" || seen[code] {
			continue
		}
		seen[code] = true
		names = append(names, *common.NewNameXWithCodeName(code, defaultHeadOfficeNameForLanguage(code)))
	}
	if len(names) == 0 {
		names = append(names,
			*common.NewNameXWithCodeName("th", "สำนักงานใหญ่"),
			*common.NewNameXWithCodeName("en", "Head Office"),
		)
	}
	return names
}

func defaultHeadOfficeNameForLanguage(code string) string {
	switch strings.ToLower(strings.TrimSpace(code)) {
	case "en":
		return "Head Office"
	case "lo":
		return "ສຳນັກງານໃຫຍ່"
	case "km":
		return "ការិយាល័យកណ្តាល"
	case "vi":
		return "Trụ sở chính"
	default:
		return "สำนักงานใหญ่"
	}
}

func (h CompanyHttp) SearchCompany(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode

	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())

	var list []companyModels.CompanyDoc
	opts := options.Find().SetSort(bson.D{{Key: "createdat", Value: 1}, {Key: "code", Value: 1}})
	if err := pst.Find(mongoCtx, companyModels.CompanyDoc{}, visibleCompanyFilter(holdingCode), &list, opts); err != nil {
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

	var data companyModels.CompanyDoc
	if err := pst.FindOne(mongoCtx, companyModels.CompanyDoc{}, bson.M{"holdingcode": holdingCode, "guidfixed": id, "deletedat": bson.M{"$exists": false}}, &data); err != nil {
		ctx.ResponseError(http.StatusNotFound, "Company not found")
		return err
	}
	if strings.TrimSpace(data.GuidFixed) == "" {
		ctx.ResponseError(http.StatusNotFound, "Company not found")
		return errors.New("Company not found")
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    data,
	})
	return nil
}

func (h CompanyHttp) UpdateCompany(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	authUsername := ctx.UserInfo().Username
	id := ctx.Param("id")
	input := ctx.ReadInput()

	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())

	var existing companyModels.CompanyDoc
	if err := pst.FindOne(mongoCtx, companyModels.CompanyDoc{}, bson.M{"holdingcode": holdingCode, "guidfixed": id, "deletedat": bson.M{"$exists": false}}, &existing); err != nil {
		ctx.ResponseError(http.StatusNotFound, "Company not found")
		return err
	}
	if strings.TrimSpace(existing.GuidFixed) == "" {
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
	if err := ensureCompanyCodeAvailable(mongoCtx, pst, holdingCode, req.Code, id); err != nil {
		ctx.ResponseError(http.StatusConflict, err.Error())
		return err
	}

	existing.Names = req.Names
	existing.TaxID = req.TaxID
	existing.Code = req.Code
	existing.LogoURI = req.LogoURI
	existing.IsActive = req.IsActive
	existing.UpdatedAt = time.Now()
	existing.UpdatedBy = authUsername

	if err := pst.Update(mongoCtx, companyModels.CompanyDoc{}, bson.M{"holdingcode": holdingCode, "guidfixed": id, "deletedat": bson.M{"$exists": false}}, bson.M{"$set": bson.M{
		"names":     existing.Names,
		"taxid":     existing.TaxID,
		"code":      existing.Code,
		"logouri":   existing.LogoURI,
		"isactive":  existing.IsActive,
		"updatedat": existing.UpdatedAt,
		"updatedby": existing.UpdatedBy,
	}}); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	kafkaSync, err := orgEvents.PublishOrOutbox(mongoCtx, pst, h.ms.Producer(h.cfg.MQConfig()), holdingCode, "organizationcompany", "updated", companyTopicUpdated, id, existing)
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

func (h CompanyHttp) DeleteCompany(ctx microservice.IContext) error {
	holdingCode := ctx.UserInfo().HoldingCode
	id := ctx.Param("id")

	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())

	companyFilter := bson.M{"holdingcode": holdingCode, "guidfixed": id, "deletedat": bson.M{"$exists": false}}
	var data companyModels.CompanyDoc
	if err := pst.FindOne(mongoCtx, companyModels.CompanyDoc{}, companyFilter, &data); err != nil {
		ctx.ResponseError(http.StatusNotFound, "Company not found")
		return err
	}
	if strings.TrimSpace(data.GuidFixed) == "" {
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
	return bson.M{"holdingcode": holdingCode, "deletedat": bson.M{"$exists": false}}
}

func ensureCompanyCodeAvailable(ctx context.Context, pst microservice.IPersisterMongo, holdingCode string, code string, excludeGuid string) error {
	filter := visibleCompanyFilter(holdingCode)
	filter["code"] = code
	if excludeGuid != "" {
		filter["guidfixed"] = bson.M{"$ne": excludeGuid}
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
