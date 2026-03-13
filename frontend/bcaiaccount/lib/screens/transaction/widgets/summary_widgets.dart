// ignore_for_file: depend_on_referenced_packages

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/screens/transaction/components/coupon_widget.dart';
import 'package:smlaicloud/screens/transaction/components/deposit_payment_widget.dart';
import 'package:smlaicloud/screens/transaction/components/pay_advance_payment_widget.dart';
import 'package:smlaicloud/screens/transaction/components/payment_summary_widget.dart';
import 'package:smlaicloud/screens/transaction/components/payment_method_item_widget.dart';
import 'package:smlaicloud/screens/transaction/components/total_text_controller.dart';
import 'package:smlaicloud/components/numpad.dart';

/// Helper class สำหรับ Summary Widgets
/// ใช้ class นี้เพื่อแยก summary widget logic ออกจาก main screen
/// รองรับทั้ง TransactionEditScreenState และ PurchaseOrderEditScreenState
class SummaryWidgets {
  final dynamic state;

  SummaryWidgets(this.state);

  /// Widget หลักสำหรับแสดงสรุปยอดและการชำระเงิน
  Widget buildEditSummaryWidget() {
    List<Widget> sumDetails = [];
    List<Widget> paymentDetail = [];

    // สร้าง sections ต่างๆ
    _buildRefTotalSection(sumDetails);
    _buildTotalValueSection(sumDetails);
    _buildDiscountSection(sumDetails);
    _buildDepositSection(sumDetails);
    _buildVatSection(sumDetails);
    _buildFinalTotalSection(sumDetails);
    _buildPaymentSection(paymentDetail);

    return SingleChildScrollView(
      child: Column(
        children: [
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(15),
            child: Column(
              children: [
                Column(children: sumDetails),
                const SizedBox(height: 10),
                Column(children: paymentDetail),
              ],
            ),
          ),
        ],
      ),
    );
  }

  /// ส่วนแสดงผลต่างจากบิลเดิม (สำหรับ return)
  void _buildRefTotalSection(List<Widget> sumDetails) {
    if (state.widget.type == global.TransactionTypeEnum.salereturn ||
        state.widget.type == global.TransactionTypeEnum.purchasereturn) {
      sumDetails.add(
        Container(
          margin: EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: Row(
                  children: [
                    Expanded(
                      child: TotalTextController(
                        readOnly: !state.screenData.ismanualamount,
                        title: global.language("text_form_ref_total_original"),
                        data: state.screenData.reftotaloriginal,
                        icon: null,
                        useColor: false,
                        onChanged: (value) {
                          if (value != '') {
                            state.screenData.totalamount = double.parse(
                              value.replaceAll(',', ''),
                            );
                          }
                        },
                      ),
                    ),
                    SizedBox(width: 10),
                    Expanded(
                      child: TotalTextController(
                        readOnly: !state.screenData.ismanualamount,
                        title: global.language("text_form_ref_total_correct"),
                        data: state.screenData.reftotalcorrect,
                        icon: null,
                        useColor: false,
                        onChanged: (value) {
                          if (value != '') {
                            state.screenData.totalamount = double.parse(
                              value.replaceAll(',', ''),
                            );
                          }
                        },
                      ),
                    ),
                  ],
                ),
              ),
              SizedBox(width: 10),
              Expanded(
                child: TotalTextController(
                  readOnly: true,
                  title: global.language("text_form_ref_total_diff"),
                  data: state.screenData.reftotaldiff,
                  icon: null,
                  useColor: false,
                  onChanged: (value) {},
                ),
              ),
            ],
          ),
        ),
      );
    }
  }

  /// ส่วนแสดงรวมมูลค่า
  void _buildTotalValueSection(List<Widget> sumDetails) {
    if (state.widget.type != global.TransactionTypeEnum.advancePaymentRefund &&
        state.widget.type != global.TransactionTypeEnum.depositRefund &&
        state.widget.type != global.TransactionTypeEnum.paidAdvanceRefund) {
      sumDetails.add(
        Container(
          margin: EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: TotalTextController(
                  readOnly: !state.screenData.ismanualamount,
                  title: global.language("doc_total_value"),
                  data: state.screenData.totalvalue,
                  icon: null,
                  onChanged: (value) {
                    if (value != '' && state.screenData.ismanualamount) {
                      state.screenData.totalvalue = double.parse(
                        value.replaceAll(',', ''),
                      );
                    }
                  },
                ),
              ),
              _buildManualAmountCheckbox(),
            ],
          ),
        ),
      );
    }
  }

  /// Checkbox สำหรับเลือก manual amount
  Widget _buildManualAmountCheckbox() {
    if (state.widget.type == global.TransactionTypeEnum.purchase ||
        state.widget.type == global.TransactionTypeEnum.purchaseorder ||
        state.widget.type == global.TransactionTypeEnum.purchasereturn ||
        state.widget.type == global.TransactionTypeEnum.purchasepartial ||
        state.widget.type == global.TransactionTypeEnum.accrualreceive) {
      return Expanded(
        child: Row(
          children: [
            const SizedBox(width: 10),
            Checkbox(
              fillColor: WidgetStateProperty.resolveWith((states) {
                if (states.contains(WidgetState.selected)) {
                  return global.theme.negativeHighlightTextColor;
                }
                return null;
              }),
              value: state.screenData.ismanualamount,
              onChanged: (bool? value) {
                state.screenData.ismanualamount = value!;
                if (!value) {
                  state.transactionCalculator.calTotalValue();
                  state.transactionCalculator.calPayTotal();
                }
                state.setState(() {});
              },
            ),
            Expanded(
              child: Text(
                global.language("is_manual_amount"),
                style: TextStyle(
                  color: global.theme.textColor,
                  fontSize: 16,
                  overflow: TextOverflow.ellipsis,
                ),
                maxLines: 2,
              ),
            ),
          ],
        ),
      );
    }
    return Container();
  }

  /// ส่วนแสดงส่วนลด
  void _buildDiscountSection(List<Widget> sumDetails) {
    if (state.widget.type == global.TransactionTypeEnum.sale ||
        state.widget.type == global.TransactionTypeEnum.saleorder ||
        state.widget.type == global.TransactionTypeEnum.salereturn ||
        state.widget.type == global.TransactionTypeEnum.purchase ||
        state.widget.type == global.TransactionTypeEnum.purchaseorder ||
        state.widget.type == global.TransactionTypeEnum.accrualreceive) {
      // ส่วนลดก่อนชำระเงิน, ส่วนลดจากแต้ม, ส่วนลดจากคูปอง
      sumDetails.add(_buildDiscountFormulaRow());
      // รวมส่วนลดก่อนชำระเงิน
      sumDetails.add(_buildTotalDiscountRow());
      // ส่วนลดสินค้ามีภาษี, ส่วนลดสินค้ายกเว้นภาษี
      sumDetails.add(_buildVatDiscountRow());
    }
  }

  /// Row สำหรับส่วนลดก่อนชำระเงิน
  Widget _buildDiscountFormulaRow() {
    return Container(
      margin: EdgeInsets.only(bottom: 10),
      child: Row(
        children: [
          // ส่วนลดก่อนชำระเงิน
          Expanded(
            child: TextFormField(
              readOnly: false,
              enabled: true,
              decoration: InputDecoration(
                border: OutlineInputBorder(),
                labelText: global.language("detail_discount_formula"),
              ),
              controller: state.detaildiscountformulaController,
              keyboardType: const TextInputType.numberWithOptions(
                decimal: true,
              ),
              inputFormatters: [
                FilteringTextInputFormatter.allow(RegExp(r'[0-9%.]')),
              ],
              onChanged: (value) {
                if (value != '' && value.isNotEmpty) {
                  state
                      .detaildiscountformulaController
                      .value = TextEditingValue(
                    text: value,
                    selection: state.detaildiscountformulaController.selection,
                  );
                  state.screenData.detaildiscountformula = value;

                  RegExp regExp = RegExp(r'[0-9%.]');
                  if (regExp.hasMatch(value)) {
                    state.transactionCalculator.calTotalValue();
                  }
                } else {
                  state.transactionCalculator.calTotalValue();
                }
              },
              onEditingComplete: () {
                if (state.detaildiscountformulaController.text == '') {
                  state.detaildiscountformulaController.text = '0';
                  state.screenData.detaildiscountformula = '0';
                  state.transactionCalculator.calTotalValue();
                }
              },
            ),
          ),
          const SizedBox(width: 10),
          // ส่วนลดจากแต้ม
          Expanded(
            child: TotalTextController(
              readOnly: true,
              title: global.language('discount_from_points'),
              data: state.screenData.pointdiscountamount,
              icon: null,
              onChanged: (value) {
                if (value != '') {
                  state.screenData.pointdiscountamount = int.parse(
                    value.replaceAll(',', ''),
                  );
                }
              },
            ),
          ),
          const SizedBox(width: 10),
          // ส่วนลดจากคูปอง
          Expanded(
            child: TotalTextController(
              readOnly: true,
              title: global.language('discount_from_coupon'),
              data: state.screenData.coupondiscountamount,
              icon: null,
              onChanged: (value) {
                if (value != '') {
                  state.screenData.coupondiscountamount = double.parse(
                    value.replaceAll(',', ''),
                  );
                }
              },
            ),
          ),
        ],
      ),
    );
  }

  /// Row สำหรับรวมส่วนลดก่อนชำระเงิน
  Widget _buildTotalDiscountRow() {
    return Container(
      margin: EdgeInsets.only(bottom: 10),
      child: Row(
        children: [
          Expanded(
            child: TotalTextController(
              readOnly: !state.screenData.ismanualamount,
              title: global.language("total_detail_total_discount"),
              data: state.screenData.detailtotaldiscount,
              icon: null,
              onChanged: (value) {
                if (value != '') {
                  state.screenData.detailtotaldiscount = double.parse(
                    value.replaceAll(',', ''),
                  );
                }
              },
            ),
          ),
        ],
      ),
    );
  }

  /// Row สำหรับส่วนลดสินค้ามีภาษี/ยกเว้นภาษี
  Widget _buildVatDiscountRow() {
    return Container(
      margin: EdgeInsets.only(bottom: 10),
      child: Row(
        children: [
          Expanded(
            child: TotalTextController(
              readOnly: true,
              title: global.language("total_discount_vat_amount"),
              data: state.screenData.totaldiscountvatamount,
              icon: null,
              onChanged: (value) {
                if (value != '') {
                  state.screenData.totaldiscountvatamount = double.parse(
                    value.replaceAll(',', ''),
                  );
                }
              },
            ),
          ),
          SizedBox(width: 10),
          Expanded(
            child: TotalTextController(
              readOnly: true,
              title: global.language("total_discount_except_vat_amount"),
              data: state.screenData.totaldiscountexceptvatamount,
              icon: null,
              onChanged: (value) {
                if (value != '') {
                  state.screenData.totaldiscountexceptvatamount = double.parse(
                    value.replaceAll(',', ''),
                  );
                }
              },
            ),
          ),
        ],
      ),
    );
  }

  /// ส่วนแสดงเงินมัดจำ/เงินล่วงหน้า
  void _buildDepositSection(List<Widget> sumDetails) {
    if (state.widget.type == global.TransactionTypeEnum.purchase ||
        state.widget.type == global.TransactionTypeEnum.sale ||
        state.widget.type == global.TransactionTypeEnum.accrualreceive) {
      sumDetails.add(
        Container(
          margin: const EdgeInsets.only(bottom: 20),
          child: DepositPaymentWidget(
            type: state.widget.type,
            custcode: state.screenData.custcode,
            depositDocs: state.screenData.depositdocs ?? [],
            onDepositDocsChanged: (List<DepositDocModel> updatedDocs) {
              state.setState(() {
                state.screenData.depositdocs = updatedDocs;
              });
            },
            onSumDepositChanged: (double sumDeposit) {
              state.setState(() {
                state.screenData.sumdeposit = sumDeposit;
                state.transactionCalculator.calTotalValue();
              });
            },
          ),
        ),
      );
    }
  }

  /// ส่วนแสดงยอดก่อนภาษี, ภาษี, ยอดหลังภาษี
  void _buildVatSection(List<Widget> sumDetails) {
    if (state.widget.type == global.TransactionTypeEnum.stockreceiveproduct) {
      return;
    }

    bool isExcludedType =
        state.widget.type == global.TransactionTypeEnum.advancePayment ||
        state.widget.type == global.TransactionTypeEnum.advancePaymentRefund ||
        state.widget.type == global.TransactionTypeEnum.deposit ||
        state.widget.type == global.TransactionTypeEnum.depositRefund ||
        state.widget.type == global.TransactionTypeEnum.paidAdvance ||
        state.widget.type == global.TransactionTypeEnum.paidAdvanceRefund;

    if (!isExcludedType) {
      // ยอดก่อนภาษี, ภาษีมูลค่าเพิ่ม
      sumDetails.add(_buildBeforeVatRow());
      // ยอดรวมสินค้ามีภาษี, ยอดรวมสินค้ายกเว้นภาษี
      sumDetails.add(_buildAfterVatRow());
      // ยอดรวมก่อนหักส่วนลดท้ายบิล
      sumDetails.add(_buildDetailTotalRow());
    }
  }

  Widget _buildBeforeVatRow() {
    return Container(
      margin: EdgeInsets.only(bottom: 10),
      child: Row(
        children: [
          Expanded(
            child: TotalTextController(
              readOnly: !state.screenData.ismanualamount,
              title: global.language("doc_before_vat_amount"),
              data: state.screenData.totalbeforevat,
              icon: null,
              onChanged: (value) {
                if (value != '' && state.screenData.ismanualamount) {
                  state.screenData.totalbeforevat = double.parse(
                    value.replaceAll(',', ''),
                  );
                }
              },
            ),
          ),
          SizedBox(width: 10),
          Expanded(
            child: TotalTextController(
              readOnly: !state.screenData.ismanualamount,
              title: global.language("doc_vat_amount"),
              data: state.screenData.totalvatvalue,
              icon: null,
              onChanged: (value) {
                if (value != '' && state.screenData.ismanualamount) {
                  state.screenData.totalvatvalue = double.parse(
                    value.replaceAll(',', ''),
                  );
                }
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildAfterVatRow() {
    return Container(
      margin: EdgeInsets.only(bottom: 10),
      child: Row(
        children: [
          Expanded(
            child: TotalTextController(
              readOnly: !state.screenData.ismanualamount,
              title: global.language("doc_after_vat_amount"),
              data: state.screenData.totalaftervat,
              icon: null,
              onChanged: (value) {
                if (value != '' && state.screenData.ismanualamount) {
                  state.screenData.totalaftervat = double.parse(
                    value.replaceAll(',', ''),
                  );
                }
              },
            ),
          ),
          SizedBox(width: 10),
          Expanded(
            child: TotalTextController(
              readOnly: !state.screenData.ismanualamount,
              title: global.language("doc_except_vat_amount"),
              data: state.screenData.totalexceptvat,
              icon: null,
              onChanged: (value) {
                if (value != '' && state.screenData.ismanualamount) {
                  state.screenData.totalexceptvat = double.parse(
                    value.replaceAll(',', ''),
                  );
                }
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDetailTotalRow() {
    return Container(
      margin: EdgeInsets.only(bottom: 10),
      child: Row(
        children: [
          Expanded(
            child: TotalTextController(
              readOnly: !state.screenData.ismanualamount,
              title: global.language("detail_total_amount"),
              data: state.screenData.detailtotalamount,
              icon: null,
              onChanged: (value) {
                if (value != '' && state.screenData.ismanualamount) {
                  state.screenData.detailtotalamount = double.parse(
                    value.replaceAll(',', ''),
                  );
                }
              },
            ),
          ),
        ],
      ),
    );
  }

  /// ส่วนแสดงยอดรวมสุทธิและปัดเศษ
  void _buildFinalTotalSection(List<Widget> sumDetails) {
    if (state.widget.type == global.TransactionTypeEnum.stockreceiveproduct) {
      return;
    }

    // ยอดรวมทั้งสิ้น, ปัดเศษ
    sumDetails.add(_buildTotalAndRoundRow());
    // ยอดรวมสุทธิ, รวมเงินทั้งสิ้น
    sumDetails.add(_buildNetTotalRow());
  }

  Widget _buildTotalAndRoundRow() {
    return Container(
      margin: EdgeInsets.only(bottom: 10),
      child: Row(
        children: [
          Expanded(
            child: TotalTextController(
              readOnly: !state.screenData.ismanualamount,
              title: global.language("doc_total_amount"),
              data: state.screenData.totalamountafterdiscount,
              icon: null,
              onChanged: (value) {
                if (value != '') {
                  state.screenData.totalamountafterdiscount = double.parse(
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
              enabled: true,
              decoration: InputDecoration(
                border: OutlineInputBorder(),
                labelText: global.language("round_amount"),
              ),
              controller: state.roundAmountController,
              keyboardType: const TextInputType.numberWithOptions(
                decimal: true,
              ),
              inputFormatters: [global.NumberInputFormatter()],
              onChanged: (value) {
                if (value != '' && value.isNotEmpty) {
                  state.roundAmountController.value = TextEditingValue(
                    text: value,
                    selection: state.roundAmountController.selection,
                  );

                  RegExp regExp = RegExp(r'[0-9]');
                  if (regExp.hasMatch(value)) {
                    if (!state.screenData.ismanualamount) {
                      state.transactionCalculator.calPayTotal();
                    }
                  }
                }
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildNetTotalRow() {
    return Row(
      children: [
        Expanded(
          child: GestureDetector(
            onTap: () async {
              if (!state.screenData.ismanualamount) return;

              final result = await showNumericInputDialog(
                state.context,
                title: global.language("sum_pay"),
                initialValue: state.screenData.totalamount.toString(),
                decimalPlaces: 2,
              );

              if (result != null && result.isNotEmpty) {
                state.setState(() {
                  state.screenData.totalamount = double.parse(
                    result.replaceAll(',', ''),
                  );
                  state.transactionCalculator.calPayTotal();
                });
              }
            },
            child: TotalTextController(
              readOnly: !state.screenData.ismanualamount,
              title: global.language("sum_pay"),
              data:
                  state.screenData.totalamount +
                  (state.screenData.roundamount ?? 0),
              icon: null,
              useColor: true,
              onChanged: (value) {
                if (value != '') {
                  state.screenData.totalamount = double.parse(
                    value.replaceAll(',', ''),
                  );
                }
              },
            ),
          ),
        ),
        SizedBox(width: 10),
        Expanded(
          child: GestureDetector(
            onTap: () async {
              final result = await showNumericInputDialog(
                state.context,
                title: global.language("total_amount"),
                initialValue: state.payTotalBill.toString(),
                decimalPlaces: 2,
              );

              if (result != null && result.isNotEmpty) {
                state.setState(() {
                  state.payTotalBill = double.parse(result.replaceAll(',', ''));
                  state.transactionCalculator.calPayTotal();
                });
              }
            },
            child: TotalTextController(
              readOnly: true,
              title: global.language("total_amount"),
              data: state.payTotalBill,
              icon: null,
              useColor: true,
              onChanged: (value) {},
            ),
          ),
        ),
      ],
    );
  }

  /// ส่วนแสดงการชำระเงิน
  void _buildPaymentSection(List<Widget> paymentDetail) {
    if (state.widget.type == global.TransactionTypeEnum.stockreceiveproduct) {
      return;
    }

    // สร้าง lists สำหรับ payment methods
    List<Widget> listTransfer = _buildTransferList();
    List<Widget> listCredit = _buildCreditCardList();
    List<Widget> listCheque = _buildChequeList();
    List<Widget> listCoupon = _buildCouponList();
    List<Widget> listQr = _buildQrList();

    bool shouldShowPayment = _shouldShowPaymentSection();

    if (shouldShowPayment) {
      paymentDetail.add(
        Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  state.paymentWidgets.buildPayMenuWidget(),
                  _buildPaymentDetailWidget(
                    listTransfer,
                    listCredit,
                    listCheque,
                    listCoupon,
                    listQr,
                  ),
                ],
              ),
            ),
            const SizedBox(width: 10),
            Expanded(
              child: PaymentSummaryWidget(
                screenData: state.screenData,
                type: state.widget.type,
                payCashAmountController: state.payCashAmountController,
                payDeliveryCashAmountController:
                    state.payDeliveryCashAmountController,
                onPaymentTypeSelected: (index) {
                  state.showPayDetail = index;
                  state.setState(() {});
                },
              ),
            ),
          ],
        ),
      );
    }
  }

  bool _shouldShowPaymentSection() {
    return state.widget.type == global.TransactionTypeEnum.purchase ||
        state.widget.type == global.TransactionTypeEnum.purchaseorder ||
        state.widget.type == global.TransactionTypeEnum.purchasepartial ||
        state.widget.type == global.TransactionTypeEnum.purchasereturn ||
        state.widget.type == global.TransactionTypeEnum.accrualreceive ||
        state.widget.type == global.TransactionTypeEnum.sale ||
        state.widget.type == global.TransactionTypeEnum.saleorder ||
        state.widget.type == global.TransactionTypeEnum.salereturn ||
        state.widget.type == global.TransactionTypeEnum.advancePayment ||
        state.widget.type == global.TransactionTypeEnum.advancePaymentRefund ||
        state.widget.type == global.TransactionTypeEnum.deposit ||
        state.widget.type == global.TransactionTypeEnum.depositRefund ||
        state.widget.type == global.TransactionTypeEnum.paidAdvance ||
        state.widget.type == global.TransactionTypeEnum.paidAdvanceRefund ||
        state.widget.type == global.TransactionTypeEnum.receiveDeposit ||
        state.widget.type == global.TransactionTypeEnum.receiveDepositRefund;
  }

  Widget _buildPaymentDetailWidget(
    List<Widget> listTransfer,
    List<Widget> listCredit,
    List<Widget> listCheque,
    List<Widget> listCoupon,
    List<Widget> listQr,
  ) {
    switch (state.showPayDetail) {
      case 0:
        return state.paymentWidgets.buildPayCashWidget();
      case 1:
        return state.paymentWidgets.buildPayTransferWidget(listTransfer);
      case 2:
        return state.paymentWidgets.buildPayCreditCardWidget(listCredit);
      case 3:
        return state.paymentWidgets.buildPayChequeWidget(listCheque);
      case 4:
        return state.paymentWidgets.buildPayCouponWidget(listCoupon);
      case 5:
        return state.paymentWidgets.buildPayQrWidget(listQr);
      case 6:
        return state.paymentWidgets.buildPayDeliveryWidget();
      case 7:
        return state.paymentWidgets.buildPayPointWidget();
      case 8:
        return PayAdvancePaymentWidget(
          advancePaymentDocs: state.screenData.advancepaymentdocs ?? [],
          onAdvancePaymentDocsChanged: (docs) {
            state.setState(() {
              state.screenData.advancepaymentdocs = docs;
            });
            state.transactionCalculator.calPayTotal();
          },
          onSumAdvancePaymentChanged: (sum) {
            state.setState(() {
              state.screenData.sumadvancepayment = sum;
            });
          },
          custcode: state.screenData.custcode,
          inquirytype: state.screenData.inquirytype,
          type: state.widget.type,
        );
      default:
        return Container();
    }
  }

  // === Payment Method Lists ===

  List<Widget> _buildTransferList() {
    List<Widget> list = [];
    for (var i = 0; i < state.payTransfer.length; i++) {
      list.add(
        PaymentMethodItemWidget(
          paymentType: PaymentMethodType.transfer,
          itemIndex: i,
          paymentData: state.payTransfer[i],
          amountController: state.payTransferAmountController[i],
          onDelete: () => _onDeleteTransfer(i),
          onAmountChanged: (amount) => _onTransferAmountChanged(i, amount),
          onBankSearch: () => _onTransferBankSearch(i),
          onDateSelected: (date) => _onTransferDateSelected(i, date),
        ),
      );
    }
    return list;
  }

  void _onDeleteTransfer(int i) {
    state.setState(() {
      state.payTransfer.removeAt(i);
      state.transferDateController.removeAt(i);
      state.payTransferAmountController.removeAt(i);
      state.transactionCalculator.calPayTotal();
    });
  }

  void _onTransferAmountChanged(int i, double amount) {
    state.setState(() {
      state.payTransfer[i].amount = amount;
      state.transactionCalculator.calPayTotal();
    });
  }

  void _onTransferBankSearch(int i) {
    state.bookBankSearch().then((value) {
      if (value != null) {
        state.setState(() {
          state.payTransfer[i].book_bank_code = value.passbook;
          state.payTransfer[i].bank_code = value.bankcode;
          state.payTransfer[i].bank_name = value.banknames![0].name;
        });
      }
    });
  }

  void _onTransferDateSelected(int i, DateTime? date) {
    if (date != null) {
      state.setState(() {
        final currentTime = state.payTransfer[i].doc_date_time!;
        final combinedDateTime = DateTime(
          date.year,
          date.month,
          date.day,
          currentTime.hour,
          currentTime.minute,
          currentTime.second,
          currentTime.millisecond,
        );
        state.docDateTimeValidated = true;
        state.payTransfer[i].doc_date_time = combinedDateTime.toLocal();
      });
    }
  }

  List<Widget> _buildCreditCardList() {
    List<Widget> list = [];
    for (var i = 0; i < state.payCreditCard.length; i++) {
      list.add(
        PaymentMethodItemWidget(
          paymentType: PaymentMethodType.creditCard,
          itemIndex: i,
          paymentData: state.payCreditCard[i],
          amountController: state.payCreditCardAmountController[i],
          onDelete: () {
            state.setState(() {
              state.payCreditCard.removeAt(i);
              state.creditCardDateController.removeAt(i);
              state.payCreditCardAmountController.removeAt(i);
              state.transactionCalculator.calPayTotal();
            });
          },
          onAmountChanged: (amount) {
            state.setState(() {
              state.payCreditCard[i].amount = amount;
              state.transactionCalculator.calPayTotal();
            });
          },
          onBankSearch: () {
            state.bookBankSearch().then((value) {
              if (value != null) {
                state.setState(() {
                  state.payCreditCard[i].book_bank_code = value.passbook;
                  state.payCreditCard[i].bank_code = value.bankcode;
                  state.payCreditCard[i].bank_name = value.banknames![0].name;
                });
              }
            });
          },
          onDateSelected: (date) {
            if (date != null) {
              state.setState(() {
                final currentTime = state.payCreditCard[i].doc_date_time!;
                final combinedDateTime = DateTime(
                  date.year,
                  date.month,
                  date.day,
                  currentTime.hour,
                  currentTime.minute,
                  currentTime.second,
                  currentTime.millisecond,
                );
                state.docDateTimeValidated = true;
                state.payCreditCard[i].doc_date_time = combinedDateTime
                    .toLocal();
              });
            }
          },
        ),
      );
    }
    return list;
  }

  List<Widget> _buildChequeList() {
    List<Widget> list = [];
    for (var i = 0; i < state.payCheque.length; i++) {
      list.add(
        PaymentMethodItemWidget(
          paymentType: PaymentMethodType.cheque,
          itemIndex: i,
          paymentData: state.payCheque[i],
          amountController: state.payChequeAmountController[i],
          onDelete: () {
            state.setState(() {
              state.payCheque.removeAt(i);
              state.chequeDateController.removeAt(i);
              state.chequeDueDateDateController.removeAt(i);
              state.payChequeAmountController.removeAt(i);
              state.transactionCalculator.calPayTotal();
            });
          },
          onAmountChanged: (amount) {
            state.setState(() {
              state.payCheque[i].amount = amount;
              state.transactionCalculator.calPayTotal();
            });
          },
          onBankSearch: () {
            state.bookBankSearch().then((value) {
              if (value != null) {
                state.setState(() {
                  state.payCheque[i].book_bank_code = value.passbook;
                  state.payCheque[i].bank_code = value.bankcode;
                  state.payCheque[i].bank_name = value.banknames![0].name;
                });
              }
            });
          },
          onDateSelected: (date) {
            if (date != null) {
              state.setState(() {
                final currentTime = state.payCheque[i].doc_date_time!;
                final combinedDateTime = DateTime(
                  date.year,
                  date.month,
                  date.day,
                  currentTime.hour,
                  currentTime.minute,
                  currentTime.second,
                  currentTime.millisecond,
                );
                state.docDateTimeValidated = true;
                state.payCheque[i].doc_date_time = combinedDateTime.toLocal();
              });
            }
          },
          onDueDateSelected: (date) {
            if (date != null) {
              state.setState(() {
                final currentTime = state.payCheque[i].due_date!;
                final combinedDateTime = DateTime(
                  date.year,
                  date.month,
                  date.day,
                  currentTime.hour,
                  currentTime.minute,
                  currentTime.second,
                  currentTime.millisecond,
                );
                state.docDateTimeValidated = true;
                state.payCheque[i].due_date = combinedDateTime.toLocal();
              });
            }
          },
        ),
      );
    }
    return list;
  }

  List<Widget> _buildCouponList() {
    List<Widget> list = [];

    // Coupon widget (new)
    if (state.couPons!.isNotEmpty) {
      list.add(
        CouponWidget(
          couPons: state.couPons!,
          isReadOnly: true,
          onCouponDeleted: (index) {
            state.couPons!.removeAt(index);
            state.transactionCalculator.calPayTotal();
          },
          onAmountChanged: () {
            state.transactionCalculator.calPayTotal();
          },
        ),
      );
    }

    // Coupon payment items
    for (var i = 0; i < state.payCoupon.length; i++) {
      list.add(
        PaymentMethodItemWidget(
          paymentType: PaymentMethodType.coupon,
          itemIndex: i,
          paymentData: state.payCoupon[i],
          amountController: state.payCouponAmountController[i],
          onDelete: () {
            state.setState(() {
              state.payCoupon.removeAt(i);
              state.couponDateController.removeAt(i);
              state.payCouponAmountController.removeAt(i);
              state.transactionCalculator.calPayTotal();
            });
          },
          onAmountChanged: (amount) {
            state.setState(() {
              state.payCoupon[i].amount = amount;
              state.transactionCalculator.calPayTotal();
            });
          },
          onDateSelected: (date) {
            if (date != null) {
              state.setState(() {
                final currentTime = state.payCoupon[i].doc_date_time!;
                final combinedDateTime = DateTime(
                  date.year,
                  date.month,
                  date.day,
                  currentTime.hour,
                  currentTime.minute,
                  currentTime.second,
                  currentTime.millisecond,
                );
                state.docDateTimeValidated = true;
                state.payCoupon[i].doc_date_time = combinedDateTime.toLocal();
              });
            }
          },
        ),
      );
    }
    return list;
  }

  List<Widget> _buildQrList() {
    List<Widget> list = [];
    for (var i = 0; i < state.payQr.length; i++) {
      list.add(
        PaymentMethodItemWidget(
          paymentType: PaymentMethodType.qrCode,
          itemIndex: i,
          paymentData: state.payQr[i],
          amountController: state.payQrAmountController[i],
          onDelete: () {
            state.setState(() {
              state.payQr.removeAt(i);
              state.qrDateController.removeAt(i);
              state.payQrAmountController.removeAt(i);
              state.transactionCalculator.calPayTotal();
            });
          },
          onAmountChanged: (amount) {
            state.setState(() {
              state.payQr[i].amount = amount;
              state.transactionCalculator.calPayTotal();
            });
          },
          onDateSelected: (date) {
            if (date != null) {
              state.setState(() {
                final currentTime = state.payQr[i].doc_date_time!;
                final combinedDateTime = DateTime(
                  date.year,
                  date.month,
                  date.day,
                  currentTime.hour,
                  currentTime.minute,
                  currentTime.second,
                  currentTime.millisecond,
                );
                state.docDateTimeValidated = true;
                state.payQr[i].doc_date_time = combinedDateTime.toLocal();
              });
            }
          },
        ),
      );
    }
    return list;
  }
}
