import 'package:smlaicloud/global.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/repositories/client.dart';
import 'package:smlaicloud/repositories/trans_repository.dart';
import 'package:uuid/uuid.dart';

import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:equatable/equatable.dart';

import 'package:smlaicloud/global.dart' as global;

part 'trans_event.dart';
part 'trans_state.dart';

class TransBloc extends Bloc<TransEvent, TransState> {
  final TransRepository _transRepository;

  TransBloc({required TransRepository transRepository}) : _transRepository = transRepository, super(TransInitial()) {
    on<TransLoad>(onTransLoad);
    on<TransSave>(onTransSave);
    on<TransUpdate>(onTransUpdate);
    on<TransDelete>(onTransDelete);
    on<TransCreateFullInvoice>(onTransCreateFullInvoice);
  }

  void onTransLoad(TransLoad event, Emitter<TransState> emit) async {
    emit(TransInProgress());

    try {
      int apiMode = 0; // 0=จาก mongo,1=จาก go api (postgres)
      late ApiResponse<dynamic> results;
      late GoApiQueryManyResultResponse goApiQueryResult;

      if (event.type == global.TransactionTypeEnum.purchase) {
        results = await _transRepository.getPurchaseList(limit: event.limit, offset: event.offset, search: event.search, custcode: event.custcode);
      } else if (event.type == global.TransactionTypeEnum.purchasereturn) {
        results = await _transRepository.getPurchaseReturnList(limit: event.limit, offset: event.offset, search: event.search, custcode: event.custcode);
      } else if (event.type == global.TransactionTypeEnum.sale) {
        results = await _transRepository.getSaleList(limit: event.limit, offset: event.offset, search: event.search, custcode: event.custcode, ispos: event.ispos);
      } else if (event.type == global.TransactionTypeEnum.salereturn) {
        results = await _transRepository.getSaleReturnList(limit: event.limit, offset: event.offset, search: event.search, custcode: event.custcode);
      } else if (event.type == global.TransactionTypeEnum.adjust) {
        results = await _transRepository.getAdjustList(limit: event.limit, offset: event.offset, search: event.search);
      } else if (event.type == global.TransactionTypeEnum.stockpickupproduct) {
        results = await _transRepository.getStockPickupList(limit: event.limit, offset: event.offset, search: event.search);
      } else if (event.type == global.TransactionTypeEnum.stockreceiveproduct) {
        results = await _transRepository.getStockReceiveList(limit: event.limit, offset: event.offset, search: event.search);
      } else if (event.type == global.TransactionTypeEnum.stockreturnproduct) {
        results = await _transRepository.getStockReturnList(limit: event.limit, offset: event.offset, search: event.search);
      } else if (event.type == global.TransactionTypeEnum.stocktransfer) {
        results = await _transRepository.getTransferList(limit: event.limit, offset: event.offset, search: event.search);
      } else if (event.type == global.TransactionTypeEnum.paid) {
        results = await _transRepository.getSaleByCode(limit: event.limit, offset: event.offset, custcode: event.search);
      } else if (event.type == global.TransactionTypeEnum.pay) {
        results = await _transRepository.getPurchaseByCode(limit: event.limit, offset: event.offset, custcode: event.search);
      } else if (event.type == global.TransactionTypeEnum.stockbalance) {
        results = await _transRepository.getStockBalanceList(limit: event.limit, offset: event.offset, search: event.search, custcode: event.custcode);
      } else if (event.type == global.TransactionTypeEnum.saleorder) {
        results = await _transRepository.getSaleOrderList(limit: event.limit, offset: event.offset, search: event.search, custcode: event.custcode);
      } else if (event.type == global.TransactionTypeEnum.purchaseorder) {
        apiMode = 1;
        goApiQueryResult = await _transRepository.getPurchaseOrderList(
          limit: event.limit,
          offset: event.offset,
          search: event.search,
          custcode: event.custcode,
          dateorder: event.dateorder,
          fromDate: event.fromDate,
          toDate: event.toDate,
          minAmount: event.minAmount,
          maxAmount: event.maxAmount,
          custCodes: event.custCodes,
        );
      } else if (event.type == global.TransactionTypeEnum.quotation) {
        results = await _transRepository.getQuotationList(limit: event.limit, offset: event.offset, search: event.search, custcode: event.custcode);
      } else if (event.type == global.TransactionTypeEnum.purchasepartial) {
        results = await _transRepository.getPurchasePartialList(limit: event.limit, offset: event.offset, search: event.search, custcode: event.custcode);
      } else if (event.type == global.TransactionTypeEnum.accrualreceive) {
        results = await _transRepository.getAccrualReceiveList(limit: event.limit, offset: event.offset, search: event.search, custcode: event.custcode);
      } else if (event.type == global.TransactionTypeEnum.advancePayment) {
        results = await _transRepository.getAdvancePaymentList(limit: event.limit, offset: event.offset, search: event.search, custcode: event.custcode);
      } else if (event.type == global.TransactionTypeEnum.advancePaymentRefund) {
        results = await _transRepository.getAdvancePaymentRefundList(limit: event.limit, offset: event.offset, search: event.search, custcode: event.custcode);
      } else if (event.type == global.TransactionTypeEnum.deposit) {
        results = await _transRepository.getDepositList(limit: event.limit, offset: event.offset, search: event.search, custcode: event.custcode);
      } else if (event.type == global.TransactionTypeEnum.depositRefund) {
        results = await _transRepository.getDepositRefundList(limit: event.limit, offset: event.offset, search: event.search, custcode: event.custcode);
      } else if (event.type == global.TransactionTypeEnum.paidAdvance) {
        results = await _transRepository.getPaidAdvanceList(limit: event.limit, offset: event.offset, search: event.search, custcode: event.custcode);
      } else if (event.type == global.TransactionTypeEnum.paidAdvanceRefund) {
        results = await _transRepository.getPaidAdvanceRefundList(limit: event.limit, offset: event.offset, search: event.search, custcode: event.custcode);
      } else if (event.type == global.TransactionTypeEnum.receiveDeposit) {
        results = await _transRepository.getReceiveDepositList(limit: event.limit, offset: event.offset, search: event.search, custcode: event.custcode);
      } else if (event.type == global.TransactionTypeEnum.receiveDepositRefund) {
        results = await _transRepository.getReceiveDepositRefundList(limit: event.limit, offset: event.offset, search: event.search, custcode: event.custcode);
      } else {
        emit(const TransLoadFailed(message: 'Trans Not Found'));
      }
      // go api query
      if (apiMode == 1) {
        List<TransactionModel> trans = [];
        // Debug: แสดงข้อมูลที่ได้จาก API เพื่อตรวจสอบ isref
        if (goApiQueryResult.results.isNotEmpty) {
          final firstItem = goApiQueryResult.results.first;
          AppLogger.debug('🔍 [TransBloc] First item from API - isref: ${firstItem['isref']}, islocked: ${firstItem['islocked']}, isclosed: ${firstItem['isclosed']}, docno: ${firstItem['docno']}');
          AppLogger.debug('🔍 [TransBloc] Currency fields - doc_currency: ${firstItem['doc_currency']}, doc_currencysymbol: ${firstItem['doc_currencysymbol']}, exchangerate: ${firstItem['exchangerate']}, totalamount_doc: ${firstItem['totalamount_doc']}');
        }
        for (var data in goApiQueryResult.results) {
          List<LanguageDataModel> custNames = [];
          custNames.addAll([LanguageDataModel(code: "th", name: data['custname'] ?? "")]);
          String docDateTime = "";
          if (data['docdatetime'] != null) {
            // แปลงเป็น MongoDB datetime format (ISO 8601 with Z)
            String dateTime = data['docdatetime']; // เช่น "2025-08-06"
            if (dateTime.length == 10) {
              // ถ้ามีแค่วันที่ ให้เพิ่มเวลา default
              docDateTime = "${dateTime}T00:00:00.000Z";
            } else if (dateTime.length == 19) {
              // ถ้ามีวันที่และเวลา แต่ไม่มี Z ให้เพิ่ม Z
              docDateTime = "$dateTime.000Z";
            } else if (dateTime.length == 24 && !dateTime.endsWith('Z')) {
              // ถ้ามีวันที่ เวลา และ milliseconds แต่ไม่มี Z ให้เพิ่ม Z
              docDateTime = "${dateTime}Z";
            } else {
              docDateTime = dateTime; // มีครบแล้ว
            }
          } else {
            docDateTime = "error";
          }
          // Parse details จาก API response (ถ้ามี)
          List<TransactionDetailModel> details = [];
          if (data['details'] != null && data['details'] is List) {
            details = (data['details'] as List).map((detail) {
              return TransactionDetailModel(
                docdatetime: docDateTime, // ใช้ docDateTime จาก header
                itemguid: detail['itemguid'] ?? const Uuid().v4(),
                barcode: detail['barcode'] ?? "",
                itemcode: detail['itemcode'] ?? "",
                itemnames: [LanguageDataModel(code: "th", name: detail['itemname'] ?? "")],
                unitcode: detail['unitcode'] ?? "",
                qty: global.safeToDouble(detail['qty']),
                price: global.safeToDouble(detail['price']),
                discount: detail['discount']?.toString() ?? "",
                discountamount: global.safeToDouble(detail['discountamount']),
                sumofcost: global.safeToDouble(detail['sumofcost']),
                sumamount: global.safeToDouble(detail['sumamount']),
                remark: detail['remark'] ?? "",
                linenumber: global.safeToInt(detail['linenumber']),
                whcode: detail['whcode'] ?? "",
                whnames: [LanguageDataModel(code: "th", name: detail['whname'] ?? "")],
                locationcode: detail['locationcode'] ?? "",
                locationnames: [LanguageDataModel(code: "th", name: detail['locationname'] ?? "")],
                towhcode: detail['towhcode'] ?? "",
                tolocationcode: detail['tolocationcode'] ?? "",
                shelfcode: detail['shelfcode'] ?? "",
                totalqty: global.safeToDouble(detail['totalqty']),
                averagecost: global.safeToDouble(detail['averagecost']),
                totalvaluevat: global.safeToDouble(detail['totalvaluevat']),
                sumamountexcludevat: global.safeToDouble(detail['sumamountexcludevat']),
                priceexcludevat: global.safeToDouble(detail['priceexcludevat']),
                standvalue: global.safeToDouble(detail['standvalue']),
                dividevalue: global.safeToDouble(detail['dividevalue']),
                calcflag: global.safeToInt(detail['calcflag']),
                vattype: global.safeToInt(detail['vattype']),
                ispos: global.safeToInt(detail['ispos']),
                laststatus: global.safeToInt(detail['laststatus']),
                itemtype: global.safeToInt(detail['itemtype']),
                taxtype: global.safeToInt(detail['taxtype']),
                inquirytype: global.safeToInt(detail['inquirytype']),
                multiunit: detail['multiunit'] ?? false,
                unitnames: [LanguageDataModel(code: "th", name: detail['unitname'] ?? "")],
              );
            }).toList();
          }

          var dataTrans = TransactionModel(
            guidref: "",
            guidfixed: data['guidfixed'] ?? "",
            docrefno: "",
            docrefdate: "",
            docreftype: 0,
            docno: data['docno'] ?? "",
            docdatetime: docDateTime,
            doctype: 0,
            vattype: 0,
            custcode: data['custcode'] ?? "",
            custnames: custNames,
            salecode: "",
            salename: "",
            discountword: "",
            totalcost: 0.0,
            totalvalue: 0.0,
            totaldiscount: 0.0,
            totalvatvalue: 0.0,
            totalbeforevat: 0.0,
            totalaftervat: 0.0,
            totalexceptvat: 0.0,
            totalamount: global.safeToDouble(data['totalamount']),
            cashiercode: "",
            posid: "",
            membercode: "",
            vatrate: 0.0,
            status: 0,
            inquirytype: 0,
            taxdocno: "",
            taxdocdate: "",
            transflag: global.safeToInt(data['transflag']),
            isref: data['isref'] ?? false,
            islocked: data['islocked'] ?? false,
            isclosed: data['isclosed'] ?? false,
            iscancel: data['iscancel'] ?? false,
            ismanualamount: false,
            detailcount: global.safeToInt(data['detailcount']),
            // ข้อมูลผู้สร้างจาก /getdoc API (PostgreSQL) - modifier อยู่ใน datahistory แทน
            creatorcode: data['creator_code']?.toString() ?? "",
            creatorname: data['creator_name']?.toString() ?? "",
            createdat: data['created_at']?.toString() ?? "",
            details: details, // เพิ่ม details ที่ parse มาจาก API
          );
          // เพิ่ม currency fields สำหรับแสดง multi-currency บน list card
          dataTrans.exchangerate = global.safeToDouble(data['exchangerate']);
          dataTrans.totalamountDoc = global.safeToDouble(data['totalamount_doc']);
          dataTrans.docCurrency = data['doc_currency']?.toString();
          dataTrans.docCurrencySymbol = data['doc_currencysymbol']?.toString();
          dataTrans.isdelete = data['isdelete'] ?? false;
          dataTrans.iscomparedsuccess = global.safeToInt(data['iscomparedsuccess']);
          dataTrans.isclosedmanual = data['isclosedmanual'] ?? false;
          dataTrans.closedmanualByCode = data['closedmanual_by_code']?.toString();
          dataTrans.closedmanualByName = data['closedmanual_by_name']?.toString();
          dataTrans.closedmanualAt = data['closedmanual_at']?.toString();
          dataTrans.closedmanualReason = data['closedmanual_reason']?.toString();
          trans.add(dataTrans);
        }

        emit(TransLoadSuccess(totalRecord: goApiQueryResult.total, trans: trans));
      } else {
        // // print(results.data);
        if (results.success) {
          List<TransactionModel> trans = (results.data as List).map((trans) => TransactionModel.fromJson(trans)).toList();

          // Parse transaction data หลังจากแปลงเป็น TransactionModel แล้ว
          for (var transaction in trans) {
            global.parseTransactionJsonData(transaction);
          }

          if (event.type == global.TransactionTypeEnum.stocktransfer) {
            for (var tran in trans) {
              tran.details!.removeWhere((ele) => ele.calcflag == 1);
            }
          }
          emit(TransLoadSuccess(totalRecord: 0, trans: trans));
        } else {
          emit(const TransLoadFailed(message: 'Trans Not Found'));
        }
      }
    } catch (e) {
      emit(TransLoadFailed(message: e.toString()));
    }
  }

  void onTransSave(TransSave event, Emitter<TransState> emit) async {
    AppLogger.info('🟢 [TransBloc.onTransSave] START - type: ${event.type}');
    AppLogger.info('🟢 [TransBloc.onTransSave] docno: ${event.trans.docno}, guidfixed: ${event.trans.guidfixed}');
    emit(TransSaveInProgress());
    try {
      late ApiResponse<dynamic> results;
      if (event.type == global.TransactionTypeEnum.purchase) {
        results = await _transRepository.savePurchase(event.trans);
      } else if (event.type == global.TransactionTypeEnum.purchasereturn) {
        results = await _transRepository.savePurchaseReturn(event.trans);
      } else if (event.type == global.TransactionTypeEnum.sale) {
        results = await _transRepository.saveSale(event.trans);
      } else if (event.type == global.TransactionTypeEnum.salereturn) {
        results = await _transRepository.saveSaleReturn(event.trans);
      } else if (event.type == global.TransactionTypeEnum.adjust) {
        if (event.trans.transflag == 66 || event.trans.transflag == 866) {
          for (var data in event.trans.details!) {
            data.calcflag = 1;
          }
        } else if (event.trans.transflag == 68 || event.trans.transflag == 868) {
          for (var data in event.trans.details!) {
            data.calcflag = -1;
          }
        }
        results = await _transRepository.saveAdjust(event.trans);
      } else if (event.type == global.TransactionTypeEnum.stockpickupproduct) {
        results = await _transRepository.saveStockPickup(event.trans);
      } else if (event.type == global.TransactionTypeEnum.stockreceiveproduct) {
        results = await _transRepository.saveStockReceive(event.trans);
      } else if (event.type == global.TransactionTypeEnum.stockreturnproduct) {
        results = await _transRepository.saveStockReturn(event.trans);
      } else if (event.type == global.TransactionTypeEnum.stocktransfer) {
        event.trans.transflag = 72;
        List<TransactionDetailModel> detailsOut = [];
        for (var data in event.trans.details!) {
          data.calcflag = -1;
          TransactionDetailModel newData = TransactionDetailModel(
            docdatetime: data.docdatetime,
            docrefdatetime: data.docrefdatetime,
            itemguid: data.itemguid,
            barcode: data.barcode,
            itemcode: data.itemcode,
            itemnames: data.itemnames,
            unitcode: data.unitcode,
            unitnames: data.unitnames,
            qty: data.qty,
            price: data.price,
            discount: data.discount,
            sumofcost: data.sumofcost,
            remark: data.remark,
            linenumber: data.linenumber,
            whcode: data.whcode,
            shelfcode: data.shelfcode,
            sumamount: data.sumamount,
            locationcode: data.locationcode,
            totalvaluevat: data.totalvaluevat,
            totalqty: data.totalqty,
            sumamountexcludevat: data.sumamountexcludevat,
            priceexcludevat: data.priceexcludevat,
            discountamount: data.discountamount,
            standvalue: data.standvalue,
            dividevalue: data.dividevalue,
            calcflag: 1,
            vattype: data.vattype,
            averagecost: data.averagecost,
            ispos: data.ispos,
            laststatus: data.laststatus,
            itemtype: data.itemtype,
            taxtype: data.taxtype,
            inquirytype: data.inquirytype,
            multiunit: data.multiunit,
          );

          newData.whcode = data.towhcode ?? '';
          newData.whnames = data.towhnames ?? [];
          newData.locationcode = data.tolocationcode ?? '';
          newData.locationnames = data.tolocationnames ?? [];

          detailsOut.add(newData);
        }
        event.trans.details!.addAll(detailsOut);
        results = await _transRepository.saveTransfer(event.trans);
      } else if (event.type == global.TransactionTypeEnum.quotation) {
        results = await _transRepository.saveQuotation(event.trans);
      } else if (event.type == global.TransactionTypeEnum.saleorder) {
        results = await _transRepository.saveSaleOrder(event.trans);
      } else if (event.type == global.TransactionTypeEnum.purchaseorder) {
        AppLogger.info('🟢 [TransBloc] Calling savePurchaseOrder...');
        results = await _transRepository.savePurchaseOrder(event.trans);
        AppLogger.info('🟢 [TransBloc] savePurchaseOrder result: success=${results.success}, data=${results.data}, message=${results.message}');
      } else if (event.type == global.TransactionTypeEnum.purchasepartial) {
        results = await _transRepository.savePurchasePartial(event.trans);
      } else if (event.type == global.TransactionTypeEnum.accrualreceive) {
        results = await _transRepository.saveAccrualReceive(event.trans);
      } else if (event.type == global.TransactionTypeEnum.advancePayment) {
        results = await _transRepository.saveAdvancePayment(event.trans);
      } else if (event.type == global.TransactionTypeEnum.advancePaymentRefund) {
        results = await _transRepository.saveAdvancePaymentRefund(event.trans);
      } else if (event.type == global.TransactionTypeEnum.deposit) {
        results = await _transRepository.saveDeposit(event.trans);
      } else if (event.type == global.TransactionTypeEnum.depositRefund) {
        results = await _transRepository.saveDepositRefund(event.trans);
      } else if (event.type == global.TransactionTypeEnum.paidAdvance) {
        results = await _transRepository.savePaidAdvance(event.trans);
      } else if (event.type == global.TransactionTypeEnum.paidAdvanceRefund) {
        results = await _transRepository.savePaidAdvanceRefund(event.trans);
      } else if (event.type == global.TransactionTypeEnum.receiveDeposit) {
        results = await _transRepository.saveReceiveDeposit(event.trans);
      } else if (event.type == global.TransactionTypeEnum.receiveDepositRefund) {
        results = await _transRepository.saveReceiveDepositRefund(event.trans);
      } else {
        emit(const TransSaveFailed(message: "No type"));
      }
      if (results.success) {
        emit(TransSaveSuccess(docno: results.data));
      } else {
        emit(TransSaveFailed(message: results.message));
      }
    } catch (e) {
      emit(TransSaveFailed(message: e.toString()));
    }
  }

  void onTransDelete(TransDelete event, Emitter<TransState> emit) async {
    emit(TransDeleteInProgress());
    try {
      if (event.type == global.TransactionTypeEnum.purchase) {
        await _transRepository.deletePurchase(event.guid);
        emit(TransDeleteSuccess());
      } else if (event.type == global.TransactionTypeEnum.purchasereturn) {
        await _transRepository.deletePurchaseReturn(event.guid);
        emit(TransDeleteSuccess());
      } else if (event.type == global.TransactionTypeEnum.sale) {
        await _transRepository.deleteSale(event.guid);
        emit(TransDeleteSuccess());
      } else if (event.type == global.TransactionTypeEnum.salereturn) {
        await _transRepository.deleteSaleReturn(event.guid);
        emit(TransDeleteSuccess());
      } else if (event.type == global.TransactionTypeEnum.adjust) {
        await _transRepository.deleteAdjust(event.guid);
        emit(TransDeleteSuccess());
      } else if (event.type == global.TransactionTypeEnum.stockpickupproduct) {
        await _transRepository.deleteStockPickup(event.guid);
        emit(TransDeleteSuccess());
      } else if (event.type == global.TransactionTypeEnum.stockreceiveproduct) {
        await _transRepository.deleteReceive(event.guid);
        emit(TransDeleteSuccess());
      } else if (event.type == global.TransactionTypeEnum.stockreturnproduct) {
        await _transRepository.deleteReturn(event.guid);
        emit(TransDeleteSuccess());
      } else if (event.type == global.TransactionTypeEnum.stocktransfer) {
        await _transRepository.deleteTransfer(event.guid);
        emit(TransDeleteSuccess());
      } else if (event.type == global.TransactionTypeEnum.stockbalance) {
        await _transRepository.deleteStockBalance(event.guid);
        emit(TransDeleteSuccess());
      } else if (event.type == global.TransactionTypeEnum.saleorder) {
        await _transRepository.deleteSaleOrder(event.guid);
        emit(TransDeleteSuccess());
      } else if (event.type == global.TransactionTypeEnum.purchaseorder) {
        await _transRepository.deletePurchaseOrder(event.guid);
        emit(TransDeleteSuccess());
      } else if (event.type == global.TransactionTypeEnum.quotation) {
        await _transRepository.deleteQuotation(event.guid);
        emit(TransDeleteSuccess());
      } else if (event.type == global.TransactionTypeEnum.purchasepartial) {
        await _transRepository.deletePurchasePartial(event.guid);
        emit(TransDeleteSuccess());
      } else if (event.type == global.TransactionTypeEnum.accrualreceive) {
        await _transRepository.deleteAccrualReceive(event.guid);
        emit(TransDeleteSuccess());
      } else if (event.type == global.TransactionTypeEnum.advancePayment) {
        await _transRepository.deleteAdvancePayment(event.guid);
        emit(TransDeleteSuccess());
      } else if (event.type == global.TransactionTypeEnum.advancePaymentRefund) {
        await _transRepository.deleteAdvancePaymentRefund(event.guid);
        emit(TransDeleteSuccess());
      } else if (event.type == global.TransactionTypeEnum.deposit) {
        await _transRepository.deleteDeposit(event.guid);
        emit(TransDeleteSuccess());
      } else if (event.type == global.TransactionTypeEnum.depositRefund) {
        await _transRepository.deleteDepositRefund(event.guid);
        emit(TransDeleteSuccess());
      } else if (event.type == global.TransactionTypeEnum.paidAdvance) {
        await _transRepository.deletePaidAdvance(event.guid);
        emit(TransDeleteSuccess());
      } else if (event.type == global.TransactionTypeEnum.paidAdvanceRefund) {
        await _transRepository.deletePaidAdvanceRefund(event.guid);
        emit(TransDeleteSuccess());
      } else if (event.type == global.TransactionTypeEnum.receiveDeposit) {
        await _transRepository.deleteReceiveDeposit(event.guid);
        emit(TransDeleteSuccess());
      } else if (event.type == global.TransactionTypeEnum.receiveDepositRefund) {
        await _transRepository.deleteReceiveDepositRefund(event.guid);
        emit(TransDeleteSuccess());
      } else {
        emit(const TransDeleteFailed(message: "No type"));
      }
    } catch (e) {
      emit(TransDeleteFailed(message: e.toString()));
    }
  }

  void onTransUpdate(TransUpdate event, Emitter<TransState> emit) async {
    emit(TransUpdateInProgress());
    try {
      if (event.type == global.TransactionTypeEnum.purchase) {
        await _transRepository.updatePurchase(event.guid, event.trans);
        emit(TransUpdateSuccess());
      } else if (event.type == global.TransactionTypeEnum.purchasereturn) {
        await _transRepository.updatePurchaseReturn(event.guid, event.trans);
        emit(TransUpdateSuccess());
      } else if (event.type == global.TransactionTypeEnum.sale) {
        await _transRepository.updateSale(event.guid, event.trans);
        emit(TransUpdateSuccess());
      } else if (event.type == global.TransactionTypeEnum.salereturn) {
        await _transRepository.updateSaleReturn(event.guid, event.trans);
        emit(TransUpdateSuccess());
      } else if (event.type == global.TransactionTypeEnum.adjust) {
        if (event.trans.transflag == 66 || event.trans.transflag == 866) {
          for (var data in event.trans.details!) {
            data.calcflag = 1;
          }
        } else if (event.trans.transflag == 68 || event.trans.transflag == 868) {
          for (var data in event.trans.details!) {
            data.calcflag = -1;
          }
        }
        await _transRepository.updateAdjust(event.guid, event.trans);
        emit(TransUpdateSuccess());
      } else if (event.type == global.TransactionTypeEnum.stockpickupproduct) {
        await _transRepository.updateStockPickup(event.guid, event.trans);
        emit(TransUpdateSuccess());
      } else if (event.type == global.TransactionTypeEnum.stockreceiveproduct) {
        await _transRepository.updateStockReceive(event.guid, event.trans);
        emit(TransUpdateSuccess());
      } else if (event.type == global.TransactionTypeEnum.stockreturnproduct) {
        await _transRepository.updateStockReturn(event.guid, event.trans);
        emit(TransUpdateSuccess());
      } else if (event.type == global.TransactionTypeEnum.stocktransfer) {
        event.trans.transflag = 72;
        List<TransactionDetailModel> detailsOut = [];
        for (var data in event.trans.details!) {
          data.calcflag = -1;
          TransactionDetailModel newData = TransactionDetailModel(
            docdatetime: data.docdatetime,
            docrefdatetime: data.docrefdatetime,
            itemguid: data.itemguid,
            barcode: data.barcode,
            itemcode: data.itemcode,
            itemnames: data.itemnames,
            unitcode: data.unitcode,
            unitnames: data.unitnames,
            qty: data.qty,
            price: data.price,
            discount: data.discount,
            sumofcost: data.sumofcost,
            remark: data.remark,
            linenumber: data.linenumber,
            whcode: data.whcode,
            shelfcode: data.shelfcode,
            sumamount: data.sumamount,
            locationcode: data.locationcode,
            totalvaluevat: data.totalvaluevat,
            totalqty: data.totalqty,
            sumamountexcludevat: data.sumamountexcludevat,
            priceexcludevat: data.priceexcludevat,
            discountamount: data.discountamount,
            standvalue: data.standvalue,
            dividevalue: data.dividevalue,
            calcflag: 1,
            vattype: data.vattype,
            averagecost: data.averagecost,
            ispos: data.ispos,
            laststatus: data.laststatus,
            itemtype: data.itemtype,
            taxtype: data.taxtype,
            inquirytype: data.inquirytype,
            multiunit: data.multiunit,
          );

          newData.whcode = data.towhcode ?? '';
          newData.whnames = data.towhnames ?? [];
          newData.locationcode = data.tolocationcode ?? '';
          newData.locationnames = data.tolocationnames ?? [];

          detailsOut.add(newData);
        }
        event.trans.details!.addAll(detailsOut);

        await _transRepository.updateTransfer(event.guid, event.trans);
        emit(TransUpdateSuccess());
      } else if (event.type == global.TransactionTypeEnum.quotation) {
        await _transRepository.updateQuotation(event.guid, event.trans);
        emit(TransUpdateSuccess());
      } else if (event.type == global.TransactionTypeEnum.saleorder) {
        await _transRepository.updateSaleOrder(event.guid, event.trans);
        emit(TransUpdateSuccess());
      } else if (event.type == global.TransactionTypeEnum.purchaseorder) {
        await _transRepository.updatePurchaseOrder(event.guid, event.trans);
        emit(TransUpdateSuccess());
      } else if (event.type == global.TransactionTypeEnum.purchasepartial) {
        await _transRepository.updatePurchasePartial(event.guid, event.trans);
        emit(TransUpdateSuccess());
      } else if (event.type == global.TransactionTypeEnum.accrualreceive) {
        await _transRepository.updateAccrualReceive(event.guid, event.trans);
        emit(TransUpdateSuccess());
      } else if (event.type == global.TransactionTypeEnum.advancePayment) {
        await _transRepository.updateAdvancePayment(event.guid, event.trans);
        emit(TransUpdateSuccess());
      } else if (event.type == global.TransactionTypeEnum.advancePaymentRefund) {
        await _transRepository.updateAdvancePaymentRefund(event.guid, event.trans);
        emit(TransUpdateSuccess());
      } else if (event.type == global.TransactionTypeEnum.deposit) {
        await _transRepository.updateDeposit(event.guid, event.trans);
        emit(TransUpdateSuccess());
      } else if (event.type == global.TransactionTypeEnum.depositRefund) {
        await _transRepository.updateDepositRefund(event.guid, event.trans);
        emit(TransUpdateSuccess());
      } else if (event.type == global.TransactionTypeEnum.paidAdvance) {
        await _transRepository.updatePaidAdvance(event.guid, event.trans);
        emit(TransUpdateSuccess());
      } else if (event.type == global.TransactionTypeEnum.paidAdvanceRefund) {
        await _transRepository.updatePaidAdvanceRefund(event.guid, event.trans);
        emit(TransUpdateSuccess());
      } else if (event.type == global.TransactionTypeEnum.receiveDeposit) {
        await _transRepository.updateReceiveDeposit(event.guid, event.trans);
        emit(TransUpdateSuccess());
      } else if (event.type == global.TransactionTypeEnum.receiveDepositRefund) {
        await _transRepository.updateReceiveDepositRefund(event.guid, event.trans);
        emit(TransUpdateSuccess());
      } else {
        emit(const TransUpdateFailed(message: "No type"));
      }
    } catch (e) {
      emit(TransUpdateFailed(message: e.toString()));
    }
  }

  void onTransCreateFullInvoice(TransCreateFullInvoice event, Emitter<TransState> emit) async {
    emit(TransFullInvoiceInProgress());

    try {
      // ตรวจสอบข้อมูลก่อนดำเนินการ
      if (event.trans.guidfixed?.isEmpty ?? true) {
        emit(TransFullInvoiceFailed(message: 'ไม่พบข้อมูลเอกสาร'));
        return;
      }

      // แก้ไขเฉพาะส่วนการยกเลิก
      event.trans.iscancel = true;
      event.trans.cancelreason = 'ใบกำกับภาษีแบบเต็มออกแทน';
      event.trans.cancelusercode = global.profileData.username;
      event.trans.cancelusername = global.profileData.name;
      event.trans.canceldatetime = DateTime.now().toLocal().toIso8601String();
      event.trans.canceltime = DateTime.now().toLocal().toIso8601String();

      // Update เอกสารเดิมให้เป็น cancelled
      late ApiResponse<dynamic> updateResult;
      if (event.type == global.TransactionTypeEnum.sale) {
        updateResult = await _transRepository.updateSale(event.guid, event.trans);
      }

      // ตรวจสอบการ update สำเร็จ
      if (!updateResult.success) {
        emit(TransFullInvoiceFailed(message: 'Failed to cancel original document: ${updateResult.message}'));
        return;
      }

      // Step 2: สร้างเอกสารใหม่สำหรับใบกำกับภาษีแบบเต็ม - คงข้อมูลเดิมทุกส่วน
      await Future.delayed(const Duration(milliseconds: 500));

      // แก้ไขเฉพาะส่วนที่จำเป็นสำหรับใบกำกับภาษีแบบเต็ม
      event.trans.guidfixed = ''; // Clear guid เพื่อสร้างใหม่
      event.trans.taxdocno = event.trans.docno;
      event.trans.taxdocdate = event.trans.docdatetime;
      event.trans.docno = ''; // mainapi จะสร้าง docno เอง
      // อัพเดท timezone info สำหรับสร้าง docno ใหม่
      final timezoneInfo = global.getLocalTimezoneInfo();
      event.trans.docdatelocal = timezoneInfo['docdatelocal'];
      event.trans.doctimelocal = timezoneInfo['doctimelocal'];
      event.trans.timezone = timezoneInfo['timezone'];
      // Reset cancel status สำหรับเอกสารใหม่
      event.trans.iscancel = false;
      event.trans.cancelreason = null;
      event.trans.cancelusercode = null;
      event.trans.cancelusername = null;
      event.trans.canceldatetime = null;
      event.trans.canceltime = null;

      late ApiResponse<dynamic> saveResult;
      if (event.type == global.TransactionTypeEnum.sale) {
        saveResult = await _transRepository.saveSale(event.trans);
      }

      if (saveResult.success) {
        emit(TransFullInvoiceSuccess(docno: saveResult.data));
      } else {
        emit(TransFullInvoiceFailed(message: 'Failed to create full tax invoice: ${saveResult.message}'));
      }
    } catch (e) {
      emit(TransFullInvoiceFailed(message: e.toString()));
    }
  }
}
