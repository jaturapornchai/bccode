import 'dart:async';
import 'dart:convert';
import 'dart:io';
import 'package:smlaicloud/environment.dart';
import 'package:smlaicloud/flavors.dart';
import 'package:smlaicloud/model/user_login_model.dart';
import 'package:smlaicloud/select_language_screen.dart';
import 'package:smlaicloud/usersystem/utils.dart';
import 'package:firebase_auth/firebase_auth.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:google_sign_in/google_sign_in.dart';
import 'package:smlaicloud/bloc/login_bloc/login_bloc.dart';

import 'package:smlaicloud/global.dart' as global;
import 'package:url_launcher/url_launcher.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/screens/setup/setup_screen.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => LoginScreenState();
}

class LoginScreenState extends State<LoginScreen> with global.ThemeRefreshMixin {
  late FirebaseAuth _auth;

  Future<UserCredential?> googleSignInForWeb() async {
    try {
      // Trigger the authentication flow
      // ใช้ clientId เฉพาะสำหรับ Windows Desktop
      final googleSignIn = GoogleSignIn.instance;
      if (!kIsWeb && Platform.isWindows) {
        await googleSignIn.initialize(clientId: global.googleSignInClientIdWindows);
      }
      final GoogleSignInAccount googleUser = await googleSignIn.authenticate();

      // authentication is synchronous in v7.x, not a Future
      final GoogleSignInAuthentication googleAuth = googleUser.authentication;

      // Create a new credential
      final OAuthCredential credential = GoogleAuthProvider.credential(
        idToken: googleAuth.idToken,
        // Note: accessToken is no longer available in google_sign_in 7.x
      );

      // print(googleUser);

      // Sign in to Firebase with the Google [UserCredential]

      return await _auth.signInWithCredential(credential);
    } catch (e) {
      // print(e);
      return null;
    }
  }

  Future<String?> getCurrentUserIdToken() async {
    User? currentUser = _auth.currentUser;
    if (currentUser != null) {
      String? idToken = await currentUser.getIdToken();
      return idToken;
    } else {
      // No user is signed in.
      // print("No user is signed in");
      return null;
    }
  }

  @override
  void initState() {
    if (kIsWeb) {
      _auth = FirebaseAuth.instance;
    } else if (Platform.isWindows || Platform.isMacOS) {
      _auth = FirebaseAuth.instance;
    }
    try {
      // load user profile from local storage
      var userValue = global.prefs.getString("user");
      if (userValue != null) {
        global.userLoginData = UserLoginModel.fromJson(jsonDecode(userValue));
        setState(() {});
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.debug(e);
      }
    }

    // Debug: ตรวจสอบค่า myAppConfig
    if (kDebugMode) {
      AppLogger.debug("=== DEBUG myAppConfig ===");
      AppLogger.debug("serviceApi: ${global.myAppConfig.serviceApi}");
      AppLogger.debug("servicePort: ${global.myAppConfig.servicePort}");
      AppLogger.debug("reportApiPath: ${global.myAppConfig.reportApiPath}");
      AppLogger.debug("reportApiPort: ${global.myAppConfig.reportApiPort}");
      AppLogger.debug("Environment serviceApi: ${Environment().config.serviceApi}");
      AppLogger.debug("=========================");
    }

    super.initState();
  }

  /// Get language selection text based on current language
  /// แสดงข้อความเลือกภาษาตามภาษาปัจจุบัน (ใช้ multi-language)
  String _getLanguageText() {
    switch (global.userLanguage) {
      case 'en':
        return global.language('select_language_english');
      case 'th':
        return global.language('select_language_thai');
      case 'lo':
        return global.language('select_language_lao');
      case 'cn':
        return global.language('select_language_chinese');
      case 'ja':
        return global.language('select_language_japanese');
      case 'ko':
        return global.language('select_language_korean');
      case 'my':
        return global.language('select_language_myanmar');
      case 'km':
        return global.language('select_language_khmer');
      case 'vi':
        return global.language('select_language_vietnamese');
      default:
        return global.language('select_language_english');
    }
  }

  @override
  Widget build(BuildContext context) {
    return BlocListener<LoginBloc, LoginState>(
      listener: (context, state) {
        if (state is TokenLoginSuccess) {
          if (state.userLogin.token != '') {
            Navigator.pushNamedAndRemoveUntil(
              context,
              '/login_screen_shop',
              (route) => false,
            );
          }
        } else if (state is TokenLoginFailed) {
          global.showSnackBar(
            context,
            const Icon(Icons.error, color: Colors.white),
            state.message,
            Colors.red,
          );
          removeUser();
        }
      },
      child: Scaffold(
        body: Stack(
          children: [
            Container(color: const Color(0xFFF5F7FA)), // SAP Fiori Background
            Center(
              child: Container(
                padding: const EdgeInsets.all(20),
                child: loginByGoogle(),
              ),
            ),
          ],
        ),
      ),
    );
  }

  void removeUser() {
    global.prefs.remove("user");
    global.userLoginData = UserLoginModel(
      name: "",
      code: "",
      token: "",
      email: "",
      refreshtoken: "",
      photourl: "",
    );
    setState(() {});
  }

  String _getEnvironmentName() {
    switch (F.appFlavor) {
      case Flavor.bcaidev:
        return 'BC DEV';
      case Flavor.bcaiuat:
        return 'BC UAT';
      case Flavor.bcaiprod:
        return 'BC PROD';
      default:
        return kDebugMode ? 'DEV' : 'PROD';
    }
  }

  Widget loginByGoogle() {
    Widget userLoginWidget = (global.userLoginData.token.isNotEmpty)
        ? Column(
            children: [
              Container(
                margin: const EdgeInsets.all(10),
                decoration: BoxDecoration(
                  color: global.theme.cardColor,
                  boxShadow: [
                    BoxShadow(
                      color: Colors.grey,
                      blurRadius: 10.0,
                      spreadRadius: 0.0,
                      offset: Offset(2.0, 2.0),
                    ),
                  ],
                ),
                child: global.userLoginData.photourl.isNotEmpty
                    ? Image.network(
                        global.userLoginData.photourl,
                        width: 100,
                        height: 100,
                        fit: BoxFit.cover,
                        errorBuilder: (context, error, stackTrace) {
                          return Container(
                            width: 100,
                            height: 100,
                            color: Colors.grey.shade300,
                            child: const Icon(
                              Icons.person,
                              size: 60,
                              color: Colors.grey,
                            ),
                          );
                        },
                        loadingBuilder: (context, child, loadingProgress) {
                          if (loadingProgress == null) return child;
                          return Container(
                            width: 100,
                            height: 100,
                            color: Colors.grey.shade200,
                            child: Center(
                              child: CircularProgressIndicator(
                                value:
                                    loadingProgress.expectedTotalBytes != null
                                    ? loadingProgress.cumulativeBytesLoaded /
                                          loadingProgress.expectedTotalBytes!
                                    : null,
                              ),
                            ),
                          );
                        },
                      )
                    : Container(
                        width: 100,
                        height: 100,
                        color: Colors.grey.shade300,
                        child: const Icon(
                          Icons.person,
                          size: 60,
                          color: Colors.grey,
                        ),
                      ),
              ),
              const SizedBox(height: 10),
              Text(
                global.userLoginData.email,
                style: const TextStyle(
                  color: Color(0xFF354A5F), // SAP Shell Header Color
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                ),
              ),
              SizedBox(height: 10),
              FittedBox(
                child: Text(
                  "${global.language("welcome")} ${global.userLoginData.name}",
                  style: const TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
              SizedBox(height: 10),
              SizedBox(
                width: double.infinity,
                height: 50,
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(0xFF0A6ED1), // SAP Action Blue
                  ),
                  onPressed: () {
                    if (mounted) {
                      context.read<LoginBloc>().add(
                        TokenLogin(token: global.userLoginData.token),
                      );
                    }
                  },
                  child: Container(
                    width: double.infinity,
                    padding: EdgeInsets.all(10),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(Icons.login),
                        SizedBox(width: 12),
                        Text(
                          global.language("login"),
                          style: TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
              SizedBox(height: 10),
              SizedBox(
                width: double.infinity,
                height: 50,
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: const Color(
                      0xFFAB2217,
                    ), // SAP Negative Red
                  ),
                  onPressed: () {
                    removeUser();
                  },
                  child: Container(
                    width: double.infinity,
                    padding: EdgeInsets.all(10),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(Icons.logout),
                        SizedBox(width: 8),
                        Text(
                          global.language("logout"),
                          style: TextStyle(
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ],
                    ),
                  ),
                ),
              ),
            ],
          )
        : Container();

    return SingleChildScrollView(
      child: ConstrainedBox(
        constraints: const BoxConstraints(maxWidth: 400),
        child: Container(
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            color: global.theme.cardColor,
            borderRadius: const BorderRadius.all(Radius.circular(10)),
            boxShadow: [
              // เงาด้านบนซ้าย
              BoxShadow(
                color: Colors.grey.withValues(alpha: 0.3),
                blurRadius: 10.0,
                spreadRadius: 2.0,
                offset: const Offset(-2.0, -2.0), // shadow direction: top left
              ),
              // เงาด้านล่างขวา
              BoxShadow(
                color: Colors.grey.withValues(alpha: 0.5),
                blurRadius: 10.0,
                spreadRadius: 2.0,
                offset: const Offset(
                  2.0,
                  2.0,
                ), // shadow direction: bottom right
              ),
              // เงาด้านบนขวา
              BoxShadow(
                color: Colors.grey.withValues(alpha: 0.3),
                blurRadius: 10.0,
                spreadRadius: 2.0,
                offset: const Offset(2.0, -2.0), // shadow direction: top right
              ),
              // เงาด้านล่างซ้าย
              BoxShadow(
                color: Colors.grey.withValues(alpha: 0.3),
                blurRadius: 10.0,
                spreadRadius: 2.0,
                offset: const Offset(
                  -2.0,
                  2.0,
                ), // shadow direction: bottom left
              ),
            ],
          ),
          child: Stack(
            children: [
              // เปลี่ยนภาษา
              Positioned(
                top: 0,
                right: 0,
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    padding: const EdgeInsets.all(12),
                    foregroundColor: Colors.black,
                    backgroundColor: Colors.white,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(8),
                      side: BorderSide(color: Colors.grey.shade400, width: 1),
                    ),
                  ),
                  onPressed: () async {
                    final value = await showLanguageSelectionDialog(context);
                    if (value != null) {
                      global.userLanguage = value;
                      global.appConfig.setString(
                        'language',
                        global.userLanguage,
                      );
                      await global.languageSelect(global.userLanguage);
                      setState(() {});
                    }
                  },
                  child: Row(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      Image(
                        image: AssetImage(
                          "assets/flags/${global.userLanguage}.png",
                        ),
                        width: 30,
                        height: 30,
                      ),
                      const SizedBox(width: 8),
                      Text(
                        _getLanguageText(),
                        style: const TextStyle(fontSize: 14),
                      ),
                    ],
                  ),
                ),
              ),

              Column(
                mainAxisAlignment: MainAxisAlignment.center,
                mainAxisSize: MainAxisSize.min,
                children: [
                  imageLogo(),
                  if (global.userLoginData.token.isEmpty)
                    const SizedBox(height: 20),
                  if (global.userLoginData.token.isEmpty)
                    ElevatedButton(
                      style: ElevatedButton.styleFrom(
                        foregroundColor: Colors.black,
                        backgroundColor: Colors.white,
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(10),
                        ),
                      ),
                      child: Container(
                        padding: const EdgeInsets.all(10),
                        width: double.infinity,
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.center,
                          children: [
                            Image(
                              width: 45,
                              height: 45,
                              image: AssetImage("assets/img/google_logo.png"),
                            ),
                            SizedBox(width: 10),
                            Text(
                              'Sign in with google',
                              overflow: TextOverflow.ellipsis,
                              style: TextStyle(
                                color: Colors.black,
                                fontSize: 18,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ],
                        ),
                      ),
                      onPressed: () async {
                        if (kIsWeb) {
                          try {
                            final value = await googleSignInForWeb();
                            if (value != null) {
                              String? userIdToken =
                                  await getCurrentUserIdToken();
                              if (userIdToken != null) {
                                global.userLoginData = UserLoginModel(
                                  name: value.user!.displayName ?? "",
                                  code: "",
                                  token: userIdToken,
                                  email: value.user!.email ?? "",
                                  refreshtoken: userIdToken,
                                  photourl: value.user!.photoURL ?? "",
                                );

                                String stringUserLoginData = jsonEncode(
                                  global.userLoginData,
                                );

                                global.prefs.setString(
                                  "user",
                                  stringUserLoginData,
                                );
                              }
                            }
                            setState(() {});
                          } catch (e) {
                            if (kDebugMode) {
                              AppLogger.error("Google sign-in error: $e");
                            }
                          }
                        } else if (Platform.isWindows || Platform.isMacOS) {
                          // ใช้ GoogleAuthHelper สำหรับ Windows/MacOS
                          // รองรับ OAuth2 flow กับ localhost redirect
                          try {
                            final googleAuthHelper = GoogleAuthHelper();
                            final result = await googleAuthHelper.signIn();
                            if (result != null) {
                              String? userIdToken =
                                  await getCurrentUserIdToken();
                              if (userIdToken != null) {
                                global.userLoginData = UserLoginModel(
                                  name: result.user!.displayName ?? "",
                                  code: "",
                                  token: userIdToken,
                                  email: result.user!.email ?? "",
                                  refreshtoken: userIdToken,
                                  photourl: result.user!.photoURL ?? "",
                                );

                                String stringUserLoginData = jsonEncode(
                                  global.userLoginData,
                                );

                                global.prefs.setString(
                                  "user",
                                  stringUserLoginData,
                                );
                              }
                            }
                            setState(() {});
                          } catch (e) {
                            if (kDebugMode) {
                              AppLogger.error("Google sign-in error: $e");
                            }
                          }
                        }
                      },
                    ),
                  const SizedBox(height: 20),
                  userLoginWidget,
                  const SizedBox(height: 20),
                  Container(
                    width: double.infinity,
                    padding: const EdgeInsets.all(10),
                    decoration: BoxDecoration(
                      color: global.theme.cardColor,
                      border: Border.all(color: Colors.white),
                      borderRadius: BorderRadius.circular(10),
                      boxShadow: const [
                        BoxShadow(
                          color: Colors.grey,
                          blurRadius: 10.0,
                          spreadRadius: 0.0,
                          offset: Offset(2.0, 2.0),
                        ),
                      ],
                    ),
                    child: Column(
                      children: [
                        RichText(
                          text: TextSpan(
                            style: const TextStyle(
                              decoration: TextDecoration.none,
                            ),
                            children: <TextSpan>[
                              TextSpan(
                                text: global.language('you_have_accepted'),
                                style: TextStyle(
                                  color: Colors.black,
                                  fontSize: 14,
                                ),
                              ),
                              TextSpan(
                                text: ' ${global.language('terms_of_use')} ',
                                style: const TextStyle(
                                  color: Color(0xFF0A6ED1), // SAP Action Blue
                                  fontWeight: FontWeight.bold,
                                  fontSize: 14,
                                ),
                                recognizer: TapGestureRecognizer()
                                  ..onTap = () {
                                    // ignore: deprecated_member_use
                                    launch('https://www.smlsoft.com/terms');
                                  },
                              ),
                            ],
                          ),
                        ),
                        RichText(
                          text: TextSpan(
                            style: const TextStyle(
                              decoration: TextDecoration.none,
                            ),
                            children: <TextSpan>[
                              TextSpan(
                                text: ' ${global.language('read')} ',
                                style: TextStyle(
                                  color: Colors.black,
                                  fontSize: 14,
                                ),
                              ),
                              TextSpan(
                                text: ' ${global.language('privacy_policy')} ',
                                style: const TextStyle(
                                  color: Color(0xFF0A6ED1), // SAP Action Blue
                                  fontSize: 14,
                                  fontWeight: FontWeight.bold,
                                ),
                                recognizer: TapGestureRecognizer()
                                  ..onTap = () {
                                    // ignore: deprecated_member_use
                                    launch('https://www.smlsoft.com/privacy');
                                  },
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),

                  /// detail version
                  Padding(
                    padding: const EdgeInsets.all(8.0),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        // แสดง Version และ Flavor
                        Text(
                          'VERSION : ${F.title}',
                          style: TextStyle(
                            color: Colors.black,
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        const SizedBox(height: 4),

                        // แสดง Environment ที่ใช้งานจริงตาม Flavor
                        Text(
                          'ENVIRONMENT: ${_getEnvironmentName()}',
                          style: TextStyle(
                            color: Colors.grey[700],
                            fontSize: 12,
                          ),
                        ),

                        // แสดง Service API URL (ใช้ myAppConfig ที่ถูกต้อง)
                        Text(
                          'SERVICE API: ${global.myAppConfig.servicePort.isNotEmpty ? "${global.myAppConfig.serviceApi}:${global.myAppConfig.servicePort}" : global.myAppConfig.serviceApi}',
                          style: TextStyle(
                            color: Colors.grey[700],
                            fontSize: 12,
                          ),
                          overflow: TextOverflow.ellipsis,
                        ),

                        // แสดง Report API URL (ใช้ myAppConfig ที่ถูกต้อง)
                        Text(
                          'REPORT API: ${global.myAppConfig.reportApiPort.isNotEmpty ? "${global.myAppConfig.reportApiPath}:${global.myAppConfig.reportApiPort}" : global.myAppConfig.reportApiPath}',
                          style: TextStyle(
                            color: Colors.grey[700],
                            fontSize: 12,
                          ),
                          overflow: TextOverflow.ellipsis,
                        ),

                        // แสดง WebSocket Cart Service URL
                        Text(
                          'WEBSOCKET: ${Environment().config.webSocketCartService}',
                          style: TextStyle(
                            color: Colors.grey[700],
                            fontSize: 12,
                          ),
                          overflow: TextOverflow.ellipsis,
                        ),

                        const SizedBox(height: 8),

                        // ปุ่ม ธีม + Setup Config
                        Center(
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              // ปุ่มเปลี่ยนธีม
                              TextButton.icon(
                                onPressed: () async {
                                  // สลับ light ↔ dark
                                  global.displayThemeMode =
                                      global.isDarkMode() ? 'light' : 'dark';
                                  await global.saveThemeSettings();
                                  setState(() {});
                                },
                                icon: Icon(
                                  _getThemeModeIcon(),
                                  size: 16,
                                  color: _getThemeModeIconColor(),
                                ),
                                label: Text(
                                  _getThemeModeLabel(),
                                  style: TextStyle(
                                    color: global.theme.iconSecondaryColor,
                                    fontSize: 12,
                                  ),
                                ),
                              ),
                              const SizedBox(width: 8),
                              // ปุ่ม Setup Config
                              TextButton.icon(
                                onPressed: () {
                                  Navigator.push(
                                    context,
                                    MaterialPageRoute(
                                      builder: (_) => const SetupScreen(),
                                    ),
                                  );
                                },
                                icon: Icon(Icons.settings, size: 16, color: global.theme.iconSecondaryColor),
                                label: Text(
                                  'Setup',
                                  style: TextStyle(
                                    color: global.theme.iconSecondaryColor,
                                    fontSize: 12,
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  /// Widget แสดง Logo และชื่อ App ที่ใช้งานได้ทุก flavor
  /// โดยเปลี่ยนแค่ชื่อ App, Logo และสีธีมตาม flavor ที่รัน
  Widget imageLogo() {
    return Column(
      children: [
        // Logo with gradient background - ใช้สีธีมจาก F
        Container(
          padding: const EdgeInsets.all(20),
          decoration: BoxDecoration(
            gradient: LinearGradient(
              colors: [F.primaryGradientColor, F.secondaryGradientColor],
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
            ),
            borderRadius: BorderRadius.circular(20),
            boxShadow: [
              BoxShadow(
                color: F.gradientShadowColor.withValues(alpha: 0.4),
                blurRadius: 20,
                offset: const Offset(0, 10),
              ),
            ],
          ),
          child: ClipRRect(
            borderRadius: BorderRadius.circular(12),
            child: Image.asset(
              F.logoPath,
              width: 120,
              height: 120,
              fit: BoxFit.contain,
              errorBuilder: (context, error, stackTrace) {
                // Fallback icon ถ้าไม่มีรูป
                return Container(
                  width: 120,
                  height: 120,
                  decoration: BoxDecoration(
                    color: Colors.white.withValues(alpha: 0.2),
                    borderRadius: BorderRadius.circular(12),
                  ),
                  child: const Icon(
                    Icons.cloud,
                    size: 80,
                    color: Colors.white,
                  ),
                );
              },
            ),
          ),
        ),
        const SizedBox(height: 16),
        // ชื่อแอป - ใช้ F.appName พร้อมสีธีม
        ShaderMask(
          shaderCallback: (bounds) => LinearGradient(
            colors: [F.primaryGradientColor, F.secondaryGradientColor],
          ).createShader(bounds),
          child: Text(
            F.appName,
            style: const TextStyle(
              fontSize: 28,
              fontWeight: FontWeight.bold,
              color: Colors.white,
            ),
          ),
        ),
        const SizedBox(height: 8),
        // ชื่อบริษัท - ใช้ F.companyName
        if (F.companyName.isNotEmpty)
          Text(
            F.companyName,
            style: TextStyle(
              fontSize: 14,
              color: Colors.grey.shade600,
              fontWeight: FontWeight.w500,
            ),
          ),
        const SizedBox(height: 8),
        // Environment badge - แสดงเฉพาะ DEV และ UAT
        if (!F.isProd)
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
            decoration: BoxDecoration(
              color: _getEnvironmentColor().withValues(alpha: 0.1),
              borderRadius: BorderRadius.circular(12),
              border: Border.all(
                color: _getEnvironmentColor().withValues(alpha: 0.5),
              ),
            ),
            child: Text(
              _getEnvironmentBadgeText(),
              style: TextStyle(
                fontSize: 12,
                color: _getEnvironmentColor(),
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
      ],
    );
  }

  // ================== Theme Mode Helpers ==================

  IconData _getThemeModeIcon() {
    return global.isDarkMode() ? Icons.dark_mode : Icons.light_mode;
  }

  Color _getThemeModeIconColor() {
    return global.isDarkMode() ? Colors.indigo.shade300 : Colors.orange;
  }

  String _getThemeModeLabel() {
    return global.isDarkMode() ? 'Dark' : 'Light';
  }

  /// ดึงสี Environment badge ตาม flavor
  Color _getEnvironmentColor() {
    if (F.isDev) {
      return Colors.orange;
    } else if (F.isUat) {
      return Colors.blue;
    } else if (F.isProd) {
      return Colors.green;
    }
    return Colors.grey;
  }

  /// ดึงข้อความ Environment badge ตาม flavor
  String _getEnvironmentBadgeText() {
    if (F.isDev) {
      return '🔧 Development';
    } else if (F.isUat) {
      return '🧪 UAT';
    } else if (F.isProd) {
      return '🚀 Production';
    }
    return '';
  }
}
