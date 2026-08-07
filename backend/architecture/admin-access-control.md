# Admin And Multi-Company Access Control

## Objective

Design access control for BC Ai Account so:

- Admin can have many people.
- The first admin can assign which email addresses are platform admins.
- Admin can assign which email addresses are owners for all companies in a company group.
- Admin/owner can assign which email addresses can access which company/business and which branches.
- Existing core `holdingcode` data remains unchanged. Logical `tenant_id` uses the same value as existing `holdingcode`. Some newer GoAPI modules use `holdingcode` as an API/DTO field and must map it explicitly.

## Scope Model

Use these scopes:

```text
platform          = whole BC Ai Account installation
company_group_id  = owner group that contains many companies/businesses
tenant_id         = one company/business, same value as core holdingcode for existing data
branch_id         = one branch under tenant_id
user_id/email     = person identity from login provider
```

## Role Model

Use role names that describe authority, not storage tables:

| Role | Scope | Can do |
| --- | --- | --- |
| `platform_owner` | platform | First admin, can manage platform admins and system-level policies |
| `platform_admin` | platform | Manage tenants/groups/support tasks, but cannot view tenant data unless granted |
| `group_owner` | company_group_id | See/manage all tenants and branches in the group |
| `group_admin` | company_group_id | Manage users/access inside the group, limited by granted permissions |
| `tenant_owner` | tenant_id | Full access to one company/business and all branches |
| `tenant_admin` | tenant_id | Manage users/settings inside one company/business |
| `branch_admin` | tenant_id + branch_id | Manage one or more branches |
| `branch_user` | tenant_id + branch_id | Work in one or more branches |
| `viewer` | any business scope | Read-only dashboard/report access |

Rules:

- `platform_admin` is not automatically data owner. To view customer data, the user must also have group/tenant/branch access or use an audited support mode.
- `group_owner` is the role for "email ไหนเป็นเจ้าของกิจการทั้งหมด" inside one company group.
- Branch-level access is always under a `tenant_id`; branch access without tenant is invalid.
- At least one active `platform_owner` must remain.

## Data Model

### Users

Keep user identity based on normalized email.

```text
users(
  id,
  email_normalized,
  display_name,
  provider,
  status,
  createdat,
  updatedat
)
```

Email normalization:

- Trim spaces.
- Lowercase.
- Store original display email separately if needed.

### Platform Admins

```text
platform_admins(
  email_normalized,
  role,              # platform_owner | platform_admin
  status,            # active | invited | disabled
  created_by_email,
  createdat,
  updatedat
)
```

Current platform policy:

1. The Global Admin email is `jaturapornchai@gmail.com`.
2. Do not auto-promote other verified-email accounts to a platform role.
3. Any active account with a valid linked email may create a Holding; the creator becomes that Holding's `OWNER`.
4. A Holding may have multiple `ADMIN` members. Holding roles remain separate from Global Admin.
5. The system must reject disabling/removing the last active platform owner after platform-role enforcement is implemented.

Runtime note: the current backend enforces the Holding rules above but does not yet resolve a platform Global Admin role. Do not claim cross-Holding Global Admin authority until its operations and enforcement are implemented.

Optional production hardening:

```text
BC_AI_ERP_BOOTSTRAP_ADMIN_EMAILS=jaturapornchai@gmail.com
```

Use this only during first deployment from a controlled secret source or environment config.

### Company Groups

```text
company_groups(
  company_group_id,
  name,
  status,
  created_by_email,
  createdat,
  updatedat
)
```

### Tenants

Use existing core `holdingcode` as tenant id.

```text
tenants(
  tenant_id,          # same value as core holdingcode
  company_group_id,
  name,
  status,
  createdat,
  updatedat
)
```

For existing records, do not add a new tenant id only for naming consistency.

### Access Grants

One collection/table should represent all company/tenant/branch grants.

```text
access_grants(
  id,
  email_normalized,
  company_group_id,
  tenant_id,          # same value as core holdingcode, nullable only for group-level grants
  branch_ids,         # empty means all branches for that tenant when branch_scope = all
  branch_scope,       # all | selected
  role,
  permissions,
  status,             # active | invited | disabled
  created_by_email,
  updated_by_email,
  createdat,
  updatedat
)
```

Recommended indexes:

```text
(email_normalized, status)
(company_group_id, status)
(tenant_id, email_normalized, status)
(tenant_id, branch_scope, status)
```

## Permission Rules

### Platform Admin

Can:

- Add/remove platform admins.
- Create company groups.
- Assign group owners/admins.
- View system health and audit logs.

Cannot by default:

- Read tenant business data.
- Export customer data.
- Edit tenant documents.

Those require explicit tenant/group grant or audited support mode.

### Group Owner

Can:

- See all tenants in the `company_group_id`.
- See all branches under those tenants.
- Assign group admins, tenant owners, tenant admins, branch admins, branch users, and viewers.
- Open owner-level BI overview across the group.

### Tenant Owner/Admin

Can:

- Access the selected `tenant_id`.
- Manage branch grants under that tenant.
- Use existing `shopUsers` compatibility where needed.

### Branch User

Can:

- Access only granted branches under the selected `tenant_id`.
- See only reports/transactions filtered by `tenant_id + branch_id`.

## Runtime Resolution

Every request must resolve access like this:

```text
JWT/email
  -> normalize email
  -> load platform_admins
  -> load access_grants by email
  -> derive allowed company_group_id list
  -> derive allowed tenant_id list
  -> derive allowed branch_id list per tenant
  -> inject into request context
```

Context shape:

```json
{
  "email": "owner@example.com",
  "platform_roles": ["platform_admin"],
  "groups": ["group_01"],
  "tenants": [
    {
      "tenant_id": "existing_holdingcode",
      "branch_scope": "selected",
      "branch_ids": ["B001", "B002"],
      "roles": ["branch_admin"]
    }
  ]
}
```

Backend rules:

- Frontend may send selected `tenant_id` / `branch_id`, but backend must validate them against resolved access.
- Query builders must reject requests where selected tenant/branch is not allowed.
- Owner overview must use only the authorized `tenant_id` list.
- ClickHouse must be queried only through backend APIs after access resolution.

## API Design

### Bootstrap / Platform Admin

```text
GET  /api/v1/admin/bootstrap/status
POST /api/v1/admin/bootstrap/claim
GET  /api/v1/admin/platform-admins
PUT  /api/v1/admin/platform-admins/:email
DELETE /api/v1/admin/platform-admins/:email
```

### Company Group Access

```text
GET /api/v1/company-groups/:companyGroupId/access
PUT /api/v1/company-groups/:companyGroupId/access/:email
DELETE /api/v1/company-groups/:companyGroupId/access/:email
```

### Tenant And Branch Access

```text
GET /api/v1/tenants/:tenantId/access
PUT /api/v1/tenants/:tenantId/access/:email
DELETE /api/v1/tenants/:tenantId/access/:email
```

Request example:

```json
{
  "email": "staff@example.com",
  "role": "branch_user",
  "branch_scope": "selected",
  "branch_ids": ["B001", "B002"],
  "permissions": ["sales.read", "sales.write", "stock.read"]
}
```

### Current User Access

```text
GET /api/v1/me/access
```

Returns all groups, tenants, branches, and roles the current user may select.

## Workflow

### First Setup

1. First verified admin logs in.
2. If no active `platform_owner` exists, backend creates `platform_owner`.
3. Platform owner enters admin emails.
4. Backend stores them in `platform_admins` as active or invited.
5. Every action is audit logged.

### Create Company Group

1. Platform owner/admin creates a company group.
2. Assign one or more `group_owner` emails.
3. Link existing `holdingcode` values into the group as `tenant_id`.
4. Group owner can open overview for all linked tenants.

### Grant Tenant/Branch Access

1. Owner/admin selects a company/business.
2. Enters an email.
3. Selects role.
4. Selects all branches or specific branches.
5. Backend validates actor permission.
6. Backend writes `access_grants`.
7. On next request, the target email can see only allowed companies/branches.

## ClickHouse And BI

Owner overview:

```text
allowed_tenants = resolveAccess(email).tenants
query ClickHouse WHERE company_group_id = ? AND tenant_id IN allowed_tenants
```

Branch report:

```text
WHERE tenant_id = ?
  AND branch_id IN allowed_branches_for_tenant
```

Never trust tenant/branch lists from the browser.

## Audit

Every admin/access change must write audit data:

```text
actor_email
target_email
action
company_group_id
tenant_id
branch_ids
old_value
new_value
ip
user_agent
createdat
```

Sensitive actions:

- Add/remove platform admin.
- Add/remove group owner.
- Grant all-branch access.
- Grant owner/admin roles.
- Support cross-tenant access.

## Config

```text
BC_AI_ERP_BOOTSTRAP_ADMIN_EMAILS
BC_AI_ERP_REQUIRE_ADMIN_AUDIT=true
BC_AI_ERP_SUPPORT_ACCESS_DEFAULT=deny
```

Production must load these from a controlled secret source or environment config.

## Dependencies

- Auth provider must provide verified email.
- Existing `shopUsers` can remain for tenant-level compatibility.
- Branch data comes from existing branch collections using `holdingcode` + branch code/id.
- Backend request context must support resolved access.
- BI/report APIs must accept authorized tenant/branch filters from backend context.

## Migration Strategy

Phase 1:

- Keep `shopUsers` and existing role behavior.
- Add `platform_admins`, `company_groups`, `tenants`, and `access_grants`.
- Set `tenant_id = holdingcode`.
- Resolve access from new grants first, fallback to `shopUsers` for legacy screens.

Phase 2:

- Update menu/workspace/report APIs to use resolved access context.
- Add branch-level enforcement for screens that need branch isolation.
- Add audit logs.

Phase 3:

- Gradually replace direct role checks such as `ROLE_OWNER` with policy checks.
- Keep `holdingcode` physical storage unless a functional migration is required.

## Limitations

- This design does not implement row-level security by itself; every API/repository still needs enforced filters.
- Branch-level restrictions only work correctly after all branch-sensitive APIs use resolved access.
- Old code that checks only `ROLE_OWNER` must be migrated carefully to avoid privilege regressions.
- Email-based invites depend on verified login email. Unverified email must not receive access.
