// หน้าจัดการข้อมูลลูกหนี้

import 'dart:async';
import 'dart:io';

import 'package:smlaicloud/bloc/debtor/debtor_bloc.dart';
import 'package:smlaicloud/bloc/debtor_group/debtor_group_bloc.dart';
import 'package:smlaicloud/model/customer_address_model.dart';
import 'package:smlaicloud/model/debtor_group_model.dart';
import 'package:smlaicloud/model/debtor_model.dart';
import 'package:smlaicloud/model/point_transaction_model.dart';
import 'package:smlaicloud/model/bulk_points_request_model.dart';
import 'package:smlaicloud/model/price_model.dart';
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
import 'package:smlaicloud/screens/config/point_transaction_screen.dart';
import 'package:smlaicloud/bloc/point_transaction/point_transaction_bloc.dart';
import 'package:smlaicloud/repositories/point_transaction_repository.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/components/loading_overlay.dart';
import 'package:smlaicloud/widgets/manual_button.dart';

class DebtorScreen extends StatefulWidget {
  const DebtorScreen({super.key});

  @override
  State<DebtorScreen> createState() => DebtorScreenState();
}

class DebtorScreenState extends State<DebtorScreen>
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
  List<DebtorModel> listData = [];
  List<DebtorGroupModel> listDataGroup = [];
  List<String> guidListChecked = [];
  List<LanguageDataModel> names = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  bool isDataChange = false;
  bool isSaveAllow = false;
  late DebtorState blocCurrentState;
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  int _hoverIndex = -1;
  bool isEditMode = false;
  late DebtorModel screenData;
  List<Uint8List> imageWeb = [];
  final ImagePicker imagePicker = ImagePicker();
  late DropzoneViewController dropZoneController;
  Color colorSelected = Colors.white;
  final _debouncer = global.Debouncer(1000);
  late Timer screenTimer;
  bool loadingData = false;
  List<DebtorGroupModel> groupSelected = [];
  List<File> imageFile = [];
  global.ScreenEventEnum screenEvent = global.ScreenEventEnum.list;
  late SplitViewController splitViewController;
  final _popupBuilderKey = GlobalKey<DropdownSearchState<String>>();
  final _popupBuilderKey2 = GlobalKey<DropdownSearchState<String>>();
  TextEditingController debtorCode = TextEditingController();

  List<DebtorGroupModel> selectedFilters = [];
  List<String> selectedFilterCodes = [];

  bool isLoadTranslation = false;

  final _formKey = GlobalKey<FormState>();
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

    // แก้ไข: ใช้ milliseconds แทน microseconds (เปลี่ยนจาก 2000 ครั้ง/วินาที → 2 ครั้ง/วินาที)
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
    screenTimer.cancel(); // ✅ เพิ่ม: cancel timer ก่อน dispose
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

  void loadDataGroupList() {
    context.read<DebtorGroupBloc>().add(
      const DebtorGroupLoadList(offset: 0, limit: 1000, search: ""),
    );
  }

  void loadDataList(String search, List<String>? filter) {
    // ป้องกัน infinite loop - ถ้ากำลัง loading อยู่แล้ว ไม่ต้องโหลดซ้ำ
    if (loadingData) {
      if (kDebugMode) {
        AppLogger.debug('⚠️ loadDataList called but already loading - SKIPPED');
      }
      return;
    }

    if (kDebugMode) {
      AppLogger.debug(
        '📋 loadDataList called - offset: ${listData.isEmpty ? 0 : listData.length}, search: "$search"',
      );
    }
    setState(() {
      loadingData = true;
    });
    searchText = search;
    context.read<DebtorBloc>().add(
      DebtorLoadList(
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
    debtorCode.text = "";
    screenData = DebtorModel(
      guidfixed: "",
      code: "",
      names: names,
      customertype: 1,
      branchnumber: '00000',
      personaltype: 1,
      addressforbilling: CustomerAddressModel(
        guid: "",
        address: [""],
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
      pointbalance: 0,
      pointscode: "",
      pricelevel: "1",
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
    context.read<DebtorBloc>().add(DebtorGet(guid: guid));
  }

  // Dialog สำหรับเลือกประเภทการนำเข้า
  Future<void> _showImportTypeDialog() async {
    await showDialog(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          title: Row(
            children: [
              Icon(Icons.upload_file, color: global.theme.primaryColor),
              SizedBox(width: 10),
              Text(global.language('import_from_excel')),
            ],
          ),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                global.language('select_import_type'),
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
              ),
              const SizedBox(height: 20),

              // ตัวเลือกที่ 1: นำเข้าข้อมูลลูกหนี้
              ListTile(
                leading: Icon(Icons.people, color: global.theme.primaryColor, size: 32),
                title: Text(
                  global.language('import_debtor_data'),
                  style: TextStyle(fontWeight: FontWeight.bold),
                ),
                subtitle: Text(global.language('import_debtor_list_data')),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(8),
                  side: BorderSide(color: Colors.grey.shade300),
                ),
                onTap: () {
                  Navigator.of(context).pop();
                  _importDebtorFromExcel();
                },
              ),

              const SizedBox(height: 12),

              // ตัวเลือกที่ 2: นำเข้าแต้มลูกหนี้
              ListTile(
                leading: const Icon(
                  Icons.stars,
                  color: Colors.orange,
                  size: 32,
                ),
                title: Text(
                  global.language('import_debtor_points'),
                  style: TextStyle(fontWeight: FontWeight.bold),
                ),
                subtitle: Text(global.language('import_debtor_points_data')),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(8),
                  side: BorderSide(color: Colors.grey.shade300),
                ),
                onTap: () {
                  Navigator.of(context).pop();
                  _importPointsFromExcel();
                },
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(context).pop(),
              child: Text(global.language('cancel')),
            ),
          ],
        );
      },
    );
  }

  // Dialog สำหรับเลือกประเภท template ที่จะดาวน์โหลด
  Future<void> _showDownloadTemplateDialog() async {
    await showDialog(
      context: context,
      builder: (BuildContext context) {
        return AlertDialog(
          title: Row(
            children: [
              const Icon(Icons.download, color: Colors.green),
              SizedBox(width: 10),
              Text(global.language('download_template')),
            ],
          ),
          content: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Text(
                global.language('select_template_to_download'),
                style: TextStyle(fontSize: 14, fontWeight: FontWeight.w500),
              ),
              const SizedBox(height: 20),

              // ตัวเลือกที่ 1: Template ลูกหนี้
              ListTile(
                leading: Icon(Icons.people, color: global.theme.primaryColor, size: 32),
                title: Text(
                  global.language('debtor_data_template'),
                  style: TextStyle(fontWeight: FontWeight.bold),
                ),
                subtitle: Text(global.language('sample_file_import_debtor')),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(8),
                  side: BorderSide(color: Colors.grey.shade300),
                ),
                onTap: () {
                  Navigator.of(context).pop();
                  _downloadDebtorTemplate();
                },
              ),

              const SizedBox(height: 12),

              // ตัวเลือกที่ 2: Template แต้มลูกหนี้
              ListTile(
                leading: const Icon(
                  Icons.stars,
                  color: Colors.orange,
                  size: 32,
                ),
                title: Text(
                  global.language('debtor_points_template'),
                  style: TextStyle(fontWeight: FontWeight.bold),
                ),
                subtitle: Text(global.language('sample_file_import_points')),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(8),
                  side: BorderSide(color: Colors.grey.shade300),
                ),
                onTap: () {
                  Navigator.of(context).pop();
                  _downloadPointsTemplate();
                },
              ),
            ],
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(context).pop(),
              child: Text(global.language('cancel')),
            ),
          ],
        );
      },
    );
  }

  // Dialog สำหรับแสดง Loading ระหว่างการนำเข้าข้อมูล
  Future<void> _showImportLoadingDialog({
    required BuildContext context,
    required String importType,
    required int itemCount,
    required Function onImport,
  }) async {
    String title = importType == 'debtor' ? global.language('debtor_data') : global.language('debtor_points');

    // เรียกฟังก์ชันนำเข้าก่อน
    onImport();

    // แสดง dialog หลังจากเรียกฟังก์ชันแล้ว
    await showDialog(
      context: context,
      barrierDismissible: false,
      builder: (BuildContext dialogContext) {
        return BlocListener<DebtorBloc, DebtorState>(
          listener: (context, state) {
            // สำหรับนำเข้าลูกหนี้
            if (importType == 'debtor') {
              if (state is DebtorBulkSaveSuccess) {
                Navigator.of(dialogContext).pop();
              } else if (state is DebtorBulkSaveFailed) {
                Navigator.of(dialogContext).pop();
              }
            }
            // สำหรับนำเข้าแต้ม
            else if (importType == 'points') {
              if (state is DebtorBulkAddPointsSuccess) {
                Navigator.of(dialogContext).pop();
              } else if (state is DebtorBulkAddPointsFailed) {
                Navigator.of(dialogContext).pop();
              }
            }
          },
          child: Dialog(
            backgroundColor: Colors.transparent,
            elevation: 0,
            child: Card(
              child: Padding(
                padding: const EdgeInsets.all(32),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    const SizedBox(
                      width: 48,
                      height: 48,
                      child: CircularProgressIndicator(
                        strokeWidth: 3,
                        valueColor: AlwaysStoppedAnimation<Color>(Colors.blue),
                        backgroundColor: Color(0x1A2196F3),
                      ),
                    ),
                    const SizedBox(height: 24),
                    Text(
                      '${global.language("importing")} $title...',
                      style: const TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w500,
                      ),
                      textAlign: TextAlign.center,
                    ),
                    const SizedBox(height: 8),
                    Text(
                      '${global.language("processing")} $itemCount ${global.language("items")}',
                      style: TextStyle(
                        color: Colors.black87.withValues(alpha: 0.6),
                        fontSize: 13,
                        fontWeight: FontWeight.w400,
                      ),
                      textAlign: TextAlign.center,
                    ),
                  ],
                ),
              ),
            ),
          ),
        );
      },
    );
  }

  // ฟังก์ชันสำหรับนำเข้าข้อมูลลูกหนี้ (เปลี่ยนชื่อจาก _importFromExcel)
  Future<void> _importDebtorFromExcel() async {
    try {
      FilePickerResult? result = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: ['xlsx', 'xls'],
      );

      if (result != null && result.files.single.bytes != null) {
        var bytes = result.files.single.bytes!;
        var excel = xl.Excel.decodeBytes(bytes);

        List<DebtorRequestModel> importedData = [];

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

                DebtorRequestModel? debtor = _parseExcelRowToDebtor(rowData);
                if (debtor != null) {
                  importedData.add(debtor);
                }
              }
            }
          }
        }

        if (importedData.isNotEmpty) {
          // Debug: แสดงข้อมูลที่จะส่งไป bulk save
          if (kDebugMode) {
            AppLogger.debug(
              '📤 Sending ${importedData.length} records to bulk save:',
            );
            for (int i = 0; i < importedData.length; i++) {
              AppLogger.debug(
                '   [$i] code=${importedData[i].code}, pointscode=${importedData[i].pointscode}, groups=${importedData[i].groups}',
              );
            }
          }

          // แสดง Loading Dialog พร้อมบันทึกข้อมูล
          await _showImportLoadingDialog(
            context: context,
            importType: 'debtor',
            itemCount: importedData.length,
            onImport: () => _saveBulkDebtors(importedData),
          );
        } else {
          global.showWarningSnackBar(context, global.language('no_valid_data_in_excel'));
        }
      }
    } catch (e) {
      global.showErrorSnackBar(context, 'เกิดข้อผิดพลาดในการ Import: $e');
    }
  }

  // ฟังก์ชันแปลง Excel row เป็น DebtorRequestModel
  DebtorRequestModel? _parseExcelRowToDebtor(List<xl.Data?> rowData) {
    try {
      // Expected columns: code, names, personaltype, taxid, customertype, branchnumber, email, creditday, address, phonepreimary, ismember, pricelevel, pointscode, groupcode
      if (rowData.length < 14) return null;

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
      int ismemberValue =
          int.tryParse(rowData[10]?.value?.toString() ?? '0') ?? 0;
      int pricelevelValue =
          int.tryParse(rowData[11]?.value?.toString() ?? '0') ?? 0;
      String pointscode = rowData[12]?.value?.toString() ?? '';
      String groupcode = rowData[13]?.value?.toString() ?? '';

      if (kDebugMode) {
        AppLogger.debug(
          '📝 Excel Row: code=$code, pointscode=$pointscode, groupcode=$groupcode',
        );
      }

      if (code.isEmpty || names.isEmpty) return null;

      // เงื่อนไขพิเศษสำหรับ ismember
      bool ismember = ismemberValue == 1;

      // เงื่อนไขพิเศษสำหรับ pricelevel
      // pricelevel 0 = keyNumber 1 (ราคาปลีก)
      // pricelevel 1-9 = keyNumber 9-17 (ราคาลู่ 1-9)
      String pricelevel;
      switch (pricelevelValue) {
        case 0:
          pricelevel = "1"; // ราคาปลีก
          break;
        case 1:
          pricelevel = "9"; // ราคาลู่ 1
          break;
        case 2:
          pricelevel = "10"; // ราคาลู่ 2
          break;
        case 3:
          pricelevel = "11"; // ราคาลู่ 3
          break;
        case 4:
          pricelevel = "12"; // ราคาลู่ 4
          break;
        case 5:
          pricelevel = "13"; // ราคาลู่ 5
          break;
        case 6:
          pricelevel = "14"; // ราคาลู่ 6
          break;
        case 7:
          pricelevel = "15"; // ราคาลู่ 7
          break;
        case 8:
          pricelevel = "16"; // ราคาลู่ 8
          break;
        case 9:
          pricelevel = "17"; // ราคาลู่ 9
          break;
        default:
          pricelevel = "1"; // ราคาปลีก (default)
          break;
      }

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

      // ประมวลผล groupcode และค้นหาใน listDataGroup
      List<String> groups = [];
      if (groupcode.isNotEmpty) {
        // แยก groupcode ด้วย comma (เช่น "111,222,333")
        List<String> groupCodes = groupcode
            .split(',')
            .map((e) => e.trim())
            .where((e) => e.isNotEmpty)
            .toList();

        if (kDebugMode) {
          AppLogger.debug('🔍 Searching for group codes: $groupCodes');
        }

        // ค้นหาแต่ละ groupcode ใน listDataGroup
        for (String code in groupCodes) {
          try {
            DebtorGroupModel? foundGroup = listDataGroup.firstWhere(
              (group) => group.groupcode == code,
              orElse: () => DebtorGroupModel(guidfixed: "", groupcode: ""),
            );

            // เพิ่มเฉพาะ guidfixed ของ group ที่เจอจริง
            if (foundGroup.guidfixed.isNotEmpty) {
              groups.add(foundGroup.guidfixed);
              if (kDebugMode) {
                AppLogger.debug(
                  '✅ พบ group code: $code → guid: ${foundGroup.guidfixed}',
                );
              }
            } else {
              if (kDebugMode) {
                AppLogger.debug('⚠️ ไม่พบ group code: $code');
              }
            }
          } catch (e) {
            if (kDebugMode) {
              AppLogger.error('⚠️ Error finding group code $code: $e');
            }
          }
        }
      }

      return DebtorRequestModel(
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
        groups: groups, // ใช้ groups ที่ค้นหาได้จาก listDataGroup
        images: [], // ใช้ค่าเริ่มต้น
        guidfixed: "", // จะถูกกำหนดในการบันทึก
        ismember: ismember,
        pricelevel: pricelevel,
        fundcode: "", // ใช้ค่าเริ่มต้น
        pointscode: pointscode, // ใช้ pointscode จาก Excel
      );
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error parsing row data: $e');
      }
      return null;
    }
  }

  // ฟังก์ชันสำหรับบันทึกข้อมูลลูกหนี้หลายรายการ
  Future<void> _saveBulkDebtors(List<DebtorRequestModel> debtors) async {
    try {
      // ใช้ bulk save method ใหม่
      context.read<DebtorBloc>().add(DebtorBulkSave(debtors: debtors));
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error saving bulk debtors: $e');
      }
      rethrow;
    }
  }

  // ฟังก์ชันสำหรับนำเข้าแต้มลูกหนี้
  Future<void> _importPointsFromExcel() async {
    try {
      FilePickerResult? result = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: ['xlsx', 'xls'],
      );

      if (result != null && result.files.single.bytes != null) {
        var bytes = result.files.single.bytes!;
        var excel = xl.Excel.decodeBytes(bytes);

        List<BulkPointsRequestModel> importedPoints = [];

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

                BulkPointsRequestModel? pointData = _parseExcelRowToPoints(
                  rowData,
                );
                if (pointData != null) {
                  importedPoints.add(pointData);
                }
              }
            }
          }
        }

        if (importedPoints.isNotEmpty) {
          // Debug: แสดงข้อมูลที่จะส่งไป bulk add points
          if (kDebugMode) {
            AppLogger.debug(
              '📤 Sending ${importedPoints.length} point records to bulk add:',
            );
            for (int i = 0; i < importedPoints.length; i++) {
              AppLogger.debug(
                '   [$i] pointscode=${importedPoints[i].pointscode}, amount=${importedPoints[i].pointamount}, desc=${importedPoints[i].description}',
              );
            }
          }

          // แสดง Loading Dialog พร้อมบันทึกข้อมูล
          await _showImportLoadingDialog(
            context: context,
            importType: 'points',
            itemCount: importedPoints.length,
            onImport: () => _saveBulkPoints(importedPoints),
          );
        } else {
          global.showWarningSnackBar(context, global.language('no_valid_points_in_file'));
        }
      }
    } catch (e) {
      global.showErrorSnackBar(context, '${global.language("import_points_error")}: $e');
    }
  }

  // ฟังก์ชันแปลง Excel row เป็น BulkPointsRequestModel
  BulkPointsRequestModel? _parseExcelRowToPoints(List<xl.Data?> rowData) {
    try {
      // Expected columns: pointscode, pointamount, description
      if (rowData.length < 2) {
        return null; // อย่างน้อยต้องมี pointscode และ pointamount
      }

      String pointscode = rowData[0]?.value?.toString().trim() ?? '';
      int pointamount = int.tryParse(rowData[1]?.value?.toString() ?? '0') ?? 0;
      String description = rowData.length > 2
          ? (rowData[2]?.value?.toString().trim() ?? '')
          : '';

      // ตรวจสอบว่าข้อมูลถูกต้อง
      if (pointscode.isEmpty || pointamount <= 0) {
        if (kDebugMode) {
          AppLogger.debug(
            'Skipping invalid row: pointscode=$pointscode, amount=$pointamount',
          );
        }
        return null;
      }

      return BulkPointsRequestModel(
        pointscode: pointscode,
        pointamount: pointamount,
        description: description,
      );
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error parsing point row data: $e');
      }
      return null;
    }
  }

  // ฟังก์ชันสำหรับบันทึกแต้มหลายรายการ
  Future<void> _saveBulkPoints(List<BulkPointsRequestModel> points) async {
    try {
      context.read<DebtorBloc>().add(DebtorBulkAddPoints(pointsList: points));
    } catch (e) {
      if (kDebugMode) {
        AppLogger.error('Error saving bulk points: $e');
      }
      rethrow;
    }
  }

  // ฟังก์ชันสำหรับดาวน์โหลด template ข้อมูลลูกหนี้
  Future<void> _downloadDebtorTemplate() async {
    await downloadAssetFile(
      'assets/file_import/import_debtor.xlsx',
      'import_debtor_template.xlsx',
    );
  }

  // ฟังก์ชันสำหรับดาวน์โหลด template แต้มลูกหนี้
  Future<void> _downloadPointsTemplate() async {
    await downloadAssetFile(
      'assets/file_import/import_pointbalance.xlsx',
      'import_pointbalance_template.xlsx',
    );
  }

  // ฟังก์ชันสำหรับดาวน์โหลดไฟล์ template Excel
  Future<void> downloadAssetFile(String assetPath, String saveFileName) async {
    try {
      // Load the asset file
      final ByteData data = await rootBundle.load(assetPath);
      final Uint8List bytes = data.buffer.asUint8List();

      // Use the download function
      bool success = await downloadAssetFileBytes(bytes, saveFileName);

      if (success) {
        global.showSnackBar(
          context,
          const Icon(Icons.download_done, color: Colors.white),
          global.language('download_template_success'),
          Colors.green,
        );
        if (kDebugMode) {
          AppLogger.info("File successfully downloaded.");
        }
      } else {
        global.showSnackBar(
          context,
          const Icon(Icons.error, color: Colors.white),
          global.language('download_template_failed'),
          Colors.red,
        );
        if (kDebugMode) {
          AppLogger.error("Failed to download file.");
        }
      }
    } catch (e) {
      global.showSnackBar(
        context,
        Icon(Icons.error, color: Colors.white),
        "เกิดข้อผิดพลาด: ${e.toString()}",
        Colors.red,
      );
    }
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('debtor')),
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
                            context.read<DebtorBloc>().add(
                              DebtorDeleteMany(guid: guidListChecked),
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
            padding: EdgeInsets.only(right: 20.0),
            child: IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () {
                _showImportTypeDialog();
              },
              icon: Icon(Icons.upload_file),
              tooltip: global.language('import_from_excel'),
            ),
          ),

          /// Download Template
          Padding(
            padding: EdgeInsets.only(right: 20.0),
            child: IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () {
                _showDownloadTemplateDialog();
              },
              icon: Icon(Icons.download),
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
                      selectedFilters = await filterDebtorGroup(
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
                      global.language("debtor_code"),
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
                      global.language("debtor_name"),
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

  void switchToEdit(DebtorModel value) {
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

  Widget listObject(int index, DebtorModel value, bool showCheckBox) {
    bool isCheck = false;
    for (int i = 0; i < guidListChecked.length; i++) {
      if (guidListChecked[i] == value.guidfixed) {
        isCheck = true;
        break;
      }
    }
    // ลบการ add key ออก - keys ถูกสร้างใน DebtorLoadSuccess แล้ว
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
      if (screenData.pointscode!.isEmpty) {
        screenData.pointscode = screenData.code;
      }

      if (imageFile.isNotEmpty) {
        context.read<DebtorBloc>().add(
          DebtorWithImageSave(
            debtor: screenData,
            imageFile: imageFile,
            imageWeb: imageWeb,
          ),
        );
      } else {
        context.read<DebtorBloc>().add(DebtorSave(debtor: screenData));
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
      context.read<DebtorBloc>().add(
        DebtorWithImageUpdate(
          guid: guid,
          debtor: screenData,
          imageFiles: imageFile,
          imagesUris: imageUris,
          imageWeb: imageWeb,
        ),
      );
    } else {
      screenData.images = [];
      context.read<DebtorBloc>().add(
        DebtorUpdate(guid: guid, debtorModel: screenData),
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

  void getDataToEditScreen(DebtorModel debtor) {
    isDataChange = false;
    selectGuid = debtor.guidfixed;
    debtorCode.text = debtor.code;
    screenData.code = debtor.code;
    screenData.email = debtor.email;
    screenData.fundcode = debtor.fundcode;
    screenData.creditday = debtor.creditday;
    screenData.guidfixed = debtor.guidfixed;
    screenData.images = debtor.images;
    screenData.ismember = debtor.ismember;
    screenData.personaltype = debtor.personaltype;
    screenData.customertype = debtor.customertype;
    screenData.branchnumber = debtor.branchnumber;
    screenData.taxid = debtor.taxid;
    screenData.groups = debtor.groups;
    groupSelected = debtor.groups;
    screenData.pointbalance = debtor.pointbalance;
    screenData.pointscode = debtor.pointscode;
    screenData.pricelevel = debtor.pricelevel ?? "1";
    screenData.addressforbilling.address = debtor.addressforbilling.address;
    screenData.addressforbilling.phoneprimary =
        debtor.addressforbilling.phoneprimary;
    screenData.addressforbilling.phonesecondary =
        debtor.addressforbilling.phonesecondary;
    screenData.pricelevel = debtor.pricelevel;

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

    for (var addressforshipping in debtor.addressforshipping) {
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
    for (var data in debtor.addressforbilling.contactnames!) {
      for (var ele in screenData.addressforbilling.contactnames!) {
        if (data.code == ele.code) {
          ele.name = data.name;
        }
      }
    }

    for (var data in debtor.names) {
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

    for (int i = 0; i < debtor.addressforshipping.length; i++) {
      for (var data in debtor.addressforshipping[i].contactnames!) {
        for (var ele in screenData.addressforshipping[i].contactnames!) {
          if (data.code == ele.code) {
            ele.name = data.name;
          }
        }
      }
    }

    //เก็บค่าที่ไม่ได้เปิดใช้งานภาษาเข้าทาง array
    for (var defaultValueLang in debtor.names) {
      LanguageDataModel result = screenData.names.firstWhere(
        (data) => data.code == defaultValueLang.code,
        orElse: () => LanguageDataModel(code: '', name: ''),
      );
      if (result.code == '') {
        screenData.names.add(defaultValueLang);
      }
    }

    for (var defaultValueLang in debtor.addressforbilling.contactnames!) {
      LanguageDataModel result = screenData.addressforbilling.contactnames!
          .firstWhere(
            (data) => data.code == defaultValueLang.code,
            orElse: () => LanguageDataModel(code: '', name: ''),
          );
      if (result.code == '') {
        screenData.addressforbilling.contactnames!.add(defaultValueLang);
      }
    }

    for (var addressforshipping in debtor.addressforshipping) {
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

    for (int i = 0; i < debtor.images.length; i++) {
      imageWeb.add(Uint8List(0));
      imageFile.add(File(''));
    }
  }

  Future<List<DebtorGroupModel>> getDataGroup(filter) async {
    return listDataGroup;
  }

  DebtorGroupModel getGroup(String guid) {
    DebtorGroupModel data = DebtorGroupModel(guidfixed: "", groupcode: "");
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
          readOnly: !isEditMode,
          onEditingComplete: () {
            if (kIsWeb) {
              findFocusNext(focusNodeIndex);
            }
          },
          focusNode: fieldFocusNodes[focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: debtorCode,
          textCapitalization: TextCapitalization.characters,
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.allow(RegExp('[a-z A-Z 0-9 -]')),
          ],
          onChanged: (value) {
            isDataChange = true;
            screenData.code = value.toUpperCase();
            debtorCode.value = TextEditingValue(
              text: value.toUpperCase(),
              selection: debtorCode.selection,
            );
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("debtor_code"),
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
      LanguageDataModel debtorName = screenData.names.firstWhere(
        (element) => element.code == languageList[languageIndex].code,
        orElse: () => LanguageDataModel(code: '', name: ''),
      );
      if (debtorName.code == '') {
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
                  "${global.language("debtor_name")} (${getLangName(screenData.names[languageIndex].code)})",
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
            FilteringTextInputFormatter.allow(RegExp('[a-z A-Z 0-9 -]')),
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
            FilteringTextInputFormatter.allow(RegExp('[a-z A-Z 0-9 -]')),
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
            bool isValid = (screenData.email!.isEmpty)
                ? true
                : EmailValidator.validate(screenData.email!);
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
              child: DropdownSearch<DebtorGroupModel>.multiSelection(
                enabled: isEditMode,
                key: _popupBuilderKey,
                items: (String filter, infiniteScrollProps) =>
                    getDataGroup(filter),
                compareFn: (item, selectedItem) =>
                    item.guidfixed == selectedItem.guidfixed,
                itemAsString: (DebtorGroupModel? group) {
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
                onChanged: (List<DebtorGroupModel> value) {
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
              child: DropdownSearch<DebtorGroupModel>.multiSelection(
                enabled: isEditMode,
                compareFn: (item, selectedItem) =>
                    item.guidfixed == selectedItem.guidfixed,
                items: (String filter, infiniteScrollProps) =>
                    getDataGroup(filter),
                itemAsString: (DebtorGroupModel? group) {
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
                onChanged: (List<DebtorGroupModel> value) {
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

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Row(
          children: [
            Radio(
              focusNode: FocusNode(skipTraversal: true),
              value: false,
              groupValue: screenData.ismember,
              onChanged: (value) {
                setState(() {
                  screenData.ismember = value!;
                });
              },
            ),
            Text(global.language("not_member")),
            const SizedBox(width: 10),
            Radio(
              focusNode: FocusNode(skipTraversal: true),
              value: true,
              groupValue: screenData.ismember,
              onChanged: (value) {
                setState(() {
                  screenData.ismember = value!;
                });
              },
            ),
            Text(global.language("is_member")),
          ],
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: TextFormField(
          readOnly: !isEditMode,
          onFieldSubmitted: (value) {
            findFocusNext(focusNodeIndex);
          },
          focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
          controller: TextEditingController(text: screenData.pointscode),
          onChanged: (value) {
            isDataChange = true;
            screenData.pointscode = value;
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language('loyalty_code'),
          ),
        ),
      ),
    );

    // เพิ่มฟิลด์ pricelevel
    List<PriceModel> availablePrices = global.config.prices
        .where(
          (price) =>
              price.keyNumber == 1 ||
              (price.keyNumber >= 9 && price.keyNumber <= 17),
        )
        .toList();

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: DropdownButtonFormField<String>(
          initialValue:
              availablePrices.any(
                (p) => p.keyNumber.toString() == screenData.pricelevel,
              )
              ? screenData.pricelevel
              : "1",
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language('product_details_level'),
          ),
          items: availablePrices.map<DropdownMenuItem<String>>((
            PriceModel price,
          ) {
            String priceName = "";
            if (price.names.isNotEmpty) {
              // ใช้ภาษาแรกที่มีอยู่
              priceName = price.names.first.name ?? "";
            }
            return DropdownMenuItem<String>(
              value: price.keyNumber.toString(),
              child: Text(priceName),
            );
          }).toList(),
          onChanged: isEditMode
              ? (String? newValue) {
                  if (newValue != null) {
                    setState(() {
                      isDataChange = true;
                      screenData.pricelevel = newValue;
                    });
                  }
                }
              : null,
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Row(
          children: [
            Expanded(
              flex: 3,
              child: TextFormField(
                readOnly: true,
                onFieldSubmitted: (value) {
                  findFocusNext(focusNodeIndex);
                },
                keyboardType: TextInputType.number,
                inputFormatters: <TextInputFormatter>[
                  FilteringTextInputFormatter.digitsOnly,
                ],
                focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
                controller: TextEditingController(
                  text: screenData.pointbalance.toString(),
                ),
                onChanged: (value) {
                  isDataChange = true;
                  screenData.pointbalance = int.parse(
                    value.isEmpty ? "0" : value,
                  );
                },
                decoration: InputDecoration(
                  floatingLabelBehavior: FloatingLabelBehavior.always,
                  border: const OutlineInputBorder(),
                  labelText: global.language('points_balance'),
                ),
              ),
            ),
            const SizedBox(width: 10),
            Expanded(
              flex: 1,
              child: SizedBox(
                height: 50,
                child: ElevatedButton.icon(
                  focusNode: FocusNode(skipTraversal: true),
                  onPressed: screenData.code.isNotEmpty
                      ? () async {
                          await Navigator.push(
                            context,
                            MaterialPageRoute(
                              builder: (context) => BlocProvider(
                                create: (context) => PointTransactionBloc(
                                  pointTransactionRepository:
                                      PointTransactionRepository(),
                                ),
                                child: PointTransactionScreen(
                                  debtorCode: screenData.code,
                                  debtorName: global.packName(screenData.names),
                                  pointsCode: screenData.pointscode!,
                                ),
                              ),
                            ),
                          );
                          // Reload ข้อมูลเมื่อกลับมาจากหน้าประวัติ
                          if (selectGuid.isNotEmpty) {
                            getData(selectGuid);
                          }
                        }
                      : null,
                  icon: Icon(Icons.history, size: 16),
                  label: Text(global.language('edit_history'), style: TextStyle(fontSize: 12)),
                  style: ElevatedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 8,
                      vertical: 12,
                    ),
                  ),
                ),
              ),
            ),
            const SizedBox(width: 10),

            /// เพิ่มแต้ม
            Expanded(
              flex: 1,
              child: SizedBox(
                height: 50,
                child: ElevatedButton.icon(
                  focusNode: FocusNode(skipTraversal: true),
                  onPressed: isEditMode && screenData.code.isNotEmpty
                      ? () {
                          showAddPointDialog(context);
                        }
                      : null,
                  icon: Icon(Icons.add, size: 16),
                  label: Text(
                    global.language('add_points'),
                    style: TextStyle(fontSize: 12),
                  ),
                  style: ElevatedButton.styleFrom(
                    padding: const EdgeInsets.symmetric(
                      horizontal: 8,
                      vertical: 12,
                    ),
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Text(global.language("address_for_billing")),
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
                  "${global.language("debtor_name")} (${getLangName(screenData.addressforbilling.contactnames![languageIndex].code)})",
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
    if (screenData.addressforbilling.address!.isEmpty) {
      screenData.addressforbilling.address!.add('');
    }

    for (int addressIndex = 0; addressIndex < 1; addressIndex++) {
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
              labelText: global.language("address"),
              // labelText: global.language("tax_address_${addressIndex + 1}"),
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
                    "${global.language("debtor_name")} (${getLangName(screenData.addressforshipping[addressShipIndex].contactnames![languageIndex].code)})",
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
              child: Text(global.language("debtor_image")),
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
    formWidgets.add(const SizedBox(height: 5));
    if (isEditMode) {
      formWidgets.add(
        Container(
          width: double.infinity,
          padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: ElevatedButton.icon(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () {
              setState(() {
                List<LanguageDataModel> names = [];
                for (int k = 0; k < languageList.length; k++) {
                  names.add(
                    LanguageDataModel(code: languageList[k].code!, name: ""),
                  );
                }
                screenData.addressforshipping.add(
                  CustomerAddressModel(
                    guid: "",
                    address: ["", "", ""],
                    countrycode: "",
                    provincecode: "",
                    districtcode: "",
                    subdistrictcode: "",
                    zipcode: "",
                    latitude: 0,
                    longitude: 0,
                    phoneprimary: "",
                    phonesecondary: "",
                    contactnames: names,
                  ),
                );
              });
            },
            icon: Icon(Icons.add),
            label: Text(global.language("add_address_shipping")),
          ),
        ),
      );
    }

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
        title: Text(headerEdit + global.language("debtor")),
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
                            context.read<DebtorBloc>().add(
                              DebtorDelete(guid: selectGuid),
                            );
                          },
                          child: Text(global.language('confirm')),
                        ),
                      ],
                    ),
                  );
                  // ลบ setState() เพราะ BlocListener จะ handle การ rebuild
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
                  bool needsUpdate = false;
                  for (int i = 0; i < screenData.names.length; i++) {
                    if (screenData.names[i].name.trim().isEmpty) {
                      var data = await global.translateNames(
                        namesData: screenData.names,
                      );
                      screenData.names[i].name = data[i].name;
                      needsUpdate = true;
                    }
                  }

                  for (
                    int i = 0;
                    i < screenData.addressforbilling.contactnames!.length;
                    i++
                  ) {
                    if (screenData.addressforbilling.contactnames![i].name
                        .trim()
                        .isEmpty) {
                      var data = await global.translateNames(
                        namesData: screenData.addressforbilling.contactnames!,
                      );
                      screenData.addressforbilling.contactnames![i].name =
                          data[i].name;
                      needsUpdate = true;
                    }
                  }

                  // เรียก setState เพียงครั้งเดียวหลังจบ loop
                  if (needsUpdate || isLoadTranslation) {
                    setState(() {
                      isLoadTranslation = false;
                    });
                  }
                },
                icon: const Icon(Icons.translate),
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
          const ManualButton(path: 'debtor'),
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

  int _buildCount = 0;

  @override
  Widget build(BuildContext context) {
    _buildCount++;

    // ป้องกัน infinite loop - ถ้า build เกิน 100 ครั้ง ให้หยุด
    if (_buildCount > 100) {
      if (kDebugMode) {
        AppLogger.debug('🚨 INFINITE LOOP DETECTED! Build count: $_buildCount');
      }
      return Scaffold(
        appBar: AppBar(
          title: Text(global.language('error_infinite_loop')),
          backgroundColor: Colors.red,
        ),
        body: Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              const Icon(Icons.error, size: 64, color: Colors.red),
              const SizedBox(height: 16),
              Text(
                'Infinite Loop Detected!\nBuild count: $_buildCount',
                textAlign: TextAlign.center,
                style: const TextStyle(fontSize: 18),
              ),
              const SizedBox(height: 16),
              ElevatedButton(
                onPressed: () => Navigator.pop(context),
                child: const Text('Go Back'),
              ),
            ],
          ),
        ),
      );
    }

    if (kDebugMode) {
      AppLogger.debug('═══════════════════════════════════════');
      AppLogger.debug('🔄 Debtor Screen build() called: $_buildCount');
      AppLogger.debug('listData.length: ${listData.length}');
      AppLogger.debug('listKeys.length: ${listKeys.length}');
      AppLogger.debug('loadingData: $loadingData');
      AppLogger.debug('═══════════════════════════════════════');
    }

    queryData = MediaQuery.of(context);

    // ย้ายการ clear ออกมาใช้เฉพาะเมื่อจำเป็น
    // ไม่ควร mutate state ใน build method

    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return MultiBlocListener(
            listeners: [
              BlocListener<DebtorBloc, DebtorState>(
                listener: (context, state) {
                  if (kDebugMode) {
                    AppLogger.debug(
                      '📥 DebtorBloc State: ${state.runtimeType}',
                    );
                  }
                  blocCurrentState = state;
                  // Load
                  if (state is DebtorLoadSuccess) {
                    if (kDebugMode) {
                      AppLogger.debug(
                        '✅ DebtorLoadSuccess: ${state.debtors.length} items, total now: ${listData.length + state.debtors.length}',
                      );
                    }
                    setState(() {
                      if (kDebugMode) {
                        AppLogger.info('   🔄 setState in DebtorLoadSuccess');
                      }
                      loadingData = false;
                      if (state.debtors.isNotEmpty) {
                        listData.addAll(state.debtors);
                        // Clear และ recreate keys เฉพาะเมื่อมีการเพิ่ม data
                        listKeys.clear();
                        // สร้าง keys ใหม่ให้ตรงกับจำนวน items
                        for (int i = 0; i < listData.length; i++) {
                          listKeys.add(GlobalKey());
                        }
                        if (kDebugMode) {
                          AppLogger.debug(
                            '   🔑 listKeys recreated, listData.length: ${listData.length}, listKeys.length: ${listKeys.length}',
                          );
                        }
                      }
                      // Clear checked list เมื่อไม่ได้อยู่ในโหมด checkbox
                      if (showCheckBox == false) {
                        guidListChecked.clear();
                      }
                    });
                  }
                  if (state is DebtorLoadFailed) {
                    if (kDebugMode) {
                      AppLogger.error('❌ DebtorLoadFailed: ${state.message}');
                    }
                    setState(() {
                      loadingData = false;
                    });
                    // แสดง error แต่ไม่ retry อัตโนมัติ
                    if (mounted) {
                      global.showSnackBar(
                        context,
                        const Icon(Icons.error, color: Colors.white),
                        'โหลดข้อมูลล้มเหลว: ${state.message}',
                        Colors.red,
                      );
                    }
                  }
                  // Save
                  if (state is DebtorSaveSuccess) {
                    if (kDebugMode) {
                      AppLogger.debug(
                        '💾 DebtorSaveSuccess - clearing data and reloading',
                      );
                    }
                    global.showSnackBar(
                      context,
                      Icon(Icons.save, color: Colors.white),
                      global.language("save_success"),
                      Colors.blue,
                    );
                    setState(() {
                      clearEditData();
                      listData.clear();
                      listKeys.clear();
                    });
                    // เรียกหลัง setState เพื่อป้องกัน infinite loop
                    loadDataList(searchText, selectedFilterCodes);
                  }
                  if (state is DebtorSaveFailed) {
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
                  if (state is DebtorUpdateSuccess) {
                    global.showSnackBar(
                      context,
                      Icon(Icons.edit, color: Colors.white),
                      global.language("edit_success"),
                      Colors.blue,
                    );
                    setState(() {
                      clearEditData();
                      listData.clear();
                      listKeys.clear();
                      isSaveAllow = false;
                    });
                    // เรียกหลัง setState เพื่อป้องกัน infinite loop
                    WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                      tabController.animateTo(0);
                    });
                    loadDataList(searchText, selectedFilterCodes);
                    getData(selectGuid);
                  }
                  if (state is DebtorUpdateFailed) {
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
                  if (state is DebtorDeleteSuccess) {
                    global.showSnackBar(
                      context,
                      Icon(Icons.delete, color: Colors.white),
                      global.language("delete_success"),
                      Colors.blue,
                    );
                    setState(() {
                      listData.clear();
                      listKeys.clear();
                      clearEditData();
                    });
                    // เรียกหลัง setState เพื่อป้องกัน infinite loop
                    WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                      tabController.animateTo(0);
                    });
                    loadDataList(searchText, selectedFilterCodes);
                  }
                  // Delete Many
                  if (state is DebtorDeleteManySuccess) {
                    global.showSnackBar(
                      context,
                      Icon(Icons.delete, color: Colors.white),
                      global.language("not_delete_success"),
                      Colors.blue,
                    );
                    setState(() {
                      listData.clear();
                      listKeys.clear();
                      clearEditData();
                      showCheckBox = false;
                    });
                    // เรียกหลัง setState เพื่อป้องกัน infinite loop
                    loadDataList(searchText, selectedFilterCodes);
                  }
                  // Get
                  if (state is DebtorGetSuccess) {
                    setState(() {
                      getDataToEditScreen(state.debtors);
                    });
                    // ย้าย side effects ออกมาจาก setState
                    if (isEditMode) {
                      WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                        tabController.animateTo(1);
                        // เรียก findFocusNext หลังจาก animate เสร็จ
                        findFocusNext(0);
                      });
                    }
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
                          // ย้าย animateTo ออกมาจาก setState
                          listScrollController.animateTo(
                            listScrollController.offset -
                                (boxHeader.size.height + box.size.height),
                            duration: const Duration(milliseconds: 100),
                            curve: Curves.ease,
                          );
                          setState(() {
                            isKeyUp = false;
                          });
                        }
                        // Scroll Down
                        if (isKeyDown &&
                            position.dy > (queryData.size.height - 100)) {
                          // ย้าย animateTo ออกมาจาก setState
                          listScrollController.animateTo(
                            listScrollController.offset +
                                (position.dy - (queryData.size.height - 100)),
                            duration: const Duration(milliseconds: 100),
                            curve: Curves.easeOut,
                          );
                          setState(() {
                            isKeyDown = false;
                          });
                        }
                      }
                    }
                  }
                  // Bulk Save
                  if (state is DebtorBulkSaveSuccess) {
                    global.showSnackBar(
                      context,
                      const Icon(Icons.upload_file, color: Colors.white),
                      'Import สำเร็จ ${state.savedCount} รายการ',
                      Colors.green,
                    );
                    setState(() {
                      listData.clear();
                      listKeys.clear();
                    });
                    // เรียกหลัง setState เพื่อป้องกัน infinite loop
                    loadDataList(searchText, selectedFilterCodes);
                  }
                  if (state is DebtorBulkSaveFailed) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        const Icon(Icons.upload_file, color: Colors.white),
                        'Import ไม่สำเร็จ: ${state.message}',
                        Colors.red,
                      );
                    });
                  }
                  // Bulk Add Points
                  if (state is DebtorBulkAddPointsSuccess) {
                    // สร้างข้อความสำหรับแสดงผล
                    String message =
                        '${global.language("import_points_success")} ${state.successCount} ${global.language("items")}';

                    if (state.failedCount > 0) {
                      message += ' (${global.language("failed")} ${state.failedCount} ${global.language("items")})';

                      // แสดงรายละเอียดของรายการที่ล้มเหลว
                      if (state.failedItems != null &&
                          state.failedItems!.isNotEmpty) {
                        String failedDetails = '\n';
                        for (var item in state.failedItems!.take(3)) {
                          String pointscode = item['pointscode'] ?? 'N/A';
                          String reason = item['reason'] ?? 'Unknown error';
                          failedDetails += '• $pointscode: $reason\n';
                        }
                        if (state.failedItems!.length > 3) {
                          failedDetails +=
                              '• ${global.language("and_more")} ${state.failedItems!.length - 3} ${global.language("items")}';
                        }
                        message += failedDetails;
                      }
                    }

                    global.showSnackBar(
                      context,
                      const Icon(Icons.stars, color: Colors.white),
                      message,
                      state.failedCount > 0 ? Colors.orange : Colors.green,
                    );

                    if (kDebugMode) {
                      AppLogger.debug(
                        '✅ Bulk add points success: ${state.successCount}/${state.totalCount}',
                      );
                      if (state.successItems != null) {
                        AppLogger.debug(
                          '   Success items: ${state.successItems!.join(", ")}',
                        );
                      }
                      if (state.failedItems != null) {
                        AppLogger.debug(
                          '   Failed items: ${state.failedItems!.length}',
                        );
                      }
                    }
                  }
                  if (state is DebtorBulkAddPointsFailed) {
                    global.showSnackBar(
                      context,
                      const Icon(Icons.error, color: Colors.white),
                      '${global.language("import_points_failed")}: ${state.message}',
                      Colors.red,
                    );
                    if (kDebugMode) {
                      AppLogger.error(
                        '❌ Bulk add points failed: ${state.message}',
                      );
                    }
                  }
                },
              ),
              BlocListener<DebtorGroupBloc, DebtorGroupState>(
                listener: (context, state) {
                  // Load
                  if (state is DebtorGroupLoadSuccess) {
                    setState(() {
                      if (state.debtorGroups.isNotEmpty) {
                        listDataGroup = state.debtorGroups;
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

  Future<List<DebtorGroupModel>> filterDebtorGroup(
    List<DebtorGroupModel> selectedFilters,
  ) async {
    List<DebtorGroupModel> selectedValues = selectedFilters;

    await showDialog(
      context: context,
      builder: (BuildContext context) {
        return StatefulBuilder(
          builder: (BuildContext context, StateSetter setState) {
            return AlertDialog(
              title: Column(
                children: [
                  Text(global.language('filter_debtor_group')),
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

  // ฟังก์ชันสำหรับเพิ่มแต้ม
  Future<void> showAddPointDialog(BuildContext context) async {
    if (screenData.pointscode == null || screenData.pointscode!.isEmpty) {
      showDialog(
        context: context,
        builder: (BuildContext context) {
          return AlertDialog(
            title: Text(global.language("error")),
            content: const Text("กรุณาเลือกรหัสแต้มก่อน"),
            actions: [
              TextButton(
                onPressed: () => Navigator.of(context).pop(),
                child: Text(global.language("ok")),
              ),
            ],
          );
        },
      );
      return;
    }

    final formKey = GlobalKey<FormState>();
    TextEditingController pointAmountController = TextEditingController();
    TextEditingController descriptionController = TextEditingController();

    await showDialog(
      context: context,
      barrierDismissible: false,
      builder: (BuildContext dialogContext) {
        bool isLoading = false;

        return StatefulBuilder(
          builder: (context, setDialogState) {
            return BlocListener<DebtorBloc, DebtorState>(
              listener: (context, state) {
                if (state is DebtorAddPointSuccess) {
                  // ปิด dialog
                  Navigator.of(dialogContext).pop();

                  // แสดงข้อความสำเร็จ
                  global.showSuccessSnackBar(context, global.language('add_point_success'));

                  // Reload ข้อมูลลูกหนี้เพื่อ อัพเดทยอดแต้ม
                  if (selectGuid.isNotEmpty) {
                    getData(selectGuid);
                  }
                } else if (state is DebtorAddPointFailed) {
                  // ปิด dialog
                  Navigator.of(dialogContext).pop();

                  // แสดงข้อความ error
                  global.showErrorSnackBar(context, 'เพิ่มแต้มไม่สำเร็จ: ${state.message}');
                }
              },
              child: Stack(
                children: [
                  AlertDialog(
                    title: Text(global.language("add_points")),
                    content: Form(
                      key: formKey,
                      child: SizedBox(
                        width: 400,
                        child: Column(
                          mainAxisSize: MainAxisSize.min,
                          children: [
                            // แสดงรหัสแต้ม
                            Padding(
                              padding: const EdgeInsets.only(bottom: 16),
                              child: Row(
                                children: [
                                  const Text(
                                    'รหัสแต้ม: ',
                                    style: TextStyle(
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                  Text(screenData.pointscode!),
                                ],
                              ),
                            ),
                            // จำนวนแต้ม
                            TextFormField(
                              controller: pointAmountController,
                              keyboardType: TextInputType.number,
                              enabled: !isLoading,
                              decoration: const InputDecoration(
                                labelText: 'จำนวนแต้ม *',
                                border: OutlineInputBorder(),
                                hintText: 'ระบุจำนวนแต้มที่ต้องการเพิ่ม',
                              ),
                              inputFormatters: [
                                FilteringTextInputFormatter.digitsOnly,
                              ],
                              validator: (value) {
                                if (value == null || value.trim().isEmpty) {
                                  return 'กรุณาระบุจำนวนแต้ม';
                                }
                                int? amount = int.tryParse(value.trim());
                                if (amount == null || amount <= 0) {
                                  return 'จำนวนแต้มต้องมากกว่า 0';
                                }
                                return null;
                              },
                            ),
                            const SizedBox(height: 16),
                            // คำอธิบาย
                            TextFormField(
                              controller: descriptionController,
                              maxLines: 3,
                              enabled: !isLoading,
                              decoration: const InputDecoration(
                                labelText: 'คำอธิบาย *',
                                border: OutlineInputBorder(),
                                hintText: 'ระบุเหตุผลในการเพิ่มแต้ม',
                              ),
                              validator: (value) {
                                if (value == null || value.trim().isEmpty) {
                                  return 'กรุณาระบุคำอธิบาย';
                                }
                                return null;
                              },
                            ),
                          ],
                        ),
                      ),
                    ),
                    actions: [
                      TextButton(
                        onPressed: isLoading
                            ? null
                            : () => Navigator.of(dialogContext).pop(),
                        child: Text(global.language("cancel")),
                      ),
                      ElevatedButton(
                        onPressed: isLoading
                            ? null
                            : () async {
                                // Validate form
                                if (!formKey.currentState!.validate()) {
                                  return;
                                }

                                // แสดง loading ทันที
                                setDialogState(() {
                                  isLoading = true;
                                });

                                int pointAmount = int.parse(
                                  pointAmountController.text.trim(),
                                );

                                // สร้าง model
                                final pointTransaction =
                                    PointTransactionCreateModel(
                                      pointscode: screenData.pointscode!,
                                      pointamount: pointAmount,
                                      description: descriptionController.text
                                          .trim(),
                                    );

                                // เรียก API ผ่าน Bloc
                                context.read<DebtorBloc>().add(
                                  DebtorAddPoint(
                                    pointscode: screenData.pointscode!,
                                    pointTransaction: pointTransaction,
                                  ),
                                );
                              },
                        child: Text(global.language('add_points')),
                      ),
                    ],
                  ),
                  // Loading Overlay
                  if (isLoading)
                    LoadingOverlay(
                      message: 'กำลังเพิ่มแต้ม...',
                      subtitle: global.language('please_wait'),
                    ),
                ],
              ),
            );
          },
        );
      },
    );
  }
}
