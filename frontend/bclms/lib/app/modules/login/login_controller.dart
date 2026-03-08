import 'package:dio/dio.dart';
import 'package:flutter/material.dart';
import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:bclms/app/core/utils/app_logger.dart';
import 'package:bclms/app/providers/service_providers.dart';

class LoginState {
  final bool isLoading;
  final bool obscurePassword;
  final bool rememberLogin;
  final String? errorMessage;
  final bool loginSuccess;
  final String backendUrl;
  final List<String> urlHistory;
  final String? connectionTestStatus; // null, 'testing', 'success', 'failed'

  const LoginState({
    this.isLoading = false,
    this.obscurePassword = true,
    this.rememberLogin = false,
    this.errorMessage,
    this.loginSuccess = false,
    this.backendUrl = '',
    this.urlHistory = const [],
    this.connectionTestStatus,
  });

  LoginState copyWith({
    bool? isLoading,
    bool? obscurePassword,
    bool? rememberLogin,
    String? errorMessage,
    bool? loginSuccess,
    String? backendUrl,
    List<String>? urlHistory,
    String? Function()? connectionTestStatus,
  }) {
    return LoginState(
      isLoading: isLoading ?? this.isLoading,
      obscurePassword: obscurePassword ?? this.obscurePassword,
      rememberLogin: rememberLogin ?? this.rememberLogin,
      errorMessage: errorMessage,
      loginSuccess: loginSuccess ?? this.loginSuccess,
      backendUrl: backendUrl ?? this.backendUrl,
      urlHistory: urlHistory ?? this.urlHistory,
      connectionTestStatus: connectionTestStatus != null
          ? connectionTestStatus()
          : this.connectionTestStatus,
    );
  }
}

class LoginNotifier extends Notifier<LoginState> {
  late final TextEditingController usernameController;
  late final TextEditingController passwordController;
  final formKey = GlobalKey<FormState>();

  @override
  LoginState build() {
    usernameController = TextEditingController();
    passwordController = TextEditingController();

    ref.onDispose(() {
      usernameController.dispose();
      passwordController.dispose();
    });

    _loadInitialData();
    return const LoginState();
  }

  Future<void> _loadInitialData() async {
    await _loadBackendUrl();
    await _loadSavedCredentials();
  }

  Future<void> _loadBackendUrl() async {
    try {
      final api = ref.read(apiServiceProvider);
      final storage = ref.read(storageServiceProvider);
      final history = await storage.getBackendUrlHistory();
      state = state.copyWith(
        backendUrl: api.baseUrl,
        urlHistory: history,
      );
    } catch (e) {
      AppLogger.error('[Login] โหลด Backend URL ล้มเหลว', error: e);
    }
  }

  Future<void> _loadSavedCredentials() async {
    try {
      final storage = ref.read(storageServiceProvider);
      final remember = await storage.getRememberLogin();
      AppLogger.debug('[Login] โหลดรหัส — remember: $remember');
      state = state.copyWith(rememberLogin: remember);
      if (remember) {
        final savedUser = await storage.getSavedUsername();
        final savedPass = await storage.getSavedPassword();
        AppLogger.debug('[Login] โหลดรหัส — username: "$savedUser", password: ${savedPass != null ? "(${savedPass.length} ตัวอักษร)" : "null"}');
        if (savedUser != null) usernameController.text = savedUser;
        if (savedPass != null) passwordController.text = savedPass;
        AppLogger.info('[Login] โหลดข้อมูลเข้าสู่ระบบที่บันทึกไว้สำเร็จ');
      }
    } catch (e) {
      AppLogger.error('[Login] โหลดข้อมูลเข้าสู่ระบบล้มเหลว', error: e);
    }
  }

  // === Backend URL ===

  Future<void> setBackendUrl(String url) async {
    final storage = ref.read(storageServiceProvider);
    final api = ref.read(apiServiceProvider);
    await storage.saveBackendUrl(url);
    api.updateBaseUrl(url);
    final history = await storage.getBackendUrlHistory();
    state = state.copyWith(
      backendUrl: url,
      urlHistory: history,
      connectionTestStatus: () => null,
    );
  }

  Future<void> addBackendUrl(String url) async {
    if (url.trim().isEmpty) return;
    await setBackendUrl(url);
  }

  Future<void> removeUrlFromHistory(String url) async {
    final storage = ref.read(storageServiceProvider);
    await storage.removeBackendUrlFromHistory(url);
    final history = await storage.getBackendUrlHistory();
    state = state.copyWith(urlHistory: history);
  }

  Future<void> testConnection() async {
    state = state.copyWith(connectionTestStatus: () => 'testing');
    try {
      final testDio = Dio(BaseOptions(
        connectTimeout: const Duration(seconds: 5),
        receiveTimeout: const Duration(seconds: 5),
      ));
      final url = state.backendUrl.endsWith('/')
          ? state.backendUrl
          : '${state.backendUrl}/';
      final response = await testDio.get(url);
      state = state.copyWith(
        connectionTestStatus: () =>
            response.statusCode != null && response.statusCode! < 500
                ? 'success'
                : 'failed',
      );
    } catch (e) {
      AppLogger.error('[Login] ทดสอบเชื่อมต่อ Server ล้มเหลว', error: e);
      state = state.copyWith(connectionTestStatus: () => 'failed');
    }
  }

  // === Auth ===

  void togglePasswordVisibility() {
    state = state.copyWith(obscurePassword: !state.obscurePassword);
  }

  void toggleRememberLogin(bool? value) {
    state = state.copyWith(rememberLogin: value ?? false);
  }

  Future<void> login() async {
    if (!formKey.currentState!.validate()) return;

    state = state.copyWith(isLoading: true, loginSuccess: false);
    try {
      final authService = ref.read(authServiceProvider);
      final storage = ref.read(storageServiceProvider);
      final username = usernameController.text.trim();
      final password = passwordController.text;

      final result = await authService.login(username, password);

      if (result.success) {
        if (state.rememberLogin) {
          AppLogger.info('[Login] บันทึกรหัส — remember: true');
          await storage.saveCredentials(username: username, password: password);
        } else {
          AppLogger.info('[Login] ลบรหัส — remember: false');
          await storage.clearCredentials();
        }
        state = state.copyWith(isLoading: false, loginSuccess: true);
        return;
      } else {
        state = state.copyWith(
          isLoading: false,
          errorMessage: result.message ?? 'ชื่อผู้ใช้หรือรหัสผ่านไม่ถูกต้อง',
        );
      }
    } catch (e) {
      AppLogger.error('เข้าสู่ระบบผิดพลาด', error: e);
      state = state.copyWith(
        isLoading: false,
        errorMessage: 'ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้',
      );
    }
  }
}

final loginProvider = NotifierProvider<LoginNotifier, LoginState>(
  LoginNotifier.new,
);
