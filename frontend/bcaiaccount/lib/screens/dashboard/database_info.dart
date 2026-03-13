import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/screens/dashboard/cubit/database_info_cubit.dart';
import 'package:smlaicloud/global.dart' as global;

class DashBoardDatabaseInfo extends StatefulWidget {
  const DashBoardDatabaseInfo({super.key});

  @override
  _DashBoardDatabaseInfoState createState() => _DashBoardDatabaseInfoState();
}

class _DashBoardDatabaseInfoState extends State<DashBoardDatabaseInfo> with global.ThemeRefreshMixin {
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
                  global.language('daily_info'),
                  style: TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                  ),
                ),
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
                  return Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Table(
                        border: TableBorder.all(color: global.theme.dividerBorderColor),
                        columnWidths: const {
                          0: FixedColumnWidth(60),
                          1: FlexColumnWidth(3),
                          2: FlexColumnWidth(1),
                        },
                        children: [
                          // Header Row
                          TableRow(
                            decoration: BoxDecoration(color: global.theme.columnHeaderColor),
                            children: [
                              const Padding(
                                padding: EdgeInsets.all(8.0),
                                child: Text(
                                  '#',
                                  style: TextStyle(
                                    fontWeight: FontWeight.bold,
                                    fontSize: 13,
                                  ),
                                  textAlign: TextAlign.center,
                                ),
                              ),
                              Padding(
                                padding: EdgeInsets.all(8.0),
                                child: Text(
                                  global.language('document_type'),
                                  style: TextStyle(
                                    fontWeight: FontWeight.bold,
                                    fontSize: 13,
                                  ),
                                ),
                              ),
                              Padding(
                                padding: EdgeInsets.all(8.0),
                                child: Text(
                                  global.language('item_count'),
                                  style: TextStyle(
                                    fontWeight: FontWeight.bold,
                                    fontSize: 13,
                                  ),
                                  textAlign: TextAlign.center,
                                ),
                              ),
                            ],
                          ),
                          // Data Rows
                          ...state.data.asMap().entries.map((entry) {
                            int index = entry.key;
                            Map<String, dynamic> item = entry.value;
                            int transFlag = item['trans_flag'];

                            return TableRow(
                              decoration: BoxDecoration(
                                color: index % 2 == 0
                                    ? global.theme.columnAlternateEvenColor
                                    : global.theme.columnAlternateOddColor,
                              ),
                              children: [
                                Padding(
                                  padding: const EdgeInsets.all(8.0),
                                  child: Text(
                                    '${index + 1}',
                                    style: TextStyle(
                                      color: global.theme.iconSecondaryColor,
                                      fontSize: 12,
                                    ),
                                    textAlign: TextAlign.center,
                                  ),
                                ),
                                Padding(
                                  padding: const EdgeInsets.all(8.0),
                                  child: Text(
                                    "${global.getTransFlagText(transFlag)} ($transFlag)",
                                    style: TextStyle(fontSize: 13),
                                  ),
                                ),
                                Padding(
                                  padding: const EdgeInsets.all(8.0),
                                  child: Center(
                                    child: Container(
                                      padding: const EdgeInsets.symmetric(
                                        horizontal: 12,
                                        vertical: 4,
                                      ),
                                      decoration: BoxDecoration(
                                        color: global.theme.columnHeaderColor.withValues(alpha: 0.5),
                                        borderRadius: BorderRadius.circular(12),
                                      ),
                                      child: Text(
                                        '${item['xcount']}',
                                        style: TextStyle(
                                          color: global.theme.columnHeaderTextColor,
                                          fontWeight: FontWeight.bold,
                                          fontSize: 12,
                                        ),
                                      ),
                                    ),
                                  ),
                                ),
                              ],
                            );
                          }),
                        ],
                      ),
                    ],
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
