import 'dart:io';

import 'package:smlaicloud/bloc/qr/qr_bloc.dart';
import 'package:smlaicloud/model/qr_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;

class QrSearchScreen extends StatefulWidget {
  const QrSearchScreen({super.key, required this.word});
  final String word;

  @override
  State<QrSearchScreen> createState() => QrSearchScreenState();
}

class QrSearchScreenState extends State<QrSearchScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  ScrollController listScrollController = ScrollController();
  String searchText = "";
  List<QrModel> qrcodeListData = [];
  bool isKeyUp = false;
  bool isKeyDown = false;
  String selectGuid = "";
  int _hoverIndex = -1;
  int currentListIndex = 0;
  final _debouncer = global.Debouncer(1000);

  void setSystemLanguageList() async {
    await global.setSystemLanguage(context);

    loadDataList(searchText);
  }

  @override
  void initState() {
    setSystemLanguageList();
    listScrollController.addListener(onScrollList);
    searchText = widget.word;
    searchController.text = searchText;

    super.initState();
  }

  void loadDataList(String search) {
    context.read<QrBloc>().add(
      QrLoadList(
        offset: (qrcodeListData.isEmpty) ? 0 : qrcodeListData.length,
        limit: global.loadDataPerPage,
        search: search,
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
        title: Text(global.language('search_qr_provider')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: Icon(Icons.arrow_back),
          onPressed: () {
            Navigator.pop(
              context,
              QrModel(
                guidfixed: '',
                isactive: true,
                logo: '',
                qrtype: 0,
                code: '',
                qrnames: [],
                bankcode: '',
                banknames: [],
                bookbankcode: '',
                bookbanknames: [],
                bookbankimages: [],
                qrcode: '',
                apikey: '',
                billerCode: '',
                billerID: '',
                storeID: '',
                terminalID: '',
                merchantName: '',
                accessCode: '',
                bankcharge: '',
                customercharge: '',
                closeqr: 0,
              ),
            );
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
                Navigator.pop(
                  context,
                  QrModel(
                    guidfixed: '',
                    isactive: true,
                    logo: '',
                    qrtype: 0,
                    code: '',
                    qrnames: [],
                    bankcode: '',
                    banknames: [],
                    bookbankcode: '',
                    bookbanknames: [],
                    bookbankimages: [],
                    qrcode: '',
                    apikey: '',
                    billerCode: '',
                    billerID: '',
                    storeID: '',
                    terminalID: '',
                    merchantName: '',
                    accessCode: '',
                    bankcharge: '',
                    customercharge: '',
                    closeqr: 0,
                  ),
                );
              }
              if (event.logicalKey == LogicalKeyboardKey.tab) {
                if (selectGuid != "") {
                  Navigator.pop(
                    context,
                    QrModel(
                      guidfixed: qrcodeListData[currentListIndex].guidfixed,
                      isactive: qrcodeListData[currentListIndex].isactive,
                      logo: qrcodeListData[currentListIndex].logo,
                      qrtype: qrcodeListData[currentListIndex].qrtype,
                      code: qrcodeListData[currentListIndex].code,
                      qrnames: qrcodeListData[currentListIndex].qrnames,
                      bankcode: qrcodeListData[currentListIndex].bankcode,
                      banknames: qrcodeListData[currentListIndex].banknames,
                      bookbankcode:
                          qrcodeListData[currentListIndex].bookbankcode,
                      bookbanknames:
                          qrcodeListData[currentListIndex].bookbanknames,
                      bookbankimages:
                          qrcodeListData[currentListIndex].bookbankimages,
                      qrcode: qrcodeListData[currentListIndex].qrcode,
                      apikey: qrcodeListData[currentListIndex].apikey,
                      billerCode: qrcodeListData[currentListIndex].billerCode,
                      billerID: qrcodeListData[currentListIndex].billerID,
                      storeID: qrcodeListData[currentListIndex].storeID,
                      terminalID: qrcodeListData[currentListIndex].terminalID,
                      merchantName:
                          qrcodeListData[currentListIndex].merchantName,
                      accessCode: qrcodeListData[currentListIndex].accessCode,
                      bankcharge: qrcodeListData[currentListIndex].bankcharge,
                      customercharge:
                          qrcodeListData[currentListIndex].customercharge,
                      closeqr: qrcodeListData[currentListIndex].closeqr,
                    ),
                  );
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
                isKeyDown = false;
                int index = qrcodeListData.indexOf(
                  qrcodeListData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index > 0) {
                  selectGuid = qrcodeListData[index - 1].guidfixed!;
                  isKeyUp = true;
                }
                setState(() {});
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                isKeyUp = false;
                int index = qrcodeListData.indexOf(
                  qrcodeListData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index < qrcodeListData.length - 1) {
                  selectGuid = qrcodeListData[index + 1].guidfixed!;
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
                            qrcodeListData = [];
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
              child: Row(
                children: [
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("qr_code_code"),
                      style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 10,
                    child: Text(
                      global.language("qr_code_name"),
                      style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                        fontWeight: FontWeight.bold,
                      ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                ],
              ),
            ),
            Expanded(
              child: SingleChildScrollView(
                controller: listScrollController,
                child: Column(
                  children: qrcodeListData
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

  Widget listObject(QrModel value, int index) {
    final isSelected = selectGuid == value.guidfixed;
    TextStyle textStyle = isSelected
        ? TextStyle(fontSize: global.deviceConfig.listDataFontSize, fontWeight: FontWeight.w700, color: global.theme.textColor)
        : TextStyle(fontSize: global.deviceConfig.listDataFontSize, fontWeight: FontWeight.w400, color: global.theme.textSecondaryColor);
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      onEnter: (_) => setState(() => _hoverIndex = index),
      onExit: (_) => setState(() => _hoverIndex = -1),
      child: GestureDetector(
      onTap: () {
        Navigator.pop(
          context,
          QrModel(
            code: value.code,
            qrnames: value.qrnames,
            guidfixed: value.guidfixed,
            qrtype: value.qrtype,
            logo: value.logo,
            isactive: value.isactive,
            bankcode: value.bankcode,
            banknames: value.banknames,
            bookbankcode: value.bookbankcode,
            bookbanknames: value.bookbanknames,
            bookbankimages: value.bookbankimages,
            qrcode: value.qrcode,
            apikey: value.apikey,
            billerCode: value.billerCode,
            billerID: value.billerID,
            storeID: value.storeID,
            terminalID: value.terminalID,
            merchantName: value.merchantName,
            accessCode: value.accessCode,
            bankcharge: value.bankcharge,
            customercharge: value.customercharge,
            closeqr: value.closeqr,
          ),
        );
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
          children: [
            Expanded(
              flex: 5,
              child: Text(
                value.code!,
                style: textStyle,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            Expanded(
              flex: 10,
              child: Text(
                global.packName(value.qrnames!),
                style: textStyle,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
          ],
        ),
      ),
    ),
    );
  }

  @override
  Widget build(BuildContext context) {
    for (int i = 0; i < qrcodeListData.length; i++) {
      if (qrcodeListData[i].guidfixed == selectGuid) {
        currentListIndex = i;
        break;
      }
    }
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return BlocListener<QrBloc, QrState>(
            listener: (context, state) {
              // Load
              if (state is QrLoadSuccess) {
                setState(() {
                  if (state.qrs.isNotEmpty) {
                    qrcodeListData.addAll(state.qrs);
                    if (qrcodeListData.isNotEmpty) {
                      selectGuid = qrcodeListData[0].guidfixed!;
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
