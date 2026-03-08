import 'package:dio/dio.dart';
import 'package:flutter/foundation.dart';
import 'package:smlaicloud/model/slip_image_model.dart';
import 'package:smlaicloud/global.dart' as global;

/// Repository สำหรับจัดการรูปภาพ SLIP เงินเข้า/ออก
/// ใช้ Image API: /image/upload, /image/list, /image/delete
/// ทุก endpoint ใช้ goApiUrlPath (mainapi:9090/goapi)
class SlipImageRepository {
  /// อัพโหลดรูปภาพ SLIP
  /// [imageBytes] - ข้อมูลรูปภาพแบบ bytes
  /// [fileName] - ชื่อไฟล์
  /// [slipType] - ประเภท SLIP (moneyIn/moneyOut) ใช้เป็น category
  Future<SlipUploadResponse> uploadSlipImage({
    required Uint8List imageBytes,
    required String fileName,
    required SlipType slipType,
    String? description,
  }) async {
    debugPrint('[SLIP REPO] uploadSlipImage called');
    debugPrint('[SLIP REPO] fileName: $fileName, category: ${slipType.value}');

    String shopId = global.getShopId();
    debugPrint('[SLIP REPO] shopId: $shopId');

    // ใช้ goApiUrlPath (mainapi:9090/goapi)
    final url = global.goApiUrlPath('image/upload');
    debugPrint('[SLIP REPO] Full URL: $url');

    // สร้าง FormData ตาม API spec
    // Request (multipart/form-data): shopid, file, category, description (optional)
    Dio dio = Dio();
    FormData formData = FormData.fromMap({
      'shopid': shopId,
      'file': MultipartFile.fromBytes(imageBytes, filename: fileName),
      'category': slipType.value,
      if (description != null) 'description': description,
    });

    try {
      debugPrint('[SLIP REPO] Posting to $url...');
      final response = await dio.post(url, data: formData);
      debugPrint('[SLIP REPO] Response: ${response.data}');
      return SlipUploadResponse.fromJson(response.data);
    } on DioException catch (ex) {
      debugPrint('[SLIP REPO] DioException: ${ex.message}');
      debugPrint('[SLIP REPO] Response: ${ex.response?.data}');
      String errorMessage = ex.response?.data?['message'] ?? 'Upload failed';
      throw Exception(errorMessage);
    }
  }

  /// ดึงรายการรูปภาพ SLIP
  /// [slipType] - ประเภท SLIP
  /// [limit] - จำนวนรายการต่อหน้า
  /// [skip] - จำนวนรายการที่จะข้าม (สำหรับ pagination)
  /// [includeData] - ถ้า true จะส่ง base64 data มาด้วย
  Future<SlipImageListResponse> getSlipImageList({
    required SlipType slipType,
    int limit = 100,
    int skip = 0,
    bool includeData = false,
  }) async {
    debugPrint('[SLIP REPO] getSlipImageList called');
    debugPrint('[SLIP REPO] category: ${slipType.value}, limit: $limit, skip: $skip');

    String shopId = global.getShopId();
    debugPrint('[SLIP REPO] shopId: $shopId');

    // ใช้ goApiUrlPath เหมือน function อื่นๆ (mainapi:9090/goapi)
    final url = global.goApiUrlPath('image/list');
    debugPrint('[SLIP REPO] Full URL: $url');

    // Request body ตาม API spec
    Map<String, dynamic> requestBody = {
      'shopid': shopId,
      'category': slipType.value,
      'limit': limit,
      'skip': skip,
      'include_data': includeData,
    };

    try {
      debugPrint('[SLIP REPO] Request body: $requestBody');
      final result = await global.goApiPost(url, requestBody);
      debugPrint('[SLIP REPO] Response: $result');

      if (result is Map<String, dynamic>) {
        return SlipImageListResponse.fromJson(result);
      } else {
        debugPrint('[SLIP REPO] Unexpected response type: ${result.runtimeType}');
        return SlipImageListResponse(success: false, data: [], total: 0);
      }
    } catch (e) {
      debugPrint('[SLIP REPO] Error: $e');
      throw Exception('Failed to load images: $e');
    }
  }

  /// ดึงข้อมูลรูปภาพ (รวม base64 data)
  /// [filename] - ชื่อไฟล์ที่ต้องการ
  Future<SlipImageDataResponse> getSlipImageData({
    required String filename,
  }) async {
    debugPrint('[SLIP REPO] getSlipImageData called');
    debugPrint('[SLIP REPO] filename: $filename');

    String shopId = global.getShopId();

    // ใช้ goApiUrlPath (mainapi:9090/goapi)
    final url = global.goApiUrlPath('image/get');
    debugPrint('[SLIP REPO] Full URL: $url');

    Map<String, dynamic> requestBody = {
      'shopid': shopId,
      'filename': filename,
    };

    try {
      debugPrint('[SLIP REPO] Request body: $requestBody');
      final result = await global.goApiPost(url, requestBody);
      debugPrint('[SLIP REPO] Response received');

      if (result is Map<String, dynamic>) {
        return SlipImageDataResponse.fromJson(result);
      } else {
        debugPrint('[SLIP REPO] Unexpected response type: ${result.runtimeType}');
        return SlipImageDataResponse(success: false, message: 'Unexpected response');
      }
    } catch (e) {
      debugPrint('[SLIP REPO] Error: $e');
      throw Exception('Failed to get image: $e');
    }
  }

  /// ลบรูปภาพ SLIP
  /// [filename] - ชื่อไฟล์ที่ต้องการลบ (หรือใช้ imageId)
  /// [imageId] - ID ของรูปภาพ (ถ้ามี)
  Future<SlipDeleteResponse> deleteSlipImage({
    String? filename,
    String? imageId,
  }) async {
    debugPrint('[SLIP REPO] deleteSlipImage called');
    debugPrint('[SLIP REPO] filename: $filename, imageId: $imageId');

    if (filename == null && imageId == null) {
      throw Exception('ต้องระบุ filename หรือ imageId');
    }

    String shopId = global.getShopId();

    // ใช้ goApiUrlPath เหมือน function อื่นๆ (mainapi:9090/goapi)
    final url = global.goApiUrlPath('image/delete');
    debugPrint('[SLIP REPO] Full URL: $url');

    Map<String, dynamic> requestBody = {
      'shopid': shopId,
      if (filename != null) 'filename': filename,
      if (imageId != null) 'image_id': imageId,
    };

    try {
      debugPrint('[SLIP REPO] Request body: $requestBody');
      final result = await global.goApiPost(url, requestBody);
      debugPrint('[SLIP REPO] Response: $result');

      if (result is Map<String, dynamic>) {
        return SlipDeleteResponse.fromJson(result);
      } else {
        debugPrint('[SLIP REPO] Unexpected response type: ${result.runtimeType}');
        return SlipDeleteResponse(success: false, message: 'Unexpected response');
      }
    } catch (e) {
      debugPrint('[SLIP REPO] Error: $e');
      throw Exception('Failed to delete image: $e');
    }
  }

  /// ตรวจสอบ Slip PromptPay/TrueMoney
  /// [imageId] - ID ของรูปภาพ (ใช้อย่างใดอย่างหนึ่งกับ filename)
  /// [filename] - ชื่อไฟล์ (ใช้อย่างใดอย่างหนึ่งกับ imageId)
  /// [checkDuplicate] - ตรวจสอบ slip ซ้ำหรือไม่ (default: true)
  /// [type] - ประเภท "bank" หรือ "truewallet" (default: "bank")
  Future<SlipVerifyResponse> verifySlip({
    String? imageId,
    String? filename,
    bool checkDuplicate = true,
    String type = 'bank',
  }) async {
    debugPrint('[SLIP REPO] verifySlip called');
    debugPrint('[SLIP REPO] imageId: $imageId, filename: $filename, type: $type');

    if (imageId == null && filename == null) {
      throw Exception('ต้องระบุ imageId หรือ filename');
    }

    String shopId = global.getShopId();
    debugPrint('[SLIP REPO] shopId: $shopId');

    // ใช้ goApiUrlPath (mainapi:9090/goapi)
    final url = global.goApiUrlPath('image/promptpayverify');
    debugPrint('[SLIP REPO] Full URL: $url');

    Map<String, dynamic> requestBody = {
      'shopid': shopId,
      if (imageId != null) 'image_id': imageId,
      if (filename != null) 'filename': filename,
      'check_duplicate': checkDuplicate,
      'type': type,
    };

    try {
      debugPrint('[SLIP REPO] Request body: $requestBody');
      final result = await global.goApiPost(url, requestBody);
      debugPrint('[SLIP REPO] Response: $result');

      if (result is Map<String, dynamic>) {
        return SlipVerifyResponse.fromJson(result);
      } else {
        debugPrint('[SLIP REPO] Unexpected response type: ${result.runtimeType}');
        return SlipVerifyResponse(success: false, message: 'Unexpected response');
      }
    } catch (e) {
      debugPrint('[SLIP REPO] Error: $e');
      throw Exception('Failed to verify slip: $e');
    }
  }
}
