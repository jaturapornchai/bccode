// ignore_for_file: use_build_context_synchronously

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/employee_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/sale_channel_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/model/transport_channel_model.dart';
import 'package:smlaicloud/screen_search/customer_search_screen.dart';
import 'package:smlaicloud/screen_search/employee_search_screen.dart';
import 'package:smlaicloud/screen_search/sale_chanels_search_screen.dart';
import 'package:smlaicloud/screen_search/supplier_search_screen.dart';
import 'package:smlaicloud/screen_search/transaction_search_screen.dart';
import 'package:smlaicloud/screen_search/transport_search_screen.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_calculator.dart';

/// Helper class สำหรับ Search handler functions
/// รองรับทั้ง TransactionEditScreenState และ PurchaseOrderEditScreenState
class TransactionSearchHandlers {
  final dynamic state;
  late final TransactionCalculator _calculator;

  TransactionSearchHandlers(this.state) {
    _calculator = TransactionCalculator(state);
  }

  /// ค้นหาพนักงานขาย
  void searchSale({required String word}) {
    Navigator.push(
      state.context,
      MaterialPageRoute(builder: (context) => EmployeeSearchScreen(word: word)),
    ).then((value) {
      EmployeeModel result = value;
      if (result.code.trim().isNotEmpty) {
        state.setState(() {
          state.saleCodeController.text = result.code;
          state.saleNameController.text = result.name;

          state.screenData.salecode = result.code;
          state.screenData.salename = result.name;
        });
      }
    });
  }

  /// ค้นหาช่องทางการขาย
  void searchSaleChannel({required String word}) {
    Navigator.push(
      state.context,
      MaterialPageRoute(
        builder: (context) => SalechannelsSearchScreen(word: word),
      ),
    ).then((value) {
      SaleChannelModel result = value;
      if (result.code!.trim().isNotEmpty) {
        state.setState(() {
          state.screenData.salechannelcode = result.code;
          state.screenData.salechannelgp = result.gp!;
          state.screenData.salechannelgptype = result.gptype!;
        });
      }
    });
  }

  /// ค้นหาช่องทางการขนส่ง
  void searchTransport({required String word}) {
    Navigator.push(
      state.context,
      MaterialPageRoute(
        builder: (context) => TransportSearchScreen(word: word),
      ),
    ).then((value) {
      TransportChannelModel result = value;
      if (result.code.trim().isNotEmpty) {
        state.setState(() {
          state.screenData.transportcode = result.code;
        });
      }
    });
  }

  /// ค้นหาเอกสารอ้างอิง
  void searchDocRef({required String word}) {
    // ตรวจสอบว่าเลือกลูกค้าแล้วหรือไม่ (ยกเว้น stockreturnproduct)
    if (state.transactionType != global.TransactionTypeEnum.stockreturnproduct &&
        state.custCodeController.text.trim().isEmpty) {
      String message = "";
      if (state.transactionType == global.TransactionTypeEnum.sale ||
          state.transactionType == global.TransactionTypeEnum.saleorder ||
          state.transactionType == global.TransactionTypeEnum.salereturn ||
          state.transactionType == global.TransactionTypeEnum.advancePayment ||
          state.transactionType == global.TransactionTypeEnum.deposit ||
          state.transactionType == global.TransactionTypeEnum.paidAdvance ||
          state.transactionType == global.TransactionTypeEnum.paidAdvanceRefund ||
          state.transactionType == global.TransactionTypeEnum.receiveDepositRefund) {
        message = global.language("please_input_custcode_debtor");
      } else {
        message = global.language("please_input_custcode_creditor");
      }

      showDialog(
        context: state.context,
        builder: (BuildContext context) {
          return AlertDialog(
            title: Text(global.language("alert")),
            content: Text(message),
            actions: [
              TextButton(
                child: Text(global.language("close")),
                onPressed: () {
                  Navigator.pop(context);
                },
              ),
            ],
          );
        },
      );
      return;
    }

    // กำหนด mapping ระหว่าง transaction type และ search type
    final Map<global.TransactionTypeEnum, global.TransactionTypeEnum>
        typeMapping = {
      global.TransactionTypeEnum.purchase:
          global.TransactionTypeEnum.purchaseorder,
      global.TransactionTypeEnum.purchasereturn:
          global.TransactionTypeEnum.purchase,
      global.TransactionTypeEnum.sale: global.TransactionTypeEnum.saleorder,
      global.TransactionTypeEnum.salereturn: global.TransactionTypeEnum.sale,
      global.TransactionTypeEnum.stockreturnproduct:
          global.TransactionTypeEnum.stockpickupproduct,
      global.TransactionTypeEnum.advancePaymentRefund:
          global.TransactionTypeEnum.advancePayment,
      global.TransactionTypeEnum.depositRefund:
          global.TransactionTypeEnum.deposit,
      global.TransactionTypeEnum.receiveDepositRefund:
          global.TransactionTypeEnum.receiveDeposit,
      global.TransactionTypeEnum.accrualreceive:
          global.TransactionTypeEnum.purchasepartial,
      global.TransactionTypeEnum.purchasepartial:
          global.TransactionTypeEnum.purchaseorder,
      global.TransactionTypeEnum.paidAdvanceRefund:
          global.TransactionTypeEnum.paidAdvance,
    };

    // ตรวจสอบ transaction type และกำหนด search type
    global.TransactionTypeEnum searchType =
        typeMapping[state.transactionType] ?? state.transactionType;

    // เก็บข้อมูลปัจจุบันไว้ก่อนค้นหา (สำหรับ transaction type ที่มี mapping)
    final Map<String, String>? currentState = typeMapping.containsKey(state.transactionType)
        ? {
            'docno': state.screenData.docno,
            'docdatetime': state.screenData.docdatetime,
            'guidfixed': state.screenData.guidfixed ?? '',
          }
        : null;

    Navigator.push(
      state.context,
      MaterialPageRoute(
        builder: (context) => TransSearchScreen(
          custcode: state.screenData.custcode,
          type: searchType,
        ),
      ),
    ).then((value) {
      if (value != null) {
        TransactionModel result = value;
        if (result.docno.isNotEmpty) {
          handleSearchResult(result, currentState);
        }
      }
    });
  }

  /// จัดการผลลัพธ์จากการค้นหาเอกสารอ้างอิง
  void handleSearchResult(
    TransactionModel? result,
    Map<String, String>? currentState,
  ) {
    try {
      // ตรวจสอบ result ก่อน
      if (result == null) {
        if (kDebugMode) {
          AppLogger.debug('result is null');
        }
        return;
      }

      // ตรวจสอบข้อมูลพื้นฐานก่อน
      if (result.docno.isEmpty) {
        if (kDebugMode) {
          AppLogger.debug('เอกสารที่เลือกไม่มีหมายเลขเอกสาร');
        }
        return;
      }

      bool shouldReplaceData =
          currentState != null &&
          (state.transactionType == global.TransactionTypeEnum.purchasereturn ||
              state.transactionType == global.TransactionTypeEnum.salereturn ||
              state.transactionType == global.TransactionTypeEnum.stockreturnproduct ||
              state.transactionType == global.TransactionTypeEnum.advancePaymentRefund ||
              state.transactionType == global.TransactionTypeEnum.depositRefund ||
              state.transactionType == global.TransactionTypeEnum.paidAdvanceRefund ||
              state.transactionType == global.TransactionTypeEnum.receiveDepositRefund);

      if (shouldReplaceData) {
        try {
          // คืนค่าข้อมูลที่ต้องการเก็บไว้
          state.screenData.guidfixed = currentState['guidfixed'];
          state.screenData.docno = currentState['docno']!;
          state.screenData.docdatetime = currentState['docdatetime']!;
          // ตั้งค่าข้อมูลอ้างอิง
          state.screenData.docrefno = result.docno;
          state.screenData.docrefdate = result.docdatetime;
        } catch (e) {
          if (kDebugMode) {
            AppLogger.error('Error replacing data: $e');
          }
          if (state.mounted && state.context.mounted) {
            global.showErrorSnackBar(state.context, global.language('error_update_data'));
          }
          return;
        }
      }

      // เพิ่มเอกสารอ้างอิงใหม่เข้าไปใน array
      try {
        final newDocRef = DocReferenceModel(
          docdatetime: result.docdatetime,
          docno: result.docno,
          guidfixed: result.guidfixed ?? '',
        );

        // ตรวจสอบว่าเอกสารนี้มีอยู่แล้วหรือไม่
        state.screenData.docreferences ??= <DocReferenceModel>[];

        bool exists = state.screenData.docreferences!.any(
          (ref) => ref.docno == newDocRef.docno,
        );
        if (exists) {
          if (state.mounted && state.context.mounted) {
            showDialog(
              context: state.context,
              builder: (BuildContext context) {
                return AlertDialog(
                  title: Text(global.language("alert")),
                  content: Text(
                    '${global.language("document")} ${newDocRef.docno} ${global.language("document_already_exists")}',
                  ),
                  actions: [
                    TextButton(
                      child: Text(global.language("close")),
                      onPressed: () {
                        Navigator.pop(context);
                      },
                    ),
                  ],
                );
              },
            );
          }
          return;
        }

        // เพิ่มเอกสารอ้างอิง
        state.screenData.docreferences!.add(newDocRef);
      } catch (e) {
        if (kDebugMode) {
          AppLogger.error('Error adding document reference: $e');
        }
        if (state.mounted && state.context.mounted) {
          global.showErrorSnackBar(state.context, global.language('error_add_reference'));
        }
        return;
      }

      // เพิ่มรายการสินค้าจากเอกสารอ้างอิง (ถ้าจำเป็น)
      addProductsFromDocRef(result);

      // Handle special case for advance payment refund
      if (state.transactionType == global.TransactionTypeEnum.advancePaymentRefund ||
          state.transactionType == global.TransactionTypeEnum.depositRefund) {
        try {
          if (result.details != null) {
            state.screenData.details = result.details!
                .map(
                  (detail) => TransactionDetailModel.fromJson(detail.toJson()),
                )
                .toList();
          }
        } catch (e) {
          if (kDebugMode) {
            AppLogger.error('Error setting details for refund: $e');
          }
        }
      }

      if (state.mounted) {
        state.setState(() {});

        if (shouldReplaceData) {
          try {
            state.loadDataToScreen();
          } catch (e) {
            if (kDebugMode) {
              AppLogger.error('Error loading data to screen: $e');
            }
          }
        }
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error in handleSearchResult: $e');
        if (kDebugMode) {
          AppLogger.debug('Stack trace: ${StackTrace.current}');
        }
      }

      if (state.mounted && state.context.mounted) {
        global.showErrorSnackBar(state.context, '${global.language("error_process_document")}: ${e.toString()}');
      }
    }
  }

  /// เพิ่มรายการสินค้าจากเอกสารอ้างอิง
  void addProductsFromDocRef(TransactionModel result) {
    try {
      // เพิ่มการตรวจสอบ null ที่เข้มงวดขึ้น
      bool shouldAddProducts =
          state.transactionType != global.TransactionTypeEnum.advancePaymentRefund &&
          state.transactionType != global.TransactionTypeEnum.depositRefund &&
          result.details != null &&
          result.details!.isNotEmpty;

      if (!shouldAddProducts) {
        if (kDebugMode) {
          AppLogger.debug(
            'ไม่ต้องเพิ่มรายการสินค้าสำหรับ transaction type: ${state.transactionType}',
          );
        }
        return;
      }

      // เพิ่มการตรวจสอบ screenData.details
      state.screenData.details ??= <TransactionDetailModel>[];

      // ตรวจสอบ result.details อีกครั้ง
      if (result.details == null || result.details!.isEmpty) {
        if (kDebugMode) {
          AppLogger.debug('ไม่มีรายการสินค้าในเอกสารอ้างอิง');
        }
        return;
      }

      // คำนวณยอดรวมเอกสารอ้างอิง - เพิ่มการตรวจสอบ null safety
      double reftotaloriginal = 0;
      try {
        double beforeVat = result.totalbeforevat ?? 0;
        double exceptVat = result.totalexceptvat ?? 0;
        reftotaloriginal = beforeVat + exceptVat;
      } catch (e) {
        if (kDebugMode) {
          AppLogger.error('Error calculating reftotaloriginal: $e');
        }
        reftotaloriginal = 0;
      }

      // เพิ่มข้อมูลเอกสารอ้างอิงใน docrefs array
      try {
        String docno = result.docno;
        String docdatetime = result.docdatetime;

        state.docrefs.add({
          'docref': docno,
          'totaloriginal': reftotaloriginal,
          'docdatetime': docdatetime,
        });
      } catch (e) {
        if (kDebugMode) {
          AppLogger.error('Error adding to docrefs: $e');
        }
      }

      // อัพเดทยอดรวมเอกสารเดิมใน screenData
      try {
        state.screenData.reftotaloriginal =
            (state.screenData.reftotaloriginal ?? 0) + reftotaloriginal;
      } catch (e) {
        if (kDebugMode) {
          AppLogger.error('Error updating reftotaloriginal: $e');
        }
        state.screenData.reftotaloriginal = reftotaloriginal;
      }

      // เพิ่มรายการสินค้าทั้งหมด
      for (var detail in result.details!) {
        try {
          // ตรวจสอบ calcflag ก่อนใช้
          int safeCalcflag = state.calcflag ?? 0;

          var newDetail = TransactionDetailModel(
            docdatetime: state.screenData.docdatetime,
            itemguid: detail.itemguid,
            barcode: detail.barcode,
            itemcode: detail.itemcode,
            docref: result.docno,
            docrefdatetime:
                result.docdatetime ?? DateTime.now().toIso8601String(),
            itemnames: detail.itemnames ?? [],
            unitcode: detail.unitcode ?? '',
            qty: detail.qty ?? 0,
            price: detail.price ?? 0,
            discount: detail.discount ?? '',
            sumofcost: detail.sumofcost ?? 0,
            sumamount: detail.sumamount ?? 0,
            remark: detail.remark ?? '',
            linenumber: (state.screenData.details?.length ?? 0) + 1,
            whcode: detail.whcode ?? '',
            whnames: detail.whnames ?? [],
            locationcode: detail.locationcode ?? '',
            locationnames: detail.locationnames ?? [],
            towhcode: detail.towhcode ?? '',
            towhnames: detail.towhnames ?? [],
            tolocationcode: detail.tolocationcode ?? '',
            tolocationnames: detail.tolocationnames ?? [],
            shelfcode: detail.shelfcode ?? '',
            totalqty: detail.totalqty ?? 0,
            discountamount: detail.discountamount ?? 0,
            averagecost: detail.averagecost ?? 0,
            totalvaluevat: detail.totalvaluevat ?? 0,
            sumamountexcludevat: detail.sumamountexcludevat ?? 0,
            priceexcludevat: detail.priceexcludevat ?? 0,
            standvalue: detail.standvalue ?? 1,
            dividevalue: detail.dividevalue ?? 1,
            calcflag: safeCalcflag,
            vattype: detail.vattype ?? 0,
            ispos: detail.ispos ?? 0,
            laststatus: detail.laststatus ?? 0,
            itemtype: detail.itemtype ?? 0,
            taxtype: detail.taxtype ?? 0,
            vatcal: detail.vatcal ?? 0,
            inquirytype: detail.inquirytype ?? 0,
            multiunit: detail.multiunit ?? false,
            unitnames: detail.unitnames ?? [],
            refbarcodes: detail.refbarcodes ?? [],
            sku: detail.sku ?? '',
            extrajson: detail.extrajson ?? '',
            extrajsonlist: detail.extrajsonlist ?? [],
            manufacturerguid: detail.manufacturerguid ?? '',
            description: detail.description ?? '',
            imageuri: detail.imageuri ?? '',
            eventqty: detail.qty ?? 0,
          );

          // เพิ่มรายการใหม่เข้าไปใน screenData.details
          state.screenData.details!.add(newDetail);
        } catch (e) {
          if (kDebugMode) {
            AppLogger.error('Error creating detail item: $e');
            if (kDebugMode) {
              AppLogger.debug('Stack trace: ${StackTrace.current}');
            }
          }
          // ข้ามรายการนี้และไปรายการถัดไป
          continue;
        }
      }

      // คำนวณยอดรวมใหม่
      try {
        if (state.mounted) {
          // ตรวจสอบว่า widget ยังอยู่หรือไม่
          _calculator.calTotalValue();
        }
      } catch (e) {
        if (kDebugMode) {
          AppLogger.error('Error calculating total value: $e');
        }

        if (state.mounted && state.context.mounted) {
          global.showWarningSnackBar(state.context, global.language('error_calculate_total'));
        }
      }
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error in addProductsFromDocRef: $e');
        if (kDebugMode) {
          AppLogger.debug('Stack trace: ${StackTrace.current}');
        }
      }

      // แสดงข้อความแจ้งเตือนให้ผู้ใช้ทราบ
      if (state.mounted && state.context.mounted) {
        global.showErrorSnackBar(state.context, '${global.language("error_add_product")}: ${e.toString()}');
      }
    }
  }

  /// ค้นหาธุรกรรม
  /// จอใหญ่ (> 800px) จะเปิด side panel แทน popup
  void searchTrans() {
    final screenWidth = MediaQuery.of(state.context).size.width;
    final isLargeScreen = screenWidth > 800;

    // จอใหญ่ เปิด side panel แทน popup
    if (isLargeScreen) {
      state.setState(() {
        state.showDocSearchPanel = true;
      });
      return;
    }

    // จอเล็ก ใช้ popup เหมือนเดิม
    Navigator.push(
      state.context,
      MaterialPageRoute(
        builder: (context) => TransSearchScreen(
          custcode: '',
          type: state.transactionType,
        ),
      ),
    ).then((value) {
      TransactionModel result = value;
      if (result.docno.isNotEmpty) {
        DateTime localDateTime = DateTime.parse(result.docdatetime);
        state.screenData = result;
        // Store the initial copy in screenDataTemp to keep it unchanged
        state.screenDataTemp = result.copyWith(
          details: result.details?.map((detail) => detail.copyWith()).toList(),
          paymentdetail: result.paymentdetail?.copyWith(),
          paymentStructs: result.paymentStructs
              ?.map((struct) => struct.copyWith())
              .toList(),
        );

        /// sort  linenumber
        state.screenData.details!.sort(
          (a, b) => a.linenumber.compareTo(b.linenumber),
        );
        state.screenData.docdatetime = localDateTime.toUtc().toIso8601String(); // แปลงเป็น UTC ตามกฎ CLAUDE.md
        state.showPayDetail == 0;

        state.loadDataToScreen();
        // คำนวณยอดรวมใหม่
        _calculator.calTotalValue();
        //
        state.setState(() {});
      }
    });
  }

  /// เลือกเอกสารอ้างอิงสำหรับรายการสินค้า
  void selectDocRefForItem({required int itemIndex}) {
    // ตรวจสอบว่าเป็น purchase transaction
    if (state.transactionType != global.TransactionTypeEnum.purchase &&
        state.transactionType != global.TransactionTypeEnum.purchasereturn &&
        state.transactionType != global.TransactionTypeEnum.purchasepartial &&
        state.transactionType != global.TransactionTypeEnum.accrualreceive &&
        state.transactionType != global.TransactionTypeEnum.sale &&
        state.transactionType != global.TransactionTypeEnum.salereturn) {
      return;
    }

    // ตรวจสอบว่าเลือกลูกค้าแล้วหรือไม่
    if (state.custCodeController.text.trim().isEmpty) {
      showDialog(
        context: state.context,
        builder: (BuildContext context) {
          return AlertDialog(
            title: Text(global.language("warning")),
            content: Text(global.language("please_input_custcode_creditor")),
            actions: [
              TextButton(
                child: Text(global.language("confirm")),
                onPressed: () {
                  Navigator.of(context).pop();
                },
              ),
            ],
          );
        },
      );
      return;
    }

    // กำหนดประเภทเอกสารที่จะค้นหาตามประเภทเอกสารปัจจุบัน
    global.TransactionTypeEnum transType;

    if (state.transactionType == global.TransactionTypeEnum.purchase ||
        state.transactionType == global.TransactionTypeEnum.purchasepartial) {
      transType = global.TransactionTypeEnum.purchaseorder;
    } else if (state.transactionType == global.TransactionTypeEnum.accrualreceive) {
      transType = global.TransactionTypeEnum.purchasepartial;
    } else if (state.transactionType == global.TransactionTypeEnum.sale) {
      transType = global.TransactionTypeEnum.saleorder;
    } else if (state.transactionType == global.TransactionTypeEnum.salereturn) {
      transType = global.TransactionTypeEnum.sale;
    } else {
      transType = global.TransactionTypeEnum.purchase;
    }

    Navigator.push(
      state.context,
      MaterialPageRoute(
        builder: (context) => TransSearchScreen(
          custcode: state.custCodeController.text.trim(),
          type: transType,
        ),
      ),
    ).then((value) {
      if (value.guidfixed.isNotEmpty) {
        TransactionModel result = value;

        // ตรวจสอบว่ารายการปัจจุบันมี barcode หรือไม่
        String currentBarcode = state.screenData.details![itemIndex].barcode;
        if (currentBarcode.trim().isEmpty) {
          // ถ้าไม่มี barcode ให้แสดงข้อความแจ้งเตือน
          showDialog(
            context: state.context,
            builder: (BuildContext context) {
              return AlertDialog(
                title: Text(global.language("warning")),
                content: Text(global.language('no_barcode_cannot_select')),
                actions: [
                  TextButton(
                    child: Text(global.language("confirm")),
                    onPressed: () {
                      Navigator.of(context).pop();
                    },
                  ),
                ],
              );
            },
          );
        } else {
          // ตรวจสอบว่า barcode ของรายการปัจจุบันมีอยู่ใน PO ที่เลือกหรือไม่
          bool barcodeExistsInPO = false;
          if (result.details != null && result.details!.isNotEmpty) {
            barcodeExistsInPO = result.details!.any(
              (detail) => detail.barcode == currentBarcode,
            );
          }

          if (barcodeExistsInPO) {
            // ถ้ามี barcode ใน PO ให้อัพเดทได้
            state.setState(() {
              state.screenData.details![itemIndex].docref = result.docno;
              state.screenData.details![itemIndex].docrefdatetime =
                  result.docdatetime;
            });

            // แสดงข้อความแจ้งเตือน
            global.showSuccessSnackBar(state.context, '${global.language("update_reference_doc")}: ${result.docno}');
          } else {
            // ถ้าไม่มี barcode ใน PO ให้แสดงข้อความแจ้งเตือน
            showDialog(
              context: state.context,
              builder: (BuildContext context) {
                return AlertDialog(
                  title: Text(global.language("warning")),
                  content: Text(
                    '${global.language("no_product_in_po")} "$currentBarcode"',
                  ),
                  actions: [
                    TextButton(
                      child: Text(global.language("confirm")),
                      onPressed: () {
                        Navigator.of(context).pop();
                      },
                    ),
                  ],
                );
              },
            );
          }
        }
      }
    });
  }

  /// ค้นหาลูกค้า
  void searchCustomer({required String word}) {
    Navigator.push(
      state.context,
      MaterialPageRoute(builder: (context) => CustomerSearchScreen(word: word)),
    ).then((value) {
      global.SearchDebtorModel result = value;
      if (result.code.trim().isNotEmpty) {
        // ตรวจสอบว่ามีรายการสินค้าอยู่แล้วและรหัสลูกค้าใหม่แตกต่างจากเดิมหรือไม่
        if (state.screenData.details!.isNotEmpty &&
            state.custCodeController.text.trim() != result.code.trim()) {
          // แสดงข้อความแจ้งเตือน
          showDialog(
            context: state.context,
            builder: (BuildContext context) {
              return AlertDialog(
                title: Text(global.language("alert")),
                content: Text(global.language("customer_code_changed_warning")),
                actions: [
                  TextButton(
                    child: Text(global.language("cancel")),
                    onPressed: () {
                      Navigator.of(context).pop();
                      // ไม่ทำอะไรถ้าผู้ใช้กด cancel
                    },
                  ),
                  TextButton(
                    child: Text(global.language("confirm")),
                    onPressed: () {
                      Navigator.of(context).pop();
                      // ดำเนินการต่อเมื่อผู้ใช้กด confirm
                      updateCustomerData(result);
                    },
                  ),
                ],
              );
            },
          );
        } else {
          // ถ้าไม่มีรายการสินค้าหรือรหัสลูกค้าเดิม ให้ดำเนินการปกติ
          updateCustomerData(result);
        }
      }
    });
  }

  /// อัพเดทข้อมูลลูกค้า
  void updateCustomerData(global.SearchDebtorModel result) {
    state.setState(() {
      // เก็บ pricelevel ของลูกค้าที่เลือก
      state.customerPriceLevel = (result.pricelevel.isNotEmpty)
          ? int.parse(result.pricelevel)
          : 1;

      // ตั้งค่า pointscode
      state.screenData.pointscode = result.pointscode;

      if (state.transactionType == global.TransactionTypeEnum.purchasereturn ||
          state.transactionType == global.TransactionTypeEnum.salereturn) {
        if (state.screenData.details!.isNotEmpty) {
          showDialog(
            context: state.context,
            builder: (BuildContext context) {
              return Center(
                child: AlertDialog(
                  content: Text(global.language("remove_all_detail")),
                  actions: [
                    TextButton(
                      child: Text(global.language("cancel")),
                      onPressed: () {
                        Navigator.of(context).pop();
                      },
                    ),
                    TextButton(
                      child: Text(
                        global.language("confirm"),
                        style: const TextStyle(color: Colors.red),
                      ),
                      onPressed: () {
                        state.screenData.discountword = '';
                        state.screenData.totalcost = 0;
                        state.screenData.totalvalue = 0;
                        state.screenData.totaldiscount = 0;
                        state.screenData.totalvatvalue = 0;
                        state.screenData.totalaftervat = 0;
                        state.screenData.totalexceptvat = 0;
                        state.screenData.totalamount = 0;
                        state.screenData.totalbeforevat = 0;
                        state.screenData.details = <TransactionDetailModel>[];
                        state.screenData.ismanualamount = false;
                        state.screenData.paymentdetailraw = "";
                        state.screenData.paymentStructs = [];
                        state.screenData.reftotaloriginal = 0;
                        state.screenData.reftotalcorrect = 0;
                        state.screenData.reftotaldiff = 0;

                        state.custCodeController.text = result.code;
                        state.custnamesController.text = global.activeLangName(
                          result.names,
                        );
                        state.screenData.custcode = result.code;
                        state.screenData.custnames = result.names;
                        _calculator.calTotalValue();
                        Navigator.of(context).pop();
                      },
                    ),
                  ],
                ),
              );
            },
          );
        } else {
          state.custCodeController.text = result.code;
          state.custnamesController.text = global.activeLangName(result.names);
          state.screenData.custcode = result.code;
          state.screenData.custnames = result.names;

          // อัพเดทราคาสินค้าที่มีอยู่แล้วตาม pricelevel ใหม่
          updateExistingProductPrices();
        }
      } else {
        state.custCodeController.text = result.code;
        state.custnamesController.text = global.activeLangName(result.names);
        state.screenData.custcode = result.code;
        state.screenData.custnames = result.names;

        // อัพเดทราคาสินค้าที่มีอยู่แล้วตาม pricelevel ใหม่
        updateExistingProductPrices();
      }
    });
  }

  /// อัพเดทราคาสินค้าที่มีอยู่แล้วตาม pricelevel ใหม่
  void updateExistingProductPrices() {
    for (int i = 0; i < state.screenData.details!.length; i++) {
      if (state.screenData.details![i].itemguid.isNotEmpty) {
        // ใช้ getProductBarcodeDetail แทน getByGuid
        state.productBarcodeRepository
            .getProductBarcodeDetail(state.screenData.details![i].barcode)
            .then((response) {
              if (response.success && response.data != null) {
                ProductBarcodeModel product = ProductBarcodeModel.fromJson(
                  response.data,
                );
                double newPrice = state.getPrice(product.prices);

                state.setState(() {
                  state.screenData.details![i].price = newPrice;
                  // คำนวณยอดรวมใหม่สำหรับรายการนี้
                  state.screenData.details![i].sumamount =
                      state.screenData.details![i].qty * newPrice;
                  _calculator.calTotalValue();
                });
              }
            })
            .catchError((error) {
              // จัดการ error ถ้าดึงข้อมูลไม่สำเร็จ
              if (kDebugMode) {
                AppLogger.debug(
                  'Error updating price for item ${state.screenData.details![i].barcode}: $error',
                );
              }
            });
      }
    }
  }

  /// ค้นหาผู้จัดจำหน่าย
  void searchSupplier({required String word}) {
    Navigator.push(
      state.context,
      MaterialPageRoute(builder: (context) => SupplierSearchScreen(word: word)),
    ).then((value) {
      SearchGuidCodeNameModel result = value;
      if (result.code.trim().isNotEmpty) {
        if (state.transactionType == global.TransactionTypeEnum.purchasereturn ||
            state.transactionType == global.TransactionTypeEnum.salereturn) {
          if (state.screenData.details!.isNotEmpty) {
            showDialog(
              context: state.context,
              builder: (BuildContext context) {
                return Center(
                  child: AlertDialog(
                    content: Text(global.language("remove_all_detail")),
                    actions: [
                      TextButton(
                        child: Text(global.language("cancel")),
                        onPressed: () {
                          Navigator.of(context).pop();
                        },
                      ),
                      TextButton(
                        child: Text(
                          global.language("confirm"),
                          style: const TextStyle(color: Colors.red),
                        ),
                        onPressed: () {
                          state.screenData.discountword = '';
                          state.screenData.totalcost = 0;
                          state.screenData.totalvalue = 0;
                          state.screenData.totaldiscount = 0;
                          state.screenData.totalvatvalue = 0;
                          state.screenData.totalaftervat = 0;
                          state.screenData.totalexceptvat = 0;
                          state.screenData.totalamount = 0;
                          state.screenData.totalbeforevat = 0;
                          state.screenData.details = <TransactionDetailModel>[];
                          state.screenData.ismanualamount = false;
                          state.screenData.paymentdetailraw = "";
                          state.screenData.paymentStructs = [];
                          state.screenData.reftotaloriginal = 0;
                          state.screenData.reftotalcorrect = 0;
                          state.screenData.reftotaldiff = 0;

                          state.custCodeController.text = result.code;
                          state.custnamesController.text = global.activeLangName(
                            result.names,
                          );
                          state.screenData.custcode = result.code;
                          state.screenData.custnames = result.names;
                          _calculator.calTotalValue();
                          Navigator.of(context).pop();
                        },
                      ),
                    ],
                  ),
                );
              },
            );
          } else {
            state.custCodeController.text = result.code;
            state.custnamesController.text = global.activeLangName(result.names);
            state.screenData.custcode = result.code;
            state.screenData.custnames = result.names;
          }
        } else {
          state.custCodeController.text = result.code;
          state.custnamesController.text = global.activeLangName(result.names);
          state.screenData.custcode = result.code;
          state.screenData.custnames = result.names;
        }
      }
      if (state.mounted) {
        state.setState(() {});
      }
    });
  }
}
