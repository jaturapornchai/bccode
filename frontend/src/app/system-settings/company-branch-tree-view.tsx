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
} from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import type { LanguageCode } from "@/lib/i18n";
import { cn } from "@/lib/utils";
import { normalizeLanguageConfigs } from "./system-settings-screen";
import { deriveMainApiUrl } from "@/lib/backend-url";
import { NamesEditor } from "@/components/product-barcode/names-editor";
import { isThaiHeadOfficeBranchCode, normalizeThaiTaxBranchCode } from "@/lib/thai-branch-code";

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
type ConfirmAction = "save" | "delete";
type OrganizationFormType = "edit_company" | "edit_branch" | "create_company" | "create_branch";

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
  const [companies, setCompanies] = useState<CompanyRecord[]>([]);
  const [branches, setBranches] = useState<BranchRecord[]>([]);
  const [loading, setLoading] = useState(false);
  const [saving, setSaving] = useState(false);
  const [saveSuccess, setSaveSuccess] = useState(false);
  const [saveError, setSaveError] = useState("");
  const [deleteError, setDeleteError] = useState("");

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
  const [formNames, setFormNames] = useState<LocalizedNameEntry[]>([]);

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

    setFormCode(selectedNode.data.code || "");
    setFormIsActive(selectedNode.data.is_active !== false);
    setFormNames(list);

    if (selectedNode.type === "company") {
      setFormTaxId((selectedNode.data as CompanyRecord).tax_id || "");
    }
  }, [selectedNode, editorLanguages]);

  // Handle Save
  const handleSave = async () => {
    if (!auth || !selectedNode || !formType) return;
    setSaving(true);
    setSaveError("");

    try {
      const namesList = formNames;
      const normalizedBranchCode = formType.includes("branch") ? normalizeThaiTaxBranchCode(formCode) : "";

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
          code: normalizedBranchCode,
          names: namesList,
          is_active: formIsActive,
        };
      } else if (formType === "edit_branch") {
        url = `${mainApiUrl}/organization/branch/${selectedNode.guid_fixed}`;
        method = "PUT";
        body = {
          company_guid: selectedNode.company_guid,
          code: normalizedBranchCode,
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

      const json = (await res.json().catch(() => ({}))) as { success?: boolean; id?: string; message?: string };
      if (!res.ok || json.success === false) {
        setSaveError(saveErrorMessage(json.message, formType));
        return;
      }
      if (json.success) {
        setSaveSuccess(true);
        setTimeout(() => setSaveSuccess(false), 2000);
        await loadData();
        if (formType.startsWith("create")) {
          const createdCode = formType === "create_branch" ? normalizedBranchCode : formCode;
          const createdData = {
            guid_fixed: json.id,
            code: createdCode,
            names: namesList,
            is_active: formIsActive,
            ...(formType === "create_company" ? { tax_id: formTaxId } : {}),
          };
          if (formType === "create_company") {
            setCompanies((prev) => prev.some((row) => row.guid_fixed === json.id) ? prev : [...prev, createdData]);
          } else {
            setBranches((prev) => prev.some((row) => row.guid_fixed === json.id) ? prev : [
              ...prev,
              { ...createdData, company_guid: selectedNode.company_guid },
            ]);
          }
          // Select newly created node
          setSelectedNode({
            type: formType === "create_company" ? "company" : "branch",
            guid_fixed: json.id,
            company_guid: selectedNode.company_guid,
            data: createdData,
          });
          setFormType(formType === "create_company" ? "edit_company" : "edit_branch");
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
    if (!auth || !node.guid_fixed) return;

    setLoading(true);
    try {
      const url = `${mainApiUrl}/organization/${node.type}/${node.guid_fixed}`;
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
          {deleteError && (
            <div className="mb-3 rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs font-semibold text-destructive">
              {deleteError}
            </div>
          )}

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
                      <div className="flex items-center gap-1">
                        <Button
                          size="icon"
                          variant="ghost"
                          className="w-7 h-7 text-sky-500 hover:text-sky-600 hover:bg-sky-500/10"
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
                          className="w-7 h-7 text-muted-foreground/60 hover:text-destructive hover:bg-destructive/10"
                          onClick={(e) => {
                            e.stopPropagation();
                            showConfirmCodeDialog("delete", {
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
                          const cannotDeleteBranch = isThaiHeadOfficeBranchCode(br.code) || compBranches.length <= 1;

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
                              <div className="flex items-center gap-1">
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
              <div className="flex items-center justify-between border-b pb-4 gap-4">
                <div>
                  <h3 className="text-xl font-bold text-foreground flex items-center gap-2">
                    {formType.includes("company") ? (
                      <Building2 className="w-5.5 h-5.5 text-primary shrink-0" />
                    ) : (
                      <GitBranch className="w-5.5 h-5.5 text-sky-500 shrink-0" />
                    )}
                    {formType === "create_company" && "เพิ่มบริษัทใหม่"}
                    {formType === "edit_company" && "แก้ไขข้อมูลบริษัท"}
                    {formType === "create_branch" && "เพิ่มสาขาใหม่"}
                    {formType === "edit_branch" && "แก้ไขข้อมูลสาขา"}
                  </h3>
                  <p className="text-sm text-muted-foreground mt-1">
                    {formType.startsWith("create") ? "ระบุข้อมูลรายละเอียดหลักเพื่อเพิ่มข้อมูลเข้าระบบ" : "แก้ไขรายละเอียดข้อมูลและบันทึกประวัติ"}
                  </p>
                </div>
                {formType === "edit_company" && selectedNode?.guid_fixed && (
                  <Button
                    size="sm"
                    variant="outline"
                    className="gap-1 text-xs border-sky-500/30 text-sky-600 hover:bg-sky-500/10 hover:text-sky-700 font-bold shrink-0"
                    onClick={() => {
                      setSelectedNode({
                        type: "branch",
                        company_guid: selectedNode.guid_fixed,
                        data: {},
                      });
                      setFormType("create_branch");
                    }}
                  >
                    <Plus className="w-3.5 h-3.5" />
                    เพิ่มสาขาในบริษัทนี้
                  </Button>
                )}
              </div>

              <div className="space-y-4">
                <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                  <div className="space-y-2">
                    <label className="text-sm font-semibold text-foreground">
                      {formType.includes("company") ? "รหัสบริษัท *" : "รหัสสาขา *"}
                    </label>
                    <Input
                      value={formCode}
                      onChange={(e) => setFormCode(e.target.value)}
                      placeholder={formType.includes("company") ? "ระบุรหัสบริษัท เช่น 00000" : "ระบุรหัสสาขา 5 หลัก เช่น 00001"}
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
                  <NamesEditor
                    names={formNames}
                    onChange={setFormNames}
                    languages={editorLanguages}
                    label={formType.includes("company") ? "ชื่อบริษัท" : "ชื่อสาขา"}
                    language={language}
                  />
                </div>
                {saveError && (
                  <div className="rounded-lg border border-destructive/30 bg-destructive/10 px-3 py-2 text-xs font-semibold text-destructive">
                    {saveError}
                  </div>
                )}

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
                  <Button
                    onClick={() => showConfirmCodeDialog("save")}
                    disabled={!formCode.trim() || saving || saveSuccess}
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
          <div className="bg-card border rounded-2xl w-full max-w-sm p-6 shadow-2xl space-y-4 animate-in fade-in zoom-in duration-200">
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
    case "company_guid is required":
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
