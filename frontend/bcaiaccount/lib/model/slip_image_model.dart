import 'dart:typed_data';
import '../global.dart' as global;

// Model สำหรับข้อมูล SLIP เงินเข้า/ออก
// ใช้กับ API: /image/upload, /image/list, /image/delete, /image/get

/// ประเภทของ SLIP
enum SlipType {
  moneyIn, // SLIP เงินเข้า
  moneyOut, // SLIP เงินออก
}

/// Extension เพื่อแปลง SlipType เป็น String สำหรับใช้กับ API
extension SlipTypeExtension on SlipType {
  /// ค่า category สำหรับส่งไป API
  String get value {
    switch (this) {
      case SlipType.moneyIn:
        return 'slip_money_in';
      case SlipType.moneyOut:
        return 'slip_money_out';
    }
  }

  /// ชื่อแสดงผลภาษาไทย
  String get displayName {
    switch (this) {
      case SlipType.moneyIn:
        return global.language('slip_money_in');
      case SlipType.moneyOut:
        return global.language('slip_money_out');
    }
  }
}

/// Model สำหรับข้อมูล SLIP แต่ละรายการ
/// ตาม API Response: id, shopid, filename, original_name, content_type, size, category, created_at
class SlipImageModel {
  final String id;
  final String shopId;
  final String filename;
  final String originalName;
  final String contentType;
  final int size;
  final String category;
  final String? description;
  final List<String>? tags;
  final String? uploadedBy;
  final DateTime? createdAt;
  final String? base64Data; // ข้อมูล base64 ถ้าใช้ include_data=true (deprecated)
  final String? url; // Presigned URL จาก Cloudflare R2 (หมดอายุใน 60 นาที)
  final Uint8List? imageBytes; // ข้อมูลรูปภาพแบบ bytes สำหรับแสดงผลโดยตรง

  // สถานะการตรวจสอบสลิป
  final bool? verified; // true = ตรวจสอบแล้ว, false/null = ยังไม่ตรวจสอบ
  final String? verifyStatus; // success, duplicate, not_found, error
  final bool? isDuplicate; // true = slip ซ้ำ
  final String? transRef; // หมายเลขอ้างอิง
  final double? transAmount; // จำนวนเงิน
  final String? senderName; // ชื่อผู้โอน
  final String? receiverName; // ชื่อผู้รับ

  SlipImageModel({
    required this.id,
    required this.shopId,
    required this.filename,
    required this.originalName,
    required this.contentType,
    required this.size,
    required this.category,
    this.description,
    this.tags,
    this.uploadedBy,
    this.createdAt,
    this.base64Data,
    this.url,
    this.imageBytes,
    this.verified,
    this.verifyStatus,
    this.isDuplicate,
    this.transRef,
    this.transAmount,
    this.senderName,
    this.receiverName,
  });

  /// สร้าง SlipImageModel จาก JSON (ตาม API Response)
  factory SlipImageModel.fromJson(Map<String, dynamic> json) {
    return SlipImageModel(
      id: json['id'] ?? '',
      shopId: json['shopid'] ?? '',
      filename: json['filename'] ?? '',
      originalName: json['original_name'] ?? '',
      contentType: json['content_type'] ?? '',
      size: json['size'] ?? 0,
      category: json['category'] ?? '',
      description: json['description'],
      tags: json['tags'] != null ? List<String>.from(json['tags']) : null,
      uploadedBy: json['uploaded_by'],
      createdAt: json['created_at'] != null
          ? DateTime.tryParse(json['created_at'].toString())
          : null,
      base64Data: json['data'], // base64 data URL ถ้ามี (deprecated)
      url: json['url'], // Presigned URL จาก Cloudflare R2
      verified: json['verified'],
      verifyStatus: json['verify_status'],
      isDuplicate: json['is_duplicate'],
      transRef: json['trans_ref'],
      transAmount: json['trans_amount'] != null
          ? double.tryParse(json['trans_amount'].toString())
          : null,
      senderName: json['sender_name'],
      receiverName: json['receiver_name'],
    );
  }

  /// แปลง SlipImageModel เป็น JSON
  Map<String, dynamic> toJson() {
    return {
      'id': id,
      'shopid': shopId,
      'filename': filename,
      'original_name': originalName,
      'content_type': contentType,
      'size': size,
      'category': category,
      'description': description,
      'tags': tags,
      'uploaded_by': uploadedBy,
      'created_at': createdAt?.toIso8601String(),
      'data': base64Data,
      'url': url,
      'verified': verified,
      'verify_status': verifyStatus,
      'is_duplicate': isDuplicate,
      'trans_ref': transRef,
      'trans_amount': transAmount,
      'sender_name': senderName,
      'receiver_name': receiverName,
    };
  }

  /// คัดลอก SlipImageModel พร้อมอัพเดตข้อมูลการตรวจสอบ
  SlipImageModel copyWithVerification({
    bool? verified,
    String? verifyStatus,
    bool? isDuplicate,
    String? transRef,
    double? transAmount,
    String? senderName,
    String? receiverName,
  }) {
    return SlipImageModel(
      id: id,
      shopId: shopId,
      filename: filename,
      originalName: originalName,
      contentType: contentType,
      size: size,
      category: category,
      description: description,
      tags: tags,
      uploadedBy: uploadedBy,
      createdAt: createdAt,
      base64Data: base64Data,
      url: url,
      imageBytes: imageBytes,
      verified: verified ?? this.verified,
      verifyStatus: verifyStatus ?? this.verifyStatus,
      isDuplicate: isDuplicate ?? this.isDuplicate,
      transRef: transRef ?? this.transRef,
      transAmount: transAmount ?? this.transAmount,
      senderName: senderName ?? this.senderName,
      receiverName: receiverName ?? this.receiverName,
    );
  }
}

/// Model สำหรับ Response จาก API /image/list
/// API Response: { status: "success", data: [...], total: N }
class SlipImageListResponse {
  final bool success;
  final List<SlipImageModel> data;
  final int total;

  SlipImageListResponse({
    required this.success,
    required this.data,
    required this.total,
  });

  factory SlipImageListResponse.fromJson(Map<String, dynamic> json) {
    List<SlipImageModel> images = [];
    if (json['data'] != null) {
      images = (json['data'] as List)
          .map((item) => SlipImageModel.fromJson(item))
          .toList();
    }

    // API ใช้ status: "success" แทน success: true
    bool isSuccess = json['status'] == 'success' || json['success'] == true;

    return SlipImageListResponse(
      success: isSuccess,
      data: images,
      total: json['total'] ?? images.length,
    );
  }
}

/// Model สำหรับ Response จาก API /upload/images หรือ /image/upload
/// /upload/images Response: { success: true, data: { uri: "https://..." } }
/// /image/upload Response: { status: "success", data: { id, filename, ... } }
class SlipUploadResponse {
  final bool success;
  final String? uri; // URL จาก /upload/images
  final String? id;
  final String? filename;
  final String? originalName;
  final String? contentType;
  final int? size;
  final String? category;
  final String? message;

  SlipUploadResponse({
    required this.success,
    this.uri,
    this.id,
    this.filename,
    this.originalName,
    this.contentType,
    this.size,
    this.category,
    this.message,
  });

  factory SlipUploadResponse.fromJson(Map<String, dynamic> json) {
    // รองรับทั้ง success: true และ status: "success"
    bool isSuccess = json['success'] == true || json['status'] == 'success';

    // ข้อมูลไฟล์อยู่ใน data object
    Map<String, dynamic>? data = json['data'];

    return SlipUploadResponse(
      success: isSuccess,
      uri: data?['uri'] ?? json['uri'], // URL จาก /upload/images
      id: data?['id'] ?? json['id'],
      filename: data?['filename'] ?? json['filename'],
      originalName: data?['original_name'] ?? json['original_name'],
      contentType: data?['content_type'] ?? json['content_type'],
      size: data?['size'] ?? json['size'],
      category: data?['category'] ?? json['category'],
      message: json['message'],
    );
  }
}

/// Model สำหรับ Response จาก API /image/get
/// API Response: { status: "success", data: { ..., data: "data:image/jpeg;base64,..." } }
class SlipImageDataResponse {
  final bool success;
  final String? id;
  final String? filename;
  final String? originalName;
  final String? contentType;
  final int? size;
  final String? category;
  final String? base64Data; // data:image/jpeg;base64,...
  final String? message;

  SlipImageDataResponse({
    required this.success,
    this.id,
    this.filename,
    this.originalName,
    this.contentType,
    this.size,
    this.category,
    this.base64Data,
    this.message,
  });

  factory SlipImageDataResponse.fromJson(Map<String, dynamic> json) {
    bool isSuccess = json['status'] == 'success' || json['success'] == true;
    Map<String, dynamic>? data = json['data'];

    return SlipImageDataResponse(
      success: isSuccess,
      id: data?['id'],
      filename: data?['filename'],
      originalName: data?['original_name'],
      contentType: data?['content_type'],
      size: data?['size'],
      category: data?['category'],
      base64Data: data?['data'], // base64 data URL
      message: json['message'],
    );
  }
}

/// Model สำหรับ Response จาก API /image/delete
/// API Response: { status: "success", message: "..." }
class SlipDeleteResponse {
  final bool success;
  final String? message;

  SlipDeleteResponse({
    required this.success,
    this.message,
  });

  factory SlipDeleteResponse.fromJson(Map<String, dynamic> json) {
    bool isSuccess = json['status'] == 'success' || json['success'] == true;

    return SlipDeleteResponse(
      success: isSuccess,
      message: json['message'],
    );
  }
}

/// Model สำหรับ Response จาก API /image/promptpayverify
/// ใช้ตรวจสอบสลิป PromptPay ว่าเป็นสลิปจริงหรือปลอม
class SlipVerifyResponse {
  final bool success;
  final bool? verified; // true = ตรวจสอบผ่าน
  final String? verifyStatus; // success, duplicate, not_found, error
  final bool? isDuplicate; // true = slip ซ้ำ
  final String? slipType; // "bank" หรือ "truewallet"
  final String? transRef; // หมายเลขอ้างอิง
  final double? transAmount; // จำนวนเงิน
  final String? senderName; // ชื่อผู้โอน
  final String? senderAccount; // เลขบัญชีผู้โอน
  final String? senderBank; // ธนาคารผู้โอน
  final String? receiverName; // ชื่อผู้รับ
  final String? receiverAccount; // เลขบัญชีผู้รับ
  final String? receiverBank; // ธนาคารผู้รับ
  final DateTime? transDate; // วันที่ทำรายการ
  final String? message;
  final String? errorCode;

  SlipVerifyResponse({
    required this.success,
    this.verified,
    this.verifyStatus,
    this.isDuplicate,
    this.slipType,
    this.transRef,
    this.transAmount,
    this.senderName,
    this.senderAccount,
    this.senderBank,
    this.receiverName,
    this.receiverAccount,
    this.receiverBank,
    this.transDate,
    this.message,
    this.errorCode,
  });

  factory SlipVerifyResponse.fromJson(Map<String, dynamic> json) {
    bool isSuccess = json['status'] == 'success' || json['success'] == true;
    Map<String, dynamic>? data = json['data'];

    return SlipVerifyResponse(
      success: isSuccess,
      verified: data?['verified'] ?? json['verified'],
      verifyStatus: data?['verify_status'] ?? json['verify_status'],
      isDuplicate: data?['is_duplicate'] ?? json['is_duplicate'],
      slipType: data?['slip_type'] ?? json['slip_type'],
      transRef: data?['trans_ref'] ?? json['trans_ref'],
      transAmount: data?['trans_amount'] != null
          ? double.tryParse(data!['trans_amount'].toString())
          : (json['trans_amount'] != null
              ? double.tryParse(json['trans_amount'].toString())
              : null),
      senderName: data?['sender_name'] ?? json['sender_name'],
      senderAccount: data?['sender_account'] ?? json['sender_account'],
      senderBank: data?['sender_bank'] ?? json['sender_bank'],
      receiverName: data?['receiver_name'] ?? json['receiver_name'],
      receiverAccount: data?['receiver_account'] ?? json['receiver_account'],
      receiverBank: data?['receiver_bank'] ?? json['receiver_bank'],
      transDate: (data?['trans_date'] ?? json['trans_date']) != null
          ? DateTime.tryParse((data?['trans_date'] ?? json['trans_date']).toString())
          : null,
      message: json['message'],
      errorCode: data?['error_code'] ?? json['error_code'],
    );
  }
}
