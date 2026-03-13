// import 'package:flutter/material.dart';
// import 'package:get_storage/get_storage.dart';
// import 'package:uuid/uuid.dart';
// import 'package:smlaicloud/global.dart' as global;
// import 'package:smlaicloud/environment.dart';
// // ignore: depend_on_referenced_packages
// import 'package:intl/intl.dart';

// // ignore: avoid_web_libraries_in_flutter
// import "dart:html" as html;
// import 'dart:ui_web' as ui;

// class ReportScreen extends StatefulWidget {
//   final global.ReportEnum type;
//   const ReportScreen({super.key, required this.type});

//   @override
//   State<ReportScreen> createState() => _PurchaseReportState();
// }

// class _PurchaseReportState extends State<ReportScreen> with global.ThemeRefreshMixin {
//   final TextEditingController search = TextEditingController();
//   final TextEditingController fromDate = TextEditingController();
//   final TextEditingController toDate = TextEditingController();
//   String viewType = 'initpdfIframe';
//   String selectedYear = (DateTime.now().year + 543).toString();

//   List<String> monthsThai = ["มกราคม", "กุมภาพันธ์", "มีนาคม", "เมษายน", "พฤษภาคม", "มิถุนายน", "กรกฎาคม", "สิงหาคม", "กันยายน", "ตุลาคม", "พฤศจิกายน", "ธันวาคม"];

//   String selectedMonth = "0";
//   final appConfig = GetStorage("AppConfig");

//   @override
//   void initState() {
//     // ignore: undefined_prefixed_name
//     ui.platformViewRegistry.registerViewFactory(
//       viewType,
//       (int viewId) => html.DivElement()..text = '',
//     );
//     selectedMonth = monthsThai[DateTime.now().month - 1];
//     super.initState();
//   }

//   void getReport() {
//     viewType = const Uuid().v4().split("-")[3];
//     String querySearch = "";
//     String queryFromdate = "";
//     String queryTodate = "";
//     // print(appConfig.read("token"));
//     if (widget.type == global.ReportEnum.vatpurchase || widget.type == global.ReportEnum.vatsale) {
//       if (search.text != '') {
//         querySearch = "&search=${search.text}";
//       }
//       var monthnum = monthsThai.indexOf(selectedMonth);
//       var yearnum = int.parse(selectedYear) - 543;
//       queryFromdate = "&year=$yearnum&month=$monthnum";
//     } else {
//       if (search.text != '') {
//         querySearch = "&search=${search.text}";
//       }

//       if (fromDate.text != '') {
//         queryFromdate = "&fromdate=${fromDate.text}";
//       }

//       if (toDate.text != '') {
//         queryTodate = "&todate=${toDate.text}";
//       }
//     }
//     print(getreportUrl() + "/pdfview?token=${appConfig.read("token")}$querySearch$queryFromdate$queryTodate");
//     // ignore: undefined_prefixed_name
//     ui.platformViewRegistry.registerViewFactory(
//       viewType,
//       (int viewId) => html.IFrameElement()
//         ..width = '640'
//         ..height = '360'
//         ..src = '${Environment().config.reportApi}${getreportUrl()}/pdfview?token=${appConfig.read("token")}$querySearch$queryFromdate$queryTodate'
//         ..style.border = 'none',
//     );

//     // print('135113');
//     setState(() {});
//   }

//   String getreportUrl() {
//     switch (widget.type) {
//       case global.ReportEnum.product:
//         return "/product";
//       case global.ReportEnum.saleinvoice:
//         return "/saleinvoice";
//       case global.ReportEnum.debtor:
//         return "/debtor";
//       case global.ReportEnum.creditor:
//         return "/creditor";
//       case global.ReportEnum.bookbank:
//         return "/bookbank";
//       case global.ReportEnum.purchase:
//         return "/purchase";
//       case global.ReportEnum.purchasereturn:
//         return "/purchasereturn";
//       case global.ReportEnum.saleinvoicereturn:
//         return "/saleinvoicereturn";
//       case global.ReportEnum.transfer:
//         return "/transfer";
//       case global.ReportEnum.receive:
//         return "/receive";
//       case global.ReportEnum.pickup:
//         return "/pickup";
//       case global.ReportEnum.returnproduct:
//         return "/returnproduct";
//       case global.ReportEnum.stockadjustment:
//         return "/stockadjustment";
//       case global.ReportEnum.paid:
//         return "/paid";
//       case global.ReportEnum.pay:
//         return "/pay";
//       case global.ReportEnum.getpaid:
//         return "/getpaid";
//       case global.ReportEnum.getpay:
//         return "/getpay";
//       case global.ReportEnum.vatsale:
//         return "/vatsale";
//       case global.ReportEnum.vatpurchase:
//         return "/vatpurchase";
//       case global.ReportEnum.salebydebtor:
//         return "/salebydebtor";
//       case global.ReportEnum.receivemoney:
//         return "/receivemoney";
//       default:
//         return "";
//     }
//   }

//   Widget datefilterBox() {
//     if (widget.type != global.ReportEnum.product &&
//         widget.type != global.ReportEnum.debtor &&
//         widget.type != global.ReportEnum.creditor &&
//         widget.type != global.ReportEnum.bookbank &&
//         widget.type != global.ReportEnum.vatsale &&
//         widget.type != global.ReportEnum.vatpurchase) {
//       return Container(
// margin: EdgeInsets.only(bottom: 10),
//           child: Row(children: [
//             Expanded(
//                 child: TextField(
//               decoration: InputDecoration(
// border: OutlineInputBorder(),
//                   labelText: global.language("from_date"),
//                   suffixIcon: Row(
//                     mainAxisAlignment: MainAxisAlignment.spaceBetween, // added line
//                     mainAxisSize: MainAxisSize.min,
//                     children: [
//                       IconButton(
//                         focusNode: FocusNode(skipTraversal: true),
//                         icon: Icon(Icons.calendar_month),
//                         onPressed: () {
//                           viewType = const Uuid().v4().split("-")[3];
//                           // ignore: undefined_prefixed_name
//                           ui.platformViewRegistry.registerViewFactory(
//                             viewType,
//                             (int viewId) => html.DivElement()..text = '',
//                           );
//                           setState(() {});
//                           _selectFromDate(context);
//                         },
//                       ),
//                     ],
//                   )),
//               controller: fromDate,
//               onChanged: (value) {},
//             )),
//             const SizedBox(
//               width: 5,
//             ),
//             Expanded(
//                 child: TextField(
//               decoration: InputDecoration(
// border: OutlineInputBorder(),
//                   labelText: global.language("to_date"),
//                   suffixIcon: Row(
//                     mainAxisAlignment: MainAxisAlignment.spaceBetween, // added line
//                     mainAxisSize: MainAxisSize.min,
//                     children: [
//                       IconButton(
//                         focusNode: FocusNode(skipTraversal: true),
//                         icon: Icon(Icons.calendar_month),
//                         onPressed: () {
//                           viewType = const Uuid().v4().split("-")[3];
//                           // ignore: undefined_prefixed_name
//                           ui.platformViewRegistry.registerViewFactory(
//                             viewType,
//                             (int viewId) => html.DivElement()..text = '',
//                           );
//                           setState(() {});
//                           _selectToDate(context);
//                         },
//                       ),
//                     ],
//                   )),
//               controller: toDate,
//               onChanged: (value) {},
//             )),
//           ]));
//     } else {
//       return Container();
//     }
//   }

//   Widget vatBox() {
//     return Container(
//         margin: const EdgeInsets.only(bottom: 10),
//         child: Row(children: [
//           Expanded(
//             child: TextField(
//               decoration: const InputDecoration(
//                 labelText: 'ปีภาษี',
//               ),
//               onTap: () async {
//                 viewType = const Uuid().v4().split("-")[3];
//                 // ignore: undefined_prefixed_name
//                 ui.platformViewRegistry.registerViewFactory(
//                   viewType,
//                   (int viewId) => html.DivElement()..text = '',
//                 );
//                 setState(() {});

//                 _showDropdownYear();
//               },
//               controller: TextEditingController(text: selectedYear),
//             ),
//           ),
//           const SizedBox(
//             width: 5,
//           ),
//           Expanded(
//             child: TextField(
//               decoration: const InputDecoration(
//                 labelText: 'เดือนภาษี',
//               ),
//               onTap: () async {
//                 viewType = const Uuid().v4().split("-")[3];
//                 // ignore: undefined_prefixed_name
//                 ui.platformViewRegistry.registerViewFactory(
//                   viewType,
//                   (int viewId) => html.DivElement()..text = '',
//                 );
//                 setState(() {});

//                 _showDropdownMonth();
//               },
//               controller: TextEditingController(text: selectedMonth),
//             ),
//           ),
//         ]));
//   }

//   void _showDropdownYear() {
//     List<String> yearItems = [];

//     for (int year = DateTime.now().year; year >= DateTime.now().year - 10; year--) {
//       yearItems.add(year.toString());
//     }
//     showDialog(
//       context: context,
//       builder: (BuildContext context) {
//         return AlertDialog(
//           content: Column(
//             mainAxisSize: MainAxisSize.min,
//             children: yearItems.map((item) {
//               return ListTile(
//                 title: Text((int.parse(item) + 543).toString()),
//                 onTap: () {
//                   setState(() {
//                     selectedYear = (int.parse(item) + 543).toString();
//                   });
//                   Navigator.of(context).pop();
//                 },
//               );
//             }).toList(),
//           ),
//         );
//       },
//     );
//   }

//   void _showDropdownMonth() {
//     showDialog(
//       context: context,
//       builder: (BuildContext context) {
//         return AlertDialog(
//           content: Column(
//             mainAxisSize: MainAxisSize.min,
//             children: monthsThai.map((item) {
//               return ListTile(
//                 title: Text(item),
//                 onTap: () {
//                   setState(() {
//                     selectedMonth = item;
//                   });
//                   Navigator.of(context).pop();
//                 },
//               );
//             }).toList(),
//           ),
//         );
//       },
//     );
//   }

//   Widget searchBox() {
//     return Container(
// margin: EdgeInsets.only(bottom: 10),
//         child: Row(children: [
//           Expanded(
//             child: TextFormField(
//               controller: search,
//               decoration: InputDecoration(
// border: OutlineInputBorder(),
//                 labelText: global.language("search"),
//               ),
//             ),
//           ),
//         ]));
//   }

//   void _selectFromDate(BuildContext context) async {
//     final DateTime? pickedDate = await showDatePicker(
//       context: context,
//       initialDate: DateTime.now(),
//       firstDate: DateTime(2000),
//       lastDate: DateTime.now(),
//     );

//     if (pickedDate != null) {
//       setState(() {
//         fromDate.text = DateFormat('yyyy-MM-dd').format(pickedDate);
//       });
//     }
//   }

//   void _selectToDate(BuildContext context) async {
//     final DateTime? pickedDate = await showDatePicker(
//       context: context,
//       initialDate: DateTime.now(),
//       firstDate: DateTime(2000),
//       lastDate: DateTime.now(),
//     );

//     if (pickedDate != null) {
//       setState(() {
//         toDate.text = DateFormat('yyyy-MM-dd').format(pickedDate);
//       });
//     }
//   }

//   @override
//   Widget build(BuildContext context) {
//     return Scaffold(
//       appBar: AppBar(
//         leading: IconButton(
//           focusNode: FocusNode(skipTraversal: true),
//           icon: Icon(Icons.arrow_back),
//           onPressed: () {
//             Navigator.pushNamed(context, '/menu');
//           },
//         ),
//         backgroundColor: global.theme.appBarColor,
//         title: Text(global.getreportName(widget.type)),
//       ),
//       body: Card(
//           child: Container(
//         margin: const EdgeInsets.all(10),
//         child: Column(children: [
//           datefilterBox(),
//           (widget.type == global.ReportEnum.paid ||
//                   widget.type == global.ReportEnum.pay ||
//                   widget.type == global.ReportEnum.vatpurchase ||
//                   widget.type == global.ReportEnum.vatsale)
//               ? Container()
//               : searchBox(),
//           (widget.type == global.ReportEnum.vatpurchase || widget.type == global.ReportEnum.vatsale) ? vatBox() : Container(),
//           Container(
//             alignment: Alignment.centerLeft,
//             margin: const EdgeInsets.only(bottom: 10),
//             child: ElevatedButton(
//                 onPressed: () {
//                   getReport();
//                 },
//                 child: Text("process")),
//           ),
//           const Divider(
//             height: 3,
//           ),
//           Expanded(
//             child: HtmlElementView(viewType: viewType),
//           ),
//         ]),
//       )),
//     );
//   }
// }
// // //
