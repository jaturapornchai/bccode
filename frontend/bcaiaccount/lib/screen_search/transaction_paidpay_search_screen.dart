import 'dart:io';

import 'package:smlaicloud/bloc/transaction_paidpay/transaction_paidpay_bloc.dart';
import 'package:smlaicloud/global.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:intl/intl.dart';

class TransPaidPaySearchScreen extends StatefulWidget {
  const TransPaidPaySearchScreen({super.key, required this.type});
  final TransactionTypeEnum type;

  @override
  State<TransPaidPaySearchScreen> createState() =>
      TransPaidPaySearchScreenState();
}

class TransPaidPaySearchScreenState extends State<TransPaidPaySearchScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  ScrollController listScrollController = ScrollController();
  String searchText = "";
  List<TransactionPaidPayModel> transListData = [];
  bool isKeyUp = false;
  bool isKeyDown = false;
  String selectGuid = "";
  int _hoverIndex = -1;
  int currentListIndex = 0;
  final _debouncer = global.Debouncer(1000);

  TransactionPaidPayModel emptyResult = TransactionPaidPayModel(
    docno: '',
    docdatetime: DateTime.now().toUtc().toIso8601String(),
    doctype: 0,
    custcode: '',
    salecode: '',
    salename: '',
    totalpaymentamount: 0,
    totalvalue: 0,
    totalamount: 0,
    totalbalance: 0,
    details: [],
    transflag: 50,
  );

  List<Widget> tableHeader = [];

  void setSystemLanguageList() async {
    await global.setSystemLanguage(context);

    loadDataList(searchText);
  }

  @override
  void initState() {
    setSystemLanguageList();
    listScrollController.addListener(onScrollList);
    searchText = '';
    searchController.text = searchText;

    tableHeader.add(
      Expanded(
        flex: 1,
        child: Text(
          global.language("doc_date"),
          style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
            fontWeight: FontWeight.bold,
          ),
        ),
      ),
    );
    tableHeader.add(
      Expanded(
        flex: 1,
        child: Text(
          global.language("docno"),
          style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
            fontWeight: FontWeight.bold,
          ),
        ),
      ),
    );

    tableHeader.add(
      Expanded(
        flex: 1,
        child: Text(
          (widget.type == TransactionTypeEnum.pay)
              ? global.language("supplier")
              : global.language("customer"),
          style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
            fontWeight: FontWeight.bold,
          ),
          textAlign: TextAlign.center,
        ),
      ),
    );
    tableHeader.add(
      Expanded(
        flex: 1,
        child: Text(
          textAlign: TextAlign.center,
          global.language("sale_name"),
          style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
            fontWeight: FontWeight.bold,
          ),
        ),
      ),
    );

    tableHeader.add(
      Expanded(
        flex: 1,
        child: Text(
          textAlign: TextAlign.center,
          global.language("product_list"),
          style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
            fontWeight: FontWeight.bold,
          ),
        ),
      ),
    );

    tableHeader.add(
      Expanded(
        flex: 1,
        child: Text(
          textAlign: TextAlign.center,
          global.language("total_value"),
          style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
            fontWeight: FontWeight.bold,
          ),
        ),
      ),
    );

    super.initState();
  }

  void loadDataList(String search) {
    context.read<TransactionPaidPayBloc>().add(
      TransactionPaidPayLoad(
        offset: (transListData.isEmpty) ? 0 : transListData.length,
        limit: global.loadDataPerPage,
        search: search,
        type: widget.type,
      ),
    );
  }

  void onScrollList() {
    if (listScrollController.offset >=
            listScrollController.position.maxScrollExtent &&
        !listScrollController.position.outOfRange) {
      loadDataList(searchText);
    }
  }

  @override
  void dispose() {
    listScrollController.dispose();
    searchController.dispose();
    super.dispose();
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.transactionName(widget.type)),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: Icon(Icons.arrow_back),
          onPressed: () {
            Navigator.pop(context, emptyResult);
          },
        ),
      ),
      body: Focus(
        focusNode: FocusNode(skipTraversal: true, canRequestFocus: true),
        onKey: (node, event) {
          if (kIsWeb ||
              Platform.isWindows ||
              Platform.isLinux ||
              Platform.isMacOS) {
            if (event is RawKeyDownEvent) {
              if (event.logicalKey == LogicalKeyboardKey.escape) {
                Navigator.pop(context, emptyResult);
              }
              if (event.logicalKey == LogicalKeyboardKey.tab) {
                if (selectGuid != "") {
                  Navigator.pop(context, emptyResult);
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
                isKeyDown = false;
                int index = transListData.indexOf(
                  transListData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index > 0) {
                  selectGuid = transListData[index - 1].guidfixed!;
                  isKeyUp = true;
                }
                setState(() {});
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                isKeyUp = false;
                int index = transListData.indexOf(
                  transListData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index < transListData.length - 1) {
                  selectGuid = transListData[index + 1].guidfixed!;
                }
                isKeyDown = true;
                setState(() {});
              }
            }
          }
          return KeyEventResult.ignored;
        },
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.all(5),
              color: global.theme.appBarColor,
              child: Container(
                padding: const EdgeInsets.all(5),
                decoration: BoxDecoration(
                  color: global.theme.surfaceColor,
                  borderRadius: BorderRadius.circular(2),
                  boxShadow: [
                    BoxShadow(
                      color: global.theme.dividerBorderColor.withValues(alpha: 0.5),
                      spreadRadius: 5,
                      blurRadius: 7,
                      offset: const Offset(0, 2),
                    ),
                  ],
                ),
                child: Padding(
                  padding: const EdgeInsets.only(left: 10, right: 10),
                  child: TextFormField(
                    onFieldSubmitted: (value) {
                      searchFocusNode.requestFocus();
                    },
                    onChanged: (value) {
                      _debouncer.run(() {
                        try {
                          setState(() {
                            transListData = [];
                          });
                          loadDataList(value);
                        } catch (_) {}
                      });
                    },
                    autofocus: true,
                    focusNode: searchFocusNode,
                    controller: searchController,
                    style: TextStyle(color: global.theme.textColor),
                    decoration: InputDecoration(
                      isDense: true,
                      filled: false,
                      contentPadding: const EdgeInsets.only(
                        top: 10,
                        bottom: 10,
                      ),
                      border: InputBorder.none,
                      prefixIcon: Icon(Icons.search, size: 20, color: global.theme.iconColor),
                      prefixIconConstraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                      hintText: global.language('search'),
                      hintStyle: TextStyle(color: global.theme.formHintColor),
                    ),
                  ),
                ),
              ),
            ),
            Container(
              padding: const EdgeInsets.only(
                left: 10,
                right: 10,
                top: 5,
                bottom: 5,
              ),
              decoration: BoxDecoration(
                color: global.theme.columnHeaderColor,
                border: Border(
                bottom: BorderSide(width: 1.0, color: global.theme.dividerBorderColor),
                ),
              ),
              child: Row(children: tableHeader),
            ),
            Expanded(
              child: SingleChildScrollView(
                controller: listScrollController,
                child: Column(
                  children: transListData
                      .asMap().entries.map((e) => listObject(e.value, e.key))
                      .toList(),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }


  Color _getContainerColor(String itemGuid, int index) {
    if (selectGuid.isNotEmpty && selectGuid == itemGuid) {
      return global.theme.rowSelectedColor;
    }
    if (_hoverIndex == index) {
      return global.theme.rowHoverColor;
    }
    return (index % 2 == 0)
        ? global.theme.columnAlternateEvenColor
        : global.theme.columnAlternateOddColor;
  }

  Widget listObject(TransactionPaidPayModel value, int index) {
    final isSelected = selectGuid == value.guidfixed;
    TextStyle textStyle = isSelected
        ? TextStyle(fontSize: global.deviceConfig.listDataFontSize, fontWeight: FontWeight.w700, color: global.theme.textColor)
        : TextStyle(fontSize: global.deviceConfig.listDataFontSize, fontWeight: FontWeight.w400, color: global.theme.textSecondaryColor);
    DateTime docDateTime = DateTime.parse(value.docdatetime);
    List<Widget> tableDetails = [];

    tableDetails.add(
      Expanded(
        flex: 1,
        child: Text(
          DateFormat('dd/MM/yyyy HH:mm').format(docDateTime.toLocal()),
          style: textStyle,
        ),
      ),
    );
    tableDetails.add(Expanded(flex: 1, child: Text(value.docno, style: textStyle)));

    tableDetails.add(
      Expanded(
        flex: 1,
        child: Text(
          "${value.custcode} : ${global.activeLangName(value.custnames ?? [])}",
          textAlign: TextAlign.center,
          style: textStyle,
        ),
      ),
    );

    tableDetails.add(
      Expanded(
        flex: 1,
        child: Text(
          value.details!.length.toString(),
          textAlign: TextAlign.center,
          style: textStyle,
        ),
      ),
    );

    tableDetails.add(
      Expanded(
        flex: 1,
        child: Text(
          global.formatNumber(value.totalamount),
          textAlign: TextAlign.right,
          style: textStyle,
        ),
      ),
    );
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      onEnter: (_) => setState(() => _hoverIndex = index),
      onExit: (_) => setState(() => _hoverIndex = -1),
      child: GestureDetector(
      onTap: () {
        Navigator.pop(context, value);
      },
      child: Container(
        decoration: BoxDecoration(
          color: _getContainerColor(value.guidfixed ?? "", index),
          border: Border(
                bottom: BorderSide(width: 1.0, color: global.theme.dividerBorderColor),
          ),
        ),
        padding: const EdgeInsets.only(left: 10, right: 10, top: 5, bottom: 5),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: tableDetails,
        ),
      ),
    ),
    );
  }

  @override
  Widget build(BuildContext context) {
    for (int i = 0; i < transListData.length; i++) {
      if (transListData[i].guidfixed == selectGuid) {
        currentListIndex = i;
        break;
      }
    }
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return BlocListener<TransactionPaidPayBloc, TransactionPaidPayState>(
            listener: (context, state) {
              // Load
              if (state is TransactionPaidPayLoadSuccess) {
                setState(() {
                  if (state.transactionPaidPay.isNotEmpty) {
                    transListData.addAll(state.transactionPaidPay);
                    if (transListData.isNotEmpty) {
                      selectGuid = transListData[0].guidfixed!;
                    } else {
                      selectGuid = "";
                    }
                  }
                });
              }
            },
            child: (constraints.maxWidth > 800)
                ? listScreen(mobileScreen: false)
                : listScreen(mobileScreen: true),
          );
        },
      ),
    );
  }
}
