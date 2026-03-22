import 'dart:convert';
import 'package:http/http.dart' as http;
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/approval_model.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/environment.dart';

/// Approval API Service
/// บริการเชื่อมต่อระบบอนุมัติผ่าน Go Backend API (MongoDB Atlas)
/// NOTE: purchase-type ถูกย้ายไป mainapi แล้ว - ใช้ _mainApiBaseUrl สำหรับ purchase-type
class ApprovalApiService {
  /// Base URL สำหรับ goapi (approval settings, status)
  static String get _baseUrl => global.goApiBaseUrl;

  /// Base URL สำหรับ mainapi (purchase-type)
  /// ใช้ Environment().config.serviceApi โดยตรง เพราะ myAppConfig.serviceApi ชี้ไป goapi (reportApiPath)
  static String get _mainApiBaseUrl {
    final env = Environment();
    String apiPath = env.config.serviceApi;
    // Remove trailing slash if present
    if (apiPath.endsWith('/')) {
      apiPath = apiPath.substring(0, apiPath.length - 1);
    }
    return apiPath;
  }

  /// Headers สำหรับ API calls (รวม Authorization token)
  static Map<String, String> get _headers {
    final headers = <String, String>{'Content-Type': 'application/json'};
    final token = global.appConfig.getString("token") ?? '';
    if (token.isNotEmpty) {
      headers['Authorization'] = 'Bearer $token';
    }
    return headers;
  }

  /// ดึงรายการประเภทการจัดซื้อทั้งหมด (จาก mainapi)
  static Future<PurchaseTypeListResult> getPurchaseTypes() async {
    final httpClient = http.Client();
    try {
      // ใช้ mainapi endpoint: GET /purchase-type
      final url = Uri.parse("$_mainApiBaseUrl/purchase-type?limit=1000");

      AppLogger.info('[ApprovalAPI] GET $url');

      final response = await httpClient.get(
        url,
        headers: _headers,
      );

      AppLogger.info('[ApprovalAPI] Response status: ${response.statusCode}');
      AppLogger.info('[ApprovalAPI] Response body: ${response.body}');

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['success'] == true) {
          return PurchaseTypeListResult.fromJson(json);
        }
      }

      return PurchaseTypeListResult.error(
        "API call failed: ${response.statusCode} - ${response.reasonPhrase}",
      );
    } catch (e, stackTrace) {
      AppLogger.error('[ApprovalAPI] Error: $e', stackTrace: stackTrace);
      return PurchaseTypeListResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }

  /// บันทึกประเภทการจัดซื้อ (จาก mainapi)
  /// ถ้า guid ว่าง = POST create, ถ้ามี guid = PUT update
  static Future<ApprovalSaveResult> savePurchaseType(
      PurchaseTypeModel model) async {
    final httpClient = http.Client();
    try {
      final bool isUpdate = model.guid != null && model.guid!.isNotEmpty;
      final Uri url;

      if (isUpdate) {
        // PUT /purchase-type/:id
        url = Uri.parse("$_mainApiBaseUrl/purchase-type/${model.guid}");
      } else {
        // POST /purchase-type
        url = Uri.parse("$_mainApiBaseUrl/purchase-type");
      }

      final body = model.toJson();

      AppLogger.info('[ApprovalAPI] ${isUpdate ? "PUT" : "POST"} $url');
      AppLogger.info('[ApprovalAPI] Body: $body');

      final http.Response response;
      if (isUpdate) {
        response = await httpClient.put(
          url,
          headers: _headers,
          body: jsonEncode(body),
        );
      } else {
        response = await httpClient.post(
          url,
          headers: _headers,
          body: jsonEncode(body),
        );
      }

      AppLogger.info('[ApprovalAPI] Response: ${response.body}');

      if (response.statusCode == 200 || response.statusCode == 201) {
        final json = jsonDecode(response.body);
        if (json['success'] == true) {
          return ApprovalSaveResult.success(json['id'] ?? model.guid ?? '');
        }
        return ApprovalSaveResult.error(json['message'] ?? 'Unknown error');
      }

      return ApprovalSaveResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      AppLogger.error('[ApprovalAPI] Save error: $e');
      return ApprovalSaveResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }

  /// ลบประเภทการจัดซื้อ (จาก mainapi)
  /// [guid] - guidfixed ของ purchase type ที่ต้องการลบ
  static Future<bool> deletePurchaseType(String guid) async {
    final httpClient = http.Client();
    try {
      // DELETE /purchase-type/:id
      final url = Uri.parse("$_mainApiBaseUrl/purchase-type/$guid");

      AppLogger.info('[ApprovalAPI] DELETE purchase type: $guid');

      final response = await httpClient.delete(
        url,
        headers: _headers,
      );

      AppLogger.info('[ApprovalAPI] Response: ${response.body}');

      final json = jsonDecode(response.body);
      return response.statusCode == 200 && json['success'] == true;
    } catch (e) {
      AppLogger.error('[ApprovalAPI] Delete error: $e');
      return false;
    } finally {
      httpClient.close();
    }
  }

  /// ดึงรายการการตั้งค่าอนุมัติใบสั่งซื้อทั้งหมด
  static Future<POApprovalSettingListResult> getPOApprovalSettings() async {
    final httpClient = http.Client();
    try {
      final shopId = global.getShopId();
      final url = Uri.parse("$_baseUrl/api/approval/po-settings");

      final body = {"shop_id": shopId};

      AppLogger.info('[ApprovalAPI] POST $url');

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      AppLogger.info('[ApprovalAPI] Response: ${response.body}');

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['success'] == true) {
          return POApprovalSettingListResult.fromJson(json);
        }
      }

      return POApprovalSettingListResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      AppLogger.error('[ApprovalAPI] Get PO settings error: $e');
      return POApprovalSettingListResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }

  /// ดึงการตั้งค่าอนุมัติใบสั่งซื้อตามประเภทการจัดซื้อ
  static Future<POApprovalSettingResult> getPOApprovalSettingByPurchaseType(
      String purchaseTypeCode) async {
    final httpClient = http.Client();
    try {
      final shopId = global.getShopId();
      final url = Uri.parse("$_baseUrl/api/approval/po-setting");

      final body = {
        "shop_id": shopId,
        "purchase_type_code": purchaseTypeCode,
      };

      AppLogger.info('[ApprovalAPI] POST $url');

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      AppLogger.info('[ApprovalAPI] Response: ${response.body}');

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['success'] == true) {
          if (json['found'] == true) {
            return POApprovalSettingResult.fromJson(json);
          } else {
            return POApprovalSettingResult.notFound();
          }
        }
      }

      return POApprovalSettingResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      AppLogger.error('[ApprovalAPI] Get PO setting error: $e');
      return POApprovalSettingResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }

  /// บันทึกการตั้งค่าอนุมัติใบสั่งซื้อ
  static Future<ApprovalSaveResult> savePOApprovalSetting(
      POApprovalSettingModel model) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/approval/po-setting/save");

      final body = {
        "shop_id": global.getShopId(),
        ...model.toJson(),
      };

      AppLogger.info('[ApprovalAPI] POST $url');
      AppLogger.info('[ApprovalAPI] Body: $body');

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      AppLogger.info('[ApprovalAPI] Response: ${response.body}');

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['success'] == true) {
          return ApprovalSaveResult.success(json['data']?['guid'] ?? '');
        }
        return ApprovalSaveResult.error(json['message'] ?? 'Unknown error');
      }

      return ApprovalSaveResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      AppLogger.error('[ApprovalAPI] Save PO setting error: $e');
      return ApprovalSaveResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }

  /// ลบการตั้งค่าอนุมัติใบสั่งซื้อ
  static Future<bool> deletePOApprovalSetting(String purchaseTypeCode) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/approval/po-setting/delete");

      final body = {
        "shop_id": global.getShopId(),
        "purchase_type_code": purchaseTypeCode,
      };

      AppLogger.info('[ApprovalAPI] DELETE PO setting: $purchaseTypeCode');

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      final json = jsonDecode(response.body);
      return response.statusCode == 200 && json['success'] == true;
    } catch (e) {
      AppLogger.error('[ApprovalAPI] Delete PO setting error: $e');
      return false;
    } finally {
      httpClient.close();
    }
  }

  // NOTE: savePODocumentType และ getPODocumentType ถูกลบแล้ว
  // เพราะ purchasetypecode และ purchasetypenames เก็บใน TransactionHeader แทน

  // =====================================================
  // PO Approval Status (สถานะการอนุมัติ PO)
  // =====================================================

  /// ดึงสถานะการอนุมัติของ PO
  static Future<POApprovalStatusResult> getPOApprovalStatus(String docNo) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/approval/po-status/get");

      final body = {
        "shop_id": global.getShopId(),
        "docno": docNo,
      };

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['success'] == true) {
          if (json['found'] == true) {
            return POApprovalStatusResult.fromJson(json);
          } else {
            return POApprovalStatusResult.notFound();
          }
        }
      }

      return POApprovalStatusResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      AppLogger.error('[ApprovalAPI] Get PO approval status error: $e');
      return POApprovalStatusResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }

  /// ดึง Timeline ประวัติการอนุมัติแบบละเอียด (รวม notification sent/opened)
  static Future<ApprovalTimelineResult> getApprovalTimeline(String docNo) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/approval/timeline");

      final body = {
        "shop_id": global.getShopId(),
        "docno": docNo,
      };

      AppLogger.debug('[ApprovalAPI] Timeline request URL: $url');
      AppLogger.debug('[ApprovalAPI] Timeline request body: $body');

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      AppLogger.debug('[ApprovalAPI] Timeline response status: ${response.statusCode}');
      AppLogger.debug('[ApprovalAPI] Timeline response body: ${response.body}');

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['success'] == true) {
          return ApprovalTimelineResult.fromJson(json);
        }
      }

      return ApprovalTimelineResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      AppLogger.error('[ApprovalAPI] Get approval timeline error: $e');
      return ApprovalTimelineResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }

  /// ดึงสถานะการอนุมัติของ PO หลายเอกสารพร้อมกัน
  /// คืนค่า Map<docNo, POApprovalStatusInfo>
  static Future<Map<String, POApprovalStatusInfo>> getBatchPOApprovalStatus(List<String> docNos) async {
    if (docNos.isEmpty) return {};

    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/approval/po-status/batch");

      final body = {
        "shop_id": global.getShopId(),
        "docnos": docNos,
      };

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['success'] == true && json['data'] != null) {
          final Map<String, POApprovalStatusInfo> result = {};
          final data = json['data'] as Map<String, dynamic>;
          for (final entry in data.entries) {
            final statusData = entry.value as Map<String, dynamic>;
            // Parse latest_action_at (DateTime)
            DateTime? latestActionAt;
            if (statusData['latest_action_at'] != null) {
              try {
                latestActionAt = DateTime.parse(statusData['latest_action_at']);
              } catch (_) {}
            }
            result[entry.key] = POApprovalStatusInfo(
              status: statusData['status'] ?? '',
              requiredLevel: statusData['required_level'] ?? 0,
              requiredLevelName: statusData['required_level_name'] ?? '',
              currentApprovedLevel: statusData['current_approved_level'] ?? 0,
              latestAction: statusData['latest_action'],
              latestActionText: statusData['latest_action_text'],
              latestActionAt: latestActionAt,
              latestActionBy: statusData['latest_action_by'],
            );
          }
          return result;
        }
      }
      AppLogger.warning('[ApprovalAPI] Batch API returned no data or failed');
      return {};
    } catch (e) {
      AppLogger.error('[ApprovalAPI] Get batch PO approval status error: $e');
      return {};
    } finally {
      httpClient.close();
    }
  }

  /// ส่ง PO เข้าระบบอนุมัติ
  /// [isModified] - true เมื่อเป็นการแก้ไขเอกสารที่เคยส่งอนุมัติแล้ว
  /// [docDatetime] - วันที่เอกสาร (สำหรับแสดงใน LIFF)
  /// [custCode] - รหัสลูกค้า/Supplier (สำหรับแสดงใน LIFF)
  /// [custName] - ชื่อลูกค้า/Supplier (สำหรับแสดงใน LIFF)
  /// [items] - รายการสินค้า (สำหรับแสดงใน LIFF)
  static Future<POApprovalSubmitResult> submitPOApproval({
    required String docNo,
    required String guidFixed,
    required String purchaseTypeCode,
    required String purchaseTypeName,
    required double totalAmount,
    required String actionBy,
    required String actionByName,
    required int approvalLevel,
    String? sourceDocNo,
    String? sourceGuidFixed,
    String? comment,
    bool isModified = false,
    // Transaction details for LIFF display
    String? docDatetime,
    String? custCode,
    String? custName,
    List<Map<String, dynamic>>? items,
  }) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/approval/po-status/submit");

      final body = {
        "shop_id": global.getShopId(),
        "docno": docNo,
        "guidfixed": guidFixed,
        "purchase_type_code": purchaseTypeCode,
        "purchase_type_name": purchaseTypeName,
        "total_amount": totalAmount,
        "action_by": actionBy,
        "action_by_name": actionByName,
        "approval_level": approvalLevel,
        "is_modified": isModified,
        if (sourceDocNo != null) "source_docno": sourceDocNo,
        if (sourceGuidFixed != null) "source_guidfixed": sourceGuidFixed,
        if (comment != null) "comment": comment,
        // Transaction details for LIFF display
        if (docDatetime != null) "docdatetime": docDatetime,
        if (custCode != null) "custcode": custCode,
        if (custName != null) "custname": custName,
        if (items != null) "items": items,
      };

      AppLogger.info('[ApprovalAPI] POST $url');
      AppLogger.debug('[ApprovalAPI] Body: $body');

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      AppLogger.info('[ApprovalAPI] Response: ${response.body}');

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['success'] == true) {
          return POApprovalSubmitResult.fromJson(json);
        }
        return POApprovalSubmitResult.error(json['message'] ?? 'Unknown error');
      }

      return POApprovalSubmitResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      AppLogger.error('[ApprovalAPI] Submit PO approval error: $e');
      return POApprovalSubmitResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }

  /// ถอนการส่งอนุมัติ PO (เปลี่ยนสถานะจาก pending → rejected)
  /// เฉพาะผู้สร้างเอกสาร หรือผู้มีสิทธิ์อนุมัติเท่านั้นที่ถอนได้
  static Future<WithdrawApprovalResult> withdrawPOApproval({
    required String docNo,
    required String actionBy,
    required String actionByName,
    String comment = '',
  }) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/approval/po-status/withdraw");

      final body = {
        "shop_id": global.getShopId(),
        "docno": docNo,
        "action_by": actionBy,
        "action_by_name": actionByName,
        "comment": comment,
      };

      AppLogger.info('[ApprovalAPI] POST $url');
      AppLogger.debug('[ApprovalAPI] Body: $body');

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      AppLogger.info('[ApprovalAPI] Response: ${response.body}');

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['success'] == true) {
          return WithdrawApprovalResult.fromJson(json);
        }
        return WithdrawApprovalResult.error(json['message'] ?? 'Unknown error');
      }

      if (response.statusCode == 403) {
        final json = jsonDecode(response.body);
        return WithdrawApprovalResult.error(json['message'] ?? 'ไม่มีสิทธิ์ถอนการอนุมัติ');
      }

      return WithdrawApprovalResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      AppLogger.error('[ApprovalAPI] Withdraw PO approval error: $e');
      return WithdrawApprovalResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }

  /// อนุมัติ PO (สำหรับผู้มีสิทธิ์อนุมัติ)
  static Future<WithdrawApprovalResult> approvePOApproval({
    required String docNo,
    required String actionBy,
    required String actionByName,
    String comment = '',
  }) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/approval/po-status/approve");

      final body = {
        "shop_id": global.getShopId(),
        "docno": docNo,
        "action_by": actionBy,
        "action_by_name": actionByName,
        "comment": comment,
      };

      AppLogger.info('[ApprovalAPI] POST $url');
      AppLogger.debug('[ApprovalAPI] Body: $body');

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      AppLogger.info('[ApprovalAPI] Response: ${response.body}');

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['success'] == true) {
          return WithdrawApprovalResult.fromJson(json);
        }
        return WithdrawApprovalResult.error(json['message'] ?? 'Unknown error');
      }

      if (response.statusCode == 403) {
        final json = jsonDecode(response.body);
        return WithdrawApprovalResult.error(json['message'] ?? 'ไม่มีสิทธิ์อนุมัติ');
      }

      return WithdrawApprovalResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      AppLogger.error('[ApprovalAPI] Approve PO error: $e');
      return WithdrawApprovalResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }

  /// ปฏิเสธ PO พร้อมเหตุผล (สำหรับผู้มีสิทธิ์อนุมัติ)
  static Future<WithdrawApprovalResult> rejectPOApproval({
    required String docNo,
    required String actionBy,
    required String actionByName,
    String comment = '',
  }) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/approval/po-status/reject");

      final body = {
        "shop_id": global.getShopId(),
        "docno": docNo,
        "action_by": actionBy,
        "action_by_name": actionByName,
        "comment": comment,
      };

      AppLogger.info('[ApprovalAPI] POST $url');
      AppLogger.debug('[ApprovalAPI] Body: $body');

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      AppLogger.info('[ApprovalAPI] Response: ${response.body}');

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        if (json['success'] == true) {
          return WithdrawApprovalResult.fromJson(json);
        }
        return WithdrawApprovalResult.error(json['message'] ?? 'Unknown error');
      }

      if (response.statusCode == 403) {
        final json = jsonDecode(response.body);
        return WithdrawApprovalResult.error(json['message'] ?? 'ไม่มีสิทธิ์ปฏิเสธ');
      }

      return WithdrawApprovalResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      AppLogger.error('[ApprovalAPI] Reject PO error: $e');
      return WithdrawApprovalResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }

  /// ส่งคำเตือนผู้อนุมัติซ้ำ (สำหรับ PO ที่รออนุมัติอยู่)
  /// จะส่ง notification ใหม่แม้จะเคยส่งไปแล้ว
  static Future<ResendNotificationResult> resendApprovalNotification({
    required String docNo,
  }) async {
    final httpClient = http.Client();
    try {
      final url = Uri.parse("$_baseUrl/api/approval/notification/resend");

      final body = {
        "shop_id": global.getShopId(),
        "docno": docNo,
      };

      AppLogger.info('[ApprovalAPI] POST $url');
      AppLogger.debug('[ApprovalAPI] Body: $body');

      final response = await httpClient.post(
        url,
        headers: _headers,
        body: jsonEncode(body),
      );

      AppLogger.info('[ApprovalAPI] Response: ${response.body}');

      if (response.statusCode == 200) {
        final json = jsonDecode(response.body);
        return ResendNotificationResult.fromJson(json);
      }

      return ResendNotificationResult.error(
        "API call failed: ${response.reasonPhrase}",
      );
    } catch (e) {
      AppLogger.error('[ApprovalAPI] Resend notification error: $e');
      return ResendNotificationResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }
}

// =====================================================
// Resend Notification Result
// =====================================================

/// ผลลัพธ์การส่งคำเตือนซ้ำ
class ResendNotificationResult {
  final bool isSuccess;
  final String? errorMessage;
  final String? message;
  final int sent;
  final int emailSent;
  final int lineSent;

  ResendNotificationResult({
    required this.isSuccess,
    this.errorMessage,
    this.message,
    this.sent = 0,
    this.emailSent = 0,
    this.lineSent = 0,
  });

  factory ResendNotificationResult.fromJson(Map<String, dynamic> json) {
    return ResendNotificationResult(
      isSuccess: json['success'] ?? false,
      message: json['message'],
      sent: json['sent'] ?? 0,
      emailSent: json['email_sent'] ?? 0,
      lineSent: json['line_sent'] ?? 0,
    );
  }

  factory ResendNotificationResult.error(String message) => ResendNotificationResult(
        isSuccess: false,
        errorMessage: message,
      );
}

// =====================================================
// Withdraw Approval Result
// =====================================================

/// ผลลัพธ์การถอนการส่งอนุมัติ PO
class WithdrawApprovalResult {
  final bool isSuccess;
  final String? errorMessage;
  final String? message;
  final POApprovalStatusModel? data;

  WithdrawApprovalResult({
    required this.isSuccess,
    this.errorMessage,
    this.message,
    this.data,
  });

  factory WithdrawApprovalResult.fromJson(Map<String, dynamic> json) {
    return WithdrawApprovalResult(
      isSuccess: true,
      message: json['message'],
      data: json['data'] != null
          ? POApprovalStatusModel.fromJson(json['data'])
          : null,
    );
  }

  factory WithdrawApprovalResult.error(String message) => WithdrawApprovalResult(
        isSuccess: false,
        errorMessage: message,
      );
}

// =====================================================
// PO Approval Submit Result
// =====================================================

/// ผลลัพธ์การส่ง PO เข้าระบบอนุมัติ
class POApprovalSubmitResult {
  final bool isSuccess;
  final String? errorMessage;
  final String? message;
  final POApprovalStatusModel? data;
  final int notificationsSent;

  POApprovalSubmitResult({
    required this.isSuccess,
    this.errorMessage,
    this.message,
    this.data,
    this.notificationsSent = 0,
  });

  factory POApprovalSubmitResult.fromJson(Map<String, dynamic> json) {
    return POApprovalSubmitResult(
      isSuccess: true,
      message: json['message'],
      data: json['data'] != null
          ? POApprovalStatusModel.fromJson(json['data'])
          : null,
      notificationsSent: json['notifications_sent'] ?? 0,
    );
  }

  factory POApprovalSubmitResult.error(String message) => POApprovalSubmitResult(
        isSuccess: false,
        errorMessage: message,
      );
}

// =====================================================
// PO Approval Status Result
// =====================================================

/// ผลลัพธ์สถานะการอนุมัติ PO
class POApprovalStatusResult {
  final bool isSuccess;
  final bool isFound;
  final String? errorMessage;
  final POApprovalStatusModel? data;

  POApprovalStatusResult({
    required this.isSuccess,
    this.isFound = false,
    this.errorMessage,
    this.data,
  });

  factory POApprovalStatusResult.fromJson(Map<String, dynamic> json) {
    return POApprovalStatusResult(
      isSuccess: true,
      isFound: json['found'] == true,
      data: json['data'] != null
          ? POApprovalStatusModel.fromJson(json['data'])
          : null,
    );
  }

  factory POApprovalStatusResult.notFound() => POApprovalStatusResult(
        isSuccess: true,
        isFound: false,
      );

  factory POApprovalStatusResult.error(String message) => POApprovalStatusResult(
        isSuccess: false,
        errorMessage: message,
      );
}

// =====================================================
// PO Approval Status Info (สำหรับ Batch API)
// =====================================================

/// ข้อมูลสถานะการอนุมัติ PO แบบย่อ (สำหรับแสดงในรายการ)
class POApprovalStatusInfo {
  final String status;
  final int requiredLevel;
  final String requiredLevelName;
  final int currentApprovedLevel;
  // Latest action info
  final String? latestAction;
  final String? latestActionText;
  final DateTime? latestActionAt;
  final String? latestActionBy;

  POApprovalStatusInfo({
    required this.status,
    required this.requiredLevel,
    required this.requiredLevelName,
    required this.currentApprovedLevel,
    this.latestAction,
    this.latestActionText,
    this.latestActionAt,
    this.latestActionBy,
  });

  /// แปลงสถานะเป็นข้อความภาษาไทย
  String get statusText {
    switch (status) {
      case 'pending':
        return global.language('pending_approval');
      case 'approved':
        return global.language('approval_approved');
      case 'rejected':
        return global.language('reject');
      case 'auto_approved':
        return global.language('approval_auto_approved');
      default:
        return '';
    }
  }

  /// สีสำหรับแสดงสถานะ
  int get statusColorValue {
    switch (status) {
      case 'pending':
        return 0xFFFFA000; // Orange
      case 'approved':
        return 0xFF4CAF50; // Green
      case 'rejected':
        return 0xFFF44336; // Red
      case 'auto_approved':
        return 0xFF2196F3; // Blue
      default:
        return 0xFF9E9E9E; // Grey
    }
  }
}

// =====================================================
// Users for Approvers Selection
// =====================================================

/// ดึงรายชื่อผู้ใช้ทั้งหมดสำหรับเลือกผู้อนุมัติ
class ApproverUserListResult {
  final bool isSuccess;
  final String? errorMessage;
  final List<ApproverInfoModel> items;

  ApproverUserListResult({
    required this.isSuccess,
    this.errorMessage,
    this.items = const [],
  });

  factory ApproverUserListResult.success(List<ApproverInfoModel> items) =>
      ApproverUserListResult(isSuccess: true, items: items);

  factory ApproverUserListResult.error(String message) =>
      ApproverUserListResult(isSuccess: false, errorMessage: message);
}

/// Extension สำหรับดึงรายชื่อผู้ใช้
class ApprovalApiServiceUsers {
  /// ดึงรายชื่อผู้ใช้ทั้งหมดสำหรับเลือกผู้อนุมัติ (รวม LINE และ Email info)
  static Future<ApproverUserListResult> getApproversUserList() async {
    final httpClient = http.Client();
    try {
      final shopId = global.getShopId();

      // 1. ดึงรายการผู้ใช้จาก /shop/users API
      // ใช้ Environment.config.serviceApi แทน myAppConfig.serviceApi
      // เนื่องจาก myAppConfig อาจถูก override เป็น localhost ในโหมด dev
      // แต่ /shop/users อยู่ที่ Main SML Cloud API
      final env = Environment();
      String apiPath = env.config.serviceApi;
      AppLogger.debug('[ApprovalAPI] Cloud serviceApi: $apiPath');

      if (!apiPath.startsWith('http')) {
        apiPath = 'https://$apiPath';
      }
      if (!apiPath.endsWith('/')) {
        apiPath += '/';
      }

      final userUrl = Uri.parse(
        "${apiPath}shop/users?offset=0&limit=1000&q=&sort=code:1",
      );

      AppLogger.info('[ApprovalAPI] GET $userUrl');
      AppLogger.debug('[ApprovalAPI] Token: ${global.appConfig.getString("token")?.substring(0, 20) ?? "NO TOKEN"}...');

      // ดึง token จาก global.appConfig เหมือนกับ client.dart
      final token = global.appConfig.getString("token") ?? '';

      final userResponse = await httpClient.get(
        userUrl,
        headers: {
          'Content-Type': 'application/json',
          if (token.isNotEmpty) 'Authorization': 'Bearer $token',
        },
      );

      AppLogger.info('[ApprovalAPI] User list status: ${userResponse.statusCode}');

      if (userResponse.statusCode != 200) {
        return ApproverUserListResult.error(
          "Failed to fetch users: ${userResponse.statusCode}",
        );
      }

      final userData = jsonDecode(userResponse.body);
      AppLogger.debug('[ApprovalAPI] Response body: ${userResponse.body.length > 500 ? userResponse.body.substring(0, 500) : userResponse.body}');

      // Response จาก /shop/users มี format: { "success": true, "data": [...] }
      final userList = userData['data'] as List? ?? [];

      AppLogger.info('[ApprovalAPI] Found ${userList.length} users');

      // Debug: แสดง user แรกถ้ามี
      if (userList.isNotEmpty) {
        AppLogger.debug('[ApprovalAPI] First user: ${userList.first}');
      }

      // 2. แปลงเป็น ApproverInfoModel
      final List<ApproverInfoModel> approvers = [];

      for (final user in userList) {
        final userCode = user['username'] ?? '';
        final userName = user['editusername'] ?? user['username'] ?? '';
        final email = user['email'] ?? '';

        // ข้อมูลจาก user model (ใช้ snake_case ตาม JSON response)
        String? position = user['position'];
        String? department = user['department'];
        // LINE fields ใช้ snake_case ตาม JsonKey ใน UserModel
        String? lineUserId = user['line_user_id'];
        String? lineDisplayName = user['line_display_name'];

        // ถ้าไม่มี LINE info ใน user model ลองดึงจาก LINE OA API
        if ((lineUserId == null || lineUserId.isEmpty) && userCode.isNotEmpty) {
          try {
            final lineUrl = Uri.parse(
              'https://dev-api.bcaicloud.com/liff/api/link?employee_code=$userCode&shop_id=$shopId',
            );
            final lineResponse = await httpClient.get(lineUrl).timeout(
              const Duration(seconds: 3),
            );

            if (lineResponse.statusCode == 200) {
              final lineData = jsonDecode(lineResponse.body);
              if (lineData['success'] == true && lineData['linked'] == true) {
                lineUserId = lineData['data']['userId'];
                lineDisplayName = lineData['data']['displayName'];
              }
            }
          } catch (e) {
            // ข้ามถ้าดึง LINE ไม่ได้ (timeout หรือ error)
          }
        }

        approvers.add(ApproverInfoModel(
          userCode: userCode,
          userName: userName,
          email: email,
          lineUserId: lineUserId,
          lineDisplayName: lineDisplayName,
          position: position,
          department: department,
        ));
      }

      return ApproverUserListResult.success(approvers);
    } catch (e, stackTrace) {
      AppLogger.error('[ApprovalAPI] Get approvers error: $e',
          stackTrace: stackTrace);
      return ApproverUserListResult.error("Connection failed: $e");
    } finally {
      httpClient.close();
    }
  }
}

// NOTE: PODocumentTypeResult class ถูกลบแล้ว - ไม่ใช้แล้วเพราะเก็บใน TransactionHeader

/// ผลลัพธ์รายการประเภทการจัดซื้อ
class PurchaseTypeListResult {
  final bool isSuccess;
  final String? errorMessage;
  final List<PurchaseTypeModel> items;

  PurchaseTypeListResult({
    required this.isSuccess,
    this.errorMessage,
    this.items = const [],
  });

  factory PurchaseTypeListResult.fromJson(Map<String, dynamic> json) {
    final dataList = (json['data'] as List?)
            ?.map((item) => PurchaseTypeModel.fromJson(item))
            .toList() ??
        [];

    return PurchaseTypeListResult(
      isSuccess: true,
      items: dataList,
    );
  }

  factory PurchaseTypeListResult.error(String message) =>
      PurchaseTypeListResult(
        isSuccess: false,
        errorMessage: message,
      );
}

/// ผลลัพธ์การบันทึก
class ApprovalSaveResult {
  final bool isSuccess;
  final String? errorMessage;
  final String? guid;

  ApprovalSaveResult({
    required this.isSuccess,
    this.errorMessage,
    this.guid,
  });

  factory ApprovalSaveResult.success(String guid) => ApprovalSaveResult(
        isSuccess: true,
        guid: guid,
      );

  factory ApprovalSaveResult.error(String message) => ApprovalSaveResult(
        isSuccess: false,
        errorMessage: message,
      );
}

/// ผลลัพธ์รายการการตั้งค่าอนุมัติ PO
class POApprovalSettingListResult {
  final bool isSuccess;
  final String? errorMessage;
  final List<POApprovalSettingModel> items;

  POApprovalSettingListResult({
    required this.isSuccess,
    this.errorMessage,
    this.items = const [],
  });

  factory POApprovalSettingListResult.fromJson(Map<String, dynamic> json) {
    final dataList = (json['data'] as List?)
            ?.map((item) => POApprovalSettingModel.fromJson(item))
            .toList() ??
        [];

    return POApprovalSettingListResult(
      isSuccess: true,
      items: dataList,
    );
  }

  factory POApprovalSettingListResult.error(String message) =>
      POApprovalSettingListResult(
        isSuccess: false,
        errorMessage: message,
      );
}

/// ผลลัพธ์การตั้งค่าอนุมัติ PO เดียว
class POApprovalSettingResult {
  final bool isSuccess;
  final bool isFound;
  final String? errorMessage;
  final POApprovalSettingModel? setting;

  POApprovalSettingResult({
    required this.isSuccess,
    this.isFound = false,
    this.errorMessage,
    this.setting,
  });

  factory POApprovalSettingResult.fromJson(Map<String, dynamic> json) {
    return POApprovalSettingResult(
      isSuccess: true,
      isFound: true,
      setting: POApprovalSettingModel.fromJson(json['data']),
    );
  }

  factory POApprovalSettingResult.notFound() => POApprovalSettingResult(
        isSuccess: true,
        isFound: false,
      );

  factory POApprovalSettingResult.error(String message) =>
      POApprovalSettingResult(
        isSuccess: false,
        errorMessage: message,
      );
}
