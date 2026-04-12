import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:intl/intl.dart';
import 'package:flutter_map/flutter_map.dart';
import 'package:latlong2/latlong.dart';
import 'package:geolocator/geolocator.dart';
import 'package:http/http.dart' as http;
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/model/job_project_model.dart';
import 'package:smlaicloud/model/cost_center_model.dart';
import 'package:smlaicloud/repositories/job_project_repository.dart';
import 'package:smlaicloud/repositories/cost_center_repository.dart';
import 'package:smlaicloud/model/currency_model.dart';
import 'package:smlaicloud/utils/date_picker.dart';
import 'package:smlaicloud/utils/time_picker.dart';
import 'package:smlaicloud/screens/purchaserequisition/utils/pr_form_controller.dart';
import 'package:smlaicloud/global.dart' as global;

/// PR Header Widget — ส่วนหัวเอกสารใบขอซื้อ
///
/// ฟีเจอร์ครบ:
/// - เลขที่เอกสาร + วันที่ + เวลา
/// - ผู้ขอซื้อ + แผนก
/// - ความเร่งด่วน + วันที่ต้องการรับ
/// - Supplier (optional)
/// - เงื่อนไขชำระเงิน (credit days)
/// - เอกสารอ้างอิง
/// - Job/โครงการ
/// - ที่อยู่จัดส่ง
/// - วัตถุประสงค์/เหตุผล
class PRHeaderWidget extends StatefulWidget {
  final TransactionModel screenData;
  final PRFormController formController;
  final Function(void Function()) setState;
  final BuildContext parentContext;
  final bool docDateTimeValidated;
  final Function({required String word}) searchSupplier;
  final Function({required String word})? searchDepartment;
  final Function({required String word})? searchJob;

  // Multi-Currency
  final List<CurrencyModel>? currencies;
  final String? selectedDocCurrencyCode;
  final String? selectedDocCurrencySymbol;
  final double? selectedExchangeRate;
  final String? baseCurrency;
  final String? baseCurrencySymbol;
  final void Function(String? code, String? symbol, double? exchangeRate)? onCurrencyChanged;
  final void Function(double? exchangeRate)? onExchangeRateChanged;

  const PRHeaderWidget({
    super.key,
    required this.screenData,
    required this.formController,
    required this.setState,
    required this.parentContext,
    required this.docDateTimeValidated,
    required this.searchSupplier,
    this.searchDepartment,
    this.searchJob,
    this.currencies,
    this.selectedDocCurrencyCode,
    this.selectedDocCurrencySymbol,
    this.selectedExchangeRate,
    this.baseCurrency,
    this.baseCurrencySymbol,
    this.onCurrencyChanged,
    this.onExchangeRateChanged,
  });

  @override
  State<PRHeaderWidget> createState() => _PRHeaderWidgetState();
}

class _PRHeaderWidgetState extends State<PRHeaderWidget> with global.ThemeRefreshMixin {
  int _selectedUrgency = 1; // 1=ปกติ, 2=เร่งด่วน, 3=เร่งด่วนมาก

  // Repositories สำหรับ Job/Project และ Cost Center
  final _jobProjectRepo = JobProjectRepository();
  final _costCenterRepo = CostCenterRepository();

  // Cache ข้อมูล master
  List<JobProjectModel>? _cachedJobProjects;
  List<CostCenterModel>? _cachedCostCenters;

  // Department dropdown overlay
  final _deptLayerLink = LayerLink();
  final _deptSearchCtrl = TextEditingController();
  OverlayEntry? _deptOverlay;

  // โครงการที่เลือกอยู่ (local state — ไม่ save ไป backend, derive จาก jobcode)
  String _selectedProjectCode = '';
  String _selectedProjectName = '';
  // Display controllers (แสดง "รหัส / ชื่อ" — ไม่ส่งไป backend โดยตรง)
  final _projectNameCtrl = TextEditingController();
  final _jobDisplayCtrl = TextEditingController();
  final _costCenterDisplayCtrl = TextEditingController();

  /// Dynamic vendor preferences list — แต่ละรายการ = {vendor, reason}
  List<Map<String, TextEditingController>> _vendorPreferences = [];

  PRFormController get fc => widget.formController;

  @override
  void initState() {
    super.initState();
    _loadFromScreenData();
  }

  void _loadFromScreenData() {
    _loadVendorPreferences();
    _selectedUrgency = widget.screenData.inquirytype;
    if (_selectedUrgency < 1 || _selectedUrgency > 3) _selectedUrgency = 1;
    // ตั้งค่า display controllers จาก controllers ที่โหลดมาแล้ว (edit mode)
    final jobCode = fc.jobCodeController.text;
    final jobName = fc.jobNameController.text;
    if (jobCode.isNotEmpty) {
      _jobDisplayCtrl.text = jobName.isNotEmpty ? '$jobCode / $jobName' : jobCode;
      _resolveProjectFromJob();
    }
    final ccCode = fc.costCenterCodeController.text;
    final ccName = fc.costCenterNameController.text;
    if (ccCode.isNotEmpty) {
      _costCenterDisplayCtrl.text = ccName.isNotEmpty ? '$ccCode / $ccName' : ccCode;
    }
  }

  /// โหลด job list แล้ว find project ของ job ที่มีอยู่ (สำหรับ edit mode)
  Future<void> _resolveProjectFromJob() async {
    final jobCode = fc.jobCodeController.text;
    if (jobCode.isEmpty) return;
    await _ensureJobProjectsLoaded();
    final all = _cachedJobProjects ?? [];
    final job = all.where((e) => e.code == jobCode).firstOrNull;
    if (job == null || job.parentcode.isEmpty) return;
    final project = all.where((e) => e.code == job.parentcode).firstOrNull;
    if (project == null) return;
    if (mounted) {
      setState(() {
        _selectedProjectCode = project.code;
        _selectedProjectName = global.activeLangName(project.names).isNotEmpty
            ? global.activeLangName(project.names)
            : project.code;
        _projectNameCtrl.text = '${project.code} / $_selectedProjectName';
      });
    }
  }

  Future<void> _ensureJobProjectsLoaded() async {
    if (_cachedJobProjects != null) return;
    try {
      final result = await _jobProjectRepo.getJobProjectList(limit: 999);
      if (result.success) {
        _cachedJobProjects = (result.data as List).map((e) => JobProjectModel.fromJson(e)).toList();
      } else {
        _cachedJobProjects = [];
      }
    } catch (_) {
      _cachedJobProjects = [];
    }
  }

  /// โหลด vendor preferences จาก screenData (JSON array ใน prPreferredVendor)
  void _loadVendorPreferences() {
    _disposeVendorPreferences();
    _vendorPreferences = [];
    final pv = widget.screenData.prPreferredVendor;
    if (pv != null && pv.isNotEmpty) {
      try {
        final list = (json.decode(pv) as List).cast<Map<String, dynamic>>();
        for (final item in list) {
          final v = TextEditingController(text: item['vendor'] ?? '');
          final r = TextEditingController(text: item['reason'] ?? '');
          v.addListener(_autoSyncVendors);
          r.addListener(_autoSyncVendors);
          _vendorPreferences.add({'vendor': v, 'reason': r});
        }
      } catch (_) {
        // fallback: ข้อมูลเก่าเป็น string ไม่ใช่ JSON array
        final v = TextEditingController(text: pv);
        final r = TextEditingController(text: widget.screenData.prReasonPreferred ?? '');
        v.addListener(_autoSyncVendors);
        r.addListener(_autoSyncVendors);
        _vendorPreferences.add({'vendor': v, 'reason': r});
      }
    }
    // เพิ่มอย่างน้อย 1 row ว่าง
    if (_vendorPreferences.isEmpty) {
      _addVendorPreference();
    }
  }

  void _addVendorPreference() {
    setState(() {
      final v = TextEditingController();
      final r = TextEditingController();
      v.addListener(_autoSyncVendors);
      r.addListener(_autoSyncVendors);
      _vendorPreferences.add({'vendor': v, 'reason': r});
    });
  }

  /// Auto-sync vendor data → screenData ทุกครั้งที่พิมพ์
  void _autoSyncVendors() {
    widget.screenData.prPreferredVendor = _serializeVendorPreferences();
  }

  void _removeVendorPreference(int index) {
    if (_vendorPreferences.length <= 1) return;
    setState(() {
      _vendorPreferences[index]['vendor']!.dispose();
      _vendorPreferences[index]['reason']!.dispose();
      _vendorPreferences.removeAt(index);
    });
  }

  /// Serialize vendor preferences เป็น JSON array สำหรับ save
  String _serializeVendorPreferences() {
    final list = _vendorPreferences
        .where((vp) => vp['vendor']!.text.isNotEmpty || vp['reason']!.text.isNotEmpty)
        .map((vp) => {'vendor': vp['vendor']!.text, 'reason': vp['reason']!.text})
        .toList();
    return list.isEmpty ? '' : json.encode(list);
  }

  void _disposeVendorPreferences() {
    for (final vp in _vendorPreferences) {
      vp['vendor']?.dispose();
      vp['reason']?.dispose();
    }
  }

  @override
  void dispose() {
    _deptOverlay?.remove();
    _deptSearchCtrl.dispose();
    _projectNameCtrl.dispose();
    _jobDisplayCtrl.dispose();
    _costCenterDisplayCtrl.dispose();
    _disposeVendorPreferences();
    super.dispose();
  }

  /// Sync vendor preferences กลับไป screenData (เรียกจาก parent ก่อน save)
  void syncVendorPreferencesToScreenData() {
    widget.screenData.prPreferredVendor = _serializeVendorPreferences();
  }

  @override
  void didUpdateWidget(PRHeaderWidget oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (oldWidget.screenData != widget.screenData) {
      _selectedProjectCode = '';
      _selectedProjectName = '';
      _projectNameCtrl.clear();
      _jobDisplayCtrl.clear();
      _costCenterDisplayCtrl.clear();
      _loadFromScreenData();
    }
  }

  // ═══ Shared field decoration ═══
  static const _ic = BoxConstraints(minWidth: 36, minHeight: 0);

  InputDecoration _fd(String label, {IconData? icon, Widget? suffix, bool? isDense}) {
    return InputDecoration(
      labelText: label,
      prefixIcon: icon != null ? Icon(icon, size: 16, color: global.theme.iconColor) : null,
      prefixIconConstraints: _ic,
      suffixIcon: suffix,
      suffixIconConstraints: suffix != null ? _ic : null,
      border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
      contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
      isDense: isDense ?? true,
    );
  }

  @override
  Widget build(BuildContext context) {
    const g = SizedBox(height: 10); // gap between rows within a group
    const fg = SizedBox(width: 10); // gap between fields

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // ═══════════════════════════════════════════════════
          // GROUP 1: ข้อมูลเอกสาร
          // ═══════════════════════════════════════════════════
          _buildGroup(
            children: [
              // Row 1: DocNo + Date + Time + Urgency + DeliveryDate
              Row(children: [
                Expanded(flex: 3, child: _buildReadOnlyField(
                  label: global.language('docno'),
                  value: widget.screenData.docno.isEmpty ? global.language('auto_generate') : widget.screenData.docno,
                  icon: Icons.assignment_outlined,
                )),
                fg,
                Expanded(flex: 2, child: CustomDatePicker(
                  labelText: global.language('doc_date'),
                  useIconSelectDate: true,
                  initialDate: global.safeParseDatetime(widget.screenData.docdatetime).toLocal(),
                  onDateSelected: (date) {
                    if (date != null) {
                      widget.setState(() {
                        final t = global.safeParseDatetime(widget.screenData.docdatetime).toLocal();
                        final combined = DateTime(date.year, date.month, date.day, t.hour, t.minute, t.second);
                        widget.screenData.docdatetime = combined.toUtc().toIso8601String();
                        widget.screenData.docdatelocal = "${date.year}${date.month.toString().padLeft(2, '0')}${date.day.toString().padLeft(2, '0')}";
                        fc.docTimeController.text = DateFormat('HH:mm').format(t);
                      });
                    }
                  },
                  isDense: true,
                  contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                  borderRadius: 8,
                  decoration: const InputDecoration(),
                )),
                fg,
                SizedBox(width: 90, child: CustomTimePicker(
                  labelText: global.language('doc_time'),
                  controller: fc.docTimeController,
                  useIconSelectTime: true,
                  isDense: true,
                  contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                  borderRadius: 8,
                  initialTime: TimeOfDay(
                    hour: global.safeParseDatetime(widget.screenData.docdatetime).toLocal().hour,
                    minute: global.safeParseDatetime(widget.screenData.docdatetime).toLocal().minute,
                  ),
                  onTimeSelected: (time) {
                    if (time != null) {
                      widget.setState(() {
                        final d = global.safeParseDatetime(widget.screenData.docdatetime).toLocal();
                        final combined = DateTime(d.year, d.month, d.day, time.hour, time.minute, 0, 0);
                        widget.screenData.docdatetime = combined.toUtc().toIso8601String();
                        fc.docTimeController.text = '${time.hour.toString().padLeft(2, '0')}:${time.minute.toString().padLeft(2, '0')}';
                      });
                    }
                  },
                )),
                fg,
                Expanded(flex: 2, child: _buildCompactUrgencySelector()),
                fg,
                Expanded(flex: 2, child: CustomDatePicker(
                  key: ValueKey(widget.screenData.docrefdate),
                  labelText: global.language('requested_delivery_date'),
                  useIconSelectDate: true,
                  initialDate: widget.screenData.docrefdate.isNotEmpty
                      ? global.safeParseDatetime(widget.screenData.docrefdate) : DateTime.now(),
                  firstDate: DateTime.now(),
                  lastDate: DateTime.now().add(const Duration(days: 365)),
                  onDateSelected: (date) {
                    if (date != null) {
                      widget.setState(() { widget.screenData.docrefdate = date.toIso8601String(); });
                      setState(() {});
                    }
                  },
                  isDense: true,
                  contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                  borderRadius: 8,
                  decoration: const InputDecoration(),
                )),
              ]),
              g,
              // Row 2: Requester + Department + Purpose
              Row(children: [
                Expanded(child: _buildReadOnlyField(
                  label: global.language('requester_name'),
                  value: global.profileData.name ?? global.profileData.username ?? '',
                  icon: Icons.person_outline,
                )),
                fg,
                Expanded(child: CompositedTransformTarget(
                  link: _deptLayerLink,
                  child: TextFormField(
                    controller: fc.departmentNameController,
                    readOnly: true,
                    decoration: _fd(global.language('department'), icon: Icons.business_outlined,
                      suffix: fc.departmentCodeController.text.isNotEmpty
                        ? _clearBtn(() { widget.setState(() { fc.departmentCodeController.clear(); fc.departmentNameController.clear(); _closeDeptOverlay(); }); })
                        : null,
                    ),
                    style: const TextStyle(fontSize: 13),
                    onTap: _toggleDeptOverlay,
                  ),
                )),
                fg,
                Expanded(flex: 2, child: TextFormField(
                  controller: fc.descriptionController,
                  decoration: _fd('${global.language("purpose")} *', icon: Icons.description_outlined),
                  style: const TextStyle(fontSize: 13),
                  onChanged: (v) { widget.screenData.description = v; },
                )),
              ]),
            ],
          ),
          const SizedBox(height: 10),

          // ═══════════════════════════════════════════════════
          // GROUP 2: คู่ค้า + เงื่อนไข
          // ═══════════════════════════════════════════════════
          _buildGroup(
            children: [
              // Row: Supplier + Credit + DocRef
              Row(children: [
                Expanded(child: TextFormField(
                  controller: fc.custCodeController,
                  readOnly: true,
                  decoration: _fd(global.language('doc_supplier_code'), icon: Icons.store_outlined,
                    suffix: fc.custCodeController.text.isNotEmpty
                      ? _clearBtn(() { widget.setState(() { fc.custCodeController.clear(); fc.custnamesController.clear(); widget.screenData.custcode = ''; widget.screenData.custnames = []; }); })
                      : null,
                  ),
                  style: const TextStyle(fontSize: 13),
                  onTap: () => widget.searchSupplier(word: ''),
                )),
                fg,
                Expanded(flex: 2, child: TextFormField(
                  controller: fc.custnamesController,
                  readOnly: true,
                  decoration: _fd(global.language('doc_supplier_name')),
                  style: const TextStyle(fontSize: 13),
                  onTap: () => widget.searchSupplier(word: ''),
                )),
                fg,
                SizedBox(width: 100, child: TextFormField(
                  controller: fc.creditDaysController,
                  keyboardType: TextInputType.number,
                  decoration: _fd(global.language('credit_days'), icon: Icons.schedule_outlined),
                  style: const TextStyle(fontSize: 13),
                  onChanged: (v) { widget.screenData.creditdays = int.tryParse(v) ?? 0; },
                )),
                fg,
                Expanded(child: TextFormField(
                  controller: fc.docRefNumberController,
                  decoration: _fd(global.language('doc_ref'), icon: Icons.link_outlined),
                  style: const TextStyle(fontSize: 13),
                  onChanged: (v) { widget.screenData.docrefno = v; },
                )),
              ]),
              g,
              // Row: VAT Type + Rate + Currency (if multi)
              Row(children: [
                Expanded(child: _buildVatTypeRow()),
                if (widget.currencies != null && widget.currencies!.length > 1) ...[
                  fg, fg,
                  Expanded(child: _buildCurrencySection()),
                ],
              ]),
            ],
          ),
          const SizedBox(height: 10),

          // ═══════════════════════════════════════════════════
          // GROUP 3: โครงการ + งบประมาณ
          // ═══════════════════════════════════════════════════
          _buildGroup(
            children: [
              // Row: Project + Job + CostCenter + ShipTo
              _buildJobAndShipToRow(),
              g,
              // Row: Budget + Estimated Cost (5 fields)
              Row(children: [
                Expanded(child: TextFormField(
                  controller: fc.budgetCodeController,
                  decoration: _fd(global.language('budget_code'), icon: Icons.tag),
                  style: const TextStyle(fontSize: 13),
                )),
                fg,
                Expanded(child: TextFormField(
                  controller: fc.budgetAmountController,
                  keyboardType: const TextInputType.numberWithOptions(decimal: true),
                  decoration: _fd(global.language('budget_amount'), icon: Icons.monetization_on_outlined),
                  style: const TextStyle(fontSize: 13),
                  onChanged: (_) => setState(() {}),
                )),
                fg,
                Expanded(child: TextFormField(
                  controller: fc.estimatedUnitCostController,
                  keyboardType: const TextInputType.numberWithOptions(decimal: true),
                  decoration: _fd(global.language('estimated_unit_cost'), icon: Icons.calculate_outlined),
                  style: const TextStyle(fontSize: 13),
                )),
                fg,
                Expanded(child: TextFormField(
                  controller: fc.estimatedFreightController,
                  keyboardType: const TextInputType.numberWithOptions(decimal: true),
                  decoration: _fd(global.language('estimated_freight'), icon: Icons.local_shipping_outlined),
                  style: const TextStyle(fontSize: 13),
                )),
                fg,
                Expanded(child: TextFormField(
                  controller: fc.estimatedDutyController,
                  keyboardType: const TextInputType.numberWithOptions(decimal: true),
                  decoration: _fd(global.language('estimated_duty'), icon: Icons.account_balance_outlined),
                  style: const TextStyle(fontSize: 13),
                )),
              ]),
              // Budget Warning
              Builder(builder: (_) {
                final ba = double.tryParse(fc.budgetAmountController.text) ?? 0;
                final ta = widget.screenData.totalamount;
                if (ba > 0 && ta > ba) {
                  final p = ((ta - ba) / ba * 100).toStringAsFixed(1);
                  return Container(
                    margin: const EdgeInsets.only(top: 4),
                    padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 4),
                    decoration: BoxDecoration(
                      color: global.theme.negativeHighlightColor.withValues(alpha: 0.15),
                      borderRadius: BorderRadius.circular(6),
                    ),
                    child: Row(children: [
                      Icon(Icons.warning_amber_rounded, size: 14, color: global.theme.negativeHighlightColor),
                      const SizedBox(width: 6),
                      Text('${global.language("budget_over_warning")} $p% (${global.formatNumber(ta)} / ${global.formatNumber(ba)})',
                        style: TextStyle(fontSize: 11, color: global.theme.negativeHighlightColor, fontWeight: FontWeight.bold)),
                    ]),
                  );
                }
                return const SizedBox.shrink();
              }),
            ],
          ),
          const SizedBox(height: 10),

          // ═══════════════════════════════════════════════════
          // GROUP 4: อื่น ๆ
          // ═══════════════════════════════════════════════════
          _buildGroup(
            children: [
              // Vendor preferences
              _buildVendorPreferencesList(),
              g,
              // Row: Internal Note + Group + Tolerance + Approval
              Row(children: [
                Expanded(flex: 2, child: TextFormField(
                  controller: fc.internalNoteController,
                  decoration: _fd(global.language('internal_note'), icon: Icons.lock_outline),
                  style: const TextStyle(fontSize: 13),
                )),
                fg,
                Expanded(child: TextFormField(
                  controller: fc.purchasingGroupController,
                  decoration: _fd(global.language('purchasing_group'), icon: Icons.groups_outlined),
                  style: const TextStyle(fontSize: 13),
                )),
                fg,
                SizedBox(width: 130, child: TextFormField(
                  controller: fc.overDeliveryToleranceController,
                  keyboardType: const TextInputType.numberWithOptions(decimal: true),
                  decoration: _fd('${global.language("over_delivery_tolerance")} %', icon: Icons.add_circle_outline),
                  style: const TextStyle(fontSize: 13),
                )),
                fg,
                SizedBox(width: 130, child: TextFormField(
                  controller: fc.underDeliveryToleranceController,
                  keyboardType: const TextInputType.numberWithOptions(decimal: true),
                  decoration: _fd('${global.language("under_delivery_tolerance")} %', icon: Icons.remove_circle_outline),
                  style: const TextStyle(fontSize: 13),
                )),
                fg,
                SizedBox(width: 200, child: CustomDatePicker(
                  labelText: global.language('approval_deadline'),
                  useIconSelectDate: true,
                  initialDate: fc.approvalDeadlineController.text.isNotEmpty
                      ? global.safeParseDatetime(fc.approvalDeadlineController.text)
                      : DateTime.now().add(const Duration(days: 3)),
                  firstDate: DateTime.now(),
                  lastDate: DateTime.now().add(const Duration(days: 365)),
                  onDateSelected: (date) {
                    if (date != null) { setState(() { fc.approvalDeadlineController.text = date.toIso8601String(); }); }
                  },
                  isDense: true,
                  contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                  borderRadius: 8,
                  decoration: const InputDecoration(),
                )),
              ]),
              // Tracking (if any)
              if (_hasTrackingData()) ...[g, _buildTrackingRow()],
            ],
          ),
        ],
      ),
    );
  }

  /// Visual group container — subtle background + rounded corners
  Widget _buildGroup({required List<Widget> children}) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 12),
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: global.theme.dividerBorderColor.withValues(alpha: 0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: children,
      ),
    );
  }

  /// Clear button สำหรับ field ที่มีค่า
  Widget _clearBtn(VoidCallback onTap) {
    return IconButton(
      icon: Icon(Icons.clear, size: 14, color: global.theme.textSecondaryColor),
      padding: EdgeInsets.zero,
      constraints: const BoxConstraints(minWidth: 24, minHeight: 24),
      onPressed: onTap,
    );
  }

  /// Urgency selector แบบกระชับ (ไม่มี InputDecorator wrapper)
  Widget _buildCompactUrgencySelector() {
    return Container(
      height: 38,
      decoration: BoxDecoration(
        border: Border.all(color: global.theme.dividerBorderColor),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Row(
        children: [
          _buildUrgencyChip(1, global.language('urgency_normal'), global.theme.positiveHighlightTextColor),
          _buildUrgencyChip(2, global.language('urgency_urgent'), global.theme.warningHighlightTextColor),
          _buildUrgencyChip(3, global.language('urgency_critical'), global.theme.negativeHighlightTextColor),
        ],
      ),
    );
  }

  Widget _buildUrgencyChip(int level, String label, Color activeColor) {
    final isSelected = _selectedUrgency == level;
    return Expanded(
      child: InkWell(
        onTap: () {
          setState(() { _selectedUrgency = level; });
          widget.setState(() { widget.screenData.inquirytype = level; });
        },
        child: Container(
          padding: const EdgeInsets.symmetric(vertical: 10),
          decoration: BoxDecoration(
            color: isSelected ? activeColor.withValues(alpha: 0.15) : null,
            borderRadius: BorderRadius.circular(7),
          ),
          child: Center(
            child: Text(
              label,
              style: TextStyle(
                color: isSelected ? activeColor : global.theme.textSecondaryColor,
                fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
                fontSize: 13,
              ),
            ),
          ),
        ),
      ),
    );
  }

  // ──────────────────────────────────────────────────────
  // Row 5: ประเภทภาษี VAT
  // ──────────────────────────────────────────────────────
  Widget _buildVatTypeRow() {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Expanded(
          child: InputDecorator(
            decoration: InputDecoration(
              labelText: global.language('vat_type'),
              border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
              contentPadding: const EdgeInsets.symmetric(horizontal: 4, vertical: 4),
            ),
            isEmpty: false,
            child: Row(
              children: [
                _buildVatTypeChip(0, global.language('vat_exclude')),
                _buildVatTypeChip(1, global.language('vat_include')),
                _buildVatTypeChip(2, global.language('vat_zero')),
                _buildVatTypeChip(3, global.language('vat_none')),
              ],
            ),
          ),
        ),
        const SizedBox(width: 12),
        // VAT Rate
        SizedBox(
          width: 120,
          child: TextFormField(
            controller: fc.vatRateController,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            decoration: InputDecoration(
              labelText: global.language('vat_rate'),
              suffixText: '%',
              border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
              contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
            ),
            onChanged: (value) {
              widget.screenData.vatrate = double.tryParse(value) ?? 0;
            },
          ),
        ),
      ],
    );
  }

  Widget _buildVatTypeChip(int type, String label) {
    final isSelected = widget.screenData.vattype == type;
    return Expanded(
      child: InkWell(
        onTap: () {
          widget.setState(() {
            widget.screenData.vattype = type;
          });
          setState(() {});
        },
        child: Container(
          padding: const EdgeInsets.symmetric(vertical: 10),
          decoration: BoxDecoration(
            color: isSelected ? global.theme.infoHighlightTextColor.withValues(alpha: 0.15) : null,
            borderRadius: BorderRadius.circular(7),
          ),
          child: Center(
            child: Text(
              label,
              style: TextStyle(
                color: isSelected ? global.theme.infoHighlightTextColor : global.theme.textSecondaryColor,
                fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
                fontSize: 13,
              ),
            ),
          ),
        ),
      ),
    );
  }

  // ──────────────────────────────────────────────────────
  // Row 6: โครงการ + งาน + Cost Center
  // Row 7: ที่อยู่จัดส่ง (แยกออกมาเป็น multiline)
  // ──────────────────────────────────────────────────────
  Widget _buildJobAndShipToRow() {
    final hasProject = _selectedProjectCode.isNotEmpty;
    final hasJob = fc.jobNameController.text.isNotEmpty || fc.jobCodeController.text.isNotEmpty;
    final hasCostCenter = fc.costCenterNameController.text.isNotEmpty || fc.costCenterCodeController.text.isNotEmpty;

    return Column(
      children: [
        // แถว 1: โครงการ + งาน + ศูนย์ต้นทุน
        Row(
          children: [
            // Field 1: โครงการ — เลือกก่อน
            Expanded(
              child: GestureDetector(
                onTap: _showProjectDialog,
                child: AbsorbPointer(
                  child: TextFormField(
                    controller: _projectNameCtrl,
                    readOnly: true,
                    decoration: InputDecoration(
                      labelText: global.language('project'),
                      prefixIcon: Icon(Icons.account_tree_outlined, size: 20, color: global.theme.iconColor),
                      prefixIconConstraints: const BoxConstraints(minWidth: 36, minHeight: 0),
                      suffixIcon: hasProject
                          ? IconButton(
                              icon: Icon(Icons.clear, size: 16, color: global.theme.iconSecondaryColor),
                              onPressed: () {
                                setState(() {
                                  _selectedProjectCode = '';
                                  _selectedProjectName = '';
                                  _projectNameCtrl.clear();
                                  fc.jobCodeController.clear();
                                  fc.jobNameController.clear();
                                });
                              },
                            )
                          : Icon(Icons.search, size: 18, color: global.theme.iconSecondaryColor),
                      suffixIconConstraints: const BoxConstraints(minWidth: 36, minHeight: 0),
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                      contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                      isDense: true,
                    ),
                    style: const TextStyle(fontSize: 13),
                  ),
                ),
              ),
            ),
            const SizedBox(width: 10),
            // Field 2: งาน — เลือกได้เมื่อเลือกโครงการแล้ว
            Expanded(
              child: GestureDetector(
                onTap: hasProject ? _showJobDialog : null,
                child: AbsorbPointer(
                  child: TextFormField(
                    controller: _jobDisplayCtrl,
                    readOnly: true,
                    enabled: hasProject,
                    decoration: InputDecoration(
                      labelText: global.language('job'),
                      prefixIcon: Icon(Icons.work_outline, size: 20, color: hasProject ? global.theme.iconColor : global.theme.iconSecondaryColor),
                      prefixIconConstraints: const BoxConstraints(minWidth: 36, minHeight: 0),
                      suffixIcon: hasJob
                          ? IconButton(
                              icon: Icon(Icons.clear, size: 16, color: global.theme.iconSecondaryColor),
                              onPressed: () {
                                setState(() {
                                  fc.jobCodeController.clear();
                                  fc.jobNameController.clear();
                                  _jobDisplayCtrl.clear();
                                });
                              },
                            )
                          : Icon(Icons.search, size: 18, color: global.theme.iconSecondaryColor),
                      suffixIconConstraints: const BoxConstraints(minWidth: 36, minHeight: 0),
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                      contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                      isDense: true,
                    ),
                    style: const TextStyle(fontSize: 13),
                  ),
                ),
              ),
            ),
            const SizedBox(width: 10),
            // Cost Center — searchable datalist
            Expanded(
              child: GestureDetector(
                onTap: _showCostCenterDialog,
                child: AbsorbPointer(
                  child: TextFormField(
                    controller: _costCenterDisplayCtrl,
                    readOnly: true,
                    decoration: InputDecoration(
                      labelText: global.language('cost_center'),
                      prefixIcon: Icon(Icons.account_balance_outlined, size: 20, color: global.theme.iconColor),
                      prefixIconConstraints: const BoxConstraints(minWidth: 36, minHeight: 0),
                      suffixIcon: hasCostCenter
                          ? IconButton(
                              icon: Icon(Icons.clear, size: 16, color: global.theme.iconSecondaryColor),
                              onPressed: () {
                                setState(() {
                                  fc.costCenterCodeController.clear();
                                  fc.costCenterNameController.clear();
                                  _costCenterDisplayCtrl.clear();
                                });
                              },
                            )
                          : Icon(Icons.search, size: 18, color: global.theme.iconSecondaryColor),
                      suffixIconConstraints: const BoxConstraints(minWidth: 36, minHeight: 0),
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                      contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                      isDense: true,
                    ),
                    style: const TextStyle(fontSize: 13),
                  ),
                ),
              ),
            ),
          ],
        ),
        const SizedBox(height: 10),
        // แถว 2: ที่อยู่จัดส่ง (พิมพ์เองได้) + ปุ่ม GPS + พิกัด — ความสูงเท่ากัน
        IntrinsicHeight(
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.stretch,
            children: [
              // ที่อยู่จัดส่ง
              Expanded(
                child: TextFormField(
                  controller: fc.shipToAddressController,
                  maxLines: 3,
                  minLines: 2,
                  decoration: InputDecoration(
                    labelText: global.language('ship_to_address'),
                    prefixIcon: Padding(
                      padding: const EdgeInsets.only(bottom: 24),
                      child: Icon(Icons.local_shipping_outlined, size: 20, color: global.theme.iconColor),
                    ),
                    prefixIconConstraints: const BoxConstraints(minWidth: 36, minHeight: 0),
                    border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                    contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                  ),
                  style: const TextStyle(fontSize: 13),
                ),
              ),
              const SizedBox(width: 8),
              // ปุ่ม GPS + พิกัด — ใช้ Container เดียวให้สูงเท่ากัน
              Container(
                width: fc.shipToLatController.text.isNotEmpty ? 170 : 42,
                decoration: BoxDecoration(
                  color: fc.shipToLatController.text.isNotEmpty
                      ? global.theme.positiveHighlightTextColor.withValues(alpha: 0.06)
                      : global.theme.surfaceColor,
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: fc.shipToLatController.text.isNotEmpty
                      ? global.theme.positiveHighlightTextColor.withValues(alpha: 0.3)
                      : global.theme.dividerBorderColor),
                ),
                child: InkWell(
                  onTap: fc.shipToLatController.text.isEmpty ? _showShipToDialog : null,
                  borderRadius: BorderRadius.circular(8),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      // ปุ่ม GPS
                      IconButton(
                        onPressed: _showShipToDialog,
                        icon: Icon(
                          fc.shipToLatController.text.isNotEmpty ? Icons.gps_fixed : Icons.gps_not_fixed,
                          size: 20,
                          color: fc.shipToLatController.text.isNotEmpty
                              ? global.theme.positiveHighlightTextColor
                              : global.theme.iconColor,
                        ),
                        tooltip: global.language('select_location'),
                        padding: EdgeInsets.zero,
                        constraints: const BoxConstraints(minWidth: 36, minHeight: 36),
                      ),
                      // พิกัด + ปุ่ม copy
                      if (fc.shipToLatController.text.isNotEmpty) ...[
                        Expanded(
                          child: Text(
                            '${fc.shipToLatController.text}, ${fc.shipToLngController.text}',
                            style: TextStyle(
                              fontSize: 11,
                              color: global.theme.positiveHighlightTextColor,
                              fontWeight: FontWeight.w500,
                            ),
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                        IconButton(
                          onPressed: () {
                            final coords = '${fc.shipToLatController.text}, ${fc.shipToLngController.text}';
                            Clipboard.setData(ClipboardData(text: coords));
                            ScaffoldMessenger.of(context).showSnackBar(
                              SnackBar(
                                content: Row(children: [
                                  const Icon(Icons.check, color: Colors.white, size: 16),
                                  const SizedBox(width: 8),
                                  Text(coords),
                                ]),
                                duration: const Duration(seconds: 2),
                              ),
                            );
                          },
                          icon: Icon(Icons.copy, size: 14, color: global.theme.iconSecondaryColor),
                          padding: EdgeInsets.zero,
                          constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                          tooltip: 'Copy',
                        ),
                      ],
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  // ──────────────────────────────────────────────────────
  // Dialog: เลือกโครงการ
  // ──────────────────────────────────────────────────────
  Future<void> _showProjectDialog() async {
    await _ensureJobProjectsLoaded();
    if (!mounted) return;

    String searchText = '';

    await showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) {
          final all = _cachedJobProjects ?? [];
          final filtered = all
              .where((e) => e.parentcode.isEmpty)
              .where((e) => searchText.isEmpty || global.activeLangName(e.names).toLowerCase().contains(searchText.toLowerCase()) || e.code.toLowerCase().contains(searchText.toLowerCase()))
              .toList();

          return AlertDialog(
            title: Text(global.language('project'), style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
            contentPadding: const EdgeInsets.fromLTRB(12, 8, 12, 0),
            content: SizedBox(
              width: 360,
              height: 420,
              child: Column(
                children: [
                  TextField(
                    autofocus: true,
                    decoration: InputDecoration(
                      hintText: global.language('search'),
                      prefixIcon: const Icon(Icons.search, size: 18),
                      contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                      isDense: true,
                    ),
                    onChanged: (v) => setDialogState(() => searchText = v),
                  ),
                  const SizedBox(height: 10),
                  Expanded(
                    child: filtered.isEmpty
                        ? Center(child: Text(global.language('no_data'), style: TextStyle(color: global.theme.textSecondaryColor)))
                        : ListView.builder(
                            itemCount: filtered.length,
                            itemBuilder: (_, i) {
                              final item = filtered[i];
                              final name = global.activeLangName(item.names);
                              return ListTile(
                                dense: true,
                                leading: Icon(Icons.account_tree_outlined, size: 18, color: global.theme.iconColor),
                                title: Text(name.isNotEmpty ? name : item.code, style: const TextStyle(fontSize: 14)),
                                subtitle: Text(item.code, style: TextStyle(fontSize: 12, color: global.theme.textSecondaryColor)),
                                trailing: const Icon(Icons.chevron_right, size: 16),
                                onTap: () {
                                  setState(() {
                                    _selectedProjectCode = item.code;
                                    _selectedProjectName = name.isNotEmpty ? name : item.code;
                                    _projectNameCtrl.text = '${item.code} / $_selectedProjectName';
                                    // ล้างงานเก่าเมื่อเปลี่ยนโครงการ
                                    fc.jobCodeController.clear();
                                    fc.jobNameController.clear();
                                    _jobDisplayCtrl.clear();
                                  });
                                  Navigator.pop(ctx);
                                },
                              );
                            },
                          ),
                  ),
                ],
              ),
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(ctx),
                child: Text(global.language('cancel')),
              ),
            ],
          );
        },
      ),
    );
  }

  // ──────────────────────────────────────────────────────
  // Dialog: เลือกงาน (ภายใต้โครงการที่เลือกไว้)
  // ──────────────────────────────────────────────────────
  Future<void> _showJobDialog() async {
    if (_selectedProjectCode.isEmpty) return;
    await _ensureJobProjectsLoaded();
    if (!mounted) return;

    String searchText = '';

    await showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) {
          final all = _cachedJobProjects ?? [];
          final filtered = all
              .where((e) => e.parentcode == _selectedProjectCode)
              .where((e) => searchText.isEmpty || global.activeLangName(e.names).toLowerCase().contains(searchText.toLowerCase()) || e.code.toLowerCase().contains(searchText.toLowerCase()))
              .toList();

          return AlertDialog(
            title: Text('${global.language('job')} — $_selectedProjectCode / $_selectedProjectName', style: const TextStyle(fontSize: 14, fontWeight: FontWeight.bold), overflow: TextOverflow.ellipsis),
            contentPadding: const EdgeInsets.fromLTRB(12, 8, 12, 0),
            content: SizedBox(
              width: 360,
              height: 420,
              child: Column(
                children: [
                  TextField(
                    autofocus: true,
                    decoration: InputDecoration(
                      hintText: global.language('search'),
                      prefixIcon: const Icon(Icons.search, size: 18),
                      contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                      isDense: true,
                    ),
                    onChanged: (v) => setDialogState(() => searchText = v),
                  ),
                  const SizedBox(height: 10),
                  Expanded(
                    child: filtered.isEmpty
                        ? Center(child: Text(global.language('no_data'), style: TextStyle(color: global.theme.textSecondaryColor)))
                        : ListView.builder(
                            itemCount: filtered.length,
                            itemBuilder: (_, i) {
                              final item = filtered[i];
                              final name = global.activeLangName(item.names);
                              return ListTile(
                                dense: true,
                                leading: Icon(Icons.work_outline, size: 18, color: global.theme.iconColor),
                                title: Text(name.isNotEmpty ? name : item.code, style: const TextStyle(fontSize: 14)),
                                subtitle: Text(item.code, style: TextStyle(fontSize: 12, color: global.theme.textSecondaryColor)),
                                onTap: () {
                                  setState(() {
                                    fc.jobCodeController.text = item.code;
                                    fc.jobNameController.text = name.isNotEmpty ? '${item.code} / $name' : item.code;
                                    _jobDisplayCtrl.text = name.isNotEmpty ? '${item.code} / $name' : item.code;
                                  });
                                  Navigator.pop(ctx);
                                },
                              );
                            },
                          ),
                  ),
                ],
              ),
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(ctx),
                child: Text(global.language('cancel')),
              ),
            ],
          );
        },
      ),
    );
  }

  // ──────────────────────────────────────────────────────
  // Cost Center searchable dialog
  // ──────────────────────────────────────────────────────
  // ──────────────────────────────────────────────────────
  // Dropdown overlay: เลือกแผนกจาก master data (แสดงใต้ textbox)
  // ──────────────────────────────────────────────────────
  void _toggleDeptOverlay() {
    if (_deptOverlay != null) {
      _closeDeptOverlay();
    } else {
      _showDeptOverlay();
    }
  }

  void _closeDeptOverlay() {
    _deptOverlay?.remove();
    _deptOverlay = null;
    _deptSearchCtrl.clear();
  }

  void _showDeptOverlay() {
    _closeDeptOverlay();
    final departments = global.companyBranchSelectData.departments;
    if (departments.isEmpty) return;

    _deptOverlay = OverlayEntry(
      builder: (context) => _DeptDropdownOverlay(
        link: _deptLayerLink,
        departments: departments,
        searchCtrl: _deptSearchCtrl,
        onSelect: (item) {
          final name = global.activeLangName(item.names);
          setState(() {
            fc.departmentCodeController.text = item.code;
            fc.departmentNameController.text = name.isNotEmpty ? '${item.code} / $name' : item.code;
          });
          _closeDeptOverlay();
        },
        onClose: _closeDeptOverlay,
      ),
    );
    Overlay.of(context).insert(_deptOverlay!);
  }

  Future<void> _showCostCenterDialog() async {
    if (_cachedCostCenters == null) {
      try {
        final result = await _costCenterRepo.getCostCenterList(limit: 999);
        if (result.success) {
          _cachedCostCenters = (result.data as List).map((e) => CostCenterModel.fromJson(e)).toList();
        } else {
          _cachedCostCenters = [];
        }
      } catch (_) {
        _cachedCostCenters = [];
      }
    }

    if (!mounted) return;

    String searchText = '';

    await showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) {
          final all = _cachedCostCenters ?? [];
          final filtered = all.where((e) => searchText.isEmpty || global.activeLangName(e.names).toLowerCase().contains(searchText.toLowerCase()) || e.code.toLowerCase().contains(searchText.toLowerCase())).toList();

          return AlertDialog(
            title: Text(global.language('cost_center'), style: const TextStyle(fontSize: 16, fontWeight: FontWeight.bold)),
            contentPadding: const EdgeInsets.fromLTRB(12, 8, 12, 0),
            content: SizedBox(
              width: 360,
              height: 380,
              child: Column(
                children: [
                  TextField(
                    autofocus: true,
                    decoration: InputDecoration(
                      hintText: global.language('search'),
                      prefixIcon: const Icon(Icons.search, size: 18),
                      contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                      isDense: true,
                    ),
                    onChanged: (v) => setDialogState(() => searchText = v),
                  ),
                  const SizedBox(height: 10),
                  Expanded(
                    child: filtered.isEmpty
                        ? Center(child: Text(global.language('no_data'), style: TextStyle(color: global.theme.textSecondaryColor)))
                        : ListView.builder(
                            itemCount: filtered.length,
                            itemBuilder: (_, i) {
                              final item = filtered[i];
                              final name = global.activeLangName(item.names);
                              return ListTile(
                                dense: true,
                                leading: Icon(Icons.account_balance_outlined, size: 18, color: global.theme.iconColor),
                                title: Text(name.isNotEmpty ? name : item.code, style: const TextStyle(fontSize: 14)),
                                subtitle: Text(item.code, style: TextStyle(fontSize: 12, color: global.theme.textSecondaryColor)),
                                onTap: () {
                                  setState(() {
                                    fc.costCenterCodeController.text = item.code;
                                    fc.costCenterNameController.text = name.isNotEmpty ? '${item.code} / $name' : item.code;
                                    _costCenterDisplayCtrl.text = name.isNotEmpty ? '${item.code} / $name' : item.code;
                                  });
                                  Navigator.pop(ctx);
                                },
                              );
                            },
                          ),
                  ),
                ],
              ),
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(ctx),
                child: Text(global.language('cancel')),
              ),
            ],
          );
        },
      ),
    );
  }

  // ──────────────────────────────────────────────────────
  // Ship-to dialog with GPS
  // ──────────────────────────────────────────────────────
  Future<void> _showShipToDialog() async {
    final addressCtrl = TextEditingController(text: fc.shipToAddressController.text);
    final latCtrl = TextEditingController(text: fc.shipToLatController.text);
    final lngCtrl = TextEditingController(text: fc.shipToLngController.text);
    final searchCtrl = TextEditingController();
    final mapController = MapController();

    final initLat = double.tryParse(fc.shipToLatController.text) ?? 0.0;
    final initLng = double.tryParse(fc.shipToLngController.text) ?? 0.0;
    LatLng? selectedPoint = (initLat != 0 || initLng != 0) ? LatLng(initLat, initLng) : null;
    bool isGettingLocation = false;
    int mapType = 0; // 0 = OSM, 1 = Google
    List<Map<String, dynamic>> searchResults = [];
    bool isSearching = false;
    bool showSearchResults = false;

    await showDialog<void>(
      context: widget.parentContext,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) {
          Future<void> searchLocation(String query) async {
            if (query.trim().isEmpty) {
              setDialogState(() { searchResults.clear(); showSearchResults = false; });
              return;
            }
            setDialogState(() => isSearching = true);
            try {
              final uri = Uri.https('nominatim.openstreetmap.org', '/search', {
                'q': query.trim(),
                'format': 'json',
                'accept-language': 'th,en',
                'countrycodes': 'th',
                'limit': '8',
              });
              final resp = await http.get(uri, headers: {
                'User-Agent': 'BCAccountApp/1.0 (support@bcaicloud.com)',
                'Accept-Language': 'th,en;q=0.9',
                'Accept': 'application/json',
              });
              if (resp.statusCode == 200) {
                final body = utf8.decode(resp.bodyBytes);
                final list = json.decode(body) as List<dynamic>;
                setDialogState(() {
                  searchResults = list.cast<Map<String, dynamic>>();
                  showSearchResults = true;
                });
              }
            } catch (_) {
              // ignore network errors silently
            } finally {
              setDialogState(() => isSearching = false);
            }
          }

          Future<void> getCurrentLocation() async {
            setDialogState(() => isGettingLocation = true);
            try {
              bool serviceEnabled = await Geolocator.isLocationServiceEnabled();
              if (!serviceEnabled) {
                setDialogState(() => isGettingLocation = false);
                return;
              }
              LocationPermission permission = await Geolocator.checkPermission();
              if (permission == LocationPermission.denied) {
                permission = await Geolocator.requestPermission();
              }
              if (permission == LocationPermission.denied || permission == LocationPermission.deniedForever) {
                setDialogState(() => isGettingLocation = false);
                return;
              }
              final position = await Geolocator.getCurrentPosition(
                locationSettings: const LocationSettings(accuracy: LocationAccuracy.high, timeLimit: Duration(seconds: 15)),
              );
              final pt = LatLng(position.latitude, position.longitude);
              setDialogState(() {
                isGettingLocation = false;
                selectedPoint = pt;
                latCtrl.text = pt.latitude.toStringAsFixed(6);
                lngCtrl.text = pt.longitude.toStringAsFixed(6);
              });
              mapController.move(pt, 15);
            } catch (_) {
              setDialogState(() => isGettingLocation = false);
            }
          }

          final sw = MediaQuery.of(ctx).size.width;
          final sh = MediaQuery.of(ctx).size.height;
          double currentZoom = selectedPoint != null ? 15 : 7;
          return Dialog(
            insetPadding: EdgeInsets.symmetric(
              horizontal: sw * 0.025,
              vertical: sh * 0.025,
            ),
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Title bar
                Padding(
                  padding: const EdgeInsets.fromLTRB(20, 14, 8, 10),
                  child: Row(
                    children: [
                      Icon(Icons.local_shipping_outlined, size: 20, color: global.theme.primaryColor),
                      const SizedBox(width: 8),
                      Expanded(child: Text(global.language('ship_to_address'), style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600))),
                      IconButton(icon: const Icon(Icons.close, size: 20), onPressed: () => Navigator.pop(ctx), padding: EdgeInsets.zero),
                    ],
                  ),
                ),
                const Divider(height: 1),
                // Address field
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 10, 16, 8),
                  child: TextFormField(
                    controller: addressCtrl,
                    maxLines: 2,
                    decoration: InputDecoration(
                      labelText: global.language('address'),
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                      contentPadding: const EdgeInsets.all(12),
                    ),
                  ),
                ),
                // Map — Expanded เต็มพื้นที่ที่เหลือ
                Expanded(
                  child: Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 16),
                    child: ClipRRect(
                      borderRadius: BorderRadius.circular(8),
                      child: Stack(
                        children: [
                          FlutterMap(
                            mapController: mapController,
                            options: MapOptions(
                              initialCenter: selectedPoint ?? const LatLng(13.7563, 100.5018),
                              initialZoom: currentZoom,
                              interactionOptions: const InteractionOptions(
                                flags: ~InteractiveFlag.doubleTapZoom,
                              ),
                              onTap: (_, point) {
                                setDialogState(() {
                                  selectedPoint = point;
                                  latCtrl.text = point.latitude.toStringAsFixed(6);
                                  lngCtrl.text = point.longitude.toStringAsFixed(6);
                                });
                              },
                              onPositionChanged: (pos, _) {
                                currentZoom = pos.zoom;
                              },
                            ),
                            children: [
                              if (mapType == 0)
                                TileLayer(
                                  urlTemplate: 'https://tile.openstreetmap.org/{z}/{x}/{y}.png',
                                  userAgentPackageName: 'com.bcaicloud.bcaiaccount',
                                )
                              else
                                TileLayer(
                                  urlTemplate: 'https://mt{s}.google.com/vt/lyrs=y&x={x}&y={y}&z={z}',
                                  subdomains: const ['0', '1', '2', '3'],
                                  userAgentPackageName: 'com.bcaicloud.bcaiaccount',
                                ),
                              if (selectedPoint != null)
                                MarkerLayer(markers: [
                                  Marker(
                                    point: selectedPoint!,
                                    width: 48,
                                    height: 48,
                                    child: Icon(Icons.location_pin, size: 48, color: global.theme.negativeHighlightTextColor),
                                  ),
                                ]),
                            ],
                          ),
                          // Zoom + GPS buttons (column, right side)
                          Positioned(
                            right: 8,
                            bottom: 8,
                            child: Column(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                // Toggle map type
                                FloatingActionButton.small(
                                  heroTag: 'shipToMapType',
                                  onPressed: () => setDialogState(() => mapType = mapType == 0 ? 1 : 0),
                                  backgroundColor: mapType == 1 ? global.theme.primaryColor : global.theme.cardColor,
                                  child: Icon(Icons.satellite_alt, size: 18, color: mapType == 1 ? global.theme.cardColor : global.theme.textColor),
                                ),
                                const SizedBox(height: 6),
                                // Zoom in
                                FloatingActionButton.small(
                                  heroTag: 'shipToZoomIn',
                                  onPressed: () => mapController.move(mapController.camera.center, mapController.camera.zoom + 1),
                                  backgroundColor: global.theme.cardColor,
                                  child: Icon(Icons.add, size: 20, color: global.theme.textColor),
                                ),
                                const SizedBox(height: 6),
                                // Zoom out
                                FloatingActionButton.small(
                                  heroTag: 'shipToZoomOut',
                                  onPressed: () => mapController.move(mapController.camera.center, mapController.camera.zoom - 1),
                                  backgroundColor: global.theme.cardColor,
                                  child: Icon(Icons.remove, size: 20, color: global.theme.textColor),
                                ),
                                const SizedBox(height: 6),
                                // GPS
                                FloatingActionButton.small(
                                  heroTag: 'shipToGPS',
                                  onPressed: isGettingLocation ? null : getCurrentLocation,
                                  backgroundColor: global.theme.cardColor,
                                  child: isGettingLocation
                                      ? const SizedBox(width: 16, height: 16, child: CircularProgressIndicator(strokeWidth: 2))
                                      : Icon(Icons.my_location, size: 20, color: global.theme.infoHighlightTextColor),
                                ),
                              ],
                            ),
                          ),
                          // Search bar overlay (top, leave right gap for FAB column)
                          Positioned(
                            top: 8,
                            left: 8,
                            right: 52,
                            child: Column(
                              children: [
                                Material(
                                  elevation: 4,
                                  borderRadius: BorderRadius.circular(8),
                                  color: global.theme.cardColor,
                                  child: TextField(
                                    controller: searchCtrl,
                                    decoration: InputDecoration(
                                      hintText: global.language('search_location'),
                                      hintStyle: TextStyle(color: global.theme.textSecondaryColor, fontSize: 13),
                                      prefixIcon: isSearching
                                          ? const Padding(
                                              padding: EdgeInsets.all(10),
                                              child: SizedBox(width: 18, height: 18, child: CircularProgressIndicator(strokeWidth: 2)),
                                            )
                                          : Icon(Icons.search, size: 20, color: global.theme.iconColor),
                                      suffixIcon: searchCtrl.text.isNotEmpty
                                          ? IconButton(
                                              icon: Icon(Icons.clear, size: 16, color: global.theme.textSecondaryColor),
                                              onPressed: () => setDialogState(() {
                                                searchCtrl.clear();
                                                searchResults.clear();
                                                showSearchResults = false;
                                              }),
                                            )
                                          : null,
                                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(8), borderSide: BorderSide.none),
                                      filled: true,
                                      fillColor: global.theme.cardColor,
                                      contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                                    ),
                                    style: TextStyle(fontSize: 13, color: global.theme.textColor),
                                    onSubmitted: searchLocation,
                                    onChanged: (v) {
                                      if (v.isEmpty) {
                                        setDialogState(() { searchResults.clear(); showSearchResults = false; });
                                      } else {
                                        setDialogState(() {});
                                      }
                                    },
                                  ),
                                ),
                                if (showSearchResults && searchResults.isNotEmpty)
                                  Material(
                                    elevation: 4,
                                    borderRadius: BorderRadius.circular(8),
                                    color: global.theme.cardColor,
                                    child: ConstrainedBox(
                                      constraints: const BoxConstraints(maxHeight: 200),
                                      child: ListView.separated(
                                        shrinkWrap: true,
                                        padding: EdgeInsets.zero,
                                        itemCount: searchResults.length,
                                        separatorBuilder: (context, i) => Divider(height: 1, color: global.theme.textSecondaryColor.withValues(alpha: 0.2)),
                                        itemBuilder: (_, i) {
                                          final r = searchResults[i];
                                          final lat = double.tryParse(r['lat']?.toString() ?? '') ?? 0;
                                          final lng = double.tryParse(r['lon']?.toString() ?? '') ?? 0;
                                          return ListTile(
                                            dense: true,
                                            leading: Icon(Icons.place_outlined, size: 18, color: global.theme.iconColor),
                                            title: Text(
                                              r['display_name']?.toString() ?? '',
                                              maxLines: 2,
                                              overflow: TextOverflow.ellipsis,
                                              style: TextStyle(fontSize: 12, color: global.theme.textColor),
                                            ),
                                            onTap: () {
                                              final pt = LatLng(lat, lng);
                                              setDialogState(() {
                                                selectedPoint = pt;
                                                latCtrl.text = lat.toStringAsFixed(6);
                                                lngCtrl.text = lng.toStringAsFixed(6);
                                                searchCtrl.clear();
                                                searchResults.clear();
                                                showSearchResults = false;
                                              });
                                              mapController.move(pt, 15);
                                            },
                                          );
                                        },
                                      ),
                                    ),
                                  ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                ),
                // Lat + Lng row
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 8, 16, 8),
                  child: Row(
                    children: [
                      Expanded(
                        child: TextFormField(
                          controller: latCtrl,
                          keyboardType: const TextInputType.numberWithOptions(decimal: true, signed: true),
                          decoration: InputDecoration(
                            labelText: global.language('company_latitude'),
                            border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                            contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                          ),
                          onChanged: (v) {
                            final lat = double.tryParse(v);
                            final lng = double.tryParse(lngCtrl.text);
                            if (lat != null && lng != null) {
                              final pt = LatLng(lat, lng);
                              setDialogState(() => selectedPoint = pt);
                              mapController.move(pt, 15);
                            }
                          },
                        ),
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: TextFormField(
                          controller: lngCtrl,
                          keyboardType: const TextInputType.numberWithOptions(decimal: true, signed: true),
                          decoration: InputDecoration(
                            labelText: global.language('company_longitude'),
                            border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                            contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                          ),
                          onChanged: (v) {
                            final lat = double.tryParse(latCtrl.text);
                            final lng = double.tryParse(v);
                            if (lat != null && lng != null) {
                              final pt = LatLng(lat, lng);
                              setDialogState(() => selectedPoint = pt);
                              mapController.move(pt, 15);
                            }
                          },
                        ),
                      ),
                    ],
                  ),
                ),
                // Actions
                const Divider(height: 1),
                Padding(
                  padding: const EdgeInsets.fromLTRB(16, 8, 16, 12),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.end,
                    children: [
                      TextButton(
                        onPressed: () => Navigator.pop(ctx),
                        child: Text(global.language('cancel')),
                      ),
                      const SizedBox(width: 8),
                      ElevatedButton(
                        onPressed: () {
                          widget.setState(() {
                            fc.shipToAddressController.text = addressCtrl.text;
                            fc.shipToLatController.text = latCtrl.text;
                            fc.shipToLngController.text = lngCtrl.text;
                            widget.screenData.prShipToAddress = addressCtrl.text;
                            widget.screenData.prShipToLat = latCtrl.text;
                            widget.screenData.prShipToLng = lngCtrl.text;
                          });
                          setState(() {});
                          Navigator.pop(ctx);
                        },
                        child: Text(global.language('save')),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          );
        },
      ),
    );

    mapController.dispose();
    addressCtrl.dispose();
    latCtrl.dispose();
    lngCtrl.dispose();
  }

  // ──────────────────────────────────────────────────────
  // ประมาณการต้นทุน (Estimated Landed Cost)
  // ──────────────────────────────────────────────────────
  // ──────────────────────────────────────────────────────
  // สกุลเงิน (Currency Section)
  // ──────────────────────────────────────────────────────
  Widget _buildCurrencySection() {
    final currencies = widget.currencies ?? [];
    final selectedCode = widget.selectedDocCurrencyCode;
    final baseCur = widget.baseCurrency ?? 'THB';
    final baseSym = widget.baseCurrencySymbol ?? '฿';
    final rate = widget.selectedExchangeRate ?? 1.0;
    final isDiffCurrency = selectedCode != null && selectedCode != baseCur;

    return Column(
      children: [
        Row(
          children: [
            // Base Currency (read-only)
            Expanded(
              child: InputDecorator(
                decoration: InputDecoration(
                  labelText: global.language('base_currency'),
                  prefixIcon: Icon(Icons.account_balance, size: 20, color: global.theme.iconColor),
                  prefixIconConstraints: const BoxConstraints(minWidth: 36, minHeight: 0),
                  border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                  contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                ),
                child: Text('$baseCur ($baseSym)', style: TextStyle(fontSize: 14, color: global.theme.textColor)),
              ),
            ),
            const SizedBox(width: 12),
            // Document Currency (dropdown)
            Expanded(
              flex: 2,
              child: DropdownButtonFormField<String>(
                initialValue: selectedCode,
                decoration: InputDecoration(
                  labelText: global.language('doc_currency'),
                  prefixIcon: Icon(Icons.currency_exchange, size: 20, color: global.theme.primaryColor),
                  prefixIconConstraints: const BoxConstraints(minWidth: 36, minHeight: 0),
                  border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                  contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                ),
                items: currencies.map((c) => DropdownMenuItem(
                  value: c.code,
                  child: Text('${c.code} (${c.symbol}) - ${c.name}', style: const TextStyle(fontSize: 13)),
                )).toList(),
                onChanged: (code) {
                  if (code == null) return;
                  final cur = currencies.firstWhere((c) => c.code == code);
                  final latestRate = cur.exchangeRates.isNotEmpty ? cur.exchangeRates.first.rate : 1.0;
                  widget.onCurrencyChanged?.call(code, cur.symbol, latestRate);
                },
              ),
            ),
            const SizedBox(width: 12),
            // Exchange Rate
            Expanded(
              child: GestureDetector(
                onTap: isDiffCurrency ? () async {
                  final result = await _showExchangeRateDialog(rate);
                  if (result != null) {
                    widget.onExchangeRateChanged?.call(result);
                  }
                } : null,
                child: AbsorbPointer(
                  child: InputDecorator(
                    decoration: InputDecoration(
                      labelText: global.language('exchange_rate'),
                      prefixIcon: Icon(Icons.sync_alt, size: 20, color: isDiffCurrency ? global.theme.positiveHighlightTextColor : global.theme.iconSecondaryColor),
                      prefixIconConstraints: const BoxConstraints(minWidth: 36, minHeight: 0),
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                      contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                    ),
                    child: Text(
                      isDiffCurrency ? '1 $selectedCode = ${rate.toStringAsFixed(4)} $baseCur' : '1.0000',
                      style: TextStyle(fontSize: 13, color: isDiffCurrency ? global.theme.positiveHighlightTextColor : global.theme.textSecondaryColor),
                    ),
                  ),
                ),
              ),
            ),
          ],
        ),
      ],
    );
  }

  /// Dialog แก้ไขอัตราแลกเปลี่ยน
  Future<double?> _showExchangeRateDialog(double currentRate) async {
    final controller = TextEditingController(text: currentRate.toStringAsFixed(4));
    return showDialog<double>(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(global.language('exchange_rate'), style: const TextStyle(fontSize: 15)),
        content: SizedBox(
          width: 280,
          child: TextField(
            controller: controller,
            autofocus: true,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            decoration: InputDecoration(
              labelText: global.language('exchange_rate'),
              border: const OutlineInputBorder(),
              isDense: true,
            ),
            onSubmitted: (v) => Navigator.pop(ctx, double.tryParse(v)),
          ),
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: Text(global.language('cancel'))),
          FilledButton(onPressed: () => Navigator.pop(ctx, double.tryParse(controller.text)), child: Text(global.language('save'))),
        ],
      ),
    );
  }

  // ──────────────────────────────────────────────────────
  // ร้านที่ต้องการ (Dynamic List)
  // ──────────────────────────────────────────────────────
  Widget _buildVendorPreferencesList() {
    return Column(
      children: [
        ...List.generate(_vendorPreferences.length, (i) {
          final vp = _vendorPreferences[i];
          return Padding(
            padding: EdgeInsets.only(bottom: i < _vendorPreferences.length - 1 ? 8 : 0),
            child: Row(
              children: [
                // ลำดับ
                Container(
                  width: 24, height: 24,
                  alignment: Alignment.center,
                  decoration: BoxDecoration(
                    color: global.theme.primaryColor.withValues(alpha: 0.1),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: Text('${i + 1}', style: TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: global.theme.primaryColor)),
                ),
                const SizedBox(width: 8),
                // ชื่อร้าน
                Expanded(
                  flex: 2,
                  child: TextFormField(
                    controller: vp['vendor'],
                    decoration: InputDecoration(
                      labelText: i == 0 ? global.language('preferred_vendor') : global.language('alternative_vendor'),
                      prefixIcon: Icon(i == 0 ? Icons.star_outline : Icons.store_outlined, size: 18, color: global.theme.iconColor),
                      prefixIconConstraints: const BoxConstraints(minWidth: 36, minHeight: 0),
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                      contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                      isDense: true,
                    ),
                    style: const TextStyle(fontSize: 13),
                  ),
                ),
                const SizedBox(width: 8),
                // เหตุผล
                Expanded(
                  flex: 3,
                  child: TextFormField(
                    controller: vp['reason'],
                    decoration: InputDecoration(
                      labelText: global.language('reason_preferred'),
                      prefixIcon: Icon(Icons.info_outline, size: 18, color: global.theme.iconColor),
                      prefixIconConstraints: const BoxConstraints(minWidth: 36, minHeight: 0),
                      border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                      contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                      isDense: true,
                    ),
                    style: const TextStyle(fontSize: 13),
                  ),
                ),
                const SizedBox(width: 4),
                // ปุ่มลบ
                if (_vendorPreferences.length > 1)
                  IconButton(
                    icon: Icon(Icons.remove_circle_outline, size: 20, color: global.theme.negativeHighlightTextColor),
                    padding: EdgeInsets.zero,
                    constraints: const BoxConstraints(minWidth: 28, minHeight: 28),
                    onPressed: () => _removeVendorPreference(i),
                    tooltip: global.language('delete'),
                  )
                else
                  const SizedBox(width: 28),
              ],
            ),
          );
        }),
        const SizedBox(height: 6),
        // ปุ่มเพิ่มร้าน
        Align(
          alignment: Alignment.centerLeft,
          child: TextButton.icon(
            onPressed: _addVendorPreference,
            icon: Icon(Icons.add, size: 16),
            label: Text(global.language('add_vendor'), style: const TextStyle(fontSize: 12)),
            style: TextButton.styleFrom(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              foregroundColor: global.theme.primaryColor,
            ),
          ),
        ),
      ],
    );
  }

  // ──────────────────────────────────────────────────────
  // Tracking (read-only) — สถานะแปลง + Ref RFQ/PO
  // ──────────────────────────────────────────────────────
  bool _hasTrackingData() {
    final sd = widget.screenData;
    return (sd.prConversionStatus != null && sd.prConversionStatus!.isNotEmpty && sd.prConversionStatus != 'none') ||
        (sd.prRefRfqDocNo != null && sd.prRefRfqDocNo!.isNotEmpty) ||
        (sd.prRefPoDocNo != null && sd.prRefPoDocNo!.isNotEmpty);
  }

  Widget _buildTrackingRow() {
    final sd = widget.screenData;
    return Row(
      children: [
        Expanded(
          child: _buildReadOnlyField(
            label: global.language('conversion_status'),
            value: sd.prConversionStatus ?? 'none',
            icon: Icons.sync_alt,
          ),
        ),
        if (sd.prRefRfqDocNo != null && sd.prRefRfqDocNo!.isNotEmpty) ...[
          const SizedBox(width: 12),
          Expanded(
            child: _buildReadOnlyField(
              label: global.language('ref_rfq_docno'),
              value: sd.prRefRfqDocNo!,
              icon: Icons.request_quote_outlined,
            ),
          ),
        ],
        if (sd.prRefPoDocNo != null && sd.prRefPoDocNo!.isNotEmpty) ...[
          const SizedBox(width: 12),
          Expanded(
            child: _buildReadOnlyField(
              label: global.language('ref_po_docno'),
              value: sd.prRefPoDocNo!,
              icon: Icons.shopping_cart_outlined,
            ),
          ),
        ],
      ],
    );
  }

  // ──────────────────────────────────────────────────────
  // Helpers
  // ──────────────────────────────────────────────────────
  Widget _buildReadOnlyField({
    required String label,
    required String value,
    IconData? icon,
  }) {
    return TextFormField(
      initialValue: value,
      readOnly: true,
      enabled: false,
      decoration: InputDecoration(
        labelText: label,
        prefixIcon: icon != null ? Icon(icon, size: 18, color: global.theme.iconColor) : null,
        prefixIconConstraints: const BoxConstraints(minWidth: 36, minHeight: 0),
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
        disabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
          borderSide: BorderSide(color: global.theme.dividerBorderColor),
        ),
        contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
        isDense: true,
        filled: true,
        fillColor: global.theme.cardColor,
      ),
      style: TextStyle(fontSize: 13, color: global.theme.textSecondaryColor),
    );
  }
}

/// Dropdown overlay สำหรับเลือกแผนก — แสดงใต้ textbox พร้อมค้นหา
class _DeptDropdownOverlay extends StatefulWidget {
  final LayerLink link;
  final List<dynamic> departments;
  final TextEditingController searchCtrl;
  final void Function(dynamic item) onSelect;
  final VoidCallback onClose;

  const _DeptDropdownOverlay({
    required this.link,
    required this.departments,
    required this.searchCtrl,
    required this.onSelect,
    required this.onClose,
  });

  @override
  State<_DeptDropdownOverlay> createState() => _DeptDropdownOverlayState();
}

class _DeptDropdownOverlayState extends State<_DeptDropdownOverlay> {
  String _search = '';

  @override
  Widget build(BuildContext context) {
    final filtered = widget.departments.where((e) {
      final name = global.activeLangName(e.names);
      return _search.isEmpty ||
          name.toLowerCase().contains(_search.toLowerCase()) ||
          e.code.toLowerCase().contains(_search.toLowerCase());
    }).toList();

    return Stack(
      children: [
        // กด backdrop ปิด overlay
        Positioned.fill(
          child: GestureDetector(
            behavior: HitTestBehavior.opaque,
            onTap: widget.onClose,
            child: const SizedBox.expand(),
          ),
        ),
        CompositedTransformFollower(
          link: widget.link,
          targetAnchor: Alignment.bottomLeft,
          followerAnchor: Alignment.topLeft,
          child: Material(
            elevation: 4,
            borderRadius: BorderRadius.circular(8),
            color: global.theme.cardColor,
            child: Container(
              width: 320,
              constraints: const BoxConstraints(maxHeight: 280),
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: global.theme.inputTextBoxColor),
              ),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  // Search field
                  Padding(
                    padding: const EdgeInsets.fromLTRB(8, 8, 8, 4),
                    child: TextField(
                      controller: widget.searchCtrl,
                      autofocus: true,
                      decoration: InputDecoration(
                        hintText: global.language('search'),
                        prefixIcon: const Icon(Icons.search, size: 16),
                        contentPadding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                        border: OutlineInputBorder(borderRadius: BorderRadius.circular(6)),
                        isDense: true,
                      ),
                      style: const TextStyle(fontSize: 13),
                      onChanged: (v) => setState(() => _search = v),
                    ),
                  ),
                  const Divider(height: 1),
                  // List
                  Flexible(
                    child: filtered.isEmpty
                        ? Padding(
                            padding: const EdgeInsets.all(16),
                            child: Text(global.language('no_data'),
                                style: TextStyle(fontSize: 13, color: global.theme.textSecondaryColor)),
                          )
                        : ListView.builder(
                            shrinkWrap: true,
                            padding: EdgeInsets.zero,
                            itemCount: filtered.length,
                            itemBuilder: (_, i) {
                              final item = filtered[i];
                              final name = global.activeLangName(item.names);
                              return InkWell(
                                onTap: () => widget.onSelect(item),
                                child: Padding(
                                  padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                                  child: Row(
                                    children: [
                                      Icon(Icons.business_outlined, size: 16, color: global.theme.iconColor),
                                      const SizedBox(width: 8),
                                      Expanded(
                                        child: Text(
                                          name.isNotEmpty ? '${item.code} / $name' : item.code,
                                          style: const TextStyle(fontSize: 13),
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                              );
                            },
                          ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ],
    );
  }
}
