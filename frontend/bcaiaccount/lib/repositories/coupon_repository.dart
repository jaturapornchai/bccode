import 'package:flutter/foundation.dart';
import 'package:dio/dio.dart';
import '../model/coupon_model.dart';
import '../model/responses_model.dart';
import '../model/coupon_import_model.dart';
import 'client.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class CouponRepository {
  Future<CouponResponseModel> getCoupons() async {
    Dio client = Client().init();
    try {
      final response = await client.get('/coupon');

      if (response.statusCode == 200) {
        return CouponResponseModel.fromJson(response.data);
      } else {
        throw Exception('Failed to load coupons: ${response.statusCode}');
      }
    } on DioException catch (ex) {
      String errorMessage =
          ex.response?.data?.toString() ?? ex.message ?? 'Unknown error';
      throw Exception('Error fetching coupons: $errorMessage');
    }
  }

  Future<CouponModel> getCouponById(String id) async {
    Dio client = Client().init();
    try {
      if (kDebugMode) {
        AppLogger.debug('🌐 API Call: GET /coupon/$id');
      }
      final response = await client.get('/coupon/$id');

      if (response.statusCode == 200) {
        if (kDebugMode) {
          AppLogger.debug('📡 API Response: ${response.data}');
        }
        // Parse เฉพาะส่วน data ที่อยู่ใน response.data['data']
        final couponData = response.data['data'];
        if (kDebugMode) {
          AppLogger.debug('📊 Coupon Data to parse: $couponData');
        }
        final coupon = CouponModel.fromJson(couponData);
        if (kDebugMode) {
          AppLogger.debug('🔄 Parsed CouponModel: ${coupon.toJson()}');
        }
        return coupon;
      } else {
        throw Exception('Failed to load coupon: ${response.statusCode}');
      }
    } on DioException catch (ex) {
      String errorMessage =
          ex.response?.data?.toString() ?? ex.message ?? 'Unknown error';
      if (kDebugMode) {
        AppLogger.error('❌ DioException in getCouponById: $errorMessage');
      }
      throw Exception('Error fetching coupon: $errorMessage');
    }
  }

  Future<ResponsesModel> createCoupon(CouponModel coupon) async {
    Dio client = Client().init();
    try {
      final response = await client.post('/coupon', data: coupon.toJson());

      if (response.statusCode == 200 || response.statusCode == 201) {
        return ResponsesModel.fromJson(response.data);
      } else {
        throw Exception('Failed to create coupon: ${response.statusCode}');
      }
    } on DioException catch (ex) {
      String errorMessage =
          ex.response?.data?.toString() ?? ex.message ?? 'Unknown error';
      throw Exception('Error creating coupon: $errorMessage');
    }
  }

  Future<ResponsesModel> updateCoupon(String id, CouponModel coupon) async {
    Dio client = Client().init();
    try {
      final response = await client.put('/coupon/$id', data: coupon.toJson());

      if (response.statusCode == 200 || response.statusCode == 201) {
        return ResponsesModel.fromJson(response.data);
      } else {
        throw Exception('Failed to update coupon: ${response.statusCode}');
      }
    } on DioException catch (ex) {
      String errorMessage =
          ex.response?.data?.toString() ?? ex.message ?? 'Unknown error';
      throw Exception('Error updating coupon: $errorMessage');
    }
  }

  Future<ResponsesModel> deleteCoupon(String id) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/coupon/$id');

      if (response.statusCode == 200) {
        return ResponsesModel.fromJson(response.data);
      } else {
        throw Exception('Failed to delete coupon: ${response.statusCode}');
      }
    } on DioException catch (ex) {
      String errorMessage =
          ex.response?.data?.toString() ?? ex.message ?? 'Unknown error';
      throw Exception('Error deleting coupon: $errorMessage');
    }
  }

  /// Upload Excel file for coupon import
  /// [file] - Excel file bytes
  /// [filename] - Name of the file
  /// [validateOnly] - true for validation mode, false for actual import
  Future<CouponImportUploadResponse> uploadExcelFile({
    required Uint8List file,
    required String filename,
    required bool validateOnly,
  }) async {
    Dio client = Client().init();
    try {
      if (kDebugMode) {
        AppLogger.debug(
          '🌐 API Call: POST /coupon/upload-excel (validateOnly: $validateOnly)',
        );
      }

      FormData formData = FormData.fromMap({
        'file': MultipartFile.fromBytes(file, filename: filename),
        'validate_only': validateOnly,
      });

      final response = await client.post(
        '/coupon/upload-excel',
        data: formData,
      );

      if (response.statusCode == 200 || response.statusCode == 201) {
        if (kDebugMode) {
          AppLogger.debug('📡 Upload Response: ${response.data}');
        }
        return CouponImportUploadResponse.fromJson(response.data);
      } else {
        throw Exception('Failed to upload file: ${response.statusCode}');
      }
    } on DioException catch (ex) {
      String errorMessage =
          ex.response?.data?.toString() ?? ex.message ?? 'Unknown error';
      if (kDebugMode) {
        AppLogger.error('❌ DioException in uploadExcelFile: $errorMessage');
      }
      throw Exception('Error uploading file: $errorMessage');
    }
  }

  /// Get import status by batch ID
  /// [batchId] - Batch ID from upload response
  Future<CouponImportStatusResponse> getImportStatus(String batchId) async {
    Dio client = Client().init();
    try {
      if (kDebugMode) {
        AppLogger.debug('🌐 API Call: GET /coupon/import-status/$batchId');
      }

      final response = await client.get('/coupon/import-status/$batchId');

      if (response.statusCode == 200) {
        if (kDebugMode) {
          AppLogger.debug('📡 Status Response: ${response.data}');
        }
        return CouponImportStatusResponse.fromJson(response.data);
      } else {
        throw Exception('Failed to get import status: ${response.statusCode}');
      }
    } on DioException catch (ex) {
      String errorMessage =
          ex.response?.data?.toString() ?? ex.message ?? 'Unknown error';
      if (kDebugMode) {
        AppLogger.error('❌ DioException in getImportStatus: $errorMessage');
      }
      throw Exception('Error getting import status: $errorMessage');
    }
  }
}
