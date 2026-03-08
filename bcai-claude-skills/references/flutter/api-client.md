# API Client — Dio, ApiResponse, GoAPI URLs

## Client Class

ใช้ `Client` จาก `lib/repositories/client.dart` เสมอ — ห้ามสร้างใหม่

```dart
import 'package:smlaicloud/repositories/client.dart';

// ใช้ใน repository
final dio = Client().init();         // MainAPI (auth, CRUD ทั่วไป)
final dio = Client().initBiReport(); // BI Report API
```

Client inject Bearer token อัตโนมัติผ่าน `ApiInterceptors`
- Token มาจาก `global.appConfig.getString("token")`
- 401 → trigger `TokenInvalid` event อัตโนมัติ → redirect ไป login

## Response Types

### ApiResponse — สำหรับ endpoint ทั่วไป

```dart
// fields ที่ใช้บ่อย
response.success   // bool
response.data      // dynamic (List หรือ Map)
response.message   // String — error message
response.id        // String — id ของ record ที่ insert
response.page      // Page? — pagination info
response.uri       // String? — URL ของ file ที่ upload

// Pagination
response.page?.page       // หน้าปัจจุบัน
response.page?.perPage    // จำนวนต่อหน้า
response.page?.total      // จำนวนทั้งหมด
response.page?.totalPage  // จำนวนหน้าทั้งหมด

// parse จาก response
return ApiResponse.fromMap(response.data);
```

### GoApiQueryManyResultResponse — สำหรับ query หลาย records

```dart
final res = GoApiQueryManyResultResponse.fromMap(response.data);
res.status   // int
res.results  // dynamic (List)
res.total    // int
```

### GoApiQueryOneResultResponse — สำหรับ query 1 record

```dart
final res = GoApiQueryOneResultResponse.fromMap(response.data);
res.status  // int
res.result  // dynamic (Map)
```

## Repository Pattern

```dart
class FeatureRepository {
  Future<ApiResponse> getList({
    required String shopId,
    int offset = 0,
    int limit = 20,
    String search = '',
  }) async {
    final dio = Client().init();
    final response = await dio.post(
      'goapi/api/feature/list',
      data: {'shopid': shopId, 'offset': offset, 'limit': limit, 'search': search},
    );
    return ApiResponse.fromMap(response.data);
  }

  Future<ApiResponse> save(FeatureModel item) async {
    final dio = Client().init();
    final response = await dio.post('goapi/api/feature/insert', data: item.toJson());
    return ApiResponse.fromMap(response.data);
  }

  Future<ApiResponse> update(String guid, FeatureModel item) async {
    final dio = Client().init();
    final response = await dio.put('goapi/api/feature/$guid', data: item.toJson());
    return ApiResponse.fromMap(response.data);
  }

  Future<ApiResponse> delete(String guid) async {
    final dio = Client().init();
    final response = await dio.delete('goapi/api/feature/$guid');
    return ApiResponse.fromMap(response.data);
  }
}
```

## GoAPI URL Patterns

ทุก request ผ่าน Backend URL เดียว — path แยกด้วย prefix:

| Prefix | ใช้กับ |
|--------|--------|
| `goapi/api/...` | Go backend endpoints ใหม่ |
| `goapi/get` | PostgreSQL SELECT query |
| `goapi/exec` | PostgreSQL execute |
| `goapi/genpdf` | สร้าง PDF |
| `goapi/mcp/...` | MCP General tools (ห้าม Flutter เรียกตรง) |
| `goapi/mcp/dev/...` | MCP Dev tools (ห้าม Flutter เรียกตรง) |
| (ไม่มี prefix) | MainAPI — auth, CRUD พื้นฐาน |

## Error Handling

```dart
try {
  final result = await _repository.save(item);
  if (result.success) {
    emit(FeatureSaveSuccess());
  } else {
    emit(FeatureSaveFailed(message: result.message));
  }
} on DioException catch (e) {
  final message = e.response?.data['message'] ?? e.message ?? 'เกิดข้อผิดพลาด';
  emit(FeatureSaveFailed(message: message));
} catch (e) {
  try {
    final error = jsonDecode(e.toString());
    emit(FeatureSaveFailed(message: error['message']));
  } catch (_) {
    emit(FeatureSaveFailed(message: e.toString()));
  }
}
```

## Authentication Flow

```dart
import 'package:smlaicloud/global.dart' as global;

final shopId = global.appConfig.getString("shopid") ?? '';
final token  = global.appConfig.getString("token") ?? '';
```

## Flavors และ Environment

| Flavor | ใช้กับ | Entry point |
|--------|--------|-------------|
| `bcaidev` | localhost | `lib/main_bcaidev.dart` |
| `bcaiuat` | UAT server | `lib/main_bcaiuat.dart` |
| `bcaiprod` | Production | `lib/main_bcaiprod.dart` |

Backend URL ตั้งค่าผ่าน `web/config.json` → `BackendUrlManager` → `Client().init()` ดึงอัตโนมัติ
