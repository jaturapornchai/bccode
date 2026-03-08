import 'package:flutter/material.dart';
import 'package:smlaicloud/bloc/export_csv/export_csv_bloc.dart';
import 'package:smlaicloud/imports_bloc.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/model/warehouse_model.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_ui_utils.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/global.dart' as global;

/// Handler สำหรับจัดการ Bloc Side Effects ในหน้าจอ PO
///
/// รับผิดชอบ:
/// - จัดการ TransBloc states (Save, Update, Delete, Get)
/// - จัดการ WarehouseBloc states
/// - จัดการ ExportCsvBloc states
class POBlocEffectHandler {
  final bool Function() isMounted;
  final VoidCallback onStateChanged;
  final void Function(bool isLoading) setLoading;
  final TransactionModel Function() getScreenData;
  final void Function(TransactionModel) setScreenData;
  final void Function(TransactionModel) setScreenDataTemp;
  final VoidCallback onClearScreen;
  final void Function(String docNo) onShowSaveDialog;
  final void Function(String docNo, global.TransactionTypeEnum transType) onSubmitPOAndHandlePrint;
  final void Function(global.TransactionTypeEnum transType) onSubmitPOAndHandleUpdatePrint;
  final void Function(String docNo, global.TransactionTypeEnum transType) onHandleSaveSuccessWithPrintCheck;
  final void Function(global.TransactionTypeEnum transType) onHandleUpdateSuccessWithPrintCheck;
  final VoidCallback onRefreshListAfterSave;
  final void Function(String deletedGuid) onRefreshListAfterDelete;
  final void Function(String docNo) onLoadPurchaseTypesAndExisting;
  final VoidCallback onLoadCurrencies;
  final void Function(String docNo) onLoadPOApprovalStatus;
  final VoidCallback onShowDuplicateDocNoDialog;

  // Warehouse callbacks
  final void Function(List<WarehouseModel>) onWarehouseListLoaded;
  final void Function({
    required String defaultWarehouse,
    required List<LanguageDataModel> defaultWarehouseNames,
    required String defaultLocation,
    required List<LanguageDataModel> defaultLocationNames,
    required String defaultToWarehouse,
    required List<LanguageDataModel> defaultToWarehouseNames,
    required String defaultToLocation,
    required List<LanguageDataModel> defaultToLocationNames,
  }) onSetDefaultWarehouse;

  final global.TransactionTypeEnum transactionType;
  final bool showDocSearchPanel;

  POBlocEffectHandler({
    required this.isMounted,
    required this.onStateChanged,
    required this.setLoading,
    required this.getScreenData,
    required this.setScreenData,
    required this.setScreenDataTemp,
    required this.onClearScreen,
    required this.onShowSaveDialog,
    required this.onSubmitPOAndHandlePrint,
    required this.onSubmitPOAndHandleUpdatePrint,
    required this.onHandleSaveSuccessWithPrintCheck,
    required this.onHandleUpdateSuccessWithPrintCheck,
    required this.onRefreshListAfterSave,
    required this.onRefreshListAfterDelete,
    required this.onLoadPurchaseTypesAndExisting,
    required this.onLoadCurrencies,
    required this.onLoadPOApprovalStatus,
    required this.onShowDuplicateDocNoDialog,
    required this.onWarehouseListLoaded,
    required this.onSetDefaultWarehouse,
    required this.transactionType,
    required this.showDocSearchPanel,
  });

  /// จัดการ state จาก TransBloc
  void handleTransState(BuildContext context, TransState state) {
    // จัดการ Loading States
    if (state is TransInProgress ||
        state is TransSaveInProgress ||
        state is TransUpdateInProgress ||
        state is TransDeleteInProgress ||
        state is TransDeleteManyInProgress ||
        state is TransGetInProgress ||
        state is TransFullInvoiceInProgress) {
      if (state is TransSaveInProgress) {
        AppLogger.debug('🖨️ [TransSaveInProgress] กำลังบันทึก...');
      }
      if (isMounted()) setLoading(true);
      return;
    }

    // จัดการ Success/Failed States
    if (state is TransLoadSuccess) {
      if (isMounted()) setLoading(false);
    } else if (state is TransLoadFailed) {
      if (isMounted()) setLoading(false);
      context.ui.showError(state.message);
    } else if (state is TransSaveSuccess) {
      _handleTransSaveSuccess(context, state);
    } else if (state is TransSaveFailed) {
      _handleTransSaveFailed(context, state);
    } else if (state is TransUpdateSuccess) {
      _handleTransUpdateSuccess(context, state);
    } else if (state is TransUpdateFailed) {
      if (isMounted()) setLoading(false);
      context.ui.showError(state.message);
    } else if (state is TransDeleteSuccess) {
      _handleTransDeleteSuccess(context);
    } else if (state is TransDeleteFailed) {
      if (isMounted()) setLoading(false);
      context.ui.showError(state.message);
    } else if (state is TransDeleteManySuccess) {
      if (isMounted()) setLoading(false);
      context.ui.showSuccess(global.language('delete_multiple_success'));
    } else if (state is TransDeleteManyFailed) {
      if (isMounted()) setLoading(false);
      context.ui.showError(global.language('delete_multiple_failed'));
    } else if (state is TransGetSuccess) {
      _handleTransGetSuccess(context, state);
    } else if (state is TransGetFailed) {
      if (isMounted()) setLoading(false);
      context.ui.showError(state.message);
    } else if (state is TransFullInvoiceSuccess) {
      if (isMounted()) setLoading(false);
      onShowSaveDialog(state.docno);
      context.ui.showSuccess(global.language('create_full_tax_invoice_success'));
    } else if (state is TransFullInvoiceFailed) {
      if (isMounted()) setLoading(false);
      context.ui.showError('${global.language("create_full_tax_invoice_failed")}: ${state.message}');
    } else {
      // Reset loading สำหรับ states อื่นๆ
      if (isMounted()) setLoading(false);
    }
  }

  /// จัดการ TransSaveSuccess
  void _handleTransSaveSuccess(BuildContext context, TransSaveSuccess state) {
    AppLogger.debug('🖨️ [TransSaveSuccess] ENTERED - docno: ${state.docno}');
    setLoading(false);

    final screenData = getScreenData();
    screenData.docno = state.docno;
    setScreenData(screenData);

    // ส่ง PO เข้าระบบอนุมัติ + จัดการพิมพ์ (ทั้ง auto-approved และ manual)
    // submitAndHandlePrint จัดการทั้ง 2 กรณีภายในแล้ว:
    // - auto_approved → เปิด PDF อัตโนมัติ
    // - ไม่ auto → เรียก handleSaveSuccessWithPrintCheck (แสดง dialog/พิมพ์ตาม prefs)
    onSubmitPOAndHandlePrint(state.docno, transactionType);

    if (isMounted()) onStateChanged();

    // Refresh list ด้านซ้าย
    onRefreshListAfterSave();
  }

  /// จัดการ TransSaveFailed
  void _handleTransSaveFailed(BuildContext context, TransSaveFailed state) {
    if (isMounted()) setLoading(false);

    // ตรวจสอบ duplicate key error
    final isDuplicateKeyError = state.message.toLowerCase().contains('duplicate key') ||
        state.message.contains('E11000') ||
        state.message.toLowerCase().contains('docno') && state.message.toLowerCase().contains(global.language('duplicate_short'));

    if (isDuplicateKeyError) {
      onShowDuplicateDocNoDialog();
    } else {
      context.ui.showError(state.message);
    }
  }

  /// จัดการ TransUpdateSuccess
  void _handleTransUpdateSuccess(BuildContext context, TransUpdateSuccess state) {
    AppLogger.debug('🖨️ [TransUpdateSuccess] ENTERED');
    if (isMounted()) setLoading(false);

    // ส่ง PO เข้าระบบอนุมัติ
    onSubmitPOAndHandleUpdatePrint(transactionType);

    // Refresh list ด้านซ้าย
    onRefreshListAfterSave();
  }

  /// จัดการ TransDeleteSuccess
  void _handleTransDeleteSuccess(BuildContext context) {
    final screenData = getScreenData();
    String deletedGuid = screenData.guidfixed ?? '';
    onClearScreen();
    if (isMounted()) setLoading(false);
    context.ui.showSuccess(global.language("delete_success"));

    // Refresh list ด้านซ้าย
    if (showDocSearchPanel && deletedGuid.isNotEmpty) {
      Future.delayed(const Duration(milliseconds: 2000), () {
        if (isMounted()) {
          onRefreshListAfterDelete(deletedGuid);
        }
      });
    }
  }

  /// จัดการ TransGetSuccess
  void _handleTransGetSuccess(BuildContext context, TransGetSuccess state) {
    if (isMounted()) setLoading(false);

    setScreenData(state.trans);
    setScreenDataTemp(TransactionModel.fromJson(state.trans.toJson()));

    final screenData = getScreenData();
    AppLogger.info('[PO] TransGetSuccess - docCurrency: ${screenData.docCurrency}, '
        'docCurrencySymbol: ${screenData.docCurrencySymbol}, '
        'currency: ${screenData.currency}, '
        'exchangerate: ${screenData.exchangerate}');

    if (screenData.docno.isNotEmpty) {
      onLoadPurchaseTypesAndExisting(screenData.docno);
      onLoadCurrencies();
      onLoadPOApprovalStatus(screenData.docno);
    }
  }

  /// จัดการ state จาก WarehouseBloc
  void handleWarehouseState(BuildContext context, WarehouseState state) {
    if (state is WarehouseLoadSuccess) {
      onWarehouseListLoaded(state.warehouses);
      if (isMounted() && state.warehouses.isNotEmpty) {
        final warehouse = state.warehouses[0];
        String defaultLocation = '';
        List<LanguageDataModel> defaultLocationNames = [];

        if (warehouse.location.isNotEmpty) {
          defaultLocation = warehouse.location[0].code;
          defaultLocationNames = warehouse.location[0].names;
        }

        onSetDefaultWarehouse(
          defaultWarehouse: warehouse.code,
          defaultWarehouseNames: warehouse.names,
          defaultLocation: defaultLocation,
          defaultLocationNames: defaultLocationNames,
          defaultToWarehouse: warehouse.code,
          defaultToWarehouseNames: warehouse.names,
          defaultToLocation: defaultLocation,
          defaultToLocationNames: defaultLocationNames,
        );
      }
    }
  }

  /// จัดการ state จาก ExportCsvBloc
  void handleExportCsvState(BuildContext context, ExportCsvState state) {
    if (state is SaleInvoiceExportSuccess) {
      global.showSnackBar(context, Icon(Icons.save, color: Colors.white), global.language("export_success"), Colors.blue);
    } else if (state is SaleInvoiceExportFailed) {
      global.showSnackBar(context, Icon(Icons.save, color: Colors.white), "${global.language("not_export_success")} : ${state.message}", Colors.red);
    }
  }
}
