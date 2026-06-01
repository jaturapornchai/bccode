package company

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"smlcloudplatform/internal/config"
	common "smlcloudplatform/internal/models"
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
	companyTopicCreated = "when-organization-company-created"
	companyTopicUpdated = "when-organization-company-updated"
	companyTopicDeleted = "when-organization-company-deleted"
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
	shopID := ctx.UserInfo().ShopID
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

	req.Code = strings.TrimSpace(req.Code)
	if req.Code == "" {
		ctx.ResponseError(http.StatusBadRequest, "company code is required")
		return errors.New("company code is required")
	}
	if err := ensureCompanyCodeAvailable(mongoCtx, pst, shopID, req.Code, ""); err != nil {
		ctx.ResponseError(http.StatusConflict, err.Error())
		return err
	}

	now := time.Now()
	req.ShopID = shopID
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

	kafkaSync, err := orgEvents.PublishOrOutbox(mongoCtx, pst, h.ms.Producer(h.cfg.MQConfig()), shopID, "organization_company", "created", companyTopicCreated, req.GuidFixed, req)
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

func (h CompanyHttp) SearchCompany(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID

	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())

	var list []companyModels.CompanyDoc
	opts := options.Find().SetSort(bson.D{{Key: "created_at", Value: 1}, {Key: "code", Value: 1}})
	if err := pst.Find(mongoCtx, companyModels.CompanyDoc{}, visibleCompanyFilter(shopID), &list, opts); err != nil {
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
	shopID := ctx.UserInfo().ShopID
	id := ctx.Param("id")

	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())

	var data companyModels.CompanyDoc
	if err := pst.FindOne(mongoCtx, companyModels.CompanyDoc{}, bson.M{"shopid": shopID, "guid_fixed": id, "deleted_at": bson.M{"$exists": false}}, &data); err != nil {
		ctx.ResponseError(http.StatusNotFound, "Company not found")
		return err
	}

	ctx.Response(http.StatusOK, common.ApiResponse{
		Success: true,
		Data:    data,
	})
	return nil
}

func (h CompanyHttp) UpdateCompany(ctx microservice.IContext) error {
	shopID := ctx.UserInfo().ShopID
	authUsername := ctx.UserInfo().Username
	id := ctx.Param("id")
	input := ctx.ReadInput()

	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())

	var existing companyModels.CompanyDoc
	if err := pst.FindOne(mongoCtx, companyModels.CompanyDoc{}, bson.M{"shopid": shopID, "guid_fixed": id, "deleted_at": bson.M{"$exists": false}}, &existing); err != nil {
		ctx.ResponseError(http.StatusNotFound, "Company not found")
		return err
	}

	var req companyModels.CompanyDoc
	if err := json.Unmarshal([]byte(input), &req); err != nil {
		ctx.ResponseError(http.StatusBadRequest, err.Error())
		return err
	}
	req.Code = strings.TrimSpace(req.Code)
	if req.Code == "" {
		ctx.ResponseError(http.StatusBadRequest, "company code is required")
		return errors.New("company code is required")
	}
	if err := ensureCompanyCodeAvailable(mongoCtx, pst, shopID, req.Code, id); err != nil {
		ctx.ResponseError(http.StatusConflict, err.Error())
		return err
	}

	existing.Names = req.Names
	existing.TaxID = req.TaxID
	existing.Code = req.Code
	existing.IsActive = req.IsActive
	existing.UpdatedAt = time.Now()
	existing.UpdatedBy = authUsername

	if err := pst.Update(mongoCtx, companyModels.CompanyDoc{}, bson.M{"shopid": shopID, "guid_fixed": id, "deleted_at": bson.M{"$exists": false}}, bson.M{"$set": bson.M{
		"names":      existing.Names,
		"tax_id":     existing.TaxID,
		"code":       existing.Code,
		"is_active":  existing.IsActive,
		"updated_at": existing.UpdatedAt,
		"updatedby":  existing.UpdatedBy,
	}}); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	kafkaSync, err := orgEvents.PublishOrOutbox(mongoCtx, pst, h.ms.Producer(h.cfg.MQConfig()), shopID, "organization_company", "updated", companyTopicUpdated, id, existing)
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
	shopID := ctx.UserInfo().ShopID
	authUsername := ctx.UserInfo().Username
	id := ctx.Param("id")

	mongoCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	pst := h.ms.MongoPersister(h.cfg.MongoPersisterConfig())

	var data companyModels.CompanyDoc
	if err := pst.FindOne(mongoCtx, companyModels.CompanyDoc{}, bson.M{"shopid": shopID, "guid_fixed": id, "deleted_at": bson.M{"$exists": false}}, &data); err != nil {
		ctx.ResponseError(http.StatusNotFound, "Company not found")
		return err
	}

	now := time.Now()
	data.DeletedAt = &now
	data.DeletedBy = authUsername
	data.UpdatedAt = now
	if err := pst.Update(mongoCtx, companyModels.CompanyDoc{}, bson.M{"shopid": shopID, "guid_fixed": id, "deleted_at": bson.M{"$exists": false}}, bson.M{"$set": bson.M{
		"deleted_at": data.DeletedAt,
		"deletedby":  data.DeletedBy,
		"updated_at": data.UpdatedAt,
	}}); err != nil {
		ctx.ResponseError(http.StatusInternalServerError, err.Error())
		return err
	}

	kafkaSync, err := orgEvents.PublishOrOutbox(mongoCtx, pst, h.ms.Producer(h.cfg.MQConfig()), shopID, "organization_company", "deleted", companyTopicDeleted, id, data)
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

func visibleCompanyFilter(shopID string) bson.M {
	return bson.M{"shopid": shopID, "deleted_at": bson.M{"$exists": false}}
}

func ensureCompanyCodeAvailable(ctx context.Context, pst microservice.IPersisterMongo, shopID string, code string, excludeGuid string) error {
	filter := visibleCompanyFilter(shopID)
	filter["code"] = code
	if excludeGuid != "" {
		filter["guid_fixed"] = bson.M{"$ne": excludeGuid}
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
