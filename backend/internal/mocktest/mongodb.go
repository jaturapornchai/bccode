package mocktest

import (
	"testing"

	"go.mongodb.org/mongo-driver/mongo/integration/mtest"
)

func MockMogodb(t *testing.T) {
	mt := mtest.New(t, mtest.NewOptions().ClientType(mtest.Mock))
	mt.Run("test name", func(mt *mtest.T) {
		// test code
	})
}
