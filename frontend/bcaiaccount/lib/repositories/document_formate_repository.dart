import 'dart:convert';

import 'package:smlaicloud/model/doc_format_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class DocumentFormateRepository {
  Future<ApiResponse> getDocFormatList({
    int limit = 0,
    int offset = 0,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/transaction/document-formate/list?offset=$offset&limit=$limit&q=$search";
      final response = await client.get(query);
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

  Future<ApiResponse> deleteDocFormat(String guid) async {
    Dio client = Client().init();
    try {
      final response =
          await client.delete('/transaction/document-formate/$guid');
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
  Future<ApiResponse> deleteDocFormatMany(List<String> guids) async {
    Dio client = Client().init();
    try {
      final guidStrings = jsonEncode(guids);
      final response = await client.delete('/transaction/document-formate',
          data: guidStrings);
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

  Future<ApiResponse> getDocFormat(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/transaction/document-formate/$guid');
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

  Future<ApiResponse> saveDocFormat(DocFormatModel docFormatModel) async {
    Dio client = Client().init();
    final data = docFormatModel.toJson();
    try {
      final response =
          await client.post('/transaction/document-formate', data: data);
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

  Future<ApiResponse> saveDocFormatBulk(
      List<DocFormatModel> docFormatModel) async {
    Dio client = Client().init();
    try {
      final jsonList =
          docFormatModel.map((docFormat) => docFormat.toJson()).toList();
      final jsonString = jsonEncode(jsonList);
      // print(jsonString);

      final response = await client.post('/transaction/document-formate/bulk',
          data: jsonString);
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

  Future<ApiResponse> updateDocFormat(
      String guid, DocFormatModel docFormatModel) async {
    Dio client = Client().init();
    final data = docFormatModel.toJson();
    try {
      final response =
          await client.put('/transaction/document-formate/$guid', data: data);
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

  Future<ApiResponse> getDocFormatDefault() async {
    Dio client = Client().init();
    try {
      final response =
          await client.get('/transaction/document-formate/default');
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
}
