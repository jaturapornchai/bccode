import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const source = readFileSync(
  fileURLToPath(new URL("./company-branch-tree-view.tsx", import.meta.url)),
  "utf8",
);

describe("CompanyBranchTreeView organization-management contract", () => {
  it("loads the complete management tree and keeps inactive records manageable", () => {
    expect(source).toContain("/organization/company?management=true&_");
    expect(source).toContain("/organization/branch?management=true&_");
    expect(source).not.toContain("if (record.isactive === false) return false");
  });

  it("requires a nonblank primary-language name before saving", () => {
    expect(source).toContain("const hasRequiredName = formNames.some");
    expect(source).toContain("!hasRequiredName ||");
    expect(source).toContain("firstRequired");
  });

  it("updates the tree after editing and does not expose unsupported organization deletion", () => {
    expect(source).toContain("setCompanies((prev) => prev.map");
    expect(source).toContain("setBranches((prev) => prev.map");
    expect(source).not.toContain('method: "DELETE"');
    expect(source).not.toContain('showConfirmCodeDialog("delete"');
  });

  it("renders created records from the saved backend entity without synthesizing identities", () => {
    expect(source).toContain("entity?: CompanyRecord | BranchRecord");
    expect(source).toContain("const createdData = formType.startsWith(\"create\") ? json.data?.entity : undefined");
    expect(source).toContain("data: savedEntity");
    expect(source).not.toContain("guidfixed: json.id");
    expect(source).not.toContain("const createdData = {");
  });

  it("forces newly created organizations active and filters both deletion markers", () => {
    const createCompany = source.slice(
      source.indexOf('if (formType === "createcompany")'),
      source.indexOf('} else if (formType === "editcompany")'),
    );
    const createBranch = source.slice(
      source.indexOf('} else if (formType === "createbranch")'),
      source.indexOf('} else if (formType === "editbranch")'),
    );

    expect(createCompany).toContain("isactive: true");
    expect(createBranch).toContain("isactive: true");
    expect(source).toContain('if (record.isdeleted === true) return false');
    expect(source).toContain('disabled={isReadOnlyMode || formType.startsWith("create")}');
  });

  it("uses canonical companyuid lineage and blocks new branches under legacy companies", () => {
    const createBranch = source.slice(
      source.indexOf('} else if (formType === "createbranch")'),
      source.indexOf('} else if (formType === "editbranch")'),
    );

    expect(createBranch).toContain("companyuid: selectedNode.companyuid");
    expect(createBranch).not.toContain("companyguid:");
    expect(source).toContain('company.companyuid?.trim() || company.guidfixed?.trim() || ""');
    expect(source).toContain('branch.companyuid?.trim() || branch.companyguid?.trim() || ""');
    expect(source).toContain("const canAddBranch = canCreateOrganization && Boolean(companyUID)");
    expect(source).toContain("disabled={!canCreateOrganization || !selectedCompanyUID}");
  });
});
