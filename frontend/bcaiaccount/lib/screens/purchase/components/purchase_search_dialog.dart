import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/widgets/date_picker_widget.dart';

/// Model สำหรับเก็บเงื่อนไขการค้นหาใบสั่งซื้อ
class PurchaseSearchCondition {
  DateTime? fromDate;
  DateTime? toDate;
  double? minAmount;
  double? maxAmount;
  String searchText;
  String docNo;
  String creditorCode;
  String creditorName;

  PurchaseSearchCondition({
    this.fromDate,
    this.toDate,
    this.minAmount,
    this.maxAmount,
    this.searchText = '',
    this.docNo = '',
    this.creditorCode = '',
    this.creditorName = '',
  });

  /// สร้าง copy ของ condition
  PurchaseSearchCondition copyWith({
    DateTime? fromDate,
    DateTime? toDate,
    double? minAmount,
    double? maxAmount,
    String? searchText,
    String? docNo,
    String? creditorCode,
    String? creditorName,
  }) {
    return PurchaseSearchCondition(
      fromDate: fromDate ?? this.fromDate,
      toDate: toDate ?? this.toDate,
      minAmount: minAmount ?? this.minAmount,
      maxAmount: maxAmount ?? this.maxAmount,
      searchText: searchText ?? this.searchText,
      docNo: docNo ?? this.docNo,
      creditorCode: creditorCode ?? this.creditorCode,
      creditorName: creditorName ?? this.creditorName,
    );
  }

  /// ล้างเงื่อนไขทั้งหมด
  void clear() {
    fromDate = null;
    toDate = null;
    minAmount = null;
    maxAmount = null;
    searchText = '';
    docNo = '';
    creditorCode = '';
    creditorName = '';
  }

  /// ตรวจสอบว่ามีเงื่อนไขหรือไม่
  bool get hasConditions {
    return fromDate != null ||
        toDate != null ||
        minAmount != null ||
        maxAmount != null ||
        searchText.isNotEmpty ||
        docNo.isNotEmpty ||
        creditorCode.isNotEmpty ||
        creditorName.isNotEmpty;
  }
}

/// Dialog สำหรับตั้งค่าเงื่อนไขการค้นหาใบสั่งซื้อ
class PurchaseSearchDialog extends StatefulWidget {
  final PurchaseSearchCondition initialCondition;
  final Function(PurchaseSearchCondition) onSearch;

  const PurchaseSearchDialog({
    super.key,
    required this.initialCondition,
    required this.onSearch,
  });

  /// แสดง dialog และรับผลลัพธ์
  static Future<PurchaseSearchCondition?> show(
    BuildContext context, {
    required PurchaseSearchCondition initialCondition,
    required Function(PurchaseSearchCondition) onSearch,
  }) {
    return showDialog<PurchaseSearchCondition>(
      context: context,
      builder: (context) => PurchaseSearchDialog(
        initialCondition: initialCondition,
        onSearch: onSearch,
      ),
    );
  }

  @override
  State<PurchaseSearchDialog> createState() => _PurchaseSearchDialogState();
}

class _PurchaseSearchDialogState extends State<PurchaseSearchDialog> {
  late PurchaseSearchCondition _condition;

  // Controllers สำหรับ text fields
  late TextEditingController _searchTextController;
  late TextEditingController _docNoController;
  late TextEditingController _creditorCodeController;
  late TextEditingController _creditorNameController;
  late TextEditingController _minAmountController;
  late TextEditingController _maxAmountController;

  @override
  void initState() {
    super.initState();
    // สร้าง copy ของ condition เพื่อไม่ให้กระทบต้นฉบับ
    _condition = widget.initialCondition.copyWith();

    // สร้าง controllers
    _searchTextController = TextEditingController(text: _condition.searchText);
    _docNoController = TextEditingController(text: _condition.docNo);
    _creditorCodeController = TextEditingController(text: _condition.creditorCode);
    _creditorNameController = TextEditingController(text: _condition.creditorName);
    _minAmountController = TextEditingController(
      text: _condition.minAmount?.toStringAsFixed(2) ?? '',
    );
    _maxAmountController = TextEditingController(
      text: _condition.maxAmount?.toStringAsFixed(2) ?? '',
    );
  }

  @override
  void dispose() {
    _searchTextController.dispose();
    _docNoController.dispose();
    _creditorCodeController.dispose();
    _creditorNameController.dispose();
    _minAmountController.dispose();
    _maxAmountController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final size = MediaQuery.of(context).size;
    final isWideScreen = size.width > 600;
    final dialogWidth = isWideScreen ? 500.0 : size.width * 0.9;

    return Dialog(
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      child: Container(
        width: dialogWidth,
        constraints: BoxConstraints(maxHeight: size.height * 0.85),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            // Header
            _buildHeader(),

            // Content - Scrollable
            Flexible(
              child: SingleChildScrollView(
                padding: EdgeInsets.all(16),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // ช่วงวันที่
                    _buildSectionTitle(global.language('date_range'), Icons.date_range),
                    const SizedBox(height: 8),
                    _buildDateRangeSection(),

                    const SizedBox(height: 20),

                    // ช่วงจำนวนเงิน
                    _buildSectionTitle(global.language('amount_range'), Icons.attach_money),
                    const SizedBox(height: 8),
                    _buildAmountRangeSection(),

                    const SizedBox(height: 20),

                    // ค้นหาเอกสาร
                    _buildSectionTitle(global.language('search_document'), Icons.description),
                    const SizedBox(height: 8),
                    _buildDocumentSearchSection(),

                    const SizedBox(height: 20),

                    // ค้นหาผู้ขาย/เจ้าหนี้
                    _buildSectionTitle(global.language('search_seller_creditor'), Icons.business),
                    const SizedBox(height: 8),
                    _buildCreditorSearchSection(),

                    const SizedBox(height: 20),

                    // Full-text search
                    _buildSectionTitle(global.language('fulltext_search'), Icons.search),
                    const SizedBox(height: 8),
                    _buildFullTextSearchSection(),
                  ],
                ),
              ),
            ),

            // Footer buttons
            _buildFooter(),
          ],
        ),
      ),
    );
  }

  /// สร้าง header ของ dialog
  Widget _buildHeader() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.blue.shade600,
        borderRadius: const BorderRadius.only(
          topLeft: Radius.circular(16),
          topRight: Radius.circular(16),
        ),
      ),
      child: Row(
        children: [
          const Icon(Icons.filter_list, color: Colors.white),
          SizedBox(width: 12),
          Text(
            global.language('search_condition'),
            style: const TextStyle(
              color: Colors.white,
              fontSize: 18,
              fontWeight: FontWeight.bold,
            ),
          ),
          const Spacer(),
          IconButton(
            icon: const Icon(Icons.close, color: Colors.white),
            onPressed: () => Navigator.pop(context),
          ),
        ],
      ),
    );
  }

  /// สร้าง section title
  Widget _buildSectionTitle(String title, IconData icon) {
    return Row(
      children: [
        Container(
          padding: const EdgeInsets.all(6),
          decoration: BoxDecoration(
            color: Colors.blue.shade50,
            borderRadius: BorderRadius.circular(8),
          ),
          child: Icon(icon, size: 18, color: Colors.blue.shade600),
        ),
        const SizedBox(width: 8),
        Text(
          title,
          style: TextStyle(
            fontSize: 14,
            fontWeight: FontWeight.w600,
            color: Colors.grey.shade800,
          ),
        ),
      ],
    );
  }

  /// สร้าง section เลือกช่วงวันที่
  Widget _buildDateRangeSection() {
    return Row(
      children: [
        Expanded(
          child: DatePickerWidget(
            label: global.language('from_date'),
            selectedDate: _condition.fromDate,
            onDateSelected: (date) {
              setState(() => _condition.fromDate = date);
            },
          ),
        ),
        const SizedBox(width: 12),
        const Text('-', style: TextStyle(fontSize: 20)),
        SizedBox(width: 12),
        Expanded(
          child: DatePickerWidget(
            label: global.language('to_date'),
            selectedDate: _condition.toDate,
            onDateSelected: (date) {
              setState(() => _condition.toDate = date);
            },
          ),
        ),
      ],
    );
  }

  /// สร้าง section ช่วงจำนวนเงิน
  Widget _buildAmountRangeSection() {
    return Row(
      children: [
        Expanded(
          child: TextField(
            controller: _minAmountController,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            decoration: InputDecoration(
              labelText: global.language('min_amount'),
              hintText: '0.00',
              prefixIcon: const Icon(Icons.money),
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(8),
              ),
              contentPadding: const EdgeInsets.symmetric(
                horizontal: 12,
                vertical: 12,
              ),
            ),
            onChanged: (value) {
              _condition.minAmount = double.tryParse(value);
            },
          ),
        ),
        const SizedBox(width: 12),
        const Text('-', style: TextStyle(fontSize: 20)),
        const SizedBox(width: 12),
        Expanded(
          child: TextField(
            controller: _maxAmountController,
            keyboardType: const TextInputType.numberWithOptions(decimal: true),
            decoration: InputDecoration(
              labelText: global.language('max_amount'),
              hintText: '999,999.00',
              prefixIcon: const Icon(Icons.money),
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(8),
              ),
              contentPadding: const EdgeInsets.symmetric(
                horizontal: 12,
                vertical: 12,
              ),
            ),
            onChanged: (value) {
              _condition.maxAmount = double.tryParse(value);
            },
          ),
        ),
      ],
    );
  }

  /// สร้าง section ค้นหาเอกสาร
  Widget _buildDocumentSearchSection() {
    return TextField(
      controller: _docNoController,
      decoration: InputDecoration(
        labelText: global.language('docno'),
        hintText: 'PO-2024-001',
        prefixIcon: const Icon(Icons.receipt_long),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
        ),
        contentPadding: const EdgeInsets.symmetric(
          horizontal: 12,
          vertical: 12,
        ),
      ),
      onChanged: (value) {
        _condition.docNo = value;
      },
    );
  }

  /// สร้าง section ค้นหาผู้ขาย/เจ้าหนี้
  Widget _buildCreditorSearchSection() {
    return Column(
      children: [
        TextField(
          controller: _creditorCodeController,
          decoration: InputDecoration(
            labelText: global.language('seller_code'),
            hintText: 'C001',
            prefixIcon: const Icon(Icons.badge),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(8),
            ),
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 12,
              vertical: 12,
            ),
          ),
          onChanged: (value) {
            _condition.creditorCode = value;
          },
        ),
        const SizedBox(height: 12),
        TextField(
          controller: _creditorNameController,
          decoration: InputDecoration(
            labelText: global.language('seller_name'),
            hintText: 'บริษัท ABC จำกัด',
            prefixIcon: const Icon(Icons.business),
            border: OutlineInputBorder(
              borderRadius: BorderRadius.circular(8),
            ),
            contentPadding: const EdgeInsets.symmetric(
              horizontal: 12,
              vertical: 12,
            ),
          ),
          onChanged: (value) {
            _condition.creditorName = value;
          },
        ),
      ],
    );
  }

  /// สร้าง section Full-text search
  Widget _buildFullTextSearchSection() {
    return TextField(
      controller: _searchTextController,
      decoration: InputDecoration(
        labelText: global.language('search_all_fields'),
        hintText: 'พิมพ์คำค้นหาภาษาไทยหรืออังกฤษ...',
        prefixIcon: const Icon(Icons.search),
        border: OutlineInputBorder(
          borderRadius: BorderRadius.circular(8),
        ),
        contentPadding: const EdgeInsets.symmetric(
          horizontal: 12,
          vertical: 12,
        ),
        helperText: 'รองรับการค้นหาภาษาไทย เช่น ชื่อสินค้า, หมายเหตุ',
        helperStyle: TextStyle(color: Colors.grey.shade600),
      ),
      onChanged: (value) {
        _condition.searchText = value;
      },
    );
  }

  /// สร้าง footer buttons
  Widget _buildFooter() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.grey.shade100,
        borderRadius: const BorderRadius.only(
          bottomLeft: Radius.circular(16),
          bottomRight: Radius.circular(16),
        ),
      ),
      child: Row(
        children: [
          // ปุ่มล้างเงื่อนไข
          OutlinedButton.icon(
            onPressed: _clearConditions,
            icon: Icon(Icons.clear_all),
            label: Text(global.language('clear_conditions')),
            style: OutlinedButton.styleFrom(
              foregroundColor: Colors.red.shade600,
              side: BorderSide(color: Colors.red.shade600),
            ),
          ),
          const Spacer(),
          // ปุ่มยกเลิก
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text(global.language('cancel')),
          ),
          const SizedBox(width: 12),
          // ปุ่มค้นหา
          ElevatedButton.icon(
            onPressed: _doSearch,
            icon: Icon(Icons.search),
            label: Text(global.language('search')),
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.blue.shade600,
              foregroundColor: Colors.white,
              padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
            ),
          ),
        ],
      ),
    );
  }

  /// ล้างเงื่อนไขทั้งหมด
  void _clearConditions() {
    setState(() {
      _condition.clear();
      _searchTextController.clear();
      _docNoController.clear();
      _creditorCodeController.clear();
      _creditorNameController.clear();
      _minAmountController.clear();
      _maxAmountController.clear();
    });
  }

  /// ดำเนินการค้นหา
  void _doSearch() {
    widget.onSearch(_condition);
    Navigator.pop(context, _condition);
  }
}
