import 'package:smlaicloud/bloc/table/table_bloc.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/table_model.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;

import 'package:qr_flutter/qr_flutter.dart';
import 'dart:convert';

import 'package:printing/printing.dart';
import 'package:pdf/pdf.dart';
import 'package:pdf/widgets.dart' as pw;

class QrcodeOrderScreen extends StatefulWidget {
  final int groupnumber;
  const QrcodeOrderScreen({super.key, required this.groupnumber});

  @override
  State<QrcodeOrderScreen> createState() => QrcodeOrderScreenState();
}

class QrcodeOrderScreenState extends State<QrcodeOrderScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  TextEditingController searchController = TextEditingController();
  List<ProductBarcodeModel> listData = [];
  bool loadingData = false;
  List<TableModel> tableList = [];
  TableModel? table;
  List<GlobalKey> globalKey = [];
  @override
  void initState() {
    loadDataList("");

    super.initState();
  }

  Future<void> generateAndShowQrcodeDialog(
    String qrData,
    String tableNumber,
  ) async {
    String shopName =
        "${global.language('shop_name')} : ${global.getShopName()}";
    String table = " | ${global.language('table_number')} : $tableNumber";

    if (tableNumber == "0") {
      table = "";
    }

    final image = await QrPainter(
      data: qrData,
      version: QrVersions.auto,
      gapless: false,
      emptyColor: global.theme.onPrimaryColor,
    ).toImageData(500.0);

    // ignore: use_build_context_synchronously
    showDialog(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          title: Text("$shopName$table"),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Image.memory(image!.buffer.asUint8List()),
              const SizedBox(height: 20),
              ElevatedButton.icon(
                onPressed: () async {
                  await Printing.sharePdf(
                    bytes: image.buffer.asUint8List(),
                    filename:
                        'qr_code_order_${global.getShopName()}_$tableNumber.png',
                  );
                  // ignore: use_build_context_synchronously
                  Navigator.of(context).pop();
                },
                icon: Icon(Icons.download),
                label: Text(global.language('download')),
              ),
            ],
          ),
        );
      },
    );
  }

  /// generateAndPrintPdf function is used to generate pdf and print pdf file using printing package in flutter
  Future<void> generateAndPrintPdf(List<TableModel> tableList) async {
    final pdf = pw.Document();
    PdfPageFormat pageFormat = PdfPageFormat.a4.landscape;
    pw.PageTheme buildTheme(
      PdfPageFormat pageFormat,
      pw.Font base,
      pw.Font bold,
      pw.Font italic,
    ) {
      return pw.PageTheme(
        pageFormat: pageFormat,
        theme: pw.ThemeData.withFont(base: base, bold: bold, italic: italic),
      );
    }

    List<pw.Widget> pageContent = [];

    for (int i = 0; i < tableList.length; i++) {
      if (i % 12 == 0) {
        pageContent.add(
          pw.SizedBox(
            height: 481,
            child: pw.GridView(
              crossAxisCount: 4,
              childAspectRatio: 1,
              children: List.generate(12, (index) {
                int tableIndex = i + index;
                if (tableIndex >= tableList.length) {
                  return pw.Container();
                }
                return pw.Container(
                  decoration: pw.BoxDecoration(
                    border: pw.Border.all(color: PdfColors.grey),
                  ),
                  child: pw.Column(
                    children: [
                      pw.Expanded(
                        child: pw.Column(
                          children: [
                            pw.SizedBox(height: 10),
                            pw.Text(
                              global.language("qrcode_order_food"),
                              style: pw.TextStyle(
                                fontSize: 16,
                                fontWeight: pw.FontWeight.bold,
                              ),
                            ),
                            pw.Expanded(
                              child: pw.BarcodeWidget(
                                data:
                                    "https://dedefoodorder.web.app/?q=${tableList[tableIndex].orderendcode}",
                                barcode: pw.Barcode.qrCode(),
                                drawText: false,
                              ),
                            ),
                            pw.Padding(
                              padding: pw.EdgeInsets.all(8.0),
                              child: pw.Column(
                                children: [
                                  pw.Text(
                                    "${global.language('shop_name')} : ${global.getShopName()}",
                                    style: pw.TextStyle(
                                      fontSize: 14,
                                      fontWeight: pw.FontWeight.bold,
                                    ),
                                  ),
                                  pw.Text(
                                    "${global.language('table_name')} : ${tableList[tableIndex].number}",
                                    style: pw.TextStyle(
                                      fontSize: 14,
                                      fontWeight: pw.FontWeight.bold,
                                      color:
                                          (tableList[tableIndex].number != "0")
                                          ? PdfColors.black
                                          : PdfColors.white,
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                );
              }),
            ),
          ),
        );
      }
    }

    for (int i = 0; i < pageContent.length; i++) {
      pdf.addPage(
        pw.MultiPage(
          pageTheme: buildTheme(
            pageFormat,
            await PdfGoogleFonts.iBMPlexSansThaiMedium(),
            await PdfGoogleFonts.iBMPlexSansThaiBold(),
            await PdfGoogleFonts.iBMPlexSansThaiBold(),
          ),
          build: (pw.Context context) {
            return [pw.Center(child: pageContent[i])];
          },
        ),
      );
    }

    await Printing.layoutPdf(
      onLayout: (PdfPageFormat format) async => pdf.save(),
    );
  }

  Widget objectBox(TableModel data, int index) {
    return RepaintBoundary(
      key: globalKey[index],
      child: MouseRegion(
        cursor: SystemMouseCursors.click,
        child: GestureDetector(
          onTap: () async {
            await generateAndShowQrcodeDialog(
              'https://dedefoodorder.web.app/?q=${data.orderendcode}',
              data.number,
            );
          },
          child: Container(
            width: 150,
            height: 150,
            decoration: BoxDecoration(
              color: global.theme.cardColor,
              border: Border.all(color: global.theme.dividerBorderColor),
            ),
            child: Column(
              children: [
                Expanded(
                  child: Column(
                    children: [
                      Expanded(
                        child: QrImageView(
                          data:
                              "https://dedefoodorder.web.app/?q=${data.orderendcode}",
                          size: null,
                          version: QrVersions.auto,
                          gapless: false,
                        ),
                      ),
                      Padding(
                        padding: EdgeInsets.all(8.0),
                        child: Column(
                          children: [
                            Text(
                              maxLines: 2,
                              overflow: TextOverflow.ellipsis,
                              "${global.language('shop_name')} : ${global.getShopName()}",
                              style: TextStyle(
                                fontSize: 14,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                            Text(
                              maxLines: 2,
                              overflow: TextOverflow.ellipsis,
                              "${global.language('table_name')} : ${data.number}",
                              style: TextStyle(
                                fontSize: 14,
                                fontWeight: FontWeight.bold,
                                color: (data.number != "0")
                                    ? global.theme.textColor
                                    : global.theme.onPrimaryColor,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  void loadDataList(String search) {
    setState(() {
      loadingData = true;
    });
    context.read<TableBloc>().add(
      TableLoadList(
        offset: (listData.isEmpty) ? 0 : listData.length,
        limit: global.loadDataPerPage,
        search: search,
        groupNumber: widget.groupnumber,
      ),
    );
  }

  Widget gridView() {
    return LayoutBuilder(
      builder: (context, constraints) {
        return Container(
          padding: EdgeInsets.all(5),
          child: Column(
            children: [
              Expanded(
                child: (tableList.isEmpty)
                    ? Center(child: Text(global.language('no_setting_table')))
                    : GridView.builder(
                        gridDelegate:
                            const SliverGridDelegateWithFixedCrossAxisCount(
                              crossAxisCount: 5,
                              crossAxisSpacing: 5,
                              mainAxisSpacing: 5,
                              childAspectRatio: 1,
                            ),
                        itemCount: tableList.length,
                        itemBuilder: (context, index) {
                          return objectBox(tableList[index], index);
                        },
                      ),
              ),
            ],
          ),
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        return MultiBlocListener(
          listeners: [
            BlocListener<TableBloc, TableState>(
              listener: (context, state) {
                // Load
                if (state is TableLoadSuccess) {
                  if (state.tables.isNotEmpty) {
                    tableList = [];
                    tableList.add(
                      TableModel(
                        guidfixed: "x",
                        number: "0",
                        names: [],
                        zone: "x",
                        xorder: 0,
                        orderendcode: base64.encode(
                          utf8.encode('{"shopid": "${global.getShopId()}"}'),
                        ),
                      ),
                    );
                    globalKey.add(GlobalKey());
                    for (int i = 0; i < state.tables.length; i++) {
                      tableList.add(
                        TableModel(
                          guidfixed: state.tables[i].guidfixed,
                          number: state.tables[i].number,
                          names: state.tables[i].names,
                          zone: state.tables[i].zone,
                          xorder: state.tables[i].xorder,
                          orderendcode: base64.encode(
                            utf8.encode(
                              '{"shopid": "${global.getShopId()}", "table": "${state.tables[i].number}"}',
                            ),
                          ),
                        ),
                      );
                      globalKey.add(GlobalKey());
                    }

                    /// sort by xorder
                    tableList.sort((a, b) => (a.xorder!).compareTo(b.xorder!));
                  }

                  setState(() {});
                }
              },
            ),
          ],
          child: Scaffold(
            resizeToAvoidBottomInset: true,
            appBar: AppBar(
              backgroundColor: global.theme.appBarColor,
              automaticallyImplyLeading: false,
              title: Text(global.language('qr_code_order')),
              leading: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                icon: Icon(Icons.arrow_back),
                onPressed: () {
                  global.gotoMainMenu(context);
                },
              ),
              actions: [
                /// print qr code
                Padding(
                  padding: const EdgeInsets.only(right: 10),
                  child: IconButton(
                    focusNode: FocusNode(skipTraversal: true),
                    icon: Icon(Icons.print),
                    onPressed: () async {
                      await generateAndPrintPdf(tableList);
                    },
                  ),
                ),
              ],
            ),
            body: gridView(),
          ),
        );
      },
    );
  }
}
