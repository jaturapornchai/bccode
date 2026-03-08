import 'package:smlaicloud/model/profile_model.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class ProfileRepository {
  /// get profile
  Future<ApiResponse> getProfile() async {
    Dio client = Client().init();
    try {
      final response = await client.get('/profile');
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      throw Exception(ex);
    }
  }

  /// update profile
  Future<ApiResponse> updateProfile(ProfileModel profileModel) async {
    Dio client = Client().init();
    final data = profileModel.toJson();
    try {
      final response = await client.put('/profile', data: data);
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      throw Exception(ex);
    }
  }

  /// เชื่อมต่อ LINE กับบัญชีผู้ใช้ (ระดับ user — ใช้ร่วมทุก shop)
  Future<ApiResponse> linkLine({
    required String lineUserId,
    required String lineDisplayName,
    required String linePictureUrl,
  }) async {
    Dio client = Client().init();
    final data = {
      'line_user_id': lineUserId,
      'line_display_name': lineDisplayName,
      'line_picture_url': linePictureUrl,
    };

    try {
      final response = await client.put('/profile/link-line', data: data);
      return ApiResponse.fromMap(response.data);
    } on DioException catch (ex) {
      String errorMessage = ex.response?.data?['message'] ?? ex.response.toString();
      AppLogger.error('linkLine error: $errorMessage');
      throw Exception(errorMessage);
    }
  }

  /// ยกเลิกการเชื่อมต่อ LINE จากบัญชีผู้ใช้
  Future<ApiResponse> unlinkLine() async {
    Dio client = Client().init();

    try {
      final response = await client.delete('/profile/link-line');
      return ApiResponse.fromMap(response.data);
    } on DioException catch (ex) {
      String errorMessage = ex.response?.data?['message'] ?? ex.response.toString();
      AppLogger.error('unlinkLine error: $errorMessage');
      throw Exception(errorMessage);
    }
  }
}
