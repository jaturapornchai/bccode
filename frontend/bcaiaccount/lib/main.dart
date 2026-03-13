import 'dart:async';

import 'package:smlaicloud/api/app_const.dart';
import 'package:smlaicloud/flavors.dart';
import 'package:smlaicloud/main_app.dart';
import 'package:smlaicloud/environment.dart';
import 'package:smlaicloud/firebase_options.dart';
// import 'package:smlaicloud/utils/google_sheet.dart'; // Not used
import 'package:firebase_core/firebase_core.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'global.dart' as global;
import 'package:smlaicloud/utils/backend_url_manager.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

void initializeEnvironmentConfig() {
  // เชื่อมโยง Flavor กับ Environment
  String environment;

  switch (F.appFlavor) {
    case Flavor.bcaidev:
      environment = Environment.DEV;
      break;
    case Flavor.bcaiuat:
      environment = Environment.UAT;
      break;
    case Flavor.bcaiprod:
      environment = Environment.PROD;
      break;
    default:
      environment = Environment.DEV;
      break;
  }

  Environment().initConfig(environment);
  global.isdevPin = global.checkDeveloperMode(environment);
  if (kDebugMode) {
    AppLogger.debug("isdev for active pin :  ${global.isdevPin}");
    if (kDebugMode) {
      AppLogger.debug("Current Flavor: ${F.title}");
    }
    if (kDebugMode) {
      AppLogger.debug("Current Environment: $environment");
    }
  }
}

FutureOr<void> main() async {
  initializeEnvironmentConfig();

  // Dev mode: ล้าง log file ทุกครั้งที่ restart
  if (kDebugMode) {
    AppLogger.clearLogFile();
    AppLogger.info('[App] เริ่มต้นแอป — log file ถูกล้างแล้ว');
  }
  global.prefs = await SharedPreferences.getInstance();
  // global.posVersion = global.PosVersionEnum.restaurant;
  global.posVersion = global.PosVersionEnum.pos;
  WidgetsFlutterBinding.ensureInitialized();

  // appConfig (SharedPreferences) ต้อง init ก่อน appConfigInit()
  // เพราะ BackendUrlManager อ่าน URL จาก appConfig
  global.appConfig = await SharedPreferences.getInstance();

  // โหลด config.json — frontend ชี้ไปที่ backend
  // web: อ่านจาก /config.json บน web root (environment-specific)
  // native: อ่านจาก assets/config.json (default, user override ได้ผ่าน Setup screen)
  await BackendUrlManager.loadConfig();

  try {
    global.myAppConfig = appConfigInit();
  } catch (e) {
    if (kDebugMode) {
      AppLogger.debug('Using default web config: $e');
    }
  }

  // โหลดการตั้งค่าการแสดงผล (Font Family, Zoom Level)
  await global.loadDisplaySettings();

  global.themeSelect(0);
  global.applyThemeMode();

  // if (global.apiConnected == false) {
  //   if (!global.isLoginProcess) {
  //     global.isLoginProcess = true;
  //     UserRepository userRepository = UserRepository();
  //     await userRepository
  //         .authenUser(global.apiUserName, global.apiUserPassword)
  //         .then((_result) async {
  //       if (_result.success) {
  //         global.apiConnected = true;
  //         global.apiToken = _result.data["token"];
  //         global.appConfig .write("token", _result.data["token"]);
  //         // print("Login Succerss");
  //         ApiResponse selectShop =
  //             await userRepository.selectShop(global.apiShopCode);
  //         if (selectShop.success) {
  //           // print("Select Shop Sucess");
  //         }
  //       }
  //     }).catchError((e) {
  //       // print("Login Error");
  //       // print(e);
  //     }).whenComplete(() async {
  //       global.isLoginProcess = false;
  //     });
  //   }
  // }

  // SystemChrome.setEnabledSystemUIMode(SystemUiMode.manual, overlays: []);
  // await SystemChrome.setPreferredOrientations(
  //     [DeviceOrientation.landscapeLeft, DeviceOrientation.landscapeRight]);

  if (kIsWeb) {
    // Web Base
    try {
      await Firebase.initializeApp(
        options: DefaultFirebaseOptions.currentPlatform,
      );
      if (kDebugMode) {
        AppLogger.info("Firebase initialized successfully on web");
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error initializing Firebase on web: $e');
      }
    }
  } else {
    // Non-web platforms
    try {
      await Firebase.initializeApp(
        options: DefaultFirebaseOptions.currentPlatform,
      );
      if (kDebugMode) {
        AppLogger.info("Firebase initialized successfully on native platform");
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error initializing Firebase: $e');
      }
    }
  }

  // if (shouldUseFirebaseEmulator) {
  //   await FirebaseAuth.instance.useAuthEmulator('localhost', 9099);
  // }

  ///โหลดภาษาจากไฟล์แยก (assets/language/{lang}.json)
  /// ปรับปรุงใหม่: โหลดเฉพาะภาษาที่ใช้งาน แทนการโหลดทุกภาษา
  /// ลด memory usage ~89%
  /// จำภาษาล่าสุดที่เลือกจาก SharedPreferences

  /*BlocOverrides.runZoned(
    () => runApp(const MyApp()),
    blocObserver: AppObserver(), 
  );*/
  // สร้าง Json จาก Google Sheet
  if (global.developerMode && kIsWeb == false) {
    // createJsonFromGoogleSheet();
  }

  global.deviceConfigLoad();

  /// load timezonelist
  global.getTimezones();

  // อ่านภาษาจาก SharedPreferences
  global.userLanguage = global.appConfig.getString("language") ?? "th";
  // Language จะถูกโหลดจาก API หลังจาก login สำเร็จ และตั้งค่า Backend URL แล้ว

  if (kDebugMode) {
    AppLogger.info("App started with language: ${global.userLanguage}");
  }

  runApp(const MyApp());
}
