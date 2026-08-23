package shop_test

import (
	"context"
	"testing"
	"time"

	"smlcloudplatform/internal/shop"
	"smlcloudplatform/pkg/microservice"
	micromodels "smlcloudplatform/pkg/microservice/models"

	"github.com/smlsoft/mongopagination"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
)

type listPagePersisterSpy struct {
	microservice.IPersisterMongo
	criteria []interface{}
}

func (spy *listPagePersisterSpy) AggregatePage(_ context.Context, _ interface{}, _ micromodels.Pageable, criteria ...interface{}) (*mongopagination.PaginatedData, error) {
	spy.criteria = criteria
	return &mongopagination.PaginatedData{}, nil
}

func TestFindByUserUIDPageUsesStableActiveMembershipPipeline(t *testing.T) {
	spy := &listPagePersisterSpy{}
	repo := shop.NewShopUserRepository(spy)
	before := time.Now().UTC()

	_, _, err := repo.FindByUserUIDPage(context.Background(), "user-uid", micromodels.Pageable{})

	after := time.Now().UTC()
	require.NoError(t, err)
	require.Len(t, spy.criteria, 7)

	membershipMatch := spy.criteria[0].(bson.M)["$match"].(bson.M)
	require.Equal(t, "user-uid", membershipMatch["useruid"])
	require.Equal(t, false, membershipMatch["isdeleted"])
	require.Equal(t, false, membershipMatch["isaccessdisabled"])
	expiryRules := membershipMatch["$or"].(bson.A)
	require.Equal(t, bson.M{"accessexpirydate": bson.M{"$exists": false}}, expiryRules[0])
	expiryCutoff := expiryRules[1].(bson.M)["accessexpirydate"].(bson.M)["$gt"].(time.Time)
	require.False(t, expiryCutoff.Before(before))
	require.False(t, expiryCutoff.After(after))

	holdingLookup := spy.criteria[1].(bson.M)["$lookup"].(bson.M)
	require.Equal(t, "holdinguid", holdingLookup["localField"])
	require.Equal(t, "holdinguid", holdingLookup["foreignField"])

	holdingMatch := spy.criteria[3].(bson.M)["$match"].(bson.M)
	require.Equal(t, bson.M{"$elemMatch": bson.M{
		"isactive":  true,
		"isdeleted": false,
	}}, holdingMatch["shopInfo"])

	projection := spy.criteria[4].(bson.M)["$project"].(bson.M)
	require.Equal(t, 1, projection["holdinguid"])
}
