package services

import (
	"encoding/json"
	"testing"
	"time"

	"smlcloudplatform/internal/vfgl/journal/models"
	micromodels "smlcloudplatform/pkg/microservice/models"
)

func TestApplyJournalDeletedActorIncludesMQPayload(t *testing.T) {
	doc := models.JournalDoc{}
	actor := micromodels.UserInfo{Username: "user-001", Name: "สมชาย"}
	deletedAt := time.Date(2026, time.August, 5, 10, 0, 0, 0, time.UTC)

	applyJournalDeletedActor(&doc, actor, deletedAt)

	if doc.ActivityDoc.DeletedBy != actor.Username || doc.JournalInfo.DeletedBy != actor.Username {
		t.Fatal("deleted actor code was not copied to both persistence and event models")
	}
	payload, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	decoded := map[string]interface{}{}
	if err := json.Unmarshal(payload, &decoded); err != nil {
		t.Fatal(err)
	}
	if decoded["deletedby"] != actor.Username || decoded["deletedbyname"] != actor.Name {
		t.Fatalf("delete actor missing from MQ payload: %s", payload)
	}
}
