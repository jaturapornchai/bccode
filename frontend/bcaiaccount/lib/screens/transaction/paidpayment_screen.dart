import 'dart:async';

import 'package:smlaicloud/model/book_bank_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/screen_search/bookbank_select_screen.dart';
import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/widgets/edit_font_size_control.dart';
import 'package:flutter_rounded_date_picker/flutter_rounded_date_picker.dart';
import 'package:smlaicloud/components/numpad.dart';

// ignore: depend_on_referenced_packages
import 'package:intl/intl.dart';

class PaidPaymentScreen extends StatefulWidget {
  final TransactionPaidPayModel screenData;
  final global.TransactionTypeEnum type;

  const PaidPaymentScreen({
    super.key,
    required this.screenData,
    required this.type,
  });

  @override
  State<PaidPaymentScreen> createState() => _PaidPaymentScreenState();
}

class _PaidPaymentScreenState extends State<PaidPaymentScreen>
    with global.ThemeRefreshMixin {
  late TransactionPaidPayModel screenData;

  int showPayDetail = 0;
  bool docDateTimeValidated = false;

  List<BillPayStruct> payTransfer = [];
  List<BillPayStruct> payCreditCard = [];
  List<BillPayStruct> payCheque = [];
  List<BillPayStruct> payCoupon = [];
  List<BillPayStruct> payQr = [];

  List<TextEditingController> creditCardDateController = [];
  List<TextEditingController> transferDateController = [];
  List<TextEditingController> chequeDateController = [];
  List<TextEditingController> chequeDueDateDateController = [];
  List<TextEditingController> couponDateController = [];
  List<TextEditingController> qrDateController = [];

  List<TextEditingController> payTransferAmountController = [];
  List<TextEditingController> payCreditCardAmountController = [];
  List<TextEditingController> payChequeAmountController = [];
  List<TextEditingController> payCouponAmountController = [];
  List<TextEditingController> payQrAmountController = [];
  TextEditingController roundAmountController = TextEditingController();
  TextEditingController payCashAmountController = TextEditingController();
  double totalPayCreditCard = 0;
  double totalPayTranfer = 0;
  double totalPayCheque = 0;
  double totalPayCoupon = 0;
  double totalPayQr = 0;
  double payTotalBill = 0;

  DateTime selectedDate = DateTime.now();

  @override
  void initState() {
    screenData = widget.screenData;

    loadPayDetail();
    super.initState();
  }

  @override
  void dispose() {
    super.dispose();
  }

  void loadPayDetail() {
    payTransfer = [];
    payCreditCard = [];
    payCheque = [];
    payCoupon = [];
    payQr = [];

    creditCardDateController = [];
    transferDateController = [];
    chequeDateController = [];
    chequeDueDateDateController = [];
    couponDateController = [];
    qrDateController = [];

    for (int i = 0; i < screenData.paymentStructs!.length; i++) {
      /// บัตรเครดิต
      if (screenData.paymentStructs![i].trans_flag == 1) {
        payCreditCard.add(screenData.paymentStructs![i]);
        totalPayCreditCard += screenData.paymentStructs![i].amount!;
        creditCardDateController.add(TextEditingController());
        payCreditCardAmountController.add(TextEditingController());

        for (int i = 0; i < creditCardDateController.length; i++) {
          if (global.profileData.yeartype == "buddhist") {
            creditCardDateController[i].text = global.dateTimeBuddhist(
              screenData.paymentStructs![i].doc_date_time!,
              format: global.DateTimeFormatEnum.dateDay,
            );
          } else {
            creditCardDateController[i].text = DateFormat(
              'dd/MM/yyyy',
            ).format(screenData.paymentStructs![i].doc_date_time!);
          }
          payCreditCardAmountController[i].text = screenData
              .paymentStructs![i]
              .amount
              .toString();
        }
      }

      /// เงินโอน
      if (screenData.paymentStructs![i].trans_flag == 2) {
        payTransfer.add(screenData.paymentStructs![i]);
        totalPayTranfer += screenData.paymentStructs![i].amount!;
        transferDateController.add(TextEditingController());
        payTransferAmountController.add(TextEditingController());

        for (int i = 0; i < transferDateController.length; i++) {
          if (global.profileData.yeartype == "buddhist") {
            transferDateController[i].text = global.dateTimeBuddhist(
              screenData.paymentStructs![i].doc_date_time!,
              format: global.DateTimeFormatEnum.dateDay,
            );
          } else {
            transferDateController[i].text = DateFormat(
              'dd/MM/yyyy',
            ).format(screenData.paymentStructs![i].doc_date_time!);
          }
          payTransferAmountController[i].text = global
              .formatNumber(screenData.paymentStructs![i].amount!)
              .toString();
        }
      }

      /// เช็ค
      if (screenData.paymentStructs![i].trans_flag == 3) {
        payCheque.add(screenData.paymentStructs![i]);
        totalPayCheque += screenData.paymentStructs![i].amount!;
        chequeDateController.add(TextEditingController());
        chequeDueDateDateController.add(TextEditingController());
        payChequeAmountController.add(TextEditingController());

        for (int i = 0; i < transferDateController.length; i++) {
          if (global.profileData.yeartype == "buddhist") {
            chequeDateController[i].text = global.dateTimeBuddhist(
              screenData.paymentStructs![i].doc_date_time!,
              format: global.DateTimeFormatEnum.dateDay,
            );
            chequeDueDateDateController[i].text = global.dateTimeBuddhist(
              screenData.paymentStructs![i].due_date!,
              format: global.DateTimeFormatEnum.dateDay,
            );
          } else {
            chequeDateController[i].text = DateFormat(
              'dd/MM/yyyy',
            ).format(screenData.paymentStructs![i].doc_date_time!);
            chequeDueDateDateController[i].text = DateFormat(
              'dd/MM/yyyy',
            ).format(screenData.paymentStructs![i].due_date!);
          }
          payChequeAmountController[i].text = screenData
              .paymentStructs![i]
              .amount
              .toString();
        }
      }

      /// คูปอง
      if (screenData.paymentStructs![i].trans_flag == 4) {
        payCoupon.add(screenData.paymentStructs![i]);
        totalPayCoupon += screenData.paymentStructs![i].amount!;
        couponDateController.add(TextEditingController());
        payCouponAmountController.add(TextEditingController());

        for (int i = 0; i < couponDateController.length; i++) {
          if (global.profileData.yeartype == "buddhist") {
            couponDateController[i].text = global.dateTimeBuddhist(
              screenData.paymentStructs![i].doc_date_time!,
              format: global.DateTimeFormatEnum.dateDay,
            );
          } else {
            couponDateController[i].text = DateFormat(
              'dd/MM/yyyy',
            ).format(screenData.paymentStructs![i].doc_date_time!);
          }
          payCouponAmountController[i].text = screenData
              .paymentStructs![i]
              .amount
              .toString();
        }
      }

      /// QR
      if (screenData.paymentStructs![i].trans_flag == 5) {
        payQr.add(screenData.paymentStructs![i]);
        totalPayQr += screenData.paymentStructs![i].amount!;
        qrDateController.add(TextEditingController());
        payQrAmountController.add(TextEditingController());

        for (int i = 0; i < qrDateController.length; i++) {
          if (global.profileData.yeartype == "buddhist") {
            qrDateController[i].text = global.dateTimeBuddhist(
              screenData.paymentStructs![i].doc_date_time!,
              format: global.DateTimeFormatEnum.dateDay,
            );
          } else {
            qrDateController[i].text = DateFormat(
              'dd/MM/yyyy',
            ).format(screenData.paymentStructs![i].doc_date_time!);
          }
          payQrAmountController[i].text = screenData
              .paymentStructs![i]
              .amount
              .toString();
        }
      }
    }

    screenData.sumcreditcard = totalPayCreditCard;
    screenData.summoneytransfer = totalPayTranfer;
    screenData.sumcheque = totalPayCheque;
    screenData.sumcoupon = totalPayCoupon;
    screenData.sumqrcode = totalPayQr;
    payCashAmountController.text = global
        .formatNumber(screenData.paycashamount!)
        .toString();
    roundAmountController.text = (screenData.roundamount != 0)
        ? (screenData.roundamount!).toString()
        : "0";
    _calPayTotal();
  }

  Future<DateTime?> selectPayDate(BuildContext context, int index) async {
    DateTime? pickDateTimeFormat;
    final DateTime? pickedDate = await showRoundedDatePicker(
      context: context,
      initialDate: payTransfer[index].doc_date_time!,
      firstDate: DateTime(2020),
      lastDate: DateTime.now(),
      locale: global.local,
      era: global.eraMode,
      borderRadius: 16,
    );

    if (pickedDate != null && pickedDate != selectedDate) {
      setState(() {
        pickDateTimeFormat = DateTime.parse(
          '${DateFormat('yyyy-MM-dd').format(pickedDate)} ${DateFormat('HH:mm:ss.sss').format(DateTime.now())}',
        );

        docDateTimeValidated = true;
      });
    }

    return pickDateTimeFormat;
  }

  Future<BookBankModel?> bookBankSearch() async {
    Completer<BookBankModel?> completer = Completer<BookBankModel?>();

    Navigator.push(
      context,
      MaterialPageRoute(builder: (context) => const BookBankSelectScreen()),
    ).then((value) {
      completer.complete(value);
    });

    return completer.future;
  }

  void _calPayTotal() {
    double totalPayCash = 0;
    double roundAmount = 0;
    double totalPayCreditCard = 0;
    double totalPayCheque = 0;
    double totalPayTransfer = 0;
    double totalPayCoupon = 0;
    double totalQr = 0;

    if (payCashAmountController.text.isEmpty ||
        payCashAmountController.text == "0") {
      totalPayCash = 0.0;
    } else {
      totalPayCash = double.parse(
        payCashAmountController.text.replaceAll(',', ''),
      );
    }
    if (roundAmountController.text.isEmpty ||
        roundAmountController.text == "0") {
      roundAmount = 0.0;
    } else {
      roundAmount = double.parse(
        roundAmountController.text.replaceAll(',', ''),
      );
    }

    for (var element in payTransferAmountController) {
      totalPayTransfer += double.parse(element.text.replaceAll(',', ''));
    }

    for (var element in payCreditCardAmountController) {
      totalPayCreditCard += double.parse(element.text.replaceAll(',', ''));
    }

    for (var element in payChequeAmountController) {
      totalPayCheque += double.parse(element.text.replaceAll(',', ''));
    }

    for (var element in payCouponAmountController) {
      totalPayCoupon += double.parse(element.text.replaceAll(',', ''));
    }

    for (var element in payQrAmountController) {
      totalQr += double.parse(element.text.replaceAll(',', ''));
    }

    screenData.paycashamount = totalPayCash;
    screenData.summoneytransfer = totalPayTransfer;
    screenData.sumcreditcard = totalPayCreditCard;
    screenData.sumcheque = totalPayCheque;
    screenData.sumcoupon = totalPayCoupon;
    screenData.sumqrcode = totalQr;
    screenData.roundamount = roundAmount;
    payTotalBill =
        totalPayCash +
        totalPayTransfer +
        totalPayCreditCard +
        totalPayCheque +
        totalPayCoupon +
        totalQr +
        screenData.sumcredit!;

    setState(() {});
  }

  Widget payMenuWidget() {
    return SizedBox(
      width: double.infinity,
      child: Column(
        children: [
          Row(
            children: [
              Expanded(
                child: Padding(
                  padding: EdgeInsets.all(5.0),
                  child: ElevatedButton.icon(
                    onPressed: () {
                      showPayDetail = 0;
                      setState(() {});
                    },
                    icon: Icon(Icons.money),
                    label: Text(global.language("cash")),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: (showPayDetail == 0)
                          ? global.theme.infoHighlightTextColor
                          : global.theme.iconSecondaryColor,
                    ),
                  ),
                ),
              ),
              Expanded(
                child: Padding(
                  padding: EdgeInsets.all(5.0),
                  child: ElevatedButton.icon(
                    onPressed: () {
                      showPayDetail = 1;
                      setState(() {});
                    },
                    icon: Icon(Icons.transform_rounded),
                    label: Text(global.language("money_transfer")),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: (showPayDetail == 1)
                          ? global.theme.infoHighlightTextColor
                          : global.theme.iconSecondaryColor,
                    ),
                  ),
                ),
              ),
              (widget.type == global.TransactionTypeEnum.paid)
                  ? Expanded(
                      child: Padding(
                        padding: EdgeInsets.all(5.0),
                        child: ElevatedButton.icon(
                          onPressed: () {
                            showPayDetail = 2;
                            setState(() {});
                          },
                          icon: Icon(Icons.credit_card),
                          label: Text(global.language("credit_card")),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: (showPayDetail == 2)
                                ? global.theme.infoHighlightTextColor
                                : global.theme.iconSecondaryColor,
                          ),
                        ),
                      ),
                    )
                  : Container(),
            ],
          ),
          (widget.type == global.TransactionTypeEnum.paid)
              ? Row(
                  children: [
                    Expanded(
                      child: Padding(
                        padding: EdgeInsets.all(5.0),
                        child: ElevatedButton.icon(
                          onPressed: () {
                            showPayDetail = 3;
                            setState(() {});
                          },
                          icon: Icon(Icons.featured_play_list_outlined),
                          label: Text(global.language("cheque")),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: (showPayDetail == 3)
                                ? global.theme.infoHighlightTextColor
                                : global.theme.iconSecondaryColor,
                          ),
                        ),
                      ),
                    ),
                    Expanded(
                      child: Padding(
                        padding: EdgeInsets.all(5.0),
                        child: ElevatedButton.icon(
                          onPressed: () {
                            showPayDetail = 4;
                            setState(() {});
                          },
                          icon: Icon(Icons.card_giftcard),
                          label: Text(global.language("coupon")),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: (showPayDetail == 4)
                                ? global.theme.infoHighlightTextColor
                                : global.theme.iconSecondaryColor,
                          ),
                        ),
                      ),
                    ),
                    Expanded(
                      child: Padding(
                        padding: EdgeInsets.all(5.0),
                        child: ElevatedButton.icon(
                          onPressed: () {
                            showPayDetail = 5;
                            setState(() {});
                          },
                          icon: Icon(Icons.qr_code),
                          label: Text(global.language("qr_code")),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: (showPayDetail == 5)
                                ? global.theme.infoHighlightTextColor
                                : global.theme.iconSecondaryColor,
                          ),
                        ),
                      ),
                    ),
                  ],
                )
              : Container(),
        ],
      ),
    );
  }

  Widget payCashWidget() {
    return Center(
      child: Padding(
        padding: EdgeInsets.all(30.0),
        child: Column(
          children: [
            Container(
              padding: EdgeInsets.all(16), // Adjust the padding as needed
              child: TextFormField(
                textAlign: TextAlign.center, // Center-align the text
                enabled: true,
                decoration: InputDecoration(
                  border: OutlineInputBorder(),
                  labelText: global.language("cash"),
                ),
                style: TextStyle(
                  fontSize: 28,
                ), // Adjust the font size as needed
                controller: payCashAmountController,
                keyboardType: const TextInputType.numberWithOptions(
                  decimal: true,
                ),
                inputFormatters: [global.NumberInputFormatter()],
                onChanged: (value) {
                  if (value == '' && value.isEmpty) {
                    payCashAmountController.text = "0";
                  } else {
                    payCashAmountController.value = TextEditingValue(
                      text: value.toUpperCase(),
                      selection: payCashAmountController.selection,
                    );
                  }
                  _calPayTotal();
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget payCreditCardWidget(List<Widget> listCredit) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          margin: const EdgeInsets.only(left: 10, top: 5, bottom: 5),
          child: ElevatedButton.icon(
            onPressed: () {
              creditCardDateController.add(TextEditingController());
              payCreditCardAmountController.add(TextEditingController());

              if (global.profileData.yeartype == "buddhist") {
                creditCardDateController[creditCardDateController.length - 1]
                    .text = global.dateTimeBuddhist(
                  DateTime.now(),
                  format: global.DateTimeFormatEnum.dateDay,
                );
              } else {
                creditCardDateController[creditCardDateController.length - 1]
                    .text = DateFormat('dd/MM/yyyy').format(
                  DateTime.parse(DateTime.now().toUtc().toIso8601String()),
                );
              }
              payCreditCardAmountController[creditCardDateController.length - 1]
                      .text =
                  "0";

              payCreditCard.add(BillPayStruct(trans_flag: 1));

              setState(() {});
            },
            icon: Icon(Icons.add),
            label: Text(global.language("add_credit")),
          ),
        ),
        Container(
          padding: const EdgeInsets.all(5),
          child: Column(children: listCredit),
        ),
      ],
    );
  }

  Widget payTransferWidget(List<Widget> listTransfer) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          margin: const EdgeInsets.only(left: 10, top: 5, bottom: 5),
          child: ElevatedButton.icon(
            onPressed: () {
              transferDateController.add(TextEditingController());
              payTransferAmountController.add(TextEditingController());

              if (global.profileData.yeartype == "buddhist") {
                transferDateController[transferDateController.length - 1].text =
                    global.dateTimeBuddhist(
                      DateTime.now(),
                      format: global.DateTimeFormatEnum.dateDay,
                    );
              } else {
                transferDateController[transferDateController.length - 1].text =
                    DateFormat('dd/MM/yyyy').format(
                      DateTime.parse(DateTime.now().toUtc().toIso8601String()),
                    );
              }
              payTransferAmountController[transferDateController.length - 1]
                      .text =
                  "0";

              payTransfer.add(BillPayStruct(trans_flag: 2));

              setState(() {});
            },
            icon: Icon(Icons.add),
            label: Text(global.language("add_transfer")),
          ),
        ),
        Container(
          padding: const EdgeInsets.all(5),
          child: Column(children: listTransfer),
        ),
      ],
    );
  }

  Widget payChequeWidget(List<Widget> listCheque) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          margin: const EdgeInsets.only(left: 10, top: 5, bottom: 5),
          child: ElevatedButton.icon(
            onPressed: () {
              chequeDateController.add(TextEditingController());
              chequeDueDateDateController.add(TextEditingController());
              payChequeAmountController.add(TextEditingController());

              if (global.profileData.yeartype == "buddhist") {
                chequeDateController[chequeDateController.length - 1].text =
                    global.dateTimeBuddhist(
                      DateTime.now(),
                      format: global.DateTimeFormatEnum.dateDay,
                    );
              } else {
                chequeDateController[chequeDateController.length - 1].text =
                    DateFormat('dd/MM/yyyy').format(
                      DateTime.parse(DateTime.now().toUtc().toIso8601String()),
                    );
              }

              if (global.profileData.yeartype == "buddhist") {
                chequeDueDateDateController[chequeDueDateDateController.length -
                        1]
                    .text = global.dateTimeBuddhist(
                  DateTime.now(),
                  format: global.DateTimeFormatEnum.dateDay,
                );
              } else {
                chequeDueDateDateController[chequeDueDateDateController.length -
                        1]
                    .text = DateFormat('dd/MM/yyyy').format(
                  DateTime.parse(DateTime.now().toUtc().toIso8601String()),
                );
              }

              payChequeAmountController[chequeDueDateDateController.length - 1]
                      .text =
                  "0";

              payCheque.add(BillPayStruct(trans_flag: 3));
              setState(() {});
            },
            icon: Icon(Icons.add),
            label: Text(global.language("add_cheque")),
          ),
        ),
        Container(
          padding: const EdgeInsets.all(5),
          child: Column(children: listCheque),
        ),
      ],
    );
  }

  Widget payCouponWidget(List<Widget> listCoupon) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          margin: const EdgeInsets.only(left: 10, top: 5, bottom: 5),
          child: ElevatedButton.icon(
            onPressed: () {
              couponDateController.add(TextEditingController());
              payCouponAmountController.add(TextEditingController());

              if (global.profileData.yeartype == "buddhist") {
                couponDateController[couponDateController.length - 1].text =
                    global.dateTimeBuddhist(
                      DateTime.now(),
                      format: global.DateTimeFormatEnum.dateDay,
                    );
              } else {
                couponDateController[couponDateController.length - 1].text =
                    DateFormat('dd/MM/yyyy').format(
                      DateTime.parse(DateTime.now().toUtc().toIso8601String()),
                    );
              }

              payCouponAmountController[couponDateController.length - 1].text =
                  "0";

              payCoupon.add(BillPayStruct(trans_flag: 4));
              setState(() {});
            },
            icon: Icon(Icons.add),
            label: Text(global.language("add_coupon")),
          ),
        ),
        Container(
          padding: const EdgeInsets.all(5),
          child: Column(children: listCoupon),
        ),
      ],
    );
  }

  Widget payQrWidget(List<Widget> listQr) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          margin: const EdgeInsets.only(left: 10, top: 5, bottom: 5),
          child: ElevatedButton.icon(
            onPressed: () {
              qrDateController.add(TextEditingController());
              payQrAmountController.add(TextEditingController());

              if (global.profileData.yeartype == "buddhist") {
                qrDateController[qrDateController.length - 1].text = global
                    .dateTimeBuddhist(
                      DateTime.now(),
                      format: global.DateTimeFormatEnum.dateDay,
                    );
              } else {
                qrDateController[qrDateController.length - 1].text =
                    DateFormat('dd/MM/yyyy').format(
                      DateTime.parse(DateTime.now().toUtc().toIso8601String()),
                    );
              }

              payQrAmountController[qrDateController.length - 1].text = "0";

              payQr.add(BillPayStruct(trans_flag: 5));
              setState(() {});
            },
            icon: Icon(Icons.add),
            label: Text(global.language("add_qr")),
          ),
        ),
        Container(
          padding: const EdgeInsets.all(5),
          child: Column(children: listQr),
        ),
      ],
    );
  }

  Widget editSummeryWidget() {
    List<Widget> paymentDetail = [];

    List<Widget> listTransfer = [];
    List<Widget> listCredit = [];
    List<Widget> listCheque = [];
    List<Widget> listCoupon = [];
    List<Widget> listQr = [];

    /// widget for transfer
    for (var i = 0; i < payTransfer.length; i++) {
      listTransfer.add(
        Card(
          child: Container(
            padding: EdgeInsets.all(10),
            child: Column(
              children: [
                Padding(
                  padding: EdgeInsets.only(bottom: 10),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        "${global.language("list_transfer")} ${i + 1}",
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      Container(
                        margin: const EdgeInsets.only(left: 15),
                        child: IconButton(
                          onPressed: () {
                            payTransfer.removeAt(i);
                            transferDateController.removeAt(i);
                            payTransferAmountController.removeAt(i);
                            _calPayTotal();
                          },
                          icon: Icon(Icons.delete),
                        ),
                      ),
                    ],
                  ),
                ),
                Row(
                  children: [
                    Expanded(
                      child: TextField(
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
                                icon: Icon(Icons.calendar_month),
                                onPressed: () {
                                  selectPayDate(context, i).then((value) {
                                    if (value != null) {
                                      if (global.profileData.yeartype ==
                                          "buddhist") {
                                        transferDateController[i].text = global
                                            .dateTimeBuddhist(
                                              value,
                                              format: global
                                                  .DateTimeFormatEnum
                                                  .dateDay,
                                            );
                                      } else {
                                        transferDateController[i].text =
                                            DateFormat(
                                              'dd/MM/yyyy',
                                            ).format(value);
                                      }
                                      payTransfer[i].doc_date_time = value
                                          .toLocal();

                                      setState(() {});
                                    }
                                  });
                                },
                              ),
                            ],
                          ),
                        ),
                        controller: transferDateController[i],
                        onChanged: (value) {
                          setState(() {
                            docDateTimeValidated = false;
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
                              if (transferDateController[i].text
                                  .trim()
                                  .isEmpty) {
                                if (global.isValidDate(value)) {
                                  payTransfer[i].doc_date_time = DateTime.parse(
                                    value,
                                  ).toLocal();
                                  docDateTimeValidated = true;
                                }
                              } else {
                                if (global.isValidDate(value)) {
                                  payTransfer[i].doc_date_time = DateTime.parse(
                                    '$value ${transferDateController[i].text}',
                                  ).toLocal();
                                  docDateTimeValidated = true;
                                }
                              }
                            } catch (e) {
                              // print(e);
                            }
                          });
                        },
                        onSubmitted: (value) {
                          transferDateController[i].text = DateFormat(
                            'dd/MM/yyyy',
                          ).format(payTransfer[i].doc_date_time!);
                        },
                      ),
                    ),
                    const SizedBox(width: 5),
                    Expanded(
                      child: RawKeyboardListener(
                        focusNode: FocusNode(),
                        child: TextField(
                          readOnly: true,
                          textInputAction: TextInputAction.next,
                          controller: TextEditingController(
                            text: payTransfer[i].book_bank_code,
                          ),
                          textAlign: TextAlign.left,
                          textCapitalization: TextCapitalization.characters,
                          decoration: InputDecoration(
                            enabledBorder: OutlineInputBorder(
                              borderSide: BorderSide(
                                color: global.theme.dividerBorderColor,
                                width: 0.0,
                              ),
                            ),
                            floatingLabelBehavior: FloatingLabelBehavior.always,
                            suffixIcon: Row(
                              mainAxisAlignment:
                                  MainAxisAlignment.spaceBetween, // added line
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                IconButton(
                                  focusNode: FocusNode(skipTraversal: true),
                                  icon: Icon(Icons.search),
                                  onPressed: () {
                                    bookBankSearch().then((value) {
                                      if (value != null) {
                                        payTransfer[i].book_bank_code =
                                            value.passbook;
                                        payTransfer[i].bank_code =
                                            value.bankcode;
                                        payTransfer[i].bank_name =
                                            value.banknames![0].name;
                                        setState(() {});
                                      }
                                    });
                                  },
                                ),
                              ],
                            ),
                            border: OutlineInputBorder(),
                            labelText: global.language("pass_book"),
                          ),
                        ),
                      ),
                    ),
                    SizedBox(width: 5),
                    Expanded(
                      child: TextField(
                        readOnly: true,
                        decoration: InputDecoration(
                          border: OutlineInputBorder(),
                          labelText: global.language("bank"),
                          floatingLabelBehavior: FloatingLabelBehavior.always,
                        ),
                        controller: TextEditingController(
                          text: (payTransfer[i].bank_code != '')
                              ? " ${payTransfer[i].bank_code} ~ ${payTransfer[i].bank_name ?? []}"
                              : "",
                        ),
                      ),
                    ),
                  ],
                ),
                Container(
                  margin: const EdgeInsets.only(top: 10),
                  child: Row(
                    children: [
                      Expanded(
                        child: TextFormField(
                          textAlign: TextAlign.center, // Center-align the text
                          decoration: const InputDecoration(
                            border: OutlineInputBorder(),
                          ),
                          controller: payTransferAmountController[i],
                          keyboardType: const TextInputType.numberWithOptions(
                            decimal: true,
                          ),
                          inputFormatters: [global.NumberInputFormatter()],
                          onChanged: (value) {
                            if (value == '' && value.isEmpty) {
                              payTransferAmountController[i].text = "0";
                              payTransfer[i].amount = 0;
                            } else {
                              payTransferAmountController[i]
                                  .value = TextEditingValue(
                                text: value.toUpperCase(),
                                selection:
                                    payTransferAmountController[i].selection,
                              );
                              payTransfer[i].amount = double.parse(
                                value.replaceAll(',', ''),
                              );
                            }

                            _calPayTotal();
                          },
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      );
    }

    /// widget for credit card
    for (var i = 0; i < payCreditCard.length; i++) {
      listCredit.add(
        Card(
          child: Container(
            padding: EdgeInsets.all(10),
            child: Column(
              children: [
                Padding(
                  padding: EdgeInsets.only(bottom: 10),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        "${global.language("list_creditcard")} ${i + 1}",
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      Container(
                        margin: const EdgeInsets.only(left: 15),
                        child: IconButton(
                          onPressed: () {
                            payCreditCard.removeAt(i);
                            creditCardDateController.removeAt(i);
                            payCreditCardAmountController.removeAt(i);
                            _calPayTotal();
                          },
                          icon: Icon(Icons.delete),
                        ),
                      ),
                    ],
                  ),
                ),
                Row(
                  children: [
                    Expanded(
                      child: TextField(
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
                                icon: Icon(Icons.calendar_month),
                                onPressed: () {
                                  selectPayDate(context, i).then((value) {
                                    if (value != null) {
                                      if (global.profileData.yeartype ==
                                          "buddhist") {
                                        creditCardDateController[i]
                                            .text = global.dateTimeBuddhist(
                                          value,
                                          format:
                                              global.DateTimeFormatEnum.dateDay,
                                        );
                                      } else {
                                        creditCardDateController[i].text =
                                            DateFormat(
                                              'dd/MM/yyyy',
                                            ).format(value);
                                      }
                                      payCreditCard[i].doc_date_time = value
                                          .toLocal();

                                      setState(() {});
                                    }
                                  });
                                },
                              ),
                            ],
                          ),
                        ),
                        controller: creditCardDateController[i],
                        onChanged: (value) {
                          setState(() {
                            docDateTimeValidated = false;
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
                              if (creditCardDateController[i].text
                                  .trim()
                                  .isEmpty) {
                                if (global.isValidDate(value)) {
                                  payCreditCard[i].doc_date_time =
                                      DateTime.parse(value).toLocal();
                                  docDateTimeValidated = true;
                                }
                              } else {
                                if (global.isValidDate(value)) {
                                  payCreditCard[i]
                                      .doc_date_time = DateTime.parse(
                                    '$value ${creditCardDateController[i].text}',
                                  ).toLocal();
                                  docDateTimeValidated = true;
                                }
                              }
                            } catch (e) {
                              // print(e);
                            }
                          });
                        },
                        onSubmitted: (value) {
                          creditCardDateController[i].text = DateFormat(
                            'dd/MM/yyyy',
                          ).format(payCreditCard[i].doc_date_time!);
                        },
                      ),
                    ),
                    SizedBox(width: 5),
                    Expanded(
                      child: TextField(
                        decoration: InputDecoration(
                          border: OutlineInputBorder(),
                          labelText: global.language("creditnumber"),
                          floatingLabelBehavior: FloatingLabelBehavior.always,
                        ),
                        controller: TextEditingController(
                          text: payCreditCard[i].card_number,
                        ),
                        onChanged: (value) {
                          payCreditCard[i].card_number = value;
                        },
                        onSubmitted: (value) {},
                      ),
                    ),
                  ],
                ),
                Padding(
                  padding: const EdgeInsets.only(top: 10),
                  child: Row(
                    children: [
                      Expanded(
                        child: RawKeyboardListener(
                          focusNode: FocusNode(),
                          child: TextField(
                            readOnly: true,
                            textInputAction: TextInputAction.next,
                            controller: TextEditingController(
                              text: payCreditCard[i].book_bank_code,
                            ),
                            textAlign: TextAlign.left,
                            textCapitalization: TextCapitalization.characters,
                            decoration: InputDecoration(
                              enabledBorder: OutlineInputBorder(
                                borderSide: BorderSide(
                                  color: global.theme.dividerBorderColor,
                                  width: 0.0,
                                ),
                              ),
                              floatingLabelBehavior:
                                  FloatingLabelBehavior.always,
                              suffixIcon: Row(
                                mainAxisAlignment: MainAxisAlignment
                                    .spaceBetween, // added line
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  IconButton(
                                    focusNode: FocusNode(skipTraversal: true),
                                    icon: Icon(Icons.search),
                                    onPressed: () {
                                      bookBankSearch().then((value) {
                                        if (value != null) {
                                          payCreditCard[i].book_bank_code =
                                              value.passbook;
                                          payCreditCard[i].bank_code =
                                              value.bankcode;
                                          payCreditCard[i].bank_name =
                                              value.banknames![0].name;
                                          setState(() {});
                                        }
                                      });
                                    },
                                  ),
                                ],
                              ),
                              border: OutlineInputBorder(),
                              labelText: global.language("pass_book"),
                            ),
                          ),
                        ),
                      ),
                      SizedBox(width: 5),
                      Expanded(
                        child: TextField(
                          readOnly: true,
                          decoration: InputDecoration(
                            border: OutlineInputBorder(),
                            labelText: global.language("bank"),
                            floatingLabelBehavior: FloatingLabelBehavior.always,
                          ),
                          controller: TextEditingController(
                            text: (payCreditCard[i].bank_code != '')
                                ? " ${payCreditCard[i].bank_code} ~ ${payCreditCard[i].bank_name ?? []}"
                                : "",
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
                Container(
                  margin: const EdgeInsets.only(top: 10),
                  child: Row(
                    children: [
                      Expanded(
                        child: TextFormField(
                          textAlign: TextAlign.center, // Center-align the text
                          decoration: const InputDecoration(
                            border: OutlineInputBorder(),
                          ),
                          controller: payCreditCardAmountController[i],
                          keyboardType: const TextInputType.numberWithOptions(
                            decimal: true,
                          ),
                          inputFormatters: [global.NumberInputFormatter()],
                          onChanged: (value) {
                            if (value == '' && value.isEmpty) {
                              payCreditCardAmountController[i].text = "0";
                              payCreditCard[i].amount = 0;
                            } else {
                              payCreditCardAmountController[i]
                                  .value = TextEditingValue(
                                text: value.toUpperCase(),
                                selection:
                                    payCreditCardAmountController[i].selection,
                              );
                              payCreditCard[i].amount = double.parse(
                                value.replaceAll(',', ''),
                              );
                            }

                            _calPayTotal();
                          },
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      );
    }

    /// widget for cheque
    for (var i = 0; i < payCheque.length; i++) {
      listCheque.add(
        Card(
          child: Container(
            padding: EdgeInsets.all(10),
            child: Column(
              children: [
                Padding(
                  padding: EdgeInsets.only(bottom: 10),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        "${global.language("list_cheque")} ${i + 1}",
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      Container(
                        margin: const EdgeInsets.only(left: 15),
                        child: IconButton(
                          onPressed: () {
                            payCheque.removeAt(i);
                            chequeDateController.removeAt(i);
                            chequeDueDateDateController.removeAt(i);
                            payChequeAmountController.removeAt(i);
                            _calPayTotal();
                          },
                          icon: Icon(Icons.delete),
                        ),
                      ),
                    ],
                  ),
                ),
                Row(
                  children: [
                    Expanded(
                      child: TextField(
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
                                icon: Icon(Icons.calendar_month),
                                onPressed: () {
                                  selectPayDate(context, i).then((value) {
                                    if (value != null) {
                                      if (global.profileData.yeartype ==
                                          "buddhist") {
                                        chequeDateController[i].text = global
                                            .dateTimeBuddhist(
                                              value,
                                              format: global
                                                  .DateTimeFormatEnum
                                                  .dateDay,
                                            );
                                      } else {
                                        chequeDateController[i].text =
                                            DateFormat(
                                              'dd/MM/yyyy',
                                            ).format(value);
                                      }
                                      payCheque[i].doc_date_time = value
                                          .toLocal();

                                      setState(() {});
                                    }
                                  });
                                },
                              ),
                            ],
                          ),
                        ),
                        controller: chequeDateController[i],
                        onChanged: (value) {
                          setState(() {
                            docDateTimeValidated = false;
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
                              if (transferDateController[i].text
                                  .trim()
                                  .isEmpty) {
                                if (global.isValidDate(value)) {
                                  payCheque[i].doc_date_time = DateTime.parse(
                                    value,
                                  ).toLocal();
                                  docDateTimeValidated = true;
                                }
                              } else {
                                if (global.isValidDate(value)) {
                                  payCheque[i].doc_date_time = DateTime.parse(
                                    '$value ${chequeDateController[i].text}',
                                  ).toLocal();
                                  docDateTimeValidated = true;
                                }
                              }
                            } catch (e) {
                              // print(e);
                            }
                          });
                        },
                        onSubmitted: (value) {
                          chequeDateController[i].text = DateFormat(
                            'dd/MM/yyyy',
                          ).format(payCheque[i].doc_date_time!);
                        },
                      ),
                    ),
                    const SizedBox(width: 5),
                    Expanded(
                      child: RawKeyboardListener(
                        focusNode: FocusNode(),
                        child: TextField(
                          readOnly: true,
                          textInputAction: TextInputAction.next,
                          controller: TextEditingController(
                            text: payCheque[i].book_bank_code,
                          ),
                          textAlign: TextAlign.left,
                          textCapitalization: TextCapitalization.characters,
                          decoration: InputDecoration(
                            enabledBorder: OutlineInputBorder(
                              borderSide: BorderSide(
                                color: global.theme.dividerBorderColor,
                                width: 0.0,
                              ),
                            ),
                            floatingLabelBehavior: FloatingLabelBehavior.always,
                            suffixIcon: Row(
                              mainAxisAlignment:
                                  MainAxisAlignment.spaceBetween, // added line
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                IconButton(
                                  focusNode: FocusNode(skipTraversal: true),
                                  icon: Icon(Icons.search),
                                  onPressed: () {
                                    bookBankSearch().then((value) {
                                      if (value != null) {
                                        payCheque[i].book_bank_code =
                                            value.passbook;
                                        payCheque[i].bank_code = value.bankcode;
                                        payCheque[i].bank_name =
                                            value.banknames![0].name;
                                        setState(() {});
                                      }
                                    });
                                  },
                                ),
                              ],
                            ),
                            border: OutlineInputBorder(),
                            labelText: global.language("pass_book"),
                          ),
                        ),
                      ),
                    ),
                    SizedBox(width: 5),
                    Expanded(
                      child: TextField(
                        readOnly: true,
                        decoration: InputDecoration(
                          border: OutlineInputBorder(),
                          labelText: global.language("bank"),
                          floatingLabelBehavior: FloatingLabelBehavior.always,
                        ),
                        controller: TextEditingController(
                          text: (payCheque[i].bank_code != '')
                              ? " ${payCheque[i].bank_code} ~ ${payCheque[i].bank_name ?? []}"
                              : "",
                        ),
                      ),
                    ),
                  ],
                ),
                Container(
                  margin: EdgeInsets.only(top: 10),
                  child: Row(
                    children: [
                      Expanded(
                        child: TextField(
                          decoration: InputDecoration(
                            border: OutlineInputBorder(),
                            labelText: global.language("due_date"),
                            suffixIcon: Row(
                              mainAxisAlignment:
                                  MainAxisAlignment.spaceBetween, // added line
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                IconButton(
                                  focusNode: FocusNode(skipTraversal: true),
                                  icon: Icon(Icons.calendar_month),
                                  onPressed: () {
                                    selectPayDate(context, i).then((value) {
                                      if (value != null) {
                                        if (global.profileData.yeartype ==
                                            "buddhist") {
                                          chequeDueDateDateController[i].text =
                                              global.dateTimeBuddhist(
                                                value,
                                                format: global
                                                    .DateTimeFormatEnum
                                                    .dateDay,
                                              );
                                        } else {
                                          chequeDueDateDateController[i].text =
                                              DateFormat(
                                                'dd/MM/yyyy',
                                              ).format(value);
                                        }
                                        payCheque[i].due_date = value.toLocal();

                                        setState(() {});
                                      }
                                    });
                                  },
                                ),
                              ],
                            ),
                          ),
                          controller: chequeDueDateDateController[i],
                          onChanged: (value) {
                            setState(() {
                              docDateTimeValidated = false;
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
                                if (transferDateController[i].text
                                    .trim()
                                    .isEmpty) {
                                  if (global.isValidDate(value)) {
                                    payCheque[i].due_date = DateTime.parse(
                                      value,
                                    ).toLocal();
                                    docDateTimeValidated = true;
                                  }
                                } else {
                                  if (global.isValidDate(value)) {
                                    payCheque[i].due_date = DateTime.parse(
                                      '$value ${chequeDueDateDateController[i].text}',
                                    ).toLocal();
                                    docDateTimeValidated = true;
                                  }
                                }
                              } catch (e) {
                                // print(e);
                              }
                            });
                          },
                          onSubmitted: (value) {
                            chequeDueDateDateController[i].text = DateFormat(
                              'dd/MM/yyyy',
                            ).format(payCheque[i].due_date!);
                          },
                        ),
                      ),
                      SizedBox(width: 5),
                      Expanded(
                        child: TextField(
                          decoration: InputDecoration(
                            border: OutlineInputBorder(),
                            labelText: global.language("cheque_number"),
                            floatingLabelBehavior: FloatingLabelBehavior.always,
                          ),
                          controller: TextEditingController(
                            text: payCheque[i].cheque_number,
                          ),
                          onChanged: (value) {
                            payCheque[i].card_number = value;
                          },
                          onSubmitted: (value) {},
                        ),
                      ),
                    ],
                  ),
                ),
                Container(
                  margin: const EdgeInsets.only(top: 10),
                  child: Row(
                    children: [
                      Expanded(
                        child: TextFormField(
                          textAlign: TextAlign.center, // Center-align the text
                          decoration: const InputDecoration(
                            border: OutlineInputBorder(),
                          ),
                          controller: payChequeAmountController[i],
                          keyboardType: const TextInputType.numberWithOptions(
                            decimal: true,
                          ),
                          inputFormatters: [global.NumberInputFormatter()],
                          onChanged: (value) {
                            if (value == '' && value.isEmpty) {
                              payChequeAmountController[i].text = "0";
                              payCheque[i].amount = 0;
                            } else {
                              payChequeAmountController[i].value =
                                  TextEditingValue(
                                    text: value.toUpperCase(),
                                    selection:
                                        payChequeAmountController[i].selection,
                                  );
                              payCheque[i].amount = double.parse(
                                value.replaceAll(',', ''),
                              );
                            }

                            _calPayTotal();
                          },
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      );
    }

    /// widget for coupon
    for (var i = 0; i < payCoupon.length; i++) {
      listCoupon.add(
        Card(
          child: Container(
            padding: EdgeInsets.all(10),
            child: Column(
              children: [
                Padding(
                  padding: EdgeInsets.only(bottom: 10),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        "${global.language("list_coupon")} ${i + 1}",
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      Container(
                        margin: const EdgeInsets.only(left: 15),
                        child: IconButton(
                          onPressed: () {
                            payCoupon.removeAt(i);
                            couponDateController.removeAt(i);
                            payCouponAmountController.removeAt(i);
                            _calPayTotal();
                          },
                          icon: Icon(Icons.delete),
                        ),
                      ),
                    ],
                  ),
                ),
                Row(
                  children: [
                    Expanded(
                      child: TextField(
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
                                icon: Icon(Icons.calendar_month),
                                onPressed: () {
                                  selectPayDate(context, i).then((value) {
                                    if (value != null) {
                                      if (global.profileData.yeartype ==
                                          "buddhist") {
                                        couponDateController[i].text = global
                                            .dateTimeBuddhist(
                                              value,
                                              format: global
                                                  .DateTimeFormatEnum
                                                  .dateDay,
                                            );
                                      } else {
                                        couponDateController[i].text =
                                            DateFormat(
                                              'dd/MM/yyyy',
                                            ).format(value);
                                      }
                                      payCoupon[i].doc_date_time = value
                                          .toLocal();

                                      setState(() {});
                                    }
                                  });
                                },
                              ),
                            ],
                          ),
                        ),
                        controller: couponDateController[i],
                        onChanged: (value) {
                          setState(() {
                            docDateTimeValidated = false;
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
                              if (couponDateController[i].text.trim().isEmpty) {
                                if (global.isValidDate(value)) {
                                  payCoupon[i].doc_date_time = DateTime.parse(
                                    value,
                                  ).toLocal();
                                  docDateTimeValidated = true;
                                }
                              } else {
                                if (global.isValidDate(value)) {
                                  payCoupon[i].doc_date_time = DateTime.parse(
                                    '$value ${couponDateController[i].text}',
                                  ).toLocal();
                                  docDateTimeValidated = true;
                                }
                              }
                            } catch (e) {
                              // print(e);
                            }
                          });
                        },
                        onSubmitted: (value) {
                          couponDateController[i].text = DateFormat(
                            'dd/MM/yyyy',
                          ).format(payCoupon[i].doc_date_time!);
                        },
                      ),
                    ),
                    SizedBox(width: 5),
                    Expanded(
                      child: TextField(
                        decoration: InputDecoration(
                          border: OutlineInputBorder(),
                          labelText: global.language("coupon_number"),
                          floatingLabelBehavior: FloatingLabelBehavior.always,
                        ),
                        controller: TextEditingController(
                          text: payCoupon[i].number,
                        ),
                        onChanged: (value) {
                          payCoupon[i].number = value;
                        },
                        onSubmitted: (value) {},
                      ),
                    ),
                  ],
                ),
                Container(
                  margin: EdgeInsets.only(top: 10),
                  child: TextField(
                    decoration: InputDecoration(
                      border: OutlineInputBorder(),
                      labelText: global.language("description"),
                      floatingLabelBehavior: FloatingLabelBehavior.always,
                    ),
                    controller: TextEditingController(
                      text: payCoupon[i].description,
                    ),
                    onChanged: (value) {
                      payCoupon[i].description = value;
                    },
                    onSubmitted: (value) {},
                  ),
                ),
                Container(
                  margin: const EdgeInsets.only(top: 10),
                  child: Row(
                    children: [
                      Expanded(
                        child: TextFormField(
                          textAlign: TextAlign.center, // Center-align the text
                          decoration: const InputDecoration(
                            border: OutlineInputBorder(),
                          ),
                          controller: payCouponAmountController[i],
                          keyboardType: const TextInputType.numberWithOptions(
                            decimal: true,
                          ),
                          inputFormatters: [global.NumberInputFormatter()],
                          onChanged: (value) {
                            if (value == '' && value.isEmpty) {
                              payCouponAmountController[i].text = "0";
                              payCoupon[i].amount = 0;
                            } else {
                              payCouponAmountController[i].value =
                                  TextEditingValue(
                                    text: value.toUpperCase(),
                                    selection:
                                        payCouponAmountController[i].selection,
                                  );
                              payCoupon[i].amount = double.parse(
                                value.replaceAll(',', ''),
                              );
                            }

                            _calPayTotal();
                          },
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      );
    }

    /// widget for Qr code
    for (var i = 0; i < payQr.length; i++) {
      listQr.add(
        Card(
          child: Container(
            padding: EdgeInsets.all(10),
            child: Column(
              children: [
                Padding(
                  padding: EdgeInsets.only(bottom: 10),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Text(
                        "${global.language("list_qr")} ${i + 1}",
                        style: TextStyle(
                          fontSize: 16,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      Container(
                        margin: const EdgeInsets.only(left: 15),
                        child: IconButton(
                          onPressed: () {
                            payQr.removeAt(i);
                            qrDateController.removeAt(i);
                            payQrAmountController.removeAt(i);
                            _calPayTotal();
                          },
                          icon: Icon(Icons.delete),
                        ),
                      ),
                    ],
                  ),
                ),
                Row(
                  children: [
                    Expanded(
                      child: TextField(
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
                                icon: Icon(Icons.calendar_month),
                                onPressed: () {
                                  selectPayDate(context, i).then((value) {
                                    if (value != null) {
                                      if (global.profileData.yeartype ==
                                          "buddhist") {
                                        qrDateController[i].text = global
                                            .dateTimeBuddhist(
                                              value,
                                              format: global
                                                  .DateTimeFormatEnum
                                                  .dateDay,
                                            );
                                      } else {
                                        qrDateController[i].text = DateFormat(
                                          'dd/MM/yyyy',
                                        ).format(value);
                                      }
                                      payQr[i].doc_date_time = value.toLocal();

                                      setState(() {});
                                    }
                                  });
                                },
                              ),
                            ],
                          ),
                        ),
                        controller: qrDateController[i],
                        onChanged: (value) {
                          setState(() {
                            docDateTimeValidated = false;
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
                              if (qrDateController[i].text.trim().isEmpty) {
                                if (global.isValidDate(value)) {
                                  payQr[i].doc_date_time = DateTime.parse(
                                    value,
                                  ).toLocal();
                                  docDateTimeValidated = true;
                                }
                              } else {
                                if (global.isValidDate(value)) {
                                  payQr[i].doc_date_time = DateTime.parse(
                                    '$value ${qrDateController[i].text}',
                                  ).toLocal();
                                  docDateTimeValidated = true;
                                }
                              }
                            } catch (e) {
                              // print(e);
                            }
                          });
                        },
                        onSubmitted: (value) {
                          qrDateController[i].text = DateFormat(
                            'dd/MM/yyyy',
                          ).format(payQr[i].doc_date_time!);
                        },
                      ),
                    ),
                    SizedBox(width: 5),
                    Expanded(
                      child: TextField(
                        decoration: InputDecoration(
                          border: OutlineInputBorder(),
                          labelText: global.language("provider_code"),
                          floatingLabelBehavior: FloatingLabelBehavior.always,
                        ),
                        controller: TextEditingController(
                          text: payQr[i].provider_code,
                        ),
                        onChanged: (value) {
                          payQr[i].provider_code = value;
                        },
                        onSubmitted: (value) {},
                      ),
                    ),
                    SizedBox(width: 5),
                    Expanded(
                      child: TextField(
                        decoration: InputDecoration(
                          border: OutlineInputBorder(),
                          labelText: global.language("provider_name"),
                          floatingLabelBehavior: FloatingLabelBehavior.always,
                        ),
                        controller: TextEditingController(
                          text: payQr[i].provider_name,
                        ),
                        onChanged: (value) {
                          payQr[i].provider_name = value;
                        },
                        onSubmitted: (value) {},
                      ),
                    ),
                  ],
                ),
                Container(
                  margin: const EdgeInsets.only(top: 10),
                  child: Row(
                    children: [
                      Expanded(
                        child: TextFormField(
                          textAlign: TextAlign.center, // Center-align the text
                          decoration: const InputDecoration(
                            border: OutlineInputBorder(),
                          ),
                          controller: payQrAmountController[i],
                          keyboardType: const TextInputType.numberWithOptions(
                            decimal: true,
                          ),
                          inputFormatters: [global.NumberInputFormatter()],
                          onChanged: (value) {
                            if (value == '' && value.isEmpty) {
                              payQrAmountController[i].text = "0";
                              payQr[i].amount = 0;
                            } else {
                              payQrAmountController[i].value = TextEditingValue(
                                text: value.toUpperCase(),
                                selection: payQrAmountController[i].selection,
                              );
                              payQr[i].amount = double.parse(
                                value.replaceAll(',', ''),
                              );
                            }

                            _calPayTotal();
                          },
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      );
    }

    paymentDetail.add(
      Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                /// pay menu
                payMenuWidget(),
                (showPayDetail == 0)
                    ? payCashWidget()
                    : (showPayDetail == 1)
                    ? payTransferWidget(listTransfer)
                    : (showPayDetail == 2)
                    ? payCreditCardWidget(listCredit)
                    : (showPayDetail == 3)
                    ? payChequeWidget(listCheque)
                    : (showPayDetail == 4)
                    ? payCouponWidget(listCoupon)
                    : (showPayDetail == 5)
                    ? payQrWidget(listQr)
                    : Container(),
              ],
            ),
          ),
          SizedBox(width: 10),
          Expanded(
            child: Column(
              children: [
                Container(
                  margin: EdgeInsets.only(bottom: 10),
                  child: TextFormField(
                    readOnly: true,
                    enabled: true,
                    decoration: InputDecoration(
                      border: OutlineInputBorder(),
                      labelText: global.language("cash"),
                      suffixIcon: const IconButton(
                        icon: Icon(Icons.money),
                        onPressed: null,
                      ),
                    ),
                    controller: payCashAmountController,
                    keyboardType: const TextInputType.numberWithOptions(
                      decimal: true,
                    ),
                    inputFormatters: [global.NumberInputFormatter()],
                  ),
                ),
                Container(
                  margin: EdgeInsets.only(bottom: 10),
                  child: TextField(
                    readOnly: true,
                    enabled: true,
                    decoration: InputDecoration(
                      border: OutlineInputBorder(),
                      labelText: global.language("transfer"),
                      suffixIcon: const IconButton(
                        icon: Icon(Icons.transform_rounded),
                        onPressed: null,
                      ),
                    ),
                    controller: TextEditingController(
                      text: global.formatNumber(screenData.summoneytransfer!),
                    ),
                  ),
                ),
                (widget.type == global.TransactionTypeEnum.paid)
                    ? Container(
                        margin: EdgeInsets.only(bottom: 10),
                        child: TextField(
                          readOnly: true,
                          enabled: true,
                          decoration: InputDecoration(
                            border: OutlineInputBorder(),
                            labelText: global.language("credit_card"),
                            suffixIcon: const IconButton(
                              icon: Icon(Icons.credit_card),
                              onPressed: null,
                            ),
                          ),
                          controller: TextEditingController(
                            text: global.formatNumber(
                              screenData.sumcreditcard!,
                            ),
                          ),
                        ),
                      )
                    : Container(),
                Container(
                  margin: EdgeInsets.only(bottom: 10),
                  child: TextField(
                    readOnly: true,
                    enabled: true,
                    decoration: InputDecoration(
                      border: OutlineInputBorder(),
                      labelText: global.language("cheque"),
                      suffixIcon: const IconButton(
                        icon: Icon(Icons.featured_play_list_outlined),
                        onPressed: null,
                      ),
                    ),
                    controller: TextEditingController(
                      text: global.formatNumber(screenData.sumcheque!),
                    ),
                  ),
                ),
                Container(
                  margin: EdgeInsets.only(bottom: 10),
                  child: TextField(
                    readOnly: true,
                    enabled: true,
                    decoration: InputDecoration(
                      border: OutlineInputBorder(),
                      labelText: global.language("coupon"),
                      suffixIcon: const IconButton(
                        icon: Icon(Icons.card_giftcard),
                        onPressed: null,
                      ),
                    ),
                    controller: TextEditingController(
                      text: global.formatNumber(screenData.sumcoupon!),
                    ),
                  ),
                ),
                (widget.type == global.TransactionTypeEnum.paid)
                    ? Container(
                        margin: EdgeInsets.only(bottom: 10),
                        child: TextField(
                          readOnly: true,
                          enabled: true,
                          decoration: InputDecoration(
                            border: OutlineInputBorder(),
                            labelText: global.language("qr_code"),
                            suffixIcon: const IconButton(
                              icon: Icon(Icons.qr_code),
                              onPressed: null,
                            ),
                          ),
                          controller: TextEditingController(
                            text: global.formatNumber(screenData.sumqrcode!),
                          ),
                        ),
                      )
                    : Container(),
              ],
            ),
          ),
        ],
      ),
    );
    return SingleChildScrollView(
      child: Container(
        width: double.infinity,
        padding: EdgeInsets.all(15),
        child: Column(
          children: [
            Container(
              margin: EdgeInsets.only(bottom: 10),
              child: Row(
                children: [
                  Expanded(
                    child: TotalTextController(
                      readOnly: true,
                      title: global.language("doc_total_amount"),
                      data: screenData.totalpaymentamount,
                      icon: null,
                      onChanged: (value) {
                        if (value != '') {
                          screenData.totalpaymentamount = double.parse(
                            value.replaceAll(',', ''),
                          );
                        }
                      },
                    ),
                  ),
                  SizedBox(width: 10),
                  Expanded(
                    child: TextFormField(
                      readOnly: false,
                      decoration: InputDecoration(
                        border: OutlineInputBorder(),
                        labelText: global.language("round_amount"),
                      ),
                      controller: roundAmountController,
                      keyboardType: const TextInputType.numberWithOptions(
                        decimal: true,
                      ),
                      inputFormatters: [global.NumberInputFormatter()],
                      onChanged: (value) {
                        if (value != '' && value.isNotEmpty) {
                          roundAmountController.value = TextEditingValue(
                            text: value,
                            selection: roundAmountController.selection,
                          );

                          /// RegExp 0-9 if macth to function  _calPayTotal();
                          RegExp regExp = RegExp(r'[0-9]');
                          if (regExp.hasMatch(value)) {
                            _calPayTotal();
                          }
                        }
                      },
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
                    child: GestureDetector(
                      onTap: () async {
                        final result = await showNumericInputDialog(
                          context,
                          title: global.language("sum_pay"),
                          initialValue:
                              (screenData.totalpaymentamount +
                                      (screenData.roundamount!))
                                  .toString(),
                          decimalPlaces: 2,
                        );

                        if (result != null && result.isNotEmpty) {
                          setState(() {
                            screenData.totalpaymentamount =
                                double.parse(result.replaceAll(',', '')) -
                                (screenData.roundamount!);
                            _calPayTotal();
                          });
                        }
                      },
                      child: TotalTextController(
                        readOnly: true,
                        title: global.language("sum_pay"),
                        data:
                            screenData.totalpaymentamount +
                            (screenData.roundamount!),
                        icon: null,
                        useColor: true,
                        onChanged: (value) {},
                      ),
                    ),
                  ),
                  SizedBox(width: 10),
                  Expanded(
                    child: GestureDetector(
                      onTap: () async {
                        final result = await showNumericInputDialog(
                          context,
                          title: global.language("total_amount"),
                          initialValue: payTotalBill.toString(),
                          decimalPlaces: 2,
                        );

                        if (result != null && result.isNotEmpty) {
                          setState(() {
                            payTotalBill = double.parse(
                              result.replaceAll(',', ''),
                            );
                            _calPayTotal();
                          });
                        }
                      },
                      child: TotalTextController(
                        readOnly: true,
                        title: global.language("total_amount"),
                        data: payTotalBill,
                        icon: null,
                        useColor: true,
                        onChanged: (value) {},
                      ),
                    ),
                  ),
                ],
              ),
            ),
            Column(children: paymentDetail),
            Container(
              margin: const EdgeInsets.only(top: 10),
              child: Row(
                children: [
                  Expanded(
                    child: SizedBox(
                      height: 50,
                      child: ElevatedButton.icon(
                        onPressed: () {
                          screenData.paymentStructs = [];
                          if (payTransfer.isNotEmpty) {
                            for (var i = 0; i < payTransfer.length; i++) {
                              screenData.paymentStructs!.addAll(
                                payTransfer,
                              );
                            }
                          }
                          if (payCreditCard.isNotEmpty) {
                            for (var i = 0; i < payCreditCard.length; i++) {
                              screenData.paymentStructs!.addAll(
                                payCreditCard,
                              );
                            }
                          }

                          if (payCheque.isNotEmpty) {
                            for (var i = 0; i < payCheque.length; i++) {
                              screenData.paymentStructs!.addAll(
                                payCheque,
                              );
                            }
                          }

                          if (payCoupon.isNotEmpty) {
                            for (var i = 0; i < payCoupon.length; i++) {
                              screenData.paymentStructs!.addAll(
                                payCoupon,
                              );
                            }
                          }

                          if (payQr.isNotEmpty) {
                            for (var i = 0; i < payQr.length; i++) {
                              screenData.paymentStructs!.addAll(payQr);
                            }
                          }

                          Navigator.pop(context, ['null', screenData]);
                        },
                        icon: Icon(Icons.arrow_back),
                        label: Text(global.language("back")),

                        /// set color button
                        style: ElevatedButton.styleFrom(
                          foregroundColor: global.theme.onPrimaryColor,
                          backgroundColor: global.theme.iconSecondaryColor,
                        ),
                      ),
                    ),
                  ),
                  const SizedBox(width: 10),
                  Expanded(
                    child: SizedBox(
                      height: 50,
                      child: ElevatedButton.icon(
                        onPressed: () {
                          if (verifyPayment()) {
                            screenData.paymentStructs = [];
                            if (payTransfer.isNotEmpty) {
                              for (var i = 0; i < payTransfer.length; i++) {
                                screenData.paymentStructs!.addAll(
                                  payTransfer,
                                );
                              }
                            }
                            if (payCreditCard.isNotEmpty) {
                              for (var i = 0; i < payCreditCard.length; i++) {
                                screenData.paymentStructs!.addAll(
                                  payCreditCard,
                                );
                              }
                            }

                            if (payCheque.isNotEmpty) {
                              for (var i = 0; i < payCheque.length; i++) {
                                screenData.paymentStructs!.addAll(
                                  payCheque,
                                );
                              }
                            }

                            if (payCoupon.isNotEmpty) {
                              for (var i = 0; i < payCoupon.length; i++) {
                                screenData.paymentStructs!.addAll(
                                  payCoupon,
                                );
                              }
                            }

                            if (payQr.isNotEmpty) {
                              for (var i = 0; i < payQr.length; i++) {
                                screenData.paymentStructs!.addAll(
                                  payQr,
                                );
                              }
                            }

                            Navigator.pop(context, ['save', screenData]);
                          }
                        },
                        icon: Icon(Icons.save),
                        label: Text(global.language("save")),
                      ),
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

  bool verifyPayment() {
    List<String> errorList = [];

    if (payTotalBill >
        screenData.totalpaymentamount + (screenData.roundamount!)) {
      errorList.add(global.language("payment_over"));
    } else if (payTotalBill <
        screenData.totalpaymentamount + (screenData.roundamount!)) {
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
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language("payment_details")),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: Icon(Icons.arrow_back),
          onPressed: () {
            screenData.paymentStructs = [];
            if (payTransfer.isNotEmpty) {
              for (var i = 0; i < payTransfer.length; i++) {
                screenData.paymentStructs!.addAll(payTransfer);
              }
            }
            if (payCreditCard.isNotEmpty) {
              for (var i = 0; i < payCreditCard.length; i++) {
                screenData.paymentStructs!.addAll(payCreditCard);
              }
            }

            if (payCheque.isNotEmpty) {
              for (var i = 0; i < payCheque.length; i++) {
                screenData.paymentStructs!.addAll(payCheque);
              }
            }

            if (payCoupon.isNotEmpty) {
              for (var i = 0; i < payCoupon.length; i++) {
                screenData.paymentStructs!.addAll(payCoupon);
              }
            }

            if (payQr.isNotEmpty) {
              for (var i = 0; i < payQr.length; i++) {
                screenData.paymentStructs!.addAll(payQr);
              }
            }

            Navigator.pop(context, ['null', screenData]);
          },
        ),

        /// button save
        actions: [
          EditFontSizeControl(onChanged: () => setState(() {})),
          IconButton(
            focusNode: FocusNode(skipTraversal: true),
            icon: Icon(Icons.save, size: 26.0),
            onPressed: () {
              if (verifyPayment()) {
                screenData.paymentStructs = [];
                if (payTransfer.isNotEmpty) {
                  for (var i = 0; i < payTransfer.length; i++) {
                    screenData.paymentStructs!.addAll(payTransfer);
                  }
                }
                if (payCreditCard.isNotEmpty) {
                  for (var i = 0; i < payCreditCard.length; i++) {
                    screenData.paymentStructs!.addAll(payCreditCard);
                  }
                }
                if (payCheque.isNotEmpty) {
                  for (var i = 0; i < payCheque.length; i++) {
                    screenData.paymentStructs!.addAll(payCheque);
                  }
                }
                if (payCoupon.isNotEmpty) {
                  for (var i = 0; i < payCoupon.length; i++) {
                    screenData.paymentStructs!.addAll(payCoupon);
                  }
                }
                if (payQr.isNotEmpty) {
                  for (var i = 0; i < payQr.length; i++) {
                    screenData.paymentStructs!.addAll(payQr);
                  }
                }
                Navigator.pop(context, ['save', screenData]);
              }
            },
          ),
        ],
      ),
      body: Builder(
        builder: (context) {
          final scaleFactor = global.editFontScaleFactor;
          return MediaQuery(
            data: MediaQuery.of(context).copyWith(
              textScaler: TextScaler.linear(scaleFactor),
            ),
            child: Theme(
              data: Theme.of(context).copyWith(
                inputDecorationTheme: Theme.of(context).inputDecorationTheme.copyWith(
                  contentPadding: EdgeInsets.symmetric(
                    horizontal: 12 * scaleFactor,
                    vertical: 10 * scaleFactor,
                  ),
                ),
                iconTheme: IconThemeData(size: 24 * scaleFactor),
              ),
              child: editSummeryWidget(),
            ),
          );
        },
      ),
    );
  }
}

class TotalTextController extends StatelessWidget {
  const TotalTextController({
    super.key,
    required this.readOnly,
    required this.title,
    required this.data,
    this.icon,
    required this.onChanged,
    this.useColor,
    this.enabled,
  });

  final dynamic data;
  final String title;
  final Icon? icon;
  final Function(String) onChanged;
  final bool readOnly;
  final bool? useColor;
  final bool? enabled;

  @override
  Widget build(BuildContext context) {
    String initialText = '';
    if (data != null) {
      if (data is String) {
        try {
          // Attempt to parse the string as a double
          final double parsedValue = double.parse(data);
          initialText = global.formatNumber(parsedValue);
        } catch (e) {
          // Handle the case where parsing fails (e.g., if the string is not a valid number)
          initialText =
              ''; // Set to an empty string or any other default value you prefer
        }
      } else if (data is double) {
        // If data is already a double, just format it
        initialText = global.formatNumber(data);
      }
    }
    return TextField(
      readOnly: readOnly,
      enabled: enabled,
      decoration: InputDecoration(
        border: const OutlineInputBorder(),
        labelText: title,
        floatingLabelBehavior: FloatingLabelBehavior.always,
        suffixIcon: icon,
        filled: useColor,
        fillColor: Colors.yellow[100],
      ),
      controller: TextEditingController(text: initialText),
      onChanged: onChanged,
      keyboardType: const TextInputType.numberWithOptions(decimal: true),
      inputFormatters: [global.NumberInputFormatter()],
    );
  }
}
