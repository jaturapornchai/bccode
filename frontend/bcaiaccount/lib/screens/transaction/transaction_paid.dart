import 'dart:convert';

import 'package:smlaicloud/bloc/transaction_paidpay/transaction_paidpay_bloc.dart';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/employee_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/pdfgen/pdfpreview.dart';
import 'package:smlaicloud/screen_search/company_branch_search_screen.dart';
import 'package:smlaicloud/screen_search/customer_search_screen.dart';
import 'package:smlaicloud/screen_search/employee_search_screen.dart';
import 'package:smlaicloud/screen_search/supplier_search_screen.dart';
import 'package:smlaicloud/screen_search/transaction_paidpay_search_screen.dart';
import 'package:smlaicloud/screens/transaction/paidpayment_screen.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import 'package:smlaicloud/global.dart' as global;
// ignore: depend_on_referenced_packages
import 'package:intl/intl.dart';
import 'package:smlaicloud/utils/date_picker.dart';

class TransactionPaidScreen extends StatefulWidget {
  final global.TransactionTypeEnum type;

  const TransactionPaidScreen({super.key, required this.type});

  @override
  State<TransactionPaidScreen> createState() => TransactionPaidScreenState();
}

class TransactionPaidScreenState extends State<TransactionPaidScreen>
    with TickerProviderStateMixin {
  final _debouncer = global.Debouncer(1000);
  TextEditingController docDateController = TextEditingController();
  TextEditingController docNumberController = TextEditingController();
  TextEditingController docTimeController = TextEditingController();

  late TransactionPaidPayModel screenData;
  bool docDateTimeValidated = false;
  int paidType = 0;
  void setSystemLanguageList() async {
    clearScreenData();
    await global.setSystemLanguage(context);
  }

  @override
  void initState() {
    super.initState();
    setSystemLanguageList();
  }

  @override
  void dispose() {
    docNumberController.dispose();
    docDateController.dispose();
    docTimeController.dispose();

    super.dispose();
  }

  /// funtion  fetchScreenData

  void calTotalValue() {
    double totalbalance = 0.0;
    double totalpaymentamount = 0.0;
    double totalvalue = 0.0;

    double totalamounttemp = screenData.totalpaymentamount;
    double totalamounthold = screenData.totalpaymentamount;

    screenData.totalamount =
        screenData.paycashamount! +
        screenData.summoneytransfer! +
        screenData.sumcheque! +
        screenData.sumcreditcard! +
        screenData.sumcoupon! +
        screenData.sumcredit! +
        screenData.sumqrcode!;
    for (var data in screenData.details!) {
      if (screenData.doctype == 0) {
        data.selected = true;
      }
      if (screenData.doctype == 0) {
        data.paymentamount = data.value;
        if (data.selected!) {
          totalpaymentamount += data.paymentamount;
        }
      }

      if (screenData.doctype == 1) {
        if (data.selected!) {
          totalpaymentamount += data.paymentamount;
        }
      }

      totalvalue += data.value;
      totalbalance += data.balance;
    }
    screenData.totalvalue = totalvalue;

    screenData.totalbalance = totalbalance;
    screenData.totalpaymentamount = totalpaymentamount;

    if (screenData.doctype == 2) {
      for (var data in screenData.details!) {
        totalpaymentamount += data.paymentamount;
        if ((totalamounttemp - data.value) > -1) {
          data.selected = true;
          data.paymentamount = data.value;
        } else if ((data.value - totalamounttemp) > -1) {
          if (totalamounttemp > 0) {
            data.selected = true;
            data.paymentamount = totalamounttemp;
          } else {
            data.selected = false;
            data.paymentamount = data.value;
          }
        } else {
          data.selected = false;
          data.paymentamount = data.value;
        }
        totalamounttemp -= data.value;
      }
      screenData.totalpaymentamount = totalamounthold;
    }

    setState(() {});
  }

  void deleteDoc() {
    context.read<TransactionPaidPayBloc>().add(
      TransactionPaidPayDelete(guid: screenData.guidfixed!, type: widget.type),
    );
  }

  void clearScreenData() {
    screenData = TransactionPaidPayModel(
      docno: '', // mainapi จะสร้าง docno เอง
      docdatetime: DateTime.now().toUtc().toIso8601String(),
      doctype: 0,
      custcode: '',
      salecode: '',
      salename: '',
      totalpaymentamount: 0,
      totalvalue: 0,
      totalamount: 0,
      totalbalance: 0,
      paycashamount: 0,
      details: [],
      transflag: 50,
      custnames: [],
      paymentdetailraw: '',
      paymentStructs: [],
      sumqrcode: 0,
      sumcreditcard: 0,
      summoneytransfer: 0,
      sumcheque: 0,
      sumcoupon: 0,
      branch: BranchModel(
        guidfixed: global.companyBranchSelectData.guidfixed,
        code: global.companyBranchSelectData.code,
        names: global.companyBranchSelectData.names,
      ),
    );

    DateTime docDateTimeFormat = DateTime.parse(screenData.docdatetime);
    docTimeController.text = DateFormat(
      'HH:mm',
    ).format(docDateTimeFormat.toLocal());
    docDateController.text = DateFormat(
      'dd/MM/yyyy',
    ).format(DateTime.parse(screenData.docdatetime));
    docNumberController.text = screenData.docno;

    if (global.profileData.yeartype == "buddhist") {
      docDateController.text = global.dateTimeBuddhist(
        docDateTimeFormat,
        format: global.DateTimeFormatEnum.dateDay,
      );
    } else {
      docDateController.text = DateFormat(
        'dd/MM/yyyy',
      ).format(docDateTimeFormat);
    }

    setState(() {});
  }

  void saveOrUpdateData() {
    if (verifyPayment()) {
      DateTime docDatetimeUtc = DateTime.parse(screenData.docdatetime);
      screenData.docdatetime = docDatetimeUtc.toUtc().toIso8601String();

      for (var data in screenData.details!) {
        DateTime docdatetimeDetailUtc = DateTime.parse(data.docdatetime);
        data.docdatetime = docdatetimeDetailUtc.toUtc().toIso8601String();
      }

      if (widget.type == global.TransactionTypeEnum.paid) {
        screenData.transflag = 50;
      } else if (widget.type == global.TransactionTypeEnum.pay) {
        screenData.transflag = 51;
      }

      screenData.paymentdetailraw = jsonEncode(
        screenData.paymentStructs,
      );

      if (screenData.guidfixed != '') {
        context.read<TransactionPaidPayBloc>().add(
          TransactionPaidPayUpdate(
            guid: screenData.guidfixed!,
            transactionPaidPay: screenData,
            type: widget.type,
          ),
        );
      } else {
        context.read<TransactionPaidPayBloc>().add(
          TransactionPaidPaySave(
            transactionPaidPay: screenData,
            type: widget.type,
          ),
        );
      }
    }
  }

  bool verifyPayment() {
    List<String> errorList = [];

    if (screenData.totalamount > screenData.totalpaymentamount) {
      errorList.add(global.language("payment_over"));
    } else if (screenData.totalamount < screenData.totalpaymentamount) {
      errorList.add(global.language("payment_less"));
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

  @override
  Widget build(BuildContext context) {
    return MultiBlocListener(
      listeners: [
        BlocListener<TransactionPaidPayBloc, TransactionPaidPayState>(
          listener: (context, state) {
            if (state is TransactionPaidPaySaveSuccess) {
              showDialog(
                context: context,
                builder: (BuildContext context) {
                  return AlertDialog(
                    title: Text(global.language("save_success")),
                    content: Text(
                      "${global.language("docno")} : ${state.docno}",
                    ),
                    actions: [
                      TextButton(
                        child: Text(global.language("confirm")),
                        onPressed: () {
                          clearScreenData();
                          Navigator.of(context).pop();
                        },
                      ),
                    ],
                  );
                },
              );
            } else if (state is TransactionPaidPaySaveFailed) {
              clearScreenData();
              global.showSnackBar(
                context,
                Icon(Icons.save, color: Colors.white),
                global.language(state.message),
                Colors.red,
              );
            } else if (state is TransactionPaidPayDeleteSuccess) {
              clearScreenData();
              global.showSnackBar(
                context,
                Icon(Icons.save, color: Colors.white),
                global.language("delete_success"),
                Colors.blue,
              );
            }

            if (state is TransactionPaidPayUpdateSuccess) {
              showDialog(
                context: context,
                builder: (BuildContext context) {
                  return AlertDialog(
                    title: Text(global.language("update_success")),
                    content: Text(
                      "${global.language("docno")} : ${screenData.docno}",
                    ),
                    actions: [
                      TextButton(
                        child: Text(global.language("confirm")),
                        onPressed: () {
                          clearScreenData();
                          Navigator.of(context).pop();
                        },
                      ),
                    ],
                  );
                },
              );
            } else if (state is TransactionPaidPayUpdateFailed) {
              clearScreenData();
              global.showSnackBar(
                context,
                Icon(Icons.save, color: Colors.white),
                global.language(state.message),
                Colors.red,
              );
            }

            if (state is GetCustcodeTransactionSuccess) {
              setState(() {
                if (state.getCustcodeTransationModel.isNotEmpty) {
                  screenData.details!.clear();
                  for (var element in state.getCustcodeTransationModel) {
                    double balanceCal =
                        double.parse(element.totalamount.toString()) -
                        double.parse(element.paidamount.toString());

                    screenData.details!.add(
                      TransactionPaidPayDetailModel(
                        value: double.parse(element.totalamount.toString()),
                        balance: balanceCal,
                        paymentamount: balanceCal,
                        docno: element.docno!,
                        docdatetime: element.docdate.toString(),
                        transflag: element.transflag8!,
                      ),
                    );
                  }
                  calTotalValue();
                }
              });
            } else if (state is GetCustcodeTransactionFailed) {
              global.showSnackBar(
                context,
                Icon(Icons.save, color: Colors.white),
                global.language(state.message),
                Colors.red,
              );
            }
          },
        ),
      ],
      child: Scaffold(
        appBar: AppBar(
          backgroundColor: global.theme.appBarColor,
          title: Text(global.transactionName(widget.type)),
          leading: IconButton(
            focusNode: FocusNode(skipTraversal: true),
            icon: const Icon(Icons.arrow_back),
            onPressed: () {
              Navigator.pop(context);
            },
          ),
          actions: <Widget>[
            (screenData.guidfixed != '')
                ? Padding(
                    padding: EdgeInsets.only(right: 20.0),
                    child: IconButton(
                      focusNode: FocusNode(skipTraversal: true),
                      onPressed: () async {
                        setState(() {});
                        if (screenData.details!.isEmpty) {
                          showDialog(
                            context: context,
                            builder: (BuildContext context) {
                              return AlertDialog(
                                title: Text(global.language("add_payment_entry")),
                                actions: [
                                  TextButton(
                                    child: Text(global.language("confirm")),
                                    onPressed: () {
                                      clearScreenData();
                                      Navigator.of(context).pop();
                                    },
                                  ),
                                ],
                              );
                            },
                          );
                        } else {
                          Navigator.push(
                            context,
                            MaterialPageRoute(
                              builder: (context) => PdfPreviewPage(
                                screenDataPayPaid: screenData,
                                type: widget.type,
                              ),
                            ),
                          );
                        }
                      },
                      icon: const Icon(Icons.print, size: 26.0),
                    ),
                  )
                : Container(),
            Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () async {
                  searchTrans();
                },
                icon: const Icon(Icons.list_alt, size: 26.0),
              ),
            ),
            (screenData.guidfixed != '')
                ? Padding(
                    padding: const EdgeInsets.only(right: 20.0),
                    child: IconButton(
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
                      icon: const Icon(Icons.delete, size: 26.0),
                    ),
                  )
                : Container(),
            Padding(
              padding: EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () async {
                  if (screenData.details!.isNotEmpty) {
                    saveOrUpdateData();
                  } else {
                    /// show dialog please payment
                    showDialog(
                      context: context,
                      builder: (BuildContext context) {
                        return Center(
                          child: AlertDialog(
                            title: Text(global.language("alert")),
                            content: Text(global.language("please_payment")),
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
                  }
                },
                icon: const Icon(Icons.save, size: 26.0),
              ),
            ),
          ],
        ),
        body: Card(
          child: Container(
            margin: EdgeInsets.all(10),
            child: Column(
              children: [
                Container(
                  margin: EdgeInsets.only(top: 10, bottom: 10),
                  child: Row(
                    children: [
                      Expanded(
                        child: Row(
                          children: [
                            Expanded(
                              child: CustomDatePicker(
                                key: ValueKey(
                                  screenData.docdatetime,
                                ), // เพิ่ม key เพื่อบังคับให้ rebuild เมื่อค่าเปลี่ยน
                                labelText: global.language("doc_date"),
                                initialDate: DateTime.parse(
                                  screenData.docdatetime,
                                ),
                                // useBuddhistCalendar และ languageCode จะอ่านจากสาขาอัตโนมัติ
                                onDateSelected: (date) {
                                  if (date != null) {
                                    setState(() {
                                      // กำหนดเวลาจากวันที่ที่เลือกโดยรักษาเวลาเดิม
                                      final currentTime = DateTime.parse(
                                        screenData.docdatetime,
                                      );
                                      final combinedDateTime = DateTime(
                                        date.year,
                                        date.month,
                                        date.day,
                                        currentTime.hour,
                                        currentTime.minute,
                                        currentTime.second,
                                        currentTime.millisecond,
                                      );

                                      screenData.docdatetime = combinedDateTime
                                          .toLocal()
                                          .toIso8601String();

                                      docTimeController.text = DateFormat(
                                        'HH:mm',
                                      ).format(combinedDateTime);
                                    });
                                  }
                                },
                                decoration: InputDecoration(
                                  border: OutlineInputBorder(),
                                  labelText: global.language("doc_date"),
                                  suffixIcon: IconButton(
                                    focusNode: FocusNode(skipTraversal: true),
                                    icon: const Icon(Icons.calendar_today),
                                    onPressed: () {},
                                  ),
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                      SizedBox(width: 5),
                      Expanded(
                        child: TextField(
                          readOnly: true,
                          decoration: InputDecoration(
                            border: OutlineInputBorder(),
                            labelText: global.language('branch'),
                            suffixIcon: Row(
                              mainAxisAlignment:
                                  MainAxisAlignment.spaceBetween, // added line
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                IconButton(
                                  focusNode: FocusNode(skipTraversal: true),
                                  icon: const Icon(Icons.search),
                                  onPressed: () {
                                    Navigator.push(
                                      context,
                                      MaterialPageRoute(
                                        builder: (context) =>
                                            const CompanyBranchSearchScreen(
                                              word: "",
                                            ),
                                      ),
                                    ).then((value) {
                                      setState(() {
                                        SearchGuidCodeNameModel result = value;
                                        if (result.isCancel == false) {
                                          screenData.branch!.guidfixed =
                                              result.guid;
                                          screenData.branch!.code = result.code;
                                          screenData.branch!.names =
                                              result.names;
                                        }
                                      });
                                    });
                                  },
                                ),
                              ],
                            ),
                          ),
                          controller: TextEditingController(
                            text:
                                "${screenData.branch!.code!} ~ ${global.activeLangName(screenData.branch!.names!)}",
                          ),
                        ),
                      ),

                      // Expanded(
                      //   child: TextField(
                      //     decoration: InputDecoration(
                      //       floatingLabelBehavior: FloatingLabelBehavior.always,
                      // border: OutlineInputBorder(),
                      //       labelText: global.language("doc_number"),
                      //     ),
                      //     controller: docNumberController,
                      //     onChanged: (value) {},
                      //   ),
                      // ),
                    ],
                  ),
                ),
                Container(
                  margin: EdgeInsets.only(bottom: 10),
                  child: Row(
                    children: [
                      Expanded(
                        child: TextField(
                          onSubmitted: (value) {
                            if (kIsWeb) {}
                          },
                          onChanged: (code) {
                            _debouncer.run(() {});
                          },
                          textAlign: TextAlign.left,
                          controller: TextEditingController(
                            text:
                                '${screenData.custcode}~${global.activeLangName(screenData.custnames!)}',
                          ),
                          decoration: InputDecoration(
                            floatingLabelBehavior: FloatingLabelBehavior.always,
                            border: const OutlineInputBorder(),
                            labelText:
                                (widget.type == global.TransactionTypeEnum.paid)
                                ? global.language("customer")
                                : global.language("supplier"),
                            suffixIcon: Row(
                              mainAxisAlignment:
                                  MainAxisAlignment.spaceBetween, // added line
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                IconButton(
                                  focusNode: FocusNode(skipTraversal: true),
                                  icon: const Icon(Icons.search),
                                  onPressed: () {
                                    if ((widget.type ==
                                        global.TransactionTypeEnum.paid)) {
                                      searchCustomer(word: "");
                                    } else if (widget.type ==
                                        global.TransactionTypeEnum.pay) {
                                      searchSupplier(word: "");
                                    }
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
                          onSubmitted: (value) {
                            if (kIsWeb) {}
                          },
                          onChanged: (code) {
                            _debouncer.run(() {});
                          },
                          textAlign: TextAlign.left,
                          controller: TextEditingController(
                            text:
                                '${screenData.salecode}~${screenData.salename}',
                          ),
                          decoration: InputDecoration(
                            floatingLabelBehavior: FloatingLabelBehavior.always,
                            border: OutlineInputBorder(),
                            labelText: global.language("employee"),
                            suffixIcon: Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                IconButton(
                                  focusNode: FocusNode(skipTraversal: true),
                                  icon: const Icon(Icons.search),
                                  onPressed: () {
                                    searchSale(word: "");
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
                        child: SizedBox(
                          height: 50,
                          child: InputDecorator(
                            decoration: InputDecoration(
                              labelText: global.language('paid_type'),
                              enabledBorder: OutlineInputBorder(
                                borderRadius: BorderRadius.circular(4.0),
                                borderSide: const BorderSide(
                                  color: Colors.grey,
                                  width: 0.0,
                                ),
                              ),
                            ),
                            child: Row(
                              mainAxisAlignment: MainAxisAlignment.start,
                              children: [
                                Expanded(
                                  child: Row(
                                    children: [
                                      Radio(
                                        value: 0,
                                        groupValue: screenData.doctype,
                                        onChanged: (value) {
                                          setState(() {
                                            screenData.doctype = 0;
                                          });
                                          calTotalValue();
                                        },
                                      ),
                                      Expanded(
                                        child: Text(
                                          global.language("paid_close"),
                                          overflow: TextOverflow.clip,
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                                Expanded(
                                  child: Row(
                                    children: [
                                      Radio(
                                        value: 1,
                                        groupValue: screenData.doctype,
                                        onChanged: (value) {
                                          setState(() {
                                            for (var element
                                                in screenData.details!) {
                                              element.selected = false;
                                            }
                                            screenData.doctype = 1;
                                          });
                                          calTotalValue();
                                        },
                                      ),
                                      Expanded(
                                        child: Text(
                                          global.language("paid_select"),
                                          overflow: TextOverflow.clip,
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                                Expanded(
                                  child: Row(
                                    children: [
                                      Radio(
                                        value: 2,
                                        groupValue: screenData.doctype,
                                        onChanged: (value) {
                                          setState(() {
                                            for (var element
                                                in screenData.details!) {
                                              element.selected = false;
                                            }
                                            screenData.totalpaymentamount = 0;
                                            screenData.doctype = 2;
                                          });
                                          calTotalValue();
                                        },
                                      ),
                                      Expanded(
                                        child: Text(
                                          global.language("paid_input"),
                                          overflow: TextOverflow.clip,
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
                      const SizedBox(width: 5),
                      Expanded(
                        child: Row(
                          children: [
                            Expanded(
                              child: SizedBox(
                                height: 50,
                                child: TextField(
                                  readOnly: screenData.doctype != 2,
                                  enabled: screenData.doctype == 2,
                                  onSubmitted: (value) {
                                    if (kIsWeb) {}
                                  },
                                  onChanged: (value) {
                                    _debouncer.run(() {
                                      if (value != '') {
                                        screenData.totalpaymentamount =
                                            double.parse(value);
                                        calTotalValue();
                                      }
                                    });
                                  },
                                  textAlign: TextAlign.left,
                                  controller: TextEditingController(
                                    text: global.formatNumber(
                                      screenData.totalpaymentamount,
                                    ),
                                  ),
                                  decoration: InputDecoration(
                                    floatingLabelBehavior:
                                        FloatingLabelBehavior.always,
                                    border: OutlineInputBorder(),
                                    labelText: global.language(
                                      "payment_amount",
                                    ),
                                  ),
                                ),
                              ),
                            ),
                            SizedBox(width: 5),
                            SizedBox(
                              height: 50,
                              child: ElevatedButton.icon(
                                onPressed: () {
                                  Navigator.push(
                                    context,
                                    MaterialPageRoute(
                                      builder: (context) => PaidPaymentScreen(
                                        screenData: screenData,
                                        type: widget.type,
                                      ),
                                    ),
                                  ).then((value) {
                                    setState(() {
                                      if (value[0] == "save") {
                                        calTotalValue();
                                        saveOrUpdateData();
                                      }
                                    });
                                  });
                                },
                                icon: Icon(Icons.money),
                                label: Text(
                                  "${global.language("pay_amount")} ",
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
                // Container(
                //   alignment: Alignment.centerLeft,
                //   margin: const EdgeInsets.only(bottom: 10),
                //   child: ElevatedButton(
                //       onPressed: () {
                //         processDoc();
                //       },
                //       child: const Text("process")),
                // ),
                Container(
                  margin: EdgeInsets.only(bottom: 10),
                  child: Column(
                    children: [
                      Row(
                        mainAxisAlignment: MainAxisAlignment.start,
                        children: [
                          Text(
                            global.language("detail"),
                            style: const TextStyle(fontSize: 17),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
                listDetails(),
                Divider(height: 3),
                Container(
                  margin: EdgeInsets.all(13),
                  child: Column(
                    children: [
                      Row(
                        mainAxisAlignment: MainAxisAlignment.end,
                        children: [
                          Text(
                            global.language("total_amount"),
                            style: const TextStyle(
                              fontSize: 17,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          const SizedBox(width: 15),
                          Text(
                            global.formatNumber(screenData.totalvalue),
                            style: const TextStyle(
                              fontSize: 17,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          SizedBox(width: 50),
                          Text(
                            global.language("balance"),
                            style: const TextStyle(
                              fontSize: 17,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          const SizedBox(width: 15),
                          Text(
                            global.formatNumber(screenData.totalbalance),
                            style: const TextStyle(
                              fontSize: 17,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          SizedBox(width: 50),
                          Text(
                            global.language("payment_amount"),
                            style: const TextStyle(
                              fontSize: 17,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                          const SizedBox(width: 15),
                          Text(
                            global.formatNumber(screenData.totalpaymentamount),
                            style: const TextStyle(
                              fontSize: 17,
                              fontWeight: FontWeight.bold,
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
        ),
      ),
    );
  }

  Widget listDetails() {
    List<TableRow> tableDetails = [];

    tableDetails.add(
      TableRow(
        children: [
          Text(''),
          Text(
            global.language('doc_number'),
            style: const TextStyle(fontWeight: FontWeight.bold),
            textAlign: TextAlign.center,
          ),
          Text(
            global.language('doc_date'),
            style: const TextStyle(fontWeight: FontWeight.bold),
            textAlign: TextAlign.center,
          ),
          Text(
            global.language('doc_type'),
            style: const TextStyle(fontWeight: FontWeight.bold),
            textAlign: TextAlign.center,
          ),
          Text(
            global.language('cash_amount'),
            style: const TextStyle(fontWeight: FontWeight.bold),
            textAlign: TextAlign.center,
          ),
          Text(
            global.language('balance'),
            style: const TextStyle(fontWeight: FontWeight.bold),
            textAlign: TextAlign.center,
          ),
          Text(
            global.language('payment_amount'),
            style: const TextStyle(fontWeight: FontWeight.bold),
            textAlign: TextAlign.center,
          ),
          const Text(''),
        ],
      ),
    );

    for (int i = 0; i < screenData.details!.length; i++) {
      tableDetails.add(
        TableRow(
          children: [
            Checkbox(
              value: screenData.details![i].selected,
              onChanged: (value) {
                if (screenData.doctype == 1) {
                  screenData.details![i].selected = value;
                  setState(() {});
                  calTotalValue();
                }
              },
            ),
            Text(screenData.details![i].docno),
            Text(
              global.docDateTimeFormateDDMMYYY(
                screenData.details![i].docdatetime,
              ),
              textAlign: TextAlign.center,
            ),
            Text(
              global.language(
                global.getDocType(screenData.details![i].transflag),
              ),
              textAlign: TextAlign.center,
            ),
            Text(
              global.formatNumber(screenData.details![i].value),
              textAlign: TextAlign.right,
            ),
            Text(
              global.formatNumber(screenData.details![i].balance),
              textAlign: TextAlign.right,
            ),
            (screenData.doctype != 1)
                ? Text(
                    global.formatNumber(screenData.details![i].paymentamount),
                    textAlign: TextAlign.right,
                  )
                : Container(
                    padding: EdgeInsets.symmetric(horizontal: 16.0),
                    child: TextFormField(
                      readOnly: false,
                      onChanged: (value) {
                        _debouncer.run(() {
                          double paymentamount = double.parse(
                            value.replaceAll(',', ''),
                          );
                          paymentamount = double.parse(
                            value.replaceAll(',', ''),
                          );

                          if (paymentamount > screenData.details![i].balance) {
                            showDialog(
                              context: context,
                              builder: (BuildContext context) {
                                return Center(
                                  child: AlertDialog(
                                    content: Text(
                                      global.language("payment_over"),
                                    ),
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

                            screenData.details![i].paymentamount = 0;
                            setState(() {});
                          } else {
                            screenData.details![i].paymentamount =
                                paymentamount;
                          }
                          calTotalValue();
                        });
                      },
                      textAlign: TextAlign.right,
                      controller: TextEditingController(
                        text: global.formatNumber(
                          screenData.details![i].paymentamount,
                        ),
                      ),
                      keyboardType: TextInputType.number,
                      inputFormatters: [
                        FilteringTextInputFormatter.allow(
                          RegExp(r'^\d+\.?\d{0,2}'),
                        ),
                      ],
                    ),
                  ),
            IconButton(
              onPressed: () {
                screenData.details!.removeAt(i);
                setState(() {});
                calTotalValue();
              },
              icon: const Icon(Icons.delete, color: Colors.red),
            ),
          ],
        ),
      );
    }

    Widget tableWidget = Expanded(
      child: SingleChildScrollView(
        scrollDirection: Axis.vertical,
        child: Center(
          child: Table(
            columnWidths: const <int, TableColumnWidth>{
              0: FlexColumnWidth(0.5),
              1: FlexColumnWidth(1.5),
              2: FlexColumnWidth(1),
              3: FlexColumnWidth(1),
              4: FlexColumnWidth(1),
              5: FlexColumnWidth(1),
              6: FlexColumnWidth(1),
              7: FlexColumnWidth(0.5),
            },
            defaultVerticalAlignment: TableCellVerticalAlignment.middle,
            children: tableDetails,
          ),
        ),
      ),
    );

    return tableWidget;
  }

  void searchTrans() {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => TransPaidPaySearchScreen(type: widget.type),
      ),
    ).then((value) {
      TransactionPaidPayModel result = value;
      if (result.docno.isNotEmpty) {
        setState(() {
          DateTime localDateTime = DateTime.parse(result.docdatetime);

          screenData = result;
          screenData.docdatetime = localDateTime.toUtc().toIso8601String(); // แปลงเป็น UTC ตามกฎ CLAUDE.md
          docTimeController.text = DateFormat(
            'HH:mm',
          ).format(localDateTime.toLocal());

          DateTime docDateTimeFormat = DateTime.parse(screenData.docdatetime);
          if (global.profileData.yeartype == "buddhist") {
            docDateController.text = global.dateTimeBuddhist(
              docDateTimeFormat,
              format: global.DateTimeFormatEnum.dateDay,
            );
          } else {
            docDateController.text = DateFormat(
              'dd/MM/yyyy',
            ).format(docDateTimeFormat);
          }
        });
      }
    });
  }

  Future<void> processDoc() async {
    setState(() {
      screenData.details!.clear();
    });

    /// รับชำระ
    if (widget.type == global.TransactionTypeEnum.paid) {
      context.read<TransactionPaidPayBloc>().add(
        GetCustcodeTransaction(
          type: global.TransactionTypeEnum.paid,
          custcode: screenData.custcode,
        ),
      );

      /// จ่ายชำระ
    } else if (widget.type == global.TransactionTypeEnum.pay) {
      context.read<TransactionPaidPayBloc>().add(
        GetCustcodeTransaction(
          type: global.TransactionTypeEnum.pay,
          custcode: screenData.custcode,
        ),
      );
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
          screenData.salecode = result.code;
          screenData.salename = result.name;
        });
      }
    });
  }

  void searchSupplier({required String word}) {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (context) => SupplierSearchScreen(word: word)),
    ).then((value) {
      SearchGuidCodeNameModel result = value;
      if (result.code.trim().isNotEmpty) {
        setState(() {
          screenData.custcode = result.code;
          screenData.custnames = result.names;
          processDoc();
        });
      }
    });
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

  void searchCustomer({required String word}) {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (context) => CustomerSearchScreen(word: word)),
    ).then((value) {
      global.SearchDebtorModel result = value;
      if (result.code.trim().isNotEmpty) {
        setState(() {
          screenData.custcode = result.code;
          screenData.custnames = result.names;
          processDoc();
        });
      }
    });
  }
}
