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
} from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { LANGUAGES, type LanguageCode } from "@/lib/i18n";
import { cn } from "@/lib/utils";
import { normalizeLanguageConfigs } from "./system-settings-screen";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { deriveMainApiUrl } from "@/lib/backend-url";

interface CompanyBranchTreeViewProps {
  auth: { token: string; backendUrl: string } | null;
  workspace: CompanyWorkspace | null;
  language: LanguageCode;
  onRefresh?: () => void;
}

interface CompanyWorkspace {
  shop: { shopid: string };
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
  guid_fixed?: string;
  code?: string;
  names?: LocalizedNames;
  tax_id?: string;
  is_active?: boolean;
}

interface BranchRecord {
  guid_fixed?: string;
  company_guid?: string;
  code?: string;
  names?: LocalizedNames;
  is_active?: boolean;
}

type NodeType = "company" | "branch";

interface SelectedNode {
  type: NodeType;
  guid_fixed?: string;
  company_guid?: string;
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

export function CompanyBranchTreeView({
  auth,
  workspace,
  language,
  onRefresh,
}: CompanyBranchTreeViewProps) {
  const { confirm, confirmationDialog } = useConfirmDialog();
  const [companies, setCompanies] = useState<CompanyRecord[]>([]);
  const [branches, setBranches] = useState<BranchRecord[]>([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [saveSuccess, setSaveSuccess] = useState(false);

  const mainApiUrl = useMemo(() => {
    if (!auth?.backendUrl) return "";
    try {
      return deriveMainApiUrl(auth.backendUrl);
    } catch {
      return auth.backendUrl;
    }
  }, [auth]);

  // Active languages for multilingual names
  const editorLanguages = useMemo(() => {
    if (!workspace) return ["th"];
    const configs = workspace.shopInfo?.settings?.languageconfigs || [];
    const defaultCode = workspace.shopInfo?.settings?.language || "th";
    return normalizeLanguageConfigs(configs, defaultCode).map((row) => row.code);
  }, [workspace]);

  // Collapsed states
  const [collapsedCompanies, setCollapsedCompanies] = useState<Record<string, boolean>>({});

  // Selected Node for Right Form Editing
  const [selectedNode, setSelectedNode] = useState<SelectedNode | null>(null);
  const [formType, setFormType] = useState<"edit_company" | "edit_branch" | "create_company" | "create_branch" | null>(null);

  // Fetch Companies & Branches
  const loadData = useCallback(async () => {
    if (!auth || !mainApiUrl) return;
    setLoading(true);
    try {
      // Load Companies
      const resComp = await fetch(`${mainApiUrl}/organization/company`, {
        headers: { Authorization: `Bearer ${auth.token}` },
      });
      const jsonComp = await resComp.json();
      if (jsonComp.success && Array.isArray(jsonComp.data)) {
        setCompanies(jsonComp.data);
      }

      // Load Branches
      const resBranch = await fetch(`${mainApiUrl}/organization/branch`, {
        headers: { Authorization: `Bearer ${auth.token}` },
      });
      const jsonBranch = await resBranch.json();
      if (jsonBranch.success && Array.isArray(jsonBranch.data)) {
        setBranches(jsonBranch.data);
      }
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  }, [auth, mainApiUrl]);

  useEffect(() => {
    void loadData();
  }, [loadData]);

  // Initialize selected node
  useEffect(() => {
    if (companies.length > 0 && !selectedNode) {
      const first = companies[0];
      setSelectedNode({
        type: "company",
        guid_fixed: first.guid_fixed,
        data: first,
      });
      setFormType("edit_company");
    }
  }, [companies, selectedNode]);

  // Right Form Values
  const [formCode, setFormCode] = useState("");
  const [formTaxId, setFormTaxId] = useState("");
  const [formIsActive, setFormIsActive] = useState(true);
  const [formNames, setFormNames] = useState<Record<string, string>>({});

  useEffect(() => {
    if (!selectedNode) return;
    const namesObj: Record<string, string> = {};
    const rawNames = selectedNode.data.names;
    editorLanguages.forEach((lang) => {
      namesObj[lang] = getNameFromObject(rawNames, lang);
    });

    setFormCode(selectedNode.data.code || "");
    setFormIsActive(selectedNode.data.is_active !== false);
    setFormNames(namesObj);

    if (selectedNode.type === "company") {
      setFormTaxId((selectedNode.data as CompanyRecord).tax_id || "");
    }
  }, [selectedNode, editorLanguages]);

  // Handle Save
  const handleSave = async () => {
    if (!auth || !selectedNode || !formType) return;
    setSaving(true);

    try {
      const namesList = Object.entries(formNames).map(([code, name]) => ({
        code,
        name,
        isauto: false,
        isdelete: false,
      }));

      let url = "";
      let method = "POST";
      let body: Record<string, unknown> = {};

      if (formType === "create_company") {
        url = `${mainApiUrl}/organization/company`;
        method = "POST";
        body = {
          code: formCode,
          names: namesList,
          tax_id: formTaxId,
          is_active: formIsActive,
        };
      } else if (formType === "edit_company") {
        url = `${mainApiUrl}/organization/company/${selectedNode.guid_fixed}`;
        method = "PUT";
        body = {
          code: formCode,
          names: namesList,
          tax_id: formTaxId,
          is_active: formIsActive,
        };
      } else if (formType === "create_branch") {
        url = `${mainApiUrl}/organization/branch`;
        method = "POST";
        body = {
          company_guid: selectedNode.company_guid,
          code: formCode,
          names: namesList,
          is_active: formIsActive,
        };
      } else if (formType === "edit_branch") {
        url = `${mainApiUrl}/organization/branch/${selectedNode.guid_fixed}`;
        method = "PUT";
        body = {
          company_guid: selectedNode.company_guid,
          code: formCode,
          names: namesList,
          is_active: formIsActive,
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

      const json = await res.json();
      if (json.success) {
        setSaveSuccess(true);
        setTimeout(() => setSaveSuccess(false), 2000);
        await loadData();
        if (formType.startsWith("create")) {
          // Select newly created node
          setSelectedNode({
            type: formType === "create_company" ? "company" : "branch",
            guid_fixed: json.id,
            company_guid: selectedNode.company_guid,
            data: {
              guid_fixed: json.id,
              code: formCode,
              names: namesList,
              is_active: formIsActive,
              ...(formType === "create_company" ? { tax_id: formTaxId } : {}),
            },
          });
          setFormType(formType === "create_company" ? "edit_company" : "edit_branch");
        }
      }
    } catch (e) {
      console.error(e);
    } finally {
      setSaving(false);
    }
  };

  // Handle Delete
  const handleDelete = async (node: SelectedNode) => {
    if (!auth || !node.guid_fixed) return;

    const confirmed = await confirm({
      title: node.type === "company" ? "ยืนยันการลบบริษัท" : "ยืนยันการลบสาขา",
      description: "ข้อมูลที่ถูกลบจะไม่แสดงในระบบการทำงาน แต่ประวัติข้อมูลเดิมจะยังถูกรักษาไว้ในฐานข้อมูล",
      confirmLabel: "ลบข้อมูล",
      cancelLabel: "ยกเลิก",
      tone: "danger",
    });

    if (!confirmed) return;

    setLoading(true);
    try {
      const url = `${mainApiUrl}/organization/${node.type}/${node.guid_fixed}`;
      const res = await fetch(url, {
        method: "DELETE",
        headers: { Authorization: `Bearer ${auth.token}` },
      });
      const json = await res.json();
      if (json.success) {
        await loadData();
        setSelectedNode(null);
        setFormType(null);
      }
    } catch (e) {
      console.error(e);
    } finally {
      setLoading(false);
    }
  };

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
                setFormType("create_company");
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

          {loading ? (
            <div className="flex justify-center py-12">
              <Loader2 className="w-8 h-8 animate-spin text-primary" />
            </div>
          ) : (
            <div className="space-y-2">
              {companies.map((comp) => {
                const compGuid = comp.guid_fixed || "";
                const isCollapsed = collapsedCompanies[compGuid];
                const isSelected = selectedNode?.type === "company" && selectedNode.guid_fixed === compGuid;
                const compBranches = branches.filter((b) => b.company_guid === compGuid);

                return (
                  <div key={compGuid} className="space-y-1">
                    <div
                      className={cn(
                        "group flex items-center justify-between p-2 rounded-lg cursor-pointer transition-all",
                        isSelected ? "bg-primary/10 border-l-4 border-primary pl-1 font-bold" : "hover:bg-accent pl-2"
                      )}
                      onClick={() => {
                        setSelectedNode({
                          type: "company",
                          guid_fixed: compGuid,
                          data: comp,
                        });
                        setFormType("edit_company");
                      }}
                    >
                      <div className="flex items-center gap-2 min-w-0">
                        <button
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
                        <Building2 className="w-4 h-4 text-primary shrink-0" />
                        <span className="truncate">
                          [{comp.code}] {getNameFromObject(comp.names, language) || comp.code}
                        </span>
                      </div>
                      <div className="opacity-0 group-hover:opacity-100 flex items-center gap-1">
                        <Button
                          size="icon"
                          variant="ghost"
                          className="w-7 h-7 hover:text-primary"
                          title="เพิ่มสาขา"
                          onClick={(e) => {
                            e.stopPropagation();
                            setSelectedNode({
                              type: "branch",
                              company_guid: compGuid,
                              data: {},
                            });
                            setFormType("create_branch");
                          }}
                        >
                          <Plus className="w-3.5 h-3.5" />
                        </Button>
                        <Button
                          size="icon"
                          variant="ghost"
                          className="w-7 h-7 hover:text-destructive"
                          onClick={(e) => {
                            e.stopPropagation();
                            void handleDelete({
                              type: "company",
                              guid_fixed: compGuid,
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
                          const brGuid = br.guid_fixed || "";
                          const isBrSelected = selectedNode?.type === "branch" && selectedNode.guid_fixed === brGuid;

                          return (
                            <div
                              key={brGuid}
                              className={cn(
                                "group flex items-center justify-between p-2 rounded-lg cursor-pointer transition-all",
                                isBrSelected ? "bg-sky-500/10 border-l-4 border-sky-500 pl-1 font-semibold" : "hover:bg-accent pl-2"
                              )}
                              onClick={() => {
                                setSelectedNode({
                                  type: "branch",
                                  guid_fixed: brGuid,
                                  company_guid: compGuid,
                                  data: br,
                                });
                                setFormType("edit_branch");
                              }}
                            >
                              <div className="flex items-center gap-2 min-w-0">
                                <GitBranch className="w-3.5 h-3.5 text-sky-500 shrink-0" />
                                <span className="text-sm truncate">
                                  [{br.code}] {getNameFromObject(br.names, language) || br.code}
                                </span>
                              </div>
                              <div className="opacity-0 group-hover:opacity-100 flex items-center gap-1">
                                <Button
                                  size="icon"
                                  variant="ghost"
                                  className="w-7 h-7 hover:text-destructive"
                                  onClick={(e) => {
                                    e.stopPropagation();
                                    void handleDelete({
                                      type: "branch",
                                      guid_fixed: brGuid,
                                      company_guid: compGuid,
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
              <div className="flex items-center justify-between border-b pb-4">
                <div>
                  <h3 className="text-xl font-bold text-foreground">
                    {formType === "create_company" && "เพิ่มบริษัทใหม่"}
                    {formType === "edit_company" && "แก้ไขข้อมูลบริษัท"}
                    {formType === "create_branch" && "เพิ่มสาขาใหม่"}
                    {formType === "edit_branch" && "แก้ไขข้อมูลสาขา"}
                  </h3>
                  <p className="text-sm text-muted-foreground mt-1">
                    {formType.startsWith("create") ? "ระบุข้อมูลรายละเอียดหลักเพื่อเพิ่มข้อมูลเข้าระบบ" : "แก้ไขรายละเอียดข้อมูลและบันทึกประวัติ"}
                  </p>
                </div>
              </div>

              {/* Form fields */}
              <div className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <label className="text-sm font-semibold text-foreground">รหัสบริษัท / รหัสสาขา *</label>
                    <Input
                      value={formCode}
                      onChange={(e) => setFormCode(e.target.value)}
                      placeholder="เช่น COMP01, BR01"
                      className="bg-accent/20"
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
                      />
                    </div>
                  )}
                </div>

                {/* Multilingual names */}
                <div className="space-y-3">
                  <label className="text-sm font-semibold text-foreground block">ชื่อบริษัท / สาขา (แยกตามภาษาที่ใช้งาน)</label>
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-4 border p-4 rounded-lg bg-accent/5">
                    {editorLanguages.map((lang) => {
                      const nativeLang = LANGUAGES.find((l) => l.code === lang);
                      return (
                        <div key={lang} className="space-y-1.5">
                          <span className="text-xs font-semibold text-muted-foreground flex items-center gap-1.5">
                            <span className="w-2 h-2 rounded-full bg-primary" />
                            {nativeLang?.name || lang.toUpperCase()} ({lang})
                          </span>
                          <Input
                            value={formNames[lang] || ""}
                            onChange={(e) =>
                              setFormNames((prev) => ({
                                ...prev,
                                [lang]: e.target.value,
                              }))
                            }
                            placeholder="ระบุชื่อภาษาท้องถิ่น"
                          />
                        </div>
                      );
                    })}
                  </div>
                </div>

                 <div className="flex items-center gap-2 pt-2">
                  <input
                    type="checkbox"
                    id="is_active"
                    checked={formIsActive}
                    onChange={(e) => setFormIsActive(e.target.checked)}
                    className="w-4 h-4 text-primary border-gray-300 rounded focus:ring-primary"
                  />
                  <label htmlFor="is_active" className="text-sm font-semibold text-foreground cursor-pointer select-none">
                    เปิดใช้งานในระบบ
                  </label>
                </div>

                <div className="pt-6 border-t mt-4">
                  <SlideToConfirm
                    onConfirm={handleSave}
                    disabled={!formCode.trim()}
                    saving={saving}
                    success={saveSuccess}
                  />
                </div>
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
      {confirmationDialog}
    </div>
  );
}

interface SlideToConfirmProps {
  onConfirm: () => void;
  disabled?: boolean;
  saving?: boolean;
  success?: boolean;
}

function SlideToConfirm({ onConfirm, disabled, saving, success }: SlideToConfirmProps) {
  const [position, setPosition] = useState(0); // 0 to 100
  const [isDragging, setIsDragging] = useState(false);
  const containerRef = React.useRef<HTMLDivElement>(null);
  const startX = React.useRef(0);
  const startPos = React.useRef(0);

  const handleStart = (clientX: number) => {
    if (disabled || saving || success) return;
    setIsDragging(true);
    startX.current = clientX;
    startPos.current = position;
  };

  const handleMove = useCallback((clientX: number) => {
    if (!isDragging || !containerRef.current) return;
    const rect = containerRef.current.getBoundingClientRect();
    const maxDragWidth = rect.width - 48; // 40px handle + 4px padding each side
    if (maxDragWidth <= 0) return;
    const deltaX = clientX - startX.current;
    const deltaPercent = (deltaX / maxDragWidth) * 100;
    const percent = Math.min(Math.max(startPos.current + deltaPercent, 0), 100);
    setPosition(percent);

    if (percent >= 98) {
      setIsDragging(false);
      setPosition(100);
      onConfirm();
    }
  }, [isDragging, onConfirm]);

  const handleEnd = useCallback(() => {
    if (!isDragging) return;
    setIsDragging(false);
    if (position < 98) {
      setPosition(0);
    }
  }, [isDragging, position]);

  useEffect(() => {
    const onMouseMove = (e: MouseEvent) => handleMove(e.clientX);
    const onTouchMove = (e: TouchEvent) => {
      if (e.touches.length > 0) {
        handleMove(e.touches[0].clientX);
      }
    };
    const onMouseUp = () => handleEnd();
    const onTouchEnd = () => handleEnd();

    if (isDragging) {
      window.addEventListener("mousemove", onMouseMove);
      window.addEventListener("mouseup", onMouseUp);
      window.addEventListener("touchmove", onTouchMove, { passive: true });
      window.addEventListener("touchend", onTouchEnd);
    }

    return () => {
      window.removeEventListener("mousemove", onMouseMove);
      window.removeEventListener("mouseup", onMouseUp);
      window.removeEventListener("touchmove", onTouchMove);
      window.removeEventListener("touchend", onTouchEnd);
    };
  }, [isDragging, handleMove, handleEnd]);

  useEffect(() => {
    if (!saving && !success) {
      setPosition(0);
    }
  }, [saving, success]);

  return (
    <div
      ref={containerRef}
      className={cn(
        "relative flex items-center justify-center h-12 w-full rounded-full overflow-hidden select-none transition-all duration-300 border",
        success
          ? "bg-emerald-500 border-emerald-500 text-white shadow-lg shadow-emerald-500/20"
          : disabled
          ? "bg-accent/30 border-muted text-muted-foreground/60 cursor-not-allowed"
          : "bg-accent border-primary/10 text-accent-foreground shadow-inner"
      )}
      onMouseDown={(e) => handleStart(e.clientX)}
      onTouchStart={(e) => {
        if (e.touches.length > 0) {
          handleStart(e.touches[0].clientX);
        }
      }}
    >
      {/* Dynamic Background Fill */}
      {!success && !disabled && (
        <div
          className="absolute left-0 top-0 bottom-0 bg-primary/20 rounded-l-full pointer-events-none transition-all duration-75"
          style={{ width: `calc(${position}% + 20px)` }}
        />
      )}

      {/* Label Text */}
      <span className="relative z-10 text-xs sm:text-sm font-semibold pointer-events-none transition-colors duration-300">
        {saving && (
          <span className="flex items-center gap-2">
            <Loader2 className="w-4 h-4 animate-spin" />
            กำลังบันทึกข้อมูล...
          </span>
        )}
        {success && (
          <span className="flex items-center gap-2 text-white">
            <Check className="w-4 h-4 animate-bounce" />
            บันทึกข้อมูลสำเร็จ
          </span>
        )}
        {!saving && !success && (
          <span className={cn(isDragging && "opacity-40 transition-opacity")}>
            {disabled ? "กรุณากรอกรหัสข้อมูลเพื่อเปิดใช้งาน" : "เลื่อนเพื่อยืนยันการบันทึก"}
          </span>
        )}
      </span>

      {/* Slide Handle */}
      {!success && !disabled && (
        <div
          className={cn(
            "absolute top-1 bottom-1 w-10 h-10 rounded-full flex items-center justify-center bg-primary text-primary-foreground shadow-md cursor-grab active:cursor-grabbing transition-all duration-75",
            isDragging && "scale-105 shadow-lg shadow-primary/30"
          )}
          style={{
            left: `calc(4px + ${position}% - ${position * 0.48}px)`,
          }}
        >
          <ArrowRight className="w-5 h-5 animate-pulse" />
        </div>
      )}
    </div>
  );
}
