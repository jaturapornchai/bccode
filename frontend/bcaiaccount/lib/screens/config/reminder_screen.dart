import 'package:smlaicloud/bloc/line_notify/line_notify_bloc.dart';
import 'package:smlaicloud/model/line_notify_model.dart';
import 'package:smlaicloud/screen_search/company_branch_search_screen.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/widgets/list_font_size_control.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:loading_animation_widget/loading_animation_widget.dart';
import 'package:split_view/split_view.dart';

class ReminderScreen extends StatefulWidget {
  const ReminderScreen({super.key});

  @override
  State<ReminderScreen> createState() => ReminderScreenState();
}

class ReminderScreenState extends State<ReminderScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  late TabController tabController;
  ScrollController editScrollController = ScrollController();
  TextEditingController searchController = TextEditingController();
  List<LanguageModel> languageList = <LanguageModel>[];
  List<LineNotifyModel> listData = [];
  List<String> guidListChecked = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  bool isDataChange = false;
  bool isSaveAllow = false;
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  int _hoverIndex = -1;
  bool isEditMode = false;
  late LineNotifyModel screenData;
  final _debouncer = global.Debouncer(1000);
  bool loadingData = false;
  late SplitViewController splitViewController;
  final _formKey = GlobalKey<FormState>();

  TextEditingController lineNotifyNameController = TextEditingController();
  TextEditingController lineNotifyTokenController = TextEditingController();

  @override
  void initState() {
    clearEditData();
    loadDataList("");
    tabController = TabController(vsync: this, length: 2);
    tabController.addListener(() {
      setState(() {});
    });

    splitViewController = SplitViewController(
      limits: [null, WeightLimit(min: 0.2, max: 0.8)],
    );
    listScrollController.addListener(onScrollList);

    super.initState();
  }

  @override
  void dispose() {
    listScrollController.dispose();
    tabController.dispose();
    editScrollController.dispose();
    searchController.dispose();
    lineNotifyNameController.dispose();
    lineNotifyTokenController.dispose();

    super.dispose();
  }

  void loadDataList(String search) {
    setState(() {
      loadingData = true;
    });
    searchText = search;
    context.read<LineNotifyBloc>().add(
      LineNotifyLoadList(
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

  void clearEditData() {
    setState(() {
      screenData = LineNotifyModel(
        guidfixed: "",
        branchevents: [
          BranchEvent(
            branch: Branch(
              code: global.companyBranchSelectData.code,
              guidfixed: global.companyBranchSelectData.guidfixed,
              names: global.companyBranchSelectData.names,
            ),
            isenable: true,
            isnearoutofstock: true,
            isoutofstock: true,
            issavebill: true,
            ispreorder: true,
          ),
        ],
        name: "",
        token: "",
        type: "",
      );
      isDataChange = false;

      lineNotifyNameController.text = "";
      lineNotifyTokenController.text = "";
    });
  }

  void discardData({required Function callBack}) {
    if (isEditMode && isDataChange) {
      showDialog<String>(
        context: context,
        builder: (BuildContext context) => AlertDialog(
          title: Text(global.language('data_editing')),
          content: Text(global.language('leave_this_screen')),
          actions: <Widget>[
            ElevatedButton(
              style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
              onPressed: () => Navigator.pop(context),
              child: Text(global.language('no')),
            ),
            ElevatedButton(
              style: ElevatedButton.styleFrom(backgroundColor: Colors.blue),
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
    isEditMode = false;
    context.read<LineNotifyBloc>().add(LineNotifyGet(guid: guid));
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('line_notify')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () {
            discardData(
              callBack: () {
                Navigator.pop(context);
                isEditMode = false;
              },
            );
          },
        ),
        actions: <Widget>[
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
                        guidListChecked.clear();
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
          if (guidListChecked.isNotEmpty)
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
                            backgroundColor: Colors.red,
                          ),
                          onPressed: () => Navigator.pop(context),
                          child: Text(global.language('no')),
                        ),
                        ElevatedButton(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: global.theme.primaryColor,
                          ),
                          onPressed: () {
                            Navigator.pop(context);
                            context.read<LineNotifyBloc>().add(
                              LineNotifyDeleteMany(guid: guidListChecked),
                            );
                          },
                          child: Text(global.language('delete')),
                        ),
                      ],
                    ),
                  );
                  setState(() {});
                },
                icon: const Icon(Icons.delete),
              ),
            ),

          /// เพิ่มข้อมูลใหม่
          Padding(
            padding: EdgeInsets.only(right: 20.0),
            child: IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () {
                discardData(
                  callBack: () {
                    setState(() {
                      isEditMode = true;
                      selectGuid = "";
                      showCheckBox = false;
                      isDataChange = false;
                      clearEditData();
                      headerEdit = global.language("append");
                      isSaveAllow = true;
                      if (mobileScreen) {
                        WidgetsBinding.instance.addPostFrameCallback((
                          timeStamp,
                        ) {
                          tabController.animateTo(1);
                        });
                      }
                    });
                  },
                );
              },
              icon: const Icon(Icons.add),
            ),
          ),
        ],
      ),
      body: Focus(
        focusNode: FocusNode(skipTraversal: true, canRequestFocus: true),
        onKey: (node, event) {
          if (kIsWeb) {
            if (event is RawKeyDownEvent) {
              if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
                isKeyDown = false;
                int index = listData.indexOf(
                  listData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index > 0) {
                  selectGuid = listData[index - 1].guidfixed!;
                  currentListIndex = index + 1;
                  isKeyUp = true;
                  getData(selectGuid);
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                isKeyUp = false;
                int index = listData.indexOf(
                  listData.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                selectGuid = listData[index + 1].guidfixed!;
                currentListIndex = index + 1;
                isKeyDown = true;
                getData(selectGuid);
              }
            }
          }
          return KeyEventResult.ignored;
        },
        child: Column(
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
                    child: TextField(
                      onSubmitted: (value) {},
                      onChanged: (value) {
                        _debouncer.run(() {
                          setState(() {
                            listData = [];
                          });
                          loadDataList(value);
                        });
                      },
                      autofocus: false,
                      controller: searchController,
                      decoration: InputDecoration(
                        isDense: true,
                        contentPadding: EdgeInsets.only(
                          top: 0,
                          bottom: 0,
                          left: 0,
                          right: 0,
                        ),
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
                    flex: 5,
                    child: Text(
                      global.language("name"),
                      style: TextStyle(
                        color: global.theme.columnHeaderTextColor,
                        fontWeight: FontWeight.bold,
                        fontSize: global.deviceConfig.listDataFontSize + 2,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("token"),
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
            if (loadingData)
              Center(
                child: LoadingAnimationWidget.staggeredDotsWave(
                  color: global.theme.primaryColor,
                  size: 50,
                ),
              ),
          ],
        ),
      ),
    );
  }

  void switchToEdit(LineNotifyModel value) {
    setState(() {
      selectGuid = value.guidfixed!;
      getData(selectGuid);
      headerEdit = global.language("edit");
      isSaveAllow = true;
      isEditMode = true;
    });
  }

  Widget listObject(int index, LineNotifyModel value, bool showCheckBox) {
    bool isCheck = false;
    for (int i = 0; i < guidListChecked.length; i++) {
      if (guidListChecked[i] == value.guidfixed) {
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
            selectGuid = value.guidfixed!;
            if (isCheck == true) {
              guidListChecked.remove(value.guidfixed);
            } else {
              guidListChecked.add(value.guidfixed!);
            }
            global.showSnackBar(
              context,
              Icon(Icons.check, color: Colors.white),
              "${global.language("chosen")} ${guidListChecked.length} ${global.language("list")}",
              Colors.blue,
            );
          });
        } else {
          setState(() {
            discardData(
              callBack: () {
                isSaveAllow = false;
                isEditMode = false;
                selectGuid = value.guidfixed!;
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
              flex: 1,
              child: Text(
                value.name!,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 2,
              child: Text(
                value.token!,
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

  void saveOrUpdateData() {
    showCheckBox = false;

    if (selectGuid.trim().isEmpty) {
      context.read<LineNotifyBloc>().add(
        LineNotifySave(lineNotify: screenData),
      );
    } else {
      updateData(selectGuid);
    }
  }

  void updateData(String guid) {
    showCheckBox = false;

    context.read<LineNotifyBloc>().add(
      LineNotifyUpdate(guid: guid, lineNotify: screenData),
    );
  }

  void getDataToEditScreen(LineNotifyModel lineNotify) {
    isDataChange = false;
    selectGuid = lineNotify.guidfixed!;
    screenData = lineNotify;
    lineNotifyNameController.text = lineNotify.name!;
    lineNotifyTokenController.text = lineNotify.token!;
  }

  Widget editScreen({mobileScreen}) {
    List<Widget> formWidgets = [];

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          onEditingComplete: () {},
          textAlign: TextAlign.left,
          controller: lineNotifyNameController,
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isDataChange = true;
            screenData.name = value;
            lineNotifyNameController.value = TextEditingValue(
              text: value.toUpperCase(),
              selection: lineNotifyNameController.selection,
            );
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("name"),
          ),
          validator: (value) {
            if (value == null || value.isEmpty) {
              return 'This field is required';
            }
            return null;
          },
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Row(
          children: [
            Expanded(
              child: TextFormField(
                onEditingComplete: () {},
                textAlign: TextAlign.left,
                controller: lineNotifyTokenController,
                textCapitalization: TextCapitalization.characters,
                onChanged: (value) {
                  isDataChange = true;
                  screenData.token = value;
                  lineNotifyTokenController.value = TextEditingValue(
                    text: value,
                    selection: lineNotifyTokenController.selection,
                  );
                },
                decoration: InputDecoration(
                  floatingLabelBehavior: FloatingLabelBehavior.always,
                  border: OutlineInputBorder(),
                  labelText: global.language("token"),
                ),
                validator: (value) {
                  if (value == null || value.isEmpty) {
                    return 'This field is required';
                  }
                  return null;
                },
              ),
            ),
            SizedBox(width: 10),
            IconButton(
              onPressed: () {
                context.read<LineNotifyBloc>().add(
                  LineNotifyTest(
                    token: lineNotifyTokenController.text,
                    message: global.language("test_send_message"),
                  ),
                );
              },
              icon: const Icon(Icons.send_to_mobile_rounded),
            ),
          ],
        ),
      ),
    );

    for (int i = 0; i < screenData.branchevents!.length; i++) {
      formWidgets.add(
        Padding(
          padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: Container(
            padding: const EdgeInsets.only(bottom: 5),
            width: double.infinity,
            child: Column(
              children: [
                Row(
                  children: [
                    Expanded(
                      child: SizedBox(
                        height: 50,
                        child: ElevatedButton(
                          style: ButtonStyle(
                            backgroundColor: WidgetStateProperty.all<Color>(
                              const Color.fromARGB(255, 238, 86, 144),
                            ),
                          ),
                          onPressed: (isEditMode)
                              ? () {
                                  Navigator.push(
                                    context,
                                    MaterialPageRoute(
                                      builder: (context) =>
                                          const CompanyBranchSearchScreen(
                                            word: "",
                                          ),
                                    ),
                                  ).then((value) {
                                    setState(() {
                                      SearchGuidCodeNameModel result = value;
                                      if (result.guid.isNotEmpty) {
                                        screenData
                                                .branchevents![i]
                                                .branch!
                                                .guidfixed =
                                            result.guid;
                                        screenData
                                                .branchevents![i]
                                                .branch!
                                                .code =
                                            result.code;
                                        screenData
                                                .branchevents![i]
                                                .branch!
                                                .names =
                                            result.names;
                                      }
                                    });
                                  });
                                }
                              : null,
                          child: Row(
                            children: [
                              Icon(
                                (screenData
                                        .branchevents![i]
                                        .branch!
                                        .guidfixed!
                                        .isEmpty)
                                    ? Icons.search
                                    : Icons.business,
                              ),
                              SizedBox(width: 10),
                              Text(
                                (screenData
                                        .branchevents![i]
                                        .branch!
                                        .guidfixed!
                                        .isEmpty)
                                    ? global.language("serch_branch")
                                    : "${screenData.branchevents![i].branch!.code} ~  ${global.packName(screenData.branchevents![i].branch!.names!)}",
                              ),
                            ],
                          ),
                        ),
                      ),
                    ),
                    const SizedBox(width: 10),
                    IconButton(
                      color: Colors.red,
                      onPressed: (isEditMode)
                          ? () {
                              setState(() {
                                screenData.branchevents!.removeAt(i);
                              });
                            }
                          : null,
                      icon: const Icon(Icons.delete),
                    ),
                  ],
                ),
                Padding(
                  padding: EdgeInsets.all(8.0),
                  child: SwitchListTile(
                    title: Text("1. ${global.language("enable_notify")}"),
                    value: screenData.branchevents![i].isenable!,
                    onChanged: (value) {
                      setState(() {
                        screenData.branchevents![i].isenable = value;
                      });
                    },
                  ),
                ),
                Padding(
                  padding: EdgeInsets.all(8.0),
                  child: SwitchListTile(
                    title: Text("2. ${global.language("out_of_stock")}"),
                    value: screenData.branchevents![i].isoutofstock!,
                    onChanged: (value) {
                      setState(() {
                        screenData.branchevents![i].isoutofstock = value;
                      });
                    },
                  ),
                ),
                Padding(
                  padding: EdgeInsets.all(8.0),
                  child: SwitchListTile(
                    title: Text("3. ${global.language("near_out_of_stock")}"),
                    value: screenData.branchevents![i].isnearoutofstock!,
                    onChanged: (value) {
                      setState(() {
                        screenData.branchevents![i].isnearoutofstock = value;
                      });
                    },
                  ),
                ),
                Padding(
                  padding: EdgeInsets.all(8.0),
                  child: SwitchListTile(
                    title: Text("4. ${global.language("save_bill")}"),
                    value: screenData.branchevents![i].issavebill!,
                    onChanged: (value) {
                      setState(() {
                        screenData.branchevents![i].issavebill = value;
                      });
                    },
                  ),
                ),
                Padding(
                  padding: EdgeInsets.all(8.0),
                  child: SwitchListTile(
                    title: Text("5. ${global.language("is_preorder")}"),
                    value: screenData.branchevents![i].ispreorder!,
                    onChanged: (value) {
                      setState(() {
                        screenData.branchevents![i].ispreorder = value;
                      });
                    },
                  ),
                ),
              ],
            ),
          ),
        ),
      );
    }

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: SizedBox(
          width: double.infinity,
          height: 50,
          child: ElevatedButton.icon(
            style: ButtonStyle(
              backgroundColor: WidgetStateProperty.all<Color>(
                const Color.fromARGB(255, 148, 160, 194),
              ),
              foregroundColor: WidgetStateProperty.all<Color>(
                const Color.fromARGB(255, 0, 0, 0),
              ),
            ),
            focusNode: FocusNode(skipTraversal: true),
            onPressed: (isEditMode)
                ? () {
                    setState(() {
                      screenData.branchevents?.add(BranchEvent());
                    });
                  }
                : null,
            icon: Icon(Icons.business),
            label: Text(global.language("add_branch")),
          ),
        ),
      ),
    );

    if (isSaveAllow) {
      formWidgets.add(
        Container(
          width: double.infinity,
          height: 60,
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: ElevatedButton.icon(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () {
              if (_formKey.currentState!.validate()) {
                saveOrUpdateData();
              }
            },
            icon: Icon(Icons.save),
            label: Text(global.language("save") + ((kIsWeb) ? " (F10)" : "")),
          ),
        ),
      );
    }

    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: (isEditMode)
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
                      isEditMode = false;
                      WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                        tabController.animateTo(0);
                      });
                    },
                  );
                },
              )
            : null,
        title: Text(headerEdit + global.language("line_notify")),
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
                            backgroundColor: Colors.red,
                          ),
                          onPressed: () => Navigator.pop(context),
                          child: Text(global.language('no')),
                        ),
                        ElevatedButton(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: global.theme.primaryColor,
                          ),
                          onPressed: () {
                            Navigator.pop(context);
                            context.read<LineNotifyBloc>().add(
                              LineNotifyDelete(guid: selectGuid),
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
          if (isSaveAllow == true)
            Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () {
                  if (_formKey.currentState!.validate()) {
                    saveOrUpdateData();
                  }
                },
                icon: const Icon(Icons.save),
              ),
            ),
        ],
      ),
      body: RawKeyboardListener(
        focusNode: FocusNode(),
        onKey: (RawKeyEvent event) {
          if (event is RawKeyDownEvent) {
            // print(event.logicalKey);
            if (event.logicalKey == LogicalKeyboardKey.f10) {
              if (_formKey.currentState!.validate()) {
                saveOrUpdateData();
              }
            }
            if (event.logicalKey == LogicalKeyboardKey.tab ||
                event.logicalKey == LogicalKeyboardKey.enter) {
              if (event.isShiftPressed) {
                //findFocusPrev(focusNodeIndex);
              } else {}
            }
          }
        },
        child: SingleChildScrollView(
          controller: editScrollController,
          child: Container(
            width: double.infinity,
            padding: const EdgeInsets.only(top: 10, bottom: 15),
            child: Form(
              child: Form(
                key: _formKey,
                child: Column(children: formWidgets),
              ),
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
      guidListChecked.clear();
    }
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return MultiBlocListener(
            listeners: [
              BlocListener<LineNotifyBloc, LineNotifyState>(
                listener: (context, state) {
                  // Load
                  if (state is LineNotifyLoadSuccess) {
                    setState(() {
                      loadingData = false;
                      if (state.lineNotifys.isNotEmpty) {
                        listData.addAll(state.lineNotifys);
                      }
                    });
                  }
                  if (state is LineNotifyLoadFailed) {
                    setState(() {
                      loadingData = false;
                    });
                  }
                  // Save
                  if (state is LineNotifySaveSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.save, color: Colors.white),
                        global.language("save_success"),
                        Colors.blue,
                      );
                      clearEditData();
                      listData.clear();
                      loadDataList(searchText);
                    });
                  }
                  if (state is LineNotifySaveFailed) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.save, color: Colors.white),
                        "//${global.language("not_success_save")} : ${state.message}",
                        Colors.red,
                      );
                    });
                  }
                  // Update
                  if (state is LineNotifyUpdateSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.edit, color: Colors.white),
                        global.language("edit_success"),
                        Colors.blue,
                      );
                      clearEditData();
                      listData.clear();
                      WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                        tabController.animateTo(0);
                      });
                      loadDataList(searchText);
                      isSaveAllow = false;
                      // getData(selectGuid);
                    });
                  }
                  if (state is LineNotifyUpdateFailed) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.edit, color: Colors.white),
                        "${global.language("not_edit_success")} : ${state.message}",
                        Colors.red,
                      );
                    });
                  }
                  // Delete
                  if (state is LineNotifyDeleteSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.delete, color: Colors.white),
                        global.language("delete_success"),
                        Colors.blue,
                      );
                      listData.clear();
                      clearEditData();
                      WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                        tabController.animateTo(0);
                      });
                      loadDataList(searchText);
                    });
                  }
                  // Delete Many
                  if (state is LineNotifyDeleteManySuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.delete, color: Colors.white),
                        global.language("not_delete_success"),
                        Colors.blue,
                      );
                      listData.clear();
                      clearEditData();
                      loadDataList(searchText);
                      showCheckBox = false;
                    });
                  }
                  // Get
                  if (state is LineNotifyGetSuccess) {
                    setState(() {
                      getDataToEditScreen(state.lineNotify);
                      if (isEditMode) {
                        WidgetsBinding.instance.addPostFrameCallback((
                          timeStamp,
                        ) {
                          tabController.animateTo(1);
                        });
                        setState(() {});
                      }
                    });
                    if (currentListIndex >= 0) {
                      RenderBox? boxHeader =
                          headerKey.currentContext?.findRenderObject()
                              as RenderBox?;
                      Offset? positionheader = boxHeader?.localToGlobal(
                        Offset.zero,
                      );
                      RenderBox? box =
                          listKeys[currentListIndex].currentContext
                                  ?.findRenderObject()
                              as RenderBox?;
                      Offset? position = box?.localToGlobal(Offset.zero);
                      if (position != null &&
                          positionheader != null &&
                          boxHeader != null &&
                          box != null) {
                        // Scroll Up
                        if (isKeyUp &&
                            position.dy <=
                                (positionheader.dy +
                                    (boxHeader.size.height +
                                        (box.size.height * 2)))) {
                          setState(() {
                            listScrollController.animateTo(
                              listScrollController.offset -
                                  (boxHeader.size.height + box.size.height),
                              duration: const Duration(milliseconds: 100),
                              curve: Curves.ease,
                            );
                            isKeyUp = false;
                          });
                        }
                        // Scroll Down
                        if (isKeyDown &&
                            position.dy > (queryData.size.height - 100)) {
                          setState(() {
                            listScrollController.animateTo(
                              listScrollController.offset +
                                  (position.dy - (queryData.size.height - 100)),
                              duration: const Duration(milliseconds: 100),
                              curve: Curves.easeOut,
                            );
                            isKeyDown = false;
                          });
                        }
                      }
                    }
                  }
                  if (state is LineNotifyTestSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        const Icon(
                          Icons.send_to_mobile_rounded,
                          color: Colors.white,
                        ),
                        global.language("token_is_valid"),
                        Colors.blue,
                      );
                    });
                  }

                  if (state is LineNotifyTestFailed) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        const Icon(
                          Icons.send_to_mobile_rounded,
                          color: Colors.white,
                        ),
                        global.language("token_is_invalid"),
                        Colors.red,
                      );
                    });
                  }
                },
              ),
            ],
            child: (constraints.maxWidth > 800)
                ? SplitView(
                    controller: splitViewController,
                    gripSize: 8,
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
