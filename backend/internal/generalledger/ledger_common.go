package generalledger

import (
	"crypto/sha256"
	"encoding/hex"
	"time"
)

type Result struct {
	CreatedJournals   int    `json:"createdjournals"`
	ID                string `json:"id"`
	Version           int64  `json:"version"`
	Sequence          int64  `json:"sequence"`
	ProjectionPending bool   `json:"projectionpending"`
	// Lines answers budgets action "spread" (nothing is saved).
	Lines []BudgetLine `json:"lines,omitempty"`
}

func digest(value []byte) string { sum := sha256.Sum256(value); return hex.EncodeToString(sum[:]) }
func scopeID(s Scope) string     { return digest([]byte(s.Holding + "\x00" + s.Company)) }

func checkVersion(cmd Command, old Identity) error {
	if old.IsDeleted {
		return ErrNotFound
	}
	if old.Version != cmd.Version || cmd.Version < 1 {
		return userError(CodeStaleVersion, "รายการนี้เปลี่ยนไปแล้ว กรุณาโหลดข้อมูลล่าสุดก่อนบันทึก")
	}
	return nil
}

func single(c Change, err error) ([]Change, error) {
	if err != nil {
		return nil, err
	}
	return []Change{c}, nil
}

func identityFor(scope Scope, cmd Command, old Identity, now time.Time) Identity {
	if cmd.Action == "create" {
		return Identity{ID: digest([]byte(scopeID(scope) + ":" + cmd.Resource + ":" + cmd.RequestID))[:32], HoldingCode: scope.Holding, BusinessCode: scope.Company, Version: 1, CreatedAt: now, CreatedBy: scope.Actor, UpdatedAt: now, UpdatedBy: scope.Actor}
	}
	old.Version++
	old.UpdatedAt = now
	old.UpdatedBy = scope.Actor
	return old
}
