"use client";

import { useBackendText } from "@/components/backend-text-provider";
import { authFetch } from "@/lib/client-auth-session";
import React, { useCallback, useMemo, useState, useEffect } from "react";
import {
  MapPin,
  Plus,
  Edit3,
  Trash2,
  Save,
  Loader2,
  Warehouse,
  Boxes,
  AlertTriangle,
  X,
  Search,
  CheckCircle2,
  RotateCcw,
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
import { ResizableSplitter } from "@/components/ui/resizable-splitter";

interface WarehouseTreeViewProps {
  auth: { token: string; backendUrl: string } | null;
  workspace: WarehouseWorkspace | null;
  language: LanguageCode;
  onRefresh?: () => void;
}

interface WarehouseWorkspace {
  shop: { holdingcode: string };
  shopInfo?: {
    settings?: {
      language?: string;
      languageconfigs?: unknown;
    };
  } | null;
}

const WAREHOUSE_SIDEBAR_MIN_WIDTH = 220;
const WAREHOUSE_SIDEBAR_MAX_WIDTH = 520;
const WAREHOUSE_SIDEBAR_DEFAULT_WIDTH = 280;
const WAREHOUSE_SIDEBAR_STORAGE_KEY = "bc_warehouse_sidebar_width";

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

interface EditableLocationRow {
  tempId: string;
  guidfixed?: string;
  code: string;
  names: Record<string, string>;
  isNew?: boolean;
  isModified?: boolean;
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

interface WarehouseFormFields {
  code: string;
  names: Record<string, string>;
  latitude: string;
  longitude: string;
  companyguids: string[];
}

const emptyWarehouseForm = (languages: string[]): WarehouseFormFields => {
  const names: Record<string, string> = {};
  languages.forEach((l) => (names[l] = ""));
  return { code: "", names, latitude: "", longitude: "", companyguids: [] };
};

export function WarehouseTreeView({ auth, workspace, language, onRefresh }: WarehouseTreeViewProps) {
  const tr = useBackendText();
  const { confirm, confirmationDialog } = useConfirmDialog({
    defaultConfirmLabel: tr("common_confirm", "ยืนยัน"),
    defaultCancelLabel: tr("common_cancel", "ยกเลิก"),
  });

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

  // Tree data fetched from GET /warehouse/tree
  const [tree, setTree] = useState<WarehouseRecord[]>([]);
  const [loading, setLoading] = useState(false);
  const [loadError, setLoadError] = useState("");

  const loadTree = useCallback(async (): Promise<WarehouseRecord[]> => {
    if (!auth || !mainApiUrl) return [];
    setLoading(true);
    setLoadError("");
    try {
      const res = await authFetch(`${mainApiUrl}/warehouse/tree`, {
        headers: { Authorization: `Bearer ${auth.token}` },
        cache: "no-store",
      });
      const json = (await res.json()) as ApiResponse;
      if (!res.ok || json.success === false) {
        throw new Error(json.message || tr("st_load_warehouse_structure_failed", "โหลดโครงสร้างคลังสินค้าไม่สำเร็จ"));
      }
      const data = Array.isArray(json.data) ? (json.data as WarehouseRecord[]) : [];
      setTree(data);
      return data;
    } catch (err) {
      setLoadError(errorMessage(err, tr("st_load_warehouse_structure_failed", "โหลดโครงสร้างคลังสินค้าไม่สำเร็จ")));
      setTree([]);
      return [];
    } finally {
      setLoading(false);
    }
  }, [auth, mainApiUrl, tr]);

  useEffect(() => {
    void loadTree();
  }, [loadTree]);

  // Selected warehouse ID
  const [selectedWarehouseId, setSelectedWarehouseId] = useState<string>("");

  useEffect(() => {
    if (tree.length > 0) {
      if (!selectedWarehouseId || !tree.some((w) => w.guidfixed === selectedWarehouseId)) {
        setSelectedWarehouseId(tree[0].guidfixed || "");
      }
    } else {
      setSelectedWarehouseId("");
    }
  }, [tree, selectedWarehouseId]);

  const activeWarehouse = useMemo(
    () => tree.find((w) => w.guidfixed === selectedWarehouseId) || (tree.length > 0 ? tree[0] : null),
    [tree, selectedWarehouseId]
  );

  // ---- Resizable Sidebar Width State ----
  const [sidebarWidth, setSidebarWidth] = useState<number>(() => {
    if (typeof window !== "undefined") {
      const saved = window.localStorage.getItem(WAREHOUSE_SIDEBAR_STORAGE_KEY);
      if (saved) {
        const parsed = Number(saved);
        if (Number.isFinite(parsed) && parsed >= WAREHOUSE_SIDEBAR_MIN_WIDTH && parsed <= WAREHOUSE_SIDEBAR_MAX_WIDTH) {
          return parsed;
        }
      }
    }
    return WAREHOUSE_SIDEBAR_DEFAULT_WIDTH;
  });
  const [isResizingSidebar, setIsResizingSidebar] = useState(false);
  const containerRef = React.useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    if (!isResizingSidebar) return;
    const prevCursor = document.body.style.cursor;
    const prevUserSelect = document.body.style.userSelect;
    document.body.style.cursor = "col-resize";
    document.body.style.userSelect = "none";
    return () => {
      document.body.style.cursor = prevCursor;
      document.body.style.userSelect = prevUserSelect;
    };
  }, [isResizingSidebar]);

  const handleSidebarResizeStart = useCallback(
    (event: React.PointerEvent<HTMLDivElement>) => {
      if (event.button !== 0) return;
      event.preventDefault();
      setIsResizingSidebar(true);

      const container = containerRef.current;
      const containerLeft = container ? container.getBoundingClientRect().left : 0;

      const onPointerMove = (moveEvent: PointerEvent) => {
        const maxWidth = typeof window !== "undefined"
          ? Math.min(WAREHOUSE_SIDEBAR_MAX_WIDTH, Math.floor(window.innerWidth * 0.6))
          : WAREHOUSE_SIDEBAR_MAX_WIDTH;
        const currentOffset = moveEvent.clientX - containerLeft;
        const next = Math.min(maxWidth, Math.max(WAREHOUSE_SIDEBAR_MIN_WIDTH, currentOffset));
        setSidebarWidth(Math.round(next));
      };

      const onPointerUp = (upEvent: PointerEvent) => {
        window.removeEventListener("pointermove", onPointerMove);
        window.removeEventListener("pointerup", onPointerUp);
        setIsResizingSidebar(false);
        const maxWidth = typeof window !== "undefined"
          ? Math.min(WAREHOUSE_SIDEBAR_MAX_WIDTH, Math.floor(window.innerWidth * 0.6))
          : WAREHOUSE_SIDEBAR_MAX_WIDTH;
        const currentOffset = upEvent.clientX - containerLeft;
        const finalWidth = Math.min(maxWidth, Math.max(WAREHOUSE_SIDEBAR_MIN_WIDTH, currentOffset));
        const rounded = Math.round(finalWidth);
        setSidebarWidth(rounded);
        if (typeof window !== "undefined") {
          window.localStorage.setItem(WAREHOUSE_SIDEBAR_STORAGE_KEY, String(rounded));
        }
      };

      window.addEventListener("pointermove", onPointerMove);
      window.addEventListener("pointerup", onPointerUp);
    },
    [],
  );

  const handleSidebarResizeReset = useCallback(() => {
    setSidebarWidth(WAREHOUSE_SIDEBAR_DEFAULT_WIDTH);
    if (typeof window !== "undefined") {
      window.localStorage.setItem(WAREHOUSE_SIDEBAR_STORAGE_KEY, String(WAREHOUSE_SIDEBAR_DEFAULT_WIDTH));
    }
  }, []);

  const handleSidebarKeyDown = useCallback((event: React.KeyboardEvent<HTMLDivElement>) => {
    let next: ((prev: number) => number) | number | null = null;
    if (event.key === "ArrowLeft") {
      next = (prev: number) => Math.max(WAREHOUSE_SIDEBAR_MIN_WIDTH, prev - 16);
    } else if (event.key === "ArrowRight") {
      next = (prev: number) => Math.min(WAREHOUSE_SIDEBAR_MAX_WIDTH, prev + 16);
    } else if (event.key === "Home") {
      next = WAREHOUSE_SIDEBAR_MIN_WIDTH;
    } else if (event.key === "End") {
      next = WAREHOUSE_SIDEBAR_MAX_WIDTH;
    }
    if (next !== null) {
      event.preventDefault();
      setSidebarWidth((prev) => {
        const val = typeof next === "function" ? next(prev) : next;
        if (typeof window !== "undefined") {
          window.localStorage.setItem(WAREHOUSE_SIDEBAR_STORAGE_KEY, String(val));
        }
        return val;
      });
    }
  }, []);

  // ---- Warehouse Dialog (Create/Edit) ----
  const [isWarehouseDialogOpen, setIsWarehouseDialogOpen] = useState(false);
  const [editingWarehouseId, setEditingWarehouseId] = useState<string | null>(null);
  const [warehouseForm, setWarehouseForm] = useState<WarehouseFormFields>(() => emptyWarehouseForm(["th"]));
  const [warehouseFormError, setWarehouseFormError] = useState("");
  const [isSavingWarehouse, setIsSavingWarehouse] = useState(false);
  const [isMapOpen, setIsMapOpen] = useState(false);


  const openCreateWarehouse = () => {
    setEditingWarehouseId(null);
    setWarehouseForm(emptyWarehouseForm(editorLanguages));
    setWarehouseFormError("");
    setIsWarehouseDialogOpen(true);
  };

  const openEditWarehouse = (warehouseId: string) => {
    const w = tree.find((x) => x.guidfixed === warehouseId);
    if (!w) return;
    const names: Record<string, string> = {};
    editorLanguages.forEach((l) => {
      names[l] = w.names?.find((n) => n.code === l)?.name || "";
    });
    setEditingWarehouseId(warehouseId);
    setWarehouseForm({
      code: w.code || "",
      names,
      latitude: w.latitude ? String(w.latitude) : "",
      longitude: w.longitude ? String(w.longitude) : "",
      companyguids: w.companyguids || [],
    });
    setWarehouseFormError("");
    setIsWarehouseDialogOpen(true);
  };

  const handleSaveWarehouse = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!auth || !mainApiUrl) return;

    const code = warehouseForm.code.trim();
    if (!code) {
      setWarehouseFormError(tr("st_please_enter_warehouse_code", "กรุณากรอกรหัสคลังสินค้า"));
      return;
    }
    const namesArray = editorLanguages
      .map((code) => ({ code, name: warehouseForm.names[code]?.trim() || "" }))
      .filter((item) => item.name !== "");

    if (namesArray.length === 0) {
      setWarehouseFormError(tr("st_enter_warehouse_name_one_lang", "กรุณากรอกชื่อคลังอย่างน้อยหนึ่งภาษา"));
      return;
    }

    setIsSavingWarehouse(true);
    setWarehouseFormError("");
    try {
      const isCreate = !editingWarehouseId;
      const url = isCreate ? `${mainApiUrl}/warehouse` : `${mainApiUrl}/warehouse/${editingWarehouseId}`;
      const payload = {
        code,
        names: namesArray,
        latitude: parseFloat(warehouseForm.latitude) || 0,
        longitude: parseFloat(warehouseForm.longitude) || 0,
        companyguids: warehouseForm.companyguids,
        status: "active",
      };
      const res = await authFetch(url, {
        method: isCreate ? "POST" : "PUT",
        headers: { "Content-Type": "application/json", Authorization: `Bearer ${auth.token}` },
        body: JSON.stringify(payload),
      });
      const json = (await res.json()) as ApiResponse;
      if (!res.ok || json.success === false) throw new Error(json.message || "Failed to save warehouse");

      await loadTree();
      const targetId = isCreate ? json.id || "" : editingWarehouseId;
      if (targetId) setSelectedWarehouseId(targetId);
      setIsWarehouseDialogOpen(false);
      if (onRefresh) onRefresh();
    } catch (err) {
      setWarehouseFormError(errorMessage(err, "Error saving warehouse"));
    } finally {
      setIsSavingWarehouse(false);
    }
  };

  const handleDeleteWarehouse = async (warehouseId: string, e: React.MouseEvent) => {
    e.stopPropagation();
    if (!auth || !mainApiUrl) return;
    const w = tree.find((x) => x.guidfixed === warehouseId);
    if (!w) return;
    const confirmed = await confirm({
      title: tr("st_confirm_delete_warehouse", "ยืนยันการลบคลังสินค้า"),
      description:
        tr("st_confirm_delete_warehouse_2", "ต้องการลบคลังสินค้า \"{0} - {1}\" ใช่หรือไม่?").replace("{0}", String(w.code)).replace("{1}", String(displayName(w.names, language))),
      confirmLabel: tr("delete", "ลบ"),
      tone: "danger",
    });
    if (!confirmed) return;

    try {
      const res = await authFetch(`${mainApiUrl}/warehouse/${encodeURIComponent(warehouseId)}`, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${auth.token}` },
      });
      const json = (await res.json()) as ApiResponse;
      if (!res.ok || json.success === false) throw new Error(json.message || "Failed to delete warehouse");

      await loadTree();
      if (onRefresh) onRefresh();
    } catch (err) {
      alert(errorMessage(err, "Error deleting warehouse"));
    }
  };

  // ---- EDITABLE LOCATIONS TABLE (ตารางที่เก็บสินค้าแบบแก้ไขและเซฟพร้อมกันทีเดียว) ----
  const [locationRows, setLocationRows] = useState<EditableLocationRow[]>([]);
  const [deletedLocationGuids, setDeletedLocationGuids] = useState<string[]>([]);
  const [locationSearchQuery, setLocationSearchQuery] = useState("");
  const [warehouseSearchQuery, setWarehouseSearchQuery] = useState("");
  const [tableError, setTableError] = useState("");
  const [successMessage, setSuccessMessage] = useState("");
  const [isSavingLocations, setIsSavingLocations] = useState(false);

  // Helper to convert backend record to editable row
  const toEditableRow = useCallback(
    (loc: WarehouseLocationRecord): EditableLocationRow => {
      const names: Record<string, string> = {};
      editorLanguages.forEach((l) => {
        names[l] = loc.names?.find((n) => n.code === l)?.name || "";
      });
      return {
        tempId: loc.guidfixed || `new-${Math.random().toString(36).slice(2)}`,
        guidfixed: loc.guidfixed,
        code: loc.code || "",
        names,
        isNew: false,
        isModified: false,
      };
    },
    [editorLanguages]
  );

  // Sync rows from active warehouse
  useEffect(() => {
    if (activeWarehouse) {
      const rows = (activeWarehouse.locations || []).map((l) => toEditableRow(l));
      setLocationRows(rows);
      setDeletedLocationGuids([]);
      setTableError("");
    } else {
      setLocationRows([]);
    }
  }, [activeWarehouse, toEditableRow]);

  const handleAddRow = () => {
    const names: Record<string, string> = {};
    editorLanguages.forEach((l) => (names[l] = ""));
    const newRow: EditableLocationRow = {
      tempId: `new-${Date.now()}-${Math.random().toString(36).slice(2, 7)}`,
      code: "",
      names,
      isNew: true,
      isModified: false,
    };
    setLocationRows((prev) => [...prev, newRow]);
  };

  const handleUpdateRowCode = (tempId: string, code: string) => {
    setLocationRows((prev) =>
      prev.map((row) => (row.tempId === tempId ? { ...row, code, isModified: !row.isNew } : row))
    );
  };

  const handleUpdateRowName = (tempId: string, langCode: string, name: string) => {
    setLocationRows((prev) =>
      prev.map((row) =>
        row.tempId === tempId
          ? { ...row, names: { ...row.names, [langCode]: name }, isModified: !row.isNew }
          : row
      )
    );
  };

  const handleDeleteRow = (tempId: string) => {
    const target = locationRows.find((r) => r.tempId === tempId);
    if (!target) return;
    if (target.guidfixed) {
      setDeletedLocationGuids((prev) => [...prev, target.guidfixed!]);
    }
    setLocationRows((prev) => prev.filter((r) => r.tempId !== tempId));
  };

  const namesToNameX = (names: Record<string, string>): NameX[] =>
    editorLanguages
      .map((code) => ({ code, name: names[code]?.trim() || "" }))
      .filter((item) => item.name !== "");

  // BATCH SAVE: Save all rows at once
  const handleSaveAllLocations = async () => {
    if (!auth || !mainApiUrl || !activeWarehouse?.guidfixed) return;
    const whGuid = activeWarehouse.guidfixed;
    setTableError("");

    // Filter out rows that are completely empty
    const activeRows = locationRows.filter(
      (r) => r.code.trim() || Object.values(r.names).some((v) => v.trim())
    );

    // Validation
    const codes = new Set<string>();
    for (let i = 0; i < activeRows.length; i++) {
      const r = activeRows[i];
      const code = r.code.trim();
      if (!code) {
        setTableError(
          tr("st_row_enter_storage_loc_code", "แถวที่ {0}: กรุณากรอกรหัสที่เก็บสินค้า").replace("{0}", String(i + 1))
        );
        return;
      }
      if (codes.has(code.toLowerCase())) {
        setTableError(
          tr("st_duplicate_storage_loc_code", "รหัสที่เก็บสินค้าซ้ำกันในตาราง: \"{0}\"").replace("{0}", String(code))
        );
        return;
      }
      codes.add(code.toLowerCase());

      const namesArray = namesToNameX(r.names);
      if (namesArray.length === 0) {
        setTableError(
          tr("st_row_enter_storage_loc_name", "แถวที่ {0} ({1}): กรุณากรอกชื่อที่เก็บสินค้า").replace("{0}", String(i + 1)).replace("{1}", String(code))
        );
        return;
      }
    }

    const toCreate = activeRows.filter((r) => r.isNew);
    const toUpdate = activeRows.filter((r) => !r.isNew && r.isModified);
    const toDelete = deletedLocationGuids;

    if (toCreate.length === 0 && toUpdate.length === 0 && toDelete.length === 0) {
      return;
    }

    setIsSavingLocations(true);
    try {
      // 1. Process deletions in parallel
      // Each step commits on the server immediately, so local state is updated
      // per row: a failure mid-batch must not re-send work that already succeeded.
      if (toDelete.length > 0) {
        const deleteResults = await Promise.allSettled(
          toDelete.map(async (guid) => {
            const res = await authFetch(`${mainApiUrl}/warehouse/${whGuid}/location/${encodeURIComponent(guid)}`, {
              method: "DELETE",
              headers: { Authorization: `Bearer ${auth.token}` },
            });
            const json = (await res.json()) as ApiResponse;
            if (!res.ok || json.success === false) {
              throw new Error(json.message || "Failed to delete location");
            }
            return guid;
          })
        );

        const successfulGuids = new Set<string>();
        let firstDeleteError: Error | null = null;
        for (const res of deleteResults) {
          if (res.status === "fulfilled") {
            successfulGuids.add(res.value);
          } else if (!firstDeleteError) {
            firstDeleteError = res.reason instanceof Error ? res.reason : new Error(String(res.reason));
          }
        }

        if (successfulGuids.size > 0) {
          setDeletedLocationGuids((prev) => prev.filter((g) => !successfulGuids.has(g)));
        }

        if (firstDeleteError) {
          throw firstDeleteError;
        }
      }

      // 2. Process creations in parallel
      if (toCreate.length > 0) {
        const createResults = await Promise.allSettled(
          toCreate.map(async (r) => {
            const payload = {
              code: r.code.trim(),
              names: namesToNameX(r.names),
              status: "active",
            };
            const res = await authFetch(`${mainApiUrl}/warehouse/${whGuid}/location`, {
              method: "POST",
              headers: { "Content-Type": "application/json", Authorization: `Bearer ${auth.token}` },
              body: JSON.stringify(payload),
            });
            const json = (await res.json()) as ApiResponse;
            if (!res.ok || json.success === false) {
              throw new Error(json.message || tr("st_cannot_create_storage", "ไม่สามารถสร้างที่เก็บ \"{0}\"").replace("{0}", String(r.code)));
            }
            return { tempId: r.tempId, guidfixed: json.id || r.guidfixed };
          })
        );

        const successfulCreated = new Map<string, string | undefined>();
        let firstCreateError: Error | null = null;
        for (const res of createResults) {
          if (res.status === "fulfilled") {
            successfulCreated.set(res.value.tempId, res.value.guidfixed);
          } else if (!firstCreateError) {
            firstCreateError = res.reason instanceof Error ? res.reason : new Error(String(res.reason));
          }
        }

        if (successfulCreated.size > 0) {
          setLocationRows((prev) =>
            prev.map((row) => {
              if (successfulCreated.has(row.tempId)) {
                const newGuid = successfulCreated.get(row.tempId);
                return { ...row, guidfixed: newGuid || row.guidfixed, isNew: false, isModified: false };
              }
              return row;
            })
          );
        }

        if (firstCreateError) {
          throw firstCreateError;
        }
      }

      // 3. Process updates in parallel
      const validToUpdate = toUpdate.filter((r) => !!r.guidfixed);
      if (validToUpdate.length > 0) {
        const updateResults = await Promise.allSettled(
          validToUpdate.map(async (r) => {
            // UpdateLocation replaces the embedded document, so fields this table
            // cannot edit (companyguids, allow*/blocked*, sortcode) must be carried over.
            const { guidfixed: _guid, ...existing } =
              activeWarehouse?.locations?.find((l) => l.guidfixed === r.guidfixed) ?? {};
            const payload = {
              ...existing,
              code: r.code.trim(),
              names: namesToNameX(r.names),
              status: "active",
            };
            const res = await authFetch(`${mainApiUrl}/warehouse/${whGuid}/location/${r.guidfixed}`, {
              method: "PUT",
              headers: { "Content-Type": "application/json", Authorization: `Bearer ${auth.token}` },
              body: JSON.stringify(payload),
            });
            const json = (await res.json()) as ApiResponse;
            if (!res.ok || json.success === false) {
              throw new Error(json.message || tr("st_cannot_edit_storage", "ไม่สามารถแก้ไขที่เก็บ \"{0}\"").replace("{0}", String(r.code)));
            }
            return r.tempId;
          })
        );

        const successfulUpdatedTempIds = new Set<string>();
        let firstUpdateError: Error | null = null;
        for (const res of updateResults) {
          if (res.status === "fulfilled") {
            successfulUpdatedTempIds.add(res.value);
          } else if (!firstUpdateError) {
            firstUpdateError = res.reason instanceof Error ? res.reason : new Error(String(res.reason));
          }
        }

        if (successfulUpdatedTempIds.size > 0) {
          setLocationRows((prev) =>
            prev.map((row) => (successfulUpdatedTempIds.has(row.tempId) ? { ...row, isModified: false } : row))
          );
        }

        if (firstUpdateError) {
          throw firstUpdateError;
        }
      }


      await loadTree();
      setDeletedLocationGuids([]);
      setSuccessMessage(tr("st_all_storage_loc_saved", "บันทึกที่เก็บสินค้าทั้งหมดเรียบร้อยแล้ว"));
      setTimeout(() => setSuccessMessage(""), 3000);
      if (onRefresh) onRefresh();
    } catch (err) {
      setTableError(errorMessage(err, tr("st_error_saving_storage", "เกิดข้อผิดพลาดในการบันทึกที่เก็บสินค้า")));
    } finally {
      setIsSavingLocations(false);
    }
  };

  // Reset table changes
  const handleResetTable = () => {
    if (activeWarehouse) {
      setLocationRows((activeWarehouse.locations || []).map((l) => toEditableRow(l)));
      setDeletedLocationGuids([]);
      setTableError("");
    }
  };

  // Filtered rows for display
  const filteredLocationRows = useMemo(() => {
    if (!locationSearchQuery.trim()) return locationRows;
    const needle = locationSearchQuery.toLowerCase();
    return locationRows.filter((r) => {
      const codeMatch = r.code.toLowerCase().includes(needle);
      const nameMatch = Object.values(r.names).some((v) => v.toLowerCase().includes(needle));
      return codeMatch || nameMatch;
    });
  }, [locationRows, locationSearchQuery]);

  const hasUnsavedChanges = useMemo(
    () => locationRows.some((r) => r.isNew || r.isModified) || deletedLocationGuids.length > 0,
    [locationRows, deletedLocationGuids]
  );

  // Filtered warehouses for Left Pane
  const filteredWarehouses = useMemo(() => {
    if (!warehouseSearchQuery.trim()) return tree;
    const needle = warehouseSearchQuery.toLowerCase();
    return tree.filter((w) => {
      const codeMatch = w.code?.toLowerCase().includes(needle);
      const nameMatch = displayName(w.names, language).toLowerCase().includes(needle);
      return codeMatch || nameMatch;
    });
  }, [tree, warehouseSearchQuery, language]);

  return (
    <div
      ref={containerRef}
      style={{ "--warehouse-sidebar-width": `${sidebarWidth}px` } as React.CSSProperties}
      className="relative grid w-full min-w-0 items-stretch gap-2 min-h-[calc(100dvh-12rem)] grid-cols-1 lg:grid-cols-[var(--warehouse-sidebar-width)_auto_minmax(0,1fr)]"
    >
      {/* LEFT PANE: Warehouses List (คลังสินค้า) */}
      <Card className="flex h-full min-h-0 flex-col overflow-hidden border-border bg-card shadow-sm">
        <div className="flex items-center justify-between border-b border-border/40 p-2.5 bg-secondary/5 shrink-0">
          <div className="flex items-center gap-1.5 font-bold text-xs text-foreground uppercase tracking-wider">
            <Warehouse className="size-4 text-primary" />
            <span>{tr("form_design_cat_inventory", "คลังสินค้า")}</span>
            <span className="rounded-full bg-primary/10 px-1.5 py-0.2 text-[10px] font-bold text-primary">
              {tree.length}
            </span>
          </div>
          <Button
            type="button"
            size="sm"
            className="h-8 shrink-0 rounded-lg gap-1 bg-primary text-primary-foreground hover:bg-primary/95 font-semibold text-xs"
            onClick={openCreateWarehouse}
            disabled={loading}
          >
            <Plus className="size-3.5" />
            {tr("st_add_warehouse", "เพิ่มคลังสินค้า")}
          </Button>
        </div>

        {tree.length > 3 && (
          <div className="p-2 border-b border-border/30 bg-secondary/5 shrink-0">
            <div className="relative">
              <Input
                value={warehouseSearchQuery}
                onChange={(e) => setWarehouseSearchQuery(e.target.value)}
                placeholder={tr("st_search_warehouse", "ค้นหาคลัง...")}
                className="h-7 !pl-8 pr-2 text-xs rounded-md"
              />
              <span className="pointer-events-none absolute left-2.5 top-1/2 -translate-y-1/2 text-muted-foreground/60">
                <Search className="size-3" />
              </span>
            </div>
          </div>
        )}

        <CardContent className="flex min-h-0 flex-1 flex-col p-0 overflow-y-auto overscroll-contain">
          {loading ? (
            <div className="flex h-full min-h-24 items-center justify-center gap-2 p-6 text-sm text-muted-foreground">
              <Loader2 className="animate-spin size-5 text-primary" />
              {tr("st_loading", "กำลังโหลด...")}
            </div>
          ) : loadError ? (
            <div className="flex h-full min-h-24 flex-col items-center justify-center gap-2 p-4 text-center text-sm">
              <AlertTriangle className="size-6 text-destructive" />
              <span className="font-medium text-destructive">{loadError}</span>
              <Button type="button" size="sm" variant="outline" onClick={() => void loadTree()}>
                {tr("export_report_retry", "ลองใหม่")}
              </Button>
            </div>
          ) : filteredWarehouses.length === 0 ? (
            <div className="flex h-full min-h-24 flex-col items-center justify-center gap-1.5 p-4 text-center text-xs text-muted-foreground">
              <span className="font-medium">
                {tr("st_no_warehouse_data_found", "ไม่พบข้อมูลคลังสินค้า")}
              </span>
            </div>
          ) : (
            <div className="flex flex-col">
              {filteredWarehouses.map((w, warehouseIdx) => {
                const warehouseId = w.guidfixed || "";
                const isWarehouseSelected = selectedWarehouseId === warehouseId;
                const locCount = w.locations?.length || 0;
                const warehouseTitle = displayName(w.names, language) || w.code || "-";

                return (
                  <div
                    key={warehouseId}
                    className={cn(
                      "group/wh relative flex items-center justify-between border-b border-border/40 py-2.5 px-3 transition-colors cursor-pointer",
                      isWarehouseSelected ? "bg-primary/10 text-primary font-semibold" : "hover:bg-muted/40"
                    )}
                    onClick={async () => {
                      if (hasUnsavedChanges) {
                        const proceed = await confirm({ title: tr("st_warning", "เตือน"), description: tr("st_unsaved_data_change_warehouse", "มีข้อมูลที่ยังไม่ได้บันทึก ต้องการเปลี่ยนคลังหรือไม่?") });
                        if (!proceed) {
                          return;
                        }
                      }
                      setSelectedWarehouseId(warehouseId);
                    }}
                  >
                    {isWarehouseSelected && (
                      <span className="pointer-events-none absolute bottom-0 left-0 top-0 w-1 bg-primary" />
                    )}
                    <div className="flex min-w-0 flex-1 flex-col gap-1 pr-2">
                      <div className="flex items-center gap-1.5">
                        <span className="text-xs font-bold text-muted-foreground shrink-0">{warehouseIdx + 1}.</span>
                        <span className="inline-flex items-center rounded border border-primary/20 bg-primary/10 px-1.5 py-0.2 text-[11px] font-mono font-bold text-primary shrink-0">
                          {w.code}
                        </span>
                        <span className="min-w-0 flex-1 truncate text-xs font-bold text-foreground">
                          {warehouseTitle}
                        </span>
                      </div>
                      <div className="flex items-center gap-2 pl-4 text-[11px] text-muted-foreground">
                        <span className="shrink-0 rounded-full border border-primary/20 bg-primary/5 px-2 py-0.2 font-medium text-primary">
                          {tr("st_subcategory_placeholder", "ลูก {0}").replace("{0}", String(locCount))}
                        </span>
                      </div>
                    </div>

                    <div className="flex shrink-0 items-center gap-1 opacity-70 group-hover/wh:opacity-100 transition-opacity">
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="size-7 rounded-full text-emerald-600 hover:text-emerald-700 hover:bg-emerald-50 dark:hover:bg-emerald-950/40"
                        title={tr("st_add_storage_location", "เพิ่มที่เก็บสินค้า")}
                        onClick={(e) => {
                          e.stopPropagation();
                          setSelectedWarehouseId(warehouseId);
                          handleAddRow();
                        }}
                      >
                        <Boxes className="size-3.5" />
                      </Button>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="size-7 rounded-full text-blue-600 hover:text-blue-700 hover:bg-blue-50 dark:hover:bg-blue-950/40"
                        title={tr("st_edit_warehouse", "แก้ไขคลังสินค้า")}
                        onClick={(e) => {
                          e.stopPropagation();
                          openEditWarehouse(warehouseId);
                        }}
                      >
                        <Edit3 className="size-3.5" />
                      </Button>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        className="size-7 rounded-full text-destructive hover:bg-destructive/10"
                        title={tr("st_delete_warehouse", "ลบคลังสินค้า")}
                        onClick={(e) => handleDeleteWarehouse(warehouseId, e)}
                      >
                        <Trash2 className="size-3.5" />
                      </Button>
                    </div>
                  </div>
                );
              })}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Draggable vertical splitter for resizing sidebar */}
      <ResizableSplitter
        value={sidebarWidth}
        min={WAREHOUSE_SIDEBAR_MIN_WIDTH}
        max={WAREHOUSE_SIDEBAR_MAX_WIDTH}
        label={
          tr("st_adjust_warehouse_list_width", "ปรับขนาดความกว้างรายชื่อคลัง (ลากเพื่อปรับ, ดับเบิ้ลคลิกเพื่อรีเซ็ต)")
        }
        isResizing={isResizingSidebar}
        onPointerDown={handleSidebarResizeStart}
        onDoubleClick={handleSidebarResizeReset}
        onKeyDown={handleSidebarKeyDown}
        breakpoint="lg"
      />

      {/* RIGHT PANE: Editable Locations Table (ตารางที่เก็บสินค้า เซฟพร้อมกันทีเดียว) */}
      <Card className="flex h-full min-h-0 flex-col overflow-hidden border-border bg-card shadow-sm">
        {/* Header Bar */}
        <div className="flex flex-wrap items-center justify-between gap-2 border-b border-border/40 p-2.5 bg-secondary/5 shrink-0">
          <div className="flex min-w-0 flex-1 items-center gap-2">
            <Boxes className="size-4 text-primary shrink-0" />
            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2 flex-wrap">
                <h2 className="text-xs font-bold text-foreground truncate">
                  {activeWarehouse
                    ? `${activeWarehouse.code} - ${displayName(activeWarehouse.names, language)}`
                    : tr("st_no_warehouse_selected", "ยังไม่ได้เลือกคลังสินค้า")}
                </h2>
                {activeWarehouse && (
                  <div className="flex items-center gap-1.5 shrink-0">
                    <span className="rounded-full bg-primary/10 border border-primary/20 px-2 py-0.5 text-[11px] font-semibold text-primary">
                      {tr("st_storage_location_count", "{0} ที่เก็บสินค้า").replace("{0}", String(locationRows.length))}
                    </span>
                    {hasUnsavedChanges && (
                      <span className="rounded-full bg-amber-500/10 border border-amber-500/30 px-2 py-0.5 text-[11px] font-semibold text-amber-700 dark:text-amber-300 flex items-center gap-1">
                        <AlertTriangle className="size-3" />
                        {tr("st_not_saved_yet", "ยังไม่ได้บันทึก")}
                      </span>
                    )}
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Action Buttons */}
          {activeWarehouse && (
            <div className="flex items-center gap-2 shrink-0">
              {hasUnsavedChanges && (
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="h-8 gap-1 text-xs font-semibold"
                  onClick={handleResetTable}
                  disabled={isSavingLocations}
                >
                  <RotateCcw className="size-3.5" />
                  {tr("st_reset_to_default", "คืนค่าเดิม")}
                </Button>
              )}
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="h-8 gap-1 text-xs font-semibold border-primary/30 text-primary hover:bg-primary/5"
                title={tr("st_add_storage_location", "เพิ่มที่เก็บสินค้า")}
                onClick={handleAddRow}
                disabled={isSavingLocations}
              >
                <Plus className="size-3.5" />
                {tr("bill_add_row", "เพิ่มแถว")}
              </Button>
              <Button
                type="button"
                size="sm"
                className="h-8 gap-1.5 bg-primary text-primary-foreground hover:bg-primary/95 font-semibold text-xs shadow-sm"
                onClick={handleSaveAllLocations}
                disabled={isSavingLocations || !hasUnsavedChanges}
              >
                {isSavingLocations ? <Loader2 className="size-3.5 animate-spin" /> : <Save className="size-3.5" />}
                {tr("st_save_all", "บันทึกทั้งหมด")}
              </Button>
            </div>
          )}
        </div>

        {/* Toolbar: Search filter */}
        {activeWarehouse && (
          <div className="flex items-center gap-2 border-b border-border/30 p-2 bg-secondary/5 shrink-0">
            <div className="relative flex-1">
              <Input
                value={locationSearchQuery}
                onChange={(e) => setLocationSearchQuery(e.target.value)}
                placeholder={
                  tr("st_search_storage_location_table", "ค้นหาในตารางที่เก็บสินค้า...")
                }
                className="h-8 !pl-10 pr-8 text-xs rounded-lg"
              />
              <span className="pointer-events-none absolute left-3 top-1/2 -translate-y-1/2 text-muted-foreground/60">
                <Search className="size-3.5" />
              </span>
              {locationSearchQuery && (
                <button
                  type="button"
                  onClick={() => setLocationSearchQuery("")}
                  className="absolute right-2.5 top-1/2 -translate-y-1/2 text-muted-foreground hover:text-foreground p-0.5 rounded"
                >
                  <X className="size-3" />
                </button>
              )}
            </div>
          </div>
        )}

        {/* Feedback Alerts */}
        {tableError && (
          <div className="px-3 py-2 bg-destructive/10 border-b border-destructive/20 text-destructive text-xs font-medium flex items-center gap-2 shrink-0">
            <AlertTriangle className="size-4 shrink-0" />
            <span>{tableError}</span>
          </div>
        )}
        {successMessage && (
          <div className="px-3 py-2 bg-emerald-500/10 border-b border-emerald-500/20 text-emerald-700 dark:text-emerald-300 text-xs font-medium flex items-center gap-2 shrink-0">
            <CheckCircle2 className="size-4 shrink-0" />
            <span>{successMessage}</span>
          </div>
        )}

        {/* Table Container */}
        <div className="flex min-h-0 flex-1 flex-col overflow-hidden">
          {!activeWarehouse ? (
            <div className="flex h-full min-h-40 flex-col items-center justify-center gap-2 p-6 text-center text-xs text-muted-foreground">
              <Warehouse className="size-8 text-muted-foreground/40 animate-pulse" />
              <span>{tr("st_select_or_create_warehouse_left", "กรุณาเลือกหรือสร้างคลังสินค้าทางด้านซ้าย")}</span>
            </div>
          ) : locationRows.length === 0 ? (
            <div className="flex h-full min-h-40 flex-col items-center justify-center gap-2 p-6 text-center text-xs text-muted-foreground">
              <Boxes className="size-8 text-muted-foreground/40" />
              <span className="font-semibold text-foreground">
                {tr("st_no_storage_locations_in_warehouse", "ยังไม่มีที่เก็บสินค้าในคลังนี้")}
              </span>
              <Button
                type="button"
                size="sm"
                onClick={handleAddRow}
                className="gap-1.5 h-8 text-xs font-semibold bg-primary text-primary-foreground"
              >
                <Plus className="size-3.5" />
                {tr("st_add_first_row", "เพิ่มแถวแรก")}
              </Button>
            </div>
          ) : (
            <div className="min-h-0 flex-1 overflow-auto overscroll-contain">
              <table className="w-full text-left text-xs border-collapse">
                <thead className="sticky top-0 z-10 bg-muted/90 backdrop-blur border-b border-border shadow-2xs">
                  <tr className="text-[11px] font-bold text-muted-foreground uppercase tracking-wider">
                    <th className="py-2.5 px-3 w-14 text-center shrink-0">#</th>
                    <th className="py-2.5 px-3 w-48 shrink-0">
                      {tr("location_code", "รหัสที่เก็บสินค้า")} <span className="text-destructive">*</span>
                    </th>
                    <th className="py-2.5 px-3 min-w-[200px]">
                      {tr("st_storage_name_th", "ชื่อที่เก็บสินค้า (ไทย)")} <span className="text-destructive">*</span>
                    </th>
                    {editorLanguages.includes("en") && (
                      <th className="py-2.5 px-3 min-w-[200px]">
                        {tr("st_storage_name_en", "ชื่อที่เก็บสินค้า (EN)")}
                      </th>
                    )}
                    <th className="py-2.5 px-3 w-24 text-center shrink-0">
                      {tr("manage", "จัดการ")}
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/30 bg-card">
                  {filteredLocationRows.map((row, rowIdx) => {
                    const rowText = `${row.code} - ${row.names.th || row.names.en || ""}`;

                    return (
                      <tr
                        key={row.tempId}
                        className={cn(
                          "transition-colors hover:bg-muted/40 group/row",
                          row.isNew && "border-l-2 border-l-emerald-500 bg-emerald-500/[0.02]",
                          row.isModified && !row.isNew && "border-l-2 border-l-amber-500 bg-amber-500/[0.02]"
                        )}
                      >
                        {/* 1. Index */}
                        <td className="py-2 px-3 text-center text-muted-foreground font-mono font-medium whitespace-nowrap">
                          <div className="flex items-center justify-center gap-1">
                            <span>{rowIdx + 1}</span>
                            {row.isNew && (
                              <span className="rounded bg-emerald-500/15 px-1 py-0.2 text-[9px] font-bold text-emerald-700 dark:text-emerald-300">
                                NEW
                              </span>
                            )}
                            {row.isModified && !row.isNew && (
                              <span className="rounded bg-amber-500/15 px-1 py-0.2 text-[9px] font-bold text-amber-700 dark:text-amber-300">
                                MOD
                              </span>
                            )}
                          </div>
                        </td>

                        {/* 2. Code Input */}
                        <td className="py-2 px-3">
                          <Input
                            value={row.code}
                            onChange={(e) => handleUpdateRowCode(row.tempId, e.target.value)}
                            placeholder="e.g. ZONE-A"
                            className="h-8 text-xs font-mono font-bold uppercase rounded-md border-border/70 bg-background/80 hover:border-border hover:bg-background focus:border-primary focus:bg-background"
                          />
                        </td>

                        {/* 3. Thai Name Input */}
                        <td className="py-2 px-3">
                          <Input
                            value={row.names.th || ""}
                            onChange={(e) => handleUpdateRowName(row.tempId, "th", e.target.value)}
                            placeholder={tr("st_storage_name_th", "ชื่อที่เก็บสินค้า (ไทย)")}
                            className="h-8 text-xs rounded-md border-border/70 bg-background/80 hover:border-border hover:bg-background focus:border-primary focus:bg-background"
                          />
                        </td>

                        {/* 4. English Name Input (if EN active) */}
                        {editorLanguages.includes("en") && (
                          <td className="py-2 px-3">
                            <Input
                              value={row.names.en || ""}
                              onChange={(e) => handleUpdateRowName(row.tempId, "en", e.target.value)}
                              placeholder="Location Name (EN)"
                              className="h-8 text-xs rounded-md border-border/70 bg-background/80 hover:border-border hover:bg-background focus:border-primary focus:bg-background"
                            />
                          </td>
                        )}

                        {/* 5. Actions Column */}
                        <td className="py-2 px-3 text-center">
                          <div className="flex items-center justify-center gap-1">
                            {/* Hidden row text container for E2E text locator matching */}
                            <span className="sr-only">{rowText}</span>
                            <Button
                              type="button"
                              variant="ghost"
                              size="icon"
                              className="size-7 rounded-full text-destructive hover:bg-destructive/10"
                              title={tr("st_delete_storage_location", "ลบที่เก็บสินค้า")}
                              onClick={() => handleDeleteRow(row.tempId)}
                            >
                              <Trash2 className="size-3.5" />
                            </Button>
                          </div>
                        </td>
                      </tr>
                    );
                  })}
                </tbody>
              </table>
            </div>
          )}

          {/* Pinned Bottom Footer (Always visible outside scroll container) */}
          {activeWarehouse && (
            <div className="shrink-0 border-t border-border/40 bg-secondary/5 px-3 py-2 flex items-center justify-between gap-2">
              <div className="flex items-center gap-2 text-xs text-muted-foreground">
                <span className="font-semibold text-foreground">
                  {tr("st_total_items", "รวม {0} รายการ").replace("{0}", String(locationRows.length))}
                </span>
                {hasUnsavedChanges && (
                  <span className="text-amber-600 dark:text-amber-400 font-medium">
                    ({locationRows.filter((r) => r.isNew).length > 0 && `${locationRows.filter((r) => r.isNew).length} ${tr("st_new_row", "แถวใหม่")}`}
                    {locationRows.filter((r) => r.isNew).length > 0 && locationRows.filter((r) => r.isModified && !r.isNew).length > 0 && ", "}
                    {locationRows.filter((r) => r.isModified && !r.isNew).length > 0 && `${locationRows.filter((r) => r.isModified && !r.isNew).length} ${tr("st_edit_row", "แถวแก้ไข")}`})
                  </span>
                )}
              </div>

              <div className="flex items-center gap-2">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  className="h-8 gap-1 text-xs font-semibold border-primary/30 text-primary hover:bg-primary/5"
                  title={tr("st_add_storage_location", "เพิ่มที่เก็บสินค้า")}
                  onClick={handleAddRow}
                  disabled={isSavingLocations}
                >
                  <Plus className="size-3.5" />
                  {tr("bill_add_row", "เพิ่มแถว")}
                </Button>

                {hasUnsavedChanges && (
                  <Button
                    type="button"
                    size="sm"
                    className="h-8 gap-1.5 bg-primary text-primary-foreground hover:bg-primary/95 font-semibold text-xs shadow-sm"
                    onClick={handleSaveAllLocations}
                    disabled={isSavingLocations}
                  >
                    {isSavingLocations ? <Loader2 className="size-3.5 animate-spin" /> : <Save className="size-3.5" />}
                    {tr("st_save_all", "บันทึกทั้งหมด")}
                  </Button>
                )}
              </div>
            </div>
          )}
        </div>
      </Card>

      {/* DIALOG 1: Warehouse Create/Edit Modal */}
      {isWarehouseDialogOpen && (
        <div className="dialog-backdrop" role="dialog" aria-modal="true">
          <div className="w-full max-w-md rounded-2xl border border-border bg-card p-4 text-card-foreground shadow-xl">
            <div className="flex items-center justify-between border-b pb-2">
              <div className="text-sm font-bold flex items-center gap-1.5">
                <Warehouse className="size-4 text-primary" />
                {editingWarehouseId
                  ? tr("st_edit_warehouse", "แก้ไขคลังสินค้า")
                  : tr("st_add_warehouse", "เพิ่มคลังสินค้า")}
              </div>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="size-7"
                onClick={() => setIsWarehouseDialogOpen(false)}
                aria-label={tr("bill_close", "ปิด")}
              >
                <X className="size-4" />
              </Button>
            </div>

            <form id="warehouse-modal-form" onSubmit={handleSaveWarehouse} className="flex flex-col gap-3 mt-3">
              <div className="flex flex-col gap-1">
                <label className="text-[11px] font-bold text-muted-foreground uppercase">
                  {tr("st_warehouse_code", "รหัสคลังสินค้า")} <span className="text-destructive">*</span>
                </label>
                <Input
                  value={warehouseForm.code}
                  onChange={(e) => setWarehouseForm((prev) => ({ ...prev, code: e.target.value }))}
                  placeholder="e.g. 00000"
                  className="h-9 text-xs"
                />
              </div>

              <div className="grid grid-cols-2 gap-2">
                <div className="flex flex-col gap-1">
                  <label className="text-[11px] font-bold text-muted-foreground uppercase">
                    {tr("company_latitude", "ละติจูด")}
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
                    {tr("st_longitude", "ลองจิจูด")}
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
                className="h-8 gap-1 text-xs font-semibold text-primary border-primary/30"
              >
                <MapPin className="size-3.5" />
                {tr("st_select_location_from_map", "เลือกตำแหน่งจากแผนที่")}
              </Button>

              <div className="border-t border-border/30 pt-2">
                <NamesEditor
                  names={editorLanguages.map((l) => ({ code: l, name: warehouseForm.names[l] || "" }))}
                  onChange={(next) =>
                    setWarehouseForm((prev) => ({
                      ...prev,
                      names: next.reduce((acc, n) => ({ ...acc, [n.code || ""]: n.name || "" }), {} as Record<string, string>),
                    }))
                  }
                  languages={editorLanguages}
                  label={tr("st_warehouse_name_multilingual", "ชื่อคลังสินค้าหลายภาษา")}
                  firstRequired
                  language={language}
                />
              </div>

              {warehouseFormError && (
                <div className="text-xs font-medium text-destructive">{warehouseFormError}</div>
              )}

              <div className="mt-3 flex items-center justify-end gap-2 border-t pt-2">
                <Button
                  type="button"
                  variant="outline"
                  size="sm"
                  onClick={() => setIsWarehouseDialogOpen(false)}
                  disabled={isSavingWarehouse}
                  className="h-8 text-xs font-semibold"
                >
                  {tr("cancel", "ยกเลิก")}
                </Button>
                <Button
                  type="submit"
                  size="sm"
                  disabled={isSavingWarehouse}
                  className="h-8 gap-1 bg-primary text-xs font-semibold text-primary-foreground"
                >
                  {isSavingWarehouse ? <Loader2 className="size-3.5 animate-spin" /> : <Save className="size-3.5" />}
                  {tr("fd_save", "บันทึก")}
                </Button>
              </div>
            </form>
          </div>
        </div>
      )}

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
