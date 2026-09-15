import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import { describe, expect, it } from "vitest";

const splitterSource = readFileSync(
  fileURLToPath(new URL("./resizable-splitter.tsx", import.meta.url)),
  "utf8",
);

const settingSource = readFileSync(
  fileURLToPath(new URL("../../app/system-settings/system-settings-screen.tsx", import.meta.url)),
  "utf8",
);

const warehouseSource = readFileSync(
  fileURLToPath(new URL("../../app/system-settings/warehouse-tree-view.tsx", import.meta.url)),
  "utf8",
);

const companyBranchSource = readFileSync(
  fileURLToPath(new URL("../../app/system-settings/company-branch-tree-view.tsx", import.meta.url)),
  "utf8",
);

const productSource = readFileSync(
  fileURLToPath(new URL("../../app/menu/product-screen.tsx", import.meta.url)),
  "utf8",
);

const barcodeSource = readFileSync(
  fileURLToPath(new URL("../../app/menu/product-barcode-screen.tsx", import.meta.url)),
  "utf8",
);

const productSetSource = readFileSync(
  fileURLToPath(new URL("../../app/menu/product-set-screen.tsx", import.meta.url)),
  "utf8",
);

const barcodeShelfSource = readFileSync(
  fileURLToPath(new URL("../../app/menu/product-barcode-shelf-screen.tsx", import.meta.url)),
  "utf8",
);

const shortcutsSource = readFileSync(
  fileURLToPath(new URL("../../app/menu/manage-shortcuts-screen.tsx", import.meta.url)),
  "utf8",
);

const erpCrudSource = readFileSync(
  fileURLToPath(new URL("../../app/crud/erp-crud-workbench.tsx", import.meta.url)),
  "utf8",
);

describe("ResizableSplitter component & usage", () => {
  it("defines the standard floating pill with GripVertical icon and responsive breakpoints", () => {
    expect(splitterSource).toContain("GripVertical");
    expect(splitterSource).toContain('role="separator"');
    expect(splitterSource).toContain('aria-orientation="vertical"');
    expect(splitterSource).toContain("cursor-col-resize");
    expect(splitterSource).toContain('"hidden md:flex"');
    expect(splitterSource).toContain('"hidden lg:flex"');
    expect(splitterSource).toContain('"hidden xl:flex"');
  });

  it("handles keyboard interaction and double-click reset", () => {
    expect(splitterSource).toContain("onKeyDown");
    expect(splitterSource).toContain("onDoubleClick");
    expect(splitterSource).toContain("onPointerDown");
    expect(splitterSource).toContain("aria-valuenow");
    expect(splitterSource).toContain("aria-valuemin");
    expect(splitterSource).toContain("aria-valuemax");
  });

  it("is adopted by SettingMasterDetail in system-settings-screen", () => {
    expect(settingSource).toContain("<ResizableSplitter");
    expect(settingSource).toContain("isResizingPane");
    expect(settingSource).toContain("handlePaneResizeReset");
    expect(settingSource).toContain("handlePaneKeyDown");
    // Ensure old plain divider is gone
    expect(settingSource).not.toContain('className="hidden w-1.5 shrink-0 cursor-col-resize bg-border/70 transition hover:bg-primary/50 lg:block"');
  });

  it("is adopted by BOM, ProductGroup, ProductSubgroup, and ProductCategory in system-settings-screen", () => {
    expect(settingSource).toContain("BOM_SPLIT_DEFAULT_LEFT");
    expect(settingSource).toContain("startBomSplitResize");
    expect(settingSource).toContain("adjustBomSplitWithKeyboard");
    expect(settingSource).toContain("TREE_SPLIT_DEFAULT_LEFT");
    expect(settingSource).toContain("startTreeSplitResize");
    expect(settingSource).toContain("adjustTreeSplitWithKeyboard");
    expect(settingSource).toContain("CATEGORY_SPLIT_DEFAULT_LEFT");
    expect(settingSource).toContain("startCategorySplitResize");
    expect(settingSource).toContain("adjustCategorySplitWithKeyboard");
  });

  it("is adopted by warehouse-tree-view", () => {
    expect(warehouseSource).toContain("<ResizableSplitter");
    expect(warehouseSource).toContain("handleSidebarKeyDown");
    expect(warehouseSource).toContain("handleSidebarResizeReset");
  });

  it("is adopted by company-branch-tree-view", () => {
    expect(companyBranchSource).toContain("<ResizableSplitter");
    expect(companyBranchSource).toContain("handleSidebarKeyDown");
    expect(companyBranchSource).toContain("handleSidebarResizeReset");
  });

  it("is adopted by product-screen", () => {
    expect(productSource).toContain("<ResizableSplitter");
    expect(productSource).toContain("PRODUCT_SPLIT_DEFAULT_LEFT");
    expect(productSource).toContain("startSplitResize");
    expect(productSource).toContain("adjustSplitWithKeyboard");
  });

  it("is adopted by product-barcode-screen", () => {
    expect(barcodeSource).toContain("<ResizableSplitter");
    expect(barcodeSource).toContain("PRODUCT_SPLIT_DEFAULT_LEFT");
    expect(barcodeSource).toContain("startSplitResize");
    expect(barcodeSource).toContain("adjustSplitWithKeyboard");
  });

  it("is adopted by product-set-screen", () => {
    expect(productSetSource).toContain("<ResizableSplitter");
    expect(productSetSource).toContain("PRODUCT_SET_SIDEBAR_DEFAULT_WIDTH");
    expect(productSetSource).toContain("handleSidebarResizeStart");
    expect(productSetSource).toContain("handleSidebarKeyDown");
  });

  it("is adopted by product-barcode-shelf-screen", () => {
    expect(barcodeShelfSource).toContain("<ResizableSplitter");
    expect(barcodeShelfSource).toContain("SHELF_SPLIT_DEFAULT_LEFT");
    expect(barcodeShelfSource).toContain("startSplitResize");
    expect(barcodeShelfSource).toContain("adjustSplitWithKeyboard");
  });

  it("is adopted by manage-shortcuts-screen", () => {
    expect(shortcutsSource).toContain("<ResizableSplitter");
    expect(shortcutsSource).toContain("SHORTCUTS_SPLIT_DEFAULT_LEFT");
    expect(shortcutsSource).toContain("startSplitResize");
    expect(shortcutsSource).toContain("adjustSplitWithKeyboard");
  });

  it("is adopted by erp-crud-workbench", () => {
    expect(erpCrudSource).toContain("<ResizableSplitter");
    expect(erpCrudSource).toContain("bc_erp_crud_splitter_width");
  });

  it("exports useSplitPercent and is adopted across screens", () => {
    expect(splitterSource).toContain("useSplitPercent");
    expect(shortcutsSource).toContain("useSplitPercent");
    expect(barcodeShelfSource).toContain("useSplitPercent");
    expect(productSetSource).toContain("useSplitPercent");
    expect(companyBranchSource).toContain("useSplitPercent");
    expect(settingSource).toContain("useSplitPercent");
    expect(productSource).toContain("useSplitPercent");
    expect(erpCrudSource).toContain("useSplitPercent");
  });
});

