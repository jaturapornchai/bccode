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

class ReportStockBalanceWhBarcode extends StatefulWidget {
  const ReportStockBalanceWhBarcode({super.key});

  @override
  _ReportStockBalanceWhBarcodeState createState() =>
      _ReportStockBalanceWhBarcodeState();
}

class _ReportStockBalanceWhBarcodeState
    extends State<ReportStockBalanceWhBarcode>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  bool processSuccess = false;
  bool pdfCreated = false;
  bool pdfDownloaded = false;
  String pdfPath = "";
  String guid = "";
  Widget processWidgetStatus = Container();
  global.ProcessState processState = global.ProcessState.idle;
  Widget resultScreenWidget = Container();
  Widget pdfViewWidget = Container();
  PdfViewerController pdfViewerController = PdfViewerController();
  TextEditingController pdfSearchController = TextEditingController();
  late Uint8List pdfData;
  DateTime conditionFinalDate = DateTime.now();
  late TabController tabController;
  double resultFontScale = 1.0;
  List<ReportStockBalanceWhBarcodeModel> dataList = [];
  late ReportStockConditionClass reportCondition;
  ScrollController scrollController = ScrollController();
  int dataOffset = 200;
  bool showHelp = false;
  bool isLoadingMore = false;

  // สีและสไตล์ที่ใช้ในหน้ารายงาน
  final Color primaryColor = global.theme.infoHighlightTextColor;
  Color get secondaryColor => global.theme.columnHeaderColor;
  final Color accentColor = Colors.amber.shade600;
  Color get headerColor => global.theme.columnHeaderColor;
  Color get footerColor => global.theme.surfaceColor;
  final Color rowEvenColor = global.theme.backgroundColor;
  final Color rowOddColor = global.theme.cardColor;
  Color get warehouseColor => global.theme.surfaceColor;
  Color get locationColor => global.theme.surfaceColor;
  Color get helpColor => global.theme.surfaceColor;
  final Color successColor = global.theme.positiveHighlightTextColor;
  final Color errorColor = global.theme.negativeHighlightTextColor;
  final Color backgroundColor = global.theme.surfaceColor;

  // รูปแบบตัวอักษรแบบต่างๆ
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

    normalStyle = TextStyle(fontSize: 13, color: global.theme.textColor);

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
    scrollController.dispose();
    super.dispose();
  }

  Future<void> reloadPdf() async {
    try {
      if (pdfCreated == false) {
        var payload = {
          "shop_id": global.getShopId(),
          "command_id": "stock_balance_by_warehouse_and_product_create_pdf",
          "guid": guid,
          "condition": "1",
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
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error("Error loading PDF: $e");
      }
      setState(() {
        pdfViewWidget = Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(Icons.error_outline, color: errorColor, size: 48),
              SizedBox(height: 16),
              Text(
                global.language("cannot_load_pdf"),
                style: TextStyle(
                  color: errorColor,
                  fontWeight: FontWeight.bold,
                ),
              ),
              SizedBox(height: 8),
              Text(
                global.language("please_try_process_report_again"),
                style: TextStyle(color: global.theme.textColor),
              ),
            ],
          ),
        );
      });
    }
  }

  Widget renderData(List<ReportStockBalanceWhBarcodeModel> dataList) {
    double fontSize = 12.0 * resultFontScale;
    var header = Container(
      decoration: BoxDecoration(
        color: headerColor,
        border: Border.all(color: global.theme.infoHighlightTextColor),
        borderRadius: BorderRadius.only(
          topLeft: Radius.circular(8),
          topRight: Radius.circular(8),
        ),
      ),
      padding: EdgeInsets.all(10),
      child: Column(
        children: [
          Row(
            children: [
              Expanded(
                flex: 1,
                child: Padding(
                  padding: EdgeInsets.all(2),
                  child: Text(
                    global.language("warehouse_code"),
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
        ],
      ),
    );

    List<Widget> data = [];
    bool isEvenRow = true;
    TextStyle itemStyle = TextStyle(
      fontSize: fontSize,
      fontWeight: FontWeight.normal,
    );

    for (var item in dataList) {
      data.add(
        Container(
          decoration: BoxDecoration(
            color: isEvenRow ? rowEvenColor : rowOddColor,
            border: Border(
              bottom: BorderSide(color: global.theme.dividerBorderColor, width: 1),
            ),
          ),
          child: Row(
            children: [
              Expanded(
                flex: 1,
                child: Padding(
                  padding: EdgeInsets.all(4),
                  child: Text(item.whCode, style: itemStyle),
                ),
              ),
              Expanded(
                flex: 1,
                child: Padding(
                  padding: EdgeInsets.all(4),
                  child: Text(item.itemCode, style: itemStyle),
                ),
              ),
              Expanded(
                flex: 2,
                child: Padding(
                  padding: EdgeInsets.all(4),
                  child: Text(item.itemName, style: itemStyle),
                ),
              ),
              Expanded(
                flex: 1,
                child: Padding(
                  padding: EdgeInsets.all(4),
                  child: Text(item.barcodeList, style: itemStyle),
                ),
              ),
              Expanded(
                flex: 1,
                child: Padding(
                  padding: EdgeInsets.all(4),
                  child: Text(
                    "${item.unitCode ?? ""}/${item.unitName ?? ""}",
                    style: itemStyle,
                  ),
                ),
              ),
              Expanded(
                flex: 1,
                child: Padding(
                  padding: EdgeInsets.all(4),
                  child: Text(
                    global.formatNumber(item.balanceQty ?? 0),
                    style: itemStyle,
                    textAlign: TextAlign.right,
                  ),
                ),
              ),
              Expanded(
                flex: 2,
                child: Padding(
                  padding: EdgeInsets.all(4),
                  child: Text(item.balanceWord ?? "", style: itemStyle),
                ),
              ),
            ],
          ),
        ),
      );

      isEvenRow = !isEvenRow;
    }

    resultScreenWidget = Container(
      margin: EdgeInsets.all(8),
      child: Card(
        elevation: 4,
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
        child: Column(
          children: [
            header,
            Expanded(
              child: dataList.isEmpty
                  ? Center(
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        children: [
                          Icon(
                            Icons.inbox,
                            size: 48,
                            color: global.theme.iconSecondaryColor,
                          ),
                          SizedBox(height: 16),
                          Text(
                            global.language("no_data_for_selected_criteria"),
                            style: TextStyle(
                              color: global.theme.textSecondaryColor,
                              fontSize: 15,
                            ),
                          ),
                          SizedBox(height: 8),
                          Text(
                            "กรุณาเปลี่ยนเงื่อนไขการค้นหาและลองใหม่อีกครั้ง",
                            style: TextStyle(
                              color: global.theme.textSecondaryColor,
                              fontSize: 13,
                            ),
                          ),
                        ],
                      ),
                    )
                  : SingleChildScrollView(
                      physics: const AlwaysScrollableScrollPhysics(),
                      controller: scrollController,
                      scrollDirection: Axis.vertical,
                      child: Column(
                        children: [
                          Container(
                            decoration: BoxDecoration(
                              color: global.theme.cardColor,
                              border: Border.all(color: global.theme.dividerBorderColor),
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
    );

    return resultScreenWidget;
  }

  Future<void> reloadData(int offset, int limit) async {
    try {
      // ใช้ ReportRepository.loadReportData แทน - generic function
      await ReportRepository.loadReportData<ReportStockBalanceWhBarcodeModel>(
        guid: guid,
        offset: offset,
        limit: limit,
        dataList: dataList,
        fromJson: ReportStockBalanceWhBarcodeModel.fromJson,
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
                    icon: Icon(Icons.close, color: global.theme.iconColor),
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
                global.language("select_date"),
                "เลือกวันที่ที่ต้องการแสดงยอดคงเหลือ (ณ สิ้นวันของวันที่เลือก)",
                Icons.calendar_today,
              ),
              _buildHelpStep(
                "2",
                global.language("display_options"),
                global.language("show_only_with_balance_or_all"),
                Icons.filter_alt,
              ),
              _buildHelpStep(
                "3",
                global.language("report_condition_select_barcode"),
                global.language("select_products_or_show_all"),
                Icons.inventory,
              ),
              _buildHelpStep(
                "4",
                global.language("select_warehouse"),
                global.language("select_warehouse_or_show_all"),
                Icons.store,
              ),
              _buildHelpStep(
                "5",
                global.language("process"),
                "กดปุ่ม \"ประมวลผล\" เพื่อสร้างรายงาน",
                Icons.search,
              ),
              _buildHelpStep(
                "6",
                global.language("view_results"),
                "หลังจากประมวลผลเสร็จ ดูผลลัพธ์ที่แท็บ \"แสดงผล\" หรือดูในรูปแบบ PDF ที่แท็บ \"PDF\"",
                Icons.visibility,
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
                    backgroundColor: global.theme.columnHeaderColor,
                    foregroundColor: global.theme.columnHeaderTextColor,
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
                  color: global.theme.textColor.withValues(alpha: 0.1),
                  blurRadius: 3,
                  offset: Offset(0, 1),
                ),
              ],
            ),
            child: Center(
              child: Text(
                step,
                style: TextStyle(
                  color: global.theme.onPrimaryColor,
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
                  style: TextStyle(fontSize: 14, color: global.theme.textColor),
                ),
              ],
            ),
          ),
        ],
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
            _buildDateSelectionPanel(),
            SizedBox(height: 12),
            _buildFilterOptionsPanel(),
            SizedBox(height: 12),
            _buildProductSelectionPanel(),
            SizedBox(height: 12),
            _buildWarehouseSelectionPanel(),
            SizedBox(height: 12),
            _buildProcessButtonPanel(),
            SizedBox(height: 20),
          ],
        ),
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
                Icon(Icons.inventory_2, color: global.theme.onPrimaryColor, size: 28),
                SizedBox(width: 12),
                Text(
                  global.language("inventory_report"),
                  style: TextStyle(
                    fontSize: 22,
                    fontWeight: FontWeight.bold,
                    color: global.theme.cardColor,
                  ),
                ),
              ],
            ),
            SizedBox(height: 8),
            Divider(color: global.theme.cardColor.withValues(alpha: 0.3), height: 24),
            Text(
              "ตามคลังสินค้า -> สินค้า",
              style: TextStyle(
                fontSize: 16,
                fontWeight: FontWeight.w500,
                color: global.theme.cardColor.withValues(alpha: 0.95),
              ),
            ),
            SizedBox(height: 8),
            Text(
              "รายงานนี้แสดงยอดคงเหลือของสินค้าแยกตามคลังสินค้า โดยสามารถเลือกวันที่ต้องการดูข้อมูล และกรองข้อมูลตามสินค้าและคลังสินค้าที่ต้องการได้",
              style: TextStyle(
                fontSize: 14,
                color: global.theme.cardColor.withValues(alpha: 0.85),
              ),
            ),
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
                    color: global.theme.textSecondaryColor,
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
                  borderSide: BorderSide(color: global.theme.iconSecondaryColor),
                ),
                focusedBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: BorderSide(color: primaryColor, width: 2),
                ),
                labelText: global.language('select_date'),
                hintText: global.language('select_date_for_balance'),
                filled: true,
                fillColor: global.theme.formFillColor,
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
                color: global.theme.backgroundColor,
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: global.theme.dividerBorderColor),
              ),
              child: CheckboxListTile(
                title: Text(
                  global.language("show_only_with_balance"),
                  style: emphasisStyle,
                ),
                subtitle: Text(
                  "หากเลือก จะไม่แสดงสินค้าที่มียอดเป็น 0 ทำให้รายงานกระชับขึ้น",
                  style: TextStyle(fontSize: 12, color: global.theme.iconColor),
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
                      List<SelectedProductItemStruct>? result =
                          await showDialog<List<SelectedProductItemStruct>>(
                            context: context,
                            builder: (context) => SelectProductBarcodeWidget(
                              selectedItems:
                                  reportCondition.conditionItemCodeList,
                            ),
                          );
                      if (result != null) {
                        for (var item in result) {
                          if (!reportCondition.conditionItemCodeList.contains(
                            item,
                          )) {
                            reportCondition.conditionItemCodeList.add(item);
                          }
                        }
                        await reportCondition.reloadWareHouseCode();
                        setState(() {});
                      }
                    },
                    icon: Icon(Icons.add_shopping_cart, size: 18),
                    label: Text(global.language("report_condition_select_barcode")),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: primaryColor,
                      foregroundColor: global.theme.onPrimaryColor,
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
                    foregroundColor: global.theme.onPrimaryColor,
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
                  color: global.theme.surfaceColor,
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: global.theme.dividerBorderColor),
                ),
                child: Row(
                  children: [
                    Icon(
                      Icons.info_outline,
                      color: global.theme.textSecondaryColor,
                      size: 18,
                    ),
                    SizedBox(width: 12),
                    Expanded(
                      child: Text(
                        global.language("no_products_selected_show_all"),
                        style: TextStyle(
                          color: global.theme.iconColor,
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
                        color: global.theme.iconColor,
                      ),
                      SizedBox(width: 8),
                      Text(
                        "สินค้าที่เลือก (${reportCondition.conditionItemCodeList.length} รายการ):",
                        style: TextStyle(
                          fontWeight: FontWeight.w500,
                          fontSize: 14,
                          color: global.theme.textColor,
                        ),
                      ),
                    ],
                  ),
                  SizedBox(height: 12),
                  Container(
                    padding: EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: helpColor,
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(color: global.theme.infoHighlightColor),
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
                          backgroundColor: global.theme.cardColor,
                          side: BorderSide(color: global.theme.infoHighlightColor),
                          labelStyle: TextStyle(
                            color: primaryColor,
                            fontWeight: FontWeight.w500,
                          ),
                          deleteIconColor: global.theme.negativeHighlightTextColor,
                          avatar: Icon(
                            Icons.inventory_2_outlined,
                            size: 16,
                            color: primaryColor,
                          ),
                          elevation: 1,
                          shadowColor: global.theme.dividerBorderColor,
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
                Icon(Icons.store, color: primaryColor, size: 20),
                SizedBox(width: 10),
                Text(global.language("select_warehouse"), style: titleStyle),
              ],
            ),
            SizedBox(height: 8),
            Text(
              global.language("select_warehouse_for_report_or_show_all"),
              style: normalStyle,
            ),
            SizedBox(height: 16),
            reportCondition.selectWareHouseWidget(
              primaryColor: primaryColor,
              secondaryColor: secondaryColor,
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildProcessButtonPanel() {
    return Card(
      margin: EdgeInsets.symmetric(horizontal: 12),
      elevation: 2,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
      child: Container(
        padding: EdgeInsets.all(16),
        decoration: BoxDecoration(
          gradient: LinearGradient(
            colors: [helpColor, secondaryColor],
            begin: Alignment.topCenter,
            end: Alignment.bottomCenter,
          ),
          borderRadius: BorderRadius.circular(12),
          border: Border.all(color: global.theme.infoHighlightColor),
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
                      foregroundColor: global.theme.onPrimaryColor,
                      elevation: 3,
                      shadowColor: global.theme.infoHighlightTextColor,
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
                        return helpColor;
                      case global.ProcessState.success:
                        return global.theme.positiveHighlightColor;
                      case global.ProcessState.error:
                        return global.theme.negativeHighlightColor;
                      default:
                        return global.theme.backgroundColor;
                    }
                  }(),
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(
                    color: () {
                      switch (processState) {
                        case global.ProcessState.processing:
                          return global.theme.infoHighlightTextColor;
                        case global.ProcessState.success:
                          return global.theme.positiveHighlightTextColor;
                        case global.ProcessState.error:
                          return global.theme.negativeHighlightTextColor;
                        default:
                          return global.theme.dividerBorderColor;
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
        "command_id": "stock_balance_by_warehouse_and_product_process",
        "condition": "1",
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
                        color: global.theme.iconColor,
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
        tabController.index = 1;
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
                color: global.theme.infoHighlightTextColor,
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
                  Icon(Icons.inventory_2, color: global.theme.onPrimaryColor, size: 24),
                  SizedBox(width: 12),
                  Expanded(
                    child: Text(
                      "รายงานยอดคงเหลือสินค้า ตามคลังสินค้า -> สินค้า",
                      overflow: TextOverflow.ellipsis,
                      maxLines: 1,
                      style: TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                        color: global.theme.cardColor,
                      ),
                    ),
                  ),
                  Tooltip(
                    message: global.language("decrease_font_size"),
                    child: IconButton(
                      icon: Icon(Icons.remove, color: global.theme.cardColor),
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
                      icon: Icon(Icons.add, color: global.theme.cardColor),
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
              Divider(color: global.theme.cardColor.withValues(alpha: 0.2), height: 16),
              Text(
                "ข้อมูล ณ วันที่ ${conditionFinalDate.day}/${conditionFinalDate.month}/${conditionFinalDate.year}",
                style: TextStyle(color: global.theme.onPrimaryColor, fontSize: 13),
              ),
              SizedBox(height: 4),
              Text(
                global.language("report_zoom_tip"),
                style: TextStyle(
                  color: global.theme.cardColor.withValues(alpha: 0.9),
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
                  Row(
                    children: [
                      Expanded(
                        child: TextField(
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
                              borderSide: BorderSide(
                                color: global.theme.iconSecondaryColor,
                              ),
                            ),
                            focusedBorder: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(8),
                              borderSide: BorderSide(
                                color: primaryColor,
                                width: 2,
                              ),
                            ),
                            filled: true,
                            fillColor: global.theme.formFillColor,
                            suffixIcon: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                IconButton(
                                  icon: Icon(
                                    Icons.clear,
                                    color: global.theme.textSecondaryColor,
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
                      ),
                    ],
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
                            foregroundColor: global.theme.onPrimaryColor,
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
                                      "รายงานสินค้าคงเหลือ_$formattedDate.pdf";
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
                            foregroundColor: global.theme.onPrimaryColor,
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
                  color: global.theme.cardColor,
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: global.theme.dividerBorderColor),
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
                                      color: global.theme.iconColor,
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
            icon: Icon(Icons.close, color: global.theme.cardColor),
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
                            style: TextStyle(color: global.theme.iconColor),
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
                ).then((onValue) {
                  if (onValue != null && onValue) {
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
              icon: Icon(Icons.help_outline, color: global.theme.cardColor),
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
            indicatorColor: global.theme.onPrimaryColor,
            indicatorWeight: 3,
            labelColor: global.theme.onPrimaryColor,
            unselectedLabelColor: global.theme.onPrimaryColor.withValues(alpha: 0.7),
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
                color: helpColor,
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
