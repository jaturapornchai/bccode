"use client";

import { useEffect, useMemo, useState, useCallback } from "react";
import {
  FileText,
  Plus,
  Pencil,
  Trash2,
  Save,
  X,
  Printer,
  Search,
  AlertCircle,
  Info,
  Clock,
} from "lucide-react";
import { ErpDocEventsDialog } from "@/components/erp-doc-events-dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
import { Badge } from "@/components/ui/badge";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { ResizableSplitter, useSplitPercent } from "@/components/ui/resizable-splitter";
import { type LanguageCode } from "@/lib/i18n";
import { flattenMenuItems, menuText } from "@/lib/menu-data";
import { useBackendLanguage, backendText } from "@/lib/backend-language";
import { getAuthSession, restoreAuthSession } from "@/lib/client-auth-session";
import { useFormShortcuts } from "@/hooks/use-form-shortcuts";
import { useTabularEnterNav } from "@/hooks/use-tabular-enter-nav";
import { SmartBreadcrumb } from "@/components/smart-breadcrumb";
import { ThaiDocumentPrintModal, type ThaiDocPrintType } from "@/components/thai-document-print-modal";
import {
  type ErpTransactionDoc,
  type ErpDetailItem,
  erpDocLifecycle,
  erpDocTotals,
  erpLineAmount,
  erpMoneyDisplay,
  getErpModuleConfig,
  fetchErpTransactions,
  saveErpTransaction,
  deleteErpTransaction,
} from "@/lib/erp-transaction";
import { ERP_AUTO_POSTING_MODULES } from "@/lib/erp-transaction-modules";
import { formatAmount } from "@/lib/general-ledger";
import { useScreenActions } from "@/lib/use-screen-actions";
import type { AuthSession, WorkspaceSession } from "@/lib/workspace-models";

// The screen title is the menu label the user clicked, so it follows every language the menu does.
const MENU_LABEL_BY_ROUTE = new Map(flattenMenuItems().map((item) => [item.route, item.label]));

interface ErpCrudWorkbenchProps {
  route: string;
  embedded?: boolean;
  language?: LanguageCode;
  auth?: AuthSession | null;
  workspace?: WorkspaceSession | null;
}

// ผลลงบัญชีที่ backend แจ้งกลับตอนบันทึก — ข้อความอยู่ใน languages.tsv (erp_gl_notice_*)
function glNoticeFor(status?: string, docno?: string): { key: string; fallback: string; docno?: string } | null {
  if (status === "posted") return { key: "erp_gl_notice_posted", fallback: "ลงบัญชีแล้ว เลขที่ใบสำคัญ {0}", docno };
  if (status === "no_mapping") {
    return {
      key: "erp_gl_notice_no_mapping",
      fallback: "บันทึกเอกสารแล้ว แต่ยังไม่ได้ลงบัญชี เพราะยังไม่ได้ตั้งรูปแบบลงบัญชีของเอกสารประเภทนี้ (ตั้งได้ที่เมนูการจับคู่บัญชี)",
    };
  }
  if (status === "no_amount") {
    return { key: "erp_gl_notice_no_amount", fallback: "บันทึกเอกสารแล้ว แต่ไม่ได้ลงบัญชี เพราะยอดรวมของเอกสารเป็นศูนย์" };
  }
  return null;
}

export function ErpCrudWorkbench({
  route,
  embedded = false,
  language = "th",
  auth = null,
  workspace = null,
}: ErpCrudWorkbenchProps) {
  // สิทธิ์ <เมนู>:create|update|delete ของผู้ใช้บนจอนี้ — backend บังคับจริงด้วย 403 (UI แค่ไม่เสนอปุ่มที่ทำไม่ได้)
  const screenActions = useScreenActions(auth, workspace, route);
  const [backendUrl, setBackendUrl] = useState(() => getAuthSession()?.backendUrl ?? "");
  useEffect(() => {
    if (backendUrl) return;
    let active = true;
    void restoreAuthSession().then((session) => {
      if (active && session) setBackendUrl(session.backendUrl);
    });
    return () => {
      active = false;
    };
  }, [backendUrl]);
  const dictionary = useBackendLanguage(language, backendUrl || undefined);
  const [alertKey, setAlertKey] = useState<string | null>(null);
  // ข้อความอธิบายจาก backend (แปลตามภาษาแล้ว) เมื่อบันทึก/ลบถูกปฏิเสธด้วยเหตุผลทางบัญชี เช่น ปีบัญชีปิด เลขที่ซ้ำ ถูกตัดหนี้แล้ว
  const [alertDetail, setAlertDetail] = useState<string | null>(null);
  const [glNotice, setGlNotice] = useState<{ key: string; fallback: string; docno?: string } | null>(null);
  const showAlert = (key: string | null, detail?: string) => {
    setAlertKey(key);
    setAlertDetail(detail ?? null);
  };
  const moduleUnavailable = alertKey === "module_not_available";
  const config = useMemo(() => getErpModuleConfig(route), [route]);

  const [items, setItems] = useState<ErpTransactionDoc[]>([]);
  const [selectedDoc, setSelectedDoc] = useState<ErpTransactionDoc | null>(null);
  const [isEditing, setIsEditing] = useState(false);
  const [isDirty, setIsDirty] = useState(false);
  const [loading, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [statusFilter, setStatusFilter] = useState<string>("all");

  // Form State
  const [formDoc, setFormDoc] = useState<Partial<ErpTransactionDoc>>({});
  const [showPrintModal, setShowPrintModal] = useState(false);
  const [showEventsModal, setShowEventsModal] = useState(false);

  const docPrintType: ThaiDocPrintType = useMemo(() => {
    if (route.includes("purchase") || route.includes("rfq")) return "purchase_order";
    if (route.includes("pay") || route.includes("expense")) return "payment_voucher";
    if (route.includes("delivery") || route.includes("saleorder")) return "delivery_order";
    return "tax_invoice";
  }, [route]);

  const { confirm, confirmationDialog } = useConfirmDialog({
    defaultConfirmLabel: backendText(dictionary, "confirm", "ยืนยัน"),
    defaultCancelLabel: backendText(dictionary, "cancel", "ยกเลิก"),
  });

  // Splitter state persistence (localStorage)
  const {
    splitPercent: splitterWidth,
    isResizing,
    startResize,
    resetSplit,
    adjustWithKeyboard,
  } = useSplitPercent({
    defaultLeft: 38,
    storageKey: "bc_erp_crud_splitter_width",
    min: 25,
    max: 60,
  });

  // Load Data
  const loadData = useCallback(async () => {
    if (!config) return;
    setLoading(true);
    try {
      const res = await fetchErpTransactions(config, { q: searchQuery });
      if (res.error) {
        showAlert(res.error);
        setItems([]);
        return;
      }
      setItems(res.items);
      showAlert(null);
      if (res.items.length > 0 && !selectedDoc && !isEditing) {
        setSelectedDoc(res.items[0]);
      }
    } finally {
      setLoading(false);
    }
  }, [config, searchQuery, selectedDoc, isEditing]);

  useEffect(() => {
    loadData();
  }, [loadData]);

  // Dirty Guard for Row Selection
  const handleSelectRow = async (doc: ErpTransactionDoc) => {
    if (selectedDoc?.id === doc.id && !isEditing) return;

    if (isDirty) {
      const ok = await confirm({
        title: backendText(dictionary, "unsaved_changes", "มีการเปลี่ยนแปลงที่ยังไม่ได้บันทึก"),
        description: backendText(
          dictionary,
          "unsaved_switch_warning",
          "ท่านมีการแก้ไขข้อมูลค้างอยู่ การเปลี่ยนรายการจะละทิ้งการเปลี่ยนแปลง ต้องการดำเนินการต่อหรือไม่?",
        ),
        confirmLabel: backendText(dictionary, "discard_and_switch", "ละทิ้งและเปลี่ยนรายการ"),
      });
      if (!ok) return;
    }

    setSelectedDoc(doc);
    setIsEditing(false);
    setIsDirty(false);
  };

  // Start Edit Mode explicitly from Pencil
  const handleStartEdit = (doc: ErpTransactionDoc) => {
    if (!screenActions.update || erpDocLifecycle(doc) !== "draft") return;
    setSelectedDoc(doc);
    setFormDoc(JSON.parse(JSON.stringify(doc)));
    setIsEditing(true);
    setIsDirty(false);
  };

  // Create New Document
  const handleStartCreate = async () => {
    if (!screenActions.create) return;
    if (isDirty) {
      const ok = await confirm({
        title: backendText(dictionary, "unsaved_changes", "มีการเปลี่ยนแปลงที่ยังไม่ได้บันทึก"),
        description: backendText(
          dictionary,
          "discard_and_create_new_question",
          "ละทิ้งการเปลี่ยนแปลงปัจจุบัน และสร้างเอกสารใหม่?",
        ),
        confirmLabel: backendText(dictionary, "create_new", "สร้างใหม่"),
      });
      if (!ok) return;
    }

    const today = new Date().toISOString().split("T")[0];
    const initialNew: Partial<ErpTransactionDoc> = {
      docno: "",
      docdatetime: `${today}T09:00:00Z`,
      custcode: "",
      custname: "",
      description: "",
      totalamount: "0",
      totalbeforevat: "0",
      totalvatvalue: "0",
      vattype: 1,
      vatrate: "7",
      status: 0,
      details: config?.hasLineItems
        ? [
            {
              linenumber: 1,
              itemcode: "",
              itemname: "",
              unitcode: "PCS",
              qty: "1",
              price: "0",
              sumamount: "0",
            },
          ]
        : [],
    };
    setFormDoc(initialNew);
    setIsEditing(true);
    setIsDirty(false);
  };

  // Cancel Edit
  const handleCancelEdit = async () => {
    if (isDirty) {
      const ok = await confirm({
        title: backendText(dictionary, "discard_changes_question", "ยกเลิกการแก้ไข?"),
        description: backendText(
          dictionary,
          "discard_edits_confirm",
          "ข้อมูลที่ท่านแก้ไขจะถูกละทิ้ง ยืนยันการยกเลิก?",
        ),
        confirmLabel: backendText(dictionary, "confirm_discard", "ยืนยันยกเลิก"),
      });
      if (!ok) return;
    }
    setIsEditing(false);
    setIsDirty(false);
  };

  // เอกสารที่อนุมัติหรือยกเลิกแล้ว backend ไม่ให้แก้ไข/ลบ (erp_doc_locked) — ว่าง = ยังแก้ไขได้
  const lockedHintFor = (doc: Parameters<typeof erpDocLifecycle>[0]): string => {
    const lifecycle = erpDocLifecycle(doc);
    if (lifecycle === "cancelled") {
      return backendText(dictionary, "erp_doc_locked_cancelled_hint", "เอกสารนี้ถูกยกเลิกแล้ว จึงแก้ไขหรือลบไม่ได้");
    }
    if (lifecycle === "approved") {
      return backendText(
        dictionary,
        "erp_doc_locked_approved_hint",
        "เอกสารนี้อนุมัติแล้ว จึงแก้ไขหรือลบไม่ได้ ต้องยกเลิกการอนุมัติก่อน",
      );
    }
    return "";
  };

  const lifecycleBadge = (doc: Parameters<typeof erpDocLifecycle>[0], approvedLabel: string) => {
    const lifecycle = erpDocLifecycle(doc);
    if (lifecycle === "cancelled") {
      return { label: backendText(dictionary, "cancelled", "ยกเลิกแล้ว"), className: "bg-destructive/15 text-destructive" };
    }
    if (lifecycle === "approved") {
      return { label: approvedLabel, className: "bg-emerald-500/15 text-emerald-600 dark:text-emerald-400" };
    }
    return {
      label: backendText(dictionary, "draft", "ฉบับร่าง"),
      className: "bg-amber-500/15 text-amber-600 dark:text-amber-400",
    };
  };

  // Delete Document
  const handleDeleteDoc = async (doc: ErpTransactionDoc) => {
    if (!config || !doc.id || !screenActions.delete || erpDocLifecycle(doc) !== "draft") return;
    const ok = await confirm({
      title: backendText(dictionary, "delete_document_confirm_title", "ยืนยันการลบเอกสาร {0}?").replace(
        "{0}",
        doc.docno ?? "",
      ),
      description: backendText(
        dictionary,
        "delete_cannot_undo_confirm",
        "การลบรายการนี้จะไม่สามารถกู้คืนได้ ยืนยันที่จะลบข้อมูลหรือไม่?",
      ),
      confirmLabel: backendText(dictionary, "delete_document", "ลบเอกสาร"),
    });
    if (!ok) return;

    setLoading(true);
    const res = await deleteErpTransaction(config, doc.id, language);
    if (!res.success) {
      showAlert(res.message ?? "delete_failed", res.detail);
      setLoading(false);
      return;
    }
    if (selectedDoc?.id === doc.id) {
      setSelectedDoc(null);
      setIsEditing(false);
      setIsDirty(false);
    }
    showAlert(null);
    setGlNotice(null);
    await loadData();
  };

  // ยอดบรรทัด/ยอดรวมระหว่างแก้เป็นตัวอย่างทศนิยมตรงตัว (erpLineAmount/erpDocTotals) — หลังบันทึกแสดงยอดที่ backend ตอบกลับ
  const updateLineItem = (index: number, patch: Partial<ErpDetailItem>) => {
    if (!formDoc.details) return;
    const nextDetails = [...formDoc.details];
    const item = { ...nextDetails[index], ...patch };
    item.sumamount = erpLineAmount(item);
    nextDetails[index] = item;

    setFormDoc({
      ...formDoc,
      details: nextDetails,
      ...erpDocTotals(nextDetails, formDoc.vattype, formDoc.vatrate),
    });
    setIsDirty(true);
  };

  const tabularEnterNav = useTabularEnterNav({
    onAddNewRow: () => {
      addLineItem();
    },
  });

  const addLineItem = () => {
    const details = formDoc.details || [];
    const newItem: ErpDetailItem = {
      linenumber: details.length + 1,
      itemcode: "",
      itemname: "",
      unitcode: "PCS",
      qty: "1",
      price: "0",
      sumamount: "0",
    };
    setFormDoc((prev) => ({ ...prev, details: [...(prev.details || []), newItem] }));
    setIsDirty(true);
    requestAnimationFrame(() => {
      const container = document.querySelector(".erp-detail-table-container");
      if (!container) return;
      const rows = container.querySelectorAll("tbody tr");
      const lastRow = rows[rows.length - 1];
      if (lastRow) {
        const firstInput = lastRow.querySelector<HTMLInputElement>("input:not([disabled]):not([readonly])");
        firstInput?.focus?.();
        firstInput?.select?.();
      }
    });
  };

  const removeLineItem = (index: number) => {
    if (!formDoc.details || formDoc.details.length <= 1) return;
    const nextDetails = formDoc.details
      .filter((_, i) => i !== index)
      .map((item, i) => ({ ...item, linenumber: i + 1 }));

    setFormDoc({
      ...formDoc,
      details: nextDetails,
      ...erpDocTotals(nextDetails, formDoc.vattype, formDoc.vatrate),
    });
    setIsDirty(true);
  };

  // Save Document
  const handleSaveDoc = async () => {
    if (!config) return;
    if (!(formDoc.id ? screenActions.update : screenActions.create)) return;
    // เอกสารใหม่ปล่อยว่างได้ — backend ออกเลขที่ให้ตอนบันทึก; เอกสารเดิมต้องมีเลขที่
    if (formDoc.id && !formDoc.docno?.trim()) {
      showAlert("docno_required");
      return;
    }

    setLoading(true);
    const res = await saveErpTransaction(config, formDoc, Boolean(formDoc.id), language);
    setLoading(false);

    if (!res.success) {
      showAlert(res.message ?? "save_failed", res.detail);
      return;
    }

    if (res.data) {
      setSelectedDoc(res.data);
    }
    setIsEditing(false);
    setIsDirty(false);
    showAlert(null);
    setGlNotice(glNoticeFor(res.data?.glstatus, res.data?.gldocno));
    await loadData();
  };

  // Global Keyboard Shortcuts (Ctrl+S, Alt+N, Esc)
  useFormShortcuts({
    onSave: handleSaveDoc,
    onNew: screenActions.create ? handleStartCreate : undefined,
    onCancel: isEditing ? handleCancelEdit : undefined,
    canSave: isEditing && !loading,
    disabled: loading || moduleUnavailable,
  });

  // Filtered list
  const filteredItems = useMemo(() => {
    return items.filter((item) => {
      if (statusFilter === "draft" && erpDocLifecycle(item) !== "draft") return false;
      if (statusFilter === "approved" && erpDocLifecycle(item) !== "approved") return false;
      if (!searchQuery.trim()) return true;
      const q = searchQuery.toLowerCase();
      return (
        item.docno.toLowerCase().includes(q) ||
        (item.custname && item.custname.toLowerCase().includes(q)) ||
        (item.custcode && item.custcode.toLowerCase().includes(q)) ||
        (item.description && item.description.toLowerCase().includes(q))
      );
    });
  }, [items, statusFilter, searchQuery]);

  const selectedBadge = lifecycleBadge(selectedDoc ?? {}, backendText(dictionary, "approved_status", "อนุมัติเรียบร้อย"));
  const selectedLockedHint = lockedHintFor(selectedDoc ?? {});

  if (!config) {
    return (
      <div className="p-8 text-center text-muted-foreground">
        <AlertCircle className="mx-auto mb-2 size-8 text-warning" />
        {backendText(dictionary, "not_found", "ไม่พบการตั้งค่าสำหรับหน้านี้")}
      </div>
    );
  }

  const titleEn = config.title.en;
  const menuLabel = MENU_LABEL_BY_ROUTE.get(route);
  const title = menuLabel ? menuText(menuLabel, language, dictionary) : config.title[language === "en" ? "en" : "th"];
  const cpLabel = backendText(dictionary, config.counterpartyKey, config.counterpartyLabel.th);

  const containerClass = embedded
    ? "h-full min-h-0 flex flex-col overflow-hidden text-[0.95rem]"
    : "mx-auto max-w-[1800px] p-4 min-h-[calc(100dvh-2rem)] flex flex-col text-[0.95rem]";

  return (
    <main className={containerClass} data-erp-module={config.moduleKey}>
      {confirmationDialog}

      <ThaiDocumentPrintModal
        open={showPrintModal}
        onClose={() => setShowPrintModal(false)}
        doc={selectedDoc}
        docType={docPrintType}
      />

      <ErpDocEventsDialog
        module={config.apiPath}
        docno={selectedDoc?.docno ?? ""}
        isOpen={showEventsModal}
        onClose={() => setShowEventsModal(false)}
        language={language}
      />

      {!embedded && <SmartBreadcrumb currentTitle={title} className="mb-3" />}

      {/* Top Header Toolbar */}
      <header className="shrink-0 flex flex-wrap items-center justify-between gap-3 border-b border-border/60 pb-3 mb-2">
        <div className="flex items-center gap-3">
          <div className="h-10 w-10 rounded-xl bg-primary/10 flex items-center justify-center text-primary font-bold text-lg">
            <FileText className="size-5" />
          </div>
          <div>
            <h1 className="text-xl font-bold text-foreground leading-tight">
              {title}
            </h1>
            <p className="text-xs text-muted-foreground">
              {titleEn} &bull; {backendText(dictionary, "system_standard", "ระบบมาตรฐานการบัญชี ERP")}
            </p>
          </div>
        </div>

        <div className="flex items-center gap-2">
          {moduleUnavailable || !screenActions.create ? null : (
            <Button
              onClick={handleStartCreate}
              className="h-10 bg-primary text-primary-foreground font-semibold px-4 shadow-sm gap-1.5"
            >
              <Plus className="size-4" />
              <span>{"+ " + backendText(dictionary, "create_new_document", "สร้างเอกสารใหม่")}</span>
              <kbd className="ml-1 hidden sm:inline-block rounded border border-primary-foreground/30 bg-primary-foreground/15 px-1.5 py-0.5 text-[10px] font-mono text-primary-foreground">
                Alt+N
              </kbd>
            </Button>
          )}
        </div>
      </header>

      {moduleUnavailable ? (
        // Not an error the user caused or can retry: this document type has no
        // endpoint yet. A red "โหลดข้อมูลไม่สำเร็จ" would send them hunting.
        <div
          role="status"
          className="mb-3 shrink-0 flex items-start gap-3 rounded-xl border border-border bg-muted/40 px-4 py-3 text-sm text-muted-foreground"
        >
          <Info className="mt-0.5 size-5 shrink-0" />
          <p className="flex-1">
            {backendText(
              dictionary,
              "module_not_available",
              "จอนี้ยังไม่เปิดใช้งาน ระบบยังไม่รองรับเอกสารชนิดนี้ กรุณาติดต่อผู้ดูแลระบบ",
            )}
          </p>
        </div>
      ) : null}

      {!moduleUnavailable && !ERP_AUTO_POSTING_MODULES.includes(config.apiPath) ? (
        // เอกสารชนิดนี้ backend เก็บอย่างเดียว ไม่ลงบัญชี/ตัดสต็อก — บอกผู้ใช้ตรง ๆ กันเข้าใจผิดว่าลงบัญชีแล้ว
        <div
          role="status"
          className="mb-3 shrink-0 flex items-start gap-3 rounded-xl border border-border bg-muted/40 px-4 py-3 text-[0.95rem] leading-relaxed text-foreground"
        >
          <Info className="mt-0.5 size-5 shrink-0" />
          <p className="flex-1">
            {backendText(
              dictionary,
              "erp_store_only_notice",
              "เอกสารชนิดนี้บันทึกเก็บไว้เท่านั้น ยังไม่ลงบัญชีและไม่ตัดสต็อกอัตโนมัติ",
            )}
          </p>
        </div>
      ) : null}

      {glNotice ? (
        // เอกสารบันทึกแล้วแต่ไม่ได้ลงบัญชี (ยังไม่ตั้งรูปแบบลงบัญชี/ยอดเป็นศูนย์) หรือลงบัญชีแล้ว — บอกผลตรง ๆ ไม่ให้เข้าใจผิด
        <div
          role="status"
          className="mb-3 shrink-0 flex items-start gap-3 rounded-xl border border-border bg-muted/40 px-4 py-3 text-[0.95rem] leading-relaxed text-foreground"
        >
          <Info className="mt-0.5 size-5 shrink-0" />
          <p className="flex-1">
            {backendText(dictionary, glNotice.key, glNotice.fallback).replace("{0}", glNotice.docno ?? "")}
          </p>
          <button
            type="button"
            aria-label="ปิดข้อความแจ้งเตือน"
            title="ปิดข้อความแจ้งเตือน"
            onClick={() => setGlNotice(null)}
            className="rounded-md p-1 hover:opacity-70 focus:outline-none"
          >
            <X className="size-4" />
          </button>
        </div>
      ) : null}

      {alertKey && !moduleUnavailable ? (
        <div
          role="alert"
          className="mb-3 shrink-0 flex items-start gap-3 rounded-xl border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive"
        >
          <AlertCircle className="mt-0.5 size-5 shrink-0" />
          <p className="flex-1">
            {alertDetail ?? backendText(
              dictionary,
              alertKey,
              alertKey === "save_failed"
                ? "บันทึกข้อมูลไม่สำเร็จ กรุณาลองใหม่อีกครั้ง"
                : alertKey === "delete_failed"
                  ? "ลบเอกสารไม่สำเร็จ กรุณาลองใหม่อีกครั้ง"
                  : alertKey === "load_failed"
                    ? "โหลดข้อมูลไม่สำเร็จ"
                    : alertKey === "connection_error"
                      ? "เชื่อมต่อเซิร์ฟเวอร์ไม่ได้ กรุณาตรวจสอบอินเทอร์เน็ต"
                      : alertKey === "unauthorized"
                        ? "ไม่มีสิทธิ์เข้าถึงข้อมูล"
                        : alertKey === "erp_err_read_forbidden"
                          ? "ไม่มีสิทธิ์เปิดดูเอกสารประเภทนี้"
                          : alertKey === "erp_err_action_forbidden"
                            ? "ไม่มีสิทธิ์เพิ่ม แก้ไข หรือลบเอกสารประเภทนี้"
                        : alertKey === "docno_required"
                          ? "กรุณาระบุเลขที่เอกสาร"
                          : "เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง",
            )}
          </p>
          <button
            type="button"
            aria-label="ปิดข้อความแจ้งเตือน"
            title="ปิดข้อความแจ้งเตือน"
            onClick={() => showAlert(null)}
            className="rounded-md p-1 text-destructive hover:opacity-70 focus:outline-none"
          >
            <X className="size-4" />
          </button>
        </div>
      ) : null}

      {/* Master-Detail Split Pane */}
      <div className="flex-1 min-h-0 flex flex-col lg:flex-row overflow-hidden border border-border/60 rounded-xl bg-background shadow-xs">
        {/* Left Data List Pane */}
        <section
          style={{ width: `${splitterWidth}%` }}
          className="min-h-0 flex flex-col overflow-hidden border-b lg:border-b-0 lg:border-r border-border/60 bg-card"
          aria-label={backendText(dictionary, "document_list", "รายการเอกสาร")}
        >
          {/* .bc-list-toolbar */}
          <div className="bc-list-toolbar shrink-0 p-2.5 border-b border-border/60 bg-muted/20 flex flex-col gap-2">
            <div className="flex items-center justify-between gap-2">
              <span className="text-xs font-semibold text-foreground">
                {backendText(dictionary, "all_records", "รายการทั้งหมด")}:{" "}
                <strong className="text-primary">{filteredItems.length}</strong>
              </span>
              <div className="flex items-center gap-1">
                <button
                  type="button"
                  onClick={() => setStatusFilter("all")}
                  className={`px-2 py-0.5 rounded-full text-xs font-medium transition-colors ${
                    statusFilter === "all"
                      ? "bg-primary text-primary-foreground"
                      : "bg-muted text-muted-foreground hover:bg-muted/80"
                  }`}
                >
                  {backendText(dictionary, "all", "ทั้งหมด")}
                </button>
                <button
                  type="button"
                  onClick={() => setStatusFilter("approved")}
                  className={`px-2 py-0.5 rounded-full text-xs font-medium transition-colors ${
                    statusFilter === "approved"
                      ? "bg-emerald-600 text-white"
                      : "bg-muted text-muted-foreground hover:bg-muted/80"
                  }`}
                >
                  {backendText(dictionary, "approve", "อนุมัติ")}
                </button>
                <button
                  type="button"
                  onClick={() => setStatusFilter("draft")}
                  className={`px-2 py-0.5 rounded-full text-xs font-medium transition-colors ${
                    statusFilter === "draft"
                      ? "bg-amber-600 text-white"
                      : "bg-muted text-muted-foreground hover:bg-muted/80"
                  }`}
                >
                  {backendText(dictionary, "draft", "ฉบับร่าง")}
                </button>
              </div>
            </div>

            <div className="relative">
              <Search className="absolute left-2.5 top-2.5 size-4 text-muted-foreground" />
              <Input
                value={searchQuery}
                onChange={(e) => setSearchQuery(e.target.value)}
                placeholder={backendText(dictionary, "search_document_placeholder", "ค้นหาเลขที่, คู่ค้า, รายละเอียด...")}
                className="h-8 pl-8 text-xs bg-background"
              />
            </div>
          </div>

          {/* List Table Container */}
          <div className="flex-1 min-h-0 overflow-y-auto scrollbar-thin">
            <table className="w-full text-left border-collapse">
              {/* .bc-list-header */}
              <thead className="bc-list-header sticky top-0 bg-muted/95 backdrop-blur z-10 border-b border-border/80 text-[0.7rem] font-extrabold uppercase tracking-wider text-muted-foreground">
                <tr>
                  <th className="py-2 px-2.5">{backendText(dictionary, "doc_no_column", "เลขที่")}</th>
                  <th className="py-2 px-2">{backendText(dictionary, "date", "วันที่")}</th>
                  <th className="py-2 px-2">{cpLabel}</th>
                  <th className="py-2 px-2 text-right">{backendText(dictionary, "amount", "ยอดรวม")}</th>
                  <th className="py-2 px-2 text-center">{backendText(dictionary, "status", "สถานะ")}</th>
                  <th className="py-2 px-2 text-center">{backendText(dictionary, "manage", "จัดการ")}</th>
                </tr>
              </thead>
              <tbody className="divide-y divide-border/40">
                {filteredItems.map((item) => {
                  const isSelected = selectedDoc?.id === item.id;
                  const lockedHint = lockedHintFor(item);
                  const badge = lifecycleBadge(item, backendText(dictionary, "approve", "อนุมัติ"));
                  return (
                    <tr
                      key={item.id || item.docno}
                      onClick={() => handleSelectRow(item)}
                      className={`bc-list-row cursor-pointer transition-colors text-[0.75rem] ${
                        isSelected
                          ? "bg-primary/10 font-medium text-foreground"
                          : "hover:bg-muted/40 text-foreground/90"
                      }`}
                      style={{ padding: "3px 8px" }}
                    >
                      <td className="py-2 px-2.5 font-semibold text-primary">{item.docno}</td>
                      <td className="py-2 px-2 text-muted-foreground font-mono text-[0.7rem]">
                        {item.docdatetime?.split("T")[0] || "-"}
                      </td>
                      <td className="py-2 px-2 truncate max-w-[130px]" title={item.custname || item.custcode}>
                        {item.custname || item.custcode || "-"}
                      </td>
                      <td className="py-2 px-2 text-right font-mono font-medium">
                        {formatAmount(item.totalamount || "0")}
                      </td>
                      <td className="py-2 px-2 text-center">
                        <span
                          className={`inline-block px-1.5 py-0.5 rounded text-[0.65rem] font-semibold ${badge.className}`}
                        >
                          {badge.label}
                        </span>
                      </td>
                      <td className="py-2 px-2 text-center">
                        <div className="flex items-center justify-center gap-1">
                          {screenActions.update ? (
                            <button
                              type="button"
                              disabled={!!lockedHint}
                              title={lockedHint || backendText(dictionary, "edit_document", "แก้ไขเอกสาร")}
                              aria-label={lockedHint || backendText(dictionary, "edit_document", "แก้ไขเอกสาร")}
                              onClick={(e) => {
                                e.stopPropagation();
                                handleStartEdit(item);
                              }}
                              className="p-1 rounded text-primary hover:bg-primary/20 transition-colors disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent"
                            >
                              <Pencil className="size-3.5" />
                            </button>
                          ) : null}
                          {screenActions.delete ? (
                            <button
                              type="button"
                              disabled={!!lockedHint}
                              title={lockedHint || backendText(dictionary, "delete_document", "ลบเอกสาร")}
                              aria-label={lockedHint || backendText(dictionary, "delete_document", "ลบเอกสาร")}
                              onClick={(e) => {
                                e.stopPropagation();
                                handleDeleteDoc(item);
                              }}
                              className="p-1 rounded text-destructive hover:bg-destructive/20 transition-colors disabled:cursor-not-allowed disabled:opacity-40 disabled:hover:bg-transparent"
                            >
                              <Trash2 className="size-3.5" />
                            </button>
                          ) : null}
                        </div>
                      </td>
                    </tr>
                  );
                })}
                {filteredItems.length === 0 && !loading && (
                  <tr>
                    <td colSpan={6} className="py-10 text-center text-muted-foreground text-xs">
                      {backendText(dictionary, "no_documents_found", "ไม่พบรายการเอกสาร")}
                    </td>
                  </tr>
                )}
              </tbody>
            </table>
          </div>
        </section>

        {/* Resizable Splitter */}
        <ResizableSplitter
          value={Math.round(splitterWidth)}
          min={25}
          max={60}
          isResizing={isResizing}
          onPointerDown={startResize}
          onDoubleClick={resetSplit}
          onKeyDown={adjustWithKeyboard}
          label={
            backendText(
              dictionary,
              "resize_panes_hint",
              "ปรับความกว้างรายการและรายละเอียด (ลากเพื่อปรับ, ดับเบิ้ลคลิกเพื่อคืนค่า)",
            )
          }
          breakpoint="lg"
        />

        {/* Right Detail / Edit Form Pane */}
        <section
          className="min-h-0 flex-1 flex flex-col overflow-y-auto scrollbar-thin bg-background p-4"
          aria-label={backendText(dictionary, "document_detail_and_form", "รายละเอียดและฟอร์มเอกสาร")}
        >
          {isEditing ? (
            /* ================= EDIT MODE FORM ================= */
            <div className="flex flex-col gap-4">
              {/* Sticky Top Actions Header */}
              <div className="sticky top-0 z-20 flex items-center justify-between bg-background/95 backdrop-blur py-2 border-b border-border/80">
                <div className="flex items-center gap-2">
                  <span className="font-bold text-foreground text-base">
                    {formDoc.id
                      ? `${backendText(dictionary, "edit_document", "แก้ไขเอกสาร")}: ${formDoc.docno}`
                      : (backendText(dictionary, "create_new_document", "สร้างเอกสารใหม่"))}
                  </span>
                  {isDirty && (
                    <Badge variant="outline" className="text-amber-600 border-amber-400 text-xs">
                      {backendText(dictionary, "editing_now", "กำลังแก้ไข")}
                    </Badge>
                  )}
                </div>
                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    onClick={handleCancelEdit}
                    className="h-9 text-sm"
                  >
                    <X className="mr-1 size-4" />
                    {backendText(dictionary, "cancel", "ยกเลิก")}
                  </Button>
                  <Button
                    onClick={handleSaveDoc}
                    className="h-9 bg-primary text-primary-foreground font-semibold px-4"
                  >
                    <Save className="mr-1 size-4" />
                    {backendText(dictionary, "save_document", "บันทึกเอกสาร")}
                  </Button>
                </div>
              </div>

              {/* Form Fields Grid */}
              <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 border border-border/60 rounded-xl p-4 bg-card">
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {backendText(dictionary, "document_no", "เลขที่เอกสาร") + (formDoc.id ? " *" : "")}
                  </label>
                  <Input
                    value={formDoc.docno || ""}
                    onChange={(e) => {
                      setFormDoc({ ...formDoc, docno: e.target.value });
                      setIsDirty(true);
                    }}
                    className="h-9 text-sm"
                    placeholder={backendText(
                      dictionary,
                      "docno_auto_placeholder",
                      "ระบบออกเลขที่เอกสารให้อัตโนมัติเมื่อบันทึก",
                    )}
                  />
                </div>

                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {backendText(dictionary, "document_date", "วันที่เอกสาร")}
                  </label>
                  <Input
                    type="date"
                    value={formDoc.docdatetime?.split("T")[0] || ""}
                    onChange={(e) => {
                      setFormDoc({ ...formDoc, docdatetime: `${e.target.value}T09:00:00Z` });
                      setIsDirty(true);
                    }}
                    className="h-9 text-sm"
                  />
                </div>

                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {cpLabel}
                  </label>
                  <Input
                    value={formDoc.custname || ""}
                    onChange={(e) => {
                      setFormDoc({ ...formDoc, custname: e.target.value });
                      setIsDirty(true);
                    }}
                    className="h-9 text-sm"
                    placeholder={backendText(dictionary, "party_name_placeholder", "ระบุชื่อคู่ค้า / ผู้ติดต่อ")}
                  />
                </div>

                <div className="md:col-span-2 lg:col-span-3">
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">
                    {backendText(dictionary, "description_or_remarks", "คำอธิบาย / หมายเหตุ")}
                  </label>
                  <Input
                    value={formDoc.description || ""}
                    onChange={(e) => {
                      setFormDoc({ ...formDoc, description: e.target.value });
                      setIsDirty(true);
                    }}
                    className="h-9 text-sm"
                    placeholder={backendText(dictionary, "extra_notes_placeholder", "บันทึกข้อความเพิ่มเติม")}
                  />
                </div>
              </div>

              {/* Line Items Section */}
              {config.hasLineItems && (
                <div className="border border-border/60 rounded-xl p-4 bg-card flex flex-col gap-3">
                  <div className="flex items-center justify-between">
                    <span className="font-semibold text-sm text-foreground">
                      {backendText(dictionary, "line_items_title", "รายการสินค้า / บริการ")}
                    </span>
                    <Button
                      type="button"
                      variant="outline"
                      size="sm"
                      onClick={addLineItem}
                      className="h-8 text-xs"
                    >
                      <Plus className="mr-1 size-3.5" />
                      {"+ " + backendText(dictionary, "add_item", "เพิ่มรายการ")}
                    </Button>
                  </div>

                  <div className="overflow-x-auto erp-detail-table-container" onKeyDown={tabularEnterNav.onKeyDown}>
                    <table className="w-full text-left text-xs border-collapse">
                      <thead className="bg-muted/80 text-muted-foreground font-semibold border-b border-border">
                        <tr>
                          <th className="p-2 w-10 text-center">#</th>
                          <th className="p-2 min-w-[140px]">{backendText(dictionary, "item_code_or_barcode", "รหัสสินค้า / บาร์โค้ด")}</th>
                          <th className="p-2 min-w-[200px]">{backendText(dictionary, "product_name", "ชื่อสินค้า")}</th>
                          <th className="p-2 w-20 text-center">{backendText(dictionary, "unit", "หน่วยนับ")}</th>
                          <th className="p-2 w-24 text-right">{backendText(dictionary, "qty", "จำนวน")}</th>
                          <th className="p-2 w-28 text-right">{backendText(dictionary, "price", "ราคา")}</th>
                          <th className="p-2 w-28 text-right">{backendText(dictionary, "line_amount", "จำนวนเงิน")}</th>
                          <th className="p-2 w-12 text-center"></th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-border/40">
                        {(formDoc.details || []).map((item, idx) => (
                          <tr key={idx} className="hover:bg-muted/20">
                            <td className="p-2 text-center text-muted-foreground">{idx + 1}</td>
                            <td className="p-1.5">
                              <Input
                                value={item.itemcode || ""}
                                onChange={(e) => updateLineItem(idx, { itemcode: e.target.value })}
                                className="h-7 text-xs"
                                placeholder="P-001"
                              />
                            </td>
                            <td className="p-1.5">
                              <Input
                                value={item.itemname || ""}
                                onChange={(e) => updateLineItem(idx, { itemname: e.target.value })}
                                className="h-7 text-xs"
                                placeholder={backendText(dictionary, "product_name", "ชื่อสินค้า")}
                              />
                            </td>
                            <td className="p-1.5">
                              <Input
                                value={item.unitcode || "PCS"}
                                onChange={(e) => updateLineItem(idx, { unitcode: e.target.value })}
                                className="h-7 text-xs text-center"
                              />
                            </td>
                            <td className="p-1.5">
                              <Input
                                type="number"
                                value={item.qty}
                                onChange={(e) => updateLineItem(idx, { qty: e.target.value })}
                                className="h-7 text-xs text-right font-mono"
                              />
                            </td>
                            <td className="p-1.5">
                              <Input
                                type="number"
                                value={item.price}
                                onChange={(e) => updateLineItem(idx, { price: e.target.value })}
                                className="h-7 text-xs text-right font-mono"
                              />
                            </td>
                            <td className="p-2 text-right font-mono font-semibold text-foreground">
                              {formatAmount(erpMoneyDisplay(item.sumamount))}
                            </td>
                            <td className="p-1.5 text-center">
                              <button
                                type="button"
                                onClick={() => removeLineItem(idx)}
                                className="p-1 text-destructive hover:bg-destructive/15 rounded transition-colors"
                              >
                                <Trash2 className="size-3.5" />
                              </button>
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>

                  {/* Summary Footer */}
                  <div className="flex flex-col items-end gap-1.5 border-t border-border/80 pt-3 text-xs">
                    <div className="flex justify-between w-64 text-muted-foreground">
                      <span>{backendText(dictionary, "goods_subtotal", "รวมมูลค่าสินค้า") + ":"}</span>
                      <span className="font-mono font-medium text-foreground">
                        {formatAmount(erpMoneyDisplay(formDoc.totalbeforevat))}
                      </span>
                    </div>
                    <div className="flex justify-between w-64 text-muted-foreground">
                      <span>{backendText(dictionary, "vat_7_percent", "ภาษีมูลค่าเพิ่ม (7%)") + ":"}</span>
                      <span className="font-mono font-medium text-foreground">
                        {formatAmount(erpMoneyDisplay(formDoc.totalvatvalue))}
                      </span>
                    </div>
                    <div className="flex justify-between w-64 text-sm font-bold border-t border-border pt-1 text-primary">
                      <span>{backendText(dictionary, "net_total", "รวมทั้งสิ้น") + ":"}</span>
                      <span className="font-mono">
                        {formatAmount(erpMoneyDisplay(formDoc.totalamount))} ฿
                      </span>
                    </div>
                  </div>
                </div>
              )}

              {/* Pinned Bottom Actions */}
              <div className="sticky bottom-0 z-20 flex justify-end gap-2 bg-background/95 backdrop-blur py-3 border-t border-border/80">
                <Button variant="outline" onClick={handleCancelEdit} className="h-10 text-sm px-4 gap-1.5">
                  <span>{backendText(dictionary, "cancel", "ยกเลิก")}</span>
                  <kbd className="hidden sm:inline-block rounded border border-border bg-muted/60 px-1.5 py-0.5 text-[10px] font-mono text-muted-foreground">
                    Esc
                  </kbd>
                </Button>
                <Button onClick={handleSaveDoc} className="h-10 bg-primary text-primary-foreground font-semibold px-5 gap-1.5">
                  <Save className="size-4" />
                  <span>{backendText(dictionary, "save_document", "บันทึกเอกสาร")}</span>
                  <kbd className="hidden sm:inline-block rounded border border-primary-foreground/30 bg-primary-foreground/15 px-1.5 py-0.5 text-[10px] font-mono text-primary-foreground">
                    Ctrl+S
                  </kbd>
                </Button>
              </div>
            </div>
          ) : selectedDoc ? (
            /* ================= READ-ONLY VIEW MODE ================= */
            <div className="flex flex-col gap-4">
              {/* Document View Top Header */}
              <div className="flex flex-wrap items-center justify-between gap-2 border-b border-border/60 pb-3">
                <div className="flex items-center gap-3">
                  <div>
                    <div className="flex items-center gap-2">
                      <span className="text-lg font-bold text-primary font-mono">{selectedDoc.docno}</span>
                      <span className={`px-2 py-0.5 rounded text-xs font-semibold ${selectedBadge.className}`}>
                        {selectedBadge.label}
                      </span>
                    </div>
                    <p className="text-xs text-muted-foreground mt-0.5">
                      {backendText(dictionary, "document_date", "วันที่เอกสาร")}:{" "}
                      <span className="font-mono">{selectedDoc.docdatetime?.split("T")[0] || "-"}</span>
                    </p>
                  </div>
                </div>

                <div className="flex items-center gap-2">
                  <Button
                    variant="outline"
                    size="sm"
                    className="h-9 text-xs"
                    onClick={() => setShowEventsModal(true)}
                  >
                    <Clock className="mr-1.5 size-3.5" />
                    {backendText(dictionary, "ops_doc_events", "ประวัติเอกสาร")}
                  </Button>
                  <Button
                    variant="outline"
                    size="sm"
                    className="h-9 text-xs"
                    onClick={() => setShowPrintModal(true)}
                  >
                    <Printer className="mr-1.5 size-3.5" />
                    {backendText(dictionary, "print_document", "พิมพ์เอกสาร")}
                  </Button>
                  {screenActions.update ? (
                    <Button
                      variant="default"
                      size="sm"
                      className="h-9 text-xs bg-primary text-primary-foreground font-semibold"
                      disabled={!!selectedLockedHint}
                      title={selectedLockedHint || undefined}
                      aria-label={selectedLockedHint || backendText(dictionary, "edit_document", "แก้ไขเอกสาร")}
                      onClick={() => handleStartEdit(selectedDoc)}
                    >
                      <Pencil className="mr-1.5 size-3.5" />
                      {backendText(dictionary, "edit_document", "แก้ไขเอกสาร")}
                    </Button>
                  ) : null}
                </div>
              </div>

              {/* Header Info Cards */}
              <div className="grid grid-cols-1 md:grid-cols-3 gap-3 border border-border/60 rounded-xl p-4 bg-card">
                <div>
                  <span className="block text-xs text-muted-foreground">{cpLabel}</span>
                  <span className="font-semibold text-foreground text-sm">
                    {selectedDoc.custname || selectedDoc.custcode || "-"}
                  </span>
                </div>
                <div>
                  <span className="block text-xs text-muted-foreground">
                    {backendText(dictionary, "description", "รายละเอียด")}
                  </span>
                  <span className="text-foreground text-sm">{selectedDoc.description || "-"}</span>
                </div>
                <div>
                  <span className="block text-xs text-muted-foreground">
                    {backendText(dictionary, "net_total", "รวมทั้งสิ้น")}
                  </span>
                  <span className="font-bold text-primary font-mono text-base">
                    {formatAmount(selectedDoc.totalamount || "0")} ฿
                  </span>
                </div>
              </div>

              {/* Line Items Table */}
              {config.hasLineItems && (
                <div className="border border-border/60 rounded-xl p-4 bg-card flex flex-col gap-3">
                  <span className="font-semibold text-sm text-foreground">
                    {backendText(dictionary, "document_items_and_details", "รายการสินค้า / รายละเอียดเอกสาร")}
                  </span>

                  <div className="overflow-x-auto">
                    <table className="w-full text-left text-xs border-collapse">
                      <thead className="bg-muted/80 text-muted-foreground font-semibold border-b border-border">
                        <tr>
                          <th className="p-2 w-10 text-center">#</th>
                          <th className="p-2">{backendText(dictionary, "product_code", "รหัสสินค้า")}</th>
                          <th className="p-2">{backendText(dictionary, "product_name", "ชื่อสินค้า")}</th>
                          <th className="p-2 w-20 text-center">{backendText(dictionary, "unit", "หน่วยนับ")}</th>
                          <th className="p-2 w-24 text-right">{backendText(dictionary, "qty", "จำนวน")}</th>
                          <th className="p-2 w-28 text-right">{backendText(dictionary, "price", "ราคา")}</th>
                          <th className="p-2 w-28 text-right">{backendText(dictionary, "line_amount", "จำนวนเงิน")}</th>
                        </tr>
                      </thead>
                      <tbody className="divide-y divide-border/40">
                        {(selectedDoc.details || []).map((item, idx) => (
                          <tr key={idx} className="hover:bg-muted/20">
                            <td className="p-2 text-center text-muted-foreground">{idx + 1}</td>
                            <td className="p-2 font-mono font-medium text-primary">{item.itemcode || "-"}</td>
                            <td className="p-2 text-foreground font-medium">{item.itemname || "-"}</td>
                            <td className="p-2 text-center text-muted-foreground">{item.unitcode || "PCS"}</td>
                            <td className="p-2 text-right font-mono">{item.qty}</td>
                            <td className="p-2 text-right font-mono">
                              {formatAmount(item.price || "0")}
                            </td>
                            <td className="p-2 text-right font-mono font-semibold text-foreground">
                              {formatAmount(item.sumamount || "0")}
                            </td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>

                  {/* Summary Footer */}
                  <div className="flex flex-col items-end gap-1.5 border-t border-border/80 pt-3 text-xs">
                    <div className="flex justify-between w-64 text-muted-foreground">
                      <span>{backendText(dictionary, "goods_subtotal", "รวมมูลค่าสินค้า") + ":"}</span>
                      <span className="font-mono font-medium text-foreground">
                        {formatAmount(selectedDoc.totalbeforevat || "0")}
                      </span>
                    </div>
                    <div className="flex justify-between w-64 text-muted-foreground">
                      <span>{backendText(dictionary, "vat_7_percent", "ภาษีมูลค่าเพิ่ม (7%)") + ":"}</span>
                      <span className="font-mono font-medium text-foreground">
                        {formatAmount(selectedDoc.totalvatvalue || "0")}
                      </span>
                    </div>
                    <div className="flex justify-between w-64 text-sm font-bold border-t border-border pt-1 text-primary">
                      <span>{backendText(dictionary, "net_total", "รวมทั้งสิ้น") + ":"}</span>
                      <span className="font-mono">
                        {formatAmount(selectedDoc.totalamount || "0")} ฿
                      </span>
                    </div>
                  </div>
                </div>
              )}
            </div>
          ) : (
            <div className="flex-1 flex flex-col items-center justify-center text-center p-8 text-muted-foreground">
              <FileText className="size-12 mb-3 text-muted-foreground/40" />
              <p className="font-medium text-sm">
                {moduleUnavailable
                  ? backendText(
                      dictionary,
                      "module_not_available",
                      "จอนี้ยังไม่เปิดใช้งาน ระบบยังไม่รองรับเอกสารชนิดนี้ กรุณาติดต่อผู้ดูแลระบบ",
                    )
                  : backendText(dictionary, "select_document_hint", "เลือกรายการเอกสารจากตารางด้านซ้ายเพื่อดูรายละเอียด")}
              </p>
              {moduleUnavailable ? null : (
                <p className="text-xs text-muted-foreground/70 mt-1">
                  {backendText(dictionary, "create_document_hint", "หรือคลิก \"สร้างเอกสารใหม่\" เพื่อเริ่มบันทึก")}
                </p>
              )}
            </div>
          )}
        </section>
      </div>
    </main>
  );
}
