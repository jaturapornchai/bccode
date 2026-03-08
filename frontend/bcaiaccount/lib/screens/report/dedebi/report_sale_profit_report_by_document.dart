import 'dart:collection';
import 'dart:convert';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:intl/intl.dart';
import 'package:pdf/pdf.dart' as pdf;
import 'package:printing/printing.dart' as printing;
import 'package:smlaicloud/bloc/company_branch/company_branch_bloc.dart';
import 'package:smlaicloud/bloc/debtor/debtor_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/bi_report/branch_selection_model.dart';
import 'package:smlaicloud/model/bi_report/entity_selection_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/result_table_models.dart';
import 'package:smlaicloud/repositories/company_branch_repository.dart';
import 'package:smlaicloud/repositories/debtor_repository.dart';
import 'package:smlaicloud/screens/report/pdf_saver.dart';
import 'package:smlaicloud/screen_search/dedebi/multi_branch_search_screen.dart';
import 'package:smlaicloud/screen_search/dedebi/multi_entity_search_screen.dart';
import 'package:smlaicloud/services/result_table_api_service.dart';
import 'package:smlaicloud/utils/date_picker.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/utils/select_product_barcode.dart';
import 'package:smlaicloud/services/sales_report_api_service.dart';
import 'package:syncfusion_flutter_pdfviewer/pdfviewer.dart';
import 'dart:typed_data';

/// หน้าจอเงื่อนไขสำหรับรายงานกำไรขั้นต้นตามเอกสาร
/// ใช้เป็นพื้นที่ทดลอง UX ก่อนเชื่อมต่อ API รายงานจริง
class ReportSaleProfitReportByDocumentScreen extends StatefulWidget {
  const ReportSaleProfitReportByDocumentScreen({super.key});

  @override
  State<ReportSaleProfitReportByDocumentScreen> createState() =>
      _ReportSaleProfitReportByDocumentScreenState();
}

class _ReportSaleProfitReportByDocumentScreenState
    extends State<ReportSaleProfitReportByDocumentScreen>
    with SingleTickerProviderStateMixin {
  late DateTime _fromDate;
  late DateTime _toDate;
  BranchSelectionModel _selectedBranches = const BranchSelectionModel(
    selectedBranches: [],
    isCancel: false,
  );
  EntitySelectionModel _selectedCreditors = const EntitySelectionModel(
    selectedEntities: [],
    isCancel: false,
  );
  final List<SelectedProductItemStruct> _selectedProducts = [];
  bool _includeItemDetails = true;
  bool _sortAscending = true;
  String? _selectedQuickRangeKey = 'thisMonth';
  bool _isProcessing = false;
  late TabController _tabController;

  // Result Table State
  String? _currentGuid;
  List<Map<String, dynamic>> _resultData = [];
  PaginationInfo? _paginationInfo;
  bool _isLoadingResults = false;
  bool _isLoadingMore = false;
  bool _hasMoreData = true;
  final int _pageSize = 100;
  final ScrollController _scrollController = ScrollController();
  double _grandTotalCount = 0;
  double _grandTotalAmount = 0;
  double _grandTotalCost = 0;
  double _grandTotalGrossProfit = 0;

  final Map<String, NumberFormat> _numberFormatCache = {};
  final Map<String, DateFormat> _dateFormatCache = {};

  // PDF State
  Uint8List? _pdfBytes;
  bool _isGeneratingPdf = false;
  String _selectedOrientation = 'P';
  String _selectedPageSize = 'A4';

  final Color _primaryColor = Colors.indigo.shade700;
  final Color _secondaryColor = Colors.indigo.shade50;
  final Color _level0RowColor = const Color(0xFFEAF0FF);
  final Color _level1RowColor = const Color(0xFFF1FFF7);
  final Color _summaryRowColor = const Color(0xFFFFF4E3);

  late TextStyle _titleStyle;
  late TextStyle _subtitleStyle;
  late TextStyle _bodyStyle;
  late final Map<String, DateTimeRange Function()> _quickRanges;

  static const _headerAlias = ResultTableQueryAlias.documentHeader;
  static const _detailAlias = ResultTableQueryAlias.documentLines;
  static final List<_ReportColumn> _documentColumns = [
    _ReportColumn(
      key: 'docdate',
      label: global.language('date'),
      flex: 8,
      dataType: _ReportColumnType.date,
      format: 'dd/MM/yyyy',
      useBuddhistYear: true,
    ),
    _ReportColumn(
      key: 'doctime',
      label: global.language('doc_time'),
      flex: 6,
      hideWhenSummary: true,
      dataType: _ReportColumnType.time,
      format: 'HH:mm',
    ),
    _ReportColumn(key: 'docno', label: global.language('docno'), flex: 11),
    _ReportColumn(
      key: 'debtorcode',
      label: global.language('customer_code'),
      flex: 8,
      hideWhenSummary: true,
    ),
    _ReportColumn(
      key: 'debtorname',
      label: global.language('customer_name'),
      flex: 27,
      hideWhenSummary: true,
    ),
    _ReportColumn(
      key: 'totalqty',
      label: global.language('total_documnet'),
      flex: 8,
      align: TextAlign.right,
      dataType: _ReportColumnType.number,
      format: '#,##0.00',
    ),
    _ReportColumn(
      key: 'totalamount',
      label: global.language('payment_daily_total_amount_column'),
      flex: 10,
      align: TextAlign.right,
      dataType: _ReportColumnType.number,
      format: '#,##0.00',
    ),
    _ReportColumn(
      key: 'calcamount',
      label: global.language('total_cost'),
      flex: 10,
      align: TextAlign.right,
      dataType: _ReportColumnType.number,
      format: '#,##0.00',
    ),
    _ReportColumn(
      key: 'grossprofit',
      label: global.language('gross_profit'),
      flex: 10,
      align: TextAlign.right,
      dataType: _ReportColumnType.number,
      format: '#,##0.00',
    ),
  ];

  static final List<_ReportColumn> _detailColumns = [
    _ReportColumn(key: 'itemcode', label: global.language('product_code'), flex: 11),
    _ReportColumn(key: 'itemname', label: global.language('import_product_detail.product_name'), flex: 17),
    _ReportColumn(key: 'barcode', label: global.language('barcode'), flex: 10),
    _ReportColumn(key: 'unitname', label: 'ชื่อหน่วย', flex: 6),
    _ReportColumn(
      key: 'price',
      label: global.language('price'),
      flex: 8,
      align: TextAlign.right,
      dataType: _ReportColumnType.number,
      format: '#,##0.00',
    ),
    _ReportColumn(
      key: 'averagecost',
      label: global.language('average_cost'),
      flex: 8,
      align: TextAlign.right,
      dataType: _ReportColumnType.number,
      format: '#,##0.00',
    ),
    _ReportColumn(
      key: 'totalqty',
      label: global.language('enter_qty'),
      flex: 8,
      align: TextAlign.right,
      dataType: _ReportColumnType.number,
      format: '#,##0.###',
    ),
    _ReportColumn(
      key: 'totalamount',
      label: global.language('payment_daily_total_amount_column'),
      flex: 10,
      align: TextAlign.right,
      dataType: _ReportColumnType.number,
      format: '#,##0.00',
    ),
    _ReportColumn(
      key: 'calcamount',
      label: global.language('total_cost'),
      flex: 10,
      align: TextAlign.right,
      dataType: _ReportColumnType.number,
      format: '#,##0.00',
    ),
    _ReportColumn(
      key: 'grossprofit',
      label: global.language('gross_profit'),
      flex: 10,
      align: TextAlign.right,
      dataType: _ReportColumnType.number,
      format: '#,##0.00',
    ),
  ];

  static Map<String, String> get _orientationLabels => {
    'L': global.language("landscape"),
    'P': global.language("portrait"),
  };

  static const List<String> _pageSizeOptions = [
    'A4',
    'A3',
    'A5',
    'Letter',
    'Legal',
  ];

  @override
  void initState() {
    super.initState();
    _tabController = TabController(length: 3, vsync: this);
    _scrollController.addListener(_onScroll);
    final now = DateTime.now();
    _fromDate = DateTime(now.year, now.month, 1);
    _toDate = DateTime(now.year, now.month, now.day);
    _initializeStyles();
    _quickRanges = {
      'today': () {
        final today = _normalize(DateTime.now());
        return DateTimeRange(start: today, end: today);
      },
      'yesterday': () {
        final yesterday = _normalize(
          DateTime.now().subtract(const Duration(days: 1)),
        );
        return DateTimeRange(start: yesterday, end: yesterday);
      },
      'thisWeek': () {
        final now = DateTime.now();
        final start = _normalize(now.subtract(Duration(days: now.weekday - 1)));
        final end = _normalize(start.add(const Duration(days: 6)));
        return DateTimeRange(start: start, end: end);
      },
      'lastWeek': () {
        final now = DateTime.now();
        final thisWeekStart = _normalize(
          now.subtract(Duration(days: now.weekday - 1)),
        );
        final start = thisWeekStart.subtract(const Duration(days: 7));
        final end = _normalize(start.add(const Duration(days: 6)));
        return DateTimeRange(start: start, end: end);
      },
      'thisMonth': () {
        final now = DateTime.now();
        final start = DateTime(now.year, now.month, 1);
        final end = DateTime(now.year, now.month + 1, 0);
        return DateTimeRange(start: start, end: end);
      },
      'lastMonth': () {
        final now = DateTime.now();
        final start = DateTime(now.year, now.month - 1, 1);
        final end = DateTime(now.year, now.month, 0);
        return DateTimeRange(start: start, end: end);
      },
    };
    for (var i = 0; i < 12; i++) {
      final date = DateTime(now.year, now.month - i, 1);
      final key = 'month_$i';
      _quickRanges[key] = () {
        final start = DateTime(date.year, date.month, 1);
        final end = DateTime(date.year, date.month + 1, 0);
        return DateTimeRange(start: start, end: end);
      };
    }
    _quickRanges['year_this'] = () {
      final start = DateTime(now.year, 1, 1);
      final end = DateTime(now.year + 1, 1, 0);
      return DateTimeRange(start: start, end: end);
    };
    _quickRanges['year_prev'] = () {
      final start = DateTime(now.year - 1, 1, 1);
      final end = DateTime(now.year, 1, 0);
      return DateTimeRange(start: start, end: end);
    };
    for (var yearsAgo = 2; yearsAgo <= 6; yearsAgo++) {
      final year = now.year - yearsAgo;
      final key = 'year_history_$yearsAgo';
      _quickRanges[key] = () {
        final start = DateTime(year, 1, 1);
        final end = DateTime(year + 1, 1, 0);
        return DateTimeRange(start: start, end: end);
      };
    }
    _applyQuickRange(_selectedQuickRangeKey!);
  }

  void _initializeStyles() {
    _titleStyle = TextStyle(
      fontSize: 16,
      fontWeight: FontWeight.w700,
      color: _primaryColor,
    );
    _subtitleStyle = TextStyle(
      fontSize: 14,
      fontWeight: FontWeight.w500,
      color: _primaryColor.withValues(alpha: 0.85),
    );
    _bodyStyle = TextStyle(fontSize: 13, color: Colors.grey.shade800);
  }

  @override
  void dispose() {
    _tabController.dispose();
    _scrollController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Colors.grey.shade100,
      appBar: AppBar(
        backgroundColor: _primaryColor,
        foregroundColor: Colors.white,
        title: TabBar(
          controller: _tabController,
          indicatorColor: Colors.white,
          indicatorWeight: 3,
          labelColor: Colors.white,
          unselectedLabelColor: Colors.white.withValues(alpha: 0.7),
          labelStyle: const TextStyle(fontWeight: FontWeight.bold),
          tabs: [
            Tab(
              child: Wrap(
                alignment: WrapAlignment.center,
                crossAxisAlignment: WrapCrossAlignment.center,
                spacing: 8,
                children: [Icon(Icons.settings), Text(global.language('choice'))],
              ),
            ),
            Tab(
              child: Wrap(
                alignment: WrapAlignment.center,
                crossAxisAlignment: WrapCrossAlignment.center,
                spacing: 8,
                children: [Icon(Icons.table_chart), Text(global.language('display'))],
              ),
            ),
            Tab(
              child: Wrap(
                alignment: WrapAlignment.center,
                crossAxisAlignment: WrapCrossAlignment.center,
                spacing: 8,
                children: [Icon(Icons.picture_as_pdf), Text('PDF')],
              ),
            ),
          ],
        ),
      ),
      body: TabBarView(
        controller: _tabController,
        children: [_buildConditionTab(), _buildResultTab(), _buildPdfTab()],
      ),
    );
  }

  Widget _buildConditionTab() {
    return SafeArea(
      child: ListView(
        padding: const EdgeInsets.only(bottom: 24),
        children: [
          const SizedBox(height: 12),
          _buildDateRangePanel(),
          _buildBranchAndCustomerPanel(),
          _buildSummaryPanel(),
          _buildDetailOptionsPanel(),
          _buildActionButtons(),
        ],
      ),
    );
  }

  Widget _buildResultTab() {
    if (_currentGuid == null || _resultData.isEmpty) {
      return Center(
        child: Column(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(Icons.insert_chart, size: 64, color: Colors.grey.shade400),
            const SizedBox(height: 16),
            Text(
              global.language('analysis_result_here'),
              style: TextStyle(fontSize: 18, color: Colors.grey.shade600),
            ),
            SizedBox(height: 8),
            Text(
              'กรุณากดปุ่ม global.language("process") ในแท็บเงื่อนไข',
              style: TextStyle(fontSize: 14, color: Colors.grey.shade500),
            ),
          ],
        ),
      );
    }

    return Column(
      children: [
        _buildResultHeaderRow(),
        Expanded(
          child: _isLoadingResults
              ? const Center(child: CircularProgressIndicator())
              : _buildSequentialResultList(),
        ),
        if (_grandTotalCount > 0) _buildGrandTotalBar(),
      ],
    );
  }

  Widget _buildSequentialResultList() {
    if (_resultData.isEmpty) {
      return const SizedBox.shrink();
    }

    return ListView.builder(
      controller: _scrollController,
      padding: const EdgeInsets.all(0),
      itemCount: _resultData.length + (_isLoadingMore ? 1 : 0),
      itemBuilder: (context, index) {
        if (index >= _resultData.length) {
          return _buildLoadMoreIndicator();
        }

        final row = _resultData[index];
        final level = _resolveRowLevel(row);

        if (level == 0) {
          return _buildDocumentRow(row);
        }

        if (level == 1) {
          if (!_includeItemDetails) return const SizedBox.shrink();
          return _buildProductRow(row);
        }

        return _buildUnknownRow(row);
      },
    );
  }

  Widget _buildLoadMoreIndicator() {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 16),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          const SizedBox(
            width: 20,
            height: 20,
            child: CircularProgressIndicator(strokeWidth: 2),
          ),
          const SizedBox(width: 12),
          Text(global.language('loading_more_data'), style: _bodyStyle),
        ],
      ),
    );
  }

  Widget _buildResultHeaderRow() {
    final headerStyle = _bodyStyle.copyWith(
      fontWeight: FontWeight.bold,
      color: Colors.black,
      fontSize: 14,
    );

    return Container(
      width: double.infinity,
      decoration: BoxDecoration(
        color: Colors.white,
        border: Border.all(color: Colors.green),
        boxShadow: [
          BoxShadow(
            color: Colors.black,
            blurRadius: 4,
            offset: const Offset(0, 2),
          ),
        ],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: _documentColumns
                .map(
                  (column) => _buildHeaderCell(
                    column.label,
                    column.flex,
                    headerStyle,
                    align: column.align,
                  ),
                )
                .toList(),
          ),
          if (_includeItemDetails)
            Padding(
              padding: const EdgeInsets.only(top: 6),
              child: Row(
                children: _detailColumns
                    .map(
                      (column) => _buildHeaderCell(
                        column.label,
                        column.flex,
                        headerStyle,
                        align: column.align,
                      ),
                    )
                    .toList(),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildHeaderCell(
    String label,
    int flex,
    TextStyle style, {
    TextAlign align = TextAlign.left,
  }) {
    return Expanded(
      flex: flex,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 4),
        child: Text(
          label,
          style: style,
          textAlign: align,
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
        ),
      ),
    );
  }

  Widget _buildDocumentRow(Map<String, dynamic> row) {
    final data = _extractRowData(row);
    final typejson = (row['typejson'] as num?)?.toInt() ?? 0;
    final isSummary = typejson == 1;

    final backgroundColor = isSummary ? _summaryRowColor : _level0RowColor;
    final baseTextColor = isSummary
        ? Colors.deepOrange.shade800
        : Colors.indigo.shade900;
    final textStyle = (_includeItemDetails)
        ? _bodyStyle.copyWith(color: baseTextColor, fontWeight: FontWeight.bold)
        : _bodyStyle.copyWith(color: baseTextColor);
    final numericStyle = (_includeItemDetails)
        ? textStyle.copyWith(
            fontFeatures: const [FontFeature.tabularFigures()],
            fontWeight: FontWeight.bold,
          )
        : textStyle.copyWith(
            fontFeatures: const [FontFeature.tabularFigures()],
          );

    return Container(
      decoration: BoxDecoration(
        color: backgroundColor,
        border: Border.all(
          color: isSummary
              ? Colors.deepOrange.shade100
              : Colors.indigo.shade100,
        ),
      ),
      child: Row(
        children: _documentColumns.map((column) {
          final effectiveText = _formatCellValue(
            column,
            data,
            isSummary: isSummary,
          );
          final style = column.isNumeric ? numericStyle : textStyle;
          return _buildDocumentCell(
            text: effectiveText,
            flex: column.flex,
            style: style,
            align: column.align,
          );
        }).toList(),
      ),
    );
  }

  Widget _buildDocumentCell({
    required String text,
    required int flex,
    required TextStyle style,
    TextAlign align = TextAlign.left,
  }) {
    return Expanded(
      flex: flex,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 4),
        child: Text(
          text,
          style: style,
          textAlign: align,
          maxLines: 1,
          overflow: TextOverflow.ellipsis,
        ),
      ),
    );
  }

  Widget _buildDetailCell({
    required String text,
    required int flex,
    required TextStyle style,
    TextAlign align = TextAlign.left,
  }) {
    return Expanded(
      flex: flex,
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 4),
        child: Text(text, style: style, textAlign: align),
      ),
    );
  }

  String _resolveUnitDisplay(Map<String, dynamic> data) {
    final rawUnitName = (data['unitname']?.toString() ?? '').trim();
    if (rawUnitName.isNotEmpty) return rawUnitName;
    final unitCode = data['unitcode']?.toString();
    if (unitCode != null && unitCode.isNotEmpty) return unitCode;
    return '-';
  }

  String _formatCellValue(
    _ReportColumn column,
    Map<String, dynamic> data, {
    bool isSummary = false,
    String? detailUnit,
  }) {
    if (isSummary && column.hideWhenSummary) {
      return '';
    }

    if (column.key == 'docno' && isSummary) {
      return global.language('total_by_date');
    }

    if (column.key == 'unitname' && detailUnit != null) {
      return detailUnit;
    }

    final value = data[column.key];
    return _formatValueByType(column, value);
  }

  String _formatValueByType(_ReportColumn column, dynamic value) {
    switch (column.dataType) {
      case _ReportColumnType.number:
        return _formatNumber(value, pattern: column.format);
      case _ReportColumnType.date:
        return _formatDateValue(value, column);
      case _ReportColumnType.time:
        return _formatTimeValue(value, column);
      case _ReportColumnType.text:
        return _formatTextValue(value);
    }
  }

  String _formatTextValue(dynamic value) {
    if (value == null) return '-';
    final text = value.toString().trim();
    return text.isEmpty ? '-' : text;
  }

  String _formatDateValue(dynamic value, _ReportColumn column) {
    final date = _parseDateValue(value);
    if (date == null) {
      return value == null ? '' : value.toString();
    }

    final pattern = column.format ?? 'dd/MM/yyyy';
    final displayDate = column.useBuddhistYear
        ? DateTime(
            date.year + 543,
            date.month,
            date.day,
            date.hour,
            date.minute,
            date.second,
            date.millisecond,
            date.microsecond,
          )
        : date;
    return _resolveDateFormat(pattern).format(displayDate);
  }

  String _formatTimeValue(dynamic value, _ReportColumn column) {
    final raw = value?.toString().trim();
    if (raw == null || raw.isEmpty) return '';
    final parsed = _parseTimeValue(raw);
    if (parsed == null) {
      return raw;
    }
    final pattern = column.format ?? 'HH:mm';
    return _resolveDateFormat(pattern).format(parsed);
  }

  DateTime? _parseDateValue(dynamic value) {
    if (value == null) return null;
    if (value is DateTime) return value;
    if (value is int) {
      return DateTime.fromMillisecondsSinceEpoch(value);
    }
    if (value is String && value.isNotEmpty) {
      final normalized = value.replaceFirst(' ', 'T');
      return DateTime.tryParse(normalized) ?? DateTime.tryParse(value);
    }
    return null;
  }

  DateTime? _parseTimeValue(String raw) {
    final normalized = raw.replaceAll('.', ':');
    if (normalized.contains(':')) {
      final parts = normalized.split(':');
      final values = <int>[];
      for (final part in parts) {
        final parsed = int.tryParse(part);
        if (parsed == null) {
          return null;
        }
        values.add(parsed);
      }
      final hour = values.isNotEmpty ? values[0] : 0;
      final minute = values.length > 1 ? values[1] : 0;
      final second = values.length > 2 ? values[2] : 0;
      return DateTime(1970, 1, 1, hour, minute, second);
    }

    if (RegExp(r'^\d{6}$').hasMatch(normalized)) {
      final hour = int.parse(normalized.substring(0, 2));
      final minute = int.parse(normalized.substring(2, 4));
      final second = int.parse(normalized.substring(4, 6));
      return DateTime(1970, 1, 1, hour, minute, second);
    }

    if (RegExp(r'^\d{4}$').hasMatch(normalized)) {
      final hour = int.parse(normalized.substring(0, 2));
      final minute = int.parse(normalized.substring(2, 4));
      return DateTime(1970, 1, 1, hour, minute);
    }

    return null;
  }

  List<Map<String, dynamic>> _buildPdfColumns(List<_ReportColumn> columns) {
    return columns
        .map(
          (column) => {
            'field': column.key,
            'label': column.label,
            'flex': column.flex,
            'align': _textAlignToString(column.align),
            'is_numeric': column.isNumeric,
            'hide_when_summary': column.hideWhenSummary,
            'data_type': column.dataType.wireValue,
            if (column.format != null) 'format': column.format,
            if (column.useBuddhistYear) 'use_buddhist_year': true,
          },
        )
        .toList();
  }

  Map<String, Map<String, dynamic>> _buildColumnSchema(bool includeDetails) {
    final schema = <String, Map<String, dynamic>>{};

    void addColumns(List<_ReportColumn> columns) {
      for (final column in columns) {
        schema.putIfAbsent(column.key, () {
          return {
            'label': column.label,
            'align': _textAlignToString(column.align),
            'flex': column.flex,
            'is_numeric': column.isNumeric,
            'data_type': column.dataType.wireValue,
            'hide_when_summary': column.hideWhenSummary,
            if (column.format != null) 'format': column.format,
            if (column.useBuddhistYear) 'use_buddhist_year': true,
          };
        });
      }
    }

    addColumns(_documentColumns);
    if (includeDetails) {
      addColumns(_detailColumns);
    }

    return schema;
  }

  Map<String, dynamic> _buildPdfStyleGuide() => {
    'palette': 'grayscale',
    'use_fill': false,
    'header': {
      'background': '#FFFFFF',
      'text': '#111111',
      'border': '#444444',
      'font_weight': 'bold',
    },
    'detail': {'background': '#FFFFFF', 'text': '#222222', 'border': '#888888'},
    'summary': {
      'background': '#FFFFFF',
      'text': '#000000',
      'border': '#000000',
      'font_weight': 'bold',
    },
    'table': {'row_spacing': 4, 'column_spacing': 8, 'grid_color': '#CCCCCC'},
  };

  Map<String, String> _buildNumberFormats() {
    final formats = <String, String>{};
    void addFormats(List<_ReportColumn> columns) {
      for (final column in columns) {
        if (column.dataType == _ReportColumnType.number &&
            column.format != null) {
          formats[column.key] = column.format!;
        }
      }
    }

    addFormats(_documentColumns);
    addFormats(_detailColumns);
    return formats;
  }

  String _textAlignToString(TextAlign align) {
    switch (align) {
      case TextAlign.right:
        return 'right';
      case TextAlign.center:
        return 'center';
      case TextAlign.justify:
        return 'justify';
      case TextAlign.end:
        return 'end';
      case TextAlign.start:
        return 'start';
      case TextAlign.left:
        return 'left';
    }
  }

  Widget _buildProductRow(Map<String, dynamic> row) {
    final data = _extractRowData(row);
    final detailTextColor = Colors.teal.shade900;
    final rowColor = _level1RowColor;
    final textStyle = _bodyStyle.copyWith(color: detailTextColor);
    final numberStyle = _bodyStyle.copyWith(
      fontWeight: FontWeight.w400,
      color: detailTextColor,
      fontFeatures: const [FontFeature.tabularFigures()],
    );
    final displayUnit = _resolveUnitDisplay(data);

    return Container(
      decoration: BoxDecoration(
        color: rowColor,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.teal.shade100),
      ),
      child: Row(
        children: _detailColumns.map((column) {
          final value = _formatCellValue(column, data, detailUnit: displayUnit);
          final style = column.isNumeric ? numberStyle : textStyle;
          return _buildDetailCell(
            text: value,
            flex: column.flex,
            style: style,
            align: column.align,
          );
        }).toList(),
      ),
    );
  }

  Widget _buildUnknownRow(Map<String, dynamic> row) {
    const encoder = JsonEncoder.withIndent('  ');
    final pretty = encoder.convert(row);

    return Container(
      margin: const EdgeInsets.symmetric(vertical: 4),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: Colors.yellow.shade50,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.yellow.shade200),
      ),
      child: Text(
        pretty,
        style: _bodyStyle.copyWith(fontSize: 12, fontFamily: 'monospace'),
      ),
    );
  }

  Map<String, dynamic> _extractRowData(Map<String, dynamic> row) {
    final data = row['data'];
    if (data is Map<String, dynamic>) {
      return data;
    }
    return row;
  }

  int _resolveRowLevel(Map<String, dynamic> row) {
    final level = row['level'];
    if (level is num) return level.toInt();
    final typeJson = row['typejson'];
    if (typeJson is num) return typeJson.toInt();
    return 0;
  }

  Widget _buildGrandTotalBar() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
      decoration: BoxDecoration(
        color: Colors.green.shade50,
        border: Border(top: BorderSide(color: Colors.green.shade300, width: 2)),
        boxShadow: [
          BoxShadow(
            color: Colors.black.withValues(alpha: 0.1),
            blurRadius: 4,
            offset: const Offset(0, -2),
          ),
        ],
      ),
      child: Row(
        children: [
          Icon(Icons.calculate, color: Colors.green.shade700, size: 20),
          const SizedBox(width: 12),
          Text(
            global.language('grand_total'),
            style: _titleStyle.copyWith(
              color: Colors.green.shade800,
              fontSize: 16,
            ),
          ),
          SizedBox(width: 24),
          Expanded(
            child: Wrap(
              spacing: 24,
              runSpacing: 8,
              children: [
                _buildTotalItem(global.language('enter_qty'), _grandTotalCount, format: '#,##0.###'),
                _buildTotalItem(global.language('payment_daily_total_amount_column'), _grandTotalAmount),
                _buildTotalItem(global.language('total_cost'), _grandTotalCost),
                _buildTotalItem(global.language('gross_profit'), _grandTotalGrossProfit),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTotalItem(String label, dynamic value, {String? format}) {
    final displayValue = value == null
        ? '-'
        : _formatNumber(value, pattern: format ?? '#,##0.00');

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 6),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.green.shade200),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            label,
            style: _bodyStyle.copyWith(
              color: Colors.green.shade700,
              fontWeight: FontWeight.w600,
            ),
          ),
          const SizedBox(width: 8),
          Text(
            displayValue,
            style: _titleStyle.copyWith(
              color: Colors.green.shade800,
              fontSize: 16,
              fontWeight: FontWeight.bold,
            ),
          ),
        ],
      ),
    );
  }

  double? _extractNumericValue(dynamic value) {
    if (value == null) return null;
    if (value is num) return value.toDouble();
    if (value is String) {
      final sanitized = value.replaceAll(',', '').trim();
      if (sanitized.isEmpty) return null;
      return double.tryParse(sanitized);
    }
    if (value is Map) {
      for (final candidate in ['sum', 'value', 'total', 'amount']) {
        if (value.containsKey(candidate)) {
          final nested = _extractNumericValue(value[candidate]);
          if (nested != null) return nested;
        }
      }
      for (final nested in value.values) {
        final resolved = _extractNumericValue(nested);
        if (resolved != null) return resolved;
      }
    }
    if (value is Iterable) {
      for (final item in value) {
        final resolved = _extractNumericValue(item);
        if (resolved != null) return resolved;
      }
    }
    return null;
  }

  void _onScroll() {
    // Check if scrolled near bottom and need to load more
    if (_scrollController.position.pixels >=
        _scrollController.position.maxScrollExtent - 200) {
      _loadMoreData();
    }
  }

  Future<void> _loadMoreData() async {
    if (_isLoadingMore || !_hasMoreData || _paginationInfo == null) {
      return;
    }

    setState(() => _isLoadingMore = true);

    try {
      final nextOffset = _resultData.length;
      await _loadResultPage(nextOffset ~/ _pageSize, append: true);
    } finally {
      if (mounted) {
        setState(() => _isLoadingMore = false);
      }
    }
  }

  String _formatNumber(dynamic value, {String? pattern}) {
    final numeric = _extractNumericValue(value);
    if (numeric == null) return value == null ? '' : value.toString();
    if (pattern == null || pattern.isEmpty) {
      return global.formatNumber(numeric);
    }
    return _resolveNumberFormat(pattern).format(numeric);
  }

  NumberFormat _resolveNumberFormat(String pattern) {
    return _numberFormatCache.putIfAbsent(
      pattern,
      () => NumberFormat(pattern, 'th_TH'),
    );
  }

  DateFormat _resolveDateFormat(String pattern) {
    return _dateFormatCache.putIfAbsent(
      pattern,
      () => DateFormat(pattern, 'th_TH'),
    );
  }

  Widget _buildPdfTab() {
    final hasPdf = _pdfBytes != null;
    final toolbarChildren = [
      Icon(Icons.picture_as_pdf, color: Colors.red.shade700),
      SizedBox(width: 8),
      Text(global.language('report_process_stock_cost'), style: _titleStyle),
      const Spacer(),
      DropdownButton<String>(
        value: _selectedOrientation,
        items: _orientationLabels.entries
            .map(
              (entry) =>
                  DropdownMenuItem(value: entry.key, child: Text(entry.value)),
            )
            .toList(),
        onChanged: (value) {
          if (value != null) {
            setState(() => _selectedOrientation = value);
          }
        },
        underline: const SizedBox.shrink(),
        style: _bodyStyle,
      ),
      const SizedBox(width: 12),
      DropdownButton<String>(
        value: _selectedPageSize,
        items: _pageSizeOptions
            .map(
              (size) =>
                  DropdownMenuItem(value: size, child: Text('ขนาด $size')),
            )
            .toList(),
        onChanged: (value) {
          if (value != null) {
            setState(() => _selectedPageSize = value);
          }
        },
        underline: const SizedBox.shrink(),
        style: _bodyStyle,
      ),
      SizedBox(width: 16),
      OutlinedButton.icon(
        onPressed: _currentGuid != null && !_isGeneratingPdf
            ? _generatePdf
            : null,
        icon: _isGeneratingPdf
            ? const SizedBox(
                width: 16,
                height: 16,
                child: CircularProgressIndicator(strokeWidth: 2),
              )
            : Icon(Icons.refresh),
        label: Text(_isGeneratingPdf ? global.language('generating') : global.language('import.action.created')),
      ),
      SizedBox(width: 8),
      ElevatedButton.icon(
        onPressed: hasPdf ? _downloadPdf : null,
        icon: Icon(Icons.download),
        label: Text(global.language('download')),
        style: ElevatedButton.styleFrom(
          backgroundColor: _primaryColor,
          foregroundColor: Colors.white,
        ),
      ),
      SizedBox(width: 8),
      ElevatedButton.icon(
        onPressed: hasPdf ? _printPdf : null,
        icon: Icon(Icons.print),
        label: Text(global.language('print')),
        style: ElevatedButton.styleFrom(
          backgroundColor: Colors.teal.shade600,
          foregroundColor: Colors.white,
        ),
      ),
    ];

    if (!hasPdf) {
      return Column(
        children: [
          Container(
            padding: const EdgeInsets.all(16),
            color: _secondaryColor,
            child: Row(children: toolbarChildren),
          ),
          Expanded(
            child: Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    Icons.picture_as_pdf,
                    size: 64,
                    color: Colors.grey.shade400,
                  ),
                  const SizedBox(height: 16),
                  Text(
                    'รายงาน PDF จะแสดงที่นี่',
                    style: TextStyle(fontSize: 18, color: Colors.grey.shade600),
                  ),
                  SizedBox(height: 8),
                  Text(
                    'กรุณาประมวลผลรายงานและกดปุ่ม global.language("create_pdf")',
                    style: TextStyle(fontSize: 14, color: Colors.grey.shade500),
                  ),
                  SizedBox(height: 24),
                  ElevatedButton.icon(
                    onPressed: _currentGuid != null && !_isGeneratingPdf
                        ? _generatePdf
                        : null,
                    icon: _isGeneratingPdf
                        ? const SizedBox(
                            width: 16,
                            height: 16,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : Icon(Icons.picture_as_pdf),
                    label: Text(global.language('create_pdf')),
                  ),
                ],
              ),
            ),
          ),
        ],
      );
    }

    return Column(
      children: [
        Container(
          padding: const EdgeInsets.all(16),
          color: _secondaryColor,
          child: Row(children: toolbarChildren),
        ),
        Expanded(
          child: SfPdfViewer.memory(
            _pdfBytes!,
            canShowScrollHead: true,
            canShowScrollStatus: true,
            enableDoubleTapZooming: true,
          ),
        ),
      ],
    );
  }

  /// ดาวน์โหลด PDF
  void _downloadPdf() {
    final bytes = _pdfBytes;
    if (bytes == null) return;

    final timestamp = DateTime.now();
    final filename =
        'sales-report-${timestamp.year}${timestamp.month.toString().padLeft(2, '0')}${timestamp.day.toString().padLeft(2, '0')}_${timestamp.hour.toString().padLeft(2, '0')}${timestamp.minute.toString().padLeft(2, '0')}.pdf';

    savePdf(Uint8List.fromList(bytes), filename, context);
  }

  Future<void> _printPdf() async {
    final bytes = _pdfBytes;
    if (bytes == null) return;

    await printing.Printing.layoutPdf(
      usePrinterSettings: true,
      dynamicLayout: true,
      onLayout: (pdf.PdfPageFormat format) async => bytes,
    );
  }

  Widget _buildDateRangePanel() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Card(
        elevation: 2,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        child: Padding(
          padding: EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Icon(Icons.calendar_month, color: _primaryColor),
                  const SizedBox(width: 8),
                  Text(global.language('analysis_date_range'), style: _titleStyle),
                ],
              ),
              const SizedBox(height: 12),
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: _buildQuickRangeSections(),
              ),
              SizedBox(height: 16),
              Row(
                children: [
                  Expanded(
                    child: CustomDatePicker(
                      labelText: global.language('start_date'),
                      initialDate: _fromDate,
                      // useBuddhistCalendar และ languageCode จะอ่านจากสาขาอัตโนมัติ
                      onDateSelected: (date) {
                        if (date != null) {
                          setState(() {
                            _fromDate = date;
                            _selectedQuickRangeKey = null;
                          });
                        }
                      },
                      decoration: _dateDecoration(global.language('kb_select_start_date')),
                    ),
                  ),
                  SizedBox(width: 12),
                  Expanded(
                    child: CustomDatePicker(
                      labelText: global.language('end_date'),
                      initialDate: _toDate,
                      // useBuddhistCalendar และ languageCode จะอ่านจากสาขาอัตโนมัติ
                      onDateSelected: (date) {
                        if (date != null) {
                          setState(() {
                            _toDate = date;
                            _selectedQuickRangeKey = null;
                          });
                        }
                      },
                      decoration: _dateDecoration(global.language('kb_select_end_date')),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  InputDecoration _dateDecoration(String placeholder) {
    return InputDecoration(
      hintText: placeholder,
      prefixIcon: Icon(Icons.event, color: _primaryColor),
      border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
      filled: true,
      fillColor: Colors.white,
    );
  }

  Widget _buildBranchAndCustomerPanel() {
    return Padding(
      padding: EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Card(
        elevation: 2,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        child: Padding(
          padding: EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Icon(Icons.store_mall_directory, color: _primaryColor),
                  const SizedBox(width: 8),
                  Text(global.language('analysis_dimension'), style: _titleStyle),
                ],
              ),
              SizedBox(height: 12),
              _buildSelectionTile(
                icon: Icons.apartment,
                title: 'สาขาที่ต้องการดู',
                subtitle: '',
                actionLabel: global.language('select_branch'),
                onTap: _openBranchSelector,
                customContent: _buildBranchSelectionWidget(),
              ),
              const SizedBox(height: 12),
              _buildSelectionTile(
                icon: Icons.handshake,
                title: global.language('related_customer_debtor'),
                subtitle: '',
                actionLabel: 'เลือกคู่ค้า',
                onTap: _openDebtorSelector,
                customContent: _buildCreditorSelectionWidget(),
              ),
              SizedBox(height: 12),
              _buildSelectionTile(
                icon: Icons.inventory_2,
                title: global.language('product'),
                subtitle: '',
                actionLabel: global.language('report_condition_select_barcode'),
                onTap: _openProductSelector,
                customContent: _buildProductSelectionWidget(),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildSelectionTile({
    required IconData icon,
    required String title,
    required String subtitle,
    required String actionLabel,
    required VoidCallback onTap,
    Widget? customContent,
  }) {
    return Container(
      padding: const EdgeInsets.all(14),
      decoration: BoxDecoration(
        color: Colors.grey.shade50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Row(
        children: [
          Container(
            padding: const EdgeInsets.all(10),
            decoration: BoxDecoration(
              color: _secondaryColor,
              borderRadius: BorderRadius.circular(10),
            ),
            child: Icon(icon, color: _primaryColor),
          ),
          const SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(title, style: _subtitleStyle),
                const SizedBox(height: 4),
                if (customContent != null)
                  customContent
                else
                  Text(subtitle, style: _bodyStyle),
              ],
            ),
          ),
          const SizedBox(width: 12),
          OutlinedButton(onPressed: onTap, child: Text(actionLabel)),
        ],
      ),
    );
  }

  /// สร้าง widget สำหรับแสดงรายการสินค้าที่เลือกแบบ wrap
  Widget _buildProductSelectionWidget() {
    if (_selectedProducts.isEmpty) {
      return Text('เลือกสินค้าตามเอกสาร (ทั้งหมด)', style: _bodyStyle);
    }

    return Wrap(
      spacing: 8,
      runSpacing: 6,
      children: [
        ..._selectedProducts.map((product) => _buildProductChip(product)),
      ],
    );
  }

  /// สร้าง chip สำหรับแต่ละสินค้า
  Widget _buildProductChip(SelectedProductItemStruct product) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: _secondaryColor,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _primaryColor.withValues(alpha: 0.3)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Flexible(
            child: Text(
              product.itemCode,
              style: _bodyStyle.copyWith(
                color: _primaryColor,
                fontSize: 11,
                fontWeight: FontWeight.w500,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          if (product.barcode.isNotEmpty) ...[
            const SizedBox(width: 4),
            Flexible(
              child: Text(
                '(${product.barcode})',
                style: _bodyStyle.copyWith(
                  color: Colors.grey.shade600,
                  fontSize: 10,
                ),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),
          ],
          const SizedBox(width: 4),
          Flexible(
            child: Text(
              product.itemName,
              style: _bodyStyle.copyWith(
                color: _primaryColor,
                fontSize: 11,
                fontWeight: FontWeight.w500,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          const SizedBox(width: 4),
          GestureDetector(
            onTap: () => _removeProduct(product),
            child: Icon(Icons.close, size: 14, color: Colors.grey.shade600),
          ),
        ],
      ),
    );
  }

  /// ลบสินค้าออกจากรายการ
  void _removeProduct(SelectedProductItemStruct product) {
    setState(() {
      _selectedProducts.removeWhere((p) => p.itemCode == product.itemCode);
    });
  }

  /// สร้าง widget สำหรับแสดงรายการสาขาที่เลือกแบบ wrap
  Widget _buildBranchSelectionWidget() {
    if (_selectedBranches.selectedBranches.isEmpty) {
      return Text(global.language('kb_all_branches'), style: _bodyStyle);
    }

    return Wrap(
      spacing: 8,
      runSpacing: 6,
      children: [
        ..._selectedBranches.selectedBranches.map(
          (branch) => _buildBranchChip(branch),
        ),
      ],
    );
  }

  /// สร้าง chip สำหรับแต่ละสาขา
  Widget _buildBranchChip(dynamic branch) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: Colors.blue.shade50,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.blue.shade300),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Flexible(
            child: Text(
              '${branch.code}',
              style: _bodyStyle.copyWith(
                color: Colors.blue.shade700,
                fontSize: 11,
                fontWeight: FontWeight.w500,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          const SizedBox(width: 4),
          Flexible(
            child: Text(
              branch.names.first.name,
              style: _bodyStyle.copyWith(
                color: Colors.blue.shade700,
                fontSize: 11,
                fontWeight: FontWeight.w500,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          const SizedBox(width: 4),
          GestureDetector(
            onTap: () => _removeBranch(branch),
            child: Icon(Icons.close, size: 14, color: Colors.grey.shade600),
          ),
        ],
      ),
    );
  }

  /// ลบสาขาออกจากรายการ
  void _removeBranch(dynamic branch) {
    setState(() {
      _selectedBranches.selectedBranches.removeWhere(
        (b) => b.code == branch.code,
      );
    });
  }

  /// สร้าง widget สำหรับแสดงรายการลูกค้า/ลูกหนี้ที่เลือกแบบ wrap (แบบสินค้า)
  Widget _buildCreditorSelectionWidget() {
    if (_selectedCreditors.selectedEntities.isEmpty) {
      return Text('เลือกคู่ค้า (ทั้งหมด)', style: _bodyStyle);
    }

    return Wrap(
      spacing: 8,
      runSpacing: 6,
      children: [
        ..._selectedCreditors.selectedEntities.map(
          (entity) => _buildCreditorChip(entity),
        ),
      ],
    );
  }

  /// สร้าง chip สำหรับแต่ละลูกค้า/ลูกหนี้ (แบบสินค้า)
  Widget _buildCreditorChip(dynamic entity) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      decoration: BoxDecoration(
        color: _secondaryColor,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: _primaryColor.withValues(alpha: 0.3)),
      ),
      child: Row(
        mainAxisSize: MainAxisSize.min,
        children: [
          Flexible(
            child: Text(
              entity.code,
              style: _bodyStyle.copyWith(
                color: _primaryColor,
                fontSize: 11,
                fontWeight: FontWeight.w500,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          const SizedBox(width: 4),
          Flexible(
            child: Text(
              entity.names.first.name,
              style: _bodyStyle.copyWith(
                color: _primaryColor,
                fontSize: 11,
                fontWeight: FontWeight.w500,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
          const SizedBox(width: 4),
          GestureDetector(
            onTap: () => _removeCreditor(entity),
            child: Icon(Icons.close, size: 14, color: Colors.grey.shade600),
          ),
        ],
      ),
    );
  }

  /// ลบลูกค้า/ลูกหนี้ออกจากรายการ
  void _removeCreditor(dynamic entity) {
    setState(() {
      _selectedCreditors.selectedEntities.removeWhere(
        (e) => e.code == entity.code,
      );
    });
  }

  Widget _buildSummaryPanel() {
    return Padding(
      padding: EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Card(
        elevation: 2,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        child: Padding(
          padding: EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Icon(Icons.receipt_long, color: _primaryColor),
                  const SizedBox(width: 8),
                  Text(global.language('selected_values_summary'), style: _titleStyle),
                ],
              ),
              SizedBox(height: 12),
              Wrap(
                spacing: 10,
                runSpacing: 10,
                children: [
                  _buildSummaryChip(
                    label: global.language('date'),
                    value:
                        '${_formatDate(_fromDate)} - ${_formatDate(_toDate)}',
                  ),
                  _buildSummaryChip(
                    label: global.language('branch'),
                    value: _selectedBranches.getBranchDisplayString(),
                  ),
                  _buildSummaryChip(
                    label: global.language('customer_debtor'),
                    value: _selectedCreditors.getEntityDisplayString(),
                  ),
                  _buildSummaryChip(
                    label: global.language('product'),
                    value: _selectedProducts.isEmpty
                        ? global.language('all')
                        : '${_selectedProducts.length} รายการ',
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildDetailOptionsPanel() {
    return Padding(
      padding: EdgeInsets.symmetric(horizontal: 16, vertical: 8),
      child: Card(
        elevation: 2,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        child: Padding(
          padding: EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Icon(Icons.tune, color: _primaryColor),
                  SizedBox(width: 8),
                  Text(global.language('display_options'), style: _titleStyle),
                ],
              ),
              const SizedBox(height: 12),
              SwitchListTile.adaptive(
                contentPadding: EdgeInsets.zero,
                value: _includeItemDetails,
                onChanged: (value) =>
                    setState(() => _includeItemDetails = value),
                title: Text(global.language('show_product_sub_items'), style: _subtitleStyle),
                subtitle: Text(
                  'เปิดเพื่อดึงข้อมูลสินค้าในแต่ละเอกสาร',
                  style: _bodyStyle.copyWith(fontSize: 12),
                ),
                activeTrackColor: _primaryColor,
              ),
              const Divider(height: 24),
              Text(global.language('sort_by_datetime'), style: _subtitleStyle),
              const SizedBox(height: 8),
              Row(
                children: [
                  Expanded(
                    child: ChoiceChip(
                      label: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(Icons.arrow_upward, size: 16),
                          const SizedBox(width: 4),
                          Text(global.language('sort_ascending')),
                        ],
                      ),
                      selected: _sortAscending,
                      onSelected: (_) => setState(() => _sortAscending = true),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Expanded(
                    child: ChoiceChip(
                      label: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(Icons.arrow_downward, size: 16),
                          const SizedBox(width: 4),
                          Text(global.language('sort_descending')),
                        ],
                      ),
                      selected: !_sortAscending,
                      onSelected: (_) => setState(() => _sortAscending = false),
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildSummaryChip({required String label, required String value}) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: _primaryColor.withValues(alpha: 0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(label, style: _bodyStyle.copyWith(color: Colors.grey.shade600)),
          const SizedBox(height: 4),
          Text(value, style: _subtitleStyle.copyWith(fontSize: 13)),
        ],
      ),
    );
  }

  Widget _buildActionButtons() {
    return Padding(
      padding: EdgeInsets.symmetric(horizontal: 16, vertical: 16),
      child: Row(
        children: [
          Expanded(
            child: ElevatedButton.icon(
              onPressed: _isProcessing ? null : _handleProcess,
              icon: _isProcessing
                  ? const SizedBox(
                      width: 16,
                      height: 16,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    )
                  : Icon(Icons.play_arrow),
              label: Text(global.language('เริ่มประมวลผล')),
              style: ElevatedButton.styleFrom(
                backgroundColor: _primaryColor,
                foregroundColor: Colors.white,
                padding: const EdgeInsets.symmetric(vertical: 14),
              ),
            ),
          ),
          SizedBox(width: 12),
          OutlinedButton.icon(
            onPressed: _resetAllConditions,
            icon: Icon(Icons.restore),
            label: Text(global.language('clear_conditions')),
          ),
        ],
      ),
    );
  }

  void _applyQuickRange(String key) {
    final rangeBuilder = _quickRanges[key];
    if (rangeBuilder == null) return;
    final range = rangeBuilder();
    setState(() {
      _selectedQuickRangeKey = key;
      _fromDate = range.start;
      _toDate = range.end;
    });
  }

  List<Widget> _buildQuickRangeSections() {
    final now = DateTime.now();

    final popularChips = <MapEntry<String, String>>[
      MapEntry('today', global.language('alert_today')),
      MapEntry('yesterday', 'เมื่อวานนี้'),
      MapEntry('thisWeek', global.language('alert_this_week')),
      MapEntry('lastWeek', 'สัปดาห์ก่อน'),
      MapEntry('thisMonth', global.language('alert_this_month')),
      const MapEntry('lastMonth', 'เดือนก่อน'),
    ];

    final monthChips = <MapEntry<String, String>>[];
    for (var i = 0; i < 12; i++) {
      final key = 'month_$i';
      final date = DateTime(now.year, now.month - i, 1);
      final monthName = global.getThaiMonthNameShort(date.month);
      final label = '$monthName ${(date.year + 543).toString().substring(2)}';
      monthChips.add(MapEntry(key, label));
    }

    final yearChips = <MapEntry<String, String>>[
      MapEntry('year_this', 'ปีนี้ (${now.year + 543})'),
      MapEntry('year_prev', 'ปีก่อน (${now.year - 1 + 543})'),
    ];
    for (var yearsAgo = 2; yearsAgo <= 6; yearsAgo++) {
      final year = now.year - yearsAgo;
      yearChips.add(MapEntry('year_history_$yearsAgo', 'พ.ศ. ${year + 543}'));
    }

    return [
      _buildQuickRangeSection('ช่วงยอดนิยม', popularChips),
      _buildQuickRangeSection('รายเดือน (ย้อนหลัง 12 เดือน)', monthChips),
      _buildQuickRangeSection('รายปี', yearChips),
    ];
  }

  Widget _buildQuickRangeSection(
    String title,
    List<MapEntry<String, String>> chips,
  ) {
    if (chips.isEmpty) return const SizedBox.shrink();
    return Padding(
      padding: const EdgeInsets.only(bottom: 12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(title, style: _subtitleStyle),
          const SizedBox(height: 8),
          Wrap(
            spacing: 8,
            runSpacing: 8,
            children: chips
                .map((entry) => _buildQuickChip(entry.key, entry.value))
                .toList(),
          ),
        ],
      ),
    );
  }

  Widget _buildQuickChip(String key, String label) {
    return ChoiceChip(
      label: Text(label),
      selected: _selectedQuickRangeKey == key,
      onSelected: (_) => _applyQuickRange(key),
    );
  }

  Future<void> _openBranchSelector() async {
    final result = await Navigator.push<BranchSelectionModel?>(
      context,
      MaterialPageRoute(
        builder: (_) => BlocProvider(
          create: (_) => CompanyBranchBloc(
            companyBranchRepository: CompanyBranchRepository(),
          ),
          child: MultiBranchSearchScreen(
            word: '',
            preSelectedBranches: _selectedBranches.selectedBranches,
          ),
        ),
      ),
    );

    if (result != null && !result.isCancel) {
      setState(() => _selectedBranches = result);
    }
  }

  Future<void> _openDebtorSelector() async {
    final result = await Navigator.push<EntitySelectionModel?>(
      context,
      MaterialPageRoute(
        builder: (_) => BlocProvider(
          create: (_) => DebtorBloc(debtorRepository: DebtorRepository()),
          child: MultiEntitySearchScreen(
            word: '',
            entityType: EntityType.debtor,
            preSelectedEntities: _selectedCreditors.selectedEntities,
          ),
        ),
      ),
    );

    if (result != null && !result.isCancel) {
      setState(() => _selectedCreditors = result);
    }
  }

  Future<void> _openProductSelector() async {
    final result = await Navigator.push<List<SelectedProductItemStruct>>(
      context,
      MaterialPageRoute(
        builder: (_) => SelectProductBarcodeWidget(
          selectedItems: List<SelectedProductItemStruct>.from(
            _selectedProducts,
          ),
        ),
      ),
    );

    if (result != null) {
      setState(() {
        _selectedProducts
          ..clear()
          ..addAll(result);
      });
    }
  }

  Future<void> _resetAllConditions() async {
    final now = DateTime.now();
    setState(() {
      _fromDate = DateTime(now.year, now.month, 1);
      _toDate = DateTime(now.year, now.month, now.day);
      _selectedQuickRangeKey = 'thisMonth';
      _selectedBranches = const BranchSelectionModel(
        selectedBranches: [],
        isCancel: false,
      );
      _selectedCreditors = const EntitySelectionModel(
        selectedEntities: [],
        isCancel: false,
      );
      _selectedProducts.clear();
    });
    _applyQuickRange('thisMonth');
  }

  Future<void> _handleProcess() async {
    if (_isProcessing) return;

    final shopid = _resolveShopId();
    if (shopid == null) {
      return;
    }

    setState(() => _isProcessing = true);

    try {
      final branchCodes = _selectedBranches.selectedBranches
          .map((b) => b.code)
          .whereType<String>()
          .toList();

      final productCodes = _selectedProducts
          .map((p) => p.itemCode)
          .whereType<String>()
          .toList();

      final fromDateStr = DateFormat('yyyy-MM-dd').format(_fromDate);
      final toDateStr = DateFormat('yyyy-MM-dd').format(_toDate);

      // 1. ดึงข้อมูล header จาก Backend API (parameterized queries, ปลอดภัยจาก SQL injection)
      AppLogger.info('[SaleProfitByDoc] Calling SalesReportApiService.getSalesReportHeader');
      final headerResult = await SalesReportApiService.getSalesReportHeader(
        fromDate: fromDateStr,
        toDate: toDateStr,
        branchCodes: branchCodes.isEmpty ? null : branchCodes,
        productCodes: productCodes.isEmpty ? null : productCodes,
        sortAscending: _sortAscending,
        limit: 10000,
        offset: 0,
      );

      if (!headerResult.isSuccess) {
        throw Exception(headerResult.errorMessage ?? 'API call failed');
      }
      AppLogger.info('[SaleProfitByDoc] API returned ${headerResult.count} header rows');

      // 2. ดึงข้อมูล detail (ถ้าต้องการ)
      List<Map<String, dynamic>> detailData = [];
      if (_includeItemDetails) {
        final detailResult = await SalesReportApiService.getSalesReportDetail(
          fromDate: fromDateStr,
          toDate: toDateStr,
          branchCodes: branchCodes.isEmpty ? null : branchCodes,
          productCodes: productCodes.isEmpty ? null : productCodes,
          sortAscending: _sortAscending,
          limit: 50000,
          offset: 0,
        );
        if (detailResult.isSuccess) {
          detailData = detailResult.data;
          AppLogger.info('[SaleProfitByDoc] API returned ${detailResult.count} detail rows');
        }
      }

      // 3. ดึงยอดรวม
      _grandTotalCount = 0;
      _grandTotalAmount = 0;
      _grandTotalCost = 0;
      _grandTotalGrossProfit = 0;

      final summary = await SalesReportApiService.getSalesSummary(
        fromDate: fromDateStr,
        toDate: toDateStr,
        branchCodes: branchCodes.isEmpty ? null : branchCodes,
        productCodes: productCodes.isEmpty ? null : productCodes,
      );

      if (summary.isSuccess) {
        _grandTotalCount = summary.totalDocuments.toDouble();
        _grandTotalAmount = summary.totalAmount;
        _grandTotalCost = summary.totalCost;
        _grandTotalGrossProfit = summary.totalProfit;
      }

      // 4. รวม header + detail เป็น flat list พร้อม level field สำหรับ UI
      final List<Map<String, dynamic>> mergedData = [];
      for (final header in headerResult.data) {
        mergedData.add({...header, 'level': 0});
        if (_includeItemDetails) {
          final docno = header['docno'];
          final details = detailData.where((d) => d['docno'] == docno);
          for (final detail in details) {
            mergedData.add({...detail, 'level': 1});
          }
        }
      }

      setState(() {
        _currentGuid = null; // Direct API mode — ไม่ใช้ result table
        _resultData = mergedData;
        _paginationInfo = null;
        _pdfBytes = null;
        _hasMoreData = false; // โหลดข้อมูลทั้งหมดแล้ว
      });

      // เปลี่ยนไปที่ Result tab
      _tabController.animateTo(1);

      if (!mounted) return;
      global.showSnackBar(
        context,
        const Icon(Icons.check_circle, color: Colors.white),
        'ประมวลผลสำเร็จ: ${headerResult.count} รายการ',
        Colors.green.shade700,
      );
    } catch (e) {
      AppLogger.error('_handleProcess error: $e');
      if (!mounted) return;
      global.showSnackBar(
        context,
        const Icon(Icons.error, color: Colors.white),
        _resolveFriendlyError(e),
        Colors.red.shade700,
      );
    } finally {
      if (mounted) {
        setState(() => _isProcessing = false);
      }
    }
  }

  /// โหลดข้อมูล result ตามหน้าที่ระบุ
  Future<void> _loadResultPage(int page, {bool append = false}) async {
    if (_currentGuid == null) return;

    final shopid = _resolveShopId();
    if (shopid == null) {
      return;
    }

    if (!append) {
      setState(() => _isLoadingResults = true);
    }

    try {
      final offset = page * _pageSize;
      AppLogger.info(
        '[ReportSalesByDocument] /resultget request | shopid=$shopid, guid=${_currentGuid!}, page=$page, limit=$_pageSize, offset=$offset, append=$append',
      );

      final response = await ResultTableApiService.getResults(
        shopid: shopid,
        guid: _currentGuid!,
        limit: _pageSize,
        offset: offset,
      );

      if (!mounted) return;

      final fetchedRows = List<Map<String, dynamic>>.from(response.data);
      final hasMoreAfterThisPage = fetchedRows.length >= _pageSize;

      setState(() {
        if (append) {
          _resultData.addAll(fetchedRows);
        } else {
          _resultData = fetchedRows;
        }

        _paginationInfo = response.pagination;
        _hasMoreData = hasMoreAfterThisPage;
      });
    } catch (e) {
      AppLogger.error('_loadResultPage error: $e');
      if (!mounted) return;
      global.showSnackBar(
        context,
        const Icon(Icons.error, color: Colors.white),
        _resolveFriendlyError(e),
        Colors.red.shade700,
      );
    } finally {
      if (mounted && !append) {
        setState(() => _isLoadingResults = false);
      }
    }
  }

  /// สร้าง PDF
  Future<void> _generatePdf() async {
    if (_resultData.isEmpty) {
      global.showSnackBar(
        context,
        const Icon(Icons.warning, color: Colors.white),
        'กรุณาประมวลผลข้อมูลก่อน',
        Colors.orange.shade700,
      );
      return;
    }

    // Direct API mode — PDF ต้องใช้ result table guid
    if (_currentGuid == null) {
      global.showSnackBar(
        context,
        const Icon(Icons.info, color: Colors.white),
        'PDF ยังไม่รองรับในโหมด API โดยตรง — กำลังพัฒนา',
        Colors.orange.shade700,
      );
      return;
    }

    final shopid = _resolveShopId();
    if (shopid == null) {
      return;
    }

    setState(() => _isGeneratingPdf = true);

    try {
      AppLogger.info(
        'Calling endpoint: ${ResultTableApiService.baseUrl}/resulttopdf',
      );

      final showDetails = _includeItemDetails;
      final orderedKeys = LinkedHashSet<String>();
      void addColumnKeys(List<_ReportColumn> columns) {
        for (final column in columns) {
          orderedKeys.add(column.key);
        }
      }

      addColumnKeys(_documentColumns);
      if (showDetails) {
        addColumnKeys(_detailColumns);
      }

      final columnOrder = orderedKeys.toList();
      final columnNames = <String, String>{
        for (final column in _documentColumns) column.key: column.label,
      };
      if (showDetails) {
        for (final column in _detailColumns) {
          columnNames.putIfAbsent(column.key, () => column.label);
        }
      }

      final layoutSections = [
        {
          'alias': _headerAlias.value,
          'title': global.language('document_info'),
          'row_type': 'header',
          'columns': _buildPdfColumns(_documentColumns),
        },
        {
          'alias': _detailAlias.value,
          'title': global.language('product_detail'),
          'row_type': 'detail',
          'visible': showDetails,
          'columns': _buildPdfColumns(_detailColumns),
        },
      ];

      final pdfConfig = PdfConfig(
        title: global.language('report_process_stock_cost'),
        orientation: _selectedOrientation,
        pageSize: _selectedPageSize,
      );

      final layoutConfig = <String, dynamic>{
        'schema_version': 2,
        'sections': layoutSections,
        'column_schema': _buildColumnSchema(showDetails),
        'styles': _buildPdfStyleGuide(),
      };

      final numberFormats = _buildNumberFormats();
      if (numberFormats.isNotEmpty) {
        layoutConfig['number_formats'] = numberFormats;
      }

      final pdfBytes = await ResultTableApiService.generatePdf(
        shopid: shopid,
        guid: _currentGuid!,
        pdfConfig: pdfConfig,
        columnOrder: columnOrder,
        columnNames: columnNames,
        layoutConfig: layoutConfig,
      );

      if (!mounted) return;
      setState(() {
        _pdfBytes = Uint8List.fromList(pdfBytes);
      });

      // เปลี่ยนไปที่ PDF tab
      _tabController.animateTo(2);

      global.showSnackBar(
        context,
        const Icon(Icons.check_circle, color: Colors.white),
        'สร้าง PDF สำเร็จ',
        Colors.green.shade700,
      );
    } catch (e) {
      AppLogger.error('_generatePdf error: $e');
      if (!mounted) return;
      global.showSnackBar(
        context,
        const Icon(Icons.error, color: Colors.white),
        _resolveFriendlyError(e),
        Colors.red.shade700,
      );
    } finally {
      if (mounted) {
        setState(() => _isGeneratingPdf = false);
      }
    }
  }

  String? _resolveShopId() {
    final shopId = global.getShopId();
    if (shopId.isEmpty) {
      AppLogger.warn('ไม่พบ shopId ใน global appConfig');
      if (mounted) {
        global.showSnackBar(
          context,
          const Icon(Icons.warning, color: Colors.white),
          'ไม่พบรหัสร้านค้า กรุณาเลือกร้านค้าก่อนใช้งานรายงาน',
          Colors.orange.shade700,
        );
      }
      return null;
    }
    return shopId;
  }

  String _resolveFriendlyError(Object error) {
    if (error is SocketException) {
      return 'ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้ กรุณาตรวจสอบเครือข่าย';
    }
    final raw = error.toString();
    final lower = raw.toLowerCase();
    if (lower.contains('503')) {
      return 'ระบบรายงานไม่พร้อมให้บริการชั่วคราว (503) กรุณาลองใหม่ภายหลัง';
    }
    if (lower.contains('timeout')) {
      return 'การเชื่อมต่อหมดเวลา กรุณาลองใหม่อีกครั้ง';
    }
    return 'เกิดข้อผิดพลาด: $raw';
  }

  String _formatDate(DateTime date) {
    return '${date.day.toString().padLeft(2, '0')}/'
        '${date.month.toString().padLeft(2, '0')}/'
        '${date.year + 543}';
  }

  DateTime _normalize(DateTime date) =>
      DateTime(date.year, date.month, date.day);
}

enum _ReportColumnType { text, number, date, time }

extension _ReportColumnTypeWire on _ReportColumnType {
  String get wireValue {
    switch (this) {
      case _ReportColumnType.text:
        return 'text';
      case _ReportColumnType.number:
        return 'number';
      case _ReportColumnType.date:
        return 'date';
      case _ReportColumnType.time:
        return 'time';
    }
  }
}

class _ReportColumn {
  final String key;
  final String label;
  final int flex;
  final TextAlign align;
  final bool isNumeric;
  final bool hideWhenSummary;
  final _ReportColumnType dataType;
  final String? format;
  final bool useBuddhistYear;

  const _ReportColumn({
    required this.key,
    required this.label,
    required this.flex,
    this.align = TextAlign.left,
    bool? isNumeric,
    this.hideWhenSummary = false,
    this.dataType = _ReportColumnType.text,
    this.format,
    this.useBuddhistYear = false,
  }) : isNumeric = isNumeric ?? dataType == _ReportColumnType.number;
}

/// Query Builder สำหรับรายงานนี้โดยเฉพาะ
// Routine classes moved to lib/utils/report_builder.
