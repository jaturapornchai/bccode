"use client";

import { useMemo, useState } from "react";
import {
  MapPin,
  Plus,
  ChevronRight,
  ChevronDown,
  Edit3,
  Trash2,
  X,
  Layers,
  GripVertical,
  ArrowUp,
  ArrowDown,
  FolderPlus,
  Scale,
  Ruler,
  Package,
} from "lucide-react";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { LANGUAGES, type LanguageCode } from "@/lib/i18n";
import { normalizeLanguageConfigs } from "./system-settings-screen";
import { cn } from "@/lib/utils";

export function WarehouseLocationsEditor({
  form,
  setForm,
  language,
  workspace,
  readOnly = false,
}: {
  form: any;
  setForm?: (form: any) => void;
  language: LanguageCode;
  workspace: any;
  readOnly?: boolean;
}) {
  const isEdit = !readOnly && typeof setForm === "function";

  const locationsList = useMemo<any[]>(() => {
    const raw = form?.location;
    if (Array.isArray(raw)) return raw;
    if (typeof raw === "string" && raw.trim()) {
      try {
        const parsed = JSON.parse(raw);
        if (Array.isArray(parsed)) return parsed;
      } catch (e) {}
    }
    return [];
  }, [form?.location]);

  // Collapsed locations state
  const [collapsedLocs, setCollapsedLocs] = useState<Record<string, boolean>>({});

  const toggleLocation = (code: string) => {
    setCollapsedLocs((prev) => ({ ...prev, [code]: !prev[code] }));
  };

  // Active languages for name editing
  const editorLanguages = useMemo(() => {
    if (!workspace) return ["th", "en"];
    const configs = workspace.shopInfo?.settings?.languageconfigs || [];
    const defaultCode = workspace.shopInfo?.settings?.language || "th";
    return normalizeLanguageConfigs(configs, defaultCode).map((row) => row.code);
  }, [workspace]);

  // Dialog state for Location
  const [locDialog, setLocDialog] = useState<{
    open: boolean;
    mode: "create" | "edit";
    index?: number;
    code: string;
    names: Record<string, string>;
    max_weight: number | string;
    width: number | string;
    length: number | string;
    height: number | string;
    suitable_product_types: string;
    error?: string;
  }>({
    open: false,
    mode: "create",
    code: "",
    names: {},
    max_weight: "",
    width: "",
    length: "",
    height: "",
    suitable_product_types: "",
  });

  // Dialog state for Shelf
  const [shelfDialog, setShelfDialog] = useState<{
    open: boolean;
    mode: "create" | "edit";
    locIndex?: number;
    shelfIndex?: number;
    code: string;
    name: string;
    max_weight: number | string;
    width: number | string;
    length: number | string;
    height: number | string;
    suitable_product_types: string;
    error?: string;
  }>({
    open: false,
    mode: "create",
    code: "",
    name: "",
    max_weight: "",
    width: "",
    length: "",
    height: "",
    suitable_product_types: "",
  });

  // Dialog state for Bulk Shelf Addition
  const [bulkShelfDialog, setBulkShelfDialog] = useState<{
    open: boolean;
    locIndex?: number;
    prefix: string;
    startNum: number | string;
    endNum: number | string;
    padding: number | string;
    namePattern: string;
    max_weight: number | string;
    width: number | string;
    length: number | string;
    height: number | string;
    suitable_product_types: string;
    error?: string;
  }>({
    open: false,
    prefix: "SH-",
    startNum: 1,
    endNum: 10,
    padding: 2,
    namePattern: language === "th" ? "ชั้นวาง {number}" : "Shelf {number}",
    max_weight: "",
    width: "",
    length: "",
    height: "",
    suitable_product_types: "",
  });

  const getLanguageName = (code: string) => {
    const found = LANGUAGES.find((item) => item.code === code);
    return found ? found.name : code.toUpperCase();
  };

  // Helper to get localized location name
  const displayLocationName = (names: any) => {
    if (Array.isArray(names)) {
      const found = names.find((item) => item.code === language);
      if (found?.name) return found.name;
      const th = names.find((item) => item.code === "th");
      if (th?.name) return th.name;
      const en = names.find((item) => item.code === "en");
      if (en?.name) return en.name;
      if (names.length > 0) return names[0].name || "";
    } else if (names && typeof names === "object") {
      return names[language] || names.th || names.en || "";
    }
    return "";
  };

  // Handlers for Location
  const openAddLocation = () => {
    const initialNames: Record<string, string> = {};
    editorLanguages.forEach((lang) => {
      initialNames[lang] = "";
    });
    setLocDialog({
      open: true,
      mode: "create",
      code: "",
      names: initialNames,
      max_weight: "",
      width: "",
      length: "",
      height: "",
      suitable_product_types: "",
    });
  };

  const openEditLocation = (index: number, loc: any) => {
    const initialNames: Record<string, string> = {};
    editorLanguages.forEach((lang) => {
      let val = "";
      if (Array.isArray(loc.names)) {
        val = loc.names.find((n: any) => n.code === lang)?.name || "";
      } else if (loc.names && typeof loc.names === "object") {
        val = loc.names[lang] || "";
      }
      initialNames[lang] = val;
    });
    setLocDialog({
      open: true,
      mode: "edit",
      index,
      code: loc.code || "",
      names: initialNames,
      max_weight: loc.max_weight !== undefined && loc.max_weight !== 0 ? loc.max_weight : "",
      width: loc.width !== undefined && loc.width !== 0 ? loc.width : "",
      length: loc.length !== undefined && loc.length !== 0 ? loc.length : "",
      height: loc.height !== undefined && loc.height !== 0 ? loc.height : "",
      suitable_product_types: loc.suitable_product_types || "",
    });
  };

  const deleteLocation = (index: number) => {
    if (!setForm) return;
    const confirmed = window.confirm(
      language === "th"
        ? `ต้องการลบโซนเก็บสินค้า "${displayLocationName(locationsList[index].names)}" ใช่หรือไม่?`
        : `Are you sure you want to delete storage zone "${displayLocationName(locationsList[index].names)}"?`
    );
    if (!confirmed) return;
    const updated = locationsList.filter((_, idx) => idx !== index);
    setForm({ ...form, location: updated });
  };

  const saveLocation = () => {
    if (!setForm) return;
    const code = locDialog.code.trim();
    if (!code) {
      setLocDialog((prev) => ({ ...prev, error: language === "th" ? "กรุณากรอกรหัสโซนเก็บสินค้า" : "Please fill in storage zone code" }));
      return;
    }

    // Check duplicate code
    const isDuplicate = locationsList.some((loc, idx) => {
      if (locDialog.mode === "edit" && locDialog.index === idx) return false;
      return loc.code?.toLowerCase() === code.toLowerCase();
    });

    if (isDuplicate) {
      setLocDialog((prev) => ({ ...prev, error: language === "th" ? "รหัสโซนเก็บสินค้านี้มีอยู่แล้ว" : "Storage zone code already exists" }));
      return;
    }

    // Validate at least one name
    const hasName = Object.values(locDialog.names).some((name) => name.trim());
    if (!hasName) {
      setLocDialog((prev) => ({ ...prev, error: language === "th" ? "กรุณากรอกชื่ออย่างน้อยหนึ่งภาษา" : "Please fill in at least one language name" }));
      return;
    }

    const namesArray = editorLanguages.map((lang) => ({
      code: lang,
      name: locDialog.names[lang]?.trim() || "",
      isauto: false,
      isdelete: false,
    }));

    const maxWeight = parseFloat(String(locDialog.max_weight)) || 0;
    const w = parseFloat(String(locDialog.width)) || 0;
    const l = parseFloat(String(locDialog.length)) || 0;
    const h = parseFloat(String(locDialog.height)) || 0;
    const suitable = locDialog.suitable_product_types?.trim() || "";

    let updated = [...locationsList];
    if (locDialog.mode === "create") {
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
    } else if (locDialog.mode === "edit" && locDialog.index !== undefined) {
      updated[locDialog.index] = {
        ...updated[locDialog.index],
        code,
        names: namesArray,
        max_weight: maxWeight,
        width: w,
        length: l,
        height: h,
        suitable_product_types: suitable,
        shelf: updated[locDialog.index].shelf || [],
      };
    }

    setForm({ ...form, location: updated });
    setLocDialog((prev) => ({ ...prev, open: false }));
  };

  // Handlers for Shelf
  const openAddShelf = (locIndex: number) => {
    setShelfDialog({
      open: true,
      mode: "create",
      locIndex,
      code: "",
      name: "",
      max_weight: "",
      width: "",
      length: "",
      height: "",
      suitable_product_types: "",
    });
  };

  const openEditShelf = (locIndex: number, shelfIndex: number, shelf: any) => {
    setShelfDialog({
      open: true,
      mode: "edit",
      locIndex,
      shelfIndex,
      code: shelf.code || "",
      name: shelf.name || "",
      max_weight: shelf.max_weight !== undefined && shelf.max_weight !== 0 ? shelf.max_weight : "",
      width: shelf.width !== undefined && shelf.width !== 0 ? shelf.width : "",
      length: shelf.length !== undefined && shelf.length !== 0 ? shelf.length : "",
      height: shelf.height !== undefined && shelf.height !== 0 ? shelf.height : "",
      suitable_product_types: shelf.suitable_product_types || "",
    });
  };

  const deleteShelf = (locIndex: number, shelfIndex: number) => {
    if (!setForm) return;
    const loc = locationsList[locIndex];
    const shelf = loc.shelf[shelfIndex];
    const confirmed = window.confirm(
      language === "th"
        ? `ต้องการลบชั้นวาง "${shelf.name || shelf.code}" ใช่หรือไม่?`
        : `Are you sure you want to delete shelf "${shelf.name || shelf.code}"?`
    );
    if (!confirmed) return;

    const updatedShelves = loc.shelf.filter((_: any, idx: number) => idx !== shelfIndex);
    const updatedLocs = [...locationsList];
    updatedLocs[locIndex] = {
      ...loc,
      shelf: updatedShelves,
    };
    setForm({ ...form, location: updatedLocs });
  };

  const saveShelf = () => {
    if (!setForm || shelfDialog.locIndex === undefined) return;
    const code = shelfDialog.code.trim();
    const name = shelfDialog.name.trim();

    if (!code) {
      setShelfDialog((prev) => ({ ...prev, error: language === "th" ? "กรุณากรอกรหัสชั้นวาง" : "Please fill in shelf code" }));
      return;
    }
    if (!name) {
      setShelfDialog((prev) => ({ ...prev, error: language === "th" ? "กรุณากรอกชื่อชั้นวาง" : "Please fill in shelf name" }));
      return;
    }

    const loc = locationsList[shelfDialog.locIndex];
    const shelves = loc.shelf || [];

    // Check duplicate code
    const isDuplicate = shelves.some((shelf: any, idx: number) => {
      if (shelfDialog.mode === "edit" && shelfDialog.shelfIndex === idx) return false;
      return shelf.code?.toLowerCase() === code.toLowerCase();
    });

    if (isDuplicate) {
      setShelfDialog((prev) => ({ ...prev, error: language === "th" ? "รหัสชั้นวางนี้มีอยู่แล้วในโซนเก็บสินค้านี้" : "Shelf code already exists in this storage zone" }));
      return;
    }

    const maxWeight = parseFloat(String(shelfDialog.max_weight)) || 0;
    const w = parseFloat(String(shelfDialog.width)) || 0;
    const l = parseFloat(String(shelfDialog.length)) || 0;
    const h = parseFloat(String(shelfDialog.height)) || 0;
    const suitable = shelfDialog.suitable_product_types?.trim() || "";

    let updatedShelves = [...shelves];
    if (shelfDialog.mode === "create") {
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
    } else if (shelfDialog.mode === "edit" && shelfDialog.shelfIndex !== undefined) {
      updatedShelves[shelfDialog.shelfIndex] = {
        ...updatedShelves[shelfDialog.shelfIndex],
        code,
        name,
        max_weight: maxWeight,
        width: w,
        length: l,
        height: h,
        suitable_product_types: suitable,
        productitems: updatedShelves[shelfDialog.shelfIndex].productitems || [],
      };
    }

    const updatedLocs = [...locationsList];
    updatedLocs[shelfDialog.locIndex] = {
      ...loc,
      shelf: updatedShelves,
    };

    setForm({ ...form, location: updatedLocs });
    setShelfDialog((prev) => ({ ...prev, open: false }));
  };

  // Handlers for Bulk Shelf Generation
  const openBulkAddShelf = (locIndex: number) => {
    setBulkShelfDialog({
      open: true,
      locIndex,
      prefix: "SH-",
      startNum: 1,
      endNum: 10,
      padding: 2,
      namePattern: language === "th" ? "ชั้นวาง {number}" : "Shelf {number}",
      max_weight: "",
      width: "",
      length: "",
      height: "",
      suitable_product_types: "",
    });
  };

  const saveBulkShelves = () => {
    if (!setForm || bulkShelfDialog.locIndex === undefined) return;
    const prefix = bulkShelfDialog.prefix.trim();
    const startNum = parseInt(String(bulkShelfDialog.startNum));
    const endNum = parseInt(String(bulkShelfDialog.endNum));
    const padding = parseInt(String(bulkShelfDialog.padding)) || 0;
    const namePattern = bulkShelfDialog.namePattern.trim();

    if (isNaN(startNum) || isNaN(endNum) || startNum < 0 || endNum < 0) {
      setBulkShelfDialog((prev) => ({ ...prev, error: language === "th" ? "กรุณากรอกช่วงตัวเลขที่ถูกต้อง" : "Please fill in a valid number range" }));
      return;
    }
    if (startNum > endNum) {
      setBulkShelfDialog((prev) => ({ ...prev, error: language === "th" ? "ตัวเลขเริ่มต้นต้องไม่มากกว่าตัวเลขสิ้นสุด" : "Start number cannot be greater than end number" }));
      return;
    }
    if (endNum - startNum > 100) {
      setBulkShelfDialog((prev) => ({ ...prev, error: language === "th" ? "สร้างได้สูงสุดครั้งละ 100 ชั้นวาง" : "You can create up to 100 shelves at a time" }));
      return;
    }

    const loc = locationsList[bulkShelfDialog.locIndex];
    const shelves = loc.shelf || [];
    const updatedShelves = [...shelves];

    const maxWeight = parseFloat(String(bulkShelfDialog.max_weight)) || 0;
    const w = parseFloat(String(bulkShelfDialog.width)) || 0;
    const l = parseFloat(String(bulkShelfDialog.length)) || 0;
    const h = parseFloat(String(bulkShelfDialog.height)) || 0;
    const suitable = bulkShelfDialog.suitable_product_types?.trim() || "";

    const duplicates: string[] = [];

    for (let i = startNum; i <= endNum; i++) {
      let numStr = String(i);
      if (padding > 0) {
        numStr = numStr.padStart(padding, "0");
      }
      const code = `${prefix}${numStr}`;
      const name = namePattern.replace("{number}", numStr);

      // Check if code is duplicate
      const isDuplicate = updatedShelves.some((shelf: any) => shelf.code?.toLowerCase() === code.toLowerCase());
      if (isDuplicate) {
        duplicates.push(code);
        continue;
      }

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
    }

    if (duplicates.length > 0 && duplicates.length === (endNum - startNum + 1)) {
      setBulkShelfDialog((prev) => ({ ...prev, error: language === "th" ? "รหัสชั้นวางที่สร้างมีอยู่แล้วทั้งหมด" : "All generated shelf codes already exist" }));
      return;
    }

    const updatedLocs = [...locationsList];
    updatedLocs[bulkShelfDialog.locIndex] = {
      ...loc,
      shelf: updatedShelves,
    };

    setForm({ ...form, location: updatedLocs });
    setBulkShelfDialog((prev) => ({ ...prev, open: false }));

    if (duplicates.length > 0) {
      alert(
        language === "th"
          ? `ข้ามรหัสที่ซ้ำกัน: ${duplicates.join(", ")}`
          : `Skipped duplicate codes: ${duplicates.join(", ")}`
      );
    }
  };

  // Reorder Locations
  const moveLocation = (fromIdx: number, toIdx: number) => {
    if (!setForm || toIdx < 0 || toIdx >= locationsList.length) return;
    const updated = [...locationsList];
    const [removed] = updated.splice(fromIdx, 1);
    updated.splice(toIdx, 0, removed);
    setForm({ ...form, location: updated });
  };

  // Reorder Shelves
  const moveShelf = (locIdx: number, fromIdx: number, toIdx: number) => {
    if (!setForm) return;
    const loc = locationsList[locIdx];
    const shelves = loc.shelf || [];
    if (toIdx < 0 || toIdx >= shelves.length) return;
    const updatedShelves = [...shelves];
    const [removed] = updatedShelves.splice(fromIdx, 1);
    updatedShelves.splice(toIdx, 0, removed);

    const updatedLocs = [...locationsList];
    updatedLocs[locIdx] = { ...loc, shelf: updatedShelves };
    setForm({ ...form, location: updatedLocs });
  };

  return (
    <div className="flex flex-col gap-4 border border-border bg-card rounded-2xl p-5 shadow-[0_4px_12px_rgba(160,64,53,0.08)] mt-3 border-l-4 border-l-primary/90 transition-all duration-200">
      {/* Editor Header */}
      <div className="flex items-center justify-between">
        <h3 className="text-sm font-bold text-foreground flex items-center gap-2">
          <MapPin className="size-4 text-primary shrink-0 animate-pulse" />
          {language === "th" ? "คลังสินค้า → โซนเก็บสินค้า → ชั้นวาง" : "Warehouse → Storage Zone → Shelf"}
        </h3>
        {isEdit && (
          <Button
            type="button"
            variant="outline"
            size="sm"
            onClick={openAddLocation}
            className="h-8 gap-1 border-primary/30 text-primary hover:bg-primary/5 hover:border-primary/50 transition-all text-xs font-semibold"
          >
            <Plus className="size-3.5" />
            {language === "th" ? "เพิ่มโซนเก็บสินค้า" : "Add Storage Zone"}
          </Button>
        )}
      </div>

      {locationsList.length === 0 ? (
        <div className="flex flex-col items-center justify-center p-8 text-center rounded-xl bg-secondary/5 border border-dashed border-border/70 hover:border-border transition-all">
          <MapPin className="size-10 text-muted-foreground/30 mb-3" />
          <p className="text-xs text-muted-foreground font-semibold">
            {language === "th"
              ? "ยังไม่มีโซนเก็บสินค้าในคลังนี้"
              : "No storage zones found in this warehouse."}
          </p>
          {isEdit && (
            <p className="text-[10px] text-muted-foreground/80 mt-1">
              {language === "th"
                ? "คลิกปุ่ม 'เพิ่มโซนเก็บสินค้า' ด้านบนเพื่อเริ่มจัดตั้งโครงสร้างโซนเก็บสินค้า"
                : "Click 'Add Storage Zone' above to set up your storage structure."}
            </p>
          )}
        </div>
      ) : (
        /* Tree Layout Container (matches Category List style) */
        <div className="flex flex-col border border-border/80 bg-card rounded-xl overflow-hidden shadow-sm">
          {locationsList.map((loc, locIdx) => {
            const isCollapsed = collapsedLocs[loc.code] ?? false;
            const shelves = loc.shelf || [];
            const hasAttributes = loc.max_weight > 0 || loc.width > 0 || loc.length > 0 || loc.height > 0 || loc.suitable_product_types;

            return (
              <div key={loc.code || locIdx} className="flex flex-col">
                {/* Level 1: Location Row (corresponds to Level 0 of Category Tree) */}
                <div
                  className="group/row relative flex items-center justify-between border-b border-border/40 py-2 px-3 hover:bg-muted/40 transition-[background-color,border-color,box-shadow,opacity,transform] duration-200 ease-out flex-wrap sm:flex-nowrap gap-2"
                  style={{ paddingLeft: "12px" }}
                >
                  {/* Left Controls & Code/Names */}
                  <div className="flex items-center gap-2 min-w-0 flex-1 sm:flex-none">
                    {/* Drag Handle representation & manual arrows */}
                    <div className="flex items-center gap-0.5 shrink-0">
                      <GripVertical className="size-4 text-primary/45 group-hover/row:text-primary transition-colors cursor-grab" />
                      {isEdit && (
                        <div className="flex flex-col shrink-0 ml-0.5 opacity-0 group-hover/row:opacity-100 transition-opacity duration-200">
                          <button
                            type="button"
                            onClick={() => moveLocation(locIdx, locIdx - 1)}
                            disabled={locIdx === 0}
                            className="text-muted-foreground/40 hover:text-primary disabled:opacity-20 transition-colors"
                            title={language === "th" ? "ย้ายขึ้น" : "Move Up"}
                          >
                            <ArrowUp className="size-2.5" />
                          </button>
                          <button
                            type="button"
                            onClick={() => moveLocation(locIdx, locIdx + 1)}
                            disabled={locIdx === locationsList.length - 1}
                            className="text-muted-foreground/40 hover:text-primary disabled:opacity-20 transition-colors"
                            title={language === "th" ? "ย้ายลง" : "Move Down"}
                          >
                            <ArrowDown className="size-2.5" />
                          </button>
                        </div>
                      )}
                    </div>

                    {/* Collapse Button */}
                    <button
                      type="button"
                      onClick={() => toggleLocation(loc.code)}
                      className={cn(
                        "size-6 flex items-center justify-center rounded hover:bg-muted text-muted-foreground transition-all shrink-0",
                        shelves.length === 0 && "invisible"
                      )}
                    >
                      {isCollapsed ? (
                        <ChevronRight className="size-4 text-primary" />
                      ) : (
                        <ChevronDown className="size-4 text-primary" />
                      )}
                    </button>

                    {/* Index Number */}
                    <span className="font-extrabold text-sm text-primary shrink-0 min-w-[20px] text-center">
                      {locIdx + 1}
                    </span>

                    {/* Code & Localized Name */}
                    <span className="font-bold text-[15px] text-foreground truncate max-w-[200px] sm:max-w-xs">
                      {loc.code} - {displayLocationName(loc.names)}
                    </span>

                    {/* Category style "ลูก" badge (Shelves count) */}
                    {shelves.length > 0 && (
                      <span className="shrink-0 text-[11px] bg-primary/10 text-primary px-2.5 py-0.5 rounded-full font-bold ml-1.5 border border-primary/20">
                        {language === "th" ? `ลูก ${shelves.length}` : `${shelves.length} child`}
                      </span>
                    )}
                  </div>

                  {/* Middle Physical Attributes Badges */}
                  <div className="flex flex-wrap items-center gap-1.5 flex-1 justify-start sm:justify-center px-2">
                    {loc.max_weight > 0 && (
                      <span className="inline-flex items-center gap-1 bg-primary/5 text-primary border border-primary/10 px-2 py-0.5 rounded-md font-semibold text-[9px] shrink-0" title={language === "th" ? "น้ำหนักบรรทุกสูงสุด" : "Max Capacity"}>
                        <Scale className="size-2.5 shrink-0" />
                        {loc.max_weight.toLocaleString()} {language === "th" ? "กก." : "kg"}
                      </span>
                    )}
                    {(loc.width > 0 || loc.length > 0 || loc.height > 0) && (
                      <span className="inline-flex items-center gap-1 bg-amber-500/10 text-amber-600 dark:text-amber-500 border border-amber-500/15 px-2 py-0.5 rounded-md font-semibold text-[9px] shrink-0" title={language === "th" ? "กว้าง x ยาว x สูง" : "Dimensions"}>
                        <Ruler className="size-2.5 shrink-0" />
                        {loc.width || "-"}x{loc.length || "-"}x{loc.height || "-"}
                      </span>
                    )}
                    {loc.suitable_product_types && (
                      <span
                        className="inline-flex items-center gap-1 bg-teal-500/10 text-teal-600 dark:text-teal-500 border border-teal-500/15 px-2 py-0.5 rounded-md font-semibold text-[9px] truncate max-w-[150px] shrink-0"
                        title={loc.suitable_product_types}
                      >
                        <Package className="size-2.5 shrink-0" />
                        {loc.suitable_product_types}
                      </span>
                    )}
                  </div>

                  {/* Right Actions */}
                  {isEdit && (
                    <div className="flex items-center gap-1 shrink-0 ml-auto sm:ml-0 opacity-60 transition-opacity duration-200 group-hover/row:opacity-100">
                      {/* Plus icon like in screenshot for adding child shelf */}
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        onClick={() => openAddShelf(locIdx)}
                        className="size-7 rounded-full text-emerald-600 hover:text-emerald-700 hover:bg-emerald-50 dark:hover:bg-emerald-950/40 shrink-0"
                        title={language === "th" ? "เพิ่มชั้นวาง" : "Add Shelf"}
                      >
                        <FolderPlus className="size-3.5" />
                      </Button>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        onClick={() => openBulkAddShelf(locIdx)}
                        className="size-7 rounded-full text-orange-600 hover:text-orange-700 hover:bg-orange-50 dark:hover:bg-orange-950/40 shrink-0"
                        title={language === "th" ? "เพิ่มกลุ่มชั้นวาง (Bulk)" : "Bulk Add Shelves"}
                      >
                        <Layers className="size-3.5" />
                      </Button>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        onClick={() => openEditLocation(locIdx, loc)}
                        className="size-7 rounded-full text-blue-600 hover:text-blue-700 hover:bg-blue-50 dark:hover:bg-blue-950/40 shrink-0"
                        title={language === "th" ? "แก้ไข" : "Edit"}
                      >
                        <Edit3 className="size-3.5" />
                      </Button>
                      <Button
                        type="button"
                        variant="ghost"
                        size="icon"
                        onClick={() => deleteLocation(locIdx)}
                        className="size-7 rounded-full text-destructive hover:bg-destructive/10 shrink-0"
                        title={language === "th" ? "ลบ" : "Delete"}
                      >
                        <Trash2 className="size-3.5" />
                      </Button>
                    </div>
                  )}
                </div>

                {/* Level 2: Shelves Rows (Indented, matches Level 1 of Category Tree) */}
                {!isCollapsed && (
                  <div className="flex flex-col">
                    {shelves.length === 0 ? (
                      <div className="flex items-center py-2 px-4 pl-14 border-b border-border/40 text-[11px] text-muted-foreground/60 italic bg-secondary/5">
                        {language === "th" ? "ไม่มีชั้นวางในโซนเก็บสินค้านี้" : "No shelves in this storage zone."}
                      </div>
                    ) : (
                      shelves.map((shelf: any, shelfIdx: number) => {
                        const pCount = shelf.productitems?.length || 0;
                        const hasShelfAttributes = shelf.max_weight > 0 || shelf.width > 0 || shelf.length > 0 || shelf.height > 0 || shelf.suitable_product_types;

                        return (
                          <div
                            key={shelf.code || shelfIdx}
                            className="group/shelf relative flex items-center justify-between py-2 px-3 border-b border-border/40 bg-secondary/5 hover:bg-muted/40 transition-[background-color,border-color,box-shadow,opacity,transform] duration-200 ease-out flex-wrap sm:flex-nowrap gap-2"
                            style={{ paddingLeft: "36px" }}
                          >
                            {/* Left Controls & Code/Name */}
                            <div className="flex items-center gap-2 min-w-0 flex-1 sm:flex-none">
                              {/* Grip Handle and manual arrows for shelf sorting */}
                              <div className="flex items-center gap-0.5 shrink-0">
                                <GripVertical className="size-3.5 text-sky-500/55 group-hover/shelf:text-sky-600 dark:group-hover/shelf:text-sky-300 transition-colors cursor-grab" />
                                {isEdit && (
                                  <div className="flex flex-col shrink-0 ml-0.5 opacity-0 group-hover/shelf:opacity-100 transition-opacity duration-200">
                                    <button
                                      type="button"
                                      onClick={() => moveShelf(locIdx, shelfIdx, shelfIdx - 1)}
                                      disabled={shelfIdx === 0}
                                      className="text-muted-foreground/40 hover:text-sky-600 disabled:opacity-20 transition-colors"
                                      title={language === "th" ? "ย้ายขึ้น" : "Move Up"}
                                    >
                                      <ArrowUp className="size-2.5" />
                                    </button>
                                    <button
                                      type="button"
                                      onClick={() => moveShelf(locIdx, shelfIdx, shelfIdx + 1)}
                                      disabled={shelfIdx === shelves.length - 1}
                                      className="text-muted-foreground/40 hover:text-sky-600 disabled:opacity-20 transition-colors"
                                      title={language === "th" ? "ย้ายลง" : "Move Down"}
                                    >
                                      <ArrowDown className="size-2.5" />
                                    </button>
                                  </div>
                                )}
                              </div>

                              {/* Index Number (Secondary blue-ish style) */}
                              <span className="font-bold text-xs text-sky-600 dark:text-sky-300 shrink-0 min-w-[16px] text-center">
                                {shelfIdx + 1}
                              </span>

                              {/* Code & Shelf Name */}
                              <span className="font-semibold text-sm text-sky-900 dark:text-sky-100 truncate max-w-[180px] sm:max-w-xs">
                                {shelf.code} - {shelf.name}
                              </span>

                              {/* Product count badge */}
                              {pCount > 0 && (
                                <span className="shrink-0 text-[10px] bg-emerald-500/10 text-emerald-700 dark:text-emerald-300 border border-emerald-500/20 px-2 py-0.5 rounded-full font-bold ml-1.5">
                                  {language === "th" ? `สินค้า ${pCount}` : `${pCount} items`}
                                </span>
                              )}
                            </div>

                            {/* Middle Shelf Attributes */}
                            <div className="flex flex-wrap items-center gap-1.5 flex-1 justify-start sm:justify-center px-2">
                              {shelf.max_weight > 0 && (
                                <span className="inline-flex items-center gap-0.5 bg-primary/5 text-primary/80 border border-primary/5 px-1.5 py-0.2 rounded text-[9px] font-medium shrink-0">
                                  <Scale className="size-2.5 shrink-0" />
                                  {shelf.max_weight.toLocaleString()} {language === "th" ? "กก." : "kg"}
                                </span>
                              )}
                              {(shelf.width > 0 || shelf.length > 0 || shelf.height > 0) && (
                                <span className="inline-flex items-center gap-0.5 bg-amber-500/10 text-amber-600 border border-amber-500/10 px-1.5 py-0.2 rounded text-[9px] font-medium shrink-0">
                                  <Ruler className="size-2.5 shrink-0" />
                                  {shelf.width || "-"}x{shelf.length || "-"}x{shelf.height || "-"}
                                </span>
                              )}
                              {shelf.suitable_product_types && (
                                <span
                                  className="inline-flex items-center gap-0.5 bg-teal-500/10 text-teal-600 border border-teal-500/10 px-1.5 py-0.2 rounded text-[9px] font-medium truncate max-w-[120px] shrink-0"
                                  title={shelf.suitable_product_types}
                                >
                                  <Package className="size-2.5 shrink-0" />
                                  {shelf.suitable_product_types}
                                </span>
                              )}
                            </div>

                            {/* Right Actions */}
                            {isEdit && (
                              <div className="flex items-center gap-1 shrink-0 ml-auto sm:ml-0 opacity-60 transition-opacity duration-200 group-hover/shelf:opacity-100">
                                <Button
                                  type="button"
                                  variant="ghost"
                                  size="icon"
                                  onClick={() => openEditShelf(locIdx, shelfIdx, shelf)}
                                  className="size-7 rounded-full text-blue-600 hover:text-blue-700 hover:bg-blue-50 dark:hover:bg-blue-950/40 shrink-0"
                                  title={language === "th" ? "แก้ไข" : "Edit"}
                                >
                                  <Edit3 className="size-3.5" />
                                </Button>
                                <Button
                                  type="button"
                                  variant="ghost"
                                  size="icon"
                                  onClick={() => deleteShelf(locIdx, shelfIdx)}
                                  className="size-7 rounded-full text-destructive hover:bg-destructive/10 shrink-0"
                                  title={language === "th" ? "ลบ" : "Delete"}
                                >
                                  <Trash2 className="size-3.5" />
                                </Button>
                              </div>
                            )}
                          </div>
                        );
                      })
                    )}
                  </div>
                )}
              </div>
            );
          })}
        </div>
      )}

      {/* Location Dialog */}
      {locDialog.open && (
        <div className="backdrop-blur-sm bg-black/40 z-50 fixed inset-0 flex items-center justify-center p-4" onClick={() => setLocDialog((prev) => ({ ...prev, open: false }))} role="presentation">
          <section
            className="flex flex-col gap-4 max-h-[90vh] w-[min(520px,calc(100vw-24px))] overflow-hidden rounded-2xl border border-border bg-card/95 p-5 text-foreground shadow-2xl z-50 relative"
            role="dialog"
            aria-modal="true"
            onClick={(e) => e.stopPropagation()}
          >
            <header className="flex items-center justify-between border-b border-border/50 pb-2">
              <h2 className="text-sm font-bold text-foreground">
                {locDialog.mode === "create"
                  ? language === "th"
                    ? "เพิ่มโซนเก็บสินค้า"
                    : "Add Storage Zone"
                  : language === "th"
                    ? "แก้ไขโซนเก็บสินค้า"
                    : "Edit Storage Zone"}
              </h2>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="size-7 text-muted-foreground hover:bg-secondary/10"
                onClick={() => setLocDialog((prev) => ({ ...prev, open: false }))}
              >
                <X className="size-4" />
              </Button>
            </header>

            <div className="flex flex-col gap-3.5 my-1 overflow-y-auto pr-1">
              <div className="grid grid-cols-2 gap-3">
                <div className="flex flex-col gap-1">
                  <label className="text-[11px] font-bold text-muted-foreground uppercase">
                    {language === "th" ? "รหัสโซนเก็บสินค้า" : "Storage Zone Code"} <span className="text-destructive">*</span>
                  </label>
                  <Input
                    value={locDialog.code}
                    onChange={(e) => setLocDialog((prev) => ({ ...prev, code: e.target.value }))}
                    placeholder={language === "th" ? "เช่น ZONE-A" : "e.g. ZONE-A"}
                    className="h-9 text-xs"
                  />
                </div>

                <div className="flex flex-col gap-1">
                  <label className="text-[11px] font-bold text-muted-foreground uppercase">
                    {language === "th" ? "ประเภทสินค้าที่เหมาะสม" : "Suitable Product Types"}
                  </label>
                  <Input
                    value={locDialog.suitable_product_types}
                    onChange={(e) => setLocDialog((prev) => ({ ...prev, suitable_product_types: e.target.value }))}
                    placeholder={language === "th" ? "เช่น สินค้าแช่แข็ง, ของเปราะบาง" : "e.g. Frozen, Fragile"}
                    className="h-9 text-xs"
                  />
                </div>
              </div>

              <div className="border-t border-border/30 my-1 pt-3">
                <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider block mb-2">
                  {language === "th" ? "ชื่อโซนเก็บสินค้าหลายภาษา" : "Multilingual Storage Zone Names"}
                </span>
                <div className="grid grid-cols-2 gap-3">
                  {editorLanguages.map((lang) => (
                    <div key={lang} className="flex flex-col gap-1">
                      <label className="text-[10px] font-bold text-muted-foreground uppercase flex items-center gap-1">
                        {getLanguageName(lang)} {lang === "th" && <span className="text-destructive">*</span>}
                      </label>
                      <Input
                        value={locDialog.names[lang] || ""}
                        onChange={(e) =>
                          setLocDialog((prev) => ({
                            ...prev,
                            names: { ...prev.names, [lang]: e.target.value },
                          }))
                        }
                        placeholder={
                          lang === "th"
                            ? "เช่น โซนเอ"
                            : lang === "en"
                              ? "e.g. Zone A"
                              : ""
                        }
                        className="h-9 text-xs"
                      />
                    </div>
                  ))}
                </div>
              </div>

              {locDialog.error && (
                <span className="text-[11px] font-semibold text-destructive mt-1">
                  {locDialog.error}
                </span>
              )}
            </div>

            <footer className="flex justify-end gap-2 border-t border-border/50 pt-3 mt-auto">
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="h-8 text-xs font-semibold"
                onClick={() => setLocDialog((prev) => ({ ...prev, open: false }))}
              >
                {language === "th" ? "ยกเลิก" : "Cancel"}
              </Button>
              <Button
                type="button"
                size="sm"
                className="h-8 text-xs font-semibold"
                onClick={saveLocation}
              >
                {language === "th" ? "ตกลง" : "OK"}
              </Button>
            </footer>
          </section>
        </div>
      )}

      {/* Shelf Dialog */}
      {shelfDialog.open && (
        <div className="backdrop-blur-sm bg-black/40 z-50 fixed inset-0 flex items-center justify-center p-4" onClick={() => setShelfDialog((prev) => ({ ...prev, open: false }))} role="presentation">
          <section
            className="flex flex-col gap-4 max-h-[90vh] w-[min(520px,calc(100vw-24px))] overflow-hidden rounded-2xl border border-border bg-card/95 p-5 text-foreground shadow-2xl z-50 relative"
            role="dialog"
            aria-modal="true"
            onClick={(e) => e.stopPropagation()}
          >
            <header className="flex items-center justify-between border-b border-border/50 pb-2">
              <h2 className="text-sm font-bold text-foreground">
                {shelfDialog.mode === "create"
                  ? language === "th"
                    ? "เพิ่มชั้นวางสินค้า"
                    : "Add Shelf"
                  : language === "th"
                    ? "แก้ไขชั้นวางสินค้า"
                    : "Edit Shelf"}
              </h2>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="size-7 text-muted-foreground hover:bg-secondary/10"
                onClick={() => setShelfDialog((prev) => ({ ...prev, open: false }))}
              >
                <X className="size-4" />
              </Button>
            </header>

            <div className="flex flex-col gap-3.5 my-1 overflow-y-auto pr-1">
              <div className="grid grid-cols-2 gap-3">
                <div className="flex flex-col gap-1">
                  <label className="text-[11px] font-bold text-muted-foreground uppercase">
                    {language === "th" ? "รหัสชั้นวาง" : "Shelf Code"} <span className="text-destructive">*</span>
                  </label>
                  <Input
                    value={shelfDialog.code}
                    onChange={(e) => setShelfDialog((prev) => ({ ...prev, code: e.target.value }))}
                    placeholder={language === "th" ? "เช่น SH-01" : "e.g. SH-01"}
                    className="h-9 text-xs"
                  />
                </div>

                <div className="flex flex-col gap-1">
                  <label className="text-[11px] font-bold text-muted-foreground uppercase">
                    {language === "th" ? "ชื่อชั้นวาง" : "Shelf Name"} <span className="text-destructive">*</span>
                  </label>
                  <Input
                    value={shelfDialog.name}
                    onChange={(e) => setShelfDialog((prev) => ({ ...prev, name: e.target.value }))}
                    placeholder={language === "th" ? "เช่น แถว A ชั้น 1" : "e.g. Row A Tier 1"}
                    className="h-9 text-xs"
                  />
                </div>
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div className="flex flex-col gap-1">
                  <label className="text-[11px] font-bold text-muted-foreground uppercase">
                    {language === "th" ? "รับน้ำหนักสูงสุด (กก.)" : "Max Weight (kg)"}
                  </label>
                  <Input
                    type="number"
                    value={shelfDialog.max_weight}
                    onChange={(e) => setShelfDialog((prev) => ({ ...prev, max_weight: e.target.value }))}
                    placeholder="e.g. 200"
                    className="h-9 text-xs"
                  />
                </div>

                <div className="flex flex-col gap-1">
                  <label className="text-[11px] font-bold text-muted-foreground uppercase">
                    {language === "th" ? "ประเภทสินค้าที่เหมาะสม" : "Suitable Product Types"}
                  </label>
                  <Input
                    value={shelfDialog.suitable_product_types}
                    onChange={(e) => setShelfDialog((prev) => ({ ...prev, suitable_product_types: e.target.value }))}
                    placeholder={language === "th" ? "เช่น ของเหลว, ยา" : "e.g. Liquids, Medicine"}
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
                    value={shelfDialog.width}
                    onChange={(e) => setShelfDialog((prev) => ({ ...prev, width: e.target.value }))}
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
                    value={shelfDialog.length}
                    onChange={(e) => setShelfDialog((prev) => ({ ...prev, length: e.target.value }))}
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
                    value={shelfDialog.height}
                    onChange={(e) => setShelfDialog((prev) => ({ ...prev, height: e.target.value }))}
                    placeholder="H"
                    className="h-9 text-xs"
                  />
                </div>
              </div>

              {shelfDialog.error && (
                <span className="text-[11px] font-semibold text-destructive mt-1">
                  {shelfDialog.error}
                </span>
              )}
            </div>

            <footer className="flex justify-end gap-2 border-t border-border/50 pt-3 mt-auto">
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="h-8 text-xs font-semibold"
                onClick={() => setShelfDialog((prev) => ({ ...prev, open: false }))}
              >
                {language === "th" ? "ยกเลิก" : "Cancel"}
              </Button>
              <Button
                type="button"
                size="sm"
                className="h-8 text-xs font-semibold"
                onClick={saveShelf}
              >
                {language === "th" ? "ตกลง" : "OK"}
              </Button>
            </footer>
          </section>
        </div>
      )}

      {/* Bulk Shelf Dialog */}
      {bulkShelfDialog.open && (
        <div className="backdrop-blur-sm bg-black/40 z-50 fixed inset-0 flex items-center justify-center p-4" onClick={() => setBulkShelfDialog((prev) => ({ ...prev, open: false }))} role="presentation">
          <section
            className="flex flex-col gap-4 max-h-[90vh] w-[min(520px,calc(100vw-24px))] overflow-hidden rounded-2xl border border-border bg-card/95 p-5 text-foreground shadow-2xl z-50 relative"
            role="dialog"
            aria-modal="true"
            onClick={(e) => e.stopPropagation()}
          >
            <header className="flex items-center justify-between border-b border-border/50 pb-2">
              <h2 className="text-sm font-bold text-foreground flex items-center gap-1.5">
                <Layers className="size-4 text-primary shrink-0" />
                {language === "th" ? "เพิ่มชั้นวางแบบกลุ่ม" : "Bulk Add Shelves"}
              </h2>
              <Button
                type="button"
                variant="ghost"
                size="icon"
                className="size-7 text-muted-foreground hover:bg-secondary/10"
                onClick={() => setBulkShelfDialog((prev) => ({ ...prev, open: false }))}
              >
                <X className="size-4" />
              </Button>
            </header>

            <div className="flex flex-col gap-3.5 my-1 overflow-y-auto pr-1">
              <div className="bg-primary/5 border border-primary/10 rounded-xl p-3 text-[11px] text-foreground/80 leading-relaxed">
                {language === "th"
                  ? "ระบบจะสร้างรหัสชั้นวางแบบเรียงลำดับให้อัตโนมัติ เช่น หากป้อนคำนำหน้า SH- ลำดับเริ่ม 1 ถึง 10 จะได้รหัส SH-01 ถึง SH-10"
                  : "The system will generate sequential shelf codes automatically. E.g. Prefix 'SH-', Start 1, End 10 will produce SH-01 to SH-10."}
              </div>

              <div className="grid grid-cols-2 gap-3">
                <div className="flex flex-col gap-1">
                  <label className="text-[11px] font-bold text-muted-foreground uppercase">
                    {language === "th" ? "คำนำหน้ารหัส (Prefix)" : "Code Prefix"} <span className="text-destructive">*</span>
                  </label>
                  <Input
                    value={bulkShelfDialog.prefix}
                    onChange={(e) => setBulkShelfDialog((prev) => ({ ...prev, prefix: e.target.value }))}
                    placeholder="e.g. SH-"
                    className="h-9 text-xs"
                  />
                </div>

                <div className="flex flex-col gap-1">
                  <label className="text-[11px] font-bold text-muted-foreground uppercase">
                    {language === "th" ? "รูปแบบชื่อ (Pattern)" : "Name Pattern"} <span className="text-destructive">*</span>
                  </label>
                  <Input
                    value={bulkShelfDialog.namePattern}
                    onChange={(e) => setBulkShelfDialog((prev) => ({ ...prev, namePattern: e.target.value }))}
                    placeholder="e.g. ชั้นวาง {number}"
                    className="h-9 text-xs"
                  />
                </div>
              </div>

              <div className="grid grid-cols-3 gap-2">
                <div className="flex flex-col gap-1">
                  <label className="text-[11px] font-bold text-muted-foreground uppercase">
                    {language === "th" ? "เลขเริ่มต้น" : "Start Number"} <span className="text-destructive">*</span>
                  </label>
                  <Input
                    type="number"
                    value={bulkShelfDialog.startNum}
                    onChange={(e) => setBulkShelfDialog((prev) => ({ ...prev, startNum: e.target.value }))}
                    placeholder="1"
                    className="h-9 text-xs"
                  />
                </div>
                <div className="flex flex-col gap-1">
                  <label className="text-[11px] font-bold text-muted-foreground uppercase">
                    {language === "th" ? "เลขสิ้นสุด" : "End Number"} <span className="text-destructive">*</span>
                  </label>
                  <Input
                    type="number"
                    value={bulkShelfDialog.endNum}
                    onChange={(e) => setBulkShelfDialog((prev) => ({ ...prev, endNum: e.target.value }))}
                    placeholder="10"
                    className="h-9 text-xs"
                  />
                </div>
                <div className="flex flex-col gap-1">
                  <label className="text-[11px] font-bold text-muted-foreground uppercase">
                    {language === "th" ? "จำนวนหลัก (Padding)" : "Zero Padding"}
                  </label>
                  <Input
                    type="number"
                    value={bulkShelfDialog.padding}
                    onChange={(e) => setBulkShelfDialog((prev) => ({ ...prev, padding: e.target.value }))}
                    placeholder="2"
                    className="h-9 text-xs"
                  />
                </div>
              </div>

              <div className="border-t border-border/30 my-1 pt-3">
                <span className="text-[10px] font-bold text-muted-foreground uppercase tracking-wider block mb-2">
                  {language === "th" ? "คุณสมบัติเริ่มต้นของชั้นวางทั้งหมด" : "Default Shelf Attributes"}
                </span>

                <div className="grid grid-cols-2 gap-3">
                  <div className="flex flex-col gap-1">
                    <label className="text-[11px] font-bold text-muted-foreground uppercase">
                      {language === "th" ? "รับน้ำหนักสูงสุด (กก.)" : "Max Weight (kg)"}
                    </label>
                    <Input
                      type="number"
                      value={bulkShelfDialog.max_weight}
                      onChange={(e) => setBulkShelfDialog((prev) => ({ ...prev, max_weight: e.target.value }))}
                      placeholder="e.g. 200"
                      className="h-9 text-xs"
                    />
                  </div>

                  <div className="flex flex-col gap-1">
                    <label className="text-[11px] font-bold text-muted-foreground uppercase">
                      {language === "th" ? "ประเภทสินค้าที่เหมาะสม" : "Suitable Product Types"}
                    </label>
                    <Input
                      value={bulkShelfDialog.suitable_product_types}
                      onChange={(e) => setBulkShelfDialog((prev) => ({ ...prev, suitable_product_types: e.target.value }))}
                      placeholder={language === "th" ? "เช่น ยา, สินค้าทั่วไป" : "e.g. Medicine, General"}
                      className="h-9 text-xs"
                    />
                  </div>
                </div>

                <div className="grid grid-cols-3 gap-2 mt-3">
                  <div className="flex flex-col gap-1">
                    <label className="text-[11px] font-bold text-muted-foreground uppercase">
                      {language === "th" ? "กว้าง (ซม.)" : "Width (cm)"}
                    </label>
                    <Input
                      type="number"
                      value={bulkShelfDialog.width}
                      onChange={(e) => setBulkShelfDialog((prev) => ({ ...prev, width: e.target.value }))}
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
                      value={bulkShelfDialog.length}
                      onChange={(e) => setBulkShelfDialog((prev) => ({ ...prev, length: e.target.value }))}
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
                      value={bulkShelfDialog.height}
                      onChange={(e) => setBulkShelfDialog((prev) => ({ ...prev, height: e.target.value }))}
                      placeholder="H"
                      className="h-9 text-xs"
                    />
                  </div>
                </div>
              </div>

              {bulkShelfDialog.error && (
                <span className="text-[11px] font-semibold text-destructive mt-1">
                  {bulkShelfDialog.error}
                </span>
              )}
            </div>

            <footer className="flex justify-end gap-2 border-t border-border/50 pt-3 mt-auto">
              <Button
                type="button"
                variant="outline"
                size="sm"
                className="h-8 text-xs font-semibold"
                onClick={() => setBulkShelfDialog((prev) => ({ ...prev, open: false }))}
              >
                {language === "th" ? "ยกเลิก" : "Cancel"}
              </Button>
              <Button
                type="button"
                size="sm"
                className="h-8 text-xs font-semibold"
                onClick={saveBulkShelves}
              >
                {language === "th" ? "สร้างทั้งหมด" : "Generate"}
              </Button>
            </footer>
          </section>
        </div>
      )}
    </div>
  );
}
