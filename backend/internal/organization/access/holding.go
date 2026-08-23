package access

import (
	"context"
	"errors"
	"fmt"
	"strings"

	shopmodels "smlcloudplatform/internal/shop/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

var ErrActiveHoldingRequired = errors.New("active Holding required")

func RequireActiveHolding(ctx context.Context, finder MembershipFinder, holdingCode string) error {
	_, err := FindActiveHolding(ctx, finder, holdingCode)
	return err
}

func FindActiveHolding(ctx context.Context, finder MembershipFinder, holdingCode string) (shopmodels.ShopDoc, error) {
	holdingCode = strings.TrimSpace(holdingCode)
	if holdingCode == "" {
		return shopmodels.ShopDoc{}, ErrActiveHoldingRequired
	}

	var holding shopmodels.ShopDoc
	err := finder.FindOne(ctx, &shopmodels.ShopDoc{}, bson.M{
		"holdingcode": holdingCode,
		"isactive":    true,
		"isdeleted":   false,
		"deletedat":   bson.M{"$exists": false},
	}, &holding)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return shopmodels.ShopDoc{}, ErrActiveHoldingRequired
		}
		return shopmodels.ShopDoc{}, fmt.Errorf("load Holding: %w", err)
	}
	if holding.ID == primitive.NilObjectID || holding.IsDeleted || !holding.IsActive {
		return shopmodels.ShopDoc{}, ErrActiveHoldingRequired
	}
	return holding, nil
}
