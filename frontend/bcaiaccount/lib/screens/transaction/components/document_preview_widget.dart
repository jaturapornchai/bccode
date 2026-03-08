import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/screens/transaction/components/document_total_widget.dart';
import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;
import 'package:printing/printing.dart';
// import 'package:intl/intl.dart'; // Not used

// Enum สำหรับกำหนดรูปแบบการแสดงผล
enum ViewMode { list, table }

// Configuration class for different transaction types
class TransactionDisplayConfig {
  final global.TransactionTypeEnum transactionType;

  TransactionDisplayConfig(this.transactionType);

  // ==================== การตั้งค่าส่วนหัวเอกสาร ====================

  /// แสดงข้อมูลลูกค้า/ผู้จำหน่าย
  /// ซ่อนสำหรับ: การโอนสต๊อก
  bool get showCustomerInfo => transactionType != global.TransactionTypeEnum.stocktransfer;

  /// แสดงข้อมูลภาษีมูลค่าเพิ่มและประเภทการสอบถาม
  /// ซ่อนสำหรับ: การจัดการสต๊อก, การปรับปรุง, การชำระเงินล่วงหน้า, การคืนเงินล่วงหน้า
  bool get showVatAndInquiryType =>
      !(transactionType == global.TransactionTypeEnum.stocktransfer ||
          transactionType == global.TransactionTypeEnum.stockreceiveproduct ||
          transactionType == global.TransactionTypeEnum.stockpickupproduct ||
          transactionType == global.TransactionTypeEnum.stockreturnproduct ||
          transactionType == global.TransactionTypeEnum.adjust ||
          transactionType == global.TransactionTypeEnum.advancePayment ||
          transactionType == global.TransactionTypeEnum.advancePaymentRefund);

  /// แสดงข้อมูลเอกสารอ้างอิง
  /// ซ่อนสำหรับ: การจัดการสต๊อก, การปรับปรุง, การชำระเงินล่วงหน้า, การคืนเงินล่วงหน้า
  bool get showRefDocument =>
      !(transactionType == global.TransactionTypeEnum.stocktransfer ||
          transactionType == global.TransactionTypeEnum.stockreceiveproduct ||
          transactionType == global.TransactionTypeEnum.stockpickupproduct ||
          transactionType == global.TransactionTypeEnum.stockreturnproduct ||
          transactionType == global.TransactionTypeEnum.adjust ||
          transactionType == global.TransactionTypeEnum.advancePayment);

  /// แสดงข้อมูลเอกสารภาษี
  /// ซ่อนสำหรับ: การจัดการสต๊อก, การปรับปรุง, การชำระเงินล่วงหน้า, การคืนเงินล่วงหน้า
  bool get showTaxDocument =>
      !(transactionType == global.TransactionTypeEnum.stocktransfer ||
          transactionType == global.TransactionTypeEnum.stockreceiveproduct ||
          transactionType == global.TransactionTypeEnum.stockpickupproduct ||
          transactionType == global.TransactionTypeEnum.stockreturnproduct ||
          transactionType == global.TransactionTypeEnum.adjust ||
          transactionType == global.TransactionTypeEnum.advancePayment ||
          transactionType == global.TransactionTypeEnum.advancePaymentRefund);

  // ==================== การตั้งค่าตารางรายการสินค้า ====================

  /// แสดงคอลัมน์ราคาในตารางสินค้า (จำนวน, ราคา, มูลค่ารวม)
  /// ซ่อนสำหรับ: การจัดการสต๊อก, การปรับปรุง, ใบสั่ง, การชำระเงินล่วงหน้า, การคืนเงินล่วงหน้า
  bool get showPriceColumns =>
      !(transactionType == global.TransactionTypeEnum.stocktransfer ||
          transactionType == global.TransactionTypeEnum.stockpickupproduct ||
          transactionType == global.TransactionTypeEnum.stockreturnproduct ||
          transactionType == global.TransactionTypeEnum.adjust ||
          transactionType == global.TransactionTypeEnum.saleorder ||
          transactionType == global.TransactionTypeEnum.quotation ||
          transactionType == global.TransactionTypeEnum.advancePayment ||
          transactionType == global.TransactionTypeEnum.advancePaymentRefund);

  /// ใช้รูปแบบตารางพิเศษสำหรับการชำระเงินล่วงหน้า (ลำดับ, รายละเอียด, มูลค่า)
  bool get isAdvancePayment => transactionType == global.TransactionTypeEnum.advancePayment || transactionType == global.TransactionTypeEnum.advancePaymentRefund;

  // ==================== การตั้งค่าสรุปยอดการเงิน ====================

  /// แสดงส่วนสรุปยอดการเงิน
  /// ซ่อนสำหรับ: การจัดการสต๊อก, การปรับปรุง
  bool get showFinancialSummary =>
      !(transactionType == global.TransactionTypeEnum.stocktransfer ||
          transactionType == global.TransactionTypeEnum.stockpickupproduct ||
          transactionType == global.TransactionTypeEnum.stockreturnproduct ||
          transactionType == global.TransactionTypeEnum.adjust);

  /// แสดงเฉพาะมูลค่ารวมในสรุปยอดการเงิน (สำหรับการชำระเงินล่วงหน้า, การคืนเงินล่วงหน้า)
  bool get showOnlyTotalValue => transactionType == global.TransactionTypeEnum.advancePayment || transactionType == global.TransactionTypeEnum.advancePaymentRefund;
}

class DocumentPreviewWidget extends StatefulWidget {
  final TransactionModel screenData;
  final global.TransactionTypeEnum transactionType;
  final bool docDateTimeValidated;
  final String fieldNameCustCode;
  final String fieldNameCustName;

  const DocumentPreviewWidget({
    super.key,
    required this.screenData,
    required this.transactionType,
    required this.docDateTimeValidated,
    required this.fieldNameCustCode,
    required this.fieldNameCustName,
  });

  @override
  State<DocumentPreviewWidget> createState() => _DocumentPreviewWidgetState();
}

class _DocumentPreviewWidgetState extends State<DocumentPreviewWidget> {
  // สถานะการแสดงผล เริ่มต้นเป็น table ตามที่ผู้ใช้ต้องการ
  ViewMode _viewMode = ViewMode.table;

  @override
  Widget build(BuildContext context) {
    return Column(children: _buildPreviewContent());
  }

  // Configuration methods for different transaction types
  TransactionDisplayConfig get _config => TransactionDisplayConfig(widget.transactionType);

  List<Widget> _buildPreviewContent() {
    List<Widget> previewResult = [];

    // Scrollable content section
    previewResult.add(_buildScrollableContent());

    // Financial summary section (fixed at bottom)
    if (_config.showFinancialSummary) {
      _buildFinancialSummary(previewResult);
    }

    return previewResult;
  }

  Widget _buildScrollableContent() {
    return Expanded(
      child: SingleChildScrollView(
        child: Column(
          children: [
            // Header section
            _buildHeaderSectionContent(),

            // Product list section
            _buildProductTableContent(),
          ],
        ),
      ),
    );
  }

  Widget _buildHeaderSectionContent() {
    final allFields = _getAllFieldValues();

    if (allFields.isEmpty) return const SizedBox();

    return Container(
      width: double.infinity,
      margin: const EdgeInsets.all(4.0),
      padding: const EdgeInsets.all(8.0),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.grey[300]!),
      ),
      child: LayoutBuilder(
        builder: (context, constraints) {
          return Wrap(
            spacing: 12.0,
            runSpacing: 6.0,
            crossAxisAlignment: WrapCrossAlignment.center,
            children: allFields.map((field) {
              return ConstrainedBox(
                constraints: BoxConstraints(maxWidth: constraints.maxWidth),
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Padding(
                      padding: const EdgeInsets.only(top: 2.0),
                      child: Icon(field['icon'] as IconData, size: 14, color: field['color'] as Color? ?? Colors.grey[600]),
                    ),
                    const SizedBox(width: 4),
                    Flexible(
                      child: Text.rich(
                        TextSpan(
                          children: [
                            TextSpan(
                              text: "${field['label']}: ",
                              style: TextStyle(fontSize: 12, color: Colors.grey[700], fontWeight: FontWeight.w500),
                            ),
                            TextSpan(
                              text: "${field['value']}",
                              style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: field['color'] as Color? ?? Colors.black87),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ],
                ),
              );
            }).toList(),
          );
        },
      ),
    );
  }

  List<Map<String, dynamic>> _getAllFieldValues() {
    final data = widget.screenData;
    final List<Map<String, dynamic>> fields = [];

    void addField(String label, String? value, IconData icon, {Color? color}) {
      if (value != null && value.isNotEmpty && value != "null") {
        fields.add({'label': label, 'value': value, 'icon': icon, 'color': color});
      }
    }

    // 1. Basic Doc Info
    addField(global.language("doc_number"), data.docno, Icons.numbers, color: Colors.blue[700]);
    addField(
      global.language("doc_date"),
      (widget.docDateTimeValidated)
          ? "${global.dateTimeBuddhist(DateTime.parse(data.docdatetime), format: global.DateTimeFormatEnum.dateDay)} ${global.dateTimeBuddhist(DateTime.parse(data.docdatetime), format: global.DateTimeFormatEnum.time)}"
          : global.language("invalid_date"),
      Icons.calendar_today,
      color: Colors.blue[700],
    );

    // Branch / Shop (Moved to top)
    if (data.branch != null && data.branch!.names != null) {
      addField(global.language("branch"), global.activeLangName(data.branch!.names!), Icons.domain, color: Colors.purple[700]);
    }
    addField("POS ID", data.posid, Icons.point_of_sale);

    // Doc Ref
    if (data.docrefno.isNotEmpty) {
      addField(global.language("doc_ref"), data.docrefno, Icons.link);
      if (data.docrefdate.isNotEmpty) {
        addField(global.language("doc_ref_date"), global.dateTimeBuddhist(DateTime.parse(data.docrefdate), format: global.DateTimeFormatEnum.dateDay), Icons.date_range);
      }
    }

    addField(global.language("document_type"), data.doctype.toString(), Icons.category);

    // 2. Customer / People
    addField(global.language(widget.fieldNameCustCode), data.custcode, Icons.business, color: Colors.green[700]);
    if (data.custnames != null && data.custnames!.isNotEmpty) {
      addField(global.language(widget.fieldNameCustName), global.activeLangName(data.custnames!), Icons.store, color: Colors.green[700]);
    }
    addField(global.language('sale_code'), data.salecode, Icons.person);
    addField(global.language('sale_name'), data.salename, Icons.badge);
    addField("พนักงานขาย (Cashier)", data.cashiername, Icons.support_agent);
    addField("รหัสสมาชิก", data.membercode, Icons.card_membership);
    addField("เบอร์โทรศัพท์", data.customertelephone, Icons.phone);

    // 3. System / Tax
    addField(global.language("vat_type"), global.getVatName(data.vattype), Icons.receipt);
    addField(global.language("inquiry_type"), global.getInquiryName(data.inquirytype), Icons.help_outline);
    addField(global.language('tax_docno'), data.taxdocno, Icons.receipt_long);
    if (data.taxdocdate.isNotEmpty) {
      addField(global.language('tax_doc_date'), global.dateTimeBuddhist(DateTime.parse(data.taxdocdate), format: global.DateTimeFormatEnum.dateDay), Icons.event);
    }
    if (data.vatrate != 0) {
      addField(global.language("var_rate"), "${data.vatrate}%", Icons.percent);
    }
    addField("คำอธิบายส่วนลด", data.discountword, Icons.discount);
    addField(global.language("remark"), data.description, Icons.note);

    // 4. Full Tax Info
    addField("ชื่อลูกค้า (เต็ม)", data.fullvatname, Icons.text_fields);
    addField("เลขผู้เสียภาษี (เต็ม)", data.fullvattaxid, Icons.pin);
    addField("ที่อยู่ (เต็ม)", data.fullvataddress, Icons.home);
    addField("สาขา (เต็ม)", data.fullvatbranchnumber, Icons.storefront);

    // 5. Cancel Info
    if (data.iscancel) {
      addField(global.language("cancel_status"), global.language("cancelled"), Icons.cancel, color: Colors.red);
      addField("เวลายกเลิก", data.canceldatetime, Icons.access_time, color: Colors.red);
      addField("ผู้ยกเลิก", "${data.cancelusercode} ${data.cancelusername}", Icons.person_off, color: Colors.red);
      addField(global.language("cancel_reason"), data.cancelreason, Icons.report, color: Colors.red);
    }

    // 6. Table / Restaurant Info
    addField(global.language("table"), data.tablenumber, Icons.table_restaurant);
    if (data.mancount != null && data.mancount! > 0) {
      addField("จำนวนลูกค้า (ชาย)", data.mancount.toString(), Icons.man);
    }
    if (data.womancount != null && data.womancount! > 0) {
      addField("จำนวนลูกค้า (หญิง)", data.womancount.toString(), Icons.woman);
    }
    if (data.childcount != null && data.childcount! > 0) {
      addField("จำนวนลูกค้า (เด็ก)", data.childcount.toString(), Icons.child_care);
    }
    addField("Buffet Code", data.buffetcode, Icons.food_bank);

    // 7. Delivery / Transport
    if (data.isdelivery == true) {
      addField("จัดส่ง", global.language("yes"), Icons.delivery_dining);
      if (data.deliveryamount != null && data.deliveryamount! > 0) {
        addField(global.language("transportchannel_amount"), global.formatNumber(data.deliveryamount!), Icons.attach_money);
      }
    }
    if (data.istransport == true) {
      addField("ขนส่ง", data.transportcode, Icons.local_shipping);
      if (data.transportamount != null && data.transportamount! > 0) {
        addField("ค่าขนส่ง", global.formatNumber(data.transportamount!), Icons.attach_money);
      }
    }

    // 8. Points / Coupons
    if (data.getpoint != null && data.getpoint! > 0) {
      addField("แต้มที่ได้", data.getpoint.toString(), Icons.star);
    }
    if (data.usepoint != null && data.usepoint! > 0) {
      addField(global.language("point_transaction.used_points"), data.usepoint.toString(), Icons.star_border);
    }
    if (data.totalcouponamount != null && data.totalcouponamount! > 0) {
      addField("มูลค่าคูปองรวม", global.formatNumber(data.totalcouponamount!), Icons.confirmation_number);
    }

    // 9. Other
    addField("Sale Channel", data.salechannelcode, Icons.router);
    if (data.slipurl != null && data.slipurl!.isNotEmpty) {
      addField("Slip URL", "มีรูปสลิป", Icons.image);
    }

    return fields;
  }

  Widget _buildProductTableContent() {
    if (_config.isAdvancePayment) {
      return _buildAdvancePaymentTableContent();
    } else if (_config.showPriceColumns) {
      return _buildFullProductTableContent();
    } else {
      return _buildSimpleProductTableContent();
    }
  }

  Widget _buildAdvancePaymentTableContent() {
    return Column(
      children: [
        // Header Card with buttons
        Card(
          margin: const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
          elevation: 4,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          child: Container(
            decoration: BoxDecoration(
              gradient: LinearGradient(begin: Alignment.topLeft, end: Alignment.bottomRight, colors: [Colors.blue[600]!, Colors.blue[800]!]),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Padding(
              padding: const EdgeInsets.all(12.0),
              child: Row(
                children: [
                  Icon(Icons.shopping_cart, color: Colors.white, size: 20),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      '${global.language("product_list")} ${widget.screenData.details!.length} ${global.language("items")}',
                      style: const TextStyle(fontWeight: FontWeight.bold, color: Colors.white, fontSize: 14),
                    ),
                  ),
                  // Export PDF button
                  IconButton(
                    icon: const Icon(Icons.picture_as_pdf, color: Colors.white, size: 20),
                    tooltip: global.language('export_pdf'),
                    onPressed: _exportPdf,
                  ),
                ],
              ),
            ),
          ),
        ),
        // Column headers
        Container(
          margin: const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
          decoration: BoxDecoration(
            color: Colors.blue[50],
            borderRadius: BorderRadius.circular(6),
            border: Border.all(color: Colors.blue[200]!),
          ),
          child: Padding(
            padding: EdgeInsets.all(6.0),
            child: Row(
              children: [
                SizedBox(
                  width: 50,
                  child: Text(
                    global.language('pdf_sequence'),
                    style: TextStyle(fontWeight: FontWeight.bold, fontSize: 11),
                    textAlign: TextAlign.center,
                  ),
                ),
                Expanded(
                  flex: 3,
                  child: Text(global.language('product'), style: TextStyle(fontWeight: FontWeight.bold, fontSize: 11)),
                ),
                Expanded(
                  child: Text(
                    global.language('enter_qty'),
                    textAlign: TextAlign.right,
                    style: TextStyle(fontWeight: FontWeight.bold, fontSize: 11),
                  ),
                ),
                Expanded(
                  child: Text(
                    global.language('price'),
                    textAlign: TextAlign.right,
                    style: TextStyle(fontWeight: FontWeight.bold, fontSize: 11),
                  ),
                ),
                Expanded(
                  child: Text(
                    global.language('receivelist_value'),
                    textAlign: TextAlign.right,
                    style: TextStyle(fontWeight: FontWeight.bold, fontSize: 11),
                  ),
                ),
              ],
            ),
          ),
        ),
        // Product Items
        ...widget.screenData.details!.map((item) {
          int index = widget.screenData.details!.indexOf(item);
          bool isEvenRow = index % 2 == 0;
          return Container(
            margin: const EdgeInsets.symmetric(vertical: 1, horizontal: 4),
            decoration: BoxDecoration(
              color: isEvenRow ? Colors.grey[50] : Colors.white,
              borderRadius: BorderRadius.circular(6),
              border: Border.all(color: Colors.grey[200]!),
            ),
            child: Padding(
              padding: const EdgeInsets.all(6.0),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  SizedBox(
                    width: 50,
                    child: Text(
                      (index + 1).toString(),
                      textAlign: TextAlign.center,
                      style: const TextStyle(fontWeight: FontWeight.bold),
                    ),
                  ),
                  Expanded(
                    flex: 3,
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(global.activeLangName(item.itemnames!), style: const TextStyle(fontWeight: FontWeight.w500)),
                        if (item.description!.isNotEmpty)
                          Text(
                            'หมายเหตุ: ${item.description!}',
                            style: TextStyle(fontSize: 12, color: Colors.grey[600], fontStyle: FontStyle.italic),
                          ),
                      ],
                    ),
                  ),
                  Expanded(
                    child: Text(
                      global.formatQuantity(item.qty).toString(),
                      textAlign: TextAlign.right,
                      style: TextStyle(fontWeight: FontWeight.w600, fontSize: 11, color: Colors.orange[700]),
                    ),
                  ),
                  Expanded(
                    child: Text(
                      global.formatUnitPrice(item.price).toString(),
                      textAlign: TextAlign.right,
                      style: TextStyle(fontWeight: FontWeight.w600, fontSize: 11, color: Colors.purple[700]),
                    ),
                  ),
                  Expanded(
                    child: Text(
                      global.formatNumber(item.sumamount).toString(),
                      textAlign: TextAlign.right,
                      style: TextStyle(fontWeight: FontWeight.bold, fontSize: 11, color: Colors.green[700]),
                    ),
                  ),
                ],
              ),
            ),
          );
        }),
      ],
    );
  }

  Widget _buildFullProductTableContent() {
    return Column(
      children: [
        // Header Card
        Card(
          margin: const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
          elevation: 4,
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
          child: Container(
            decoration: BoxDecoration(
              gradient: LinearGradient(begin: Alignment.topLeft, end: Alignment.bottomRight, colors: [Colors.blue[600]!, Colors.blue[800]!]),
              borderRadius: BorderRadius.circular(12),
            ),
            child: Padding(
              padding: EdgeInsets.all(12.0),
              child: Row(
                children: [
                  Icon(Icons.shopping_cart, color: Colors.white, size: 20),
                  const SizedBox(width: 8),
                  Expanded(
                    child: Text(
                      '${global.language("product_list")} ${widget.screenData.details!.length} ${global.language("items")}',
                      style: const TextStyle(fontWeight: FontWeight.bold, color: Colors.white, fontSize: 14),
                    ),
                  ),
                  // Toggle view mode button
                  IconButton(
                    icon: Icon(_viewMode == ViewMode.list ? Icons.table_chart : Icons.view_list, color: Colors.white, size: 20),
                    tooltip: _viewMode == ViewMode.list ? global.language('show_as_table') : global.language('show_as_list'),
                    onPressed: () {
                      setState(() {
                        _viewMode = _viewMode == ViewMode.list ? ViewMode.table : ViewMode.list;
                      });
                    },
                  ),
                  const SizedBox(width: 4),
                  // Export PDF button
                  IconButton(
                    icon: const Icon(Icons.picture_as_pdf, color: Colors.white, size: 20),
                    tooltip: global.language('export_pdf'),
                    onPressed: _exportPdf,
                  ),
                ],
              ),
            ),
          ),
        ),
        // Product Items - สลับแสดงตาม viewMode
        if (_viewMode == ViewMode.list)
          // List View (แบบ Card เดิม)
          ...widget.screenData.details!.map((item) {
            int index = widget.screenData.details!.indexOf(item);
            return Card(
              margin: const EdgeInsets.symmetric(vertical: 4, horizontal: 8),
              elevation: 2,
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
              child: Container(
                decoration: BoxDecoration(
                  borderRadius: BorderRadius.circular(12),
                  gradient: LinearGradient(begin: Alignment.topLeft, end: Alignment.bottomRight, colors: [Colors.indigo[50]!, Colors.white]),
                ),
                child: Padding(
                  padding: const EdgeInsets.all(8.0),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      // Header row with item number and icon
                      Row(
                        children: [
                          Container(
                            width: 24,
                            height: 24,
                            decoration: BoxDecoration(
                              gradient: LinearGradient(colors: [Colors.blue[400]!, Colors.blue[600]!]),
                              borderRadius: BorderRadius.circular(12),
                            ),
                            child: Center(
                              child: Text(
                                (index + 1).toString(),
                                style: const TextStyle(fontWeight: FontWeight.bold, color: Colors.white, fontSize: 10),
                              ),
                            ),
                          ),
                          const SizedBox(width: 8),
                          Icon(Icons.inventory_2, color: Colors.indigo[600], size: 16),
                          const SizedBox(width: 4),
                          Text(
                            'สินค้าลำดับที่ ${index + 1}',
                            style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: Colors.indigo[700]),
                          ),
                        ],
                      ),
                      const Divider(height: 12),

                      // Product information using _buildInfoRow like header cards
                      _buildInfoRow(Icons.shopping_bag, global.language('import_product_detail.product_name'), global.activeLangName(item.itemnames!)),

                      if (item.itemcode.isNotEmpty) _buildInfoRow(Icons.qr_code_scanner, global.language('product_code'), item.itemcode),

                      if (item.barcode.isNotEmpty) _buildInfoRow(Icons.qr_code, global.language('barcode'), item.barcode),

                      _buildInfoRow(Icons.warehouse, global.language('warehouse'), global.language('default_warehouse')),

                      _buildInfoRow(Icons.location_on, global.language('area'), item.locationcode.isEmpty ? 'A001' : item.locationcode),

                      _buildInfoRow(Icons.numbers, global.language('enter_qty'), '${global.formatQuantity(item.qty)} ${item.unitcode.isEmpty ? global.language('chatbot_product_unit') : item.unitcode}'),

                      // ราคาต่อหน่วย — แสดง 2 สกุลเงินถ้า multi-currency
                      _isMultiCurrency()
                          ? _buildDualCurrencyInfoRow(Icons.attach_money, global.language('price_per_unit'), item.priceDoc ?? (_safeRate() > 0 ? item.price / _safeRate() : 0), item.price)
                          : _buildInfoRow(Icons.attach_money, global.language('price_per_unit'), global.formatUnitPrice(item.price).toString()),

                      if (item.discountamount > 0) _buildInfoRow(Icons.discount, global.language('discount'), global.formatNumber(item.discountamount).toString()),

                      // มูลค่ารวม — แสดง 2 สกุลเงินถ้า multi-currency
                      _isMultiCurrency()
                          ? _buildDualCurrencyInfoRow(Icons.calculate, global.language('payment_daily_total_amount_column'), item.sumAmountDoc ?? (_safeRate() > 0 ? item.sumamount / _safeRate() : 0), item.sumamount)
                          : _buildInfoRow(Icons.calculate, global.language('payment_daily_total_amount_column'), global.formatNumber(item.sumamount).toString()),

                      if (item.description!.isNotEmpty) _buildInfoRow(Icons.note, global.language('remark'), item.description!),
                      // Extra items (if any)
                      if (item.extrajsonlist!.isNotEmpty) ...[
                        const SizedBox(height: 8),
                        Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: Colors.orange[50],
                            borderRadius: BorderRadius.circular(8),
                            border: Border.all(color: Colors.orange[200]!),
                          ),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Row(
                                children: [
                                  Icon(Icons.add_circle, color: Colors.orange[700], size: 16),
                                  const SizedBox(width: 4),
                                  Text(
                                    global.language('additional_items'),
                                    style: TextStyle(fontWeight: FontWeight.bold, color: Colors.orange[700], fontSize: 12),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 4),
                              ...item.extrajsonlist!.map((extra) {
                                return Padding(
                                  padding: const EdgeInsets.symmetric(vertical: 2.0),
                                  child: Row(
                                    children: [
                                      Expanded(
                                        flex: 3,
                                        child: Text("• ${global.activeLangName(extra.itemnames!)}", style: TextStyle(fontSize: 11, color: Colors.grey[700])),
                                      ),
                                      SizedBox(
                                        width: 40,
                                        child: Text(
                                          global.formatQuantity((extra.qty! != 0.0) ? extra.qty! : 1.0).toString(),
                                          textAlign: TextAlign.right,
                                          style: TextStyle(fontSize: 11, color: Colors.grey[700]),
                                        ),
                                      ),
                                      const SizedBox(width: 8),
                                      SizedBox(
                                        width: 40,
                                        child: Text(
                                          global.formatUnitPrice(extra.price!).toString(),
                                          textAlign: TextAlign.right,
                                          style: TextStyle(fontSize: 11, color: Colors.grey[700]),
                                        ),
                                      ),
                                    ],
                                  ),
                                );
                              }),
                            ],
                          ),
                        ),
                      ],
                    ],
                  ),
                ),
              ),
            );
          }),

        // Table View
        if (_viewMode == ViewMode.table) _buildTableView(),
      ],
    );
  }

  /// สร้าง Table View สำหรับแสดงรายการสินค้าแบบตาราง (ความกว้าง 100%)
  Widget _buildTableView() {
    // คำนวณยอดรวมทั้งหมด
    double totalQty = 0;
    double totalAmount = 0;
    double totalDiscountAmount = 0;
    double totalAmountDoc = 0;

    final isMulti = _isMultiCurrency();
    final rate = _safeRate();

    for (var item in widget.screenData.details!) {
      totalQty += item.qty;
      totalAmount += item.sumamount;
      totalDiscountAmount += item.discountamount;
      if (isMulti) {
        totalAmountDoc += (item.sumAmountDoc ?? (rate > 0 ? item.sumamount / rate : 0));
      }
    }

    return Padding(
      padding: EdgeInsets.symmetric(horizontal: 8.0, vertical: 4.0),
      child: Table(
        border: TableBorder.all(color: Colors.grey[300]!, width: 1),
        columnWidths: const {
          0: FixedColumnWidth(40), // #
          1: FlexColumnWidth(1), // รหัสสินค้า
          2: FlexColumnWidth(1), // บาร์โค้ด
          3: FlexColumnWidth(2.5), // ชื่อสินค้า
          4: FlexColumnWidth(0.8), // หน่วย
          5: FlexColumnWidth(0.8), // จำนวน
          6: FlexColumnWidth(1), // ราคา/หน่วย
          7: FlexColumnWidth(0.8), // ส่วนลด
          8: FlexColumnWidth(1.2), // มูลค่ารวม
        },
        children: [
          // Header Row
          TableRow(
            decoration: BoxDecoration(color: Colors.blue[100]),
            children: [
              _buildTableHeaderCell('#'),
              _buildTableHeaderCell(global.language('product_code')),
              _buildTableHeaderCell(global.language('barcode')),
              _buildTableHeaderCell(global.language('import_product_detail.product_name')),
              _buildTableHeaderCell(global.language('chatbot_product_unit')),
              _buildTableHeaderCell(global.language('enter_qty'), align: TextAlign.right),
              _buildTableHeaderCell(isMulti ? 'ราคา/หน่วย (${_getDocSymbol()})' : global.language('pdf_price_per_unit'), align: TextAlign.right),
              _buildTableHeaderCell(global.language('discount'), align: TextAlign.right),
              _buildTableHeaderCell(isMulti ? 'มูลค่ารวม (${_getDocSymbol()})' : global.language('payment_daily_total_amount_column'), align: TextAlign.right),
            ],
          ),
          // Data Rows
          ...widget.screenData.details!.asMap().entries.map((entry) {
            int index = entry.key;
            var item = entry.value;
            bool isEven = index % 2 == 0;

            return TableRow(
              decoration: BoxDecoration(color: isEven ? Colors.grey[50] : Colors.white),
              children: [
                _buildTableDataCell('${index + 1}', align: TextAlign.center, maxLines: null, overflow: TextOverflow.visible),
                _buildTableDataCell(item.itemcode, maxLines: null, overflow: TextOverflow.visible),
                _buildTableDataCell(item.barcode, maxLines: null, overflow: TextOverflow.visible),
                // ชื่อสินค้า พร้อมรายละเอียดเพิ่มเติม (ถ้ามี)
                _buildNameWithDescriptionCell(
                  global.activeLangName(item.itemnames!),
                  item.description,
                  item.remark,
                ),
                _buildTableDataCell(item.unitcode, maxLines: null, overflow: TextOverflow.visible),
                _buildTableDataCell(global.formatQuantity(item.qty), align: TextAlign.right, maxLines: null, overflow: TextOverflow.visible),
                // ราคา — แสดง 2 สกุลเงินถ้า multi-currency
                isMulti
                    ? _buildDualCurrencyTableCell(item.priceDoc ?? (rate > 0 ? item.price / rate : 0), item.price)
                    : _buildTableDataCell(global.formatUnitPrice(item.price), align: TextAlign.right, maxLines: null, overflow: TextOverflow.visible),
                _buildTableDataCell(item.discount, align: TextAlign.right, maxLines: null, overflow: TextOverflow.visible),
                // มูลค่ารวม — แสดง 2 สกุลเงินถ้า multi-currency
                isMulti
                    ? _buildDualCurrencyTableCell(item.sumAmountDoc ?? (rate > 0 ? item.sumamount / rate : 0), item.sumamount, isBold: true)
                    : _buildTableDataCell(global.formatNumber(item.sumamount), align: TextAlign.right, bold: true, maxLines: null, overflow: TextOverflow.visible),
              ],
            );
          }),
          // Footer Row - ยอดรวมด้านล่าง
          TableRow(
            decoration: BoxDecoration(color: Colors.blue[50]),
            children: [
              _buildTableFooterCell('', align: TextAlign.center),
              _buildTableFooterCell(''),
              _buildTableFooterCell(''),
              _buildTableFooterCell(global.language('total'), bold: true),
              _buildTableFooterCell(''),
              _buildTableFooterCell(global.formatNumber(totalQty), align: TextAlign.right, bold: true),
              _buildTableFooterCell(''),
              _buildTableFooterCell(global.formatNumber(totalDiscountAmount), align: TextAlign.right, bold: true),
              // ยอดรวม — แสดง 2 สกุลเงินถ้า multi-currency
              isMulti
                  ? _buildDualCurrencyTableCell(totalAmountDoc, totalAmount, isBold: true)
                  : _buildTableFooterCell(global.formatNumber(totalAmount), align: TextAlign.right, bold: true, color: Colors.blue[800]),
            ],
          ),
        ],
      ),
    );
  }

  /// สร้าง cell สำหรับชื่อสินค้าพร้อมรายละเอียดเพิ่มเติม (ถ้ามี)
  Widget _buildNameWithDescriptionCell(String itemName, String? description, String remark) {
    // ตรวจสอบว่ามีรายละเอียดเพิ่มเติมหรือไม่ (description หรือ remark)
    final hasDescription = description != null && description.isNotEmpty;
    final hasRemark = remark.isNotEmpty;
    final additionalInfo = hasDescription ? description : (hasRemark ? remark : null);

    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4.0, vertical: 4.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          // ชื่อสินค้า
          Text(
            itemName,
            style: const TextStyle(fontSize: 11, color: Colors.black87),
          ),
          // รายละเอียดเพิ่มเติม (ถ้ามี)
          if (additionalInfo != null && additionalInfo.isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(top: 2),
              child: Text(
                additionalInfo,
                style: TextStyle(
                  color: Colors.blue[700],
                  fontSize: 10,
                  fontStyle: FontStyle.italic,
                ),
              ),
            ),
        ],
      ),
    );
  }

  /// สร้าง Footer Cell สำหรับแถวยอดรวม
  Widget _buildTableFooterCell(String text, {TextAlign align = TextAlign.left, bool bold = false, Color? color}) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4.0, vertical: 6.0),
      child: Text(
        text,
        style: TextStyle(
          fontSize: 11,
          fontWeight: bold ? FontWeight.bold : FontWeight.normal,
          color: color ?? Colors.black87,
        ),
        textAlign: align,
      ),
    );
  }

  /// สร้าง Header Cell สำหรับตาราง
  Widget _buildTableHeaderCell(String text, {TextAlign align = TextAlign.left}) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 6.0, vertical: 4.0),
      child: Text(
        text,
        style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 11, color: Colors.black87),
        textAlign: align,
      ),
    );
  }

  /// สร้าง Data Cell สำหรับตาราง
  Widget _buildTableDataCell(
    String text, {
    TextAlign align = TextAlign.left,
    int? maxLines = 1, // เปลี่ยนเป็น nullable เพื่อให้แสดงได้ไม่จำกัดบรรทัด
    bool bold = false,
    TextOverflow overflow = TextOverflow.ellipsis, // เพิ่ม parameter overflow
  }) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4.0, vertical: 2.0),
      child: Text(
        text,
        style: TextStyle(fontSize: 10, fontWeight: bold ? FontWeight.bold : FontWeight.normal),
        textAlign: align,
        overflow: overflow, // ใช้ overflow ที่ส่งมา
        maxLines: maxLines, // ถ้าเป็น null จะแสดงได้ไม่จำกัดบรรทัด
      ),
    );
  }

  /// สร้าง Table Cell แสดง 2 สกุลเงิน (doc=น้ำเงิน, base=เทา)
  Widget _buildDualCurrencyTableCell(double docValue, double baseValue, {bool isBold = false}) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 4.0, vertical: 2.0),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.end,
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            '${_getDocSymbol()} ${global.formatNumber(docValue)}',
            style: TextStyle(fontSize: 10, fontWeight: isBold ? FontWeight.bold : FontWeight.normal, color: Colors.blue[700]),
            textAlign: TextAlign.right,
          ),
          Text(
            '${_getBaseSymbol()} ${global.formatNumber(baseValue)}',
            style: TextStyle(fontSize: 8, color: Colors.grey[500], fontStyle: FontStyle.italic),
            textAlign: TextAlign.right,
          ),
        ],
      ),
    );
  }

  /// Export เอกสารเป็น PDF
  Future<void> _exportPdf() async {
    try {
      // โหลด Thai font
      final fontData = await rootBundle.load('assets/fonts/THSarabunNew.ttf');
      final fontBoldData = await rootBundle.load('assets/fonts/THSarabunNew-Bold.ttf');
      final ttf = pw.Font.ttf(fontData);
      final ttfBold = pw.Font.ttf(fontBoldData);

      final pdf = pw.Document();

      // สร้าง PDF page (A4 แบบบีบระหว่างบรรทัด)
      pdf.addPage(
        pw.MultiPage(
          pageFormat: PdfPageFormat.a4,
          margin: const pw.EdgeInsets.all(20), // ลด margin จาก 32 เป็น 20
          theme: pw.ThemeData.withFont(base: ttf, bold: ttfBold),
          build: (pw.Context context) {
            return [
              // Header
              _buildPdfHeader(ttfBold),
              pw.SizedBox(height: 8), // ลดจาก 20 เป็น 8
              // ข้อมูลเอกสารแบบละเอียด
              _buildPdfDocumentInfoDetailed(ttf, ttfBold),
              pw.SizedBox(height: 6), // ลดจาก 15 เป็น 6
              // ข้อมูลลูกค้า/ผู้จำหน่ายแบบละเอียด
              if (_config.showCustomerInfo) ...[
                _buildPdfCustomerInfoDetailed(ttf, ttfBold),
                pw.SizedBox(height: 6), // ลดจาก 15 เป็น 6
              ],

              // ตารางสินค้าแบบละเอียด
              _buildPdfProductTableDetailed(ttf, ttfBold),
              pw.SizedBox(height: 6), // ลดจาก 15 เป็น 6
              // สรุปยอดแบบละเอียด
              if (_config.showFinancialSummary) _buildPdfSummaryDetailed(ttf, ttfBold),
            ];
          },
          footer: (pw.Context context) {
            return _buildPdfFooter(context, ttf);
          },
        ),
      );

      // แสดง PDF ใน popup dialog
      final pdfBytes = await pdf.save();

      if (mounted) {
        final dialogWidth = MediaQuery.of(context).size.width * 0.9;
        final dialogHeight = MediaQuery.of(context).size.height * 0.9;

        showDialog(
          context: context,
          builder: (BuildContext dialogContext) {
            return Dialog(
              child: SizedBox(
                width: dialogWidth,
                height: dialogHeight,
                child: Padding(
                  padding: const EdgeInsets.all(16),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      // Header
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text(global.language('pdf_preview'), style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                          Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              // ปุ่ม Print
                              IconButton(
                                icon: const Icon(Icons.print),
                                tooltip: global.language('print'),
                                onPressed: () async {
                                  await Printing.layoutPdf(onLayout: (PdfPageFormat format) async => pdfBytes);
                                },
                              ),
                              // ปุ่ม Download
                              IconButton(
                                icon: const Icon(Icons.download),
                                tooltip: global.language('download'),
                                onPressed: () async {
                                  await Printing.sharePdf(bytes: pdfBytes, filename: 'document_${widget.screenData.docno}.pdf');
                                },
                              ),
                              // ปุ่มปิด
                              IconButton(
                                icon: Icon(Icons.close),
                                tooltip: global.language('close'),
                                onPressed: () {
                                  Navigator.of(dialogContext).pop();
                                },
                              ),
                            ],
                          ),
                        ],
                      ),
                      const Divider(),
                      // PDF Viewer - ใช้ SizedBox แทน Expanded เพื่อหลีกเลี่ยง ParentDataWidget error
                      SizedBox(
                        height: dialogHeight - 100,
                        child: PdfPreview(
                          build: (format) async => pdfBytes,
                          allowPrinting: false,
                          allowSharing: false,
                          canChangePageFormat: false,
                          canChangeOrientation: false,
                          canDebug: false,
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            );
          },
        );
      }
    } catch (e) {
      if (mounted) {
        global.showErrorSnackBar(context, 'เกิดข้อผิดพลาดในการสร้าง PDF: $e');
      }
    }
  }

  /// สร้าง PDF Header — ใช้ language keys
  pw.Widget _buildPdfHeader(pw.Font fontBold) {
    String docTitle = global.language('purchase_order');
    if (widget.transactionType == global.TransactionTypeEnum.sale) {
      docTitle = global.language('sale_order');
    } else if (widget.transactionType == global.TransactionTypeEnum.purchase) {
      docTitle = global.language('purchase_order');
    }

    return pw.Column(
      crossAxisAlignment: pw.CrossAxisAlignment.start,
      children: [
        pw.Text(
          docTitle,
          style: pw.TextStyle(fontSize: 18, font: fontBold), // ลดจาก 24 เป็น 18
        ),
        pw.Divider(thickness: 1.5), // ลดจาก 2 เป็น 1.5
      ],
    );
  }

  /// สร้างข้อมูลเอกสารแบบละเอียด รองรับ Multi-Currency + language keys
  pw.Widget _buildPdfDocumentInfoDetailed(pw.Font font, pw.Font fontBold) {
    final isMulti = _isMultiCurrency();
    final sd = widget.screenData;

    return pw.Container(
      padding: const pw.EdgeInsets.all(8),
      decoration: pw.BoxDecoration(
        border: pw.Border.all(color: PdfColors.grey400),
        borderRadius: const pw.BorderRadius.all(pw.Radius.circular(4)),
      ),
      child: pw.Column(
        crossAxisAlignment: pw.CrossAxisAlignment.start,
        children: [
          pw.Text(global.language('document_info'), style: pw.TextStyle(fontSize: 14, font: fontBold)),
          pw.Divider(),
          pw.Row(
            mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
            children: [
              pw.Column(
                crossAxisAlignment: pw.CrossAxisAlignment.start,
                children: [
                  pw.Text('${global.language("doc_number")}: ${sd.docno}', style: pw.TextStyle(font: font, fontSize: 11)),
                  pw.SizedBox(height: 2),
                  pw.Text('${global.language("doc_date")}: ${_formatDate(sd.docdatetime)}', style: pw.TextStyle(font: font, fontSize: 11)),
                  pw.SizedBox(height: 2),
                  pw.Text('${global.language("doc_ref")}: ${sd.docrefno ?? "-"}', style: pw.TextStyle(font: font, fontSize: 11)),
                  if (isMulti) ...[
                    pw.SizedBox(height: 2),
                    pw.Text(
                      '${global.language("currency")}: ${sd.docCurrency ?? ''} > ${sd.currency ?? ''}',
                      style: pw.TextStyle(font: font, fontSize: 11, color: PdfColors.blue700),
                    ),
                  ],
                ],
              ),
              pw.Column(
                crossAxisAlignment: pw.CrossAxisAlignment.start,
                children: [
                  pw.Text('${global.language("doc_type")}: ${sd.doctype ?? "-"}', style: pw.TextStyle(font: font, fontSize: 11)),
                  pw.SizedBox(height: 2),
                  pw.Text('${global.language("sale_code")}: ${sd.salecode ?? "-"}', style: pw.TextStyle(font: font, fontSize: 11)),
                  if (isMulti) ...[
                    pw.SizedBox(height: 2),
                    pw.Text(
                      '${global.language("exchange_rate")}: ${global.formatNumber(_safeRate())}',
                      style: pw.TextStyle(font: font, fontSize: 11, color: PdfColors.blue700),
                    ),
                  ],
                ],
              ),
            ],
          ),
        ],
      ),
    );
  }

  /// สร้างข้อมูลลูกค้า/ผู้จำหน่ายแบบละเอียด — ใช้ language keys
  pw.Widget _buildPdfCustomerInfoDetailed(pw.Font font, pw.Font fontBold) {
    return pw.Container(
      padding: const pw.EdgeInsets.all(8),
      decoration: pw.BoxDecoration(
        border: pw.Border.all(color: PdfColors.grey400),
        borderRadius: const pw.BorderRadius.all(pw.Radius.circular(4)),
      ),
      child: pw.Column(
        crossAxisAlignment: pw.CrossAxisAlignment.start,
        children: [
          pw.Text(
            '${global.language(widget.fieldNameCustCode)}/${global.language(widget.fieldNameCustName)}',
            style: pw.TextStyle(fontSize: 14, font: fontBold),
          ),
          pw.Divider(),
          pw.Text(
            '${global.language(widget.fieldNameCustCode)}: ${widget.screenData.custcode}',
            style: pw.TextStyle(font: font, fontSize: 11),
          ),
          pw.SizedBox(height: 2),
          pw.Text('${global.language(widget.fieldNameCustName)}: ${global.activeLangName(widget.screenData.custnames!)}', style: pw.TextStyle(font: fontBold, fontSize: 11)),
          if (widget.screenData.customertelephone != null && widget.screenData.customertelephone!.isNotEmpty) ...[
            pw.SizedBox(height: 2),
            pw.Text('${global.language("telephone")}: ${widget.screenData.customertelephone}', style: pw.TextStyle(font: font, fontSize: 11)),
          ],
        ],
      ),
    );
  }

  /// สร้างตารางสินค้าแบบละเอียด 100% รองรับภาษาไทยและ Multi-Currency
  pw.Widget _buildPdfProductTableDetailed(pw.Font font, pw.Font fontBold) {
    final isMulti = _isMultiCurrency();
    final rate = _safeRate();
    final docSymbol = _getDocSymbol();
    final baseSymbol = _getBaseSymbol();

    // Header — แสดงสัญลักษณ์สกุลเงินเอกสารเมื่อ multi-currency
    final priceLabel = global.language('price');
    final amountLabel = global.language('amount');
    final priceHeader = isMulti ? '$priceLabel ($docSymbol)' : priceLabel;
    final amountHeader = isMulti ? '$amountLabel ($docSymbol)' : amountLabel;

    return pw.TableHelper.fromTextArray(
      context: null,
      headers: [
        '#',
        global.language('product_code'),
        global.language('barcode'),
        global.language('product_name'),
        global.language('unit'),
        global.language('qty'),
        priceHeader,
        global.language('discount'),
        amountHeader,
      ],
      data: widget.screenData.details!.asMap().entries.map((entry) {
        int index = entry.key;
        var item = entry.value;

        // เมื่อ multi-currency แสดงทั้ง 2 สกุลเงิน (doc + base) ในข้อความเดียว
        String priceText;
        String amountText;
        if (isMulti) {
          final docPrice = item.priceDoc ?? (rate > 0 ? item.price / rate : 0);
          final docAmount = item.sumAmountDoc ?? (rate > 0 ? item.sumamount / rate : 0);
          priceText = '${global.formatUnitPrice(docPrice)}\n($baseSymbol ${global.formatUnitPrice(item.price)})';
          amountText = '${global.formatNumber(docAmount)}\n($baseSymbol ${global.formatNumber(item.sumamount)})';
        } else {
          priceText = global.formatUnitPrice(item.price);
          amountText = global.formatNumber(item.sumamount);
        }

        return [
          '${index + 1}',
          item.itemcode,
          item.barcode,
          global.activeLangName(item.itemnames!),
          item.unitcode,
          global.formatQuantity(item.qty),
          priceText,
          item.discount.isEmpty ? '-' : item.discount,
          amountText,
        ];
      }).toList(),
      border: pw.TableBorder.all(color: PdfColors.grey400),
      headerStyle: pw.TextStyle(font: fontBold, fontSize: 9),
      headerDecoration: const pw.BoxDecoration(color: PdfColors.grey300),
      cellStyle: pw.TextStyle(font: font, fontSize: 8),
      cellHeight: isMulti ? 30 : 20,
      cellPadding: const pw.EdgeInsets.symmetric(horizontal: 3, vertical: 2),
      cellAlignments: {0: pw.Alignment.center, 5: pw.Alignment.centerRight, 6: pw.Alignment.centerRight, 7: pw.Alignment.centerRight, 8: pw.Alignment.centerRight},
      columnWidths: {
        0: const pw.FixedColumnWidth(25),
        1: const pw.FlexColumnWidth(1),
        2: const pw.FlexColumnWidth(1.2),
        3: const pw.FlexColumnWidth(2.5),
        4: const pw.FlexColumnWidth(0.7),
        5: const pw.FlexColumnWidth(0.7),
        6: const pw.FlexColumnWidth(0.9),
        7: const pw.FlexColumnWidth(0.7),
        8: const pw.FlexColumnWidth(1.1),
      },
    );
  }

  /// สร้างส่วนสรุปยอดแบบละเอียด รองรับ Multi-Currency + ใช้ language keys
  pw.Widget _buildPdfSummaryDetailed(pw.Font font, pw.Font fontBold) {
    final sd = widget.screenData;
    final isMulti = _isMultiCurrency();
    final rate = _safeRate();
    final docSymbol = _getDocSymbol();
    final baseSymbol = _getBaseSymbol();

    List<pw.Widget> rows = [];

    // มูลค่ารวม
    rows.add(_buildPdfSummaryRow(global.language('total_value'), sd.totalvalue, font, fontBold, isMulti: isMulti, rate: rate, docSymbol: docSymbol, baseSymbol: baseSymbol));

    // ส่วนลดรายการ
    if ((sd.detailtotaldiscount ?? 0) != 0) {
      String discLabel = global.language('detail_total_discount');
      if (sd.detaildiscountformula != null && sd.detaildiscountformula!.isNotEmpty) {
        discLabel += ' (${sd.detaildiscountformula})';
      }
      rows.add(_buildPdfSummaryRow(discLabel, sd.detailtotaldiscount!, font, fontBold, isMulti: isMulti, rate: rate, docSymbol: docSymbol, baseSymbol: baseSymbol));
    }

    // ส่วนลด VAT
    if ((sd.totaldiscountvatamount ?? 0) != 0) {
      rows.add(_buildPdfSummaryRow(global.language('total_discount_vat_amount'), sd.totaldiscountvatamount!, font, fontBold, isMulti: isMulti, rate: rate, docSymbol: docSymbol, baseSymbol: baseSymbol));
    }

    // ยกเว้นส่วนลด VAT
    if ((sd.totaldiscountexceptvatamount ?? 0) != 0) {
      rows.add(_buildPdfSummaryRow(global.language('total_discount_except_vat_amount'), sd.totaldiscountexceptvatamount!, font, fontBold, isMulti: isMulti, rate: rate, docSymbol: docSymbol, baseSymbol: baseSymbol));
    }

    // มูลค่าก่อนภาษี
    if (sd.totalbeforevat != 0) {
      rows.add(_buildPdfSummaryRow(global.language('total_before_vat'), sd.totalbeforevat, font, fontBold, isMulti: isMulti, rate: rate, docSymbol: docSymbol, baseSymbol: baseSymbol));
    }

    // ภาษีมูลค่าเพิ่ม
    if (sd.totalvatvalue != 0) {
      String vatLabel = global.language('doc_vat_amount');
      if (sd.vatrate != 0) {
        vatLabel += ' (${global.formatNumber(sd.vatrate)}%)';
      }
      rows.add(_buildPdfSummaryRow(vatLabel, sd.totalvatvalue, font, fontBold, isMulti: isMulti, rate: rate, docSymbol: docSymbol, baseSymbol: baseSymbol));
    }

    // มูลค่าหลังภาษี
    if (sd.totalaftervat != 0) {
      rows.add(_buildPdfSummaryRow(global.language('total_after_vat'), sd.totalaftervat, font, fontBold, isMulti: isMulti, rate: rate, docSymbol: docSymbol, baseSymbol: baseSymbol));
    }

    // มูลค่ายกเว้นภาษี
    if (sd.totalexceptvat != 0) {
      rows.add(_buildPdfSummaryRow(global.language('total_except_vat'), sd.totalexceptvat, font, fontBold, isMulti: isMulti, rate: rate, docSymbol: docSymbol, baseSymbol: baseSymbol));
    }

    // ส่วนลดท้ายบิล
    if (sd.totaldiscount != 0) {
      rows.add(_buildPdfSummaryRow(global.language('discount_bill'), sd.totaldiscount, font, fontBold, isMulti: isMulti, rate: rate, docSymbol: docSymbol, baseSymbol: baseSymbol));
    }

    // ยอดหลังหักส่วนลด
    if ((sd.totalamountafterdiscount ?? 0) != 0) {
      rows.add(_buildPdfSummaryRow(global.language('total_amount_after_discount'), sd.totalamountafterdiscount!, font, fontBold, isMulti: isMulti, rate: rate, docSymbol: docSymbol, baseSymbol: baseSymbol));
    }

    // ปัดเศษ
    if ((sd.roundamount ?? 0) != 0) {
      rows.add(_buildPdfSummaryRow(global.language('round_amount'), sd.roundamount!, font, fontBold, isMulti: isMulti, rate: rate, docSymbol: docSymbol, baseSymbol: baseSymbol));
    }

    // เส้นแบ่ง + ยอดรวมสุทธิ
    rows.add(pw.Divider(thickness: 1.5));
    rows.add(_buildPdfSummaryRow(global.language('sum_pay'), sd.totalamount, font, fontBold, isBold: true, isMulti: isMulti, rate: rate, docSymbol: docSymbol, baseSymbol: baseSymbol));

    // ข้อมูลอัตราแลกเปลี่ยน
    if (isMulti) {
      rows.add(pw.SizedBox(height: 4));
      rows.add(pw.Divider(color: PdfColors.grey300));
      rows.add(pw.Text(
        '${global.language("exchange_rate")}: 1 ${sd.docCurrency ?? ''} = ${global.formatNumber(rate)} ${sd.currency ?? ''}',
        style: pw.TextStyle(font: font, fontSize: 9, color: PdfColors.grey600),
      ));
    }

    return pw.Container(
      alignment: pw.Alignment.centerRight,
      child: pw.Container(
        width: isMulti ? 320 : 230,
        padding: const pw.EdgeInsets.all(8),
        decoration: pw.BoxDecoration(
          border: pw.Border.all(color: PdfColors.grey400),
          borderRadius: const pw.BorderRadius.all(pw.Radius.circular(4)),
        ),
        child: pw.Column(
          crossAxisAlignment: pw.CrossAxisAlignment.stretch,
          children: rows,
        ),
      ),
    );
  }

  /// Helper สำหรับแถวสรุปยอด — รองรับ Multi-Currency
  pw.Widget _buildPdfSummaryRow(
    String label,
    num value,
    pw.Font font,
    pw.Font fontBold, {
    bool isBold = false,
    bool isMulti = false,
    double rate = 1.0,
    String docSymbol = '',
    String baseSymbol = '',
  }) {
    final baseAmount = global.formatNumber(value.toDouble());

    if (isMulti && rate > 0) {
      final docAmount = global.formatNumber(value.toDouble() / rate);
      return pw.Padding(
        padding: const pw.EdgeInsets.symmetric(vertical: 1),
        child: pw.Row(
          mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
          crossAxisAlignment: pw.CrossAxisAlignment.start,
          children: [
            pw.Expanded(
              flex: 3,
              child: pw.Text(label, style: pw.TextStyle(font: font, fontSize: isBold ? 12 : 10)),
            ),
            pw.Expanded(
              flex: 4,
              child: pw.Column(
                crossAxisAlignment: pw.CrossAxisAlignment.end,
                children: [
                  pw.Text(
                    '$docSymbol $docAmount',
                    style: pw.TextStyle(font: isBold ? fontBold : font, fontSize: isBold ? 12 : 10, color: PdfColors.blue700),
                  ),
                  pw.Text(
                    '$baseSymbol $baseAmount',
                    style: pw.TextStyle(font: font, fontSize: 8, color: PdfColors.grey500),
                  ),
                ],
              ),
            ),
          ],
        ),
      );
    }

    return pw.Padding(
      padding: const pw.EdgeInsets.symmetric(vertical: 1),
      child: pw.Row(
        mainAxisAlignment: pw.MainAxisAlignment.spaceBetween,
        children: [
          pw.Text(label, style: pw.TextStyle(font: font, fontSize: 11)),
          pw.Text(
            baseAmount,
            style: pw.TextStyle(font: isBold ? fontBold : font, fontSize: isBold ? 13 : 11),
          ),
        ],
      ),
    );
  }

  /// สร้าง PDF Footer รองรับภาษาไทย
  pw.Widget _buildPdfFooter(pw.Context context, pw.Font font) {
    return pw.Container(
      alignment: pw.Alignment.centerRight,
      margin: pw.EdgeInsets.only(top: 10),
      child: pw.Text('${global.language("page")} ${context.pageNumber} / ${context.pagesCount}', style: pw.TextStyle(fontSize: 10, font: font)),
    );
  }

  /// จัดรูปแบบวันที่
  String _formatDate(String? dateTime) {
    if (dateTime == null || dateTime.isEmpty) return '-';
    try {
      final date = DateTime.parse(dateTime);
      return '${date.day.toString().padLeft(2, '0')}/${date.month.toString().padLeft(2, '0')}/${date.year}';
    } catch (e) {
      return dateTime;
    }
  }

  Widget _buildInfoRow(IconData icon, String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2.0),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, size: 14, color: Colors.grey[600]),
          const SizedBox(width: 4),
          Expanded(
            flex: 2,
            child: Text(label, style: const TextStyle(fontSize: 12, color: Colors.grey)),
          ),
          Expanded(
            flex: 3,
            child: Text(
              value,
              style: const TextStyle(fontSize: 12, fontWeight: FontWeight.w500, color: Colors.black87),
            ),
          ),
        ],
      ),
    );
  }

  /// ตรวจสอบว่าเป็นเอกสาร Multi-Currency หรือไม่
  bool _isMultiCurrency() {
    final sd = widget.screenData;
    return sd.docCurrency != null &&
        sd.docCurrency!.isNotEmpty &&
        sd.currency != null &&
        sd.currency!.isNotEmpty &&
        sd.docCurrency != sd.currency &&
        (sd.exchangerate ?? 1.0) != 1.0;
  }

  /// ดึง exchange rate อย่างปลอดภัย
  double _safeRate() {
    final r = widget.screenData.exchangerate ?? 1.0;
    return (r > 0 && !r.isNaN && !r.isInfinite) ? r : 1.0;
  }

  /// ดึงสัญลักษณ์สกุลเงินเอกสาร
  String _getDocSymbol() {
    return widget.screenData.docCurrencySymbol ?? widget.screenData.docCurrency ?? '';
  }

  /// ดึงสัญลักษณ์สกุลเงินหลัก
  String _getBaseSymbol() {
    return widget.screenData.currencysymbol ?? widget.screenData.currency ?? '฿';
  }

  /// สร้าง InfoRow แสดง 2 สกุลเงิน (doc=น้ำเงิน, base=เทา)
  Widget _buildDualCurrencyInfoRow(IconData icon, String label, double docValue, double baseValue) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 2.0),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(icon, size: 14, color: Colors.grey[600]),
          const SizedBox(width: 4),
          Expanded(
            flex: 2,
            child: Text(label, style: const TextStyle(fontSize: 12, color: Colors.grey)),
          ),
          Expanded(
            flex: 3,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  '${_getDocSymbol()} ${global.formatNumber(docValue)}',
                  style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: Colors.blue[700]),
                ),
                Text(
                  '${_getBaseSymbol()} ${global.formatNumber(baseValue)}',
                  style: TextStyle(fontSize: 10, color: Colors.grey[500], fontStyle: FontStyle.italic),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildInfoTile(String label, String value, IconData icon, Color color) {
    return Container(
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: color.withOpacity(0.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(icon, size: 12, color: color),
              const SizedBox(width: 4),
              Text(
                label,
                style: TextStyle(fontSize: 10, fontWeight: FontWeight.w500, color: color),
              ),
            ],
          ),
          const SizedBox(height: 2),
          Text(
            value,
            style: const TextStyle(fontSize: 11, fontWeight: FontWeight.bold, color: Colors.black87),
          ),
        ],
      ),
    );
  }

  Widget _buildSimpleProductTable() {
    return Expanded(
      child: SingleChildScrollView(
        child: Column(
          children: [
            // Product Items as Cards
            ...widget.screenData.details!.map((item) {
              int index = widget.screenData.details!.indexOf(item);
              return Card(
                margin: const EdgeInsets.symmetric(vertical: 4, horizontal: 8),
                elevation: 2,
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
                child: Container(
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(12),
                    gradient: LinearGradient(begin: Alignment.topLeft, end: Alignment.bottomRight, colors: [Colors.white, Colors.green[50]!]),
                  ),
                  child: Padding(
                    padding: const EdgeInsets.all(12.0),
                    child: Row(
                      children: [
                        // Item number badge
                        Container(
                          width: 35,
                          height: 35,
                          decoration: BoxDecoration(
                            gradient: LinearGradient(colors: [Colors.green[400]!, Colors.green[600]!]),
                            borderRadius: BorderRadius.circular(17.5),
                            boxShadow: [BoxShadow(color: Colors.green.withOpacity(0.3), spreadRadius: 1, blurRadius: 3, offset: const Offset(0, 2))],
                          ),
                          child: Center(
                            child: Text(
                              (index + 1).toString(),
                              style: const TextStyle(fontWeight: FontWeight.bold, color: Colors.white, fontSize: 14),
                            ),
                          ),
                        ),
                        const SizedBox(width: 12),
                        // Product info
                        Expanded(
                          flex: 3,
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                global.activeLangName(item.itemnames!),
                                style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: Colors.black87),
                              ),
                              if (item.description!.isNotEmpty)
                                Padding(
                                  padding: const EdgeInsets.only(top: 4.0),
                                  child: Text(
                                    'หมายเหตุ: ${item.description!}',
                                    style: TextStyle(fontSize: 11, color: Colors.grey[600], fontStyle: FontStyle.italic),
                                  ),
                                ),
                            ],
                          ),
                        ),
                        // Quantity with icon
                        Container(
                          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                          decoration: BoxDecoration(
                            color: Colors.green[100],
                            borderRadius: BorderRadius.circular(8),
                            border: Border.all(color: Colors.green[300]!),
                          ),
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              Icon(Icons.numbers, size: 14, color: Colors.green[700]),
                              const SizedBox(width: 4),
                              Text(
                                global.formatQuantity(item.qty).toString(),
                                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13, color: Colors.green[700]),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              );
            }),
          ],
        ),
      ),
    );
  }

  Widget _buildSimpleProductTableContent() {
    return Column(
      children: [
        // Product Items as Cards
        ...widget.screenData.details!.map((item) {
          int index = widget.screenData.details!.indexOf(item);
          return Card(
            margin: const EdgeInsets.symmetric(vertical: 4, horizontal: 8),
            elevation: 2,
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
            child: Container(
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(12),
                gradient: LinearGradient(begin: Alignment.topLeft, end: Alignment.bottomRight, colors: [Colors.white, Colors.green[50]!]),
              ),
              child: Padding(
                padding: const EdgeInsets.all(12.0),
                child: Row(
                  children: [
                    // Item number badge
                    Container(
                      width: 35,
                      height: 35,
                      decoration: BoxDecoration(
                        gradient: LinearGradient(colors: [Colors.green[400]!, Colors.green[600]!]),
                        borderRadius: BorderRadius.circular(17.5),
                        boxShadow: [BoxShadow(color: Colors.green.withOpacity(0.3), spreadRadius: 1, blurRadius: 3, offset: const Offset(0, 2))],
                      ),
                      child: Center(
                        child: Text(
                          (index + 1).toString(),
                          style: const TextStyle(fontWeight: FontWeight.bold, color: Colors.white, fontSize: 14),
                        ),
                      ),
                    ),
                    const SizedBox(width: 12),
                    // Product info
                    Expanded(
                      flex: 3,
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            global.activeLangName(item.itemnames!),
                            style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: Colors.black87),
                          ),
                          if (item.description!.isNotEmpty)
                            Padding(
                              padding: const EdgeInsets.only(top: 4.0),
                              child: Text(
                                'หมายเหตุ: ${item.description!}',
                                style: TextStyle(fontSize: 11, color: Colors.grey[600], fontStyle: FontStyle.italic),
                              ),
                            ),
                        ],
                      ),
                    ),
                    // Quantity with icon
                    Container(
                      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                      decoration: BoxDecoration(
                        color: Colors.green[100],
                        borderRadius: BorderRadius.circular(8),
                        border: Border.all(color: Colors.green[300]!),
                      ),
                      child: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(Icons.numbers, size: 14, color: Colors.green[700]),
                          const SizedBox(width: 4),
                          Text(
                            global.formatQuantity(item.qty).toString(),
                            style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13, color: Colors.green[700]),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
          );
        }),
      ],
    );
  }

  void _buildFinancialSummary(List<Widget> previewResult) {
    previewResult.add(DocumentTotalWidget(screenData: widget.screenData, showOnlyTotalValue: _config.showOnlyTotalValue));
  }
}
