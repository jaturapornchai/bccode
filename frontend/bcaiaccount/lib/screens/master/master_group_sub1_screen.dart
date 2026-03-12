import 'dart:io';
import 'package:smlaicloud/bloc/master_group_sub1/master_group_sub1_bloc.dart';
import 'package:smlaicloud/widgets/manual_button.dart';
import 'package:smlaicloud/model/master_group_sub1_model.dart';
import 'package:smlaicloud/model/master_group_model.dart';
import 'package:smlaicloud/screens/master/group_selection_screens.dart';
import 'package:smlaicloud/widgets/list_font_size_control.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:split_view/split_view.dart';
import 'package:translator/translator.dart';

class MasterGroupSub1Screen extends StatefulWidget {
  const MasterGroupSub1Screen({super.key});

  @override
  State<MasterGroupSub1Screen> createState() => MasterGroupSub1ScreenState();
}

class MasterGroupSub1ScreenState extends State<MasterGroupSub1Screen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  final translator = GoogleTranslator();
  late TabController tabController;
  ScrollController editScrollController = ScrollController();
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  List<LanguageModel> languageList = <LanguageModel>[];
  List<TextEditingController> fieldTextController = [];
  List<global.FieldFocusModel> fieldFocusNodes = [];
  int focusNodeIndex = 0;
  List<MasterGroupSub1Model> listData = [];
  List<String> groupSub1GuidListChecked = [];
  List<LanguageDataModel> names = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  bool isChange = false;
  bool isSaveAllow = false;
  late MasterGroupSub1State blocMasterGroupSub1State;
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  int _hoverIndex = -1;
  global.ScreenEventEnum screenEvent = global.ScreenEventEnum.list;
  late SplitViewController splitViewController;
  final debouncer = global.Debouncer(1000);
  String selectedGroupMainGuid = "";
  String selectedGroupMainName = "";

  bool isLoadTranslation = false;

  void changeScreenEvent(global.ScreenEventEnum event) {
    screenEvent = event;
    for (int index = 0; index < fieldFocusNodes.length; index++) {
      fieldFocusNodes[index].isReadOnly =
          (screenEvent == global.ScreenEventEnum.list) ? true : false;
    }
    if (screenEvent == global.ScreenEventEnum.edit) {
      fieldFocusNodes[0].isReadOnly = true;
    }
    setState(() {});
  }

  void setSystemLanguageList() async {
    await global.setSystemLanguage(context);

    for (int i = 0; i < global.config.languages.length; i++) {
      if (global.config.languages[i].isuse!) {
        languageList.add(global.config.languages[i]);
      }
    }
    for (int i = 0; i < languageList.length; i++) {
      fieldTextController.add(TextEditingController());
      FocusNode focusNode = FocusNode();
      focusNode.addListener(() {
        focusNodeIndex = i;
      });
      fieldFocusNodes.add(global.FieldFocusModel(focusNode: focusNode));
    }
    loadDataList("");
  }

  @override
  void initState() {
    tabController = TabController(vsync: this, length: 2);
    tabController.addListener(() {
      setState(() {});
    });

    splitViewController = SplitViewController(
      limits: [null, WeightLimit(min: 0.2, max: 0.8)],
    );

    FocusNode focusNode = FocusNode();
    focusNode.addListener(() {
      focusNodeIndex = 0;
    });
    fieldFocusNodes.add(global.FieldFocusModel(focusNode: focusNode));
    fieldTextController.add(TextEditingController());

    // Add field for groupMainGuid
    fieldTextController.add(TextEditingController());
    FocusNode groupMainFocusNode = FocusNode();
    fieldFocusNodes.add(global.FieldFocusModel(focusNode: groupMainFocusNode));

    setSystemLanguageList();
    listScrollController.addListener(onScrollList);

    super.initState();
  }

  void loadDataList(String search) {
    context.read<MasterGroupSub1Bloc>().add(
      MasterGroupSub1LoadList(
        offset: (listData.isEmpty) ? 0 : listData.length,
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
    tabController.dispose();
    editScrollController.dispose();
    searchController.dispose();
    for (int i = 0; i < fieldTextController.length; i++) {
      fieldTextController[i].dispose();
    }
    for (int i = 0; i < fieldFocusNodes.length; i++) {
      fieldFocusNodes[i].focusNode.dispose();
    }
    super.dispose();
  }

  void clearEditData() {
    for (int i = 0; i < fieldTextController.length; i++) {
      fieldTextController[i].clear();
    }
    selectedGroupMainGuid = "";
    selectedGroupMainName = "";
    isChange = false;
    focusNodeIndex = 0;
    fieldFocusNodes[focusNodeIndex].focusNode.requestFocus();
  }

  void showGroupSelectionDialog() async {
    GroupSelectionHelper.showMasterGroupSelectionScreen(
      context: context,
      onGroupSelected: (MasterGroupModel selectedGroup) {
        setState(() {
          selectedGroupMainGuid = selectedGroup.guidfixed;
          selectedGroupMainName = global.packName(selectedGroup.names);
          fieldTextController[1].text = selectedGroup.guidfixed;
          isChange = true;
        });
      },
    );
  }

  void discardData({required Function callBack}) {
    if ((screenEvent == global.ScreenEventEnum.add ||
            screenEvent == global.ScreenEventEnum.edit) &&
        isChange) {
      showDialog<String>(
        context: context,
        builder: (BuildContext context) => AlertDialog(
          title: Text(global.language('data_editing')),
          content: Text(global.language('leave_this_screen')),
          actions: <Widget>[
            ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: global.theme.buttonNoColor,
              ),
              onPressed: () => Navigator.pop(context),
              child: Text(global.language('no')),
            ),
            ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: global.theme.buttonYesColor,
              ),
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

  void getData(String guid) {
    headerEdit = global.language("show");
    changeScreenEvent(global.ScreenEventEnum.list);
    context.read<MasterGroupSub1Bloc>().add(MasterGroupSub1Get(guid: guid));
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('group_sub1')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () {
            discardData(
              callBack: () {
                Navigator.pop(context);
                changeScreenEvent(global.ScreenEventEnum.list);
              },
            );
          },
        ),
        actions: <Widget>[
          const ManualButton(path: 'product-sub-group-1'),
          Padding(
            padding: EdgeInsets.only(right: 20.0),
            child: IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () {
                discardData(
                  callBack: () {
                    setState(() {
                      if (showCheckBox) {
                        showCheckBox = false;
                        groupSub1GuidListChecked.clear();
                      } else {
                        showCheckBox = true;
                        global.showSnackBar(
                          context,
                          Icon(Icons.delete, color: Colors.white),
                          global.language("choose_item_delete"),
                          Colors.blue,
                        );
                      }
                    });
                  },
                );
              },
              icon: (showCheckBox)
                  ? const Icon(Icons.close)
                  : const Icon(Icons.check_box),
            ),
          ),
          if (groupSub1GuidListChecked.isNotEmpty)
            Padding(
              padding: EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () {
                  showDialog<String>(
                    context: context,
                    builder: (BuildContext context) => AlertDialog(
                      title: Text(global.language('confirm_delete')),
                      actions: <Widget>[
                        ElevatedButton(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: global.theme.buttonNoColor,
                          ),
                          onPressed: () => Navigator.pop(context),
                          child: Text(global.language('no')),
                        ),
                        ElevatedButton(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: global.theme.buttonYesColor,
                          ),
                          onPressed: () {
                            Navigator.pop(context);
                            context.read<MasterGroupSub1Bloc>().add(
                              MasterGroupSub1DeleteMany(
                                guid: groupSub1GuidListChecked,
                              ),
                            );
                          },
                          child: Text(global.language('confirm')),
                        ),
                      ],
                    ),
                  );
                  setState(() {});
                },
                icon: const Icon(Icons.delete),
              ),
            ),
          Padding(
            padding: EdgeInsets.only(right: 20.0),
            child: IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () {
                discardData(
                  callBack: () {
                    setState(() {
                      changeScreenEvent(global.ScreenEventEnum.add);
                      selectGuid = "";
                      showCheckBox = false;
                      isChange = false;
                      clearEditData();
                      headerEdit = global.language("append");
                      isSaveAllow = true;
                      WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                        tabController.animateTo(1);
                        fieldFocusNodes[0].focusNode.requestFocus();
                      });
                    });
                  },
                );
              },
              icon: const Icon(Icons.add),
            ),
          ),
        ],
      ),
      body: Column(
        children: [
          Container(
            padding: const EdgeInsets.all(5),
            decoration: BoxDecoration(
              color: global.theme.searchBarColor,
              borderRadius: BorderRadius.circular(2),
            ),
            child: Row(
              children: [
                Expanded(
                  child: TextFormField(
                    onFieldSubmitted: (value) {
                      searchFocusNode.requestFocus();
                    },
                    onChanged: (value) {
                      debouncer.run(() {
                        setState(() {
                          listData = [];
                        });
                        loadDataList(value);
                      });
                    },
                    autofocus: false,
                    focusNode: searchFocusNode,
                    controller: searchController,
                    decoration: InputDecoration(
                      isDense: true,
                      contentPadding: EdgeInsets.symmetric(vertical: 8, horizontal: 4),
                      border: InputBorder.none,
                      prefixIcon: Icon(Icons.search, size: 20, color: global.theme.iconColor),
                      prefixIconConstraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                      hintText: global.language('search'),
                    ),
                  ),
                ),
                ListFontSizeControl(onChanged: () => setState(() {})),
                const SizedBox(width: 4),
                if (listData.isNotEmpty)
                  Text(
                    '(${listData.length})',
                    style: TextStyle(fontSize: 11, color: global.theme.textSecondaryColor),
                  ),
                IconButton(
                  focusNode: FocusNode(skipTraversal: true),
                  icon: const Icon(Icons.line_weight),
                  onPressed: () async {
                    setState(() {
                      global.listDataLineSpaceChange();
                    });
                  },
                ),
              ],
            ),
          ),
          Container(color: global.theme.appBarColor, height: 6),
          Container(
            color: global.theme.columnHeaderColor,
            key: headerKey,
            padding: const EdgeInsets.only(
              left: 10,
              right: 10,
              top: 5,
              bottom: 5,
            ),
            child: Row(
              children: [
                Expanded(
                  flex: 4,
                  child: Text(
                    global.language("group_sub1_code"),
                    style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                      fontWeight: FontWeight.bold,
                      fontSize: global.deviceConfig.listDataFontSize + 2,
                    ),
                  ),
                ),
                Expanded(
                  flex: 6,
                  child: Text(
                    global.language("group_sub1_name"),
                    style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                      fontWeight: FontWeight.bold,
                      fontSize: global.deviceConfig.listDataFontSize + 2,
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                Expanded(
                  flex: 6,
                  child: Text(
                    global.language("group_main_name"),
                    style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                      fontWeight: FontWeight.bold,
                      fontSize: global.deviceConfig.listDataFontSize + 2,
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
                if (showCheckBox)
                  Expanded(
                    flex: 1,
                    child: Icon(
                      Icons.check,
                      color: global.theme.columnHeaderTextColor,
                      size: 12,
                    ),
                  ),
              ],
            ),
          ),
          Expanded(
            child: RawKeyboardListener(
              autofocus: true,
              focusNode: FocusNode(skipTraversal: true, canRequestFocus: true),
              onKey: (RawKeyEvent event) {
                if (kIsWeb ||
                    Platform.isWindows ||
                    Platform.isLinux ||
                    Platform.isMacOS) {
                  if (event is RawKeyDownEvent) {
                    if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
                      isKeyUp = true;
                      isKeyDown = false;
                      if (currentListIndex > 0) {
                        currentListIndex--;
                        setState(() {
                          selectGuid = listData[currentListIndex].guidfixed;
                          getData(selectGuid);
                        });
                        scrollToListItem();
                      }
                    }
                    if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                      isKeyDown = true;
                      isKeyUp = false;
                      if (currentListIndex < listData.length - 1) {
                        currentListIndex++;
                        setState(() {
                          selectGuid = listData[currentListIndex].guidfixed;
                          getData(selectGuid);
                        });
                        scrollToListItem();
                      }
                    }
                    if (event.logicalKey == LogicalKeyboardKey.enter) {
                      if (currentListIndex >= 0 &&
                          currentListIndex < listData.length) {
                        switchToEdit(listData[currentListIndex]);
                      }
                    }
                    if (event.logicalKey == LogicalKeyboardKey.f2) {
                      setState(() {
                        changeScreenEvent(global.ScreenEventEnum.add);
                        selectGuid = "";
                        showCheckBox = false;
                        isChange = false;
                        clearEditData();
                        headerEdit = global.language("append");
                        isSaveAllow = true;
                        WidgetsBinding.instance.addPostFrameCallback((
                          timeStamp,
                        ) {
                          tabController.animateTo(1);
                          fieldFocusNodes[0].focusNode.requestFocus();
                        });
                      });
                    }
                  }
                }
              },
              child: ListView(
                controller: listScrollController,
                children: listData
                    .map(
                      (value) => listObject(
                        listData.indexOf(value),
                        value,
                        showCheckBox,
                      ),
                    )
                    .toList(),
              ),
            ),
          ),
        ],
      ),
    );
  }

  void scrollToListItem() {
    if (currentListIndex >= 0 && currentListIndex < listKeys.length) {
      WidgetsBinding.instance.addPostFrameCallback((_) {
        final RenderBox? renderBox =
            listKeys[currentListIndex].currentContext?.findRenderObject()
                as RenderBox?;
        if (renderBox != null) {
          final position = renderBox.localToGlobal(Offset.zero);
          final RenderBox? headerRenderBox =
              headerKey.currentContext?.findRenderObject() as RenderBox?;
          double headerHeight = 0;
          if (headerRenderBox != null) {
            headerHeight = headerRenderBox.size.height;
          }

          double screenHeight = queryData.size.height;
          double itemTop = position.dy;
          double itemHeight = renderBox.size.height;
          double visibleTop = headerHeight + 100;
          double visibleBottom = screenHeight - 100;

          if (itemTop < visibleTop || itemTop + itemHeight > visibleBottom) {
            double targetOffset =
                listScrollController.offset + (itemTop - visibleTop);
            if (isKeyUp && itemTop < visibleTop) {
              targetOffset =
                  listScrollController.offset - (visibleTop - itemTop);
            }

            listScrollController.animateTo(
              targetOffset.clamp(
                0.0,
                listScrollController.position.maxScrollExtent,
              ),
              duration: const Duration(milliseconds: 200),
              curve: Curves.easeInOut,
            );
          }
        }
      });
    }
  }

  void switchToEdit(MasterGroupSub1Model value) {
    setState(() {
      selectGuid = value.guidfixed;
      getData(selectGuid);
      headerEdit = global.language("edit");
      isSaveAllow = true;
      changeScreenEvent(global.ScreenEventEnum.edit);
    });
  }

  Widget listObject(int index, MasterGroupSub1Model value, bool showCheckBox) {
    bool isCheck = false;
    for (int i = 0; i < groupSub1GuidListChecked.length; i++) {
      if (groupSub1GuidListChecked[i] == value.guidfixed) {
        isCheck = true;
        break;
      }
    }
    // ลบการ add key ออก - keys ถูกสร้างใน LoadSuccess แล้ว
    // listKeys.add(GlobalKey());  // ❌ ตรงนี้ทำให้เกิด infinite loop!
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
        if (showCheckBox == true) {
          setState(() {
            selectGuid = value.guidfixed;
            if (isCheck == true) {
              groupSub1GuidListChecked.remove(value.guidfixed);
            } else {
              groupSub1GuidListChecked.add(value.guidfixed);
            }
            global.showSnackBar(
              context,
              Icon(Icons.check, color: Colors.white),
              "${global.language("chosen")} ${groupSub1GuidListChecked.length} ${global.language("list")}",
              Colors.blue,
            );
          });
        } else {
          setState(() {
            discardData(
              callBack: () {
                isSaveAllow = false;
                changeScreenEvent(global.ScreenEventEnum.list);
                selectGuid = value.guidfixed;
                getData(selectGuid);
                WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                  tabController.animateTo(1);
                });
              },
            );
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
          color: (selectGuid == value.guidfixed)
              ? global.theme.rowSelectedColor
              : (index % 2 == 0)
              ? global.theme.columnAlternateEvenColor
              : global.theme.columnAlternateOddColor,
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
              flex: 4,
              child: Text(
                value.code,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 6,
              child: Text(
                global.packName(value.names),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 6,
              child: Text(
                global.packName(value.groupMainNames),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            if (showCheckBox)
              Expanded(
                flex: 1,
                child: (isCheck)
                    ? Icon(
                        Icons.check,
                        size: global.deviceConfig.listDataFontSize,
                      )
                    : Container(),
              ),
          ],
        ),
      ),
    ),
    );
  }


  Color _getContainerColor(String itemGuid, int index) {
    if (selectGuid.isNotEmpty && selectGuid == itemGuid) {
      return (screenEvent == global.ScreenEventEnum.edit)
          ? global.theme.rowEditColor
          : global.theme.rowSelectedColor;
    }
    if (_hoverIndex == index) {
      return global.theme.rowHoverColor;
    }
    return (index % 2 == 0)
        ? global.theme.columnAlternateEvenColor
        : global.theme.columnAlternateOddColor;
  }

  List<LanguageDataModel> packLanguage() {
    List<LanguageDataModel> names = [];
    for (int i = 0; i < languageList.length; i++) {
      if (languageList[i].code!.trim().isNotEmpty &&
          fieldTextController[i + 2].text.trim().isNotEmpty) {
        names.add(
          LanguageDataModel(
            code: languageList[i].code!,
            name: fieldTextController[i + 2].text,
          ),
        );
      }
    }
    return names;
  }

  bool verifyData(MasterGroupSub1Model value) {
    List<String> errorList = [];
    if (value.code.isEmpty) {
      errorList.add(global.language("code"));
    }
    if (value.groupMainGuid.isEmpty) {
      errorList.add(global.language("group_main_guid"));
    }
    if (value.names.isEmpty || value.names[0].name.isEmpty) {
      errorList.add(global.language("name"));
    }
    if (errorList.isNotEmpty) {
      global.showSnackBar(
        context,
        Icon(Icons.error, color: Colors.white),
        "${global.language("not_success_save")} ${errorList.join(", ")}",
        Colors.red,
      );
      return false;
    }
    return true;
  }

  void saveOrUpdateData() {
    if (screenEvent == global.ScreenEventEnum.edit ||
        screenEvent == global.ScreenEventEnum.add) {
      MasterGroupSub1Model groupSub1 = MasterGroupSub1Model(
        guidfixed: "",
        code: fieldTextController[0].text,
        groupMainGuid: fieldTextController[1].text,
        names: packLanguage(),
        groupMainNames: [],
      );
      if (verifyData(groupSub1)) {
        showCheckBox = false;
        if (selectGuid.trim().isEmpty) {
          context.read<MasterGroupSub1Bloc>().add(
            MasterGroupSub1Save(groupSub1Model: groupSub1),
          );
        } else {
          context.read<MasterGroupSub1Bloc>().add(
            MasterGroupSub1Update(guid: selectGuid, groupSub1Model: groupSub1),
          );
        }
      }
    }
  }

  void getDataToEditScreen(MasterGroupSub1Model groupSub1) {
    isChange = false;
    selectGuid = groupSub1.guidfixed;
    fieldTextController[0].text = groupSub1.code;
    fieldTextController[1].text = groupSub1.groupMainGuid;
    selectedGroupMainGuid = groupSub1.groupMainGuid;
    selectedGroupMainName = global.packName(groupSub1.groupMainNames);
    for (int i = 0; i < languageList.length; i++) {
      fieldTextController[i + 2].text = "";
    }
    for (int i = 0; i < languageList.length; i++) {
      for (int j = 0; j < groupSub1.names.length; j++) {
        if (languageList[i].code! == groupSub1.names[j].code) {
          fieldTextController[i + 2].text = groupSub1.names[j].name;
        }
      }
    }
  }

  void findFocusNext(int index) {
    focusNodeIndex = index + 1;
    if (focusNodeIndex > fieldFocusNodes.length - 1) {
      focusNodeIndex = 0;
    }
    while (true) {
      if (fieldFocusNodes[focusNodeIndex].isReadOnly == true) {
        focusNodeIndex++;
        if (focusNodeIndex > fieldFocusNodes.length - 1) {
          break;
        }
      } else {
        fieldFocusNodes[focusNodeIndex].focusNode.requestFocus();
        fieldTextController[focusNodeIndex]
            .selection = TextSelection.fromPosition(
          TextPosition(offset: fieldTextController[focusNodeIndex].text.length),
        );
        break;
      }
    }
  }

  Widget editScreen({mobileScreen}) {
    return Scaffold(
      backgroundColor: (screenEvent == global.ScreenEventEnum.edit || screenEvent == global.ScreenEventEnum.add) ? global.theme.toolBarEditModeColor.withValues(alpha: 0.05) : global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor:
            (screenEvent == global.ScreenEventEnum.edit ||
                screenEvent == global.ScreenEventEnum.add)
            ? global.theme.toolBarEditModeColor
            : global.theme.appBarColor,
        automaticallyImplyLeading: false,
        leading: mobileScreen
            ? IconButton(
                focusNode: FocusNode(skipTraversal: true),
                icon: const Icon(Icons.arrow_back),
                onPressed: () async {
                  showCheckBox = false;
                  discardData(
                    callBack: () {
                      changeScreenEvent(global.ScreenEventEnum.list);
                      WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                        tabController.animateTo(0);
                      });
                    },
                  );
                },
              )
            : null,
        title: Text(headerEdit + global.language("group_sub1")),
        actions: <Widget>[
          if (selectGuid.isNotEmpty)
            Padding(
              padding: EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () {
                  showCheckBox = false;
                  showDialog<String>(
                    context: context,
                    builder: (BuildContext context) => AlertDialog(
                      title: Text(global.language('delete_confirm')),
                      actions: <Widget>[
                        ElevatedButton(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: global.theme.buttonNoColor,
                          ),
                          onPressed: () => Navigator.pop(context),
                          child: Text(global.language('no')),
                        ),
                        ElevatedButton(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: global.theme.buttonYesColor,
                          ),
                          onPressed: () {
                            Navigator.pop(context);
                            context.read<MasterGroupSub1Bloc>().add(
                              MasterGroupSub1Delete(guid: selectGuid),
                            );
                          },
                          child: Text(global.language('confirm')),
                        ),
                      ],
                    ),
                  );
                },
                icon: const Icon(Icons.delete),
              ),
            ),
          if (isSaveAllow == false && selectGuid.trim().isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () {
                  showCheckBox = false;
                  switchToEdit(
                    listData[listData.indexOf(
                      listData.firstWhere(
                        (element) => element.guidfixed == selectGuid,
                      ),
                    )],
                  );
                },
                icon: const Icon(Icons.edit),
              ),
            ),
          if ((screenEvent == global.ScreenEventEnum.edit ||
                  screenEvent == global.ScreenEventEnum.add) &&
              global.systemLanguage.length > 1)
            Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () async {
                  setState(() {
                    isLoadTranslation = true;
                  });
                  for (int i = 2; i <= languageList.length + 1; i++) {
                    try {
                      var translation = await translator.translate(
                        fieldTextController[2].text,
                        to: languageList[i - 2].codeTranslator!,
                      );
                      if (fieldTextController[i].text.isEmpty) {
                        fieldTextController[i].text = translation.text;
                      }
                    } catch (_) {}
                  }
                  setState(() {
                    isLoadTranslation = false;
                  });
                },
                icon: const Icon(Icons.translate),
              ),
            ),
          if (isSaveAllow == true)
            Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () => saveOrUpdateData(),
                icon: const Icon(Icons.save),
              ),
            ),
        ],
      ),
      body: RawKeyboardListener(
        focusNode: FocusNode(skipTraversal: true),
        onKey: (event) {
          if (kIsWeb ||
              Platform.isWindows ||
              Platform.isLinux ||
              Platform.isMacOS) {
            if (event is RawKeyUpEvent) {
              if (event.logicalKey == LogicalKeyboardKey.f10) {
                saveOrUpdateData();
              }
            }
          }
        },
        child: SingleChildScrollView(
          controller: editScrollController,
          child: Container(
            color: global.theme.cardColor,
            width: double.infinity,
            padding: const EdgeInsets.all(10),
            child: Column(
              children: [
                      const SizedBox(height: 10),
                      TextFormField(
                        inputFormatters: <TextInputFormatter>[
                          FilteringTextInputFormatter.allow(RegExp('[a-zA-Z0-9-]')),
                          FilteringTextInputFormatter.deny(
                            ' ',
                          ), // Exclude space character
                        ],
                        readOnly: fieldFocusNodes[0].isReadOnly,
                        onFieldSubmitted: (value) {
                          findFocusNext(0);
                        },
                        textInputAction: TextInputAction.next,
                        focusNode: fieldFocusNodes[0].focusNode,
                        textAlign: TextAlign.left,
                        controller: fieldTextController[0],
                        textCapitalization: TextCapitalization.characters,
                        onChanged: (value) {
                          isChange = true;
                          fieldTextController[0].value = TextEditingValue(
                            text: value.toUpperCase(),
                            selection: fieldTextController[0].selection,
                          );
                        },
                        decoration: InputDecoration(
                          border: OutlineInputBorder(),
                          contentPadding: EdgeInsets.only(
                            left: 10,
                            top: 0,
                            bottom: 0,
                            right: 10,
                          ),
                          enabledBorder: OutlineInputBorder(
                            borderSide: BorderSide(color: global.theme.dividerBorderColor, width: 0.0),
                          ),
                          floatingLabelBehavior: FloatingLabelBehavior.always,
                          labelText: global.language("group_sub1_code"),
                          labelStyle: TextStyle(
                            color: global.theme.inputTextBoxForceColor,
                          ),
                        ),
                      ),
                      const SizedBox(height: 15),
                      // Group Main Selection Field
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            global.language("group_main"),
                            style: TextStyle(
                              color: global.theme.inputTextBoxForceColor,
                              fontSize: 12,
                            ),
                          ),
                          SizedBox(height: 8),
                          Container(
                            decoration: BoxDecoration(
                              border: Border.all(color: global.theme.textSecondaryColor),
                              borderRadius: BorderRadius.circular(4),
                            ),
                            child: Row(
                              children: [
                                Expanded(
                                  child: Container(
                                    padding: const EdgeInsets.symmetric(
                                      horizontal: 12,
                                      vertical: 16,
                                    ),
                                    child: Text(
                                      selectedGroupMainName.isEmpty
                                          ? global.language("select_group_main")
                                          : selectedGroupMainName,
                                      style: TextStyle(
                                        color: selectedGroupMainName.isEmpty
                                            ? global.theme.textSecondaryColor
                                            : Colors.black,
                                        fontSize: 16,
                                      ),
                                    ),
                                  ),
                                ),
                                Container(width: 1, height: 40, color: global.theme.textSecondaryColor),
                                if (selectedGroupMainGuid.isNotEmpty &&
                                    !fieldFocusNodes[1].isReadOnly)
                                  InkWell(
                                    onTap: () {
                                      setState(() {
                                        selectedGroupMainGuid = "";
                                        selectedGroupMainName = "";
                                        fieldTextController[1].text = "";
                                        isChange = true;
                                      });
                                    },
                                    child: Container(
                                      padding: const EdgeInsets.all(12),
                                      child: Icon(
                                        Icons.clear,
                                        color: Colors.red,
                                        size: 20,
                                      ),
                                    ),
                                  ),
                                InkWell(
                                  onTap: fieldFocusNodes[1].isReadOnly
                                      ? null
                                      : showGroupSelectionDialog,
                                  child: Container(
                                    padding: EdgeInsets.all(12),
                                    child: Icon(
                                      Icons.search,
                                      color: fieldFocusNodes[1].isReadOnly
                                          ? Colors.grey
                                          : global.theme.appBarColor,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                      const SizedBox(height: 15),
                      for (int i = 0; i < languageList.length; i++)
                        Padding(
                          padding: const EdgeInsets.only(bottom: 15),
                          child: TextFormField(
                            readOnly: fieldFocusNodes[i + 2].isReadOnly,
                            onChanged: (value) {
                              isChange = true;
                            },
                            onFieldSubmitted: (value) {
                              findFocusNext(i + 2);
                            },
                            textInputAction: TextInputAction.next,
                            focusNode: fieldFocusNodes[i + 2].focusNode,
                            textAlign: TextAlign.left,
                            controller: fieldTextController[i + 2],
                            decoration: InputDecoration(
                              contentPadding: EdgeInsets.only(
                                left: 10,
                                top: 0,
                                bottom: 0,
                                right: 10,
                              ),
                              enabledBorder: OutlineInputBorder(
                                borderSide: BorderSide(
                                  color: global.theme.textSecondaryColor,
                                  width: 0.0,
                                ),
                              ),
                              floatingLabelBehavior: FloatingLabelBehavior.always,
                              labelStyle: TextStyle(
                                color: global.theme.inputTextBoxColor,
                              ),
                              border: const OutlineInputBorder(),
                              labelText:
                                  "${global.language("group_sub1_name")} (${languageList[i].name})",
                              suffixIcon: isLoadTranslation
                                  ? const Padding(
                                      padding: EdgeInsets.all(8.0),
                                      child: SizedBox(
                                        width: 20,
                                        height: 20,
                                        child: CircularProgressIndicator(
                                          strokeWidth: 2,
                                        ),
                                      ),
                                    )
                                  : null,
                            ),
                          ),
                        ),
                if (isSaveAllow)
                  SizedBox(
                    width: double.infinity,
                    child: ElevatedButton.icon(
                      style: ElevatedButton.styleFrom(
                        backgroundColor: global.theme.buttonColor,
                      ),
                      focusNode: FocusNode(skipTraversal: true),
                      onPressed: () {
                        saveOrUpdateData();
                      },
                      icon: Icon(Icons.save),
                      label: Text(
                        global.language("save") +
                            ((kIsWeb ||
                                    Platform.isWindows ||
                                    Platform.isLinux ||
                                    Platform.isMacOS)
                                ? " (F10)"
                                : ""),
                      ),
                    ),
                  ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    queryData = MediaQuery.of(context);
    listKeys.clear();
    if (showCheckBox == false) {
      groupSub1GuidListChecked.clear();
    }
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return BlocListener<MasterGroupSub1Bloc, MasterGroupSub1State>(
            listener: (context, state) {
              blocMasterGroupSub1State = state;
              if (state is MasterGroupSub1LoadSuccess) {
                setState(() {
                  if (state.groupSub1s.isNotEmpty) {
                    listData.addAll(state.groupSub1s);
                  }
                });
              }
              if (state is MasterGroupSub1SaveSuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.save, color: Colors.white),
                    global.language("save_success"),
                    global.theme.primaryColor,
                  );
                  clearEditData();
                  listData.clear();
                  loadDataList(searchText);
                });
              }
              if (state is MasterGroupSub1SaveFailed) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.save, color: Colors.white),
                    "${global.language("not_success_save")} : ${state.message}",
                    Colors.red,
                  );
                });
              }
              if (state is MasterGroupSub1UpdateSuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.edit, color: Colors.white),
                    global.language("edit_success"),
                    global.theme.primaryColor,
                  );
                  clearEditData();
                  listData.clear();
                  WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                    tabController.animateTo(0);
                  });
                  loadDataList(searchText);
                  isSaveAllow = false;
                  getData(selectGuid);
                });
              }
              if (state is MasterGroupSub1UpdateFailed) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.edit, color: Colors.white),
                    "${global.language("not_edit_success")} : ${state.message}",
                    Colors.red,
                  );
                });
              }
              if (state is MasterGroupSub1DeleteSuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.delete, color: Colors.white),
                    global.language("delete_success"),
                    global.theme.primaryColor,
                  );
                  listData.clear();
                  clearEditData();
                  WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                    tabController.animateTo(0);
                  });
                  loadDataList(searchText);
                });
              }
              if (state is MasterGroupSub1DeleteManySuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.delete, color: Colors.white),
                    global.language("delete_success"),
                    global.theme.primaryColor,
                  );
                  listData.clear();
                  clearEditData();
                  loadDataList(searchText);
                  showCheckBox = false;
                });
              }
              if (state is MasterGroupSub1GetSuccess) {
                setState(() {
                  getDataToEditScreen(state.groupSub1);
                  if (screenEvent == global.ScreenEventEnum.edit) {
                    WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                      tabController.animateTo(1);
                    });
                    setState(() {
                      findFocusNext(0);
                    });
                  }
                });
              }
            },
            child: (constraints.maxWidth > 800)
                ? SplitView(
                    controller: splitViewController,
                    gripSize: 14,
                    gripColor: global.theme.appBarColor,
                    gripColorActive: global.theme.primaryColor,
                    viewMode: SplitViewMode.Horizontal,
                    indicator: const SplitIndicator(
                      viewMode: SplitViewMode.Horizontal,
                    ),
                    activeIndicator: const SplitIndicator(
                      viewMode: SplitViewMode.Horizontal,
                      isActive: true,
                    ),
                    children: [
                      listScreen(mobileScreen: false),
                      editScreen(mobileScreen: false),
                    ],
                  )
                : TabBarView(
                    physics: const NeverScrollableScrollPhysics(),
                    controller: tabController,
                    children: [
                      listScreen(mobileScreen: true),
                      editScreen(mobileScreen: true),
                    ],
                  ),
          );
        },
      ),
    );
  }
}
