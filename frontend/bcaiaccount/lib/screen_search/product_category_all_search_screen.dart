import 'dart:io';

import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/master_category/master_category_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/master_category_model.dart';
import 'package:smlaicloud/components/loading_overlay.dart';

class ProductCategorySearchScreenNew extends StatefulWidget {
  const ProductCategorySearchScreenNew({super.key, required this.word});
  final String word;

  @override
  State<ProductCategorySearchScreenNew> createState() =>
      ProductCategorySearchScreenNewState();
}

class ProductCategorySearchScreenNewState
    extends State<ProductCategorySearchScreenNew>
    with SingleTickerProviderStateMixin {
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  ScrollController listScrollController = ScrollController();
  String searchText = "";
  List<MasterCategoryModel> masterCategoryListData = [];
  bool isKeyUp = false;
  bool isKeyDown = false;
  String selectGuid = "";
  int currentListIndex = 0;
  final _debouncer = global.Debouncer(1000);
  bool loadingData = false;
  bool isLoading = false; // สำหรับ LoadingOverlay

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
    setState(() {
      loadingData = true;
    });
    searchText = search;

    context.read<MasterCategoryBloc>().add(
      MasterCategoryLoadList(
        offset: (masterCategoryListData.isEmpty)
            ? 0
            : masterCategoryListData.length,
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
        title: Text(global.language('category')),
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
                      code: masterCategoryListData[currentListIndex].code,
                      names: masterCategoryListData[currentListIndex].names,
                      isCancel: false,
                      guidfixed: '',
                    ),
                  );
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
                isKeyDown = false;
                int index = masterCategoryListData.indexOf(
                  masterCategoryListData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index > 0) {
                  selectGuid = masterCategoryListData[index - 1].guidfixed;
                  isKeyUp = true;
                }
                setState(() {});
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                isKeyDown = false;
                int index = masterCategoryListData.indexOf(
                  masterCategoryListData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index < masterCategoryListData.length - 1) {
                  selectGuid = masterCategoryListData[index + 1].guidfixed;
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
                      try {
                        _debouncer.run(() {
                          setState(() {
                            masterCategoryListData = [];
                          });
                          loadDataList(value);
                        });
                      } catch (_) {}
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
                      global.language("category_code"),
                      style: const TextStyle(
                        color: Colors.black,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 10,
                    child: Text(
                      global.language("category_name"),
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
                  children: masterCategoryListData
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

  Widget listObject(MasterCategoryModel value) {
    return GestureDetector(
      onTap: () {
        Navigator.pop(
          context,
          global.SearchCodeNameModel(
            code: value.code,
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
                value.code,
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
    for (int i = 0; i < masterCategoryListData.length; i++) {
      if (masterCategoryListData[i].guidfixed == selectGuid) {
        currentListIndex = i;
        break;
      }
    }
    return Scaffold(
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return BlocListener<MasterCategoryBloc, MasterCategoryState>(
            listener: (context, state) {
              // Load
              if (state is MasterCategoryLoadSuccess) {
                setState(() {
                  loadingData = false;
                  if (state.categories.isNotEmpty) {
                    masterCategoryListData.addAll(state.categories);
                    if (masterCategoryListData.isNotEmpty) {
                      selectGuid = masterCategoryListData[0].guidfixed;
                    } else {
                      selectGuid = "";
                    }
                  }
                });
              }
            },
            child: Stack(
              children: [
                // Main content
                (constraints.maxWidth > 800)
                    ? listScreen(mobileScreen: false)
                    : listScreen(mobileScreen: true),

                // Loading Overlay
                if (loadingData)
                  LoadingOverlay(message: global.language('loading')),
              ],
            ),
          );
        },
      ),
    );
  }
}
