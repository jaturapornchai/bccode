import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/point_transaction/point_transaction_bloc.dart';
import 'package:smlaicloud/model/point_transaction_model.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:loading_animation_widget/loading_animation_widget.dart';
import 'package:intl/intl.dart';
import 'package:smlaicloud/components/loading_overlay.dart';

class PointTransactionScreen extends StatefulWidget {
  final String debtorCode;
  final String debtorName;
  final String pointsCode;

  const PointTransactionScreen({
    super.key,
    required this.debtorCode,
    required this.debtorName,
    required this.pointsCode,
  });

  @override
  State<PointTransactionScreen> createState() => _PointTransactionScreenState();
}

class _PointTransactionScreenState extends State<PointTransactionScreen>
    with global.ThemeRefreshMixin {
  List<PointTransactionModel> transactions = [];
  bool loadingData = false;

  @override
  void initState() {
    super.initState();
    loadTransactions();
  }

  void loadTransactions() {
    context.read<PointTransactionBloc>().add(
      PointTransactionLoadByDebtorCode(debtorCode: widget.pointsCode),
    );
  }

  String getTransactionTypeText(int type) {
    switch (type) {
      case 1:
        return global.language('point_transaction.received_points');
      case 2:
        return global.language('point_transaction.used_points');
      case 3:
        return global.language('point_transaction.add_points');
      default:
        return global.language('point_transaction.unknown');
    }
  }

  Color getTransactionTypeColor(int type) {
    switch (type) {
      case 1:
        return global.theme.positiveHighlightTextColor;
      case 2:
        return global.theme.negativeHighlightTextColor;
      case 3:
        return global.theme.infoHighlightTextColor;
      default:
        return global.theme.iconSecondaryColor;
    }
  }

  Widget transactionListItem(PointTransactionModel transaction) {
    // แสดงปุ่มลบเฉพาะ type 3 (เพิ่มแต้มแบบ manual)
    bool canDelete = transaction.transactiontype == 3;

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 10, vertical: 5),
      child: Padding(
        padding: const EdgeInsets.all(15),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Expanded(
                  child: Text(
                    transaction.transactiondocno,
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 16,
                    ),
                  ),
                ),
                Row(
                  children: [
                    Container(
                      padding: const EdgeInsets.symmetric(
                        horizontal: 8,
                        vertical: 4,
                      ),
                      decoration: BoxDecoration(
                        color: getTransactionTypeColor(
                          transaction.transactiontype,
                        ),
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Text(
                        getTransactionTypeText(transaction.transactiontype),
                        style: TextStyle(
                          color: global.theme.onPrimaryColor,
                          fontSize: 12,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                    if (canDelete) ...[
                      const SizedBox(width: 8),
                      IconButton(
                        icon: Icon(Icons.delete, color: global.theme.negativeHighlightTextColor),
                        iconSize: 20,
                        constraints: const BoxConstraints(),
                        padding: EdgeInsets.zero,
                        onPressed: () => _showDeleteConfirmDialog(transaction),
                      ),
                    ],
                  ],
                ),
              ],
            ),
            const SizedBox(height: 8),
            Text(
              DateFormat(
                'dd/MM/yyyy HH:mm:ss',
              ).format(transaction.transactiondate),
              style: TextStyle(color: global.theme.textSecondaryColor, fontSize: 14),
            ),
            SizedBox(height: 8),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  '${global.language('point_transaction.points_code')}: ${transaction.pointscode}',
                ),
                Text(
                  '${transaction.pointamount > 0 ? '+' : ''}${transaction.pointamount}',
                  style: TextStyle(
                    color: getTransactionTypeColor(transaction.transactiontype),
                    fontWeight: FontWeight.bold,
                    fontSize: 16,
                  ),
                ),
              ],
            ),
            SizedBox(height: 5),
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  '${global.language('point_transaction.balance_before')}: ${transaction.balancebefore}',
                ),
                Text(
                  '${global.language('point_transaction.balance_after')}: ${transaction.balanceafter}',
                ),
              ],
            ),
            if (transaction.description.isNotEmpty) ...[
              const SizedBox(height: 8),
              Text(
                transaction.description,
                style: TextStyle(color: global.theme.textSecondaryColor, fontSize: 13),
              ),
            ],
          ],
        ),
      ),
    );
  }

  void _showDeleteConfirmDialog(PointTransactionModel transaction) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (BuildContext dialogContext) {
        bool isLoading = false;
        String loadingMessage = global.language('deleting_items');

        return BlocListener<PointTransactionBloc, PointTransactionState>(
          listener: (context, state) {
            if (state is PointTransactionDeleteSuccess) {
              // ปิด dialog เมื่อลบสำเร็จ
              Navigator.of(dialogContext).pop();
              // โหลดข้อมูลใหม่
              loadTransactions();
            } else if (state is PointTransactionDeleteFailed) {
              // ปิด dialog เมื่อลบไม่สำเร็จ
              Navigator.of(dialogContext).pop();
            }
          },
          child: StatefulBuilder(
            builder: (context, setDialogState) {
              return Stack(
                children: [
                  AlertDialog(
                    shape: RoundedRectangleBorder(
                      borderRadius: BorderRadius.circular(16),
                    ),
                    title: Row(
                      children: [
                        Container(
                          padding: EdgeInsets.all(8),
                          decoration: BoxDecoration(
                            color: Colors.red.withValues(alpha: 0.1),
                            borderRadius: BorderRadius.circular(8),
                          ),
                          child: Icon(
                            Icons.warning_rounded,
                            color: global.theme.negativeHighlightTextColor,
                            size: 28,
                          ),
                        ),
                        SizedBox(width: 12),
                        Expanded(
                          child: Text(
                            global.language('confirm'),
                            style: TextStyle(
                              fontSize: 20,
                              fontWeight: FontWeight.bold,
                            ),
                          ),
                        ),
                      ],
                    ),
                    content: Container(
                      padding: const EdgeInsets.symmetric(vertical: 8),
                      child: Column(
                        mainAxisSize: MainAxisSize.min,
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            global.language('confirm_delete_point_transaction'),
                            style: TextStyle(
                              fontSize: 16,
                              fontWeight: FontWeight.w500,
                            ),
                          ),
                          const SizedBox(height: 16),
                          Container(
                            padding: EdgeInsets.all(12),
                            decoration: BoxDecoration(
                              color: global.theme.surfaceColor,
                              borderRadius: BorderRadius.circular(8),
                              border: Border.all(color: global.theme.dividerBorderColor),
                            ),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Row(
                                  children: [
                                    Icon(
                                      Icons.receipt_long,
                                      size: 16,
                                      color: global.theme.iconSecondaryColor,
                                    ),
                                    const SizedBox(width: 8),
                                    Text(
                                      '${global.language("docno")}:',
                                      style: TextStyle(
                                        color: global.theme.iconSecondaryColor,
                                        fontSize: 13,
                                      ),
                                    ),
                                    const SizedBox(width: 8),
                                    Expanded(
                                      child: Text(
                                        transaction.transactiondocno,
                                        style: TextStyle(
                                          fontWeight: FontWeight.bold,
                                          fontSize: 14,
                                        ),
                                      ),
                                    ),
                                  ],
                                ),
                                const Divider(height: 16),
                                Row(
                                  children: [
                                    Icon(
                                      Icons.stars_rounded,
                                      size: 16,
                                      color: global.theme.iconSecondaryColor,
                                    ),
                                    const SizedBox(width: 8),
                                    Text(
                                      '${global.language("point_amount")}:',
                                      style: TextStyle(
                                        color: global.theme.iconSecondaryColor,
                                        fontSize: 13,
                                      ),
                                    ),
                                    const SizedBox(width: 8),
                                    Text(
                                      '${transaction.pointamount > 0 ? '+' : ''}${transaction.pointamount}',
                                      style: TextStyle(
                                        fontWeight: FontWeight.bold,
                                        fontSize: 16,
                                        color: getTransactionTypeColor(
                                          transaction.transactiontype,
                                        ),
                                      ),
                                    ),
                                  ],
                                ),
                                if (transaction.description.isNotEmpty) ...[
                                  const Divider(height: 16),
                                  Row(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
                                      Icon(
                                        Icons.description,
                                        size: 16,
                                        color: global.theme.iconSecondaryColor,
                                      ),
                                      const SizedBox(width: 8),
                                      Expanded(
                                        child: Column(
                                          crossAxisAlignment:
                                              CrossAxisAlignment.start,
                                          children: [
                                            Text(
                                              '${global.language("description")}:',
                                              style: TextStyle(
                                                color: global.theme.iconSecondaryColor,
                                                fontSize: 13,
                                              ),
                                            ),
                                            const SizedBox(height: 4),
                                            Text(
                                              transaction.description,
                                              style: TextStyle(
                                                fontSize: 13,
                                              ),
                                            ),
                                          ],
                                        ),
                                      ),
                                    ],
                                  ),
                                ],
                              ],
                            ),
                          ),
                          const SizedBox(height: 16),
                          Container(
                            padding: const EdgeInsets.all(8),
                            decoration: BoxDecoration(
                              color: Colors.orange[50],
                              borderRadius: BorderRadius.circular(8),
                              border: Border.all(color: Colors.orange[200]!),
                            ),
                            child: Row(
                              children: [
                                Icon(
                                  Icons.info_outline,
                                  color: Colors.orange[700],
                                  size: 20,
                                ),
                                const SizedBox(width: 8),
                                Expanded(
                                  child: Text(
                                    global.language('delete_cannot_undo'),
                                    style: TextStyle(
                                      color: Colors.orange[900],
                                      fontSize: 12,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                    actions: [
                      TweenAnimationBuilder<double>(
                        tween: Tween(begin: 0.0, end: 1.0),
                        duration: const Duration(milliseconds: 300),
                        curve: Curves.easeOut,
                        builder: (context, value, child) {
                          return Transform.scale(
                            scale: value,
                            child: Opacity(opacity: value, child: child),
                          );
                        },
                        child: TextButton(
                          onPressed: isLoading
                              ? null
                              : () async {
                                  await Future.delayed(
                                    const Duration(milliseconds: 150),
                                  );
                                  if (dialogContext.mounted) {
                                    Navigator.of(dialogContext).pop();
                                  }
                                },
                          child: Text(global.language('cancel')),
                        ),
                      ),
                      TweenAnimationBuilder<double>(
                        tween: Tween(begin: 0.0, end: 1.0),
                        duration: const Duration(milliseconds: 400),
                        curve: Curves.easeOut,
                        builder: (context, value, child) {
                          return Transform.scale(
                            scale: value,
                            child: Opacity(opacity: value, child: child),
                          );
                        },
                        child: ElevatedButton.icon(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: global.theme.negativeHighlightTextColor,
                            foregroundColor: global.theme.onPrimaryColor,
                            padding: const EdgeInsets.symmetric(
                              horizontal: 20,
                              vertical: 12,
                            ),
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(8),
                            ),
                          ),
                          onPressed: isLoading
                              ? null
                              : () async {
                                  // Animation feedback when pressed
                                  await Future.delayed(
                                    const Duration(milliseconds: 150),
                                  );

                                  // แสดง loading
                                  setDialogState(() {
                                    isLoading = true;
                                    loadingMessage = global.language('deleting_items');
                                  });

                                  // เรียก API ลบ
                                  context.read<PointTransactionBloc>().add(
                                    PointTransactionDeleteManual(
                                      pointscode: transaction.pointscode,
                                      docno: transaction.transactiondocno,
                                    ),
                                  );
                                },
                          icon: Icon(Icons.delete_outline, size: 20),
                          label: Text(global.language('delete')),
                        ),
                      ),
                    ],
                  ),
                  // Loading Overlay
                  if (isLoading) LoadingOverlay(message: loadingMessage),
                ],
              );
            },
          ),
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        title: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Text(global.language('point_transaction.points_history')),
            Text(
              '${widget.debtorCode} - ${widget.debtorName}',
              style: TextStyle(fontSize: 14),
            ),
          ],
        ),
        actions: [
          IconButton(
            icon: Icon(Icons.refresh),
            onPressed: loadTransactions,
          ),
        ],
      ),
      body: BlocListener<PointTransactionBloc, PointTransactionState>(
        listener: (context, state) {
          if (state is PointTransactionLoadSuccess) {
            setState(() {
              loadingData = false;
              transactions = state.transactions;
            });
          }
          if (state is PointTransactionLoadFailed) {
            setState(() {
              loadingData = false;
            });
            global.showErrorSnackBar(context, '${global.language('point_transaction.error_occurred')}: ${state.message}');
          }
          if (state is PointTransactionInProgress) {
            setState(() {
              loadingData = true;
            });
          }
          // จัดการผลลัพธ์การลบ
          if (state is PointTransactionDeleteSuccess) {
            global.showSuccessSnackBar(context, global.language('delete_point_transaction_success'));
            // โหลดข้อมูลใหม่
            loadTransactions();
          }
          if (state is PointTransactionDeleteFailed) {
            global.showErrorSnackBar(context, '${global.language("delete_point_transaction_failed")}: ${state.message}');
          }
        },
        child: Column(
          children: [
            if (loadingData)
              Center(
                child: Padding(
                  padding: EdgeInsets.all(20),
                  child: LoadingAnimationWidget.staggeredDotsWave(
                    color: global.theme.infoHighlightTextColor,
                    size: 50,
                  ),
                ),
              ),
            if (!loadingData && transactions.isEmpty)
              Expanded(
                child: Center(
                  child: Text(
                    global.language(
                      'point_transaction.no_points_history_found',
                    ),
                    style: TextStyle(fontSize: 16, color: global.theme.iconSecondaryColor),
                  ),
                ),
              ),
            if (!loadingData && transactions.isNotEmpty)
              Expanded(
                child: ListView.builder(
                  itemCount: transactions.length,
                  itemBuilder: (context, index) {
                    return transactionListItem(transactions[index]);
                  },
                ),
              ),
          ],
        ),
      ),
    );
  }
}
