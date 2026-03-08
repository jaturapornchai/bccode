package journalreport_test

import (
	"context"
	"fmt"
	"os"
	"smlcloudplatform/internal/vfgl/journalreport"
	"smlcloudplatform/mock"
	"smlcloudplatform/pkg/microservice"
	"testing"

	"github.com/stretchr/testify/assert"
)

var repoMock journalreport.IJournalReportMongoRepository

func init() {
	mongoPersisterConfig := mock.NewPersisterMongoConfig()
	mongoPersister := microservice.NewPersisterMongo(mongoPersisterConfig)
	repoMock = journalreport.NewJournalMongoRepository(mongoPersister)
}

func TestFindDetailByGUIDs(t *testing.T) {

	if os.Getenv("SERVERLESS") == "serverless" {
		t.Skip()
	}

	docList, err := repoMock.FindCountDetailByDocs(context.Background(), "27dcEdktOoaSBYFmnN6G6ett4Jb", []string{"JO-20220706F8F4CA", "JO-202207069A2102"})

	if err != nil {
		fmt.Println(err)
	}

	assert.Nil(t, err)
	assert.Equal(t, 2, len(docList))

}
