import 'dart:async';

import 'package:smlaicloud/bloc/stock_balance/stock_balance_bloc.dart';
import 'package:smlaicloud/imports_bloc.dart';
import 'package:smlaicloud/model/employee_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/pagination.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/stock_balance_import_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/repositories/employee_repository.dart';
import 'package:smlaicloud/repositories/product_barcode_repository.dart';
import 'package:smlaicloud/screen_search/barcode_search_screen.dart';
import 'package:smlaicloud/screen_search/employee_search_screen.dart';
import 'package:smlaicloud/screen_search/product_location_search_screen.dart';
import 'package:smlaicloud/screen_search/product_warehouse_search_screen.dart';
import 'package:smlaicloud/screen_search/transaction_search_screen.dart';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/utils/date_picker.dart';
import 'package:split_view/split_view.dart';
import 'package:uuid/uuid.dart';
import 'package:smlaicloud/global.dart' as global;
// ignore: depend_on_referenced_packages
import 'package:intl/intl.dart';
import 'package:smlaicloud/screens/report/file_download.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

enum TransactionStockBalaceScreenModule { header, detail, footer }

class TransactionStockBalaceScreen extends StatefulWidget {
  final global.TransactionTypeEnum type;

  const TransactionStockBalaceScreen({super.key, required this.type});

  @override
  State<TransactionStockBalaceScreen> createState() =>
      TransactionStockBalaceScreenState();
}

class TransactionStockBalaceScreenState
    extends State<TransactionStockBalaceScreen>
    with TickerProviderStateMixin, global.ThemeRefreshMixin {
  late SplitViewController splitViewController;
  late TabController editTabController;
  late TransactionModel screenData;
  late TransactionStockBalaceScreenModule moduleEdit;
  final ProductBarcodeRepository _productBarcodeRepository =
      ProductBarcodeRepository();
  TextEditingController docNumberController = TextEditingController();
  TextEditingController docDateController = TextEditingController();
  TextEditingController docTimeController = TextEditingController();
  TextEditingController docRefNumberController = TextEditingController();
  TextEditingController docRefDateController = TextEditingController();
  TextEditingController saleCodeController = TextEditingController();
  TextEditingController saleNameController = TextEditingController();
  TextEditingController descriptionController = TextEditingController();

  final DateTime _selectedDate = DateTime.now();
  List<Widget> tab = [];
  int transflag = 0;
  int calcflag = 0;
  String docnoFormat = "";

  bool docDateTimeValidated = false;
  final _debouncer = global.Debouncer(1000);

  String fileImportName = "";
  bool isLoading = false;
  bool isLoadingText = false;
  bool isLoadingTable = false;
  bool isShowTable = false;

  List<StockBalanceImportModel> detailStockBalanceImportModel = [];
  List<TransactionDetailModel> detailTransactionModel = [];
  Pagination pagination = Pagination(
    page: 1,
    perPage: 20,
    total: 10,
    totalPage: 1,
    next: 0,
    prev: 0,
  );

  final List<int> _limitOptions = [20, 50, 100];
  String taskid = "";
  String q = "";
  int limit = 20;
  int page = 1;
  double screenHeight = 0;

  TotalModel totalModel = TotalModel(totalitem: 0, totalamount: 0);

  void setSystemLanguageList() async {
    clearScreenData();
    await global.setSystemLanguage(context);
  }

  @override
  void initState() {
    super.initState();

    global.getDeviceModel(context);
    setSystemLanguageList();
    splitViewController = SplitViewController(
      limits: [null, WeightLimit(min: 0.1, max: 0.9)],
    );
    splitViewController.weights = [0.30, 0.70];

    editTabController = TabController(length: 2, vsync: this);
  }

  @override
  void dispose() {
    splitViewController.dispose();
    editTabController.dispose();
    docNumberController.dispose();
    docDateController.dispose();
    docTimeController.dispose();
    docRefNumberController.dispose();
    docRefDateController.dispose();
    saleCodeController.dispose();
    saleNameController.dispose();
    descriptionController.dispose();

    super.dispose();
  }

  void loadDataToScreen() {
    DateTime docDateTimeFormat = DateTime.parse(screenData.docdatetime);
    DateTime docRefDateTimeFormat = DateTime.parse(screenData.docrefdate);

    docNumberController.text = screenData.docno;
    docDateController.text = DateFormat(
      'dd/MM/yyyy',
    ).format(docDateTimeFormat.toLocal());
    docTimeController.text = DateFormat(
      'HH:mm',
    ).format(docDateTimeFormat.toLocal());
    docRefNumberController.text = screenData.docrefno;
    docRefDateController.text = DateFormat(
      'dd/MM/yyyy',
    ).format(docRefDateTimeFormat.toLocal());

    saleCodeController.text = screenData.salecode;
    saleNameController.text = screenData.salename;

    descriptionController.text = screenData.description.toString();

    if (global.profileData.yeartype == "buddhist") {
      docDateController.text = global.dateTimeBuddhist(
        docDateTimeFormat,
        format: global.DateTimeFormatEnum.dateDay,
      );
      docRefDateController.text = global.dateTimeBuddhist(
        docRefDateTimeFormat,
        format: global.DateTimeFormatEnum.dateDay,
      );
    } else {
      docDateController.text = DateFormat(
        'dd/MM/yyyy',
      ).format(docDateTimeFormat);
      docRefDateController.text = DateFormat(
        'dd/MM/yyyy',
      ).format(docRefDateTimeFormat);
    }
  }

  void clearScreenData() {
    if (widget.type == global.TransactionTypeEnum.stockreceiveproduct) {
      transflag = 60;
      calcflag = 1;
      docnoFormat = "IF";
    } else if (widget.type == global.TransactionTypeEnum.stockbalance) {
      transflag = 54;
      calcflag = 1;
      docnoFormat = "IB";
    }

    const uuid = Uuid();
    moduleEdit = TransactionStockBalaceScreenModule.header;
    // ดึงข้อมูล timezone สำหรับส่งไป mainapi
    final timezoneInfo = global.getLocalTimezoneInfo();
    screenData = TransactionModel(
      shopid: global.apiShopCode,
      guidref: uuid.v4(),
      docno: '', // mainapi จะสร้าง docno เอง
      docdatetime: DateTime.now().toUtc().toIso8601String(), // แปลงเป็น UTC ตามกฎ CLAUDE.md
      docdatelocal: timezoneInfo['docdatelocal'], // วันที่ local สำหรับ docno
      doctimelocal: timezoneInfo['doctimelocal'], // เวลา local
      timezone: timezoneInfo['timezone'], // timezone ของ frontend
      docrefno: '',
      docrefdate: DateTime.now().toUtc().toIso8601String(), // แปลงเป็น UTC ตามกฎ CLAUDE.md
      docreftype: 0,
      doctype: 0,
      vattype: 0,
      custcode: '',
      custnames: [],
      salecode: '',
      salename: '',
      discountword: '',
      totalcost: 0,
      totalvalue: 0,
      totaldiscount: 0,
      totalvatvalue: 0,
      totalaftervat: 0,
      totalexceptvat: 0,
      totalamount: 0,
      cashiercode: '',
      posid: '',
      membercode: '',
      vatrate: global.config.vatrate,
      status: 0,
      inquirytype: 0,
      taxdocdate: DateTime.now().toLocal().toIso8601String(),
      taxdocno: '',
      totalbeforevat: 0,
      transflag: transflag,
      details: <TransactionDetailModel>[],
      iscancel: false,
      description: '',
      ismanualamount: false,
      isref: false,
      islocked: false,
      isclosed: false,
      detailcount: 0,
      paymentdetail: TransactionPayModel(
        cashamount: 0,
        cashamounttext: '',
        paymentcreditcards: [],
        paymenttransfers: [],
      ),
    );

    pagination = Pagination(
      page: 1,
      perPage: 20,
      total: 10,
      totalPage: 1,
      next: 0,
      prev: 0,
    );
    detailStockBalanceImportModel = <StockBalanceImportModel>[];
    isLoading = false;
    isLoadingText = false;
    isLoadingTable = false;
    isShowTable = false;
    fileImportName = '';

    setState(() {
      loadDataToScreen();
      docDateTimeValidated = true;
    });
  }

  Widget editDocumentWidget() {
    return Container(
      width: double.infinity,
      padding: EdgeInsets.all(10),
      child: SingleChildScrollView(
        child: Column(
          children: [
            Row(
              children: [
                Expanded(
                  child: CustomDatePicker(
                    key: ValueKey(screenData.docdatetime),
                    labelText: global.language("doc_date"),
                    initialDate: DateTime.parse(screenData.docdatetime),
                    useBuddhistCalendar:
                        global.profileData.yeartype == "buddhist",
                    onDateSelected: (date) {
                      setState(() {
                        docDateTimeValidated = true;

                        // รวมวันที่ใหม่กับเวลาเดิม
                        DateTime currentDateTime = DateTime.parse(
                          screenData.docdatetime,
                        );
                        DateTime newDateTime = DateTime(
                          date!.year,
                          date.month,
                          date.day,
                          currentDateTime.hour,
                          currentDateTime.minute,
                          currentDateTime.second,
                          currentDateTime.millisecond,
                        );

                        screenData.docdatetime = newDateTime
                            .toLocal()
                            .toIso8601String();

                        // อัพเดท controller สำหรับการแสดงผล
                        if (global.profileData.yeartype == "buddhist") {
                          docDateController.text = global.dateTimeBuddhist(
                            newDateTime,
                            format: global.DateTimeFormatEnum.dateDay,
                          );
                        } else {
                          docDateController.text = DateFormat(
                            'dd/MM/yyyy',
                          ).format(newDateTime);
                        }
                      });
                    },
                    decoration: InputDecoration(
                      border: OutlineInputBorder(),
                      labelText: global.language("doc_date"),
                    ),
                  ),
                ),
                SizedBox(width: 10),
                Expanded(
                  child: TextField(
                    decoration: InputDecoration(
                      border: OutlineInputBorder(),
                      labelText: global.language("doc_time"),
                    ),
                    controller: docTimeController,
                    onChanged: (value) {
                      setState(() {
                        try {
                          // รวมวันที่เดิมกับเวลาใหม่
                          DateTime currentDate = DateTime.parse(
                            screenData.docdatetime,
                          );
                          List<String> timeParts = value.split(':');
                          if (timeParts.length >= 2) {
                            int hour =
                                int.tryParse(timeParts[0]) ?? currentDate.hour;
                            int minute =
                                int.tryParse(timeParts[1]) ??
                                currentDate.minute;

                            DateTime newDateTime = DateTime(
                              currentDate.year,
                              currentDate.month,
                              currentDate.day,
                              hour,
                              minute,
                              currentDate.second,
                              currentDate.millisecond,
                            );

                            screenData.docdatetime = newDateTime
                                .toLocal()
                                .toIso8601String();
                          }
                        } catch (e) {
                          // Handle error
                        }
                      });
                    },
                  ),
                ),
              ],
            ),
            SizedBox(height: 10),
            Row(
              children: [
                Expanded(
                  child: RawKeyboardListener(
                    focusNode: FocusNode(),
                    onKey: (RawKeyEvent event) {
                      if (event is RawKeyDownEvent) {
                        if (event.logicalKey == LogicalKeyboardKey.f2) {
                          searchSale(word: "");
                        }
                      }
                    },
                    child: TextField(
                      onSubmitted: (value) {
                        if (kIsWeb) {}
                      },
                      onChanged: (code) {
                        _debouncer.run(() {
                          if (code.trim().isNotEmpty) {
                            EmployeeRepository()
                                .getEmployeeByCode(code.trim())
                                .then((value) {
                                  if (value.success && value.data != null) {
                                    EmployeeModel emp = EmployeeModel.fromJson(
                                      value.data,
                                    );
                                    saleNameController.text = emp.name;
                                    screenData.salename = emp.name;
                                    screenData.salecode = code;
                                  } else {
                                    saleNameController.text = global.language(
                                      "Employee_not_found",
                                    );
                                    screenData.salename = "";
                                    screenData.salecode = "";
                                  }
                                })
                                .onError((error, stackTrace) {
                                  saleNameController.text = global.language(
                                    "Employee_not_found",
                                  );
                                  screenData.salename = "";
                                  screenData.salecode = "";
                                });
                          } else {
                            screenData.salename = "";
                            screenData.salecode = "";
                            saleNameController.text = global.language(
                              "Employee_not_found",
                            );
                          }
                        });
                        Future.delayed(const Duration(seconds: 3), () {
                          setState(() {});
                        });
                      },
                      textAlign: TextAlign.left,
                      controller: saleCodeController,
                      decoration: InputDecoration(
                        floatingLabelBehavior: FloatingLabelBehavior.always,
                        border: OutlineInputBorder(),
                        labelText: global.language("sale_code"),
                        suffixIcon: Row(
                          mainAxisAlignment:
                              MainAxisAlignment.spaceBetween, // added line
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            IconButton(
                              focusNode: FocusNode(skipTraversal: true),
                              icon: Icon(Icons.search),
                              onPressed: () {
                                searchSale(word: "");
                              },
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                ),
                SizedBox(width: 5),
                Expanded(
                  child: TextField(
                    readOnly: true,
                    focusNode: null,
                    textAlign: TextAlign.left,
                    controller: saleNameController,
                    decoration: InputDecoration(
                      contentPadding: const EdgeInsets.all(10.0),
                      enabledBorder: OutlineInputBorder(
                        borderSide: BorderSide(color: global.theme.dividerBorderColor, width: 0.0),
                      ),
                      floatingLabelBehavior: FloatingLabelBehavior.always,
                      border: OutlineInputBorder(),
                      labelText: global.language("sale_name"),
                    ),
                  ),
                ),
              ],
            ),
            SizedBox(height: 10),
            Container(
              margin: EdgeInsets.only(bottom: 10),
              child: Row(
                children: [
                  (widget.type == global.TransactionTypeEnum.stockbalance)
                      ? Expanded(
                          child: RawKeyboardListener(
                            focusNode: FocusNode(),
                            child: TextField(
                              onSubmitted: (value) {
                                if (kIsWeb) {}
                              },
                              onChanged: (value) {
                                setState(() {
                                  screenData.docrefno = value;
                                });
                              },
                              textAlign: TextAlign.left,
                              controller: docRefNumberController,
                              decoration: InputDecoration(
                                floatingLabelBehavior:
                                    FloatingLabelBehavior.always,
                                border: OutlineInputBorder(),
                                labelText: global.language("doc_ref"),
                              ),
                            ),
                          ),
                        )
                      : Container(),
                  SizedBox(width: 10),
                  Expanded(
                    child: CustomDatePicker(
                      key: ValueKey(screenData.docrefdate),
                      labelText: global.language("doc_ref_date"),
                      initialDate: DateTime.parse(screenData.docrefdate),
                      useBuddhistCalendar:
                          global.profileData.yeartype == "buddhist",
                      onDateSelected: (date) {
                        setState(() {
                          // รวมวันที่ใหม่กับเวลาเดิม
                          DateTime currentDateTime = DateTime.parse(
                            screenData.docrefdate,
                          );
                          DateTime newDateTime = DateTime(
                            date!.year,
                            date.month,
                            date.day,
                            currentDateTime.hour,
                            currentDateTime.minute,
                            currentDateTime.second,
                            currentDateTime.millisecond,
                          );

                          screenData.docrefdate = newDateTime
                              .toLocal()
                              .toIso8601String();

                          // อัพเดท controller สำหรับการแสดงผล
                          if (global.profileData.yeartype == "buddhist") {
                            docRefDateController.text = global.dateTimeBuddhist(
                              newDateTime,
                              format: global.DateTimeFormatEnum.dateDay,
                            );
                          } else {
                            docRefDateController.text = DateFormat(
                              'dd/MM/yyyy',
                            ).format(newDateTime);
                          }
                        });
                      },
                      decoration: InputDecoration(
                        border: OutlineInputBorder(),
                        labelText: global.language("doc_ref_date"),
                      ),
                    ),
                  ),
                ],
              ),
            ),
            SizedBox(height: 10),
            Container(
              margin: EdgeInsets.only(bottom: 10),
              child: Row(
                children: [
                  Expanded(
                    child: TextField(
                      maxLines: 4,
                      decoration: InputDecoration(
                        border: OutlineInputBorder(),
                        labelText: global.language('disciption'),
                      ),
                      controller: descriptionController,
                      onChanged: (value) {
                        setState(() {
                          screenData.description = value;
                        });
                      },
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget editProductListDetailWidget(
    double maxWidth,
    double sumWidth,
    double screenHeight,
  ) {
    return (isShowTable)
        ? Column(
            children: [
              Padding(
                padding: const EdgeInsets.all(8.0),
                child: Row(
                  children: [
                    Expanded(
                      flex: 2,
                      child: Row(
                        children: [
                          /// frist page
                          IconButton(
                            icon: Icon(Icons.first_page),
                            onPressed: pagination.page > 1
                                ? () => fetchData(1, limit, q)
                                : null,
                          ),
                          IconButton(
                            icon: Icon(Icons.chevron_left),
                            onPressed: pagination.page > 1
                                ? () => fetchData(pagination.page - 1, limit, q)
                                : null,
                          ),
                          Text(
                            'Page ${pagination.page} of ${pagination.totalPage}',
                          ),
                          IconButton(
                            icon: Icon(Icons.chevron_right),
                            onPressed: pagination.page < pagination.totalPage
                                ? () => fetchData(pagination.page + 1, limit, q)
                                : null,
                          ),

                          /// last page
                          IconButton(
                            icon: Icon(Icons.last_page),
                            onPressed: pagination.page < pagination.totalPage
                                ? () =>
                                      fetchData(pagination.totalPage, limit, q)
                                : null,
                          ),

                          /// size box width 10
                          SizedBox(width: 10),
                          DropdownButton<int>(
                            value: limit,
                            items: _limitOptions.map<DropdownMenuItem<int>>((
                              int value,
                            ) {
                              return DropdownMenuItem<int>(
                                value: value,
                                child: Text(value.toString()),
                              );
                            }).toList(),
                            onChanged: (int? newValue) {
                              setState(() {
                                limit = newValue!;
                              });
                              fetchData(
                                page,
                                limit,
                                q,
                              ); // Go to the first page with the new rows per page
                            },
                          ),
                        ],
                      ),
                    ),
                    Expanded(
                      flex: 1,
                      child: TextField(
                        decoration: InputDecoration(
                          labelText: global.language('search'),
                          suffixIcon: isLoadingText
                              ? Padding(
                                  padding: EdgeInsets.all(8.0),
                                  child: CircularProgressIndicator(),
                                )
                              : null, // Loading indicator
                          prefixIcon: Icon(Icons.search),
                        ),
                        onChanged: (value) {
                          setState(() {
                            isLoadingText = true;
                            q = value;

                            /// set delay 3 second
                            Duration duration = const Duration(seconds: 2);
                            Timer(duration, () {
                              fetchData(1, limit, value);
                            });
                          });
                        },
                      ),
                    ),
                  ],
                ),
              ),
              SizedBox(
                width: double.infinity,
                height: screenHeight * 0.7,
                child: SingleChildScrollView(
                  child: Column(
                    children: [
                      SizedBox(
                        width: double.infinity,
                        child: DataTable(
                          columns: (screenData.guidfixed!.isEmpty)
                              ? <DataColumn>[
                                  DataColumn(
                                    label: Text(global.language('Barcode')),
                                  ),
                                  DataColumn(
                                    label: Text(global.language('item_name')),
                                  ),
                                  DataColumn(
                                    label: Text(global.language('unit_code')),
                                  ),
                                  DataColumn(
                                    label: Text(
                                      global.language('warehouse_code'),
                                    ),
                                  ),
                                  DataColumn(
                                    label: Text(global.language('shelf_code')),
                                  ),
                                  DataColumn(
                                    label: Text(global.language('qty')),
                                  ),
                                  DataColumn(
                                    label: Text(
                                      global.language('product_cost'),
                                    ),
                                  ),
                                  DataColumn(
                                    label: Text(global.language('amount')),
                                  ),
                                  DataColumn(
                                    label: Text(global.language('edit')),
                                  ),
                                  DataColumn(
                                    label: Text(global.language('delete')),
                                  ),
                                  // Add more columns as needed
                                ]
                              : <DataColumn>[
                                  DataColumn(
                                    label: Text(global.language('Barcode')),
                                  ),
                                  DataColumn(
                                    label: Text(global.language('item_name')),
                                  ),
                                  DataColumn(
                                    label: Text(global.language('unit_code')),
                                  ),
                                  DataColumn(
                                    label: Text(
                                      global.language('warehouse_code'),
                                    ),
                                  ),
                                  DataColumn(
                                    label: Text(global.language('shelf_code')),
                                  ),
                                  DataColumn(
                                    label: Text(global.language('qty')),
                                  ),
                                  DataColumn(
                                    label: Text(
                                      global.language('product_cost'),
                                    ),
                                  ),
                                  DataColumn(
                                    label: Text(global.language('amount')),
                                  ),
                                  // Add more columns as needed
                                ],
                          rows: (screenData.guidfixed!.isEmpty)
                              ? (!isLoadingTable)
                                    ? detailStockBalanceImportModel
                                          .map<DataRow>(
                                            (item) => DataRow(
                                              cells: <DataCell>[
                                                DataCell(
                                                  Text(
                                                    item.barcode!,
                                                    style: (item.isnotexist!)
                                                        ? TextStyle(
                                                            fontWeight:
                                                                FontWeight.bold,
                                                            color: global.theme.negativeHighlightTextColor,
                                                          )
                                                        : TextStyle(
                                                            fontWeight:
                                                                FontWeight
                                                                    .normal,
                                                          ),
                                                  ),
                                                ),
                                                DataCell(
                                                  Text(
                                                    item.name!,
                                                    style: (item.isnotexist!)
                                                        ? TextStyle(
                                                            fontWeight:
                                                                FontWeight.bold,
                                                            color: global.theme.negativeHighlightTextColor,
                                                          )
                                                        : TextStyle(
                                                            fontWeight:
                                                                FontWeight
                                                                    .normal,
                                                          ),
                                                  ),
                                                ),
                                                DataCell(
                                                  Text(
                                                    item.unitcode!,
                                                    style: (item.isnotexist!)
                                                        ? TextStyle(
                                                            fontWeight:
                                                                FontWeight.bold,
                                                            color: global.theme.negativeHighlightTextColor,
                                                          )
                                                        : TextStyle(
                                                            fontWeight:
                                                                FontWeight
                                                                    .normal,
                                                          ),
                                                  ),
                                                ),
                                                DataCell(
                                                  Text(
                                                    item.warehousecode!,
                                                    style: (item.isnotexist!)
                                                        ? TextStyle(
                                                            fontWeight:
                                                                FontWeight.bold,
                                                            color: global.theme.negativeHighlightTextColor,
                                                          )
                                                        : TextStyle(
                                                            fontWeight:
                                                                FontWeight
                                                                    .normal,
                                                          ),
                                                  ),
                                                ),
                                                DataCell(
                                                  Text(
                                                    item.shelfcode!,
                                                    style: (item.isnotexist!)
                                                        ? TextStyle(
                                                            fontWeight:
                                                                FontWeight.bold,
                                                            color: global.theme.negativeHighlightTextColor,
                                                          )
                                                        : TextStyle(
                                                            fontWeight:
                                                                FontWeight
                                                                    .normal,
                                                          ),
                                                  ),
                                                ),
                                                DataCell(
                                                  Align(
                                                    alignment:
                                                        Alignment.centerRight,
                                                    child: Text(
                                                      item.qty.toString(),
                                                      style: (item.isnotexist!)
                                                          ? TextStyle(
                                                              fontWeight:
                                                                  FontWeight
                                                                      .bold,
                                                              color: global.theme.negativeHighlightTextColor,
                                                            )
                                                          : TextStyle(
                                                              fontWeight:
                                                                  FontWeight
                                                                      .normal,
                                                            ),
                                                    ),
                                                  ),
                                                ),
                                                DataCell(
                                                  Align(
                                                    alignment:
                                                        Alignment.centerRight,
                                                    child: Text(
                                                      item.price.toString(),
                                                      style: (item.isnotexist!)
                                                          ? TextStyle(
                                                              fontWeight:
                                                                  FontWeight
                                                                      .bold,
                                                              color: global.theme.negativeHighlightTextColor,
                                                            )
                                                          : TextStyle(
                                                              fontWeight:
                                                                  FontWeight
                                                                      .normal,
                                                            ),
                                                    ),
                                                  ),
                                                ),
                                                DataCell(
                                                  Align(
                                                    alignment:
                                                        Alignment.centerRight,
                                                    child: Text(
                                                      item.sumamount.toString(),
                                                      style: (item.isnotexist!)
                                                          ? TextStyle(
                                                              fontWeight:
                                                                  FontWeight
                                                                      .bold,
                                                              color: global.theme.negativeHighlightTextColor,
                                                            )
                                                          : TextStyle(
                                                              fontWeight:
                                                                  FontWeight
                                                                      .normal,
                                                            ),
                                                    ),
                                                  ),
                                                ),
                                                DataCell(
                                                  Align(
                                                    alignment: Alignment.center,
                                                    child: IconButton(
                                                      icon: Icon(
                                                        Icons.edit,
                                                      ),
                                                      onPressed: () {
                                                        _showBarcodeDialog(
                                                          context,
                                                          item,
                                                        );
                                                      },
                                                    ),
                                                  ),
                                                ),
                                                DataCell(
                                                  Align(
                                                    alignment: Alignment.center,
                                                    child: IconButton(
                                                      color: global.theme.negativeHighlightTextColor,
                                                      icon: Icon(
                                                        Icons.delete,
                                                      ),
                                                      onPressed: () {
                                                        /// show dialog confirm delete
                                                        showDialog(
                                                          context: context,
                                                          builder: (BuildContext context) {
                                                            return AlertDialog(
                                                              title: Text(
                                                                global.language(
                                                                  'confirm_delete',
                                                                ),
                                                              ),
                                                              content: Text(
                                                                global.language(
                                                                  'are_you_sure_you_want_to_delete_this_item',
                                                                ),
                                                              ),
                                                              actions: <Widget>[
                                                                TextButton(
                                                                  onPressed: () {
                                                                    Navigator.of(
                                                                      context,
                                                                    ).pop();
                                                                  },
                                                                  child: Text(
                                                                    global.language(
                                                                      'close',
                                                                    ),
                                                                  ),
                                                                ),
                                                                ElevatedButton(
                                                                  style: ElevatedButton.styleFrom(
                                                                    backgroundColor:
                                                                        Colors
                                                                            .red,
                                                                  ),
                                                                  onPressed: () {
                                                                    Navigator.pop(
                                                                      context,
                                                                    );
                                                                    deleteItemDetailStockBalanceImpart(
                                                                      item.guidfixed!,
                                                                    );
                                                                  },
                                                                  child: Text(
                                                                    global.language(
                                                                      'delete',
                                                                    ),
                                                                  ),
                                                                ),
                                                              ],
                                                            );
                                                          },
                                                        );
                                                      },
                                                    ),
                                                  ),
                                                ),
                                              ],
                                            ),
                                          )
                                          .toList()
                                    : []
                              : (!isLoadingTable)
                              ? detailTransactionModel
                                    .map<DataRow>(
                                      (item) => DataRow(
                                        cells: <DataCell>[
                                          DataCell(Text(item.barcode)),
                                          DataCell(
                                            Text(
                                              global.activeLangName(
                                                item.itemnames!,
                                              ),
                                            ),
                                          ),
                                          DataCell(Text(item.unitcode)),
                                          DataCell(Text(item.whcode)),
                                          DataCell(Text(item.shelfcode)),
                                          DataCell(
                                            Align(
                                              alignment: Alignment.centerRight,
                                              child: Text(item.qty.toString()),
                                            ),
                                          ),
                                          DataCell(
                                            Align(
                                              alignment: Alignment.centerRight,
                                              child: Text(
                                                item.price.toString(),
                                              ),
                                            ),
                                          ),
                                          DataCell(
                                            Align(
                                              alignment: Alignment.centerRight,
                                              child: Text(
                                                item.sumamount.toString(),
                                              ),
                                            ),
                                          ),
                                        ],
                                      ),
                                    )
                                    .toList()
                              : [],
                        ),
                      ),
                      (isLoadingTable)
                          ? Padding(
                              padding: EdgeInsets.only(top: 10.0),
                              child: Column(
                                mainAxisAlignment: MainAxisAlignment.center,
                                crossAxisAlignment: CrossAxisAlignment.center,
                                children: [
                                  SizedBox(
                                    width: 100, // Set the desired width
                                    height: 100, // Set the desired height
                                    child: CircularProgressIndicator(
                                      strokeWidth:
                                          4, // Optional: Set the thickness of the indicator
                                      valueColor: AlwaysStoppedAnimation<Color>(
                                        global.theme.infoHighlightTextColor,
                                      ), // Optional: Set the color
                                    ),
                                  ),
                                ],
                              ),
                            )
                          : Container(),
                    ],
                  ),
                ),
              ),
            ],
          )
        : Container();
  }

  void fetchData(int page, int limit, String q) {
    setState(() {
      isLoadingTable = true;
    });

    page = page;
    limit = limit;
    q = q;

    context.read<StockBalanceBloc>().add(
      LoadStockBalanceImportByTaskid(
        taskid: taskid,
        q: q,
        limit: limit,
        page: page,
      ),
    );
  }

  void deleteItemDetailStockBalanceImpart(String guid) {
    context.read<StockBalanceBloc>().add(DeleteDetailByGuid(guid: guid));
  }

  Future<void> downloadAssetFile(String assetPath, String saveFileName) async {
    // Load the asset file
    final ByteData data = await rootBundle.load(assetPath);
    final Uint8List bytes = data.buffer.asUint8List();

    // Use the download function
    bool success = await downloadAssetFileBytes(bytes, saveFileName);

    if (success) {
      if (kDebugMode) {
        AppLogger.info("File successfully downloaded.");
      }
    } else {
      if (kDebugMode) {
        AppLogger.error("Failed to download file.");
      }
    }
  }

  Widget editProductListWidget() {
    Widget resultWidget;

    resultWidget = wideListScreen();

    return resultWidget;
  }

  Widget wideListScreen() {
    double sumWidth = 0;
    double screenHeight = MediaQuery.of(
      context,
    ).size.height; // Get the screen height

    return Container(
      width: double.infinity,
      padding: const EdgeInsets.only(top: 10),
      child: LayoutBuilder(
        builder: (context, constraints) {
          return Column(
            children: [
              Expanded(
                child: (isLoading)
                    ? Center(
                        child: SizedBox(
                          width: 100, // Set the desired width
                          height: 100, // Set the desired height
                          child: CircularProgressIndicator(
                            strokeWidth:
                                4, // Optional: Set the thickness of the indicator
                            valueColor: AlwaysStoppedAnimation<Color>(
                              global.theme.infoHighlightTextColor,
                            ), // Optional: Set the color
                          ),
                        ),
                      )
                    : (fileImportName.isNotEmpty ||
                          screenData.guidfixed!.isNotEmpty)
                    ? editProductListDetailWidget(
                        constraints.maxWidth,
                        sumWidth,
                        screenHeight,
                      )
                    : Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        crossAxisAlignment: CrossAxisAlignment.center,
                        children: [
                          IconButton(
                            iconSize: 150,
                            color: global.theme.iconSecondaryColor,
                            tooltip: global.language('import_product_balance'),
                            icon: Icon(Icons.upload_file),
                            onPressed: () {
                              uploadFileExcel();
                            },
                          ),
                          Text(
                            global.language(
                              'import_stock_balance_from_excel_file_size_2mb',
                            ),
                            style: TextStyle(
                              fontSize: 20,
                              color: global.theme.iconSecondaryColor,
                            ),
                          ),

                          /// download file excel ex sample stock balance
                          TextButton(
                            onPressed: () {
                              try {
                                downloadAssetFile(
                                  'assets/file_import/import_balance_final.xlsx',
                                  'import_balance.xlsx',
                                );
                              } catch (e) {
                                if (kDebugMode) {
                                  AppLogger.error('An error occurred: $e');
                                }
                                // Handle the error or show a message to the user

                                /// show dialog error
                                showDialog(
                                  context: context,
                                  builder: (BuildContext context) {
                                    return AlertDialog(
                                      title: Text(global.language('error')),
                                      content: Text(e.toString()),
                                      actions: [
                                        TextButton(
                                          child: Text(
                                            global.language('confirm'),
                                          ),
                                          onPressed: () {
                                            Navigator.of(context).pop();
                                          },
                                        ),
                                      ],
                                    );
                                  },
                                );
                              }
                            },
                            child: Text(
                              global.language(
                                'download_file_excel_ex_sample_stock_balance',
                              ),
                              style: TextStyle(fontSize: 20),
                            ),
                          ),
                        ],
                      ),
              ),
              productListButtomBar(),
            ],
          );
        },
      ),
    );
  }

  Widget productListButtomBar() {
    return SizedBox(
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Row(
            children: [
              (fileImportName.isNotEmpty)
                  ? Padding(
                      padding: EdgeInsets.all(8.0),
                      child: ElevatedButton(
                        onPressed: () {
                          _showBarcodeDialog(context, null);
                        },
                        child: Text(global.language('add_new_line_by_barcode')),
                      ),
                    )
                  : Container(),
              Padding(
                padding: const EdgeInsets.all(8.0),
                child: (fileImportName.isEmpty)
                    ? Container()
                    : ElevatedButton.icon(
                        /// button red
                        style: ElevatedButton.styleFrom(
                          foregroundColor: global.theme.onPrimaryColor,
                          backgroundColor: global.theme.negativeHighlightTextColor,
                        ),

                        onPressed: () {
                          discardData(
                            callBack: () {
                              context.read<StockBalanceBloc>().add(
                                DeleteTaskid(taskid: taskid),
                              );
                            },
                          );
                        },
                        icon: Icon(Icons.clear),
                        label: Text(fileImportName),
                      ),
              ),
            ],
          ),

          /// show totalModel
          (fileImportName.isNotEmpty || screenData.guidfixed!.isNotEmpty)
              ? Padding(
                  padding: EdgeInsets.all(8.0),
                  child: Row(
                    children: [
                      Text(
                        '${global.language('total_item')} : ${global.formatNumber((screenData.guidfixed!.isEmpty) ? double.parse(totalModel.totalitem.toString()) : screenData.totalqty!)}',
                        style: TextStyle(
                          fontSize: 20,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      SizedBox(width: 10),
                      Text(
                        '${global.language('total_amount')} : ${global.formatNumber((screenData.guidfixed!.isEmpty) ? double.parse(totalModel.totalamount.toString()) : screenData.totalamount)}',
                        style: TextStyle(
                          fontSize: 20,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ],
                  ),
                )
              : Container(),
        ],
      ),
    );
  }

  Widget editWidget() {
    List<Widget> tabx = [
      Tab(text: global.language("doc_header")),
      Tab(text: global.language("doc_details")),
    ];

    List<Widget> childrenx = [editDocumentWidget(), editProductListWidget()];

    return Scaffold(
      appBar: AppBar(
        backgroundColor: global.theme.surfaceColor,
        automaticallyImplyLeading: false,
        title: TabBar(controller: editTabController, tabs: tabx),
      ),
      body: RawKeyboardListener(
        focusNode: FocusNode(skipTraversal: true),
        child: TabBarView(controller: editTabController, children: childrenx),
      ),
    );
  }

  void deleteDoc() {
    context.read<TransBloc>().add(
      TransDelete(guid: screenData.guidfixed!, type: widget.type),
    );
  }

  void saveOrUpdateData() {
    DateTime docDatetimeUtc = DateTime.parse(screenData.docdatetime);
    screenData.docdatetime = docDatetimeUtc.toUtc().toIso8601String();
    DateTime docRefDatetimeUtc = DateTime.parse(screenData.docrefdate);
    screenData.docrefdate = docRefDatetimeUtc.toUtc().toIso8601String();
    DateTime taxDocDatetimeUtc = DateTime.parse(screenData.taxdocdate);
    screenData.taxdocdate = taxDocDatetimeUtc.toUtc().toIso8601String();

    if (verify()) {
      context.read<StockBalanceBloc>().add(
        SaveTransStockBalance(taskid: taskid, transactionModel: screenData),
      );
    }
  }

  bool verify() {
    List<String> errorList = [];

    if (fileImportName == "") {
      errorList.add(global.language("please_select_a_file"));
    }

    if (errorList.isNotEmpty) {
      showDialog(
        context: context,
        builder: (BuildContext context) {
          return Center(
            child: AlertDialog(
              title: Text(global.language("not_success_save")),
              content: Text(errorList.join(", ")),
              actions: [
                TextButton(
                  child: Text(global.language("confirm")),
                  onPressed: () {
                    Navigator.of(context).pop();
                  },
                ),
              ],
            ),
          );
        },
      );

      return false;
    } else {
      return true;
    }
  }

  Future<void> uploadFileExcel() async {
    double filesize = 0;

    /// select file excel from device
    FilePickerResult? result = await FilePicker.platform.pickFiles(
      type: FileType.custom,
      allowedExtensions: ['xlsx'],
    );
    if (result != null) {
      filesize = result.files.first.size / 1048576;

      /// set filesize 2 decimal
      filesize = double.parse(filesize.toStringAsFixed(2));

      if (filesize <= 2.0) {
        if (mounted) {
          setState(() {
            isLoading = true;
            fileImportName = result.files.first.name;
          });

          final Uint8List file = result.files.first.bytes!;
          context.read<StockBalanceBloc>().add(
            UploadFileExcel(file: file, filename: fileImportName),
          );
        }
      } else {
        setState(() {
          showDialog(
            context: context,
            builder: (BuildContext context) {
              return AlertDialog(
                title: Text(global.language('file_size_exceeds_2_mb')),
                content: RichText(
                  text: TextSpan(
                    children: <TextSpan>[
                      TextSpan(
                        text: global.language(
                          'please_select_a_file_that_does_not_exceed_2_mb',
                        ),
                      ),
                      TextSpan(text: ' '),
                      TextSpan(
                        text: '${global.language('file_size')} : $filesize MB',
                        style: TextStyle(color: global.theme.negativeHighlightTextColor),
                      ),
                    ],
                  ),
                ),
                actions: [
                  TextButton(
                    onPressed: () {
                      Navigator.of(context).pop();
                    },
                    child: Text(global.language('close')),
                  ),
                ],
              );
            },
          );
        });
      }
    }
  }

  void searchSale({required String word}) {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (context) => EmployeeSearchScreen(word: word)),
    ).then((value) {
      EmployeeModel result = value;
      if (result.code.trim().isNotEmpty) {
        setState(() {
          saleCodeController.text = result.code;
          saleNameController.text = result.name;

          screenData.salecode = result.code;
          screenData.salename = result.name;
        });
      }
    });
  }

  void searchTrans() {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) =>
            TransSearchScreen(custcode: '', type: widget.type),
      ),
    ).then((value) {
      TransactionModel result = value;
      if (result.docno.isNotEmpty) {
        setState(() {
          DateTime localDateTime = DateTime.parse(result.docdatetime);
          screenData = result;
          screenData.docdatetime = localDateTime.toUtc().toIso8601String(); // แปลงเป็น UTC ตามกฎ CLAUDE.md
          loadDataToScreen();

          /// call bloc get detail by docno
          if (screenData.guidfixed!.isNotEmpty) {
            isLoading = true;
            context.read<StockBalanceBloc>().add(
              LoadTransStockBalanceDetailByDocno(
                docno: screenData.docno,
                q: q,
                limit: limit,
                page: page,
              ),
            );
          }
        });
      }
    });
  }

  void _showBarcodeDialog(
    BuildContext context,
    StockBalanceImportModel? item,
  ) async {
    TextEditingController barcode = TextEditingController();
    TextEditingController name = TextEditingController();
    TextEditingController unitcode = TextEditingController();
    TextEditingController warehousecode = TextEditingController();
    TextEditingController shelfcode = TextEditingController();
    TextEditingController qty = TextEditingController();
    TextEditingController price = TextEditingController();
    TextEditingController totalamount = TextEditingController();

    ProductBarcodeModel result = ProductBarcodeModel(guidfixed: '');

    barcode.text = "";
    name.text = "";
    unitcode.text = "";
    warehousecode.text = "00000";
    shelfcode.text = "";
    qty.text = "";
    price.text = "";
    totalamount.text = "";

    FocusNode barcodefocus = FocusNode();
    FocusNode namefocus = FocusNode();
    FocusNode unitcodefocus = FocusNode();
    FocusNode warehousecodefocus = FocusNode();
    FocusNode shelfcodefocus = FocusNode();
    FocusNode qtyfocus = FocusNode();
    FocusNode pricefocus = FocusNode();

    if (item != null) {
      barcode.text = item.barcode!;
      name.text = item.name!;
      unitcode.text = item.unitcode!;
      warehousecode.text = item.warehousecode!;
      shelfcode.text = item.shelfcode!;
      qty.text = item.qty.toString();
      price.text = item.price.toString();
      totalamount.text = item.sumamount.toString();
    }
    // totalamount.text = (double.parse(qty.text.trim()) * double.parse(price.text.trim())).toString();

    return showDialog(
      context: context,
      barrierDismissible: true, // Allows tapping outside the dialog to close it
      builder: (BuildContext context) {
        return AlertDialog(
          title: Text(global.language("barcode")),
          content: SizedBox(
            width: (global.isMobileScreen(context)) ? 350 : 500,
            height: 200,
            child: Column(
              children: [
                Row(
                  children: [
                    Expanded(
                      child: TextField(
                        autofocus: true,
                        controller: barcode,
                        focusNode: barcodefocus,
                        decoration: InputDecoration(
                          labelText: '',
                          border: const OutlineInputBorder(),
                          prefixIcon: IconButton(
                            onPressed: () {
                              Navigator.push(
                                context,
                                MaterialPageRoute(
                                  builder: (context) =>
                                      const BarcodeSearchScreen(
                                        word: '',
                                        screen: 'not_material',
                                      ),
                                ),
                              ).then((value) {
                                ProductBarcodeModel result = value;
                                if (result.barcode!.trim().isNotEmpty) {
                                  setState(() {
                                    barcode.text = result.barcode!;
                                    name.text = global.activeLangName(
                                      result.names!,
                                    );
                                    unitcode.text = result.itemunitcode;

                                    Future.delayed(
                                      const Duration(milliseconds: 200),
                                      () {
                                        FocusScope.of(
                                          context,
                                        ).requestFocus(qtyfocus);
                                      },
                                    );
                                  });
                                }
                              });
                            },
                            icon: Icon(Icons.search),
                          ),
                          suffixIcon: IconButton(
                            icon: Icon(Icons.clear),
                            onPressed: () {
                              barcode.text = '';
                              name.text = '';
                              unitcode.text = '';
                              qty.text = "";
                              price.text = "";
                              totalamount.text = "";
                              Future.delayed(
                                const Duration(milliseconds: 200),
                                () {
                                  FocusScope.of(
                                    context,
                                  ).requestFocus(barcodefocus);
                                },
                              );
                              setState(() {});
                            },
                          ),
                        ),
                        onSubmitted: (value) {
                          if (value.trim().isNotEmpty) {
                            String barcodeValue = "";
                            barcodeValue = value.trim();
                            _productBarcodeRepository
                                .getProductBarcodeDetail(barcodeValue)
                                .then((value) {
                                  if (value.success && value.data != null) {
                                    result = ProductBarcodeModel.fromJson(
                                      value.data,
                                    );

                                    if (result.itemtype == 3) {
                                      return;
                                    }
                                    name.text = global.activeLangName(
                                      result.names!,
                                    );
                                    unitcode.text = result.itemunitcode;

                                    Future.delayed(
                                      const Duration(milliseconds: 200),
                                      () {
                                        FocusScope.of(
                                          context,
                                        ).requestFocus(qtyfocus);
                                      },
                                    );
                                  }
                                  setState(() {});
                                })
                                .onError((error, stackTrace) {
                                  barcode.text = '';
                                  Future.delayed(
                                    const Duration(milliseconds: 200),
                                    () {
                                      FocusScope.of(
                                        context,
                                      ).requestFocus(barcodefocus);
                                    },
                                  );
                                  setState(() {});
                                });
                          }
                        },
                      ),
                    ),
                    SizedBox(width: 10),
                    Expanded(
                      child: TextField(
                        controller: name,
                        focusNode: namefocus,
                        decoration: InputDecoration(
                          labelText: global.language("product_name"),
                          border: const OutlineInputBorder(),
                        ),
                      ),
                    ),
                  ],
                ),
                SizedBox(height: 10),
                Row(
                  children: [
                    Expanded(
                      child: TextField(
                        controller: unitcode,
                        focusNode: unitcodefocus,
                        decoration: InputDecoration(
                          labelText: global.language("unit_code"),
                          border: const OutlineInputBorder(),
                        ),
                      ),
                    ),
                    SizedBox(width: 10),
                    Expanded(
                      child: TextField(
                        controller: warehousecode,
                        focusNode: warehousecodefocus,
                        decoration: InputDecoration(
                          labelText: global.language("warehouse_code"),
                          border: const OutlineInputBorder(),
                          suffixIcon: IconButton(
                            onPressed: () {
                              Navigator.push(
                                context,
                                MaterialPageRoute(
                                  builder: (context) =>
                                      const ProductWarehouseSearchScreen(
                                        word: '',
                                      ),
                                ),
                              ).then((value) {
                                SearchGuidCodeNameModel result = value;
                                if (result.code.trim().isNotEmpty) {
                                  setState(() {
                                    warehousecode.text = result.code;
                                    Future.delayed(
                                      const Duration(milliseconds: 200),
                                      () {
                                        FocusScope.of(
                                          context,
                                        ).requestFocus(qtyfocus);
                                      },
                                    );
                                  });
                                }
                              });
                            },
                            icon: Icon(Icons.search),
                          ),
                        ),
                      ),
                    ),
                    SizedBox(width: 10),
                    Expanded(
                      child: TextField(
                        controller: shelfcode,
                        focusNode: shelfcodefocus,
                        decoration: InputDecoration(
                          labelText: global.language("shelf_code"),
                          border: const OutlineInputBorder(),
                          suffixIcon: IconButton(
                            onPressed: () {
                              Navigator.push(
                                context,
                                MaterialPageRoute(
                                  builder: (context) =>
                                      ProductLocationSearchScreen(
                                        whcode: warehousecode.text,
                                      ),
                                ),
                              ).then((value) {
                                SearchGuidCodeNameModel result = value;
                                if (result.code.trim().isNotEmpty) {
                                  setState(() {
                                    shelfcode.text = result.code;
                                    Future.delayed(
                                      const Duration(milliseconds: 200),
                                      () {
                                        FocusScope.of(
                                          context,
                                        ).requestFocus(qtyfocus);
                                      },
                                    );
                                  });
                                }
                              });
                            },
                            icon: Icon(Icons.search),
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
                SizedBox(height: 10),
                Row(
                  children: [
                    Expanded(
                      child: TextField(
                        controller: qty,
                        focusNode: qtyfocus,
                        keyboardType: const TextInputType.numberWithOptions(
                          decimal: true,
                        ),
                        inputFormatters: [global.NumberInputFormatter()],
                        decoration: InputDecoration(
                          labelText: global.language("qty"),
                          border: const OutlineInputBorder(),
                        ),
                        onChanged: (value) {
                          if (value.isNotEmpty) {
                            if (value == '0') {
                              qty.selection = TextSelection.fromPosition(
                                TextPosition(offset: qty.text.length),
                              );
                            } else {
                              final numbercheck = value.replaceAll(',', '');

                              if (price.text.trim().isEmpty) {
                                price.text = '0';
                              }

                              totalamount.text =
                                  (double.parse(numbercheck) *
                                          double.parse((price.text)))
                                      .toString(); // calculate total amount
                              totalamount.text = global.formatNumber(
                                double.parse(totalamount.text),
                              );
                            }
                          } else {
                            totalamount.text = '0';
                          }
                        },
                        onSubmitted: (value) {
                          Future.delayed(const Duration(milliseconds: 200), () {
                            FocusScope.of(context).requestFocus(pricefocus);
                          });
                        },
                      ),
                    ),
                    SizedBox(width: 10),
                    Expanded(
                      child: TextField(
                        controller: price,
                        focusNode: pricefocus,
                        keyboardType: const TextInputType.numberWithOptions(
                          decimal: true,
                        ),
                        inputFormatters: [global.NumberInputFormatter()],
                        decoration: InputDecoration(
                          labelText: global.language("product_cost"),
                          border: const OutlineInputBorder(),
                        ),
                        onChanged: (value) {
                          if (value.isNotEmpty) {
                            if (value == '0') {
                              price.selection = TextSelection.fromPosition(
                                TextPosition(offset: price.text.length),
                              );
                            } else {
                              final numbercheck = value.replaceAll(',', '');

                              if (qty.text.trim().isEmpty) {
                                qty.text = '0';
                              }

                              totalamount.text =
                                  (double.parse(numbercheck) *
                                          double.parse((qty.text)))
                                      .toString(); // calculate total amount
                              totalamount.text = global.formatNumber(
                                double.parse(totalamount.text),
                              );
                            }
                          } else {
                            totalamount.text = '0';
                          }
                        },
                      ),
                    ),
                    SizedBox(width: 10),
                    Expanded(
                      child: TextField(
                        controller: totalamount,
                        decoration: InputDecoration(
                          labelText: global.language("amount"),
                          border: const OutlineInputBorder(),
                        ),
                      ),
                    ),
                  ],
                ),
              ],
            ),
          ),
          actions: <Widget>[
            TextButton(
              child: Text(global.language("close")),
              onPressed: () {
                Navigator.pop(context);
              },
            ),

            /// button Save
            ElevatedButton(
              onPressed: () {
                if (barcode.text.isEmpty) {
                  Future.delayed(const Duration(milliseconds: 200), () {
                    FocusScope.of(context).requestFocus(barcodefocus);
                  });
                  return;
                } else {
                  StockBalanceImportModel detail = StockBalanceImportModel(
                    guidfixed: (item == null) ? '' : item.guidfixed!,
                    barcode: barcode.text,
                    name: name.text,
                    unitcode: unitcode.text,
                    warehousecode: warehousecode.text,
                    shelfcode: shelfcode.text,
                    qty: double.parse(qty.text.replaceAll(',', '')),
                    price: double.parse(price.text.replaceAll(',', '')),
                    sumamount: double.parse(
                      totalamount.text.replaceAll(',', ''),
                    ),
                    taskid: taskid,
                    rownumber: (item == null) ? 0 : item.rownumber!,
                  );
                  if (item == null) {
                    context.read<StockBalanceBloc>().add(
                      AddDetail(stockBalanceImportModel: detail),
                    );
                  } else {
                    context.read<StockBalanceBloc>().add(
                      UpdateDetail(
                        guid: item.guidfixed!,
                        stockBalanceImportModel: detail,
                      ),
                    );
                  }
                  Navigator.pop(context);
                }
              },
              child: Text(global.language("save")),
            ),
          ],
        );
      },
    );
  }

  void discardData({required Function callBack}) {
    if (fileImportName.isNotEmpty || screenData.guidfixed!.isNotEmpty) {
      showDialog<String>(
        context: context,
        builder: (BuildContext context) => AlertDialog(
          title: Text(global.language('data_uploaded')),
          content: Text(
            '${global.language('how_do_you_want_to_discard_the_data')} : $fileImportName ?',
          ),
          actions: <Widget>[
            ElevatedButton(
              style: ElevatedButton.styleFrom(backgroundColor: global.theme.negativeHighlightTextColor),
              onPressed: () => Navigator.pop(context),
              child: Text(global.language('no')),
            ),
            ElevatedButton(
              style: ElevatedButton.styleFrom(backgroundColor: global.theme.infoHighlightTextColor),
              onPressed: () {
                Navigator.pop(context);
                callBack();
              },
              child: Text(global.language('yes')),
            ),
          ],
        ),
      );
    } else {
      callBack();
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: Icon(Icons.arrow_back),
          onPressed: () {
            setState(() {
              if (fileImportName.isNotEmpty) {
                context.read<StockBalanceBloc>().add(
                  DeleteTaskid(taskid: taskid),
                );
              }
              global.gotoMainMenu(context);
            });
          },
        ),
        backgroundColor: global.theme.appBarColor,
        title: Text(global.transactionName(widget.type)),
        actions: <Widget>[
          IconButton(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () async {
              searchTrans();
            },
            icon: Icon(Icons.list_alt, size: 26.0),
          ),
          (screenData.guidfixed == '')
              ? IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () async {
                  final result = await _showAlertConfirmSaveDialog(
                    context,
                    screenData.guidfixed ?? '',
                  );
                  if (result != null && result) {
                    saveOrUpdateData();
                  }
                },
                icon: Icon(Icons.save, size: 26.0),
              )
              : IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () async {
                  final result = await _showAlertConfirmDeleteDialog(
                    context,
                    screenData.docno,
                  );
                  if (result != null && result) {
                    deleteDoc();
                  }
                },
                icon: Icon(Icons.delete, size: 26.0),
              ),
        ],
      ),
      body: LayoutBuilder(
        builder: (context, constraints) {
          return MultiBlocListener(
            listeners: [
              BlocListener<StockBalanceBloc, StockBalanceState>(
                listener: (context, state) async {
                  if (state is UploadFileExcelSuccess) {
                    taskid = state.response.id;
                    context.read<StockBalanceBloc>().add(
                      LoadStockBalanceImportByTaskid(
                        taskid: taskid,
                        q: "",
                        limit: 20,
                        page: 1,
                      ),
                    );
                  }
                  if (state is UploadFileExcelFailed) {
                    setState(() {
                      fileImportName = '';
                      isLoading = false;

                      /// show dialog error
                      showDialog(
                        context: context,
                        builder: (BuildContext context) {
                          return AlertDialog(
                            title: Text(global.language('error')),
                            content: Text(state.message),
                            actions: [
                              TextButton(
                                onPressed: () {
                                  Navigator.of(context).pop();
                                },
                                child: Text(global.language('close')),
                              ),
                            ],
                          );
                        },
                      );
                    });
                  }
                  if (state is LoadStockBalanceImportByTaskidSuccess) {
                    /// call bloc totalitem and totalamount
                    context.read<StockBalanceBloc>().add(
                      LoadTotal(taskid: taskid),
                    );

                    setState(() {
                      pagination = state.pagination;
                      detailStockBalanceImportModel = state.data;

                      isLoading = false;
                      isLoadingText = false;
                      isLoadingTable = false;
                      isShowTable = true;
                    });
                  } else if (state is LoadStockBalanceImportByTaskidFailed) {
                    setState(() {
                      isLoading = false;
                      isLoadingText = false;
                      isLoadingTable = false;
                      isShowTable = false;
                      fileImportName = '';
                    });
                    global.showSnackBar(
                      context,
                      Icon(Icons.error, color: global.theme.onPrimaryColor),
                      global.language(state.message),
                      global.theme.negativeHighlightTextColor,
                    );
                  }

                  if (state is DeleteDetailByGuidSuccess) {
                    setState(() {
                      fetchData(page, limit, q);
                    });
                  }

                  if (state is UpdateDetailSuccess) {
                    context.read<StockBalanceBloc>().add(
                      LoadStockBalanceImportByTaskid(
                        taskid: taskid,
                        q: q,
                        limit: limit,
                        page: page,
                      ),
                    );
                    global.showSnackBar(
                      context,
                      Icon(Icons.save, color: global.theme.onPrimaryColor),
                      global.language("update_success"),
                      global.theme.infoHighlightTextColor,
                    );
                  }

                  if (state is AddDetailSuccess) {
                    context.read<StockBalanceBloc>().add(
                      LoadStockBalanceImportByTaskid(
                        taskid: taskid,
                        q: q,
                        limit: limit,
                        page: page,
                      ),
                    );
                    global.showSnackBar(
                      context,
                      Icon(Icons.save, color: global.theme.onPrimaryColor),
                      global.language("save_success"),
                      global.theme.infoHighlightTextColor,
                    );
                  }

                  if (state is LoadTotalSuccess) {
                    setState(() {
                      screenData.totalamount = double.parse(
                        state.total.totalamount.toString(),
                      );
                      screenData.totalqty = double.parse(
                        state.total.totalitem.toString(),
                      );

                      totalModel = state.total;
                    });
                  } else if (state is LoadTotalFailed) {
                    setState(() {
                      screenData.totalamount = 0;
                      screenData.totalqty = 0;
                    });

                    global.showSnackBar(
                      context,
                      Icon(Icons.error, color: global.theme.onPrimaryColor),
                      global.language(state.message),
                      global.theme.negativeHighlightTextColor,
                    );
                  }

                  if (state is SaveTransStockBalanceSuccess) {
                    clearScreenData();
                    global.showSnackBar(
                      context,
                      Icon(Icons.save, color: global.theme.onPrimaryColor),
                      global.language("save_success"),
                      global.theme.infoHighlightTextColor,
                    );
                  }

                  if (state is SaveTransStockBalanceFailed) {
                    context.read<StockBalanceBloc>().add(
                      LoadStockBalanceImportByTaskid(
                        taskid: taskid,
                        q: "",
                        limit: 20,
                        page: 1,
                      ),
                    );

                    /// show dialog error
                    showDialog(
                      context: context,
                      builder: (BuildContext context) {
                        return AlertDialog(
                          title: Text(global.language('not_success_save')),
                          content: Text(global.language(state.message)),
                          actions: [
                            TextButton(
                              onPressed: () {
                                Navigator.of(context).pop();
                              },
                              child: Text(global.language('close')),
                            ),
                          ],
                        );
                      },
                    );
                  }

                  if (state is DeleteTaskidSuccess) {
                    setState(() {
                      pagination = Pagination(
                        page: 1,
                        perPage: 20,
                        total: 10,
                        totalPage: 1,
                        next: 0,
                        prev: 0,
                      );
                      detailStockBalanceImportModel =
                          <StockBalanceImportModel>[];
                      isLoading = false;
                      isLoadingText = false;
                      isLoadingTable = false;
                      isShowTable = false;
                      fileImportName = '';
                    });
                  }
                  if (state is LoadTransStockBalanceDetailByDocnoSuccess) {
                    setState(() {
                      pagination = state.pagination;
                      detailTransactionModel = state.data;

                      isLoading = false;
                      isLoadingText = false;
                      isLoadingTable = false;
                      isShowTable = true;
                    });
                  }
                },
              ),
              BlocListener<TransBloc, TransState>(
                listener: (context, state) {
                  if (state is TransDeleteSuccess) {
                    clearScreenData();
                    global.showSnackBar(
                      context,
                      Icon(Icons.delete, color: global.theme.onPrimaryColor),
                      global.language("delete_success"),
                      global.theme.infoHighlightTextColor,
                    );
                  }
                },
              ),
            ],
            child: editWidget(),
          );
        },
      ),
    );
  }
}

Future<bool?> _showAlertConfirmDeleteDialog(
  BuildContext context,
  String docno,
) async {
  return showDialog<bool?>(
    context: context,
    builder: (BuildContext context) {
      // Return the AlertDialog.
      return AlertDialog(
        title: Text(global.language("alert")),
        content: Text(global.language("confirm_delete")),
        actions: [
          TextButton(
            child: Text(global.language("cancel")),
            onPressed: () {
              // Do something here.
              Navigator.pop(context, false);
            },
          ),
          TextButton(
            child: Text(global.language("delete")),
            onPressed: () {
              Navigator.pop(context, true);
            },
          ),
        ],
      );
    },
  );
}

Future<bool?> _showAlertConfirmSaveDialog(
  BuildContext context,
  String guid,
) async {
  return showDialog<bool?>(
    context: context,
    builder: (BuildContext context) {
      // Return the AlertDialog.
      return AlertDialog(
        title: Text(global.language("alert")),
        content: (guid == '')
            ? Text(global.language("confirm_save"))
            : Text(global.language("confirm_update")),
        actions: [
          TextButton(
            child: Text(global.language("cancel")),
            onPressed: () {
              // Do something here.
              Navigator.pop(context, false);
            },
          ),
          TextButton(
            child: Text(global.language("save")),
            onPressed: () {
              Navigator.pop(context, true);
            },
          ),
        ],
      );
    },
  );
}
