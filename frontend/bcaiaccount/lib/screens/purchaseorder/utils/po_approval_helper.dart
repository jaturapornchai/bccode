import 'package:flutter/material.dart';
import 'package:smlaicloud/model/approval_model.dart';
import 'package:smlaicloud/global.dart' as global;

/// Helper class สำหรับจัดการ PO Approval Logic
/// แยกออกมาจาก PurchaseOrderEditScreen เพื่อความสะอาดและ reusability
class POApprovalHelper {
  /// ตรวจสอบว่า PO สามารถพิมพ์ได้หรือไม่
  /// - อนุมัติอัตโนมัติ (auto_approved): พิมพ์ได้ทันทีหลัง save
  /// - อนุมัติแล้ว (approved): พิมพ์ได้
  /// - รออนุมัติ (pending): พิมพ์ไม่ได้
  /// - ถูกปฏิเสธ (rejected): พิมพ์ไม่ได้
  /// - ยังไม่ส่งอนุมัติ: พิมพ์ไม่ได้
  static bool canPrint(POApprovalStatusModel? approvalStatus) {
    // ถ้าไม่มีสถานะการอนุมัติ = ยังไม่ส่งอนุมัติ = พิมพ์ไม่ได้
    if (approvalStatus == null) {
      return false;
    }

    // พิมพ์ได้เฉพาะเมื่ออนุมัติแล้ว หรืออนุมัติอัตโนมัติ
    return approvalStatus.status == POApprovalStatus.approved ||
           approvalStatus.status == POApprovalStatus.autoApproved;
  }

  /// ตรวจสอบว่า PO สามารถแก้ไขได้หรือไม่
  /// [docNo] - เลขที่เอกสาร (ถ้าว่าง = เอกสารใหม่ = แก้ไขได้เสมอ)
  /// [approvalStatus] - สถานะการอนุมัติปัจจุบัน
  /// [creatorCode] - รหัสผู้สร้างเอกสาร
  /// [currentUserCode] - รหัสผู้ใช้งานปัจจุบัน
  ///
  /// Rules:
  /// - เอกสารใหม่ (docNo ว่าง): แก้ไขได้เสมอ
  /// - ไม่ใช่ผู้สร้างเอกสาร: ห้ามแก้ไข
  /// - ยังไม่ส่งอนุมัติ (null): แก้ไขได้
  /// - มีการอนุมัติบางส่วนแล้ว (currentApprovedLevel > 0): ห้ามแก้ไข
  /// - ถูกปฏิเสธ (rejected): แก้ไขได้ (เพื่อส่งอนุมัติใหม่)
  /// - รออนุมัติ (pending): ห้ามแก้ไข
  /// - อนุมัติอัตโนมัติ + ยังไม่ถูกอ้างอิง: แก้ไขได้
  /// - อนุมัติอัตโนมัติ + ถูกอ้างอิงแล้ว: ห้ามแก้ไข
  /// - อนุมัติแล้ว (manual): ห้ามแก้ไข
  static bool canEdit({
    required String docNo,
    required POApprovalStatusModel? approvalStatus,
    bool isref = false,
    String? creatorCode,
    String? currentUserCode,
  }) {
    // เอกสารใหม่ที่ยังไม่มี docno = แก้ไขได้เสมอ
    if (docNo.isEmpty) {
      return true;
    }

    // ตรวจสอบว่าเป็นผู้สร้างเอกสารหรือไม่ — เฉพาะผู้สร้างเท่านั้นที่แก้ไขได้
    if (creatorCode != null && creatorCode.isNotEmpty &&
        currentUserCode != null && currentUserCode.isNotEmpty &&
        creatorCode != currentUserCode) {
      return false;
    }

    // ถ้าไม่มีสถานะการอนุมัติ = ยังไม่ส่งอนุมัติ = แก้ไขได้
    if (approvalStatus == null) {
      return true;
    }

    // ตรวจสอบว่ามีการอนุมัติบางส่วนไปแล้วหรือไม่
    if (approvalStatus.currentApprovedLevel > 0) {
      return false;
    }

    // ถูกปฏิเสธ = แก้ไขได้ (เพื่อส่งอนุมัติใหม่)
    if (approvalStatus.status == POApprovalStatus.rejected) {
      return true;
    }

    // รออนุมัติ (pending) = ห้ามแก้ไข
    if (approvalStatus.status == POApprovalStatus.pending) {
      return false;
    }

    // อนุมัติอัตโนมัติ + ยังไม่ถูกอ้างอิง = แก้ไขได้
    if (approvalStatus.status == POApprovalStatus.autoApproved && !isref) {
      return true;
    }

    // อนุมัติอัตโนมัติ (ถูกอ้างอิงแล้ว) หรือ อนุมัติแล้ว = ห้ามแก้ไข
    if (approvalStatus.status == POApprovalStatus.autoApproved ||
        approvalStatus.status == POApprovalStatus.approved) {
      return false;
    }

    // กรณีอื่นๆ (draft) = แก้ไขได้
    return true;
  }

  /// ข้อความแสดงเหตุผลที่ไม่สามารถแก้ไขได้
  static String getEditDisabledReason({
    required String docNo,
    required POApprovalStatusModel? approvalStatus,
    bool isref = false,
    String? creatorCode,
    String? currentUserCode,
  }) {
    // เอกสารใหม่ = ไม่มีเหตุผลที่จะห้ามแก้ไข
    if (docNo.isEmpty) {
      return '';
    }

    // ตรวจสอบว่าเป็นผู้สร้างเอกสารหรือไม่
    if (creatorCode != null && creatorCode.isNotEmpty &&
        currentUserCode != null && currentUserCode.isNotEmpty &&
        creatorCode != currentUserCode) {
      return global.language("only_creator_can_edit");
    }

    if (approvalStatus == null) {
      return '';
    }

    if (approvalStatus.currentApprovedLevel > 0) {
      return global.language("already_approved_cannot_edit");
    }

    switch (approvalStatus.status) {
      case POApprovalStatus.pending:
        return global.language("pending_approval_cannot_edit");
      case POApprovalStatus.approved:
        return global.language("approved_cannot_edit");
      case POApprovalStatus.autoApproved:
        if (isref) {
          return global.language("document_referenced_cannot_edit");
        }
        return '';
      default:
        return '';
    }
  }

  /// ดึงสี และ icon สำหรับแต่ละสถานะ
  static ApprovalStatusStyle getStatusStyle(POApprovalStatus status) {
    switch (status) {
      case POApprovalStatus.pending:
        return ApprovalStatusStyle(
          bgColor: Colors.orange.shade50,
          textColor: Colors.orange.shade800,
          borderColor: Colors.orange.shade300,
          statusText: global.language("approval_pending"),
          icon: Icons.hourglass_empty,
        );
      case POApprovalStatus.approved:
        return ApprovalStatusStyle(
          bgColor: Colors.green.shade50,
          textColor: Colors.green.shade800,
          borderColor: Colors.green.shade300,
          statusText: global.language("approval_approved"),
          icon: Icons.check_circle,
        );
      case POApprovalStatus.rejected:
        return ApprovalStatusStyle(
          bgColor: Colors.red.shade50,
          textColor: Colors.red.shade800,
          borderColor: Colors.red.shade300,
          statusText: global.language("approval_rejected"),
          icon: Icons.cancel,
        );
      case POApprovalStatus.autoApproved:
        return ApprovalStatusStyle(
          bgColor: Colors.blue.shade50,
          textColor: Colors.blue.shade800,
          borderColor: Colors.blue.shade300,
          statusText: global.language("approval_auto_approved"),
          icon: Icons.verified,
        );
      default:
        return ApprovalStatusStyle(
          bgColor: Colors.grey.shade50,
          textColor: Colors.grey.shade800,
          borderColor: Colors.grey.shade300,
          statusText: global.language("unknown_status"),
          icon: Icons.help_outline,
        );
    }
  }
}

/// Data class สำหรับเก็บ style ของแต่ละสถานะ
class ApprovalStatusStyle {
  final Color bgColor;
  final Color textColor;
  final Color borderColor;
  final String statusText;
  final IconData icon;

  const ApprovalStatusStyle({
    required this.bgColor,
    required this.textColor,
    required this.borderColor,
    required this.statusText,
    required this.icon,
  });
}
