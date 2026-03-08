package shopadmin

import (
	"encoding/json"
	"fmt"
	"smlcloudplatform/internal/config"
	"smlcloudplatform/pkg/microservice"

	goMicroModels "smlcloudplatform/internal/models"
	shopModels "smlcloudplatform/internal/shop/models"
)

type IShopAdminHttp interface {
	ListShop(ctx microservice.IContext) error
	CreateShop(ctx microservice.IContext) error
	GetShopByProjectNo(ctx microservice.IContext) error
	RegisterHttp(ms *microservice.Microservice, prefix string)
	ListShopUsers(ctx microservice.IContext) error
	ListShopUsersByShopId(ctx microservice.IContext) error
}

type ShopAdminHttp struct {
	ms           *microservice.Microservice
	cfg          config.IConfig
	shopadminsvc IShopAdminService
}

func NewShopAdminHttp(ms *microservice.Microservice, cfg config.IConfig) IShopAdminHttp {

	shopAdminService := NewShopAdminService(microservice.NewPersisterMongo(cfg.MongoPersisterConfig()))
	return &ShopAdminHttp{
		ms:           ms,
		cfg:          cfg,
		shopadminsvc: shopAdminService,
	}
}

func (s *ShopAdminHttp) RegisterHttp(ms *microservice.Microservice, prefix string) {
	ms.GET(prefix+"/shopadmin/listshop", s.ListShop)
	ms.POST(prefix+"/shopadmin/newshop", s.CreateShop)
	ms.GET(prefix+"/shopadmin/getshopbybranchcode", s.GetShopByProjectNo)
	ms.GET(prefix+"/shopadmin/listshopusers", s.ListShopUsers)
	ms.GET(prefix+"/shopadmin/shopusers", s.ListShopUsersByShopId)

}

func (s *ShopAdminHttp) ListShop(ctx microservice.IContext) error {

	shopList, err := s.shopadminsvc.ListShop()

	if err != nil {
		ctx.Response(500, goMicroModels.RestErrorResponse{
			ErrStatus: 500,
			ErrError:  err.Error(),
		})
		return err

	}
	ctx.Response(200, goMicroModels.ApiResponse{
		Success: true,
		Data:    shopList,
	})
	return nil
}

func (s *ShopAdminHttp) CreateShop(ctx microservice.IContext) error {

	input := ctx.ReadInput()
	var doc shopModels.ShopDoc

	err := json.Unmarshal([]byte(input), &doc)

	if err != nil {
		ctx.Response(500, goMicroModels.RestErrorResponse{
			ErrStatus: 500,
			ErrError:  err.Error(),
		})
		return err
	}

	err = s.shopadminsvc.CreateShop(doc)

	if err != nil {
		ctx.Response(500, goMicroModels.RestErrorResponse{
			ErrStatus: 500,
			ErrError:  err.Error(),
		})
		return err
	}

	ctx.Response(200, goMicroModels.ResponseSuccess{
		Success: true,
	})
	return nil
}

func (s *ShopAdminHttp) GetShopByProjectNo(ctx microservice.IContext) error {

	branchCode := ctx.QueryParam("branchcode")

	shop, err := s.shopadminsvc.FindShopByProjectNo(branchCode)

	if err != nil {
		ctx.Response(500, goMicroModels.RestErrorResponse{
			ErrStatus: 500,
			ErrError:  err.Error(),
		})
		return err

	}

	if shop.GuidFixed == "" {
		ctx.Response(200, goMicroModels.ApiResponse{
			Success: false,
			Message: fmt.Sprintf("code %s not found", branchCode),
		})
		return nil
	}

	ctx.Response(200, goMicroModels.ApiResponse{
		Success: true,
		Data:    shop,
	})
	return nil
}

func (s *ShopAdminHttp) ListShopUsers(ctx microservice.IContext) error {

	shopList, err := s.shopadminsvc.ListShopUsersAll()

	if err != nil {
		ctx.Response(500, goMicroModels.RestErrorResponse{
			ErrStatus: 500,
			ErrError:  err.Error(),
		})
		return err

	}
	ctx.Response(200, goMicroModels.ApiResponse{
		Success: true,
		Data:    shopList,
	})
	return nil
}

func (s *ShopAdminHttp) ListShopUsersByShopId(ctx microservice.IContext) error {

	shopId := ctx.QueryParam("shopid")

	shopList, err := s.shopadminsvc.ListShopUsersByShopId(shopId)

	if err != nil {
		ctx.Response(500, goMicroModels.RestErrorResponse{
			ErrStatus: 500,
			ErrError:  err.Error(),
		})
		return err

	}
	ctx.Response(200, goMicroModels.ApiResponse{
		Success: true,
		Data:    shopList,
	})
	return nil
}
