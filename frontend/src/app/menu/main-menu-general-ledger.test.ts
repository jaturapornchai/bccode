import { readFileSync } from "node:fs";
import { fileURLToPath } from "node:url";
import ts from "typescript";
import { describe, expect, it, vi } from "vitest";
import { isGeneralLedgerRoute } from "@/lib/general-ledger";

const source = readFileSync(fileURLToPath(new URL("./main-menu-screen.tsx", import.meta.url)), "utf8");
const file = ts.createSourceFile("main-menu-screen.tsx", source, ts.ScriptTarget.Latest, true, ts.ScriptKind.TSX);
function functionSource(name: string) {
  let text = "";
  function visit(node: ts.Node) { if (ts.isFunctionDeclaration(node) && node.name?.text === name) text = node.getText(file); ts.forEachChild(node, visit); }
  visit(file); if (!text) throw new Error(`Missing menu action ${name}`); return text;
}
type Tab = { id: string; route: string; title: string; closable: boolean };
function actions(initial: Tab[], dirty: string[], accepted: boolean) {
  let current = [...initial];
  const setTabs = vi.fn((update: (tabs: Tab[]) => Tab[]) => { current = update(current); });
  const setActiveTabId = vi.fn(), confirmDiscard = vi.fn(async () => accepted);
  const code = ts.transpileModule(["confirmLeaveLedger", "closeTab", "selectWorkTab", "openMenuItem"].map(functionSource).join("\n"), { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.None } }).outputText;
  const create = new Function("tabs", "dirtyLedgerRoutes", "confirmDiscard", "setTabs", "setActiveTabId", "isGeneralLedgerRoute", `const activeTabId = tabs[0]?.id, firstTab = {id:'home',title:'หน้าหลัก',route:'/',closable:false}; const canAccessMenuItem=()=>true, menuText=(label)=>label.th, language='th', backendLanguage={}, menuUsageKey='', randomId=()=> 'new-id'; ${code}; return {closeTab,selectWorkTab,openMenuItem};`);
  return { ...create(initial, { current: new Set(dirty) }, confirmDiscard, setTabs, setActiveTabId, isGeneralLedgerRoute), setTabs, setActiveTabId, confirmDiscard, current: () => current } as { closeTab: (id: string) => Promise<void>; selectWorkTab: (id: string) => void; openMenuItem: (item: { id: string; route: string; label: { th: string } }, options: { forceNew: boolean }) => void; setTabs: typeof setTabs; setActiveTabId: typeof setActiveTabId; confirmDiscard: typeof confirmDiscard; current: () => Tab[] };
}
const ledger: Tab = { id: "ledger-one", route: "/gl/journal/jv", title: "รายวันทั่วไป", closable: true };
const home: Tab = { id: "home", route: "/", title: "หน้าหลัก", closable: false };
describe("ledger WorkTab data preservation", () => {
  it("cancelling close retains the dirty editor and active tab", async () => {
    const menu = actions([ledger, home], [ledger.route], false);
    await menu.closeTab(ledger.id);
    expect(menu.confirmDiscard).toHaveBeenCalledOnce();
    expect(menu.setTabs).not.toHaveBeenCalled();
    expect(menu.setActiveTabId).not.toHaveBeenCalled();
  });
  it("confirmed close removes only its target and clean close needs no confirmation", async () => {
    const menu = actions([ledger, home], [ledger.route], true);
    await menu.closeTab(ledger.id);
    expect(menu.current()).toEqual([home]);
    const clean = actions([ledger, home], [], false);
    await clean.closeTab(ledger.id);
    expect(clean.confirmDiscard).not.toHaveBeenCalled();
    expect(clean.current()).toEqual([home]);
  });
  it("switching a tab preserves its editor and selecting New reuses the GL route", () => {
    const menu = actions([ledger, home], [ledger.route], false);
    menu.selectWorkTab(home.id);
    expect(menu.current()).toEqual([ledger, home]);
    expect(menu.setTabs).not.toHaveBeenCalled();
    menu.openMenuItem({ id: "gl-journals", route: ledger.route, label: { th: ledger.title } }, { forceNew: true });
    expect(menu.current()).toEqual([ledger, home]);
    expect(menu.setActiveTabId).toHaveBeenLastCalledWith(ledger.id);
    expect(menu.confirmDiscard).not.toHaveBeenCalled();
  });
  it("keeps the WorkTab container mounted outside the search conditional", () => {
    let container: ts.JsxElement | undefined;
    function visit(node: ts.Node) { if (ts.isJsxElement(node) && node.openingElement.attributes.properties.some((attribute) => ts.isJsxAttribute(attribute) && attribute.name.getText(file) === "data-worktabs-preserved")) container = node; ts.forEachChild(node, visit); }
    visit(file);
    expect(container).toBeDefined();
    expect(container!.getText(file)).toContain("<WorkTabPanel");
    expect(container!.openingElement.attributes.properties.some((attribute) => ts.isJsxAttribute(attribute) && attribute.name.getText(file) === "hidden")).toBe(true);
    for (let parent = container!.parent; parent; parent = parent.parent) {
      if (ts.isConditionalExpression(parent)) expect(parent.condition.getText(file)).not.toBe("topSearchResults");
    }
  });
});

// review 2026-09-24: จอแบบยื่นภาษี (50 ทวิ แก้ผู้จ่าย/ผู้รับเงิน) ต้องอยู่ในการ์ดถามก่อนปิดแท็บ/เปลี่ยนบริษัท/ออกจากระบบ เหมือนจอบัญชีแยกประเภท
describe("unsaved-data guard covers tax filing tabs", () => {
  it("tracks ledger and Thai tax routes, not other screens", async () => {
    const { isThaiTaxRoute } = await import("@/lib/thai-tax");
    const code = ts.transpileModule(functionSource("guardsUnsavedRoute"), { compilerOptions: { target: ts.ScriptTarget.ES2022, module: ts.ModuleKind.None } }).outputText;
    const guards = new Function("isGeneralLedgerRoute", "isThaiTaxRoute", `${code}; return guardsUnsavedRoute;`)(isGeneralLedgerRoute, isThaiTaxRoute) as (route: string) => boolean;
    expect(guards("/gl/journal/jv")).toBe(true);
    expect(guards("/report/whtcertificate")).toBe(true);
    expect(guards("/report/wht-reports")).toBe(true);
    expect(guards("/product")).toBe(false);
    expect(source).toContain("!guardsUnsavedRoute(detail.route)");
  });
  it("the tax workbench reports unsaved 50 ทวิ edits to the main menu", () => {
    const workbench = readFileSync(fileURLToPath(new URL("../tax/tax-filing-workbench.tsx", import.meta.url)), "utf8");
    expect(workbench).toMatch(/const reportWhtDirty = useCallback\(\(dirty: boolean\) => \{ whtDirtyRef\.current = dirty; setWhtDirty\(dirty\); \}, \[\]\);/);
    expect(workbench).toContain("useDirtyGuard(route, whtDirty);");
  });
});
