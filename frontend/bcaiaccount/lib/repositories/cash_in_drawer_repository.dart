import 'client.dart';
import 'package:dio/dio.dart';

class CashInDrawerRepository {
  Future<ApiResponse> getShiftList({
    String fromdate = '',
    String todate = '',
    List<int> doctype = const [],
    int limit = 0,
    int page = 0,
    String usercode = '',
    String posid = '',
  }) async {
    Dio client = Client().init();

    try {
      String query =
          "/pos/shift?limit=$limit&page=$page&doctype=${doctype.join(',')}&fromdate=$fromdate&todate=$todate&usercode=$usercode&posid=$posid";
      final response = await client.get(query);
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

  Future<ApiResponse> getShiftReportDetails({
    required String docno,
  }) async {
    Dio client = Client().init();

    try {
      String query = "/pos/shift/report/$docno";
      final response = await client.get(query);
      try {
        var data = response.data['data'] ?? response.data;

        return ApiResponse.fromMap({
          'success': true,
          'data': data,
        });
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      String errorMessage = ex.response.toString();
      throw Exception(errorMessage);
    }
  }
}
