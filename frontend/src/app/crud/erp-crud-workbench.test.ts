import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const workbenchSource = readFileSync(
  resolve(process.cwd(), "src/app/crud/erp-crud-workbench.tsx"),
  "utf8",
);

const mainMenuSource = readFileSync(
  resolve(process.cwd(), "src/app/menu/main-menu-screen.tsx"),
  "utf8",
);

const statusSource = readFileSync(
  resolve(process.cwd(), "src/lib/menu-screen-status.ts"),
  "utf8",
);

describe("ErpCrudWorkbench Master-Detail Contract (.agents/skills/datacrud/SKILL.md)", () => {
  it("enforces ResizableSplitter and persistent localStorage width", () => {
    expect(workbenchSource).toContain("<ResizableSplitter");
    expect(workbenchSource).toContain("useSplitPercent");
    expect(workbenchSource).toContain("bc_erp_crud_splitter_width");
    expect(workbenchSource).toContain("breakpoint=");
    expect(workbenchSource).toContain("onDoubleClick=");
  });

  it("follows exact CSS class standards for high-density Master-Detail list", () => {
    expect(workbenchSource).toContain("bc-list-toolbar");
    expect(workbenchSource).toContain("bc-list-header");
    expect(workbenchSource).toContain("bc-list-row");
    expect(workbenchSource).toContain("3px 8px"); // compact padding
  });

  it("strictly enforces that row click views read-only and edit mode requires pencil click", () => {
    // Row click only selects row and does not enter edit mode
    expect(workbenchSource).toContain("handleSelectRow");
    expect(workbenchSource).toContain("setIsEditing(false)");

    // Edit mode explicitly triggered by pencil icon
    expect(workbenchSource).toContain("handleStartEdit");
    expect(workbenchSource).toContain("Pencil");
    expect(workbenchSource).toContain("setIsEditing(true)");

    // Stop propagation on edit and delete buttons to prevent row trigger
    expect(workbenchSource).toContain("e.stopPropagation()");
  });

  it("implements Dirty Form Guard with Thai confirmation dialog", () => {
    expect(workbenchSource).toContain("isDirty");
    expect(workbenchSource).toContain("useConfirmDialog");
    expect(workbenchSource).toContain("มีการเปลี่ยนแปลงที่ยังไม่ได้บันทึก");
  });

  it("pins action buttons at header and footer for accessible saving", () => {
    expect(workbenchSource).toContain("sticky top-0");
    expect(workbenchSource).toContain("sticky bottom-0");
    expect(workbenchSource).toContain("handleSaveDoc");
    expect(workbenchSource).toContain("handleCancelEdit");
  });

  it("lets the backend issue the document number and warns when a type is store-only", () => {
    expect(workbenchSource).not.toContain("defaultDocPrefix");
    expect(workbenchSource).toContain("docno_auto_placeholder");
    expect(workbenchSource).toContain("formDoc.id && !formDoc.docno?.trim()");
    expect(workbenchSource).toContain("ERP_AUTO_POSTING_MODULES.includes(config.apiPath)");
    expect(workbenchSource).toContain("erp_store_only_notice");
  });

  it("offers create, edit and delete only when the menu permission allows them", () => {
    expect(workbenchSource).toContain("useScreenActions(auth, workspace, route)");
    expect(workbenchSource).toContain("onNew: screenActions.create ? handleStartCreate : undefined");
    expect(workbenchSource).toContain("!screenActions.create ? null");
    expect(workbenchSource).toContain("{screenActions.update ? (");
    expect(workbenchSource).toContain("{screenActions.delete ? (");
    expect(workbenchSource).toContain("if (!screenActions.update || erpDocLifecycle(doc)");
    expect(workbenchSource).toContain("!screenActions.delete || erpDocLifecycle(doc)");
    expect(workbenchSource).toContain("erp_err_action_forbidden");
    expect(mainMenuSource).toContain("route={activeTab.route} auth={auth} workspace={workspace}");
  });

  it("guards the screens that have no backend and no longer lists live routes by hand", () => {
    expect(mainMenuSource).toContain("isRetiredBackendScreen(activeTab.route)) return plannedCard");
    expect(statusSource).not.toContain("LIVE_ERP_TRANSACTION_ROUTES");
  });

  it("is registered and connected in MainMenu and MenuScreenStatus", () => {
    expect(mainMenuSource).toContain("isErpTransactionRoute");
    expect(mainMenuSource).toContain("<ErpCrudWorkbench");
    expect(statusSource).toContain("isErpTransactionRoute");
  });
});

describe("ErpCrudWorkbench approved / cancelled document lock", () => {
  it("derives the lifecycle from the shared helper instead of status === 1 only", () => {
    expect(workbenchSource).toContain("erpDocLifecycle");
    expect(workbenchSource).not.toContain("item.status === 1");
    expect(workbenchSource).not.toContain("selectedDoc.status === 1");
  });

  it("filters the draft/approved tabs by the lifecycle helper because the backend omits status 0", () => {
    expect(workbenchSource).toContain('statusFilter === "draft" && erpDocLifecycle(item) !== "draft"');
    expect(workbenchSource).toContain('statusFilter === "approved" && erpDocLifecycle(item) !== "approved"');
    expect(workbenchSource).not.toContain("item.status !== 0");
    expect(workbenchSource).not.toContain("item.status !== 1");
  });

  it("shows a cancelled badge and a lock hint with language keys", () => {
    expect(workbenchSource).toContain('backendText(dictionary, "cancelled"');
    expect(workbenchSource).toContain("erp_doc_locked_cancelled_hint");
    expect(workbenchSource).toContain("erp_doc_locked_approved_hint");
  });

  it("disables edit and delete for locked rows with title and aria-label, and guards the handlers", () => {
    expect(workbenchSource).toContain("disabled={!!lockedHint}");
    expect(workbenchSource).toContain("aria-label={lockedHint ||");
    expect(workbenchSource).toContain("disabled={!!selectedLockedHint}");
    expect(workbenchSource).toContain('erpDocLifecycle(doc) !== "draft"');
  });
});

// Rule 20: ยอดเงินบนจอเป็นข้อความทศนิยม — แสดงด้วย formatAmount คำนวณตัวอย่างด้วย erpLineAmount/erpDocTotals ไม่ใช้ Number()/Math
describe("ErpCrudWorkbench keeps amounts as decimal strings (Rule 20)", () => {
  it("never does money arithmetic with Number(), parseFloat or Math.round", () => {
    expect(workbenchSource).not.toMatch(/Number\(\s*(item|formDoc|selectedDoc|d|e\.target)\b/);
    expect(workbenchSource).not.toContain("parseFloat");
    expect(workbenchSource).not.toMatch(/Math\.round\([^)]*(vat|total|Vat|Total)/);
    expect(workbenchSource).not.toContain("toLocaleString(\"th-TH\"");
  });

  it("previews line and document totals with the exact helpers and shows them via formatAmount", () => {
    expect(workbenchSource).toContain("item.sumamount = erpLineAmount(item)");
    expect(workbenchSource).toContain("...erpDocTotals(nextDetails, formDoc.vattype, formDoc.vatrate)");
    expect(workbenchSource).toContain('formatAmount(selectedDoc.totalamount || "0")');
    expect(workbenchSource).toContain("updateLineItem(idx, { qty: e.target.value })");
  });
});
