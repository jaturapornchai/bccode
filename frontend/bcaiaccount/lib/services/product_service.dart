import 'dart:async';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/repositories/product_barcode_repository.dart';

class ProductService {
  final ProductBarcodeRepository _repository;
  final Map<String, List<ProductBarcodeModel>> _cache = {};
  final Map<String, DateTime> _cacheTimestamps = {};
  static const Duration _cacheTimeout = Duration(minutes: 5);
  static const int _maxCacheSize = 1000;

  // Cancellation tokens for ongoing requests
  final Map<String, CancelToken> _pendingRequests = {};

  ProductService(this._repository);

  /// Get products with caching and cancellation support
  Future<List<ProductBarcodeModel>> getProducts({
    required int offset,
    required int limit,
    required String search,
    required String branchcode,
    required String businesstypecode,
    bool useCache = true,
    CancelToken? cancelToken,
  }) async {
    final cacheKey = _generateCacheKey(
      offset,
      limit,
      search,
      branchcode,
      businesstypecode,
    );

    // Cancel any existing request for the same key
    _cancelRequest(cacheKey);

    // Create new cancel token if not provided
    final token = cancelToken ?? CancelToken();
    _pendingRequests[cacheKey] = token;

    try {
      // Check cache first
      if (useCache && _isCacheValid(cacheKey)) {
        _pendingRequests.remove(cacheKey);
        return _cache[cacheKey]!;
      }

      // Make API call with retry mechanism
      final result = await _makeRequestWithRetry(
        () => _repository.getProductBarcodeList(
          offset: offset,
          limit: limit,
          search: search,
          branchcode: branchcode,
          businesstypecode: businesstypecode,
        ),
        token,
        maxRetries: 3,
      );

      if (result.success && !token.isCancelled) {
        final products = (result.data as List)
            .map((product) => ProductBarcodeModel.fromJson(product))
            .toList();

        // Update cache
        _updateCache(cacheKey, products);

        _pendingRequests.remove(cacheKey);
        return products;
      } else {
        _pendingRequests.remove(cacheKey);
        throw Exception('Failed to load products');
      }
    } catch (e) {
      _pendingRequests.remove(cacheKey);
      if (token.isCancelled) {
        throw const CancelledException();
      }
      rethrow;
    }
  }

  /// Make API request with retry mechanism
  Future<T> _makeRequestWithRetry<T>(
    Future<T> Function() requestFunction,
    CancelToken cancelToken, {
    int maxRetries = 3,
  }) async {
    int retryCount = 0;

    while (retryCount < maxRetries) {
      if (cancelToken.isCancelled) {
        throw const CancelledException();
      }

      try {
        return await requestFunction();
      } catch (e) {
        retryCount++;

        if (retryCount >= maxRetries || cancelToken.isCancelled) {
          rethrow;
        }

        // Exponential backoff
        final delay = Duration(milliseconds: 500 * (1 << (retryCount - 1)));
        await Future.delayed(delay);
      }
    }

    throw Exception('Max retries exceeded');
  }

  /// Generate cache key
  String _generateCacheKey(
    int offset,
    int limit,
    String search,
    String branchcode,
    String businesstypecode,
  ) {
    return 'products_${offset}_${limit}_${search}_${branchcode}_$businesstypecode';
  }

  /// Check if cache is valid
  bool _isCacheValid(String cacheKey) {
    if (!_cache.containsKey(cacheKey)) return false;

    final timestamp = _cacheTimestamps[cacheKey];
    if (timestamp == null) return false;

    return DateTime.now().difference(timestamp) < _cacheTimeout;
  }

  /// Update cache with size management
  void _updateCache(String cacheKey, List<ProductBarcodeModel> products) {
    // Remove oldest entries if cache is full
    if (_cache.length >= _maxCacheSize) {
      _cleanupCache();
    }

    _cache[cacheKey] = products;
    _cacheTimestamps[cacheKey] = DateTime.now();
  }

  /// Cleanup old cache entries
  void _cleanupCache() {
    final entries = _cacheTimestamps.entries.toList()
      ..sort((a, b) => a.value.compareTo(b.value));

    // Remove oldest 20% of entries
    final removeCount = (entries.length * 0.2).ceil();
    for (int i = 0; i < removeCount && i < entries.length; i++) {
      final key = entries[i].key;
      _cache.remove(key);
      _cacheTimestamps.remove(key);
    }
  }

  /// Cancel request
  void _cancelRequest(String cacheKey) {
    final token = _pendingRequests[cacheKey];
    if (token != null && !token.isCancelled) {
      token.cancel();
      _pendingRequests.remove(cacheKey);
    }
  }

  /// Cancel all pending requests
  void cancelAllRequests() {
    for (final token in _pendingRequests.values) {
      if (!token.isCancelled) {
        token.cancel();
      }
    }
    _pendingRequests.clear();
  }

  /// Clear cache
  void clearCache() {
    _cache.clear();
    _cacheTimestamps.clear();
  }

  /// Dispose service
  void dispose() {
    cancelAllRequests();
    clearCache();
  }
}

/// Simple cancellation token implementation
class CancelToken {
  bool _isCancelled = false;

  bool get isCancelled => _isCancelled;

  void cancel() {
    _isCancelled = true;
  }
}

/// Exception thrown when operation is cancelled
class CancelledException implements Exception {
  const CancelledException();

  @override
  String toString() => 'Operation was cancelled';
}
