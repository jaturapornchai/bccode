import 'dart:io';
import 'package:smlaicloud/widgets/manual_button.dart';
import 'package:smlaicloud/bloc/cost_center/cost_center_bloc.dart';
import 'package:smlaicloud/model/cost_center_model.dart';
import 'package:smlaicloud/widgets/list_font_size_control.dart';
import 'package:smlaicloud/widgets/edit_font_size_control.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/utils/focus_utils.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:split_view/split_view.dart';
import 'package:translator/translator.dart';

class CostCenterScreen extends StatefulWidget {
  const CostCenterScreen({super.key});

  @override
  State<CostCenterScreen> createState() => CostCenterScreenState();
}

class CostCenterScreenState extends State<CostCenterScreen>
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
  List<CostCenterModel> listData = [];
  List<String> costCenterGuidListChecked = [];
  List<LanguageDataModel> names = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectCode = "";
  bool isChange = false;
  bool isSaveAllow = false;
  late CostCenterState blocCostCenterState;
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  global.ScreenEventEnum screenEvent = global.ScreenEventEnum.list;
  late SplitViewController splitViewController;
  final debouncer = global.Debouncer(1000);
  int _hoverIndex = -1;

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
    setState(() {});
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
    // เรียงลำดับ Focus
    // Focus รหัส
    FocusNode focusNode = FocusNode();
    focusNode.addListener(() {
      focusNodeIndex = 0;
    });
    fieldFocusNodes.add(global.FieldFocusModel(focusNode: focusNode));
    fieldTextController.add(TextEditingController());

    setSystemLanguageList();
    super.initState();

    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (mounted) {
        context.read<CostCenterBloc>().add(
              const CostCenterLoadList(offset: 0, limit: 999999, search: ""),
            );
      }
    });
  }

  void searchDataList(String searchTerm) {
    setState(() {
      listData = listData
          .where(
            (item) =>
                item.names.any(
                  (n) => n.name.contains(searchTerm),
                ) ||
                item.code.toString().contains(searchTerm),
          )
          .toList();
    });
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
    isChange = false;
    focusNodeIndex = 0;
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
    context.read<CostCenterBloc>().add(CostCenterGet(guid: guid));
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('cost_center')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: Icon(Icons.arrow_back),
          onPressed: () {
            discardData(
              callBack: () {
                global.gotoMainMenu(context);
                changeScreenEvent(global.ScreenEventEnum.list);
              },
            );
          },
        ),
        actions: <Widget>[
          IconButton(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () {
              discardData(
                callBack: () {
                  setState(() {
                    if (showCheckBox) {
                      showCheckBox = false;
                      costCenterGuidListChecked.clear();
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
          if (costCenterGuidListChecked.isNotEmpty)
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
                          context.read<CostCenterBloc>().add(
                                CostCenterDeleteMany(
                                    guid: costCenterGuidListChecked),
                              );
                        },
                        child: Text(global.language('confirm')),
                      ),
                    ],
                  ),
                );
                setState(() {});
              },
              icon: Icon(Icons.delete),
            ),
          IconButton(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () {
              discardData(
                callBack: () {
                  setState(() {
                    changeScreenEvent(global.ScreenEventEnum.add);
                    selectCode = "";
                    showCheckBox = false;
                    isChange = false;
                    clearEditData();
                    headerEdit = global.language("append");
                    isSaveAllow = true;
                    WidgetsBinding.instance.addPostFrameCallback((
                      timeStamp,
                    ) {
                      tabController.animateTo(1);
                      focusFirstTextField(context);
                    });
                  });
                },
              );
            },
            icon: Icon(Icons.add),
          ),
          const ManualButton(path: 'settings-cost-center'),
        ],
      ),
      body: Column(
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
                        context.read<CostCenterBloc>().add(
                              CostCenterLoadList(
                                  offset: 0,
                                  limit: 999999,
                                  search: value),
                            );
                      });
                    },
                    autofocus: false,
                    focusNode: searchFocusNode,
                    controller: searchController,
                    decoration: InputDecoration(
                      isDense: true,
                      contentPadding: const EdgeInsets.symmetric(
                          vertical: 8, horizontal: 4),
                      border: InputBorder.none,
                      hintText: global.language('search'),
                      prefixIcon: Icon(Icons.search, size: 20),
                      prefixIconConstraints:
                          const BoxConstraints(minWidth: 32, minHeight: 32),
                    ),
                  ),
                ),
                ListFontSizeControl(onChanged: () => setState(() {})),
                const SizedBox(width: 4),
                if (listData.isNotEmpty)
                  Text('(${listData.length})',
                      style: TextStyle(
                          fontSize: 11,
                          color: global.theme.textSecondaryColor)),
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
                  flex: 5,
                  child: Text(
                    global.language("cost_center_code"),
                    style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                      fontWeight: FontWeight.bold,
                      fontSize: global.deviceConfig.listDataFontSize + 2,
                    ),
                  ),
                ),
                Expanded(
                  flex: 10,
                  child: Text(
                    global.language("cost_center_name"),
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
              focusNode:
                  FocusNode(skipTraversal: true, canRequestFocus: true),
              onKey: (RawKeyEvent event) {
                if (screenEvent == global.ScreenEventEnum.list) {
                  if (kIsWeb ||
                      Platform.isWindows ||
                      Platform.isLinux ||
                      Platform.isMacOS) {
                    if (event is RawKeyUpEvent) {
                      try {
                        if (event.logicalKey == LogicalKeyboardKey.f2) {
                          isKeyDown = false;
                          searchFocusNode.requestFocus();
                        }
                        if (event.logicalKey ==
                            LogicalKeyboardKey.arrowUp) {
                          isKeyDown = false;
                          int index = listData.indexOf(
                            listData.firstWhere(
                              (element) =>
                                  element.guidfixed == selectCode,
                            ),
                          );
                          if (index > 0) {
                            selectCode =
                                listData[index - 1].guidfixed;
                            currentListIndex = index - 1;
                            isKeyUp = true;
                            getData(selectCode);
                          }
                        }
                        if (event.logicalKey ==
                            LogicalKeyboardKey.arrowDown) {
                          isKeyUp = false;
                          int index = listData.indexOf(
                            listData.firstWhere(
                              (element) =>
                                  element.guidfixed == selectCode,
                            ),
                          );
                          selectCode =
                              listData[index + 1].guidfixed;
                          currentListIndex = index + 1;
                          isKeyDown = true;
                          getData(selectCode);
                        }
                      } catch (_) {}
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

  void switchToEdit(CostCenterModel value) {
    setState(() {
      selectCode = value.guidfixed;
      getData(selectCode);
      headerEdit = global.language("edit");
      isSaveAllow = true;
      changeScreenEvent(global.ScreenEventEnum.edit);

      if (screenEvent == global.ScreenEventEnum.edit) {
        WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
          tabController.animateTo(1);
          focusFirstTextField(context);
        });
      }
    });
  }

  Color _getContainerColor(String itemGuid, int index) {
    if (selectCode.isNotEmpty && selectCode == itemGuid) {
      return isSaveAllow
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

  Widget listObject(int index, CostCenterModel value, bool showCheckBox) {
    bool isCheck = false;
    for (int i = 0; i < costCenterGuidListChecked.length; i++) {
      if (costCenterGuidListChecked[i] == value.guidfixed) {
        isCheck = true;
        break;
      }
    }
    final isSelected = selectCode == value.guidfixed;
    TextStyle textStyle = isSelected
        ? TextStyle(
            fontSize: global.deviceConfig.listDataFontSize,
            fontWeight: FontWeight.w700,
            color: global.theme.textColor)
        : TextStyle(
            fontSize: global.deviceConfig.listDataFontSize,
            fontWeight: FontWeight.w400,
            color: global.theme.textSecondaryColor);
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      onEnter: (_) => setState(() => _hoverIndex = index),
      onExit: (_) => setState(() => _hoverIndex = -1),
      child: GestureDetector(
        onTap: () {
          if (showCheckBox == true) {
            setState(() {
              selectCode = value.guidfixed;
              if (isCheck == true) {
                costCenterGuidListChecked.remove(value.guidfixed);
              } else {
                costCenterGuidListChecked.add(value.guidfixed);
              }
              global.showSnackBar(
                context,
                Icon(Icons.check, color: global.theme.onPrimaryColor),
                "${global.language("chosen")} ${costCenterGuidListChecked.length} ${global.language("list")}",
                global.theme.infoHighlightTextColor,
              );
            });
          } else {
            setState(() {
              discardData(
                callBack: () {
                  isSaveAllow = false;
                  changeScreenEvent(global.ScreenEventEnum.list);
                  selectCode = value.guidfixed;
                  getData(selectCode);
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
            color: _getContainerColor(value.guidfixed, index),
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
                  value.code,
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                  style: textStyle,
                ),
              ),
              Expanded(
                flex: 10,
                child: Text(
                  global.packName(value.names),
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

  List<LanguageDataModel> packLanguage() {
    List<LanguageDataModel> names = [];
    for (int i = 0; i < languageList.length; i++) {
      if (languageList[i].code!.trim().isNotEmpty &&
          fieldTextController[i + 1].text.trim().isNotEmpty) {
        names.add(
          LanguageDataModel(
            code: languageList[i].code!,
            name: fieldTextController[i + 1].text,
          ),
        );
      }
    }
    return names;
  }

  bool verifyData(CostCenterModel value) {
    List<String> errorList = [];
    if (value.code.isEmpty) {
      errorList.add(global.language("code"));
    }
    if (value.names.isEmpty || value.names[0].name.isEmpty) {
      errorList.add(global.language("name"));
    }
    if (errorList.isNotEmpty) {
      global.showSnackBar(
        context,
        Icon(Icons.error, color: global.theme.onPrimaryColor),
        "${global.language("not_success_save")} ${errorList.join(", ")}",
        global.theme.negativeHighlightTextColor,
      );
      return false;
    }
    return true;
  }

  void saveOrUpdateData() {
    if (screenEvent == global.ScreenEventEnum.edit ||
        screenEvent == global.ScreenEventEnum.add) {
      CostCenterModel costCenterModel = CostCenterModel(
        guidfixed: selectCode,
        code: fieldTextController[0].text,
        names: packLanguage(),
      );

      if (verifyData(costCenterModel)) {
        showCheckBox = false;
        if (screenEvent == global.ScreenEventEnum.add) {
          context
              .read<CostCenterBloc>()
              .add(CostCenterSave(costCenter: costCenterModel));
        } else {
          context.read<CostCenterBloc>().add(
                CostCenterUpdate(
                    guid: selectCode, costCenter: costCenterModel),
              );
        }
      }
    }
  }

  void getDataToEditScreen(CostCenterModel costCenterModel) {
    isChange = false;
    selectCode = costCenterModel.guidfixed;
    fieldTextController[0].text = costCenterModel.code;
    for (int i = 0; i < languageList.length; i++) {
      fieldTextController[i + 1].text = "";
    }
    for (int i = 0; i < languageList.length; i++) {
      for (int j = 0; j < costCenterModel.names.length; j++) {
        if (languageList[i].code! == costCenterModel.names[j].code) {
          fieldTextController[i + 1].text = costCenterModel.names[j].name;
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
        focusAndCursorToEnd(fieldFocusNodes[focusNodeIndex].focusNode);
        break;
      }
    }
  }

  Widget editScreen({mobileScreen}) {
    return Scaffold(
      backgroundColor: global.theme.cardColor,
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
                icon: Icon(Icons.arrow_back),
                onPressed: () async {
                  showCheckBox = false;
                  discardData(
                    callBack: () {
                      changeScreenEvent(global.ScreenEventEnum.list);
                      WidgetsBinding.instance
                          .addPostFrameCallback((timeStamp) {
                        tabController.animateTo(0);
                      });
                    },
                  );
                },
              )
            : null,
        title: Text(headerEdit + global.language("cost_center")),
        actions: <Widget>[
          EditFontSizeControl(onChanged: () => setState(() {})),
          const SizedBox(width: 8),
          if (selectCode.isNotEmpty)
            IconButton(
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
                          context.read<CostCenterBloc>().add(
                                CostCenterDelete(guid: selectCode),
                              );
                        },
                        child: Text(global.language('confirm')),
                      ),
                    ],
                  ),
                );
                setState(() {});
              },
              icon: Icon(Icons.delete),
            ),
          if (isSaveAllow == false && selectCode.trim().isNotEmpty)
            IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () {
                showCheckBox = false;
                switchToEdit(
                  listData[listData.indexOf(
                    listData.firstWhere(
                      (element) => element.guidfixed == selectCode,
                    ),
                  )],
                );
              },
              icon: Icon(Icons.edit),
            ),
          if (isSaveAllow == true)
            IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () => saveOrUpdateData(),
              icon: Icon(Icons.save),
            ),
        ],
      ),
      body: Focus(
        skipTraversal: true,
        onKeyEvent: (node, event) {
          if (event is KeyUpEvent) return KeyEventResult.ignored;
          if (event.logicalKey == LogicalKeyboardKey.f10) {
            if (event is KeyDownEvent) saveOrUpdateData();
            return KeyEventResult.handled;
          }
          if (event.logicalKey == LogicalKeyboardKey.enter) {
            if (event is KeyDownEvent) {
              FocusManager.instance.primaryFocus?.nextFocus();
            }
            return KeyEventResult.handled;
          }
          return KeyEventResult.ignored;
        },
        child: FocusTraversalGroup(
          policy: TextFieldTraversalPolicy(),
          child: Builder(
            builder: (context) {
              final scale = global.editFontScaleFactor;
              return MediaQuery(
                data: MediaQuery.of(context).copyWith(
                  textScaler: TextScaler.linear(scale),
                ),
                child: Theme(
                  data: Theme.of(context).copyWith(
                    inputDecorationTheme:
                        Theme.of(context).inputDecorationTheme.copyWith(
                              contentPadding: EdgeInsets.fromLTRB(
                                12 * scale,
                                20 * scale,
                                12 * scale,
                                12 * scale,
                              ),
                            ),
                  ),
                  child: IconTheme(
                    data: IconTheme.of(context).copyWith(
                      size: 24.0 * scale,
                    ),
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
                              readOnly: fieldFocusNodes[0].isReadOnly,
                              onFieldSubmitted: (value) {
                                findFocusNext(0);
                              },
                              textInputAction: TextInputAction.next,
                              focusNode: fieldFocusNodes[0].focusNode,
                              textAlign: TextAlign.left,
                              controller: fieldTextController[0],
                              textCapitalization:
                                  TextCapitalization.characters,
                              inputFormatters: <TextInputFormatter>[
                                FilteringTextInputFormatter.allow(
                                  RegExp('[a-z A-Z 0-9 -]'),
                                ),
                              ],
                              onChanged: (value) {
                                isChange = true;
                                fieldTextController[0].value =
                                    TextEditingValue(
                                  text: value.toUpperCase(),
                                  selection:
                                      fieldTextController[0].selection,
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
                                  borderSide: BorderSide(
                                      color:
                                          global.theme.dividerBorderColor,
                                      width: 0.0),
                                ),
                                floatingLabelBehavior:
                                    FloatingLabelBehavior.always,
                                labelText:
                                    global.language("cost_center_code"),
                                labelStyle: TextStyle(
                                  color:
                                      global.theme.inputTextBoxForceColor,
                                ),
                              ),
                            ),
                            const SizedBox(height: 15),
                            for (int i = 0;
                                i < languageList.length;
                                i++)
                              Padding(
                                padding:
                                    const EdgeInsets.only(bottom: 15),
                                child: TextFormField(
                                  readOnly:
                                      fieldFocusNodes[i + 1].isReadOnly,
                                  onChanged: (value) {
                                    isChange = true;
                                  },
                                  onFieldSubmitted: (value) {
                                    findFocusNext(i + 1);
                                  },
                                  textInputAction: TextInputAction.next,
                                  focusNode:
                                      fieldFocusNodes[i + 1].focusNode,
                                  textAlign: TextAlign.left,
                                  controller:
                                      fieldTextController[i + 1],
                                  decoration: InputDecoration(
                                    contentPadding: EdgeInsets.only(
                                      left: 10,
                                      top: 0,
                                      bottom: 0,
                                      right: 10,
                                    ),
                                    enabledBorder: OutlineInputBorder(
                                      borderSide: BorderSide(
                                        color: global
                                            .theme.dividerBorderColor,
                                        width: 0.0,
                                      ),
                                    ),
                                    floatingLabelBehavior:
                                        FloatingLabelBehavior.always,
                                    labelStyle: TextStyle(
                                      color: global
                                          .theme.inputTextBoxColor,
                                    ),
                                    border: const OutlineInputBorder(),
                                    labelText:
                                        "${global.language("cost_center_name")} (${languageList[i].name})",
                                  ),
                                ),
                              ),
                            if (isSaveAllow)
                              SizedBox(
                                width: double.infinity,
                                child: ElevatedButton.icon(
                                  style: ElevatedButton.styleFrom(
                                    backgroundColor:
                                        global.theme.buttonColor,
                                  ),
                                  focusNode:
                                      FocusNode(skipTraversal: true),
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
                ),
              );
            },
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
      costCenterGuidListChecked.clear();
    }
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return MultiBlocListener(
            listeners: [
              BlocListener<CostCenterBloc, CostCenterState>(
                listener: (context, state) {
                  if (state is CostCenterLoadSuccess) {
                    setState(() {
                      listData = state.costCenter;
                    });
                  }
                  if (state is CostCenterGetSuccess) {
                    setState(() {
                      getDataToEditScreen(state.costCenter);
                    });
                  }
                  if (state is CostCenterSaveSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.save,
                            color: global.theme.onPrimaryColor),
                        global.language("save_success"),
                        global.theme.infoHighlightTextColor,
                      );
                      clearEditData();
                      isSaveAllow = false;
                      selectCode = state.responsesID;
                      context.read<CostCenterBloc>().add(
                            const CostCenterLoadList(
                                offset: 0, limit: 999999, search: ""),
                          );
                    });
                  }
                  if (state is CostCenterUpdateSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.save,
                            color: global.theme.onPrimaryColor),
                        global.language("save_success"),
                        global.theme.infoHighlightTextColor,
                      );
                      clearEditData();
                      isSaveAllow = false;
                      context.read<CostCenterBloc>().add(
                            const CostCenterLoadList(
                                offset: 0, limit: 999999, search: ""),
                          );
                    });
                  }
                  if (state is CostCenterDeleteSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.delete,
                            color: global.theme.onPrimaryColor),
                        global.language("delete_success"),
                        global.theme.infoHighlightTextColor,
                      );
                      selectCode = "";
                      clearEditData();
                      isSaveAllow = false;
                      changeScreenEvent(global.ScreenEventEnum.list);
                      context.read<CostCenterBloc>().add(
                            const CostCenterLoadList(
                                offset: 0, limit: 999999, search: ""),
                          );
                    });
                  }
                  if (state is CostCenterDeleteManySuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.delete,
                            color: global.theme.onPrimaryColor),
                        global.language("delete_success"),
                        global.theme.infoHighlightTextColor,
                      );
                      costCenterGuidListChecked.clear();
                      showCheckBox = false;
                      selectCode = "";
                      clearEditData();
                      isSaveAllow = false;
                      changeScreenEvent(global.ScreenEventEnum.list);
                      context.read<CostCenterBloc>().add(
                            const CostCenterLoadList(
                                offset: 0, limit: 999999, search: ""),
                          );
                    });
                  }
                  if (state is CostCenterSaveFailed) {
                    global.showSnackBar(
                      context,
                      Icon(Icons.error,
                          color: global.theme.onPrimaryColor),
                      "${global.language("not_success_save")} ${state.message}",
                      global.theme.negativeHighlightTextColor,
                    );
                  }
                  if (state is CostCenterUpdateFailed) {
                    global.showSnackBar(
                      context,
                      Icon(Icons.error,
                          color: global.theme.onPrimaryColor),
                      "${global.language("not_success_save")} ${state.message}",
                      global.theme.negativeHighlightTextColor,
                    );
                  }
                },
              ),
            ],
            child: (constraints.maxWidth > 800)
                ? SplitView(
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
