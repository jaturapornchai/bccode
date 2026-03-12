// ignore_for_file: deprecated_member_use

import 'dart:async';
import 'dart:io';

import 'package:smlaicloud/widgets/manual_button.dart';
import 'package:smlaicloud/bloc/company_branch/company_branch_bloc.dart';
import 'package:smlaicloud/bloc/image/image_upload_bloc.dart';
import 'package:smlaicloud/model/business_type_model.dart';
import 'package:smlaicloud/model/currency_model.dart';
import 'package:smlaicloud/services/currency_api_service.dart';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/contact_model.dart';
import 'package:smlaicloud/screens/config/map_get_location_screen.dart';
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
import 'package:dropdown_search/dropdown_search.dart';

// Import custom widgets
import 'package:smlaicloud/screens/components/company_branch_screen/company_info_form_widget.dart';
import 'package:smlaicloud/screens/components/company_branch_screen/branch_info_form_widget.dart';
import 'package:smlaicloud/screens/components/company_branch_screen/tax_id_widget.dart';
import 'package:smlaicloud/screens/components/company_branch_screen/business_type_selector_widget.dart';
import 'package:smlaicloud/screens/components/company_branch_screen/business_property_widget.dart';
import 'package:smlaicloud/screens/components/company_branch_screen/language_selector_widget.dart';
import 'package:smlaicloud/screens/components/company_branch_screen/vat_config_widget.dart';
import 'package:smlaicloud/screens/components/company_branch_screen/payment_rounding_widget.dart';
import 'package:smlaicloud/screens/components/company_branch_screen/receipt_config_widget.dart';
import 'package:smlaicloud/screens/components/company_branch_screen/image_picker_widget.dart';
import 'package:smlaicloud/screens/components/company_branch_screen/coupon_usage_type_widget.dart';

class CompanyBranchScreen extends StatefulWidget {
  const CompanyBranchScreen({super.key});

  @override
  State<CompanyBranchScreen> createState() => CompanyBranchScreenState();
}

class CompanyBranchScreenState extends State<CompanyBranchScreen>
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

  List<LanguageModel> defaultlanguageList = [
    LanguageModel(code: "th", codeTranslator: "th", name: "Thai", isuse: false),
    LanguageModel(
      code: "en",
      codeTranslator: "en",
      name: "English",
      isuse: false,
    ),
    LanguageModel(
      code: "cn",
      codeTranslator: "cn",
      name: "Chinese",
      isuse: false,
    ),
    LanguageModel(
      code: "ja",
      codeTranslator: "ja",
      name: "Japanese",
      isuse: false,
    ),
    LanguageModel(
      code: "ko",
      codeTranslator: "ko",
      name: "Korean",
      isuse: false,
    ),
    LanguageModel(code: "lo", codeTranslator: "lo", name: "Lao", isuse: false),
    LanguageModel(
      code: "my",
      codeTranslator: "my",
      name: "Burmese",
      isuse: false,
    ),
    LanguageModel(
      code: "ms",
      codeTranslator: "ms",
      name: "Malaysian",
      isuse: false,
    ),
    LanguageModel(
      code: "vi",
      codeTranslator: "vi",
      name: "Vietnamese",
      isuse: false,
    ),
    LanguageModel(
      code: "km",
      codeTranslator: "km",
      name: "Khmer",
      isuse: false,
    ),
  ];

  bool businesstypeNull = false;

  bool isExpanded = false;

  // Currency service และรายการสกุลเงินจาก API
  final CurrencyApiService _currencyApiService = CurrencyApiService();
  List<CurrencyModel> _currencies = [];
  bool _currenciesLoading = false;

  /// โหลดรายการสกุลเงินจาก API
  Future<void> _loadCurrencies() async {
    if (_currenciesLoading) return;
    setState(() {
      _currenciesLoading = true;
    });

    try {
      final String shopId = global.prefs.getString("shopid") ?? "";
      final response = await _currencyApiService.getCurrencies(
        shopId: shopId,
        disabled: false, // โหลดเฉพาะสกุลเงินที่ใช้งานอยู่
        limit: 100,
      );

      if (response.success && response.data != null) {
        setState(() {
          _currencies = (response.data as List)
              .map((e) => CurrencyModel.fromJson(e))
              .toList();
        });
      }
    } catch (e) {
      // ถ้าโหลดไม่ได้ ใช้ค่า default THB
      if (_currencies.isEmpty) {
        setState(() {
          _currencies = [
            CurrencyModel(code: 'THB', name: 'บาท', symbol: '฿'),
          ];
        });
      }
    } finally {
      setState(() {
        _currenciesLoading = false;
      });
    }
  }

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
    _loadCurrencies(); // โหลดรายการสกุลเงิน
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

    // Initialize payment rounding controllers
    initPaymentRoundingControllers();

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
      baseCurrency: 'THB',
      language: 'th',
      timezone: 'Asia/Bangkok',
      yeartype: 'christian',
      timezonelabel: '(UTC+07:00) Bangkok, Hanoi, Jakarta, Ho Chi Minh',
      timezoneoffset: '+07:00',
      dateFormat: 'dd/MM/yyyy',
      decimalQuantity: 2,
      decimalPrice: 2,
      decimalDocument: 2,
    );

    isDataChange = false;
    focusNodeIndex = 0;
    refreshFocus = true;

    imageFile = [File('')];
    imageWeb = [Uint8List(0)];
    logoFile = [File('')];
    logoWeb = [Uint8List(0)];

    businesstypeNull = false;

    // Reset payment rounding controllers
    initPaymentRoundingControllers();
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
        title: Text(global.language('companyBranch')),
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
          const ManualButton(path: 'settings-branch'),
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
                            context.read<CompanyBranchBloc>().add(
                              CompanyBranchDeleteMany(guid: guidListChecked),
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
                      screenData.companynames = global.shopSelectData.names!;
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

  void saveOrUpdateData() {
    // sync languages จาก UI state → model ก่อน save
    screenData.languages = List<String>.from(selectlanguageList);

    if (verifyData(screenData)) {
      showCheckBox = false;

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
    screenData.machinetype = companyBranch.machinetype;
    screenData.couponusetype = companyBranch.couponusetype;
    screenData.departments = companyBranch.departments;
    // ใช้ isNotEmpty เพราะ MongoDB อาจส่ง empty string "" กลับมา ซึ่ง ?? ไม่จัดการ
    screenData.baseCurrency = (companyBranch.baseCurrency != null && companyBranch.baseCurrency!.isNotEmpty)
        ? companyBranch.baseCurrency
        : 'THB';
    screenData.language = (companyBranch.language != null && companyBranch.language!.isNotEmpty)
        ? companyBranch.language
        : 'th';
    screenData.timezone = (companyBranch.timezone != null && companyBranch.timezone!.isNotEmpty)
        ? companyBranch.timezone
        : 'Asia/Bangkok';
    screenData.yeartype = (companyBranch.yeartype != null && companyBranch.yeartype!.isNotEmpty)
        ? companyBranch.yeartype
        : 'christian';
    screenData.timezonelabel = (companyBranch.timezonelabel != null && companyBranch.timezonelabel!.isNotEmpty)
        ? companyBranch.timezonelabel
        : '(UTC+07:00) Bangkok, Hanoi, Jakarta, Ho Chi Minh';
    screenData.timezoneoffset = (companyBranch.timezoneoffset != null && companyBranch.timezoneoffset!.isNotEmpty)
        ? companyBranch.timezoneoffset
        : '+07:00';
    screenData.dateFormat = (companyBranch.dateFormat != null && companyBranch.dateFormat!.isNotEmpty)
        ? companyBranch.dateFormat
        : 'dd/MM/yyyy';
    // ค่าทศนิยมต้องอยู่ในช่วง 2-8 ถ้าเป็น 0 หรือ null ให้ใช้ค่าเริ่มต้น 2
    screenData.decimalQuantity = (companyBranch.decimalQuantity != null && companyBranch.decimalQuantity! >= 2)
        ? companyBranch.decimalQuantity
        : 2;
    screenData.decimalPrice = (companyBranch.decimalPrice != null && companyBranch.decimalPrice! >= 2)
        ? companyBranch.decimalPrice
        : 2;
    screenData.decimalDocument = (companyBranch.decimalDocument != null && companyBranch.decimalDocument! >= 2)
        ? companyBranch.decimalDocument
        : 2;

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

    // sync languages list จาก model → UI state
    selectlanguageList = List<String>.from(companyBranch.languages);
    if (selectlanguageList.isEmpty) {
      selectlanguageList = ['th'];
    }

    screenData.paymentrounding =
        companyBranch.paymentrounding ?? PaymentRoundingModel();

    // Set up controllers for the existing rounding rules
    updateRoundingControllers();
  }

  void updateRoundingControllers() {
    // Clear existing controllers
    if (screenData.guidfixed.isEmpty) {
      initPaymentRoundingControllers();
    }

    if (screenData.paymentrounding == null) return;

    // Setup controllers for cash
    updateMethodControllers('cash', screenData.paymentrounding!.cash.rules);
    // Setup controllers for creditcard
    updateMethodControllers(
      'creditcard',
      screenData.paymentrounding!.creditcard.rules,
    );
    // Setup controllers for banktransfer
    updateMethodControllers(
      'banktransfer',
      screenData.paymentrounding!.banktransfer.rules,
    );
    // Setup controllers for cheque
    updateMethodControllers('cheque', screenData.paymentrounding!.cheque.rules);
    // Setup controllers for coupon
    updateMethodControllers('coupon', screenData.paymentrounding!.coupon.rules);
    // Setup controllers for delivery
    updateMethodControllers(
      'delivery',
      screenData.paymentrounding!.delivery.rules,
    );
    // Setup controllers for qrcode
    updateMethodControllers('qrcode', screenData.paymentrounding!.qrcode.rules);
  }

  void updateMethodControllers(String method, List<RoundingRuleModel> rules) {
    roundingControllers[method]?.clear();
    for (var rule in rules) {
      List<TextEditingController> ruleControllers = [
        TextEditingController(text: rule.lowerbound.toString()),
        TextEditingController(text: rule.upperbound.toString()),
        TextEditingController(text: rule.roundto.toString()),
      ];
      roundingControllers[method]?.add(ruleControllers);
    }
    // If no rules exist, add an empty one
    if (rules.isEmpty) {
      addNewRoundingRule(method);
    }
  }

  void addNewRoundingRule(String method) {
    List<TextEditingController> controllers = [
      TextEditingController(text: "0.01"),
      TextEditingController(text: "0.99"),
      TextEditingController(text: "0"),
    ];
    roundingControllers[method]?.add(controllers);

    // Update the model
    RoundingRuleModel newRule = RoundingRuleModel(
      lowerbound: 0.01,
      upperbound: 0.99,
      roundto: 0,
    );

    // Update the model
    switch (method) {
      case 'cash':
        screenData.paymentrounding?.cash.rules.add(newRule);
        break;
      case 'creditcard':
        screenData.paymentrounding?.creditcard.rules.add(newRule);
        break;
      case 'banktransfer':
        screenData.paymentrounding?.banktransfer.rules.add(newRule);
        break;
      case 'cheque':
        screenData.paymentrounding?.cheque.rules.add(newRule);
        break;
      case 'coupon':
        screenData.paymentrounding?.coupon.rules.add(newRule);
        break;
      case 'delivery':
        screenData.paymentrounding?.delivery.rules.add(newRule);
        break;
      case 'qrcode':
        screenData.paymentrounding?.qrcode.rules.add(newRule);
        break;
    }
    setState(() {});
  }

  void removeRoundingRule(String method, int index) {
    if (roundingControllers[method]!.length > 1) {
      roundingControllers[method]?.removeAt(index);

      // Update the model
      switch (method) {
        case 'cash':
          if (index < screenData.paymentrounding!.cash.rules.length) {
            screenData.paymentrounding!.cash.rules.removeAt(index);
          }
          break;
        case 'creditcard':
          if (index < screenData.paymentrounding!.creditcard.rules.length) {
            screenData.paymentrounding!.creditcard.rules.removeAt(index);
          }
          break;
        case 'banktransfer':
          if (index < screenData.paymentrounding!.banktransfer.rules.length) {
            screenData.paymentrounding!.banktransfer.rules.removeAt(index);
          }
          break;
        case 'cheque':
          if (index < screenData.paymentrounding!.cheque.rules.length) {
            screenData.paymentrounding!.cheque.rules.removeAt(index);
          }
          break;
        case 'coupon':
          if (index < screenData.paymentrounding!.coupon.rules.length) {
            screenData.paymentrounding!.coupon.rules.removeAt(index);
          }
          break;
        case 'delivery':
          if (index < screenData.paymentrounding!.delivery.rules.length) {
            screenData.paymentrounding!.delivery.rules.removeAt(index);
          }
          break;
        case 'qrcode':
          if (index < screenData.paymentrounding!.qrcode.rules.length) {
            screenData.paymentrounding!.qrcode.rules.removeAt(index);
          }
          break;
      }
      setState(() {});
    }
  }

  Widget buildPaymentRoundingSection() {
    List<Widget> paymentRoundingWidgets = [];

    // Title
    paymentRoundingWidgets.add(
      Padding(
        padding: EdgeInsets.all(10.0),
        child: Container(
          margin: EdgeInsets.only(top: 10, bottom: 10),
          padding: EdgeInsets.all(10),
          decoration: BoxDecoration(
            border: Border.all(color: global.theme.primaryColor),
            borderRadius: BorderRadius.circular(5),
            color: Colors.white,
          ),
          child: Column(
            children: [
              Text(
                global.language("payment_rounding"),
                style: const TextStyle(
                  fontSize: 20,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 20),

              // Payment Methods Tabs
              DefaultTabController(
                length: 7,
                child: Column(
                  children: [
                    TabBar(
                      isScrollable: true,
                      labelColor: global.theme.primaryColor,
                      unselectedLabelColor: Colors.grey,
                      tabs: [
                        Tab(text: global.language("cash")),
                        Tab(text: global.language("credit_card")),
                        Tab(text: global.language("bank_transfer")),
                        Tab(text: global.language("cheque")),
                        Tab(text: global.language("coupon")),
                        Tab(text: global.language("delivery")),
                        Tab(text: global.language("qrcode")),
                      ],
                    ),
                    const SizedBox(height: 10),
                    SizedBox(
                      height: 500,
                      child: TabBarView(
                        children: [
                          buildPaymentMethodRounding('cash'),
                          buildPaymentMethodRounding('creditcard'),
                          buildPaymentMethodRounding('banktransfer'),
                          buildPaymentMethodRounding('cheque'),
                          buildPaymentMethodRounding('coupon'),
                          buildPaymentMethodRounding('delivery'),
                          buildPaymentMethodRounding('qrcode'),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );

    return Column(children: paymentRoundingWidgets);
  }

  Widget buildPaymentMethodRounding(String method) {
    bool isEnabled = false;

    // Get enabled state from the model
    switch (method) {
      case 'cash':
        isEnabled = screenData.paymentrounding?.cash.enabled ?? false;
        break;
      case 'creditcard':
        isEnabled = screenData.paymentrounding?.creditcard.enabled ?? false;
        break;
      case 'banktransfer':
        isEnabled = screenData.paymentrounding?.banktransfer.enabled ?? false;
        break;
      case 'cheque':
        isEnabled = screenData.paymentrounding?.cheque.enabled ?? false;
        break;
      case 'coupon':
        isEnabled = screenData.paymentrounding?.coupon.enabled ?? false;
        break;
      case 'delivery':
        isEnabled = screenData.paymentrounding?.delivery.enabled ?? false;
        break;
      case 'qrcode':
        isEnabled = screenData.paymentrounding?.qrcode.enabled ?? false;
        break;
    }

    return Column(
      children: [
        // Enable/Disable switch
        Row(
          mainAxisAlignment: MainAxisAlignment.start,
          children: [
            Switch(
              value: isEnabled,
              onChanged: (value) {
                setState(() {
                  isDataChange = true;
                  // Update the model
                  switch (method) {
                    case 'cash':
                      screenData.paymentrounding?.cash.enabled = value;
                      break;
                    case 'creditcard':
                      screenData.paymentrounding?.creditcard.enabled = value;
                      break;
                    case 'banktransfer':
                      screenData.paymentrounding?.banktransfer.enabled = value;
                      break;
                    case 'cheque':
                      screenData.paymentrounding?.cheque.enabled = value;
                      break;
                    case 'coupon':
                      screenData.paymentrounding?.coupon.enabled = value;
                      break;
                    case 'delivery':
                      screenData.paymentrounding?.delivery.enabled = value;
                      break;
                    case 'qrcode':
                      screenData.paymentrounding?.qrcode.enabled = value;
                      break;
                  }
                });
              },
            ),
            Text(global.language("enable_rounding")),
          ],
        ),
        const SizedBox(height: 10),

        // Rules table header
        if (isEnabled)
          Container(
            padding: EdgeInsets.all(8),
            color: global.theme.dividerBorderColor,
            child: Row(
              children: [
                // Add sequence number header
                SizedBox(
                  width: 30,
                  child: Text(
                    global.language("#"),
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                ),

                Expanded(
                  flex: 3,
                  child: Text(
                    global.language("lower_bound"),
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                ),
                Expanded(
                  flex: 3,
                  child: Text(
                    global.language("upper_bound"),
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                ),
                Expanded(
                  flex: 3,
                  child: Text(
                    global.language("round_to"),
                    style: const TextStyle(fontWeight: FontWeight.bold),
                  ),
                ),
                Expanded(flex: 1, child: Container()),
              ],
            ),
          ),

        // Rules list
        if (isEnabled)
          Expanded(
            child: ListView.builder(
              itemCount: roundingControllers[method]!.length,
              itemBuilder: (context, index) {
                return Padding(
                  padding: const EdgeInsets.symmetric(vertical: 4.0),
                  child: Row(
                    children: [
                      // Add sequence number column
                      SizedBox(
                        width: 30,
                        child: Padding(
                          padding: const EdgeInsets.all(8.0),
                          child: Text(
                            "${index + 1}",
                            style: const TextStyle(fontWeight: FontWeight.bold),
                          ),
                        ),
                      ),
                      const SizedBox(width: 8),

                      // Lower bound
                      Expanded(
                        flex: 3,
                        child: TextField(
                          controller: roundingControllers[method]![index][0],
                          decoration: InputDecoration(
                            border: const OutlineInputBorder(),
                            contentPadding: const EdgeInsets.symmetric(
                              horizontal: 8,
                            ),
                            hintText: global.language("lower_bound"),
                          ),
                          keyboardType: const TextInputType.numberWithOptions(
                            decimal: true,
                          ),
                          onChanged: (value) {
                            isDataChange = true;
                            double val = double.tryParse(value) ?? 0;

                            // Update the model
                            switch (method) {
                              case 'cash':
                                if (index <
                                    screenData
                                        .paymentrounding!
                                        .cash
                                        .rules
                                        .length) {
                                  screenData
                                          .paymentrounding!
                                          .cash
                                          .rules[index]
                                          .lowerbound =
                                      val;
                                } else {
                                  screenData.paymentrounding!.cash.rules.add(
                                    RoundingRuleModel(
                                      lowerbound: val,
                                      upperbound: 0,
                                      roundto: 0,
                                    ),
                                  );
                                }
                                break;
                              case 'creditcard':
                                if (index <
                                    screenData
                                        .paymentrounding!
                                        .creditcard
                                        .rules
                                        .length) {
                                  screenData
                                          .paymentrounding!
                                          .creditcard
                                          .rules[index]
                                          .lowerbound =
                                      val;
                                } else {
                                  screenData.paymentrounding!.creditcard.rules
                                      .add(
                                        RoundingRuleModel(
                                          lowerbound: val,
                                          upperbound: 0,
                                          roundto: 0,
                                        ),
                                      );
                                }
                                break;
                              case 'banktransfer':
                                if (index <
                                    screenData
                                        .paymentrounding!
                                        .banktransfer
                                        .rules
                                        .length) {
                                  screenData
                                          .paymentrounding!
                                          .banktransfer
                                          .rules[index]
                                          .lowerbound =
                                      val;
                                } else {
                                  screenData.paymentrounding!.banktransfer.rules
                                      .add(
                                        RoundingRuleModel(
                                          lowerbound: val,
                                          upperbound: 0,
                                          roundto: 0,
                                        ),
                                      );
                                }
                                break;
                              case 'cheque':
                                if (index <
                                    screenData
                                        .paymentrounding!
                                        .cheque
                                        .rules
                                        .length) {
                                  screenData
                                          .paymentrounding!
                                          .cheque
                                          .rules[index]
                                          .lowerbound =
                                      val;
                                } else {
                                  screenData.paymentrounding!.cheque.rules.add(
                                    RoundingRuleModel(
                                      lowerbound: val,
                                      upperbound: 0,
                                      roundto: 0,
                                    ),
                                  );
                                }
                                break;
                              case 'coupon':
                                if (index <
                                    screenData
                                        .paymentrounding!
                                        .coupon
                                        .rules
                                        .length) {
                                  screenData
                                          .paymentrounding!
                                          .coupon
                                          .rules[index]
                                          .lowerbound =
                                      val;
                                } else {
                                  screenData.paymentrounding!.coupon.rules.add(
                                    RoundingRuleModel(
                                      lowerbound: val,
                                      upperbound: 0,
                                      roundto: 0,
                                    ),
                                  );
                                }
                                break;
                              case 'delivery':
                                if (index <
                                    screenData
                                        .paymentrounding!
                                        .delivery
                                        .rules
                                        .length) {
                                  screenData
                                          .paymentrounding!
                                          .delivery
                                          .rules[index]
                                          .lowerbound =
                                      val;
                                } else {
                                  screenData.paymentrounding!.delivery.rules
                                      .add(
                                        RoundingRuleModel(
                                          lowerbound: val,
                                          upperbound: 0,
                                          roundto: 0,
                                        ),
                                      );
                                }
                                break;
                              case 'qrcode':
                                if (index <
                                    screenData
                                        .paymentrounding!
                                        .qrcode
                                        .rules
                                        .length) {
                                  screenData
                                          .paymentrounding!
                                          .qrcode
                                          .rules[index]
                                          .lowerbound =
                                      val;
                                } else {
                                  screenData.paymentrounding!.qrcode.rules.add(
                                    RoundingRuleModel(
                                      lowerbound: val,
                                      upperbound: 0,
                                      roundto: 0,
                                    ),
                                  );
                                }
                                break;
                            }
                          },
                        ),
                      ),
                      const SizedBox(width: 8),

                      // Upper bound
                      Expanded(
                        flex: 3,
                        child: TextField(
                          controller: roundingControllers[method]![index][1],
                          decoration: InputDecoration(
                            border: const OutlineInputBorder(),
                            contentPadding: const EdgeInsets.symmetric(
                              horizontal: 8,
                            ),
                            hintText: global.language("upper_bound"),
                          ),
                          keyboardType: const TextInputType.numberWithOptions(
                            decimal: true,
                          ),
                          onChanged: (value) {
                            isDataChange = true;
                            double val = double.tryParse(value) ?? 0;

                            // Update the model
                            switch (method) {
                              case 'cash':
                                if (index <
                                    screenData
                                        .paymentrounding!
                                        .cash
                                        .rules
                                        .length) {
                                  screenData
                                          .paymentrounding!
                                          .cash
                                          .rules[index]
                                          .upperbound =
                                      val;
                                }
                                break;
                              case 'creditcard':
                                if (index <
                                    screenData
                                        .paymentrounding!
                                        .creditcard
                                        .rules
                                        .length) {
                                  screenData
                                          .paymentrounding!
                                          .creditcard
                                          .rules[index]
                                          .upperbound =
                                      val;
                                }
                                break;
                              case 'banktransfer':
                                if (index <
                                    screenData
                                        .paymentrounding!
                                        .banktransfer
                                        .rules
                                        .length) {
                                  screenData
                                          .paymentrounding!
                                          .banktransfer
                                          .rules[index]
                                          .upperbound =
                                      val;
                                }
                                break;
                              case 'cheque':
                                if (index <
                                    screenData
                                        .paymentrounding!
                                        .cheque
                                        .rules
                                        .length) {
                                  screenData
                                          .paymentrounding!
                                          .cheque
                                          .rules[index]
                                          .upperbound =
                                      val;
                                }
                                break;
                              case 'coupon':
                                if (index <
                                    screenData
                                        .paymentrounding!
                                        .coupon
                                        .rules
                                        .length) {
                                  screenData
                                          .paymentrounding!
                                          .coupon
                                          .rules[index]
                                          .upperbound =
                                      val;
                                }
                                break;
                              case 'delivery':
                                if (index <
                                    screenData
                                        .paymentrounding!
                                        .delivery
                                        .rules
                                        .length) {
                                  screenData
                                          .paymentrounding!
                                          .delivery
                                          .rules[index]
                                          .upperbound =
                                      val;
                                }
                                break;
                              case 'qrcode':
                                if (index <
                                    screenData
                                        .paymentrounding!
                                        .qrcode
                                        .rules
                                        .length) {
                                  screenData
                                          .paymentrounding!
                                          .qrcode
                                          .rules[index]
                                          .upperbound =
                                      val;
                                }
                                break;
                            }
                          },
                        ),
                      ),
                      const SizedBox(width: 8),

                      // Round to
                      Expanded(
                        flex: 3,
                        child: TextField(
                          controller: roundingControllers[method]![index][2],
                          decoration: InputDecoration(
                            border: const OutlineInputBorder(),
                            contentPadding: const EdgeInsets.symmetric(
                              horizontal: 8,
                            ),
                            hintText: global.language("round_to"),
                          ),
                          keyboardType: const TextInputType.numberWithOptions(
                            decimal: true,
                          ),
                          onChanged: (value) {
                            isDataChange = true;
                            double val = double.tryParse(value) ?? 0;

                            // Update the model
                            switch (method) {
                              case 'cash':
                                if (index <
                                    screenData
                                        .paymentrounding!
                                        .cash
                                        .rules
                                        .length) {
                                  screenData
                                          .paymentrounding!
                                          .cash
                                          .rules[index]
                                          .roundto =
                                      val;
                                }
                                break;
                              case 'creditcard':
                                if (index <
                                    screenData
                                        .paymentrounding!
                                        .creditcard
                                        .rules
                                        .length) {
                                  screenData
                                          .paymentrounding!
                                          .creditcard
                                          .rules[index]
                                          .roundto =
                                      val;
                                }
                                break;
                              case 'banktransfer':
                                if (index <
                                    screenData
                                        .paymentrounding!
                                        .banktransfer
                                        .rules
                                        .length) {
                                  screenData
                                          .paymentrounding!
                                          .banktransfer
                                          .rules[index]
                                          .roundto =
                                      val;
                                }
                                break;
                              case 'cheque':
                                if (index <
                                    screenData
                                        .paymentrounding!
                                        .cheque
                                        .rules
                                        .length) {
                                  screenData
                                          .paymentrounding!
                                          .cheque
                                          .rules[index]
                                          .roundto =
                                      val;
                                }
                                break;
                              case 'coupon':
                                if (index <
                                    screenData
                                        .paymentrounding!
                                        .coupon
                                        .rules
                                        .length) {
                                  screenData
                                          .paymentrounding!
                                          .coupon
                                          .rules[index]
                                          .roundto =
                                      val;
                                }
                                break;
                              case 'delivery':
                                if (index <
                                    screenData
                                        .paymentrounding!
                                        .delivery
                                        .rules
                                        .length) {
                                  screenData
                                          .paymentrounding!
                                          .delivery
                                          .rules[index]
                                          .roundto =
                                      val;
                                }
                                break;
                              case 'qrcode':
                                if (index <
                                    screenData
                                        .paymentrounding!
                                        .qrcode
                                        .rules
                                        .length) {
                                  screenData
                                          .paymentrounding!
                                          .qrcode
                                          .rules[index]
                                          .roundto =
                                      val;
                                }
                                break;
                            }
                          },
                        ),
                      ),
                      const SizedBox(width: 8),

                      // Delete button
                      Expanded(
                        flex: 1,
                        child: IconButton(
                          icon: const Icon(Icons.delete, color: Colors.red),
                          onPressed: () => removeRoundingRule(method, index),
                        ),
                      ),
                    ],
                  ),
                );
              },
            ),
          ),

        // Add new rule button
        if (isEnabled)
          Padding(
            padding: EdgeInsets.only(top: 8.0),
            child: ElevatedButton.icon(
              icon: Icon(Icons.add),
              label: Text(global.language("add_rule")),
              onPressed: () => addNewRoundingRule(method),
            ),
          ),
      ],
    );
  }

  /// สร้าง Widget สำหรับเลือกสกุลเงินหลักของสาขา
  Widget _buildBaseCurrencySelector() {
    // ถ้ายังไม่มีสกุลเงิน ใช้ค่า default THB
    final currencies = _currencies.isEmpty
        ? [CurrencyModel(code: 'THB', name: 'บาท', symbol: '฿')]
        : _currencies;

    // ตรวจสอบว่า value ปัจจุบันมีอยู่ใน list หรือไม่
    final currentValue = screenData.baseCurrency ?? 'THB';
    final valueExists = currencies.any((c) => c.code == currentValue);
    final dropdownValue = valueExists ? currentValue : currencies.first.code;

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.monetization_on, color: Colors.amber.shade700, size: 20),
              SizedBox(width: 8),
              Text(
                global.language("base_currency"),
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: global.theme.textColor,
                ),
              ),
              if (_currenciesLoading) ...[
                const SizedBox(width: 8),
                const SizedBox(
                  width: 16,
                  height: 16,
                  child: CircularProgressIndicator(strokeWidth: 2),
                ),
              ],
            ],
          ),
          const SizedBox(height: 12),
          DropdownButtonFormField<String>(
            value: dropdownValue,
            decoration: InputDecoration(
              filled: true,
              fillColor: Colors.grey.shade50,
              contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(8),
                borderSide: BorderSide(color: Colors.grey.shade300),
              ),
              enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(8),
                borderSide: BorderSide(color: Colors.grey.shade300),
              ),
            ),
            items: currencies.map((currency) {
              return DropdownMenuItem<String>(
                value: currency.code,
                child: Text('${currency.symbol} ${currency.name} (${currency.code})'),
              );
            }).toList(),
            onChanged: isSaveAllow
                ? (value) {
                    setState(() {
                      isDataChange = true;
                      screenData.baseCurrency = value;
                    });
                  }
                : null,
          ),
        ],
      ),
    );
  }

  /// สร้าง Widget สำหรับเลือกภาษาที่ใช้ในสาขา
  Widget _buildBranchLanguageSelector() {
    // รายการภาษาที่รองรับ
    final languages = [
      {'code': 'th', 'name': 'ไทย', 'flag': '🇹🇭'},
      {'code': 'en', 'name': 'English', 'flag': '🇺🇸'},
      {'code': 'vi', 'name': 'Tiếng Việt', 'flag': '🇻🇳'},
      {'code': 'lo', 'name': 'ລາວ', 'flag': '🇱🇦'},
      {'code': 'km', 'name': 'ខ្មែរ', 'flag': '🇰🇭'},
      {'code': 'my', 'name': 'မြန်မာ', 'flag': '🇲🇲'},
      {'code': 'cn', 'name': '中文', 'flag': '🇨🇳'},
      {'code': 'ja', 'name': '日本語', 'flag': '🇯🇵'},
      {'code': 'ko', 'name': '한국어', 'flag': '🇰🇷'},
    ];

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.language, color: Colors.blue.shade700, size: 20),
              SizedBox(width: 8),
              Text(
                global.language("branch_language"),
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: global.theme.textColor,
                ),
              ),
            ],
          ),
          const SizedBox(height: 12),
          DropdownButtonFormField<String>(
            value: (screenData.language != null && screenData.language!.isNotEmpty)
                ? screenData.language
                : 'th',
            decoration: InputDecoration(
              filled: true,
              fillColor: Colors.grey.shade50,
              contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
              border: OutlineInputBorder(
                borderRadius: BorderRadius.circular(8),
                borderSide: BorderSide(color: Colors.grey.shade300),
              ),
              enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(8),
                borderSide: BorderSide(color: Colors.grey.shade300),
              ),
            ),
            items: languages.map((lang) {
              return DropdownMenuItem<String>(
                value: lang['code'],
                child: Text('${lang['flag']} ${lang['name']}'),
              );
            }).toList(),
            onChanged: isSaveAllow
                ? (value) {
                    setState(() {
                      isDataChange = true;
                      screenData.language = value;
                    });
                  }
                : null,
          ),
        ],
      ),
    );
  }

  /// รายการ timezone ทั่วโลก
  List<Map<String, String>> get _worldTimezones => [
    // UTC-12 to UTC-1
    {'label': '(UTC-12:00) Baker Island, Howland Island', 'offset': '-12:00', 'code': 'Etc/GMT+12'},
    {'label': '(UTC-11:00) American Samoa, Niue', 'offset': '-11:00', 'code': 'Pacific/Pago_Pago'},
    {'label': '(UTC-10:00) Hawaii, Honolulu', 'offset': '-10:00', 'code': 'Pacific/Honolulu'},
    {'label': '(UTC-09:30) Marquesas Islands', 'offset': '-09:30', 'code': 'Pacific/Marquesas'},
    {'label': '(UTC-09:00) Alaska, Anchorage', 'offset': '-09:00', 'code': 'America/Anchorage'},
    {'label': '(UTC-08:00) Los Angeles, San Francisco, Vancouver', 'offset': '-08:00', 'code': 'America/Los_Angeles'},
    {'label': '(UTC-07:00) Denver, Phoenix, Calgary', 'offset': '-07:00', 'code': 'America/Denver'},
    {'label': '(UTC-06:00) Chicago, Mexico City, Guatemala', 'offset': '-06:00', 'code': 'America/Chicago'},
    {'label': '(UTC-05:00) New York, Toronto, Bogota, Lima', 'offset': '-05:00', 'code': 'America/New_York'},
    {'label': '(UTC-04:00) Santiago, Caracas, Halifax', 'offset': '-04:00', 'code': 'America/Santiago'},
    {'label': '(UTC-03:30) Newfoundland', 'offset': '-03:30', 'code': 'America/St_Johns'},
    {'label': '(UTC-03:00) São Paulo, Buenos Aires, Montevideo', 'offset': '-03:00', 'code': 'America/Sao_Paulo'},
    {'label': '(UTC-02:00) Fernando de Noronha', 'offset': '-02:00', 'code': 'America/Noronha'},
    {'label': '(UTC-01:00) Azores, Cape Verde', 'offset': '-01:00', 'code': 'Atlantic/Azores'},
    // UTC
    {'label': '(UTC+00:00) London, Dublin, Lisbon, Casablanca', 'offset': '+00:00', 'code': 'Europe/London'},
    // UTC+1 to UTC+14
    {'label': '(UTC+01:00) Paris, Berlin, Rome, Madrid, Amsterdam', 'offset': '+01:00', 'code': 'Europe/Paris'},
    {'label': '(UTC+02:00) Cairo, Athens, Helsinki, Johannesburg', 'offset': '+02:00', 'code': 'Europe/Athens'},
    {'label': '(UTC+03:00) Moscow, Istanbul, Riyadh, Nairobi', 'offset': '+03:00', 'code': 'Europe/Moscow'},
    {'label': '(UTC+03:30) Tehran', 'offset': '+03:30', 'code': 'Asia/Tehran'},
    {'label': '(UTC+04:00) Dubai, Baku, Tbilisi, Yerevan', 'offset': '+04:00', 'code': 'Asia/Dubai'},
    {'label': '(UTC+04:30) Kabul', 'offset': '+04:30', 'code': 'Asia/Kabul'},
    {'label': '(UTC+05:00) Karachi, Tashkent, Yekaterinburg', 'offset': '+05:00', 'code': 'Asia/Karachi'},
    {'label': '(UTC+05:30) Mumbai, Kolkata, New Delhi, Chennai', 'offset': '+05:30', 'code': 'Asia/Kolkata'},
    {'label': '(UTC+05:45) Kathmandu', 'offset': '+05:45', 'code': 'Asia/Kathmandu'},
    {'label': '(UTC+06:00) Dhaka, Almaty, Omsk', 'offset': '+06:00', 'code': 'Asia/Dhaka'},
    {'label': '(UTC+06:30) Yangon, Cocos Islands', 'offset': '+06:30', 'code': 'Asia/Yangon'},
    {'label': '(UTC+07:00) Bangkok, Hanoi, Jakarta, Ho Chi Minh', 'offset': '+07:00', 'code': 'Asia/Bangkok'},
    {'label': '(UTC+08:00) Singapore, Kuala Lumpur, Hong Kong, Perth', 'offset': '+08:00', 'code': 'Asia/Singapore'},
    {'label': '(UTC+08:00) Beijing, Shanghai, Taipei', 'offset': '+08:00', 'code': 'Asia/Shanghai'},
    {'label': '(UTC+08:00) Manila, Philippines', 'offset': '+08:00', 'code': 'Asia/Manila'},
    {'label': '(UTC+08:45) Eucla', 'offset': '+08:45', 'code': 'Australia/Eucla'},
    {'label': '(UTC+09:00) Tokyo, Seoul, Pyongyang', 'offset': '+09:00', 'code': 'Asia/Tokyo'},
    {'label': '(UTC+09:30) Adelaide, Darwin', 'offset': '+09:30', 'code': 'Australia/Adelaide'},
    {'label': '(UTC+10:00) Sydney, Melbourne, Brisbane, Guam', 'offset': '+10:00', 'code': 'Australia/Sydney'},
    {'label': '(UTC+10:30) Lord Howe Island', 'offset': '+10:30', 'code': 'Australia/Lord_Howe'},
    {'label': '(UTC+11:00) Solomon Islands, New Caledonia', 'offset': '+11:00', 'code': 'Pacific/Guadalcanal'},
    {'label': '(UTC+12:00) Auckland, Fiji, Kamchatka', 'offset': '+12:00', 'code': 'Pacific/Auckland'},
    {'label': '(UTC+12:45) Chatham Islands', 'offset': '+12:45', 'code': 'Pacific/Chatham'},
    {'label': '(UTC+13:00) Nuku\'alofa, Samoa', 'offset': '+13:00', 'code': 'Pacific/Tongatapu'},
    {'label': '(UTC+14:00) Line Islands', 'offset': '+14:00', 'code': 'Pacific/Kiritimati'},
  ];

  /// สร้าง Widget สำหรับเลือก Timezone ของสาขา (ค้นหาได้)
  Widget _buildTimezoneSelector() {
    final timezones = _worldTimezones;

    // หา timezone ปัจจุบัน
    final currentTimezone = screenData.timezone ?? 'Asia/Bangkok';
    final currentTz = timezones.firstWhere(
      (tz) => tz['code'] == currentTimezone,
      orElse: () => timezones.firstWhere((tz) => tz['code'] == 'Asia/Bangkok'),
    );

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.access_time, color: Colors.teal.shade700, size: 20),
              SizedBox(width: 8),
              Text(
                global.language("timezone"),
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: global.theme.textColor,
                ),
              ),
            ],
          ),
          SizedBox(height: 12),
          DropdownSearch<Map<String, String>>(
            items: (filter, infiniteScrollProps) => timezones,
            selectedItem: currentTz,
            enabled: isSaveAllow,
            compareFn: (item1, item2) => item1['code'] == item2['code'],
            itemAsString: (item) => item['label'] ?? '',
            filterFn: (item, filter) {
              final label = item['label']?.toLowerCase() ?? '';
              final code = item['code']?.toLowerCase() ?? '';
              final search = filter.toLowerCase();
              return label.contains(search) || code.contains(search);
            },
            popupProps: PopupProps.menu(
              showSearchBox: true,
              searchFieldProps: TextFieldProps(
                decoration: InputDecoration(
                  hintText: global.language("search"),
                  prefixIcon: const Icon(Icons.search),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                  ),
                ),
              ),
              constraints: const BoxConstraints(maxHeight: 350),
              itemBuilder: (context, item, isDisabled, isSelected) {
                return ListTile(
                  title: Text(
                    item['label'] ?? '',
                    style: TextStyle(
                      fontSize: 13,
                      fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
                      color: isSelected ? Colors.teal.shade700 : Colors.black87,
                    ),
                  ),
                  dense: true,
                  selected: isSelected,
                );
              },
            ),
            decoratorProps: DropDownDecoratorProps(
              decoration: InputDecoration(
                filled: true,
                fillColor: Colors.grey.shade50,
                contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                border: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: BorderSide(color: Colors.grey.shade300),
                ),
                enabledBorder: OutlineInputBorder(
                  borderRadius: BorderRadius.circular(8),
                  borderSide: BorderSide(color: Colors.grey.shade300),
                ),
              ),
            ),
            onChanged: (value) {
              if (value != null) {
                setState(() {
                  isDataChange = true;
                  screenData.timezone = value['code'];
                  screenData.timezonelabel = value['label'];
                  screenData.timezoneoffset = value['offset'];
                });
              }
            },
          ),
        ],
      ),
    );
  }

  /// สร้าง Widget สำหรับเลือกประเภทปี (พ.ศ. หรือ ค.ศ.)
  Widget _buildYeartypeSelector() {
    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.calendar_today, color: Colors.purple.shade700, size: 20),
              SizedBox(width: 8),
              Text(
                global.language("yeartype"),
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: global.theme.textColor,
                ),
              ),
            ],
          ),
          SizedBox(height: 12),
          Row(
            children: [
              Expanded(
                child: RadioListTile<String>(
                  title: Text(global.language("christian_era")),
                  subtitle: Text('${global.language("christian_era_short")} (2026)'),
                  value: 'christian',
                  groupValue: screenData.yeartype ?? 'christian',
                  onChanged: isSaveAllow
                      ? (value) {
                          setState(() {
                            isDataChange = true;
                            screenData.yeartype = value;
                          });
                        }
                      : null,
                  contentPadding: EdgeInsets.zero,
                  dense: true,
                ),
              ),
              Expanded(
                child: RadioListTile<String>(
                  title: Text(global.language("buddhist_era")),
                  subtitle: Text('${global.language("buddhist_era_short")} (2569)'),
                  value: 'buddhist',
                  groupValue: screenData.yeartype ?? 'christian',
                  onChanged: isSaveAllow
                      ? (value) {
                          setState(() {
                            isDataChange = true;
                            screenData.yeartype = value;
                          });
                        }
                      : null,
                  contentPadding: EdgeInsets.zero,
                  dense: true,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  /// สร้าง Widget สำหรับตั้งค่าทศนิยม (Decimal Places)
  Widget _buildDecimalSettingsSection() {
    final decimalOptions = List.generate(7, (i) => i + 2); // 2-8

    Widget buildDecimalDropdown({
      required String label,
      required int currentValue,
      required ValueChanged<int?> onChanged,
    }) {
      return Padding(
        padding: EdgeInsets.only(bottom: 12),
        child: Row(
          children: [
            Expanded(
              flex: 3,
              child: Text(
                label,
                style: TextStyle(fontSize: 13, color: global.theme.textColor),
              ),
            ),
            Expanded(
              flex: 2,
              child: DropdownButtonFormField<int>(
                value: currentValue,
                decoration: InputDecoration(
                  filled: true,
                  fillColor: Colors.grey.shade50,
                  contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                  border: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: BorderSide(color: Colors.grey.shade300),
                  ),
                  enabledBorder: OutlineInputBorder(
                    borderRadius: BorderRadius.circular(8),
                    borderSide: BorderSide(color: Colors.grey.shade300),
                  ),
                ),
                items: decimalOptions.map((v) {
                  return DropdownMenuItem<int>(
                    value: v,
                    child: Text('$v ${global.language("decimal_places")}'),
                  );
                }).toList(),
                onChanged: isSaveAllow ? onChanged : null,
              ),
            ),
          ],
        ),
      );
    }

    return Container(
      margin: const EdgeInsets.only(bottom: 16),
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: Colors.grey.shade200),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.numbers, color: Colors.teal.shade700, size: 20),
              SizedBox(width: 8),
              Text(
                global.language("decimal_settings"),
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: global.theme.textColor,
                ),
              ),
            ],
          ),
          SizedBox(height: 12),
          buildDecimalDropdown(
            label: global.language("decimal_quantity"),
            currentValue: (screenData.decimalQuantity != null && screenData.decimalQuantity! >= 2) ? screenData.decimalQuantity! : 2,
            onChanged: (value) {
              setState(() {
                isDataChange = true;
                screenData.decimalQuantity = value;
              });
            },
          ),
          buildDecimalDropdown(
            label: global.language("decimal_price"),
            currentValue: (screenData.decimalPrice != null && screenData.decimalPrice! >= 2) ? screenData.decimalPrice! : 2,
            onChanged: (value) {
              setState(() {
                isDataChange = true;
                screenData.decimalPrice = value;
              });
            },
          ),
          buildDecimalDropdown(
            label: global.language("decimal_document"),
            currentValue: (screenData.decimalDocument != null && screenData.decimalDocument! >= 2) ? screenData.decimalDocument! : 2,
            onChanged: (value) {
              setState(() {
                isDataChange = true;
                screenData.decimalDocument = value;
              });
            },
          ),
        ],
      ),
    );
  }

  Widget editScreen({mobileScreen}) {
    List<Widget> formWidgets = [];
    List<Widget> formWidgetsExpansion = [];

    focusNodeMax = 0;
    int currentFocusNodeMax = focusNodeMax;

    // Company Info Form Widget
    formWidgets.add(
      CompanyInfoFormWidget(
        languageList: languageList,
        companyNames: screenData.companynames,
        isEditMode: isEditMode,
        disableBranch: disableBranch,
        onCompanyNameChanged: (languageIndex, value) {
          setState(() {
            isDataChange = true;
            // Ensure the index exists
            while (screenData.companynames.length <= languageIndex) {
              screenData.companynames.add(
                LanguageDataModel(
                  code: languageList[screenData.companynames.length].code!,
                  name: '',
                ),
              );
            }
            screenData.companynames[languageIndex].name = value;
          });
        },
        findFocusNext: (index) {
          if (kIsWeb) {
            findFocusNext(index);
          }
        },
        fieldFocusNodes: fieldFocusNodes,
        focusNodeMax: focusNodeMax,
      ),
    );

    // Update focusNodeMax after company names
    currentFocusNodeMax = focusNodeMax + languageList.length;

    // Branch Info Form Widget (includes branch code, names, addresses, location, phone)
    focusNodeMax = currentFocusNodeMax;
    int branchInfoStartIndex = focusNodeMax + 1;

    formWidgets.add(
      BranchInfoFormWidget(
        languageList: languageList,
        branchCodeController: branchCode,
        branchNames: screenData.names,
        contact: screenData.contact,
        isEditMode: isEditMode,
        disableBranch: disableBranch,
        onBranchCodeChanged: (value) {
          setState(() {
            isDataChange = true;
            screenData.code = value.toUpperCase();
            branchCode.value = TextEditingValue(
              text: value.toUpperCase(),
              selection: branchCode.selection,
            );
          });
        },
        onBranchNameChanged: (languageIndex, value) {
          setState(() {
            isDataChange = true;
            // Ensure the index exists
            while (screenData.names.length <= languageIndex) {
              screenData.names.add(
                LanguageDataModel(
                  code: languageList[screenData.names.length].code!,
                  name: '',
                ),
              );
            }
            screenData.names[languageIndex].name = value;
          });
        },
        onAddressChanged: (languageIndex, value) {
          setState(() {
            isDataChange = true;
            // Ensure the index exists
            while (screenData.contact!.address!.length <= languageIndex) {
              screenData.contact!.address!.add(
                LanguageDataModel(
                  code: languageList[screenData.contact!.address!.length].code!,
                  name: '',
                ),
              );
            }
            screenData.contact!.address![languageIndex].name = value;
          });
        },
        onLatitudeChanged: (value) {
          setState(() {
            isDataChange = true;
            screenData.contact?.latitude = double.tryParse(value) ?? 0;
          });
        },
        onLongitudeChanged: (value) {
          setState(() {
            isDataChange = true;
            screenData.contact?.longitude = double.tryParse(value) ?? 0;
          });
        },
        onPhoneChanged: (value) {
          setState(() {
            isDataChange = true;
            screenData.contact?.phonenumber = value;
          });
        },
        onMapLocationPressed: () async {
          final result = await Navigator.push(
            context,
            MaterialPageRoute(
              builder: (context) => MapGetLocationScreen(
                latitude: screenData.contact!.latitude!,
                longitude: screenData.contact!.longitude!,
              ),
            ),
          );
          if (result != null) {
            setState(() {
              isDataChange = true;
              screenData.contact!.latitude = result.latitude;
              screenData.contact!.longitude = result.longitude;
            });
          }
        },
        findFocusNext: (index) {
          if (kIsWeb) {
            findFocusNext(index);
          }
        },
        fieldFocusNodes: fieldFocusNodes,
        initialFocusNodeIndex: branchInfoStartIndex,
      ),
    );

    // Update focusNodeMax after branch info
    focusNodeMax = branchInfoStartIndex + 1 + (languageList.length * 3) + 3;

    // Tax ID Widget
    formWidgets.add(
      TaxIdWidget(
        taxId: screenData.pos!.taxid,
        disableBranch: disableBranch,
        isEditMode: isEditMode,
        onChanged: (value) {
          setState(() {
            isDataChange = true;
            screenData.pos!.taxid = value;
          });
        },
        onSubmitted: () {
          if (kIsWeb) {
            findFocusNext(focusNodeIndex);
          }
        },
        focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
      ),
    );

    // Business Type Selector Widget
    formWidgets.add(
      BusinessTypeSelectorWidget(
        businessType: screenData.businesstype,
        businessTypeNull: businesstypeNull,
        onBusinessTypeSelected: (BusinessTypeModel result) {
          setState(() {
            businesstypeNull = false;
            isDataChange = true;
            screenData.businesstype!.guidfixed = result.guidfixed;
            screenData.businesstype!.code = result.code;
            screenData.businesstype!.names = result.names;
          });
        },
        onBusinessTypeCleared: () {
          setState(() {
            isDataChange = true;
            screenData.businesstype!.guidfixed = '';
            screenData.businesstype!.code = '';
            screenData.businesstype!.names = [];
          });
        },
      ),
    );

    // คุณสมบัติธุรกิจ
    formWidgets.add(
      BusinessPropertyWidget(
        isRestaurant: screenData.isRestaurant ?? false,
        isTire: screenData.isTire ?? false,
        isAgriculture: screenData.isAgriculture ?? false,
        isPharmacy: screenData.isPharmacy ?? false,
        onRestaurantChanged: (v) {
          setState(() { isDataChange = true; screenData.isRestaurant = v; });
        },
        onTireChanged: (v) {
          setState(() { isDataChange = true; screenData.isTire = v; });
        },
        onAgricultureChanged: (v) {
          setState(() { isDataChange = true; screenData.isAgriculture = v; });
        },
        onPharmacyChanged: (v) {
          setState(() { isDataChange = true; screenData.isPharmacy = v; });
        },
      ),
    );

    // Language Selector Widget
    formWidgetsExpansion.add(
      LanguageSelectorWidget(
        selectedLanguages: selectlanguageList,
        defaultLanguageList: defaultlanguageList,
        isSaveAllow: isSaveAllow,
        onLanguagesChanged: (List<String> newList) {
          setState(() {
            selectlanguageList = newList;
          });
        },
      ),
    );

    // Base Currency Selector Widget
    formWidgetsExpansion.add(
      _buildBaseCurrencySelector(),
    );

    // Branch Language Selector Widget
    formWidgetsExpansion.add(
      _buildBranchLanguageSelector(),
    );

    // Timezone Selector Widget
    formWidgetsExpansion.add(
      _buildTimezoneSelector(),
    );

    // Yeartype Selector Widget (พ.ศ. / ค.ศ.)
    formWidgetsExpansion.add(
      _buildYeartypeSelector(),
    );

    // Decimal Settings Widget (ตั้งค่าทศนิยม)
    formWidgetsExpansion.add(
      _buildDecimalSettingsSection(),
    );

    // VAT Config Widget
    formWidgetsExpansion.add(
      VatConfigWidget(
        posData: screenData.pos,
        onDataChanged: (value) {
          setState(() {
            isDataChange = value;
          });
        },
      ),
    );

    // Payment Rounding Widget
    formWidgetsExpansion.add(
      PaymentRoundingWidget(
        paymentRounding: screenData.paymentrounding,
        roundingControllers: roundingControllers,
        onRemoveRule: removeRoundingRule,
        onAddRule: addNewRoundingRule,
        onDataChanged: (value) {
          setState(() {
            isDataChange = value;
          });
        },
      ),
    );

    // Coupon Usage Type Widget
    formWidgetsExpansion.add(
      CouponUsageTypeWidget(
        couponUseType: screenData.couponusetype,
        onChanged: (value) {
          setState(() {
            isDataChange = true;
            screenData.couponusetype = value;
          });
        },
      ),
    );

    // Receipt Config Widget
    formWidgetsExpansion.add(
      ReceiptConfigWidget(
        posData: screenData.pos,
        onDataChanged: (value) {
          setState(() {
            isDataChange = value;
          });
        },
      ),
    );

    // Image Picker Widget (Branch Image)
    formWidgetsExpansion.add(
      ImagePickerWidget(
        title: global.language("image"),
        imageFiles: imageFile,
        imageWeb: imageWeb,
        imageUri: screenData.imageuri,
        enabled: disableBranch(),
        onUpload: () {
          setState(() {
            upLoadImage();
          });
        },
        picker: imagePicker,
      ),
    );

    // Image Picker Widget (Logo)
    formWidgetsExpansion.add(
      ImagePickerWidget(
        title: global.language("logo"),
        imageFiles: logoFile,
        imageWeb: logoWeb,
        imageUri: screenData.logouri,
        enabled: disableBranch(),
        onUpload: () {
          setState(() {
            upLoadLogo();
          });
        },
        picker: logoPicker,
      ),
    );

    formWidgets.add(const SizedBox(height: 10));

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
        title: Text(headerEdit + global.language("company_branch")),
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
                            context.read<CompanyBranchBloc>().add(
                              CompanyBranchDelete(guid: selectGuid),
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
                  for (
                    int indexLanguage = 2;
                    indexLanguage <= languageList.length;
                    indexLanguage++
                  ) {
                    try {} catch (_) {}
                  }
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
      body: RawKeyboardListener(
        focusNode: FocusNode(),
        onKey: (RawKeyEvent event) {
          if (event is RawKeyDownEvent) {
            // print(event.logicalKey);
            if (event.logicalKey == LogicalKeyboardKey.f10) {
              saveOrUpdateData();
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
            padding: const EdgeInsets.only(top: 10, bottom: 10),
            child: Form(
              child: Column(
                children: [
                  Column(children: formWidgets),
                  const SizedBox(height: 10),
                  // กำหนดค่าเพิ่มเติม - แสดงตลอดไม่มีปุ่ม expand
                  Container(
                    decoration: BoxDecoration(
                      color: Colors.grey[50],
                      borderRadius: BorderRadius.circular(12),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        // Header
                        Padding(
                          padding: EdgeInsets.all(16),
                          child: Row(
                            children: [
                              Container(
                                padding: EdgeInsets.all(6),
                                decoration: BoxDecoration(
                                  color: Colors.blue.withValues(alpha: 0.1),
                                  borderRadius: BorderRadius.circular(6),
                                ),
                                child: Icon(
                                  Icons.settings,
                                  color: global.theme.primaryColor,
                                  size: 20,
                                ),
                              ),
                              SizedBox(width: 12),
                              Text(
                                global.language("advanced_settings"),
                                style: TextStyle(
                                  fontSize: 16,
                                  fontWeight: FontWeight.w600,
                                  color: global.theme.textColor,
                                ),
                              ),
                            ],
                          ),
                        ),
                        // Content
                        Padding(
                          padding: const EdgeInsets.fromLTRB(16, 0, 16, 16),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: formWidgetsExpansion,
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
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
              BlocListener<CompanyBranchBloc, CompanyBranchState>(
                listener: (context, state) {
                  blocCurrentState = state;
                  // Load
                  if (state is CompanyBranchLoadSuccess) {
                    setState(() {
                      loadingData = false;
                      if (state.companyBranch.isNotEmpty) {
                        listData.addAll(state.companyBranch);
                        // สร้าง keys ใหม่ให้ตรงกับจำนวน items
                        listKeys.clear();
                        for (int i = 0; i < listData.length; i++) {
                          listKeys.add(GlobalKey());
                        }
                      }
                    });
                  }
                  if (state is CompanyBranchLoadFailed) {
                    setState(() {
                      loadingData = false;
                    });
                  }
                  // Save
                  if (state is CompanyBranchSaveSuccess) {
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
                  if (state is CompanyBranchSaveFailed) {
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
                  // Delete
                  if (state is CompanyBranchDeleteSuccess) {
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
                  if (state is CompanyBranchDeleteManySuccess) {
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

  // Add the missing roundingControllers property
  Map<String, List<List<TextEditingController>>> roundingControllers = {};
  void initPaymentRoundingControllers() {
    // Initialize controllers for all payment methods
    List<String> paymentMethods = [
      'cash',
      'creditcard',
      'banktransfer',
      'cheque',
      'coupon',
      'delivery',
      'qrcode',
    ];

    List<Map<String, double>> defaultRules = [
      {'lowerbound': 0.01, 'upperbound': 0.12, 'roundto': 0.0},
      {'lowerbound': 0.13, 'upperbound': 0.37, 'roundto': 0.25},
      {'lowerbound': 0.38, 'upperbound': 0.62, 'roundto': 0.5},
      {'lowerbound': 0.63, 'upperbound': 0.87, 'roundto': 0.75},
      {'lowerbound': 0.88, 'upperbound': 0.99, 'roundto': 1.0},
    ];

    for (String method in paymentMethods) {
      roundingControllers[method] = [];

      // Initialize the model if needed
      if (screenData.paymentrounding != null) {
        // Make sure the rules list is empty before adding default rules
        switch (method) {
          case 'cash':
            screenData.paymentrounding!.cash.rules.clear();
            for (var rule in defaultRules) {
              screenData.paymentrounding!.cash.rules.add(
                RoundingRuleModel(
                  lowerbound: rule['lowerbound']!,
                  upperbound: rule['upperbound']!,
                  roundto: rule['roundto']!,
                ),
              );
            }
            break;
          case 'creditcard':
            screenData.paymentrounding!.creditcard.rules.clear();
            for (var rule in defaultRules) {
              screenData.paymentrounding!.creditcard.rules.add(
                RoundingRuleModel(
                  lowerbound: rule['lowerbound']!,
                  upperbound: rule['upperbound']!,
                  roundto: rule['roundto']!,
                ),
              );
            }
            break;
          case 'banktransfer':
            screenData.paymentrounding!.banktransfer.rules.clear();
            for (var rule in defaultRules) {
              screenData.paymentrounding!.banktransfer.rules.add(
                RoundingRuleModel(
                  lowerbound: rule['lowerbound']!,
                  upperbound: rule['upperbound']!,
                  roundto: rule['roundto']!,
                ),
              );
            }
            break;
          case 'cheque':
            screenData.paymentrounding!.cheque.rules.clear();
            for (var rule in defaultRules) {
              screenData.paymentrounding!.cheque.rules.add(
                RoundingRuleModel(
                  lowerbound: rule['lowerbound']!,
                  upperbound: rule['upperbound']!,
                  roundto: rule['roundto']!,
                ),
              );
            }
            break;
          case 'coupon':
            screenData.paymentrounding!.coupon.rules.clear();
            for (var rule in defaultRules) {
              screenData.paymentrounding!.coupon.rules.add(
                RoundingRuleModel(
                  lowerbound: rule['lowerbound']!,
                  upperbound: rule['upperbound']!,
                  roundto: rule['roundto']!,
                ),
              );
            }
            break;
          case 'delivery':
            screenData.paymentrounding!.delivery.rules.clear();
            for (var rule in defaultRules) {
              screenData.paymentrounding!.delivery.rules.add(
                RoundingRuleModel(
                  lowerbound: rule['lowerbound']!,
                  upperbound: rule['upperbound']!,
                  roundto: rule['roundto']!,
                ),
              );
            }
            break;
          case 'qrcode':
            screenData.paymentrounding!.qrcode.rules.clear();
            for (var rule in defaultRules) {
              screenData.paymentrounding!.qrcode.rules.add(
                RoundingRuleModel(
                  lowerbound: rule['lowerbound']!,
                  upperbound: rule['upperbound']!,
                  roundto: rule['roundto']!,
                ),
              );
            }
            break;
        }
      }

      // For each default rule, create controllers
      for (var rule in defaultRules) {
        List<TextEditingController> controllers = [
          TextEditingController(text: rule['lowerbound']!.toString()),
          TextEditingController(text: rule['upperbound']!.toString()),
          TextEditingController(text: rule['roundto']!.toString()),
        ];
        roundingControllers[method]!.add(controllers);
      }
    }
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
