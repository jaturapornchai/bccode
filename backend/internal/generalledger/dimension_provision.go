package generalledger

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

// Moving a company that already keeps department / project on its journal lines onto the dimension switches (design R9).
//
// Before the switches existed any company could put a department or project code on a line, and an account whose
// "แยกแผนก" flag was on with an empty allowed list accepted every code. Champ reads an empty list as "nothing allowed" and
// has the "เลือกทั้งหมด" box for "every code", and a new code on a line is refused while the company switch is off. So an
// existing company needs, once: the switches of the registers it already uses turned on, and "เลือกทั้งหมด" ticked on every
// flagged account that had no list. Nothing else changes: no line, balance or code is touched.

// DimensionProvisionPlan lists what ApplyDimensionProvision would write for one company; empty = nothing to do.
type DimensionProvisionPlan struct {
	Switches []string // register kinds whose company switch will be turned on
	Accounts []string // account codes that get "เลือกทั้งหมด" for at least one register

	settings        DimensionSettingsInput
	settingsVersion int64
	updates         []Command
}

// Empty reports whether the company already matches the Champ rules.
func (p DimensionProvisionPlan) Empty() bool { return len(p.Switches) == 0 && len(p.Accounts) == 0 }

// tickAllCodes sets "เลือกทั้งหมด" on every register whose flag is on while no code is listed and returns the registers it changed.
func tickAllCodes(a *Account) []string {
	var changed []string
	for _, d := range dimensionRegistry {
		dim := a.accountDimension(d.kind)
		if *dim.separate && !*dim.all && len(*dim.codes) == 0 {
			*dim.all = true
			changed = append(changed, d.kind)
		}
	}
	return changed
}

// PlanDimensionProvision reads the company and works out the plan; it writes nothing.
func (s *PostgresStore) PlanDimensionProvision(ctx context.Context, scope Scope) (DimensionProvisionPlan, error) {
	var plan DimensionProvisionPlan
	db, err := s.pg.database(ctx, scope.Holding)
	if err != nil {
		return plan, err
	}
	settings, err := loadDimensionSettings(ctx, db, scope.Company)
	if err != nil {
		return plan, err
	}
	plan.settings, plan.settingsVersion = settings.DimensionSettingsInput, settings.Version
	flagged := map[string]bool{}
	rows, err := db.QueryContext(ctx, `SELECT payload || jsonb_build_object('id', id, 'version', version) FROM gl_records WHERE company=$1 AND kind='accounts' AND NOT COALESCE((payload->>'isdeleted')::boolean,false) ORDER BY code`, scope.Company)
	if err != nil {
		return plan, err
	}
	defer rows.Close()
	for rows.Next() {
		var payload []byte
		var account Account
		if err = rows.Scan(&payload); err != nil {
			return plan, err
		}
		if err = json.Unmarshal(payload, &account); err != nil {
			return plan, err
		}
		for _, d := range dimensionRegistry {
			flagged[d.kind] = flagged[d.kind] || *account.accountDimension(d.kind).separate
		}
		if len(tickAllCodes(&account)) > 0 {
			plan.Accounts = append(plan.Accounts, account.AccountCode)
			plan.updates = append(plan.updates, Command{Resource: "accounts", Action: "update", RequestID: fmt.Sprintf("dimension-provision-account-%s-v%d", account.AccountCode, account.Version), ID: account.ID, Version: account.Version, Account: &account})
		}
	}
	if err = rows.Err(); err != nil {
		return plan, err
	}
	for _, d := range dimensionRegistry {
		if settings.enabled(d.kind) {
			continue
		}
		used := flagged[d.kind]
		if !used {
			if used, err = dimensionColumnUsed(ctx, db, scope.Company, d); err != nil {
				return plan, err
			}
		}
		if used {
			plan.Switches = append(plan.Switches, d.kind)
			setSwitch(&plan.settings, d.kind)
		}
	}
	return plan, nil
}

// dimensionColumnUsed: the company already has a ledger line, a draft journal or a budget naming a code of the register.
func dimensionColumnUsed(ctx context.Context, db *sql.DB, company string, d dimensionKind) (bool, error) {
	var used bool
	// d.column / d.jsonKey come from the dimensionRegistry table, never from the request.
	err := db.QueryRowContext(ctx, fmt.Sprintf(`SELECT EXISTS(SELECT 1 FROM gl_lines WHERE company=$1 AND %[1]s<>'')
		OR EXISTS(SELECT 1 FROM gl_budgets WHERE company=$1 AND %[1]s<>'')
		OR EXISTS(SELECT 1 FROM gl_records WHERE company=$1 AND kind='journals' AND NOT COALESCE((payload->>'isdeleted')::boolean,false)
			AND EXISTS(SELECT 1 FROM jsonb_array_elements(COALESCE(payload->'lines','[]'::jsonb)) l WHERE COALESCE(l->>'%[2]s','')<>''))`, d.column, d.jsonKey), company).Scan(&used)
	return used, err
}

func setSwitch(s *DimensionSettingsInput, kind string) {
	switch kind {
	case DepartmentKind:
		s.DepartSeparate = true
	case ProjectKind:
		s.ProjectSeparate = true
	case AllocateKind:
		s.AllocateSeparate = true
	case JobKind:
		s.JobSeparate = true
	case PartKind:
		s.PartSeparate = true
	case SideKind:
		s.SideSeparate = true
	}
}

// ApplyDimensionProvision writes the plan through the same store.Execute every screen uses (events, versions, audit trail).
// Accounts first, then the switches; it stops at the first refusal and says which command failed.
func (s *PostgresStore) ApplyDimensionProvision(ctx context.Context, scope Scope, plan DimensionProvisionPlan) error {
	for i, cmd := range plan.updates {
		if _, err := s.Execute(ctx, scope, cmd); err != nil {
			return fmt.Errorf("บัญชี %s: %w", plan.Accounts[i], err)
		}
	}
	if len(plan.Switches) == 0 {
		return nil
	}
	settings := plan.settings
	cmd := Command{Resource: DimensionSettingsKind, Action: "update", RequestID: fmt.Sprintf("dimension-provision-settings-v%d", plan.settingsVersion), ID: dimensionSettingsID, Version: plan.settingsVersion, DimensionSettings: &settings}
	if _, err := s.Execute(ctx, scope, cmd); err != nil {
		return fmt.Errorf("สวิตช์มิติ: %w", err)
	}
	return nil
}
