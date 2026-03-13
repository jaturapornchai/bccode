import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/screens/dashboard/cubit/database_info_cubit.dart';
import 'package:smlaicloud/modules/cart-system/cartsystem.dart';
import 'package:smlaicloud/global.dart' as global;

class DashBoardDatabaseDocInfo extends StatefulWidget {
  const DashBoardDatabaseDocInfo({super.key});

  @override
  _DashBoardDatabaseDocInfoState createState() =>
      _DashBoardDatabaseDocInfoState();
}

class _DashBoardDatabaseDocInfoState extends State<DashBoardDatabaseDocInfo> with global.ThemeRefreshMixin {
  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (context) => DatabaseInfoCubit()..loadDocCountByTransFlag(),
      child: Container(
        padding: EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  global.language('daily_data'),
                  style: TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                Row(
                  children: [
                    // Cart System Icon
                    IconButton(
                      icon: Icon(
                        Icons.shopping_cart,
                        color: global.theme.warningHighlightTextColor,
                      ),
                      onPressed: () {
                        Navigator.push(
                          context,
                          MaterialPageRoute(
                            builder: (context) => const CartSystemScreen(),
                          ),
                        );
                      },
                      tooltip: global.language('shopping_cart_system'),
                    ),
                    // AI Chat Assistant Icon
                    IconButton(
                      icon: Icon(Icons.chat, color: global.theme.infoHighlightTextColor),
                      onPressed: () {
                        Navigator.pushNamed(context, '/ai-chat');
                      },
                      tooltip: global.language('ai_chat_assistant'),
                    ),
                    // Refresh Icon
                    BlocBuilder<DatabaseInfoCubit, DatabaseInfoState>(
                      builder: (context, state) {
                        return IconButton(
                          icon: Icon(Icons.refresh),
                          onPressed: () {
                            context.read<DatabaseInfoCubit>().refresh();
                          },
                          tooltip: global.language('refresh'),
                        );
                      },
                    ),
                  ],
                ),
              ],
            ),
            const SizedBox(height: 10),
            BlocBuilder<DatabaseInfoCubit, DatabaseInfoState>(
              builder: (context, state) {
                if (state is DatabaseInfoLoading) {
                  return const Card(
                    child: ListTile(
                      leading: CircularProgressIndicator(),
                      title: Text('Loading tables...'),
                    ),
                  );
                }

                if (state is DatabaseInfoError) {
                  return Card(
                    child: ListTile(
                      leading: Icon(Icons.error, color: global.theme.negativeHighlightTextColor),
                      title: Text(global.language('error')),
                      subtitle: Text(state.message),
                    ),
                  );
                }

                if (state is DatabaseInfoLoaded) {
                  return SizedBox(
                    width: double.infinity,
                    child: Table(
                      border: TableBorder.all(color: global.theme.dividerBorderColor),
                      columnWidths: const {
                        0: FlexColumnWidth(0.5),
                        1: FlexColumnWidth(2),
                        2: FlexColumnWidth(1),
                        3: FlexColumnWidth(1.5),
                        4: FlexColumnWidth(1.2),
                        5: FlexColumnWidth(1.5),
                      },
                      children: [
                        // Header Row
                        TableRow(
                          decoration: BoxDecoration(color: global.theme.columnHeaderColor),
                          children: [
                            const Padding(
                              padding: EdgeInsets.all(4.0),
                              child: Text(
                                '#',
                                style: TextStyle(
                                  fontWeight: FontWeight.bold,
                                  fontSize: 12,
                                ),
                                textAlign: TextAlign.center,
                              ),
                            ),
                            Padding(
                              padding: EdgeInsets.all(4.0),
                              child: Text(
                                global.language('doc_type'),
                                style: TextStyle(
                                  fontWeight: FontWeight.bold,
                                  fontSize: 12,
                                ),
                              ),
                            ),
                            Padding(
                              padding: EdgeInsets.all(4.0),
                              child: Text(
                                global.language('quantity'),
                                style: TextStyle(
                                  fontWeight: FontWeight.bold,
                                  fontSize: 12,
                                ),
                                textAlign: TextAlign.center,
                              ),
                            ),
                            Padding(
                              padding: EdgeInsets.all(4.0),
                              child: Text(
                                global.language('latest_number'),
                                style: TextStyle(
                                  fontWeight: FontWeight.bold,
                                  fontSize: 12,
                                ),
                                textAlign: TextAlign.center,
                              ),
                            ),
                            Padding(
                              padding: EdgeInsets.all(4.0),
                              child: Text(
                                global.language('latest_date'),
                                style: TextStyle(
                                  fontWeight: FontWeight.bold,
                                  fontSize: 12,
                                ),
                                textAlign: TextAlign.center,
                              ),
                            ),
                            Padding(
                              padding: EdgeInsets.all(4.0),
                              child: Text(
                                global.language('document_value'),
                                style: TextStyle(
                                  fontWeight: FontWeight.bold,
                                  fontSize: 12,
                                ),
                                textAlign: TextAlign.right,
                              ),
                            ),
                          ],
                        ),
                        // Data Rows
                        ...state.data.asMap().entries.map((entry) {
                          int index = entry.key;
                          Map<String, dynamic> item = entry.value;
                          int transFlag = item['trans_flag'];
                          String? lastDocDate = item['last_doc_date'];
                          String? lastDocNo = item['last_doc_no'];
                          double totalAmount = item['total_amount'] ?? 0.0;

                          // Format date in Thai format with time
                          String formattedDate = '-';
                          if (lastDocDate != null && lastDocDate.isNotEmpty) {
                            try {
                              final dateTime = DateTime.parse(lastDocDate);
                              formattedDate = global.formatThaiDateTime(
                                dateTime: dateTime,
                                showTime: true,
                              );
                            } catch (e) {
                              formattedDate = lastDocDate.split(' ')[0];
                            }
                          }

                          return TableRow(
                            decoration: BoxDecoration(
                              color: index % 2 == 0
                                  ? global.theme.columnAlternateEvenColor
                                  : global.theme.columnAlternateOddColor,
                            ),
                            children: [
                              Padding(
                                padding: const EdgeInsets.all(4.0),
                                child: Text(
                                  '${index + 1}',
                                  style: TextStyle(
                                    color: global.theme.iconSecondaryColor,
                                    fontSize: 11,
                                  ),
                                  textAlign: TextAlign.center,
                                ),
                              ),
                              Padding(
                                padding: const EdgeInsets.all(4.0),
                                child: Text(
                                  "${global.getTransFlagText(transFlag)} ($transFlag)",
                                  style: TextStyle(fontSize: 11),
                                ),
                              ),
                              Padding(
                                padding: const EdgeInsets.all(4.0),
                                child: Center(
                                  child: Container(
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 8,
                                      vertical: 4,
                                    ),
                                    decoration: BoxDecoration(
                                      color: global.theme.columnHeaderColor.withValues(alpha: 0.5),
                                      borderRadius: BorderRadius.circular(12),
                                    ),
                                    child: Text(
                                      global.formatNumberRemoveRightZero(
                                        item['xcount']?.toDouble() ?? 0.0,
                                      ),
                                      style: TextStyle(
                                        color: global.theme.columnHeaderTextColor,
                                        fontWeight: FontWeight.bold,
                                        fontSize: 11,
                                      ),
                                    ),
                                  ),
                                ),
                              ),
                              Padding(
                                padding: const EdgeInsets.all(4.0),
                                child: Text(
                                  lastDocNo ?? '-',
                                  style: TextStyle(
                                    fontSize: 11,
                                    fontFamily: 'monospace',
                                  ),
                                  textAlign: TextAlign.center,
                                ),
                              ),
                              Padding(
                                padding: const EdgeInsets.all(4.0),
                                child: Text(
                                  formattedDate,
                                  style: TextStyle(fontSize: 11),
                                  textAlign: TextAlign.center,
                                ),
                              ),
                              Padding(
                                padding: const EdgeInsets.all(4.0),
                                child: Text(
                                  global.formatNumber(totalAmount),
                                  style: TextStyle(
                                    fontSize: 11,
                                    fontWeight: FontWeight.bold,
                                    color: totalAmount > 0
                                        ? (global.isDarkMode() ? global.theme.positiveHighlightTextColor : Colors.green[700])
                                        : global.theme.textSecondaryColor,
                                  ),
                                  textAlign: TextAlign.right,
                                ),
                              ),
                            ],
                          );
                        }),
                      ],
                    ),
                  );
                } else {
                  return Card(
                    child: ListTile(
                      leading: Icon(Icons.info, color: global.theme.infoHighlightTextColor),
                      title: Text(global.language('no_data_available')),
                    ),
                  );
                }
              },
            ),
          ],
        ),
      ),
    );
  }
}
