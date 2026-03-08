import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/cart_model.dart';
import '../models/cart_system_type.dart';
import '../../../utils/logger/app_logger.dart';
import '../../../global.dart' as global;

/// Service สำหรับจัดการตะกร้าสินค้าผ่าน MongoDB API
class MongoDBCartService {
  final String baseUrl;

  MongoDBCartService({String? baseUrl})
    : baseUrl = baseUrl ?? global.goApiUrlPath('atlas');

  // Headers สำหรับ API (ไม่ต้องการ Authorization ตาม API doc)
  Map<String, String> get _headers => {'Content-Type': 'application/json'};

  /// ดึงตะกร้าทั้งหมดของ user
  ///
  /// Query โดยระบุเฉพาะ shopid และ email (ไม่ต้องระบุ cartid)
  /// จะได้ตะกร้าทั้งหมดของ user นั้น
  Future<List<CartModel>> getCarts(String shopId, String email) async {
    try {
      AppLogger.info(
        '🛒 MongoDB: Getting all carts - ShopId: $shopId, Email: $email',
      );

      final url = Uri.parse('$baseUrl/get');
      AppLogger.debug('🛒 MongoDB: Request URL: $url');

      final response = await http.post(
        url,
        headers: _headers,
        body: jsonEncode({
          'collection': 'carts',
          'shopid': shopId,
          'email': email,
          // ไม่ระบุ cartid เพื่อดึงตะกร้าทั้งหมด
        }),
      );

      AppLogger.debug('🛒 MongoDB Response: ${response.statusCode}');

      if (response.statusCode == 200) {
        final responseData = jsonDecode(response.body) as Map<String, dynamic>;
        AppLogger.debug('🛒 MongoDB Response body: ${response.body}');

        // API response format: {"status": "success", "code": 200, "count": 1, "data": [...]}
        final status = responseData['status'] as String?;
        final documents = responseData['data'] as List<dynamic>?;

        if (status == 'success' && documents != null) {
          if (documents.isNotEmpty) {
            AppLogger.info('🛒 MongoDB: Found ${documents.length} cart(s)');
            return documents
                .map((doc) => CartModel.fromJson(doc as Map<String, dynamic>))
                .toList();
          }
        }

        AppLogger.info('🛒 MongoDB: No carts found');
        return [];
      } else if (response.statusCode == 404) {
        AppLogger.info('🛒 MongoDB: No carts exist yet');
        return [];
      } else if (response.statusCode == 502) {
        throw Exception(
          global.language("server_connection_error"),
        );
      } else {
        AppLogger.error(
          '🛒 MongoDB: Failed to load carts - Status: ${response.statusCode}, Body: ${response.body}',
        );
        throw Exception('Failed to load carts: ${response.statusCode}');
      }
    } catch (e, stackTrace) {
      AppLogger.error(
        '🛒 MongoDB: Error loading carts',
        error: e,
        stackTrace: stackTrace,
      );
      throw Exception('Error loading carts: $e');
    }
  }

  /// ดึงตะกร้าเฉพาะ ID
  Future<CartModel?> getCartById(
    String shopId,
    String email,
    String cartId,
  ) async {
    try {
      AppLogger.info(
        '🛒 MongoDB: Getting cart by ID - ShopId: $shopId, Email: $email, CartId: $cartId',
      );

      final url = Uri.parse('$baseUrl/get');
      AppLogger.debug('🛒 MongoDB: Request URL: $url');
      final response = await http.post(
        url,
        headers: _headers,
        body: jsonEncode({
          'collection': 'carts',
          'shopid': shopId,
          'email': email,
          'cartid': cartId,
        }),
      );

      AppLogger.debug('🛒 MongoDB Response: ${response.statusCode}');

      if (response.statusCode == 200) {
        final responseData = jsonDecode(response.body) as Map<String, dynamic>;
        AppLogger.debug('🛒 MongoDB Response body: ${response.body}');

        // API response format: {"status": "success", "code": 200, "count": 1, "data": [...]}
        final status = responseData['status'] as String?;
        final documents = responseData['data'] as List<dynamic>?;

        if (status == 'success' && documents != null && documents.isNotEmpty) {
          AppLogger.info('🛒 MongoDB: Cart found');
          final cartData = documents.first as Map<String, dynamic>;
          return CartModel.fromJson(cartData);
        }

        AppLogger.warning('🛒 MongoDB: Cart not found');
        return null;
      } else if (response.statusCode == 502) {
        throw Exception(
          global.language("server_connection_error"),
        );
      } else {
        AppLogger.error(
          '🛒 MongoDB: Failed to load cart - Status: ${response.statusCode}, Body: ${response.body}',
        );
        throw Exception('Failed to load cart: ${response.statusCode}');
      }
    } catch (e, stackTrace) {
      AppLogger.error(
        '🛒 MongoDB: Error loading cart',
        error: e,
        stackTrace: stackTrace,
      );
      throw Exception('Error loading cart: $e');
    }
  }

  /// สร้างตะกร้าใหม่
  Future<CartModel> createCart(CartModel cart) async {
    try {
      AppLogger.info(
        '🛒 MongoDB: Creating cart - ShopId: ${cart.shopId}, Email: ${cart.email}, CartId: ${cart.cartId}',
      );

      final url = Uri.parse('$baseUrl/update');
      AppLogger.debug('🛒 MongoDB: Request URL: $url');
      final response = await http.post(
        url,
        headers: _headers,
        body: jsonEncode({
          'collection': 'carts',
          'shopid': cart.shopId,
          'email': cart.email,
          'cartid': cart.cartId,
          'data': cart.toJson(),
          'upsert': true,
        }),
      );

      AppLogger.debug('🛒 MongoDB Response: ${response.statusCode}');
      AppLogger.debug('🛒 MongoDB Response body: ${response.body}');

      if (response.statusCode == 200) {
        final responseData = jsonDecode(response.body) as Map<String, dynamic>;
        final status = responseData['status'] as String?;

        if (status == 'success') {
          AppLogger.info('🛒 MongoDB: Cart created successfully');
          return cart;
        } else {
          throw Exception(
            'Create failed: ${responseData['message'] ?? 'Unknown error'}',
          );
        }
      } else if (response.statusCode == 502) {
        throw Exception(
          global.language("server_connection_error"),
        );
      } else {
        AppLogger.error(
          '🛒 MongoDB: Failed to create cart - Status: ${response.statusCode}, Body: ${response.body}',
        );
        throw Exception('Failed to create cart: ${response.statusCode}');
      }
    } catch (e, stackTrace) {
      AppLogger.error(
        '🛒 MongoDB: Error creating cart',
        error: e,
        stackTrace: stackTrace,
      );
      throw Exception('Error creating cart: $e');
    }
  }

  /// อัพเดทตะกร้า
  Future<CartModel> updateCart(CartModel cart) async {
    try {
      AppLogger.info(
        '🛒 MongoDB: Updating cart - ShopId: ${cart.shopId}, Email: ${cart.email}, CartId: ${cart.cartId}',
      );

      final url = Uri.parse('$baseUrl/update');
      AppLogger.debug('🛒 MongoDB: Request URL: $url');
      final response = await http.post(
        url,
        headers: _headers,
        body: jsonEncode({
          'collection': 'carts',
          'shopid': cart.shopId,
          'email': cart.email,
          'cartid': cart.cartId,
          'data': cart.toJson(),
          'upsert': false, // ไม่สร้างใหม่ถ้าไม่มี
        }),
      );

      AppLogger.debug('🛒 MongoDB Response: ${response.statusCode}');
      AppLogger.debug('🛒 MongoDB Response body: ${response.body}');

      if (response.statusCode == 200) {
        final responseData = jsonDecode(response.body) as Map<String, dynamic>;
        final status = responseData['status'] as String?;

        if (status == 'success') {
          AppLogger.info('🛒 MongoDB: Cart updated successfully');
          return cart;
        } else {
          throw Exception(
            'Update failed: ${responseData['message'] ?? 'Unknown error'}',
          );
        }
      } else if (response.statusCode == 502) {
        throw Exception(
          global.language("server_connection_error"),
        );
      } else {
        AppLogger.error(
          '🛒 MongoDB: Failed to update cart - Status: ${response.statusCode}, Body: ${response.body}',
        );
        throw Exception('Failed to update cart: ${response.statusCode}');
      }
    } catch (e, stackTrace) {
      AppLogger.error(
        '🛒 MongoDB: Error updating cart',
        error: e,
        stackTrace: stackTrace,
      );
      throw Exception('Error updating cart: $e');
    }
  }

  /// ลบตะกร้า
  Future<void> deleteCart(String shopId, String email, String cartId) async {
    try {
      AppLogger.info(
        '🛒 MongoDB: Deleting cart - ShopId: $shopId, Email: $email, CartId: $cartId',
      );

      final url = Uri.parse('$baseUrl/delete');
      AppLogger.debug('🛒 MongoDB: Request URL: $url');
      final response = await http.post(
        url,
        headers: _headers,
        body: jsonEncode({
          'collection': 'carts',
          'shopid': shopId,
          'email': email,
          'cartid': cartId,
          'delete_many': false,
        }),
      );

      AppLogger.debug('🛒 MongoDB Response: ${response.statusCode}');
      AppLogger.debug('🛒 MongoDB Response body: ${response.body}');

      if (response.statusCode == 200) {
        final responseData = jsonDecode(response.body) as Map<String, dynamic>;
        final status = responseData['status'] as String?;

        if (status == 'success') {
          AppLogger.info('🛒 MongoDB: Cart deleted successfully');
        } else {
          throw Exception(
            'Delete failed: ${responseData['message'] ?? 'Unknown error'}',
          );
        }
      } else if (response.statusCode == 502) {
        throw Exception(
          global.language("server_connection_error"),
        );
      } else {
        AppLogger.error(
          '🛒 MongoDB: Failed to delete cart - Status: ${response.statusCode}, Body: ${response.body}',
        );
        throw Exception('Failed to delete cart: ${response.statusCode}');
      }
    } catch (e, stackTrace) {
      AppLogger.error(
        '🛒 MongoDB: Error deleting cart',
        error: e,
        stackTrace: stackTrace,
      );
      throw Exception('Error deleting cart: $e');
    }
  }

  /// ดึงตะกร้าตาม systemType (sales, purchase, inventory)
  Future<List<CartModel>> getCartsByType(
    String shopId,
    String email,
    String systemType,
  ) async {
    try {
      AppLogger.info(
        '🛒 MongoDB: Getting carts by type - ShopId: $shopId, Email: $email, Type: $systemType',
      );

      // ดึงตะกร้าทั้งหมดก่อน แล้ว filter ด้วย Dart
      final allCarts = await getCarts(shopId, email);

      // กรองเฉพาะตะกร้าที่ตรงกับ systemType และ status = 'active'
      final filteredCarts = allCarts.where((cart) {
        return cart.systemType.toJson() == systemType &&
            cart.status == 'active';
      }).toList();

      AppLogger.info(
        '🛒 MongoDB: Found ${filteredCarts.length} carts for type: $systemType',
      );

      return filteredCarts;
    } catch (e, stackTrace) {
      AppLogger.error(
        '🛒 MongoDB: Error loading carts by type',
        error: e,
        stackTrace: stackTrace,
      );
      throw Exception('Error loading carts by type: $e');
    }
  }

  /// นับจำนวนตะกร้าที่ active
  Future<int> getActiveCartCount(String shopId, String email) async {
    try {
      final carts = await getCarts(shopId, email);
      return carts.where((cart) => cart.status == 'active').length;
    } catch (e) {
      throw Exception('Error counting active carts: $e');
    }
  }
}
