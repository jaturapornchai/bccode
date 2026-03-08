import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:http/http.dart' as http;

import 'package:firebase_auth/firebase_auth.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/gestures.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart' show Clipboard, ClipboardData;
import 'package:smlaicloud/utils/focus_utils.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/login_bloc/login_bloc.dart';
import 'package:smlaicloud/flavors.dart';
import 'package:smlaicloud/model/user_login_model.dart';
import 'package:smlaicloud/select_language_screen.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:url_launcher/url_launcher.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/usersystem/utils.dart';
import 'package:smlaicloud/screens/setup/setup_screen.dart';
import 'package:smlaicloud/utils/backend_url_manager.dart';
import 'package:qr_flutter/qr_flutter.dart';
import 'package:smlaicloud/widgets/line_login_qr_dialog.dart';
import 'package:smlaicloud/usersystem/register_username_screen.dart';

/// Login Screen using Username/Password and Google Sign-In (สำหรับ DoHome flavors)
class LoginPasswordScreen extends StatefulWidget {
  const LoginPasswordScreen({super.key});

  @override
  State<LoginPasswordScreen> createState() => LoginPasswordScreenState();
}

class LoginPasswordScreenState extends State<LoginPasswordScreen> {
  final TextEditingController _usernameController = TextEditingController();
  final TextEditingController _passwordController = TextEditingController();
  final TextEditingController _backendUrlController = TextEditingController();
  late final List<FocusNode> _loginFocusNodes;
  FocusNode get _usernameFocus => _loginFocusNodes[0];
  FocusNode get _passwordFocus => _loginFocusNodes[1];
  List<String> _urlHistory = [];
  // สถานะทดสอบเชื่อมต่อ Server: null=ยังไม่ทดสอบ, 'testing', 'success', 'failed'
  String? _connectionTestStatus;
  bool _isLoggingIn = false;
  bool _showPassword = false;
  bool _isSigningIn = false;
  bool _isLineLoggingIn = false;
  bool _isGoogleLoggingIn = false;
  bool _rememberPassword = false;
  late FirebaseAuth _auth;

  // แยกการ handle successful sign-in ออกมาเป็น method แยก
  Future<void> _handleSuccessfulSignIn(User user) async {
    try {
      if (kDebugMode) {
        AppLogger.debug('🔄 _handleSuccessfulSignIn started for user: ${user.email}');
        AppLogger.debug('🔄 Getting ID Token from Firebase...');
      }

      String? userIdToken = await getCurrentUserIdToken();

      if (kDebugMode) {
        AppLogger.debug('🔄 ID Token result: ${userIdToken != null ? "Got token (${userIdToken.length} chars)" : "null"}');
      }

      if (userIdToken != null) {
        if (kDebugMode) {
          AppLogger.debug('✅ ID Token obtained successfully');
        }

        global.userLoginData = UserLoginModel(name: user.displayName ?? "", code: "", token: userIdToken, email: user.email ?? "", refreshtoken: userIdToken, photourl: user.photoURL ?? "");

        // เก็บ profile data ไว้ใน global ด้วย
        global.loginName = user.displayName ?? '';
        global.loginEmail = user.email ?? '';
        global.loginPhotoUrl = user.photoURL ?? '';
        // ระบุ login method เป็น google ทันที ก่อน dispatch TokenLogin
        global.currentLoginMethod = global.LoginEnum.google;

        String stringUserLoginData = jsonEncode(global.userLoginData);
        global.prefs.setString("user", stringUserLoginData);

        // ล้าง saved credentials เดิม — login ด้วย social แล้วไม่ต้องจำ username/password เก่า
        await global.appConfig.remove('saved_username');
        await global.appConfig.remove('saved_password');
        await global.appConfig.setBool('remember_password', false);

        if (kDebugMode) {
          AppLogger.debug('💾 User data saved to local storage');
          AppLogger.debug('🧹 Cleared saved credentials (social login)');
          AppLogger.debug('🚀 Dispatching TokenLogin event to LoginBloc...');
        }

        // เรียก LoginBloc เพื่อไปหน้าถัดไป
        if (mounted) {
          context.read<LoginBloc>().add(TokenLogin(token: userIdToken));
          if (kDebugMode) {
            AppLogger.debug('✅ TokenLogin event dispatched!');
          }
        } else {
          if (kDebugMode) {
            AppLogger.error('❌ Widget is not mounted - cannot dispatch TokenLogin');
          }
        }
      } else {
        if (kDebugMode) {
          AppLogger.error('❌ ID Token is null - getCurrentUserIdToken() failed');
          AppLogger.error('❌ _auth.currentUser: ${_auth.currentUser}');
        }
        throw Exception('ไม่สามารถรับ ID Token ได้');
      }
    } catch (e, stackTrace) {
      if (kDebugMode) {
        AppLogger.error('❌ Error handling successful sign-in: $e');
        AppLogger.error('📍 Stack trace: $stackTrace');
      }
      rethrow;
    }
  }

  Future<String?> getCurrentUserIdToken() async {
    if (kDebugMode) {
      AppLogger.debug('🔍 getCurrentUserIdToken() called');
      AppLogger.debug('🔍 _auth instance: ${_auth.hashCode}');
    }

    User? currentUser = _auth.currentUser;

    if (kDebugMode) {
      AppLogger.debug('🔍 _auth.currentUser: ${currentUser != null ? currentUser.email : "null"}');
    }

    if (currentUser != null) {
      if (kDebugMode) {
        AppLogger.debug('🔍 Getting ID Token from currentUser...');
      }
      String? idToken = await currentUser.getIdToken();
      if (kDebugMode) {
        AppLogger.debug('🔍 ID Token obtained: ${idToken != null ? "${idToken.length} chars" : "null"}');
      }
      return idToken;
    } else {
      if (kDebugMode) {
        AppLogger.error('🔍 No currentUser in _auth - this should not happen after Firebase signIn');
      }
      return null;
    }
  }

  @override
  void initState() {
    super.initState();
    // Tab trap: วน focus เฉพาะ ชื่อผู้ใช้ ↔ รหัสผ่าน
    _loginFocusNodes = createTabCycleFocusNodes(2);
    if (kIsWeb) {
      _auth = FirebaseAuth.instance;
    } else if (Platform.isWindows || Platform.isMacOS) {
      _auth = FirebaseAuth.instance;
    }
    // โหลดรหัสล่าสุดจาก SharedPreferences (แยกตามเครื่อง)
    _loadSavedCredentials();
    // โหลด Backend URL จาก SharedPreferences
    _loadBackendUrl();
  }

  /// โหลดรหัสล่าสุดที่บันทึกไว้ (แยกตามเครื่องผ่าน SharedPreferences)
  void _loadSavedCredentials() {
    final savedUsername = global.appConfig.getString('saved_username') ?? '';
    final savedRemember = global.appConfig.getBool('remember_password') ?? false;
    AppLogger.info('[Login] โหลดรหัส — username: "$savedUsername", remember: $savedRemember');
    _usernameController.text = savedUsername;
    _rememberPassword = savedRemember;
    if (savedRemember) {
      final savedPassword = global.appConfig.getString('saved_password') ?? '';
      _passwordController.text = savedPassword;
      AppLogger.info('[Login] โหลดรหัสผ่านสำเร็จ (${savedPassword.length} ตัวอักษร)');
    }
  }

  /// บันทึกรหัสหลัง login สำเร็จ
  Future<void> _saveCredentials() async {
    final username = _usernameController.text.trim();
    await global.appConfig.setString('saved_username', username);
    await global.appConfig.setBool('remember_password', _rememberPassword);
    if (_rememberPassword) {
      await global.appConfig.setString('saved_password', _passwordController.text);
    } else {
      await global.appConfig.remove('saved_password');
    }
    AppLogger.info('[Login] บันทึกรหัสสำเร็จ — username: "$username", remember: $_rememberPassword');
  }

  @override
  void dispose() {
    _usernameController.dispose();
    _passwordController.dispose();
    _backendUrlController.dispose();
    disposeFocusNodes(_loginFocusNodes);
    _stopLineLoginPolling(); // หยุด polling timer ถ้ายังทำงานอยู่
    _stopGoogleLoginPolling(); // หยุด Google polling timer ถ้ายังทำงานอยู่
    super.dispose();
  }

  /// หยุด LINE polling timer
  void _stopLineLoginPolling() {
    _lineLoginPollingTimer?.cancel();
    _lineLoginPollingTimer = null;
  }

  /// หยุด Google polling timer
  void _stopGoogleLoginPolling() {
    _googleLoginPollingTimer?.cancel();
    _googleLoginPollingTimer = null;
  }

  /// โหลด Backend URL จาก SharedPreferences + แสดงใน textfield
  void _loadBackendUrl() {
    _backendUrlController.text = BackendUrlManager.getEffectiveUrl() ?? '';
    _urlHistory = BackendUrlManager.getHistory();
  }

  /// Multi-language text สำหรับหน้า Login — hardcode ไม่ต้อง load จาก backend
  /// เพราะตอนเปิด app ครั้งแรกอาจยังไม่มี URL ของ backend
  // 9 ภาษา: en, th, lo, zh, ja, ko, my, km, vi
  static const Map<String, Map<String, String>> _loginTexts = {
    'select_server': {
      'en': 'Select Server',
      'th': 'เลือกเซิร์ฟเวอร์',
      'lo': 'ເລືອກເຊີບເວີ',
      'zh': '选择服务器',
      'ja': 'サーバーを選択',
      'ko': '서버 선택',
      'my': 'ဆာဗာရွေးပါ',
      'km': 'ជ្រើសរើorg​សឺវើ',
      'vi': 'Chọn máy chủ',
    },
    'add_server': {'en': 'Add Server', 'th': 'เพิ่มเซิร์ฟเวอร์', 'lo': 'ເພີ່ມເຊີບເວີ', 'zh': '添加服务器', 'ja': 'サーバーを追加', 'ko': '서버 추가', 'my': 'ဆာဗာထည့်ပါ', 'km': 'បន្ថែមសឺវើ', 'vi': 'Thêm máy chủ'},
    'server_url': {'en': 'Server URL', 'th': 'URL เซิร์ฟเวอร์', 'lo': 'URL ເຊີບເວີ', 'zh': '服务器地址', 'ja': 'サーバーURL', 'ko': '서버 URL', 'my': 'ဆာဗာ URL', 'km': 'URL សឺវើ', 'vi': 'URL máy chủ'},
    'cancel': {'en': 'Cancel', 'th': 'ยกเลิก', 'lo': 'ຍົກເລີກ', 'zh': '取消', 'ja': 'キャンセル', 'ko': '취소', 'my': 'ပယ်ဖျက်', 'km': 'បោះបង់', 'vi': 'Hủy'},
    'save': {'en': 'Save', 'th': 'บันทึก', 'lo': 'ບັນທຶກ', 'zh': '保存', 'ja': '保存', 'ko': '저장', 'my': 'သိမ်းဆည်း', 'km': 'រក្សាទុក', 'vi': 'Lưu'},
    'welcome': {'en': 'Welcome', 'th': 'ยินดีต้อนรับ', 'lo': 'ຍິນດີຕ້ອນຮັບ', 'zh': '欢迎', 'ja': 'ようこそ', 'ko': '환영합니다', 'my': 'ကြိုဆိုပါသည်', 'km': 'សូមស្វាគមន៍', 'vi': 'Chào mừng'},
    'login': {'en': 'Login', 'th': 'เข้าสู่ระบบ', 'lo': 'ເຂົ້າສູ່ລະບົບ', 'zh': '登录', 'ja': 'ログイン', 'ko': '로그인', 'my': 'ဝင်ရောက်', 'km': 'ចូលប្រើ', 'vi': 'Đăng nhập'},
    'logout': {'en': 'Logout', 'th': 'ออกจากระบบ', 'lo': 'ອອກຈາກລະບົບ', 'zh': '退出', 'ja': 'ログアウト', 'ko': '로그아웃', 'my': 'ထွက်ရန်', 'km': 'ចាកចេញ', 'vi': 'Đăng xuất'},
    'username': {'en': 'Username', 'th': 'ชื่อผู้ใช้', 'lo': 'ຊື່ຜູ້ໃຊ້', 'zh': '用户名', 'ja': 'ユーザー名', 'ko': '사용자명', 'my': 'အသုံးပြုသူအမည်', 'km': 'ឈ្មោះអ្នកប្រើ', 'vi': 'Tên đăng nhập'},
    'password': {'en': 'Password', 'th': 'รหัสผ่าน', 'lo': 'ລະຫັດຜ່ານ', 'zh': '密码', 'ja': 'パスワード', 'ko': '비밀번호', 'my': 'စကားဝှက်', 'km': 'ពាក្យសម្ងាត់', 'vi': 'Mật khẩu'},
    'please_enter_username': {
      'en': 'Please enter username',
      'th': 'กรุณากรอกชื่อผู้ใช้',
      'lo': 'ກະລຸນາປ້ອນຊື່ຜູ້ໃຊ້',
      'zh': '请输入用户名',
      'ja': 'ユーザー名を入力してください',
      'ko': '사용자명을 입력하세요',
      'my': 'အသုံးပြုသူအမည်ထည့်ပါ',
      'km': 'សូមបញ្ចូលឈ្មោះអ្នកប្រើ',
      'vi': 'Vui lòng nhập tên đăng nhập',
    },
    'please_enter_password': {
      'en': 'Please enter password',
      'th': 'กรุณากรอกรหัสผ่าน',
      'lo': 'ກະລຸນາປ້ອນລະຫັດຜ່ານ',
      'zh': '请输入密码',
      'ja': 'パスワードを入力してください',
      'ko': '비밀번호를 입력하세요',
      'my': 'စကားဝှက်ထည့်ပါ',
      'km': 'សូមបញ្ចូលពាក្យសម្ងាត់',
      'vi': 'Vui lòng nhập mật khẩu',
    },
    'you_have_accepted': {
      'en': 'You have accepted',
      'th': 'คุณยอมรับ',
      'lo': 'ທ່ານຍອมຮັບ',
      'zh': '您已接受',
      'ja': '同意済み',
      'ko': '동의하셨습니다',
      'my': 'သင်လက်ခံပြီး',
      'km': 'អ្នកបានទទួលយក',
      'vi': 'Bạn đã chấp nhận',
    },
    'terms_of_use': {
      'en': 'Terms of Use',
      'th': 'เงื่อนไขการใช้งาน',
      'lo': 'ເງື່ອນໄຂການນໍາໃຊ້',
      'zh': '使用条款',
      'ja': '利用規約',
      'ko': '이용약관',
      'my': 'အသုံးပြုမှုစည်းမျဉ်း',
      'km': 'លក្ខខណ្ឌប្រើប្រាស់',
      'vi': 'Điều khoản sử dụng',
    },
    'read': {'en': 'Read', 'th': 'อ่าน', 'lo': 'ອ່ານ', 'zh': '阅读', 'ja': '読む', 'ko': '읽기', 'my': 'ဖတ်ရန်', 'km': 'អាន', 'vi': 'Đọc'},
    'privacy_policy': {
      'en': 'Privacy Policy',
      'th': 'นโยบายความเป็นส่วนตัว',
      'lo': 'ນະໂຍບາຍຄວາມເປັນສ່ວນຕົວ',
      'zh': '隐私政策',
      'ja': 'プライバシーポリシー',
      'ko': '개인정보처리방침',
      'my': 'ကိုယ်ရေးအချက်အလက်မူဝါဒ',
      'km': 'គោលការណ៍ភាពឯកជន',
      'vi': 'Chính sách bảo mật',
    },
    'test_connection': {
      'en': 'Test Connection',
      'th': 'ทดสอบเชื่อมต่อ',
      'lo': 'ທົດສອບເຊື່ອມຕໍ່',
      'zh': '测试连接',
      'ja': '接続テスト',
      'ko': '연결 테스트',
      'my': 'ချိတ်ဆက်မှုစမ်းသပ်',
      'km': 'សាកល្បងការតភ្ជាប់',
      'vi': 'Kiểm tra kết nối',
    },
    'testing': {
      'en': 'Testing...',
      'th': 'กำลังทดสอบ...',
      'lo': 'ກຳລັງທົດສອບ...',
      'zh': '测试中...',
      'ja': 'テスト中...',
      'ko': '테스트 중...',
      'my': 'စမ်းသပ်နေသည်...',
      'km': 'កំពុងសាកល្បង...',
      'vi': 'Đang kiểm tra...',
    },
    'connected': {'en': 'Connected', 'th': 'เชื่อมต่อสำเร็จ', 'lo': 'ເຊື່ອມຕໍ່ສຳເລັດ', 'zh': '连接成功', 'ja': '接続成功', 'ko': '연결 성공', 'my': 'ချိတ်ဆက်ပြီး', 'km': 'បានតភ្ជាប់', 'vi': 'Kết nối thành công'},
    'connection_failed': {
      'en': 'Connection failed',
      'th': 'เชื่อมต่อไม่สำเร็จ',
      'lo': 'ເຊື່ອມຕໍ່ບໍ່ສຳເລັດ',
      'zh': '连接失败',
      'ja': '接続失敗',
      'ko': '연결 실패',
      'my': 'ချိတ်ဆက်မှုမအောင်မြင်',
      'km': 'ការតភ្ជាប់បរាជ័យ',
      'vi': 'Kết nối thất bại',
    },
    'remember_password': {
      'en': 'Remember password',
      'th': 'บันทึกรหัสผ่าน',
      'lo': 'ບັນທຶກລະຫັດຜ່ານ',
      'zh': '记住密码',
      'ja': 'パスワードを記憶',
      'ko': '비밀번호 저장',
      'my': 'စကားဝှက်မှတ်ထား',
      'km': 'ចងចាំពាក្យសម្ងាត់',
      'vi': 'Nhớ mật khẩu',
    },
    'register': {'en': 'Register', 'th': 'สมัครผู้ใช้งาน', 'lo': 'ລົງທະບຽນ', 'zh': '注册', 'ja': '登録', 'ko': '회원가입', 'my': 'စာရင်းသွင်း', 'km': 'ចុះឈ្មោះ', 'vi': 'Đăng ký'},
    'manual': {'en': 'Manual', 'th': 'คู่มือ', 'lo': 'ຄູ່ມື', 'zh': '使用手册', 'ja': 'マニュアル', 'ko': '매뉴얼', 'my': 'လမ်းညွှန်', 'km': 'សៀវភៅណែនាំ', 'vi': 'Hướng dẫn'},
    'setup': {'en': 'Setup', 'th': 'ตั้งค่า', 'lo': 'ຕັ້ງຄ່າ', 'zh': '设置', 'ja': '設定', 'ko': '설정', 'my': 'ပြင်ဆင်', 'km': 'ការកំណត់', 'vi': 'Cài đặt'},
  };

  String _loginText(String key) {
    final lang = global.userLanguage; // en, th, lo, zh, ja, ko, my, km, vi
    return _loginTexts[key]?[lang] ?? _loginTexts[key]?['en'] ?? key;
  }

  /// ทดสอบเชื่อมต่อ Server — ลอง HTTP GET ไปที่ backend URL
  Future<void> _testServerConnection() async {
    final url = _backendUrlController.text.trim();
    if (url.isEmpty) return;

    setState(() => _connectionTestStatus = 'testing');

    try {
      final uri = Uri.parse(url.endsWith('/') ? url : '$url/');
      final response = await http.get(uri).timeout(const Duration(seconds: 5));
      if (!mounted) return;
      setState(() {
        _connectionTestStatus = (response.statusCode < 500) ? 'success' : 'failed';
      });
    } catch (e) {
      if (!mounted) return;
      AppLogger.error('[Login] ทดสอบเชื่อมต่อ Server ล้มเหลว: $e');
      setState(() => _connectionTestStatus = 'failed');
    }
  }

  /// Navigate ไปหน้าเลือก shop — โหลดภาษาจาก API ก่อน ถ้าล้มเหลวแสดง error
  Future<void> _navigateToShopScreen() async {
    try {
      await global.languageSelect(global.userLanguage);
    } catch (e) {
      AppLogger.error('[Login] โหลดภาษาจาก API ไม่สำเร็จ: $e');
      if (mounted) {
        global.showSnackBar(context, const Icon(Icons.error, color: Colors.white), 'ไม่สามารถเชื่อมต่อ Server ได้ กรุณาตรวจสอบ Backend URL', Colors.red);
      }
      return;
    }
    if (!mounted) return;
    Navigator.pushNamedAndRemoveUntil(context, '/login_screen_shop', (route) => false);
  }

  /// บันทึก Backend URL ที่ผู้ใช้กรอก + reload app config
  Future<void> _saveBackendUrl() async {
    final url = _backendUrlController.text.trim();
    if (url.isEmpty) return;

    await BackendUrlManager.saveUrl(url);
    setState(() {
      _urlHistory = BackendUrlManager.getHistory();
    });
    AppLogger.info('[Login] บันทึก Backend URL: $url');
  }

  /// Get Line Login text based on current language
  /// แสดงข้อความปุ่ม Line Login ตามภาษาปัจจุบัน
  String _getLineLoginText() {
    switch (global.userLanguage) {
      case 'en':
        return 'Login with LINE';
      case 'th':
        return 'เข้าสู่ระบบด้วย LINE';
      case 'lo':
        return 'ເຂົ້າສູ່ລະບົບດ້ວຍ LINE';
      case 'cn':
        return '使用LINE登录';
      case 'ja':
        return 'LINEでログイン';
      case 'ko':
        return 'LINE으로 로그인';
      case 'my':
        return 'LINE ဖြင့်ဝင်ရောက်ပါ';
      case 'km':
        return 'ចូលដោយ LINE';
      case 'vi':
        return 'Đăng nhập bằng LINE';
      default:
        return 'Login with LINE';
    }
  }

  /// Get language selection text based on current language
  String _getLanguageText() {
    switch (global.userLanguage) {
      case 'en':
        return 'Select language English';
      case 'th':
        return 'เลือกภาษา ไทย';
      case 'lo':
        return 'ເລືອກພາສາ ລາວ';
      case 'cn':
        return '选择语言 中文';
      case 'ja':
        return '言語を選択 日本語';
      case 'ko':
        return '언어 선택 한국어';
      case 'my':
        return 'ဘာသာစကား မြန်မာ';
      case 'km':
        return 'ជ្រើសរើសភាសា ខ្មែរ';
      case 'vi':
        return 'Chọn ngôn ngữ Tiếng Việt';
      default:
        return 'Select language English';
    }
  }

  @override
  Widget build(BuildContext context) {
    final size = MediaQuery.of(context).size;
    final isSmallScreen = size.width < 600;

    return BlocListener<LoginBloc, LoginState>(
      listener: (context, state) {
        // จัดการกับ state ที่เกิดจากการ login ด้วย token
        if (state is TokenLoginSuccess) {
          if (state.userLogin.token != '') {
            // อัพเดท token เป็น JWT จาก backend เก็บ name/email/photo จาก Google ไว้
            global.userLoginData = UserLoginModel(
              name: global.userLoginData.name,
              code: global.userLoginData.code,
              token: state.userLogin.token,
              email: global.userLoginData.email,
              refreshtoken: global.userLoginData.refreshtoken,
              photourl: global.userLoginData.photourl,
            );
            global.prefs.setString("user", jsonEncode(global.userLoginData));
            _navigateToShopScreen();
          }
        } else if (state is TokenLoginFailed) {
          global.showSnackBar(context, const Icon(Icons.error, color: Colors.white), state.message, Colors.red);
          removeUser();
        }
        // จัดการกับ state ที่เกิดจากการ login ด้วย username/password
        else if (state is LoginSuccess) {
          AppLogger.info('[Login] LoginSuccess — กำลังบันทึกรหัส...');
          setState(() {
            _isLoggingIn = false;
          });
          global.userLoginData = state.userLogin;
          global.currentLoginMethod = global.LoginEnum.password;
          // ล้าง Google/LINE profile data — login ด้วย password แล้วไม่ต้องแสดง social profile
          global.loginName = '';
          global.loginEmail = '';
          global.loginPhotoUrl = '';
          _saveCredentials().then((_) {
            if (mounted) _navigateToShopScreen();
          });
        } else if (state is LoginFailed) {
          setState(() {
            _isLoggingIn = false;
          });

          global.showSnackBar(context, const Icon(Icons.error, color: Colors.white), state.message, Colors.red);
        }
        // จัดการกับ state ที่เกิดจากการ login ด้วย LINE
        else if (state is LineLoginSuccess) {
          setState(() {
            _isLineLoggingIn = false;
          });
          global.userLoginData = state.userLogin;
          // บันทึกข้อมูล user
          String stringUserLoginData = jsonEncode(global.userLoginData);
          global.prefs.setString("user", stringUserLoginData);
          _navigateToShopScreen();
        } else if (state is LineLoginFailed) {
          setState(() {
            _isLineLoggingIn = false;
          });
          global.showSnackBar(context, const Icon(Icons.error, color: Colors.white), state.message, Colors.red);
        } else if (state is LineLoginInProgress) {
          setState(() {
            _isLineLoggingIn = true;
          });
        }
        // จัดการกับ state ที่เกิดจากการ login ด้วย Google
        else if (state is GoogleLoginSuccess) {
          setState(() {
            _isGoogleLoggingIn = false;
          });
          global.userLoginData = state.userLogin;
          global.currentLoginMethod = global.LoginEnum.google;
          // บันทึกข้อมูล user
          String stringUserLoginData = jsonEncode(global.userLoginData);
          global.prefs.setString("user", stringUserLoginData);
          _navigateToShopScreen();
        } else if (state is GoogleLoginFailed) {
          setState(() {
            _isGoogleLoggingIn = false;
          });
          global.showSnackBar(context, const Icon(Icons.error, color: Colors.white), state.message, Colors.red);
        } else if (state is GoogleLoginInProgress) {
          setState(() {
            _isGoogleLoggingIn = true;
          });
        }
      },
      child: Scaffold(
        body: Container(
          width: double.infinity,
          height: double.infinity,
          decoration: BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
              colors: [
                _darkenColor(F.primaryGradientColor, 0.3), // สีเข้มขึ้น
                F.primaryGradientColor,
                F.secondaryGradientColor,
              ],
              stops: const [0.0, 0.5, 1.0],
            ),
          ),
          child: SafeArea(
            child: Stack(
              children: [
                // Decorative circles
                Positioned(
                  top: -100,
                  right: -100,
                  child: Container(
                    width: 300,
                    height: 300,
                    decoration: BoxDecoration(shape: BoxShape.circle, color: Colors.white.withValues(alpha: 0.05)),
                  ),
                ),
                Positioned(
                  bottom: -150,
                  left: -100,
                  child: Container(
                    width: 400,
                    height: 400,
                    decoration: BoxDecoration(shape: BoxShape.circle, color: Colors.white.withValues(alpha: 0.05)),
                  ),
                ),
                Positioned(
                  top: size.height * 0.3,
                  left: -50,
                  child: Container(
                    width: 150,
                    height: 150,
                    decoration: BoxDecoration(shape: BoxShape.circle, color: Colors.white.withValues(alpha: 0.03)),
                  ),
                ),
                // Main content
                Center(
                  child: SingleChildScrollView(
                    padding: EdgeInsets.symmetric(horizontal: isSmallScreen ? 24 : 40, vertical: 20),
                    child: buildLoginForm(),
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  void removeUser() {
    global.prefs.remove("user");
    global.userLoginData = UserLoginModel(name: "", code: "", token: "", email: "", refreshtoken: "", photourl: "");
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

  /// ทำให้สีเข้มขึ้น (darken)
  Color _darkenColor(Color color, double factor) {
    final int a = (color.a * 255.0).round().clamp(0, 255);
    final int r = ((color.r * 255.0) * (1 - factor)).round().clamp(0, 255);
    final int g = ((color.g * 255.0) * (1 - factor)).round().clamp(0, 255);
    final int b = ((color.b * 255.0) * (1 - factor)).round().clamp(0, 255);
    return Color.fromARGB(a, r, g, b);
  }

  Widget buildLoginForm() {
    Widget userLoginWidget = (global.userLoginData.token.isNotEmpty) ? _buildLoggedInUserCard() : Container();

    // ตรวจสอบความสูงหน้าจอเพื่อปรับ UI ให้พอดี
    final screenHeight = MediaQuery.of(context).size.height;
    final isCompactScreen = screenHeight < 800; // จอเล็กกว่า 800px
    final cardPadding = isCompactScreen ? 20.0 : 32.0;
    final sectionSpacing = isCompactScreen ? 16.0 : 32.0;
    final smallSpacing = isCompactScreen ? 12.0 : 24.0;

    return ConstrainedBox(
      constraints: const BoxConstraints(maxWidth: 420),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          // Logo section
          _buildLogoSection(isCompact: isCompactScreen),
          SizedBox(height: isCompactScreen ? 8 : 16),

          // Backend URL selector — ซ่อนบน Web (URL มาจาก config.json)
          if (!kIsWeb) ...[_buildBackendUrlSelector(), SizedBox(height: isCompactScreen ? 8 : 16)],

          // Main login card
          Container(
            padding: EdgeInsets.all(cardPadding),
            decoration: BoxDecoration(
              color: Colors.white,
              borderRadius: BorderRadius.circular(24),
              boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.15), blurRadius: 30, offset: const Offset(0, 15))],
            ),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Welcome text
                if (global.userLoginData.token.isEmpty) ...[
                  Text(
                    _loginText("welcome"),
                    style: TextStyle(fontSize: isCompactScreen ? 22 : 28, fontWeight: FontWeight.bold, color: const Color(0xFF1A237E)),
                  ),
                  SizedBox(height: isCompactScreen ? 4 : 8),
                  Text(
                    _loginText("login"),
                    style: TextStyle(fontSize: isCompactScreen ? 14 : 16, color: Colors.grey[600]),
                  ),
                  SizedBox(height: sectionSpacing),
                  _buildUsernamePasswordLogin(isCompact: isCompactScreen),
                ],
                userLoginWidget,
              ],
            ),
          ),

          SizedBox(height: smallSpacing),

          // Terms and Privacy
          _buildTermsAndPrivacyPolicy(isCompact: isCompactScreen),

          SizedBox(height: isCompactScreen ? 4 : 8),

          // ปุ่ม คู่มือ + Setup Config
          Center(
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextButton.icon(
                  onPressed: () {
                    launchUrl(Uri.parse('https://bcaicloud.com/manual/loginscreen'), mode: LaunchMode.externalApplication);
                  },
                  icon: Icon(Icons.menu_book_rounded, size: 14, color: Colors.white.withValues(alpha: 0.6)),
                  label: Text(_loginText('manual'), style: TextStyle(color: Colors.white.withValues(alpha: 0.6), fontSize: 12)),
                ),
                const SizedBox(width: 8),
                TextButton.icon(
                  onPressed: () {
                    Navigator.push(context, MaterialPageRoute(builder: (_) => const SetupScreen()));
                  },
                  icon: Icon(Icons.settings, size: 14, color: Colors.white.withValues(alpha: 0.6)),
                  label: Text(_loginText('setup'), style: TextStyle(color: Colors.white.withValues(alpha: 0.6), fontSize: 12)),
                ),
              ],
            ),
          ),

          SizedBox(height: isCompactScreen ? 4 : 8),
        ],
      ),
    );
  }

  Widget _buildLogoSection({bool isCompact = false}) {
    final logoSize = isCompact ? 60.0 : 80.0;
    final logoPadding = isCompact ? 14.0 : 20.0;
    final appNameSize = isCompact ? 20.0 : 24.0;

    return Column(
      children: [
        // Logo with gradient background - ใช้ design เดียวกันทุก flavor
        Container(
          padding: EdgeInsets.all(logoPadding),
          decoration: BoxDecoration(
            color: Colors.white.withValues(alpha: 0.15),
            borderRadius: BorderRadius.circular(20),
            boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.1), blurRadius: 20, offset: const Offset(0, 10))],
          ),
          child: ClipRRect(
            borderRadius: BorderRadius.circular(12),
            child: Image.asset(
              F.logoPath,
              width: logoSize,
              height: logoSize,
              fit: BoxFit.contain,
              errorBuilder: (context, error, stackTrace) {
                // Fallback icon ถ้าไม่มีรูป
                return Icon(Icons.cloud_outlined, size: logoSize * 0.75, color: Colors.white);
              },
            ),
          ),
        ),
        SizedBox(height: isCompact ? 10 : 16),
        // ชื่อ App
        Text(
          F.appName,
          style: TextStyle(fontSize: appNameSize, fontWeight: FontWeight.bold, color: Colors.white, letterSpacing: 1.2),
        ),
        // Environment badge - แสดงเฉพาะ DEV และ UAT
        if (!F.isProd) ...[
          SizedBox(height: isCompact ? 4 : 8),
          Container(
            padding: EdgeInsets.symmetric(horizontal: isCompact ? 10 : 12, vertical: isCompact ? 2 : 4),
            decoration: BoxDecoration(color: Colors.white.withValues(alpha: 0.2), borderRadius: BorderRadius.circular(12)),
            child: Text(
              _getEnvironmentBadgeText(),
              style: TextStyle(fontSize: isCompact ? 10 : 12, color: Colors.white, fontWeight: FontWeight.bold),
            ),
          ),
        ],
      ],
    );
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

  Widget _buildLoggedInUserCard() {
    return Column(
      children: [
        // User avatar with glow effect
        Container(
          padding: const EdgeInsets.all(4),
          decoration: BoxDecoration(
            shape: BoxShape.circle,
            gradient: LinearGradient(colors: [const Color(0xFF3949AB), const Color(0xFF5C6BC0)]),
            boxShadow: [BoxShadow(color: const Color(0xFF3949AB).withValues(alpha: 0.4), blurRadius: 20, spreadRadius: 2)],
          ),
          child: CircleAvatar(
            radius: 50,
            backgroundColor: Colors.white,
            child: global.userLoginData.photourl.isNotEmpty
                ? ClipOval(
                    child: Image.network(
                      global.userLoginData.photourl,
                      width: 92,
                      height: 92,
                      fit: BoxFit.cover,
                      errorBuilder: (context, error, stackTrace) {
                        return Icon(Icons.person, size: 50, color: const Color(0xFF3949AB));
                      },
                    ),
                  )
                : Icon(Icons.person, size: 50, color: const Color(0xFF3949AB)),
          ),
        ),
        const SizedBox(height: 20),

        // User info
        Text(
          global.userLoginData.name.isNotEmpty ? global.userLoginData.name : global.userLoginData.email,
          style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: const Color(0xFF1A237E)),
        ),
        if (global.userLoginData.email.isNotEmpty && global.userLoginData.name.isNotEmpty)
          Padding(
            padding: const EdgeInsets.only(top: 4),
            child: Text(global.userLoginData.email, style: TextStyle(fontSize: 14, color: Colors.grey[600])),
          ),
        const SizedBox(height: 8),
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 6),
          decoration: BoxDecoration(color: Colors.green.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(20)),
          child: Text(
            "${_loginText("welcome")}!",
            style: TextStyle(color: Colors.green[700], fontWeight: FontWeight.w500),
          ),
        ),
        const SizedBox(height: 28),

        // Continue button
        SizedBox(
          width: double.infinity,
          height: 54,
          child: ElevatedButton(
            focusNode: FocusNode(canRequestFocus: false),
            style: ElevatedButton.styleFrom(
              backgroundColor: const Color(0xFF3949AB),
              foregroundColor: Colors.white,
              elevation: 4,
              shadowColor: const Color(0xFF3949AB).withValues(alpha: 0.4),
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
            ),
            onPressed: () {
              if (mounted) {
                context.read<LoginBloc>().add(TokenLogin(token: global.userLoginData.token));
              }
            },
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(Icons.login_rounded, size: 22),
                const SizedBox(width: 10),
                Text(_loginText("login"), style: TextStyle(fontSize: 17, fontWeight: FontWeight.w600, letterSpacing: 0.5)),
              ],
            ),
          ),
        ),
        const SizedBox(height: 14),

        // Logout button
        SizedBox(
          width: double.infinity,
          height: 54,
          child: OutlinedButton(
            focusNode: FocusNode(canRequestFocus: false),
            style: OutlinedButton.styleFrom(
              foregroundColor: Colors.red[600],
              side: BorderSide(color: Colors.red.shade300, width: 1.5),
              shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
            ),
            onPressed: () {
              removeUser();
            },
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(Icons.logout_rounded, size: 22),
                const SizedBox(width: 10),
                Text(_loginText("logout"), style: TextStyle(fontSize: 17, fontWeight: FontWeight.w600)),
              ],
            ),
          ),
        ),
      ],
    );
  }

  /// Backend URL selector — dropdown เลือก server + เพิ่ม/ลบ URL
  Widget _buildBackendUrlSelector() {
    final currentUrl = _backendUrlController.text;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(color: Colors.white.withValues(alpha: 0.12), borderRadius: BorderRadius.circular(12)),
      child: Row(
        children: [
          Icon(Icons.dns_outlined, size: 16, color: Colors.white70),
          const SizedBox(width: 8),
          // Dropdown เลือก URL จากประวัติ
          Expanded(
            child: DropdownButtonHideUnderline(
              child: DropdownButton<String>(
                focusNode: FocusNode(canRequestFocus: false),
                value: _urlHistory.contains(currentUrl) ? currentUrl : null,
                hint: Text(
                  currentUrl.isNotEmpty ? currentUrl : _loginText('select_server'),
                  style: const TextStyle(color: Colors.white, fontSize: 13),
                  overflow: TextOverflow.ellipsis,
                ),
                dropdownColor: const Color(0xFF1A237E),
                iconEnabledColor: Colors.white70,
                iconSize: 20,
                isExpanded: true,
                style: const TextStyle(color: Colors.white, fontSize: 13),
                items: _urlHistory.map((url) {
                  return DropdownMenuItem<String>(
                    value: url,
                    child: Row(
                      children: [
                        Expanded(child: Text(url, overflow: TextOverflow.ellipsis)),
                        // ปุ่มลบ
                        InkWell(
                          focusNode: FocusNode(canRequestFocus: false),
                          onTap: () {
                            Navigator.pop(context);
                            BackendUrlManager.removeFromHistory(url);
                            setState(() {
                              _urlHistory = BackendUrlManager.getHistory();
                            });
                          },
                          child: Padding(
                            padding: const EdgeInsets.all(4),
                            child: Icon(Icons.close, size: 14, color: Colors.white38),
                          ),
                        ),
                      ],
                    ),
                  );
                }).toList(),
                onChanged: (url) {
                  if (url != null) {
                    _backendUrlController.text = url;
                    _saveBackendUrl();
                    setState(() => _connectionTestStatus = null);
                  }
                },
              ),
            ),
          ),
          const SizedBox(width: 4),
          // ปุ่ม copy URL
          InkWell(
            focusNode: FocusNode(canRequestFocus: false),
            onTap: () {
              final url = _backendUrlController.text;
              if (url.isNotEmpty) {
                Clipboard.setData(ClipboardData(text: url));
                global.showInfoSnackBar(context, 'คัดลอก URL แล้ว');
              }
            },
            borderRadius: BorderRadius.circular(6),
            child: Container(
              padding: const EdgeInsets.all(6),
              decoration: BoxDecoration(color: Colors.white.withValues(alpha: 0.15), borderRadius: BorderRadius.circular(6)),
              child: Icon(Icons.copy, size: 18, color: Colors.white70),
            ),
          ),
          const SizedBox(width: 4),
          // ปุ่มเพิ่ม URL ใหม่
          InkWell(
            focusNode: FocusNode(canRequestFocus: false),
            onTap: () => _showAddUrlDialog(),
            borderRadius: BorderRadius.circular(6),
            child: Container(
              padding: const EdgeInsets.all(6),
              decoration: BoxDecoration(color: Colors.white.withValues(alpha: 0.15), borderRadius: BorderRadius.circular(6)),
              child: Icon(Icons.add, size: 18, color: Colors.white70),
            ),
          ),
          const SizedBox(width: 4),
          // ปุ่มทดสอบเชื่อมต่อ Server
          InkWell(
            focusNode: FocusNode(canRequestFocus: false),
            onTap: _connectionTestStatus == 'testing' ? null : _testServerConnection,
            borderRadius: BorderRadius.circular(6),
            child: Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
              decoration: BoxDecoration(
                color: _connectionTestStatus == 'success'
                    ? Colors.green.withValues(alpha: 0.3)
                    : _connectionTestStatus == 'failed'
                    ? Colors.red.withValues(alpha: 0.3)
                    : Colors.white.withValues(alpha: 0.15),
                borderRadius: BorderRadius.circular(6),
              ),
              child: _connectionTestStatus == 'testing'
                  ? const SizedBox(width: 14, height: 14, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.white70))
                  : Icon(
                      _connectionTestStatus == 'success'
                          ? Icons.check_circle
                          : _connectionTestStatus == 'failed'
                          ? Icons.error
                          : Icons.wifi_tethering,
                      size: 14,
                      color: _connectionTestStatus == 'success'
                          ? Colors.greenAccent
                          : _connectionTestStatus == 'failed'
                          ? Colors.redAccent
                          : Colors.white70,
                    ),
            ),
          ),
        ],
      ),
    );
  }

  /// Dialog เพิ่ม URL ใหม่
  void _showAddUrlDialog() {
    final addController = TextEditingController();
    showDialog(
      context: context,
      builder: (ctx) => AlertDialog(
        title: Text(_loginText('add_server')),
        content: TextField(
          controller: addController,
          autofocus: true,
          decoration: InputDecoration(hintText: 'http://192.168.0.5:9090', labelText: _loginText('server_url')),
          onSubmitted: (_) {
            final url = addController.text.trim();
            if (url.isNotEmpty) {
              Navigator.pop(ctx);
              _backendUrlController.text = url;
              _saveBackendUrl();
              setState(() => _connectionTestStatus = null);
            }
          },
        ),
        actions: [
          TextButton(onPressed: () => Navigator.pop(ctx), child: Text(_loginText('cancel'))),
          FilledButton(
            onPressed: () {
              final url = addController.text.trim();
              if (url.isNotEmpty) {
                Navigator.pop(ctx);
                _backendUrlController.text = url;
                _saveBackendUrl();
                setState(() => _connectionTestStatus = null);
              }
            },
            child: Text(_loginText('save')),
          ),
        ],
      ),
    );
  }

  Widget imageLogo() {
    // For DoHome flavors, use DoHome logo
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(20),
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.1), blurRadius: 20, offset: const Offset(0, 10))],
      ),
      child: Image.asset("assets/icons/logo_marine_new_color_white.png", height: 80),
    );
  }

  Widget _buildUsernamePasswordLogin({bool isCompact = false}) {
    final fieldSpacing = isCompact ? 10.0 : 16.0;
    final sectionSpacing = isCompact ? 16.0 : 28.0;
    final buttonHeight = isCompact ? 46.0 : 56.0;
    final fontSize = isCompact ? 14.0 : 16.0;
    final iconSize = isCompact ? 18.0 : 22.0;
    final iconMargin = isCompact ? 8.0 : 12.0;
    final iconPadding = isCompact ? 6.0 : 8.0;
    final contentPaddingV = isCompact ? 12.0 : 18.0;

    return Column(
      children: [
        // Tab วนกลับ: ชื่อผู้ใช้ ↔ รหัสผ่าน
        FocusTraversalGroup(
          child: Column(
            children: [
        // Username field with modern styling
        Container(
          decoration: BoxDecoration(
            color: Colors.grey[50],
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: Colors.grey.shade200),
          ),
          child: TextField(
            controller: _usernameController,
            focusNode: _usernameFocus,
            autofocus: true,
            textInputAction: TextInputAction.next,
            onSubmitted: (_) => _passwordFocus.requestFocus(),
            style: TextStyle(fontSize: fontSize),
            decoration: InputDecoration(
              labelText: _loginText('username'),
              labelStyle: TextStyle(color: Colors.grey[600], fontSize: fontSize),
              border: InputBorder.none,
              prefixIcon: Container(
                margin: EdgeInsets.all(iconMargin),
                padding: EdgeInsets.all(iconPadding),
                decoration: BoxDecoration(color: const Color(0xFF3949AB).withValues(alpha: 0.1), borderRadius: BorderRadius.circular(10)),
                child: Icon(Icons.person_outline_rounded, color: const Color(0xFF3949AB), size: iconSize),
              ),
              contentPadding: EdgeInsets.symmetric(horizontal: 20, vertical: contentPaddingV),
            ),
            enabled: !_isLoggingIn,
          ),
        ),
        SizedBox(height: fieldSpacing),

        // Password field with modern styling
        Container(
          decoration: BoxDecoration(
            color: Colors.grey[50],
            borderRadius: BorderRadius.circular(16),
            border: Border.all(color: Colors.grey.shade200),
          ),
          child: TextField(
            controller: _passwordController,
            focusNode: _passwordFocus,
            obscureText: !_showPassword,
            textInputAction: TextInputAction.done,
            onSubmitted: (_) => _login(),
            style: TextStyle(fontSize: fontSize),
            decoration: InputDecoration(
              labelText: _loginText('password'),
              labelStyle: TextStyle(color: Colors.grey[600], fontSize: fontSize),
              border: InputBorder.none,
              prefixIcon: Container(
                margin: EdgeInsets.all(iconMargin),
                padding: EdgeInsets.all(iconPadding),
                decoration: BoxDecoration(color: const Color(0xFF3949AB).withValues(alpha: 0.1), borderRadius: BorderRadius.circular(10)),
                child: Icon(Icons.lock_outline_rounded, color: const Color(0xFF3949AB), size: iconSize),
              ),
              suffixIcon: ExcludeFocus(
                child: IconButton(
                  icon: Icon(_showPassword ? Icons.visibility_off_rounded : Icons.visibility_rounded, color: Colors.grey[500], size: iconSize),
                  onPressed: () {
                    setState(() {
                      _showPassword = !_showPassword;
                    });
                  },
                ),
              ),
              contentPadding: EdgeInsets.symmetric(horizontal: 20, vertical: contentPaddingV),
            ),
            enabled: !_isLoggingIn,
          ),
        ),
            ],
          ),
        ),
        SizedBox(height: fieldSpacing),

        // Checkbox บันทึกรหัสผ่าน
        Row(
          children: [
            SizedBox(
              width: 24,
              height: 24,
              child: Checkbox(
                focusNode: FocusNode(canRequestFocus: false),
                value: _rememberPassword,
                onChanged: (value) {
                  setState(() {
                    _rememberPassword = value ?? false;
                  });
                },
                activeColor: const Color(0xFF3949AB),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(4)),
              ),
            ),
            const SizedBox(width: 8),
            GestureDetector(
              onTap: () {
                setState(() {
                  _rememberPassword = !_rememberPassword;
                });
              },
              child: Text(
                _loginText('remember_password'),
                style: TextStyle(fontSize: fontSize, color: Colors.grey[600]),
              ),
            ),
          ],
        ),
        SizedBox(height: sectionSpacing),

        // Login button with gradient
        SizedBox(
          width: double.infinity,
          height: buttonHeight,
          child: Container(
            decoration: BoxDecoration(
              gradient: _isLoggingIn ? null : LinearGradient(colors: [const Color(0xFF3949AB), const Color(0xFF5C6BC0)]),
              color: _isLoggingIn ? Colors.grey[300] : null,
              borderRadius: BorderRadius.circular(16),
              boxShadow: _isLoggingIn ? null : [BoxShadow(color: const Color(0xFF3949AB).withValues(alpha: 0.4), blurRadius: 15, offset: const Offset(0, 8))],
            ),
            child: ElevatedButton(
              focusNode: FocusNode(canRequestFocus: false),
              style: ElevatedButton.styleFrom(
                backgroundColor: Colors.transparent,
                shadowColor: Colors.transparent,
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
              ),
              onPressed: _isLoggingIn ? null : () => _login(),
              child: _isLoggingIn
                  ? Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        SizedBox(width: 20, height: 20, child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2.5)),
                        const SizedBox(width: 10),
                        Text(
                          'กำลังเข้าสู่ระบบ...',
                          style: TextStyle(fontSize: fontSize, fontWeight: FontWeight.w600, color: Colors.grey[600]),
                        ),
                      ],
                    )
                  : Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(Icons.login_rounded, color: Colors.white, size: iconSize),
                        const SizedBox(width: 8),
                        Text(
                          _loginText("login"),
                          style: TextStyle(fontSize: fontSize + 1, fontWeight: FontWeight.w600, color: Colors.white, letterSpacing: 0.5),
                        ),
                      ],
                    ),
            ),
          ),
        ),

        SizedBox(height: isCompact ? 8 : 14),

        // ปุ่มสมัครสมาชิก
        TextButton(
          focusNode: FocusNode(canRequestFocus: false),
          onPressed: _isLoggingIn
              ? null
              : () {
                  Navigator.push(context, MaterialPageRoute(builder: (_) => const RegisterUsernameScreen()));
                },
          child: Text(
            _loginText('register'),
            style: TextStyle(color: const Color(0xFF3949AB), fontSize: isCompact ? 13 : 15, fontWeight: FontWeight.w500),
          ),
        ),

        SizedBox(height: isCompact ? 4 : 8),

        // Language selector with modern style
        InkWell(
          focusNode: FocusNode(canRequestFocus: false),
          onTap: () async {
            final value = await showLanguageSelectionDialog(context);
            if (value != null) {
              global.userLanguage = value;
              global.appConfig.setString('language', global.userLanguage);
              // ไม่ load จาก API ตอนหน้า login — ใช้ hardcode _loginText() แทน
              // จะ load จาก API หลัง login สำเร็จใน _navigateToShopScreen()
              setState(() {});
            }
          },
          borderRadius: BorderRadius.circular(12),
          child: Container(
            padding: EdgeInsets.symmetric(horizontal: isCompact ? 14 : 20, vertical: isCompact ? 10 : 14),
            decoration: BoxDecoration(
              color: Colors.grey[50],
              borderRadius: BorderRadius.circular(12),
              border: Border.all(color: Colors.grey.shade200),
            ),
            child: Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                ClipRRect(
                  borderRadius: BorderRadius.circular(4),
                  child: Image(image: AssetImage("assets/flags/${global.userLanguage}.png"), width: isCompact ? 22 : 28, height: isCompact ? 16 : 20, fit: BoxFit.cover),
                ),
                SizedBox(width: isCompact ? 8 : 12),
                Text(
                  _getLanguageText(),
                  style: TextStyle(fontSize: isCompact ? 12 : 14, fontWeight: FontWeight.w500, color: Colors.grey[700]),
                ),
                SizedBox(width: isCompact ? 4 : 8),
                Icon(Icons.keyboard_arrow_down_rounded, color: Colors.grey[500], size: isCompact ? 16 : 20),
              ],
            ),
          ),
        ),

        SizedBox(height: sectionSpacing),

        // Divider with "OR" text
        ExcludeFocus(
          child: Row(
            children: [
              Expanded(
                child: Container(
                  height: 1,
                  decoration: BoxDecoration(gradient: LinearGradient(colors: [Colors.transparent, Colors.grey.shade300])),
                ),
              ),
              Padding(
                padding: EdgeInsets.symmetric(horizontal: isCompact ? 12 : 20),
                child: Text(
                  "OR",
                  style: TextStyle(color: Colors.grey[500], fontWeight: FontWeight.w500, fontSize: isCompact ? 11 : 13),
                ),
              ),
              Expanded(
                child: Container(
                  height: 1,
                  decoration: BoxDecoration(gradient: LinearGradient(colors: [Colors.grey.shade300, Colors.transparent])),
                ),
              ),
            ],
          ),
        ),

        SizedBox(height: sectionSpacing),

        // Google Sign-In button
        _buildGoogleSignInButton(isCompact: isCompact),

        SizedBox(height: isCompact ? 8 : 16),

        // Line Login button
        _buildLineLoginButton(isCompact: isCompact),
      ],
    );
  }

  Widget _buildGoogleSignInButton({bool isCompact = false}) {
    final buttonHeight = isCompact ? 44.0 : 54.0;

    return Column(
      children: [
        // สำหรับ Web - ใช้ Firebase popup
        if (kIsWeb) ...[
          SizedBox(
            width: double.infinity,
            height: buttonHeight,
            child: OutlinedButton(
              focusNode: FocusNode(canRequestFocus: false),
              style: OutlinedButton.styleFrom(
                foregroundColor: Colors.grey[800],
                backgroundColor: Colors.white,
                side: BorderSide(color: Colors.grey.shade300, width: 1.5),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                elevation: 0,
              ),
              onPressed: _isSigningIn
                  ? null
                  : () async {
                      if (_isSigningIn) return;

                      setState(() {
                        _isSigningIn = true;
                      });

                      try {
                        await Future.delayed(Duration(milliseconds: 100));

                        if (kDebugMode) {
                          AppLogger.debug('🚀 Starting Google Sign-In...');
                        }

                        final GoogleAuthProvider googleProvider = GoogleAuthProvider();
                        googleProvider.addScope('email');
                        googleProvider.addScope('profile');
                        googleProvider.setCustomParameters({'prompt': 'select_account', 'access_type': 'online'});

                        if (kDebugMode) {
                          AppLogger.debug('📱 Opening Google Sign-In popup...');
                        }

                        UserCredential? userCredential;
                        try {
                          userCredential = await _auth.signInWithPopup(googleProvider);
                        } catch (popupError) {
                          if (kDebugMode) {
                            AppLogger.error('❌ Popup error: $popupError');
                          }
                          if (popupError.toString().contains('popup_blocked') || popupError.toString().contains('popup_closed')) {
                            if (kDebugMode) {
                              AppLogger.debug('🔄 Trying redirect method...');
                            }
                            await _auth.signInWithRedirect(googleProvider);
                            return;
                          }
                          rethrow;
                        }

                        if (userCredential.user != null) {
                          if (kDebugMode) {
                            AppLogger.debug('✅ Google Sign-In successful: ${userCredential.user!.email}');
                          }
                          await _handleSuccessfulSignIn(userCredential.user!);
                        } else {
                          throw Exception('ไม่สามารถรับข้อมูลผู้ใช้ได้');
                        }
                      } catch (e) {
                        if (kDebugMode) {
                          AppLogger.error('❌ Google Sign-In Error: $e');
                        }

                        String errorMessage = 'Google Sign-In Error';
                        if (e.toString().contains('popup_blocked')) {
                          errorMessage = 'กรุณาอนุญาต popup ในเบราว์เซอร์';
                        } else if (e.toString().contains('popup_closed')) {
                          errorMessage = 'ผู้ใช้ปิด popup ก่อนเสร็จสิ้นการเข้าสู่ระบบ';
                        } else if (e.toString().contains('network')) {
                          errorMessage = 'ปัญหาการเชื่อมต่อเครือข่าย กรุณาลองใหม่';
                        } else {
                          errorMessage = 'เกิดข้อผิดพลาดในการเข้าสู่ระบบด้วย Google';
                        }

                        if (mounted) {
                          global.showSnackBar(context, const Icon(Icons.error, color: Colors.white), errorMessage, Colors.red);
                        }
                      } finally {
                        if (mounted) {
                          setState(() {
                            _isSigningIn = false;
                          });
                        }
                      }
                    },
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Image(width: 24, height: 24, image: AssetImage("assets/img/google_logo.png")),
                  const SizedBox(width: 14),
                  Text(_isSigningIn ? 'Signing in...' : 'Sign in with Google', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
                  if (_isSigningIn) ...[const SizedBox(width: 12), SizedBox(width: 18, height: 18, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.grey[600]))],
                ],
              ),
            ),
          ),
          const SizedBox(height: 10),
        ] else ...[
          // สำหรับ Desktop (Windows/macOS/Linux)
          // Windows ใช้ GoogleAuthHelper กับ localhost redirect
          // อื่นๆ ใช้ Browser OAuth + Polling ผ่าน Vercel
          SizedBox(
            width: double.infinity,
            height: 54,
            child: OutlinedButton(
              focusNode: FocusNode(canRequestFocus: false),
              style: OutlinedButton.styleFrom(
                foregroundColor: Colors.grey[800],
                backgroundColor: Colors.white,
                side: BorderSide(color: Colors.grey.shade300, width: 1.5),
                shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                elevation: 0,
              ),
              onPressed: _isGoogleLoggingIn
                  ? null
                  : () async {
                      // สำหรับ Windows ใช้ GoogleAuthHelper กับ localhost redirect
                      if (Platform.isWindows || Platform.isMacOS) {
                        await _handleGoogleSignInWithHelper();
                      } else {
                        // สำหรับ Linux หรือ platform อื่น ใช้ Vercel backend
                        _startGoogleLogin();
                      }
                    },
              // ใช้ Flexible สำหรับ Text เพื่อป้องกัน overflow เมื่อ zoom สูง
              child: _isGoogleLoggingIn
                  ? Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        const SizedBox(width: 24, height: 24, child: CircularProgressIndicator(strokeWidth: 2, color: Colors.grey)),
                        const SizedBox(width: 12),
                        Flexible(
                          child: Text(
                            _getGoogleLoggingInText(),
                            style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                      ],
                    )
                  : Row(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Image(width: 24, height: 24, image: AssetImage("assets/img/google_logo.png")),
                        const SizedBox(width: 14),
                        Flexible(
                          child: Text(
                            _getGoogleLoginText(),
                            style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600),
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                      ],
                    ),
            ),
          ),
          const SizedBox(height: 10),
        ],
      ],
    );
  }

  // เมธอดสำหรับดำเนินการ login ด้วยชื่อผู้ใช้และรหัสผ่าน
  void _login() {
    // ตรวจสอบข้อมูลที่ป้อนเข้ามา
    String username = _usernameController.text.trim();
    String password = _passwordController.text;

    if (username.isEmpty) {
      global.showSnackBar(context, const Icon(Icons.error, color: Colors.white), _loginText("please_enter_username"), Colors.red);
      return;
    }

    if (password.isEmpty) {
      global.showSnackBar(context, const Icon(Icons.error, color: Colors.white), _loginText("please_enter_password"), Colors.red);
      return;
    }

    // แสดงสถานะกำลังเข้าสู่ระบบ
    setState(() {
      _isLoggingIn = true;
    });

    // เรียกใช้ Bloc เพื่อดำเนินการ login
    context.read<LoginBloc>().add(LoginOnLoad(userName: username, passWord: password));

    // หลังจากเรียก event ให้ติดตาม state ใน listener
    Future.delayed(const Duration(seconds: 3), () {
      if (mounted && _isLoggingIn) {
        setState(() {
          _isLoggingIn = false;
        });
      }
    });
  }

  /// สร้างปุ่ม Line Login
  Widget _buildLineLoginButton({bool isCompact = false}) {
    final buttonHeight = isCompact ? 44.0 : 54.0;

    return SizedBox(
      width: double.infinity,
      height: buttonHeight,
      child: ElevatedButton(
        focusNode: FocusNode(canRequestFocus: false),
        style: ElevatedButton.styleFrom(
          foregroundColor: Colors.white,
          backgroundColor: const Color(0xFF00B900), // LINE Green color
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
          elevation: 2,
        ),
        onPressed: _isLineLoggingIn ? null : _startLineLogin,
        child: _isLineLoggingIn
            ? const Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  SizedBox(width: 24, height: 24, child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2)),
                  SizedBox(width: 12),
                  Text('กำลังเข้าสู่ระบบ...', style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
                ],
              )
            : Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Image.asset('assets/img/line_logo.png', width: 28, height: 28),
                  const SizedBox(width: 12),
                  Text(_getLineLoginText(), style: TextStyle(fontSize: 16, fontWeight: FontWeight.w600)),
                ],
              ),
      ),
    );
  }

  // Timer สำหรับ polling LINE Login status
  Timer? _lineLoginPollingTimer;

  // Timer สำหรับ polling Google Login status
  Timer? _googleLoginPollingTimer;

  /// LINE Backend URL for LIFF
  String _getLineBackendUrl() {
    return 'https://dev-api.bcaicloud.com/liff';
  }

  /// LIFF ID for LINE Login
  String _getLiffId() {
    return '2008792333-tziDpgnJ';
  }

  /// เริ่มกระบวนการ LINE Login - แสดง QR Code Dialog
  /// ใช้ระบบ QR Code - user scan ด้วย LINE แล้วยืนยันการเข้าสู่ระบบ
  Future<void> _startLineLogin() async {
    try {
      setState(() {
        _isLineLoggingIn = true;
      });

      // แสดง QR Code Dialog และรอผลลัพธ์
      final result = await showLineLoginQRDialog(context: context, liffId: _getLiffId(), liffBaseUrl: _getLineBackendUrl());

      setState(() {
        _isLineLoggingIn = false;
      });

      // ถ้า user ยืนยันการเข้าสู่ระบบสำเร็จ
      if (result != null && mounted) {
        if (kDebugMode) {
          AppLogger.debug('LINE Login Success!');
          AppLogger.debug('LINE User ID: ${result.userId}');
          AppLogger.debug('LINE Display Name: ${result.displayName}');
          AppLogger.debug('LINE Email: ${result.email}');
        }

        // เรียก LineLogin event ใน bloc
        context.read<LoginBloc>().add(LineLogin(lineUserId: result.userId, displayName: result.displayName, pictureUrl: result.pictureUrl ?? '', email: result.email ?? ''));
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Start LINE Login error: $e');
      }
      if (mounted) {
        setState(() {
          _isLineLoggingIn = false;
        });
        global.showSnackBar(context, const Icon(Icons.error, color: Colors.white), 'เกิดข้อผิดพลาด: $e', Colors.red);
      }
    }
  }

  /// เริ่ม polling ตรวจสอบสถานะ LINE Login
  void _startLineLoginPolling(String sessionId, int pollIntervalMs, BuildContext dialogContext) {
    _stopLineLoginPolling();

    final lineBackendUrl = _getLineBackendUrl();

    _lineLoginPollingTimer = Timer.periodic(Duration(milliseconds: pollIntervalMs), (timer) async {
      try {
        final statusResponse = await http.get(Uri.parse('$lineBackendUrl/api/line/session/$sessionId/status'));

        if (statusResponse.statusCode != 200) {
          return; // ลองใหม่รอบหน้า
        }

        final statusData = jsonDecode(statusResponse.body);
        final status = statusData['data']?['status'];

        if (status == 'success') {
          // Login สำเร็จ!
          timer.cancel();
          _lineLoginPollingTimer = null;

          final token = statusData['data']['token'];
          final user = statusData['data']['user'];

          if (kDebugMode) {
            AppLogger.debug('LINE Login Success! Token: $token');
            AppLogger.debug('LINE Login User Data: $user');
            AppLogger.debug('LINE Login display_name: ${user?['display_name']}');
            AppLogger.debug('LINE Login picture_url: ${user?['picture_url']}');
            AppLogger.debug('LINE Login email: ${user?['email']}');
          }

          // ปิด dialog
          if (mounted && Navigator.canPop(dialogContext)) {
            Navigator.pop(dialogContext);
          }

          // ส่งข้อมูลไปยัง Bloc
          if (mounted) {
            context.read<LoginBloc>().add(
              LineLogin(lineUserId: user['line_user_id'] ?? '', accessToken: token, displayName: user['display_name'], pictureUrl: user['picture_url'], email: user['email']),
            );
          }
        } else if (status == 'expired' || status == 'failed') {
          // Session หมดอายุหรือ login ล้มเหลว
          timer.cancel();
          _lineLoginPollingTimer = null;

          final errorMessage = statusData['error']?['message'] ?? 'Session expired';

          if (mounted && Navigator.canPop(dialogContext)) {
            Navigator.pop(dialogContext);
          }

          if (mounted) {
            global.showSnackBar(context, const Icon(Icons.error, color: Colors.white), errorMessage, Colors.red);
          }
        }
        // status == 'pending' → ไม่ต้องทำอะไร รอ poll รอบหน้า
      } catch (e) {
        if (kDebugMode) {
          AppLogger.error('Polling error: $e');
        }
        // ไม่หยุด polling เมื่อเกิด error เพราะอาจเป็นปัญหาชั่วคราว
      }
    });
  }

  /// แสดง Login Code Dialog พร้อม polling รอ Backend ตอบกลับ
  /// user ต้องพิมพ์รหัส 4 หลักส่งไปหา LINE Bot
  void _showLineLoginCodeDialogWithPolling(String loginCode, String sessionId, int pollIntervalMs) {
    showDialog(
      context: context,
      barrierDismissible: false, // ไม่ให้ปิดโดยกดข้างนอก เพราะต้องหยุด polling ก่อน
      builder: (dialogContext) {
        // เริ่ม polling เมื่อ dialog แสดง
        _startLineLoginPolling(sessionId, pollIntervalMs, dialogContext);

        return Dialog(
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
          child: Container(
            width: 380,
            padding: const EdgeInsets.all(24),
            child: SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  // Title
                  Row(
                    children: [
                      Image.asset('assets/img/line_logo.png', width: 36, height: 36),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Text(
                          _getLineLoginTitle(),
                          style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Color(0xFF00B900)),
                        ),
                      ),
                      // ปุ่มปิด
                      IconButton(
                        icon: const Icon(Icons.close, color: Colors.grey),
                        onPressed: () {
                          _stopLineLoginPolling();
                          Navigator.pop(dialogContext);
                        },
                      ),
                    ],
                  ),
                  const SizedBox(height: 24),

                  // เพิ่มเพื่อน LINE Bot - แสดง QR Code สำหรับ user ที่ยังไม่ได้เพิ่มเพื่อน (อยู่ด้านบน)
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: Colors.orange.shade50,
                      borderRadius: BorderRadius.circular(12),
                      border: Border.all(color: Colors.orange.shade200),
                    ),
                    child: Column(
                      children: [
                        Row(
                          children: [
                            Icon(Icons.person_add_alt_1, color: Colors.orange[700], size: 20),
                            const SizedBox(width: 8),
                            Expanded(
                              child: Text(
                                _getAddFriendHintText(),
                                style: TextStyle(fontSize: 13, fontWeight: FontWeight.bold, color: Colors.orange[800]),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 10),
                        // QR Code สำหรับเพิ่มเพื่อน LINE Bot @bcaccount
                        Container(
                          padding: const EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: Colors.white,
                            borderRadius: BorderRadius.circular(8),
                            border: Border.all(color: const Color(0xFF00B900), width: 2),
                          ),
                          child: SizedBox(
                            width: 100,
                            height: 100,
                            child: QrImageView(
                              data: 'https://line.me/R/ti/p/@bcaccount',
                              version: QrVersions.auto,
                              size: 100,
                              backgroundColor: Colors.white,
                              errorStateBuilder: (context, error) {
                                return const Center(child: Text('QR Error', style: TextStyle(fontSize: 10)));
                              },
                            ),
                          ),
                        ),
                        const SizedBox(height: 6),
                        Text(
                          '@bcaccount',
                          style: TextStyle(fontSize: 12, color: Colors.grey[600], fontWeight: FontWeight.w500),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 16),

                  // Login Code - แสดงรหัสตัวใหญ่ๆ
                  Container(
                    padding: const EdgeInsets.symmetric(vertical: 24, horizontal: 32),
                    decoration: BoxDecoration(
                      color: const Color(0xFF00B900).withValues(alpha: 0.1),
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(color: const Color(0xFF00B900), width: 3),
                    ),
                    child: Column(
                      children: [
                        Text(_getLoginCodeLabelText(), style: TextStyle(fontSize: 14, color: Colors.grey[600])),
                        const SizedBox(height: 8),
                        // รหัส Login Code ตัวใหญ่
                        Text(
                          loginCode,
                          style: const TextStyle(fontSize: 56, fontWeight: FontWeight.bold, color: Color(0xFF00B900), letterSpacing: 12, fontFamily: 'monospace'),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 20),

                  // คำแนะนำ - บอกให้พิมพ์รหัสส่งไปหา LINE Bot
                  Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(color: Colors.green.shade50, borderRadius: BorderRadius.circular(12)),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            const Icon(Icons.chat_bubble_outline, color: Color(0xFF00B900), size: 24),
                            const SizedBox(width: 10),
                            Expanded(
                              child: Text(
                                _getLoginCodeInstructionTitle(),
                                style: const TextStyle(fontSize: 15, fontWeight: FontWeight.bold, color: Color(0xFF00B900)),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 12),
                        // ขั้นตอน
                        _buildInstructionStep(1, _getLoginCodeStep1Text()),
                        const SizedBox(height: 8),
                        _buildInstructionStep(2, _getLoginCodeStep2Text()),
                        const SizedBox(height: 8),
                        _buildInstructionStep(3, _getLoginCodeStep3Text()),
                      ],
                    ),
                  ),
                  const SizedBox(height: 16),

                  // สถานะ polling - แสดงว่ากำลังรอ
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(color: Colors.blue.shade50, borderRadius: BorderRadius.circular(12)),
                    child: Row(
                      children: [
                        const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2, color: Color(0xFF00B900))),
                        const SizedBox(width: 12),
                        Expanded(
                          child: Text(_getWaitingForLoginCodeText(), style: TextStyle(fontSize: 14, color: Colors.grey[700])),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 16),

                  // ปุ่มยกเลิก
                  SizedBox(
                    width: double.infinity,
                    child: TextButton(
                      onPressed: () {
                        _stopLineLoginPolling();
                        Navigator.pop(dialogContext);
                      },
                      child: Text(_getCancelText(), style: const TextStyle(color: Colors.grey, fontSize: 16)),
                    ),
                  ),
                ],
              ),
            ),
          ),
        );
      },
    ).then((_) {
      // หยุด polling เมื่อ dialog ถูกปิด
      _stopLineLoginPolling();
    });
  }

  /// สร้าง widget แสดงขั้นตอน
  Widget _buildInstructionStep(int stepNumber, String text) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          width: 22,
          height: 22,
          decoration: BoxDecoration(color: const Color(0xFF00B900), shape: BoxShape.circle),
          child: Center(
            child: Text(
              '$stepNumber',
              style: const TextStyle(color: Colors.white, fontSize: 12, fontWeight: FontWeight.bold),
            ),
          ),
        ),
        const SizedBox(width: 10),
        Expanded(
          child: Text(text, style: TextStyle(fontSize: 14, color: Colors.grey[700])),
        ),
      ],
    );
  }

  // Localized texts for Login Code Dialog
  String _getLineLoginTitle() {
    switch (global.userLanguage) {
      case 'th':
        return 'เข้าสู่ระบบด้วย LINE';
      default:
        return 'Login with LINE';
    }
  }

  String _getLoginCodeLabelText() {
    switch (global.userLanguage) {
      case 'th':
        return 'รหัสเข้าสู่ระบบ';
      default:
        return 'Login Code';
    }
  }

  String _getLoginCodeInstructionTitle() {
    switch (global.userLanguage) {
      case 'th':
        return 'วิธีเข้าสู่ระบบ';
      default:
        return 'How to login';
    }
  }

  String _getLoginCodeStep1Text() {
    switch (global.userLanguage) {
      case 'th':
        return 'เปิดแอป LINE บนมือถือของคุณ';
      default:
        return 'Open LINE app on your phone';
    }
  }

  String _getLoginCodeStep2Text() {
    switch (global.userLanguage) {
      case 'th':
        return 'เข้าห้องแชท LINE Bot ของเรา';
      default:
        return 'Open our LINE Bot chat room';
    }
  }

  String _getLoginCodeStep3Text() {
    switch (global.userLanguage) {
      case 'th':
        return 'พิมพ์รหัสด้านบนส่งไปหา Bot';
      default:
        return 'Type the code above and send to Bot';
    }
  }

  String _getWaitingForLoginCodeText() {
    switch (global.userLanguage) {
      case 'th':
        return 'รอการเข้าสู่ระบบ... กรุณาพิมพ์รหัสส่งไปหา LINE Bot';
      default:
        return 'Waiting for login... Please send the code to LINE Bot';
    }
  }

  String _getAddFriendHintText() {
    switch (global.userLanguage) {
      case 'th':
        return 'ยังไม่ได้เพิ่มเพื่อน?';
      default:
        return 'Not a friend yet?';
    }
  }

  String _getCancelText() {
    switch (global.userLanguage) {
      case 'th':
        return 'ยกเลิก';
      default:
        return 'Cancel';
    }
  }

  /// Get Google Login text based on current language
  String _getGoogleLoginText() {
    switch (global.userLanguage) {
      case 'en':
        return 'Sign in with Google';
      case 'th':
        return 'เข้าสู่ระบบด้วย Google';
      case 'lo':
        return 'ເຂົ້າສູ່ລະບົບດ້ວຍ Google';
      case 'cn':
        return '使用Google登录';
      case 'ja':
        return 'Googleでログイン';
      case 'ko':
        return 'Google로 로그인';
      case 'my':
        return 'Google ဖြင့်ဝင်ရောက်ပါ';
      case 'km':
        return 'ចូលដោយ Google';
      case 'vi':
        return 'Đăng nhập bằng Google';
      default:
        return 'Sign in with Google';
    }
  }

  /// Google Sign-In สำหรับ Windows/macOS ใช้ GoogleAuthHelper กับ localhost redirect
  Future<void> _handleGoogleSignInWithHelper() async {
    if (_isGoogleLoggingIn) return;

    setState(() {
      _isGoogleLoggingIn = true;
    });

    try {
      if (kDebugMode) {
        AppLogger.debug('🚀 Starting Google Sign-In with GoogleAuthHelper...');
      }

      final googleAuthHelper = GoogleAuthHelper();
      final result = await googleAuthHelper.signIn();

      if (kDebugMode) {
        AppLogger.debug('📦 GoogleAuthHelper.signIn() returned: ${result != null ? "UserCredential" : "null"}');
        if (result != null) {
          AppLogger.debug('📦 result.user: ${result.user != null ? result.user!.email : "null"}');
        }
      }

      if (result != null && result.user != null) {
        if (kDebugMode) {
          AppLogger.debug('✅ Google Sign-In successful: ${result.user!.email}');
          AppLogger.debug('🔄 Calling _handleSuccessfulSignIn...');
        }
        await _handleSuccessfulSignIn(result.user!);
        if (kDebugMode) {
          AppLogger.debug('✅ _handleSuccessfulSignIn completed!');
        }
      } else {
        if (kDebugMode) {
          AppLogger.debug('❌ Google Sign-In cancelled or failed - result is null');
        }
        // แสดง error message ให้ผู้ใช้รู้ว่า login ไม่สำเร็จ
        if (mounted) {
          global.showSnackBar(context, const Icon(Icons.warning, color: Colors.white), 'การเข้าสู่ระบบถูกยกเลิกหรือไม่สำเร็จ', Colors.orange);
        }
      }
    } catch (e, stackTrace) {
      if (kDebugMode) {
        AppLogger.error('❌ Google Sign-In Error: $e');
        AppLogger.error('📍 Stack trace: $stackTrace');
      }

      if (mounted) {
        global.showSnackBar(context, const Icon(Icons.error, color: Colors.white), 'เกิดข้อผิดพลาด: ${e.toString().length > 50 ? e.toString().substring(0, 50) : e.toString()}', Colors.red);
      }
    } finally {
      if (mounted) {
        setState(() {
          _isGoogleLoggingIn = false;
        });
      }
    }
  }

  /// Get Google logging in text based on current language
  String _getGoogleLoggingInText() {
    switch (global.userLanguage) {
      case 'th':
        return 'กำลังเข้าสู่ระบบ...';
      default:
        return 'Signing in...';
    }
  }

  /// Google Backend URL - lineoa-liff บน Hetzner (ผ่าน Caddy /liff/*)
  String _getGoogleBackendUrl() {
    return 'https://dev-api.bcaicloud.com/liff';
  }

  /// เปิด Browser สำหรับ Google Login
  Future<void> _openBrowser(String url) async {
    await launchUrl(Uri.parse(url), mode: LaunchMode.externalApplication);
  }

  /// เริ่มกระบวนการ Google Login - สร้าง session และเปิด Browser
  Future<void> _startGoogleLogin() async {
    try {
      setState(() {
        _isGoogleLoggingIn = true;
      });

      final googleBackendUrl = _getGoogleBackendUrl();

      // 1. เรียก Backend สร้าง session
      final sessionResponse = await http.post(
        Uri.parse('$googleBackendUrl/api/google/session'),
        headers: {'Content-Type': 'application/json'},
        body: jsonEncode({'device_info': Platform.operatingSystem}),
      );

      if (sessionResponse.statusCode != 200) {
        throw Exception('Failed to create Google session');
      }

      final sessionData = jsonDecode(sessionResponse.body);
      if (sessionData['success'] != true) {
        throw Exception(sessionData['error']?['message'] ?? 'Failed to create session');
      }

      final sessionId = sessionData['data']['session_id'];
      final loginUrl = sessionData['data']['login_url'];
      final pollInterval = sessionData['data']['poll_interval'] ?? 2000;

      if (kDebugMode) {
        AppLogger.debug('Google Login Session ID: $sessionId');
        AppLogger.debug('Google Login URL: $loginUrl');
      }

      // 2. เปิด Browser (สำหรับ Desktop) หรือ แสดง QR Code (สำหรับ Mobile)
      if (Platform.isWindows || Platform.isMacOS || Platform.isLinux) {
        // Desktop: เปิด Browser ขนาดเล็ก (popup window)
        await _openBrowser(loginUrl);

        setState(() {
          _isGoogleLoggingIn = false;
        });

        // แสดง Dialog รอ login พร้อม polling
        if (mounted) {
          _showGoogleWaitingDialogWithPolling(sessionId, pollInterval);
        }
      } else {
        // Mobile: แสดง QR Code
        setState(() {
          _isGoogleLoggingIn = false;
        });

        if (mounted) {
          _showGoogleQRCodeDialogWithPolling(loginUrl, sessionId, pollInterval);
        }
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Start Google Login error: $e');
      }
      if (mounted) {
        setState(() {
          _isGoogleLoggingIn = false;
        });
        global.showSnackBar(context, const Icon(Icons.error, color: Colors.white), 'เกิดข้อผิดพลาด: $e', Colors.red);
      }
    }
  }

  /// เริ่ม polling ตรวจสอบสถานะ Google Login
  void _startGoogleLoginPolling(String sessionId, int pollIntervalMs, BuildContext dialogContext) {
    _stopGoogleLoginPolling();

    final googleBackendUrl = _getGoogleBackendUrl();

    _googleLoginPollingTimer = Timer.periodic(Duration(milliseconds: pollIntervalMs), (timer) async {
      try {
        final statusResponse = await http.get(Uri.parse('$googleBackendUrl/api/google/session/$sessionId/status'));

        if (statusResponse.statusCode != 200) {
          return; // ลองใหม่รอบหน้า
        }

        final statusData = jsonDecode(statusResponse.body);
        final status = statusData['data']?['status'];

        if (status == 'success') {
          // Login สำเร็จ!
          timer.cancel();
          _googleLoginPollingTimer = null;

          final token = statusData['data']['token'];
          final user = statusData['data']['user'];

          if (kDebugMode) {
            AppLogger.debug('Google Login Success! Token: $token');
            AppLogger.debug('Google Login User Data: $user');
            AppLogger.debug('Google Login display_name: ${user?['display_name']}');
            AppLogger.debug('Google Login picture_url: ${user?['picture_url']}');
            AppLogger.debug('Google Login email: ${user?['email']}');
          }

          // ปิด dialog
          if (mounted && Navigator.canPop(dialogContext)) {
            Navigator.pop(dialogContext);
          }

          // ส่งข้อมูลไปยัง Bloc
          if (mounted) {
            context.read<LoginBloc>().add(
              GoogleLogin(googleUserId: user['google_user_id'] ?? '', accessToken: token, displayName: user['display_name'], pictureUrl: user['picture_url'], email: user['email']),
            );
          }
        } else if (status == 'expired' || status == 'failed') {
          // Session หมดอายุหรือ login ล้มเหลว
          timer.cancel();
          _googleLoginPollingTimer = null;

          final errorMessage = statusData['error']?['message'] ?? 'Session expired';

          if (mounted && Navigator.canPop(dialogContext)) {
            Navigator.pop(dialogContext);
          }

          if (mounted) {
            global.showSnackBar(context, const Icon(Icons.error, color: Colors.white), errorMessage, Colors.red);
          }
        }
        // status == 'pending' → ไม่ต้องทำอะไร รอ poll รอบหน้า
      } catch (e) {
        if (kDebugMode) {
          AppLogger.error('Google Polling error: $e');
        }
        // ไม่หยุด polling เมื่อเกิด error เพราะอาจเป็นปัญหาชั่วคราว
      }
    });
  }

  /// แสดง Dialog รอ login (สำหรับ Desktop ที่เปิด Browser แล้ว)
  void _showGoogleWaitingDialogWithPolling(String sessionId, int pollIntervalMs) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (dialogContext) {
        // เริ่ม polling เมื่อ dialog แสดง
        _startGoogleLoginPolling(sessionId, pollIntervalMs, dialogContext);

        return Dialog(
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
          child: Container(
            width: 350,
            padding: const EdgeInsets.all(24),
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Title
                Row(
                  children: [
                    Image(width: 36, height: 36, image: AssetImage("assets/img/google_logo.png")),
                    const SizedBox(width: 12),
                    Expanded(
                      child: Text(
                        _getGoogleLoginTitle(),
                        style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Color(0xFF4285F4)),
                      ),
                    ),
                    // ปุ่มปิด
                    IconButton(
                      icon: const Icon(Icons.close, color: Colors.grey),
                      onPressed: () {
                        _stopGoogleLoginPolling();
                        Navigator.pop(dialogContext);
                      },
                    ),
                  ],
                ),
                const SizedBox(height: 24),

                // Icon แสดงว่าเปิด Browser แล้ว
                Container(
                  padding: const EdgeInsets.all(20),
                  decoration: BoxDecoration(color: Colors.blue.shade50, shape: BoxShape.circle),
                  child: Icon(Icons.open_in_browser, size: 48, color: const Color(0xFF4285F4)),
                ),
                const SizedBox(height: 20),

                // คำแนะนำ
                Text(
                  _getGoogleBrowserInstructionText(),
                  textAlign: TextAlign.center,
                  style: TextStyle(fontSize: 16, color: Colors.grey[700]),
                ),
                const SizedBox(height: 20),

                // สถานะ polling
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(color: Colors.blue.shade50, borderRadius: BorderRadius.circular(12)),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.center,
                    children: [
                      const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2, color: Color(0xFF4285F4))),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Text(_getWaitingForGoogleLoginText(), style: TextStyle(fontSize: 14, color: Colors.grey[700])),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 20),

                // ปุ่มยกเลิก
                SizedBox(
                  width: double.infinity,
                  child: TextButton(
                    onPressed: () {
                      _stopGoogleLoginPolling();
                      Navigator.pop(dialogContext);
                    },
                    child: Text(_getCancelText(), style: const TextStyle(color: Colors.grey, fontSize: 16)),
                  ),
                ),
              ],
            ),
          ),
        );
      },
    ).then((_) {
      // หยุด polling เมื่อ dialog ถูกปิด
      _stopGoogleLoginPolling();
    });
  }

  /// แสดง QR Code Dialog สำหรับ Mobile (คล้าย LINE Login)
  void _showGoogleQRCodeDialogWithPolling(String loginUrl, String sessionId, int pollIntervalMs) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (dialogContext) {
        // เริ่ม polling เมื่อ dialog แสดง
        _startGoogleLoginPolling(sessionId, pollIntervalMs, dialogContext);

        return Dialog(
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(20)),
          child: Container(
            width: 350,
            padding: const EdgeInsets.all(20),
            child: SingleChildScrollView(
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  // Title
                  Row(
                    children: [
                      Image(width: 36, height: 36, image: AssetImage("assets/img/google_logo.png")),
                      const SizedBox(width: 12),
                      Expanded(
                        child: Text(
                          _getGoogleLoginTitle(),
                          style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold, color: Color(0xFF4285F4)),
                        ),
                      ),
                      // ปุ่มปิด
                      IconButton(
                        icon: const Icon(Icons.close, color: Colors.grey),
                        onPressed: () {
                          _stopGoogleLoginPolling();
                          Navigator.pop(dialogContext);
                        },
                      ),
                    ],
                  ),
                  const SizedBox(height: 16),

                  // QR Code
                  Container(
                    padding: const EdgeInsets.all(16),
                    decoration: BoxDecoration(
                      color: Colors.white,
                      borderRadius: BorderRadius.circular(16),
                      border: Border.all(color: const Color(0xFF4285F4), width: 3),
                    ),
                    child: SizedBox(
                      width: 200,
                      height: 200,
                      child: QrImageView(
                        data: loginUrl,
                        version: QrVersions.auto,
                        size: 200,
                        backgroundColor: Colors.white,
                        errorStateBuilder: (context, error) {
                          return const Center(child: Text('ไม่สามารถสร้าง QR Code ได้'));
                        },
                      ),
                    ),
                  ),
                  const SizedBox(height: 16),

                  // คำแนะนำ
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(color: Colors.blue.shade50, borderRadius: BorderRadius.circular(12)),
                    child: Row(
                      children: [
                        const Icon(Icons.qr_code_scanner, color: Color(0xFF4285F4), size: 24),
                        const SizedBox(width: 8),
                        Expanded(
                          child: Text(_getGoogleQRCodeInstructionText(), style: TextStyle(fontSize: 14, color: Colors.grey[700])),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 16),

                  // สถานะ polling
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(color: Colors.blue.shade50, borderRadius: BorderRadius.circular(12)),
                    child: Row(
                      children: [
                        const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(strokeWidth: 2, color: Color(0xFF4285F4))),
                        const SizedBox(width: 12),
                        Expanded(
                          child: Text(_getWaitingForGoogleLoginText(), style: TextStyle(fontSize: 14, color: Colors.grey[700])),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 16),

                  // ปุ่มยกเลิก
                  SizedBox(
                    width: double.infinity,
                    child: TextButton(
                      onPressed: () {
                        _stopGoogleLoginPolling();
                        Navigator.pop(dialogContext);
                      },
                      child: Text(_getCancelText(), style: const TextStyle(color: Colors.grey, fontSize: 16)),
                    ),
                  ),
                ],
              ),
            ),
          ),
        );
      },
    ).then((_) {
      // หยุด polling เมื่อ dialog ถูกปิด
      _stopGoogleLoginPolling();
    });
  }

  // Localized texts for Google Login Dialog
  String _getGoogleLoginTitle() {
    switch (global.userLanguage) {
      case 'th':
        return 'เข้าสู่ระบบด้วย Google';
      default:
        return 'Sign in with Google';
    }
  }

  String _getGoogleBrowserInstructionText() {
    switch (global.userLanguage) {
      case 'th':
        return 'เบราว์เซอร์ถูกเปิดแล้ว\nกรุณาเข้าสู่ระบบด้วย Google Account ของคุณ';
      default:
        return 'Browser has been opened\nPlease sign in with your Google Account';
    }
  }

  String _getGoogleQRCodeInstructionText() {
    switch (global.userLanguage) {
      case 'th':
        return 'สแกน QR Code ด้วยมือถือของคุณเพื่อเข้าสู่ระบบ Google';
      default:
        return 'Scan QR Code with your phone to sign in with Google';
    }
  }

  String _getWaitingForGoogleLoginText() {
    switch (global.userLanguage) {
      case 'th':
        return 'รอการเข้าสู่ระบบ... กรุณาอนุญาตการเข้าถึงในเบราว์เซอร์';
      default:
        return 'Waiting for login... Please allow access in browser';
    }
  }

  Widget _buildTermsAndPrivacyPolicy({bool isCompact = false}) {
    final fontSize = isCompact ? 11.0 : 13.0;
    final padding = isCompact ? 8.0 : 12.0;

    return Container(
      padding: EdgeInsets.symmetric(horizontal: 16, vertical: padding),
      decoration: BoxDecoration(color: Colors.white.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(12)),
      child: Column(
        children: [
          RichText(
            textAlign: TextAlign.center,
            text: TextSpan(
              style: TextStyle(color: Colors.white.withValues(alpha: 0.8), fontSize: fontSize, height: 1.5),
              children: <TextSpan>[
                TextSpan(text: _loginText('you_have_accepted')),
                TextSpan(
                  text: ' ${_loginText('terms_of_use')}',
                  style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, decoration: TextDecoration.underline),
                  recognizer: TapGestureRecognizer()
                    ..onTap = () {
                      // ignore: deprecated_member_use
                      launch('https://www.dohome.co.th/terms');
                    },
                ),
              ],
            ),
          ),
          SizedBox(height: isCompact ? 2 : 4),
          RichText(
            textAlign: TextAlign.center,
            text: TextSpan(
              style: TextStyle(color: Colors.white.withValues(alpha: 0.8), fontSize: fontSize),
              children: <TextSpan>[
                TextSpan(text: _loginText('read')),
                TextSpan(
                  text: ' ${_loginText('privacy_policy')}',
                  style: TextStyle(color: Colors.white, fontWeight: FontWeight.bold, decoration: TextDecoration.underline),
                  recognizer: TapGestureRecognizer()
                    ..onTap = () {
                      // ignore: deprecated_member_use
                      launch('https://www.dohome.co.th/privacy');
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
