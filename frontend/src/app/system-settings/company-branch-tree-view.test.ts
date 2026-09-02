import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const source = readFileSync(
  fileURLToPath(new URL("./company-branch-tree-view.tsx", import.meta.url)),
  "utf8",
);
const namesEditorSource = readFileSync(
  fileURLToPath(new URL("../../components/product-barcode/names-editor.tsx", import.meta.url)),
  "utf8",
);

describe("CompanyBranchTreeView organization-management contract", () => {
  it("loads the complete management tree and keeps inactive records manageable", () => {
    expect(source).toContain("/organization/company?management=true&_");
    expect(source).toContain("/organization/branch?management=true&_");
    expect(source).not.toContain("if (record.isactive === false) return false");
  });

  it("explains missing required fields before confirmation instead of silently disabling save", () => {
    expect(source).toContain("const hasRequiredName = formNames.some");
    expect(source).toContain("const requiredFormError =");
    expect(source).toContain("if (requiredFormError)");
    expect(source).toContain('disabled={saving || saveSuccess || logoUploading}');
    expect(source).toContain('id="company-branch-required-fields"');
    expect(source).toContain("รหัสภาษาไม่ใช่ชื่อ");
    expect(source).toContain("firstRequired");
  });

  it("keeps company and branch logos independent in both payloads and UI guidance", () => {
    const createCompany = source.slice(
      source.indexOf('if (formType === "createcompany")'),
      source.indexOf('} else if (formType === "editcompany")'),
    );
    const createBranch = source.slice(
      source.indexOf('} else if (formType === "createbranch")'),
      source.indexOf('} else if (formType === "editbranch")'),
    );

    expect(createCompany).toContain("logouri: formLogoUri");
    expect(createBranch).toContain("logouri: formLogoUri");
    expect(source).toContain("โลโก้นี้บันทึกเฉพาะสาขานี้ และแตกต่างจากสาขาอื่นได้");
  });

  it("selects name language from the active languages configured in step one", () => {
    expect(source).toContain("normalizeLanguageConfigs(configs, defaultCode, { forcePrimaryFirst: true })");
    expect(source).toContain("languageSelect");
    expect(namesEditorSource).toContain("if (languageSelect)");
    expect(namesEditorSource).toContain('aria-label={language === "th" ? "เลือกภาษาของชื่อ"');
    expect(namesEditorSource).toContain("selectableLanguages.map((code) =>");
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
