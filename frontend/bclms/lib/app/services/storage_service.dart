import 'dart:convert';
import 'package:flutter_secure_storage/flutter_secure_storage.dart';
import 'package:bclms/app/core/utils/app_logger.dart';

class StorageService {
  late final FlutterSecureStorage _storage;

  static const _keyToken = 'token';
  static const _keyRefresh = 'refresh';
  static const _keyShopId = 'shopid';
  static const _keyShopName = 'shopname';
  static const _keyRole = 'role';
  static const _keyBranchGuidFixed = 'branch_guidfixed';
  static const _keyUsername = 'username';
  static const _keyName = 'name';
  static const _keySavedUsername = 'saved_username';
  static const _keySavedPassword = 'saved_password';
  static const _keyRememberLogin = 'remember_login';
  static const _keyBackendUrl = 'backend_url';
  static const _keyBackendUrlHistory = 'backend_url_history';
  static const _maxUrlHistory = 10;

  Future<StorageService> init() async {
    _storage = const FlutterSecureStorage(
      aOptions: AndroidOptions(encryptedSharedPreferences: true),
      iOptions: IOSOptions(accessibility: KeychainAccessibility.first_unlock),
      webOptions: WebOptions(dbName: 'bclms', publicKey: 'bclms_public_key'),
    );
    AppLogger.info('[Storage] StorageService เริ่มต้นสำเร็จ');
    return this;
  }

  // === Helper: เขียนทีละตัว (ป้องกัน concurrent write issue บน Windows) ===
  Future<void> _writeSequential(Map<String, String> entries) async {
    for (final entry in entries.entries) {
      await _storage.write(key: entry.key, value: entry.value);
    }
  }

  Future<void> _deleteSequential(List<String> keys) async {
    for (final key in keys) {
      await _storage.delete(key: key);
    }
  }

  // Token
  Future<void> saveTokens({
    required String token,
    required String refresh,
  }) async {
    await _writeSequential({
      _keyToken: token,
      _keyRefresh: refresh,
    });
    AppLogger.debug('[Storage] บันทึก tokens สำเร็จ');
  }

  Future<String?> getToken() => _storage.read(key: _keyToken);
  Future<String?> getRefreshToken() => _storage.read(key: _keyRefresh);

  // Shop
  Future<void> saveShopData({
    required String shopId,
    required String shopName,
    required int role,
  }) async {
    await _writeSequential({
      _keyShopId: shopId,
      _keyShopName: shopName,
      _keyRole: role.toString(),
    });
    AppLogger.info('[Storage] บันทึกข้อมูลร้านค้า: $shopName สำเร็จ');
  }

  Future<String?> getShopId() => _storage.read(key: _keyShopId);
  Future<String?> getShopName() => _storage.read(key: _keyShopName);
  Future<int?> getRole() async {
    final value = await _storage.read(key: _keyRole);
    return value != null ? int.tryParse(value) : null;
  }

  // Branch
  Future<void> saveBranchGuidFixed(String guidFixed) async {
    await _storage.write(key: _keyBranchGuidFixed, value: guidFixed);
    AppLogger.info('[Storage] บันทึกสาขา: $guidFixed สำเร็จ');
  }

  Future<String?> getBranchGuidFixed() =>
      _storage.read(key: _keyBranchGuidFixed);

  // User info
  Future<void> saveUserInfo({
    required String username,
    required String name,
  }) async {
    await _writeSequential({
      _keyUsername: username,
      _keyName: name,
    });
    AppLogger.info('[Storage] บันทึกข้อมูลผู้ใช้: $username สำเร็จ');
  }

  Future<String?> getUsername() => _storage.read(key: _keyUsername);
  Future<String?> getName() => _storage.read(key: _keyName);

  // บันทึกรหัสผ่าน (Remember login)
  Future<void> saveCredentials({
    required String username,
    required String password,
  }) async {
    // เขียนทีละตัว ป้องกัน Windows credential storage เสียหาย
    await _writeSequential({
      _keySavedUsername: username,
      _keySavedPassword: password,
      _keyRememberLogin: 'true',
    });

    // ตรวจสอบว่าเขียนสำเร็จจริง
    final verifyUser = await _storage.read(key: _keySavedUsername);
    final verifyRemember = await _storage.read(key: _keyRememberLogin);
    if (verifyUser == null || verifyRemember == null) {
      AppLogger.error('[Storage] ❌ เขียน credentials ไม่สำเร็จ! verifyUser: $verifyUser, verifyRemember: $verifyRemember');
      // ลองเขียนใหม่อีกครั้ง
      await _storage.write(key: _keySavedUsername, value: username);
      await _storage.write(key: _keySavedPassword, value: password);
      await _storage.write(key: _keyRememberLogin, value: 'true');
      AppLogger.info('[Storage] ลองเขียน credentials ใหม่แล้ว');
    } else {
      AppLogger.info('[Storage] บันทึก credentials สำเร็จ (verified: user=$verifyUser)');
    }
  }

  Future<void> clearCredentials() async {
    await _deleteSequential([
      _keySavedUsername,
      _keySavedPassword,
      _keyRememberLogin,
    ]);
    AppLogger.debug('[Storage] ลบ credentials สำเร็จ');
  }

  Future<String?> getSavedUsername() => _storage.read(key: _keySavedUsername);
  Future<String?> getSavedPassword() => _storage.read(key: _keySavedPassword);
  Future<bool> getRememberLogin() async {
    final value = await _storage.read(key: _keyRememberLogin);
    return value == 'true';
  }

  // Backend URL
  Future<void> saveBackendUrl(String url) async {
    final normalized = _normalizeUrl(url);
    if (normalized.isEmpty) return;

    await _storage.write(key: _keyBackendUrl, value: normalized);

    // เพิ่มใน history (newest first, max 10)
    final history = await getBackendUrlHistory();
    history.remove(normalized);
    history.insert(0, normalized);
    if (history.length > _maxUrlHistory) {
      history.removeRange(_maxUrlHistory, history.length);
    }
    await _storage.write(key: _keyBackendUrlHistory, value: jsonEncode(history));
    AppLogger.info('[Storage] บันทึก Backend URL: $normalized');
  }

  Future<String?> getBackendUrl() => _storage.read(key: _keyBackendUrl);

  Future<List<String>> getBackendUrlHistory() async {
    final raw = await _storage.read(key: _keyBackendUrlHistory);
    if (raw == null || raw.isEmpty) return [];
    try {
      final List<dynamic> decoded = jsonDecode(raw);
      return decoded.cast<String>();
    } catch (_) {
      return [];
    }
  }

  Future<void> removeBackendUrlFromHistory(String url) async {
    final history = await getBackendUrlHistory();
    history.remove(url);
    await _storage.write(key: _keyBackendUrlHistory, value: jsonEncode(history));
  }

  String _normalizeUrl(String url) {
    var trimmed = url.trim();
    if (trimmed.isEmpty) return '';
    if (!trimmed.startsWith('http://') && !trimmed.startsWith('https://')) {
      trimmed = 'http://$trimmed';
    }
    if (trimmed.endsWith('/')) {
      trimmed = trimmed.substring(0, trimmed.length - 1);
    }
    return trimmed;
  }

  // Clear all (เก็บ saved credentials + backend URL ไว้)
  Future<void> clearAll() async {
    final rememberLogin = await getRememberLogin();
    final savedUser = await getSavedUsername();
    final savedPass = await getSavedPassword();
    final backendUrl = await getBackendUrl();
    final urlHistory = await _storage.read(key: _keyBackendUrlHistory);
    AppLogger.debug('[Storage] clearAll — remember: $rememberLogin, savedUser: "$savedUser", savedPass: ${savedPass != null ? "(${savedPass.length} ตัวอักษร)" : "null"}');

    await _storage.deleteAll();

    // คืนค่า saved credentials ถ้าเปิด remember login
    if (rememberLogin && savedUser != null && savedPass != null) {
      await saveCredentials(username: savedUser, password: savedPass);
    }
    // คืนค่า backend URL
    if (backendUrl != null) {
      await _storage.write(key: _keyBackendUrl, value: backendUrl);
    }
    if (urlHistory != null) {
      await _storage.write(key: _keyBackendUrlHistory, value: urlHistory);
    }
    AppLogger.info('[Storage] ลบข้อมูลทั้งหมดใน storage สำเร็จ');
  }
}
