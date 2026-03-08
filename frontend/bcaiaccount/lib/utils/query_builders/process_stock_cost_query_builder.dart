// Query Builder สำหรับ processstockcost table
// สร้าง SQL query ตามเงื่อนไขที่เลือก (date range, branch, customer, product)

@Deprecated('Use StockCostApiService instead — backend parameterized queries are safer')
class ProcessStockCostQueryBuilder {
  /// สร้าง SQL query สำหรับ processstockcost table
  ///
  /// Parameters:
  /// - [fromDate] วันที่เริ่มต้น
  /// - [toDate] วันที่สิ้นสุด
  /// - [branchCodes] รหัสสาขาที่เลือก (ถ้าว่างเปล่า = ทุกสาขา)
  /// - [customerCodes] รหัสลูกค้าที่เลือก (ถ้าว่างเปล่า = ทุกลูกค้า)
  /// - [productCodes] รหัสสินค้าที่เลือก (ถ้าว่างเปล่า = ทุกสินค้า)
  static String buildQuery({
    required DateTime fromDate,
    required DateTime toDate,
    List<String>? branchCodes,
    List<String>? customerCodes,
    List<String>? productCodes,
  }) {
    // Format วันที่เป็น ISO 8601
    final fromDateStr = _formatDate(fromDate);
    final toDateStr = _formatDate(
      toDate.add(const Duration(days: 1)),
    ); // +1 วัน เพื่อรวมวันสุดท้าย

    // Base query
    final buffer = StringBuffer();
    buffer.writeln('SELECT');
    buffer.writeln('  id,');
    buffer.writeln('  docdatetime,');
    buffer.writeln('  docno,');
    buffer.writeln('  docref,');
    buffer.writeln('  linenumber,');
    buffer.writeln('  transflag,');
    buffer.writeln('  itemcode,');
    buffer.writeln('  itemname,');
    buffer.writeln('  barcode,');
    buffer.writeln('  unitcode,');
    buffer.writeln('  whcode,');
    buffer.writeln('  locationcode,');
    buffer.writeln('  debtorcode,');
    buffer.writeln('  debtorname,');
    buffer.writeln('  totalqty,');
    buffer.writeln('  unitstand,');
    buffer.writeln('  unitdivide,');
    buffer.writeln('  price,');
    buffer.writeln('  averagecost,');
    buffer.writeln('  calcamount,');
    buffer.writeln('  balanceqty,');
    buffer.writeln('  balanceamount,');
    buffer.writeln('  unitcost,');
    buffer.writeln('  guid');
    buffer.writeln('FROM public.processstockcost');

    // WHERE clause
    final conditions = <String>[];

    // เงื่อนไขวันที่ (บังคับ)
    conditions.add("docdatetime >= '$fromDateStr'");
    conditions.add("docdatetime < '$toDateStr'");

    // เงื่อนไขสาขา
    if (branchCodes != null && branchCodes.isNotEmpty) {
      final branchList = branchCodes.map((code) => "'$code'").join(', ');
      conditions.add("whcode IN ($branchList)");
    }

    // เงื่อนไขลูกค้า (ถ้ามี - อาจต้องเชื่อมกับ table อื่น)
    // หมายเหตุ: processstockcost ไม่มี customer code โดยตรง
    // อาจต้อง JOIN กับ transaction table ถ้าต้องการกรองตามลูกค้า
    if (customerCodes != null && customerCodes.isNotEmpty) {
      // ตัวอย่าง: กรองจาก docno pattern หรือ JOIN table อื่น
      // สำหรับตอนนี้ข้ามไปก่อน
      // TODO: เพิ่ม JOIN ถ้าจำเป็น
    }

    // เงื่อนไขสินค้า
    if (productCodes != null && productCodes.isNotEmpty) {
      final productList = productCodes.map((code) => "'$code'").join(', ');
      conditions.add("itemcode IN ($productList)");
    }

    // เพิ่ม WHERE clause
    if (conditions.isNotEmpty) {
      buffer.writeln('WHERE ${conditions.join(' AND ')}');
    }

    // ORDER BY
    buffer.writeln('ORDER BY docdatetime DESC, docno, linenumber');

    return buffer.toString();
  }

  /// สร้าง SQL query แบบกำหนดเอง
  ///
  /// สำหรับกรณีที่ต้องการ query ที่ซับซ้อนกว่า
  static String buildCustomQuery({
    required String selectColumns,
    String? whereClause,
    String? orderByClause,
    int? limit,
  }) {
    final buffer = StringBuffer();
    buffer.writeln('SELECT $selectColumns');
    buffer.writeln('FROM public.processstockcost');

    if (whereClause != null && whereClause.isNotEmpty) {
      buffer.writeln('WHERE $whereClause');
    }

    if (orderByClause != null && orderByClause.isNotEmpty) {
      buffer.writeln('ORDER BY $orderByClause');
    }

    if (limit != null && limit > 0) {
      buffer.writeln('LIMIT $limit');
    }

    return buffer.toString();
  }

  /// สร้าง query สำหรับสรุปยอดรวม (aggregate)
  static String buildSummaryQuery({
    required DateTime fromDate,
    required DateTime toDate,
    List<String>? branchCodes,
    List<String>? productCodes,
  }) {
    final fromDateStr = _formatDate(fromDate);
    final toDateStr = _formatDate(toDate.add(const Duration(days: 1)));

    final buffer = StringBuffer();
    buffer.writeln('SELECT');
    buffer.writeln('  itemcode,');
    buffer.writeln('  SUM(totalqty) as total_quantity,');
    buffer.writeln('  SUM(calcamount) as total_amount,');
    buffer.writeln('  AVG(averagecost) as avg_cost,');
    buffer.writeln('  COUNT(*) as transaction_count');
    buffer.writeln('FROM public.processstockcost');

    final conditions = <String>[];
    conditions.add("docdatetime >= '$fromDateStr'");
    conditions.add("docdatetime < '$toDateStr'");

    if (branchCodes != null && branchCodes.isNotEmpty) {
      final branchList = branchCodes.map((code) => "'$code'").join(', ');
      conditions.add("whcode IN ($branchList)");
    }

    if (productCodes != null && productCodes.isNotEmpty) {
      final productList = productCodes.map((code) => "'$code'").join(', ');
      conditions.add("itemcode IN ($productList)");
    }

    if (conditions.isNotEmpty) {
      buffer.writeln('WHERE ${conditions.join(' AND ')}');
    }

    buffer.writeln('GROUP BY itemcode');
    buffer.writeln('ORDER BY total_amount DESC');
    buffer.writeln('LIMIT 10000');

    return buffer.toString();
  }

  /// Format DateTime เป็น ISO 8601 string สำหรับ PostgreSQL
  static String _formatDate(DateTime date) {
    return date.toUtc().toIso8601String();
  }

  /// Escape string สำหรับป้องกัน SQL injection
  static String _escapeString(String value) {
    return value.replaceAll("'", "''");
  }

  /// สร้าง query สำหรับตรวจสอบข้อมูลที่มีอยู่
  static String buildDataCheckQuery({
    required DateTime fromDate,
    required DateTime toDate,
  }) {
    final fromDateStr = _formatDate(fromDate);
    final toDateStr = _formatDate(toDate.add(const Duration(days: 1)));

    return '''
SELECT
  COUNT(*) as total_rows,
  COUNT(DISTINCT docno) as total_documents,
  COUNT(DISTINCT itemcode) as total_products,
  MIN(docdatetime) as earliest_date,
  MAX(docdatetime) as latest_date
FROM public.processstockcost
WHERE docdatetime >= '$fromDateStr'
  AND docdatetime < '$toDateStr'
''';
  }
}
