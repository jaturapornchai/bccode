import 'dart:async';
import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:http/http.dart' as http;
import 'package:smlaicloud/model/permission_model.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// Service สำหรับจัดการสิทธิ์ผ่าน MongoDB Atlas
/// - ใช้ /atlas/get, /atlas/update, /atlas/delete เท่านั้น
/// - มี cache และ refresh อัตโนมัติทุก 1 นาที
class PermissionService {
  static final PermissionService _instance = PermissionService._internal();
  factory PermissionService() => _instance;
  PermissionService._internal();

  // Collection names
  static const String _permissionDefinitionsCollection = 'permission_definitions';
  static const String _employeePermissionsCollection = 'employee_permissions';

  // Cache
  UserEffectivePermissionModel? _cachedPermissions;
  // Timer สำหรับ refresh อัตโนมัติ
  Timer? _refreshTimer;

  // Callback เมื่อสิทธิ์เปลี่ยน
  final List<VoidCallback> _onPermissionChangedListeners = [];

  /// เริ่ม service และ timer refresh (ทุก 5 นาที)
  void startAutoRefresh() {
    _refreshTimer?.cancel();
    _refreshTimer = Timer.periodic(const Duration(minutes: 5), (_) {
      _refreshPermissions();
    });
    if (kDebugMode) {
      AppLogger.debug('🔄 PermissionService: Auto refresh started (every 5 minutes)');
    }
  }

  /// หยุด service และ timer
  void stopAutoRefresh() {
    _refreshTimer?.cancel();
    _refreshTimer = null;
    if (kDebugMode) {
      AppLogger.debug('🔄 PermissionService: Auto refresh stopped');
    }
  }

  /// เพิ่ม listener เมื่อสิทธิ์เปลี่ยน
  void addOnPermissionChangedListener(VoidCallback callback) {
    _onPermissionChangedListeners.add(callback);
  }

  /// ลบ listener
  void removeOnPermissionChangedListener(VoidCallback callback) {
    _onPermissionChangedListeners.remove(callback);
  }

  /// แจ้ง listeners ว่าสิทธิ์เปลี่ยน
  void _notifyPermissionChanged() {
    for (var callback in _onPermissionChangedListeners) {
      callback();
    }
  }

  /// Refresh สิทธิ์ (เรียกจาก timer หรือ manual)
  Future<void> _refreshPermissions() async {
    try {
      final employeeCode = global.profileData.username ?? '';
      if (employeeCode.isEmpty) return;

      final oldPermissions = _cachedPermissions;
      await loadUserPermissions(employeeCode, forceRefresh: true);

      // ตรวจสอบว่าสิทธิ์เปลี่ยนหรือไม่
      if (_hasPermissionsChanged(oldPermissions, _cachedPermissions)) {
        if (kDebugMode) {
          AppLogger.debug('🔄 PermissionService: Permissions changed, notifying listeners');
        }
        _notifyPermissionChanged();
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ PermissionService: Error refreshing permissions: $e');
      }
    }
  }

  /// ตรวจสอบว่าสิทธิ์เปลี่ยนหรือไม่
  bool _hasPermissionsChanged(
      UserEffectivePermissionModel? oldPermissions,
      UserEffectivePermissionModel? newPermissions) {
    if (oldPermissions == null && newPermissions == null) return false;
    if (oldPermissions == null || newPermissions == null) return true;
    // เปรียบเทียบ JSON string
    return jsonEncode(oldPermissions.permissions) !=
        jsonEncode(newPermissions.permissions);
  }

  // ============================================
  // MongoDB Atlas API Methods
  // ============================================

  /// สร้าง headers สำหรับ API (รวม Authorization token)
  Map<String, String> get _headers {
    final headers = <String, String>{'Content-Type': 'application/json'};
    final token = global.appConfig.getString("token") ?? '';
    if (token.isNotEmpty) {
      headers['Authorization'] = 'Bearer $token';
    }
    return headers;
  }

  /// สร้าง base URL สำหรับ Atlas API
  String get _baseUrl => global.goApiUrlPath('atlas');

  /// GET - ดึงข้อมูลจาก MongoDB Atlas
  Future<List<Map<String, dynamic>>> _atlasGet({
    required String collection,
    required String shopid,
    String? email,
    String? cartid,
  }) async {
    try {
      final url = '$_baseUrl/get';
      final body = {
        'collection': collection,
        'shopid': shopid,
        if (email != null) 'email': email,
        if (cartid != null) 'cartid': cartid,
      };

      if (kDebugMode) {
        AppLogger.debug('📤 Atlas GET: $url');
        AppLogger.debug('📤 Body: ${jsonEncode(body)}');
      }

      final response = await http.post(
        Uri.parse(url),
        headers: _headers,
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        final data = jsonDecode(response.body);
        if (data['status'] == 'success' && data['data'] != null) {
          return List<Map<String, dynamic>>.from(data['data']);
        }
      }

      if (kDebugMode) {
        AppLogger.debug('📥 Atlas GET Response: ${response.body}');
      }

      return [];
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Atlas GET Error: $e');
      }
      return [];
    }
  }

  /// UPDATE - บันทึก/อัปเดตข้อมูลใน MongoDB Atlas
  Future<bool> _atlasUpdate({
    required String collection,
    required String shopid,
    required String email,
    required String cartid,
    required Map<String, dynamic> data,
    bool upsert = true,
  }) async {
    try {
      final url = '$_baseUrl/update';

      // ลบ _id ออกจาก data เพราะ MongoDB ไม่อนุญาตให้แก้ไข _id
      final cleanData = Map<String, dynamic>.from(data);
      cleanData.remove('_id');

      final body = {
        'collection': collection,
        'shopid': shopid,
        'email': email,
        'cartid': cartid,
        'data': cleanData,
        'upsert': upsert,
      };

      if (kDebugMode) {
        AppLogger.debug('📤 Atlas UPDATE: $url');
        AppLogger.debug('📤 Body: ${jsonEncode(body)}');
      }

      final response = await http.post(
        Uri.parse(url),
        headers: _headers,
        body: jsonEncode(body),
      );

      if (kDebugMode) {
        AppLogger.debug('📥 Atlas UPDATE Response: ${response.body}');
      }

      return response.statusCode == 200;
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Atlas UPDATE Error: $e');
      }
      return false;
    }
  }

  /// DELETE - ลบข้อมูลจาก MongoDB Atlas
  Future<bool> _atlasDelete({
    required String collection,
    required String shopid,
    required String email,
    required String cartid,
  }) async {
    try {
      final url = '$_baseUrl/delete';
      final body = {
        'collection': collection,
        'shopid': shopid,
        'email': email,
        'cartid': cartid,
        'delete_many': false,
      };

      if (kDebugMode) {
        AppLogger.debug('📤 Atlas DELETE: $url');
        AppLogger.debug('📤 Body: ${jsonEncode(body)}');
      }

      final response = await http.post(
        Uri.parse(url),
        headers: _headers,
        body: jsonEncode(body),
      );

      if (kDebugMode) {
        AppLogger.debug('📥 Atlas DELETE Response: ${response.body}');
      }

      return response.statusCode == 200;
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ Atlas DELETE Error: $e');
      }
      return false;
    }
  }

  // ============================================
  // Permission Definition Methods
  // ============================================

  /// ดึงรายการกำหนดสิทธิ์ทั้งหมดของ shop
  Future<List<PermissionDefinitionModel>> getPermissionDefinitions(String shopid) async {
    final results = await _atlasGet(
      collection: _permissionDefinitionsCollection,
      shopid: shopid,
    );

    return results.map((json) => PermissionDefinitionModel.fromJson(json)).toList();
  }

  /// ดึงกำหนดสิทธิ์ตามรหัส
  Future<PermissionDefinitionModel?> getPermissionDefinitionByCode(
      String shopid, String permissionCode) async {
    final results = await _atlasGet(
      collection: _permissionDefinitionsCollection,
      shopid: shopid,
      email: permissionCode, // ใช้ email field เป็น query key
    );

    if (results.isNotEmpty) {
      return PermissionDefinitionModel.fromJson(results.first);
    }
    return null;
  }

  /// บันทึกกำหนดสิทธิ์
  Future<bool> savePermissionDefinition(PermissionDefinitionModel definition) async {
    definition.updatedAt = DateTime.now().toIso8601String();
    definition.updatedBy = global.profileData.username ?? '';

    return await _atlasUpdate(
      collection: _permissionDefinitionsCollection,
      shopid: definition.shopid,
      email: definition.permissionCode, // ใช้ permissionCode เป็น key
      cartid: definition.permissionCode,
      data: definition.toJson(),
      upsert: true,
    );
  }

  /// ลบกำหนดสิทธิ์
  Future<bool> deletePermissionDefinition(String shopid, String permissionCode) async {
    return await _atlasDelete(
      collection: _permissionDefinitionsCollection,
      shopid: shopid,
      email: permissionCode,
      cartid: permissionCode,
    );
  }

  /// คัดลอกกำหนดสิทธิ์จาก source → destination พร้อมรหัสและชื่อใหม่
  Future<bool> copyPermissionDefinition(
    String shopid,
    String sourceCode,
    String destinationCode,
    String destinationName,
  ) async {
    final source = await getPermissionDefinitionByCode(shopid, sourceCode);
    if (source == null) {
      throw Exception('ไม่พบสิทธิ์ต้นแบบ: $sourceCode');
    }

    final now = DateTime.now().toIso8601String();
    final username = global.profileData.username ?? '';

    // สร้าง model ใหม่โดยตรง (ไม่ใช้ copyWith เพราะ copyWith จะ fallback id เดิม)
    final copied = PermissionDefinitionModel(
      id: null, // ให้ Atlas generate _id ใหม่
      shopid: source.shopid,
      permissionCode: destinationCode,
      permissionName: destinationName,
      description: source.description,
      branches: source.branches.map((b) => b.copyWith()).toList(),
      isActive: source.isActive,
      createdAt: now,
      updatedAt: now,
      createdBy: username,
      updatedBy: username,
    );

    return savePermissionDefinition(copied);
  }

  // ============================================
  // Employee Permission Methods
  // ============================================

  /// ดึงการเชื่อมสิทธิ์ของพนักงานทั้งหมดของ shop
  Future<List<EmployeePermissionModel>> getEmployeePermissions(String shopid) async {
    final results = await _atlasGet(
      collection: _employeePermissionsCollection,
      shopid: shopid,
    );

    return results.map((json) => EmployeePermissionModel.fromJson(json)).toList();
  }

  /// ดึงการเชื่อมสิทธิ์ของพนักงานคนเดียว
  Future<EmployeePermissionModel?> getEmployeePermission(
      String shopid, String employeeCode) async {
    final results = await _atlasGet(
      collection: _employeePermissionsCollection,
      shopid: shopid,
      email: employeeCode,
    );

    if (results.isNotEmpty) {
      return EmployeePermissionModel.fromJson(results.first);
    }
    return null;
  }

  /// บันทึกการเชื่อมสิทธิ์พนักงาน
  Future<bool> saveEmployeePermission(EmployeePermissionModel employeePermission) async {
    employeePermission.updatedAt = DateTime.now().toIso8601String();
    employeePermission.updatedBy = global.profileData.username ?? '';

    return await _atlasUpdate(
      collection: _employeePermissionsCollection,
      shopid: employeePermission.shopid,
      email: employeePermission.employeeCode,
      cartid: employeePermission.employeeCode,
      data: employeePermission.toJson(),
      upsert: true,
    );
  }

  /// ลบการเชื่อมสิทธิ์พนักงาน
  Future<bool> deleteEmployeePermission(String shopid, String employeeCode) async {
    return await _atlasDelete(
      collection: _employeePermissionsCollection,
      shopid: shopid,
      email: employeeCode,
      cartid: employeeCode,
    );
  }

  // ============================================
  // User Effective Permissions (OR combined)
  // ============================================

  /// โหลดสิทธิ์รวมของผู้ใช้ (OR จากทุก permission codes ที่เชื่อม)
  Future<UserEffectivePermissionModel?> loadUserPermissions(
    String employeeCode, {
    bool forceRefresh = false,
  }) async {
    // ใช้ cache ถ้ายังไม่หมดอายุ
    if (!forceRefresh &&
        _cachedPermissions != null &&
        _cachedPermissions!.employeeCode == employeeCode &&
        !_cachedPermissions!.isExpired()) {
      return _cachedPermissions;
    }

    try {
      final shopid = global.getShopId();

      // 1. ดึงการเชื่อมสิทธิ์ของพนักงาน
      final employeePermission = await getEmployeePermission(shopid, employeeCode);
      if (employeePermission == null || employeePermission.permissionCodes.isEmpty) {
        _cachedPermissions = UserEffectivePermissionModel(
          employeeCode: employeeCode,
          permissions: {},
        );
        return _cachedPermissions;
      }

      // 2. ดึง definition ของแต่ละ permission code
      final List<PermissionDefinitionModel> definitions = [];
      for (var code in employeePermission.permissionCodes) {
        final def = await getPermissionDefinitionByCode(shopid, code);
        if (def != null && def.isActive) {
          definitions.add(def);
        }
      }

      // 3. OR สิทธิ์ทั้งหมดเข้าด้วยกัน
      final combinedPermissions = _combinePermissions(definitions);

      _cachedPermissions = UserEffectivePermissionModel(
        employeeCode: employeeCode,
        permissions: combinedPermissions,
      );

      if (kDebugMode) {
        AppLogger.debug('✅ PermissionService: Loaded permissions for $employeeCode');
        AppLogger.debug('   Branches: ${combinedPermissions.keys.toList()}');
      }

      return _cachedPermissions;
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('❌ PermissionService: Error loading permissions: $e');
      }
      return null;
    }
  }

  /// OR สิทธิ์จากหลาย definitions เข้าด้วยกัน
  Map<String, Map<String, ScreenPermissionModel>> _combinePermissions(
    List<PermissionDefinitionModel> definitions,
  ) {
    final Map<String, Map<String, ScreenPermissionModel>> result = {};

    for (var def in definitions) {
      for (var branch in def.branches) {
        // สร้าง branch ถ้ายังไม่มี
        result.putIfAbsent(branch.branchCode, () => {});

        for (var screen in branch.screens) {
          final existing = result[branch.branchCode]![screen.screenCode];

          if (existing == null) {
            // ยังไม่มี - ใช้ค่าจาก definition นี้
            result[branch.branchCode]![screen.screenCode] = screen.copyWith();
          } else {
            // มีแล้ว - OR สิทธิ์เข้าด้วยกัน
            result[branch.branchCode]![screen.screenCode] = ScreenPermissionModel(
              screenCode: screen.screenCode,
              screenName: screen.screenName.isNotEmpty ? screen.screenName : existing.screenName,
              canView: existing.canView || screen.canView,
              canAdd: existing.canAdd || screen.canAdd,
              canEditOwn: existing.canEditOwn || screen.canEditOwn,
              canEditAll: existing.canEditAll || screen.canEditAll,
              canDeleteOwn: existing.canDeleteOwn || screen.canDeleteOwn,
              canDeleteAll: existing.canDeleteAll || screen.canDeleteAll,
              canPrint: existing.canPrint || screen.canPrint,
            );
          }
        }
      }
    }

    return result;
  }

  /// ดึงสิทธิ์ปัจจุบันจาก cache
  UserEffectivePermissionModel? get currentPermissions => _cachedPermissions;

  /// ตรวจสอบสิทธิ์ (ใช้ cache)
  bool hasPermission(String branchCode, String screenCode, String action) {
    if (_cachedPermissions == null) return false;

    final screenPerm = _cachedPermissions!.getScreenPermission(branchCode, screenCode);
    if (screenPerm == null) return false;

    switch (action) {
      case 'add':
        return screenPerm.canAdd;
      case 'edit':
        return screenPerm.canEdit;
      case 'delete':
        return screenPerm.canDelete;
      case 'print':
        return screenPerm.canPrint;
      case 'view':
        return screenPerm.canView;
      default:
        return false;
    }
  }

  /// Clear cache
  void clearCache() {
    _cachedPermissions = null;
  }

  /// Dispose
  void dispose() {
    stopAutoRefresh();
    _onPermissionChangedListeners.clear();
    clearCache();
  }
}

/// Global instance
final permissionService = PermissionService();
