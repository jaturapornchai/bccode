import 'dart:async';
import 'dart:io';

import 'package:smlaicloud/bloc/company_branch/company_branch_bloc.dart';
import 'package:smlaicloud/bloc/image/image_upload_bloc.dart';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/contact_model.dart';
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
import 'package:smlaicloud/widgets/point_config_widget.dart';

class PointSettingScreen extends StatefulWidget {
  const PointSettingScreen({super.key});

  @override
  State<PointSettingScreen> createState() => PointSettingScreenState();
}

class PointSettingScreenState extends State<PointSettingScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
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
  List<CompanyBranchModel> listData = [];
  List<String> guidListChecked = [];
  List<LanguageDataModel> names = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  bool isDataChange = false;
  bool isSaveAllow = false;
  late CompanyBranchState blocCurrentState;
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  int _hoverIndex = -1;
  bool isEditMode = false;
  late CompanyBranchModel screenData;

  late DropzoneViewController dropZoneController;
  Color colorSelected = Colors.white;
  final _debouncer = global.Debouncer(1000);
  late Timer screenTimer;
  bool loadingData = false;

  global.ScreenEventEnum screenEvent = global.ScreenEventEnum.list;
  late SplitViewController splitViewController;
  TextEditingController branchCode = TextEditingController();

  final ImagePicker imagePicker = ImagePicker();
  List<File> imageFile = [File('')];
  List<Uint8List> imageWeb = [Uint8List(0)];

  final ImagePicker logoPicker = ImagePicker();
  List<File> logoFile = [File('')];
  List<Uint8List> logoWeb = [Uint8List(0)];

  List<String> selectlanguageList = [];

  bool businesstypeNull = false;

  void setSystemLanguageList() async {
    clearEditData();
    await global.setSystemLanguage(context);

    for (int i = 0; i < global.config.languages.length; i++) {
      if (global.config.languages[i].isuse!) {
        languageList.add(global.config.languages[i]);
      }
    }
    loadDataList("");
  }

  @override
  void initState() {
    selectlanguageList.add("th");

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
    splitViewController = SplitViewController(
      limits: [null, WeightLimit(min: 0.2, max: 0.8)],
    );
    listScrollController.addListener(onScrollList);

    // แก้ไข: ใช้ milliseconds แทน microseconds
    screenTimer = Timer.periodic(const Duration(milliseconds: 500), (timer) {
      if (refreshFocus) {
        fieldFocusNodes[focusNodeIndex].focusNode.requestFocus();
        refreshFocus = false;
      }
    });

    super.initState();
  }

  @override
  void dispose() {
    screenTimer.cancel(); // ✅ cancel timer
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
    context.read<CompanyBranchBloc>().add(
      CompanyBranchLoadList(
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

  String getLangName(String? code) {
    LanguageModel name = languageList.firstWhere(
      (element) => element.code == (code ?? ''),
      orElse: () =>
          LanguageModel(code: '', codeTranslator: '', name: '', isuse: false),
    );
    return name.name!;
  }

  void clearEditData() {
    List<LanguageDataModel> names = [];
    for (int k = 0; k < languageList.length; k++) {
      names.add(LanguageDataModel(code: languageList[k].code!, name: ""));
    }
    branchCode.text = "";
    screenData = CompanyBranchModel(
      code: '',
      guidfixed: '',
      names: names,
      contact: ContactModel(
        address: [],
        phonenumber: '',
        latitude: 0,
        longitude: 0,
      ),
      pointconfig: PointConfigModel(
        generalrules: [],
        specialrules: [],
        pointusagetype: 1,
      ),
    );

    isDataChange = false;
    focusNodeIndex = 0;
    refreshFocus = true;

    imageFile = [File('')];
    imageWeb = [Uint8List(0)];
    logoFile = [File('')];
    logoWeb = [Uint8List(0)];

    businesstypeNull = false;
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
    context.read<CompanyBranchBloc>().add(CompanyBranchGet(guid: guid));
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('point_setting')),
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
        actions: <Widget>[],
      ),
      body: Focus(
        focusNode: FocusNode(skipTraversal: true, canRequestFocus: true),
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
                        _debouncer.run(() {
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
                      global.language("company_branch_code"),
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
                      global.language("company_branch_name"),
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

  void switchToEdit(CompanyBranchModel value) {
    setState(() {
      selectGuid = value.guidfixed;
      getData(selectGuid);
      headerEdit = global.language("edit");
      isSaveAllow = true;
      isEditMode = true;
    });
  }

  Widget listObject(int index, CompanyBranchModel value, bool showCheckBox) {
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

  void upLoadLogo() {
    if (logoFile.isNotEmpty) {
      if (logoFile[0].path != '') {
        context.read<ImageUploadBloc>().add(
          LogoUploadFileSaved(imageFiles: logoFile, imageWeb: logoWeb),
        );
      }
    }
  }

  void upLoadImage() {
    if (imageFile.isNotEmpty) {
      if (imageFile[0].path != '') {
        context.read<ImageUploadBloc>().add(
          ImageUploadFileSaved(imageFiles: imageFile, imageWeb: imageWeb),
        );
      }
    }
  }

  bool verifyData(CompanyBranchModel value) {
    List<String> errorList = [];
    if (screenData.businesstype!.code!.isEmpty) {
      setState(() {
        businesstypeNull = true;
      });
      errorList.add(global.language("please_select_business_type"));
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

  void saveOrUpdateData() {
    if (verifyData(screenData)) {
      showCheckBox = false;
      // print(screenData.toJson());

      if (selectGuid.trim().isEmpty) {
        context.read<CompanyBranchBloc>().add(
          CompanyBranchSave(companyBranch: screenData),
        );
      } else {
        updateData(selectGuid);
      }
    }
  }

  void updateData(String guid) {
    showCheckBox = false;
    context.read<CompanyBranchBloc>().add(
      CompanyBranchUpdate(guid: guid, companyBranch: screenData),
    );
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

  void getDataToEditScreen(CompanyBranchModel companyBranch) {
    isDataChange = false;
    selectGuid = companyBranch.guidfixed;
    branchCode.text = companyBranch.code;
    screenData.code = companyBranch.code;
    screenData.guidfixed = companyBranch.guidfixed;
    screenData.contact?.address = companyBranch.contact?.address;
    screenData.contact?.phonenumber = companyBranch.contact?.phonenumber;
    screenData.contact?.latitude = companyBranch.contact?.latitude;
    screenData.contact?.longitude = companyBranch.contact?.longitude;
    screenData.imageuri = companyBranch.imageuri;
    screenData.logouri = companyBranch.logouri;
    screenData.names = [];
    screenData.pos!.taxid = companyBranch.pos!.taxid;
    screenData.pos!.vatrate = companyBranch.pos!.vatrate;
    screenData.pos!.vattypesale = companyBranch.pos!.vattypesale;
    screenData.pos!.vattypepurchase = companyBranch.pos!.vattypepurchase;
    screenData.pos!.inquirytypesale = companyBranch.pos!.inquirytypesale;
    screenData.pos!.inquirytypepurchase =
        companyBranch.pos!.inquirytypepurchase;
    screenData.pos!.headerreceiptpos = companyBranch.pos!.headerreceiptpos;
    screenData.pos!.footerreceiptpos = companyBranch.pos!.footerreceiptpos;
    screenData.pos!.isbom = companyBranch.pos!.isbom;
    screenData.companynames = [];
    screenData.businesstype = companyBranch.businesstype;
    screenData.pointconfig = companyBranch.pointconfig ?? PointConfigModel();

    // Ensure pointusagetype has default value of 1 if it's 0 or null
    if (screenData.pointconfig!.pointusagetype == 0) {
      screenData.pointconfig!.pointusagetype = 1;
    }

    //เพิ่มภาษาตาม config
    for (var lang in languageList) {
      screenData.names.add(LanguageDataModel(code: lang.code!, name: ""));
      screenData.companynames.add(
        LanguageDataModel(code: lang.code!, name: ""),
      );
    }

    for (var data in companyBranch.names) {
      for (var ele in screenData.names) {
        if (data.code == ele.code) {
          ele.name = data.name;
        }
      }
    }

    for (var data in companyBranch.companynames) {
      for (var ele in screenData.companynames) {
        if (data.code == ele.code) {
          ele.name = data.name;
        }
      }
    }

    //เก็บค่าที่ไม่ได้เปิดใช้งานภาษาเข้าทาง array
    for (var defaultValueLang in companyBranch.names) {
      LanguageDataModel result = screenData.names.firstWhere(
        (data) => data.code == defaultValueLang.code,
        orElse: () => LanguageDataModel(code: '', name: ''),
      );
      if (result.code == '') {
        screenData.names.add(defaultValueLang);
      }
    }

    for (var defaultValueLang in companyBranch.companynames) {
      LanguageDataModel result = screenData.companynames.firstWhere(
        (data) => data.code == defaultValueLang.code,
        orElse: () => LanguageDataModel(code: '', name: ''),
      );
      if (result.code == '') {
        screenData.companynames.add(defaultValueLang);
      }
    }

    screenData.paymentrounding =
        companyBranch.paymentrounding ?? PaymentRoundingModel();
  }

  Widget editScreen({mobileScreen}) {
    List<Widget> formWidgets = [];

    /// ระบบแต้มสะสม (Point System)
    formWidgets.add(
      PointConfigWidget(
        pointConfig: screenData.pointconfig ?? PointConfigModel(),
        onChanged: (updatedPointConfig) {
          setState(() {
            isDataChange = true;
            screenData.pointconfig = updatedPointConfig;
          });
        },
        isEditMode: isEditMode,
      ),
    );

    if (isSaveAllow) {
      formWidgets.add(
        Container(
          width: double.infinity,
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
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
        title: Text('$headerEdit${global.language('point_setting')}'),
        actions: <Widget>[
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
                onPressed: () => saveOrUpdateData(),
                icon: const Icon(Icons.save),
              ),
            ),
        ],
      ),
      body: KeyboardListener(
        focusNode: FocusNode(),
        onKeyEvent: (KeyEvent event) {
          if (event is KeyDownEvent) {
            // print(event.logicalKey);
            if (event.logicalKey == LogicalKeyboardKey.f10) {
              saveOrUpdateData();
            }
            if (event.logicalKey == LogicalKeyboardKey.tab ||
                event.logicalKey == LogicalKeyboardKey.enter) {
              if (HardwareKeyboard.instance.isShiftPressed) {
                //findFocusPrev(focusNodeIndex);
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
            padding: const EdgeInsets.only(top: 10, bottom: 10),
            child: Form(child: Column(children: formWidgets)),
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
              BlocListener<CompanyBranchBloc, CompanyBranchState>(
                listener: (context, state) {
                  blocCurrentState = state;
                  // Load
                  if (state is CompanyBranchLoadSuccess) {
                    setState(() {
                      loadingData = false;
                      if (state.companyBranch.isNotEmpty) {
                        listData.addAll(state.companyBranch);
                      }
                    });
                  }
                  if (state is CompanyBranchLoadFailed) {
                    setState(() {
                      loadingData = false;
                    });
                  }

                  // Update
                  if (state is CompanyBranchUpdateSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.edit, color: Colors.white),
                        global.language("edit_success"),
                        Colors.blue,
                      );
                      isSaveAllow = false;
                      clearEditData();
                      listData.clear();
                      loadDataList(searchText);
                      getData(selectGuid);
                    });
                  }
                  if (state is CompanyBranchUpdateFailed) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.edit, color: Colors.white),
                        "${global.language("not_edit_success")} : ${state.message}",
                        Colors.red,
                      );
                    });
                  }

                  // Get
                  if (state is CompanyBranchGetSuccess) {
                    setState(() {
                      businesstypeNull = false;
                      getDataToEditScreen(state.companyBranch);
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
              BlocListener<ImageUploadBloc, ImageUploadState>(
                listener: (context, state) {
                  if (state is ImageUploadSaveSuccess) {
                    screenData.imageuri = state.imageUpload.uri;
                  } else if (state is LogoUploadSaveSuccess) {
                    screenData.logouri = state.imageUpload.uri;
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

  // Add the missing disableBranch method
  bool disableBranch() {
    // if (screenData.code == '00000') {
    //   return false;
    // } else {
    //   return true;
    // }
    return true;
  }
}
