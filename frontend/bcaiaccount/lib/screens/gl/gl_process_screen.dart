import 'package:smlaicloud/imports_bloc.dart';
import 'package:smlaicloud/model/doc_format_model.dart';
import 'package:smlaicloud/model/journal_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:loading_animation_widget/loading_animation_widget.dart';
import 'package:flutter_rounded_date_picker/flutter_rounded_date_picker.dart';
import 'package:intl/intl.dart';
import 'package:percent_indicator/linear_percent_indicator.dart';
import 'package:smlaicloud/widgets/list_font_size_control.dart';

class GlProcessScreen extends StatefulWidget {
  const GlProcessScreen({super.key});

  @override
  State<GlProcessScreen> createState() => GlProcessScreenState();
}

class GlProcessScreenState extends State<GlProcessScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  List<DocFormatModel> listData = [];
  List<String> guidListChecked = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  bool isDataChange = false;
  late MediaQueryData queryData;
  GlobalKey headerKey = GlobalKey();
  int _hoverIndex = -1;
  bool isEditMode = false;
  bool loadingData = false;
  bool isCheckBoxAll = false;
  bool isProscess = false;
  List<JournalModel> journalData = [];
  final DateTime _selectedDate = DateTime.now();

  List<DefaultDocFormatModel> defaultData = [];
  String fromDate = DateTime.now().toLocal().toIso8601String();
  String toDate = DateTime.now().toLocal().toIso8601String();
  TextEditingController fromDateController = TextEditingController();
  TextEditingController toDateController = TextEditingController();

  @override
  void initState() {
    fromDateController.text = global.dateTimeBuddhist(
      DateTime.now(),
      format: global.DateTimeFormatEnum.dateDay,
    );
    toDateController.text = global.dateTimeBuddhist(
      DateTime.now(),
      format: global.DateTimeFormatEnum.dateDay,
    );
    listScrollController.addListener(onScrollList);
    context.read<DocFormatBloc>().add(const DocFormatLoadDefault());

    super.initState();
  }

  @override
  void dispose() {
    listScrollController.dispose();

    super.dispose();
  }

  void loadDataList(String search) {
    setState(() {
      loadingData = true;
    });
    searchText = search;

    context.read<DocFormatBloc>().add(
      DocFormatLoadList(
        offset: (listData.isEmpty) ? 0 : listData.length,
        limit: global.loadDataPerPage,
        search: search,
      ),
    );
  }

  void onScrollList() {
    if (listScrollController.offset >=
            listScrollController.position.maxScrollExtent &&
        !listScrollController.position.outOfRange) {
      loadDataList(searchText);
    }
  }

  void discardData({required Function callBack}) {
    if (isEditMode && isDataChange) {
      showDialog<String>(
        context: context,
        builder: (BuildContext context) => AlertDialog(
          title: Text(global.language('data_editing')),
          content: Text(global.language('leave_this_screen')),
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

  void selectFromDate(BuildContext context) async {
    late Locale local = const Locale('th', 'TH');
    late EraMode eraMode = EraMode.BUDDHIST_YEAR;

    final DateTime? pickedDate = await showRoundedDatePicker(
      context: context,
      initialDate: DateTime.parse(fromDate),
      firstDate: DateTime(2020),
      lastDate: DateTime(2030),
      locale: local,
      era: eraMode,
      borderRadius: 16,
    );

    if (pickedDate != null && pickedDate != _selectedDate) {
      setState(() {
        DateTime? pickDateTimeFormat = DateTime.parse(
          '${DateFormat('yyyy-MM-dd').format(pickedDate)} ${DateFormat('HH:mm:ss.sss').format(DateTime.now())}',
        );
        fromDateController.text = global.dateTimeBuddhist(
          pickDateTimeFormat,
          format: global.DateTimeFormatEnum.dateDay,
        );
        fromDate = pickDateTimeFormat.toLocal().toIso8601String();
      });
    }
  }

  void selectToDate(BuildContext context) async {
    late Locale local = const Locale('th', 'TH');
    late EraMode eraMode = EraMode.BUDDHIST_YEAR;

    final DateTime? pickedDate = await showRoundedDatePicker(
      context: context,
      initialDate: DateTime.parse(fromDate),
      firstDate: DateTime(2020),
      lastDate: DateTime(2030),
      locale: local,
      era: eraMode,
      borderRadius: 16,
    );

    if (pickedDate != null && pickedDate != _selectedDate) {
      setState(() {
        DateTime? pickDateTimeFormat = DateTime.parse(
          '${DateFormat('yyyy-MM-dd').format(pickedDate)} ${DateFormat('HH:mm:ss.sss').format(DateTime.now())}',
        );
        toDateController.text = global.dateTimeBuddhist(
          pickDateTimeFormat,
          format: global.DateTimeFormatEnum.dateDay,
        );
        toDate = pickDateTimeFormat.toLocal().toIso8601String();
      });
    }
  }

  void loadDataDocument() {
    /// find listdata where ischeck == true
    List<DocFormatModel> listDataChecked = listData
        .where((element) => element.ischeck == true)
        .toList();

    /// convert fromDate to format yyyy-MM-dd
    fromDate = DateFormat('yyyy-MM-dd').format(DateTime.parse(fromDate));

    /// convert toDate to format yyyy-MM-dd
    toDate = DateFormat('yyyy-MM-dd').format(DateTime.parse(toDate));

    if (listDataChecked.isNotEmpty) {
      for (var element in listDataChecked) {
        isProscess = true;
        element.successdocument = 0;

        if (element.module == 'PU') {
          context.read<GlProcessBloc>().add(
            GetPurchaseList(fromDate: fromDate, toDate: toDate),
          );
        }

        if (element.module == 'PT') {
          context.read<GlProcessBloc>().add(
            GetPurchaseReturnList(fromDate: fromDate, toDate: toDate),
          );
        }

        if (element.module == 'SI') {
          context.read<GlProcessBloc>().add(
            GetSaleList(fromDate: fromDate, toDate: toDate),
          );
        }

        if (element.module == 'ST') {
          context.read<GlProcessBloc>().add(
            GetSaleReturnList(fromDate: fromDate, toDate: toDate),
          );
        }

        if (element.module == 'IR') {
          context.read<GlProcessBloc>().add(
            GetStockReturnProductList(fromDate: fromDate, toDate: toDate),
          );
        }

        if (element.module == 'AJ') {
          context.read<GlProcessBloc>().add(
            GetStockAdjustList(fromDate: fromDate, toDate: toDate),
          );
        }

        if (element.module == 'IM') {
          context.read<GlProcessBloc>().add(
            GetStockPickupList(fromDate: fromDate, toDate: toDate),
          );
        }

        if (element.module == 'IF') {
          context.read<GlProcessBloc>().add(
            GetStockReceiveList(fromDate: fromDate, toDate: toDate),
          );
        }

        if (element.module == 'EE') {
          context.read<GlProcessBloc>().add(
            GetPaidList(fromDate: fromDate, toDate: toDate),
          );
        }

        if (element.module == 'DE') {
          context.read<GlProcessBloc>().add(
            GetPayList(fromDate: fromDate, toDate: toDate),
          );
        }
      }
      context.read<GlProcessBloc>().add(
        SaveJournalBulk(journalData: journalData),
      );
    } else {
      global.showSnackBar(
        context,
        Icon(Icons.info, color: global.theme.onPrimaryColor),
        "${global.language("warning")} : ไม่ได้เลือกประเภทเอกสาร",
        global.theme.negativeHighlightTextColor,
      );
    }
  }

  Object? getTransByKey(TransactionModel trans, String key) {
    switch (key) {
      case 'discountword':
        return trans.discountword;
      case 'totalcost':
        return trans.totalcost;
      case 'totalvalue':
        return trans.totalvalue;
      case 'totaldiscount':
        return trans.totaldiscount;
      case 'totalvatvalue':
        return trans.totalvatvalue;
      case 'totalbeforevat':
        return trans.totalbeforevat;
      case 'totalaftervat':
        return trans.totalaftervat;
      case 'totalexceptvat':
        return trans.totalexceptvat;
      case 'totalamount':
        return trans.totalamount;
      case 'cashamount':
        return trans.paymentdetail!.cashamount;

      default:
        throw 0;
    }
  }

  double? getTransPaidPayByKey(TransactionPaidPayModel trans, String key) {
    switch (key) {
      case 'totalpaymentamount':
        return trans.totalpaymentamount;
      case 'totalamount':
        return trans.totalamount;
      case 'totalbalance':
        return trans.totalbalance;
      case 'totalvalue':
        return trans.totalvalue;
      case 'totalvatvalue':
        return trans.totalamount;
      case 'cashamount':
        return trans.paycashamount;

      default:
        throw 0;
    }
  }

  String getNameVattype(int vattype) {
    switch (vattype) {
      case 0:
        return 'vat_exclude';
      case 1:
        return 'vat_include';
      case 2:
        return 'vat_zero';
      default:
        return 'vat_none';
    }
  }

  String getNameInquirytype(int inquirytype, String menu) {
    if (menu == "PT") {
      switch (inquirytype) {
        case 0:
          return 'return_credit_purchaser';
        case 1:
          return 'reduce_credit_purchaser';
        case 2:
          return 'return_cash_purchaser';
        default:
          return 'reduce_cash_purchaser';
      }
    } else if (menu == "ST") {
      switch (inquirytype) {
        case 0:
          return 'return_credit_sale';
        case 1:
          return 'reduce_credit_sale';
        case 2:
          return 'return_cash_sale';
        default:
          return 'reduce_cash_sale';
      }
    }
    return ''; // Default return for unknown menu types
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('gl_process')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: Icon(Icons.arrow_back),
          onPressed: () {
            discardData(
              callBack: () {
                // Navigator.pop(context);
                global.gotoMainMenu(context);
                isEditMode = false;
              },
            );
          },
        ),
        actions: <Widget>[
          ListFontSizeControl(onChanged: () => setState(() {})),
        ],
      ),
      body: Focus(
        focusNode: FocusNode(skipTraversal: true, canRequestFocus: true),
        child: Column(
          children: [
            /// fromdate todate
            Padding(
              padding: EdgeInsets.all(20.0),
              child: Column(
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: TextField(
                          onTap: () {
                            selectFromDate(context);
                          },
                          readOnly: true,
                          decoration: InputDecoration(
                            border: OutlineInputBorder(),
                            labelText: global.language("doc_date"),
                            suffixIcon: Row(
                              mainAxisAlignment:
                                  MainAxisAlignment.spaceBetween, // added line
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                IconButton(
                                  focusNode: FocusNode(skipTraversal: true),
                                  icon: Icon(Icons.calendar_today),
                                  onPressed: () {
                                    selectFromDate(context);
                                  },
                                ),
                              ],
                            ),
                          ),
                          controller: fromDateController,
                          onChanged: (value) {
                            setState(() {
                              try {
                                List<String> valueSplit = value
                                    .replaceAll(".", "/")
                                    .split("/");
                                if (valueSplit.length == 3) {
                                  if (valueSplit[2].length == 2) {
                                    valueSplit[2] = '25${valueSplit[2]}';
                                  }
                                  int year = int.tryParse(valueSplit[2]) ?? 0;
                                  year = year - 543;
                                  int month = int.tryParse(valueSplit[1]) ?? 0;
                                  int day = int.tryParse(valueSplit[0]) ?? 0;
                                  value =
                                      "$year-${month.toString().padLeft(2, '0')}-${day.toString().padLeft(2, '0')}";
                                }
                              } catch (e) {
                                // print(e);
                              }
                            });
                          },
                          onSubmitted: (value) {
                            fromDateController.text = DateFormat(
                              'dd/MM/yyyy',
                            ).format(DateTime.parse(value));
                          },
                        ),
                      ),
                      SizedBox(width: 10),
                      Expanded(
                        child: TextField(
                          onTap: () {
                            selectToDate(context);
                          },
                          readOnly: true,
                          decoration: InputDecoration(
                            border: OutlineInputBorder(),
                            labelText: global.language("to_date"),
                            suffixIcon: Row(
                              mainAxisAlignment:
                                  MainAxisAlignment.spaceBetween, // added line
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                IconButton(
                                  focusNode: FocusNode(skipTraversal: true),
                                  icon: Icon(Icons.calendar_today),
                                  onPressed: () {
                                    selectToDate(context);
                                  },
                                ),
                              ],
                            ),
                          ),
                          controller: toDateController,
                          onChanged: (value) {
                            setState(() {
                              try {
                                List<String> valueSplit = value
                                    .replaceAll(".", "/")
                                    .split("/");
                                if (valueSplit.length == 3) {
                                  if (valueSplit[2].length == 2) {
                                    valueSplit[2] = '25${valueSplit[2]}';
                                  }
                                  int year = int.tryParse(valueSplit[2]) ?? 0;
                                  year = year - 543;
                                  int month = int.tryParse(valueSplit[1]) ?? 0;
                                  int day = int.tryParse(valueSplit[0]) ?? 0;
                                  value =
                                      "$year-${month.toString().padLeft(2, '0')}-${day.toString().padLeft(2, '0')}";
                                }
                              } catch (e) {
                                // print(e);
                              }
                            });
                          },
                          onSubmitted: (value) {
                            toDateController.text = DateFormat(
                              'dd/MM/yyyy',
                            ).format(DateTime.parse(value));
                          },
                        ),
                      ),
                    ],
                  ),
                  const SizedBox(height: 10),
                  SizedBox(
                    height: 40,
                    width: double.infinity,
                    child: ElevatedButton.icon(
                      style: ButtonStyle(
                        backgroundColor: WidgetStateProperty.all<Color>(
                          isProscess
                              ? global.theme.iconSecondaryColor
                              : Colors
                                    .blue, // Adjust colors according to your design
                        ),
                      ),
                      icon: isProscess
                          ? SizedBox(
                              width: 20,
                              height: 20,
                              child: CircularProgressIndicator(
                                strokeWidth: 2,
                                valueColor: AlwaysStoppedAnimation<Color>(
                                  global.theme.onPrimaryColor,
                                ),
                              ),
                            )
                          : Icon(Icons.save),
                      onPressed:
                          (!isProscess) // Disable button during loading
                          ? () {
                              loadDataDocument();
                            }
                          : null,
                      label: Text(global.language('process_gl')),
                    ),
                  ),
                ],
              ),
            ),
            Container(color: global.theme.appBarColor, height: 6),
            Container(
              key: headerKey,
              padding: const EdgeInsets.only(
                left: 10,
                right: 10,
                top: 5,
                bottom: 5,
              ),
              color: global.theme.columnHeaderColor,
              child: Row(
                children: [
                  Expanded(
                    flex: 1,
                    child: Checkbox(
                      value: isCheckBoxAll,
                      onChanged: (bool? value) {
                        setState(() {
                          isCheckBoxAll = value!;
                          for (int i = 0; i < listData.length; i++) {
                            listData[i].ischeck = value;
                          }
                        });
                      },
                    ),
                  ),
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("menu"),
                      style: TextStyle(
                        color: global.theme.columnHeaderTextColor,
                        fontWeight: FontWeight.bold,
                        fontSize: global.deviceConfig.listDataFontSize + 2,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("doc_format_code"),
                      style: TextStyle(
                        color: global.theme.columnHeaderTextColor,
                        fontWeight: FontWeight.bold,
                        fontSize: global.deviceConfig.listDataFontSize + 2,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("desciption"),
                      style: TextStyle(
                        color: global.theme.columnHeaderTextColor,
                        fontWeight: FontWeight.bold,
                        fontSize: global.deviceConfig.listDataFontSize + 2,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 2,
                    child: Text(
                      global.language("total_documnet"),
                      style: TextStyle(
                        color: global.theme.columnHeaderTextColor,
                        fontWeight: FontWeight.bold,
                        fontSize: global.deviceConfig.listDataFontSize + 2,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("total_success"),
                      style: TextStyle(
                        color: global.theme.columnHeaderTextColor,
                        fontWeight: FontWeight.bold,
                        fontSize: global.deviceConfig.listDataFontSize + 2,
                      ),
                    ),
                  ),
                ],
              ),
            ),
            Expanded(
              child: ListView(
                controller: listScrollController,
                children: listData
                    .map((value) => listObject(listData.indexOf(value), value))
                    .toList(),
              ),
            ),
            if (loadingData)
              Center(
                child: LoadingAnimationWidget.staggeredDotsWave(
                  color: global.theme.infoHighlightTextColor,
                  size: 50,
                ),
              ),
          ],
        ),
      ),
    );
  }

  Widget listObject(int index, DocFormatModel value) {
    // ลบการ add key ออก - keys ถูกสร้างใน LoadSuccess แล้ว
    // listKeys.add(GlobalKey());  // ❌ ตรงนี้ทำให้เกิด infinite loop!
    bool selected = selectGuid == value.guidfixed;
    TextStyle textStyle = TextStyle(
      fontWeight: (selected) ? FontWeight.bold : FontWeight.normal,
      fontSize: (selected)
          ? global.deviceConfig.listDataFontSize + 2.0
          : global.deviceConfig.listDataFontSize,
    );

    return MouseRegion(


      cursor: SystemMouseCursors.click,


      onEnter: (_) => setState(() => _hoverIndex = index),


      onExit: (_) => setState(() => _hoverIndex = -1),


      child: GestureDetector(
      child: Container(
        width: double.infinity,
        key: index < listKeys.length ? listKeys[index] : null,
        color: _getContainerColor(value.guidfixed ?? '', index),
        padding: EdgeInsets.only(
          left: 10,
          right: 10,
          top: global.deviceConfig.listDataLineSpace,
          bottom: global.deviceConfig.listDataLineSpace,
        ),
        child: Center(
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.center,
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Expanded(
                flex: 1,
                child: Checkbox(
                  value: value.ischeck,
                  onChanged: (bool? value) {
                    setState(() {
                      isDataChange = true;
                      listData[index].ischeck = value;
                    });
                  },
                ),
              ),
              Expanded(
                flex: 5,
                child: Text(
                  global.language(
                    defaultData
                        .firstWhere(
                          (item) => item.doccode == value.module,
                          orElse: () => DefaultDocFormatModel(
                            name: '',
                            dateformate: '',
                            doccode: '',
                            docnumber: 0,
                          ),
                        )
                        .name,
                  ),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: textStyle,
                ),
              ),
              Expanded(
                flex: 5,
                child: Text(
                  value.doccode,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: textStyle,
                ),
              ),
              Expanded(
                flex: 5,
                child: Text(
                  value.description!,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: textStyle,
                ),
              ),
              Expanded(
                flex: 2,
                child: Text(
                  value.totaldocument!.toString(),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: textStyle,
                ),
              ),
              Expanded(
                flex: 5,
                child: LinearPercentIndicator(
                  width: 150,
                  animation: false,
                  animationDuration: 1000,
                  lineHeight: (global.deviceConfig.listDataFontSize + 8) + 2,
                  percent: value.successdocument!,
                  center: Text(
                    '${(value.successdocument! * 100).toStringAsFixed(2)}%',
                    style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                      fontWeight: FontWeight.bold,
                      fontSize: global.deviceConfig.listDataFontSize + 2,
                    ),
                  ),
                  progressColor: global.theme.positiveHighlightTextColor,
                ),
              ),
            ],
          ),
        ),
      ),
    ),
    );
  }

  Color? _getContainerColor(String itemGuid, int index) {
    if (selectGuid.isNotEmpty && selectGuid == itemGuid) {
      return isEditMode ? global.theme.rowEditColor : global.theme.rowSelectedColor;
    }
    if (_hoverIndex == index) {
      return global.theme.rowHoverColor;
    }
    return (index % 2 == 0)
        ? global.theme.columnAlternateEvenColor
        : global.theme.columnAlternateOddColor;
  }


  @override
  Widget build(BuildContext context) {
    queryData = MediaQuery.of(context);
    listKeys.clear();

    return Scaffold(
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return MultiBlocListener(
            listeners: [
              BlocListener<DocFormatBloc, DocFormatState>(
                listener: (context, state) {
                  if (state is DocFormatLoadDefaultSuccess) {
                    setState(() {
                      /// remove data where doccode == GL
                      state.docFormats.removeWhere(
                        (element) =>
                            element.doccode == 'GL' || element.doccode == 'TF',
                      );
                      defaultData = state.docFormats;

                      loadDataList("");
                    });
                  }
                  // Load
                  if (state is DocFormatLoadListSuccess) {
                    setState(() {
                      loadingData = false;
                      if (state.docFormat.isNotEmpty) {
                        /// remove data state.docFormats where module == GL
                        state.docFormat.removeWhere(
                          (element) =>
                              element.module == 'GL' || element.doccode == 'TF',
                        );
                        listData.addAll(state.docFormat);

                        /// add guidfixed to guidListChecked
                        for (int i = 0; i < listData.length; i++) {
                          if (guidListChecked.contains(listData[i].guidfixed)) {
                            listKeys[i].currentState!.setState(() {
                              selectGuid = listData[i].guidfixed!;
                            });
                          }
                        }
                      }
                    });
                  }
                },
              ),
              BlocListener<GlProcessBloc, GlProcessState>(
                listener: (context, state) {
                  if (state is TransPurchaseLoadSuccess) {
                    setState(() {
                      /// find listdata where module == PU
                      List<DocFormatModel> listDataPu = listData
                          .where((element) => element.module == 'PU')
                          .toList();

                      if (state.trans.isNotEmpty) {
                        for (var element in state.trans) {
                          List<JournalDetailModel> journaldetail = [];

                          for (
                            int i = 0;
                            i < listDataPu[0].details.length;
                            i++
                          ) {
                            dynamic documentFormat = listDataPu[0].details[i];
                            if (documentFormat.debit.isNotEmpty) {
                              journaldetail.add(
                                JournalDetailModel(
                                  accountcode:
                                      documentFormat.accountdebit!.accountcode,
                                  accountname:
                                      documentFormat.accountdebit!.accountname,
                                  debitamount:
                                      (getTransByKey(
                                                element,
                                                documentFormat.actioncode,
                                              )
                                              as num?)
                                          ?.toDouble(),
                                  creditamount: 0,
                                ),
                              );
                            } else {
                              journaldetail.add(
                                JournalDetailModel(
                                  accountcode:
                                      documentFormat.accountcredit!.accountcode,
                                  accountname:
                                      documentFormat.accountcredit!.accountname,
                                  debitamount: 0,
                                  creditamount:
                                      (getTransByKey(
                                                element,
                                                documentFormat.actioncode,
                                              )
                                              as num?)
                                          ?.toDouble(),
                                ),
                              );
                            }
                          }

                          /// remove row journaldetail where debitamount == 0 and creditamount == 0
                          journaldetail.removeWhere(
                            (element) =>
                                element.debitamount == 0 &&
                                element.creditamount == 0,
                          );

                          String dateString = element.docdatetime;
                          DateTime dateTime = DateTime.parse(dateString);
                          int month = dateTime.month;
                          int year = dateTime.year + 543;

                          String accountdescription = "";
                          String menu = "${global.language("gl_purchase")} /";
                          String customer =
                              "${global.language("gl_customer_name")} : ${element.custcode} ${global.activeLangName(element.custnames!)} / ";
                          String vattype =
                              "${global.language("gl_vat_type")} : ${global.language(getNameVattype(element.vattype))} / ";
                          String inquirytype =
                              "${global.language("gl_inquiry_type")} : ${(element.inquirytype == 0) ? global.language("credit") : global.language("cash")} / ";
                          String employee =
                              "${global.language("gl_employee_name")} : ${element.salecode} ${element.salename} / ";
                          String desciption = (element.description!.isNotEmpty)
                              ? "${global.language("gl_remark")} : ${element.description}  "
                              : "";

                          accountdescription =
                              menu +
                              customer +
                              vattype +
                              inquirytype +
                              employee +
                              desciption;

                          journalData.add(
                            JournalModel(
                              accountdescription: accountdescription,
                              accountgroup: listDataPu[0].accountgroup,
                              accountperiod: month,
                              accountyear: year,
                              amount: element.totalamount,
                              batchid: '',
                              bookcode: listDataPu[0].bookcode,
                              docdate: element.docdatetime,
                              docformat: '',
                              docno: element.docno,
                              documentref: '',
                              exdocrefdate: element.docdatetime,
                              exdocrefno: '',
                              journaltype: 0,
                              journaldetail: journaldetail,
                              taxes: [],
                              vats: [],
                            ),
                          );
                        }

                        listDataPu[0].totaldocument = state.trans.length;

                        if (listDataPu[0].details.isEmpty) {
                          setState(() {
                            global.showSnackBar(
                              context,
                              Icon(Icons.save, color: global.theme.onPrimaryColor),
                              "${global.language("not_success_save")} : ไม่ได้กำหนดผังบัญชี : PU",
                              global.theme.negativeHighlightTextColor,
                            );
                          });

                          return;
                        }
                      }
                    });
                  }

                  if (state is TransSaleLoadSuccess) {
                    setState(() {
                      /// find listdata where module == SI
                      List<DocFormatModel> listDataSi = listData
                          .where((element) => element.module == 'SI')
                          .toList();
                      if (state.trans.isNotEmpty) {
                        for (var element in state.trans) {
                          List<JournalDetailModel> journaldetail = [];

                          for (
                            int i = 0;
                            i < listDataSi[0].details.length;
                            i++
                          ) {
                            dynamic documentFormat = listDataSi[0].details[i];
                            if (listDataSi[0].details[i].debit.isNotEmpty) {
                              journaldetail.add(
                                JournalDetailModel(
                                  accountcode:
                                      documentFormat.accountdebit!.accountcode,
                                  accountname:
                                      documentFormat.accountdebit!.accountname,
                                  debitamount:
                                      (getTransByKey(
                                                element,
                                                documentFormat.actioncode,
                                              )
                                              as num?)
                                          ?.toDouble(),
                                  creditamount: 0,
                                ),
                              );
                            } else {
                              journaldetail.add(
                                JournalDetailModel(
                                  accountcode:
                                      documentFormat.accountcredit!.accountcode,
                                  accountname:
                                      documentFormat.accountcredit!.accountname,
                                  debitamount: 0,
                                  creditamount:
                                      (getTransByKey(
                                                element,
                                                documentFormat.actioncode,
                                              )
                                              as num?)
                                          ?.toDouble(),
                                ),
                              );
                            }
                          }

                          /// remove row journaldetail where debitamount == 0 and creditamount == 0
                          journaldetail.removeWhere(
                            (element) =>
                                element.debitamount == 0 &&
                                element.creditamount == 0,
                          );

                          String dateString = element.docdatetime;
                          DateTime dateTime = DateTime.parse(dateString);
                          int month = dateTime.month;
                          int year = dateTime.year + 543;

                          String accountdescription = "";
                          String menu = "${global.language("gl_sale")} /";
                          String customer =
                              "${global.language("gl_customer_name")} : ${element.custcode} ${global.activeLangName(element.custnames!)} / ";
                          String vattype =
                              "${global.language("gl_vat_type")} : ${global.language(getNameVattype(element.vattype))} / ";
                          String inquirytype =
                              "${global.language("gl_inquiry_type")} : ${(element.inquirytype == 0) ? global.language("credit") : global.language("cash")} / ";
                          String employee =
                              "${global.language("gl_employee_name")} : ${element.salecode} ${element.salename} / ";
                          String desciption = (element.description!.isNotEmpty)
                              ? "${global.language("gl_remark")} : ${element.description}  "
                              : "";

                          accountdescription =
                              menu +
                              customer +
                              vattype +
                              inquirytype +
                              employee +
                              desciption;

                          journalData.add(
                            JournalModel(
                              accountdescription: accountdescription,
                              accountgroup: listDataSi[0].accountgroup,
                              accountperiod: month,
                              accountyear: year,
                              amount: element.totalamount,
                              batchid: '',
                              bookcode: listDataSi[0].bookcode,
                              docdate: element.docdatetime,
                              docformat: '',
                              docno: element.docno,
                              documentref: '',
                              exdocrefdate: element.docdatetime,
                              exdocrefno: '',
                              journaltype: 0,
                              journaldetail: journaldetail,
                              taxes: [],
                              vats: [],
                            ),
                          );
                        }

                        listDataSi[0].totaldocument = state.trans.length;
                      }
                    });
                  }

                  if (state is TransPurchaseReturnLoadSuccess) {
                    setState(() {
                      /// find listdata where module == PT
                      List<DocFormatModel> listDataPt = listData
                          .where((element) => element.module == 'PT')
                          .toList();
                      if (state.trans.isNotEmpty) {
                        for (var element in state.trans) {
                          List<JournalDetailModel> journaldetail = [];

                          for (
                            int i = 0;
                            i < listDataPt[0].details.length;
                            i++
                          ) {
                            dynamic documentFormat = listDataPt[0].details[i];
                            if (listDataPt[0].details[i].debit.isNotEmpty) {
                              journaldetail.add(
                                JournalDetailModel(
                                  accountcode:
                                      documentFormat.accountdebit!.accountcode,
                                  accountname:
                                      documentFormat.accountdebit!.accountname,
                                  debitamount:
                                      (getTransByKey(
                                                element,
                                                documentFormat.actioncode,
                                              )
                                              as num?)
                                          ?.toDouble(),
                                  creditamount: 0,
                                ),
                              );
                            } else {
                              journaldetail.add(
                                JournalDetailModel(
                                  accountcode:
                                      documentFormat.accountcredit!.accountcode,
                                  accountname:
                                      documentFormat.accountcredit!.accountname,
                                  debitamount: 0,
                                  creditamount:
                                      (getTransByKey(
                                                element,
                                                documentFormat.actioncode,
                                              )
                                              as num?)
                                          ?.toDouble(),
                                ),
                              );
                            }
                          }

                          /// remove row journaldetail where debitamount == 0 and creditamount == 0
                          journaldetail.removeWhere(
                            (element) =>
                                element.debitamount == 0 &&
                                element.creditamount == 0,
                          );

                          String dateString = element.docdatetime;
                          DateTime dateTime = DateTime.parse(dateString);
                          int month = dateTime.month;
                          int year = dateTime.year + 543;

                          String accountdescription = "";
                          String menu = "${global.language("gl_return_product")} /";
                          String customer =
                              "${global.language("gl_customer_name")} : ${element.custcode} ${global.activeLangName(element.custnames!)} / ";
                          String vattype =
                              "${global.language("gl_vat_type")} : ${global.language(getNameVattype(element.vattype))} / ";
                          String inquirytype =
                              "${global.language("gl_inquiry_type")} : ${global.language(getNameInquirytype(element.inquirytype, 'PT'))}/ ";
                          String employee =
                              "${global.language("gl_employee_name")} : ${element.salecode} ${element.salename} / ";
                          String desciption = (element.description!.isNotEmpty)
                              ? "${global.language("gl_remark")} : ${element.description}  "
                              : "";

                          accountdescription =
                              menu +
                              customer +
                              vattype +
                              inquirytype +
                              employee +
                              desciption;
                          journalData.add(
                            JournalModel(
                              accountdescription: accountdescription,
                              accountgroup: listDataPt[0].accountgroup,
                              accountperiod: month,
                              accountyear: year,
                              amount: element.totalamount,
                              batchid: '',
                              bookcode: listDataPt[0].bookcode,
                              docdate: element.docdatetime,
                              docformat: '',
                              docno: element.docno,
                              documentref: '',
                              exdocrefdate: element.docdatetime,
                              exdocrefno: '',
                              journaltype: 0,
                              journaldetail: journaldetail,
                              taxes: [],
                              vats: [],
                            ),
                          );
                        }

                        listDataPt[0].totaldocument = state.trans.length;
                      }
                    });
                  }

                  if (state is TransSaleReturnLoadSuccess) {
                    setState(() {
                      /// find listdata where module == ST
                      List<DocFormatModel> listDataSt = listData
                          .where((element) => element.module == 'ST')
                          .toList();

                      if (state.trans.isNotEmpty) {
                        for (var element in state.trans) {
                          List<JournalDetailModel> journaldetail = [];

                          for (
                            int i = 0;
                            i < listDataSt[0].details.length;
                            i++
                          ) {
                            dynamic documentFormat = listDataSt[0].details[i];
                            if (listDataSt[0].details[i].debit.isNotEmpty) {
                              journaldetail.add(
                                JournalDetailModel(
                                  accountcode:
                                      documentFormat.accountdebit!.accountcode,
                                  accountname:
                                      documentFormat.accountdebit!.accountname,
                                  debitamount:
                                      (getTransByKey(
                                                element,
                                                documentFormat.actioncode,
                                              )
                                              as num?)
                                          ?.toDouble(),
                                  creditamount: 0,
                                ),
                              );
                            } else {
                              journaldetail.add(
                                JournalDetailModel(
                                  accountcode:
                                      documentFormat.accountcredit!.accountcode,
                                  accountname:
                                      documentFormat.accountcredit!.accountname,
                                  debitamount: 0,
                                  creditamount:
                                      (getTransByKey(
                                                element,
                                                documentFormat.actioncode,
                                              )
                                              as num?)
                                          ?.toDouble(),
                                ),
                              );
                            }
                          }

                          /// remove row journaldetail where debitamount == 0 and creditamount == 0
                          journaldetail.removeWhere(
                            (element) =>
                                element.debitamount == 0 &&
                                element.creditamount == 0,
                          );

                          String dateString = element.docdatetime;
                          DateTime dateTime = DateTime.parse(dateString);
                          int month = dateTime.month;
                          int year = dateTime.year + 543;

                          String accountdescription = "";
                          String menu = "${global.language("gl_receive_return")} /";
                          String customer =
                              "${global.language("gl_customer_name")} : ${element.custcode} ${global.activeLangName(element.custnames!)} / ";
                          String vattype =
                              "${global.language("gl_vat_type")} : ${global.language(getNameVattype(element.vattype))} / ";
                          String inquirytype =
                              "${global.language("gl_inquiry_type")} : ${global.language(getNameInquirytype(element.inquirytype, 'ST'))}/ ";
                          String employee =
                              "${global.language("gl_employee_name")} : ${element.salecode} ${element.salename} / ";
                          String desciption = (element.description!.isNotEmpty)
                              ? "${global.language("gl_remark")} : ${element.description}  "
                              : "";

                          accountdescription =
                              menu +
                              customer +
                              vattype +
                              inquirytype +
                              employee +
                              desciption;

                          journalData.add(
                            JournalModel(
                              accountdescription: accountdescription,
                              accountgroup: listDataSt[0].accountgroup,
                              accountperiod: month,
                              accountyear: year,
                              amount: element.totalamount,
                              batchid: '',
                              bookcode: listDataSt[0].bookcode,
                              docdate: element.docdatetime,
                              docformat: '',
                              docno: element.docno,
                              documentref: '',
                              exdocrefdate: element.docdatetime,
                              exdocrefno: '',
                              journaltype: 0,
                              journaldetail: journaldetail,
                              taxes: [],
                              vats: [],
                            ),
                          );
                        }

                        listDataSt[0].totaldocument = state.trans.length;
                      }
                    });
                  }

                  if (state is TransStockReturnProductLoadSuccess) {
                    setState(() {
                      /// find listdata where module == IR
                      List<DocFormatModel> listDataIr = listData
                          .where((element) => element.module == 'IR')
                          .toList();

                      if (state.trans.isNotEmpty) {
                        for (var element in state.trans) {
                          List<JournalDetailModel> journaldetail = [];

                          for (
                            int i = 0;
                            i < listDataIr[0].details.length;
                            i++
                          ) {
                            dynamic documentFormat = listDataIr[0].details[i];
                            if (listDataIr[0].details[i].debit.isNotEmpty) {
                              journaldetail.add(
                                JournalDetailModel(
                                  accountcode:
                                      documentFormat.accountdebit!.accountcode,
                                  accountname:
                                      documentFormat.accountdebit!.accountname,
                                  debitamount:
                                      (getTransByKey(
                                                element,
                                                documentFormat.actioncode,
                                              )
                                              as num?)
                                          ?.toDouble(),
                                  creditamount: 0,
                                ),
                              );
                            } else {
                              journaldetail.add(
                                JournalDetailModel(
                                  accountcode:
                                      documentFormat.accountcredit!.accountcode,
                                  accountname:
                                      documentFormat.accountcredit!.accountname,
                                  debitamount: 0,
                                  creditamount:
                                      (getTransByKey(
                                                element,
                                                documentFormat.actioncode,
                                              )
                                              as num?)
                                          ?.toDouble(),
                                ),
                              );
                            }
                          }

                          /// remove row journaldetail where debitamount == 0 and creditamount == 0
                          journaldetail.removeWhere(
                            (element) =>
                                element.debitamount == 0 &&
                                element.creditamount == 0,
                          );

                          String dateString = element.docdatetime;
                          DateTime dateTime = DateTime.parse(dateString);
                          int month = dateTime.month;
                          int year = dateTime.year + 543;

                          String accountdescription = "";
                          String menu = "${global.language("gl_receive_from_requisition")} /";
                          String employee =
                              "${global.language("gl_employee_name")} : ${element.salecode} ${element.salename} / ";
                          String desciption = (element.description!.isNotEmpty)
                              ? "${global.language("gl_remark")} : ${element.description}  "
                              : "";

                          accountdescription = menu + employee + desciption;

                          journalData.add(
                            JournalModel(
                              accountdescription: accountdescription,
                              accountgroup: listDataIr[0].accountgroup,
                              accountperiod: month,
                              accountyear: year,
                              amount: element.totalamount,
                              batchid: '',
                              bookcode: listDataIr[0].bookcode,
                              docdate: element.docdatetime,
                              docformat: '',
                              docno: element.docno,
                              documentref: '',
                              exdocrefdate: element.docdatetime,
                              exdocrefno: '',
                              journaltype: 0,
                              journaldetail: journaldetail,
                              taxes: [],
                              vats: [],
                            ),
                          );
                        }

                        listDataIr[0].totaldocument = state.trans.length;
                      }
                    });
                  }

                  if (state is TransStockAdjustLoadSuccess) {
                    setState(() {
                      /// find listdata where module == AJ
                      List<DocFormatModel> listDataAj = listData
                          .where((element) => element.module == 'AJ')
                          .toList();

                      if (state.trans.isNotEmpty) {
                        for (var element in state.trans) {
                          List<JournalDetailModel> journaldetail = [];

                          for (
                            int i = 0;
                            i < listDataAj[0].details.length;
                            i++
                          ) {
                            dynamic documentFormat = listDataAj[0].details[i];
                            if (listDataAj[0].details[i].debit.isNotEmpty) {
                              journaldetail.add(
                                JournalDetailModel(
                                  accountcode:
                                      documentFormat.accountdebit!.accountcode,
                                  accountname:
                                      documentFormat.accountdebit!.accountname,
                                  debitamount:
                                      (getTransByKey(
                                                element,
                                                documentFormat.actioncode,
                                              )
                                              as num?)
                                          ?.toDouble(),
                                  creditamount: 0,
                                ),
                              );
                            } else {
                              journaldetail.add(
                                JournalDetailModel(
                                  accountcode:
                                      documentFormat.accountcredit!.accountcode,
                                  accountname:
                                      documentFormat.accountcredit!.accountname,
                                  debitamount: 0,
                                  creditamount:
                                      (getTransByKey(
                                                element,
                                                documentFormat.actioncode,
                                              )
                                              as num?)
                                          ?.toDouble(),
                                ),
                              );
                            }
                          }

                          /// remove row journaldetail where debitamount == 0 and creditamount == 0
                          journaldetail.removeWhere(
                            (element) =>
                                element.debitamount == 0 &&
                                element.creditamount == 0,
                          );

                          String dateString = element.docdatetime;
                          DateTime dateTime = DateTime.parse(dateString);
                          int month = dateTime.month;
                          int year = dateTime.year + 543;

                          String accountdescription = "";
                          String menu = "${global.language("gl_stock_adjustment")} /";
                          String transflag =
                              "${global.language("gl_adjustment_type")} : ${(element.transflag == 66)
                                  ? global.language("increase")
                                  : (element.transflag == 68)
                                  ? global.language("decrease")
                                  : global.language("adjust_cost")} / ";
                          String employee =
                              "${global.language("gl_employee_name")} : ${element.salecode} ${element.salename} / ";
                          String desciption = (element.description!.isNotEmpty)
                              ? "${global.language("gl_remark")} : ${element.description}  "
                              : "";

                          accountdescription =
                              menu + transflag + employee + desciption;

                          journalData.add(
                            JournalModel(
                              accountdescription: accountdescription,
                              accountgroup: listDataAj[0].accountgroup,
                              accountperiod: month,
                              accountyear: year,
                              amount: element.totalamount,
                              batchid: '',
                              bookcode: listDataAj[0].bookcode,
                              docdate: element.docdatetime,
                              docformat: '',
                              docno: element.docno,
                              documentref: '',
                              exdocrefdate: element.docdatetime,
                              exdocrefno: '',
                              journaltype: 0,
                              journaldetail: journaldetail,
                              taxes: [],
                              vats: [],
                            ),
                          );
                        }

                        listDataAj[0].totaldocument = state.trans.length;
                      }
                    });
                  }

                  if (state is TransStockPickupLoadSuccess) {
                    setState(() {
                      /// find listdata where module == IM
                      List<DocFormatModel> listDataIm = listData
                          .where((element) => element.module == 'IM')
                          .toList();

                      if (state.trans.isNotEmpty) {
                        for (var element in state.trans) {
                          List<JournalDetailModel> journaldetail = [];

                          for (
                            int i = 0;
                            i < listDataIm[0].details.length;
                            i++
                          ) {
                            dynamic documentFormat = listDataIm[0].details[i];
                            if (listDataIm[0].details[i].debit.isNotEmpty) {
                              journaldetail.add(
                                JournalDetailModel(
                                  accountcode:
                                      documentFormat.accountdebit!.accountcode,
                                  accountname:
                                      documentFormat.accountdebit!.accountname,
                                  debitamount:
                                      (getTransByKey(
                                                element,
                                                documentFormat.actioncode,
                                              )
                                              as num?)
                                          ?.toDouble(),
                                  creditamount: 0,
                                ),
                              );
                            } else {
                              journaldetail.add(
                                JournalDetailModel(
                                  accountcode:
                                      documentFormat.accountcredit!.accountcode,
                                  accountname:
                                      documentFormat.accountcredit!.accountname,
                                  debitamount: 0,
                                  creditamount:
                                      (getTransByKey(
                                                element,
                                                documentFormat.actioncode,
                                              )
                                              as num?)
                                          ?.toDouble(),
                                ),
                              );
                            }
                          }

                          /// remove row journaldetail where debitamount == 0 and creditamount == 0
                          journaldetail.removeWhere(
                            (element) =>
                                element.debitamount == 0 &&
                                element.creditamount == 0,
                          );

                          String dateString = element.docdatetime;
                          DateTime dateTime = DateTime.parse(dateString);
                          int month = dateTime.month;
                          int year = dateTime.year + 543;

                          String accountdescription = "";
                          String menu = "${global.language("gl_requisition")} /";
                          String employee =
                              "${global.language("gl_employee_name")} : ${element.salecode} ${element.salename} / ";
                          String desciption = (element.description!.isNotEmpty)
                              ? "${global.language("gl_remark")} : ${element.description}  "
                              : "";

                          accountdescription = menu + employee + desciption;

                          journalData.add(
                            JournalModel(
                              accountdescription: accountdescription,
                              accountgroup: listDataIm[0].accountgroup,
                              accountperiod: month,
                              accountyear: year,
                              amount: element.totalamount,
                              batchid: '',
                              bookcode: listDataIm[0].bookcode,
                              docdate: element.docdatetime,
                              docformat: '',
                              docno: element.docno,
                              documentref: '',
                              exdocrefdate: element.docdatetime,
                              exdocrefno: '',
                              journaltype: 0,
                              journaldetail: journaldetail,
                              taxes: [],
                              vats: [],
                            ),
                          );
                        }

                        listDataIm[0].totaldocument = state.trans.length;
                      }
                    });
                  }

                  if (state is TransStockReceiveLoadSuccess) {
                    setState(() {
                      /// find listdata where module == IF
                      List<DocFormatModel> listDataIf = listData
                          .where((element) => element.module == 'IF')
                          .toList();

                      if (state.trans.isNotEmpty) {
                        for (var element in state.trans) {
                          List<JournalDetailModel> journaldetail = [];

                          for (
                            int i = 0;
                            i < listDataIf[0].details.length;
                            i++
                          ) {
                            dynamic documentFormat = listDataIf[0].details[i];
                            if (listDataIf[0].details[i].debit.isNotEmpty) {
                              journaldetail.add(
                                JournalDetailModel(
                                  accountcode:
                                      documentFormat.accountdebit!.accountcode,
                                  accountname:
                                      documentFormat.accountdebit!.accountname,
                                  debitamount:
                                      (getTransByKey(
                                                element,
                                                documentFormat.actioncode,
                                              )
                                              as num?)
                                          ?.toDouble(),
                                  creditamount: 0,
                                ),
                              );
                            } else {
                              journaldetail.add(
                                JournalDetailModel(
                                  accountcode:
                                      documentFormat.accountcredit!.accountcode,
                                  accountname:
                                      documentFormat.accountcredit!.accountname,
                                  debitamount: 0,
                                  creditamount:
                                      (getTransByKey(
                                                element,
                                                documentFormat.actioncode,
                                              )
                                              as num?)
                                          ?.toDouble(),
                                ),
                              );
                            }
                          }

                          /// remove row journaldetail where debitamount == 0 and creditamount == 0
                          journaldetail.removeWhere(
                            (element) =>
                                element.debitamount == 0 &&
                                element.creditamount == 0,
                          );

                          String dateString = element.docdatetime;
                          DateTime dateTime = DateTime.parse(dateString);
                          int month = dateTime.month;
                          int year = dateTime.year + 543;

                          String accountdescription = "";
                          String menu = "${global.language("gl_receive_product")} /";
                          String employee =
                              "${global.language("gl_employee_name")} : ${element.salecode} ${element.salename} / ";
                          String desciption = (element.description!.isNotEmpty)
                              ? "${global.language("gl_remark")} : ${element.description}  "
                              : "";

                          accountdescription = menu + employee + desciption;

                          journalData.add(
                            JournalModel(
                              accountdescription: accountdescription,
                              accountgroup: listDataIf[0].accountgroup,
                              accountperiod: month,
                              accountyear: year,
                              amount: element.totalamount,
                              batchid: '',
                              bookcode: listDataIf[0].bookcode,
                              docdate: element.docdatetime,
                              docformat: '',
                              docno: element.docno,
                              documentref: '',
                              exdocrefdate: element.docdatetime,
                              exdocrefno: '',
                              journaltype: 0,
                              journaldetail: journaldetail,
                              taxes: [],
                              vats: [],
                            ),
                          );
                        }

                        listDataIf[0].totaldocument = state.trans.length;
                      }
                    });
                  }

                  if (state is TransPaidLoadSuccess) {
                    setState(() {
                      /// find listdata where module == EE
                      List<DocFormatModel> listDataEe = listData
                          .where((element) => element.module == 'EE')
                          .toList();

                      if (state.trans.isNotEmpty) {
                        for (var element in state.trans) {
                          List<JournalDetailModel> journaldetail = [];

                          for (
                            int i = 0;
                            i < listDataEe[0].details.length;
                            i++
                          ) {
                            dynamic documentFormat = listDataEe[0].details[i];
                            if (listDataEe[0].details[i].debit.isNotEmpty) {
                              journaldetail.add(
                                JournalDetailModel(
                                  accountcode:
                                      documentFormat.accountdebit!.accountcode,
                                  accountname:
                                      documentFormat.accountdebit!.accountname,
                                  debitamount: getTransPaidPayByKey(
                                    element,
                                    documentFormat.actioncode,
                                  ),
                                  creditamount: 0,
                                ),
                              );
                            } else {
                              journaldetail.add(
                                JournalDetailModel(
                                  accountcode:
                                      documentFormat.accountcredit!.accountcode,
                                  accountname:
                                      documentFormat.accountcredit!.accountname,
                                  debitamount: 0,
                                  creditamount: getTransPaidPayByKey(
                                    element,
                                    documentFormat.actioncode,
                                  ),
                                ),
                              );
                            }
                          }

                          /// remove row journaldetail where debitamount == 0 and creditamount == 0
                          journaldetail.removeWhere(
                            (element) =>
                                element.debitamount == 0 &&
                                element.creditamount == 0,
                          );

                          String dateString = element.docdatetime;
                          DateTime dateTime = DateTime.parse(dateString);
                          int month = dateTime.month;
                          int year = dateTime.year + 543;

                          String accountdescription = "";
                          String menu = "${global.language("gl_receive_payment")} /";
                          String doctype = (element.doctype == 0)
                              ? global.language("paid_close")
                              : (element.doctype == 1)
                              ? global.language("paid_select")
                              : global.language("paid_input");
                          String customer =
                              "${global.language("gl_customer_name")} : ${element.custcode} ${global.activeLangName(element.custnames!)} / ";
                          String employee =
                              "${global.language("gl_employee_name")} : ${element.salecode} ${element.salename} / ";

                          accountdescription =
                              "$menu$customer$employee ${global.language("gl_payment_type")} : $doctype";

                          journalData.add(
                            JournalModel(
                              accountdescription: accountdescription,
                              accountgroup: listDataEe[0].accountgroup,
                              accountperiod: month,
                              accountyear: year,
                              amount: element.totalamount,
                              batchid: '',
                              bookcode: listDataEe[0].bookcode,
                              docdate: element.docdatetime,
                              docformat: '',
                              docno: element.docno,
                              documentref: '',
                              exdocrefdate: element.docdatetime,
                              exdocrefno: '',
                              journaltype: 0,
                              journaldetail: journaldetail,
                              taxes: [],
                              vats: [],
                            ),
                          );
                        }

                        listDataEe[0].totaldocument = state.trans.length;
                      }
                    });
                  }

                  if (state is TransPayLoadSuccess) {
                    setState(() {
                      /// find listdata where module == DE
                      List<DocFormatModel> listDataDe = listData
                          .where((element) => element.module == 'DE')
                          .toList();

                      if (state.trans.isNotEmpty) {
                        for (var element in state.trans) {
                          List<JournalDetailModel> journaldetail = [];

                          for (
                            int i = 0;
                            i < listDataDe[0].details.length;
                            i++
                          ) {
                            dynamic documentFormat = listDataDe[0].details[i];
                            if (listDataDe[0].details[i].debit.isNotEmpty) {
                              journaldetail.add(
                                JournalDetailModel(
                                  accountcode:
                                      documentFormat.accountdebit!.accountcode,
                                  accountname:
                                      documentFormat.accountdebit!.accountname,
                                  debitamount: getTransPaidPayByKey(
                                    element,
                                    documentFormat.actioncode,
                                  ),
                                  creditamount: 0,
                                ),
                              );
                            } else {
                              journaldetail.add(
                                JournalDetailModel(
                                  accountcode:
                                      documentFormat.accountcredit!.accountcode,
                                  accountname:
                                      documentFormat.accountcredit!.accountname,
                                  debitamount: 0,
                                  creditamount: getTransPaidPayByKey(
                                    element,
                                    documentFormat.actioncode,
                                  ),
                                ),
                              );
                            }
                          }

                          /// remove row journaldetail where debitamount == 0 and creditamount == 0
                          journaldetail.removeWhere(
                            (element) =>
                                element.debitamount == 0 &&
                                element.creditamount == 0,
                          );

                          String dateString = element.docdatetime;
                          DateTime dateTime = DateTime.parse(dateString);
                          int month = dateTime.month;
                          int year = dateTime.year + 543;

                          String accountdescription = "";
                          String menu = "${global.language("gl_make_payment")} /";
                          String doctype = (element.doctype == 0)
                              ? global.language("paid_close")
                              : (element.doctype == 1)
                              ? global.language("paid_select")
                              : global.language("paid_input");
                          String customer =
                              "${global.language("gl_creditor_name")} : ${element.custcode} ${global.activeLangName(element.custnames!)} / ";
                          String employee =
                              "${global.language("gl_employee_name")} : ${element.salecode} ${element.salename} / ";

                          accountdescription =
                              "$menu$customer$employee ${global.language("gl_payment_type")} : $doctype";

                          journalData.add(
                            JournalModel(
                              accountdescription: accountdescription,
                              accountgroup: listDataDe[0].accountgroup,
                              accountperiod: month,
                              accountyear: year,
                              amount: element.totalamount,
                              batchid: '',
                              bookcode: listDataDe[0].bookcode,
                              docdate: element.docdatetime,
                              docformat: '',
                              docno: element.docno,
                              documentref: '',
                              exdocrefdate: element.docdatetime,
                              exdocrefno: '',
                              journaltype: 0,
                              journaldetail: journaldetail,
                              taxes: [],
                              vats: [],
                            ),
                          );
                        }

                        listDataDe[0].totaldocument = state.trans.length;
                      }
                    });
                  }

                  if (state is SaveJournalBulkSuccess) {
                    List<String> docsuccess = [];
                    List<String> docfail = [];
                    setState(() {
                      List<DocFormatModel> listDataPu = listData
                          .where(
                            (element) =>
                                element.module == 'PU' &&
                                element.ischeck == true,
                          )
                          .toList();
                      if (listDataPu.isNotEmpty &&
                          listDataPu[0].details.isNotEmpty) {
                        docsuccess.add('PU');
                        listDataPu[0].successdocument = 1;
                      } else {
                        docfail.add('PU');
                      }

                      List<DocFormatModel> listDataSi = listData
                          .where(
                            (element) =>
                                element.module == 'SI' &&
                                element.ischeck == true,
                          )
                          .toList();
                      if (listDataSi.isNotEmpty &&
                          listDataSi[0].details.isNotEmpty) {
                        docsuccess.add('SI');
                        listDataSi[0].successdocument = 1;
                      } else {
                        docfail.add('SI');
                      }

                      List<DocFormatModel> listDataIr = listData
                          .where(
                            (element) =>
                                element.module == 'IR' &&
                                element.ischeck == true,
                          )
                          .toList();
                      if (listDataIr.isNotEmpty &&
                          listDataIr[0].details.isNotEmpty) {
                        docsuccess.add('IR');
                        listDataIr[0].successdocument = 1;
                      } else {
                        docfail.add('IR');
                      }

                      List<DocFormatModel> listDataPt = listData
                          .where(
                            (element) =>
                                element.module == 'PT' &&
                                element.ischeck == true,
                          )
                          .toList();
                      if (listDataPt.isNotEmpty &&
                          listDataPt[0].details.isNotEmpty) {
                        docsuccess.add('PT');
                        listDataPt[0].successdocument = 1;
                      } else {
                        docfail.add('PT');
                      }

                      List<DocFormatModel> listDataSt = listData
                          .where(
                            (element) =>
                                element.module == 'ST' &&
                                element.ischeck == true,
                          )
                          .toList();
                      if (listDataSt.isNotEmpty &&
                          listDataSt[0].details.isNotEmpty) {
                        docsuccess.add('ST');
                        listDataSt[0].successdocument = 1;
                      } else {
                        docfail.add('ST');
                      }

                      List<DocFormatModel> listDataAj = listData
                          .where(
                            (element) =>
                                element.module == 'AJ' &&
                                element.ischeck == true,
                          )
                          .toList();
                      if (listDataAj.isNotEmpty &&
                          listDataAj[0].details.isNotEmpty) {
                        docsuccess.add('AJ');
                        listDataAj[0].successdocument = 1;
                      } else {
                        docfail.add('AJ');
                      }

                      List<DocFormatModel> listDataIm = listData
                          .where(
                            (element) =>
                                element.module == 'IM' &&
                                element.ischeck == true,
                          )
                          .toList();
                      if (listDataIm.isNotEmpty &&
                          listDataIm[0].details.isNotEmpty) {
                        docsuccess.add('IM');
                        listDataIm[0].successdocument = 1;
                      } else {}

                      List<DocFormatModel> listDataIf = listData
                          .where(
                            (element) =>
                                element.module == 'IF' &&
                                element.ischeck == true,
                          )
                          .toList();
                      if (listDataIf.isNotEmpty &&
                          listDataIf[0].details.isNotEmpty) {
                        docsuccess.add('IF');
                        listDataIf[0].successdocument = 1;
                      } else {
                        docfail.add('IF');
                      }

                      List<DocFormatModel> listDataEe = listData
                          .where(
                            (element) =>
                                element.module == 'EE' &&
                                element.ischeck == true,
                          )
                          .toList();
                      if (listDataEe.isNotEmpty &&
                          listDataEe[0].details.isNotEmpty) {
                        docsuccess.add('EE');
                        listDataEe[0].successdocument = 1;
                      } else {
                        docfail.add('EE');
                      }

                      List<DocFormatModel> listDataDe = listData
                          .where(
                            (element) =>
                                element.module == 'DE' &&
                                element.ischeck == true,
                          )
                          .toList();
                      if (listDataDe.isNotEmpty &&
                          listDataDe[0].details.isNotEmpty) {
                        docsuccess.add('DE');
                        listDataDe[0].successdocument = 1;
                      } else {
                        docfail.add('DE');
                      }

                      Future.delayed( Duration(seconds: 1), () {
                        setState(() {
                          isProscess = false;
                        });
                        if (docsuccess.isNotEmpty) {
                          global.showSnackBar(
                            context,
                            Icon(Icons.save, color: global.theme.onPrimaryColor),
                            "${global.language("save_success")} : ${docsuccess.join(', ')}",
                            global.theme.infoHighlightTextColor,
                          );
                        }

                        if (docfail.isNotEmpty) {
                          global.showSnackBar(
                            context,
                            Icon(Icons.save, color: global.theme.onPrimaryColor),
                            "${global.language("not_success_save")} : ไม่ได้กำหนดผังบัญชี : ${docfail.join(', ')}",
                            global.theme.negativeHighlightTextColor,
                          );
                        }
                      });
                    });
                  }
                  if (state is SaveJournalBulkFailed) {
                    setState(() {
                      isProscess = false;

                      global.showSnackBar(
                        context,
                        Icon(Icons.save, color: global.theme.onPrimaryColor),
                        "${global.language("not_success_save")} : ${state.message}",
                        global.theme.negativeHighlightTextColor,
                      );
                      return;
                    });
                  }
                },
              ),
            ],
            child: listScreen(mobileScreen: false),
          );
        },
      ),
    );
  }
}
