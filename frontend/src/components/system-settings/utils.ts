// System Settings utility functions extracted from the god file
// These are pure functions used across system-settings panels and editors

import type { AuthSession, WorkspaceSession } from "@/lib/workspace-models";
import { getAuthSession } from "@/lib/client-auth-session";
import { LANGUAGES, type LanguageCode } from "@/lib/i18n";
import type { BackendLanguageDictionary } from "@/lib/backend-language";
import { backendText } from "@/lib/backend-language";
import {
  systemSettingLabel,
  type SystemSettingConfig,
  type SystemSettingField,
  type SystemSettingOption,
} from "@/lib/system-setting-screens";
import {
  type SettingRecord,
  isRecord,
  stringValue,
  getByPath,
  localizedValue,
  booleanLikeValue,
} from "./types";

// ─── UI Text Constants ───────────────────────────────────────────────────────

export const uiEn = {
  active: "Active",
  add: "Add",
  addItem: "Add item",
  all: "All",
  accessEnabled: "Can access",
  accessStatus: "Access status",
  accessTemporarilyDisabled: "Temporarily disabled",
  allBranches: "All branches",
  branch: "Branch",
  cancel: "Cancel",
  close: "Close",
  company: "Company",
  copyNow: "Copy now",
  copyMondaySchedule: "Copy Monday time",
  creator: "Creator",
  creatorAlwaysEnabled: "Creator can always access.",
  creatorCannotDelete: "Creator cannot be deleted.",
  selfPermissionCannotEdit: "You cannot edit your own permission.",
  delete: "Delete",
  deleteConfirm: "Confirm delete this item?",
  edit: "Edit",
  empty: "No data",
  emptyHint: "Add a new item or change the search term.",
  readOnlyEmptyHint: "No data found from the real database.",
  enableAccess: "Enable access",
  enabled: "Enabled",
  errorRequired: "Please fill required fields.",
  fullDay: "Full day",
  id: "ID",
  jsonInvalid: "JSON format is invalid.",
  details: "Details",
  items: "items",
  list: "List",
  loading: "Loading data",
  manual: "Manual",
  newItem: "New item",
  findStandardUnits: "Find standard units",
  standardUnits: "Standard units",
  addSelected: "Add selected",
  clearSelection: "Clear selection",
  noStandardUnits: "No standard units to add.",
  selectAll: "Select all",
  selected: "Selected",
  off: "Closed",
  preview: "Preview",
  quickAdd: "Quick add",
  range: "Range",
  readAccess: "Access",
  refresh: "Refresh",
  removeRange: "Remove range",
  requestFailed: "Request failed.",
  resetPassword: "Send reset link",
  resetPasswordConfirm: "Send this user a one-time password reset link?",
  resetPasswordDone: "Password reset link sent.",
  save: "Save",
  saveFailed: "Save failed",
  saveSucceeded: "Saved successfully",
  saved: "Saved.",
  search: "Search",
  selfOnly: "Own data only",
  selectSourceShop: "Select source shop",
  sourceEnvironment: "Source environment",
  sourcePro: "PRO",
  sourceShop: "Source shop",
  sourceUat: "UAT",
  startTime: "Start",
  status: "Status",
  targetShop: "Target shop",
  timezone: "Timezone",
  endTime: "End",
  timeInvalid: "Start time must be before end time.",
  timeOverlap: "Time ranges overlap.",
  total: "Total",
  totalRanges: "Ranges",
  updated: "Updated",
  updateAccess: "Edit",
  workTime: "Work time",
  writeAccess: "Add",
  afternoonShift: "Afternoon",
  eveningShift: "Evening",
  morningShift: "Morning",
  standardTwoShifts: "2 shifts",
  temporarilyDisableAccess: "Disable temporarily",
} as const;

export type UiTextKey = keyof typeof uiEn;

export const uiText: Partial<
  Record<LanguageCode, Partial<Record<UiTextKey, string>>>
> = {
  th: {
    active: "ใช้งาน",
    add: "เพิ่ม",
    addItem: "เพิ่มรายการ",
    all: "ทั้งหมด",
    accessEnabled: "เข้าใช้งานได้",
    accessStatus: "สถานะเข้าใช้งาน",
    accessTemporarilyDisabled: "เข้าใช้งานไม่ได้ชั่วคราว",
    allBranches: "ใช้กับทุกสาขา",
    branch: "สาขา",
    cancel: "ยกเลิก",
    close: "ปิด",
    company: "บริษัท",
    copyNow: "โอนข้อมูล",
    copyMondaySchedule: "คัดลอกเวลาจันทร์",
    creator: "ผู้สร้าง",
    creatorAlwaysEnabled: "ผู้สร้างเข้าใช้งานได้ตลอด",
    creatorCannotDelete: "ผู้สร้างไม่สามารถลบได้",
    selfPermissionCannotEdit: "ไม่สามารถแก้ไขสิทธิ์ของตัวเองได้",
    delete: "ลบ",
    deleteConfirm: "ต้องการลบจริงหรือไม่",
    edit: "แก้ไข",
    empty: "ไม่มีข้อมูล",
    emptyHint: "เพิ่มรายการใหม่ หรือเปลี่ยนคำค้นหา",
    readOnlyEmptyHint: "ไม่พบข้อมูลจากฐานข้อมูลจริง",
    enableAccess: "ให้เข้าใช้งานได้",
    enabled: "เปิดใช้งาน",
    errorRequired: "กรุณากรอกช่องที่จำเป็น",
    fullDay: "ทั้งวัน",
    id: "รหัส",
    jsonInvalid: "รูปแบบ JSON ไม่ถูกต้อง",
    details: "รายละเอียด",
    items: "รายการ",
    list: "รายการ",
    loading: "กำลังโหลดข้อมูล",
    manual: "คู่มือ",
    newItem: "รายการใหม่",
    findStandardUnits: "ค้นหาหน่วยนับมาตรฐาน",
    standardUnits: "หน่วยนับมาตรฐาน",
    addSelected: "เพิ่มรายการที่เลือก",
    clearSelection: "ล้างการเลือก",
    noStandardUnits: "ไม่พบหน่วยนับมาตรฐานที่เพิ่มได้",
    selectAll: "เลือกทั้งหมด",
    selected: "เลือกแล้ว",
    off: "หยุด",
    preview: "ตรวจสอบก่อนโอน",
    quickAdd: "เพิ่มด่วน",
    range: "ช่วง",
    readAccess: "เข้าถึง",
    refresh: "โหลดใหม่",
    removeRange: "ลบช่วงเวลา",
    requestFailed: "เรียกข้อมูลไม่สำเร็จ",
    resetPassword: "ส่งลิงก์รีเซ็ตรหัสผ่าน",
    resetPasswordConfirm: "ยืนยันส่งลิงก์รีเซ็ตรหัสผ่านแบบใช้ครั้งเดียวให้ผู้ใช้นี้?",
    resetPasswordDone: "ส่งลิงก์รีเซ็ตรหัสผ่านแล้ว",
    save: "บันทึก",
    saveFailed: "บันทึกไม่สำเร็จ",
    saveSucceeded: "บันทึกสำเร็จ",
    saved: "บันทึกแล้ว",
    search: "ค้นหา",
    selfOnly: "เห็นข้อมูลตัวเองเท่านั้น",
    selectSourceShop: "เลือก shop ต้นทาง",
    sourceEnvironment: "ฐานข้อมูลต้นทาง",
    sourcePro: "PRO ใช้งานจริง",
    sourceShop: "Shop ต้นทาง",
    sourceUat: "UAT ทดสอบ",
    startTime: "เริ่ม",
    status: "สถานะ",
    targetShop: "Shop ปลายทาง",
    timezone: "เขตเวลา",
    endTime: "สิ้นสุด",
    timeInvalid: "เวลาเริ่มต้องน้อยกว่าเวลาจบ",
    timeOverlap: "ช่วงเวลาทับซ้อนกัน",
    total: "ทั้งหมด",
    totalRanges: "ช่วงเวลา",
    updated: "อัปเดต",
    updateAccess: "แก้ไขได้",
    workTime: "เวลาทำงาน",
    writeAccess: "เพิ่มได้",
    afternoonShift: "บ่าย",
    eveningShift: "เย็น",
    morningShift: "เช้า",
    standardTwoShifts: "2 ช่วง",
    temporarilyDisableAccess: "ปิดชั่วคราว",
  },
  en: uiEn,
};

export const uiBackendKeys: Partial<Record<UiTextKey, string>> = {
  accessEnabled: "access_enabled",
  accessStatus: "access_status",
  accessTemporarilyDisabled: "access_temporarily_disabled",
  active: "active",
  add: "add",
  all: "all",
  allBranches: "allbranches",
  branch: "branch",
  cancel: "cancel",
  close: "close",
  company: "company",
  creator: "creator",
  creatorAlwaysEnabled: "creator_always_enabled",
  creatorCannotDelete: "creator_cannot_delete",
  selfPermissionCannotEdit: "self_permission_cannot_edit",
  delete: "delete",
  edit: "edit",
  empty: "empty_data",
  emptyHint: "press_add_item_to_start",
  enableAccess: "enable_access",
  enabled: "enabled",
  id: "code",
  loading: "loading",
  manual: "manual",
  newItem: "new_item",
  findStandardUnits: "find_standard_units",
  standardUnits: "standard_units",
  addSelected: "add_selected",
  clearSelection: "clear_selection",
  noStandardUnits: "no_standard_units",
  selectAll: "select_all",
  selected: "selected",
  readAccess: "access",
  refresh: "refresh",
  resetPassword: "reset_password",
  resetPasswordConfirm: "reset_password_confirm",
  resetPasswordDone: "reset_password_done",
  save: "save",
  search: "search",
  selfOnly: "own_data_only",
  status: "status",
  timezone: "timezone",
  total: "total",
  updated: "updated",
  updateAccess: "can_edit",
  temporarilyDisableAccess: "temporarily_disable_access",
  writeAccess: "can_add",
};

export const systemSettingBackendKeys: Record<string, string> = {
  branch: "branch",
  businesstypescreen: "business_type",
  company: "company",
  department: "department",
  employee: "employee",
  holidayscreen: "holiday",
  permissiondefinition: "permissiondefinition",
  permissionlink: "permissionlink",
  approvalsetting: "approvalsetting",
  user: "user",
  workdayscreen: "work_day",
};

export const fieldBackendKeys: Record<string, string> = {
  "branch.basecurrency": "basecurrency",
  "branch.code": "branchcode",
  "branch.companyregistrationno": "companyregistrationno",
  "branch.dateformat": "dateformat",
  "branch.decimaldocument": "decimaldocument",
  "branch.decimalprice": "decimalprice",
  "branch.decimalquantity": "decimalquantity",
  "branch.isvatregistered": "vat_status",
  "branch.language": "default_language",
  "branch.machinetype": "machine_type",
  "branch.names": "branch_name",
  "branch.pointconfig": "point_config",
  "branch.timezone": "timezone",
  "branch.yeartype": "yeartype",
  "branch.languages": "select_data_language",
  "branch.businesstype": "business_type",
  "branch.pos.taxid": "company_taxid",
  "branch.pos.vatrate": "vat_rate",
  "branch.pos.isbom": "cut_stock_by_bom",
  "branch.pos.vattypepurchase": "vattype_purchase",
  "branch.pos.inquirytypepurchase": "inquirytype_purchase",
  "branch.pos.vattypesale": "vattype_sale",
  "branch.pos.inquirytypesale": "inquirytype_sale",
  "businesstypescreen.code": "code",
  "businesstypescreen.names": "business_type",
  "company.address": "company_address",
  "company.logo": "company_logo",
  "company.names": "company_name",
  "company.settings.basecurrency": "basecurrency",
  "company.settings.companyregistrationno": "companyregistrationno",
  "company.settings.country_code": "country_code",
  "company.settings.dateformat": "dateformat",
  "company.settings.decimaldocument": "decimaldocument",
  "company.settings.decimalprice": "decimalprice",
  "company.settings.decimalquantity": "decimalquantity",
  "company.settings.isvatregistered": "vat_status",
  "company.settings.isusebranch": "use_branch_system",
  "company.settings.isusedepartment": "use_department_system",
  "company.settings.language": "default_language",
  "company.settings.languageconfigs": "activelanguages",
  "company.settings.taxid": "taxid",
  "company.settings.timezone": "timezone",
  "company.settings.usebuddhistcalendar": "yeartype",
  "company.settings.vatrate": "vat_rate",
  "company.telephone": "telephone",
  "department.code": "department_code",
  "department.names": "department_name",
  "holidayscreen.date": "date",
  "holidayscreen.desc": "description",
  "user.isaccessdisabled": "access_status",
  "user.uid": "user_id_guid",
  "user.userprofilename": "user_name",
  "user.email": "registered_email",
  "user.role": "user_role",
  "user.position": "user_position",
  "user.department": "department",
  "user.accessscopes": "accessscopes",
  "user.lineuserid": "lineuserid",
  "user.linedisplayname": "linedisplayname",
  "permissiondefinition.permissioncode": "permissioncode",
  "permissiondefinition.permissionname": "permissionname",
  "permissiondefinition.scoperules": "scoperules",
  "permissiondefinition.accessrules": "accessrules",
  "approvalsetting.approvalcode": "approvalcode",
  "approvalsetting.approvalname": "approvalname",
  "approvalsetting.approvalrules": "approvalrules",
  "approvalsetting.approvals": "approval_permission",
  "permissionlink.employeecode": "user_employeecode",
  "permissionlink.employeename": "name",
  "permissionlink.scoperules": "scoperules",
  "permissionlink.permissioncodes": "permissioncodes",
  "permissionlink.approvalcodes": "approvalcodes",
};

export const fieldValueAliases: Record<string, string[]> = {
  "company.settings.languageconfigs": ["settings.languageconfigs"],
  "productunit.unitcode": ["unitcode"],
  "productunit.businesscodes": ["companyguids"],
  "employee.businesscodes": ["companyguids"],
  "user.businesscodes": ["companyguids"],
  "user.accessscopes": ["scoperules", "businesscodes", "companyguids"],
  "approvalsetting.approvalcode": ["approvalCode"],
  "approvalsetting.approvalname": ["approvalName"],
  "approvalsetting.isactive": ["isActive"],
  "approvalsetting.approvalrules": ["scoperules"],
  "permissiondefinition.permissioncode": ["permissionCode"],
  "permissiondefinition.permissionname": ["permissionName"],
  "permissiondefinition.isactive": ["isActive"],
  "permissiondefinition.scoperules": ["accessscopes"],
  "permissiondefinition.accessrules": ["branches"],
  "permissiongroup.groupcode": ["groupCode"],
  "permissiongroup.groupname": ["groupName"],
  "permissiongroup.isactive": ["isActive"],
  "permissiongroup.scoperules": ["accessscopes"],
  "permissiongroup.permissioncodes": ["permissionCodes"],
  "permissionlink.employeecode": ["employeeCode"],
  "permissionlink.employeename": ["employeeName"],
  "permissionlink.groupcode": ["groupCode"],
  "permissionlink.scoperules": ["accessscopes", "businesscodes", "companyguids"],
  "permissionlink.businesscodes": ["companyguids"],
  "permissionlink.permissioncodes": ["permissionCodes"],
  "permissionlink.approvalcodes": ["approvalCodes"],
  "branch.pos.taxid": ["pos.taxid"],
  "branch.yeartype": ["yeartype"],
  "productcategorygroupselectscreen.groupnumber": ["groupnumber"],
  "productcategorygroupselectscreen.parentguid": ["parentguid"],
  "productcategorylist.groupnumber": ["groupnumber"],
  "productcategorylist.parentguid": ["parentguid"],
};

// ─── Record Identity Functions ───────────────────────────────────────────────

export function firstRecordValue(...values: unknown[]): string {
  for (const value of values) {
    const text = stringValue(value);
    if (text) return text;
  }
  return "";
}

export function recordId(
  record: SettingRecord | null | undefined,
  config: SystemSettingConfig,
): string {
  if (!record) return "";
  return firstRecordValue(
    config.idField ? getByPath(record, config.idField) : undefined,
    record.guidfixed,
    record.guid,
    record.id,
    record._id,
    record.unitcode,
    record.unitCode,
    record.approvalcode,
    record.permissioncode,
    record.groupcode,
    record.employeecode,
    record.approvalCode,
    record.permissionCode,
    record.groupCode,
    record.employeeCode,
    record.code,
    record.holdingcode,
    record.providername,
  );
}

export function recordDisplayCode(
  record: SettingRecord | null | undefined,
  config: SystemSettingConfig,
): string {
  if (!record) return "";
  const codeField = config.fields.find((field) => field.businessCode);
  if (codeField) {
    const code = stringValue(getByPath(record, codeField.key));
    if (code) return code;
  }
  const idKey = (config.idField ?? "").toLowerCase();
  if (idKey && idKey !== "guidfixed" && idKey !== "guid") {
    return recordId(record, config);
  }
  return recordBusinessLookup(record, config) || recordId(record, config);
}

export function recordDetailId(
  record: SettingRecord | null | undefined,
  config: SystemSettingConfig,
): string {
  if (!record) return "";
  return firstRecordValue(
    config.idField ? getByPath(record, config.idField) : undefined,
    record.guidfixed,
    record.guid,
    record.id,
    record._id,
  );
}

export function recordBusinessLookup(
  record: SettingRecord | null | undefined,
  config: SystemSettingConfig,
): string {
  if (!record) return "";
  return firstRecordValue(
    config.fields?.[0]?.key ? getByPath(record, config.fields[0].key) : undefined,
    record.unitcode,
    record.unitCode,
    record.code,
    record.approvalcode,
    record.permissioncode,
    record.groupcode,
    record.employeecode,
    record.approvalCode,
    record.permissionCode,
    record.groupCode,
    record.employeeCode,
  );
}

export function recordMatchesBusinessLookup(
  record: SettingRecord,
  config: SystemSettingConfig,
  lookup: string,
): boolean {
  const target = lookup.trim().toUpperCase();
  if (!target) return false;
  const candidates = [
    config.fields?.[0]?.key ? getByPath(record, config.fields[0].key) : undefined,
    record.unitcode,
    record.unitCode,
    record.code,
    record.approvalcode,
    record.permissioncode,
    record.groupcode,
    record.employeecode,
    record.approvalCode,
    record.permissionCode,
    record.groupCode,
    record.employeeCode,
  ];
  return candidates.some((value) => stringValue(value).toUpperCase() === target);
}

export function recordTitle(
  record: SettingRecord,
  config: SystemSettingConfig | undefined,
  language: LanguageCode,
): string {
  if (config?.slug === "activelanguages")
    return systemSettingLabel(config, language);
  const names =
    getByPath(record, "names") ??
    getByPath(record, "name") ??
    getByPath(record, "desc");
  const localized = localizedValue(names, language);
  if (localized) return localized;
  return String(
    record.userprofilename ??
      record.name ??
      record.name1 ??
      record.permissionname ??
      record.approvalname ??
      record.employeename ??
      record.permissionName ??
      record.approvalName ??
      record.employeeName ??
      record.username ??
      record.providername ??
      record.code ??
      record.employeecode ??
      record.approvalcode ??
      record.permissioncode ??
      record.groupcode ??
      record.guidfixed ??
      config?.slug ??
      "",
  );
}

export function recordBranchCaption(record: SettingRecord): string {
  return (
    stringValue(record.branchcode) ||
    stringValue(record.branchguid) ||
    stringValue(record.branchkey)
  );
}

// ─── Record State Functions ──────────────────────────────────────────────────

export function shouldHydrateRecordDetail(config: SystemSettingConfig): boolean {
  return config.kind === "main-crud" || config.kind === "atlas";
}

export function isCreatorRecord(
  record: SettingRecord,
  workspace: WorkspaceSession | null,
): boolean {
  if (Boolean(record.iscreator)) return true;
  const creator = stringValue(
    record.createdby ??
      workspace?.shop.createdby ??
      getByPath(workspace?.shopInfo ?? {}, "createdby"),
  );
  const username = stringValue(record.username ?? record.email ?? record.code);
  return Boolean(
    creator && username && creator.toLowerCase() === username.toLowerCase(),
  );
}

export function isSelfUserRecord(
  record: SettingRecord,
  auth: AuthSession | null,
): boolean {
  const username = stringValue(
    record.username ?? record.email ?? record.code,
  ).toLowerCase();
  const identities = [auth?.username, auth?.profile?.email]
    .map((item) => stringValue(item).toLowerCase())
    .filter(Boolean);
  return Boolean(username && identities.includes(username));
}

export function userAccessDisabled(record: SettingRecord): boolean {
  return Boolean(record.isaccessdisabled ?? false);
}

export function isEmailLike(value: unknown): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(stringValue(value));
}

// ─── Label Functions ─────────────────────────────────────────────────────────

export function uiBackendKey(key: UiTextKey): string {
  return uiBackendKeys[key] ?? key;
}

export function systemSettingTitle(
  config: SystemSettingConfig,
  language: LanguageCode,
  dictionary: BackendLanguageDictionary,
): string {
  const fallback = systemSettingLabel(config, language);
  const backendKey = systemSettingBackendKeys[config.slug] ?? config.slug;
  const value = backendText(dictionary, backendKey, fallback);
  return value === backendKey || value === config.slug ? fallback : value;
}

export function fieldLabel(
  field: SystemSettingField,
  language: LanguageCode,
  config?: SystemSettingConfig,
  dictionary?: BackendLanguageDictionary,
): string {
  const fallback = field.label[language] ?? field.label.en ?? field.label.th;
  const backendKey = config
    ? (fieldBackendKeys[`${config.slug}.${field.key}`] ??
      fieldBackendKeys[field.key])
    : undefined;
  if (!backendKey || !dictionary) return fallback;
  const value = backendText(dictionary, backendKey, fallback);
  return value === backendKey || value === field.key ? fallback : value;
}

export function optionLabel(
  option: SystemSettingOption,
  language: LanguageCode,
): string {
  return (
    option.labels?.[language] ??
    option.labels?.en ??
    option.labels?.th ??
    option.label
  );
}

// ─── Misc Helpers ────────────────────────────────────────────────────────────

export function wait(milliseconds: number): Promise<void> {
  return new Promise((resolve) => setTimeout(resolve, milliseconds));
}

// ─── Storage Helpers ─────────────────────────────────────────────────────────

import { workspaceStorageKeys } from "@/lib/workspace-models";

export function readAuth(): AuthSession | null {
  return getAuthSession();
}

export function readWorkspace(): WorkspaceSession | null {
  try {
    const raw = localStorage.getItem(workspaceStorageKeys.workspace);
    if (!raw) return null;
    const workspace = JSON.parse(raw) as WorkspaceSession;
    return workspace?.shop?.holdingcode ? workspace : null;
  } catch {
    return null;
  }
}

// ─── Field Type Predicates ───────────────────────────────────────────────────

export function isPermissionAccessRulesField(field: SystemSettingField): boolean {
  return field.key === "accessrules" || field.key === "branches";
}

export function isPermissionCodesField(field: SystemSettingField): boolean {
  return field.key === "permissioncodes" || field.key === "permissionCodes";
}

export function isApprovalCodesField(field: SystemSettingField): boolean {
  return field.key === "approvalcodes" || field.key === "approvalCodes";
}

export function isEmployeeCodeField(field: SystemSettingField): boolean {
  return field.key === "employeecode" || field.key === "employeeCode";
}

export function isEmployeeNameField(field: SystemSettingField): boolean {
  return field.key === "employeename" || field.key === "employeeName";
}

const productVariantStructuredFieldKeys = new Set([
  "optiontiers",
  "skucombinations",
  "mediaassets",
  "specificationgroups",
  "importattributemaps",
  "integrationprofiles",
  "payloadexamples",
]);

export function isProductVariantStructuredField(
  config: SystemSettingConfig,
  field: SystemSettingField,
): boolean {
  return (
    config.slug === "productvariantmatrix" &&
    productVariantStructuredFieldKeys.has(field.key)
  );
}

export function isBranchStructuredSettingField(
  config: SystemSettingConfig,
  field: SystemSettingField,
): boolean {
  return (
    config.slug === "branch" &&
    field.type === "json" &&
    (field.key === "paymentrounding" || field.key === "pointconfig")
  );
}

export function isBranchLatitudeField(
  config: SystemSettingConfig,
  field: SystemSettingField,
): boolean {
  return config.slug === "branch" && field.key === "contact.latitude";
}

// ─── Record Value Helpers ────────────────────────────────────────────────────

export function recordValueForField(
  record: SettingRecord,
  config: SystemSettingConfig,
  field: SystemSettingField,
): unknown {
  if (
    (config.slug === "creditor" || config.slug === "debtor") &&
    field.key === "creditlimitbaht"
  ) {
    const satang = Number(getByPath(record, "creditlimitsatang") ?? 0);
    return Number.isFinite(satang) && satang !== 0 ? satang / 100 : undefined;
  }
  const value = getByPath(record, field.key);
  if (value !== undefined) return value;
  const aliases = fieldValueAliases[`${config.slug}.${field.key}`] ?? [];
  for (const alias of aliases) {
    const aliasValue = getByPath(record, alias);
    if (aliasValue !== undefined) return aliasValue;
  }
  return undefined;
}

export function productCategoryParentGuid(
  record: SettingRecord | null | undefined,
): string {
  if (!record) return "";
  return stringValue(record.parentguid ?? record.parentguid);
}

export function productCategoryGroupNumber(
  record: SettingRecord | null | undefined,
): number {
  if (!record) return 0;
  const value = Number(record.groupnumber ?? record.groupnumber ?? 0);
  return Number.isFinite(value) ? value : 0;
}

// ─── Field Grid Layout ───────────────────────────────────────────────────────

export function fieldGridItemClass(
  field: SystemSettingField,
  config: SystemSettingConfig,
): string {
  if (config.kind === "company") return "min-w-0 md:col-span-2";
  if (isBranchLatitudeField(config, field)) return "min-w-0 md:col-span-2";
  if (
    field.type === "bank-accounts" ||
    field.type === "branch-multi-select" ||
    field.type === "company-multi-select" ||
    field.type === "holding-scope-rules" ||
    field.type === "image-upload" ||
    field.type === "image-gallery" ||
    field.type === "json" ||
    field.type === "language-configs" ||
    field.type === "language-list" ||
    field.type === "master-picker" ||
    field.type === "master-multi-picker" ||
    field.type === "names" ||
    field.type === "radio" ||
    field.type === "string-list" ||
    field.type === "thai-address" ||
    field.type === "time-sale-list" ||
    field.type === "textarea"
  ) {
    return "min-w-0 md:col-span-2";
  }
  if (
    (config.slug === "permissiondefinition" && isPermissionAccessRulesField(field)) ||
    (config.slug === "approvalsetting" && field.key === "approvals") ||
    (config.slug === "permissionlink" &&
      (isEmployeeCodeField(field) ||
        isPermissionCodesField(field) ||
        isApprovalCodesField(field)))
  ) {
    return "min-w-0 md:col-span-2";
  }
  return "min-w-0";
}

export function isBranchLongitudeField(
  config: SystemSettingConfig,
  field: SystemSettingField,
): boolean {
  return config.slug === "branch" && field.key === "contact.longitude";
}

export function isColorHexField(
  config: SystemSettingConfig,
  field: SystemSettingField,
): boolean {
  return (
    (config.slug === "productcategorygroupselectscreen" &&
      field.key === "colorselecthex") ||
    (config.slug === "productcolor" && field.key === "hexcolor")
  );
}

// ─── Form Value Helpers ──────────────────────────────────────────────────────

export function radioFormValue(value: unknown, field: SystemSettingField): string {
  const fallback = field.options?.[0]?.value ?? "";
  if (field.valueType !== "boolean") return stringValue(value) || fallback;
  if (typeof value === "boolean") return value ? "true" : "false";
  const normalized = stringValue(value).toLowerCase();
  if (normalized === "true" || normalized === "false") return normalized;
  if (
    normalized === "1" ||
    normalized === "yes" ||
    normalized === "y" ||
    normalized === "buddhist" ||
    normalized === "be" ||
    normalized === "พ.ศ."
  )
    return "true";
  if (
    normalized === "0" ||
    normalized === "no" ||
    normalized === "n" ||
    normalized === "christian" ||
    normalized === "ce" ||
    normalized === "ค.ศ."
  )
    return "false";
  return fallback;
}

export function radioValueToFormValue(
  value: string,
  field: SystemSettingField,
): string | boolean | number {
  if (field.valueType === "boolean") return booleanLikeValue(value);
  if (field.valueType === "number") return Number(value);
  return value;
}

export function optionValueToFormValue(
  value: string,
  field: SystemSettingField,
): string | number {
  if (field.valueType === "number") {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : 0;
  }
  return value;
}

export function normalizeHexColor(value: unknown): string {
  const raw = stringValue(value).replace(/^#/, "");
  if (/^[0-9a-fA-F]{6}$/.test(raw)) return `#${raw}`;
  if (/^[0-9a-fA-F]{8}$/.test(raw)) return `#${raw.slice(2)}`;
  return "#ffffff";
}

export function languageName(code: string, language: LanguageCode): string {
  const item = LANGUAGES.find((entry) => entry.code === code);
  if (!item) return code.toUpperCase();
  if (language === "th") return item.name;
  return `${item.name} (${item.code.toUpperCase()})`;
}
