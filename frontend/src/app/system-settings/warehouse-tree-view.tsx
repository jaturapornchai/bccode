"use client";

import React, { useCallback, useMemo, useState, useEffect } from "react";
import {
  MapPin,
  Plus,
  ChevronRight,
  ChevronDown,
  Edit3,
  Trash2,
  Layers,
  GripVertical,
  ArrowUp,
  ArrowDown,
  FolderPlus,
  Save,
  Loader2,
  Warehouse,
} from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { LANGUAGES, type LanguageCode } from "@/lib/i18n";
import { cn } from "@/lib/utils";
import { normalizeLanguageConfigs } from "./system-settings-screen";
import { MapPickerDialog } from "@/components/map-picker-dialog";

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
  records: WarehouseRecord[];
  onRefresh?: () => void;
  saving: boolean;
  loading: boolean;
}

type NodeType = "warehouse" | "location" | "shelf";

interface WarehouseWorkspace {
  shop: { holding_code: string };
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

interface WarehouseShelf {
  code?: string;
  name?: string;
  productitems?: unknown[];
  max_weight?: number;
  width?: number;
  length?: number;
  height?: number;
  suitable_product_types?: string;
  [key: string]: unknown;
}

interface WarehouseLocation {
  code?: string;
  names?: LocalizedNames;
  shelf?: WarehouseShelf[];
  max_weight?: number;
  width?: number;
  length?: number;
  height?: number;
  suitable_product_types?: string;
  [key: string]: unknown;
}

interface WarehouseRecord {
  guid_fixed?: string;
  code?: string;
  names?: LocalizedNames;
  location?: WarehouseLocation[] | string;
  latitude?: number;
  longitude?: number;
  [key: string]: unknown;
}

interface SelectedNode {
  type: NodeType;
  warehouseId: string;
  locIndex?: number;
  shelfIndex?: number;
  data: Partial<WarehouseRecord> | WarehouseLocation | WarehouseShelf;
}

interface CompanyRecord {
  guid_fixed?: string;
  code?: string;
  names?: LocalizedNames;
  tax_id?: string;
  is_active?: boolean;
}

interface ApiResponse {
  message?: string;
}

const getNameFromObject = (names: LocalizedNames, code: string): string => {
  if (!names || Array.isArray(names) || typeof names !== "object") return "";
  const value = names[code];
  return typeof value === "string" ? value : "";
};

const errorMessage = (err: unknown, fallback: string): string =>
  err instanceof Error && err.message ? err.message : fallback;

export function WarehouseTreeView({
  auth,
  workspace,
  language,
  records,
  onRefresh,
  saving: globalSaving,
  loading,
}: WarehouseTreeViewProps) {
  // Active languages for multilingual names
  const editorLanguages = useMemo(() => {
    if (!workspace) return ["th"];
    const configs = workspace.shopInfo?.settings?.languageconfigs || [];
    const defaultCode = workspace.shopInfo?.settings?.language || "th";
    return normalizeLanguageConfigs(configs, defaultCode).map((row) => row.code);
  }, [workspace]);

  // Collapsed states
  const [collapsedWarehouses, setCollapsedWarehouses] = useState<Record<string, boolean>>({});
  const [collapsedLocs, setCollapsedLocs] = useState<Record<string, boolean>>({});

  const [companiesList, setCompaniesList] = useState<CompanyRecord[]>([]);

  useEffect(() => {
    if (!auth) return;
    fetch(`${auth.backendUrl}/organization/company`, {
      headers: { Authorization: `Bearer ${auth.token}` },
    })
      .then((res) => res.json())
      .then((json) => {
        if (json.success && Array.isArray(json.data)) {
          setCompaniesList(json.data);
        }
      })
      .catch((err) => console.error(err));
  }, [auth]);

  // Helper to get locations list for any warehouse record
  const getLocationsList = useCallback((warehouse: WarehouseRecord): WarehouseLocation[] => {
    if (!warehouse) return [];
    const raw = warehouse.location;
    if (Array.isArray(raw)) return raw;
    if (typeof raw === "string" && raw.trim()) {
      try {
        const parsed = JSON.parse(raw);
        if (Array.isArray(parsed)) return parsed as WarehouseLocation[];
      } catch {}
    }
    return [];
  }, []);

  // Selected Node for Right Form Editing
  const [selectedNode, setSelectedNode] = useState<SelectedNode | null>(null);

  // Initialize selected node on load or change
  useEffect(() => {
    if (records.length > 0 && !selectedNode) {
      const first = records[0];
      setSelectedNode({
        type: "warehouse",
        warehouseId: first.guid_fixed || "",
        data: {
          code: first.code || "",
          names: first.names || [],
        },
      });
      setFormType("edit_warehouse");
    }
  }, [records, selectedNode]);

  // Search filter
  const [searchQuery, setSearchQuery] = useState("");

  // Edit / Form state on the right pane
  const [formType, setFormType] = useState<
    "create_warehouse" | "edit_warehouse" | "create_location" | "edit_location" | "create_shelf" | "edit_shelf" | "bulk_shelf" | null
  >("edit_warehouse");

  // Local state for the inline form fields
  const [formFields, setFormFields] = useState<{
    code: string;
    name: string;
    names: Record<string, string>; // Multilingual names
    max_weight: string;
    width: string;
    length: string;
    height: string;
    suitable_product_types: string;
    latitude: string;
    longitude: string;
    company_guids?: string[];
    // Bulk parameters
    bulkPrefix?: string;
    bulkStartNum?: string;
    bulkEndNum?: string;
    bulkPadding?: string;
    bulkNamePattern?: string;
  }>({
    code: "",
    name: "",
    names: {},
    max_weight: "",
    width: "",
    length: "",
    height: "",
    suitable_product_types: "",
    latitude: "",
    longitude: "",
    company_guids: [],
  });

  const [formError, setFormError] = useState("");
  const [isSavingLocal, setIsSavingLocal] = useState(false);
  const [isMapOpen, setIsMapOpen] = useState(false);

  // Helper to sync Form state when selectedNode or formType changes
  useEffect(() => {
    if (formType === "create_warehouse") {
      setFormError("");
      const namesMap: Record<string, string> = {};
      editorLanguages.forEach((lang) => {
        namesMap[lang] = "";
      });
      setFormFields({
        code: "",
        name: "",
        names: namesMap,
        max_weight: "",
        width: "",
        length: "",
        height: "",
        suitable_product_types: "",
        latitude: "",
        longitude: "",
        company_guids: [],
      });
      return;
    }

    if (!selectedNode) return;
    setFormError("");

    const warehouseId = selectedNode.warehouseId;
    const warehouse = records.find((r) => r.guid_fixed === warehouseId) || records[0];

    if (formType === "edit_warehouse") {
      const namesMap: Record<string, string> = {};
      editorLanguages.forEach((lang) => {
        let val = "";
        const names = warehouse?.names;
        if (Array.isArray(names)) {
          val = names.find((n) => n.code === lang)?.name || "";
        } else if (names && typeof names === "object") {
          val = getNameFromObject(names, lang);
        }
        namesMap[lang] = val;
      });

      setFormFields({
        code: warehouse?.code || "",
        name: "",
        names: namesMap,
        max_weight: "",
        width: "",
        length: "",
        height: "",
        suitable_product_types: "",
        latitude: warehouse?.latitude !== undefined && warehouse?.latitude !== 0 ? String(warehouse.latitude) : "",
        longitude: warehouse?.longitude !== undefined && warehouse?.longitude !== 0 ? String(warehouse.longitude) : "",
        company_guids: (warehouse as any)?.company_guids || [],
      });
    } else if (formType === "edit_location" && selectedNode.type === "location") {
      const loc = selectedNode.data as WarehouseLocation;
      const namesMap: Record<string, string> = {};
      editorLanguages.forEach((lang) => {
        let val = "";
        if (Array.isArray(loc.names)) {
          val = loc.names.find((n) => n.code === lang)?.name || "";
        } else if (loc.names && typeof loc.names === "object") {
          val = getNameFromObject(loc.names, lang);
        }
        namesMap[lang] = val;
      });

      setFormFields({
        code: loc.code || "",
        name: "",
        names: namesMap,
        max_weight: loc.max_weight !== undefined && loc.max_weight !== 0 ? String(loc.max_weight) : "",
        width: loc.width !== undefined && loc.width !== 0 ? String(loc.width) : "",
        length: loc.length !== undefined && loc.length !== 0 ? String(loc.length) : "",
        height: loc.height !== undefined && loc.height !== 0 ? String(loc.height) : "",
        suitable_product_types: loc.suitable_product_types || "",
        latitude: "",
        longitude: "",
      });
    } else if (formType === "create_location") {
      const namesMap: Record<string, string> = {};
      editorLanguages.forEach((lang) => {
        namesMap[lang] = "";
      });
      setFormFields({
        code: "",
        name: "",
        names: namesMap,
        max_weight: "",
        width: "",
        length: "",
        height: "",
        suitable_product_types: "",
        latitude: "",
        longitude: "",
      });
    } else if (formType === "edit_shelf" && selectedNode.type === "shelf") {
      const shelf = selectedNode.data as WarehouseShelf;
      setFormFields({
        code: shelf.code || "",
        name: shelf.name || "",
        names: {},
        max_weight: shelf.max_weight !== undefined && shelf.max_weight !== 0 ? String(shelf.max_weight) : "",
        width: shelf.width !== undefined && shelf.width !== 0 ? String(shelf.width) : "",
        length: shelf.length !== undefined && shelf.length !== 0 ? String(shelf.length) : "",
        height: shelf.height !== undefined && shelf.height !== 0 ? String(shelf.height) : "",
        suitable_product_types: shelf.suitable_product_types || "",
        latitude: "",
        longitude: "",
      });
    } else if (formType === "create_shelf") {
      setFormFields({
        code: "",
        name: "",
        names: {},
        max_weight: "",
        width: "",
        length: "",
        height: "",
        suitable_product_types: "",
        latitude: "",
        longitude: "",
      });
    } else if (formType === "bulk_shelf") {
      setFormFields({
        code: "",
        name: "",
        names: {},
        max_weight: "",
        width: "",
        length: "",
        height: "",
        suitable_product_types: "",
        latitude: "",
        longitude: "",
        bulkPrefix: "SH-",
        bulkStartNum: "1",
        bulkEndNum: "10",
        bulkPadding: "2",
        bulkNamePattern: language === "th" ? "ชั้นวาง {number}" : "Shelf {number}",
      });
    }
  }, [selectedNode, formType, records, editorLanguages, language]);

  const getLanguageName = (code: string) => {
    const found = LANGUAGES.find((item) => item.code === code);
    return found ? found.name : code.toUpperCase();
  };

  // Helper to get localized name
  const displayWarehouseName = useCallback((names: LocalizedNames) => {
    if (Array.isArray(names)) {
      const found = names.find((item) => item.code === language);
      if (found?.name) return found.name;
      const th = names.find((item) => item.code === "th");
      if (th?.name) return th.name;
      if (names.length > 0) return names[0].name || "";
    } else if (names && typeof names === "object") {
      return getNameFromObject(names, language) || getNameFromObject(names, "th") || getNameFromObject(names, "en");
    }
    return "";
  }, [language]);

  const displayLocationName = useCallback((names: LocalizedNames) => {
    if (Array.isArray(names)) {
      const found = names.find((item) => item.code === language);
      if (found?.name) return found.name;
      const th = names.find((item) => item.code === "th");
      if (th?.name) return th.name;
      const en = names.find((item) => item.code === "en");
      if (en?.name) return en.name;
      if (names.length > 0) return names[0].name || "";
    } else if (names && typeof names === "object") {
      return getNameFromObject(names, language) || getNameFromObject(names, "th") || getNameFromObject(names, "en");
    }
    return "";
  }, [language]);

  // Save the Warehouse record using PUT API
  const saveWarehousePayload = async (
    targetWarehouse: WarehouseRecord,
    updatedLocations: WarehouseLocation[],
  ) => {
    if (!auth || !workspace || !targetWarehouse.guid_fixed) return;
    setIsSavingLocal(true);
    setFormError("");
    try {
      const payload = {
        ...targetWarehouse,
        location: updatedLocations,
        backendUrl: auth.backendUrl,
        holding_code: workspace.shop.holding_code,
      };

      const response = await fetch(
        `/api/system-settings/product_warehouse_screen/${targetWarehouse.guid_fixed}`,
        {
          method: "PUT",
          headers: {
            "Content-Type": "application/json",
            "x-bc-backend-url": auth.backendUrl,
            Authorization: `Bearer ${auth.token}`,
          },
          body: JSON.stringify(payload),
        }
      );

      const data = (await response.json()) as ApiResponse;
      if (!response.ok) {
        throw new Error(data.message || "Failed to save warehouse structure");
      }

      if (onRefresh) onRefresh();
    } catch (err: unknown) {
      setFormError(errorMessage(err, "An error occurred"));
    } finally {
      setIsSavingLocal(false);
    }
  };

  // Save changes from Right Pane Form
  const handleSaveForm = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!workspace) return;
    const warehouseId = selectedNode?.warehouseId;
    const warehouse = records.find((r) => r.guid_fixed === warehouseId) || records[0];
    if (!warehouse && formType !== "create_warehouse") return;

    const code = formFields.code.trim();
    if (formType !== "bulk_shelf" && !code) {
      setFormError(language === "th" ? "กรุณากรอกรหัส" : "Code is required");
      return;
    }

    if (formType === "create_warehouse" || formType === "edit_warehouse") {
      // Validate names
      const hasName = Object.values(formFields.names).some((n) => n.trim());
      if (!hasName) {
        setFormError(language === "th" ? "กรุณากรอกชื่อคลังอย่างน้อยหนึ่งภาษา" : "Please fill in at least one language name");
        return;
      }

      setIsSavingLocal(true);
      setFormError("");
      try {
        const namesArray = editorLanguages
          .map((lang) => ({
            code: lang,
            name: formFields.names[lang]?.trim() || "",
            isauto: false,
            isdelete: false,
          }))
          .filter((item) => item.name !== "");

        const latVal = parseFloat(formFields.latitude) || 0;
        const lngVal = parseFloat(formFields.longitude) || 0;

        const isCreate = formType === "create_warehouse";
        const payload = {
          code,
          names: namesArray,
          latitude: latVal,
          longitude: lngVal,
          company_guids: formFields.company_guids || [],
        };

        const url = isCreate
          ? `${auth?.backendUrl}/warehouse`
          : `${auth?.backendUrl}/warehouse/${warehouse.guid_fixed || ""}`;

        const response = await fetch(
          url,
          {
            method: isCreate ? "POST" : "PUT",
            headers: {
              "Content-Type": "application/json",
              Authorization: `Bearer ${auth?.token || ""}`,
            },
            body: JSON.stringify(payload),
          }
        );
        const data = (await response.json()) as ApiResponse;
        if (!response.ok) throw new Error(data.message || "Failed to save warehouse");

        if (onRefresh) onRefresh();
        setFormType("edit_warehouse");
      } catch (err: unknown) {
        setFormError(errorMessage(err, "Error saving warehouse"));
      } finally {
        setIsSavingLocal(false);
      }
    } else if (formType === "create_location" || formType === "edit_location") {
      // Validate names
      const hasName = Object.values(formFields.names).some((n) => n.trim());
      if (!hasName) {
        setFormError(language === "th" ? "กรุณากรอกชื่อโซนเก็บสินค้าอย่างน้อยหนึ่งภาษา" : "Please fill in at least one language name");
        return;
      }

      const locationsList = getLocationsList(warehouse);

      // Check duplicates
      const isDuplicate = locationsList.some((loc, idx) => {
        if (formType === "edit_location" && selectedNode?.locIndex === idx) return false;
        return loc.code?.toLowerCase() === code.toLowerCase();
      });

      if (isDuplicate) {
        setFormError(language === "th" ? "รหัสโซนเก็บสินค้านี้มีอยู่แล้ว" : "Location code already exists");
        return;
      }

      const namesArray = editorLanguages
        .map((lang) => ({
          code: lang,
          name: formFields.names[lang]?.trim() || "",
          isauto: false,
          isdelete: false,
        }))
        .filter((item) => item.name !== "");

      const maxWeight = parseFloat(formFields.max_weight) || 0;
      const w = parseFloat(formFields.width) || 0;
      const l = parseFloat(formFields.length) || 0;
      const h = parseFloat(formFields.height) || 0;
      const suitable = formFields.suitable_product_types.trim();

      const updated = [...locationsList];
      if (formType === "create_location") {
        updated.push({
          code,
          names: namesArray,
          shelf: [],
          max_weight: maxWeight,
          width: w,
          length: l,
          height: h,
          suitable_product_types: suitable,
        });
      } else if (formType === "edit_location" && selectedNode?.locIndex !== undefined) {
        updated[selectedNode.locIndex] = {
          ...updated[selectedNode.locIndex],
          code,
          names: namesArray,
          max_weight: maxWeight,
          width: w,
          length: l,
          height: h,
          suitable_product_types: suitable,
        };
      }

      await saveWarehousePayload(warehouse, updated);
      // Reset form to warehouse view
      setFormType("edit_warehouse");
    } else if (formType === "create_shelf" || formType === "edit_shelf") {
      const name = formFields.name.trim();
      if (!name) {
        setFormError(language === "th" ? "กรุณากรอกชื่อชั้นวาง" : "Shelf name is required");
        return;
      }

      const locIdx = selectedNode?.locIndex;
      if (locIdx === undefined) return;

      const locationsList = getLocationsList(warehouse);
      const loc = locationsList[locIdx];
      const shelves = loc.shelf || [];

      // Check duplicate
      const isDuplicate = shelves.some((shelf: WarehouseShelf, idx: number) => {
        if (formType === "edit_shelf" && selectedNode?.shelfIndex === idx) return false;
        return shelf.code?.toLowerCase() === code.toLowerCase();
      });

      if (isDuplicate) {
        setFormError(language === "th" ? "รหัสชั้นวางนี้มีอยู่แล้วในโซนเก็บสินค้านี้" : "Shelf code already exists in this location");
        return;
      }

      const maxWeight = parseFloat(formFields.max_weight) || 0;
      const w = parseFloat(formFields.width) || 0;
      const l = parseFloat(formFields.length) || 0;
      const h = parseFloat(formFields.height) || 0;
      const suitable = formFields.suitable_product_types.trim();

      const updatedShelves = [...shelves];
      if (formType === "create_shelf") {
        updatedShelves.push({
          code,
          name,
          productitems: [],
          max_weight: maxWeight,
          width: w,
          length: l,
          height: h,
          suitable_product_types: suitable,
        });
      } else if (formType === "edit_shelf" && selectedNode?.shelfIndex !== undefined) {
        updatedShelves[selectedNode.shelfIndex] = {
          ...updatedShelves[selectedNode.shelfIndex],
          code,
          name,
          max_weight: maxWeight,
          width: w,
          length: l,
          height: h,
          suitable_product_types: suitable,
        };
      }

      const updatedLocs = [...locationsList];
      updatedLocs[locIdx] = {
        ...loc,
        shelf: updatedShelves,
      };

      await saveWarehousePayload(warehouse, updatedLocs);
      setFormType("edit_warehouse");
    } else if (formType === "bulk_shelf") {
      const prefix = formFields.bulkPrefix?.trim() || "";
      const startNum = parseInt(formFields.bulkStartNum || "1");
      const endNum = parseInt(formFields.bulkEndNum || "10");
      const padding = parseInt(formFields.bulkPadding || "2") || 0;
      const namePattern = formFields.bulkNamePattern?.trim() || "";

      if (isNaN(startNum) || isNaN(endNum) || startNum < 0 || endNum < 0) {
        setFormError(language === "th" ? "กรุณากรอกช่วงตัวเลขที่ถูกต้อง" : "Please fill in a valid number range");
        return;
      }
      if (startNum > endNum) {
        setFormError(language === "th" ? "ตัวเลขเริ่มต้นต้องไม่มากกว่าตัวเลขสิ้นสุด" : "Start number cannot be greater than end number");
        return;
      }
      if (endNum - startNum > 100) {
        setFormError(language === "th" ? "สร้างได้สูงสุดครั้งละ 100 ชั้นวาง" : "You can create up to 100 shelves at a time");
        return;
      }

      const locIdx = selectedNode?.locIndex;
      if (locIdx === undefined) return;

      const locationsList = getLocationsList(warehouse);
      const loc = locationsList[locIdx];
      const shelves = loc.shelf || [];
      const updatedShelves = [...shelves];

      const maxWeight = parseFloat(formFields.max_weight) || 0;
      const w = parseFloat(formFields.width) || 0;
      const l = parseFloat(formFields.length) || 0;
      const h = parseFloat(formFields.height) || 0;
      const suitable = formFields.suitable_product_types.trim();

      const duplicates: string[] = [];

      for (let i = startNum; i <= endNum; i++) {
        let numStr = String(i);
        if (padding > 0) {
          numStr = numStr.padStart(padding, "0");
        }
        const shelfCode = `${prefix}${numStr}`;
        const shelfName = namePattern.replace("{number}", numStr);

        const isDuplicate = updatedShelves.some((shelf) => shelf.code?.toLowerCase() === shelfCode.toLowerCase());
        if (isDuplicate) {
          duplicates.push(shelfCode);
          continue;
        }

        updatedShelves.push({
          code: shelfCode,
          name: shelfName,
          productitems: [],
          max_weight: maxWeight,
          width: w,
          length: l,
          height: h,
          suitable_product_types: suitable,
        });
      }

      if (duplicates.length > 0 && duplicates.length === (endNum - startNum + 1)) {
        setFormError(language === "th" ? "รหัสชั้นวางที่สร้างมีอยู่แล้วทั้งหมด" : "All generated shelf codes already exist");
        return;
      }

      const updatedLocs = [...locationsList];
      updatedLocs[locIdx] = {
        ...loc,
        shelf: updatedShelves,
      };

      await saveWarehousePayload(warehouse, updatedLocs);
      setFormType("edit_warehouse");
      if (duplicates.length > 0) {
        alert(
          language === "th"
            ? `ข้ามรหัสที่ซ้ำกัน: ${duplicates.join(", ")}`
            : `Skipped duplicate codes: ${duplicates.join(", ")}`
        );
      }
    }
  };

  // Reorder Locations
  const moveLocation = async (warehouseId: string, fromIdx: number, toIdx: number, e: React.MouseEvent) => {
    e.stopPropagation();
    const warehouse = records.find((r) => r.guid_fixed === warehouseId);
    if (!warehouse) return;
    const locationsList = getLocationsList(warehouse);
    if (toIdx < 0 || toIdx >= locationsList.length) return;
    const updated = [...locationsList];
    const [removed] = updated.splice(fromIdx, 1);
    updated.splice(toIdx, 0, removed);
    await saveWarehousePayload(warehouse, updated);
  };

  // Reorder Shelves
  const moveShelf = async (warehouseId: string, locIdx: number, fromIdx: number, toIdx: number, e: React.MouseEvent) => {
    e.stopPropagation();
    const warehouse = records.find((r) => r.guid_fixed === warehouseId);
    if (!warehouse) return;
    const locationsList = getLocationsList(warehouse);
    const loc = locationsList[locIdx];
    const shelves = loc.shelf || [];
    if (toIdx < 0 || toIdx >= shelves.length) return;
    const updatedShelves = [...shelves];
    const [removed] = updatedShelves.splice(fromIdx, 1);
    updatedShelves.splice(toIdx, 0, removed);

    const updatedLocs = [...locationsList];
    updatedLocs[locIdx] = { ...loc, shelf: updatedShelves };
    await saveWarehousePayload(warehouse, updatedLocs);
  };

  // Delete handlers
  const handleDeleteWarehouse = async (warehouseId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    const warehouse = records.find((r) => r.guid_fixed === warehouseId);
    if (!warehouse) return;
    const confirmed = window.confirm(
      language === "th"
        ? `ต้องการลบคลังสินค้า "${displayWarehouseName(warehouse.names)}" ใช่หรือไม่?`
        : `Are you sure you want to delete warehouse "${displayWarehouseName(warehouse.names)}"?`
    );
    if (!confirmed) return;

    setIsSavingLocal(true);
    setFormError("");
    try {
      const response = await fetch(
        `${auth?.backendUrl}/warehouse/${encodeURIComponent(warehouseId)}`,
        {
          method: "DELETE",
          headers: {
            "Content-Type": "application/json",
            Authorization: `Bearer ${auth?.token || ""}`,
          },
        }
      );

      const data = (await response.json()) as ApiResponse;
      if (!response.ok) throw new Error(data.message || "Failed to delete warehouse");

      if (onRefresh) onRefresh();

      const remaining = records.filter((r) => r.guid_fixed !== warehouseId);
      if (remaining.length > 0) {
        setSelectedNode({
          type: "warehouse",
          warehouseId: remaining[0].guid_fixed || "",
          data: {
            code: remaining[0].code || "",
            names: remaining[0].names || [],
          },
        });
        setFormType("edit_warehouse");
      } else {
        setSelectedNode(null);
        setFormType(null);
      }
    } catch (err: unknown) {
      setFormError(errorMessage(err, "Error deleting warehouse"));
    } finally {
      setIsSavingLocal(false);
    }
  };

  const handleDeleteLocation = async (warehouseId: string, index: number, e: React.MouseEvent) => {
    e.stopPropagation();
    const warehouse = records.find((r) => r.guid_fixed === warehouseId);
    if (!warehouse) return;
    const locationsList = getLocationsList(warehouse);
    const location = locationsList[index];
    if (!location) return;
    const confirmed = window.confirm(
      language === "th"
        ? `ต้องการลบโซนเก็บสินค้า "${displayLocationName(location.names)}" ใช่หรือไม่?`
        : `Are you sure you want to delete location "${displayLocationName(location.names)}"?`
    );
    if (!confirmed) return;
    const updated = locationsList.filter((_, idx) => idx !== index);
    await saveWarehousePayload(warehouse, updated);
    setFormType("edit_warehouse");
  };

  const handleDeleteShelf = async (warehouseId: string, locIndex: number, shelfIndex: number, e: React.MouseEvent) => {
    e.stopPropagation();
    const warehouse = records.find((r) => r.guid_fixed === warehouseId);
    if (!warehouse) return;
    const locationsList = getLocationsList(warehouse);
    const loc = locationsList[locIndex];
    const shelf = loc?.shelf?.[shelfIndex];
    if (!loc || !shelf) return;
    const confirmed = window.confirm(
      language === "th"
        ? `ต้องการลบชั้นวาง "${shelf.name || shelf.code}" ใช่หรือไม่?`
        : `Are you sure you want to delete shelf "${shelf.name || shelf.code}"?`
    );
    if (!confirmed) return;

    const updatedShelves = (loc.shelf || []).filter((_, idx: number) => idx !== shelfIndex);
    const updatedLocs = [...locationsList];
    updatedLocs[locIndex] = {
      ...loc,
      shelf: updatedShelves,
    };
    await saveWarehousePayload(warehouse, updatedLocs);
    setFormType("edit_warehouse");
  };

  // Filtered Warehouses based on search query
  const filteredWarehouses = useMemo(() => {
    if (!searchQuery.trim()) return records;
    const needle = searchQuery.toLowerCase();
    return records.filter((w) => {
      const codeMatch = w.code?.toLowerCase().includes(needle);
      const nameMatch = displayWarehouseName(w.names).toLowerCase().includes(needle);
      const locs = getLocationsList(w);
      const locMatch = locs.some((loc) => {
        const locCodeMatch = loc.code?.toLowerCase().includes(needle);
        const locNameMatch = displayLocationName(loc.names).toLowerCase().includes(needle);
        const shelfMatch = (loc.shelf || []).some((sh) =>
          sh.code?.toLowerCase().includes(needle) || sh.name?.toLowerCase().includes(needle)
        );
        return locCodeMatch || locNameMatch || shelfMatch;
      });
      return codeMatch || nameMatch || locMatch;
    });
  }, [records, searchQuery, displayWarehouseName, displayLocationName, getLocationsList]);

  // Render the whole tree view
  return (
    <div className="grid w-full min-w-0 items-stretch gap-3 min-h-[calc(100dvh-12rem)] xl:grid-cols-[minmax(320px,0.95fr)_minmax(420px,1.05fr)]">
      {/* LEFT COLUMN: Warehouse Tree list */}
      <Card className="flex h-full min-h-0 flex-col overflow-hidden border-border bg-card shadow-sm">
        {/* Search Header */}
        <div className="flex items-center gap-2 border-b border-border/40 p-2.5 bg-secondary/5">
          <div className="relative flex-1">
            <Input
              value={searchQuery}
              onChange={(e) => setSearchQuery(e.target.value)}
              placeholder={
                language === "th"
                  ? "ค้นหาคลังสินค้า โซน หรือชั้นวาง..."
                  : "Search warehouse, location, or shelf..."
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
              setFormType("create_warehouse");
              setSelectedNode({
                type: "warehouse",
                warehouseId: "",
                data: {},
              });
            }}
            disabled={loading || globalSaving}
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
          ) : filteredWarehouses.length === 0 ? (
            <div className="flex h-full min-h-24 flex-col items-center justify-center gap-1.5 p-4 text-center text-sm text-muted-foreground">
              <span className="font-medium">
                {language === "th" ? "ไม่พบข้อมูลคลังสินค้า" : "No warehouses found"}
              </span>
            </div>
          ) : (
            <div className="flex min-h-0 flex-1 flex-col overflow-y-auto overscroll-contain">
              {filteredWarehouses.map((w, warehouseIdx) => {
                const warehouseId = w.guid_fixed || "";
                const isWarehouseSelected =
                  selectedNode?.type === "warehouse" && selectedNode?.warehouseId === warehouseId;
                const isWarehouseCollapsed = collapsedWarehouses[warehouseId] ?? false;
                const locations = getLocationsList(w);
                const level0Style = warehouseLevelStyle(0);

                return (
                  <div key={warehouseId} className="flex flex-col border-b border-border/30 last:border-b-0">
                    {/* Level 0: Warehouse Node */}
                    <div
                      className={cn(
                        "group/row relative flex items-center justify-between border-b border-border/40 py-2 px-3 transition-colors cursor-pointer",
                        isWarehouseSelected ? level0Style.selectedBg : "hover:bg-muted/40"
                      )}
                      style={{ paddingLeft: "12px" }}
                      onClick={() => {
                        setSelectedNode({
                          type: "warehouse",
                          warehouseId,
                          data: {
                            code: w.code || "",
                            names: w.names || [],
                          },
                        });
                        setFormType("edit_warehouse");
                      }}
                    >
                      {isWarehouseSelected && (
                        <span className={cn("pointer-events-none absolute bottom-0 left-0 top-0 w-1", level0Style.borderLeft)} />
                      )}
                      <div className="flex min-w-0 flex-1 flex-wrap items-center gap-x-2 gap-y-1 pr-2">
                        <GripVertical className={cn("size-4 shrink-0 transition-colors", level0Style.grip)} />

                        {/* Expand/Collapse Toggle */}
                        <button
                          type="button"
                          className={cn(
                            "size-6 flex items-center justify-center rounded hover:bg-muted text-muted-foreground shrink-0",
                            locations.length === 0 && "invisible"
                          )}
                          onClick={(e) => {
                            e.stopPropagation();
                            setCollapsedWarehouses((prev) => ({
                              ...prev,
                              [warehouseId]: !isWarehouseCollapsed,
                            }));
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
                          {w.code} - {displayWarehouseName(w.names)}
                        </span>

                        {locations.length > 0 ? (
                          <button
                            type="button"
                            className="shrink-0 rounded-full border border-primary/20 bg-primary/10 px-2 py-0.5 text-[11px] font-semibold leading-5 text-primary transition-colors hover:border-primary/40 hover:bg-primary/15"
                            onClick={(e) => {
                              e.stopPropagation();
                              setCollapsedWarehouses((prev) => ({
                                ...prev,
                                [warehouseId]: !isWarehouseCollapsed,
                              }));
                            }}
                          >
                            {language === "th"
                              ? `ลูก ${locations.length}`
                              : `${locations.length} ${locations.length === 1 ? "child" : "children"}`}
                          </button>
                        ) : null}
                      </div>

                      {/* Warehouse Row Actions */}
                      <div className="flex shrink-0 items-center gap-1 opacity-60 group-hover/row:opacity-100 transition-opacity">
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          className="size-7 rounded-full text-emerald-600 hover:text-emerald-700 hover:bg-emerald-50 dark:hover:bg-emerald-950/40"
                          title={language === "th" ? "เพิ่มโซนเก็บสินค้า" : "Add Location"}
                          onClick={(e) => {
                            e.stopPropagation();
                            setSelectedNode({
                              type: "warehouse",
                              warehouseId,
                              data: {
                                code: w.code || "",
                                names: w.names || [],
                              },
                            });
                            setFormType("create_location");
                            setCollapsedWarehouses((prev) => ({ ...prev, [warehouseId]: false }));
                          }}
                        >
                          <FolderPlus className="size-3.5" />
                        </Button>
                        <Button
                          type="button"
                          variant="ghost"
                          size="icon"
                          className="size-7 rounded-full text-blue-600 hover:text-blue-700 hover:bg-blue-50 dark:hover:bg-blue-950/40"
                          title={language === "th" ? "แก้ไขคลังสินค้า" : "Edit Warehouse"}
                          onClick={(e) => {
                            e.stopPropagation();
                            setSelectedNode({
                              type: "warehouse",
                              warehouseId,
                              data: {
                                code: w.code || "",
                                names: w.names || [],
                              },
                            });
                            setFormType("edit_warehouse");
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
                          const isLocSelected =
                            selectedNode?.type === "location" &&
                            selectedNode?.warehouseId === warehouseId &&
                            selectedNode?.locIndex === locIdx;
                          const locKey = `${warehouseId}-${locIdx}`;
                          const isLocCollapsed = collapsedLocs[locKey] ?? false;
                          const shelves = loc.shelf || [];
                          const level1Style = warehouseLevelStyle(1);

                          return (
                            <div key={locIdx} className="flex flex-col">
                              {/* Location Node Row */}
                              <div
                                className={cn(
                                  "group/row relative flex items-center justify-between border-b border-border/40 py-2 px-3 transition-colors cursor-pointer",
                                  isLocSelected ? level1Style.selectedBg : "hover:bg-muted/40"
                                )}
                                style={{ paddingLeft: "36px" }}
                                onClick={() => {
                                  setSelectedNode({
                                    type: "location",
                                    warehouseId,
                                    locIndex: locIdx,
                                    data: loc,
                                  });
                                  setFormType("edit_location");
                                }}
                              >
                                {isLocSelected && (
                                  <span className={cn("pointer-events-none absolute bottom-0 left-0 top-0 w-1", level1Style.borderLeft)} />
                                )}
                                <div className="flex min-w-0 flex-1 flex-wrap items-center gap-x-2 gap-y-1 pr-2">
                                  <GripVertical className={cn("size-4 shrink-0 transition-colors", level1Style.grip)} />

                                  {/* Expand/Collapse Toggle */}
                                  <button
                                    type="button"
                                    className={cn(
                                      "size-6 flex items-center justify-center rounded hover:bg-muted text-muted-foreground shrink-0",
                                      shelves.length === 0 && "invisible"
                                    )}
                                    onClick={(e) => {
                                      e.stopPropagation();
                                      setCollapsedLocs((prev) => ({
                                        ...prev,
                                        [locKey]: !isLocCollapsed,
                                      }));
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
                                    {loc.code} - {displayLocationName(loc.names)}
                                  </span>

                                  {shelves.length > 0 ? (
                                    <button
                                      type="button"
                                      className="shrink-0 rounded-full border border-sky-200 bg-sky-100 dark:border-sky-800 dark:bg-sky-950 px-2 py-0.5 text-[11px] font-semibold leading-5 text-sky-700 dark:text-sky-300 transition-colors hover:bg-sky-200 dark:hover:bg-sky-900"
                                      onClick={(e) => {
                                        e.stopPropagation();
                                        setCollapsedLocs((prev) => ({
                                          ...prev,
                                          [locKey]: !isLocCollapsed,
                                        }));
                                      }}
                                    >
                                      {language === "th"
                                        ? `ลูก ${shelves.length}`
                                        : `${shelves.length} ${shelves.length === 1 ? "child" : "children"}`}
                                    </button>
                                  ) : null}
                                </div>

                                {/* Location Row Actions */}
                                <div className="flex shrink-0 items-center gap-1 opacity-60 group-hover/row:opacity-100 transition-opacity">
                                  {/* Reordering locations */}
                                  <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    disabled={locIdx === 0}
                                    className="size-7 rounded-full text-muted-foreground disabled:opacity-30 hover:bg-muted"
                                    title={language === "th" ? "เลื่อนขึ้น" : "Move Up"}
                                    onClick={(e) => moveLocation(warehouseId, locIdx, locIdx - 1, e)}
                                  >
                                    <ArrowUp className="size-3.5" />
                                  </Button>
                                  <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    disabled={locIdx === locations.length - 1}
                                    className="size-7 rounded-full text-muted-foreground disabled:opacity-30 hover:bg-muted"
                                    title={language === "th" ? "เลื่อนลง" : "Move Down"}
                                    onClick={(e) => moveLocation(warehouseId, locIdx, locIdx + 1, e)}
                                  >
                                    <ArrowDown className="size-3.5" />
                                  </Button>
                                  <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    className="size-7 rounded-full text-emerald-600 hover:text-emerald-700 hover:bg-emerald-50 dark:hover:bg-emerald-950/40"
                                    title={language === "th" ? "เพิ่มชั้นวาง" : "Add Shelf"}
                                    onClick={(e) => {
                                      e.stopPropagation();
                                      setSelectedNode({
                                        type: "location",
                                        warehouseId,
                                        locIndex: locIdx,
                                        data: loc,
                                      });
                                      setFormType("create_shelf");
                                      setCollapsedLocs((prev) => ({ ...prev, [locKey]: false }));
                                    }}
                                  >
                                    <Plus className="size-3.5" />
                                  </Button>
                                  <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    className="size-7 rounded-full text-orange-600 hover:text-orange-700 hover:bg-orange-50 dark:hover:bg-orange-950/40"
                                    title={language === "th" ? "เพิ่มกลุ่มชั้นวาง" : "Bulk Add Shelves"}
                                    onClick={(e) => {
                                      e.stopPropagation();
                                      setSelectedNode({
                                        type: "location",
                                        warehouseId,
                                        locIndex: locIdx,
                                        data: loc,
                                      });
                                      setFormType("bulk_shelf");
                                      setCollapsedLocs((prev) => ({ ...prev, [locKey]: false }));
                                    }}
                                  >
                                    <Layers className="size-3.5" />
                                  </Button>
                                  <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    className="size-7 rounded-full text-blue-600 hover:text-blue-700 hover:bg-blue-50 dark:hover:bg-blue-950/40"
                                    title={language === "th" ? "แก้ไข" : "Edit Location"}
                                    onClick={(e) => {
                                      e.stopPropagation();
                                      setSelectedNode({
                                        type: "location",
                                        warehouseId,
                                        locIndex: locIdx,
                                        data: loc,
                                      });
                                      setFormType("edit_location");
                                    }}
                                  >
                                    <Edit3 className="size-3.5" />
                                  </Button>
                                  <Button
                                    type="button"
                                    variant="ghost"
                                    size="icon"
                                    className="size-7 rounded-full text-destructive hover:bg-destructive/10"
                                    title={language === "th" ? "ลบโซนเก็บสินค้า" : "Delete Location"}
                                    onClick={(e) => handleDeleteLocation(warehouseId, locIdx, e)}
                                  >
                                    <Trash2 className="size-3.5" />
                                  </Button>
                                </div>
                              </div>

                              {/* Level 2: Shelves List */}
                              {!isLocCollapsed && shelves.length > 0 && (
                                <div className="flex flex-col bg-secondary/10">
                                  {shelves.map((shelf, shelfIdx) => {
                                    const isShelfSelected =
                                      selectedNode?.type === "shelf" &&
                                      selectedNode?.warehouseId === warehouseId &&
                                      selectedNode?.locIndex === locIdx &&
                                      selectedNode?.shelfIndex === shelfIdx;
                                    const level2Style = warehouseLevelStyle(2);

                                    return (
                                      <div
                                        key={shelfIdx}
                                        className={cn(
                                          "group/row relative flex items-center justify-between border-b border-border/40 py-1.5 px-3 transition-colors cursor-pointer",
                                          isShelfSelected ? level2Style.selectedBg : "hover:bg-muted/40"
                                        )}
                                        style={{ paddingLeft: "60px" }}
                                        onClick={() => {
                                          setSelectedNode({
                                            type: "shelf",
                                            warehouseId,
                                            locIndex: locIdx,
                                            shelfIndex: shelfIdx,
                                            data: shelf,
                                          });
                                          setFormType("edit_shelf");
                                        }}
                                      >
                                        {isShelfSelected && (
                                          <span className={cn("pointer-events-none absolute bottom-0 left-0 top-0 w-1", level2Style.borderLeft)} />
                                        )}
                                        <div className="flex min-w-0 flex-1 flex-wrap items-center gap-x-2 gap-y-1 pr-2">
                                          <GripVertical className={cn("size-4 shrink-0 transition-colors", level2Style.grip)} />

                                          {/* Empty Caret Spacer */}
                                          <div className="size-6 shrink-0" />

                                          <span className={cn("min-w-[1.25rem] text-sm font-bold", level2Style.order)}>
                                            {shelfIdx + 1}
                                          </span>

                                          <span className={cn("min-w-0 flex-1 basis-40 whitespace-normal break-words text-sm", level2Style.name)}>
                                            {shelf.code} - {shelf.name}
                                          </span>
                                        </div>

                                        {/* Shelf Row Actions */}
                                        <div className="flex shrink-0 items-center gap-1 opacity-60 group-hover/row:opacity-100 transition-opacity">
                                          <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon"
                                            disabled={shelfIdx === 0}
                                            className="size-7 rounded-full text-muted-foreground disabled:opacity-30 hover:bg-muted"
                                            title={language === "th" ? "เลื่อนขึ้น" : "Move Up"}
                                            onClick={(e) => moveShelf(warehouseId, locIdx, shelfIdx, shelfIdx - 1, e)}
                                          >
                                            <ArrowUp className="size-3.5" />
                                          </Button>
                                          <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon"
                                            disabled={shelfIdx === shelves.length - 1}
                                            className="size-7 rounded-full text-muted-foreground disabled:opacity-30 hover:bg-muted"
                                            title={language === "th" ? "เลื่อนลง" : "Move Down"}
                                            onClick={(e) => moveShelf(warehouseId, locIdx, shelfIdx, shelfIdx + 1, e)}
                                          >
                                            <ArrowDown className="size-3.5" />
                                          </Button>
                                          <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon"
                                            className="size-7 rounded-full text-blue-600 hover:text-blue-700 hover:bg-blue-50 dark:hover:bg-blue-950/40"
                                            title={language === "th" ? "แก้ไข" : "Edit Shelf"}
                                            onClick={(e) => {
                                              e.stopPropagation();
                                              setSelectedNode({
                                                type: "shelf",
                                                warehouseId,
                                                locIndex: locIdx,
                                                shelfIndex: shelfIdx,
                                                data: shelf,
                                              });
                                              setFormType("edit_shelf");
                                            }}
                                          >
                                            <Edit3 className="size-3.5" />
                                          </Button>
                                          <Button
                                            type="button"
                                            variant="ghost"
                                            size="icon"
                                            className="size-7 rounded-full text-destructive hover:bg-destructive/10"
                                            title={language === "th" ? "ลบชั้นวาง" : "Delete Shelf"}
                                            onClick={(e) => handleDeleteShelf(warehouseId, locIdx, shelfIdx, e)}
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
            {formType === "create_warehouse" && (
              <>
                <Plus className="size-4 text-emerald-600 shrink-0" />
                {language === "th" ? "เพิ่มคลังสินค้า" : "Add Warehouse"}
              </>
            )}
            {formType === "edit_warehouse" && (
              <>
                <Warehouse className="size-4 text-primary shrink-0" />
                {language === "th" ? "แก้ไขคลังสินค้า" : "Edit Warehouse"}
              </>
            )}
            {formType === "create_location" && (
              <>
                <Plus className="size-4 text-emerald-600 shrink-0" />
                {language === "th" ? "เพิ่มโซนเก็บสินค้า" : "Add Location"}
              </>
            )}
            {formType === "edit_location" && (
              <>
                <Edit3 className="size-4 text-blue-600 shrink-0" />
                {language === "th" ? "แก้ไขโซนเก็บสินค้า" : "Edit Location"}
              </>
            )}
            {formType === "create_shelf" && (
              <>
                <Plus className="size-4 text-emerald-600 shrink-0" />
                {language === "th" ? "เพิ่มชั้นวางสินค้า" : "Add Shelf"}
              </>
            )}
            {formType === "edit_shelf" && (
              <>
                <Edit3 className="size-4 text-blue-600 shrink-0" />
                {language === "th" ? "แก้ไขชั้นวางสินค้า" : "Edit Shelf"}
              </>
            )}
            {formType === "bulk_shelf" && (
              <>
                <Layers className="size-4 text-orange-600 shrink-0" />
                {language === "th" ? "เพิ่มกลุ่มชั้นวางสินค้า" : "Bulk Add Shelves"}
              </>
            )}
          </h3>
        </div>

        <CardContent className="min-h-0 flex-1 overflow-y-auto overscroll-contain p-3">
          {!formType ? (
            <div className="flex flex-col items-center justify-center h-48 text-center text-xs text-muted-foreground italic gap-2">
              <MapPin className="size-8 text-muted-foreground/30 animate-pulse" />
              {language === "th"
                ? "เลือกคลังสินค้า โซนเก็บสินค้า หรือชั้นวาง ในโครงสร้างด้านซ้ายเพื่อทำการแก้ไข"
                : "Select warehouse, location, or shelf on the left to edit."}
            </div>
          ) : (
            <form onSubmit={handleSaveForm} className="flex flex-col gap-3">
              {/* Form fields for WAREHOUSE editing */}
              {(formType === "create_warehouse" || formType === "edit_warehouse") && (
                <div className="flex flex-col gap-3">
                  <div className="flex flex-col gap-1">
                    <label className="text-[11px] font-bold text-muted-foreground uppercase">
                      {language === "th" ? "รหัสคลังสินค้า" : "Warehouse Code"} <span className="text-destructive">*</span>
                    </label>
                    <Input
                      value={formFields.code}
                      onChange={(e) => setFormFields((prev) => ({ ...prev, code: e.target.value }))}
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
                          value={formFields.latitude}
                          onChange={(e) => setFormFields((prev) => ({ ...prev, latitude: e.target.value }))}
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
                          value={formFields.longitude}
                          onChange={(e) => setFormFields((prev) => ({ ...prev, longitude: e.target.value }))}
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

                  <div className="border-t border-border/30 pt-3 flex flex-col gap-3">
                    <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider block">
                      {language === "th" ? "ชื่อคลังสินค้าหลายภาษา" : "Multilingual Warehouse Names"}
                    </span>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                      {editorLanguages.map((lang) => (
                        <div key={lang} className="flex flex-col gap-1">
                          <label className="text-[10px] font-bold text-muted-foreground uppercase flex items-center gap-1">
                            {getLanguageName(lang)} {lang === "th" && <span className="text-destructive">*</span>}
                          </label>
                          <Input
                            value={formFields.names[lang] || ""}
                            onChange={(e) =>
                              setFormFields((prev) => ({
                                ...prev,
                                names: { ...prev.names, [lang]: e.target.value },
                              }))
                            }
                            placeholder={lang === "th" ? "เช่น คลังสินค้าหลัก" : "e.g. Main Warehouse"}
                            className="h-9 text-xs"
                          />
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              )}

              {/* Form fields for LOCATION creation/editing */}
              {(formType === "create_location" || formType === "edit_location") && (
                <div className="flex flex-col gap-3">
                  <div className="grid grid-cols-2 gap-3">
                    <div className="flex flex-col gap-1">
                      <label className="text-[11px] font-bold text-muted-foreground uppercase">
                        {language === "th" ? "รหัสโซนเก็บสินค้า" : "Storage Zone Code"} <span className="text-destructive">*</span>
                      </label>
                      <Input
                        value={formFields.code}
                        onChange={(e) => setFormFields((prev) => ({ ...prev, code: e.target.value }))}
                        placeholder="e.g. ZONE-A"
                        className="h-9 text-xs"
                      />
                    </div>

                    <div className="flex flex-col gap-1">
                      <label className="text-[11px] font-bold text-muted-foreground uppercase">
                        {language === "th" ? "ประเภทสินค้าที่เหมาะสม" : "Suitable Product Types"}
                      </label>
                      <Input
                        value={formFields.suitable_product_types}
                        onChange={(e) => setFormFields((prev) => ({ ...prev, suitable_product_types: e.target.value }))}
                        placeholder={language === "th" ? "เช่น ของแช่แข็ง, ของเหลว" : "e.g. Frozen, Liquids"}
                        className="h-9 text-xs"
                      />
                    </div>
                  </div>

                  <div className="border-t border-border/30 pt-3 flex flex-col gap-3">
                    <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider block">
                      {language === "th" ? "ชื่อโซนเก็บสินค้าหลายภาษา" : "Multilingual Storage Zone Names"}
                    </span>
                    <div className="grid grid-cols-1 md:grid-cols-2 gap-3">
                      {editorLanguages.map((lang) => (
                        <div key={lang} className="flex flex-col gap-1">
                          <label className="text-[10px] font-bold text-muted-foreground uppercase flex items-center gap-1">
                            {getLanguageName(lang)} {lang === "th" && <span className="text-destructive">*</span>}
                          </label>
                          <Input
                            value={formFields.names[lang] || ""}
                            onChange={(e) =>
                              setFormFields((prev) => ({
                                ...prev,
                                names: { ...prev.names, [lang]: e.target.value },
                              }))
                            }
                            placeholder={lang === "th" ? "เช่น โซนเอ" : "e.g. Zone A"}
                            className="h-9 text-xs"
                          />
                        </div>
                      ))}
                    </div>
                  </div>
                </div>
              )}

              {/* Form fields for SHELF creation/editing */}
              {(formType === "create_shelf" || formType === "edit_shelf") && (
                  <div className="flex flex-col gap-3">
                  <div className="grid grid-cols-2 gap-3">
                    <div className="flex flex-col gap-1">
                      <label className="text-[11px] font-bold text-muted-foreground uppercase">
                        {language === "th" ? "รหัสชั้นวาง" : "Shelf Code"} <span className="text-destructive">*</span>
                      </label>
                      <Input
                        value={formFields.code}
                        onChange={(e) => setFormFields((prev) => ({ ...prev, code: e.target.value }))}
                        placeholder="e.g. SH-01"
                        className="h-9 text-xs"
                      />
                    </div>

                    <div className="flex flex-col gap-1">
                      <label className="text-[11px] font-bold text-muted-foreground uppercase">
                        {language === "th" ? "ชื่อชั้นวาง" : "Shelf Name"} <span className="text-destructive">*</span>
                      </label>
                      <Input
                        value={formFields.name}
                        onChange={(e) => setFormFields((prev) => ({ ...prev, name: e.target.value }))}
                        placeholder="e.g. Row A, Tier 1"
                        className="h-9 text-xs"
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-2 gap-3">
                    <div className="flex flex-col gap-1">
                      <label className="text-[11px] font-bold text-muted-foreground uppercase">
                        {language === "th" ? "น้ำหนักบรรทุกสูงสุด (กก.)" : "Max Weight (kg)"}
                      </label>
                      <Input
                        type="number"
                        value={formFields.max_weight}
                        onChange={(e) => setFormFields((prev) => ({ ...prev, max_weight: e.target.value }))}
                        placeholder="e.g. 200"
                        className="h-9 text-xs"
                      />
                    </div>

                    <div className="flex flex-col gap-1">
                      <label className="text-[11px] font-bold text-muted-foreground uppercase">
                        {language === "th" ? "ประเภทสินค้าที่เหมาะสม" : "Suitable Product Types"}
                      </label>
                      <Input
                        value={formFields.suitable_product_types}
                        onChange={(e) => setFormFields((prev) => ({ ...prev, suitable_product_types: e.target.value }))}
                        placeholder="e.g. General"
                        className="h-9 text-xs"
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-3 gap-2">
                    <div className="flex flex-col gap-1">
                      <label className="text-[11px] font-bold text-muted-foreground uppercase">
                        {language === "th" ? "กว้าง (ซม.)" : "Width (cm)"}
                      </label>
                      <Input
                        type="number"
                        value={formFields.width}
                        onChange={(e) => setFormFields((prev) => ({ ...prev, width: e.target.value }))}
                        placeholder="W"
                        className="h-9 text-xs"
                      />
                    </div>
                    <div className="flex flex-col gap-1">
                      <label className="text-[11px] font-bold text-muted-foreground uppercase">
                        {language === "th" ? "ยาว (ซม.)" : "Length (cm)"}
                      </label>
                      <Input
                        type="number"
                        value={formFields.length}
                        onChange={(e) => setFormFields((prev) => ({ ...prev, length: e.target.value }))}
                        placeholder="L"
                        className="h-9 text-xs"
                      />
                    </div>
                    <div className="flex flex-col gap-1">
                      <label className="text-[11px] font-bold text-muted-foreground uppercase">
                        {language === "th" ? "สูง (ซม.)" : "Height (cm)"}
                      </label>
                      <Input
                        type="number"
                        value={formFields.height}
                        onChange={(e) => setFormFields((prev) => ({ ...prev, height: e.target.value }))}
                        placeholder="H"
                        className="h-9 text-xs"
                      />
                    </div>
                  </div>
                </div>
              )}

              {/* Form fields for BULK shelves generation */}
              {formType === "bulk_shelf" && (
                  <div className="flex flex-col gap-3">
                  <div className="bg-primary/5 border border-primary/10 rounded-xl p-3 text-[11px] text-foreground/80 leading-relaxed">
                    {language === "th"
                      ? "ระบุช่วงหมายเลขเพื่อสร้างชั้นวางสินค้าหลายรายการในคลิกเดียว ตัวอย่างเช่น หากระบุคำนำหน้า SH- หมายเลขเริ่มต้น 1 สิ้นสุด 10 จะได้รหัส SH-01 ถึง SH-10"
                      : "Provide a number range to generate sequential shelves in one click. E.g. Prefix 'SH-', Start 1, End 10 generates SH-01 to SH-10."}
                  </div>

                  <div className="grid grid-cols-2 gap-3">
                    <div className="flex flex-col gap-1">
                      <label className="text-[11px] font-bold text-muted-foreground uppercase">
                        {language === "th" ? "คำนำหน้ารหัส (Prefix)" : "Code Prefix"} <span className="text-destructive">*</span>
                      </label>
                      <Input
                        value={formFields.bulkPrefix}
                        onChange={(e) => setFormFields((prev) => ({ ...prev, bulkPrefix: e.target.value }))}
                        placeholder="e.g. SH-"
                        className="h-9 text-xs"
                      />
                    </div>

                    <div className="flex flex-col gap-1">
                      <label className="text-[11px] font-bold text-muted-foreground uppercase">
                        {language === "th" ? "รูปแบบชื่อ (Pattern)" : "Name Pattern"} <span className="text-destructive">*</span>
                      </label>
                      <Input
                        value={formFields.bulkNamePattern}
                        onChange={(e) => setFormFields((prev) => ({ ...prev, bulkNamePattern: e.target.value }))}
                        placeholder="e.g. ชั้นวาง {number}"
                        className="h-9 text-xs"
                      />
                    </div>
                  </div>

                  <div className="grid grid-cols-3 gap-2">
                    <div className="flex flex-col gap-1">
                      <label className="text-[11px] font-bold text-muted-foreground uppercase">
                        {language === "th" ? "เริ่มที่หมายเลข" : "Start Number"} <span className="text-destructive">*</span>
                      </label>
                      <Input
                        type="number"
                        value={formFields.bulkStartNum}
                        onChange={(e) => setFormFields((prev) => ({ ...prev, bulkStartNum: e.target.value }))}
                        placeholder="1"
                        className="h-9 text-xs"
                      />
                    </div>
                    <div className="flex flex-col gap-1">
                      <label className="text-[11px] font-bold text-muted-foreground uppercase">
                        {language === "th" ? "สิ้นสุดที่หมายเลข" : "End Number"} <span className="text-destructive">*</span>
                      </label>
                      <Input
                        type="number"
                        value={formFields.bulkEndNum}
                        onChange={(e) => setFormFields((prev) => ({ ...prev, bulkEndNum: e.target.value }))}
                        placeholder="10"
                        className="h-9 text-xs"
                      />
                    </div>
                    <div className="flex flex-col gap-1">
                      <label className="text-[11px] font-bold text-muted-foreground uppercase">
                        {language === "th" ? "จำนวนหลัก (Zero Padding)" : "Zero Padding"}
                      </label>
                      <Input
                        type="number"
                        value={formFields.bulkPadding}
                        onChange={(e) => setFormFields((prev) => ({ ...prev, bulkPadding: e.target.value }))}
                        placeholder="2"
                        className="h-9 text-xs"
                      />
                    </div>
                  </div>

                  <div className="border-t border-border/30 pt-3 flex flex-col gap-3">
                    <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider block">
                      {language === "th" ? "คุณสมบัติเริ่มต้นของชั้นวางทั้งหมด" : "Default Shelf Attributes"}
                    </span>
                    <div className="grid grid-cols-2 gap-3">
                      <div className="flex flex-col gap-1">
                        <label className="text-[11px] font-bold text-muted-foreground uppercase">
                          {language === "th" ? "รับน้ำหนักสูงสุด (กก.)" : "Max Weight (kg)"}
                        </label>
                        <Input
                          type="number"
                          value={formFields.max_weight}
                          onChange={(e) => setFormFields((prev) => ({ ...prev, max_weight: e.target.value }))}
                          placeholder="e.g. 200"
                          className="h-9 text-xs"
                        />
                      </div>

                      <div className="flex flex-col gap-1">
                        <label className="text-[11px] font-bold text-muted-foreground uppercase">
                          {language === "th" ? "ประเภทสินค้าที่เหมาะสม" : "Suitable Product Types"}
                        </label>
                        <Input
                          value={formFields.suitable_product_types}
                          onChange={(e) => setFormFields((prev) => ({ ...prev, suitable_product_types: e.target.value }))}
                          placeholder="e.g. Medicine"
                          className="h-9 text-xs"
                        />
                      </div>
                    </div>

                    <div className="grid grid-cols-3 gap-2">
                      <div className="flex flex-col gap-1">
                        <label className="text-[11px] font-bold text-muted-foreground uppercase">
                          {language === "th" ? "กว้าง (ซม.)" : "Width (cm)"}
                        </label>
                        <Input
                          type="number"
                          value={formFields.width}
                          onChange={(e) => setFormFields((prev) => ({ ...prev, width: e.target.value }))}
                          placeholder="W"
                          className="h-9 text-xs"
                        />
                      </div>
                      <div className="flex flex-col gap-1">
                        <label className="text-[11px] font-bold text-muted-foreground uppercase">
                          {language === "th" ? "ยาว (ซม.)" : "Length (cm)"}
                        </label>
                        <Input
                          type="number"
                          value={formFields.length}
                          onChange={(e) => setFormFields((prev) => ({ ...prev, length: e.target.value }))}
                          placeholder="L"
                          className="h-9 text-xs"
                        />
                      </div>
                      <div className="flex flex-col gap-1">
                        <label className="text-[11px] font-bold text-muted-foreground uppercase">
                          {language === "th" ? "สูง (ซม.)" : "Height (cm)"}
                        </label>
                        <Input
                          type="number"
                          value={formFields.height}
                          onChange={(e) => setFormFields((prev) => ({ ...prev, height: e.target.value }))}
                          placeholder="H"
                          className="h-9 text-xs"
                        />
                      </div>
                    </div>
                  </div>
                </div>
              )}

              {/* Error warning and footer */}
              {formError && (
                <div className="text-xs font-semibold text-destructive mt-1">
                  {formError}
                </div>
              )}

              <div className="flex justify-end gap-2 border-t border-border/30 pt-4 mt-2">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => {
                    setFormType("edit_warehouse");
                    const targetId = selectedNode?.warehouseId || records[0]?.guid_fixed || "";
                    const w = records.find((r) => r.guid_fixed === targetId);
                    if (w) {
                      setSelectedNode({
                        type: "warehouse",
                        warehouseId: w.guid_fixed || "",
                        data: {
                          code: w.code || "",
                          names: w.names || [],
                        },
                      });
                    }
                  }}
                  className="h-8 text-xs font-semibold"
                >
                  {language === "th" ? "ยกเลิก" : "Cancel"}
                </Button>
                <Button
                  type="submit"
                  size="sm"
                  disabled={isSavingLocal || globalSaving}
                  className="h-8 text-xs font-semibold gap-1 bg-primary text-primary-foreground hover:bg-primary/95"
                >
                  {isSavingLocal ? (
                    <Loader2 className="animate-spin size-3" />
                  ) : (
                    <Save className="size-3.5" />
                  )}
                  {language === "th" ? "บันทึก" : "Save"}
                </Button>
              </div>
            </form>
          )}
        </CardContent>
      </Card>

      <MapPickerDialog
        open={isMapOpen}
        initialLat={parseFloat(formFields.latitude) || null}
        initialLng={parseFloat(formFields.longitude) || null}
        language={language}
        onCancel={() => setIsMapOpen(false)}
        onSelect={(lat, lng) => {
          setFormFields((prev) => ({
            ...prev,
            latitude: String(lat),
            longitude: String(lng),
          }));
          setIsMapOpen(false);
        }}
      />
    </div>
  );
}
