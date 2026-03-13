import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/report_movement.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/repositories/report_repository.dart';
import 'package:smlaicloud/screen_search/barcode_search_screen.dart';
import 'package:smlaicloud/screen_search/product_location_search_screen.dart';
import 'package:smlaicloud/screen_search/product_warehouse_search_screen.dart';
import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
// ignore: depend_on_referenced_packages
import 'package:intl/intl.dart';
import 'package:smlaicloud/utils/date_picker.dart';

// ignore: avoid_web_libraries_in_flutter

class ReportMovementScreen extends StatefulWidget {
  const ReportMovementScreen({super.key});

  @override
  State<ReportMovementScreen> createState() => _ReportMovementState();
}

class _ReportMovementState extends State<ReportMovementScreen> with global.ThemeRefreshMixin {
  final TextEditingController search = TextEditingController();
  final TextEditingController barcode = TextEditingController();
  final TextEditingController whcode = TextEditingController();
  final TextEditingController lccode = TextEditingController();
  final TextEditingController fromDate = TextEditingController();
  final TextEditingController toDate = TextEditingController();
  ReportMovementModel reportMovements = ReportMovementModel(
    balance: 0,
    details: [],
  );

  final GlobalKey<FormState> _formKey = GlobalKey<FormState>();

  @override
  void initState() {
    super.initState();

    fromDate.text = DateFormat('yyyy-MM-dd').format(DateTime.now());
    toDate.text = DateFormat('yyyy-MM-dd').format(DateTime.now());
  }

  Future<void> getReport() async {
    reportMovements = ReportMovementModel(balance: 0, details: []);
    setState(() {});
    String queryBarcode = "";
    String queryFromdate = "";
    String queryTodate = "";
    String querywhcode = "";
    String querylccode = "";

    if (barcode.text != '') {
      queryBarcode = "&barcode=${barcode.text}";
    }

    if (whcode.text != '') {
      querywhcode = "&whcode =${whcode.text}";
    }
    if (lccode.text != '') {
      querylccode = "&lccode =${lccode.text}";
    }

    if (fromDate.text != '') {
      queryFromdate = "&fromdate=${fromDate.text}";
    }

    if (toDate.text != '') {
      queryTodate = "&todate=${toDate.text}";
    }

    ReportRepository reportRepository = ReportRepository();

    ApiResponse result = await reportRepository.getReportMovement(
      queryBarcode,
      queryFromdate,
      queryTodate,
      querywhcode,
      querylccode,
    );
    if (result.success) {
      /// map json to model
      reportMovements = ReportMovementModel.fromJson(result.data);

      reportMovements.details!.insert(
        0,
        ReportMovementDetail(
          docdate: "",
          barcode: "",
          qty: reportMovements.balance.toString(),
          transflag: 0,
          calcflag: 0,
          balance: 0,
        ),
      );

      _setBalance();

      setState(() {});
    }
  }

  void _setBalance() {
    for (int i = 0; i < reportMovements.details!.length; i++) {
      if (i == 0) {
        reportMovements.details![i].balance = double.parse(
          reportMovements.details![i].qty,
        );
      } else {
        reportMovements.details![i].balance =
            (double.parse(reportMovements.details![i].qty) *
                reportMovements.details![i].calcflag) +
            reportMovements.details![i - 1].balance;
      }
    }
  }

  // _selectFromDate / _selectToDate ถูกแทนที่ด้วย CustomDatePicker widgets

  void barcodeSearch() {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => const BarcodeSearchScreen(word: "", screen: ""),
      ),
    ).then((value) {
      if (value != null) {
        setState(() {
          ProductBarcodeModel result = value;
          if (result.guidfixed.isNotEmpty) {
            search.text =
                "${result.barcode!} ~ ${global.activeLangName(result.names!)}";
            barcode.text = result.barcode!;
          }
        });
      }
    });
  }

  Future<SearchGuidCodeNameModel> warehouseSearch() async {
    SearchGuidCodeNameModel res = SearchGuidCodeNameModel(
      guid: '',
      code: '',
      names: [],
    );
    res = await Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => const ProductWarehouseSearchScreen(word: ''),
      ),
    );
    return res;
  }

  Future<SearchGuidCodeNameModel> locationSearch(String whcode) async {
    SearchGuidCodeNameModel res = SearchGuidCodeNameModel(
      guid: '',
      code: '',
      names: [],
    );
    res = await Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => ProductLocationSearchScreen(whcode: whcode),
      ),
    );
    return res;
  }

  @override
  Widget build(BuildContext context) {
    List<DataRow> dataRows = [];

    for (int i = 0; i < reportMovements.details!.length; i++) {
      dataRows.add(
        DataRow(
          cells: [
            DataCell(Text(reportMovements.details![i].docdate)),
            DataCell(
              Text(
                (i != 0)
                    ? reportMovements.details![i].docno
                    : global.language('balanceamount'),
              ),
            ),
            DataCell(
              Text(
                (i != 0)
                    ? global.language(
                        global.getDocType(
                          reportMovements.details![i].transflag,
                        ),
                      )
                    : '',
                textAlign: TextAlign.center,
              ),
            ),
            DataCell(
              Text(
                (i != 0) ? reportMovements.details![i].qty.toString() : '',
                textAlign: TextAlign.center,
              ),
            ),
            DataCell(
              Text(
                ((double.parse(reportMovements.details![i].qty) *
                            reportMovements.details![i].calcflag.toDouble()) >
                        0)
                    ? (double.parse(reportMovements.details![i].qty) *
                              reportMovements.details![i].calcflag.toDouble())
                          .toString()
                    : "",
                textAlign: TextAlign.right,
              ),
            ),
            DataCell(
              Text(
                ((double.parse(reportMovements.details![i].qty) *
                            reportMovements.details![i].calcflag.toDouble()) <
                        0)
                    ? (double.parse(reportMovements.details![i].qty) *
                              reportMovements.details![i].calcflag.toDouble())
                          .toString()
                    : "",
                textAlign: TextAlign.right,
              ),
            ),
            DataCell(
              Text(
                reportMovements.details![i].balance.toString(),
                textAlign: TextAlign.right,
              ),
            ),
          ],
        ),
      );
    }
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
        title: Text(global.language("report_movement")),
      ),
      body: SingleChildScrollView(
        child: Column(
          children: [
            Card(
              child: Container(
                margin: EdgeInsets.all(10),
                child: Form(
                  key: _formKey,
                  child: Column(
                    children: [
                      Container(
                        margin: EdgeInsets.only(bottom: 10),
                        child: Row(
                          children: [
                            Expanded(
                              child: CustomDatePicker(
                                labelText: global.language("from_date"),
                                initialDate: DateTime.now(),
                                lastDate: DateTime.now(),
                                onDateSelected: (date) {
                                  if (date != null) {
                                    setState(() {
                                      fromDate.text = DateFormat('yyyy-MM-dd').format(date);
                                    });
                                  }
                                },
                                decoration: InputDecoration(),
                              ),
                            ),
                            SizedBox(width: 5),
                            Expanded(
                              child: CustomDatePicker(
                                labelText: global.language("to_date"),
                                initialDate: DateTime.now(),
                                lastDate: DateTime.now(),
                                onDateSelected: (date) {
                                  if (date != null) {
                                    setState(() {
                                      toDate.text = DateFormat('yyyy-MM-dd').format(date);
                                    });
                                  }
                                },
                                decoration: InputDecoration(),
                              ),
                            ),
                          ],
                        ),
                      ),
                      Container(
                        margin: EdgeInsets.only(bottom: 10),
                        child: Row(
                          children: [
                            Expanded(
                              child: TextField(
                                controller: whcode,
                                decoration: InputDecoration(
                                  border: OutlineInputBorder(),
                                  labelText: global.language("warehouse"),
                                  suffixIcon: Row(
                                    mainAxisAlignment: MainAxisAlignment
                                        .spaceBetween, // added line
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      IconButton(
                                        focusNode: FocusNode(
                                          skipTraversal: true,
                                        ),
                                        icon: Icon(Icons.search),
                                        onPressed: () {
                                          warehouseSearch().then((result) {
                                            if (result.code.isNotEmpty) {
                                              whcode.text = result.code;
                                            }
                                            setState(() {});
                                          });
                                        },
                                      ),
                                    ],
                                  ),
                                ),
                              ),
                            ),
                            SizedBox(width: 5),
                            Expanded(
                              child: TextField(
                                enabled: whcode.text.isNotEmpty,
                                controller: lccode,
                                decoration: InputDecoration(
                                  border: OutlineInputBorder(),
                                  labelText: global.language("location"),
                                  suffixIcon: Row(
                                    mainAxisAlignment: MainAxisAlignment
                                        .spaceBetween, // added line
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      IconButton(
                                        focusNode: FocusNode(
                                          skipTraversal: true,
                                        ),
                                        icon: Icon(Icons.search),
                                        onPressed: () {
                                          locationSearch(whcode.text).then((
                                            result,
                                          ) {
                                            if (result.code.isNotEmpty) {
                                              lccode.text = result.code;
                                            }
                                            setState(() {});
                                          });
                                        },
                                      ),
                                    ],
                                  ),
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                      Container(
                        margin: EdgeInsets.only(bottom: 10),
                        child: Row(
                          children: [
                            Expanded(
                              child: Column(
                                children: [
                                  TextFormField(
                                    readOnly: true,
                                    controller: search,
                                    decoration: InputDecoration(
                                      border: OutlineInputBorder(),
                                      labelText: global.language("barcode"),
                                      suffixIcon: Row(
                                        mainAxisAlignment:
                                            MainAxisAlignment.spaceBetween,
                                        mainAxisSize: MainAxisSize.min,
                                        children: [
                                          IconButton(
                                            focusNode: FocusNode(
                                              skipTraversal: true,
                                            ),
                                            icon: Icon(Icons.search),
                                            onPressed: () {
                                              barcodeSearch();
                                            },
                                          ),
                                        ],
                                      ),
                                    ),
                                    onTap: () {
                                      barcodeSearch();
                                    },
                                    validator: (value) {
                                      if (value == null || value.isEmpty) {
                                        return global.language(
                                          "please_select_barcode",
                                        );
                                      }
                                      return null;
                                    },
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                      ),
                      Container(
                        width: double.infinity,
                        height: 40,
                        margin: EdgeInsets.only(bottom: 10),
                        child: ElevatedButton(
                          onPressed: () {
                            if (_formKey.currentState!.validate()) {
                              // If the form is valid, perform the action
                              getReport();
                            }
                          },
                          child: Text(global.language("process")),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
            SizedBox(
              width: double.infinity,
              child: DataTable(
                columns: [
                  DataColumn(label: Text(global.language("doc_date"))),
                  DataColumn(label: Text(global.language("doc_no"))),
                  DataColumn(label: Text(global.language("trans_flag"))),
                  DataColumn(label: Text(global.language("qty"))),
                  DataColumn(label: Text(global.language("in"))),
                  DataColumn(label: Text(global.language("out"))),
                  DataColumn(label: Text(global.language("balance"))),
                ],
                rows: dataRows,
              ),
            ),
          ],
        ),
      ),
    );
  }
}
