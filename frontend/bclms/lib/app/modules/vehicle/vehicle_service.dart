import 'package:dio/dio.dart';
import 'package:bclms/app/core/constants/api_constants.dart';
import 'package:bclms/app/core/utils/app_logger.dart';
import 'package:bclms/app/modules/vehicle/vehicle_model.dart';
import 'package:bclms/app/services/api_service.dart';

/// Service สำหรับ CRUD ยานพาหนะ — ตาม Backend API /logistics/vehicle
class VehicleService {
  final ApiService _api;

  VehicleService(this._api);

  Dio get _dio => _api.dio;

  /// ดึงรายการยานพาหนะ (Pagination)
  /// GET /logistics/vehicle?q=&vehicletype=&status=&page=1&limit=20
  Future<List<VehicleModel>> list({
    String? search,
    String? vehicleType,
    int? status,
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
      if (vehicleType != null && vehicleType.isNotEmpty) {
        queryParams['vehicletype'] = vehicleType;
      }
      if (status != null) {
        queryParams['status'] = status;
      }

      final response = await _dio.get(
        ApiConstants.vehicle,
        queryParameters: queryParams,
      );

      if (response.data['success'] == true) {
        final list = response.data['data'] as List? ?? [];
        AppLogger.info(
            '[Vehicle] โหลดรายการยานพาหนะสำเร็จ: ${list.length} คัน');
        return list.map((e) => VehicleModel.fromJson(e)).toList();
      }
      return [];
    } catch (e) {
      AppLogger.error('[Vehicle] โหลดรายการยานพาหนะล้มเหลว', error: e);
      return [];
    }
  }

  /// ดึงข้อมูลยานพาหนะตาม ID
  /// GET /logistics/vehicle/:id
  Future<VehicleModel?> getById(String guidFixed) async {
    try {
      final response = await _dio.get('${ApiConstants.vehicle}/$guidFixed');

      if (response.data['success'] == true) {
        AppLogger.info('[Vehicle] โหลดข้อมูลยานพาหนะสำเร็จ: $guidFixed');
        return VehicleModel.fromJson(response.data['data']);
      }
      return null;
    } catch (e) {
      AppLogger.error('[Vehicle] โหลดข้อมูลยานพาหนะล้มเหลว: $guidFixed',
          error: e);
      return null;
    }
  }

  /// สร้างยานพาหนะใหม่
  /// POST /logistics/vehicle
  Future<bool> create(VehicleModel vehicle) async {
    try {
      final response = await _dio.post(
        ApiConstants.vehicle,
        data: vehicle.toJson(),
      );

      if (response.data['success'] == true) {
        AppLogger.info('[Vehicle] สร้างยานพาหนะสำเร็จ: ${vehicle.code}');
        return true;
      }
      AppLogger.warning(
          '[Vehicle] สร้างยานพาหนะไม่สำเร็จ: ${response.data['message']}');
      return false;
    } catch (e) {
      AppLogger.error('[Vehicle] สร้างยานพาหนะล้มเหลว', error: e);
      return false;
    }
  }

  /// อัพเดทยานพาหนะ
  /// PUT /logistics/vehicle/:id
  Future<bool> update(String guidFixed, VehicleModel vehicle) async {
    try {
      final response = await _dio.put(
        '${ApiConstants.vehicle}/$guidFixed',
        data: vehicle.toJson(),
      );

      if (response.data['success'] == true) {
        AppLogger.info('[Vehicle] อัพเดทยานพาหนะสำเร็จ: $guidFixed');
        return true;
      }
      AppLogger.warning(
          '[Vehicle] อัพเดทยานพาหนะไม่สำเร็จ: ${response.data['message']}');
      return false;
    } catch (e) {
      AppLogger.error('[Vehicle] อัพเดทยานพาหนะล้มเหลว: $guidFixed',
          error: e);
      return false;
    }
  }

  /// ลบยานพาหนะ
  /// DELETE /logistics/vehicle/:id
  Future<bool> delete(String guidFixed) async {
    try {
      final response =
          await _dio.delete('${ApiConstants.vehicle}/$guidFixed');

      if (response.data['success'] == true) {
        AppLogger.info('[Vehicle] ลบยานพาหนะสำเร็จ: $guidFixed');
        return true;
      }
      return false;
    } catch (e) {
      AppLogger.error('[Vehicle] ลบยานพาหนะล้มเหลว: $guidFixed', error: e);
      return false;
    }
  }

  /// ลบยานพาหนะหลายรายการ
  /// DELETE /logistics/vehicle — body: ["guid1", "guid2"]
  Future<bool> deleteMultiple(List<String> guidFixeds) async {
    try {
      final response = await _dio.delete(
        ApiConstants.vehicle,
        data: guidFixeds,
      );

      if (response.data['success'] == true) {
        AppLogger.info(
            '[Vehicle] ลบยานพาหนะหลายรายการสำเร็จ: ${guidFixeds.length} คัน');
        return true;
      }
      return false;
    } catch (e) {
      AppLogger.error('[Vehicle] ลบยานพาหนะหลายรายการล้มเหลว', error: e);
      return false;
    }
  }
}
