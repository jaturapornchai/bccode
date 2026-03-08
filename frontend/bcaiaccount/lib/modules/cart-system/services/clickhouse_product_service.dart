import 'dart:convert';
import 'package:http/http.dart' as http;
import '../models/product_search_model.dart';
import '../models/product_transaction_model.dart';
import '../../../global.dart' as global;
import '../../../utils/logger/app_logger.dart';
import 'thai_word_tokenizer.dart';
import 'transliteration_service.dart';

/// Service สำหรบคนหาสนคาจาก ClickHouse API
class ClickHouseProductService {
  final String baseUrl;
  late final TransliterationService _transliterationService;

  // ✅ Constants สำหรับ SELECT fields ที่ใช้ซ้ำ
  static const String _PRODUCT_FIELDS = 'shopid, itemcode, barcode, name0, unitcode, unitname, price1, unitstand, unitdivide, imageuri';

  static const String _PRODUCT_TABLE = '${global.clickHouseDatabaseName}.productbarcode';
  static const String _DOCDETAIL_TABLE = '${global.clickHouseDatabaseName}.docdetail';

  ClickHouseProductService({String? baseUrl}) : baseUrl = baseUrl ?? _getDefaultUrl() {
    // Initialize transliteration service in constructor body
    _transliterationService = TransliterationService(baseUrl: this.baseUrl);
  }

  static String _getDefaultUrl() {
    // URL มาจากตั้งค่าระบบ (BackendUrlManager) เสมอ
    final path = global.goApiUrlPath('');
    final url = path.endsWith('/') ? path.substring(0, path.length - 1) : path;
    AppLogger.info('ClickHouse Base URL: $url');
    return url;
  }

  // ✅ Helper method: Execute ClickHouse query
  Future<Map<String, dynamic>> _executeQuery(String query, {Duration timeout = const Duration(seconds: 10), String context = 'query'}) async {
    try {
      final url = Uri.parse('$baseUrl/clickhouse/query');

      AppLogger.debug('🌐 [$context] API URL: $url');
      AppLogger.debug('📤 [$context] Query: $query');

      final response = await http.post(url, headers: {'Content-Type': 'application/json'}, body: jsonEncode({'query': query})).timeout(timeout);

      AppLogger.debug('📥 [$context] Status: ${response.statusCode}');
      AppLogger.debug('📥 [$context] Response: ${response.body}');

      if (response.statusCode == 200) {
        return jsonDecode(response.body);
      } else {
        AppLogger.error('ClickHouse API error: ${response.statusCode} - ${response.body}');
        return {'status': 'error', 'code': response.statusCode, 'message': response.body};
      }
    } catch (e) {
      AppLogger.error('Exception in _executeQuery ($context): $e');
      return {'status': 'error', 'message': e.toString()};
    }
  }

  // ✅ Helper method: Parse response เป็น ProductSearchResult
  ProductSearchResult _parseProductListResponse(Map<String, dynamic> jsonData, String context) {
    if (jsonData['status'] == 'success' && jsonData['data'] != null) {
      final List<dynamic> products = jsonData['data'];

      if (products.isEmpty) {
        AppLogger.info('[$context] No products found');
        return ProductSearchResult(status: 'success', count: 0, products: []);
      }

      final productList = products.map((p) => ProductSearchModel.fromJson(p)).toList();

      AppLogger.info('[$context] Found ${productList.length} product(s)');

      return ProductSearchResult(status: 'success', count: productList.length, products: productList);
    }

    AppLogger.warning('[$context] Invalid response format');
    return ProductSearchResult(status: 'error', message: jsonData['message'] ?? 'Invalid response format', count: 0, products: []);
  }

  /// คนหาสนคาดวย barcode
  Future<ProductSearchResult> searchByBarcode(String barcode, String shopId) async {
    try {
      AppLogger.info('Searching product by barcode: $barcode for shop: $shopId');

      final query =
          "SELECT $_PRODUCT_FIELDS FROM $_PRODUCT_TABLE "
          "WHERE shopid = '$shopId' AND barcode = '$barcode' LIMIT 1";

      final jsonData = await _executeQuery(query, context: 'searchByBarcode');
      return _parseProductListResponse(jsonData, 'searchByBarcode');
    } catch (e) {
      AppLogger.error('Exception in searchByBarcode: $e');
      return ProductSearchResult(status: 'error', message: e.toString(), count: 0, products: []);
    }
  }

  /// ค้นหาสินค้าด้วยชื่อ (Full Text Search)
  /// ค้นหาใน: name0, unitname, unitcode, barcode, itemcode
  ///
  /// วิธีการค้นหา:
  /// 1. User ป้อนคำค้น (ไทย/อังกฤษ/ปนกัน)
  /// 2. ตัดคำ แล้วค้นหาด้วยคำที่ป้อน → ผล 1
  /// 3. เอาคำที่ตัดไปแปลเป็นอังกฤษ แล้วค้นหา → ผล 2
  /// 4. เอาคำที่ตัดไปแปลเป็นไทย แล้วค้นหา → ผล 3
  /// 5. รวม ผล 1,2,3 ด้วย itemcode+barcode (ไม่ซ้ำ)
  Future<ProductSearchResult> searchByName(String keyword, String shopId) async {
    try {
      AppLogger.info('Searching products by name: "$keyword" for shop: $shopId');

      // ตัดคำเพื่อใส่ space ระหว่างคำก่อนแปลง
      final keywords = await ThaiWordTokenizer.tokenize(keyword);
      final keywordWithSpaces = keywords.join(' ');
      AppLogger.info('🔪 Tokenized: "$keyword" → "$keywordWithSpaces" (${keywords.length} words)');

      // Step 1: ค้นหาด้วยคำเดิม (ที่ป้อนมา)
      AppLogger.info('📍 Step 1: Search with original keyword...');
      final result1 = await _searchByNameInternal(keyword, shopId);
      AppLogger.info('✅ Step 1: Found ${result1.count} products');

      // Step 2: แปลง ไทย→อังกฤษ แล้วค้นหา
      ProductSearchResult result2 = ProductSearchResult(status: 'success', count: 0, products: []);
      String translatedThToEn = ''; // เก็บคำที่แปลไว้
      List<String> keywordsThToEn = []; // เก็บ keywords ที่แปลไว้

      AppLogger.info('📍 Step 2: Translate Thai→English and search...');
      try {
        final thToEn = await _transliterationService.transliterateThaiToEnglish(
          keywordWithSpaces, // ใช้คำที่มี space
        );
        AppLogger.info('🔄 Thai→English: "${thToEn.input}" → "${thToEn.output}" (${thToEn.matched} matches)');

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
      ProductSearchResult result3 = ProductSearchResult(status: 'success', count: 0, products: []);
      String translatedEnToTh = ''; // เก็บคำที่แปลไว้
      List<String> keywordsEnToTh = []; // เก็บ keywords ที่แปลไว้

      AppLogger.info('📍 Step 3: Translate English→Thai and search...');
      try {
        final enToTh = await _transliterationService.transliterateEnglishToThai(
          keywordWithSpaces, // ใช้คำที่มี space
        );
        AppLogger.info('🔄 English→Thai: "${enToTh.input}" → "${enToTh.output}" (${enToTh.matched} matches)');

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
      AppLogger.debug('📊 Before merge: result1=${result1.products.length}, result2=${result2.products.length}, result3=${result3.products.length}');
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
      AppLogger.debug('🔑 All unique keys:');
      for (final key in mergedProducts.keys) {
        AppLogger.debug('  - $key');
      }

      // Step 5: คำนวณคะแนนและเรียงใหม่
      AppLogger.info('📍 Step 5: Calculating scores for ranking...');

      final finalProducts = mergedProducts.values.toList();

      // รวมคำค้นหาทั้งหมด (เดิม + ที่แปล) เพื่อให้คะแนนครบถ้วน
      final allSearchKeywords = <String>[];
      allSearchKeywords.add(keyword); // คำเดิม
      allSearchKeywords.addAll(keywords); // keywords จากคำเดิม
      if (translatedThToEn.isNotEmpty) {
        allSearchKeywords.add(translatedThToEn); // คำแปลไทย→อังกฤษ
        allSearchKeywords.addAll(keywordsThToEn); // keywords จากคำแปล
      }
      if (translatedEnToTh.isNotEmpty) {
        allSearchKeywords.add(translatedEnToTh); // คำแปลอังกฤษ→ไทย
        allSearchKeywords.addAll(keywordsEnToTh); // keywords จากคำแปล
      }

      // คำนวณคะแนนสำหรับแต่ละสินค้า
      final scoredProducts = finalProducts.map((product) {
        final score = _calculateProductScore(
          product: product,
          originalKeyword: keyword,
          keywords: keywords,
          translatedKeywords: allSearchKeywords, // ส่งคำที่แปลไปด้วย
        );
        return {'product': product, 'score': score};
      }).toList();

      // เรียงตาม: 1) คะแนน (มากไปน้อย) 2) itemcode (A-Z) 3) อัตราส่วนหน่วยนับ (มากไปน้อย)
      scoredProducts.sort((a, b) {
        final productA = a['product'] as ProductSearchModel;
        final productB = b['product'] as ProductSearchModel;
        final scoreA = a['score'] as int;
        final scoreB = b['score'] as int;

        // 1. เรียงตามคะแนนก่อน (มากไปน้อย)
        final scoreCompare = scoreB.compareTo(scoreA);
        if (scoreCompare != 0) return scoreCompare;

        // 2. ถ้าคะแนนเท่ากัน เรียงตาม itemcode (A-Z)
        final itemCodeCompare = productA.itemCode.compareTo(productB.itemCode);
        if (itemCodeCompare != 0) return itemCodeCompare;

        // 3. ถ้า itemcode เท่ากัน เรียงตามอัตราส่วนหน่วยนับ (มากไปน้อย)
        // อัตราส่วน = unitstand / unitdivide
        final ratioA = productA.unitStand / productA.unitDive;
        final ratioB = productB.unitStand / productB.unitDive;
        final ratioCompare = ratioB.compareTo(ratioA); // มากไปน้อย
        if (ratioCompare != 0) return ratioCompare;

        // 4. ถ้าอัตราส่วนเท่ากัน เรียงตาม barcode (A-Z)
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

      final rankedProducts = scoredProducts.map((item) => item['product'] as ProductSearchModel).toList();

      AppLogger.info('🎯 Ranked Result: ${rankedProducts.length} unique products (${result1.count} + ${result2.count} + ${result3.count} merged & ranked)');

      // Step 6: ขยายผลลัพธ์ให้ครบทุก barcode ของแต่ละ itemcode
      AppLogger.info('📍 Step 6: Expanding to include all barcodes per itemcode...');

      // เก็บ unique itemcodes จากผลลัพธ์
      final uniqueItemCodes = rankedProducts.map((p) => p.itemCode).where((code) => code.isNotEmpty).toSet();

      AppLogger.info('📦 Found ${uniqueItemCodes.length} unique itemcodes, fetching all barcodes...');

      // ✅ ดึงทุก barcode ครั้งเดียว (Batch Query)
      final allUnitsMap = await getAllUnitsForMultipleProducts(uniqueItemCodes.toList(), shopId);

      // สร้าง map จากผลลัพธ์
      final expandedProductsMap = <String, ProductSearchModel>{};

      for (final entry in allUnitsMap.entries) {
        final itemCode = entry.key;
        final units = entry.value;

        AppLogger.debug('  ✓ ItemCode $itemCode: ${units.length} barcode(s)');

        // เพิ่มทุก barcode เข้า map
        for (final unit in units) {
          final key = '${unit.itemCode}_${unit.barcode}';
          expandedProductsMap[key] = unit;
        }
      }

      AppLogger.info('✅ Expanded to ${expandedProductsMap.length} total products (all barcodes)');

      // แสดงรายละเอียดทุก barcode ที่ได้จาก expansion
      AppLogger.debug('📋 All barcodes after expansion:');
      for (final key in expandedProductsMap.keys) {
        final product = expandedProductsMap[key]!;
        AppLogger.debug('  - $key | ${product.unitStand}:${product.unitDive} ${product.unitName}');
      }

      // แปลง map กลับเป็น list และเรียงใหม่
      final expandedProducts = expandedProductsMap.values.toList();

      // เรียงตาม: 1) itemcode (A-Z) 2) อัตราส่วน (มาก→น้อย) 3) barcode (A-Z)
      expandedProducts.sort((a, b) {
        // 1. เรียงตาม itemcode
        final itemCodeCompare = a.itemCode.compareTo(b.itemCode);
        if (itemCodeCompare != 0) return itemCodeCompare;

        // 2. เรียงตามอัตราส่วน (มาก→น้อย)
        final ratioA = a.unitStand / a.unitDive;
        final ratioB = b.unitStand / b.unitDive;
        final ratioCompare = ratioB.compareTo(ratioA);
        if (ratioCompare != 0) return ratioCompare;

        // 3. เรียงตาม barcode
        return a.barcode.compareTo(b.barcode);
      });

      AppLogger.info('🎯 Final Result: ${expandedProducts.length} products (expanded from ${rankedProducts.length})');

      return ProductSearchResult(status: 'success', count: expandedProducts.length, products: expandedProducts);
    } catch (e) {
      AppLogger.error('Exception in searchByName: $e');
      return ProductSearchResult(status: 'error', message: e.toString(), count: 0, products: []);
    }
  }

  /// Internal method สำหรับค้นหาด้วย keyword (ไม่มี transliteration)
  Future<ProductSearchResult> _searchByNameInternal(String keyword, String shopId) async {
    try {
      AppLogger.debug('Internal search with keyword: "$keyword" for shop: $shopId');

      // ตัดคำภาษาไทยเป็น Array ผ่าน API
      final keywords = await ThaiWordTokenizer.tokenize(keyword);
      AppLogger.debug('Keywords for search: $keywords');

      // ✅ Sanitize keywords for SQL (escape single quotes)
      final sanitizedKeywords = keywords.map((k) => k.replaceAll("'", "''")).toList();
      final sanitizedKeyword = keyword.replaceAll("'", "''").trim();
      final searchKeywordLower = sanitizedKeyword.toLowerCase();

      // ✅ สร้าง combined text: itemcode + barcode + name0 + unitcode + unitname
      // มี space คั่นกลาง เพื่อไม่ให้ปนกัน
      const combinedField = '''concat(
        ifNull(itemcode, ''), ' ',
        ifNull(barcode, ''), ' ',
        ifNull(name0, ''), ' ',
        ifNull(unitcode, ''), ' ',
        ifNull(unitname, '')
      )''';

      // สร้าง WHERE clause: ค้นหาทุกคำใน combined text
      final whereConditions = sanitizedKeywords.map((kw) => "lower($combinedField) LIKE lower('%$kw%')").join(' AND ');

      final whereClause = whereConditions.isNotEmpty ? whereConditions : "lower($combinedField) LIKE lower('%$searchKeywordLower%')";

      // คำแรกของ keywords มีความสำคัญสุด (น่าจะเป็นคำหลัก)
      final firstKeyword = sanitizedKeywords.isNotEmpty ? sanitizedKeywords.first : '';

      // สร้าง ORDER BY clauses (ใช้ combined field)
      final orderByClauses = <String>[];

      // === Priority 1: Exact Matches (สำคัญสุด!) ===

      // 1. ตรงทั้งหมดกับ barcode
      orderByClauses.add("lower(barcode) = lower('$searchKeywordLower') DESC");

      // 2. ตรงทั้งหมดกับ itemcode
      orderByClauses.add("lower(itemcode) = lower('$searchKeywordLower') DESC");

      // 3. ตรงทั้งหมดกับ name0
      orderByClauses.add("lower(name0) = lower('$searchKeywordLower') DESC");

      // === Priority 2: All Keywords Match (มีครบทุกคำใน combined text) ===

      // 4. ถ้ามีหลายคำ → ต้องมีครบทุกคำใน combined text (สำคัญมาก!)
      if (sanitizedKeywords.length > 1) {
        final conditions = sanitizedKeywords.map((kw) => "lower($combinedField) LIKE lower('%$kw%')").join(' AND ');
        orderByClauses.add("if($conditions, 1, 0) DESC");
      }

      // === Priority 3: Prefix Matches (ขึ้นต้นด้วย) ===

      // 5. ขึ้นต้นด้วย barcode
      orderByClauses.add("lower(barcode) LIKE lower('$searchKeywordLower%') DESC");

      // 6. ขึ้นต้นด้วย itemcode
      orderByClauses.add("lower(itemcode) LIKE lower('$searchKeywordLower%') DESC");

      // 7. ขึ้นต้นด้วย name0
      orderByClauses.add("lower(name0) LIKE lower('$searchKeywordLower%') DESC");

      // 8. ขึ้นต้นด้วยคำแรก (ใน combined text)
      if (firstKeyword.isNotEmpty) {
        orderByClauses.add("lower($combinedField) LIKE lower('$firstKeyword%') DESC");
      }
      // === Priority 4: Partial Matches (มีอยู่บางส่วน ใน combined text) ===

      // 9. นับจำนวนคำที่พบ (ใน combined text)
      if (sanitizedKeywords.isNotEmpty) {
        final conditions = sanitizedKeywords.map((kw) => "if(lower($combinedField) LIKE lower('%$kw%'), 1, 0)").join(' + ');
        orderByClauses.add("($conditions) DESC");
      }

      // 10. มีคำค้นหาทั้งหมด (ใน combined text)
      orderByClauses.add("lower($combinedField) LIKE lower('%$searchKeywordLower%') DESC");

      // 11. ความยาวสั้น (name0)
      orderByClauses.add("length(name0) ASC");

      // 12. เรียงตาม itemcode (กลุ่มสินค้าเดียวกัน)
      orderByClauses.add("itemcode");

      // 13. เรียงตาม ratio (unitstand/unitdivide) จากน้อยไปมาก
      orderByClauses.add("unitstand ASC");
      orderByClauses.add("unitdivide DESC");

      // 14. เรียง A-Z
      orderByClauses.add("name0");

      final orderByClause = orderByClauses.join(', ');

      final query =
          '''
        SELECT 
          $_PRODUCT_FIELDS
        FROM $_PRODUCT_TABLE 
        WHERE shopid = '$shopId' AND ($whereClause)
        ORDER BY $orderByClause
        LIMIT 50
      '''
              .trim()
              .replaceAll(RegExp(r'\s+'), ' ');

      AppLogger.debug('Full Text Search Query: $query');

      final jsonData = await _executeQuery(query, context: '_searchByNameInternal');
      return _parseProductListResponse(jsonData, '_searchByNameInternal');
    } catch (e) {
      AppLogger.error('Exception in _searchByNameInternal: $e');
      return ProductSearchResult(status: 'error', message: e.toString(), count: 0, products: []);
    }
  }

  /// คนหาสนคาดวย itemcode
  Future<ProductSearchResult> searchByItemCode(String itemCode, String shopId) async {
    try {
      AppLogger.info('Searching product by itemcode: $itemCode for shop: $shopId');

      final query =
          "SELECT $_PRODUCT_FIELDS FROM $_PRODUCT_TABLE "
          "WHERE shopid = '$shopId' AND itemcode = '$itemCode' LIMIT 1";

      final jsonData = await _executeQuery(query, context: 'searchByItemCode');
      return _parseProductListResponse(jsonData, 'searchByItemCode');
    } catch (e) {
      AppLogger.error('Exception in searchByItemCode: $e');
      return ProductSearchResult(status: 'error', message: e.toString(), count: 0, products: []);
    }
  }

  /// ค้นหาสินค้าแบบ Universal (ค้นหาทั้ง barcode, itemcode, name, unitcode, unitname)
  /// ใช้ Full Text Search ค้นหาทุก fields ในครั้งเดียว
  ///
  /// ✅ วิธีการทำงาน:
  /// - ค้นหาทุกฟิลด์พร้อมกัน: itemcode, barcode, name0, unitcode, unitname
  /// - ใช้ Full Text Search (3 Steps: Original + Thai→EN + EN→TH)
  /// - ไม่แยก exact match ออกมาต่างหาก จะได้ผลลัพธ์เหมือนกัน
  Future<ProductSearchResult> universalSearch(String keyword, String shopId) async {
    try {
      AppLogger.info('🔍 Universal search with keyword: "$keyword" for shop: $shopId');

      // ✅ ใช้ Full Text Search ค้นหาทุก field ในครั้งเดียว
      // ค้นหา: itemcode, barcode, name0, unitcode, unitname
      return await searchByName(keyword, shopId);
    } catch (e) {
      AppLogger.error('Exception in universalSearch: $e');
      return ProductSearchResult(status: 'error', message: e.toString(), count: 0, products: []);
    }
  }

  /// ดึงสินค้ายอดนิยม พร้อมขยายให้ครบทุก barcode
  Future<ProductSearchResult> getPopularProducts(String shopId, {int limit = 300}) async {
    try {
      AppLogger.info('Getting popular products for shop: $shopId (limit: $limit)');

      // ดึงสินค้ายอดนิยม (เอาแค่ itemcode ไม่ซ้ำ)
      // ใช้ subquery เพื่อหลีกเลี่ยง aggregate function ใน WHERE
      final query =
          "SELECT "
          "shopid, "
          "itemcode, "
          "any(barcode) as barcode, "
          "any(name0) as name0, "
          "any(unitcode) as unitcode, "
          "any(unitname) as unitname, "
          "MAX(price1) as price1, "
          "any(unitstand) as unitstand, "
          "any(unitdivide) as unitdivide "
          "FROM ${global.clickHouseDatabaseName}.productbarcode "
          "WHERE shopid = '$shopId' "
          "GROUP BY itemcode, shopid "
          "ORDER BY price1 DESC "
          "LIMIT $limit";

      final url = Uri.parse('$baseUrl/clickhouse/query');
      final response = await http.post(url, headers: {'Content-Type': 'application/json'}, body: jsonEncode({'query': query})).timeout(const Duration(seconds: 10));

      if (response.statusCode == 200) {
        final jsonData = jsonDecode(response.body);

        if (jsonData['status'] == 'success' && jsonData['data'] != null) {
          final List<dynamic> products = jsonData['data'];

          if (products.isEmpty) {
            AppLogger.info('No popular products found');
            return ProductSearchResult(status: 'success', count: 0, products: []);
          }

          final productList = products.map((p) => ProductSearchModel.fromJson(p)).toList();
          AppLogger.info('Found ${productList.length} popular itemcode(s)');

          // ขยายให้ครบทุก barcode ของแต่ละ itemcode
          AppLogger.info('📍 Expanding to include all barcodes per itemcode...');

          final uniqueItemCodes = productList.map((p) => p.itemCode).where((code) => code.isNotEmpty).toSet();

          AppLogger.info('📦 Found ${uniqueItemCodes.length} unique itemcodes, fetching all barcodes...');

          // ✅ ดึงทุก barcode ครั้งเดียว (Batch Query)
          final allUnitsMap = await getAllUnitsForMultipleProducts(uniqueItemCodes.toList(), shopId);

          // สร้าง map จากผลลัพธ์
          final expandedProductsMap = <String, ProductSearchModel>{};

          for (final entry in allUnitsMap.entries) {
            final itemCode = entry.key;
            final units = entry.value;

            AppLogger.debug('  ✓ ItemCode $itemCode: ${units.length} barcode(s)');

            // เพิ่มทุก barcode เข้า map
            for (final unit in units) {
              final key = '${unit.itemCode}_${unit.barcode}';
              expandedProductsMap[key] = unit;
            }
          }

          AppLogger.info('✅ Expanded to ${expandedProductsMap.length} total products (all barcodes)');

          // แปลง map กลับเป็น list และเรียงใหม่
          final expandedProducts = expandedProductsMap.values.toList();

          // เรียงตาม: 1) itemcode (A-Z) 2) อัตราส่วน (มาก→น้อย) 3) barcode (A-Z)
          expandedProducts.sort((a, b) {
            // 1. เรียงตาม itemcode
            final itemCodeCompare = a.itemCode.compareTo(b.itemCode);
            if (itemCodeCompare != 0) return itemCodeCompare;

            // 2. เรียงตามอัตราส่วน (มาก→น้อย)
            final ratioA = a.unitStand / a.unitDive;
            final ratioB = b.unitStand / b.unitDive;
            final ratioCompare = ratioB.compareTo(ratioA);
            if (ratioCompare != 0) return ratioCompare;

            // 3. เรียงตาม barcode
            return a.barcode.compareTo(b.barcode);
          });

          AppLogger.info('🎯 Final Result: ${expandedProducts.length} products (expanded from ${productList.length})');

          return ProductSearchResult(status: 'success', count: expandedProducts.length, products: expandedProducts);
        }
      }

      AppLogger.error('ClickHouse API error: ${response.statusCode} - ${response.body}');
      return ProductSearchResult(status: 'error', message: 'API returned status ${response.statusCode}', count: 0, products: []);
    } catch (e) {
      AppLogger.error('Exception in getPopularProducts: $e');
      return ProductSearchResult(status: 'error', message: e.toString(), count: 0, products: []);
    }
  }

  /// คำนวณคะแนนความเกี่ยวข้องของสินค้ากับคำค้นหา
  /// คะแนนสูง = เกี่ยวข้องมาก
  int _calculateProductScore({
    required ProductSearchModel product,
    required String originalKeyword,
    required List<String> keywords,
    List<String>? translatedKeywords, // คำที่แปลแล้ว (รวมทั้งหมด)
  }) {
    int score = 0;
    final keywordLower = originalKeyword.toLowerCase().trim();
    final name = product.name0.toLowerCase();
    final barcode = product.barcode.toLowerCase();
    final itemcode = product.itemCode.toLowerCase();
    final unitcode = product.unitCode.toLowerCase();

    // ใช้ทั้งคำเดิมและคำที่แปล (สำหรับการเช็คความครบถ้วน)
    final allKeywords = translatedKeywords ?? keywords;

    // === Priority 1: Exact Matches (1000+ points) ===

    if (barcode == keywordLower) {
      score += 1000; // Barcode ตรงทุกตัว (สำคัญสุด!)
    } else if (itemcode == keywordLower) {
      score += 900; // ItemCode ตรงทุกตัว
    } else if (name == keywordLower) {
      score += 800; // ชื่อตรงทุกตัว
    }

    // === Priority 2: All Keywords Match (500+ points) ===

    if (allKeywords.length > 1) {
      // เช็คว่ามีคำทุกคำในชื่อหรือไม่ (รวมคำที่แปลด้วย)
      final allKeywordsInName = allKeywords.every((kw) => name.contains(kw.toLowerCase()));
      if (allKeywordsInName) {
        score += 500; // มีครบทุกคำในชื่อ (สำคัญมาก!)
      }

      // เช็คว่ามีคำทุกคำในหลายฟิลด์หรือไม่
      final allKeywordsAnywhere = allKeywords.every((kw) {
        final kwLower = kw.toLowerCase();
        return name.contains(kwLower) || barcode.contains(kwLower) || itemcode.contains(kwLower) || unitcode.contains(kwLower);
      });
      if (allKeywordsAnywhere) {
        score += 300; // มีครบทุกคำในฟิลด์ใดฟิลด์หนึ่ง
      }
    }

    // === Priority 3: Prefix Matches (200+ points) ===

    if (barcode.startsWith(keywordLower)) {
      score += 250; // Barcode ขึ้นต้นด้วยคำค้น
    }
    if (itemcode.startsWith(keywordLower)) {
      score += 230; // ItemCode ขึ้นต้นด้วยคำค้น
    }
    if (name.startsWith(keywordLower)) {
      score += 220; // ชื่อขึ้นต้นด้วยคำค้น
    }

    // ขึ้นต้นด้วยคำแรก (จากคำเดิม)
    if (keywords.isNotEmpty) {
      final firstKeyword = keywords.first.toLowerCase();
      if (name.startsWith(firstKeyword)) {
        score += 200;
      }
    }

    // === Priority 4: Partial Keyword Matches (10-100 points each) ===

    for (final keyword in allKeywords) {
      final kw = keyword.toLowerCase();

      // นับจำนวนครั้งที่เจอในแต่ละฟิลด์
      if (name.contains(kw)) {
        score += 50; // เจอในชื่อ
      }
      if (barcode.contains(kw)) {
        score += 40; // เจอใน barcode
      }
      if (itemcode.contains(kw)) {
        score += 30; // เจอใน itemcode
      }
      if (unitcode.contains(kw)) {
        score += 20; // เจอใน unitcode
      }
    }

    // === Bonus: Shorter name = more relevant ===

    // ชื่อสั้นกว่า = น่าจะเกี่ยวข้องมากกว่า
    if (name.length < 20) {
      score += 10;
    } else if (name.length < 40) {
      score += 5;
    }

    return score;
  }

  /// ดึงหน่วยนับทั้งหมดสำหรับหลาย itemcode ในครั้งเดียว (Batch Query)
  /// เพื่อลด N+1 query problem
  ///
  /// คืนค่า: Map<itemCode, List<ProductSearchModel>>
  Future<Map<String, List<ProductSearchModel>>> getAllUnitsForMultipleProducts(
    List<String> itemCodes,
    String shopId, {
    Map<String, List<ProductSearchModel>>? unitsMap, // cache
  }) async {
    try {
      if (itemCodes.isEmpty) {
        return {};
      }

      // กรองเฉพาะ itemcode ที่ยังไม่มีใน cache
      final itemCodesToFetch = <String>[];
      final result = <String, List<ProductSearchModel>>{};

      for (final itemCode in itemCodes) {
        if (unitsMap != null && unitsMap.containsKey(itemCode)) {
          // ใช้ cache
          result[itemCode] = unitsMap[itemCode]!;
        } else {
          // ต้อง fetch จาก API
          itemCodesToFetch.add(itemCode);
        }
      }

      if (itemCodesToFetch.isEmpty) {
        AppLogger.debug('✅ [getAllUnitsForMultipleProducts] All ${itemCodes.length} itemcodes found in cache');
        return result;
      }

      AppLogger.info('🌐 [getAllUnitsForMultipleProducts] Fetching ${itemCodesToFetch.length} itemcodes from API (${itemCodes.length - itemCodesToFetch.length} from cache)');

      // สร้าง IN clause สำหรับ itemcodes
      final itemCodesCondition = itemCodesToFetch.map((code) => "'$code'").join(',');

      // Query ครั้งเดียวสำหรับทุก itemcode
      // เรียงตาม: itemcode, อัตราส่วน (มาก→น้อย), barcode (A→Z)
      final query =
          "SELECT $_PRODUCT_FIELDS "
          "FROM $_PRODUCT_TABLE "
          "WHERE shopid = '$shopId' AND itemcode IN ($itemCodesCondition) "
          "ORDER BY itemcode ASC, unitstand DESC, unitdivide ASC, barcode ASC";

      final jsonData = await _executeQuery(query, context: 'getAllUnitsForMultipleProducts');

      if (jsonData['status'] == 'success' && jsonData['data'] != null) {
        final List<dynamic> products = jsonData['data'];

        // จัดกลุ่มตาม itemcode
        for (final productJson in products) {
          final product = ProductSearchModel.fromJson(productJson);
          result.putIfAbsent(product.itemCode, () => []).add(product);
        }

        AppLogger.info('✅ [getAllUnitsForMultipleProducts] Fetched ${products.length} units for ${result.length} itemcodes');

        // แสดงสรุปแต่ละ itemcode
        for (final entry in result.entries) {
          if (!itemCodesToFetch.contains(entry.key)) {
            continue; // skip cached items
          }
          AppLogger.debug('   ${entry.key}: ${entry.value.length} unit(s)');
        }

        return result;
      }

      return result;
    } catch (e) {
      AppLogger.error('Exception in getAllUnitsForMultipleProducts: $e');
      return {};
    }
  }

  /// ดึงหน่วยนับทั้งหมดของสินค้าตาม itemcode
  /// เรียงตามอัตราส่วน (unitstand) จากมากไปน้อย
  ///
  /// ถ้ามี unitsMap (cache) จะดึงจาก memory ก่อน
  /// ถ้าไม่มีค่อยเรียก API
  Future<List<ProductSearchModel>> getAllUnitsForProduct(
    String itemCode,
    String shopId, {
    Map<String, List<ProductSearchModel>>? unitsMap, // cache
  }) async {
    try {
      // ✅ ถ้ามี cache ให้ดึงจาก memory ก่อน
      if (unitsMap != null && unitsMap.containsKey(itemCode)) {
        final cachedUnits = unitsMap[itemCode]!;
        AppLogger.debug('� [getAllUnitsForProduct] Using cached data for $itemCode (${cachedUnits.length} units)');
        return cachedUnits;
      }

      // ถ้าไม่มี cache ค่อยเรียก API
      AppLogger.debug('🌐 [getAllUnitsForProduct] Fetching from API for $itemCode');

      // เรียงตาม: อัตราส่วน (มาก→น้อย), barcode (A→Z)
      final query =
          "SELECT $_PRODUCT_FIELDS "
          "FROM $_PRODUCT_TABLE "
          "WHERE shopid = '$shopId' AND itemcode = '$itemCode' "
          "ORDER BY unitstand DESC, unitdivide ASC, barcode ASC";

      final jsonData = await _executeQuery(query, context: 'getAllUnitsForProduct');

      if (jsonData['status'] == 'success' && jsonData['data'] != null) {
        final List<dynamic> products = jsonData['data'];
        final productList = products.map((p) => ProductSearchModel.fromJson(p)).toList();

        AppLogger.info('✅ [getAllUnitsForProduct] Found ${productList.length} unit(s) for itemcode: $itemCode');

        // แสดงรายละเอียดทุก barcode ที่ได้
        for (var i = 0; i < productList.length; i++) {
          final p = productList[i];
          AppLogger.debug('   ${i + 1}. ${p.unitStand}:${p.unitDive} ${p.unitName} | Barcode: ${p.barcode}');
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
  /// ยอดคงเหลือ = sum((qty * calcflag) * stand/divide)
  /// จาก table ${global.clickHouseDatabaseName}.docdetail where iscalstock=1
  ///
  /// ⚠️ **REAL-TIME DATA - NO CACHE**
  /// ยอดคงเหลือต้องเป็น real-time เสมอ ไม่ cache เพราะเปลี่ยนแปลงตลอดเวลา
  ///
  /// คืนค่า: Map<itemCode, ProductBalanceModel>
  Future<Map<String, ProductBalanceModel>> getProductBalances(
    List<String> itemCodes,
    String? shopId, // ถ้า null = ดึงทั้งหมด, ถ้าระบุ = ดึงเฉพาะคลังนั้น
  ) async {
    try {
      if (itemCodes.isEmpty) {
        return {};
      }

      AppLogger.info('📊 Getting balances for ${itemCodes.length} items${shopId != null ? ' (shop: $shopId)' : ' (all shops)'}');

      // สร้าง WHERE clause สำหรับ itemcodes
      final itemCodesCondition = itemCodes.map((code) => "'$code'").join(',');

      // สร้าง query สำหรับยอดคงเหลือ แบบ flat (ไม่ใช้ nested groupArray)
      final query =
          '''
        SELECT
          ifNull(itemcode, '') as itemcode,
          shopid,
          sum((qty * calcflag) * unitstand / unitdivide) as balance
        FROM $_DOCDETAIL_TABLE
        WHERE iscalcstock = 1
          AND itemcode IN ($itemCodesCondition)
          ${shopId != null ? "AND shopid = '$shopId'" : ''}
        GROUP BY itemcode, shopid
      '''
              .trim()
              .replaceAll(RegExp(r'\s+'), ' ');

      final url = Uri.parse('$baseUrl/clickhouse/query');
      final requestBody = jsonEncode({'query': query});

      AppLogger.debug('🌐 [getProductBalances] API URL: $url');
      AppLogger.debug('📤 [getProductBalances] Request: $requestBody');

      final response = await http.post(url, headers: {'Content-Type': 'application/json'}, body: requestBody).timeout(const Duration(seconds: 15));

      AppLogger.debug('📥 [getProductBalances] Status: ${response.statusCode}');
      AppLogger.debug('📥 [getProductBalances] Response: ${response.body}');

      if (response.statusCode == 200) {
        final jsonData = jsonDecode(response.body);

        if (jsonData['status'] == 'success' && jsonData['data'] != null) {
          final List<dynamic> balances = jsonData['data'];
          final result = <String, ProductBalanceModel>{};

          // จัดกลุ่มข้อมูลตาม itemcode
          final groupedData = <String, Map<String, double>>{};

          for (final item in balances) {
            final itemCode = item['itemcode'] as String? ?? '';
            final sid = item['shopid'] as String? ?? '';
            final balance = (item['balance'] as num?)?.toDouble() ?? 0.0;

            if (itemCode.isEmpty) continue;

            if (!groupedData.containsKey(itemCode)) {
              groupedData[itemCode] = {};
            }

            if (sid.isNotEmpty) {
              groupedData[itemCode]![sid] = balance;
            }
          }

          // สร้าง ProductBalanceModel สำหรับแต่ละ itemcode
          for (final entry in groupedData.entries) {
            final itemCode = entry.key;
            final balanceByShop = entry.value;

            // คำนวณ total balance
            final totalBalance = balanceByShop.values.fold(0.0, (sum, val) => sum + val);

            result[itemCode] = ProductBalanceModel(itemCode: itemCode, totalBalance: totalBalance, balanceByShop: balanceByShop);

            AppLogger.debug('   ✓ $itemCode: Total=${totalBalance.toStringAsFixed(2)}, Shops=${balanceByShop.length}');
          }

          AppLogger.info('✅ [getProductBalances] Got balances for ${result.length} items');

          return result;
        }
      }

      AppLogger.error('ClickHouse API error: ${response.statusCode} - ${response.body}');
      return {};
    } catch (e) {
      AppLogger.error('Exception in getProductBalances: $e');
      return {};
    }
  }

  /// ดึงรายการซื้อล่าสุดของสินค้า (จาก docdetail)
  /// ดึงจากเอกสารประเภท PO (ใบสั่งซื้อ) หรือ GR (รับของ)
  Future<List<ProductTransactionModel>> getRecentPurchases({required String itemCode, required String shopId, int limit = 5}) async {
    try {
      AppLogger.info('🛒 Getting recent purchases for item: $itemCode (limit: $limit)');

      // Query สำหรับดึงรายการซื้อ
      // ดึงจากเอกสารที่เป็นการซื้อ (calcflag = 1)
      final query =
          '''
        SELECT
          docno,
          docdatetime,
          ifNull(itemname, '') as itemname,
          qty,
          unitcode as unitname,
          price,
          (qty * price) as amount
        FROM $_DOCDETAIL_TABLE
        WHERE itemcode = '$itemCode'
          AND shopid = '$shopId'
          AND calcflag = 1
          AND iscalcstock = 1
        ORDER BY docdatetime DESC, docno DESC
        LIMIT $limit
      '''
              .trim()
              .replaceAll(RegExp(r'\s+'), ' ');

      final jsonData = await _executeQuery(query, context: 'getRecentPurchases');

      if (jsonData['status'] == 'success' && jsonData['data'] != null) {
        final List<dynamic> transactions = jsonData['data'];
        final result = transactions.map((t) => ProductTransactionModel.fromJson(t)).toList();

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
  /// ดึงจากเอกสารประเภท SO (ใบสั่งขาย) หรือ IV (ใบกำกับภาษี)
  Future<List<ProductTransactionModel>> getRecentSales({required String itemCode, required String shopId, int limit = 5}) async {
    try {
      AppLogger.info('💰 Getting recent sales for item: $itemCode (limit: $limit)');

      // Query สำหรับดึงรายการขาย
      // ดึงจากเอกสารที่เป็นการขาย (calcflag = -1)
      final query =
          '''
        SELECT
          docno,
          docdatetime,
          ifNull(itemname, '') as itemname,
          qty,
          unitcode as unitname,
          price,
          (qty * price) as amount
        FROM $_DOCDETAIL_TABLE
        WHERE itemcode = '$itemCode'
          AND shopid = '$shopId'
          AND calcflag = -1
          AND iscalcstock = 1
        ORDER BY docdatetime DESC, docno DESC
        LIMIT $limit
      '''
              .trim()
              .replaceAll(RegExp(r'\s+'), ' ');

      final jsonData = await _executeQuery(query, context: 'getRecentSales');

      if (jsonData['status'] == 'success' && jsonData['data'] != null) {
        final List<dynamic> transactions = jsonData['data'];
        final result = transactions.map((t) => ProductTransactionModel.fromJson(t)).toList();

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
  Future<Map<String, ProductWarehouseBalanceModel>> getDetailedBalances({required String itemCode, required String shopId}) async {
    try {
      AppLogger.info('📦 Getting detailed balances for item: $itemCode (shop: $shopId)');

      // Query สำหรับดึงยอดคงเหลือแยกตาม warehouse (whcode) และ location (locationcode)
      // ใช้ shopid, whcode, locationcode เพื่อแยกยอดคงเหลือได้ละเอียด
      // shopid ใช้สำหรับ filter เท่านั้น ไม่แสดงในผลลัพธ์
      final query =
          '''
        SELECT
          ifNull(whcode, '') as warehouse_id,
          ifNull(whcode, '') as warehouse_name,
          ifNull(locationcode, '') as location_id,
          ifNull(locationcode, '') as location_name,
          sum((qty * calcflag) * unitstand / unitdivide) as balance
        FROM $_DOCDETAIL_TABLE
        WHERE itemcode = '$itemCode'
          AND shopid = '$shopId'
          AND iscalcstock = 1
        GROUP BY whcode, locationcode
        ORDER BY whcode, locationcode
      '''
              .trim()
              .replaceAll(RegExp(r'\s+'), ' ');

      final jsonData = await _executeQuery(query, context: 'getDetailedBalances');

      if (jsonData['status'] == 'success' && jsonData['data'] != null) {
        final List<dynamic> balances = jsonData['data'];

        AppLogger.info('📦 Raw balance data from query:');
        for (final item in balances) {
          AppLogger.info(
            '  - warehouse_id: ${item['warehouse_id']}, '
            'warehouse_name: ${item['warehouse_name']}, '
            'location_id: ${item['location_id']}, '
            'location_name: ${item['location_name']}, '
            'balance: ${item['balance']}',
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
            final balance = (item['balance'] as num?)?.toDouble() ?? 0.0;

            totalBalance += balance;

            if (locationId.isNotEmpty || balance != 0) {
              final locKey = locationId.isEmpty ? 'default' : locationId;
              locationBalances[locKey] = ProductLocationBalanceModel(locationId: locationId, locationName: locationName, balance: balance);
            }
          }

          final warehouseNameRaw = items.isNotEmpty ? (items.first['warehouse_name'] as String? ?? '') : '';
          // ถ้า warehouseName เป็นค่าว่าง ให้ใช้ "คลังหลัก" แทน
          final warehouseName = warehouseNameRaw.isEmpty ? global.language("default_warehouse") : warehouseNameRaw;

          result[warehouseId] = ProductWarehouseBalanceModel(warehouseId: warehouseId, warehouseName: warehouseName, totalBalance: totalBalance, locationBalances: locationBalances);

          AppLogger.info(
            '🏭 Warehouse: $warehouseId (name: $warehouseName), '
            'total: $totalBalance, locations: ${locationBalances.length}',
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
