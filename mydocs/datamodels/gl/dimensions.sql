-- ============================================================================
-- ทะเบียนมิติวิเคราะห์ GL 6 มิติ (GL Dimension Registers)
-- ตามต้นแบบ Champ 5.5 Pro (BCDepartment, BCProject, BCAllocate, BCJob, BCPart, BCSide)
-- ============================================================================

-- 1. แผนก (Department)
CREATE TABLE IF NOT EXISTS gl_departments (
    company_code VARCHAR(20) NOT NULL,
    code         VARCHAR(10) NOT NULL,
    name         VARCHAR(50) NOT NULL,
    is_active    BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    PRIMARY KEY (company_code, code)
);

-- 2. โครงการ (Project)
CREATE TABLE IF NOT EXISTS gl_projects (
    company_code VARCHAR(20) NOT NULL,
    code         VARCHAR(10) NOT NULL,
    name         VARCHAR(50) NOT NULL,
    is_active    BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    PRIMARY KEY (company_code, code)
);

-- 3. การจัดสรร (Allocation)
CREATE TABLE IF NOT EXISTS gl_allocates (
    company_code VARCHAR(20) NOT NULL,
    code         VARCHAR(10) NOT NULL,
    name         VARCHAR(50) NOT NULL,
    is_active    BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    PRIMARY KEY (company_code, code)
);

-- 4. งาน (Job)
CREATE TABLE IF NOT EXISTS gl_jobs (
    company_code VARCHAR(20) NOT NULL,
    code         VARCHAR(15) NOT NULL,
    name         VARCHAR(255) NOT NULL,
    is_active    BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    PRIMARY KEY (company_code, code)
);

-- 5. หน่วยงาน (Part)
CREATE TABLE IF NOT EXISTS gl_parts (
    company_code VARCHAR(20) NOT NULL,
    code         VARCHAR(10) NOT NULL,
    name         VARCHAR(255) NOT NULL,
    is_active    BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    PRIMARY KEY (company_code, code)
);

-- 6. ฝ่าย (Side)
CREATE TABLE IF NOT EXISTS gl_sides (
    company_code VARCHAR(20) NOT NULL,
    code         VARCHAR(10) NOT NULL,
    name         VARCHAR(255) NOT NULL,
    is_active    BOOLEAN NOT NULL DEFAULT true,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC'),
    PRIMARY KEY (company_code, code)
);

-- สวิตช์เปิด/ปิดมิติวิเคราะห์ระดับบริษัท (Company Dimension Switches)
-- ค่าเริ่มต้น 0=ปิด, 1=เปิด
CREATE TABLE IF NOT EXISTS gl_dimension_settings (
    company_code     VARCHAR(20) NOT NULL PRIMARY KEY,
    depart_separate  BOOLEAN NOT NULL DEFAULT false,
    project_separate BOOLEAN NOT NULL DEFAULT false,
    allocate_separate BOOLEAN NOT NULL DEFAULT false,
    job_separate     BOOLEAN NOT NULL DEFAULT false,
    part_separate    BOOLEAN NOT NULL DEFAULT false,
    side_separate    BOOLEAN NOT NULL DEFAULT false,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT (CURRENT_TIMESTAMP AT TIME ZONE 'UTC')
);
