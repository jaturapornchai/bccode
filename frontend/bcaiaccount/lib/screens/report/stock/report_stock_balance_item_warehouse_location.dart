import 'dart:convert';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:pdf/pdf.dart' as pdf;
import 'package:printing/printing.dart' as printing;
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/repositories/report_repository.dart';
import 'package:smlaicloud/screens/report/report_stock_widget.dart';
import 'package:smlaicloud/utils/date_picker.dart';
import 'package:smlaicloud/utils/select_product_barcode.dart';
import 'package:syncfusion_flutter_pdfviewer/pdfviewer.dart';
import '../pdf_saver.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class ReportStockBalanceItemWareHouseLocation extends StatefulWidget {
  const ReportStockBalanceItemWareHouseLocation({super.key});

  @override
  ReportStockBalanceItemWareHouseLocationState createState() =>
      ReportStockBalanceItemWareHouseLocationState();
}

class ReportStockBalanceItemWareHouseLocationState
    extends State<ReportStockBalanceItemWareHouseLocation>
    with SingleTickerProviderStateMixin {
  int condition = 1;
  List<String> titles = [
    "แสดงสินค้าคงเหลือตามสินค้า",
    "แสดงสินค้าคงเหลือตามสินค้า -> คลัง",
    "แสดงสินค้าคงเหลือตามสินค้า -> คลัง -> ที่เก็บ",
  ];
  bool processSuccess = false;
  bool pdfCreated = false;
  bool pdfDownloaded = false;
  String pdfPath = "";
  String guid = "";
  global.ProcessState processState = global.ProcessState.idle;
  Widget processWidgetStatus = const SizedBox.shrink();
  Widget resultScreenWidget = Container();
  Widget pdfViewWidget = Container();
  PdfViewerController pdfViewerController = PdfViewerController();
  TextEditingController pdfSearchController = TextEditingController();
  late Uint8List pdfData;
  DateTime conditionFinalDate = DateTime.now();
  late TabController tabController;
  double resultFontScale = 1.0;
  List<ReportStockBalanceModel> dataList = [];
  late ReportStockConditionClass reportCondition;
  ScrollController scrollController = ScrollController();
  int dataOffset = 200;
  bool showHelp = false;
  bool isLoadingMore = false;

  // Enhanced color scheme for better UI
  final Color primaryColor = Colors.blue.shade700;
  final Color secondaryColor = Colors.blue.shade100;
  final Color accentColor = Colors.amber.shade600;
  final Color successColor = Colors.green.shade600;
  final Color errorColor = Colors.red.shade600;
  final Color backgroundColor = Colors.grey.shade50;

  final Color headerColor = Color(0xFFE3F2FD);
  final Color footerColor = Color(0xFFE1F5FE);
  final Color rowEvenColor = Color(0xFFF5F5F5);
  final Color rowOddColor = Colors.white;
  final Color warehouseColor = Color(0xFFFFF8E1);
  final Color locationColor = Color(0xFFE8F5E9);
  final Color helpColor = Colors.blue.shade50;

  // Text styles for consistent typography
  late TextStyle titleStyle;
  late TextStyle subtitleStyle;
  late TextStyle normalStyle;
  late TextStyle emphasisStyle;
  late TextStyle highlightStyle;

  @override
  void initState() {
    super.initState();
    _initializeStyles();
    reportCondition = ReportStockConditionClass(
      onStateUpdate: () {
        setState(() {});
      },
    );
    tabController = TabController(length: 3, vsync: this);
    tabController.addListener(() async {
      if (tabController.indexIsChanging) {
        if (tabController.index == 2) {
          if (pdfDownloaded == false) {
            await reloadPdf();
          }
        }
        setState(() {});
      }
    });
    scrollController.addListener(() {
      if (scrollController.position.pixels ==
              scrollController.position.maxScrollExtent &&
          !isLoadingMore) {
        _loadMoreData();
      }
    });

    // Show help panel with slight delay for better UX
    Future.delayed(Duration(milliseconds: 500), () {
      setState(() {
        showHelp = true;
      });
    });

    reportCondition.reloadWareHouseCode();
  }

  void _initializeStyles() {
    titleStyle = TextStyle(
      fontSize: 16,
      fontWeight: FontWeight.bold,
      color: primaryColor,
    );

    subtitleStyle = TextStyle(
      fontSize: 14,
      fontWeight: FontWeight.w500,
      color: primaryColor.withValues(alpha: 0.8),
    );

    normalStyle = TextStyle(fontSize: 13, color: Colors.grey.shade800);

    emphasisStyle = TextStyle(
      fontSize: 13,
      fontWeight: FontWeight.w500,
      color: primaryColor.withValues(alpha: 0.9),
    );

    highlightStyle = TextStyle(
      fontSize: 13,
      fontWeight: FontWeight.bold,
      color: accentColor,
    );
  }

  void _loadMoreData() {
    if (isLoadingMore) return;
    setState(() {
      isLoadingMore = true;
    });

    reloadData(dataList.length, dataOffset)
        .then((_) {
          if (mounted) {
            setState(() {
              isLoadingMore = false;
            });
          }
        })
        .catchError((error) {
          if (mounted) {
            setState(() {
              isLoadingMore = false;
            });
          }
          if (kDebugMode) {
            AppLogger.error("Error loading more data: $error");
          }
        });
  }

  @override
  void dispose() {
    tabController.dispose();
    pdfViewerController.dispose();
    pdfSearchController.dispose();
    super.dispose();
  }

  Future<void> reloadPdf() async {
    if (pdfCreated == false) {
      var payload = {
        "shop_id": global.getShopId(),
        "command_id":
            "stock_balance_by_product_and_warehouse_and_location_create_pdf",
        "guid": guid,
        "condition": condition.toString(),
        "final_date":
            "${conditionFinalDate.year}-${conditionFinalDate.month.toString().padLeft(2, '0')}-${conditionFinalDate.day.toString().padLeft(2, '0')}",
      };
      var jsonPayload = jsonEncode(payload);
      var jsonResult = await global.reportServicePost(jsonPayload);
      if (jsonResult['code'] == 200) {
        pdfCreated = true;
        pdfDownloaded = false;
        pdfPath = jsonResult['path'];
        await reloadPdf();
      } else {
        throw Exception('PDF Creation Error: ${jsonResult['message']}');
      }
    } else {
      if (pdfDownloaded == false) {
        var payload = {
          "shop_id": global.getShopId(),
          "command_id": "data_bin",
          "guid": guid,
          "path": pdfPath,
        };
        var jsonResult = await global.reportServiceGetBinary(payload);
        pdfDownloaded = true;
        pdfData = jsonResult['binaryData'];
        if (kIsWeb) {
          String base64Data = base64Encode(pdfData);
          String pdfInBase64 = "data:application/pdf;base64,$base64Data";
          pdfViewWidget = SfPdfViewer.network(
            pdfInBase64,
            controller: pdfViewerController,
          );
        } else {
          pdfViewWidget = SfPdfViewer.memory(
            pdfData,
            controller: pdfViewerController,
          );
        }
        setState(() {});
      }
    }
  }

  Widget renderData(List<ReportStockBalanceModel> dataList) {
    double fontSize = 12.0 * resultFontScale;
    var header = Container(
      decoration: BoxDecoration(
        color: headerColor,
        border: Border.all(color: Colors.blue.shade300),
        borderRadius: BorderRadius.only(
          topLeft: Radius.circular(8),
          topRight: Radius.circular(8),
        ),
      ),
      padding: EdgeInsets.all(8),
      child: Column(
        children: [
          Row(
            children: [
              Expanded(
                flex: 1,
                child: Padding(
                  padding: EdgeInsets.all(2),
                  child: Text(
                    global.language("product_code"),
                    style: TextStyle(
                      fontSize: fontSize,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ),
              Expanded(
                flex: 2,
                child: Padding(
                  padding: EdgeInsets.all(2),
                  child: Text(
                    global.language("import_product_detail.product_name"),
                    style: TextStyle(
                      fontSize: fontSize,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ),
              Expanded(
                flex: 1,
                child: Padding(
                  padding: EdgeInsets.all(2),
                  child: Text(
                    global.language("barcode"),
                    style: TextStyle(
                      fontSize: fontSize,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ),
              Expanded(
                flex: 1,
                child: Padding(
                  padding: EdgeInsets.all(2),
                  child: Text(
                    global.language("unit"),
                    style: TextStyle(
                      fontSize: fontSize,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ),
              Expanded(
                flex: 1,
                child: Padding(
                  padding: EdgeInsets.all(2),
                  child: Text(
                    global.language("balance"),
                    textAlign: TextAlign.right,
                    style: TextStyle(
                      fontSize: fontSize,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ),
              Expanded(
                flex: 1,
                child: Padding(
                  padding: EdgeInsets.all(2),
                  child: Text(
                    global.language("average_cost"),
                    textAlign: TextAlign.right,
                    style: TextStyle(
                      fontSize: fontSize,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ),
              Expanded(
                flex: 1,
                child: Padding(
                  padding: EdgeInsets.all(2),
                  child: Text(
                    global.language("balance_amount"),
                    textAlign: TextAlign.right,
                    style: TextStyle(
                      fontSize: fontSize,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ),
              Expanded(
                flex: 2,
                child: Padding(
                  padding: EdgeInsets.all(2),
                  child: Text(
                    global.language("balance_qty_readable"),
                    style: TextStyle(
                      fontSize: fontSize,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                ),
              ),
            ],
          ),
          if (condition == 2 || condition == 3)
            Row(
              children: [
                Expanded(
                  flex: 1,
                  child: Padding(
                    padding: EdgeInsets.all(2),
                    child: Text(
                      "",
                      style: TextStyle(
                        fontSize: fontSize,
                        fontWeight: FontWeight.normal,
                      ),
                    ),
                  ),
                ),
                Expanded(
                  flex: 2,
                  child: Padding(
                    padding: EdgeInsets.all(2),
                    child: Text(
                      global.language("warehouse"),
                      style: TextStyle(
                        fontSize: fontSize,
                        fontWeight: FontWeight.normal,
                      ),
                    ),
                  ),
                ),
                Expanded(
                  flex: 1,
                  child: Padding(
                    padding: EdgeInsets.all(2),
                    child: Text(
                      "",
                      style: TextStyle(
                        fontSize: fontSize,
                        fontWeight: FontWeight.normal,
                      ),
                    ),
                  ),
                ),
                Expanded(
                  flex: 1,
                  child: Padding(
                    padding: EdgeInsets.all(2),
                    child: Text(
                      global.language("balance"),
                      textAlign: TextAlign.right,
                      style: TextStyle(
                        fontSize: fontSize,
                        fontWeight: FontWeight.normal,
                      ),
                    ),
                  ),
                ),
                Expanded(
                  flex: 1,
                  child: Padding(
                    padding: EdgeInsets.all(2),
                    child: Text(
                      global.language("average_cost"),
                      textAlign: TextAlign.right,
                      style: TextStyle(
                        fontSize: fontSize,
                        fontWeight: FontWeight.normal,
                      ),
                    ),
                  ),
                ),
                Expanded(
                  flex: 1,
                  child: Padding(
                    padding: EdgeInsets.all(2),
                    child: Text(
                      global.language("balance_amount"),
                      textAlign: TextAlign.right,
                      style: TextStyle(
                        fontSize: fontSize,
                        fontWeight: FontWeight.normal,
                      ),
                    ),
                  ),
                ),
                Expanded(
                  flex: 2,
                  child: Padding(
                    padding: EdgeInsets.all(2),
                    child: Text(
                      global.language("balance_qty_readable"),
                      style: TextStyle(
                        fontSize: fontSize,
                        fontWeight: FontWeight.normal,
                      ),
                    ),
                  ),
                ),
              ],
            ),
          if (condition == 3)
            Row(
              children: [
                Expanded(
                  flex: 1,
                  child: Padding(
                    padding: EdgeInsets.all(2),
                    child: Text(
                      "",
                      style: TextStyle(
                        fontSize: fontSize,
                        fontWeight: FontWeight.normal,
                      ),
                    ),
                  ),
                ),
                Expanded(
                  flex: 1,
                  child: Padding(
                    padding: EdgeInsets.all(2),
                    child: Text(
                      "",
                      style: TextStyle(
                        fontSize: fontSize,
                        fontWeight: FontWeight.normal,
                      ),
                    ),
                  ),
                ),
                Expanded(
                  flex: 2,
                  child: Padding(
                    padding: EdgeInsets.all(2),
                    child: Text(
                      "",
                      style: TextStyle(
                        fontSize: fontSize,
                        fontWeight: FontWeight.normal,
                      ),
                    ),
                  ),
                ),
                Expanded(
                  flex: 1,
                  child: Padding(
                    padding: EdgeInsets.all(2),
                    child: Text(
                      global.language("location"),
                      style: TextStyle(
                        fontSize: fontSize,
                        fontWeight: FontWeight.normal,
                      ),
                    ),
                  ),
                ),
                Expanded(
                  flex: 1,
                  child: Padding(
                    padding: EdgeInsets.all(2),
                    child: Text(
                      global.language("balance"),
                      textAlign: TextAlign.right,
                      style: TextStyle(
                        fontSize: fontSize,
                        fontWeight: FontWeight.normal,
                      ),
                    ),
                  ),
                ),
                Expanded(
                  flex: 1,
                  child: Padding(
                    padding: EdgeInsets.all(2),
                    child: Text(
                      global.language("average_cost"),
                      textAlign: TextAlign.right,
                      style: TextStyle(
                        fontSize: fontSize,
                        fontWeight: FontWeight.normal,
                      ),
                    ),
                  ),
                ),
                Expanded(
                  flex: 1,
                  child: Padding(
                    padding: EdgeInsets.all(2),
                    child: Text(
                      global.language("balance_amount"),
                      textAlign: TextAlign.right,
                      style: TextStyle(
                        fontSize: fontSize,
                        fontWeight: FontWeight.normal,
                      ),
                    ),
                  ),
                ),
                Expanded(
                  flex: 2,
                  child: Padding(
                    padding: EdgeInsets.all(2),
                    child: Text(
                      global.language("balance_qty_readable"),
                      style: TextStyle(
                        fontSize: fontSize,
                        fontWeight: FontWeight.normal,
                      ),
                    ),
                  ),
                ),
              ],
            ),
        ],
      ),
    );
    List<Widget> data = [];
    bool isEvenRow = true;
    TextStyle itemStyle = (condition == 1)
        ? TextStyle(fontSize: fontSize, fontWeight: FontWeight.normal)
        : TextStyle(fontSize: fontSize, fontWeight: FontWeight.bold);
    for (var item in dataList) {
      data.add(
        Container(
          decoration: BoxDecoration(
            color: isEvenRow ? rowEvenColor : rowOddColor,
            border: Border(
              bottom: BorderSide(color: Colors.grey.shade300, width: 1),
            ),
          ),
          child: Row(
            children: [
              Expanded(
                flex: 1,
                child: Padding(
                  padding: EdgeInsets.all(2),
                  child: Text(item.itemCode, style: itemStyle),
                ),
              ),
              Expanded(
                flex: 2,
                child: Padding(
                  padding: EdgeInsets.all(2),
                  child: Text(item.itemNameName, style: itemStyle),
                ),
              ),
              Expanded(
                flex: 1,
                child: Padding(
                  padding: EdgeInsets.all(2),
                  child: Text(item.barCodeList, style: itemStyle),
                ),
              ),
              Expanded(
                flex: 1,
                child: Padding(
                  padding: EdgeInsets.all(2),
                  child: Text(
                    "${item.unitCode ?? ""}/${item.unitName ?? ""}",
                    style: itemStyle,
                  ),
                ),
              ),
              Expanded(
                flex: 1,
                child: Padding(
                  padding: EdgeInsets.all(2),
                  child: Text(
                    global.formatNumber(item.balanceQty ?? 0),
                    style: itemStyle,
                    textAlign: TextAlign.right,
                  ),
                ),
              ),
              Expanded(
                flex: 1,
                child: Padding(
                  padding: EdgeInsets.all(2),
                  child: Text(
                    global.formatNumber(item.averageCost ?? 0),
                    style: itemStyle,
                    textAlign: TextAlign.right,
                  ),
                ),
              ),
              Expanded(
                flex: 1,
                child: Padding(
                  padding: EdgeInsets.all(2),
                  child: Text(
                    global.formatNumber(item.balanceAmount ?? 0),
                    style: itemStyle,
                    textAlign: TextAlign.right,
                  ),
                ),
              ),
              Expanded(
                flex: 2,
                child: Padding(
                  padding: EdgeInsets.all(2),
                  child: Text(item.balanceWord ?? "", style: itemStyle),
                ),
              ),
            ],
          ),
        ),
      );
      isEvenRow = !isEvenRow;
      if (condition == 2 || condition == 3) {
        for (var warehouse in item.warehouses ?? []) {
          data.add(
            Container(
              decoration: BoxDecoration(
                color: warehouseColor,
                border: Border(
                  bottom: BorderSide(color: Colors.grey.shade300, width: 1),
                ),
              ),
              child: Row(
                children: [
                  Expanded(
                    flex: 1,
                    child: Padding(
                      padding: EdgeInsets.all(2),
                      child: Text(
                        "",
                        style: TextStyle(
                          fontSize: fontSize,
                          fontWeight: FontWeight.normal,
                        ),
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 2,
                    child: Padding(
                      padding: EdgeInsets.all(2),
                      child: Text(
                        warehouse.warehouseCode,
                        style: TextStyle(
                          fontSize: fontSize,
                          fontWeight: FontWeight.normal,
                        ),
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 1,
                    child: Padding(
                      padding: EdgeInsets.all(2),
                      child: Text(
                        "",
                        style: TextStyle(
                          fontSize: fontSize,
                          fontWeight: FontWeight.normal,
                        ),
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 1,
                    child: Padding(
                      padding: EdgeInsets.all(2),
                      child: Text(
                        global.formatNumber(warehouse.balanceQty),
                        textAlign: TextAlign.right,
                        style: TextStyle(
                          fontSize: fontSize,
                          fontWeight: FontWeight.normal,
                        ),
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 1,
                    child: Padding(
                      padding: EdgeInsets.all(2),
                      child: Text(
                        global.formatNumber(warehouse.averageCost),
                        textAlign: TextAlign.right,
                        style: TextStyle(
                          fontSize: fontSize,
                          fontWeight: FontWeight.normal,
                        ),
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 1,
                    child: Padding(
                      padding: EdgeInsets.all(2),
                      child: Text(
                        global.formatNumber(warehouse.balanceAmount),
                        textAlign: TextAlign.right,
                        style: TextStyle(
                          fontSize: fontSize,
                          fontWeight: FontWeight.normal,
                        ),
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 2,
                    child: Padding(
                      padding: EdgeInsets.all(2),
                      child: Text(
                        warehouse.balanceWord ?? "",
                        style: TextStyle(
                          fontSize: fontSize,
                          fontWeight: FontWeight.normal,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          );
          if (condition == 3) {
            for (var location in warehouse.locations ?? []) {
              data.add(
                Container(
                  decoration: BoxDecoration(
                    color: locationColor,
                    border: Border(
                      bottom: BorderSide(color: Colors.grey.shade300, width: 1),
                    ),
                  ),
                  child: Row(
                    children: [
                      Expanded(
                        flex: 1,
                        child: Padding(
                          padding: EdgeInsets.all(2),
                          child: Text(
                            "",
                            style: TextStyle(
                              fontSize: fontSize,
                              fontWeight: FontWeight.normal,
                            ),
                          ),
                        ),
                      ),
                      Expanded(
                        flex: 1,
                        child: Padding(
                          padding: EdgeInsets.all(2),
                          child: Text(
                            "",
                            style: TextStyle(
                              fontSize: fontSize,
                              fontWeight: FontWeight.normal,
                            ),
                          ),
                        ),
                      ),
                      Expanded(
                        flex: 2,
                        child: Padding(
                          padding: EdgeInsets.all(2),
                          child: Text(
                            "",
                            style: TextStyle(
                              fontSize: fontSize,
                              fontWeight: FontWeight.normal,
                            ),
                          ),
                        ),
                      ),
                      Expanded(
                        flex: 1,
                        child: Padding(
                          padding: EdgeInsets.all(2),
                          child: Text(
                            location.locationCode,
                            style: TextStyle(
                              fontSize: fontSize,
                              fontWeight: FontWeight.normal,
                            ),
                          ),
                        ),
                      ),
                      Expanded(
                        flex: 1,
                        child: Padding(
                          padding: EdgeInsets.all(2),
                          child: Text(
                            global.formatNumber(location.balanceQty),
                            textAlign: TextAlign.right,
                            style: TextStyle(
                              fontSize: fontSize,
                              fontWeight: FontWeight.normal,
                            ),
                          ),
                        ),
                      ),
                      Expanded(
                        flex: 1,
                        child: Padding(
                          padding: EdgeInsets.all(2),
                          child: Text(
                            "",
                            textAlign: TextAlign.right,
                            style: TextStyle(
                              fontSize: fontSize,
                              fontWeight: FontWeight.normal,
                            ),
                          ),
                        ),
                      ),
                      Expanded(
                        flex: 1,
                        child: Padding(
                          padding: EdgeInsets.all(2),
                          child: Text(
                            global.formatNumber(
                              (location.balanceQty * warehouse.averageCost),
                            ),
                            textAlign: TextAlign.right,
                            style: TextStyle(
                              fontSize: fontSize,
                              fontWeight: FontWeight.normal,
                            ),
                          ),
                        ),
                      ),
                      Expanded(
                        flex: 2,
                        child: Padding(
                          padding: EdgeInsets.all(2),
                          child: Text(
                            location.balanceWord ?? "",
                            style: TextStyle(
                              fontSize: fontSize,
                              fontWeight: FontWeight.normal,
                            ),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              );
            }
          }
        }
      }
    }
    resultScreenWidget = Container(
      margin: EdgeInsets.symmetric(horizontal: 8, vertical: 8),
      child: Card(
        elevation: 4,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        child: Container(
          color: Colors.white,
          child: Column(
            children: [
              header,
              Expanded(
                child: SingleChildScrollView(
                  physics: const AlwaysScrollableScrollPhysics(),
                  controller: scrollController,
                  scrollDirection: Axis.vertical,
                  child: Column(
                    children: [
                      Container(
                        decoration: BoxDecoration(
                          color: Colors.white,
                          border: Border.all(color: Colors.grey.shade300),
                        ),
                        child: Column(children: data),
                      ),
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
    return resultScreenWidget;
  }

  Future<void> reloadData(int offset, int limit) async {
    try {
      // ใช้ ReportRepository.loadReportData แทน - generic function
      await ReportRepository.loadReportData<ReportStockBalanceModel>(
        guid: guid,
        offset: offset,
        limit: limit,
        dataList: dataList,
        fromJson: ReportStockBalanceModel.fromJson,
      );

      // Render เฉพาะตอนสำเร็จ
      setState(() {
        renderData(dataList);
      });
    } catch (e, t) {
      if (kDebugMode) {
        AppLogger.error("Error reloadData: $e\n$t");
      }
      // แสดง error ผ่าน snackbar แทน resultScreenWidget
      if (mounted) {
        global.showErrorSnackBar(context, "${global.language("cannot_load_data")}: $e");
      }
    }
  }

  Widget _buildHelpPanel() {
    return AnimatedContainer(
      duration: Duration(milliseconds: 300),
      height: showHelp ? null : 0,
      child: Card(
        margin: EdgeInsets.all(8),
        elevation: 4,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        color: helpColor,
        child: Padding(
          padding: EdgeInsets.all(16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Icon(Icons.help_outline, color: primaryColor, size: 24),
                  SizedBox(width: 12),
                  Text(
                    global.language("stock_report_usage_guide"),
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 16,
                      color: primaryColor,
                    ),
                  ),
                  Spacer(),
                  IconButton(
                    icon: Icon(Icons.close, color: Colors.grey.shade700),
                    onPressed: () {
                      setState(() {
                        showHelp = false;
                      });
                    },
                    tooltip: global.language("close_help_window"),
                  ),
                ],
              ),
              Divider(color: secondaryColor, thickness: 1.5),
              _buildHelpStep(
                "1",
                global.language("select_report_type"),
                "เลือกรูปแบบรายงานที่ต้องการจาก 3 รูปแบบ",
                Icons.category,
              ),
              _buildHelpStep(
                "2",
                global.language("select_date"),
                "เลือกวันที่ที่ต้องการแสดงยอดคงเหลือ",
                Icons.calendar_today,
              ),
              _buildHelpStep(
                "3",
                global.language("display_options"),
                global.language("show_only_with_balance_or_all"),
                Icons.filter_alt,
              ),
              _buildHelpStep(
                "4",
                global.language("report_condition_select_barcode"),
                global.language("select_products_or_show_all"),
                Icons.inventory,
              ),
              _buildHelpStep(
                "5",
                "เลือกคลัง/ที่เก็บ",
                "เลือกคลังและที่เก็บที่ต้องการแสดง",
                Icons.store,
              ),
              _buildHelpStep(
                "6",
                global.language("process"),
                "กดปุ่ม \"ประมวลผล\" เพื่อสร้างรายงาน",
                Icons.play_arrow,
              ),
              SizedBox(height: 8),
              Center(
                child: ElevatedButton.icon(
                  icon: Icon(Icons.help_outline),
                  label: Text(global.language("hide_tips")),
                  onPressed: () {
                    setState(() {
                      showHelp = false;
                    });
                  },
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.grey.shade200,
                    foregroundColor: Colors.grey.shade800,
                    elevation: 0,
                    padding: EdgeInsets.symmetric(horizontal: 20, vertical: 10),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(8),
                    ),
                  ),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildHelpStep(
    String step,
    String title,
    String description,
    IconData icon,
  ) {
    return Padding(
      padding: EdgeInsets.symmetric(vertical: 10),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Container(
            width: 28,
            height: 28,
            decoration: BoxDecoration(
              color: primaryColor,
              shape: BoxShape.circle,
              boxShadow: [
                BoxShadow(
                  color: Colors.black.withValues(alpha: 0.1),
                  blurRadius: 3,
                  offset: Offset(0, 1),
                ),
              ],
            ),
            child: Center(
              child: Text(
                step,
                style: TextStyle(
                  color: Colors.white,
                  fontWeight: FontWeight.bold,
                  fontSize: 14,
                ),
              ),
            ),
          ),
          SizedBox(width: 12),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Row(
                  children: [
                    Icon(icon, size: 18, color: primaryColor),
                    SizedBox(width: 8),
                    Text(
                      title,
                      style: TextStyle(
                        fontWeight: FontWeight.bold,
                        color: primaryColor,
                        fontSize: 15,
                      ),
                    ),
                  ],
                ),
                SizedBox(height: 4),
                Text(
                  description,
                  style: TextStyle(fontSize: 14, color: Colors.grey.shade800),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildReportHeader() {
    return Card(
      margin: EdgeInsets.symmetric(horizontal: 12),
      elevation: 4,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Container(
        padding: EdgeInsets.all(16),
        decoration: BoxDecoration(
          gradient: LinearGradient(
            colors: [primaryColor, primaryColor.withBlue(255)],
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
          ),
          borderRadius: BorderRadius.circular(12),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.inventory_2, color: Colors.white, size: 28),
                SizedBox(width: 12),
                Text(
                  global.language("inventory_report"),
                  style: TextStyle(
                    fontSize: 22,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                  ),
                ),
              ],
            ),
            SizedBox(height: 8),
            Divider(color: Colors.white.withValues(alpha: 0.3), height: 24),
            Text(
              "แสดงข้อมูลสินค้าคงเหลือตามสินค้า คลัง และที่เก็บ",
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w500,
                color: Colors.white.withValues(alpha: 0.95),
              ),
            ),
            SizedBox(height: 8),
            Text(
              "เลือกรูปแบบรายงานและเงื่อนไขต่างๆ ด้านล่างเพื่อสร้างรายงานตามที่ต้องการ",
              style: TextStyle(
                fontSize: 14,
                color: Colors.white.withValues(alpha: 0.85),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildConditionSelect() {
    return Container(
      color: backgroundColor,
      child: SingleChildScrollView(
        padding: EdgeInsets.symmetric(vertical: 8),
        child: Column(
          children: [
            _buildHelpPanel(),
            _buildReportHeader(),
            SizedBox(height: 12),
            _buildReportTypePanel(),
            SizedBox(height: 12),
            _buildDateSelectionPanel(),
            SizedBox(height: 12),
            _buildFilterOptionsPanel(),
            SizedBox(height: 12),
            _buildProductSelectionPanel(),
            if (condition == 2 || condition == 3)
              _buildWarehouseSelectionPanel(),
            if (condition == 3) _buildLocationSelectionPanel(),
            _buildProcessButtonPanel(),
            SizedBox(height: 20),
          ],
        ),
      ),
    );
  }

  Widget _buildReportTypePanel() {
    return Card(
      margin: EdgeInsets.symmetric(horizontal: 12),
      elevation: 2,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Padding(
        padding: EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.category, color: primaryColor, size: 20),
                SizedBox(width: 10),
                Text(global.language("select_report_type"), style: titleStyle),
              ],
            ),
            SizedBox(height: 12),
            Container(
              decoration: BoxDecoration(
                color: Colors.grey.shade50,
                borderRadius: BorderRadius.circular(8),
              ),
              child: Column(
                children: [
                  _buildReportTypeButton(
                    1,
                    Icons.inventory,
                    global.language('balance_by_product'),
                    'แสดงสินค้าและยอดคงเหลือรวมทั้งหมด',
                  ),
                  Divider(height: 1, thickness: 1, color: Colors.grey.shade200),
                  _buildReportTypeButton(
                    2,
                    Icons.store,
                    global.language('balance_by_warehouse'),
                    'แสดงสินค้าและแยกตามคลังสินค้า',
                  ),
                  Divider(height: 1, thickness: 1, color: Colors.grey.shade200),
                  _buildReportTypeButton(
                    3,
                    Icons.grid_view,
                    'ยอดคงเหลือตามที่เก็บ',
                    'แสดงสินค้า คลัง และแยกตามที่เก็บ',
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildReportTypeButton(
    int value,
    IconData icon,
    String title,
    String subtitle,
  ) {
    final bool isSelected = condition == value;

    return InkWell(
      onTap: () {
        setState(() {
          condition = value;
        });
      },
      borderRadius: BorderRadius.circular(8),
      splashColor: secondaryColor,
      child: Container(
        padding: EdgeInsets.symmetric(vertical: 16, horizontal: 16),
        decoration: BoxDecoration(
          color: isSelected
              ? secondaryColor.withValues(alpha: 0.3)
              : Colors.transparent,
          borderRadius: BorderRadius.circular(8),
        ),
        child: Row(
          children: [
            Container(
              padding: EdgeInsets.all(10),
              decoration: BoxDecoration(
                color: isSelected ? primaryColor : Colors.grey.shade200,
                shape: BoxShape.circle,
              ),
              child: Icon(
                icon,
                color: isSelected ? Colors.white : Colors.grey.shade600,
                size: 22,
              ),
            ),
            SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: TextStyle(
                      fontSize: 15,
                      fontWeight: FontWeight.bold,
                      color: isSelected ? primaryColor : Colors.grey.shade800,
                    ),
                  ),
                  SizedBox(height: 4),
                  Text(
                    subtitle,
                    style: TextStyle(
                      fontSize: 13,
                      color: isSelected
                          ? primaryColor.withValues(alpha: 0.8)
                          : Colors.grey.shade600,
                    ),
                  ),
                ],
              ),
            ),
            if (isSelected)
              Icon(Icons.check_circle, color: primaryColor, size: 24),
          ],
        ),
      ),
    );
  }

  Widget _buildDateSelectionPanel() {
    return Card(
      margin: EdgeInsets.symmetric(horizontal: 12),
      elevation: 2,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Padding(
        padding: EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.calendar_today, color: primaryColor, size: 20),
                SizedBox(width: 10),
                Text(global.language("report_date"), style: titleStyle),
                Tooltip(
                  message:
                      "เลือกวันที่ที่ต้องการแสดงยอดคงเหลือ ณ สิ้นวันของวันที่เลือก",
                  child: Icon(
                    Icons.info_outline,
                    size: 16,
                    color: Colors.grey.shade600,
                  ),
                ),
              ],
            ),
            SizedBox(height: 12),
            Text(
              "โปรดเลือกวันที่ที่ต้องการทราบยอดคงเหลือ (ระบบจะคำนวณยอด ณ สิ้นวันของวันที่เลือก)",
              style: normalStyle,
            ),
            SizedBox(height: 12),
            CustomDatePicker(
              labelText: global.language('balance_as_of_date'),
              initialDate: conditionFinalDate,
              // useBuddhistCalendar และ languageCode จะอ่านจากสาขาอัตโนมัติ
              onDateSelected: (date) {
                if (date != null) {
                  setState(() {
                    conditionFinalDate = date;
                  });
                }
              },
              decoration: InputDecoration(
                prefixIcon: Icon(Icons.calendar_today, color: primaryColor),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: BorderSide(color: Colors.grey.shade400),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: BorderSide(color: primaryColor, width: 2),
                ),
                labelText: global.language('select_date'),
                hintText: global.language('select_date_for_balance'),
                filled: true,
                fillColor: Colors.white,
                contentPadding: EdgeInsets.symmetric(
                  horizontal: 16,
                  vertical: 16,
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildFilterOptionsPanel() {
    return Card(
      margin: EdgeInsets.symmetric(horizontal: 12),
      elevation: 2,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Padding(
        padding: EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.filter_alt, color: primaryColor, size: 20),
                SizedBox(width: 10),
                Text(global.language("display_options"), style: titleStyle),
              ],
            ),
            SizedBox(height: 8),
            Text(global.language("customize_report_display"), style: normalStyle),
            SizedBox(height: 12),
            Container(
              decoration: BoxDecoration(
                color: Colors.grey.shade50,
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: Colors.grey.shade300),
              ),
              child: CheckboxListTile(
                title: Text(
                  global.language("show_only_with_balance"),
                  style: emphasisStyle,
                ),
                subtitle: Text(
                  "หากเลือก จะไม่แสดงสินค้าที่มียอดเป็นศูนย์ ทำให้รายงานกระชับขึ้น",
                  style: TextStyle(fontSize: 12, color: Colors.grey.shade700),
                ),
                value: reportCondition.showOnlyBalance,
                secondary: Icon(
                  reportCondition.showOnlyBalance
                      ? Icons.visibility
                      : Icons.visibility_off,
                  color: primaryColor,
                ),
                onChanged: (value) {
                  setState(() {
                    reportCondition.showOnlyBalance = value!;
                  });
                },
                controlAffinity: ListTileControlAffinity.leading,
                activeColor: primaryColor,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildProductSelectionPanel() {
    return Card(
      margin: EdgeInsets.symmetric(horizontal: 12),
      elevation: 2,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Padding(
        padding: EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.inventory, color: primaryColor, size: 20),
                SizedBox(width: 10),
                Text(global.language("report_condition_select_barcode"), style: titleStyle),
              ],
            ),
            SizedBox(height: 8),
            Text(
              global.language("select_products_for_report_or_show_all"),
              style: normalStyle,
            ),
            SizedBox(height: 16),
            Row(
              children: [
                Expanded(
                  child: ElevatedButton.icon(
                    onPressed: () async {
                      List<SelectedProductItemStruct>? results =
                          await showDialog<List<SelectedProductItemStruct>>(
                            context: context,
                            builder: (context) => SelectProductBarcodeWidget(
                              selectedItems:
                                  reportCondition.conditionItemCodeList,
                            ),
                          );
                      if (results != null) {
                        for (var item in results) {
                          if (!reportCondition.conditionItemCodeList.contains(
                            item,
                          )) {
                            reportCondition.conditionItemCodeList.add(item);
                          }
                        }
                      }
                      await reportCondition.reloadWareHouseCode();
                      setState(() {});
                    },
                    icon: Icon(Icons.add_shopping_cart, size: 18),
                    label: Text(global.language("report_condition_select_barcode")),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: primaryColor,
                      foregroundColor: Colors.white,
                      elevation: 2,
                      padding: EdgeInsets.symmetric(vertical: 12),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                    ),
                  ),
                ),
                SizedBox(width: 12),
                ElevatedButton.icon(
                  onPressed: reportCondition.conditionItemCodeList.isNotEmpty
                      ? () async {
                          setState(() {
                            reportCondition.conditionItemCodeList.clear();
                          });
                          await reportCondition.reloadWareHouseCode();
                        }
                      : null,
                  icon: Icon(Icons.refresh, size: 18),
                  label: Text(global.language("start_over")),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: accentColor,
                    foregroundColor: Colors.white,
                    elevation: 2,
                    padding: EdgeInsets.symmetric(vertical: 12),
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(8),
                    ),
                  ),
                ),
              ],
            ),
            SizedBox(height: 16),
            if (reportCondition.conditionItemCodeList.isEmpty)
              Container(
                padding: EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: Colors.grey.shade100,
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: Colors.grey.shade300),
                ),
                child: Row(
                  children: [
                    Icon(
                      Icons.info_outline,
                      color: Colors.grey.shade600,
                      size: 18,
                    ),
                    SizedBox(width: 12),
                    Expanded(
                      child: Text(
                        global.language("no_products_selected_show_all"),
                        style: TextStyle(
                          color: Colors.grey.shade700,
                          fontStyle: FontStyle.italic,
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            if (reportCondition.conditionItemCodeList.isNotEmpty)
              Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(
                        Icons.shopping_cart,
                        size: 16,
                        color: Colors.grey.shade700,
                      ),
                      SizedBox(width: 8),
                      Text(
                        "สินค้าที่เลือก (${reportCondition.conditionItemCodeList.length} รายการ):",
                        style: TextStyle(
                          fontWeight: FontWeight.w500,
                          fontSize: 14,
                          color: Colors.grey.shade800,
                        ),
                      ),
                    ],
                  ),
                  SizedBox(height: 12),
                  Container(
                    padding: EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: Colors.blue.shade50,
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(color: Colors.blue.shade200),
                    ),
                    child: Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: reportCondition.conditionItemCodeList.map((
                        item,
                      ) {
                        return Chip(
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(6),
                          ),
                          label: Text(item.itemCode),
                          backgroundColor: Colors.white,
                          side: BorderSide(color: Colors.blue.shade200),
                          labelStyle: TextStyle(
                            color: primaryColor,
                            fontWeight: FontWeight.w500,
                          ),
                          deleteIconColor: Colors.red.shade400,
                          avatar: Icon(
                            Icons.inventory_2_outlined,
                            size: 16,
                            color: primaryColor,
                          ),
                          elevation: 1,
                          shadowColor: Colors.grey.shade200,
                          onDeleted: () async {
                            reportCondition.conditionItemCodeList.remove(item);
                            await reportCondition.reloadWareHouseCode();
                            setState(() {});
                          },
                        );
                      }).toList(),
                    ),
                  ),
                ],
              ),
          ],
        ),
      ),
    );
  }

  Widget _buildWarehouseSelectionPanel() {
    return Card(
      margin: EdgeInsets.only(left: 12, right: 12, top: 12),
      elevation: 2,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Padding(
        padding: EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.store, color: Colors.amber.shade800, size: 20),
                SizedBox(width: 10),
                Text(
                  global.language("select_warehouse"),
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                    color: Colors.amber.shade900,
                  ),
                ),
              ],
            ),
            SizedBox(height: 8),
            Text(
              global.language("select_warehouse_for_report_or_show_all"),
              style: normalStyle,
            ),
            SizedBox(height: 16),
            reportCondition.selectWareHouseWidget(
              primaryColor: Colors.amber.shade600,
              secondaryColor: Colors.amber.shade100,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildLocationSelectionPanel() {
    return Card(
      margin: EdgeInsets.only(left: 12, right: 12, top: 12),
      elevation: 2,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Padding(
        padding: EdgeInsets.all(16),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.grid_view, color: Colors.green.shade800, size: 20),
                SizedBox(width: 10),
                Text(
                  "เลือกที่เก็บสินค้า",
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                    color: Colors.green.shade900,
                  ),
                ),
              ],
            ),
            SizedBox(height: 8),
            Text(
              "เลือกที่เก็บสินค้าที่ต้องการแสดงในรายงาน หากไม่เลือก ระบบจะแสดงที่เก็บทั้งหมด",
              style: normalStyle,
            ),
            SizedBox(height: 16),
            reportCondition.selectLocationWidget(
              primaryColor: Colors.green.shade600,
              secondaryColor: Colors.green.shade100,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildProcessButtonPanel() {
    return Card(
      margin: EdgeInsets.only(left: 12, right: 12, top: 20),
      elevation: 2,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Container(
        padding: EdgeInsets.all(16),
        decoration: BoxDecoration(
          gradient: LinearGradient(
            colors: [Colors.blue.shade50, Colors.blue.shade100],
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
          ),
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: Colors.blue.shade200),
        ),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(
              "เมื่อตั้งค่าเงื่อนไขเรียบร้อยแล้ว กดปุ่ม \"ประมวลผล\" เพื่อสร้างรายงาน",
              style: emphasisStyle,
            ),
            SizedBox(height: 16),
            Row(
              children: [
                Expanded(
                  child: ElevatedButton.icon(
                    icon: Icon(Icons.play_arrow),
                    label: Text(
                      global.language("process_report"),
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: primaryColor,
                      foregroundColor: Colors.white,
                      elevation: 3,
                      shadowColor: Colors.blue.shade300,
                      padding: EdgeInsets.symmetric(vertical: 16),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                    ),
                    onPressed: () async {
                      await _processReport();
                    },
                  ),
                ),
              ],
            ),
            SizedBox(height: 16),
            if (processState != global.ProcessState.idle)
              Container(
                padding: EdgeInsets.all(12),
                decoration: BoxDecoration(
                  color: () {
                    switch (processState) {
                      case global.ProcessState.processing:
                        return Colors.blue.shade50;
                      case global.ProcessState.success:
                        return Colors.green.shade50;
                      case global.ProcessState.error:
                        return Colors.red.shade50;
                      default:
                        return Colors.grey.shade50;
                    }
                  }(),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(
                    color: () {
                      switch (processState) {
                        case global.ProcessState.processing:
                          return Colors.blue.shade300;
                        case global.ProcessState.success:
                          return Colors.green.shade300;
                        case global.ProcessState.error:
                          return Colors.red.shade300;
                        default:
                          return Colors.grey.shade300;
                      }
                    }(),
                  ),
                ),
                child: processWidgetStatus,
              ),
          ],
        ),
      ),
    );
  }

  Future<void> _processReport() async {
    dataList.clear();
    setState(() {
      // Set state to processing first
      processState = global.ProcessState.processing;
      processWidgetStatus = Row(
        children: [
          SizedBox(
            width: 20,
            height: 20,
            child: CircularProgressIndicator(color: primaryColor),
          ),
          SizedBox(width: 12),
          Text(global.language("processing"), style: TextStyle(color: primaryColor)),
        ],
      );
    });

    try {
      // Add a small delay to ensure the spinner is shown
      await Future.delayed(Duration(milliseconds: 300));

      String itemCodeList = "";
      for (var item in reportCondition.conditionItemCodeList) {
        if (itemCodeList.isNotEmpty) {
          itemCodeList += ",";
        }
        itemCodeList += item.itemCode;
      }

      var payload = {
        "shop_id": global.getShopId(),
        "command_id":
            "stock_balance_by_product_and_warehouse_and_location_process",
        "condition": condition.toString(),
        "balance_only": reportCondition.showOnlyBalance.toString(),
        "item_code_list": itemCodeList.toString(),
        "warehouse_list": reportCondition.conditionWareHouseCodeList,
        "final_date":
            "${conditionFinalDate.year}-${conditionFinalDate.month.toString().padLeft(2, '0')}-${conditionFinalDate.day.toString().padLeft(2, '0')}",
      };

      var jsonPayload = jsonEncode(payload);
      var jsonResult = await global.reportServicePost(jsonPayload);

      if (jsonResult['code'] == 200) {
        pdfCreated = false;
        pdfDownloaded = false;
        setState(() {
          processSuccess = true;
          processState = global.ProcessState.success;
          processWidgetStatus = Row(
            children: [
              Icon(Icons.check_circle, color: successColor),
              SizedBox(width: 12),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      global.language("process_success"),
                      style: TextStyle(
                        color: successColor,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    Text(
                      "คลิกที่แท็บ \"แสดงผล\" หรือ \"PDF\" เพื่อดูผลลัพธ์",
                      style: TextStyle(
                        color: Colors.grey.shade700,
                        fontSize: 12,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          );
        });

        guid = jsonResult['guid'];
        tabController.index = 1; // Auto-switch to results tab
        await reloadData(0, dataOffset);
        setState(() {});
      } else {
        setState(() {
          processSuccess = false;
          processState = global.ProcessState.error;
          processWidgetStatus = Row(
            children: [
              Icon(Icons.error, color: errorColor),
              SizedBox(width: 12),
              Expanded(
                child: Text(
                  jsonResult['message'],
                  style: TextStyle(
                    color: errorColor,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
            ],
          );
        });
      }
    } catch (e) {
      setState(() {
        processSuccess = false;
        processState = global.ProcessState.error;
        processWidgetStatus = Row(
          children: [
            Icon(Icons.error, color: errorColor),
            SizedBox(width: 12),
            Expanded(
              child: Text(
                "${global.language("error_occurred")}: $e",
                style: TextStyle(
                  color: errorColor,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
          ],
        );
      });
    }
  }

  // Improved data display tab
  Widget _buildResultTab() {
    return Column(
      children: [
        Container(
          decoration: BoxDecoration(
            gradient: LinearGradient(
              begin: Alignment.topLeft,
              end: Alignment.bottomRight,
              colors: [primaryColor, primaryColor.withBlue(255)],
            ),
            boxShadow: [
              BoxShadow(
                color: Colors.blue.shade300,
                spreadRadius: 1,
                blurRadius: 4,
                offset: Offset(0, 2),
              ),
            ],
          ),
          padding: EdgeInsets.symmetric(vertical: 12, horizontal: 16),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Icon(Icons.inventory_2, color: Colors.white, size: 24),
                  SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      titles[condition - 1],
                      overflow: TextOverflow.ellipsis,
                      maxLines: 1,
                      style: TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                        color: Colors.white,
                      ),
                    ),
                  ),
                  Tooltip(
                    message: global.language("decrease_font_size"),
                    child: IconButton(
                      icon: Icon(Icons.remove, color: Colors.white),
                      onPressed: () {
                        if (resultFontScale > 0.5) {
                          setState(() {
                            resultFontScale = resultFontScale - 0.1;
                            renderData(dataList);
                          });
                        }
                      },
                    ),
                  ),
                  Tooltip(
                    message: global.language("increase_font_size"),
                    child: IconButton(
                      icon: Icon(Icons.add, color: Colors.white),
                      onPressed: () {
                        setState(() {
                          resultFontScale = resultFontScale + 0.1;
                          renderData(dataList);
                        });
                      },
                    ),
                  ),
                ],
              ),
              Divider(color: Colors.white.withValues(alpha: 0.2), height: 16),
              Text(
                "ข้อมูล ณ วันที่ ${conditionFinalDate.day}/${conditionFinalDate.month}/${conditionFinalDate.year}",
                style: TextStyle(color: Colors.white, fontSize: 13),
              ),
              SizedBox(height: 4),
              Text(
                global.language("report_zoom_tip"),
                style: TextStyle(
                  color: Colors.white.withValues(alpha: 0.9),
                  fontSize: 12,
                ),
              ),
            ],
          ),
        ),
        Expanded(
          child: dataList.isEmpty && processSuccess
              ? Center(
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      CircularProgressIndicator(color: primaryColor),
                      SizedBox(height: 16),
                      Text(global.language("loading"), style: emphasisStyle),
                    ],
                  ),
                )
              : resultScreenWidget,
        ),
      ],
    );
  }

  // Enhanced PDF tab
  Widget _buildPdfTab() {
    return Container(
      color: backgroundColor,
      padding: EdgeInsets.all(12),
      child: Column(
        children: [
          Card(
            elevation: 2,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12),
            ),
            child: Padding(
              padding: EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Icon(Icons.info_outline, color: primaryColor),
                      SizedBox(width: 12),
                      Text(global.language("pdf_usage"), style: titleStyle),
                    ],
                  ),
                  SizedBox(height: 8),
                  Text(
                    global.language("report_pdf_tab_desc"),
                    style: normalStyle,
                  ),
                ],
              ),
            ),
          ),
          SizedBox(height: 12),
          Card(
            elevation: 2,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12),
            ),
            child: Padding(
              padding: EdgeInsets.all(16),
              child: Column(
                children: [
                  TextField(
                    controller: pdfSearchController,
                    decoration: InputDecoration(
                      labelText: global.language('search_text'),
                      hintText: global.language('type_search_text_in_document'),
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                      prefixIcon: Icon(Icons.search, color: primaryColor),
                      enabledBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                        borderSide: BorderSide(color: Colors.grey.shade400),
                      ),
                      focusedBorder: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(8),
                        borderSide: BorderSide(color: primaryColor, width: 2),
                      ),
                      filled: true,
                      fillColor: Colors.white,
                      suffixIcon: Row(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          IconButton(
                            icon: Icon(
                              Icons.clear,
                              color: Colors.grey.shade600,
                            ),
                            tooltip: global.language("clear_search"),
                            onPressed: () {
                              pdfSearchController.clear();
                              pdfViewerController.searchText("");
                            },
                          ),
                          IconButton(
                            icon: Icon(Icons.search, color: primaryColor),
                            tooltip: global.language("search_text"),
                            onPressed: () {
                              pdfViewerController.searchText(
                                pdfSearchController.text,
                              );
                              setState(() {});
                            },
                          ),
                        ],
                      ),
                    ),
                    onSubmitted: (value) {
                      pdfViewerController.searchText(value);
                    },
                  ),
                  SizedBox(height: 16),
                  Row(
                    children: [
                      Expanded(
                        child: ElevatedButton.icon(
                          icon: Icon(Icons.save),
                          label: Text(global.language("save_as_pdf")),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: primaryColor,
                            foregroundColor: Colors.white,
                            padding: EdgeInsets.symmetric(vertical: 14),
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(8),
                            ),
                          ),
                          onPressed: pdfDownloaded
                              ? () async {
                                  final dateTimeNow = DateTime.now();
                                  final formattedDate =
                                      "${dateTimeNow.year}-${dateTimeNow.month}-${dateTimeNow.day}-${dateTimeNow.hour}-${dateTimeNow.minute}-${dateTimeNow.second}";
                                  final pdfFileName =
                                      "${titles[condition - 1].replaceAll(" ", "_")}_$formattedDate.pdf";
                                  savePdf(pdfData, pdfFileName, context);
                                }
                              : null,
                        ),
                      ),
                      SizedBox(width: 12),
                      Expanded(
                        child: ElevatedButton.icon(
                          icon: Icon(Icons.print),
                          label: Text(global.language("print_report")),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: accentColor,
                            foregroundColor: Colors.white,
                            padding: EdgeInsets.symmetric(vertical: 14),
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(8),
                            ),
                          ),
                          onPressed: pdfDownloaded
                              ? () async {
                                  await printing.Printing.layoutPdf(
                                    usePrinterSettings: true,
                                    dynamicLayout: true,
                                    onLayout:
                                        (pdf.PdfPageFormat format) async =>
                                            pdfData,
                                  );
                                }
                              : null,
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ),
          SizedBox(height: 12),
          Expanded(
            child: Card(
              elevation: 2,
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
              ),
              child: Container(
                decoration: BoxDecoration(
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: Colors.grey.shade300),
                ),
                child: ClipRRect(
                  borderRadius: BorderRadius.circular(12),
                  child: pdfDownloaded
                      ? pdfViewWidget
                      : Center(
                          child: Column(
                            mainAxisSize: MainAxisSize.min,
                            children: [
                              CircularProgressIndicator(color: primaryColor),
                              SizedBox(height: 16),
                              Text(
                                global.language("loading_pdf_please_wait"),
                                style: emphasisStyle,
                              ),
                              if (!processSuccess)
                                Padding(
                                  padding: EdgeInsets.only(top: 16),
                                  child: Text(
                                    global.language("report_not_processed_hint"),
                                    style: TextStyle(
                                      color: Colors.grey.shade700,
                                    ),
                                    textAlign: TextAlign.center,
                                  ),
                                ),
                            ],
                          ),
                        ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return DefaultTabController(
      length: 3,
      child: Scaffold(
        appBar: AppBar(
          backgroundColor: primaryColor,
          automaticallyImplyLeading: false,
          elevation: 4,
          leading: IconButton(
            icon: const Icon(Icons.close, color: Colors.white),
            onPressed: () {
              if (processSuccess == true) {
                showDialog(
                  context: context,
                  builder: (context) {
                    return AlertDialog(
                      title: Row(
                        children: [
                          Icon(Icons.warning, color: accentColor),
                          SizedBox(width: 12),
                          Text(global.language("confirm_exit_report")),
                        ],
                      ),
                      content: Text(global.language("confirm_exit_report_question")),
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                      actions: [
                        TextButton(
                          onPressed: () {
                            global.gotoMainMenu(context);
                          },
                          child: Text(
                            global.language("cancel"),
                            style: TextStyle(color: Colors.grey.shade700),
                          ),
                        ),
                        ElevatedButton(
                          onPressed: () {
                            Navigator.of(context).pop(true);
                          },
                          style: ElevatedButton.styleFrom(
                            backgroundColor: primaryColor,
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(8),
                            ),
                          ),
                          child: Text(global.language("ok")),
                        ),
                      ],
                    );
                  },
                ).then((value) {
                  if (value == true) {
                    global.gotoMainMenu(context);
                  }
                });
              } else {
                global.gotoMainMenu(context);
              }
            },
            tooltip: global.language("close_report_page"),
          ),
          actions: [
            IconButton(
              icon: Icon(Icons.help_outline, color: Colors.white),
              onPressed: () {
                setState(() {
                  showHelp = !showHelp;
                });
              },
              tooltip: global.language("show_hide_usage_guide"),
            ),
          ],
          title: TabBar(
            controller: tabController,
            indicatorColor: Colors.white,
            indicatorWeight: 3,
            labelColor: Colors.white,
            unselectedLabelColor: Colors.white.withValues(alpha: 0.7),
            labelStyle: TextStyle(fontWeight: FontWeight.bold),
            tabs: [
              Tab(
                child: Wrap(
                  alignment: WrapAlignment.center,
                  crossAxisAlignment: WrapCrossAlignment.center,
                  spacing: 8,
                  children: [Icon(Icons.settings), Text(global.language("condition"))],
                ),
              ),
              Tab(
                child: Wrap(
                  alignment: WrapAlignment.center,
                  crossAxisAlignment: WrapCrossAlignment.center,
                  spacing: 8,
                  children: [Icon(Icons.table_chart), Text(global.language("display"))],
                ),
              ),
              Tab(
                child: Wrap(
                  alignment: WrapAlignment.center,
                  crossAxisAlignment: WrapCrossAlignment.center,
                  spacing: 8,
                  children: [Icon(Icons.picture_as_pdf), Text("PDF")],
                ),
              ),
            ],
          ),
        ),
        body: TabBarView(
          controller: tabController,
          children: [
            _buildConditionSelect(),
            _buildResultTab(),
            _buildPdfTab(),
          ],
        ),
        bottomNavigationBar: processSuccess
            ? Container(
                color: Colors.blue.shade50,
                padding: EdgeInsets.symmetric(vertical: 10, horizontal: 16),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Icon(Icons.info_outline, color: primaryColor, size: 18),
                    SizedBox(width: 8),
                    Text(
                      "เลือกแท็บ \"แสดงผล\" สำหรับดูข้อมูล หรือ \"PDF\" สำหรับพิมพ์และบันทึกรายงาน",
                      style: TextStyle(
                        color: primaryColor,
                        fontStyle: FontStyle.italic,
                      ),
                    ),
                  ],
                ),
              )
            : null,
      ),
    );
  }
}
