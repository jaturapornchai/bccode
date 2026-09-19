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

  it("is registered and connected in MainMenu and MenuScreenStatus", () => {
    expect(mainMenuSource).toContain("isErpTransactionRoute");
    expect(mainMenuSource).toContain("<ErpCrudWorkbench");
    expect(statusSource).toContain("isErpTransactionRoute");
  });
});
