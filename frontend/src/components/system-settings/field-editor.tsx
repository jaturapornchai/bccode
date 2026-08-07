"use client";

// Field Editor - extracted from system-settings-screen.tsx
// Core dispatcher component that renders the appropriate editor for each field type

import { FormEvent } from "react";
import type { LanguageCode } from "@/lib/i18n";
import type { AuthSession, WorkspaceSession } from "@/lib/workspace-models";
import type { BackendLanguageDictionary } from "@/lib/backend-language";
import { backendText } from "@/lib/backend-language";
import type {
  SystemSettingConfig,
  SystemSettingField,
} from "@/lib/system-setting-screens";
import { DateField } from "@/components/ui/date-time-field";
import { Input } from "@/components/ui/input";
import { cn } from "@/lib/utils";
import {
  type SettingRecord,
  type FormState,
  type DateTimeScope,
  isRecord,
  stringValue,
} from "./types";
import {
  uiEn,
  uiText,
  uiBackendKey,
  fieldLabel,
  optionLabel,
  isPermissionAccessRulesField,
  isPermissionCodesField,
  isApprovalCodesField,
  isEmployeeCodeField,
  isEmployeeNameField,
  isProductVariantStructuredField,
  isBranchStructuredSettingField,
  isBranchLatitudeField,
  isBranchLongitudeField,
  isColorHexField,
  radioFormValue,
  radioValueToFormValue,
  optionValueToFormValue,
  normalizeHexColor,
  languageName,
  type UiTextKey,
} from "./utils";

// Import field editors from extracted modules
import { ThailandAddressFieldEditor } from "./field-editors/thailand-address-editor";
import {
  ImageUploadFieldEditor,
  ImageGalleryFieldEditor,
} from "./field-editors/image-upload-editor";
import {
  BranchCoordinatePairEditor,
  CompanyMultiSelectFieldEditor,
} from "./field-editors/branch-company-selectors";
import {
  PermissionMatrixEditor,
  PermissionLinkMultiSelectEditor,
} from "./field-editors/permission-editors";
import { HoldingScopeRulesEditor } from "./field-editors/holding-scope-editor";
import {
  ProductVariantStructuredFieldEditor,
  BranchStructuredSettingEditor,
  TimeSaleListEditor,
  BankAccountsEditor,
} from "./field-editors/structured-field-editors";
import {
  StringListFieldEditor,
  ComboFieldEditor,
  MasterPickerFieldEditor,
  MasterMultiPickerFieldEditor,
} from "./field-editors/misc-field-editors";
import {
  LanguageConfigsEditor,
  LanguageListEditor,
  LanguageFlag,
  nameEditorLanguageCodes,
  setDefaultLanguageConfig,
} from "./field-editors/language-editors";
import { ApprovalSettingEditor } from "@/app/system-settings/approval-setting-editor";

// ─── FieldEditor Component ───────────────────────────────────────────────────

export function FieldEditor({
  auth,
  config,
  dateTimeScope,
  dictionary,
  field,
  form,
  language,
  setForm,
  workspace,
}: {
  auth: AuthSession | null;
  config: SystemSettingConfig;
  dateTimeScope: DateTimeScope;
  dictionary: BackendLanguageDictionary;
  field: SystemSettingField;
  form: FormState;
  language: LanguageCode;
  setForm: (form: FormState) => void;
  workspace: WorkspaceSession | null;
}) {
  const label = fieldLabel(field, language, config, dictionary);
  const helper =
    field.helper?.[language] ?? field.helper?.en ?? field.helper?.th;
  const value = form[field.key];

  if (isProductVariantStructuredField(config, field)) {
    return (
      <ProductVariantStructuredFieldEditor
        field={field}
        form={form}
        label={label}
        language={language}
        setForm={setForm}
      />
    );
  }

  if (config.slug === "permissiondefinition" && isPermissionAccessRulesField(field)) {
    return (
      <PermissionMatrixEditor
        dateTimeScope={dateTimeScope}
        dictionary={dictionary}
        form={form}
        language={language}
        setForm={setForm}
      />
    );
  }

  if (config.slug === "approvalsetting" && field.key === "approvals") {
    return (
      <ApprovalSettingEditor
        dictionary={dictionary}
        form={form}
        setForm={setForm}
      />
    );
  }

  if (config.slug === "permissionlink" && isEmployeeCodeField(field)) {
    return (
      <label className="grid gap-1 text-sm font-semibold">
        <span>{label}</span>
        <Input value={String(value ?? "")} readOnly disabled aria-readonly />
      </label>
    );
  }

  if (config.slug === "permissionlink" && isEmployeeNameField(field)) {
    return (
      <label className="grid gap-1 text-sm font-semibold">
        <span>{label}</span>
        <Input value={String(value ?? "")} readOnly disabled aria-readonly />
      </label>
    );
  }

  if (
    (config.slug === "permissionlink" || config.slug === "permissiongroup") &&
    (isPermissionCodesField(field) || isApprovalCodesField(field))
  ) {
    return (
      <PermissionLinkMultiSelectEditor
        config={config}
        auth={auth}
        dictionary={dictionary}
        field={field}
        form={form}
        language={language}
        normalizeRecords={(payload) => {
          if (!isRecord(payload)) return Array.isArray(payload) ? payload.filter(isRecord) : [];
          const data = payload.data;
          if (Array.isArray(data)) return data.filter(isRecord);
          if (isRecord(data) && Array.isArray(data.data)) return data.data.filter(isRecord);
          return [];
        }}
        setForm={setForm}
        workspace={workspace}
      />
    );
  }

  if (field.type === "holding-scope-rules") {
    return (
      <HoldingScopeRulesEditor
        auth={auth}
        field={field}
        form={form}
        label={label}
        language={language}
        setForm={setForm}
        workspace={workspace}
      />
    );
  }

  if (field.type === "thai-address") {
    return (
      <ThailandAddressFieldEditor
        backendUrl={auth?.backendUrl}
        copyFromPrefix={field.copyFromPrefix}
        form={form}
        language={language}
        prefix={field.key}
        setForm={setForm}
      />
    );
  }

  if (field.type === "checkbox") {
    return (
      <label className="flex min-h-12 items-center gap-2 rounded-2xl border border-border bg-background p-3 text-sm font-semibold">
        <input
          className="size-4 accent-primary"
          type="checkbox"
          checked={Boolean(value)}
          onChange={(event) =>
            setForm({ ...form, [field.key]: event.target.checked })
          }
        />
        <span>{label}</span>
      </label>
    );
  }

  if (field.type === "radio") {
    const options = field.options ?? [];
    const selectedValue = radioFormValue(value, field);
    const hasUnknownValue =
      Boolean(selectedValue) &&
      options.length > 0 &&
      !options.some((option) => option.value === selectedValue);
    const unknownValueText =
      language === "th"
        ? `ค่าปัจจุบันไม่ตรงกับบทบาทที่ระบบรองรับ: ${selectedValue} กรุณาเลือกใหม่`
        : `Current value is not a supported role: ${selectedValue}. Please choose a valid role.`;
    return (
      <section className="grid gap-1 text-sm font-semibold">
        <span>
          {label}
          {field.required ? " *" : ""}
        </span>
        <div
          className="grid w-full auto-rows-fr gap-2"
          style={{ gridTemplateColumns: "repeat(auto-fit, minmax(min(100%, 8rem), 1fr))" }}
          role="radiogroup"
          aria-label={label}
        >
          {options.map((option) => {
            const checked = selectedValue === option.value;
            return (
              <label
                className={cn(
                  "flex h-10 cursor-pointer items-center justify-center gap-2 rounded-2xl border px-3 transition-colors",
                  checked
                    ? "border-primary bg-primary/10 text-primary"
                    : "border-border bg-background text-foreground hover:bg-muted/60",
                )}
                key={option.value}
              >
                <input
                  className="size-4 accent-primary"
                  name={field.key}
                  type="radio"
                  value={option.value}
                  checked={checked}
                  onChange={() =>
                    setForm({
                      ...form,
                      [field.key]: radioValueToFormValue(option.value, field),
                    })
                  }
                />
                <span className="min-w-0 text-center leading-tight">{optionLabel(option, language)}</span>
              </label>
            );
          })}
        </div>
        {hasUnknownValue ? (
          <span className="text-xs font-medium text-destructive">
            {unknownValueText}
          </span>
        ) : null}
      </section>
    );
  }

  if (field.type === "names") {
    const names = isRecord(value) ? value : {};
    const editorLanguages = nameEditorLanguageCodes(form, config, language, workspace);
    return (
      <section className="grid gap-1 rounded-2xl border border-border bg-background p-2 md:col-span-2">
        <div className="text-sm font-semibold">
          {label}
          {field.required ? " *" : ""}
        </div>
        <div
          className={cn(
            "grid gap-2",
            editorLanguages.length > 1 && "md:grid-cols-2",
          )}
        >
          {editorLanguages.map((code, index) => {
            const currentValue =
              typeof names[code] === "string" ? String(names[code]) : "";
            return (
              <label className="grid gap-1 text-sm font-semibold" key={code}>
                <span className="flex min-w-0 items-center gap-2 text-xs text-muted-foreground">
                  <LanguageFlag code={code} />
                  <span className="truncate">
                    {index === 0
                      ? language === "th"
                        ? "ภาษาแรก"
                        : "Primary"
                      : languageName(code, language)}
                  </span>
                  <span className="uppercase">{code}</span>
                </span>
                {field.multiline ? (
                  <textarea
                    className="min-h-20 w-full rounded-2xl border border-input bg-background px-3 py-2 text-sm text-foreground shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
                    value={currentValue}
                    onChange={(event) =>
                      setForm({
                        ...form,
                        [field.key]: { ...names, [code]: event.target.value },
                      })
                    }
                  />
                ) : (
                  <Input
                    value={currentValue}
                    onChange={(event) =>
                      setForm({
                        ...form,
                        [field.key]: { ...names, [code]: event.target.value },
                      })
                    }
                  />
                )}
              </label>
            );
          })}
        </div>
      </section>
    );
  }

  if (field.type === "language-configs") {
    return (
      <LanguageConfigsEditor
        form={form}
        language={language}
        label={label}
        setForm={setForm}
      />
    );
  }

  if (field.type === "language-list") {
    return (
      <LanguageListEditor
        form={form}
        field={field}
        language={language}
        label={label}
        setForm={setForm}
      />
    );
  }

  if (isBranchStructuredSettingField(config, field)) {
    return (
      <BranchStructuredSettingEditor
        dictionary={dictionary}
        field={field}
        form={form}
        label={label}
        language={language}
        setForm={setForm}
      />
    );
  }

  if (field.type === "time-sale-list") {
    return (
      <TimeSaleListEditor
        dateTimeScope={dateTimeScope}
        field={field}
        form={form}
        label={label}
        language={language}
        setForm={setForm}
      />
    );
  }

  if (field.type === "bank-accounts") {
    return (
      <BankAccountsEditor
        field={field}
        form={form}
        label={label}
        language={language}
        setForm={setForm}
      />
    );
  }

  if (field.type === "string-list") {
    return (
      <StringListFieldEditor
        field={field}
        form={form}
        label={label}
        language={language}
        setForm={setForm}
      />
    );
  }

  if (field.type === "textarea" || field.type === "json") {
    return (
      <label className="grid gap-1 text-sm font-semibold md:col-span-2">
        <span>
          {label}
          {field.required ? " *" : ""}
        </span>
        <textarea
          className="min-h-24 w-full rounded-2xl border border-input bg-background px-3 py-2 text-sm text-foreground shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
          value={String(value ?? "")}
          onChange={(event) =>
            setForm({ ...form, [field.key]: event.target.value })
          }
          placeholder={field.placeholder ?? (field.type === "json" ? "[]" : undefined)}
        />
      </label>
    );
  }

  if (field.type === "combo") {
    return (
      <ComboFieldEditor
        field={field}
        form={form}
        label={label}
        language={language}
        setForm={setForm}
      />
    );
  }

  if (field.type === "master-picker") {
    return (
      <MasterPickerFieldEditor
        auth={auth}
        field={field}
        form={form}
        label={label}
        language={language}
        setForm={setForm}
      />
    );
  }

  if (field.type === "master-multi-picker") {
    return (
      <MasterMultiPickerFieldEditor
        auth={auth}
        field={field}
        form={form}
        label={label}
        language={language}
        setForm={setForm}
      />
    );
  }

  if (field.type === "image-upload") {
    return (
      <ImageUploadFieldEditor
        auth={auth}
        field={field}
        form={form}
        label={label}
        language={language}
        setForm={setForm}
      />
    );
  }

  if (field.type === "image-gallery") {
    return (
      <ImageGalleryFieldEditor
        auth={auth}
        field={field}
        form={form}
        label={label}
        language={language}
        setForm={setForm}
      />
    );
  }

  if (field.type === "branch-multi-select") {
    // Branch access is inherited from company access in these CRUD screens.
    return null;
  }

  if (field.type === "company-multi-select") {
    return (
      <CompanyMultiSelectFieldEditor
        auth={auth}
        field={field}
        form={form}
        language={language}
        setForm={setForm}
        workspace={workspace}
      />
    );
  }

  if (isBranchLongitudeField(config, field)) {
    return null;
  }

  if (isBranchLatitudeField(config, field)) {
    return (
      <BranchCoordinatePairEditor
        config={config}
        form={form}
        language={language}
        setForm={setForm}
      />
    );
  }

  if (field.type === "select") {
    const onSelectChange = (nextValue: string) => {
      const nextForm = {
        ...form,
        [field.key]: optionValueToFormValue(nextValue, field),
      };
      if (config.kind === "company" && field.key === "settings.language") {
        nextForm["settings.languageconfigs"] = setDefaultLanguageConfig(
          nextForm["settings.languageconfigs"] ?? nextForm["settings.languageconfigs"],
          nextValue,
        );
      }
      setForm(nextForm);
    };
    return (
      <label className="grid gap-1 text-sm font-semibold">
        <span>
          {label}
          {field.required ? " *" : ""}
        </span>
        <select
          className="min-h-10 w-full rounded-2xl border border-input bg-background px-3 text-sm text-foreground shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
          value={String(value ?? "")}
          onChange={(event) => onSelectChange(event.target.value)}
        >
          <option value=""></option>
          {field.options?.map((item) => (
            <option key={item.value} value={item.value}>
              {optionLabel(item, language)}
            </option>
          ))}
        </select>
      </label>
    );
  }

  if (field.type === "date") {
    return (
      <DateField
        language={language}
        yearType={dateTimeScope.calendarYearType}
        label={`${label}${field.required ? " *" : ""}`}
        timezoneLabel={
          dateTimeScope.timezonelabel || dateTimeScope.timezoneoffset
        }
        value={String(value ?? "")}
        onChange={(event) =>
          setForm({ ...form, [field.key]: event.target.value })
        }
      />
    );
  }

  if (isColorHexField(config, field)) {
    const hex = normalizeHexColor(value);
    return (
      <label className="grid gap-1 text-sm font-semibold">
        <span>{label}</span>
        <div className="flex w-full flex-wrap items-center gap-2 rounded-2xl border border-input bg-background px-2 py-2 shadow-sm">
          <input
            aria-label={label}
            className="h-9 w-12 rounded-lg border border-input bg-background"
            type="color"
            value={hex}
            onChange={(event) =>
              setForm({ ...form, [field.key]: event.target.value })
            }
          />
          <Input
            className="min-w-32 flex-1"
            value={String(value ?? "")}
            onChange={(event) =>
              setForm({ ...form, [field.key]: event.target.value })
            }
            placeholder="#000000"
          />
        </div>
      </label>
    );
  }

  // Default: text/number input
  return (
    <label className="grid gap-1 text-sm font-semibold">
      <span>
        {label}
        {field.required ? " *" : ""}
      </span>
      <Input
        type={field.type === "number" ? "number" : "text"}
        value={String(value ?? "")}
        readOnly={field.readOnly}
        disabled={field.readOnly}
        aria-readonly={field.readOnly}
        onChange={(event) => {
          if (field.readOnly) return;
          setForm({ ...form, [field.key]: event.target.value });
        }}
        placeholder={field.placeholder}
      />
      {helper ? (
        <span className="text-xs font-medium leading-snug text-muted-foreground">
          {helper}
        </span>
      ) : null}
    </label>
  );
}
