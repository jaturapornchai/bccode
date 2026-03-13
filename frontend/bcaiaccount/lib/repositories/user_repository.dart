import 'package:flutter/foundation.dart';
import 'dart:convert';

import 'package:smlaicloud/model/profile_model.dart';
import 'package:smlaicloud/model/user_model.dart';
import 'package:smlaicloud/model/create_shop_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class UserRepository {
  Future<ApiResponse> registerUser(String userName, String passWord, String timezonelabel, String timezoneoffset, String yeartype) async {
    Dio client = Client().init();

    try {
      final response = await client.post(
        '/register',
        data: {"name": userName, "username": userName, "password": passWord, "timezonelabel": timezonelabel, "timezoneoffset": timezoneoffset, "yeartype": yeartype},
      );
      try {
        final result = json.decode(response.toString());
        final rawData = {"success": result["success"], "data": result};

        if (rawData['error'] != null) {
          String errorMessage = '${rawData['code']}: ${rawData['message']}';
          throw Exception(errorMessage);
        }

        return ApiResponse.fromMap(rawData);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      if (ex.type == DioExceptionType.connectionTimeout) {
        throw Exception('Connection Timeout');
      }
      if (ex.type == DioExceptionType.receiveTimeout) {
        throw Exception('unable to connect to the server');
      }
      throw Exception(errorMessage);
    }
  }

  /// สมัครสมาชิกด้วยรหัสพนักงาน + รหัสผ่าน (ไม่ต้องมี email)
  Future<ApiResponse> registerByUsername(String username, String password, String timezonelabel, String timezoneoffset, String yeartype, {String? name}) async {
    Dio client = Client().init();

    try {
      final response = await client.post(
        '/register-username',
        data: {
          "username": username,
          "password": password,
          "name": name ?? username,
          "timezonelabel": timezonelabel,
          "timezoneoffset": timezoneoffset,
          "yeartype": yeartype,
        },
      );
      try {
        final result = json.decode(response.toString());
        final rawData = {"success": result["success"], "data": result};

        if (rawData['error'] != null) {
          String errorMessage = '${rawData['code']}: ${rawData['message']}';
          throw Exception(errorMessage);
        }

        return ApiResponse.fromMap(rawData);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      if (ex.type == DioExceptionType.connectionTimeout) {
        throw Exception('Connection Timeout');
      }
      if (ex.type == DioExceptionType.receiveTimeout) {
        throw Exception('unable to connect to the server');
      }
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> authenUser(String userName, String passWord) async {
    Dio client = Client().init();

    if (kDebugMode) {
      AppLogger.info('=== authenUser (Login) ===');
      AppLogger.info('API Base URL: ${client.options.baseUrl}');
      AppLogger.info('Full URL: ${client.options.baseUrl}login');
      AppLogger.info('Username: $userName');
    }

    try {
      final response = await client.post('/login', data: {"username": userName, "password": passWord});

      if (kDebugMode) {
        AppLogger.info('Login response received: ${response.statusCode}');
        AppLogger.debug('Response data: ${response.toString()}');
      }

      try {
        final result = json.decode(response.toString());
        final rawData = {"success": result["success"], "data": result};

        if (rawData['error'] != null) {
          throw Exception('${rawData['code']}: ${rawData['message']}');
        }

        if (kDebugMode) {
          AppLogger.info('✅ Login successful!');
        }

        return ApiResponse.fromMap(rawData);
      } catch (ex) {
        if (kDebugMode) {
          AppLogger.error('❌ Error parsing login response: $ex');
        }
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      // ดึง error message จาก response body ถ้ามี
      String errorMessage = 'เข้าสู่ระบบไม่สำเร็จ';

      if (kDebugMode) {
        AppLogger.error('=== Login DioException ===');
        AppLogger.error('Exception type: ${ex.type}');
        AppLogger.error('Error message: ${ex.message}');
        AppLogger.error('Response status: ${ex.response?.statusCode}');
        AppLogger.error('Response data: ${ex.response?.data}');
      }

      if (ex.response?.data != null) {
        final data = ex.response!.data;
        if (data is Map && data['message'] != null) {
          errorMessage = data['message'].toString();
        } else if (data is String) {
          errorMessage = data;
        }
      } else if (ex.type == DioExceptionType.connectionTimeout) {
        errorMessage = 'เชื่อมต่อ Server ไม่ทันเวลา (Connection Timeout)';
      } else if (ex.type == DioExceptionType.receiveTimeout) {
        errorMessage = 'Server ไม่ตอบกลับทันเวลา (Receive Timeout)';
      } else if (ex.type == DioExceptionType.connectionError) {
        errorMessage = 'ไม่สามารถเชื่อมต่อ Server ได้ (Server อาจไม่ได้เปิด)';
      } else if (ex.type == DioExceptionType.unknown) {
        errorMessage = 'ไม่สามารถเชื่อมต่อ Server ได้ (ตรวจสอบว่า Server เปิดอยู่)';
      } else if (ex.type == DioExceptionType.sendTimeout) {
        errorMessage = 'ส่งข้อมูลไม่ทันเวลา (Send Timeout)';
      } else if (ex.type == DioExceptionType.badResponse) {
        errorMessage = 'Server ตอบกลับผิดปกติ (${ex.response?.statusCode ?? 'unknown'})';
      }

      if (kDebugMode) {
        AppLogger.error('Final error message: $errorMessage');
      }

      throw Exception(errorMessage);
    }
  }

  /// token login
  Future<ApiResponse> authenUserByToken(String token) async {
    Dio client = Client().init();
    if (kDebugMode) {
      AppLogger.debug('📡 [authenUserByToken] baseUrl: ${client.options.baseUrl}');
      AppLogger.debug('📡 [authenUserByToken] endpoint: POST /tokenlogin');
      AppLogger.debug('📡 [authenUserByToken] token length: ${token.length}');
    }

    try {
      final response = await client.post('/tokenlogin', data: {"token": token});
      if (kDebugMode) {
        AppLogger.debug('📥 [authenUserByToken] status: ${response.statusCode}');
        AppLogger.debug('📥 [authenUserByToken] response: ${response.data}');
      }
      try {
        final result = json.decode(response.toString());
        final rawData = {"success": result["success"], "data": result};

        if (rawData['error'] != null) {
          throw Exception('${rawData['code']}: ${rawData['message']}');
        }

        return ApiResponse.fromMap(rawData);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      if (kDebugMode) {
        AppLogger.error('❌ [authenUserByToken] DioException: type=${ex.type}');
        AppLogger.error('❌ [authenUserByToken] status: ${ex.response?.statusCode}');
        AppLogger.error('❌ [authenUserByToken] response: ${ex.response?.data}');
        AppLogger.error('❌ [authenUserByToken] requestUrl: ${ex.requestOptions.uri}');
      }
      String errorMessage = ex.response.toString();
      if (ex.type == DioExceptionType.connectionTimeout) {
        throw Exception('Connection Timeout');
      }
      if (ex.type == DioExceptionType.receiveTimeout) {
        throw Exception('unable to connect to the server');
      }

      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> getShopList() async {
    Dio client = Client().init();

    try {
      final response = await client.get('/list-shop?limit=100');
      return ApiResponse.fromMap(response.data);
    } on DioException catch (ex) {
      throw Exception(ex.response?.data?.toString() ?? ex.message);
    }
  }

  Future<ApiResponse> selectShop(String shopid) async {
    Dio client = Client().init();

    try {
      final response = await client.post('/select-shop', data: {"shopid": shopid});
      return ApiResponse.fromMap(response.data);
    } on DioException catch (ex) {
      throw Exception(ex.response?.data?.toString() ?? ex.message);
    }
  }

  Future<ApiResponse> createShop(CreateShopModel createShop) async {
    Dio client = Client().init();
    createShop.settings!.languageconfigs![0].isdefault = true;
    final data = createShop.toJson();
    if (kDebugMode) {
      AppLogger.debug('📡 [createShop] baseUrl: ${client.options.baseUrl}');
      AppLogger.debug('📡 [createShop] POST /create-shop');
      AppLogger.debug('📡 [createShop] data: $data');
    }
    try {
      final response = await client.post('/create-shop', data: data);
      if (kDebugMode) {
        AppLogger.debug('📥 [createShop] status: ${response.statusCode}');
        AppLogger.debug('📥 [createShop] response: ${response.data}');
      }
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      if (kDebugMode) {
        AppLogger.error('❌ [createShop] DioException: ${ex.type}');
        AppLogger.error('❌ [createShop] status: ${ex.response?.statusCode}');
        AppLogger.error('❌ [createShop] response: ${ex.response?.data}');
        AppLogger.error('❌ [createShop] url: ${ex.requestOptions.uri}');
      }
      String errorMessage = ex.response.toString();
      throw errorMessage;
    }
  }

  Future<dynamic> requestOPT(String telephoneNumber) async {
    Dio client = Dio();

    try {
      final response = await client.post(
        'https://smsapi.deecommerce.co.th:4300/service/v1/otp/request',
        data: {
          "accountId": "08992231310610",
          "secretKey": "U2FsdGVkX19gSK0SR/xX5DAa6B2Mn1wDyEo1es83LNQ=",
          "type": "OTP",
          "lang": 'th',
          "to": telephoneNumber,
          "sender": "deeSMS.OTP",
          "isShowRef": '1',
        },
      );
      try {
        final rawData = json.decode(response.toString());

        if (rawData['error'] != '0') {
          throw Exception('${rawData['msg']}');
        }

        return rawData['result'];
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  Future<bool> verifyOPT(String token, String pin) async {
    Dio client = Dio();

    try {
      final response = await client.post(
        'https://smsapi.deecommerce.co.th:4300/service/v1/otp/verify',
        data: {"accountId": "08992231310610", "secretKey": "U2FsdGVkX19gSK0SR/xX5DAa6B2Mn1wDyEo1es83LNQ=", "token": token, "pin": pin},
      );
      try {
        final rawData = json.decode(response.toString());

        if (rawData['error'] != '0') {
          throw Exception('${rawData['msg']}');
        }

        return true;
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  /// ดึงผู้ใช้งานทั้งหมด
  Future<ApiResponse> getUserList({int limit = 0, int offset = 0, String search = ""}) async {
    Dio client = Client().init();

    try {
      String query = "/shop/users?offset=$offset&limit=$limit&q=$search&sort=code:1";
      final response = await client.get(query);
      try {
        final rawData = json.decode(response.toString());
        if (rawData['error'] != null) {
          throw Exception('${rawData['code']}: ${rawData['message']}');
        }
        return ApiResponse.fromMap(rawData);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  /// ลบผู้ใช้งาน
  Future<ApiResponse> deleteUser(String username) async {
    Dio client = Client().init();
    // URL encode username เพราะอาจมี @ หรืออักขระพิเศษ
    final encodedUsername = Uri.encodeComponent(username);
    debugPrint('deleteUser - username: $username, encoded: $encodedUsername');
    try {
      final response = await client.delete('/shop/permission/$encodedUsername');
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  /// บันทึกและแก้ไขผู้ใช้งาน
  Future<ApiResponse> saveAndUpdateUser(UserModel userModel) async {
    Dio client = Client().init();
    final data = userModel.toJson();

    // Debug log: ดูข้อมูลที่ส่งไป mainapi
    if (kDebugMode) {
      AppLogger.debug('=== saveAndUpdateUser ===');
      AppLogger.debug('API Base URL: ${client.options.baseUrl}');
      AppLogger.debug('Full URL: ${client.options.baseUrl}shop/permission');
      AppLogger.debug('Sending to /shop/permission:');
      AppLogger.debug('  username: ${data['username']}');
      AppLogger.debug('  line_user_id: ${data['line_user_id']}');
      AppLogger.debug('  line_display_name: ${data['line_display_name']}');
      AppLogger.debug('  line_picture_url: ${data['line_picture_url']}');
      AppLogger.debug('  position: ${data['position']}');
      AppLogger.debug('  department: ${data['department']}');
      AppLogger.debug('  Full JSON: $data');
    }

    try {
      final response = await client.put('/shop/permission', data: data);
      try {
        if (kDebugMode) {
          AppLogger.debug('Response from /shop/permission: ${response.data}');
        }
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  /// ดึงผู้ใช้งานตาม username
  Future<ApiResponse> getUser(String username) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/shop/permission/$username');
      try {
        // Debug log: ดูข้อมูลที่ได้รับจาก mainapi
        if (kDebugMode) {
          AppLogger.debug('=== getUser ===');
          AppLogger.debug('API Base URL: ${client.options.baseUrl}');
          AppLogger.debug('Full URL: ${client.options.baseUrl}shop/permission/$username');
          AppLogger.debug('Response from /shop/permission/$username: ${response.data}');
          final data = response.data['data'];
          if (data != null) {
            AppLogger.debug('  line_user_id: ${data['line_user_id']}');
            AppLogger.debug('  line_display_name: ${data['line_display_name']}');
            AppLogger.debug('  line_picture_url: ${data['line_picture_url']}');
          }
        }
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

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

  /// LINE Login - ใช้ข้อมูล LINE user ในการเข้าสู่ระบบผ่าน Main API
  Future<ApiResponse> authenUserByLineToken(String lineUserId, {String? displayName, String? pictureUrl, String? email}) async {
    Dio client = Client().init();
    if (kDebugMode) {
      AppLogger.debug('LINE Login with line_user_id: $lineUserId');
      AppLogger.debug('LINE Login display_name: $displayName');
      AppLogger.debug('LINE Login picture_url: $pictureUrl');
      AppLogger.debug('LINE Login email: $email');
    }

    try {
      final response = await client.post('/linelogin', data: {"line_user_id": lineUserId, "display_name": displayName ?? "", "picture_url": pictureUrl ?? "", "email": email ?? ""});
      try {
        final result = json.decode(response.toString());
        final rawData = {"success": result["success"], "data": result};

        if (rawData['error'] != null) {
          throw Exception('${rawData['code']}: ${rawData['message']}');
        }

        return ApiResponse.fromMap(rawData);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      if (ex.type == DioExceptionType.connectionTimeout) {
        throw Exception('Connection Timeout');
      }
      if (ex.type == DioExceptionType.receiveTimeout) {
        throw Exception('unable to connect to the server');
      }
      // ดึง message จาก API response JSON (เช่น "ไม่พบบัญชีที่เชื่อมต่อ LINE นี้")
      String errorMessage = ex.response.toString();
      if (ex.response?.data is Map) {
        errorMessage = ex.response?.data['message'] ?? errorMessage;
      } else {
        try {
          final decoded = json.decode(ex.response.toString());
          if (decoded is Map && decoded['message'] != null) {
            errorMessage = decoded['message'];
          }
        } catch (_) {}
      }
      throw Exception(errorMessage);
    }
  }

  /// Google Login - ใช้ข้อมูล Google user ในการเข้าสู่ระบบผ่าน Main API
  Future<ApiResponse> authenUserByGoogleToken(String googleUserId, {String? displayName, String? pictureUrl, String? email}) async {
    Dio client = Client().init();
    if (kDebugMode) {
      AppLogger.debug('Google Login with google_user_id: $googleUserId');
      AppLogger.debug('Google Login display_name: $displayName');
      AppLogger.debug('Google Login picture_url: $pictureUrl');
      AppLogger.debug('Google Login email: $email');
    }

    try {
      final response = await client.post('/googlelogin', data: {"google_user_id": googleUserId, "display_name": displayName ?? "", "picture_url": pictureUrl ?? "", "email": email ?? ""});
      try {
        final result = json.decode(response.toString());
        final rawData = {"success": result["success"], "data": result};

        if (rawData['error'] != null) {
          throw Exception('${rawData['code']}: ${rawData['message']}');
        }

        return ApiResponse.fromMap(rawData);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      if (ex.type == DioExceptionType.connectionTimeout) {
        throw Exception('Connection Timeout');
      }
      if (ex.type == DioExceptionType.receiveTimeout) {
        throw Exception('unable to connect to the server');
      }
      throw Exception(errorMessage);
    }
  }

  /// Logout
  Future<ApiResponse> logout() async {
    Dio client = Client().init();
    try {
      final response = await client.post('/logout');
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  /// profile shop
  Future<ApiResponse> loadProfileShop() async {
    Dio client = Client().init();
    try {
      final response = await client.get('/profileshop');
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      throw Exception(ex);
    }
  }

  /// บันทึก LINE data ของตัวเอง (ใช้หลัง LIFF linking)
  /// ไม่ต้องใช้ owner permission เพราะเป็นการแก้ไขข้อมูลตัวเอง
  Future<ApiResponse> saveMyLineData({
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

    if (kDebugMode) {
      AppLogger.debug('=== saveMyLineData ===');
      AppLogger.debug('API Base URL: ${client.options.baseUrl}');
      AppLogger.debug('Sending to /profile/my-line:');
      AppLogger.debug('  line_user_id: $lineUserId');
      AppLogger.debug('  line_display_name: $lineDisplayName');
      AppLogger.debug('  line_picture_url: $linePictureUrl');
    }

    try {
      final response = await client.put('/profile/my-line', data: data);
      try {
        if (kDebugMode) {
          AppLogger.debug('Response from /profile/my-line: ${response.data}');
        }
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }
}
