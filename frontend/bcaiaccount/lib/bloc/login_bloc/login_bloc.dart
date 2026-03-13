// ignore_for_file: avoid_print

import 'dart:convert';

import 'package:smlaicloud/model/create_shop_model.dart';
import 'package:smlaicloud/model/timezones_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';
import 'package:smlaicloud/repositories/user_repository.dart';
import 'package:smlaicloud/model/user_login_model.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/logger/app_logger.dart';

part 'login_event.dart';
part 'login_state.dart';

class LoginBloc extends Bloc<LoginEvent, LoginState> {
  final UserRepository _userRepository;

  LoginBloc({required UserRepository userRepository})
    : _userRepository = userRepository,
      super(LoginInitial()) {
    on<LoginOnLoad>(_onLoginLoad);
    on<RegisterUser>(_onRegisterUser);
    on<TokenLogin>(_onTokenLogin);
    on<LineLogin>(_onLineLogin);
    on<GoogleLogin>(_onGoogleLogin);
    on<CreateShop>(_onCreateShop);
    on<Logout>(_onLogout);
    on<TokenInvalid>(_onTokenInvalid);
  }

  void _onLoginLoad(LoginOnLoad event, Emitter<LoginState> emit) async {
    AppLogger.info('🔐 เริ่มต้น Login process...');
    AppLogger.info('Username: ${event.userName}');

    emit(LoginInProgress());
    try {
      AppLogger.info('📡 กำลังเรียก API authenUser...');
      final result = await _userRepository.authenUser(
        event.userName,
        event.passWord,
      );

      AppLogger.info('📥 ได้รับ response จาก API');
      AppLogger.debug('Success: ${result.success}');

      if (result.success) {
        UserLoginModel userLogin = UserLoginModel(
          name: event.userName,
          token: result.data["token"],
          refreshtoken: "",
          email: "",
          code: "",
          photourl: "",
        );
        global.appConfig.setString("refreshtoken", result.data["refresh"]);
        global.appConfig.setString("token", result.data["token"]);
        global.appConfig.setString("user", event.userName);

        AppLogger.info('✅ Login สำเร็จ!');
        AppLogger.debug('Token: ${result.data["token"]}');
        emit(LoginSuccess(userLogin: userLogin));
      } else {
        AppLogger.warning('⚠️  Login ไม่สำเร็จ: result.success = false');
        emit(LoginFailed(message: 'User Not Found'));
      }
    } on Exception catch (exception) {
      AppLogger.error('❌ Login Exception: $exception');
      final msg = exception.toString().replaceFirst('Exception: ', '');
      emit(LoginFailed(message: msg));
    } catch (e) {
      AppLogger.error('❌ Login Error: $e');
      emit(LoginFailed(message: 'เกิดข้อผิดพลาดที่ไม่คาดคิด: $e'));
    }
  }

  void _onRegisterUser(RegisterUser event, Emitter<LoginState> emit) async {
    emit(RegisterInProgress());
    try {
      DateTime now = DateTime.now();
      int timeZoneOffset = now.timeZoneOffset.inHours;
      String timezonelabel = "";
      String timezonefffset = "";
      String yeartype = "christian";

      String currentTimeZone = "";
      if (kIsWeb == false) {
        // เอา timezone จาก DateTime.now()
        currentTimeZone = DateTime.now().timeZoneName;
      } else {
        currentTimeZone = "Asia/Bangkok";
      }

      // Find the first timezone that matches the given conditions
      TimezonesModel timezone = global.timezonesListData.firstWhere(
        (element) => element.offset == timeZoneOffset.abs().toString(),
      );
      for (var element in timezone.utc) {
        if (element == currentTimeZone) {
          timezonelabel = timezone.text;
          timezonefffset = timezone.offset;
          break;
        }
      }

      final result = await _userRepository.registerUser(
        event.userName,
        event.passWord,
        timezonelabel,
        timezonefffset,
        yeartype,
      );

      if (result.success) {
        emit(RegisterSuccess());
      } else {
        emit(const RegisterFailed(message: 'Not Found'));
      }
    } on Exception catch (exception) {
      final msg = exception.toString().replaceFirst('Exception: ', '');
      emit(RegisterFailed(message: msg));
    } catch (e) {
      emit(RegisterFailed(message: 'เกิดข้อผิดพลาดที่ไม่คาดคิด: $e'));
    }
  }

  void _onTokenLogin(TokenLogin event, Emitter<LoginState> emit) async {
    emit(TokenLoginInProgress());
    try {
      if (kDebugMode) {
        AppLogger.debug('🔐 _onTokenLogin: calling authenUserByToken...');
        AppLogger.debug('🔐 Token length: ${event.token.length}');
        AppLogger.debug('🔐 Token (first 80): ${event.token.substring(0, event.token.length > 80 ? 80 : event.token.length)}');
      }
      final result = await _userRepository.authenUserByToken(event.token);

      if (kDebugMode) {
        AppLogger.debug('🔐 authenUserByToken result: success=${result.success}, data=${result.data}');
      }

      if (result.success) {
        // print(result.data);
        UserLoginModel userLogin = UserLoginModel(
          name: "",
          token: result.data["token"],
          refreshtoken: "",
          email: "",
          code: "",
          photourl: "",
        );

        global.appConfig.setString("token", result.data["token"]);
        if (kDebugMode) {
          AppLogger.debug("Token: ${result.data["token"]}");
        }
        emit(TokenLoginSuccess(userLogin: userLogin));
      } else {
        emit(const TokenLoginFailed(message: 'User Not Found'));
      }
    } on Exception catch (exception) {
      final msg = exception.toString().replaceFirst('Exception: ', '');
      emit(TokenLoginFailed(message: msg));
    } catch (e) {
      emit(TokenLoginFailed(message: 'เกิดข้อผิดพลาดที่ไม่คาดคิด: $e'));
    }
  }

  /// LINE Login handler
  /// ส่งข้อมูล LINE user ไปยัง Main API เพื่อสร้าง/ดึง user และรับ JWT token กลับมา
  void _onLineLogin(LineLogin event, Emitter<LoginState> emit) async {
    emit(LineLoginInProgress());
    try {
      // เรียก Main API เพื่อ login ด้วยข้อมูล LINE user
      final result = await _userRepository.authenUserByLineToken(
        event.lineUserId,
        displayName: event.displayName,
        pictureUrl: event.pictureUrl,
        email: event.email,
      );

      if (result.success) {
        // ใช้ username จริงจาก API (email ของ user ที่เชื่อมต่อ LINE)
        // ไม่ใช่ LINE display name
        final actualUsername = result.data["username"] ?? event.email ?? "";

        UserLoginModel userLogin = UserLoginModel(
          name: event.displayName ?? "",
          token: result.data["token"],
          refreshtoken: "",
          email: actualUsername,
          code: "",
          photourl: event.pictureUrl ?? "",
        );

        // บันทึก token + username จริง
        global.appConfig.setString("token", result.data["token"]);
        global.appConfig.setString("user", actualUsername);

        if (kDebugMode) {
          AppLogger.debug("LINE Login Success!");
          AppLogger.debug("LINE Login Token: ${result.data["token"]}");
          AppLogger.debug("LINE Login Username: $actualUsername");
          AppLogger.debug("LINE Login LINE Name: ${event.displayName}");
          AppLogger.debug("LINE Login Photo: ${event.pictureUrl}");
        }

        emit(LineLoginSuccess(userLogin: userLogin));
      } else {
        emit(const LineLoginFailed(message: 'ไม่สามารถเข้าสู่ระบบด้วย LINE ได้'));
      }
    } on Exception catch (exception) {
      // ดึง message จาก Exception — ตัด "Exception: " prefix ออก
      String msg = exception.toString();
      if (msg.startsWith('Exception: ')) {
        msg = msg.substring('Exception: '.length);
      }
      emit(LineLoginFailed(message: msg));
    } catch (e) {
      emit(LineLoginFailed(message: 'เกิดข้อผิดพลาด: $e'));
    }
  }

  /// Google Login handler
  /// ส่งข้อมูล Google user ไปยัง Main API เพื่อสร้าง/ดึง user และรับ JWT token กลับมา
  void _onGoogleLogin(GoogleLogin event, Emitter<LoginState> emit) async {
    emit(GoogleLoginInProgress());
    try {
      // เรียก Main API เพื่อ login ด้วยข้อมูล Google user
      final result = await _userRepository.authenUserByGoogleToken(
        event.googleUserId,
        displayName: event.displayName,
        pictureUrl: event.pictureUrl,
        email: event.email,
      );

      if (result.success) {
        // สร้าง UserLoginModel จากข้อมูลที่ได้จาก Main API
        UserLoginModel userLogin = UserLoginModel(
          name: event.displayName ?? "",
          token: result.data["token"], // JWT token จาก Main API
          refreshtoken: "",
          email: event.email ?? "",
          code: "",
          photourl: event.pictureUrl ?? "",
        );

        // บันทึก token สำหรับใช้งาน API อื่นๆ
        global.appConfig.setString("token", result.data["token"]);

        if (kDebugMode) {
          AppLogger.debug("Google Login Success!");
          AppLogger.debug("Google Login Token: ${result.data["token"]}");
          AppLogger.debug("Google Login Name: ${event.displayName}");
          AppLogger.debug("Google Login Email: ${event.email}");
          AppLogger.debug("Google Login Photo: ${event.pictureUrl}");
        }

        emit(GoogleLoginSuccess(userLogin: userLogin));
      } else {
        emit(const GoogleLoginFailed(message: 'Google Login Failed'));
      }
    } on Exception catch (exception) {
      emit(GoogleLoginFailed(message: 'Google Login Failed: $exception'));
    } catch (e) {
      emit(GoogleLoginFailed(message: 'Google Login Failed: $e'));
    }
  }

  void _onCreateShop(CreateShop event, Emitter<LoginState> emit) async {
    if (kDebugMode) {
      AppLogger.debug('🏪 _onCreateShop: start');
      AppLogger.debug('🏪 shopData: ${jsonEncode(event.createShop.toJson())}');
    }
    emit(CreateShopInProgress());
    try {
      await _userRepository.createShop(event.createShop);
      if (kDebugMode) {
        AppLogger.debug('🏪 createShop: success!');
      }
      emit(CreateShopSuccess());
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('🏪 createShop error: $e');
        AppLogger.error('🏪 createShop error type: ${e.runtimeType}');
      }
      try {
        final error = jsonDecode(e.toString());
        emit(CreateShopFailed(message: error['message']));
      } catch (_) {
        emit(CreateShopFailed(message: e.toString()));
      }
    }
  }

  /// Logout
  /// Clear all data in GetStorage
  void _onLogout(Logout event, Emitter<LoginState> emit) async {
    emit(LogoutInProgress());
    try {
      // แก้ปัญหาเบื้องต้น ไม่ให้ลบ token เพราะ ใน pos และ kiosk จะใช้ token นี้ ไม่ได้
      // await _userRepository.logout();

      // clear เฉพาะ session token — ไม่แตะ saved_username/saved_password/remember_password
      global.appConfig.remove('token');
      global.appConfig.remove('shopid');
      global.appConfig.remove('shopname');
      global.appConfig.remove('role');

      global.prefs.remove("user");
      global.userLoginData = UserLoginModel(
        name: "",
        code: "",
        token: "",
        email: "",
        refreshtoken: "",
        photourl: "",
      );
      // ล้าง social login profile data (Google/LINE)
      global.loginName = '';
      global.loginEmail = '';
      global.loginPhotoUrl = '';
      emit(LogoutSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(LogoutFailed(message: error['message']));
    }
  }

  /// handler token invalid
  /// Clear all data in GetStorage
  void _onTokenInvalid(TokenInvalid event, Emitter<LoginState> emit) async {
    emit(LogoutInProgress());
    try {
      // clear เฉพาะ session token — ไม่แตะ saved credentials
      global.appConfig.remove('token');
      global.appConfig.remove('shopid');
      global.appConfig.remove('shopname');
      global.appConfig.remove('role');
      emit(LogoutSuccess());
    } catch (e) {
      final error = jsonDecode(e.toString());
      emit(LogoutFailed(message: error['message']));
    }
  }
}
