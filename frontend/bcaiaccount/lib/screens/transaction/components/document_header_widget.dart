import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/model/customer_model.dart';
import 'package:smlaicloud/model/employee_model.dart';
import 'package:smlaicloud/model/approval_model.dart';
import 'package:smlaicloud/repositories/customer_repository.dart';
import 'package:smlaicloud/repositories/employee_repository.dart';
import 'package:smlaicloud/screens/transaction/components/document_total_widget.dart';
import 'package:smlaicloud/screens/transaction/transaction_edit_product_list_widget.dart';
import 'package:smlaicloud/screens/transaction/components/doc_flow_widget.dart';
import 'package:smlaicloud/utils/date_picker.dart';
import 'package:smlaicloud/utils/time_picker.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/screen_search/company_branch_search_screen.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:intl/intl.dart';
import 'package:smlaicloud/screens/transaction/components/document_references_widget.dart';
import 'package:smlaicloud/utils/textfield_custom.dart';
import 'package:smlaicloud/widgets/document_creator_info_widget.dart';
import 'package:smlaicloud/model/currency_model.dart';
import 'package:smlaicloud/components/numpad.dart';
import 'package:smlaicloud/screens/transaction/components/wht_section_widget.dart';
import 'package:smlaicloud/screens/transaction/components/credit_terms_section_widget.dart';

class DocumentHeaderWidget extends StatelessWidget {
  final TransactionModel screenData;
  final global.TransactionTypeEnum transactionType;
  final Function(void Function()) setState;
  final BuildContext context;
  final TextEditingController custCodeController;
  final TextEditingController custnamesController;
  final TextEditingController saleCodeController;
  final TextEditingController saleNameController;
  final TextEditingController docRefNumberController;
  final TextEditingController taxDocNoController;
  final TextEditingController vatRateController;
  final TextEditingController descriptionController;
  final TextEditingController transportAmountController;
  final TextEditingController docDateController;
  final TextEditingController docTimeController;
  final bool docDateTimeValidated;
  final global.Debouncer debouncer;
  final Function() calTotalValue;
  final Function() headerTableDetail;
  final Function({required String word}) searchCustomer;
  final Function({required String word}) searchSupplier;
  final Function({required String word}) searchSale;
  final Function({required String word}) searchSaleChannel;
  final Function({required String word}) searchDocRef;
  final Function(int index, String docno)? onDeleteDocReference;

  // ประเภทการจัดซื้อ (สำหรับ PO)
  final List<PurchaseTypeModel>? purchaseTypes;
  final String? selectedPurchaseTypeCode;
  final Function(String? code, String? name)? onPurchaseTypeChanged;

  // สกุลเงิน (สำหรับ Multi-Currency)
  final List<CurrencyModel>? currencies;
  final String? selectedDocCurrencyCode;
  final String? selectedDocCurrencySymbol;
  final double? selectedExchangeRate;
  final Function(String? code, String? symbol, double? exchangeRate)? onCurrencyChanged;
  final Function(double? exchangeRate)? onExchangeRateChanged; // แก้ไขอัตราแลกเปลี่ยนเอง
  
  // สกุลเงินหลัก (Base Currency สำหรับลงบัญชี)
  final String? baseCurrency;
  final String? baseCurrencySymbol;

  const DocumentHeaderWidget({
    super.key,
    required this.screenData,
    required this.transactionType,
    required this.setState,
    required this.context,
    required this.custCodeController,
    required this.custnamesController,
    required this.saleCodeController,
    required this.saleNameController,
    required this.docRefNumberController,
    required this.taxDocNoController,
    required this.vatRateController,
    required this.descriptionController,
    required this.transportAmountController,
    required this.docDateController,
    required this.docTimeController,
    required this.docDateTimeValidated,
    required this.debouncer,
    required this.calTotalValue,
    required this.headerTableDetail,
    required this.searchCustomer,
    required this.searchSupplier,
    required this.searchSale,
    required this.searchSaleChannel,
    required this.searchDocRef,
    this.onDeleteDocReference,
    this.purchaseTypes,
    this.selectedPurchaseTypeCode,
    this.onPurchaseTypeChanged,
    this.currencies,
    this.selectedDocCurrencyCode,
    this.selectedDocCurrencySymbol,
    this.selectedExchangeRate,
    this.onCurrencyChanged,
    this.onExchangeRateChanged,
    this.baseCurrency,
    this.baseCurrencySymbol,
  });

  // 🎨 Modern Color Palette
  static const Color _primaryColor = Color(0xFF667eea);
  static const Color _accentColor = Color(0xFF764ba2);

  /// 🎨 Helper method สำหรับสร้าง Section Card ที่มี icon, title และ child widget
  Widget _buildSectionCard({required IconData icon, required String title, required Widget child}) {
    return Container(
      margin: const EdgeInsets.only(bottom: 12),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(12),
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.06), blurRadius: 8, offset: const Offset(0, 2))],
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 14, vertical: 10),
            decoration: BoxDecoration(
              gradient: LinearGradient(colors: [_primaryColor.withValues(alpha: 0.1), _accentColor.withValues(alpha: 0.05)], begin: Alignment.centerLeft, end: Alignment.centerRight),
              borderRadius: const BorderRadius.only(topLeft: Radius.circular(12), topRight: Radius.circular(12)),
            ),
            child: Row(
              children: [
                Icon(icon, color: _primaryColor, size: 20),
                const SizedBox(width: 8),
                Text(
                  title,
                  style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: _primaryColor),
                ),
              ],
            ),
          ),
          // Content
          Padding(padding: const EdgeInsets.all(14), child: child),
        ],
      ),
    );
  }

  /// 🎨 Helper method สำหรับสร้าง InputDecoration ที่มี style เดียวกัน
  InputDecoration _buildInputDecoration({required String labelText, Widget? suffixIcon, Widget? prefixIcon, bool readOnly = false, bool hasError = false}) {
    return InputDecoration(
      filled: true,
      fillColor: readOnly ? Colors.grey.shade50 : Colors.white,
      floatingLabelBehavior: FloatingLabelBehavior.always,
      contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 14),
      labelText: labelText,
      labelStyle: TextStyle(color: hasError ? Colors.red : Colors.black87, fontWeight: FontWeight.bold, fontSize: 14),
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(10),
        borderSide: BorderSide(color: hasError ? Colors.red : Colors.grey.shade300),
      ),
      enabledBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(10),
        borderSide: BorderSide(color: hasError ? Colors.red : Colors.grey.shade300),
      ),
      focusedBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(10),
        borderSide: BorderSide(color: hasError ? Colors.red : _primaryColor, width: 1.5),
      ),
      suffixIcon: suffixIcon,
      prefixIcon: prefixIcon,
    );
  }

  /// 🎨 Helper method สำหรับสร้าง InputDecoration สำหรับ Radio/Dropdown groups
  InputDecoration _buildGroupDecoration({required String labelText}) {
    return InputDecoration(
      filled: true,
      fillColor: Colors.white,
      floatingLabelBehavior: FloatingLabelBehavior.always,
      contentPadding: const EdgeInsets.only(left: 12.0, top: 0, bottom: 0, right: 12.0),
      labelText: labelText,
      labelStyle: const TextStyle(color: Colors.black87, fontWeight: FontWeight.bold, fontSize: 14),
      border: OutlineInputBorder(
        borderRadius: BorderRadius.circular(10),
        borderSide: BorderSide(color: Colors.grey.shade300),
      ),
      enabledBorder: OutlineInputBorder(
        borderRadius: BorderRadius.circular(10),
        borderSide: BorderSide(color: Colors.grey.shade300),
      ),
    );
  }

  /// 🏷️ Helper method สำหรับสร้าง Section Header (ไม่ใช่ Card)
  /// ปิดการแสดง section header เพื่อประหยัดพื้นที่หน้าจอ
  Widget _buildSectionHeader({required IconData icon, required String title}) {
    // ไม่แสดง section header เพื่อประหยัดพื้นที่
    return const SizedBox.shrink();
  }

  /// 📦 สร้าง Dropdown สำหรับเลือกประเภทการจัดซื้อ (PO)
  Widget _buildPurchaseTypeDropdown() {
    // ตรวจสอบว่า selectedPurchaseTypeCode มีอยู่ใน purchaseTypes หรือไม่
    final isValidSelection = selectedPurchaseTypeCode != null &&
        purchaseTypes != null &&
        purchaseTypes!.any((t) => t.code == selectedPurchaseTypeCode);

    return Container(
      // ใช้ key เพื่อบังคับ rebuild เมื่อค่าเปลี่ยน
      key: ValueKey('purchase_type_dropdown_$selectedPurchaseTypeCode'),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        border: Border.all(color: Colors.grey.shade300),
      ),
      child: DropdownButtonFormField<String>(
        // ใช้ initialValue ร่วมกับ ValueKey เพื่อให้ dropdown อัปเดตตามค่าที่เปลี่ยน
        initialValue: isValidSelection ? selectedPurchaseTypeCode : null,
        decoration: InputDecoration(
          filled: true,
          fillColor: Colors.white,
          floatingLabelBehavior: FloatingLabelBehavior.always,
          contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 14),
          labelText: global.language("purchase_type"),
          labelStyle: const TextStyle(color: Colors.black87, fontWeight: FontWeight.bold, fontSize: 14),
          border: InputBorder.none,
          prefixIcon: Icon(Icons.category_outlined, color: _primaryColor),
        ),
        items: purchaseTypes?.map((type) {
          return DropdownMenuItem<String>(
            value: type.code,
            child: Text(type.getName(global.systemLanguage)),
          );
        }).toList(),
        onChanged: (value) {
          if (onPurchaseTypeChanged != null && value != null) {
            // หาชื่อของประเภทที่เลือก
            final selectedType = purchaseTypes?.firstWhere(
              (t) => t.code == value,
              orElse: () => PurchaseTypeModel(
                code: value,
                names: [LanguageDataModel(code: global.systemLanguage, name: value)],
              ),
            );
            onPurchaseTypeChanged!(value, selectedType?.getName(global.systemLanguage));
          }
        },
        icon: Icon(Icons.arrow_drop_down, color: _primaryColor),
        dropdownColor: Colors.white,
        isExpanded: true,
      ),
    );
  }

  /// 💱 สร้าง Dropdown สำหรับเลือกสกุลเงิน (สำหรับ PO Multi-Currency)
  /// แสดงทั้ง Base Currency (สำหรับลงบัญชี) และ Current Currency (สกุลเงินเอกสาร)
  Widget _buildCurrencySection() {
    // ตรวจสอบว่า selectedDocCurrencyCode มีอยู่ใน currencies หรือไม่
    final isValidSelection = selectedDocCurrencyCode != null &&
        currencies != null &&
        currencies!.any((c) => c.code == selectedDocCurrencyCode);

    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // 🔷 Base Currency (สกุลเงินหลักสำหรับลงบัญชี) - Readonly
        Container(
          decoration: BoxDecoration(
            color: Colors.grey.shade100,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: Colors.grey.shade300),
          ),
          child: InputDecorator(
            decoration: InputDecoration(
              filled: true,
              fillColor: Colors.grey.shade100,
              floatingLabelBehavior: FloatingLabelBehavior.always,
              contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 14),
              labelText: global.language('base_currency_accounting'),
              labelStyle: const TextStyle(color: Colors.black87, fontWeight: FontWeight.bold, fontSize: 14),
              border: InputBorder.none,
              prefixIcon: Icon(Icons.account_balance, color: Colors.grey.shade600),
              suffixIcon: Container(
                margin: const EdgeInsets.all(8),
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: Colors.grey.shade300,
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Text(
                  'BASE',
                  style: TextStyle(
                    fontSize: 10,
                    fontWeight: FontWeight.bold,
                    color: Colors.grey.shade700,
                  ),
                ),
              ),
            ),
            child: Text(
              baseCurrency != null 
                ? '$baseCurrency ${baseCurrencySymbol != null ? '($baseCurrencySymbol)' : ''}'
                : global.language('not_specified'),
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.bold,
                color: baseCurrency != null ? Colors.black87 : Colors.grey,
              ),
            ),
          ),
        ),
        const SizedBox(height: 10),
        
        // 🔶 Current Currency (สกุลเงินเอกสาร) - Dropdown
        Container(
          key: ValueKey('currency_dropdown_$selectedDocCurrencyCode'),
          decoration: BoxDecoration(
            color: Colors.white,
            borderRadius: BorderRadius.circular(10),
            border: Border.all(color: Colors.grey.shade300),
          ),
          child: DropdownButtonFormField<String>(
            initialValue: isValidSelection ? selectedDocCurrencyCode : null,
            decoration: InputDecoration(
              filled: true,
              fillColor: Colors.white,
              floatingLabelBehavior: FloatingLabelBehavior.always,
              contentPadding: EdgeInsets.symmetric(horizontal: 12, vertical: 14),
              labelText: global.language('doc_currency'),
              labelStyle: const TextStyle(color: Colors.black87, fontWeight: FontWeight.bold, fontSize: 14),
              border: InputBorder.none,
              prefixIcon: Icon(Icons.currency_exchange, color: _primaryColor),
              suffixIcon: Container(
                margin: const EdgeInsets.all(8),
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                decoration: BoxDecoration(
                  color: _primaryColor.withValues(alpha: 0.2),
                  borderRadius: BorderRadius.circular(4),
                ),
                child: Text(
                  'DOC',
                  style: TextStyle(
                    fontSize: 10,
                    fontWeight: FontWeight.bold,
                    color: _primaryColor,
                  ),
                ),
              ),
            ),
            items: currencies?.map((currency) {
              return DropdownMenuItem<String>(
                value: currency.code,
                child: Text('${currency.code} (${currency.symbol}) - ${currency.name}'),
              );
            }).toList(),
            onChanged: (value) {
              if (onCurrencyChanged != null && value != null) {
                // หาข้อมูลสกุลเงินที่เลือก
                final selectedCurrency = currencies?.firstWhere(
                  (c) => c.code == value,
                  orElse: () => CurrencyModel(code: value, name: value, symbol: value),
                );
                // ดึงอัตราแลกเปลี่ยนล่าสุด (ถ้ามี)
                double? latestRate;
                if (selectedCurrency?.exchangeRates != null && selectedCurrency!.exchangeRates.isNotEmpty) {
                  // หาอัตราล่าสุด (เรียงตามวันที่ล่าสุด)
                  final sortedRates = List<ExchangeRateEntryModel>.from(selectedCurrency.exchangeRates)
                    ..sort((a, b) => b.date.compareTo(a.date));
                  latestRate = sortedRates.first.rate;
                }
                onCurrencyChanged!(value, selectedCurrency?.symbol, latestRate);
              }
            },
            icon: Icon(Icons.arrow_drop_down, color: _primaryColor),
            dropdownColor: Colors.white,
            isExpanded: true,
          ),
        ),
        
        // 📊 ช่องกรอกอัตราแลกเปลี่ยน (แก้ไขได้)
        if (selectedDocCurrencyCode != null && selectedDocCurrencyCode != baseCurrency) ...[
          SizedBox(height: 8),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            decoration: BoxDecoration(
              color: Colors.green.shade50,
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: Colors.green.shade200),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Icon(Icons.sync_alt, size: 16, color: Colors.green.shade700),
                    SizedBox(width: 8),
                    Text(
                      global.language('exchange_rate'),
                      style: TextStyle(
                        fontSize: 12,
                        color: Colors.green.shade700,
                        fontWeight: FontWeight.w500,
                      ),
                    ),
                  ],
                ),
                SizedBox(height: 4),
                Row(
                  children: [
                    Text(
                      '1 $selectedDocCurrencyCode = ',
                      style: TextStyle(
                        fontSize: 14,
                        color: Colors.green.shade800,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    SizedBox(width: 8),
                    Expanded(
                      child: GestureDetector(
                        onTap: () async {
                          final result = await showNumericInputDialog(
                            context,
                            title: global.language('exchange_rate'),
                            initialValue: selectedExchangeRate?.toStringAsFixed(4) ?? '1.0000',
                            maxLength: 12,
                            decimalPlaces: 4,
                            allowNegative: false,
                          );
                          if (result != null && onExchangeRateChanged != null) {
                            final rate = double.tryParse(result) ?? 1.0;
                            onExchangeRateChanged!(rate);
                          }
                        },
                        child: Container(
                          padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                          decoration: BoxDecoration(
                            color: Colors.white,
                            borderRadius: BorderRadius.circular(6),
                            border: Border.all(color: Colors.green.shade300),
                          ),
                          child: Text(
                            selectedExchangeRate?.toStringAsFixed(4) ?? '1.0000',
                            textAlign: TextAlign.center,
                            style: TextStyle(
                              fontSize: 14,
                              fontWeight: FontWeight.bold,
                              color: Colors.green.shade800,
                            ),
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(width: 8),
                    Text(
                      baseCurrency ?? '',
                      style: TextStyle(
                        fontSize: 14,
                        color: Colors.green.shade800,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ] else if (selectedDocCurrencyCode == baseCurrency) ...[
          // ถ้าเป็นสกุลเงินหลัก แสดงอัตราแลกเปลี่ยนเป็น 1 (ไม่ให้แก้ไข)
          const SizedBox(height: 8),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            decoration: BoxDecoration(
              color: Colors.grey.shade100,
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: Colors.grey.shade300),
            ),
            child: Row(
              children: [
                Icon(Icons.sync_alt, size: 16, color: Colors.grey.shade600),
                const SizedBox(width: 8),
                Text(
                  '1 $selectedDocCurrencyCode = 1.0000 ${baseCurrency ?? ''}',
                  style: TextStyle(
                    fontSize: 13,
                    color: Colors.grey.shade600,
                    fontWeight: FontWeight.w500,
                  ),
                ),
              ],
            ),
          ),
        ],
      ],
    );
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        gradient: LinearGradient(
          colors: [const Color(0xFFF8F9FD), const Color(0xFFEEF1F8), const Color(0xFFE8EBF5)],
          begin: Alignment.topCenter,
          end: Alignment.bottomCenter,
          stops: const [0.0, 0.5, 1.0],
        ),
      ),
      width: double.infinity,
      padding: EdgeInsets.all(16),
      child: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            // 📅 Section: วันที่และเวลา
            _buildSectionHeader(icon: Icons.notes_rounded, title: global.language("date_time_section")),
            SizedBox(height: 4), // เพิ่มที่ว่างด้านบนให้ floating label ไม่โดนตัด
            Row(
              children: [
                Expanded(
                  child: FocusTraversalOrder(
                    order: NumericFocusOrder(1.0),
                    child: CustomDatePicker(
                      labelText: global.language("doc_date"),
                      // useBuddhistCalendar และ languageCode จะอ่านจากสาขาอัตโนมัติ
                      useIconSelectDate: true,
                      // ต้อง pass initialDate เพื่อให้ CustomDatePicker แสดงวันที่ที่ถูกต้อง (local time)
                      initialDate: global.safeParseDatetime(screenData.docdatetime).toLocal(),
                      // callback เมื่อเลือกวันที่จาก CustomDatePicker
                      onDateSelected: (date) {
                        if (date != null) {
                          setState(() {
                            // แปลง UTC เป็น local ก่อนเพื่อให้ได้เวลา local ที่ถูกต้อง
                            final currentTimeLocal = global.safeParseDatetime(screenData.docdatetime).toLocal();
                            // รวมวันที่ที่เลือกกับเวลา local
                            final combinedDateTime = DateTime(date.year, date.month, date.day, currentTimeLocal.hour, currentTimeLocal.minute, currentTimeLocal.second, currentTimeLocal.millisecond);

                            // แปลงเป็น UTC ก่อนบันทึก ตามกฎ CLAUDE.md
                            screenData.docdatetime = combinedDateTime.toUtc().toIso8601String();

                            // อัพเดท docdatelocal ด้วย - สำหรับ mainapi สร้าง docno
                            screenData.docdatelocal = "${date.year}${date.month.toString().padLeft(2, '0')}${date.day.toString().padLeft(2, '0')}";

                            // เมื่อเปลี่ยน docdatetime ให้ taxdocdate เปลี่ยนตามด้วย
                            final currentTaxTimeLocal = global.safeParseDatetime(screenData.taxdocdate).toLocal();
                            final newTaxDateTime = DateTime(date.year, date.month, date.day, currentTaxTimeLocal.hour, currentTaxTimeLocal.minute, currentTaxTimeLocal.second, currentTaxTimeLocal.millisecond);
                            // แปลงเป็น UTC ก่อนบันทึก ตามกฎ CLAUDE.md
                            screenData.taxdocdate = newTaxDateTime.toUtc().toIso8601String();

                            // อัพเดท docTimeController ด้วย (เวลา local ไม่เปลี่ยน)
                            docTimeController.text = DateFormat('HH:mm').format(currentTimeLocal);
                          });
                        }
                      },
                      // decoration ไม่ได้ใช้โดย CustomDatePicker (widget build InputDecoration เอง)
                      // แต่ต้อง pass เพราะเป็น required parameter
                      decoration: _buildInputDecoration(labelText: global.language("doc_date")),
                    ),
                  ),
                ),
                SizedBox(width: 5),
                Expanded(
                  child: FocusTraversalOrder(
                    order: NumericFocusOrder(2.0),
                    child: CustomTimePicker(
                      labelText: global.language("doc_time"),
                      controller: docTimeController,
                      useIconSelectTime: true,
                      initialTime: TimeOfDay(hour: global.safeParseDatetime(screenData.docdatetime).toLocal().hour, minute: global.safeParseDatetime(screenData.docdatetime).toLocal().minute),
                      onTimeSelected: (time) {
                        if (time != null) {
                          setState(() {
                            // แปลง UTC เป็น local ก่อนเพื่อให้ได้วันที่ local ที่ถูกต้อง
                            final currentDateLocal = global.safeParseDatetime(screenData.docdatetime).toLocal();
                            // รวมวันที่ local กับเวลาที่เลือก
                            final combinedDateTime = DateTime(currentDateLocal.year, currentDateLocal.month, currentDateLocal.day, time.hour, time.minute, 0, 0);
                            // แปลงเป็น UTC ก่อนบันทึก ตามกฎ CLAUDE.md
                            screenData.docdatetime = combinedDateTime.toUtc().toIso8601String();
                            docTimeController.text = '${time.hour.toString().padLeft(2, '0')}:${time.minute.toString().padLeft(2, '0')}';
                          });
                        }
                      },
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 10),

            // 📦 Section: ประเภทการจัดซื้อ/ใบเสนอราคา (สำหรับ PO, QT)
            if ((transactionType == global.TransactionTypeEnum.purchaseorder || transactionType == global.TransactionTypeEnum.quotation) && purchaseTypes != null && purchaseTypes!.isNotEmpty) ...[
              _buildPurchaseTypeDropdown(),
              const SizedBox(height: 10),
            ],

            // 💱 Section: สกุลเงิน (สำหรับ PO, QT Multi-Currency)
            // แสดงเฉพาะเมื่อมีสกุลเงินมากกว่า 1 สกุล (ถ้ามีแค่ 1 สกุล ถือว่าเป็นสกุลเงินหลักตัวเดียว)
            if ((transactionType == global.TransactionTypeEnum.purchaseorder || transactionType == global.TransactionTypeEnum.quotation) && currencies != null && currencies!.length > 1) ...[
              _buildCurrencySection(),
              const SizedBox(height: 10),
            ],

            // 👤 Section: ลูกค้า/ผู้จำหน่าย
            (transactionType == global.TransactionTypeEnum.purchase ||
                    transactionType == global.TransactionTypeEnum.purchaseorder ||
                    transactionType == global.TransactionTypeEnum.purchasepartial ||
                    transactionType == global.TransactionTypeEnum.purchasereturn ||
                    transactionType == global.TransactionTypeEnum.sale ||
                    transactionType == global.TransactionTypeEnum.saleorder ||
                    transactionType == global.TransactionTypeEnum.salereturn ||
                    transactionType == global.TransactionTypeEnum.quotation ||
                    transactionType == global.TransactionTypeEnum.accrualreceive ||
                    transactionType == global.TransactionTypeEnum.advancePayment ||
                    transactionType == global.TransactionTypeEnum.advancePaymentRefund ||
                    transactionType == global.TransactionTypeEnum.deposit ||
                    transactionType == global.TransactionTypeEnum.depositRefund ||
                    transactionType == global.TransactionTypeEnum.paidAdvance ||
                    transactionType == global.TransactionTypeEnum.paidAdvanceRefund ||
                    transactionType == global.TransactionTypeEnum.receiveDeposit ||
                    transactionType == global.TransactionTypeEnum.receiveDepositRefund)
                ? _buildSectionHeader(
                    icon:
                        (transactionType == global.TransactionTypeEnum.purchase ||
                            transactionType == global.TransactionTypeEnum.purchaseorder ||
                            transactionType == global.TransactionTypeEnum.purchasepartial ||
                            transactionType == global.TransactionTypeEnum.purchasereturn ||
                            transactionType == global.TransactionTypeEnum.accrualreceive ||
                            transactionType == global.TransactionTypeEnum.advancePayment ||
                            transactionType == global.TransactionTypeEnum.advancePaymentRefund ||
                            transactionType == global.TransactionTypeEnum.deposit ||
                            transactionType == global.TransactionTypeEnum.depositRefund)
                        ? Icons.storefront_rounded
                        : Icons.person_rounded,
                    title:
                        (transactionType == global.TransactionTypeEnum.purchase ||
                            transactionType == global.TransactionTypeEnum.purchaseorder ||
                            transactionType == global.TransactionTypeEnum.purchasepartial ||
                            transactionType == global.TransactionTypeEnum.purchasereturn ||
                            transactionType == global.TransactionTypeEnum.accrualreceive ||
                            transactionType == global.TransactionTypeEnum.advancePayment ||
                            transactionType == global.TransactionTypeEnum.advancePaymentRefund ||
                            transactionType == global.TransactionTypeEnum.deposit ||
                            transactionType == global.TransactionTypeEnum.depositRefund)
                        ? global.language("supplier_section")
                        : global.language("customer_section"),
                  )
                : Container(),
            (transactionType == global.TransactionTypeEnum.purchase ||
                    transactionType == global.TransactionTypeEnum.purchaseorder ||
                    transactionType == global.TransactionTypeEnum.purchasepartial ||
                    transactionType == global.TransactionTypeEnum.purchasereturn ||
                    transactionType == global.TransactionTypeEnum.sale ||
                    transactionType == global.TransactionTypeEnum.saleorder ||
                    transactionType == global.TransactionTypeEnum.salereturn ||
                    transactionType == global.TransactionTypeEnum.quotation ||
                    transactionType == global.TransactionTypeEnum.accrualreceive ||
                    transactionType == global.TransactionTypeEnum.advancePayment ||
                    transactionType == global.TransactionTypeEnum.advancePaymentRefund ||
                    transactionType == global.TransactionTypeEnum.deposit ||
                    transactionType == global.TransactionTypeEnum.depositRefund ||
                    transactionType == global.TransactionTypeEnum.paidAdvance ||
                    transactionType == global.TransactionTypeEnum.paidAdvanceRefund ||
                    transactionType == global.TransactionTypeEnum.receiveDeposit ||
                    transactionType == global.TransactionTypeEnum.receiveDepositRefund)
                ? Row(
                    children: [
                      Expanded(
                        child: RawKeyboardListener(
                          focusNode: FocusNode(),
                          onKey: (RawKeyEvent event) {
                            if (event is RawKeyDownEvent) {
                              if (event.logicalKey == LogicalKeyboardKey.f2) {
                                if (transactionType == global.TransactionTypeEnum.purchase ||
                                    transactionType == global.TransactionTypeEnum.purchaseorder ||
                                    transactionType == global.TransactionTypeEnum.purchasepartial ||
                                    transactionType == global.TransactionTypeEnum.purchasereturn ||
                                    transactionType == global.TransactionTypeEnum.accrualreceive ||
                                    transactionType == global.TransactionTypeEnum.advancePayment ||
                                    transactionType == global.TransactionTypeEnum.advancePaymentRefund ||
                                    transactionType == global.TransactionTypeEnum.deposit ||
                                    transactionType == global.TransactionTypeEnum.depositRefund) {
                                  searchSupplier(word: "");
                                } else {
                                  searchCustomer(word: "");
                                }
                              }
                            }
                          },
                          child: FocusTraversalOrder(
                            order: const NumericFocusOrder(3.0),
                            child: TextFieldCustom(
                              textInputAction: TextInputAction.next,
                              onSubmitted: (value) {
                                FocusScope.of(context).nextFocus();
                              },
                              onChanged: (code) {
                                debouncer.run(() {
                                  try {
                                    if (code.trim().isNotEmpty) {
                                      if (transactionType == global.TransactionTypeEnum.purchase ||
                                          transactionType == global.TransactionTypeEnum.purchaseorder ||
                                          transactionType == global.TransactionTypeEnum.purchasepartial ||
                                          transactionType == global.TransactionTypeEnum.purchasereturn ||
                                          transactionType == global.TransactionTypeEnum.accrualreceive ||
                                          transactionType == global.TransactionTypeEnum.advancePayment ||
                                          transactionType == global.TransactionTypeEnum.advancePaymentRefund ||
                                          transactionType == global.TransactionTypeEnum.deposit ||
                                          transactionType == global.TransactionTypeEnum.depositRefund) {
                                        CustomerRepository()
                                            .getSupplierByCode(code.trim())
                                            .then((value) {
                                              if (value.success && value.data != null) {
                                                CustomerModel cust = CustomerModel.fromJson(value.data);
                                                if (cust.iscreditor) {
                                                  custnamesController.text = global.activeLangName(cust.names);
                                                  screenData.custnames = cust.names;
                                                } else {
                                                  custnamesController.text = global.language("Supplier Not Found");
                                                  screenData.custnames = [];
                                                }
                                              } else {
                                                custnamesController.text = global.language("Supplier Not Found");
                                              }
                                            })
                                            .onError((error, stackTrace) {
                                              custnamesController.text = global.language("Supplier Not Found");
                                            });
                                        screenData.custcode = code;
                                      } else {
                                        CustomerRepository()
                                            .getCustomerByCode(code.trim())
                                            .then((value) {
                                              if (value.success && value.data != null) {
                                                CustomerModel cust = CustomerModel.fromJson(value.data);
                                                custnamesController.text = global.activeLangName(cust.names);
                                                screenData.custnames = cust.names;
                                              } else {
                                                custnamesController.text = global.language("Customer Not Found");
                                              }
                                            })
                                            .onError((error, stackTrace) {
                                              custnamesController.text = global.language("Customer Not Found");
                                            });
                                        screenData.custcode = code;
                                      }
                                    } else {
                                      custnamesController.text = global.language("Customer Not Found");
                                    }
                                  } catch (_) {}
                                });
                              },
                              textAlign: TextAlign.left,
                              controller: custCodeController,
                              decoration: _buildInputDecoration(
                                labelText:
                                    (transactionType == global.TransactionTypeEnum.purchase ||
                                        transactionType == global.TransactionTypeEnum.purchaseorder ||
                                        transactionType == global.TransactionTypeEnum.purchasepartial ||
                                        transactionType == global.TransactionTypeEnum.purchasereturn ||
                                        transactionType == global.TransactionTypeEnum.accrualreceive ||
                                        transactionType == global.TransactionTypeEnum.advancePayment ||
                                        transactionType == global.TransactionTypeEnum.advancePaymentRefund ||
                                        transactionType == global.TransactionTypeEnum.deposit ||
                                        transactionType == global.TransactionTypeEnum.depositRefund)
                                    ? global.language("supplier_code")
                                    : global.language("customer_code"),
                                suffixIcon: Row(
                                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    IconButton(
                                      focusNode: FocusNode(skipTraversal: true),
                                      icon: const Icon(Icons.search),
                                      onPressed: () {
                                        if (transactionType == global.TransactionTypeEnum.purchase ||
                                            transactionType == global.TransactionTypeEnum.purchaseorder ||
                                            transactionType == global.TransactionTypeEnum.purchasepartial ||
                                            transactionType == global.TransactionTypeEnum.purchasereturn ||
                                            transactionType == global.TransactionTypeEnum.accrualreceive ||
                                            transactionType == global.TransactionTypeEnum.advancePayment ||
                                            transactionType == global.TransactionTypeEnum.advancePaymentRefund ||
                                            transactionType == global.TransactionTypeEnum.deposit ||
                                            transactionType == global.TransactionTypeEnum.depositRefund) {
                                          searchSupplier(word: "");
                                        } else {
                                          searchCustomer(word: "");
                                        }
                                      },
                                    ),
                                  ],
                                ),
                              ),
                            ),
                          ),
                        ),
                      ),
                      SizedBox(width: 5),
                      Expanded(
                        child: Focus(
                          skipTraversal: true,
                          canRequestFocus: false,
                          child: TextFieldCustom(
                            readOnly: true,
                            focusNode: null,
                            textAlign: TextAlign.left,
                            controller: custnamesController,
                            decoration: _buildInputDecoration(
                              labelText:
                                  (transactionType == global.TransactionTypeEnum.purchase ||
                                      transactionType == global.TransactionTypeEnum.purchaseorder ||
                                      transactionType == global.TransactionTypeEnum.purchasepartial ||
                                      transactionType == global.TransactionTypeEnum.purchasereturn ||
                                      transactionType == global.TransactionTypeEnum.accrualreceive ||
                                      transactionType == global.TransactionTypeEnum.advancePayment ||
                                      transactionType == global.TransactionTypeEnum.advancePaymentRefund ||
                                      transactionType == global.TransactionTypeEnum.deposit ||
                                      transactionType == global.TransactionTypeEnum.depositRefund)
                                  ? global.language("supplier_name")
                                  : global.language("customer_name"),
                              readOnly: true,
                            ),
                          ),
                        ),
                      ),
                    ],
                  )
                : Container(),
            const SizedBox(height: 10),
            // 👔 Section: พนักงานขาย
            _buildSectionHeader(icon: Icons.badge_rounded, title: global.language("salesperson_section")),
            Row(
              children: [
                Expanded(
                  child: FocusTraversalOrder(
                    order: NumericFocusOrder(4.0),
                    child: TextFieldCustom(
                      textInputAction: TextInputAction.next,
                      onSubmitted: (value) {
                        FocusScope.of(context).nextFocus();
                      },
                      onChanged: (code) {
                        debouncer.run(() {
                          if (code.trim().isNotEmpty) {
                            EmployeeRepository()
                                .getEmployeeByCode(code.trim())
                                .then((value) {
                                  if (value.success && value.data != null) {
                                    EmployeeModel emp = EmployeeModel.fromJson(value.data);
                                    saleNameController.text = emp.name;
                                    screenData.salename = emp.name;
                                    screenData.salecode = code;
                                  } else {
                                    saleNameController.text = global.language("Employee_not_found");
                                    screenData.salename = "";
                                    screenData.salecode = "";
                                  }
                                })
                                .onError((error, stackTrace) {
                                  saleNameController.text = global.language("Employee_not_found");
                                  screenData.salename = "";
                                  screenData.salecode = "";
                                });
                          } else {
                            screenData.salename = "";
                            screenData.salecode = "";
                            saleNameController.text = global.language("Employee_not_found");
                          }
                        });
                      },
                      textAlign: TextAlign.left,
                      controller: saleCodeController,
                      decoration: _buildInputDecoration(
                        labelText: global.language("sale_code"),
                        suffixIcon: Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            IconButton(
                              focusNode: FocusNode(skipTraversal: true),
                              icon: Icon(Icons.search, color: _primaryColor),
                              onPressed: () {
                                searchSale(word: "");
                              },
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                ),

                SizedBox(width: 5),
                Expanded(
                  child: Focus(
                    skipTraversal: true,
                    canRequestFocus: false,
                    child: TextFieldCustom(
                      readOnly: true,
                      focusNode: null,
                      textAlign: TextAlign.left,
                      controller: saleNameController,
                      decoration: _buildInputDecoration(labelText: global.language("sale_name"), readOnly: true),
                    ),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 10),
            // 💰 Section: ภาษีมูลค่าเพิ่ม
            (transactionType == global.TransactionTypeEnum.purchase ||
                    transactionType == global.TransactionTypeEnum.purchaseorder ||
                    transactionType == global.TransactionTypeEnum.purchasepartial ||
                    transactionType == global.TransactionTypeEnum.purchasereturn ||
                    transactionType == global.TransactionTypeEnum.sale ||
                    transactionType == global.TransactionTypeEnum.saleorder ||
                    transactionType == global.TransactionTypeEnum.salereturn ||
                    transactionType == global.TransactionTypeEnum.quotation ||
                    transactionType == global.TransactionTypeEnum.accrualreceive ||
                    transactionType == global.TransactionTypeEnum.deposit ||
                    transactionType == global.TransactionTypeEnum.depositRefund ||
                    transactionType == global.TransactionTypeEnum.receiveDeposit ||
                    transactionType == global.TransactionTypeEnum.receiveDepositRefund)
                ? _buildSectionHeader(icon: Icons.percent_rounded, title: global.language("vat_section"))
                : Container(),
            (transactionType == global.TransactionTypeEnum.purchase ||
                    transactionType == global.TransactionTypeEnum.purchaseorder ||
                    transactionType == global.TransactionTypeEnum.purchasepartial ||
                    transactionType == global.TransactionTypeEnum.purchasereturn ||
                    transactionType == global.TransactionTypeEnum.sale ||
                    transactionType == global.TransactionTypeEnum.saleorder ||
                    transactionType == global.TransactionTypeEnum.salereturn ||
                    transactionType == global.TransactionTypeEnum.quotation ||
                    transactionType == global.TransactionTypeEnum.accrualreceive ||
                    transactionType == global.TransactionTypeEnum.deposit ||
                    transactionType == global.TransactionTypeEnum.depositRefund ||
                    transactionType == global.TransactionTypeEnum.receiveDeposit ||
                    transactionType == global.TransactionTypeEnum.receiveDepositRefund)
                ? Container(
                    margin: EdgeInsets.only(bottom: 10),
                    child: InputDecorator(
                      decoration: _buildGroupDecoration(labelText: global.language('vat_type')),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.start,
                        children: [
                          Expanded(
                            child: Row(
                              children: [
                                Radio(
                                  value: 0,
                                  groupValue: screenData.vattype,
                                  onChanged: (value) {
                                    setState(() {
                                      screenData.vattype = 0;
                                      Future.delayed(const Duration(milliseconds: 200), () {
                                        calTotalValue();
                                      });
                                    });
                                  },
                                ),
                                Expanded(child: Text(global.language("vat_exclude"), overflow: TextOverflow.clip)),
                              ],
                            ),
                          ),
                          Expanded(
                            child: Row(
                              children: [
                                Radio(
                                  value: 1,
                                  groupValue: screenData.vattype,
                                  onChanged: (value) {
                                    setState(() {
                                      screenData.vattype = 1;
                                      Future.delayed(const Duration(milliseconds: 200), () {
                                        calTotalValue();
                                      });
                                    });
                                  },
                                ),
                                Expanded(child: Text(global.language("vat_include"), overflow: TextOverflow.clip)),
                              ],
                            ),
                          ),
                          Expanded(
                            child: Row(
                              children: [
                                Radio(
                                  value: 2,
                                  groupValue: screenData.vattype,
                                  activeColor: Colors.blue,
                                  onChanged: (value) {
                                    setState(() {
                                      screenData.vattype = 2;
                                      Future.delayed(const Duration(milliseconds: 200), () {
                                        calTotalValue();
                                      });
                                    });
                                  },
                                ),
                                Expanded(child: Text(global.language("vat_zero"), overflow: TextOverflow.clip)),
                              ],
                            ),
                          ),
                          Expanded(
                            child: Row(
                              children: [
                                Radio(
                                  value: 3,
                                  groupValue: screenData.vattype,
                                  activeColor: Colors.blue,
                                  onChanged: (value) {
                                    setState(() {
                                      screenData.vattype = 3;
                                      Future.delayed(const Duration(milliseconds: 200), () {
                                        calTotalValue();
                                      });
                                    });
                                  },
                                ),
                                Expanded(child: Text(global.language("vat_none"), overflow: TextOverflow.clip)),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                  )
                : Container(),
            (transactionType == global.TransactionTypeEnum.purchase || transactionType == global.TransactionTypeEnum.sale || transactionType == global.TransactionTypeEnum.accrualreceive)
                ? Container(
                    margin: EdgeInsets.only(bottom: 10),
                    child: InputDecorator(
                      decoration: _buildGroupDecoration(labelText: global.language('inquiry_type')),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.start,
                        children: [
                          Expanded(
                            child: Row(
                              children: [
                                Radio(
                                  value: 0,
                                  groupValue: screenData.inquirytype,
                                  onChanged: (value) {
                                    setState(() {
                                      screenData.inquirytype = 0;
                                    });
                                  },
                                ),
                                Expanded(child: Text(global.language("credit"), overflow: TextOverflow.clip)),
                              ],
                            ),
                          ),
                          Expanded(
                            child: Row(
                              children: [
                                Radio(
                                  value: 1,
                                  groupValue: screenData.inquirytype,
                                  onChanged: (value) {
                                    setState(() {
                                      screenData.inquirytype = 1;
                                    });
                                  },
                                ),
                                Expanded(child: Text(global.language("cash"), overflow: TextOverflow.clip)),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                  )
                : (transactionType == global.TransactionTypeEnum.purchasereturn || transactionType == global.TransactionTypeEnum.salereturn)
                ? Container(
                    margin: EdgeInsets.only(bottom: 10),
                    child: InputDecorator(
                      decoration: _buildGroupDecoration(labelText: global.language('inquiry_type')),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.start,
                        children: [
                          Expanded(
                            child: Row(
                              children: [
                                Radio(
                                  value: 0,
                                  groupValue: screenData.inquirytype,
                                  onChanged: (value) {
                                    setState(() {
                                      screenData.inquirytype = 0;
                                    });
                                  },
                                ),
                                Expanded(
                                  child: Text(
                                    (transactionType == global.TransactionTypeEnum.purchasereturn) ? global.language("return_credit_purchaser") : global.language("return_credit_sale"),
                                    overflow: TextOverflow.clip,
                                  ),
                                ),
                              ],
                            ),
                          ),
                          Expanded(
                            child: Row(
                              children: [
                                Radio(
                                  value: 1,
                                  groupValue: screenData.inquirytype,
                                  onChanged: (value) {
                                    setState(() {
                                      screenData.inquirytype = 1;
                                    });
                                  },
                                ),
                                Expanded(
                                  child: Text(
                                    (transactionType == global.TransactionTypeEnum.purchasereturn) ? global.language("reduce_credit_purchaser") : global.language("reduce_credit_sale"),
                                    overflow: TextOverflow.clip,
                                  ),
                                ),
                              ],
                            ),
                          ),
                          Expanded(
                            child: Row(
                              children: [
                                Radio(
                                  value: 2,
                                  groupValue: screenData.inquirytype,
                                  onChanged: (value) {
                                    setState(() {
                                      screenData.inquirytype = 2;
                                    });
                                  },
                                ),
                                Expanded(
                                  child: Text(
                                    (transactionType == global.TransactionTypeEnum.purchasereturn) ? global.language("return_cash_purchaser") : global.language("return_cash_sale"),
                                    overflow: TextOverflow.clip,
                                  ),
                                ),
                              ],
                            ),
                          ),
                          Expanded(
                            child: Row(
                              children: [
                                Radio(
                                  value: 3,
                                  groupValue: screenData.inquirytype,
                                  onChanged: (value) {
                                    setState(() {
                                      screenData.inquirytype = 3;
                                    });
                                  },
                                ),
                                Expanded(
                                  child: Text(
                                    (transactionType == global.TransactionTypeEnum.purchasereturn) ? global.language("reduce_cash_purchaser") : global.language("reduce_cash_sale"),
                                    overflow: TextOverflow.clip,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                  )
                : Container(),

            (transactionType == global.TransactionTypeEnum.stockpickupproduct)
                ? Container(
                    margin: EdgeInsets.only(bottom: 10),
                    child: InputDecorator(
                      decoration: _buildGroupDecoration(labelText: global.language('inquiry_type')),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.start,
                        children: [
                          Expanded(
                            child: Row(
                              children: [
                                Radio(
                                  value: 0,
                                  groupValue: screenData.inquirytype,
                                  onChanged: (value) {
                                    setState(() {
                                      screenData.inquirytype = 0;
                                    });
                                  },
                                ),
                                Expanded(child: Text(global.language("withdraw_production"), overflow: TextOverflow.clip)),
                              ],
                            ),
                          ),
                          Expanded(
                            child: Row(
                              children: [
                                Radio(
                                  value: 1,
                                  groupValue: screenData.inquirytype,
                                  onChanged: (value) {
                                    setState(() {
                                      screenData.inquirytype = 1;
                                    });
                                  },
                                ),
                                Expanded(child: Text(global.language("withdraw_for_your_own_use"), overflow: TextOverflow.clip)),
                              ],
                            ),
                          ),
                          Expanded(
                            child: Row(
                              children: [
                                Radio(
                                  value: 2,
                                  groupValue: screenData.inquirytype,
                                  onChanged: (value) {
                                    setState(() {
                                      screenData.inquirytype = 2;
                                    });
                                  },
                                ),
                                Expanded(child: Text(global.language("pick_up_damaged_items"), overflow: TextOverflow.clip)),
                              ],
                            ),
                          ),
                          Expanded(
                            child: Row(
                              children: [
                                Radio(
                                  value: 9,
                                  groupValue: screenData.inquirytype,
                                  onChanged: (value) {
                                    setState(() {
                                      screenData.inquirytype = 9;
                                    });
                                  },
                                ),
                                Expanded(child: Text(global.language("other"), overflow: TextOverflow.clip)),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                  )
                : Container(),
            (transactionType == global.TransactionTypeEnum.adjust)
                ? Container(
                    margin: EdgeInsets.only(bottom: 10),
                    child: InputDecorator(
                      decoration: _buildGroupDecoration(labelText: global.language('adjust_type')),
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.start,
                        children: [
                          Expanded(
                            child: Row(
                              children: [
                                Radio(
                                  value: 66,
                                  groupValue: screenData.transflag,
                                  onChanged: (value) {
                                    setState(() {
                                      if (screenData.transflag == 866 || screenData.transflag == 868) {
                                        for (var element in screenData.details!) {
                                          element.sumamount = 0;
                                        }
                                      }
                                      screenData.transflag = 66;

                                      headerTableDetail();
                                    });
                                  },
                                ),
                                Expanded(child: Text(global.language("increase"), overflow: TextOverflow.clip)),
                              ],
                            ),
                          ),
                          Expanded(
                            child: Row(
                              children: [
                                Radio(
                                  value: 68,
                                  groupValue: screenData.transflag,
                                  onChanged: (value) {
                                    setState(() {
                                      if (screenData.transflag == 866 || screenData.transflag == 868) {
                                        for (var element in screenData.details!) {
                                          element.sumamount = 0;
                                        }
                                      }
                                      screenData.transflag = 68;
                                      headerTableDetail();
                                    });
                                  },
                                ),
                                Expanded(child: Text(global.language("decrease"), overflow: TextOverflow.clip)),
                              ],
                            ),
                          ),
                          Expanded(
                            child: Row(
                              children: [
                                Radio(
                                  value: 866,
                                  groupValue: screenData.transflag,
                                  onChanged: (value) {
                                    setState(() {
                                      if (screenData.transflag == 66 || screenData.transflag == 68) {
                                        for (var element in screenData.details!) {
                                          element.qty = 0;
                                        }
                                      }
                                      screenData.transflag = 866;
                                      headerTableDetail();
                                    });
                                  },
                                ),
                                Expanded(child: Text(global.language('adjust_value_increase'), overflow: TextOverflow.clip)),
                              ],
                            ),
                          ),
                          Expanded(
                            child: Row(
                              children: [
                                Radio(
                                  value: 868,
                                  groupValue: screenData.transflag,
                                  onChanged: (value) {
                                    setState(() {
                                      if (screenData.transflag == 66 || screenData.transflag == 68) {
                                        for (var element in screenData.details!) {
                                          element.qty = 0;
                                        }
                                      }
                                      screenData.transflag = 868;
                                      headerTableDetail();
                                    });
                                  },
                                ),
                                Expanded(child: Text(global.language('adjust_value_decrease'), overflow: TextOverflow.clip)),
                              ],
                            ),
                          ),
                          // Expanded(
                          //     child: Row(
                          //   children: [
                          //     Radio(
                          //         value: 966,
                          //         groupValue: screenData.transflag,
                          //         onChanged: (value) {
                          //           setState(() {
                          //             screenData.transflag = 966;
                          //             headerTableDetail();
                          //           });
                          //         }),
                          //     Expanded(
                          //         child: Text(
                          //       global.language("adjust_cost"),
                          //       overflow: TextOverflow.clip,
                          //     ))
                          //   ],
                          // )),
                        ],
                      ),
                    ),
                  )
                : Container(),
            (transactionType == global.TransactionTypeEnum.purchase ||
                    transactionType == global.TransactionTypeEnum.purchaseorder ||
                    transactionType == global.TransactionTypeEnum.purchasepartial ||
                    transactionType == global.TransactionTypeEnum.purchasereturn ||
                    transactionType == global.TransactionTypeEnum.sale ||
                    transactionType == global.TransactionTypeEnum.saleorder ||
                    transactionType == global.TransactionTypeEnum.salereturn ||
                    transactionType == global.TransactionTypeEnum.quotation ||
                    transactionType == global.TransactionTypeEnum.accrualreceive ||
                    transactionType == global.TransactionTypeEnum.deposit ||
                    transactionType == global.TransactionTypeEnum.depositRefund ||
                    transactionType == global.TransactionTypeEnum.receiveDepositRefund)
                ? Container(
                    margin: EdgeInsets.only(bottom: 10),
                    child: Row(
                      children: [
                        Expanded(
                          child: FocusTraversalOrder(
                            order: NumericFocusOrder(5.0),
                            child: TextFieldCustom(
                              textInputAction: TextInputAction.next,
                              onSubmitted: (value) {
                                FocusScope.of(context).nextFocus();
                              },
                              decoration: _buildInputDecoration(labelText: global.language('var_rate')),
                              keyboardType: const TextInputType.numberWithOptions(decimal: true),
                              inputFormatters: [global.NumberInputFormatter()],
                              controller: vatRateController,
                              onChanged: (value) {
                                setState(() {
                                  screenData.vatrate = double.parse(value);
                                  calTotalValue();
                                });
                              },
                            ),
                          ),
                        ),
                        SizedBox(width: 10),
                        Expanded(
                          child: Focus(
                            skipTraversal: true,
                            canRequestFocus: false,
                            child: TextFieldCustom(
                              readOnly: true,
                              decoration: _buildInputDecoration(
                                labelText: global.language('branch'),
                                readOnly: true,
                                suffixIcon: Row(
                                  mainAxisAlignment: MainAxisAlignment.spaceBetween, // added line
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    IconButton(
                                      focusNode: FocusNode(skipTraversal: true),
                                      icon: Icon(Icons.search, color: _primaryColor),
                                      onPressed: () {
                                        Navigator.push(context, MaterialPageRoute(builder: (context) => const CompanyBranchSearchScreen(word: ""))).then((value) {
                                          setState(() {
                                            SearchGuidCodeNameModel result = value;
                                            if (result.isCancel == false) {
                                              screenData.branch!.guidfixed = result.guid;
                                              screenData.branch!.code = result.code;
                                              screenData.branch!.names = result.names;
                                            }
                                          });
                                        });
                                      },
                                    ),
                                  ],
                                ),
                              ),
                              controller: TextEditingController(text: "${screenData.branch!.code!} ~ ${global.activeLangName(screenData.branch!.names!)}"),
                            ), // ปิด TextField
                          ), // ปิด Focus
                        ), // ปิด Expanded
                      ],
                    ),
                  )
                : Container(
                    margin: EdgeInsets.only(bottom: 10),
                    child: Row(
                      children: [
                        Expanded(
                          child: Focus(
                            skipTraversal: true,
                            canRequestFocus: false,
                            child: TextFieldCustom(
                              readOnly: true,
                              decoration: _buildInputDecoration(
                                labelText: global.language('branch'),
                                readOnly: true,
                                suffixIcon: Row(
                                  mainAxisAlignment: MainAxisAlignment.spaceBetween, // added line
                                  mainAxisSize: MainAxisSize.min,
                                  children: [
                                    IconButton(
                                      focusNode: FocusNode(skipTraversal: true),
                                      icon: Icon(Icons.search, color: _primaryColor),
                                      onPressed: () {
                                        Navigator.push(context, MaterialPageRoute(builder: (context) => const CompanyBranchSearchScreen(word: ""))).then((value) {
                                          setState(() {
                                            SearchGuidCodeNameModel result = value;
                                            if (result.isCancel == false) {
                                              screenData.branch!.guidfixed = result.guid;
                                              screenData.branch!.code = result.code;
                                              screenData.branch!.names = result.names;
                                            }
                                          });
                                        });
                                      },
                                    ),
                                  ],
                                ),
                              ),
                              controller: TextEditingController(text: "${screenData.branch!.code!} ~ ${global.activeLangName(screenData.branch!.names!)}"),
                            ), // ปิด TextField
                          ), // ปิด Focus
                        ), // ปิด Expanded
                      ],
                    ),
                  ),
            _buildSectionHeader(icon: Icons.link_rounded, title: global.language("reference_section")),
            Container(
              margin: const EdgeInsets.only(bottom: 10),
              child: Row(
                children: [
                  (transactionType == global.TransactionTypeEnum.stockreceiveproduct || transactionType == global.TransactionTypeEnum.stockpickupproduct)
                      ? Expanded(
                          child: RawKeyboardListener(
                            focusNode: FocusNode(),
                            // onKey: (RawKeyEvent event) {
                            //   if (event is RawKeyDownEvent) {
                            //     if (event.logicalKey == LogicalKeyboardKey.f2) {
                            //       searchDocRef();
                            //     }
                            //   }
                            // },
                            child: FocusTraversalOrder(
                              order: NumericFocusOrder(6.0),
                              child: TextFieldCustom(
                                textInputAction: TextInputAction.next,
                                onSubmitted: (value) {
                                  FocusScope.of(context).nextFocus();
                                },
                                onChanged: (value) {
                                  setState(() {
                                    screenData.docrefno = value;
                                  });
                                },
                                textAlign: TextAlign.left,
                                controller: docRefNumberController,
                                decoration: _buildInputDecoration(labelText: global.language("doc_ref")),
                              ),
                            ),
                          ),
                        )
                      : (transactionType != global.TransactionTypeEnum.purchasereturn &&
                            transactionType != global.TransactionTypeEnum.salereturn &&
                            transactionType != global.TransactionTypeEnum.stockreturnproduct &&
                            transactionType != global.TransactionTypeEnum.advancePayment &&
                            transactionType != global.TransactionTypeEnum.deposit)
                      ? Expanded(
                          child: TextFieldCustom(
                            textInputAction: TextInputAction.next,
                            onSubmitted: (value) {
                              FocusScope.of(context).nextFocus();
                            },
                            onChanged: (value) {
                              setState(() {
                                screenData.docrefno = value;
                              });
                            },
                            textAlign: TextAlign.left,
                            controller: docRefNumberController,
                            readOnly:
                                (transactionType == global.TransactionTypeEnum.advancePaymentRefund ||
                                    transactionType == global.TransactionTypeEnum.depositRefund ||
                                    transactionType == global.TransactionTypeEnum.paidAdvanceRefund ||
                                    transactionType == global.TransactionTypeEnum.receiveDepositRefund)
                                ? true
                                : false,
                            decoration: _buildInputDecoration(
                              labelText: global.language("doc_ref"),
                              suffixIcon:
                                  (transactionType == global.TransactionTypeEnum.advancePaymentRefund ||
                                      transactionType == global.TransactionTypeEnum.depositRefund ||
                                      transactionType == global.TransactionTypeEnum.paidAdvanceRefund ||
                                      transactionType == global.TransactionTypeEnum.receiveDepositRefund)
                                  ? Row(
                                      mainAxisAlignment: MainAxisAlignment.spaceBetween,
                                      mainAxisSize: MainAxisSize.min,
                                      children: [
                                        IconButton(
                                          focusNode: FocusNode(skipTraversal: true),
                                          icon: const Icon(Icons.search),
                                          onPressed: () {
                                            searchDocRef(word: docRefNumberController.text);
                                          },
                                        ),
                                      ],
                                    )
                                  : null,
                            ),
                          ),
                        )
                      : Container(),
                  const SizedBox(width: 10),
                  if (transactionType != global.TransactionTypeEnum.purchasereturn &&
                      transactionType != global.TransactionTypeEnum.salereturn &&
                      transactionType != global.TransactionTypeEnum.stockreturnproduct &&
                      transactionType != global.TransactionTypeEnum.advancePayment &&
                      transactionType != global.TransactionTypeEnum.deposit)
                    Expanded(
                      child: CustomDatePicker(
                        useIconSelectDate: true,
                        key: ValueKey(screenData.docrefdate),
                        labelText: global.language("doc_ref_date"),
                        initialDate: global.safeParseDatetime(screenData.docrefdate),
                        // useBuddhistCalendar และ languageCode จะอ่านจากสาขาอัตโนมัติ
                        onDateSelected: (date) {
                          if (date != null) {
                            setState(() {
                              // กำหนดเวลาจากวันที่ที่เลือกโดยรักษาเวลาเดิม
                              final currentTime = global.safeParseDatetime(screenData.docrefdate);
                              final combinedDateTime = DateTime(date.year, date.month, date.day, currentTime.hour, currentTime.minute, currentTime.second, currentTime.millisecond);

                              screenData.docrefdate = combinedDateTime.toLocal().toIso8601String();
                            });
                          }
                        },
                        decoration: _buildInputDecoration(labelText: global.language("doc_ref_date")),
                      ),
                    ),
                ],
              ),
            ),

            (transactionType == global.TransactionTypeEnum.purchase ||
                    transactionType == global.TransactionTypeEnum.purchasepartial ||
                    transactionType == global.TransactionTypeEnum.purchasereturn ||
                    transactionType == global.TransactionTypeEnum.sale ||
                    transactionType == global.TransactionTypeEnum.saleorder ||
                    transactionType == global.TransactionTypeEnum.salereturn ||
                    transactionType == global.TransactionTypeEnum.stockreturnproduct ||
                    transactionType == global.TransactionTypeEnum.accrualreceive ||
                    transactionType == global.TransactionTypeEnum.deposit ||
                    transactionType == global.TransactionTypeEnum.depositRefund ||
                    transactionType == global.TransactionTypeEnum.receiveDeposit ||
                    transactionType == global.TransactionTypeEnum.receiveDepositRefund)
                ? Padding(
                    padding: EdgeInsets.only(bottom: 10.0),
                    child: Row(
                      children: [
                        Expanded(
                          child: FocusTraversalOrder(
                            order: NumericFocusOrder(7.0),
                            child: TextFieldCustom(
                              textInputAction: TextInputAction.next,
                              onSubmitted: (value) {
                                FocusScope.of(context).nextFocus();
                              },
                              decoration: _buildInputDecoration(
                                labelText: (transactionType == global.TransactionTypeEnum.purchase && (screenData.vattype == 0 || screenData.vattype == 1 || screenData.vattype == 2))
                                    ? "${global.language('tax_docno')} *"
                                    : global.language('tax_docno'),
                                hasError:
                                    (transactionType == global.TransactionTypeEnum.purchase &&
                                    (screenData.vattype == 0 || screenData.vattype == 1 || screenData.vattype == 2) &&
                                    screenData.taxdocno.trim().isEmpty),
                              ),
                              controller: taxDocNoController,
                              onChanged: (value) {
                                setState(() {
                                  screenData.taxdocno = value;
                                });
                              },
                            ),
                          ),
                        ),
                        SizedBox(width: 10),
                        Expanded(
                          child: CustomDatePicker(
                            key: ValueKey(screenData.taxdocdate),
                            labelText: global.language("tax_doc_date"),
                            initialDate: global.safeParseDatetime(screenData.taxdocdate),
                            // useBuddhistCalendar และ languageCode จะอ่านจากสาขาอัตโนมัติ
                            onDateSelected: (date) {
                              if (date != null) {
                                setState(() {
                                  // กำหนดเวลาจากวันที่ที่เลือกโดยรักษาเวลาเดิม
                                  final currentTime = global.safeParseDatetime(screenData.taxdocdate);
                                  final combinedDateTime = DateTime(date.year, date.month, date.day, currentTime.hour, currentTime.minute, currentTime.second, currentTime.millisecond);

                                  screenData.taxdocdate = combinedDateTime.toLocal().toIso8601String();
                                });
                              }
                            },
                            decoration: _buildInputDecoration(labelText: global.language("tax_doc_date")),
                          ),
                        ),
                      ],
                    ),
                  )
                : Container(),

            ///  docreferences - Array Management
            (transactionType == global.TransactionTypeEnum.purchase ||
                    transactionType == global.TransactionTypeEnum.purchasereturn ||
                    transactionType == global.TransactionTypeEnum.purchasepartial ||
                    transactionType == global.TransactionTypeEnum.accrualreceive ||
                    transactionType == global.TransactionTypeEnum.sale ||
                    transactionType == global.TransactionTypeEnum.salereturn)
                ? DocumentReferencesWidget(screenData: screenData, transactionType: transactionType, setState: setState, searchDocRef: searchDocRef, onDeleteDocReference: onDeleteDocReference)
                : Container(),

            /// Description field
            // 📝 Section: หมายเหตุ
            _buildSectionHeader(icon: Icons.notes_rounded, title: global.language("description_section")),
            Container(
              margin: EdgeInsets.only(bottom: 10),
              child: Row(
                children: [
                  Expanded(
                    child: FocusTraversalOrder(
                      order: NumericFocusOrder(8.0),
                      child: TextFieldCustom(
                        maxLines: 4,
                        textInputAction: TextInputAction.done,
                        onSubmitted: (value) {
                          // วนกลับไปยัง field แรก (วันที่) เมื่อจาก field สุดท้าย
                          FocusScope.of(context).unfocus();
                          Future.delayed(const Duration(milliseconds: 100), () {
                            FocusScope.of(context).focusInDirection(TraversalDirection.down);
                          });
                        },
                        decoration: _buildInputDecoration(labelText: global.language('disciption')),
                        controller: descriptionController,
                        onChanged: (value) {
                          setState(() {
                            screenData.description = value;
                          });
                        },
                      ),
                    ),
                  ),
                ],
              ),
            ),

            /// isusedelivery and is use transport
            (transactionType == global.TransactionTypeEnum.sale)
                ? Column(
                    children: [
                      if (global.posVersion == global.PosVersionEnum.restaurant)
                        Row(
                          children: [
                            Switch(
                              value: screenData.isdelivery!,
                              onChanged: (value) {
                                setState(() {
                                  screenData.isdelivery = value;
                                });
                              },
                              activeTrackColor: Colors.lightGreenAccent,
                              activeThumbColor: Colors.green,
                            ),
                            Text(global.language("is_use_delivery")),
                          ],
                        ),
                      const SizedBox(height: 10),
                      (screenData.isdelivery!)
                          ? Container(
                              padding: EdgeInsets.only(left: 20, right: 20),
                              child: Row(
                                mainAxisAlignment: MainAxisAlignment.start,
                                children: [
                                  Expanded(
                                    child: Focus(
                                      skipTraversal: true,
                                      canRequestFocus: false,
                                      child: TextFieldCustom(
                                        readOnly: true,
                                        textAlign: TextAlign.left,
                                        controller: TextEditingController(text: screenData.salechannelcode!),
                                        decoration: _buildInputDecoration(
                                          labelText: global.language("sale_channel_code"),
                                          readOnly: true,
                                          suffixIcon: Row(
                                            mainAxisAlignment: MainAxisAlignment.spaceBetween, // added line
                                            mainAxisSize: MainAxisSize.min,
                                            children: [
                                              IconButton(
                                                focusNode: FocusNode(skipTraversal: true),
                                                icon: Icon(Icons.search, color: _primaryColor),
                                                onPressed: () {
                                                  searchSaleChannel(word: "");
                                                },
                                              ),
                                            ],
                                          ),
                                        ),
                                      ),
                                    ), // ปิด Focus
                                  ), // ปิด Expanded
                                  SizedBox(width: 5),
                                  Expanded(
                                    child: Focus(
                                      skipTraversal: true,
                                      canRequestFocus: false,
                                      child: TextFieldCustom(
                                        enabled: false,
                                        focusNode: null,
                                        textAlign: TextAlign.left,
                                        controller: TextEditingController(text: screenData.salechannelgp!.toString()),
                                        decoration: _buildInputDecoration(labelText: global.language("sale_channel_gp"), readOnly: true),
                                      ),
                                    ), // ปิด Focus
                                  ), // ปิด Expanded
                                  SizedBox(width: 5),
                                  Text((screenData.salechannelgptype! == 0) ? "%" : global.language("money_symbol"), style: TextStyle(fontSize: 14)),
                                ],
                              ),
                            )
                          : Container(),

                      // /// transport
                      // Row(
                      //   children: [
                      //     Switch(
                      //       value: screenData.istransport!,
                      //       onChanged: (value) {
                      //         setState(() {
                      //           screenData.istransport = value;
                      //         });
                      //       },
                      //       activeTrackColor: Colors.lightGreenAccent,
                      //       activeColor: Colors.green,
                      //     ),
                      //     Text(global.language("is_use_transport")),
                      //   ],
                      // ),
                      // const SizedBox(
                      //   height: 10,
                      // ),
                      // (screenData.istransport!)
                      //     ? Container(
                      // padding: EdgeInsets.only(left: 20, right: 20),
                      //         child: Row(
                      //           mainAxisAlignment: MainAxisAlignment.start,
                      //           children: [
                      //             Expanded(
                      //               child: TextFieldCustom(
                      //                 readOnly: true,
                      //                 textAlign: TextAlign.left,
                      //                 controller: TextEditingController(text: screenData.transportcode!),
                      //                 decoration: InputDecoration(
                      //                   floatingLabelBehavior: FloatingLabelBehavior.always,
                      // border: OutlineInputBorder(),
                      //                   labelText: global.language("transportchannel_code"),
                      //                   suffixIcon: Row(
                      //                     mainAxisAlignment: MainAxisAlignment.spaceBetween, // added line
                      //                     mainAxisSize: MainAxisSize.min,
                      //                     children: [
                      //                       IconButton(
                      //                         focusNode: FocusNode(skipTraversal: true),
                      //                         icon: const Icon(Icons.search),
                      //                         onPressed: () {
                      //                           searchTransport(word: "");
                      //                         },
                      //                       ),
                      //                     ],
                      //                   ),
                      //                 ),
                      //               ),
                      //             ),
                      // SizedBox(width: 5),
                      //             Expanded(
                      //               ///transportamount  input number only and format number
                      //               child: TextFieldCustom(
                      //                 keyboardType: const TextInputType.numberWithOptions(decimal: true),
                      //                 inputFormatters: [global.NumberInputFormatter()],
                      //                 textAlign: TextAlign.left,
                      //                 controller: transportAmountController,
                      //                 decoration: InputDecoration(
                      //                   floatingLabelBehavior: FloatingLabelBehavior.always,
                      // border: OutlineInputBorder(),
                      //                   labelText: global.language("transportchannel_amount"),
                      //                 ),
                      //                 onChanged: (value) {
                      //                   /// check value null
                      //                   if (value.isEmpty) {
                      //                     screenData.transportamount = 0.00;
                      //                   } else {
                      //                     screenData.transportamount = double.parse(value.replaceAll(',', ''));
                      //                   }
                      //                 },
                      //               ),
                      //             ),
                      //           ],
                      //         ),
                      //       )
                      //     : Container(),
                    ],
                  )
                : Container(),
            const SizedBox(height: 20),

            (screenData.cancelreason!.isNotEmpty)
                ? Center(
                    child: Container(
                      margin: EdgeInsets.only(bottom: 10),
                      child: Text(
                        "***${global.language('cancel_reason')} : ${screenData.cancelreason!}***",
                        style: const TextStyle(color: Colors.red, fontSize: 14, fontWeight: FontWeight.bold),
                      ),
                    ),
                  )
                : Container(),

            // 💰 Section: ภาษีหัก ณ ที่จ่าย (WHT) - สำหรับ PO, QT
            if (transactionType == global.TransactionTypeEnum.purchaseorder || transactionType == global.TransactionTypeEnum.quotation)
              Padding(
                padding: const EdgeInsets.only(bottom: 10),
                child: WHTSectionWidget(
                  screenData: screenData,
                  setState: setState,
                  context: context,
                ),
              ),

            // 📅 Section: เครดิตเทอม - สำหรับ PO, QT
            if (transactionType == global.TransactionTypeEnum.purchaseorder || transactionType == global.TransactionTypeEnum.quotation)
              Padding(
                padding: const EdgeInsets.only(bottom: 10),
                child: CreditTermsSectionWidget(
                  screenData: screenData,
                  parentSetState: setState,
                ),
              ),

            DocumentTotalWidget(screenData: screenData, showOnlyTotalValue: false),
            TransactionEditProductListWidget(screenData: screenData),

            // 👤 Section: ข้อมูลผู้สร้าง/ผู้แก้ไข (สำหรับ PO, QT และต้องมี docno)
            // ดึงข้อมูลจาก MongoDB datahistory + fallback จากข้อมูลเอกสาร
            if ((transactionType == global.TransactionTypeEnum.purchaseorder || transactionType == global.TransactionTypeEnum.quotation) && screenData.docno.isNotEmpty)
              DocumentCreatorInfoWidget(
                docNo: screenData.docno,
                fallbackCreatorCode: screenData.creatorcode,
                fallbackCreatorName: screenData.creatorname,
                fallbackCreatedAt: screenData.createdat,
              ),

            // 📊 Section: การไหลของเอกสาร (Document Flow) - หลังรายการสินค้า
            if (screenData.docno.isNotEmpty)
              DocFlowWidget(
                docno: screenData.docno,
                transflag: screenData.transflag,
                onDocumentTap: (docno, transflag) {
                  // TODO: Navigate to document or show document details
                  if (kDebugMode) {
                    print('Tapped document: $docno, transflag: $transflag');
                  }
                },
              ),

          ],
        ),
      ),
    );
  }
}

// Helper function for backward compatibility
Widget editDocumentWidget({
  required TransactionModel screenData,
  required global.TransactionTypeEnum transactionType,
  required Function(void Function()) setState,
  required BuildContext context,
  required TextEditingController custCodeController,
  required TextEditingController custnamesController,
  required TextEditingController saleCodeController,
  required TextEditingController saleNameController,
  required TextEditingController docRefNumberController,
  required TextEditingController taxDocNoController,
  required TextEditingController vatRateController,
  required TextEditingController descriptionController,
  required TextEditingController transportAmountController,
  required TextEditingController docDateController,
  required TextEditingController docTimeController,
  required bool docDateTimeValidated,
  required global.Debouncer debouncer,
  required Function() calTotalValue,
  required Function() headerTableDetail,
  required Function({required String word}) searchCustomer,
  required Function({required String word}) searchSupplier,
  required Function({required String word}) searchSale,
  required Function({required String word}) searchSaleChannel,
  required Function({required String word}) searchDocRef,
  Function(int index, String docno)? onDeleteDocReference,
}) {
  return DocumentHeaderWidget(
    screenData: screenData,
    transactionType: transactionType,
    setState: setState,
    context: context,
    custCodeController: custCodeController,
    custnamesController: custnamesController,
    saleCodeController: saleCodeController,
    saleNameController: saleNameController,
    docRefNumberController: docRefNumberController,
    taxDocNoController: taxDocNoController,
    vatRateController: vatRateController,
    descriptionController: descriptionController,
    transportAmountController: transportAmountController,
    docDateController: docDateController,
    docTimeController: docTimeController,
    docDateTimeValidated: docDateTimeValidated,
    debouncer: debouncer,
    calTotalValue: calTotalValue,
    headerTableDetail: headerTableDetail,
    searchCustomer: searchCustomer,
    searchSupplier: searchSupplier,
    searchSale: searchSale,
    searchSaleChannel: searchSaleChannel,
    searchDocRef: searchDocRef,
    onDeleteDocReference: onDeleteDocReference,
  );
}
