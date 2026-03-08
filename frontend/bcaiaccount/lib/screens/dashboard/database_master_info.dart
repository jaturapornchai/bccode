import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/screens/dashboard/cubit/database_master_info_cubit.dart';
import 'package:smlaicloud/global.dart' as global;

class DashBoardDatabaseMasterInfo extends StatefulWidget {
  const DashBoardDatabaseMasterInfo({super.key});

  @override
  _DashBoardDatabaseMasterInfoState createState() =>
      _DashBoardDatabaseMasterInfoState();
}

class _DashBoardDatabaseMasterInfoState
    extends State<DashBoardDatabaseMasterInfo> {
  @override
  Widget build(BuildContext context) {
    return BlocProvider(
      create: (context) => DatabaseMasterInfoCubit()..loadMasterDataCount(),
      child: Container(
        padding: EdgeInsets.all(16.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [
                Text(
                  global.language('database_master_info.data_by_type'),
                  style: const TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                BlocBuilder<DatabaseMasterInfoCubit, DatabaseMasterInfoState>(
                  builder: (context, state) {
                    return IconButton(
                      icon: const Icon(Icons.refresh),
                      onPressed: () {
                        context.read<DatabaseMasterInfoCubit>().refresh();
                      },
                      tooltip: global.language('database_master_info.refresh'),
                    );
                  },
                ),
              ],
            ),
            const SizedBox(height: 10),
            BlocBuilder<DatabaseMasterInfoCubit, DatabaseMasterInfoState>(
              builder: (context, state) {
                if (state is DatabaseMasterInfoLoading) {
                  return const Card(
                    child: ListTile(
                      leading: CircularProgressIndicator(),
                      title: Text('Loading master data...'),
                    ),
                  );
                }

                if (state is DatabaseMasterInfoError) {
                  return Card(
                    child: ListTile(
                      leading: const Icon(Icons.error, color: Colors.red),
                      title: Text(global.language('error')),
                      subtitle: Text(state.message),
                    ),
                  );
                }

                if (state is DatabaseMasterInfoLoaded) {
                  return Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Table(
                        border: TableBorder.all(color: Colors.grey[300]!),
                        columnWidths: const {
                          0: FixedColumnWidth(60),
                          1: FlexColumnWidth(3),
                          2: FlexColumnWidth(1),
                        },
                        children: [
                          // Header Row
                          TableRow(
                            decoration: BoxDecoration(color: Colors.blue[100]),
                            children: [
                              const Padding(
                                padding: EdgeInsets.all(4.0),
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
                                padding: EdgeInsets.all(4.0),
                                child: Text(
                                  global.language(
                                    'database_master_info.data_type',
                                  ),
                                  style: const TextStyle(
                                    fontWeight: FontWeight.bold,
                                    fontSize: 13,
                                  ),
                                ),
                              ),
                              Padding(
                                padding: EdgeInsets.all(4.0),
                                child: Text(
                                  global.language(
                                    'database_master_info.item_count',
                                  ),
                                  style: const TextStyle(
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
                            String tableName = item['table_name'] ?? '';

                            // แปลงชื่อประเภทข้อมูลโดยใช้ระบบแปลภาษา
                            String getDataTypeName(String tableName) {
                              switch (tableName) {
                                case 'product':
                                  return global.language(
                                    'database_master_info.product',
                                  );
                                case 'productbarcode':
                                  return global.language(
                                    'database_master_info.productbarcode',
                                  );
                                case 'customer':
                                  return global.language(
                                    'database_master_info.customer',
                                  );
                                case 'debtor':
                                  return global.language(
                                    'database_master_info.debtor',
                                  );
                                case 'creditor':
                                  return global.language(
                                    'database_master_info.creditor',
                                  );
                                case 'ic_warehouse':
                                  return global.language(
                                    'database_master_info.ic_warehouse',
                                  );
                                case 'ic_shelf':
                                  return global.language(
                                    'database_master_info.ic_shelf',
                                  );
                                case 'doc':
                                  return global.language(
                                    'database_master_info.doc',
                                  );
                                case 'docdetail':
                                  return global.language(
                                    'database_master_info.docdetail',
                                  );
                                default:
                                  return tableName;
                              }
                            }

                            return TableRow(
                              decoration: BoxDecoration(
                                color: index % 2 == 0
                                    ? Colors.white
                                    : Colors.grey[50],
                              ),
                              children: [
                                Padding(
                                  padding: const EdgeInsets.all(4.0),
                                  child: Text(
                                    '${index + 1}',
                                    style: TextStyle(
                                      color: Colors.grey[600],
                                      fontSize: 12,
                                    ),
                                    textAlign: TextAlign.center,
                                  ),
                                ),
                                Padding(
                                  padding: const EdgeInsets.all(4.0),
                                  child: Text(
                                    getDataTypeName(tableName),
                                    style: const TextStyle(fontSize: 13),
                                  ),
                                ),
                                Padding(
                                  padding: const EdgeInsets.all(4.0),
                                  child: Center(
                                    child: Container(
                                      padding: const EdgeInsets.symmetric(
                                        horizontal: 12,
                                        vertical: 4,
                                      ),
                                      decoration: BoxDecoration(
                                        color: Colors.blue[100],
                                        borderRadius: BorderRadius.circular(12),
                                      ),
                                      child: Text(
                                        global.formatNumberRemoveRightZero(
                                          item['xcount']?.toDouble() ?? 0.0,
                                        ),
                                        style: TextStyle(
                                          color: Colors.blue[900],
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
                      leading: const Icon(Icons.info, color: Colors.blue),
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
