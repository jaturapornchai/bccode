import 'dart:io';

import 'package:smlaicloud/bloc/devices/devices_bloc.dart';
import 'package:smlaicloud/model/devices_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';

import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';

import 'package:split_view/split_view.dart';
import 'package:translator/translator.dart';

class DevicesScreen extends StatefulWidget {
  const DevicesScreen({super.key});

  @override
  State<DevicesScreen> createState() => DevicesScreenState();
}

class DevicesScreenState extends State<DevicesScreen>
    with SingleTickerProviderStateMixin {
  final translator = GoogleTranslator();
  late TabController tabController;
  ScrollController editScrollController = ScrollController();
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  List<LanguageModel> languageList = <LanguageModel>[];
  List<FocusNode> fieldFocusNodes = [];
  int focusNodeIndex = 0;
  List<File> imageFile = [];
  List<DevicesModel> listDatas = [];
  List<String> guidListChecked = [];
  List<LanguageDataModel> names = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  late SplitViewController splitViewController;
  String searchText = "";
  String selectGuid = "";
  bool isChange = false;
  bool isSaveAllow = false;
  late DevicesState blocDevicesState;
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  bool isEditMode = false;
  late DevicesModel screenData;
  final _debouncer = global.Debouncer(1000);
  void setSystemLanguageList() async {
    await global.setSystemLanguage(context);

    for (int i = 0; i < global.config.languages.length; i++) {
      if (global.config.languages[i].isuse!) {
        languageList.add(global.config.languages[i]);
      }
    }
    clearEditData();
    loadDataList("");
  }

  @override
  void initState() {
    clearEditData();
    setSystemLanguageList();
    splitViewController = SplitViewController(
      limits: [null, WeightLimit(min: 0.2, max: 0.8)],
    );
    tabController = TabController(vsync: this, length: 2);
    tabController.addListener(() {
      setState(() {});
    });

    // เรียงลำดับ Focus
    // Focus รหัส
    for (var i = 0; i < 100; i++) {
      fieldFocusNodes.add(FocusNode());
    }

    listScrollController.addListener(onScrollList);

    super.initState();
  }

  void loadDataList(String search) {
    context.read<DevicesBloc>().add(
      DevicesLoadList(
        offset: (listDatas.isEmpty) ? 0 : listDatas.length,
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

  String getLangName(String? code) {
    LanguageModel name = languageList.firstWhere(
      (element) => element.code == (code ?? ''),
      orElse: () =>
          LanguageModel(code: '', codeTranslator: '', name: '', isuse: false),
    );
    return name.name!;
  }

  @override
  void dispose() {
    listScrollController.dispose();
    tabController.dispose();
    editScrollController.dispose();
    searchController.dispose();
    for (var i = 0; i < 100; i++) {
      fieldFocusNodes[i].dispose();
    }

    super.dispose();
  }

  void clearEditData() {
    List<LanguageDataModel> names = [];
    List<LanguageDataModel> addressNames = [];
    for (int k = 0; k < languageList.length; k++) {
      names.add(LanguageDataModel(code: languageList[k].code!, name: ""));
      addressNames.add(
        LanguageDataModel(code: languageList[k].code!, name: ""),
      );
    }
    List<LanguageDataModel> nameBill = [];
    for (int k = 0; k < languageList.length; k++) {
      nameBill.add(LanguageDataModel(code: languageList[k].code!, name: ""));
    }
    screenData = DevicesModel(guidfixed: "", code: "", name1: "", type: 0);
    isChange = false;
    focusNodeIndex = 0;
  }

  void discardData({required Function callBack}) {
    if (isEditMode && isChange) {
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
    context.read<DevicesBloc>().add(DevicesGet(guid: guid));
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('Devices')),
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
              icon: const Icon(Icons.check_box),
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
                            backgroundColor: Colors.blue,
                          ),
                          onPressed: () {
                            Navigator.pop(context);
                            context.read<DevicesBloc>().add(
                              DevicesDeleteMany(guid: guidListChecked),
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
                      isChange = false;
                      clearEditData();
                      headerEdit = global.language("append");
                      isSaveAllow = true;
                      WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                        tabController.animateTo(1);
                        fieldFocusNodes[0].requestFocus();
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
      body: Focus(
        focusNode: FocusNode(skipTraversal: true, canRequestFocus: true),
        onKey: (node, event) {
          if (kIsWeb) {
            if (event is RawKeyDownEvent) {
              if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
                isKeyDown = false;
                int index = listDatas.indexOf(
                  listDatas.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index > 0) {
                  selectGuid = listDatas[index - 1].guidfixed;
                  currentListIndex = index + 1;
                  isKeyUp = true;
                  getData(selectGuid);
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                isKeyUp = false;
                int index = listDatas.indexOf(
                  listDatas.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                selectGuid = listDatas[index + 1].guidfixed;
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
                  padding: EdgeInsets.only(left: 10, right: 10),
                  child: TextFormField(
                    onFieldSubmitted: (value) {
                      searchFocusNode.requestFocus();
                    },
                    onChanged: (value) {
                      _debouncer.run(() {
                        setState(() {
                          listDatas = [];
                        });
                        loadDataList(value);
                      });
                    },
                    autofocus: false,
                    focusNode: searchFocusNode,
                    controller: searchController,
                    decoration: InputDecoration(
                      isDense: true,
                      contentPadding: const EdgeInsets.only(
                        top: 10,
                        bottom: 10,
                      ),
                      border: InputBorder.none,
                      hintText: (kIsWeb)
                          ? "${global.language('search')} (F2)"
                          : global.language('search'),
                    ),
                  ),
                ),
              ),
            ),
            Container(color: global.theme.appBarColor, height: 2),
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
                      global.language("Devices_code"),
                      style: const TextStyle(
                        color: Colors.black,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 10,
                    child: Text(
                      global.language("Devices_name"),
                      style: const TextStyle(
                        color: Colors.black,
                        fontWeight: FontWeight.bold,
                      ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  if (showCheckBox)
                    const Expanded(
                      flex: 1,
                      child: Icon(Icons.check, color: Colors.black, size: 12),
                    ),
                ],
              ),
            ),
            Expanded(
              child: SingleChildScrollView(
                controller: listScrollController,
                child: Column(
                  children: listDatas
                      .asMap()
                      .entries
                      .map(
                        (entry) =>
                            listObject(entry.key, entry.value, showCheckBox),
                      )
                      .toList(),
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }

  void switchToEdit(DevicesModel value) {
    setState(() {
      selectGuid = value.guidfixed;
      getData(selectGuid);
      headerEdit = global.language("edit");
      isSaveAllow = true;
      isEditMode = true;
    });
  }

  Widget listObject(int index, DevicesModel value, bool showCheckBox) {
    bool isCheck = false;
    for (int i = 0; i < guidListChecked.length; i++) {
      if (guidListChecked[i] == value.guidfixed) {
        isCheck = true;
        break;
      }
    }
    // ลบการ add key ออก - keys ถูกสร้างใน LoadSuccess แล้ว
    // listKeys.add(GlobalKey());  // ❌ ตรงนี้ทำให้เกิด infinite loop!
    return GestureDetector(
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
                selectGuid = value.guidfixed;
                getData(selectGuid);
                searchFocusNode.requestFocus();
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
                value.name1,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            if (showCheckBox)
              Expanded(
                flex: 1,
                child: (isCheck)
                    ? const Icon(Icons.check, size: 12)
                    : Container(),
              ),
          ],
        ),
      ),
    );
  }

  void saveOrUpdateData() {
    showCheckBox = false;

    if (selectGuid.trim().isEmpty) {
      if (screenData.code != "" && screenData.name1 != "") {
        context.read<DevicesBloc>().add(DevicesSave(devices: screenData));
      } else {
        global.showSnackBar(
          context,
          Icon(Icons.edit, color: Colors.white),
          global.language("empty_data"),
          Colors.red,
        );
      }
    } else {
      updateData(selectGuid);
    }
  }

  void updateData(String guid) {
    showCheckBox = false;

    if (screenData.code != "" && screenData.name1 != "") {
      context.read<DevicesBloc>().add(
        DevicesUpdate(guid: guid, devices: screenData),
      );
    } else {
      global.showSnackBar(
        context,
        Icon(Icons.edit, color: Colors.white),
        global.language("empty_data"),
        Colors.red,
      );
    }
  }

  void getDataToEditScreen(DevicesModel devices) {
    isChange = false;
    selectGuid = devices.guidfixed;
    screenData.code = devices.code;
    screenData.name1 = devices.name1;
    screenData.guidfixed = devices.guidfixed;
    screenData.type = devices.type;
  }

  void findFocusNext(int index) {
    focusNodeIndex = index + 1;
    if (focusNodeIndex > fieldFocusNodes.length - 1) {
      focusNodeIndex = 0;
    }
    fieldFocusNodes[focusNodeIndex].requestFocus();
  }

  Widget editScreen({mobileScreen}) {
    int nodeIndex = 0;
    List<Widget> formWidgets = [];
    formWidgets.add(
      Padding(
        padding: const EdgeInsets.only(bottom: 10),
        child: TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {
            findFocusNext(0);
          },
          textInputAction: TextInputAction.next,
          focusNode: fieldFocusNodes[nodeIndex],
          textAlign: TextAlign.left,
          controller: TextEditingController(text: screenData.code),
          textCapitalization: TextCapitalization.characters,
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.allow(RegExp('[a-z A-Z 0-9 -]')),
          ],
          onChanged: (value) {
            isChange = true;
            screenData.code = value;
          },
          decoration: InputDecoration(
            border: const OutlineInputBorder(),
            contentPadding: const EdgeInsets.only(
              left: 10,
              top: 0,
              bottom: 0,
              right: 10,
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            labelText: global.language("Devices_code"),
          ),
        ),
      ),
    );
    nodeIndex++;
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(bottom: 10),
        child: TextFormField(
          readOnly: !isEditMode,
          onChanged: (value) {
            isChange = true;
            screenData.name1 = value;
          },
          onFieldSubmitted: (value) {
            findFocusNext(focusNodeIndex);
          },
          textInputAction: TextInputAction.next,
          focusNode: fieldFocusNodes[nodeIndex],
          textAlign: TextAlign.left,
          controller: TextEditingController(text: screenData.name1),
          decoration: InputDecoration(
            border: const OutlineInputBorder(),
            contentPadding: const EdgeInsets.only(
              left: 10,
              top: 0,
              bottom: 0,
              right: 10,
            ),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            labelText: global.language("Devices_name"),
          ),
        ),
      ),
    );
    nodeIndex++;
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(bottom: 10),
        child: Column(
          children: [
            Row(
              mainAxisAlignment: MainAxisAlignment.start,
              children: [
                Radio(
                  value: 0,
                  groupValue: screenData.type,
                  onChanged: (value) {
                    setState(() {
                      screenData.type = value!;
                    });
                  },
                ),
                Text(global.language("cashier")),
              ],
            ),
            Row(
              mainAxisAlignment: MainAxisAlignment.start,
              children: [
                Radio(
                  value: 1,
                  groupValue: screenData.type,
                  onChanged: (value) {
                    setState(() {
                      screenData.type = value!;
                    });
                  },
                ),
                Text(global.language("customer_display")),
              ],
            ),
            Row(
              mainAxisAlignment: MainAxisAlignment.start,
              children: [
                Radio(
                  value: 2,
                  groupValue: screenData.type,
                  onChanged: (value) {
                    setState(() {
                      screenData.type = value!;
                    });
                  },
                ),
                Text(global.language("chef_que_display")),
              ],
            ),
            Row(
              mainAxisAlignment: MainAxisAlignment.start,
              children: [
                Radio(
                  value: 3,
                  groupValue: screenData.type,
                  onChanged: (value) {
                    setState(() {
                      screenData.type = value!;
                    });
                  },
                ),
                Text(global.language("kitchen_display")),
              ],
            ),
          ],
        ),
      ),
    );
    if (isSaveAllow) {
      formWidgets.add(
        SizedBox(
          width: double.infinity,
          child: ElevatedButton.icon(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () {
              saveOrUpdateData();
            },
            icon: Icon(Icons.save),
            label: Text(global.language("save") + ((kIsWeb) ? " (F10)" : "")),
          ),
        ),
      );
    }

    return Scaffold(
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
        title: Text(headerEdit + global.language("Devices")),
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
                            backgroundColor: Colors.blue,
                          ),
                          onPressed: () {
                            Navigator.pop(context);
                            context.read<DevicesBloc>().add(
                              DevicesDelete(guid: selectGuid),
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
          if (isEditMode && global.systemLanguage.length > 1)
            Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () async {
                  setState(() {});
                },
                icon: const Icon(Icons.translate),
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
                    listDatas[listDatas.indexOf(
                      listDatas.firstWhere(
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
                onPressed: () => saveOrUpdateData(),
                icon: const Icon(Icons.save),
              ),
            ),
        ],
      ),
      body: Focus(
        focusNode: FocusNode(skipTraversal: true),
        onKey: (node, event) {
          if (kIsWeb) {
            if (event is RawKeyDownEvent) {
              if (event.logicalKey == LogicalKeyboardKey.f2) {
                searchFocusNode.requestFocus();
              }
              if (event.logicalKey == LogicalKeyboardKey.f10) {
                saveOrUpdateData();
              }
            }
          }
          return KeyEventResult.ignored;
        },
        child: SingleChildScrollView(
          controller: editScrollController,
          child: Container(
            width: double.infinity,
            padding: const EdgeInsets.all(10),
            child: Column(children: formWidgets),
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
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return BlocListener<DevicesBloc, DevicesState>(
            listener: (context, state) {
              blocDevicesState = state;
              // Load
              if (state is DevicesLoadSuccess) {
                setState(() {
                  if (state.devices.isNotEmpty) {
                    listDatas.addAll(state.devices);
                  }
                });
              }
              // Save
              if (state is DevicesSaveSuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.save, color: Colors.white),
                    global.language("save_success"),
                    Colors.blue,
                  );
                  clearEditData();
                  listDatas.clear();
                  loadDataList(searchText);
                });
              }
              if (state is DevicesSaveFailed) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.save, color: Colors.white),
                    "${global.language("not_success_save")} : ${state.message}",
                    Colors.red,
                  );
                });
              }
              // Update
              if (state is DevicesUpdateSuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.edit, color: Colors.white),
                    global.language("edit_success"),
                    Colors.blue,
                  );
                  clearEditData();
                  listDatas.clear();
                  WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                    tabController.animateTo(0);
                  });
                  loadDataList(searchText);
                  isSaveAllow = false;
                  getData(selectGuid);
                });
              }
              if (state is DevicesUpdateFailed) {
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
              if (state is DevicesDeleteSuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.delete, color: Colors.white),
                    global.language("delete_success"),
                    Colors.blue,
                  );
                  listDatas.clear();
                  clearEditData();
                  WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                    tabController.animateTo(0);
                  });
                  loadDataList(searchText);
                });
              }
              // Delete Many
              if (state is DevicesDeleteManySuccess) {
                setState(() {
                  global.showSnackBar(
                    context,
                    Icon(Icons.delete, color: Colors.white),
                    global.language("not_delete_success"),
                    Colors.blue,
                  );
                  listDatas.clear();
                  clearEditData();
                  loadDataList(searchText);
                  showCheckBox = false;
                });
              }
              // Get
              if (state is DevicesGetSuccess) {
                setState(() {
                  getDataToEditScreen(state.devices);
                  if (isEditMode) {
                    WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                      tabController.animateTo(1);
                    });
                    setState(() {
                      findFocusNext(0);
                    });
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
            },
            child: (constraints.maxWidth > 800)
                ? SplitView(
                    controller: splitViewController,
                    gripSize: 8,
                    gripColor: global.theme.appBarColor,
                    gripColorActive: Colors.blue,
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
