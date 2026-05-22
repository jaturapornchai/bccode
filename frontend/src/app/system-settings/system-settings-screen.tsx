"use client";

import {
  AlertCircle,
  BadgeCheck,
  Bot,
  Building2,
  CalendarCheck2,
  CalendarDays,
  Check,
  Copy,
  Crown,
  DownloadCloud,
  Edit3,
  FileCog,
  GitBranch,
  ImageIcon,
  KeyRound,
  Link2,
  Loader2,
  Network,
  Plus,
  RefreshCcw,
  Save,
  Search,
  ShieldCheck,
  Tag,
  Trash2,
  UserCheck,
  UserRound,
  UserX,
  UsersRound,
  ChevronsUpDown,
  UploadCloud,
  X,
} from "lucide-react";
import Image from "next/image";
import { useRouter } from "next/navigation";
import { type CSSProperties, FormEvent, ReactNode, type PointerEvent as ReactPointerEvent, useCallback, useEffect, useLayoutEffect, useMemo, useRef, useState } from "react";
import { MasterPicker } from "@/components/product-barcode/master-picker";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { DateField, TimeField } from "@/components/ui/date-time-field";
import { Input } from "@/components/ui/input";
import { backendText, useBackendLanguage, type BackendLanguageDictionary } from "@/lib/backend-language";
import { localTimeToUtcTime, normalizeTimeInput, timeToMinutes, type CalendarYearType } from "@/lib/date-time";
import {
  getSystemSettingConfig,
  systemSettingLabel,
  type SystemSettingConfig,
  type SystemSettingField,
  type SystemSettingOption,
} from "@/lib/system-setting-screens";
import { LANGUAGES, normalizeLanguage, type LanguageCode } from "@/lib/i18n";
import { MENU_SECTIONS, menuText } from "@/lib/menu-data";
import type { MasterEntry, MasterName } from "@/lib/product-barcode/api";
import {
  branchDisplayName,
  shopDisplayName,
  type AuthSession,
  type WorkspaceSession,
  workspaceStorageKeys,
} from "@/lib/workspace-models";
import { cn } from "@/lib/utils";
import { LanguageDialog } from "../language-dialog";
import { ManualLink } from "../manual-link";
import { ThemeToggle } from "../theme-toggle";

type SystemSettingsScreenProps = {
  embedded?: boolean;
  initialBackendLanguage?: BackendLanguageDictionary;
  initialBackendUrl?: string;
  initialLanguage?: LanguageCode;
  language?: LanguageCode;
  route: string;
};

type SettingRecord = Record<string, unknown>;
type FormState = Record<string, unknown>;
type Notice = { type: "error" | "info" | "success"; text: string } | null;
type ProductUnitOption = {
  unitcode: string;
  names?: { code?: string; name?: string }[];
};
type LanguageConfigFormRow = {
  code: LanguageCode;
  codetranslator: string;
  name: string;
  is_use: boolean;
  isdefault: boolean;
};
type StandardUnitDialogState = {
  open: boolean;
  loading: boolean;
  saving: boolean;
  query: string;
  options: ProductUnitOption[];
  selectedCodes: string[];
  source: string;
  error: string;
};

const SETTINGS_LIST_PAGE_SIZE = 100;

type WorkDayTime = {
  starttime?: string;
  endtime?: string;
  start?: string;
  end?: string;
  starttimeutc?: string;
  endtimeutc?: string;
  startdayoffsetutc?: number;
  enddayoffsetutc?: number;
  timezone?: string;
  timezone_offset?: string;
};

type WorkDay = {
  code: string;
  name: string;
  isactive: boolean;
  fullday: boolean;
  worktimes: WorkDayTime[];
};

type DateTimeScope = {
  key: string;
  branchcode: string;
  branchguid: string;
  timezone: string;
  timezone_label: string;
  timezone_offset: string;
  calendarYearType: CalendarYearType;
};

const emptyStandardUnitDialog: StandardUnitDialogState = {
  open: false,
  loading: false,
  saving: false,
  query: "",
  options: [],
  selectedCodes: [],
  source: "",
  error: "",
};

const companySetupDefaults: FormState = {
  "settings.country_code": "TH",
  "settings.language": "th",
  "settings.languageconfigs": defaultLanguageConfigs("th"),
  "settings.base_currency": "THB",
  "settings.timezone": "Asia/Bangkok",
  "settings.date_format": "dd/MM/yyyy",
  "settings.decimal_quantity": "2",
  "settings.decimal_price": "2",
  "settings.decimal_document": "2",
  "settings.is_vat_registered": false,
  "settings.vatrate": "0",
  "settings.usebuddhistcalendar": true,
};

const branchSetupDefaults: FormState = {
  languages: ["th"],
  language: "th",
  "contact.latitude": 0,
  "contact.longitude": 0,
  "contact.phonenumber": "",
  "pos.taxid": "",
  "pos.vatrate": 0,
  "pos.vattypesale": 0,
  "pos.vattypepurchase": 0,
  "pos.inquirytypesale": 0,
  "pos.inquirytypepurchase": 0,
  "pos.headerreceiptpos": "",
  "pos.footerreceiptpos": "",
  "pos.isbom": false,
  decimal_quantity: 2,
  decimal_price: 2,
  decimal_document: 2,
  couponusetype: 0,
  paymentrounding: defaultPaymentRoundingJson(),
};

const uiEn = {
  active: "Active",
  add: "Add",
  all: "All",
  accessEnabled: "Can access",
  accessStatus: "Access status",
  accessTemporarilyDisabled: "Temporarily disabled",
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
  resetPassword: "Reset password",
  resetPasswordConfirm: "Reset this user's password to 12345?",
  resetPasswordDone: "Password was reset to 12345.",
  save: "Save",
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

const uiText: Partial<Record<LanguageCode, Partial<Record<keyof typeof uiEn, string>>>> = {
  th: {
    active: "ใช้งาน",
    add: "เพิ่ม",
    all: "ทั้งหมด",
    accessEnabled: "เข้าใช้งานได้",
    accessStatus: "สถานะเข้าใช้งาน",
    accessTemporarilyDisabled: "เข้าใช้งานไม่ได้ชั่วคราว",
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
    deleteConfirm: "ยืนยันลบรายการนี้?",
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
    resetPassword: "รีเซ็ตรหัสผ่าน",
    resetPasswordConfirm: "ยืนยันรีเซ็ตรหัสผ่านผู้ใช้นี้กลับเป็น 12345?",
    resetPasswordDone: "รีเซ็ตรหัสผ่านเป็น 12345 แล้ว",
    save: "บันทึก",
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

const uiBackendKeys: Partial<Record<keyof typeof uiEn, string>> = {
  accessEnabled: "access_enabled",
  accessStatus: "access_status",
  accessTemporarilyDisabled: "access_temporarily_disabled",
  active: "active",
  add: "add",
  all: "all",
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

const systemSettingBackendKeys: Record<string, string> = {
  branch: "branch",
  business_type_screen: "business_type",
  company: "company",
  department: "department",
  employee: "employee",
  holiday_screen: "holiday",
  permission_definition: "permission_definition",
  permission_link: "permission_link",
  approval_setting: "approval_setting",
  user: "user",
  work_day_screen: "work_day",
};

const fieldBackendKeys: Record<string, string> = {
  "branch.base_currency": "base_currency",
  "branch.code": "branch_code",
  "branch.date_format": "date_format",
  "branch.decimal_document": "decimal_document",
  "branch.decimal_price": "decimal_price",
  "branch.decimal_quantity": "decimal_quantity",
  "branch.language": "default_language",
  "branch.names": "branch_name",
  "branch.timezone": "timezone",
  "branch.yeartype": "year_type",
  "branch.languages": "select_data_language",
  "branch.businesstype": "business_type",
  "branch.pos.taxid": "company_tax_id",
  "branch.pos.vatrate": "vat_rate",
  "branch.pos.isbom": "cut_stock_by_bom",
  "branch.pos.vattypepurchase": "vattype_purchase",
  "branch.pos.inquirytypepurchase": "inquirytype_purchase",
  "branch.pos.vattypesale": "vattype_sale",
  "branch.pos.inquirytypesale": "inquirytype_sale",
  "business_type_screen.code": "code",
  "business_type_screen.names": "business_type",
  "company.address": "company_address",
  "company.logo": "company_logo",
  "company.names": "company_name",
  "company.settings.base_currency": "base_currency",
  "company.settings.company_registration_no": "company_registration_no",
  "company.settings.country_code": "country_code",
  "company.settings.date_format": "date_format",
  "company.settings.decimal_document": "decimal_document",
  "company.settings.decimal_price": "decimal_price",
  "company.settings.decimal_quantity": "decimal_quantity",
  "company.settings.is_vat_registered": "vat_status",
  "company.settings.isusebranch": "use_branch_system",
  "company.settings.isusedepartment": "use_department_system",
  "company.settings.language": "default_language",
  "company.settings.languageconfigs": "active_languages",
  "company.settings.tax_id": "tax_id",
  "company.settings.timezone": "timezone",
  "company.settings.usebuddhistcalendar": "year_type",
  "company.settings.vatrate": "vat_rate",
  "company.telephone": "telephone",
  "department.code": "department_code",
  "department.names": "department_name",
  "holiday_screen.date": "date",
  "holiday_screen.desc": "description",
  "user.is_access_disabled": "access_status",
  "user.uid": "user_id_guid",
  "user.user_profile_name": "user_name",
  "user.email": "registered_email",
  "user.role": "user_role",
  "user.position": "user_position",
  "user.department": "department",
  "user.line_user_id": "line_user_id",
  "user.line_display_name": "line_display_name",
  "permission_definition.permissionCode": "permission_code",
  "permission_definition.permissionName": "permission_name",
  "permission_definition.branches": "branch_permissions",
  "approval_setting.approvalCode": "approval_code",
  "approval_setting.approvalName": "approval_name",
  "approval_setting.approvals": "approval_permission",
  "permission_link.employeeCode": "user_employee_code",
  "permission_link.employeeName": "name",
  "permission_link.permissionCodes": "permission_codes",
  "permission_link.approvalCodes": "approval_codes",
};

const dayNames: Record<LanguageCode, string[]> = {
  th: ["จันทร์", "อังคาร", "พุธ", "พฤหัสบดี", "ศุกร์", "เสาร์", "อาทิตย์"],
  en: ["Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday", "Sunday"],
  cn: ["星期一", "星期二", "星期三", "星期四", "星期五", "星期六", "星期日"],
  ja: ["月曜", "火曜", "水曜", "木曜", "金曜", "土曜", "日曜"],
  ko: ["월요일", "화요일", "수요일", "목요일", "금요일", "토요일", "일요일"],
  lo: ["ຈັນ", "ອັງຄານ", "ພຸດ", "ພະຫັດ", "ສຸກ", "ເສົາ", "ອາທິດ"],
  my: ["တနင်္လာ", "အင်္ဂါ", "ဗုဒ္ဓဟူး", "ကြာသပတေး", "သောကြာ", "စနေ", "တနင်္ဂနွေ"],
  km: ["ចន្ទ", "អង្គារ", "ពុធ", "ព្រហស្បតិ៍", "សុក្រ", "សៅរ៍", "អាទិត្យ"],
  vi: ["Thứ hai", "Thứ ba", "Thứ tư", "Thứ năm", "Thứ sáu", "Thứ bảy", "Chủ nhật"],
  ms: ["Isnin", "Selasa", "Rabu", "Khamis", "Jumaat", "Sabtu", "Ahad"],
  id: ["Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu", "Minggu"],
  fil: ["Lunes", "Martes", "Miyerkules", "Huwebes", "Biyernes", "Sabado", "Linggo"],
};

export function SystemSettingsScreen({ embedded = false, initialBackendLanguage, initialBackendUrl, initialLanguage = "th", language: externalLanguage, route }: SystemSettingsScreenProps) {
  const router = useRouter();
  const config = getSystemSettingConfig(route);
  const [language, setLanguage] = useState<LanguageCode>(externalLanguage ?? initialLanguage);
  const [auth, setAuth] = useState<AuthSession | null>(null);
  const [workspace, setWorkspace] = useState<WorkspaceSession | null>(null);
  const [records, setRecords] = useState<SettingRecord[]>([]);
  const [query, setQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  const [recordTotal, setRecordTotal] = useState(0);
  const [allRecordTotal, setAllRecordTotal] = useState(0);
  const [allRecordsLoaded, setAllRecordsLoaded] = useState(false);
  const [saving, setSaving] = useState(false);
  const [notice, setNotice] = useState<Notice>(null);
  const [editing, setEditing] = useState<SettingRecord | null>(null);
  const [form, setForm] = useState<FormState>({});
  const [formOpen, setFormOpen] = useState(false);
  const [selectedRecordId, setSelectedRecordId] = useState("");
  const [autoOpenedCompanyId, setAutoOpenedCompanyId] = useState("");
  const [workDays, setWorkDays] = useState<WorkDay[]>([]);
  const [sourceShopId, setSourceShopId] = useState("");
  const [copySourceEnvironment, setCopySourceEnvironment] = useState<"uat" | "pro">("uat");
  const [copyPreview, setCopyPreview] = useState<unknown>(null);
  const [standardUnitDialog, setStandardUnitDialog] = useState<StandardUnitDialogState>(emptyStandardUnitDialog);
  const { confirm, confirmationDialog } = useConfirmDialog();
  const activeBackendUrl = auth?.backendUrl ?? initialBackendUrl;
  const backendLanguage = useBackendLanguage(language, activeBackendUrl, language === initialLanguage ? initialBackendLanguage : undefined);

  const text = useCallback((key: keyof typeof uiEn) => {
    const fallback = uiText[language]?.[key] ?? uiEn[key];
    const backendKey = uiBackendKey(key);
    const value = backendText(backendLanguage, backendKey, fallback);
    return value === backendKey || value === key ? fallback : value;
  }, [backendLanguage, language]);
  const errorText = useCallback((error: unknown, fallback = text("requestFailed")) => {
    const message = error instanceof Error ? error.message : "";
    return message ? backendText(backendLanguage, message, message) : fallback;
  }, [backendLanguage, text]);
  const title = config ? systemSettingTitle(config, language, backendLanguage) : route;
  const subtitle = config ? config.subtitle[language] ?? config.subtitle.en ?? config.subtitle.th : "";
  const dateTimeScope = useMemo(() => resolveDateTimeScope(workspace), [workspace]);

  const loadRecords = useCallback(async (
    currentAuth: AuthSession | null,
    currentWorkspace: WorkspaceSession | null,
    currentConfig: SystemSettingConfig | undefined,
    offset = 0,
    append = false,
  ) => {
    if (!currentAuth || !currentWorkspace || !currentConfig) return;
    if (append) setLoadingMore(true);
    else setLoading(true);
    setNotice(null);
    try {
      const searchParams = new URLSearchParams({
        limit: String(SETTINGS_LIST_PAGE_SIZE),
        offset: String(offset),
        page: String(Math.floor(offset / SETTINGS_LIST_PAGE_SIZE) + 1),
        q: query,
        shopid: currentWorkspace.shop.shopid,
      });
      const scope = resolveDateTimeScope(currentWorkspace);
      if (currentConfig.slug === "department" || currentConfig.kind === "restaurant-setting") {
        if (scope.key) searchParams.set("branch_key", scope.key);
        if (scope.branchcode) searchParams.set("branchcode", scope.branchcode);
        if (scope.branchguid) searchParams.set("branchguid", scope.branchguid);
      }
      if (currentConfig.kind === "copy-uat") searchParams.set("source_environment", copySourceEnvironment);
      const response = await fetch(`/api/system-settings/${currentConfig.slug}?${searchParams.toString()}`, {
        headers: {
          "x-bc-backend-url": currentAuth.backendUrl,
          Authorization: `Bearer ${currentAuth.token}`,
        },
        cache: "no-store",
      });
      const payload = await response.json() as unknown;
      if (!response.ok || isFailed(payload)) throw new Error(extractMessage(payload) ?? text("requestFailed"));
      const nextRecords = normalizeRecords(payload, currentConfig);
      const scopedRecords = currentConfig.slug === "holiday_screen" || currentConfig.slug === "department" ? filterRecordsByDateTimeScope(nextRecords, scope) : nextRecords;
      const nextTotal = extractRecordTotal(payload, scopedRecords.length);
      setRecordTotal(nextTotal);
      if (!query.trim()) setAllRecordTotal(nextTotal);
      setAllRecordsLoaded(scopedRecords.length < SETTINGS_LIST_PAGE_SIZE || offset + scopedRecords.length >= nextTotal);
      setRecords((currentRecords) => append ? mergeRecords(currentRecords, scopedRecords, currentConfig) : scopedRecords);
      if (currentConfig.slug === "work_day_screen") setWorkDays(normalizeWorkDays(nextRecords, language, scope));
    } catch (error) {
      setNotice({ type: "error", text: errorText(error) });
    } finally {
      if (append) setLoadingMore(false);
      else setLoading(false);
    }
  }, [copySourceEnvironment, errorText, language, query, setLoading, setNotice, setRecords, setWorkDays, text]);

  useEffect(() => {
    if (externalLanguage) setLanguage(externalLanguage);
  }, [externalLanguage]);

  useEffect(() => {
    if (config?.kind !== "copy-uat" || !auth || !workspace) return;
    setSourceShopId("");
    setCopyPreview(null);
    void loadRecords(auth, workspace, config);
  }, [auth, config, copySourceEnvironment, loadRecords, workspace]);

  useEffect(() => {
    if (!config) return;
    const savedLanguage = normalizeLanguage(localStorage.getItem("user_language") ?? externalLanguage ?? initialLanguage);
    if (!externalLanguage) setLanguage(savedLanguage);

    const nextAuth = readAuth();
    const nextWorkspace = readWorkspace();
    if (!nextAuth || !nextWorkspace) {
      setNotice({ type: "info", text: "กรุณาเข้าสู่ระบบและเลือกบริษัทก่อนเปิดหน้าจอนี้" });
      if (!embedded) router.replace("/");
      return;
    }

    setAuth(nextAuth);
    setWorkspace(nextWorkspace);
    void loadRecords(nextAuth, nextWorkspace, config);
  }, [config, embedded, externalLanguage, initialLanguage, loadRecords, router]);

  useEffect(() => {
    document.documentElement.lang = language;
    if (!externalLanguage) localStorage.setItem("user_language", language);
  }, [externalLanguage, language]);

  const visibleRecords = useMemo(() => {
    if (!config) return records;
    const needle = query.trim().toLowerCase();
    if (!needle) return records;
    return records.filter((record) => `${recordTitle(record, config, language)} ${recordId(record, config)} ${JSON.stringify(record)}`.toLowerCase().includes(needle));
  }, [config, language, query, records]);
  const displayTotal = recordTotal || visibleRecords.length;
  const totalAllRecords = allRecordTotal || (!query.trim() ? displayTotal : 0);
  const hasMoreRecords = Boolean(config && records.length > 0 && records.length < displayTotal && !allRecordsLoaded);
  const loadMoreRecords = useCallback(() => {
    if (!auth || !workspace || !config || loading || loadingMore || !hasMoreRecords) return;
    void loadRecords(auth, workspace, config, records.length, true);
  }, [auth, config, hasMoreRecords, loadRecords, loading, loadingMore, records.length, workspace]);
  const selectedRecord = useMemo(() => {
    if (!config || !visibleRecords.length) return null;
    return visibleRecords.find((record) => recordId(record, config) === selectedRecordId) ?? visibleRecords[0] ?? null;
  }, [config, selectedRecordId, visibleRecords]);

  useEffect(() => {
    if (!config || !visibleRecords.length) {
      setSelectedRecordId("");
      return;
    }
    if (selectedRecordId && visibleRecords.some((record) => recordId(record, config) === selectedRecordId)) return;
    setSelectedRecordId(recordId(visibleRecords[0], config));
  }, [config, selectedRecordId, visibleRecords]);

  useEffect(() => {
    if (!config || config.kind !== "company" || formOpen || editing) return;
    const record = companyRecordForEdit(records, workspace);
    if (!record) return;
    const id = recordId(record, config);
    if (autoOpenedCompanyId === id) return;
    setEditing(record);
    setForm(formFromRecord(record, config, language));
    setFormOpen(true);
    setNotice(null);
    setAutoOpenedCompanyId(id);
  }, [autoOpenedCompanyId, config, editing, formOpen, language, records, workspace]);

  if (!config) {
    return (
      <main className="min-h-dvh w-full bg-background p-3 text-foreground">
        <Card><CardContent className="p-4 text-sm text-destructive">ไม่พบหน้าจอ {route}</CardContent></Card>
      </main>
    );
  }
  const currentConfig = config;
  const canEdit = currentConfig.editable !== false;

  function openCreate() {
    setEditing(null);
    setForm(defaultForm(currentConfig, language));
    setFormOpen(true);
    setSelectedRecordId("");
    setNotice(null);
  }

  async function openEdit(record: SettingRecord) {
    setSelectedRecordId(recordId(record, currentConfig));
    setEditing(record);
    setForm(formFromRecord(record, currentConfig, language));
    setFormOpen(true);
    setNotice(null);
    if (!auth || !workspace || currentConfig.kind !== "main-crud") return;
    const id = recordId(record, currentConfig);
    if (!id) return;
    try {
      const params = new URLSearchParams({ shopid: workspace.shop.shopid });
      const response = await fetch(`/api/system-settings/${currentConfig.slug}/${encodeURIComponent(id)}?${params.toString()}`, {
        headers: requestHeaders(auth),
        cache: "no-store",
      });
      const payload = await response.json() as unknown;
      if (!response.ok || isFailed(payload)) return;
      const detail = normalizeRecords(payload, currentConfig)[0];
      if (!detail) return;
      setEditing(detail);
      setForm(formFromRecord(detail, currentConfig, language));
    } catch {
      // Keep the list-row data visible if detail fetch is unavailable.
    }
  }

  async function saveRecord(event?: FormEvent<HTMLFormElement>) {
    event?.preventDefault();
    if (!auth || !workspace) return;

    let payload: SettingRecord;
    try {
      payload = buildPayload(form, editing, currentConfig, workspace, auth, language);
    } catch (error) {
      setNotice({ type: "error", text: error instanceof Error ? error.message : text("jsonInvalid") });
      return;
    }

    const missingRequired = currentConfig.fields.some((field) => field.required && !String(getByPath(payload, field.key) ?? "").trim());
    if (missingRequired) {
      setNotice({ type: "error", text: text("errorRequired") });
      return;
    }

    if (!editing && currentConfig.slug === "user") {
      const username = stringValue(payload.username).toLowerCase();
      const duplicateUser = records.some((record) => stringValue(record.username ?? record.email ?? record.code).toLowerCase() === username);
      if (duplicateUser) {
        setNotice({ type: "error", text: language === "th" ? "รหัสผู้ใช้งานซ้ำ" : "User code already exists." });
        return;
      }
    }

    const id = recordId(editing ?? payload, currentConfig);
    setSaving(true);
    setNotice(null);
    try {
      const response = await fetch(`/api/system-settings/${currentConfig.slug}${editing && id ? `/${encodeURIComponent(id)}` : ""}`, {
        method: editing ? "PUT" : "POST",
        headers: requestHeaders(auth),
        body: JSON.stringify({ ...payload, backendUrl: auth.backendUrl, shopid: workspace.shop.shopid }),
      });
      const data = await response.json() as unknown;
      if (!response.ok || isFailed(data)) throw new Error(extractMessage(data) ?? text("requestFailed"));
      setNotice({ type: "success", text: text("saved") });
      if (currentConfig.kind === "company") {
        const nextWorkspace = { ...workspace, shopInfo: payload };
        setWorkspace(nextWorkspace);
        localStorage.setItem(workspaceStorageKeys.workspace, JSON.stringify(nextWorkspace));
        localStorage.setItem(workspaceStorageKeys.shopInfo, JSON.stringify(payload));
        setFormOpen(true);
        setEditing(payload);
      } else {
        setFormOpen(false);
        setEditing(null);
      }
      await loadRecords(auth, workspace, currentConfig);
    } catch (error) {
      setNotice({ type: "error", text: errorText(error) });
    } finally {
      setSaving(false);
    }
  }

  async function deleteRecord(record: SettingRecord) {
    if (!auth || !workspace) return;
    if (currentConfig.slug === "user" && isSelfUserRecord(record, auth)) {
      setNotice({ type: "info", text: text("selfPermissionCannotEdit") });
      return;
    }
    if (currentConfig.slug === "user" && isCreatorRecord(record, workspace)) {
      setNotice({ type: "info", text: text("creatorCannotDelete") });
      return;
    }
    const id = recordId(record, currentConfig);
    const confirmed = await confirm({
      title: text("deleteConfirm"),
      description: recordTitle(record, currentConfig, language),
      details: id ? `${language === "th" ? "รหัสอ้างอิง" : "Reference ID"}: ${id}` : undefined,
      confirmLabel: text("delete"),
      cancelLabel: text("cancel"),
      tone: "danger",
    });
    if (!confirmed) return;
    setLoading(true);
    setNotice(null);
    try {
      const response = await fetch(`/api/system-settings/${currentConfig.slug}/${encodeURIComponent(id)}?shopid=${encodeURIComponent(workspace.shop.shopid)}`, {
        method: "DELETE",
        headers: requestHeaders(auth),
        body: JSON.stringify({ backendUrl: auth.backendUrl, shopid: workspace.shop.shopid }),
      });
      const data = await response.json() as unknown;
      if (!response.ok || isFailed(data)) throw new Error(extractMessage(data) ?? text("requestFailed"));
      setNotice({ type: "success", text: text("saved") });
      await loadRecords(auth, workspace, currentConfig);
    } catch (error) {
      setNotice({ type: "error", text: errorText(error) });
    } finally {
      setLoading(false);
    }
  }

  async function toggleUserAccess(record: SettingRecord, disabled: boolean) {
    if (!auth || !workspace || currentConfig.slug !== "user") return;
    if (isSelfUserRecord(record, auth)) {
      setNotice({ type: "info", text: text("selfPermissionCannotEdit") });
      return;
    }
    if (isCreatorRecord(record, workspace)) {
      setNotice({ type: "info", text: text("creatorAlwaysEnabled") });
      return;
    }

    const id = recordId(record, currentConfig);
    const nextForm = formFromRecord(record, currentConfig, language);
    nextForm.is_access_disabled = disabled;

    let payload: SettingRecord;
    try {
      payload = buildPayload(nextForm, record, currentConfig, workspace, auth, language);
    } catch (error) {
      setNotice({ type: "error", text: error instanceof Error ? error.message : text("jsonInvalid") });
      return;
    }

    setSaving(true);
    setNotice(null);
    try {
      const response = await fetch(`/api/system-settings/${currentConfig.slug}/${encodeURIComponent(id)}`, {
        method: "PUT",
        headers: requestHeaders(auth),
        body: JSON.stringify({ ...payload, backendUrl: auth.backendUrl, shopid: workspace.shop.shopid }),
      });
      const data = await response.json() as unknown;
      if (!response.ok || isFailed(data)) throw new Error(extractMessage(data) ?? text("requestFailed"));
      setNotice({ type: "success", text: text("saved") });
      await loadRecords(auth, workspace, currentConfig);
    } catch (error) {
      setNotice({ type: "error", text: errorText(error) });
    } finally {
      setSaving(false);
    }
  }

  async function resetUserPassword(record: SettingRecord) {
    if (!auth || !workspace || currentConfig.slug !== "user") return;
    if (isSelfUserRecord(record, auth)) {
      setNotice({ type: "info", text: text("selfPermissionCannotEdit") });
      return;
    }
    const username = stringValue(record.username ?? record.email ?? record.code);
    if (!username) return;
    const confirmed = await confirm({
      title: text("resetPassword"),
      description: text("resetPasswordConfirm"),
      details: `${language === "th" ? "ผู้ใช้" : "User"}: ${username}`,
      confirmLabel: text("resetPassword"),
      cancelLabel: text("cancel"),
      tone: "warning",
    });
    if (!confirmed) return;

    setSaving(true);
    setNotice(null);
    try {
      const response = await fetch("/api/auth/profile/reset-password", {
        method: "PUT",
        headers: requestHeaders(auth),
        body: JSON.stringify({
          backendUrl: auth.backendUrl,
          username,
          shopid: workspace.shop.shopid,
        }),
      });
      const data = await response.json() as unknown;
      if (!response.ok || isFailed(data)) throw new Error(extractMessage(data) ?? text("requestFailed"));
      setNotice({ type: "success", text: text("resetPasswordDone") });
    } catch (error) {
      setNotice({ type: "error", text: errorText(error) });
    } finally {
      setSaving(false);
    }
  }

  async function saveWorkDays() {
    if (!auth || !workspace) return;
    const currentId = recordId(records[0], currentConfig);
    setSaving(true);
    setNotice(null);
    try {
      const response = await fetch(`/api/system-settings/${currentConfig.slug}${currentId ? `/${encodeURIComponent(currentId)}` : ""}`, {
        method: currentId ? "PUT" : "POST",
        headers: requestHeaders(auth),
        body: JSON.stringify({
          backendUrl: auth.backendUrl,
          body: JSON.stringify(buildBranchScopedWorkDayBody(records[0], workspace, workDays)),
          shopid: workspace.shop.shopid,
        }),
      });
      const data = await response.json() as unknown;
      if (!response.ok || isFailed(data)) throw new Error(extractMessage(data) ?? text("requestFailed"));
      setNotice({ type: "success", text: text("saved") });
      await loadRecords(auth, workspace, currentConfig);
    } catch (error) {
      setNotice({ type: "error", text: errorText(error) });
    } finally {
      setSaving(false);
    }
  }

  async function runCopy(action: "copy" | "preview") {
    if (!auth || !workspace || !sourceShopId) return;
    setSaving(true);
    setNotice(null);
    try {
      const response = await fetch(`/api/system-settings/${currentConfig.slug}?shopid=${encodeURIComponent(workspace.shop.shopid)}`, {
        method: "POST",
        headers: requestHeaders(auth),
        body: JSON.stringify({
          action,
          backendUrl: auth.backendUrl,
          source_environment: copySourceEnvironment,
          target_environment: "dev",
          source_shop_id: sourceShopId,
          target_shop_id: workspace.shop.shopid,
        }),
      });
      const data = await response.json() as unknown;
      if (!response.ok || isFailed(data)) throw new Error(extractMessage(data) ?? text("requestFailed"));
      setCopyPreview(data);
      setNotice({ type: "success", text: action === "copy" ? text("saved") : text("preview") });
    } catch (error) {
      setNotice({ type: "error", text: errorText(error) });
    } finally {
      setSaving(false);
    }
  }

  async function openStandardUnitDialog() {
    if (!auth || !workspace) return;
    setStandardUnitDialog({ ...emptyStandardUnitDialog, open: true, loading: true });
    await loadStandardUnitOptions("");
  }

  async function loadStandardUnitOptions(searchText = standardUnitDialog.query) {
    if (!auth || !workspace) return;
    setStandardUnitDialog((current) => ({ ...current, loading: true, error: "", query: searchText }));
    try {
      const params = new URLSearchParams({
        backendUrl: auth.backendUrl,
        mainShopId: getMainShopIdFromWorkspace(workspace),
        q: searchText,
      });
      const response = await fetch(`/api/workspace/product-units/standard?${params.toString()}`, {
        headers: requestHeaders(auth),
        cache: "no-store",
      });
      const payload = await response.json() as unknown;
      if (!response.ok || isFailed(payload)) throw new Error(extractMessage(payload) ?? text("requestFailed"));
      const options = isRecord(payload) && Array.isArray(payload.data) ? payload.data.filter(isProductUnitOption) : [];
      setStandardUnitDialog((current) => ({
        ...current,
        loading: false,
        options,
        selectedCodes: options.map((unit) => unit.unitcode),
        source: isRecord(payload) ? stringValue(payload.source) : "",
      }));
    } catch (error) {
      setStandardUnitDialog((current) => ({ ...current, loading: false, error: errorText(error) }));
    }
  }

  async function saveStandardUnits() {
    if (!auth || !workspace || standardUnitDialog.selectedCodes.length === 0) return;
    setStandardUnitDialog((current) => ({ ...current, saving: true, error: "" }));
    setNotice(null);
    try {
      const response = await fetch("/api/workspace/product-units/defaults", {
        method: "POST",
        headers: requestHeaders(auth),
        body: JSON.stringify({
          backendUrl: auth.backendUrl,
          mainShopId: getMainShopIdFromWorkspace(workspace),
          unitcodes: standardUnitDialog.selectedCodes,
        }),
      });
      const payload = await response.json() as unknown;
      if (!response.ok || isFailed(payload)) throw new Error(extractMessage(payload) ?? text("requestFailed"));
      setStandardUnitDialog(emptyStandardUnitDialog);
      setNotice({ type: "success", text: extractMessage(payload) ?? text("saved") });
      await loadRecords(auth, workspace, currentConfig);
    } catch (error) {
      setStandardUnitDialog((current) => ({ ...current, saving: false, error: errorText(error) }));
    }
  }

  function toggleStandardUnit(unitcode: string, checked: boolean) {
    setStandardUnitDialog((current) => {
      const selected = new Set(current.selectedCodes);
      if (checked) selected.add(unitcode);
      else selected.delete(unitcode);
      return { ...current, selectedCodes: Array.from(selected) };
    });
  }

  function selectAllStandardUnits() {
    setStandardUnitDialog((current) => ({ ...current, selectedCodes: current.options.map((unit) => unit.unitcode) }));
  }

  function clearStandardUnits() {
    setStandardUnitDialog((current) => ({ ...current, selectedCodes: [] }));
  }

  const content = (
    <div className="grid w-full min-w-0 gap-3">
      <header className="rounded-2xl border border-border bg-card p-3 shadow-sm">
        <div className="flex w-full flex-wrap items-start justify-between gap-2">
          <div className="flex min-w-0 items-start gap-2">
            <span className="grid size-10 shrink-0 place-items-center rounded-2xl bg-primary/10 text-primary">
              {settingIcon(config.icon)}
            </span>
            <div className="min-w-0">
              <p className="text-xs font-semibold uppercase text-muted-foreground">BC Ai Account</p>
              <h1 className="truncate text-xl font-semibold sm:text-2xl">{title}</h1>
              <p className="max-w-[92ch] text-sm leading-6 text-muted-foreground">{subtitle}</p>
            </div>
          </div>
          <div className="flex min-w-0 flex-wrap justify-end gap-2">
            {embedded ? null : <LanguageDialog language={language} onLanguageChange={setLanguage} />}
            <ManualLink compact language={language} screen={config.manual} />
            {embedded ? null : <ThemeToggle language={language} />}
          </div>
        </div>
        {workspace ? (
          <div className="mt-2 flex flex-wrap gap-2 text-xs text-muted-foreground">
            <Badge variant="outline">{text("company")}: {shopDisplayName(workspace.shop)}</Badge>
            <Badge variant="outline">{text("branch")}: {workspace.branch ? branchDisplayName(workspace.branch) : "-"}</Badge>
            <Badge variant="outline">{text("timezone")}: {dateTimeScope.timezone_label || dateTimeScope.timezone || dateTimeScope.timezone_offset || "-"}</Badge>
          </div>
        ) : null}
      </header>

      {notice ? (
        <div className={`message ${notice.type === "success" ? "success" : notice.type === "error" ? "error" : "info"}`}>
          {notice.type === "success" ? <BadgeCheck size={18} /> : <AlertCircle size={18} />}
          <span>{notice.text}</span>
        </div>
      ) : null}

      {config.kind === "copy-uat" ? (
        <CopyUatPanel
          config={config}
          copyPreview={copyPreview}
          language={language}
          loading={loading}
          records={records}
          runCopy={runCopy}
          saving={saving}
          selected={sourceShopId}
          setSelected={setSourceShopId}
          sourceEnvironment={copySourceEnvironment}
          setSourceEnvironment={setCopySourceEnvironment}
          targetShopId={workspace?.shop.shopid ?? ""}
          text={text}
        />
      ) : config.slug === "work_day_screen" ? (
        <WorkDayPanel
          language={language}
          loading={loading}
          dateTimeScope={dateTimeScope}
          onRefresh={() => void loadRecords(auth, workspace, config)}
          onSave={saveWorkDays}
          saving={saving}
          setWorkDays={setWorkDays}
          text={text}
          workDays={workDays}
        />
      ) : config.kind === "company" ? (
        loading && !companyRecordForEdit(records, workspace) ? (
          <Card><CardContent className="flex min-h-40 items-center justify-center gap-2 p-4 text-sm text-muted-foreground"><Loader2 className="animate-spin" />{text("loading")}</CardContent></Card>
        ) : formOpen ? (
          <SettingFormDialog
            inline
            auth={auth}
            config={config}
            dictionary={backendLanguage}
            editing={editing}
            form={form}
            dateTimeScope={dateTimeScope}
            language={language}
            saving={saving}
            setForm={setForm}
            workspace={workspace}
            onClose={() => undefined}
            onSubmit={saveRecord}
            text={text}
          />
        ) : (
          <Card>
            <CardContent className="grid min-h-44 place-items-center p-4 text-center">
              <div className="grid gap-2">
                <Loader2 className="mx-auto size-8 animate-spin text-muted-foreground" />
                <h2 className="text-base font-semibold">{text("loading")}</h2>
                <p className="text-sm text-muted-foreground">{language === "th" ? "กำลังเตรียมข้อมูลบริษัทปัจจุบัน" : "Preparing current company settings."}</p>
              </div>
            </CardContent>
          </Card>
        )
      ) : (
        <>
          <Card>
            <CardContent className="grid gap-2 p-3">
              <div className="grid gap-2 lg:grid-cols-[minmax(0,1fr)_auto_auto_auto_auto]">
                <label className="relative block min-w-0">
                  <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                  <Input className="pl-9" placeholder={`${text("search")} ${title}`} value={query} onChange={(event) => setQuery(event.target.value)} />
                </label>
                <div className="flex min-h-9 items-center gap-2 rounded-lg border border-border bg-background px-2 text-sm shadow-sm">
                  <span className="font-medium text-muted-foreground">{language === "th" ? "ทั้งหมด" : "Total"}</span>
                  <b className="text-foreground">{totalAllRecords.toLocaleString(localeOf(language))}</b>
                  <span className="h-4 w-px bg-border" aria-hidden="true" />
                  <span className="font-medium text-muted-foreground">{text("active")}</span>
                  <b className="text-foreground">{visibleRecords.filter(isActiveRecord).length.toLocaleString(localeOf(language))}</b>
                </div>
                {currentConfig.slug === "productunit" && canEdit ? (
                  <Button type="button" variant="outline" onClick={() => void openStandardUnitDialog()} disabled={loading || saving || !auth}>
                    <DownloadCloud />
                    {text("findStandardUnits")}
                  </Button>
                ) : null}
                <Button type="button" variant="outline" onClick={() => void loadRecords(auth, workspace, config)} disabled={loading || !auth}>
                  {loading ? <Loader2 className="animate-spin" /> : <RefreshCcw />}
                  {text("refresh")}
                </Button>
                {canEdit ? (
                  <Button type="button" onClick={openCreate} disabled={!auth}>
                    <Plus />
                    {text("add")}
                  </Button>
                ) : null}
              </div>
            </CardContent>
          </Card>

          {loading && !records.length ? (
            <Card><CardContent className="flex min-h-40 items-center justify-center gap-2 p-4 text-sm text-muted-foreground"><Loader2 className="animate-spin" />{text("loading")}</CardContent></Card>
          ) : visibleRecords.length || formOpen ? (
            <SettingDataList
              auth={auth}
              config={config}
              dateTimeScope={dateTimeScope}
              dictionary={backendLanguage}
              editing={editing}
              form={form}
              formOpen={formOpen}
              language={language}
              onCloseForm={() => {
                if (!saving) setFormOpen(false);
              }}
              onDelete={deleteRecord}
              onEdit={openEdit}
              onResetPassword={resetUserPassword}
              onSelect={(record) => {
                setSelectedRecordId(recordId(record, config));
                setFormOpen(false);
                setEditing(null);
              }}
              onSubmit={saveRecord}
              onToggleAccess={toggleUserAccess}
              hasMore={hasMoreRecords}
              loadingMore={loadingMore}
              onLoadMore={loadMoreRecords}
              records={visibleRecords}
              totalRecords={displayTotal}
              saving={saving}
              selectedRecord={selectedRecord}
              setForm={setForm}
              text={text}
              workspace={workspace}
            />
          ) : (
            <Card>
              <CardContent className="grid min-h-44 place-items-center p-4 text-center">
                <div className="grid gap-2">
                  <AlertCircle className="mx-auto size-10 text-muted-foreground" />
                  <h2 className="text-base font-semibold">{text("empty")}</h2>
                  <p className="text-sm text-muted-foreground">{canEdit ? text("emptyHint") : text("readOnlyEmptyHint")}</p>
                  {canEdit ? <Button type="button" onClick={openCreate} disabled={!auth}><Plus />{text("add")}</Button> : null}
                </div>
              </CardContent>
            </Card>
          )}

          {standardUnitDialog.open ? (
            <StandardUnitDialog
              dialog={standardUnitDialog}
              language={language}
              onClear={clearStandardUnits}
              onClose={() => {
                if (!standardUnitDialog.saving) setStandardUnitDialog(emptyStandardUnitDialog);
              }}
              onQueryChange={(value) => setStandardUnitDialog((current) => ({ ...current, query: value }))}
              onRefresh={() => void loadStandardUnitOptions()}
              onSave={() => void saveStandardUnits()}
              onSelectAll={selectAllStandardUnits}
              onToggle={toggleStandardUnit}
              text={text}
            />
          ) : null}
        </>
      )}
      {confirmationDialog}
    </div>
  );

  if (embedded) return <section className="grid w-full min-w-0 gap-3">{content}</section>;
  return <main className="min-h-dvh w-full overflow-x-hidden bg-background p-2 text-foreground sm:p-3">{content}</main>;
}

function SettingCard({
  auth,
  config,
  dictionary,
  language,
  onDelete,
  onEdit,
  onResetPassword,
  onToggleAccess,
  record,
  saving,
  text,
  workspace,
}: {
  auth: AuthSession | null;
  config: SystemSettingConfig;
  dictionary: BackendLanguageDictionary;
  language: LanguageCode;
  onDelete: (record: SettingRecord) => void;
  onEdit: (record: SettingRecord) => void;
  onResetPassword: (record: SettingRecord) => void;
  onToggleAccess: (record: SettingRecord, disabled: boolean) => void;
  record: SettingRecord;
  saving: boolean;
  text: (key: keyof typeof uiEn) => string;
  workspace: WorkspaceSession | null;
}) {
  const id = recordId(record, config);
  const title = recordTitle(record, config, language);
  const branchCaption = config.slug === "department" || config.kind === "restaurant-setting" ? recordBranchCaption(record) : "";
  const isUser = config.slug === "user";
  const isCreator = isUser && isCreatorRecord(record, workspace);
  const isSelf = isUser && isSelfUserRecord(record, auth);
  const accessDisabled = isUser && userAccessDisabled(record);
  const canEdit = config.editable !== false;
  return (
    <Card className={cn(
      "min-w-0 shadow-sm",
      isCreator && "border-amber-200 bg-amber-50/40 dark:border-amber-900 dark:bg-amber-950/15",
      !isCreator && accessDisabled && "border-amber-200 bg-amber-50/30 dark:border-amber-900 dark:bg-amber-950/10",
    )}>
      <CardContent className="grid gap-2 p-3">
        <div className="flex min-w-0 items-start justify-between gap-2">
          <div className="min-w-0">
            <h2 className="truncate text-base font-semibold">{title || id || "-"}</h2>
            <p className="truncate text-xs text-muted-foreground">{text("id")}: {id || "-"}</p>
            <div className="mt-1 flex flex-wrap gap-1">
              {branchCaption ? <Badge variant="outline" className="max-w-full truncate">{text("branch")}: {branchCaption}</Badge> : null}
              {isCreator ? <Badge variant="warning" className="gap-1"><Crown className="size-3" />{text("creator")}</Badge> : null}
              {isUser ? (
                <Badge variant={accessDisabled ? "warning" : "success"} className="gap-1">
                  {accessDisabled ? <UserX className="size-3" /> : <UserCheck className="size-3" />}
                  {accessDisabled ? text("accessTemporarilyDisabled") : text("accessEnabled")}
                </Badge>
              ) : null}
            </div>
          </div>
          <div className="flex shrink-0 flex-wrap justify-end gap-1">
            {canEdit ? (
              <Button type="button" size="icon" variant="outline" onClick={() => onEdit(record)} disabled={isSelf} aria-label={isSelf ? text("selfPermissionCannotEdit") : text("edit")} title={isSelf ? text("selfPermissionCannotEdit") : text("edit")}>
                <Edit3 />
              </Button>
            ) : null}
            {isUser && !isCreator && !isSelf ? (
              <Button
                type="button"
                size="icon"
                variant="outline"
                onClick={() => onResetPassword(record)}
                disabled={saving}
                aria-label={text("resetPassword")}
                title={text("resetPassword")}
              >
                <KeyRound />
              </Button>
            ) : null}
            {isUser && !isCreator && !isSelf ? (
              <Button
                type="button"
                size="sm"
                variant={accessDisabled ? "outline" : "destructive"}
                onClick={() => onToggleAccess(record, !accessDisabled)}
                disabled={saving}
                aria-label={accessDisabled ? text("enableAccess") : text("temporarilyDisableAccess")}
                title={accessDisabled ? text("enableAccess") : text("temporarilyDisableAccess")}
              >
                {accessDisabled ? <UserCheck /> : <UserX />}
                <span className="hidden sm:inline">{accessDisabled ? text("enableAccess") : text("temporarilyDisableAccess")}</span>
              </Button>
            ) : null}
            {config.kind === "company" || !canEdit ? null : (
              <Button type="button" size="icon" variant="outline" onClick={() => onDelete(record)} disabled={isCreator || isSelf} aria-label={isCreator ? text("creatorCannotDelete") : isSelf ? text("selfPermissionCannotEdit") : text("delete")} title={isCreator ? text("creatorCannotDelete") : isSelf ? text("selfPermissionCannotEdit") : text("delete")}>
                <Trash2 />
              </Button>
            )}
          </div>
        </div>
        <div className="grid gap-1 text-xs text-muted-foreground">
          {config.fields.slice(0, 4).map((field) => (
            <span className="flex min-w-0 items-center justify-between gap-2 rounded-xl border border-border bg-background px-2 py-1.5" key={field.key}>
              <span className="truncate">{fieldLabel(field, language, config, dictionary)}</span>
              <b className="min-w-0 max-w-[60%] truncate text-right text-foreground">{shortValue(getByPath(record, field.key), language)}</b>
            </span>
          ))}
        </div>
      </CardContent>
    </Card>
  );
}

function SettingDataList({
  auth,
  config,
  dateTimeScope,
  dictionary,
  editing,
  form,
  formOpen,
  language,
  onCloseForm,
  onDelete,
  onEdit,
  onResetPassword,
  onSelect,
  onSubmit,
  onToggleAccess,
  hasMore,
  loadingMore,
  onLoadMore,
  records,
  totalRecords,
  saving,
  selectedRecord,
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
  formOpen: boolean;
  language: LanguageCode;
  onCloseForm: () => void;
  onDelete: (record: SettingRecord) => void;
  onEdit: (record: SettingRecord) => void;
  onResetPassword: (record: SettingRecord) => void;
  onSelect: (record: SettingRecord) => void;
  onSubmit: (event: FormEvent<HTMLFormElement>) => void;
  onToggleAccess: (record: SettingRecord, disabled: boolean) => void;
  hasMore: boolean;
  loadingMore: boolean;
  onLoadMore: () => void;
  records: SettingRecord[];
  totalRecords: number;
  saving: boolean;
  selectedRecord: SettingRecord | null;
  setForm: (form: FormState) => void;
  text: (key: keyof typeof uiEn) => string;
  workspace: WorkspaceSession | null;
}) {
  const selectedId = selectedRecord ? recordId(selectedRecord, config) : "";
  const editingId = editing ? recordId(editing, config) : "";
  const listScrollRef = useRef<HTMLDivElement | null>(null);
  const loadMoreRef = useRef<HTMLDivElement | null>(null);
  const [listMaxHeight, setListMaxHeight] = useState(420);
  const [leftPanePercent, setLeftPanePercent] = useState(30);
  const [hoveredRecordId, setHoveredRecordId] = useState("");
  const columns = useMemo(() => settingListColumns(config, language, dictionary, text), [config, dictionary, language, text]);

  useEffect(() => {
    const target = loadMoreRef.current;
    if (!target || !hasMore || loadingMore) return;
    const observer = new IntersectionObserver((entries) => {
      if (entries.some((entry) => entry.isIntersecting)) onLoadMore();
    }, { root: listScrollRef.current, rootMargin: "240px 0px", threshold: 0.01 });
    observer.observe(target);
    return () => observer.disconnect();
  }, [hasMore, loadingMore, onLoadMore]);

  useLayoutEffect(() => {
    const element = listScrollRef.current;
    if (!element) return;
    const updateHeight = () => {
      const top = element.getBoundingClientRect().top;
      const bottomGap = 12;
      setListMaxHeight(Math.max(260, Math.floor(window.innerHeight - top - bottomGap)));
    };
    updateHeight();
    const resizeObserver = new ResizeObserver(updateHeight);
    resizeObserver.observe(document.body);
    resizeObserver.observe(element);
    window.addEventListener("resize", updateHeight);
    window.addEventListener("orientationchange", updateHeight);
    return () => {
      resizeObserver.disconnect();
      window.removeEventListener("resize", updateHeight);
      window.removeEventListener("orientationchange", updateHeight);
    };
  }, [formOpen, records.length, selectedId]);

  const startPaneResize = useCallback((event: ReactPointerEvent<HTMLDivElement>) => {
    const container = event.currentTarget.parentElement;
    if (!container) return;
    event.preventDefault();
    const rect = container.getBoundingClientRect();
    const update = (clientX: number) => {
      const next = ((clientX - rect.left) / rect.width) * 100;
      setLeftPanePercent(Math.min(95, Math.max(5, next)));
    };
    update(event.clientX);
    const onMove = (moveEvent: PointerEvent) => update(moveEvent.clientX);
    const onUp = () => {
      window.removeEventListener("pointermove", onMove);
      window.removeEventListener("pointerup", onUp);
      document.body.style.cursor = "";
      document.body.style.userSelect = "";
    };
    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";
    window.addEventListener("pointermove", onMove);
    window.addEventListener("pointerup", onUp);
  }, []);

  return (
    <Card className="overflow-hidden shadow-sm">
      <CardContent className="flex min-h-[520px] flex-col gap-0 p-0 lg:flex-row">
        <section className="min-h-0 min-w-0 border-b border-border lg:border-b-0" style={{ flexBasis: `${leftPanePercent}%` }} aria-label={text("list")}>
          <div className="flex items-center justify-between gap-2 border-b border-border bg-muted/30 px-3 py-1.5">
            <div className="min-w-0">
              <h2 className="truncate text-sm font-semibold">{systemSettingTitle(config, language, dictionary)}</h2>
              <p className="text-xs text-muted-foreground">{records.length.toLocaleString(localeOf(language))} / {totalRecords.toLocaleString(localeOf(language))} {text("items")}</p>
            </div>
          </div>
          <div ref={listScrollRef} className="min-h-[260px] w-full overflow-x-hidden overflow-y-auto" style={{ maxHeight: listMaxHeight }}>
            <div className="sticky top-0 z-10 flex flex-wrap items-center gap-2 border-b border-border bg-muted/95 px-3 py-1.5 text-xs font-semibold text-muted-foreground backdrop-blur">
              {columns.map((column, columnIndex) => (
                <span className={cn("min-w-0 break-words", columnIndex === 0 ? "basis-36 grow-[2]" : "basis-24 grow")} key={column.key}>{column.label}</span>
              ))}
              <span className="basis-24 grow text-right">{language === "th" ? "จัดการ" : "Actions"}</span>
            </div>
            {records.map((record, index) => {
              const id = recordId(record, config);
              const active = id === selectedId;
              const isHovered = id === hoveredRecordId;
              const isEditing = Boolean(editingId && id === editingId);
              const isCreator = config.slug === "user" && isCreatorRecord(record, workspace);
              const accessDisabled = config.slug === "user" && userAccessDisabled(record);
              const rowStyle = settingListRowStyle({ active, index, isEditing, isHovered });
              return (
                <div
                  className={cn(
                    "flex w-full cursor-pointer flex-wrap items-center gap-2 overflow-x-hidden border-b border-border px-3 py-1.5 text-left text-sm transition-colors last:border-b-0",
                    isEditing
                      ? "bg-amber-100 text-amber-950 dark:bg-amber-950/40 dark:text-amber-100"
                      : active
                        ? "bg-primary/10 text-primary"
                        : index % 2 === 0
                          ? "bg-background"
                          : "bg-muted/30",
                  )}
                  style={rowStyle}
                  aria-current={isEditing ? "step" : active ? "true" : undefined}
                  key={`${id}-${index}`}
                  onClick={() => onSelect(record)}
                  onMouseEnter={() => setHoveredRecordId(id)}
                  onMouseLeave={() => setHoveredRecordId((current) => current === id ? "" : current)}
                  onFocus={() => setHoveredRecordId(id)}
                  onBlur={() => setHoveredRecordId((current) => current === id ? "" : current)}
                  onKeyDown={(event) => {
                    if (event.key === "Enter" || event.key === " ") {
                      event.preventDefault();
                      onSelect(record);
                    }
                  }}
                  role="button"
                  tabIndex={0}
                >
                  {columns.map((column, columnIndex) => (
                    <span className={cn("min-w-0 break-words", columnIndex === 0 ? "basis-36 grow-[2]" : "basis-24 grow")} key={column.key}>
                      {column.render(record, { isCreator, accessDisabled, id, columnIndex })}
                    </span>
                  ))}
                  <span className="flex min-w-0 basis-24 grow flex-wrap justify-end gap-1" onClick={(event) => event.stopPropagation()}>
                    {config.editable !== false ? (
                      <Button type="button" size="icon" variant="outline" onClick={() => onEdit(record)} disabled={config.slug === "user" && isSelfUserRecord(record, auth)} aria-label={text("edit")} title={text("edit")}>
                        <Edit3 />
                      </Button>
                    ) : null}
                    {config.slug === "user" && !isCreator && !isSelfUserRecord(record, auth) ? (
                      <Button type="button" size="icon" variant="outline" onClick={() => onResetPassword(record)} disabled={saving} aria-label={text("resetPassword")} title={text("resetPassword")}>
                        <KeyRound />
                      </Button>
                    ) : null}
                    {config.kind === "company" || config.editable === false ? null : (
                      <Button type="button" size="icon" variant="outline" onClick={() => onDelete(record)} disabled={isCreator || isSelfUserRecord(record, auth)} aria-label={text("delete")} title={text("delete")}>
                        <Trash2 />
                      </Button>
                    )}
                  </span>
                </div>
              );
            })}
            <div ref={loadMoreRef} className="flex min-h-10 items-center justify-center py-2 text-xs text-muted-foreground">
              {loadingMore ? <><Loader2 className="mr-2 size-4 animate-spin" />{text("loading")}</> : hasMore ? text("loading") : null}
            </div>
          </div>
        </section>

        <div
          className="hidden w-1.5 shrink-0 cursor-col-resize bg-border/70 transition hover:bg-primary/50 lg:block"
          onPointerDown={startPaneResize}
          role="separator"
          aria-orientation="vertical"
          aria-label={language === "th" ? "ปรับความกว้างรายการและรายละเอียด" : "Resize list and detail panes"}
        />

        <section className="min-h-0 min-w-0 flex-1 p-3 lg:sticky lg:top-3 lg:self-start" aria-label={text("details")}>
          {formOpen ? (
            <SettingFormDialog
              inline
              auth={auth}
              config={config}
              dateTimeScope={dateTimeScope}
              dictionary={dictionary}
              editing={editing}
              form={form}
              language={language}
              onClose={onCloseForm}
              onSubmit={onSubmit}
              saving={saving}
              setForm={setForm}
              text={text}
              workspace={workspace}
            />
          ) : selectedRecord ? (
            <SettingDetailPanel
              auth={auth}
              config={config}
              dictionary={dictionary}
              language={language}
              onDelete={onDelete}
              onEdit={onEdit}
              onResetPassword={onResetPassword}
              onToggleAccess={onToggleAccess}
              record={selectedRecord}
              saving={saving}
              text={text}
              workspace={workspace}
            />
          ) : (
            <div className="grid min-h-44 place-items-center rounded-2xl border border-dashed border-border bg-muted/20 p-4 text-center text-sm text-muted-foreground">
              {text("empty")}
            </div>
          )}
        </section>
      </CardContent>
    </Card>
  );
}

type SettingListColumn = {
  key: string;
  label: string;
  render: (record: SettingRecord, meta: { accessDisabled: boolean; columnIndex: number; id: string; isCreator: boolean }) => ReactNode;
};

function settingListRowStyle({
  active,
  index,
  isEditing,
  isHovered,
}: {
  active: boolean;
  index: number;
  isEditing: boolean;
  isHovered: boolean;
}): CSSProperties | undefined {
  if (!isHovered) return undefined;
  if (isEditing) {
    return {
      backgroundColor: "rgba(251, 191, 36, 0.18)",
      borderColor: "rgba(245, 158, 11, 0.32)",
      color: "var(--foreground)",
    };
  }
  if (active) {
    return {
      backgroundColor: "rgba(14, 165, 233, 0.12)",
      borderColor: "rgba(14, 116, 144, 0.24)",
      color: "var(--foreground)",
    };
  }
  const backgroundColor = index % 2 === 0
    ? "rgba(14, 165, 233, 0.055)"
    : "rgba(14, 165, 233, 0.09)";
  return {
    backgroundColor,
    borderColor: "rgba(14, 116, 144, 0.18)",
    color: "var(--foreground)",
  };
}

function settingListColumns(
  config: SystemSettingConfig,
  language: LanguageCode,
  dictionary: BackendLanguageDictionary,
  text: (key: keyof typeof uiEn) => string,
): SettingListColumn[] {
  if (config.slug === "user") {
    const roleField = config.fields.find((field) => field.key === "role");
    return [
      {
        key: "username",
        label: language === "th" ? "รหัสผู้ใช้ หรือ email" : "User code or email",
        render: (record, meta) => (
          <span className="flex min-w-0 items-center gap-1.5">
            <b className="min-w-0 break-words">{stringValue(record.username ?? record.email ?? meta.id) || "-"}</b>
            {meta.isCreator ? <Badge variant="warning" className="shrink-0 gap-1"><Crown className="size-3" />{text("creator")}</Badge> : null}
          </span>
        ),
      },
      {
        key: "user_profile_name",
        label: language === "th" ? "ชื่อผู้ใช้งาน" : "User name",
        render: (record) => stringValue(record.user_profile_name ?? record.userprofilename ?? record.name) || "-",
      },
      {
        key: "email",
        label: language === "th" ? "อีเมล" : "Email",
        render: (record) => stringValue(record.email) || "-",
      },
      {
        key: "role",
        label: language === "th" ? "สิทธิ์" : "Role",
        render: (record) => roleField ? fieldDisplayValue(roleField, record.role, language) : shortValue(record.role, language),
      },
      {
        key: "status",
        label: language === "th" ? "สถานะ" : "Status",
        render: (_record, meta) => <Badge variant={meta.accessDisabled ? "warning" : "success"}>{meta.accessDisabled ? text("accessTemporarilyDisabled") : text("accessEnabled")}</Badge>,
      },
    ];
  }

  if (config.slug === "employee") {
    return [
      {
        key: "code",
        label: language === "th" ? "รหัสพนักงาน" : "Employee code",
        render: (record, meta) => <b>{stringValue(record.code ?? meta.id) || "-"}</b>,
      },
      {
        key: "name",
        label: language === "th" ? "ชื่อพนักงาน" : "Employee name",
        render: (record) => stringValue(record.name ?? record.employeeName) || "-",
      },
      {
        key: "email",
        label: language === "th" ? "อีเมล" : "Email",
        render: (record) => stringValue(record.email) || "-",
      },
      {
        key: "status",
        label: language === "th" ? "สถานะ" : "Status",
        render: (record) => <Badge variant={isActiveRecord(record) ? "success" : "warning"}>{isActiveRecord(record) ? text("active") : language === "th" ? "ปิดใช้งาน" : "Inactive"}</Badge>,
      },
    ];
  }

  const fields = config.fields.filter((field) => !["image-upload", "json", "language-configs", "language-list"].includes(field.type)).slice(0, 4);
  if (!fields.length) {
    return [
      {
        key: "title",
        label: systemSettingTitle(config, language, dictionary),
        render: (record, meta) => recordTitle(record, config, language) || meta.id || "-",
      },
    ];
  }
  return fields.map((field) => ({
    key: field.key,
    label: fieldLabel(field, language, config, dictionary),
    render: (record) => fieldDisplayValue(field, getByPath(record, field.key), language),
  }));
}

function SettingDetailPanel({
  auth,
  config,
  dictionary,
  language,
  onDelete,
  onEdit,
  onResetPassword,
  onToggleAccess,
  record,
  saving,
  text,
  workspace,
}: {
  auth: AuthSession | null;
  config: SystemSettingConfig;
  dictionary: BackendLanguageDictionary;
  language: LanguageCode;
  onDelete: (record: SettingRecord) => void;
  onEdit: (record: SettingRecord) => void;
  onResetPassword: (record: SettingRecord) => void;
  onToggleAccess: (record: SettingRecord, disabled: boolean) => void;
  record: SettingRecord;
  saving: boolean;
  text: (key: keyof typeof uiEn) => string;
  workspace: WorkspaceSession | null;
}) {
  const id = recordId(record, config);
  const title = recordTitle(record, config, language);
  const isUser = config.slug === "user";
  const isCreator = isUser && isCreatorRecord(record, workspace);
  const isSelf = isUser && isSelfUserRecord(record, auth);
  const accessDisabled = isUser && userAccessDisabled(record);
  const canEdit = config.editable !== false;
  return (
    <section className="grid gap-3">
      <header className="flex min-w-0 flex-wrap items-start justify-between gap-2">
        <div className="min-w-0">
          <h2 className="truncate text-lg font-semibold">{title || id || "-"}</h2>
          <p className="truncate text-sm text-muted-foreground">{text("id")}: {id || "-"}</p>
          <div className="mt-1 flex flex-wrap gap-1">
            {isCreator ? <Badge variant="warning" className="gap-1"><Crown className="size-3" />{text("creator")}</Badge> : null}
            {isUser ? <Badge variant={accessDisabled ? "warning" : "success"}>{accessDisabled ? text("accessTemporarilyDisabled") : text("accessEnabled")}</Badge> : null}
          </div>
        </div>
        <div className="flex flex-wrap justify-end gap-1">
          {canEdit ? (
            <Button type="button" size="sm" variant="outline" onClick={() => onEdit(record)} disabled={isSelf}>
              <Edit3 />{text("edit")}
            </Button>
          ) : null}
          {isUser && !isCreator && !isSelf ? (
            <Button type="button" size="sm" variant="outline" onClick={() => onResetPassword(record)} disabled={saving}>
              <KeyRound />{text("resetPassword")}
            </Button>
          ) : null}
          {isUser && !isCreator && !isSelf ? (
            <Button type="button" size="sm" variant={accessDisabled ? "outline" : "destructive"} onClick={() => onToggleAccess(record, !accessDisabled)} disabled={saving}>
              {accessDisabled ? <UserCheck /> : <UserX />}
              {accessDisabled ? text("enableAccess") : text("temporarilyDisableAccess")}
            </Button>
          ) : null}
          {config.kind === "company" || !canEdit ? null : (
            <Button type="button" size="sm" variant="outline" onClick={() => onDelete(record)} disabled={isCreator || isSelf}>
              <Trash2 />{text("delete")}
            </Button>
          )}
        </div>
      </header>
      <div className="grid gap-1.5 md:grid-cols-2">
        {config.fields.map((field) => (
          <div className="grid gap-1 rounded-xl border border-border bg-background px-2 py-1.5 text-sm" key={field.key}>
            <span className="text-xs font-medium text-muted-foreground">{fieldLabel(field, language, config, dictionary)}</span>
            <b className="min-w-0 break-words text-foreground">{fieldDisplayValue(field, getByPath(record, field.key), language)}</b>
          </div>
        ))}
      </div>
    </section>
  );
}

function SettingFormDialog({
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
  setForm: (form: FormState) => void;
  text: (key: keyof typeof uiEn) => string;
  workspace: WorkspaceSession | null;
}) {
  const formElement = (
      <form
        className={cn(
          "grid gap-3 rounded-2xl border border-border bg-card p-3 text-foreground shadow-sm",
          inline ? "w-full" : "max-h-[calc(100dvh-24px)] w-[min(860px,calc(100vw-24px))] overflow-hidden shadow-xl",
        )}
        onSubmit={onSubmit}
        role={inline ? undefined : "dialog"}
        aria-modal={inline ? undefined : true}
        aria-label={editing ? text("edit") : text("newItem")}
      >
        <header className="flex min-w-0 items-center justify-between gap-2">
          <div className="min-w-0">
            <h2 className="truncate text-lg font-semibold">{editing ? text("edit") : text("newItem")}: {systemSettingTitle(config, language, dictionary)}</h2>
            <div className="flex flex-wrap gap-1 text-xs text-muted-foreground">
              <span>{config.route}</span>
              {config.slug === "department" || config.kind === "restaurant-setting" ? (
                <Badge variant="outline" className="text-[11px]">{text("branch")}: {dateTimeScope.branchcode || dateTimeScope.branchguid || dateTimeScope.key}</Badge>
              ) : null}
            </div>
          </div>
          {inline ? null : <Button type="button" variant="outline" size="icon" onClick={onClose} disabled={saving} aria-label={text("close")}>
            <X />
          </Button>}
        </header>

        <div className={cn("grid min-h-0 gap-2 pr-1", inline ? "" : "overflow-y-auto")}>
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
            <div className="grid gap-2 md:grid-cols-2">
              {config.fields.map((field) => (
                <div className={fieldGridItemClass(field, config)} key={field.key}>
                  <FieldEditor auth={auth} config={config} dateTimeScope={dateTimeScope} dictionary={dictionary} field={field} form={form} language={language} setForm={setForm} workspace={workspace} />
                </div>
              ))}
            </div>
          )}
        </div>

        <footer className="flex flex-wrap justify-end gap-2">
          {inline ? null : <Button type="button" variant="outline" onClick={onClose} disabled={saving}>{text("cancel")}</Button>}
          <Button type="submit" disabled={saving}>
            {saving ? <Loader2 className="animate-spin" /> : <Save />}
            {text("save")}
          </Button>
        </footer>
      </form>
  );
  if (inline) return formElement;
  return <div className="dialog-backdrop" role="presentation">{formElement}</div>;
}

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
  setForm: (form: FormState) => void;
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
      keys: ["uid", "username", "user_profile_name", "email"],
      title: language === "th" ? "บัญชีเข้าสู่ระบบ" : "Sign-in account",
      description: loginHint,
    },
    {
      keys: ["role", "is_access_disabled"],
      title: language === "th" ? "สิทธิ์และสถานะ" : "Permission and status",
      description: language === "th" ? "กำหนดระดับสิทธิ์ในร้าน และเปิดหรือปิดการเข้าใช้งานของผู้ใช้นี้" : "Set the user's shop role and whether this user can access the system.",
    },
    {
      keys: ["position", "department", "line_user_id", "line_display_name"],
      title: language === "th" ? "ข้อมูลองค์กรและ LINE" : "Organization and LINE",
      description: language === "th" ? "ใช้สำหรับอ้างอิงตำแหน่ง แผนก และข้อมูล LINE ที่ผูกกับผู้ใช้งาน" : "Reference position, department, and LINE data linked to this user.",
    },
  ];

  return (
    <div className="grid gap-2">
      {sections.map((section) => {
        const fields = section.keys.map((key) => fieldsByKey.get(key)).filter((field): field is SystemSettingField => Boolean(field));
        if (fields.length === 0) return null;
        return (
          <section className="grid gap-2 rounded-2xl border border-border bg-background/70 p-2" key={section.title}>
            <header className="grid gap-0.5">
              <h3 className="text-sm font-semibold">{section.title}</h3>
              <p className="text-xs leading-snug text-muted-foreground">{section.description}</p>
            </header>
            <div className="grid items-start gap-2 md:grid-cols-2">
              {fields.map((field) => (
                <div className={fieldGridItemClass(field, config)} key={field.key}>
                  <FieldEditor auth={auth} config={config} dateTimeScope={dateTimeScope} dictionary={dictionary} field={field} form={form} language={language} setForm={setForm} workspace={workspace} />
                </div>
              ))}
            </div>
          </section>
        );
      })}
    </div>
  );
}

function fieldGridItemClass(field: SystemSettingField, config: SystemSettingConfig): string {
  if (config.kind === "company") return "min-w-0 md:col-span-2";
  if (
    field.type === "image-upload"
    || field.type === "json"
    || field.type === "language-configs"
    || field.type === "language-list"
    || field.type === "master-picker"
    || field.type === "names"
    || field.type === "textarea"
  ) {
    return "min-w-0 md:col-span-2";
  }
  if (
    (config.slug === "permission_definition" && field.key === "branches")
    || (config.slug === "approval_setting" && field.key === "approvals")
    || (config.slug === "permission_link" && (field.key === "employeeCode" || field.key === "permissionCodes" || field.key === "approvalCodes"))
  ) {
    return "min-w-0 md:col-span-2";
  }
  return "min-w-0";
}

function StandardUnitDialog({
  dialog,
  language,
  onClear,
  onClose,
  onQueryChange,
  onRefresh,
  onSave,
  onSelectAll,
  onToggle,
  text,
}: {
  dialog: StandardUnitDialogState;
  language: LanguageCode;
  onClear: () => void;
  onClose: () => void;
  onQueryChange: (value: string) => void;
  onRefresh: () => void;
  onSave: () => void;
  onSelectAll: () => void;
  onToggle: (unitcode: string, checked: boolean) => void;
  text: (key: keyof typeof uiEn) => string;
}) {
  return (
    <div className="dialog-backdrop" role="presentation">
      <section className="grid max-h-[calc(100dvh-24px)] w-[min(720px,calc(100vw-24px))] gap-3 overflow-hidden rounded-2xl border border-border bg-card p-3 text-foreground shadow-xl" role="dialog" aria-modal="true" aria-label={text("standardUnits")}>
        <header className="flex min-w-0 items-center justify-between gap-2">
          <div className="min-w-0">
            <p className="text-xs font-semibold uppercase text-muted-foreground">PRODUCT UNIT</p>
            <h2 className="truncate text-lg font-semibold">{text("standardUnits")}</h2>
          </div>
          <Button type="button" variant="outline" size="icon" onClick={onClose} disabled={dialog.saving} aria-label={text("close")}>
            <X />
          </Button>
        </header>

        <div className="grid gap-2 sm:grid-cols-[minmax(0,1fr)_auto]">
          <label className="relative block min-w-0">
            <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input className="pl-9" placeholder={text("search")} value={dialog.query} onChange={(event) => onQueryChange(event.target.value)} disabled={dialog.loading || dialog.saving} />
          </label>
          <Button type="button" variant="outline" onClick={onRefresh} disabled={dialog.loading || dialog.saving}>
            {dialog.loading ? <Loader2 className="animate-spin" /> : <Search />}
            {text("search")}
          </Button>
        </div>

        <div className="flex flex-wrap items-center justify-between gap-2 text-sm font-semibold">
          <Badge variant="outline">{text("selected")}: {dialog.selectedCodes.length.toLocaleString(localeOf(language))}/{dialog.options.length.toLocaleString(localeOf(language))}</Badge>
          <div className="flex flex-wrap gap-2">
            <Button type="button" variant="outline" size="sm" onClick={onSelectAll} disabled={dialog.loading || dialog.saving || dialog.options.length === 0}>
              {text("selectAll")}
            </Button>
            <Button type="button" variant="outline" size="sm" onClick={onClear} disabled={dialog.loading || dialog.saving || dialog.selectedCodes.length === 0}>
              {text("clearSelection")}
            </Button>
          </div>
        </div>

        {dialog.error ? (
          <div className="message error">
            <AlertCircle size={18} />
            <span>{dialog.error}</span>
          </div>
        ) : null}

        <div className="grid min-h-0 gap-2 overflow-y-auto pr-1">
          {dialog.loading ? (
            <div className="flex min-h-32 items-center justify-center gap-2 rounded-2xl border border-border bg-background p-4 text-sm font-semibold text-muted-foreground">
              <Loader2 className="animate-spin" />
              {text("loading")}
            </div>
          ) : dialog.options.length ? dialog.options.map((unit) => {
            const checked = dialog.selectedCodes.includes(unit.unitcode);
            return (
              <label className="flex min-w-0 cursor-pointer items-center gap-2 rounded-2xl border border-border bg-background px-3 py-2 text-sm font-semibold" key={unit.unitcode}>
                <input
                  className="size-4 shrink-0 accent-primary"
                  type="checkbox"
                  checked={checked}
                  disabled={dialog.saving}
                  onChange={(event) => onToggle(unit.unitcode, event.target.checked)}
                />
                <span className="min-w-0 flex-1 truncate">{unitDisplayName(unit, language)}</span>
                <b className="shrink-0 text-xs text-muted-foreground">{unit.unitcode}</b>
              </label>
            );
          }) : (
            <div className="grid min-h-32 place-items-center rounded-2xl border border-border bg-background p-4 text-center text-sm font-semibold text-muted-foreground">
              {text("noStandardUnits")}
            </div>
          )}
        </div>

        <footer className="flex flex-wrap justify-end gap-2">
          <Button type="button" variant="outline" onClick={onClose} disabled={dialog.saving}>{text("cancel")}</Button>
          <Button type="button" onClick={onSave} disabled={dialog.loading || dialog.saving || dialog.selectedCodes.length === 0}>
            {dialog.saving ? <Loader2 className="animate-spin" /> : <Plus />}
            {text("addSelected")}
          </Button>
        </footer>
      </section>
    </div>
  );
}

type MenuPermissionAction = "access" | "create" | "update" | "delete" | "own_only";
type ApprovalTarget = "pr" | "po" | "quotation" | "sale_order";

const menuPermissionActions: { key: MenuPermissionAction; textKey: keyof typeof uiEn }[] = [
  { key: "access", textKey: "readAccess" },
  { key: "create", textKey: "writeAccess" },
  { key: "update", textKey: "updateAccess" },
  { key: "delete", textKey: "delete" },
  { key: "own_only", textKey: "selfOnly" },
];

const approvalTargets: { key: ApprovalTarget; label: string }[] = [
  { key: "pr", label: "PR" },
  { key: "po", label: "PO" },
  { key: "quotation", label: "Quotation" },
  { key: "sale_order", label: "Sale Order" },
];

function ApprovalSettingEditor({
  dictionary,
  form,
  setForm,
}: {
  dictionary: BackendLanguageDictionary;
  form: FormState;
  setForm: (form: FormState) => void;
}) {
  const approvals = approvalsFromForm(form.approvals);

  function updateApproval(target: ApprovalTarget, key: "enabled" | "approval_role" | "max_approval_amount", value: boolean | number) {
    const nextApprovals = approvalsFromForm(form.approvals);
    const currentApproval = isRecord(nextApprovals[target]) ? { ...nextApprovals[target] } : {};
    currentApproval[key] = value;
    nextApprovals[target] = currentApproval;
    setForm({ ...form, approvals: nextApprovals });
  }

  return (
    <section className="grid gap-3 rounded-2xl border border-border bg-background p-3 md:col-span-2">
      <h3 className="text-base font-semibold">{backendText(dictionary, "approval", "Approval")}</h3>
      <div className="grid gap-2 md:grid-cols-2">
        {approvalTargets.map((target) => {
          const approval: SettingRecord = isRecord(approvals[target.key]) ? approvals[target.key] as SettingRecord : {};
          return (
            <div className="grid gap-2 rounded-xl border border-border bg-card p-2" key={target.key}>
              <label className="flex items-center gap-2 text-sm font-semibold">
                <input
                  className="size-4 accent-primary"
                  type="checkbox"
                  checked={Boolean(approval.enabled)}
                  onChange={(event) => updateApproval(target.key, "enabled", event.target.checked)}
                />
                <span>{target.label}</span>
              </label>
              <label className="grid gap-1 text-xs font-semibold">
                <span>{approvalLabelText(dictionary, "approval_role", "Approval role")}</span>
                <Input
                  type="number"
                  min="0"
                  value={String(approval.approval_role ?? "")}
                  onChange={(event) => updateApproval(target.key, "approval_role", Number(event.target.value || 0))}
                />
              </label>
              <label className="grid gap-1 text-xs font-semibold">
                <span>{approvalLabelText(dictionary, "max_approval_amount_label", "Maximum approval amount")}</span>
                <Input
                  type="number"
                  min="0"
                  value={String(approval.max_approval_amount ?? "")}
                  onChange={(event) => updateApproval(target.key, "max_approval_amount", Number(event.target.value || 0))}
                />
              </label>
            </div>
          );
        })}
      </div>
    </section>
  );
}

function approvalsFromForm(value: unknown): SettingRecord {
  if (isRecord(value)) return { ...value };
  if (typeof value !== "string" || !value.trim()) return {};
  const parsed = safeJsonParse(value, {});
  return isRecord(parsed) ? { ...parsed } : {};
}

function approvalLabelText(dictionary: BackendLanguageDictionary, key: string, fallback: string): string {
  const value = backendText(dictionary, key, fallback);
  return value === key ? fallback : value;
}

function permissionLinkOption(record: SettingRecord, isApproval: boolean): PermissionLinkOption {
  const code = stringValue(isApproval ? record.approvalCode : record.permissionCode);
  const name = stringValue(isApproval ? record.approvalName : record.permissionName) || code;
  const description = stringValue(record.description);
  return { code, description, isActive: Boolean(record.isActive ?? record.isactive ?? true), name };
}

function stringArrayFromForm(value: unknown): string[] {
  if (Array.isArray(value)) return uniqueStrings(value.map(stringValue).filter(Boolean));
  if (typeof value !== "string") return [];
  const trimmed = value.trim();
  if (!trimmed) return [];
  try {
    const parsed = JSON.parse(trimmed) as unknown;
    return Array.isArray(parsed) ? uniqueStrings(parsed.map(stringValue).filter(Boolean)) : [];
  } catch {
    return uniqueStrings(trimmed.split(",").map((item) => item.trim()).filter(Boolean));
  }
}

function uniqueStrings(values: string[]): string[] {
  return Array.from(new Set(values.filter(Boolean)));
}

type PermissionLinkUserOption = {
  code: string;
  name: string;
  subtitle: string;
  isDisabled: boolean;
};

function permissionLinkUserOption(record: SettingRecord): PermissionLinkUserOption {
  const code = stringValue(record.username ?? record.email ?? record.uid);
  const name = stringValue(record.name ?? record.email ?? record.username) || code;
  const subtitle = stringValue(record.email) || stringValue(record.uid);
  return { code, name, subtitle, isDisabled: Boolean(record.is_access_disabled ?? record.is_access_disabled ?? false) };
}

function PermissionLinkUserSelector({
  auth,
  dictionary,
  field,
  form,
  language,
  setForm,
  workspace,
}: {
  auth: AuthSession | null;
  dictionary: BackendLanguageDictionary;
  field: SystemSettingField;
  form: FormState;
  language: LanguageCode;
  setForm: (form: FormState) => void;
  workspace: WorkspaceSession | null;
}) {
  const userConfig = useMemo(() => getSystemSettingConfig("user"), []);
  const [users, setUsers] = useState<PermissionLinkUserOption[]>([]);
  const [query, setQuery] = useState("");
  const [debouncedQuery, setDebouncedQuery] = useState("");
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const selectedCode = stringValue(form.employeeCode);
  const selectedName = stringValue(form.employeeName);
  const fallbackLabel = field.label[language] ?? field.label.en ?? field.label.th;
  const labelKey = fieldBackendKeys[`permission_link.${field.key}`] ?? field.key;
  const translatedLabel = backendText(dictionary, labelKey, fallbackLabel);
  const label = translatedLabel === labelKey || translatedLabel === field.key ? fallbackLabel : translatedLabel;
  const searchText = debouncedQuery.trim();
  const canSearch = searchText.length >= 2;

  useEffect(() => {
    const timer = window.setTimeout(() => setDebouncedQuery(query), 250);
    return () => window.clearTimeout(timer);
  }, [query]);

  useEffect(() => {
    let cancelled = false;
    async function loadUsers() {
      if (!auth || !workspace || !userConfig || !canSearch) {
        setUsers([]);
        setLoading(false);
        setError("");
        return;
      }
      setLoading(true);
      setError("");
      try {
        const searchParams = new URLSearchParams({
          limit: "20",
          offset: "0",
          q: searchText,
          shopid: workspace.shop.shopid,
        });
        const response = await fetch(`/api/system-settings/${userConfig.slug}?${searchParams.toString()}`, {
          headers: requestHeaders(auth),
          cache: "no-store",
        });
        const payload = await response.json() as unknown;
        if (!response.ok || isFailed(payload)) throw new Error(extractMessage(payload) ?? backendText(dictionary, "request_failed", "Request failed."));
        if (cancelled) return;
        setUsers(normalizeRecords(payload, userConfig).map(permissionLinkUserOption).filter((user) => user.code));
      } catch (loadError) {
        if (!cancelled) setError(loadError instanceof Error && loadError.message ? loadError.message : backendText(dictionary, "request_failed", "Request failed."));
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    void loadUsers();
    return () => {
      cancelled = true;
    };
  }, [auth, canSearch, dictionary, searchText, userConfig, workspace]);

  function choose(user: PermissionLinkUserOption) {
    if (user.isDisabled) return;
    setForm({ ...form, employeeCode: user.code, employeeName: user.name });
    setQuery("");
    setUsers([]);
  }

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-3 text-sm font-semibold md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span>{label}{field.required ? " *" : ""}</span>
        {selectedCode ? <Badge variant="success">{selectedCode}</Badge> : <Badge variant="outline">{backendText(dictionary, "select_user", "Select user")}</Badge>}
      </div>
      <Input
        value={query}
        onChange={(event) => setQuery(event.target.value)}
        placeholder={backendText(dictionary, "search", "Search")}
      />
      {selectedCode ? (
        <div className="grid gap-1 rounded-xl border border-primary/40 bg-primary/10 p-2 text-primary">
          <span className="truncate font-semibold">{selectedName || selectedCode}</span>
          <span className="truncate text-xs">code: {selectedCode}</span>
        </div>
      ) : null}
      {error ? <p className="rounded-xl border border-destructive/40 bg-destructive/10 px-2 py-1 text-xs text-destructive">{error}</p> : null}
      <div className="grid max-h-60 gap-2 overflow-y-auto pr-1 md:grid-cols-2">
        {!canSearch ? (
          <div className="rounded-xl border border-border bg-card p-3 text-sm text-muted-foreground md:col-span-2">
            {backendText(dictionary, "search_user_hint", "Type at least 2 characters to search users.")}
          </div>
        ) : loading ? (
          <div className="flex min-h-16 items-center gap-2 rounded-xl border border-border bg-card p-3 text-muted-foreground md:col-span-2">
            <Loader2 className="animate-spin" />
            {backendText(dictionary, "loading", "Loading data")}
          </div>
        ) : users.length ? users.map((user) => {
          const checked = selectedCode === user.code;
          return (
            <button
              className={cn(
                "grid min-h-16 gap-1 rounded-xl border p-2 text-left transition-colors",
                checked ? "border-primary bg-primary/10 text-primary" : "border-border bg-card text-foreground hover:bg-muted/60",
                user.isDisabled && "cursor-not-allowed opacity-50",
              )}
              disabled={user.isDisabled}
              key={user.code}
              onClick={() => choose(user)}
              type="button"
            >
              <span className="flex min-w-0 items-center gap-2">
                {checked ? <Check className="size-4 shrink-0" /> : <UserRound className="size-4 shrink-0 text-muted-foreground" />}
                <span className="min-w-0 truncate font-semibold">{user.name}</span>
              </span>
              <span className="truncate text-xs text-muted-foreground">code: {user.code}</span>
              {user.subtitle ? <span className="truncate text-xs font-normal text-muted-foreground">{user.subtitle}</span> : null}
            </button>
          );
        }) : (
          <div className="rounded-xl border border-border bg-card p-3 text-sm text-muted-foreground md:col-span-2">
            {backendText(dictionary, "empty_data", "No data")}
          </div>
        )}
      </div>
    </section>
  );
}

type PermissionLinkOption = {
  code: string;
  description: string;
  isActive: boolean;
  name: string;
};

function PermissionLinkMultiSelectEditor({
  auth,
  dictionary,
  field,
  form,
  language,
  setForm,
  workspace,
}: {
  auth: AuthSession | null;
  dictionary: BackendLanguageDictionary;
  field: SystemSettingField;
  form: FormState;
  language: LanguageCode;
  setForm: (form: FormState) => void;
  workspace: WorkspaceSession | null;
}) {
  const isApproval = field.key === "approvalCodes";
  const sourceConfig = useMemo(() => getSystemSettingConfig(isApproval ? "approval_setting" : "permission_definition"), [isApproval]);
  const [options, setOptions] = useState<PermissionLinkOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const selectedCodes = stringArrayFromForm(form[field.key]);
  const selectedUserCode = stringValue(form.employeeCode);
  const fallbackLabel = field.label[language] ?? field.label.en ?? field.label.th;
  const labelKey = fieldBackendKeys[`permission_link.${field.key}`] ?? field.key;
  const translatedLabel = backendText(dictionary, labelKey, fallbackLabel);
  const label = translatedLabel === labelKey || translatedLabel === field.key ? fallbackLabel : translatedLabel;

  useEffect(() => {
    let cancelled = false;
    async function loadOptions() {
      if (!auth || !workspace || !sourceConfig) {
        setOptions([]);
        return;
      }
      setLoading(true);
      setError("");
      try {
        const searchParams = new URLSearchParams({
          limit: "1000",
          offset: "0",
          shopid: workspace.shop.shopid,
        });
        const response = await fetch(`/api/system-settings/${sourceConfig.slug}?${searchParams.toString()}`, {
          headers: requestHeaders(auth),
          cache: "no-store",
        });
        const payload = await response.json() as unknown;
        if (!response.ok || isFailed(payload)) throw new Error(extractMessage(payload) ?? backendText(dictionary, "request_failed", "Request failed."));
        if (cancelled) return;
        setOptions(normalizeRecords(payload, sourceConfig).map((record) => permissionLinkOption(record, isApproval)).filter((option) => option.code));
      } catch (loadError) {
        if (!cancelled) setError(loadError instanceof Error && loadError.message ? loadError.message : backendText(dictionary, "request_failed", "Request failed."));
      } finally {
        if (!cancelled) setLoading(false);
      }
    }
    void loadOptions();
    return () => {
      cancelled = true;
    };
  }, [auth, dictionary, isApproval, sourceConfig, workspace]);

  function toggle(code: string, checked: boolean) {
    if (!selectedUserCode) return;
    const next = checked ? uniqueStrings([...selectedCodes, code]) : selectedCodes.filter((item) => item !== code);
    setForm({ ...form, [field.key]: next });
  }

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-3 text-sm font-semibold md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span>{label}{field.required ? " *" : ""}</span>
        <Badge variant="outline">{selectedCodes.length.toLocaleString(localeOf(language))}</Badge>
      </div>
      {!selectedUserCode ? (
        <p className="rounded-xl border border-amber-300 bg-amber-50 px-2 py-1 text-xs text-amber-800 dark:border-amber-900 dark:bg-amber-950/30 dark:text-amber-200">
          {backendText(dictionary, "select_user_first", "Select user first")}
        </p>
      ) : null}
      {error ? <p className="rounded-xl border border-destructive/40 bg-destructive/10 px-2 py-1 text-xs text-destructive">{error}</p> : null}
      <div className="grid max-h-72 gap-2 overflow-y-auto pr-1 md:grid-cols-2">
        {loading ? (
          <div className="flex min-h-20 items-center gap-2 rounded-xl border border-border bg-card p-3 text-muted-foreground md:col-span-2">
            <Loader2 className="animate-spin" />
            {backendText(dictionary, "loading", "Loading data")}
          </div>
        ) : options.length ? options.map((option) => {
          const checked = selectedCodes.includes(option.code);
          return (
            <label
              className={cn(
                "grid cursor-pointer gap-1 rounded-xl border p-2 transition-colors",
                checked ? "border-primary bg-primary/10 text-primary" : "border-border bg-card text-foreground hover:bg-muted/60",
                !option.isActive && "opacity-60",
              )}
              key={option.code}
            >
              <span className="flex min-w-0 items-center gap-2">
                <input
                  className="size-4 shrink-0 accent-primary"
                  type="checkbox"
                  checked={checked}
                  disabled={!selectedUserCode}
                  onChange={(event) => toggle(option.code, event.target.checked)}
                />
                <span className="min-w-0 truncate font-semibold">{option.name || option.code}</span>
              </span>
              <span className="truncate text-xs text-muted-foreground">code: {option.code}</span>
              {option.description ? <span className="line-clamp-2 text-xs font-normal text-muted-foreground">{option.description}</span> : null}
            </label>
          );
        }) : (
          <div className="rounded-xl border border-border bg-card p-3 text-sm text-muted-foreground md:col-span-2">
            {backendText(dictionary, "empty_data", "No data")}
          </div>
        )}
      </div>
    </section>
  );
}

function PermissionMatrixEditor({
  dateTimeScope,
  dictionary,
  form,
  language,
  setForm,
  text,
}: {
  dateTimeScope: DateTimeScope;
  dictionary: BackendLanguageDictionary;
  form: FormState;
  language: LanguageCode;
  setForm: (form: FormState) => void;
  text: (key: keyof typeof uiEn) => string;
}) {
  const branchKey = dateTimeScope.key || "company";
  const branches = permissionBranchesFromForm(form.branches);
  const branchPermission = permissionBranchValue(branches, branchKey);
  const menus = isRecord(branchPermission.menus) ? branchPermission.menus : {};

  function updateMenuPermission(menuId: string, action: MenuPermissionAction, checked: boolean) {
    const nextBranches = permissionBranchesFromForm(form.branches);
    const nextBranch = permissionBranchValue(nextBranches, branchKey);
    const nextMenus = isRecord(nextBranch.menus) ? { ...nextBranch.menus } : {};
    const currentMenu = isRecord(nextMenus[menuId]) ? { ...nextMenus[menuId] } : {};
    currentMenu[action] = checked;
    nextMenus[menuId] = currentMenu;
    nextBranches[branchKey] = { ...nextBranch, ...dateTimeScopePayload(dateTimeScope), menus: nextMenus };
    setForm({ ...form, branches: nextBranches });
  }

  return (
    <section className="grid gap-3 rounded-2xl border border-border bg-background p-3 md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h3 className="text-base font-semibold">{backendText(dictionary, "permission_definition", "Permission Definition")}</h3>
          <p className="text-xs text-muted-foreground">{text("branch")}: {dateTimeScope.branchcode || dateTimeScope.branchguid || branchKey}</p>
        </div>
        <Badge variant="outline">{MENU_SECTIONS.flatMap((section) => section.groups.flatMap((group) => group.items)).length} menu codes</Badge>
      </div>

      <div className="grid max-h-[52dvh] gap-3 overflow-y-auto pr-1">
        {MENU_SECTIONS.map((section) => (
          <section className="grid gap-2 rounded-2xl border border-border bg-card p-2" key={section.id}>
            <h4 className="text-sm font-semibold">{menuText(section.title, language, dictionary)}</h4>
            {section.groups.map((group) => (
              <div className="grid gap-1 rounded-xl border border-border bg-background p-2" key={group.id}>
                <p className="text-xs font-semibold text-muted-foreground">{menuText(group.title, language, dictionary)}</p>
                {group.items.map((item) => {
                  const permission: SettingRecord = isRecord(menus[item.id]) ? menus[item.id] as SettingRecord : {};
                  return (
                    <div className="grid gap-2 rounded-xl border border-border bg-card px-2 py-2 xl:grid-cols-[minmax(180px,1fr)_auto]" key={item.id}>
                      <div className="min-w-0">
                        <p className="truncate text-sm font-semibold">{menuText(item.label, language, dictionary)}</p>
                        <p className="truncate text-xs text-muted-foreground">code: {item.id}</p>
                      </div>
                      <div className="flex flex-wrap gap-1">
                        {menuPermissionActions.map((action) => (
                          <label className="flex min-h-8 items-center gap-1 rounded-full border border-border bg-background px-2 text-xs font-semibold" key={action.key}>
                            <input
                              className="size-3.5 accent-primary"
                              type="checkbox"
                              checked={Boolean(permission[action.key])}
                              onChange={(event) => updateMenuPermission(item.id, action.key, event.target.checked)}
                            />
                            <span>{permissionActionText(dictionary, language, action.textKey)}</span>
                          </label>
                        ))}
                      </div>
                    </div>
                  );
                })}
              </div>
            ))}
          </section>
        ))}

      </div>
    </section>
  );
}

function permissionBranchesFromForm(value: unknown): SettingRecord {
  if (isRecord(value)) return { ...value };
  if (typeof value !== "string" || !value.trim()) return {};
  const parsed = safeJsonParse(value, {});
  return isRecord(parsed) ? { ...parsed } : {};
}

function permissionBranchValue(branches: SettingRecord, branchKey: string): SettingRecord {
  const branch = branches[branchKey];
  return isRecord(branch) ? { ...branch } : {};
}

function permissionActionText(dictionary: BackendLanguageDictionary, language: LanguageCode, key: keyof typeof uiEn): string {
  const fallback = uiText[language]?.[key] ?? uiEn[key];
  const backendKey = uiBackendKey(key);
  const value = backendText(dictionary, backendKey, fallback);
  return value === backendKey || value === key ? fallback : value;
}

function FieldEditor({
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
  const value = form[field.key];

  if (config.slug === "permission_definition" && field.key === "branches") {
    return (
      <PermissionMatrixEditor
        dateTimeScope={dateTimeScope}
        dictionary={dictionary}
        form={form}
        language={language}
        setForm={setForm}
        text={(key) => {
          const fallback = uiText[language]?.[key] ?? uiEn[key];
          return backendText(dictionary, uiBackendKey(key), fallback);
        }}
      />
    );
  }

  if (config.slug === "approval_setting" && field.key === "approvals") {
    return <ApprovalSettingEditor dictionary={dictionary} form={form} setForm={setForm} />;
  }

  if (config.slug === "permission_link" && field.key === "employeeCode") {
    return <PermissionLinkUserSelector auth={auth} dictionary={dictionary} field={field} form={form} language={language} setForm={setForm} workspace={workspace} />;
  }

  if (config.slug === "permission_link" && field.key === "employeeName") {
    return (
      <label className="grid gap-1 text-sm font-semibold">
        <span>{label}</span>
        <Input value={String(value ?? "")} readOnly disabled aria-readonly />
      </label>
    );
  }

  if (config.slug === "permission_link" && (field.key === "permissionCodes" || field.key === "approvalCodes")) {
    return <PermissionLinkMultiSelectEditor auth={auth} dictionary={dictionary} field={field} form={form} language={language} setForm={setForm} workspace={workspace} />;
  }

  if (field.type === "checkbox") {
    return (
      <label className="flex min-h-12 items-center gap-2 rounded-2xl border border-border bg-background p-3 text-sm font-semibold">
        <input className="size-4 accent-primary" type="checkbox" checked={Boolean(value)} onChange={(event) => setForm({ ...form, [field.key]: event.target.checked })} />
        <span>{label}</span>
      </label>
    );
  }

  if (field.type === "radio") {
    const options = field.options ?? [];
    const selectedValue = radioFormValue(value, field);
    const hasUnknownValue = Boolean(selectedValue) && options.length > 0 && !options.some((option) => option.value === selectedValue);
    const unknownValueText = language === "th"
      ? `ค่าปัจจุบันไม่ตรงกับบทบาทที่ระบบรองรับ: ${selectedValue} กรุณาเลือกใหม่`
      : `Current value is not a supported role: ${selectedValue}. Please choose a valid role.`;
    return (
      <section className="grid gap-1 text-sm font-semibold">
        <span>{label}{field.required ? " *" : ""}</span>
        <div className="flex w-full flex-wrap gap-2" role="radiogroup" aria-label={label}>
          {options.map((option) => {
            const checked = selectedValue === option.value;
            return (
              <label
                className={cn(
                  "flex min-h-10 flex-1 basis-24 cursor-pointer items-center gap-2 rounded-2xl border px-3 py-2 transition-colors",
                  checked ? "border-primary bg-primary/10 text-primary" : "border-border bg-background text-foreground hover:bg-muted/60",
                )}
                key={option.value}
              >
                <input
                  className="size-4 accent-primary"
                  name={field.key}
                  type="radio"
                  value={option.value}
                  checked={checked}
                  onChange={() => setForm({ ...form, [field.key]: radioValueToFormValue(option.value, field) })}
                />
                <span>{optionLabel(option, language)}</span>
              </label>
            );
          })}
        </div>
        {hasUnknownValue ? <span className="text-xs font-medium text-destructive">{unknownValueText}</span> : null}
      </section>
    );
  }

  if (field.type === "names") {
    const names = isRecord(value) ? value : {};
    const editorLanguages = nameEditorLanguageCodes(form, config, language);
    return (
      <section className="grid gap-1 rounded-2xl border border-border bg-background p-2 md:col-span-2">
        <div className="text-sm font-semibold">{label}{field.required ? " *" : ""}</div>
        <div className={cn("grid gap-2", editorLanguages.length > 1 && "md:grid-cols-2")}>
          {editorLanguages.map((code, index) => {
            const currentValue = typeof names[code] === "string" ? String(names[code]) : "";
            return (
              <label className="grid gap-1 text-sm font-semibold" key={code}>
                <span className="flex min-w-0 items-center gap-2 text-xs text-muted-foreground">
                  <LanguageFlag code={code} />
                  <span className="truncate">{index === 0 ? (language === "th" ? "ภาษาแรก" : "Primary") : languageName(code, language)}</span>
                  <span className="uppercase">{code}</span>
                </span>
                {field.multiline ? (
                  <textarea
                    className="min-h-20 w-full rounded-2xl border border-input bg-background px-3 py-2 text-sm text-foreground shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
                    value={currentValue}
                    onChange={(event) => setForm({ ...form, [field.key]: { ...names, [code]: event.target.value } })}
                  />
                ) : (
                  <Input
                    value={currentValue}
                    onChange={(event) => setForm({ ...form, [field.key]: { ...names, [code]: event.target.value } })}
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

  if (field.type === "textarea" || field.type === "json") {
    return (
      <label className="grid gap-1 text-sm font-semibold md:col-span-2">
        <span>{label}{field.required ? " *" : ""}</span>
        <textarea
          className="min-h-24 w-full rounded-2xl border border-input bg-background px-3 py-2 text-sm text-foreground shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
          value={String(value ?? "")}
          onChange={(event) => setForm({ ...form, [field.key]: event.target.value })}
          placeholder={field.type === "json" ? "[]" : field.placeholder}
        />
      </label>
    );
  }

  if (field.type === "combo") {
    return <ComboFieldEditor field={field} form={form} label={label} language={language} setForm={setForm} />;
  }

  if (field.type === "master-picker") {
    return <MasterPickerFieldEditor auth={auth} field={field} form={form} label={label} language={language} setForm={setForm} />;
  }

  if (field.type === "image-upload") {
    return <ImageUploadFieldEditor auth={auth} field={field} form={form} label={label} language={language} setForm={setForm} />;
  }

  if (field.type === "select") {
    const onSelectChange = (nextValue: string) => {
      const nextForm = { ...form, [field.key]: optionValueToFormValue(nextValue, field) };
      if (config.kind === "company" && field.key === "settings.language") {
        nextForm["settings.languageconfigs"] = setDefaultLanguageConfig(nextForm["settings.languageconfigs"], nextValue);
      }
      setForm(nextForm);
    };
    return (
      <label className="grid gap-1 text-sm font-semibold">
        <span>{label}{field.required ? " *" : ""}</span>
        <select
          className="min-h-10 w-full rounded-2xl border border-input bg-background px-3 text-sm text-foreground shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
          value={String(value ?? "")}
          onChange={(event) => onSelectChange(event.target.value)}
        >
          <option value=""></option>
          {field.options?.map((item) => <option key={item.value} value={item.value}>{optionLabel(item, language)}</option>)}
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
        timezoneLabel={dateTimeScope.timezone_label || dateTimeScope.timezone_offset}
        value={String(value ?? "")}
        onChange={(event) => setForm({ ...form, [field.key]: event.target.value })}
      />
    );
  }

  const emailLockedByLoginCode = config.slug === "user" && field.key === "email" && isEmailLike(form.username);

  return (
    <label className="grid gap-1 text-sm font-semibold">
      <span>{label}{field.required ? " *" : ""}</span>
      <Input
        type={field.type === "number" ? "number" : "text"}
        value={String(value ?? "")}
        readOnly={field.readOnly || emailLockedByLoginCode}
        disabled={field.readOnly || emailLockedByLoginCode}
        aria-readonly={field.readOnly || emailLockedByLoginCode}
        onChange={(event) => {
          if (field.readOnly || emailLockedByLoginCode) return;
          setForm({ ...form, [field.key]: event.target.value });
        }}
        placeholder={field.placeholder}
      />
    </label>
  );
}

function LanguageConfigsEditor({
  form,
  language,
  label,
  setForm,
}: {
  form: FormState;
  language: LanguageCode;
  label: string;
  setForm: (form: FormState) => void;
}) {
  const defaultCode = supportedLanguageCode(form["settings.language"], "th");
  const rows = normalizeLanguageConfigs(form["settings.languageconfigs"], defaultCode);
  const usedCodes = new Set(rows.map((row) => row.code));
  const availableLanguages = LANGUAGES.filter((item) => !usedCodes.has(item.code));
  const [languageToAdd, setLanguageToAdd] = useState<string>(availableLanguages[0]?.code ?? "");

  function commit(nextRows: LanguageConfigFormRow[], nextDefault = defaultCode) {
    const normalized = normalizeLanguageConfigs(nextRows, nextDefault);
    const primary = normalized[0]?.code ?? supportedLanguageCode(nextDefault, "th");
    setForm({
      ...form,
      "settings.language": primary,
      "settings.languageconfigs": normalized,
    });
  }

  function move(index: number, direction: -1 | 1) {
    const targetIndex = index + direction;
    if (index <= 0 || targetIndex <= 0 || targetIndex >= rows.length) return;
    const nextRows = [...rows];
    const [row] = nextRows.splice(index, 1);
    nextRows.splice(targetIndex, 0, row);
    commit(nextRows);
  }

  const textPrimary = language === "th" ? "ภาษาแรก" : "Primary language";
  const textAdd = language === "th" ? "เพิ่มภาษา" : "Add language";
  const textSetPrimary = language === "th" ? "ตั้งเป็นภาษาแรก" : "Set primary";
  const textRemove = language === "th" ? "เอาออก" : "Remove";
  const textNoMore = language === "th" ? "เพิ่มครบทุกภาษาที่รองรับแล้ว" : "All supported languages are already added.";

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="min-w-0">
          <div className="font-semibold">{label}</div>
          <div className="text-xs text-muted-foreground">{language === "th" ? "ลำดับแรกคือภาษาแรกของบริษัท" : "The first row is the company primary language."}</div>
        </div>
        {availableLanguages.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            <select
              className="min-h-9 min-w-44 rounded-xl border border-input bg-background px-2 text-sm text-foreground shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
              value={languageToAdd || availableLanguages[0]?.code || ""}
              onChange={(event) => setLanguageToAdd(event.target.value)}
              aria-label={textAdd}
            >
              {availableLanguages.map((item) => (
                <option key={item.code} value={item.code}>
                  {item.name}
                </option>
              ))}
            </select>
            <Button
              type="button"
              size="sm"
              onClick={() => {
                const code = supportedLanguageCode(languageToAdd || availableLanguages[0]?.code, "");
                if (!code) return;
                commit([...rows, languageConfigRow(code, false)]);
                setLanguageToAdd("");
              }}
            >
              <Plus />
              {textAdd}
            </Button>
          </div>
        ) : null}
      </div>
      <div className="grid gap-1">
        {rows.map((row, index) => (
          <div
            key={row.code}
            className={cn(
              "grid gap-2 rounded-xl border border-border bg-card px-2 py-1.5 sm:grid-cols-[56px_minmax(0,1fr)_auto] sm:items-center",
              index === 0 && "border-primary/40 bg-primary/5",
            )}
          >
            <div className={cn("grid size-10 place-items-center rounded-full font-semibold", index === 0 ? "bg-primary text-primary-foreground" : "bg-muted text-foreground")}>
              {index + 1}
            </div>
            <div className="flex min-w-0 items-center gap-2">
              <LanguageFlag code={row.code} />
              <div className="min-w-0">
                <div className="truncate font-semibold">{languageName(row.code, language)}</div>
                <div className="flex flex-wrap items-center gap-2 text-xs uppercase text-muted-foreground">
                  <span>{row.code}</span>
                  {index === 0 ? <Badge variant="outline">{textPrimary}</Badge> : null}
                </div>
              </div>
            </div>
            <div className="flex flex-wrap justify-end gap-1">
              {index > 0 ? (
                <>
                  <Button type="button" variant="outline" size="sm" onClick={() => commit(rows, row.code)} title={textSetPrimary}>
                    <Edit3 />
                    <span className="sr-only">{textSetPrimary}</span>
                  </Button>
                  <Button type="button" variant="outline" size="sm" onClick={() => move(index, -1)} disabled={index <= 1} title={language === "th" ? "เลื่อนขึ้น" : "Move up"}>
                    ↑
                  </Button>
                  <Button type="button" variant="outline" size="sm" onClick={() => move(index, 1)} disabled={index >= rows.length - 1} title={language === "th" ? "เลื่อนลง" : "Move down"}>
                    ↓
                  </Button>
                  <Button type="button" variant="ghost" size="sm" onClick={() => commit(rows.filter((item) => item.code !== row.code))} title={textRemove}>
                    <Trash2 />
                    <span className="sr-only">{textRemove}</span>
                  </Button>
                </>
              ) : null}
            </div>
          </div>
        ))}
      </div>
      {availableLanguages.length === 0 ? <p className="text-xs text-muted-foreground">{textNoMore}</p> : null}
    </section>
  );
}

function LanguageListEditor({
  field,
  form,
  language,
  label,
  setForm,
}: {
  field: SystemSettingField;
  form: FormState;
  language: LanguageCode;
  label: string;
  setForm: (form: FormState) => void;
}) {
  const rows = normalizeLanguageList(form[field.key], form.language);
  const usedCodes = new Set(rows);
  const availableLanguages = LANGUAGES.filter((item) => !usedCodes.has(item.code));
  const [languageToAdd, setLanguageToAdd] = useState<string>(availableLanguages[0]?.code ?? "");
  const textPrimary = language === "th" ? "ภาษาแรก" : "Primary";
  const textAdd = language === "th" ? "เพิ่มภาษา" : "Add language";
  const textRemove = language === "th" ? "เอาออก" : "Remove";
  const textNoMore = language === "th" ? "เพิ่มครบทุกภาษาที่รองรับแล้ว" : "All supported languages are already added.";

  function commit(nextCodes: string[]) {
    const normalized = normalizeLanguageList(nextCodes, nextCodes[0] ?? form.language);
    setForm({
      ...form,
      [field.key]: normalized,
      language: normalized[0] ?? "th",
    });
  }

  function move(index: number, direction: -1 | 1) {
    const targetIndex = index + direction;
    if (targetIndex < 0 || targetIndex >= rows.length) return;
    const nextRows = [...rows];
    const [row] = nextRows.splice(index, 1);
    nextRows.splice(targetIndex, 0, row);
    commit(nextRows);
  }

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="min-w-0">
          <div className="font-semibold">{label}</div>
          <div className="text-xs text-muted-foreground">{language === "th" ? "ลำดับแรกคือภาษาแรกของข้อมูลนี้" : "The first row is the primary language for this record."}</div>
        </div>
        {availableLanguages.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            <select
              className="min-h-9 min-w-44 rounded-xl border border-input bg-background px-2 text-sm text-foreground shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
              value={languageToAdd || availableLanguages[0]?.code || ""}
              onChange={(event) => setLanguageToAdd(event.target.value)}
              aria-label={textAdd}
            >
              {availableLanguages.map((item) => (
                <option key={item.code} value={item.code}>
                  {item.name}
                </option>
              ))}
            </select>
            <Button
              type="button"
              size="sm"
              onClick={() => {
                const code = supportedLanguageCode(languageToAdd || availableLanguages[0]?.code, "");
                if (!code) return;
                commit([...rows, code]);
                setLanguageToAdd("");
              }}
            >
              <Plus />
              {textAdd}
            </Button>
          </div>
        ) : null}
      </div>
      <div className="grid gap-1">
        {rows.map((code, index) => (
          <div
            key={code}
            className={cn(
              "grid gap-2 rounded-xl border border-border bg-card px-2 py-1.5 sm:grid-cols-[56px_minmax(0,1fr)_auto] sm:items-center",
              index === 0 && "border-primary/40 bg-primary/5",
            )}
          >
            <div className={cn("grid size-10 place-items-center rounded-full font-semibold", index === 0 ? "bg-primary text-primary-foreground" : "bg-muted text-foreground")}>
              {index + 1}
            </div>
            <div className="flex min-w-0 items-center gap-2">
              <LanguageFlag code={code} />
              <div className="min-w-0">
                <div className="truncate font-semibold">{languageName(code, language)}</div>
                <div className="flex flex-wrap items-center gap-2 text-xs uppercase text-muted-foreground">
                  <span>{code}</span>
                  {index === 0 ? <Badge variant="outline">{textPrimary}</Badge> : null}
                </div>
              </div>
            </div>
            <div className="flex flex-wrap justify-end gap-1">
              <Button type="button" variant="outline" size="sm" onClick={() => move(index, -1)} disabled={index === 0} title={language === "th" ? "เลื่อนขึ้น" : "Move up"}>
                ↑
              </Button>
              <Button type="button" variant="outline" size="sm" onClick={() => move(index, 1)} disabled={index >= rows.length - 1} title={language === "th" ? "เลื่อนลง" : "Move down"}>
                ↓
              </Button>
              {index > 0 ? (
                <Button type="button" variant="ghost" size="sm" onClick={() => commit(rows.filter((item) => item !== code))} title={textRemove}>
                  <Trash2 />
                  <span className="sr-only">{textRemove}</span>
                </Button>
              ) : null}
            </div>
          </div>
        ))}
      </div>
      {availableLanguages.length === 0 ? <p className="text-xs text-muted-foreground">{textNoMore}</p> : null}
    </section>
  );
}

function MasterPickerFieldEditor({
  auth,
  field,
  form,
  label,
  language,
  setForm,
}: {
  auth: AuthSession | null;
  field: SystemSettingField;
  form: FormState;
  label: string;
  language: LanguageCode;
  setForm: (form: FormState) => void;
}) {
  const [open, setOpen] = useState(false);
  const anchorRef = useRef<HTMLButtonElement | null>(null);
  const rawValue = form[field.key];
  const value: SettingRecord = isRecord(rawValue) ? rawValue : {};
  const names = getLocalizedNameArray(value.names);
  const displayName = localizedNameForLanguage(names, language) || stringValue(value.name);
  const code = stringValue(value.code);
  const master = (field.master ?? "businesstype") as MasterName;

  function select(entry: MasterEntry) {
    setForm({
      ...form,
      [field.key]: {
        guidfixed: entry.guidfixed,
        code: entry.code,
        names: entry.names,
      },
    });
    setOpen(false);
  }

  return (
    <section className="grid gap-1 text-sm font-semibold md:col-span-2">
      <span>{label}{field.required ? " *" : ""}</span>
      <div className="flex min-w-0 gap-2">
        <button
          ref={anchorRef}
          type="button"
          className="flex min-h-10 min-w-0 flex-1 items-center justify-between gap-2 rounded-2xl border border-input bg-background px-3 text-left text-sm text-foreground shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
          onClick={() => setOpen(true)}
        >
          <span className={cn("min-w-0 truncate", !code && !displayName && "text-muted-foreground")}>
            {code || displayName ? `${code}${code && displayName ? " - " : ""}${displayName}` : (language === "th" ? "เลือกข้อมูล" : "Select")}
          </span>
          <Search className="size-4 shrink-0 text-muted-foreground" />
        </button>
        {code || displayName ? (
          <Button type="button" variant="outline" size="icon" onClick={() => setForm({ ...form, [field.key]: {} })} aria-label={language === "th" ? "ล้างค่า" : "Clear"}>
            <X />
          </Button>
        ) : null}
      </div>
      <MasterPicker
        open={open}
        onClose={() => setOpen(false)}
        auth={auth}
        language={language}
        master={master}
        placement="field"
        anchorRef={anchorRef}
        title={label}
        onSelect={select}
      />
    </section>
  );
}

function defaultLanguageConfigs(defaultCode: unknown): LanguageConfigFormRow[] {
  return [languageConfigRow(supportedLanguageCode(defaultCode, "th"), true)];
}

function setDefaultLanguageConfig(value: unknown, defaultCode: unknown): LanguageConfigFormRow[] {
  return normalizeLanguageConfigs(value, defaultCode);
}

function normalizeLanguageConfigs(value: unknown, defaultCode: unknown): LanguageConfigFormRow[] {
  const requestedPrimaryCode = supportedLanguageCode(defaultCode, "");
  const source = typeof value === "string" ? safeJsonParse(value, []) : value;
  const input = Array.isArray(source) ? source : [];
  const seen = new Set<string>();
  const rows: LanguageConfigFormRow[] = [];

  for (const item of input) {
    if (!isRecord(item)) continue;
    const code = supportedLanguageCode(item.code, "");
    if (!code || seen.has(code)) continue;
    const isDefault = booleanLikeValue(item.isdefault) || code === requestedPrimaryCode;
    const isUse = item.is_use === undefined && item.isuse === undefined ? true : booleanLikeValue(item.is_use ?? item.isuse);
    if (!isUse && !isDefault) continue;
    seen.add(code);
    rows.push({
      code,
      codetranslator: stringValue(item.codetranslator ?? item.codeTranslator) || code,
      name: stringValue(item.name) || languageName(code, "en"),
      is_use: true,
      isdefault: isDefault,
    });
  }

  const primaryCode = supportedLanguageCode(requestedPrimaryCode || rows.find((row) => row.isdefault)?.code || "th", "th");
  if (!seen.has(primaryCode)) rows.unshift(languageConfigRow(primaryCode, true));
  const defaultRow = rows.find((row) => row.code === primaryCode) ?? rows.find((row) => row.isdefault) ?? rows[0] ?? languageConfigRow("th", true);
  const ordered = [
    { ...defaultRow, is_use: true, isdefault: true },
    ...rows.filter((row) => row.code !== defaultRow.code).map((row) => ({ ...row, is_use: true, isdefault: false })),
  ];
  return ordered.map((row) => ({
    ...row,
    codetranslator: row.codetranslator || row.code,
    name: row.name || languageName(row.code, "en"),
  }));
}

function languageConfigRow(code: LanguageCode, isDefault: boolean): LanguageConfigFormRow {
  return {
    code,
    codetranslator: code,
    name: languageName(code, "en"),
    is_use: true,
    isdefault: isDefault,
  };
}

function supportedLanguageCode(value: unknown, fallback: LanguageCode): LanguageCode;
function supportedLanguageCode(value: unknown, fallback: ""): LanguageCode | "";
function supportedLanguageCode(value: unknown, fallback: LanguageCode | ""): LanguageCode | "" {
  const raw = stringValue(value).toLowerCase();
  if (!raw) return fallback;
  const normalized = normalizeLanguage(raw);
  return LANGUAGES.some((item) => item.code === normalized) ? normalized : fallback;
}

function languageName(code: string, language: LanguageCode): string {
  const item = LANGUAGES.find((entry) => entry.code === code);
  if (!item) return code.toUpperCase();
  if (language === "th") return item.name;
  return `${item.name} (${item.code.toUpperCase()})`;
}

function LanguageFlag({ code }: { code: string }) {
  const normalized = supportedLanguageCode(code, "th");
  return (
    <span className="grid size-7 shrink-0 place-items-center overflow-hidden rounded border border-border bg-card">
      <Image alt="" height={18} src={`/flags/${normalized}.png`} width={27} />
    </span>
  );
}

function normalizeLanguageList(value: unknown, defaultCode: unknown): string[] {
  const requestedPrimaryCode = supportedLanguageCode(defaultCode, "");
  const source = typeof value === "string"
    ? (value.trim().startsWith("[") ? safeJsonParse(value, []) : value.split(","))
    : value;
  const input = Array.isArray(source) ? source : [];
  const seen = new Set<string>();
  const rows: string[] = [];

  for (const item of input) {
    const code = supportedLanguageCode(item, "");
    if (!code || seen.has(code)) continue;
    seen.add(code);
    rows.push(code);
  }

  const primaryCode = requestedPrimaryCode || rows[0] || "th";
  const ordered = [primaryCode, ...rows.filter((code) => code !== primaryCode)];
  return ordered.length ? ordered : ["th"];
}

function nameEditorLanguageCodes(form: FormState, config: SystemSettingConfig, language: LanguageCode): string[] {
  if (config.kind === "company") {
    return normalizeLanguageConfigs(form["settings.languageconfigs"], form["settings.language"]).map((row) => row.code);
  }
  if (config.slug === "branch") {
    return normalizeLanguageList(form.languages, form.language);
  }
  return [language];
}

function getLocalizedNameArray(value: unknown): Array<{ code?: string; name?: string }> {
  if (!Array.isArray(value)) return [];
  return value.filter(isRecord).map((item) => ({
    code: stringValue(item.code),
    name: stringValue(item.name),
  }));
}

function localizedNameForLanguage(names: Array<{ code?: string; name?: string }>, language: LanguageCode): string {
  return names.find((item) => item.code === language && item.name)?.name
    ?? names.find((item) => item.code === "th" && item.name)?.name
    ?? names.find((item) => item.name)?.name
    ?? "";
}

function defaultPaymentRoundingJson(): string {
  const defaultRules = [
    { lowerbound: 0.01, upperbound: 0.12, roundto: 0 },
    { lowerbound: 0.13, upperbound: 0.37, roundto: 0.25 },
    { lowerbound: 0.38, upperbound: 0.62, roundto: 0.5 },
    { lowerbound: 0.63, upperbound: 0.87, roundto: 0.75 },
    { lowerbound: 0.88, upperbound: 0.99, roundto: 1 },
  ];
  const method = { enabled: true, rules: defaultRules };
  return JSON.stringify({
    banktransfer: method,
    cash: method,
    cheque: method,
    coupon: method,
    creditcard: method,
    delivery: method,
    qrcode: method,
  }, null, 2);
}

type ComboOption = SystemSettingOption;
type ComboPlacement = {
  left: number;
  top: number;
  width: number;
  maxHeight: number;
};

const COMBO_VIEWPORT_MARGIN = 12;
const COMBO_GAP = 4;
const COMBO_MAX_HEIGHT = 360;
const COMBO_MIN_USABLE_HEIGHT = 80;
type UploadTextKey = "applyCrop" | "cancel" | "chooseImage" | "cropImage" | "cropTitle" | "imageTooLarge" | "imageUploadFailed" | "imageUploadHint" | "removeImage" | "uploading";

const uploadText: Record<UploadTextKey, Partial<Record<LanguageCode, string>> & { en: string; th: string }> = {
  applyCrop: {
    th: "ใช้รูปนี้",
    en: "Apply crop",
  },
  cancel: {
    th: "ยกเลิก",
    en: "Cancel",
  },
  chooseImage: {
    th: "เลือกรูป",
    en: "Choose image",
    cn: "选择图片",
    ja: "画像を選択",
    ko: "이미지 선택",
    lo: "ເລືອກຮູບ",
    my: "ပုံရွေးပါ",
    km: "ជ្រើសរូបភាព",
    vi: "Chọn hình ảnh",
    ms: "Pilih imej",
    id: "Pilih gambar",
    fil: "Pumili ng larawan",
  },
  cropImage: {
    th: "แก้ไขรูป",
    en: "Edit image",
  },
  cropTitle: {
    th: "ครอปรูป",
    en: "Crop image",
  },
  imageTooLarge: {
    th: "ไฟล์รูปต้องไม่เกิน 12 MB",
    en: "Image must not exceed 12 MB",
    cn: "图片不能超过 12 MB",
    ja: "画像は 12 MB 以下にしてください",
    ko: "이미지는 12 MB를 초과할 수 없습니다",
    lo: "ຮູບຕ້ອງບໍ່ເກີນ 12 MB",
    my: "ပုံဖိုင်သည် 12 MB ထက်မကြီးရပါ",
    km: "រូបភាពមិនត្រូវលើស 12 MB",
    vi: "Ảnh không được vượt quá 12 MB",
    ms: "Imej tidak boleh melebihi 12 MB",
    id: "Gambar tidak boleh lebih dari 12 MB",
    fil: "Hindi dapat lumampas sa 12 MB ang larawan",
  },
  imageUploadFailed: {
    th: "อัปโหลดรูปไม่สำเร็จ",
    en: "Image upload failed",
    cn: "图片上传失败",
    ja: "画像のアップロードに失敗しました",
    ko: "이미지 업로드 실패",
    lo: "ອັບໂຫຼດຮູບບໍ່ສຳເລັດ",
    my: "ပုံတင်ခြင်း မအောင်မြင်ပါ",
    km: "អាប់ឡូតរូបភាពមិនបានសម្រេច",
    vi: "Tải ảnh lên thất bại",
    ms: "Muat naik imej gagal",
    id: "Unggah gambar gagal",
    fil: "Nabigo ang pag-upload ng larawan",
  },
  imageUploadHint: {
    th: "รองรับ JPG, PNG, WebP และย่อรูปก่อนอัปโหลด",
    en: "JPG, PNG, WebP. The image is resized before upload.",
    cn: "支持 JPG、PNG、WebP，上传前会缩小图片",
    ja: "JPG、PNG、WebP 対応。アップロード前にリサイズします。",
    ko: "JPG, PNG, WebP 지원. 업로드 전에 크기를 줄입니다.",
    lo: "ຮອງຮັບ JPG, PNG, WebP ແລະຫຍໍ້ຮູບກ່ອນອັບໂຫຼດ",
    my: "JPG, PNG, WebP. မတင်မီ အရွယ်အစားလျှော့ပါမည်",
    km: "គាំទ្រ JPG, PNG, WebP ហើយបង្រួមរូបភាពមុនអាប់ឡូត",
    vi: "Hỗ trợ JPG, PNG, WebP và tự thu nhỏ trước khi tải lên",
    ms: "Sokong JPG, PNG, WebP dan saiz imej dikecilkan sebelum muat naik",
    id: "Mendukung JPG, PNG, WebP dan gambar diperkecil sebelum diunggah",
    fil: "Suportado ang JPG, PNG, WebP at nire-resize bago i-upload",
  },
  removeImage: {
    th: "ลบรูป",
    en: "Remove image",
    cn: "移除图片",
    ja: "画像を削除",
    ko: "이미지 제거",
    lo: "ລຶບຮູບ",
    my: "ပုံဖယ်ရှားပါ",
    km: "លុបរូបភាព",
    vi: "Xóa hình ảnh",
    ms: "Buang imej",
    id: "Hapus gambar",
    fil: "Alisin ang larawan",
  },
  uploading: {
    th: "กำลังอัปโหลด",
    en: "Uploading",
    cn: "正在上传",
    ja: "アップロード中",
    ko: "업로드 중",
    lo: "ກຳລັງອັບໂຫຼດ",
    my: "တင်နေသည်",
    km: "កំពុងអាប់ឡូត",
    vi: "Đang tải lên",
    ms: "Sedang memuat naik",
    id: "Mengunggah",
    fil: "Ina-upload",
  },
};

function ImageUploadFieldEditor({
  auth,
  field,
  form,
  label,
  language,
  setForm,
}: {
  auth: AuthSession | null;
  field: SystemSettingField;
  form: FormState;
  label: string;
  language: LanguageCode;
  setForm: (form: FormState) => void;
}) {
  const inputRef = useRef<HTMLInputElement>(null);
  const [error, setError] = useState("");
  const [cropSource, setCropSource] = useState("");
  const [localPreview, setLocalPreview] = useState("");
  const [uploading, setUploading] = useState(false);
  const value = stringValue(form[field.key]);
  const previewValue = localPreview || imageDisplayUrl(value, auth);
  const previewStyle = previewValue ? { backgroundImage: `url(${JSON.stringify(previewValue)})` } : undefined;

  useEffect(() => {
    if (!localPreview) return;
    return () => URL.revokeObjectURL(localPreview);
  }, [localPreview]);

  function clearImage() {
    setLocalPreview("");
    setCropSource("");
    setForm({ ...form, [field.key]: "" });
  }

  async function uploadImageFile(file: File) {
    if (!auth) {
      setError(uploadUiText(language, "imageUploadFailed"));
      return;
    }

    setUploading(true);
    setError("");
    try {
      const resizedFile = await resizeLogoFile(file);
      const uploadForm = new FormData();
      uploadForm.append("file", resizedFile, resizedFile.name);
      const response = await fetch("/api/upload/image", {
        method: "POST",
        headers: {
          "x-bc-backend-url": auth.backendUrl,
          Authorization: `Bearer ${auth.token}`,
        },
        body: uploadForm,
      });
      const payload = await response.json() as unknown;
      if (!response.ok || isFailed(payload)) throw new Error(extractMessage(payload) ?? uploadUiText(language, "imageUploadFailed"));
      const uri = extractUploadUri(payload);
      if (!uri) throw new Error(uploadUiText(language, "imageUploadFailed"));
      setForm({ ...form, [field.key]: uri });
    } catch (uploadError) {
      setError(uploadError instanceof Error && uploadError.message ? uploadError.message : uploadUiText(language, "imageUploadFailed"));
    } finally {
      setUploading(false);
      if (inputRef.current) inputRef.current.value = "";
    }
  }

  async function applyCroppedFile(file: File) {
    const objectUrl = URL.createObjectURL(file);
    setLocalPreview(objectUrl);
    setCropSource("");
    await uploadImageFile(file);
  }

  async function handleFile(file: File | undefined) {
    if (!file || uploading) return;
    if (!file.type.startsWith("image/") || file.size > 12 * 1024 * 1024) {
      setError(uploadUiText(language, "imageTooLarge"));
      return;
    }
    setLocalPreview(URL.createObjectURL(file));
    await uploadImageFile(file);
  }

  return (
    <section className="grid gap-1 rounded-2xl border border-border bg-background p-2 text-sm font-semibold md:col-span-2">
      <span>{label}{field.required ? " *" : ""}</span>
      <div className="flex w-full flex-wrap items-center gap-3">
        <button
          aria-label={label}
          className="grid size-20 place-items-center overflow-hidden rounded-2xl border border-input bg-card bg-contain bg-center bg-no-repeat text-muted-foreground shadow-sm"
          onClick={() => inputRef.current?.click()}
          style={previewStyle}
          type="button"
        >
          {previewValue ? <span className="sr-only">{label}</span> : <ImageIcon className="size-8" />}
        </button>
        <div className="grid min-w-48 flex-1 gap-2">
          <div className="flex w-full flex-wrap gap-2">
            <Button type="button" variant="outline" onClick={() => inputRef.current?.click()} disabled={uploading || !auth}>
              {uploading ? <Loader2 className="animate-spin" /> : <UploadCloud />}
              {uploading ? uploadUiText(language, "uploading") : uploadUiText(language, "chooseImage")}
            </Button>
            {previewValue ? (
              <Button type="button" variant="outline" onClick={() => setCropSource(previewValue)} disabled={uploading}>
                <Edit3 />
                {uploadUiText(language, "cropImage")}
              </Button>
            ) : null}
            {previewValue ? (
              <Button type="button" variant="outline" onClick={clearImage} disabled={uploading}>
                <Trash2 />
                {uploadUiText(language, "removeImage")}
              </Button>
            ) : null}
          </div>
          <p className="text-xs font-normal text-muted-foreground">{uploadUiText(language, "imageUploadHint")}</p>
          {error ? <p className="text-xs font-semibold text-destructive">{error}</p> : null}
        </div>
      </div>
      <input
        ref={inputRef}
        className="sr-only"
        type="file"
        accept="image/png,image/jpeg,image/webp,image/gif"
        onChange={(event) => void handleFile(event.target.files?.[0])}
      />
      {cropSource ? (
        <ImageCropDialog
          imageUrl={cropSource}
          language={language}
          onCancel={() => setCropSource("")}
          onApply={(file) => void applyCroppedFile(file)}
        />
      ) : null}
    </section>
  );
}

function ComboFieldEditor({
  field,
  form,
  label,
  language,
  setForm,
}: {
  field: SystemSettingField;
  form: FormState;
  label: string;
  language: LanguageCode;
  setForm: (form: FormState) => void;
}) {
  const [open, setOpen] = useState(false);
  const buttonRef = useRef<HTMLButtonElement | null>(null);
  const panelRef = useRef<HTMLDivElement | null>(null);
  const [placement, setPlacement] = useState<ComboPlacement | null>(null);
  const [query, setQuery] = useState("");
  const value = String(form[field.key] ?? "");
  const options = useMemo(() => comboOptionsForField(field, language, value), [field, language, value]);
  const currentOption = options.find((option) => option.value === value);
  const needle = query.trim().toLowerCase();
  const visibleOptions = needle
    ? options.filter((option) => `${option.value} ${optionLabel(option, language)}`.toLowerCase().includes(needle))
    : options;

  const updatePlacement = useCallback(() => {
    if (!open || typeof window === "undefined") return;
    const rect = buttonRef.current?.getBoundingClientRect();
    if (!rect) return;

    const margin = COMBO_VIEWPORT_MARGIN;
    const viewportWidth = window.innerWidth;
    const viewportHeight = window.innerHeight;
    const availableWidth = Math.max(120, viewportWidth - margin * 2);
    const width = Math.min(Math.max(rect.width, Math.min(240, availableWidth)), availableWidth);
    const maxLeft = Math.max(margin, viewportWidth - margin - width);
    const left = Math.min(Math.max(rect.left, margin), maxLeft);
    const below = Math.max(0, viewportHeight - rect.bottom - COMBO_GAP - margin);
    const above = Math.max(0, rect.top - COMBO_GAP - margin);
    const bestSpace = Math.max(below, above);

    let top: number;
    let maxHeight: number;
    if (bestSpace < COMBO_MIN_USABLE_HEIGHT) {
      top = margin;
      maxHeight = Math.max(COMBO_MIN_USABLE_HEIGHT, viewportHeight - margin * 2);
    } else if (below >= COMBO_MIN_USABLE_HEIGHT || below >= above) {
      maxHeight = Math.min(COMBO_MAX_HEIGHT, below);
      top = rect.bottom + COMBO_GAP;
    } else {
      maxHeight = Math.min(COMBO_MAX_HEIGHT, above);
      top = Math.max(margin, rect.top - COMBO_GAP - maxHeight);
    }

    setPlacement((current) => {
      if (
        current &&
        current.left === left &&
        current.top === top &&
        current.width === width &&
        current.maxHeight === maxHeight
      ) {
        return current;
      }
      return { left, top, width, maxHeight };
    });
  }, [open]);

  useLayoutEffect(() => {
    if (!open) {
      setPlacement(null);
      return;
    }

    updatePlacement();
    let frameId = 0;
    const scheduleUpdate = () => {
      window.cancelAnimationFrame(frameId);
      frameId = window.requestAnimationFrame(updatePlacement);
    };
    const resizeObserver =
      typeof ResizeObserver !== "undefined" && buttonRef.current ? new ResizeObserver(scheduleUpdate) : null;

    window.addEventListener("resize", scheduleUpdate);
    window.addEventListener("scroll", scheduleUpdate, true);
    resizeObserver?.observe(buttonRef.current as Element);

    return () => {
      window.cancelAnimationFrame(frameId);
      window.removeEventListener("resize", scheduleUpdate);
      window.removeEventListener("scroll", scheduleUpdate, true);
      resizeObserver?.disconnect();
    };
  }, [open, updatePlacement]);

  useEffect(() => {
    if (!open) return;
    const handler = (event: KeyboardEvent) => {
      if (event.key === "Escape") setOpen(false);
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [open]);

  useEffect(() => {
    if (!open) return;
    const handler = (event: PointerEvent) => {
      const target = event.target;
      if (!(target instanceof Node)) return;
      if (panelRef.current?.contains(target)) return;
      if (buttonRef.current?.contains(target)) return;
      setOpen(false);
    };
    window.addEventListener("pointerdown", handler, true);
    return () => window.removeEventListener("pointerdown", handler, true);
  }, [open]);

  function choose(option: ComboOption) {
    const nextForm = { ...form, [field.key]: option.value };
    if (field.optionSource === "timezones") {
      const meta = timezoneMeta(option.value);
      const prefix = field.key.includes(".") ? `${field.key.split(".").slice(0, -1).join(".")}.` : "";
      nextForm[`${prefix}timezone_label`] = meta.label;
      nextForm[`${prefix}timezone_offset`] = meta.offset;
    }
    if (field.optionSource === "countries") applyCountryDefaultsToForm(nextForm, option.value);
    setForm(nextForm);
    setQuery("");
    setOpen(false);
  }

  return (
    <div className="grid gap-1 text-sm font-semibold">
      <span>{label}{field.required ? " *" : ""}</span>
      <button
        ref={buttonRef}
        aria-expanded={open}
        className="flex min-h-10 w-full items-center justify-between gap-2 rounded-2xl border border-input bg-background px-3 text-left text-sm text-foreground shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
        onClick={() => setOpen((current) => !current)}
        type="button"
      >
        <span className={cn("min-w-0 truncate", !value && "text-muted-foreground")}>
          {currentOption ? optionLabel(currentOption, language) : value || field.placeholder || ""}
        </span>
        <ChevronsUpDown className="size-4 shrink-0 text-muted-foreground" />
      </button>
      {open ? (
        <div
          ref={panelRef}
          className="fixed z-50 grid grid-rows-[auto_minmax(0,1fr)] gap-1 overflow-hidden rounded-2xl border border-border bg-popover p-2 text-popover-foreground shadow-lg"
          style={{
            left: placement?.left ?? COMBO_VIEWPORT_MARGIN,
            top: placement?.top ?? COMBO_VIEWPORT_MARGIN,
            width: placement?.width ?? 240,
            maxHeight: placement?.maxHeight,
            visibility: placement ? "visible" : "hidden",
          }}
        >
          <Input
            autoFocus
            className="h-9 min-h-9 rounded-xl px-3 text-sm font-normal"
            value={query}
            onChange={(event) => setQuery(event.target.value)}
            placeholder={language === "th" ? "ค้นหา" : "Search"}
          />
          <div className="grid min-h-0 gap-1 overflow-y-auto">
            {visibleOptions.length ? visibleOptions.map((option) => (
              <button
                className={cn(
                  "flex min-h-8 w-full items-center justify-between gap-2 rounded-xl px-2 py-1 text-left text-sm font-medium hover:bg-accent hover:text-accent-foreground",
                  option.value === value && "bg-primary/10 text-primary",
                )}
                key={option.value}
                onClick={() => choose(option)}
                type="button"
              >
                <span className="min-w-0 truncate">{optionLabel(option, language)}</span>
                {option.value === value ? <Check className="size-4 shrink-0" /> : null}
              </button>
            )) : (
              <div className="rounded-xl px-2 py-3 text-sm text-muted-foreground">{language === "th" ? "ไม่พบข้อมูล" : "No options found"}</div>
            )}
          </div>
        </div>
      ) : null}
    </div>
  );
}

function WorkDayPanel({
  dateTimeScope,
  language,
  loading,
  onRefresh,
  onSave,
  saving,
  setWorkDays,
  text,
  workDays,
}: {
  dateTimeScope: DateTimeScope;
  language: LanguageCode;
  loading: boolean;
  onRefresh: () => void;
  onSave: () => void;
  saving: boolean;
  setWorkDays: (workDays: WorkDay[]) => void;
  text: (key: keyof typeof uiEn) => string;
  workDays: WorkDay[];
}) {
  const activeDays = workDays.filter((day) => day.isactive).length;
  const totalRanges = workDays.reduce((sum, day) => sum + (day.isactive ? (day.fullday ? 1 : Math.max(day.worktimes.length, 1)) : 0), 0);
  const invalidCount = workDays.reduce((sum, day) => sum + day.worktimes.filter((time, timeIndex) => getWorkTimeIssue(day, time, timeIndex)).length, 0);

  function updateDay(index: number, patch: Partial<WorkDay>) {
    setWorkDays(workDays.map((day, dayIndex) => dayIndex === index ? { ...day, ...patch } : day));
  }

  function updateTime(dayIndex: number, timeIndex: number, key: keyof WorkDayTime, value: string) {
    const day = workDays[dayIndex];
    const nextTimes = day.worktimes.length ? [...day.worktimes] : [createWorkTime("", "", dateTimeScope)];
    const currentTime = nextTimes[timeIndex] ?? createWorkTime("", "", dateTimeScope);
    nextTimes[timeIndex] = withWorkTimePatch(currentTime, key, value, dateTimeScope);
    updateDay(dayIndex, { worktimes: nextTimes });
  }

  function addTimeRange(dayIndex: number) {
    const day = workDays[dayIndex];
    updateDay(dayIndex, {
      fullday: false,
      isactive: true,
      worktimes: [...(day.worktimes.length ? day.worktimes : []), createWorkTime("", "", dateTimeScope)],
    });
  }

  function removeTimeRange(dayIndex: number, timeIndex: number) {
    const day = workDays[dayIndex];
    const nextTimes = day.worktimes.filter((_, index) => index !== timeIndex);
    updateDay(dayIndex, { worktimes: nextTimes.length ? nextTimes : [createWorkTime("", "", dateTimeScope)] });
  }

  function copyMondaySchedule() {
    const monday = workDays[0];
    if (!monday) return;
    setWorkDays(workDays.map((day, index) => index === 0 || !day.isactive ? day : { ...day, fullday: monday.fullday, worktimes: cloneWorkTimes(monday.worktimes) }));
  }

  function toggleFullDay(index: number, checked: boolean) {
    const day = workDays[index];
    updateDay(index, { fullday: checked, worktimes: day.worktimes.length ? day.worktimes : [createWorkTime("", "", dateTimeScope)] });
  }

  return (
    <Card className="w-full overflow-hidden">
      <CardHeader className="p-3 pb-1">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <div className="min-w-0">
            <CardTitle>{text("workTime")}</CardTitle>
            <CardDescription className="truncate">{text("updated")}: restaurant settings code workDay</CardDescription>
          </div>
          <div className="flex w-full flex-wrap gap-2 sm:w-auto">
            <Badge variant={invalidCount ? "warning" : "success"} className="min-h-8 justify-center px-3">{text("active")}: {activeDays}</Badge>
            <Badge variant="outline" className="min-h-8 justify-center px-3">{text("totalRanges")}: {totalRanges}</Badge>
            <Button className="w-full sm:w-auto" size="sm" type="button" variant="outline" onClick={copyMondaySchedule} disabled={loading || !workDays[0]}>
              <Copy />
              {text("copyMondaySchedule")}
            </Button>
            <Button className="w-full sm:w-auto" size="sm" type="button" variant="outline" onClick={onRefresh} disabled={loading}>
              {loading ? <Loader2 className="animate-spin" /> : <RefreshCcw />}
              {text("refresh")}
            </Button>
            <Button className="w-full sm:w-auto" size="sm" type="button" onClick={onSave} disabled={saving || invalidCount > 0}>
              {saving ? <Loader2 className="animate-spin" /> : <Save />}
              {text("save")}
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className="grid gap-2 p-3">
        {workDays.length ? workDays.map((day, index) => {
          const times = day.worktimes.length ? day.worktimes : [createWorkTime("", "", dateTimeScope)];
          return (
            <div className={cn("grid gap-2 rounded-2xl border border-border bg-background p-2", !day.isactive && "bg-muted/30")} key={day.code}>
              <div className="flex flex-wrap items-center justify-between gap-2">
                <label className="flex min-w-40 flex-1 items-center gap-2 text-sm font-semibold">
                  <input className="size-4 accent-primary" type="checkbox" checked={day.isactive} onChange={(event) => updateDay(index, { isactive: event.target.checked })} />
                  <span className="truncate">{dayNames[language]?.[index] ?? day.name}</span>
                  <Badge variant={day.isactive ? "success" : "outline"}>{day.isactive ? text("active") : text("off")}</Badge>
                </label>
                {day.isactive ? (
                  <div className="flex w-full flex-wrap items-center gap-2 sm:w-auto">
                    <label className="flex h-8 items-center gap-2 rounded-xl border border-border bg-card px-2 text-xs font-semibold">
                      <input className="size-4 accent-primary" type="checkbox" checked={day.fullday} onChange={(event) => toggleFullDay(index, event.target.checked)} />
                      {text("fullDay")}
                    </label>
                    {!day.fullday ? (
                      <Button size="sm" type="button" variant="outline" onClick={() => addTimeRange(index)}>
                        <Plus />
                        {text("add")} {text("range")}
                      </Button>
                    ) : null}
                  </div>
                ) : null}
              </div>
              {day.isactive && !day.fullday ? (
                <div className="grid gap-2 md:grid-cols-2 xl:grid-cols-3">
                  {times.map((time, timeIndex) => {
                    const issue = getWorkTimeIssue(day, time, timeIndex);
                    return (
                      <div className={cn("grid gap-1 rounded-xl border border-border bg-card p-2", issue && "border-amber-300 bg-amber-50/70 dark:border-amber-900 dark:bg-amber-950/30")} key={`${day.code}-${timeIndex}`}>
                        <div className="flex items-center justify-between gap-2">
                          <b className="text-xs">{text("range")} {timeIndex + 1}</b>
                          <Button
                            aria-label={`${text("removeRange")} ${timeIndex + 1}`}
                            disabled={times.length <= 1}
                            onClick={() => removeTimeRange(index, timeIndex)}
                            size="icon"
                            type="button"
                            variant="ghost"
                            className="h-7 w-7 rounded-xl"
                          >
                            <Trash2 />
                          </Button>
                        </div>
                        <div className="grid grid-cols-2 gap-2">
                          <TimeField
                            aria-label={`${dayNames[language]?.[index] ?? day.name} ${text("range")} ${timeIndex + 1} ${text("workTime")} start`}
                            label={text("startTime")}
                            timezoneLabel={dateTimeScope.timezone_offset}
                            utcPreview={formatUtcPreview(time.starttimeutc, time.startdayoffsetutc)}
                            value={workTimeStart(time)}
                            onChange={(event) => updateTime(index, timeIndex, "starttime", event.target.value)}
                          />
                          <TimeField
                            aria-label={`${dayNames[language]?.[index] ?? day.name} ${text("range")} ${timeIndex + 1} ${text("workTime")} end`}
                            label={text("endTime")}
                            timezoneLabel={dateTimeScope.timezone_offset}
                            utcPreview={formatUtcPreview(time.endtimeutc, time.enddayoffsetutc)}
                            value={workTimeEnd(time)}
                            onChange={(event) => updateTime(index, timeIndex, "endtime", event.target.value)}
                          />
                        </div>
                        {issue ? <span className="text-xs font-semibold text-amber-700 dark:text-amber-300">{text(issue)}</span> : null}
                      </div>
                    );
                  })}
                  <Button type="button" variant="outline" className="min-h-20 w-full border-dashed" onClick={() => addTimeRange(index)}>
                    <Plus />
                    {text("add")} {text("range")}
                  </Button>
                </div>
              ) : null}
              {day.isactive && day.fullday ? <div className="rounded-xl border border-dashed border-border bg-muted/30 px-3 py-2 text-sm font-semibold text-muted-foreground">{text("fullDay")}</div> : null}
            </div>
          );
        }) : <div className="rounded-2xl border border-dashed border-border p-4 text-sm text-muted-foreground">{text("empty")}</div>}
      </CardContent>
    </Card>
  );
}

function CopyUatPanel({
  config,
  copyPreview,
  language,
  loading,
  records,
  runCopy,
  saving,
  selected,
  setSelected,
  sourceEnvironment,
  setSourceEnvironment,
  targetShopId,
  text,
}: {
  config: SystemSettingConfig;
  copyPreview: unknown;
  language: LanguageCode;
  loading: boolean;
  records: SettingRecord[];
  runCopy: (action: "copy" | "preview") => void;
  saving: boolean;
  selected: string;
  setSelected: (value: string) => void;
  sourceEnvironment: "uat" | "pro";
  setSourceEnvironment: (value: "uat" | "pro") => void;
  targetShopId: string;
  text: (key: keyof typeof uiEn) => string;
}) {
  return (
    <Card>
      <CardContent className="grid gap-3 p-3">
        <div className="grid gap-2 lg:grid-cols-[minmax(180px,220px)_minmax(0,1fr)_auto_auto]">
          <label className="grid gap-1 text-sm font-semibold">
            <span>{text("sourceEnvironment")}</span>
            <select
              className="min-h-10 w-full rounded-2xl border border-input bg-background px-3 text-sm text-foreground shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
              disabled={loading || saving}
              onChange={(event) => {
                setSelected("");
                setSourceEnvironment(event.target.value === "pro" ? "pro" : "uat");
              }}
              value={sourceEnvironment}
            >
              <option value="uat">{text("sourceUat")}</option>
              <option value="pro">{text("sourcePro")}</option>
            </select>
          </label>
          <label className="grid gap-1 text-sm font-semibold">
            <span>{text("sourceShop")}</span>
            <select
              className="min-h-10 w-full rounded-2xl border border-input bg-background px-3 text-sm text-foreground shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
              disabled={loading}
              onChange={(event) => setSelected(event.target.value)}
              value={selected}
            >
              <option value="">{text("selectSourceShop")}</option>
              {records.map((record, index) => {
                const id = String(record.guid_fixed ?? record.shopid ?? record._id ?? "");
                return <option key={`${id}-${index}`} value={id}>{recordTitle(record, config, language)} ({id})</option>;
              })}
            </select>
          </label>
          <Button type="button" variant="outline" onClick={() => runCopy("preview")} disabled={!selected || saving}>
            {saving ? <Loader2 className="animate-spin" /> : <Search />}
            {text("preview")}
          </Button>
          <Button type="button" onClick={() => runCopy("copy")} disabled={!selected || saving}>
            {saving ? <Loader2 className="animate-spin" /> : <Copy />}
            {text("copyNow")}
          </Button>
        </div>
        <div className="grid gap-2 rounded-2xl border border-border bg-background p-3 text-sm">
          <b>{text("sourceEnvironment")}: {sourceEnvironment.toUpperCase()} → DEV</b>
          <b>{text("targetShop")}: {targetShopId || "-"}</b>
          <p className="text-muted-foreground">Preview ก่อน copy ทุกครั้ง เพราะ action นี้มีผลกับข้อมูล MongoDB ของ shop ปัจจุบัน</p>
        </div>
        {copyPreview ? (
          <pre className="max-h-80 overflow-auto rounded-2xl border border-border bg-muted p-3 text-xs text-muted-foreground">{JSON.stringify(copyPreview, null, 2)}</pre>
        ) : null}
      </CardContent>
    </Card>
  );
}

function StatCard({ label, value }: { label: string; value: string }) {
  return (
    <Card className="shadow-sm">
      <CardHeader className="p-3 pb-1">
        <CardDescription>{label}</CardDescription>
        <CardTitle>{value}</CardTitle>
      </CardHeader>
    </Card>
  );
}

function readAuth(): AuthSession | null {
  try {
    const raw = localStorage.getItem(workspaceStorageKeys.auth);
    if (!raw) return null;
    const auth = JSON.parse(raw) as AuthSession;
    return auth.token && auth.backendUrl ? auth : null;
  } catch {
    return null;
  }
}

function readWorkspace(): WorkspaceSession | null {
  try {
    const raw = localStorage.getItem(workspaceStorageKeys.workspace);
    if (!raw) return null;
    const workspace = JSON.parse(raw) as WorkspaceSession;
    return workspace?.shop?.shopid ? workspace : null;
  } catch {
    return null;
  }
}

function companyRecordForEdit(records: SettingRecord[], workspace: WorkspaceSession | null): SettingRecord | null {
  return records[0] ?? companyRecordFromWorkspace(workspace);
}

function companyRecordFromWorkspace(workspace: WorkspaceSession | null): SettingRecord | null {
  if (!workspace?.shop?.shopid) return null;
  const shopInfo = isRecord(workspace.shopInfo) ? { ...workspace.shopInfo } : {};
  const names = Array.isArray(shopInfo.names)
    ? shopInfo.names
    : Array.isArray(workspace.shop.names)
      ? workspace.shop.names
      : workspace.shop.name
        ? [{ code: "th", name: workspace.shop.name }]
        : [];
  return {
    ...shopInfo,
    shopid: stringValue(shopInfo.shopid) || workspace.shop.shopid,
    names,
    settings: isRecord(shopInfo.settings) ? shopInfo.settings : {},
  };
}

function getMainShopIdFromWorkspace(workspace: WorkspaceSession | null): string {
  if (!workspace?.shopInfo) return "";
  const mainShopId = workspace.shopInfo.main_shop_id ?? workspace.shopInfo.mainshopid ?? workspace.shopInfo.mainShopId;
  return typeof mainShopId === "string" ? mainShopId.trim() : "";
}

function requestHeaders(auth: AuthSession): HeadersInit {
  return {
    "Content-Type": "application/json",
    "x-bc-backend-url": auth.backendUrl,
    Authorization: `Bearer ${auth.token}`,
  };
}

function normalizeRecords(payload: unknown, config: SystemSettingConfig): SettingRecord[] {
  if (Array.isArray(payload)) return payload.filter(isRecord);
  if (!isRecord(payload)) return [];
  if (config.kind === "company") {
    const company = extractCompanyRecord(payload);
    return company ? [company] : [];
  }
  if (config.kind === "ai-provider" && Array.isArray(payload.providers)) return payload.providers.filter(isRecord);
  if (Array.isArray(payload.data)) return payload.data.filter(isRecord).map((record) => normalizeRestaurantRecord(record, config));
  if (isRecord(payload.data)) return [payload.data];
  return [];
}

function extractRecordTotal(payload: unknown, fallback: number): number {
  if (!isRecord(payload)) return fallback;
  const directTotal = payload.total ?? payload.count ?? payload.total_count ?? payload.totalCount;
  if (typeof directTotal === "number" && Number.isFinite(directTotal)) return directTotal;
  const pagination = payload.pagination;
  if (isRecord(pagination)) {
    const pageTotal = pagination.total ?? pagination.total_count ?? pagination.totalCount;
    if (typeof pageTotal === "number" && Number.isFinite(pageTotal)) return pageTotal;
  }
  const meta = payload.meta;
  if (isRecord(meta)) {
    const metaTotal = meta.total ?? meta.total_count ?? meta.totalCount;
    if (typeof metaTotal === "number" && Number.isFinite(metaTotal)) return metaTotal;
  }
  return fallback;
}

function mergeRecords(currentRecords: SettingRecord[], nextRecords: SettingRecord[], config: SystemSettingConfig): SettingRecord[] {
  const seen = new Set<string>();
  const merged: SettingRecord[] = [];
  for (const record of [...currentRecords, ...nextRecords]) {
    const id = recordId(record, config);
    const key = id || JSON.stringify(record);
    if (seen.has(key)) continue;
    seen.add(key);
    merged.push(record);
  }
  return merged;
}

function extractCompanyRecord(payload: Record<string, unknown>): SettingRecord | null {
  const data = payload.data;
  if (isRecord(data)) return data;
  const shop = payload.shop;
  if (isRecord(shop)) return shop;
  const result = payload.result;
  if (isRecord(result)) return result;
  if (payload.shopid || payload.names || payload.settings) return payload;
  return null;
}

function normalizeRestaurantRecord(record: SettingRecord, config: SystemSettingConfig): SettingRecord {
  if (config.kind !== "restaurant-setting") return record;
  const body = typeof record.body === "string" ? safeJsonParse(record.body, {}) : isRecord(record.body) ? record.body : {};
  return isRecord(body) ? { ...record, ...body } : record;
}

function normalizeWorkDays(records: SettingRecord[], language: LanguageCode, scope: DateTimeScope): WorkDay[] {
  const first = records[0];
  const body = scopedRestaurantBody(first, scope);
  const raw = Array.isArray(body.workdays) ? body.workdays : [];
  const names = dayNames[language] ?? dayNames.en;
  if (!raw.length) {
    return names.map((name, index) => ({
      code: String(index + 1),
      name,
      isactive: false,
      fullday: false,
      worktimes: [],
    }));
  }
  return raw.map((item, index) => {
    const record = isRecord(item) ? item : {};
    return {
      code: String(record.code ?? index + 1),
      name: String(record.name ?? names[index] ?? index + 1),
      isactive: Boolean(record.isactive ?? true),
      fullday: Boolean(record.fullday ?? false),
      worktimes: Array.isArray(record.worktimes)
        ? record.worktimes.filter(isRecord).map((time) => normalizeWorkTimeRecord(time, scope))
        : [],
    };
  });
}

function normalizeWorkTimeRecord(record: SettingRecord, scope: DateTimeScope): WorkDayTime {
  const starttime = normalizeTimeInput(record.starttime ?? record.start);
  const endtime = normalizeTimeInput(record.endtime ?? record.end);
  return applyUtcFields({ ...record, starttime, endtime, start: toLegacyTime(starttime), end: toLegacyTime(endtime) }, scope);
}

function createWorkTime(starttime = "", endtime = "", scope?: DateTimeScope): WorkDayTime {
  const start = normalizeTimeInput(starttime);
  const end = normalizeTimeInput(endtime);
  return applyUtcFields({ starttime: start, endtime: end, start: toLegacyTime(start), end: toLegacyTime(end) }, scope);
}

function withWorkTimePatch(time: WorkDayTime, key: keyof WorkDayTime, value: string, scope: DateTimeScope): WorkDayTime {
  const normalized = key === "starttime" || key === "endtime" || key === "start" || key === "end" ? normalizeTimeInput(value) : value;
  const next: WorkDayTime = { ...time, [key]: normalized };
  if (key === "starttime" || key === "start") {
    next.starttime = normalized;
    next.start = toLegacyTime(normalized);
  }
  if (key === "endtime" || key === "end") {
    next.endtime = normalized;
    next.end = toLegacyTime(normalized);
  }
  return applyUtcFields(next, scope);
}

function applyUtcFields(time: WorkDayTime, scope?: DateTimeScope): WorkDayTime {
  const next: WorkDayTime = { ...time };
  const starttime = workTimeStart(next);
  const endtime = workTimeEnd(next);
  next.starttime = starttime;
  next.endtime = endtime;
  next.start = toLegacyTime(starttime);
  next.end = toLegacyTime(endtime);
  if (scope) {
    next.timezone = scope.timezone;
    next.timezone_offset = scope.timezone_offset;
  }
  const startUtc = scope?.timezone_offset ? localTimeToUtcTime(starttime, scope.timezone_offset) : null;
  const endUtc = scope?.timezone_offset ? localTimeToUtcTime(endtime, scope.timezone_offset) : null;
  if (startUtc) {
    next.starttimeutc = startUtc.time;
    next.startdayoffsetutc = startUtc.dayOffset;
  } else {
    delete next.starttimeutc;
    delete next.startdayoffsetutc;
  }
  if (endUtc) {
    next.endtimeutc = endUtc.time;
    next.enddayoffsetutc = endUtc.dayOffset;
  } else {
    delete next.endtimeutc;
    delete next.enddayoffsetutc;
  }
  return next;
}

function cloneWorkTimes(times: WorkDayTime[]): WorkDayTime[] {
  return times.map((time) => ({ ...time }));
}

function workTimeStart(time: WorkDayTime): string {
  return normalizeTimeInput(time.starttime ?? time.start);
}

function workTimeEnd(time: WorkDayTime): string {
  return normalizeTimeInput(time.endtime ?? time.end);
}

function toLegacyTime(value: string): string {
  const normalized = normalizeTimeInput(value);
  return normalized ? normalized.replace(":", "") : "";
}

function getWorkTimeIssue(day: WorkDay, time: WorkDayTime, timeIndex: number): "timeInvalid" | "timeOverlap" | null {
  if (!day.isactive || day.fullday) return null;
  const start = timeToMinutes(workTimeStart(time));
  const end = timeToMinutes(workTimeEnd(time));
  if (!Number.isFinite(start) || !Number.isFinite(end) || start >= end) return "timeInvalid";
  const overlaps = day.worktimes.some((other, otherIndex) => {
    if (otherIndex === timeIndex) return false;
    const otherStart = timeToMinutes(workTimeStart(other));
    const otherEnd = timeToMinutes(workTimeEnd(other));
    if (!Number.isFinite(otherStart) || !Number.isFinite(otherEnd) || otherStart >= otherEnd) return false;
    return start < otherEnd && end > otherStart;
  });
  return overlaps ? "timeOverlap" : null;
}

function formatUtcPreview(time: string | undefined, dayOffset: number | undefined): string {
  if (!time) return "";
  const suffix = dayOffset ? ` ${dayOffset > 0 ? "+" : ""}${dayOffset}d` : "";
  return `UTC+0 ${time}${suffix}`;
}

function buildBranchScopedWorkDayBody(record: SettingRecord | undefined, workspace: WorkspaceSession, workDays: WorkDay[]): SettingRecord {
  const scope = resolveDateTimeScope(workspace);
  const base = restaurantBody(record);
  const branches = isRecord(base.branches) ? { ...base.branches } : {};
  const scopedWorkDays = workDays.map((day) => ({
    ...day,
    worktimes: day.worktimes.map((time) => applyUtcFields(time, scope)),
  }));
  const branchPayload = {
    ...(isRecord(branches[scope.key]) ? branches[scope.key] as SettingRecord : {}),
    ...dateTimeScopePayload(scope),
    workdays: scopedWorkDays,
  };
  branches[scope.key] = branchPayload;
  return {
    ...base,
    ...dateTimeScopePayload(scope),
    workdays: scopedWorkDays,
    branches,
  };
}

function scopedRestaurantBody(record: SettingRecord | undefined, scope: DateTimeScope): SettingRecord {
  const body = restaurantBody(record);
  const branches = isRecord(body.branches) ? body.branches : {};
  const direct = branches[scope.key];
  if (isRecord(direct)) return direct;
  const matched = Object.values(branches).find((item) => {
    if (!isRecord(item)) return false;
    return String(item.branchguid ?? "") === scope.branchguid || String(item.branchcode ?? "") === scope.branchcode;
  });
  if (isRecord(matched)) return matched;
  return Object.keys(branches).length ? {} : body;
}

function restaurantBody(record: SettingRecord | undefined): SettingRecord {
  if (!record) return {};
  if (typeof record.body === "string") {
    const parsed = safeJsonParse(record.body, {});
    return isRecord(parsed) ? parsed : {};
  }
  if (isRecord(record.body)) return record.body;
  return record;
}

function filterRecordsByDateTimeScope(records: SettingRecord[], scope: DateTimeScope): SettingRecord[] {
  return records.filter((record) => {
    const key = String(record.branch_key ?? "");
    const branchguid = String(record.branchguid ?? "");
    const branchcode = String(record.branchcode ?? "");
    if (!key && !branchguid && !branchcode) return true;
    return key === scope.key || branchguid === scope.branchguid || branchcode === scope.branchcode;
  });
}

function defaultForm(config: SystemSettingConfig, language: LanguageCode): FormState {
  const form: FormState = {};
  for (const field of config.fields) {
    if (field.type === "checkbox") form[field.key] = field.key === "isActive" || field.key === "is_active" || field.key === "isenabled";
    else if (field.type === "names") form[field.key] = Object.fromEntries(LANGUAGES.map((item) => [item.code, item.code === language ? "" : ""]));
    else if (field.type === "language-configs") form[field.key] = defaultLanguageConfigs("th");
    else if (field.type === "language-list") form[field.key] = ["th"];
    else if (field.type === "master-picker") form[field.key] = {};
    else if (field.type === "json") form[field.key] = field.key === "paymentrounding"
      ? defaultPaymentRoundingJson()
      : field.key === "permissionCodes" || field.key === "approvalCodes" || field.key === "allowed_tools" ? "[]" : "{}";
    else if (field.type === "number") form[field.key] = isDecimalSettingField(field.key) ? 2 : "";
    else if (field.type === "radio") form[field.key] = radioValueToFormValue(field.options?.[0]?.value ?? "", field);
    else form[field.key] = "";
  }
  if (config.slug === "mcp_apikey") {
    form.allowed_tools = JSON.stringify(["readonly"], null, 2);
    form.rate_limit_per_minute = "600";
    form.is_active = true;
  }
  if (config.slug === "ai_provider") {
    form.provider_name = "ollama";
    form.is_active = true;
    form.priority = "1";
  }
  if (config.slug === "approval_setting") {
    form.approvalCode = "default";
    form.approvalName = "Default";
    form.isActive = true;
    form.approvals = {};
  }
  if (config.slug === "permission_definition") {
    form.branches = {};
  }
  if (config.slug === "permission_link") {
    form.permissionCodes = [];
    form.approvalCodes = [];
  }
  applyCompanyDefaults(form, config);
  applyBranchDefaults(form, config);
  return form;
}

function formFromRecord(record: SettingRecord, config: SystemSettingConfig, language: LanguageCode): FormState {
  const form = defaultForm(config, language);
  for (const field of config.fields) {
    const value = getByPath(record, field.key);
    if (field.type === "names") form[field.key] = namesToObject(value);
    else if (field.type === "language-configs") form[field.key] = normalizeLanguageConfigs(value, getByPath(record, "settings.language"));
    else if (field.type === "language-list") form[field.key] = normalizeLanguageList(value, getByPath(record, "language"));
    else if (field.type === "master-picker") form[field.key] = isRecord(value) ? value : {};
    else if (field.type === "json" && config.slug === "permission_definition" && field.key === "branches") form[field.key] = isRecord(value) ? value : {};
    else if (field.type === "json" && config.slug === "approval_setting" && field.key === "approvals") form[field.key] = isRecord(value) ? value : {};
    else if (field.type === "json" && config.slug === "permission_link" && (field.key === "permissionCodes" || field.key === "approvalCodes")) form[field.key] = stringArrayFromForm(value);
    else if (field.type === "json") form[field.key] = JSON.stringify(value ?? (field.key === "permissionCodes" || field.key === "approvalCodes" || field.key === "allowed_tools" ? [] : {}), null, 2);
    else if (field.type === "radio") form[field.key] = radioValueToFormValue(radioFormValue(value, field), field);
    else if (field.key === "api_key") form[field.key] = "";
    else if (field.type === "number" && isDecimalSettingField(field.key)) form[field.key] = normalizeDecimalPlaces(value);
    else form[field.key] = value ?? "";
  }
  applyCompanyDefaults(form, config);
  applyBranchDefaults(form, config);
  return form;
}

function applyCompanyDefaults(form: FormState, config: SystemSettingConfig) {
  if (config.slug !== "company") return;
  for (const [key, value] of Object.entries(companySetupDefaults)) {
    setFormValueIfEmpty(form, key, value);
  }
  const timezone = stringValue(form["settings.timezone"]);
  if (timezone) applyTimezoneMetaToForm(form, "settings.timezone", timezone);
  syncCompanyLanguageForm(form);
}

function applyBranchDefaults(form: FormState, config: SystemSettingConfig) {
  if (config.slug !== "branch") return;
  for (const [key, value] of Object.entries(branchSetupDefaults)) {
    setFormValueIfEmpty(form, key, value);
  }
  form.languages = normalizeLanguageList(form.languages, form.language);
  if (!stringValue(form.language)) form.language = Array.isArray(form.languages) ? form.languages[0] ?? "th" : "th";
  const timezone = stringValue(form.timezone);
  if (timezone) applyTimezoneMetaToForm(form, "timezone", timezone);
}

function applyCountryDefaultsToForm(form: FormState, countryCode: string) {
  if (countryCode !== "TH") return;
  for (const [key, value] of Object.entries(companySetupDefaults)) {
    if (key === "settings.country_code") continue;
    setFormValueIfEmpty(form, key, value);
  }
  const timezone = stringValue(form["settings.timezone"]);
  if (timezone) applyTimezoneMetaToForm(form, "settings.timezone", timezone);
  syncCompanyLanguageForm(form);
}

function syncCompanyLanguageForm(form: FormState) {
  const defaultCode = supportedLanguageCode(form["settings.language"], "th");
  const configs = normalizeLanguageConfigs(form["settings.languageconfigs"], defaultCode);
  form["settings.language"] = configs[0]?.code ?? defaultCode;
  form["settings.languageconfigs"] = configs;
}

function setFormValueIfEmpty(form: FormState, key: string, value: unknown) {
  if (isEmptyFormValue(form[key])) form[key] = value;
}

function isEmptyFormValue(value: unknown): boolean {
  return value === "" || value === null || value === undefined;
}

function applyTimezoneMetaToForm(form: FormState, key: string, timeZone: string) {
  const meta = timezoneMeta(timeZone);
  const prefix = key.includes(".") ? `${key.split(".").slice(0, -1).join(".")}.` : "";
  setFormValueIfEmpty(form, `${prefix}timezone_label`, meta.label);
  setFormValueIfEmpty(form, `${prefix}timezone_offset`, meta.offset);
}

function buildPayload(form: FormState, editing: SettingRecord | null, config: SystemSettingConfig, workspace: WorkspaceSession, auth: AuthSession, language: LanguageCode): SettingRecord {
  const payload: SettingRecord = { ...(editing ?? {}) };
  for (const field of config.fields) {
    if (field.readOnly) continue;
    const value = form[field.key];
    if (field.type === "names") {
      setByPath(payload, field.key, objectToNames(
        value,
        language,
        getByPath(payload, field.key),
        nameEditorLanguageCodes(form, config, language),
      ));
    }
    else if (field.type === "language-configs") setByPath(payload, field.key, normalizeLanguageConfigs(value, form["settings.language"]));
    else if (field.type === "language-list") setByPath(payload, field.key, normalizeLanguageList(value, form.language));
    else if (field.type === "master-picker") setByPath(payload, field.key, isRecord(value) ? value : {});
    else if (field.type === "json") setByPath(payload, field.key, parseJsonField(value, field.key));
    else if (field.type === "checkbox") setByPath(payload, field.key, Boolean(value));
    else if (field.type === "radio") setByPath(payload, field.key, radioValueToFormValue(radioFormValue(value, field), field));
    else if (field.type === "select") setByPath(payload, field.key, optionValueToFormValue(String(value ?? ""), field));
    else if (field.type === "combo") {
      const selectedValue = String(value ?? "").trim();
      setByPath(payload, field.key, selectedValue);
      if (field.optionSource === "timezones" && selectedValue) setTimezoneDerivedPayload(payload, field.key, selectedValue);
    }
    else if (field.type === "number") setByPath(
      payload,
      field.key,
      isDecimalSettingField(field.key) ? normalizeDecimalPlaces(value) : value === "" || value === null || value === undefined ? 0 : Number(value),
    );
    else if (field.key === "api_key" && !String(value ?? "").trim()) {
      deleteByPath(payload, field.key);
    } else {
      setByPath(payload, field.key, String(value ?? "").trim());
    }
  }

  if (config.kind === "atlas") {
    const now = new Date().toISOString();
    payload.shopid = workspace.shop.shopid;
    payload.updatedAt = now;
    payload.updatedBy = auth.username;
    if (!editing) {
      payload.createdAt = now;
      payload.createdBy = auth.username;
    }
  }

  if (config.kind === "goapi-crud") {
    payload.shop_id = workspace.shop.shopid;
    payload.created_by = payload.created_by ?? auth.username;
    if (!editing && config.slug === "mcp_apikey") payload.createWithExport = true;
  }

  if (config.kind === "ai-provider") {
    payload.shop_id = workspace.shop.shopid;
    if (!payload.api_key && String(payload.provider_name) !== "ollama") {
      throw new Error(language === "th" ? "กรุณากรอก API Key เมื่อบันทึก AI Provider" : "Please enter API Key when saving AI Provider.");
    }
  }

  if (config.kind === "company") {
    const configs = normalizeLanguageConfigs(getByPath(payload, "settings.languageconfigs"), form["settings.language"] ?? getByPath(payload, "settings.language"));
    setByPath(payload, "settings.languageconfigs", configs);
    setByPath(payload, "settings.language", configs[0]?.code ?? "th");
    payload.shopid = workspace.shop.shopid;
  }

  if (config.slug === "branch") {
    const languages = normalizeLanguageList(getByPath(payload, "languages"), form.language ?? getByPath(payload, "language"));
    setByPath(payload, "languages", languages);
    setByPath(payload, "language", languages[0] ?? "th");
    const timezone = stringValue(getByPath(payload, "timezone"));
    if (timezone) {
      const meta = timezoneMeta(timezone);
      setByPath(payload, "timezonelabel", stringValue(getByPath(payload, "timezonelabel")) || meta.label);
      setByPath(payload, "timezoneoffset", normalizeUtcOffset(stringValue(getByPath(payload, "timezoneoffset")) || meta.offset));
      setByPath(payload, "timezone_label", stringValue(getByPath(payload, "timezone_label")) || meta.label);
      setByPath(payload, "timezone_offset", normalizeUtcOffset(stringValue(getByPath(payload, "timezone_offset")) || meta.offset));
    }
  }

  if (config.slug === "user") {
    payload.shopid = workspace.shop.shopid;
    if (editing) payload.editusername = recordId(editing, config);
    if (isEmailLike(payload.username)) {
      payload.email = payload.username;
    }
    if (isCreatorRecord(payload, workspace)) payload.is_access_disabled = false;
  }

  if (config.slug === "department") {
    Object.assign(payload, dateTimeScopePayload(resolveDateTimeScope(workspace)));
  }

  if (config.kind === "restaurant-setting") {
    const scope = resolveDateTimeScope(workspace);
    Object.assign(payload, dateTimeScopePayload(scope));
    if (config.slug === "holiday_screen") {
      const localDate = String(payload.date ?? "").trim();
      const utcDate = localDateToUtcIso(localDate, scope.timezone_offset);
      if (utcDate) payload.date_utc = utcDate;
    }
  }

  return payload;
}

function resolveDateTimeScope(workspace: WorkspaceSession | null): DateTimeScope {
  const branch = workspace?.branch ?? null;
  const shopInfo = workspace?.shopInfo ?? null;
  const branchcode = stringValue(branch?.code ?? workspace?.shop.branchcode);
  const branchguid = stringValue(branch?.guid_fixed);
  const shopTimezone = stringValue(getByPath(shopInfo, "settings.timezone"));
  const shopTimezoneLabel = stringValue(getByPath(shopInfo, "settings.timezone_label"));
  const shopTimezoneOffset = stringValue(getByPath(shopInfo, "settings.timezone_offset"));
  const branchYear = stringValue(branch?.year_type).toLowerCase();
  const useBuddhistCalendar = booleanLikeValue(getByPath(shopInfo, "settings.usebuddhistcalendar"));
  const calendarYearType: CalendarYearType =
    branchYear === "buddhist" || branchYear === "be" || branchYear === "พ.ศ."
      ? "buddhist"
      : branchYear === "christian" || branchYear === "ce" || branchYear === "ค.ศ."
        ? "christian"
        : useBuddhistCalendar
          ? "buddhist"
          : "christian";
  return {
    key: branchguid || branchcode || "company",
    branchcode,
    branchguid,
    timezone: stringValue(branch?.timezone) || shopTimezone,
    timezone_label: stringValue(branch?.timezone_label) || shopTimezoneLabel || stringValue(branch?.timezone) || shopTimezone,
    timezone_offset: normalizeUtcOffset(stringValue(branch?.timezone_offset) || shopTimezoneOffset),
    calendarYearType,
  };
}

function dateTimeScopePayload(scope: DateTimeScope): SettingRecord {
  return {
    branch_key: scope.key,
    branchcode: scope.branchcode,
    branchguid: scope.branchguid,
    timezone: scope.timezone,
    timezone_label: scope.timezone_label,
    timezone_offset: scope.timezone_offset,
    calendar_year_type: scope.calendarYearType,
  };
}

function localDateToUtcIso(localDate: string, utcOffset: string): string {
  const [year, month, day] = localDate.split("-").map(Number);
  if (!year || !month || !day || !utcOffset) return "";
  const utc = localTimeToUtcTime("00:00", utcOffset);
  if (!utc) return "";
  const [hours, minutes] = utc.time.split(":").map(Number);
  return new Date(Date.UTC(year, month - 1, day + utc.dayOffset, hours, minutes, 0, 0)).toISOString();
}

function normalizeUtcOffset(value: string): string {
  const raw = value.trim().replace(/^UTC/i, "");
  const match = raw.match(/^([+-])(\d{1,2}):?(\d{2})$/);
  if (!match) return raw;
  return `${match[1]}${match[2].padStart(2, "0")}:${match[3]}`;
}

function stringValue(value: unknown): string {
  return typeof value === "string" ? value.trim() : value === null || value === undefined ? "" : String(value).trim();
}

function isEmailLike(value: unknown): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(stringValue(value));
}

function recordId(record: SettingRecord | null | undefined, config: SystemSettingConfig): string {
  if (!record) return "";
  const value = config.idField ? getByPath(record, config.idField) : undefined;
  return String(value ?? record.guid_fixed ?? record.id ?? record._id ?? record.code ?? record.shopid ?? record.provider_name ?? "");
}

function recordTitle(record: SettingRecord, config: SystemSettingConfig | undefined, language: LanguageCode): string {
  const names = getByPath(record, "names") ?? getByPath(record, "name") ?? getByPath(record, "desc");
  const localized = localizedValue(names, language);
  if (localized) return localized;
  return String(
    record.name ??
    record.name1 ??
    record.permissionName ??
    record.approvalName ??
    record.employeeName ??
    record.username ??
    record.provider_name ??
    record.code ??
    record.guid_fixed ??
    config?.slug ??
    "",
  );
}

function recordBranchCaption(record: SettingRecord): string {
  return stringValue(record.branchcode) || stringValue(record.branchguid) || stringValue(record.branch_key);
}

function isCreatorRecord(record: SettingRecord, workspace: WorkspaceSession | null): boolean {
  if (Boolean(record.is_creator)) return true;
  const creator = stringValue(record.createdby ?? workspace?.shop.createdby ?? getByPath(workspace?.shopInfo ?? {}, "createdby"));
  const username = stringValue(record.username ?? record.email ?? record.code);
  return Boolean(creator && username && creator.toLowerCase() === username.toLowerCase());
}

function isSelfUserRecord(record: SettingRecord, auth: AuthSession | null): boolean {
  const username = stringValue(record.username ?? record.email ?? record.code).toLowerCase();
  const identities = [auth?.username, auth?.profile?.email]
    .map((item) => stringValue(item).toLowerCase())
    .filter(Boolean);
  return Boolean(username && identities.includes(username));
}

function userAccessDisabled(record: SettingRecord): boolean {
  return Boolean(record.is_access_disabled ?? record.is_access_disabled ?? false);
}

function localizedValue(value: unknown, language: LanguageCode): string {
  if (Array.isArray(value)) {
    const found = value.find((item) => isRecord(item) && String(item.code).toLowerCase() === language && typeof item.name === "string");
    if (isRecord(found) && typeof found.name === "string") return found.name;
    const fallback = value.find((item) => isRecord(item) && typeof item.name === "string");
    return isRecord(fallback) && typeof fallback.name === "string" ? fallback.name : "";
  }
  if (typeof value === "string") return value;
  return "";
}

function uiBackendKey(key: keyof typeof uiEn): string {
  return uiBackendKeys[key] ?? key;
}

function systemSettingTitle(config: SystemSettingConfig, language: LanguageCode, dictionary: BackendLanguageDictionary): string {
  const fallback = systemSettingLabel(config, language);
  const backendKey = systemSettingBackendKeys[config.slug] ?? config.slug;
  const value = backendText(dictionary, backendKey, fallback);
  return value === backendKey || value === config.slug ? fallback : value;
}

function fieldLabel(field: SystemSettingField, language: LanguageCode, config?: SystemSettingConfig, dictionary?: BackendLanguageDictionary): string {
  const fallback = field.label[language] ?? field.label.en ?? field.label.th;
  const backendKey = config ? fieldBackendKeys[`${config.slug}.${field.key}`] ?? fieldBackendKeys[field.key] : undefined;
  if (!backendKey || !dictionary) return fallback;
  const value = backendText(dictionary, backendKey, fallback);
  return value === backendKey || value === field.key ? fallback : value;
}

function optionLabel(option: SystemSettingOption, language: LanguageCode): string {
  return option.labels?.[language] ?? option.labels?.en ?? option.labels?.th ?? option.label;
}

function isProductUnitOption(value: unknown): value is ProductUnitOption {
  return isRecord(value) && typeof value.unitcode === "string" && value.unitcode.trim().length > 0;
}

function unitDisplayName(unit: ProductUnitOption, language: LanguageCode): string {
  const names = Array.isArray(unit.names) ? unit.names : [];
  return names.find((name) => name.code?.toLowerCase() === language && name.name?.trim())?.name?.trim()
    ?? names.find((name) => name.code?.toLowerCase() === "th" && name.name?.trim())?.name?.trim()
    ?? names.find((name) => name.name?.trim())?.name?.trim()
    ?? unit.unitcode;
}

function uploadUiText(language: LanguageCode, key: UploadTextKey): string {
  return uploadText[key][language] ?? uploadText[key].en;
}

function ImageCropDialog({
  imageUrl,
  language,
  onApply,
  onCancel,
}: {
  imageUrl: string;
  language: LanguageCode;
  onApply: (file: File) => void;
  onCancel: () => void;
}) {
  const canvasRef = useRef<HTMLCanvasElement | null>(null);
  const [image, setImage] = useState<HTMLImageElement | null>(null);
  const [offsetX, setOffsetX] = useState(0);
  const [offsetY, setOffsetY] = useState(0);
  const [zoom, setZoom] = useState(1);
  const [error, setError] = useState("");

  useEffect(() => {
    let active = true;
    const nextImage = new window.Image();
    nextImage.crossOrigin = "anonymous";
    nextImage.onload = () => {
      if (!active) return;
      setImage(nextImage);
      setOffsetX(0);
      setOffsetY(0);
      setZoom(1);
      setError("");
    };
    nextImage.onerror = () => {
      if (active) setError(uploadUiText(language, "imageUploadFailed"));
    };
    nextImage.src = imageUrl;
    return () => {
      active = false;
    };
  }, [imageUrl, language]);

  useEffect(() => {
    if (!image || !canvasRef.current) return;
    drawCroppedImage(canvasRef.current, image, { offsetX, offsetY, zoom });
  }, [image, offsetX, offsetY, zoom]);

  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      if (event.key === "Escape") onCancel();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [onCancel]);

  async function apply() {
    if (!canvasRef.current) return;
    try {
      const blob = await canvasToBlob(canvasRef.current, "image/webp", 0.86);
      if (!blob) throw new Error(uploadUiText(language, "imageUploadFailed"));
      onApply(new File([blob], "cropped-image.webp", { type: "image/webp", lastModified: Date.now() }));
    } catch {
      setError(uploadUiText(language, "imageUploadFailed"));
    }
  }

  return (
    <div className="fixed inset-0 z-50 grid place-items-center bg-black/40 p-2" role="dialog" aria-modal="true">
      <section className="grid w-[min(420px,calc(100vw-16px))] gap-2 rounded-2xl border border-border bg-card p-3 text-card-foreground shadow-xl">
        <div className="flex items-center justify-between gap-2">
          <div className="text-sm font-semibold">{uploadUiText(language, "cropTitle")}</div>
          <Button type="button" variant="ghost" size="sm" onClick={onCancel} aria-label={uploadUiText(language, "cancel")}>
            <X />
          </Button>
        </div>
        <div className="grid place-items-center rounded-2xl border border-input bg-muted/40 p-2">
          <canvas ref={canvasRef} width={512} height={512} className="aspect-square w-full max-w-80 rounded-xl bg-background object-contain" />
        </div>
        <label className="grid gap-1 text-xs font-semibold">
          <span>Zoom</span>
          <input type="range" min="1" max="3" step="0.01" value={zoom} onChange={(event) => setZoom(Number(event.target.value))} />
        </label>
        <label className="grid gap-1 text-xs font-semibold">
          <span>X</span>
          <input type="range" min="-100" max="100" step="1" value={offsetX} onChange={(event) => setOffsetX(Number(event.target.value))} />
        </label>
        <label className="grid gap-1 text-xs font-semibold">
          <span>Y</span>
          <input type="range" min="-100" max="100" step="1" value={offsetY} onChange={(event) => setOffsetY(Number(event.target.value))} />
        </label>
        {error ? <p className="text-xs font-semibold text-destructive">{error}</p> : null}
        <div className="flex flex-wrap justify-end gap-2">
          <Button type="button" variant="outline" onClick={onCancel}>{uploadUiText(language, "cancel")}</Button>
          <Button type="button" onClick={() => void apply()} disabled={!image}>{uploadUiText(language, "applyCrop")}</Button>
        </div>
      </section>
    </div>
  );
}

function drawCroppedImage(
  canvas: HTMLCanvasElement,
  image: HTMLImageElement,
  crop: { offsetX: number; offsetY: number; zoom: number },
) {
  const size = Math.min(image.naturalWidth, image.naturalHeight);
  const cropSize = Math.max(1, size / Math.max(1, crop.zoom));
  const maxX = Math.max(0, image.naturalWidth - cropSize);
  const maxY = Math.max(0, image.naturalHeight - cropSize);
  const sourceX = clampNumber((image.naturalWidth - cropSize) / 2 + (crop.offsetX / 100) * (maxX / 2), 0, maxX);
  const sourceY = clampNumber((image.naturalHeight - cropSize) / 2 + (crop.offsetY / 100) * (maxY / 2), 0, maxY);
  const context = canvas.getContext("2d");
  if (!context) return;
  context.clearRect(0, 0, canvas.width, canvas.height);
  context.imageSmoothingEnabled = true;
  context.imageSmoothingQuality = "high";
  context.drawImage(image, sourceX, sourceY, cropSize, cropSize, 0, 0, canvas.width, canvas.height);
}

function clampNumber(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value));
}

function extractUploadUri(payload: unknown): string {
  if (!isRecord(payload)) return "";
  const direct = stringValue(payload.uri ?? payload.url ?? payload.file_url ?? payload.fileUrl ?? payload.imageurl ?? payload.image_uri);
  if (direct) return direct;
  const data = payload.data;
  if (!isRecord(data)) return "";
  return stringValue(data.uri ?? data.url ?? data.file_url ?? data.fileUrl ?? data.imageurl ?? data.image_uri);
}

function imageDisplayUrl(value: unknown, auth: AuthSession | null): string {
  const raw = stringValue(value).trim();
  if (!raw) return "";
  if (/^(blob:|data:|https?:\/\/)/i.test(raw)) return raw;
  if (raw.startsWith("//")) return typeof window === "undefined" ? raw : `${window.location.protocol}${raw}`;
  if (raw.startsWith("/api/")) return raw;

  const base = mainApiDisplayBase(auth?.backendUrl);
  if (!base) return raw;
  if (raw.startsWith("/")) return `${base}${raw}`;
  if (raw.toLowerCase().startsWith("images/")) return `${base}/${raw.replace(/^\/+/, "")}`;
  return `${base}/images/${raw.replace(/^\/+/, "")}`;
}

function mainApiDisplayBase(rawBackendUrl: unknown): string {
  const raw = stringValue(rawBackendUrl).trim();
  if (!raw) return "";
  try {
    const withProtocol = /^https?:\/\//i.test(raw) ? raw : `http://${raw}`;
    const parsed = new URL(withProtocol);
    const path = parsed.pathname.replace(/\/+$/, "");
    parsed.pathname = path.toLowerCase().endsWith("/goapi") ? path.slice(0, -"/goapi".length) || "/" : "/";
    parsed.search = "";
    parsed.hash = "";
    return parsed.toString().replace(/\/$/, "");
  } catch {
    return "";
  }
}

async function resizeLogoFile(file: File): Promise<File> {
  const objectUrl = URL.createObjectURL(file);
  try {
    const image = await loadImageElement(objectUrl);
    const maxSide = Math.max(image.naturalWidth, image.naturalHeight);
    if (!maxSide) return file;
    const scale = Math.min(1, 512 / maxSide);
    if (scale === 1 && file.size <= 300 * 1024) return file;

    const canvas = document.createElement("canvas");
    canvas.width = Math.max(1, Math.round(image.naturalWidth * scale));
    canvas.height = Math.max(1, Math.round(image.naturalHeight * scale));
    const context = canvas.getContext("2d");
    if (!context) return file;
    context.imageSmoothingEnabled = true;
    context.imageSmoothingQuality = "high";
    context.drawImage(image, 0, 0, canvas.width, canvas.height);

    const blob = await canvasToBlob(canvas, "image/webp", 0.82);
    if (!blob) return file;
    const baseName = file.name.replace(/\.[^.]+$/, "") || "company-logo";
    return new File([blob], `${baseName}.webp`, { type: "image/webp", lastModified: Date.now() });
  } finally {
    URL.revokeObjectURL(objectUrl);
  }
}

function loadImageElement(src: string): Promise<HTMLImageElement> {
  return new Promise((resolve, reject) => {
    const image = new window.Image();
    image.onload = () => resolve(image);
    image.onerror = () => reject(new Error("image load failed"));
    image.src = src;
  });
}

function canvasToBlob(canvas: HTMLCanvasElement, type: string, quality: number): Promise<Blob | null> {
  return new Promise((resolve) => canvas.toBlob(resolve, type, quality));
}

function radioFormValue(value: unknown, field: SystemSettingField): string {
  const fallback = field.options?.[0]?.value ?? "";
  if (field.valueType !== "boolean") return stringValue(value) || fallback;
  if (typeof value === "boolean") return value ? "true" : "false";
  const normalized = stringValue(value).toLowerCase();
  if (normalized === "true" || normalized === "false") return normalized;
  if (normalized === "1" || normalized === "yes" || normalized === "y" || normalized === "buddhist" || normalized === "be" || normalized === "พ.ศ.") return "true";
  if (normalized === "0" || normalized === "no" || normalized === "n" || normalized === "christian" || normalized === "ce" || normalized === "ค.ศ.") return "false";
  return fallback;
}

function radioValueToFormValue(value: string, field: SystemSettingField): string | boolean | number {
  if (field.valueType === "boolean") return booleanLikeValue(value);
  if (field.valueType === "number") return Number(value);
  return value;
}

function optionValueToFormValue(value: string, field: SystemSettingField): string | number {
  if (field.valueType === "number") {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : 0;
  }
  return value;
}

function booleanLikeValue(value: unknown): boolean {
  if (typeof value === "boolean") return value;
  const normalized = stringValue(value).toLowerCase();
  return normalized === "true" || normalized === "1" || normalized === "yes" || normalized === "y" || normalized === "buddhist" || normalized === "be" || normalized === "พ.ศ.";
}

function isDecimalSettingField(key: string): boolean {
  return key === "decimal_quantity"
    || key === "decimal_price"
    || key === "decimal_document"
    || key === "settings.decimal_quantity"
    || key === "settings.decimal_price"
    || key === "settings.decimal_document";
}

function normalizeDecimalPlaces(value: unknown): number {
  const numeric = Number(value);
  if (!Number.isFinite(numeric) || numeric < 2) return 2;
  return Math.min(8, Math.trunc(numeric));
}

function comboOptionsForField(field: SystemSettingField, language: LanguageCode, currentValue: string): ComboOption[] {
  const sourceOptions = field.optionSource === "timezones" ? timezoneOptions(language) : field.options ?? [];
  if (!currentValue || sourceOptions.some((option) => option.value === currentValue)) return sourceOptions;
  return [{ value: currentValue, label: currentValue }, ...sourceOptions];
}

function timezoneOptions(language: LanguageCode): ComboOption[] {
  return supportedTimeZones().map((timeZone) => {
    const meta = timezoneMeta(timeZone);
    return { value: timeZone, label: `${meta.offset ? `UTC${meta.offset} ` : ""}${timeZone}` };
  }).sort((first, second) => first.label.localeCompare(second.label, localeOf(language)));
}

function supportedTimeZones(): string[] {
  const intl = Intl as typeof Intl & { supportedValuesOf?: (key: "timeZone") => string[] };
  if (typeof intl.supportedValuesOf !== "function") return [];
  try {
    return intl.supportedValuesOf("timeZone");
  } catch {
    return [];
  }
}

function timezoneMeta(timeZone: string): { label: string; offset: string } {
  const offset = timezoneUtcOffset(timeZone);
  return {
    label: offset ? `(UTC${offset}) ${timeZone}` : timeZone,
    offset,
  };
}

function setTimezoneDerivedPayload(payload: SettingRecord, key: string, timeZone: string) {
  const meta = timezoneMeta(timeZone);
  const prefix = key.includes(".") ? `${key.split(".").slice(0, -1).join(".")}.` : "";
  setByPath(payload, `${prefix}timezone_label`, meta.label);
  setByPath(payload, `${prefix}timezone_offset`, meta.offset);
}

function timezoneUtcOffset(timeZone: string): string {
  try {
    const parts = new Intl.DateTimeFormat("en-US", {
      hour: "2-digit",
      minute: "2-digit",
      timeZone,
      timeZoneName: "longOffset",
    }).formatToParts(new Date());
    const zone = parts.find((part) => part.type === "timeZoneName")?.value ?? "";
    const rawOffset = zone.replace(/^GMT/i, "").trim();
    return rawOffset ? normalizeUtcOffset(rawOffset) : "+00:00";
  } catch {
    return "";
  }
}

function namesToObject(value: unknown): Record<string, string> {
  if (isRecord(value)) return Object.fromEntries(Object.entries(value).map(([key, item]) => [key, String(item ?? "")]));
  if (!Array.isArray(value)) return {};
  return Object.fromEntries(value.filter(isRecord).map((item) => [String(item.code ?? ""), String(item.name ?? "")]));
}

function objectToNames(
  value: unknown,
  language: LanguageCode,
  previousValue?: unknown,
  activeLanguages?: string[],
): { code: string; name: string; isauto: boolean; isdelete: boolean }[] {
  const record = namesToObject(value);
  const previous = namesToObject(previousValue);
  const active = new Set((activeLanguages?.length ? activeLanguages : LANGUAGES.map((item) => item.code)).map((code) => code.toLowerCase()));
  return LANGUAGES.map((item) => ({
    code: item.code,
    name: active.has(item.code)
      ? record[item.code]?.trim() || (item.code === "en" ? record[language]?.trim() : "") || ""
      : previous[item.code]?.trim() || record[item.code]?.trim() || "",
    isauto: false,
    isdelete: false,
  }));
}

function parseJsonField(value: unknown, key: string): unknown {
  if (typeof value !== "string") return value;
  const trimmed = value.trim();
  if (!trimmed) return key === "permissionCodes" || key === "approvalCodes" || key === "allowed_tools" ? [] : {};
  return JSON.parse(trimmed);
}

function getByPath(record: unknown, path: string): unknown {
  if (!isRecord(record)) return undefined;
  return path.split(".").reduce<unknown>((current, key) => isRecord(current) ? current[key] : undefined, record);
}

function setByPath(record: SettingRecord, path: string, value: unknown) {
  const parts = path.split(".");
  let current: SettingRecord = record;
  parts.slice(0, -1).forEach((part) => {
    if (!isRecord(current[part])) current[part] = {};
    current = current[part] as SettingRecord;
  });
  current[parts[parts.length - 1]] = value;
}

function deleteByPath(record: SettingRecord, path: string) {
  const parts = path.split(".");
  let current: unknown = record;
  parts.slice(0, -1).forEach((part) => {
    current = isRecord(current) ? current[part] : undefined;
  });
  if (isRecord(current)) delete current[parts[parts.length - 1]];
}

function shortValue(value: unknown, language: LanguageCode): string {
  if (Array.isArray(value)) {
    const localized = localizedValue(value, language);
    return localized || `${value.length}`;
  }
  if (typeof value === "boolean") return value ? "✓" : "-";
  if (typeof value === "object" && value !== null) return JSON.stringify(value);
  return String(value ?? "-");
}

function fieldDisplayValue(field: SystemSettingField, value: unknown, language: LanguageCode): string {
  if (field.type === "radio" || field.type === "select") {
    const optionValue = field.type === "radio" ? radioFormValue(value, field) : String(value ?? "");
    const option = field.options?.find((item) => item.value === optionValue);
    if (option) return optionLabel(option, language);
  }
  return shortValue(value, language);
}

function isRecord(value: unknown): value is SettingRecord {
  return typeof value === "object" && value !== null && !Array.isArray(value);
}

function isFailed(payload: unknown): boolean {
  if (!isRecord(payload)) return false;
  return payload.success === false || payload.status === "error";
}

function extractMessage(payload: unknown): string | undefined {
  if (typeof payload === "string") return payload;
  if (!isRecord(payload)) return undefined;
  return typeof payload.message === "string" ? payload.message : typeof payload.error === "string" ? payload.error : undefined;
}

function safeJsonParse(value: string, fallback: unknown): unknown {
  try {
    return JSON.parse(value);
  } catch {
    return fallback;
  }
}

function isActiveRecord(record: SettingRecord): boolean {
  if ("is_access_disabled" in record || "is_access_disabled" in record) return !userAccessDisabled(record);
  if ("is_active" in record) return Boolean(record.is_active);
  if ("isActive" in record) return Boolean(record.isActive);
  if ("isenabled" in record) return Boolean(record.isenabled);
  if ("isdisabled" in record) return !record.isdisabled;
  return true;
}

function localeOf(language: LanguageCode): string {
  const map: Record<LanguageCode, string> = {
    th: "th-TH",
    en: "en-US",
    cn: "zh-CN",
    ja: "ja-JP",
    ko: "ko-KR",
    lo: "lo-LA",
    my: "my-MM",
    km: "km-KH",
    vi: "vi-VN",
    ms: "ms-MY",
    id: "id-ID",
    fil: "fil-PH",
  };
  return map[language];
}

function settingIcon(icon: string): ReactNode {
  const props = { size: 22 };
  if (icon === "bot") return <Bot {...props} />;
  if (icon === "building") return <Building2 {...props} />;
  if (icon === "calendar") return <CalendarDays {...props} />;
  if (icon === "calendar-check") return <CalendarCheck2 {...props} />;
  if (icon === "download-cloud") return <DownloadCloud {...props} />;
  if (icon === "file-cog") return <FileCog {...props} />;
  if (icon === "key") return <KeyRound {...props} />;
  if (icon === "link") return <Link2 {...props} />;
  if (icon === "network") return <Network {...props} />;
  if (icon === "shield") return <ShieldCheck {...props} />;
  if (icon === "user-round") return <UserRound {...props} />;
  if (icon === "users") return <UsersRound {...props} />;
  if (icon === "branch") return <GitBranch {...props} />;
  return <Tag {...props} />;
}
