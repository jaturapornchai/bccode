import 'package:flutter/material.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/price_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/screen_search/barcode_search_screen.dart';

/// Handler สำหรับค้นหาสินค้าเฉพาะใบเสนอราคา (Quotation)
/// Quotation ใช้ screen: '' สำหรับค้นหาสินค้าทั่วไป (ไม่ filter เฉพาะวัตถุดิบ)
class QTProductSearchHandler {
  /// Screen type สำหรับ Quotation = '' (สินค้าทั่วไป)
  static const String screenType = '';

  /// ค้นหาสินค้าด้วย Barcode
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
  static void updateDetailFromProduct({
    required TransactionDetailModel detail,
    required ProductBarcodeModel product,
    required QTWarehouseInfo warehouseInfo,
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
  static void populateDetailFromProduct({
    required TransactionDetailModel detail,
    required ProductBarcodeModel product,
    required QTWarehouseInfo warehouseInfo,
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
class QTWarehouseInfo {
  final String defaultWarehouse;
  final List<LanguageDataModel> defaultWarehouseNames;
  final String defaultLocation;
  final List<LanguageDataModel> defaultLocationNames;
  final String defaultToWarehouse;
  final List<LanguageDataModel> defaultToWarehouseNames;
  final String defaultToLocation;
  final List<LanguageDataModel> defaultToLocationNames;

  const QTWarehouseInfo({
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
