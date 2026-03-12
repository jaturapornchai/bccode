import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:smlaicloud/bloc/creditor/creditor_bloc.dart';
import 'package:smlaicloud/bloc/creditor_group/creditor_group_bloc.dart';
import 'package:smlaicloud/model/creditor_group_model.dart';
import 'package:smlaicloud/model/creditor_model.dart';
import 'package:smlaicloud/model/customer_address_model.dart';
import 'package:smlaicloud/model/thai_province_model.dart';
import 'package:dropdown_search/dropdown_search.dart';
import 'package:email_validator/email_validator.dart';
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
import 'package:file_picker/file_picker.dart';
import 'package:excel/excel.dart' as xl;
import 'package:smlaicloud/screens/report/file_download.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/widgets/manual_button.dart';

class CreditorScreen extends StatefulWidget {
  const CreditorScreen({super.key});

  @override
  State<CreditorScreen> createState() => CreditorScreenState();
}

class CreditorScreenState extends State<CreditorScreen>
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
  List<CreditorModel> listData = [];
  List<CreditorGroupModel> listDataGroup = [];
  List<String> guidListChecked = [];
  List<LanguageDataModel> names = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  bool isDataChange = false;
  bool isSaveAllow = false;
  late CreditorState blocCurrentState;
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  int _hoverIndex = -1;
  bool isEditMode = false;
  late CreditorModel screenData;
  List<Uint8List> imageWeb = [];
  final ImagePicker imagePicker = ImagePicker();
  late DropzoneViewController dropZoneController;
  Color colorSelected = Colors.white;
  final _debouncer = global.Debouncer(1000);
  late Timer screenTimer;
  bool loadingData = false;
  List<CreditorGroupModel> groupSelected = [];
  List<File> imageFile = [];
  global.ScreenEventEnum screenEvent = global.ScreenEventEnum.list;
  late SplitViewController splitViewController;
  final _popupBuilderKey = GlobalKey<DropdownSearchState<String>>();
  final _popupBuilderKey2 = GlobalKey<DropdownSearchState<String>>();
  TextEditingController creditorCode = TextEditingController();

  List<CreditorGroupModel> selectedFilters = [];
  List<String> selectedFilterCodes = [];

  final _formKey = GlobalKey<FormState>();

  bool isLoadTranslation = false;

  List<ThaiProvinceModel> thaiProvinceList = [];

  void setSystemLanguageList() async {
    clearEditData();
    await global.setSystemLanguage(context);

    for (int i = 0; i < global.config.languages.length; i++) {
      if (global.config.languages[i].isuse!) {
        languageList.add(global.config.languages[i]);
      }
    }
    loadDataList("", []);
    loadDataGroupList();
  }

  @override
  void initState() {
    getProvinceData();
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

  Future<void> getProvinceData() async {
    const githubRawUrl =
        'https://raw.githubusercontent.com/smlsoft/dedepos_template/main/thai-province-data/thai_provinces.json';
    try {
      final fileContent = await global.readFileFromGithub(githubRawUrl);
      final provinces = (json.decode(fileContent) as List)
          .map((province) => ThaiProvinceModel.fromJson(province))
          .toList();
      thaiProvinceList = [];

      for (int i = 0; i < provinces.length; i++) {
        thaiProvinceList.add(provinces[i]);
      }
    } catch (error) {
      // Handle error
      // ignore: avoid_print
      if (kDebugMode) {
        AppLogger.error('Error reading file: $error');
      }
    }
  }

  void loadDataGroupList() {
    context.read<CreditorGroupBloc>().add(
      const CreditorGroupLoadList(offset: 0, limit: 1000, search: ""),
    );
  }

  void loadDataList(String search, List<String>? filter) {
    setState(() {
      loadingData = true;
    });
    searchText = search;

    context.read<CreditorBloc>().add(
      CreditorLoadList(
        offset: (listData.isEmpty) ? 0 : listData.length,
        limit: global.loadDataPerPage,
        search: search,
        groups: filter,
      ),
    );
  }

  void onScrollList() {
    if (listScrollController.offset >=
            listScrollController.position.maxScrollExtent &&
        !listScrollController.position.outOfRange) {
      loadDataList(searchText, selectedFilterCodes);
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
    creditorCode.text = "";
    screenData = CreditorModel(
      guidfixed: "",
      code: "",
      names: names,
      customertype: 1,
      branchnumber: '00000',
      personaltype: 1,
      addressforbilling: CustomerAddressModel(
        guid: "",
        address: ["", "", ""],
        countrycode: "",
        provincecode: "",
        districtcode: "",
        subdistrictcode: "",
        zipcode: "",
        latitude: 0,
        longitude: 0,
        contactnames: nameBill,
        phoneprimary: "",
        phonesecondary: "",
      ),
      addressforshipping: [],
      images: [],
      groups: [],
      taxid: "",
      email: "",
    );
    imageFile = [];
    imageWeb = [];
    groupSelected = [];
    isDataChange = false;
    focusNodeIndex = 0;
    refreshFocus = true;
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
    context.read<CreditorBloc>().add(CreditorGet(guid: guid));
  }

  // ฟังก์ชันสำหรับ import Excel
  Future<void> _importFromExcel() async {
    try {
      FilePickerResult? result = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: ['xlsx', 'xls'],
      );

      if (result != null && result.files.single.bytes != null) {
        var bytes = result.files.single.bytes!;
        var excel = xl.Excel.decodeBytes(bytes);

        List<CreditorRequestModel> importedData = [];

        for (var table in excel.tables.keys) {
          var sheet = excel.tables[table];
          if (sheet != null) {
            // ข้าม header row (row 0)
            for (int row = 1; row < sheet.maxRows; row++) {
              var rowData = <xl.Data?>[];
              if (sheet.maxColumns > 0) {
                for (int col = 0; col < sheet.maxColumns; col++) {
                  rowData.add(
                    sheet.cell(
                      xl.CellIndex.indexByColumnRow(
                        columnIndex: col,
                        rowIndex: row,
                      ),
                    ),
                  );
                }

                CreditorRequestModel? creditor = _parseExcelRowToCreditor(
                  rowData,
                );
                if (creditor != null) {
                  importedData.add(creditor);
                }
              }
            }
          }
        }

        if (importedData.isNotEmpty) {
          // บันทึกข้อมูลเจ้าหนี้หลายรายการ
          _saveBulkCreditors(importedData);

          // แสดงข้อความสำเร็จจะทำใน BlocListener
        } else {
          global.showWarningSnackBar(context, global.language('no_valid_data_in_excel'));
        }
      }
    } catch (e) {
      global.showErrorSnackBar(context, '${global.language('import_error')}: $e');
    }
  }

  // ฟังก์ชันแปลง Excel row เป็น CreditorRequestModel
  CreditorRequestModel? _parseExcelRowToCreditor(List<xl.Data?> rowData) {
    try {
      // Expected columns: code, names, personaltype, taxid, customertype, branchnumber, email, creditday, address, phonepreimary
      if (rowData.length < 10) return null;

      String code = rowData[0]?.value?.toString() ?? '';
      String names = rowData[1]?.value?.toString() ?? '';
      int personaltype =
          int.tryParse(rowData[2]?.value?.toString() ?? '1') ?? 1;
      String taxid = rowData[3]?.value?.toString() ?? '';
      int customertype =
          int.tryParse(rowData[4]?.value?.toString() ?? '1') ?? 1;
      String branchnumber = rowData[5]?.value?.toString() ?? '';
      String email = rowData[6]?.value?.toString() ?? '';
      int creditday = int.tryParse(rowData[7]?.value?.toString() ?? '0') ?? 0;
      String address = rowData[8]?.value?.toString() ?? '';
      String phonepreimary = rowData[9]?.value?.toString() ?? '';

      if (code.isEmpty || names.isEmpty) return null;

      // สร้าง LanguageDataModel สำหรับ names
      List<LanguageDataModel> namesList = [
        LanguageDataModel(code: "th", name: names),
      ];

      // สร้าง CustomerAddressModel สำหรับที่อยู่
      CustomerAddressModel addressforbilling = CustomerAddressModel(
        address: [address, "", ""], // แบ่งเป็น 3 ส่วน
        phoneprimary: phonepreimary,
        contactnames: [], // ใช้ค่าเริ่มต้น
        zipcode: "",
        provincecode: "",
        districtcode: "",
        subdistrictcode: "",
        phonesecondary: "",
      );

      return CreditorRequestModel(
        code: code,
        personaltype: personaltype,
        names: namesList,
        addressforbilling: addressforbilling,
        addressforshipping: [], // ใช้ค่าเริ่มต้น
        taxid: taxid,
        customertype: customertype,
        branchnumber: branchnumber,
        email: email,
        creditday: creditday,
        groups: [], // ใช้ค่าเริ่มต้น
        images: [], // ใช้ค่าเริ่มต้น
        guidfixed: "", // จะถูกกำหนดในการบันทึก
      );
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error parsing row data: $e');
      }
      return null;
    }
  }

  // ฟังก์ชันสำหรับบันทึกข้อมูลเจ้าหนี้หลายรายการ
  Future<void> _saveBulkCreditors(List<CreditorRequestModel> creditors) async {
    try {
      // ใช้ bulk save method ใหม่
      context.read<CreditorBloc>().add(CreditorBulkSave(creditors: creditors));
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error saving bulk creditors: $e');
      }
      rethrow;
    }
  }

  Future<void> downloadAssetFile() async {
    try {
      // Load the asset file
      ByteData data = await rootBundle.load(
        'assets/file_import/import_creditor.xlsx',
      );
      Uint8List bytes = data.buffer.asUint8List();

      // Download the file
      await downloadAssetFileBytes(bytes, 'import_creditor_template.xlsx');

      if (mounted) {
        global.showInfoSnackBar(context, 'Template downloaded successfully');
      }
    } catch (e) {
      if (mounted) {
        global.showInfoSnackBar(context, 'Error downloading template: $e');
      }
    }
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('creditor')),
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
                            context.read<CreditorBloc>().add(
                              CreditorDeleteMany(guid: guidListChecked),
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

          /// Import Excel
          Padding(
            padding: const EdgeInsets.only(right: 20.0),
            child: IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () {
                _importFromExcel();
              },
              icon: const Icon(Icons.upload_file),
              tooltip: global.language('import_from_excel'),
            ),
          ),

          /// Download Template
          Padding(
            padding: const EdgeInsets.only(right: 20.0),
            child: IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () {
                downloadAssetFile();
              },
              icon: const Icon(Icons.download),
              tooltip: global.language('download_template'),
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
                      fieldFocusNodes[0].focusNode.requestFocus();
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
                        _debouncer.run(() {
                          setState(() {
                            listData = [];
                          });
                          loadDataList(value, selectedFilterCodes);
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
                  IconButton(
                    onPressed: () async {
                      selectedFilters = await filterCrediterGroup(
                        selectedFilters,
                      );
                      if (selectedFilters.isNotEmpty) {
                        selectedFilterCodes.clear();
                        for (var element in selectedFilters) {
                          selectedFilterCodes.add(element.guidfixed);
                        }
                      } else {
                        selectedFilterCodes.clear();
                      }
                      listData = [];
                      loadDataList(searchText, selectedFilterCodes);
                      setState(() {});
                    },
                    icon: Icon(
                      (selectedFilters.isEmpty)
                          ? Icons.filter_alt_off
                          : Icons.filter_alt,
                      color: (selectedFilters.isEmpty)
                          ? global.theme.iconColor
                          : global.theme.primaryColor,
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
                    flex: 3,
                    child: Text(
                      global.language("creditor_code"),
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
                      global.language("creditor_name"),
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
                      global.language("address"),
                      style: TextStyle(
                        color: global.theme.columnHeaderTextColor,
                        fontWeight: FontWeight.bold,
                        fontSize: global.deviceConfig.listDataFontSize + 2,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 3,
                    child: Text(
                      global.language("telephone"),
                      style: TextStyle(
                        color: global.theme.columnHeaderTextColor,
                        fontWeight: FontWeight.bold,
                        fontSize: global.deviceConfig.listDataFontSize + 2,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 3,
                    child: Text(
                      global.language("fundcode"),
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

  void switchToEdit(CreditorModel value) {
    setState(() {
      selectGuid = value.guidfixed;
      getData(selectGuid);
      headerEdit = global.language("edit");
      isSaveAllow = true;
      isEditMode = true;
    });
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

  Widget listObject(int index, CreditorModel value, bool showCheckBox) {
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
              flex: 3,
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
                global.packName(value.names),
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 5,
              child: Text(
                (value.addressforbilling.address!.isNotEmpty)
                    ? value.addressforbilling.address![0]
                    : '',
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 3,
              child: Text(
                (value.addressforbilling.phoneprimary!.isNotEmpty)
                    ? value.addressforbilling.phoneprimary!
                    : '',
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
                style: textStyle,
              ),
            ),
            Expanded(
              flex: 3,
              child: Text(
                value.fundcode,
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

    screenData.groups = groupSelected;

    if (selectGuid.trim().isEmpty) {
      if (imageFile.isNotEmpty) {
        context.read<CreditorBloc>().add(
          CreditorWithImageSave(
            creditor: screenData,
            imageFile: imageFile,
            imageWeb: imageWeb,
          ),
        );
      } else {
        context.read<CreditorBloc>().add(CreditorSave(creditor: screenData));
      }
    } else {
      updateData(selectGuid);
    }
  }

  void updateData(String guid) {
    showCheckBox = false;
    List<File> imageFileUpdate = [];
    List<Uint8List> imageWebUpdate = [];
    List<ImagesModel> imageUris = [];
    for (int i = 0; i < imageWeb.length; i++) {
      // print(imageWeb.length);
      // print(imageFile.length);
      // print(screenData.images.length);
      if (imageWeb[i].isNotEmpty || screenData.images[i].uri != '') {
        imageFileUpdate.add(imageFile[i]);
        imageWebUpdate.add(imageWeb[i]);
        imageUris.add(ImagesModel(uri: screenData.images[i].uri, xorder: i));
      }
    }
    // print("imageWebUpdate.isNotEmpty " + imageWebUpdate.isNotEmpty.toString());
    if (imageWebUpdate.isNotEmpty) {
      context.read<CreditorBloc>().add(
        CreditorWithImageUpdate(
          guid: guid,
          creditor: screenData,
          imageFiles: imageFile,
          imagesUris: imageUris,
          imageWeb: imageWeb,
        ),
      );
    } else {
      screenData.images = [];
      context.read<CreditorBloc>().add(
        CreditorUpdate(guid: guid, creditorModel: screenData),
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

  void getDataToEditScreen(CreditorModel creditor) {
    isDataChange = false;
    selectGuid = creditor.guidfixed;
    creditorCode.text = creditor.code;
    screenData.code = creditor.code;
    screenData.email = creditor.email;
    screenData.fundcode = creditor.fundcode;
    screenData.creditday = creditor.creditday;
    screenData.guidfixed = creditor.guidfixed;
    screenData.images = creditor.images;

    screenData.personaltype = creditor.personaltype;
    screenData.customertype = creditor.customertype;
    screenData.branchnumber = creditor.branchnumber;
    screenData.taxid = creditor.taxid;
    screenData.groups = creditor.groups;
    groupSelected = creditor.groups;
    screenData.addressforbilling.address = creditor.addressforbilling.address;
    screenData.addressforbilling.phoneprimary =
        creditor.addressforbilling.phoneprimary;
    screenData.addressforbilling.phonesecondary =
        creditor.addressforbilling.phonesecondary;

    screenData.names = [];
    screenData.addressforbilling.contactnames = [];
    screenData.addressforshipping = [];
    //เพิ่มภาษาตาม config
    for (var lang in languageList) {
      screenData.names.add(LanguageDataModel(code: lang.code!, name: ""));
      screenData.addressforbilling.contactnames!.add(
        LanguageDataModel(code: lang.code!, name: ""),
      );
    }

    for (var addressforshipping in creditor.addressforshipping) {
      CustomerAddressModel newData = CustomerAddressModel(
        guid: addressforshipping.guid,
        address: addressforshipping.address,
        countrycode: addressforshipping.countrycode,
        provincecode: addressforshipping.provincecode,
        districtcode: addressforshipping.districtcode,
        subdistrictcode: addressforshipping.subdistrictcode,
        zipcode: addressforshipping.zipcode,
        latitude: addressforshipping.latitude,
        longitude: addressforshipping.longitude,
        contactnames: [],
        phoneprimary: addressforshipping.phoneprimary,
        phonesecondary: addressforshipping.phonesecondary,
      );

      for (var lang in languageList) {
        newData.contactnames!.add(
          LanguageDataModel(code: lang.code!, name: ""),
        );
      }
      screenData.addressforshipping.add(newData);
    }

    //ใส่ value ตามภาษา
    for (var data in creditor.addressforbilling.contactnames!) {
      for (var ele in screenData.addressforbilling.contactnames!) {
        if (data.code == ele.code) {
          ele.name = data.name;
        }
      }
    }

    for (var data in creditor.names) {
      for (var ele in screenData.names) {
        if (data.code == ele.code) {
          ele.name = data.name;
        }
      }
    }

    // for (var addressforshipping in customer.addressforshipping) {
    //   for (var data in addressforshipping.contactnames) {
    //     for (var arr in screenData.addressforshipping) {
    //       for (var ele in arr.contactnames) {
    //         if (data.code == ele.code) {
    //           ele.name = data.name;
    //         }
    //       }
    //     }
    //   }
    // }

    for (int i = 0; i < creditor.addressforshipping.length; i++) {
      for (var data in creditor.addressforshipping[i].contactnames!) {
        for (var ele in screenData.addressforshipping[i].contactnames!) {
          if (data.code == ele.code) {
            ele.name = data.name;
          }
        }
      }
    }

    //เก็บค่าที่ไม่ได้เปิดใช้งานภาษาเข้าทาง array
    for (var defaultValueLang in creditor.names) {
      LanguageDataModel result = screenData.names.firstWhere(
        (data) => data.code == defaultValueLang.code,
        orElse: () => LanguageDataModel(code: '', name: ''),
      );
      if (result.code == '') {
        screenData.names.add(defaultValueLang);
      }
    }

    for (var defaultValueLang in creditor.addressforbilling.contactnames!) {
      LanguageDataModel result = screenData.addressforbilling.contactnames!
          .firstWhere(
            (data) => data.code == defaultValueLang.code,
            orElse: () => LanguageDataModel(code: '', name: ''),
          );
      if (result.code == '') {
        screenData.addressforbilling.contactnames!.add(defaultValueLang);
      }
    }

    for (var addressforshipping in creditor.addressforshipping) {
      for (var defaultValueLang in addressforshipping.contactnames!) {
        for (var arr in screenData.addressforshipping) {
          LanguageDataModel result = arr.contactnames!.firstWhere(
            (data) => data.code == defaultValueLang.code,
            orElse: () => LanguageDataModel(code: '', name: ''),
          );
          if (result.code == '') {
            arr.contactnames!.add(defaultValueLang);
          }
        }
      }
    }

    imageWeb = [];
    imageFile = [];

    for (int i = 0; i < creditor.images.length; i++) {
      imageWeb.add(Uint8List(0));
      imageFile.add(File(''));
    }
  }

  Future<List<CreditorGroupModel>> getDataGroup(filter) async {
    return listDataGroup;
  }

  CreditorGroupModel getGroup(String guid) {
    CreditorGroupModel data = CreditorGroupModel(guidfixed: "", groupcode: "");
    for (var ele in listDataGroup) {
      if (ele.guidfixed == guid) {
        data = ele;
      }
    }
    return data;
  }

  Widget editScreen({mobileScreen}) {
    List<Widget> formWidgets = [];

    focusNodeMax = 0;
    formWidgets.add(
      Padding(
        padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          enabled: screenData.code.isEmpty,
          onEditingComplete: () {
            if (kIsWeb) {
              findFocusNext(focusNodeIndex);
            }
          },
          focusNode: fieldFocusNodes[focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: creditorCode,
          textCapitalization: TextCapitalization.characters,
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.allow(RegExp('[a-z A-Z 0-9 -]')),
          ],
          onChanged: (value) {
            isDataChange = true;
            screenData.code = value.toUpperCase();
            creditorCode.value = TextEditingValue(
              text: value.toUpperCase(),
              selection: creditorCode.selection,
            );
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("creditor_code"),
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
    for (
      int languageIndex = 0;
      languageIndex < languageList.length;
      languageIndex++
    ) {
      LanguageDataModel creditorName = screenData.names.firstWhere(
        (element) => element.code == languageList[languageIndex].code,
        orElse: () => LanguageDataModel(code: '', name: ''),
      );
      if (creditorName.code == '') {
        screenData.names.add(
          LanguageDataModel(code: languageList[languageIndex].code!, name: ''),
        );
      }
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: TextFormField(
            readOnly: !isEditMode,
            onChanged: (value) {
              isDataChange = true;
              screenData.names[languageIndex].name = value;
            },
            onEditingComplete: () {
              if (kIsWeb) {
                findFocusNext(focusNodeIndex);
              }
            },
            focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
            textAlign: TextAlign.left,
            controller: TextEditingController(
              text: screenData.names[languageIndex].name,
            ),
            decoration: InputDecoration(
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: const OutlineInputBorder(),
              labelText:
                  "${global.language("creditor_name")} (${getLangName(screenData.names[languageIndex].code)})",
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
            validator: (value) {
              if (languageIndex == 0) {
                if (value == null || value.isEmpty) {
                  return global.language('this_field_is_required');
                }
              }

              return null;
            },
          ),
        ),
      );
    }
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Row(
          children: [
            Radio(
              focusNode: FocusNode(skipTraversal: true),
              value: 1,
              groupValue: screenData.personaltype,
              onChanged: (value) {
                setState(() {
                  screenData.personaltype = value!;
                  focusNodeIndex = 3;
                  refreshFocus = true;
                });
              },
            ),
            Text(global.language("customer_individual")),
            const SizedBox(width: 10),
            Radio(
              focusNode: FocusNode(skipTraversal: true),
              value: 2,
              groupValue: screenData.personaltype,
              onChanged: (value) {
                setState(() {
                  screenData.personaltype = value!;
                  focusNodeIndex = 3;
                  refreshFocus = true;
                });
              },
            ),
            Text(global.language("customer_company")),
          ],
        ),
      ),
    );
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          maxLength: 100,
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {
            findFocusNext(focusNodeIndex);
          },
          textInputAction: TextInputAction.next,
          focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: TextEditingController(text: screenData.taxid),
          textCapitalization: TextCapitalization.characters,
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.allow(RegExp('[a-z A-Z 0-9 -*/]')),
          ],
          onChanged: (value) {
            isDataChange = true;
            screenData.taxid = value;
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language(
              (screenData.personaltype == 1)
                  ? "customer_tax_id_card_number"
                  : "customer_tax_id_bussiness",
            ),
          ),
          validator: (value) {
            if (screenData.personaltype == 2) {
              if (value == null || value.isEmpty) {
                return global.language('this_field_is_required');
              }
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
            Radio(
              focusNode: FocusNode(skipTraversal: true),
              value: 1,
              groupValue: screenData.customertype,
              onChanged: (screenData.personaltype == 2)
                  ? (value) {
                      setState(() {
                        screenData.customertype = 1;
                        screenData.branchnumber = '00000';
                        focusNodeIndex = 5;
                        refreshFocus = true;
                      });
                    }
                  : null,
            ),
            Text(global.language("head_office")),
            const SizedBox(width: 10),
            Radio(
              focusNode: FocusNode(skipTraversal: true),
              value: 2,
              groupValue: screenData.customertype,
              onChanged: (screenData.personaltype == 2)
                  ? (value) {
                      setState(() {
                        screenData.customertype = 2;
                        screenData.branchnumber = '';
                        focusNodeIndex = 4;
                        refreshFocus = true;
                      });
                    }
                  : null,
            ),
            Text(global.language("branch")),
          ],
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          enabled: screenData.personaltype == 2,
          readOnly: !isEditMode || screenData.customertype == 1,
          onFieldSubmitted: (value) {
            findFocusNext(focusNodeIndex);
          },
          key: _popupBuilderKey2,
          textInputAction: TextInputAction.next,
          focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: TextEditingController(text: screenData.branchnumber),
          textCapitalization: TextCapitalization.characters,
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.allow(RegExp('[0-9]')),
          ],
          onChanged: (value) {
            isDataChange = true;
            screenData.branchnumber = value;
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("branch_number"),
          ),
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {
            bool isValid = (screenData.email.isEmpty)
                ? true
                : EmailValidator.validate(screenData.email);
            if (!isValid) {
              global.showSnackBar(
                context,
                Icon(Icons.delete, color: Colors.white),
                global.language("email_error"),
                Colors.red,
              );
            }
            findFocusNext(focusNodeIndex);
          },
          textInputAction: TextInputAction.next,
          focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: TextEditingController(text: screenData.email),
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isDataChange = true;
            screenData.email = value;
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("email"),
          ),
        ),
      ),
    );

    // formWidgets.add(Padding(
    // padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
    //     child: TextFormField(
    //         readOnly: !isEditMode,
    //         onFieldSubmitted: (value) {
    //           findFocusNext(focusNodeIndex);
    //         },
    //         focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
    //         controller: TextEditingController(text: screenData.fundcode),
    //         onChanged: (value) {
    //           isDataChange = true;
    //           screenData.fundcode = value;
    //         },
    //         decoration: InputDecoration(
    //           floatingLabelBehavior: FloatingLabelBehavior.always,
    // border: OutlineInputBorder(),
    //           labelText: global.language("fund_code"),
    //         ))));

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {
            findFocusNext(focusNodeIndex);
          },
          keyboardType: TextInputType.number,
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.digitsOnly,
          ],
          focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
          controller: TextEditingController(
            text: screenData.creditday.toString(),
          ),
          onChanged: (value) {
            isDataChange = true;
            screenData.creditday = int.parse(value);
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("credit_day"),
          ),
        ),
      ),
    );

    formWidgets.add(
      (!kIsWeb)
          ? Padding(
              padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
              child: DropdownSearch<CreditorGroupModel>.multiSelection(
                enabled: isEditMode,
                key: _popupBuilderKey,
                items: (String filter, infiniteScrollProps) =>
                    getDataGroup(filter),
                compareFn: (item, selectedItem) =>
                    item.guidfixed == selectedItem.guidfixed,
                itemAsString: (CreditorGroupModel? group) {
                  if (group == null) return '';
                  return '${group.groupcode} - ${global.activeLangName(group.names)}';
                },
                decoratorProps: DropDownDecoratorProps(
                  decoration: InputDecoration(
                    border: const OutlineInputBorder(),
                    contentPadding: const EdgeInsets.only(
                      left: 10,
                      top: 0,
                      bottom: 0,
                      right: 10,
                    ),
                    floatingLabelBehavior: FloatingLabelBehavior.always,
                    labelText: global.language("group"),
                  ),
                ),
                onChanged: (List<CreditorGroupModel> value) {
                  setState(() {
                    groupSelected = value;
                  });
                },
                popupProps: const PopupPropsMultiSelection.modalBottomSheet(
                  showSearchBox: true,
                  showSelectedItems: true,
                ),
                selectedItems: groupSelected,
              ),
            )
          : Padding(
              padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
              child: DropdownSearch<CreditorGroupModel>.multiSelection(
                enabled: isEditMode,
                compareFn: (item, selectedItem) =>
                    item.guidfixed == selectedItem.guidfixed,
                items: (String filter, infiniteScrollProps) =>
                    getDataGroup(filter),
                itemAsString: (CreditorGroupModel? group) {
                  if (group == null) return '';
                  return '${group.groupcode} - ${global.activeLangName(group.names)}';
                },
                decoratorProps: DropDownDecoratorProps(
                  decoration: InputDecoration(
                    border: const OutlineInputBorder(),
                    contentPadding: const EdgeInsets.only(
                      left: 10,
                      top: 0,
                      bottom: 0,
                      right: 10,
                    ),
                    floatingLabelBehavior: FloatingLabelBehavior.always,
                    labelText: global.language("group"),
                  ),
                ),
                onChanged: (List<CreditorGroupModel> value) {
                  setState(() {
                    groupSelected = value;
                  });
                },
                popupProps: const PopupPropsMultiSelection.dialog(
                  showSearchBox: true,
                  showSelectedItems: true,
                ),
                selectedItems: groupSelected,
              ),
            ),
    );

    /// add divider
    formWidgets.add(
      const Padding(
        padding: EdgeInsets.all(8.0),
        child: Divider(height: 1, thickness: 1),
      ),
    );

    /// add title "ข้อมูลผู้ติดต่อ"
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Text(
          global.language("contact_information"),
          style: const TextStyle(fontWeight: FontWeight.bold),
        ),
      ),
    );

    for (
      int languageIndex = 0;
      languageIndex < languageList.length;
      languageIndex++
    ) {
      LanguageDataModel contactnames = screenData
          .addressforbilling
          .contactnames!
          .firstWhere(
            (element) => element.code == languageList[languageIndex].code,
            orElse: () => LanguageDataModel(code: '', name: ''),
          );
      if (contactnames.code == '') {
        screenData.addressforbilling.contactnames!.add(
          LanguageDataModel(code: languageList[languageIndex].code!, name: ''),
        );
      }
      formWidgets.add(
        Padding(
          padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: TextField(
            readOnly: !isEditMode,
            onChanged: (value) {
              isDataChange = true;
              screenData.addressforbilling.contactnames![languageIndex].name =
                  value;
            },
            onSubmitted: (value) {
              if (kIsWeb) {
                findFocusNext(focusNodeIndex);
              }
            },
            focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
            textAlign: TextAlign.left,
            controller: TextEditingController(
              text: screenData
                  .addressforbilling
                  .contactnames![languageIndex]
                  .name,
            ),
            decoration: InputDecoration(
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: const OutlineInputBorder(),
              labelText:
                  "${global.language("creditor_name")} (${getLangName(screenData.addressforbilling.contactnames![languageIndex].code)})",
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

    // Ensure address list has enough elements
    screenData.addressforbilling.address ??= [];
    while (screenData.addressforbilling.address!.length < 3) {
      screenData.addressforbilling.address!.add('');
    }

    for (int addressIndex = 0; addressIndex < 3; addressIndex++) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: TextFormField(
            readOnly: !isEditMode,
            onChanged: (value) {
              isDataChange = true;
              screenData.addressforbilling.address![addressIndex] = value;
            },
            onFieldSubmitted: (value) {
              findFocusNext(focusNodeIndex);
            },
            textInputAction: TextInputAction.next,
            focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
            textAlign: TextAlign.left,
            controller: TextEditingController(
              text: screenData.addressforbilling.address![addressIndex],
            ),
            decoration: InputDecoration(
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: OutlineInputBorder(),
              labelText: global.language("tax_address_${addressIndex + 1}"),
            ),
          ),
        ),
      );
    }

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          maxLength: 10,
          readOnly: !isEditMode,
          onChanged: (value) {
            isDataChange = true;
            screenData.addressforbilling.phoneprimary = value;
          },
          onFieldSubmitted: (value) {
            findFocusNext(focusNodeIndex);
          },
          textInputAction: TextInputAction.next,
          focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: TextEditingController(
            text: screenData.addressforbilling.phoneprimary,
          ),
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.allow(RegExp('[0-9]')),
          ],
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("telephone_primary"),
          ),
        ),
      ),
    );
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          maxLength: 10,
          readOnly: !isEditMode,
          onChanged: (value) {
            isDataChange = true;
            screenData.addressforbilling.phonesecondary = value;
          },
          onFieldSubmitted: (value) {
            findFocusNext(focusNodeIndex);
          },
          textInputAction: TextInputAction.next,
          focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: TextEditingController(
            text: screenData.addressforbilling.phonesecondary,
          ),
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.allow(RegExp('[0-9]')),
          ],
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("telephone_secondary"),
          ),
        ),
      ),
    );

    for (
      int addressShipIndex = 0;
      addressShipIndex < screenData.addressforshipping.length;
      addressShipIndex++
    ) {
      formWidgets.add(
        (screenData.addressforshipping.isNotEmpty)
            ? Padding(
                padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
                child: Text(
                  global.language(
                    "address_for_shipping ${addressShipIndex + 1}",
                  ),
                ),
              )
            : Container(),
      );
      for (
        int languageIndex = 0;
        languageIndex < languageList.length;
        languageIndex++
      ) {
        LanguageDataModel contactnames = screenData
            .addressforshipping[addressShipIndex]
            .contactnames!
            .firstWhere(
              (element) => element.code == languageList[languageIndex].code,
              orElse: () => LanguageDataModel(code: '', name: ''),
            );
        if (contactnames.code == '') {
          screenData.addressforshipping[addressShipIndex].contactnames!.add(
            LanguageDataModel(
              code: languageList[languageIndex].code!,
              name: '',
            ),
          );
        }
        formWidgets.add(
          Padding(
            padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
            child: TextField(
              readOnly: !isEditMode,
              onChanged: (value) {
                isDataChange = true;
                screenData
                        .addressforshipping[addressShipIndex]
                        .contactnames![languageIndex]
                        .name =
                    value;
              },
              onSubmitted: (value) {
                if (kIsWeb) {
                  findFocusNext(focusNodeIndex);
                }
              },
              focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
              textAlign: TextAlign.left,
              controller: TextEditingController(
                text: screenData
                    .addressforshipping[addressShipIndex]
                    .contactnames![languageIndex]
                    .name,
              ),
              decoration: InputDecoration(
                floatingLabelBehavior: FloatingLabelBehavior.always,
                border: const OutlineInputBorder(),
                labelText:
                    "${global.language("creditor_name")} (${getLangName(screenData.addressforshipping[addressShipIndex].contactnames![languageIndex].code)})",
              ),
            ),
          ),
        );
      }
      for (
        int addressIndex = 0;
        addressIndex <
            screenData.addressforshipping[addressShipIndex].address!.length;
        addressIndex++
      ) {
        formWidgets.add(
          Padding(
            padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
            child: TextFormField(
              readOnly: !isEditMode,
              onChanged: (value) {
                isDataChange = true;
                screenData
                        .addressforshipping[addressShipIndex]
                        .address![addressIndex] =
                    value;
              },
              onFieldSubmitted: (value) {
                findFocusNext(focusNodeIndex);
              },
              textInputAction: TextInputAction.next,
              focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
              textAlign: TextAlign.left,
              controller: TextEditingController(
                text: screenData
                    .addressforshipping[addressShipIndex]
                    .address![addressIndex],
              ),
              decoration: InputDecoration(
                floatingLabelBehavior: FloatingLabelBehavior.always,
                border: OutlineInputBorder(),
                labelText: global.language(
                  "shipping_address_${addressIndex + 1}",
                ),
              ),
            ),
          ),
        );
      }
      formWidgets.add(
        Padding(
          padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: TextFormField(
            maxLength: 10,
            readOnly: !isEditMode,
            onChanged: (value) {
              isDataChange = true;
              screenData.addressforshipping[addressShipIndex].phoneprimary =
                  value;
            },
            onFieldSubmitted: (value) {
              findFocusNext(focusNodeIndex);
            },
            textInputAction: TextInputAction.next,
            focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
            textAlign: TextAlign.left,
            controller: TextEditingController(
              text:
                  screenData.addressforshipping[addressShipIndex].phoneprimary,
            ),
            inputFormatters: <TextInputFormatter>[
              FilteringTextInputFormatter.allow(RegExp('[0-9]')),
            ],
            decoration: InputDecoration(
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: OutlineInputBorder(),
              labelText: global.language("telephone_primary"),
            ),
          ),
        ),
      );
      formWidgets.add(
        Padding(
          padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: TextFormField(
            maxLength: 10,
            readOnly: !isEditMode,
            onChanged: (value) {
              isDataChange = true;
              screenData.addressforshipping[addressShipIndex].phonesecondary =
                  value;
            },
            onFieldSubmitted: (value) {
              findFocusNext(focusNodeIndex);
            },
            textInputAction: TextInputAction.next,
            focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
            textAlign: TextAlign.left,
            controller: TextEditingController(
              text: screenData
                  .addressforshipping[addressShipIndex]
                  .phonesecondary,
            ),
            inputFormatters: <TextInputFormatter>[
              FilteringTextInputFormatter.allow(RegExp('[0-9]')),
            ],
            decoration: InputDecoration(
              floatingLabelBehavior: FloatingLabelBehavior.always,
              border: OutlineInputBorder(),
              labelText: global.language("telephone_secondary"),
            ),
          ),
        ),
      );
    }

    formWidgets.add(
      (screenData.images.isNotEmpty)
          ? Padding(
              padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
              child: Text(global.language("customer_image")),
            )
          : Container(),
    );

    List<Widget> imageList = [];
    for (
      int imageIndex = 0;
      imageIndex < screenData.images.length;
      imageIndex++
    ) {
      imageList.add(
        Container(
          width: 300,
          padding: EdgeInsets.only(left: 5, right: 5, bottom: 5, top: 5),
          decoration: BoxDecoration(
            border: Border.all(color: global.theme.textSecondaryColor),
            borderRadius: const BorderRadius.all(Radius.circular(5.0)),
          ),
          child: Column(
            children: [
              Row(
                children: [
                  (isEditMode)
                      ? Expanded(
                          child: IconButton(
                            focusNode: FocusNode(skipTraversal: true),
                            onPressed: () async {
                              screenData.images.removeAt(imageIndex);
                              imageWeb.removeAt(imageIndex);
                              imageFile.removeAt(imageIndex);
                              setState(() {});
                            },
                            icon: const Icon(Icons.delete),
                          ),
                        )
                      : Container(),
                  const SizedBox(width: 5),
                  (isEditMode)
                      ? Expanded(
                          child: IconButton(
                            focusNode: FocusNode(skipTraversal: true),
                            onPressed: (kIsWeb)
                                ? () async {
                                    XFile? image = await imagePicker.pickImage(
                                      source: ImageSource.gallery,
                                      maxHeight: 480,
                                      maxWidth: 640,
                                      imageQuality: 60,
                                    );
                                    if (image != null) {
                                      var f = await image.readAsBytes();
                                      setState(() {
                                        imageWeb[imageIndex] = f;
                                        imageFile[imageIndex] = File(
                                          image.path,
                                        );
                                        FocusScope.of(context).unfocus();
                                      });
                                    }
                                  }
                                : () async {
                                    final XFile? photo = await imagePicker
                                        .pickImage(
                                          source: ImageSource.camera,
                                          maxHeight: 480,
                                          maxWidth: 640,
                                          imageQuality: 60,
                                        );
                                    if (photo != null) {
                                      var f = await photo.readAsBytes();
                                      imageWeb[imageIndex] = f;
                                      imageFile.add(File(photo.path));
                                      setState(() {
                                        FocusScope.of(context).unfocus();
                                      });
                                    }
                                  },
                            icon: const Icon(Icons.folder),
                          ),
                        )
                      : Container(),
                  const SizedBox(width: 5),
                  if (kIsWeb == false)
                    Expanded(
                      child: IconButton(
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
                            imageWeb[imageIndex] = f;
                            imageFile.add(File(photo.path));
                            setState(() {});
                          }
                        },
                        icon: const Icon(Icons.camera_alt),
                      ),
                    ),
                ],
              ),
              SizedBox(
                width: 300,
                height: 300,
                child: Stack(
                  children: [
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
                          imageWeb[imageIndex] = bytes;
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
                          boxShadow: const [
                            BoxShadow(
                              offset: Offset(0, 4),
                              color: Colors.cyan, //edited
                              spreadRadius: 4,
                              blurRadius: 10, //edited
                            ),
                          ],
                          image: (imageWeb[imageIndex].isNotEmpty)
                              ? DecorationImage(
                                  image: MemoryImage(imageWeb[imageIndex]),
                                  fit: BoxFit.fill,
                                )
                              : (screenData.images[imageIndex].uri != '')
                              ? DecorationImage(
                                  image: NetworkImage(
                                    screenData.images[imageIndex].uri,
                                  ),
                                  fit: BoxFit.fill,
                                )
                              : const DecorationImage(
                                  image: AssetImage('assets/img/noimage.png'),
                                ),
                        ),
                        child: const SizedBox(width: 500, height: 500),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      );
    }

    formWidgets.add(Wrap(children: imageList));
    formWidgets.add(const SizedBox(height: 5));
    if (isEditMode) {
      formWidgets.add(
        Container(
          width: double.infinity,
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: ElevatedButton.icon(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () {
              setState(() {
                screenData.images.add(ImagesModel(uri: '', xorder: 0));
                imageWeb.add(Uint8List(0));
                imageFile.add(File(''));
                FocusScope.of(context).unfocus();
              });
            },
            icon: Icon(Icons.add),
            label: Text(global.language("add_picture")),
          ),
        ),
      );
    }
    // formWidgets.add(const SizedBox(
    //   height: 5,
    // ));
    // if (isEditMode) {
    //   formWidgets.add(Container(
    //     width: double.infinity,
    //     padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
    //     child: ElevatedButton.icon(
    //         focusNode: FocusNode(skipTraversal: true),
    //         onPressed: () {
    //           setState(() {
    //             List<LanguageDataModel> names = [];
    //             for (int k = 0; k < languageList.length; k++) {
    //               names.add(LanguageDataModel(code: languageList[k].code!, name: ""));
    //             }
    //             screenData.addressforshipping.add(CustomerAddressModel(
    //                 guid: "",
    //                 address: ["", "", ""],
    //                 countrycode: "",
    //                 provincecode: "",
    //                 districtcode: "",
    //                 subdistrictcode: "",
    //                 zipcode: "",
    //                 latitude: 0,
    //                 longitude: 0,
    //                 phoneprimary: "",
    //                 phonesecondary: "",
    //                 contactnames: names));
    //           });
    //         },
    // icon: Icon(Icons.add),
    //         label: Text(global.language("add_address_shipping"))),
    //   ));
    // }

    formWidgets.add(const SizedBox(height: 5));

    if (isSaveAllow) {
      formWidgets.add(
        Container(
          width: double.infinity,
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
        title: Text(headerEdit + global.language("creditor")),
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
                            context.read<CreditorBloc>().add(
                              CreditorDelete(guid: selectGuid),
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
                  setState(() {
                    isLoadTranslation = true;
                  });

                  /// Translate names in global
                  for (var name in screenData.names) {
                    if (name.name.trim().isEmpty) {
                      var data = await global.translateNames(
                        namesData: screenData.names,
                      );
                      setState(() {
                        screenData.names = data;
                      });
                      break;
                    }
                  }

                  for (var name in screenData.addressforbilling.contactnames!) {
                    if (name.name.trim().isEmpty) {
                      var data = await global.translateNames(
                        namesData: screenData.addressforbilling.contactnames!,
                      );
                      setState(() {
                        screenData.addressforbilling.contactnames = data;
                      });
                      break;
                    }
                  }
                  setState(() {
                    isLoadTranslation = false;
                  });
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
          const ManualButton(path: 'creditor'),
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
              BlocListener<CreditorBloc, CreditorState>(
                listener: (context, state) {
                  blocCurrentState = state;
                  // Load
                  if (state is CreditorLoadSuccess) {
                    setState(() {
                      loadingData = false;
                      if (state.creditors.isNotEmpty) {
                        listData.addAll(state.creditors);
                        // สร้าง keys ใหม่ให้ตรงกับจำนวน items
                        listKeys.clear();
                        for (int i = 0; i < listData.length; i++) {
                          listKeys.add(GlobalKey());
                        }
                      }
                    });
                  }
                  if (state is CreditorLoadFailed) {
                    setState(() {
                      loadingData = false;
                      global.showSnackBar(
                        context,
                        Icon(Icons.error, color: Colors.white),
                        "${global.language("not_success_load")} : ${state.message}",
                        Colors.red,
                      );
                    });
                  }
                  // Save
                  if (state is CreditorSaveSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.save, color: Colors.white),
                        global.language("save_success"),
                        Colors.blue,
                      );
                      clearEditData();
                      listData.clear();
                      loadDataList(searchText, selectedFilterCodes);
                    });
                  }
                  if (state is CreditorSaveFailed) {
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
                  if (state is CreditorUpdateSuccess) {
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
                      loadDataList(searchText, selectedFilterCodes);
                      isSaveAllow = false;
                      getData(selectGuid);
                    });
                  }
                  if (state is CreditorUpdateFailed) {
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
                  if (state is CreditorDeleteSuccess) {
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
                      loadDataList(searchText, selectedFilterCodes);
                    });
                  }
                  // Delete Many
                  if (state is CreditorDeleteManySuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.delete, color: Colors.white),
                        global.language("not_delete_success"),
                        Colors.blue,
                      );
                      listData.clear();
                      clearEditData();
                      loadDataList(searchText, selectedFilterCodes);
                      showCheckBox = false;
                    });
                  }
                  // Get
                  if (state is CreditorGetSuccess) {
                    setState(() {
                      getDataToEditScreen(state.creditors);
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
                  // Bulk Save
                  if (state is CreditorBulkSaveSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        const Icon(Icons.upload_file, color: Colors.white),
                        'Import สำเร็จ ${state.savedCount} รายการ',
                        Colors.green,
                      );
                      listData.clear();
                      loadDataList(searchText, selectedFilterCodes);
                    });
                  }
                  if (state is CreditorBulkSaveFailed) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        const Icon(Icons.upload_file, color: Colors.white),
                        'Import ไม่สำเร็จ: ${state.message}',
                        Colors.red,
                      );
                    });
                  }
                },
              ),
              BlocListener<CreditorGroupBloc, CreditorGroupState>(
                listener: (context, state) {
                  // Load
                  if (state is CreditorGroupLoadSuccess) {
                    setState(() {
                      if (state.creditorGroups.isNotEmpty) {
                        listDataGroup = state.creditorGroups;
                      }
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

  Future<List<CreditorGroupModel>> filterCrediterGroup(
    List<CreditorGroupModel> selectedFilters,
  ) async {
    List<CreditorGroupModel> selectedValues = selectedFilters;

    await showDialog(
      context: context,
      builder: (BuildContext context) {
        return StatefulBuilder(
          builder: (BuildContext context, StateSetter setState) {
            return AlertDialog(
              title: Column(
                children: [
                  Text(global.language("filter_credit_group")),
                  const Divider(),
                  Wrap(
                    spacing: 8.0,
                    children: selectedValues.map((filter) {
                      return InputChip(
                        label: Text(filter.groupcode),
                        deleteIcon: const Icon(Icons.close),
                        onDeleted: () {
                          setState(() {
                            selectedValues.remove(filter);
                          });
                        },
                      );
                    }).toList(),
                  ),
                  const Divider(),
                ],
              ),
              content: SizedBox(
                width: double.maxFinite,
                child: ListView.builder(
                  itemCount: listDataGroup.length,
                  itemBuilder: (BuildContext context, int index) {
                    final filter = listDataGroup[index];
                    final isSelected = selectedValues.contains(filter);
                    return CheckboxListTile(
                      title: Text(
                        "${filter.groupcode} ~ ${(global.packName(filter.names))} ",
                      ),
                      value: isSelected,
                      onChanged: (bool? value) {
                        setState(() {
                          if (value == true) {
                            selectedValues.add(filter);
                          } else {
                            selectedValues.remove(filter);
                          }
                        });
                      },
                    );
                  },
                ),
              ),
              actions: [
                ElevatedButton(
                  onPressed: () {
                    Navigator.pop(context);
                  },
                  child: Text(global.language("filter")),
                ),
                ElevatedButton(
                  onPressed: () {
                    setState(() {
                      Navigator.pop(context);
                      selectedValues.clear();
                    });
                  },
                  child: Text(global.language("cancel")),
                ),
              ],
            );
          },
        );
      },
    );

    return selectedValues;
  }
}
