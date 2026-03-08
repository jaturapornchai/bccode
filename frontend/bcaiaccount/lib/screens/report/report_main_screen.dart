import 'package:flutter/foundation.dart';
import 'dart:async';

import 'package:smlaicloud/bloc/report/report_bloc.dart';
import 'package:smlaicloud/model/report_main_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/repositories/report_repository.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
// ignore: depend_on_referenced_packages
import 'package:intl/intl.dart';
import 'package:data_table_2/data_table_2.dart';
import 'file_download.dart';
import 'package:shared_preferences/shared_preferences.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class ReportMainScreen extends StatefulWidget {
  final global.ReportEnum type;
  const ReportMainScreen({super.key, required this.type});

  @override
  State<ReportMainScreen> createState() => _ReportMainScreenState();
}

class _ReportMainScreenState extends State<ReportMainScreen> {
  final ReportRepository _reportRepository = ReportRepository();

  final TextEditingController fromDate = TextEditingController();
  final TextEditingController toDate = TextEditingController();
  List<DataRow> dataRows = [];
  List<ReportSaleByDateModel> reportSaleByDateModels = [];
  List<TransactionModel> reportSaleInvoiceModels = [];
  List<ReportReceiveMoneyModel> reportReceiveMoneyModels = [];
  final GlobalKey<FormState> _formKey = GlobalKey<FormState>();
  bool isLoading = false;
  bool showTable = false;
  bool showDetail = false;
  bool showSumByDate = false;
  late Timer _timer;
  String fromDatePref = "";
  String toDatePref = "";
  bool showDetailPref = false;
  bool showSumByDatePref = false;
  String jobId = "";
  String downloadPath = "";
  bool isDownloadReady = true;
  String modulename = "";
  int totalPage = 0;
  int pageActive = 1;
  int perPage = 100;
  bool _isLoading = false;

  @override
  void initState() {
    super.initState();
    if (widget.type == global.ReportEnum.saleinvoice) {
      modulename = "saleinvoice";
    } else if (widget.type == global.ReportEnum.salebydate) {
      modulename = "salebydate";
    } else if (widget.type == global.ReportEnum.receivemoney) {
      modulename = "receivemoney";
    }
    fromDate.text = DateFormat('dd/MM/yyyy').format(DateTime.now());
    toDate.text = DateFormat('dd/MM/yyyy').format(DateTime.now());
    _loadDownloadPath();
    startTimer();

    /// set delay 1 second
    Future.delayed(const Duration(seconds: 1), () {
      _showDateFilterDialog();
    });
  }

  Future<void> _loadMore() async {
    if (pageActive < totalPage) {
      if (_isLoading) return;

      setState(() {
        _isLoading = true;
      });
      String queryFromdate = "";
      String queryTodate = "";

      if (fromDate.text != '') {
        DateTime parsedDate = DateFormat("dd/MM/yyyy").parse(fromDate.text);
        String formattedDate = DateFormat("yyyy-MM-dd").format(parsedDate);

        queryFromdate = "&fromdate=$formattedDate";
      }

      if (toDate.text != '') {
        DateTime parsedToDate = DateFormat("dd/MM/yyyy").parse(toDate.text);
        String formattedToDate = DateFormat("yyyy-MM-dd").format(parsedToDate);

        queryTodate = "&todate=$formattedToDate";
      }
      context.read<ReportBloc>().add(
        GetReport(
          type: widget.type,
          fromdate: queryFromdate,
          todate: queryTodate,
          page: pageActive,
          perpage: perPage,
        ),
      );
      Future.delayed(const Duration(milliseconds: 800), () {
        setState(() {
          pageActive++;
          _isLoading = false;
        });
      });
    }
  }

  void checkFileStatus() async {
    _loadDownloadPath();
    if (downloadPath.isNotEmpty) {
      await _reportRepository
          .checkFileStatus(widget.type, downloadPath, jobId)
          .then((value) {
            if (value.success) {
              isDownloadReady = true;
              _timer.cancel();
              setState(() {});
            } else {
              if (value.message == "regenerated") {
                setState(() {
                  isDownloadReady = false;
                  _timer.cancel();
                });
                pdfDownloadCache();
              }
            }
          })
          .onError((error, stackTrace) {
            if (kDebugMode) {
              AppLogger.error(error.toString());
            }
          });
    }
  }

  Future<void> _saveDownloadPath(String path, jobid) async {
    final prefs = await SharedPreferences.getInstance();
    await prefs.setString('${modulename}_jobid', jobid);
    await prefs.setString('${modulename}_downloadPath', path);
    await prefs.setString('${modulename}_fromDate', fromDate.text);
    await prefs.setString('${modulename}_toDate', toDate.text);
    await prefs.setString('${modulename}_showDetail', showDetail ? "1" : "0");
    await prefs.setString(
      '${modulename}_showSumByDate',
      showSumByDate ? "1" : "0",
    );
  }

  Future<void> _loadDownloadPath() async {
    final prefs = await SharedPreferences.getInstance();
    final jobid = prefs.getString('${modulename}_jobid') ?? '';
    final savedPath = prefs.getString('${modulename}_downloadPath') ?? '';
    fromDatePref = prefs.getString('${modulename}_fromDate') ?? '';
    toDatePref = prefs.getString('${modulename}_toDate') ?? '';
    final showDetailres = prefs.getString('${modulename}_showDetail') ?? '0';
    final showSumByDateres =
        prefs.getString('${modulename}_showSumByDate') ?? '0';
    setState(() {
      downloadPath = savedPath;
      jobId = jobid;
      showDetailPref = showDetailres == "1" ? true : false;
      showSumByDatePref = showSumByDateres == "1" ? true : false;
    });
  }

  /// start timer
  Future<void> startTimer() async {
    _timer = Timer.periodic(
      const Duration(seconds: 6),
      (Timer t) => checkFileStatus(),
    );
  }

  Future<void> getReport() async {
    String queryFromdate = "";
    String queryTodate = "";

    if (fromDate.text != '') {
      DateTime parsedDate = DateFormat("dd/MM/yyyy").parse(fromDate.text);
      String formattedDate = DateFormat("yyyy-MM-dd").format(parsedDate);

      queryFromdate = "&fromdate=$formattedDate";
    }

    if (toDate.text != '') {
      DateTime parsedToDate = DateFormat("dd/MM/yyyy").parse(toDate.text);
      String formattedToDate = DateFormat("yyyy-MM-dd").format(parsedToDate);

      queryTodate = "&todate=$formattedToDate";
    }
    setState(() {
      pageActive = 1;
      totalPage = 0;
      reportSaleByDateModels = [];
    });

    context.read<ReportBloc>().add(
      GetReport(
        type: widget.type,
        fromdate: queryFromdate,
        todate: queryTodate,
        page: pageActive,
        perpage: perPage,
      ),
    );
  }

  void _selectFromDate(BuildContext context) async {
    final DateTime? pickedDate = await showDatePicker(
      context: context,
      initialDate: DateTime.now(),
      firstDate: DateTime(2000),
      lastDate: DateTime.now(),
    );

    if (pickedDate != null) {
      setState(() {
        fromDate.text = DateFormat('dd/MM/yyyy').format(pickedDate);
      });
    }
  }

  @override
  void dispose() {
    _timer.cancel();
    super.dispose();
  }

  void _selectToDate(BuildContext context) async {
    final DateTime? pickedDate = await showDatePicker(
      context: context,
      initialDate: DateTime.now(),
      firstDate: DateTime(2000),
      lastDate: DateTime.now(),
    );

    if (pickedDate != null) {
      setState(() {
        toDate.text = DateFormat('dd/MM/yyyy').format(pickedDate);
      });
    }
  }

  Future<void> _showDateFilterDialog() async {
    return showDialog<void>(
      context: context,
      barrierDismissible: true, // User must tap a button to close the dialog.
      builder: (BuildContext dialogContext) {
        return StatefulBuilder(
          builder: (BuildContext context, StateSetter setState) {
            return AlertDialog(
              title: Text(global.language("filter")),
              content: SingleChildScrollView(
                child: Container(
                  width: 600,
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
                                child: TextField(
                                  decoration: InputDecoration(
                                    border: OutlineInputBorder(),
                                    labelText: global.language("from_date"),
                                    suffixIcon: Row(
                                      mainAxisAlignment: MainAxisAlignment
                                          .spaceBetween, // added line
                                      mainAxisSize: MainAxisSize.min,
                                      children: [
                                        IconButton(
                                          focusNode: FocusNode(
                                            skipTraversal: true,
                                          ),
                                          icon: const Icon(
                                            Icons.calendar_month,
                                          ),
                                          onPressed: () {
                                            _selectFromDate(context);
                                          },
                                        ),
                                      ],
                                    ),
                                  ),
                                  controller: fromDate,
                                  onChanged: (value) {},
                                ),
                              ),
                              SizedBox(width: 5),
                              Expanded(
                                child: TextField(
                                  decoration: InputDecoration(
                                    border: OutlineInputBorder(),
                                    labelText: global.language("to_date"),
                                    suffixIcon: Row(
                                      mainAxisAlignment: MainAxisAlignment
                                          .spaceBetween, // added line
                                      mainAxisSize: MainAxisSize.min,
                                      children: [
                                        IconButton(
                                          focusNode: FocusNode(
                                            skipTraversal: true,
                                          ),
                                          icon: const Icon(
                                            Icons.calendar_month,
                                          ),
                                          onPressed: () {
                                            _selectToDate(context);
                                          },
                                        ),
                                      ],
                                    ),
                                  ),
                                  controller: toDate,
                                  onChanged: (value) {},
                                ),
                              ),
                            ],
                          ),
                        ),
                        (widget.type == global.ReportEnum.saleinvoice)
                            ? Column(
                                children: [
                                  SizedBox(
                                    /// check book show detail
                                    child: CheckboxListTile(
                                      title: Text(
                                        global.language("show_detail"),
                                      ),
                                      value: showDetail,
                                      onChanged: (bool? value) {
                                        setState(() {
                                          showDetail = value!;
                                        });
                                      },
                                      controlAffinity: ListTileControlAffinity
                                          .leading, //  <-- leading Checkbox
                                    ),
                                  ),
                                  SizedBox(
                                    /// check book show detail
                                    child: CheckboxListTile(
                                      title: Text(
                                        global.language("show_sum_by_date"),
                                      ),
                                      value: showSumByDate,
                                      onChanged: (bool? value) {
                                        setState(() {
                                          showSumByDate = value!;
                                        });
                                      },
                                      controlAffinity: ListTileControlAffinity
                                          .leading, //  <-- leading Checkbox
                                    ),
                                  ),
                                ],
                              )
                            : Container(),
                      ],
                    ),
                  ),
                ),
              ),
              actions: <Widget>[
                TextButton(
                  child: Text(global.language("cancel")),
                  onPressed: () {
                    Navigator.of(dialogContext).pop(); // Close the dialog
                  },
                ),
                ElevatedButton(
                  child: Text(global.language("process")),
                  onPressed: () {
                    if (_formKey.currentState!.validate()) {
                      // If the form is valid, perform the action
                      getReport();
                    }
                    Navigator.of(dialogContext).pop(); // Close the dialog
                  },
                ),
              ],
            );
          },
        );
      },
    );
  }

  /// get header report
  List<DataColumn2> getHeaderReport() {
    List<Map<String, String>> columnDefinitions = [];
    Set<String> nonNumericColumns = {
      'date',
      'doc_date',
      'doc_number',
      'debtor',
    };

    if (widget.type == global.ReportEnum.salebydate) {
      columnDefinitions = [
        {"key": "date", "label": global.language("date"), "label2": ""},
        {
          "key": "doc_total_value",
          "label": global.language("doc_total_value"),
          "label2": "",
        },
        {
          "key": "total_discount",
          "label": global.language("total_discount"),
          "label2": "",
        },
        {
          "key": "total_exceptvat",
          "label": global.language("total_exceptvat"),
          "label2": "",
        },
        {
          "key": "total_beforevat",
          "label": global.language("total_beforevat"),
          "label2": "",
        },
        {
          "key": "total_vatvalue",
          "label": global.language("total_vatvalue"),
          "label2": "",
        },
        {
          "key": "total_amount",
          "label": global.language("totalnetworth"),
          "label2": "",
        },
      ];
    } else if (widget.type == global.ReportEnum.receivemoney) {
      columnDefinitions = [
        {"key": "date", "label": global.language("date"), "label2": ""},
        {"key": "cash", "label": global.language("cash"), "label2": ""},
        {
          "key": "credit_amount",
          "label": global.language("credit_amount"),
          "label2": "",
        },
        {
          "key": "transfer_amount",
          "label": global.language("transfer_amount"),
          "label2": "",
        },
        {
          "key": "coupon_amount",
          "label": global.language("coupon_amount"),
          "label2": "",
        },
        {
          "key": "cheque_amount",
          "label": global.language("cheque_amount"),
          "label2": "",
        },
        {
          "key": "total_amount",
          "label": global.language("total_amount"),
          "label2": "",
        },
      ];
    } else if (widget.type == global.ReportEnum.saleinvoice) {
      columnDefinitions = [
        {"key": "doc_date", "label": global.language("doc_date"), "label2": ""},
        {
          "key": "doc_number",
          "label": global.language("doc_number"),
          "label2": global.language("product"),
        },
        {
          "key": "debtor",
          "label": global.language("debtor"),
          "label2": global.language("warehouse"),
        },
        {
          "key": "doc_total_value",
          "label": global.language("doc_total_value"),
          "label2": global.language("storage_area"),
        },
        {
          "key": "discount_amount",
          "label": global.language("discount_amount"),
          "label2": global.language("unit"),
        },
        {
          "key": "after_discount",
          "label": global.language("after_discount"),
          "label2": global.language("quantity"),
        },
        {
          "key": "total_exceptvat",
          "label": global.language("total_exceptvat"),
          "label2": global.language("price"),
        },
        {
          "key": "doc_vat_amount",
          "label": global.language("doc_vat_amount"),
          "label2": global.language("discount"),
        },
        {
          "key": "totalnetworth",
          "label": global.language("totalnetworth"),
          "label2": global.language("total_value"),
        },
      ];
    }

    return columnDefinitions.map((column) {
      return DataColumn2(
        label: (widget.type != global.ReportEnum.saleinvoice)
            ? SizedBox(
                child: Text(
                  column["label"]!,
                  style: const TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              )
            : Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  SizedBox(
                    width: double.infinity,
                    child: Text(
                      column["label"]!,
                      textAlign: TextAlign.center,
                      style: const TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  (showDetail)
                      ? const Divider(
                          height: 10,
                          thickness: 1,
                          indent: 0,
                          endIndent: 0,
                        )
                      : Container(),
                  (showDetail)
                      ? SizedBox(
                          width: double.infinity,
                          child: Text(
                            column["label2"]!,
                            textAlign: TextAlign.center,
                            style: const TextStyle(fontSize: 16),
                          ),
                        )
                      : Container(),
                ],
              ),
        numeric: !nonNumericColumns.contains(column["key"]),
      );
    }).toList();
  }

  String getTitle() {
    if (widget.type == global.ReportEnum.salebydate) {
      return global.language("report_salebydate");
    } else if (widget.type == global.ReportEnum.receivemoney) {
      return global.language("report_receivemoney");
    } else if (widget.type == global.ReportEnum.saleinvoice) {
      return global.language("report_saleinvoice");
    } else {
      return "";
    }
  }

  Future<void> pdfDownloadCache() async {
    await _loadDownloadPath();

    String queryFromdate = "";
    String queryTodate = "";

    DateTime parsedDate = DateFormat("dd/MM/yyyy").parse(fromDatePref);
    String formattedDate = DateFormat("yyyy-MM-dd").format(parsedDate);
    queryFromdate = "&fromdate=$formattedDate";

    DateTime parsedToDate = DateFormat("dd/MM/yyyy").parse(toDatePref);
    String formattedToDate = DateFormat("yyyy-MM-dd").format(parsedToDate);

    queryTodate = "&todate=$formattedToDate";
    if (mounted) {
      context.read<ReportBloc>().add(
        DownloadReport(
          type: widget.type,
          fromdate: queryFromdate,
          todate: queryTodate,
          showDetail: showDetailPref ? 1 : 0,
          showSumByDate: showSumByDatePref ? 1 : 0,
          xorder: 1,
        ),
      );
    }
  }

  void pdfDownload() {
    String queryFromdate = "";
    String queryTodate = "";

    if (fromDate.text != '') {
      DateTime parsedDate = DateFormat("dd/MM/yyyy").parse(fromDate.text);
      String formattedDate = DateFormat("yyyy-MM-dd").format(parsedDate);
      queryFromdate = "&fromdate=$formattedDate";
    }

    if (toDate.text != '') {
      DateTime parsedToDate = DateFormat("dd/MM/yyyy").parse(toDate.text);
      String formattedToDate = DateFormat("yyyy-MM-dd").format(parsedToDate);

      queryTodate = "&todate=$formattedToDate";
    }
    context.read<ReportBloc>().add(
      DownloadReport(
        type: widget.type,
        fromdate: queryFromdate,
        todate: queryTodate,
        showDetail: showDetail ? 1 : 0,
        showSumByDate: showSumByDate ? 1 : 0,
        xorder: 1,
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () {
            _timer.cancel();
            global.gotoMainMenu(context);
          },
        ),
        backgroundColor: global.theme.appBarColor,
        title: Center(child: Text(getTitle())),
        actions: [
          /// download pdf
          // IconButton(
          //   focusNode: FocusNode(skipTraversal: true),
          //   icon: const Icon(Icons.picture_as_pdf),
          //   onPressed: () {
          //     String queryFromdate = "";
          //     String queryTodate = "";

          //     if (fromDate.text != '') {
          //       // Parse the original date string
          //       DateTime parsedDate = DateFormat("dd/MM/yyyy").parse(fromDate.text);

          //       // Format it to the new date string
          //       String formattedDate = DateFormat("yyyy-MM-dd").format(parsedDate);

          //       queryFromdate = "&fromdate=$formattedDate";
          //     }

          //     if (toDate.text != '') {
          //       // Parse the original date string
          //       DateTime parsedToDate = DateFormat("dd/MM/yyyy").parse(toDate.text);

          //       // Format it to the new date string
          //       String formattedToDate = DateFormat("yyyy-MM-dd").format(parsedToDate);

          //       queryTodate = "&todate=$formattedToDate";
          //     }
          //     context.read<ReportBloc>().add(DownloadReport(type: widget.type, fromdate: queryFromdate, todate: queryTodate));
          //   },
          // ),
          IconButton(
            focusNode: FocusNode(skipTraversal: true),
            icon: const Icon(Icons.file_download),
            onPressed: () {
              pdfDownload();
            },
          ),
          const SizedBox(width: 5),
          IconButton(
            focusNode: FocusNode(skipTraversal: true),
            icon: const Icon(Icons.search),
            onPressed: () {
              _showDateFilterDialog();
            },
          ),
        ],
      ),
      body: BlocListener<ReportBloc, ReportState>(
        listener: (context, state) {
          if (state is GetReportInProgress) {
            _setLoading(true);
          } else if (state is GetReportSaleByDateSuccess) {
            _updateSaleByDateReport(state);
          } else if (state is GetReportReceiveMoneySuccess) {
            _updateReceiveMoneyReport(state);
          } else if (state is GetReportSaleInvoiceSuccess) {
            _updateSaleInvoiceReport(state);
          } else if (state is DownloadReportSuccess) {
            _timer.cancel();
            _saveDownloadPath(state.savePath, state.jobId);
            setState(() {
              isDownloadReady = false;
              jobId = state.jobId;
              downloadPath = state.savePath;
              startTimer();
            });
          }
        },
        child: _buildReportView(),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () {
          setState(() {
            pageActive++;
            _loadMore();
          });
        },
        child: const Icon(Icons.navigate_next),
      ),
    );
  }

  void _setLoading(bool loading) {
    setState(() {
      isLoading = loading;
    });
  }

  void _updateSaleByDateReport(GetReportSaleByDateSuccess state) {
    setState(() {
      isLoading = false;
      showTable = true;
      pageActive = state.page.page;
      totalPage = state.page.total;
      reportSaleByDateModels.addAll(state.reportSaleByDateModels);
      dataRows = _buildSaleByDateRows(reportSaleByDateModels);
    });
  }

  void _updateReceiveMoneyReport(GetReportReceiveMoneySuccess state) {
    setState(() {
      isLoading = false;
      showTable = true;
      pageActive = state.page.page;
      totalPage = state.page.total;
      reportReceiveMoneyModels.addAll(state.reportReceiveMoneyModels);
      dataRows = _buildReceiveMoneyRows(reportReceiveMoneyModels);
    });
  }

  void _updateSaleInvoiceReport(GetReportSaleInvoiceSuccess state) {
    setState(() {
      isLoading = false;
      showTable = true;
      pageActive = state.page.page;
      totalPage = state.page.total;
      reportSaleInvoiceModels.addAll(state.reportSaleInvoiceModels);
      dataRows = _buildSaleInvoiceRows(reportSaleInvoiceModels);
    });
  }

  List<DataRow> _buildSaleByDateRows(List<ReportSaleByDateModel> models) {
    List<DataRow> rows = [];
    double totalDetailtotalamount = 0;
    double totalTotaldiscount = 0;
    double totalTotalexceptvat = 0;
    double totalTotalbeforevat = 0;
    double totalTotalvatvalue = 0;
    double totalTotalamount = 0;

    TextStyle styleTextFooter = const TextStyle(
      fontSize: 18,
      fontWeight: FontWeight.bold,
    );

    setState(() {
      if (models.isNotEmpty) {
        for (int i = 0; i < models.length; i++) {
          totalDetailtotalamount += models[i].detailtotalamount;
          totalTotaldiscount += models[i].totaldiscount;
          totalTotalexceptvat += models[i].totalexceptvat;
          totalTotalbeforevat += models[i].totalbeforevat;
          totalTotalvatvalue += models[i].totalvatvalue;
          totalTotalamount += models[i].totalamount;

          DateTime parsedDate = DateFormat(
            "yyyy-MM-dd",
          ).parse(models[i].docdatetime);
          String formattedDate = DateFormat("dd/MM/yyyy").format(parsedDate);

          rows.add(
            DataRow(
              cells: [
                DataCell(SizedBox(child: Text(formattedDate))),
                DataCell(
                  SizedBox(
                    child: Text(
                      global.formatNumber(models[i].detailtotalamount),
                    ),
                  ),
                ),
                DataCell(
                  SizedBox(
                    child: Text(global.formatNumber(models[i].totaldiscount)),
                  ),
                ),
                DataCell(
                  SizedBox(
                    child: Text(global.formatNumber(models[i].totalexceptvat)),
                  ),
                ),
                DataCell(
                  SizedBox(
                    child: Text(global.formatNumber(models[i].totalbeforevat)),
                  ),
                ),
                DataCell(
                  SizedBox(
                    child: Text(global.formatNumber(models[i].totalvatvalue)),
                  ),
                ),
                DataCell(
                  SizedBox(
                    child: Text(global.formatNumber(models[i].totalamount)),
                  ),
                ),
              ],
            ),
          );
        }

        /// footer total
        rows.add(
          DataRow(
            cells: [
              DataCell(
                SizedBox(
                  child: Text(global.language("total"), style: styleTextFooter),
                ),
              ),
              DataCell(
                SizedBox(
                  child: Text(
                    global.formatNumber(totalDetailtotalamount),
                    style: styleTextFooter,
                  ),
                ),
              ),
              DataCell(
                SizedBox(
                  child: Text(
                    global.formatNumber(totalTotaldiscount),
                    style: styleTextFooter,
                  ),
                ),
              ),
              DataCell(
                SizedBox(
                  child: Text(
                    global.formatNumber(totalTotalexceptvat),
                    style: styleTextFooter,
                  ),
                ),
              ),
              DataCell(
                SizedBox(
                  child: Text(
                    global.formatNumber(totalTotalbeforevat),
                    style: styleTextFooter,
                  ),
                ),
              ),
              DataCell(
                SizedBox(
                  child: Text(
                    global.formatNumber(totalTotalvatvalue),
                    style: styleTextFooter,
                  ),
                ),
              ),
              DataCell(
                SizedBox(
                  child: Text(
                    global.formatNumber(totalTotalamount),
                    style: styleTextFooter,
                  ),
                ),
              ),
            ],
          ),
        );
      }
    });
    return rows;
  }

  List<DataRow> _buildReceiveMoneyRows(List<ReportReceiveMoneyModel> models) {
    List<DataRow> rows = [];
    double totalcashamount = 0;
    double totalcreditamount = 0;
    double totaltransferamount = 0;
    double totalcouponamount = 0;
    double totalchequeamount = 0;
    double totaltotalamount = 0;

    TextStyle styleTextFooter = const TextStyle(
      fontSize: 18,
      fontWeight: FontWeight.bold,
    );

    setState(() {
      for (int i = 0; i < models.length; i++) {
        totalcashamount += models[i].data.cashAmount;
        totalcreditamount += models[i].data.creditAmount;
        totaltransferamount += models[i].data.transferAmount;
        totalcouponamount += models[i].data.couponAmount;
        totalchequeamount += models[i].data.chequeAmount;
        totaltotalamount += models[i].data.totalAmount;

        // Parse the original date string
        DateTime parsedDate = DateFormat("yyyy-MM-dd").parse(models[i].date);

        // Format it to the new date string
        String formattedDate = DateFormat("dd/MM/yyyy").format(parsedDate);

        rows.add(
          DataRow(
            cells: [
              DataCell(SizedBox(child: Text(formattedDate))),
              DataCell(
                SizedBox(
                  child: Text(global.formatNumber(models[i].data.cashAmount)),
                ),
              ),
              DataCell(
                SizedBox(
                  child: Text(global.formatNumber(models[i].data.creditAmount)),
                ),
              ),
              DataCell(
                SizedBox(
                  child: Text(
                    global.formatNumber(models[i].data.transferAmount),
                  ),
                ),
              ),
              DataCell(
                SizedBox(
                  child: Text(global.formatNumber(models[i].data.couponAmount)),
                ),
              ),
              DataCell(
                SizedBox(
                  child: Text(global.formatNumber(models[i].data.chequeAmount)),
                ),
              ),
              DataCell(
                SizedBox(
                  child: Text(global.formatNumber(models[i].data.totalAmount)),
                ),
              ),
            ],
          ),
        );
      }

      /// footer total
      rows.add(
        DataRow(
          cells: [
            DataCell(
              SizedBox(
                child: Text(global.language("total"), style: styleTextFooter),
              ),
            ),
            DataCell(
              SizedBox(
                child: Text(
                  global.formatNumber(totalcashamount),
                  style: styleTextFooter,
                ),
              ),
            ),
            DataCell(
              SizedBox(
                child: Text(
                  global.formatNumber(totalcreditamount),
                  style: styleTextFooter,
                ),
              ),
            ),
            DataCell(
              SizedBox(
                child: Text(
                  global.formatNumber(totaltransferamount),
                  style: styleTextFooter,
                ),
              ),
            ),
            DataCell(
              SizedBox(
                child: Text(
                  global.formatNumber(totalcouponamount),
                  style: styleTextFooter,
                ),
              ),
            ),
            DataCell(
              SizedBox(
                child: Text(
                  global.formatNumber(totalchequeamount),
                  style: styleTextFooter,
                ),
              ),
            ),
            DataCell(
              SizedBox(
                child: Text(
                  global.formatNumber(totaltotalamount),
                  style: styleTextFooter,
                ),
              ),
            ),
          ],
        ),
      );
    });

    return rows;
  }

  List<DataRow> _buildSaleInvoiceRows(List<TransactionModel> models) {
    List<DataRow> rows = [];
    TextStyle styleTextRowHerder = const TextStyle(
      fontSize: 16,
      fontWeight: FontWeight.bold,
    );
    TextStyle styleTextFooter = const TextStyle(
      fontSize: 18,
      fontWeight: FontWeight.bold,
    );
    TextStyle styleTextSubFooter = const TextStyle(
      fontSize: 16,
      fontWeight: FontWeight.bold,
    );

    if (!showSumByDate) {
      double totalDetailtotalamount = 0;
      double totalTotaldiscount = 0;
      double totalTotalexceptvat = 0;
      double totalTotalbeforevat = 0;
      double totalTotalvatvalue = 0;
      double totalTotalamount = 0;

      setState(() {
        for (int i = 0; i < models.length; i++) {
          totalDetailtotalamount += models[i].detailtotalamount!;
          totalTotaldiscount += models[i].totaldiscount;
          totalTotalexceptvat += models[i].totalexceptvat;
          totalTotalbeforevat += models[i].totalbeforevat;
          totalTotalvatvalue += models[i].totalvatvalue;
          totalTotalamount += models[i].totalamount;

          DateTime parsedDate = DateFormat(
            "yyyy-MM-dd",
          ).parse(models[i].docdatetime);

          String formattedDate = DateFormat("dd/MM/yyyy").format(parsedDate);

          rows.add(
            DataRow(
              color: (!showDetail)
                  ? WidgetStateColor.resolveWith((states) => Colors.white)
                  : WidgetStateColor.resolveWith(
                      (states) => Colors.yellow[100]!,
                    ),
              cells: [
                DataCell(
                  SizedBox(
                    child: Text(
                      formattedDate,
                      style: (showDetail) ? styleTextRowHerder : null,
                    ),
                  ),
                ),
                DataCell(
                  SizedBox(
                    child: Text(
                      models[i].docno,
                      style: (showDetail) ? styleTextRowHerder : null,
                    ),
                  ),
                ),
                DataCell(
                  SizedBox(
                    child: Text(
                      "${models[i].custcode} ~ ${global.activeLangName(models[i].custnames!)}",
                      style: (showDetail) ? styleTextRowHerder : null,
                    ),
                  ),
                ),
                DataCell(
                  SizedBox(
                    child: Text(
                      global.formatNumber(models[i].detailtotalamount!),
                      style: (showDetail) ? styleTextRowHerder : null,
                    ),
                  ),
                ),
                DataCell(
                  SizedBox(
                    child: Text(
                      global.formatNumber(models[i].totaldiscount),
                      style: (showDetail) ? styleTextRowHerder : null,
                    ),
                  ),
                ),
                DataCell(
                  SizedBox(
                    child: Text(
                      global.formatNumber(models[i].totalexceptvat),
                      style: (showDetail) ? styleTextRowHerder : null,
                    ),
                  ),
                ),
                DataCell(
                  SizedBox(
                    child: Text(
                      global.formatNumber(models[i].totalbeforevat),
                      style: (showDetail) ? styleTextRowHerder : null,
                    ),
                  ),
                ),
                DataCell(
                  SizedBox(
                    child: Text(
                      global.formatNumber(models[i].totalvatvalue),
                      style: (showDetail) ? styleTextRowHerder : null,
                    ),
                  ),
                ),
                DataCell(
                  SizedBox(
                    child: Text(
                      global.formatNumber(models[i].totalamount),
                      style: (showDetail) ? styleTextRowHerder : null,
                    ),
                  ),
                ),
              ],
            ),
          );

          /// add rows detail product
          if (models[i].details != null && showDetail) {
            for (int j = 0; j < models[i].details!.length; j++) {
              rows.add(
                DataRow(
                  color: WidgetStateColor.resolveWith(
                    (states) => Colors.grey[100]!,
                  ),
                  cells: [
                    const DataCell(SizedBox(child: Text(""))),
                    DataCell(
                      SizedBox(
                        child: Text(
                          "${models[i].details![j].itemcode} ~ ${global.activeLangName(models[i].details![j].itemnames!)}",
                        ),
                      ),
                    ),
                    DataCell(
                      SizedBox(
                        child: Text(
                          "${models[i].details![j].whcode} ~ ${global.activeLangName(models[i].details![j].whnames!)}",
                        ),
                      ),
                    ),
                    DataCell(
                      SizedBox(
                        child: Text(
                          "${models[i].details![j].locationcode} ~ ${global.activeLangName(models[i].details![j].locationnames!)}",
                        ),
                      ),
                    ),
                    DataCell(
                      SizedBox(
                        child: Text(
                          "${models[i].details![j].unitcode} ~ ${global.activeLangName(models[i].details![j].unitnames!)}",
                        ),
                      ),
                    ),
                    DataCell(
                      SizedBox(
                        child: Text(
                          global.formatNumber(models[i].details![j].qty),
                        ),
                      ),
                    ),
                    DataCell(
                      SizedBox(
                        child: Text(
                          global.formatNumber(models[i].details![j].price),
                        ),
                      ),
                    ),
                    DataCell(
                      SizedBox(child: Text(models[i].details![j].discount)),
                    ),
                    DataCell(
                      SizedBox(
                        child: Text(
                          global.formatNumber(models[i].details![j].sumamount),
                        ),
                      ),
                    ),
                  ],
                ),
              );
            }
          }
        }

        /// footer total
        rows.add(
          DataRow(
            cells: [
              const DataCell(SizedBox(child: Text(""))),
              DataCell(SizedBox(child: Text(""))),
              DataCell(
                SizedBox(
                  child: Text(global.language("total"), style: styleTextFooter),
                ),
              ),
              DataCell(
                SizedBox(
                  child: Text(
                    global.formatNumber(totalDetailtotalamount),
                    style: styleTextFooter,
                  ),
                ),
              ),
              DataCell(
                SizedBox(
                  child: Text(
                    global.formatNumber(totalTotaldiscount),
                    style: styleTextFooter,
                  ),
                ),
              ),
              DataCell(
                SizedBox(
                  child: Text(
                    global.formatNumber(totalTotalexceptvat),
                    style: styleTextFooter,
                  ),
                ),
              ),
              DataCell(
                SizedBox(
                  child: Text(
                    global.formatNumber(totalTotalbeforevat),
                    style: styleTextFooter,
                  ),
                ),
              ),
              DataCell(
                SizedBox(
                  child: Text(
                    global.formatNumber(totalTotalvatvalue),
                    style: styleTextFooter,
                  ),
                ),
              ),
              DataCell(
                SizedBox(
                  child: Text(
                    global.formatNumber(totalTotalamount),
                    style: styleTextFooter,
                  ),
                ),
              ),
            ],
          ),
        );
      });
    } else {
      Map<String, Map<String, dynamic>> groupedResults = {};

      for (var entry in models) {
        String dateKey = DateTime.parse(
          entry.docdatetime,
        ).toIso8601String().split('T')[0];

        groupedResults.putIfAbsent(
          dateKey,
          () => {
            'detailtotalamount': 0.0,
            'totaldiscount': 0.0,
            'totalexceptvat': 0.0,
            'totalbeforevat': 0.0,
            'totalvatvalue': 0.0,
            'totalamount': 0.0,
            'details': [],
          },
        );

        groupedResults[dateKey]!['detailtotalamount'] =
            (groupedResults[dateKey]!['detailtotalamount'] ?? 0.0) +
            (entry.detailtotalamount ?? 0.0);
        groupedResults[dateKey]!['totaldiscount'] =
            (groupedResults[dateKey]!['totaldiscount'] ?? 0.0) +
            (entry.totaldiscount);
        groupedResults[dateKey]!['totalexceptvat'] =
            (groupedResults[dateKey]!['totalexceptvat'] ?? 0.0) +
            (entry.totalexceptvat);
        groupedResults[dateKey]!['totalbeforevat'] =
            (groupedResults[dateKey]!['totalbeforevat'] ?? 0.0) +
            (entry.totalbeforevat);
        groupedResults[dateKey]!['totalvatvalue'] =
            (groupedResults[dateKey]!['totalvatvalue'] ?? 0.0) +
            (entry.totalvatvalue);
        groupedResults[dateKey]!['totalamount'] =
            (groupedResults[dateKey]!['totalamount'] ?? 0.0) +
            (entry.totalamount);

        groupedResults[dateKey]!['details'].add({
          'docdatetime': entry.docdatetime,
          'docno': entry.docno,
          'custcode': entry.custcode,
          'custnames': entry.custnames,
          'detailtotalamount': entry.detailtotalamount,
          'totaldiscount': entry.totaldiscount,
          'totalexceptvat': entry.totalexceptvat,
          'totalbeforevat': entry.totalbeforevat,
          'totalvatvalue': entry.totalvatvalue,
          'totalamount': entry.totalamount,
          'details': entry.details,
        });
      }

      List<Map<String, dynamic>> resultList = groupedResults.entries
          .map(
            (e) => {
              'docdate': e.key,
              'detailtotalamount': e.value['detailtotalamount'],
              'totaldiscount': e.value['totaldiscount'],
              'totalexceptvat': e.value['totalexceptvat'],
              'totalbeforevat': e.value['totalbeforevat'],
              'totalvatvalue': e.value['totalvatvalue'],
              'totalamount': e.value['totalamount'],
              'details': e.value['details'],
            },
          )
          .toList();

      for (var data in resultList) {
        if (data['details'] != null) {
          DateTime parsedDate = DateFormat("yyyy-MM-dd").parse(data['docdate']);
          String formattedDate = DateFormat("dd/MM/yyyy").format(parsedDate);
          for (var detail in data['details'] as List<dynamic>) {
            rows.add(
              DataRow(
                color: (!showDetail)
                    ? WidgetStateColor.resolveWith((states) => Colors.white)
                    : WidgetStateColor.resolveWith(
                        (states) => Colors.yellow[100]!,
                      ),
                cells: [
                  DataCell(
                    SizedBox(
                      child: Text(
                        formattedDate,
                        style: (showDetail) ? styleTextRowHerder : null,
                      ),
                    ),
                  ),
                  DataCell(
                    SizedBox(
                      child: Text(
                        detail['docno'],
                        style: (showDetail) ? styleTextRowHerder : null,
                      ),
                    ),
                  ),
                  DataCell(
                    SizedBox(
                      child: Text(
                        "${detail['custcode']} ~ ${global.activeLangName(detail['custnames'])}",
                        style: (showDetail) ? styleTextRowHerder : null,
                      ),
                    ),
                  ),
                  DataCell(
                    SizedBox(
                      child: Text(
                        global.formatNumber(detail['detailtotalamount']),
                        style: (showDetail) ? styleTextRowHerder : null,
                      ),
                    ),
                  ),
                  DataCell(
                    SizedBox(
                      child: Text(
                        global.formatNumber(detail['totaldiscount']),
                        style: (showDetail) ? styleTextRowHerder : null,
                      ),
                    ),
                  ),
                  DataCell(
                    SizedBox(
                      child: Text(
                        global.formatNumber(detail['totalexceptvat']),
                        style: (showDetail) ? styleTextRowHerder : null,
                      ),
                    ),
                  ),
                  DataCell(
                    SizedBox(
                      child: Text(
                        global.formatNumber(detail['totalbeforevat']),
                        style: (showDetail) ? styleTextRowHerder : null,
                      ),
                    ),
                  ),
                  DataCell(
                    SizedBox(
                      child: Text(
                        global.formatNumber(detail['totalvatvalue']),
                        style: (showDetail) ? styleTextRowHerder : null,
                      ),
                    ),
                  ),
                  DataCell(
                    SizedBox(
                      child: Text(
                        global.formatNumber(detail['totalamount']),
                        style: (showDetail) ? styleTextRowHerder : null,
                      ),
                    ),
                  ),
                ],
              ),
            );

            if (detail['details'] != null && showDetail) {
              for (var items in detail['details'] as List<dynamic>) {
                if (items != null) {
                  rows.add(
                    DataRow(
                      color: WidgetStateColor.resolveWith(
                        (states) => Colors.grey[100]!,
                      ),
                      cells: [
                        const DataCell(SizedBox(child: Text(""))),
                        DataCell(
                          SizedBox(
                            child: Text(
                              "${items.itemcode ?? ''} ~ ${global.activeLangName(items.itemnames)}",
                            ),
                          ),
                        ),
                        DataCell(
                          SizedBox(
                            child: Text(
                              "${items.whcode ?? ''} ~ ${global.activeLangName(items.whnames)}",
                            ),
                          ),
                        ),
                        DataCell(
                          SizedBox(
                            child: Text(
                              "${items.locationcode ?? ''} ~ ${global.activeLangName(items.locationnames)}",
                            ),
                          ),
                        ),
                        DataCell(
                          SizedBox(
                            child: Text(
                              "${items.unitcode ?? ''} ~ ${global.activeLangName(items.unitnames)}",
                            ),
                          ),
                        ),
                        DataCell(
                          SizedBox(child: Text(global.formatNumber(items.qty))),
                        ),
                        DataCell(
                          SizedBox(
                            child: Text(global.formatNumber(items.price)),
                          ),
                        ),
                        DataCell(SizedBox(child: Text(items.discount ?? ''))),
                        DataCell(
                          SizedBox(
                            child: Text(global.formatNumber(items.sumamount)),
                          ),
                        ),
                      ],
                    ),
                  );
                }
              }
            }
          }

          rows.add(
            DataRow(
              cells: [
                const DataCell(SizedBox(child: Text(""))),
                DataCell(SizedBox(child: Text(""))),
                DataCell(
                  SizedBox(
                    child: Text(
                      "${global.language("total")} $formattedDate",
                      style: styleTextSubFooter,
                    ),
                  ),
                ),
                DataCell(
                  SizedBox(
                    child: Text(
                      global.formatNumber(data['detailtotalamount']),
                      style: styleTextSubFooter,
                    ),
                  ),
                ),
                DataCell(
                  SizedBox(
                    child: Text(
                      global.formatNumber(data['totaldiscount']),
                      style: styleTextSubFooter,
                    ),
                  ),
                ),
                DataCell(
                  SizedBox(
                    child: Text(
                      global.formatNumber(data['totalexceptvat']),
                      style: styleTextSubFooter,
                    ),
                  ),
                ),
                DataCell(
                  SizedBox(
                    child: Text(
                      global.formatNumber(data['totalbeforevat']),
                      style: styleTextSubFooter,
                    ),
                  ),
                ),
                DataCell(
                  SizedBox(
                    child: Text(
                      global.formatNumber(data['totalvatvalue']),
                      style: styleTextSubFooter,
                    ),
                  ),
                ),
                DataCell(
                  SizedBox(
                    child: Text(
                      global.formatNumber(data['totalamount']),
                      style: styleTextSubFooter,
                    ),
                  ),
                ),
              ],
            ),
          );
        }
      }

      double totalDetailtotalamount = 0;
      double totalTotaldiscount = 0;
      double totalTotalexceptvat = 0;
      double totalTotalbeforevat = 0;
      double totalTotalvatvalue = 0;
      double totalTotalamount = 0;

      for (int i = 0; i < models.length; i++) {
        totalDetailtotalamount += models[i].detailtotalamount!;
        totalTotaldiscount += models[i].totaldiscount;
        totalTotalexceptvat += models[i].totalexceptvat;
        totalTotalbeforevat += models[i].totalbeforevat;
        totalTotalvatvalue += models[i].totalvatvalue;
        totalTotalamount += models[i].totalamount;
      }
      rows.add(
        DataRow(
          cells: [
            const DataCell(SizedBox(child: Text(""))),
            DataCell(SizedBox(child: Text(""))),
            DataCell(
              SizedBox(
                child: Text(global.language("total"), style: styleTextFooter),
              ),
            ),
            DataCell(
              SizedBox(
                child: Text(
                  global.formatNumber(totalDetailtotalamount),
                  style: styleTextFooter,
                ),
              ),
            ),
            DataCell(
              SizedBox(
                child: Text(
                  global.formatNumber(totalTotaldiscount),
                  style: styleTextFooter,
                ),
              ),
            ),
            DataCell(
              SizedBox(
                child: Text(
                  global.formatNumber(totalTotalexceptvat),
                  style: styleTextFooter,
                ),
              ),
            ),
            DataCell(
              SizedBox(
                child: Text(
                  global.formatNumber(totalTotalbeforevat),
                  style: styleTextFooter,
                ),
              ),
            ),
            DataCell(
              SizedBox(
                child: Text(
                  global.formatNumber(totalTotalvatvalue),
                  style: styleTextFooter,
                ),
              ),
            ),
            DataCell(
              SizedBox(
                child: Text(
                  global.formatNumber(totalTotalamount),
                  style: styleTextFooter,
                ),
              ),
            ),
          ],
        ),
      );
    }
    return rows;
  }

  Widget _downloadView() {
    return (downloadPath.isNotEmpty)
        ? Card(
            child: Container(
              margin: const EdgeInsets.all(10),
              child: Row(
                children: [
                  Expanded(
                    child: Text(
                      "PDF Download : $downloadPath",
                      style: const TextStyle(
                        fontSize: 18,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  (isDownloadReady)
                      ? TextButton(
                          onPressed: () {
                            downloadFile(
                              downloadPath,
                              downloadPath.split('/').last,
                            );
                            Future.delayed(const Duration(seconds: 1), () {
                              setState(() {
                                _saveDownloadPath("", "");
                                downloadPath = "";
                                jobId = "";
                                isDownloadReady = false;
                                _timer.cancel();
                              });
                            });
                          },
                          child: Text(global.language('download')),
                        )
                      : Text(global.language('processing')),
                ],
              ),
            ),
          )
        : Container();
  }

  Widget _buildReportView() {
    if (isLoading) {
      return const Center(child: CircularProgressIndicator());
    } else if (showTable) {
      return Column(
        children: [
          _downloadView(),
          Padding(
            padding: EdgeInsets.all(8.0),
            child: Container(
              margin: EdgeInsets.only(left: 10, bottom: 10),
              child: Text(
                "${global.language("from_date")} ${fromDate.text} - ${global.language("to_date")} ${toDate.text}",
                style: const TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(8.0),
            child: Container(
              margin: const EdgeInsets.only(left: 10, bottom: 10),
              child: Text(
                "Page $pageActive of $totalPage",
                style: const TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
          ),
          Expanded(
            child: DataTable2(
              minWidth: 500,
              headingTextStyle: const TextStyle(color: Colors.black),
              horizontalMargin: 20,
              columnSpacing: 0,
              headingRowHeight: (widget.type != global.ReportEnum.saleinvoice)
                  ? 40
                  : (showDetail)
                  ? 70
                  : 40,
              headingRowColor: WidgetStateColor.resolveWith(
                (states) => Colors.grey[200]!,
              ),
              border: (widget.type != global.ReportEnum.saleinvoice)
                  ? TableBorder(
                      top: const BorderSide(color: Colors.black),
                      bottom: BorderSide(color: Colors.grey[300]!),
                      left: BorderSide(color: Colors.grey[300]!),
                      right: BorderSide(color: Colors.grey[300]!),
                      verticalInside: BorderSide(color: Colors.grey[300]!),
                      horizontalInside: const BorderSide(
                        color: Colors.grey,
                        width: 1,
                      ),
                    )
                  : null,
              columns: getHeaderReport(),
              rows: [...dataRows],
              empty: Center(
                child: Container(
                  padding: const EdgeInsets.all(20),
                  color: Colors.grey[200],
                  child: const Text('No data'),
                ),
              ),
            ),
          ),
          Padding(
            padding: const EdgeInsets.all(8.0),
            child: Container(
              margin: const EdgeInsets.only(left: 10, bottom: 10),
              child: Text(
                "Page $pageActive of $totalPage",
                style: const TextStyle(
                  fontSize: 18,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ),
          ),
        ],
      );
    } else {
      return Column(
        children: [
          _downloadView(),
          const SizedBox(height: 30),
          const Center(child: Text('No data available')),
        ],
      );
    }
  }
}
