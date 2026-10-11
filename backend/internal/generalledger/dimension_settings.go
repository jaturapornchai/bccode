package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"
)

// DimensionSettingsKind is the company switches "ใช้ผังบัญชีแบบแยก…" of Champ's company-defaults screen (Option 2, items 8-14;
// myfrmsysconfig.cpp:927-946, stored as XSeparate in BCConfigurations SQLSERVER_Script.sql:743-750 and BCConfigurations2 :947-954).
// One record per company in gl_records (id and code "default"), every switch OFF until the owner turns it on (Champ DDL DEFAULT 0).
// Design: docs/kms/champ/design-2026-10-11/gl-dimensions.md §3.4.
const (
	DimensionSettingsKind = "dimension-settings"
	dimensionSettingsID   = "default"
)

// DimensionSettingsInput is what a client sends and what the screens read back. Order and meaning follow Champ's seven check boxes;
// BranchSeparate is stored to keep all seven but has no effect until the owner decides whether branch becomes a line-level
// dimension (design OD-2): BC keeps the branch on the journal header.
type DimensionSettingsInput struct {
	DepartSeparate   bool `json:"departseparate" bson:"departseparate"`
	ProjectSeparate  bool `json:"projectseparate" bson:"projectseparate"`
	AllocateSeparate bool `json:"allocateseparate" bson:"allocateseparate"`
	JobSeparate      bool `json:"jobseparate" bson:"jobseparate"`
	PartSeparate     bool `json:"partseparate" bson:"partseparate"`
	SideSeparate     bool `json:"sideseparate" bson:"sideseparate"`
	BranchSeparate   bool `json:"branchseparate" bson:"branchseparate"`
}

// DimensionSettings is the stored record.
type DimensionSettings struct {
	Identity               `bson:",inline"`
	DimensionSettingsInput `bson:",inline"`
}

// DimensionSettingsView is the answer of GET /gl/v2/dimension-settings: the seven switches and the record version
// (0 = nothing saved yet, every switch off) the next update must send back.
type DimensionSettingsView struct {
	DimensionSettingsInput
	Version int64 `json:"version"`
}

// View is the screen's copy of the settings.
func (s DimensionSettings) View() DimensionSettingsView {
	return DimensionSettingsView{DimensionSettingsInput: s.DimensionSettingsInput, Version: s.Version}
}

// enabled reports whether the company switch of the register is on.
func (s DimensionSettings) enabled(kind string) bool {
	switch kind {
	case DepartmentKind:
		return s.DepartSeparate
	case ProjectKind:
		return s.ProjectSeparate
	case AllocateKind:
		return s.AllocateSeparate
	case JobKind:
		return s.JobSeparate
	case PartKind:
		return s.PartSeparate
	case SideKind:
		return s.SideSeparate
	}
	return false
}

type rowQuerier interface {
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

// loadDimensionSettings reads the company's switches; a company that never saved them has every switch off and version 0.
func loadDimensionSettings(ctx context.Context, q rowQuerier, company string) (DimensionSettings, error) {
	var payload []byte
	err := q.QueryRowContext(ctx, `SELECT payload || jsonb_build_object('id', id, 'version', version) FROM gl_records WHERE company=$1 AND kind=$2 AND id=$3 AND NOT COALESCE((payload->>'isdeleted')::boolean,false)`,
		company, DimensionSettingsKind, dimensionSettingsID).Scan(&payload)
	if errors.Is(err, sql.ErrNoRows) {
		return DimensionSettings{}, nil
	}
	if err != nil {
		return DimensionSettings{}, err
	}
	var settings DimensionSettings
	if err = json.Unmarshal(payload, &settings); err != nil {
		return DimensionSettings{}, err
	}
	return settings, nil
}

// DimensionSettings answers GET /gl/v2/dimension-settings for the screens that hide, lock or disable dimension controls by the switches.
func (p *Postgres) DimensionSettings(ctx context.Context, scope Scope) (DimensionSettingsView, error) {
	db, err := p.database(ctx, scope.Holding)
	if err != nil {
		return DimensionSettingsView{}, err
	}
	settings, err := loadDimensionSettings(ctx, db, scope.Company)
	if err != nil {
		return DimensionSettingsView{}, err
	}
	return settings.View(), nil
}

// mutateDimensionSettings saves the seven switches (resource dimension-settings, action update). The first save sends version 0;
// later saves must send the stored version (stale_version otherwise). Switching a dimension off is always allowed even when
// data uses it — the data stays and reports still read it (Champ myfrmsysconfig.cpp:1176-1183 does not check). The command is
// stored as a gl_events event, which is the audit trail of who changed which switch and when.
func (s *PostgresStore) mutateDimensionSettings(ctx context.Context, tx *sql.Tx, scope Scope, cmd Command, now time.Time) ([]Change, error) {
	if cmd.Action != "update" || cmd.ID != dimensionSettingsID {
		return nil, userError(CodeUnsupported, "ไม่รองรับคำสั่งนี้สำหรับการตั้งค่ามิติ")
	}
	if cmd.DimensionSettings == nil {
		return nil, userError(CodePayloadMissing, "ไม่พบข้อมูลการตั้งค่ามิติ")
	}
	old, err := loadDimensionSettings(ctx, tx, scope.Company)
	if err != nil {
		return nil, err
	}
	var next DimensionSettings
	if old.Version == 0 {
		if cmd.Version != 0 {
			return nil, userError(CodeStaleVersion, "รายการนี้เปลี่ยนไปแล้ว กรุณาโหลดข้อมูลล่าสุดก่อนบันทึก")
		}
		next.Identity = newIdentity(scope, now)
		next.ID = dimensionSettingsID
	} else {
		if err = checkVersion(cmd, old.Identity); err != nil {
			return nil, err
		}
		next.Identity = updateIdentity(scope, old.Identity, now)
	}
	next.DimensionSettingsInput = *cmd.DimensionSettings
	data, err := json.Marshal(next)
	if err != nil {
		return nil, err
	}
	return []Change{{Kind: DimensionSettingsKind, ID: dimensionSettingsID, Code: dimensionSettingsID, Payload: string(data)}}, nil
}
