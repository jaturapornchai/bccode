import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/util.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class AuditScreen extends StatefulWidget {
  const AuditScreen({super.key});

  @override
  State<AuditScreen> createState() => _AuditScreenState();
}

class _AuditScreenState extends State<AuditScreen> with global.ThemeRefreshMixin {
  Container bottomPanel = Container();


  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(
          global.language('audit_system'),
          style: TextStyle(color: global.theme.onPrimaryColor, fontWeight: FontWeight.bold),
        ),
        backgroundColor: global.theme.appBarColor,
        iconTheme: IconThemeData(color: global.theme.onPrimaryColor),
        leading: IconButton(
          icon: const Icon(Icons.arrow_back),
          onPressed: () {
            Navigator.pop(context);
          },
        ),
      ),
      body: Column(
        children: [
          _buildTopPanel(),
          Expanded(child: bottomPanel),
        ],
      ),
    );
  }

  Widget _buildTopPanel() {
    return Container(
      color: global.theme.surfaceColor,
      child: Column(
        children: [
          Container(
            padding: const EdgeInsets.all(4.0),
            margin: const EdgeInsets.all(4.0),
            decoration: BoxDecoration(
              color: global.theme.cardColor,
              border: Border.all(color: global.theme.dividerBorderColor),
              boxShadow: [BoxShadow(color: global.theme.dividerBorderColor.withValues(alpha: 0.1), spreadRadius: 1, blurRadius: 2, offset: const Offset(0, 1))],
            ),
            child: Wrap(
              spacing: 8,
              runSpacing: 8,
              children: [
                SizedBox(
                  width: 200,
                  height: 100,
                  child: ElevatedButton(
                    onPressed: () async {
                      await _auditBarcodeItemCodeNotFound();
                    },
                    style: ElevatedButton.styleFrom(backgroundColor: global.theme.primaryColor, foregroundColor: global.theme.onPrimaryColor, padding: EdgeInsets.all(4.0)),
                    child: Text(global.language('audit_barcode_without_item_code'), style: TextStyle(fontSize: 13)),
                  ),
                ),
                SizedBox(
                  width: 200,
                  height: 100,
                  child: ElevatedButton(
                    onPressed: () async {
                      await auditBarcodeUnitFormula();
                    },
                    style: ElevatedButton.styleFrom(padding: EdgeInsets.all(4.0), backgroundColor: global.theme.primaryColor, foregroundColor: global.theme.onPrimaryColor),
                    child: Text(global.language('audit_barcode_unit_ratio_error'), style: TextStyle(fontSize: 13)),
                  ),
                ),
                SizedBox(
                  width: 200,
                  height: 100,
                  child: ElevatedButton(
                    onPressed: () async {
                      await auditBarcodeUnitFormulaNoBarcodeRef();
                    },
                    style: ElevatedButton.styleFrom(backgroundColor: global.theme.primaryColor, foregroundColor: global.theme.onPrimaryColor, padding: EdgeInsets.all(4.0)),
                    child: Text(global.language('audit_barcode_non_1_1_ratio_no_ref'), style: TextStyle(fontSize: 13)),
                  ),
                ),
                SizedBox(
                  width: 200,
                  height: 100,
                  child: ElevatedButton(
                    onPressed: () async {
                      await _showAuditMultiUnitDialog();
                    },
                    style: ElevatedButton.styleFrom(backgroundColor: global.theme.primaryColor, foregroundColor: global.theme.onPrimaryColor, padding: EdgeInsets.all(4.0)),
                    child: Text(global.language('audit_products_with_multiple_units'), style: TextStyle(fontSize: 13)),
                  ),
                ),
                SizedBox(
                  width: 200,
                  height: 100,
                  child: ElevatedButton(
                    onPressed: () async {
                      await _auditTableProcessStock();
                    },
                    style: ElevatedButton.styleFrom(backgroundColor: global.theme.primaryColor, foregroundColor: global.theme.onPrimaryColor, padding: EdgeInsets.all(4.0)),
                    child: Text(global.language("audit_check_cost_table"), style: const TextStyle(fontSize: 13)),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Future<void> _auditTableProcessStock() async {
    final infoMessages = <Widget>[];
    final comparisonRows = <_ProcessStockComparisonRow>[];

    bottomPanel = Container(
      color: global.theme.surfaceColor,
      child: const Center(child: CircularProgressIndicator()),
    );
    setState(() {});
    List<dynamic> pgSqlrows = [];
    List<dynamic> clickHouserows = [];
    try {
      // ตรวจสอบตาราง processstock ที่มีข้อมูลผิดปกติ
      infoMessages.add(Text("ตรวจสอบตาราง processstock เทียบวัน PgSql กับ ClickHouse"));
      String queryPgSql =
          "select to_char(docdatetime, 'YYYY-MM-DD') as doc_date,docno,count(*) as xcount,sum(balanceqty) as balanceqty,sum(balanceamount) as balanceamount from processstockcost group by to_char(docdatetime, 'YYYY-MM-DD'),docno order by doc_date,docno";
      infoMessages.add(Text(queryPgSql));
      final pgSqlResponse = await pgSqlSelectGroup([queryPgSql]);
      pgSqlrows = (pgSqlResponse["results"]?[0]?["rows"] as List?) ?? [];

      if (pgSqlrows.isEmpty) {
        AppLogger.debug('PgSql No data found');
        return;
      }

      String clickHouseQuery =
          "SELECT toString(toDate(docdatetime)) AS doc_date,docno,COUNT(*) AS xcount,SUM(balanceqty) AS balanceqty,SUM(balanceamount) AS balanceamount FROM ${global.clickHouseDatabaseName}.processstockcost WHERE shopid = '${global.getShopId()}' GROUP BY toDate(docdatetime),docno ORDER BY doc_date,docno";
      infoMessages.add(Text(clickHouseQuery));
      final clickHouseResponse = await global.clickhouseSelect(clickHouseQuery);
      clickHouserows = (clickHouseResponse["data"] as List?) ?? [];
      if (clickHouserows.isEmpty) {
        AppLogger.debug('ClickHouse No data found');
        return;
      }
    } catch (e) {
      AppLogger.error('Error: $e');
      bottomPanel = Container(
        color: global.theme.surfaceColor,
        child: Center(
          child: Text('Error: $e', style: TextStyle(color: global.theme.negativeHighlightTextColor)),
        ),
      );
      setState(() {});
      return;
    }
    // compare results and show messages
    Set<String> keys = {};
    for (var row in pgSqlrows) {
      keys.add('${row['doc_date']}|${row['docno']}');
    }
    for (var row in clickHouserows) {
      keys.add('${row['doc_date']}|${row['docno']}');
    }
    List<String> sortedKeys = keys.toList()..sort();

    String currentDate = "";
    int dailyPgSqlCount = 0;
    int dailyClickHouseCount = 0;
    double dailyPgSqlBalanceQty = 0.0;
    double dailyClickHouseBalanceQty = 0.0;
    double dailyPgSqlBalanceAmount = 0.0;
    double dailyClickHouseBalanceAmount = 0.0;

    int totalPgSqlCount = 0;
    int totalClickHouseCount = 0;
    double totalPgSqlBalanceQty = 0.0;
    double totalClickHouseBalanceQty = 0.0;
    double totalPgSqlBalanceAmount = 0.0;
    double totalClickHouseBalanceAmount = 0.0;

    for (var key in sortedKeys) {
      var parts = key.split('|');
      var date = parts[0];
      var docNo = parts[1];

      if (currentDate != date) {
        if (currentDate.isNotEmpty) {
          final isMatch =
              dailyPgSqlCount == dailyClickHouseCount &&
              double.parse(dailyPgSqlBalanceQty.toStringAsFixed(6)) == double.parse(dailyClickHouseBalanceQty.toStringAsFixed(6)) &&
              double.parse(dailyPgSqlBalanceAmount.toStringAsFixed(6)) == double.parse(dailyClickHouseBalanceAmount.toStringAsFixed(6));

          comparisonRows.add(
            _ProcessStockComparisonRow(
              date: currentDate,
              docNo: "Total",
              pgSqlCount: dailyPgSqlCount,
              clickHouseCount: dailyClickHouseCount,
              pgSqlBalanceQty: dailyPgSqlBalanceQty,
              clickHouseBalanceQty: dailyClickHouseBalanceQty,
              pgSqlBalanceAmount: dailyPgSqlBalanceAmount,
              clickHouseBalanceAmount: dailyClickHouseBalanceAmount,
              isMatch: isMatch,
              isSummary: true,
            ),
          );
        }
        currentDate = date;
        dailyPgSqlCount = 0;
        dailyClickHouseCount = 0;
        dailyPgSqlBalanceQty = 0.0;
        dailyClickHouseBalanceQty = 0.0;
        dailyPgSqlBalanceAmount = 0.0;
        dailyClickHouseBalanceAmount = 0.0;
      }

      var pgSqlRow = pgSqlrows.firstWhere((element) => element['doc_date'].toString() == date && element['docno'].toString() == docNo, orElse: () => null);
      var clickHouseRow = clickHouserows.firstWhere((element) => element['doc_date'].toString() == date && element['docno'].toString() == docNo, orElse: () => null);
      int pgSqlCount = pgSqlRow != null ? int.parse(pgSqlRow['xcount'].toString()) : 0;
      int clickHouseCount = clickHouseRow != null ? int.parse(clickHouseRow['xcount'].toString()) : 0;
      double pgSqlBalanceQty = pgSqlRow != null ? double.parse(pgSqlRow['balanceqty'].toString()) : 0.0;
      double clickHouseBalanceQty = clickHouseRow != null ? double.parse(clickHouseRow['balanceqty'].toString()) : 0.0;
      double pgSqlBalanceAmount = pgSqlRow != null ? double.parse(pgSqlRow['balanceamount'].toString()) : 0.0;
      double clickHouseBalanceAmount = clickHouseRow != null ? double.parse(clickHouseRow['balanceamount'].toString()) : 0.0;

      dailyPgSqlCount += pgSqlCount;
      dailyClickHouseCount += clickHouseCount;
      dailyPgSqlBalanceQty += pgSqlBalanceQty;
      dailyClickHouseBalanceQty += clickHouseBalanceQty;
      dailyPgSqlBalanceAmount += pgSqlBalanceAmount;
      dailyClickHouseBalanceAmount += clickHouseBalanceAmount;

      totalPgSqlCount += pgSqlCount;
      totalClickHouseCount += clickHouseCount;
      totalPgSqlBalanceQty += pgSqlBalanceQty;
      totalClickHouseBalanceQty += clickHouseBalanceQty;
      totalPgSqlBalanceAmount += pgSqlBalanceAmount;
      totalClickHouseBalanceAmount += clickHouseBalanceAmount;

      final isMatch =
          pgSqlCount == clickHouseCount &&
          double.parse(pgSqlBalanceQty.toStringAsFixed(6)) == double.parse(clickHouseBalanceQty.toStringAsFixed(6)) &&
          double.parse(pgSqlBalanceAmount.toStringAsFixed(6)) == double.parse(clickHouseBalanceAmount.toStringAsFixed(6));

      comparisonRows.add(
        _ProcessStockComparisonRow(
          date: date,
          docNo: docNo,
          pgSqlCount: pgSqlCount,
          clickHouseCount: clickHouseCount,
          pgSqlBalanceQty: pgSqlBalanceQty,
          clickHouseBalanceQty: clickHouseBalanceQty,
          pgSqlBalanceAmount: pgSqlBalanceAmount,
          clickHouseBalanceAmount: clickHouseBalanceAmount,
          isMatch: isMatch,
        ),
      );
    }

    if (currentDate.isNotEmpty) {
      final isMatch =
          dailyPgSqlCount == dailyClickHouseCount &&
          double.parse(dailyPgSqlBalanceQty.toStringAsFixed(6)) == double.parse(dailyClickHouseBalanceQty.toStringAsFixed(6)) &&
          double.parse(dailyPgSqlBalanceAmount.toStringAsFixed(6)) == double.parse(dailyClickHouseBalanceAmount.toStringAsFixed(6));

      comparisonRows.add(
        _ProcessStockComparisonRow(
          date: currentDate,
          docNo: "Total",
          pgSqlCount: dailyPgSqlCount,
          clickHouseCount: dailyClickHouseCount,
          pgSqlBalanceQty: dailyPgSqlBalanceQty,
          clickHouseBalanceQty: dailyClickHouseBalanceQty,
          pgSqlBalanceAmount: dailyPgSqlBalanceAmount,
          clickHouseBalanceAmount: dailyClickHouseBalanceAmount,
          isMatch: isMatch,
          isSummary: true,
        ),
      );
    }

    // Grand Total
    final isTotalMatch =
        totalPgSqlCount == totalClickHouseCount &&
        double.parse(totalPgSqlBalanceQty.toStringAsFixed(6)) == double.parse(totalClickHouseBalanceQty.toStringAsFixed(6)) &&
        double.parse(totalPgSqlBalanceAmount.toStringAsFixed(6)) == double.parse(totalClickHouseBalanceAmount.toStringAsFixed(6));

    comparisonRows.add(
      _ProcessStockComparisonRow(
        date: "Grand Total",
        docNo: "",
        pgSqlCount: totalPgSqlCount,
        clickHouseCount: totalClickHouseCount,
        pgSqlBalanceQty: totalPgSqlBalanceQty,
        clickHouseBalanceQty: totalClickHouseBalanceQty,
        pgSqlBalanceAmount: totalPgSqlBalanceAmount,
        clickHouseBalanceAmount: totalClickHouseBalanceAmount,
        isMatch: isTotalMatch,
        isSummary: true,
      ),
    );
    bottomPanel = Container(
      padding: const EdgeInsets.all(16.0),
      color: global.theme.surfaceColor,
      width: double.infinity,
      child: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            ...infoMessages.map((message) => Padding(padding: const EdgeInsets.symmetric(vertical: 2.0), child: message)),
            const SizedBox(height: 12),
            _buildComparisonTable(comparisonRows),
          ],
        ),
      ),
    );
    setState(() {});
  }

  Widget _buildComparisonTable(List<_ProcessStockComparisonRow> rows) {
    if (rows.isEmpty) {
      return Text(global.language('no_data'), style: TextStyle(color: global.theme.textSecondaryColor));
    }

    return SingleChildScrollView(
      scrollDirection: Axis.horizontal,
      child: DataTable(
        dataRowMinHeight: 20,
        dataRowMaxHeight: 30,
        headingRowColor: WidgetStateProperty.all(global.theme.columnHeaderColor),
        columns: [
          DataColumn(label: Text(global.language('date'))),
          DataColumn(label: Text(global.language('docno'))),
          DataColumn(label: Text(global.language('pgsql_count'))),
          DataColumn(label: Text(global.language('clickhouse_count'))),
          DataColumn(label: Text(global.language('pgsql_qty'))),
          DataColumn(label: Text(global.language('clickhouse_qty'))),
          DataColumn(label: Text(global.language('pgsql_amount'))),
          DataColumn(label: Text(global.language('clickhouse_amount'))),
          DataColumn(label: Text(global.language('status'))),
        ],
        rows: rows
            .map(
              (row) => DataRow(
                color: WidgetStateProperty.all(row.isSummary ? global.theme.infoHighlightColor : (row.isMatch ? global.theme.positiveHighlightColor : global.theme.negativeHighlightColor)),
                cells: [
                  DataCell(Text(row.date, style: row.isSummary ? const TextStyle(fontWeight: FontWeight.bold) : null)),
                  DataCell(Text(row.docNo, style: row.isSummary ? const TextStyle(fontWeight: FontWeight.bold) : null)),
                  DataCell(Text(row.pgSqlCount.toString(), style: row.isSummary ? const TextStyle(fontWeight: FontWeight.bold) : null)),
                  DataCell(Text(row.clickHouseCount.toString(), style: row.isSummary ? const TextStyle(fontWeight: FontWeight.bold) : null)),
                  DataCell(Text(global.formatNumberRemoveRightZero(row.pgSqlBalanceQty), style: row.isSummary ? const TextStyle(fontWeight: FontWeight.bold) : null)),
                  DataCell(Text(global.formatNumberRemoveRightZero(row.clickHouseBalanceQty), style: row.isSummary ? const TextStyle(fontWeight: FontWeight.bold) : null)),
                  DataCell(Text(global.formatNumberRemoveRightZero(row.pgSqlBalanceAmount), style: row.isSummary ? const TextStyle(fontWeight: FontWeight.bold) : null)),
                  DataCell(Text(global.formatNumberRemoveRightZero(row.clickHouseBalanceAmount), style: row.isSummary ? const TextStyle(fontWeight: FontWeight.bold) : null)),
                  DataCell(
                    Text(
                      row.isMatch ? global.language('audit_match') : global.language('audit_not_match'),
                      style: TextStyle(color: row.isMatch ? global.theme.positiveHighlightTextColor : global.theme.negativeHighlightTextColor, fontWeight: FontWeight.w600),
                    ),
                  ),
                ],
              ),
            )
            .toList(),
      ),
    );
  }

  Future<void> _auditBarcodeItemCodeNotFound() async {
    List<String> messages = [];

    bottomPanel = Container(
      color: global.theme.surfaceColor,
      child: const Center(child: CircularProgressIndicator()),
    );
    setState(() {});
    try {
      // ตรวจสอบบาร์โค้ดที่ไม่มีรหัสสินค้า
      messages.add(global.language('audit_barcode_without_item_code_header'));
      String query = "select * from productbarcode where itemcode='' or itemcode is null order by barcode limit 1000";
      messages.add(query);
      final response = await pgSqlSelectGroup([query]);
      final rows = response["results"]?[0]?["rows"] as List?;

      if (rows == null) {
        AppLogger.debug('No data found');
        return;
      }
      if (rows.isEmpty) {
        messages.add(global.language('audit_barcode_without_item_code_no_error'));
      }

      for (var row in rows) {
        messages.add('Barcode: ${row['barcode']} {${row['name0']}} - ${global.language('item_code_not_found')}');
      }
    } catch (e) {
      AppLogger.error('Error: $e');
      bottomPanel = Container(
        color: global.theme.surfaceColor,
        child: Center(
          child: Text('Error: $e', style: TextStyle(color: global.theme.negativeHighlightTextColor)),
        ),
      );
      setState(() {});
    }
    bottomPanel = Container(
      padding: const EdgeInsets.all(16.0),
      color: global.theme.surfaceColor,
      width: double.infinity,
      child: SingleChildScrollView(
        child: Column(children: [...messages.map((error) => Text(error, style: const TextStyle(fontSize: 13)))]),
      ),
    );
    setState(() {});
  }

  Future<void> auditBarcodeUnitFormula() async {
    List<String> messages = [];

    bottomPanel = Container(
      color: global.theme.surfaceColor,
      child: const Center(child: CircularProgressIndicator()),
    );
    setState(() {});
    try {
      // ตรวจสอบบาร์โค้ดที่ไม่มีรหัสสินค้า
      messages.add(global.language('audit_barcode_unit_ratio_error_header'));
      String query = "select * from productbarcode where barcoderefunitstand is null or barcoderefunitstand=0 or barcoderefunitdivide is null or barcoderefunitdivide=0 order by barcode limit 1000";
      messages.add(query);
      final response = await pgSqlSelectGroup([query]);
      final rows = response["results"]?[0]?["rows"] as List?;

      if (rows == null) {
        AppLogger.debug('No data found');
        return;
      }
      if (rows.isEmpty) {
        messages.add(global.language('audit_barcode_unit_ratio_error_no_error'));
      }

      for (var row in rows) {
        messages.add('Barcode: ${row['barcode']} {${row['name0']}} - ${global.language('unit_ratio_error')}');
      }
    } catch (e) {
      AppLogger.error('Error: $e');
      bottomPanel = Container(
        color: global.theme.surfaceColor,
        child: Center(
          child: Text('Error: $e', style: TextStyle(color: global.theme.negativeHighlightTextColor)),
        ),
      );
      setState(() {});
    }
    bottomPanel = Container(
      padding: const EdgeInsets.all(16.0),
      color: global.theme.surfaceColor,
      width: double.infinity,
      child: SingleChildScrollView(
        child: Column(children: [...messages.map((error) => Text(error, style: const TextStyle(fontSize: 13)))]),
      ),
    );
    setState(() {});
  }

  Future<void> auditBarcodeUnitFormulaNoBarcodeRef() async {
    List<String> messages = [];

    bottomPanel = Container(
      color: global.theme.surfaceColor,
      child: const Center(child: CircularProgressIndicator()),
    );
    setState(() {});
    try {
      // ตรวจสอบบาร์โค้ดที่ไม่มีรหัสสินค้า
      String query = "select * from productbarcode where (barcoderef is null or barcoderef='') and barcoderefunitstand /  barcoderefunitdivide <> 1 order by barcode limit 1000";
      messages.add(query);
      messages.add(global.language('audit_barcode_non_1_1_ratio_no_ref_header'));
      final response = await pgSqlSelectGroup([query]);
      final rows = response["results"]?[0]?["rows"] as List?;

      if (rows == null) {
        AppLogger.debug('No data found');
        return;
      }
      if (rows.isEmpty) {
        messages.add(global.language('audit_barcode_non_1_1_ratio_no_ref_no_error'));
      }

      for (var row in rows) {
        messages.add('Barcode: ${row['barcode']} {${row['name0']}} {${row['barcoderefunitstand']}}: {${row['barcoderefunitdivide']}} - ${global.language('unit_ratio_error')}');
      }
    } catch (e) {
      AppLogger.error('Error: $e');
      bottomPanel = Container(
        color: global.theme.surfaceColor,
        child: Center(
          child: Text('Error: $e', style: TextStyle(color: global.theme.negativeHighlightTextColor)),
        ),
      );
      setState(() {});
    }
    bottomPanel = Container(
      padding: const EdgeInsets.all(16.0),
      color: global.theme.surfaceColor,
      width: double.infinity,
      child: SingleChildScrollView(
        child: Column(children: [...messages.map((error) => Text(error, style: const TextStyle(fontSize: 13)))]),
      ),
    );
    setState(() {});
  }

  Future<void> _showAuditMultiUnitDialog() async {
    return showDialog<void>(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          title: Text(global.language('audit_products_with_multiple_units')),
          content: Text(global.language('select_audit_type')),
          actions: <Widget>[
            TextButton(
              child: Text(global.language('all')),
              onPressed: () {
                Navigator.of(context).pop();
                auditProductMultiUnit(null);
              },
            ),
            TextButton(
              child: Text(global.language('by_item')),
              onPressed: () {
                Navigator.of(context).pop();
                _showItemCodeInputDialog();
              },
            ),
            TextButton(
              child: Text(global.language('cancel')),
              onPressed: () {
                Navigator.of(context).pop();
              },
            ),
          ],
        );
      },
    );
  }

  Future<void> _showItemCodeInputDialog() async {
    final TextEditingController itemCodeController = TextEditingController();

    return showDialog<void>(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          title: Text(global.language('specify_item_code')),
          content: TextField(
            controller: itemCodeController,
            decoration: InputDecoration(labelText: global.language('item_code'), hintText: global.language('enter_item_code_to_audit'), border: OutlineInputBorder()),
            autofocus: true,
          ),
          actions: <Widget>[
            TextButton(
              child: Text(global.language('cancel')),
              onPressed: () {
                Navigator.of(context).pop();
              },
            ),
            TextButton(
              child: Text(global.language('audit')),
              onPressed: () {
                final itemCode = itemCodeController.text.trim();
                Navigator.of(context).pop();
                if (itemCode.isNotEmpty) {
                  auditProductMultiUnit(itemCode);
                }
              },
            ),
          ],
        );
      },
    );
  }

  Future<void> auditProductMultiUnit(String? specificItemCode) async {
    List<Widget> widgetList = [];

    bottomPanel = Container(
      color: global.theme.surfaceColor,
      child: const Center(child: CircularProgressIndicator()),
    );
    setState(() {});
    try {
      // ดึงข้อมูลสินค้าและ barcode ทั้งหมดมาครั้งเดียว
      String query =
          "select itemcode, barcode, barcoderef, name0, unitcode, unitname, barcoderefunitstand, barcoderefunitdivide from productbarcode {where}order by itemcode, barcoderefunitstand/barcoderefunitdivide, barcode limit 5000";
      if (specificItemCode != null && specificItemCode.isNotEmpty) {
        query = query.replaceFirst('{where}', "where itemcode='$specificItemCode' ");
      } else {
        query = query.replaceFirst('{where}', '');
      }

      final response = await pgSqlSelectGroup([query]);
      final rows = response["results"]?[0]?["rows"] as List?;

      if (rows == null) {
        AppLogger.debug('No data found');
        return;
      }
      if (rows.isEmpty) {
        widgetList.add(Text(global.language('audit_products_with_multiple_units_no_error')));
      }

      // จัดกลุ่มข้อมูลตาม itemcode
      Map<String, List<dynamic>> groupedByItemCode = {};
      for (var row in rows) {
        String itemCode = row['itemcode'] ?? '';
        if (itemCode.isEmpty) continue;

        if (!groupedByItemCode.containsKey(itemCode)) {
          groupedByItemCode[itemCode] = [];
        }
        groupedByItemCode[itemCode]!.add(row);
      }

      // กรองเฉพาะสินค้าที่มี barcode มากกว่า 1
      var multiUnitProducts = groupedByItemCode.entries.where((entry) => entry.value.length > 1).toList();

      if (multiUnitProducts.isEmpty) {
        if (specificItemCode != null && specificItemCode.isNotEmpty) {
          widgetList.add(Text('${global.language('product_code')} $specificItemCode ${global.language('item_no_multiple_units_or_not_found')}'));
        } else {
          widgetList.add(Text(global.language('audit_products_with_multiple_units_no_error')));
        }
      }

      // สร้าง widget สำหรับแต่ละสินค้า
      for (var entry in multiUnitProducts) {
        String itemCode = entry.key;
        List<dynamic> barcodes = entry.value;

        widgetList.add(
          Container(
            padding: const EdgeInsets.all(8.0),
            width: 500,
            decoration: BoxDecoration(
              color: global.theme.infoHighlightColor,
              border: Border.all(color: global.theme.infoHighlightTextColor),
              borderRadius: BorderRadius.circular(4.0),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  '${global.language('item_code_colon')} $itemCode',
                  style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold, color: global.theme.infoHighlightTextColor),
                ),
                SizedBox(height: 8),
                Table(
                  border: TableBorder.all(color: global.theme.dividerBorderColor),
                  columnWidths: const {0: FlexColumnWidth(2), 1: FlexColumnWidth(3), 2: FlexColumnWidth(2), 3: FlexColumnWidth(1.5)},
                  children: [
                    TableRow(
                      decoration: BoxDecoration(color: global.theme.infoHighlightColor),
                      children: [
                        const Padding(
                          padding: EdgeInsets.all(2.0),
                          child: Text('Barcode', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13)),
                        ),
                        Padding(
                          padding: EdgeInsets.all(2.0),
                          child: Text(global.language('product_name'), style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13)),
                        ),
                        Padding(
                          padding: EdgeInsets.all(2.0),
                          child: Text(global.language('unit'), style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13)),
                        ),
                        Padding(
                          padding: EdgeInsets.all(2.0),
                          child: Text(global.language('ratio'), style: TextStyle(fontWeight: FontWeight.bold, fontSize: 13)),
                        ),
                      ],
                    ),
                    ...barcodes.map<TableRow>((barcodeRow) {
                      return TableRow(
                        children: [
                          Padding(
                            padding: const EdgeInsets.all(2.0),
                            child: Text(
                              '${barcodeRow['barcode']}${barcodeRow['barcoderef'] != null && barcodeRow['barcoderef'].toString().isNotEmpty ? ' (${barcodeRow['barcoderef']})' : ''}',
                              style: const TextStyle(fontSize: 12),
                            ),
                          ),
                          Padding(
                            padding: const EdgeInsets.all(2.0),
                            child: Text('${barcodeRow['name0']}', style: const TextStyle(fontSize: 12)),
                          ),
                          Padding(
                            padding: const EdgeInsets.all(2.0),
                            child: Text('${barcodeRow['unitcode']} - ${barcodeRow['unitname']}', style: const TextStyle(fontSize: 12)),
                          ),
                          Padding(
                            padding: const EdgeInsets.all(2.0),
                            child: Center(
                              child: Text(
                                '${global.formatNumberRemoveRightZero(double.tryParse(barcodeRow['barcoderefunitstand']?.toString() ?? '0') ?? 0)}:${global.formatNumberRemoveRightZero(double.tryParse(barcodeRow['barcoderefunitdivide']?.toString() ?? '0') ?? 0)}',
                                style: const TextStyle(fontSize: 12),
                              ),
                            ),
                          ),
                        ],
                      );
                    }),
                  ],
                ),
              ],
            ),
          ),
        );
      }
    } catch (e) {
      AppLogger.error('Error: $e');
      bottomPanel = Container(
        color: global.theme.surfaceColor,
        child: Center(
          child: Text('Error: $e', style: TextStyle(color: global.theme.negativeHighlightTextColor)),
        ),
      );
      setState(() {});
      return;
    }
    bottomPanel = Container(
      padding: const EdgeInsets.all(16.0),
      color: global.theme.surfaceColor,
      width: double.infinity,
      child: ListView.builder(
        itemCount: widgetList.length,
        itemBuilder: (context, index) {
          return Padding(
            padding: const EdgeInsets.all(4.0),
            child: Center(child: widgetList[index]),
          );
        },
      ),
    );
    setState(() {});
  }

}

class _ProcessStockComparisonRow {
  const _ProcessStockComparisonRow({
    required this.date,
    required this.docNo,
    required this.pgSqlCount,
    required this.clickHouseCount,
    required this.pgSqlBalanceQty,
    required this.clickHouseBalanceQty,
    required this.pgSqlBalanceAmount,
    required this.clickHouseBalanceAmount,
    required this.isMatch,
    this.isSummary = false,
  });

  final String date;
  final String docNo;
  final int pgSqlCount;
  final int clickHouseCount;
  final double pgSqlBalanceQty;
  final double clickHouseBalanceQty;
  final double pgSqlBalanceAmount;
  final double clickHouseBalanceAmount;
  final bool isMatch;
  final bool isSummary;
}
