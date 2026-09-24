package shop

import (
	"context"
	"database/sql"
	"time"

	authmodels "smlcloudplatform/internal/authentication/models"
	common "smlcloudplatform/internal/models"
	orgaccess "smlcloudplatform/internal/organization"
	"smlcloudplatform/internal/shop/models"
	micromodels "smlcloudplatform/pkg/microservice/models"
)

// IShopRepository stores Holdings (the "shop" of the legacy API) in the central database.
type IShopRepository interface {
	// Transaction runs fn in one database transaction; repositories built on the same
	// *sql.DB join it through the context.
	Transaction(ctx context.Context, fn func(context.Context) error) error
	Create(ctx context.Context, shop models.ShopDoc) error
	// Update rewrites the Holding profile while the Holding still has expectedActive.
	Update(ctx context.Context, holdingCode string, expectedActive bool, shop models.ShopDoc) error
	// UpdateStatus changes is_active only when it still equals expectedActive.
	UpdateStatus(ctx context.Context, holdingCode string, expectedActive bool, active bool, username string, now time.Time) error
	FindByHoldingCode(ctx context.Context, holdingCode string) (models.ShopDoc, error)
	RecordAudit(ctx context.Context, audit orgaccess.Audit) error
}

// IShopUserRepository stores Holding memberships (holding_members) and reads user profiles.
type IShopUserRepository interface {
	FindByHoldingCodeAndUsername(ctx context.Context, holdingCode string, username string) (authmodels.ShopUser, error)
	FindByHoldingCodeAndUserUID(ctx context.Context, holdingCode string, userUID string) (authmodels.ShopUser, error)
	FindByUserUIDPage(ctx context.Context, userUID string, pageable micromodels.Pageable) ([]authmodels.ShopUserInfo, common.PaginationData, error)
	FindByUsernamePage(ctx context.Context, username string, pageable micromodels.Pageable) ([]authmodels.ShopUserInfo, common.PaginationData, error)
	FindByUserInShopPageWithProfileMatches(ctx context.Context, holdingCode string, pageable micromodels.Pageable, profileUsernames []string) ([]authmodels.ShopUser, common.PaginationData, error)
	FindUsernamesByProfileQuery(ctx context.Context, query string) ([]string, error)
	FindUserProfileByUsernames(ctx context.Context, usernames []string) ([]authmodels.UserProfile, error)
	// FindShopCreatedBy returns the user UID that created the Holding.
	FindShopCreatedBy(ctx context.Context, holdingCode string) (string, error)
	SaveFullProfile(ctx context.Context, holdingCode string, req *authmodels.UserRoleRequest) error
	SaveStable(ctx context.Context, holdingCode string, userUID string, role authmodels.UserRole) error
	// Delete removes a membership; actorUID (the admin) is recorded with the removal.
	Delete(ctx context.Context, holdingCode string, username string, actorUID string) error
	UpdateLastAccess(ctx context.Context, holdingCode string, userUID string, lastAccessedAt time.Time) error
	SaveFavorite(ctx context.Context, holdingCode string, userUID string, isFavorite bool) error
	ResolveHoldingCodeByHoldingCode(ctx context.Context, holdingCode string) (string, error)
	ResolveCompanyUID(ctx context.Context, holdingCode string, businessCode string) (string, error)
}

// IShopUserAccessLogRepository records Holding selections.
type IShopUserAccessLogRepository interface {
	Create(ctx context.Context, shopUserAccessLog authmodels.ShopUserAccessLog) error
}

type txKey struct{}

type dbtx interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// conn returns the transaction carried by ctx, or the pool.
func conn(ctx context.Context, db *sql.DB) dbtx {
	if tx, ok := ctx.Value(txKey{}).(*sql.Tx); ok && tx != nil {
		return tx
	}
	return db
}

func runInTx(ctx context.Context, db *sql.DB, fn func(context.Context) error) error {
	if _, ok := ctx.Value(txKey{}).(*sql.Tx); ok {
		return fn(ctx)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := fn(context.WithValue(ctx, txKey{}, tx)); err != nil {
		return err
	}
	return tx.Commit()
}
