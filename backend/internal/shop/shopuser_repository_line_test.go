package shop_test

import (
	"context"
	"testing"

	"smlcloudplatform/internal/shop"
	"smlcloudplatform/pkg/microservice"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type lineFieldsPersisterSpy struct {
	microservice.IPersisterMongo
	filter interface{}
	update interface{}
}

func (spy *lineFieldsPersisterSpy) Update(_ context.Context, _ interface{}, filter interface{}, update interface{}, _ ...*options.UpdateOptions) error {
	spy.filter = filter
	spy.update = update
	return nil
}

func TestUpdateLineFieldsDoesNotWriteAuthorizationFields(t *testing.T) {
	spy := &lineFieldsPersisterSpy{}
	repo := shop.NewShopUserRepository(spy)

	err := repo.UpdateLineFields(context.Background(), "holdingcode", "user-uid", "line-id", "Line Name", "https://example.test/line.png")

	require.NoError(t, err)
	require.Equal(t, bson.M{
		"holdingcode": "holdingcode",
		"useruid":     "user-uid",
		"isdeleted":   bson.M{"$ne": true},
	}, spy.filter)
	require.Equal(t, bson.M{"$set": bson.M{
		"lineuserid":      "line-id",
		"linedisplayname": "Line Name",
		"linepictureurl":  "https://example.test/line.png",
	}}, spy.update)
}
