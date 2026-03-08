// ignore_for_file: deprecated_member_use

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/imports_bloc.dart';
import 'package:smlaicloud/model/bi_report/bi_report_models.dart';
import 'package:smlaicloud/repositories/creditor_repository.dart';
import 'package:smlaicloud/repositories/debtor_repository.dart';
import 'package:smlaicloud/repositories/employee_repository.dart';
import 'package:smlaicloud/model/bi_report/branch_selection_model.dart';
import 'package:smlaicloud/model/bi_report/entity_selection_model.dart';
import 'package:smlaicloud/screen_search/dedebi/multi_branch_search_screen.dart';
import 'package:smlaicloud/screen_search/dedebi/multi_entity_search_screen.dart';
import 'package:smlaicloud/utils/date_picker.dart';

class ReportConditions {
  final DateTime? fromDate;
  final DateTime? toDate;
  final bool? showDetails;
  final String? showCancelled;
  final String? saleType;
  final String? posType;
  final BranchSelectionModel? selectedBranches;
  final EntitySelectionModel? selectedCreditors;
  final EntitySelectionModel? selectedSalespersons;
  final EntitySelectionModel? selectedBarcodes;

  const ReportConditions({
    this.fromDate,
    this.toDate,
    this.showDetails,
    this.showCancelled,
    this.saleType,
    this.posType,
    this.selectedBranches,
    this.selectedCreditors,
    this.selectedSalespersons,
    this.selectedBarcodes,
  });
}

class ReportConditionDialog extends StatefulWidget {
  final DateTime? initialFromDate;
  final DateTime? initialToDate;
  final bool? initialShowDetails;
  final String? initialShowCancelled;
  final String? initialSaleType;
  final String? initialPosType;
  final BranchSelectionModel? initialSelectedBranches;
  final EntitySelectionModel? initialSelectedCreditors;
  final EntitySelectionModel? initialSelectedSalespersons;
  final EntitySelectionModel? initialSelectedBarcodes;
  final Function(ReportConditions) onConditionsSet;
  final BiReportType reportType;

  ReportConditionDialog({
    super.key,
    this.initialFromDate,
    this.initialToDate,
    this.initialShowDetails = true,
    String? initialShowCancelled,
    String? initialSaleType,
    String? initialPosType,
    this.initialSelectedBranches = const BranchSelectionModel(
      selectedBranches: [],
      isCancel: false,
    ),
    this.initialSelectedCreditors = const EntitySelectionModel(
      selectedEntities: [],
      isCancel: false,
    ),
    this.initialSelectedSalespersons = const EntitySelectionModel(
      selectedEntities: [],
      isCancel: false,
    ),
    this.initialSelectedBarcodes = const EntitySelectionModel(
      selectedEntities: [],
      isCancel: false,
    ),
    required this.onConditionsSet,
    required this.reportType,
  })  : initialShowCancelled = initialShowCancelled ?? global.language('all'),
        initialSaleType = initialSaleType ?? global.language('all'),
        initialPosType = initialPosType ?? global.language('all');

  @override
  State<ReportConditionDialog> createState() => _ReportConditionDialogState();
}

class _ReportConditionDialogState extends State<ReportConditionDialog> {
  late DateTime? _fromDate;
  late DateTime? _toDate;
  late bool _showDetails;
  late String _showCancelled;
  late String _saleType;
  late String _posType;
  late BranchSelectionModel _selectedBranches;
  late EntitySelectionModel _selectedCreditors;
  late EntitySelectionModel _selectedSalespersons;
  late EntitySelectionModel _selectedBarcodes;

  @override
  void initState() {
    super.initState();
    _fromDate = widget.initialFromDate;
    _toDate = widget.initialToDate;
    _showDetails = widget.initialShowDetails!;
    _showCancelled = widget.initialShowCancelled!;
    _saleType = widget.initialSaleType!;
    _posType = widget.initialPosType!;
    _selectedBranches = widget.initialSelectedBranches!;
    _selectedCreditors = widget.initialSelectedCreditors!;
    _selectedSalespersons = widget.initialSelectedSalespersons!;
    _selectedBarcodes = widget.initialSelectedBarcodes!;
  }

  @override
  Widget build(BuildContext context) {
    final screenWidth = MediaQuery.of(context).size.width;
    final screenHeight = MediaQuery.of(context).size.height;

    // กำหนดความกว้างตามหน้าจอ
    double dialogWidth;
    double maxDialogHeight = screenHeight * 0.9; // ความสูงสูงสุด 90% ของหน้าจอ

    if (screenWidth >= 1300) {
      dialogWidth = screenWidth * 0.5;
    } else if (screenWidth > 800) {
      dialogWidth = screenWidth * 0.6;
    } else {
      dialogWidth = screenWidth * 0.95;
    }

    return Dialog(
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(16),
      ), // ลดจาก 20 เป็น 16
      child: Container(
        width: dialogWidth,
        constraints: BoxConstraints(
          maxHeight: maxDialogHeight, // จำกัดความสูงสูงสุด
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min, // ใช้พื้นที่น้อยที่สุด
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // Header (ไม่ scroll)
            Padding(
              padding: const EdgeInsets.all(16), // ลดจาก 24 เป็น 16
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  _buildHeader(),
                  const SizedBox(height: 12), // ลดจาก 16 เป็น 12
                  const Divider(),
                ],
              ),
            ),
            // Content ที่สามารถ scroll ได้
            Flexible(
              // ใช้ Flexible เพื่อให้ปรับขนาดตามเนื้อหา
              child: SingleChildScrollView(
                padding: const EdgeInsets.symmetric(
                  horizontal: 16,
                ), // ลดจาก 24 เป็น 16
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    _buildDateRangeSection(),
                    if (widget.reportType == BiReportType.paymentDaily ||
                        widget.reportType == BiReportType.saleReturn) ...[
                      const SizedBox(height: 12),
                      _buildBranchSelectionSection(),
                      const SizedBox(height: 12),
                    ] else if (widget.reportType ==
                            BiReportType.stockMovement ||
                        widget.reportType == BiReportType.stockBalance) ...[
                      const SizedBox(height: 12),
                      _buildBarcodeSection(),
                      const SizedBox(height: 12),
                    ] else if (widget.reportType ==
                        BiReportType.grossProfitByDocument) ...[
                      // Gross Profit By Document: วันที่ + สาขา + เจ้าหนี้ + เอกสารยกเลิก
                      const SizedBox(height: 12),
                      _buildBranchSelectionAndCreditorSection(),
                      const SizedBox(height: 12),
                      _buildShowCancelledToggle(),
                      const SizedBox(height: 16),
                    ] else if (widget.reportType ==
                        BiReportType.grossProfitByProduct) ...[
                      // Gross Profit By Product: วันที่ + บาร์โค้ด + สาขา
                      const SizedBox(height: 12),
                      _buildBarcodeSection(),
                      const SizedBox(height: 12),
                      _buildBranchSelectionSection(),
                      const SizedBox(height: 16),
                    ] else if (widget.reportType == BiReportType.vatSale) ...[
                      // VAT Sale: วันที่ + สาขา
                      const SizedBox(height: 12),
                      _buildBranchSelectionSection(),
                      const SizedBox(height: 16),
                    ] else if (widget.reportType == BiReportType.vatBuy) ...[
                      // VAT Buy: วันที่ + สาขา
                      const SizedBox(height: 12),
                      _buildBranchSelectionSection(),
                      const SizedBox(height: 16),
                    ] else if (widget.reportType !=
                            BiReportType.stockMovement &&
                        widget.reportType != BiReportType.stockBalance) ...[
                      const SizedBox(height: 12),
                      if (widget.reportType ==
                          BiReportType.purchasepartial) ...[
                        _buildBranchSelectionAndCreditorSection(),
                      ] else ...[
                        _buildBranchSelectionSection(),
                      ],
                      const SizedBox(height: 12),
                      if (widget.reportType == BiReportType.sale) ...[
                        _buildCreditorAndSalespersonSection(),
                        const SizedBox(height: 12),
                      ],
                      _buildFilterOptionsSection(),
                      const SizedBox(height: 16),
                    ],
                  ],
                ),
              ),
            ),
            // Footer (ปุ่มแอ็คชัน - ไม่ scroll)
            Container(
              padding: const EdgeInsets.all(16), // ลดจาก 24 เป็น 16
              decoration: BoxDecoration(
                color: Colors.white,
                border: Border(top: BorderSide(color: Colors.grey.shade200)),
              ),
              child: _buildActionButtons(),
            ),
          ],
        ), // ปิด Column
      ), // ปิด Container
    ); // ปิด Dialog
  }

  Widget _buildHeader() {
    return Row(
      children: [
        Container(
          padding: const EdgeInsets.all(8),
          decoration: BoxDecoration(
            color: Colors.indigo.shade100,
            borderRadius: BorderRadius.circular(10),
          ),
          child: Icon(
            Icons.tune,
            color: Colors.indigo.shade700,
            size: 20, // ลดจาก 24 เป็น 20
          ),
        ),
        SizedBox(width: 12), // ลดจาก 16 เป็น 12
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                widget.reportType.displayName,
                style: const TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                  color: Colors.indigo,
                ),
              ),
              Text(
                global.language('report_condition_specify_date'),
                style: const TextStyle(fontSize: 14, color: Colors.grey),
              ),
            ],
          ),
        ),
        IconButton(
          onPressed: () => Navigator.of(context).pop(),
          icon: Icon(Icons.close),
          tooltip: global.language('report_condition_close'),
        ),
      ],
    );
  }

  Widget _buildDateRangeSection() {
    return Container(
      padding: const EdgeInsets.all(16), // ลดจาก 20 เป็น 16
      decoration: BoxDecoration(
        color: Colors.grey.shade50,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.date_range, color: Colors.indigo.shade600, size: 18),
              SizedBox(width: 6),
              Text(
                (widget.reportType == BiReportType.stockBalance)
                    ? global.language('report_condition_as_of_date')
                    : global.language('report_condition_date_range'),
                style: TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.w600,
                  color: Colors.indigo,
                ),
              ),
            ],
          ),
          SizedBox(height: 12), // ลดจาก 16 เป็น 12
          Row(
            children: [
              if (widget.reportType != BiReportType.stockBalance) ...[
                Expanded(
                  child: CustomDatePicker(
                    decoration: InputDecoration(
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(6),
                      ),
                      suffixIcon: const Icon(Icons.calendar_today),
                      filled: true,
                      fillColor: Colors.white,
                    ),
                    initialDate: _fromDate,
                    // useBuddhistCalendar และ languageCode จะอ่านจากสาขาอัตโนมัติ
                    onDateSelected: (date) {
                      setState(() {
                        _fromDate = date;
                      });
                    },
                    labelText: global.language('report_condition_start_date'),
                  ),
                ),
              ],
              SizedBox(width: 12), // ลดจาก 16 เป็น 12
              Expanded(
                child: CustomDatePicker(
                  decoration: InputDecoration(
                    labelText: (widget.reportType == BiReportType.stockBalance)
                        ? global.language('report_condition_as_of_date')
                        : global.language('report_condition_end_date'),
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(6),
                    ),
                    suffixIcon: const Icon(Icons.calendar_today),
                    filled: true,
                    fillColor: Colors.white,
                  ),
                  initialDate: _toDate,
                  // useBuddhistCalendar และ languageCode จะอ่านจากสาขาอัตโนมัติ
                  onDateSelected: (date) {
                    setState(() {
                      _toDate = date;
                    });
                  },
                  labelText: (widget.reportType == BiReportType.stockBalance)
                      ? global.language('report_condition_as_of_date')
                      : global.language('report_condition_end_date'),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  // _buildBarcodeSection
  Widget _buildBarcodeSection() {
    return Container(
      padding: const EdgeInsets.all(12), // ลดจาก 16 เป็น 12
      decoration: BoxDecoration(
        color: Colors.purple.shade50,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: Colors.purple.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.qr_code, color: Colors.purple.shade600, size: 18),
              SizedBox(width: 6),
              Text(
                global.language('report_condition_select_barcode'),
                style: const TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.w600,
                  color: Colors.purple,
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          InkWell(
            onTap: () async {
              final result = await Navigator.push<EntitySelectionModel?>(
                context,
                MaterialPageRoute(
                  builder: (context) => MultiEntitySearchScreen(
                    word: '',
                    entityType: EntityType.barcode,
                    preSelectedEntities: _selectedBarcodes.selectedEntities,
                  ),
                ),
              );

              if (result != null && !result.isCancel) {
                setState(() {
                  _selectedBarcodes = result;
                });
              }
            },
            child: Container(
              width: double.infinity,
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(6),
                border: Border.all(color: Colors.purple.shade300),
              ),
              child: Row(
                children: [
                  Expanded(
                    child: Text(
                      _selectedBarcodes.selectedEntities.isEmpty
                          ? global.language(
                              'report_condition_click_select_barcode',
                            )
                          : '${_selectedBarcodes.selectedEntities.first.code} : ${_selectedBarcodes.selectedEntities.first.getDisplayName()}',
                      style: TextStyle(
                        color: _selectedBarcodes.selectedEntities.isEmpty
                            ? Colors.grey.shade600
                            : Colors.black,
                        fontSize: 14,
                      ),
                    ),
                  ),
                  Icon(
                    Icons.arrow_forward_ios,
                    size: 16,
                    color: Colors.purple.shade600,
                  ),
                ],
              ),
            ),
          ),
          if (_selectedBarcodes.selectedEntities.isNotEmpty) ...[
            SizedBox(height: 6),
            Row(
              children: [
                Icon(
                  Icons.info_outline,
                  size: 14, // ลดจาก 16 เป็น 14
                  color: Colors.purple.shade600,
                ),
                SizedBox(width: 3), // ลดจาก 4 เป็น 3
                Text(
                  _selectedBarcodes.selectedEntities.isEmpty
                      ? global.language('report_condition_not_selected_barcode')
                      : '${global.language('report_condition_selected')}: ${_selectedBarcodes.selectedEntities.first.code} : ${_selectedBarcodes.selectedEntities.first.getDisplayName()}',
                  style: TextStyle(fontSize: 14, color: Colors.purple.shade700),
                ),
                Spacer(),
                InkWell(
                  onTap: () {
                    setState(() {
                      _selectedBarcodes = const EntitySelectionModel(
                        selectedEntities: [],
                        isCancel: false,
                      );
                    });
                  },
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 6,
                      vertical: 3,
                    ), // ลดจาก 8,4 เป็น 6,3
                    decoration: BoxDecoration(
                      color: Colors.purple.shade100,
                      borderRadius: BorderRadius.circular(3), // ลดจาก 4 เป็น 3
                    ),
                    child: Text(
                      global.language('report_condition_delete_all'),
                      style: TextStyle(
                        fontSize: 10, // ลดจาก 11 เป็น 10
                        color: Colors.purple.shade700,
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildBranchSelectionAndCreditorSection() {
    return Row(
      children: [
        Expanded(child: _buildBranchSelectionSection()),
        const SizedBox(width: 12),
        Expanded(child: _buildCreditorSelectionSection()),
      ],
    );
  }

  Widget _buildBranchSelectionSection() {
    return Container(
      padding: const EdgeInsets.all(12), // ลดจาก 16 เป็น 12
      decoration: BoxDecoration(
        color: Colors.orange.shade50,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: Colors.orange.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.business, color: Colors.orange.shade700, size: 18),
              SizedBox(width: 6),
              Text(
                global.language('report_condition_select_branch'),
                style: const TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.w600,
                  color: Colors.orange,
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          InkWell(
            onTap: () async {
              final result = await Navigator.push<BranchSelectionModel?>(
                context,
                MaterialPageRoute(
                  builder: (context) => MultiBranchSearchScreen(
                    word: '',
                    preSelectedBranches: _selectedBranches.selectedBranches,
                  ),
                ),
              );

              if (result != null && !result.isCancel) {
                setState(() {
                  _selectedBranches = result;
                });
              }
            },
            child: Container(
              width: double.infinity,
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(6),
                border: Border.all(color: Colors.orange.shade300),
              ),
              child: Row(
                children: [
                  Expanded(
                    child: Text(
                      _selectedBranches.selectedBranches.isEmpty
                          ? global.language(
                              'report_condition_click_select_branch',
                            )
                          : _selectedBranches.getBranchDisplayString(),
                      style: TextStyle(
                        color: _selectedBranches.selectedBranches.isEmpty
                            ? Colors.grey.shade600
                            : Colors.black,
                        fontSize: 14,
                      ),
                    ),
                  ),
                  Icon(
                    Icons.arrow_forward_ios,
                    size: 16,
                    color: Colors.orange.shade600,
                  ),
                ],
              ),
            ),
          ),
          if (_selectedBranches.selectedBranches.isNotEmpty) ...[
            SizedBox(height: 6),
            Row(
              children: [
                Icon(
                  Icons.info_outline,
                  size: 14, // ลดจาก 16 เป็น 14
                  color: Colors.orange.shade600,
                ),
                SizedBox(width: 3), // ลดจาก 4 เป็น 3
                Text(
                  '${global.language('report_condition_selected')} ${_selectedBranches.selectedBranches.length} ${global.language('report_condition_branch')}',
                  style: TextStyle(fontSize: 14, color: Colors.orange.shade700),
                ),
                Spacer(),
                InkWell(
                  onTap: () {
                    setState(() {
                      _selectedBranches = const BranchSelectionModel(
                        selectedBranches: [],
                        isCancel: false,
                      );
                    });
                  },
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 6,
                      vertical: 3,
                    ), // ลดจาก 8,4 เป็น 6,3
                    decoration: BoxDecoration(
                      color: Colors.orange.shade100,
                      borderRadius: BorderRadius.circular(3), // ลดจาก 4 เป็น 3
                    ),
                    child: Text(
                      global.language('report_condition_delete_all'),
                      style: TextStyle(
                        fontSize: 10, // ลดจาก 11 เป็น 10
                        color: Colors.orange.shade700,
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildCreditorAndSalespersonSection() {
    return Row(
      children: [
        Expanded(child: _buildCreditorSelectionSection()),
        const SizedBox(width: 12), // ลดจาก 16 เป็น 12
        Expanded(child: _buildSalespersonSelectionSection()),
      ],
    );
  }

  Widget _buildCreditorSelectionSection() {
    // ตรวจสอบประเภทรายงานเพื่อแสดงข้อความและใช้ EntityType ที่เหมาะสม
    final bool isCreditorReport =
        widget.reportType == BiReportType.purchasepartial ||
        widget.reportType == BiReportType.grossProfitByDocument;
    final String entityLabel = isCreditorReport ? global.language('creditor') : global.language('customer');
    final EntityType entityType = isCreditorReport
        ? EntityType.creditor
        : EntityType.debtor;
    final MaterialColor sectionColor = isCreditorReport
        ? Colors.purple
        : Colors.blue;
    final IconData sectionIcon = isCreditorReport
        ? Icons.business
        : Icons.people;

    return Container(
      padding: const EdgeInsets.all(12), // ลดจาก 16 เป็น 12
      decoration: BoxDecoration(
        color: sectionColor.shade50,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: sectionColor.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(sectionIcon, color: sectionColor.shade700, size: 18),
              const SizedBox(width: 6),
              Text(
                'เลือก$entityLabel',
                style: TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.w600,
                  color: sectionColor,
                ),
              ),
            ],
          ),
          const SizedBox(height: 6),
          InkWell(
            onTap: () async {
              final result = await Navigator.push<EntitySelectionModel?>(
                context,
                MaterialPageRoute(
                  builder: (context) => MultiBlocProvider(
                    providers: isCreditorReport
                        ? [
                            BlocProvider<CreditorBloc>(
                              create: (context) => CreditorBloc(
                                creditorRepository: CreditorRepository(),
                              ),
                            ),
                          ]
                        : [
                            BlocProvider<DebtorBloc>(
                              create: (context) => DebtorBloc(
                                debtorRepository: DebtorRepository(),
                              ),
                            ),
                          ],
                    child: MultiEntitySearchScreen(
                      word: '',
                      entityType: entityType,
                      preSelectedEntities: _selectedCreditors.selectedEntities,
                    ),
                  ),
                ),
              );

              if (result != null && !result.isCancel) {
                setState(() {
                  _selectedCreditors = result;
                });
              }
            },
            child: Container(
              width: double.infinity,
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(6),
                border: Border.all(color: sectionColor.shade300),
              ),
              child: Row(
                children: [
                  Expanded(
                    child: Text(
                      _selectedCreditors.selectedEntities.isEmpty
                          ? 'กดเพื่อเลือก$entityLabel (ทั้งหมด)'
                          : _selectedCreditors.getEntityDisplayString(),
                      style: TextStyle(
                        color: _selectedCreditors.selectedEntities.isEmpty
                            ? Colors.grey.shade600
                            : Colors.black,
                        fontSize: 14,
                      ),
                    ),
                  ),
                  Icon(
                    Icons.arrow_forward_ios,
                    size: 16,
                    color: sectionColor.shade600,
                  ),
                ],
              ),
            ),
          ),
          if (_selectedCreditors.selectedEntities.isNotEmpty) ...[
            const SizedBox(height: 6),
            Row(
              children: [
                Icon(
                  Icons.info_outline,
                  size: 14, // ลดจาก 16 เป็น 14
                  color: sectionColor.shade600,
                ),
                const SizedBox(width: 3), // ลดจาก 4 เป็น 3
                Text(
                  'เลือกแล้ว ${_selectedCreditors.selectedEntities.length} $entityLabel',
                  style: TextStyle(fontSize: 14, color: sectionColor.shade700),
                ),
                Spacer(),
                InkWell(
                  onTap: () {
                    setState(() {
                      _selectedCreditors = const EntitySelectionModel(
                        selectedEntities: [],
                        isCancel: false,
                      );
                    });
                  },
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 6,
                      vertical: 3,
                    ), // ลดจาก 8,4 เป็น 6,3
                    decoration: BoxDecoration(
                      color: sectionColor.shade100,
                      borderRadius: BorderRadius.circular(3), // ลดจาก 4 เป็น 3
                    ),
                    child: Text(
                      global.language('report_condition_delete_all'),
                      style: TextStyle(
                        fontSize: 10, // ลดจาก 11 เป็น 10
                        color: sectionColor.shade700,
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildSalespersonSelectionSection() {
    return Container(
      padding: const EdgeInsets.all(12), // ลดจาก 16 เป็น 12
      decoration: BoxDecoration(
        color: Colors.green.shade50,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: Colors.green.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.person_pin, color: Colors.green.shade700, size: 18),
              SizedBox(width: 6),
              Text(
                global.language('select_salesperson'),
                style: TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.w600,
                  color: Colors.green,
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          InkWell(
            onTap: () async {
              final result = await Navigator.push<EntitySelectionModel?>(
                context,
                MaterialPageRoute(
                  builder: (context) => MultiBlocProvider(
                    providers: [
                      BlocProvider<EmployeeBloc>(
                        create: (context) => EmployeeBloc(
                          employeeRepository: EmployeeRepository(),
                        ),
                      ),
                    ],
                    child: MultiEntitySearchScreen(
                      word: '',
                      entityType: EntityType.employee,
                      preSelectedEntities:
                          _selectedSalespersons.selectedEntities,
                    ),
                  ),
                ),
              );

              if (result != null && !result.isCancel) {
                setState(() {
                  _selectedSalespersons = result;
                });
              }
            },
            child: Container(
              width: double.infinity,
              padding: const EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(6),
                border: Border.all(color: Colors.green.shade300),
              ),
              child: Row(
                children: [
                  Expanded(
                    child: Text(
                      _selectedSalespersons.selectedEntities.isEmpty
                          ? 'กดเพื่อเลือกพนักงานขาย (ทั้งหมด)'
                          : _selectedSalespersons.getEntityDisplayString(),
                      style: TextStyle(
                        color: _selectedSalespersons.selectedEntities.isEmpty
                            ? Colors.grey.shade600
                            : Colors.black,
                        fontSize: 14,
                      ),
                    ),
                  ),
                  Icon(
                    Icons.arrow_forward_ios,
                    size: 16,
                    color: Colors.green.shade600,
                  ),
                ],
              ),
            ),
          ),
          if (_selectedSalespersons.selectedEntities.isNotEmpty) ...[
            const SizedBox(height: 6),
            Row(
              children: [
                Icon(
                  Icons.info_outline,
                  size: 14, // ลดจาก 16 เป็น 14
                  color: Colors.green.shade600,
                ),
                const SizedBox(width: 3), // ลดจาก 4 เป็น 3
                Text(
                  'เลือกแล้ว ${_selectedSalespersons.selectedEntities.length} พนักงาน',
                  style: TextStyle(fontSize: 14, color: Colors.green.shade700),
                ),
                Spacer(),
                InkWell(
                  onTap: () {
                    setState(() {
                      _selectedSalespersons = const EntitySelectionModel(
                        selectedEntities: [],
                        isCancel: false,
                      );
                    });
                  },
                  child: Container(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 6,
                      vertical: 3,
                    ), // ลดจาก 8,4 เป็น 6,3
                    decoration: BoxDecoration(
                      color: Colors.green.shade100,
                      borderRadius: BorderRadius.circular(3), // ลดจาก 4 เป็น 3
                    ),
                    child: Text(
                      global.language('report_condition_delete_all'),
                      style: TextStyle(
                        fontSize: 10, // ลดจาก 11 เป็น 10
                        color: Colors.green.shade700,
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildFilterOptionsSection() {
    return Container(
      padding: const EdgeInsets.all(16), // ลดจาก 20 เป็น 16
      decoration: BoxDecoration(
        color: Colors.green.shade50,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: Colors.green.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.filter_list, color: Colors.green.shade700, size: 18),
              SizedBox(width: 6),
              Text(
                global.language('data_filter'),
                style: TextStyle(
                  fontSize: 15,
                  fontWeight: FontWeight.w600,
                  color: Colors.green,
                ),
              ),
            ],
          ),
          const SizedBox(height: 10),
          // _buildShowDetailsToggle(),
          // const SizedBox(height: 12),
          _buildShowCancelledToggle(),
          const SizedBox(height: 10),
          if (widget.reportType != BiReportType.purchasepartial) ...[
            _buildSaleTypeSelection(),
            const SizedBox(height: 10),
            _buildPosTypeSelection(),
          ],
        ],
      ),
    );
  }

  // Widget _buildShowDetailsToggle() {
  //   return Container(
  //     padding: const EdgeInsets.all(16),
  //     decoration: BoxDecoration(
  //       color: Colors.white,
  //       borderRadius: BorderRadius.circular(8),
  //     ),
  //     child: Row(
  //       children: [
  //         Expanded(
  //           child: Column(
  //             crossAxisAlignment: CrossAxisAlignment.start,
  //             children: [
  //               const Text(
  //                 'แสดงรายละเอียดสินค้า',
  //                 style: TextStyle(
  //                   fontSize: 14,
  //                   fontWeight: FontWeight.w600,
  //                 ),
  //               ),
  //               const SizedBox(height: 4),
  //               Text(
  //                 _showDetails ? 'แสดงข้อมูลรายการสินค้าในแต่ละใบเสร็จ' : 'แสดงเฉพาะยอดรวมของแต่ละใบเสร็จ',
  //                 style: TextStyle(
  //                   fontSize: 12,
  //                   color: Colors.grey.shade600,
  //                 ),
  //               ),
  //             ],
  //           ),
  //         ),
  //         Switch.adaptive(
  //           value: _showDetails,
  //           onChanged: (value) {
  //             setState(() {
  //               _showDetails = value;
  //             });
  //           },
  //           activeColor: Colors.indigo.shade600,
  //         ),
  //       ],
  //     ),
  //   );
  // }

  Widget _buildShowCancelledToggle() {
    return Container(
      padding: const EdgeInsets.all(12), // ลดจาก 16 เป็น 12
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(6),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            global.language('show_cancelled_documents'),
            style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
          ),
          SizedBox(height: 8),
          Row(
            children: [
              Expanded(
                child: RadioListTile<String>(
                  title: Text(
                    global.language('hide_cancelled_documents'),
                    style: TextStyle(fontSize: 14),
                  ),
                  value: global.language('hide_cancelled_documents'),
                  groupValue: _showCancelled,
                  onChanged: (value) {
                    setState(() {
                      _showCancelled = value!;
                    });
                  },
                  dense: true,
                  contentPadding: EdgeInsets.zero,
                  activeColor: Colors.green.shade600,
                ),
              ),
              Expanded(
                child: RadioListTile<String>(
                  title: Text(
                    global.language('show_cancelled_documents'),
                    style: TextStyle(fontSize: 14),
                  ),
                  value: global.language('show_cancelled_documents'),
                  groupValue: _showCancelled,
                  onChanged: (value) {
                    setState(() {
                      _showCancelled = value!;
                    });
                  },
                  dense: true,
                  contentPadding: EdgeInsets.zero,
                  activeColor: Colors.green.shade600,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildSaleTypeSelection() {
    return Container(
      padding: const EdgeInsets.all(12), // ลดจาก 16 เป็น 12
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(6),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            global.language('sale_type'),
            style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
          ),
          SizedBox(height: 8),
          Row(
            children: [
              Expanded(
                child: RadioListTile<String>(
                  title: Text(global.language('all'), style: TextStyle(fontSize: 14)),
                  value: global.language('all'),
                  groupValue: _saleType,
                  onChanged: (value) {
                    setState(() {
                      _saleType = value!;
                    });
                  },
                  dense: true,
                  contentPadding: EdgeInsets.zero,
                  activeColor: Colors.green.shade600,
                ),
              ),
              Expanded(
                child: RadioListTile<String>(
                  title: Text(
                    global.language('inquiry_credit'),
                    style: TextStyle(fontSize: 14),
                  ),
                  value: global.language('inquiry_credit'),
                  groupValue: _saleType,
                  onChanged: (value) {
                    setState(() {
                      _saleType = value!;
                    });
                  },
                  dense: true,
                  contentPadding: EdgeInsets.zero,
                  activeColor: Colors.green.shade600,
                ),
              ),
              Expanded(
                child: RadioListTile<String>(
                  title: Text(global.language('cash'), style: TextStyle(fontSize: 14)),
                  value: global.language('cash'),
                  groupValue: _saleType,
                  onChanged: (value) {
                    setState(() {
                      _saleType = value!;
                    });
                  },
                  dense: true,
                  contentPadding: EdgeInsets.zero,
                  activeColor: Colors.green.shade600,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildPosTypeSelection() {
    return Container(
      padding: const EdgeInsets.all(12), // ลดจาก 16 เป็น 12
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(6),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            global.language('show_items'),
            style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600),
          ),
          SizedBox(height: 8),
          Row(
            children: [
              Expanded(
                child: RadioListTile<String>(
                  title: Text(global.language('all'), style: TextStyle(fontSize: 14)),
                  value: global.language('all'),
                  groupValue: _posType,
                  onChanged: (value) {
                    setState(() {
                      _posType = value!;
                    });
                  },
                  dense: true,
                  contentPadding: EdgeInsets.zero,
                  activeColor: Colors.green.shade600,
                ),
              ),
              Expanded(
                child: RadioListTile<String>(
                  title: Text(
                    global.language('pos_only'),
                    style: TextStyle(fontSize: 14),
                  ),
                  value: global.language('pos_only'),
                  groupValue: _posType,
                  onChanged: (value) {
                    setState(() {
                      _posType = value!;
                    });
                  },
                  dense: true,
                  contentPadding: EdgeInsets.zero,
                  activeColor: Colors.green.shade600,
                ),
              ),
              Expanded(
                child: RadioListTile<String>(
                  title: Text(global.language('back_office'), style: TextStyle(fontSize: 14)),
                  value: global.language('back_office'),
                  groupValue: _posType,
                  onChanged: (value) {
                    setState(() {
                      _posType = value!;
                    });
                  },
                  dense: true,
                  contentPadding: EdgeInsets.zero,
                  activeColor: Colors.green.shade600,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildActionButtons() {
    // ตรวจสอบเงื่อนไขการค้นหา
    bool canSearch = true;
    String errorMessage = '';

    // ตรวจสอบวันที่สำหรับรายงานที่ต้องการวันที่
    if (widget.reportType != BiReportType.stockMovement &&
        widget.reportType != BiReportType.stockBalance) {
      // รายงานทั่วไปต้องมีทั้ง fromDate และ toDate
      if (_fromDate == null || _toDate == null) {
        canSearch = false;
        errorMessage = global.language('report_dedebi_select_date_range');
      } else if (_fromDate!.isAfter(_toDate!)) {
        canSearch = false;
        errorMessage = global.language(
          'report_dedebi_from_date_not_exceed_to_date',
        );
      }
    } else if (widget.reportType == BiReportType.stockBalance) {
      // Stock Balance ต้องมี toDate
      if (_toDate == null) {
        canSearch = false;
        errorMessage = global.language('report_condition_select_date');
      }
    }

    // เฉพาะ Stock Movement ต้องเลือกบาร์โค้ด (Stock Balance ไม่บังคับ)
    if (widget.reportType == BiReportType.stockMovement) {
      if (_selectedBarcodes.selectedEntities.isEmpty) {
        canSearch = false;
        errorMessage = global.language('report_condition_select_barcode_first');
      }
    }

    return Column(
      children: [
        // แสดงข้อความแจ้งเตือนหากไม่สามารถค้นหาได้
        if (!canSearch) ...[
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(10),
            margin: const EdgeInsets.only(bottom: 12), // ลดจาก 16 เป็น 12
            decoration: BoxDecoration(
              color: Colors.red.shade50,
              borderRadius: BorderRadius.circular(6),
              border: Border.all(color: Colors.red.shade200),
            ),
            child: Row(
              children: [
                Icon(Icons.warning_amber, color: Colors.red.shade600, size: 18),
                const SizedBox(width: 6),
                Expanded(
                  child: Text(
                    errorMessage,
                    style: TextStyle(
                      color: Colors.red.shade700,
                      fontSize: 14,
                      fontWeight: FontWeight.w500,
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],

        // ปุ่มแอ็คชัน
        Row(
          children: [
            Expanded(
              child: OutlinedButton.icon(
                onPressed: () => Navigator.of(context).pop(),
                icon: Icon(Icons.close),
                label: Text(global.language('report_condition_cancel')),
                style: OutlinedButton.styleFrom(
                  padding: const EdgeInsets.symmetric(
                    vertical: 12,
                  ), // ลดจาก 16 เป็น 12
                  side: BorderSide(color: Colors.grey.shade400),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(6),
                  ),
                ),
              ),
            ),
            const SizedBox(width: 12), // ลดจาก 16 เป็น 12
            Expanded(
              child: ElevatedButton.icon(
                onPressed: canSearch
                    ? () {
                        late ReportConditions conditions;
                        if (widget.reportType == BiReportType.sale) {
                          conditions = ReportConditions(
                            fromDate: _fromDate,
                            toDate: _toDate,
                            showDetails: _showDetails,
                            showCancelled: _showCancelled,
                            saleType: _saleType,
                            posType: _posType,
                            selectedBranches: _selectedBranches,
                            selectedCreditors: _selectedCreditors,
                            selectedSalespersons: _selectedSalespersons,
                          );
                        } else if (widget.reportType ==
                            BiReportType.saleDaily) {
                          conditions = ReportConditions(
                            fromDate: _fromDate,
                            toDate: _toDate,
                            showDetails: _showDetails,
                            showCancelled: _showCancelled,
                            saleType: _saleType,
                            posType: _posType,
                            selectedBranches: _selectedBranches,
                          );
                        } else if (widget.reportType ==
                            BiReportType.stockMovement) {
                          conditions = ReportConditions(
                            fromDate: _fromDate,
                            toDate: _toDate,
                            selectedBarcodes: _selectedBarcodes,
                          );
                        } else if (widget.reportType ==
                            BiReportType.stockBalance) {
                          conditions = ReportConditions(
                            toDate: _toDate,
                            selectedBarcodes: _selectedBarcodes,
                          );
                        } else if (widget.reportType ==
                            BiReportType.paymentDaily) {
                          conditions = ReportConditions(
                            fromDate: _fromDate,
                            toDate: _toDate,
                            selectedBranches: _selectedBranches,
                          );
                        } else if (widget.reportType ==
                            BiReportType.saleReturn) {
                          conditions = ReportConditions(
                            fromDate: _fromDate,
                            toDate: _toDate,
                            showDetails: _showDetails,
                            selectedBranches: _selectedBranches,
                          );
                        } else if (widget.reportType ==
                            BiReportType.purchasepartial) {
                          conditions = ReportConditions(
                            fromDate: _fromDate,
                            toDate: _toDate,
                            showDetails: _showDetails,
                            showCancelled: _showCancelled,
                            selectedBranches: _selectedBranches,
                            selectedCreditors: _selectedCreditors,
                          );
                        } else if (widget.reportType ==
                            BiReportType.grossProfitByDocument) {
                          conditions = ReportConditions(
                            fromDate: _fromDate,
                            toDate: _toDate,
                            showCancelled: _showCancelled,
                            selectedBranches: _selectedBranches,
                            selectedCreditors: _selectedCreditors,
                          );
                        } else if (widget.reportType ==
                            BiReportType.grossProfitByProduct) {
                          conditions = ReportConditions(
                            fromDate: _fromDate,
                            toDate: _toDate,
                            selectedBarcodes: _selectedBarcodes,
                            selectedBranches: _selectedBranches,
                          );
                        } else if (widget.reportType == BiReportType.vatSale) {
                          conditions = ReportConditions(
                            fromDate: _fromDate,
                            toDate: _toDate,
                            selectedBranches: _selectedBranches,
                          );
                        } else if (widget.reportType == BiReportType.vatBuy) {
                          conditions = ReportConditions(
                            fromDate: _fromDate,
                            toDate: _toDate,
                            selectedBranches: _selectedBranches,
                          );
                        } else {
                          // Show error
                          global.showInfoSnackBar(context, global.language('report_condition_cannot_create'));
                          return; // เพิ่ม return เพื่อไม่ให้ทำงานต่อ
                        }

                        Navigator.of(context).pop();
                        widget.onConditionsSet(conditions);
                      }
                    : null, // ปิดการใช้งานปุ่มหากไม่สามารถค้นหาได้
                icon: Icon(Icons.search),
                label: Text(global.language('report_condition_search')),
                style: ElevatedButton.styleFrom(
                  backgroundColor: canSearch
                      ? Colors.indigo.shade600
                      : Colors.grey.shade400,
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(
                    vertical: 12,
                  ), // ลดจาก 16 เป็น 12
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(6),
                  ),
                  elevation: canSearch ? 2 : 0,
                ),
              ),
            ),
          ],
        ),
      ],
    );
  }
}
