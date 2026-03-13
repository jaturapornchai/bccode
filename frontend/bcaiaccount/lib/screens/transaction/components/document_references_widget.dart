import 'package:flutter/material.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:intl/intl.dart';

class DocumentReferencesWidget extends StatelessWidget {
  final TransactionModel screenData;
  final global.TransactionTypeEnum transactionType;
  final Function(void Function()) setState;
  final Function({required String word}) searchDocRef;
  final Function(int index, String docno)? onDeleteDocReference;

  const DocumentReferencesWidget({
    super.key,
    required this.screenData,
    required this.transactionType,
    required this.setState,
    required this.searchDocRef,
    this.onDeleteDocReference,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      margin: const EdgeInsets.only(bottom: 15),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header with compact design
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            decoration: BoxDecoration(
              color: Colors.teal[50],
              borderRadius: BorderRadius.circular(8),
              border: Border.all(color: Colors.teal[200]!),
            ),
            child: Row(
              children: [
                Icon(
                  Icons.receipt_long,
                  size: 18,
                  color: Colors.teal[600],
                ),
                const SizedBox(width: 8),
                Expanded(
                  child: Text(
                    _getHeaderText(),
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.w600,
                      color: Colors.teal[700],
                    ),
                  ),
                ),
                // Add button with compact design
                Container(
                  decoration: BoxDecoration(
                    color: Colors.teal[600],
                    borderRadius: BorderRadius.circular(6),
                  ),
                  child: Material(
                    color: Colors.transparent,
                    child: InkWell(
                      borderRadius: BorderRadius.circular(6),
                      onTap: () => searchDocRef(word: ""),
                      child: Padding(
                        padding: EdgeInsets.all(6),
                        child: Row(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            Icon(
                              Icons.add,
                              size: 16,
                              color: global.theme.onPrimaryColor,
                            ),
                            SizedBox(width: 4),
                            Text(
                              global.language('add'),
                              style: TextStyle(
                                color: global.theme.onPrimaryColor,
                                fontSize: 12,
                                fontWeight: FontWeight.w500,
                              ),
                            ),
                          ],
                        ),
                      ),
                    ),
                  ),
                ),
              ],
            ),
          ),

          const SizedBox(height: 8),

          // Document list with compact cards
          if (screenData.docreferences != null &&
              screenData.docreferences!.isNotEmpty)
            Container(
              decoration: BoxDecoration(
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: global.theme.dividerBorderColor),
              ),
              child: Column(
                children:
                    screenData.docreferences!.asMap().entries.map((entry) {
                  int index = entry.key;
                  var docRef = entry.value;
                  final isLast = index == screenData.docreferences!.length - 1;

                  return Container(
                    decoration: BoxDecoration(
                      color: index % 2 == 0 ? global.theme.cardColor : global.theme.surfaceColor,
                      borderRadius: BorderRadius.only(
                        topLeft: Radius.circular(index == 0 ? 8 : 0),
                        topRight: Radius.circular(index == 0 ? 8 : 0),
                        bottomLeft: Radius.circular(isLast ? 8 : 0),
                        bottomRight: Radius.circular(isLast ? 8 : 0),
                      ),
                    ),
                    child: Padding(
                      padding: const EdgeInsets.symmetric(
                          horizontal: 12, vertical: 8),
                      child: Row(
                        children: [
                          // Order number badge
                          Container(
                            width: 24,
                            height: 24,
                            decoration: BoxDecoration(
                              color: Colors.teal[100],
                              borderRadius: BorderRadius.circular(12),
                              border: Border.all(color: Colors.teal[300]!),
                            ),
                            child: Center(
                              child: Text(
                                '${index + 1}',
                                style: TextStyle(
                                  fontSize: 11,
                                  fontWeight: FontWeight.bold,
                                  color: Colors.teal[700],
                                ),
                              ),
                            ),
                          ),

                          const SizedBox(width: 12),

                          // Document info - compact layout
                          Expanded(
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Text(
                                  docRef.docno,
                                  style: TextStyle(
                                    fontWeight: FontWeight.w600,
                                    fontSize: 13,
                                  ),
                                  overflow: TextOverflow.ellipsis,
                                ),
                                Text(
                                  DateFormat('dd/MM/yy HH:mm').format(
                                      DateTime.parse(docRef.docdatetime)),
                                  style: TextStyle(
                                    fontSize: 11,
                                    color: global.theme.textSecondaryColor,
                                  ),
                                ),
                              ],
                            ),
                          ),

                          // Delete button - compact
                          Container(
                            width: 32,
                            height: 32,
                            decoration: BoxDecoration(
                              color: Colors.red[50],
                              borderRadius: BorderRadius.circular(6),
                              border: Border.all(color: Colors.red[200]!),
                            ),
                            child: Material(
                              color: Colors.transparent,
                              child: InkWell(
                                borderRadius: BorderRadius.circular(6),
                                onTap: () => _handleDeleteDocReference(
                                    index, docRef.docno),
                                child: Icon(
                                  Icons.close,
                                  size: 16,
                                  color: global.theme.negativeHighlightTextColor,
                                ),
                              ),
                            ),
                          ),
                        ],
                      ),
                    ),
                  );
                }).toList(),
              ),
            )
          else
            // Empty state - more compact
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 16),
              decoration: BoxDecoration(
                color: global.theme.surfaceColor,
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: global.theme.dividerBorderColor),
              ),
              child: Row(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  Icon(
                    Icons.inbox_outlined,
                    size: 20,
                    color: global.theme.iconSecondaryColor,
                  ),
                  const SizedBox(width: 8),
                  Text(
                    _getEmptyStateText(),
                    style: TextStyle(
                      color: global.theme.textSecondaryColor,
                      fontSize: 13,
                      fontStyle: FontStyle.italic,
                    ),
                  ),
                ],
              ),
            ),
        ],
      ),
    );
  }

  String _getHeaderText() {
    switch (transactionType) {
      case global.TransactionTypeEnum.purchase:
        return global.language('po_number');
      case global.TransactionTypeEnum.purchasereturn:
        return global.language('purchase_doc_number');
      case global.TransactionTypeEnum.purchasepartial:
        return global.language('po_number');
      case global.TransactionTypeEnum.accrualreceive:
        return global.language('receive_doc_number');
      case global.TransactionTypeEnum.sale:
        return global.language('so_number');
      case global.TransactionTypeEnum.salereturn:
        return global.language('sale_doc_number');
      default:
        return 'not config type';
    }
  }

  String _getEmptyStateText() {
    switch (transactionType) {
      case global.TransactionTypeEnum.purchase:
        return global.language('no_po_number_yet');
      case global.TransactionTypeEnum.purchasereturn:
        return global.language('no_purchase_doc_yet');
      case global.TransactionTypeEnum.purchasepartial:
        return global.language('no_po_number_yet');
      case global.TransactionTypeEnum.accrualreceive:
        return global.language('no_receive_doc_yet');
      case global.TransactionTypeEnum.sale:
        return global.language('no_so_number_yet');
      case global.TransactionTypeEnum.salereturn:
        return global.language('no_sale_doc_yet');
      default:
        return 'not config type';
    }
  }

  void _handleDeleteDocReference(int index, String docno) {
    if (onDeleteDocReference != null) {
      onDeleteDocReference!(index, docno);
    } else {
      setState(() {
        screenData.docreferences!.removeAt(index);
      });
    }
  }
}
