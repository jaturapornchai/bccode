import 'dart:io';

import 'package:smlaicloud/bloc/transport_channel/transport_channel_bloc.dart';
import 'package:smlaicloud/model/transport_channel_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;

class TransportSearchScreen extends StatefulWidget {
  const TransportSearchScreen({super.key, required this.word});
  final String word;

  @override
  State<TransportSearchScreen> createState() => TransportSearchScreenState();
}

class TransportSearchScreenState extends State<TransportSearchScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  ScrollController listScrollController = ScrollController();
  String searchText = "";
  List<TransportChannelModel> listData = [];
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
    context.read<TransportChannelBloc>().add(TransportChannelLoadList(
        offset: (listData.isEmpty) ? 0 : listData.length,
        limit: global.loadDataPerPage,
        search: search));
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
        title: Text(global.language('sale_channel')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: Icon(Icons.arrow_back),
          onPressed: () {
            Navigator.pop(context,
                TransportChannelModel(code: "", name: "", guidfixed: ''));
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
                  Navigator.pop(context,
                      TransportChannelModel(code: "", name: "", guidfixed: ''));
                }
                if (event.logicalKey == LogicalKeyboardKey.tab) {
                  if (selectGuid != "") {
                    Navigator.pop(
                        context,
                        TransportChannelModel(
                            code: listData[currentListIndex].code,
                            name: listData[currentListIndex].name,
                            guidfixed: listData[currentListIndex].guidfixed));
                  }
                }
                if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
                  isKeyDown = false;
                  int index = listData.indexOf(listData.firstWhere(
                      (element) => element.guidfixed == selectGuid));
                  if (index > 0) {
                    selectGuid = listData[index - 1].guidfixed;
                    isKeyUp = true;
                  }
                  setState(() {});
                }
                if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                  isKeyUp = false;
                  int index = listData.indexOf(listData.firstWhere(
                      (element) => element.guidfixed == selectGuid));
                  if (index < listData.length - 1) {
                    selectGuid = listData[index + 1].guidfixed;
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
                          padding: EdgeInsets.only(left: 10, right: 10),
                          child: TextFormField(
                              onFieldSubmitted: (value) {
                                searchFocusNode.requestFocus();
                              },
                              onChanged: (value) {
                                _debouncer.run(() {
                                  try {
                                    setState(() {
                                      listData = [];
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
                                contentPadding:
                                    const EdgeInsets.only(top: 10, bottom: 10),
                                border: InputBorder.none,
                                hintText: global.language('search'),
                                hintStyle: TextStyle(color: global.theme.formHintColor),
                              ))))),
              Container(
                  padding: const EdgeInsets.only(
                      left: 10, right: 10, top: 5, bottom: 5),
                  decoration: BoxDecoration(
                      color: global.theme.columnHeaderColor,
                      border: Border(
                bottom: BorderSide(width: 1.0, color: global.theme.dividerBorderColor),
                      )),
                  child: Row(children: [
                    Expanded(
                        flex: 5,
                        child: Text(global.language("sale_channel_code"),
                            style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                                fontWeight: FontWeight.bold))),
                    Expanded(
                        flex: 10,
                        child: Text(
                          global.language("sale_channel_name"),
                          style: TextStyle(
                      color: global.theme.columnHeaderTextColor, fontWeight: FontWeight.bold),
                          maxLines: 2,
                          overflow: TextOverflow.ellipsis,
                        )),
                  ])),
              Expanded(
                  child: SingleChildScrollView(
                      controller: listScrollController,
                      child: Column(
                          children: listData
                              .asMap().entries.map((e) => listObject(e.value, e.key))
                              .toList())))
            ],
          )),
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

  Widget listObject(TransportChannelModel value, int index) {
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
              TransportChannelModel(
                  code: value.code,
                  name: value.name,
                  guidfixed: value.guidfixed));
        },
        child: Container(
            decoration: BoxDecoration(
              color: _getContainerColor(value.guidfixed ?? "", index),
              border: Border(
                bottom: BorderSide(width: 1.0, color: global.theme.dividerBorderColor),
              ),
            ),
            padding:
                const EdgeInsets.only(left: 10, right: 10, top: 5, bottom: 5),
            child: Row(crossAxisAlignment: CrossAxisAlignment.start, children: [
              Expanded(
                  flex: 5,
                  child: Text(value.code,
                      style: textStyle,
                      maxLines: 2, overflow: TextOverflow.ellipsis)),
              Expanded(
                  flex: 10,
                  child: Text(
                    value.name,
                    style: textStyle,
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  )),
            ]))),
    );
  }

  @override
  Widget build(BuildContext context) {
    for (int i = 0; i < listData.length; i++) {
      if (listData[i].guidfixed == selectGuid) {
        currentListIndex = i;
        break;
      }
    }
    return Scaffold(
        backgroundColor: global.theme.backgroundColor,
        resizeToAvoidBottomInset: true,
        body: LayoutBuilder(builder: (context, constraints) {
          return BlocListener<TransportChannelBloc, TransportChannelState>(
              listener: (context, state) {
                // Load
                if (state is TransportChannelLoadSuccess) {
                  setState(() {
                    if (state.transportchannel.isNotEmpty) {
                      listData.addAll(state.transportchannel);
                      if (listData.isNotEmpty) {
                        selectGuid = listData[0].guidfixed;
                      } else {
                        selectGuid = "";
                      }
                    }
                  });
                }
              },
              child: (constraints.maxWidth > 800)
                  ? listScreen(mobileScreen: false)
                  : listScreen(mobileScreen: true));
        }));
  }
}
