package services

import (
	"context"
	salechannel_models "smlcloudplatform/internal/channel/salechannel/models"
	salechannel_repo "smlcloudplatform/internal/channel/salechannel/repositories"
	notify_repositories "smlcloudplatform/internal/notify/repositories"
	order_device_repo "smlcloudplatform/internal/order/device/repositories"
	order_setting_repo "smlcloudplatform/internal/order/setting/repositories"
	branch_repo "smlcloudplatform/internal/organization/branch/repositories"
	media_repo "smlcloudplatform/internal/pos/media/repositories"
	"smlcloudplatform/internal/product/eorder/models"
	"smlcloudplatform/internal/restaurant/kitchen"
	"smlcloudplatform/internal/restaurant/table"
	"smlcloudplatform/internal/shop"
	"time"
)

type EOrderService struct {
	shopRepo        shop.IShopRepository
	tableRepo       table.ITableRepository
	repoOrder       order_setting_repo.ISettingRepository
	repoMedia       media_repo.IMediaRepository
	repoKitchen     kitchen.IKitchenRepository
	repoDevice      order_device_repo.IDeviceRepository
	repoSaleChannel salechannel_repo.ISaleChannelRepository
	repoBranch      branch_repo.IBranchRepository
	repoNotify      notify_repositories.INotifyRepository
	contextTimeout  time.Duration
}

func NewEOrderService(
	shopRepo shop.IShopRepository,
	tableRepo table.ITableRepository,
	repoOrder order_setting_repo.ISettingRepository,
	repoMedia media_repo.IMediaRepository,
	repoKitchen kitchen.IKitchenRepository,
	repoDevice order_device_repo.IDeviceRepository,
	repoSaleChannel salechannel_repo.ISaleChannelRepository,
	repoBranch branch_repo.IBranchRepository,
	repoNotify notify_repositories.INotifyRepository,
) EOrderService {
	contextTimeout := time.Duration(15) * time.Second
	return EOrderService{
		shopRepo:        shopRepo,
		tableRepo:       tableRepo,
		repoOrder:       repoOrder,
		repoMedia:       repoMedia,
		repoKitchen:     repoKitchen,
		repoDevice:      repoDevice,
		repoSaleChannel: repoSaleChannel,
		repoBranch:      repoBranch,
		repoNotify:      repoNotify,
		contextTimeout:  contextTimeout,
	}
}

func (svc EOrderService) getContextTimeout() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), svc.contextTimeout)
}

func (svc EOrderService) GetShopInfoOld(shopID string, orderStationCode string) (models.EOrderShopOld, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	result := models.EOrderShopOld{}

	shopInfo, err := svc.shopRepo.FindByGuid(ctx, shopID)

	if err != nil {
		return models.EOrderShopOld{}, err
	}

	tableCount, err := svc.tableRepo.Count(ctx, shopID)

	if err != nil {
		return models.EOrderShopOld{}, err
	}

	if orderStationCode != "" {
		orderDevice, err := svc.repoDevice.FindByDocIndentityGuid(ctx, shopID, "code", orderStationCode)

		if err != nil {
			return models.EOrderShopOld{}, err
		}

		tempOrderStation := models.EOrderShopOrderOld{}
		if orderDevice.Code != "" {
			order, err := svc.repoOrder.FindByDocIndentityGuid(ctx, shopID, "guidfixed", orderDevice.SettingCode)

			if err != nil {
				return models.EOrderShopOld{}, err
			}

			tempOrderStation.OrderSetting = order.OrderSetting

			if order.Code != "" {
				// Media
				media, err := svc.repoMedia.FindByGuid(ctx, shopID, order.MediaGUID)

				if err != nil {
					return models.EOrderShopOld{}, err
				}

				tempOrderStation.Media = media.Media

				// Device info
				tempOrderStation.DeviceInfo = orderDevice.OrderDevice

				saleChannelGUIDs := []string{}
				if order.SaleChannels != nil {
					saleChannelGUIDs = *order.SaleChannels
				}

				// Sale channel
				saleChannels, err := svc.repoSaleChannel.FindByGuids(ctx, shopID, saleChannelGUIDs)

				if err != nil {
					return models.EOrderShopOld{}, err
				}

				tempOrderStation.SaleChannels = []salechannel_models.SaleChannel{}
				for _, tempSaleChannel := range saleChannels {
					tempOrderStation.SaleChannels = append(tempOrderStation.SaleChannels, tempSaleChannel.SaleChannel)
				}

				// Branch
				branch, err := svc.repoBranch.FindByGuid(ctx, shopID, order.Branch.GuidFixed)

				if err != nil {
					return models.EOrderShopOld{}, err
				}
				tempOrderStation.Branch = branch.Branch

				// Kitchen
				kitchens, err := svc.repoKitchen.Find(ctx, shopID, map[string]interface{}{
					"groupnumber": order.KitchenGroupNumber,
				})

				if err != nil {
					return models.EOrderShopOld{}, err
				}

				for _, tempKitchen := range kitchens {
					result.Kitchens = append(result.Kitchens, tempKitchen.Kitchen)
				}

			}

			result.OrderStation = tempOrderStation
		}

	}

	result.ShopID = shopInfo.GuidFixed
	result.Name1 = shopInfo.Name1
	result.ProfilePicture = shopInfo.ProfilePicture
	result.IsBcMember = shopInfo.IsBcMember
	result.ApiKey = shopInfo.ApiKey
	result.TotalTable = tableCount

	return result, nil
}

func (svc EOrderService) GetShopInfo(shopID string, orderStationCode string) (models.EOrderShop, error) {
	ctx, ctxCancel := svc.getContextTimeout()
	defer ctxCancel()

	result := models.EOrderShop{}

	shopInfo, err := svc.shopRepo.FindByGuid(ctx, shopID)

	if err != nil {
		return models.EOrderShop{}, err
	}

	tableCount, err := svc.tableRepo.Count(ctx, shopID)

	if err != nil {
		return models.EOrderShop{}, err
	}

	notify, err := svc.repoNotify.Find(ctx, shopID, map[string]interface{}{})
	if err != nil {
		return models.EOrderShop{}, err
	}

	result.Notify = notify

	if orderStationCode != "" {
		orderDevice, err := svc.repoDevice.FindByDocIndentityGuid(ctx, shopID, "code", orderStationCode)

		if err != nil {
			return models.EOrderShop{}, err
		}

		tempOrderStation := models.EOrderShopOrderStation{}
		if orderDevice.Code != "" {
			tempOrderStation.OrderDevice = orderDevice.OrderDevice

			order, err := svc.repoOrder.FindByDocIndentityGuid(ctx, shopID, "guidfixed", orderDevice.SettingCode)

			if err != nil {
				return models.EOrderShop{}, err
			}

			if order.Code != "" {
				// Order setting
				tempOrderStation.Setting.OrderSetting = order.OrderSetting

				// Media
				media, err := svc.repoMedia.FindByGuid(ctx, shopID, order.MediaGUID)

				if err != nil {
					return models.EOrderShop{}, err
				}

				tempOrderStation.Setting.Media = media.Media

				saleChannelGUIDs := []string{}
				if order.SaleChannels != nil {
					saleChannelGUIDs = *order.SaleChannels
				}

				// Sale channel
				saleChannels, err := svc.repoSaleChannel.FindByGuids(ctx, shopID, saleChannelGUIDs)

				if err != nil {
					return models.EOrderShop{}, err
				}

				tempOrderStation.Setting.SaleChannels = []salechannel_models.SaleChannel{}
				for _, tempSaleChannel := range saleChannels {
					tempOrderStation.Setting.SaleChannels = append(tempOrderStation.Setting.SaleChannels, tempSaleChannel.SaleChannel)
				}

				// Branch
				branch, err := svc.repoBranch.FindByGuid(ctx, shopID, order.Branch.GuidFixed)

				if err != nil {
					return models.EOrderShop{}, err
				}
				tempOrderStation.Setting.Branch = branch.Branch

				// Kitchen
				kitchens, err := svc.repoKitchen.Find(ctx, shopID, map[string]interface{}{
					"groupnumber": order.KitchenGroupNumber,
				})

				if err != nil {
					return models.EOrderShop{}, err
				}

				for _, tempKitchen := range kitchens {
					result.Kitchens = append(result.Kitchens, tempKitchen.Kitchen)
				}

			}

			result.OrderStation = tempOrderStation
		}

	}

	result.ShopID = shopInfo.GuidFixed
	result.Name1 = shopInfo.Name1
	result.ProfilePicture = shopInfo.ProfilePicture
	result.TotalTable = tableCount

	return result, nil
}
