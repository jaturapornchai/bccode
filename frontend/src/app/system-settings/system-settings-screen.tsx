"use client";

import {
  AlertCircle,
  ArrowLeft,
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
  Pencil,
  FileCog,
  FolderOpen,
  FolderPlus,
  GitBranch,
  Globe,
  ImageIcon,
  KeyRound,
  Link2,
  Loader2,
  MapPin,
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
  ChevronDown,
  ChevronRight,
  UploadCloud,
  X,
} from "lucide-react";
import Image from "next/image";
import { useRouter } from "next/navigation";
import {
  type CSSProperties,
  type DragEvent as ReactDragEvent,
  FormEvent,
  ReactNode,
  type PointerEvent as ReactPointerEvent,
  useCallback,
  useEffect,
  useLayoutEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import { MasterPicker } from "@/components/product-barcode/master-picker";
import { Badge } from "@/components/ui/badge";
import { Button } from "@/components/ui/button";
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { DateField, TimeField } from "@/components/ui/date-time-field";
import { Input } from "@/components/ui/input";
import {
  backendText,
  useBackendLanguage,
  type BackendLanguageDictionary,
} from "@/lib/backend-language";
import {
  localTimeToUtcTime,
  normalizeTimeInput,
  timeToMinutes,
  type CalendarYearType,
} from "@/lib/date-time";
import {
  getSystemSettingConfig,
  systemSettingLabel,
  type SystemSettingConfig,
  type SystemSettingField,
  type SystemSettingOption,
} from "@/lib/system-setting-screens";
import { MapPickerDialog } from "@/components/map-picker-dialog";
import { ProductCategoryTreeView } from "./product-category-tree-view";
import { ProductCategoryItemsEditor } from "./product-category-items-editor";
import { WarehouseTreeView } from "./warehouse-tree-view";
import { CompanyBranchTreeView } from "./company-branch-tree-view";
import { WarehouseLocationsEditor } from "./warehouse-locations-editor";
import { ProductBomEditor } from "./product-bom-editor";
import { imageNeedsAuthenticatedFetch } from "@/lib/image-upload-proxy";
import { normalizeThaiTaxBranchCode } from "@/lib/thai-branch-code";
import {
  findThailandAddressMatchesByPostalCode,
  findThailandDistrict,
  findThailandProvince,
  findThailandSubdistrict,
  getThailandDistrictPostalCode,
  getThailandCountry,
  getThailandSubdistrictPostalCode,
  loadThailandAddressData,
  normalizeThaiPostalCode,
  onlyUniqueCodes,
  thailandAddressLabel,
  type ThailandAddressData,
  type ThailandAddressMatch,
  type ThailandDistrict,
  type ThailandProvince,
  type ThailandSubdistrict,
} from "@/lib/thailand-addresses";
import { LANGUAGES, normalizeLanguage, type LanguageCode } from "@/lib/i18n";
import { MENU_SECTIONS, menuText } from "@/lib/menu-data";
import type { MasterEntry, MasterName } from "@/lib/product-barcode/api";
import { pickName } from "@/lib/product-barcode/utils";
import {
  branchDisplayName,
  shopDisplayName,
  type AuthSession,
  type BranchListItem,
  type WorkspaceSession,
  workspaceStorageKeys,
  WORKSPACE_CHANGED_EVENT,
  notifyWorkspaceChanged,
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
  branchOverride?: BranchListItem | null;
  hideChrome?: boolean;
};

type SettingRecord = Record<string, unknown>;
type FormState = Record<string, unknown>;
type Notice = { type: "error" | "info" | "success"; text: string } | null;
type ProductUnitOption = {
  unitcode: string;
  names?: { code?: string; name?: string }[];
};
type AuthenticatedImageCacheEntry = {
  objectUrl: string;
  lastAccessAt: number;
};

const THAILAND_ADDRESS_PRIMARY_FIELD = "contact.province_code";
const THAILAND_ADDRESS_SECONDARY_FIELDS = new Set([
  "contact.district_code",
  "contact.sub_district_code",
  "contact.zip_code",
]);
const AUTHENTICATED_IMAGE_CACHE_MAX_ENTRIES = 80;

type BranchTabId =
  | "general"
  | "address"
  | "department"
  | "workday"
  | "holiday"
  | "pos"
  | "business";

const BRANCH_TABS: ReadonlyArray<{
  id: BranchTabId;
  labels: Record<LanguageCode, string>;
}> = [
  {
    id: "general",
    labels: {
      th: "ทั่วไป",
      en: "General",
      cn: "通用",
      ja: "全般",
      ko: "일반",
      lo: "ທົ່ວໄປ",
      my: "အထွေထွေ",
      km: "ទូទៅ",
      vi: "Chung",
      ms: "Umum",
      id: "Umum",
      fil: "Pangkalahatan",
    },
  },
  {
    id: "address",
    labels: {
      th: "ที่อยู่",
      en: "Address",
      cn: "地址",
      ja: "住所",
      ko: "주소",
      lo: "ທີ່ຢູ່",
      my: "လိပ်စာ",
      km: "អាសយដ្ឋាន",
      vi: "Địa chỉ",
      ms: "Alamat",
      id: "Alamat",
      fil: "Address",
    },
  },
  {
    id: "department",
    labels: {
      th: "แผนก",
      en: "Department",
      cn: "部门",
      ja: "部門",
      ko: "부서",
      lo: "ພະແນກ",
      my: "ဌာန",
      km: "ផ្នែក",
      vi: "Phòng ban",
      ms: "Jabatan",
      id: "Departemen",
      fil: "Departamento",
    },
  },
  {
    id: "workday",
    labels: {
      th: "วันทำงาน",
      en: "Work Day",
      cn: "工作日",
      ja: "勤務日",
      ko: "근무일",
      lo: "ວັນເຮັດວຽກ",
      my: "အလုပ်ရက်",
      km: "ថ្ងៃធ្វើការ",
      vi: "Ngày làm việc",
      ms: "Hari Bekerja",
      id: "Hari Kerja",
      fil: "Araw ng Trabaho",
    },
  },
  {
    id: "holiday",
    labels: {
      th: "วันหยุด",
      en: "Holiday",
      cn: "假期",
      ja: "休日",
      ko: "휴일",
      lo: "ວັນພັກ",
      my: "ပိတ်ရက်",
      km: "ថ្ងៃឈប់សម្រាក",
      vi: "Ngày nghỉ",
      ms: "Cuti",
      id: "Hari Libur",
      fil: "Pista Opisyal",
    },
  },
  {
    id: "pos",
    labels: {
      th: "POS / ภาษี",
      en: "POS / Tax",
      cn: "POS / 税务",
      ja: "POS / 税",
      ko: "POS / 세금",
      lo: "POS / ພາສີ",
      my: "POS / အခွန်",
      km: "POS / ពន្ធ",
      vi: "POS / Thuế",
      ms: "POS / Cukai",
      id: "POS / Pajak",
      fil: "POS / Buwis",
    },
  },
  {
    id: "business",
    labels: {
      th: "รองรับธุรกิจ",
      en: "Business Types",
      cn: "业务类型",
      ja: "業種",
      ko: "사업 유형",
      lo: "ປະເພດທຸລະກິດ",
      my: "လုပ်ငန်းအမျိုးအစား",
      km: "ប្រភេទអាជីវកម្ម",
      vi: "Loại hình",
      ms: "Jenis Perniagaan",
      id: "Jenis Usaha",
      fil: "Uri ng Negosyo",
    },
  },
];

const BRANCH_FIELD_TAB: Record<string, BranchTabId> = {
  languages: "general",
  code: "general",
  companynames: "general",
  names: "general",
  base_currency: "general",
  language: "general",
  timezone: "general",
  timezone_label: "general",
  timezone_offset: "general",
  date_format: "general",
  year_type: "general",
  decimal_quantity: "general",
  decimal_price: "general",
  decimal_document: "general",
  imageuri: "general",
  imageuris: "general",
  logouri: "general",
  "contact.address": "address",
  "contact.country_code": "address",
  "contact.province_code": "address",
  "contact.district_code": "address",
  "contact.sub_district_code": "address",
  "contact.zip_code": "address",
  "contact.latitude": "address",
  "contact.longitude": "address",
  "contact.phone_number": "address",
  businesstype: "pos",
  company_registration_no: "pos",
  is_vat_registered: "pos",
  "pos.tax_id": "pos",
  "pos.vatrate": "pos",
  "pos.isbom": "pos",
  "pos.vattypepurchase": "pos",
  "pos.inquirytypepurchase": "pos",
  "pos.vattypesale": "pos",
  "pos.inquirytypesale": "pos",
  "pos.headerreceiptpos": "pos",
  "pos.footerreceiptpos": "pos",
  couponusetype: "pos",
  paymentrounding: "pos",
  pointconfig: "pos",
  machinetype: "pos",
};

function branchTabForField(fieldKey: string): BranchTabId {
  return BRANCH_FIELD_TAB[fieldKey] ?? "business";
}

function branchTabLabel(
  tabId: BranchTabId,
  language: LanguageCode,
): string {
  const tab = BRANCH_TABS.find((item) => item.id === tabId);
  if (!tab) return tabId;
  return tab.labels[language] ?? tab.labels.en ?? tabId;
}

const POS_TAB_SECTIONS: ReadonlyArray<{
  id: string;
  title: Record<LanguageCode, string>;
  fieldKeys: string[];
}> = [
  {
    id: "registration",
    title: {
      th: "ทะเบียน / VAT",
      en: "Registration / VAT",
      cn: "注册 / VAT",
      ja: "登録 / VAT",
      ko: "등록 / VAT",
      lo: "ການລົງທະບຽນ / VAT",
      my: "မှတ်ပုံတင် / VAT",
      km: "ការចុះបញ្ជី / VAT",
      vi: "Đăng ký / VAT",
      ms: "Pendaftaran / VAT",
      id: "Registrasi / VAT",
      fil: "Pagrehistro / VAT",
    },
    fieldKeys: [
      "businesstype",
      "company_registration_no",
      "is_vat_registered",
      "pos.tax_id",
      "pos.vatrate",
    ],
  },
  {
    id: "tax-types",
    title: {
      th: "ประเภทภาษีซื้อ-ขาย",
      en: "Purchase / Sale Tax",
      cn: "购销税",
      ja: "仕入・売上税",
      ko: "구매/판매 세금",
      lo: "ພາສີຊື້-ຂາຍ",
      my: "ဝယ်/ရောင်း အခွန်",
      km: "ពន្ធទិញ-លក់",
      vi: "Thuế mua-bán",
      ms: "Cukai Beli/Jual",
      id: "Pajak Beli/Jual",
      fil: "Buwis Bili/Benta",
    },
    fieldKeys: [
      "pos.vattypepurchase",
      "pos.inquirytypepurchase",
      "pos.vattypesale",
      "pos.inquirytypesale",
    ],
  },
  {
    id: "receipt",
    title: {
      th: "ใบเสร็จ",
      en: "Receipt",
      cn: "收据",
      ja: "レシート",
      ko: "영수증",
      lo: "ໃບບິນ",
      my: "ပြေစာ",
      km: "បង្កាន់ដៃ",
      vi: "Biên lai",
      ms: "Resit",
      id: "Struk",
      fil: "Resibo",
    },
    fieldKeys: ["pos.headerreceiptpos", "pos.footerreceiptpos"],
  },
  {
    id: "coupons-points",
    title: {
      th: "คูปอง / แต้ม",
      en: "Coupons / Points",
      cn: "优惠券 / 积分",
      ja: "クーポン / ポイント",
      ko: "쿠폰 / 포인트",
      lo: "ຄູປ໋ອງ / ຄະແນນ",
      my: "ကူပွန် / အမှတ်",
      km: "គូប៉ុង / ពិន្ទុ",
      vi: "Coupon / Điểm",
      ms: "Kupon / Mata",
      id: "Kupon / Poin",
      fil: "Kupon / Puntos",
    },
    fieldKeys: ["couponusetype", "pointconfig"],
  },
  {
    id: "rounding",
    title: {
      th: "การปัดเศษ",
      en: "Rounding",
      cn: "舍入",
      ja: "丸め",
      ko: "반올림",
      lo: "ການປັດເສດ",
      my: "အလုံးအလုံး",
      km: "ការបង្គត់",
      vi: "Làm tròn",
      ms: "Pembundaran",
      id: "Pembulatan",
      fil: "Pag-round",
    },
    fieldKeys: ["paymentrounding"],
  },
  {
    id: "pos-other",
    title: {
      th: "POS อื่น ๆ",
      en: "POS Other",
      cn: "POS 其他",
      ja: "POS その他",
      ko: "POS 기타",
      lo: "POS ອື່ນໆ",
      my: "POS အခြား",
      km: "POS ផ្សេងៗ",
      vi: "POS khác",
      ms: "POS Lain",
      id: "POS Lainnya",
      fil: "POS Iba pa",
    },
    fieldKeys: ["pos.isbom", "machinetype"],
  },
];

function posSectionTitle(
  section: (typeof POS_TAB_SECTIONS)[number],
  language: LanguageCode,
): string {
  return section.title[language] ?? section.title.en ?? section.id;
}
const authenticatedImageObjectUrlCache = new Map<
  string,
  AuthenticatedImageCacheEntry
>();
let authenticatedImageCacheOwner = "";

type LanguageConfigFormRow = {
  code: LanguageCode;
  codetranslator: string;
  name: string;
  is_use: boolean;
  isdefault: boolean;
};
type NormalizeLanguageOptions = {
  forcePrimaryFirst?: boolean;
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
  "settings.language": "th",
  "settings.languageconfigs": defaultLanguageConfigs("th"),
};

const branchSetupDefaults: FormState = {
  languages: ["th"],
  language: "th",
  "contact.country_code": "TH",
  "contact.province_code": "",
  "contact.district_code": "",
  "contact.sub_district_code": "",
  "contact.zip_code": "",
  "contact.latitude": 0,
  "contact.longitude": 0,
  "contact.phone_number": "",
  company_registration_no: "",
  is_vat_registered: false,
  "pos.tax_id": "",
  "pos.vatrate": 0,
  "pos.vattypesale": 0,
  "pos.vattypepurchase": 0,
  "pos.inquirytypesale": 0,
  "pos.inquirytypepurchase": 0,
  "pos.headerreceiptpos": "",
  "pos.footerreceiptpos": "",
  "pos.isbom": false,
  base_currency: "THB",
  timezone: "Asia/Bangkok",
  date_format: "dd/MM/yyyy",
  year_type: "buddhist",
  decimal_quantity: 2,
  decimal_price: 2,
  decimal_document: 2,
  machinetype: 0,
  couponusetype: 0,
  paymentrounding: defaultPaymentRoundingJson(),
  pointconfig: defaultPointConfigJson(),
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

const uiText: Partial<
  Record<LanguageCode, Partial<Record<keyof typeof uiEn, string>>>
> = {
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
    resetPassword: "รีเซ็ตรหัสผ่าน",
    resetPasswordConfirm: "ยืนยันรีเซ็ตรหัสผ่านผู้ใช้นี้กลับเป็น 12345?",
    resetPasswordDone: "รีเซ็ตรหัสผ่านเป็น 12345 แล้ว",
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
  "branch.company_registration_no": "company_registration_no",
  "branch.contact.country_code": "country_code",
  "branch.contact.district_code": "district",
  "branch.contact.phone_number": "telephone",
  "branch.contact.province_code": "province",
  "branch.contact.sub_district_code": "subdistrict",
  "branch.contact.zip_code": "zip_code",
  "branch.date_format": "date_format",
  "branch.decimal_document": "decimal_document",
  "branch.decimal_price": "decimal_price",
  "branch.decimal_quantity": "decimal_quantity",
  "branch.is_vat_registered": "vat_status",
  "branch.language": "default_language",
  "branch.machinetype": "machine_type",
  "branch.names": "branch_name",
  "branch.pointconfig": "point_config",
  "branch.timezone": "timezone",
  "branch.year_type": "year_type",
  "branch.languages": "select_data_language",
  "branch.businesstype": "business_type",
  "branch.pos.tax_id": "company_tax_id",
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

const fieldValueAliases: Record<string, string[]> = {
  "branch.contact.country_code": ["contact.countrycode"],
  "branch.contact.district_code": ["contact.districtcode"],
  "branch.contact.phone_number": ["contact.phonenumber"],
  "branch.contact.province_code": ["contact.provincecode"],
  "branch.contact.sub_district_code": ["contact.subdistrictcode"],
  "branch.contact.zip_code": ["contact.zipcode"],
  "branch.pos.tax_id": ["pos.taxid"],
  "branch.year_type": ["yeartype"],
  "product_category_group_select_screen.group_number": ["groupnumber"],
  "product_category_group_select_screen.parent_guid": ["parentguid"],
  "productcategorylist.group_number": ["groupnumber"],
  "productcategorylist.parent_guid": ["parentguid"],
};

const dayNames: Record<LanguageCode, string[]> = {
  th: ["จันทร์", "อังคาร", "พุธ", "พฤหัสบดี", "ศุกร์", "เสาร์", "อาทิตย์"],
  en: [
    "Monday",
    "Tuesday",
    "Wednesday",
    "Thursday",
    "Friday",
    "Saturday",
    "Sunday",
  ],
  cn: ["星期一", "星期二", "星期三", "星期四", "星期五", "星期六", "星期日"],
  ja: ["月曜", "火曜", "水曜", "木曜", "金曜", "土曜", "日曜"],
  ko: ["월요일", "화요일", "수요일", "목요일", "금요일", "토요일", "일요일"],
  lo: ["ຈັນ", "ອັງຄານ", "ພຸດ", "ພະຫັດ", "ສຸກ", "ເສົາ", "ອາທິດ"],
  my: [
    "တနင်္လာ",
    "အင်္ဂါ",
    "ဗုဒ္ဓဟူး",
    "ကြာသပတေး",
    "သောကြာ",
    "စနေ",
    "တနင်္ဂနွေ",
  ],
  km: ["ចន្ទ", "អង្គារ", "ពុធ", "ព្រហស្បតិ៍", "សុក្រ", "សៅរ៍", "អាទិត្យ"],
  vi: [
    "Thứ hai",
    "Thứ ba",
    "Thứ tư",
    "Thứ năm",
    "Thứ sáu",
    "Thứ bảy",
    "Chủ nhật",
  ],
  ms: ["Isnin", "Selasa", "Rabu", "Khamis", "Jumaat", "Sabtu", "Ahad"],
  id: ["Senin", "Selasa", "Rabu", "Kamis", "Jumat", "Sabtu", "Minggu"],
  fil: [
    "Lunes",
    "Martes",
    "Miyerkules",
    "Huwebes",
    "Biyernes",
    "Sabado",
    "Linggo",
  ],
};

function isCreatorRecord(
  record: SettingRecord,
  workspace: WorkspaceSession | null,
): boolean {
  if (Boolean(record.is_creator)) return true;
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

function isSelfUserRecord(
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

export function SystemSettingsScreen({
  embedded = false,
  initialBackendLanguage,
  initialBackendUrl,
  initialLanguage = "th",
  language: externalLanguage,
  route,
  branchOverride = null,
  hideChrome = false,
}: SystemSettingsScreenProps) {
  const router = useRouter();
  const config = getSystemSettingConfig(route);
  const [language, setLanguage] = useState<LanguageCode>(
    externalLanguage ?? initialLanguage,
  );
  const [auth, setAuth] = useState<AuthSession | null>(null);
  const [workspaceState, setWorkspace] = useState<WorkspaceSession | null>(
    null,
  );
  const workspace = useMemo<WorkspaceSession | null>(() => {
    if (!workspaceState) return workspaceState;
    if (!branchOverride) return workspaceState;
    return { ...workspaceState, branch: branchOverride };
  }, [branchOverride, workspaceState]);
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
  const [groupNumber, setGroupNumber] = useState<number | null>(null);
  const [categorySelectedGuid, setCategorySelectedGuid] = useState("");
  const [categorySearchQuery, setCategorySearchQuery] = useState("");
  const [categoryUnsavedChanges, setCategoryUnsavedChanges] = useState(false);
  const groupNumberRef = useRef<number | null>(null);
  groupNumberRef.current = groupNumber;
  const [copySourceEnvironment, setCopySourceEnvironment] = useState<
    "uat" | "pro"
  >("uat");
  const [copyPreview, setCopyPreview] = useState<unknown>(null);
  const [standardUnitDialog, setStandardUnitDialog] =
    useState<StandardUnitDialogState>(emptyStandardUnitDialog);
  const { confirm, confirmationDialog } = useConfirmDialog();
  const activeBackendUrl = auth?.backendUrl ?? initialBackendUrl;

  // BOM Resizable Split States
  const BOM_SPLIT_DEFAULT_LEFT = 30;
  const BOM_SPLIT_STORAGE_KEY = "bc_bom_split_left";
  const BOM_SPLIT_MIN_LEFT = 15;
  const BOM_SPLIT_MAX_LEFT = 85;

  const [bomSplitLeftPercent, setBomSplitLeftPercent] = useState(BOM_SPLIT_DEFAULT_LEFT);
  const [resizingBomSplit, setResizingBomSplit] = useState(false);
  const bomSplitContainerRef = useRef<HTMLDivElement | null>(null);

  // Restore split settings
  useEffect(() => {
    if (typeof window === "undefined") return;
    const saved = window.localStorage.getItem(BOM_SPLIT_STORAGE_KEY);
    if (saved) {
      const parsed = Number(saved);
      if (Number.isFinite(parsed)) {
        setBomSplitLeftPercent(Math.min(BOM_SPLIT_MAX_LEFT, Math.max(BOM_SPLIT_MIN_LEFT, parsed)));
      }
    }
  }, []);

  const bomSplitStyle = useMemo(
    () =>
      ({
        "--bom-list-fr": `${bomSplitLeftPercent}fr`,
        "--bom-detail-fr": `${100 - bomSplitLeftPercent}fr`,
      } as React.CSSProperties),
    [bomSplitLeftPercent],
  );

  const updateBomSplitFromClientX = useCallback((clientX: number) => {
    const container = bomSplitContainerRef.current;
    if (!container) return;
    const rect = container.getBoundingClientRect();
    if (rect.width <= 0) return;
    const offset = clientX - rect.left;
    const next = (offset / rect.width) * 100;
    setBomSplitLeftPercent(Math.min(BOM_SPLIT_MAX_LEFT, Math.max(BOM_SPLIT_MIN_LEFT, next)));
  }, []);

  useEffect(() => {
    if (!resizingBomSplit) return;
    const onPointerMove = (event: PointerEvent) => {
      updateBomSplitFromClientX(event.clientX);
    };
    const onPointerUp = () => {
      setResizingBomSplit(false);
    };
    window.addEventListener("pointermove", onPointerMove);
    window.addEventListener("pointerup", onPointerUp);
    return () => {
      window.removeEventListener("pointermove", onPointerMove);
      window.removeEventListener("pointerup", onPointerUp);
    };
  }, [resizingBomSplit, updateBomSplitFromClientX]);

  const startBomSplitResize = useCallback(
    (event: React.PointerEvent<HTMLDivElement>) => {
      setResizingBomSplit(true);
      updateBomSplitFromClientX(event.clientX);
    },
    [updateBomSplitFromClientX],
  );

  const startBomSplitMouseResize = useCallback(
    (event: React.MouseEvent<HTMLDivElement>) => {
      setResizingBomSplit(true);
      updateBomSplitFromClientX(event.clientX);
    },
    [updateBomSplitFromClientX],
  );

  const moveBomSplitResize = useCallback(
    (event: React.PointerEvent<HTMLDivElement>) => {
      if (!resizingBomSplit) return;
      updateBomSplitFromClientX(event.clientX);
    },
    [resizingBomSplit, updateBomSplitFromClientX],
  );

  const stopBomSplitResize = useCallback((event: React.PointerEvent<HTMLDivElement>) => {
    event.currentTarget.releasePointerCapture(event.pointerId);
    setResizingBomSplit(false);
    setBomSplitLeftPercent((current) => {
      const next = Math.min(BOM_SPLIT_MAX_LEFT, Math.max(BOM_SPLIT_MIN_LEFT, current));
      if (typeof window !== "undefined") {
        window.localStorage.setItem(BOM_SPLIT_STORAGE_KEY, String(Math.round(next)));
      }
      return next;
    });
  }, []);

  const adjustBomSplitWithKeyboard = useCallback((event: React.KeyboardEvent<HTMLDivElement>) => {
    let direction = 0;
    if (event.key === "ArrowLeft") direction = -2;
    else if (event.key === "ArrowRight") direction = 2;
    if (direction === 0) return;
    event.preventDefault();
    setBomSplitLeftPercent((current) => Math.min(BOM_SPLIT_MAX_LEFT, Math.max(BOM_SPLIT_MIN_LEFT, current + direction)));
  }, []);
  const backendLanguage = useBackendLanguage(
    language,
    activeBackendUrl,
    language === initialLanguage ? initialBackendLanguage : undefined,
  );

  const text = useCallback(
    (key: keyof typeof uiEn) => {
      const fallback = uiText[language]?.[key] ?? uiEn[key];
      const backendKey = uiBackendKey(key);
      const value = backendText(backendLanguage, backendKey, fallback);
      return value === backendKey || value === key ? fallback : value;
    },
    [backendLanguage, language],
  );
  const errorText = useCallback(
    (error: unknown, fallback = text("requestFailed")) => {
      const message = error instanceof Error ? error.message : "";
      return message
        ? backendText(backendLanguage, message, message)
        : fallback;
    },
    [backendLanguage, text],
  );
  const saveSuccessText = useCallback(
    (message?: string) => {
      const success = text("saveSucceeded");
      const detail = String(message ?? "").trim();
      return detail && detail !== success && detail !== text("saved")
        ? `${success}: ${detail}`
        : success;
    },
    [text],
  );
  const saveErrorText = useCallback(
    (error: unknown) => {
      return `${text("saveFailed")}: ${errorText(error)}`;
    },
    [errorText, text],
  );
  const title = config
    ? systemSettingTitle(config, language, backendLanguage)
    : route;
  const subtitle = config
    ? (config.subtitle[language] ?? config.subtitle.en ?? config.subtitle.th)
    : "";
  const dateTimeScope = useMemo(
    () => resolveDateTimeScope(workspace),
    [workspace],
  );

  const loadRecords = useCallback(
    async (
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
        const limitValue = (currentConfig.slug === "product_category_group_select_screen" || currentConfig.slug === "productcategorylist") ? "100000" : String(SETTINGS_LIST_PAGE_SIZE);
        const searchParams = new URLSearchParams({
          limit: limitValue,
          offset: String(offset),
          page: String(Math.floor(offset / SETTINGS_LIST_PAGE_SIZE) + 1),
          q: query,
          shopid: currentWorkspace.shop.shopid,
        });
        if ((currentConfig.slug === "product_category_group_select_screen" || currentConfig.slug === "productcategorylist") && groupNumberRef.current !== null) {
          searchParams.set("group-number", String(groupNumberRef.current));
        }
        const scope = resolveDateTimeScope(currentWorkspace);
        if (
          currentConfig.slug === "department" ||
          currentConfig.kind === "restaurant-setting"
        ) {
          if (scope.key) searchParams.set("branch_key", scope.key);
          if (scope.branchcode)
            searchParams.set("branchcode", scope.branchcode);
          if (scope.branchguid)
            searchParams.set("branchguid", scope.branchguid);
        }
        if (currentConfig.kind === "copy-uat")
          searchParams.set("source_environment", copySourceEnvironment);
        const fetchSlug = currentConfig.slug === "permission_link" ? "user" : currentConfig.slug;
        const response = await fetch(
          `/api/system-settings/${fetchSlug}?${searchParams.toString()}`,
          {
            headers: {
              "x-bc-backend-url": currentAuth.backendUrl,
              Authorization: `Bearer ${currentAuth.token}`,
            },
            cache: "no-store",
          },
        );
        const payload = (await response.json()) as unknown;
        if (!response.ok || isFailed(payload))
          throw new Error(extractMessage(payload) ?? text("requestFailed"));
        let nextRecords = normalizeRecords(payload, currentConfig.slug === "permission_link" ? getSystemSettingConfig("user")! : currentConfig);
        if (currentConfig.slug === "permission_link") {
          nextRecords = nextRecords.map((r: any) => ({
            ...r,
            employeeCode: r.username,
            employeeName: r.user_profile_name || r.name || r.username,
          }));
        }
        const scopedRecords =
          currentConfig.slug === "holiday_screen" ||
          currentConfig.slug === "department"
            ? filterRecordsByDateTimeScope(nextRecords, scope)
            : nextRecords;
        const nextTotal = extractRecordTotal(payload, scopedRecords.length);
        setRecordTotal(nextTotal);
        if (!query.trim()) setAllRecordTotal(nextTotal);
        setAllRecordsLoaded(
          scopedRecords.length < SETTINGS_LIST_PAGE_SIZE ||
            offset + scopedRecords.length >= nextTotal,
        );
        setRecords((currentRecords) =>
          append
            ? mergeRecords(currentRecords, scopedRecords, currentConfig)
            : scopedRecords,
        );
        if (currentConfig.slug === "work_day_screen")
          setWorkDays(normalizeWorkDays(nextRecords, language, scope));
      } catch (error) {
        setNotice({ type: "error", text: errorText(error) });
      } finally {
        if (append) setLoadingMore(false);
        else setLoading(false);
      }
    },
    [
      copySourceEnvironment,
      errorText,
      language,
      query,
      setLoading,
      setNotice,
      setRecords,
      setWorkDays,
      text,
    ],
  );

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
    setGroupNumber(null);
    const savedLanguage = normalizeLanguage(
      localStorage.getItem("user_language") ??
        externalLanguage ??
        initialLanguage,
    );
    if (!externalLanguage) setLanguage(savedLanguage);

    const nextAuth = readAuth();
    const nextWorkspace = readWorkspace();
    if (!nextAuth || !nextWorkspace) {
      setNotice({
        type: "info",
        text: "กรุณาเข้าสู่ระบบและเลือกบริษัทก่อนเปิดหน้าจอนี้",
      });
      if (!embedded) router.replace("/");
      return;
    }

    setAuth(nextAuth);
    setWorkspace(nextWorkspace);
    void loadRecords(nextAuth, nextWorkspace, config);
  }, [
    config,
    embedded,
    externalLanguage,
    initialLanguage,
    loadRecords,
    router,
  ]);

  useEffect(() => {
    const handleWorkspaceChange = () => {
      const nextWorkspace = readWorkspace();
      if (nextWorkspace) {
        setWorkspace(nextWorkspace);
      }
    };
    window.addEventListener(WORKSPACE_CHANGED_EVENT, handleWorkspaceChange);
    window.addEventListener("storage", handleWorkspaceChange);
    return () => {
      window.removeEventListener(WORKSPACE_CHANGED_EVENT, handleWorkspaceChange);
      window.removeEventListener("storage", handleWorkspaceChange);
    };
  }, []);

  useEffect(() => {
    if ((config?.slug === "product_category_group_select_screen" || config?.slug === "productcategorylist") && auth && workspace) {
      void loadRecords(auth, workspace, config);
    }
  }, [groupNumber, config, auth, workspace, loadRecords]);

  useEffect(() => {
    setCategorySelectedGuid("");
    setCategorySearchQuery("");
  }, [groupNumber]);

  useEffect(() => {
    document.documentElement.lang = language;
    if (!externalLanguage) localStorage.setItem("user_language", language);
  }, [externalLanguage, language]);

  const visibleRecords = useMemo(() => {
    if (!config) return records;
    const needle = query.trim().toLowerCase();
    if (!needle) return records;
    return records.filter((record) =>
      `${recordTitle(record, config, language)} ${recordId(record, config)} ${JSON.stringify(record)}`
        .toLowerCase()
        .includes(needle),
    );
  }, [config, language, query, records]);
  const displayTotal = recordTotal || visibleRecords.length;
  const totalAllRecords = allRecordTotal || (!query.trim() ? displayTotal : 0);
  const hasMoreRecords = Boolean(
    config &&
    records.length > 0 &&
    records.length < displayTotal &&
    !allRecordsLoaded,
  );
  const loadMoreRecords = useCallback(() => {
    if (
      !auth ||
      !workspace ||
      !config ||
      loading ||
      loadingMore ||
      !hasMoreRecords
    )
      return;
    void loadRecords(auth, workspace, config, records.length, true);
  }, [
    auth,
    config,
    hasMoreRecords,
    loadRecords,
    loading,
    loadingMore,
    records.length,
    workspace,
  ]);
  const selectedRecord = useMemo(() => {
    if (!config || !visibleRecords.length) return null;
    return (
      visibleRecords.find(
        (record) => recordId(record, config) === selectedRecordId,
      ) ??
      visibleRecords[0] ??
      null
    );
  }, [config, selectedRecordId, visibleRecords]);

  useEffect(() => {
    if (!config || !visibleRecords.length) {
      setSelectedRecordId("");
      return;
    }
    if (
      selectedRecordId &&
      visibleRecords.some(
        (record) => recordId(record, config) === selectedRecordId,
      )
    )
      return;
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
  }, [
    autoOpenedCompanyId,
    config,
    editing,
    formOpen,
    language,
    records,
    workspace,
  ]);

  if (!config) {
    return (
      <main className="min-h-dvh w-full bg-background p-3 text-foreground">
        <Card>
          <CardContent className="p-4 text-sm text-destructive">
            ไม่พบหน้าจอ {route}
          </CardContent>
        </Card>
      </main>
    );
  }
  const currentConfig = config;
  const isCreator = workspace?.shop?.is_creator === true ||
    Boolean(auth?.username && workspace?.shop?.createdby && workspace?.shop?.createdby.trim().toLowerCase() === auth.username.trim().toLowerCase());
  const isOwnerOrAdmin = isCreator || workspace?.shop?.role === 1 || workspace?.shop?.role === 2;
  const isProductUnit = currentConfig.slug === "productunit";
  const canEdit = currentConfig.editable !== false && (!isProductUnit || isOwnerOrAdmin);

  function openCreate() {
    setEditing(null);
    setForm(defaultForm(currentConfig, language));
    setFormOpen(true);
    setSelectedRecordId("");
    setNotice(null);
  }

  function handleOpenCategoryCreate(parentGuid?: string) {
    setEditing(null);
    const initialForm = defaultForm(currentConfig, language);
    initialForm.group_number = groupNumber;
    initialForm.parent_guid = parentGuid ?? "";
    setForm(initialForm);
    setFormOpen(true);
    setSelectedRecordId("");
    setNotice(null);
  }

  async function openEdit(record: SettingRecord) {
    const newId = recordId(record, currentConfig);
    const oldId = editing ? recordId(editing, currentConfig) : "";
    if (categoryUnsavedChanges && newId !== oldId) {
      const confirmLeave = await confirm({
        title: language === "th" ? "คุณมีข้อมูลที่ยังไม่ได้บันทึก" : "You have unsaved changes",
        description: language === "th"
          ? "คุณมีข้อมูลสินค้าที่ผูกในหมวดหมู่ที่ยังไม่ได้บันทึก ต้องการเปลี่ยนหมวดหมู่โดยไม่บันทึกหรือไม่?"
          : "You have unsaved changes. Do you want to leave without saving?",
        confirmLabel: language === "th" ? "เปลี่ยนหมวดหมู่โดยไม่บันทึก" : "Leave without saving",
        cancelLabel: language === "th" ? "กลับไปแก้ไข" : "Cancel",
        tone: "warning",
      });
      if (!confirmLeave) {
        setCategorySelectedGuid(categorySelectedGuid);
        return;
      }
      setCategoryUnsavedChanges(false);
    }
    setSelectedRecordId(recordId(record, currentConfig));
    setEditing(record);
    setForm(formFromRecord(record, currentConfig, language));
    setFormOpen(true);
    setNotice(null);

    if (auth && workspace && currentConfig.slug === "permission_link") {
      const empCode = record.employeeCode || recordId(record, currentConfig);
      if (empCode) {
        try {
          const params = new URLSearchParams({ shopid: workspace.shop.shopid });
          const response = await fetch(
            `/api/system-settings/permission_link/${encodeURIComponent(String(empCode))}?${params.toString()}`,
            {
              headers: requestHeaders(auth),
              cache: "no-store",
            },
          );
          const payload = (await response.json()) as unknown;
          if (response.ok && !isFailed(payload)) {
            const detail = normalizeRecords(payload, currentConfig)[0];
            if (detail) {
              const mergedRecord = {
                ...record,
                ...detail,
                employeeCode: record.employeeCode,
                employeeName: record.employeeName,
              };
              setEditing(mergedRecord);
              setForm(formFromRecord(mergedRecord, currentConfig, language));
            }
          }
        } catch {
          // Keep base form if fetch fails or no existing permission config exists
        }
      }
      return;
    }

    if (!auth || !workspace || currentConfig.kind !== "main-crud") return;
    const id = recordId(record, currentConfig);
    if (!id) return;
    try {
      const params = new URLSearchParams({ shopid: workspace.shop.shopid });
      const response = await fetch(
        `/api/system-settings/${currentConfig.slug}/${encodeURIComponent(id)}?${params.toString()}`,
        {
          headers: requestHeaders(auth),
          cache: "no-store",
        },
      );
      const payload = (await response.json()) as unknown;
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
      payload = buildPayload(
        form,
        editing,
        currentConfig,
        workspace,
        auth,
        language,
      );
    } catch (error) {
      setNotice({
        type: "error",
        text: error instanceof Error ? error.message : text("jsonInvalid"),
      });
      return;
    }

    if (currentConfig.slug === "product_category_group_select_screen" || currentConfig.slug === "productcategorylist") {
      const parentGuid = stringValue(
        payload.parent_guid ?? payload.parentguid ?? form.parent_guid,
      );
      const getParentGuidAll = (list: SettingRecord[], pGuid: string): string => {
        if (!pGuid) return "";
        const parent = list.find((r) => productCategoryGuid(r) === pGuid);
        if (!parent) return pGuid;
        const grandParentGuid = productCategoryParentGuid(parent);
        const grandParents = getParentGuidAll(list, grandParentGuid);
        return grandParents ? `${grandParents},${pGuid}` : pGuid;
      };
      payload.parent_guid = parentGuid;
      payload.group_number = Number(
        payload.group_number ?? form.group_number ?? groupNumber ?? 0,
      );
      payload.parentguidall = getParentGuidAll(records, parentGuid);
      const existingXSorts = Array.isArray(payload.xsorts)
        ? (payload.xsorts as unknown[])
        : [];
      if (existingXSorts.length === 0) {
        const currentGuid = productCategoryGuid(editing ?? payload);
        const siblingOrders = records
          .filter(
            (record) =>
              productCategoryGroupNumber(record) === Number(payload.group_number) &&
              productCategoryParentGuid(record) === parentGuid &&
              productCategoryGuid(record) !== currentGuid,
          )
          .map(productCategoryXOrder)
          .filter((order) => Number.isFinite(order));
        payload.xsorts = [
          {
            code: "X",
            xorder:
              siblingOrders.length > 0 ? Math.max(...siblingOrders) + 1 : 1,
          },
        ];
      }
      if (!Array.isArray(payload.codelist)) payload.codelist = [];
      if (!Array.isArray(payload.timeforsales)) payload.timeforsales = [];
      if (payload.colorselecthex) {
        payload.colorselecthex = normalizeHexColor(payload.colorselecthex);
        payload.colorselect = payload.colorselecthex;
      }
    }

    const missingRequired = currentConfig.fields.some(
      (field) =>
        field.required && !String(getByPath(payload, field.key) ?? "").trim(),
    );
    if (missingRequired) {
      setNotice({ type: "error", text: text("errorRequired") });
      return;
    }

    if (!editing && currentConfig.slug === "user") {
      const username = stringValue(payload.username).toLowerCase();
      const duplicateUser = records.some(
        (record) =>
          stringValue(
            record.username ?? record.email ?? record.code,
          ).toLowerCase() === username,
      );
      if (duplicateUser) {
        setNotice({
          type: "error",
          text:
            language === "th"
              ? "รหัสผู้ใช้งานซ้ำ"
              : "User code already exists.",
        });
        return;
      }
    }

    const id = recordId(editing ?? payload, currentConfig);
    setSaving(true);
    setNotice(null);
    try {
      const response = await fetch(
        `/api/system-settings/${currentConfig.slug}${editing && id ? `/${encodeURIComponent(id)}` : ""}`,
        {
          method: editing ? "PUT" : "POST",
          headers: requestHeaders(auth),
          body: JSON.stringify({
            ...payload,
            backendUrl: auth.backendUrl,
            shopid: workspace.shop.shopid,
          }),
        },
      );
      const data = (await response.json()) as unknown;
      if (!response.ok || isFailed(data))
        throw new Error(extractMessage(data) ?? text("requestFailed"));
      const saveMessage = extractMessage(data);
      if (currentConfig.kind === "company") {
        const existingShopInfo = workspace.shopInfo ?? {};
        const mergedShopInfo = { ...existingShopInfo, ...payload };
        const nextWorkspace = { ...workspace, shopInfo: mergedShopInfo };
        setWorkspace(nextWorkspace);
        localStorage.setItem(
          workspaceStorageKeys.workspace,
          JSON.stringify(nextWorkspace),
        );
        localStorage.setItem(
          workspaceStorageKeys.shopInfo,
          JSON.stringify(mergedShopInfo),
        );
        notifyWorkspaceChanged();
        setFormOpen(true);
        setEditing(payload);
      } else {
        if (currentConfig.slug === "branch") {
          const payloadGuid = payload.guid_fixed ?? payload.guid ?? id;
          if (
            payloadGuid &&
            workspace.branch &&
            (payloadGuid === workspace.branch.guid_fixed ||
              payloadGuid === workspace.branch.code)
          ) {
            const nextBranch = { ...workspace.branch, ...payload };
            const nextWorkspace = { ...workspace, branch: nextBranch };
            setWorkspace(nextWorkspace);
            localStorage.setItem(
              workspaceStorageKeys.workspace,
              JSON.stringify(nextWorkspace),
            );
            localStorage.setItem(
              workspaceStorageKeys.branch,
              JSON.stringify(nextBranch),
            );
          }
        }
        if (currentConfig.slug === "user" && isSelfUserRecord(payload, auth)) {
          const nextAuth = {
            ...auth,
            username: stringValue(
              payload.username ?? payload.email ?? payload.code ?? auth.username,
            ),
            profile: {
              ...(auth?.profile ?? {}),
              name: stringValue(
                payload.name1 ?? payload.name ?? auth?.profile?.name,
              ),
              email: stringValue(payload.email ?? auth?.profile?.email),
            },
          };
          localStorage.setItem(
            workspaceStorageKeys.auth,
            JSON.stringify(nextAuth),
          );
        }
        setFormOpen(false);
        setEditing(null);
        notifyWorkspaceChanged();
      }
      await loadRecords(auth, workspace, currentConfig);
      setNotice({ type: "success", text: saveSuccessText(saveMessage) });
    } catch (error) {
      setNotice({ type: "error", text: saveErrorText(error) });
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
      details: id
        ? `${language === "th" ? "รหัสอ้างอิง" : "Reference ID"}: ${id}`
        : undefined,
      confirmLabel: text("delete"),
      cancelLabel: text("cancel"),
      tone: "danger",
    });
    if (!confirmed) return;
    setLoading(true);
    setNotice(null);
    try {
      const response = await fetch(
        `/api/system-settings/${currentConfig.slug}/${encodeURIComponent(id)}?shopid=${encodeURIComponent(workspace.shop.shopid)}`,
        {
          method: "DELETE",
          headers: requestHeaders(auth),
          body: JSON.stringify({
            backendUrl: auth.backendUrl,
            shopid: workspace.shop.shopid,
          }),
        },
      );
      const data = (await response.json()) as unknown;
      if (!response.ok || isFailed(data))
        throw new Error(extractMessage(data) ?? text("requestFailed"));
      setNotice({ type: "success", text: text("saved") });
      notifyWorkspaceChanged();
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
      payload = buildPayload(
        nextForm,
        record,
        currentConfig,
        workspace,
        auth,
        language,
      );
    } catch (error) {
      setNotice({
        type: "error",
        text: error instanceof Error ? error.message : text("jsonInvalid"),
      });
      return;
    }

    setSaving(true);
    setNotice(null);
    try {
      const response = await fetch(
        `/api/system-settings/${currentConfig.slug}/${encodeURIComponent(id)}`,
        {
          method: "PUT",
          headers: requestHeaders(auth),
          body: JSON.stringify({
            ...payload,
            backendUrl: auth.backendUrl,
            shopid: workspace.shop.shopid,
          }),
        },
      );
      const data = (await response.json()) as unknown;
      if (!response.ok || isFailed(data))
        throw new Error(extractMessage(data) ?? text("requestFailed"));
      setNotice({ type: "success", text: text("saved") });
      notifyWorkspaceChanged();
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
    const username = stringValue(
      record.username ?? record.email ?? record.code,
    );
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
      const data = (await response.json()) as unknown;
      if (!response.ok || isFailed(data))
        throw new Error(extractMessage(data) ?? text("requestFailed"));
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
      const response = await fetch(
        `/api/system-settings/${currentConfig.slug}${currentId ? `/${encodeURIComponent(currentId)}` : ""}`,
        {
          method: currentId ? "PUT" : "POST",
          headers: requestHeaders(auth),
          body: JSON.stringify({
            backendUrl: auth.backendUrl,
            body: JSON.stringify(
              buildBranchScopedWorkDayBody(records[0], workspace, workDays),
            ),
            shopid: workspace.shop.shopid,
          }),
        },
      );
      const data = (await response.json()) as unknown;
      if (!response.ok || isFailed(data))
        throw new Error(extractMessage(data) ?? text("requestFailed"));
      setNotice({ type: "success", text: text("saved") });
      notifyWorkspaceChanged();
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
      const response = await fetch(
        `/api/system-settings/${currentConfig.slug}?shopid=${encodeURIComponent(workspace.shop.shopid)}`,
        {
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
        },
      );
      const data = (await response.json()) as unknown;
      if (!response.ok || isFailed(data))
        throw new Error(extractMessage(data) ?? text("requestFailed"));
      setCopyPreview(data);
      setNotice({
        type: "success",
        text: action === "copy" ? text("saved") : text("preview"),
      });
    } catch (error) {
      setNotice({ type: "error", text: errorText(error) });
    } finally {
      setSaving(false);
    }
  }

  async function openStandardUnitDialog() {
    if (!auth || !workspace) return;
    setStandardUnitDialog({
      ...emptyStandardUnitDialog,
      open: true,
      loading: true,
    });
    await loadStandardUnitOptions("");
  }

  async function loadStandardUnitOptions(
    searchText = standardUnitDialog.query,
  ) {
    if (!auth || !workspace) return;
    setStandardUnitDialog((current) => ({
      ...current,
      loading: true,
      error: "",
      query: searchText,
    }));
    try {
      const params = new URLSearchParams({
        backendUrl: auth.backendUrl,
        mainShopId: getMainShopIdFromWorkspace(workspace),
        q: searchText,
      });
      const response = await fetch(
        `/api/workspace/product-units/standard?${params.toString()}`,
        {
          headers: requestHeaders(auth),
          cache: "no-store",
        },
      );
      const payload = (await response.json()) as unknown;
      if (!response.ok || isFailed(payload))
        throw new Error(extractMessage(payload) ?? text("requestFailed"));
      const options =
        isRecord(payload) && Array.isArray(payload.data)
          ? payload.data.filter(isProductUnitOption)
          : [];
      setStandardUnitDialog((current) => ({
        ...current,
        loading: false,
        options,
        selectedCodes: options.map((unit) => unit.unitcode),
        source: isRecord(payload) ? stringValue(payload.source) : "",
      }));
    } catch (error) {
      setStandardUnitDialog((current) => ({
        ...current,
        loading: false,
        error: errorText(error),
      }));
    }
  }

  async function saveStandardUnits() {
    if (!auth || !workspace || standardUnitDialog.selectedCodes.length === 0)
      return;
    setStandardUnitDialog((current) => ({
      ...current,
      saving: true,
      error: "",
    }));
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
      const payload = (await response.json()) as unknown;
      if (!response.ok || isFailed(payload))
        throw new Error(extractMessage(payload) ?? text("requestFailed"));
      setStandardUnitDialog(emptyStandardUnitDialog);
      setNotice({
        type: "success",
        text: extractMessage(payload) ?? text("saved"),
      });
      notifyWorkspaceChanged();
      await loadRecords(auth, workspace, currentConfig);
    } catch (error) {
      setStandardUnitDialog((current) => ({
        ...current,
        saving: false,
        error: errorText(error),
      }));
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
    setStandardUnitDialog((current) => ({
      ...current,
      selectedCodes: current.options.map((unit) => unit.unitcode),
    }));
  }

  function clearStandardUnits() {
    setStandardUnitDialog((current) => ({ ...current, selectedCodes: [] }));
  }

  const showProductCategoryHeaderControls =
    (config.slug === "product_category_group_select_screen" || config.slug === "productcategorylist") &&
    !hideChrome &&
    groupNumber !== null;

  const content = (
    <div className="grid w-full min-w-0 gap-2">
      {hideChrome ? null : (
      <header className="rounded-xl border border-border/80 bg-gradient-to-r from-secondary/15 via-secondary/5 to-transparent px-2.5 py-1.5 shadow-[0_1px_2px_rgba(0,0,0,0.015)]">
        <div className="flex w-full flex-wrap items-center justify-between gap-2">
          <div className="flex min-w-0 flex-1 items-center gap-2">
            <span className="grid size-7 shrink-0 place-items-center rounded-lg bg-primary/10 text-primary shadow-sm">
              {settingIcon(config.icon)}
            </span>
            <div className="min-w-0 flex-1">
              <div className="flex flex-wrap items-center gap-x-2 gap-y-1">
                <h1 className="truncate text-sm sm:text-base font-bold text-foreground leading-none">
                  {title}
                </h1>
                {workspace ? (
                  <div className="flex flex-wrap gap-1 text-[10px] text-muted-foreground">
                    <Badge
                      variant="outline"
                      className="px-1.5 py-0 h-4.5 text-[9px] font-medium border-secondary/40 bg-secondary/5"
                    >
                      {text("company")}: {shopDisplayName(workspace.shop)}
                    </Badge>
                    <Badge
                      variant="outline"
                      className="px-1.5 py-0 h-4.5 text-[9px] font-medium border-secondary/40 bg-secondary/5"
                    >
                      {text("branch")}:{" "}
                      {workspace.branch
                        ? branchDisplayName(workspace.branch)
                        : "-"}
                    </Badge>
                    <Badge
                      variant="outline"
                      className="px-1.5 py-0 h-4.5 text-[9px] font-medium border-secondary/40 bg-secondary/5"
                    >
                      {text("timezone")}:{" "}
                      {dateTimeScope.timezone_label ||
                        dateTimeScope.timezone ||
                        dateTimeScope.timezone_offset ||
                        "-"}
                    </Badge>
                  </div>
                ) : null}
              </div>
              <p className="max-w-[92ch] truncate text-[10px] text-muted-foreground mt-0.5 leading-tight">
                {subtitle}
              </p>
            </div>
          </div>
          <div
            className={cn(
              "flex min-w-0 flex-wrap items-center justify-end gap-1.5",
              showProductCategoryHeaderControls ? "flex-[1_1_34rem]" : "shrink-0",
            )}
          >
            {showProductCategoryHeaderControls ? (
              <div className="flex min-w-0 flex-[1_1_28rem] flex-wrap items-center justify-end gap-1.5">
                <Button
                  type="button"
                  variant="outline"
                  size="icon"
                  className="size-8 shrink-0 rounded-lg"
                  onClick={async () => {
                    if (categoryUnsavedChanges) {
                      const confirmLeave = await confirm({
                        title: language === "th" ? "คุณมีข้อมูลที่ยังไม่ได้บันทึก" : "You have unsaved changes",
                        description: language === "th"
                          ? "คุณมีข้อมูลสินค้าที่ผูกในหมวดหมู่ที่ยังไม่ได้บันทึก ต้องการกลับโดยไม่บันทึกหรือไม่?"
                          : "You have unsaved changes. Do you want to leave without saving?",
                        confirmLabel: language === "th" ? "กลับโดยไม่บันทึก" : "Leave without saving",
                        cancelLabel: language === "th" ? "กลับไปแก้ไข" : "Cancel",
                        tone: "warning",
                      });
                      if (!confirmLeave) return;
                      setCategoryUnsavedChanges(false);
                    }
                    setGroupNumber(null);
                    setCategorySelectedGuid("");
                    setCategorySearchQuery("");
                  }}
                >
                  <ArrowLeft className="size-4" />
                </Button>
                <Badge
                  variant="outline"
                  className="h-8 shrink-0 px-2 text-xs font-semibold"
                >
                  {language === "th" ? `กลุ่ม ${groupNumber}` : `Group ${groupNumber}`}
                </Badge>
                <Input
                  className="h-8 min-w-40 flex-[1_1_14rem] rounded-lg text-sm md:max-w-72"
                  placeholder={language === "th" ? "ค้นหา..." : "Search..."}
                  value={categorySearchQuery}
                  onChange={(event) => setCategorySearchQuery(event.target.value)}
                />
                {config.slug === "product_category_group_select_screen" && (
                  <>
                    <Button
                      type="button"
                      size="sm"
                      className="h-8 shrink-0 rounded-lg gap-1.5"
                      onClick={() => handleOpenCategoryCreate()}
                      disabled={loading || saving}
                    >
                      <Plus className="size-4" />
                      {language === "th" ? "เพิ่มหมวดหลัก" : "Add Root"}
                    </Button>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      className="h-8 shrink-0 rounded-lg gap-1.5"
                      onClick={() => categorySelectedGuid && handleOpenCategoryCreate(categorySelectedGuid)}
                      disabled={loading || saving || !categorySelectedGuid}
                    >
                      <FolderPlus className="size-4 text-emerald-600 dark:text-emerald-400" />
                      {language === "th" ? "เพิ่มหมวดย่อย" : "Add Subcategory"}
                    </Button>
                  </>
                )}
              </div>
            ) : null}
            {embedded ? null : (
              <LanguageDialog
                language={language}
                onLanguageChange={setLanguage}
              />
            )}
            <ManualLink compact language={language} screen={config.manual} />
            {embedded ? null : <ThemeToggle language={language} />}
          </div>
        </div>
      </header>
      )}

      {notice ? (
        <div
          className={`message ${notice.type === "success" ? "success" : notice.type === "error" ? "error" : "info"}`}
          role={notice.type === "error" ? "alert" : "status"}
          aria-live={notice.type === "error" ? "assertive" : "polite"}
        >
          {notice.type === "success" ? (
            <BadgeCheck size={18} />
          ) : (
            <AlertCircle size={18} />
          )}
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
      ) : config.slug === "productgroup" && !hideChrome ? (
        <ProductGroupUnifiedView
          initialBackendLanguage={initialBackendLanguage}
          initialBackendUrl={initialBackendUrl}
          language={language}
        />
      ) : config.slug === "product_warehouse_screen" && !hideChrome ? (
        <WarehouseTreeView
          auth={auth}
          workspace={workspace}
          language={language}
          records={records}
          onRefresh={() => void loadRecords(auth, workspace, config)}
          saving={saving}
          loading={loading}
        />
      ) : (config.slug === "company" || config.slug === "branch") && !hideChrome ? (
        <CompanyBranchTreeView
          auth={auth}
          workspace={workspace}
          language={language}
          onRefresh={() => void loadRecords(auth, workspace, config)}
        />
      ) : config.slug === "product_bom" && !hideChrome ? (
        <div
          ref={bomSplitContainerRef}
          className="grid min-h-0 min-w-0 gap-3 xl:grid-cols-[minmax(0,var(--bom-list-fr))_8px_minmax(0,var(--bom-detail-fr))] xl:gap-0 flex-1 xl:h-full flex-col xl:flex-row items-stretch w-full min-h-[calc(100dvh-12rem)]"
          style={bomSplitStyle}
        >
          {/* Left Column: BOM List / Sidebar */}
          <Card className={cn(
            "min-w-0 overflow-hidden xl:flex xl:h-full xl:min-h-0 xl:flex-col border-b xl:border-b-0 xl:border-r border-border bg-muted/10 rounded-none border-y-0 border-l-0 shadow-none bg-card flex flex-col min-h-[300px] lg:min-h-0",
            selectedRecordId ? "hidden xl:flex" : "flex"
          )}>
            <CardHeader className="p-3 border-b border-border bg-muted/20 shrink-0">
              <div className="flex items-center justify-between gap-2 mb-2">
                <CardTitle className="text-sm font-bold flex items-center gap-1.5">
                  <span>{language === "th" ? "รายการสูตรผลิต" : "Recipes"}</span>
                  <Badge variant="secondary" className="text-[10px] font-semibold h-4 px-1.5">
                    {visibleRecords.length} / {records.length} {language === "th" ? "รายการ" : "items"}
                  </Badge>
                </CardTitle>
                <Button
                  type="button"
                  size="icon"
                  variant="outline"
                  className="h-8 w-8 rounded-full"
                  onClick={() => {
                    const today = new Date().toISOString().substring(0, 10);
                    const newVirtualBOM: any = {
                      guid_fixed: "virtual-" + Math.random().toString(36).substring(2, 11),
                      isNew: true,
                      barcode: "",
                      itemcode: "",
                      names: [{ code: language, name: "" }],
                      item_unit_code: "RECIPE",
                      itemunitnames: [{ code: language, name: language === "th" ? "สูตร" : "Recipe" }],
                      price: 0,
                      bom: [],
                      boms: [{ guid_fixed: "", start_date: today, end_date: null, bom: [] }],
                    };
                    setRecords((prev) => {
                      const clean = prev.filter((r: any) => !r.guid_fixed?.startsWith("virtual-"));
                      return [newVirtualBOM, ...clean];
                    });
                    setSelectedRecordId(newVirtualBOM.guid_fixed);
                  }}
                  disabled={loading}
                >
                  <Plus className="size-4" />
                </Button>
              </div>
              <div className="relative">
                <Search className="absolute left-2.5 top-2.5 size-4 text-muted-foreground" />
                <Input
                  className="pl-9 h-9"
                  placeholder={language === "th" ? "ค้นหาสูตร..." : "Search recipe..."}
                  value={query}
                  onChange={(e) => setQuery(e.target.value)}
                />
              </div>
            </CardHeader>
            <CardContent className="p-0 flex-1 overflow-y-auto scrollbar-thin">
              {loading && !records.length ? (
                <div className="flex items-center justify-center p-8 gap-2 text-sm text-muted-foreground">
                  <Loader2 className="animate-spin size-4" />
                  {text("loading")}
                </div>
              ) : visibleRecords.length === 0 ? (
                <div className="text-center p-8 text-sm text-muted-foreground">
                  {language === "th" ? "ไม่พบสูตรผลิต" : "No recipes found"}
                </div>
              ) : (
                <div className="divide-y divide-border/60">
                  {visibleRecords.map((record: any, index: number) => {
                    const guid = recordId(record, config);
                    const isSelected = guid === selectedRecordId;
                    const name = pickName(record.names, language) || record.barcode || (language === "th" ? "(สูตรใหม่)" : "(New Recipe)");
                    return (
                      <button
                        key={guid}
                        className={cn(
                          "w-full text-left p-3 hover:bg-muted/40 transition-colors flex flex-col gap-1",
                          isSelected
                            ? "bg-muted/80 border-r-2 border-primary"
                            : index % 2 === 0
                              ? "bg-background"
                              : "bg-muted/5"
                        )}
                        onClick={() => {
                          setSelectedRecordId(guid);
                          setEditing(null);
                        }}
                      >
                        <span className="font-semibold text-foreground text-sm break-words">
                          {name}
                        </span>
                        <div className="flex items-center justify-between text-xs text-muted-foreground font-medium">
                          <span className="font-mono">{record.barcode || (language === "th" ? "ยังไม่กำหนด" : "Not set")}</span>
                          <span className="bg-muted px-1.5 py-0.5 rounded font-semibold text-[10px]">
                            {record.item_unit_code || "RECIPE"}
                          </span>
                        </div>
                      </button>
                    );
                  })}
                </div>
              )}
            </CardContent>
          </Card>

          {/* Resizable split separator bar */}
          <div
            aria-label="Adjust layout split"
            aria-orientation="vertical"
            aria-valuemax={BOM_SPLIT_MAX_LEFT}
            aria-valuemin={BOM_SPLIT_MIN_LEFT}
            aria-valuenow={Math.round(bomSplitLeftPercent)}
            className={cn(
              "group hidden cursor-col-resize touch-none items-stretch justify-center rounded-md outline-none xl:flex",
              resizingBomSplit && "cursor-col-resize",
            )}
            onKeyDown={adjustBomSplitWithKeyboard}
            onMouseDown={startBomSplitMouseResize}
            onPointerCancel={stopBomSplitResize}
            onPointerDown={startBomSplitResize}
            onPointerMove={moveBomSplitResize}
            onPointerUp={stopBomSplitResize}
            role="separator"
            tabIndex={0}
          >
            <div
              className={cn(
                "my-1 w-1 rounded-full bg-border transition-colors group-hover:bg-primary group-focus-visible:bg-primary",
                resizingBomSplit && "bg-primary",
              )}
            />
          </div>

          {/* Right Column: Custom BOM Editor */}
          <div className={cn(
            "flex-1 overflow-y-auto bg-background min-h-0 xl:h-full xl:min-h-0 p-4",
            !selectedRecordId ? "hidden xl:block" : "block"
          )}>
            {selectedRecordId && (
              <Button
                variant="ghost"
                size="sm"
                className="mb-4 xl:hidden flex items-center gap-2"
                onClick={() => setSelectedRecordId("")}
              >
                <ArrowLeft className="h-4 w-4" />
                {language === "th" ? "กลับไปที่รายการ" : "Back to list"}
              </Button>
            )}
            <ProductBomEditor
              auth={auth}
              workspace={workspace}
              language={language}
              selectedRecord={selectedRecord as any}
              records={records as any}
              onRefresh={() => void loadRecords(auth, workspace, config)}
              onClose={() => setSelectedRecordId("")}
              setSelectedRecordId={setSelectedRecordId}
              setRecords={setRecords}
            />
          </div>

        </div>
      ) : (config.slug === "product_category_group_select_screen" || config.slug === "productcategorylist") && !hideChrome ? (
        <div
          className={cn(
            "grid w-full min-w-0 items-stretch gap-3",
            groupNumber === null
              ? "grid-cols-1"
              : "min-h-[calc(100dvh-12rem)] xl:grid-cols-[minmax(320px,0.95fr)_minmax(420px,1.05fr)]",
          )}
        >
          <ProductCategoryTreeView
            auth={auth}
            workspace={workspace}
            language={language}
            records={records}
            groupNumber={groupNumber}
            setGroupNumber={setGroupNumber}
            selectedGuid={categorySelectedGuid}
            setSelectedGuid={setCategorySelectedGuid}
            searchQuery={categorySearchQuery}
            onOpenCreate={handleOpenCategoryCreate}
            onOpenEdit={openEdit}
            onDeleteRecord={deleteRecord}
            onRefresh={() => void loadRecords(auth, workspace, config)}
            saving={saving}
            loading={loading}
            readOnly={config.slug === "productcategorylist"}
          />
          {groupNumber === null ? null : (
            <div className="min-h-0 h-full">
              {config.slug === "productcategorylist" ? (
                <ProductCategoryItemsEditor
                  auth={auth}
                  workspace={workspace}
                  language={language}
                  categorySelectedGuid={categorySelectedGuid}
                  categoryRecord={editing}
                  setEditing={setEditing}
                  saving={saving}
                  setSaving={setSaving}
                  onRefresh={() => void loadRecords(auth, workspace, config)}
                  onUnsavedChangesChange={setCategoryUnsavedChanges}
                />
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
                  notice={notice}
                  onClose={() => {
                    if (!saving) setFormOpen(false);
                  }}
                  onSubmit={saveRecord}
                  saving={saving}
                  setForm={setForm}
                  text={text}
                  workspace={workspace}
                />
              ) : (
                <Card className="h-full border-border bg-card shadow-sm">
                  <CardContent className="grid h-full min-h-60 place-items-center p-4 text-center text-sm text-muted-foreground">
                    <div className="grid gap-2">
                      <FolderOpen className="mx-auto size-8 text-primary/70" />
                      <b className="text-foreground">
                        {language === "th" ? "เลือกหรือเพิ่มหมวดสินค้า" : "Select or add a category"}
                      </b>
                      <span>
                        {language === "th"
                          ? "เลือกแถวด้านซ้ายเพื่อแก้ไข หรือกดเพิ่มหมวดหลัก/หมวดย่อย"
                          : "Select a row on the left to edit, or add a root/subcategory."}
                      </span>
                    </div>
                  </CardContent>
                </Card>
              )}
            </div>
          )}
        </div>
      ) : config.slug === "branch" && !branchOverride ? (
        <BranchUnifiedView
          auth={auth}
          config={config}
          dictionary={backendLanguage}
          editing={editing}
          form={form}
          formOpen={formOpen}
          initialBackendLanguage={initialBackendLanguage}
          initialBackendUrl={initialBackendUrl}
          language={language}
          loading={loading}
          notice={notice}
          onCloseForm={() => {
            if (!saving) setFormOpen(false);
          }}
          onDelete={deleteRecord}
          onOpenCreate={openCreate}
          onRefresh={() => void loadRecords(auth, workspace, config)}
          onSelect={(record) => {
            setSelectedRecordId(recordId(record, config));
            void openEdit(record);
          }}
          onSubmit={saveRecord}
          query={query}
          setForm={setForm}
          setQuery={setQuery}
          saving={saving}
          selectedRecord={selectedRecord}
          dateTimeScope={dateTimeScope}
          records={visibleRecords}
          text={text}
          workspace={workspace}
        />
      ) : config.kind === "company" ? (
        loading && !companyRecordForEdit(records, workspace) ? (
          <Card>
            <CardContent className="flex min-h-40 items-center justify-center gap-2 p-4 text-sm text-muted-foreground">
              <Loader2 className="animate-spin" />
              {text("loading")}
            </CardContent>
          </Card>
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
            notice={notice}
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
                <p className="text-sm text-muted-foreground">
                  {language === "th"
                    ? "กำลังเตรียมข้อมูลบริษัทปัจจุบัน"
                    : "Preparing current company settings."}
                </p>
              </div>
            </CardContent>
          </Card>
        )
      ) : (
        <>
          <Card>
            <CardContent className="grid gap-1.5 p-2">
              <div className="grid gap-1.5 lg:grid-cols-[minmax(0,1fr)_auto_auto_auto_auto]">
                <label className="relative block min-w-0">
                  <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    className="h-8 pl-9"
                    placeholder={`${text("search")} ${title}`}
                    value={query}
                    onChange={(event) => setQuery(event.target.value)}
                  />
                </label>
                <div className="flex min-h-8 items-center gap-1.5 rounded-lg border border-border bg-background px-2 text-xs shadow-sm">
                  <span className="font-medium text-muted-foreground">
                    {language === "th" ? "ทั้งหมด" : "Total"}
                  </span>
                  <b className="text-foreground">
                    {totalAllRecords.toLocaleString(localeOf(language))}
                  </b>
                  <span className="h-4 w-px bg-border" aria-hidden="true" />
                  <span className="font-medium text-muted-foreground">
                    {text("active")}
                  </span>
                  <b className="text-foreground">
                    {visibleRecords
                      .filter(isActiveRecord)
                      .length.toLocaleString(localeOf(language))}
                  </b>
                </div>
                {currentConfig.slug === "productunit" && canEdit ? (
                  <Button
                    type="button"
                    variant="outline"
                    size="sm"
                    onClick={() => void openStandardUnitDialog()}
                    disabled={loading || saving || !auth}
                  >
                    <DownloadCloud />
                    {text("findStandardUnits")}
                  </Button>
                ) : null}
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => void loadRecords(auth, workspace, config)}
                  disabled={loading || !auth}
                >
                  {loading ? (
                    <Loader2 className="animate-spin" />
                  ) : (
                    <RefreshCcw />
                  )}
                  {text("refresh")}
                </Button>
                {canEdit && currentConfig.slug !== "permission_link" ? (
                  <Button
                    type="button"
                    size="sm"
                    onClick={openCreate}
                    disabled={!auth}
                  >
                    <Plus />
                    {text("add")}
                  </Button>
                ) : null}
              </div>
            </CardContent>
          </Card>

          {loading && !records.length ? (
            <Card>
              <CardContent className="flex min-h-40 items-center justify-center gap-2 p-4 text-sm text-muted-foreground">
                <Loader2 className="animate-spin" />
                {text("loading")}
              </CardContent>
            </Card>
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
              notice={notice}
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
                  <p className="text-sm text-muted-foreground">
                    {canEdit ? text("emptyHint") : text("readOnlyEmptyHint")}
                  </p>
                  {canEdit && currentConfig.slug !== "permission_link" ? (
                    <Button type="button" onClick={openCreate} disabled={!auth}>
                      <Plus />
                      {text("add")}
                    </Button>
                  ) : null}
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
                if (!standardUnitDialog.saving)
                  setStandardUnitDialog(emptyStandardUnitDialog);
              }}
              onQueryChange={(value) =>
                setStandardUnitDialog((current) => ({
                  ...current,
                  query: value,
                }))
              }
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

  if (embedded)
    return <section className="grid w-full min-w-0 gap-3">{content}</section>;
  return (
    <main className="min-h-dvh w-full overflow-x-hidden bg-background p-2 text-foreground sm:p-3">
      {content}
    </main>
  );
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
  const branchCaption =
    config.slug === "department" || config.kind === "restaurant-setting"
      ? recordBranchCaption(record)
      : "";
  const isUser = config.slug === "user";
  const isCreator = isUser && isCreatorRecord(record, workspace);
  const isSelf = isUser && isSelfUserRecord(record, auth);
  const accessDisabled = isUser && userAccessDisabled(record);
  const isOwnerCreator = workspace?.shop?.is_creator === true ||
    Boolean(auth?.username && workspace?.shop?.createdby && workspace?.shop?.createdby.trim().toLowerCase() === auth.username.trim().toLowerCase());
  const isOwnerOrAdmin = isOwnerCreator || workspace?.shop?.role === 1 || workspace?.shop?.role === 2;
  const isProductUnit = config.slug === "productunit";
  const canEdit = config.editable !== false && (!isProductUnit || isOwnerOrAdmin);
  return (
    <Card
      className={cn(
        "min-w-0 shadow-sm",
        isCreator &&
          "border-amber-200 bg-amber-50/40 dark:border-amber-900 dark:bg-amber-950/15",
        !isCreator &&
          accessDisabled &&
          "border-amber-200 bg-amber-50/30 dark:border-amber-900 dark:bg-amber-950/10",
      )}
    >
      <CardContent className="grid gap-2 p-3">
        <div className="flex min-w-0 items-start justify-between gap-2">
          <div className="min-w-0">
            <h2 className="truncate text-base font-semibold">
              {title || id || "-"}
            </h2>
            <p className="truncate text-xs text-muted-foreground">
              {text("id")}: {id || "-"}
            </p>
            <div className="mt-1 flex flex-wrap gap-1">
              {branchCaption ? (
                <Badge variant="outline" className="max-w-full truncate">
                  {text("branch")}: {branchCaption}
                </Badge>
              ) : null}
              {isCreator ? (
                <Badge variant="warning" className="gap-1">
                  <Crown className="size-3" />
                  {text("creator")}
                </Badge>
              ) : null}
              {isUser ? (
                <Badge
                  variant={accessDisabled ? "warning" : "success"}
                  className="gap-1"
                >
                  {accessDisabled ? (
                    <UserX className="size-3" />
                  ) : (
                    <UserCheck className="size-3" />
                  )}
                  {accessDisabled
                    ? text("accessTemporarilyDisabled")
                    : text("accessEnabled")}
                </Badge>
              ) : null}
            </div>
          </div>
          <div className="flex shrink-0 flex-wrap justify-end gap-1">
            {canEdit ? (
              <Button
                type="button"
                size="icon"
                variant="outline"
                onClick={() => onEdit(record)}
                disabled={isSelf}
                aria-label={
                  isSelf ? text("selfPermissionCannotEdit") : text("edit")
                }
                title={isSelf ? text("selfPermissionCannotEdit") : text("edit")}
              >
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
                aria-label={
                  accessDisabled
                    ? text("enableAccess")
                    : text("temporarilyDisableAccess")
                }
                title={
                  accessDisabled
                    ? text("enableAccess")
                    : text("temporarilyDisableAccess")
                }
              >
                {accessDisabled ? <UserCheck /> : <UserX />}
                <span className="hidden sm:inline">
                  {accessDisabled
                    ? text("enableAccess")
                    : text("temporarilyDisableAccess")}
                </span>
              </Button>
            ) : null}
            {config.kind === "company" || !canEdit ? null : (
              <Button
                type="button"
                size="icon"
                variant="outline"
                onClick={() => onDelete(record)}
                disabled={isCreator || isSelf}
                aria-label={
                  isCreator
                    ? text("creatorCannotDelete")
                    : isSelf
                      ? text("selfPermissionCannotEdit")
                      : text("delete")
                }
                title={
                  isCreator
                    ? text("creatorCannotDelete")
                    : isSelf
                      ? text("selfPermissionCannotEdit")
                      : text("delete")
                }
              >
                <Trash2 />
              </Button>
            )}
          </div>
        </div>
        <div className="grid gap-1 text-xs text-muted-foreground">
          {config.fields.slice(0, 4).map((field) => (
            <span
              className="flex min-w-0 items-center justify-between gap-2 rounded-xl border border-border bg-background px-2 py-1.5"
              key={field.key}
            >
              <span className="truncate">
                {fieldLabel(field, language, config, dictionary)}
              </span>
              <b className="min-w-0 max-w-[60%] truncate text-right text-foreground">
                {shortValue(getByPath(record, field.key), language)}
              </b>
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
  notice,
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
  notice: Notice;
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
  const detailScrollRef = useRef<HTMLDivElement | null>(null);
  const loadMoreRef = useRef<HTMLDivElement | null>(null);
  const [listMaxHeight, setListMaxHeight] = useState(420);
  const [detailMaxHeight, setDetailMaxHeight] = useState(520);
  const [leftPanePercent, setLeftPanePercent] = useState(30);
  const [hoveredRecordId, setHoveredRecordId] = useState("");
  const [mounted, setMounted] = useState(false);
  const columns = useMemo(
    () => settingListColumns(config, language, dictionary, text, auth),
    [config, dictionary, language, text, auth],
  );

  useEffect(() => {
    setMounted(true);
  }, []);

  useEffect(() => {
    const target = loadMoreRef.current;
    if (!target || !hasMore || loadingMore) return;
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) onLoadMore();
      },
      { root: listScrollRef.current, rootMargin: "240px 0px", threshold: 0.01 },
    );
    observer.observe(target);
    return () => observer.disconnect();
  }, [hasMore, loadingMore, onLoadMore]);

  useLayoutEffect(() => {
    const listElement = listScrollRef.current;
    const detailElement = detailScrollRef.current;
    const updateHeight = () => {
      const isDesktop = window.innerWidth >= 1024;
      if (!isDesktop) {
        setListMaxHeight(99999);
        setDetailMaxHeight(99999);
        return;
      }

      if (listElement) {
        const top = listElement.getBoundingClientRect().top;
        const bottomGap = 12;
        setListMaxHeight(
          Math.max(260, Math.floor(window.innerHeight - top - bottomGap)),
        );
      }
      if (detailElement) {
        const top = detailElement.getBoundingClientRect().top;
        const bottomGap = 12;
        setDetailMaxHeight(
          Math.max(260, Math.floor(window.innerHeight - top - bottomGap)),
        );
      }
    };
    updateHeight();
    const resizeObserver = new ResizeObserver(updateHeight);
    resizeObserver.observe(document.body);
    if (listElement) resizeObserver.observe(listElement);
    if (detailElement) resizeObserver.observe(detailElement);
    window.addEventListener("resize", updateHeight);
    window.addEventListener("orientationchange", updateHeight);
    return () => {
      resizeObserver.disconnect();
      window.removeEventListener("resize", updateHeight);
      window.removeEventListener("orientationchange", updateHeight);
    };
  }, [formOpen, records.length, selectedId]);

  const startPaneResize = useCallback(
    (event: ReactPointerEvent<HTMLDivElement>) => {
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
    },
    [],
  );

  return (
    <Card className="overflow-hidden shadow-sm">
      <CardContent className="flex min-h-[520px] lg:min-h-0 flex-col gap-0 p-0 lg:flex-row">
        <section
          className="min-h-0 min-w-0 border-b border-border lg:border-b-0 lg:border-r"
          style={
            mounted &&
            typeof window !== "undefined" &&
            window.innerWidth >= 1024
              ? { flexBasis: `${leftPanePercent}%` }
              : undefined
          }
          aria-label={text("list")}
        >
          <div className="bc-list-toolbar">
            <div className="min-w-0">
              <h2 className="truncate text-sm font-semibold">
                {systemSettingTitle(config, language, dictionary)}
              </h2>
              <p className="text-xs text-muted-foreground">
                {records.length.toLocaleString(localeOf(language))} /{" "}
                {totalRecords.toLocaleString(localeOf(language))}{" "}
                {text("items")}
              </p>
            </div>
          </div>
          <div
            ref={listScrollRef}
            className="min-h-[260px] w-full overflow-x-hidden overflow-y-auto scrollbar-thin"
            style={{ maxHeight: listMaxHeight }}
          >
            <div className="bc-list-header">
              {columns.map((column) => (
                <span
                  className={cn("min-w-0 break-words", column.className)}
                  key={column.key}
                >
                  {column.label}
                </span>
              ))}
              <span className="basis-24 grow text-right">
                {language === "th" ? "จัดการ" : "Actions"}
              </span>
            </div>
            {records.map((record, index) => {
              const id = recordId(record, config);
              const active = id === selectedId;
              const isHovered = id === hoveredRecordId;
              const isEditing = Boolean(editingId && id === editingId);
              const isCreator =
                config.slug === "user" && isCreatorRecord(record, workspace);
              const accessDisabled =
                config.slug === "user" && userAccessDisabled(record);
              const rowStyle = settingListRowStyle({
                active,
                index,
                isEditing,
                isHovered,
              });
              return (
                <div
                  className={cn(
                    "bc-list-row",
                    isEditing
                      ? "bg-amber-100 text-amber-950 dark:bg-amber-950/40 dark:text-amber-100"
                      : active
                        ? "bg-primary/10 text-primary"
                        : index % 2 === 0
                          ? "bg-background"
                          : "bg-muted/30",
                  )}
                  style={rowStyle}
                  aria-current={
                    isEditing ? "step" : active ? "true" : undefined
                  }
                  key={`${id}-${index}`}
                  onClick={() => onSelect(record)}
                  onMouseEnter={() => setHoveredRecordId(id)}
                  onMouseLeave={() =>
                    setHoveredRecordId((current) =>
                      current === id ? "" : current,
                    )
                  }
                  onFocus={() => setHoveredRecordId(id)}
                  onBlur={() =>
                    setHoveredRecordId((current) =>
                      current === id ? "" : current,
                    )
                  }
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
                    <span
                      className={cn("min-w-0 break-words", column.className)}
                      key={column.key}
                    >
                      {column.render(record, {
                        isCreator,
                        accessDisabled,
                        id,
                        columnIndex,
                      })}
                    </span>
                  ))}
                  <span
                    className="flex min-w-0 basis-24 grow flex-wrap justify-end gap-1"
                    onClick={(event) => event.stopPropagation()}
                  >
                    {config.editable !== false ? (
                      <Button
                        type="button"
                        size="icon"
                        variant="outline"
                        className="size-7 rounded-lg bg-background text-primary hover:bg-primary/10 border-border"
                        onClick={() => onEdit(record)}
                        disabled={
                          config.slug === "user" &&
                          isSelfUserRecord(record, auth)
                        }
                        aria-label={text("edit")}
                        title={text("edit")}
                      >
                        <Pencil className="size-3.5" />
                      </Button>
                    ) : null}
                    {config.slug === "user" &&
                    !isCreator &&
                    !isSelfUserRecord(record, auth) ? (
                      <Button
                        type="button"
                        size="icon"
                        variant="outline"
                        className="size-7 rounded-lg bg-background text-primary hover:bg-primary/10 border-border"
                        onClick={() => onResetPassword(record)}
                        disabled={saving}
                        aria-label={text("resetPassword")}
                        title={text("resetPassword")}
                      >
                        <KeyRound className="size-3.5" />
                      </Button>
                    ) : null}
                    {config.kind === "company" ||
                    config.editable === false ? null : (
                      <Button
                        type="button"
                        size="icon"
                        variant="outline"
                        className="size-7 rounded-lg bg-background text-destructive hover:bg-destructive/10 border-border"
                        onClick={() => onDelete(record)}
                        disabled={isCreator || isSelfUserRecord(record, auth)}
                        aria-label={text("delete")}
                        title={text("delete")}
                      >
                        <Trash2 className="size-3.5" />
                      </Button>
                    )}
                  </span>
                </div>
              );
            })}
            <div
              ref={loadMoreRef}
              className="flex min-h-10 items-center justify-center py-2 text-xs text-muted-foreground"
            >
              {loadingMore ? (
                <>
                  <Loader2 className="mr-2 size-4 animate-spin" />
                  {text("loading")}
                </>
              ) : hasMore ? (
                text("loading")
              ) : null}
            </div>
          </div>
        </section>

        <div
          className="hidden w-1.5 shrink-0 cursor-col-resize bg-border/70 transition hover:bg-primary/50 lg:block"
          onPointerDown={startPaneResize}
          role="separator"
          aria-orientation="vertical"
          aria-label={
            language === "th"
              ? "ปรับความกว้างรายการและรายละเอียด"
              : "Resize list and detail panes"
          }
        />

        <section
          ref={detailScrollRef}
          className="min-h-0 min-w-0 flex-1 p-3 overflow-y-auto scrollbar-thin lg:h-full lg:max-h-full"
          style={{ maxHeight: detailMaxHeight }}
          aria-label={text("details")}
        >
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
              notice={notice}
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
  className?: string;
  render: (
    record: SettingRecord,
    meta: {
      accessDisabled: boolean;
      columnIndex: number;
      id: string;
      isCreator: boolean;
    },
  ) => ReactNode;
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
  const backgroundColor =
    index % 2 === 0 ? "rgba(14, 165, 233, 0.055)" : "rgba(14, 165, 233, 0.09)";
  return {
    backgroundColor,
    borderColor: "rgba(14, 116, 144, 0.18)",
    color: "var(--foreground)",
  };
}

function CompanyMultiSelectCell({
  value,
  language,
  auth,
}: {
  value: unknown;
  language: LanguageCode;
  auth: AuthSession | null;
}) {
  const [options, setOptions] = useState<MasterEntry[]>([]);
  const selectedGuids = useMemo(() => {
    if (Array.isArray(value)) {
      return value.map((item) => {
        if (typeof item === "string") return item.trim();
        if (isRecord(item)) return stringValue(item.guid_fixed ?? item.guidfixed ?? item.guid ?? "");
        return "";
      }).filter(Boolean);
    }
    return [];
  }, [value]);

  useEffect(() => {
    if (!auth || selectedGuids.length === 0) return;
    const controller = new AbortController();
    void fetch(`/api/workspace/shops`, {
      headers: requestHeaders(auth),
      cache: "no-store",
      signal: controller.signal,
    })
      .then(async (response) => {
        if (response.ok) return response.json();
      })
      .then((payload) => {
        if (payload && payload.success && Array.isArray(payload.data)) {
          const parsed = payload.data.map((shop: any) => ({
            guidfixed: shop.shopid,
            code: "",
            names: shop.names || [{ code: "th", name: shop.name1 || shop.name || "" }]
          }));
          setOptions(parsed);
        }
      })
      .catch(() => {});
    return () => controller.abort();
  }, [auth, selectedGuids.length]);

  if (selectedGuids.length === 0) return <span>-</span>;

  const names = selectedGuids.map((guid) => {
    const match = options.find((opt) => opt.guidfixed === guid);
    if (match) {
      return companyOptionDisplayName(match, language);
    }
    return "";
  }).filter(Boolean);

  if (names.length === 0) {
    return <span>{selectedGuids.length} บริษัท</span>;
  }

  return <span title={names.join(", ")}>{names.join(", ")}</span>;
}

function settingListColumns(
  config: SystemSettingConfig,
  language: LanguageCode,
  dictionary: BackendLanguageDictionary,
  text: (key: keyof typeof uiEn) => string,
  auth: AuthSession | null,
): SettingListColumn[] {
  if (config.slug === "user") {
    const roleField = config.fields.find((field) => field.key === "role");
    return [
      {
        key: "username",
        label:
          language === "th" ? "รหัสผู้ใช้ หรือ email" : "User code or email",
        className: "basis-36 grow-[2] min-w-[120px] shrink-0",
        render: (record, meta) => (
          <span className="flex min-w-0 items-center gap-1.5 w-full">
            <b
              className="min-w-0 truncate"
              title={stringValue(record.username ?? record.email ?? meta.id)}
            >
              {stringValue(record.username ?? record.email ?? meta.id) || "-"}
            </b>
            {meta.isCreator ? (
              <Badge
                variant="warning"
                className="shrink-0 gap-1 px-1.5 py-0 text-[10px]"
              >
                <Crown className="size-2.5" />
                {text("creator")}
              </Badge>
            ) : null}
          </span>
        ),
      },
      {
        key: "user_profile_name",
        label: language === "th" ? "ชื่อผู้ใช้งาน" : "User name",
        className: "basis-24 grow min-w-[100px] shrink-0",
        render: (record) => (
          <span
            className="block truncate"
            title={stringValue(
              record.user_profile_name ?? record.userprofilename ?? record.name,
            )}
          >
            {stringValue(
              record.user_profile_name ?? record.userprofilename ?? record.name,
            ) || "-"}
          </span>
        ),
      },
      {
        key: "email",
        label: language === "th" ? "อีเมล" : "Email",
        className: "basis-32 grow min-w-[130px] shrink-0 hidden xl:inline-flex",
        render: (record) => (
          <span className="block truncate" title={stringValue(record.email)}>
            {stringValue(record.email) || "-"}
          </span>
        ),
      },
      {
        key: "role",
        label: language === "th" ? "สิทธิ์" : "Role",
        className: "basis-24 grow min-w-[95px] shrink-0 hidden md:inline-flex",
        render: (record) => (
          <span
            className="block truncate"
            title={
              roleField
                ? String(fieldDisplayValue(roleField, record.role, language))
                : String(shortValue(record.role, language))
            }
          >
            {roleField
              ? fieldDisplayValue(roleField, record.role, language)
              : shortValue(record.role, language)}
          </span>
        ),
      },
      {
        key: "status",
        label: language === "th" ? "สถานะ" : "Status",
        className: "basis-24 grow min-w-[95px] shrink-0 hidden lg:inline-flex",
        render: (_record, meta) => (
          <Badge
            variant={meta.accessDisabled ? "warning" : "success"}
            className="shrink-0"
          >
            {meta.accessDisabled
              ? text("accessTemporarilyDisabled")
              : text("accessEnabled")}
          </Badge>
        ),
      },
    ];
  }

  if (config.slug === "employee") {
    return [
      {
        key: "code",
        label: language === "th" ? "รหัสพนักงาน" : "Employee code",
        className: "basis-36 grow-[2] min-w-[120px] shrink-0",
        render: (record, meta) => (
          <b
            className="min-w-0 truncate"
            title={stringValue(record.code ?? meta.id)}
          >
            {stringValue(record.code ?? meta.id) || "-"}
          </b>
        ),
      },
      {
        key: "name",
        label: language === "th" ? "ชื่อพนักงาน" : "Employee name",
        className: "basis-28 grow min-w-[100px] shrink-0",
        render: (record) => (
          <span
            className="block truncate"
            title={stringValue(record.name ?? record.employeeName)}
          >
            {stringValue(record.name ?? record.employeeName) || "-"}
          </span>
        ),
      },
      {
        key: "email",
        label: language === "th" ? "อีเมล" : "Email",
        className: "basis-28 grow min-w-[120px] shrink-0 hidden xl:inline-flex",
        render: (record) => (
          <span className="block truncate" title={stringValue(record.email)}>
            {stringValue(record.email) || "-"}
          </span>
        ),
      },
      {
        key: "status",
        label: language === "th" ? "สถานะ" : "Status",
        className: "basis-24 grow min-w-[90px] shrink-0 hidden sm:inline-flex",
        render: (record) => (
          <Badge
            variant={isActiveRecord(record) ? "success" : "warning"}
            className="shrink-0"
          >
            {isActiveRecord(record)
              ? text("active")
              : language === "th"
                ? "ปิดใช้งาน"
                : "Inactive"}
          </Badge>
        ),
      },
    ];
  }

  const fields = config.fields
    .filter(
      (field) =>
        !["image-upload", "json", "language-configs", "language-list"].includes(
          field.type,
        ),
    )
    .slice(0, 4);
  if (!fields.length) {
    return [
      {
        key: "title",
        label: systemSettingTitle(config, language, dictionary),
        className: "basis-36 grow-[2] min-w-[120px] shrink-0",
        render: (record, meta) =>
          recordTitle(record, config, language) || meta.id || "-",
      },
    ];
  }
  return fields.map((field, columnIndex) => {
    let className = "";
    if (columnIndex === 0) {
      className = "basis-36 grow-[2] min-w-[120px] shrink-0";
    } else if (columnIndex === 1) {
      className = "basis-24 grow min-w-[100px] shrink-0 hidden sm:inline-flex";
    } else if (columnIndex === 2) {
      className = "basis-24 grow min-w-[100px] shrink-0 hidden md:inline-flex";
    } else {
      className = "basis-24 grow min-w-[100px] shrink-0 hidden lg:inline-flex";
    }
    return {
      key: field.key,
      label: fieldLabel(field, language, config, dictionary),
      className,
      render: (record) => {
        const val = getByPath(record, field.key);
        if (field.type === "company-multi-select") {
          return (
            <CompanyMultiSelectCell
              value={val}
              language={language}
              auth={auth}
            />
          );
        }
        const displayVal = fieldDisplayValue(field, val, language);
        return (
          <span className="block truncate" title={String(displayVal)}>
            {displayVal}
          </span>
        );
      },
    };
  });
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
  const isOwnerCreator = workspace?.shop?.is_creator === true ||
    Boolean(auth?.username && workspace?.shop?.createdby && workspace?.shop?.createdby.trim().toLowerCase() === auth.username.trim().toLowerCase());
  const isOwnerOrAdmin = isOwnerCreator || workspace?.shop?.role === 1 || workspace?.shop?.role === 2;
  const isProductUnit = config.slug === "productunit";
  const canEdit = config.editable !== false && (!isProductUnit || isOwnerOrAdmin);
  return (
    <section className="grid gap-4">
      <div className="flex flex-col gap-3 rounded-2xl border border-border/80 bg-gradient-to-r from-secondary/15 via-secondary/5 to-transparent p-4 shadow-sm">
        <header className="flex min-w-0 flex-wrap items-start justify-between gap-3">
          <div className="flex min-w-0 items-start gap-2.5">
            <div className="flex size-10 shrink-0 items-center justify-center rounded-xl bg-primary/10 text-primary mt-0.5 shadow-sm">
              {config.slug === "user" ? (
                <UserRound className="size-5" />
              ) : config.slug === "employee" ? (
                <UsersRound className="size-5" />
              ) : config.slug === "active_languages" ? (
                <Globe className="size-5" />
              ) : config.kind === "company" ? (
                <Building2 className="size-5" />
              ) : (
                <FileCog className="size-5" />
              )}
            </div>
            <div className="min-w-0">
              <h2 className="truncate text-lg font-bold text-foreground leading-snug">
                {title || id || "-"}
              </h2>
              <p className="truncate text-xs text-muted-foreground mt-0.5">
                {text("id")}:{" "}
                <code className="rounded bg-secondary/30 px-1.5 py-0.5 font-mono text-[10px] text-primary">
                  {id || "-"}
                </code>
              </p>
              <div className="mt-2 flex flex-wrap gap-1.5">
                {isCreator ? (
                  <Badge variant="warning" className="gap-1 shadow-sm">
                    <Crown className="size-3" />
                    {text("creator")}
                  </Badge>
                ) : null}
                {isUser ? (
                  <Badge
                    variant={accessDisabled ? "warning" : "success"}
                    className="gap-1 shadow-sm"
                  >
                    {accessDisabled ? (
                      <UserX className="size-3" />
                    ) : (
                      <UserCheck className="size-3" />
                    )}
                    {accessDisabled
                      ? text("accessTemporarilyDisabled")
                      : text("accessEnabled")}
                  </Badge>
                ) : null}
              </div>
            </div>
          </div>
          <div className="flex flex-wrap justify-end gap-1.5">
            {canEdit ? (
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={() => onEdit(record)}
                disabled={isSelf}
                className="h-8 hover:bg-primary/5 hover:text-primary transition-colors text-xs font-semibold"
              >
                <Edit3 className="size-3.5" />
                {text("edit")}
              </Button>
            ) : null}
            {isUser && !isCreator && !isSelf ? (
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={() => onResetPassword(record)}
                disabled={saving}
                className="h-8 hover:bg-primary/5 hover:text-primary transition-colors text-xs font-semibold"
              >
                <KeyRound className="size-3.5" />
                {text("resetPassword")}
              </Button>
            ) : null}
            {isUser && !isCreator && !isSelf ? (
              <Button
                type="button"
                size="sm"
                variant={accessDisabled ? "outline" : "destructive"}
                onClick={() => onToggleAccess(record, !accessDisabled)}
                disabled={saving}
                className="h-8 text-xs font-semibold"
              >
                {accessDisabled ? (
                  <UserCheck className="size-3.5" />
                ) : (
                  <UserX className="size-3.5" />
                )}
                {accessDisabled
                  ? text("enableAccess")
                  : text("temporarilyDisableAccess")}
              </Button>
            ) : null}
            {config.kind === "company" || !canEdit ? null : (
              <Button
                type="button"
                size="sm"
                variant="outline"
                onClick={() => onDelete(record)}
                disabled={isCreator || isSelf}
                className="h-8 hover:bg-destructive/5 hover:text-destructive hover:border-destructive/30 transition-colors text-xs font-semibold"
              >
                <Trash2 className="size-3.5" />
                {text("delete")}
              </Button>
            )}
          </div>
        </header>
      </div>

      <div className="grid gap-2 md:grid-cols-2">
        {config.fields.map((field) => {
          if (field.type === "language-list") {
            return (
              <div className={fieldGridItemClass(field, config)} key={field.key}>
                <LanguageListEditor
                  field={field}
                  form={record}
                  language={language}
                  label={fieldLabel(field, language, config, dictionary)}
                  readOnly
                />
              </div>
            );
          }
          if (field.type === "names") {
            return (
              <div className={fieldGridItemClass(field, config)} key={field.key}>
                <LocalizedNamesReadOnlyDetail
                  config={config}
                  field={field}
                  form={record}
                  language={language}
                  label={fieldLabel(field, language, config, dictionary)}
                  workspace={workspace}
                />
              </div>
            );
          }
          if (isThailandAddressSecondaryField(config, field)) return null;
          if (field.type === "image-upload") {
            const label = fieldLabel(field, language, config, dictionary);
            return (
              <div className={fieldGridItemClass(field, config)} key={field.key}>
                <ImageUploadReadOnlyDetail
                  auth={auth}
                  label={label}
                  language={language}
                  value={getByPath(record, field.key)}
                />
              </div>
            );
          }
          if (field.type === "image-gallery") {
            const label = fieldLabel(field, language, config, dictionary);
            return (
              <div className={fieldGridItemClass(field, config)} key={field.key}>
                <ImageGalleryReadOnlyDetail
                  auth={auth}
                  label={label}
                  language={language}
                  value={getByPath(record, field.key)}
                />
              </div>
            );
          }
          if (field.type === "branch-multi-select") {
            const label = fieldLabel(field, language, config, dictionary);
            return (
              <div className={fieldGridItemClass(field, config)} key={field.key}>
                <BranchMultiSelectReadOnlyDetail
                  label={label}
                  language={language}
                  value={getByPath(record, field.key)}
                />
              </div>
            );
          }
          if (field.type === "company-multi-select") {
            const label = fieldLabel(field, language, config, dictionary);
            return (
              <div className={fieldGridItemClass(field, config)} key={field.key}>
                <CompanyMultiSelectReadOnlyDetail
                  label={label}
                  language={language}
                  value={getByPath(record, field.key)}
                  auth={auth}
                />
              </div>
            );
          }
          if (isThailandAddressPrimaryField(config, field)) {
            return (
              <div className={fieldGridItemClass(field, config)} key={field.key}>
                <ThailandAddressReadOnlyDetail
                  backendUrl={auth?.backendUrl}
                  form={record}
                  language={language}
                />
              </div>
            );
          }
          if (field.type === "time-sale-list") {
            return (
              <div className={fieldGridItemClass(field, config)} key={field.key}>
                <TimeSaleListReadOnlyDetail
                  label={fieldLabel(field, language, config, dictionary)}
                  language={language}
                  value={getByPath(record, field.key)}
                />
              </div>
            );
          }
          if (isBranchStructuredSettingField(config, field)) {
            return (
              <div className={fieldGridItemClass(field, config)} key={field.key}>
                <BranchStructuredSettingEditor
                  dictionary={dictionary}
                  field={field}
                  form={record}
                  label={fieldLabel(field, language, config, dictionary)}
                  language={language}
                  readOnly
                />
              </div>
            );
          }
          if (
            config.slug === "permission_definition" &&
            field.key === "branches"
          ) {
            const dateTimeScope = resolveDateTimeScope(workspace);
            return (
              <div className="md:col-span-2" key={field.key}>
                <PermissionMatrixEditor
                  dateTimeScope={dateTimeScope}
                  dictionary={dictionary}
                  form={record}
                  language={language}
                  readOnly
                  text={text}
                />
              </div>
            );
          }
          if (config.slug === "approval_setting" && field.key === "approvals") {
            return (
              <div className="md:col-span-2" key={field.key}>
                <ApprovalSettingEditor
                  dictionary={dictionary}
                  form={record}
                  readOnly
                />
              </div>
            );
          }
          if (
            (config.slug === "permission_link" ||
              config.slug === "permission_group") &&
            (field.key === "permissionCodes" || field.key === "approvalCodes")
          ) {
            return (
              <div className="md:col-span-2" key={field.key}>
                <PermissionLinkMultiSelectEditor
                  config={config}
                  auth={auth}
                  dictionary={dictionary}
                  field={field}
                  form={record}
                  language={language}
                  readOnly
                  workspace={workspace}
                />
              </div>
            );
          }
          return (
            <div
              className="grid gap-1 rounded-xl border border-border bg-card hover:bg-secondary/5 px-3 py-2 text-sm shadow-[0_1px_2px_rgba(0,0,0,0.02)] transition-all duration-200 border-l-2 border-l-secondary"
              key={field.key}
            >
              <span className="text-xs font-semibold text-muted-foreground">
                {fieldLabel(field, language, config, dictionary)}
              </span>
              <b className="min-w-0 break-words text-foreground font-medium">
                {fieldDisplayValue(
                  field,
                  getByPath(record, field.key),
                  language,
                )}
              </b>
            </div>
          );
        })}
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
  notice,
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
  notice?: Notice;
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
        inline
          ? "w-full"
          : "max-h-[calc(100dvh-24px)] w-[min(860px,calc(100vw-24px))] overflow-hidden shadow-xl",
      )}
      onSubmit={onSubmit}
      role={inline ? undefined : "dialog"}
      aria-modal={inline ? undefined : true}
      aria-label={editing ? text("edit") : text("newItem")}
    >
      <header className="flex min-w-0 items-center justify-between gap-2">
        <div className="min-w-0">
          <h2 className="truncate text-lg font-semibold">
            {editing ? text("edit") : text("newItem")}:{" "}
            {systemSettingTitle(config, language, dictionary)}
          </h2>
          <div className="flex flex-wrap gap-1 text-xs text-muted-foreground">
            <span>{config.route}</span>
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
        {inline ? null : (
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
        className={cn(
          "grid min-h-0 gap-2 pr-1",
          inline ? "" : "overflow-y-auto",
        )}
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
          <div className="grid gap-2 md:grid-cols-2">
            {config.fields.map((field) => {
              if (isThailandAddressSecondaryField(config, field)) return null;
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

      <footer className="flex flex-wrap items-center justify-end gap-2">
        {inline && notice ? (
          <div
            className={`message ${notice.type === "success" ? "success" : notice.type === "error" ? "error" : "info"} min-w-0 flex-1`}
            role={notice.type === "error" ? "alert" : "status"}
            aria-live={notice.type === "error" ? "assertive" : "polite"}
          >
            {notice.type === "success" ? (
              <BadgeCheck size={18} />
            ) : (
              <AlertCircle size={18} />
            )}
            <span>{notice.text}</span>
          </div>
        ) : null}
        {inline ? null : (
          <Button
            type="button"
            variant="outline"
            onClick={onClose}
            disabled={saving}
          >
            {text("cancel")}
          </Button>
        )}
        <Button type="submit" disabled={saving}>
          {saving ? <Loader2 className="animate-spin" /> : <Save />}
          {text("save")}
        </Button>
      </footer>
    </form>
  );
  if (inline) return formElement;
  return (
    <div className="dialog-backdrop" role="presentation">
      {formElement}
    </div>
  );
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
      description:
        language === "th"
          ? "กำหนดระดับสิทธิ์ในร้าน และเปิดหรือปิดการเข้าใช้งานของผู้ใช้นี้"
          : "Set the user's shop role and whether this user can access the system.",
    },
    {
      keys: ["position", "department", "line_user_id", "line_display_name"],
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
            className="grid gap-2 rounded-2xl border border-border bg-background/70 p-2"
            key={section.title}
          >
            <header className="grid gap-0.5">
              <h3 className="text-sm font-semibold">{section.title}</h3>
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

function fieldGridItemClass(
  field: SystemSettingField,
  config: SystemSettingConfig,
): string {
  if (config.kind === "company") return "min-w-0 md:col-span-2";
  if (isThailandAddressPrimaryField(config, field))
    return "min-w-0 md:col-span-2";
  if (isBranchLatitudeField(config, field)) return "min-w-0 md:col-span-2";
  if (
    field.type === "branch-multi-select" ||
    field.type === "company-multi-select" ||
    field.type === "image-upload" ||
    field.type === "image-gallery" ||
    field.type === "json" ||
    field.type === "language-configs" ||
    field.type === "language-list" ||
    field.type === "master-picker" ||
    field.type === "names" ||
    field.type === "time-sale-list" ||
    field.type === "textarea"
  ) {
    return "min-w-0 md:col-span-2";
  }
  if (
    (config.slug === "permission_definition" && field.key === "branches") ||
    (config.slug === "approval_setting" && field.key === "approvals") ||
    (config.slug === "permission_link" &&
      (field.key === "employeeCode" ||
        field.key === "permissionCodes" ||
        field.key === "approvalCodes"))
  ) {
    return "min-w-0 md:col-span-2";
  }
  return "min-w-0";
}

function isThailandAddressPrimaryField(
  config: SystemSettingConfig,
  field: SystemSettingField,
): boolean {
  return (
    config.slug === "branch" && field.key === THAILAND_ADDRESS_PRIMARY_FIELD
  );
}

function isThailandAddressSecondaryField(
  config: SystemSettingConfig,
  field: SystemSettingField,
): boolean {
  return (
    config.slug === "branch" && THAILAND_ADDRESS_SECONDARY_FIELDS.has(field.key)
  );
}

function isBranchLatitudeField(
  config: SystemSettingConfig,
  field: SystemSettingField,
): boolean {
  return config.slug === "branch" && field.key === "contact.latitude";
}

function isBranchLongitudeField(
  config: SystemSettingConfig,
  field: SystemSettingField,
): boolean {
  return config.slug === "branch" && field.key === "contact.longitude";
}

function parseCoordinateValue(value: unknown): number | null {
  if (typeof value === "number" && Number.isFinite(value)) return value;
  if (typeof value === "string") {
    const text = value.trim();
    if (!text) return null;
    const parsed = Number(text);
    return Number.isFinite(parsed) ? parsed : null;
  }
  return null;
}

function LocalizedNamesReadOnlyDetail({
  config,
  field,
  form,
  label,
  language,
  workspace,
}: {
  config: SystemSettingConfig;
  field: SystemSettingField;
  form: SettingRecord;
  label: string;
  language: LanguageCode;
  workspace?: WorkspaceSession | null;
}) {
  const names = getLocalizedNameArray(getPathOrFlatValue(form, field.key));
  const editorLanguages = nameEditorLanguageCodes(form, config, language, workspace);
  return (
    <section className="grid gap-1 rounded-2xl border border-border bg-background p-2 text-sm font-semibold md:col-span-2">
      <div>{label}</div>
      <div
        className={cn(
          "grid gap-2",
          editorLanguages.length > 1 && "md:grid-cols-2",
        )}
      >
        {editorLanguages.map((code, index) => {
          const currentValue =
            names.find((item) => item.code === code)?.name ?? "";
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
                  aria-readonly
                  className="min-h-20 w-full rounded-2xl border border-input bg-background px-3 py-2 text-sm text-foreground shadow-sm outline-none"
                  readOnly
                  value={currentValue || "-"}
                />
              ) : (
                <Input aria-readonly readOnly value={currentValue || "-"} />
              )}
            </label>
          );
        })}
      </div>
    </section>
  );
}

function ThailandAddressReadOnlyDetail({
  backendUrl,
  form,
  language,
}: {
  backendUrl?: string;
  form: SettingRecord;
  language: LanguageCode;
}) {
  const [data, setData] = useState<ThailandAddressData | null>(null);
  const [loadError, setLoadError] = useState("");
  const countryCode = stringValue(getPathOrFlatValue(form, "contact.country_code")) || "TH";
  const provinceCode = stringValue(getPathOrFlatValue(form, "contact.province_code"));
  const districtCode = stringValue(getPathOrFlatValue(form, "contact.district_code"));
  const subdistrictCode = stringValue(
    getPathOrFlatValue(form, "contact.sub_district_code"),
  );
  const postalCode = normalizeThaiPostalCode(
    getPathOrFlatValue(form, "contact.zip_code"),
  );
  const labels = thailandAddressUi(language);

  useEffect(() => {
    if (countryCode !== "TH") return;
    let active = true;
    setLoadError("");
    loadThailandAddressData(backendUrl)
      .then((nextData) => {
        if (active) setData(nextData);
      })
      .catch((error: unknown) => {
        if (!active) return;
        setLoadError(error instanceof Error ? error.message : "load failed");
      });
    return () => {
      active = false;
    };
  }, [backendUrl, countryCode]);

  const country = data ? getThailandCountry(data) : undefined;
  const selectedProvince = data
    ? findThailandProvince(data, provinceCode)
    : undefined;
  const selectedDistrict = data
    ? findThailandDistrict(data, provinceCode, districtCode)
    : undefined;
  const selectedSubdistrict = data
    ? findThailandSubdistrict(
        data,
        provinceCode,
        districtCode,
        subdistrictCode,
      )
    : undefined;
  const loading = countryCode === "TH" && !country && !loadError;

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm font-semibold md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span>{labels.title}</span>
        <span className="text-xs font-medium text-muted-foreground">
          {loading
            ? labels.loading
            : loadError
              ? labels.loadError
              : countryCode === "TH"
                ? `${country?.provinces.length ?? 0} ${labels.provinces}`
                : countryCode}
        </span>
      </div>
      <div className="grid gap-2 md:grid-cols-2">
        <ReadOnlyDetailValue
          label={labels.province}
          value={thailandAddressReadOnlyValue(
            provinceCode,
            selectedProvince
              ? thailandAddressLabel(selectedProvince.name, language)
              : "",
          )}
        />
        <ReadOnlyDetailValue
          label={labels.district}
          value={thailandAddressReadOnlyValue(
            districtCode,
            selectedDistrict
              ? thailandAddressLabel(selectedDistrict.name, language)
              : "",
          )}
        />
        <ReadOnlyDetailValue
          label={labels.subdistrict}
          value={thailandAddressReadOnlyValue(
            subdistrictCode,
            selectedSubdistrict
              ? thailandAddressLabel(selectedSubdistrict.name, language)
              : "",
          )}
        />
        <ReadOnlyDetailValue label={labels.postalCode} value={postalCode || "-"} />
      </div>
    </section>
  );
}

function ReadOnlyDetailValue({
  label,
  value,
}: {
  label: string;
  value: string;
}) {
  return (
    <label className="grid gap-1 text-sm font-semibold">
      <span>{label}</span>
      <Input aria-readonly readOnly value={value || "-"} />
    </label>
  );
}

function thailandAddressReadOnlyValue(code: string, label: string): string {
  if (code && label) return `${label} (${code})`;
  return code || label || "-";
}

function ThailandAddressFieldEditor({
  backendUrl,
  form,
  language,
  setForm,
}: {
  backendUrl?: string;
  form: FormState;
  language: LanguageCode;
  setForm: (form: FormState) => void;
}) {
  const [data, setData] = useState<ThailandAddressData | null>(null);
  const [loadError, setLoadError] = useState("");
  const countryCode = stringValue(form["contact.country_code"]) || "TH";
  const provinceCode = stringValue(form["contact.province_code"]);
  const districtCode = stringValue(form["contact.district_code"]);
  const subdistrictCode = stringValue(form["contact.sub_district_code"]);
  const postalCode = normalizeThaiPostalCode(form["contact.zip_code"]);

  useEffect(() => {
    if (countryCode !== "TH") return;
    let active = true;
    setLoadError("");
    loadThailandAddressData(backendUrl)
      .then((nextData) => {
        if (active) setData(nextData);
      })
      .catch((error: unknown) => {
        if (!active) return;
        setLoadError(error instanceof Error ? error.message : "load failed");
      });
    return () => {
      active = false;
    };
  }, [backendUrl, countryCode]);

  const country = data ? getThailandCountry(data) : undefined;
  const postalMatches = useMemo(
    () =>
      data && postalCode.length === 5
        ? findThailandAddressMatchesByPostalCode(data, postalCode)
        : [],
    [data, postalCode],
  );
  const selectedProvince = data
    ? findThailandProvince(data, provinceCode)
    : undefined;
  const selectedDistrict = data
    ? findThailandDistrict(data, provinceCode, districtCode)
    : undefined;
  const selectedSubdistrict = data
    ? findThailandSubdistrict(
        data,
        provinceCode,
        districtCode,
        subdistrictCode,
      )
    : undefined;

  const provinces = filterThailandProvinces(
    country?.provinces ?? [],
    postalMatches,
  );
  const districts = filterThailandDistricts(
    selectedProvince?.districts ?? [],
    postalMatches,
    provinceCode,
  );
  const subdistricts = filterThailandSubdistricts(
    selectedDistrict?.subdistricts ?? [],
    postalMatches,
    provinceCode,
    districtCode,
  );

  function setAddress(nextValues: Partial<FormState>) {
    setForm({ ...form, ...nextValues });
  }

  function chooseProvince(nextProvinceCode: string) {
    const nextValues: Partial<FormState> = {
      "contact.province_code": nextProvinceCode,
      "contact.district_code": "",
      "contact.sub_district_code": "",
      "contact.zip_code": "",
    };
    if (data && postalCode.length === 5 && nextProvinceCode) {
      const matches = postalMatches.filter(
        (item) => item.provinceCode === nextProvinceCode,
      );
      if (matches.length > 0) {
        nextValues["contact.zip_code"] = postalCode;
        const nextDistrictCode = singleThailandAddressCode(
          matches,
          (item) => item.districtCode,
        );
        if (nextDistrictCode) {
          nextValues["contact.district_code"] = nextDistrictCode;
          const nextSubdistrictCode = singleThailandAddressCode(
            matches.filter((item) => item.districtCode === nextDistrictCode),
            (item) => item.subdistrictCode,
          );
          if (nextSubdistrictCode)
            nextValues["contact.sub_district_code"] = nextSubdistrictCode;
        }
      }
    }
    setAddress(nextValues);
  }

  function chooseDistrict(nextDistrictCode: string) {
    const nextValues: Partial<FormState> = {
      "contact.district_code": nextDistrictCode,
      "contact.sub_district_code": "",
      "contact.zip_code": "",
    };
    if (data && provinceCode && nextDistrictCode) {
      const matches =
        postalCode.length === 5
          ? postalMatches.filter(
              (item) =>
                item.provinceCode === provinceCode &&
                item.districtCode === nextDistrictCode,
            )
          : [];
      nextValues["contact.zip_code"] =
        matches.length > 0
          ? postalCode
          : getThailandDistrictPostalCode(data, provinceCode, nextDistrictCode);
      if (postalCode.length === 5) {
        const nextSubdistrictCode = singleThailandAddressCode(
          matches,
          (item) => item.subdistrictCode,
        );
        if (nextSubdistrictCode)
          nextValues["contact.sub_district_code"] = nextSubdistrictCode;
      }
    }
    setAddress(nextValues);
  }

  function chooseSubdistrict(nextSubdistrictCode: string) {
    const nextValues: Partial<FormState> = {
      "contact.sub_district_code": nextSubdistrictCode,
      "contact.zip_code": "",
    };
    if (data && provinceCode && districtCode) {
      nextValues["contact.zip_code"] = nextSubdistrictCode
        ? getThailandSubdistrictPostalCode(
            data,
            provinceCode,
            districtCode,
            nextSubdistrictCode,
          )
        : getThailandDistrictPostalCode(data, provinceCode, districtCode);
    }
    setAddress(nextValues);
  }

  function applyPostalCode(rawValue: string) {
    const normalized = normalizeThaiPostalCode(rawValue);
    const nextValues: Partial<FormState> = { "contact.zip_code": normalized };
    if (data && normalized.length === 5) {
      const matches = findThailandAddressMatchesByPostalCode(data, normalized);
      const nextProvinceCode = singleThailandAddressCode(
        matches,
        (item) => item.provinceCode,
      );
      if (nextProvinceCode) {
        nextValues["contact.province_code"] = nextProvinceCode;
        const provinceMatches = matches.filter(
          (item) => item.provinceCode === nextProvinceCode,
        );
        const nextDistrictCode = singleThailandAddressCode(
          provinceMatches,
          (item) => item.districtCode,
        );
        nextValues["contact.district_code"] = nextDistrictCode;
        if (nextDistrictCode) {
          nextValues["contact.sub_district_code"] = singleThailandAddressCode(
            provinceMatches.filter(
              (item) => item.districtCode === nextDistrictCode,
            ),
            (item) => item.subdistrictCode,
          );
        } else {
          nextValues["contact.sub_district_code"] = "";
        }
      } else {
        nextValues["contact.province_code"] = "";
        nextValues["contact.district_code"] = "";
        nextValues["contact.sub_district_code"] = "";
      }
    }
    setAddress(nextValues);
  }

  if (countryCode !== "TH") {
    return (
      <ThailandAddressFreeTextEditor
        form={form}
        language={language}
        setForm={setForm}
      />
    );
  }

  const loading = !country && !loadError;
  const labels = thailandAddressUi(language);

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm font-semibold md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span>{labels.title}</span>
        <span className="text-xs font-medium text-muted-foreground">
          {loading
            ? labels.loading
            : loadError
              ? labels.loadError
              : `${country?.provinces.length ?? 0} ${labels.provinces}`}
        </span>
      </div>
      <div className="grid gap-2 md:grid-cols-2">
        <ThailandAddressSelect
          label={labels.province}
          value={provinceCode}
          disabled={loading || Boolean(loadError)}
          placeholder={labels.selectProvince}
          onChange={chooseProvince}
          options={provinces.map((province) => ({
            code: province.code,
            label: thailandAddressOptionLabel(province, language),
          }))}
        />
        <ThailandAddressSelect
          label={labels.district}
          value={districtCode}
          disabled={!selectedProvince || loading || Boolean(loadError)}
          placeholder={labels.selectDistrict}
          onChange={chooseDistrict}
          options={districts.map((district) => ({
            code: district.code,
            label: thailandAddressOptionLabel(district, language),
          }))}
        />
        <ThailandAddressSelect
          label={labels.subdistrict}
          value={subdistrictCode}
          disabled={!selectedDistrict || loading || Boolean(loadError)}
          placeholder={labels.selectSubdistrict}
          onChange={chooseSubdistrict}
          options={subdistricts.map((subdistrict) => ({
            code: subdistrict.code,
            label: thailandAddressOptionLabel(subdistrict, language),
          }))}
        />
        <label className="grid gap-1">
          <span>{labels.postalCode}</span>
          <Input
            inputMode="numeric"
            maxLength={5}
            value={postalCode}
            onChange={(event) => applyPostalCode(event.target.value)}
            placeholder="10200"
          />
        </label>
      </div>
      <p className="text-xs font-medium text-muted-foreground">
        {postalAddressHint(
          postalCode,
          postalMatches,
          selectedSubdistrict,
          language,
        )}
      </p>
    </section>
  );
}

function ThailandAddressFreeTextEditor({
  form,
  language,
  setForm,
}: {
  form: FormState;
  language: LanguageCode;
  setForm: (form: FormState) => void;
}) {
  const labels = thailandAddressUi(language);
  const fields = [
    ["contact.province_code", labels.province],
    ["contact.district_code", labels.district],
    ["contact.sub_district_code", labels.subdistrict],
    ["contact.zip_code", labels.postalCode],
  ] as const;
  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm font-semibold md:col-span-2">
      <div className="grid gap-2 md:grid-cols-2">
        {fields.map(([key, label]) => (
          <label className="grid gap-1" key={key}>
            <span>{label}</span>
            <Input
              value={String(form[key] ?? "")}
              onChange={(event) =>
                setForm({ ...form, [key]: event.target.value })
              }
            />
          </label>
        ))}
      </div>
    </section>
  );
}

function ThailandAddressSelect({
  disabled,
  label,
  onChange,
  options,
  placeholder,
  value,
}: {
  disabled?: boolean;
  label: string;
  onChange: (value: string) => void;
  options: Array<{ code: string; label: string }>;
  placeholder: string;
  value: string;
}) {
  return (
    <label className="grid gap-1">
      <span>{label}</span>
      <select
        className="min-h-10 w-full rounded-2xl border border-input bg-background px-3 text-sm text-foreground shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
        disabled={disabled}
        value={value}
        onChange={(event) => onChange(event.target.value)}
      >
        <option value="">{placeholder}</option>
        {options.map((option) => (
          <option key={option.code} value={option.code}>
            {option.label}
          </option>
        ))}
      </select>
    </label>
  );
}

function thailandAddressUi(language: LanguageCode) {
  if (language === "th") {
    return {
      title: "ที่อยู่ประเทศไทย",
      loading: "กำลังโหลดข้อมูลที่อยู่",
      loadError: "โหลดข้อมูลที่อยู่ไม่สำเร็จ",
      provinces: "จังหวัด",
      province: "จังหวัด",
      district: "อำเภอ/เขต",
      subdistrict: "ตำบล/แขวง",
      postalCode: "รหัสไปรษณีย์",
      selectProvince: "เลือกจังหวัด",
      selectDistrict: "เลือกอำเภอ/เขต",
      selectSubdistrict: "เลือกตำบล/แขวง",
    };
  }
  return {
    title: "Thailand address",
    loading: "Loading address data",
    loadError: "Address data failed to load",
    provinces: "provinces",
    province: "Province",
    district: "District",
    subdistrict: "Subdistrict",
    postalCode: "Postal code",
    selectProvince: "Select province",
    selectDistrict: "Select district",
    selectSubdistrict: "Select subdistrict",
  };
}

function filterThailandProvinces(
  provinces: ThailandProvince[],
  matches: ThailandAddressMatch[],
): ThailandProvince[] {
  if (!matches.length) return provinces;
  const allowed = new Set(matches.map((item) => item.provinceCode));
  return provinces.filter((province) => allowed.has(province.code));
}

function filterThailandDistricts(
  districts: ThailandDistrict[],
  matches: ThailandAddressMatch[],
  provinceCode: string,
): ThailandDistrict[] {
  if (!matches.length || !provinceCode) return districts;
  const allowed = new Set(
    matches
      .filter((item) => item.provinceCode === provinceCode)
      .map((item) => item.districtCode),
  );
  return districts.filter((district) => allowed.has(district.code));
}

function filterThailandSubdistricts(
  subdistricts: ThailandSubdistrict[],
  matches: ThailandAddressMatch[],
  provinceCode: string,
  districtCode: string,
): ThailandSubdistrict[] {
  if (!matches.length || !provinceCode || !districtCode) return subdistricts;
  const allowed = new Set(
    matches
      .filter(
        (item) =>
          item.provinceCode === provinceCode &&
          item.districtCode === districtCode,
      )
      .map((item) => item.subdistrictCode),
  );
  return subdistricts.filter((subdistrict) => allowed.has(subdistrict.code));
}

function singleThailandAddressCode<T>(
  rows: T[],
  getCode: (row: T) => string,
): string {
  const codes = onlyUniqueCodes(rows, getCode);
  return codes.length === 1 ? codes[0] : "";
}

function thailandAddressOptionLabel(
  item: ThailandProvince | ThailandDistrict | ThailandSubdistrict,
  language: LanguageCode,
): string {
  return `${thailandAddressLabel(item.name, language)} (${item.code})`;
}

function postalAddressHint(
  postalCode: string,
  matches: ThailandAddressMatch[],
  selectedSubdistrict: ThailandSubdistrict | undefined,
  language: LanguageCode,
): string {
  if (postalCode.length !== 5) {
    return language === "th"
      ? "เลือกจังหวัด > อำเภอ/เขต > ตำบล/แขวง แล้วระบบจะเติมรหัสไปรษณีย์ หรือกรอกรหัสไปรษณีย์ 5 หลักเพื่อกรองตัวเลือก"
      : "Select province > district > subdistrict to fill the postal code, or enter a 5-digit postal code to filter choices.";
  }
  if (!matches.length) {
    return language === "th"
      ? "ไม่พบข้อมูลจากรหัสไปรษณีย์นี้"
      : "No address found for this postal code.";
  }
  if (selectedSubdistrict) {
    return language === "th"
      ? `เลือกแล้ว: ${selectedSubdistrict.name.th} ${postalCode}`
      : `Selected: ${selectedSubdistrict.name.en || selectedSubdistrict.name.th} ${postalCode}`;
  }
  return language === "th"
    ? `พบ ${matches.length} ตำบล/แขวงจากรหัสนี้ เลือกตำบล/แขวงเพื่อยืนยัน`
    : `${matches.length} subdistricts found for this postal code. Select one to confirm.`;
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
      <section
        className="grid max-h-[calc(100dvh-24px)] w-[min(720px,calc(100vw-24px))] gap-3 overflow-hidden rounded-2xl border border-border bg-card p-3 text-foreground shadow-xl"
        role="dialog"
        aria-modal="true"
        aria-label={text("standardUnits")}
      >
        <header className="flex min-w-0 items-center justify-between gap-2">
          <div className="min-w-0">
            <p className="text-xs font-semibold uppercase text-muted-foreground">
              PRODUCT UNIT
            </p>
            <h2 className="truncate text-lg font-semibold">
              {text("standardUnits")}
            </h2>
          </div>
          <Button
            type="button"
            variant="outline"
            size="icon"
            onClick={onClose}
            disabled={dialog.saving}
            aria-label={text("close")}
          >
            <X />
          </Button>
        </header>

        <div className="grid gap-2 sm:grid-cols-[minmax(0,1fr)_auto]">
          <label className="relative block min-w-0">
            <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              className="pl-9"
              placeholder={text("search")}
              value={dialog.query}
              onChange={(event) => onQueryChange(event.target.value)}
              disabled={dialog.loading || dialog.saving}
            />
          </label>
          <Button
            type="button"
            variant="outline"
            onClick={onRefresh}
            disabled={dialog.loading || dialog.saving}
          >
            {dialog.loading ? <Loader2 className="animate-spin" /> : <Search />}
            {text("search")}
          </Button>
        </div>

        <div className="flex flex-wrap items-center justify-between gap-2 text-sm font-semibold">
          <Badge variant="outline">
            {text("selected")}:{" "}
            {dialog.selectedCodes.length.toLocaleString(localeOf(language))}/
            {dialog.options.length.toLocaleString(localeOf(language))}
          </Badge>
          <div className="flex flex-wrap gap-2">
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={onSelectAll}
              disabled={
                dialog.loading || dialog.saving || dialog.options.length === 0
              }
            >
              {text("selectAll")}
            </Button>
            <Button
              type="button"
              variant="outline"
              size="sm"
              onClick={onClear}
              disabled={
                dialog.loading ||
                dialog.saving ||
                dialog.selectedCodes.length === 0
              }
            >
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
          ) : dialog.options.length ? (
            dialog.options.map((unit) => {
              const checked = dialog.selectedCodes.includes(unit.unitcode);
              return (
                <label
                  className="flex min-w-0 cursor-pointer items-center gap-2 rounded-2xl border border-border bg-background px-3 py-2 text-sm font-semibold"
                  key={unit.unitcode}
                >
                  <input
                    className="size-4 shrink-0 accent-primary"
                    type="checkbox"
                    checked={checked}
                    disabled={dialog.saving}
                    onChange={(event) =>
                      onToggle(unit.unitcode, event.target.checked)
                    }
                  />
                  <span className="min-w-0 flex-1 truncate">
                    {unitDisplayName(unit, language)}
                  </span>
                  <b className="shrink-0 text-xs text-muted-foreground">
                    {unit.unitcode}
                  </b>
                </label>
              );
            })
          ) : (
            <div className="grid min-h-32 place-items-center rounded-2xl border border-border bg-background p-4 text-center text-sm font-semibold text-muted-foreground">
              {text("noStandardUnits")}
            </div>
          )}
        </div>

        <footer className="flex flex-wrap justify-end gap-2">
          <Button
            type="button"
            variant="outline"
            onClick={onClose}
            disabled={dialog.saving}
          >
            {text("cancel")}
          </Button>
          <Button
            type="button"
            onClick={onSave}
            disabled={
              dialog.loading ||
              dialog.saving ||
              dialog.selectedCodes.length === 0
            }
          >
            {dialog.saving ? <Loader2 className="animate-spin" /> : <Plus />}
            {text("addSelected")}
          </Button>
        </footer>
      </section>
    </div>
  );
}

type MenuPermissionAction =
  | "access"
  | "create"
  | "update"
  | "delete"
  | "own_only";
type ApprovalTarget = "pr" | "po" | "quotation" | "sale_order";

const menuPermissionActions: {
  key: MenuPermissionAction;
  textKey: keyof typeof uiEn;
}[] = [
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
  readOnly = false,
  setForm,
}: {
  dictionary: BackendLanguageDictionary;
  form: FormState;
  readOnly?: boolean;
  setForm?: (form: FormState) => void;
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

function approvalsFromForm(value: unknown): SettingRecord {
  if (isRecord(value)) return { ...value };
  if (typeof value !== "string" || !value.trim()) return {};
  const parsed = safeJsonParse(value, {});
  return isRecord(parsed) ? { ...parsed } : {};
}

function approvalLabelText(
  dictionary: BackendLanguageDictionary,
  key: string,
  fallback: string,
): string {
  const value = backendText(dictionary, key, fallback);
  return value === key ? fallback : value;
}

function permissionLinkOption(
  record: SettingRecord,
  isApproval: boolean,
): PermissionLinkOption {
  const code = stringValue(
    isApproval ? record.approvalCode : record.permissionCode,
  );
  const name =
    stringValue(isApproval ? record.approvalName : record.permissionName) ||
    code;
  const description = stringValue(record.description);
  return {
    code,
    description,
    isActive: Boolean(record.isActive ?? record.isactive ?? true),
    name,
  };
}

function stringArrayFromForm(value: unknown): string[] {
  if (Array.isArray(value))
    return uniqueStrings(value.map(stringValue).filter(Boolean));
  if (typeof value !== "string") return [];
  const trimmed = value.trim();
  if (!trimmed) return [];
  try {
    const parsed = JSON.parse(trimmed) as unknown;
    return Array.isArray(parsed)
      ? uniqueStrings(parsed.map(stringValue).filter(Boolean))
      : [];
  } catch {
    return uniqueStrings(
      trimmed
        .split(",")
        .map((item) => item.trim())
        .filter(Boolean),
    );
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

function permissionLinkUserOption(
  record: SettingRecord,
): PermissionLinkUserOption {
  const code = stringValue(record.username ?? record.email ?? record.uid);
  const name =
    stringValue(record.name ?? record.email ?? record.username) || code;
  const subtitle = stringValue(record.email) || stringValue(record.uid);
  return {
    code,
    name,
    subtitle,
    isDisabled: Boolean(
      record.is_access_disabled ?? record.is_access_disabled ?? false,
    ),
  };
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
  const fallbackLabel =
    field.label[language] ?? field.label.en ?? field.label.th;
  const labelKey =
    fieldBackendKeys[`permission_link.${field.key}`] ?? field.key;
  const translatedLabel = backendText(dictionary, labelKey, fallbackLabel);
  const label =
    translatedLabel === labelKey || translatedLabel === field.key
      ? fallbackLabel
      : translatedLabel;
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
        const response = await fetch(
          `/api/system-settings/${userConfig.slug}?${searchParams.toString()}`,
          {
            headers: requestHeaders(auth),
            cache: "no-store",
          },
        );
        const payload = (await response.json()) as unknown;
        if (!response.ok || isFailed(payload))
          throw new Error(
            extractMessage(payload) ??
              backendText(dictionary, "request_failed", "Request failed."),
          );
        if (cancelled) return;
        setUsers(
          normalizeRecords(payload, userConfig)
            .map(permissionLinkUserOption)
            .filter((user) => user.code),
        );
      } catch (loadError) {
        if (!cancelled)
          setError(
            loadError instanceof Error && loadError.message
              ? loadError.message
              : backendText(dictionary, "request_failed", "Request failed."),
          );
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
        <span>
          {label}
          {field.required ? " *" : ""}
        </span>
        {selectedCode ? (
          <Badge variant="success">{selectedCode}</Badge>
        ) : (
          <Badge variant="outline">
            {backendText(dictionary, "select_user", "Select user")}
          </Badge>
        )}
      </div>
      <Input
        value={query}
        onChange={(event) => setQuery(event.target.value)}
        placeholder={backendText(dictionary, "search", "Search")}
      />
      {selectedCode ? (
        <div className="grid gap-1 rounded-xl border border-primary/40 bg-primary/10 p-2 text-primary">
          <span className="truncate font-semibold">
            {selectedName || selectedCode}
          </span>
          <span className="truncate text-xs">code: {selectedCode}</span>
        </div>
      ) : null}
      {error ? (
        <p className="rounded-xl border border-destructive/40 bg-destructive/10 px-2 py-1 text-xs text-destructive">
          {error}
        </p>
      ) : null}
      <div className="grid max-h-60 gap-2 overflow-y-auto pr-1 md:grid-cols-2">
        {!canSearch ? (
          <div className="rounded-xl border border-border bg-card p-3 text-sm text-muted-foreground md:col-span-2">
            {backendText(
              dictionary,
              "search_user_hint",
              "Type at least 2 characters to search users.",
            )}
          </div>
        ) : loading ? (
          <div className="flex min-h-16 items-center gap-2 rounded-xl border border-border bg-card p-3 text-muted-foreground md:col-span-2">
            <Loader2 className="animate-spin" />
            {backendText(dictionary, "loading", "Loading data")}
          </div>
        ) : users.length ? (
          users.map((user) => {
            const checked = selectedCode === user.code;
            return (
              <button
                className={cn(
                  "grid min-h-16 gap-1 rounded-xl border p-2 text-left transition-colors",
                  checked
                    ? "border-primary bg-primary/10 text-primary"
                    : "border-border bg-card text-foreground hover:bg-muted/60",
                  user.isDisabled && "cursor-not-allowed opacity-50",
                )}
                disabled={user.isDisabled}
                key={user.code}
                onClick={() => choose(user)}
                type="button"
              >
                <span className="flex min-w-0 items-center gap-2">
                  {checked ? (
                    <Check className="size-4 shrink-0" />
                  ) : (
                    <UserRound className="size-4 shrink-0 text-muted-foreground" />
                  )}
                  <span className="min-w-0 truncate font-semibold">
                    {user.name}
                  </span>
                </span>
                <span className="truncate text-xs text-muted-foreground">
                  code: {user.code}
                </span>
                {user.subtitle ? (
                  <span className="truncate text-xs font-normal text-muted-foreground">
                    {user.subtitle}
                  </span>
                ) : null}
              </button>
            );
          })
        ) : (
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
  config,
  auth,
  dictionary,
  field,
  form,
  language,
  readOnly = false,
  setForm,
  workspace,
}: {
  config?: SystemSettingConfig;
  auth: AuthSession | null;
  dictionary: BackendLanguageDictionary;
  field: SystemSettingField;
  form: FormState;
  language: LanguageCode;
  readOnly?: boolean;
  setForm?: (form: FormState) => void;
  workspace: WorkspaceSession | null;
}) {
  const isApproval = field.key === "approvalCodes";
  const sourceConfig = useMemo(
    () =>
      getSystemSettingConfig(
        isApproval ? "approval_setting" : "permission_definition",
      ),
    [isApproval],
  );
  const [options, setOptions] = useState<PermissionLinkOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const selectedCodes = stringArrayFromForm(form[field.key]);
  const isGroup = config?.slug === "permission_group";
  const selectedUserCode = isGroup ? "group" : stringValue(form.employeeCode);
  const fallbackLabel =
    field.label[language] ?? field.label.en ?? field.label.th;
  const labelKey =
    fieldBackendKeys[`permission_link.${field.key}`] ?? field.key;
  const translatedLabel = backendText(dictionary, labelKey, fallbackLabel);
  const label =
    translatedLabel === labelKey || translatedLabel === field.key
      ? fallbackLabel
      : translatedLabel;

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
        const response = await fetch(
          `/api/system-settings/${sourceConfig.slug}?${searchParams.toString()}`,
          {
            headers: requestHeaders(auth),
            cache: "no-store",
          },
        );
        const payload = (await response.json()) as unknown;
        if (!response.ok || isFailed(payload))
          throw new Error(
            extractMessage(payload) ??
              backendText(dictionary, "request_failed", "Request failed."),
          );
        if (cancelled) return;
        setOptions(
          normalizeRecords(payload, sourceConfig)
            .map((record) => permissionLinkOption(record, isApproval))
            .filter((option) => option.code),
        );
      } catch (loadError) {
        if (!cancelled)
          setError(
            loadError instanceof Error && loadError.message
              ? loadError.message
              : backendText(dictionary, "request_failed", "Request failed."),
          );
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
    if (readOnly || !setForm || !selectedUserCode) return;
    const next = checked
      ? uniqueStrings([...selectedCodes, code])
      : selectedCodes.filter((item) => item !== code);
    setForm({ ...form, [field.key]: next });
  }

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-3 text-sm font-semibold md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span>
          {label}
          {field.required ? " *" : ""}
        </span>
        <Badge variant="outline">
          {selectedCodes.length.toLocaleString(localeOf(language))}
        </Badge>
      </div>
      {!selectedUserCode ? (
        <p className="rounded-xl border border-amber-300 bg-amber-50 px-2 py-1 text-xs text-amber-800 dark:border-amber-900 dark:bg-amber-950/30 dark:text-amber-200">
          {backendText(dictionary, "select_user_first", "Select user first")}
        </p>
      ) : null}
      {error ? (
        <p className="rounded-xl border border-destructive/40 bg-destructive/10 px-2 py-1 text-xs text-destructive">
          {error}
        </p>
      ) : null}
      <div className="grid gap-2 md:grid-cols-2">
        {loading ? (
          <div className="flex min-h-20 items-center gap-2 rounded-xl border border-border bg-card p-3 text-muted-foreground md:col-span-2">
            <Loader2 className="animate-spin" />
            {backendText(dictionary, "loading", "Loading data")}
          </div>
        ) : options.length ? (
          options.map((option) => {
            const checked = selectedCodes.includes(option.code);
            return (
              <label
                className={cn(
                  "grid cursor-pointer gap-1 rounded-xl border p-2 transition-colors",
                  checked
                    ? "border-primary bg-primary/10 text-primary"
                    : "border-border bg-card text-foreground hover:bg-muted/60",
                  !option.isActive && "opacity-60",
                )}
                key={option.code}
              >
                <span className="flex min-w-0 items-center gap-2">
                  <input
                    className="size-4 shrink-0 accent-primary"
                    type="checkbox"
                    checked={checked}
                    disabled={readOnly || !selectedUserCode}
                    onChange={(event) =>
                      toggle(option.code, event.target.checked)
                    }
                  />
                  <span className="min-w-0 truncate font-semibold">
                    {option.name || option.code}
                  </span>
                </span>
                <span className="truncate text-xs text-muted-foreground">
                  code: {option.code}
                </span>
                {option.description ? (
                  <span className="line-clamp-2 text-xs font-normal text-muted-foreground">
                    {option.description}
                  </span>
                ) : null}
              </label>
            );
          })
        ) : (
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
  readOnly = false,
  setForm,
  text,
}: {
  dateTimeScope: DateTimeScope;
  dictionary: BackendLanguageDictionary;
  form: FormState;
  language: LanguageCode;
  readOnly?: boolean;
  setForm?: (form: FormState) => void;
  text: (key: keyof typeof uiEn) => string;
}) {
  const branchKey = dateTimeScope.key || "company";
  const branches = permissionBranchesFromForm(form.branches);
  const branchPermission = permissionBranchValue(branches, branchKey);
  const menus = isRecord(branchPermission.menus) ? branchPermission.menus : {};

  function updateMenuPermission(
    menuId: string,
    action: MenuPermissionAction,
    checked: boolean,
  ) {
    if (readOnly || !setForm) return;
    const nextBranches = permissionBranchesFromForm(form.branches);
    const nextBranch = permissionBranchValue(nextBranches, branchKey);
    const nextMenus = isRecord(nextBranch.menus) ? { ...nextBranch.menus } : {};
    const currentMenu = isRecord(nextMenus[menuId])
      ? { ...nextMenus[menuId] }
      : {};
    currentMenu[action] = checked;
    nextMenus[menuId] = currentMenu;
    nextBranches[branchKey] = {
      ...nextBranch,
      ...dateTimeScopePayload(dateTimeScope),
      menus: nextMenus,
    };
    setForm({ ...form, branches: nextBranches });
  }

  return (
    <section className="grid gap-3 rounded-2xl border border-border bg-background p-3 md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div>
          <h3 className="text-base font-semibold">
            {backendText(
              dictionary,
              "permission_definition",
              "Permission Definition",
            )}
          </h3>
          <p className="text-xs text-muted-foreground">
            {text("branch")}:{" "}
            {dateTimeScope.branchcode || dateTimeScope.branchguid || branchKey}
          </p>
        </div>
        <Badge variant="outline">
          {
            MENU_SECTIONS.flatMap((section) =>
              section.groups.flatMap((group) => group.items),
            ).length
          }{" "}
          menu codes
        </Badge>
      </div>

      <div className="grid gap-3">
        {MENU_SECTIONS.map((section) => (
          <section
            className="grid gap-2 rounded-2xl border border-border bg-card p-2"
            key={section.id}
          >
            <h4 className="text-sm font-semibold">
              {menuText(section.title, language, dictionary)}
            </h4>
            {section.groups.map((group) => (
              <div
                className="grid gap-1 rounded-xl border border-border bg-background p-2"
                key={group.id}
              >
                <p className="text-xs font-semibold text-muted-foreground">
                  {menuText(group.title, language, dictionary)}
                </p>
                {group.items.map((item) => {
                  const permission: SettingRecord = isRecord(menus[item.id])
                    ? (menus[item.id] as SettingRecord)
                    : {};
                  return (
                    <div
                      className="grid gap-2 rounded-xl border border-border bg-card px-3 py-2.5 xl:grid-cols-[minmax(180px,1fr)_auto] items-center shadow-[0_1px_2px_rgba(0,0,0,0.01)] hover:bg-secondary/5 transition-colors duration-150"
                      key={item.id}
                    >
                      <div className="min-w-0">
                        <p className="truncate text-sm font-bold text-foreground">
                          {menuText(item.label, language, dictionary)}
                        </p>
                        <p className="truncate text-[10px] text-muted-foreground mt-0.5">
                          <span className="font-mono bg-secondary/35 px-1 py-0.5 rounded text-primary">
                            code: {item.id}
                          </span>
                        </p>
                      </div>
                      <div className="grid grid-cols-2 gap-1.5 sm:grid-cols-5 md:gap-2">
                        {menuPermissionActions.map((action) => (
                          <label
                            className="inline-flex h-8 w-auto max-w-none items-center gap-1.5 rounded-lg border border-border/80 bg-background/50 hover:bg-secondary/5 px-2 text-xs font-semibold cursor-pointer transition-colors shadow-sm"
                            key={action.key}
                          >
                            <input
                              className="size-3.5 accent-primary shrink-0"
                              type="checkbox"
                              checked={Boolean(permission[action.key])}
                              disabled={readOnly}
                              onChange={(event) =>
                                updateMenuPermission(
                                  item.id,
                                  action.key,
                                  event.target.checked,
                                )
                              }
                            />
                            <span className="truncate text-foreground/85">
                              {permissionActionText(
                                dictionary,
                                language,
                                action.textKey,
                              )}
                            </span>
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

function permissionBranchValue(
  branches: SettingRecord,
  branchKey: string,
): SettingRecord {
  const branch = branches[branchKey];
  return isRecord(branch) ? { ...branch } : {};
}

function permissionActionText(
  dictionary: BackendLanguageDictionary,
  language: LanguageCode,
  key: keyof typeof uiEn,
): string {
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
  const helper =
    field.helper?.[language] ?? field.helper?.en ?? field.helper?.th;
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
    return (
      <ApprovalSettingEditor
        dictionary={dictionary}
        form={form}
        setForm={setForm}
      />
    );
  }

  if (config.slug === "permission_link" && field.key === "employeeCode") {
    return (
      <label className="grid gap-1 text-sm font-semibold">
        <span>{label}</span>
        <Input value={String(value ?? "")} readOnly disabled aria-readonly />
      </label>
    );
  }

  if (config.slug === "permission_link" && field.key === "employeeName") {
    return (
      <label className="grid gap-1 text-sm font-semibold">
        <span>{label}</span>
        <Input value={String(value ?? "")} readOnly disabled aria-readonly />
      </label>
    );
  }

  if (
    (config.slug === "permission_link" || config.slug === "permission_group") &&
    (field.key === "permissionCodes" || field.key === "approvalCodes")
  ) {
    return (
      <PermissionLinkMultiSelectEditor
        config={config}
        auth={auth}
        dictionary={dictionary}
        field={field}
        form={form}
        language={language}
        setForm={setForm}
        workspace={workspace}
      />
    );
  }

  if (config.slug === "product_warehouse_screen" && field.key === "location") {
    return (
      <div className="md:col-span-2">
        <label className="text-sm font-semibold mb-1 block">{label}</label>
        <WarehouseLocationsEditor
          form={form}
          setForm={setForm}
          language={language}
          workspace={workspace}
        />
      </div>
    );
  }

  if (isThailandAddressPrimaryField(config, field)) {
    return (
      <ThailandAddressFieldEditor
        backendUrl={auth?.backendUrl}
        form={form}
        language={language}
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
          className="flex w-full flex-wrap gap-2"
          role="radiogroup"
          aria-label={label}
        >
          {options.map((option) => {
            const checked = selectedValue === option.value;
            return (
              <label
                className={cn(
                  "flex min-h-10 flex-1 basis-24 cursor-pointer items-center gap-2 rounded-2xl border px-3 py-2 transition-colors",
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
                <span>{optionLabel(option, language)}</span>
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
          placeholder={field.type === "json" ? "[]" : field.placeholder}
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
    // Hidden because it is integrated into CompanyBranchTreeSelector
    return null;
  }

  if (field.type === "company-multi-select") {
    return (
      <CompanyBranchTreeSelector
        auth={auth}
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
          nextForm["settings.languageconfigs"],
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
          dateTimeScope.timezone_label || dateTimeScope.timezone_offset
        }
        value={String(value ?? "")}
        onChange={(event) =>
          setForm({ ...form, [field.key]: event.target.value })
        }
      />
    );
  }

  if (
    config.slug === "product_category_group_select_screen" &&
    field.key === "colorselecthex"
  ) {
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

  const emailLockedByLoginCode =
    config.slug === "user" &&
    field.key === "email" &&
    isEmailLike(form.username);

  return (
    <label className="grid gap-1 text-sm font-semibold">
      <span>
        {label}
        {field.required ? " *" : ""}
      </span>
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
      {helper ? (
        <span className="text-xs font-medium leading-snug text-muted-foreground">
          {helper}
        </span>
      ) : null}
    </label>
  );
}

function isBranchStructuredSettingField(
  config: SystemSettingConfig,
  field: SystemSettingField,
): boolean {
  return (
    config.slug === "branch" &&
    field.type === "json" &&
    (field.key === "paymentrounding" || field.key === "pointconfig")
  );
}

type TimeSaleRow = {
  daysofweek: number[];
  fromdate: string;
  todate: string;
  fromtime: string;
  totime: string;
};

function TimeSaleListEditor({
  dateTimeScope,
  field,
  form,
  label,
  language,
  setForm,
}: {
  dateTimeScope: DateTimeScope;
  field: SystemSettingField;
  form: FormState;
  label: string;
  language: LanguageCode;
  setForm: (form: FormState) => void;
}) {
  const rows = normalizeTimeSaleFormList(form[field.key]);
  const enabled = rows.length > 0;

  function commit(nextRows: TimeSaleRow[]) {
    setForm({ ...form, [field.key]: nextRows });
  }

  function updateRow(index: number, patch: Partial<TimeSaleRow>) {
    commit(rows.map((row, rowIndex) => (rowIndex === index ? { ...row, ...patch } : row)));
  }

  function toggleEnabled(checked: boolean) {
    commit(checked ? [emptyTimeSaleRow()] : []);
  }

  function toggleDay(index: number, day: number, checked: boolean) {
    const current = new Set(rows[index]?.daysofweek ?? []);
    if (checked) current.add(day);
    else current.delete(day);
    updateRow(index, { daysofweek: Array.from(current).sort((a, b) => a - b) });
  }

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm font-semibold">
      <label className="flex min-h-10 items-center gap-2">
        <input
          className="size-4 accent-primary"
          type="checkbox"
          checked={enabled}
          onChange={(event) => toggleEnabled(event.target.checked)}
        />
        <span>
          {label}
          {enabled ? ` (${rows.length})` : ""}
        </span>
      </label>
      {enabled ? (
        <div className="grid gap-2">
          {rows.map((row, index) => (
            <div
              className="grid gap-2 rounded-xl border border-border bg-card p-2"
              key={index}
            >
              <div className="flex flex-wrap items-center justify-between gap-2">
                <b>
                  {language === "th" ? "ช่วงเวลา" : "Time window"} {index + 1}
                </b>
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="h-8"
                  onClick={() => commit(rows.filter((_, rowIndex) => rowIndex !== index))}
                >
                  <Trash2 className="size-3.5" />
                  {language === "th" ? "ลบ" : "Delete"}
                </Button>
              </div>
              <div className="grid gap-2 md:grid-cols-2">
                <DateField
                  language={language}
                  yearType={dateTimeScope.calendarYearType}
                  label={language === "th" ? "วันที่เริ่ม" : "From date"}
                  value={row.fromdate}
                  onChange={(event) => updateRow(index, { fromdate: event.target.value })}
                />
                <DateField
                  language={language}
                  yearType={dateTimeScope.calendarYearType}
                  label={language === "th" ? "วันที่สิ้นสุด" : "To date"}
                  value={row.todate}
                  onChange={(event) => updateRow(index, { todate: event.target.value })}
                />
                <TimeField
                  label={language === "th" ? "เวลาเริ่ม" : "From time"}
                  value={row.fromtime}
                  onChange={(event) => updateRow(index, { fromtime: normalizeTimeInput(event.target.value) })}
                />
                <TimeField
                  label={language === "th" ? "เวลาสิ้นสุด" : "To time"}
                  value={row.totime}
                  onChange={(event) => updateRow(index, { totime: normalizeTimeInput(event.target.value) })}
                />
              </div>
              <div className="flex flex-wrap gap-1">
                {dayNames[language].map((dayName, dayIndex) => {
                  const day = dayIndex + 1;
                  const checked = row.daysofweek.includes(day);
                  return (
                    <label
                      className={cn(
                        "inline-flex min-h-8 cursor-pointer items-center gap-1 rounded-lg border px-2 text-xs font-semibold",
                        checked
                          ? "border-primary bg-primary/10 text-primary"
                          : "border-border bg-background text-muted-foreground",
                      )}
                      key={day}
                    >
                      <input
                        className="size-3 accent-primary"
                        type="checkbox"
                        checked={checked}
                        onChange={(event) => toggleDay(index, day, event.target.checked)}
                      />
                      {dayName}
                    </label>
                  );
                })}
              </div>
            </div>
          ))}
          <Button
            type="button"
            variant="outline"
            className="h-9 justify-center"
            onClick={() => commit([...rows, emptyTimeSaleRow()])}
          >
            <Plus className="size-4" />
            {language === "th" ? "เพิ่มช่วงเวลา" : "Add time window"}
          </Button>
        </div>
      ) : null}
    </section>
  );
}

function TimeSaleListReadOnlyDetail({
  label,
  language,
  value,
}: {
  label: string;
  language: LanguageCode;
  value: unknown;
}) {
  const rows = normalizeTimeSaleFormList(value);
  return (
    <div className="grid gap-2 rounded-xl border border-border bg-card px-3 py-2 text-sm shadow-[0_1px_2px_rgba(0,0,0,0.02)]">
      <span className="text-xs font-semibold text-muted-foreground">
        {label}
      </span>
      {rows.length === 0 ? (
        <b className="text-foreground">-</b>
      ) : (
        <div className="grid gap-1">
          {rows.map((row, index) => (
            <div className="rounded-lg border border-border bg-background p-2" key={index}>
              <b>
                {language === "th" ? "ช่วงเวลา" : "Time window"} {index + 1}
              </b>
              <p className="text-xs font-medium text-muted-foreground">
                {row.fromdate || "-"} {row.fromtime || "--:--"} - {row.todate || "-"} {row.totime || "--:--"}
              </p>
              <p className="text-xs font-medium text-muted-foreground">
                {(row.daysofweek.length ? row.daysofweek : [1, 2, 3, 4, 5, 6, 7])
                  .map((day) => dayNames[language][day - 1])
                  .filter(Boolean)
                  .join(", ")}
              </p>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}

function emptyTimeSaleRow(): TimeSaleRow {
  return {
    daysofweek: [1, 2, 3, 4, 5, 6, 7],
    fromdate: "",
    todate: "",
    fromtime: "",
    totime: "",
  };
}

function normalizeTimeSaleFormList(value: unknown): TimeSaleRow[] {
  if (!Array.isArray(value)) return [];
  return value.filter(isRecord).map((item) => ({
    daysofweek: normalizeTimeSaleDays(item.daysofweek),
    fromdate: timeSaleDateInputValue(item.fromdate),
    todate: timeSaleDateInputValue(item.todate),
    fromtime: normalizeTimeInput(stringValue(item.fromtime)),
    totime: normalizeTimeInput(stringValue(item.totime)),
  }));
}

function normalizeTimeSalePayload(value: unknown): TimeSaleRow[] {
  return normalizeTimeSaleFormList(value).map((item) => ({
    daysofweek: item.daysofweek,
    fromdate: timeSaleDateInputToIso(item.fromdate),
    todate: timeSaleDateInputToIso(item.todate),
    fromtime: item.fromtime,
    totime: item.totime,
  }));
}

function normalizeTimeSaleDays(value: unknown): number[] {
  const values = Array.isArray(value) ? value : [];
  const normalized = values
    .map((item) => Number(item))
    .filter((item) => Number.isInteger(item) && item >= 1 && item <= 7);
  return Array.from(new Set(normalized)).sort((a, b) => a - b);
}

function timeSaleDateInputValue(value: unknown): string {
  const raw = stringValue(value);
  if (!raw) return "";
  const match = raw.match(/^(\d{4}-\d{2}-\d{2})/);
  return match?.[1] ?? "";
}

function timeSaleDateInputToIso(value: unknown): string {
  const raw = timeSaleDateInputValue(value);
  if (!raw) return "";
  return new Date(`${raw}T00:00:00.000Z`).toISOString();
}

function normalizeHexColor(value: unknown): string {
  const raw = stringValue(value).replace(/^#/, "");
  if (/^[0-9a-fA-F]{6}$/.test(raw)) return `#${raw}`;
  if (/^[0-9a-fA-F]{8}$/.test(raw)) return `#${raw.slice(2)}`;
  return "#ffffff";
}

const paymentMethodKeys = [
  "cash",
  "creditcard",
  "banktransfer",
  "cheque",
  "coupon",
  "delivery",
  "qrcode",
] as const;

type PaymentMethodKey = (typeof paymentMethodKeys)[number];
type RoundingRule = {
  lowerbound: number;
  upperbound: number;
  roundto: number;
};
type PaymentMethodRounding = {
  enabled: boolean;
  rules: RoundingRule[];
};
type PaymentRoundingConfig = Record<PaymentMethodKey, PaymentMethodRounding>;
type GeneralPointRule = {
  startdate: string;
  enddate: string;
  payperpoint: number;
  pointvalue: number;
};
type SpecialPointRule = {
  startdate: string;
  enddate: string;
  multiplier: number;
  sunday: boolean;
  monday: boolean;
  tuesday: boolean;
  wednesday: boolean;
  thursday: boolean;
  friday: boolean;
  saturday: boolean;
  maxpointperbill: number;
};
type PointConfig = {
  generalrules: GeneralPointRule[];
  specialrules: SpecialPointRule[];
  pointusagetype: number;
};
type BranchStructuredParseResult<T> = {
  isRecovered: boolean;
  value: T;
};

const pointWeekDays = [
  "monday",
  "tuesday",
  "wednesday",
  "thursday",
  "friday",
  "saturday",
  "sunday",
] as const;

type PointWeekDay = (typeof pointWeekDays)[number];

function BranchStructuredSettingEditor({
  dictionary,
  field,
  form,
  label,
  language,
  readOnly = false,
  setForm,
}: {
  dictionary: BackendLanguageDictionary;
  field: SystemSettingField;
  form: FormState;
  label: string;
  language: LanguageCode;
  readOnly?: boolean;
  setForm?: (form: FormState) => void;
}) {
  const commit = (nextValue: unknown) => {
    if (readOnly || !setForm) return;
    setForm({ ...form, [field.key]: JSON.stringify(nextValue, null, 2) });
  };

  if (field.key === "paymentrounding") {
    const parsed = normalizePaymentRoundingConfig(form[field.key]);
    return (
      <PaymentRoundingTableEditor
  dictionary={dictionary}
  isRecovered={parsed.isRecovered}
      label={label}
      language={language}
      onChange={commit}
      readOnly={readOnly}
      value={parsed.value}
    />
  );
}

  const parsed = normalizePointConfig(form[field.key]);
  return (
    <PointConfigTableEditor
      dictionary={dictionary}
      isRecovered={parsed.isRecovered}
    label={label}
    language={language}
    onChange={commit}
    readOnly={readOnly}
    value={parsed.value}
  />
);
}

function PaymentRoundingTableEditor({
  dictionary,
  isRecovered,
  label,
  language,
  onChange,
  readOnly = false,
  value,
}: {
  dictionary: BackendLanguageDictionary;
  isRecovered: boolean;
  label: string;
  language: LanguageCode;
  onChange: (value: PaymentRoundingConfig) => void;
  readOnly?: boolean;
  value: PaymentRoundingConfig;
}) {
  const text = (key: string, fallback: string) =>
    backendText(dictionary, key, fallback);

  function updateMethod(
    method: PaymentMethodKey,
    patch: Partial<PaymentMethodRounding>,
  ) {
    if (readOnly) return;
    const current = value[method];
    onChange({
      ...value,
      [method]: {
        ...current,
        ...patch,
        rules: patch.rules ?? current.rules,
      },
    });
  }

  function updateRule(
    method: PaymentMethodKey,
    index: number,
    patch: Partial<RoundingRule>,
  ) {
    if (readOnly) return;
    const current = value[method];
    updateMethod(method, {
      rules: current.rules.map((rule, ruleIndex) =>
        ruleIndex === index ? { ...rule, ...patch } : rule,
      ),
    });
  }

  function addRule(method: PaymentMethodKey) {
    if (readOnly) return;
    updateMethod(method, {
      enabled: true,
      rules: [...value[method].rules, defaultRoundingRule()],
    });
  }

  function removeRule(method: PaymentMethodKey, index: number) {
    if (readOnly) return;
    const current = value[method];
    if (current.rules.length <= 1) return;
    updateMethod(method, {
      rules: current.rules.filter((_, ruleIndex) => ruleIndex !== index),
    });
  }

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <span className="font-semibold">{label}</span>
            {isRecovered ? <Badge variant="warning">{invalidConfigText(language)}</Badge> : null}
          </div>
          <p className="mt-1 text-xs font-medium text-muted-foreground">
            {paymentRoundingSummary(value, language)}
          </p>
        </div>
      </div>
      <div className="overflow-x-auto rounded-xl border border-border">
        <table className="w-full min-w-[760px] border-collapse text-left text-xs">
          <thead className="bg-muted/60 text-muted-foreground">
            <tr>
              <th className="w-52 px-2 py-2 font-semibold">{text("payment_method", "ช่องทาง")}</th>
              <th className="px-2 py-2 font-semibold">{text("lower_bound", "ต่ำสุด")}</th>
              <th className="px-2 py-2 font-semibold">{text("upper_bound", "สูงสุด")}</th>
              <th className="px-2 py-2 font-semibold">{text("round_to", "ปัดเป็น")}</th>
              {!readOnly ? (
                <th className="w-36 px-2 py-2 text-right font-semibold">{text("manage", "จัดการ")}</th>
              ) : null}
            </tr>
          </thead>
          <tbody>
            {paymentMethodKeys.map((method) => {
              const methodConfig = value[method];
              if (!methodConfig.enabled) {
                return (
                  <tr className="border-t border-border" key={method}>
                    <td className="px-2 py-2 align-top">
                      <PaymentMethodToggle
                        checked={methodConfig.enabled}
                        label={paymentMethodLabel(method, dictionary)}
                        disabled={readOnly}
                        onChange={(checked) =>
                          updateMethod(method, {
                            enabled: checked,
                            rules: methodConfig.rules.length
                              ? methodConfig.rules
                              : defaultRoundingRules(),
                          })
                        }
                        text={text}
                      />
                    </td>
                    <td className="px-2 py-2 text-muted-foreground" colSpan={3}>
                      {language === "th" ? "ปิดการปัดเศษ" : "Rounding disabled"}
                    </td>
                    {!readOnly ? (
                      <td className="px-2 py-2 text-right">
                        <Button type="button" size="sm" variant="outline" onClick={() => addRule(method)}>
                          <Plus />
                          {text("add_rule", "เพิ่มกฎ")}
                        </Button>
                      </td>
                    ) : null}
                  </tr>
                );
              }
              return (
                <FragmentLikeRows key={method}>
                  {methodConfig.rules.map((rule, index) => (
                    <tr className="border-t border-border" key={`${method}-${index}`}>
                      <td className="px-2 py-2 align-top">
                        {index === 0 ? (
                          <PaymentMethodToggle
                            checked={methodConfig.enabled}
                            label={paymentMethodLabel(method, dictionary)}
                            disabled={readOnly}
                            onChange={(checked) => updateMethod(method, { enabled: checked })}
                            text={text}
                          />
                        ) : null}
                      </td>
                      <td className="px-2 py-1">
                        <CompactNumberInput
                          disabled={readOnly}
                          value={rule.lowerbound}
                          onChange={(nextValue) => updateRule(method, index, { lowerbound: nextValue })}
                        />
                      </td>
                      <td className="px-2 py-1">
                        <CompactNumberInput
                          disabled={readOnly}
                          value={rule.upperbound}
                          onChange={(nextValue) => updateRule(method, index, { upperbound: nextValue })}
                        />
                      </td>
                      <td className="px-2 py-1">
                        <CompactNumberInput
                          disabled={readOnly}
                          value={rule.roundto}
                          onChange={(nextValue) => updateRule(method, index, { roundto: nextValue })}
                        />
                      </td>
                      {!readOnly ? (
                        <td className="px-2 py-1">
                          <div className="flex justify-end gap-1">
                            {index === 0 ? (
                              <Button type="button" size="sm" variant="outline" onClick={() => addRule(method)}>
                                <Plus />
                              </Button>
                            ) : null}
                            <Button
                              type="button"
                              size="sm"
                              variant="outline"
                              disabled={methodConfig.rules.length <= 1}
                              onClick={() => removeRule(method, index)}
                            >
                              <Trash2 />
                            </Button>
                          </div>
                        </td>
                      ) : null}
                    </tr>
                  ))}
                </FragmentLikeRows>
              );
            })}
          </tbody>
        </table>
      </div>
    </section>
  );
}

function PointConfigTableEditor({
  dictionary,
  isRecovered,
  label,
  language,
  onChange,
  readOnly = false,
  value,
}: {
  dictionary: BackendLanguageDictionary;
  isRecovered: boolean;
  label: string;
  language: LanguageCode;
  onChange: (value: PointConfig) => void;
  readOnly?: boolean;
  value: PointConfig;
}) {
  const text = (key: string, fallback: string) =>
    backendText(dictionary, key, fallback);

  function updateGeneralRule(index: number, patch: Partial<GeneralPointRule>) {
    if (readOnly) return;
    onChange({
      ...value,
      generalrules: value.generalrules.map((rule, ruleIndex) =>
        ruleIndex === index ? { ...rule, ...patch } : rule,
      ),
    });
  }

  function updateSpecialRule(index: number, patch: Partial<SpecialPointRule>) {
    if (readOnly) return;
    onChange({
      ...value,
      specialrules: value.specialrules.map((rule, ruleIndex) =>
        ruleIndex === index ? { ...rule, ...patch } : rule,
      ),
    });
  }

  return (
    <section className="grid gap-3 rounded-2xl border border-border bg-background p-2 text-sm md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="min-w-0">
          <div className="flex flex-wrap items-center gap-2">
            <span className="font-semibold">{label}</span>
            {isRecovered ? <Badge variant="warning">{invalidConfigText(language)}</Badge> : null}
          </div>
          <p className="mt-1 text-xs font-medium text-muted-foreground">
            {pointConfigSummary(value, language)}
          </p>
        </div>
        <div className="flex flex-wrap gap-1">
          {[1, 2].map((usageType) => (
            <Button
              className={value.pointusagetype === usageType ? "" : "bg-background text-foreground"}
              key={usageType}
              size="sm"
              type="button"
              disabled={readOnly}
              variant={value.pointusagetype === usageType ? "default" : "outline"}
              onClick={() => {
                if (!readOnly) onChange({ ...value, pointusagetype: usageType });
              }}
            >
              {usageType === 1
                ? text("discount", "ส่วนลด")
                : text("cash", "เงินสด")}
            </Button>
          ))}
        </div>
      </div>

      <PointGeneralRulesTable
        dictionary={dictionary}
        language={language}
        onAdd={() =>
          !readOnly
            ? onChange({
                ...value,
                generalrules: [...value.generalrules, defaultGeneralPointRule()],
              })
            : undefined
        }
        onRemove={(index) =>
          !readOnly
            ? onChange({
                ...value,
                generalrules: value.generalrules.filter((_, ruleIndex) => ruleIndex !== index),
              })
            : undefined
        }
        onUpdate={updateGeneralRule}
        readOnly={readOnly}
        rules={value.generalrules}
      />

      <PointSpecialRulesTable
        dictionary={dictionary}
        onAdd={() =>
          !readOnly
            ? onChange({
                ...value,
                specialrules: [...value.specialrules, defaultSpecialPointRule()],
              })
            : undefined
        }
        onRemove={(index) =>
          !readOnly
            ? onChange({
                ...value,
                specialrules: value.specialrules.filter((_, ruleIndex) => ruleIndex !== index),
              })
            : undefined
        }
        onUpdate={updateSpecialRule}
        readOnly={readOnly}
        rules={value.specialrules}
      />
    </section>
  );
}

function PointGeneralRulesTable({
  dictionary,
  language,
  onAdd,
  onRemove,
  onUpdate,
  readOnly = false,
  rules,
}: {
  dictionary: BackendLanguageDictionary;
  language: LanguageCode;
  onAdd: () => void;
  onRemove: (index: number) => void;
  onUpdate: (index: number, patch: Partial<GeneralPointRule>) => void;
  readOnly?: boolean;
  rules: GeneralPointRule[];
}) {
  const text = (key: string, fallback: string) =>
    backendText(dictionary, key, fallback);
  return (
    <div className="grid gap-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h4 className="font-semibold">{text("general_rules", "กฎทั่วไป")}</h4>
        {!readOnly ? (
          <Button type="button" size="sm" onClick={onAdd}>
            <Plus />
            {text("add_rule", "เพิ่มกฎ")}
          </Button>
        ) : null}
      </div>
      <div className="overflow-x-auto rounded-xl border border-border">
        <table className="w-full min-w-[720px] border-collapse text-left text-xs">
          <thead className="bg-muted/60 text-muted-foreground">
            <tr>
              <th className="px-2 py-2 font-semibold">#</th>
              <th className="px-2 py-2 font-semibold">{text("start_date", "วันที่เริ่มต้น")}</th>
              <th className="px-2 py-2 font-semibold">{text("end_date", "วันที่สิ้นสุด")}</th>
              <th className="px-2 py-2 font-semibold">{text("amount_per_point", "จำนวนเงินต่อ 1 แต้ม")}</th>
              <th className="px-2 py-2 font-semibold">{text("points_per_baht", "แต้มต่อบาท")}</th>
              {!readOnly ? (
                <th className="w-20 px-2 py-2 text-right font-semibold">{text("manage", "จัดการ")}</th>
              ) : null}
            </tr>
          </thead>
          <tbody>
            {rules.length === 0 ? (
              <tr className="border-t border-border">
                <td className="px-2 py-3 text-muted-foreground" colSpan={readOnly ? 5 : 6}>
                  {language === "th" ? "ยังไม่มีกฎทั่วไป" : "No general rules"}
                </td>
              </tr>
            ) : (
              rules.map((rule, index) => (
                <tr className="border-t border-border" key={index}>
                  <td className="px-2 py-1 font-semibold">{index + 1}</td>
                  <td className="px-2 py-1">
                    <CompactDateInput
                      disabled={readOnly}
                      value={rule.startdate}
                      onChange={(nextValue) => onUpdate(index, { startdate: nextValue })}
                    />
                  </td>
                  <td className="px-2 py-1">
                    <CompactDateInput
                      disabled={readOnly}
                      endOfDay
                      value={rule.enddate}
                      onChange={(nextValue) => onUpdate(index, { enddate: nextValue })}
                    />
                  </td>
                  <td className="px-2 py-1">
                    <CompactNumberInput
                      disabled={readOnly}
                      value={rule.payperpoint}
                      onChange={(nextValue) => onUpdate(index, { payperpoint: nextValue })}
                    />
                  </td>
                  <td className="px-2 py-1">
                    <CompactNumberInput
                      disabled={readOnly}
                      value={rule.pointvalue}
                      onChange={(nextValue) => onUpdate(index, { pointvalue: nextValue })}
                    />
                  </td>
                  {!readOnly ? (
                    <td className="px-2 py-1 text-right">
                      <Button type="button" size="sm" variant="outline" onClick={() => onRemove(index)}>
                        <Trash2 />
                      </Button>
                    </td>
                  ) : null}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function PointSpecialRulesTable({
  dictionary,
  onAdd,
  onRemove,
  onUpdate,
  readOnly = false,
  rules,
}: {
  dictionary: BackendLanguageDictionary;
  onAdd: () => void;
  onRemove: (index: number) => void;
  onUpdate: (index: number, patch: Partial<SpecialPointRule>) => void;
  readOnly?: boolean;
  rules: SpecialPointRule[];
}) {
  const text = (key: string, fallback: string) =>
    backendText(dictionary, key, fallback);
  return (
    <div className="grid gap-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <h4 className="font-semibold">{text("special_rules", "กฎพิเศษ")}</h4>
        {!readOnly ? (
          <Button type="button" size="sm" onClick={onAdd}>
            <Plus />
            {text("add_rule", "เพิ่มกฎ")}
          </Button>
        ) : null}
      </div>
      <div className="overflow-x-auto rounded-xl border border-border">
        <table className="w-full min-w-[980px] border-collapse text-left text-xs">
          <thead className="bg-muted/60 text-muted-foreground">
            <tr>
              <th className="px-2 py-2 font-semibold">#</th>
              <th className="px-2 py-2 font-semibold">{text("start_date", "วันที่เริ่มต้น")}</th>
              <th className="px-2 py-2 font-semibold">{text("end_date", "วันที่สิ้นสุด")}</th>
              <th className="px-2 py-2 font-semibold">{text("point_multiplier", "ตัวคูณแต้ม")}</th>
              <th className="px-2 py-2 font-semibold">{text("max_points_per_bill", "แต้มสูงสุดต่อบิล")}</th>
              <th className="px-2 py-2 font-semibold">{text("days_of_week", "วันในสัปดาห์")}</th>
              {!readOnly ? (
                <th className="w-20 px-2 py-2 text-right font-semibold">{text("manage", "จัดการ")}</th>
              ) : null}
            </tr>
          </thead>
          <tbody>
            {rules.length === 0 ? (
              <tr className="border-t border-border">
                <td className="px-2 py-3 text-muted-foreground" colSpan={readOnly ? 6 : 7}>
                  {text("no_special_rules", "ไม่มีกฎพิเศษ")}
                </td>
              </tr>
            ) : (
              rules.map((rule, index) => (
                <tr className="border-t border-border" key={index}>
                  <td className="px-2 py-1 font-semibold">{index + 1}</td>
                  <td className="px-2 py-1">
                    <CompactDateInput
                      disabled={readOnly}
                      value={rule.startdate}
                      onChange={(nextValue) => onUpdate(index, { startdate: nextValue })}
                    />
                  </td>
                  <td className="px-2 py-1">
                    <CompactDateInput
                      disabled={readOnly}
                      endOfDay
                      value={rule.enddate}
                      onChange={(nextValue) => onUpdate(index, { enddate: nextValue })}
                    />
                  </td>
                  <td className="px-2 py-1">
                    <CompactNumberInput
                      disabled={readOnly}
                      value={rule.multiplier}
                      onChange={(nextValue) => onUpdate(index, { multiplier: nextValue })}
                    />
                  </td>
                  <td className="px-2 py-1">
                    <CompactNumberInput
                      disabled={readOnly}
                      value={rule.maxpointperbill}
                      onChange={(nextValue) => onUpdate(index, { maxpointperbill: nextValue })}
                    />
                  </td>
                  <td className="px-2 py-1">
                    <div className="flex min-w-[340px] flex-wrap gap-1">
                      {pointWeekDays.map((day) => (
                        <label className="inline-flex w-auto max-w-none items-center gap-1 rounded-lg border border-border bg-muted/20 px-2 py-1" key={day}>
                          <input
                            aria-readonly={readOnly}
                            className="size-3.5 accent-primary"
                            type="checkbox"
                            checked={rule[day]}
                            onClick={(event) => {
                              if (readOnly) event.preventDefault();
                            }}
                            onChange={(event) => {
                              if (readOnly) return;
                              onUpdate(index, { [day]: event.target.checked } as Partial<SpecialPointRule>);
                            }}
                          />
                          <span>{shortDayLabel(day, dictionary)}</span>
                        </label>
                      ))}
                    </div>
                  </td>
                  {!readOnly ? (
                    <td className="px-2 py-1 text-right">
                      <Button type="button" size="sm" variant="outline" onClick={() => onRemove(index)}>
                        <Trash2 />
                      </Button>
                    </td>
                  ) : null}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
    </div>
  );
}

function PaymentMethodToggle({
  checked,
  disabled = false,
  label,
  onChange,
  text,
}: {
  checked: boolean;
  disabled?: boolean;
  label: string;
  onChange: (checked: boolean) => void;
  text: (key: string, fallback: string) => string;
}) {
  return (
    <div className="grid gap-1">
      <span className="font-semibold">{label}</span>
      <label className="inline-flex w-auto max-w-none items-center gap-2 text-xs font-medium text-muted-foreground">
        <input
          aria-readonly={disabled}
          className="size-4 accent-primary"
          type="checkbox"
          checked={checked}
          onClick={(event) => {
            if (disabled) event.preventDefault();
          }}
          onChange={(event) => {
            if (disabled) return;
            onChange(event.target.checked);
          }}
        />
        {text("enable_rounding", "เปิดการปัดเศษ")}
      </label>
    </div>
  );
}

function CompactNumberInput({
  disabled,
  onChange,
  value,
}: {
  disabled?: boolean;
  onChange: (value: number) => void;
  value: number;
}) {
  return (
    <input
      className="h-8 w-24 rounded-lg border border-input bg-background px-2 text-right text-xs text-foreground outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-60"
      disabled={disabled}
      inputMode="decimal"
      type="number"
      step="0.01"
      value={formatConfigNumber(value)}
      onChange={(event) => onChange(numberFromInput(event.target.value))}
    />
  );
}

function CompactDateInput({
  disabled,
  endOfDay,
  onChange,
  value,
}: {
  disabled?: boolean;
  endOfDay?: boolean;
  onChange: (value: string) => void;
  value: string;
}) {
  return (
    <input
      className="h-8 w-36 rounded-lg border border-input bg-background px-2 text-xs text-foreground outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:opacity-60"
      disabled={disabled}
      type="date"
      value={isoDateToInput(value)}
      onChange={(event) => onChange(dateInputToIso(event.target.value, endOfDay))}
    />
  );
}

function FragmentLikeRows({ children }: { children: ReactNode }) {
  return <>{children}</>;
}

function normalizePaymentRoundingConfig(
  value: unknown,
): BranchStructuredParseResult<PaymentRoundingConfig> {
  const parsed = parseBranchSettingRecord(value, defaultPaymentRoundingJson());
  const source = parsed.record;
  return {
    isRecovered: parsed.isRecovered,
    value: Object.fromEntries(
      paymentMethodKeys.map((method) => {
        const methodValue = isRecord(source[method]) ? source[method] : {};
        const rules = Array.isArray(methodValue.rules)
          ? methodValue.rules.map(normalizeRoundingRule)
          : defaultRoundingRules();
        return [
          method,
          {
            enabled:
              typeof methodValue.enabled === "boolean"
                ? methodValue.enabled
                : true,
            rules: rules.length ? rules : [defaultRoundingRule()],
          },
        ];
      }),
    ) as PaymentRoundingConfig,
  };
}

function normalizePointConfig(
  value: unknown,
): BranchStructuredParseResult<PointConfig> {
  const parsed = parseBranchSettingRecord(value, defaultPointConfigJson());
  const source = parsed.record;
  const pointUsageType = Number(source.pointusagetype);
  return {
    isRecovered: parsed.isRecovered,
    value: {
      generalrules: Array.isArray(source.generalrules)
        ? source.generalrules.map(normalizeGeneralPointRule)
        : [],
      specialrules: Array.isArray(source.specialrules)
        ? source.specialrules.map(normalizeSpecialPointRule)
        : [],
      pointusagetype: pointUsageType === 2 ? 2 : 1,
    },
  };
}

function parseBranchSettingRecord(
  value: unknown,
  fallbackJson: string,
): { isRecovered: boolean; record: SettingRecord } {
  const fallback = JSON.parse(fallbackJson) as SettingRecord;
  if (isRecord(value)) return { isRecovered: false, record: value };
  if (typeof value !== "string" || !value.trim())
    return { isRecovered: false, record: fallback };
  try {
    const parsed = JSON.parse(value) as unknown;
    if (isRecord(parsed)) return { isRecovered: false, record: parsed };
  } catch {
    return { isRecovered: true, record: fallback };
  }
  return { isRecovered: true, record: fallback };
}

function normalizeRoundingRule(value: unknown): RoundingRule {
  const record = isRecord(value) ? value : {};
  return {
    lowerbound: normalizeConfigNumber(record.lowerbound, 0),
    upperbound: normalizeConfigNumber(record.upperbound, 0),
    roundto: normalizeConfigNumber(record.roundto, 0),
  };
}

function normalizeGeneralPointRule(value: unknown): GeneralPointRule {
  const record = isRecord(value) ? value : {};
  const fallback = defaultGeneralPointRule();
  return {
    startdate: normalizeIsoDate(record.startdate, fallback.startdate),
    enddate: normalizeIsoDate(record.enddate, fallback.enddate),
    payperpoint: normalizeConfigNumber(record.payperpoint, 20),
    pointvalue: normalizeConfigNumber(record.pointvalue, 1),
  };
}

function normalizeSpecialPointRule(value: unknown): SpecialPointRule {
  const record = isRecord(value) ? value : {};
  const fallback = defaultSpecialPointRule();
  return {
    startdate: normalizeIsoDate(record.startdate, fallback.startdate),
    enddate: normalizeIsoDate(record.enddate, fallback.enddate),
    multiplier: normalizeConfigNumber(record.multiplier, 2),
    sunday: Boolean(record.sunday),
    monday: record.monday === undefined ? true : Boolean(record.monday),
    tuesday: record.tuesday === undefined ? true : Boolean(record.tuesday),
    wednesday: record.wednesday === undefined ? true : Boolean(record.wednesday),
    thursday: record.thursday === undefined ? true : Boolean(record.thursday),
    friday: record.friday === undefined ? true : Boolean(record.friday),
    saturday: Boolean(record.saturday),
    maxpointperbill: normalizeConfigNumber(record.maxpointperbill, 100),
  };
}

function defaultRoundingRules(): RoundingRule[] {
  return [
    { lowerbound: 0.01, upperbound: 0.12, roundto: 0 },
    { lowerbound: 0.13, upperbound: 0.37, roundto: 0.25 },
    { lowerbound: 0.38, upperbound: 0.62, roundto: 0.5 },
    { lowerbound: 0.63, upperbound: 0.87, roundto: 0.75 },
    { lowerbound: 0.88, upperbound: 0.99, roundto: 1 },
  ];
}

function defaultRoundingRule(): RoundingRule {
  return { lowerbound: 0.01, upperbound: 0.99, roundto: 0 };
}

function defaultGeneralPointRule(): GeneralPointRule {
  const start = new Date();
  const end = new Date(start);
  end.setDate(end.getDate() + 365);
  return {
    startdate: start.toISOString(),
    enddate: end.toISOString(),
    payperpoint: 20,
    pointvalue: 1,
  };
}

function defaultSpecialPointRule(): SpecialPointRule {
  const start = new Date();
  const end = new Date(start);
  end.setDate(end.getDate() + 30);
  return {
    startdate: start.toISOString(),
    enddate: end.toISOString(),
    multiplier: 2,
    sunday: false,
    monday: true,
    tuesday: true,
    wednesday: true,
    thursday: true,
    friday: true,
    saturday: false,
    maxpointperbill: 100,
  };
}

function paymentRoundingSummary(
  value: PaymentRoundingConfig,
  language: LanguageCode,
): string {
  const methods = paymentMethodKeys.map((method) => value[method]);
  const enabledCount = methods.filter((method) => Boolean(method.enabled)).length;
  const maxRules = Math.max(
    0,
    ...methods.map((method) => method.rules.length),
  );
  return language === "th"
    ? `เปิดใช้ ${enabledCount}/${methods.length} ช่องทาง, กฎสูงสุด ${maxRules} รายการ`
    : `Enabled ${enabledCount}/${methods.length} methods, max ${maxRules} rules`;
}

function pointConfigSummary(value: PointConfig, language: LanguageCode): string {
  return language === "th"
    ? `กฎทั่วไป ${value.generalrules.length} รายการ, กฎพิเศษ ${value.specialrules.length} รายการ, ประเภทใช้แต้ม ${value.pointusagetype}`
    : `${value.generalrules.length} general rules, ${value.specialrules.length} special rules, usage type ${value.pointusagetype}`;
}

function paymentMethodLabel(
  method: PaymentMethodKey,
  dictionary: BackendLanguageDictionary,
): string {
  const keys: Record<PaymentMethodKey, [string, string]> = {
    banktransfer: ["bank_transfer", "การโอนเงิน"],
    cash: ["cash", "เงินสด"],
    cheque: ["cheque", "เช็ค"],
    coupon: ["coupon", "คูปอง"],
    creditcard: ["credit_card", "บัตรเครดิต"],
    delivery: ["delivery", "เดลิเวอรี่"],
    qrcode: ["qrcode", "คิวอาร์โค้ด"],
  };
  const [key, fallback] = keys[method];
  return backendText(dictionary, key, fallback);
}

function shortDayLabel(
  day: PointWeekDay,
  dictionary: BackendLanguageDictionary,
): string {
  const fallback: Record<PointWeekDay, string> = {
    friday: "ศุกร์",
    monday: "จันทร์",
    saturday: "เสาร์",
    sunday: "อาทิตย์",
    thursday: "พฤหัส",
    tuesday: "อังคาร",
    wednesday: "พุธ",
  };
  return backendText(dictionary, day, fallback[day]);
}

function invalidConfigText(language: LanguageCode): string {
  return language === "th"
    ? "ข้อมูลเดิมไม่ถูกต้อง ใช้ค่าเริ่มต้น"
    : "Invalid saved config; default values are shown";
}

function normalizeConfigNumber(value: unknown, fallback: number): number {
  const next = Number(value);
  return Number.isFinite(next) ? next : fallback;
}

function numberFromInput(value: string): number {
  const next = Number(value);
  return Number.isFinite(next) ? next : 0;
}

function formatConfigNumber(value: number): string {
  return Number.isFinite(value) ? String(value) : "";
}

function normalizeIsoDate(value: unknown, fallback: string): string {
  if (typeof value !== "string" || !value.trim()) return fallback;
  const date = new Date(value);
  return Number.isNaN(date.getTime()) ? fallback : date.toISOString();
}

function isoDateToInput(value: string): string {
  const date = new Date(value);
  if (Number.isNaN(date.getTime())) return "";
  return date.toISOString().slice(0, 10);
}

function dateInputToIso(value: string, endOfDay?: boolean): string {
  if (!value) return "";
  const suffix = endOfDay ? "T23:59:59" : "T00:00:00";
  return new Date(`${value}${suffix}`).toISOString();
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
  const rows = normalizeLanguageConfigs(
    form["settings.languageconfigs"],
    defaultCode,
  );
  const usedCodes = new Set(rows.map((row) => row.code));
  const availableLanguages = LANGUAGES.filter(
    (item) => !usedCodes.has(item.code),
  );
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [draggingCode, setDraggingCode] = useState("");

  function commit(
    nextRows: LanguageConfigFormRow[],
    nextDefault = nextRows[0]?.code ?? defaultCode,
  ) {
    const normalized = normalizeLanguageConfigs(nextRows, nextDefault, {
      forcePrimaryFirst: true,
    });
    const primary =
      normalized[0]?.code ?? supportedLanguageCode(nextDefault, "th");
    setForm({
      ...form,
      "settings.language": primary,
      "settings.languageconfigs": normalized,
    });
  }

  function reorderByCode(sourceCode: string, targetCode: string) {
    if (!sourceCode || sourceCode === targetCode) return;
    const sourceIndex = rows.findIndex((row) => row.code === sourceCode);
    const targetIndex = rows.findIndex((row) => row.code === targetCode);
    if (sourceIndex < 0 || targetIndex < 0 || sourceIndex === targetIndex)
      return;
    commit(moveArrayItem(rows, sourceIndex, targetIndex));
  }

  function handleDragStart(event: ReactDragEvent<HTMLElement>, code: string) {
    if (rows.length <= 1) return;
    setDraggingCode(code);
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", code);
  }

  function handleDragOver(
    event: ReactDragEvent<HTMLElement>,
    targetCode: string,
  ) {
    const sourceCode = draggingCode || event.dataTransfer.getData("text/plain");
    if (!sourceCode || sourceCode === targetCode) return;
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
    reorderByCode(sourceCode, targetCode);
  }

  function handleDrop(event: ReactDragEvent<HTMLElement>, targetCode: string) {
    event.preventDefault();
    reorderByCode(
      draggingCode || event.dataTransfer.getData("text/plain"),
      targetCode,
    );
    setDraggingCode("");
  }

  const textPrimary = language === "th" ? "ภาษาแรก" : "Primary language";
  const textAdd = language === "th" ? "เพิ่มภาษา" : "Add language";
  const textRemove = language === "th" ? "เอาออก" : "Remove";
  const textNoMore =
    language === "th"
      ? "เพิ่มครบทุกภาษาที่รองรับแล้ว"
      : "All supported languages are already added.";
  const textDrag =
    language === "th" ? "ลากเพื่อจัดลำดับภาษา" : "Drag to reorder language";

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="min-w-0">
          <div className="font-semibold">{label}</div>
          <div className="text-xs text-muted-foreground">
            {language === "th"
              ? "ลำดับแรกคือภาษาแรกของบริษัท"
              : "The first row is the company primary language."}
          </div>
        </div>
        {availableLanguages.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            <Button
              type="button"
              size="sm"
              onClick={() => setAddDialogOpen(true)}
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
            draggable={rows.length > 1}
            onDragEnd={() => setDraggingCode("")}
            onDragOver={(event) => handleDragOver(event, row.code)}
            onDragStart={(event) => handleDragStart(event, row.code)}
            onDrop={(event) => handleDrop(event, row.code)}
            title={textDrag}
            className={cn(
              "grid cursor-grab gap-2 rounded-xl border border-border bg-card px-2 py-1.5 transition-[transform,box-shadow,border-color,background-color,opacity] duration-150 ease-out hover:-translate-y-0.5 hover:shadow-sm active:cursor-grabbing sm:grid-cols-[84px_minmax(0,1fr)_auto] sm:items-center",
              index === 0 && "border-primary/40 bg-primary/5",
              draggingCode === row.code &&
                "scale-[0.99] opacity-60 ring-2 ring-primary/30",
            )}
          >
            <div className="flex items-center gap-1">
              <span className="grid size-8 place-items-center rounded-lg border border-border bg-background text-muted-foreground">
                <ChevronsUpDown className="size-4" aria-hidden="true" />
              </span>
              <div
                className={cn(
                  "grid size-10 place-items-center rounded-full font-semibold",
                  index === 0
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted text-foreground",
                )}
              >
                {index + 1}
              </div>
            </div>
            <div className="flex min-w-0 items-center gap-2">
              <LanguageFlag code={row.code} />
              <div className="min-w-0">
                <div className="truncate font-semibold">
                  {languageName(row.code, language)}
                </div>
                <div className="flex flex-wrap items-center gap-2 text-xs uppercase text-muted-foreground">
                  <span>{row.code}</span>
                  {index === 0 ? (
                    <Badge variant="outline">{textPrimary}</Badge>
                  ) : null}
                </div>
              </div>
            </div>
            <div className="flex flex-wrap justify-end gap-1">
              {rows.length > 1 ? (
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() =>
                    commit(rows.filter((item) => item.code !== row.code))
                  }
                  title={textRemove}
                >
                  <Trash2 />
                  <span className="sr-only">{textRemove}</span>
                </Button>
              ) : null}
            </div>
          </div>
        ))}
      </div>
      {availableLanguages.length === 0 ? (
        <p className="text-xs text-muted-foreground">{textNoMore}</p>
      ) : null}
      <LanguageAddDialog
        activeCodes={rows.map((row) => row.code)}
        language={language}
        open={addDialogOpen}
        onClose={() => setAddDialogOpen(false)}
        onToggle={(code) => {
          const isActive = rows.some((row) => row.code === code);
          if (isActive) {
            if (rows[0]?.code === code) return;
            commit(rows.filter((row) => row.code !== code));
          } else {
            commit([...rows, languageConfigRow(code, false)]);
          }
        }}
      />
    </section>
  );
}

function LanguageListEditor({
  field,
  form,
  language,
  label,
  readOnly = false,
  setForm,
}: {
  field: SystemSettingField;
  form: FormState;
  language: LanguageCode;
  label: string;
  readOnly?: boolean;
  setForm?: (form: FormState) => void;
}) {
  const rows = normalizeLanguageList(form[field.key], form.language);
  const usedCodes = new Set(rows);
  const availableLanguages = LANGUAGES.filter(
    (item) => !usedCodes.has(item.code),
  );
  const [addDialogOpen, setAddDialogOpen] = useState(false);
  const [draggingCode, setDraggingCode] = useState("");
  const textPrimary = language === "th" ? "ภาษาแรก" : "Primary";
  const textAdd = language === "th" ? "เพิ่มภาษา" : "Add language";
  const textRemove = language === "th" ? "เอาออก" : "Remove";
  const textNoMore =
    language === "th"
      ? "เพิ่มครบทุกภาษาที่รองรับแล้ว"
      : "All supported languages are already added.";
  const textDrag =
    language === "th" ? "ลากเพื่อจัดลำดับภาษา" : "Drag to reorder language";

  function commit(nextCodes: string[]) {
    if (readOnly || !setForm) return;
    const normalized = normalizeLanguageList(
      nextCodes,
      nextCodes[0] ?? form.language,
      { forcePrimaryFirst: true },
    );
    setForm({
      ...form,
      [field.key]: normalized,
      language: normalized[0] ?? "th",
    });
  }

  function reorderByCode(sourceCode: string, targetCode: string) {
    if (readOnly) return;
    if (!sourceCode || sourceCode === targetCode) return;
    const sourceIndex = rows.findIndex((code) => code === sourceCode);
    const targetIndex = rows.findIndex((code) => code === targetCode);
    if (sourceIndex < 0 || targetIndex < 0 || sourceIndex === targetIndex)
      return;
    commit(moveArrayItem(rows, sourceIndex, targetIndex));
  }

  function handleDragStart(event: ReactDragEvent<HTMLElement>, code: string) {
    if (readOnly) return;
    if (rows.length <= 1) return;
    setDraggingCode(code);
    event.dataTransfer.effectAllowed = "move";
    event.dataTransfer.setData("text/plain", code);
  }

  function handleDragOver(
    event: ReactDragEvent<HTMLElement>,
    targetCode: string,
  ) {
    if (readOnly) return;
    const sourceCode = draggingCode || event.dataTransfer.getData("text/plain");
    if (!sourceCode || sourceCode === targetCode) return;
    event.preventDefault();
    event.dataTransfer.dropEffect = "move";
    reorderByCode(sourceCode, targetCode);
  }

  function handleDrop(event: ReactDragEvent<HTMLElement>, targetCode: string) {
    if (readOnly) return;
    event.preventDefault();
    reorderByCode(
      draggingCode || event.dataTransfer.getData("text/plain"),
      targetCode,
    );
    setDraggingCode("");
  }

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <div className="min-w-0">
          <div className="font-semibold">{label}</div>
          <div className="text-xs text-muted-foreground">
            {language === "th"
              ? "ลำดับแรกคือภาษาแรกของข้อมูลนี้"
              : "The first row is the primary language for this record."}
          </div>
        </div>
        {!readOnly && availableLanguages.length > 0 ? (
          <div className="flex flex-wrap gap-2">
            <Button
              type="button"
              size="sm"
              onClick={() => setAddDialogOpen(true)}
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
            draggable={!readOnly && rows.length > 1}
            onDragEnd={() => setDraggingCode("")}
            onDragOver={(event) => handleDragOver(event, code)}
            onDragStart={(event) => handleDragStart(event, code)}
            onDrop={(event) => handleDrop(event, code)}
            title={textDrag}
            className={cn(
              "grid cursor-grab gap-2 rounded-xl border border-border bg-card px-2 py-1.5 transition-[transform,box-shadow,border-color,background-color,opacity] duration-150 ease-out hover:-translate-y-0.5 hover:shadow-sm active:cursor-grabbing sm:grid-cols-[84px_minmax(0,1fr)_auto] sm:items-center",
              readOnly &&
                "cursor-default hover:translate-y-0 active:cursor-default",
              index === 0 && "border-primary/40 bg-primary/5",
              draggingCode === code &&
                "scale-[0.99] opacity-60 ring-2 ring-primary/30",
            )}
          >
            <div className="flex items-center gap-1">
              <span className="grid size-8 place-items-center rounded-lg border border-border bg-background text-muted-foreground">
                <ChevronsUpDown className="size-4" aria-hidden="true" />
              </span>
              <div
                className={cn(
                  "grid size-10 place-items-center rounded-full font-semibold",
                  index === 0
                    ? "bg-primary text-primary-foreground"
                    : "bg-muted text-foreground",
                )}
              >
                {index + 1}
              </div>
            </div>
            <div className="flex min-w-0 items-center gap-2">
              <LanguageFlag code={code} />
              <div className="min-w-0">
                <div className="truncate font-semibold">
                  {languageName(code, language)}
                </div>
                <div className="flex flex-wrap items-center gap-2 text-xs uppercase text-muted-foreground">
                  <span>{code}</span>
                  {index === 0 ? (
                    <Badge variant="outline">{textPrimary}</Badge>
                  ) : null}
                </div>
              </div>
            </div>
            <div className="flex flex-wrap justify-end gap-1">
              {!readOnly && rows.length > 1 ? (
                <Button
                  type="button"
                  variant="ghost"
                  size="sm"
                  onClick={() => commit(rows.filter((item) => item !== code))}
                  title={textRemove}
                >
                  <Trash2 />
                  <span className="sr-only">{textRemove}</span>
                </Button>
              ) : null}
            </div>
          </div>
        ))}
      </div>
      {!readOnly && availableLanguages.length === 0 ? (
        <p className="text-xs text-muted-foreground">{textNoMore}</p>
      ) : null}
      {!readOnly ? (
        <LanguageAddDialog
          activeCodes={rows}
          language={language}
          open={addDialogOpen}
          onClose={() => setAddDialogOpen(false)}
          onToggle={(code) => {
            const isActive = rows.includes(code);
            if (isActive) {
              if (rows[0] === code) return;
              commit(rows.filter((item) => item !== code));
            } else {
              commit([...rows, code]);
            }
          }}
        />
      ) : null}
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
  const displayName =
    localizedNameForLanguage(names, language) || stringValue(value.name);
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
      <span>
        {label}
        {field.required ? " *" : ""}
      </span>
      <div className="flex min-w-0 gap-2">
        <button
          ref={anchorRef}
          type="button"
          className="flex min-h-10 min-w-0 flex-1 items-center justify-between gap-2 rounded-2xl border border-input bg-background px-3 text-left text-sm text-foreground shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
          onClick={() => setOpen(true)}
        >
          <span
            className={cn(
              "min-w-0 truncate",
              !code && !displayName && "text-muted-foreground",
            )}
          >
            {code || displayName
              ? `${code}${code && displayName ? " - " : ""}${displayName}`
              : language === "th"
                ? "เลือกข้อมูล"
                : "Select"}
          </span>
          <Search className="size-4 shrink-0 text-muted-foreground" />
        </button>
        {code || displayName ? (
          <Button
            type="button"
            variant="outline"
            size="icon"
            onClick={() => setForm({ ...form, [field.key]: {} })}
            aria-label={language === "th" ? "ล้างค่า" : "Clear"}
          >
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

function moveArrayItem<T>(
  items: readonly T[],
  fromIndex: number,
  toIndex: number,
): T[] {
  const next = [...items];
  const [item] = next.splice(fromIndex, 1);
  if (item === undefined) return next;
  next.splice(toIndex, 0, item);
  return next;
}

function setDefaultLanguageConfig(
  value: unknown,
  defaultCode: unknown,
): LanguageConfigFormRow[] {
  return normalizeLanguageConfigs(value, defaultCode, {
    forcePrimaryFirst: true,
  });
}

export function normalizeLanguageConfigs(
  value: unknown,
  defaultCode: unknown,
  options: NormalizeLanguageOptions = {},
): LanguageConfigFormRow[] {
  const requestedPrimaryCode = supportedLanguageCode(defaultCode, "");
  const source = typeof value === "string" ? safeJsonParse(value, []) : value;
  const input = Array.isArray(source) ? source : [];
  const seen = new Set<string>();
  const rows: LanguageConfigFormRow[] = [];

  for (const item of input) {
    if (!isRecord(item)) continue;
    const code = supportedLanguageCode(item.code, "");
    if (!code || seen.has(code)) continue;
    const isDefault = booleanLikeValue(item.isdefault);
    const isUse =
      item.is_use === undefined && item.isuse === undefined
        ? true
        : booleanLikeValue(item.is_use ?? item.isuse);
    if (!isUse && !isDefault) continue;
    seen.add(code);
    rows.push({
      code,
      codetranslator:
        stringValue(item.codetranslator ?? item.codeTranslator) || code,
      name: stringValue(item.name) || languageName(code, "en"),
      is_use: true,
      isdefault: isDefault,
    });
  }

  if (!rows.length)
    return [languageConfigRow(requestedPrimaryCode || "th", true)];

  const requestedExists = Boolean(
    requestedPrimaryCode && seen.has(requestedPrimaryCode),
  );
  const primaryCode = supportedLanguageCode(
    options.forcePrimaryFirst && requestedPrimaryCode
      ? requestedPrimaryCode
      : rows.find((row) => row.isdefault)?.code ||
          rows[0]?.code ||
          requestedPrimaryCode ||
          "th",
    "th",
  );
  if (!seen.has(primaryCode))
    rows.unshift(languageConfigRow(primaryCode, true));
  const ordered =
    options.forcePrimaryFirst && requestedExists
      ? [
          rows.find((row) => row.code === primaryCode) ??
            languageConfigRow(primaryCode, true),
          ...rows.filter((row) => row.code !== primaryCode),
        ]
      : rows;
  return ordered.map((row) => ({
    ...row,
    codetranslator: row.codetranslator || row.code,
    isdefault: row.code === primaryCode,
    is_use: true,
    name: row.name || languageName(row.code, "en"),
  }));
}

function languageConfigRow(
  code: LanguageCode,
  isDefault: boolean,
): LanguageConfigFormRow {
  return {
    code,
    codetranslator: code,
    name: languageName(code, "en"),
    is_use: true,
    isdefault: isDefault,
  };
}

function supportedLanguageCode(
  value: unknown,
  fallback: LanguageCode,
): LanguageCode;
function supportedLanguageCode(value: unknown, fallback: ""): LanguageCode | "";
function supportedLanguageCode(
  value: unknown,
  fallback: LanguageCode | "",
): LanguageCode | "" {
  const raw = stringValue(value).toLowerCase();
  if (!raw) return fallback;
  const normalized = normalizeLanguage(raw);
  return LANGUAGES.some((item) => item.code === normalized)
    ? normalized
    : fallback;
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

function LanguageAddDialog({
  activeCodes,
  language,
  onClose,
  onToggle,
  open,
}: {
  activeCodes: string[];
  language: LanguageCode;
  onClose: () => void;
  onToggle: (code: LanguageCode) => void;
  open: boolean;
}) {
  useEffect(() => {
    if (!open) return;
    const previousOverflow = document.body.style.overflow;
    const onKeyDown = (event: KeyboardEvent) => {
      if (event.key === "Escape") onClose();
    };
    document.body.style.overflow = "hidden";
    document.addEventListener("keydown", onKeyDown);
    return () => {
      document.body.style.overflow = previousOverflow;
      document.removeEventListener("keydown", onKeyDown);
    };
  }, [onClose, open]);

  if (!open) return null;

  const title = language === "th" ? "เลือกภาษา" : "Select Language";
  const closeText = language === "th" ? "ปิด" : "Close";

  return (
    <div className="dialog-backdrop" onClick={onClose} role="presentation">
      <section
        aria-label={title}
        aria-modal="true"
        className="language-dialog"
        onClick={(event) => event.stopPropagation()}
        role="dialog"
      >
        <div className="dialog-header">
          <div className="min-w-0">
            <p className="eyebrow">{language === "th" ? "ภาษา" : "Language"}</p>
            <h2>{title}</h2>
          </div>
          <Button
            aria-label={closeText}
            type="button"
            variant="outline"
            size="icon"
            onClick={onClose}
          >
            <X aria-hidden="true" />
          </Button>
        </div>

        <div className="language-grid">
          {LANGUAGES.map((item) => {
            const isPrimary = activeCodes[0] === item.code;
            const isActive = activeCodes.includes(item.code);
            return (
              <button
                className={cn(
                  "language-choice transition-all",
                  isPrimary &&
                    "border-primary/30 bg-primary/10 text-primary cursor-default hover:bg-primary/10 hover:border-primary/30 shadow-none",
                  isActive &&
                    !isPrimary &&
                    "border-primary text-primary bg-background hover:bg-primary/5",
                  !isActive &&
                    "border-border bg-card text-foreground hover:border-primary/50 hover:bg-accent/50",
                )}
                key={item.code}
                onClick={() => {
                  if (isPrimary) return;
                  onToggle(item.code);
                }}
                type="button"
              >
                <LanguageFlag code={item.code} />
                <span className="font-semibold">
                  {languageName(item.code, language)}
                </span>
                {isPrimary && (
                  <Check
                    className="ml-auto size-4 text-primary"
                    aria-hidden="true"
                  />
                )}
              </button>
            );
          })}
        </div>
      </section>
    </div>
  );
}

function normalizeLanguageList(
  value: unknown,
  defaultCode: unknown,
  options: NormalizeLanguageOptions = {},
): string[] {
  const requestedPrimaryCode = supportedLanguageCode(defaultCode, "");
  const source =
    typeof value === "string"
      ? value.trim().startsWith("[")
        ? safeJsonParse(value, [])
        : value.split(",")
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

  if (!rows.length) return [requestedPrimaryCode || "th"];
  if (!options.forcePrimaryFirst || !requestedPrimaryCode) return rows;
  if (!seen.has(requestedPrimaryCode)) return [requestedPrimaryCode, ...rows];
  return [
    requestedPrimaryCode,
    ...rows.filter((code) => code !== requestedPrimaryCode),
  ];
}

function nameEditorLanguageCodes(
  form: FormState,
  config: SystemSettingConfig,
  language: LanguageCode,
  workspace?: WorkspaceSession | null,
): string[] {
  if (config.kind === "company") {
    return normalizeLanguageConfigs(
      form["settings.languageconfigs"],
      form["settings.language"],
    ).map((row) => row.code);
  }
  if (config.slug === "branch") {
    return normalizeLanguageList(form.languages, form.language);
  }
  if (workspace) {
    const shopInfo = isRecord(workspace.shopInfo) ? workspace.shopInfo : {};
    const settings = isRecord(shopInfo.settings) ? shopInfo.settings : {};

    const rawConfigs = settings.languageconfigs ?? shopInfo["settings.languageconfigs"] ?? shopInfo.languageconfigs;
    const rawLang = settings.language ?? shopInfo["settings.language"] ?? shopInfo.language;

    if (Array.isArray(rawConfigs)) {
      const activeCodes = rawConfigs
        .filter(isRecord)
        .map((item) => {
          const isUse = item.is_use === undefined && item.isuse === undefined ? true : (item.is_use === true || item.is_use === "true" || item.isuse === true || item.isuse === "true");
          return {
            code: supportedLanguageCode(item.code, ""),
            isUse,
          };
        })
        .filter((item) => item.code && item.isUse)
        .map((item) => item.code);
      if (activeCodes.length > 0) {
        const primary = supportedLanguageCode(rawLang, language);
        return [primary, ...activeCodes.filter((code) => code !== primary)];
      }
    }
  }
  return [language];
}

function getLocalizedNameArray(
  value: unknown,
): Array<{ code?: string; name?: string }> {
  if (!Array.isArray(value)) return [];
  return value.filter(isRecord).map((item) => ({
    code: stringValue(item.code),
    name: stringValue(item.name),
  }));
}

function localizedNameForLanguage(
  names: Array<{ code?: string; name?: string }>,
  language: LanguageCode,
): string {
  return (
    names.find((item) => item.code === language && item.name)?.name ??
    names.find((item) => item.code === "th" && item.name)?.name ??
    names.find((item) => item.name)?.name ??
    ""
  );
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
  return JSON.stringify(
    {
      banktransfer: method,
      cash: method,
      cheque: method,
      coupon: method,
      creditcard: method,
      delivery: method,
      qrcode: method,
    },
    null,
    2,
  );
}

function defaultPointConfigJson(): string {
  return JSON.stringify(
    {
      generalrules: [],
      specialrules: [],
      pointusagetype: 1,
    },
    null,
    2,
  );
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
type UploadTextKey =
  | "applyCrop"
  | "cancel"
  | "chooseImage"
  | "cropImage"
  | "cropTitle"
  | "imageDisplayFailed"
  | "imageTooLarge"
  | "imageUploadFailed"
  | "imageUploadHint"
  | "removeImage"
  | "uploading";

const uploadText: Record<
  UploadTextKey,
  Partial<Record<LanguageCode, string>> & { en: string; th: string }
> = {
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
  imageDisplayFailed: {
    th: "ไม่สามารถแสดงรูปได้",
    en: "Cannot display image",
    cn: "无法显示图片",
    ja: "画像を表示できません",
    ko: "이미지를 표시할 수 없습니다",
    lo: "ບໍ່ສາມາດສະແດງຮູບໄດ້",
    my: "ပုံကို ပြသ၍မရပါ",
    km: "មិនអាចបង្ហាញរូបភាពបាន",
    vi: "Không thể hiển thị ảnh",
    ms: "Tidak dapat memaparkan imej",
    id: "Tidak dapat menampilkan gambar",
    fil: "Hindi maipakita ang larawan",
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
  const {
    displayUrl: previewValue,
    failed: previewLoadFailed,
    loading: previewLoading,
  } = useAuthenticatedImageDisplaySource(localPreview || value, auth);
  const hasImageValue = Boolean(localPreview || value);
  const previewStyle = previewValue
    ? { backgroundImage: `url(${JSON.stringify(previewValue)})` }
    : undefined;

  // Reset local preview and errors when active form record changes (e.g., category changes)
  const formGuid = String(form.guid_fixed || form.guidfixed || form.guid || "");
  useEffect(() => {
    setLocalPreview("");
    setError("");
  }, [formGuid]);

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
      uploadForm.append("category", `system-settings/${field.key}`);
      const response = await fetch("/api/upload/image", {
        method: "POST",
        headers: {
          "x-bc-backend-url": auth.backendUrl,
          Authorization: `Bearer ${auth.token}`,
        },
        body: uploadForm,
      });
      const payload = (await response.json()) as unknown;
      if (!response.ok || isFailed(payload))
        throw new Error(
          extractMessage(payload) ??
            uploadUiText(language, "imageUploadFailed"),
        );
      const uri = extractUploadUri(payload);
      if (!uri) throw new Error(uploadUiText(language, "imageUploadFailed"));
      setForm({ ...form, [field.key]: uri });
    } catch (uploadError) {
      setError(
        uploadError instanceof Error && uploadError.message
          ? uploadError.message
          : uploadUiText(language, "imageUploadFailed"),
      );
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
      <span>
        {label}
        {field.required ? " *" : ""}
      </span>
      <div className="flex w-full flex-wrap items-center gap-3">
        <button
          aria-label={label}
          className="grid size-20 place-items-center overflow-hidden rounded-2xl border border-input bg-card bg-contain bg-center bg-no-repeat text-muted-foreground shadow-sm"
          onClick={() => inputRef.current?.click()}
          style={previewStyle}
          type="button"
        >
          {previewValue ? (
            <span className="sr-only">{label}</span>
          ) : previewLoading ? (
            <Loader2 className="size-8 animate-spin" />
          ) : (
            <ImageIcon className="size-8" />
          )}
        </button>
        <div className="grid min-w-48 flex-1 gap-2">
          <div className="flex w-full flex-wrap gap-2">
            <Button
              type="button"
              variant="outline"
              onClick={() => inputRef.current?.click()}
              disabled={uploading || !auth}
            >
              {uploading ? (
                <Loader2 className="animate-spin" />
              ) : (
                <UploadCloud />
              )}
              {uploading
                ? uploadUiText(language, "uploading")
                : uploadUiText(language, "chooseImage")}
            </Button>
            {previewValue ? (
              <Button
                type="button"
                variant="outline"
                onClick={() => setCropSource(previewValue)}
                disabled={uploading || previewLoading}
              >
                <Edit3 />
                {uploadUiText(language, "cropImage")}
              </Button>
            ) : null}
            {hasImageValue ? (
              <Button
                type="button"
                variant="outline"
                onClick={clearImage}
                disabled={uploading}
              >
                <Trash2 />
                {uploadUiText(language, "removeImage")}
              </Button>
            ) : null}
          </div>
          <p className="text-xs font-normal text-muted-foreground">
            {uploadUiText(language, "imageUploadHint")}
          </p>
          {error ? (
            <p className="text-xs font-semibold text-destructive">{error}</p>
          ) : null}
          {previewLoadFailed && value ? (
            <p className="break-all text-xs font-semibold text-destructive">
              {uploadUiText(language, "imageDisplayFailed")}
              <br />
              {value}
            </p>
          ) : null}
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

function ImageUploadReadOnlyDetail({
  auth,
  label,
  language,
  value,
}: {
  auth: AuthSession | null;
  label: string;
  language: LanguageCode;
  value: unknown;
}) {
  const rawValue = stringValue(value);
  const {
    displayUrl,
    failed: fetchFailed,
    requestedUrl,
  } = useAuthenticatedImageDisplaySource(rawValue, auth);
  const [elementFailed, setElementFailed] = useState(false);
  const canShowImage = Boolean(displayUrl) && !fetchFailed && !elementFailed;

  useEffect(() => {
    setElementFailed(false);
  }, [displayUrl]);

  return (
    <div className="grid gap-2 rounded-xl border border-border bg-card px-3 py-2 text-sm shadow-[0_1px_2px_rgba(0,0,0,0.02)] transition-all duration-200 border-l-2 border-l-secondary">
      <span className="text-xs font-semibold text-muted-foreground">
        {label}
      </span>
      {canShowImage ? (
        <div className="flex w-full flex-wrap items-start gap-3">
          <div className="relative h-28 w-28 shrink-0 overflow-hidden rounded-xl border border-input bg-muted">
            <Image
              alt={label}
              className="object-cover"
              fill
              sizes="112px"
              src={displayUrl}
              unoptimized
              onError={() => setElementFailed(true)}
            />
          </div>
          <b className="min-w-0 flex-1 break-all text-sm font-medium text-foreground">
            {rawValue || requestedUrl}
          </b>
        </div>
      ) : (
        <div className="grid gap-1">
          {(fetchFailed || elementFailed) && rawValue ? (
            <span className="text-xs font-semibold text-destructive">
              {uploadUiText(language, "imageDisplayFailed")}
            </span>
          ) : null}
          <b className="min-w-0 break-words text-foreground font-medium">
            {rawValue || "-"}
          </b>
        </div>
      )}
    </div>
  );
}

function toUriArray(value: unknown): string[] {
  if (Array.isArray(value)) {
    return value.map((item) => stringValue(item)).filter(Boolean);
  }
  if (typeof value !== "string") return [];
  const trimmed = value.trim();
  if (!trimmed) return [];
  if (trimmed.startsWith("[")) {
    try {
      const parsed = JSON.parse(trimmed) as unknown;
      if (Array.isArray(parsed)) {
        return parsed.map((item) => stringValue(item)).filter(Boolean);
      }
    } catch {
      // fall through to single value
    }
  }
  return [trimmed];
}

function ImageGalleryFieldEditor({
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
  const [uploading, setUploading] = useState(false);
  const [cropTarget, setCropTarget] = useState<{
    url: string;
    index: number;
  } | null>(null);
  const values = useMemo(
    () => toUriArray(form[field.key]),
    [field.key, form],
  );

  const addLabel =
    language === "th" ? "เพิ่มรูป" : "Add image";
  const editLabel = language === "th" ? "แก้ไข" : "Edit";
  const removeLabel = language === "th" ? "ลบ" : "Remove";

  function updateValues(next: string[]) {
    setForm({ ...form, [field.key]: next });
  }

  function removeAt(index: number) {
    updateValues(values.filter((_, idx) => idx !== index));
  }

  function replaceAt(index: number, uri: string) {
    const next = values.slice();
    next[index] = uri;
    updateValues(next);
  }

  async function uploadFile(file: File, replaceIndex?: number) {
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
      uploadForm.append("category", `system-settings/${field.key}`);
      const response = await fetch("/api/upload/image", {
        method: "POST",
        headers: {
          "x-bc-backend-url": auth.backendUrl,
          Authorization: `Bearer ${auth.token}`,
        },
        body: uploadForm,
      });
      const payload = (await response.json()) as unknown;
      if (!response.ok || isFailed(payload))
        throw new Error(
          extractMessage(payload) ??
            uploadUiText(language, "imageUploadFailed"),
        );
      const uri = extractUploadUri(payload);
      if (!uri) throw new Error(uploadUiText(language, "imageUploadFailed"));
      if (typeof replaceIndex === "number") {
        replaceAt(replaceIndex, uri);
      } else {
        updateValues([...values, uri]);
      }
    } catch (uploadError) {
      setError(
        uploadError instanceof Error && uploadError.message
          ? uploadError.message
          : uploadUiText(language, "imageUploadFailed"),
      );
    } finally {
      setUploading(false);
      if (inputRef.current) inputRef.current.value = "";
    }
  }

  async function handleFile(file: File | undefined) {
    if (!file || uploading) return;
    if (!file.type.startsWith("image/") || file.size > 12 * 1024 * 1024) {
      setError(uploadUiText(language, "imageTooLarge"));
      return;
    }
    await uploadFile(file);
  }

  async function applyCroppedFile(file: File) {
    const target = cropTarget;
    setCropTarget(null);
    if (!target) return;
    await uploadFile(file, target.index);
  }

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm font-semibold md:col-span-2">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span>
          {label}
          {field.required ? " *" : ""}
        </span>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={() => inputRef.current?.click()}
          disabled={uploading || !auth}
        >
          {uploading ? (
            <Loader2 className="animate-spin" />
          ) : (
            <UploadCloud />
          )}
          {uploading ? uploadUiText(language, "uploading") : addLabel}
        </Button>
      </div>
      {values.length === 0 ? (
        <div className="grid h-20 place-items-center rounded-xl border border-dashed border-input text-xs font-normal text-muted-foreground">
          {language === "th"
            ? "ยังไม่มีรูป — กด \"เพิ่มรูป\" เพื่ออัปโหลด"
            : 'No images yet — click "Add image" to upload'}
        </div>
      ) : (
        <ul className="grid gap-2 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-4">
          {values.map((uri, index) => (
            <GalleryItemCard
              key={`${uri}-${index}`}
              auth={auth}
              editLabel={editLabel}
              language={language}
              onCrop={(displayUrl) =>
                setCropTarget({ url: displayUrl, index })
              }
              onRemove={() => removeAt(index)}
              removeLabel={removeLabel}
              uri={uri}
            />
          ))}
        </ul>
      )}
      <p className="text-xs font-normal text-muted-foreground">
        {uploadUiText(language, "imageUploadHint")}
      </p>
      {error ? (
        <p className="text-xs font-semibold text-destructive">{error}</p>
      ) : null}
      <input
        ref={inputRef}
        className="sr-only"
        type="file"
        accept="image/png,image/jpeg,image/webp,image/gif"
        onChange={(event) => void handleFile(event.target.files?.[0])}
      />
      {cropTarget ? (
        <ImageCropDialog
          imageUrl={cropTarget.url}
          language={language}
          onCancel={() => setCropTarget(null)}
          onApply={(file) => void applyCroppedFile(file)}
        />
      ) : null}
    </section>
  );
}

function GalleryItemCard({
  auth,
  editLabel,
  language,
  onCrop,
  onRemove,
  removeLabel,
  uri,
}: {
  auth: AuthSession | null;
  editLabel: string;
  language: LanguageCode;
  onCrop: (displayUrl: string) => void;
  onRemove: () => void;
  removeLabel: string;
  uri: string;
}) {
  const { displayUrl, failed, loading } = useAuthenticatedImageDisplaySource(
    uri,
    auth,
  );
  return (
    <li className="grid gap-1 rounded-xl border border-input bg-card p-1 shadow-sm">
      <div
        className="relative h-24 w-full overflow-hidden rounded-lg border border-border bg-muted bg-contain bg-center bg-no-repeat"
        style={
          displayUrl
            ? { backgroundImage: `url(${JSON.stringify(displayUrl)})` }
            : undefined
        }
      >
        {!displayUrl ? (
          <div className="grid h-full place-items-center text-muted-foreground">
            {loading ? (
              <Loader2 className="size-6 animate-spin" />
            ) : failed ? (
              <span className="px-2 text-center text-[10px] font-semibold text-destructive">
                {uploadUiText(language, "imageDisplayFailed")}
              </span>
            ) : (
              <ImageIcon className="size-6" />
            )}
          </div>
        ) : null}
      </div>
      <div className="flex flex-wrap items-center justify-between gap-1 px-1 text-[11px] text-muted-foreground">
        <span className="line-clamp-1 min-w-0 break-all" title={uri}>
          {uri}
        </span>
        <div className="flex shrink-0 gap-1">
          {displayUrl ? (
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => onCrop(displayUrl)}
              aria-label={editLabel}
            >
              <Edit3 className="size-3" />
            </Button>
          ) : null}
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={onRemove}
            aria-label={removeLabel}
          >
            <Trash2 className="size-3" />
          </Button>
        </div>
      </div>
    </li>
  );
}

function ImageGalleryReadOnlyDetail({
  auth,
  label,
  language,
  value,
}: {
  auth: AuthSession | null;
  label: string;
  language: LanguageCode;
  value: unknown;
}) {
  const uris = useMemo(() => toUriArray(value), [value]);
  if (uris.length === 0) {
    return (
      <div className="grid gap-2 rounded-xl border border-border bg-card px-3 py-2 text-sm shadow-[0_1px_2px_rgba(0,0,0,0.02)] border-l-2 border-l-secondary">
        <span className="text-xs font-semibold text-muted-foreground">
          {label}
        </span>
        <b className="text-foreground font-medium">-</b>
      </div>
    );
  }
  return (
    <div className="grid gap-2 rounded-xl border border-border bg-card px-3 py-2 text-sm shadow-[0_1px_2px_rgba(0,0,0,0.02)] border-l-2 border-l-secondary">
      <span className="text-xs font-semibold text-muted-foreground">
        {label}
      </span>
      <ul className="grid gap-2 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-4">
        {uris.map((uri, index) => (
          <GalleryReadOnlyItem
            key={`${uri}-${index}`}
            auth={auth}
            label={label}
            language={language}
            uri={uri}
          />
        ))}
      </ul>
    </div>
  );
}

function GalleryReadOnlyItem({
  auth,
  label,
  language,
  uri,
}: {
  auth: AuthSession | null;
  label: string;
  language: LanguageCode;
  uri: string;
}) {
  const { displayUrl, failed } = useAuthenticatedImageDisplaySource(
    uri,
    auth,
  );
  const [elementFailed, setElementFailed] = useState(false);
  const canShow = Boolean(displayUrl) && !failed && !elementFailed;

  useEffect(() => {
    setElementFailed(false);
  }, [displayUrl]);

  return (
    <li className="grid gap-1 rounded-lg border border-input bg-background p-1">
      <div className="relative h-24 w-full overflow-hidden rounded-md border border-border bg-muted">
        {canShow ? (
          <Image
            alt={label}
            className="object-cover"
            fill
            sizes="160px"
            src={displayUrl}
            unoptimized
            onError={() => setElementFailed(true)}
          />
        ) : (
          <div className="grid h-full place-items-center text-[10px] font-semibold text-muted-foreground">
            {failed || elementFailed
              ? uploadUiText(language, "imageDisplayFailed")
              : "…"}
          </div>
        )}
      </div>
      <span className="line-clamp-2 break-all px-1 text-[10px] text-muted-foreground" title={uri}>
        {uri}
      </span>
    </li>
  );
}

type BranchOption = {
  guid_fixed: string;
  code: string;
  names: Array<{ code?: string; name?: string }>;
  shopid?: string;
  shop_name?: string;
};

function branchOptionDisplayName(
  option: BranchOption,
  language: LanguageCode,
): string {
  const names = option.names ?? [];
  const localized =
    names.find((item) => item.code?.toLowerCase() === language && item.name)
      ?.name ??
    names.find((item) => item.code?.toLowerCase() === "th" && item.name)
      ?.name ??
    names.find((item) => item.name)?.name;
  const branchName = (localized ?? option.code ?? option.guid_fixed).trim();
  if (option.shop_name) {
    return `${option.shop_name} - ${branchName}`;
  }
  return branchName;
}

function recordToBranchOption(record: SettingRecord): BranchOption {
  const namesRaw = Array.isArray(record.names) ? record.names : [];
  const names: BranchOption["names"] = namesRaw
    .filter(isRecord)
    .map((entry) => ({
      code: stringValue(entry.code),
      name: stringValue(entry.name),
    }));
  return {
    guid_fixed: stringValue(record.guid_fixed ?? record.guidfixed ?? record.guid ?? record.guidFixed),
    code: stringValue(record.code),
    names,
  };
}

function selectedBranchesFromValue(value: unknown): BranchOption[] {
  if (!Array.isArray(value)) return [];
  return value
    .map((item) => {
      if (typeof item === "string") {
        const trimmed = item.trim();
        if (!trimmed) return null;
        return { guid_fixed: trimmed, code: "", names: [] } as BranchOption;
      }
      if (!isRecord(item)) return null;
      const namesRaw = Array.isArray(item.names) ? item.names : [];
      const names: BranchOption["names"] = namesRaw
        .filter(isRecord)
        .map((entry) => ({
          code: stringValue(entry.code),
          name: stringValue(entry.name),
        }));
      const guid = stringValue(item.guid_fixed ?? item.guidfixed ?? item.guid);
      const code = stringValue(item.code);
      if (!guid && !code) return null;
      return { guid_fixed: guid, code, names } as BranchOption;
    })
    .filter((item): item is BranchOption => item !== null);
}

function branchKeyOf(option: { guid_fixed?: string; code?: string; shopid?: string }): string {
  const shopPrefix = option.shopid ? `${option.shopid}_` : "";
  const coreKey = stringValue(option.guid_fixed) || stringValue(option.code);
  return `${shopPrefix}${coreKey}`;
}

function BranchMultiSelectFieldEditor({
  auth,
  field,
  form,
  label,
  language,
  setForm,
  workspace,
}: {
  auth: AuthSession | null;
  field: SystemSettingField;
  form: FormState;
  label: string;
  language: LanguageCode;
  setForm: (form: FormState) => void;
  workspace: WorkspaceSession | null;
}) {
  const [options, setOptions] = useState<BranchOption[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [dialogOpen, setDialogOpen] = useState(false);
  const selected = useMemo(() => {
    const rawSelected = selectedBranchesFromValue(form[field.key]);
    return rawSelected.map((item) => {
      const match = options.find((opt) => branchKeyOf(opt) === branchKeyOf(item));
      if (match) {
        return {
          ...item,
          shopid: match.shopid,
          shop_name: match.shop_name,
          names: match.names,
        };
      }
      return item;
    });
  }, [field.key, form, options]);

  useEffect(() => {
    if (!auth || !workspace) return;
    const controller = new AbortController();
    let cancelled = false;
    setLoading(true);
    setError("");

    let shopids: string[] = [];
    const formCompanies = form.company_guids;
    if (Array.isArray(formCompanies) && formCompanies.length > 0) {
      shopids = formCompanies
        .map((s) => (typeof s === "string" ? s.trim() : stringValue(s?.guid_fixed ?? s?.shopid ?? "")))
        .filter(Boolean);
    }
    if (shopids.length === 0) {
      shopids = [workspace.shop.shopid];
    }

    void fetch(`/api/workspace/shops`, {
      headers: requestHeaders(auth),
      cache: "no-store",
      signal: controller.signal,
    })
      .then(async (res) => {
        if (!res.ok) throw new Error("Load shops failed");
        return res.json() as Promise<any>;
      })
      .then(async (shopsPayload) => {
        if (cancelled) return;
        const shopsList = Array.isArray(shopsPayload.data) ? shopsPayload.data : [];
        const shopNameMap = new Map<string, string>();
        for (const s of shopsList) {
          shopNameMap.set(s.shopid, s.name1 || s.name || s.shopid);
        }

        const fetchPromises = shopids.map(async (sid) => {
          const params = new URLSearchParams({
            limit: "1000",
            offset: "0",
            shopid: sid,
          });
          const response = await fetch(`/api/system-settings/branch?${params.toString()}`, {
            headers: requestHeaders(auth),
            cache: "no-store",
            signal: controller.signal,
          });
          if (!response.ok) throw new Error(`HTTP ${response.status}`);
          const payload = await response.json();
          const records = extractListRecords(payload);
          return records
            .map(recordToBranchOption)
            .filter((opt) => opt.guid_fixed || opt.code)
            .map((opt) => ({
              ...opt,
              shopid: sid,
              shop_name: shopNameMap.get(sid) || sid,
            }));
        });

        const allBranchesResults = await Promise.all(fetchPromises);
        const mergedBranches = allBranchesResults.flat();

        if (!cancelled) {
          setOptions(mergedBranches);
          setLoading(false);
        }
      })
      .catch((catchError: unknown) => {
        if (cancelled) return;
        if (
          catchError instanceof DOMException &&
          catchError.name === "AbortError"
        )
          return;
        setError(
          catchError instanceof Error && catchError.message
            ? catchError.message
            : language === "th"
              ? "โหลดสาขาไม่สำเร็จ"
              : "Failed to load branches",
        );
        setLoading(false);
      });

    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [auth, language, workspace, form.company_guids]);

  function commitSelection(next: BranchOption[]) {
    setForm({ ...form, [field.key]: next });
  }

  function removeBranch(option: BranchOption) {
    const key = branchKeyOf(option);
    commitSelection(selected.filter((item) => branchKeyOf(item) !== key));
  }

  const summary =
    language === "th"
      ? `เลือก ${selected.length} / ${options.length} สาขา`
      : `${selected.length} / ${options.length} branches selected`;
  const pickLabel =
    language === "th" ? "เลือกสาขา" : "Pick branches";

  return (
    <section className="grid w-full gap-2 rounded-2xl border border-border bg-background p-2 text-sm font-semibold">
      <div className="flex flex-wrap items-center justify-between gap-2">
        <span>
          {label}
          {field.required ? " *" : ""}
        </span>
        <div className="flex flex-wrap items-center gap-2 text-xs font-normal text-muted-foreground">
          <span>{summary}</span>
          <Button
            type="button"
            size="sm"
            onClick={() => setDialogOpen(true)}
            disabled={loading || (options.length === 0 && !error)}
          >
            {loading ? (
              <Loader2 className="animate-spin" />
            ) : (
              <Search />
            )}
            {pickLabel}
          </Button>
        </div>
      </div>
      {loading ? (
        <div className="flex items-center gap-2 px-1 text-xs text-muted-foreground">
          <Loader2 className="size-3 animate-spin" />
          {language === "th" ? "กำลังโหลดสาขา…" : "Loading branches…"}
        </div>
      ) : null}
      {error ? (
        <p className="text-xs font-semibold text-destructive">{error}</p>
      ) : null}
      {options.length === 0 && !loading && !error ? (
        <p className="text-xs font-normal text-muted-foreground">
          {language === "th"
            ? "ยังไม่มีสาขาให้เลือก — เพิ่มสาขาในหน้า \"สาขา\" ก่อน"
            : 'No branches to choose yet — add one on the "Branch" screen first.'}
        </p>
      ) : null}
      {selected.length > 0 ? (
        <ul className="flex flex-wrap gap-1.5">
          {selected.map((option) => {
            const key = branchKeyOf(option);
            return (
              <li key={key}>
                <span className="inline-flex items-center gap-1.5 rounded-md border border-border bg-card px-2 py-1 text-xs font-medium">
                  <span className="max-w-[200px] truncate">
                    {branchOptionDisplayName(option, language)}
                  </span>
                  {option.code ? (
                    <span className="text-[10px] font-normal text-muted-foreground">
                      {option.code}
                    </span>
                  ) : null}
                  <button
                    type="button"
                    className="ml-1 grid size-4 place-items-center rounded text-muted-foreground hover:bg-muted hover:text-foreground"
                    onClick={() => removeBranch(option)}
                    aria-label={
                      language === "th" ? "ลบสาขา" : "Remove branch"
                    }
                  >
                    <X className="size-3" />
                  </button>
                </span>
              </li>
            );
          })}
        </ul>
      ) : !loading && !error && options.length > 0 ? (
        <p className="text-xs font-normal text-muted-foreground">
          {language === "th"
            ? "ยังไม่มีสาขาที่เลือก — กด \"เลือกสาขา\" เพื่อเลือก"
            : 'No branches selected yet — click "Pick branches" to choose.'}
        </p>
      ) : null}
      {dialogOpen ? (
        <BranchPickerDialog
          initialSelected={selected}
          language={language}
          onCancel={() => setDialogOpen(false)}
          onConfirm={(next) => {
            commitSelection(next);
            setDialogOpen(false);
          }}
          options={options}
        />
      ) : null}
    </section>
  );
}

function BranchPickerDialog({
  initialSelected,
  language,
  onCancel,
  onConfirm,
  options,
}: {
  initialSelected: BranchOption[];
  language: LanguageCode;
  onCancel: () => void;
  onConfirm: (selected: BranchOption[]) => void;
  options: BranchOption[];
}) {
  const [draft, setDraft] = useState<BranchOption[]>(initialSelected);
  const [query, setQuery] = useState("");
  const draftKeys = useMemo(
    () => new Set(draft.map((item) => branchKeyOf(item))),
    [draft],
  );
  const filteredOptions = useMemo(() => {
    const needle = query.trim().toLowerCase();
    if (!needle) return options;
    return options.filter((option) => {
      const name = branchOptionDisplayName(option, language).toLowerCase();
      const code = option.code.toLowerCase();
      return name.includes(needle) || code.includes(needle);
    });
  }, [language, options, query]);

  useEffect(() => {
    const handler = (event: KeyboardEvent) => {
      if (event.key === "Escape") onCancel();
    };
    window.addEventListener("keydown", handler);
    return () => window.removeEventListener("keydown", handler);
  }, [onCancel]);

  function toggleDraft(option: BranchOption, checked: boolean) {
    const key = branchKeyOf(option);
    const without = draft.filter((item) => branchKeyOf(item) !== key);
    setDraft(checked ? [...without, option] : without);
  }

  function selectAllVisible() {
    const map = new Map<string, BranchOption>();
    for (const item of draft) map.set(branchKeyOf(item), item);
    for (const item of filteredOptions) map.set(branchKeyOf(item), item);
    setDraft(Array.from(map.values()));
  }

  function clearVisible() {
    if (!query.trim()) {
      setDraft([]);
      return;
    }
    const visibleKeys = new Set(
      filteredOptions.map((item) => branchKeyOf(item)),
    );
    setDraft(draft.filter((item) => !visibleKeys.has(branchKeyOf(item))));
  }

  const title = language === "th" ? "เลือกสาขา" : "Pick branches";
  const searchPlaceholder =
    language === "th"
      ? "ค้นหารหัสหรือชื่อสาขา"
      : "Search branch code or name";
  const summary =
    language === "th"
      ? `เลือก ${draft.length} / ${options.length} สาขา (กรอง ${filteredOptions.length})`
      : `${draft.length} / ${options.length} selected (${filteredOptions.length} filtered)`;
  const selectAllLabel =
    language === "th"
      ? query.trim()
        ? "เลือกทั้งหมดที่กรอง"
        : "เลือกทุกสาขา"
      : query.trim()
        ? "Select all filtered"
        : "Select all";
  const clearLabel =
    language === "th"
      ? query.trim()
        ? "ล้างที่กรอง"
        : "ล้างทั้งหมด"
      : query.trim()
        ? "Clear filtered"
        : "Clear all";

  return (
    <div
      className="fixed inset-0 z-50 flex flex-col bg-card text-card-foreground"
      role="dialog"
      aria-modal="true"
      aria-label={title}
    >
      <header className="flex items-center justify-between gap-2 border-b border-border px-3 py-2">
        <div className="flex items-center gap-2 text-sm font-semibold">
          <GitBranch className="size-4" />
          {title}
        </div>
        <Button
          type="button"
          variant="ghost"
          size="sm"
          onClick={onCancel}
          aria-label={language === "th" ? "ยกเลิก" : "Cancel"}
        >
          <X />
        </Button>
      </header>
      <div className="flex flex-wrap items-center gap-2 border-b border-border px-3 py-2">
        <label className="relative flex min-w-0 flex-1 items-center">
          <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            autoFocus
            className="h-9 pl-9"
            placeholder={searchPlaceholder}
            value={query}
            onChange={(event) => setQuery(event.target.value)}
          />
        </label>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={selectAllVisible}
          disabled={filteredOptions.length === 0}
        >
          {selectAllLabel}
        </Button>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={clearVisible}
          disabled={draft.length === 0}
        >
          {clearLabel}
        </Button>
      </div>
      <div className="min-h-0 flex-1 overflow-y-auto px-3 py-2">
        {filteredOptions.length === 0 ? (
          <p className="grid h-full place-items-center text-sm text-muted-foreground">
            {language === "th" ? "ไม่พบสาขา" : "No branches found"}
          </p>
        ) : (
          <ul className="grid gap-1 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-4">
            {filteredOptions.map((option) => {
              const key = branchKeyOf(option);
              const checked = draftKeys.has(key);
              return (
                <li key={key}>
                  <label
                    className={cn(
                      "flex cursor-pointer items-center gap-2 rounded-md border px-2 py-1.5 text-xs font-medium transition-colors",
                      checked
                        ? "border-primary/40 bg-primary/5 hover:bg-primary/10"
                        : "border-border bg-card hover:border-primary/40 hover:bg-accent/40",
                    )}
                  >
                    <input
                      type="checkbox"
                      checked={checked}
                      onChange={(event) =>
                        toggleDraft(option, event.target.checked)
                      }
                      className="size-4 accent-primary"
                    />
                    <span className="min-w-0 flex-1 truncate">
                      {branchOptionDisplayName(option, language)}
                    </span>
                    {option.code ? (
                      <span className="shrink-0 text-[10px] text-muted-foreground">
                        {option.code}
                      </span>
                    ) : null}
                  </label>
                </li>
              );
            })}
          </ul>
        )}
      </div>
      <footer className="flex flex-wrap items-center justify-between gap-2 border-t border-border px-3 py-2">
        <span className="text-xs text-muted-foreground">{summary}</span>
        <div className="flex flex-wrap gap-2">
          <Button type="button" variant="outline" onClick={onCancel}>
            {language === "th" ? "ยกเลิก" : "Cancel"}
          </Button>
          <Button type="button" onClick={() => onConfirm(draft)}>
            <BadgeCheck />
            {language === "th" ? "ยืนยัน" : "Confirm"}
          </Button>
        </div>
      </footer>
    </div>
  );
}

function BranchMultiSelectReadOnlyDetail({
  label,
  language,
  value,
}: {
  label: string;
  language: LanguageCode;
  value: unknown;
}) {
  const selected = selectedBranchesFromValue(value);
  return (
    <div className="grid gap-2 rounded-xl border border-border bg-card px-3 py-2 text-sm shadow-[0_1px_2px_rgba(0,0,0,0.02)] border-l-2 border-l-secondary">
      <span className="text-xs font-semibold text-muted-foreground">
        {label}
      </span>
      {selected.length === 0 ? (
        <b className="text-foreground font-medium">-</b>
      ) : (
        <ul className="flex flex-wrap gap-1">
          {selected.map((option) => {
            const key = branchKeyOf(option);
            return (
              <li
                key={key}
                className="rounded-md border border-border bg-background px-2 py-0.5 text-xs font-semibold"
              >
                {branchOptionDisplayName(option, language)}
                {option.code ? (
                  <span className="ml-1 text-[10px] font-normal text-muted-foreground">
                    {option.code}
                  </span>
                ) : null}
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}

function companyOptionDisplayName(
  option: { guidfixed?: string; code?: string; names?: any },
  language: LanguageCode,
): string {
  const names = option.names ?? [];
  const localized =
    names.find((item: any) => item.code?.toLowerCase() === language && item.name)
      ?.name ??
    names.find((item: any) => item.code?.toLowerCase() === "th" && item.name)
      ?.name ??
    names.find((item: any) => item.name)?.name;
  return (localized ?? option.code ?? option.guidfixed ?? "").trim();
}

function CompanyBranchTreeSelector({
  auth,
  form,
  language,
  setForm,
  workspace,
}: {
  auth: AuthSession | null;
  form: FormState;
  language: LanguageCode;
  setForm: (form: FormState) => void;
  workspace: WorkspaceSession | null;
}) {
  const [shops, setShops] = useState<any[]>([]);
  const [branchesMap, setBranchesMap] = useState<Record<string, BranchOption[]>>({});
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState("");
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({});

  const selectedShopIds = useMemo(() => {
    const val = form.company_guids;
    if (Array.isArray(val)) {
      return val.map((item) => {
        if (typeof item === "string") return item.trim();
        if (isRecord(item)) return stringValue(item.guid_fixed ?? item.guidfixed ?? item.guid ?? "");
        return "";
      }).filter(Boolean);
    }
    return [];
  }, [form.company_guids]);

  const selectedBranches = useMemo(() => {
    const raw = selectedBranchesFromValue(form.branches);
    return raw.map((item) => {
      const itemCoreKey = stringValue(item.guid_fixed) || stringValue(item.code);
      if (!itemCoreKey) return item;
      for (const [shopid, shopBranches] of Object.entries(branchesMap)) {
        const match = shopBranches.find((br) => 
          (stringValue(br.guid_fixed) || stringValue(br.code)) === itemCoreKey
        );
        if (match) {
          return {
            ...item,
            shopid: shopid,
          };
        }
      }
      return item;
    });
  }, [form.branches, branchesMap]);

  useEffect(() => {
    if (!auth || !workspace) return;
    const controller = new AbortController();
    let cancelled = false;
    setLoading(true);
    setError("");

    void fetch(`/api/workspace/shops`, {
      headers: requestHeaders(auth),
      cache: "no-store",
      signal: controller.signal,
    })
      .then(async (res) => {
        if (!res.ok) throw new Error("Load shops failed");
        return res.json() as Promise<any>;
      })
      .then(async (shopsPayload) => {
        if (cancelled) return;
        const shopsList = Array.isArray(shopsPayload.data) ? shopsPayload.data : [];
        setShops(shopsList);

        // Fetch branches for all shops
        const fetchPromises = shopsList.map(async (shop: any) => {
          const sid = shop.shopid;
          const compGuid = String(shop.guid_fixed ?? shop.guidfixed ?? shop.guid ?? "").trim();
          const params = new URLSearchParams({
            limit: "1000",
            offset: "0",
            company_guid: compGuid,
          });
          const response = await fetch(`/api/system-settings/branch?${params.toString()}`, {
            headers: requestHeaders(auth),
            cache: "no-store",
            signal: controller.signal,
          });
          if (!response.ok) return { sid, data: [] };
          const payload = await response.json();
          const records = extractListRecords(payload);
          const parsed = records
            .map(recordToBranchOption)
            .filter((opt) => opt.guid_fixed || opt.code)
            .map((opt) => ({
              ...opt,
              shopid: sid,
            }));
          return { sid, data: parsed };
        });

        const results = await Promise.all(fetchPromises);
        const nextMap: Record<string, BranchOption[]> = {};
        for (const res of results) {
          nextMap[res.sid] = res.data;
        }

        if (!cancelled) {
          setBranchesMap(nextMap);
          setLoading(false);
        }
      })
      .catch((catchError: unknown) => {
        if (cancelled) return;
        setError(
          catchError instanceof Error && catchError.message
            ? catchError.message
            : language === "th"
              ? "โหลดข้อมูลบริษัท/สาขาไม่สำเร็จ"
              : "Failed to load organizational structure",
        );
        setLoading(false);
      });

    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [auth, language, workspace]);

  function handleToggleShop(shopId: string, checked: boolean) {
    let nextShops = [...selectedShopIds];
    if (checked) {
      if (!nextShops.includes(shopId)) nextShops.push(shopId);
    } else {
      nextShops = nextShops.filter((id) => id !== shopId);
    }

    // Toggle branches under this shop as well
    const shopBranches = branchesMap[shopId] || [];
    let nextBranches = [...selectedBranches];
    if (checked) {
      for (const br of shopBranches) {
        if (!nextBranches.some((item) => branchKeyOf(item) === branchKeyOf(br))) {
          nextBranches.push(br);
        }
      }
    } else {
      const shopBranchKeys = new Set(shopBranches.map((br) => branchKeyOf(br)));
      nextBranches = nextBranches.filter((item) => !shopBranchKeys.has(branchKeyOf(item)));
    }

    setForm({
      ...form,
      company_guids: nextShops,
      branches: nextBranches,
    });
  }

  function handleToggleBranch(shopId: string, br: BranchOption, checked: boolean) {
    const brKey = branchKeyOf(br);
    let nextBranches = [...selectedBranches];
    if (checked) {
      if (!nextBranches.some((item) => branchKeyOf(item) === brKey)) {
        nextBranches.push(br);
      }
    } else {
      nextBranches = nextBranches.filter((item) => branchKeyOf(item) !== brKey);
    }

    const shopBranches = branchesMap[shopId] || [];
    const hasAnyChecked = shopBranches.some((b) =>
      checked ? branchKeyOf(b) === brKey || nextBranches.some((item) => branchKeyOf(item) === branchKeyOf(b))
              : nextBranches.some((item) => branchKeyOf(item) === branchKeyOf(b))
    );

    let nextShops = [...selectedShopIds];
    if (hasAnyChecked) {
      if (!nextShops.includes(shopId)) nextShops.push(shopId);
    } else {
      nextShops = nextShops.filter((id) => id !== shopId);
    }

    setForm({
      ...form,
      company_guids: nextShops,
      branches: nextBranches,
    });
  }

  function toggleCollapsed(shopId: string) {
    setCollapsed((prev) => ({ ...prev, [shopId]: !prev[shopId] }));
  }

  return (
    <section className="grid w-full gap-3 rounded-2xl border border-border bg-background p-4 text-sm font-semibold shadow-sm">
      <div className="flex flex-wrap items-center justify-between gap-2 border-b border-border pb-2">
        <span>{language === "th" ? "สิทธิ์การเข้าถึงบริษัทและสาขา" : "Company & Branch Access"}</span>
        {loading ? (
          <span className="flex items-center gap-1 text-xs font-normal text-muted-foreground animate-pulse">
            <Loader2 className="size-3 animate-spin" />
            {language === "th" ? "กำลังโหลดข้อมูลโครงสร้าง…" : "Loading structure…"}
          </span>
        ) : null}
      </div>
      {error ? <p className="text-xs font-semibold text-destructive">{error}</p> : null}

      {!loading && !error && shops.length === 0 ? (
        <p className="text-xs font-normal text-muted-foreground py-2">
          {language === "th" ? "ไม่พบข้อมูลบริษัทในระบบ" : "No companies found."}
        </p>
      ) : null}

      {!loading && !error && shops.length > 0 ? (
        <div className="flex flex-col gap-2.5 py-1">
          {shops.map((shop) => {
            const sid = shop.shopid;
            const shopName = shop.names?.find((n: any) => n.code === language)?.name || shop.name1 || shop.name || sid;
            const isShopChecked = selectedShopIds.includes(sid);
            const shopBranches = branchesMap[sid] || [];
            const isCollapsed = collapsed[sid];
            const selectedShopBranchesCount = shopBranches.filter((b) =>
              selectedBranches.some((item) => branchKeyOf(item) === branchKeyOf(b))
            ).length;

            return (
              <div key={sid} className="flex flex-col gap-1.5 border border-border/60 bg-card rounded-xl p-3">
                <div className="flex items-center justify-between gap-2">
                  <div className="flex items-center gap-2.5 min-w-0">
                    <button
                      type="button"
                      onClick={() => toggleCollapsed(sid)}
                      className="p-1 hover:bg-muted rounded text-muted-foreground"
                    >
                      {isCollapsed ? (
                        <ChevronRight className="size-4" />
                      ) : (
                        <ChevronDown className="size-4" />
                      )}
                    </button>
                    <input
                      type="checkbox"
                      id={`shop-${sid}`}
                      checked={isShopChecked}
                      onChange={(e) => handleToggleShop(sid, e.target.checked)}
                      className="size-4 text-primary border-border rounded cursor-pointer"
                    />
                    <label htmlFor={`shop-${sid}`} className="font-bold cursor-pointer truncate flex items-center gap-1.5 select-none">
                      <Building2 className="size-4 text-primary shrink-0" />
                      <span>{shopName}</span>
                    </label>
                  </div>
                  {shopBranches.length > 0 ? (
                    <span className="text-xs font-normal text-muted-foreground shrink-0">
                      {language === "th"
                        ? `เลือก ${selectedShopBranchesCount} / ${shopBranches.length} สาขา`
                        : `${selectedShopBranchesCount} / ${shopBranches.length} selected`}
                    </span>
                  ) : null}
                </div>

                {!isCollapsed && shopBranches.length > 0 ? (
                  <div className="pl-9 border-l border-border/80 ml-3.5 mt-1 flex flex-col gap-2">
                    {shopBranches.map((br) => {
                      const brKey = branchKeyOf(br);
                      const isBrChecked = selectedBranches.some((item) => branchKeyOf(item) === brKey);
                      const brName = branchOptionDisplayName(br, language);

                      return (
                        <label
                          key={brKey}
                          className="flex items-center gap-2.5 cursor-pointer py-0.5 select-none font-normal text-muted-foreground hover:text-foreground transition-colors"
                        >
                          <input
                            type="checkbox"
                            checked={isBrChecked}
                            onChange={(e) => handleToggleBranch(sid, br, e.target.checked)}
                            className="size-3.5 text-primary border-border rounded cursor-pointer"
                          />
                          <GitBranch className="size-3.5 text-sky-500 shrink-0" />
                          <span className="text-xs">{brName}</span>
                        </label>
                      );
                    })}
                  </div>
                ) : null}
              </div>
            );
          })}
        </div>
      ) : null}
    </section>
  );
}

function CompanyMultiSelectReadOnlyDetail({
  label,
  language,
  value,
  auth,
}: {
  label: string;
  language: LanguageCode;
  value: unknown;
  auth: AuthSession | null;
}) {
  const [options, setOptions] = useState<MasterEntry[]>([]);
  const selectedGuids = useMemo(() => {
    if (Array.isArray(value)) {
      return value.map((item) => {
        if (typeof item === "string") return item.trim();
        if (isRecord(item)) return stringValue(item.guid_fixed ?? item.guidfixed ?? item.guid ?? "");
        return "";
      }).filter(Boolean);
    }
    return [];
  }, [value]);

  useEffect(() => {
    if (!auth || selectedGuids.length === 0) return;
    const controller = new AbortController();
    void fetch(`/api/workspace/shops`, {
      headers: requestHeaders(auth),
      cache: "no-store",
      signal: controller.signal,
    })
      .then(async (response) => {
        if (response.ok) return response.json();
      })
      .then((payload) => {
        if (payload && payload.success && Array.isArray(payload.data)) {
          const parsed = payload.data.map((shop: any) => ({
            guidfixed: shop.shopid,
            code: "",
            names: shop.names || [{ code: "th", name: shop.name1 || shop.name || "" }]
          }));
          setOptions(parsed);
        }
      })
      .catch(() => {});
    return () => controller.abort();
  }, [auth, selectedGuids.length]);

  const selectedOptions = useMemo(() => {
    return selectedGuids.map((guid) => {
      const match = options.find((opt) => opt.guidfixed === guid);
      if (match) return match;
      return { guidfixed: guid, code: "", names: [] } as MasterEntry;
    });
  }, [selectedGuids, options]);

  return (
    <div className="grid gap-2 rounded-xl border border-border bg-card px-3 py-2 text-sm shadow-[0_1px_2px_rgba(0,0,0,0.02)] border-l-2 border-l-secondary">
      <span className="text-xs font-semibold text-muted-foreground">
        {label}
      </span>
      {selectedGuids.length === 0 ? (
        <b className="text-foreground font-medium">-</b>
      ) : (
        <ul className="flex flex-wrap gap-1">
          {selectedOptions.map((option) => {
            return (
              <li
                key={option.guidfixed}
                className="rounded-md border border-border bg-background px-2 py-0.5 text-xs font-semibold"
              >
                {companyOptionDisplayName(option, language)}
                {option.code ? (
                  <span className="ml-1 text-[10px] font-normal text-muted-foreground">
                    ({option.code})
                  </span>
                ) : null}
              </li>
            );
          })}
        </ul>
      )}
    </div>
  );
}

function extractListRecords(payload: unknown): SettingRecord[] {
  if (!isRecord(payload)) return Array.isArray(payload) ? payload.filter(isRecord) : [];
  const data = payload.data;
  if (Array.isArray(data)) return data.filter(isRecord);
  if (isRecord(data) && Array.isArray(data.data)) return data.data.filter(isRecord);
  return [];
}

function BranchCoordinatePairEditor({
  config,
  form,
  language,
  setForm,
}: {
  config: SystemSettingConfig;
  form: FormState;
  language: LanguageCode;
  setForm: (form: FormState) => void;
}) {
  const [mapOpen, setMapOpen] = useState(false);
  const latitudeField = config.fields.find(
    (item) => item.key === "contact.latitude",
  );
  const longitudeField = config.fields.find(
    (item) => item.key === "contact.longitude",
  );
  const latitudeLabel = latitudeField
    ? (latitudeField.label[language] ?? latitudeField.label.en ?? "Latitude")
    : "Latitude";
  const longitudeLabel = longitudeField
    ? (longitudeField.label[language] ?? longitudeField.label.en ?? "Longitude")
    : "Longitude";
  const latitudeRaw = stringValue(form["contact.latitude"]);
  const longitudeRaw = stringValue(form["contact.longitude"]);
  const pickLabel =
    language === "th" ? "เลือกจากแผนที่" : "Pick on map";

  return (
    <section className="grid gap-2 rounded-2xl border border-border bg-background p-2 text-sm font-semibold md:col-span-2">
      <div className="grid gap-2 md:grid-cols-[1fr_1fr_auto] md:items-end">
        <label className="grid gap-1 text-sm font-semibold">
          <span>{latitudeLabel}</span>
          <Input
            type="number"
            inputMode="decimal"
            step="any"
            value={latitudeRaw}
            placeholder="13.756300"
            onChange={(event) =>
              setForm({ ...form, "contact.latitude": event.target.value })
            }
          />
        </label>
        <label className="grid gap-1 text-sm font-semibold">
          <span>{longitudeLabel}</span>
          <Input
            type="number"
            inputMode="decimal"
            step="any"
            value={longitudeRaw}
            placeholder="100.501800"
            onChange={(event) =>
              setForm({ ...form, "contact.longitude": event.target.value })
            }
          />
        </label>
        <Button
          type="button"
          variant="outline"
          onClick={() => setMapOpen(true)}
        >
          <MapPin />
          {pickLabel}
        </Button>
      </div>
      <MapPickerDialog
        open={mapOpen}
        initialLat={parseCoordinateValue(latitudeRaw)}
        initialLng={parseCoordinateValue(longitudeRaw)}
        language={language}
        onCancel={() => setMapOpen(false)}
        onSelect={(lat, lng) => {
          setForm({
            ...form,
            "contact.latitude": lat.toFixed(6),
            "contact.longitude": lng.toFixed(6),
          });
          setMapOpen(false);
        }}
      />
    </section>
  );
}

type BranchUnifiedViewProps = {
  auth: AuthSession | null;
  config: SystemSettingConfig;
  dateTimeScope: DateTimeScope;
  dictionary: BackendLanguageDictionary;
  editing: SettingRecord | null;
  form: FormState;
  formOpen: boolean;
  initialBackendLanguage?: BackendLanguageDictionary;
  initialBackendUrl?: string;
  language: LanguageCode;
  loading: boolean;
  notice: Notice;
  onCloseForm: () => void;
  onDelete: (record: SettingRecord) => void;
  onOpenCreate: () => void;
  onRefresh: () => void;
  onSelect: (record: SettingRecord) => void;
  onSubmit: (event?: FormEvent<HTMLFormElement>) => void;
  query: string;
  records: SettingRecord[];
  saving: boolean;
  selectedRecord: SettingRecord | null;
  setForm: (form: FormState) => void;
  setQuery: (query: string) => void;
  text: (key: keyof typeof uiEn) => string;
  workspace: WorkspaceSession | null;
};

function BranchUnifiedView({
  auth,
  config,
  dateTimeScope,
  dictionary,
  editing,
  form,
  formOpen,
  initialBackendLanguage,
  initialBackendUrl,
  language,
  loading,
  notice,
  onCloseForm,
  onDelete,
  onOpenCreate,
  onRefresh,
  onSelect,
  onSubmit,
  query,
  records,
  saving,
  selectedRecord,
  setForm,
  setQuery,
  text,
  workspace,
}: BranchUnifiedViewProps) {
  const [activeTab, setActiveTab] = useState<BranchTabId>("general");
  const editingId = editing ? recordId(editing, config) : "";
  const selectedId = selectedRecord
    ? recordId(selectedRecord, config)
    : editingId;
  const activeBranchListItem = useMemo<BranchListItem | null>(() => {
    const source = editing ?? selectedRecord;
    if (!source) return null;
    return branchRecordToListItem(source);
  }, [editing, selectedRecord]);

  useEffect(() => {
    setActiveTab("general");
  }, [selectedId]);

  const grouped = useMemo(() => {
    const buckets: Record<BranchTabId, SystemSettingField[]> = {
      general: [],
      address: [],
      department: [],
      workday: [],
      holiday: [],
      pos: [],
      business: [],
    };
    for (const field of config.fields) {
      if (isThailandAddressSecondaryField(config, field)) continue;
      if (isBranchLongitudeField(config, field)) continue;
      buckets[branchTabForField(field.key)].push(field);
    }
    return buckets;
  }, [config]);

  const editorEnabled = Boolean(editing);
  const isSubScreenTab =
    activeTab === "department" ||
    activeTab === "workday" ||
    activeTab === "holiday";
  const showFormFields =
    !isSubScreenTab && editorEnabled && grouped[activeTab].length > 0;
  const subScreenRoute: Record<BranchTabId, string | null> = {
    general: null,
    address: null,
    department: "/department",
    workday: "/work_day_screen",
    holiday: "/holiday_screen",
    pos: null,
    business: null,
  };
  const subRoute = subScreenRoute[activeTab];

  return (
    <div className="grid w-full min-w-0 gap-2 md:grid-cols-[280px_minmax(0,1fr)]">
      <aside className="grid min-w-0 content-start gap-2 rounded-2xl border border-border bg-card p-2 text-sm shadow-sm md:sticky md:top-2 md:max-h-[calc(100dvh-1rem)] md:overflow-y-auto">
        <div className="flex items-center justify-between gap-2">
          <span className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            {text("branch")}
          </span>
          <Button
            type="button"
            size="sm"
            onClick={onOpenCreate}
            disabled={!auth}
          >
            <Plus />
            {text("add")}
          </Button>
        </div>
        <label className="relative block">
          <Search className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground" />
          <Input
            className="h-8 pl-9"
            placeholder={`${text("search")} ${branchDisplayLabel(language)}`}
            value={query}
            onChange={(event) => setQuery(event.target.value)}
          />
        </label>
        <Button
          type="button"
          variant="outline"
          size="sm"
          onClick={onRefresh}
          disabled={loading || !auth}
        >
          {loading ? (
            <Loader2 className="animate-spin" />
          ) : (
            <RefreshCcw />
          )}
          {text("refresh")}
        </Button>
        <ul className="grid gap-1">
          {records.length === 0 && !loading ? (
            <li className="rounded-lg border border-dashed border-border p-2 text-center text-xs text-muted-foreground">
              {text("empty")}
            </li>
          ) : null}
          {records.map((record) => {
            const id = recordId(record, config);
            const isActive = id === selectedId;
            const code = stringValue(record.code);
            const displayName = recordTitle(record, config, language) || id;
            return (
              <li key={id || displayName}>
                <button
                  type="button"
                  className={cn(
                    "grid w-full gap-0.5 rounded-lg border border-transparent px-2 py-1.5 text-left text-sm transition-colors",
                    isActive
                      ? "border-primary/40 bg-primary/10 text-primary"
                      : "hover:border-border hover:bg-muted",
                  )}
                  onClick={() => onSelect(record)}
                >
                  <span className="truncate font-semibold">{displayName}</span>
                  {code ? (
                    <span className="text-[11px] text-muted-foreground">
                      {language === "th" ? "รหัส" : "Code"}: {code}
                    </span>
                  ) : null}
                </button>
              </li>
            );
          })}
        </ul>
      </aside>

      <main className="grid min-w-0 gap-2">
        {notice ? (
          <div
            className={`message ${notice.type === "success" ? "success" : notice.type === "error" ? "error" : "info"}`}
            role={notice.type === "error" ? "alert" : "status"}
          >
            {notice.type === "success" ? (
              <BadgeCheck size={18} />
            ) : (
              <AlertCircle size={18} />
            )}
            <span>{notice.text}</span>
          </div>
        ) : null}

        {!editorEnabled ? (
          <Card>
            <CardContent className="grid min-h-40 place-items-center p-4 text-center">
              <div className="grid gap-2">
                <Loader2
                  className={cn(
                    "mx-auto size-8 text-muted-foreground",
                    loading ? "animate-spin" : "",
                  )}
                />
                <p className="text-sm text-muted-foreground">
                  {language === "th"
                    ? "เลือกสาขาทางซ้ายเพื่อแก้ไข หรือกดเพิ่มสาขาใหม่"
                    : "Select a branch on the left, or add a new branch."}
                </p>
              </div>
            </CardContent>
          </Card>
        ) : (
          <Card className="overflow-hidden">
            <div className="flex flex-wrap items-center gap-1 border-b border-border bg-muted/40 px-2 py-1.5">
              {BRANCH_TABS.map((tab) => {
                const isActive = activeTab === tab.id;
                return (
                  <button
                    key={tab.id}
                    type="button"
                    className={cn(
                      "rounded-md px-2.5 py-1 text-xs font-semibold transition-colors",
                      isActive
                        ? "bg-primary text-primary-foreground shadow-sm"
                        : "text-foreground hover:bg-background",
                    )}
                    onClick={() => setActiveTab(tab.id)}
                  >
                    {branchTabLabel(tab.id, language)}
                  </button>
                );
              })}
            </div>

            {showFormFields ? (
              <form
                onSubmit={(event) => {
                  event.preventDefault();
                  onSubmit(event);
                }}
                className="grid gap-3 p-3"
              >
                {activeTab === "pos" ? (
                  <BranchPosSections
                    auth={auth}
                    config={config}
                    dateTimeScope={dateTimeScope}
                    dictionary={dictionary}
                    fields={grouped[activeTab]}
                    form={form}
                    language={language}
                    setForm={setForm}
                    workspace={workspace}
                  />
                ) : activeTab === "business" ? (
                  <BranchBusinessFlags
                    auth={auth}
                    config={config}
                    dateTimeScope={dateTimeScope}
                    dictionary={dictionary}
                    fields={grouped[activeTab]}
                    form={form}
                    language={language}
                    setForm={setForm}
                    workspace={workspace}
                  />
                ) : (
                  <div className="grid gap-2 md:grid-cols-2">
                    {grouped[activeTab].map((field) => (
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
                )}
                <footer className="flex flex-wrap items-center justify-end gap-2 border-t border-border pt-2">
                  <Button
                    type="button"
                    variant="outline"
                    onClick={onCloseForm}
                    disabled={saving}
                  >
                    {text("close")}
                  </Button>
                  <Button type="submit" disabled={saving || !auth}>
                    {saving ? (
                      <Loader2 className="animate-spin" />
                    ) : (
                      <BadgeCheck />
                    )}
                    {text("save")}
                  </Button>
                </footer>
              </form>
            ) : null}

            {isSubScreenTab && activeBranchListItem && subRoute ? (
              <div className="p-2">
                <BranchEmbeddedSubScreen
                  branch={activeBranchListItem}
                  initialBackendLanguage={initialBackendLanguage}
                  initialBackendUrl={initialBackendUrl}
                  language={language}
                  route={subRoute}
                />
              </div>
            ) : null}

            {isSubScreenTab && !activeBranchListItem ? (
              <div className="p-4 text-sm text-muted-foreground">
                {language === "th"
                  ? "เลือกสาขาก่อนเข้าใช้แท็บนี้"
                  : "Select a branch before opening this tab."}
              </div>
            ) : null}
          </Card>
        )}

        {editing && !isSubScreenTab ? (
          <div className="flex flex-wrap items-center justify-between gap-2 px-1 text-xs text-muted-foreground">
            <span>
              {language === "th" ? "รหัส" : "Code"}: {stringValue(editing.code) || "-"} · {text("branch")}: {recordTitle(editing, config, language)}
            </span>
            <Button
              type="button"
              variant="ghost"
              size="sm"
              onClick={() => onDelete(editing)}
              disabled={saving}
            >
              <Trash2 />
              {text("delete")}
            </Button>
          </div>
        ) : null}
      </main>

      {formOpen && !editing ? (
        <SettingFormDialog
          auth={auth}
          config={config}
          dictionary={dictionary}
          editing={editing}
          form={form}
          dateTimeScope={dateTimeScope}
          language={language}
          notice={notice}
          saving={saving}
          setForm={setForm}
          workspace={workspace}
          onClose={onCloseForm}
          onSubmit={onSubmit}
          text={text}
        />
      ) : null}
    </div>
  );
}

function BranchEmbeddedSubScreen({
  branch,
  initialBackendLanguage,
  initialBackendUrl,
  language,
  route,
}: {
  branch: BranchListItem;
  initialBackendLanguage?: BackendLanguageDictionary;
  initialBackendUrl?: string;
  language: LanguageCode;
  route: string;
}) {
  return (
    <SystemSettingsScreen
      branchOverride={branch}
      embedded
      hideChrome
      initialBackendLanguage={initialBackendLanguage}
      initialBackendUrl={initialBackendUrl}
      language={language}
      route={route}
    />
  );
}

type ProductGroupTabKey =
  | "productgroup"
  | "master_group_screen"
  | "master_group_sub1_screen"
  | "master_group_sub2_screen";

const PRODUCT_GROUP_TABS: ReadonlyArray<{
  id: ProductGroupTabKey;
  route: string;
  labels: Record<LanguageCode, string>;
}> = [
  {
    id: "productgroup",
    route: "/productgroup",
    labels: {
      th: "กลุ่มสินค้า",
      en: "Product Group",
      cn: "产品组",
      ja: "商品グループ",
      ko: "상품 그룹",
      lo: "ກຸ່ມສິນຄ້າ",
      my: "ကုန်ပစ္စည်းအုပ်စု",
      km: "ក្រុមផលិតផល",
      vi: "Nhóm sản phẩm",
      ms: "Kumpulan Produk",
      id: "Grup Produk",
      fil: "Grupo ng Produkto",
    },
  },
  {
    id: "master_group_screen",
    route: "/master_group_screen",
    labels: {
      th: "กลุ่มหลัก",
      en: "Main Group",
      cn: "主组",
      ja: "メイングループ",
      ko: "메인 그룹",
      lo: "ກຸ່ມຫຼັກ",
      my: "ပင်မ အုပ်စု",
      km: "ក្រុមមេ",
      vi: "Nhóm chính",
      ms: "Kumpulan Utama",
      id: "Grup Utama",
      fil: "Pangunahing Grupo",
    },
  },
  {
    id: "master_group_sub1_screen",
    route: "/master_group_sub1_screen",
    labels: {
      th: "กลุ่มย่อย 1",
      en: "Sub Group 1",
      cn: "子组 1",
      ja: "サブグループ 1",
      ko: "하위 그룹 1",
      lo: "ກຸ່ມຍ່ອຍ 1",
      my: "အောက်အုပ်စု ၁",
      km: "ក្រុមរង ១",
      vi: "Nhóm phụ 1",
      ms: "Kumpulan Kecil 1",
      id: "Sub Grup 1",
      fil: "Sub Grupo 1",
    },
  },
  {
    id: "master_group_sub2_screen",
    route: "/master_group_sub2_screen",
    labels: {
      th: "กลุ่มย่อย 2",
      en: "Sub Group 2",
      cn: "子组 2",
      ja: "サブグループ 2",
      ko: "하위 그룹 2",
      lo: "ກຸ່ມຍ່ອຍ 2",
      my: "အောက်အုပ်စု ၂",
      km: "ក្រុមរង ២",
      vi: "Nhóm phụ 2",
      ms: "Kumpulan Kecil 2",
      id: "Sub Grup 2",
      fil: "Sub Grupo 2",
    },
  },
];

function productGroupTabLabel(
  tab: (typeof PRODUCT_GROUP_TABS)[number],
  language: LanguageCode,
): string {
  return tab.labels[language] ?? tab.labels.en ?? tab.id;
}

function ProductGroupUnifiedView({
  initialBackendLanguage,
  initialBackendUrl,
  language,
}: {
  initialBackendLanguage?: BackendLanguageDictionary;
  initialBackendUrl?: string;
  language: LanguageCode;
}) {
  const [activeTab, setActiveTab] = useState<ProductGroupTabKey>(
    "productgroup",
  );
  const activeTabEntry =
    PRODUCT_GROUP_TABS.find((tab) => tab.id === activeTab) ??
    PRODUCT_GROUP_TABS[0];
  return (
    <div className="grid w-full min-w-0 gap-2">
      <Card className="overflow-hidden">
        <div className="flex flex-wrap items-center gap-1 border-b border-border bg-muted/40 px-2 py-1.5">
          {PRODUCT_GROUP_TABS.map((tab) => {
            const isActive = activeTab === tab.id;
            return (
              <button
                key={tab.id}
                type="button"
                className={cn(
                  "rounded-md px-2.5 py-1 text-xs font-semibold transition-colors",
                  isActive
                    ? "bg-primary text-primary-foreground shadow-sm"
                    : "text-foreground hover:bg-background",
                )}
                onClick={() => setActiveTab(tab.id)}
              >
                {productGroupTabLabel(tab, language)}
              </button>
            );
          })}
        </div>
        <div className="p-2">
          <SystemSettingsScreen
            embedded
            hideChrome
            initialBackendLanguage={initialBackendLanguage}
            initialBackendUrl={initialBackendUrl}
            language={language}
            route={activeTabEntry.route}
          />
        </div>
      </Card>
    </div>
  );
}

function branchRecordToListItem(record: SettingRecord): BranchListItem {
  return {
    guid_fixed: stringValue(record.guid_fixed),
    code: stringValue(record.code),
    names: Array.isArray(record.names)
      ? (record.names as BranchListItem["names"])
      : undefined,
    companynames: Array.isArray(record.companynames)
      ? (record.companynames as BranchListItem["companynames"])
      : undefined,
    base_currency: stringValue(record.base_currency) || undefined,
    language: stringValue(record.language) || undefined,
    timezone: stringValue(record.timezone) || undefined,
    timezone_offset: stringValue(record.timezone_offset) || undefined,
    timezone_label: stringValue(record.timezone_label) || undefined,
    year_type: stringValue(record.year_type) || undefined,
  };
}

function branchDisplayLabel(language: LanguageCode): string {
  return language === "th" ? "สาขา" : "branch";
}

type BranchTabSectionProps = {
  auth: AuthSession | null;
  config: SystemSettingConfig;
  dateTimeScope: DateTimeScope;
  dictionary: BackendLanguageDictionary;
  fields: SystemSettingField[];
  form: FormState;
  language: LanguageCode;
  setForm: (form: FormState) => void;
  workspace: WorkspaceSession | null;
};

function BranchPosSections({
  auth,
  config,
  dateTimeScope,
  dictionary,
  fields,
  form,
  language,
  setForm,
  workspace,
}: BranchTabSectionProps) {
  const fieldByKey = useMemo(() => {
    const map = new Map<string, SystemSettingField>();
    for (const field of fields) map.set(field.key, field);
    return map;
  }, [fields]);
  const sectionEntries = POS_TAB_SECTIONS.map((section) => ({
    section,
    sectionFields: section.fieldKeys
      .map((key) => fieldByKey.get(key))
      .filter((item): item is SystemSettingField => Boolean(item)),
  })).filter((entry) => entry.sectionFields.length > 0);
  const claimedKeys = new Set(
    sectionEntries.flatMap((entry) =>
      entry.sectionFields.map((field) => field.key),
    ),
  );
  const orphanFields = fields.filter((field) => !claimedKeys.has(field.key));

  return (
    <div className="grid gap-3">
      {sectionEntries.map(({ section, sectionFields }) => (
        <section
          key={section.id}
          className="grid gap-2 rounded-xl border border-border bg-background/40 p-3"
        >
          <header className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            {posSectionTitle(section, language)}
          </header>
          <div className="grid gap-2 md:grid-cols-2">
            {sectionFields.map((field) => (
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
      ))}
      {orphanFields.length > 0 ? (
        <section className="grid gap-2 rounded-xl border border-border bg-background/40 p-3">
          <header className="text-xs font-semibold uppercase tracking-wider text-muted-foreground">
            {language === "th" ? "อื่น ๆ" : "Other"}
          </header>
          <div className="grid gap-2 md:grid-cols-2">
            {orphanFields.map((field) => (
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
      ) : null}
    </div>
  );
}

function BranchBusinessFlags({
  auth,
  config,
  dateTimeScope,
  dictionary,
  fields,
  form,
  language,
  setForm,
  workspace,
}: BranchTabSectionProps) {
  return (
    <div className="grid gap-2 sm:grid-cols-2 md:grid-cols-3 xl:grid-cols-4">
      {fields.map((field) => (
        <div className="min-w-0" key={field.key}>
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
  const options = useMemo(
    () => comboOptionsForField(field, language, value),
    [field, language, value],
  );
  const currentOption = options.find((option) => option.value === value);
  const needle = query.trim().toLowerCase();
  const visibleOptions = needle
    ? options.filter((option) =>
        `${option.value} ${optionLabel(option, language)}`
          .toLowerCase()
          .includes(needle),
      )
    : options;

  const updatePlacement = useCallback(() => {
    if (!open || typeof window === "undefined") return;
    const rect = buttonRef.current?.getBoundingClientRect();
    if (!rect) return;

    const margin = COMBO_VIEWPORT_MARGIN;
    const viewportWidth = window.innerWidth;
    const viewportHeight = window.innerHeight;
    const availableWidth = Math.max(120, viewportWidth - margin * 2);
    const width = Math.min(
      Math.max(rect.width, Math.min(240, availableWidth)),
      availableWidth,
    );
    const maxLeft = Math.max(margin, viewportWidth - margin - width);
    const left = Math.min(Math.max(rect.left, margin), maxLeft);
    const below = Math.max(
      0,
      viewportHeight - rect.bottom - COMBO_GAP - margin,
    );
    const above = Math.max(0, rect.top - COMBO_GAP - margin);
    const bestSpace = Math.max(below, above);

    let top: number;
    let maxHeight: number;
    if (bestSpace < COMBO_MIN_USABLE_HEIGHT) {
      top = margin;
      maxHeight = Math.max(
        COMBO_MIN_USABLE_HEIGHT,
        viewportHeight - margin * 2,
      );
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
      typeof ResizeObserver !== "undefined" && buttonRef.current
        ? new ResizeObserver(scheduleUpdate)
        : null;

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
      const prefix = field.key.includes(".")
        ? `${field.key.split(".").slice(0, -1).join(".")}.`
        : "";
      nextForm[`${prefix}timezone_label`] = meta.label;
      nextForm[`${prefix}timezone_offset`] = meta.offset;
    }
    if (field.optionSource === "countries")
      applyCountryDefaultsToForm(nextForm, field.key, option.value);
    setForm(nextForm);
    setQuery("");
    setOpen(false);
  }

  return (
    <div className="grid gap-1 text-sm font-semibold">
      <span>
        {label}
        {field.required ? " *" : ""}
      </span>
      <button
        ref={buttonRef}
        aria-expanded={open}
        className="flex min-h-10 w-full items-center justify-between gap-2 rounded-2xl border border-input bg-background px-3 text-left text-sm text-foreground shadow-sm outline-none focus-visible:ring-2 focus-visible:ring-ring"
        onClick={() => setOpen((current) => !current)}
        type="button"
      >
        <span
          className={cn("min-w-0 truncate", !value && "text-muted-foreground")}
        >
          {currentOption
            ? optionLabel(currentOption, language)
            : value || field.placeholder || ""}
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
            {visibleOptions.length ? (
              visibleOptions.map((option) => (
                <button
                  className={cn(
                    "flex min-h-8 w-full items-center justify-between gap-2 rounded-xl px-2 py-1 text-left text-sm font-medium hover:bg-accent hover:text-accent-foreground",
                    option.value === value && "bg-primary/10 text-primary",
                  )}
                  key={option.value}
                  onClick={() => choose(option)}
                  type="button"
                >
                  <span className="min-w-0 truncate">
                    {optionLabel(option, language)}
                  </span>
                  {option.value === value ? (
                    <Check className="size-4 shrink-0" />
                  ) : null}
                </button>
              ))
            ) : (
              <div className="rounded-xl px-2 py-3 text-sm text-muted-foreground">
                {language === "th" ? "ไม่พบข้อมูล" : "No options found"}
              </div>
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
  const totalRanges = workDays.reduce(
    (sum, day) =>
      sum +
      (day.isactive
        ? day.fullday
          ? 1
          : Math.max(day.worktimes.length, 1)
        : 0),
    0,
  );
  const invalidCount = workDays.reduce(
    (sum, day) =>
      sum +
      day.worktimes.filter((time, timeIndex) =>
        getWorkTimeIssue(day, time, timeIndex),
      ).length,
    0,
  );

  function updateDay(index: number, patch: Partial<WorkDay>) {
    setWorkDays(
      workDays.map((day, dayIndex) =>
        dayIndex === index ? { ...day, ...patch } : day,
      ),
    );
  }

  function updateTime(
    dayIndex: number,
    timeIndex: number,
    key: keyof WorkDayTime,
    value: string,
  ) {
    const day = workDays[dayIndex];
    const nextTimes = day.worktimes.length
      ? [...day.worktimes]
      : [createWorkTime("", "", dateTimeScope)];
    const currentTime =
      nextTimes[timeIndex] ?? createWorkTime("", "", dateTimeScope);
    nextTimes[timeIndex] = withWorkTimePatch(
      currentTime,
      key,
      value,
      dateTimeScope,
    );
    updateDay(dayIndex, { worktimes: nextTimes });
  }

  function addTimeRange(dayIndex: number) {
    const day = workDays[dayIndex];
    updateDay(dayIndex, {
      fullday: false,
      isactive: true,
      worktimes: [
        ...(day.worktimes.length ? day.worktimes : []),
        createWorkTime("", "", dateTimeScope),
      ],
    });
  }

  function removeTimeRange(dayIndex: number, timeIndex: number) {
    const day = workDays[dayIndex];
    const nextTimes = day.worktimes.filter((_, index) => index !== timeIndex);
    updateDay(dayIndex, {
      worktimes: nextTimes.length
        ? nextTimes
        : [createWorkTime("", "", dateTimeScope)],
    });
  }

  function copyMondaySchedule() {
    const monday = workDays[0];
    if (!monday) return;
    setWorkDays(
      workDays.map((day, index) =>
        index === 0 || !day.isactive
          ? day
          : {
              ...day,
              fullday: monday.fullday,
              worktimes: cloneWorkTimes(monday.worktimes),
            },
      ),
    );
  }

  function toggleFullDay(index: number, checked: boolean) {
    const day = workDays[index];
    updateDay(index, {
      fullday: checked,
      worktimes: day.worktimes.length
        ? day.worktimes
        : [createWorkTime("", "", dateTimeScope)],
    });
  }

  return (
    <Card className="w-full overflow-hidden">
      <CardHeader className="p-3 pb-1">
        <div className="flex flex-wrap items-center justify-between gap-2">
          <div className="min-w-0">
            <CardTitle>{text("workTime")}</CardTitle>
            <CardDescription className="truncate">
              {text("updated")}: restaurant settings code workDay
            </CardDescription>
          </div>
          <div className="flex w-full flex-wrap gap-2 sm:w-auto">
            <Badge
              variant={invalidCount ? "warning" : "success"}
              className="min-h-8 justify-center px-3"
            >
              {text("active")}: {activeDays}
            </Badge>
            <Badge variant="outline" className="min-h-8 justify-center px-3">
              {text("totalRanges")}: {totalRanges}
            </Badge>
            <Button
              className="w-full sm:w-auto"
              size="sm"
              type="button"
              variant="outline"
              onClick={copyMondaySchedule}
              disabled={loading || !workDays[0]}
            >
              <Copy />
              {text("copyMondaySchedule")}
            </Button>
            <Button
              className="w-full sm:w-auto"
              size="sm"
              type="button"
              variant="outline"
              onClick={onRefresh}
              disabled={loading}
            >
              {loading ? <Loader2 className="animate-spin" /> : <RefreshCcw />}
              {text("refresh")}
            </Button>
            <Button
              className="w-full sm:w-auto"
              size="sm"
              type="button"
              onClick={onSave}
              disabled={saving || invalidCount > 0}
            >
              {saving ? <Loader2 className="animate-spin" /> : <Save />}
              {text("save")}
            </Button>
          </div>
        </div>
      </CardHeader>
      <CardContent className="grid gap-2 p-3">
        {workDays.length ? (
          workDays.map((day, index) => {
            const times = day.worktimes.length
              ? day.worktimes
              : [createWorkTime("", "", dateTimeScope)];
            return (
              <div
                className={cn(
                  "grid gap-2 rounded-2xl border border-border bg-background p-2",
                  !day.isactive && "bg-muted/30",
                )}
                key={day.code}
              >
                <div className="flex flex-wrap items-center justify-between gap-2">
                  <label className="flex min-w-40 flex-1 items-center gap-2 text-sm font-semibold">
                    <input
                      className="size-4 accent-primary"
                      type="checkbox"
                      checked={day.isactive}
                      onChange={(event) =>
                        updateDay(index, { isactive: event.target.checked })
                      }
                    />
                    <span className="truncate">
                      {dayNames[language]?.[index] ?? day.name}
                    </span>
                    <Badge variant={day.isactive ? "success" : "outline"}>
                      {day.isactive ? text("active") : text("off")}
                    </Badge>
                  </label>
                  {day.isactive ? (
                    <div className="flex w-full flex-wrap items-center gap-2 sm:w-auto">
                      <label className="flex h-8 items-center gap-2 rounded-xl border border-border bg-card px-2 text-xs font-semibold">
                        <input
                          className="size-4 accent-primary"
                          type="checkbox"
                          checked={day.fullday}
                          onChange={(event) =>
                            toggleFullDay(index, event.target.checked)
                          }
                        />
                        {text("fullDay")}
                      </label>
                      {!day.fullday ? (
                        <Button
                          size="sm"
                          type="button"
                          variant="outline"
                          onClick={() => addTimeRange(index)}
                        >
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
                        <div
                          className={cn(
                            "grid gap-1 rounded-xl border border-border bg-card p-2",
                            issue &&
                              "border-amber-300 bg-amber-50/70 dark:border-amber-900 dark:bg-amber-950/30",
                          )}
                          key={`${day.code}-${timeIndex}`}
                        >
                          <div className="flex items-center justify-between gap-2">
                            <b className="text-xs">
                              {text("range")} {timeIndex + 1}
                            </b>
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
                              utcPreview={formatUtcPreview(
                                time.starttimeutc,
                                time.startdayoffsetutc,
                              )}
                              value={workTimeStart(time)}
                              onChange={(event) =>
                                updateTime(
                                  index,
                                  timeIndex,
                                  "starttime",
                                  event.target.value,
                                )
                              }
                            />
                            <TimeField
                              aria-label={`${dayNames[language]?.[index] ?? day.name} ${text("range")} ${timeIndex + 1} ${text("workTime")} end`}
                              label={text("endTime")}
                              timezoneLabel={dateTimeScope.timezone_offset}
                              utcPreview={formatUtcPreview(
                                time.endtimeutc,
                                time.enddayoffsetutc,
                              )}
                              value={workTimeEnd(time)}
                              onChange={(event) =>
                                updateTime(
                                  index,
                                  timeIndex,
                                  "endtime",
                                  event.target.value,
                                )
                              }
                            />
                          </div>
                          {issue ? (
                            <span className="text-xs font-semibold text-amber-700 dark:text-amber-300">
                              {text(issue)}
                            </span>
                          ) : null}
                        </div>
                      );
                    })}
                    <Button
                      type="button"
                      variant="outline"
                      className="min-h-20 w-full border-dashed"
                      onClick={() => addTimeRange(index)}
                    >
                      <Plus />
                      {text("add")} {text("range")}
                    </Button>
                  </div>
                ) : null}
                {day.isactive && day.fullday ? (
                  <div className="rounded-xl border border-dashed border-border bg-muted/30 px-3 py-2 text-sm font-semibold text-muted-foreground">
                    {text("fullDay")}
                  </div>
                ) : null}
              </div>
            );
          })
        ) : (
          <div className="rounded-2xl border border-dashed border-border p-4 text-sm text-muted-foreground">
            {text("empty")}
          </div>
        )}
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
                setSourceEnvironment(
                  event.target.value === "pro" ? "pro" : "uat",
                );
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
                const id = String(
                  record.guid_fixed ?? record.shopid ?? record._id ?? "",
                );
                return (
                  <option key={`${id}-${index}`} value={id}>
                    {recordTitle(record, config, language)} ({id})
                  </option>
                );
              })}
            </select>
          </label>
          <Button
            type="button"
            variant="outline"
            onClick={() => runCopy("preview")}
            disabled={!selected || saving}
          >
            {saving ? <Loader2 className="animate-spin" /> : <Search />}
            {text("preview")}
          </Button>
          <Button
            type="button"
            onClick={() => runCopy("copy")}
            disabled={!selected || saving}
          >
            {saving ? <Loader2 className="animate-spin" /> : <Copy />}
            {text("copyNow")}
          </Button>
        </div>
        <div className="grid gap-2 rounded-2xl border border-border bg-background p-3 text-sm">
          <b>
            {text("sourceEnvironment")}: {sourceEnvironment.toUpperCase()} → DEV
          </b>
          <b>
            {text("targetShop")}: {targetShopId || "-"}
          </b>
          <p className="text-muted-foreground">
            Preview ก่อน copy ทุกครั้ง เพราะ action นี้มีผลกับข้อมูล MongoDB ของ
            shop ปัจจุบัน
          </p>
        </div>
        {copyPreview ? (
          <pre className="max-h-80 overflow-auto rounded-2xl border border-border bg-muted p-3 text-xs text-muted-foreground">
            {JSON.stringify(copyPreview, null, 2)}
          </pre>
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

function companyRecordForEdit(
  records: SettingRecord[],
  workspace: WorkspaceSession | null,
): SettingRecord | null {
  return records[0] ?? companyRecordFromWorkspace(workspace);
}

function companyRecordFromWorkspace(
  workspace: WorkspaceSession | null,
): SettingRecord | null {
  if (!workspace?.shop?.shopid) return null;
  const shopInfo = isRecord(workspace.shopInfo)
    ? { ...workspace.shopInfo }
    : {};
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

function getMainShopIdFromWorkspace(
  workspace: WorkspaceSession | null,
): string {
  if (!workspace?.shopInfo) return "";
  const mainShopId =
    workspace.shopInfo.main_shop_id ??
    workspace.shopInfo.mainshopid ??
    workspace.shopInfo.mainShopId;
  return typeof mainShopId === "string" ? mainShopId.trim() : "";
}

function requestHeaders(auth: AuthSession): HeadersInit {
  return {
    "Content-Type": "application/json",
    "x-bc-backend-url": auth.backendUrl,
    Authorization: `Bearer ${auth.token}`,
  };
}

function normalizeRecords(
  payload: unknown,
  config: SystemSettingConfig,
): SettingRecord[] {
  if (Array.isArray(payload)) return payload.filter(isRecord);
  if (!isRecord(payload)) return [];
  if (config.kind === "company") {
    const company = extractCompanyRecord(payload);
    return company ? [company] : [];
  }
  if (config.kind === "ai-provider" && Array.isArray(payload.providers))
    return payload.providers.filter(isRecord);
  if (Array.isArray(payload.data))
    return payload.data
      .filter(isRecord)
      .map((record) => normalizeRestaurantRecord(record, config));
  if (isRecord(payload.data)) return [payload.data];
  return [];
}

function extractRecordTotal(payload: unknown, fallback: number): number {
  if (!isRecord(payload)) return fallback;
  const directTotal =
    payload.total ?? payload.count ?? payload.total_count ?? payload.totalCount;
  if (typeof directTotal === "number" && Number.isFinite(directTotal))
    return directTotal;
  const pagination = payload.pagination;
  if (isRecord(pagination)) {
    const pageTotal =
      pagination.total ?? pagination.total_count ?? pagination.totalCount;
    if (typeof pageTotal === "number" && Number.isFinite(pageTotal))
      return pageTotal;
  }
  const meta = payload.meta;
  if (isRecord(meta)) {
    const metaTotal = meta.total ?? meta.total_count ?? meta.totalCount;
    if (typeof metaTotal === "number" && Number.isFinite(metaTotal))
      return metaTotal;
  }
  return fallback;
}

function mergeRecords(
  currentRecords: SettingRecord[],
  nextRecords: SettingRecord[],
  config: SystemSettingConfig,
): SettingRecord[] {
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

function extractCompanyRecord(
  payload: Record<string, unknown>,
): SettingRecord | null {
  const data = payload.data;
  if (isRecord(data)) return data;
  const shop = payload.shop;
  if (isRecord(shop)) return shop;
  const result = payload.result;
  if (isRecord(result)) return result;
  if (payload.shopid || payload.names || payload.settings) return payload;
  return null;
}

function normalizeRestaurantRecord(
  record: SettingRecord,
  config: SystemSettingConfig,
): SettingRecord {
  if (config.kind !== "restaurant-setting") return record;
  const body =
    typeof record.body === "string"
      ? safeJsonParse(record.body, {})
      : isRecord(record.body)
        ? record.body
        : {};
  return isRecord(body) ? { ...record, ...body } : record;
}

function normalizeWorkDays(
  records: SettingRecord[],
  language: LanguageCode,
  scope: DateTimeScope,
): WorkDay[] {
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
        ? record.worktimes
            .filter(isRecord)
            .map((time) => normalizeWorkTimeRecord(time, scope))
        : [],
    };
  });
}

function normalizeWorkTimeRecord(
  record: SettingRecord,
  scope: DateTimeScope,
): WorkDayTime {
  const starttime = normalizeTimeInput(record.starttime ?? record.start);
  const endtime = normalizeTimeInput(record.endtime ?? record.end);
  return applyUtcFields(
    {
      ...record,
      starttime,
      endtime,
      start: toLegacyTime(starttime),
      end: toLegacyTime(endtime),
    },
    scope,
  );
}

function createWorkTime(
  starttime = "",
  endtime = "",
  scope?: DateTimeScope,
): WorkDayTime {
  const start = normalizeTimeInput(starttime);
  const end = normalizeTimeInput(endtime);
  return applyUtcFields(
    {
      starttime: start,
      endtime: end,
      start: toLegacyTime(start),
      end: toLegacyTime(end),
    },
    scope,
  );
}

function withWorkTimePatch(
  time: WorkDayTime,
  key: keyof WorkDayTime,
  value: string,
  scope: DateTimeScope,
): WorkDayTime {
  const normalized =
    key === "starttime" || key === "endtime" || key === "start" || key === "end"
      ? normalizeTimeInput(value)
      : value;
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
  const startUtc = scope?.timezone_offset
    ? localTimeToUtcTime(starttime, scope.timezone_offset)
    : null;
  const endUtc = scope?.timezone_offset
    ? localTimeToUtcTime(endtime, scope.timezone_offset)
    : null;
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

function getWorkTimeIssue(
  day: WorkDay,
  time: WorkDayTime,
  timeIndex: number,
): "timeInvalid" | "timeOverlap" | null {
  if (!day.isactive || day.fullday) return null;
  const start = timeToMinutes(workTimeStart(time));
  const end = timeToMinutes(workTimeEnd(time));
  if (!Number.isFinite(start) || !Number.isFinite(end) || start >= end)
    return "timeInvalid";
  const overlaps = day.worktimes.some((other, otherIndex) => {
    if (otherIndex === timeIndex) return false;
    const otherStart = timeToMinutes(workTimeStart(other));
    const otherEnd = timeToMinutes(workTimeEnd(other));
    if (
      !Number.isFinite(otherStart) ||
      !Number.isFinite(otherEnd) ||
      otherStart >= otherEnd
    )
      return false;
    return start < otherEnd && end > otherStart;
  });
  return overlaps ? "timeOverlap" : null;
}

function formatUtcPreview(
  time: string | undefined,
  dayOffset: number | undefined,
): string {
  if (!time) return "";
  const suffix = dayOffset ? ` ${dayOffset > 0 ? "+" : ""}${dayOffset}d` : "";
  return `UTC+0 ${time}${suffix}`;
}

function buildBranchScopedWorkDayBody(
  record: SettingRecord | undefined,
  workspace: WorkspaceSession,
  workDays: WorkDay[],
): SettingRecord {
  const scope = resolveDateTimeScope(workspace);
  const base = restaurantBody(record);
  const branches = isRecord(base.branches) ? { ...base.branches } : {};
  const scopedWorkDays = workDays.map((day) => ({
    ...day,
    worktimes: day.worktimes.map((time) => applyUtcFields(time, scope)),
  }));
  const branchPayload = {
    ...(isRecord(branches[scope.key])
      ? (branches[scope.key] as SettingRecord)
      : {}),
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

function scopedRestaurantBody(
  record: SettingRecord | undefined,
  scope: DateTimeScope,
): SettingRecord {
  const body = restaurantBody(record);
  const branches = isRecord(body.branches) ? body.branches : {};
  const direct = branches[scope.key];
  if (isRecord(direct)) return direct;
  const matched = Object.values(branches).find((item) => {
    if (!isRecord(item)) return false;
    return (
      String(item.branchguid ?? "") === scope.branchguid ||
      String(item.branchcode ?? "") === scope.branchcode
    );
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

function filterRecordsByDateTimeScope(
  records: SettingRecord[],
  scope: DateTimeScope,
): SettingRecord[] {
  return records.filter((record) => {
    const key = String(record.branch_key ?? "");
    const branchguid = String(record.branchguid ?? "");
    const branchcode = String(record.branchcode ?? "");
    if (!key && !branchguid && !branchcode) return true;
    return (
      key === scope.key ||
      branchguid === scope.branchguid ||
      branchcode === scope.branchcode
    );
  });
}

function defaultForm(
  config: SystemSettingConfig,
  language: LanguageCode,
): FormState {
  const form: FormState = {};
  for (const field of config.fields) {
    if (field.type === "checkbox")
      form[field.key] =
        field.key === "isActive" ||
        field.key === "is_active" ||
        field.key === "isenabled";
    else if (field.type === "names")
      form[field.key] = Object.fromEntries(
        LANGUAGES.map((item) => [item.code, item.code === language ? "" : ""]),
      );
    else if (field.type === "language-configs")
      form[field.key] = defaultLanguageConfigs("th");
    else if (field.type === "language-list") form[field.key] = ["th"];
    else if (field.type === "master-picker") form[field.key] = {};
    else if (field.type === "time-sale-list") form[field.key] = [];
    else if (field.type === "json")
      form[field.key] =
        field.key === "paymentrounding"
          ? defaultPaymentRoundingJson()
          : field.key === "pointconfig"
            ? defaultPointConfigJson()
            : field.key === "permissionCodes" ||
                field.key === "approvalCodes" ||
                field.key === "allowed_tools"
              ? "[]"
              : "{}";
    else if (field.type === "number")
      form[field.key] = isDecimalSettingField(field.key) ? 2 : "";
    else if (field.type === "radio")
      form[field.key] = radioValueToFormValue(
        field.options?.[0]?.value ?? "",
        field,
      );
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

function formFromRecord(
  record: SettingRecord,
  config: SystemSettingConfig,
  language: LanguageCode,
): FormState {
  const form = defaultForm(config, language);
  for (const field of config.fields) {
    const value = recordValueForField(record, config, field);
    if (field.type === "names") form[field.key] = namesToObject(value);
    else if (field.type === "language-configs")
      form[field.key] = normalizeLanguageConfigs(
        value,
        getByPath(record, "settings.language"),
      );
    else if (field.type === "language-list")
      form[field.key] = normalizeLanguageList(
        value,
        getByPath(record, "language"),
      );
    else if (field.type === "master-picker")
      form[field.key] = isRecord(value) ? value : {};
    else if (field.type === "time-sale-list")
      form[field.key] = normalizeTimeSaleFormList(value);
    else if (
      field.type === "json" &&
      config.slug === "permission_definition" &&
      field.key === "branches"
    )
      form[field.key] = isRecord(value) ? value : {};
    else if (
      field.type === "json" &&
      config.slug === "approval_setting" &&
      field.key === "approvals"
    )
      form[field.key] = isRecord(value) ? value : {};
    else if (
      field.type === "json" &&
      config.slug === "permission_link" &&
      (field.key === "permissionCodes" || field.key === "approvalCodes")
    )
      form[field.key] = stringArrayFromForm(value);
    else if (field.type === "json")
      form[field.key] = JSON.stringify(
        value ??
          (field.key === "permissionCodes" ||
          field.key === "approvalCodes" ||
          field.key === "allowed_tools"
            ? []
            : {}),
        null,
        2,
      );
    else if (field.type === "radio")
      form[field.key] = radioValueToFormValue(
        radioFormValue(value, field),
        field,
      );
    else if (field.key === "api_key") form[field.key] = "";
    else if (field.type === "number" && isDecimalSettingField(field.key))
      form[field.key] = normalizeDecimalPlaces(value);
    else if (field.type === "image-gallery") {
      const direct = toUriArray(value);
      if (direct.length > 0) {
        form[field.key] = direct;
      } else {
        const legacySingle = toUriArray(getByPath(record, "imageuri"));
        form[field.key] = legacySingle;
      }
    } else if (field.type === "branch-multi-select") {
      form[field.key] = selectedBranchesFromValue(value);
    } else if (field.type === "company-multi-select") {
      form[field.key] = Array.isArray(value)
        ? value.map((v) =>
            typeof v === "string"
              ? v.trim()
              : isRecord(v)
                ? String(v.guidfixed ?? v.guid_fixed ?? "")
                : "",
          ).filter(Boolean)
        : [];
    } else form[field.key] = value ?? "";
  }
  if (config.slug === "company") {
    const languageConfigs = normalizeLanguageConfigs(
      getByPath(record, "settings.languageconfigs"),
      getByPath(record, "settings.language"),
    );
    form["settings.languageconfigs"] = languageConfigs;
    form["settings.language"] = languageConfigs[0]?.code ?? "th";
  }
  applyCompanyDefaults(form, config);
  applyBranchDefaults(form, config);
  form.guid_fixed = record.guid_fixed || record.guidfixed || record.guid || "";
  return form;
}

function recordValueForField(
  record: SettingRecord,
  config: SystemSettingConfig,
  field: SystemSettingField,
): unknown {
  const value = getByPath(record, field.key);
  if (value !== undefined) return value;
  const aliases = fieldValueAliases[`${config.slug}.${field.key}`] ?? [];
  for (const alias of aliases) {
    const aliasValue = getByPath(record, alias);
    if (aliasValue !== undefined) return aliasValue;
  }
  return undefined;
}

function applyCompanyDefaults(form: FormState, config: SystemSettingConfig) {
  if (config.slug !== "company" && config.slug !== "active_languages") return;
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
  form.language = Array.isArray(form.languages)
    ? (form.languages[0] ?? "th")
    : "th";
  const timezone = stringValue(form.timezone);
  if (timezone) applyTimezoneMetaToForm(form, "timezone", timezone);
}

function applyCountryDefaultsToForm(
  form: FormState,
  fieldKey: string,
  countryCode: string,
) {
  if (countryCode !== "TH") return;
  if (fieldKey.startsWith("contact.")) {
    setFormValueIfEmpty(form, "base_currency", "THB");
    setFormValueIfEmpty(form, "timezone", "Asia/Bangkok");
    setFormValueIfEmpty(form, "date_format", "dd/MM/yyyy");
    setFormValueIfEmpty(form, "year_type", "buddhist");
    setFormValueIfEmpty(form, "decimal_quantity", 2);
    setFormValueIfEmpty(form, "decimal_price", 2);
    setFormValueIfEmpty(form, "decimal_document", 2);
    const timezone = stringValue(form.timezone);
    if (timezone) applyTimezoneMetaToForm(form, "timezone", timezone);
    return;
  }
  for (const [key, value] of Object.entries(companySetupDefaults)) {
    setFormValueIfEmpty(form, key, value);
  }
  const timezone = stringValue(form["settings.timezone"]);
  if (timezone) applyTimezoneMetaToForm(form, "settings.timezone", timezone);
  syncCompanyLanguageForm(form);
}

function syncCompanyLanguageForm(form: FormState) {
  const defaultCode = supportedLanguageCode(form["settings.language"], "th");
  const configs = normalizeLanguageConfigs(
    form["settings.languageconfigs"],
    defaultCode,
  );
  form["settings.language"] = configs[0]?.code ?? defaultCode;
  form["settings.languageconfigs"] = configs;
}

function setFormValueIfEmpty(form: FormState, key: string, value: unknown) {
  if (isEmptyFormValue(form[key])) form[key] = value;
}

function isEmptyFormValue(value: unknown): boolean {
  return value === "" || value === null || value === undefined;
}

function applyTimezoneMetaToForm(
  form: FormState,
  key: string,
  timeZone: string,
) {
  const meta = timezoneMeta(timeZone);
  const prefix = key.includes(".")
    ? `${key.split(".").slice(0, -1).join(".")}.`
    : "";
  setFormValueIfEmpty(form, `${prefix}timezone_label`, meta.label);
  setFormValueIfEmpty(form, `${prefix}timezone_offset`, meta.offset);
}

function buildPayload(
  form: FormState,
  editing: SettingRecord | null,
  config: SystemSettingConfig,
  workspace: WorkspaceSession,
  auth: AuthSession,
  language: LanguageCode,
): SettingRecord {
  const payload: SettingRecord = { ...(editing ?? {}) };
  for (const field of config.fields) {
    if (field.readOnly) continue;
    const value = form[field.key];
    if (field.type === "names") {
      setByPath(
        payload,
        field.key,
        objectToNames(
          value,
          language,
          getByPath(payload, field.key),
          nameEditorLanguageCodes(form, config, language, workspace),
        ),
      );
    } else if (field.type === "language-configs")
      setByPath(
        payload,
        field.key,
        normalizeLanguageConfigs(value, form["settings.language"], {
          forcePrimaryFirst: true,
        }),
      );
    else if (field.type === "language-list")
      setByPath(
        payload,
        field.key,
        normalizeLanguageList(value, form.language, {
          forcePrimaryFirst: true,
        }),
      );
    else if (field.type === "master-picker")
      setByPath(payload, field.key, isRecord(value) ? value : {});
    else if (field.type === "image-gallery")
      setByPath(payload, field.key, toUriArray(value));
    else if (field.type === "time-sale-list")
      setByPath(payload, field.key, normalizeTimeSalePayload(value));
    else if (field.type === "branch-multi-select")
      setByPath(payload, field.key, selectedBranchesFromValue(value));
    else if (field.type === "company-multi-select")
      setByPath(
        payload,
        field.key,
        Array.isArray(value)
          ? value.map((v) =>
              typeof v === "string"
                ? v.trim()
                : isRecord(v)
                  ? String(v.guidfixed ?? v.guid_fixed ?? "")
                  : "",
            ).filter(Boolean)
          : [],
      );
    else if (field.type === "json")
      setByPath(payload, field.key, parseJsonField(value, field.key));
    else if (field.type === "checkbox")
      setByPath(payload, field.key, Boolean(value));
    else if (field.type === "radio")
      setByPath(
        payload,
        field.key,
        radioValueToFormValue(radioFormValue(value, field), field),
      );
    else if (field.type === "select")
      setByPath(
        payload,
        field.key,
        optionValueToFormValue(String(value ?? ""), field),
      );
    else if (field.type === "combo") {
      const selectedValue = String(value ?? "").trim();
      setByPath(payload, field.key, selectedValue);
      if (field.optionSource === "timezones" && selectedValue)
        setTimezoneDerivedPayload(payload, field.key, selectedValue);
    } else if (field.type === "number")
      setByPath(
        payload,
        field.key,
        isDecimalSettingField(field.key)
          ? normalizeDecimalPlaces(value)
          : value === "" || value === null || value === undefined
            ? 0
            : Number(value),
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
    if (!editing && config.slug === "mcp_apikey")
      payload.createWithExport = true;
  }

  if (config.kind === "ai-provider") {
    payload.shop_id = workspace.shop.shopid;
    if (!payload.api_key && String(payload.provider_name) !== "ollama") {
      throw new Error(
        language === "th"
          ? "กรุณากรอก API Key เมื่อบันทึก AI Provider"
          : "Please enter API Key when saving AI Provider.",
      );
    }
  }

  if (config.kind === "company") {
    const configs = normalizeLanguageConfigs(
      getByPath(payload, "settings.languageconfigs"),
      form["settings.language"] ?? getByPath(payload, "settings.language"),
      { forcePrimaryFirst: true },
    );
    setByPath(payload, "settings.languageconfigs", configs);
    setByPath(payload, "settings.language", configs[0]?.code ?? "th");
    payload.shopid = workspace.shop.shopid;
  }

  if (config.slug === "branch") {
    try {
      setByPath(
        payload,
        "code",
        normalizeThaiTaxBranchCode(getByPath(payload, "code")),
      );
    } catch {
      throw new Error(
        language === "th"
          ? "รหัสสาขาภาษีไทยต้องเป็นเลขไม่เกิน 5 หลัก เช่น สำนักงานใหญ่ = 00000 และสาขาที่ 1 = 00001"
          : "Thai tax branch code must be numeric and no more than 5 digits. Head office = 00000; branch 1 = 00001.",
      );
    }
    const languages = normalizeLanguageList(
      getByPath(payload, "languages"),
      form.language ?? getByPath(payload, "language"),
      { forcePrimaryFirst: true },
    );
    setByPath(payload, "languages", languages);
    setByPath(payload, "language", languages[0] ?? "th");
    const timezone = stringValue(getByPath(payload, "timezone"));
    if (timezone) {
      const meta = timezoneMeta(timezone);
      setByPath(
        payload,
        "timezonelabel",
        stringValue(getByPath(payload, "timezonelabel")) || meta.label,
      );
      setByPath(
        payload,
        "timezoneoffset",
        normalizeUtcOffset(
          stringValue(getByPath(payload, "timezoneoffset")) || meta.offset,
        ),
      );
      setByPath(
        payload,
        "timezone_label",
        stringValue(getByPath(payload, "timezone_label")) || meta.label,
      );
      setByPath(
        payload,
        "timezone_offset",
        normalizeUtcOffset(
          stringValue(getByPath(payload, "timezone_offset")) || meta.offset,
        ),
      );
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
    Object.assign(
      payload,
      dateTimeScopePayload(resolveDateTimeScope(workspace)),
    );
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

function resolveDateTimeScope(
  workspace: WorkspaceSession | null,
): DateTimeScope {
  const branch = workspace?.branch ?? null;
  const shopInfo = workspace?.shopInfo ?? null;
  const branchcode = stringValue(branch?.code ?? workspace?.shop.branchcode);
  const branchguid = stringValue(branch?.guid_fixed);
  const shopTimezone = stringValue(getByPath(shopInfo, "settings.timezone"));
  const shopTimezoneLabel = stringValue(
    getByPath(shopInfo, "settings.timezone_label"),
  );
  const shopTimezoneOffset = stringValue(
    getByPath(shopInfo, "settings.timezone_offset"),
  );
  const branchYear = stringValue(branch?.year_type).toLowerCase();
  const useBuddhistCalendar = booleanLikeValue(
    getByPath(shopInfo, "settings.usebuddhistcalendar"),
  );
  const calendarYearType: CalendarYearType =
    branchYear === "buddhist" || branchYear === "be" || branchYear === "พ.ศ."
      ? "buddhist"
      : branchYear === "christian" ||
          branchYear === "ce" ||
          branchYear === "ค.ศ."
        ? "christian"
        : useBuddhistCalendar
          ? "buddhist"
          : "christian";
  return {
    key: branchguid || branchcode || "company",
    branchcode,
    branchguid,
    timezone: stringValue(branch?.timezone) || shopTimezone,
    timezone_label:
      stringValue(branch?.timezone_label) ||
      shopTimezoneLabel ||
      stringValue(branch?.timezone) ||
      shopTimezone,
    timezone_offset: normalizeUtcOffset(
      stringValue(branch?.timezone_offset) || shopTimezoneOffset,
    ),
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
  return new Date(
    Date.UTC(year, month - 1, day + utc.dayOffset, hours, minutes, 0, 0),
  ).toISOString();
}

function normalizeUtcOffset(value: string): string {
  const raw = value.trim().replace(/^UTC/i, "");
  const match = raw.match(/^([+-])(\d{1,2}):?(\d{2})$/);
  if (!match) return raw;
  return `${match[1]}${match[2].padStart(2, "0")}:${match[3]}`;
}

function stringValue(value: unknown): string {
  return typeof value === "string"
    ? value.trim()
    : value === null || value === undefined
      ? ""
      : String(value).trim();
}

function productCategoryGuid(record: SettingRecord | null | undefined): string {
  if (!record) return "";
  return stringValue(record.guid_fixed ?? record.guidfixed ?? record.guid);
}

function productCategoryParentGuid(
  record: SettingRecord | null | undefined,
): string {
  if (!record) return "";
  return stringValue(record.parent_guid ?? record.parentguid);
}

function productCategoryGroupNumber(
  record: SettingRecord | null | undefined,
): number {
  if (!record) return 0;
  const value = Number(record.group_number ?? record.groupnumber ?? 0);
  return Number.isFinite(value) ? value : 0;
}

function productCategoryXOrder(record: SettingRecord): number {
  const xsorts = Array.isArray(record.xsorts) ? record.xsorts : [];
  const first = xsorts.find(isRecord);
  const value = isRecord(first) ? Number(first.xorder ?? 0) : 0;
  return Number.isFinite(value) ? value : 0;
}

function isEmailLike(value: unknown): boolean {
  return /^[^\s@]+@[^\s@]+\.[^\s@]+$/.test(stringValue(value));
}

function recordId(
  record: SettingRecord | null | undefined,
  config: SystemSettingConfig,
): string {
  if (!record) return "";
  const value = config.idField ? getByPath(record, config.idField) : undefined;
  return String(
    value ??
      record.guid_fixed ??
      record.guidfixed ??
      record.id ??
      record._id ??
      record.code ??
      record.shopid ??
      record.provider_name ??
      "",
  );
}

function recordTitle(
  record: SettingRecord,
  config: SystemSettingConfig | undefined,
  language: LanguageCode,
): string {
  if (config?.slug === "active_languages")
    return systemSettingLabel(config, language);
  const names =
    getByPath(record, "names") ??
    getByPath(record, "name") ??
    getByPath(record, "desc");
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
      record.guidfixed ??
      config?.slug ??
      "",
  );
}

function recordBranchCaption(record: SettingRecord): string {
  return (
    stringValue(record.branchcode) ||
    stringValue(record.branchguid) ||
    stringValue(record.branch_key)
  );
}



function userAccessDisabled(record: SettingRecord): boolean {
  return Boolean(
    record.is_access_disabled ?? record.is_access_disabled ?? false,
  );
}

function localizedValue(value: unknown, language: LanguageCode): string {
  if (Array.isArray(value)) {
    const found = value.find(
      (item) =>
        isRecord(item) &&
        String(item.code).toLowerCase() === language &&
        typeof item.name === "string",
    );
    if (isRecord(found) && typeof found.name === "string") return found.name;
    const fallback = value.find(
      (item) => isRecord(item) && typeof item.name === "string",
    );
    return isRecord(fallback) && typeof fallback.name === "string"
      ? fallback.name
      : "";
  }
  if (typeof value === "string") return value;
  return "";
}

function uiBackendKey(key: keyof typeof uiEn): string {
  return uiBackendKeys[key] ?? key;
}

function systemSettingTitle(
  config: SystemSettingConfig,
  language: LanguageCode,
  dictionary: BackendLanguageDictionary,
): string {
  const fallback = systemSettingLabel(config, language);
  const backendKey = systemSettingBackendKeys[config.slug] ?? config.slug;
  const value = backendText(dictionary, backendKey, fallback);
  return value === backendKey || value === config.slug ? fallback : value;
}

function fieldLabel(
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

function optionLabel(
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

function isProductUnitOption(value: unknown): value is ProductUnitOption {
  return (
    isRecord(value) &&
    typeof value.unitcode === "string" &&
    value.unitcode.trim().length > 0
  );
}

function unitDisplayName(
  unit: ProductUnitOption,
  language: LanguageCode,
): string {
  const names = Array.isArray(unit.names) ? unit.names : [];
  return (
    names
      .find(
        (name) => name.code?.toLowerCase() === language && name.name?.trim(),
      )
      ?.name?.trim() ??
    names
      .find((name) => name.code?.toLowerCase() === "th" && name.name?.trim())
      ?.name?.trim() ??
    names.find((name) => name.name?.trim())?.name?.trim() ??
    unit.unitcode
  );
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
      onApply(
        new File([blob], "cropped-image.webp", {
          type: "image/webp",
          lastModified: Date.now(),
        }),
      );
    } catch {
      setError(uploadUiText(language, "imageUploadFailed"));
    }
  }

  return (
    <div
      className="fixed inset-0 z-50 grid place-items-center bg-black/40 p-2"
      role="dialog"
      aria-modal="true"
    >
      <section className="grid w-[min(420px,calc(100vw-16px))] gap-2 rounded-2xl border border-border bg-card p-3 text-card-foreground shadow-xl">
        <div className="flex items-center justify-between gap-2">
          <div className="text-sm font-semibold">
            {uploadUiText(language, "cropTitle")}
          </div>
          <Button
            type="button"
            variant="ghost"
            size="sm"
            onClick={onCancel}
            aria-label={uploadUiText(language, "cancel")}
          >
            <X />
          </Button>
        </div>
        <div className="grid place-items-center rounded-2xl border border-input bg-muted/40 p-2">
          <canvas
            ref={canvasRef}
            width={512}
            height={512}
            className="aspect-square w-full max-w-80 rounded-xl bg-background object-contain"
          />
        </div>
        <label className="grid gap-1 text-xs font-semibold">
          <span>Zoom</span>
          <input
            type="range"
            min="1"
            max="3"
            step="0.01"
            value={zoom}
            onChange={(event) => setZoom(Number(event.target.value))}
          />
        </label>
        <label className="grid gap-1 text-xs font-semibold">
          <span>X</span>
          <input
            type="range"
            min="-100"
            max="100"
            step="1"
            value={offsetX}
            onChange={(event) => setOffsetX(Number(event.target.value))}
          />
        </label>
        <label className="grid gap-1 text-xs font-semibold">
          <span>Y</span>
          <input
            type="range"
            min="-100"
            max="100"
            step="1"
            value={offsetY}
            onChange={(event) => setOffsetY(Number(event.target.value))}
          />
        </label>
        {error ? (
          <p className="text-xs font-semibold text-destructive">{error}</p>
        ) : null}
        <div className="flex flex-wrap justify-end gap-2">
          <Button type="button" variant="outline" onClick={onCancel}>
            {uploadUiText(language, "cancel")}
          </Button>
          <Button type="button" onClick={() => void apply()} disabled={!image}>
            {uploadUiText(language, "applyCrop")}
          </Button>
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
  const sourceX = clampNumber(
    (image.naturalWidth - cropSize) / 2 + (crop.offsetX / 100) * (maxX / 2),
    0,
    maxX,
  );
  const sourceY = clampNumber(
    (image.naturalHeight - cropSize) / 2 + (crop.offsetY / 100) * (maxY / 2),
    0,
    maxY,
  );
  const context = canvas.getContext("2d");
  if (!context) return;
  context.clearRect(0, 0, canvas.width, canvas.height);
  context.imageSmoothingEnabled = true;
  context.imageSmoothingQuality = "high";
  context.drawImage(
    image,
    sourceX,
    sourceY,
    cropSize,
    cropSize,
    0,
    0,
    canvas.width,
    canvas.height,
  );
}

function clampNumber(value: number, min: number, max: number): number {
  return Math.min(max, Math.max(min, value));
}

function extractUploadUri(payload: unknown): string {
  if (!isRecord(payload)) return "";
  const direct = stringValue(
    payload.uri ??
      payload.url ??
      payload.file_url ??
      payload.fileUrl ??
      payload.imageurl ??
      payload.image_uri,
  );
  if (direct) return direct;
  const data = payload.data;
  if (!isRecord(data)) return "";
  return stringValue(
    data.uri ??
      data.url ??
      data.file_url ??
      data.fileUrl ??
      data.imageurl ??
      data.image_uri,
  );
}

function useAuthenticatedImageDisplaySource(
  value: unknown,
  auth: AuthSession | null,
): {
  displayUrl: string;
  failed: boolean;
  loading: boolean;
  requestedUrl: string;
} {
  const authBackendUrl = auth?.backendUrl ?? "";
  const authToken = auth?.token ?? "";
  const authUsername = auth?.username ?? "";
  const requestedUrl = useMemo(
    () => imageDisplayUrl(value, authBackendUrl),
    [authBackendUrl, value],
  );
  const [state, setState] = useState({
    displayUrl: "",
    failed: false,
    loading: false,
  });

  useEffect(() => {
    if (!requestedUrl) {
      setState({ displayUrl: "", failed: false, loading: false });
      return;
    }

    if (!imageNeedsAuthenticatedFetch(requestedUrl)) {
      setState({ displayUrl: requestedUrl, failed: false, loading: false });
      return;
    }

    if (!authToken) {
      clearAuthenticatedImageObjectUrlCache();
      setState({ displayUrl: "", failed: true, loading: false });
      return;
    }

    syncAuthenticatedImageCacheOwner(authBackendUrl, authUsername, authToken);
    const cacheKey = authenticatedImageCacheKey(
      requestedUrl,
      authBackendUrl,
      authUsername,
      authToken,
    );
    const cachedObjectUrl = getCachedAuthenticatedImageObjectUrl(cacheKey);
    if (cachedObjectUrl) {
      setState({ displayUrl: cachedObjectUrl, failed: false, loading: false });
      return;
    }

    const controller = new AbortController();
    let cancelled = false;

    setState({ displayUrl: "", failed: false, loading: true });
    void fetch(requestedUrl, {
      cache: "no-store",
      headers: { Authorization: `Bearer ${authToken}` },
      signal: controller.signal,
    })
      .then(async (response) => {
        if (!response.ok) throw new Error(`HTTP ${response.status}`);
        return response.blob();
      })
      .then((blob) => {
        const nextObjectUrl = URL.createObjectURL(blob);
        if (cancelled) {
          URL.revokeObjectURL(nextObjectUrl);
          return;
        }
        cacheAuthenticatedImageObjectUrl(cacheKey, nextObjectUrl);
        setState({
          displayUrl: nextObjectUrl,
          failed: false,
          loading: false,
        });
      })
      .catch((error: unknown) => {
        if (cancelled) return;
        if (error instanceof DOMException && error.name === "AbortError")
          return;
        setState({ displayUrl: "", failed: true, loading: false });
      });

    return () => {
      cancelled = true;
      controller.abort();
    };
  }, [authBackendUrl, authToken, authUsername, requestedUrl]);

  return { ...state, requestedUrl };
}

function authenticatedImageCacheKey(
  imageUrl: string,
  backendUrl: string,
  username: string,
  token: string,
): string {
  return [
    backendUrl.trim(),
    username.trim().toLowerCase(),
    token,
    imageUrl,
  ].join("\u0000");
}

function syncAuthenticatedImageCacheOwner(
  backendUrl: string,
  username: string,
  token: string,
) {
  const owner = [backendUrl.trim(), username.trim().toLowerCase(), token].join(
    "\u0000",
  );
  if (authenticatedImageCacheOwner && authenticatedImageCacheOwner !== owner)
    clearAuthenticatedImageObjectUrlCache();
  authenticatedImageCacheOwner = owner;
}

function getCachedAuthenticatedImageObjectUrl(cacheKey: string): string {
  const cached = authenticatedImageObjectUrlCache.get(cacheKey);
  if (!cached) return "";
  cached.lastAccessAt = Date.now();
  return cached.objectUrl;
}

function cacheAuthenticatedImageObjectUrl(cacheKey: string, objectUrl: string) {
  const existing = authenticatedImageObjectUrlCache.get(cacheKey);
  if (existing?.objectUrl && existing.objectUrl !== objectUrl)
    URL.revokeObjectURL(existing.objectUrl);

  authenticatedImageObjectUrlCache.set(cacheKey, {
    objectUrl,
    lastAccessAt: Date.now(),
  });

  if (
    authenticatedImageObjectUrlCache.size <=
    AUTHENTICATED_IMAGE_CACHE_MAX_ENTRIES
  )
    return;

  const entries = [...authenticatedImageObjectUrlCache.entries()].sort(
    (first, second) => first[1].lastAccessAt - second[1].lastAccessAt,
  );
  for (const [key, entry] of entries.slice(
    0,
    authenticatedImageObjectUrlCache.size -
      AUTHENTICATED_IMAGE_CACHE_MAX_ENTRIES,
  )) {
    URL.revokeObjectURL(entry.objectUrl);
    authenticatedImageObjectUrlCache.delete(key);
  }
}

function clearAuthenticatedImageObjectUrlCache() {
  for (const entry of authenticatedImageObjectUrlCache.values()) {
    URL.revokeObjectURL(entry.objectUrl);
  }
  authenticatedImageObjectUrlCache.clear();
  authenticatedImageCacheOwner = "";
}

function imageDisplayUrl(value: unknown, backendUrl: unknown): string {
  const raw = stringValue(value).trim();
  if (!raw) return "";
  if (/^(blob:|data:|https?:\/\/)/i.test(raw)) return raw;
  if (raw.startsWith("//"))
    return typeof window === "undefined"
      ? raw
      : `${window.location.protocol}${raw}`;
  if (raw.startsWith("/api/")) return raw;

  const base = mainApiDisplayBase(backendUrl);
  if (!base) return raw;
  if (raw.startsWith("/")) return `${base}${raw}`;
  if (raw.toLowerCase().startsWith("images/"))
    return `${base}/${raw.replace(/^\/+/, "")}`;
  return `${base}/images/${raw.replace(/^\/+/, "")}`;
}

function mainApiDisplayBase(rawBackendUrl: unknown): string {
  const raw = stringValue(rawBackendUrl).trim();
  if (!raw) return "";
  try {
    const withProtocol = /^https?:\/\//i.test(raw) ? raw : `http://${raw}`;
    const parsed = new URL(withProtocol);
    const path = parsed.pathname.replace(/\/+$/, "");
    parsed.pathname = path.toLowerCase().endsWith("/goapi")
      ? path.slice(0, -"/goapi".length) || "/"
      : "/";
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
    return new File([blob], `${baseName}.webp`, {
      type: "image/webp",
      lastModified: Date.now(),
    });
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

function canvasToBlob(
  canvas: HTMLCanvasElement,
  type: string,
  quality: number,
): Promise<Blob | null> {
  return new Promise((resolve) => canvas.toBlob(resolve, type, quality));
}

function radioFormValue(value: unknown, field: SystemSettingField): string {
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

function radioValueToFormValue(
  value: string,
  field: SystemSettingField,
): string | boolean | number {
  if (field.valueType === "boolean") return booleanLikeValue(value);
  if (field.valueType === "number") return Number(value);
  return value;
}

function optionValueToFormValue(
  value: string,
  field: SystemSettingField,
): string | number {
  if (field.valueType === "number") {
    const parsed = Number(value);
    return Number.isFinite(parsed) ? parsed : 0;
  }
  return value;
}

function booleanLikeValue(value: unknown): boolean {
  if (typeof value === "boolean") return value;
  const normalized = stringValue(value).toLowerCase();
  return (
    normalized === "true" ||
    normalized === "1" ||
    normalized === "yes" ||
    normalized === "y" ||
    normalized === "buddhist" ||
    normalized === "be" ||
    normalized === "พ.ศ."
  );
}

function isDecimalSettingField(key: string): boolean {
  return (
    key === "decimal_quantity" ||
    key === "decimal_price" ||
    key === "decimal_document" ||
    key === "settings.decimal_quantity" ||
    key === "settings.decimal_price" ||
    key === "settings.decimal_document"
  );
}

function normalizeDecimalPlaces(value: unknown): number {
  const numeric = Number(value);
  if (!Number.isFinite(numeric) || numeric < 2) return 2;
  return Math.min(8, Math.trunc(numeric));
}

function comboOptionsForField(
  field: SystemSettingField,
  language: LanguageCode,
  currentValue: string,
): ComboOption[] {
  const sourceOptions =
    field.optionSource === "timezones"
      ? timezoneOptions(language)
      : (field.options ?? []);
  if (
    !currentValue ||
    sourceOptions.some((option) => option.value === currentValue)
  )
    return sourceOptions;
  return [{ value: currentValue, label: currentValue }, ...sourceOptions];
}

function timezoneOptions(language: LanguageCode): ComboOption[] {
  return supportedTimeZones()
    .map((timeZone) => {
      const meta = timezoneMeta(timeZone);
      return {
        value: timeZone,
        label: `${meta.offset ? `UTC${meta.offset} ` : ""}${timeZone}`,
      };
    })
    .sort((first, second) =>
      first.label.localeCompare(second.label, localeOf(language)),
    );
}

function supportedTimeZones(): string[] {
  const intl = Intl as typeof Intl & {
    supportedValuesOf?: (key: "timeZone") => string[];
  };
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

function setTimezoneDerivedPayload(
  payload: SettingRecord,
  key: string,
  timeZone: string,
) {
  const meta = timezoneMeta(timeZone);
  const prefix = key.includes(".")
    ? `${key.split(".").slice(0, -1).join(".")}.`
    : "";
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
    const zone =
      parts.find((part) => part.type === "timeZoneName")?.value ?? "";
    const rawOffset = zone.replace(/^GMT/i, "").trim();
    return rawOffset ? normalizeUtcOffset(rawOffset) : "+00:00";
  } catch {
    return "";
  }
}

function namesToObject(value: unknown): Record<string, string> {
  if (isRecord(value))
    return Object.fromEntries(
      Object.entries(value).map(([key, item]) => [key, String(item ?? "")]),
    );
  if (!Array.isArray(value)) return {};
  return Object.fromEntries(
    value
      .filter(isRecord)
      .map((item) => [String(item.code ?? ""), String(item.name ?? "")]),
  );
}

function objectToNames(
  value: unknown,
  language: LanguageCode,
  previousValue?: unknown,
  activeLanguages?: string[],
): { code: string; name: string; isauto: boolean; isdelete: boolean }[] {
  const record = namesToObject(value);
  const previous = namesToObject(previousValue);
  const active = new Set(
    (activeLanguages?.length
      ? activeLanguages
      : LANGUAGES.map((item) => item.code)
    ).map((code) => code.toLowerCase()),
  );
  return LANGUAGES.map((item) => ({
    code: item.code,
    name: active.has(item.code)
      ? record[item.code]?.trim() ||
        (item.code === "en" ? record[language]?.trim() : "") ||
        ""
      : previous[item.code]?.trim() || record[item.code]?.trim() || "",
    isauto: false,
    isdelete: false,
  }));
}

function parseJsonField(value: unknown, key: string): unknown {
  if (typeof value !== "string") return value;
  const trimmed = value.trim();
  if (!trimmed)
    return key === "permissionCodes" ||
      key === "approvalCodes" ||
      key === "allowed_tools"
      ? []
      : {};
  return JSON.parse(trimmed);
}

function getByPath(record: unknown, path: string): unknown {
  if (!isRecord(record)) return undefined;
  return path
    .split(".")
    .reduce<unknown>(
      (current, key) => (isRecord(current) ? current[key] : undefined),
      record,
    );
}

function getPathOrFlatValue(record: SettingRecord, path: string): unknown {
  const nested = getByPath(record, path);
  return nested ?? record[path];
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

function fieldDisplayValue(
  field: SystemSettingField,
  value: unknown,
  language: LanguageCode,
): string {
  if (field.type === "master-picker") {
    return masterPickerDisplayValue(value, language);
  }
  if (field.type === "time-sale-list") {
    const count = normalizeTimeSaleFormList(value).length;
    return count > 0
      ? language === "th"
        ? `${count} ช่วงเวลา`
        : `${count} time windows`
      : "-";
  }
  if (
    field.type === "combo" ||
    field.type === "radio" ||
    field.type === "select"
  ) {
    const optionValue =
      field.type === "radio"
        ? radioFormValue(value, field)
        : String(value ?? "");
    const option = field.options?.find((item) => item.value === optionValue);
    if (option) return optionLabel(option, language);
  }
  return shortValue(value, language);
}

function masterPickerDisplayValue(
  value: unknown,
  language: LanguageCode,
): string {
  if (!isRecord(value)) return "-";
  const names = getLocalizedNameArray(value.names);
  const name =
    localizedNameForLanguage(names, language) || stringValue(value.name);
  const code = stringValue(value.code);
  if (name && code) return `${name} (${code})`;
  return name || code || "-";
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
  return typeof payload.message === "string"
    ? payload.message
    : typeof payload.error === "string"
      ? payload.error
      : undefined;
}

function safeJsonParse(value: string, fallback: unknown): unknown {
  try {
    return JSON.parse(value);
  } catch {
    return fallback;
  }
}

function isActiveRecord(record: SettingRecord): boolean {
  if ("is_access_disabled" in record || "is_access_disabled" in record)
    return !userAccessDisabled(record);
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
