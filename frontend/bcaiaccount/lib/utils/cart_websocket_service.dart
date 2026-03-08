import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:smlaicloud/environment.dart';
import 'package:smlaicloud/model/cart_model.dart';
import 'package:smlaicloud/model/location_model.dart';
import 'package:smlaicloud/model/warehouse_model.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:device_info_plus/device_info_plus.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:web_socket_channel/web_socket_channel.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class CartWebSocketService {
  WebSocketChannel? channel;
  Stream? broadcastStream;
  String? clientId;
  bool isConnected = false;
  String? deviceId;
  StreamSubscription? _streamSubscription;
  List<CartModel> carts = [];
  bool isLoading = true;

  // Define event handlers using functions that will be passed from screens
  Function(List<CartModel>)? onCartListReceived;
  Function(String, String, Map<String, dynamic>, WarehouseModel, LocationModel,
      WarehouseModel, LocationModel)? onCartDetailsReceived;
  Function(String)? onClientIdReceived;
  Function(bool)? onConnectionChanged;
  Function(String)? onError;
  Function(bool, String)? onCartDeleted; // เพิ่ม callback สำหรับการลบตะกร้า

  // Singleton pattern
  static final CartWebSocketService _instance =
      CartWebSocketService._internal();

  factory CartWebSocketService() {
    return _instance;
  }

  CartWebSocketService._internal();

  // Initialize the service with callbacks
  void initialize({
    Function(List<CartModel>)? onCartListReceived,
    Function(String, String, Map<String, dynamic>, WarehouseModel,
            LocationModel, WarehouseModel, LocationModel)?
        onCartDetailsReceived,
    Function(String)? onClientIdReceived,
    Function(bool)? onConnectionChanged,
    Function(String)? onError,
    Function(bool, String)? onCartDeleted, // เพิ่ม parameter สำหรับการลบตะกร้า
  }) {
    this.onCartListReceived = onCartListReceived;
    this.onCartDetailsReceived = onCartDetailsReceived;
    this.onClientIdReceived = onClientIdReceived;
    this.onConnectionChanged = onConnectionChanged;
    this.onError = onError;
    this.onCartDeleted = onCartDeleted; // กำหนดค่า callback
  }

  Future<String> _getDeviceId() async {
    DeviceInfoPlugin deviceInfo = DeviceInfoPlugin();
    try {
      if (Platform.isAndroid) {
        final androidInfo = await deviceInfo.androidInfo;
        return '${androidInfo.model}_${androidInfo.id}';
      } else if (Platform.isIOS) {
        final iosInfo = await deviceInfo.iosInfo;
        return '${iosInfo.model}_${iosInfo.identifierForVendor ?? 'unknown'}';
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error getting device ID: $e');
      }
    }
    return 'unknown';
  }

  Future<bool> connect(BuildContext context) async {
    try {
      deviceId = await _getDeviceId();

      try {
        channel = WebSocketChannel.connect(
            Uri.parse(Environment().config.webSocketCartService));
      } catch (socketError) {
        if (kDebugMode) {
          AppLogger.error('WebSocket connection error: $socketError');
        }

        if (onError != null) {
          onError!(global.language('chatbot_cannot_connect_to_server'));
        }

        isConnected = false;
        if (onConnectionChanged != null) {
          onConnectionChanged!(isConnected);
        }
        return false;
      }

      broadcastStream = channel!.stream.asBroadcastStream();

      await _streamSubscription?.cancel();

      _streamSubscription = broadcastStream!.listen(
        (message) {
          _handleMessage(message);
        },
        onDone: () {
          isConnected = false;
          if (onConnectionChanged != null) {
            onConnectionChanged!(isConnected);
          }
        },
        onError: (error) {
          isConnected = false;
          if (onConnectionChanged != null) {
            onConnectionChanged!(isConnected);
          }
          if (onError != null) {
            onError!(error.toString());
          }
        },
      );

      channel?.sink.add(jsonEncode({
        'type': 'connect',
        'deviceId': deviceId,
      }));

      isConnected = true;
      if (onConnectionChanged != null) {
        onConnectionChanged!(isConnected);
      }

      return true;
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error: $e');
      }
      isConnected = false;
      if (onConnectionChanged != null) {
        onConnectionChanged!(isConnected);
      }
      return false;
    }
  }

  void disconnect() {
    _streamSubscription?.cancel();
    channel?.sink.close();
    isConnected = false;
    if (onConnectionChanged != null) {
      onConnectionChanged!(isConnected);
    }
  }

  void _handleMessage(dynamic message) {
    try {
      final data = jsonDecode(message);

      switch (data['type']) {
        case 'init':
          clientId = '_$deviceId';
          if (onClientIdReceived != null) {
            onClientIdReceived!(clientId!);
          }
          break;

        case 'cartList':
          try {
            carts = (data['carts'] as List?)
                    ?.map((json) => CartModel.fromJson(json))
                    .toList() ??
                [];
            isLoading = false;
            if (onCartListReceived != null) {
              onCartListReceived!(carts);
            }
          } catch (e) {
            if (kDebugMode) {
              AppLogger.error('Error parsing cart list: $e');
            }
            if (onError != null) {
              onError!('เกิดข้อผิดพลาดในการอ่านข้อมูลตะกร้า: ${e.toString()}');
            }
          }
          break;

        case 'cartDetails':
          String cartName = data['cartName'] ?? '';
          String cartId = data['cartId'] ?? '';

          WarehouseModel currentWarehouse;
          if (data['warehouse'] != null) {
            currentWarehouse = WarehouseModel(
              guidfixed: data['warehouse']['guidfixed'],
              code: data['warehouse']['code'],
              names:
                  global.convertToLanguageDataList(data['warehouse']['names']),
            );
          } else {
            currentWarehouse = WarehouseModel(
              guidfixed: '',
              code: '',
              names: [],
            );
          }

          LocationModel currentLocation;
          if (data['location'] != null) {
            currentLocation = LocationModel(
              code: data['location']['code'],
              names:
                  global.convertToLanguageDataList(data['location']['names']),
            );
          } else {
            currentLocation = LocationModel(
              code: '',
              names: [],
            );
          }

          WarehouseModel currentDestWarehouse;
          if (data['destWarehouse'] != null) {
            currentDestWarehouse = WarehouseModel(
              guidfixed: data['destWarehouse']['guidfixed'],
              code: data['destWarehouse']['code'],
              names: global
                  .convertToLanguageDataList(data['destWarehouse']['names']),
            );
          } else {
            currentDestWarehouse = WarehouseModel(
              guidfixed: '',
              code: '',
              names: [],
            );
          }

          LocationModel currentDestLocation;
          if (data['destLocation'] != null) {
            currentDestLocation = LocationModel(
              code: data['destLocation']['code'],
              names: global
                  .convertToLanguageDataList(data['destLocation']['names']),
            );
          } else {
            currentDestLocation = LocationModel(
              code: '',
              names: [],
            );
          }

          // Convert map to list of cart items
          final itemDetailsMap = data['itemDetails'] as Map<String, dynamic>;

          if (onCartDetailsReceived != null) {
            onCartDetailsReceived!(
                cartName,
                cartId,
                itemDetailsMap,
                currentWarehouse,
                currentLocation,
                currentDestWarehouse,
                currentDestLocation);
          }
          break;

        case 'deleteCartResult':
          // รับผลลัพธ์การลบตะกร้า
          bool success = data['success'] ?? false;
          String message = data['message'] ?? '';

          if (onCartDeleted != null) {
            onCartDeleted!(success, message);
          }
          break;
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error handling message: $e');
      }
      if (onError != null) {
        onError!('เกิดข้อผิดพลาดในการประมวลผลข้อมูล: ${e.toString()}');
      }
    }
  }

  void requestCartList(String shopId, {String status = 'CLOSED'}) {
    isLoading = true;
    channel?.sink.add(jsonEncode({
      'type': 'getCartList',
      'shopid': shopId,
      'status': status,
    }));
  }

  void requestCartDetails(String cartId, String shopId) {
    channel?.sink.add(jsonEncode({
      'type': 'getCartDetails',
      'cartId': cartId,
      'clientId': clientId,
      'shopid': shopId,
    }));
  }

  void deleteCart(String cartId, String shopId) {
    channel?.sink.add(jsonEncode({
      'type': 'deleteCart',
      'cartId': cartId,
      'clientId': clientId,
      'shopid': shopId,
    }));
  }

  bool isSocketConnected() {
    return isConnected;
  }
}
