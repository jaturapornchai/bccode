"use client";

import { authFetch } from "@/lib/client-auth-session";
import React, { useCallback, useMemo, useState, useEffect } from "react";
import {
  MapPin,
  Plus,
  ChevronRight,
  ChevronDown,
  Edit3,
  Trash2,
  GripVertical,
  Save,
  Loader2,
  Warehouse,
  Boxes,
  PackageSearch,
  AlertTriangle,
} from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { LANGUAGES, type LanguageCode } from "@/lib/i18n";
import { deriveMainApiUrl } from "@/lib/backend-url";
import { cn } from "@/lib/utils";
import { normalizeLanguageConfigs } from "./system-settings-screen";
import { MapPickerDialog } from "@/components/map-picker-dialog";
import { NamesEditor } from "@/components/product-barcode/names-editor";
import { type NameX } from "@/lib/product-barcode/types";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";

const WAREHOUSE_LEVEL_STYLES = [
  {
    caret: "text-primary",
    grip: "text-primary/45 group-hover/row:text-primary",
    name: "text-[15px] font-bold text-foreground",
    order: "text-primary",
    selectedBg: "bg-primary/5 text-primary font-semibold",
    borderLeft: "bg-primary",
  },
  {
    caret: "text-sky-600 dark:text-sky-300",
    grip: "text-sky-500/55 group-hover/row:text-sky-600 dark:group-hover/row:text-sky-300",
    name: "font-semibold text-sky-900 dark:text-sky-100",
    order: "text-sky-600 dark:text-sky-300",
    selectedBg: "bg-sky-500/10 text-sky-600 font-semibold",
    borderLeft: "bg-sky-500",
  },
  {
    caret: "text-emerald-600 dark:text-emerald-300",
    grip: "text-emerald-500/55 group-hover/row:text-emerald-600 dark:group-hover/row:text-emerald-300",
    name: "font-semibold text-emerald-900 dark:text-emerald-100",
    order: "text-emerald-600 dark:text-emerald-300",
    selectedBg: "bg-emerald-500/10 text-emerald-600 font-semibold",
    borderLeft: "bg-emerald-500",
  },
] as const;

const warehouseLevelStyle = (level: number) =>
  WAREHOUSE_LEVEL_STYLES[Math.min(Math.max(level, 0), WAREHOUSE_LEVEL_STYLES.length - 1)];

interface WarehouseTreeViewProps {
  auth: { token: string; backendUrl: string } | null;
  workspace: WarehouseWorkspace | null;
  language: LanguageCode;
  onRefresh?: () => void;
}

type NodeType = "warehouse" | "location" | "bin";

interface WarehouseWorkspace {
  shop: { holdingcode: string };
  shopInfo?: {
    settings?: {
      language?: string;
      languageconfigs?: unknown;
    };
  } | null;
}

interface CompanyRecord {
  guidfixed?: string;
  code?: string;
  names?: NameX[] | null;
  taxid?: string;
  isactive?: boolean;
}

interface BinFixedItem {
  guidfixed?: string;
  barcode?: string;
  unitcode?: string;
  names?: NameX[] | null;
}

interface WarehouseBinRecord {
  guidfixed?: string;
  warehouseguid?: string;
  locationguid?: string;
  code?: string;
  name?: string;
  barcode?: string;
  bintype?: string;
  aislecode?: string;
  rackcode?: string;
  levelcode?: string;
  positioncode?: string;
  widthmm?: number;
  lengthmm?: number;
  heightmm?: number;
  maxweightgram?: number;
  maxvolumecm3?: number;
  mintemperaturedeci?: number;
  maxtemperaturedeci?: number;
  allowedproductclasses?: string[];
  hazardclasses?: string[];
  fixeditems?: BinFixedItem[];
  allowputaway?: boolean;
  allowpick?: boolean;
  blockedin?: boolean;
  blockedout?: boolean;
  sortcode?: string;
  status?: string;
}

interface WarehouseLocationRecord {
  guidfixed?: string;
  warehouseguid?: string;
  code?: string;
  names?: NameX[] | null;
  locationtype?: string;
  allowedproductclasses?: string[];
  hazardclasses?: string[];
  allowputaway?: boolean;
  allowpick?: boolean;
  blockedin?: boolean;
  blockedout?: boolean;
  sortcode?: string;
  status?: string;
  companyguids?: string[];
  bins?: WarehouseBinRecord[];
}

interface WarehouseRecord {
  guidfixed?: string;
  code?: string;
  names?: NameX[] | null;
  latitude?: number;
  longitude?: number;
  companyguids?: string[];
  status?: string;
  locations?: WarehouseLocationRecord[];
}

interface SelectedNode {
  type: NodeType;
  warehouseId: string;
  locationId?: string;
  binId?: string;
}

interface ApiResponse {
  success?: boolean;
  message?: string;
  id?: string;
  data?: unknown;
}

const errorMessage = (err: unknown, fallback: string): string =>
  err instanceof Error && err.message ? err.message : fallback;

const displayName = (names: NameX[] | null | undefined, language: LanguageCode): string => {
  if (!Array.isArray(names)) return "";
  const found = names.find((item) => item.code === language);
  if (found?.name) return found.name;
  const th = names.find((item) => item.code === "th");
  if (th?.name) return th.name;
  const en = names.find((item) => item.code === "en");
  if (en?.name) return en.name;
  return names[0]?.name || "";
};

const companyDisplayName = (company: CompanyRecord, language: LanguageCode): string => {
  const val = displayName(company.names, language);
  return val || company.code || company.guidfixed || "";
};

const LOCATION_TYPE_OPTIONS = [
  { value: "storage", th: "จัดเก็บ", en: "Storage" },
  { value: "picking", th: "หยิบสินค้า", en: "Picking" },
  { value: "receiving", th: "รับสินค้า", en: "Receiving" },
  { value: "qc", th: "ตรวจสอบคุณภาพ", en: "QC" },
  { value: "wip", th: "งานระหว่างทำ", en: "WIP" },
  { value: "damaged", th: "สินค้าชำรุด", en: "Damaged" },
  { value: "transit", th: "ระหว่างขนส่ง", en: "Transit" },
] as const;

const BIN_TYPE_OPTIONS = [
  { value: "storage", th: "จัดเก็บ", en: "Storage" },
  { value: "picking", th: "หยิบสินค้า", en: "Picking" },
  { value: "staging", th: "พักสินค้า", en: "Staging" },
  { value: "wip", th: "งานระหว่างทำ", en: "WIP" },
  { value: "qc", th: "ตรวจสอบคุณภาพ", en: "QC" },
  { value: "damaged", th: "สินค้าชำรุด", en: "Damaged" },
] as const;

/** <=4 (well, small fixed) option picker per Radio Buttons vs Combo Box rule — chips, not <select>. */
function RadioChipPicker<T extends string>({
  options,
  value,
  onChange,
  language,
}: {
  options: readonly { value: T; th: string; en: string }[];
  value: T;
  onChange: (next: T) => void;
  language: LanguageCode;
}) {
  return (
    <div className="flex flex-wrap gap-1.5">
      {options.map((opt) => {
        const checked = value === opt.value;
        return (
          <label
            key={opt.value}
            className={cn(
              "flex cursor-pointer items-center gap-1.5 rounded-full border px-2.5 py-1 text-[11px] font-medium transition-colors",
              checked
                ? "border-primary/40 bg-primary/10 text-primary"
                : "border-border/50 bg-secondary/10 text-foreground/80 hover:border-primary/30"
            )}
          >
            <input
              type="radio"
              checked={checked}
              onChange={() => onChange(opt.value)}
              className="size-3.5 shrink-0 accent-primary"
            />
            {language === "th" ? opt.th : opt.en}
          </label>
        );
      })}
    </div>
  );
}

/** Compact wrap-first checkbox chip group for gate flags (allowputaway/allowpick/blockedin/blockedout). */
function BooleanChip({
  label,
  checked,
  onChange,
  tone = "default",
}: {
  label: string;
  checked: boolean;
  onChange: (next: boolean) => void;
  tone?: "default" | "danger";
}) {
  return (
    <label
      className={cn(
        "flex cursor-pointer items-center gap-1.5 rounded-full border px-2.5 py-1 text-[11px] font-medium transition-colors",
        checked
          ? tone === "danger"
            ? "border-destructive/40 bg-destructive/10 text-destructive"
            : "border-primary/40 bg-primary/10 text-primary"
          : "border-border/50 bg-secondary/10 text-foreground/80 hover:border-primary/30"
      )}
    >
      <input
        type="checkbox"
        checked={checked}
        onChange={(e) => onChange(e.target.checked)}
        className="size-3.5 shrink-0 accent-primary"
      />
      {label}
    </label>
  );
}

/**
 * Compact wrap-first checkbox chip group for picking which companies may use a warehouse/location
 * (Radio/Checkbox Compact Wrap Rule). `options` may be pre-filtered by the caller (e.g. the location
 * picker only offers companies the parent warehouse itself allows). Bins have no company scope
 * per s/warehouse.md — this picker is only used for warehouse and location forms.
 */
function CompanyScopePicker({
  options,
  selected,
  onChange,
  language,
  helperText,
  emptyText,
}: {
  options: CompanyRecord[];
  selected: string[];
  onChange: (next: string[]) => void;
  language: LanguageCode;
  helperText?: string;
  emptyText?: string;
}) {
  const toggle = (guid: string) => {
    if (selected.includes(guid)) {
      onChange(selected.filter((g) => g !== guid));
    } else {
      onChange([...selected, guid]);
    }
  };

  return (
    <div className="flex flex-col gap-1.5">
      <p className="text-[11px] text-muted-foreground leading-relaxed">
        {selected.length === 0
          ? language === "th"
            ? "ว่าง = ใช้ได้ทุกบริษัท"
            : "Empty = usable by every company"
          : language === "th"
            ? `เลือกแล้ว ${selected.length} บริษัท`
            : `${selected.length} companies selected`}
      </p>
      {helperText && <p className="text-[10px] text-muted-foreground/80 italic">{helperText}</p>}
      {options.length === 0 ? (
        <p className="text-[11px] text-muted-foreground italic">
          {emptyText || (language === "th" ? "ไม่มีบริษัทให้เลือก" : "No companies available")}
        </p>
      ) : (
        <div className="flex flex-wrap gap-1.5">
          {options.map((company) => {
            const guid = company.guidfixed || "";
            if (!guid) return null;
            const checked = selected.includes(guid);
            return (
              <label
                key={guid}
                className={cn(
                  "flex cursor-pointer items-center gap-1.5 rounded-full border px-2.5 py-1 text-[11px] font-medium transition-colors",
                  checked
                    ? "border-primary/40 bg-primary/10 text-primary"
                    : "border-border/50 bg-secondary/10 text-foreground/80 hover:border-primary/30"
                )}
              >
                <input
                  type="checkbox"
                  checked={checked}
                  onChange={() => toggle(guid)}
                  className="size-3.5 shrink-0 accent-primary"
                />
                {companyDisplayName(company, language)}
              </label>
            );
          })}
        </div>
      )}
    </div>
  );
}

// Number <-> string helpers for the physical-spec integer fields (widthmm, maxweightgram, ...).
const numStr = (v: number | undefined): string => (v !== undefined && v !== 0 ? String(v) : "");
const toInt = (v: string): number => {
  const n = parseInt(v, 10);
  return Number.isFinite(n) ? n : 0;
};

type FormType =
  | "createwarehouse"
  | "editwarehouse"
  | "createlocation"
  | "editlocation"
  | "createbin"
  | "editbin";

interface WarehouseFormFields {
  code: string;
  names: Record<string, string>;
  latitude: string;
  longitude: string;
  companyguids: string[];
}

interface LocationFormFields {
  code: string;
  names: Record<string, string>;
  locationtype: string;
  allowedproductclasses: string;
  hazardclasses: string;
  allowputaway: boolean;
  allowpick: boolean;
  blockedin: boolean;
  blockedout: boolean;
  sortcode: string;
  companyguids: string[];
}

interface BinFormFields {
  code: string;
  name: string;
  barcode: string;
  bintype: string;
  aislecode: string;
  rackcode: string;
  levelcode: string;
  positioncode: string;
  widthmm: string;
  lengthmm: string;
  heightmm: string;
  maxweightgram: string;
  maxvolumecm3: string;
  mintemperaturedeci: string;
  maxtemperaturedeci: string;
  allowedproductclasses: string;
  hazardclasses: string;
  allowputaway: boolean;
  allowpick: boolean;
  blockedin: boolean;
  blockedout: boolean;
  sortcode: string;
}

const emptyWarehouseForm = (languages: string[]): WarehouseFormFields => {
  const names: Record<string, string> = {};
  languages.forEach((l) => (names[l] = ""));
  return { code: "", names, latitude: "", longitude: "", companyguids: [] };
};

const emptyLocationForm = (languages: string[]): LocationFormFields => {
  const names: Record<string, string> = {};
  languages.forEach((l) => (names[l] = ""));
  return {
    code: "",
    names,
    locationtype: "storage",
    allowedproductclasses: "",
    hazardclasses: "",
    allowputaway: true,
    allowpick: true,
    blockedin: false,
    blockedout: false,
    sortcode: "",
    companyguids: [],
  };
};

const emptyBinForm = (): BinFormFields => ({
  code: "",
  name: "",
  barcode: "",
  bintype: "storage",
  aislecode: "",
  rackcode: "",
  levelcode: "",
  positioncode: "",
  widthmm: "",
  lengthmm: "",
  heightmm: "",
  maxweightgram: "",
  maxvolumecm3: "",
  mintemperaturedeci: "",
  maxtemperaturedeci: "",
  allowedproductclasses: "",
  hazardclasses: "",
  allowputaway: true,
  allowpick: true,
  blockedin: false,
  blockedout: false,
  sortcode: "",
});

const csvToArray = (v: string): string[] =>
  v.split(",").map((s) => s.trim()).filter(Boolean);
const arrayToCsv = (v: string[] | undefined): string => (Array.isArray(v) ? v.join(", ") : "");

export function WarehouseTreeView({ auth, workspace, language, onRefresh }: WarehouseTreeViewProps) {
  const { confirm, confirmationDialog } = useConfirmDialog();

  const editorLanguages = useMemo(() => {
    if (!workspace) return ["th"];
    const configs = workspace.shopInfo?.settings?.languageconfigs || [];
    const defaultCode = workspace.shopInfo?.settings?.language || "th";
    return normalizeLanguageConfigs(configs, defaultCode).map((row) => row.code);
  }, [workspace]);

  const mainApiUrl = useMemo(() => {
    try {
      return deriveMainApiUrl(auth?.backendUrl ?? "");
    } catch {
      return "";
    }
  }, [auth]);

  // Tree data — self-fetched from GET /warehouse/tree (new contract; parent's generic `records`
  // loader is shaped for the old embedded-location document and cannot serve this screen anymore).
  const [tree, setTree] = useState<WarehouseRecord[]>([]);
  // Starts false (not true): loadTree's early-return when auth/mainApiUrl isn't ready yet never
  // flips loading back off, so initializing true would leave a permanent spinner in that window —
  // same fix as the sibling company-branch-tree-view.tsx uses for the same reason.
  const [loading, setLoading] = useState(false);
  const [loadError, setLoadError] = useState("");
  const [companiesList, setCompaniesList] = useState<CompanyRecord[]>([]);

  const loadTree = useCallback(async () => {
    if (!auth || !mainApiUrl) return;
    setLoading(true);
    setLoadError("");
    try {
      const res = await authFetch(`${mainApiUrl}/warehouse/tree`, {
        headers: { Authorization: `Bearer ${auth.token}` },
        cache: "no-store",
      });
      const json = (await res.json()) as ApiResponse;
      if (!res.ok || json.success === false) {
        throw new Error(json.message || "โหลดโครงสร้างคลังสินค้าไม่สำเร็จ");
      }
      setTree(Array.isArray(json.data) ? (json.data as WarehouseRecord[]) : []);
    } catch (err) {
      setLoadError(errorMessage(err, "โหลดโครงสร้างคลังสินค้าไม่สำเร็จ"));
      setTree([]);
    } finally {
      setLoading(false);
    }
  }, [auth, mainApiUrl]);

  useEffect(() => {
    void loadTree();
  }, [loadTree]);

  useEffect(() => {
    if (!auth || !mainApiUrl) return;
    authFetch(`${mainApiUrl}/organization/company`, {
      headers: { Authorization: `Bearer ${auth.token}` },
    })
      .then((res) => res.json())
      .then((json) => {
        if (json.success && Array.isArray(json.data)) {
          setCompaniesList(json.data);
        }
      })
      .catch((err) => console.error(err));
  }, [auth, mainApiUrl]);

  const refreshAll = useCallback(async () => {
    await loadTree();
    if (onRefresh) onRefresh();
  }, [loadTree, onRefresh]);

  // Collapsed states
  const [collapsedWarehouses, setCollapsedWarehouses] = useState<Record<string, boolean>>({});
  const [collapsedLocs, setCollapsedLocs] = useState<Record<string, boolean>>({});

  const [selectedNode, setSelectedNode] = useState<SelectedNode | null>(null);
  const [formType, setFormType] = useState<FormType | null>("createwarehouse");

  const [warehouseForm, setWarehouseForm] = useState<WarehouseFormFields>(() => emptyWarehouseForm(["th"]));
  const [locationForm, setLocationForm] = useState<LocationFormFields>(() => emptyLocationForm(["th"]));
  const [binForm, setBinForm] = useState<BinFormFields>(emptyBinForm());

  const [formError, setFormError] = useState("");
  const [isSavingLocal, setIsSavingLocal] = useState(false);
  const [isMapOpen, setIsMapOpen] = useState(false);

  const findWarehouse = useCallback(
    (warehouseId: string) => tree.find((w) => w.guidfixed === warehouseId),
    [tree]
  );
  const findLocation = useCallback(
    (warehouseId: string, locationId: string) =>
      findWarehouse(warehouseId)?.locations?.find((l) => l.guidfixed === locationId),
    [findWarehouse]
  );
  const findBin = useCallback(
    (warehouseId: string, locationId: string, binId: string) =>
      findLocation(warehouseId, locationId)?.bins?.find((b) => b.guidfixed === binId),
    [findLocation]
  );

  // Select first warehouse once loaded.
  useEffect(() => {
    if (tree.length > 0 && !selectedNode) {
      setSelectedNode({ type: "warehouse", warehouseId: tree[0].guidfixed || "" });
      setFormType("editwarehouse");
    }
  }, [tree, selectedNode]);

  // Sync the right-pane form whenever the selection or form mode changes.
  useEffect(() => {
    setFormError("");

    if (formType === "createwarehouse") {
      setWarehouseForm(emptyWarehouseForm(editorLanguages));
      return;
    }
    if (formType === "createlocation") {
      setLocationForm(emptyLocationForm(editorLanguages));
      return;
    }
    if (formType === "createbin") {
      setBinForm(emptyBinForm());
      return;
    }
    if (!selectedNode) return;

    if (formType === "editwarehouse") {
      const w = findWarehouse(selectedNode.warehouseId);
      const names: Record<string, string> = {};
      editorLanguages.forEach((l) => {
        names[l] = w?.names?.find((n) => n.code === l)?.name || "";
      });
      setWarehouseForm({
        code: w?.code || "",
        names,
        latitude: w?.latitude ? String(w.latitude) : "",
        longitude: w?.longitude ? String(w.longitude) : "",
        companyguids: w?.companyguids || [],
      });
    } else if (formType === "editlocation" && selectedNode.locationId) {
      const loc = findLocation(selectedNode.warehouseId, selectedNode.locationId);
      const names: Record<string, string> = {};
      editorLanguages.forEach((l) => {
        names[l] = loc?.names?.find((n) => n.code === l)?.name || "";
      });
      setLocationForm({
        code: loc?.code || "",
        names,
        locationtype: loc?.locationtype || "storage",
        allowedproductclasses: arrayToCsv(loc?.allowedproductclasses),
        hazardclasses: arrayToCsv(loc?.hazardclasses),
        allowputaway: loc?.allowputaway ?? true,
        allowpick: loc?.allowpick ?? true,
        blockedin: loc?.blockedin ?? false,
        blockedout: loc?.blockedout ?? false,
        sortcode: loc?.sortcode || "",
        companyguids: loc?.companyguids || [],
      });
    } else if (formType === "editbin" && selectedNode.locationId && selectedNode.binId) {
      const bin = findBin(selectedNode.warehouseId, selectedNode.locationId, selectedNode.binId);
      setBinForm({
        code: bin?.code || "",
        name: bin?.name || "",
        barcode: bin?.barcode || "",
        bintype: bin?.bintype || "storage",
        aislecode: bin?.aislecode || "",
        rackcode: bin?.rackcode || "",
        levelcode: bin?.levelcode || "",
        positioncode: bin?.positioncode || "",
        widthmm: numStr(bin?.widthmm),
        lengthmm: numStr(bin?.lengthmm),
        heightmm: numStr(bin?.heightmm),
        maxweightgram: numStr(bin?.maxweightgram),
        maxvolumecm3: numStr(bin?.maxvolumecm3),
        mintemperaturedeci: numStr(bin?.mintemperaturedeci),
        maxtemperaturedeci: numStr(bin?.maxtemperaturedeci),
        allowedproductclasses: arrayToCsv(bin?.allowedproductclasses),
        hazardclasses: arrayToCsv(bin?.hazardclasses),
        allowputaway: bin?.allowputaway ?? true,
        allowpick: bin?.allowpick ?? true,
        blockedin: bin?.blockedin ?? false,
        blockedout: bin?.blockedout ?? false,
        sortcode: bin?.sortcode || "",
      });
    }
  }, [selectedNode, formType, editorLanguages, findWarehouse, findLocation, findBin]);

  const getLanguageName = (code: string) => {
    const found = LANGUAGES.find((item) => item.code === code);
    return found ? found.name : code.toUpperCase();
  };

  const namesToNameX = (names: Record<string, string>): NameX[] =>
    editorLanguages
      .map((code) => ({ code, name: names[code]?.trim() || "" }))
      .filter((item) => item.name !== "");

  // ---- Save handlers ----

  const handleSaveForm = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!auth || !mainApiUrl) return;

    if (formType === "createwarehouse" || formType === "editwarehouse") {
      const code = warehouseForm.code.trim();
      if (!code) {
        setFormError(language === "th" ? "กรุณากรอกรหัส" : "Code is required");
        return;
      }
      const namesArray = namesToNameX(warehouseForm.names);
      if (namesArray.length === 0) {
        setFormError(language === "th" ? "กรุณากรอกชื่อคลังอย่างน้อยหนึ่งภาษา" : "Please fill in at least one language name");
        return;
      }

      setIsSavingLocal(true);
      setFormError("");
      try {
        const isCreate = formType === "createwarehouse";
        const warehouseId = selectedNode?.warehouseId || "";
        const payload = {
          code,
          names: namesArray,
          latitude: parseFloat(warehouseForm.latitude) || 0,
          longitude: parseFloat(warehouseForm.longitude) || 0,
          companyguids: warehouseForm.companyguids,
          status: "active",
        };
        const url = isCreate ? `${mainApiUrl}/warehouse` : `${mainApiUrl}/warehouse/${warehouseId}`;
        const res = await authFetch(url, {
          method: isCreate ? "POST" : "PUT",
          headers: { "Content-Type": "application/json", Authorization: `Bearer ${auth.token}` },
          body: JSON.stringify(payload),
        });
        const json = (await res.json()) as ApiResponse;
        if (!res.ok || json.success === false) throw new Error(json.message || "Failed to save warehouse");

        await loadTree();
        const newId = isCreate ? json.id || "" : warehouseId;
        setSelectedNode({ type: "warehouse", warehouseId: newId });
        setFormType("editwarehouse");
        if (onRefresh) onRefresh();
      } catch (err) {
        setFormError(errorMessage(err, "Error saving warehouse"));
      } finally {
        setIsSavingLocal(false);
      }
    } else if (formType === "createlocation" || formType === "editlocation") {
      const code = locationForm.code.trim();
      if (!code) {
        setFormError(language === "th" ? "กรุณากรอกรหัส" : "Code is required");
        return;
      }
      const namesArray = namesToNameX(locationForm.names);
      if (namesArray.length === 0) {
        setFormError(language === "th" ? "กรุณากรอกชื่อที่เก็บสินค้าอย่างน้อยหนึ่งภาษา" : "Please fill in at least one language name");
        return;
      }
      const warehouseId = selectedNode?.warehouseId;
      if (!warehouseId) return;

      setIsSavingLocal(true);
      setFormError("");
      try {
        const isCreate = formType === "createlocation";
        const locationId = selectedNode?.locationId || "";
        const payload = {
          code,
          names: namesArray,
          locationtype: locationForm.locationtype,
          allowedproductclasses: csvToArray(locationForm.allowedproductclasses),
          hazardclasses: csvToArray(locationForm.hazardclasses),
          allowputaway: locationForm.allowputaway,
          allowpick: locationForm.allowpick,
          blockedin: locationForm.blockedin,
          blockedout: locationForm.blockedout,
          sortcode: locationForm.sortcode,
          status: "active",
          companyguids: locationForm.companyguids,
        };
        const url = isCreate
          ? `${mainApiUrl}/warehouse/${warehouseId}/location`
          : `${mainApiUrl}/warehouse/${warehouseId}/location/${locationId}`;
        const res = await authFetch(url, {
          method: isCreate ? "POST" : "PUT",
          headers: { "Content-Type": "application/json", Authorization: `Bearer ${auth.token}` },
          body: JSON.stringify(payload),
        });
        const json = (await res.json()) as ApiResponse;
        if (!res.ok || json.success === false) throw new Error(json.message || "Failed to save location");

        await loadTree();
        const newLocationId = isCreate ? json.id || "" : locationId;
        setSelectedNode({ type: "location", warehouseId, locationId: newLocationId });
        setFormType("editlocation");
        setCollapsedWarehouses((prev) => ({ ...prev, [warehouseId]: false }));
        if (onRefresh) onRefresh();
      } catch (err) {
        setFormError(errorMessage(err, "Error saving location"));
      } finally {
        setIsSavingLocal(false);
      }
    } else if (formType === "createbin" || formType === "editbin") {
      const code = binForm.code.trim();
      const name = binForm.name.trim();
      if (!code || !name) {
        setFormError(language === "th" ? "กรุณากรอกรหัสและชื่อ" : "Code and name are required");
        return;
      }
      const warehouseId = selectedNode?.warehouseId;
      const locationId = selectedNode?.locationId;
      if (!warehouseId || !locationId) return;

      setIsSavingLocal(true);
      setFormError("");
      try {
        const isCreate = formType === "createbin";
        const binId = selectedNode?.binId || "";
        const payload = {
          code,
          name,
          barcode: binForm.barcode.trim(),
          bintype: binForm.bintype,
          aislecode: binForm.aislecode.trim(),
          rackcode: binForm.rackcode.trim(),
          levelcode: binForm.levelcode.trim(),
          positioncode: binForm.positioncode.trim(),
          widthmm: toInt(binForm.widthmm),
          lengthmm: toInt(binForm.lengthmm),
          heightmm: toInt(binForm.heightmm),
          maxweightgram: toInt(binForm.maxweightgram),
          maxvolumecm3: toInt(binForm.maxvolumecm3),
          mintemperaturedeci: toInt(binForm.mintemperaturedeci),
          maxtemperaturedeci: toInt(binForm.maxtemperaturedeci),
          allowedproductclasses: csvToArray(binForm.allowedproductclasses),
          hazardclasses: csvToArray(binForm.hazardclasses),
          allowputaway: binForm.allowputaway,
          allowpick: binForm.allowpick,
          blockedin: binForm.blockedin,
          blockedout: binForm.blockedout,
          sortcode: binForm.sortcode,
          status: "active",
        };
        const url = isCreate
          ? `${mainApiUrl}/warehouse/${warehouseId}/location/${locationId}/bin`
          : `${mainApiUrl}/warehouse/${warehouseId}/location/${locationId}/bin/${binId}`;
        const res = await authFetch(url, {
          method: isCreate ? "POST" : "PUT",
          headers: { "Content-Type": "application/json", Authorization: `Bearer ${auth.token}` },
          body: JSON.stringify(payload),
        });
        const json = (await res.json()) as ApiResponse;
        if (!res.ok || json.success === false) throw new Error(json.message || "Failed to save bin");

        await loadTree();
        const newBinId = isCreate ? json.id || "" : binId;
        setSelectedNode({ type: "bin", warehouseId, locationId, binId: newBinId });
        setFormType("editbin");
        setCollapsedLocs((prev) => ({ ...prev, [`${warehouseId}-${locationId}`]: false }));
        if (onRefresh) onRefresh();
      } catch (err) {
        setFormError(errorMessage(err, "Error saving bin"));
      } finally {
        setIsSavingLocal(false);
      }
    }
  };

  const handleCancelForm = () => {
    const targetId = selectedNode?.warehouseId || tree[0]?.guidfixed || "";
    if (targetId) {
      setSelectedNode({ type: "warehouse", warehouseId: targetId });
      setFormType("editwarehouse");
    } else {
      setSelectedNode(null);
      setFormType("createwarehouse");
    }
  };

  // ---- Delete handlers (custom confirm dialog — never window.confirm) ----

  const handleDeleteWarehouse = async (warehouseId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (!auth || !mainApiUrl) return;
    const w = findWarehouse(warehouseId);
    if (!w) return;
    const confirmed = await confirm({
      title: language === "th" ? "ยืนยันการลบคลังสินค้า" : "Confirm Delete Warehouse",
      description:
        language === "th"
          ? `ต้องการลบคลังสินค้า "${w.code} - ${displayName(w.names, language)}" ใช่หรือไม่?`
          : `Delete warehouse "${w.code} - ${displayName(w.names, language)}"?`,
      confirmLabel: language === "th" ? "ลบ" : "Delete",
      tone: "danger",
    });
    if (!confirmed) return;

    setIsSavingLocal(true);
    setFormError("");
    try {
      const res = await authFetch(`${mainApiUrl}/warehouse/${encodeURIComponent(warehouseId)}`, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${auth.token}` },
      });
      const json = (await res.json()) as ApiResponse;
      if (!res.ok || json.success === false) throw new Error(json.message || "Failed to delete warehouse");

      await loadTree();
      setSelectedNode(null);
      setFormType(tree.length > 1 ? "editwarehouse" : "createwarehouse");
      if (onRefresh) onRefresh();
    } catch (err) {
      setFormError(errorMessage(err, "Error deleting warehouse"));
    } finally {
      setIsSavingLocal(false);
    }
  };

  const handleDeleteLocation = async (warehouseId: string, locationId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (!auth || !mainApiUrl) return;
    const loc = findLocation(warehouseId, locationId);
    if (!loc) return;
    const confirmed = await confirm({
      title: language === "th" ? "ยืนยันการลบที่เก็บสินค้า" : "Confirm Delete Location",
      description:
        language === "th"
          ? `ต้องการลบที่เก็บสินค้า "${loc.code} - ${displayName(loc.names, language)}" ใช่หรือไม่?`
          : `Delete location "${loc.code} - ${displayName(loc.names, language)}"?`,
      confirmLabel: language === "th" ? "ลบ" : "Delete",
      tone: "danger",
    });
    if (!confirmed) return;

    setIsSavingLocal(true);
    setFormError("");
    try {
      const res = await authFetch(
        `${mainApiUrl}/warehouse/${warehouseId}/location/${encodeURIComponent(locationId)}`,
        { method: "DELETE", headers: { Authorization: `Bearer ${auth.token}` } }
      );
      const json = (await res.json()) as ApiResponse;
      if (!res.ok || json.success === false) throw new Error(json.message || "Failed to delete location");

      await loadTree();
      setSelectedNode({ type: "warehouse", warehouseId });
      setFormType("editwarehouse");
      if (onRefresh) onRefresh();
    } catch (err) {
      setFormError(errorMessage(err, "Error deleting location"));
    } finally {
      setIsSavingLocal(false);
    }
  };

  const handleDeleteBin = async (warehouseId: string, locationId: string, binId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (!auth || !mainApiUrl) return;
    const bin = findBin(warehouseId, locationId, binId);
    if (!bin) return;
    const confirmed = await confirm({
      title: language === "th" ? "ยืนยันการลบที่วางสินค้า" : "Confirm Delete Bin",
      description:
        language === "th"
          ? `ต้องการลบที่วางสินค้า "${bin.code} - ${bin.name}" ใช่หรือไม่?`
          : `Delete bin "${bin.code} - ${bin.name}"?`,
      confirmLabel: language === "th" ? "ลบ" : "Delete",
      tone: "danger",
    });
    if (!confirmed) return;

    setIsSavingLocal(true);
    setFormError("");
    try {
      const res = await authFetch(
        `${mainApiUrl}/warehouse/${warehouseId}/location/${locationId}/bin/${encodeURIComponent(binId)}`,
        { method: "DELETE", headers: { Authorization: `Bearer ${auth.token}` } }
      );
      const json = (await res.json()) as ApiResponse;
      if (!res.ok || json.success === false) throw new Error(json.message || "Failed to delete bin");

      await loadTree();
      setSelectedNode({ type: "location", warehouseId, locationId });
      setFormType("editlocation");
      if (onRefresh) onRefresh();
    } catch (err) {
      setFormError(errorMessage(err, "Error deleting bin"));
    } finally {
      setIsSavingLocal(false);
    }
  };

  // Search filter
  const [searchQuery, setSearchQuery] = useState("");
  const filteredWarehouses = useMemo(() => {
    if (!searchQuery.trim()) return tree;
    const needle = searchQuery.toLowerCase();
    return tree.filter((w) => {
      const codeMatch = w.code?.toLowerCase().includes(needle);
      const nameMatch = displayName(w.names, language).toLowerCase().includes(needle);
      const locMatch = (w.locations || []).some((loc) => {
        const locCodeMatch = loc.code?.toLowerCase().includes(needle);
        const locNameMatch = displayName(loc.names, language).toLowerCase().includes(needle);
        const binMatch = (loc.bins || []).some(
          (b) => b.code?.toLowerCase().includes(needle) || b.name?.toLowerCase().includes(needle)
        );
        return locCodeMatch || locNameMatch || binMatch;
      });
      return codeMatch || nameMatch || locMatch;
    });
  }, [tree, searchQuery, language]);

  return (
    <div className="grid w-full min-w-0 items-stretch gap-3 min-h-[calc(100dvh-12rem)] xl:grid-cols-[minmax(320px,0.95fr)_minmax(420px,1.05fr)]">
      {/* LEFT COLUMN: Warehouse Tree list */}
      <Card className="flex h-full min-h-0 flex-col overflow-hidden border-border bg-card shadow-sm">
        <div className="flex items-center gap-2 border-b border-border/40 p-2.5 bg-secondary/5">
          <div className="relative flex-1">
            <Input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder={
                language === "th"
                  ? "ค้นหาคลังสินค้า ที่เก็บสินค้า หรือที่วางสินค้า..."
                  : "Search warehouse, location, or bin..."
              }
              className="h-8 !pl-10 pr-3 text-xs rounded-lg"
            />
            <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground/60">
              <svg className="size-3.5" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </span>
          </div>
          <Button
            type="button"
            size="sm"
            className="h-8 shrink-0 rounded-lg gap-1.5 bg-primary text-primary-foreground hover:bg-primary/95"
            onClick={() => {
              setFormType("createwarehouse");
              setSelectedNode(null);
            }}
            disabled={loading || isSavingLocal}
          >
            <Plus className="size-4" />
            {language === "th" ? "เพิ่มคลังสินค้า" : "Add Warehouse"}
          </Button>
        </div>

        <CardContent className="flex min-h-0 flex-1 flex-col p-0 overflow-y-auto overscroll-contain">
          {loading ? (
            <div className="flex h-full min-h-24 items-center justify-center gap-2 p-6 text-sm text-muted-foreground">
              <Loader2 className="animate-spin size-5 text-primary" />
              {language === "th" ? "กำลังโหลดข้อมูล..." : "Loading..."}
            </div>
          ) : loadError ? (
            <div className="flex h-full min-h-24 flex-col items-center justify-center gap-2 p-4 text-center text-sm">
              <AlertTriangle className="size-6 text-destructive" />
              <span className="font-medium text-destructive">{loadError}</span>
              <Button type="button" size="sm" variant="outline" onClick={() => void loadTree()}>
                {language === "th" ? "ลองใหม่" : "Retry"}
              </Button>
            </div>
          ) : filteredWarehouses.length === 0 ? (
            <div className="flex h-full min-h-24 flex-col items-center justify-center gap-1.5 p-4 text-center text-sm text-muted-foreground">
              <span className="font-medium">
                {language === "th" ? "ไม่พบข้อมูลคลังสินค้า" : "No warehouses found"}
              </span>
            </div>
          ) : (
            <div className="flex min-h-0 flex-1 flex-col overflow-y-auto overscroll-contain">
              {filteredWarehouses.map((w, warehouseIdx) => {
                const warehouseId = w.guidfixed || "";
                const isWarehouseSelected =
                  selectedNode?.type === "warehouse" && selectedNode?.warehouseId === warehouseId;
                const isWarehouseCollapsed = collapsedWarehouses[warehouseId] ?? false;
                const locations = w.locations || [];
                const level0Style = warehouseLevelStyle(0);

                return (
                  <div key={warehouseId} className="flex flex-col border-b border-border/30 last:border-b-0">
                    {/* Level 0: Warehouse Node */}
                    <div
                      className={cn(
                        "group/row relative flex items-center justify-between border-b border-border/40 py-2 px-3 transition-colors cursor-pointer",
                        isWarehouseSelected ? level0Style.selectedBg : "hover:bg-muted/40"
                      )}
                      onClick={() => {
                        setSelectedNode({ type: "warehouse", warehouseId });
                        setFormType("editwarehouse");
                      }}
                    >
                      {isWarehouseSelected && (
                        <span className={cn("pointer-events-none absolute bottom-0 left-0 top-0 w-1", level0Style.borderLeft)} />
                      )}
                      <div className="flex min-w-0 flex-1 flex-wrap items-center gap-x-2 gap-y-1 pr-2">
                        <GripVertical className={cn("size-4 shrink-0 transition-colors", level0Style.grip)} />

                        <button
                          type="button"
                          className={cn(
                            "size-6 flex items-center justify-center rounded hover:bg-muted text-muted-foreground shrink-0",
                            locations.length === 0 && "invisible"
                          )}
                          onClick={(e) => {
                            e.stopPropagation();
                            setCollapsedWarehouses((prev) => ({ ...prev, [warehouseId]: !isWarehouseCollapsed }));
                          }}
                        >
                          {isWarehouseCollapsed ? (
                            <ChevronRight className={cn("size-4", level0Style.caret)} />
                          ) : (
                            <ChevronDown className={cn("size-4", level0Style.caret)} />
                          )}
                        </button>

                        <span className={cn("min-w-[1.25rem] text-sm font-bold", level0Style.order)}>
                          {warehouseIdx + 1}
                        </span>

                        <span className={cn("min-w-0 flex-1 basis-40 whitespace-normal break-words text-sm", level0Style.name)}>
                          {w.code} - {displayName(w.names, language)}
                        </span>

                        {locations.length > 0 ? (
                          <button
                            type="button"
                            className="shrink-0 rounded-full border border-primary/20 bg-primary/10 px-2 py-0.5 text-[11px] font-semibold leading-5 text-primary transition-colors hover:border-primary/40 hover:bg-primary/15"
                            onClick={(e) => {
                              e.stopPropagation();
                              setCollapsedWarehouses((prev) => ({ ...prev, [warehouseId]: !isWarehouseCollapsed }));
                            }}
                          >
                            {language === "th" ? `ลูก ${locations.length}` : `${locations.length} ${locations.length === 1 ? "child" : "children"}`}
                          </button>
                        ) : null}
                      </div>

                      <div className="flex shrink-0 items-center gap-1 opacity-60 group-hover/row:opacity-100 transition-opacity">
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          className="size-7 rounded-full text-emerald-600 hover:text-emerald-700 hover:bg-emerald-50 dark:hover:bg-emerald-950/40"
                          title={language === "th" ? "เพิ่มที่เก็บสินค้า" : "Add Location"}
                          onClick={(e) => {
                            e.stopPropagation();
                            setSelectedNode({ type: "warehouse", warehouseId });
                            setFormType("createlocation");
                            setCollapsedWarehouses((prev) => ({ ...prev, [warehouseId]: false }));
                          }}
                        >
                          <Boxes className="size-3.5" />
                        </Button>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          className="size-7 rounded-full text-blue-600 hover:text-blue-700 hover:bg-blue-50 dark:hover:bg-blue-950/40"
                          title={language === "th" ? "แก้ไขคลังสินค้า" : "Edit Warehouse"}
                          onClick={(e) => {
                            e.stopPropagation();
                            setSelectedNode({ type: "warehouse", warehouseId });
                            setFormType("editwarehouse");
                          }}
                        >
                          <Edit3 className="size-3.5" />
                        </Button>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          className="size-7 rounded-full text-destructive hover:bg-destructive/10"
                          title={language === "th" ? "ลบคลังสินค้า" : "Delete Warehouse"}
                          onClick={(e) => handleDeleteWarehouse(warehouseId, e)}
                        >
                          <Trash2 className="size-3.5" />
                        </Button>
                      </div>
                    </div>

                    {/* Level 1: Locations List */}
                    {!isWarehouseCollapsed && locations.length > 0 && (
                      <div className="flex flex-col bg-secondary/5">
                        {locations.map((loc, locIdx) => {
                          const locationId = loc.guidfixed || "";
                          const isLocSelected =
                            selectedNode?.type === "location" &&
                            selectedNode?.warehouseId === warehouseId &&
                            selectedNode?.locationId === locationId;
                          const locKey = `${warehouseId}-${locationId}`;
                          const isLocCollapsed = collapsedLocs[locKey] ?? false;
                          const bins = loc.bins || [];
                          const level1Style = warehouseLevelStyle(1);

                          return (
                            <div key={locationId} className="flex flex-col">
                              <div
                                className={cn(
                                  "group/row relative flex items-center justify-between border-b border-border/40 py-2 px-3 transition-colors cursor-pointer",
                                  isLocSelected ? level1Style.selectedBg : "hover:bg-muted/40"
                                )}
                                style={{ paddingLeft: "36px" }}
                                onClick={() => {
                                  setSelectedNode({ type: "location", warehouseId, locationId });
                                  setFormType("editlocation");
                                }}
                              >
                                {isLocSelected && (
                                  <span className={cn("pointer-events-none absolute bottom-0 left-0 top-0 w-1", level1Style.borderLeft)} />
                                )}
                                <div className="flex min-w-0 flex-1 flex-wrap items-center gap-x-2 gap-y-1 pr-2">
                                  <GripVertical className={cn("size-4 shrink-0 transition-colors", level1Style.grip)} />

                                  <button
                                    type="button"
                                    className={cn(
                                      "size-6 flex items-center justify-center rounded hover:bg-muted text-muted-foreground shrink-0",
                                      bins.length === 0 && "invisible"
                                    )}
                                    onClick={(e) => {
                                      e.stopPropagation();
                                      setCollapsedLocs((prev) => ({ ...prev, [locKey]: !isLocCollapsed }));
                                    }}
                                  >
                                    {isLocCollapsed ? (
                                      <ChevronRight className={cn("size-4", level1Style.caret)} />
                                    ) : (
                                      <ChevronDown className={cn("size-4", level1Style.caret)} />
                                    )}
                                  </button>

                                  <span className={cn("min-w-[1.25rem] text-sm font-bold", level1Style.order)}>
                                    {locIdx + 1}
                                  </span>

                                  <span className={cn("min-w-0 flex-1 basis-40 whitespace-normal break-words text-sm", level1Style.name)}>
                                    {loc.code} - {displayName(loc.names, language)}
                                  </span>

                                  {bins.length > 0 ? (
                                    <button
                                      type="button"
                                      className="shrink-0 rounded-full border border-sky-200 bg-sky-100 dark:border-sky-800 dark:bg-sky-950 px-2 py-0.5 text-[11px] font-semibold leading-5 text-sky-700 dark:text-sky-300 transition-colors hover:bg-sky-200 dark:hover:bg-sky-900"
                                      onClick={(e) => {
                                        e.stopPropagation();
                                        setCollapsedLocs((prev) => ({ ...prev, [locKey]: !isLocCollapsed }));
                                      }}
                                    >
                                      {language === "th" ? `ลูก ${bins.length}` : `${bins.length} ${bins.length === 1 ? "child" : "children"}`}
                                    </button>
                                  ) : null}
                                </div>

                                <div className="flex shrink-0 items-center gap-1 opacity-60 group-hover/row:opacity-100 transition-opacity">
                                  <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    className="size-7 rounded-full text-emerald-600 hover:text-emerald-700 hover:bg-emerald-50 dark:hover:bg-emerald-950/40"
                                    title={language === "th" ? "เพิ่มที่วางสินค้า" : "Add Bin"}
                                    onClick={(e) => {
                                      e.stopPropagation();
                                      setSelectedNode({ type: "location", warehouseId, locationId });
                                      setFormType("createbin");
                                      setCollapsedLocs((prev) => ({ ...prev, [locKey]: false }));
                                    }}
                                  >
                                    <Plus className="size-3.5" />
                                  </Button>
                                  <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    className="size-7 rounded-full text-blue-600 hover:text-blue-700 hover:bg-blue-50 dark:hover:bg-blue-950/40"
                                    title={language === "th" ? "แก้ไข" : "Edit Location"}
                                    onClick={(e) => {
                                      e.stopPropagation();
                                      setSelectedNode({ type: "location", warehouseId, locationId });
                                      setFormType("editlocation");
                                    }}
                                  >
                                    <Edit3 className="size-3.5" />
                                  </Button>
                                  <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    className="size-7 rounded-full text-destructive hover:bg-destructive/10"
                                    title={language === "th" ? "ลบที่เก็บสินค้า" : "Delete Location"}
                                    onClick={(e) => handleDeleteLocation(warehouseId, locationId, e)}
                                  >
                                    <Trash2 className="size-3.5" />
                                  </Button>
                                </div>
                              </div>

                              {/* Level 2: Bins List */}
                              {!isLocCollapsed && bins.length > 0 && (
                                <div className="flex flex-col bg-secondary/10">
                                  {bins.map((bin, binIdx) => {
                                    const binId = bin.guidfixed || "";
                                    const isBinSelected =
                                      selectedNode?.type === "bin" &&
                                      selectedNode?.warehouseId === warehouseId &&
                                      selectedNode?.locationId === locationId &&
                                      selectedNode?.binId === binId;
                                    const level2Style = warehouseLevelStyle(2);

                                    return (
                                      <div
                                        key={binId}
                                        className={cn(
                                          "group/row relative flex items-center justify-between border-b border-border/40 py-1.5 px-3 transition-colors cursor-pointer",
                                          isBinSelected ? level2Style.selectedBg : "hover:bg-muted/40"
                                        )}
                                        style={{ paddingLeft: "60px" }}
                                        onClick={() => {
                                          setSelectedNode({ type: "bin", warehouseId, locationId, binId });
                                          setFormType("editbin");
                                        }}
                                      >
                                        {isBinSelected && (
                                          <span className={cn("pointer-events-none absolute bottom-0 left-0 top-0 w-1", level2Style.borderLeft)} />
                                        )}
                                        <div className="flex min-w-0 flex-1 flex-wrap items-center gap-x-2 gap-y-1 pr-2">
                                          <GripVertical className={cn("size-4 shrink-0 transition-colors", level2Style.grip)} />
                                          <div className="size-6 shrink-0" />
                                          <span className={cn("min-w-[1.25rem] text-sm font-bold", level2Style.order)}>
                                            {binIdx + 1}
                                          </span>
                                          <span className={cn("min-w-0 flex-1 basis-40 whitespace-normal break-words text-sm", level2Style.name)}>
                                            {bin.code} - {bin.name}
                                          </span>
                                        </div>

                                        <div className="flex shrink-0 items-center gap-1 opacity-60 group-hover/row:opacity-100 transition-opacity">
                                          <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon"
                                            className="size-7 rounded-full text-blue-600 hover:text-blue-700 hover:bg-blue-50 dark:hover:bg-blue-950/40"
                                            title={language === "th" ? "แก้ไข" : "Edit Bin"}
                                            onClick={(e) => {
                                              e.stopPropagation();
                                              setSelectedNode({ type: "bin", warehouseId, locationId, binId });
                                              setFormType("editbin");
                                            }}
                                          >
                                            <Edit3 className="size-3.5" />
                                          </Button>
                                          <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon"
                                            className="size-7 rounded-full text-destructive hover:bg-destructive/10"
                                            title={language === "th" ? "ลบที่วางสินค้า" : "Delete Bin"}
                                            onClick={(e) => handleDeleteBin(warehouseId, locationId, binId, e)}
                                          >
                                            <Trash2 className="size-3.5" />
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
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>

      {/* RIGHT COLUMN: Edit Node Inline Form Card */}
      <Card className="flex min-h-[360px] flex-col overflow-hidden border-border bg-card shadow-sm xl:h-full xl:min-h-0">
        <div className="flex min-h-10 items-center justify-between border-b border-border/40 bg-secondary/5 px-3 py-2">
          <h3 className="text-xs font-bold text-foreground uppercase tracking-wider flex items-center gap-1.5">
            {formType === "createwarehouse" && (<><Plus className="size-4 text-emerald-600 shrink-0" />{language === "th" ? "เพิ่มคลังสินค้า" : "Add Warehouse"}</>)}
            {formType === "editwarehouse" && (<><Warehouse className="size-4 text-primary shrink-0" />{language === "th" ? "แก้ไขคลังสินค้า" : "Edit Warehouse"}</>)}
            {formType === "createlocation" && (<><Plus className="size-4 text-emerald-600 shrink-0" />{language === "th" ? "เพิ่มที่เก็บสินค้า" : "Add Location"}</>)}
            {formType === "editlocation" && (<><Edit3 className="size-4 text-blue-600 shrink-0" />{language === "th" ? "แก้ไขที่เก็บสินค้า" : "Edit Location"}</>)}
            {formType === "createbin" && (<><Plus className="size-4 text-emerald-600 shrink-0" />{language === "th" ? "เพิ่มที่วางสินค้า" : "Add Bin"}</>)}
            {formType === "editbin" && (<><PackageSearch className="size-4 text-blue-600 shrink-0" />{language === "th" ? "แก้ไขที่วางสินค้า" : "Edit Bin"}</>)}
          </h3>
          {formType ? (
            <div className="flex w-full shrink-0 flex-wrap items-center justify-end gap-2 sm:w-auto" data-testid="warehouse-form-actions">
              <Button
                type="button"
                variant="outline"
                size="sm"
                onClick={handleCancelForm}
                disabled={isSavingLocal}
                className="h-8 text-xs font-semibold"
              >
                {language === "th" ? "ยกเลิก" : "Cancel"}
              </Button>
              <Button
                type="submit"
                form="warehouse-node-form"
                size="sm"
                disabled={isSavingLocal}
                className="h-8 gap-1 bg-primary text-xs font-semibold text-primary-foreground hover:bg-primary/95"
              >
                {isSavingLocal ? <Loader2 className="size-3 animate-spin" /> : <Save className="size-3.5" />}
                {language === "th" ? "บันทึก" : "Save"}
              </Button>
            </div>
          ) : null}
        </div>

        <CardContent className="min-h-0 flex-1 overflow-y-auto overscroll-contain p-3">
          {!formType ? (
            <div className="flex flex-col items-center justify-center h-48 text-center text-xs text-muted-foreground italic gap-2">
              <MapPin className="size-8 text-muted-foreground/30 animate-pulse" />
              {language === "th"
                ? "เลือกคลังสินค้า ที่เก็บสินค้า หรือที่วางสินค้า ในโครงสร้างด้านซ้ายเพื่อทำการแก้ไข"
                : "Select warehouse, location, or bin on the left to edit."}
            </div>
          ) : (
            <form id="warehouse-node-form" onSubmit={handleSaveForm} className="flex flex-col gap-3">
              {/* WAREHOUSE form */}
              {(formType === "createwarehouse" || formType === "editwarehouse") && (
                <div className="flex flex-col gap-3">
                  <div className="flex flex-col gap-1">
                    <label className="text-[11px] font-bold text-muted-foreground uppercase">
                      {language === "th" ? "รหัสคลังสินค้า" : "Warehouse Code"} <span className="text-destructive">*</span>
                    </label>
                    <Input
                      value={warehouseForm.code}
                      onChange={(e) => setWarehouseForm((prev) => ({ ...prev, code: e.target.value }))}
                      placeholder="e.g. 00000"
                      className="h-9 text-xs"
                    />
                  </div>

                  <div className="flex flex-col gap-1.5">
                    <div className="grid grid-cols-2 gap-3">
                      <div className="flex flex-col gap-1">
                        <label className="text-[11px] font-bold text-muted-foreground uppercase">
                          {language === "th" ? "ละติจูด (Latitude)" : "Latitude"}
                        </label>
                        <Input
                          type="number"
                          step="any"
                          value={warehouseForm.latitude}
                          onChange={(e) => setWarehouseForm((prev) => ({ ...prev, latitude: e.target.value }))}
                          placeholder="e.g. 13.7563"
                          className="h-9 text-xs"
                        />
                      </div>
                      <div className="flex flex-col gap-1">
                        <label className="text-[11px] font-bold text-muted-foreground uppercase">
                          {language === "th" ? "ลองจิจูด (Longitude)" : "Longitude"}
                        </label>
                        <Input
                          type="number"
                          step="any"
                          value={warehouseForm.longitude}
                          onChange={(e) => setWarehouseForm((prev) => ({ ...prev, longitude: e.target.value }))}
                          placeholder="e.g. 100.5018"
                          className="h-9 text-xs"
                        />
                      </div>
                    </div>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={() => setIsMapOpen(true)}
                      className="mt-1 h-8 gap-1 border-primary/30 text-primary hover:bg-primary/5 hover:border-primary/50 transition-all text-xs font-semibold"
                    >
                      <MapPin className="size-3.5" />
                      {language === "th" ? "เลือกตำแหน่งจากแผนที่" : "Pick from Map"}
                    </Button>
                  </div>

                  <div className="border-t border-border/30 pt-3">
                    <NamesEditor
                      names={editorLanguages.map((l) => ({ code: l, name: warehouseForm.names[l] || "" }))}
                      onChange={(next) =>
                        setWarehouseForm((prev) => ({
                          ...prev,
                          names: next.reduce((acc, n) => ({ ...acc, [n.code || ""]: n.name || "" }), {} as Record<string, string>),
                        }))
                      }
                      languages={editorLanguages}
                      label={language === "th" ? "ชื่อคลังสินค้าหลายภาษา" : "Multilingual Warehouse Names"}
                      firstRequired
                      language={language}
                    />
                  </div>

                  <div className="border-t border-border/30 pt-3 flex flex-col gap-1.5">
                    <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider block">
                      {language === "th" ? "บริษัทที่ใช้คลังนี้ได้" : "Companies allowed to use this warehouse"}
                    </span>
                    <CompanyScopePicker
                      options={companiesList}
                      selected={warehouseForm.companyguids}
                      onChange={(next) => setWarehouseForm((prev) => ({ ...prev, companyguids: next }))}
                      language={language}
                    />
                  </div>
                </div>
              )}

              {/* LOCATION form */}
              {(formType === "createlocation" || formType === "editlocation") && (
                <div className="flex flex-col gap-3">
                  <div className="flex flex-col gap-1">
                    <label className="text-[11px] font-bold text-muted-foreground uppercase">
                      {language === "th" ? "รหัสที่เก็บสินค้า" : "Location Code"} <span className="text-destructive">*</span>
                    </label>
                    <Input
                      value={locationForm.code}
                      onChange={(e) => setLocationForm((prev) => ({ ...prev, code: e.target.value }))}
                      placeholder="e.g. ZONE-A"
                      className="h-9 text-xs"
                    />
                  </div>

                  <div className="flex flex-col gap-1.5">
                    <label className="text-[11px] font-bold text-muted-foreground uppercase">
                      {language === "th" ? "ประเภทที่เก็บสินค้า" : "Location Type"}
                    </label>
                    <RadioChipPicker
                      options={LOCATION_TYPE_OPTIONS}
                      value={locationForm.locationtype}
                      onChange={(v) => setLocationForm((prev) => ({ ...prev, locationtype: v }))}
                      language={language}
                    />
                  </div>

                  <div className="border-t border-border/30 pt-3">
                    <NamesEditor
                      names={editorLanguages.map((l) => ({ code: l, name: locationForm.names[l] || "" }))}
                      onChange={(next) =>
                        setLocationForm((prev) => ({
                          ...prev,
                          names: next.reduce((acc, n) => ({ ...acc, [n.code || ""]: n.name || "" }), {} as Record<string, string>),
                        }))
                      }
                      languages={editorLanguages}
                      label={language === "th" ? "ชื่อที่เก็บสินค้าหลายภาษา" : "Multilingual Location Names"}
                      firstRequired
                      language={language}
                    />
                  </div>

                  <div className="grid grid-cols-2 gap-3">
                    <div className="flex flex-col gap-1">
                      <label className="text-[11px] font-bold text-muted-foreground uppercase">
                        {language === "th" ? "ประเภทสินค้าที่อนุญาต (คั่นด้วย ,)" : "Allowed Product Classes (comma-separated)"}
                      </label>
                      <Input
                        value={locationForm.allowedproductclasses}
                        onChange={(e) => setLocationForm((prev) => ({ ...prev, allowedproductclasses: e.target.value }))}
                        placeholder={language === "th" ? "เช่น ของแช่แข็ง, ของเหลว" : "e.g. Frozen, Liquid"}
                        className="h-9 text-xs"
                      />
                    </div>
                    <div className="flex flex-col gap-1">
                      <label className="text-[11px] font-bold text-muted-foreground uppercase">
                        {language === "th" ? "ประเภทวัตถุอันตราย (คั่นด้วย ,)" : "Hazard Classes (comma-separated)"}
                      </label>
                      <Input
                        value={locationForm.hazardclasses}
                        onChange={(e) => setLocationForm((prev) => ({ ...prev, hazardclasses: e.target.value }))}
                        placeholder="e.g. flammable"
                        className="h-9 text-xs"
                      />
                    </div>
                  </div>

                  <div className="flex flex-col gap-1">
                    <label className="text-[11px] font-bold text-muted-foreground uppercase">
                      {language === "th" ? "ลำดับการจัดเรียง" : "Sort Code"}
                    </label>
                    <Input
                      value={locationForm.sortcode}
                      onChange={(e) => setLocationForm((prev) => ({ ...prev, sortcode: e.target.value }))}
                      className="h-9 text-xs"
                    />
                  </div>

                  <div className="flex flex-col gap-1.5">
                    <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider block">
                      {language === "th" ? "กฎการเข้า-ออก" : "In/Out Rules"}
                    </span>
                    <div className="flex flex-wrap gap-1.5">
                      <BooleanChip
                        label={language === "th" ? "รับเข้าได้ (Putaway)" : "Allow Putaway"}
                        checked={locationForm.allowputaway}
                        onChange={(v) => setLocationForm((prev) => ({ ...prev, allowputaway: v }))}
                      />
                      <BooleanChip
                        label={language === "th" ? "หยิบออกได้ (Pick)" : "Allow Pick"}
                        checked={locationForm.allowpick}
                        onChange={(v) => setLocationForm((prev) => ({ ...prev, allowpick: v }))}
                      />
                      <BooleanChip
                        label={language === "th" ? "บล็อกรับเข้า" : "Blocked In"}
                        checked={locationForm.blockedin}
                        onChange={(v) => setLocationForm((prev) => ({ ...prev, blockedin: v }))}
                        tone="danger"
                      />
                      <BooleanChip
                        label={language === "th" ? "บล็อกส่งออก" : "Blocked Out"}
                        checked={locationForm.blockedout}
                        onChange={(v) => setLocationForm((prev) => ({ ...prev, blockedout: v }))}
                        tone="danger"
                      />
                    </div>
                  </div>

                  <div className="border-t border-border/30 pt-3 flex flex-col gap-1.5">
                    <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider block">
                      {language === "th" ? "บริษัทที่ใช้ที่เก็บสินค้านี้ได้" : "Companies allowed to use this location"}
                    </span>
                    {(() => {
                      const parentWarehouse = findWarehouse(selectedNode?.warehouseId || "");
                      const warehouseGuids = parentWarehouse?.companyguids || [];
                      const restricted = warehouseGuids.length > 0;
                      const pickerOptions = restricted
                        ? companiesList.filter((c) => c.guidfixed && warehouseGuids.includes(c.guidfixed))
                        : companiesList;
                      return (
                        <CompanyScopePicker
                          options={pickerOptions}
                          selected={locationForm.companyguids}
                          onChange={(next) => setLocationForm((prev) => ({ ...prev, companyguids: next }))}
                          language={language}
                          helperText={
                            restricted
                              ? language === "th"
                                ? "แสดงเฉพาะบริษัทที่คลังนี้อนุญาต"
                                : "Showing only companies this warehouse allows"
                              : undefined
                          }
                          emptyText={
                            restricted
                              ? language === "th"
                                ? "คลังนี้ยังไม่ได้อนุญาตบริษัทใด"
                                : "This warehouse has no allowed companies"
                              : undefined
                          }
                        />
                      );
                    })()}
                  </div>
                </div>
              )}

              {/* BIN form */}
              {(formType === "createbin" || formType === "editbin") && (
                <div className="flex flex-col gap-3">
                  <div className="grid grid-cols-2 gap-3">
                    <div className="flex flex-col gap-1">
                      <label className="text-[11px] font-bold text-muted-foreground uppercase">
                        {language === "th" ? "รหัสที่วางสินค้า" : "Bin Code"} <span className="text-destructive">*</span>
                      </label>
                      <Input
                        value={binForm.code}
                        onChange={(e) => setBinForm((prev) => ({ ...prev, code: e.target.value }))}
                        placeholder="e.g. BIN-01"
                        className="h-9 text-xs"
                      />
                    </div>
                    <div className="flex flex-col gap-1">
                      <label className="text-[11px] font-bold text-muted-foreground uppercase">
                        {language === "th" ? "ชื่อที่วางสินค้า" : "Bin Name"} <span className="text-destructive">*</span>
                      </label>
                      <Input
                        value={binForm.name}
                        onChange={(e) => setBinForm((prev) => ({ ...prev, name: e.target.value }))}
                        placeholder="e.g. Row A, Tier 1"
                        className="h-9 text-xs"
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-3">
                    <div className="flex flex-col gap-1">
                      <label className="text-[11px] font-bold text-muted-foreground uppercase">
                        {language === "th" ? "บาร์โค้ด" : "Barcode"}
                      </label>
                      <Input
                        value={binForm.barcode}
                        onChange={(e) => setBinForm((prev) => ({ ...prev, barcode: e.target.value }))}
                        className="h-9 text-xs"
                      />
                    </div>
                    <div className="flex flex-col gap-1">
                      <label className="text-[11px] font-bold text-muted-foreground uppercase">
                        {language === "th" ? "ลำดับการจัดเรียง" : "Sort Code"}
                      </label>
                      <Input
                        value={binForm.sortcode}
                        onChange={(e) => setBinForm((prev) => ({ ...prev, sortcode: e.target.value }))}
                        className="h-9 text-xs"
                      />
                    </div>
                  </div>

                  <div className="flex flex-col gap-1.5">
                    <label className="text-[11px] font-bold text-muted-foreground uppercase">
                      {language === "th" ? "ประเภทที่วางสินค้า" : "Bin Type"}
                    </label>
                    <RadioChipPicker
                      options={BIN_TYPE_OPTIONS}
                      value={binForm.bintype}
                      onChange={(v) => setBinForm((prev) => ({ ...prev, bintype: v }))}
                      language={language}
                    />
                  </div>

                  <div className="border-t border-border/30 pt-3 flex flex-col gap-2">
                    <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider block">
                      {language === "th" ? "ตำแหน่งจัดเก็บ (Aisle / Rack / Level / Position)" : "Aisle / Rack / Level / Position"}
                    </span>
                    <div className="grid grid-cols-4 gap-2">
                      <Input
                        value={binForm.aislecode}
                        onChange={(e) => setBinForm((prev) => ({ ...prev, aislecode: e.target.value }))}
                        placeholder={language === "th" ? "ทางเดิน" : "Aisle"}
                        className="h-9 text-xs"
                      />
                      <Input
                        value={binForm.rackcode}
                        onChange={(e) => setBinForm((prev) => ({ ...prev, rackcode: e.target.value }))}
                        placeholder={language === "th" ? "ชั้นวาง" : "Rack"}
                        className="h-9 text-xs"
                      />
                      <Input
                        value={binForm.levelcode}
                        onChange={(e) => setBinForm((prev) => ({ ...prev, levelcode: e.target.value }))}
                        placeholder={language === "th" ? "ระดับ" : "Level"}
                        className="h-9 text-xs"
                      />
                      <Input
                        value={binForm.positioncode}
                        onChange={(e) => setBinForm((prev) => ({ ...prev, positioncode: e.target.value }))}
                        placeholder={language === "th" ? "ตำแหน่ง" : "Position"}
                        className="h-9 text-xs"
                      />
                    </div>
                  </div>

                  <div className="border-t border-border/30 pt-3 flex flex-col gap-2">
                    <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider block">
                      {language === "th" ? "ขนาดและน้ำหนักบรรทุกสูงสุด" : "Dimensions & Max Weight"}
                    </span>
                    <div className="grid grid-cols-4 gap-2">
                      <div className="flex flex-col gap-1">
                        <label className="text-[10px] text-muted-foreground">{language === "th" ? "กว้าง (มม.)" : "Width (mm)"}</label>
                        <Input
                          type="number"
                          value={binForm.widthmm}
                          onChange={(e) => setBinForm((prev) => ({ ...prev, widthmm: e.target.value }))}
                          className="h-9 text-xs"
                        />
                      </div>
                      <div className="flex flex-col gap-1">
                        <label className="text-[10px] text-muted-foreground">{language === "th" ? "ยาว (มม.)" : "Length (mm)"}</label>
                        <Input
                          type="number"
                          value={binForm.lengthmm}
                          onChange={(e) => setBinForm((prev) => ({ ...prev, lengthmm: e.target.value }))}
                          className="h-9 text-xs"
                        />
                      </div>
                      <div className="flex flex-col gap-1">
                        <label className="text-[10px] text-muted-foreground">{language === "th" ? "สูง (มม.)" : "Height (mm)"}</label>
                        <Input
                          type="number"
                          value={binForm.heightmm}
                          onChange={(e) => setBinForm((prev) => ({ ...prev, heightmm: e.target.value }))}
                          className="h-9 text-xs"
                        />
                      </div>
                      <div className="flex flex-col gap-1">
                        <label className="text-[10px] text-muted-foreground">{language === "th" ? "น้ำหนักสูงสุด (กรัม)" : "Max Weight (g)"}</label>
                        <Input
                          type="number"
                          value={binForm.maxweightgram}
                          onChange={(e) => setBinForm((prev) => ({ ...prev, maxweightgram: e.target.value }))}
                          className="h-9 text-xs"
                        />
                      </div>
                      <div className="flex flex-col gap-1">
                        <label className="text-[10px] text-muted-foreground">{language === "th" ? "ปริมาตรสูงสุด (ซม.³)" : "Max Volume (cm³)"}</label>
                        <Input
                          type="number"
                          value={binForm.maxvolumecm3}
                          onChange={(e) => setBinForm((prev) => ({ ...prev, maxvolumecm3: e.target.value }))}
                          className="h-9 text-xs"
                        />
                      </div>
                    </div>
                  </div>

                  <div className="flex flex-col gap-2">
                    <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider block">
                      {language === "th" ? "ช่วงอุณหภูมิ (°C x10)" : "Temperature Range (°C x10)"}
                    </span>
                    <div className="grid grid-cols-2 gap-2">
                      <div className="flex flex-col gap-1">
                        <label className="text-[10px] text-muted-foreground">{language === "th" ? "ต่ำสุด" : "Min"}</label>
                        <Input
                          type="number"
                          value={binForm.mintemperaturedeci}
                          onChange={(e) => setBinForm((prev) => ({ ...prev, mintemperaturedeci: e.target.value }))}
                          className="h-9 text-xs"
                        />
                      </div>
                      <div className="flex flex-col gap-1">
                        <label className="text-[10px] text-muted-foreground">{language === "th" ? "สูงสุด" : "Max"}</label>
                        <Input
                          type="number"
                          value={binForm.maxtemperaturedeci}
                          onChange={(e) => setBinForm((prev) => ({ ...prev, maxtemperaturedeci: e.target.value }))}
                          className="h-9 text-xs"
                        />
                      </div>
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-3">
                    <div className="flex flex-col gap-1">
                      <label className="text-[11px] font-bold text-muted-foreground uppercase">
                        {language === "th" ? "ประเภทสินค้าที่อนุญาต (คั่นด้วย ,)" : "Allowed Product Classes (comma-separated)"}
                      </label>
                      <Input
                        value={binForm.allowedproductclasses}
                        onChange={(e) => setBinForm((prev) => ({ ...prev, allowedproductclasses: e.target.value }))}
                        className="h-9 text-xs"
                      />
                    </div>
                    <div className="flex flex-col gap-1">
                      <label className="text-[11px] font-bold text-muted-foreground uppercase">
                        {language === "th" ? "ประเภทวัตถุอันตราย (คั่นด้วย ,)" : "Hazard Classes (comma-separated)"}
                      </label>
                      <Input
                        value={binForm.hazardclasses}
                        onChange={(e) => setBinForm((prev) => ({ ...prev, hazardclasses: e.target.value }))}
                        className="h-9 text-xs"
                      />
                    </div>
                  </div>

                  <div className="flex flex-col gap-1.5">
                    <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider block">
                      {language === "th" ? "กฎการเข้า-ออก" : "In/Out Rules"}
                    </span>
                    <div className="flex flex-wrap gap-1.5">
                      <BooleanChip
                        label={language === "th" ? "รับเข้าได้ (Putaway)" : "Allow Putaway"}
                        checked={binForm.allowputaway}
                        onChange={(v) => setBinForm((prev) => ({ ...prev, allowputaway: v }))}
                      />
                      <BooleanChip
                        label={language === "th" ? "หยิบออกได้ (Pick)" : "Allow Pick"}
                        checked={binForm.allowpick}
                        onChange={(v) => setBinForm((prev) => ({ ...prev, allowpick: v }))}
                      />
                      <BooleanChip
                        label={language === "th" ? "บล็อกรับเข้า" : "Blocked In"}
                        checked={binForm.blockedin}
                        onChange={(v) => setBinForm((prev) => ({ ...prev, blockedin: v }))}
                        tone="danger"
                      />
                      <BooleanChip
                        label={language === "th" ? "บล็อกส่งออก" : "Blocked Out"}
                        checked={binForm.blockedout}
                        onChange={(v) => setBinForm((prev) => ({ ...prev, blockedout: v }))}
                        tone="danger"
                      />
                    </div>
                  </div>
                </div>
              )}

              {formError && (
                <div className="text-xs font-semibold text-destructive mt-1">{formError}</div>
              )}

            </form>
          )}
        </CardContent>
      </Card>

      <MapPickerDialog
        open={isMapOpen}
        initialLat={parseFloat(warehouseForm.latitude) || null}
        initialLng={parseFloat(warehouseForm.longitude) || null}
        language={language}
        onCancel={() => setIsMapOpen(false)}
        onSelect={(lat, lng) => {
          setWarehouseForm((prev) => ({ ...prev, latitude: String(lat), longitude: String(lng) }));
          setIsMapOpen(false);
        }}
      />
      {confirmationDialog}
    </div>
  );
}
