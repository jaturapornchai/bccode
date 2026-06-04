package models

import (
	salechannel_models "smlcloudplatform/internal/channel/salechannel/models"
	notify_models "smlcloudplatform/internal/notify/models"
	device_models "smlcloudplatform/internal/order/device/models"
	order_models "smlcloudplatform/internal/order/setting/models"
	branch_models "smlcloudplatform/internal/organization/branch/models"
	"smlcloudplatform/internal/pos/media/models"
	kitchen_models "smlcloudplatform/internal/restaurant/kitchen/models"
)

type EOrderShop struct {
	HoldingCode    string                     `json:"holding_code"`
	Name1          string                     `json:"name1"`
	ProfilePicture string                     `json:"profilepicture"`
	TotalTable     int                        `json:"totaltable"`
	OrderStation   EOrderShopOrderStation     `json:"orderstation,omitempty"`
	Kitchens       []kitchen_models.Kitchen   `json:"kitchens" bson:"kitchens"`
	Notify         []notify_models.NotifyInfo `json:"notify"`
}

type EOrderShopOrderStation struct {
	device_models.OrderDevice
	Setting EOrderSetting `json:"ordersetting"`
}

type EOrderSetting struct {
	order_models.OrderSetting
	Branch       branch_models.Branch             `json:"branch"`
	Media        models.Media                     `json:"media"`
	SaleChannels []salechannel_models.SaleChannel `json:"salechannels" `
}

// previous version
type EOrderShopOld struct {
	HoldingCode    string `json:"holding_code"`
	Name1          string `json:"name1"`
	ProfilePicture string `json:"profilepicture"`
	TotalTable     int    `json:"totaltable"`
	IsBcMember     bool   `json:"isbcmember"`
	ApiKey         string `json:"apikey"`

	OrderStation EOrderShopOrderOld       `json:"orderstation,omitempty"`
	Kitchens     []kitchen_models.Kitchen `json:"kitchens" bson:"kitchens"`
}

type EOrderShopOrderOld struct {
	EOrderSettingOld
	DeviceInfo device_models.OrderDevice `json:"deviceinfo" bson:"deviceinfo"`
}

type EOrderSettingOld struct {
	order_models.OrderSetting
	Branch       branch_models.Branch             `json:"branch"`
	Media        models.Media                     `json:"media"`
	SaleChannels []salechannel_models.SaleChannel `json:"salechannels" `
}
