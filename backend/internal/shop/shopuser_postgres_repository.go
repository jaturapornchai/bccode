package shop

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/lib/pq"

	"smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/centraldb"
	common "smlcloudplatform/internal/models"
	micromodels "smlcloudplatform/pkg/microservice/models"
)

// ErrShopUserNotFound means the user has no membership row in the Holding (or the
// user/Holding is closed).
var ErrShopUserNotFound = errors.New("holding member not found")

// errMemberRequestInvalid wraps a save the caller can fix (no holding/username, malformed
// useruid): the handler answers 400 VALIDATION_FAILED instead of a 500.
var errMemberRequestInvalid = errors.New("member request invalid")

// Save rejections the settings screen explains per field (shopUserErrorRow).
var (
	// errMemberAlreadyExists: an ADD named a username that is already a member here.
	errMemberAlreadyExists = errors.New("member already exists")
	// errUsernameTaken: the username already has a login account (for example in another
	// business group) and the request did not explicitly ask to add that account.
	errUsernameTaken = errors.New("username taken")
	// errLoginExists: the username has exactly one login account and no other business group uses
	// it — the same save with addexistinguser attaches it. The screen asks before sending that, so
	// only this answer opens the "attach the existing account" dialog (errUsernameTaken never can
	// be attached; review 2026-09-24: the dialog used to promise an attach the save then refused).
	errLoginExists = errors.New("login account exists")
	// errUserCodeLocked / errUserEmailLocked: the account is not managed by this business
	// group (someone signed in with it, or another group uses it), so its usercode/email
	// cannot be changed from here.
	errUserCodeLocked  = errors.New("usercode locked")
	errUserEmailLocked = errors.New("email locked")
	// errSaveTargetNotFound: an EDIT addressed a member that is not in this business group.
	errSaveTargetNotFound = errors.New("user not found")
)

type ShopUserPostgresRepository struct {
	db *sql.DB
}

func NewShopUserPostgresRepository(db *sql.DB) IShopUserRepository {
	return &ShopUserPostgresRepository{db: db}
}

// memberColumns is what scanMember reads (m = holding_members, u = users, h = holdings).
const memberColumns = `m.id::text, u.id::text, u.username, m.role, m.permission_sets, m.access_scopes, m.is_active,
		m.is_favorite, COALESCE(m.last_accessed_at, 'epoch'::timestamptz), m.created_at,
		m.position, m.department, m.avatar, m.avatar_thumb, m.access_expiry_date, ` + centraldb.HoldingTimezoneSQL

// memberSelect reads one membership of an active user in an active Holding. Disabled
// memberships are returned with IsAccessDisabled so callers can explain the denial.
const memberSelect = `SELECT ` + memberColumns + `
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
		expiryDate     sql.NullTime
		timezone       string
	)
	err := row.Scan(&member.MembershipUID, &member.UserUID, &member.Username, &role, &permissionSets, &accessScopes, &isActive,
		&member.IsFavorite, &lastAccessedAt, &member.CreatedAt,
		&member.Position, &member.Department, &member.Avatar, &member.AvatarThumb, &expiryDate, &timezone)
	if err != nil {
		return models.ShopUser{}, err
	}
	member.AccessExpiryDate = centraldb.AccessExpiryInstant(expiryDate, timezone)
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
	// A manager without explicit scopes is holding-wide. Report it as the holding rule (not
	// a snapshot of today's companies) in the detail AND the list, so the editor and the
	// access audit agree and a re-save keeps it holding-wide.
	if len(member.AccessScopes) == 0 && (member.Role == models.ROLE_OWNER || member.Role == models.ROLE_ADMIN) {
		member.AccessScopes = []models.AccessScope{{ScopeType: "holding"}}
	}
	return member, nil
}

// loginAccount is the users row a membership save may touch.
type loginAccount struct {
	id       string
	username string
	email    string
	fullName string
	// claimed: somebody has signed in with the account (it has a password or a linked login).
	claimed bool
	// elsewhere / member: it is a member of another / of this business group.
	elsewhere bool
	member    bool
	// removedHere: this business group removed its membership before (organization_audits
	// member_removed, written by Delete) — re-adding it brings the same account back.
	removedHere bool
}

// managedHere: the usercode and email of an account may be changed from this business
// group's settings only while nobody has signed in with it and no other group uses it.
// Otherwise the email is someone's identity (and the key a first Google login links by),
// and another group's admin must not be able to repoint it.
func (a loginAccount) managedHere() bool {
	return !a.claimed && !a.elsewhere
}

const loginAccountSelect = `SELECT u.id::text, u.username, COALESCE(u.email, ''), u.full_name,
		(u.password_hash <> '' OR EXISTS (SELECT 1 FROM user_identities i WHERE i.user_id = u.id)),
		EXISTS (SELECT 1 FROM holding_members o WHERE o.user_id = u.id AND o.holding_code <> $1),
		EXISTS (SELECT 1 FROM holding_members o WHERE o.user_id = u.id AND o.holding_code = $1),
		EXISTS (SELECT 1 FROM organization_audits a WHERE a.holding_code = $1 AND a.target_type = '` + memberAuditTarget + `'
			AND a.target_code = u.id::text AND a.action = '` + memberRemovedAction + `')
	FROM users u WHERE `

// findLoginAccounts locks and returns the accounts matching filter ($2 = identity).
func findLoginAccounts(ctx context.Context, db dbtx, holdingCode string, filter string, identity string) ([]loginAccount, error) {
	query := loginAccountSelect + filter + ` ORDER BY u.username`
	// Lock first, then read the flags in a fresh statement: a statement that waited for a row
	// lock still evaluates its sub-queries on the snapshot it started with, so a Google login
	// that claimed the account meanwhile would be missed (and the account wrongly editable).
	if _, err := db.ExecContext(ctx, query+` FOR UPDATE OF u`, holdingCode, identity); err != nil {
		return nil, err
	}
	rows, err := db.QueryContext(ctx, query, holdingCode, identity)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	accounts := []loginAccount{}
	for rows.Next() {
		var account loginAccount
		if err := rows.Scan(&account.id, &account.username, &account.email, &account.fullName,
			&account.claimed, &account.elsewhere, &account.member, &account.removedHere); err != nil {
			return nil, err
		}
		accounts = append(accounts, account)
	}
	return accounts, rows.Err()
}

// membershipFields are the per-membership values of a save, ready as SQL arguments.
type membershipFields struct {
	role, permissionSets, accessScopes string
	isActive                           bool
	position, department               string
	avatar, avatarThumb                interface{} // nil keeps the stored picture
	accessExpiryDate                   string      // "YYYY-MM-DD" or "" (no expiry)
}

func newMembershipFields(req *models.UserRoleRequest) (membershipFields, error) {
	fields := membershipFields{
		role:       models.RoleText(req.Role),
		isActive:   !req.IsAccessDisabled,
		position:   strings.TrimSpace(req.Position),
		department: strings.TrimSpace(req.Department),
	}
	if expiry := strings.TrimSpace(req.AccessExpiryDate); expiry != "" {
		day, err := models.NormalizeAccessExpiryDate(json.RawMessage(`"` + expiry + `"`))
		if err != nil {
			return membershipFields{}, err
		}
		fields.accessExpiryDate = day
	}
	permissionSets := req.PermissionSets
	if permissionSets == nil {
		permissionSets = []string{}
	}
	permsJSON, err := json.Marshal(permissionSets)
	if err != nil {
		return membershipFields{}, err
	}
	fields.permissionSets = string(permsJSON)
	fields.accessScopes = "[]"
	if len(req.AccessScopes) > 0 {
		scopesJSON, err := json.Marshal(req.AccessScopes)
		if err != nil {
			return membershipFields{}, err
		}
		fields.accessScopes = string(scopesJSON)
	}
	if req.Avatar != nil {
		fields.avatar = strings.TrimSpace(*req.Avatar)
	}
	if req.AvatarThumb != nil {
		fields.avatarThumb = strings.TrimSpace(*req.AvatarThumb)
	}
	return fields, nil
}

// SaveFullProfile writes one membership in a single transaction.
//
// ADD (no editusername and no useruid): a new username gets a new login account; a username
// that is already a member here is errMemberAlreadyExists; a username that already has an
// account elsewhere is errUsernameTaken unless req.AddExistingUser explicitly joins that
// account and no other business group uses it. EDIT (editusername or useruid): updates that member only, never creates one.
func (r *ShopUserPostgresRepository) SaveFullProfile(ctx context.Context, holdingCode string, req *models.UserRoleRequest) error {
	holdingCode = strings.TrimSpace(holdingCode)
	if holdingCode == "" || req == nil {
		return fmt.Errorf("%w: holdingCode and request required", errMemberRequestInvalid)
	}
	username := strings.TrimSpace(req.Username)
	userUID := strings.TrimSpace(req.UserUID)
	editID := strings.TrimSpace(req.EditUsername)
	if username == "" && userUID == "" {
		return fmt.Errorf("%w: username required", errMemberRequestInvalid)
	}
	if userUID != "" {
		if _, err := uuid.Parse(userUID); err != nil {
			return fmt.Errorf("%w: invalid useruid", errMemberRequestInvalid)
		}
	}
	fields, err := newMembershipFields(req)
	if err != nil {
		return err
	}
	return runInTx(ctx, r.db, func(txCtx context.Context) error {
		db := conn(txCtx, r.db)
		if userUID == "" && editID == "" {
			return addMember(txCtx, db, holdingCode, username, req, fields)
		}
		return editMember(txCtx, db, holdingCode, userUID, editID, username, req, fields)
	})
}

func addMember(ctx context.Context, db dbtx, holdingCode string, username string, req *models.UserRoleRequest, fields membershipFields) error {
	// Two admins adding the same new usercode at once must not both create it.
	if _, err := db.ExecContext(ctx, `SELECT pg_advisory_xact_lock(hashtextextended($1, 0))`, "login-account:"+strings.ToLower(username)); err != nil {
		return err
	}
	accounts, err := findLoginAccounts(ctx, db, holdingCode, `LOWER(u.username) = LOWER($2)`, username)
	if err != nil {
		return err
	}
	for _, account := range accounts {
		if account.member {
			return errMemberAlreadyExists
		}
	}
	var userID string
	switch {
	// A login account this business group removed earlier comes back as the same account
	// (reactivated) without the addexistinguser confirmation: it is not another group's account.
	// Only while nobody has signed in with it: a claimed account (password or Google login) would
	// hand the new role to whoever still holds that sign-in, so the admin must confirm the attach
	// (errLoginExists → dialog). One still used by another group stays refused below (review 2026-09-24).
	case len(accounts) == 1 && accounts[0].removedHere && !accounts[0].elsewhere && !accounts[0].claimed:
		if err := updateLoginAccount(ctx, db, accounts[0], "", req.Email, req.UserProfileName); err != nil {
			return err
		}
		userID = accounts[0].id
	case len(accounts) == 0:
		err = db.QueryRowContext(ctx, `
			INSERT INTO users (username, password_hash, email, full_name, is_active)
			VALUES ($1, '', NULLIF($2, ''), $3, true) RETURNING id::text`,
			username, strings.TrimSpace(req.Email), strings.TrimSpace(req.UserProfileName)).Scan(&userID)
		if centraldb.IsUniqueViolation(err) {
			return errUsernameTaken
		}
		if err != nil {
			return err
		}
	// addexistinguser joins only an account no other business group uses. Attaching another
	// group's account by its usercode alone let any group admin (or the public Demo owner) read
	// that account's email/name and lock its home group out of fixing its usercode/email
	// (managedHere turns false) — cross-group access needs an owner-accepted invitation instead.
	// Same answer with or without the flag.
	case len(accounts) > 1 || accounts[0].elsewhere:
		return errUsernameTaken
	case !req.AddExistingUser:
		return errLoginExists
	default:
		if err := updateLoginAccount(ctx, db, accounts[0], "", req.Email, req.UserProfileName); err != nil {
			return err
		}
		userID = accounts[0].id
	}
	_, err = db.ExecContext(ctx, `
		INSERT INTO holding_members (holding_code, user_id, role, permission_sets, access_scopes, is_active,
			position, department, avatar, avatar_thumb, access_expiry_date)
		VALUES ($1, $2::uuid, $3, $4, $5, $6, $7, $8, COALESCE($9, ''), COALESCE($10, ''), NULLIF($11, '')::date)`,
		holdingCode, userID, fields.role, fields.permissionSets, fields.accessScopes, fields.isActive,
		fields.position, fields.department, fields.avatar, fields.avatarThumb, fields.accessExpiryDate)
	if centraldb.IsUniqueViolation(err) {
		return errMemberAlreadyExists
	}
	return err
}

func editMember(ctx context.Context, db dbtx, holdingCode, userUID, editID, username string, req *models.UserRoleRequest, fields membershipFields) error {
	filter, identity := `u.id::text = $2`, userUID
	if userUID == "" {
		filter, identity = `(LOWER(u.username) = LOWER($2) OR u.id::text = $2)`, editID
	}
	accounts, err := findLoginAccounts(ctx, db, holdingCode, filter, identity)
	if err != nil {
		return err
	}
	members := []loginAccount{}
	for _, account := range accounts {
		if account.member {
			members = append(members, account)
		}
	}
	if len(members) != 1 {
		return errSaveTargetNotFound
	}
	account := members[0]
	rename := ""
	if username != "" && !strings.EqualFold(username, account.username) {
		rename = username
	}
	if err := updateLoginAccount(ctx, db, account, rename, req.Email, req.UserProfileName); err != nil {
		return err
	}
	result, err := db.ExecContext(ctx, `
		UPDATE holding_members SET
			role = $3,
			permission_sets = $4,
			access_scopes = $5,
			is_active = $6,
			position = $7,
			department = $8,
			avatar = COALESCE($9, avatar),
			avatar_thumb = COALESCE($10, avatar_thumb),
			access_expiry_date = NULLIF($11, '')::date
		WHERE holding_code = $1 AND user_id = $2::uuid`,
		holdingCode, account.id, fields.role, fields.permissionSets, fields.accessScopes, fields.isActive,
		fields.position, fields.department, fields.avatar, fields.avatarThumb, fields.accessExpiryDate)
	if err != nil {
		return err
	}
	if affected, err := result.RowsAffected(); err != nil || affected != 1 {
		if err != nil {
			return err
		}
		return errSaveTargetNotFound
	}
	return nil
}

// updateLoginAccount applies the account-level values of a save: a new usercode, the email
// and the display name. Usercode and email change only on an account managedHere; an
// empty email/name keeps the stored one.
func updateLoginAccount(ctx context.Context, db dbtx, account loginAccount, rename string, email string, fullName string) error {
	email = strings.TrimSpace(email)
	fullName = strings.TrimSpace(fullName)
	if rename != "" && !account.managedHere() {
		return errUserCodeLocked
	}
	setEmail := ""
	if email != "" && !strings.EqualFold(email, account.email) {
		if !account.managedHere() {
			return errUserEmailLocked
		}
		setEmail = email
	}
	if rename == "" && setEmail == "" && (fullName == "" || fullName == account.fullName) {
		return nil
	}
	if rename != "" {
		var taken bool
		if err := db.QueryRowContext(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE LOWER(username) = LOWER($1) AND id::text <> $2)`,
			rename, account.id).Scan(&taken); err != nil {
			return err
		}
		if taken {
			return errUsernameTaken
		}
	}
	_, err := db.ExecContext(ctx, `
		UPDATE users SET
			username = COALESCE(NULLIF($2, ''), username),
			email = COALESCE(NULLIF($3, ''), email),
			full_name = COALESCE(NULLIF($4, ''), full_name),
			updated_at = now()
		WHERE id::text = $1`, account.id, rename, setEmail, fullName)
	if centraldb.IsUniqueViolation(err) {
		return errUsernameTaken
	}
	return err
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

// Member removal is recorded in organization_audits (append-only): who removed whom, and the
// evidence addMember uses to bring the same login account back when it is re-added.
const (
	memberAuditTarget   = "holding_member"
	memberRemovedAction = "member_removed"
)

// Delete removes the membership and records the removal (actorUID = the admin) in one statement.
func (r *ShopUserPostgresRepository) Delete(ctx context.Context, holdingCode string, username string, actorUID string) error {
	_, err := conn(ctx, r.db).ExecContext(ctx, `
		WITH removed AS (
			DELETE FROM holding_members m
			USING users u
			WHERE m.user_id = u.id AND m.holding_code = $1 AND (LOWER(u.username) = LOWER($2) OR u.id::text = $2)
			RETURNING m.holding_code, m.user_id, m.role, u.username)
		INSERT INTO organization_audits (holding_code, action, target_type, target_code, actor_uid, before_state)
		SELECT holding_code, '`+memberRemovedAction+`', '`+memberAuditTarget+`', user_id::text, $3,
			jsonb_build_object('username', username, 'role', role)
		FROM removed`,
		strings.TrimSpace(holdingCode), strings.TrimSpace(username), strings.TrimSpace(actorUID))
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
	query := `SELECT ` + memberColumns + `
		FROM holding_members m
		JOIN users u ON u.id = m.user_id
		JOIN holdings h ON h.code = m.holding_code
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

// ResolveCompanyUID maps an ACTIVE company code to its UID; in PostgreSQL the UID is the
// code. A holding-wide scope covers any company, so closed companies must fail here.
func (r *ShopUserPostgresRepository) ResolveCompanyUID(ctx context.Context, holdingCode string, businessCode string) (string, error) {
	var code string
	err := conn(ctx, r.db).QueryRowContext(ctx, `SELECT code FROM companies WHERE holding_code = $1 AND UPPER(code) = UPPER($2) AND is_active = true`,
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
