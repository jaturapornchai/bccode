package mcptoken

import (
	"context"
	"database/sql"
	_ "embed"
	"encoding/json"
	"github.com/lib/pq"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	authmodels "smlcloudplatform/internal/authentication/models"
	"smlcloudplatform/internal/goapi/language"
	"smlcloudplatform/internal/goapi/mypg"
	"smlcloudplatform/pkg/microservice"
	"smlcloudplatform/pkg/microservice/models"
)

//go:embed schema.sql
var schema string

type Http struct {
	ms      *microservice.Microservice
	connect func(string) (*sql.DB, error)
}

func NewHttp(ms *microservice.Microservice) *Http {
	return &Http{ms: ms, connect: mypg.PgSqlFastConnect}
}
func (h *Http) RegisterHttp() {
	h.ms.GET("/mcp-tokens", h.list)
	h.ms.GET("/mcp-tokens/companies", h.companies)
	h.ms.POST("/mcp-tokens", h.create)
	h.ms.POST("/mcp-tokens/:id/revoke", h.revoke)
}
func respond(r microservice.IContext, status int, data interface{}) error {
	r.ResponseWriter().Header().Set("Cache-Control", "no-store")
	if status == 200 {
		r.Response(status, map[string]interface{}{"success": true, "data": data})
	} else {
		r.Response(status, map[string]interface{}{"success": false, "message": data})
	}
	return nil
}

// requestLanguage is the caller's language for user-facing messages (query lang, then the
// Accept-Language the BFF forwards).
func requestLanguage(r microservice.IContext) string {
	if lang := strings.TrimSpace(r.Request().URL.Query().Get("lang")); lang != "" {
		return lang
	}
	return r.Request().Header.Get("Accept-Language")
}

// admin returns the control database and the caller's OWNER/ADMIN membership with scopes expanded.
func (h *Http) admin(ctx context.Context, u models.UserInfo) (*sql.DB, authmodels.ShopUser, error) {
	if !holdingPattern.MatchString(u.HoldingCode) {
		return nil, authmodels.ShopUser{}, ErrDenied
	}
	db, err := h.connect(ControlDatabase)
	if err != nil || db == nil {
		return nil, authmodels.ShopUser{}, ErrDenied
	}
	issuer, err := authorizeHolding(ctx, db, u)
	if err != nil {
		return nil, authmodels.ShopUser{}, ErrDenied
	}
	// Serialize first-time initialization across simultaneous admin requests.
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return nil, authmodels.ShopUser{}, err
	}
	defer tx.Rollback()
	if _, err = tx.ExecContext(ctx, `SELECT pg_advisory_xact_lock(194857231)`); err != nil {
		return nil, authmodels.ShopUser{}, err
	}
	if _, err = tx.ExecContext(ctx, schema); err != nil {
		return nil, authmodels.ShopUser{}, err
	}
	if err = tx.Commit(); err != nil {
		return nil, authmodels.ShopUser{}, err
	}
	return db, issuer, nil
}

func (h *Http) list(r microservice.IContext) error {
	ctx, cancel := context.WithTimeout(r.Request().Context(), 15*time.Second)
	defer cancel()
	u := r.UserInfo()
	db, _, err := h.admin(ctx, u)
	if err != nil {
		return respond(r, 403, ErrDenied.Error())
	}
	rows, err := db.QueryContext(ctx, `SELECT t.id,t.name,t.kind,t.mode,t.holding_code,t.company_code,t.branch_code,t.created_by,t.created_at,t.expires_at,t.revoked_at,t.last_used_at, ARRAY(SELECT tc.company_code FROM mcp_token_companies tc WHERE tc.token_id=t.id AND tc.holding_code=t.holding_code ORDER BY tc.company_code) FROM mcp_access_tokens t WHERE t.holding_code=$1 ORDER BY t.created_at DESC LIMIT 500`, u.HoldingCode)
	if err != nil {
		return respond(r, 503, "ไม่สามารถอ่าน token ได้")
	}
	defer rows.Close()
	result := []Metadata{}
	for rows.Next() {
		var m Metadata
		if err = rows.Scan(&m.ID, &m.Name, &m.Kind, &m.Mode, &m.HoldingCode, &m.CompanyCode, &m.BranchCode, &m.CreatedBy, &m.CreatedAt, &m.ExpiresAt, &m.RevokedAt, &m.LastUsedAt, pq.Array(&m.CompanyCodes)); err != nil {
			return respond(r, 503, "ไม่สามารถอ่าน token ได้")
		}
		result = append(result, m)
	}
	if rows.Err() != nil {
		return respond(r, 503, "ไม่สามารถอ่าน token ได้")
	}
	return respond(r, 200, result)
}

type createInput struct {
	CompanyCodes []string  `json:"companyCodes"`
	Kind         string    `json:"kind"`
	Name         string    `json:"name"`
	Mode         string    `json:"mode"`
	ExpiresAt    time.Time `json:"expiresAt"`
}

func validateInput(in createInput, now time.Time) bool {
	return len(in.CompanyCodes) > 0 && len(in.CompanyCodes) <= 100 && (in.Kind == "api" || in.Kind == "mcp") && utf8.RuneCountInString(in.Name) >= 1 && utf8.RuneCountInString(in.Name) <= 100 && (in.Mode == "readonly" || in.Mode == "readwrite") && in.ExpiresAt.After(now) && !in.ExpiresAt.After(now.AddDate(1, 0, 0))
}
func (h *Http) create(r microservice.IContext) error {
	ctx, cancel := context.WithTimeout(r.Request().Context(), 15*time.Second)
	defer cancel()
	u := r.UserInfo()
	db, issuer, err := h.admin(ctx, u)
	if err != nil {
		return respond(r, 403, ErrDenied.Error())
	}
	var in createInput
	dec := json.NewDecoder(http.MaxBytesReader(r.ResponseWriter(), r.Request().Body, 16384))
	dec.DisallowUnknownFields()
	if dec.Decode(&in) != nil || dec.Decode(&struct{}{}) != io.EOF {
		return respond(r, 400, "ข้อมูล token ไม่ถูกต้อง")
	}
	in.Name = strings.TrimSpace(in.Name)
	now := time.Now().UTC()
	if !validateInput(in, now) {
		return respond(r, 400, "ระบุชื่อ สิทธิ์ และวันหมดอายุในอนาคตไม่เกินหนึ่งปี")
	}
	id, raw, hash, err := generate(u.HoldingCode, in.Kind)
	if err != nil {
		return respond(r, 503, "ไม่สามารถสร้าง token ได้")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return respond(r, 503, "ไม่สามารถสร้าง token ได้")
	}
	defer tx.Rollback()
	if validateCompanies(ctx, tx, u.HoldingCode, in.CompanyCodes) != nil {
		return respond(r, 400, "เลือกบริษัทที่เปิดใช้งานใน Holding นี้อย่างน้อยหนึ่งบริษัท")
	}
	// A token never reaches further than its issuer: an ADMIN limited to some companies may
	// allow only companies they can open company-wide. 400, not 403: the token screen
	// treats 403 as lost administrator access and would discard the form.
	for _, code := range in.CompanyCodes {
		if !withinScope(issuer.AccessScopes, code, "") {
			return respond(r, 400, language.Text("mcp_err_company_outside_scope", requestLanguage(r)))
		}
	}
	_, err = tx.ExecContext(ctx, `INSERT INTO mcp_access_tokens(id,holding_code,company_code,branch_code,name,kind,mode,token_hash,created_by,creator_username,created_at,expires_at) VALUES($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12)`, id, u.HoldingCode, "", "", in.Name, in.Kind, in.Mode, hash, u.UID, u.Username, now, in.ExpiresAt)
	if err == nil {
		_, err = tx.ExecContext(ctx, `INSERT INTO mcp_token_companies(token_id,holding_code,company_code) SELECT $1,$2,unnest($3::text[])`, id, u.HoldingCode, pq.Array(in.CompanyCodes))
	}
	if err == nil {
		_, err = tx.ExecContext(ctx, `INSERT INTO mcp_token_audit(token_id,holding_code,company_code,branch_code,actor,action) VALUES($1,$2,$3,$4,$5,'create')`, id, u.HoldingCode, "", "", u.UID)
	}
	if err != nil || tx.Commit() != nil {
		return respond(r, 503, "ไม่สามารถสร้าง token ได้")
	}
	result := struct {
		Metadata
		Token string `json:"token"`
	}{Metadata: Metadata{ID: id, Name: in.Name, Kind: in.Kind, Mode: in.Mode, HoldingCode: u.HoldingCode, CompanyCodes: in.CompanyCodes, CreatedBy: u.UID, CreatedAt: now, ExpiresAt: in.ExpiresAt}, Token: raw}
	return respond(r, 200, result)
}
func (h *Http) revoke(r microservice.IContext) error {
	ctx, cancel := context.WithTimeout(r.Request().Context(), 15*time.Second)
	defer cancel()
	u := r.UserInfo()
	db, _, err := h.admin(ctx, u)
	if err != nil {
		return respond(r, 403, ErrDenied.Error())
	}
	id := r.Param("id")
	if !idPattern.MatchString(id) {
		return respond(r, 404, "ไม่พบ token")
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return respond(r, 503, "ไม่สามารถเพิกถอน token ได้")
	}
	defer tx.Rollback()
	var revoked *time.Time
	err = tx.QueryRowContext(ctx, `SELECT revoked_at FROM mcp_access_tokens WHERE holding_code=$1 AND id=$2 FOR UPDATE`, u.HoldingCode, id).Scan(&revoked)
	if err == sql.ErrNoRows {
		return respond(r, 404, "ไม่พบ token")
	}
	if err != nil {
		return respond(r, 503, "ไม่สามารถเพิกถอน token ได้")
	}
	if revoked == nil {
		_, err = tx.ExecContext(ctx, `UPDATE mcp_access_tokens SET revoked_at=now() WHERE id=$1`, id)
		if err == nil {
			_, err = tx.ExecContext(ctx, `INSERT INTO mcp_token_audit(token_id,holding_code,company_code,branch_code,actor,action) VALUES($1,$2,$3,$4,$5,'revoke')`, id, u.HoldingCode, "", "", u.UID)
		}
	}
	if err != nil || tx.Commit() != nil {
		return respond(r, 503, "ไม่สามารถเพิกถอน token ได้")
	}
	return respond(r, 200, map[string]string{"id": id})
}

func (h *Http) companies(r microservice.IContext) error {
	ctx, cancel := context.WithTimeout(r.Request().Context(), 15*time.Second)
	defer cancel()
	u := r.UserInfo()
	db, issuer, err := h.admin(ctx, u)
	if err != nil {
		return respond(r, 403, ErrDenied.Error())
	}
	rows, err := db.QueryContext(ctx, `SELECT code,name FROM companies WHERE holding_code=$1 AND is_active=true ORDER BY code`, u.HoldingCode)
	if err != nil {
		return respond(r, 503, "ไม่สามารถอ่านบริษัทได้")
	}
	defer rows.Close()
	type option struct {
		Code string `json:"code"`
		Name string `json:"name"`
	}
	result := []option{}
	for rows.Next() {
		var item option
		if rows.Scan(&item.Code, &item.Name) != nil {
			return respond(r, 503, "ไม่สามารถอ่านบริษัทได้")
		}
		// Offer only companies this administrator may allow; create enforces the same rule.
		if withinScope(issuer.AccessScopes, item.Code, "") {
			result = append(result, item)
		}
	}
	if rows.Err() != nil {
		return respond(r, 503, "ไม่สามารถอ่านบริษัทได้")
	}
	return respond(r, 200, result)
}
