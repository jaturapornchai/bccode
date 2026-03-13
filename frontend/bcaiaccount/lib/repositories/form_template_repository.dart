import 'dart:convert';
import 'package:dio/dio.dart';
import 'client.dart';

class FormTemplateRepository {
  /// Save (upsert by code) — สร้างใหม่หรืออัปเดตถ้า code ซ้ำ
  Future<ApiResponse> saveFormTemplate(Map<String, dynamic> data) async {
    Dio client = Client().init();
    try {
      final response =
          await client.post('/form/template/save', data: jsonEncode(data));
      return ApiResponse.fromMap(response.data);
    } on DioException catch (ex) {
      throw Exception(ex.response.toString());
    }
  }

  /// List templates with pagination + doctype filter
  Future<ApiResponse> getFormTemplateList({
    int limit = 20,
    int offset = 0,
    String search = '',
    String doctype = '',
  }) async {
    Dio client = Client().init();
    try {
      String query =
          '/form/template?offset=$offset&limit=$limit&q=$search';
      if (doctype.isNotEmpty) {
        query += '&doctype=$doctype';
      }
      final response = await client.get(query);
      final rawData = json.decode(response.toString());
      if (rawData['error'] != null && rawData['error'] == true) {
        throw Exception('${rawData['code']}: ${rawData['message']}');
      }
      return ApiResponse.fromMap(rawData);
    } on DioException catch (ex) {
      throw Exception(ex.response.toString());
    }
  }

  /// Get single template by guidfixed
  Future<ApiResponse> getFormTemplate(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/form/template/$guid');
      return ApiResponse.fromMap(response.data);
    } on DioException catch (ex) {
      throw Exception(ex.response.toString());
    }
  }

  /// Update template by guidfixed
  Future<ApiResponse> updateFormTemplate(
      String guid, Map<String, dynamic> data) async {
    Dio client = Client().init();
    try {
      final response =
          await client.put('/form/template/$guid', data: jsonEncode(data));
      return ApiResponse.fromMap(response.data);
    } on DioException catch (ex) {
      throw Exception(ex.response.toString());
    }
  }

  /// Delete template by guidfixed
  Future<ApiResponse> deleteFormTemplate(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/form/template/$guid');
      return ApiResponse.fromMap(response.data);
    } on DioException catch (ex) {
      throw Exception(ex.response.toString());
    }
  }

  /// Delete multiple templates by guidfixeds
  Future<ApiResponse> deleteFormTemplateMany(List<String> guids) async {
    Dio client = Client().init();
    try {
      final response =
          await client.delete('/form/template', data: jsonEncode(guids));
      return ApiResponse.fromMap(response.data);
    } on DioException catch (ex) {
      throw Exception(ex.response.toString());
    }
  }
}
