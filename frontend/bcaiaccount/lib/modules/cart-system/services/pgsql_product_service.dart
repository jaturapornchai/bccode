import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/product_search_model.dart';
import '../models/product_transaction_model.dart';
import '../../../global.dart' as global;
import '../../../utils/logger/app_logger.dart';
import 'thai_word_tokenizer.dart';
import 'transliteration_service.dart';

/// Service สำหรับค้นหาสินค้าจาก PostgreSQL API
/// ใช้ endpoint /pg/select แทน /clickhouse/query
class PgSQLProductService {
  final String baseUrl;
  late final TransliterationService _transliterationService;

  // Constants สำหรับ SELECT fields ที่ใช้ซ้ำ
  // Note: PostgreSQL productbarcode ไม่มี shopid column (ใช้ database per shop)
  static const String _PRODUCT_FIELDS =
      'itemcode, barcode, name0, unitcode, unitname, price1, price_retail, barcoderefunitstand as unitstand, barcoderefunitdivide as unitdivide';

  static const String _PRODUCT_TABLE = 'productbarcode';
  static const String _DOCDETAIL_TABLE = 'docdetail';

  PgSQLProductService({String? baseUrl}) : baseUrl = baseUrl ?? _getDefaultUrl() {
    _transliterationService = TransliterationService(baseUrl: this.baseUrl);
  }

  static String _getDefaultUrl() {
    // URL มาจากตั้งค่าระบบ (BackendUrlManager) เสมอ
    final path = global.goApiUrlPath('');
    final url = path.endsWith('/') ? path.substring(0, path.length - 1) : path;
    AppLogger.info('PostgreSQL Base URL: $url');
    return url;
  }

  /// Helper: Parse double from dynamic value (PostgreSQL อาจส่งตัวเลขเป็น String)
  static double _parseDouble(dynamic value) {
    if (value == null) return 0.0;
    if (value is num) return value.toDouble();
    if (value is String) return double.tryParse(value) ?? 0.0;
    return 0.0;
  }

  /// Helper method: Execute PostgreSQL query
  Future<Map<String, dynamic>> _executeQuery(
    String query, {
    required String shopId,
    Duration timeout = const Duration(seconds: 10),
    String context = 'query',
  }) async {
    try {
      final url = Uri.parse('$baseUrl/get');

      AppLogger.debug('🌐 [$context] API URL: $url');
      AppLogger.debug('📤 [$context] Query: $query');

      final response = await http
          .post(
            url,
            headers: {'Content-Type': 'application/json'},
            body: jsonEncode({
              'database': shopId,
              'queries': [query],
            }),
          )
          .timeout(timeout);

      AppLogger.debug('📥 [$context] Status: ${response.statusCode}');
      AppLogger.debug('📥 [$context] Response: ${response.body}');

      if (response.statusCode == 200) {
        final jsonData = jsonDecode(response.body);

        // PostgreSQL API returns results in different format than ClickHouse
        // { "status": "success", "database": "...", "results": [{ "query_id": 0, "rows": [...], "count": ... }] }
        if (jsonData['status'] == 'success' && jsonData['results'] != null) {
          final results = jsonData['results'] as List;
          if (results.isNotEmpty && results[0]['rows'] != null) {
            return {
              'status': 'success',
              'data': results[0]['rows'],
            };
          }
        }

        return {'status': 'error', 'message': 'Invalid response format'};
      } else {
        AppLogger.error('PostgreSQL API error: ${response.statusCode} - ${response.body}');
        return {
          'status': 'error',
          'code': response.statusCode,
          'message': response.body,
        };
      }
    } catch (e) {
      AppLogger.error('Exception in _executeQuery ($context): $e');
      return {'status': 'error', 'message': e.toString()};
    }
  }

  /// Helper method: Parse response เป็น ProductSearchResult
  ProductSearchResult _parseProductListResponse(
    Map<String, dynamic> jsonData,
    String context,
  ) {
    if (jsonData['status'] == 'success' && jsonData['data'] != null) {
      final List<dynamic> products = jsonData['data'];

      if (products.isEmpty) {
        AppLogger.info('[$context] No products found');
        return ProductSearchResult(status: 'success', count: 0, products: []);
      }

      final productList = products.map((p) => ProductSearchModel.fromJson(p)).toList();

      AppLogger.info('[$context] Found ${productList.length} product(s)');

      return ProductSearchResult(
        status: 'success',
        count: productList.length,
        products: productList,
      );
    }

    AppLogger.warning('[$context] Invalid response format');
    return ProductSearchResult(
      status: 'error',
      message: jsonData['message'] ?? 'Invalid response format',
      count: 0,
      products: [],
    );
  }

  /// ค้นหาสินค้าด้วย barcode
  Future<ProductSearchResult> searchByBarcode(String barcode, String shopId) async {
    try {
      AppLogger.info('Searching product by barcode: $barcode for shop: $shopId');

      final query =
          "SELECT $_PRODUCT_FIELDS FROM $_PRODUCT_TABLE "
          "WHERE barcode = '$barcode' LIMIT 1";

      final jsonData = await _executeQuery(
        query,
        shopId: shopId,
        context: 'searchByBarcode',
      );
      return _parseProductListResponse(jsonData, 'searchByBarcode');
    } catch (e) {
      AppLogger.error('Exception in searchByBarcode: $e');
      return ProductSearchResult(
        status: 'error',
        message: e.toString(),
        count: 0,
        products: [],
      );
    }
  }

  /// ค้นหาสินค้าด้วยชื่อ (Full Text Search)
  /// ค้นหาใน: name0, unitname, unitcode, barcode, itemcode
  Future<ProductSearchResult> searchByName(String keyword, String shopId) async {
    try {
      AppLogger.info('Searching products by name: "$keyword" for shop: $shopId');

      // ตัดคำเพื่อใส่ space ระหว่างคำก่อนแปลง
      final keywords = await ThaiWordTokenizer.tokenize(keyword);
      final keywordWithSpaces = keywords.join(' ');
      AppLogger.info(
        '🔪 Tokenized: "$keyword" → "$keywordWithSpaces" (${keywords.length} words)',
      );

      // Step 1: ค้นหาด้วยคำเดิม (ที่ป้อนมา)
      AppLogger.info('📍 Step 1: Search with original keyword...');
      final result1 = await _searchByNameInternal(keyword, shopId);
      AppLogger.info('✅ Step 1: Found ${result1.count} products');

      // Step 2: แปลง ไทย→อังกฤษ แล้วค้นหา
      ProductSearchResult result2 = ProductSearchResult(
        status: 'success',
        count: 0,
        products: [],
      );
      String translatedThToEn = '';
      List<String> keywordsThToEn = [];

      AppLogger.info('📍 Step 2: Translate Thai→English and search...');
      try {
        final thToEn = await _transliterationService.transliterateThaiToEnglish(
          keywordWithSpaces,
        );
        AppLogger.info(
          '🔄 Thai→English: "${thToEn.input}" → "${thToEn.output}" (${thToEn.matched} matches)',
        );

        if (thToEn.hasTranslation && thToEn.output != keywordWithSpaces) {
          translatedThToEn = thToEn.output;
          keywordsThToEn = await ThaiWordTokenizer.tokenize(translatedThToEn);
          result2 = await _searchByNameInternal(thToEn.output, shopId);
          AppLogger.info('✅ Step 2: Found ${result2.count} products');
        } else {
          AppLogger.info('⏭️  Step 2: No translation, skipped');
        }
      } catch (e) {
        AppLogger.warning('❌ Step 2 failed: $e');
      }

      // Step 3: แปลง อังกฤษ→ไทย แล้วค้นหา
      ProductSearchResult result3 = ProductSearchResult(
        status: 'success',
        count: 0,
        products: [],
      );
      String translatedEnToTh = '';
      List<String> keywordsEnToTh = [];

      AppLogger.info('📍 Step 3: Translate English→Thai and search...');
      try {
        final enToTh = await _transliterationService.transliterateEnglishToThai(
          keywordWithSpaces,
        );
        AppLogger.info(
          '🔄 English→Thai: "${enToTh.input}" → "${enToTh.output}" (${enToTh.matched} matches)',
        );

        if (enToTh.hasTranslation && enToTh.output != keywordWithSpaces) {
          translatedEnToTh = enToTh.output;
          keywordsEnToTh = await ThaiWordTokenizer.tokenize(translatedEnToTh);
          result3 = await _searchByNameInternal(enToTh.output, shopId);
          AppLogger.info('✅ Step 3: Found ${result3.count} products');
        } else {
          AppLogger.info('⏭️  Step 3: No translation, skipped');
        }
      } catch (e) {
        AppLogger.warning('❌ Step 3 failed: $e');
      }

      // Step 4: รวมผลลัพธ์ โดยใช้ itemcode+barcode เป็น key (ไม่ให้ซ้ำ)
      AppLogger.info('📍 Step 4: Merging and re-ranking results...');
      AppLogger.debug(
        '📊 Before merge: result1=${result1.products.length}, result2=${result2.products.length}, result3=${result3.products.length}',
      );
      final mergedProducts = <String, ProductSearchModel>{};

      // เพิ่มจาก result1 (ค้นหาด้วยคำเดิม)
      AppLogger.debug('➕ Adding from result1 (${result1.products.length} items)');
      for (final product in result1.products) {
        final key = '${product.itemCode}_${product.barcode}';
        mergedProducts[key] = product;
        AppLogger.debug('  ✓ $key | ${product.name0}');
      }

      // เพิ่มจาก result2 (ค้นหาด้วยคำแปลไทย→อังกฤษ)
      AppLogger.debug('➕ Adding from result2 (${result2.products.length} items)');
      for (final product in result2.products) {
        final key = '${product.itemCode}_${product.barcode}';
        if (!mergedProducts.containsKey(key)) {
          mergedProducts[key] = product;
          AppLogger.debug('  ✓ $key | ${product.name0} (new)');
        } else {
          AppLogger.debug('  ⊗ $key | ${product.name0} (duplicate, skipped)');
        }
      }

      // เพิ่มจาก result3 (ค้นหาด้วยคำแปลอังกฤษ→ไทย)
      AppLogger.debug('➕ Adding from result3 (${result3.products.length} items)');
      for (final product in result3.products) {
        final key = '${product.itemCode}_${product.barcode}';
        if (!mergedProducts.containsKey(key)) {
          mergedProducts[key] = product;
          AppLogger.debug('  ✓ $key | ${product.name0} (new)');
        } else {
          AppLogger.debug('  ⊗ $key | ${product.name0} (duplicate, skipped)');
        }
      }

      AppLogger.info('📊 After merge: ${mergedProducts.length} unique products');

      // Step 5: คำนวณคะแนนและเรียงใหม่
      AppLogger.info('📍 Step 5: Calculating scores for ranking...');

      final finalProducts = mergedProducts.values.toList();

      // รวมคำค้นหาทั้งหมด (เดิม + ที่แปล) เพื่อให้คะแนนครบถ้วน
      final allSearchKeywords = <String>[];
      allSearchKeywords.add(keyword);
      allSearchKeywords.addAll(keywords);
      if (translatedThToEn.isNotEmpty) {
        allSearchKeywords.add(translatedThToEn);
        allSearchKeywords.addAll(keywordsThToEn);
      }
      if (translatedEnToTh.isNotEmpty) {
        allSearchKeywords.add(translatedEnToTh);
        allSearchKeywords.addAll(keywordsEnToTh);
      }

      // คำนวณคะแนนสำหรับแต่ละสินค้า
      final scoredProducts =
          finalProducts.map((product) {
            final score = _calculateProductScore(
              product: product,
              originalKeyword: keyword,
              keywords: keywords,
              translatedKeywords: allSearchKeywords,
            );
            return {'product': product, 'score': score};
          }).toList();

      // เรียงตาม: 1) คะแนน (มากไปน้อย) 2) itemcode (A-Z) 3) อัตราส่วนหน่วยนับ (มากไปน้อย)
      scoredProducts.sort((a, b) {
        final productA = a['product'] as ProductSearchModel;
        final productB = b['product'] as ProductSearchModel;
        final scoreA = a['score'] as int;
        final scoreB = b['score'] as int;

        final scoreCompare = scoreB.compareTo(scoreA);
        if (scoreCompare != 0) return scoreCompare;

        final itemCodeCompare = productA.itemCode.compareTo(productB.itemCode);
        if (itemCodeCompare != 0) return itemCodeCompare;

        final ratioA = productA.unitStand / productA.unitDive;
        final ratioB = productB.unitStand / productB.unitDive;
        final ratioCompare = ratioB.compareTo(ratioA);
        if (ratioCompare != 0) return ratioCompare;

        return productA.barcode.compareTo(productB.barcode);
      });

      // แสดง Top 5 พร้อมคะแนน
      AppLogger.info('🏆 Top 5 After Re-ranking:');
      for (var i = 0; i < scoredProducts.length && i < 5; i++) {
        final item = scoredProducts[i];
        final p = item['product'] as ProductSearchModel;
        final score = item['score'] as int;
        AppLogger.info('  ${i + 1}. "${p.name0}" | Score: $score');
      }

      final rankedProducts =
          scoredProducts
              .map((item) => item['product'] as ProductSearchModel)
              .toList();

      AppLogger.info(
        '🎯 Ranked Result: ${rankedProducts.length} unique products (${result1.count} + ${result2.count} + ${result3.count} merged & ranked)',
      );

      // Step 6: ขยายผลลัพธ์ให้ครบทุก barcode ของแต่ละ itemcode
      AppLogger.info('📍 Step 6: Expanding to include all barcodes per itemcode...');

      final uniqueItemCodes =
          rankedProducts
              .map((p) => p.itemCode)
              .where((code) => code.isNotEmpty)
              .toSet();

      AppLogger.info(
        '📦 Found ${uniqueItemCodes.length} unique itemcodes, fetching all barcodes...',
      );

      final allUnitsMap = await getAllUnitsForMultipleProducts(
        uniqueItemCodes.toList(),
        shopId,
      );

      final expandedProductsMap = <String, ProductSearchModel>{};

      for (final entry in allUnitsMap.entries) {
        final itemCode = entry.key;
        final units = entry.value;

        AppLogger.debug('  ✓ ItemCode $itemCode: ${units.length} barcode(s)');

        for (final unit in units) {
          final key = '${unit.itemCode}_${unit.barcode}';
          expandedProductsMap[key] = unit;
        }
      }

      AppLogger.info(
        '✅ Expanded to ${expandedProductsMap.length} total products (all barcodes)',
      );

      final expandedProducts = expandedProductsMap.values.toList();

      expandedProducts.sort((a, b) {
        final itemCodeCompare = a.itemCode.compareTo(b.itemCode);
        if (itemCodeCompare != 0) return itemCodeCompare;

        final ratioA = a.unitStand / a.unitDive;
        final ratioB = b.unitStand / b.unitDive;
        final ratioCompare = ratioB.compareTo(ratioA);
        if (ratioCompare != 0) return ratioCompare;

        return a.barcode.compareTo(b.barcode);
      });

      AppLogger.info(
        '🎯 Final Result: ${expandedProducts.length} products (expanded from ${rankedProducts.length})',
      );

      return ProductSearchResult(
        status: 'success',
        count: expandedProducts.length,
        products: expandedProducts,
      );
    } catch (e) {
      AppLogger.error('Exception in searchByName: $e');
      return ProductSearchResult(
        status: 'error',
        message: e.toString(),
        count: 0,
        products: [],
      );
    }
  }

  /// Internal method สำหรับค้นหาด้วย keyword (ไม่มี transliteration)
  Future<ProductSearchResult> _searchByNameInternal(
    String keyword,
    String shopId,
  ) async {
    try {
      AppLogger.debug('Internal search with keyword: "$keyword" for shop: $shopId');

      final keywords = await ThaiWordTokenizer.tokenize(keyword);
      AppLogger.debug('Keywords for search: $keywords');

      final sanitizedKeywords = keywords.map((k) => k.replaceAll("'", "''")).toList();
      final sanitizedKeyword = keyword.replaceAll("'", "''").trim();
      final searchKeywordLower = sanitizedKeyword.toLowerCase();

      const combinedField =
          "itemcode || ' ' || COALESCE(barcode, '') || ' ' || COALESCE(name0, '') || ' ' || COALESCE(unitcode, '') || ' ' || COALESCE(unitname, '')";

      final whereConditions =
          sanitizedKeywords
              .map((kw) => "LOWER($combinedField) LIKE LOWER('%$kw%')")
              .join(' AND ');

      final whereClause =
          whereConditions.isNotEmpty
              ? whereConditions
              : "LOWER($combinedField) LIKE LOWER('%$searchKeywordLower%')";

      final firstKeyword = sanitizedKeywords.isNotEmpty ? sanitizedKeywords.first : '';

      final orderByClauses = <String>[];

      orderByClauses.add("CASE WHEN LOWER(barcode) = LOWER('$searchKeywordLower') THEN 0 ELSE 1 END");
      orderByClauses.add("CASE WHEN LOWER(itemcode) = LOWER('$searchKeywordLower') THEN 0 ELSE 1 END");
      orderByClauses.add("CASE WHEN LOWER(name0) = LOWER('$searchKeywordLower') THEN 0 ELSE 1 END");

      if (sanitizedKeywords.length > 1) {
        final conditions =
            sanitizedKeywords
                .map((kw) => "LOWER($combinedField) LIKE LOWER('%$kw%')")
                .join(' AND ');
        orderByClauses.add("CASE WHEN $conditions THEN 0 ELSE 1 END");
      }

      orderByClauses.add("CASE WHEN LOWER(barcode) LIKE LOWER('$searchKeywordLower%') THEN 0 ELSE 1 END");
      orderByClauses.add("CASE WHEN LOWER(itemcode) LIKE LOWER('$searchKeywordLower%') THEN 0 ELSE 1 END");
      orderByClauses.add("CASE WHEN LOWER(name0) LIKE LOWER('$searchKeywordLower%') THEN 0 ELSE 1 END");

      if (firstKeyword.isNotEmpty) {
        orderByClauses.add("CASE WHEN LOWER($combinedField) LIKE LOWER('$firstKeyword%') THEN 0 ELSE 1 END");
      }

      if (sanitizedKeywords.isNotEmpty) {
        final conditions =
            sanitizedKeywords
                .map((kw) => "CASE WHEN LOWER($combinedField) LIKE LOWER('%$kw%') THEN 1 ELSE 0 END")
                .join(' + ');
        orderByClauses.add("($conditions) DESC");
      }

      orderByClauses.add("CASE WHEN LOWER($combinedField) LIKE LOWER('%$searchKeywordLower%') THEN 0 ELSE 1 END");
      orderByClauses.add("LENGTH(name0) ASC");
      orderByClauses.add("itemcode ASC");
      orderByClauses.add("unitstand ASC");
      orderByClauses.add("unitdivide DESC");
      orderByClauses.add("name0 ASC");

      final orderByClause = orderByClauses.join(', ');

      final query =
          '''
        SELECT
          $_PRODUCT_FIELDS
        FROM $_PRODUCT_TABLE
        WHERE $whereClause
        ORDER BY $orderByClause
        LIMIT 500
      '''
              .trim()
              .replaceAll(RegExp(r'\s+'), ' ');

      AppLogger.debug('Full Text Search Query: $query');

      final jsonData = await _executeQuery(
        query,
        shopId: shopId,
        context: '_searchByNameInternal',
      );
      return _parseProductListResponse(jsonData, '_searchByNameInternal');
    } catch (e) {
      AppLogger.error('Exception in _searchByNameInternal: $e');
      return ProductSearchResult(
        status: 'error',
        message: e.toString(),
        count: 0,
        products: [],
      );
    }
  }

  /// ค้นหาสินค้าด้วย itemcode
  Future<ProductSearchResult> searchByItemCode(
    String itemCode,
    String shopId,
  ) async {
    try {
      AppLogger.info('Searching product by itemcode: $itemCode for shop: $shopId');

      final query =
          "SELECT $_PRODUCT_FIELDS FROM $_PRODUCT_TABLE "
          "WHERE itemcode = '$itemCode' LIMIT 1";

      final jsonData = await _executeQuery(
        query,
        shopId: shopId,
        context: 'searchByItemCode',
      );
      return _parseProductListResponse(jsonData, 'searchByItemCode');
    } catch (e) {
      AppLogger.error('Exception in searchByItemCode: $e');
      return ProductSearchResult(
        status: 'error',
        message: e.toString(),
        count: 0,
        products: [],
      );
    }
  }

  /// ค้นหาสินค้าแบบ Universal (ค้นหาทั้ง barcode, itemcode, name, unitcode, unitname)
  Future<ProductSearchResult> universalSearch(String keyword, String shopId) async {
    try {
      AppLogger.info('🔍 Universal search with keyword: "$keyword" for shop: $shopId');
      return await searchByName(keyword, shopId);
    } catch (e) {
      AppLogger.error('Exception in universalSearch: $e');
      return ProductSearchResult(
        status: 'error',
        message: e.toString(),
        count: 0,
        products: [],
      );
    }
  }

  /// ดึงสินค้ายอดนิยม พร้อมขยายให้ครบทุก barcode
  Future<ProductSearchResult> getPopularProducts(
    String shopId, {
    int limit = 300,
  }) async {
    try {
      AppLogger.info('Getting popular products for shop: $shopId (limit: $limit)');

      final query =
          "SELECT DISTINCT ON (itemcode) "
          "itemcode, barcode, name0, unitcode, unitname, price1, barcoderefunitstand as unitstand, barcoderefunitdivide as unitdivide "
          "FROM $_PRODUCT_TABLE "
          "ORDER BY itemcode, price1 DESC "
          "LIMIT $limit";

      final jsonData = await _executeQuery(
        query,
        shopId: shopId,
        timeout: const Duration(seconds: 10),
        context: 'getPopularProducts',
      );

      if (jsonData['status'] == 'success' && jsonData['data'] != null) {
        final List<dynamic> products = jsonData['data'];

        if (products.isEmpty) {
          AppLogger.info('No popular products found');
          return ProductSearchResult(status: 'success', count: 0, products: []);
        }

        final productList =
            products.map((p) => ProductSearchModel.fromJson(p)).toList();
        AppLogger.info('Found ${productList.length} popular itemcode(s)');

        AppLogger.info('📍 Expanding to include all barcodes per itemcode...');

        final uniqueItemCodes =
            productList
                .map((p) => p.itemCode)
                .where((code) => code.isNotEmpty)
                .toSet();

        AppLogger.info(
          '📦 Found ${uniqueItemCodes.length} unique itemcodes, fetching all barcodes...',
        );

        final allUnitsMap = await getAllUnitsForMultipleProducts(
          uniqueItemCodes.toList(),
          shopId,
        );

        final expandedProductsMap = <String, ProductSearchModel>{};

        for (final entry in allUnitsMap.entries) {
          final itemCode = entry.key;
          final units = entry.value;

          AppLogger.debug('  ✓ ItemCode $itemCode: ${units.length} barcode(s)');

          for (final unit in units) {
            final key = '${unit.itemCode}_${unit.barcode}';
            expandedProductsMap[key] = unit;
          }
        }

        AppLogger.info(
          '✅ Expanded to ${expandedProductsMap.length} total products (all barcodes)',
        );

        final expandedProducts = expandedProductsMap.values.toList();

        expandedProducts.sort((a, b) {
          final itemCodeCompare = a.itemCode.compareTo(b.itemCode);
          if (itemCodeCompare != 0) return itemCodeCompare;

          final ratioA = a.unitStand / a.unitDive;
          final ratioB = b.unitStand / b.unitDive;
          final ratioCompare = ratioB.compareTo(ratioA);
          if (ratioCompare != 0) return ratioCompare;

          return a.barcode.compareTo(b.barcode);
        });

        AppLogger.info(
          '🎯 Final Result: ${expandedProducts.length} products (expanded from ${productList.length})',
        );

        return ProductSearchResult(
          status: 'success',
          count: expandedProducts.length,
          products: expandedProducts,
        );
      }

      return ProductSearchResult(
        status: 'error',
        message: 'API returned error',
        count: 0,
        products: [],
      );
    } catch (e) {
      AppLogger.error('Exception in getPopularProducts: $e');
      return ProductSearchResult(
        status: 'error',
        message: e.toString(),
        count: 0,
        products: [],
      );
    }
  }

  /// คำนวณคะแนนความเกี่ยวข้องของสินค้ากับคำค้นหา
  int _calculateProductScore({
    required ProductSearchModel product,
    required String originalKeyword,
    required List<String> keywords,
    List<String>? translatedKeywords,
  }) {
    int score = 0;
    final keywordLower = originalKeyword.toLowerCase().trim();
    final name = product.name0.toLowerCase();
    final barcode = product.barcode.toLowerCase();
    final itemcode = product.itemCode.toLowerCase();
    final unitcode = product.unitCode.toLowerCase();

    final allKeywords = translatedKeywords ?? keywords;

    if (barcode == keywordLower) {
      score += 1000;
    } else if (itemcode == keywordLower) {
      score += 900;
    } else if (name == keywordLower) {
      score += 800;
    }

    if (allKeywords.length > 1) {
      final allKeywordsInName = allKeywords.every(
        (kw) => name.contains(kw.toLowerCase()),
      );
      if (allKeywordsInName) {
        score += 500;
      }

      final allKeywordsAnywhere = allKeywords.every((kw) {
        final kwLower = kw.toLowerCase();
        return name.contains(kwLower) ||
            barcode.contains(kwLower) ||
            itemcode.contains(kwLower) ||
            unitcode.contains(kwLower);
      });
      if (allKeywordsAnywhere) {
        score += 300;
      }
    }

    if (barcode.startsWith(keywordLower)) {
      score += 250;
    }
    if (itemcode.startsWith(keywordLower)) {
      score += 230;
    }
    if (name.startsWith(keywordLower)) {
      score += 220;
    }

    if (keywords.isNotEmpty) {
      final firstKeyword = keywords.first.toLowerCase();
      if (name.startsWith(firstKeyword)) {
        score += 200;
      }
    }

    for (final keyword in allKeywords) {
      final kw = keyword.toLowerCase();

      if (name.contains(kw)) {
        score += 50;
      }
      if (barcode.contains(kw)) {
        score += 40;
      }
      if (itemcode.contains(kw)) {
        score += 30;
      }
      if (unitcode.contains(kw)) {
        score += 20;
      }
    }

    if (name.length < 20) {
      score += 10;
    } else if (name.length < 40) {
      score += 5;
    }

    return score;
  }

  /// ดึงหน่วยนับทั้งหมดสำหรับหลาย itemcode ในครั้งเดียว (Batch Query)
  Future<Map<String, List<ProductSearchModel>>> getAllUnitsForMultipleProducts(
    List<String> itemCodes,
    String shopId, {
    Map<String, List<ProductSearchModel>>? unitsMap,
  }) async {
    try {
      if (itemCodes.isEmpty) {
        return {};
      }

      final itemCodesToFetch = <String>[];
      final result = <String, List<ProductSearchModel>>{};

      for (final itemCode in itemCodes) {
        if (unitsMap != null && unitsMap.containsKey(itemCode)) {
          result[itemCode] = unitsMap[itemCode]!;
        } else {
          itemCodesToFetch.add(itemCode);
        }
      }

      if (itemCodesToFetch.isEmpty) {
        AppLogger.debug(
          '✅ [getAllUnitsForMultipleProducts] All ${itemCodes.length} itemcodes found in cache',
        );
        return result;
      }

      AppLogger.info(
        '🌐 [getAllUnitsForMultipleProducts] Fetching ${itemCodesToFetch.length} itemcodes from API (${itemCodes.length - itemCodesToFetch.length} from cache)',
      );

      final itemCodesCondition =
          itemCodesToFetch.map((code) => "'$code'").join(',');

      final query =
          "SELECT $_PRODUCT_FIELDS "
          "FROM $_PRODUCT_TABLE "
          "WHERE itemcode IN ($itemCodesCondition) "
          "ORDER BY itemcode ASC, barcoderefunitstand DESC, barcoderefunitdivide ASC, barcode ASC";

      final jsonData = await _executeQuery(
        query,
        shopId: shopId,
        context: 'getAllUnitsForMultipleProducts',
      );

      if (jsonData['status'] == 'success' && jsonData['data'] != null) {
        final List<dynamic> products = jsonData['data'];

        for (final productJson in products) {
          final product = ProductSearchModel.fromJson(productJson);
          result.putIfAbsent(product.itemCode, () => []).add(product);
        }

        AppLogger.info(
          '✅ [getAllUnitsForMultipleProducts] Fetched ${products.length} units for ${result.length} itemcodes',
        );

        return result;
      }

      return result;
    } catch (e) {
      AppLogger.error('Exception in getAllUnitsForMultipleProducts: $e');
      return {};
    }
  }

  /// ดึงหน่วยนับทั้งหมดของสินค้าตาม itemcode
  Future<List<ProductSearchModel>> getAllUnitsForProduct(
    String itemCode,
    String shopId, {
    Map<String, List<ProductSearchModel>>? unitsMap,
  }) async {
    try {
      if (unitsMap != null && unitsMap.containsKey(itemCode)) {
        final cachedUnits = unitsMap[itemCode]!;
        AppLogger.debug(
          '📦 [getAllUnitsForProduct] Using cached data for $itemCode (${cachedUnits.length} units)',
        );
        return cachedUnits;
      }

      AppLogger.debug('🌐 [getAllUnitsForProduct] Fetching from API for $itemCode');

      final query =
          "SELECT $_PRODUCT_FIELDS "
          "FROM $_PRODUCT_TABLE "
          "WHERE itemcode = '$itemCode' "
          "ORDER BY barcoderefunitstand DESC, barcoderefunitdivide ASC, barcode ASC";

      final jsonData = await _executeQuery(
        query,
        shopId: shopId,
        context: 'getAllUnitsForProduct',
      );

      if (jsonData['status'] == 'success' && jsonData['data'] != null) {
        final List<dynamic> products = jsonData['data'];
        
        // Debug: แสดง raw data ก่อน parse
        if (products.isNotEmpty) {
          AppLogger.info('📦 [getAllUnitsForProduct] Raw data sample: ${products.first}');
        }
        
        final productList =
            products.map((p) => ProductSearchModel.fromJson(p)).toList();

        AppLogger.info(
          '✅ [getAllUnitsForProduct] Found ${productList.length} unit(s) for itemcode: $itemCode',
        );
        
        // Debug: แสดงข้อมูลราคาหลัง parse
        for (var i = 0; i < productList.length; i++) {
          final p = productList[i];
          AppLogger.info('   ${i + 1}. ${p.unitStand}:${p.unitDive} ${p.unitName} | Barcode: ${p.barcode} | Price: ${p.price1}');
        }

        return productList;
      }

      return [];
    } catch (e) {
      AppLogger.error('Exception in getAllUnitsForProduct: $e');
      return [];
    }
  }

  /// ดึงยอดคงเหลือของสินค้าหลายรายการพร้อมกัน
  Future<Map<String, ProductBalanceModel>> getProductBalances(
    List<String> itemCodes,
    String? shopId,
  ) async {
    try {
      if (itemCodes.isEmpty) {
        return {};
      }

      AppLogger.info(
        '📊 Getting balances for ${itemCodes.length} items${shopId != null ? ' (shop: $shopId)' : ' (all shops)'}',
      );

      final itemCodesCondition = itemCodes.map((code) => "'$code'").join(',');

      // transflag ที่เกี่ยวกับ stock (ไม่ใช้ iscalcstock เพราะอาจเป็น 0 ตลอด)
      // calcflag จัดการทิศทางแล้ว: +1 = รับเข้า, -1 = จ่ายออก
      const stockTransFlags = '54,12,310,48,60,58,66,44,16,56,68,72';
      final query =
          '''
        SELECT
          COALESCE(itemcode, '') as itemcode,
          SUM((totalqty * calcflag) * unitstand / unitdivide) as balance
        FROM $_DOCDETAIL_TABLE
        WHERE transflag IN ($stockTransFlags)
          AND itemcode IN ($itemCodesCondition)
        GROUP BY itemcode
      '''
              .trim()
              .replaceAll(RegExp(r'\s+'), ' ');

      final effectiveShopId = shopId ?? (itemCodes.isNotEmpty ? itemCodes.first : '');
      final jsonData = await _executeQuery(
        query,
        shopId: effectiveShopId,
        timeout: const Duration(seconds: 15),
        context: 'getProductBalances',
      );

      if (jsonData['status'] == 'success' && jsonData['data'] != null) {
        final List<dynamic> balances = jsonData['data'];
        final result = <String, ProductBalanceModel>{};

        for (final item in balances) {
          final itemCode = item['itemcode'] as String? ?? '';
          // PostgreSQL อาจส่งตัวเลขเป็น String
          final balance = _parseDouble(item['balance']);

          if (itemCode.isEmpty) continue;

          result[itemCode] = ProductBalanceModel(
            itemCode: itemCode,
            totalBalance: balance,
            balanceByShop: {}, // PostgreSQL ไม่มี shopid ใน query result
          );

          AppLogger.debug(
            '   ✓ $itemCode: Total=${balance.toStringAsFixed(2)}',
          );
        }

        AppLogger.info('✅ [getProductBalances] Got balances for ${result.length} items');

        return result;
      }

      return {};
    } catch (e) {
      AppLogger.error('Exception in getProductBalances: $e');
      return {};
    }
  }

  /// ดึงรายการซื้อล่าสุดของสินค้า (จาก docdetail)
  Future<List<ProductTransactionModel>> getRecentPurchases({
    required String itemCode,
    required String shopId,
    int limit = 5,
  }) async {
    try {
      AppLogger.info('🛒 Getting recent purchases for item: $itemCode (limit: $limit)');

      final query =
          '''
        SELECT
          docno,
          docdatetime,
          COALESCE(itemcode, '') as itemname,
          totalqty as qty,
          unitcode as unitname,
          price,
          (totalqty * price) as amount
        FROM $_DOCDETAIL_TABLE
        WHERE itemcode = '$itemCode'
          AND calcflag = 1
          AND transflag IN (54,12,310,48,60)
        ORDER BY docdatetime DESC, docno DESC
        LIMIT $limit
      '''
              .trim()
              .replaceAll(RegExp(r'\s+'), ' ');

      final jsonData = await _executeQuery(
        query,
        shopId: shopId,
        context: 'getRecentPurchases',
      );

      if (jsonData['status'] == 'success' && jsonData['data'] != null) {
        final List<dynamic> transactions = jsonData['data'];
        final result =
            transactions.map((t) => ProductTransactionModel.fromJson(t)).toList();

        AppLogger.info('✅ Found ${result.length} recent purchase(s) for $itemCode');

        return result;
      }

      return [];
    } catch (e) {
      AppLogger.error('Exception in getRecentPurchases: $e');
      return [];
    }
  }

  /// ดึงรายการขายล่าสุดของสินค้า (จาก docdetail)
  Future<List<ProductTransactionModel>> getRecentSales({
    required String itemCode,
    required String shopId,
    int limit = 5,
  }) async {
    try {
      AppLogger.info('💰 Getting recent sales for item: $itemCode (limit: $limit)');

      final query =
          '''
        SELECT
          docno,
          docdatetime,
          COALESCE(itemcode, '') as itemname,
          totalqty as qty,
          unitcode as unitname,
          price,
          (totalqty * price) as amount
        FROM $_DOCDETAIL_TABLE
        WHERE itemcode = '$itemCode'
          AND calcflag = -1
          AND transflag IN (58,66,44,16,56,68,72)
        ORDER BY docdatetime DESC, docno DESC
        LIMIT $limit
      '''
              .trim()
              .replaceAll(RegExp(r'\s+'), ' ');

      final jsonData = await _executeQuery(
        query,
        shopId: shopId,
        context: 'getRecentSales',
      );

      if (jsonData['status'] == 'success' && jsonData['data'] != null) {
        final List<dynamic> transactions = jsonData['data'];
        final result =
            transactions.map((t) => ProductTransactionModel.fromJson(t)).toList();

        AppLogger.info('✅ Found ${result.length} recent sale(s) for $itemCode');

        return result;
      }

      return [];
    } catch (e) {
      AppLogger.error('Exception in getRecentSales: $e');
      return [];
    }
  }

  /// ดึงยอดคงเหลือแบบละเอียด แยกตาม warehouse และ location
  /// Note: ในระบบปัจจุบัน shopid อาจใช้แทน warehouse
  /// ถ้ามี field location ใน docdetail ให้ดึงด้วย
  ///
  /// ⚠️ **REAL-TIME DATA - NO CACHE**
  /// ยอดคงเหลือต้องเป็น real-time เสมอ ไม่ cache เพราะเปลี่ยนแปลงตลอดเวลา
  Future<Map<String, ProductWarehouseBalanceModel>> getDetailedBalances({
    required String itemCode,
    required String shopId,
  }) async {
    try {
      AppLogger.info('📦 Getting detailed balances for item: $itemCode (shop: $shopId)');

      // Query สำหรับดึงยอดคงเหลือแยกตาม warehouse (whcode) และ location (locationcode)
      // ใช้ transflag แทน iscalcstock เพราะ iscalcstock อาจเป็น 0 ตลอด
      const stockTransFlags = '54,12,310,48,60,58,66,44,16,56,68,72';
      final query =
          '''
        SELECT
          COALESCE(whcode, '') as warehouse_id,
          COALESCE(whcode, '') as warehouse_name,
          COALESCE(locationcode, '') as location_id,
          COALESCE(locationcode, '') as location_name,
          SUM((totalqty * calcflag) * unitstand / unitdivide) as balance
        FROM $_DOCDETAIL_TABLE
        WHERE itemcode = '$itemCode'
          AND transflag IN ($stockTransFlags)
        GROUP BY whcode, locationcode
        ORDER BY whcode, locationcode
      '''
              .trim()
              .replaceAll(RegExp(r'\s+'), ' ');

      final jsonData = await _executeQuery(
        query,
        shopId: shopId,
        context: 'getDetailedBalances',
      );

      if (jsonData['status'] == 'success' && jsonData['data'] != null) {
        final List<dynamic> balances = jsonData['data'];

        AppLogger.info('📦 Raw balance data from query:');
        for (final item in balances) {
          AppLogger.info(
            '  - warehouse_id: ${item['warehouse_id']?.toString() ?? ''}, warehouse_name: ${item['warehouse_name']?.toString() ?? ''}, location_id: ${item['location_id']?.toString() ?? ''}, location_name: ${item['location_name']?.toString() ?? ''}, balance: ${item['balance']?.toString() ?? ''}',
          );
        }

        // จัดกลุ่มตาม warehouse
        final Map<String, List<Map<String, dynamic>>> groupedByWarehouse = {};

        for (final item in balances) {
          final warehouseId = item['warehouse_id'] as String? ?? '';
          if (!groupedByWarehouse.containsKey(warehouseId)) {
            groupedByWarehouse[warehouseId] = [];
          }
          groupedByWarehouse[warehouseId]!.add(item);
        }

        // สร้าง ProductWarehouseBalanceModel
        final result = <String, ProductWarehouseBalanceModel>{};

        for (final entry in groupedByWarehouse.entries) {
          final warehouseId = entry.key;
          final items = entry.value;

          // สร้าง location balances
          final locationBalances = <String, ProductLocationBalanceModel>{};
          double totalBalance = 0.0;

          for (final item in items) {
            final locationId = item['location_id'] as String? ?? '';
            final locationNameRaw = item['location_name'] as String? ?? '';
            // ถ้า locationName เป็นค่าว่าง ให้ใช้ "ทั่วไป" แทน
            final locationName = locationNameRaw.isEmpty ? global.language("general_location") : locationNameRaw;
            // PostgreSQL อาจส่งตัวเลขเป็น String
            final balance = _parseDouble(item['balance']);

            totalBalance += balance;

            if (locationId.isNotEmpty || balance != 0) {
              final locKey = locationId.isEmpty ? 'default' : locationId;
              locationBalances[locKey] = ProductLocationBalanceModel(
                locationId: locationId,
                locationName: locationName,
                balance: balance,
              );
            }
          }

          final warehouseNameRaw = items.isNotEmpty ? (items.first['warehouse_name'] as String? ?? '') : '';
          // ถ้า warehouseName เป็นค่าว่าง ให้ใช้ "คลังหลัก" แทน
          final warehouseName = warehouseNameRaw.isEmpty ? global.language("default_warehouse") : warehouseNameRaw;

          result[warehouseId] = ProductWarehouseBalanceModel(
            warehouseId: warehouseId,
            warehouseName: warehouseName,
            totalBalance: totalBalance,
            locationBalances: locationBalances,
          );

          AppLogger.info(
            '🏭 Warehouse: $warehouseId (name: $warehouseName), total: $totalBalance, locations: ${locationBalances.length}',
          );
        }

        AppLogger.info('✅ Found detailed balances for ${result.length} warehouse(s)');

        return result;
      }

      return {};
    } catch (e) {
      AppLogger.error('Exception in getDetailedBalances: $e');
      return {};
    }
  }
}
