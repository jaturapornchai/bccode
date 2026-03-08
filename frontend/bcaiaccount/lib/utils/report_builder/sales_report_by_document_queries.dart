/// SQL builders for the Sales Report by Document workflow.
@Deprecated('Use SalesReportApiService instead — backend parameterized queries are safer')
class SalesReportByDocumentQueries {
  const SalesReportByDocumentQueries._();

  /// Aggregated header-level query from processstockcost grouped by docno.
  static String buildHeaderQuery({
    required DateTime fromDate,
    required DateTime toDate,
    List<String>? branchCodes,
    List<String>? productCodes,
    bool sortAscending = true,
  }) {
    final buffer = StringBuffer();
    buffer.writeln('SELECT');
    buffer.writeln(
      "  CAST(MAX(p.docdatetime) + INTERVAL '7 hour' AS DATE) as docdate,",
    );
    buffer.writeln(
      "  TO_CHAR(MAX(p.docdatetime) + INTERVAL '7 hour', 'HH24:MI:SS') as doctime,",
    );
    buffer.writeln("  MAX(p.docdatetime) + INTERVAL '7 hour' as docdatetime,");
    buffer.writeln('  p.docno,');
    buffer.writeln("  COALESCE(doc.custcode, '') as debtorcode,");
    buffer.writeln("  COALESCE(d.name0, '') as debtorname,");
    buffer.writeln('  COALESCE(MAX(doc.totalamount), 0) as totalqty,');
    buffer.writeln('  (SUM(p.totalqty * p.price) * -1) as totalamount,');
    buffer.writeln('  AVG(p.price) as price,');
    buffer.writeln('  AVG(p.averagecost) as averagecost,');
    buffer.writeln('  (SUM(p.calcamount) * -1) as calcamount,');
    buffer.writeln(
      '  ((SUM(p.totalqty * p.price) * -1) - (SUM(p.calcamount) * -1)) as grossprofit',
    );
    buffer.writeln('FROM public.processstockcost p');
    buffer.writeln('LEFT JOIN public.doc doc ON p.docno = doc.docno');
    buffer.writeln('LEFT JOIN public.debtor d ON doc.custcode = d.code');

    final conditions = _buildCommonConditions(
      fromDate: fromDate,
      toDate: toDate,
      branchCodes: branchCodes,
      productCodes: productCodes,
      includeProductFilter: false,
    );

    if (productCodes != null && productCodes.isNotEmpty) {
      final docFilter = _buildDocFilterCondition(
        fromDate: fromDate,
        toDate: toDate,
        branchCodes: branchCodes,
        productCodes: productCodes,
      );
      conditions.add(docFilter);
    }

    if (conditions.isNotEmpty) {
      buffer.writeln('WHERE ${conditions.join(' AND ')}');
    }

    buffer.writeln('GROUP BY p.docno, doc.custcode, d.name0');
    final sortDir = sortAscending ? 'ASC' : 'DESC';
    buffer.writeln('ORDER BY MAX(p.docdatetime) $sortDir, p.docno');

    return buffer.toString();
  }

  /// Detail-level query for retrieving individual line items.
  static String buildDetailQuery({
    required DateTime fromDate,
    required DateTime toDate,
    List<String>? branchCodes,
    List<String>? productCodes,
    bool sortAscending = true,
  }) {
    final buffer = StringBuffer();
    buffer.writeln('SELECT');
    buffer.writeln(
      "  CAST(p.docdatetime + INTERVAL '7 hour' AS DATE) as docdate,",
    );
    buffer.writeln(
      "  TO_CHAR(p.docdatetime + INTERVAL '7 hour', 'HH24:MI:SS') as doctime,",
    );
    buffer.writeln("  p.docdatetime + INTERVAL '7 hour' as docdatetime,");
    buffer.writeln('  p.docno,');
    buffer.writeln('  p.linenumber,');
    buffer.writeln('  p.itemcode,');
    buffer.writeln("  COALESCE(item_lookup.itemname, '') as itemname,");
    buffer.writeln('  p.barcode,');
    buffer.writeln("  COALESCE(unit_lookup.unitname, '') as unitname,");
    buffer.writeln('  (p.totalqty * -1) as totalqty,');
    buffer.writeln('  (p.totalqty * p.price * -1) as totalamount,');
    buffer.writeln('  (p.calcamount * -1) as calcamount,');
    buffer.writeln('  p.price,');
    buffer.writeln('  p.averagecost,');
    buffer.writeln(
      '  ((p.totalqty * p.price * -1) - (p.calcamount * -1)) as grossprofit',
    );
    buffer.writeln('FROM public.processstockcost p');
    buffer.writeln('LEFT JOIN LATERAL (');
    buffer.writeln("  SELECT STRING_AGG(DISTINCT pb.name0, ', ') AS itemname");
    buffer.writeln('  FROM public.productbarcode pb');
    buffer.writeln('  WHERE pb.itemcode = p.itemcode');
    buffer.writeln('    AND pb.barcoderefunitstand = 1');
    buffer.writeln('    AND pb.barcoderefunitdivide = 1');
    buffer.writeln(') item_lookup ON TRUE');
    buffer.writeln('LEFT JOIN LATERAL (');
    buffer.writeln(
      "  SELECT STRING_AGG(DISTINCT pb.unitname, ', ') AS unitname",
    );
    buffer.writeln('  FROM public.productbarcode pb');
    buffer.writeln('  WHERE pb.barcode = p.barcode');
    buffer.writeln(') unit_lookup ON TRUE');

    final conditions = _buildCommonConditions(
      fromDate: fromDate,
      toDate: toDate,
      branchCodes: branchCodes,
      productCodes: productCodes,
      includeProductFilter: false,
    );

    if (productCodes != null && productCodes.isNotEmpty) {
      final docFilter = _buildDocFilterCondition(
        fromDate: fromDate,
        toDate: toDate,
        branchCodes: branchCodes,
        productCodes: productCodes,
      );
      conditions.add(docFilter);
    }

    if (conditions.isNotEmpty) {
      buffer.writeln('WHERE ${conditions.join(' AND ')}');
    }

    final sortDir = sortAscending ? 'ASC' : 'DESC';
    buffer.writeln('ORDER BY p.docdatetime $sortDir, p.docno, p.linenumber');
    return buffer.toString();
  }

  static String _formatDate(DateTime date) {
    return date.toUtc().toIso8601String();
  }

  static List<String> _buildCommonConditions({
    required DateTime fromDate,
    required DateTime toDate,
    List<String>? branchCodes,
    List<String>? productCodes,
    bool includeProductFilter = true,
    String tableAlias = 'p',
  }) {
    final fromDateStr = _formatDate(fromDate);
    final toDateStr = _formatDate(toDate.add(const Duration(days: 1)));

    final conditions = <String>[
      "$tableAlias.transflag = 44",
      "$tableAlias.docdatetime >= '$fromDateStr'",
      "$tableAlias.docdatetime < '$toDateStr'",
    ];

    if (branchCodes != null && branchCodes.isNotEmpty) {
      final branchList = branchCodes.map((code) => "'$code'").join(', ');
      conditions.add("$tableAlias.whcode IN ($branchList)");
    }

    if (includeProductFilter &&
        productCodes != null &&
        productCodes.isNotEmpty) {
      final productList = productCodes.map((code) => "'$code'").join(', ');
      conditions.add("$tableAlias.itemcode IN ($productList)");
    }

    return conditions;
  }

  static String _buildDocFilterCondition({
    required DateTime fromDate,
    required DateTime toDate,
    List<String>? branchCodes,
    required List<String> productCodes,
  }) {
    final subConditions = _buildCommonConditions(
      fromDate: fromDate,
      toDate: toDate,
      branchCodes: branchCodes,
      productCodes: productCodes,
      tableAlias: 'sub',
    );

    return '''p.docno IN (
  SELECT DISTINCT sub.docno
  FROM public.processstockcost sub
  WHERE ${subConditions.join(' AND ')}
)''';
  }
}
