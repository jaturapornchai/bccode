// ignore_for_file: unnecessary_const, avoid_unnecessary_containers, sized_box_for_whitespace

import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

class ReceiveList extends StatefulWidget {
  const ReceiveList({super.key});

  @override
  State<ReceiveList> createState() => _ReceiveListState();
}

class _ReceiveListState extends State<ReceiveList> {
  TextEditingController txtSearch = TextEditingController();

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: Color.fromRGBO(241, 241, 255, 1),
      appBar: AppBar(
        backgroundColor: Colors.white,
        centerTitle: true,
        leading: IconButton(
          color: const Color.fromRGBO(121, 130, 142, 1),
          icon: const Icon(Icons.arrow_back),
          onPressed: () {
            // Navigator.of(context).pushAndRemoveUntil(
            //     MaterialPageRoute(builder: (_) => MenuListLandscape()),
            //     (route) => false);
          },
        ),
        title: Text(
          global.language('receivelist_title'),
          style: const TextStyle(color: Color.fromRGBO(121, 130, 142, 1)),
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () {},
        backgroundColor: Colors.blue,
        child: const Icon(Icons.add),
      ),
      body: Column(
        children: <Widget>[
          _searchInput(),
          SizedBox(height: 10),
          Expanded(
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisAlignment: MainAxisAlignment.start,
              children: <Widget>[_docList(), _detailBox()],
            ),
          ),
        ],
      ),
    );
  }

  Widget _detailBox() {
    return Container(
      width: (MediaQuery.of(context).size.width / 100) * 35,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            flex: 2,
            child: Container(
              padding: EdgeInsets.only(right: 8),
              child: Card(
                elevation: 5.0,
                color: Colors.white,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Container(
                  margin: EdgeInsets.all(20.0),
                  child: Column(
                    children: <Widget>[
                      Expanded(
                        child: Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Expanded(
                              flex: 1,
                              child: Text(
                                global.language('receivelist_document_number'),
                              ),
                            ),
                            const Expanded(
                              flex: 1,
                              child: Text("RE-202202021001"),
                            ),
                          ],
                        ),
                      ),
                      Expanded(
                        child: Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Expanded(
                              flex: 1,
                              child: Text(global.language('receivelist_date')),
                            ),
                            const Expanded(flex: 1, child: Text("10/04/2565")),
                          ],
                        ),
                      ),
                      Expanded(
                        child: Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Expanded(
                              flex: 1,
                              child: Text(
                                global.language('receivelist_transaction_type'),
                              ),
                            ),
                            Expanded(
                              flex: 1,
                              child: Text(
                                global.language('receivelist_receive_goods'),
                              ),
                            ),
                          ],
                        ),
                      ),
                      Expanded(
                        child: Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Expanded(
                              flex: 1,
                              child: Text(
                                global.language('receivelist_member'),
                              ),
                            ),
                            const Expanded(
                              flex: 1,
                              child: Text("สิธิพัช สามวง"),
                            ),
                          ],
                        ),
                      ),
                      Expanded(
                        child: Row(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Expanded(
                              flex: 1,
                              child: Text(
                                global.language('receivelist_employee'),
                              ),
                            ),
                            const Expanded(
                              flex: 1,
                              child: Text("เบอรรี่ ดูดี"),
                            ),
                          ],
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
          Expanded(
            flex: 3,
            child: Container(
              padding: const EdgeInsets.only(top: 10, right: 8),
              child: Card(
                elevation: 10.0,
                color: Colors.white,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(10),
                ),
                child: SingleChildScrollView(
                  child: Container(
                    padding: const EdgeInsets.only(left: 15.0, right: 15.0),
                    child: Column(
                      children: <Widget>[
                        _tableDetailHeader(),
                        _tableDetailDetail(),
                      ],
                    ),
                  ),
                ),
              ),
            ),
          ),
          Expanded(
            flex: 1,
            child: Container(
              padding: EdgeInsets.only(top: 10, right: 8),
              child: Card(
                elevation: 10.0,
                color: Colors.white,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.center,
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Expanded(
                      flex: 1,
                      child: Container(
                        margin: EdgeInsets.only(left: 15),
                        child: Text(
                          global.language('receivelist_total_value'),
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
                    ),
                    Expanded(
                      flex: 1,
                      child: Container(
                        alignment: Alignment.centerRight,
                        margin: EdgeInsets.only(right: 15),
                        child: Text(
                          "1,800บาท",
                          style: TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
          const Expanded(flex: 1, child: SizedBox()),
        ],
      ),
    );
  }

  Table _tableDetailHeader() {
    return Table(
      border: const TableBorder(
        bottom: const BorderSide(
          width: 1,
          color: Colors.grey,
          style: BorderStyle.solid,
        ),
      ),
      defaultVerticalAlignment: TableCellVerticalAlignment.middle,
      columnWidths: const <int, TableColumnWidth>{
        0: FlexColumnWidth(2),
        1: FlexColumnWidth(1),
        2: FlexColumnWidth(1),
      },
      children: <TableRow>[
        TableRow(
          children: <Widget>[
            Container(
              height: 45,
              child: Container(
                alignment: Alignment.centerLeft,
                child: Text(
                  global.language('receivelist_item'),
                  style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
                ),
              ),
            ),
            Center(
              child: Container(
                child: Text(
                  global.language('receivelist_quantity'),
                  style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
                ),
              ),
            ),
            Container(
              alignment: Alignment.centerRight,
              child: Text(
                global.language('receivelist_value'),
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _tableDetailDetail() {
    return Table(
      defaultVerticalAlignment: TableCellVerticalAlignment.middle,
      columnWidths: const <int, TableColumnWidth>{
        0: FlexColumnWidth(2),
        1: FlexColumnWidth(1),
        2: FlexColumnWidth(1),
      },
      children: <TableRow>[
        TableRow(
          children: <Widget>[
            Container(
              height: 35,
              child: Container(
                alignment: Alignment.centerLeft,
                child: const Text(
                  "บะหมี่กึ่งสำเร็จรูป",
                  style: TextStyle(fontSize: 14),
                ),
              ),
            ),
            Container(
              alignment: Alignment.center,
              child: const Text("10", style: TextStyle(fontSize: 14)),
            ),
            Container(
              alignment: Alignment.centerRight,
              child: const Text("700บาท", style: TextStyle(fontSize: 14)),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 35,
              child: Container(
                alignment: Alignment.centerLeft,
                child: const Text(
                  "ซอสถั่วเหลือง",
                  style: TextStyle(fontSize: 14),
                ),
              ),
            ),
            Container(
              alignment: Alignment.center,
              child: const Text("10", style: TextStyle(fontSize: 14)),
            ),
            Container(
              alignment: Alignment.centerRight,
              child: const Text("600บาท", style: TextStyle(fontSize: 14)),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 35,
              child: Container(
                alignment: Alignment.centerLeft,
                child: const Text("ผงซักฟอก", style: TextStyle(fontSize: 14)),
              ),
            ),
            Container(
              alignment: Alignment.center,
              child: const Text("10", style: TextStyle(fontSize: 14)),
            ),
            Container(
              alignment: Alignment.centerRight,
              child: const Text("500บาท", style: TextStyle(fontSize: 14)),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 35,
              child: Container(
                alignment: Alignment.centerLeft,
                child: const Text(
                  "บะหมี่กึ่งสำเร็จรูป",
                  style: TextStyle(fontSize: 14),
                ),
              ),
            ),
            Container(
              alignment: Alignment.center,
              child: const Text("10", style: TextStyle(fontSize: 14)),
            ),
            Container(
              alignment: Alignment.centerRight,
              child: const Text("700บาท", style: TextStyle(fontSize: 14)),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 35,
              child: Container(
                alignment: Alignment.centerLeft,
                child: const Text(
                  "ซอสถั่วเหลือง",
                  style: TextStyle(fontSize: 14),
                ),
              ),
            ),
            Container(
              alignment: Alignment.center,
              child: const Text("10", style: TextStyle(fontSize: 14)),
            ),
            Container(
              alignment: Alignment.centerRight,
              child: const Text("600บาท", style: TextStyle(fontSize: 14)),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 35,
              child: Container(
                alignment: Alignment.centerLeft,
                child: const Text("ผงซักฟอก", style: TextStyle(fontSize: 14)),
              ),
            ),
            Container(
              alignment: Alignment.center,
              child: const Text("10", style: TextStyle(fontSize: 14)),
            ),
            Container(
              alignment: Alignment.centerRight,
              child: const Text("500บาท", style: TextStyle(fontSize: 14)),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 35,
              child: Container(
                alignment: Alignment.centerLeft,
                child: const Text(
                  "บะหมี่กึ่งสำเร็จรูป",
                  style: TextStyle(fontSize: 14),
                ),
              ),
            ),
            Container(
              alignment: Alignment.center,
              child: const Text("10", style: TextStyle(fontSize: 14)),
            ),
            Container(
              alignment: Alignment.centerRight,
              child: const Text("700บาท", style: TextStyle(fontSize: 14)),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 35,
              child: Container(
                alignment: Alignment.centerLeft,
                child: const Text(
                  "ซอสถั่วเหลือง",
                  style: TextStyle(fontSize: 14),
                ),
              ),
            ),
            Container(
              alignment: Alignment.center,
              child: const Text("10", style: TextStyle(fontSize: 14)),
            ),
            Container(
              alignment: Alignment.centerRight,
              child: const Text("600บาท", style: TextStyle(fontSize: 14)),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 35,
              child: Container(
                alignment: Alignment.centerLeft,
                child: const Text("ผงซักฟอก", style: TextStyle(fontSize: 14)),
              ),
            ),
            Container(
              alignment: Alignment.center,
              child: const Text("10", style: TextStyle(fontSize: 14)),
            ),
            Container(
              alignment: Alignment.centerRight,
              child: const Text("500บาท", style: TextStyle(fontSize: 14)),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 35,
              child: Container(
                alignment: Alignment.centerLeft,
                child: const Text(
                  "บะหมี่กึ่งสำเร็จรูป",
                  style: TextStyle(fontSize: 14),
                ),
              ),
            ),
            Container(
              alignment: Alignment.center,
              child: const Text("10", style: TextStyle(fontSize: 14)),
            ),
            Container(
              alignment: Alignment.centerRight,
              child: const Text("700บาท", style: TextStyle(fontSize: 14)),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 35,
              child: Container(
                alignment: Alignment.centerLeft,
                child: const Text(
                  "ซอสถั่วเหลือง",
                  style: TextStyle(fontSize: 14),
                ),
              ),
            ),
            Container(
              alignment: Alignment.center,
              child: const Text("10", style: TextStyle(fontSize: 14)),
            ),
            Container(
              alignment: Alignment.centerRight,
              child: const Text("600บาท", style: TextStyle(fontSize: 14)),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 35,
              child: Container(
                alignment: Alignment.centerLeft,
                child: const Text("ผงซักฟอก", style: TextStyle(fontSize: 14)),
              ),
            ),
            Container(
              alignment: Alignment.center,
              child: const Text("10", style: TextStyle(fontSize: 14)),
            ),
            Container(
              alignment: Alignment.centerRight,
              child: const Text("500บาท", style: TextStyle(fontSize: 14)),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 35,
              child: Container(
                alignment: Alignment.centerLeft,
                child: const Text(
                  "บะหมี่กึ่งสำเร็จรูป",
                  style: TextStyle(fontSize: 14),
                ),
              ),
            ),
            Container(
              alignment: Alignment.center,
              child: const Text("10", style: TextStyle(fontSize: 14)),
            ),
            Container(
              alignment: Alignment.centerRight,
              child: const Text("700บาท", style: TextStyle(fontSize: 14)),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 35,
              child: Container(
                alignment: Alignment.centerLeft,
                child: const Text(
                  "ซอสถั่วเหลือง",
                  style: TextStyle(fontSize: 14),
                ),
              ),
            ),
            Container(
              alignment: Alignment.center,
              child: const Text("10", style: TextStyle(fontSize: 14)),
            ),
            Container(
              alignment: Alignment.centerRight,
              child: const Text("600บาท", style: TextStyle(fontSize: 14)),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 35,
              child: Container(
                alignment: Alignment.centerLeft,
                child: const Text("ผงซักฟอก", style: TextStyle(fontSize: 14)),
              ),
            ),
            Container(
              alignment: Alignment.center,
              child: const Text("10", style: TextStyle(fontSize: 14)),
            ),
            Container(
              alignment: Alignment.centerRight,
              child: const Text("500บาท", style: TextStyle(fontSize: 14)),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 35,
              child: Container(
                alignment: Alignment.centerLeft,
                child: const Text(
                  "บะหมี่กึ่งสำเร็จรูป",
                  style: TextStyle(fontSize: 14),
                ),
              ),
            ),
            Container(
              alignment: Alignment.center,
              child: const Text("10", style: TextStyle(fontSize: 14)),
            ),
            Container(
              alignment: Alignment.centerRight,
              child: const Text("700บาท", style: TextStyle(fontSize: 14)),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 35,
              child: Container(
                alignment: Alignment.centerLeft,
                child: const Text(
                  "ซอสถั่วเหลือง",
                  style: TextStyle(fontSize: 14),
                ),
              ),
            ),
            Container(
              alignment: Alignment.center,
              child: const Text("10", style: TextStyle(fontSize: 14)),
            ),
            Container(
              alignment: Alignment.centerRight,
              child: const Text("600บาท", style: TextStyle(fontSize: 14)),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 35,
              child: Container(
                alignment: Alignment.centerLeft,
                child: const Text("ผงซักฟอก", style: TextStyle(fontSize: 14)),
              ),
            ),
            Container(
              alignment: Alignment.center,
              child: const Text("10", style: TextStyle(fontSize: 14)),
            ),
            Container(
              alignment: Alignment.centerRight,
              child: const Text("500บาท", style: TextStyle(fontSize: 14)),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 35,
              child: Container(
                alignment: Alignment.centerLeft,
                child: const Text(
                  "บะหมี่กึ่งสำเร็จรูป",
                  style: TextStyle(fontSize: 14),
                ),
              ),
            ),
            Container(
              alignment: Alignment.center,
              child: const Text("10", style: TextStyle(fontSize: 14)),
            ),
            Container(
              alignment: Alignment.centerRight,
              child: const Text("700บาท", style: TextStyle(fontSize: 14)),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 35,
              child: Container(
                alignment: Alignment.centerLeft,
                child: const Text(
                  "ซอสถั่วเหลือง",
                  style: TextStyle(fontSize: 14),
                ),
              ),
            ),
            Container(
              alignment: Alignment.center,
              child: const Text("10", style: TextStyle(fontSize: 14)),
            ),
            Container(
              alignment: Alignment.centerRight,
              child: const Text("600บาท", style: TextStyle(fontSize: 14)),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 35,
              child: Container(
                alignment: Alignment.centerLeft,
                child: const Text("ผงซักฟอก", style: TextStyle(fontSize: 14)),
              ),
            ),
            Container(
              alignment: Alignment.center,
              child: const Text("10", style: TextStyle(fontSize: 14)),
            ),
            Container(
              alignment: Alignment.centerRight,
              child: const Text("500บาท", style: TextStyle(fontSize: 14)),
            ),
          ],
        ),
      ],
    );
  }

  Widget _docList() {
    return Expanded(
      child: SingleChildScrollView(
        child: Container(
          padding: const EdgeInsets.only(left: 8.0, right: 8.0),
          child: Card(
            elevation: 5.0,
            color: Colors.white,
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(10),
            ),
            child: Column(
              children: <Widget>[
                Padding(
                  padding: const EdgeInsets.only(left: 8.0, right: 8.0),
                  child: _tableHeader(),
                ),
                Padding(
                  padding: const EdgeInsets.all(8.0),
                  child: _tableDetail(),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  Widget _tableHeader() {
    return Table(
      border: const TableBorder(
        bottom: const BorderSide(
          width: 1,
          color: Colors.grey,
          style: BorderStyle.solid,
        ),
      ),
      defaultVerticalAlignment: TableCellVerticalAlignment.middle,
      defaultColumnWidth: const FlexColumnWidth(),
      children: <TableRow>[
        TableRow(
          children: <Widget>[
            Container(
              height: 45,
              child: Center(
                child: Container(
                  child: Text(
                    global.language('receivelist_document_number'),
                    style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
                  ),
                ),
              ),
            ),
            Center(
              child: Container(
                child: Text(
                  global.language('receivelist_date'),
                  style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
                ),
              ),
            ),
            Center(
              child: Container(
                child: Text(
                  global.language('receivelist_member_name'),
                  style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
                ),
              ),
            ),
            Center(
              child: Container(
                child: Text(
                  global.language('receivelist_employee'),
                  style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
                ),
              ),
            ),
            Center(
              child: Container(
                child: Text(
                  global.language('receivelist_type'),
                  style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
                ),
              ),
            ),
            Center(
              child: Container(
                child: Text(
                  global.language('receivelist_grand_total'),
                  style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
                ),
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _tableDetail() {
    return Table(
      border: const TableBorder(
        horizontalInside: const BorderSide(
          width: 1,
          color: Colors.grey,
          style: BorderStyle.solid,
        ),
      ),
      defaultVerticalAlignment: TableCellVerticalAlignment.middle,
      defaultColumnWidth: const FlexColumnWidth(),
      children: <TableRow>[
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021001",
                    textAlign: TextAlign.left,
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "10/04/2565 21:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text(
                "สิธิพัช สามวง",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("เบอรรี่ ดูดี", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: Text(
                global.language('receivelist_receive_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ),
            Container(
              child: const Text(
                "1,800.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021002",
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "15/04/2565 12:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("วงเหลา คำปลา", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: const Text(
                "สมใจ เนื่องแป้น",
                style: const TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: Text(
                global.language('receivelist_issue_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            Container(
              child: const Text(
                "900.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021001",
                    textAlign: TextAlign.left,
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "10/04/2565 21:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text(
                "สิธิพัช สามวง",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("เบอรรี่ ดูดี", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: Text(
                global.language('receivelist_receive_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ),
            Container(
              child: const Text(
                "1,800.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021002",
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "15/04/2565 12:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("วงเหลา คำปลา", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: const Text(
                "สมใจ เนื่องแป้น",
                style: const TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: Text(
                global.language('receivelist_issue_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            Container(
              child: const Text(
                "900.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021001",
                    textAlign: TextAlign.left,
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "10/04/2565 21:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text(
                "สิธิพัช สามวง",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("เบอรรี่ ดูดี", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: Text(
                global.language('receivelist_receive_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ),
            Container(
              child: const Text(
                "1,800.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021002",
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "15/04/2565 12:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("วงเหลา คำปลา", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: const Text(
                "สมใจ เนื่องแป้น",
                style: const TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: Text(
                global.language('receivelist_issue_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            Container(
              child: const Text(
                "900.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021001",
                    textAlign: TextAlign.left,
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "10/04/2565 21:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text(
                "สิธิพัช สามวง",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("เบอรรี่ ดูดี", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: Text(
                global.language('receivelist_receive_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ),
            Container(
              child: const Text(
                "1,800.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021002",
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "15/04/2565 12:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("วงเหลา คำปลา", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: const Text(
                "สมใจ เนื่องแป้น",
                style: const TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: Text(
                global.language('receivelist_issue_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            Container(
              child: const Text(
                "900.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021001",
                    textAlign: TextAlign.left,
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "10/04/2565 21:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text(
                "สิธิพัช สามวง",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("เบอรรี่ ดูดี", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: Text(
                global.language('receivelist_receive_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ),
            Container(
              child: const Text(
                "1,800.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021002",
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "15/04/2565 12:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("วงเหลา คำปลา", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: const Text(
                "สมใจ เนื่องแป้น",
                style: const TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: Text(
                global.language('receivelist_issue_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            Container(
              child: const Text(
                "900.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021001",
                    textAlign: TextAlign.left,
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "10/04/2565 21:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text(
                "สิธิพัช สามวง",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("เบอรรี่ ดูดี", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: Text(
                global.language('receivelist_receive_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ),
            Container(
              child: const Text(
                "1,800.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021002",
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "15/04/2565 12:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("วงเหลา คำปลา", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: const Text(
                "สมใจ เนื่องแป้น",
                style: const TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: Text(
                global.language('receivelist_issue_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            Container(
              child: const Text(
                "900.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021001",
                    textAlign: TextAlign.left,
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "10/04/2565 21:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text(
                "สิธิพัช สามวง",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("เบอรรี่ ดูดี", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: Text(
                global.language('receivelist_receive_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ),
            Container(
              child: const Text(
                "1,800.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021002",
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "15/04/2565 12:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("วงเหลา คำปลา", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: const Text(
                "สมใจ เนื่องแป้น",
                style: const TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: Text(
                global.language('receivelist_issue_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            Container(
              child: const Text(
                "900.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021001",
                    textAlign: TextAlign.left,
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "10/04/2565 21:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text(
                "สิธิพัช สามวง",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("เบอรรี่ ดูดี", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: Text(
                global.language('receivelist_receive_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ),
            Container(
              child: const Text(
                "1,800.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021002",
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "15/04/2565 12:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("วงเหลา คำปลา", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: const Text(
                "สมใจ เนื่องแป้น",
                style: const TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: Text(
                global.language('receivelist_issue_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            Container(
              child: const Text(
                "900.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021001",
                    textAlign: TextAlign.left,
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "10/04/2565 21:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text(
                "สิธิพัช สามวง",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("เบอรรี่ ดูดี", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: Text(
                global.language('receivelist_receive_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ),
            Container(
              child: const Text(
                "1,800.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021002",
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "15/04/2565 12:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("วงเหลา คำปลา", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: const Text(
                "สมใจ เนื่องแป้น",
                style: const TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: Text(
                global.language('receivelist_issue_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            Container(
              child: const Text(
                "900.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021001",
                    textAlign: TextAlign.left,
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "10/04/2565 21:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text(
                "สิธิพัช สามวง",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("เบอรรี่ ดูดี", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: Text(
                global.language('receivelist_receive_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ),
            Container(
              child: const Text(
                "1,800.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021002",
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "15/04/2565 12:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("วงเหลา คำปลา", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: const Text(
                "สมใจ เนื่องแป้น",
                style: const TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: Text(
                global.language('receivelist_issue_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            Container(
              child: const Text(
                "900.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021001",
                    textAlign: TextAlign.left,
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "10/04/2565 21:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text(
                "สิธิพัช สามวง",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("เบอรรี่ ดูดี", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: Text(
                global.language('receivelist_receive_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ),
            Container(
              child: const Text(
                "1,800.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021002",
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "15/04/2565 12:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("วงเหลา คำปลา", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: const Text(
                "สมใจ เนื่องแป้น",
                style: const TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: Text(
                global.language('receivelist_issue_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            Container(
              child: const Text(
                "900.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021001",
                    textAlign: TextAlign.left,
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "10/04/2565 21:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text(
                "สิธิพัช สามวง",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("เบอรรี่ ดูดี", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: Text(
                global.language('receivelist_receive_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ),
            Container(
              child: const Text(
                "1,800.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021002",
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "15/04/2565 12:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("วงเหลา คำปลา", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: const Text(
                "สมใจ เนื่องแป้น",
                style: const TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: Text(
                global.language('receivelist_issue_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            Container(
              child: const Text(
                "900.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021001",
                    textAlign: TextAlign.left,
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "10/04/2565 21:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text(
                "สิธิพัช สามวง",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("เบอรรี่ ดูดี", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: Text(
                global.language('receivelist_receive_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ),
            Container(
              child: const Text(
                "1,800.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021002",
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "15/04/2565 12:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("วงเหลา คำปลา", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: const Text(
                "สมใจ เนื่องแป้น",
                style: const TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: Text(
                global.language('receivelist_issue_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            Container(
              child: const Text(
                "900.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021001",
                    textAlign: TextAlign.left,
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "10/04/2565 21:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text(
                "สิธิพัช สามวง",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("เบอรรี่ ดูดี", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: Text(
                global.language('receivelist_receive_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ),
            Container(
              child: const Text(
                "1,800.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021002",
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "15/04/2565 12:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("วงเหลา คำปลา", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: const Text(
                "สมใจ เนื่องแป้น",
                style: const TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: Text(
                global.language('receivelist_issue_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            Container(
              child: const Text(
                "900.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021001",
                    textAlign: TextAlign.left,
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "10/04/2565 21:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text(
                "สิธิพัช สามวง",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("เบอรรี่ ดูดี", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: Text(
                global.language('receivelist_receive_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ),
            Container(
              child: const Text(
                "1,800.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021002",
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "15/04/2565 12:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("วงเหลา คำปลา", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: const Text(
                "สมใจ เนื่องแป้น",
                style: const TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: Text(
                global.language('receivelist_issue_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            Container(
              child: const Text(
                "900.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021001",
                    textAlign: TextAlign.left,
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "10/04/2565 21:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text(
                "สิธิพัช สามวง",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("เบอรรี่ ดูดี", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: Text(
                global.language('receivelist_receive_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ),
            Container(
              child: const Text(
                "1,800.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021002",
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "15/04/2565 12:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("วงเหลา คำปลา", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: const Text(
                "สมใจ เนื่องแป้น",
                style: const TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: Text(
                global.language('receivelist_issue_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            Container(
              child: const Text(
                "900.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021001",
                    textAlign: TextAlign.left,
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "10/04/2565 21:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text(
                "สิธิพัช สามวง",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("เบอรรี่ ดูดี", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: Text(
                global.language('receivelist_receive_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ),
            Container(
              child: const Text(
                "1,800.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021002",
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "15/04/2565 12:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("วงเหลา คำปลา", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: const Text(
                "สมใจ เนื่องแป้น",
                style: const TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: Text(
                global.language('receivelist_issue_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            Container(
              child: const Text(
                "900.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021001",
                    textAlign: TextAlign.left,
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "10/04/2565 21:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text(
                "สิธิพัช สามวง",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("เบอรรี่ ดูดี", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: Text(
                global.language('receivelist_receive_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ),
            Container(
              child: const Text(
                "1,800.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021002",
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "15/04/2565 12:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("วงเหลา คำปลา", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: const Text(
                "สมใจ เนื่องแป้น",
                style: const TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: Text(
                global.language('receivelist_issue_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            Container(
              child: const Text(
                "900.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021001",
                    textAlign: TextAlign.left,
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "10/04/2565 21:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text(
                "สิธิพัช สามวง",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("เบอรรี่ ดูดี", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: Text(
                global.language('receivelist_receive_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ),
            Container(
              child: const Text(
                "1,800.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021002",
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "15/04/2565 12:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("วงเหลา คำปลา", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: const Text(
                "สมใจ เนื่องแป้น",
                style: const TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: Text(
                global.language('receivelist_issue_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            Container(
              child: const Text(
                "900.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021001",
                    textAlign: TextAlign.left,
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "10/04/2565 21:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text(
                "สิธิพัช สามวง",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("เบอรรี่ ดูดี", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: Text(
                global.language('receivelist_receive_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ),
            Container(
              child: const Text(
                "1,800.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021002",
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "15/04/2565 12:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("วงเหลา คำปลา", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: const Text(
                "สมใจ เนื่องแป้น",
                style: const TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: Text(
                global.language('receivelist_issue_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            Container(
              child: const Text(
                "900.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021001",
                    textAlign: TextAlign.left,
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "10/04/2565 21:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text(
                "สิธิพัช สามวง",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("เบอรรี่ ดูดี", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: Text(
                global.language('receivelist_receive_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ),
            Container(
              child: const Text(
                "1,800.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021002",
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "15/04/2565 12:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("วงเหลา คำปลา", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: const Text(
                "สมใจ เนื่องแป้น",
                style: const TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: Text(
                global.language('receivelist_issue_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            Container(
              child: const Text(
                "900.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021001",
                    textAlign: TextAlign.left,
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "10/04/2565 21:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text(
                "สิธิพัช สามวง",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("เบอรรี่ ดูดี", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: Text(
                global.language('receivelist_receive_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.green,
                ),
              ),
            ),
            Container(
              child: const Text(
                "1,800.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
        TableRow(
          children: <Widget>[
            Container(
              height: 40,
              child: Center(
                child: Container(
                  child: const Text(
                    "RE-202202021002",
                    style: TextStyle(fontSize: 14),
                  ),
                ),
              ),
            ),
            Container(
              child: const Text(
                "15/04/2565 12:47",
                textAlign: TextAlign.left,
                style: TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: const Text("วงเหลา คำปลา", style: TextStyle(fontSize: 14)),
            ),
            Container(
              child: const Text(
                "สมใจ เนื่องแป้น",
                style: const TextStyle(fontSize: 14),
              ),
            ),
            Container(
              child: Text(
                global.language('receivelist_issue_goods'),
                textAlign: TextAlign.center,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.bold,
                  color: Colors.red,
                ),
              ),
            ),
            Container(
              child: const Text(
                "900.00 บาท",
                textAlign: TextAlign.right,
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.bold),
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _searchInput() {
    return Padding(
      padding: EdgeInsets.only(top: 12, left: 10, right: 10),
      child: TextField(
        onSubmitted: (text) {},
        controller: txtSearch,
        style: TextStyle(color: Colors.black, fontSize: 16.0),
        decoration: InputDecoration(
          border: const OutlineInputBorder(
            borderRadius: const BorderRadius.all(const Radius.circular(8.0)),
          ),
          prefixIcon: IconButton(
            icon: const Icon(Icons.search),
            color: Colors.lightBlue.shade500,
            onPressed: () {},
          ),
          fillColor: Colors.white,
          filled: true,
          hintText: global.language('receivelist_search'),
          hintStyle: const TextStyle(color: Colors.grey),
        ),
      ),
    );
  }
}
