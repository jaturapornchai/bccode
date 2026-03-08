package models

import (
	"smlcloudplatform/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

const shopcouponCollectionName = "shopCoupon"

type ShopCoupon struct {
	models.PartitionIdentity `bson:"inline"`
	CouponType               int16  `json:"coupontype" bson:"coupontype"`
	Logo                     string `json:"logo" bson:"logo"`
	models.Name              `bson:"inline"`
}

type ShopCouponInfo struct {
	models.DocIdentity `bson:"inline"`
	ShopCoupon         `bson:"inline"`
}

func (ShopCouponInfo) CollectionName() string {
	return shopcouponCollectionName
}

type ShopCouponData struct {
	models.ShopIdentity `bson:"inline"`
	ShopCouponInfo      `bson:"inline"`
}

type ShopCouponDoc struct {
	ID                 primitive.ObjectID `json:"id" bson:"_id,omitempty"`
	ShopCouponData     `bson:"inline"`
	models.ActivityDoc `bson:"inline"`
}

func (ShopCouponDoc) CollectionName() string {
	return shopcouponCollectionName
}

type ShopCouponItemGuid struct {
	DocNo string `json:"docno" bson:"docno"`
}

func (ShopCouponItemGuid) CollectionName() string {
	return shopcouponCollectionName
}

type ShopCouponActivity struct {
	ShopCouponData      `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ShopCouponActivity) CollectionName() string {
	return shopcouponCollectionName
}

type ShopCouponDeleteActivity struct {
	models.Identity     `bson:"inline"`
	models.ActivityTime `bson:"inline"`
}

func (ShopCouponDeleteActivity) CollectionName() string {
	return shopcouponCollectionName
}
