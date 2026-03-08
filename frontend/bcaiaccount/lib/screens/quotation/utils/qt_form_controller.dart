import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// Controller สำหรับจัดการ TextEditingController ทั้งหมดในหน้าจอใบเสนอราคา
///
/// รับผิดชอบ:
/// - สร้างและจัดการ TextEditingController ทั้งหมด
/// - โหลดข้อมูลจาก TransactionModel ไปยัง Controllers
/// - Dispose controllers เมื่อไม่ใช้งาน
class QTFormController {
  // Document Header Controllers
  final docNumberController = TextEditingController();
  final docDateController = TextEditingController();
  final docTimeController = TextEditingController();
  final docRefNumberController = TextEditingController();
  final docRefDateController = TextEditingController();
  final taxDocNoController = TextEditingController();
  final taxDocDateController = TextEditingController();
  final docRefTypeController = TextEditingController();
  final docTypeController = TextEditingController();
  final vatTypeController = TextEditingController();

  // Customer Controllers
  final custCodeController = TextEditingController();
  final custnamesController = TextEditingController();
  final saleCodeController = TextEditingController();
  final saleNameController = TextEditingController();

  // Discount & Amount Controllers
  final discountWordController = TextEditingController();
  final docDiscountWordController = TextEditingController();
  final totalCostController = TextEditingController();
  final totalValueController = TextEditingController();
  final totalDiscountController = TextEditingController();
  final totalVatValueController = TextEditingController();
  final totalAfterVatController = TextEditingController();
  final totalExceptVatController = TextEditingController();
  final totalAmountController = TextEditingController();

  // VAT & Status Controllers
  final vatRateController = TextEditingController();
  final statusController = TextEditingController();
  final descriptionController = TextEditingController();

  // Detail Discount Controllers
  final detailTotalDiscount = TextEditingController();
  final totalDiscountVatAmount = TextEditingController();
  final totalDiscountExceptVatAmount = TextEditingController();
  final roundAmount = TextEditingController();
  final totalAmountAfterDiscount = TextEditingController();
  final detailDiscountFormula = TextEditingController();
  final detailTotalAmount = TextEditingController();

  // Additional Controllers
  final roundAmountController = TextEditingController();
  final detaildiscountformulaController = TextEditingController();
  final transportAmountController = TextEditingController();

  /// โหลดข้อมูลจาก TransactionModel ไปยัง Controllers
  void loadDataFromModel(TransactionModel screenData) {
    DateTime docDateTimeFormat = _safeParseDatetime(screenData.docdatetime);
    DateTime docRefDateTimeFormat = _safeParseDatetime(screenData.docrefdate);
    DateTime taxDocDateTimeFormat = _safeParseDatetime(screenData.taxdocdate);

    docNumberController.text = screenData.docno;
    docDateController.text = DateFormat('dd/MM/yyyy').format(docDateTimeFormat.toLocal());
    docTimeController.text = DateFormat('HH:mm').format(docDateTimeFormat.toLocal());
    docRefNumberController.text = screenData.docrefno;
    docRefDateController.text = DateFormat('dd/MM/yyyy').format(docRefDateTimeFormat.toLocal());
    taxDocDateController.text = DateFormat('dd/MM/yyyy').format(taxDocDateTimeFormat.toLocal());
    taxDocNoController.text = screenData.taxdocno;
    docRefTypeController.text = screenData.docreftype.toString();
    docTypeController.text = screenData.doctype.toString();
    vatTypeController.text = screenData.vattype.toString();

    custCodeController.text = screenData.custcode;
    custnamesController.text = global.activeLangName(screenData.custnames ?? []);
    saleCodeController.text = screenData.salecode;
    saleNameController.text = screenData.salename;
    docDiscountWordController.text = screenData.discountword;
    totalCostController.text = screenData.totalcost.toString();
    totalValueController.text = screenData.totalvalue.toString();
    totalDiscountController.text = screenData.totaldiscount.toString();
    totalVatValueController.text = screenData.totalvatvalue.toString();
    totalAfterVatController.text = screenData.totalaftervat.toString();
    totalExceptVatController.text = screenData.totalexceptvat.toString();
    totalAmountController.text = screenData.totalamount.toString();

    totalDiscountVatAmount.text = screenData.totaldiscountvatamount.toString();
    totalDiscountExceptVatAmount.text = screenData.totaldiscountexceptvatamount.toString();
    totalAmountAfterDiscount.text = screenData.totalamountafterdiscount.toString();
    detailDiscountFormula.text = screenData.detaildiscountformula.toString();
    detailTotalAmount.text = screenData.detailtotalamount.toString();

    vatRateController.text = screenData.vatrate.toString();
    statusController.text = screenData.status.toString();
    descriptionController.text = screenData.description.toString();

    transportAmountController.text = NumberFormat('#,##0.00', 'en_US').format(screenData.transportamount);

    // จัดการรูปแบบวันที่ตาม yeartype
    if (global.profileData.yeartype == "buddhist") {
      docDateController.text = global.dateTimeBuddhist(docDateTimeFormat, format: global.DateTimeFormatEnum.dateDay);
      docRefDateController.text = global.dateTimeBuddhist(docRefDateTimeFormat, format: global.DateTimeFormatEnum.dateDay);
      taxDocDateController.text = global.dateTimeBuddhist(taxDocDateTimeFormat, format: global.DateTimeFormatEnum.dateDay);
    } else {
      docDateController.text = DateFormat('dd/MM/yyyy').format(docDateTimeFormat);
      docRefDateController.text = DateFormat('dd/MM/yyyy').format(docRefDateTimeFormat);
      taxDocDateController.text = DateFormat('dd/MM/yyyy').format(taxDocDateTimeFormat);
    }

    roundAmountController.text = (screenData.roundamount != 0) ? (screenData.roundamount!).toString() : "0";
    detaildiscountformulaController.text = (screenData.detaildiscountformula!.isNotEmpty) ? screenData.detaildiscountformula.toString() : "0";
    discountWordController.text = (screenData.discountword.isNotEmpty) ? screenData.discountword.toString() : "0";
  }

  /// Parse datetime อย่างปลอดภัย
  DateTime _safeParseDatetime(String? dateString) {
    if (dateString == null || dateString.isEmpty) {
      return DateTime.now();
    }
    try {
      return DateTime.parse(dateString);
    } catch (e) {
      AppLogger.warning('[QTFormController] แปลงวันที่ไม่สำเร็จ: "$dateString" → ใช้ DateTime.now() แทน');
      return DateTime.now();
    }
  }

  /// Clear ค่าทุก Controllers
  void clearAll() {
    docNumberController.clear();
    docDateController.clear();
    docTimeController.clear();
    docRefNumberController.clear();
    docRefDateController.clear();
    taxDocNoController.clear();
    taxDocDateController.clear();
    docRefTypeController.clear();
    docTypeController.clear();
    vatTypeController.clear();
    custCodeController.clear();
    custnamesController.clear();
    saleCodeController.clear();
    saleNameController.clear();
    discountWordController.clear();
    docDiscountWordController.clear();
    totalCostController.clear();
    totalValueController.clear();
    totalDiscountController.clear();
    totalVatValueController.clear();
    totalAfterVatController.clear();
    totalExceptVatController.clear();
    totalAmountController.clear();
    vatRateController.clear();
    statusController.clear();
    descriptionController.clear();
    detailTotalDiscount.clear();
    totalDiscountVatAmount.clear();
    totalDiscountExceptVatAmount.clear();
    roundAmount.clear();
    totalAmountAfterDiscount.clear();
    detailDiscountFormula.clear();
    detailTotalAmount.clear();
    roundAmountController.clear();
    detaildiscountformulaController.clear();
    transportAmountController.clear();
  }

  /// Dispose ทุก Controllers
  void dispose() {
    docNumberController.dispose();
    docDateController.dispose();
    docTimeController.dispose();
    docRefNumberController.dispose();
    docRefDateController.dispose();
    taxDocNoController.dispose();
    taxDocDateController.dispose();
    docRefTypeController.dispose();
    docTypeController.dispose();
    vatTypeController.dispose();
    custCodeController.dispose();
    custnamesController.dispose();
    saleCodeController.dispose();
    saleNameController.dispose();
    discountWordController.dispose();
    docDiscountWordController.dispose();
    totalCostController.dispose();
    totalValueController.dispose();
    totalDiscountController.dispose();
    totalVatValueController.dispose();
    totalAfterVatController.dispose();
    totalExceptVatController.dispose();
    totalAmountController.dispose();
    vatRateController.dispose();
    statusController.dispose();
    descriptionController.dispose();
    detailTotalDiscount.dispose();
    totalDiscountVatAmount.dispose();
    totalDiscountExceptVatAmount.dispose();
    roundAmount.dispose();
    totalAmountAfterDiscount.dispose();
    detailDiscountFormula.dispose();
    detailTotalAmount.dispose();
    roundAmountController.dispose();
    detaildiscountformulaController.dispose();
    transportAmountController.dispose();
  }
}
