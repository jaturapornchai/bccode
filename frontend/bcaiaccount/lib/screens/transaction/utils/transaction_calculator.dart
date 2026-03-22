import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/calamount.dart' as calAmount;
import 'package:smlaicloud/screens/transaction/utils/transaction_currency_utils.dart';
import 'package:smlaicloud/model/transaction_model.dart';

/// Helper class สำหรับ Calculator functions
/// ใช้สำหรับคำนวณยอดรวม, ยอดชำระ, และตรวจสอบการชำระเงิน
///
/// ใช้ร่วมกับ TransactionCurrencyUtils สำหรับ Multi-Currency calculation
///
/// รองรับทั้ง TransactionEditScreenState และ PurchaseOrderEditScreenState
/// โดยใช้ dynamic type
class TransactionCalculator {
  final dynamic state;

  TransactionCalculator(this.state);

  /// คำนวณยอดรวมทั้งหมดของเอกสาร
  void calTotalValue() {
    double totalValue = 0;
    double totalQty = 0;
    String detaildiscountformulaData = "";
    String discountWordData = "";

    /// รวมสินค้ามีภาษี
    double totalAmountVatCale0 = 0;

    /// รวมสินค้ายกเว้นภาษี
    double totalAmountVatCale1 = 0;

    if (state.screenData.details!.isEmpty) {
      state.screenData.totalvalue = 0.0;
      state.screenData.totalqty = 0.0;
      state.screenData.detailtotaldiscount = 0.0;
      state.screenData.totaldiscountvatamount = 0.0;
      state.screenData.totaldiscountexceptvatamount = 0.0;
      state.screenData.totalbeforevat = 0.0;
      state.screenData.totalvatvalue = 0.0;
      state.screenData.totalaftervat = 0.0;
      state.screenData.totalexceptvat = 0.0;
      state.screenData.detailtotalamount = 0.0;
      state.screenData.totalamountafterdiscount = 0.0;
      state.screenData.totalamount = 0.0;
      state.screenData.totalamountDoc = 0.0; // ยอดรวมในสกุลเงินเอกสาร
      state.screenData.totalvalueDoc = 0.0;
      state.screenData.totaldiscountDoc = 0.0;
      state.screenData.totalvatvalueDoc = 0.0;
      state.screenData.totalbeforevatDoc = 0.0;
      state.screenData.totalaftervatDoc = 0.0;
      state.screenData.discountword = "0";
      state.screenData.totaldiscount = 0.0;

      state.discountWordController.text = "0";
      state.detaildiscountformulaController.text = "0";

      state.setState(() {});
      return;
    }

    /// คำนวนพี่ฟิช

    state.screenData = calAmount.calTotalValue(state.screenData);

    for (var val in state.screenData.details!) {
      if (val.vatcal == 0) {
        totalAmountVatCale0 += val.sumamount;
      } else if (val.vatcal == 1) {
        totalAmountVatCale1 += val.sumamount;
      }

      totalValue += val.sumamount;
      totalQty += val.qty;
    }

    if (state.detaildiscountformulaController.text.isNotEmpty) {
      detaildiscountformulaData = state.detaildiscountformulaController.text;
    } else {
      detaildiscountformulaData = "0";
    }

    if (state.discountWordController.text.isNotEmpty) {
      discountWordData = state.discountWordController.text;
    } else {
      discountWordData = "0";
    }

    /// ส่วนลดก่อนชำระเงิน
    if (detaildiscountformulaData.contains('%')) {
      String detaildiscountformula = detaildiscountformulaData.replaceAll(
        '%',
        '',
      );
      state.screenData.detailtotaldiscount = double.parse(
        ((totalValue * double.parse(detaildiscountformula)) / 100)
            .toStringAsFixed(global.getDecimalDocument()),
      );
    } else if (detaildiscountformulaData
        .replaceAll(',', '')
        .trim()
        .isNotEmpty) {
      state.screenData.detailtotaldiscount = double.parse(
        detaildiscountformulaData.replaceAll(',', '').trim(),
      );
    }

    state.screenData.totalqty = totalQty;

    /// ส่วนลดสินค้ามีภาษี
    state.screenData.totaldiscountvatamount =
        ((state.screenData.detailtotaldiscount! * totalAmountVatCale0) /
            state.screenData.totalvalue);

    state.screenData.totaldiscountvatamount = double.parse(
      state.screenData.totaldiscountvatamount!.toStringAsFixed(global.getDecimalDocument()),
    );

    if (state.screenData.totaldiscountvatamount!.isNaN) {
      state.screenData.totaldiscountvatamount = 0.0;
    }

    /// ส่วนลดสินค้ายกเว้นภาษี
    state.screenData.totaldiscountexceptvatamount =
        state.screenData.detailtotaldiscount! -
            state.screenData.totaldiscountvatamount!;

    state.screenData.totaldiscountexceptvatamount = double.parse(
      state.screenData.totaldiscountexceptvatamount!.toStringAsFixed(global.getDecimalDocument()),
    );

    if (state.screenData.totaldiscountexceptvatamount!.isNaN) {
      state.screenData.totaldiscountexceptvatamount = 0.0;
    }

    ///  0 = แยกนอก , 1 = รวมใน
    // PO ไม่มีเงินมัดจำ — บังคับ deposit = 0
    double depositAmount = 0;
    if (state.transactionType != global.TransactionTypeEnum.purchaseorder) {
      depositAmount = state.screenData.sumdeposit ?? 0; // เงินมัดจำ
    }
    double totalAmountVatAfterDiscount = totalAmountVatCale0 -
        state.screenData
            .totaldiscountvatamount!; // รวมมูลค่าภาษี - ส่วนลดภาษี

    if (state.screenData.vattype == 0) {
      // แยกนอก
      if (depositAmount >= totalAmountVatAfterDiscount) {
        // ถ้าเงินมัดจำ มากกว่าหรือเท่ากับ รวมมูลค่าภาษี
        /// มูลค่าก่อนภาษี = 0
        state.screenData.totalbeforevat = 0.0;

        /// มูลค่าภาษี = 0
        state.screenData.totalvatvalue = 0.0;

        /// มูลค่ายกเว้นภาษี = ( รวมมูลค่ายกเว้นภาษี - ส่วนลดยกเว้นภาษี - (เงินมัดจำ - (รวมมูลค่าภาษี - ส่วนลดภาษี) ) )
        double excessDeposit = depositAmount - totalAmountVatAfterDiscount;
        state.screenData.totalexceptvat = totalAmountVatCale1 -
            state.screenData.totaldiscountexceptvatamount! -
            excessDeposit;
      } else {
        // ถ้าเงินมัดจำ น้อยกว่า รวมมูลค่าภาษี
        /// มูลค่าก่อนภาษี = (รวมมูลค่าภาษี - ส่วนลดภาษี - เงินมัดจำ)
        state.screenData.totalbeforevat = double.parse(
          (totalAmountVatAfterDiscount - depositAmount).toStringAsFixed(global.getDecimalDocument()),
        );

        /// มูลค่าภาษี = ( มูลค่าก่อนภาษี * (7/100) )
        state.screenData.totalvatvalue = double.parse(
          (state.screenData.totalbeforevat *
                  (state.screenData.vatrate / 100))
              .toStringAsFixed(global.getDecimalDocument()),
        );

        /// มูลค่ายกเว้นภาษี = เหมือนเดิม
        state.screenData.totalexceptvat =
            totalAmountVatCale1 - state.screenData.totaldiscountexceptvatamount!;
      }
    } else if (state.screenData.vattype == 1) {
      // รวมใน
      if (depositAmount >= totalAmountVatAfterDiscount) {
        // ถ้าเงินมัดจำ มากกว่าหรือเท่ากับ รวมมูลค่าภาษี
        /// มูลค่าภาษี = 0
        state.screenData.totalvatvalue = 0.0;

        /// มูลค่าก่อนภาษี = 0
        state.screenData.totalbeforevat = 0.0;

        /// มูลค่ายกเว้นภาษี = ( รวมมูลค่ายกเว้นภาษี - ส่วนลดยกเว้นภาษี - (เงินมัดจำ - (รวมมูลค่าภาษี - ส่วนลดภาษี)) )
        double excessDeposit = depositAmount - totalAmountVatAfterDiscount;
        state.screenData.totalexceptvat = totalAmountVatCale1 -
            state.screenData.totaldiscountexceptvatamount! -
            excessDeposit;
      } else {
        // ถ้าเงินมัดจำ น้อยกว่า รวมมูลค่าภาษี
        double remainingAmount = totalAmountVatAfterDiscount - depositAmount;

        /// มูลค่าภาษี = ( (รวมมูลค่าภาษี - ส่วนลดภาษี - เงินมัดจำ) * (7/107) )
        state.screenData.totalvatvalue = double.parse(
          ((remainingAmount * state.screenData.vatrate) /
                  (100 + state.screenData.vatrate))
              .toStringAsFixed(global.getDecimalDocument()),
        );

        /// มูลค่าก่อนภาษี = (รวมมูลค่าภาษี - ส่วนลดภาษี - เงินมัดจำ) - มูลค่าภาษี
        state.screenData.totalbeforevat = double.parse(
          (remainingAmount - state.screenData.totalvatvalue).toStringAsFixed(global.getDecimalDocument()),
        );

        /// มูลค่ายกเว้นภาษี = เหมือนเดิม
        state.screenData.totalexceptvat =
            totalAmountVatCale1 - state.screenData.totaldiscountexceptvatamount!;
      }
    } else {
      // Else (ไม่กระทบภาษี)
      if (depositAmount >= totalAmountVatAfterDiscount) {
        // ถ้าเงินมัดจำ มากกว่าหรือเท่ากับ รวมมูลค่าภาษี
        /// มูลค่าก่อนภาษี = 0
        state.screenData.totalbeforevat = 0.0;

        /// มูลค่าภาษี = 0
        state.screenData.totalvatvalue = 0.0;

        /// มูลค่ายกเว้นภาษี = ( รวมมูลค่ายกเว้นภาษี - ส่วนลดยกเว้นภาษี - (เงินมัดจำ - (รวมมูลค่าภาษี - ส่วนลดภาษี) ) )
        double excessDeposit = depositAmount - totalAmountVatAfterDiscount;
        state.screenData.totalexceptvat = totalAmountVatCale1 -
            state.screenData.totaldiscountexceptvatamount! -
            excessDeposit;
      } else {
        // ถ้าเงินมัดจำ น้อยกว่า รวมมูลค่าภาษี
        /// มูลค่าก่อนภาษี = (รวมมูลค่าภาษี - ส่วนลดภาษี - เงินมัดจำ)
        state.screenData.totalbeforevat = double.parse(
          (totalAmountVatAfterDiscount - depositAmount).toStringAsFixed(global.getDecimalDocument()),
        );

        /// มูลค่าภาษี = 0
        state.screenData.totalvatvalue = 0.0;

        /// มูลค่ายกเว้นภาษี = เหมือนเดิม
        state.screenData.totalexceptvat =
            totalAmountVatCale1 - state.screenData.totaldiscountexceptvatamount!;
      }
    }

    // ตรวจสอบค่า NaN และกำหนดเป็น 0 หากจำเป็น
    if (state.screenData.totalbeforevat.isNaN) {
      state.screenData.totalbeforevat = 0.0;
    }

    if (state.screenData.totalvatvalue.isNaN) {
      state.screenData.totalvatvalue = 0.0;
    }

    // ตรวจสอบไม่ให้ค่าเป็นลบ
    if (state.screenData.totalbeforevat < 0) {
      state.screenData.totalbeforevat = 0.0;
    }

    if (state.screenData.totalvatvalue < 0) {
      state.screenData.totalvatvalue = 0.0;
    }

    if (state.screenData.totalexceptvat < 0) {
      state.screenData.totalexceptvat = 0.0;
    }

    /// ยอดรวมสินค้ามีภาษี
    state.screenData.totalaftervat = double.parse(
      (state.screenData.totalbeforevat + state.screenData.totalvatvalue)
          .toStringAsFixed(global.getDecimalDocument()),
    );

    if (state.screenData.totalaftervat.isNaN) {
      state.screenData.totalaftervat = 0.0;
    }

    if (state.screenData.totalexceptvat.isNaN) {
      state.screenData.totalexceptvat = 0.0;
    }

    /// ยอดรวมก่อนหักส่วนลดท้ายบิล
    state.screenData.detailtotalamount = double.parse(
      (state.screenData.totalaftervat + state.screenData.totalexceptvat)
          .toStringAsFixed(global.getDecimalDocument()),
    );
    if (state.screenData.detailtotalamount!.isNaN) {
      state.screenData.detailtotalamount = 0.0;
    }

    /// ส่วนลดท้ายบิล

    if (discountWordData.contains('%')) {
      String discount = discountWordData.replaceAll('%', '');
      state.screenData.totaldiscount = double.parse(
        ((state.screenData.detailtotalamount! * double.parse(discount)) / 100)
            .toStringAsFixed(global.getDecimalDocument()),
      );
    } else if (discountWordData.trim().isNotEmpty) {
      state.screenData.totaldiscount = double.parse(discountWordData.trim());
    }

    /// ยอดรวมทั้งสิ้น
    state.screenData.totalamountafterdiscount = double.parse(
      (state.screenData.detailtotalamount! - state.screenData.totaldiscount)
          .toStringAsFixed(global.getDecimalDocument()),
    );
    if (state.screenData.totalamountafterdiscount!.isNaN) {
      state.screenData.totalamountafterdiscount = 0.0;
    }

    /// ยอดรวมสุทธิ (ในสกุลเงินหลัก - Base Currency)
    state.screenData.totalamount = double.parse(
      (state.screenData.detailtotalamount! - state.screenData.totaldiscount)
          .toStringAsFixed(global.getDecimalDocument()),
    );

    if (state.screenData.totalamount.isNaN) {
      state.screenData.totalamount = 0.0;
    }
    
    /// ============================================
    /// คำนวณยอดรวมในสกุลเงินเอกสาร (Document Currency)
    /// ============================================
    /// 
    /// ใช้ TransactionCurrencyUtils เพื่อความ consistency
    
    final screenData = state.screenData as TransactionModel;
    
    // ตรวจสอบว่าเป็น Multi-Currency หรือไม่
    final bool isMultiCurrency = TransactionCurrencyUtils.isMultiCurrency(
      docCurrency: screenData.docCurrency,
      baseCurrency: screenData.currency,
      exchangeRate: screenData.exchangerate,
    );

    // คำนวณยอดในสกุลเงินเอกสาร
    if (isMultiCurrency) {
      final rate = screenData.safeExchangeRate;
      screenData.totalamountDoc = TransactionCurrencyUtils.calculateDocumentAmount(screenData.totalamount, rate);
      screenData.totalvalueDoc = TransactionCurrencyUtils.calculateDocumentAmount(screenData.totalvalue.toDouble(), rate);
      screenData.totaldiscountDoc = TransactionCurrencyUtils.calculateDocumentAmount(screenData.totaldiscount.toDouble(), rate);
      screenData.totalvatvalueDoc = TransactionCurrencyUtils.calculateDocumentAmount(screenData.totalvatvalue.toDouble(), rate);
      screenData.totalbeforevatDoc = TransactionCurrencyUtils.calculateDocumentAmount(screenData.totalbeforevat.toDouble(), rate);
      screenData.totalaftervatDoc = TransactionCurrencyUtils.calculateDocumentAmount(screenData.totalaftervat.toDouble(), rate);
    } else {
      // ไม่มี Multi-Currency - copy ค่าจาก Base Currency
      screenData.totalamountDoc = screenData.totalamount;
      screenData.totalvalueDoc = screenData.totalvalue.toDouble();
      screenData.totaldiscountDoc = screenData.totaldiscount.toDouble();
      screenData.totalvatvalueDoc = screenData.totalvatvalue.toDouble();
      screenData.totalbeforevatDoc = screenData.totalbeforevat.toDouble();
      screenData.totalaftervatDoc = screenData.totalaftervat.toDouble();
    }

    // ตรวจสอบความถูกต้องของข้อมูล
    if (screenData.totalamountDoc?.isNaN ?? true) {
      screenData.totalamountDoc = screenData.totalamount;
    }
    if (screenData.totalamountDoc?.isInfinite ?? false) {
      screenData.totalamountDoc = screenData.totalamount;
    }

    // if (screenData.totaldiscount > screenData.totalvalue) {
    //   screenData.totaldiscount = screenData.totalvalue;
    //   screenData.discountword = screenData.totalvalue.toString();
    //   docDiscountWordController.text = screenData.totalvalue.toString();

    //   _showAlertDiscountDialog(context);
    // }

    /// คำนวณผลต่างจากเอกสารเดิม
    if (state.transactionType == global.TransactionTypeEnum.salereturn ||
        state.transactionType == global.TransactionTypeEnum.purchasereturn) {
      state.screenData.reftotaldiff =
          (state.screenData.totalbeforevat + state.screenData.totalexceptvat);
      state.screenData.reftotalcorrect = state.screenData.reftotaloriginal! -
          state.screenData.reftotaldiff!;
    }

    state.setState(() {});
  }

  /// คำนวณยอดชำระเงินทั้งหมด
  /// ข้าม PO, Quotation, Sale Order เพราะไม่มีการชำระเงิน
  void calPayTotal() {
    // PO, Quotation, Sale Order, PR ไม่มีการชำระเงิน - skip payment calculation
    if (state.transactionType == global.TransactionTypeEnum.purchaseorder ||
        state.transactionType == global.TransactionTypeEnum.quotation ||
        state.transactionType == global.TransactionTypeEnum.saleorder ||
        state.transactionType == global.TransactionTypeEnum.purchaserequisition) {
      return;
    }

    double totalPayCash = 0;
    double totalPayDelivery = 0;
    double roundAmount = 0;
    double totalPayCreditCard = 0;
    double totalPayTransfer = 0;
    double totalPayCheque = 0;
    double totalPayCoupon = 0;
    double totalQr = 0;

    totalPayCash = double.parse(
      state.payCashAmountController.text.replaceAll(',', ''),
    );
    totalPayDelivery = double.parse(
      state.payDeliveryCashAmountController.text.replaceAll(',', ''),
    );

    // ถ้า roundAmountController เป็นค่าว่าง ให้เป็น 0 เพื่อการคำนวณ
    String roundText = state.roundAmountController.text.replaceAll(',', '');
    roundAmount = double.parse(roundText.isEmpty ? "0" : roundText);

    if (state.payTransferAmountController.isNotEmpty) {
      for (var element in state.payTransferAmountController) {
        totalPayTransfer += double.parse(
          (element.text.isNotEmpty) ? element.text.replaceAll(',', '') : "0",
        );
      }
    }

    if (state.payCreditCardAmountController.isNotEmpty) {
      for (var element in state.payCreditCardAmountController) {
        totalPayCreditCard += double.parse(
          (element.text.isNotEmpty) ? element.text.replaceAll(',', '') : "0",
        );
      }
    }

    for (var element in state.payChequeAmountController) {
      totalPayCheque += double.parse(
        (element.text.isNotEmpty) ? element.text.replaceAll(',', '') : "0",
      );
    }

    for (var element in state.payCouponAmountController) {
      totalPayCoupon += double.parse(
        (element.text.isNotEmpty) ? element.text.replaceAll(',', '') : "0",
      );
    }

    for (var element in state.payQrAmountController) {
      totalQr += double.parse(
        (element.text.isNotEmpty) ? element.text.replaceAll(',', '') : "0",
      );
    }

    state.screenData.paycashamount = totalPayCash;
    state.screenData.deliveryamount = totalPayDelivery;
    state.screenData.summoneytransfer = totalPayTransfer;
    state.screenData.sumcreditcard = totalPayCreditCard;
    state.screenData.sumcheque = totalPayCheque;
    state.screenData.sumcoupon = totalPayCoupon;
    state.screenData.sumqrcode = totalQr;
    state.screenData.roundamount = roundAmount;

    state.payTotalBill = totalPayCash +
        totalPayTransfer +
        totalPayCreditCard +
        totalPayCheque +
        totalPayCoupon +
        totalQr +
        state.screenData.sumcredit! +
        totalPayDelivery +
        state.screenData.paypointamount! +
        (state.screenData.sumadvancepayment ?? 0);

    state.setState(() {});
  }

  /// ตรวจสอบความถูกต้องของการชำระเงิน
  bool verifyPayment() {
    if (state.screenData.iscancel) {
      return true;
    }

    List<String> errorList = [];

    // Debug log
    debugPrint('[verifyPayment] transactionType: ${state.transactionType}');
    debugPrint('[verifyPayment] inquirytype: ${state.screenData.inquirytype}');
    debugPrint('[verifyPayment] custcode: "${state.screenData.custcode}"');

    // เพิ่มการตรวจสอบ custcode เมื่อ inquirytype == 0
    if (state.screenData.inquirytype == 0 &&
        state.screenData.custcode.trim().isEmpty) {
      if (state.transactionType == global.TransactionTypeEnum.sale ||
          state.transactionType == global.TransactionTypeEnum.saleorder ||
          state.transactionType == global.TransactionTypeEnum.salereturn) {
        errorList.add(global.language("please_input_custcode_debtor"));
      } else if (state.transactionType == global.TransactionTypeEnum.purchase ||
          state.transactionType == global.TransactionTypeEnum.purchaseorder ||
          state.transactionType == global.TransactionTypeEnum.purchasereturn ||
          state.transactionType == global.TransactionTypeEnum.purchasepartial) {
        errorList.add(global.language("please_input_custcode_creditor"));
      }
    }

    // เพิ่มการตรวจสอบ tax_docno สำหรับ purchase เมื่อ vattype = 0, 1, หรือ 2
    if (state.transactionType == global.TransactionTypeEnum.purchase &&
        (state.screenData.vattype == 0 ||
            state.screenData.vattype == 1 ||
            state.screenData.vattype == 2) &&
        state.screenData.taxdocno.trim().isEmpty) {
      errorList.add(global.language("please_input_tax_docno"));
    }

    // ถ้าเป็น purchaseorder, saleorder, quotation, purchaserequisition และไม่มี error จาก custcode ให้ return true ทันที
    // (PR ไม่มี payment — ไม่ต้องตรวจ payTotalBill)
    if (state.transactionType == global.TransactionTypeEnum.purchaseorder ||
        state.transactionType == global.TransactionTypeEnum.saleorder ||
        state.transactionType == global.TransactionTypeEnum.purchasepartial ||
        state.transactionType == global.TransactionTypeEnum.quotation ||
        state.transactionType == global.TransactionTypeEnum.purchaserequisition) {
      if (errorList.isNotEmpty) {
        showDialog(
          context: state.context,
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

    // ตรวจสอบการชำระเงินสำหรับ transaction type อื่นๆ
    double totalWithRound =
        double.parse(state.screenData.totalamount.toStringAsFixed(global.getDecimalDocument())) +
            double.parse(state.screenData.roundamount!.toString());

    if (state.transactionType == global.TransactionTypeEnum.purchasereturn ||
        state.transactionType == global.TransactionTypeEnum.salereturn) {
      if (state.payTotalBill > totalWithRound &&
          state.screenData.inquirytype == 2 &&
          state.screenData.inquirytype == 3) {
        errorList.add(global.language("payment_over"));
      } else if (state.payTotalBill < (totalWithRound) &&
          state.screenData.inquirytype == 2 &&
          state.screenData.inquirytype == 3) {
        errorList.add(global.language("payment_less"));
      }
    } else {
      if (state.transactionType != global.TransactionTypeEnum.stockpickupproduct) {
        if (state.payTotalBill > totalWithRound &&
            state.screenData.inquirytype == 1) {
          errorList.add(global.language("payment_over"));
        } else if (state.payTotalBill < (totalWithRound) &&
            state.screenData.inquirytype == 1) {
          errorList.add(global.language("payment_less"));
        }
      }
    }

    if (errorList.isNotEmpty) {
      showDialog(
        context: state.context,
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
}
