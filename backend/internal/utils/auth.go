package utils

import (
	"context"
	"fmt"
	"smlcloudplatform/internal/shop/models"
	"smlcloudplatform/pkg/microservice"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
)

func NormalizeUsername(username string) string {
	username = strings.TrimSpace(username)
	username = strings.ToLower(username)
	return username
}

func NormalizePhonenumber(phoneNumber string) string {
	phoneNumber = strings.TrimSpace(phoneNumber)
	phoneNumber = strings.TrimPrefix(phoneNumber, "0")
	return phoneNumber
}

func NormalizeEmail(email string) string {
	email = strings.TrimSpace(email)
	return email
}

func NormalizeName(username string) string {
	username = strings.TrimSpace(username)
	return username
}

func HasPermissionShop(pst microservice.IPersisterMongo, ctx microservice.IContext) (bool, error) {

	shopID := ctx.Param("shop_id")

	return HasPermissionShopByID(pst, ctx, shopID)
}

func HasPermissionShopByID(pst microservice.IPersisterMongo, ctx microservice.IContext, shopID string) (bool, error) {

	authUsername := ctx.UserInfo().Username

	if len(shopID) < 1 {
		return false, fmt.Errorf("shop not found")
	}

	pstContect, pstContextCancel := context.WithTimeout(context.Background(), time.Duration(15)*time.Second)
	defer pstContextCancel()

	shop := &models.ShopDoc{}

	pst.FindOne(pstContect, &models.Shop{}, bson.M{"guidfixed": shopID, "deletedat": bson.M{"$exists": false}}, shop)

	if len(shop.GuidFixed) < 1 {
		return false, fmt.Errorf("shop invalid")
	}

	if shop.CreatedBy != authUsername {
		return false, fmt.Errorf("username invalid")
	}

	return true, nil
}
