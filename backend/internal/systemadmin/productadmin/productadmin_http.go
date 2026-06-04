package productadmin

import (
	"encoding/json"
	"net/http"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/internal/logger"
	common "smlcloudplatform/internal/models"
	"smlcloudplatform/pkg/microservice"
)

type IProductAdminHttp interface {
	ReSycProductBarcode(ctx microservice.IContext) error
	ReCalcStockBalance(ctx microservice.IContext) error
	RegisterHttp(ms *microservice.Microservice, prefix string)
	DeleteProductBarcode(ctx microservice.IContext) error
}

type ProductAdminHttp struct {
	svc IProductAdminService
}

func NewProductAdminHttp(ms *microservice.Microservice, cfg config.IConfig) IProductAdminHttp {

	producer := microservice.NewProducer(cfg.MQConfig().URI(), cfg.MQConfig().SecurityProtocol(), cfg.MQConfig().SSLCAFile(), cfg.MQConfig().SSLKeyFile(), cfg.MQConfig().SSLCertFile(), ms.Logger)
	mongoPersister := microservice.NewPersisterMongo(cfg.MongoPersisterConfig())

	svc := NewProductAdminService(mongoPersister, producer)

	return &ProductAdminHttp{
		svc: svc,
	}
}

func (s *ProductAdminHttp) RegisterHttp(ms *microservice.Microservice, prefix string) {
	ms.POST(prefix+"/productadmin/resyncproductbarcode", s.ReSycProductBarcode)
	ms.POST(prefix+"/productadmin/deleteproductbarcodeall", s.DeleteProductBarcode)
	ms.POST(prefix+"/productadmin/recalcstock", s.ReCalcStockBalance)
}

func (s *ProductAdminHttp) ReSycProductBarcode(ctx microservice.IContext) error {

	input := ctx.ReadInput()
	var req RequestReSyncProductBarcode

	err := json.Unmarshal([]byte(input), &req)

	if err != nil {
		ctx.Response(http.StatusBadRequest, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	err = s.svc.ReSyncProductBarcode(req.HoldingCode)
	if err != nil {
		logger.GetLogger().Error("ReSycProductBarcode error ", err)
		ctx.Response(http.StatusInternalServerError, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	ctx.Response(http.StatusOK, common.ResponseSuccess{
		Success: true,
	})
	return nil
}

func (s *ProductAdminHttp) ReCalcStockBalance(ctx microservice.IContext) error {

	input := ctx.ReadInput()
	var req RequestReSyncProductBarcode

	err := json.Unmarshal([]byte(input), &req)

	if err != nil {
		ctx.Response(http.StatusBadRequest, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	logger.GetLogger().Debugf("Receive ReCalcStockBalance HoldingCode:%v , Barcode:%v", req.HoldingCode, req.Barcode)
	err = s.svc.ReCalcStockBalance(req.HoldingCode, req.Barcode)
	if err != nil {
		ctx.Response(http.StatusBadRequest, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	ctx.Response(http.StatusOK, common.ResponseSuccess{
		Success: true,
	})
	return nil
}

func (s *ProductAdminHttp) DeleteProductBarcode(ctx microservice.IContext) error {

	input := ctx.ReadInput()
	var req RequestReSyncProductBarcode

	err := json.Unmarshal([]byte(input), &req)

	if err != nil {
		ctx.Response(http.StatusBadRequest, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	err = s.svc.DeleteProductBarcodeAll(req.HoldingCode, ctx.UserInfo().Username)
	if err != nil {
		ctx.Response(http.StatusBadRequest, common.ApiResponse{
			Success: false,
			Message: err.Error(),
		})
		return err
	}

	ctx.Response(http.StatusOK, common.ResponseSuccess{
		Success: true,
	})
	return nil
}
