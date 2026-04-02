# API Client — Dio, ApiResponse, GoAPI URLs

## Client Class

Always use `Client` from `lib/repositories/client.dart` — never create a new one.

```dart
import 'package:smlaicloud/repositories/client.dart';

// Usage in repository
final dio = Client().init();         // MainAPI (auth, general CRUD)
final dio = Client().initBiReport(); // BI Report API
```

Client injects Bearer token automatically via `ApiInterceptors`:
- Token from `global.appConfig.getString("token")`
- 401 -> triggers `TokenInvalid` event automatically -> redirects to login

## Response Types

### ApiResponse — for general endpoints

```dart
// Common fields
response.success   // bool
response.data      // dynamic (List or Map)
response.message   // String — error message
response.id        // String — inserted record id
response.page      // Page? — pagination info
response.uri       // String? — uploaded file URL

// Pagination
response.page?.page       // Current page
response.page?.perPage    // Items per page
response.page?.total      // Total count
response.page?.totalPage  // Total pages

// Parse from response
return ApiResponse.fromMap(response.data);
```

### GoApiQueryManyResultResponse — for multi-record queries

```dart
final res = GoApiQueryManyResultResponse.fromMap(response.data);
res.status   // int
res.results  // dynamic (List)
res.total    // int
```

### GoApiQueryOneResultResponse — for single-record queries

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

All requests go through a single backend URL — paths differentiated by prefix:

| Prefix | Used for |
|--------|----------|
| `goapi/api/...` | New Go backend endpoints |
| `goapi/get` | PostgreSQL SELECT query |
| `goapi/exec` | PostgreSQL execute |
| `goapi/genpdf` | PDF generation |
| `goapi/mcp/...` | MCP General tools (Flutter must not call directly) |
| `goapi/mcp/dev/...` | MCP Dev tools (Flutter must not call directly) |
| (no prefix) | MainAPI — auth, basic CRUD |

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
  final message = e.response?.data['message'] ?? e.message ?? 'An error occurred';
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

## Flavors and Environment

| Flavor | Environment | Entry Point |
|--------|-------------|-------------|
| `bcaidev` | localhost | `lib/main_bcaidev.dart` |
| `bcaiuat` | UAT server | `lib/main_bcaiuat.dart` |
| `bcaiprod` | Production | `lib/main_bcaiprod.dart` |

Backend URL configured via `web/config.json` -> `BackendUrlManager` -> `Client().init()` resolves automatically.
