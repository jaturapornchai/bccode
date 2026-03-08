import 'client.dart';
import 'package:dio/dio.dart';

class ApiKeyRepository {
  Future<ApiResponse> apiKeyService() async {
    Dio client = Client().init();
    try {
      final response = await client.post('/apikeyservice');
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

  Future<ApiResponse> deleteApiKeyService(String apikey) async {
    Dio client = Client().init();
    final data = {"apikey": apikey};
    try {
      final response = await client.delete('/apikeyservice', data: data);
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
