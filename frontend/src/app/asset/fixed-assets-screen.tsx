"use client";

import { useEffect, useState } from "react";
import { Button } from "@/components/ui/button";
import { useConfirmDialog } from "@/components/ui/confirm-dialog";
import { type LanguageCode } from "@/lib/i18n";
import { useBackendText } from "@/components/backend-text-provider";
import {
  type FixedAsset,
  type AssetType,
  type DepreciationScheduleItem,
  type AssetScheduleReport,
  getFixedAssets,
  getFixedAsset,
  getAssetSchedule,
  getAssetTypes,
  getFixedAssetScheduleReport,
  getTaxReconciliationReport,
  sendFixedAssetCommand,
  assetName,
} from "@/lib/fixed-assets";

/** แม่แบบอัตราค่าเสื่อมตามประมวลรัษฎากร ม.65 ทวิ (2) — ใช้เติมค่าเริ่มต้นในฟอร์มเท่านั้น ตัวเลขจริงคำนวณที่ backend */
const THAI_ASSET_CATEGORIES: { categoryCode: string; nameTh: string; standardUsefulLifeYears: number; standardDeprecPercent: number }[] = [
  { categoryCode: "BUILDING_PERM", nameTh: "อาคารถาวร", standardUsefulLifeYears: 20, standardDeprecPercent: 5.0 },
  { categoryCode: "BUILDING_TEMP", nameTh: "อาคารชั่วคราว", standardUsefulLifeYears: 1, standardDeprecPercent: 100.0 },
  { categoryCode: "VEHICLE_PASSENGER", nameTh: "ยานพาหนะ - รถยนต์นั่งไม่เกิน 10 ที่นั่ง (จำกัดภาษี 1 ลบ.)", standardUsefulLifeYears: 5, standardDeprecPercent: 20.0 },
  { categoryCode: "VEHICLE_COMMERCIAL", nameTh: "ยานพาหนะ - รถบรรทุก/เชิงพาณิชย์", standardUsefulLifeYears: 5, standardDeprecPercent: 20.0 },
  { categoryCode: "MACHINERY", nameTh: "เครื่องจักรและอุปกรณ์การผลิต", standardUsefulLifeYears: 5, standardDeprecPercent: 20.0 },
  { categoryCode: "OFFICE_EQUIPMENT", nameTh: "เครื่องใช้และอุปกรณ์สำนักงาน", standardUsefulLifeYears: 5, standardDeprecPercent: 20.0 },
  { categoryCode: "COMPUTER", nameTh: "คอมพิวเตอร์และอุปกรณ์อิเล็กทรอนิกส์ (3 ปี)", standardUsefulLifeYears: 3, standardDeprecPercent: 33.33 },
];

interface FixedAssetsScreenProps {
  route: string;
  embedded?: boolean;
  language?: LanguageCode;
}

export function FixedAssetsScreen({ route, embedded = false, language = "th" }: FixedAssetsScreenProps) {
  const tr = useBackendText();
  const [activeTab, setActiveTab] = useState<string>("registry");
  const [assets, setAssets] = useState<FixedAsset[]>([]);
  const [types, setTypes] = useState<AssetType[]>([]);
  const [loading, setLoading] = useState(false);
  const [searchQuery, setSearchQuery] = useState("");
  const [selectedAsset, setSelectedAsset] = useState<FixedAsset | null>(null);
  const [scheduleItems, setScheduleItems] = useState<DepreciationScheduleItem[]>([]);
  const [reportData, setReportData] = useState<AssetScheduleReport | null>(null);
  const [taxReportData, setTaxReportData] = useState<AssetScheduleReport | null>(null);
  const [postFiscalYear, setPostFiscalYear] = useState<string>("2026");
  const [postPeriod, setPostPeriod] = useState<number>(1);
  const [postStatusMsg, setPostStatusMsg] = useState<string>("");

  // Form State for New/Edit Asset
  const [isEditing, setIsEditing] = useState(false);
  const [editForm, setEditForm] = useState<Partial<FixedAsset>>({
    assetcode: "",
    names: [{ code: "th", name: "" }],
    assettypecode: "EQUIPMENT",
    cost: "0.00",
    scrapvalue: "1.00",
    usefullifeyears: 5,
    deprecpercent: "20.00",
    purchasedate: "2026-01-01",
    startcalcdate: "2026-01-01",
    firstyearpercent: "0.00",
    assetaccountcode: "120101",
    accumdeprecaccountcode: "129101",
    deprecexpenseaccountcode: "520103",
    status: "active",
  });

  // Disposal State
  const [isDisposing, setIsDisposing] = useState(false);
  const [disposingAsset, setDisposingAsset] = useState<FixedAsset | null>(null);
  const [disposalForm, setDisposalForm] = useState<{
    disposaldate: string;
    disposaltype: "sale" | "write_off" | "scrap";
    saleprice: string;
    vatamount: string;
    settlementaccountcode: string;
    gainlossaccountcode: string;
    reason: string;
  }>({
    disposaldate: new Date().toISOString().split("T")[0],
    disposaltype: "sale",
    saleprice: "0.00",
    vatamount: "0.00",
    settlementaccountcode: "110101",
    gainlossaccountcode: "420101",
    reason: "",
  });

  const { confirm, confirmationDialog } = useConfirmDialog({
    defaultConfirmLabel: tr("confirm", "ยืนยัน"),
    defaultCancelLabel: tr("cancel", "ยกเลิก"),
  });

  // Sync activeTab with initial route
  useEffect(() => {
    const cleanRoute = route.split("?")[0];
    if (cleanRoute === "/asset/registry") setActiveTab("registry");
    else if (cleanRoute === "/asset/depreciation") setActiveTab("depreciation");
    else if (cleanRoute === "/asset/post-gl") setActiveTab("post-gl");
    else if (cleanRoute === "/report/assetschedule") setActiveTab("schedule");
    else if (cleanRoute === "/asset/types") setActiveTab("types");
  }, [route]);

  const loadData = async () => {
    setLoading(true);
    try {
      const [resAssets, resTypes] = await Promise.all([
        getFixedAssets({ q: searchQuery }),
        getAssetTypes(),
      ]);
      if (resAssets?.items) setAssets(resAssets.items);
      if (resTypes?.items) setTypes(resTypes.items);
    } catch (err) {
      console.error("Failed to load FA data", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    loadData();
  }, [searchQuery]);

  const loadSchedule = async (assetCode: string) => {
    try {
      const res = await getAssetSchedule(assetCode);
      if (res?.items) setScheduleItems(res.items);
    } catch (err) {
      console.error("Failed to load schedule", err);
    }
  };

  const loadScheduleReport = async () => {
    setLoading(true);
    try {
      const res = await getFixedAssetScheduleReport(postFiscalYear, postPeriod);
      if (res?.report) setReportData(res.report);
    } catch (err) {
      console.error("Failed to load schedule report", err);
    } finally {
      setLoading(false);
    }
  };

  const loadTaxReport = async () => {
    setLoading(true);
    try {
      const res = await getTaxReconciliationReport(postFiscalYear);
      if (res?.report) setTaxReportData(res.report);
    } catch (err) {
      console.error("Failed to load tax report", err);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    if (activeTab === "schedule") loadScheduleReport();
    if (activeTab === "tax") loadTaxReport();
  }, [activeTab, postFiscalYear, postPeriod]);

  // Actions
  const handleSaveAsset = async () => {
    if (!editForm.assetcode?.trim()) {
      alert(tr("fa_please_specify_asset_code", "กรุณาระบุรหัสสินทรัพย์"));
      return;
    }
    const isNew = !editForm.id;
    const res = await sendFixedAssetCommand({
      resource: "assets",
      action: isNew ? "create" : "update",
      requestid: crypto.randomUUID(),
      id: editForm.id,
      version: editForm.version,
      asset: editForm,
    });

    if (res?.success) {
      alert(isNew ? tr("fa_asset_saved_successfully", "บันทึกสินทรัพย์เรียบร้อยแล้ว") : tr("fa_asset_updated_successfully", "แก้ไขสินทรัพย์เรียบร้อยแล้ว"));
      setIsEditing(false);
      loadData();
    } else {
      alert(res?.message || tr("error_saving", "เกิดข้อผิดพลาดในการบันทึก"));
    }
  };

  const handleDeleteAsset = async (ast: FixedAsset) => {
    const ok = await confirm({
      title: tr("fa_confirm_delete_asset", "ยืนยันการลบสินทรัพย์ {0}?").replace("{0}", ast.assetcode),
      description: tr("fa_unposted_gl_items_permanently_deleted", "รายการที่ยังไม่ได้ผ่านรายการเข้า GL จะถูกลบถาวร"),
      confirmLabel: tr("fa_confirm_delete", "ยืนยันลบ"),
    });
    if (!ok) return;

    const res = await sendFixedAssetCommand({
      resource: "assets",
      action: "delete",
      requestid: crypto.randomUUID(),
      id: ast.id,
      version: ast.version,
    });

    if (res?.success) {
      alert(tr("fa_asset_deleted_successfully", "ลบสินทรัพย์เรียบร้อยแล้ว"));
      loadData();
    } else {
      alert(res?.message || tr("error_deleting", "เกิดข้อผิดพลาดในการลบ"));
    }
  };

  const handlePostGL = async () => {
    const ok = await confirm({
      title: tr("fa_confirm_post_depreciation", "ยืนยันการผ่านรายการค่าเสื่อมราคาเข้า GL?"),
      description: tr("fa_post_period_summary", "ปีบัญชี {0} งวดที่ {1} (Dr. ค่าเสื่อมราคา / Cr. ค่าเสื่อมราคาสะสม)")
        .replace("{0}", postFiscalYear)
        .replace("{1}", String(postPeriod)),
      confirmLabel: tr("fa_post", "ผ่านรายการ (Post)"),
    });
    if (!ok) return;

    setLoading(true);
    setPostStatusMsg(tr("fa_posting_in_progress", "กำลังประมวลผลผ่านรายการ..."));
    const res = await sendFixedAssetCommand({
      resource: "depreciations",
      action: "post-gl",
      requestid: crypto.randomUUID(),
      fiscalyear: postFiscalYear,
      period: postPeriod,
    });
    setLoading(false);

    if (res?.success) {
      setPostStatusMsg(tr("fa_posted_journal_docno", "ผ่านรายการสำเร็จ! เลขที่ใบสำคัญสมุดรายวัน: {0}").replace("{0}", res?.journal?.docno || "JV"));
      alert(tr("fa_posted_voucher_docno", "ผ่านรายการสำเร็จ! เลขที่ใบสำคัญ: {0}").replace("{0}", String(res?.journal?.docno ?? "")));
      loadData();
    } else {
      setPostStatusMsg(tr("fa_failed_reason", "ล้มเหลว: {0}").replace("{0}", String(res?.message ?? "")));
      alert(res?.message || tr("fa_posting_error", "เกิดข้อผิดพลาดในการผ่านรายการ"));
    }
  };

  const handleConfirmDisposal = async () => {
    if (!disposingAsset) return;
    const ok = await confirm({
      title: tr("fa_confirm_disposal_title", "ยืนยันการจำหน่ายสินทรัพย์?"),
      description: tr(
        "fa_confirm_disposal_desc",
        "จำหน่ายสินทรัพย์ {0} ({1}) ระบบจะลงบัญชีกำไร/ขาดทุนจากการจำหน่าย และตัดยอดสินทรัพย์ออกจากบัญชี"
      )
        .replace("{0}", disposingAsset.assetcode)
        .replace("{1}", assetName(disposingAsset, language)),
      confirmLabel: tr("fa_confirm_disposal_btn", "ยืนยันจำหน่าย"),
      tone: "warning",
    });
    if (!ok) return;

    setLoading(true);
    try {
      const res = await sendFixedAssetCommand({
        resource: "disposals",
        action: "dispose",
        requestid: crypto.randomUUID(),
        disposal: {
          assetcode: disposingAsset.assetcode,
          disposaldate: disposalForm.disposaldate,
          disposaltype: disposalForm.disposaltype,
          saleprice: disposalForm.saleprice,
          vatamount: disposalForm.vatamount,
          settlementaccountcode: disposalForm.settlementaccountcode,
          gainlossaccountcode: disposalForm.gainlossaccountcode,
          reason: disposalForm.reason,
        },
      });
      if (res?.success) {
        setIsDisposing(false);
        setDisposingAsset(null);
        await loadData();
      } else {
        alert(res?.message || tr("fa_disposal_failed", "เกิดข้อผิดพลาดในการจำหน่ายสินทรัพย์"));
      }
    } catch (err) {
      console.error("Disposal failed", err);
      alert(tr("fa_disposal_failed", "เกิดข้อผิดพลาดในการจำหน่ายสินทรัพย์"));
    } finally {
      setLoading(false);
    }
  };

  const containerClass = embedded
    ? "p-3 h-full min-h-0 flex flex-col overflow-hidden text-[0.95rem]"
    : "mx-auto max-w-[1800px] p-4 min-h-[calc(100dvh-2rem)] flex flex-col text-[0.95rem]";

  return (
    <main className={containerClass} data-fa-tab={activeTab}>
      {/* Top Header Navigation Tabs */}
      <header className="shrink-0 flex flex-wrap items-center justify-between gap-3 border-b border-border/60 pb-3">
        <div className="flex items-center gap-3">
          <div className="h-10 w-10 rounded-xl bg-primary/10 flex items-center justify-center text-primary font-bold text-lg">
            FA
          </div>
          <div>
            <h1 className="text-xl font-bold text-foreground leading-tight">
              {tr("menu_fixed_assets_fa", "ระบบสินทรัพย์และค่าเสื่อมราคา")}
            </h1>
            <p className="text-xs text-muted-foreground">
              {tr("fa_fixed_assets_depreciation_management", "Fixed Assets & Depreciation Management (มาตรฐานสำนักงานบัญชีไทย)")}
            </p>
          </div>
        </div>

        {/* Tab Buttons */}
        <nav aria-label="Fixed Assets Navigation" className="flex flex-wrap items-center gap-2">
          <Button
            variant={activeTab === "registry" ? "default" : "outline"}
            className="h-10 text-sm font-medium"
            onClick={() => setActiveTab("registry")}
          >
            {tr("asset_registry", "รายละเอียดสินทรัพย์")}
          </Button>
          <Button
            variant={activeTab === "depreciation" ? "default" : "outline"}
            className="h-10 text-sm font-medium"
            onClick={() => setActiveTab("depreciation")}
          >
            {tr("asset_depreciation", "ประมวลผลสินทรัพย์")}
          </Button>
          <Button
            variant={activeTab === "post-gl" ? "default" : "outline"}
            className="h-10 text-sm font-medium"
            onClick={() => setActiveTab("post-gl")}
          >
            {tr("fa_post_depreciation_to_gl", "โอนค่าเสื่อมเข้าบัญชีแยกประเภท")}
          </Button>
          <Button
            variant={activeTab === "schedule" ? "default" : "outline"}
            className="h-10 text-sm font-medium"
            onClick={() => setActiveTab("schedule")}
          >
            {tr("fixed_asset_schedule", "ตารางค่าเสื่อมและสินทรัพย์")}
          </Button>
          <Button
            variant={activeTab === "tax" ? "default" : "outline"}
            className="h-10 text-sm font-medium"
            onClick={() => setActiveTab("tax")}
          >
            {tr("fa_tax_reconciliation_pnd_50", "กระทบยอดภาษี (ภ.ง.ด.50)")}
          </Button>
        </nav>
      </header>

      {/* Tab Content Area */}
      <div className="flex-1 min-h-0 pt-3 flex flex-col">
        {/* 1. ASSET REGISTRY TAB */}
        {activeTab === "registry" && (
          <div className="flex-1 min-h-0 flex flex-col gap-3">
            {/* Action Bar */}
            <div className="flex flex-wrap items-center justify-between gap-3 bg-card/60 p-3 rounded-xl border border-border/40">
              <div className="flex items-center gap-2 flex-1 min-w-[240px] max-w-md">
                <input
                  type="text"
                  placeholder={tr("fa_search_code_name_serial", "ค้นหารหัส ชื่อ หรือ Serial...")}
                  value={searchQuery}
                  onChange={(e) => setSearchQuery(e.target.value)}
                  className="w-full h-10 px-3 rounded-lg border border-input bg-background text-sm focus:outline-none focus:ring-2 focus:ring-primary/20"
                />
              </div>

              <div className="flex items-center gap-2">
                <Button
                  onClick={() => {
                    setEditForm({
                      assetcode: `FA-${new Date().getFullYear()}-${String(assets.length + 1).padStart(3, "0")}`,
                      names: [{ code: "th", name: "" }],
                      assettypecode: "EQUIPMENT",
                      cost: "10000.00",
                      scrapvalue: "1.00",
                      usefullifeyears: 5,
                      deprecpercent: "20.00",
                      purchasedate: new Date().toISOString().split("T")[0],
                      startcalcdate: new Date().toISOString().split("T")[0],
                      firstyearpercent: "0.00",
                      assetaccountcode: "120101",
                      accumdeprecaccountcode: "129101",
                      deprecexpenseaccountcode: "520103",
                      status: "active",
                    });
                    setIsEditing(true);
                  }}
                  className="h-10 bg-primary text-primary-foreground text-sm font-medium"
                >
                  + {tr("fa_add_new_asset", "เพิ่มสินทรัพย์ใหม่")}
                </Button>
              </div>
            </div>

            {/* Asset Table */}
            <div className="flex-1 min-h-0 overflow-auto border border-border/60 rounded-xl bg-card">
              <table className="w-full text-left text-sm border-collapse">
                <thead className="sticky top-0 bg-muted/90 backdrop-blur z-10 border-b border-border text-xs uppercase tracking-wider font-semibold">
                  <tr>
                    <th className="p-3">{tr("fa_asset_code", "รหัสสินทรัพย์")}</th>
                    <th className="p-3">{tr("fa_asset_name", "ชื่อสินทรัพย์")}</th>
                    <th className="p-3">{tr("cart_type_label", "ประเภท")}</th>
                    <th className="p-3 text-right">{tr("fa_cost", "ราคาทุน")}</th>
                    <th className="p-3 text-right">{tr("fa_useful_life_years", "อายุ (ปี)")}</th>
                    <th className="p-3 text-right">{tr("fa_rate", "อัตรา (%)")}</th>
                    <th className="p-3">{tr("fa_start_date", "วันที่เริ่มคิด")}</th>
                    <th className="p-3 text-center">{tr("status", "สถานะ")}</th>
                    <th className="p-3 text-center">{tr("barcode_actions", "การจัดการ")}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/40">
                  {assets.map((ast) => (
                    <tr key={ast.assetcode} className="hover:bg-muted/40 transition-colors">
                      <td className="p-3 font-semibold text-primary">{ast.assetcode}</td>
                      <td className="p-3 font-medium">{assetName(ast, language)}</td>
                      <td className="p-3 text-muted-foreground">{ast.assettypecode}</td>
                      <td className="p-3 text-right font-mono font-medium">
                        {Number(ast.cost).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                      </td>
                      <td className="p-3 text-right font-mono">{ast.usefullifeyears}</td>
                      <td className="p-3 text-right font-mono">{ast.deprecpercent}%</td>
                      <td className="p-3 font-mono text-xs">{ast.startcalcdate}</td>
                      <td className="p-3 text-center">
                        <span
                          className={`inline-block px-2 py-0.5 rounded-full text-xs font-medium ${
                            ast.status === "active"
                              ? "bg-emerald-500/15 text-emerald-600 dark:text-emerald-400"
                              : "bg-rose-500/15 text-rose-600 dark:text-rose-400"
                          }`}
                        >
                          {ast.status === "active" ? tr("fa_normal_use", "ใช้งานปกติ") : tr("fa_disposed", "จำหน่ายแล้ว")}
                        </span>
                      </td>
                      <td className="p-3 text-center">
                        <div className="flex items-center justify-center gap-1">
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8 px-2 text-xs"
                            onClick={() => {
                              setSelectedAsset(ast);
                              loadSchedule(ast.assetcode);
                              setActiveTab("depreciation");
                            }}
                          >
                            {tr("fa_schedule", "ตารางงวด")}
                          </Button>
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8 px-2 text-xs"
                            onClick={() => {
                              setEditForm(ast);
                              setIsEditing(true);
                            }}
                          >
                            {tr("edit", "แก้ไข")}
                          </Button>
                          {ast.status === "active" && (
                            <Button
                              variant="ghost"
                              size="sm"
                              className="h-8 px-2 text-xs text-amber-600 hover:text-amber-700 hover:bg-amber-500/10 dark:text-amber-400"
                              onClick={() => {
                                setDisposingAsset(ast);
                                setDisposalForm({
                                  disposaldate: new Date().toISOString().split("T")[0],
                                  disposaltype: "sale",
                                  saleprice: "0.00",
                                  vatamount: "0.00",
                                  settlementaccountcode: "110101",
                                  gainlossaccountcode: "420101",
                                  reason: "",
                                });
                                setIsDisposing(true);
                              }}
                            >
                              {tr("fa_dispose", "จำหน่าย")}
                            </Button>
                          )}
                          <Button
                            variant="ghost"
                            size="sm"
                            className="h-8 px-2 text-xs text-destructive hover:text-destructive"
                            onClick={() => handleDeleteAsset(ast)}
                          >
                            {tr("gl_delete", "ลบ")}
                          </Button>
                        </div>
                      </td>
                    </tr>
                  ))}
                  {assets.length === 0 && !loading && (
                    <tr>
                      <td colSpan={9} className="p-8 text-center text-muted-foreground">
                        {tr("fa_no_fixed_asset_data_found", "ไม่พบข้อมูลสินทรัพย์ถาวร")}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>

            {/* Modal Dialog for Edit / Create */}
            {isEditing && (
              <div className="fixed inset-0 bg-background/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
                <div className="bg-card border border-border rounded-2xl max-w-2xl w-full p-6 shadow-2xl flex flex-col gap-4">
                  <h2 className="text-lg font-bold text-foreground border-b border-border/60 pb-3">
                    {editForm.id ? tr("fa_edit_asset_information", "แก้ไขข้อมูลสินทรัพย์") : tr("fa_add_new_fixed_asset", "เพิ่มสินทรัพย์ถาวรใหม่")}
                  </h2>
                  <div className="flex flex-wrap items-center gap-1.5 py-2">
                    <span className="text-xs font-semibold text-muted-foreground mr-1">
                      {tr("fa_preset_label", "แม่แบบภาษี ม.65 ทวิ (2):")}
                    </span>
                    {THAI_ASSET_CATEGORIES.map((cat) => (
                      <button
                        key={cat.categoryCode}
                        type="button"
                        onClick={() => {
                          setEditForm((prev) => ({
                            ...prev,
                            assettypecode: cat.categoryCode,
                            usefullifeyears: cat.standardUsefulLifeYears,
                            deprecpercent: cat.standardDeprecPercent.toFixed(2),
                          }));
                        }}
                        className={`px-2.5 py-1 text-xs rounded-full border transition-all ${
                          editForm.assettypecode === cat.categoryCode
                            ? "bg-primary text-primary-foreground border-primary font-semibold shadow-xs"
                            : "bg-muted/50 hover:bg-muted text-foreground border-border/60"
                        }`}
                      >
                        {cat.nameTh} ({cat.standardDeprecPercent}%)
                      </button>
                    ))}
                  </div>
                  <div className="grid grid-cols-2 gap-4 text-sm max-h-[60vh] overflow-y-auto pr-1">
                    <div>
                      <label className="block text-xs font-semibold text-muted-foreground mb-1">{tr("fa_asset_code", "รหัสสินทรัพย์")}</label>
                      <input
                        type="text"
                        value={editForm.assetcode || ""}
                        disabled={Boolean(editForm.id)}
                        onChange={(e) => setEditForm({ ...editForm, assetcode: e.target.value })}
                        className="w-full h-10 px-3 rounded-lg border border-input bg-background disabled:opacity-60"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-muted-foreground mb-1">{tr("fa_asset_name_thai", "ชื่อสินทรัพย์ (ไทย)")}</label>
                      <input
                        type="text"
                        value={editForm.names?.[0]?.name || ""}
                        onChange={(e) => setEditForm({ ...editForm, names: [{ code: "th", name: e.target.value }] })}
                        className="w-full h-10 px-3 rounded-lg border border-input bg-background"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-muted-foreground mb-1">{tr("fa_asset_type", "ประเภทสินทรัพย์")}</label>
                      <input
                        type="text"
                        value={editForm.assettypecode || "EQUIPMENT"}
                        onChange={(e) => setEditForm({ ...editForm, assettypecode: e.target.value })}
                        className="w-full h-10 px-3 rounded-lg border border-input bg-background"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-muted-foreground mb-1">{tr("fa_capital_cost", "ราคาทุน (Cost)")}</label>
                      <input
                        type="number"
                        step="0.01"
                        value={editForm.cost || "0.00"}
                        onChange={(e) => setEditForm({ ...editForm, cost: e.target.value })}
                        className="w-full h-10 px-3 rounded-lg border border-input bg-background font-mono"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-muted-foreground mb-1">{tr("fa_salvage_value", "ราคาซาก (Scrap Value)")}</label>
                      <input
                        type="number"
                        step="0.01"
                        value={editForm.scrapvalue || "1.00"}
                        onChange={(e) => setEditForm({ ...editForm, scrapvalue: e.target.value })}
                        className="w-full h-10 px-3 rounded-lg border border-input bg-background font-mono"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-muted-foreground mb-1">{tr("fa_useful_life", "อายุการใช้งาน (ปี)")}</label>
                      <input
                        type="number"
                        value={editForm.usefullifeyears || 5}
                        onChange={(e) => {
                          const years = Number(e.target.value);
                          setEditForm({
                            ...editForm,
                            usefullifeyears: years,
                            deprecpercent: years > 0 ? (100 / years).toFixed(2) : "0.00",
                          });
                        }}
                        className="w-full h-10 px-3 rounded-lg border border-input bg-background font-mono"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-muted-foreground mb-1">{tr("fa_depreciation_rate", "อัตราค่าเสื่อม (%)")}</label>
                      <input
                        type="number"
                        step="0.01"
                        value={editForm.deprecpercent || "20.00"}
                        onChange={(e) => setEditForm({ ...editForm, deprecpercent: e.target.value })}
                        className="w-full h-10 px-3 rounded-lg border border-input bg-background font-mono"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-muted-foreground mb-1">{tr("fa_purchase_date", "วันที่ซื้อ")}</label>
                      <input
                        type="date"
                        value={editForm.purchasedate || ""}
                        onChange={(e) => setEditForm({ ...editForm, purchasedate: e.target.value, startcalcdate: e.target.value })}
                        className="w-full h-10 px-3 rounded-lg border border-input bg-background font-mono"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-muted-foreground mb-1">{tr("fa_first_year_special_allowance", "สิทธิพิเศษปีแรก (%) เช่น 40% คอมพิวเตอร์")}</label>
                      <input
                        type="number"
                        step="0.01"
                        value={editForm.firstyearpercent || "0.00"}
                        onChange={(e) => setEditForm({ ...editForm, firstyearpercent: e.target.value })}
                        className="w-full h-10 px-3 rounded-lg border border-input bg-background font-mono"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-muted-foreground mb-1">{tr("fa_asset_account_code", "รหัสบัญชีสินทรัพย์ (GL)")}</label>
                      <input
                        type="text"
                        value={editForm.assetaccountcode || "120101"}
                        onChange={(e) => setEditForm({ ...editForm, assetaccountcode: e.target.value })}
                        className="w-full h-10 px-3 rounded-lg border border-input bg-background font-mono"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-muted-foreground mb-1">{tr("fa_accumulated_depreciation_account_code", "รหัสบัญชีค่าเสื่อมสะสม (GL)")}</label>
                      <input
                        type="text"
                        value={editForm.accumdeprecaccountcode || "129101"}
                        onChange={(e) => setEditForm({ ...editForm, accumdeprecaccountcode: e.target.value })}
                        className="w-full h-10 px-3 rounded-lg border border-input bg-background font-mono"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-muted-foreground mb-1">{tr("fa_depreciation_expense_account_code", "รหัสบัญชีค่าใช้จ่ายค่าเสื่อม (GL)")}</label>
                      <input
                        type="text"
                        value={editForm.deprecexpenseaccountcode || "520103"}
                        onChange={(e) => setEditForm({ ...editForm, deprecexpenseaccountcode: e.target.value })}
                        className="w-full h-10 px-3 rounded-lg border border-input bg-background font-mono"
                      />
                    </div>
                  </div>

                  <div className="flex items-center justify-end gap-3 pt-4 border-t border-border/60">
                    <Button variant="outline" onClick={() => setIsEditing(false)}>
                      {tr("cancel", "ยกเลิก")}
                    </Button>
                    <Button onClick={handleSaveAsset}>{tr("gl_save_data", "บันทึกข้อมูล")}</Button>
                  </div>
                </div>
              </div>
            )}

            {/* Modal Dialog for Disposal */}
            {isDisposing && disposingAsset && (
              <div className="fixed inset-0 bg-background/80 backdrop-blur-sm z-50 flex items-center justify-center p-4">
                <div className="bg-card border border-border rounded-2xl max-w-lg w-full p-6 shadow-2xl flex flex-col gap-4">
                  <h2 className="text-lg font-bold text-foreground border-b border-border/60 pb-3">
                    {tr("fa_asset_disposal_title", "บันทึกการจำหน่ายสินทรัพย์")} — {disposingAsset.assetcode}
                  </h2>
                  <div className="text-xs text-muted-foreground bg-muted/40 p-3 rounded-lg flex flex-col gap-1">
                    <div><strong>{tr("fa_asset_name", "ชื่อสินทรัพย์")}:</strong> {assetName(disposingAsset, language)}</div>
                    <div><strong>{tr("fa_cost", "ราคาทุน")}:</strong> {Number(disposingAsset.cost).toLocaleString("th-TH", { minimumFractionDigits: 2 })} บาท</div>
                  </div>
                  <div className="grid grid-cols-2 gap-3 text-sm max-h-[60vh] overflow-y-auto pr-1">
                    <div>
                      <label className="block text-xs font-semibold text-muted-foreground mb-1">
                        {tr("fa_disposal_date", "วันที่จำหน่าย")}
                      </label>
                      <input
                        type="date"
                        value={disposalForm.disposaldate}
                        onChange={(e) => setDisposalForm({ ...disposalForm, disposaldate: e.target.value })}
                        className="w-full h-10 px-3 rounded-lg border border-input bg-background"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-muted-foreground mb-1">
                        {tr("fa_disposal_type", "ประเภทการจำหน่าย")}
                      </label>
                      <select
                        value={disposalForm.disposaltype}
                        onChange={(e) => setDisposalForm({ ...disposalForm, disposaltype: e.target.value as "sale" | "write_off" | "scrap" })}
                        className="w-full h-10 px-3 rounded-lg border border-input bg-background"
                      >
                        <option value="sale">{tr("fa_disposal_type_sale", "ขายสินทรัพย์")}</option>
                        <option value="write_off">{tr("fa_disposal_type_write_off", "ตัดจำหน่าย / สูญหาย")}</option>
                        <option value="scrap">{tr("fa_disposal_type_scrap", "เศษซาก")}</option>
                      </select>
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-muted-foreground mb-1">
                        {tr("fa_sale_price", "ราคาขาย (ไม่รวม VAT)")}
                      </label>
                      <input
                        type="number"
                        step="0.01"
                        value={disposalForm.saleprice}
                        onChange={(e) => {
                          const price = Number(e.target.value) || 0;
                          const vat = Math.round(price * 0.07 * 100) / 100;
                          setDisposalForm({ ...disposalForm, saleprice: e.target.value, vatamount: String(vat) });
                        }}
                        className="w-full h-10 px-3 rounded-lg border border-input bg-background font-mono"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-muted-foreground mb-1">
                        {tr("fa_vat_amount", "ภาษีมูลค่าเพิ่ม (VAT 7%)")}
                      </label>
                      <input
                        type="number"
                        step="0.01"
                        value={disposalForm.vatamount}
                        onChange={(e) => setDisposalForm({ ...disposalForm, vatamount: e.target.value })}
                        className="w-full h-10 px-3 rounded-lg border border-input bg-background font-mono"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-muted-foreground mb-1">
                        {tr("fa_settlement_account", "บัญชีเงินสด/เงินฝาก/ลูกหนี้")}
                      </label>
                      <input
                        type="text"
                        value={disposalForm.settlementaccountcode}
                        onChange={(e) => setDisposalForm({ ...disposalForm, settlementaccountcode: e.target.value })}
                        placeholder="110101"
                        className="w-full h-10 px-3 rounded-lg border border-input bg-background font-mono"
                      />
                    </div>
                    <div>
                      <label className="block text-xs font-semibold text-muted-foreground mb-1">
                        {tr("fa_gain_loss_account", "บัญชีกำไร/ขาดทุนจากการขาย")}
                      </label>
                      <input
                        type="text"
                        value={disposalForm.gainlossaccountcode}
                        onChange={(e) => setDisposalForm({ ...disposalForm, gainlossaccountcode: e.target.value })}
                        placeholder="420101"
                        className="w-full h-10 px-3 rounded-lg border border-input bg-background font-mono"
                      />
                    </div>
                    <div className="col-span-2">
                      <label className="block text-xs font-semibold text-muted-foreground mb-1">
                        {tr("fa_disposal_reason", "เหตุผล / รายละเอียด")}
                      </label>
                      <input
                        type="text"
                        value={disposalForm.reason}
                        onChange={(e) => setDisposalForm({ ...disposalForm, reason: e.target.value })}
                        placeholder={tr("fa_disposal_reason_placeholder", "ระบุเหตุผลในการจำหน่ายสินทรัพย์")}
                        className="w-full h-10 px-3 rounded-lg border border-input bg-background"
                      />
                    </div>
                  </div>
                  <div className="flex items-center justify-end gap-2 border-t border-border/60 pt-4">
                    <Button
                      variant="outline"
                      onClick={() => {
                        setIsDisposing(false);
                        setDisposingAsset(null);
                      }}
                    >
                      {tr("cancel", "ยกเลิก")}
                    </Button>
                    <Button
                      className="bg-amber-600 hover:bg-amber-700 text-white font-medium"
                      onClick={handleConfirmDisposal}
                      disabled={loading}
                    >
                      {loading ? tr("processing", "กำลังประมวลผล...") : tr("fa_confirm_disposal_btn", "ยืนยันจำหน่าย")}
                    </Button>
                  </div>
                </div>
              </div>
            )}
          </div>
        )}

        {/* 2. DEPRECIATION SCHEDULE PREVIEW TAB */}
        {activeTab === "depreciation" && (
          <div className="flex-1 min-h-0 flex flex-col gap-3">
            <div className="flex items-center justify-between gap-3 bg-card p-3 rounded-xl border border-border/40">
              <div className="flex items-center gap-3">
                <span className="text-sm font-semibold">{tr("fa_select_asset", "เลือกสินทรัพย์:")}</span>
                <select
                  value={selectedAsset?.assetcode || ""}
                  onChange={(e) => {
                    const found = assets.find((x) => x.assetcode === e.target.value);
                    setSelectedAsset(found || null);
                    if (found) loadSchedule(found.assetcode);
                  }}
                  className="h-10 px-3 rounded-lg border border-input bg-background text-sm font-medium"
                >
                  <option value="">{tr("fa_please_select_asset", "-- กรุณาเลือกสินทรัพย์ --")}</option>
                  {assets.map((a) => (
                    <option key={a.assetcode} value={a.assetcode}>
                      {a.assetcode} : {assetName(a, language)}
                    </option>
                  ))}
                </select>
              </div>

              {selectedAsset && (
                <div className="text-xs text-muted-foreground font-mono">
                  ราคาทุน: {Number(selectedAsset.cost).toLocaleString()} | ซาก: {Number(selectedAsset.scrapvalue).toLocaleString()} | อัตรา: {selectedAsset.deprecpercent}%
                </div>
              )}
            </div>

            <div className="flex-1 min-h-0 overflow-auto border border-border/60 rounded-xl bg-card">
              <table className="w-full text-left text-sm border-collapse">
                <thead className="sticky top-0 bg-muted/90 backdrop-blur z-10 border-b border-border text-xs uppercase tracking-wider font-semibold">
                  <tr>
                    <th className="p-3">{tr("gl_fiscal_year", "ปีบัญชี")}</th>
                    <th className="p-3 text-center">{tr("fa_period_column", "งวด")}</th>
                    <th className="p-3">{tr("report_condition_start_date", "วันที่เริ่มต้น")}</th>
                    <th className="p-3">{tr("end_date", "วันที่สิ้นสุด")}</th>
                    <th className="p-3 text-right">{tr("fa_number_of_days", "จำนวนวัน")}</th>
                    <th className="p-3 text-right">{tr("fa_depreciation_this_period", "ค่าเสื่อมราคางวดนี้")}</th>
                    <th className="p-3 text-right">{tr("fa_accumulated_depreciation", "ค่าเสื่อมราคาสะสม")}</th>
                    <th className="p-3 text-right">{tr("fa_net_book_value", "มูลค่าคงเหลือสุทธิ (NBV)")}</th>
                    <th className="p-3 text-center">{tr("fa_gl_status", "สถานะ GL")}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/40">
                  {scheduleItems.map((item, idx) => (
                    <tr key={idx} className="hover:bg-muted/40 transition-colors">
                      <td className="p-3 font-semibold">{item.fiscalyear}</td>
                      <td className="p-3 text-center font-mono">{item.period}</td>
                      <td className="p-3 font-mono text-xs">{item.startdate}</td>
                      <td className="p-3 font-mono text-xs">{item.stopdate}</td>
                      <td className="p-3 text-right font-mono">{item.days}</td>
                      <td className="p-3 text-right font-mono font-medium text-amber-600 dark:text-amber-400">
                        {Number(item.perioddeprec).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                      </td>
                      <td className="p-3 text-right font-mono">
                        {Number(item.accumdeprec).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                      </td>
                      <td className="p-3 text-right font-mono font-semibold text-emerald-600 dark:text-emerald-400">
                        {Number(item.netbookvalue).toLocaleString("th-TH", { minimumFractionDigits: 2 })}
                      </td>
                      <td className="p-3 text-center">
                        <span
                          className={`inline-block px-2 py-0.5 rounded-full text-xs font-medium ${
                            item.isposted
                              ? "bg-emerald-500/15 text-emerald-600 dark:text-emerald-400"
                              : "bg-muted text-muted-foreground"
                          }`}
                        >
                          {item.isposted ? tr("fa_posted_with_docno", "ผ่านแล้ว ({0})").replace("{0}", String(item.journaldocno ?? "")) : tr("fa_not_posted", "ยังไม่ผ่าน")}
                        </span>
                      </td>
                    </tr>
                  ))}
                  {scheduleItems.length === 0 && (
                    <tr>
                      <td colSpan={9} className="p-8 text-center text-muted-foreground">
                        {tr("fa_select_asset_view_depreciation_table", "กรุณาเลือกสินทรัพย์เพื่อดูตารางคำนวณค่าเสื่อมราคา")}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* 3. POST TO GL TAB */}
        {activeTab === "post-gl" && (
          <div className="flex-1 min-h-0 flex flex-col max-w-xl mx-auto w-full gap-4 pt-6">
            <div className="bg-card border border-border/60 rounded-2xl p-6 shadow-md flex flex-col gap-4">
              <h2 className="text-lg font-bold text-foreground border-b border-border/40 pb-2">
                {tr("fa_transfer_depreciation_to_gl", "โอนค่าเสื่อมราคาเข้าบัญชีแยกประเภท (GL Journal Posting)")}
              </h2>
              <p className="text-xs text-muted-foreground leading-relaxed">
                {tr("fa_depreciation_journal_voucher", "ระบบจะรวบรวมค่าเสื่อมราคาของสินทรัพย์ทุกตัวในงวดที่เลือก นำมาบันทึกสมุดรายวันทั่วไป (JV)")}
                {tr("fa_depreciation_account_entry", "โดยเดบิตบัญชีค่าใช้จ่ายค่าเสื่อมราคา และเครดิตบัญชีค่าเสื่อมราคาสะสม")}
              </p>

              <div className="grid grid-cols-2 gap-4 pt-2">
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">{tr("gl_fiscal_year", "ปีบัญชี")}</label>
                  <input
                    type="text"
                    value={postFiscalYear}
                    onChange={(e) => setPostFiscalYear(e.target.value)}
                    className="w-full h-10 px-3 rounded-lg border border-input bg-background font-mono text-sm"
                  />
                </div>
                <div>
                  <label className="block text-xs font-semibold text-muted-foreground mb-1">{tr("fa_period", "งวดที่ (1-12)")}</label>
                  <input
                    type="number"
                    min={1}
                    max={12}
                    value={postPeriod}
                    onChange={(e) => setPostPeriod(Number(e.target.value))}
                    className="w-full h-10 px-3 rounded-lg border border-input bg-background font-mono text-sm"
                  />
                </div>
              </div>

              {postStatusMsg && (
                <div className="p-3 rounded-xl bg-muted/60 text-xs font-mono border border-border/40">
                  {postStatusMsg}
                </div>
              )}

              <Button
                onClick={handlePostGL}
                disabled={loading}
                className="w-full h-11 bg-primary text-primary-foreground font-semibold text-sm mt-2"
              >
                {loading ? tr("ops_processing", "กำลังประมวลผล...") : tr("fa_post_period_to_gl", "ผ่านรายการประจำงวด {0}/{1} เข้า GL")
                    .replace("{0}", String(postPeriod))
                    .replace("{1}", postFiscalYear)}
              </Button>
            </div>
          </div>
        )}

        {/* 4. ASSET SCHEDULE REPORT TAB */}
        {activeTab === "schedule" && (
          <div className="flex-1 min-h-0 flex flex-col gap-3">
            <div className="flex items-center justify-between gap-3 bg-card p-3 rounded-xl border border-border/40">
              <div className="flex items-center gap-3">
                <span className="text-sm font-semibold">{tr("fa_accounting_year", "ปีบัญชี:")}</span>
                <input
                  type="text"
                  value={postFiscalYear}
                  onChange={(e) => setPostFiscalYear(e.target.value)}
                  className="h-10 w-24 px-3 rounded-lg border border-input bg-background font-mono text-sm"
                />
                <Button size="sm" onClick={loadScheduleReport} className="h-10">
                  {tr("fa_pull_report", "ดึงรายงาน")}
                </Button>
              </div>
              <div className="text-xs text-muted-foreground">
                {tr("fa_fixed_asset_depreciation_schedule_report", "รายงานตารางค่าเสื่อมราคาและสินทรัพย์ถาวร (Fixed Asset Schedule)")}
              </div>
            </div>

            <div className="flex-1 min-h-0 overflow-auto border border-border/60 rounded-xl bg-card">
              <table className="w-full text-left text-sm border-collapse">
                <thead className="sticky top-0 bg-muted/90 backdrop-blur z-10 border-b border-border text-xs uppercase tracking-wider font-semibold">
                  <tr>
                    {reportData?.columns.map((col) => (
                      <th key={col.key} className={`p-3 ${col.amount ? "text-right" : ""}`}>
                        {col.label}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/40">
                  {reportData?.rows.map((r, i) => (
                    <tr key={i} className="hover:bg-muted/40 transition-colors">
                      {reportData.columns.map((col) => (
                        <td key={col.key} className={`p-3 font-mono text-xs ${col.amount ? "text-right" : ""}`}>
                          {r[col.key]}
                        </td>
                      ))}
                    </tr>
                  ))}
                  {reportData?.totals && (
                    <tr className="bg-muted/60 font-bold border-t-2 border-border text-xs font-mono">
                      {reportData.columns.map((col) => (
                        <td key={col.key} className={`p-3 ${col.amount ? "text-right" : ""}`}>
                          {reportData.totals[col.key] || ""}
                        </td>
                      ))}
                    </tr>
                  )}
                  {(!reportData || reportData.rows.length === 0) && !loading && (
                    <tr>
                      <td colSpan={13} className="p-8 text-center text-muted-foreground">
                        {tr("fa_no_report_data_this_year", "ไม่พบข้อมูลรายงานสำหรับปีนี้")}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* 5. TAX RECONCILIATION TAB (PND 50) */}
        {activeTab === "tax" && (
          <div className="flex-1 min-h-0 flex flex-col gap-3">
            <div className="flex items-center justify-between gap-3 bg-card p-3 rounded-xl border border-border/40">
              <div className="flex items-center gap-3">
                <span className="text-sm font-semibold">{tr("fa_tax_year", "ปีภาษี:")}</span>
                <input
                  type="text"
                  value={postFiscalYear}
                  onChange={(e) => setPostFiscalYear(e.target.value)}
                  className="h-10 w-24 px-3 rounded-lg border border-input bg-background font-mono text-sm"
                />
                <Button size="sm" onClick={loadTaxReport} className="h-10">
                  {tr("fa_pull_report", "ดึงรายงาน")}
                </Button>
              </div>
              <div className="text-xs text-muted-foreground">
                {tr("fa_accounting_tax_depreciation_reconciliation_report", "รายงานกระทบยอดค่าเสื่อมราคาทางบัญชี vs ทางภาษีอากร สำหรับแบบ ภ.ง.ด.50")}
              </div>
            </div>

            <div className="flex-1 min-h-0 overflow-auto border border-border/60 rounded-xl bg-card">
              <table className="w-full text-left text-sm border-collapse">
                <thead className="sticky top-0 bg-muted/90 backdrop-blur z-10 border-b border-border text-xs uppercase tracking-wider font-semibold">
                  <tr>
                    {taxReportData?.columns.map((col) => (
                      <th key={col.key} className={`p-3 ${col.amount ? "text-right" : ""}`}>
                        {col.label}
                      </th>
                    ))}
                  </tr>
                </thead>
                <tbody className="divide-y divide-border/40">
                  {taxReportData?.rows.map((r, i) => (
                    <tr key={i} className="hover:bg-muted/40 transition-colors">
                      {taxReportData.columns.map((col) => (
                        <td key={col.key} className={`p-3 font-mono text-xs ${col.amount ? "text-right" : ""}`}>
                          {r[col.key]}
                        </td>
                      ))}
                    </tr>
                  ))}
                  {taxReportData?.totals && (
                    <tr className="bg-muted/60 font-bold border-t-2 border-border text-xs font-mono">
                      {taxReportData.columns.map((col) => (
                        <td key={col.key} className={`p-3 ${col.amount ? "text-right" : ""}`}>
                          {taxReportData.totals[col.key] || ""}
                        </td>
                      ))}
                    </tr>
                  )}
                  {(!taxReportData || taxReportData.rows.length === 0) && !loading && (
                    <tr>
                      <td colSpan={7} className="p-8 text-center text-muted-foreground">
                        {tr("fa_no_data_for_tax_year", "ไม่พบข้อมูลสำหรับปีภาษีนี้")}
                      </td>
                    </tr>
                  )}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>

      {confirmationDialog}
    </main>
  );
}
