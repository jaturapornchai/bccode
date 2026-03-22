import 'package:flutter/material.dart';
import 'package:smlaicloud/screens/dashboard/database_doc_info.dart';
import 'package:smlaicloud/screens/dashboard/database_master_info.dart';
import 'package:smlaicloud/global.dart' as global;

class DashBoardScreen extends StatefulWidget {
  const DashBoardScreen({super.key});

  @override
  _DashBoardScreenState createState() => _DashBoardScreenState();
}

class _DashBoardScreenState extends State<DashBoardScreen> with global.ThemeRefreshMixin {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(global.language('data_status')),
        backgroundColor: global.theme.appBarColor,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.all(4),
        child: Column(
          children: const [
            DashBoardDatabaseDocInfo(),
            DashBoardDatabaseMasterInfo(),
          ],
        ),
      ),
    );
  }
}
