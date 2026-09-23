package organization

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"strings"
	"time"
)

// ErrStatusChangeConflict means the row no longer has the status the caller read.
var ErrStatusChangeConflict = errors.New("organization status changed concurrently")

// Audit is one append-only row of organization_audits.
type Audit struct {
	HoldingCode string
	Action      string
	TargetType  string
	TargetCode  string
	CompanyCode string
	ActorUID    string
	Reason      string
	Before      interface{}
	After       interface{}
	OccurredAt  time.Time
}

// Execer is satisfied by *sql.DB and *sql.Tx.
type Execer interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
}

// OrganizationCreateResponse is the create response body ({"entity": ...}).
type OrganizationCreateResponse struct {
	Entity interface{} `json:"entity"`
}

// RecordAudit appends an organization change inside the caller's transaction.
func RecordAudit(ctx context.Context, tx Execer, audit Audit) error {
	if strings.TrimSpace(audit.HoldingCode) == "" || strings.TrimSpace(audit.Action) == "" || strings.TrimSpace(audit.TargetCode) == "" {
		return errors.New("organization audit identity is incomplete")
	}
	before, err := optionalJSON(audit.Before)
	if err != nil {
		return err
	}
	after, err := optionalJSON(audit.After)
	if err != nil {
		return err
	}
	occurredAt := audit.OccurredAt
	if occurredAt.IsZero() {
		occurredAt = time.Now()
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO organization_audits (holding_code, action, target_type, target_code, company_code, actor_uid, reason, before_state, after_state, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
		audit.HoldingCode, audit.Action, audit.TargetType, audit.TargetCode, audit.CompanyCode,
		strings.TrimSpace(audit.ActorUID), strings.TrimSpace(audit.Reason), before, after, occurredAt.UTC())
	return err
}

func optionalJSON(value interface{}) (interface{}, error) {
	if value == nil {
		return nil, nil
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}
	return string(raw), nil
}
