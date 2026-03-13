// ignore_for_file: depend_on_referenced_packages

import 'package:flutter/material.dart';
import 'package:intl/intl.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/transaction_model.dart';

/// Helper class สำหรับ Payment Widgets
/// ใช้ class นี้เพื่อแยก payment widget logic ออกจาก main screen
/// รองรับทั้ง TransactionEditScreenState และ PurchaseOrderEditScreenState
class PaymentWidgets {
  final dynamic state;

  PaymentWidgets(this.state);

  /// แสดงเมนูเลือกประเภทการชำระเงิน
  Widget buildPayMenuWidget() {
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
                      state.showPayDetail = 0;
                      state.setState(() {});
                    },
                    icon: Icon(Icons.money),
                    label: Text(global.language("cash")),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: (state.showPayDetail == 0)
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
                      state.showPayDetail = 1;
                      state.setState(() {});
                    },
                    icon: Icon(Icons.transform_rounded),
                    label: Text(global.language("money_transfer")),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: (state.showPayDetail == 1)
                          ? global.theme.infoHighlightTextColor
                          : global.theme.iconSecondaryColor,
                    ),
                  ),
                ),
              ),

              /// เงินล่วงหน้า
              (state.widget.type == global.TransactionTypeEnum.purchase ||
                      state.widget.type ==
                          global.TransactionTypeEnum.accrualreceive)
                  ? Expanded(
                      child: Padding(
                        padding: EdgeInsets.all(5.0),
                        child: ElevatedButton.icon(
                          onPressed: () {
                            state.showPayDetail = 8;
                            state.setState(() {});
                          },
                          icon: Icon(Icons.money_off),
                          label: Text(global.language('advance_payment')),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: (state.showPayDetail == 8)
                                ? global.theme.infoHighlightTextColor
                                : global.theme.iconSecondaryColor,
                          ),
                        ),
                      ),
                    )
                  : Container(),
              (state.widget.type == global.TransactionTypeEnum.sale ||
                      state.widget.type ==
                          global.TransactionTypeEnum.paidAdvance ||
                      state.widget.type ==
                          global.TransactionTypeEnum.receiveDeposit)
                  ? Expanded(
                      child: Padding(
                        padding: EdgeInsets.all(5.0),
                        child: ElevatedButton.icon(
                          onPressed: () {
                            state.showPayDetail = 2;
                            state.setState(() {});
                          },
                          icon: Icon(Icons.credit_card),
                          label: Text(global.language("credit_card")),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: (state.showPayDetail == 2)
                                ? global.theme.infoHighlightTextColor
                                : global.theme.iconSecondaryColor,
                          ),
                        ),
                      ),
                    )
                  : Container(),
            ],
          ),
          Row(
            children: [
              (state.widget.type == global.TransactionTypeEnum.sale ||
                      state.widget.type ==
                          global.TransactionTypeEnum.saleorder ||
                      state.widget.type ==
                          global.TransactionTypeEnum.paidAdvance ||
                      state.widget.type ==
                          global.TransactionTypeEnum.receiveDeposit)
                  ? Expanded(
                      child: Padding(
                        padding: EdgeInsets.all(5.0),
                        child: ElevatedButton.icon(
                          onPressed: () {
                            state.showPayDetail = 3;
                            state.setState(() {});
                          },
                          icon: Icon(Icons.featured_play_list_outlined),
                          label: Text(global.language("cheque")),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: (state.showPayDetail == 3)
                                ? global.theme.infoHighlightTextColor
                                : global.theme.iconSecondaryColor,
                          ),
                        ),
                      ),
                    )
                  : Container(),
              (state.widget.type == global.TransactionTypeEnum.sale ||
                      state.widget.type == global.TransactionTypeEnum.saleorder)
                  ? Expanded(
                      child: Padding(
                        padding: EdgeInsets.all(5.0),
                        child: ElevatedButton.icon(
                          onPressed: () {
                            state.showPayDetail = 4;
                            state.setState(() {});
                          },
                          icon: Icon(Icons.card_giftcard),
                          label: Text(global.language("coupon")),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: (state.showPayDetail == 4)
                                ? global.theme.infoHighlightTextColor
                                : global.theme.iconSecondaryColor,
                          ),
                        ),
                      ),
                    )
                  : Container(),
              (state.widget.type == global.TransactionTypeEnum.sale ||
                      state.widget.type ==
                          global.TransactionTypeEnum.saleorder ||
                      state.widget.type ==
                          global.TransactionTypeEnum.paidAdvance ||
                      state.widget.type ==
                          global.TransactionTypeEnum.receiveDeposit)
                  ? Expanded(
                      child: Padding(
                        padding: EdgeInsets.all(5.0),
                        child: ElevatedButton.icon(
                          onPressed: () {
                            state.showPayDetail = 5;
                            state.setState(() {});
                          },
                          icon: Icon(Icons.qr_code),
                          label: Text(global.language("qr_code")),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: (state.showPayDetail == 5)
                                ? global.theme.infoHighlightTextColor
                                : global.theme.iconSecondaryColor,
                          ),
                        ),
                      ),
                    )
                  : Container(),
            ],
          ),
          (state.widget.type == global.TransactionTypeEnum.sale ||
                  state.widget.type == global.TransactionTypeEnum.saleorder)
              ? Row(
                  children: [
                    Expanded(
                      child: Padding(
                        padding: EdgeInsets.all(5.0),
                        child: ElevatedButton.icon(
                          onPressed: () {
                            state.showPayDetail = 6;
                            state.setState(() {});
                          },
                          icon: Icon(Icons.delivery_dining_rounded),
                          label: Text(global.language("delivery")),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: (state.showPayDetail == 6)
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
                            state.showPayDetail = 7;
                            state.setState(() {});
                          },
                          icon: Icon(Icons.star_rate_rounded),
                          label: Text(global.language('points')),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: (state.showPayDetail == 7)
                                ? global.theme.infoHighlightTextColor
                                : global.theme.iconSecondaryColor,
                          ),
                        ),
                      ),
                    ),

                    /// เงินล่วงหน้า เพื่อให้ครบ 3 กล่อง สำหรับ sale
                    (state.widget.type == global.TransactionTypeEnum.sale)
                        ? Expanded(
                            child: Padding(
                              padding: EdgeInsets.all(5.0),
                              child: ElevatedButton.icon(
                                onPressed: () {
                                  state.showPayDetail = 8;
                                  state.setState(() {});
                                },
                                icon: Icon(Icons.money_off),
                                label: Text(global.language('advance_payment')),
                                style: ElevatedButton.styleFrom(
                                  backgroundColor: (state.showPayDetail == 8)
                                      ? global.theme.infoHighlightTextColor
                                      : global.theme.iconSecondaryColor,
                                ),
                              ),
                            ),
                          )
                        : Container(),
                  ],
                )
              : Container(),
        ],
      ),
    );
  }

  /// แสดงฟอร์มชำระด้วยเงินสด
  Widget buildPayCashWidget() {
    return Center(
      child: Padding(
        padding: EdgeInsets.all(30.0),
        child: Column(
          children: [
            Container(
              padding: EdgeInsets.all(16),
              child: TextFormField(
                textAlign: TextAlign.center,
                enabled:
                    (state.widget.type ==
                            global.TransactionTypeEnum.purchasereturn ||
                        state.widget.type ==
                            global.TransactionTypeEnum.salereturn)
                    ? state.screenData.inquirytype != 0 &&
                          state.screenData.inquirytype != 1
                    : state.screenData.inquirytype != 0,
                decoration: InputDecoration(
                  border: OutlineInputBorder(),
                  labelText: global.language("cash"),
                ),
                style: TextStyle(fontSize: 28),
                controller: state.payCashAmountController,
                keyboardType: const TextInputType.numberWithOptions(
                  decimal: true,
                ),
                inputFormatters: [global.NumberInputFormatter()],
                onChanged: (value) {
                  if (value == '' && value.isEmpty) {
                    state.payCashAmountController.text = "0";
                  } else {
                    state.payCashAmountController.value = TextEditingValue(
                      text: value.toUpperCase(),
                      selection: state.payCashAmountController.selection,
                    );
                  }
                  state.transactionCalculator.calPayTotal();
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// แสดงฟอร์มชำระเงินปลายทาง
  Widget buildPayDeliveryWidget() {
    return Center(
      child: Padding(
        padding: EdgeInsets.all(30.0),
        child: Column(
          children: [
            Container(
              padding: EdgeInsets.all(16),
              child: TextFormField(
                textAlign: TextAlign.center,
                decoration: InputDecoration(
                  border: OutlineInputBorder(),
                  labelText: global.language("cash_amount"),
                ),
                style: TextStyle(fontSize: 28),
                controller: state.payDeliveryCashAmountController,
                keyboardType: const TextInputType.numberWithOptions(
                  decimal: true,
                ),
                inputFormatters: [global.NumberInputFormatter()],
                onChanged: (value) {
                  if (value == '' && value.isEmpty) {
                    state.payDeliveryCashAmountController.text = "0";
                  } else {
                    state.payDeliveryCashAmountController.value =
                        TextEditingValue(
                          text: value.toUpperCase(),
                          selection:
                              state.payDeliveryCashAmountController.selection,
                        );
                  }
                  state.transactionCalculator.calPayTotal();
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// แสดงฟอร์มชำระด้วยคะแนน
  Widget buildPayPointWidget() {
    return Center(
      child: Padding(
        padding: EdgeInsets.all(30.0),
        child: Column(
          children: [
            Container(
              padding: EdgeInsets.all(8),
              child: TextFormField(
                textAlign: TextAlign.center,
                readOnly: true,
                decoration: InputDecoration(
                  border: OutlineInputBorder(),
                  labelText: global.language('use_points'),
                ),
                style: TextStyle(fontSize: 20),
                controller: TextEditingController(
                  text: state.screenData.usepoint.toString(),
                ),
                keyboardType: const TextInputType.numberWithOptions(
                  decimal: true,
                ),
                inputFormatters: [global.NumberInputFormatter()],
              ),
            ),

            /// มูลค่าแต้ม
            Container(
              padding: EdgeInsets.all(8),
              child: TextFormField(
                textAlign: TextAlign.center,
                readOnly: true,
                decoration: InputDecoration(
                  border: OutlineInputBorder(),
                  labelText: global.language('points_value'),
                ),
                style: TextStyle(fontSize: 20),
                controller: TextEditingController(
                  text: state.screenData.paypointamount.toString(),
                ),
                keyboardType: const TextInputType.numberWithOptions(
                  decimal: true,
                ),
                inputFormatters: [global.NumberInputFormatter()],
              ),
            ),

            /// รหัสสะสมแต้ม
            Container(
              padding: EdgeInsets.all(8),
              child: TextFormField(
                textAlign: TextAlign.center,
                readOnly: true,
                decoration: InputDecoration(
                  border: OutlineInputBorder(),
                  labelText: global.language('points_code'),
                ),
                style: TextStyle(fontSize: 20),
                controller: TextEditingController(
                  text: state.screenData.pointscode,
                ),
                onChanged: (value) {
                  state.screenData.pointscode = value;
                },
              ),
            ),

            /// แต้มที่ได้รับ
            Container(
              padding: EdgeInsets.all(8),
              child: TextFormField(
                textAlign: TextAlign.center,
                readOnly: true,
                decoration: InputDecoration(
                  border: OutlineInputBorder(),
                  labelText: global.language('points_received'),
                ),
                style: TextStyle(fontSize: 20),
                controller: TextEditingController(
                  text: state.screenData.getpoint.toString(),
                ),
                keyboardType: const TextInputType.numberWithOptions(
                  decimal: true,
                ),
                inputFormatters: [global.NumberInputFormatter()],
                onChanged: (value) {
                  if (value.isNotEmpty) {
                    final numberValue = int.tryParse(value.replaceAll(',', ''));
                    if (numberValue != null) {
                      state.screenData.getpoint = numberValue;
                    }
                  } else {
                    state.screenData.getpoint = 0;
                  }
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  /// แสดงฟอร์มชำระด้วยบัตรเครดิต
  Widget buildPayCreditCardWidget(List<Widget> listCredit) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          margin: const EdgeInsets.only(left: 10, top: 5, bottom: 5),
          child: ElevatedButton.icon(
            onPressed: (state.screenData.inquirytype != 0)
                ? () {
                    state.creditCardDateController.add(TextEditingController());
                    state.payCreditCardAmountController.add(
                      TextEditingController(),
                    );

                    if (global.profileData.yeartype == "buddhist") {
                      state
                          .creditCardDateController[state
                                  .creditCardDateController
                                  .length -
                              1]
                          .text = global.dateTimeBuddhist(
                        DateTime.now(),
                        format: global.DateTimeFormatEnum.dateDay,
                      );
                    } else {
                      state
                          .creditCardDateController[state
                                  .creditCardDateController
                                  .length -
                              1]
                          .text = DateFormat('dd/MM/yyyy').format(
                        DateTime.parse(
                          DateTime.now().toUtc().toIso8601String(),
                        ),
                      );
                    }
                    state
                            .payCreditCardAmountController[state
                                    .creditCardDateController
                                    .length -
                                1]
                            .text =
                        "0";

                    state.payCreditCard.add(BillPayStruct(trans_flag: 1));

                    state.setState(() {});
                  }
                : null,
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

  /// แสดงฟอร์มชำระด้วยการโอนเงิน
  Widget buildPayTransferWidget(List<Widget> listTransfer) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          margin: const EdgeInsets.only(left: 10, top: 5, bottom: 5),
          child: ElevatedButton.icon(
            onPressed: (state.screenData.inquirytype != 0)
                ? () {
                    state.transferDateController.add(TextEditingController());
                    state.payTransferAmountController.add(
                      TextEditingController(),
                    );

                    if (global.profileData.yeartype == "buddhist") {
                      state
                          .transferDateController[state
                                  .transferDateController
                                  .length -
                              1]
                          .text = global.dateTimeBuddhist(
                        DateTime.now(),
                        format: global.DateTimeFormatEnum.dateDay,
                      );
                    } else {
                      state
                          .transferDateController[state
                                  .transferDateController
                                  .length -
                              1]
                          .text = DateFormat('dd/MM/yyyy').format(
                        DateTime.parse(
                          DateTime.now().toUtc().toIso8601String(),
                        ),
                      );
                    }
                    state
                            .payTransferAmountController[state
                                    .transferDateController
                                    .length -
                                1]
                            .text =
                        "0";

                    state.payTransfer.add(BillPayStruct(trans_flag: 2));

                    state.setState(() {});
                  }
                : null,
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

  /// แสดงฟอร์มชำระด้วยเช็ค
  Widget buildPayChequeWidget(List<Widget> listCheque) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          margin: const EdgeInsets.only(left: 10, top: 5, bottom: 5),
          child: ElevatedButton.icon(
            onPressed: (state.screenData.inquirytype != 0)
                ? () {
                    state.chequeDateController.add(TextEditingController());
                    state.chequeDueDateDateController.add(
                      TextEditingController(),
                    );
                    state.payChequeAmountController.add(
                      TextEditingController(),
                    );

                    if (global.profileData.yeartype == "buddhist") {
                      state
                          .chequeDateController[state
                                  .chequeDateController
                                  .length -
                              1]
                          .text = global.dateTimeBuddhist(
                        DateTime.now(),
                        format: global.DateTimeFormatEnum.dateDay,
                      );
                    } else {
                      state
                          .chequeDateController[state
                                  .chequeDateController
                                  .length -
                              1]
                          .text = DateFormat('dd/MM/yyyy').format(
                        DateTime.parse(
                          DateTime.now().toUtc().toIso8601String(),
                        ),
                      );
                    }

                    if (global.profileData.yeartype == "buddhist") {
                      state
                          .chequeDueDateDateController[state
                                  .chequeDueDateDateController
                                  .length -
                              1]
                          .text = global.dateTimeBuddhist(
                        DateTime.now(),
                        format: global.DateTimeFormatEnum.dateDay,
                      );
                    } else {
                      state
                          .chequeDueDateDateController[state
                                  .chequeDueDateDateController
                                  .length -
                              1]
                          .text = DateFormat('dd/MM/yyyy').format(
                        DateTime.parse(
                          DateTime.now().toUtc().toIso8601String(),
                        ),
                      );
                    }

                    state
                            .payChequeAmountController[state
                                    .chequeDueDateDateController
                                    .length -
                                1]
                            .text =
                        "0";

                    state.payCheque.add(BillPayStruct(trans_flag: 3));
                    state.setState(() {});
                  }
                : null,
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

  /// แสดงฟอร์มชำระด้วยคูปอง
  Widget buildPayCouponWidget(List<Widget> listCoupon) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          padding: const EdgeInsets.all(5),
          child: Column(children: listCoupon),
        ),
      ],
    );
  }

  /// แสดงฟอร์มชำระด้วย QR Code
  Widget buildPayQrWidget(List<Widget> listQr) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Container(
          margin: const EdgeInsets.only(left: 10, top: 5, bottom: 5),
          child: ElevatedButton.icon(
            onPressed: (state.screenData.inquirytype != 0)
                ? () {
                    state.qrDateController.add(TextEditingController());
                    state.payQrAmountController.add(TextEditingController());

                    if (global.profileData.yeartype == "buddhist") {
                      state
                          .qrDateController[state.qrDateController.length - 1]
                          .text = global.dateTimeBuddhist(
                        DateTime.now(),
                        format: global.DateTimeFormatEnum.dateDay,
                      );
                    } else {
                      state
                          .qrDateController[state.qrDateController.length - 1]
                          .text = DateFormat('dd/MM/yyyy').format(
                        DateTime.parse(
                          DateTime.now().toUtc().toIso8601String(),
                        ),
                      );
                    }

                    state
                            .payQrAmountController[state
                                    .qrDateController
                                    .length -
                                1]
                            .text =
                        "0";

                    state.payQr.add(BillPayStruct(trans_flag: 5));
                    state.setState(() {});
                  }
                : null,
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
}
