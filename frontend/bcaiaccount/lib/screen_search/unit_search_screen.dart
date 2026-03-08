import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/unit/unit_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/product_model.dart';

class UnitSearchScreen extends StatefulWidget {
  const UnitSearchScreen({super.key, required this.word});
  final String word;

  @override
  State<UnitSearchScreen> createState() => UnitSearchScreenState();
}

class UnitSearchScreenState extends State<UnitSearchScreen>
    with SingleTickerProviderStateMixin {
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  ScrollController listScrollController = ScrollController();
  String searchText = "";
  List<UnitModel> unitListData = [];
  bool isKeyUp = false;
  bool isKeyDown = false;
  String selectGuid = "";
  int currentListIndex = 0;
  final _debouncer = global.Debouncer(1000);

  @override
  void initState() {
    listScrollController.addListener(onScrollList);
    searchText = widget.word;
    searchController.text = searchText;
    loadDataList(searchText);
    super.initState();
  }

  void loadDataList(String search) {
    context.read<UnitBloc>().add(
      UnitLoadList(
        offset: (unitListData.isEmpty) ? 0 : unitListData.length,
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
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('product_unit')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () {
            Navigator.pop(
              context,
              global.SearchCodeNameModel(
                code: "",
                names: [],
                isCancel: true,
                guidfixed: '',
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
                  global.SearchCodeNameModel(
                    code: "",
                    names: [],
                    isCancel: true,
                    guidfixed: '',
                  ),
                );
              }
              if (event.logicalKey == LogicalKeyboardKey.tab) {
                if (selectGuid != "") {
                  Navigator.pop(
                    context,
                    global.SearchCodeNameModel(
                      code: unitListData[currentListIndex].unitcode,
                      names: unitListData[currentListIndex].names,
                      isCancel: false,
                      guidfixed: '',
                    ),
                  );
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
                isKeyDown = false;
                int index = unitListData.indexOf(
                  unitListData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index > 0) {
                  selectGuid = unitListData[index - 1].guidfixed;
                  isKeyUp = true;
                }
                setState(() {});
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                isKeyUp = false;
                int index = unitListData.indexOf(
                  unitListData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index < unitListData.length - 1) {
                  selectGuid = unitListData[index + 1].guidfixed;
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
                  color: Colors.white,
                  borderRadius: BorderRadius.circular(2),
                  boxShadow: [
                    BoxShadow(
                      color: Colors.grey.withOpacity(0.5),
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
                            unitListData = [];
                          });
                          loadDataList(value);
                        } catch (_) {}
                      });
                    },
                    autofocus: true,
                    focusNode: searchFocusNode,
                    controller: searchController,
                    decoration: InputDecoration(
                      isDense: true,
                      contentPadding: const EdgeInsets.only(
                        top: 10,
                        bottom: 10,
                      ),
                      border: InputBorder.none,
                      hintText: global.language('search'),
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
                border: const Border(
                  bottom: BorderSide(width: 1.0, color: Colors.grey),
                ),
              ),
              child: Row(
                children: [
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("unit_code"),
                      style: const TextStyle(
                        color: Colors.black,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 10,
                    child: Text(
                      global.language("unit_name"),
                      style: const TextStyle(
                        color: Colors.black,
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
                  children: unitListData
                      .map((value) => listObject(value))
                      .toList(),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget listObject(UnitModel value) {
    return GestureDetector(
      onTap: () {
        Navigator.pop(
          context,
          global.SearchCodeNameModel(
            code: value.unitcode,
            names: value.names,
            isCancel: false,
            guidfixed: '',
          ),
        );
      },
      child: Container(
        decoration: BoxDecoration(
          color: (selectGuid == value.guidfixed)
              ? Colors.cyan[100]
              : Colors.white,
          border: const Border(
            bottom: BorderSide(width: 1.0, color: Colors.grey),
          ),
        ),
        padding: const EdgeInsets.only(left: 10, right: 10, top: 5, bottom: 5),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              flex: 5,
              child: Text(
                value.unitcode,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            Expanded(
              flex: 10,
              child: Text(
                global.packName(value.names),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
          ],
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    for (int i = 0; i < unitListData.length; i++) {
      if (unitListData[i].guidfixed == selectGuid) {
        currentListIndex = i;
        break;
      }
    }
    return Scaffold(
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return BlocListener<UnitBloc, UnitState>(
            listener: (context, state) {
              // Load
              if (state is UnitLoadSuccess) {
                setState(() {
                  if (state.units.isNotEmpty) {
                    unitListData.addAll(state.units);
                    if (unitListData.isNotEmpty) {
                      selectGuid = unitListData[0].guidfixed;
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
