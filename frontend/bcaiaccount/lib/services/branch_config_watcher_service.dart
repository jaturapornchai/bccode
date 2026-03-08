import 'dart:async';
import 'dart:convert';
import 'package:crypto/crypto.dart';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/repositories/company_branch_repository.dart';
import 'package:smlaicloud/services/notification_queue_service.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// Service สำหรับตรวจสอบการเปลี่ยนแปลง config ของสาขา
/// และ reload อัตโนมัติเมื่อมีการเปลี่ยนแปลง
class BranchConfigWatcherService {
  // Singleton instance
  static final BranchConfigWatcherService _instance = BranchConfigWatcherService._internal();
  factory BranchConfigWatcherService() => _instance;
  BranchConfigWatcherService._internal();

  // Repository
  final CompanyBranchRepository _companyBranchRepository = CompanyBranchRepository();

  // Hash ของ config ปัจจุบัน
  String? _currentConfigHash;

  // Branch guidfixed ที่กำลังติดตาม
  String? _watchingBranchGuid;

  // Timer สำหรับ polling
  Timer? _pollingTimer;

  // Interval สำหรับ polling (วินาที) - default 60 วินาที
  int _pollingIntervalSeconds = 60;

  // สถานะว่ากำลัง polling อยู่หรือไม่
  bool _isPolling = false;

  // Callback เมื่อ config เปลี่ยน
  void Function(CompanyBranchModel newConfig, List<String> changedFields)? onConfigChanged;

  // Stream controller สำหรับ broadcast การเปลี่ยนแปลง
  final _configChangeController = StreamController<BranchConfigChangeEvent>.broadcast();

  // Stream สำหรับ listen การเปลี่ยนแปลง
  Stream<BranchConfigChangeEvent> get configChangeStream => _configChangeController.stream;

  /// เริ่มติดตาม config ของสาขา
  /// [branchGuidFixed] - GUID ของสาขาที่ต้องการติดตาม
  /// [intervalSeconds] - ความถี่ในการตรวจสอบ (วินาที)
  void startWatching({
    required String branchGuidFixed,
    int intervalSeconds = 60,
  }) {
    if (branchGuidFixed.isEmpty) {
      AppLogger.warning('[BranchConfigWatcher] Cannot watch empty branch GUID');
      return;
    }

    _watchingBranchGuid = branchGuidFixed;
    _pollingIntervalSeconds = intervalSeconds;

    // คำนวณ hash จาก config ปัจจุบัน
    _updateCurrentConfigHash();

    // หยุด timer เก่า (ถ้ามี) และป้องกันการสร้างซ้อน
    if (_pollingTimer != null) {
      AppLogger.debug('[BranchConfigWatcher] Already watching, restarting...');
      _stopPollingTimer();
    }

    // เริ่ม timer ใหม่
    _pollingTimer = Timer.periodic(
      Duration(seconds: _pollingIntervalSeconds),
      (_) => _checkForChanges(),
    );

    _isPolling = true;
    AppLogger.info('[BranchConfigWatcher] Started watching branch: $branchGuidFixed (interval: ${_pollingIntervalSeconds}s)');
  }

  /// หยุดติดตาม
  void stopWatching() {
    _stopPollingTimer();
    _watchingBranchGuid = null;
    _currentConfigHash = null;
    _isPolling = false;
    AppLogger.info('[BranchConfigWatcher] Stopped');
  }

  void _stopPollingTimer() {
    _pollingTimer?.cancel();
    _pollingTimer = null;
  }

  /// อัพเดท hash จาก config ปัจจุบัน
  void _updateCurrentConfigHash() {
    _currentConfigHash = _calculateConfigHash(global.companyBranchSelectData);
    AppLogger.debug('[BranchConfigWatcher] Current config hash: $_currentConfigHash');
  }

  /// คำนวณ hash จาก config
  String _calculateConfigHash(CompanyBranchModel config) {
    // สร้าง JSON string จาก config fields ที่สำคัญ
    final configData = {
      'guidfixed': config.guidfixed,
      'code': config.code,
      'names': config.names.map((e) => {'code': e.code, 'name': e.name}).toList(),
      'companynames': config.companynames.map((e) => {'code': e.code, 'name': e.name}).toList(),
      'languages': config.languages,
      'baseCurrency': config.baseCurrency,
      'language': config.language,
      'timezone': config.timezone,
      'yeartype': config.yeartype,
      'timezonelabel': config.timezonelabel,
      'timezoneoffset': config.timezoneoffset,
      'dateFormat': config.dateFormat,
      'pos': config.pos != null ? {
        'taxid': config.pos!.taxid,
        'vatrate': config.pos!.vatrate,
        'vattypesale': config.pos!.vattypesale,
        'vattypepurchase': config.pos!.vattypepurchase,
        'headerreceiptpos': config.pos!.headerreceiptpos,
        'footerreceiptpos': config.pos!.footerreceiptpos,
        'isbom': config.pos!.isbom,
      } : null,
      'paymentrounding': config.paymentrounding?.toJson(),
      'pointconfig': config.pointconfig?.toJson(),
      'couponusetype': config.couponusetype,
      'machinetype': config.machinetype,
    };

    final jsonString = jsonEncode(configData);
    final bytes = utf8.encode(jsonString);
    final digest = md5.convert(bytes);
    return digest.toString();
  }

  /// ตรวจสอบการเปลี่ยนแปลง
  Future<void> _checkForChanges() async {
    if (_watchingBranchGuid == null || _watchingBranchGuid!.isEmpty) {
      return;
    }

    try {
      AppLogger.debug('[BranchConfigWatcher] Checking for config changes...');

      // ดึง config จาก server
      final result = await _companyBranchRepository.getBranch(_watchingBranchGuid!);

      if (!result.success || result.data == null) {
        AppLogger.warning('[BranchConfigWatcher] Failed to fetch branch config');
        return;
      }

      final newConfig = CompanyBranchModel.fromJson(result.data);
      final newHash = _calculateConfigHash(newConfig);

      // เปรียบเทียบ hash
      if (_currentConfigHash != null && _currentConfigHash != newHash) {
        AppLogger.info('[BranchConfigWatcher] Config changed! Old hash: $_currentConfigHash, New hash: $newHash');

        // หา fields ที่เปลี่ยนแปลง
        final changedFields = _findChangedFields(global.companyBranchSelectData, newConfig);

        // อัพเดท global config
        final oldConfig = global.companyBranchSelectData;
        global.companyBranchSelectData = newConfig;

        // อัพเดท hash
        _currentConfigHash = newHash;

        // แจ้งเตือน
        _notifyConfigChange(oldConfig, newConfig, changedFields);

        // Callback
        onConfigChanged?.call(newConfig, changedFields);

        // Broadcast event
        _configChangeController.add(BranchConfigChangeEvent(
          oldConfig: oldConfig,
          newConfig: newConfig,
          changedFields: changedFields,
        ));
      }
    } catch (e) {
      AppLogger.error('[BranchConfigWatcher] Error checking config: $e');
    }
  }

  /// หา fields ที่เปลี่ยนแปลง
  List<String> _findChangedFields(CompanyBranchModel oldConfig, CompanyBranchModel newConfig) {
    final changedFields = <String>[];

    if (oldConfig.baseCurrency != newConfig.baseCurrency) {
      changedFields.add(global.language('base_currency'));
    }
    if (oldConfig.language != newConfig.language) {
      changedFields.add(global.language('languages'));
    }
    if (oldConfig.timezone != newConfig.timezone) {
      changedFields.add(global.language('timezone'));
    }
    if (oldConfig.yeartype != newConfig.yeartype) {
      changedFields.add(global.language('yeartype'));
    }
    if (oldConfig.dateFormat != newConfig.dateFormat) {
      changedFields.add(global.language('date_format'));
    }
    if (oldConfig.pos?.vatrate != newConfig.pos?.vatrate) {
      changedFields.add('อัตราภาษีมูลค่าเพิ่ม');
    }
    if (oldConfig.pos?.vattypesale != newConfig.pos?.vattypesale) {
      changedFields.add('ประเภท VAT ขาย');
    }
    if (oldConfig.pos?.vattypepurchase != newConfig.pos?.vattypepurchase) {
      changedFields.add('ประเภท VAT ซื้อ');
    }
    if (oldConfig.couponusetype != newConfig.couponusetype) {
      changedFields.add('ประเภทการใช้คูปอง');
    }
    if (_jsonEncode(oldConfig.paymentrounding) != _jsonEncode(newConfig.paymentrounding)) {
      changedFields.add(global.language('payment_rounding'));
    }
    if (_jsonEncode(oldConfig.pointconfig) != _jsonEncode(newConfig.pointconfig)) {
      changedFields.add('การตั้งค่าแต้ม');
    }

    // ตรวจสอบชื่อสาขา
    if (_namesChanged(oldConfig.names, newConfig.names)) {
      changedFields.add(global.language('company_branch_name'));
    }

    // ตรวจสอบชื่อบริษัท
    if (_namesChanged(oldConfig.companynames, newConfig.companynames)) {
      changedFields.add(global.language('company_name'));
    }

    return changedFields;
  }

  String _jsonEncode(dynamic obj) {
    try {
      return jsonEncode(obj?.toJson());
    } catch (e) {
      return '';
    }
  }

  bool _namesChanged(List<LanguageDataModel> oldNames, List<LanguageDataModel> newNames) {
    if (oldNames.length != newNames.length) return true;
    for (int i = 0; i < oldNames.length; i++) {
      if (oldNames[i].code != newNames[i].code || oldNames[i].name != newNames[i].name) {
        return true;
      }
    }
    return false;
  }

  /// แจ้งเตือนเมื่อ config เปลี่ยน
  void _notifyConfigChange(
    CompanyBranchModel oldConfig,
    CompanyBranchModel newConfig,
    List<String> changedFields,
  ) {
    final branchName = global.activeLangName(newConfig.names);
    final fieldsText = changedFields.isNotEmpty
        ? changedFields.join(', ')
        : global.language("general_settings");

    global.notificationQueue.add(NotificationItem(
      title: global.language("branch_config_changed"),
      message: 'สาขา "$branchName" มีการเปลี่ยนแปลง:\n$fieldsText\n\nระบบได้โหลดการตั้งค่าใหม่แล้ว',
      type: NotificationType.info,
      duration: const Duration(seconds: 8),
      data: {
        'branch_guid': newConfig.guidfixed,
        'branch_name': branchName,
        'changed_fields': changedFields,
      },
    ));
  }

  /// ตรวจสอบทันที (ไม่ต้องรอ polling interval)
  Future<bool> checkNow() async {
    final oldHash = _currentConfigHash;
    await _checkForChanges();
    return oldHash != _currentConfigHash;
  }

  /// Force reload config โดยไม่สนใจ hash
  Future<void> forceReload() async {
    if (_watchingBranchGuid == null || _watchingBranchGuid!.isEmpty) {
      return;
    }

    try {
      final result = await _companyBranchRepository.getBranch(_watchingBranchGuid!);

      if (result.success && result.data != null) {
        final newConfig = CompanyBranchModel.fromJson(result.data);
        global.companyBranchSelectData = newConfig;
        _currentConfigHash = _calculateConfigHash(newConfig);

        AppLogger.info('[BranchConfigWatcher] Force reloaded config');

        global.notificationQueue.addSimple(
          title: global.language("config_reloaded"),
          message: 'โหลดการตั้งค่าสาขาใหม่เรียบร้อยแล้ว',
          type: NotificationType.success,
          duration: const Duration(seconds: 3),
        );
      }
    } catch (e) {
      AppLogger.error('[BranchConfigWatcher] Error force reloading: $e');
    }
  }

  /// สถานะว่ากำลัง polling อยู่หรือไม่
  bool get isWatching => _isPolling;

  /// Branch GUID ที่กำลังติดตาม
  String? get watchingBranchGuid => _watchingBranchGuid;

  /// ปรับ interval
  void setPollingInterval(int seconds) {
    _pollingIntervalSeconds = seconds;
    if (_isPolling && _watchingBranchGuid != null) {
      // Restart with new interval
      startWatching(
        branchGuidFixed: _watchingBranchGuid!,
        intervalSeconds: seconds,
      );
    }
  }

  /// Dispose
  void dispose() {
    _stopPollingTimer();
    _configChangeController.close();
    _watchingBranchGuid = null;
    _currentConfigHash = null;
    _isPolling = false;
  }
}

/// Event สำหรับ broadcast เมื่อ config เปลี่ยน
class BranchConfigChangeEvent {
  final CompanyBranchModel oldConfig;
  final CompanyBranchModel newConfig;
  final List<String> changedFields;

  BranchConfigChangeEvent({
    required this.oldConfig,
    required this.newConfig,
    required this.changedFields,
  });
}

/// Global instance
final branchConfigWatcher = BranchConfigWatcherService();
