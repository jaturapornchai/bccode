package shop

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/lib/pq"

	"smlcloudplatform/internal/authentication/models"
	common "smlcloudplatform/internal/models"
	micromodels "smlcloudplatform/pkg/microservice/models"
)

// ErrShopUserNotFound means the user has no membership row in the Holding (or the
// user/Holding is closed).
var ErrShopUserNotFound = errors.New("holding member not found")

type ShopUserPostgresRepository struct {
	db *sql.DB
}

func NewShopUserPostgresRepository(db *sql.DB) IShopUserRepository {
	return &ShopUserPostgresRepository{db: db}
}

// memberSelect reads one membership of an active user in an active Holding. Disabled
// memberships are returned with IsAccessDisabled so callers can explain the denial.
const memberSelect = `SELECT m.id::text, u.id::text, u.username, m.role, m.permission_sets, m.access_scopes, m.is_active,
		m.is_favorite, COALESCE(m.last_accessed_at, 'epoch'::timestamptz), m.created_at
	FROM holding_members m
	JOIN users u ON u.id = m.user_id AND u.is_active = true
	JOIN holdings h ON h.code = m.holding_code AND h.is_active = true
	WHERE m.holding_code = $1`

func (r *ShopUserPostgresRepository) FindByHoldingCodeAndUserUID(ctx context.Context, holdingCode string, userUID string) (models.ShopUser, error) {
	return r.findMember(ctx, holdingCode, `u.id::text = $2`, userUID)
}

func (r *ShopUserPostgresRepository) FindByHoldingCodeAndUsername(ctx context.Context, holdingCode string, username string) (models.ShopUser, error) {
	return r.findMember(ctx, holdingCode, `LOWER(u.username) = LOWER($2)`, username)
}

func (r *ShopUserPostgresRepository) findMember(ctx context.Context, holdingCode string, identityFilter string, identity string) (models.ShopUser, error) {
	holdingCode = strings.TrimSpace(holdingCode)
	identity = strings.TrimSpace(identity)
	if holdingCode == "" || identity == "" {
		return models.ShopUser{}, ErrShopUserNotFound
	}
	db := conn(ctx, r.db)
	member, err := scanMember(db.QueryRowContext(ctx, memberSelect+` AND `+identityFilter, holdingCode, identity), holdingCode)
	if errors.Is(err, sql.ErrNoRows) {
		return models.ShopUser{}, ErrShopUserNotFound
	}
	if err != nil {
		return models.ShopUser{}, err
	}
	if len(member.AccessScopes) == 0 && (member.Role == models.ROLE_OWNER || member.Role == models.ROLE_ADMIN) {
		if member.AccessScopes, err = allCompanyScopes(ctx, db, holdingCode); err != nil {
			return models.ShopUser{}, err
		}
	}
	return member, nil
}

func scanMember(row interface{ Scan(...interface{}) error }, holdingCode string) (models.ShopUser, error) {
	var (
		member         models.ShopUser
		role           string
		permissionSets []byte
		accessScopes   []byte
		isActive       bool
		lastAccessedAt time.Time
	)
	err := row.Scan(&member.MembershipUID, &member.UserUID, &member.Username, &role, &permissionSets, &accessScopes, &isActive,
		&member.IsFavorite, &lastAccessedAt, &member.CreatedAt)
	if err != nil {
		return models.ShopUser{}, err
	}
	if len(permissionSets) > 0 && json.Unmarshal(permissionSets, &member.PermissionSets) != nil {
		return models.ShopUser{}, errors.New("invalid permission sets")
	}
	if member.AccessScopes, err = models.ParseAccessScopes(accessScopes); err != nil {
		return models.ShopUser{}, err
	}
	if member.PermissionSets == nil {
		member.PermissionSets = []string{}
	}
	if lastAccessedAt.After(time.Unix(0, 0)) {
		member.LastAccessedAt = lastAccessedAt
	}
	member.ID = member.MembershipUID
	member.HoldingUID = holdingCode
	member.HoldingCode = holdingCode
	member.Role = models.RoleFromText(role)
	member.IsAccessDisabled = !isActive
	return member, nil
}

// allCompanyScopes grants a Holding manager without explicit scopes every active company.
func allCompanyScopes(ctx context.Context, db dbtx, holdingCode string) ([]models.AccessScope, error) {
	rows, err := db.QueryContext(ctx, `SELECT code FROM companies WHERE holding_code = $1 AND is_active = true ORDER BY code`, holdingCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	scopes := []models.AccessScope{}
	for rows.Next() {
		var code string
		if err := rows.Scan(&code); err != nil {
			return nil, err
		}
		scopes = append(scopes, models.AccessScope{ScopeType: "company", CompanyUID: code, BusinessCode: code, AllBranches: true})
	}
	return scopes, rows.Err()
}

// SaveFullProfile upserts a membership by username, creating the user row when the
// person has not logged in yet (the username/email binds on first login).
func (r *ShopUserPostgresRepository) SaveFullProfile(ctx context.Context, holdingCode string, req *models.UserRoleRequest) error {
	holdingCode = strings.TrimSpace(holdingCode)
	if holdingCode == "" || req == nil {
		return errors.New("holdingCode and request required")
	}
	username := strings.TrimSpace(req.Username)
	if username == "" && strings.TrimSpace(req.UserUID) == "" {
		return errors.New("username required")
	}
	permissionSets := req.PermissionSets
	if permissionSets == nil {
		permissionSets = []string{}
	}
	permsJSON, err := json.Marshal(permissionSets)
	if err != nil {
		return err
	}
	scopesJSON := []byte("[]")
	if len(req.AccessScopes) > 0 {
		if scopesJSON, err = json.Marshal(req.AccessScopes); err != nil {
			return err
		}
	}

	return runInTx(ctx, r.db, func(txCtx context.Context) error {
		db := conn(txCtx, r.db)
		userID := strings.TrimSpace(req.UserUID)
		if userID == "" {
			err := db.QueryRowContext(txCtx, `SELECT id::text FROM users WHERE LOWER(username) = LOWER($1)`, username).Scan(&userID)
			if errors.Is(err, sql.ErrNoRows) {
				err = db.QueryRowContext(txCtx, `
					INSERT INTO users (username, password_hash, email, full_name, is_active)
					VALUES ($1, '', NULLIF($2, ''), $3, true) RETURNING id::text`,
					username, strings.TrimSpace(req.Email), strings.TrimSpace(req.UserProfileName)).Scan(&userID)
			}
			if err != nil {
				return err
			}
		} else if req.Email != "" || req.UserProfileName != "" {
			if _, err := db.ExecContext(txCtx, `
				UPDATE users SET email = COALESCE(NULLIF($1, ''), email), full_name = COALESCE(NULLIF($2, ''), full_name), updated_at = now()
				WHERE id::text = $3`, strings.TrimSpace(req.Email), strings.TrimSpace(req.UserProfileName), userID); err != nil {
				return err
			}
		}
		_, err := db.ExecContext(txCtx, `
			INSERT INTO holding_members (holding_code, user_id, role, permission_sets, access_scopes, is_active)
			VALUES ($1, $2::uuid, $3, $4, $5, $6)
			ON CONFLICT (holding_code, user_id) DO UPDATE SET
				role = EXCLUDED.role,
				permission_sets = EXCLUDED.permission_sets,
				access_scopes = EXCLUDED.access_scopes,
				is_active = EXCLUDED.is_active`,
			holdingCode, userID, models.RoleText(req.Role), string(permsJSON), string(scopesJSON), !req.IsAccessDisabled)
		return err
	})
}

// SaveStable makes an existing user a member with the given role (Holding creation).
// Managers keep empty scopes, which grant every active company.
func (r *ShopUserPostgresRepository) SaveStable(ctx context.Context, holdingCode string, userUID string, role models.UserRole) error {
	_, err := conn(ctx, r.db).ExecContext(ctx, `
		INSERT INTO holding_members (holding_code, user_id, role, permission_sets, access_scopes, is_active)
		VALUES ($1, $2::uuid, $3, '[]'::jsonb, '{}'::jsonb, true)
		ON CONFLICT (holding_code, user_id) DO UPDATE SET role = EXCLUDED.role, is_active = true`,
		strings.TrimSpace(holdingCode), strings.TrimSpace(userUID), models.RoleText(role))
	return err
}

func (r *ShopUserPostgresRepository) Delete(ctx context.Context, holdingCode string, username string) error {
	_, err := conn(ctx, r.db).ExecContext(ctx, `
		DELETE FROM holding_members m
		USING users u
		WHERE m.user_id = u.id AND m.holding_code = $1 AND (LOWER(u.username) = LOWER($2) OR u.id::text = $2)`,
		strings.TrimSpace(holdingCode), strings.TrimSpace(username))
	return err
}

func (r *ShopUserPostgresRepository) UpdateLastAccess(ctx context.Context, holdingCode string, userUID string, lastAccessedAt time.Time) error {
	_, err := conn(ctx, r.db).ExecContext(ctx, `UPDATE holding_members SET last_accessed_at = $3 WHERE holding_code = $1 AND user_id::text = $2`,
		strings.TrimSpace(holdingCode), strings.TrimSpace(userUID), lastAccessedAt.UTC())
	return err
}

func (r *ShopUserPostgresRepository) SaveFavorite(ctx context.Context, holdingCode string, userUID string, isFavorite bool) error {
	_, err := conn(ctx, r.db).ExecContext(ctx, `UPDATE holding_members SET is_favorite = $3 WHERE holding_code = $1 AND user_id::text = $2`,
		strings.TrimSpace(holdingCode), strings.TrimSpace(userUID), isFavorite)
	return err
}

func (r *ShopUserPostgresRepository) FindShopCreatedBy(ctx context.Context, holdingCode string) (string, error) {
	var createdBy string
	err := conn(ctx, r.db).QueryRowContext(ctx, `SELECT created_by FROM holdings WHERE code = $1`, strings.TrimSpace(holdingCode)).Scan(&createdBy)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	return createdBy, err
}

func (r *ShopUserPostgresRepository) FindByUsernamePage(ctx context.Context, username string, pageable micromodels.Pageable) ([]models.ShopUserInfo, common.PaginationData, error) {
	return r.findActiveHoldingPage(ctx, `LOWER(u.username) = LOWER($1)`, username)
}

func (r *ShopUserPostgresRepository) FindByUserUIDPage(ctx context.Context, userUID string, pageable micromodels.Pageable) ([]models.ShopUserInfo, common.PaginationData, error) {
	return r.findActiveHoldingPage(ctx, `u.id::text = $1`, userUID)
}

// holdingListProfile is the part of holdings.profile the Holding selector shows.
type holdingListProfile struct {
	MainHoldingCode string         `json:"mainholdingcode"`
	Names           []common.NameX `json:"names"`
	BranchCode      string         `json:"branchcode"`
	Settings        struct {
		Language            string                  `json:"language"`
		LanguageConfigs     []models.LanguageConfig `json:"languageconfigs"`
		BaseCurrency        string                  `json:"basecurrency"`
		CurrencyCodes       []string                `json:"currencycodes"`
		Timezone            string                  `json:"timezone"`
		TimezoneLabel       string                  `json:"timezonelabel"`
		TimezoneOffset      string                  `json:"timezoneoffset"`
		DateFormat          string                  `json:"dateformat"`
		UseBuddhistCalendar bool                    `json:"usebuddhistcalendar"`
	} `json:"settings"`
}

// findActiveHoldingPage lists the active Holdings a user can select through an active
// membership (the whole list is one page; a user belongs to a handful of Holdings).
func (r *ShopUserPostgresRepository) findActiveHoldingPage(ctx context.Context, filter string, identity string) ([]models.ShopUserInfo, common.PaginationData, error) {
	rows, err := conn(ctx, r.db).QueryContext(ctx, `
		SELECT h.code, h.name, h.profile, h.created_by, m.role, m.is_active, m.is_favorite, COALESCE(m.last_accessed_at, 'epoch'::timestamptz)
		FROM holdings h
		JOIN holding_members m ON m.holding_code = h.code AND m.is_active = true
		JOIN users u ON u.id = m.user_id AND u.is_active = true
		WHERE h.is_active = true AND `+filter+`
		ORDER BY m.is_favorite DESC, m.last_accessed_at DESC NULLS LAST, h.code`, strings.TrimSpace(identity))
	if err != nil {
		return nil, common.PaginationData{}, err
	}
	defer rows.Close()

	result := []models.ShopUserInfo{}
	for rows.Next() {
		var (
			info           models.ShopUserInfo
			role           string
			rawProfile     []byte
			isActive       bool
			lastAccessedAt time.Time
			profile        holdingListProfile
		)
		if err := rows.Scan(&info.HoldingCode, &info.Name, &rawProfile, &info.CreatedBy, &role, &isActive, &info.IsFavorite, &lastAccessedAt); err != nil {
			return nil, common.PaginationData{}, err
		}
		if len(rawProfile) > 0 {
			if err := json.Unmarshal(rawProfile, &profile); err != nil {
				return nil, common.PaginationData{}, err
			}
		}
		info.HoldingUID = info.HoldingCode
		info.Role = models.RoleFromText(role)
		info.IsAccessDisabled = !isActive
		if lastAccessedAt.After(time.Unix(0, 0)) {
			info.LastAccessedAt = lastAccessedAt
		}
		info.MainHoldingCode = profile.MainHoldingCode
		info.Names = profile.Names
		info.BranchCode = profile.BranchCode
		info.Language = profile.Settings.Language
		info.LanguageConfigs = profile.Settings.LanguageConfigs
		info.BaseCurrency = profile.Settings.BaseCurrency
		info.Currencies = profile.Settings.CurrencyCodes
		info.Timezone = profile.Settings.Timezone
		info.TimezoneLabel = profile.Settings.TimezoneLabel
		info.TimezoneOffset = profile.Settings.TimezoneOffset
		info.DateFormat = profile.Settings.DateFormat
		info.UseBuddhistCalendar = profile.Settings.UseBuddhistCalendar
		result = append(result, info)
	}
	return result, common.SinglePage(len(result)), rows.Err()
}

func (r *ShopUserPostgresRepository) FindByUserInShopPageWithProfileMatches(ctx context.Context, holdingCode string, pageable micromodels.Pageable, profileUsernames []string) ([]models.ShopUser, common.PaginationData, error) {
	holdingCode = strings.TrimSpace(holdingCode)
	query := `SELECT m.id::text, u.id::text, u.username, m.role, m.permission_sets, m.access_scopes, m.is_active,
			m.is_favorite, COALESCE(m.last_accessed_at, 'epoch'::timestamptz), m.created_at
		FROM holding_members m
		JOIN users u ON u.id = m.user_id
		WHERE m.holding_code = $1`
	args := []interface{}{holdingCode}
	if q := strings.TrimSpace(pageable.Query); q != "" {
		query += ` AND (u.username ILIKE $2 OR u.full_name ILIKE $2 OR u.email ILIKE $2 OR LOWER(u.username) = ANY($3))`
		args = append(args, "%"+escapeLike(q)+"%", pq.Array(lowerAll(profileUsernames)))
	}
	query += ` ORDER BY u.username`

	rows, err := conn(ctx, r.db).QueryContext(ctx, query, args...)
	if err != nil {
		return nil, common.PaginationData{}, err
	}
	defer rows.Close()
	result := []models.ShopUser{}
	for rows.Next() {
		member, err := scanMember(rows, holdingCode)
		if err != nil {
			return nil, common.PaginationData{}, err
		}
		result = append(result, member)
	}
	return result, common.SinglePage(len(result)), rows.Err()
}

func (r *ShopUserPostgresRepository) FindUsernamesByProfileQuery(ctx context.Context, query string) ([]string, error) {
	query = strings.TrimSpace(query)
	if query == "" {
		return []string{}, nil
	}
	rows, err := conn(ctx, r.db).QueryContext(ctx, `SELECT username FROM users WHERE username ILIKE $1 OR full_name ILIKE $1 OR email ILIKE $1 ORDER BY username LIMIT 500`, "%"+escapeLike(query)+"%")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := []string{}
	for rows.Next() {
		var username string
		if err := rows.Scan(&username); err != nil {
			return nil, err
		}
		list = append(list, username)
	}
	return list, rows.Err()
}

func (r *ShopUserPostgresRepository) FindUserProfileByUsernames(ctx context.Context, usernames []string) ([]models.UserProfile, error) {
	if len(usernames) == 0 {
		return []models.UserProfile{}, nil
	}
	rows, err := conn(ctx, r.db).QueryContext(ctx,
		`SELECT id::text, username, full_name, COALESCE(email, '') FROM users WHERE LOWER(username) = ANY($1)`, pq.Array(lowerAll(usernames)))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := []models.UserProfile{}
	for rows.Next() {
		var profile models.UserProfile
		if err := rows.Scan(&profile.UID, &profile.Username, &profile.Name, &profile.Email); err != nil {
			return nil, err
		}
		result = append(result, profile)
	}
	return result, rows.Err()
}

// ResolveHoldingCodeByHoldingCode returns the stored spelling of a Holding code.
func (r *ShopUserPostgresRepository) ResolveHoldingCodeByHoldingCode(ctx context.Context, holdingCode string) (string, error) {
	holdingCode = strings.TrimSpace(holdingCode)
	var code string
	err := conn(ctx, r.db).QueryRowContext(ctx, `SELECT code FROM holdings WHERE LOWER(code) = LOWER($1) ORDER BY code LIMIT 1`, holdingCode).Scan(&code)
	if errors.Is(err, sql.ErrNoRows) {
		return holdingCode, nil
	}
	return code, err
}

// ResolveCompanyUID maps a company code to its UID; in PostgreSQL the UID is the code.
func (r *ShopUserPostgresRepository) ResolveCompanyUID(ctx context.Context, holdingCode string, businessCode string) (string, error) {
	var code string
	err := conn(ctx, r.db).QueryRowContext(ctx, `SELECT code FROM companies WHERE holding_code = $1 AND UPPER(code) = UPPER($2)`,
		strings.TrimSpace(holdingCode), strings.TrimSpace(businessCode)).Scan(&code)
	if errors.Is(err, sql.ErrNoRows) {
		return "", errors.New("company not found")
	}
	return code, err
}

func lowerAll(values []string) []string {
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value = strings.ToLower(strings.TrimSpace(value)); value != "" {
			out = append(out, value)
		}
	}
	return out
}

func escapeLike(value string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(value)
}
