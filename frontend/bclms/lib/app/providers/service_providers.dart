import 'package:flutter_riverpod/flutter_riverpod.dart';
import 'package:bclms/app/services/storage_service.dart';
import 'package:bclms/app/services/api_service.dart';
import 'package:bclms/app/services/auth_service.dart';
import 'package:bclms/app/modules/vehicle/vehicle_service.dart';
import 'package:bclms/app/modules/warehouse/warehouse_service.dart';

/// StorageService provider — override ด้วย instance จริงใน main.dart
final storageServiceProvider = Provider<StorageService>((ref) {
  throw UnimplementedError('StorageService must be initialized before use');
});

/// ApiService provider — override ด้วย instance จริงใน main.dart
final apiServiceProvider = Provider<ApiService>((ref) {
  throw UnimplementedError('ApiService must be initialized before use');
});

/// AuthService provider - depends on ApiService and StorageService
final authServiceProvider = Provider<AuthService>((ref) {
  final api = ref.watch(apiServiceProvider);
  final storage = ref.watch(storageServiceProvider);
  return AuthService(api, storage);
});

/// VehicleService provider - depends on ApiService
final vehicleServiceProvider = Provider<VehicleService>((ref) {
  final api = ref.watch(apiServiceProvider);
  return VehicleService(api);
});

/// WarehouseService provider - depends on ApiService
final warehouseServiceProvider = Provider<WarehouseService>((ref) {
  final api = ref.watch(apiServiceProvider);
  return WarehouseService(api);
});
