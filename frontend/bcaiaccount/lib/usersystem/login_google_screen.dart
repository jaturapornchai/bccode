import 'dart:convert';

import 'package:firebase_auth/firebase_auth.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/login_bloc/login_bloc.dart';
import 'package:smlaicloud/environment.dart';
import 'package:smlaicloud/flavors.dart';
import 'package:smlaicloud/model/user_login_model.dart';
import 'package:smlaicloud/select_language_screen.dart';
import 'package:smlaicloud/utils/background.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:url_launcher/url_launcher.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// Login Screen with Google Sign-In (สำหรับ SMLAI flavors)
class LoginGoogleScreen extends StatefulWidget {
  const LoginGoogleScreen({super.key});

  @override
  State<LoginGoogleScreen> createState() => LoginGoogleScreenState();
}

class LoginGoogleScreenState extends State<LoginGoogleScreen> with global.ThemeRefreshMixin {
  late FirebaseAuth _auth;
  bool _isSigningIn = false;

  // Method สำหรับการตรวจสอบว่าได้ token ID หรือไม่
  Future<String?> getCurrentUserIdToken() async {
    try {
      User? user = FirebaseAuth.instance.currentUser;
      if (user != null) {
        String? idToken = await user.getIdToken();
        return idToken;
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error getting ID token: $e');
      }
    }
    return null;
  }

  @override
  void initState() {
    _auth = FirebaseAuth.instance;

    // ตรวจสอบว่ามี redirect result หรือไม่ (สำหรับกรณีใช้ signInWithRedirect)
    _checkForRedirectResult();

    try {
      // load user profile from local storage
      var userValue = global.prefs.getString("user");
      if (userValue != null && userValue.isNotEmpty) {
        // ตรวจสอบว่าเป็น JSON หรือไม่
        String trimmedValue = userValue.trim();
        if (trimmedValue.startsWith('{') && trimmedValue.endsWith('}')) {
          try {
            global.userLoginData = UserLoginModel.fromJson(
              jsonDecode(userValue),
            );
            // sync Google profile data ไปที่ global vars
            global.loginName = global.userLoginData.name;
            global.loginEmail = global.userLoginData.email;
            global.loginPhotoUrl = global.userLoginData.photourl;
            if (kDebugMode) {
              AppLogger.info('[LoginGoogle] User data loaded from storage');
            }
          } catch (jsonError) {
            if (kDebugMode) {
              AppLogger.debug(
                '[LoginGoogle] Invalid JSON in stored user data: $jsonError',
              );
              AppLogger.debug('[LoginGoogle] Stored data: $userValue');
            }
            // ลบข้อมูลที่เสียหาย
            global.prefs.remove("user");
          }
        } else {
          if (kDebugMode) {
            AppLogger.debug(
              '[LoginGoogle] Invalid format in stored user data: $userValue',
            );
          }
          // ลบข้อมูลที่ไม่ใช่ JSON
          global.prefs.remove("user");
        }
        setState(() {});
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('[LoginGoogle] Error loading user from storage: $e');
      }
      // ล้างข้อมูลที่อาจเสียหาย
      global.prefs.remove("user");
    }
    super.initState();
  }

  /// Get language selection text based on current language
  String _getLanguageText() {
    switch (global.userLanguage) {
      case 'en':
        return 'Select language English';
      case 'th':
        return global.language('select_language_thai');
      case 'lo':
        return 'ເລືອກພາສາ ລາວ';
      case 'cn':
        return '选择语言 中文';
      case 'ja':
        return '言語を選択 日本語';
      case 'ko':
        return '언어 선택 한국어';
      default:
        return 'Select language English';
    }
  }

  // ตรวจสอบผลลัพธ์จาก redirect
  Future<void> _checkForRedirectResult() async {
    try {
      final UserCredential result = await _auth.getRedirectResult();
      if (result.user != null) {
        if (kDebugMode) {
          AppLogger.debug('✅ Redirect result received: ${result.user?.email}');
        }
        await _handleSuccessfulSignIn(result.user!);
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Redirect result error: $e');
      }
    }
  }

  // แยกการ handle successful sign-in ออกมาเป็น method แยก
  Future<void> _handleSuccessfulSignIn(User user) async {
    try {
      String? userIdToken = await getCurrentUserIdToken();

      if (userIdToken != null) {
        if (kDebugMode) {
          AppLogger.debug('✅ ID Token obtained');
        }

        global.userLoginData = UserLoginModel(
          name: user.displayName ?? "",
          code: "",
          token: userIdToken,
          email: user.email ?? "",
          refreshtoken: userIdToken,
          photourl: user.photoURL ?? "",
        );

        // เก็บ Google profile data ไว้ใน global ด้วย
        global.loginName = user.displayName ?? '';
        global.loginEmail = user.email ?? '';
        global.loginPhotoUrl = user.photoURL ?? '';

        String stringUserLoginData = jsonEncode(global.userLoginData);
        global.prefs.setString("user", stringUserLoginData);

        if (kDebugMode) {
          AppLogger.debug('💾 User data saved to local storage');
        }

        if (mounted) setState(() {});
      } else {
        throw Exception('ไม่สามารถรับ ID Token ได้');
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Error handling successful sign-in: $e');
      }
      rethrow;
    }
  }

  @override
  Widget build(BuildContext context) {
    return BlocListener<LoginBloc, LoginState>(
      listener: (context, state) async {
        if (state is TokenLoginSuccess) {
          if (state.userLogin.token != '') {
            // อัพเดท token เป็น JWT จาก backend เก็บ name/email/photo จาก Firebase ไว้
            global.userLoginData = UserLoginModel(
              name: global.userLoginData.name,
              code: global.userLoginData.code,
              token: state.userLogin.token,
              email: global.userLoginData.email,
              refreshtoken: global.userLoginData.refreshtoken,
              photourl: global.userLoginData.photourl,
            );
            global.currentLoginMethod = global.LoginEnum.google;
            global.prefs.setString("user", jsonEncode(global.userLoginData));
            // โหลดภาษาจาก API ก่อน navigate (เหมือน password login)
            try {
              await global.languageSelect(global.userLanguage);
            } catch (e) {
              AppLogger.error('[GoogleLogin] โหลดภาษาไม่สำเร็จ: $e');
            }
            if (!context.mounted) return;
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
            CloudBackground(),
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
                      color: global.theme.iconSecondaryColor,
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
                            color: global.theme.dividerBorderColor,
                            child: Icon(
                              Icons.person,
                              size: 60,
                              color: global.theme.iconSecondaryColor,
                            ),
                          );
                        },
                        loadingBuilder: (context, child, loadingProgress) {
                          if (loadingProgress == null) return child;
                          return Container(
                            width: 100,
                            height: 100,
                            color: global.theme.dividerBorderColor,
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
                        color: global.theme.dividerBorderColor,
                        child: Icon(
                          Icons.person,
                          size: 60,
                          color: global.theme.iconSecondaryColor,
                        ),
                      ),
              ),
              const SizedBox(height: 10),
              Text(
                global.userLoginData.email,
                style: const TextStyle(
                  color: Colors.blue,
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
                    backgroundColor: Colors.green,
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
                  style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
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
            color: global.theme.dividerBorderColor,
            borderRadius: const BorderRadius.all(Radius.circular(10)),
            boxShadow: [
              BoxShadow(
                color: global.theme.iconSecondaryColor.withValues(alpha: 0.3),
                blurRadius: 10.0,
                spreadRadius: 2.0,
                offset: const Offset(-2.0, -2.0),
              ),
              BoxShadow(
                color: global.theme.iconSecondaryColor.withValues(alpha: 0.5),
                blurRadius: 10.0,
                spreadRadius: 2.0,
                offset: const Offset(2.0, 2.0),
              ),
              BoxShadow(
                color: global.theme.iconSecondaryColor.withValues(alpha: 0.3),
                blurRadius: 10.0,
                spreadRadius: 2.0,
                offset: const Offset(2.0, -2.0),
              ),
              BoxShadow(
                color: global.theme.iconSecondaryColor.withValues(alpha: 0.3),
                blurRadius: 10.0,
                spreadRadius: 2.0,
                offset: const Offset(-2.0, 2.0),
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
                    foregroundColor: global.theme.textColor,
                    backgroundColor: Colors.white,
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(8),
                      side: BorderSide(color: global.theme.iconSecondaryColor, width: 1),
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
                  const SizedBox(height: 20),
                  if (global.userLoginData.token.isEmpty)
                    _buildGoogleSignInButton(),
                  userLoginWidget,
                  const SizedBox(height: 20),
                  _buildTermsAndPrivacyPolicy(),

                  /// detail version
                  Padding(
                    padding: const EdgeInsets.all(8.0),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'VERSION : ${F.title}',
                          style: TextStyle(
                            color: global.theme.textColor,
                            fontSize: 18,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          'ENVIRONMENT: ${_getEnvironmentName()}',
                          style: TextStyle(
                            color: global.theme.textSecondaryColor,
                            fontSize: 12,
                          ),
                        ),
                        Text(
                          'SERVICE API: ${Environment().config.serviceApi}',
                          style: TextStyle(
                            color: global.theme.textSecondaryColor,
                            fontSize: 12,
                          ),
                          overflow: TextOverflow.ellipsis,
                        ),
                        Text(
                          'REPORT API: ${Environment().config.reportApi}',
                          style: TextStyle(
                            color: global.theme.textSecondaryColor,
                            fontSize: 12,
                          ),
                          overflow: TextOverflow.ellipsis,
                        ),
                        Text(
                          'WEBSOCKET: ${Environment().config.webSocketCartService}',
                          style: TextStyle(
                            color: global.theme.textSecondaryColor,
                            fontSize: 12,
                          ),
                          overflow: TextOverflow.ellipsis,
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
              color: global.theme.textSecondaryColor,
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

  /// ดึงสี Environment badge ตาม flavor
  Color _getEnvironmentColor() {
    if (F.isDev) {
      return Colors.orange;
    } else if (F.isUat) {
      return Colors.blue;
    } else if (F.isProd) {
      return Colors.green;
    }
    return global.theme.iconSecondaryColor;
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

  Widget _buildGoogleSignInButton() {
    return Column(
      children: [
        // สำหรับ Web - ใช้ custom button
        if (kIsWeb) ...[
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              foregroundColor: global.theme.textColor,
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
                  if (_isSigningIn) ...[
                    Image(
                      width: 45,
                      height: 45,
                      image: AssetImage("assets/img/google_logo.png"),
                    ),
                    SizedBox(width: 10),
                    Text(
                      'Signing in...',
                      style: TextStyle(
                        color: global.theme.textColor,
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    SizedBox(
                      width: 20,
                      height: 20,
                      child: CircularProgressIndicator(strokeWidth: 2),
                    ),
                  ] else ...[
                    Image(
                      width: 45,
                      height: 45,
                      image: AssetImage("assets/img/google_logo.png"),
                    ),
                    SizedBox(width: 10),
                    Text(
                      'Sign in with Google',
                      overflow: TextOverflow.ellipsis,
                      style: TextStyle(
                        color: global.theme.textColor,
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ],
                ],
              ),
            ),
            onPressed: () async {
              if (_isSigningIn) return; // ป้องกันการกดซ้ำ

              setState(() {
                _isSigningIn = true;
              });

              try {
                // เพิ่ม delay เล็กน้อยเพื่อให้ UI render
                await Future.delayed(Duration(milliseconds: 100));

                if (kDebugMode) {
                  AppLogger.debug('🚀 Starting Google Sign-In...');
                }

                // สำหรับ web ใช้ Firebase Auth popup แทน
                final GoogleAuthProvider googleProvider = GoogleAuthProvider();
                googleProvider.addScope('email');
                googleProvider.addScope('profile');
                // บังคับให้เลือก account ใหม่ทุกครั้ง
                googleProvider.setCustomParameters({
                  'prompt': 'select_account',
                  'access_type': 'online', // ใช้ online แทน offline สำหรับ web
                });

                if (kDebugMode) {
                  AppLogger.debug('📱 Opening Google Sign-In popup...');
                }

                // ใช้ try-catch เฉพาะสำหรับ popup
                UserCredential? userCredential;
                try {
                  userCredential = await _auth.signInWithPopup(googleProvider);
                } catch (popupError) {
                  if (kDebugMode) {
                    AppLogger.error('❌ Popup error: $popupError');
                  }

                  // ถ้า popup ล้มเหลว ลอง redirect แทน (สำหรับบางเบราว์เซอร์)
                  if (popupError.toString().contains('popup') ||
                      popupError.toString().contains('blocked')) {
                    if (kDebugMode) {
                      AppLogger.debug('🔄 Popup blocked, trying redirect...');
                    }
                    await _auth.signInWithRedirect(googleProvider);
                    return; // รอ redirect กลับมา
                  } else {
                    rethrow; // ถ้าเป็น error อื่นให้ throw ต่อ
                  }
                }

                if (userCredential.user != null) {
                  if (kDebugMode) {
                    AppLogger.debug(
                      '✅ Google Sign-In successful: ${userCredential.user!.email}',
                    );
                  }
                  await _handleSuccessfulSignIn(userCredential.user!);
                } else {
                  throw Exception('ไม่ได้รับข้อมูล user จาก Google');
                }
              } catch (e) {
                if (kDebugMode) {
                  AppLogger.error('❌ Google Sign-In Error: $e');
                }

                // แสดง error message ที่เหมาะสม
                String errorMessage = 'Google Sign-In Error';
                if (e.toString().contains('popup_blocked')) {
                  errorMessage =
                      'Popup ถูกบล็อก กรุณาเปิดใช้งาน popup ในเบราว์เซอร์';
                } else if (e.toString().contains('popup_closed')) {
                  errorMessage = 'ผู้ใช้ปิด popup ก่อนเสร็จสิ้นการเข้าสู่ระบบ';
                } else if (e.toString().contains('network')) {
                  errorMessage = 'ปัญหาการเชื่อมต่อเครือข่าย กรุณาลองใหม่';
                } else {
                  errorMessage = 'เกิดข้อผิดพลาด: ${e.toString()}';
                }

                if (mounted) {
                  global.showSnackBar(
                    context,
                    const Icon(Icons.error, color: Colors.white),
                    errorMessage,
                    Colors.red,
                  );
                }
              } finally {
                if (mounted) {
                  setState(() {
                    _isSigningIn = false;
                  });
                }
              }
            },
          ),
          const SizedBox(height: 10),
        ] else ...[
          // สำหรับ Non-web platforms แสดงข้อความเตือน
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: global.theme.infoHighlightColor,
              borderRadius: BorderRadius.circular(10),
              border: Border.all(color: Colors.orange),
            ),
            child: Column(
              children: [
                Icon(Icons.info, color: Colors.orange),
                SizedBox(height: 8),
                Text(
                  'Google Sign-In is only available on web platform',
                  style: TextStyle(
                    color: Colors.orange.shade800,
                    fontWeight: FontWeight.bold,
                  ),
                  textAlign: TextAlign.center,
                ),
              ],
            ),
          ),
          const SizedBox(height: 10),
        ],
      ],
    );
  }

  Widget _buildTermsAndPrivacyPolicy() {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        border: Border.all(color: global.theme.dividerBorderColor),
        borderRadius: BorderRadius.circular(10),
        boxShadow: [
          BoxShadow(
            color: global.theme.iconSecondaryColor,
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
              style: const TextStyle(decoration: TextDecoration.none),
              children: <TextSpan>[
                TextSpan(
                  text: global.language('you_have_accepted'),
                  style: TextStyle(color: global.theme.textColor, fontSize: 14),
                ),
                TextSpan(
                  text: ' ${global.language('terms_of_use')} ',
                  style: const TextStyle(
                    color: Colors.blue,
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
              style: const TextStyle(decoration: TextDecoration.none),
              children: <TextSpan>[
                TextSpan(
                  text: ' ${global.language('read')} ',
                  style: TextStyle(color: global.theme.textColor, fontSize: 14),
                ),
                TextSpan(
                  text: ' ${global.language('privacy_policy')} ',
                  style: const TextStyle(
                    color: Colors.blue,
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
    );
  }
}
