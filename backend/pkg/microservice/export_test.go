package microservice

import (
	"context"

	"smlcloudplatform/pkg/microservice/models"
)

// Test hooks for the external (database-backed) test package.

func AuthorizeLiveForTest(db AuthorizationFinder, selected models.UserInfo) (models.UserInfo, error) {
	return newLiveAuthorization(db).Authorize(context.Background(), selected)
}

func SelectWorkspaceLiveForTest(db AuthorizationFinder, identity models.UserInfo, holdingCode, businessCode, branchUID string) (models.UserInfo, error) {
	return newLiveAuthorization(db).SelectWorkspace(context.Background(), identity, holdingCode, businessCode, branchUID)
}
