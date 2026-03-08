import 'package:smlaicloud/model/shop_model.dart';

import 'client.dart';
import 'package:dio/dio.dart';

class ShopRepository {
  /// profile shop
  Future<ApiResponse> loadShopInfo(String shopid) async {
    Dio client = Client().init();
    try {
      final response = await client.get('/shop/$shopid');
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      throw Exception(ex);
    }
  }

  /// update shop

  Future<ApiResponse> updateShop(String shopid, ShopModel shop) async {
    Dio client = Client().init();
    final data = shop.toJson();
    try {
      final response = await client.put('/shop/$shopid', data: data);
      try {
        return ApiResponse.fromMap(response.data);
      } catch (ex) {
        throw Exception(ex);
      }
    } on DioException catch (ex) {
      throw Exception(ex);
    }
  }
}
