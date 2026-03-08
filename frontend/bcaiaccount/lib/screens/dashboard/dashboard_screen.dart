import 'package:flutter/material.dart';
import 'package:smlaicloud/screens/dashboard/database_doc_info.dart';
import 'package:smlaicloud/screens/dashboard/database_master_info.dart';

class DashBoardScreen extends StatefulWidget {
  const DashBoardScreen({super.key});

  @override
  _DashBoardScreenState createState() => _DashBoardScreenState();
}

class _DashBoardScreenState extends State<DashBoardScreen> {
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text(
          'Dashboard',
          style: TextStyle(
            fontSize: 18,
            color: Colors.white,
            fontWeight: FontWeight.bold,
            letterSpacing: 0.5,
            shadows: [
              Shadow(
                offset: Offset(1.5, 1.5),
                blurRadius: 3.0,
                color: Color.fromARGB(150, 0, 0, 0),
              ),
            ],
          ),
        ),
        backgroundColor: const Color(0xFF0A3880),
        iconTheme: const IconThemeData(color: Colors.white),
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
