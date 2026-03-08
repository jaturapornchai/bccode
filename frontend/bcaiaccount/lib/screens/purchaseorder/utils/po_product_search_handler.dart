import 'package:flutter/material.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/price_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/screen_search/barcode_search_screen.dart';

/// Handler สำหรับค้นหาสินค้าเฉพาะ PO (Purchase Order)
/// PO ใช้ screen: 'material' สำหรับค้นหาวัตถุดิบ
class POProductSearchHandler {
  /// Screen type สำหรับ PO = material (วัตถุดิบ)
  static const String screenType = 'material';

  /// ค้นหาสินค้าด้วย Barcode
  /// คืนค่า ProductBarcodeModel หรือ null ถ้ายกเลิก
  static Future<ProductBarcodeModel?> searchByBarcode(
    BuildContext context, {
    String initialWord = '',
  }) async {
    final result = await Navigator.push<ProductBarcodeModel?>(
      context,
      MaterialPageRoute(
        builder: (context) => BarcodeSearchScreen(
          word: initialWord,
          screen: screenType,
        ),
      ),
    );

    if (result != null && (result.barcode?.trim().isNotEmpty ?? false)) {
      return result;
    }
    return null;
  }

  /// ค้นหาสินค้าด้วย Item Code
  /// คืนค่า ProductBarcodeModel หรือ null ถ้ายกเลิก
  static Future<ProductBarcodeModel?> searchByItemCode(
    BuildContext context, {
    String initialWord = '',
  }) async {
    final result = await Navigator.push<ProductBarcodeModel?>(
      context,
      MaterialPageRoute(
        builder: (context) => BarcodeSearchScreen(
          word: initialWord,
          screen: screenType,
        ),
      ),
    );

    if (result != null && (result.barcode?.trim().isNotEmpty ?? false)) {
      return result;
    }
    return null;
  }

  /// อัพเดท TransactionDetailModel จาก ProductBarcodeModel ที่ค้นหาได้
  /// [detail] - รายการสินค้าที่ต้องการอัพเดท
  /// [product] - ข้อมูลสินค้าที่ค้นหาได้
  /// [warehouseInfo] - ข้อมูลคลังสินค้าเริ่มต้น
  /// [getPrice] - function สำหรับดึงราคาจาก prices array
  /// [exchangeRate] - อัตราแลกเปลี่ยน (ถ้า multi-currency จะคำนวณ priceDoc = price / rate)
  static void updateDetailFromProduct({
    required TransactionDetailModel detail,
    required ProductBarcodeModel product,
    required POWarehouseInfo warehouseInfo,
    required double Function(List<PriceDataModel>?) getPrice,
    double? exchangeRate,
  }) {
    detail.itemguid = product.guidfixed;
    detail.barcode = product.barcode ?? '';
    detail.itemcode = product.itemcode ?? '';
    detail.itemnames = product.names;
    detail.multiunit = true;
    detail.unitcode = product.itemunitcode;
    detail.dividevalue = product.dividevalue ?? 1;
    detail.standvalue = product.standvalue ?? 1;
    detail.unitnames = product.itemunitnames;

    // Warehouse info
    detail.whcode = warehouseInfo.defaultWarehouse;
    detail.whnames = warehouseInfo.defaultWarehouseNames;
    detail.locationcode = warehouseInfo.defaultLocation;
    detail.locationnames = warehouseInfo.defaultLocationNames;
    detail.towhcode = warehouseInfo.defaultToWarehouse;
    detail.towhnames = warehouseInfo.defaultToWarehouseNames;
    detail.tolocationcode = warehouseInfo.defaultToLocation;
    detail.tolocationnames = warehouseInfo.defaultToLocationNames;

    // Tax & Price
    detail.taxtype = product.taxtype ?? 1;
    detail.qty = 1;
    detail.vatcal = product.vatcal;
    detail.price = getPrice(product.prices);
    detail.manufacturerguid = product.manufacturerguid;

    // Multi-currency: คำนวณ priceDoc จาก base price
    if (exchangeRate != null && exchangeRate != 1.0 && exchangeRate > 0) {
      detail.priceDoc = detail.price / exchangeRate;
    } else {
      detail.priceDoc = null;
    }
  }

  /// อัพเดท TransactionDetailModel ที่มีอยู่แล้วจาก ProductBarcodeModel
  /// ใช้สำหรับเพิ่มข้อมูลสินค้าให้กับ detail ที่สร้างไว้แล้ว
  ///
  /// ตัวอย่างการใช้:
  /// ```dart
  /// final newDetail = TransactionDetailModel.empty(lineNumber);
  /// POProductSearchHandler.populateDetailFromProduct(
  ///   detail: newDetail,
  ///   product: searchResult,
  ///   warehouseInfo: warehouseInfo,
  ///   getPrice: getPrice,
  /// );
  /// ```
  static void populateDetailFromProduct({
    required TransactionDetailModel detail,
    required ProductBarcodeModel product,
    required POWarehouseInfo warehouseInfo,
    required double Function(List<PriceDataModel>?) getPrice,
    double? exchangeRate,
  }) {
    updateDetailFromProduct(
      detail: detail,
      product: product,
      warehouseInfo: warehouseInfo,
      getPrice: getPrice,
      exchangeRate: exchangeRate,
    );
  }
}

/// Data class สำหรับข้อมูลคลังสินค้า
class POWarehouseInfo {
  final String defaultWarehouse;
  final List<LanguageDataModel> defaultWarehouseNames;
  final String defaultLocation;
  final List<LanguageDataModel> defaultLocationNames;
  final String defaultToWarehouse;
  final List<LanguageDataModel> defaultToWarehouseNames;
  final String defaultToLocation;
  final List<LanguageDataModel> defaultToLocationNames;

  const POWarehouseInfo({
    required this.defaultWarehouse,
    required this.defaultWarehouseNames,
    required this.defaultLocation,
    required this.defaultLocationNames,
    required this.defaultToWarehouse,
    required this.defaultToWarehouseNames,
    required this.defaultToLocation,
    required this.defaultToLocationNames,
  });
}
