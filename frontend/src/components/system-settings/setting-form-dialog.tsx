"use client";

// Setting Form Dialog - extracted from system-settings-screen.tsx
// Modal/inline form for creating and editing system setting records

import { FormEvent } from "react";
import { Loader2, Save, X } from "lucide-react";
import type { LanguageCode } from "@/lib/i18n";
import type { AuthSession, WorkspaceSession } from "@/lib/workspace-models";
import type { BackendLanguageDictionary } from "@/lib/backend-language";
import type {
  SystemSettingConfig,
  SystemSettingField,
} from "@/lib/system-setting-screens";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { cn } from "@/lib/utils";
import {
  type SettingRecord,
  type FormState,
  type DateTimeScope,
} from "./types";
import {
  uiEn,
  systemSettingTitle,
  fieldGridItemClass,
  isBranchLongitudeField,
  isEmailLike,
  type UiTextKey,
} from "./utils";
import { FieldEditor } from "./field-editor";

// ─── SettingFormDialog Component ─────────────────────────────────────────────

export function SettingFormDialog({
  auth,
  config,
  dateTimeScope,
  dictionary,
  editing,
  form,
  inline = false,
  language,
  onClose,
  onSubmit,
  saving,
  setForm,
  text,
  workspace,
}: {
  auth: AuthSession | null;
  config: SystemSettingConfig;
  dateTimeScope: DateTimeScope;
  dictionary: BackendLanguageDictionary;
  editing: SettingRecord | null;
  form: FormState;
  inline?: boolean;
  language: LanguageCode;
  onClose: () => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  saving: boolean;
  setForm: (update: FormState | ((current: FormState) => FormState)) => void;
  text: (key: UiTextKey) => string;
  workspace: WorkspaceSession | null;
}) {
  const isProductCategoryForm = config.slug === "productcategorygroupselectscreen";
  const categoryUsesColor = form.useimageorcolor === true || String(form.useimageorcolor).toLowerCase() === "true";
  const shouldRenderField = (field: SystemSettingField) => {
    if (!isProductCategoryForm) return true;
    if (field.key === "groupnumber" || field.key === "parentguid") return false;
    if (field.key === "colorselecthex") return categoryUsesColor;
    if (field.key === "imageuri" || field.key === "coveruri") return !categoryUsesColor;
    return true;
  };
  const formElement = (
    <form
      className={cn(
        "grid gap-3 rounded-2xl border border-border bg-card p-3 text-foreground shadow-sm",
        inline
          ? "h-full min-h-0 w-full grid-rows-[auto_minmax(0,1fr)] overflow-hidden"
          : "max-h-[calc(100dvh-24px)] w-[min(860px,calc(100vw-24px))] overflow-hidden shadow-xl",
      )}
      onSubmit={onSubmit}
      role={inline ? undefined : "dialog"}
      aria-modal={inline ? undefined : true}
      aria-label={editing ? text("edit") : text("newItem")}
    >
      <header className="flex min-w-0 flex-wrap items-start justify-between gap-2">
        <div className="min-w-0">
          <h2 className="break-words text-lg font-semibold leading-snug">
            {editing ? text("edit") : text("newItem")}:{" "}
            {systemSettingTitle(config, language, dictionary)}
          </h2>
          <div className="flex flex-wrap gap-1 text-xs text-muted-foreground">
            {isProductCategoryForm ? (
              <Badge variant="outline" className="text-[11px]">
                {language === "th" ? "ชุด " : "Set "}
                {Number(form.groupnumber ?? 0) || "-"}
                {form.parentguid
                  ? language === "th" ? " · หมวดย่อย" : " · Subcategory"
                  : language === "th" ? " · หมวดหลัก" : " · Root category"}
              </Badge>
            ) : null}
            {config.slug === "department" ||
            config.kind === "restaurant-setting" ? (
              <Badge variant="outline" className="text-[11px]">
                {text("branch")}:{" "}
                {dateTimeScope.branchcode ||
                  dateTimeScope.branchguid ||
                  dateTimeScope.key}
              </Badge>
            ) : null}
          </div>
        </div>
        {inline ? (
          <div className="flex w-full shrink-0 flex-wrap items-center justify-end gap-2 sm:w-auto" data-testid="inline-form-actions">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={onClose}
              disabled={saving}
            >
              {text("cancel")}
            </Button>
            <Button type="submit" size="sm" disabled={saving}>
              {saving ? <Loader2 className="animate-spin" /> : <Save />}
              {text("save")}
            </Button>
          </div>
        ) : (
          <Button
            type="button"
            variant="outline"
            size="icon"
            onClick={onClose}
            disabled={saving}
            aria-label={text("close")}
          >
            <X />
          </Button>
        )}
      </header>

      <div
        className="grid min-h-0 gap-2 overflow-y-auto overscroll-contain pr-1"
      >
        {config.slug === "user" ? (
          <UserFormSections
            auth={auth}
            config={config}
            dateTimeScope={dateTimeScope}
            dictionary={dictionary}
            form={form}
            language={language}
            setForm={setForm}
            workspace={workspace}
          />
        ) : (
          <div className="grid content-start gap-2 md:grid-cols-2">
            {config.fields.filter(shouldRenderField).map((field) => {
              if (isBranchLongitudeField(config, field)) return null;
              return (
                <div
                  className={fieldGridItemClass(field, config)}
                  key={field.key}
                >
                  <FieldEditor
                    auth={auth}
                    config={config}
                    dateTimeScope={dateTimeScope}
                    dictionary={dictionary}
                    field={field}
                    form={form}
                    language={language}
                    setForm={setForm}
                    workspace={workspace}
                  />
                </div>
              );
            })}
          </div>
        )}
      </div>

      {inline ? null : (
        <footer className="flex flex-wrap items-center justify-end gap-2">
          <Button
            type="button"
            variant="outline"
            onClick={onClose}
            disabled={saving}
          >
            {text("cancel")}
          </Button>
          <Button type="submit" disabled={saving}>
            {saving ? <Loader2 className="animate-spin" /> : <Save />}
            {text("save")}
          </Button>
        </footer>
      )}
    </form>
  );
  if (inline) return formElement;
  return (
    <div className="dialog-backdrop" role="presentation">
      {formElement}
    </div>
  );
}

// ─── UserFormSections Component ──────────────────────────────────────────────

function UserFormSections({
  auth,
  config,
  dateTimeScope,
  dictionary,
  form,
  language,
  setForm,
  workspace,
}: {
  auth: AuthSession | null;
  config: SystemSettingConfig;
  dateTimeScope: DateTimeScope;
  dictionary: BackendLanguageDictionary;
  form: FormState;
  language: LanguageCode;
  setForm: (update: FormState | ((current: FormState) => FormState)) => void;
  workspace: WorkspaceSession | null;
}) {
  const fieldsByKey = new Map(config.fields.map((field) => [field.key, field]));
  const loginHint = isEmailLike(form.username)
    ? language === "th"
      ? "รหัสผู้ใช้ตอนนี้เป็นอีเมลแล้ว ผู้ใช้งานจะใช้ค่านี้เข้าสู่ระบบได้ ส่วนช่องอีเมลที่ลงทะเบียนจะใช้อีเมลเดียวกัน"
      : "The user code is already an email. This value is used for sign-in, and the registered email uses the same value."
    : language === "th"
      ? "ถ้าต้องการให้ผู้ใช้งานเข้าสู่ระบบด้วยอีเมล ให้กรอกอีเมลในช่องรหัสผู้ใช้ หรือ email ส่วนอีเมลที่ลงทะเบียนมีไว้สำหรับส่งอีเมลเท่านั้น"
      : "To let the user sign in with email, enter the email in User code or email. The registered email is only for sending email.";
  const sections = [
    {
      keys: ["avatar", "uid", "username", "userprofilename", "email"],
      title: language === "th" ? "บัญชีเข้าสู่ระบบ" : "Sign-in account",
      description: loginHint,
    },
    {
      keys: ["role", "permissionsets"],
      title: language === "th" ? "สิทธิ์ผู้ใช้งาน" : "User role",
      description:
        language === "th"
          ? "เลือกระดับสิทธิ์ แล้วเพิ่มสิทธิ์การใช้งานสำเร็จรูปได้หลายชุด (สร้างชุดที่ขั้น 2 สิทธิ์การใช้งาน)"
          : "Pick the access level, then add any number of reusable permission sets (managed in step 2).",
    },
    {
      keys: ["isaccessdisabled"],
      title: language === "th" ? "สถานะเข้าใช้งาน" : "Access status",
      description:
        language === "th"
          ? "เปิดหรือปิดการเข้าใช้งานของผู้ใช้นี้ ปิดชั่วคราวได้โดยไม่ต้องลบ"
          : "Enable or temporarily disable this user's access without deleting.",
    },
    {
      keys: ["accessscopes"],
      title:
        language === "th"
          ? "บริษัทและสาขาที่เข้าได้"
          : "Accessible companies and branches",
      description:
        language === "th"
          ? "ต้องเลือกบริษัทก่อน แล้วเลือกว่าจะเข้าได้ทุกสาขาหรือเฉพาะสาขาที่กำหนด"
          : "Select the company first, then choose all branches or specific branches.",
    },
    {
      keys: ["position", "department", "lineuserid", "linedisplayname"],
      title:
        language === "th" ? "ข้อมูลองค์กรและ LINE" : "Organization and LINE",
      description:
        language === "th"
          ? "ใช้สำหรับอ้างอิงตำแหน่ง แผนก และข้อมูล LINE ที่ผูกกับผู้ใช้งาน"
          : "Reference position, department, and LINE data linked to this user.",
    },
  ];

  return (
    <div className="grid gap-2">
      {sections.map((section) => {
        const fields = section.keys
          .map((key) => fieldsByKey.get(key))
          .filter((field): field is SystemSettingField => Boolean(field));
        if (fields.length === 0) return null;
        return (
          <section
            className="grid gap-2 rounded-2xl border border-border bg-background/70 p-2 relative overflow-hidden"
            key={section.title}
          >
            <div className="absolute left-0 top-0 bottom-0 w-1 bg-primary/40" />
            <header className="grid gap-0.5 pl-2">
              <h3 className="text-sm font-bold text-foreground">{section.title}</h3>
              <p className="text-xs leading-snug text-muted-foreground">
                {section.description}
              </p>
            </header>
            <div className="grid items-start gap-2 md:grid-cols-2">
              {fields.map((field) => (
                <div
                  className={fieldGridItemClass(field, config)}
                  key={field.key}
                >
                  <FieldEditor
                    auth={auth}
                    config={config}
                    dateTimeScope={dateTimeScope}
                    dictionary={dictionary}
                    field={field}
                    form={form}
                    language={language}
                    setForm={setForm}
                    workspace={workspace}
                  />
                </div>
              ))}
            </div>
          </section>
        );
      })}
    </div>
  );
}
