import 'package:smlaicloud/environment.dart';
import 'package:smlaicloud/global.dart';
import 'package:flutter/material.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/report_product_balance_model.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/repositories/report_repository.dart';
import 'package:smlaicloud/screen_search/barcode_search_screen.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/screens/report/file_download.dart';

class ReportProductBalanceScreen extends StatefulWidget {
  const ReportProductBalanceScreen({super.key});

  @override
  State<ReportProductBalanceScreen> createState() => _ReportMovementState();
}

class _ReportMovementState extends State<ReportProductBalanceScreen> with global.ThemeRefreshMixin {
  final TextEditingController search = TextEditingController();

  List<ReportProductBalanceModel> reportProductBalance = [];
  ReportProductBalanceModel reportProductBalanceFooter =
      ReportProductBalanceModel();

  @override
  void initState() {
    super.initState();
  }

  Future<void> getReport() async {
    reportProductBalance = [];
    setState(() {});
    String queryBarcode = "";

    if (search.text != '') {
      queryBarcode = "&barcode=${search.text}";
    }

    ReportRepository reportRepository = ReportRepository();

    ApiResponse result = await reportRepository.getReportProductBalance(
      queryBarcode,
    );
    if (result.success) {
      if (result.data.length > 0) {
        reportProductBalance = (result.data as List)
            .map((reportData) => ReportProductBalanceModel.fromJson(reportData))
            .toList();

        double sumBalanceAmount = 0.0;
        double balanceqty = 0.0;

        for (var element in reportProductBalance) {
          sumBalanceAmount += double.parse(element.balanceamount!);
          balanceqty += double.parse(element.balanceqty!);
        }

        reportProductBalanceFooter = ReportProductBalanceModel(
          barcode: "",
          names: [],
          standunit: global.language('total_amount'),
          balanceqty: global.formatNumber(double.parse(balanceqty.toString())),
          averagecost: "",
          balanceamount: global.formatNumber(
            double.parse(sumBalanceAmount.toString()),
          ),
        );

        setState(() {});
      }
    }
  }

  void downloadFileFromUrl(String url) {
    downloadFile(url, "product_balance.pdf");
  }

  Widget searchBox(ProductBarcodeModel value) {
    return Container(
      margin: EdgeInsets.only(bottom: 10),
      child: Row(
        children: [
          Expanded(
            child: Column(
              children: [
                TextField(
                  controller: search,
                  decoration: InputDecoration(
                    border: OutlineInputBorder(),
                    labelText: global.language("barcode"),
                    suffixIcon: Row(
                      mainAxisAlignment:
                          MainAxisAlignment.spaceBetween, // added line
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        IconButton(
                          focusNode: FocusNode(skipTraversal: true),
                          icon: Icon(Icons.search),
                          onPressed: () {
                            barcodeSearch().then((result) {
                              if (result.guidfixed.isNotEmpty) {
                                search.text = result.barcode!;
                              }
                              setState(() {});
                            });
                          },
                        ),
                      ],
                    ),
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Future<ProductBarcodeModel> barcodeSearch() async {
    ProductBarcodeModel res = ProductBarcodeModel(guidfixed: '', itemcode: '');
    res = await Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => const BarcodeSearchScreen(word: '', screen: ''),
      ),
    );
    return res;
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: Icon(Icons.arrow_back),
          onPressed: () {
            Navigator.pop(context);
          },
        ),
        backgroundColor: global.theme.appBarColor,
        title: Text(global.language("report_product_balance")),
      ),
      body: SingleChildScrollView(
        child: Column(
          children: [
            Card(
              child: Container(
                margin: EdgeInsets.all(10),
                child: Column(
                  children: [
                    searchBox(ProductBarcodeModel(barcode: "", guidfixed: "")),
                    Row(
                      children: [
                        Container(
                          alignment: Alignment.centerLeft,
                          margin: EdgeInsets.only(bottom: 10),
                          child: ElevatedButton(
                            onPressed: () {
                              getReport();
                            },
                            child: Text(global.language('process')),
                          ),
                        ),
                        const SizedBox(width: 10),
                        Container(
                          alignment: Alignment.centerLeft,
                          margin: const EdgeInsets.only(bottom: 10),
                          child: ElevatedButton.icon(
                            /// set color for button
                            style: ElevatedButton.styleFrom(
                              backgroundColor: global.theme.positiveHighlightTextColor,
                            ),
                            onPressed: () async {
                              final token = appConfig.getString("token");
                              downloadFileFromUrl(
                                '${Environment().config.reportApi}/productbalance/pdfdownload?token=$token&barcode=${search.text}',
                              );
                            },
                            icon: Icon(Icons.download),
                            label: Text("Download"),
                          ),
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            ),
            Padding(
              padding: EdgeInsets.all(10.0),
              child: ListView(
                shrinkWrap: true,
                children: [
                  Row(
                    children: [
                      Expanded(child: Text(global.language("sequence"))),
                      Expanded(child: Text(global.language("barcode"))),
                      Expanded(child: Text(global.language("product_name"))),
                      Expanded(child: Text(global.language("remaining_unit"))),
                      Expanded(
                        child: Text(
                          global.language("balance_qty"),
                          textAlign: TextAlign.right,
                        ),
                      ),
                      Expanded(
                        child: Text(
                          global.language("average_cost"),
                          textAlign: TextAlign.right,
                        ),
                      ),
                      Expanded(
                        child: Text(
                          global.language("balance_amount"),
                          textAlign: TextAlign.right,
                        ),
                      ),
                    ],
                  ),

                  /// for reportProductBalance
                  for (int i = 0; i < reportProductBalance.length; i++)
                    Row(
                      children: [
                        Expanded(child: Text((i + 1).toString())),
                        Expanded(child: Text(reportProductBalance[i].barcode!)),
                        Expanded(
                          child: Text(
                            "${global.packName(reportProductBalance[i].names!)}/${reportProductBalance[i].unitcode!}",
                          ),
                        ),
                        Expanded(
                          child: Text(reportProductBalance[i].standunit!),
                        ),
                        Expanded(
                          child: Text(
                            global.formatNumber(
                              double.parse(
                                reportProductBalance[i].balanceqty.toString(),
                              ),
                            ),
                            textAlign: TextAlign.right,
                          ),
                        ),
                        Expanded(
                          child: Text(
                            global.formatNumber(
                              double.parse(
                                reportProductBalance[i].averagecost.toString(),
                              ),
                            ),
                            textAlign: TextAlign.right,
                          ),
                        ),
                        Expanded(
                          child: Text(
                            global.formatNumber(
                              double.parse(
                                reportProductBalance[i].balanceamount
                                    .toString(),
                              ),
                            ),
                            textAlign: TextAlign.right,
                          ),
                        ),
                      ],
                    ),

                  /// reportProductBalanceFooter
                  Row(
                    children: [
                      const Expanded(child: Text("")),
                      Expanded(
                        child: Text(reportProductBalanceFooter.barcode!),
                      ),
                      const Expanded(child: Text("")),
                      Expanded(
                        child: Text(reportProductBalanceFooter.standunit!),
                      ),
                      Expanded(
                        child: Text(
                          reportProductBalanceFooter.balanceqty!,
                          textAlign: TextAlign.right,
                          style: TextStyle(fontWeight: FontWeight.bold),
                        ),
                      ),
                      Expanded(
                        child: Text(
                          reportProductBalanceFooter.averagecost!,
                          textAlign: TextAlign.right,
                          style: TextStyle(fontWeight: FontWeight.bold),
                        ),
                      ),
                      Expanded(
                        child: Text(
                          reportProductBalanceFooter.balanceamount!,
                          textAlign: TextAlign.right,
                          style: TextStyle(fontWeight: FontWeight.bold),
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}
