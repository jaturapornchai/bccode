import 'dart:io';

import 'package:smlaicloud/bloc/product/product_bloc.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/screens/config/product_screen/product_screen_edit.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:flutter/scheduler.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/widgets/list_font_size_control.dart';
import 'package:loading_animation_widget/loading_animation_widget.dart';
import 'package:split_view/split_view.dart';
import 'package:translator/translator.dart';

class ProductScreen extends StatefulWidget {
  const ProductScreen({super.key});

  @override
  State<ProductScreen> createState() => ProductScreenState();
}

class ProductScreenState extends State<ProductScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  late ProductScreenEdit dataEdit;
  late TabController tabController;
  GlobalKey<ProductScreenEditState> dataEditState = GlobalKey();
  String selectGuid = "";
  bool isDataChange = false;
  global.ScreenEventEnum screenEvent = global.ScreenEventEnum.list;
  final translator = GoogleTranslator();
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  int focusNodeIndex = 0;
  List<ProductModel> listData = [];
  List<String> guidListChecked = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  bool isChange = false;
  bool isSaveAllow = false;
  String headerEdit = "";
  int currentListIndex = -1;
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  bool isEditMode = false;
  final debouncer = global.Debouncer(1000);
  bool loadingData = false;
  int _hoverIndex = -1;
  late SplitViewController splitViewController;

  @override
  void initState() {
    screenEvent = global.ScreenEventEnum.list;
    dataEdit = ProductScreenEdit(
      screenEventGetValue: screenEventGetValueCallBack,
      screenEventUpdateValue: screenEventUpdateValueCallBack,
      isDataChangeGetValue: isDataChangeGetValueCallBack,
      isDataChangeUpdateValue: isDataChangeUpdateValueCallBack,
      tabChange: tabChange,
      loadDataList: loadDataList,
      key: dataEditState,
      isSaveAllowGet: isSaveAllowGet,
      isSaveAllowSet: isSaveAllowSet,
      discardData: discardDataEvent,
      headerLabel: global.language("show"),
      selectGuidSet: selectGuidSet,
      selectGuidGet: selectGuidGet,
    );
    splitViewController = SplitViewController(
      limits: [null, WeightLimit(min: 0.2, max: 0.8)],
    );
    tabController = TabController(length: 2, vsync: this);
    listScrollController.addListener(onScrollList);
    SchedulerBinding.instance.addPostFrameCallback((_) {
      loadDataList(false, "");
    });
    super.initState();
  }

  @override
  void dispose() {
    listScrollController.dispose();
    searchController.dispose();
    super.dispose();
  }

  void isSaveAllowSet(bool value) {
    setState(() {
      isSaveAllow = value;
    });
  }

  bool isSaveAllowGet() {
    return isSaveAllow;
  }

  void selectGuidSet(String value) {
    setState(() {
      selectGuid = value;
    });
  }

  String selectGuidGet() {
    return selectGuid;
  }

  void dataChecker({required Function callBack}) {
    if (dataEditState.currentState != null) {
      if ((screenEvent == global.ScreenEventEnum.add ||
              screenEvent == global.ScreenEventEnum.edit) &&
          isDataChange) {
        showDialog<String>(
          context: context,
          builder: (BuildContext context) => AlertDialog(
            title: Text(global.language('data_editing')),
            content: Text(global.language('leave_this_screen')),
            actions: <Widget>[
              ElevatedButton(
                style: ElevatedButton.styleFrom(backgroundColor: global.theme.negativeHighlightTextColor),
                onPressed: () => Navigator.pop(context),
                child: Text(global.language('no')),
              ),
              ElevatedButton(
                style: ElevatedButton.styleFrom(backgroundColor: global.theme.infoHighlightTextColor),
                onPressed: () {
                  Navigator.pop(context);
                  callBack();
                },
                child: Text(global.language('yes')),
              ),
            ],
          ),
        );
      } else {
        callBack();
      }
    } else {
      callBack();
    }
  }

  void tabChange(int index) {
    tabController.animateTo(index);
  }

  void discardDataEvent({required Function callBack}) {
    if ((screenEvent == global.ScreenEventEnum.add ||
            screenEvent == global.ScreenEventEnum.edit) &&
        isDataChange) {
      showDialog<String>(
        context: context,
        builder: (BuildContext context) => AlertDialog(
          title: Text(global.language('data_editing')),
          content: Text(global.language('leave_this_screen')),
          actions: <Widget>[
            ElevatedButton(
              style: ElevatedButton.styleFrom(backgroundColor: global.theme.negativeHighlightTextColor),
              onPressed: () => Navigator.pop(context),
              child: Text(global.language('no')),
            ),
            ElevatedButton(
              style: ElevatedButton.styleFrom(backgroundColor: global.theme.infoHighlightTextColor),
              onPressed: () {
                Navigator.pop(context);
                callBack();
              },
              child: Text(global.language('yes')),
            ),
          ],
        ),
      );
    } else {
      callBack();
    }
  }

  global.ScreenEventEnum screenEventGetValueCallBack() {
    return screenEvent;
  }

  void screenEventUpdateValueCallBack(global.ScreenEventEnum value) {
    screenEvent = value;
  }

  void isDataChangeUpdateValueCallBack(bool value) {
    isDataChange = value;
  }

  bool isDataChangeGetValueCallBack() {
    return isDataChange;
  }

  void loadDataList(bool reset, String search) {
    if (reset) {
      listData.clear();
      guidListChecked.clear();
      listKeys.clear();
    }
    setState(() {
      loadingData = true;
    });
    context.read<ProductBloc>().add(
      ProductLoadList(
        offset: listData.length,
        limit: global.loadDataPerPage,
        search: search,
      ),
    );
  }

  void onScrollList() {
    if (listScrollController.offset >=
            listScrollController.position.maxScrollExtent &&
        !listScrollController.position.outOfRange) {
      loadDataList(false, searchText);
    }
  }

  void addNewData(bool clearData) {
    tabChange(1);
    dataChecker(
      callBack: () {
        if (dataEditState.currentState != null) {
          // หน้าจอใหญ่ ข้าวไปบันทึกได้เลย
          dataEditState.currentState!.addNewData(clearData);
        } else {
          // หน้าจอเล็ก ให้ดึงข้อมูลมาแสดงก่อน
          screenEvent = global.ScreenEventEnum.add;
          dataEditState.currentState?.headerEdit = global.language("append");
          setState(() {});
        }
      },
    );
  }

  void loadDataFromServer({required String guid, required bool isEdit}) {
    if (dataEditState.currentState != null) {
      // กรณีจอใหญ่
      discardDataEvent(
        callBack: () {
          dataEditState.currentState!.getData(guid, isEdit);
        },
      );
    } else {
      // กรณีจอเล็ก
      tabChange(1);
      screenEvent = global.ScreenEventEnum.display;
      dataEditState.currentState?.headerEdit = global.language("display");
      setState(() {});
    }
  }

  Widget listScreen() {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('product')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: Icon(Icons.arrow_back),
          onPressed: () {
            dataChecker(
              callBack: () {
                Navigator.pop(context);
              },
            );
          },
        ),
        actions: <Widget>[
          IconButton(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () {
              discardDataEvent(
                callBack: () {
                  setState(() {
                    if (showCheckBox) {
                      showCheckBox = false;
                      guidListChecked.clear();
                    } else {
                      showCheckBox = true;
                      global.showSnackBar(
                        context,
                        Icon(Icons.delete, color: global.theme.onPrimaryColor),
                        global.language("choose_item_delete"),
                        global.theme.infoHighlightTextColor,
                      );
                    }
                  });
                },
              );
            },
            icon: (showCheckBox)
                ? Icon(Icons.close)
                : Icon(Icons.check_box),
          ),
          if (guidListChecked.isNotEmpty)
            IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () {
                showDialog<String>(
                  context: context,
                  builder: (BuildContext context) => AlertDialog(
                    title: Text(global.language('confirm_delete')),
                    actions: <Widget>[
                      ElevatedButton(
                        style: ElevatedButton.styleFrom(
                          backgroundColor: global.theme.negativeHighlightTextColor,
                        ),
                        onPressed: () => Navigator.pop(context),
                        child: Text(global.language('no')),
                      ),
                      ElevatedButton(
                        style: ElevatedButton.styleFrom(
                          backgroundColor: global.theme.infoHighlightTextColor,
                        ),
                        onPressed: () {
                          Navigator.pop(context);
                          context.read<ProductBloc>().add(
                            ProductDeleteMany(guid: guidListChecked),
                          );
                        },
                        child: Text(global.language('delete')),
                      ),
                    ],
                  ),
                );
                setState(() {});
              },
              icon: Icon(Icons.delete),
            ),
          if (showCheckBox == false)
            /// เพิ่มข้อมูลใหม่
            IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () {
                addNewData(true);
              },
              icon: Icon(Icons.add),
            ),
        ],
      ),
      body: Focus(
        focusNode: FocusNode(skipTraversal: true, canRequestFocus: true),
        onKey: (node, event) {
          if (kIsWeb ||
              Platform.isWindows ||
              Platform.isLinux ||
              Platform.isMacOS) {
            if (event is RawKeyDownEvent) {
              if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
                isKeyDown = false;
                int index = listData.indexOf(
                  listData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index > 0) {
                  selectGuid = listData[index - 1].guidfixed;
                  currentListIndex = index + 1;
                  isKeyUp = true;
                  loadDataFromServer(guid: selectGuid, isEdit: false);
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                isKeyUp = false;
                int index = listData.indexOf(
                  listData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                selectGuid = listData[index + 1].guidfixed;
                currentListIndex = index + 1;
                isKeyDown = true;
                loadDataFromServer(guid: selectGuid, isEdit: false);
              }
            }
          }
          return KeyEventResult.ignored;
        },
        child: Column(
          children: [
            Container(
              padding: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              color: global.theme.searchBarColor,
              child: Row(
                children: [
                  Expanded(
                    child: TextField(
                      onSubmitted: (value) {
                        searchFocusNode.requestFocus();
                      },
                      onChanged: (value) {
                        debouncer.run(() {
                          loadDataList(true, value);
                        });
                      },
                      autofocus: false,
                      focusNode: searchFocusNode,
                      controller: searchController,
                      decoration: InputDecoration(
                        isDense: true,
                        contentPadding: const EdgeInsets.symmetric(vertical: 8, horizontal: 4),
                        border: InputBorder.none,
                        hintText: global.language('search'),
                        prefixIcon: Icon(Icons.search, size: 20),
                        prefixIconConstraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                      ),
                    ),
                  ),
                  ListFontSizeControl(onChanged: () => setState(() {})),
                  const SizedBox(width: 4),
                  if (listData.isNotEmpty)
                    Text('(${listData.length})', style: TextStyle(fontSize: 11, color: global.theme.textSecondaryColor)),
                ],
              ),
            ),
            Container(color: global.theme.appBarColor, height: 6),
            Container(
              padding: EdgeInsets.only(
                left: 10,
                right: 10,
                top: global.deviceConfig.listDataLineSpace,
                bottom: global.deviceConfig.listDataLineSpace,
              ),
              color: global.theme.columnHeaderColor,
              child: Row(
                children: [
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("product_code"),
                      style: TextStyle(
                        color: global.theme.textColor,
                        fontWeight: FontWeight.bold,
                        fontSize: global.deviceConfig.listDataFontSize + 2,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 10,
                    child: Text(
                      global.language("product_name"),
                      style: TextStyle(
                        color: global.theme.textColor,
                        fontWeight: FontWeight.bold,
                        fontSize: global.deviceConfig.listDataFontSize + 2,
                      ),
                      maxLines: 1,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  if (showCheckBox)
                    Expanded(
                      flex: 1,
                      child: Icon(Icons.check, color: global.theme.textColor, size: 12),
                    ),
                ],
              ),
            ),
            Expanded(
              child: ListView.builder(
                //controller: listScrollController,
                key: const PageStorageKey("list"),
                controller: listScrollController,
                itemCount: listData.length,
                itemBuilder: (context, index) {
                  return (listData.isEmpty)
                      ? Container()
                      : listObject(index, listData[index], showCheckBox);
                },
              ),
            ),
            if (loadingData)
              Center(
                child: LoadingAnimationWidget.staggeredDotsWave(
                  color: global.theme.infoHighlightTextColor,
                  size: 50,
                ),
              ),
          ],
        ),
      ),
    );
  }

  void switchToEdit(ProductModel value) {
    tabChange(1);
    setState(() {
      selectGuid = value.guidfixed;
      loadDataFromServer(guid: selectGuid, isEdit: true);
    });
  }

  Color _getContainerColor(int index, bool isCheck, bool selected) {
    if (isCheck) return global.theme.negativeHighlightColor;
    if (selected) return global.theme.rowSelectedColor;
    if (_hoverIndex == index) return global.theme.rowHoverColor;
    return (index % 2 == 0)
        ? global.theme.columnAlternateEvenColor
        : global.theme.columnAlternateOddColor;
  }

  Widget listObject(int index, ProductModel value, bool showCheckBox) {
    bool isCheck = false;
    for (int i = 0; i < guidListChecked.length; i++) {
      if (guidListChecked[i] == value.guidfixed) {
        isCheck = true;
        break;
      }
    }
    bool selected = selectGuid == value.guidfixed;
    TextStyle textStyle = TextStyle(
      fontWeight: selected ? FontWeight.w700 : FontWeight.w400,
      color: selected ? global.theme.textColor : global.theme.textSecondaryColor,
      fontSize: (selected)
          ? global.deviceConfig.listDataFontSize + 2.0
          : global.deviceConfig.listDataFontSize,
    );
    return MouseRegion(
      onEnter: (_) => setState(() => _hoverIndex = index),
      onExit: (_) => setState(() => _hoverIndex = -1),
      child: GestureDetector(
        onTap: () {
          if (showCheckBox == true) {
            setState(() {
              selectGuid = value.guidfixed;
              if (isCheck == true) {
                guidListChecked.remove(value.guidfixed);
              } else {
                guidListChecked.add(value.guidfixed);
              }
              global.showSnackBar(
                context,
                Icon(Icons.check, color: global.theme.onPrimaryColor),
                "${global.language("chosen")} ${guidListChecked.length} ${global.language("list")}",
                global.theme.infoHighlightTextColor,
              );
            });
          } else {
            setState(() {
              selectGuid = value.guidfixed;
              isSaveAllow = false;
              loadDataFromServer(guid: selectGuid, isEdit: false);
            });
          }
        },
        onDoubleTap: () {
          if (showCheckBox == false) {
            switchToEdit(value);
          }
        },
        child: Container(
          key: index < listKeys.length ? listKeys[index] : null,
          decoration: BoxDecoration(
            color: _getContainerColor(index, isCheck, selected),
          ),
          padding: EdgeInsets.only(
            left: 10,
            right: 10,
            top: global.deviceConfig.listDataLineSpace,
            bottom: global.deviceConfig.listDataLineSpace,
          ),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Expanded(
                flex: 5,
                child: Text(
                  value.itemcode,
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                  style: textStyle,
                ),
              ),
              Expanded(
                flex: 10,
                child: Text(
                  global.packName(value.names!),
                  maxLines: 1,
                  style: textStyle,
                ),
              ),
              if (showCheckBox)
                Expanded(
                  flex: 1,
                  child: (isCheck)
                      ? Icon(Icons.check, size: 12)
                      : Container(),
                ),
            ],
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    listKeys.clear();
    if (showCheckBox == false) {
      guidListChecked.clear();
    }
    return Scaffold(
      resizeToAvoidBottomInset: true,
      body: BlocListener<ProductBloc, ProductState>(
        listener: (context, state) {
          // Load
          if (state is ProductLoadSuccess) {
            setState(() {
              loadingData = false;
              if (state.products.isNotEmpty) {
                listData.addAll(state.products);
              }
            });
          }
          if (state is ProductLoadFailed) {
            setState(() {
              loadingData = false;
            });
          }
          // Delete
          if (state is ProductDeleteSuccess) {
            setState(() {
              global.showSnackBar(
                context,
                Icon(Icons.delete, color: global.theme.onPrimaryColor),
                global.language("delete_success"),
                global.theme.infoHighlightTextColor,
              );
              listData.clear();
              WidgetsBinding.instance.addPostFrameCallback((timeStamp) {});
              loadDataList(false, searchText);
            });
          }
          // Delete Many
          if (state is ProductDeleteManySuccess) {
            setState(() {
              global.showSnackBar(
                context,
                Icon(Icons.delete, color: global.theme.onPrimaryColor),
                global.language("not_delete_success"),
                global.theme.infoHighlightTextColor,
              );
              listData.clear();
              loadDataList(false, searchText);
              showCheckBox = false;
            });
          }
        },
        child: (global.isMobileScreen(context))
            ? DefaultTabController(
                length: 2,
                child: TabBarView(
                  controller: tabController,
                  children: [listScreen(), dataEdit],
                ),
              )
            : SplitView(
                controller: splitViewController,
                gripSize: 14,
                gripColor: global.theme.appBarColor,
                gripColorActive: global.theme.infoHighlightTextColor,
                viewMode: SplitViewMode.Horizontal,
                indicator: const SplitIndicator(
                  viewMode: SplitViewMode.Horizontal,
                ),
                activeIndicator: const SplitIndicator(
                  viewMode: SplitViewMode.Horizontal,
                  isActive: true,
                ),
                children: [listScreen(), dataEdit],
              ),
      ),
    );
  }
}
