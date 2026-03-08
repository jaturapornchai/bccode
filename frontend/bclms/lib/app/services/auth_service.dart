import 'package:dio/dio.dart';
import 'package:bclms/app/core/constants/api_constants.dart';
import 'package:bclms/app/core/utils/app_logger.dart';
import 'package:bclms/app/models/branch_model.dart';
import 'package:bclms/app/models/login_response.dart';
import 'package:bclms/app/models/profile_model.dart';
import 'package:bclms/app/models/shop_model.dart';
import 'package:bclms/app/services/api_service.dart';
import 'package:bclms/app/services/storage_service.dart';

class AuthService {
  final ApiService _api;
  final StorageService _storage;

  AuthService(this._api, this._storage);

  Dio get _dio => _api.dio;

  /// POST /login
  Future<LoginResponse> login(String username, String password) async {
    try {
      final response = await _dio.post(
        ApiConstants.login,
        data: {'username': username, 'password': password},
      );
      final result = LoginResponse.fromJson(response.data);
      if (result.success) {
        await _storage.saveTokens(
          token: result.token!,
          refresh: result.refresh!,
        );
        AppLogger.info('เข้าสู่ระบบสำเร็จ: $username');
      }
      return result;
    } on DioException catch (e) {
      AppLogger.error('เข้าสู่ระบบล้มเหลว', error: e);
      final data = e.response?.data;
      if (data is Map<String, dynamic>) {
        return LoginResponse.fromJson(data);
      }
      return LoginResponse(success: false, message: 'ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้');
    }
  }

  /// GET /verify-token
  Future<bool> verifyToken() async {
    try {
      final response = await _dio.get(ApiConstants.verifyToken);
      if (response.data['success'] == true) {
        await _storage.saveUserInfo(
          username: response.data['username'] ?? '',
          name: response.data['name'] ?? '',
        );
        AppLogger.info('ยืนยัน token สำเร็จ');
        return true;
      }
      return false;
    } catch (e) {
      AppLogger.error('ยืนยัน token ล้มเหลว', error: e);
      return false;
    }
  }

  /// GET /list-shop
  Future<List<ShopModel>> listShops({int page = 1, int limit = 100}) async {
    try {
      final response = await _dio.get(
        ApiConstants.listShop,
        queryParameters: {'page': page, 'limit': limit},
      );
      if (response.data['success'] == true) {
        final list = response.data['data'] as List;
        AppLogger.info('โหลดรายการร้านค้าสำเร็จ: ${list.length} ร้าน');
        return list.map((e) => ShopModel.fromJson(e)).toList();
      }
      return [];
    } catch (e) {
      AppLogger.error('โหลดรายการร้านค้าล้มเหลว', error: e);
      return [];
    }
  }

  /// POST /select-shop
  Future<bool> selectShop(ShopModel shop) async {
    try {
      final response = await _dio.post(
        ApiConstants.selectShop,
        data: {'shopid': shop.shopId},
      );
      if (response.data['success'] == true) {
        await _storage.saveShopData(
          shopId: shop.shopId,
          shopName: shop.shopName,
          role: shop.role,
        );
        AppLogger.info('เลือกร้านค้าสำเร็จ: ${shop.shopName}');
        return true;
      }
      return false;
    } catch (e) {
      AppLogger.error('เลือกร้านค้าล้มเหลว', error: e);
      return false;
    }
  }

  /// GET /branch
  Future<List<BranchModel>> listBranches() async {
    try {
      final response = await _dio.get(ApiConstants.branch);
      if (response.data['success'] == true) {
        final list = response.data['data'] as List;
        AppLogger.info('โหลดรายการสาขาสำเร็จ: ${list.length} สาขา');
        return list.map((e) => BranchModel.fromJson(e)).toList();
      }
      return [];
    } catch (e) {
      AppLogger.error('โหลดรายการสาขาล้มเหลว', error: e);
      return [];
    }
  }

  /// GET /profile
  Future<ProfileModel?> getProfile() async {
    try {
      final response = await _dio.get(ApiConstants.profile);
      if (response.data['success'] == true) {
        return ProfileModel.fromJson(response.data['data']);
      }
      return null;
    } catch (e) {
      AppLogger.error('โหลดโปรไฟล์ล้มเหลว', error: e);
      return null;
    }
  }

  /// POST /logout
  Future<void> logout() async {
    try {
      await _dio.post(ApiConstants.logout);
      AppLogger.info('ออกจากระบบสำเร็จ');
    } catch (e) {
      AppLogger.warning('เรียก logout API ล้มเหลว (ดำเนินการต่อ)');
    } finally {
      await _storage.clearAll();
    }
  }
}
