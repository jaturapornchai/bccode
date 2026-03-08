import 'dart:convert';
import 'dart:io';
import 'package:flutter/services.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/price_history_model.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/global.dart' as global;
import 'client.dart';
import 'package:dio/dio.dart';
import 'dart:typed_data';

class ProductBarcodeRepository {
  Future<ApiResponse> getProductBarcodeList({
    int limit = 0,
    int offset = 0,
    String search = "",
    String itemtype = "",
    String branchcode = "",
    String businesstypecode = "",
    String isbom = "",
    String isusesubbarcodes = "",
    String shopsid = "",
  }) async {
    Dio client = Client().init();

    Map<String, dynamic> queryParams = {'offset': offset, 'limit': limit, 'q': search, 'sort': 'barcode:1'};

    if (isbom == "showbom") {
      queryParams['isbom'] = 'true';
    } else if (isbom == "notshowbom") {
      queryParams['isbom'] = 'false';
    }

    if (isusesubbarcodes == "showsubbarcodes") {
      queryParams['isusesubbarcodes'] = 'true';
    } else if (isusesubbarcodes == "notshowsubbarcodes") {
      queryParams['isusesubbarcodes'] = 'false';
    }

    if (itemtype.isNotEmpty) {
      queryParams['itemtype'] = itemtype;
    }

    if (branchcode.isNotEmpty) {
      queryParams['branchcode'] = branchcode;
    }

    if (businesstypecode.isNotEmpty) {
      queryParams['businesstypecode'] = businesstypecode;
    }

    if (shopsid.isNotEmpty) {
      queryParams['shopsid'] = shopsid;
    }

    try {
      final response = await client.get('/product/barcode/list', queryParameters: queryParams);
      try {
        final rawData = json.decode(response.toString());
        if (rawData['error'] != null) {
          throw Exception('${rawData['code']}: ${rawData['message']}');
        }
        return ApiResponse.fromMap(rawData);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> deleteProductBarcode(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/product/barcode/$guid');
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  /// ลบที่ละหลาย GUID
  Future<ApiResponse> deleteProductBarcodeMany(List<String> guids) async {
    Dio client = Client().init();
    final guidStrings = jsonEncode(guids);
    try {
      final response = await client.delete('/product/barcode', data: guidStrings);
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> getProductBarcode(String guid) async {
    Dio client = Client().init();
    try {
      AppLogger.debug('📡 REPOSITORY: เรียก API /product/barcode/$guid');
      final response = await client.get('/product/barcode/$guid');
      AppLogger.debug('📡 REPOSITORY: ได้ response แล้ว - status: ${response.statusCode}');
      AppLogger.debug('📡 REPOSITORY: Raw response data: ${response.data}');

      try {
        final apiResponse = ApiResponse.fromMap(response.data);
        AppLogger.info('✅ REPOSITORY: Parse ApiResponse สำเร็จ - success: ${apiResponse.success}');
        return apiResponse;
      } catch (ex) {
        AppLogger.error('❌ REPOSITORY: Parse ApiResponse ล้มเหลว - $ex');
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      AppLogger.error('❌ REPOSITORY: DioException - type: ${ex.type}, statusCode: ${ex.response?.statusCode}');

      // Handle different error types
      if (ex.response != null) {
        final statusCode = ex.response!.statusCode;
        AppLogger.error('❌ REPOSITORY: Response body: ${ex.response?.data}');

        if (statusCode == 502) {
          throw Exception('เซิร์ฟเวอร์ไม่สามารถเชื่อมต่อได้ กรุณาลองใหม่อีกครั้ง (502 Bad Gateway)');
        } else if (statusCode == 500) {
          throw Exception('เซิร์ฟเวอร์เกิดข้อผิดพลาด กรุณาลองใหม่อีกครั้ง (500 Internal Server Error)');
        } else if (statusCode == 404) {
          throw Exception('ไม่พบข้อมูลสินค้า (404 Not Found)');
        } else if (statusCode == 401 || statusCode == 403) {
          throw Exception('ไม่มีสิทธิ์เข้าถึงข้อมูล กรุณาเข้าสู่ระบบใหม่');
        } else {
          throw Exception('เกิดข้อผิดพลาดจากเซิร์ฟเวอร์ (HTTP $statusCode)');
        }
      } else if (ex.type == DioExceptionType.connectionTimeout || ex.type == DioExceptionType.receiveTimeout) {
        throw Exception('การเชื่อมต่อหมดเวลา กรุณาตรวจสอบอินเทอร์เน็ต');
      } else if (ex.type == DioExceptionType.connectionError) {
        throw Exception('ไม่สามารถเชื่อมต่อเซิร์ฟเวอร์ได้ กรุณาตรวจสอบอินเทอร์เน็ต');
      } else {
        throw Exception('เกิดข้อผิดพลาด: ${ex.message ?? "Unknown error"}');
      }
    }
  }

  Future<ApiResponse> getProductBarcodeRef(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/product/barcode/ref/$guid');
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> getProductBarcodeDetail(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/product/barcode/pk/$guid');
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> saveProductBarcode(ProductBarcodeModel productBarcode) async {
    Dio client = Client().init();
    final data = productBarcode.toJson();
    try {
      final response = await client.post('/product/barcode', data: data);
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw errorMessage;
    }
  }

  Future<ApiResponse> updateProductBarcode(String guid, ProductBarcodeModel productBarcode) async {
    Dio client = Client().init();
    final data = productBarcode.toJson();
    try {
      final response = await client.put('/product/barcode/$guid', data: data);
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> uploadImage(File file, Uint8List image) async {
    Dio client = Client().init();
    String fileName = file.path.split('/').last;
    FormData formData = FormData.fromMap({"file": MultipartFile.fromBytes(image, filename: '$fileName.png')});
    try {
      final response = await client.post('/upload/images', data: formData);
      try {
        // print(response.data);
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        // print(ex);
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      // print(ex);
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> getProductBarcodeBom(String barcode) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/product/barcode/bom/$barcode');
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> getProductBarcodeByBarcode(List<String> productbarcodes) async {
    Dio client = Client().init();
    final barcodesJson = jsonEncode(productbarcodes);
    try {
      final response = await client.get('/product/barcode/by-code?codes=$barcodesJson');
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> uploadFileNoLimitSize(File file, Uint8List fileData) async {
    Dio client = Client().init();
    String fileName = file.path.split('/').last;

    FormData formData = FormData.fromMap({"file": MultipartFile.fromBytes(fileData, filename: fileName)});

    try {
      final response = await client.post('/media/upload/video', data: formData);
      return ApiResponse.fromMap(response.data);
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  /// ค้นหารายการบาร์โค้ดจาก PostgreSQL (GoAPI)
  /// ใช้ goApiUrlPath + goApiPost เพราะเป็น GoAPI endpoint
  Future<Map<String, dynamic>> searchBarcodeListPg({
    required String keyword,
    String groupCode = '',
    String brandCode = '',
    String categoryCode = '',
    String classCode = '',
    String designCode = '',
    String gradeCode = '',
    String modelCode = '',
    String patternCode = '',
    double? priceMin,
    double? priceMax,
    int limit = 50,
    int offset = 0,
    String sortField = 'barcode',
    String sortOrder = 'asc',
  }) async {
    final url = global.goApiUrlPath('api/product/barcode/list');
    Map<String, dynamic> payload = {
      'shopid': global.getShopId(),
      'keyword': keyword,
      'limit': limit,
      'offset': offset,
      'sort_field': sortField,
      'sort_order': sortOrder,
    };
    if (groupCode.isNotEmpty) payload['groupcode'] = groupCode;
    if (brandCode.isNotEmpty) payload['brandcode'] = brandCode;
    if (categoryCode.isNotEmpty) payload['categorycode'] = categoryCode;
    if (classCode.isNotEmpty) payload['classcode'] = classCode;
    if (designCode.isNotEmpty) payload['designcode'] = designCode;
    if (gradeCode.isNotEmpty) payload['gradecode'] = gradeCode;
    if (modelCode.isNotEmpty) payload['modelcode'] = modelCode;
    if (patternCode.isNotEmpty) payload['patterncode'] = patternCode;
    if (priceMin != null) payload['price_min'] = priceMin;
    if (priceMax != null) payload['price_max'] = priceMax;

    try {
      final result = await global.goApiPost(url, payload);
      if (result is Map<String, dynamic>) {
        return result;
      }
      return {'success': false, 'message': 'Invalid response format', 'data': [], 'total': 0};
    } catch (e) {
      AppLogger.error('searchBarcodeListPg error: $e');
      return {'success': false, 'message': e.toString(), 'data': [], 'total': 0};
    }
  }

  Future<PriceHistoryResponseModel> getProductBarcodePriceHistory({required String barcode, int page = 1, int limit = 20}) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/product/barcode/price-history/$barcode?page=$page&limit=$limit');
      try {
        return PriceHistoryResponseModel.fromJson(response.data);
      } catch (ex) {
        throw Exception(ex.toString());
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }
}
