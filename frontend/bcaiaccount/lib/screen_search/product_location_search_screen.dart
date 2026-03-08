import 'dart:io';

import 'package:smlaicloud/bloc/warehouse/warehose_bloc.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/location_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;

class ProductLocationSearchScreen extends StatefulWidget {
  const ProductLocationSearchScreen({super.key, required this.whcode});
  final String whcode;

  @override
  State<ProductLocationSearchScreen> createState() =>
      ProductLocationSearchScreenState();
}

class ProductLocationSearchScreenState
    extends State<ProductLocationSearchScreen>
    with SingleTickerProviderStateMixin {
  ScrollController listScrollController = ScrollController();
  String searchWhcode = "";
  List<LocationModel> locationListData = [];
  bool isKeyUp = false;
  bool isKeyDown = false;
  String selectGuid = "";
  int currentListIndex = 0;

  void setSystemLanguageList() async {
    await global.setSystemLanguage(context);

    loadDataList(searchWhcode);
  }

  @override
  void initState() {
    setSystemLanguageList();
    listScrollController.addListener(onScrollList);
    searchWhcode = widget.whcode;

    super.initState();
  }

  void loadDataList(String whcode) {
    context.read<WarehouseBloc>().add(WarehouseGetByCode(code: whcode));
  }

  void onScrollList() {
    if (listScrollController.offset >=
            listScrollController.position.maxScrollExtent &&
        !listScrollController.position.outOfRange) {
      loadDataList(searchWhcode);
    }
  }

  @override
  void dispose() {
    listScrollController.dispose();
    super.dispose();
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('location')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () {
            Navigator.pop(
              context,
              SearchGuidCodeNameModel(
                guid: "",
                code: "",
                names: [],
                isCancel: true,
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
                  SearchGuidCodeNameModel(
                    guid: "",
                    code: "",
                    names: [],
                    isCancel: true,
                  ),
                );
              }
              if (event.logicalKey == LogicalKeyboardKey.tab) {
                if (selectGuid != "") {
                  Navigator.pop(
                    context,
                    global.SearchCodeNameModel(
                      code: locationListData[currentListIndex].code,
                      names: locationListData[currentListIndex].names,
                      isCancel: false,
                      guidfixed: '',
                    ),
                  );
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
                isKeyDown = false;
                int index = locationListData.indexOf(
                  locationListData.firstWhere(
                    (element) => element.code == selectGuid,
                  ),
                );
                if (index > 0) {
                  selectGuid = locationListData[index - 1].code;
                  isKeyUp = true;
                }
                setState(() {});
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                isKeyUp = false;
                int index = locationListData.indexOf(
                  locationListData.firstWhere(
                    (element) => element.code == selectGuid,
                  ),
                );
                if (index < locationListData.length - 1) {
                  selectGuid = locationListData[index + 1].code;
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
                      global.language("location_code"),
                      style: const TextStyle(
                        color: Colors.black,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 10,
                    child: Text(
                      global.language("location_name"),
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
                  children: locationListData
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

  Widget listObject(LocationModel value) {
    return GestureDetector(
      onTap: () {
        Navigator.pop(
          context,
          SearchGuidCodeNameModel(
            guid: value.code,
            code: value.code,
            names: value.names,
            isCancel: false,
          ),
        );
      },
      child: Container(
        decoration: BoxDecoration(
          color: (selectGuid == value.code) ? Colors.cyan[100] : Colors.white,
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
    for (int i = 0; i < locationListData.length; i++) {
      if (locationListData[i].code == selectGuid) {
        currentListIndex = i;
        break;
      }
    }
    return Scaffold(
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return BlocListener<WarehouseBloc, WarehouseState>(
            listener: (context, state) {
              // Load
              if (state is WarehouseGetSuccess) {
                setState(() {
                  if (state.warehouse.code.isNotEmpty) {
                    locationListData.addAll(state.warehouse.location);
                    if (locationListData.isNotEmpty) {
                      selectGuid = locationListData[0].code;
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
