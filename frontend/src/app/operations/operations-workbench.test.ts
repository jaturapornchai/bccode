import { readFileSync } from "node:fs";
import { resolve } from "node:path";
import { describe, expect, it } from "vitest";

const source = readFileSync(resolve(process.cwd(), "src/app/operations/operations-workbench.tsx"), "utf8");

describe("OperationsApprovalWorkbench batch + paging contract", () => {
  it("renders selection checkboxes for approve, unapprove and cancel screens (no actionType gate)", () => {
    expect(source).toContain('type="checkbox"');
    expect(source).not.toMatch(/actionType === "approve" && \(\s*<th/);
    for (const action of ["approve", "unapprove", "cancel"]) {
      expect(source, action).toContain(`actionType === "${action}" && selectedDocnos.size > 0`);
    }
  });

  it("requires a reason before a batch cancel is submitted", () => {
    expect(source).toContain('pendingBatchAction === "cancel" && !reasonInput.trim()');
    expect(source).toContain("submitBatchCancel");
    expect(source).toContain("submitBatchUnapprove");
  });

  it("pages with limit/offset and shows page x of y", () => {
    expect(source).toContain("offset: (page - 1) * PAGE_SIZE");
    expect(source).toContain("limit: PAGE_SIZE");
    expect(source).toContain("ops_page_of");
    expect(source).toContain("pageCount > 1");
  });

  it("lists the documents that failed inside a partially successful batch", () => {
    expect(source).toContain("batchFailures");
    expect(source).toContain("ops_batch_partial_success");
    expect(source).toContain("gl_err_");
  });

  it("maps permission and unavailable-module errors to readable messages and never shows a raw key", () => {
    expect(source).toContain('erp_err_action_forbidden: { key: "ops_msg_you_do_not_have_permission"');
    expect(source).toMatch(/module_not_available: \{[ \t\r\n]*key: "module_not_available"/);
    expect(source).toContain("MESSAGES[key] ?? MESSAGES.action_failed");
  });

  it("labels row checkboxes with a language key rather than hard-coded English", () => {
    expect(source).toContain('tr("ops_select_doc"');
    expect(source).not.toContain("Select ${");
  });
});
