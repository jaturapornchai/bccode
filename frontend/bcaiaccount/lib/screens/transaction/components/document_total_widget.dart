import 'package:flutter/material.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/global.dart' as global;

class DocumentTotalWidget extends StatelessWidget {
  final TransactionModel screenData;
  final bool showOnlyTotalValue;

  const DocumentTotalWidget({super.key, required this.screenData, this.showOnlyTotalValue = false});

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(top: 4.0),
      padding: const EdgeInsets.all(4.0),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.grey[300]!),
        boxShadow: [BoxShadow(color: Colors.grey.withValues(alpha: 0.15), spreadRadius: 1, blurRadius: 4, offset: const Offset(0, 2))],
      ),
      child: Column(children: _buildFinancialSummary()),
    );
  }

  List<Widget> _buildFinancialSummary() {
    List<Widget> summaryWidgets = [];

    // ตรวจสอบ multi-currency
    final String baseCurrency = screenData.currency ?? global.getBaseCurrency();
    final String docCurrency = screenData.docCurrency ?? baseCurrency;
    final String docCurrencySymbol = screenData.docCurrencySymbol ?? '';
    double safeRate = screenData.exchangerate ?? 1.0;
    if (safeRate <= 0 || safeRate.isNaN || safeRate.isInfinite) safeRate = 1.0;
    final bool isMulti = _isMultiCurrencyDisplay(docCurrency, baseCurrency, safeRate);

    // Total value — ใช้ totalvalueDoc ที่คำนวณไว้แล้ว
    summaryWidgets.add(_buildFinancialRow('total_value', screenData.totalvalue,
        isMulti: isMulti, docSymbol: docCurrencySymbol, rate: safeRate,
        docValue: screenData.totalvalueDoc));

    // For advance payment, show only total value
    if (showOnlyTotalValue) {
      summaryWidgets.add(_buildFinalTotalRow());
      return summaryWidgets;
    }

    // Detail total discount
    if (screenData.detailtotaldiscount != 0) {
      summaryWidgets.add(_buildFinancialRowWithFormula('detail_total_discount', screenData.detailtotaldiscount ?? 0, screenData.detaildiscountformula ?? '',
          isMulti: isMulti, docSymbol: docCurrencySymbol, rate: safeRate));
    }

    // VAT discount amount
    if (screenData.totaldiscountvatamount != 0) {
      summaryWidgets.add(_buildFinancialRow('total_discount_vat_amount', screenData.totaldiscountvatamount ?? 0,
          isMulti: isMulti, docSymbol: docCurrencySymbol, rate: safeRate));
    }

    // Except VAT discount amount
    if (screenData.totaldiscountexceptvatamount != 0) {
      summaryWidgets.add(_buildFinancialRow('total_discount_except_vat_amount', screenData.totaldiscountexceptvatamount!,
          isMulti: isMulti, docSymbol: docCurrencySymbol, rate: safeRate));
    }

    // Total before VAT — ใช้ totalbeforevatDoc ที่คำนวณไว้แล้ว
    if (screenData.totalbeforevat != 0) {
      summaryWidgets.add(_buildFinancialRow('total_before_vat', screenData.totalbeforevat,
          isMulti: isMulti, docSymbol: docCurrencySymbol, rate: safeRate,
          docValue: screenData.totalbeforevatDoc));
    }

    // VAT amount — ใช้ totalvatvalueDoc ที่คำนวณไว้แล้ว
    if (screenData.totalvatvalue != 0) {
      summaryWidgets.add(_buildFinancialRowWithRate('doc_vat_amount', screenData.totalvatvalue, screenData.vatrate,
          isMulti: isMulti, docSymbol: docCurrencySymbol, rate: safeRate,
          docValue: screenData.totalvatvalueDoc));
    }

    // Total after VAT — ใช้ totalaftervatDoc ที่คำนวณไว้แล้ว
    if (screenData.totalaftervat != 0) {
      summaryWidgets.add(_buildFinancialRow('total_after_vat', screenData.totalaftervat,
          isMulti: isMulti, docSymbol: docCurrencySymbol, rate: safeRate,
          docValue: screenData.totalaftervatDoc));
    }

    // Total except VAT
    if (screenData.totalexceptvat != 0) {
      summaryWidgets.add(_buildFinancialRow('total_except_vat', screenData.totalexceptvat,
          isMulti: isMulti, docSymbol: docCurrencySymbol, rate: safeRate));
    }

    // Bill discount — ใช้ totaldiscountDoc ที่คำนวณไว้แล้ว
    if (screenData.totaldiscount != 0) {
      summaryWidgets.add(_buildFinancialRow('discount_bill', screenData.totaldiscount,
          isMulti: isMulti, docSymbol: docCurrencySymbol, rate: safeRate,
          docValue: screenData.totaldiscountDoc));
    }

    // Total amount after discount
    if (screenData.totalamountafterdiscount != 0) {
      summaryWidgets.add(_buildFinancialRow('total_amount_after_discount', screenData.totalamountafterdiscount!,
          isMulti: isMulti, docSymbol: docCurrencySymbol, rate: safeRate));
    }

    // Round amount
    if (screenData.roundamount != 0) {
      summaryWidgets.add(_buildFinancialRow('round_amount', screenData.roundamount!,
          isMulti: isMulti, docSymbol: docCurrencySymbol, rate: safeRate));
    }

    // Final total
    summaryWidgets.add(_buildFinalTotalRow());

    return summaryWidgets;
  }

  /// คำนวณค่า doc currency — ใช้ docValue ที่คำนวณไว้แล้ว, fallback หาร base/rate
  double _resolveDocValue(num baseValue, {double? docValue, double rate = 1.0}) {
    if (docValue != null && docValue != 0) return docValue;
    if (rate <= 0 || rate.isNaN || rate.isInfinite) return baseValue.toDouble();
    final dp = global.getDecimalDocument();
    return double.parse((baseValue.toDouble() / rate).toStringAsFixed(dp));
  }

  Widget _buildFinancialRow(String labelKey, num value, {bool isMulti = false, String docSymbol = '', double rate = 1.0, double? docValue}) {
    return Container(
      margin: const EdgeInsets.symmetric(vertical: 1.0),
      padding: const EdgeInsets.symmetric(horizontal: 8.0, vertical: 4.0),
      decoration: BoxDecoration(
        color: Colors.grey[50],
        borderRadius: BorderRadius.circular(4),
        border: Border.all(color: Colors.grey[200]!),
      ),
      child: Row(
        children: [
          Icon(_getIconForLabel(labelKey), size: 13, color: Colors.grey[600]),
          SizedBox(width: 6),
          Expanded(
            child: Text(
              global.language(labelKey),
              style: TextStyle(fontWeight: FontWeight.w500, fontSize: 12, color: Colors.grey[700]),
            ),
          ),
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            mainAxisSize: MainAxisSize.min,
            children: [
              if (isMulti)
                Text(
                  '$docSymbol ${global.formatNumber(_resolveDocValue(value, docValue: docValue, rate: rate))}',
                  style: TextStyle(fontSize: 11, color: Colors.blue[600], fontWeight: FontWeight.w500),
                ),
              Text(
                global.formatNumber(value.toDouble()),
                style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 12, color: Colors.black87),
              ),
            ],
          ),
        ],
      ),
    );
  }

  IconData _getIconForLabel(String labelKey) {
    switch (labelKey) {
      case 'total_value':
        return Icons.inventory;
      case 'detail_total_discount':
        return Icons.discount;
      case 'total_discount_vat_amount':
      case 'total_discount_except_vat_amount':
        return Icons.percent;
      case 'total_before_vat':
      case 'total_after_vat':
      case 'total_except_vat':
        return Icons.receipt_long;
      case 'doc_vat_amount':
        return Icons.account_balance;
      case 'discount_bill':
        return Icons.local_offer;
      case 'total_amount_after_discount':
        return Icons.price_check;
      case 'round_amount':
        return Icons.rounded_corner;
      default:
        return Icons.monetization_on;
    }
  }

  Widget _buildFinancialRowWithFormula(String labelKey, num value, String formula, {bool isMulti = false, String docSymbol = '', double rate = 1.0, double? docValue}) {
    return Container(
      margin: const EdgeInsets.symmetric(vertical: 1.0),
      padding: const EdgeInsets.symmetric(horizontal: 8.0, vertical: 4.0),
      decoration: BoxDecoration(
        color: Colors.orange[50],
        borderRadius: BorderRadius.circular(4),
        border: Border.all(color: Colors.orange[200]!),
      ),
      child: Row(
        children: [
          Icon(Icons.calculate, size: 13, color: Colors.orange[600]),
          SizedBox(width: 6),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  global.language(labelKey),
                  style: TextStyle(fontWeight: FontWeight.w500, fontSize: 12, color: Colors.orange[700]),
                ),
                if (formula.isNotEmpty)
                  Text(
                    '${global.language('formula')}: $formula',
                    style: TextStyle(fontSize: 10, color: Colors.orange[600], fontStyle: FontStyle.italic),
                  ),
              ],
            ),
          ),
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            mainAxisSize: MainAxisSize.min,
            children: [
              if (isMulti)
                Text(
                  '$docSymbol ${global.formatNumber(_resolveDocValue(value, docValue: docValue, rate: rate))}',
                  style: TextStyle(fontSize: 11, color: Colors.blue[600], fontWeight: FontWeight.w500),
                ),
              Text(
                value.toString(),
                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 12, color: Colors.orange[800]),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildFinancialRowWithRate(String labelKey, num value, num vatRate, {bool isMulti = false, String docSymbol = '', double rate = 1.0, double? docValue}) {
    return Container(
      margin: const EdgeInsets.symmetric(vertical: 1.0),
      padding: const EdgeInsets.symmetric(horizontal: 8.0, vertical: 4.0),
      decoration: BoxDecoration(
        color: Colors.green[50],
        borderRadius: BorderRadius.circular(4),
        border: Border.all(color: Colors.green[200]!),
      ),
      child: Row(
        children: [
          Icon(Icons.percent, size: 13, color: Colors.green[600]),
          SizedBox(width: 6),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  global.language(labelKey),
                  style: TextStyle(fontWeight: FontWeight.w500, fontSize: 12, color: Colors.green[700]),
                ),
                Text(
                  '${global.language('rate')}: $vatRate%',
                  style: TextStyle(fontSize: 10, color: Colors.green[600], fontStyle: FontStyle.italic),
                ),
              ],
            ),
          ),
          Column(
            crossAxisAlignment: CrossAxisAlignment.end,
            mainAxisSize: MainAxisSize.min,
            children: [
              if (isMulti)
                Text(
                  '$docSymbol ${global.formatNumber(_resolveDocValue(value, docValue: docValue, rate: rate))}',
                  style: TextStyle(fontSize: 11, color: Colors.blue[600], fontWeight: FontWeight.w500),
                ),
              Text(
                global.formatNumber(value.toDouble()),
                style: TextStyle(fontWeight: FontWeight.bold, fontSize: 12, color: Colors.green[800]),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildFinalTotalRow() {
    // ใช้ totalamount ถ้ามีค่า ถ้าไม่มีให้ใช้ totalvalue แทน
    final double finalTotalBase = (screenData.totalamount != 0)
        ? screenData.totalamount + (screenData.roundamount ?? 0)
        : screenData.totalvalue.toDouble();

    // ตรวจสอบว่ามี multi-currency หรือไม่
    final String baseCurrency = screenData.currency ?? global.getBaseCurrency();
    final String baseCurrencySymbol = screenData.currencysymbol ?? global.getBaseCurrencySymbol();
    final String docCurrency = screenData.docCurrency ?? baseCurrency;
    final String docCurrencySymbol = screenData.docCurrencySymbol ?? baseCurrencySymbol;

    // ป้องกัน division by zero
    double safeExchangeRate = screenData.exchangerate ?? 1.0;
    if (safeExchangeRate <= 0 || safeExchangeRate.isNaN || safeExchangeRate.isInfinite) {
      safeExchangeRate = 1.0;
    }

    // ตรวจสอบ Multi-Currency
    final bool isMultiCurrency = _isMultiCurrencyDisplay(
      docCurrency,
      baseCurrency,
      safeExchangeRate
    );

    // ใช้ totalamountDoc ที่คำนวณไว้แล้ว (ถ้ามี) หรือคำนวณใหม่
    final double finalTotalDoc = isMultiCurrency
        ? _resolveDocValue(finalTotalBase, docValue: screenData.totalamountDoc, rate: safeExchangeRate)
        : finalTotalBase;

    return Container(
      margin: const EdgeInsets.only(top: 2.0),
      padding: const EdgeInsets.all(8.0),
      decoration: BoxDecoration(
        gradient: LinearGradient(colors: [Colors.blue[500]!, Colors.blue[600]!], begin: Alignment.topLeft, end: Alignment.bottomRight),
        borderRadius: BorderRadius.circular(6),
      ),
      child: Column(
        children: [
          // ยอดรวมในสกุลเงินเอกสาร (ถ้าเป็นสกุลเงินต่างประเทศ)
          if (isMultiCurrency) ...[
            Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Icon(Icons.receipt, color: Colors.white70, size: 14),
                SizedBox(width: 4),
                Text(
                  '${global.language('sum_pay')} ($docCurrency)',
                  style: const TextStyle(color: Colors.white70, fontSize: 12, fontWeight: FontWeight.w500),
                ),
                Spacer(),
                Text(
                  '$docCurrencySymbol ${global.formatNumber(finalTotalDoc)}',
                  style: const TextStyle(color: Colors.white70, fontSize: 14, fontWeight: FontWeight.bold),
                ),
              ],
            ),
            const SizedBox(height: 4),
            Divider(color: Colors.white.withValues(alpha: 0.3), height: 1),
            const SizedBox(height: 4),
          ],
          // ยอดรวมในสกุลเงินหลัก (Base Currency)
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Icon(isMultiCurrency ? Icons.account_balance : Icons.monetization_on, color: Colors.white, size: isMultiCurrency ? 16 : 18),
              SizedBox(width: 4),
              Text(
                isMultiCurrency
                    ? '${global.language('sum_pay')} ($baseCurrency)'
                    : global.language('sum_pay'),
                style: TextStyle(color: Colors.white, fontSize: isMultiCurrency ? 14 : 18, fontWeight: FontWeight.w500),
              ),
              Spacer(),
              Text(
                isMultiCurrency
                    ? '$baseCurrencySymbol ${global.formatNumber(finalTotalBase)}'
                    : global.formatNumber(finalTotalBase),
                style: TextStyle(color: Colors.white, fontSize: isMultiCurrency ? 16 : 18, fontWeight: FontWeight.bold),
              ),
            ],
          ),
          // แสดงอัตราแลกเปลี่ยน (ถ้ามี)
          // Exchange Rate: 1 USD = 35 THB (1 DocCurrency = X BaseCurrency)
          if (isMultiCurrency) ...[
            const SizedBox(height: 4),
            Row(
              mainAxisAlignment: MainAxisAlignment.end,
              children: [
                Icon(Icons.sync_alt, color: Colors.white54, size: 10),
                const SizedBox(width: 4),
                Text(
                  '1 $docCurrency = ${safeExchangeRate.toStringAsFixed(4)} $baseCurrency',
                  style: TextStyle(color: Colors.white54, fontSize: 10, fontStyle: FontStyle.italic),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }

  /// ตรวจสอบว่าควรแสดง Multi-Currency หรือไม่
  bool _isMultiCurrencyDisplay(String docCurrency, String baseCurrency, double exchangeRate) {
    if (docCurrency.isEmpty || baseCurrency.isEmpty) return false;
    if (docCurrency == baseCurrency) return false;
    if (exchangeRate == 1.0) return false;
    return true;
  }
}
