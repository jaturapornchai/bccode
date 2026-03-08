import 'package:flutter/material.dart';
import 'package:uuid/uuid.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/modules/cart-system/models/cart_transfer_model.dart';
import 'package:smlaicloud/modules/cart-system/models/cart_system_type.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
// ignore: library_prefixes
import 'package:smlaicloud/calamount.dart' as calAmount;

/// Handler สำหรับจัดการการรับข้อมูลจากระบบตะกร้าสินค้าสำหรับใบเสนอราคา
///
/// รับผิดชอบ:
/// - ตรวจสอบ CartTransferModel จาก Provider/Route Arguments
/// - แปลงข้อมูลจาก CartTransferModel เป็น TransactionModel
/// - อัพเดทข้อมูลหัวเอกสารและรายการสินค้า
class QTCartIntegrationHandler {
  /// ตรวจสอบและดึง CartTransferModel จาก context
  static CartTransferModel? getCartTransfer(BuildContext context) {
    // 1. ตรวจสอบจาก CartTransferProvider (สำหรับ tab ใหม่)
    CartTransferModel? cartTransfer = CartTransferProvider.of(context);

    // 2. ถ้าไม่มี ลองตรวจสอบจาก route arguments (สำหรับ Navigator.pushNamed)
    if (cartTransfer == null) {
      final args = ModalRoute.of(context)?.settings.arguments;
      if (args is CartTransferModel) {
        cartTransfer = args;
      }
    }

    return cartTransfer;
  }

  /// แปลงข้อมูลจาก CartTransferModel มาเป็น TransactionModel
  static TransactionModel populateFromCart({
    required TransactionModel screenData,
    required CartTransferModel cartTransfer,
    required int calcflag,
  }) {
    // ใบเสนอราคาใช้ข้อมูล debtor (ลูกหนี้)
    String customerCode = '';
    String customerName = '';

    switch (cartTransfer.systemType) {
      case CartSystemType.sales:
        customerCode = cartTransfer.debtorCode ?? '';
        customerName = cartTransfer.debtorName ?? '';
        break;
      case CartSystemType.purchase:
        customerCode = cartTransfer.creditorCode ?? '';
        customerName = cartTransfer.creditorName ?? '';
        break;
      case CartSystemType.inventory:
        customerCode = '';
        customerName = '';
        break;
    }

    screenData.custcode = customerCode;
    screenData.custnames = customerName.isNotEmpty
        ? [LanguageDataModel(code: 'th', name: customerName)]
        : [];

    // แปลงรายการสินค้าจาก cart เป็น transaction details
    screenData.details = cartTransfer.items.map((cartItem) {
      return TransactionDetailModel(
        docdatetime: DateTime.now().toUtc().toIso8601String(),
        itemguid: const Uuid().v4(),
        barcode: cartItem.barcode,
        itemcode: cartItem.itemCode,
        itemnames: [LanguageDataModel(code: 'th', name: cartItem.productName)],
        unitcode: cartItem.unitCode,
        qty: cartItem.quantity,
        price: cartItem.price,
        discount: cartItem.discount.toString(),
        discountamount: cartItem.discountAmount,
        sumofcost: 0,
        sumamount: cartItem.totalPrice,
        remark: cartItem.remark ?? '',
        linenumber: 0,
        whcode: cartItem.warehouseId ?? '',
        whnames: cartItem.warehouseName != null
            ? [LanguageDataModel(code: 'th', name: cartItem.warehouseName!)]
            : [],
        locationcode: cartItem.locationId ?? '',
        locationnames: cartItem.locationName != null
            ? [LanguageDataModel(code: 'th', name: cartItem.locationName!)]
            : [],
        towhcode: '',
        towhnames: [],
        tolocationcode: '',
        tolocationnames: [],
        shelfcode: '',
        totalqty: cartItem.quantity,
        averagecost: 0,
        totalvaluevat: 0,
        sumamountexcludevat: 0,
        priceexcludevat: 0,
        standvalue: cartItem.unitStand.toDouble(),
        dividevalue: cartItem.unitDive.toDouble(),
        calcflag: calcflag,
        vattype: screenData.vattype,
        ispos: 0,
        laststatus: 0,
        itemtype: 0,
        taxtype: 0,
        inquirytype: screenData.inquirytype,
        multiunit: false,
        unitnames: [LanguageDataModel(code: 'th', name: cartItem.unitName)],
        refbarcodes: [],
        imageuri: cartItem.imageuri,
      );
    }).toList();

    // คำนวณยอดรวม
    screenData.totalvalue = cartTransfer.totalAmount;
    screenData.totalamount = cartTransfer.totalAmount;

    // คำนวณยอดรวมใหม่ตามประเภทภาษี
    screenData = calAmount.calTotalValue(screenData);

    AppLogger.info('[QTCartIntegration] โหลดข้อมูลจากตะกร้าสำเร็จ: ${cartTransfer.items.length} รายการ');

    return screenData;
  }
}
