import 'dart:typed_data';

import 'package:smlaicloud/model/import_product_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class ImportProductRepository {
  Future<ApiResponse> uploadFileExcel(Uint8List file, String filename) async {
    Dio client = Client().init();

    FormData formData = FormData.fromMap({
      "file": MultipartFile.fromBytes(file, filename: filename),
    });

    try {
      final response = await client.post(
        '/productimport/upload',
        data: formData,
      );
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

  Future<ApiResponse> getImportProduct(
    String taskid,
    String q,
    int limit,
    int page,
  ) async {
    Dio client = Client().init();

    try {
      final response = await client.get(
        '/productimport/$taskid?&limit=$limit&page=$page&q=$q',
      );
      try {
        var data = response.data['data'];
        var pagination = response.data['pagination'];

        return ApiResponse.fromMap({
          'success': true,
          'data': data,
          'pagination': pagination,
        });
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }

  /// delete detail by guid
  Future<ApiResponse> deleteDetailByGuid(String guid) async {
    Dio client = Client().init();

    try {
      final response = await client.delete('/productimport/item/$guid');
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

  /// update detail
  Future<ApiResponse> updateDetail(
    String guid,
    ImportProductModel detail,
  ) async {
    Dio client = Client().init();
    final data = detail.toJson();
    try {
      final response = await client.put(
        '/productimport/item/$guid',
        data: data,
      );
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

  /// add detail
  Future<ApiResponse> addDetail(ImportProductModel detail) async {
    Dio client = Client().init();
    final data = detail.toJson();
    try {
      final response = await client.post('/productimport', data: data);
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

  /// save taskid (deprecated - use applyImport instead)
  Future<ApiResponse> saveTaskid(String taskid, String languangecode) async {
    Dio client = Client().init();

    try {
      final response = await client.post(
        '/productimport/$taskid',
        data: {'languangecode': languangecode},
      );
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

  /// apply import - async mode with batch processing
  Future<ApiResponse> applyImport(
    String taskid,
    bool async,
    int batchSize,
    bool forceUpdate,
    String importMode,
  ) async {
    Dio client = Client().init();

    try {
      final response = await client.post(
        '/productimport/$taskid/apply?async=$async&batch_size=$batchSize',
        data: {'force_update': forceUpdate, 'import_mode': importMode},
      );
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

  /// verify taskid
  Future<ApiResponse> verifyTaskid(String taskid) async {
    Dio client = Client().init();

    try {
      final response = await client.post('/productimport/$taskid/verify');
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

  /// getStatusTaskid
  Future<ApiResponse> getStatusTaskid(String taskid) async {
    Dio client = Client().init();

    try {
      final response = await client.get('/productimport/task/$taskid/status');
      try {
        // Return the full response which includes success and data
        var result = response.data;
        return ApiResponse.fromMap({
          'success': result['success'],
          'data': result['data'], // This is the status object
        });
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw errorMessage;
    }
  }

  /// compare detail
  Future<ApiResponse> compareDetail(
    String taskid,
    int page,
    int limit,
    bool fast,
  ) async {
    Dio client = Client().init();

    try {
      final response = await client.get(
        '/productimport/$taskid/compare?page=$page&limit=$limit&fast=$fast',
      );
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
}
