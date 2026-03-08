import 'package:dio/dio.dart';
import 'package:bclms/app/core/constants/api_constants.dart';
import 'package:bclms/app/core/utils/app_logger.dart';
import 'package:bclms/app/modules/warehouse/warehouse_model.dart';
import 'package:bclms/app/services/api_service.dart';

/// Service สำหรับ CRUD คลังสินค้า — ตาม Backend API /warehouse
class WarehouseService {
  final ApiService _api;

  WarehouseService(this._api);

  Dio get _dio => _api.dio;

  // ============================================================
  // Warehouse CRUD
  // ============================================================

  /// ดึงรายการคลังสินค้า (Pagination)
  /// GET /warehouse?q=&page=1&limit=50
  Future<List<WarehouseModel>> list({
    String? search,
    int page = 1,
    int limit = 50,
  }) async {
    try {
      final queryParams = <String, dynamic>{
        'page': page,
        'limit': limit,
      };
      if (search != null && search.isNotEmpty) {
        queryParams['q'] = search;
      }

      final response = await _dio.get(
        ApiConstants.warehouse,
        queryParameters: queryParams,
      );

      if (response.data['success'] == true) {
        final list = response.data['data'] as List? ?? [];
        AppLogger.info(
            '[Warehouse] โหลดรายการคลังสินค้าสำเร็จ: ${list.length} รายการ');
        return list.map((e) => WarehouseModel.fromJson(e)).toList();
      }
      return [];
    } catch (e) {
      AppLogger.error('[Warehouse] โหลดรายการคลังสินค้าล้มเหลว', error: e);
      return [];
    }
  }

  /// ดึงข้อมูลคลังสินค้าตาม ID
  /// GET /warehouse/:id
  Future<WarehouseModel?> getById(String guidFixed) async {
    try {
      final response =
          await _dio.get('${ApiConstants.warehouse}/$guidFixed');

      if (response.data['success'] == true) {
        AppLogger.info('[Warehouse] โหลดข้อมูลคลังสำเร็จ: $guidFixed');
        return WarehouseModel.fromJson(response.data['data']);
      }
      return null;
    } catch (e) {
      AppLogger.error('[Warehouse] โหลดข้อมูลคลังล้มเหลว: $guidFixed',
          error: e);
      return null;
    }
  }

  /// สร้างคลังสินค้าใหม่
  /// POST /warehouse
  Future<bool> create(WarehouseModel warehouse) async {
    try {
      final response = await _dio.post(
        ApiConstants.warehouse,
        data: warehouse.toJson(),
      );

      if (response.data['success'] == true) {
        AppLogger.info('[Warehouse] สร้างคลังสินค้าสำเร็จ: ${warehouse.code}');
        return true;
      }
      AppLogger.warning(
          '[Warehouse] สร้างคลังสินค้าไม่สำเร็จ: ${response.data['message']}');
      return false;
    } catch (e) {
      AppLogger.error('[Warehouse] สร้างคลังสินค้าล้มเหลว', error: e);
      return false;
    }
  }

  /// อัพเดทคลังสินค้า
  /// PUT /warehouse/:id
  Future<bool> update(String guidFixed, WarehouseModel warehouse) async {
    try {
      final response = await _dio.put(
        '${ApiConstants.warehouse}/$guidFixed',
        data: warehouse.toJson(),
      );

      if (response.data['success'] == true) {
        AppLogger.info('[Warehouse] อัพเดทคลังสินค้าสำเร็จ: $guidFixed');
        return true;
      }
      AppLogger.warning(
          '[Warehouse] อัพเดทคลังสินค้าไม่สำเร็จ: ${response.data['message']}');
      return false;
    } catch (e) {
      AppLogger.error('[Warehouse] อัพเดทคลังสินค้าล้มเหลว: $guidFixed',
          error: e);
      return false;
    }
  }

  /// ลบคลังสินค้า
  /// DELETE /warehouse/:id
  Future<bool> delete(String guidFixed) async {
    try {
      final response =
          await _dio.delete('${ApiConstants.warehouse}/$guidFixed');

      if (response.data['success'] == true) {
        AppLogger.info('[Warehouse] ลบคลังสินค้าสำเร็จ: $guidFixed');
        return true;
      }
      return false;
    } catch (e) {
      AppLogger.error('[Warehouse] ลบคลังสินค้าล้มเหลว: $guidFixed',
          error: e);
      return false;
    }
  }

  /// ลบคลังสินค้าหลายรายการ
  /// DELETE /warehouse — body: ["guid1", "guid2"]
  Future<bool> deleteMultiple(List<String> guidFixeds) async {
    try {
      final response = await _dio.delete(
        ApiConstants.warehouse,
        data: guidFixeds,
      );

      if (response.data['success'] == true) {
        AppLogger.info(
            '[Warehouse] ลบคลังหลายรายการสำเร็จ: ${guidFixeds.length} รายการ');
        return true;
      }
      return false;
    } catch (e) {
      AppLogger.error('[Warehouse] ลบคลังหลายรายการล้มเหลว', error: e);
      return false;
    }
  }

  // ============================================================
  // Location CRUD (nested under Warehouse)
  // ============================================================

  /// สร้าง Location ใหม่ในคลัง
  /// POST /warehouse/:warehouseCode/location
  Future<bool> createLocation(String warehouseCode, LocationModel location) async {
    try {
      final response = await _dio.post(
        '${ApiConstants.warehouse}/$warehouseCode/location',
        data: {
          'warehousecode': warehouseCode,
          'locationcode': location.code,
          'locationnames': location.names.map((n) => n.toJson()).toList(),
          if (location.shelves.isNotEmpty)
            'shelf': location.shelves.map((s) => s.toJson()).toList(),
        },
      );

      if (response.data['success'] == true) {
        AppLogger.info(
            '[Warehouse] สร้าง Location สำเร็จ: $warehouseCode/${location.code}');
        return true;
      }
      return false;
    } catch (e) {
      AppLogger.error('[Warehouse] สร้าง Location ล้มเหลว', error: e);
      return false;
    }
  }

  /// อัพเดท Location
  /// PUT /warehouse/:warehouseCode/location/:locationCode
  Future<bool> updateLocation(
      String warehouseCode, String locationCode, LocationModel location) async {
    try {
      final response = await _dio.put(
        '${ApiConstants.warehouse}/$warehouseCode/location/$locationCode',
        data: {
          'warehousecode': warehouseCode,
          'locationcode': location.code,
          'locationnames': location.names.map((n) => n.toJson()).toList(),
          if (location.shelves.isNotEmpty)
            'shelf': location.shelves.map((s) => s.toJson()).toList(),
        },
      );

      if (response.data['success'] == true) {
        AppLogger.info(
            '[Warehouse] อัพเดท Location สำเร็จ: $warehouseCode/$locationCode');
        return true;
      }
      return false;
    } catch (e) {
      AppLogger.error('[Warehouse] อัพเดท Location ล้มเหลว', error: e);
      return false;
    }
  }

  /// ลบ Location (ส่ง array ของ codes)
  /// DELETE /warehouse/:warehouseCode/location — body: ["locCode1", "locCode2"]
  Future<bool> deleteLocations(
      String warehouseCode, List<String> locationCodes) async {
    try {
      final response = await _dio.delete(
        '${ApiConstants.warehouse}/$warehouseCode/location',
        data: locationCodes,
      );

      if (response.data['success'] == true) {
        AppLogger.info(
            '[Warehouse] ลบ Location สำเร็จ: $warehouseCode (${locationCodes.length} รายการ)');
        return true;
      }
      return false;
    } catch (e) {
      AppLogger.error('[Warehouse] ลบ Location ล้มเหลว', error: e);
      return false;
    }
  }

  // ============================================================
  // Shelf CRUD (nested under Location)
  // ============================================================

  /// สร้าง Shelf ใหม่
  /// POST /warehouse/:warehouseCode/location/:locationCode/shelf
  Future<bool> createShelf(
      String warehouseCode, String locationCode, ShelfModel shelf) async {
    try {
      final response = await _dio.post(
        '${ApiConstants.warehouse}/$warehouseCode/location/$locationCode/shelf',
        data: {
          'warehousecode': warehouseCode,
          'locationcode': locationCode,
          'shelfcode': shelf.code,
          'shelfname': shelf.name,
        },
      );

      if (response.data['success'] == true) {
        AppLogger.info(
            '[Warehouse] สร้าง Shelf สำเร็จ: $warehouseCode/$locationCode/${shelf.code}');
        return true;
      }
      return false;
    } catch (e) {
      AppLogger.error('[Warehouse] สร้าง Shelf ล้มเหลว', error: e);
      return false;
    }
  }

  /// อัพเดท Shelf
  /// PUT /warehouse/:warehouseCode/location/:locationCode/shelf/:shelfCode
  Future<bool> updateShelf(String warehouseCode, String locationCode,
      String shelfCode, ShelfModel shelf) async {
    try {
      final response = await _dio.put(
        '${ApiConstants.warehouse}/$warehouseCode/location/$locationCode/shelf/$shelfCode',
        data: {
          'warehousecode': warehouseCode,
          'locationcode': locationCode,
          'shelfcode': shelf.code,
          'shelfname': shelf.name,
        },
      );

      if (response.data['success'] == true) {
        AppLogger.info(
            '[Warehouse] อัพเดท Shelf สำเร็จ: $warehouseCode/$locationCode/$shelfCode');
        return true;
      }
      return false;
    } catch (e) {
      AppLogger.error('[Warehouse] อัพเดท Shelf ล้มเหลว', error: e);
      return false;
    }
  }

  /// ลบ Shelf (ส่ง array ของ codes)
  /// DELETE /warehouse/:warehouseCode/location/:locationCode/shelf — body: ["shelfCode1"]
  Future<bool> deleteShelves(
      String warehouseCode, String locationCode, List<String> shelfCodes) async {
    try {
      final response = await _dio.delete(
        '${ApiConstants.warehouse}/$warehouseCode/location/$locationCode/shelf',
        data: shelfCodes,
      );

      if (response.data['success'] == true) {
        AppLogger.info(
            '[Warehouse] ลบ Shelf สำเร็จ: $warehouseCode/$locationCode (${shelfCodes.length} รายการ)');
        return true;
      }
      return false;
    } catch (e) {
      AppLogger.error('[Warehouse] ลบ Shelf ล้มเหลว', error: e);
      return false;
    }
  }
}
