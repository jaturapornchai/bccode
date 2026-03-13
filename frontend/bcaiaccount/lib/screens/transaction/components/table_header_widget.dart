import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

class TableHeaderWidget extends StatelessWidget {
  final global.TransactionTypeEnum transactionType;
  final List<global.DataTableHeader> headers;
  final BoxConstraints constraints;

  const TableHeaderWidget({super.key, required this.transactionType, required this.headers, required this.constraints});

  @override
  Widget build(BuildContext context) {
    // Special case for advance payment
    if (_isAdvancePaymentType()) {
      return Container(); // Header will be handled by AdvancePaymentWidget
    }

    if (MediaQuery.of(context).size.width > 799) {
      return Container(
        margin: const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
        decoration: BoxDecoration(
          color: global.theme.surfaceColor,
          border: Border.all(color: global.theme.dividerBorderColor),
          borderRadius: BorderRadius.circular(10),
        ),
        child: Table(
          columnWidths: {for (int i = 0; i < headers.length; i++) i: headers[i].code == 'line_number' ? const FixedColumnWidth(50.0) : FlexColumnWidth(headers[i].width)},
          children: [
            TableRow(
              decoration: BoxDecoration(color: global.theme.primaryColor, borderRadius: BorderRadius.circular(10)),
              children: headers
                  .map(
                    (header) => Container(
                      padding: EdgeInsets.symmetric(horizontal: 4, vertical: 4),
                      child: Semantics(
                        label: '${header.label} ${global.language('table_header')}',
                        header: true,
                        child: Align(
                          alignment: _shouldAlignRight(header.code) ? Alignment.centerRight : Alignment.centerLeft,
                          child: Text(
                            header.label,
                            textAlign: _shouldAlignRight(header.code) ? TextAlign.right : TextAlign.left,
                            style: TextStyle(fontSize: 12, color: global.theme.onPrimaryColor),
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                      ),
                    ),
                  )
                  .toList(),
            ),
          ],
        ),
      );
    }
    return Container();
  }

  bool _isAdvancePaymentType() {
    return transactionType == global.TransactionTypeEnum.advancePayment ||
        transactionType == global.TransactionTypeEnum.advancePaymentRefund ||
        transactionType == global.TransactionTypeEnum.deposit ||
        transactionType == global.TransactionTypeEnum.depositRefund ||
        transactionType == global.TransactionTypeEnum.paidAdvance ||
        transactionType == global.TransactionTypeEnum.paidAdvanceRefund ||
        transactionType == global.TransactionTypeEnum.receiveDeposit ||
        transactionType == global.TransactionTypeEnum.receiveDepositRefund;
  }

  // ตรวจสอบว่า column นี้ต้องชิดขวาหรือไม่ (column ตัวเลข)
  bool _shouldAlignRight(String code) {
    return code == 'product_qty' ||
        code == 'product_price' ||
        code == 'product_price_adjust' ||
        code == 'product_amount' ||
        code == 'eventqty' ||
        code == 'line_number';
  }
}
