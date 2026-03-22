import 'package:flutter/material.dart';
import 'package:smlaicloud/model/approval_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/services/approval_api_service.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/global.dart' as global;

/// Handler สำหรับจัดการ PO Approval และ Purchase Type
///
/// รับผิดชอบ:
/// - โหลด/บันทึกประเภทการจัดซื้อ
/// - โหลด/ส่งสถานะการอนุมัติ
/// - ตรวจสอบสิทธิ์พิมพ์/แก้ไข PO
class POApprovalHandler {
  final VoidCallback onStateChanged;
  final void Function(String) onShowSuccess;
  final void Function(String) onShowInfo;
  final void Function(String) onShowError;

  POApprovalHandler({
    required this.onStateChanged,
    required this.onShowSuccess,
    required this.onShowInfo,
    required this.onShowError,
  });

  // ==================== Purchase Type State ====================
  List<PurchaseTypeModel> _purchaseTypes = [];
  String? _selectedPurchaseTypeCode;
  String? _selectedPurchaseTypeName;

  // ==================== Approval Status State ====================
  POApprovalStatusModel? _poApprovalStatus;
  bool _isLoadingApprovalStatus = false;

  // ==================== Getters ====================
  List<PurchaseTypeModel> get purchaseTypes => _purchaseTypes;
  String? get selectedPurchaseTypeCode => _selectedPurchaseTypeCode;
  String? get selectedPurchaseTypeName => _selectedPurchaseTypeName;
  POApprovalStatusModel? get approvalStatus => _poApprovalStatus;
  bool get isLoadingApprovalStatus => _isLoadingApprovalStatus;

  // ==================== Purchase Type Methods ====================

  /// โหลดรายการประเภทการจัดซื้อสำหรับ PO (เอกสารใหม่)
  Future<void> loadPurchaseTypesForNewPO() async {
    try {
      final result = await ApprovalApiService.getPurchaseTypes();
      if (result.isSuccess) {
        final currentLang = global.appConfig.getString("language") ?? 'th';
        _purchaseTypes = result.items;
        // ถ้ามีประเภทการจัดซื้อ ให้เลือกตัวแรกเป็นค่าเริ่มต้น
        if (_purchaseTypes.isNotEmpty && _selectedPurchaseTypeCode == null) {
          _selectedPurchaseTypeCode = _purchaseTypes.first.code;
          _selectedPurchaseTypeName = _purchaseTypes.first.getName(currentLang);
        }
        onStateChanged();
        AppLogger.info('[PO] โหลดประเภทการจัดซื้อสำเร็จ: ${_purchaseTypes.length} รายการ');
      }
    } catch (e) {
      AppLogger.error('[PO] โหลดประเภทการจัดซื้อผิดพลาด: $e');
    }
  }

  /// โหลด list ประเภทการจัดซื้อ และค่าที่บันทึกไว้ (สำหรับโหลดเอกสารที่มีอยู่)
  Future<void> loadPurchaseTypesAndExisting(TransactionModel screenData) async {
    try {
      // โหลด list ประเภทการจัดซื้อก่อน
      final typesResult = await ApprovalApiService.getPurchaseTypes();
      if (typesResult.isSuccess) {
        _purchaseTypes = typesResult.items;
        AppLogger.info('[PO] โหลดประเภทการจัดซื้อสำเร็จ: ${_purchaseTypes.length} รายการ');
      }

      // อ่านค่าที่บันทึกไว้จาก TransactionModel โดยตรง
      final existingCode = screenData.purchasetypecode;
      final existingNames = screenData.purchasetypenames;

      if (existingCode != null && existingCode.isNotEmpty) {
        // ดึงชื่อจาก purchasetypenames ตามภาษาปัจจุบัน
        String? purchaseTypeName;
        if (existingNames != null && existingNames.isNotEmpty) {
          final currentLang = global.appConfig.getString("language") ?? 'th';
          final nameData = existingNames.firstWhere(
            (n) => n.code == currentLang,
            orElse: () => existingNames.first,
          );
          purchaseTypeName = nameData.name;
        }

        _selectedPurchaseTypeCode = existingCode;
        _selectedPurchaseTypeName = purchaseTypeName ?? '';
        AppLogger.info('[PO] โหลดประเภทการจัดซื้อของ PO สำเร็จ: $purchaseTypeName');
      } else if (_purchaseTypes.isNotEmpty) {
        // ถ้าไม่พบค่าที่บันทึกไว้ ให้ใช้ค่าแรกใน list
        final currentLang = global.appConfig.getString("language") ?? 'th';
        _selectedPurchaseTypeCode = _purchaseTypes.first.code;
        _selectedPurchaseTypeName = _purchaseTypes.first.getName(currentLang);
      }

      onStateChanged();
    } catch (e) {
      AppLogger.error('[PO] โหลดประเภทการจัดซื้อผิดพลาด: $e');
    }
  }

  /// บันทึกประเภทการจัดซื้อลงใน TransactionModel
  void savePurchaseTypeToModel(TransactionModel screenData) {
    if (_selectedPurchaseTypeCode == null || _selectedPurchaseTypeCode!.isEmpty) return;

    // บันทึกค่าลงใน screenData (จะถูกบันทึกไปพร้อมกับเอกสารหลัก)
    screenData.purchasetypecode = _selectedPurchaseTypeCode;

    // สร้าง purchasetypenames (หลายภาษา) จากชื่อที่เลือก
    if (_selectedPurchaseTypeName != null && _selectedPurchaseTypeName!.isNotEmpty) {
      final currentLang = global.appConfig.getString("language") ?? 'th';
      screenData.purchasetypenames = [
        LanguageDataModel(code: currentLang, name: _selectedPurchaseTypeName!),
      ];
    }

    AppLogger.info('[PO] บันทึกประเภทการจัดซื้อลงใน TransactionModel: $_selectedPurchaseTypeCode');
  }

  /// เปลี่ยนประเภทการจัดซื้อที่เลือก
  void onPurchaseTypeChanged(String? code, String? name) {
    _selectedPurchaseTypeCode = code;
    _selectedPurchaseTypeName = name;
    onStateChanged();
  }

  /// รีเซ็ตประเภทการจัดซื้อ (สำหรับเอกสารใหม่)
  void resetPurchaseType() {
    _selectedPurchaseTypeCode = null;
    _selectedPurchaseTypeName = null;
  }

  // ==================== Approval Status Methods ====================

  /// โหลดสถานะการอนุมัติของ PO
  Future<void> loadApprovalStatus(String docNo) async {
    AppLogger.debug('[PO] loadApprovalStatus เริ่มโหลดสถานะสำหรับ docNo: $docNo');
    _isLoadingApprovalStatus = true;
    onStateChanged();

    try {
      final result = await ApprovalApiService.getPOApprovalStatus(docNo);
      AppLogger.debug('[PO] loadApprovalStatus result - isSuccess: ${result.isSuccess}, isFound: ${result.isFound}');

      if (result.isSuccess && result.isFound) {
        _poApprovalStatus = result.data;
        AppLogger.info('[PO] โหลดสถานะการอนุมัติสำเร็จ: ${result.data?.status}');
      } else {
        _poApprovalStatus = null;
        AppLogger.debug('[PO] ไม่พบสถานะการอนุมัติสำหรับ docNo: $docNo');
      }
    } catch (e) {
      AppLogger.error('[PO] โหลดสถานะการอนุมัติผิดพลาด: $e');
      _poApprovalStatus = null;
    } finally {
      _isLoadingApprovalStatus = false;
      onStateChanged();
    }
  }

  /// ส่ง PO เข้าระบบอนุมัติ
  /// [isModified] - true เมื่อเป็นการแก้ไขเอกสารที่เคยส่งอนุมัติแล้ว
  /// คืนค่า true ถ้าเป็น auto_approved (สามารถพิมพ์ได้ทันที)
  Future<bool> submitForApproval({
    required TransactionModel screenData,
    required String description,
    bool isModified = false,
  }) async {
    if (_selectedPurchaseTypeCode == null || _selectedPurchaseTypeCode!.isEmpty) {
      AppLogger.warning('[PO] submitForApproval: purchaseTypeCode ว่าง — return false (ไม่ส่งอนุมัติ)');
      return false;
    }

    try {
      // คำนวณยอดรวม
      final totalAmount = screenData.totalaftervat;

      // ดึงข้อมูล user ปัจจุบัน
      final userCode = global.appConfig.getString("user") ?? '';
      final userName = userCode;
      const approvalLevel = 0;

      // ถ้ามี approval status อยู่แล้ว ถือว่าเป็นการแก้ไขเอกสาร
      final effectiveIsModified = isModified || (_poApprovalStatus != null);

      AppLogger.debug('[PO] Submit approval - isModified: $isModified, hasExistingStatus: ${_poApprovalStatus != null}, effectiveIsModified: $effectiveIsModified');
      AppLogger.debug('[PO] Description (หมายเหตุ): $description');

      // Build items list for LIFF display
      final List<Map<String, dynamic>> itemsList = [];
      if (screenData.details != null) {
        for (int i = 0; i < screenData.details!.length; i++) {
          final detail = screenData.details![i];
          itemsList.add({
            'linenumber': i + 1,
            'itemcode': detail.itemcode,
            'itemname': global.activeLangName(detail.itemnames ?? []),
            'qty': detail.qty,
            'unitname': global.activeLangName(detail.unitnames ?? []),
            'price': detail.price,
            'sumamount': detail.sumamount,
          });
        }
      }

      final result = await ApprovalApiService.submitPOApproval(
        docNo: screenData.docno,
        guidFixed: screenData.guidfixed ?? '',
        purchaseTypeCode: _selectedPurchaseTypeCode!,
        purchaseTypeName: _selectedPurchaseTypeName ?? '',
        totalAmount: totalAmount,
        actionBy: userCode,
        actionByName: userName,
        approvalLevel: approvalLevel,
        isModified: effectiveIsModified,
        comment: description.isNotEmpty ? description : null,
        docDatetime: screenData.docdatetime,
        custCode: screenData.custcode,
        custName: global.activeLangName(screenData.custnames ?? []),
        items: itemsList,
      );

      if (result.isSuccess) {
        _poApprovalStatus = result.data;
        onStateChanged();

        // ตรวจสอบว่าเป็น auto_approved หรือไม่
        final isAutoApproved = result.data?.status == POApprovalStatus.autoApproved;

        AppLogger.info('[PO] submitForApproval สำเร็จ — status: ${result.data?.status}, '
            'isAutoApproved: $isAutoApproved, message: ${result.message}');

        // แสดงข้อความแจ้งเตือน
        final message = result.message ?? global.language('submit_approval_success');
        if (isAutoApproved) {
          onShowSuccess(message);
        } else {
          onShowInfo(message);
        }

        if (result.notificationsSent > 0) {
          AppLogger.info('[PO] ส่งแจ้งเตือนผู้อนุมัติแล้ว ${result.notificationsSent} คน');
        }

        return isAutoApproved;
      } else {
        AppLogger.warning('[PO] submitForApproval ไม่สำเร็จ: ${result.errorMessage}');
        onShowError('ส่งเข้าระบบอนุมัติไม่สำเร็จ: ${result.errorMessage}');
      }
      return false;
    } catch (e) {
      AppLogger.error('[PO] ส่งเข้าระบบอนุมัติผิดพลาด: $e');
      onShowError('ส่งเข้าระบบอนุมัติผิดพลาด: $e');
      return false;
    }
  }

  /// ถอนการส่งอนุมัติ — เฉพาะสถานะ pending เท่านั้น
  /// คืนค่า true ถ้าถอนสำเร็จ
  Future<bool> withdrawApproval({
    required String docNo,
    String comment = '',
  }) async {
    try {
      final userCode = global.appConfig.getString("user") ?? '';
      final userName = userCode;

      AppLogger.info('[PO] withdrawApproval เริ่มถอนการอนุมัติ - docNo: $docNo');

      final result = await ApprovalApiService.withdrawPOApproval(
        docNo: docNo,
        actionBy: userCode,
        actionByName: userName,
        comment: comment,
      );

      if (result.isSuccess) {
        _poApprovalStatus = result.data;
        onStateChanged();

        final message = result.message ?? 'ถอนการส่งอนุมัติสำเร็จ';
        onShowSuccess(message);

        AppLogger.info('[PO] withdrawApproval สำเร็จ — status: ${result.data?.status}');
        return true;
      } else {
        AppLogger.warning('[PO] withdrawApproval ไม่สำเร็จ: ${result.errorMessage}');
        onShowError('ถอนการอนุมัติไม่สำเร็จ: ${result.errorMessage}');
        return false;
      }
    } catch (e) {
      AppLogger.error('[PO] ถอนการอนุมัติผิดพลาด: $e');
      onShowError('ถอนการอนุมัติผิดพลาด: $e');
      return false;
    }
  }

  /// อนุมัติเอกสาร (สำหรับผู้มีสิทธิ์อนุมัติ)
  Future<bool> approveApproval({
    required String docNo,
    String comment = '',
  }) async {
    try {
      final userCode = global.appConfig.getString("user") ?? '';
      final userName = userCode;

      AppLogger.info('[PO] approveApproval เริ่มอนุมัติ - docNo: $docNo');

      final result = await ApprovalApiService.approvePOApproval(
        docNo: docNo,
        actionBy: userCode,
        actionByName: userName,
        comment: comment,
      );

      if (result.isSuccess) {
        _poApprovalStatus = result.data;
        onStateChanged();

        final message = result.message ?? 'อนุมัติเอกสารสำเร็จ';
        onShowSuccess(message);

        AppLogger.info('[PO] approveApproval สำเร็จ — status: ${result.data?.status}');
        return true;
      } else {
        AppLogger.warning('[PO] approveApproval ไม่สำเร็จ: ${result.errorMessage}');
        onShowError('อนุมัติไม่สำเร็จ: ${result.errorMessage}');
        return false;
      }
    } catch (e) {
      AppLogger.error('[PO] อนุมัติผิดพลาด: $e');
      onShowError('อนุมัติผิดพลาด: $e');
      return false;
    }
  }

  /// ปฏิเสธเอกสารพร้อมเหตุผล (สำหรับผู้มีสิทธิ์อนุมัติ)
  Future<bool> rejectApproval({
    required String docNo,
    String comment = '',
  }) async {
    try {
      final userCode = global.appConfig.getString("user") ?? '';
      final userName = userCode;

      AppLogger.info('[PO] rejectApproval เริ่มปฏิเสธ - docNo: $docNo, comment: $comment');

      final result = await ApprovalApiService.rejectPOApproval(
        docNo: docNo,
        actionBy: userCode,
        actionByName: userName,
        comment: comment,
      );

      if (result.isSuccess) {
        _poApprovalStatus = result.data;
        onStateChanged();

        final message = result.message ?? 'ปฏิเสธเอกสารแล้ว';
        onShowSuccess(message);

        AppLogger.info('[PO] rejectApproval สำเร็จ — status: ${result.data?.status}');
        return true;
      } else {
        AppLogger.warning('[PO] rejectApproval ไม่สำเร็จ: ${result.errorMessage}');
        onShowError('ปฏิเสธไม่สำเร็จ: ${result.errorMessage}');
        return false;
      }
    } catch (e) {
      AppLogger.error('[PO] ปฏิเสธผิดพลาด: $e');
      onShowError('ปฏิเสธผิดพลาด: $e');
      return false;
    }
  }

  /// รีเซ็ตสถานะการอนุมัติ (สำหรับเอกสารใหม่)
  void resetApprovalStatus() {
    _poApprovalStatus = null;
    _isLoadingApprovalStatus = false;
  }
}
