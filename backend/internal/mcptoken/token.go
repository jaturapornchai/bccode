// Package mcptoken issues scoped, revocable credentials with separate API and MCP audiences for GL.
package mcptoken

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"github.com/lib/pq"
	"regexp"
	"strings"
	"time"

	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/pkg/microservice/models"
)

var ErrDenied = errors.New("ไม่มีสิทธิ์ใช้งาน token")

// ControlDatabase contains identities, Holding memberships and token grants.
const ControlDatabase = "bcai_projection"

var holdingPattern = regexp.MustCompile(`^[A-Za-z0-9_]{1,63}$`)
var idPattern = regexp.MustCompile(`^[0-9a-f]{32}$`)

type Principal struct {
	CompanyCodes []string
	ID           string
	Kind         string
	Mode         string
	User         models.UserInfo
}
type Metadata struct {
	CompanyCodes []string   `json:"companyCodes"`
	Kind         string     `json:"kind"`
	ID           string     `json:"id"`
	Name         string     `json:"name"`
	Mode         string     `json:"mode"`
	HoldingCode  string     `json:"holdingCode"`
	CompanyCode  string     `json:"companyCode"`
	BranchCode   string     `json:"branchCode"`
	CreatedBy    string     `json:"createdBy"`
	CreatedAt    time.Time  `json:"createdAt"`
	ExpiresAt    time.Time  `json:"expiresAt"`
	RevokedAt    *time.Time `json:"revokedAt"`
	LastUsedAt   *time.Time `json:"lastUsedAt"`
}

func generate(holding, kind string) (id, raw string, hash []byte, err error) {
	if !holdingPattern.MatchString(holding) || (kind != "api" && kind != "mcp") {
		return "", "", nil, ErrDenied
	}
	var identifier [16]byte
	var secret [32]byte
	if _, err = rand.Read(identifier[:]); err != nil {
		return
	}
	if _, err = rand.Read(secret[:]); err != nil {
		return
	}
	id = hex.EncodeToString(identifier[:])
	raw = "bcai" + kind + "_" + base64.RawURLEncoding.EncodeToString([]byte(holding)) + "." + id + "." + base64.RawURLEncoding.EncodeToString(secret[:])
	sum := sha256.Sum256([]byte(raw))
	hash = sum[:]
	return
}

func parse(raw, kind string) (holding, id string, err error) {
	if (kind != "api" && kind != "mcp") || len(raw) > 200 || !strings.HasPrefix(raw, "bcai"+kind+"_") {
		return "", "", ErrDenied
	}
	parts := strings.Split(strings.TrimPrefix(raw, "bcai"+kind+"_"), ".")
	if len(parts) != 3 || !idPattern.MatchString(parts[1]) {
		return "", "", ErrDenied
	}
	decoded, e := base64.RawURLEncoding.Strict().DecodeString(parts[0])
	if e != nil || !holdingPattern.Match(decoded) {
		return "", "", ErrDenied
	}
	secret, e := base64.RawURLEncoding.Strict().DecodeString(parts[2])
	if e != nil || len(secret) != 32 {
		return "", "", ErrDenied
	}
	return string(decoded), parts[1], nil
}

type querier interface {
	QueryRowContext(context.Context, string, ...interface{}) *sql.Row
}

// Resolve the role and tenant from database state, never from a session role claim.
func authorizeHolding(ctx context.Context, db querier, u models.UserInfo) error {
	if !holdingPattern.MatchString(u.HoldingCode) || u.UID == "" {
		return ErrDenied
	}
	var role string
	err := db.QueryRowContext(ctx, `SELECT m.role FROM holding_members m
 JOIN users u ON u.id=m.user_id AND u.is_active=true
 JOIN holdings h ON h.code=m.holding_code AND h.is_active=true
 WHERE m.holding_code=$1 AND u.id::text=$2 AND m.is_active=true`, u.HoldingCode, u.UID).Scan(&role)
	if err != nil || !(strings.EqualFold(role, "OWNER") || strings.EqualFold(role, "ADMIN")) {
		return ErrDenied
	}

	return nil
}

// Authenticate accepts MCP credentials only, never caller-provided tenant scope or schema creation.
func Authenticate(ctx context.Context, raw string) (Principal, error) {
	return authenticateAudience(ctx, raw, "mcp", mypg.PgSqlFastConnect, time.Now().UTC())
}

// AuthenticateAPI accepts API credentials only; MCP credentials cannot interoperate.
func AuthenticateAPI(ctx context.Context, raw string) (Principal, error) {
	return authenticateAudience(ctx, raw, "api", mypg.PgSqlFastConnect, time.Now().UTC())
}
func authenticateAudience(ctx context.Context, raw, kind string, connect func(string) (*sql.DB, error), now time.Time) (Principal, error) {
	var p Principal
	holding, id, err := parse(raw, kind)
	if err != nil {
		return p, ErrDenied
	}
	db, err := connect(ControlDatabase)
	if err != nil || db == nil {
		return p, ErrDenied
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return p, ErrDenied
	}
	defer tx.Rollback()
	var hash []byte
	var expiry time.Time
	var revoked sql.NullTime
	p.ID = id
	p.User.HoldingCode = holding
	err = tx.QueryRowContext(ctx, `SELECT token_hash,kind,mode,company_code,branch_code,created_by,creator_username,expires_at,revoked_at FROM mcp_access_tokens WHERE holding_code=$1 AND id=$2 FOR UPDATE`, holding, id).Scan(&hash, &p.Kind, &p.Mode, &p.User.BusinessCode, &p.User.BranchUID, &p.User.UID, &p.User.Username, &expiry, &revoked)
	sum := sha256.Sum256([]byte(raw))
	if err != nil || p.Kind != kind || subtle.ConstantTimeCompare(hash, sum[:]) != 1 || !now.Before(expiry) || revoked.Valid || (p.Mode != "readonly" && p.Mode != "readwrite") {
		return Principal{}, ErrDenied
	}
	if authorizeHolding(ctx, tx, p.User) != nil {
		return Principal{}, ErrDenied
	}
	// Legacy tokens retain their original company and branch even after migration.
	if p.User.BusinessCode != "" {
		if activeCompany(ctx, tx, p.User, p.User.BusinessCode) != nil {
			return Principal{}, ErrDenied
		}
		p.CompanyCodes = []string{p.User.BusinessCode}
	} else {
		rows, e := tx.QueryContext(ctx, `SELECT tc.company_code FROM mcp_token_companies tc JOIN companies c ON c.holding_code=tc.holding_code AND c.code=tc.company_code AND c.is_active=true WHERE tc.token_id=$1 AND tc.holding_code=$2 ORDER BY tc.company_code`, p.ID, p.User.HoldingCode)
		if e != nil {
			return Principal{}, ErrDenied
		}
		for rows.Next() {
			var code string
			if rows.Scan(&code) != nil {
				rows.Close()
				return Principal{}, ErrDenied
			}
			p.CompanyCodes = append(p.CompanyCodes, code)
		}
		e = rows.Err()
		rows.Close()
		if e != nil || len(p.CompanyCodes) == 0 {
			return Principal{}, ErrDenied
		}
	}
	if _, err = tx.ExecContext(ctx, `UPDATE mcp_access_tokens SET last_used_at=$3 WHERE holding_code=$1 AND id=$2`, holding, id, now); err != nil {
		return Principal{}, ErrDenied
	}
	if _, err = tx.ExecContext(ctx, `INSERT INTO mcp_token_audit(token_id,holding_code,company_code,branch_code,actor,action) VALUES($1,$2,$3,$4,$5,'use')`, id, holding, p.User.BusinessCode, p.User.BranchUID, p.User.UID); err != nil {
		return Principal{}, ErrDenied
	}
	if tx.Commit() != nil {
		return Principal{}, ErrDenied
	}
	return p, nil
}

func activeCompany(ctx context.Context, db querier, u models.UserInfo, company string) error {
	var code string
	if db.QueryRowContext(ctx, `SELECT code FROM companies WHERE holding_code=$1 AND code=$2 AND is_active=true`, u.HoldingCode, company).Scan(&code) != nil {
		return ErrDenied
	}
	if u.BranchUID != "" {
		if db.QueryRowContext(ctx, `SELECT code FROM branches WHERE holding_code=$1 AND company_code=$2 AND code=$3 AND is_active=true`, u.HoldingCode, company, u.BranchUID).Scan(&code) != nil {
			return ErrDenied
		}
	}
	return nil
}

// ResolveCompany chooses one explicit company without widening the stored grant.
// An omitted company is unambiguous only for a token with one allowed company.
func ResolveCompany(ctx context.Context, p Principal, companyCode string) (Principal, error) {
	return resolveCompany(ctx, p, companyCode, mypg.PgSqlFastConnect)
}
func resolveCompany(ctx context.Context, p Principal, companyCode string, connect func(string) (*sql.DB, error)) (Principal, error) {
	if companyCode == "" && len(p.CompanyCodes) == 1 {
		companyCode = p.CompanyCodes[0]
	}
	allowed := false
	for _, code := range p.CompanyCodes {
		if code == companyCode {
			allowed = true
			break
		}
	}
	if !allowed || companyCode == "" || !holdingPattern.MatchString(p.User.HoldingCode) {
		return Principal{}, ErrDenied
	}
	// A legacy company restriction is never replaced by the new allow-list.
	if p.User.BusinessCode != "" && p.User.BusinessCode != companyCode {
		return Principal{}, ErrDenied
	}
	db, err := connect(ControlDatabase)
	if err != nil || db == nil {
		return Principal{}, ErrDenied
	}
	if activeCompany(ctx, db, p.User, companyCode) != nil {
		return Principal{}, ErrDenied
	}
	p.User.BusinessCode = companyCode
	return p, nil
}

func validateCompanies(ctx context.Context, tx *sql.Tx, holding string, codes []string) error {
	if len(codes) == 0 || len(codes) > 100 {
		return ErrDenied
	}
	seen := map[string]bool{}
	for _, code := range codes {
		if code == "" || strings.TrimSpace(code) != code || len(code) > 100 || seen[code] {
			return ErrDenied
		}
		seen[code] = true
	}
	var count int
	if tx.QueryRowContext(ctx, `SELECT count(*) FROM companies WHERE holding_code=$1 AND code=ANY($2) AND is_active=true`, holding, pq.Array(codes)).Scan(&count) != nil || count != len(codes) {
		return ErrDenied
	}
	return nil
}
