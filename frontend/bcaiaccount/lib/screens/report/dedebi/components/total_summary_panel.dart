import 'package:flutter/material.dart';
import 'package:smlaicloud/model/bi_report/bi_report_models.dart';
import 'package:smlaicloud/model/bi_report/gross_profit_by_document_summary_model.dart';
import 'package:smlaicloud/model/bi_report/gross_profit_by_product_summary_model.dart';
import 'package:smlaicloud/model/bi_report/vat_sale_summary_model.dart';
import 'package:smlaicloud/model/bi_report/vat_buy_summary_model.dart';
import 'package:smlaicloud/model/bi_report/payment_daily_model.dart';
import 'package:smlaicloud/model/bi_report/purchase_partial_summary_model.dart';
import 'package:smlaicloud/model/bi_report/sale_daily_report_summary.dart';
import 'package:smlaicloud/model/bi_report/sale_report_summary.dart';
import 'package:smlaicloud/model/bi_report/sale_return_model.dart';
import 'package:smlaicloud/model/bi_report/stock_balance_model.dart';
import 'package:smlaicloud/model/bi_report/stock_movment_summary_model.dart';
import 'package:smlaicloud/screens/report/dedebi/utils/report_utils.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/global.dart' as global;

class TotalSummaryPanel extends StatelessWidget {
  final SaleReportSummary? totalSummary;
  final SaleDailyReportSummary? dailySummary;
  final StockMovmentSummaryModel? stockSummary;
  final PaymentDailySummaryModel? paymentSummary;
  final SaleReturnSummaryModel? saleReturnSummary;
  final StockBalanceSummaryModel? stockBalanceSummary;
  final PurchasePartialSummaryModel? purchasePartialSummary;
  final GrossProfitByDocumentSummary? grossProfitSummary;
  final GrossProfitByProductSummary? grossProfitByProductSummary;
  final VatSaleSummary? vatSaleSummary;
  final VatBuySummary? vatBuySummary;
  final bool isLoading;
  final BiReportType reportType;

  const TotalSummaryPanel({
    super.key,
    this.totalSummary,
    this.isLoading = false,
    this.dailySummary,
    this.stockSummary,
    this.paymentSummary,
    this.saleReturnSummary,
    this.stockBalanceSummary,
    this.purchasePartialSummary,
    this.grossProfitSummary,
    this.grossProfitByProductSummary,
    this.vatSaleSummary,
    this.vatBuySummary,
    required this.reportType,
  });

  @override
  Widget build(BuildContext context) {
    if (isLoading) {
      return _buildLoadingWidget();
    }

    // แยกการตรวจสอบตาม reportType
    if (reportType == BiReportType.sale) {
      if (totalSummary == null) {
        return _buildEmptyWidget();
      }
    } else if (reportType == BiReportType.saleDaily) {
      if (dailySummary == null) {
        return _buildEmptyWidget();
      }
    } else if (reportType == BiReportType.stockMovement) {
      if (stockSummary == null) {
        return _buildEmptyWidget();
      }
    } else if (reportType == BiReportType.paymentDaily) {
      if (paymentSummary == null) {
        return _buildEmptyWidget();
      }
    } else if (reportType == BiReportType.saleReturn) {
      if (saleReturnSummary == null) {
        return _buildEmptyWidget();
      }
    } else if (reportType == BiReportType.stockBalance) {
      if (stockBalanceSummary == null) {
        return _buildEmptyWidget();
      }
    } else if (reportType == BiReportType.purchasepartial) {
      AppLogger.debug("🎯 TotalSummaryPanel: checking purchasepartial");
      AppLogger.debug("📊 purchasePartialSummary: $purchasePartialSummary");
      if (purchasePartialSummary == null) {
        AppLogger.debug(
          "❌ purchasePartialSummary is null, showing empty widget",
        );
        return _buildEmptyWidget();
      }
      AppLogger.debug("✅ purchasePartialSummary is not null, showing content");
    } else if (reportType == BiReportType.grossProfitByDocument) {
      if (grossProfitSummary == null) {
        return _buildEmptyWidget();
      }
    } else if (reportType == BiReportType.grossProfitByProduct) {
      if (grossProfitByProductSummary == null) {
        return _buildEmptyWidget();
      }
    } else if (reportType == BiReportType.vatSale) {
      if (vatSaleSummary == null) {
        return _buildEmptyWidget();
      }
    } else if (reportType == BiReportType.vatBuy) {
      if (vatBuySummary == null) {
        return _buildEmptyWidget();
      }
    } else {
      return _buildEmptyWidget();
    }

    return _buildSummaryContent();
  }

  Widget _buildLoadingWidget() {
    return Container(
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: global.theme.dividerBorderColor),
        boxShadow: [
          BoxShadow(
            color: global.theme.textColor.withValues(alpha: 0.05),
            blurRadius: 10,
            offset: const Offset(0, 3),
          ),
        ],
      ),
      padding: EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(
                Icons.summarize_outlined,
                color: Colors.indigo.shade600,
                size: 20,
              ),
              SizedBox(width: 8),
              Text(
                global.language('amount'),
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: global.theme.textColor,
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),
          Center(
            child: Column(
              children: [
                SizedBox(
                  width: 30,
                  height: 30,
                  child: CircularProgressIndicator(
                    strokeWidth: 3,
                    valueColor: AlwaysStoppedAnimation<Color>(
                      Colors.indigo.shade600,
                    ),
                  ),
                ),
                const SizedBox(height: 12),
                Text(
                  'กำลังโหลดยอดรวม...',
                  style: TextStyle(fontSize: 14, color: global.theme.textSecondaryColor),
                ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildEmptyWidget() {
    return Container(
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: global.theme.dividerBorderColor),
        boxShadow: [
          BoxShadow(
            color: global.theme.textColor.withValues(alpha: 0.05),
            blurRadius: 10,
            offset: const Offset(0, 3),
          ),
        ],
      ),
      padding: EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(
                Icons.summarize_outlined,
                color: global.theme.iconSecondaryColor,
                size: 20,
              ),
              SizedBox(width: 8),
              Text(
                global.language('amount'),
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: global.theme.iconSecondaryColor,
                ),
              ),
            ],
          ),
          SizedBox(height: 16),
          Center(
            child: Text(
              global.language('no_total_data'),
              style: TextStyle(fontSize: 14, color: global.theme.textSecondaryColor),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildSummaryContent() {
    return Container(
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(16),
        border: Border.all(color: global.theme.dividerBorderColor),
        boxShadow: [
          BoxShadow(
            color: global.theme.textColor.withValues(alpha: 0.05),
            blurRadius: 10,
            offset: const Offset(0, 3),
          ),
        ],
      ),
      padding: EdgeInsets.all(16),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header
          Row(
            children: [
              Icon(
                Icons.summarize_outlined,
                color: Colors.indigo.shade600,
                size: 20,
              ),
              SizedBox(width: 8),
              Text(
                global.language('amount'),
                style: TextStyle(
                  fontSize: 16,
                  fontWeight: FontWeight.w600,
                  color: global.theme.textColor,
                ),
              ),
            ],
          ),
          const SizedBox(height: 16),

          // แยกการแสดงผลตาม reportType
          if (reportType == BiReportType.sale && totalSummary != null) ...[
            // Sale Report Summary
            _buildSummaryRow(
              global.language('database_master_info.item_count'),
              '${totalSummary!.totalRecords ?? 0} รายการ',
              global.theme.infoHighlightTextColor,
              Icons.receipt_long_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('pdf_goods_value'),
              ReportUtils.formatCurrency(totalSummary!.totalValue ?? 0),
              Colors.purple.shade600,
              Icons.inventory_2_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('discount'),
              ReportUtils.formatCurrency(
                totalSummary!.totalDetailDiscount ?? 0,
              ),
              Colors.purple.shade600,
              Icons.inventory_2_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('after_discount'),
              ReportUtils.formatCurrency(totalSummary!.totalAfterDiscount ?? 0),
              Colors.cyan.shade600,
              Icons.discount_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('sale_daily_report_except_vat'),
              ReportUtils.formatCurrency(totalSummary!.totalExceptVat ?? 0),
              Colors.teal.shade600,
              Icons.remove_circle_outline,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('before_tax'),
              ReportUtils.formatCurrency(totalSummary!.totalBeforeVat ?? 0),
              global.theme.warningHighlightTextColor,
              Icons.calculate_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('sale_daily_report_vat_value'),
              ReportUtils.formatCurrency(totalSummary!.totalVatValue ?? 0),
              global.theme.negativeHighlightTextColor,
              Icons.percent_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('net_amount'),
              ReportUtils.formatCurrency(totalSummary!.totalAmount ?? 0),
              global.theme.positiveHighlightTextColor,
              Icons.payments_outlined,
            ),

            // Branch breakdown สำหรับ Sale Report
            if (totalSummary!.totalByBranch != null &&
                totalSummary!.totalByBranch!.length > 1) ...[
              const SizedBox(height: 16),
              Divider(color: global.theme.dividerBorderColor),
              const SizedBox(height: 12),
              Text(
                'แยกตามสาขา',
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: global.theme.iconColor,
                ),
              ),
              const SizedBox(height: 8),
              ...totalSummary!.totalByBranch!.map(
                (branch) => _buildBranchRow(branch),
              ),
            ],
          ] else if (reportType == BiReportType.saleDaily &&
              dailySummary != null) ...[
            // Sale Daily Report Summary
            _buildSummaryRow(
              global.language('database_master_info.item_count'),
              '${dailySummary!.totalDays ?? 0} รายการ',
              global.theme.infoHighlightTextColor,
              Icons.receipt_long_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('total_sales'),
              ReportUtils.formatCurrency(dailySummary!.totalAmount ?? 0),
              global.theme.positiveHighlightTextColor,
              Icons.payments_outlined,
            ),
            const SizedBox(height: 12),
            _buildSummaryRow(
              'เฉลี่ยต่อวัน',
              ReportUtils.formatCurrency(dailySummary!.averageDailyAmount ?? 0),
              Colors.purple.shade600,
              Icons.account_balance_wallet_outlined,
            ),
          ] else if (reportType == BiReportType.stockMovement &&
              stockSummary != null) ...[
            // Stock Movement Report Summary
            _buildSummaryRow(
              global.language('database_master_info.item_count'),
              '${stockSummary!.totalRecords} รายการ',
              global.theme.infoHighlightTextColor,
              Icons.receipt_long_outlined,
            ),
            const SizedBox(height: 12),
            _buildSummaryRow(
              'ยอดเข้า',
              '${stockSummary!.totalQtyIn!.toStringAsFixed(2)} ',
              global.theme.positiveHighlightTextColor,
              Icons.add_box_outlined,
            ),
            const SizedBox(height: 12),
            _buildSummaryRow(
              'ต้นทุนเฉลี่ยเข้า',
              ReportUtils.formatCurrency(stockSummary!.averageCostIn!),
              global.theme.positiveHighlightTextColor,
              Icons.trending_up_outlined,
            ),
            const SizedBox(height: 12),
            _buildSummaryRow(
              'มูลค่าเข้า',
              ReportUtils.formatCurrency(stockSummary!.totalBalanceIn!),
              global.theme.positiveHighlightTextColor,
              Icons.account_balance_wallet_outlined,
            ),
            const SizedBox(height: 12),
            _buildSummaryRow(
              'ยอดออก',
              '${stockSummary!.totalQtyOut!.toStringAsFixed(2)} ',
              global.theme.negativeHighlightTextColor,
              Icons.remove_circle_outline,
            ),
            const SizedBox(height: 12),
            _buildSummaryRow(
              'ต้นทุนเฉลี่ยออก',
              ReportUtils.formatCurrency(stockSummary!.averageCostOut!),
              global.theme.negativeHighlightTextColor,
              Icons.trending_down_outlined,
            ),
            const SizedBox(height: 12),
            _buildSummaryRow(
              'มูลค่าออก',
              ReportUtils.formatCurrency(stockSummary!.totalBalanceOut!),
              global.theme.negativeHighlightTextColor,
              Icons.account_balance_wallet_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('stock_balance'),
              '${stockSummary!.finalBalanceQty!.toStringAsFixed(2)} ',
              global.theme.warningHighlightTextColor,
              Icons.inventory_outlined,
            ),
            const SizedBox(height: 12),
            _buildSummaryRow(
              'ต้นทุนเฉลี่ยสุดท้าย',
              ReportUtils.formatCurrency(stockSummary!.finalAverageCost!),
              global.theme.warningHighlightTextColor,
              Icons.calculate_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('balance_amount'),
              ReportUtils.formatCurrency(stockSummary!.finalBalanceAmount!),
              Colors.indigo.shade600,
              Icons.account_balance_wallet_outlined,
            ),
          ] else if (reportType == BiReportType.paymentDaily &&
              paymentSummary != null) ...[
            // Payment Daily Report Summary
            _buildSummaryRow(
              global.language('date_range'),
              '${ReportUtils.formatDate(paymentSummary!.fromDate)} - ${ReportUtils.formatDate(paymentSummary!.toDate)}',
              global.theme.infoHighlightTextColor,
              Icons.date_range,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('total_sales'),
              ReportUtils.formatCurrency(paymentSummary!.totalAmount),
              global.theme.positiveHighlightTextColor,
              Icons.account_balance_wallet_outlined,
            ),
            const SizedBox(height: 12),
            _buildSummaryRow(
              'ยอดชำระรวม',
              ReportUtils.formatCurrency(paymentSummary!.totalPayment),
              Colors.purple.shade600,
              Icons.payments_outlined,
            ),
          ] else if (reportType == BiReportType.saleReturn &&
              saleReturnSummary != null) ...[
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('database_master_info.item_count'),
              '${saleReturnSummary!.totalRecords} รายการ',
              global.theme.warningHighlightTextColor,
              Icons.receipt_long_outlined,
            ),
            const SizedBox(height: 12),
            _buildSummaryRow(
              'ยอดคืนรวม',
              ReportUtils.formatCurrency(saleReturnSummary!.totalAmount),
              global.theme.negativeHighlightTextColor,
              Icons.assignment_return_outlined,
            ),
          ] else if (reportType == BiReportType.stockBalance &&
              stockBalanceSummary != null) ...[
            // Stock Balance Report Summary
            _buildSummaryRow(
              global.language('date'),
              ReportUtils.formatDate(stockBalanceSummary!.toDate),
              global.theme.infoHighlightTextColor,
              Icons.date_range,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('database_master_info.item_count'),
              '${stockBalanceSummary!.totalRecords} รายการ',
              Colors.purple.shade600,
              Icons.receipt_long_outlined,
            ),
            const SizedBox(height: 12),
            _buildSummaryRow(
              'ยอดคงเหลือรวม',
              '${stockBalanceSummary!.totalBalanceQty.toStringAsFixed(2)} ',
              global.theme.warningHighlightTextColor,
              Icons.inventory_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('average_cost'),
              stockBalanceSummary!.averageCost != null
                  ? ReportUtils.formatCurrency(
                      stockBalanceSummary!.averageCost!,
                    )
                  : global.language('no_data_available'),
              global.theme.positiveHighlightTextColor,
              Icons.calculate_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              'มูลค่าคงเหลือรวม',
              stockBalanceSummary!.totalBalanceAmount != null
                  ? ReportUtils.formatCurrency(
                      stockBalanceSummary!.totalBalanceAmount!,
                    )
                  : global.language('no_data_available'),
              Colors.indigo.shade600,
              Icons.account_balance_wallet_outlined,
            ),
          ] else if (reportType == BiReportType.purchasepartial &&
              purchasePartialSummary != null) ...[
            // Purchase Partial Report Summary
            _buildSummaryRow(
              global.language('date_range'),
              purchasePartialSummary!.fromDate != null &&
                      purchasePartialSummary!.toDate != null
                  ? '${ReportUtils.formatDate(purchasePartialSummary!.fromDate!)} - ${ReportUtils.formatDate(purchasePartialSummary!.toDate!)}'
                  : global.language('no_data_available'),
              global.theme.infoHighlightTextColor,
              Icons.date_range,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('database_master_info.item_count'),
              '${purchasePartialSummary!.totalRecords ?? 0} รายการ',
              Colors.purple.shade600,
              Icons.receipt_long_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('pdf_goods_value'),
              ReportUtils.formatCurrency(
                purchasePartialSummary!.totalValue ?? 0,
              ),
              global.theme.warningHighlightTextColor,
              Icons.inventory_2_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('sale_daily_report_except_vat'),
              ReportUtils.formatCurrency(
                purchasePartialSummary!.totalExceptVat ?? 0,
              ),
              Colors.teal.shade600,
              Icons.remove_circle_outline,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('before_tax'),
              ReportUtils.formatCurrency(
                purchasePartialSummary!.totalBeforeVat ?? 0,
              ),
              Colors.cyan.shade600,
              Icons.calculate_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('sale_daily_report_vat_value'),
              ReportUtils.formatCurrency(
                purchasePartialSummary!.totalVatValue ?? 0,
              ),
              global.theme.negativeHighlightTextColor,
              Icons.percent_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('net_amount'),
              ReportUtils.formatCurrency(
                purchasePartialSummary!.totalAmount ?? 0,
              ),
              global.theme.positiveHighlightTextColor,
              Icons.payments_outlined,
            ),
          ] else if (reportType == BiReportType.grossProfitByDocument &&
              grossProfitSummary != null) ...[
            // Gross Profit By Document Report Summary
            _buildSummaryRow(
              global.language('date_range'),
              '${ReportUtils.formatDate(grossProfitSummary!.fromdate!)} - ${ReportUtils.formatDate(grossProfitSummary!.todate!)}',
              global.theme.infoHighlightTextColor,
              Icons.date_range,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('database_master_info.item_count'),
              '${grossProfitSummary!.totalRecords} รายการ',
              Colors.purple.shade600,
              Icons.receipt_long_outlined,
            ),
            const SizedBox(height: 12),
            _buildSummaryRow(
              'มูลค่าก่อน VAT',
              ReportUtils.formatCurrency(
                grossProfitSummary!.totalTotalBeforeVat!,
              ),
              global.theme.warningHighlightTextColor,
              Icons.inventory_2_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('total_discount'),
              ReportUtils.formatCurrency(
                grossProfitSummary!.totalDetailTotalDiscount!,
              ),
              Colors.cyan.shade600,
              Icons.discount_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('after_discount'),
              ReportUtils.formatCurrency(
                grossProfitSummary!.totalTotalAfterDiscount!,
              ),
              Colors.teal.shade600,
              Icons.calculate_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('total_cost'),
              ReportUtils.formatCurrency(grossProfitSummary!.totalTotalCost!),
              global.theme.negativeHighlightTextColor,
              Icons.money_off_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('total_gross_profit'),
              ReportUtils.formatCurrency(grossProfitSummary!.totalPal!),
              grossProfitSummary!.totalPal! >= 0
                  ? global.theme.positiveHighlightTextColor
                  : global.theme.negativeHighlightTextColor,
              Icons.trending_up_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('avg_profit_pct'),
              '${grossProfitSummary!.averageProfitPercentage.toStringAsFixed(2)}%',
              grossProfitSummary!.averageProfitPercentage >= 0
                  ? global.theme.positiveHighlightTextColor
                  : global.theme.negativeHighlightTextColor,
              Icons.percent_outlined,
            ),
          ] else if (reportType == BiReportType.grossProfitByProduct &&
              grossProfitByProductSummary != null) ...[
            // Gross Profit By Product Report Summary
            _buildSummaryRow(
              global.language('date_range'),
              '${ReportUtils.formatDate(grossProfitByProductSummary!.fromdate!)} - ${ReportUtils.formatDate(grossProfitByProductSummary!.todate!)}',
              global.theme.infoHighlightTextColor,
              Icons.date_range,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('database_master_info.item_count'),
              '${grossProfitByProductSummary!.totalRecords} รายการ',
              Colors.purple.shade600,
              Icons.receipt_long_outlined,
            ),
            const SizedBox(height: 12),
            _buildSummaryRow(
              'จำนวนสินค้ารวม',
              grossProfitByProductSummary!.totalQty!.toStringAsFixed(2),
              Colors.indigo.shade600,
              Icons.inventory_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('total_sales'),
              ReportUtils.formatCurrency(
                grossProfitByProductSummary!.totalSumAmountExcludeVat!,
              ),
              global.theme.warningHighlightTextColor,
              Icons.shopping_cart_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('total_cost'),
              ReportUtils.formatCurrency(
                grossProfitByProductSummary!.totalTotalCost!,
              ),
              global.theme.negativeHighlightTextColor,
              Icons.money_off_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('total_gross_profit'),
              ReportUtils.formatCurrency(
                grossProfitByProductSummary!.totalPal!,
              ),
              grossProfitByProductSummary!.totalPal! >= 0
                  ? global.theme.positiveHighlightTextColor
                  : global.theme.negativeHighlightTextColor,
              Icons.trending_up_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('avg_profit_pct'),
              '${grossProfitByProductSummary!.averageProfitPercentage.toStringAsFixed(2)}%',
              grossProfitByProductSummary!.averageProfitPercentage >= 0
                  ? global.theme.positiveHighlightTextColor
                  : global.theme.negativeHighlightTextColor,
              Icons.percent_outlined,
            ),
          ] else if (reportType == BiReportType.vatSale &&
              vatSaleSummary != null) ...[
            // VAT Sale Summary
            _buildSummaryRow(
              global.language('date_range'),
              '${ReportUtils.formatDate(vatSaleSummary!.fromdate)} - ${ReportUtils.formatDate(vatSaleSummary!.todate)}',
              global.theme.infoHighlightTextColor,
              Icons.calendar_month_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('database_master_info.item_count'),
              '${vatSaleSummary!.totalRecords} รายการ',
              Colors.indigo.shade600,
              Icons.receipt_long_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('total_tax_base'),
              ReportUtils.formatCurrency(vatSaleSummary!.totalTotalBeforeVat),
              Colors.purple.shade600,
              Icons.attach_money_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('tax_included'),
              ReportUtils.formatCurrency(vatSaleSummary!.totalTotalVatValue),
              global.theme.infoHighlightTextColor,
              Icons.money_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('total_tax_exempt'),
              ReportUtils.formatCurrency(vatSaleSummary!.totalTotalExceptVat),
              global.theme.warningHighlightTextColor,
              Icons.money_off_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('product_amount'),
              ReportUtils.formatCurrency(vatSaleSummary!.totalTotalAmount),
              global.theme.positiveHighlightTextColor,
              Icons.payments_outlined,
            ),
          ] else if (reportType == BiReportType.vatBuy &&
              vatBuySummary != null) ...[
            // VAT Buy Summary
            _buildSummaryRow(
              global.language('date_range'),
              '${ReportUtils.formatDate(vatBuySummary!.fromdate)} - ${ReportUtils.formatDate(vatBuySummary!.todate)}',
              global.theme.infoHighlightTextColor,
              Icons.calendar_month_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('database_master_info.item_count'),
              '${vatBuySummary!.totalRecords} รายการ',
              Colors.indigo.shade600,
              Icons.receipt_long_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('total_tax_base'),
              ReportUtils.formatCurrency(vatBuySummary!.totalTotalBeforeVat),
              Colors.purple.shade600,
              Icons.attach_money_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('tax_included'),
              ReportUtils.formatCurrency(vatBuySummary!.totalTotalVatValue),
              global.theme.infoHighlightTextColor,
              Icons.money_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('total_tax_exempt'),
              ReportUtils.formatCurrency(vatBuySummary!.totalTotalExceptVat),
              global.theme.warningHighlightTextColor,
              Icons.money_off_outlined,
            ),
            SizedBox(height: 12),
            _buildSummaryRow(
              global.language('product_amount'),
              ReportUtils.formatCurrency(vatBuySummary!.totalTotalAmount),
              global.theme.positiveHighlightTextColor,
              Icons.payments_outlined,
            ),
          ] else ...[
            // Fallback - No data
            Center(
              child: Text(
                global.language('no_total_data'),
                style: TextStyle(fontSize: 14, color: global.theme.textSecondaryColor),
              ),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildSummaryRow(
    String label,
    String value,
    Color color,
    IconData icon,
  ) {
    return Row(
      children: [
        Container(
          width: 32,
          height: 32,
          decoration: BoxDecoration(
            color: color.withValues(alpha: 0.1),
            borderRadius: BorderRadius.circular(8),
          ),
          child: Icon(icon, size: 16, color: color),
        ),
        const SizedBox(width: 12),
        Expanded(
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Text(
                label,
                style: TextStyle(
                  fontSize: 12,
                  color: global.theme.textSecondaryColor,
                  fontWeight: FontWeight.w500,
                ),
              ),
              Text(
                value,
                style: TextStyle(
                  fontSize: 14,
                  color: global.theme.textColor,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
          ),
        ),
      ],
    );
  }

  Widget _buildBranchRow(SaleReportSummaryByBranch branch) {
    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: global.theme.backgroundColor,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: global.theme.dividerBorderColor),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.store_outlined, size: 14, color: global.theme.textSecondaryColor),
              const SizedBox(width: 6),
              Text(
                'สาขา ${branch.branchcode ?? '-'}',
                style: TextStyle(
                  fontSize: 13,
                  fontWeight: FontWeight.w600,
                  color: global.theme.iconColor,
                ),
              ),
            ],
          ),
          const SizedBox(height: 6),
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [
              Text(
                '${branch.totalRecords ?? 0} รายการ',
                style: TextStyle(fontSize: 12, color: global.theme.textSecondaryColor),
              ),
              Text(
                ReportUtils.formatCurrency(branch.totalAmount ?? 0),
                style: TextStyle(
                  fontSize: 12,
                  color: global.theme.textColor,
                  fontWeight: FontWeight.w600,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }
}
