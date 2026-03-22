import 'dart:async';

import 'package:smlaicloud/bloc/shelf_excel_import/shelf_excel_import_bloc.dart';
import 'package:smlaicloud/widgets/manual_button.dart';
import 'package:smlaicloud/bloc/warehouse/warehose_bloc.dart';
import 'package:smlaicloud/bloc/warehouse_location/warehouse_location_bloc.dart';
import 'package:smlaicloud/model/location_model.dart';
import 'package:smlaicloud/model/shelf_model.dart';
import 'package:smlaicloud/model/warehouse_location_model.dart';
import 'package:smlaicloud/model/warehouse_location_update_model.dart';
import 'package:smlaicloud/model/warehouse_model.dart';
import 'package:smlaicloud/screens/config/shelf_product_management_screen.dart';
import 'package:smlaicloud/repositories/product_barcode_repository.dart';
import 'package:smlaicloud/screen_search/product_warehouse_search_screen.dart';
import 'package:file_picker/file_picker.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_dropzone/flutter_dropzone.dart';
import 'package:smlaicloud/widgets/list_font_size_control.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:loading_animation_widget/loading_animation_widget.dart';
import 'package:split_view/split_view.dart';
import 'package:translator/translator.dart';
import 'package:smlaicloud/utils/focus_utils.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class ProductLocaltionScreen extends StatefulWidget {
  const ProductLocaltionScreen({super.key});

  @override
  State<ProductLocaltionScreen> createState() => ProductLocaltionScreenState();
}

class ProductLocaltionScreenState extends State<ProductLocaltionScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  final translator = GoogleTranslator();
  late TabController tabController;
  ScrollController editScrollController = ScrollController();
  TextEditingController searchController = TextEditingController();
  TextEditingController warehouseCodeController = TextEditingController();
  List<LanguageModel> warehouseNameController = <LanguageModel>[];
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  List<LanguageModel> languageList = <LanguageModel>[];
  List<WarehouseLocationModel> listData = [];
  List<String> guidListChecked = [];
  List<LanguageDataModel> names = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  bool isDataChange = false;
  bool isSaveAllow = false;
  late WarehouseLocationState blocWarehouseLocationState;
  late WarehouseState blocWarehouseState;
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  int _hoverIndex = -1;
  bool isEditMode = false;
  bool isAddMode = false;
  late WarehouseModel screenData;
  late LocationModel locationData;
  late WarehouseModel warehouseData;

  late DropzoneViewController dropZoneController;
  Color colorSelected = Colors.white;
  final _debouncer = global.Debouncer(1000);
  late Timer screenTimer;

  // Loading states for shelf rendering
  bool isShelfLoading = false;
  String shelfFilter = "";
  bool loadingData = false;
  String selectWarehouseCode = "";
  String selectLocationCode = "";
  late SplitViewController splitViewController;
  global.ScreenEventEnum screenEvent = global.ScreenEventEnum.list;

  bool isLoadTranslation = false;

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
    tabController = TabController(vsync: this, length: 2);
    tabController.addListener(() {
      setState(() {});
    });

    splitViewController = SplitViewController(
      limits: [null, WeightLimit(min: 0.2, max: 0.8)],
    );
    listScrollController.addListener(onScrollList);
    setSystemLanguageList();

    super.initState();
  }

  @override
  void dispose() {
    listScrollController.dispose();
    tabController.dispose();
    editScrollController.dispose();
    searchController.dispose();
    warehouseCodeController.dispose();
    super.dispose();
  }

  void changeScreenEvent(global.ScreenEventEnum event) {
    // print(event);
    screenEvent = event;
    setState(() {});
  }

  void loadDataList(String search) {
    setState(() {
      loadingData = true;
    });
    searchText = search;
    context.read<WarehouseLocationBloc>().add(
      WarehouseLoadLocationList(
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

    List<LanguageDataModel> locationnames = [];
    for (int k = 0; k < languageList.length; k++) {
      locationnames.add(
        LanguageDataModel(code: languageList[k].code!, name: ""),
      );
    }

    warehouseCodeController.text = "";
    screenData = WarehouseModel(
      guidfixed: "",
      code: "",
      names: names,
      location: [],
    );

    locationData = LocationModel(code: "", names: names, shelf: []);

    warehouseData = WarehouseModel(
      guidfixed: "",
      code: "",
      names: names,
      location: [],
    );

    isDataChange = false;
    isDataChange = false;
  }

  void discardData({required Function callBack}) {
    // กรณีมีการแก้ไขข้อมูล ให้เตือน
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

  void getData(String warehousecode, String locationcode) {
    headerEdit = global.language("show");
    if (isAddMode) {
      changeScreenEvent(global.ScreenEventEnum.add);
    } else {
      changeScreenEvent(global.ScreenEventEnum.list);
    }
    context.read<WarehouseLocationBloc>().add(
      WarehouseLocationGetByCode(
        warehousecode: warehousecode,
        locationcode: locationcode,
      ),
    );
  }

  void getWarehouseData(String guid) {
    context.read<WarehouseBloc>().add(WarehouseGet(guid: guid));
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('product_location')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () {
            discardData(
              callBack: () {
                global.gotoMainMenu(context);
                isEditMode = false;
                isAddMode = false;
                changeScreenEvent(global.ScreenEventEnum.list);
              },
            );
          },
        ),
        actions: <Widget>[
          const ManualButton(path: 'warehouse'),
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
                          child: Text(global.language('product_location_no')),
                        ),
                        ElevatedButton(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: global.theme.primaryColor,
                          ),
                          onPressed: () {
                            Navigator.pop(context);
                            context.read<WarehouseLocationBloc>().add(
                              WarehouseLocationDeleteMany(
                                warehousecode: selectWarehouseCode,
                                locationcode: guidListChecked,
                              ),
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
                      selectWarehouseCode = "";
                      selectLocationCode = "";
                      showCheckBox = false;
                      isDataChange = false;
                      isAddMode = true;
                      clearEditData();
                      changeScreenEvent(global.ScreenEventEnum.add);
                      headerEdit = global.language("append");
                      isSaveAllow = true;
                      if (mobileScreen) {
                        WidgetsBinding.instance.addPostFrameCallback((
                          timeStamp,
                        ) {
                          tabController.animateTo(1);
                        });
                      }
                      WidgetsBinding.instance.addPostFrameCallback((_) {
                        focusFirstTextField(context);
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
                int index = listData.indexOf(
                  listData.firstWhere(
                    (element) => element.locationcode == selectGuid,
                  ),
                );
                if (index > 0) {
                  selectGuid = listData[index - 1].locationcode;
                  currentListIndex = index + 1;
                  isKeyUp = true;
                  getData(selectWarehouseCode, selectLocationCode);
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                isKeyUp = false;
                int index = listData.indexOf(
                  listData.firstWhere(
                    (element) => element.locationcode == selectGuid,
                  ),
                );
                selectGuid = listData[index + 1].locationcode;
                currentListIndex = index + 1;
                isKeyDown = true;
                getData(selectWarehouseCode, selectLocationCode);
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
                        contentPadding: const EdgeInsets.only(
                          top: 0,
                          bottom: 0,
                          left: 0,
                          right: 0,
                        ),
                        border: InputBorder.none,
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
                      global.language("warehouse_code"),
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
                      global.language("warehouse_name"),
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
                      global.language("location_code"),
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
                      global.language("location_name"),
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

  void switchToEdit(WarehouseLocationModel value) {
    setState(() {
      selectGuid = value.guidfixed;
      selectWarehouseCode = value.warehousecode;
      selectLocationCode = value.locationcode;
      getData(selectWarehouseCode, selectLocationCode);
      headerEdit = global.language("edit");
      changeScreenEvent(global.ScreenEventEnum.edit);
      isSaveAllow = true;
      isEditMode = true;
      isAddMode = false;
    });
    WidgetsBinding.instance.addPostFrameCallback((_) {
      focusFirstTextField(context);
    });
  }

  Widget listObject(
    int index,
    WarehouseLocationModel value,
    bool showCheckBox,
  ) {
    bool isCheck = false;

    for (int i = 0; i < guidListChecked.length; i++) {
      if (guidListChecked[i] == value.locationcode) {
        isCheck = true;
        break;
      }
    }
    // ลบการ add key ออก - keys ถูกสร้างใน LoadSuccess แล้ว
    // listKeys.add(GlobalKey());  // ❌ ตรงนี้ทำให้เกิด infinite loop!
    bool selected =
        selectLocationCode == value.locationcode &&
        selectWarehouseCode == value.warehousecode;
    TextStyle textStyle = TextStyle(
      fontWeight: (selected) ? FontWeight.bold : FontWeight.normal,
      fontSize: (selected)
          ? global.deviceConfig.listDataFontSize + 2.0
          : global.deviceConfig.listDataFontSize,
      color: (selected) ? global.theme.textColor : global.theme.textSecondaryColor,
    );
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      onEnter: (_) => setState(() => _hoverIndex = index),
      onExit: (_) => setState(() => _hoverIndex = -1),
      child: GestureDetector(
      onTap: () {
        selectGuid = value.guidfixed;
        if (showCheckBox == true) {
          setState(() {
            selectWarehouseCode = value.warehousecode;
            selectLocationCode = value.locationcode;
            if (isCheck == true) {
              guidListChecked.remove(value.locationcode);
            } else {
              guidListChecked.add(value.locationcode);
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
                isAddMode = false;
                selectGuid = value.guidfixed;

                selectWarehouseCode = value.warehousecode;
                selectLocationCode = value.locationcode;
                changeScreenEvent(global.ScreenEventEnum.list);
                getData(selectWarehouseCode, selectLocationCode);
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
        color:
            (selectWarehouseCode == value.warehousecode &&
                selectLocationCode == value.locationcode)
            ? global.theme.rowSelectedColor
            : (index % 2 == 0)
            ? global.theme.columnAlternateEvenColor
            : global.theme.columnAlternateOddColor,
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
                value.warehousecode,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 10,
              child: Text(
                global.packName(value.warehousenames),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 5,
              child: Text(
                value.locationcode,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 10,
              child: Text(
                global.packName(value.locationnames),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
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
    ),
    );
  }

  void saveOrUpdateData() {
    showCheckBox = false;

    if (isAddMode) {
      saveData();
    } else {
      updateData(selectWarehouseCode, selectLocationCode);
    }
  }

  void saveData() {
    // ตรวจสอบ shelf code ซ้ำใน location นี้
    List<String> shelfCodes = [];
    List<String> duplicateShelfCodes = [];

    for (var shelf in locationData.shelf) {
      String code = shelf.code.trim().toUpperCase();
      if (code.isNotEmpty) {
        if (shelfCodes.contains(code)) {
          if (!duplicateShelfCodes.contains(code)) {
            duplicateShelfCodes.add(code);
          }
        } else {
          shelfCodes.add(code);
        }
      }
    }

    // หากพบ shelf code ซ้ำ แสดงข้อความแจ้งเตือน
    if (duplicateShelfCodes.isNotEmpty) {
      showDialog(
        context: context,
        barrierDismissible: false,
        builder: (BuildContext dialogContext) {
          return AlertDialog(
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12),
            ),
            title: Row(
              children: [
                Icon(Icons.warning, color: Colors.orange, size: 28),
                SizedBox(width: 8),
                Expanded(
                  child: Text(
                    global.language('product_location_duplicate_shelf_code'),
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                      color: Colors.orange.shade700,
                    ),
                  ),
                ),
              ],
            ),
            content: SizedBox(
              width: 400,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: Colors.orange.shade50,
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(color: Colors.orange.shade200),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Icon(
                              Icons.info_outline,
                              color: Colors.orange.shade600,
                              size: 20,
                            ),
                            const SizedBox(width: 8),
                            Expanded(
                              child: Text(
                                '⚠️ พบรหัสชั้นวางที่ซ้ำกัน:',
                                style: TextStyle(
                                  fontWeight: FontWeight.bold,
                                  fontSize: 14,
                                  color: Colors.orange.shade800,
                                ),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        Wrap(
                          spacing: 6,
                          runSpacing: 4,
                          children: duplicateShelfCodes.map((code) {
                            return Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 8,
                                vertical: 4,
                              ),
                              decoration: BoxDecoration(
                                color: Colors.red.shade100,
                                border: Border.all(color: Colors.red.shade300),
                                borderRadius: BorderRadius.circular(6),
                              ),
                              child: Text(
                                code,
                                style: TextStyle(
                                  fontSize: 13,
                                  fontWeight: FontWeight.bold,
                                  color: Colors.red.shade800,
                                ),
                              ),
                            );
                          }).toList(),
                        ),
                        SizedBox(height: 8),
                        Text(
                          global.language('product_location_fix_duplicate_before_save'),
                          style: TextStyle(
                            fontSize: 13,
                            color: global.theme.textSecondaryColor,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            actions: [
              ElevatedButton(
                onPressed: () => Navigator.of(dialogContext).pop(),
                style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.orange,
                  foregroundColor: Colors.white,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                  padding: const EdgeInsets.symmetric(
                    horizontal: 24,
                    vertical: 12,
                  ),
                ),
                child:  Text(
                  global.language('product_location_ok'),
                  style: TextStyle(fontWeight: FontWeight.bold),
                ),
              ),
            ],
          );
        },
      );
      return;
    }

    // // print(jsonEncode(warehouseData.toJson()));
    // // print(jsonEncode(locationData.toJson()));

    warehouseData.location.add(
      LocationModel(
        code: locationData.code,
        names: locationData.names,
        shelf: locationData.shelf,
      ),
    );

    // Filter out locations with empty codes
    List<LocationModel> filteredLocations = warehouseData.location
        .where((location) => location.code.trim().isNotEmpty)
        .toList();

    WarehouseModel dataSave = WarehouseModel(
      guidfixed: screenData.guidfixed,
      code: screenData.code,
      names: screenData.names,
      location: filteredLocations,
    );

    // // print(jsonEncode(dataSave.toJson()));

    context.read<WarehouseBloc>().add(
      WarehouseUpdate(guid: screenData.guidfixed, warehouseModel: dataSave),
    );
  }

  void updateData(String selectWarehouseCode, String selectLocationCode) {
    // ตรวจสอบ shelf code ซ้ำใน location นี้
    List<String> shelfCodes = [];
    List<String> duplicateShelfCodes = [];

    for (var shelf in locationData.shelf) {
      String code = shelf.code.trim().toUpperCase();
      if (code.isNotEmpty) {
        if (shelfCodes.contains(code)) {
          if (!duplicateShelfCodes.contains(code)) {
            duplicateShelfCodes.add(code);
          }
        } else {
          shelfCodes.add(code);
        }
      }
    }

    // หากพบ shelf code ซ้ำ แสดงข้อความแจ้งเตือน
    if (duplicateShelfCodes.isNotEmpty) {
      showDialog(
        context: context,
        barrierDismissible: false,
        builder: (BuildContext dialogContext) {
          return AlertDialog(
            shape: RoundedRectangleBorder(
              borderRadius: BorderRadius.circular(12),
            ),
            title: Row(
              children: [
                Icon(Icons.warning, color: Colors.orange, size: 28),
                SizedBox(width: 8),
                Expanded(
                  child: Text(
                    global.language('product_location_duplicate_shelf_code'),
                    style: TextStyle(
                      fontSize: 18,
                      fontWeight: FontWeight.bold,
                      color: Colors.orange.shade700,
                    ),
                  ),
                ),
              ],
            ),
            content: SizedBox(
              width: 400,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: Colors.orange.shade50,
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(color: Colors.orange.shade200),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Row(
                          children: [
                            Icon(
                              Icons.info_outline,
                              color: Colors.orange.shade600,
                              size: 20,
                            ),
                            const SizedBox(width: 8),
                            Expanded(
                              child: Text(
                                '⚠️ พบรหัสชั้นวางที่ซ้ำกัน:',
                                style: TextStyle(
                                  fontWeight: FontWeight.bold,
                                  fontSize: 14,
                                  color: Colors.orange.shade800,
                                ),
                              ),
                            ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        Wrap(
                          spacing: 6,
                          runSpacing: 4,
                          children: duplicateShelfCodes.map((code) {
                            return Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 8,
                                vertical: 4,
                              ),
                              decoration: BoxDecoration(
                                color: Colors.red.shade100,
                                border: Border.all(color: Colors.red.shade300),
                                borderRadius: BorderRadius.circular(6),
                              ),
                              child: Text(
                                code,
                                style: TextStyle(
                                  fontSize: 13,
                                  fontWeight: FontWeight.bold,
                                  color: Colors.red.shade800,
                                ),
                              ),
                            );
                          }).toList(),
                        ),
                        SizedBox(height: 8),
                        Text(
                          global.language('product_location_fix_duplicate_before_save'),
                          style: TextStyle(
                            fontSize: 13,
                            color: global.theme.textSecondaryColor,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
            actions: [
              ElevatedButton(
                onPressed: () => Navigator.of(dialogContext).pop(),
                style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.orange,
                  foregroundColor: Colors.white,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                  padding: const EdgeInsets.symmetric(
                    horizontal: 24,
                    vertical: 12,
                  ),
                ),
                child:  Text(
                  global.language('product_location_ok'),
                  style: TextStyle(fontWeight: FontWeight.bold),
                ),
              ),
            ],
          );
        },
      );
      return;
    }

    WarehouseLocationUpdateModel dataSave = WarehouseLocationUpdateModel(
      warehousecode: screenData.code,
      locationcode: locationData.code,
      locationnames: locationData.names,
      shelf: locationData.shelf,
    );

    // print(jsonEncode(dataSave.toJson()));

    context.read<WarehouseLocationBloc>().add(
      WarehouseLocationUpdate(
        warehousecode: selectWarehouseCode,
        locationcode: selectLocationCode,
        warehouseLocationUpdateModel: dataSave,
      ),
    );
  }

  void navigateToShelfProductManagement(String shelfCode, String shelfName) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => ShelfProductManagementScreen(
          warehouseCode: selectWarehouseCode,
          locationCode: selectLocationCode,
          shelfCode: shelfCode,
          warehouseName: global.packName(screenData.names),
          locationName: global.packName(locationData.names),
          shelfName: shelfName,
          screenData: screenData, // ส่ง screenData เข้าไป
        ),
      ),
    ).then((_) {
      // Refresh data when returning from shelf management
      getData(selectWarehouseCode, selectLocationCode);
    });
  }

  /// แสดง dialog สำหรับ import Excel
  void _showImportExcelDialog() {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (BuildContext context) {
        return BlocProvider(
          create: (context) => ShelfExcelImportBloc(
            productBarcodeRepository: ProductBarcodeRepository(),
          ),
          child: BlocConsumer<ShelfExcelImportBloc, ShelfExcelImportState>(
            listener: (context, state) {
              if (state is ExcelImportSuccess) {
                // นำข้อมูล shelves ที่ได้มาเพิ่มเข้า locationData
                setState(() {
                  locationData.shelf.addAll(state.shelves);
                  isDataChange = true;
                });

                Navigator.of(context).pop();

                updateData(selectWarehouseCode, selectLocationCode);

                // แสดง dialog แจ้งผลลัพธ์เฉพาะเมื่อมี processing errors
                if (state.processingErrors.isNotEmpty) {
                  showDialog(
                    context: context,
                    barrierDismissible: false,
                    builder: (BuildContext dialogContext) {
                      return AlertDialog(
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                        title: Row(
                          children: [
                            Icon(
                              state.processingErrors.isEmpty
                                  ? Icons.check_circle
                                  : Icons.warning,
                              color: state.processingErrors.isEmpty
                                  ? Colors.green
                                  : Colors.orange,
                              size: 28,
                            ),
                            const SizedBox(width: 8),
                            Expanded(
                              child: Text(
                                state.processingErrors.isEmpty
                                    ? 'Import สำเร็จ!'
                                    : 'Import เสร็จสิ้น',
                                style: TextStyle(
                                  fontSize: 18,
                                  fontWeight: FontWeight.bold,
                                  color: state.processingErrors.isEmpty
                                      ? Colors.green.shade700
                                      : Colors.orange.shade700,
                                ),
                              ),
                            ),
                          ],
                        ),
                        content: SizedBox(
                          width: 400,
                          child: Column(
                            mainAxisSize: MainAxisSize.min,
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              // ข้อผิดพลาด (ถ้ามี)
                              if (state.processingErrors.isNotEmpty) ...[
                                Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Row(
                                      children: [
                                        Icon(
                                          Icons.error_outline,
                                          color: Colors.red.shade600,
                                          size: 18,
                                        ),
                                        const SizedBox(width: 6),
                                        Text(
                                          '❌ ข้อผิดพลาด (${state.processingErrors.length} รายการ):',
                                          style: TextStyle(
                                            fontWeight: FontWeight.bold,
                                            fontSize: 13,
                                            color: Colors.red.shade800,
                                          ),
                                        ),
                                      ],
                                    ),
                                    const SizedBox(height: 8),
                                    Container(
                                      height: 100,
                                      padding: const EdgeInsets.all(8),
                                      decoration: BoxDecoration(
                                        color: global.theme.cardColor,
                                        borderRadius: BorderRadius.circular(4),
                                        border: Border.all(
                                          color: Colors.red.shade200,
                                        ),
                                      ),
                                      child: ListView.builder(
                                        itemCount:
                                            state.processingErrors.length,
                                        itemBuilder: (context, index) {
                                          return Padding(
                                            padding: const EdgeInsets.symmetric(
                                              vertical: 2,
                                            ),
                                            child: Text(
                                              '• ${state.processingErrors[index]}',
                                              style: TextStyle(
                                                fontSize: 11,
                                                color: Colors.red.shade700,
                                              ),
                                            ),
                                          );
                                        },
                                      ),
                                    ),
                                  ],
                                ),
                              ],
                            ],
                          ),
                        ),
                        actions: [
                          ElevatedButton(
                            onPressed: () => Navigator.of(dialogContext).pop(),
                            style: ElevatedButton.styleFrom(
                              backgroundColor: state.processingErrors.isEmpty
                                  ? Colors.green
                                  : Colors.orange,
                              foregroundColor: Colors.white,
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(8),
                              ),
                              padding: const EdgeInsets.symmetric(
                                horizontal: 24,
                                vertical: 12,
                              ),
                            ),
                            child:  Text(
                              global.language('product_location_ok'),
                              style: TextStyle(fontWeight: FontWeight.bold),
                            ),
                          ),
                        ],
                      );
                    },
                  );
                }
              } else if (state is ExcelImportFailed) {
                String errorMessage = state.message;
                if (state.errorMessages.isNotEmpty) {
                  errorMessage +=
                      '\n\nข้อผิดพลาด:\n${state.errorMessages.take(3).join('\n')}';
                  if (state.errorMessages.length > 3) {
                    errorMessage +=
                        '\n... และอีก ${state.errorMessages.length - 3} รายการ';
                  }
                }

                global.showSnackBar(
                  context,
                  const Icon(Icons.error, color: Colors.white),
                  errorMessage,
                  Colors.red,
                );
              }
            },
            builder: (context, state) {
              return AlertDialog(
                title: Text(global.language('import_shelf_from_excel')),
                content: SizedBox(
                  width: 400,
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    children: [
                      if (state is ExcelImportInProgress)
                        Column(
                          children: [
                            const CircularProgressIndicator(),
                            const SizedBox(height: 16),
                            Text(state.message),
                          ],
                        )
                      else if (state is ExcelImportParsed)
                        Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            // Header ข้อมูลสรุป
                            Container(
                              width: double.infinity,
                              padding: const EdgeInsets.all(12),
                              decoration: BoxDecoration(
                                color: Colors.blue.shade50,
                                border: Border.all(color: Colors.blue.shade200),
                                borderRadius: BorderRadius.circular(8),
                              ),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Row(
                                    children: [
                                      Icon(
                                        Icons.check_circle,
                                        color: Colors.green.shade600,
                                        size: 20,
                                      ),
                                      SizedBox(width: 8),
                                      Text(
                                        global.language('product_location_found_in_file'),
                                        style: TextStyle(
                                          fontWeight: FontWeight.bold,
                                          fontSize: 14,
                                          color: Colors.blue.shade800,
                                        ),
                                      ),
                                    ],
                                  ),
                                  const SizedBox(height: 8),
                                  Text(
                                    '• ชั้นวางทั้งหมด: ${state.shelfGroups.length} ชั้น',
                                  ),
                                  Text(
                                    '• บาร์โค้ดทั้งหมด: ${state.shelfGroups.fold(0, (sum, group) => sum + group.barcodes.length)} รายการ',
                                  ),
                                ],
                              ),
                            ),
                            const SizedBox(height: 12),

                            // รายละเอียดแต่ละชั้นวาง
                            Text(
                              global.language('product_location_shelf_details'),
                              style: TextStyle(
                                fontWeight: FontWeight.bold,
                                fontSize: 14,
                                color: global.theme.textColor,
                              ),
                            ),
                            const SizedBox(height: 8),
                            Container(
                              height: 200,
                              decoration: BoxDecoration(
                                border: Border.all(color: global.theme.dividerBorderColor),
                                borderRadius: BorderRadius.circular(8),
                              ),
                              child: ListView.builder(
                                padding: const EdgeInsets.all(8),
                                itemCount: state.shelfGroups.length,
                                itemBuilder: (context, index) {
                                  final shelf = state.shelfGroups[index];
                                  return Container(
                                    margin: const EdgeInsets.only(bottom: 12),
                                    padding: const EdgeInsets.all(10),
                                    decoration: BoxDecoration(
                                      color: global.theme.inputFillColor,
                                      border: Border.all(
                                        color: global.theme.dividerBorderColor,
                                      ),
                                      borderRadius: BorderRadius.circular(6),
                                    ),
                                    child: Column(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        // ข้อมูลชั้นวาง
                                        Row(
                                          children: [
                                            Container(
                                              padding:
                                                  const EdgeInsets.symmetric(
                                                    horizontal: 8,
                                                    vertical: 4,
                                                  ),
                                              decoration: BoxDecoration(
                                                color: Colors.green.shade100,
                                                borderRadius:
                                                    BorderRadius.circular(12),
                                              ),
                                              child: Text(
                                                '${index + 1}',
                                                style: TextStyle(
                                                  fontSize: 12,
                                                  fontWeight: FontWeight.bold,
                                                  color: Colors.green.shade800,
                                                ),
                                              ),
                                            ),
                                            SizedBox(width: 8),
                                            Expanded(
                                              child: Column(
                                                crossAxisAlignment:
                                                    CrossAxisAlignment.start,
                                                children: [
                                                  Text(
                                                    '${global.language('product_location_code')}: ${shelf.shelfCode}',
                                                    style: const TextStyle(
                                                      fontWeight:
                                                          FontWeight.bold,
                                                      fontSize: 13,
                                                    ),
                                                  ),
                                                  if (shelf
                                                      .shelfName
                                                      .isNotEmpty)
                                                    Text(
                                                      '${global.language('product_location_name')}: ${shelf.shelfName}',
                                                      style: TextStyle(
                                                        fontSize: 12,
                                                        color: Colors
                                                            .grey
                                                            .shade700,
                                                      ),
                                                    ),
                                                ],
                                              ),
                                            ),
                                            Container(
                                              padding:
                                                  const EdgeInsets.symmetric(
                                                    horizontal: 6,
                                                    vertical: 2,
                                                  ),
                                              decoration: BoxDecoration(
                                                color: Colors.blue.shade100,
                                                borderRadius:
                                                    BorderRadius.circular(10),
                                              ),
                                              child: Text(
                                                '${shelf.barcodes.length} รายการ',
                                                style: TextStyle(
                                                  fontSize: 11,
                                                  fontWeight: FontWeight.w500,
                                                  color: Colors.blue.shade800,
                                                ),
                                              ),
                                            ),
                                          ],
                                        ),

                                        // รายการบาร์โค้ด
                                        if (shelf.barcodes.isNotEmpty) ...[
                                          SizedBox(height: 8),
                                          Text(
                                            global.language('product_location_barcode'),
                                            style: TextStyle(
                                              fontSize: 11,
                                              fontWeight: FontWeight.w500,
                                              color: global.theme.textSecondaryColor,
                                            ),
                                          ),
                                          const SizedBox(height: 4),
                                          Wrap(
                                            spacing: 4,
                                            runSpacing: 4,
                                            children: shelf.barcodes.map((
                                              barcode,
                                            ) {
                                              return Container(
                                                padding:
                                                    const EdgeInsets.symmetric(
                                                      horizontal: 6,
                                                      vertical: 2,
                                                    ),
                                                decoration: BoxDecoration(
                                                  color: global.theme.cardColor,
                                                  border: Border.all(
                                                    color: global.theme.dividerBorderColor,
                                                  ),
                                                  borderRadius:
                                                      BorderRadius.circular(4),
                                                ),
                                                child: Text(
                                                  barcode,
                                                  style: const TextStyle(
                                                    fontSize: 10,
                                                  ),
                                                ),
                                              );
                                            }).toList(),
                                          ),
                                        ],
                                      ],
                                    ),
                                  );
                                },
                              ),
                            ),

                            // ข้อผิดพลาด (ถ้ามี)
                            if (state.errorMessages.isNotEmpty) ...[
                              SizedBox(height: 12),
                              Container(
                                width: double.infinity,
                                padding: const EdgeInsets.all(10),
                                decoration: BoxDecoration(
                                  color: Colors.red.shade50,
                                  border: Border.all(
                                    color: Colors.red.shade200,
                                  ),
                                  borderRadius: BorderRadius.circular(8),
                                ),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Row(
                                      children: [
                                        Icon(
                                          Icons.warning,
                                          color: Colors.red.shade600,
                                          size: 18,
                                        ),
                                        SizedBox(width: 6),
                                        Text(
                                          '${global.language('product_location_errors')} (${state.errorMessages.length} ${global.language('product_location_items')}):',
                                          style: TextStyle(
                                            color: Colors.red.shade800,
                                            fontWeight: FontWeight.bold,
                                            fontSize: 13,
                                          ),
                                        ),
                                      ],
                                    ),
                                    const SizedBox(height: 8),
                                    Container(
                                      height: 80,
                                      decoration: BoxDecoration(
                                        color: global.theme.cardColor,
                                        border: Border.all(
                                          color: Colors.red.shade200,
                                        ),
                                        borderRadius: BorderRadius.circular(4),
                                      ),
                                      child: ListView.builder(
                                        padding: const EdgeInsets.all(4),
                                        itemCount: state.errorMessages.length,
                                        itemBuilder: (context, index) {
                                          return Padding(
                                            padding: const EdgeInsets.symmetric(
                                              vertical: 2,
                                            ),
                                            child: Text(
                                              '• ${state.errorMessages[index]}',
                                              style: TextStyle(
                                                fontSize: 11,
                                                color: Colors.red.shade700,
                                              ),
                                            ),
                                          );
                                        },
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ],
                          ],
                        )
                      else
                        Column(
                          crossAxisAlignment: CrossAxisAlignment.center,
                          children: [
                            // File icon ใหญ่ที่กดได้
                            GestureDetector(
                              onTap: () => _pickAndImportExcelFile(context),
                              child: Container(
                                padding: const EdgeInsets.all(20),
                                decoration: BoxDecoration(
                                  color: Colors.green.shade50,
                                  border: Border.all(
                                    color: Colors.green.shade300,
                                    width: 2,
                                  ),
                                  borderRadius: BorderRadius.circular(12),
                                ),
                                child: Column(
                                  children: [
                                    Icon(
                                      Icons.file_upload,
                                      size: 64,
                                      color: Colors.green.shade600,
                                    ),
                                    SizedBox(height: 12),
                                    Text(
                                      global.language('product_location_click_to_select_excel'),
                                      style: TextStyle(
                                        fontSize: 16,
                                        fontWeight: FontWeight.bold,
                                        color: Colors.green.shade700,
                                      ),
                                    ),
                                    SizedBox(height: 8),
                                    Text(
                                      global.language('product_location_supported_files'),
                                      style: TextStyle(
                                        fontSize: 12,
                                        color: global.theme.textSecondaryColor,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ),
                            const SizedBox(height: 16),
                            // คำอธิบายรูปแบบไฟล์
                            Container(
                              padding: const EdgeInsets.all(12),
                              decoration: BoxDecoration(
                                color: global.theme.inputFillColor,
                                border: Border.all(color: global.theme.dividerBorderColor),
                                borderRadius: BorderRadius.circular(8),
                              ),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Row(
                                    children: [
                                      Icon(
                                        Icons.info_outline,
                                        size: 16,
                                        color: Colors.blue.shade600,
                                      ),
                                      SizedBox(width: 6),
                                      Text(
                                        global.language('product_location_excel_format'),
                                        style: TextStyle(
                                          fontWeight: FontWeight.bold,
                                          color: Colors.blue.shade700,
                                          fontSize: 13,
                                        ),
                                      ),
                                    ],
                                  ),
                                  SizedBox(height: 8),
                                   Text(
                                    'Column A: shelf_code (${global.language("shelf_code")})\n'
                                    'Column B: shelf_name (${global.language("shelf_name")})\n'
                                    'Column C: barcode (${global.language("product_barcode")})\n\n'
                                    '${global.language('product_location_note_header_skip')}',
                                    style: TextStyle(fontSize: 12),
                                  ),
                                ],
                              ),
                            ),
                          ],
                        ),
                    ],
                  ),
                ),
                actions: [
                  if (state is ExcelImportParsed) ...[
                    TextButton(
                      onPressed: () => Navigator.of(context).pop(),
                      child: Text(global.language('product_location_cancel')),
                    ),
                    ElevatedButton(
                      onPressed: () {
                        context.read<ShelfExcelImportBloc>().add(
                          ProcessImportedData(shelfGroups: state.shelfGroups),
                        );
                      },
                      child: Text(global.language('product_location_process')),
                    ),
                  ] else if (state is ExcelImportInProgress)
                    const SizedBox.shrink()
                  else ...[
                    TextButton(
                      onPressed: () => Navigator.of(context).pop(),
                      child: Text(global.language('product_location_cancel')),
                    ),
                    ElevatedButton(
                      onPressed: () => _pickAndImportExcelFile(context),
                      child: Text(global.language('select_file')),
                    ),
                  ],
                ],
              );
            },
          ),
        );
      },
    );
  }

  /// เลือกไฟล์ Excel และเริ่มการ import
  Future<void> _pickAndImportExcelFile(BuildContext context) async {
    try {
      FilePickerResult? result = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: ['xlsx', 'xls'],
      );

      if (result != null && result.files.single.bytes != null) {
        Uint8List fileBytes = result.files.single.bytes!;
        String fileName = result.files.single.name;

        if (context.mounted) {
          context.read<ShelfExcelImportBloc>().add(
            ImportExcelFile(fileBytes: fileBytes, fileName: fileName),
          );
        }
      }
    } catch (e) {
      global.showSnackBar(
        context,
        Icon(Icons.error, color: Colors.white),
        '${global.language("error_selecting_file")}: ${e.toString()}',
        Colors.red,
      );
    }
  }

  void warehouseSearch() {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => const ProductWarehouseSearchScreen(word: ""),
      ),
    ).then((value) {
      if (value != null) {
        setState(() {
          SearchGuidCodeNameModel result = value;
          if (result.isCancel == false) {
            screenData.guidfixed = result.guid;
            screenData.code = result.code;
            screenData.names = result.names;

            getWarehouseData(screenData.guidfixed);
          }
        });
      }
    });
  }

  Widget multiShrlf() {
    // Filter shelves based on search text
    List<ShelfModel> filteredShelves = locationData.shelf
        .where(
          (shelf) =>
              shelfFilter.isEmpty ||
              shelf.code.toLowerCase().contains(shelfFilter.toLowerCase()) ||
              shelf.name.toLowerCase().contains(shelfFilter.toLowerCase()),
        )
        .toList();

    // Calculate responsive height based on screen size
    final screenHeight = MediaQuery.of(context).size.height;
    final availableHeight =
        screenHeight -
        300; // Reserve space for app bar, form fields, buttons etc.
    final containerHeight = availableHeight.clamp(
      300.0,
      600.0,
    ); // Min 100, max 600

    return Card(
      margin: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
      elevation: 2,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          // Header Section
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: Theme.of(context).primaryColor.withValues(alpha: 0.05),
              borderRadius: const BorderRadius.only(
                topLeft: Radius.circular(8),
                topRight: Radius.circular(8),
              ),
            ),
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Title and count
                Row(
                  children: [
                    Icon(
                      Icons.view_module,
                      color: Theme.of(context).primaryColor,
                      size: 20,
                    ),
                    SizedBox(width: 8),
                    Text(
                      global.language("shelf"),
                      style: Theme.of(context).textTheme.titleMedium?.copyWith(
                        fontWeight: FontWeight.w600,
                        color: Theme.of(context).primaryColor,
                      ),
                    ),
                    const Spacer(),
                    Chip(
                      label: Text(
                        shelfFilter.isEmpty
                            ? "${locationData.shelf.length}"
                            : "${filteredShelves.length}/${locationData.shelf.length}",
                        style: const TextStyle(
                          fontSize: 12,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                      backgroundColor: Theme.of(
                        context,
                      ).primaryColor.withValues(alpha: 0.1),
                      side: BorderSide.none,
                      padding: const EdgeInsets.symmetric(horizontal: 4),
                    ),
                    if (isEditMode) ...[
                      SizedBox(width: 8),
                      Chip(
                        label: Text(
                          global.language("edit"),
                          style: const TextStyle(
                            fontSize: 11,
                            fontWeight: FontWeight.w600,
                          ),
                        ),
                        backgroundColor: Colors.green.withValues(alpha: 0.1),
                        labelStyle: TextStyle(color: Colors.green.shade700),
                        side: BorderSide.none,
                        padding: const EdgeInsets.symmetric(horizontal: 4),
                      ),
                    ],
                  ],
                ),

                const SizedBox(height: 12),

                // Search bar
                TextField(
                  decoration: InputDecoration(
                    hintText: "${global.language("search_shelf")}...",
                    prefixIcon: const Icon(Icons.search, size: 20),
                    suffixIcon: shelfFilter.isNotEmpty
                        ? IconButton(
                            icon: const Icon(Icons.clear, size: 20),
                            onPressed: () {
                              setState(() {
                                shelfFilter = "";
                              });
                            },
                          )
                        : null,
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(8),
                      borderSide: BorderSide(color: global.theme.dividerBorderColor),
                    ),
                    enabledBorder: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(8),
                      borderSide: BorderSide(color: global.theme.dividerBorderColor),
                    ),
                    focusedBorder: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(8),
                      borderSide: BorderSide(
                        color: Theme.of(context).primaryColor,
                      ),
                    ),
                    contentPadding: const EdgeInsets.symmetric(
                      horizontal: 12,
                      vertical: 8,
                    ),
                    filled: true,
                    fillColor: global.theme.inputFillColor,
                  ),
                  onChanged: (value) {
                    setState(() {
                      shelfFilter = value;
                    });
                  },
                ),
              ],
            ),
          ),

          // Content Area
          SizedBox(
            height: containerHeight,
            child: isShelfLoading
                ? const Center(child: CircularProgressIndicator())
                : filteredShelves.isEmpty
                ? _buildEmptyState()
                : ListView.builder(
                    padding: const EdgeInsets.all(16),
                    itemCount: filteredShelves.length,
                    itemBuilder: (context, index) {
                      final shelfIndex = locationData.shelf.indexOf(
                        filteredShelves[index],
                      );
                      return Padding(
                        padding: const EdgeInsets.only(bottom: 8),
                        child: _buildShelfItem(
                          shelfIndex,
                          filteredShelves[index],
                        ),
                      );
                    },
                  ),
          ),

          // Add button
          Container(
            padding: const EdgeInsets.all(16),
            decoration: BoxDecoration(
              color: global.theme.inputFillColor,
              borderRadius: const BorderRadius.only(
                bottomLeft: Radius.circular(8),
                bottomRight: Radius.circular(8),
              ),
            ),
            child: SizedBox(
              width: double.infinity,
              child: ElevatedButton.icon(
                onPressed: isEditMode
                    ? () {
                        locationData.shelf.add(ShelfModel(code: "", name: ""));

                        setState(() {
                          isDataChange = true;
                        });
                      }
                    : null,
                icon: Icon(Icons.add),
                label: Text(
                  isEditMode ? global.language("add_shelf") : global.language("enable_edit_to_add"),
                ),
                style: ElevatedButton.styleFrom(
                  backgroundColor: isEditMode
                      ? Theme.of(context).primaryColor
                      : global.theme.textSecondaryColor,
                  foregroundColor: Colors.white,
                  padding: const EdgeInsets.symmetric(vertical: 12),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                ),
              ),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.inventory_2_outlined,
            size: 64,
            color: global.theme.textSecondaryColor,
          ),
          SizedBox(height: 16),
          Text(
            shelfFilter.isEmpty ? global.language("no_shelves_yet") : global.language("shelf_not_found"),
            style: TextStyle(
              fontSize: 18,
              color: global.theme.textSecondaryColor,
              fontWeight: FontWeight.w500,
            ),
          ),
          SizedBox(height: 8),
          Text(
            shelfFilter.isEmpty
                ? global.language("start_by_adding_first_shelf")
                : global.language("try_different_search"),
            style: TextStyle(fontSize: 14, color: global.theme.textSecondaryColor),
          ),
          if (shelfFilter.isNotEmpty) ...[
            SizedBox(height: 16),
            TextButton.icon(
              onPressed: () {
                setState(() {
                  shelfFilter = "";
                });
              },
              icon: Icon(Icons.clear, size: 16),
              label: Text(global.language("clear_search")),
            ),
          ],
        ],
      ),
    );
  }

  Widget _buildShelfItem(int shelfIndex, ShelfModel shelf) {
    final bool hasCode = shelf.code.trim().isNotEmpty;
    final bool hasName = shelf.name.trim().isNotEmpty;
    final bool isComplete = hasCode && hasName;

    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(
          color: isComplete ? Colors.green.shade300 : Colors.orange.shade300,
          width: 1,
        ),
      ),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Row(
          children: [
            // Shelf number
            Container(
              width: 28,
              height: 28,
              decoration: BoxDecoration(
                color: isComplete
                    ? Colors.green.shade100
                    : Colors.orange.shade100,
                borderRadius: BorderRadius.circular(6),
                border: Border.all(
                  color: isComplete
                      ? Colors.green.shade300
                      : Colors.orange.shade300,
                ),
              ),
              child: Center(
                child: Text(
                  "${shelfIndex + 1}",
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                    color: isComplete
                        ? Colors.green.shade700
                        : Colors.orange.shade700,
                  ),
                ),
              ),
            ),

            const SizedBox(width: 12),

            // Input fields
            Expanded(
              child: Row(
                children: [
                  // Code field
                  Expanded(
                    flex: 1,
                    child: TextFormField(
                      readOnly: !isEditMode,
                      initialValue: shelf.code,
                      textCapitalization: TextCapitalization.characters,
                      inputFormatters: <TextInputFormatter>[
                        FilteringTextInputFormatter.allow(
                          RegExp('[a-z A-Z 0-9 -]'),
                        ),
                      ],
                      onChanged: (value) {
                        isDataChange = true;
                        locationData.shelf[shelfIndex].code = value
                            .toUpperCase();
                        setState(() {});
                      },
                      style: const TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w500,
                      ),
                      decoration: InputDecoration(
                        hintText: global.language("code"),
                        hintStyle: TextStyle(
                          fontSize: 12,
                          color: global.theme.textSecondaryColor,
                        ),
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(6),
                          borderSide: BorderSide(color: global.theme.dividerBorderColor),
                        ),
                        enabledBorder: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(6),
                          borderSide: BorderSide(color: global.theme.dividerBorderColor),
                        ),
                        focusedBorder: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(6),
                          borderSide: BorderSide(
                            color: Theme.of(context).primaryColor,
                          ),
                        ),
                        contentPadding: const EdgeInsets.symmetric(
                          horizontal: 8,
                          vertical: 8,
                        ),
                        filled: true,
                        fillColor: isEditMode
                            ? global.theme.cardColor
                            : global.theme.inputFillColor,
                      ),
                    ),
                  ),

                  const SizedBox(width: 8),

                  // Name field
                  Expanded(
                    flex: 2,
                    child: TextFormField(
                      readOnly: !isEditMode,
                      initialValue: shelf.name,
                      onChanged: (value) {
                        isDataChange = true;
                        locationData.shelf[shelfIndex].name = value;
                        setState(() {});
                      },
                      style: const TextStyle(
                        fontSize: 14,
                        fontWeight: FontWeight.w500,
                      ),
                      decoration: InputDecoration(
                        hintText: global.language("shelf_name"),
                        hintStyle: TextStyle(
                          fontSize: 12,
                          color: global.theme.textSecondaryColor,
                        ),
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(6),
                          borderSide: BorderSide(color: global.theme.dividerBorderColor),
                        ),
                        enabledBorder: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(6),
                          borderSide: BorderSide(color: global.theme.dividerBorderColor),
                        ),
                        focusedBorder: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(6),
                          borderSide: BorderSide(
                            color: Theme.of(context).primaryColor,
                          ),
                        ),
                        contentPadding: const EdgeInsets.symmetric(
                          horizontal: 8,
                          vertical: 8,
                        ),
                        filled: true,
                        fillColor: isEditMode
                            ? global.theme.cardColor
                            : global.theme.inputFillColor,
                      ),
                    ),
                  ),
                ],
              ),
            ),

            const SizedBox(width: 8),

            // Action buttons
            Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                // Manage products button
                IconButton(
                  onPressed: (hasCode || hasName)
                      ? () {
                          navigateToShelfProductManagement(
                            shelf.code,
                            shelf.name.isNotEmpty ? shelf.name : shelf.code,
                          );
                        }
                      : null,
                  icon: const Icon(Icons.inventory_2_outlined, size: 18),
                  style: IconButton.styleFrom(
                    foregroundColor: (hasCode || hasName)
                        ? Theme.of(context).primaryColor
                        : global.theme.textSecondaryColor,
                    backgroundColor: (hasCode || hasName)
                        ? Theme.of(context).primaryColor.withValues(alpha: 0.1)
                        : global.theme.inputFillColor,
                    padding: const EdgeInsets.all(8),
                  ),
                  tooltip: global.language("manage_products"),
                ),

                // Delete button
                if (isEditMode && locationData.shelf.length > 1) ...[
                  SizedBox(width: 4),
                  IconButton(
                    onPressed: () => _deleteShelf(shelfIndex, shelf),
                    icon: const Icon(Icons.delete_outline, size: 18),
                    style: IconButton.styleFrom(
                      foregroundColor: Colors.red.shade600,
                      backgroundColor: Colors.red.withValues(alpha: 0.1),
                      padding: const EdgeInsets.all(8),
                    ),
                    tooltip: global.language("delete_shelf"),
                  ),
                ],
              ],
            ),
          ],
        ),
      ),
    );
  }

  void _deleteShelf(int shelfIndex, ShelfModel shelf) {
    showDialog(
      context: context,
      builder: (context) => AlertDialog(
        title: Text(global.language("confirm_delete")),
        content: Text(
          "${global.language("confirm_delete_shelf")} \"${shelf.name.isEmpty ? shelf.code : shelf.name}\"?",
        ),
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(12)),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(context).pop(),
            child: Text(global.language("cancel")),
          ),
          ElevatedButton(
            onPressed: () {
              setState(() {
                locationData.shelf.removeAt(shelfIndex);
                isDataChange = true;
              });
              Navigator.of(context).pop();
            },
            style: ElevatedButton.styleFrom(
              backgroundColor: Colors.red,
              foregroundColor: Colors.white,
            ),
            child: Text(global.language("delete")),
          ),
        ],
      ),
    );
  }

  Widget editScreen({mobileScreen}) {
    List<Widget> formWidgets = [];

    formWidgets.add(
      Container(
        margin: const EdgeInsets.only(bottom: 10),
        padding: const EdgeInsets.only(left: 10, right: 10),
        child: Row(
          children: [
            Expanded(
              child: RawKeyboardListener(
                focusNode: FocusNode(),
                onKey: (RawKeyEvent event) {
                  if (event is RawKeyDownEvent) {
                    if (event.logicalKey == LogicalKeyboardKey.f2) {
                      warehouseSearch();
                    }
                  }
                },
                child: TextField(
                  readOnly: true,
                  textInputAction: TextInputAction.next,
                  controller: TextEditingController(text: screenData.code),
                  textAlign: TextAlign.left,
                  textCapitalization: TextCapitalization.characters,
                  onChanged: (code) {
                    screenData.code = code;
                  },
                  decoration: InputDecoration(
                    contentPadding: EdgeInsets.all(10.0),
                    enabledBorder: OutlineInputBorder(
                      borderSide: BorderSide(color: global.theme.dividerBorderColor, width: 0.0),
                    ),
                    floatingLabelBehavior: FloatingLabelBehavior.always,
                    suffixIcon: Row(
                      mainAxisAlignment:
                          MainAxisAlignment.spaceBetween, // added line
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        IconButton(
                          focusNode: FocusNode(skipTraversal: true),
                          icon: const Icon(Icons.search),
                          onPressed: () {
                            if (isEditMode) {
                              warehouseSearch();
                            }
                          },
                        ),
                      ],
                    ),
                    border: OutlineInputBorder(),
                    labelText: global.language("warehouse_code"),
                  ),
                ),
              ),
            ),
            SizedBox(width: 5),
            Expanded(
              child: TextField(
                readOnly: true,
                controller: TextEditingController(
                  text: global.packName(screenData.names),
                ),
                textAlign: TextAlign.left,
                textCapitalization: TextCapitalization.characters,
                decoration: InputDecoration(
                  contentPadding: EdgeInsets.all(10.0),
                  enabledBorder: OutlineInputBorder(
                    borderSide: BorderSide(color: global.theme.dividerBorderColor, width: 0.0),
                  ),
                  floatingLabelBehavior: FloatingLabelBehavior.always,
                  border: OutlineInputBorder(),
                  labelText: global.language("warehouse_name"),
                ),
              ),
            ),
          ],
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
        child: TextField(
          readOnly: !isEditMode,
          onSubmitted: (value) {
            if (kIsWeb) {
              // findFocusNext(focusNodeIndex); // removed focus navigation
            }
          },
          textAlign: TextAlign.left,
          controller: TextEditingController(text: locationData.code),
          textCapitalization: TextCapitalization.characters,
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.allow(RegExp('[a-z A-Z 0-9 -]')),
          ],
          onChanged: (value) {
            isDataChange = true;
            locationData.code = value.toUpperCase();
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("location_code"),
          ),
        ),
      ),
    );
    for (
      int languageIndex = 0;
      languageIndex < languageList.length;
      languageIndex++
    ) {
      LanguageDataModel locationName = locationData.names.firstWhere(
        (element) => element.code == languageList[languageIndex].code,
        orElse: () => LanguageDataModel(code: '', name: ''),
      );
      if (locationName.code == '') {
        locationData.names.add(
          LanguageDataModel(code: languageList[languageIndex].code!, name: ''),
        );
      }
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
          child: TextField(
            readOnly: !isEditMode,
            onChanged: (value) {
              isDataChange = true;
              locationData.names[languageIndex].name = value;
            },
            onSubmitted: (value) {
              if (kIsWeb) {
                // findFocusNext(focusNodeIndex); // removed focus navigation
              }
            },
            textAlign: TextAlign.left,
            controller: TextEditingController(
              text: locationData.names[languageIndex].name,
            ),
            decoration: InputDecoration(
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: const OutlineInputBorder(),
              labelText:
                  "${global.language("location_name")} (${getLangName(locationData.names[languageIndex].code)})",
              suffixIcon: isLoadTranslation
                  ? const Padding(
                      padding: EdgeInsets.all(8.0),
                      child: SizedBox(
                        width: 20,
                        height: 20,
                        child: CircularProgressIndicator(strokeWidth: 2),
                      ),
                    )
                  : null,
            ),
          ),
        ),
      );
    }

    formWidgets.add(const SizedBox(height: 5));
    formWidgets.add(multiShrlf());

    formWidgets.add(const SizedBox(height: 10));

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
        title: Text(headerEdit + global.language("product_location")),
        actions: <Widget>[
          /// Import Excel
          if (screenEvent == global.ScreenEventEnum.edit && isEditMode)
            Padding(
              padding: EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                tooltip: global.language('import_shelf_from_excel'),
                onPressed: () {
                  _showImportExcelDialog();
                },
                icon: const Icon(Icons.upload_file, color: Colors.green),
              ),
            ),
          if (selectWarehouseCode.isNotEmpty && selectLocationCode.isNotEmpty)
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
                            guidListChecked.add(selectLocationCode);
                            Navigator.pop(context);
                            context.read<WarehouseLocationBloc>().add(
                              WarehouseLocationDeleteMany(
                                warehousecode: selectWarehouseCode,
                                locationcode: guidListChecked,
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
          if (isSaveAllow == false &&
              selectWarehouseCode.trim().isNotEmpty &&
              selectLocationCode.trim().isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () {
                  showCheckBox = false;
                  switchToEdit(
                    listData[listData.indexOf(
                      listData.firstWhere(
                        (element) =>
                            element.warehousecode == selectWarehouseCode &&
                            element.locationcode == selectLocationCode,
                      ),
                    )],
                  );
                },
                icon: const Icon(Icons.edit),
              ),
            ),
          if (isEditMode && global.systemLanguage.length > 1)
            Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () async {
                  setState(() {
                    isLoadTranslation = true;
                  });

                  try {
                    for (var name in locationData.names) {
                      if (name.name.trim().isEmpty) {
                        var data = await global.translateNames(
                          namesData: locationData.names,
                        );
                        setState(() {
                          locationData.names = data;
                        });
                        break;
                      }
                    }
                  } catch (e) {
                    if (kDebugMode) {
                      AppLogger.debug(e);
                    }
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
          child: SingleChildScrollView(
                  controller: editScrollController,
                  child: Container(
                    width: double.infinity,
                    color: global.theme.cardColor,
                    padding: const EdgeInsets.only(top: 10, bottom: 10),
                    child: Form(child: Column(children: formWidgets)),
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
              BlocListener<WarehouseLocationBloc, WarehouseLocationState>(
                listener: (context, state) {
                  blocWarehouseLocationState = state;

                  // Load
                  if (state is WarehouseLocationLoadSuccess) {
                    setState(() {
                      loadingData = false;
                      if (state.warehouses.isNotEmpty) {
                        listData.addAll(state.warehouses);
                      }
                    });
                  }
                  if (state is WarehouseLocationLoadFailed) {
                    setState(() {
                      loadingData = false;
                    });
                  }

                  // Update
                  if (state is WarehouseLocationUpdateSuccess) {
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
                      getData(selectWarehouseCode, selectLocationCode);
                    });
                  }
                  if (state is WarehouseLocationUpdateFailed) {
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
                  if (state is WarehouseLocationGetSuccess) {
                    setState(() {
                      isDataChange = false;

                      screenData.guidfixed = state
                          .warehouselocation
                          .guidfixed; // Fix: Set GUID from API response
                      screenData.code = state.warehouselocation.warehousecode;
                      screenData.names = state.warehouselocation.warehousenames;

                      warehouseCodeController.text = screenData.code;

                      locationData.code = state.warehouselocation.locationcode;
                      locationData.names =
                          state.warehouselocation.locationnames;

                      locationData.shelf = state.warehouselocation.shelf;

                      if (screenEvent == global.ScreenEventEnum.edit) {
                        WidgetsBinding.instance.addPostFrameCallback((
                          timeStamp,
                        ) {
                          tabController.animateTo(1);
                        });
                        setState(() {
                          // removed findFocusNext(0);
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
                  // Delete Many
                  if (state is WarehouseLocationDeleteManySuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.delete, color: Colors.white),
                        global.language("delete_success"),
                        Colors.blue,
                      );
                      listData.clear();
                      clearEditData();
                      loadDataList(searchText);
                      showCheckBox = false;
                    });
                  }
                },
              ),
              BlocListener<WarehouseBloc, WarehouseState>(
                listener: (context, state) {
                  blocWarehouseState = state;
                  // Update
                  if (state is WarehouseUpdateSuccess) {
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
                      getData(selectWarehouseCode, selectLocationCode);
                    });
                  }
                  if (state is WarehouseUpdateFailed) {
                    clearEditData();
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.edit, color: Colors.white),
                        "${global.language("not_edit_success")} : ${state.message}",
                        Colors.red,
                      );
                    });
                  }

                  /// Get Warehouse by guid
                  if (state is WarehouseGetSuccess) {
                    setState(() {
                      warehouseData = state.warehouse;
                      // removed findFocusNext(0);
                    });
                  }
                },
              ),
            ],
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
