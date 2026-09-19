package member_test

import (
	"os"
	"smlcloudplatform/internal/member"
	"smlcloudplatform/internal/member/models"
	"smlcloudplatform/mock"
	"smlcloudplatform/pkg/microservice"
	"testing"
)

func newPgRepo() member.MemberPGRepository {
	persisterConfig := mock.NewPersisterPostgresqlConfig()
	pst := microservice.NewPersister(persisterConfig)
	repo := member.NewMemberPGRepository(pst)
	return repo
}

func TestCreate(t *testing.T) {
	if os.Getenv("SERVERLESS") == "serverless" || os.Getenv("TEST_POSTGRES_HOST") == "" {
		t.Skip("skipping postgres integration test")
	}
	repo := newPgRepo()

	idx := models.MemberIndex{}
	idx.ID = "134567"
	idx.HoldingCode = "holdingcodex001"
	idx.GuidFixed = "fixguid"
	err := repo.Create(idx)

	if err != nil {
		t.Error(err)
	}
}

func TestCount(t *testing.T) {
	if os.Getenv("SERVERLESS") == "serverless" || os.Getenv("TEST_POSTGRES_HOST") == "" {
		t.Skip("skipping postgres integration test")
	}
	repo := newPgRepo()

	count, err := repo.Count("holdingcodex001", "fixguid")

	if err != nil {
		t.Error(err)
	}

	t.Log(count)
}

func TestFindByGuid(t *testing.T) {
	if os.Getenv("SERVERLESS") == "serverless" || os.Getenv("TEST_POSTGRES_HOST") == "" {
		t.Skip("skipping postgres integration test")
	}
	repo := newPgRepo()
	inv, err := repo.FindByGuid("holdingcodex001", "fixguid")

	if err != nil {
		t.Error(err)
	}

	t.Log(inv)
}

func TestDelete(t *testing.T) {
	if os.Getenv("SERVERLESS") == "serverless" || os.Getenv("TEST_POSTGRES_HOST") == "" {
		t.Skip("skipping postgres integration test")
	}
	repo := newPgRepo()

	err := repo.Delete("holdingcodex001", "fixguid")

	if err != nil {
		t.Error(err)
	}
}
