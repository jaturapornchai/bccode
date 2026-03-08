import 'dart:convert';
import 'dart:io';
import 'package:smlaicloud/model/sale_channel_model.dart';
import 'package:flutter/services.dart';
import 'client.dart';
import 'package:dio/dio.dart';

class SaleChannelRepository {
  Future<ApiResponse> getSaleChannelByCode(String code) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/sale-channel/code/$code');
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

  Future<ApiResponse> getSaleChannelList({
    int limit = 0,
    int offset = 0,
    String search = "",
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/sale-channel/list?offset=$offset&limit=$limit&q=$search&sort=code:1";
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

  Future<ApiResponse> deleteSaleChannel(String guid) async {
    Dio client = Client().init();
    // print(guid);
    try {
      final response = await client.delete('/sale-channel/$guid');
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
  Future<ApiResponse> deleteSaleChannelMany(List<String> guidList) async {
    Dio client = Client().init();
    try {
      final guidStrings = jsonEncode(guidList);
      final response = await client.delete('/sale-channel', data: guidStrings);
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

  Future<ApiResponse> getSaleChannel(String guid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/sale-channel/$guid');
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

  Future<ApiResponse> saveSaleChannel(SaleChannelModel saleChannelModel) async {
    Dio client = Client().init();
    final data = saleChannelModel.toJson();
    try {
      final response = await client.post('/sale-channel', data: data);
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

  Future<ApiResponse> saveSaleChannelBulk(
      List<SaleChannelModel> saleChannels) async {
    Dio client = Client().init();
    try {
      final jsonList =
          saleChannels.map((saleChannels) => saleChannels.toJson()).toList();
      final jsonString = jsonEncode(jsonList);
      // print(jsonString);

      final response =
          await client.post('/sale-channel/bulk', data: jsonString);
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

  Future<ApiResponse> updateSaleChannel(
      String guid, SaleChannelModel saleChannelModel) async {
    Dio client = Client().init();
    final data = saleChannelModel.toJson();
    try {
      final response = await client.put('/sale-channel/$guid', data: data);
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
    FormData formData = FormData.fromMap({
      "file": MultipartFile.fromBytes(image, filename: '$fileName.png'),
    });
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
}
