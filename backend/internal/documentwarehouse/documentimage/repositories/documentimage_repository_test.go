package repositories_test

import (
	"smlcloudplatform/internal/documentwarehouse/documentimage/repositories"
	"smlcloudplatform/mock"
	"smlcloudplatform/pkg/microservice"
	"testing"
)

var repoMock repositories.DocumentImageRepository

func init() {
	mongoPersisterConfig := mock.NewPersisterMongoConfig()
	mongoPersister := microservice.NewPersisterMongo(mongoPersisterConfig)
	repoMock = repositories.NewDocumentImageRepository(mongoPersister)
}

func TestDocumentImageGroup(t *testing.T) {
	// listx := repoMock.DocumentImageGroup("27dcEdktOoaSBYFmnN6G6ett4Jb")
	// fmt.Println(listx)
}

func TestUpdateImageDocGroup(t *testing.T) {

	// err := repoMock.SaveDocumentImageDocRefGroup("27dcEdktOoaSBYFmnN6G6ett4Jb", "tester", []string{"2AbpV69VhKfBY6zHKxaRQydmtm0", "2AeBoO0ZNzhGPRf556vySRqnqew"})

	// if err != nil {
	// 	t.Error(err)
	// 	return
	// }
}
