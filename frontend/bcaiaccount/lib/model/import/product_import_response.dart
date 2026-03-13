import 'package:flutter/material.dart';
import '../../global.dart' as global;

class DuplicateBarcode {
  final String barcode;
  final int count;
  final List<int> rowNumbers;

  DuplicateBarcode({
    required this.barcode,
    required this.count,
    required this.rowNumbers,
  });

  factory DuplicateBarcode.fromJson(Map<String, dynamic> json) {
    return DuplicateBarcode(
      barcode: json['barcode'] ?? '',
      count: json['count'] ?? 0,
      rowNumbers:
          (json['rowNumbers'] as List<dynamic>?)
              ?.map((e) => e as int)
              .toList() ??
          [],
    );
  }

  Map<String, dynamic> toJson() {
    return {'barcode': barcode, 'count': count, 'rowNumbers': rowNumbers};
  }
}

class ProductImportResponse {
  final ComparisonData? comparison;
  final String duration;
  final int errorCount;
  final String fileName;
  final String shopId;
  final int status;
  final bool success;
  final int successCount;
  final int totalRows;
  final String? errorMessage;
  final List<DuplicateBarcode>? duplicateBarcodes;

  ProductImportResponse({
    this.comparison,
    required this.duration,
    required this.errorCount,
    required this.fileName,
    required this.shopId,
    required this.status,
    required this.success,
    required this.successCount,
    required this.totalRows,
    this.errorMessage,
    this.duplicateBarcodes,
  });

  factory ProductImportResponse.fromJson(Map<String, dynamic> json) {
    return ProductImportResponse(
      comparison: json['comparison'] != null
          ? ComparisonData.fromJson(json['comparison'])
          : null,
      duration: json['duration'] ?? '',
      errorCount: json['errorCount'] ?? 0,
      fileName: json['fileName'] ?? '',
      shopId: json['shopId'] ?? '',
      status: json['status'] ?? 0,
      success: json['success'] ?? false,
      successCount: json['successCount'] ?? 0,
      totalRows: json['totalRows'] ?? 0,
      errorMessage: json['errorMessage'],
      duplicateBarcodes: (json['duplicateBarcodes'] as List<dynamic>?)
          ?.map((e) => DuplicateBarcode.fromJson(e))
          .toList(),
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'comparison': comparison?.toJson(),
      'duration': duration,
      'errorCount': errorCount,
      'fileName': fileName,
      'shopId': shopId,
      'status': status,
      'success': success,
      'successCount': successCount,
      'totalRows': totalRows,
      'errorMessage': errorMessage,
      'duplicateBarcodes': duplicateBarcodes?.map((e) => e.toJson()).toList(),
    };
  }
}

class ComparisonData {
  final String shopId;
  final int totalExcelRows;
  final int totalMongoProducts;
  final int newProductCount;
  final int updatedProductCount;
  final int unchangedCount;
  final String processTime;
  final List<ProductComparison> products;

  ComparisonData({
    required this.shopId,
    required this.totalExcelRows,
    required this.totalMongoProducts,
    required this.newProductCount,
    required this.updatedProductCount,
    required this.unchangedCount,
    required this.processTime,
    required this.products,
  });

  factory ComparisonData.fromJson(Map<String, dynamic> json) {
    return ComparisonData(
      shopId: json['shopId'] ?? '',
      totalExcelRows: json['totalExcelRows'] ?? 0,
      totalMongoProducts: json['totalMongoProducts'] ?? 0,
      newProductCount: json['newProductCount'] ?? 0,
      updatedProductCount: json['updatedProductCount'] ?? 0,
      unchangedCount: json['unchangedCount'] ?? 0,
      processTime: json['processTime'] ?? '',
      products:
          (json['products'] as List<dynamic>?)
              ?.map((e) => ProductComparison.fromJson(e))
              .toList() ??
          [],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'shopId': shopId,
      'totalExcelRows': totalExcelRows,
      'totalMongoProducts': totalMongoProducts,
      'newProductCount': newProductCount,
      'updatedProductCount': updatedProductCount,
      'unchangedCount': unchangedCount,
      'processTime': processTime,
      'products': products.map((e) => e.toJson()).toList(),
    };
  }
}

class ProductComparison {
  final String barcode;
  final ProductData? mongo;
  final ProductData? excel;
  final int action; // 1=new, 2=updated, 0=unchanged

  ProductComparison({
    required this.barcode,
    this.mongo,
    this.excel,
    required this.action,
  });

  factory ProductComparison.fromJson(Map<String, dynamic> json) {
    return ProductComparison(
      barcode: json['barcode'] ?? '',
      mongo: json['mongo'] != null ? ProductData.fromJson(json['mongo']) : null,
      excel: json['excel'] != null ? ProductData.fromJson(json['excel']) : null,
      action: json['action'] ?? 0,
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'barcode': barcode,
      'mongo': mongo?.toJson(),
      'excel': excel?.toJson(),
      'action': action,
    };
  }

  String get actionText {
    switch (action) {
      case 1:
        return global.language('import.action.created');
      case 2:
        return global.language('import.action.updated');
      default:
        return global.language('import.action.unchanged');
    }
  }

  Color get actionColor {
    switch (action) {
      case 1:
        return Colors.green;
      case 2:
        return Colors.orange;
      default:
        return global.theme.iconSecondaryColor;
    }
  }
}

class ProductData {
  final String code;
  final String name;
  final int divideValue;
  final int standValue;
  final String? unitCode;

  ProductData({
    required this.code,
    required this.name,
    required this.divideValue,
    required this.standValue,
    this.unitCode,
  });

  factory ProductData.fromJson(Map<String, dynamic> json) {
    return ProductData(
      code: json['code'] ?? '',
      name: json['name'] ?? '',
      divideValue: json['divideValue'] ?? 0,
      standValue: json['standValue'] ?? 0,
      unitCode: json['unitCode'],
    );
  }

  Map<String, dynamic> toJson() {
    return {
      'code': code,
      'name': name,
      'divideValue': divideValue,
      'standValue': standValue,
      'unitCode': unitCode,
    };
  }
}
