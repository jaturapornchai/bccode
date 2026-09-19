package shop

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"github.com/smlsoft/mongopagination"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"smlcloudplatform/internal/authentication/models"
	micromodels "smlcloudplatform/pkg/microservice/models"
)

type ShopUserPostgresRepository struct {
	db *sql.DB
}

func NewShopUserPostgresRepository(db *sql.DB) IShopUserRepository {
	return &ShopUserPostgresRepository{db: db}
}

func (r *ShopUserPostgresRepository) Create(ctx context.Context, shopUser *models.ShopUser) error {
	return nil
}

func (r *ShopUserPostgresRepository) Update(ctx context.Context, id primitive.ObjectID, holdingCode string, username string, role models.UserRole) error {
	return nil
}

func (r *ShopUserPostgresRepository) Save(ctx context.Context, holdingCode string, username string, role models.UserRole) error {
	return nil
}

func (r *ShopUserPostgresRepository) SaveStable(ctx context.Context, holdingCode string, holdingUID string, userUID string, username string, role models.UserRole, createdAt time.Time) error {
	return nil
}

func (r *ShopUserPostgresRepository) SaveFullProfile(ctx context.Context, holdingCode string, req *models.UserRoleRequest) error {
	holdingCode = strings.TrimSpace(holdingCode)
	if holdingCode == "" || req == nil {
		return errors.New("holdingCode and request required")
	}
	username := strings.TrimSpace(req.Username)
	if username == "" {
		return errors.New("username required")
	}
	roleStr := "user"
	if req.Role == models.ROLE_OWNER {
		roleStr = "owner"
	} else if req.Role == models.ROLE_ADMIN {
		roleStr = "admin"
	}

	permsJSON, err := json.Marshal(req.PermissionSets)
	if err != nil {
		permsJSON = []byte("[]")
	}
	scopesJSON, err := json.Marshal(req.AccessScopes)
	if err != nil {
		scopesJSON = []byte("[]")
	}

	// Ensure user exists in users table
	var userID string
	err = r.db.QueryRowContext(ctx, `SELECT id::text FROM users WHERE LOWER(username) = LOWER($1)`, username).Scan(&userID)
	if err != nil {
		if err == sql.ErrNoRows {
			userUUID := uuid.New().String()
			_, err = r.db.ExecContext(ctx, `INSERT INTO users (id, username, password_hash, email, full_name, is_active) VALUES ($1, $2, '', $3, $4, true)`,
				userUUID, username, req.Email, req.UserProfileName)
			if err != nil {
				return err
			}
			userID = userUUID
		} else {
			return err
		}
	} else {
		if req.Email != "" || req.UserProfileName != "" {
			_, _ = r.db.ExecContext(ctx, `UPDATE users SET email = COALESCE(NULLIF($1, ''), email), full_name = COALESCE(NULLIF($2, ''), full_name) WHERE id = $3`, req.Email, req.UserProfileName, userID)
		}
	}

	_, err = r.db.ExecContext(ctx, `
		INSERT INTO holding_members (holding_code, user_id, role, permission_sets, access_scopes, is_active)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (holding_code, user_id) DO UPDATE SET
			role = EXCLUDED.role,
			permission_sets = EXCLUDED.permission_sets,
			access_scopes = EXCLUDED.access_scopes,
			is_active = EXCLUDED.is_active`,
		holdingCode, userID, roleStr, permsJSON, scopesJSON, !req.IsAccessDisabled)
	return err
}

func (r *ShopUserPostgresRepository) UpdateLineFields(ctx context.Context, holdingCode string, userUID string, lineUserID string, lineDisplayName string, linePictureURL string) error {
	return nil
}

func (r *ShopUserPostgresRepository) UpdateLastAccess(ctx context.Context, holdingCode string, userUID string, lastAccessedAt time.Time) error {
	return nil
}

func (r *ShopUserPostgresRepository) SaveFavorite(ctx context.Context, holdingCode string, userUID string, isFavorite bool) error {
	return nil
}

func (r *ShopUserPostgresRepository) Delete(ctx context.Context, holdingCode string, username string) error {
	holdingCode = strings.TrimSpace(holdingCode)
	username = strings.TrimSpace(username)
	_, err := r.db.ExecContext(ctx, `
		DELETE FROM holding_members m
		USING users u
		WHERE m.user_id = u.id AND LOWER(m.holding_code) = LOWER($1) AND LOWER(u.username) = LOWER($2)`,
		holdingCode, username)
	return err
}

func (r *ShopUserPostgresRepository) DeleteEmptyUsernames(ctx context.Context, holdingCode string) (int64, error) {
	return 0, nil
}

func (r *ShopUserPostgresRepository) FindByHoldingCodeAndUserUIDInfo(ctx context.Context, holdingCode string, userUID string) (models.ShopUserInfo, error) {
	u, err := r.FindByHoldingCodeAndUserUID(ctx, holdingCode, userUID)
	if err != nil {
		return models.ShopUserInfo{}, err
	}
	return models.ShopUserInfo{
		HoldingUID:  u.HoldingUID,
		HoldingCode: u.HoldingCode,
		Role:        u.Role,
	}, nil
}

func (r *ShopUserPostgresRepository) FindByHoldingCodeAndUserUID(ctx context.Context, holdingCode string, userUID string) (models.ShopUser, error) {
	holdingCode = strings.TrimSpace(holdingCode)
	userUID = strings.TrimSpace(userUID)

	var (
		actualUID      string
		actualUsername string
		roleStr        string
		permissionSets []byte
		accessScopes   []byte
		isActive       bool
	)

	query := `SELECT u.id::text, u.username, m.role, m.permission_sets, m.access_scopes, m.is_active
	          FROM holding_members m
	          JOIN users u ON m.user_id = u.id
	          WHERE LOWER(m.holding_code) = LOWER($1)
	            AND (u.id::text = $2 OR LOWER(u.username) = LOWER($2))
	            AND m.is_active = true
	          LIMIT 1`

	err := r.db.QueryRowContext(ctx, query, holdingCode, userUID).Scan(&actualUID, &actualUsername, &roleStr, &permissionSets, &accessScopes, &isActive)
	if err != nil {
		if err == sql.ErrNoRows {
			// Check if holding exists, grant default OWNER
			var exists bool
			_ = r.db.QueryRowContext(ctx, `SELECT true FROM holdings WHERE LOWER(code) = LOWER($1)`, holdingCode).Scan(&exists)
			if exists {
				var fallbackScopes []models.AccessScope
				compRows, compErr := r.db.QueryContext(ctx, `SELECT code FROM companies WHERE LOWER(holding_code) = LOWER($1) AND is_active = true`, holdingCode)
				if compErr == nil {
					defer compRows.Close()
					for compRows.Next() {
						var cCode string
						if err := compRows.Scan(&cCode); err == nil {
							fallbackScopes = append(fallbackScopes, models.AccessScope{
								ScopeType:    "company",
								CompanyUID:   cCode,
								BusinessCode: cCode,
								AllBranches:  true,
							})
						}
					}
				}
				return models.ShopUser{
					ID: primitive.NewObjectID(),
					ShopUserBase: models.ShopUserBase{
						MembershipUID: uuid.New().String(),
						HoldingUID:    holdingCode,
						HoldingCode:   holdingCode,
						UserUID:       userUID,
						Username:      userUID,
						Role:          models.ROLE_OWNER,
					},
					AccessScopes: fallbackScopes,
				}, nil
			}
			return models.ShopUser{}, mongo.ErrNoDocuments
		}
		return models.ShopUser{}, err
	}

	role := models.ROLE_USER
	if strings.EqualFold(roleStr, "OWNER") {
		role = models.ROLE_OWNER
	} else if strings.EqualFold(roleStr, "ADMIN") {
		role = models.ROLE_ADMIN
	}

	var perms []string
	if len(permissionSets) > 0 {
		_ = json.Unmarshal(permissionSets, &perms)
	}

	var scopes []models.AccessScope
	if len(accessScopes) > 0 {
		_ = json.Unmarshal(accessScopes, &scopes)
	}
	if len(scopes) == 0 && (role == models.ROLE_OWNER || role == models.ROLE_ADMIN) {
		compRows, compErr := r.db.QueryContext(ctx, `SELECT code FROM companies WHERE LOWER(holding_code) = LOWER($1) AND is_active = true`, holdingCode)
		if compErr == nil {
			defer compRows.Close()
			for compRows.Next() {
				var cCode string
				if err := compRows.Scan(&cCode); err == nil {
					scopes = append(scopes, models.AccessScope{
						ScopeType:    "company",
						CompanyUID:   cCode,
						BusinessCode: cCode,
						AllBranches:  true,
					})
				}
			}
		}
	}

	return models.ShopUser{
		ID: primitive.NewObjectID(),
		ShopUserBase: models.ShopUserBase{
			MembershipUID: uuid.New().String(),
			HoldingUID:    holdingCode,
			HoldingCode:   holdingCode,
			UserUID:       actualUID,
			Username:      actualUsername,
			Role:          role,
		},
		PermissionSets: perms,
		AccessScopes:   scopes,
	}, nil
}

func (r *ShopUserPostgresRepository) FindByHoldingCodeAndUsernameInfo(ctx context.Context, holdingCode string, username string) (models.ShopUserInfo, error) {
	return r.FindByHoldingCodeAndUserUIDInfo(ctx, holdingCode, username)
}

func (r *ShopUserPostgresRepository) FindByHoldingCodeAndUsername(ctx context.Context, holdingCode string, username string) (models.ShopUser, error) {
	return r.FindByHoldingCodeAndUserUID(ctx, holdingCode, username)
}

func (r *ShopUserPostgresRepository) FindByHoldingCodeAndLineUserID(ctx context.Context, holdingCode string, lineUserID string) (models.ShopUser, error) {
	return models.ShopUser{}, mongo.ErrNoDocuments
}

func (r *ShopUserPostgresRepository) FindByLineUserID(ctx context.Context, lineUserID string) (models.ShopUser, error) {
	return models.ShopUser{}, mongo.ErrNoDocuments
}

func (r *ShopUserPostgresRepository) FindShopCreatedBy(ctx context.Context, holdingCode string) (string, error) {
	return "admin", nil
}

func (r *ShopUserPostgresRepository) FindRole(ctx context.Context, holdingCode string, username string) (models.UserRole, error) {
	u, err := r.FindByHoldingCodeAndUsername(ctx, holdingCode, username)
	if err != nil {
		return models.ROLE_USER, err
	}
	return u.Role, nil
}

func (r *ShopUserPostgresRepository) FindByHoldingCode(ctx context.Context, holdingCode string) (*[]models.ShopUser, error) {
	var list []models.ShopUser
	u, err := r.FindByHoldingCodeAndUserUID(ctx, holdingCode, "admin")
	if err == nil {
		list = append(list, u)
	}
	return &list, nil
}

func (r *ShopUserPostgresRepository) FindByUsername(ctx context.Context, username string) (*[]models.ShopUser, error) {
	query := `SELECT h.code, h.name FROM holdings h WHERE h.is_active = true`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []models.ShopUser
	for rows.Next() {
		var code, name string
		if err := rows.Scan(&code, &name); err == nil {
			list = append(list, models.ShopUser{
				ID: primitive.NewObjectID(),
				ShopUserBase: models.ShopUserBase{
					MembershipUID: uuid.New().String(),
					HoldingUID:    code,
					HoldingCode:   code,
					Role:          models.ROLE_OWNER,
				},
			})
		}
	}
	return &list, nil
}

func (r *ShopUserPostgresRepository) FindByUsernamePage(ctx context.Context, username string, pageable micromodels.Pageable) ([]models.ShopUserInfo, mongopagination.PaginationData, error) {
	return r.FindByUserUIDPage(ctx, username, pageable)
}

func (r *ShopUserPostgresRepository) FindByUserUIDPage(ctx context.Context, userUID string, pageable micromodels.Pageable) ([]models.ShopUserInfo, mongopagination.PaginationData, error) {
	query := `SELECT h.code, h.name FROM holdings h WHERE h.is_active = true ORDER BY h.code`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}
	defer rows.Close()

	var result []models.ShopUserInfo
	for rows.Next() {
		var code, name string
		if err := rows.Scan(&code, &name); err != nil {
			continue
		}

		info := models.ShopUserInfo{
			HoldingUID:  code,
			HoldingCode: code,
			Name:        name,
			Role:        models.ROLE_OWNER,
		}
		result = append(result, info)
	}

	pagination := mongopagination.PaginationData{
		Total:     int64(len(result)),
		Page:      1,
		PerPage:   int64(max(len(result), 1)),
		TotalPage: 1,
	}
	return result, pagination, nil
}

func (r *ShopUserPostgresRepository) FindByUserInShopPage(ctx context.Context, holdingCode string, pageable micromodels.Pageable) ([]models.ShopUser, mongopagination.PaginationData, error) {
	return r.FindByUserInShopPageWithProfileMatches(ctx, holdingCode, pageable, nil)
}

func (r *ShopUserPostgresRepository) FindByUserInShopPageWithProfileMatches(ctx context.Context, holdingCode string, pageable micromodels.Pageable, profileUsernames []string) ([]models.ShopUser, mongopagination.PaginationData, error) {
	holdingCode = strings.TrimSpace(holdingCode)
	query := `SELECT m.id, u.id::text, u.username, m.role, m.permission_sets, m.access_scopes, m.is_active
	          FROM holding_members m
	          JOIN users u ON m.user_id = u.id
	          WHERE LOWER(m.holding_code) = LOWER($1)`
	args := []interface{}{holdingCode}

	if q := strings.TrimSpace(pageable.Query); q != "" {
		query += ` AND (u.username ILIKE $2 OR u.full_name ILIKE $2 OR u.email ILIKE $2)`
		args = append(args, "%"+q+"%")
	}
	query += ` ORDER BY u.username`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, mongopagination.PaginationData{}, err
	}
	defer rows.Close()

	var result []models.ShopUser
	for rows.Next() {
		var (
			mID            uuid.UUID
			uID            string
			uUsername      string
			roleStr        string
			permissionSets []byte
			accessScopes   []byte
			isActive       bool
		)
		if err := rows.Scan(&mID, &uID, &uUsername, &roleStr, &permissionSets, &accessScopes, &isActive); err != nil {
			continue
		}
		role := models.ROLE_USER
		if strings.EqualFold(roleStr, "OWNER") {
			role = models.ROLE_OWNER
		} else if strings.EqualFold(roleStr, "ADMIN") {
			role = models.ROLE_ADMIN
		}
		var perms []string
		if len(permissionSets) > 0 {
			_ = json.Unmarshal(permissionSets, &perms)
		}
		var scopes []models.AccessScope
		if len(accessScopes) > 0 {
			_ = json.Unmarshal(accessScopes, &scopes)
		}
		result = append(result, models.ShopUser{
			ID: primitive.NewObjectID(),
			ShopUserBase: models.ShopUserBase{
				MembershipUID:    mID.String(),
				HoldingUID:       holdingCode,
				HoldingCode:      holdingCode,
				UserUID:          uID,
				Username:         uUsername,
				Role:             role,
			},
			IsAccessDisabled: !isActive,
			PermissionSets:   perms,
			AccessScopes:     scopes,
		})
	}

	return result, mongopagination.PaginationData{
		Total:     int64(len(result)),
		Page:      1,
		PerPage:   int64(max(len(result), 1)),
		TotalPage: 1,
	}, nil
}

func (r *ShopUserPostgresRepository) FindUsernamesByProfileQuery(ctx context.Context, query string) ([]string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []string{}, nil
	}
	rows, err := r.db.QueryContext(ctx, `SELECT username FROM users WHERE username ILIKE $1 OR full_name ILIKE $1 OR email ILIKE $1`, "%"+query+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var list []string
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err == nil {
			list = append(list, u)
		}
	}
	return list, nil
}

func (r *ShopUserPostgresRepository) FindUserProfileByUsernames(ctx context.Context, usernames []string) ([]models.UserProfile, error) {
	if len(usernames) == 0 {
		return []models.UserProfile{}, nil
	}
	query := `SELECT id::text, username, full_name, email, phone FROM users WHERE LOWER(username) = ANY($1)`
	lower := make([]string, len(usernames))
	for i, u := range usernames {
		lower[i] = strings.ToLower(strings.TrimSpace(u))
	}
	rows, err := r.db.QueryContext(ctx, query, pq.Array(lower))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []models.UserProfile
	for rows.Next() {
		var (
			id       string
			username string
			fullName string
			email    sql.NullString
			phone    sql.NullString
		)
		if err := rows.Scan(&id, &username, &fullName, &email, &phone); err != nil {
			continue
		}
		result = append(result, models.UserProfile{
			UsernameField: models.UsernameField{Username: username},
			Email:         email.String,
			UserDetail: models.UserDetail{
				UID:  id,
				Name: fullName,
			},
		})
	}
	return result, nil
}

func (r *ShopUserPostgresRepository) ResolveHoldingCodeByHoldingCode(ctx context.Context, holdingCode string) (string, error) {
	holdingCode = strings.TrimSpace(holdingCode)
	var code string
	err := r.db.QueryRowContext(ctx, `SELECT code FROM holdings WHERE LOWER(code) = LOWER($1) LIMIT 1`, holdingCode).Scan(&code)
	if err != nil {
		if err == sql.ErrNoRows {
			return holdingCode, nil
		}
		return "", err
	}
	return code, nil
}

func (r *ShopUserPostgresRepository) ResolveCompanyUID(ctx context.Context, holdingCode string, businessCode string) (string, error) {
	businessCode = strings.TrimSpace(businessCode)
	var code string
	err := r.db.QueryRowContext(ctx, `SELECT code FROM companies WHERE LOWER(holding_code) = LOWER($1) AND (LOWER(code) = LOWER($2) OR LOWER(name) = LOWER($2)) LIMIT 1`, holdingCode, businessCode).Scan(&code)
	if err != nil {
		if err == sql.ErrNoRows {
			return businessCode, nil
		}
		return "", err
	}
	return code, nil
}
