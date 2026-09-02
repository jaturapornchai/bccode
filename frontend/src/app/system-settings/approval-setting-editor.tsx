"use client";

import { Input } from "@/components/ui/input";
import {
  backendText,
  type BackendLanguageDictionary,
} from "@/lib/backend-language";
import {
  type FormState,
  type SettingRecord,
  isRecord,
  safeJsonParse,
} from "@/components/system-settings/types";

/* ------------------------------------------------------------------ */
/* Approval Setting Editor                                             */
/* ------------------------------------------------------------------ */

export type ApprovalTarget = "pr" | "po" | "quotation" | "sale_order";

export const approvalTargets: { key: ApprovalTarget; label: string }[] = [
  { key: "pr", label: "PR" },
  { key: "po", label: "PO" },
  { key: "quotation", label: "Quotation" },
  { key: "sale_order", label: "Sale Order" },
];

export function approvalsFromForm(value: unknown): SettingRecord {
  if (isRecord(value)) return { ...value };
  if (typeof value !== "string" || !value.trim()) return {};
  const parsed = safeJsonParse(value, {});
  return isRecord(parsed) ? { ...parsed } : {};
}

export function approvalLabelText(
  dictionary: BackendLanguageDictionary,
  key: string,
  fallback: string,
): string {
  const value = backendText(dictionary, key, fallback);
  return value === key ? fallback : value;
}

export function ApprovalSettingEditor({
  dictionary,
  form,
  readOnly = false,
  setForm,
}: {
  dictionary: BackendLanguageDictionary;
  form: FormState;
  readOnly?: boolean;
  setForm?: (update: FormState | ((current: FormState) => FormState)) => void;
}) {
  const approvals = approvalsFromForm(form.approvals);

  function updateApproval(
    target: ApprovalTarget,
    key: "enabled" | "approval_role" | "max_approval_amount",
    value: boolean | number,
  ) {
    if (readOnly || !setForm) return;
    const nextApprovals = approvalsFromForm(form.approvals);
    const currentApproval = isRecord(nextApprovals[target])
      ? { ...nextApprovals[target] }
      : {};
    currentApproval[key] = value;
    nextApprovals[target] = currentApproval;
    setForm({ ...form, approvals: nextApprovals });
  }

  return (
    <section className="grid gap-3 rounded-2xl border border-border bg-background p-3 md:col-span-2">
      <h3 className="text-base font-semibold">
        {backendText(dictionary, "approval", "Approval")}
      </h3>
      <div className="grid gap-2 md:grid-cols-2">
        {approvalTargets.map((target) => {
          const approval: SettingRecord = isRecord(approvals[target.key])
            ? (approvals[target.key] as SettingRecord)
            : {};
          return (
            <div
              className="grid gap-2 rounded-xl border border-border bg-card p-2"
              key={target.key}
            >
              <label className="flex items-center gap-2 text-sm font-semibold cursor-pointer">
                <input
                  className="size-4 accent-primary"
                  type="checkbox"
                  checked={Boolean(approval.enabled)}
                  disabled={readOnly}
                  onChange={(event) =>
                    updateApproval(target.key, "enabled", event.target.checked)
                  }
                />
                <span>{target.label}</span>
              </label>
              <label className="grid gap-1 text-xs font-semibold">
                <span>
                  {approvalLabelText(
                    dictionary,
                    "approval_role",
                    "Approval role",
                  )}
                </span>
                <Input
                  type="number"
                  min="0"
                  value={String(approval.approval_role ?? "")}
                  disabled={readOnly || !approval.enabled}
                  onChange={(event) =>
                    updateApproval(
                      target.key,
                      "approval_role",
                      Number(event.target.value || 0),
                    )
                  }
                />
              </label>
              <label className="grid gap-1 text-xs font-semibold">
                <span>
                  {approvalLabelText(
                    dictionary,
                    "max_approval_amount_label",
                    "Maximum approval amount",
                  )}
                </span>
                <Input
                  type="number"
                  min="0"
                  value={String(approval.max_approval_amount ?? "")}
                  disabled={readOnly || !approval.enabled}
                  onChange={(event) =>
                    updateApproval(
                      target.key,
                      "max_approval_amount",
                      Number(event.target.value || 0),
                    )
                  }
                />
              </label>
            </div>
          );
        })}
      </div>
    </section>
  );
}
