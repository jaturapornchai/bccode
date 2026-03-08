import 'dart:convert';

import 'package:smlaicloud/model/member_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class MemberRepository {
  Future<ApiResponse> getMemberList({
    int page = 0,
    int perPage = 1,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      final response = await client.get(
        '/member?page=$page&limit=$perPage&q=$search',
      );
      try {
        final rawData = json.decode(response.toString());

        // print(rawData);

        if (rawData['error'] != null) {
          String errorMessage = '${rawData['code']}: ${rawData['message']}';
          // print(errorMessage);
          throw Exception('${rawData['code']}: ${rawData['message']}');
        }

        return ApiResponse.fromMap(rawData);
      } catch (ex) {
        // print(ex);
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      // print(errorMessage);
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> getMemberId(String id) async {
    Dio client = Client().init();

    try {
      final response = await client.get('/member/$id');
      try {
        final rawData = json.decode(response.toString());

        // print(rawData);

        if (rawData['error'] != null) {
          String errorMessage = '${rawData['code']}: ${rawData['message']}';
          // print(errorMessage);
          throw Exception('${rawData['code']}: ${rawData['message']}');
        }

        return ApiResponse.fromMap(rawData);
      } catch (ex) {
        // print(ex);
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      // print(errorMessage);
      throw Exception(errorMessage);
    }
  }

  Future<ApiResponse> saveMember(MemberModel member) async {
    Dio client = Client().init();
    final data = member.toJson();
    // print(data);
    try {
      final response = await client.post('/member', data: data);
      try {
        // final rawData = json.decode(response.toString());

        // print(response.data);

        // if (rawData['error'] != null) {
        //   String errorMessage = '${rawData['code']}: ${rawData['message']}';
        //   // print(errorMessage);
        //   throw new Exception('${rawData['code']}: ${rawData['message']}');
        // }

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

  Future<ApiResponse> updateMember(MemberModel member) async {
    Dio client = Client().init();
    final data = member.toJson();
    // print(data);
    try {
      final response = await client.put(
        '/member/${member.guidfixed}',
        data: data,
      );
      try {
        // final rawData = json.decode(response.toString());

        // print(response.data);

        // if (rawData['error'] != null) {
        //   String errorMessage = '${rawData['code']}: ${rawData['message']}';
        //   // print(errorMessage);
        //   throw new Exception('${rawData['code']}: ${rawData['message']}');
        // }

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

  Future<ApiResponse> deleteMember(String id) async {
    Dio client = Client().init();
    try {
      final response = await client.delete('/member/$id');
      try {
        // final rawData = json.decode(response.toString());

        // print(response.data);

        // if (rawData['error'] != null) {
        //   String errorMessage = '${rawData['code']}: ${rawData['message']}';
        //   // print(errorMessage);
        //   throw new Exception('${rawData['code']}: ${rawData['message']}');
        // }

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
}
