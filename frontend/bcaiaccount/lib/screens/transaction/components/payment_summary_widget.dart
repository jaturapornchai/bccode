import 'package:flutter/material.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/global.dart' as global;

class PaymentSummaryWidget extends StatelessWidget {
  final TransactionModel screenData;
  final global.TransactionTypeEnum type;
  final TextEditingController payCashAmountController;
  final TextEditingController payDeliveryCashAmountController;
  final Function(int)? onPaymentTypeSelected;

  const PaymentSummaryWidget({
    super.key,
    required this.screenData,
    required this.type,
    required this.payCashAmountController,
    required this.payDeliveryCashAmountController,
    this.onPaymentTypeSelected,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.all(0),
      child: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Payment Summary Items
          _buildCardSummaryList(context),
        ],
      ),
    );
  }

  Widget _buildCardSummaryList(BuildContext context) {
    List<Widget> paymentItems = [];

    // แถวที่ 1: เงินสด, เงินโอน, บัตรเครดิต
    // 1. Cash - Always show (Icons.money)
    double cashAmount = payCashAmountController.text.isEmpty
        ? 0.0
        : double.tryParse(payCashAmountController.text.replaceAll(',', '')) ??
            0.0;
    paymentItems.add(_buildPaymentCard(
      context,
      icon: Icons.money,
      title: global.language("cash"),
      amount: cashAmount,
      showPayDetailIndex: 0,
    ));

    // 2. Transfer - Always show (Icons.transform_rounded)
    paymentItems.add(_buildPaymentCard(
      context,
      icon: Icons.transform_rounded,
      title: global.language("transfer"),
      amount: screenData.summoneytransfer ?? 0.0,
      showPayDetailIndex: 1,
    ));

    // 3. Credit Card - Show for valid transaction types (Icons.credit_card)
    if (type == global.TransactionTypeEnum.sale ||
        type == global.TransactionTypeEnum.paidAdvance ||
        type == global.TransactionTypeEnum.receiveDeposit) {
      paymentItems.add(_buildPaymentCard(
        context,
        icon: Icons.credit_card,
        title: global.language("credit_card"),
        amount: screenData.sumcreditcard ?? 0.0,
        showPayDetailIndex: 2,
      ));
    }

    // แถวที่ 2: เช็ค, คูปอง, คิวอาร์โค๊ด
    // 4. Cheque - Show for valid transaction types (Icons.featured_play_list_outlined)
    if (type == global.TransactionTypeEnum.sale ||
        type == global.TransactionTypeEnum.saleorder ||
        type == global.TransactionTypeEnum.paidAdvance ||
        type == global.TransactionTypeEnum.receiveDeposit) {
      paymentItems.add(_buildPaymentCard(
        context,
        icon: Icons.featured_play_list_outlined,
        title: global.language("cheque"),
        amount: screenData.sumcheque ?? 0.0,
        showPayDetailIndex: 3,
      ));
    }

    // 5. Coupon - Show for sale type (Icons.card_giftcard)
    if (type == global.TransactionTypeEnum.sale ||
        type == global.TransactionTypeEnum.saleorder) {
      paymentItems.add(_buildPaymentCard(
        context,
        icon: Icons.card_giftcard,
        title: global.language("coupon"),
        amount: screenData.sumcoupon ?? 0.0,
        showPayDetailIndex: 4,
      ));
    }

    // 6. QR Code - Show for valid transaction types (Icons.qr_code)
    if (type == global.TransactionTypeEnum.sale ||
        type == global.TransactionTypeEnum.saleorder ||
        type == global.TransactionTypeEnum.paidAdvance ||
        type == global.TransactionTypeEnum.receiveDeposit) {
      paymentItems.add(_buildPaymentCard(
        context,
        icon: Icons.qr_code,
        title: global.language("qr_code"),
        amount: screenData.sumqrcode ?? 0.0,
        showPayDetailIndex: 5,
      ));
    }

    // แถวที่ 3: เดริเวอรี่, แต้ม, เงินล่วงหน้า
    // 7. Delivery - Show for sale type (Icons.delivery_dining_rounded)
    if (type == global.TransactionTypeEnum.sale ||
        type == global.TransactionTypeEnum.saleorder) {
      double deliveryAmount = payDeliveryCashAmountController.text.isEmpty
          ? 0.0
          : double.tryParse(
                  payDeliveryCashAmountController.text.replaceAll(',', '')) ??
              0.0;
      paymentItems.add(_buildPaymentCard(
        context,
        icon: Icons.delivery_dining_rounded,
        title: global.language("delivery"),
        amount: deliveryAmount,
        showPayDetailIndex: 6,
      ));
    }

    // 8. Points - Show for sale type (Icons.star_rate_rounded)
    if (type == global.TransactionTypeEnum.sale ||
        type == global.TransactionTypeEnum.saleorder) {
      paymentItems.add(_buildPaymentCard(
        context,
        icon: Icons.star_rate_rounded,
        title: global.language('points'),
        amount: double.parse(screenData.paypointamount.toString()),
        showPayDetailIndex: 7,
      ));
    }

    // 9. Advance Payment - Show based on transaction type (Icons.money_off)
    if (type == global.TransactionTypeEnum.purchase ||
        type == global.TransactionTypeEnum.sale) {
      if (screenData.sumadvancepayment != null) {
        paymentItems.add(_buildPaymentCard(
          context,
          icon: Icons.money_off,
          title: global.language('advance_payment'),
          amount: screenData.sumadvancepayment ?? 0.0,
          showPayDetailIndex: 8,
        ));
      }
    }

    // Fixed 3 columns layout with larger cards
    int crossAxisCount = 3;
    double childAspectRatio = 2;
    double spacing = 10;

    return paymentItems.isEmpty
        ? Center(
            child: Padding(
              padding: EdgeInsets.all(20),
              child: Text(
                global.language('no_payment'),
                style: TextStyle(
                  color: global.theme.textSecondaryColor,
                  fontSize: 14,
                  fontStyle: FontStyle.italic,
                ),
              ),
            ),
          )
        : LayoutBuilder(
            builder: (context, constraints) {
              return GridView.builder(
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                  crossAxisCount: crossAxisCount,
                  crossAxisSpacing: spacing,
                  mainAxisSpacing: spacing,
                  childAspectRatio: childAspectRatio,
                ),
                itemCount: paymentItems.length,
                itemBuilder: (context, index) {
                  return paymentItems[index];
                },
              );
            },
          );
  }

  Widget _buildPaymentCard(
    BuildContext context, {
    required IconData icon,
    required String title,
    required double amount,
    required int showPayDetailIndex,
  }) {
    final String formattedAmount = global.formatNumber(amount);
    final bool hasAmount = amount > 0;

    // Larger sizes for better visibility
    double titleSize = 16;
    double amountSize = 22;

    return Material(
      color: Colors.transparent,
      child: InkWell(
        onTap: () {
          if (onPaymentTypeSelected != null) {
            onPaymentTypeSelected!(showPayDetailIndex);
          }
        },
        borderRadius: BorderRadius.circular(16),
        child: Container(
          decoration: BoxDecoration(
            borderRadius: BorderRadius.circular(16),
            border: Border.all(
              color: hasAmount ? global.theme.infoHighlightTextColor : global.theme.dividerBorderColor,
              width: 2,
            ),
            boxShadow: [
              BoxShadow(
                color: (hasAmount ? global.theme.infoHighlightColor : global.theme.dividerBorderColor)
                    .withValues(alpha: 0.5),
                blurRadius: 6,
                offset: const Offset(0, 3),
              ),
            ],
          ),
          child: Stack(
            children: [
              // Main content
              Container(
                width: double.infinity,
                height: double.infinity,
                decoration: BoxDecoration(
                  color: hasAmount ? global.theme.infoHighlightColor : global.theme.cardColor,
                  borderRadius: BorderRadius.circular(14),
                ),
                child: Column(
                  mainAxisAlignment: MainAxisAlignment.center,
                  crossAxisAlignment: CrossAxisAlignment.center,
                  children: [
                    // Title
                    Text(
                      title,
                      textAlign: TextAlign.center,
                      style: TextStyle(
                        fontSize: titleSize,
                        color: hasAmount
                            ? global.theme.infoHighlightTextColor
                            : global.theme.textColor,
                        fontWeight: FontWeight.w700,
                      ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),

                    const SizedBox(height: 4),

                    // Amount
                    Text(
                      formattedAmount,
                      textAlign: TextAlign.center,
                      style: TextStyle(
                        fontSize: amountSize,
                        color: hasAmount
                            ? global.theme.infoHighlightTextColor
                            : global.theme.textColor,
                        fontWeight: FontWeight.bold,
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ],
                ),
              ),

              // Icon at bottom right corner
              Positioned(
                bottom: 8,
                right: 8,
                child: Icon(
                  icon,
                  size: 70,
                  color: hasAmount
                      ? global.theme.infoHighlightTextColor.withValues(alpha: 0.1)
                      : global.theme.dividerBorderColor.withValues(alpha: 0.3),
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}
