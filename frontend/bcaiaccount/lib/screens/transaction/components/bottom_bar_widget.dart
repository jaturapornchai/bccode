import 'package:flutter/material.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/screen_search/cart_search_screen.dart';
import 'package:smlaicloud/global.dart' as global;

class BottomBarWidget extends StatelessWidget {
  final global.TransactionTypeEnum transactionType;
  final TransactionModel screenData;
  final List<String> cartList;
  final Function() onBarcodePressed;
  final Function(ProductBarcodeModel) onProductAdded;
  final Function() onEmptyProductAdded;
  final Function(void Function()) setState;
  final BuildContext context;

  /// ถ้าส่ง callback นี้มา → กดค้นหา/เพิ่ม จะเรียก callback แทน Navigator.push
  /// ใช้สำหรับ persistent search overlay (ไม่ dispose จอค้นหา)
  final VoidCallback? onSearchOverride;

  const BottomBarWidget({
    super.key,
    required this.transactionType,
    required this.screenData,
    required this.cartList,
    required this.onBarcodePressed,
    required this.onProductAdded,
    required this.onEmptyProductAdded,
    required this.setState,
    required this.context,
    this.onSearchOverride,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
      decoration: BoxDecoration(
        color: global.theme.surfaceColor,
        border: Border.all(color: global.theme.dividerBorderColor),
        borderRadius: BorderRadius.circular(10),
      ),
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      child: Row(
        children: [
          // Add Product Actions (Left side)
          Expanded(child: _buildAddProductSection()),

          const SizedBox(width: 8),

          // Cart Button (if applicable)
          if (_shouldShowCartButton()) ...[_buildCartButton(), const SizedBox(width: 8)],

          // Total Value Display (Right side)
          if (_shouldShowTotalValue()) _buildTotalValueDisplay(),
        ],
      ),
    );
  }

  Widget _buildAddProductSection() {
    // ตรวจสอบว่าเป็น transaction type ที่เกี่ยวข้องกับการชำระเงิน
    bool isPaymentType =
        transactionType == global.TransactionTypeEnum.advancePayment ||
        transactionType == global.TransactionTypeEnum.advancePaymentRefund ||
        transactionType == global.TransactionTypeEnum.deposit ||
        transactionType == global.TransactionTypeEnum.depositRefund ||
        transactionType == global.TransactionTypeEnum.paidAdvance ||
        transactionType == global.TransactionTypeEnum.paidAdvanceRefund ||
        transactionType == global.TransactionTypeEnum.receiveDeposit ||
        transactionType == global.TransactionTypeEnum.receiveDepositRefund;

    return Wrap(
      children: [
        Icon(Icons.add_shopping_cart, size: 14, color: global.theme.primaryColor),
        SizedBox(width: 4),
        Text(
          '${global.language('add_product')}:',
          style: TextStyle(fontSize: 12, fontWeight: FontWeight.w600, color: global.theme.primaryColor),
        ),
        const SizedBox(width: 8),

        // แสดงเฉพาะปุ่ม "เพิ่ม" สำหรับ transaction type ที่เกี่ยวข้องกับการชำระเงิน
        if (isPaymentType) ...[
          _buildSimpleButton(icon: Icons.add, label: global.language('add'), onPressed: onEmptyProductAdded),
        ] else ...[
          // แสดงปุ่มสแกนและเพิ่มสินค้า
          _buildSimpleButton(icon: Icons.qr_code_scanner, label: global.language('scan'), onPressed: onBarcodePressed),
          SizedBox(width: 6),
          _buildSimpleButton(icon: Icons.add, label: global.language('add'), onPressed: onEmptyProductAdded),
        ],
      ],
    );
  }

  Widget _buildCartButton() {
    return TextButton.icon(
      onPressed: _handleCartButtonPressed,
      icon: Icon(Icons.shopping_cart, size: 16),
      label: Text(global.language('cart')),
      style: TextButton.styleFrom(foregroundColor: global.theme.textSecondaryColor, padding: EdgeInsets.symmetric(horizontal: 8, vertical: 4)),
    );
  }

  Widget _buildTotalValueDisplay() {
    final bool isMultiCurrency = _isMultiCurrency();
    final String docCurrencySymbol = screenData.docCurrencySymbol ?? screenData.docCurrency ?? '';
    final String baseCurrencySymbol = screenData.currencysymbol ?? screenData.currency ?? '฿';

    return Row(
      mainAxisSize: MainAxisSize.min,
      children: [
        Icon(Icons.calculate, size: 14, color: global.theme.primaryColor),
        SizedBox(width: 4),
        Text(
          '${global.language("total_value")}: ',
          style: TextStyle(fontSize: 12, color: global.theme.primaryColor, fontWeight: FontWeight.w600),
        ),
        // Document Currency Value
        Container(
          padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
          decoration: BoxDecoration(color: global.theme.primaryColor, borderRadius: BorderRadius.circular(6)),
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              if (isMultiCurrency)
                Padding(
                  padding: const EdgeInsets.only(right: 2),
                  child: Text(
                    docCurrencySymbol,
                    style: TextStyle(fontSize: 10, fontWeight: FontWeight.normal, color: global.theme.onPrimaryColor.withValues(alpha: 0.7)),
                  ),
                ),
              Text(
                global.formatNumber(screenData.totalvalue),
                style: TextStyle(fontSize: 12, fontWeight: FontWeight.bold, color: global.theme.onPrimaryColor),
              ),
            ],
          ),
        ),
        // Base Currency Value (แสดงเมื่อเป็น Multi-Currency)
        if (isMultiCurrency) ...[
          const SizedBox(width: 4),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
            decoration: BoxDecoration(
              color: global.theme.dividerBorderColor,
              borderRadius: BorderRadius.circular(4),
              border: Border.all(color: global.theme.dividerBorderColor, width: 0.5),
            ),
            child: Text(
              '$baseCurrencySymbol ${global.formatNumber(_convertToBaseCurrency(screenData.totalvalue))}',
              style: TextStyle(fontSize: 10, fontWeight: FontWeight.w500, color: global.theme.textSecondaryColor),
            ),
          ),
        ],
      ],
    );
  }

  /// ตรวจสอบว่าเป็นเอกสาร Multi-Currency หรือไม่
  bool _isMultiCurrency() {
    return screenData.docCurrency != null &&
        screenData.docCurrency!.isNotEmpty &&
        screenData.currency != null &&
        screenData.currency!.isNotEmpty &&
        screenData.docCurrency != screenData.currency &&
        (screenData.exchangerate ?? 1.0) != 1.0;
  }

  /// แปลงจากสกุลเงินเอกสาร (Document Currency) เป็นสกุลเงินหลัก (Base Currency)
  double _convertToBaseCurrency(double docCurrencyValue) {
    final rate = screenData.exchangerate ?? 1.0;
    if (rate <= 0) return docCurrencyValue;
    return docCurrencyValue * rate;
  }

  Widget _buildSimpleButton({required IconData icon, required String label, required VoidCallback onPressed}) {
    return Container(
      decoration: BoxDecoration(
        border: Border.all(color: global.theme.dividerBorderColor),
        borderRadius: BorderRadius.circular(6),
        color: global.theme.cardColor,
      ),
      child: TextButton.icon(
        onPressed: onPressed,
        icon: Icon(icon, size: 12),
        label: Text(label),
        style: TextButton.styleFrom(
          foregroundColor: global.theme.textColor,
          padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
          minimumSize: Size.zero,
          tapTargetSize: MaterialTapTargetSize.shrinkWrap,
          textStyle: TextStyle(fontSize: 11),
        ),
      ),
    );
  }

  bool _shouldShowCartButton() {
    return transactionType == global.TransactionTypeEnum.stockreceiveproduct ||
        transactionType == global.TransactionTypeEnum.stockpickupproduct ||
        transactionType == global.TransactionTypeEnum.stocktransfer ||
        transactionType == global.TransactionTypeEnum.purchasepartial;
  }

  bool _shouldShowTotalValue() {
    return !(transactionType == global.TransactionTypeEnum.stocktransfer ||
        transactionType == global.TransactionTypeEnum.stockpickupproduct ||
        transactionType == global.TransactionTypeEnum.stockreturnproduct ||
        transactionType == global.TransactionTypeEnum.adjust);
  }

  void _handleCartButtonPressed() {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => CartListSearchScreen(
          onCartSelected: (details, cartId) {
            if (details.isNotEmpty) {
              setState(() {
                screenData.details!.addAll(details);
              });
              cartList.add(cartId);
            }
          },
        ),
      ),
    );
  }
}
