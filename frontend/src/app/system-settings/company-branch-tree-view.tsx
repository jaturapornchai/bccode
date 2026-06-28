"use client";

import React, { useCallback, useMemo, useState, useEffect } from "react";
import {
  Building2,
  GitBranch,
  Plus,
  ChevronRight,
  ChevronDown,
  Edit3,
  Trash2,
  Save,
  Loader2,
  Check,
  ArrowRight,
  KeyRound,
  UploadCloud,
  ImageIcon,
  X,
} from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { LogoAvatar } from "@/components/logo-avatar";
import { type LanguageCode, LANGUAGES } from "@/lib/i18n";
import { DEFAULT_TIME_ZONE, timezoneMeta, timezoneSelectOptions } from "@/lib/date-time";
import { cn } from "@/lib/utils";
import { normalizeLanguageConfigs } from "./system-settings-screen";
import { deriveMainApiUrl } from "@/lib/backend-url";
import { NamesEditor } from "@/components/product-barcode/names-editor";
import { AddressesEditor } from "@/components/product-barcode/addresses-editor";
import { isThaiHeadOfficeBranchCode, normalizeThaiTaxBranchCode } from "@/lib/thai-branch-code";
import { notifyWorkspaceChanged } from "@/lib/workspace-models";
import { normalizeBusinessCode } from "@/lib/business-code";

interface CompanyBranchTreeViewProps {
  auth: { token: string; backendUrl: string } | null;
  workspace: CompanyWorkspace | null;
  language: LanguageCode;
  onRefresh?: () => void;
}

interface CompanyWorkspace {
  shop: { holdingcode: string };
  shopInfo?: {
    settings?: {
      language?: string;
      languageconfigs?: unknown;
    };
  } | null;
}

interface LocalizedNameEntry {
  code?: string;
  name?: string;
  isauto?: boolean;
  isdelete?: boolean;
}

type LocalizedNames = LocalizedNameEntry[] | Record<string, unknown> | null | undefined;

interface CompanyRecord {
  guidfixed?: string;
  code?: string;
  logouri?: string;
  names?: LocalizedNames;
  taxid?: string;
  isactive?: boolean;
  deletedat?: string | null;
}

interface BranchRecord {
  guidfixed?: string;
  companyguid?: string;
  code?: string;
  logouri?: string;
  names?: LocalizedNames;
  timezone?: string;
  timezonelabel?: string;
  timezoneoffset?: string;
  language?: string;
  dateformat?: string;
  yeartype?: string;
  basecurrency?: string;
  branchtype?: string;
  isvatregistered?: boolean;
  companyregistrationno?: string;
  email?: string;
  managername?: string;
  fiscalstartmonth?: number;
  documentprefixes?: { doctype?: string; prefix?: string }[];
  addresses?: { code?: string; address?: string }[];
  etaxenabled?: boolean;
  isactive?: boolean;
  deletedat?: string | null;
}

const MONTH_OPTIONS = [
  { value: 1, label: "มกราคม" },
  { value: 2, label: "กุมภาพันธ์" },
  { value: 3, label: "มีนาคม" },
  { value: 4, label: "เมษายน" },
  { value: 5, label: "พฤษภาคม" },
  { value: 6, label: "มิถุนายน" },
  { value: 7, label: "กรกฎาคม" },
  { value: 8, label: "สิงหาคม" },
  { value: 9, label: "กันยายน" },
  { value: 10, label: "ตุลาคม" },
  { value: 11, label: "พฤศจิกายน" },
  { value: 12, label: "ธันวาคม" },
];

// ภ.พ.20 branch registration type.
const BRANCH_TYPE_OPTIONS = [
  { value: "head", label: "สำนักงานใหญ่" },
  { value: "permanent", label: "สาขาถาวร" },
  { value: "temporary", label: "สาขาชั่วคราว" },
];

const YEAR_TYPE_OPTIONS = [
  { value: "buddhist", label: "พ.ศ. (Buddhist Era)" },
  { value: "christian", label: "ค.ศ. (Christian Era)" },
];

// ตัวอย่างปีในรูปแบบวันที่เปลี่ยนตามประเภทปี: พ.ศ. -> 2568, ค.ศ. -> 2025
function dateFormatOptionsFor(yearType: string) {
  const y = yearType === "buddhist" ? 2568 : 2025;
  return [
    { value: "dd/MM/yyyy", label: `dd/MM/yyyy (31/12/${y})` },
    { value: "dd-MM-yyyy", label: `dd-MM-yyyy (31-12-${y})` },
    { value: "yyyy-MM-dd", label: `yyyy-MM-dd (${y}-12-31)` },
    { value: "MM/dd/yyyy", label: `MM/dd/yyyy (12/31/${y})` },
  ];
}

// Common currencies for Thai/ASEAN businesses (default THB).
const CURRENCY_OPTIONS = [
  { value: "THB", label: "THB — บาท" },
  { value: "USD", label: "USD — US Dollar" },
  { value: "EUR", label: "EUR — Euro" },
  { value: "GBP", label: "GBP — Pound" },
  { value: "JPY", label: "JPY — Yen" },
  { value: "CNY", label: "CNY — Renminbi" },
  { value: "LAK", label: "LAK — Lao Kip" },
  { value: "MMK", label: "MMK — Myanmar Kyat" },
  { value: "KHR", label: "KHR — Cambodian Riel" },
  { value: "VND", label: "VND — Vietnamese Dong" },
  { value: "SGD", label: "SGD — Singapore Dollar" },
  { value: "MYR", label: "MYR — Malaysian Ringgit" },
];

// ประเภทเอกสารหลักที่สาขากำหนดคำนำหน้าเลขที่เอกสารแยกได้ (subset จาก ~46 types ในระบบ).
// code ตรงกับ MODULE_NAME ของ transaction module ใน backend (เก็บค่าอย่างเดียว — generator ยังไม่ใช้).
const DOC_PREFIX_TYPES = [
  { code: "SI", label: "ใบกำกับภาษี / ใบเสร็จ" },
  { code: "ST", label: "ใบลดหนี้ (ขาย)" },
  { code: "SA", label: "ใบเพิ่มหนี้ (ขาย)" },
  { code: "SO", label: "ใบสั่งขาย" },
  { code: "QT", label: "ใบเสนอราคา" },
  { code: "PU", label: "ใบรับสินค้า (ซื้อ)" },
  { code: "PO", label: "ใบสั่งซื้อ" },
  { code: "PT", label: "ใบรับคืน (ซื้อ)" },
  { code: "TF", label: "ใบโอนสินค้าระหว่างสาขา" },
  { code: "AJ", label: "ใบปรับปรุงสต็อก" },
  { code: "EE", label: "ใบสำคัญรับเงิน" },
  { code: "DE", label: "ใบสำคัญจ่าย" },
  { code: "PC", label: "เงินสดย่อย / มัดจำ" },
];

// คำนำหน้าเลขที่เอกสาร = 2 ตัวอักษร (uppercase, ไม่มีช่องว่าง) ต่อด้วย YYMMDD + running เช่น PO -> PO26062800001.
const normalizeDocPrefix = (v: string) => v.toUpperCase().replace(/[^A-Z0-9]/g, "").slice(0, 2);
// คืนค่า prefix ที่ผู้ใช้ตั้งไว้ (1-2 ตัวอักษร/ตัวเลขล้วน) หรือ "" ถ้าเป็น legacy/ยาว (เช่น "HQ-SI") เพื่อให้ default = รหัสประเภท
const cleanDocPrefixOrEmpty = (v: string) => (/^[A-Za-z0-9]{1,2}$/.test(v.trim()) ? v.trim().toUpperCase() : "");

type NodeType = "company" | "branch";
type ConfirmAction = "save" | "delete";
type OrganizationFormType = "viewcompany" | "viewbranch" | "editcompany" | "editbranch" | "createcompany" | "createbranch";

interface SelectedNode {
  type: NodeType;
  guidfixed?: string;
  companyguid?: string;
  data: Partial<CompanyRecord> | Partial<BranchRecord>;
}

const getNameFromObject = (names: LocalizedNames, code: string): string => {
  if (!names) return "";
  if (Array.isArray(names)) {
    const entry = names.find((n) => n.code === code);
    return entry?.name || "";
  }
  const value = names[code];
  return typeof value === "string" ? value : "";
};

const getAddressFromList = (
  list: { code?: string; address?: string }[] | null | undefined,
  code: string,
): string => {
  if (!Array.isArray(list)) return "";
  const entry = list.find((item) => item?.code === code);
  return entry?.address || "";
};

const isVisibleOrganizationRecord = <T extends { isactive?: boolean; deletedat?: string | null }>(record: T): boolean => {
  if (record.isactive === false) return false;
  return !record.deletedat || String(record.deletedat).trim().length === 0;
};

export function CompanyBranchTreeView({
  auth,
  workspace,
  language,
  onRefresh,
}: CompanyBranchTreeViewProps) {
  const [companies, setCompanies] = useState<CompanyRecord[]>([]);
  const [branches, setBranches] = useState<BranchRecord[]>([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [saveSuccess, setSaveSuccess] = useState(false);
  const [saveError, setSaveError] = useState("");
  const [deleteError, setDeleteError] = useState("");
  const [loadError, setLoadError] = useState("");

  const mainApiUrl = useMemo(() => {
    if (!auth?.backendUrl) return "";
    try {
      return deriveMainApiUrl(auth.backendUrl);
    } catch {
      return auth.backendUrl;
    }
  }, [auth]);

  const ensureActiveWorkspaceHolding = useCallback(async () => {
    const holdingcode = workspace?.shop?.holdingcode?.trim();
    if (!auth || !holdingcode) return;

    const res = await fetch("/api/workspace/select-holding", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "x-bc-backend-url": auth.backendUrl,
        Authorization: `Bearer ${auth.token}`,
      },
      body: JSON.stringify({ backendUrl: auth.backendUrl, holdingcode }),
    });
    const json = (await res.json().catch(() => ({}))) as { success?: boolean; message?: string };
    if (!res.ok || json.success === false) {
      throw new Error(json.message || "เลือกบริษัทสำหรับโหลดข้อมูลไม่สำเร็จ");
    }
  }, [auth, workspace]);

  // Active languages for multilingual names
  const editorLanguages = useMemo(() => {
    if (!workspace) return ["th"];
    const configs = workspace.shopInfo?.settings?.languageconfigs || [];
    const defaultCode = workspace.shopInfo?.settings?.language || "th";
    return normalizeLanguageConfigs(configs, defaultCode).map((row) => row.code);
  }, [workspace]);

  const timezoneChoices = useMemo(() => timezoneSelectOptions(language), [language]);
  const workspaceDefaultLanguage = useMemo(
    () => workspace?.shopInfo?.settings?.language || "th",
    [workspace],
  );

  // Collapsed states
  const [collapsedCompanies, setCollapsedCompanies] = useState<Record<string, boolean>>({});

  // Selected Node for Right Form Editing
  const [selectedNode, setSelectedNode] = useState<SelectedNode | null>(null);
  const [formType, setFormType] = useState<OrganizationFormType | null>(null);

  const [confirmOpen, setConfirmOpen] = useState(false);
  const [randomCode, setRandomCode] = useState("");
  const [inputCode, setInputCode] = useState("");
  const [codeError, setCodeError] = useState(false);
  const [confirmAction, setConfirmAction] = useState<ConfirmAction>("save");
  const [pendingDeleteNode, setPendingDeleteNode] = useState<SelectedNode | null>(null);

  const showConfirmCodeDialog = (action: ConfirmAction, node?: SelectedNode) => {
    const code = Math.floor(1000 + Math.random() * 9000).toString();
    setConfirmAction(action);
    setPendingDeleteNode(action === "delete" ? node ?? null : null);
    setDeleteError("");
    setRandomCode(code);
    setInputCode("");
    setCodeError(false);
    setConfirmOpen(true);
  };

  const closeConfirmCodeDialog = () => {
    setConfirmOpen(false);
    setPendingDeleteNode(null);
  };

  const handleConfirmCodeSubmit = () => {
    if (inputCode !== randomCode) {
      setCodeError(true);
      return;
    }

    setConfirmOpen(false);
    if (confirmAction === "delete" && pendingDeleteNode) {
      void handleDelete(pendingDeleteNode);
      setPendingDeleteNode(null);
      return;
    }
    if (confirmAction === "save") {
      void handleSave();
    }
  };

  // Fetch Companies & Branches
  const loadData = useCallback(async () => {
    if (!auth || !mainApiUrl) return;
    setLoading(true);
    setLoadError("");
    try {
      await ensureActiveWorkspaceHolding();

      // Load Companies
      const cacheBuster = Date.now().toString();
      const resComp = await fetch(`${mainApiUrl}/organization/company?_=${cacheBuster}`, {
        headers: { Authorization: `Bearer ${auth.token}` },
        cache: "no-store",
      });
      const jsonComp = await resComp.json();
      if (!resComp.ok || jsonComp.success === false) {
        throw new Error(jsonComp.message || "โหลดข้อมูลบริษัทไม่สำเร็จ");
      }
      if (jsonComp.success && Array.isArray(jsonComp.data)) {
        setCompanies(jsonComp.data.filter(isVisibleOrganizationRecord));
      }

      // Load Branches
      const resBranch = await fetch(`${mainApiUrl}/organization/branch?_=${cacheBuster}`, {
        headers: { Authorization: `Bearer ${auth.token}` },
        cache: "no-store",
      });
      const jsonBranch = await resBranch.json();
      if (!resBranch.ok || jsonBranch.success === false) {
        throw new Error(jsonBranch.message || "โหลดข้อมูลสาขาไม่สำเร็จ");
      }
      if (jsonBranch.success && Array.isArray(jsonBranch.data)) {
        setBranches(jsonBranch.data.filter(isVisibleOrganizationRecord));
      }
    } catch (e) {
      setLoadError(e instanceof Error && e.message ? e.message : "โหลดข้อมูลโครงสร้างองค์กรไม่สำเร็จ");
      console.error(e);
    } finally {
      setLoading(false);
    }
  }, [auth, ensureActiveWorkspaceHolding, mainApiUrl]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  // Right Form Values
  const [formCode, setFormCode] = useState("");
  const [formTaxId, setFormTaxId] = useState("");
  const [formIsActive, setFormIsActive] = useState(true);
  const [formNames, setFormNames] = useState<LocalizedNameEntry[]>([]);
  const [formLogoUri, setFormLogoUri] = useState("");
  const [formTimezone, setFormTimezone] = useState(DEFAULT_TIME_ZONE);
  const [formLanguage, setFormLanguage] = useState("th");
  const [formDateFormat, setFormDateFormat] = useState("dd/MM/yyyy");
  const [formYearType, setFormYearType] = useState("buddhist");
  const [formCurrency, setFormCurrency] = useState("THB");
  const [formBranchType, setFormBranchType] = useState("permanent");
  const [formIsVatRegistered, setFormIsVatRegistered] = useState(false);
  const [formCompanyRegNo, setFormCompanyRegNo] = useState("");
  const [formEmail, setFormEmail] = useState("");
  const [formManagerName, setFormManagerName] = useState("");
  const [formFiscalStartMonth, setFormFiscalStartMonth] = useState(1);
  const [formDocPrefixes, setFormDocPrefixes] = useState<Record<string, string>>({});
  const [formAddresses, setFormAddresses] = useState<Record<string, string>>({});
  const [formETaxEnabled, setFormETaxEnabled] = useState(false);
  const [logoUploading, setLogoUploading] = useState(false);
  const [logoError, setLogoError] = useState("");
  const logoInputRef = React.useRef<HTMLInputElement>(null);

  useEffect(() => {
    if (!selectedNode) return;
    const list: LocalizedNameEntry[] = [];
    const rawNames = selectedNode.data.names;
    editorLanguages.forEach((lang) => {
      list.push({
        code: lang,
        name: getNameFromObject(rawNames, lang),
        isauto: false,
        isdelete: false,
      });
    });

    setFormCode(selectedNode.type === "company" ? normalizeBusinessCode(selectedNode.data.code) : selectedNode.data.code || "");
    setFormIsActive(selectedNode.data.isactive !== false);
    setFormNames(list);
    setFormLogoUri(String(selectedNode.data.logouri ?? ""));
    setLogoError("");

    if (selectedNode.type === "company") {
      setFormTaxId((selectedNode.data as CompanyRecord).taxid || "");
      setFormAddresses({});
    } else {
      const branchData = selectedNode.data as BranchRecord;
      setFormTimezone(branchData.timezone || DEFAULT_TIME_ZONE);
      setFormLanguage(branchData.language || workspaceDefaultLanguage);
      setFormDateFormat(branchData.dateformat || "dd/MM/yyyy");
      setFormYearType(branchData.yeartype || "buddhist");
      setFormCurrency(branchData.basecurrency || "THB");
      setFormBranchType(
        branchData.branchtype || (isThaiHeadOfficeBranchCode(branchData.code) ? "head" : "permanent"),
      );
      setFormIsVatRegistered(branchData.isvatregistered === true);
      setFormCompanyRegNo(branchData.companyregistrationno || "");
      setFormEmail(branchData.email || "");
      setFormManagerName(branchData.managername || "");
      setFormFiscalStartMonth(branchData.fiscalstartmonth || 1);
      setFormDocPrefixes(
        Object.fromEntries(
          (branchData.documentprefixes || [])
            .filter((e) => e.doctype)
            // legacy/long values (e.g. "HQ-SI", "BR01PO") are reset to empty so the doctype-code default shows
            .map((e) => [e.doctype as string, cleanDocPrefixOrEmpty(e.prefix || "")]),
        ),
      );
      setFormETaxEnabled(branchData.etaxenabled === true);
      const addrMap: Record<string, string> = {};
      editorLanguages.forEach((lang) => {
        addrMap[lang] = getAddressFromList(branchData.addresses, lang);
      });
      setFormAddresses(addrMap);
    }
  }, [selectedNode, editorLanguages, workspaceDefaultLanguage]);

  const handleLogoUpload = async (file: File | undefined) => {
    if (!file || !auth || logoUploading) return;
    // PNG-only for logos.
    const isPng = /\.png$/i.test(file.name) || file.type === "image/png";
    if (!isPng) {
      setLogoError("โลโก้ต้องเป็นไฟล์ PNG เท่านั้น");
      return;
    }
    setLogoUploading(true);
    setLogoError("");
    try {
      const uploadForm = new FormData();
      uploadForm.append("file", file, file.name);
      uploadForm.append("category", `system-settings/logouri`);
      const response = await fetch("/api/upload/image", {
        method: "POST",
        headers: {
          "x-bc-backend-url": auth.backendUrl,
          Authorization: `Bearer ${auth.token}`,
        },
        body: uploadForm,
      });
      const payload = (await response.json()) as {
        success?: boolean;
        message?: string;
        data?: {
          holdingcode?: string;
          filename?: string;
          category?: string;
        };
      };
      if (!response.ok || payload.success === false) {
        throw new Error(
          typeof payload.message === "string" ? payload.message : "อัปโหลดโลโก้ไม่สำเร็จ",
        );
      }
      // Backend (image_r2.go) stores images as private R2 objects and only returns
      // holdingcode/filename/category — no URL. We construct the proxy URI
      // `/s3/file/{holdingcode}/{category}/{filename}` which the backend serves
      // through the authenticated S3FileProxyHandler.
      const meta = payload.data ?? {};
      const holding = (meta.holdingcode ?? "").trim();
      const category = (meta.category ?? "").trim();
      const filename = (meta.filename ?? "").trim();
      if (!filename) throw new Error("อัปโหลดโลโก้ไม่สำเร็จ");
      const pathSegments = [holding, category, filename]
        .filter((segment) => segment.length > 0)
        .map((segment) => segment.replace(/^\/+|\/+$/g, ""));
      const proxyUri = `/goapi/s3/file/${pathSegments.join("/")}`;
      setFormLogoUri(proxyUri);
    } catch (error) {
      setLogoError(error instanceof Error && error.message ? error.message : "อัปโหลดโลโก้ไม่สำเร็จ");
    } finally {
      setLogoUploading(false);
      if (logoInputRef.current) logoInputRef.current.value = "";
    }
  };

  // Handle Save
  const handleSave = async () => {
    if (!auth || !selectedNode || !formType || formType.startsWith("view")) return;
    setSaving(true);
    setSaveError("");

    try {
      await ensureActiveWorkspaceHolding();

      const namesList = formNames;
      const normalizedCompanyCode = formType.includes("company") ? normalizeBusinessCode(formCode) : "";
      const normalizedBranchCode = formType.includes("branch") ? normalizeThaiTaxBranchCode(formCode) : "";

      const isBranchForm = formType.includes("branch");
      if (isBranchForm && (!formTimezone.trim() || !formLanguage.trim())) {
        setSaveError("กรุณาเลือกเขตเวลาและภาษาของสาขา");
        setSaving(false);
        return;
      }
      const branchTzMeta = timezoneMeta(formTimezone);
      // ทุกประเภทมีคำนำหน้า default = รหัสประเภท (เช่น PO) เมื่อผู้ใช้ไม่ได้ override
      const docPrefixesPayload = DOC_PREFIX_TYPES.map((dt) => ({
        doctype: dt.code,
        prefix: formDocPrefixes[dt.code] || dt.code,
      }));
      const addressesPayload = editorLanguages
        .map((lang) => ({ code: lang, address: (formAddresses[lang] || "").trim() }))
        .filter((item) => item.address);

      let url = "";
      let method = "POST";
      let body: Record<string, unknown> = {};

      if (formType === "createcompany") {
        url = `${mainApiUrl}/organization/company`;
        method = "POST";
        body = {
          code: normalizedCompanyCode,
          names: namesList,
          taxid: formTaxId,
          logouri: formLogoUri,
          isactive: formIsActive,
        };
      } else if (formType === "editcompany") {
        url = `${mainApiUrl}/organization/company/${selectedNode.guidfixed}`;
        method = "PUT";
        body = {
          code: normalizedCompanyCode,
          names: namesList,
          taxid: formTaxId,
          logouri: formLogoUri,
          isactive: formIsActive,
        };
      } else if (formType === "createbranch") {
        url = `${mainApiUrl}/organization/branch`;
        method = "POST";
        body = {
          companyguid: selectedNode.companyguid,
          code: normalizedBranchCode,
          names: namesList,
          logouri: formLogoUri,
          timezone: formTimezone,
          timezonelabel: branchTzMeta.label,
          timezoneoffset: branchTzMeta.offset,
          language: formLanguage,
          dateformat: formDateFormat,
          yeartype: formYearType,
          basecurrency: formCurrency,
          branchtype: formBranchType,
          isvatregistered: formIsVatRegistered,
          companyregistrationno: formCompanyRegNo,
          email: formEmail,
          managername: formManagerName,
          fiscalstartmonth: formFiscalStartMonth,
          documentprefixes: docPrefixesPayload,
          addresses: addressesPayload,
          etaxenabled: formETaxEnabled,
          isactive: formIsActive,
        };
      } else if (formType === "editbranch") {
        url = `${mainApiUrl}/organization/branch/${selectedNode.guidfixed}`;
        method = "PUT";
        body = {
          companyguid: selectedNode.companyguid,
          code: normalizedBranchCode,
          names: namesList,
          logouri: formLogoUri,
          timezone: formTimezone,
          timezonelabel: branchTzMeta.label,
          timezoneoffset: branchTzMeta.offset,
          language: formLanguage,
          dateformat: formDateFormat,
          yeartype: formYearType,
          basecurrency: formCurrency,
          branchtype: formBranchType,
          isvatregistered: formIsVatRegistered,
          companyregistrationno: formCompanyRegNo,
          email: formEmail,
          managername: formManagerName,
          fiscalstartmonth: formFiscalStartMonth,
          documentprefixes: docPrefixesPayload,
          addresses: addressesPayload,
          etaxenabled: formETaxEnabled,
          isactive: formIsActive,
        };
      }

      const res = await fetch(url, {
        method,
        headers: {
          "Content-Type": "application/json",
          Authorization: `Bearer ${auth.token}`,
        },
        body: JSON.stringify(body),
      });

      const json = (await res.json().catch(() => ({}))) as { success?: boolean; id?: string; message?: string };
      if (!res.ok || json.success === false) {
        setSaveError(saveErrorMessage(json.message, formType));
        return;
      }
      if (json.success) {
        setSaveSuccess(true);
        setTimeout(() => setSaveSuccess(false), 2000);
        await loadData();
        onRefresh?.();
        notifyWorkspaceChanged();
        // Switch back to view mode and refresh the selected node data so the
        // logo (and any other changed fields) show immediately without a manual
        // close/reopen of the form.
        if (formType.startsWith("create")) {
          const createdCode = formType === "createbranch" ? normalizedBranchCode : normalizedCompanyCode;
          const createdData = {
            guidfixed: json.id,
            code: createdCode,
            names: namesList,
            logouri: formLogoUri,
            isactive: formIsActive,
            ...(formType === "createcompany"
              ? { taxid: formTaxId }
              : {
                  timezone: formTimezone,
                  timezonelabel: branchTzMeta.label,
                  timezoneoffset: branchTzMeta.offset,
                  language: formLanguage,
                  dateformat: formDateFormat,
                  yeartype: formYearType,
                  basecurrency: formCurrency,
                  branchtype: formBranchType,
                  isvatregistered: formIsVatRegistered,
                  companyregistrationno: formCompanyRegNo,
                  email: formEmail,
                  managername: formManagerName,
                  fiscalstartmonth: formFiscalStartMonth,
                  documentprefixes: docPrefixesPayload,
                  addresses: addressesPayload,
                  etaxenabled: formETaxEnabled,
                }),
          };
          if (formType === "createcompany") {
            setCompanies((prev) => prev.some((row) => row.guidfixed === json.id) ? prev : [...prev, createdData]);
          } else {
            setBranches((prev) => prev.some((row) => row.guidfixed === json.id) ? prev : [
              ...prev,
              { ...createdData, companyguid: selectedNode.companyguid },
            ]);
          }
          // Select newly created node
          setSelectedNode({
            type: formType === "createcompany" ? "company" : "branch",
            guidfixed: json.id,
            companyguid: selectedNode.companyguid,
            data: createdData,
          });
          setFormType(formType === "createcompany" ? "editcompany" : "editbranch");
        } else {
          // Edit mode: switch back to view mode and refresh selectedNode data so
          // the logo and other changed fields display immediately without a
          // manual close/reopen of the form.
          const updatedData = {
            ...selectedNode.data,
            code: selectedNode.type === "company" ? normalizedCompanyCode : normalizedBranchCode,
            names: namesList,
            logouri: formLogoUri,
            isactive: formIsActive,
            ...(selectedNode.type === "company"
              ? { taxid: formTaxId }
              : {
                  timezone: formTimezone,
                  timezonelabel: branchTzMeta.label,
                  timezoneoffset: branchTzMeta.offset,
                  language: formLanguage,
                  dateformat: formDateFormat,
                  yeartype: formYearType,
                  basecurrency: formCurrency,
                  branchtype: formBranchType,
                  isvatregistered: formIsVatRegistered,
                  companyregistrationno: formCompanyRegNo,
                  email: formEmail,
                  managername: formManagerName,
                  fiscalstartmonth: formFiscalStartMonth,
                  documentprefixes: docPrefixesPayload,
                  addresses: addressesPayload,
                  etaxenabled: formETaxEnabled,
                }),
          };
          setSelectedNode({
            ...selectedNode,
            data: updatedData,
          });
          setFormType(selectedNode.type === "company" ? "viewcompany" : "viewbranch");
        }
      }
    } catch (e) {
      setSaveError(e instanceof Error && e.message ? e.message : "บันทึกข้อมูลไม่สำเร็จ");
      console.error(e);
    } finally {
      setSaving(false);
    }
  };

  // Handle Delete
  const handleDelete = async (node: SelectedNode) => {
    if (!auth || !node.guidfixed) return;

    setLoading(true);
    try {
      await ensureActiveWorkspaceHolding();

      const url = `${mainApiUrl}/organization/${node.type}/${node.guidfixed}`;
      const res = await fetch(url, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${auth.token}` },
      });
      const json = (await res.json().catch(() => ({}))) as { success?: boolean; message?: string };
      if (!res.ok || json.success === false) {
        throw new Error(deleteErrorMessage(json.message, node));
      }
      if (json.success) {
        await loadData();
        pruneDeletedNode(node);
        onRefresh?.();
        notifyWorkspaceChanged();
        setSelectedNode(null);
        setFormType(null);
        setDeleteError("");
      }
    } catch (e) {
      setDeleteError(e instanceof Error && e.message ? e.message : "ลบข้อมูลไม่สำเร็จ");
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

  const pruneDeletedNode = (node: SelectedNode) => {
    const deletedGuid = node.guidfixed ?? "";
    if (!deletedGuid) return;
    if (node.type === "company") {
      setCompanies((prev) => prev.filter((company) => company.guidfixed !== deletedGuid));
      setBranches((prev) => prev.filter((branch) => branch.companyguid !== deletedGuid));
      return;
    }
    setBranches((prev) => prev.filter((branch) => branch.guidfixed !== deletedGuid));
  };

  const sortedCompanies = useMemo(() => {
    return [...companies].sort((a, b) => {
      const codeA = normalizeBusinessCode(a.code);
      const codeB = normalizeBusinessCode(b.code);
      const cmp = codeA.localeCompare(codeB, undefined, { numeric: true, sensitivity: "base" });
      if (cmp !== 0) return cmp;

      const nameA = getNameFromObject(a.names, language) || "";
      const nameB = getNameFromObject(b.names, language) || "";
      return nameA.localeCompare(nameB, "th", { sensitivity: "base" });
    });
  }, [companies, language]);

  const sortedBranches = useMemo(() => {
    return [...branches].sort((a, b) => {
      const codeA = a.code || "";
      const codeB = b.code || "";
      const cmp = codeA.localeCompare(codeB, undefined, { numeric: true, sensitivity: "base" });
      if (cmp !== 0) return cmp;

      const nameA = getNameFromObject(a.names, language) || "";
      const nameB = getNameFromObject(b.names, language) || "";
      return nameA.localeCompare(nameB, "th", { sensitivity: "base" });
    });
  }, [branches, language]);

  const isReadOnlyMode = formType?.startsWith("view") ?? false;

  return (
    <div className="grid w-full grid-cols-1 gap-4 xl:grid-cols-[minmax(320px,0.85fr)_minmax(420px,1.15fr)]">
      {/* Left panel: Company -> Branch list */}
      <Card className="min-h-[calc(100vh-12rem)] shadow-lg border-primary/10">
        <CardContent className="p-4">
          <div className="flex items-center justify-between border-b pb-3 mb-4">
            <h2 className="text-lg font-bold text-foreground flex items-center gap-2">
              <Building2 className="w-5 h-5 text-primary" />
              โครงสร้างองค์กร
            </h2>
            <Button
              size="sm"
              className="gap-1 font-semibold"
              onClick={() => {
                setFormType("createcompany");
                setSelectedNode({
                  type: "company",
                  data: {},
                });
              }}
            >
              <Plus className="w-4 h-4" />
              เพิ่มบริษัท
            </Button>
          </div>
          {deleteError && (
            <div className="mb-3 rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs font-semibold text-destructive">
              {deleteError}
            </div>
          )}
          {loadError && (
            <div className="mb-3 rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs font-semibold text-destructive">
              {loadError}
            </div>
          )}

          {loading ? (
            <div className="flex justify-center py-12">
              <Loader2 className="w-8 h-8 animate-spin text-primary" />
            </div>
          ) : sortedCompanies.length === 0 ? (
            <div className="rounded-lg border border-dashed border-border bg-muted/30 px-3 py-8 text-center text-sm font-semibold text-muted-foreground">
              ยังไม่มีข้อมูลบริษัทในกิจการนี้
            </div>
          ) : (
            <div className="space-y-2">
              {sortedCompanies.map((comp) => {
                const compGuid = comp.guidfixed || "";
                const companyCode = normalizeBusinessCode(comp.code);
                const isCollapsed = collapsedCompanies[compGuid];
                const isSelected = selectedNode?.type === "company" && selectedNode.guidfixed === compGuid;
                const isEditingCompany = isSelected && formType === "editcompany";
                const compBranches = sortedBranches.filter((b) => b.companyguid === compGuid);

                return (
                  <div key={compGuid} className="space-y-1">
                    <div
                      className={cn(
                        "group flex items-center justify-between p-2 rounded-lg cursor-pointer transition-all",
                        isEditingCompany
                          ? "bg-amber-100/70 text-amber-950 dark:bg-amber-950/40 dark:text-amber-100 border-l-4 border-amber-400 pl-1 font-bold"
                          : isSelected
                            ? "bg-primary/10 border-l-4 border-primary pl-1 font-bold"
                            : "hover:bg-accent pl-2"
                      )}
                      onClick={() => {
                        setSelectedNode({
                          type: "company",
                          guidfixed: compGuid,
                          data: comp,
                        });
                        setFormType("viewcompany");
                      }}
                    >
                      <div className="flex items-center gap-2 min-w-0">
                        <button
                          type="button"
                          aria-label={language === "th" ? "ย่อ/ขยายสาขา" : "Toggle branches"}
                          onClick={(e) => {
                            e.stopPropagation();
                            setCollapsedCompanies((prev) => ({
                              ...prev,
                              [compGuid]: !prev[compGuid],
                            }));
                          }}
                          className="p-1 hover:bg-black/5 dark:hover:bg-white/5 rounded"
                        >
                          {isCollapsed ? (
                            <ChevronRight className="w-4 h-4 text-muted-foreground" />
                          ) : (
                            <ChevronDown className="w-4 h-4 text-muted-foreground" />
                          )}
                        </button>
                        <LogoAvatar
                          uri={comp.logouri}
                          auth={auth}
                          alt={getNameFromObject(comp.names, language) || companyCode}
                          sizeClass="size-5 rounded-md shrink-0"
                          iconSize={14}
                          width={64}
                          className="text-primary"
                        />
                        <span className="min-w-0 break-words">
                          {getNameFromObject(comp.names, language) || companyCode}
                        </span>
                        {companyCode ? (
                          <span className="font-mono text-[9px] px-1.5 py-0.2 bg-muted border border-border/50 text-muted-foreground rounded uppercase font-bold shrink-0">
                            {companyCode}
                          </span>
                        ) : null}
                      </div>
                      <div className="flex items-center gap-1">
                        <Button
                          size="icon"
                          variant="ghost"
                          className="w-7 h-7 text-primary hover:text-primary hover:bg-primary/10"
                          title="แก้ไขบริษัท"
                          onClick={(e) => {
                            e.stopPropagation();
                            setSelectedNode({
                              type: "company",
                              guidfixed: compGuid,
                              data: comp,
                            });
                            setFormType("editcompany");
                          }}
                        >
                          <Edit3 className="w-3.5 h-3.5" />
                        </Button>
                        <Button
                          size="icon"
                          variant="ghost"
                          className="w-7 h-7 text-sky-500 hover:text-sky-600 hover:bg-sky-500/10"
                          title="เพิ่มสาขา"
                          onClick={(e) => {
                            e.stopPropagation();
                            setSelectedNode({
                              type: "branch",
                              companyguid: compGuid,
                              data: {},
                            });
                            setFormType("createbranch");
                          }}
                        >
                          <Plus className="w-3.5 h-3.5" />
                        </Button>
                        <Button
                          size="icon"
                          variant="ghost"
                          className="w-7 h-7 text-muted-foreground/60 hover:text-destructive hover:bg-destructive/10"
                          onClick={(e) => {
                            e.stopPropagation();
                            showConfirmCodeDialog("delete", {
                              type: "company",
                              guidfixed: compGuid,
                              data: comp,
                            });
                          }}
                        >
                          <Trash2 className="w-3.5 h-3.5" />
                        </Button>
                      </div>
                    </div>

                    {/* Branches list */}
                    {!isCollapsed && compBranches.length > 0 && (
                      <div className="pl-6 border-l ml-4 space-y-1 my-1">
                        {compBranches.map((br) => {
                          const brGuid = br.guidfixed || "";
                          const isBrSelected = selectedNode?.type === "branch" && selectedNode.guidfixed === brGuid;
                          const isEditingBranch = isBrSelected && formType === "editbranch";
                          const cannotDeleteBranch = isThaiHeadOfficeBranchCode(br.code) || compBranches.length <= 1;

                          return (
                            <div
                              key={brGuid}
                              className={cn(
                                "group flex items-center justify-between p-2 rounded-lg cursor-pointer transition-all",
                                isEditingBranch
                                  ? "bg-amber-100/70 text-amber-950 dark:bg-amber-950/40 dark:text-amber-100 border-l-4 border-amber-400 pl-1 font-semibold"
                                  : isBrSelected
                                    ? "bg-sky-500/10 border-l-4 border-sky-500 pl-1 font-semibold"
                                    : "hover:bg-accent pl-2"
                              )}
                              onClick={() => {
                                setSelectedNode({
                                  type: "branch",
                                  guidfixed: brGuid,
                                  companyguid: compGuid,
                                  data: br,
                                });
                                setFormType("viewbranch");
                              }}
                            >
                              <div className="flex items-center gap-2 min-w-0">
                                <LogoAvatar
                                  uri={br.logouri}
                                  auth={auth}
                                  alt={getNameFromObject(br.names, language) || br.code || ""}
                                  sizeClass="size-5 rounded-md shrink-0"
                                  iconSize={14}
                                  width={64}
                                  className="text-sky-500"
                                />
                                <span className="min-w-0 text-sm break-words">
                                  {getNameFromObject(br.names, language) || br.code}
                                </span>
                                {br.code ? (
                                  <span className="font-mono text-[9px] px-1.5 py-0.2 bg-muted border border-border/50 text-muted-foreground rounded uppercase font-bold shrink-0">
                                    {br.code}
                                  </span>
                                ) : null}
                              </div>
                              <div className="flex items-center gap-1">
                                <Button
                                  size="icon"
                                  variant="ghost"
                                  className="w-7 h-7 text-primary hover:text-primary hover:bg-primary/10"
                                  title="แก้ไขสาขา"
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    setSelectedNode({
                                      type: "branch",
                                      guidfixed: brGuid,
                                      companyguid: compGuid,
                                      data: br,
                                    });
                                    setFormType("editbranch");
                                  }}
                                >
                                  <Edit3 className="w-3.5 h-3.5" />
                                </Button>
                                <Button
                                  size="icon"
                                  variant="ghost"
                                  className="w-7 h-7 text-muted-foreground/60 hover:text-destructive hover:bg-destructive/10"
                                  disabled={cannotDeleteBranch}
                                  title={cannotDeleteBranch ? "สาขาสำนักงานใหญ่หรือสาขาสุดท้ายของบริษัทลบไม่ได้" : "ลบสาขา"}
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    if (cannotDeleteBranch) return;
                                    showConfirmCodeDialog("delete", {
                                      type: "branch",
                                      guidfixed: brGuid,
                                      companyguid: compGuid,
                                      data: br,
                                    });
                                  }}
                                >
                                  <Trash2 className="w-3.5 h-3.5" />
                                </Button>
                              </div>
                            </div>
                          );
                        })}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Right panel: Detail / Edit Form */}
      <Card className="shadow-lg border-primary/10">
        <CardContent className="p-6">
          {formType ? (
            <div className="space-y-6">
              <div className="flex items-center justify-between border-b pb-4 gap-4">
                <div>
                  <h3 className="text-xl font-bold text-foreground flex items-center gap-2">
                    {formType.includes("company") ? (
                      <Building2 className="w-5.5 h-5.5 text-primary shrink-0" />
                    ) : (
                      <GitBranch className="w-5.5 h-5.5 text-sky-500 shrink-0" />
                    )}
                    {formType === "viewcompany" && "ข้อมูลบริษัท"}
                    {formType === "viewbranch" && "ข้อมูลสาขา"}
                    {formType === "createcompany" && "เพิ่มบริษัทใหม่"}
                    {formType === "editcompany" && "แก้ไขข้อมูลบริษัท"}
                    {formType === "createbranch" && "เพิ่มสาขาใหม่"}
                    {formType === "editbranch" && "แก้ไขข้อมูลสาขา"}
                  </h3>
                  <p className="text-sm text-muted-foreground mt-1">
                    {formType.startsWith("create")
                      ? "ระบุข้อมูลรายละเอียดหลักเพื่อเพิ่มข้อมูลเข้าระบบ"
                      : isReadOnlyMode
                        ? "แสดงรายละเอียดข้อมูลจากรายการที่เลือก"
                        : "แก้ไขรายละเอียดข้อมูลและบันทึกประวัติ"}
                  </p>
                </div>
                <div className="flex shrink-0 items-center gap-2">
                  {isReadOnlyMode && selectedNode?.guidfixed && (
                    <Button
                      size="icon"
                      variant="outline"
                      className="size-9 border-primary/30 text-primary hover:bg-primary/10"
                      title={selectedNode.type === "company" ? "แก้ไขบริษัท" : "แก้ไขสาขา"}
                      onClick={() => setFormType(selectedNode.type === "company" ? "editcompany" : "editbranch")}
                    >
                      <Edit3 className="w-4 h-4" />
                    </Button>
                  )}
                  {(formType === "viewcompany" || formType === "editcompany") && selectedNode?.guidfixed && (
                    <Button
                      size="sm"
                      variant="outline"
                      className="gap-1 text-xs border-sky-500/30 text-sky-600 hover:bg-sky-500/10 hover:text-sky-700 font-bold shrink-0"
                      onClick={() => {
                        setSelectedNode({
                          type: "branch",
                          companyguid: selectedNode.guidfixed,
                          data: {},
                        });
                        setFormType("createbranch");
                      }}
                    >
                      <Plus className="w-3.5 h-3.5" />
                      เพิ่มสาขาในบริษัทนี้
                    </Button>
                  )}
                </div>
              </div>

              <div className="space-y-4">
                {/* Logo */}
                <div className="space-y-2">
                  <label className="text-sm font-semibold text-foreground">
                    {formType.includes("company") ? "โลโก้บริษัท" : "โลโก้สาขา"}
                  </label>
                  <div className="flex items-start gap-3">
                    <LogoAvatar
                      uri={formLogoUri}
                      auth={auth}
                      alt={formType.includes("company") ? "โลโก้บริษัท" : "โลโก้สาขา"}
                      sizeClass="size-20 rounded-2xl"
                      iconSize={32}
                      width={256}
                      className="border border-input"
                    />
                    {!isReadOnlyMode && (
                      <div className="flex flex-col gap-1.5 min-w-48 flex-1">
                        <div className="flex gap-2">
                          <Button
                            type="button"
                            variant="outline"
                            size="sm"
                            disabled={logoUploading}
                            onClick={() => logoInputRef.current?.click()}
                          >
                            {logoUploading ? <Loader2 className="w-4 h-4 animate-spin" /> : <UploadCloud className="w-4 h-4" />}
                            {logoUploading ? "กำลังอัปโหลด..." : formLogoUri ? "เปลี่ยนโลโก้" : "เลือกไฟล์ PNG"}
                          </Button>
                          {formLogoUri && (
                            <Button
                              type="button"
                              variant="outline"
                              size="sm"
                              className="text-destructive hover:bg-destructive/5"
                              disabled={logoUploading}
                              onClick={() => {
                                setFormLogoUri("");
                                setLogoError("");
                              }}
                            >
                              <X className="w-4 h-4" /> ลบ
                            </Button>
                          )}
                        </div>
                        <p className="text-xs text-muted-foreground">
                          รองรับเฉพาะไฟล์ PNG พื้นหลังโปร่งใสได้ ใช้สำหรับออกแบบฟอร์มและพิมพ์เอกสาร
                        </p>
                        {logoError && (
                          <p className="text-xs font-semibold text-destructive">{logoError}</p>
                        )}
                        <input
                          ref={logoInputRef}
                          className="sr-only"
                          type="file"
                          accept="image/png"
                          onChange={(e) => void handleLogoUpload(e.target.files?.[0])}
                        />
                      </div>
                    )}
                  </div>
                </div>

                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <label className="text-sm font-semibold text-foreground">
                      {formType.includes("company") ? "รหัสบริษัท *" : "รหัสสาขา *"}
                    </label>
                    <Input
                      value={formCode}
                      onChange={(e) => {
                        setFormCode(formType.includes("company") ? normalizeBusinessCode(e.target.value) : e.target.value);
                      }}
                      placeholder={formType.includes("company") ? "ระบุรหัสบริษัท เช่น 00000" : "ระบุรหัสสาขา 5 หลัก เช่น 00001"}
                      className="bg-accent/20"
                      disabled={isReadOnlyMode}
                    />
                  </div>
                  {formType.includes("company") && (
                    <div className="space-y-2">
                      <label className="text-sm font-semibold text-foreground">เลขประจำตัวผู้เสียภาษี (Tax ID)</label>
                      <Input
                        value={formTaxId}
                        onChange={(e) => setFormTaxId(e.target.value)}
                        placeholder="เลขผู้เสียภาษี 13 หลัก"
                        className="bg-accent/20"
                        disabled={isReadOnlyMode}
                      />
                    </div>
                  )}
                </div>

                {/* Branch locale: timezone + language (both required — branches may differ) */}
                {formType.includes("branch") && (
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                    <div className="space-y-2">
                      <label className="text-sm font-semibold text-foreground">เขตเวลา (Timezone) *</label>
                      <select
                        value={formTimezone}
                        onChange={(e) => setFormTimezone(e.target.value)}
                        disabled={isReadOnlyMode}
                        className="flex h-10 w-full rounded-md border border-input bg-accent/20 px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                      >
                        {timezoneChoices.map((tz) => (
                          <option key={tz.value} value={tz.value}>
                            {tz.label}
                          </option>
                        ))}
                      </select>
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-semibold text-foreground">ภาษาของสาขา *</label>
                      <select
                        value={formLanguage}
                        onChange={(e) => setFormLanguage(e.target.value)}
                        disabled={isReadOnlyMode}
                        className="flex h-10 w-full rounded-md border border-input bg-accent/20 px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                      >
                        {LANGUAGES.map((lang) => (
                          <option key={lang.code} value={lang.code}>
                            {lang.name}
                          </option>
                        ))}
                      </select>
                    </div>
                  </div>
                )}

                {/* Branch date/currency settings (backend already carries these) */}
                {formType.includes("branch") && (
                  <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                    <div className="space-y-2">
                      <label className="text-sm font-semibold text-foreground">ประเภทปี</label>
                      <select
                        value={formYearType}
                        onChange={(e) => setFormYearType(e.target.value)}
                        disabled={isReadOnlyMode}
                        className="flex h-10 w-full rounded-md border border-input bg-accent/20 px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                      >
                        {YEAR_TYPE_OPTIONS.map((opt) => (
                          <option key={opt.value} value={opt.value}>
                            {opt.label}
                          </option>
                        ))}
                      </select>
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-semibold text-foreground">รูปแบบวันที่</label>
                      <select
                        value={formDateFormat}
                        onChange={(e) => setFormDateFormat(e.target.value)}
                        disabled={isReadOnlyMode}
                        className="flex h-10 w-full rounded-md border border-input bg-accent/20 px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                      >
                        {dateFormatOptionsFor(formYearType).map((opt) => (
                          <option key={opt.value} value={opt.value}>
                            {opt.label}
                          </option>
                        ))}
                      </select>
                    </div>
                    <div className="space-y-2">
                      <label className="text-sm font-semibold text-foreground">สกุลเงินหลัก</label>
                      <select
                        value={formCurrency}
                        onChange={(e) => setFormCurrency(e.target.value)}
                        disabled={isReadOnlyMode}
                        className="flex h-10 w-full rounded-md border border-input bg-accent/20 px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                      >
                        {CURRENCY_OPTIONS.map((opt) => (
                          <option key={opt.value} value={opt.value}>
                            {opt.label}
                          </option>
                        ))}
                      </select>
                    </div>
                  </div>
                )}

                {/* Branch tax / registration (ภ.พ.20) */}
                {formType.includes("branch") && (
                  <div className="space-y-4 rounded-lg border border-border/60 bg-muted/20 p-4">
                    <p className="text-sm font-bold text-foreground">ข้อมูลภาษี / ทะเบียน (ภ.พ.20)</p>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <label className="text-sm font-semibold text-foreground">ประเภทสาขา (ภ.พ.20)</label>
                        <select
                          value={formBranchType}
                          onChange={(e) => setFormBranchType(e.target.value)}
                          disabled={isReadOnlyMode}
                          className="flex h-10 w-full rounded-md border border-input bg-accent/20 px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                        >
                          {BRANCH_TYPE_OPTIONS.map((opt) => (
                            <option key={opt.value} value={opt.value}>
                              {opt.label}
                            </option>
                          ))}
                        </select>
                      </div>
                      <div className="space-y-2">
                        <label className="text-sm font-semibold text-foreground">เลขทะเบียนนิติบุคคล</label>
                        <Input
                          value={formCompanyRegNo}
                          onChange={(e) => setFormCompanyRegNo(e.target.value)}
                          placeholder="เลขทะเบียนนิติบุคคล 13 หลัก"
                          className="bg-accent/20"
                          disabled={isReadOnlyMode}
                        />
                      </div>
                    </div>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <label className="text-sm font-semibold text-foreground">ผู้จัดการสาขา</label>
                        <Input
                          value={formManagerName}
                          onChange={(e) => setFormManagerName(e.target.value)}
                          placeholder="ชื่อผู้จัดการสาขา"
                          className="bg-accent/20"
                          disabled={isReadOnlyMode}
                        />
                      </div>
                      <div className="space-y-2">
                        <label className="text-sm font-semibold text-foreground">อีเมลสาขา</label>
                        <Input
                          type="email"
                          value={formEmail}
                          onChange={(e) => setFormEmail(e.target.value)}
                          placeholder="อีเมลสำหรับส่งเอกสาร"
                          className="bg-accent/20"
                          disabled={isReadOnlyMode}
                        />
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <input
                        type="checkbox"
                        id="isvatregistered"
                        checked={formIsVatRegistered}
                        onChange={(e) => setFormIsVatRegistered(e.target.checked)}
                        disabled={isReadOnlyMode}
                        className="w-4 h-4 text-primary border-gray-300 rounded focus:ring-primary"
                      />
                      <label htmlFor="isvatregistered" className="text-sm font-semibold text-foreground cursor-pointer select-none">
                        จดทะเบียนภาษีมูลค่าเพิ่ม (VAT)
                      </label>
                    </div>
                  </div>
                )}

                {/* Document / fiscal config — value-only (generator + e-Tax engine are separate, not active) */}
                {formType.includes("branch") && (
                  <div className="space-y-4 rounded-lg border border-border/60 bg-muted/20 p-4">
                    <p className="text-sm font-bold text-foreground">รอบบัญชี / เอกสาร</p>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                      <div className="space-y-2">
                        <label className="text-sm font-semibold text-foreground">เดือนเริ่มรอบบัญชี</label>
                        <select
                          value={formFiscalStartMonth}
                          onChange={(e) => setFormFiscalStartMonth(Number(e.target.value))}
                          disabled={isReadOnlyMode}
                          className="flex h-10 w-full rounded-md border border-input bg-accent/20 px-3 py-2 text-sm ring-offset-background focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
                        >
                          {MONTH_OPTIONS.map((opt) => (
                            <option key={opt.value} value={opt.value}>
                              {opt.label}
                            </option>
                          ))}
                        </select>
                      </div>
                    </div>

                    {/* คำนำหน้าเลขที่เอกสาร แยกตามประเภท (เก็บค่า config — generator ยังไม่ใช้) */}
                    <div className="space-y-2">
                      <label className="text-sm font-semibold text-foreground">
                        คำนำหน้าเลขที่เอกสาร (แยกตามประเภท)
                        <span className="ml-1 font-normal text-xs text-muted-foreground">— 2 ตัวอักษร + ปีเดือนวัน + running เช่น PO → PO26062800001</span>
                      </label>
                      <div className="grid grid-cols-1 md:grid-cols-2 gap-x-4 gap-y-2">
                        {DOC_PREFIX_TYPES.map((dt) => (
                          <div key={dt.code} className="flex items-center gap-2">
                            <span className="w-36 shrink-0 text-xs text-muted-foreground">
                              {dt.label} <span className="font-mono">({dt.code})</span>
                            </span>
                            <Input
                              value={formDocPrefixes[dt.code] || dt.code}
                              onChange={(e) =>
                                setFormDocPrefixes((prev) => ({ ...prev, [dt.code]: normalizeDocPrefix(e.target.value) }))
                              }
                              maxLength={2}
                              placeholder={dt.code}
                              className="bg-accent/20 h-8 text-sm"
                              disabled={isReadOnlyMode}
                            />
                          </div>
                        ))}
                      </div>
                    </div>
                    <div className="flex items-center gap-2">
                      <input
                        type="checkbox"
                        id="etaxenabled"
                        checked={formETaxEnabled}
                        onChange={(e) => setFormETaxEnabled(e.target.checked)}
                        disabled={isReadOnlyMode}
                        className="w-4 h-4 text-primary border-gray-300 rounded focus:ring-primary"
                      />
                      <label htmlFor="etaxenabled" className="text-sm font-semibold text-foreground cursor-pointer select-none">
                        เปิดใช้ใบกำกับภาษีอิเล็กทรอนิกส์ (e-Tax)
                      </label>
                    </div>
                    <p className="text-xs text-muted-foreground">
                      * การออกเลขที่เอกสารอัตโนมัติและการส่ง e-Tax เป็นระบบแยก ยังไม่เปิดใช้งาน — ค่านี้เก็บไว้ตั้งค่าล่วงหน้า
                    </p>
                  </div>
                )}

                {/* Multilingual names */}
                <div className="space-y-3">
                  <NamesEditor
                    names={formNames}
                    onChange={setFormNames}
                    languages={editorLanguages}
                    label={formType.includes("company") ? "ชื่อบริษัท" : "ชื่อสาขา"}
                    language={language}
                    disabled={isReadOnlyMode}
                  />
                  {formType.includes("branch") && (
                    <AddressesEditor
                      addresses={formAddresses}
                      onChange={setFormAddresses}
                      languages={editorLanguages}
                      label={language === "th" ? "ที่อยู่สาขา (สำหรับออกเอกสาร)" : "Branch address (for documents)"}
                      language={language}
                      disabled={isReadOnlyMode}
                    />
                  )}
                </div>
                {saveError && (
                  <div className="rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs font-semibold text-destructive">
                    {saveError}
                  </div>
                )}

                 <div className="flex items-center gap-2 pt-2">
                  <input
                    type="checkbox"
                    id="isactive"
                    checked={formIsActive}
                    onChange={(e) => setFormIsActive(e.target.checked)}
                    disabled={isReadOnlyMode}
                    className="w-4 h-4 text-primary border-gray-300 rounded focus:ring-primary"
                  />
                  <label htmlFor="isactive" className="text-sm font-semibold text-foreground cursor-pointer select-none">
                    เปิดใช้งานในระบบ
                  </label>
                </div>

                {!isReadOnlyMode && (
                <div className="pt-6 border-t mt-4">
                  <Button
                    onClick={() => showConfirmCodeDialog("save")}
                    disabled={
                      !formCode.trim() ||
                      saving ||
                      saveSuccess ||
                      (formType.includes("branch") && (!formTimezone.trim() || !formLanguage.trim()))
                    }
                    className="w-full font-bold bg-primary text-primary-foreground hover:bg-primary/90 rounded-full h-11"
                  >
                    {saving ? (
                      <span className="flex items-center gap-2">
                        <Loader2 className="w-4 h-4 animate-spin" />
                        กำลังบันทึกข้อมูล...
                      </span>
                    ) : saveSuccess ? (
                      "บันทึกข้อมูลสำเร็จ"
                    ) : (
                      "บันทึกข้อมูล"
                    )}
                  </Button>
                </div>
                )}
              </div>
            </div>
          ) : (
            <div className="flex flex-col items-center justify-center py-24 text-muted-foreground">
              <Building2 className="w-16 h-16 text-muted-foreground/30 mb-4" />
              <p className="text-lg font-semibold">กรุณาเลือก บริษัท หรือ สาขา</p>
              <p className="text-sm mt-1">คลิกเลือกรายการที่แถบเมนูด้านซ้ายเพื่อดูหรือแก้ไขข้อมูล</p>
            </div>
          )}
        </CardContent>
      </Card>
      {confirmOpen && (
        <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 backdrop-blur-sm p-4">
          <div
            role="dialog"
            aria-modal="true"
            aria-label={confirmAction === "delete" ? "ยืนยันการลบข้อมูล" : "ยืนยันการบันทึกข้อมูล"}
            className="bg-card border rounded-2xl w-full max-w-sm p-6 shadow-2xl space-y-4 animate-in fade-in zoom-in duration-200"
          >
            <div className="text-center space-y-2">
              <div className="w-12 h-12 rounded-full bg-primary/10 text-primary flex items-center justify-center mx-auto">
                <KeyRound size={22} className="animate-pulse" />
              </div>
              <h3 className="text-lg font-bold text-foreground">
                {confirmAction === "delete" ? "ยืนยันการลบข้อมูล" : "ยืนยันการบันทึกข้อมูล"}
              </h3>
              <p className="text-xs text-muted-foreground">
                {confirmAction === "delete"
                  ? "กรุณากรอกรหัสยืนยันตัวเลข 4 หลักเพื่อดำเนินการลบข้อมูลโครงสร้างองค์กร"
                  : "กรุณากรอกรหัสยืนยันตัวเลข 4 หลักเพื่อดำเนินการบันทึกข้อมูลโครงสร้างองค์กร"}
              </p>
            </div>

            <div className="bg-accent/40 rounded-xl p-3 border border-border/80 text-center">
              <span className="text-xs font-semibold text-muted-foreground block mb-1">รหัสยืนยันของคุณคือ</span>
              <span className="text-2xl font-black tracking-widest text-primary font-mono select-none">{randomCode}</span>
            </div>

            <div className="space-y-1.5">
              <Input
                value={inputCode}
                onChange={(e) => {
                  setInputCode(e.target.value);
                  if (codeError) setCodeError(false);
                }}
                placeholder="กรอกรหัส 4 หลักที่แสดงด้านบน"
                className={`bg-accent/20 h-11 text-center font-bold tracking-widest font-mono text-base ${codeError ? "border-destructive focus-visible:ring-destructive" : ""}`}
                maxLength={4}
              />
              {codeError && (
                <p className="text-[10px] text-destructive font-semibold text-center">รหัสยืนยันไม่ถูกต้อง กรุณาลองใหม่อีกครั้ง</p>
              )}
            </div>

            <div className="flex gap-3 pt-2">
              <Button
                variant="outline"
                className="flex-1 rounded-xl h-11 text-xs font-semibold"
                onClick={closeConfirmCodeDialog}
              >
                ยกเลิก
              </Button>
              <Button
                className={cn(
                  "flex-1 rounded-xl h-11 text-xs font-semibold",
                  confirmAction === "delete" && "bg-destructive text-destructive-foreground hover:bg-destructive/90"
                )}
                onClick={handleConfirmCodeSubmit}
                disabled={inputCode.length !== 4}
              >
                {confirmAction === "delete" ? "ยืนยันลบ" : "ยืนยันบันทึก"}
              </Button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}

function deleteErrorMessage(message: string | undefined, node: SelectedNode): string {
  switch (message) {
    case "head office branch cannot be deleted":
      return "ลบสาขาสำนักงานใหญ่ไม่ได้";
    case "company must have at least one branch":
      return "บริษัทต้องมีอย่างน้อย 1 สาขา";
    case "Branch not found":
      return "ไม่พบข้อมูลสาขาที่ต้องการลบ";
    case "Company not found":
      return "ไม่พบข้อมูลบริษัทที่ต้องการลบ";
    default:
      return message || (node.type === "company" ? "ลบบริษัทไม่สำเร็จ" : "ลบสาขาไม่สำเร็จ");
  }
}

function saveErrorMessage(message: string | undefined, formType: OrganizationFormType): string {
  void formType;
  switch (message) {
    case "branch code is required":
      return "กรุณากรอกรหัสสาขา";
    case "branch code must be numeric and no more than 5 digits":
    case "branch code must be no more than 5 digits":
      return "รหัสสาขาต้องเป็นตัวเลขไม่เกิน 5 หลัก";
    case "companyguid is required":
      return "ไม่พบบริษัทของสาขาที่กำลังเพิ่ม กรุณากดเพิ่มสาขาจากบริษัทอีกครั้ง";
    case "company not found":
      return "ไม่พบบริษัทในกิจการนี้ กรุณาโหลดข้อมูลใหม่แล้วลองอีกครั้ง";
    case "branch code is exists":
      return "รหัสสาขานี้มีอยู่แล้วในบริษัทนี้";
    default:
      if (message?.includes("duplicate key")) return "รหัสนี้ซ้ำกับข้อมูลเดิม";
      return message || "บันทึกข้อมูลไม่สำเร็จ";
  }
}
