-- GL budgets (กำหนดงบประมาณ, Champ menu 5500 / table BCGLBudget): one header per budget +
-- one amount per account per fiscal period. period_no is the month index inside the fiscal
-- year (1 = month of fiscal_year start date). Dimension codes '' mean "all" (the budget
-- applies to every branch/department/project). status is Champ's open/closed flag.
-- ADR docs/kms/decisions/2026-09-25-gl-monthly-budget.md
CREATE TABLE IF NOT EXISTS gl_budgets (
  company text NOT NULL,
  code text NOT NULL,
  name text NOT NULL,
  fiscal_year text NOT NULL,
  branch_code text NOT NULL DEFAULT '',
  department_code text NOT NULL DEFAULT '',
  project_code text NOT NULL DEFAULT '',
  status text NOT NULL DEFAULT 'open' CHECK (status IN ('open','closed')),
  remark text NOT NULL DEFAULT '',
  version bigint NOT NULL DEFAULT 1 CHECK (version >= 1),
  created_by text NOT NULL DEFAULT '',
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_by text NOT NULL DEFAULT '',
  updated_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (company, code)
);
CREATE INDEX IF NOT EXISTS idx_gl_budgets_fiscal_year ON gl_budgets(company, fiscal_year);

CREATE TABLE IF NOT EXISTS gl_budget_lines (
  company text NOT NULL,
  budget_code text NOT NULL,
  account_code text NOT NULL,
  period_no smallint NOT NULL CHECK (period_no BETWEEN 1 AND 12),
  amount numeric(18,2) NOT NULL CHECK (amount >= 0),
  PRIMARY KEY (company, budget_code, account_code, period_no),
  CONSTRAINT fk_gl_budget_lines_budget FOREIGN KEY (company, budget_code) REFERENCES gl_budgets(company, code) ON DELETE CASCADE
);
