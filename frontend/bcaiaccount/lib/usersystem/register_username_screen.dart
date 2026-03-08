import 'package:flutter/material.dart';
import 'package:smlaicloud/flavors.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/repositories/user_repository.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// หน้าสมัครสมาชิกด้วยรหัสพนักงาน + รหัสผ่าน (ไม่ต้องมี email)
class RegisterUsernameScreen extends StatefulWidget {
  const RegisterUsernameScreen({super.key});

  @override
  State<RegisterUsernameScreen> createState() => _RegisterUsernameScreenState();
}

class _RegisterUsernameScreenState extends State<RegisterUsernameScreen> {
  final TextEditingController _usernameController = TextEditingController();
  final TextEditingController _nameController = TextEditingController();
  final TextEditingController _passwordController = TextEditingController();
  final TextEditingController _confirmPasswordController = TextEditingController();
  final FocusNode _usernameFocus = FocusNode();
  final FocusNode _nameFocus = FocusNode();
  final FocusNode _passwordFocus = FocusNode();
  final FocusNode _confirmPasswordFocus = FocusNode();
  bool _showPassword = false;
  bool _showConfirmPassword = false;
  bool _isRegistering = false;

  /// Multi-language text สำหรับหน้า Register
  static const Map<String, Map<String, String>> _registerTexts = {
    'register': {'en': 'Register', 'th': 'สมัครผู้ใช้งาน', 'lo': 'ລົງທະບຽນ', 'zh': '注册', 'ja': '登録', 'ko': '회원가입', 'my': 'စာရင်းသွင်းပါ', 'km': 'ចុះឈ្មោះ', 'vi': 'Đăng ký'},
    'employee_code': {
      'en': 'Employee Code',
      'th': 'รหัสพนักงาน',
      'lo': 'ລະຫັດພະນັກງານ',
      'zh': '员工编号',
      'ja': '社員コード',
      'ko': '사원코드',
      'my': 'ဝန်ထမ်းကုဒ်',
      'km': 'លេខបញ្ជាក់និយោជិក',
      'vi': 'Mã nhân viên',
    },
    'password': {'en': 'Password', 'th': 'รหัสผ่าน', 'lo': 'ລະຫັດຜ່ານ', 'zh': '密码', 'ja': 'パスワード', 'ko': '비밀번호', 'my': 'စကားဝှက်', 'km': 'ពាក្យសម្ងាត់', 'vi': 'Mật khẩu'},
    'confirm_password': {
      'en': 'Confirm Password',
      'th': 'ยืนยันรหัสผ่าน',
      'lo': 'ຢືນຢັນລະຫັດຜ່ານ',
      'zh': '确认密码',
      'ja': 'パスワード確認',
      'ko': '비밀번호 확인',
      'my': 'စကားဝှက်အတည်ပြုပါ',
      'km': 'បញ្ជាក់ពាក្យសម្ងាត់',
      'vi': 'Xác nhận mật khẩu',
    },
    'please_enter_employee_code': {
      'en': 'Please enter employee code',
      'th': 'กรุณากรอกรหัสพนักงาน',
      'lo': 'ກະລຸນາປ້ອນລະຫັດພະນັກງານ',
      'zh': '请输入员工编号',
      'ja': '社員コードを入力してください',
      'ko': '사원코드를 입력하세요',
      'my': 'ကျေးဇူးပြု၍ ဝန်ထမ်းကုဒ်ထည့်ပါ',
      'km': 'សូមបញ្ចូលលេខបញ្ជាក់និយោជិក',
      'vi': 'Vui lòng nhập mã nhân viên',
    },
    'please_enter_password': {
      'en': 'Please enter password',
      'th': 'กรุณากรอกรหัสผ่าน',
      'lo': 'ກະລຸນາປ້ອນລະຫັດຜ່ານ',
      'zh': '请输入密码',
      'ja': 'パスワードを入力してください',
      'ko': '비밀번호를 입력하세요',
      'my': 'ကျေးဇူးပြု၍ စကားဝှက်ထည့်ပါ',
      'km': 'សូមបញ្ចូលពាក្យសម្ងាត់',
      'vi': 'Vui lòng nhập mật khẩu',
    },
    'password_min_length': {
      'en': 'Password must be at least 5 characters',
      'th': 'รหัสผ่านต้องมีอย่างน้อย 5 ตัวอักษร',
      'lo': 'ລະຫັດຜ່ານຕ້ອງມີຢ່າງໜ້ອຍ 5 ຕົວອັກສອນ',
      'zh': '密码至少需要5个字符',
      'ja': 'パスワードは5文字以上必要です',
      'ko': '비밀번호는 최소 5자 이상이어야 합니다',
      'my': 'စကားဝှက်သည် အနည်းဆုံး 5 လုံးရှိရမည်',
      'km': 'ពាក្យសម្ងាត់ត្រូវមានយ៉ាងហោចណាស់ 5 តួអក្សរ',
      'vi': 'Mật khẩu phải có ít nhất 5 ký tự',
    },
    'password_not_match': {
      'en': 'Passwords do not match',
      'th': 'รหัสผ่านไม่ตรงกัน',
      'lo': 'ລະຫັດຜ່ານບໍ່ກົງກັນ',
      'zh': '密码不一致',
      'ja': 'パスワードが一致しません',
      'ko': '비밀번호가 일치하지 않습니다',
      'my': 'စကားဝှက်မတူညီပါ',
      'km': 'ពាក្យសម្ងាត់មិនត្រូវគ្នា',
      'vi': 'Mật khẩu không khớp',
    },
    'name': {
      'en': 'Name',
      'th': 'ชื่อ',
      'lo': 'ຊື່',
      'zh': '姓名',
      'ja': '名前',
      'ko': '이름',
      'my': 'အမည်',
      'km': 'ឈ្មោះ',
      'vi': 'Tên',
    },
    'please_enter_name': {
      'en': 'Please enter your name',
      'th': 'กรุณากรอกชื่อ',
      'lo': 'ກະລຸນາປ້ອນຊື່',
      'zh': '请输入姓名',
      'ja': '名前を入力してください',
      'ko': '이름을 입력하세요',
      'my': 'ကျေးဇူးပြု၍ အမည်ထည့်ပါ',
      'km': 'សូមបញ្ចូលឈ្មោះ',
      'vi': 'Vui lòng nhập tên',
    },
    'employee_code_min_length': {
      'en': 'Employee code must be at least 5 characters',
      'th': 'รหัสพนักงานต้องมีอย่างน้อย 5 ตัวอักษร',
      'lo': 'ລະຫັດພະນັກງານຕ້ອງມີຢ່າງໜ້ອຍ 5 ຕົວອັກສອນ',
      'zh': '员工编号至少需要5个字符',
      'ja': '社員コードは5文字以上必要です',
      'ko': '사원코드는 최소 5자 이상이어야 합니다',
      'my': 'ဝန်ထမ်းကုဒ်သည် အနည်းဆုံး 5 လုံးရှိရမည်',
      'km': 'លេខបញ្ជាក់និយោជិកត្រូវមានយ៉ាងហោចណាស់ 5 តួអក្សរ',
      'vi': 'Mã nhân viên phải có ít nhất 5 ký tự',
    },
    'registering': {
      'en': 'Registering...',
      'th': 'กำลังสมัคร...',
      'lo': 'ກຳລັງລົງທະບຽນ...',
      'zh': '注册中...',
      'ja': '登録中...',
      'ko': '가입 중...',
      'my': 'စာရင်းသွင်းနေသည်...',
      'km': 'កំពុងចុះឈ្មោះ...',
      'vi': 'Đang đăng ký...',
    },
    'register_success': {
      'en': 'Registration successful! Please login.',
      'th': 'สมัครผู้ใช้งานสำเร็จ! กรุณาเข้าสู่ระบบ',
      'lo': 'ລົງທະບຽນສຳເລັດ! ກະລຸນາເຂົ້າສູ່ລະບົບ',
      'zh': '注册成功！请登录',
      'ja': '登録成功！ログインしてください',
      'ko': '가입 성공! 로그인하세요',
      'my': 'စာရင်းသွင်းခြင်းအောင်မြင်သည်! ကျေးဇူးပြု၍ ဝင်ရောက်ပါ',
      'km': 'ចុះឈ្មោះដោយជោគជ័យ! សូមចូល',
      'vi': 'Đăng ký thành công! Vui lòng đăng nhập',
    },
    'register_failed': {
      'en': 'Registration failed',
      'th': 'สมัครผู้ใช้งานไม่สำเร็จ',
      'lo': 'ລົງທະບຽນບໍ່ສຳເລັດ',
      'zh': '注册失败',
      'ja': '登録失敗',
      'ko': '가입 실패',
      'my': 'စာရင်းသွင်းခြင်းမအောင်မြင်ပါ',
      'km': 'ការចុះឈ្មោះបានបរាជ័យ',
      'vi': 'Đăng ký thất bại',
    },
    'username_exists': {
      'en': 'This employee code is already taken',
      'th': 'รหัสพนักงานนี้ถูกใช้แล้ว',
      'lo': 'ລະຫັດພະນັກງານນີ້ຖືກໃຊ້ແລ້ວ',
      'zh': '此员工编号已被使用',
      'ja': 'この社員コードは既に使用されています',
      'ko': '이 사원코드는 이미 사용 중입니다',
      'my': 'ဤဝန်ထမ်းကုဒ်ကို အသုံးပြုပြီးဖြစ်သည်',
      'km': 'លេខបញ្ជាក់និយោជិកនេះត្រូវបានប្រើប្រាស់រួចហើយ',
      'vi': 'Mã nhân viên này đã được sử dụng',
    },
    'back_to_login': {
      'en': 'Back to Login',
      'th': 'กลับหน้าเข้าสู่ระบบ',
      'lo': 'ກັບຄືນໜ້າເຂົ້າສູ່ລະບົບ',
      'zh': '返回登录',
      'ja': 'ログインに戻る',
      'ko': '로그인으로 돌아가기',
      'my': 'လော့ဂ်အင်စာမျက်နှာသို့ ပြန်သွားပါ',
      'km': 'ត្រលប់ទៅទំព័រចូល',
      'vi': 'Quay lại trang đăng nhập',
    },
    'create_account': {
      'en': 'Create Account',
      'th': 'สร้างบัญชี',
      'lo': 'ສ້າງບັນຊີ',
      'zh': '创建账号',
      'ja': 'アカウント作成',
      'ko': '계정 만들기',
      'my': 'အကောင့်ဖန်တီးပါ',
      'km': 'បង្កើតគណនី',
      'vi': 'Tạo tài khoản',
    },
  };

  String _text(String key) {
    final lang = global.userLanguage;
    return _registerTexts[key]?[lang] ?? _registerTexts[key]?['en'] ?? key;
  }

  @override
  void initState() {
    super.initState();
    // auto-focus รหัสพนักงาน เมื่อเปิดหน้านี้
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) _usernameFocus.requestFocus();
    });
  }

  @override
  void dispose() {
    _usernameController.dispose();
    _nameController.dispose();
    _passwordController.dispose();
    _confirmPasswordController.dispose();
    _usernameFocus.dispose();
    _nameFocus.dispose();
    _passwordFocus.dispose();
    _confirmPasswordFocus.dispose();
    super.dispose();
  }

  Future<void> _register() async {
    final username = _usernameController.text.trim();
    final name = _nameController.text.trim();
    final password = _passwordController.text;
    final confirmPassword = _confirmPasswordController.text;

    // ตรวจสอบข้อมูล
    if (username.isEmpty) {
      _showError(_text('please_enter_employee_code'));
      return;
    }
    if (username.length < 5) {
      _showError(_text('employee_code_min_length'));
      return;
    }
    if (name.isEmpty) {
      _showError(_text('please_enter_name'));
      return;
    }
    if (password.isEmpty) {
      _showError(_text('please_enter_password'));
      return;
    }
    if (password.length < 5) {
      _showError(_text('password_min_length'));
      return;
    }
    if (password != confirmPassword) {
      _showError(_text('password_not_match'));
      return;
    }

    setState(() => _isRegistering = true);

    try {
      final userRepo = UserRepository();
      final result = await userRepo.registerByUsername(
        username,
        password,
        global.appConfig.getString('timezonelabel') ?? 'Asia/Bangkok',
        global.appConfig.getString('timezoneoffset') ?? '+07:00',
        global.appConfig.getString('yeartype') ?? 'buddhist',
        name: name,
      );

      if (!mounted) return;

      if (result.success) {
        AppLogger.info('[Register] สมัครสมาชิกสำเร็จ: $username');
        global.showSnackBar(context, const Icon(Icons.check_circle, color: Colors.white), _text('register_success'), Colors.green);
        Navigator.pop(context);
      } else {
        AppLogger.warning('[Register] สมัครสมาชิกไม่สำเร็จ: $username');
        _showError(_text('register_failed'));
      }
    } catch (e) {
      if (!mounted) return;
      AppLogger.error('[Register] เกิดข้อผิดพลาด: $e');
      final errorMsg = e.toString().toLowerCase();
      if (errorMsg.contains('exists')) {
        _showError(_text('username_exists'));
      } else {
        _showError('${_text('register_failed')}: $e');
      }
    } finally {
      if (mounted) {
        setState(() => _isRegistering = false);
      }
    }
  }

  void _showError(String message) {
    global.showSnackBar(context, const Icon(Icons.error, color: Colors.white), message, Colors.red);
  }

  @override
  Widget build(BuildContext context) {
    final size = MediaQuery.of(context).size;

    return Scaffold(
      body: Container(
        width: double.infinity,
        height: double.infinity,
        decoration: BoxDecoration(
          gradient: LinearGradient(begin: Alignment.topLeft, end: Alignment.bottomRight, colors: [F.primaryGradientColor, F.secondaryGradientColor]),
        ),
        child: SafeArea(
          child: Center(
            child: SingleChildScrollView(
              padding: EdgeInsets.symmetric(horizontal: size.width > 600 ? size.width * 0.2 : 24, vertical: 24),
              child: _buildRegisterForm(),
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildRegisterForm() {
    final screenHeight = MediaQuery.of(context).size.height;
    final isCompact = screenHeight < 800;
    final cardPadding = isCompact ? 20.0 : 32.0;
    final sectionSpacing = isCompact ? 16.0 : 28.0;
    final fieldSpacing = isCompact ? 10.0 : 16.0;
    final buttonHeight = isCompact ? 46.0 : 56.0;
    final fontSize = isCompact ? 14.0 : 16.0;
    final iconSize = isCompact ? 18.0 : 22.0;
    final iconMargin = isCompact ? 8.0 : 12.0;
    final iconPadding = isCompact ? 6.0 : 8.0;
    final contentPaddingV = isCompact ? 12.0 : 18.0;

    return ConstrainedBox(
      constraints: const BoxConstraints(maxWidth: 420),
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          // Logo
          Container(
            padding: EdgeInsets.all(isCompact ? 14 : 20),
            decoration: BoxDecoration(color: Colors.white.withValues(alpha: 0.15), borderRadius: BorderRadius.circular(20)),
            child: ClipRRect(
              borderRadius: BorderRadius.circular(12),
              child: Image.asset(
                F.logoPath,
                width: isCompact ? 60 : 80,
                height: isCompact ? 60 : 80,
                fit: BoxFit.contain,
                errorBuilder: (context, error, stackTrace) {
                  return Icon(Icons.cloud_outlined, size: isCompact ? 45 : 60, color: Colors.white);
                },
              ),
            ),
          ),
          SizedBox(height: isCompact ? 8 : 16),

          // Main card
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
                // Title
                Icon(Icons.person_add_rounded, size: isCompact ? 36 : 48, color: const Color(0xFF3949AB)),
                SizedBox(height: isCompact ? 8 : 12),
                Text(
                  _text('create_account'),
                  style: TextStyle(fontSize: isCompact ? 22 : 28, fontWeight: FontWeight.bold, color: const Color(0xFF1A237E)),
                ),
                SizedBox(height: sectionSpacing),

                // Tab วนกลับ: username → password → confirm → username
                FocusTraversalGroup(
                  child: Column(
                    children: [
                      // รหัสพนักงาน
                      _buildTextField(
                        controller: _usernameController,
                        focusNode: _usernameFocus,
                        label: _text('employee_code'),
                        icon: Icons.badge_outlined,
                        textInputAction: TextInputAction.next,
                        onSubmitted: (_) => _nameFocus.requestFocus(),
                        fontSize: fontSize,
                        iconSize: iconSize,
                        iconMargin: iconMargin,
                        iconPadding: iconPadding,
                        contentPaddingV: contentPaddingV,
                      ),
                      SizedBox(height: fieldSpacing),

                      // ชื่อ
                      _buildTextField(
                        controller: _nameController,
                        focusNode: _nameFocus,
                        label: _text('name'),
                        icon: Icons.person_outline,
                        textInputAction: TextInputAction.next,
                        onSubmitted: (_) => _passwordFocus.requestFocus(),
                        fontSize: fontSize,
                        iconSize: iconSize,
                        iconMargin: iconMargin,
                        iconPadding: iconPadding,
                        contentPaddingV: contentPaddingV,
                      ),
                      SizedBox(height: fieldSpacing),

                      // รหัสผ่าน
                      _buildTextField(
                        controller: _passwordController,
                        focusNode: _passwordFocus,
                        label: _text('password'),
                        icon: Icons.lock_outline_rounded,
                        isPassword: true,
                        showPassword: _showPassword,
                        onTogglePassword: () => setState(() => _showPassword = !_showPassword),
                        textInputAction: TextInputAction.next,
                        onSubmitted: (_) => _confirmPasswordFocus.requestFocus(),
                        fontSize: fontSize,
                        iconSize: iconSize,
                        iconMargin: iconMargin,
                        iconPadding: iconPadding,
                        contentPaddingV: contentPaddingV,
                      ),
                      SizedBox(height: fieldSpacing),

                      // ยืนยันรหัสผ่าน
                      _buildTextField(
                        controller: _confirmPasswordController,
                        focusNode: _confirmPasswordFocus,
                        label: _text('confirm_password'),
                        icon: Icons.lock_outline_rounded,
                        isPassword: true,
                        showPassword: _showConfirmPassword,
                        onTogglePassword: () => setState(() => _showConfirmPassword = !_showConfirmPassword),
                        textInputAction: TextInputAction.done,
                        onSubmitted: (_) => _register(),
                        fontSize: fontSize,
                        iconSize: iconSize,
                        iconMargin: iconMargin,
                        iconPadding: iconPadding,
                        contentPaddingV: contentPaddingV,
                      ),
                    ],
                  ),
                ),
                SizedBox(height: sectionSpacing),

                // ปุ่มสมัครสมาชิก
                SizedBox(
                  width: double.infinity,
                  height: buttonHeight,
                  child: Container(
                    decoration: BoxDecoration(
                      gradient: _isRegistering ? null : const LinearGradient(colors: [Color(0xFF3949AB), Color(0xFF5C6BC0)]),
                      color: _isRegistering ? Colors.grey[300] : null,
                      borderRadius: BorderRadius.circular(16),
                      boxShadow: _isRegistering ? null : [BoxShadow(color: const Color(0xFF3949AB).withValues(alpha: 0.4), blurRadius: 15, offset: const Offset(0, 8))],
                    ),
                    child: ExcludeFocus(
                      child: ElevatedButton(
                        focusNode: FocusNode(canRequestFocus: false),
                        style: ElevatedButton.styleFrom(
                          backgroundColor: Colors.transparent,
                          shadowColor: Colors.transparent,
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
                        ),
                        onPressed: _isRegistering ? null : _register,
                        child: _isRegistering
                            ? Row(
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  const SizedBox(width: 20, height: 20, child: CircularProgressIndicator(color: Colors.white, strokeWidth: 2.5)),
                                  const SizedBox(width: 10),
                                  Text(
                                    _text('registering'),
                                    style: TextStyle(fontSize: fontSize, fontWeight: FontWeight.w600, color: Colors.grey[600]),
                                  ),
                                ],
                              )
                            : Row(
                                mainAxisAlignment: MainAxisAlignment.center,
                                children: [
                                  Icon(Icons.person_add_rounded, color: Colors.white, size: iconSize),
                                  const SizedBox(width: 8),
                                  Text(
                                    _text('register'),
                                    style: TextStyle(fontSize: fontSize + 1, fontWeight: FontWeight.w600, color: Colors.white, letterSpacing: 0.5),
                                  ),
                                ],
                              ),
                      ),
                    ),
                  ),
                ),

                SizedBox(height: isCompact ? 12 : 20),

                // กลับหน้า Login
                ExcludeFocus(
                  child: TextButton(
                    focusNode: FocusNode(canRequestFocus: false),
                    onPressed: () => Navigator.pop(context),
                    child: Text(
                      _text('back_to_login'),
                      style: TextStyle(color: const Color(0xFF3949AB), fontSize: fontSize, fontWeight: FontWeight.w500),
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildTextField({
    required TextEditingController controller,
    required String label,
    required IconData icon,
    FocusNode? focusNode,
    bool isPassword = false,
    bool showPassword = false,
    VoidCallback? onTogglePassword,
    TextInputAction? textInputAction,
    ValueChanged<String>? onSubmitted,
    required double fontSize,
    required double iconSize,
    required double iconMargin,
    required double iconPadding,
    required double contentPaddingV,
  }) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.grey[50],
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: TextField(
        controller: controller,
        focusNode: focusNode,
        obscureText: isPassword && !showPassword,
        textInputAction: textInputAction,
        onSubmitted: onSubmitted,
        style: TextStyle(fontSize: fontSize),
        decoration: InputDecoration(
          labelText: label,
          labelStyle: TextStyle(color: Colors.grey[600], fontSize: fontSize),
          border: InputBorder.none,
          prefixIcon: Container(
            margin: EdgeInsets.all(iconMargin),
            padding: EdgeInsets.all(iconPadding),
            decoration: BoxDecoration(color: const Color(0xFF3949AB).withValues(alpha: 0.1), borderRadius: BorderRadius.circular(10)),
            child: Icon(icon, color: const Color(0xFF3949AB), size: iconSize),
          ),
          suffixIcon: isPassword
              ? ExcludeFocus(
                  child: IconButton(
                    icon: Icon(showPassword ? Icons.visibility_off_rounded : Icons.visibility_rounded, color: Colors.grey[500], size: iconSize),
                    onPressed: onTogglePassword,
                  ),
                )
              : null,
          contentPadding: EdgeInsets.symmetric(horizontal: 20, vertical: contentPaddingV),
        ),
        enabled: !_isRegistering,
      ),
    );
  }
}
