import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/bi_report/payment_daily_model.dart';
import 'package:smlaicloud/model/bi_report/purchase_partial_model.dart';
import 'package:smlaicloud/model/bi_report/sale_return_model.dart';
import 'package:smlaicloud/model/bi_report/stock_balance_model.dart';
import 'package:smlaicloud/screens/report/dedebi/components/payment_daily_table_view.dart';
import 'package:smlaicloud/screens/report/dedebi/components/purchase_partial_table_view.dart';
import 'package:smlaicloud/screens/report/dedebi/components/sale_return_table_view.dart';
import 'package:smlaicloud/screens/report/dedebi/components/stock_balance_table_view.dart';
import '../../../../model/bi_report/bi_report_models.dart';
import '../../../../model/bi_report/bi_sale_report_data.dart';
import '../../../../model/bi_report/sale_daily_report_model.dart';
import '../../../../model/bi_report/stock_movment_model.dart';
import 'sale_report_table_view.dart';
import 'sale_daily_report_table_view.dart';
import 'stock_movement_table_view.dart';

class SaleTableView extends StatelessWidget {
  final List<SaleReportData>? dataSale;
  final List<SaleDailyReportData>? dataSaleDaily;
  final List<StockMovementModel>? dataStockMovement;
  final List<PaymentDailyModel>? dataPaymentDaily;
  final List<SaleReturnModel>? dataSaleReturn;
  final List<StockBalanceModel>? dataStockBalance;
  final List<PurchasePartialModel>? dataPurchasePartial;
  final String Function(double) formatCurrency;
  final String Function(List<SaleCreditorName>) getCreditorName;
  final Function(SaleReportData)? onRowTap;
  final Function(SaleDailyReportData)? onRowSaleDailyTap;
  final Function(StockMovementModel)? onRowStockMovementTap;
  final Function(SaleReturnModel)? onRowSaleReturnTap;
  final Function(PaymentDailyModel)? onRowPaymentDailyTap;
  final Function(PurchasePartialModel)? onRowPurchasePartialTap;
  final BiReportType reportType;

  const SaleTableView({
    super.key,
    this.dataSale,
    this.dataSaleDaily,
    this.dataStockMovement,
    this.dataPaymentDaily,
    this.dataSaleReturn,
    this.dataStockBalance,
    this.dataPurchasePartial,
    required this.formatCurrency,
    required this.getCreditorName,
    this.onRowTap,
    this.onRowSaleDailyTap,
    this.onRowStockMovementTap,
    this.onRowPaymentDailyTap,
    this.onRowSaleReturnTap,
    this.onRowPurchasePartialTap,
    required this.reportType,
  });

  @override
  Widget build(BuildContext context) {
    // Use appropriate table component based on report type
    if (reportType == BiReportType.sale) {
      return SaleReportTableView(
        data: dataSale!,
        getCreditorName: getCreditorName,
        onRowTap: onRowTap,
      );
    } else if (reportType == BiReportType.saleDaily) {
      return SaleDailyReportTableView(
        data: dataSaleDaily!,
        onRowTap: onRowSaleDailyTap,
      );
    } else if (reportType == BiReportType.stockMovement) {
      return StockMovementTableView(
        data: dataStockMovement!,
        onRowTap: onRowStockMovementTap,
      );
    } else if (reportType == BiReportType.paymentDaily) {
      return PaymentDailyTableView(
        data: dataPaymentDaily!,
        onRowTap: onRowPaymentDailyTap,
      );
    } else if (reportType == BiReportType.saleReturn) {
      return SaleReturnTableView(
        data: dataSaleReturn!,
        onRowTap: onRowSaleReturnTap,
      );
    } else if (reportType == BiReportType.stockBalance) {
      return StockBalanceTableView(
        stockBalances: dataStockBalance!,
        formatCurrency: formatCurrency,
      );
    } else if (reportType == BiReportType.purchasepartial) {
      return PurchasePartialTableView(
        data: dataPurchasePartial!,
        formatCurrency: formatCurrency,
        onRowTap: onRowPurchasePartialTap,
      );
    } else {
      return Center(
        child: Text(
          'ไม่ได้กำหนด ตาราง สำหรับ ประเภทรายงาน $reportType',
          style: TextStyle(fontSize: 16, color: global.theme.textSecondaryColor),
        ),
      );
    }
  }
}
