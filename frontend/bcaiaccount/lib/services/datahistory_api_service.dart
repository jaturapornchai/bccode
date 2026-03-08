import 'package:dio/dio.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/logger/app_logger.dart';

/// Model สำหรับ Field Change
class FieldChange {
  final String field;
  final dynamic oldValue;
  final dynamic newValue;

  FieldChange({
    required this.field,
    this.oldValue,
    this.newValue,
  });

  factory FieldChange.fromJson(Map<String, dynamic> json) {
    return FieldChange(
      field: json['field'] ?? '',
      oldValue: json['old_value'],
      newValue: json['new_value'],
    );
  }
}

/// Model สำหรับ Data History
class DataHistoryModel {
  final String id;
  final String shopId;
  final String screenType;
  final String action;
  final String docNo;
  final String guidFixed;
  final String userCode;
  final String userName;
  final DateTime timestamp;
  final Map<String, dynamic>? dataBefore;
  final Map<String, dynamic>? dataAfter;
  final List<FieldChange> changes;

  DataHistoryModel({
    required this.id,
    required this.shopId,
    required this.screenType,
    required this.action,
    required this.docNo,
    required this.guidFixed,
    required this.userCode,
    required this.userName,
    required this.timestamp,
    this.dataBefore,
    this.dataAfter,
    required this.changes,
  });

  factory DataHistoryModel.fromJson(Map<String, dynamic> json) {
    List<FieldChange> changesList = [];
    if (json['changes'] != null) {
      changesList = (json['changes'] as List)
          .map((e) => FieldChange.fromJson(e as Map<String, dynamic>))
          .toList();
    }

    return DataHistoryModel(
      id: _parseId(json['_id']),
      shopId: json['shopid'] ?? '',
      screenType: json['screen_type'] ?? '',
      action: json['action'] ?? '',
      docNo: json['docno'] ?? '',
      guidFixed: json['guidfixed'] ?? '',
      userCode: json['user_code'] ?? '',
      userName: json['user_name'] ?? '',
      timestamp: _parseTimestamp(json['timestamp']),
      dataBefore: json['data_before'] as Map<String, dynamic>?,
      dataAfter: json['data_after'] as Map<String, dynamic>?,
      changes: changesList,
    );
  }

  static DateTime _parseTimestamp(dynamic timestamp) {
    if (timestamp == null) return DateTime.now();
    if (timestamp is String) {
      return DateTime.tryParse(timestamp) ?? DateTime.now();
    }
    if (timestamp is Map && timestamp['\$date'] != null) {
      return DateTime.tryParse(timestamp['\$date'].toString()) ?? DateTime.now();
    }
    return DateTime.now();
  }

  static String _parseId(dynamic id) {
    if (id == null) return '';
    if (id is String) return id;
    if (id is Map && id['\$oid'] != null) return id['\$oid'].toString();
    return id.toString();
  }

  /// แปลง action เป็นภาษาไทย
  String get actionText {
    switch (action) {
      case 'create':
        return global.language('import.action.created');
      case 'update':
        return global.language('edit');
      case 'delete':
        return global.language('delete');
      default:
        return action;
    }
  }

  /// สี badge ตาม action
  int get actionColor {
    switch (action) {
      case 'create':
        return 0xFF4CAF50; // Green
      case 'update':
        return 0xFF2196F3; // Blue
      case 'delete':
        return 0xFFF44336; // Red
      default:
        return 0xFF9E9E9E; // Grey
    }
  }
}

/// Service สำหรับดึงข้อมูล Data History
class DataHistoryApiService {
  static final Dio _dio = Dio();

  /// ดึง history ของ PO
  static Future<List<DataHistoryModel>> getPOHistory(String docNo) async {
    try {
      final url = global.goApiUrlPath('api/datahistory/po');
      final response = await _dio.get(
        url,
        queryParameters: {
          'shopid': global.getShopId(),
          'docno': docNo,
        },
      );

      if (response.statusCode == 200 && response.data['success'] == true) {
        final dynamic rawData = response.data['data'];
        // ตรวจสอบว่า data เป็น List หรือไม่
        if (rawData is! List) {
          AppLogger.debug('[DataHistoryApiService] data is not a List, got: ${rawData.runtimeType}');
          return [];
        }
        return rawData
            .map((e) => DataHistoryModel.fromJson(e as Map<String, dynamic>))
            .toList();
      }

      return [];
    } catch (e, stackTrace) {
      AppLogger.error('[DataHistoryApiService] Error fetching PO history: $e\n$stackTrace');
      return [];
    }
  }

  /// ดึง history ทั่วไป (รองรับหลาย screen types)
  static Future<List<DataHistoryModel>> getHistory({
    required String screenType,
    String? docNo,
  }) async {
    try {
      final Map<String, dynamic> params = {
        'shopid': global.getShopId(),
        'screen_type': screenType,
      };
      if (docNo != null) {
        params['docno'] = docNo;
      }

      final url = global.goApiUrlPath('api/datahistory');
      final response = await _dio.get(
        url,
        queryParameters: params,
      );

      if (response.statusCode == 200 && response.data['success'] == true) {
        final dynamic rawData = response.data['data'];
        // ตรวจสอบว่า data เป็น List หรือไม่
        if (rawData is! List) {
          AppLogger.debug('[DataHistoryApiService] data is not a List, got: ${rawData.runtimeType}');
          return [];
        }
        return rawData
            .map((e) => DataHistoryModel.fromJson(e as Map<String, dynamic>))
            .toList();
      }

      return [];
    } catch (e, stackTrace) {
      AppLogger.error('[DataHistoryApiService] Error fetching history: $e\n$stackTrace');
      return [];
    }
  }
}
