import 'dart:async';
import 'package:smlaicloud/widgets/manual_button.dart';
import 'dart:io';

import 'package:smlaicloud/bloc/employee/employee_bloc.dart';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/employee_model.dart';
import 'package:smlaicloud/screen_search/company_branch_search_screen.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_dropzone/flutter_dropzone.dart';
import 'package:smlaicloud/widgets/list_font_size_control.dart';
import 'package:image_picker/image_picker.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:loading_animation_widget/loading_animation_widget.dart';
import 'package:split_view/split_view.dart';
import 'package:translator/translator.dart';

class EmployeeScreen extends StatefulWidget {
  const EmployeeScreen({super.key});

  @override
  State<EmployeeScreen> createState() => EmployeeScreenState();
}

class EmployeeScreenState extends State<EmployeeScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  final translator = GoogleTranslator();
  late TabController tabController;
  ScrollController editScrollController = ScrollController();
  bool refreshFocus = false;
  TextEditingController searchController = TextEditingController();
  TextEditingController groupController = TextEditingController();
  int focusNodeMax = 0;
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  List<LanguageModel> languageList = <LanguageModel>[];
  List<global.FieldFocusModel> fieldFocusNodes = [];
  int focusNodeIndex = 0;
  List<EmployeeModel> listData = [];
  List<String> guidListChecked = [];
  List<LanguageDataModel> names = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  bool isDataChange = false;
  bool isSaveAllow = false;
  late EmployeeState blocCurrentState;
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  int _hoverIndex = -1;
  bool isEditMode = false;
  late EmployeeModel screenData;
  File imageFile = File('');
  Uint8List? imageWeb;
  final ImagePicker imagePicker = ImagePicker();
  late DropzoneViewController dropZoneController;
  late SplitViewController splitViewController;
  final debouncer = global.Debouncer(1000);
  late Timer screenTimer;
  bool loadingData = false;
  TextEditingController employeeCode = TextEditingController();
  bool _isLoadingSave = false;

  final _formKey = GlobalKey<FormState>();

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
    splitViewController = SplitViewController(
      limits: [null, WeightLimit(min: 0.2, max: 0.8)],
    );
    clearEditData();
    setSystemLanguageList();
    tabController = TabController(vsync: this, length: 2);
    tabController.addListener(() {
      setState(() {});
    });

    // เรียงลำดับ Focus
    for (int i = 0; i < 100; i++) {
      fieldFocusNodes.add(global.FieldFocusModel(focusNode: FocusNode()));
      fieldFocusNodes[i].focusNode.addListener(() {
        if (fieldFocusNodes[i].focusNode.hasFocus) {
          focusNodeIndex = i;
          fieldFocusNodes[focusNodeIndex].focusNode.requestFocus();
        }
      });
    }
    listScrollController.addListener(onScrollList);

    super.initState();
  }

  @override
  void dispose() {
    listScrollController.dispose();
    tabController.dispose();
    editScrollController.dispose();
    searchController.dispose();
    groupController.dispose();
    for (int i = 0; i < fieldFocusNodes.length; i++) {
      fieldFocusNodes[i].focusNode.dispose();
    }
    super.dispose();
  }

  void loadDataList(String search) {
    setState(() {
      loadingData = true;
    });
    searchText = search;
    context.read<EmployeeBloc>().add(
      EmployeeLoadList(
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
    List<LanguageDataModel> names = [];
    List<LanguageDataModel> itemunitnames = [];
    for (int k = 0; k < languageList.length; k++) {
      names.add(LanguageDataModel(code: languageList[k].code!, name: ""));
      itemunitnames.add(
        LanguageDataModel(code: languageList[k].code!, name: ""),
      );
    }
    employeeCode.text = "";
    screenData = EmployeeModel(
      guidfixed: "",
      code: "",
      name: "",
      email: "",
      profilepicture: "",
      pincode: "12345",
      isenabled: true,
      branches: [],
    );

    isDataChange = false;
    focusNodeIndex = 0;
    refreshFocus = true;
    _isLoadingSave = false;
    setState(() {
      imageFile = File('');
      imageWeb = null;
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
    imageWeb = null;
    imageFile = File('');
    context.read<EmployeeBloc>().add(EmployeeGet(guid: guid));
  }

  void switchToEdit(EmployeeModel value) {
    setState(() {
      selectGuid = value.guidfixed;
      getData(selectGuid);
      headerEdit = global.language("edit");
      isSaveAllow = true;
      isEditMode = true;
    });
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('employee')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () {
            discardData(
              callBack: () {
                // Navigator.pop(context);
                global.gotoMainMenu(context);
                isEditMode = false;
              },
            );
          },
        ),
        actions: <Widget>[
          const ManualButton(path: 'settings-employee'),
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
                            context.read<EmployeeBloc>().add(
                              EmployeeDeleteMany(guid: guidListChecked),
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

                      screenData.branches!.add(
                        CompanyBranchModel(
                          guidfixed: global.companyBranchSelectData.guidfixed,
                          code: global.companyBranchSelectData.code,
                          names: global.companyBranchSelectData.names,
                        ),
                      );
                      headerEdit = global.language("append");
                      isSaveAllow = true;
                      if (mobileScreen) {
                        WidgetsBinding.instance.addPostFrameCallback((
                          timeStamp,
                        ) {
                          tabController.animateTo(1);
                        });
                      }
                      fieldFocusNodes[0].focusNode.requestFocus();
                    });
                  },
                );
              },
              icon: const Icon(Icons.add),
            ),
          ),

          /// เพิ่มมูลใหม่จากข้อมูลเดิม (Copy)
          // if (selectGuid.isNotEmpty)
          //   Padding(
          // padding: EdgeInsets.only(right: 20.0),
          //       child: IconButton(
          //         focusNode: FocusNode(skipTraversal: true),
          //         onPressed: () {
          //           discardData(callBack: () {
          //             setState(() {
          //               isEditMode = true;
          //               showCheckBox = false;
          //               isChange = false;
          //               isChange = false;
          //               headerEdit = global.language("append");
          //               isSaveAllow = true;
          //               if (mobileScreen) {
          //                 WidgetsBinding.instance
          //                     .addPostFrameCallback((timeStamp) {
          //                   tabController.animateTo(1);
          //                 });
          //               }
          //               fieldFocusNodes[0].requestFocus();
          //             });
          //           });
          //         },
          //         icon: const Icon(
          //           Icons.copy_all,
          //         ),
          //       )),
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
                  selectGuid = listData[index - 1].guidfixed;
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
                selectGuid = listData[index + 1].guidfixed;
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
                      onSubmitted: (value) {
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
              key: headerKey,
              padding: EdgeInsets.only(
                left: 10,
                right: 10,
                top: 5,
                bottom: 5,
              ),
              color: global.theme.columnHeaderColor,
              child: Row(
                children: [
                  Expanded(
                    flex: 5,
                    child: Text(
                      global.language("employee_code"),
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
                      global.language("employee_name"),
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
                    flex: 5,
                    child: Text(
                      global.language("employee_phone"),
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
                      global.language("reset_pin_code"),
                      style: TextStyle(
                        color: global.theme.columnHeaderTextColor,
                        fontWeight: FontWeight.bold,
                        fontSize: global.deviceConfig.listDataFontSize + 2,
                      ),
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

  Widget listObject(int index, EmployeeModel value, bool showCheckBox) {
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
                //searchFocusNode.requestFocus();
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
        color: _getContainerColor(value.guidfixed, index),
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
              flex: 5,
              child: Text(
                value.name,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 5,
              child: Text(
                value.contact?.phonenumber as String,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 5,
              child: Padding(
                padding: EdgeInsets.all(5.0),
                child: ElevatedButton(
                  onPressed: () {
                    showDialog<String>(
                      context: context,
                      builder: (BuildContext context) => AlertDialog(
                        title: Text(global.language('confirm_reset_pin_code')),
                        content: Text(
                          global.language('confirm_reset_pin_code_detail'),
                        ),
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
                              value.pincode = "12345";
                              context.read<EmployeeBloc>().add(
                                EmployeeUpdate(
                                  guid: value.guidfixed,
                                  employeeModel: value,
                                ),
                              );

                              Navigator.pop(context);
                            },
                            child: Text(global.language('confirm')),
                          ),
                        ],
                      ),
                    );
                  },
                  child: Row(
                    children: [
                      Icon(
                        Icons.reset_tv,
                        size: global.deviceConfig.listDataFontSize,
                      ),
                      SizedBox(width: 5),
                      Text(global.language('reset')),
                    ],
                  ),
                ),
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

  Future<void> saveOrUpdateData() async {
    // print(jsonEncode(screenData.toJson()));
    showCheckBox = false;
    if (selectGuid.trim().isEmpty) {
      if (imageFile.path.isNotEmpty) {
        context.read<EmployeeBloc>().add(
          EmployeeWithImageSave(
            employee: screenData,
            imageFile: imageFile,
            imageWeb: imageWeb,
          ),
        );
      } else {
        context.read<EmployeeBloc>().add(
          EmployeeSave(employeeModel: screenData),
        );
      }
    } else {
      updateData(selectGuid);
    }
  }

  void updateData(String guid) {
    showCheckBox = false;
    if (imageWeb != null) {
      context.read<EmployeeBloc>().add(
        EmployeeWithImageUpdate(
          guid: guid,
          employee: screenData,
          imageFile: imageFile,
          imageWeb: imageWeb!,
        ),
      );
    } else {
      context.read<EmployeeBloc>().add(
        EmployeeUpdate(guid: guid, employeeModel: screenData),
      );
    }
  }

  void findFocusNext(int index) {
    focusNodeIndex = index;
    do {
      focusNodeIndex++;
      if (focusNodeIndex > focusNodeMax) {
        focusNodeIndex = 0;
      }
    } while (fieldFocusNodes[focusNodeIndex].isReadOnly);
    // print("findFocusNext=$focusNodeIndex");
    refreshFocus = true;
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

  Widget editScreen({mobileScreen}) {
    List<Widget> formWidgets = [];

    focusNodeMax = 0;
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          enabled: screenData.code.isEmpty,
          textAlign: TextAlign.left,
          controller: employeeCode,
          textCapitalization: TextCapitalization.characters,
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.allow(RegExp('[a-z A-Z 0-9 -]')),
          ],
          onChanged: (value) {
            isDataChange = true;
            screenData.code = value.toUpperCase();
            employeeCode.value = TextEditingValue(
              text: value.toUpperCase(),
              selection: employeeCode.selection,
            );
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("employee_code"),
          ),
          validator: (value) {
            if (value == null || value.isEmpty) {
              return global.language('this_field_is_required');
            }
            return null;
          },
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          readOnly: !isEditMode,
          onChanged: (value) {
            isDataChange = true;
            screenData.name = value;
          },
          focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: TextEditingController(text: screenData.name),
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("employee_name"),
          ),
          validator: (value) {
            if (value == null || value.isEmpty) {
              return global.language('this_field_is_required');
            }
            return null;
          },
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextField(
          readOnly: !isEditMode,
          onChanged: (value) {
            isDataChange = true;
            screenData.email = value;
          },
          onSubmitted: (value) {
            if (kIsWeb) {
              findFocusNext(focusNodeIndex);
            }
          },
          focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: TextEditingController(text: screenData.email),
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("employee_email"),
          ),
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextField(
          readOnly: !isEditMode,
          onChanged: (value) {
            isDataChange = true;
            screenData.contact?.address = value;
          },
          onSubmitted: (value) {
            if (kIsWeb) {
              findFocusNext(focusNodeIndex);
            }
          },
          focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: TextEditingController(text: screenData.contact?.address),
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("employee_address"),
          ),
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextField(
          readOnly: !isEditMode,
          onChanged: (value) {
            isDataChange = true;
            screenData.contact?.phonenumber = value;
          },
          onSubmitted: (value) {
            if (kIsWeb) {
              findFocusNext(focusNodeIndex);
            }
          },
          focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: TextEditingController(
            text: screenData.contact?.phonenumber,
          ),
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.allow(RegExp('[0-9]')),
          ],
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("employee_phone"),
          ),
        ),
      ),
    );

    formWidgets.add(
      Container(
        width: double.infinity,
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Row(
          children: [
            Radio(
              focusNode: FocusNode(skipTraversal: true),
              value: true,
              groupValue: screenData.isenabled,
              onChanged: (value) {
                setState(() {
                  screenData.isenabled = true;
                });
              },
            ),
            Text(global.language("open")),
            const SizedBox(width: 10),
            Radio(
              focusNode: FocusNode(skipTraversal: true),
              value: false,
              groupValue: screenData.isenabled,
              onChanged: (value) {
                setState(() {
                  screenData.isenabled = false;
                });
              },
            ),
            Text(global.language("close")),
          ],
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Row(
          children: [
            Checkbox(
              value: screenData.isusepos,
              onChanged: (bool? value) {
                setState(() {
                  screenData.isusepos = value ?? false;
                });
              },
            ),
            Text(global.language("use_pos")),
          ],
        ),
      ),
    );

    for (int i = 0; i < screenData.branches!.length; i++) {
      formWidgets.add(
        Padding(
          padding: const EdgeInsets.all(8.0),
          child: Container(
            padding: const EdgeInsets.only(bottom: 5),
            width: double.infinity,
            child: Column(
              children: [
                Row(
                  children: [
                    Expanded(
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
                                      screenData.branches![i].guidfixed =
                                          result.guid;
                                      screenData.branches![i].code =
                                          result.code;
                                      screenData.branches![i].names =
                                          result.names;
                                    }
                                  });
                                });
                              }
                            : null,
                        child: Row(
                          children: [
                            Icon(
                              (screenData.branches![i].guidfixed.isEmpty)
                                  ? Icons.search
                                  : Icons.business,
                            ),
                            SizedBox(width: 10),
                            Text(
                              (screenData.branches![i].guidfixed.isEmpty)
                                  ? global.language("serch_branch")
                                  : "${screenData.branches![i].code} ~  ${global.packName(screenData.branches![i].names)}",
                            ),
                          ],
                        ),
                      ),
                    ),
                    const SizedBox(width: 10),
                    IconButton(
                      onPressed: (isEditMode)
                          ? () {
                              setState(() {
                                screenData.branches!.removeAt(i);
                              });
                            }
                          : null,
                      icon: const Icon(Icons.delete),
                    ),
                  ],
                ),
              ],
            ),
          ),
        ),
      );
    }

    /// add branch list to employee
    formWidgets.add(
      Padding(
        padding: EdgeInsets.all(8.0),
        child: SizedBox(
          width: double.infinity,
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
                      screenData.branches?.add(
                        CompanyBranchModel(guidfixed: "", code: "", names: []),
                      );
                    });
                  }
                : null,
            icon: Icon(Icons.business),
            label: Text(global.language("add_branch")),
          ),
        ),
      ),
    );

    // formWidgets.add(Padding(
    // padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
    //     child: TextField(
    //       obscureText: true,
    //       enabled: false,
    //       onChanged: (value) {
    //         isDataChange = true;
    //         screenData.pincode = value;
    //       },
    //       onSubmitted: (value) {
    //         if (kIsWeb) {
    //           findFocusNext(focusNodeIndex);
    //         }
    //       },
    //       keyboardType: TextInputType.number,
    //       inputFormatters: [
    //         FilteringTextInputFormatter.digitsOnly,
    //         LengthLimitingTextInputFormatter(4),
    //       ],
    //       focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
    //       textAlign: TextAlign.left,
    //       controller: TextEditingController(text: screenData.pincode),
    //       decoration: InputDecoration(
    //         floatingLabelBehavior: FloatingLabelBehavior.always,
    // border: OutlineInputBorder(),
    //         labelText: global.language("pin_code"),
    //       ),
    //     )));

    if (isEditMode) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.all(8.0),
          child: Row(
            children: [
              Expanded(
                child: ElevatedButton.icon(
                  focusNode: FocusNode(skipTraversal: true),
                  onPressed: () async {
                    FocusScope.of(context).unfocus();
                    setState(() {
                      imageWeb = null;
                      imageFile = File('');
                      screenData.profilepicture = '';
                    });
                  },
                  icon: Icon(Icons.delete),
                  label: Text(global.language('delete_picture')),
                ),
              ),
              const SizedBox(width: 5),
              Expanded(
                child: ElevatedButton.icon(
                  focusNode: FocusNode(skipTraversal: true),
                  onPressed: (kIsWeb)
                      ? () async {
                          FocusScope.of(context).unfocus();
                          XFile? image = await imagePicker.pickImage(
                            source: ImageSource.gallery,
                            maxHeight: 480,
                            maxWidth: 640,
                            imageQuality: 60,
                          );
                          if (image != null) {
                            var f = await image.readAsBytes();
                            setState(() {
                              imageWeb = f;
                              imageFile = File(image.path);
                            });
                          }
                        }
                      : () {
                          FocusScope.of(context).unfocus();
                          setState(() async {
                            final XFile? photo = await imagePicker.pickImage(
                              source: ImageSource.gallery,
                              maxHeight: 480,
                              maxWidth: 640,
                              imageQuality: 60,
                            );
                            if (photo != null) {
                              var f = await photo.readAsBytes();
                              setState(() {
                                imageWeb = f;
                                imageFile = File(photo.path);
                              });
                            }
                          });
                        },
                  icon: Icon(Icons.folder),
                  label: Text(global.language("select_picture")),
                ),
              ),
              const SizedBox(width: 5),
              if (kIsWeb == false)
                Expanded(
                  child: ElevatedButton.icon(
                    focusNode: FocusNode(skipTraversal: true),
                    onPressed: () async {
                      FocusScope.of(context).unfocus();
                      final XFile? photo = await imagePicker.pickImage(
                        source: ImageSource.camera,
                        maxHeight: 480,
                        maxWidth: 640,
                        imageQuality: 60,
                      );
                      if (photo != null) {
                        var f = await photo.readAsBytes();
                        setState(() {
                          imageWeb = f;
                          imageFile = File(photo.path);
                        });
                      }
                    },
                    icon: Icon(Icons.camera_alt),
                    label: Text(global.language('take_photo')),
                  ),
                ),
            ],
          ),
        ),
      );
    }

    formWidgets.add(
      Container(
        padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
        width: double.infinity,
        height: 300,
        child: Stack(
          children: [
            if (kIsWeb)
              DropzoneView(
                operation: DragOperation.copy,
                cursor: CursorType.grab,
                onCreated: (ctrl) => dropZoneController = ctrl,
                onLoaded: () {},
                onError: (ev) {},
                onHover: () {},
                onLeave: () {},
                onDrop: (ev) async {
                  final bytes = await dropZoneController.getFileData(ev);
                  setState(() {
                    imageWeb = bytes;
                  });
                },
                onDropMultiple: (ev) async {},
              ),
            Center(
              child: DecoratedBox(
                decoration: BoxDecoration(
                  color: Colors.white,
                  border: Border.all(color: global.theme.textColor),
                  borderRadius: BorderRadius.circular(5),
                  image: (imageWeb != null)
                      ? DecorationImage(
                          image: MemoryImage(imageWeb!),
                          fit: BoxFit.fill,
                        )
                      : (screenData.profilepicture != '')
                      ? DecorationImage(
                          image: NetworkImage(screenData.profilepicture!),
                        )
                      : const DecorationImage(
                          image: AssetImage('assets/img/noimage.png'),
                        ),
                ),
                child: const SizedBox(width: double.infinity, height: 400),
              ),
            ),
          ],
        ),
      ),
    );

    if (isSaveAllow) {
      formWidgets.add(
        Container(
          width: double.infinity,
          padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: ElevatedButton.icon(
            onPressed:
                _isLoadingSave // Disable the button when loading
                ? null
                : () async {
                    if (_formKey.currentState!.validate()) {
                      await saveOrUpdateData(); // Perform the operation
                      // No need to set _isLoading to false here if it's already set in saveOrUpdateData
                    }
                  },
            icon: _isLoadingSave
                ? Container(
                    width: 24,
                    height: 24,
                    margin: const EdgeInsets.only(
                      right: 8,
                    ), // Add some spacing between the loader and the label text
                    child: const CircularProgressIndicator(
                      strokeWidth: 3,
                      valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                    ),
                  )
                : Icon(Icons.save),
            label: Text(
              global.language("save") + ((kIsWeb) ? " (F10)" : ""),
              // Update the text to show "saving" or similar feedback when loading
            ),
          ),
        ),
      );
    }


    return Scaffold(
      backgroundColor: isEditMode ? global.theme.toolBarEditModeColor.withValues(alpha: 0.05) : global.theme.backgroundColor,
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
        title: Text(headerEdit + global.language("employee")),
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
                            context.read<EmployeeBloc>().add(
                              EmployeeDelete(guid: selectGuid),
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
                onPressed:
                    _isLoadingSave // Disable the button when loading
                    ? null
                    : () async {
                        if (_formKey.currentState!.validate()) {
                          await saveOrUpdateData(); // Perform the operation
                          // No need to set _isLoading to false here if it's already set in saveOrUpdateData
                        }
                      },
                icon: _isLoadingSave
                    ? Container(
                        width: 24,
                        height: 24,
                        padding: const EdgeInsets.all(2.0),
                        child: const CircularProgressIndicator(
                          strokeWidth: 3,
                          valueColor: AlwaysStoppedAnimation<Color>(
                            Colors.white,
                          ),
                        ),
                      )
                    : const Icon(Icons.save),
                // Optionally, adjust icon size here if needed
              ),
            ),
        ],
      ),
      body: RawKeyboardListener(
              focusNode: FocusNode(),
              onKey: (RawKeyEvent event) {
                if (event is RawKeyDownEvent) {
                  if (event.logicalKey == LogicalKeyboardKey.f10) {
                    if (_formKey.currentState!.validate()) {
                      saveOrUpdateData();
                    }
                  }
                  if (event.logicalKey == LogicalKeyboardKey.tab ||
                      event.logicalKey == LogicalKeyboardKey.enter) {
                    if (event.isShiftPressed) {
                    } else {
                      findFocusNext(focusNodeIndex);
                    }
                  }
                }
              },
              child: SingleChildScrollView(
                    controller: editScrollController,
                    child: Container(
                      width: double.infinity,
                      padding: const EdgeInsets.only(top: 10, bottom: 15),
                      child: Form(
                        key: _formKey,
                        child: Column(children: formWidgets),
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
              BlocListener<EmployeeBloc, EmployeeState>(
                listener: (context, state) {
                  blocCurrentState = state;
                  // Load
                  if (state is EmployeeLoadSuccess) {
                    setState(() {
                      loadingData = false;
                      if (state.employees.isNotEmpty) {
                        listData.addAll(state.employees);
                      }
                    });
                  }
                  if (state is EmployeeLoadFailed) {
                    setState(() {
                      loadingData = false;
                    });
                  }
                  // Save
                  if (state is EmployeeSaveInProgress) {
                    setState(() {
                      _isLoadingSave = true;
                    });
                  } else if (state is EmployeeSaveSuccess) {
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
                  } else if (state is EmployeeSaveFailed) {
                    _isLoadingSave = false;
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
                  if (state is EmployeeUpdateSuccess) {
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
                      getData(selectGuid);
                    });
                  }
                  if (state is EmployeeUpdateFailed) {
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
                  if (state is EmployeeDeleteSuccess) {
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
                  if (state is EmployeeDeleteManySuccess) {
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
                  if (state is EmployeeGetSuccess) {
                    setState(() {
                      isDataChange = false;

                      screenData = state.employee;
                      employeeCode.text = screenData.code;

                      if (isEditMode) {
                        WidgetsBinding.instance.addPostFrameCallback((
                          timeStamp,
                        ) {
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
