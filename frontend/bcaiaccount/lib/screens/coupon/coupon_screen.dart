import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:font_awesome_flutter/font_awesome_flutter.dart';
import 'package:file_picker/file_picker.dart';

import 'package:split_view/split_view.dart';
import 'package:translator/translator.dart';
import '../../bloc/coupon/coupon_bloc.dart';
import '../../bloc/coupon/coupon_event.dart';
import '../../bloc/coupon/coupon_state.dart';
import '../../model/coupon_model.dart';
import '../../model/global_model.dart';
import '../../model/product_model.dart';
import '../../model/coupon_import_model.dart';
import '../../screen_search/customer_search_screen.dart';
import '../../screen_search/barcode_search_screen.dart';
import '../../screen_search/product_master_group_search_screen.dart';
import '../../screen_search/product_group_sub1_search_screen.dart';
import '../../screen_search/product_group_sub2_search_screen.dart';
import '../../screen_search/product_brand_search_screen.dart';
import '../../screen_search/product_design_search_screen.dart';
import '../../screen_search/product_model_search_screen.dart';
import '../../screen_search/product_pattern_search_screen.dart';
import '../../screen_search/product_grade_search_screen.dart';
import '../../screen_search/product_category_all_search_screen.dart';
import '../../screen_search/product_class_search_screen.dart';
import '../../global.dart' as global;
import '../../widgets/edit_font_size_control.dart';
import '../../components/loading_overlay.dart';
import '../../screens/report/file_download.dart';

// Import component widgets
import 'components/coupon_basic_info_widget.dart';
import 'components/coupon_type_value_widget.dart';
import 'components/coupon_conditions_widget.dart';
import 'components/coupon_status_widget.dart';
import 'components/coupon_customer_selection_widget.dart';
import 'components/coupon_list_item_widget.dart';
import 'components/coupon_remark_widget.dart';
import 'components/coupon_product_condition_widget.dart';
import '../../model/product_condition_model.dart';

// Import helpers
import 'helpers/coupon_validator.dart';
import 'helpers/coupon_error_handler.dart';
import 'helpers/datetime_extensions.dart';

class CouponScreen extends StatefulWidget {
  const CouponScreen({super.key});

  @override
  State<CouponScreen> createState() => CouponScreenState();
}

class CouponScreenState extends State<CouponScreen>
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
  List<CouponModel> listData = [];
  List<String> couponGuidListChecked = [];
  List<LanguageDataModel> names = [];
  List<LanguageDataModel> conditionNames = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  int _hoverIndex = -1;
  bool isChange = false;
  bool isSaveAllow = false;
  late CouponState blocCouponState;
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  global.ScreenEventEnum screenEvent = global.ScreenEventEnum.list;
  late SplitViewController splitViewController;
  final debouncer = global.Debouncer(1000);
  bool isLoadTranslation = false;

  // Coupon specific fields - additional to multi-language fields
  late TextEditingController couponCodeController;
  late TextEditingController couponValueController;
  late TextEditingController remarkController;
  late TextEditingController maxUsageCountController;
  late TextEditingController maxUsageCountPerCustomerController;
  late TextEditingController customerCodesController;

  late FocusNode couponCodeFocusNode;
  late FocusNode couponValueFocusNode;
  late FocusNode remarkFocusNode;
  late FocusNode maxUsageCountFocusNode;
  late FocusNode maxUsageCountPerCustomerFocusNode;
  late FocusNode customerCodesFocusNode;

  DateTime? issuedDate = CouponDateHelper.getDefaultIssuedDate();
  DateTime? expiryDate = CouponDateHelper.getDefaultExpiryDate();
  int couponType = 0;
  int status = 0;
  bool isOneTimeUse = false;
  List<String> customerCodes = [];
  ProductConditionModel? productCondition;

  // Error states for validation
  Map<String, String> fieldErrors = {};

  // Loading state
  bool isLoading = false;

  // Import progress tracking
  String? currentBatchId;
  bool isImportDialogOpen = false;

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

    // Add fieldTextController and focusNodes for each language
    for (int i = 0; i < languageList.length; i++) {
      fieldTextController.add(TextEditingController());
      FocusNode focusNode = FocusNode();
      int actualIndex = i + 1; // +1 because coupon code is at index 0
      focusNode.addListener(() {
        focusNodeIndex = actualIndex;
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

    // Initialize coupon specific controllers and focus nodes
    couponCodeController = TextEditingController();
    couponValueController = TextEditingController();
    remarkController = TextEditingController();
    maxUsageCountController = TextEditingController();
    maxUsageCountPerCustomerController = TextEditingController();
    customerCodesController = TextEditingController();

    couponCodeFocusNode = FocusNode();
    couponValueFocusNode = FocusNode();
    remarkFocusNode = FocusNode();
    maxUsageCountFocusNode = FocusNode();
    maxUsageCountPerCustomerFocusNode = FocusNode();
    customerCodesFocusNode = FocusNode();

    // Add coupon code field as first field (similar to brand code)
    FocusNode focusNode = FocusNode();
    focusNode.addListener(() {
      focusNodeIndex = 0;
    });
    fieldFocusNodes.add(global.FieldFocusModel(focusNode: focusNode));
    fieldTextController.add(TextEditingController());

    setSystemLanguageList();
    listScrollController.addListener(onScrollList);

    super.initState();
  }

  void loadDataList(String search) {
    context.read<CouponBloc>().add(GetCoupons());
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

    // Dispose coupon specific controllers
    couponCodeController.dispose();
    couponValueController.dispose();
    remarkController.dispose();
    maxUsageCountController.dispose();
    maxUsageCountPerCustomerController.dispose();
    customerCodesController.dispose();

    // Dispose coupon specific focus nodes
    couponCodeFocusNode.dispose();
    couponValueFocusNode.dispose();
    remarkFocusNode.dispose();
    maxUsageCountFocusNode.dispose();
    maxUsageCountPerCustomerFocusNode.dispose();
    customerCodesFocusNode.dispose();

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
    couponCodeController.clear();
    couponValueController.clear();
    remarkController.clear();
    maxUsageCountController.clear();
    maxUsageCountPerCustomerController.clear();

    issuedDate = CouponDateHelper.getDefaultIssuedDate();
    expiryDate = CouponDateHelper.getDefaultExpiryDate();
    couponType = 0;
    status = 0; // Default to active (0 = เปิด)
    isOneTimeUse = false;
    customerCodes.clear();
    productCondition = null;

    names.clear();
    conditionNames.clear();
    fieldErrors.clear(); // Clear validation errors

    isChange = false;
    focusNodeIndex = 0;
    if (fieldFocusNodes.isNotEmpty) {
      fieldFocusNodes[focusNodeIndex].focusNode.requestFocus();
    }
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
    context.read<CouponBloc>().add(GetCouponById(id: guid));
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('coupon')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: Icon(Icons.arrow_back),
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
          IconButton(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () {
              downloadCouponTemplate();
            },
            icon: Icon(Icons.download),
            tooltip: global.language('download_excel_template'),
          ),
          IconButton(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () {
              selectAndUploadExcelFile();
            },
            icon: Icon(Icons.file_upload),
            tooltip: global.language('import_excel'),
          ),
          IconButton(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () {
              discardData(
                callBack: () {
                  setState(() {
                    changeScreenEvent(global.ScreenEventEnum.add);
                    selectGuid = "";
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
            icon: Icon(Icons.add),
          ),
        ],
      ),
      body: Column(
        children: [
          Container(
            padding: const EdgeInsets.all(5),
            decoration: BoxDecoration(
              color: global.theme.cardColor,
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
                IconButton(
                  focusNode: FocusNode(skipTraversal: true),
                  icon: const FaIcon(FontAwesomeIcons.font),
                  onPressed: () async {
                    setState(() {
                      global.listDataFontSizeChange();
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
                    global.language("coupon_code"),
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
                    global.language("coupon_name"),
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
                  flex: 3,
                  child: Text(
                    global.language("coupon_type"),
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
                  flex: 2,
                  child: Text(
                    global.language("status"),
                    style: TextStyle(
                      color: global.theme.columnHeaderTextColor,
                      fontWeight: FontWeight.bold,
                      fontSize: global.deviceConfig.listDataFontSize + 2,
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ],
            ),
          ),
          Expanded(
            child: ListView(
              controller: listScrollController,
              children: listData
                  .map((value) => listObject(listData.indexOf(value), value))
                  .toList(),
            ),
          ),
        ],
      ),
    );
  }

  void switchToEdit(CouponModel value) {
    setState(() {
      selectGuid = value.guidfixed ?? "";
      getData(selectGuid);
      headerEdit = global.language("edit");
      isSaveAllow = true;
      changeScreenEvent(global.ScreenEventEnum.edit);
    });
  }


  Color _getContainerColor(String itemGuid, int index) {
    if (selectGuid.isNotEmpty && selectGuid == itemGuid) {
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

  Widget listObject(int index, CouponModel value) {
    return CouponListItemWidget(
      coupon: value,
      isSelected: selectGuid == (value.guidfixed ?? ""),
      fontSize: global.deviceConfig.listDataFontSize,
      lineSpace: global.deviceConfig.listDataLineSpace,
      onTap: () {
        setState(() {
          discardData(
            callBack: () {
              isSaveAllow = false;
              changeScreenEvent(global.ScreenEventEnum.list);
              selectGuid = value.guidfixed ?? "";
              getData(selectGuid);
              WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                tabController.animateTo(1);
              });
            },
          );
        });
      },
      onDoubleTap: () {
        switchToEdit(value);
      },
    );
  }

  List<LanguageDataModel> packLanguage(
    List<TextEditingController> controllers,
  ) {
    List<LanguageDataModel> names = [];
    for (int i = 0; i < languageList.length; i++) {
      if (languageList[i].code!.trim().isNotEmpty &&
          controllers[i + 1].text.trim().isNotEmpty) {
        names.add(
          LanguageDataModel(
            code: languageList[i].code!,
            name: controllers[i + 1].text,
          ),
        );
      }
    }
    return names;
  }

  void saveOrUpdateData() {
    // Use CouponValidator for validation
    fieldErrors = CouponValidator.validate(
      fieldTextController: fieldTextController,
      couponValueController: couponValueController,
      maxUsageCountController: maxUsageCountController,
      issuedDate: issuedDate,
      expiryDate: expiryDate,
      couponType: couponType,
    );

    // If there are errors, trigger setState to show them
    if (fieldErrors.isNotEmpty) {
      setState(() {});
      return;
    }

    // Pack language data for names
    names = packLanguage(fieldTextController);

    // Parse values using validator helper methods
    final couponValue = CouponValidator.parseCouponValue(
      couponValueController.text,
    );
    final maxUsageCount = CouponValidator.parseMaxUsageCount(
      maxUsageCountController.text,
    );
    final maxUsageCountPerCustomer =
        CouponValidator.parseMaxUsageCountPerCustomer(
          maxUsageCountPerCustomerController.text,
        );

    final couponData = {
      'couponcode': fieldTextController[0].text.trim().toUpperCase(),
      'names': names.map((name) => name.toJson()).toList(),
      'couponvalue': couponValue,
      'coupontype': couponType,
      'remark': remarkController.text.trim(),
      'issueddate': issuedDate?.toUtcIsoString(),
      'expirydate': expiryDate?.toUtcIsoString(),
      'maxusagecount': maxUsageCount,
      'maxusagecountpercustomer': maxUsageCountPerCustomer,
      'isonetimeuse': isOneTimeUse,
      'customercodes': customerCodes,
      'status': status,
      'product_condition': productCondition?.toJson(),
    };

    if (screenEvent == global.ScreenEventEnum.add) {
      context.read<CouponBloc>().add(CreateCoupon(couponData: couponData));
    } else if (screenEvent == global.ScreenEventEnum.edit) {
      context.read<CouponBloc>().add(
        UpdateCoupon(id: selectGuid, couponData: couponData),
      );
    }
  }

  void getDataToEditScreen(CouponModel coupon) {
    fieldTextController[0].text = coupon.couponcode ?? "";

    // Set multi-language names
    for (int i = 0; i < languageList.length; i++) {
      int controllerIndex =
          i + 1; // coupon code is at index 0, language names start at index 1

      if (controllerIndex < fieldTextController.length) {
        fieldTextController[controllerIndex].text = "";
        if (coupon.names != null) {
          for (LanguageDataModel name in coupon.names!) {
            // Case-insensitive comparison
            if (name.code?.toLowerCase() ==
                languageList[i].code!.toLowerCase()) {
              fieldTextController[controllerIndex].text = name.name;
              break;
            }
          }
        }
      }
    }

    couponValueController.text = coupon.couponvalue?.toString() ?? "";
    remarkController.text = coupon.remark ?? "";
    maxUsageCountController.text = coupon.maxusagecount?.toString() ?? "";
    maxUsageCountPerCustomerController.text =
        coupon.maxusagecountpercustomer?.toString() ?? "";

    // Convert UTC dates to local time for display using extension
    issuedDate = coupon.issueddate?.parseLocalDate();
    expiryDate = coupon.expirydate?.parseLocalDate();
    couponType = coupon.coupontype ?? 0;
    status = coupon.status ?? 0;
    isOneTimeUse = coupon.isonetimeuse ?? false;
    customerCodes = coupon.customercodes ?? [];
    productCondition = coupon.productCondition;
  }

  void findFocusNext(int currentIndex) {
    if (currentIndex < fieldFocusNodes.length - 1) {
      fieldFocusNodes[currentIndex + 1].focusNode.requestFocus();
    } else {
      // Move to coupon specific fields
      couponValueFocusNode.requestFocus();
    }
  }

  Widget editScreen({mobileScreen = false}) {
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
        title: Text(headerEdit + global.language("coupon")),
        actions: <Widget>[
          EditFontSizeControl(
            onChanged: () => setState(() {}),
          ),
          if (selectGuid.isNotEmpty)
            IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () {
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
                          context.read<CouponBloc>().add(
                            DeleteCoupon(id: selectGuid),
                          );
                        },
                        child: Text(global.language('confirm')),
                      ),
                    ],
                  ),
                );
              },
              icon: Icon(Icons.delete),
            ),
          if (isSaveAllow == false && selectGuid.trim().isNotEmpty)
            IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () {
                switchToEdit(
                  listData[listData.indexOf(
                    listData.firstWhere(
                      (element) => (element.guidfixed ?? "") == selectGuid,
                    ),
                  )],
                );
              },
              icon: Icon(Icons.edit),
            ),
          if ((screenEvent == global.ScreenEventEnum.edit ||
                  screenEvent == global.ScreenEventEnum.add) &&
              global.systemLanguage.length > 1)
            IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () async {
                setState(() {
                  isLoadTranslation = true;
                });
                for (int i = 1; i <= languageList.length; i++) {
                  try {
                    var translation = await translator.translate(
                      fieldTextController[1].text,
                      to: languageList[i - 1].codeTranslator!,
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
              icon: Icon(Icons.translate),
            ),
          if (isSaveAllow == true)
            IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () => saveOrUpdateData(),
              icon: Icon(Icons.save),
            ),
        ],
      ),
      body: Builder(
        builder: (context) {
          final scaleFactor = global.editFontScaleFactor;
          return MediaQuery(
            data: MediaQuery.of(context).copyWith(
              textScaler: TextScaler.linear(scaleFactor),
            ),
            child: Theme(
              data: Theme.of(context).copyWith(
                inputDecorationTheme: Theme.of(context).inputDecorationTheme.copyWith(
                  contentPadding: EdgeInsets.symmetric(
                    horizontal: 12 * scaleFactor,
                    vertical: 10 * scaleFactor,
                  ),
                ),
                iconTheme: IconThemeData(size: 24 * scaleFactor),
              ),
              child: SingleChildScrollView(
        controller: editScrollController,
        child: Container(
          color: global.theme.cardColor,
          width: double.infinity,
          padding: const EdgeInsets.all(20),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // ข้อมูลพื้นฐาน (รหัสและชื่อคูปอง)
              CouponBasicInfoWidget(
                fieldTextController: fieldTextController,
                fieldFocusNodes: fieldFocusNodes,
                languageList: languageList,
                fieldErrors: fieldErrors,
                isLoadTranslation: isLoadTranslation,
                onCodeChanged: (value) {
                  setState(() {
                    isChange = true;
                    fieldTextController[0].value = TextEditingValue(
                      text: value.toUpperCase(),
                      selection: fieldTextController[0].selection,
                    );
                    fieldErrors.remove('coupon_code');
                  });
                },
                onNameChanged: (value) {
                  setState(() {
                    isChange = true;
                    fieldErrors.remove('coupon_name');
                  });
                },
              ),

              // ประเภทและค่าส่วนลด
              CouponTypeValueWidget(
                couponType: couponType,
                couponValueController: couponValueController,
                couponValueFocusNode: couponValueFocusNode,
                fieldErrors: fieldErrors,
                onTypeChanged: (type) {
                  setState(() {
                    couponType = type;
                    isChange = true;
                  });
                },
                onValueChanged: (value) {
                  setState(() {
                    isChange = true;
                    fieldErrors.remove('coupon_value');
                  });
                },
              ),

              // เงื่อนไข (วันที่และจำนวนการใช้งาน)
              CouponConditionsWidget(
                issuedDate: issuedDate,
                expiryDate: expiryDate,
                maxUsageCountController: maxUsageCountController,
                maxUsageCountPerCustomerController:
                    maxUsageCountPerCustomerController,
                maxUsageCountFocusNode: maxUsageCountFocusNode,
                maxUsageCountPerCustomerFocusNode:
                    maxUsageCountPerCustomerFocusNode,
                fieldErrors: fieldErrors,
                isEditMode: screenEvent != global.ScreenEventEnum.list,
                onIssuedDateChanged: (date) {
                  setState(() {
                    issuedDate = date;
                    isChange = true;
                    fieldErrors.remove('issued_date');
                  });
                },
                onExpiryDateChanged: (date) {
                  setState(() {
                    // Set to end of day automatically using extension
                    expiryDate = date?.endOfDay;
                    isChange = true;
                    fieldErrors.remove('expiry_date');
                  });
                },
                onMaxUsageChanged: (value) {
                  setState(() {
                    isChange = true;
                    fieldErrors.remove('max_usage_count');
                  });
                },
                onMaxUsagePerCustomerChanged: (value) {
                  setState(() {
                    isChange = true;
                  });
                },
              ),

              // สถานะและตัวเลือก
              CouponStatusWidget(
                status: status,
                isOneTimeUse: isOneTimeUse,
                onStatusChanged: (newStatus) {
                  setState(() {
                    status = newStatus;
                    isChange = true;
                  });
                },
                onOneTimeUseChanged: (value) {
                  setState(() {
                    isOneTimeUse = value;
                    isChange = true;
                  });
                },
              ),

              // ลูกค้า
              CouponCustomerSelectionWidget(
                customerCodes: customerCodes,
                onAddCustomer: openCustomerSearch,
                onRemoveCustomer: removeCustomerCode,
              ),

              // เงื่อนไขสินค้า (Product Conditions)
              CouponProductConditionWidget(
                productCondition: productCondition,
                isEditMode: screenEvent != global.ScreenEventEnum.list,
                onAddProductCodes: addProductCodes,
                onAddGroupCodes: addGroupCodes,
                onAddGroupSuboneCodes: addGroupSuboneCodes,
                onAddGroupSubtwoCodes: addGroupSubtwoCodes,
                onAddBrandCodes: addBrandCodes,
                onAddDesignCodes: addDesignCodes,
                onAddModelCodes: addModelCodes,
                onAddPatternCodes: addPatternCodes,
                onAddGradeCodes: addGradeCodes,
                onAddCategoryCodes: addCategoryCodes,
                onAddClassCodes: addClassCodes,
                onEditMinimumAmount: editMinimumAmount,
                onRemoveCode: removeProductConditionCode,
              ),

              // หมายเหตุ
              CouponRemarkWidget(
                remarkController: remarkController,
                remarkFocusNode: remarkFocusNode,
                onChanged: (value) {
                  setState(() {
                    isChange = true;
                  });
                },
              ),

              const SizedBox(height: 32),

              // ปุ่มบันทึก
              if (isSaveAllow)
                SizedBox(
                  width: double.infinity,
                  height: 50,
                  child: ElevatedButton.icon(
                    style: ElevatedButton.styleFrom(
                      backgroundColor: global.theme.buttonColor,
                      foregroundColor: global.theme.onPrimaryColor,
                      elevation: 2,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(8),
                      ),
                    ),
                    onPressed: saveOrUpdateData,
                    icon: Icon(Icons.save, size: 20),
                    label: Text(
                      global.language("save") +
                          ((kIsWeb ||
                                  Platform.isWindows ||
                                  Platform.isLinux ||
                                  Platform.isMacOS)
                              ? " (F10)"
                              : ""),
                      style: TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.w600,
                      ),
                    ),
                  ),
                ),

              const SizedBox(height: 24),
            ],
          ),
        ),
            ),
          ),
          );
        },
      ),
    );
  }

  void openCustomerSearch() async {
    try {
      final result = await Navigator.push(
        context,
        MaterialPageRoute(
          builder: (context) => const CustomerSearchScreen(word: ""),
        ),
      );

      if (result != null && result is global.SearchDebtorModel) {
        if (result.guid.isNotEmpty) {
          // เพิ่มรหัสลูกค้าใหม่เข้าไปใน list
          if (!customerCodes.contains(result.code)) {
            customerCodes.add(result.code);
            isChange = true;
            setState(() {});

            CouponErrorHandler.showSuccess(
              context,
              'add_customer',
              details: "${result.code} - ${global.packName(result.names)}",
            );
          } else {
            CouponErrorHandler.showWarning(
              context,
              "${global.language("customer_already_exists")}: ${result.code}",
            );
          }
        }
      }
    } catch (e) {
      CouponErrorHandler.handleError(context, e, operation: 'add_customer');
    }
  }

  void removeCustomerCode(String customerCode) {
    showDialog<String>(
      context: context,
      builder: (BuildContext context) => AlertDialog(
        title: Text(global.language('coupon_confirm_delete_customer')),
        content: Text('$customerCode?'),
        actions: <Widget>[
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text(global.language('no')),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: global.theme.negativeHighlightTextColor,
              foregroundColor: global.theme.onPrimaryColor,
            ),
            onPressed: () {
              Navigator.pop(context);
              customerCodes.remove(customerCode);
              isChange = true;
              setState(() {});

              CouponErrorHandler.showSuccess(
                context,
                'remove_customer',
                details: customerCode,
              );
            },
            child: Text(global.language('confirm')),
          ),
        ],
      ),
    );
  }

  // ============== Product Condition Functions ==============

  /// Generic function to add product condition codes
  /// This eliminates code duplication across all 11 addXxxCodes functions
  Future<void> _addProductConditionCode({
    required Widget Function() searchScreenBuilder,
    required List<String>? Function(ProductConditionModel) getCurrentCodes,
    required ProductConditionModel Function(ProductConditionModel, List<String>)
    updateCodes,
    required bool Function(dynamic) isValidResult,
    required String? Function(dynamic) getCodeFromResult,
  }) async {
    try {
      final result = await Navigator.push(
        context,
        MaterialPageRoute(builder: (context) => searchScreenBuilder()),
      );

      if (result != null && isValidResult(result)) {
        final code = getCodeFromResult(result);
        if (code != null && code.isNotEmpty) {
          // Initialize productCondition if null
          productCondition ??= ProductConditionModel();

          // Get current codes list
          List<String> currentCodes = getCurrentCodes(productCondition!) ?? [];

          // Check if code already exists
          if (!currentCodes.contains(code)) {
            currentCodes.add(code);

            // Update productCondition with new codes
            productCondition = updateCodes(productCondition!, currentCodes);

            isChange = true;
            setState(() {});

            CouponErrorHandler.showSuccess(
              context,
              'add_product',
              details: code,
            );
          } else {
            CouponErrorHandler.showWarning(
              context,
              "${global.language("customer_already_exists")}: $code",
            );
          }
        }
      }
    } catch (e) {
      CouponErrorHandler.handleError(context, e, operation: 'add_product');
    }
  }

  void addProductCodes() async {
    await _addProductConditionCode(
      searchScreenBuilder: () =>
          const BarcodeSearchScreen(word: "", screen: "material"),
      getCurrentCodes: (model) => model.productCodes,
      updateCodes: (model, codes) => model.copyWith(productCodes: codes),
      isValidResult: (result) => result is ProductBarcodeModel,
      getCodeFromResult: (result) => (result as ProductBarcodeModel).itemcode,
    );
  }

  void addGroupCodes() async {
    await _addProductConditionCode(
      searchScreenBuilder: () => const ProductMasterGroupScreen(word: ""),
      getCurrentCodes: (model) => model.groupCodes,
      updateCodes: (model, codes) => model.copyWith(groupCodes: codes),
      isValidResult: (result) =>
          result is global.SearchCodeNameModel && !result.isCancel,
      getCodeFromResult: (result) =>
          (result as global.SearchCodeNameModel).code,
    );
  }

  void addGroupSuboneCodes() async {
    await _addProductConditionCode(
      searchScreenBuilder: () => const ProductGroupSub1SearchScreen(word: ""),
      getCurrentCodes: (model) => model.groupSuboneCodes,
      updateCodes: (model, codes) => model.copyWith(groupSuboneCodes: codes),
      isValidResult: (result) =>
          result is global.SearchCodeNameModel && !result.isCancel,
      getCodeFromResult: (result) =>
          (result as global.SearchCodeNameModel).code,
    );
  }

  void addGroupSubtwoCodes() async {
    await _addProductConditionCode(
      searchScreenBuilder: () => const ProductGroupSub2SearchScreen(word: ""),
      getCurrentCodes: (model) => model.groupSubtwoCodes,
      updateCodes: (model, codes) => model.copyWith(groupSubtwoCodes: codes),
      isValidResult: (result) =>
          result is global.SearchCodeNameModel && !result.isCancel,
      getCodeFromResult: (result) =>
          (result as global.SearchCodeNameModel).code,
    );
  }

  void addBrandCodes() async {
    await _addProductConditionCode(
      searchScreenBuilder: () => const ProductBrandSearchScreen(word: ""),
      getCurrentCodes: (model) => model.brandCodes,
      updateCodes: (model, codes) => model.copyWith(brandCodes: codes),
      isValidResult: (result) =>
          result is global.SearchCodeNameModel && !result.isCancel,
      getCodeFromResult: (result) =>
          (result as global.SearchCodeNameModel).code,
    );
  }

  void addDesignCodes() async {
    await _addProductConditionCode(
      searchScreenBuilder: () => const ProductDesignSearchScreen(word: ""),
      getCurrentCodes: (model) => model.designCodes,
      updateCodes: (model, codes) => model.copyWith(designCodes: codes),
      isValidResult: (result) =>
          result is global.SearchCodeNameModel && !result.isCancel,
      getCodeFromResult: (result) =>
          (result as global.SearchCodeNameModel).code,
    );
  }

  void addModelCodes() async {
    await _addProductConditionCode(
      searchScreenBuilder: () => const ProductModelSearchScreen(word: ""),
      getCurrentCodes: (model) => model.modelCodes,
      updateCodes: (model, codes) => model.copyWith(modelCodes: codes),
      isValidResult: (result) =>
          result is global.SearchCodeNameModel && !result.isCancel,
      getCodeFromResult: (result) =>
          (result as global.SearchCodeNameModel).code,
    );
  }

  void addPatternCodes() async {
    await _addProductConditionCode(
      searchScreenBuilder: () => const ProductPatternSearchScreen(word: ""),
      getCurrentCodes: (model) => model.patternCodes,
      updateCodes: (model, codes) => model.copyWith(patternCodes: codes),
      isValidResult: (result) =>
          result is global.SearchCodeNameModel && !result.isCancel,
      getCodeFromResult: (result) =>
          (result as global.SearchCodeNameModel).code,
    );
  }

  void addGradeCodes() async {
    await _addProductConditionCode(
      searchScreenBuilder: () => const ProductGradeSearchScreen(word: ""),
      getCurrentCodes: (model) => model.gradeCodes,
      updateCodes: (model, codes) => model.copyWith(gradeCodes: codes),
      isValidResult: (result) =>
          result is global.SearchCodeNameModel && !result.isCancel,
      getCodeFromResult: (result) =>
          (result as global.SearchCodeNameModel).code,
    );
  }

  void addCategoryCodes() async {
    await _addProductConditionCode(
      searchScreenBuilder: () => const ProductCategorySearchScreenNew(word: ""),
      getCurrentCodes: (model) => model.categoryCodes,
      updateCodes: (model, codes) => model.copyWith(categoryCodes: codes),
      isValidResult: (result) =>
          result is global.SearchCodeNameModel && !result.isCancel,
      getCodeFromResult: (result) =>
          (result as global.SearchCodeNameModel).code,
    );
  }

  void addClassCodes() async {
    await _addProductConditionCode(
      searchScreenBuilder: () => const ProductClassSearchScreen(word: ""),
      getCurrentCodes: (model) => model.classCodes,
      updateCodes: (model, codes) => model.copyWith(classCodes: codes),
      isValidResult: (result) =>
          result is global.SearchCodeNameModel && !result.isCancel,
      getCodeFromResult: (result) =>
          (result as global.SearchCodeNameModel).code,
    );
  }

  void editMinimumAmount() {
    final TextEditingController amountController = TextEditingController(
      text: productCondition?.minimumAmount?.toString() ?? '',
    );

    showDialog<String>(
      context: context,
      builder: (BuildContext context) => AlertDialog(
        title: Text(global.language('minimum_purchase_amount')),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            TextField(
              controller: amountController,
              keyboardType: TextInputType.number,
              decoration: InputDecoration(
                labelText: global.language('amount'),
                hintText: '0.00',
                suffixText: global.language('baht'),
                border: OutlineInputBorder(),
              ),
              autofocus: true,
            ),
            SizedBox(height: 12),
            Text(
              global.language('coupon_minimum_amount_description'),
              style: TextStyle(
                fontSize: 13,
                color: global.theme.textSecondaryColor,
                fontStyle: FontStyle.italic,
              ),
            ),
          ],
        ),
        actions: <Widget>[
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text(global.language('cancel')),
          ),
          if (productCondition?.minimumAmount != null &&
              productCondition!.minimumAmount! > 0)
            TextButton(
              onPressed: () {
                Navigator.pop(context);
                setState(() {
                  productCondition ??= ProductConditionModel();
                  productCondition = productCondition!.copyWith(
                    minimumAmount: 0,
                  );
                  isChange = true;
                });
                CouponErrorHandler.showSuccess(
                  context,
                  'remove_product',
                  details: global.language('minimum_purchase_amount'),
                );
              },
              style: TextButton.styleFrom(foregroundColor: global.theme.negativeHighlightTextColor),
              child: Text(global.language('remove')),
            ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: global.theme.buttonColor,
              foregroundColor: global.theme.onPrimaryColor,
            ),
            onPressed: () {
              final double? amount = double.tryParse(amountController.text);
              if (amount != null && amount > 0) {
                Navigator.pop(context);
                setState(() {
                  productCondition ??= ProductConditionModel();
                  productCondition = productCondition!.copyWith(
                    minimumAmount: amount,
                  );
                  isChange = true;
                });
                CouponErrorHandler.showSuccess(
                  context,
                  'update',
                  details:
                      '${global.language("minimum_purchase_amount")}: ${amount.toStringAsFixed(2)} ${global.language("baht")}',
                );
              } else {
                CouponErrorHandler.showWarning(
                  context,
                  global.language('coupon_error_value_invalid'),
                );
              }
            },
            child: Text(global.language('save')),
          ),
        ],
      ),
    );
  }

  void removeProductConditionCode(String code, String type) {
    showDialog<String>(
      context: context,
      builder: (BuildContext context) => AlertDialog(
        title: Text(global.language('delete_confirm')),
        content: Text('${global.language("remove")} $code?'),
        actions: <Widget>[
          TextButton(
            onPressed: () => Navigator.pop(context),
            child: Text(global.language('no')),
          ),
          ElevatedButton(
            style: ElevatedButton.styleFrom(
              backgroundColor: global.theme.negativeHighlightTextColor,
              foregroundColor: global.theme.onPrimaryColor,
            ),
            onPressed: () {
              Navigator.pop(context);
              setState(() {
                productCondition ??= ProductConditionModel();

                switch (type) {
                  case 'product':
                    productCondition!.productCodes?.remove(code);
                    break;
                  case 'group':
                    productCondition!.groupCodes?.remove(code);
                    break;
                  case 'group_subone':
                    productCondition!.groupSuboneCodes?.remove(code);
                    break;
                  case 'group_subtwo':
                    productCondition!.groupSubtwoCodes?.remove(code);
                    break;
                  case 'brand':
                    productCondition!.brandCodes?.remove(code);
                    break;
                  case 'design':
                    productCondition!.designCodes?.remove(code);
                    break;
                  case 'model':
                    productCondition!.modelCodes?.remove(code);
                    break;
                  case 'pattern':
                    productCondition!.patternCodes?.remove(code);
                    break;
                  case 'grade':
                    productCondition!.gradeCodes?.remove(code);
                    break;
                  case 'category':
                    productCondition!.categoryCodes?.remove(code);
                    break;
                  case 'class':
                    productCondition!.classCodes?.remove(code);
                    break;
                }

                isChange = true;
              });

              CouponErrorHandler.showSuccess(
                context,
                'remove_product',
                details: code,
              );
            },
            child: Text(global.language('confirm')),
          ),
        ],
      ),
    );
  }

  // ============== Import Functions ==============

  /// Download coupon import template
  Future<void> downloadCouponTemplate() async {
    try {
      // Load the asset file
      final ByteData data = await rootBundle.load(
        'assets/file_import/import_coupon.xlsx',
      );
      final Uint8List bytes = data.buffer.asUint8List();

      // Use the download function
      bool success = await downloadAssetFileBytes(
        bytes,
        'import_coupon_template.xlsx',
      );

      if (success) {
        if (mounted) {
          global.showSnackBar(
            context,
            Icon(Icons.download_done, color: global.theme.onPrimaryColor),
            global.language("template_download_success"),
            global.theme.positiveHighlightTextColor,
          );
        }
      } else {
        if (mounted) {
          global.showSnackBar(
            context,
            Icon(Icons.error, color: global.theme.onPrimaryColor),
            global.language("template_download_failed"),
            global.theme.negativeHighlightTextColor,
          );
        }
      }
    } catch (e) {
      if (mounted) {
        global.showSnackBar(
          context,
          Icon(Icons.error, color: global.theme.onPrimaryColor),
          "${global.language("error_occurred")}: ${e.toString()}",
          global.theme.negativeHighlightTextColor,
        );
      }
    }
  }

  /// Select and upload Excel file for coupon import
  Future<void> selectAndUploadExcelFile() async {
    try {
      FilePickerResult? result = await FilePicker.platform.pickFiles(
        type: FileType.custom,
        allowedExtensions: ['xlsx', 'xls'],
        allowMultiple: false,
      );

      if (result != null && result.files.single.bytes != null) {
        // Show dialog to choose: Validate or Import
        final selectedBytes = result.files.single.bytes!;
        final selectedFilename = result.files.single.name;

        if (!mounted) return;

        showDialog(
          context: context,
          builder: (BuildContext dialogContext) {
            return AlertDialog(
              title: Text(global.language("select_import_mode")),
              content: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    global.language("import_mode_description"),
                    style: TextStyle(fontSize: 14),
                  ),
                  SizedBox(height: 16),
                  ListTile(
                    leading: Icon(Icons.search, color: global.theme.infoHighlightTextColor),
                    title: Text(global.language("validate_only")),
                    subtitle: Text(
                      global.language("validate_only_description"),
                    ),
                    contentPadding: EdgeInsets.zero,
                  ),
                  SizedBox(height: 8),
                  ListTile(
                    leading: Icon(Icons.upload, color: global.theme.positiveHighlightTextColor),
                    title: Text(global.language("import_directly")),
                    subtitle: Text(
                      global.language("import_directly_description"),
                    ),
                    contentPadding: EdgeInsets.zero,
                  ),
                ],
              ),
              actions: [
                TextButton(
                  onPressed: () => Navigator.of(dialogContext).pop(),
                  child: Text(global.language("cancel")),
                ),
                ElevatedButton.icon(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: global.theme.infoHighlightTextColor,
                    foregroundColor: global.theme.onPrimaryColor,
                  ),
                  icon: Icon(Icons.search, size: 18),
                  onPressed: () {
                    Navigator.of(dialogContext).pop();
                    if (mounted) {
                      context.read<CouponBloc>().add(
                        UploadCouponExcel(
                          file: selectedBytes,
                          filename: selectedFilename,
                          validateOnly: true,
                        ),
                      );
                    }
                  },
                  label: Text(global.language("validate_only")),
                ),
                ElevatedButton.icon(
                  style: ElevatedButton.styleFrom(
                    backgroundColor: global.theme.positiveHighlightTextColor,
                    foregroundColor: global.theme.onPrimaryColor,
                  ),
                  icon: Icon(Icons.upload, size: 18),
                  onPressed: () {
                    Navigator.of(dialogContext).pop();
                    if (mounted) {
                      context.read<CouponBloc>().add(
                        UploadCouponExcel(
                          file: selectedBytes,
                          filename: selectedFilename,
                          validateOnly: false,
                        ),
                      );
                    }
                  },
                  label: Text(global.language("import_directly")),
                ),
              ],
            );
          },
        );
      }
    } catch (e) {
      if (mounted) {
        global.showSnackBar(
          context,
          Icon(Icons.error, color: global.theme.onPrimaryColor),
          "${global.language("import_failed")}: ${e.toString()}",
          global.theme.negativeHighlightTextColor,
        );
      }
    }
  }

  /// Show validation result dialog and confirm import
  void showValidationResultDialog(CouponImportUploadData validationResult) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (BuildContext dialogContext) {
        return AlertDialog(
          title: Row(
            children: [
              Icon(
                validationResult.errorCount > 0
                    ? Icons.warning
                    : Icons.check_circle,
                color: validationResult.errorCount > 0
                    ? global.theme.warningHighlightTextColor
                    : global.theme.positiveHighlightTextColor,
              ),
              SizedBox(width: 10),
              Text(global.language("validation_result")),
            ],
          ),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  validationResult.message,
                  style: TextStyle(fontWeight: FontWeight.bold),
                ),
                SizedBox(height: 16),
                _buildInfoRow(
                  global.language("total_rows"),
                  validationResult.totalRows.toString(),
                ),
                _buildInfoRow(
                  global.language("valid_rows"),
                  validationResult.importedCount.toString(),
                  global.theme.positiveHighlightTextColor,
                ),
                if (validationResult.errorCount > 0)
                  _buildInfoRow(
                    global.language("error_rows"),
                    validationResult.errorCount.toString(),
                    global.theme.negativeHighlightTextColor,
                  ),
                if (validationResult.errors != null &&
                    validationResult.errors!.isNotEmpty) ...[
                  const SizedBox(height: 16),
                  const Divider(),
                  SizedBox(height: 8),
                  Text(
                    global.language("error_details"),
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 14,
                    ),
                  ),
                  SizedBox(height: 8),
                  ...validationResult.errors!.take(10).map((error) {
                    return Padding(
                      padding: EdgeInsets.only(bottom: 8),
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Icon(
                            Icons.error_outline,
                            size: 16,
                            color: global.theme.negativeHighlightTextColor,
                          ),
                          SizedBox(width: 8),
                          Expanded(
                            child: Text(
                              '${global.language("row")} ${error.row ?? "-"}${error.couponCode != null ? " (${error.couponCode})" : ""}: ${error.errorMessage}',
                              style: TextStyle(fontSize: 13),
                            ),
                          ),
                        ],
                      ),
                    );
                  }),
                  if (validationResult.errors!.length > 10)
                    Text(
                      '... ${global.language("and")} ${validationResult.errors!.length - 10} ${global.language("more_errors")}',
                      style: TextStyle(
                        fontSize: 12,
                        fontStyle: FontStyle.italic,
                        color: global.theme.iconSecondaryColor,
                      ),
                    ),
                ],
              ],
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.of(dialogContext).pop(),
              child: Text(global.language("cancel")),
            ),
            if (validationResult.errorCount == 0)
              ElevatedButton(
                style: ElevatedButton.styleFrom(
                  backgroundColor: global.theme.buttonColor,
                  foregroundColor: global.theme.onPrimaryColor,
                ),
                onPressed: () async {
                  Navigator.of(dialogContext).pop();
                  // Step 2: Proceed with actual import
                  FilePickerResult? result = await FilePicker.platform
                      .pickFiles(
                        type: FileType.custom,
                        allowedExtensions: ['xlsx', 'xls'],
                        allowMultiple: false,
                      );

                  if (result != null && result.files.single.bytes != null) {
                    if (mounted) {
                      context.read<CouponBloc>().add(
                        UploadCouponExcel(
                          file: result.files.single.bytes!,
                          filename: result.files.single.name,
                          validateOnly: false,
                        ),
                      );
                    }
                  }
                },
                child: Text(global.language("proceed_import")),
              ),
          ],
        );
      },
    );
  }

  /// Show import progress dialog
  void showImportProgressDialog(CouponImportStatusData statusData) {
    if (isImportDialogOpen) return;

    isImportDialogOpen = true;
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (BuildContext dialogContext) {
        return WillPopScope(
          onWillPop: () async => false,
          child: AlertDialog(
            title: Row(
              children: [
                const SizedBox(
                  width: 20,
                  height: 20,
                  child: CircularProgressIndicator(strokeWidth: 2),
                ),
                SizedBox(width: 16),
                Text(global.language("importing_data")),
              ],
            ),
            content: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                LinearProgressIndicator(
                  value: statusData.progress / 100,
                  backgroundColor: global.theme.dividerBorderColor,
                  valueColor: AlwaysStoppedAnimation<Color>(
                    global.theme.appBarColor,
                  ),
                ),
                SizedBox(height: 16),
                Text(
                  '${global.language("progress")}: ${statusData.progress}%',
                  style: TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                  ),
                ),
                SizedBox(height: 8),
                _buildInfoRow(
                  global.language("processed"),
                  '${statusData.processedRows}/${statusData.totalRows}',
                ),
                _buildInfoRow(
                  global.language("success"),
                  statusData.successCount.toString(),
                  global.theme.positiveHighlightTextColor,
                ),
                if (statusData.errorCount > 0)
                  _buildInfoRow(
                    global.language("errors"),
                    statusData.errorCount.toString(),
                    global.theme.negativeHighlightTextColor,
                  ),
                const SizedBox(height: 8),
                Text(
                  statusData.message,
                  style: TextStyle(
                    fontSize: 13,
                    fontStyle: FontStyle.italic,
                    color: global.theme.iconSecondaryColor,
                  ),
                ),
              ],
            ),
          ),
        );
      },
    );
  }

  /// Show import completion dialog
  void showImportCompletionDialog(CouponImportStatusData finalStatus) {
    if (!isImportDialogOpen) return;

    // Close progress dialog
    Navigator.of(context).pop();
    isImportDialogOpen = false;

    // Reload data immediately
    setState(() {
      listData.clear();
    });
    loadDataList("");

    showDialog(
      context: context,
      builder: (BuildContext dialogContext) {
        return AlertDialog(
          title: Row(
            children: [
              Icon(
                finalStatus.errorCount == 0
                    ? Icons.check_circle
                    : Icons.warning,
                color: finalStatus.errorCount == 0
                    ? global.theme.positiveHighlightTextColor
                    : global.theme.warningHighlightTextColor,
                size: 30,
              ),
              SizedBox(width: 10),
              Text(global.language("import_completed")),
            ],
          ),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  finalStatus.message,
                  style: TextStyle(fontWeight: FontWeight.bold),
                ),
                SizedBox(height: 16),
                _buildInfoRow(
                  global.language("total_rows"),
                  finalStatus.totalRows.toString(),
                ),
                _buildInfoRow(
                  global.language("success"),
                  finalStatus.successCount.toString(),
                  global.theme.positiveHighlightTextColor,
                ),
                if (finalStatus.errorCount > 0)
                  _buildInfoRow(
                    global.language("errors"),
                    finalStatus.errorCount.toString(),
                    global.theme.negativeHighlightTextColor,
                  ),
                if (finalStatus.errors != null &&
                    finalStatus.errors!.isNotEmpty) ...[
                  const SizedBox(height: 16),
                  const Divider(),
                  SizedBox(height: 8),
                  Text(
                    global.language("error_details"),
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 14,
                    ),
                  ),
                  SizedBox(height: 8),
                  ...finalStatus.errors!.take(10).map((error) {
                    return Padding(
                      padding: EdgeInsets.only(bottom: 8),
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Icon(
                            Icons.error_outline,
                            size: 16,
                            color: global.theme.negativeHighlightTextColor,
                          ),
                          SizedBox(width: 8),
                          Expanded(
                            child: Text(
                              '${global.language("row")} ${error.row ?? "-"}${error.couponCode != null ? " (${error.couponCode})" : ""}: ${error.errorMessage}',
                              style: TextStyle(fontSize: 13),
                            ),
                          ),
                        ],
                      ),
                    );
                  }),
                  if (finalStatus.errors!.length > 10)
                    Text(
                      '... ${global.language("and")} ${finalStatus.errors!.length - 10} ${global.language("more_errors")}',
                      style: TextStyle(
                        fontSize: 12,
                        fontStyle: FontStyle.italic,
                        color: global.theme.iconSecondaryColor,
                      ),
                    ),
                ],
              ],
            ),
          ),
          actions: [
            ElevatedButton(
              style: ElevatedButton.styleFrom(
                backgroundColor: global.theme.buttonColor,
                foregroundColor: global.theme.onPrimaryColor,
              ),
              onPressed: () {
                Navigator.of(dialogContext).pop();
              },
              child: Text(global.language("ok")),
            ),
          ],
        );
      },
    );
  }

  /// Helper to build info row
  Widget _buildInfoRow(String label, String value, [Color? valueColor]) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          Text(label, style: TextStyle(fontSize: 14)),
          Text(
            value,
            style: TextStyle(
              fontSize: 14,
              fontWeight: FontWeight.bold,
              color: valueColor,
            ),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    queryData = MediaQuery.of(context);
    listKeys.clear();
    couponGuidListChecked.clear();
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return BlocListener<CouponBloc, CouponState>(
            listener: (context, state) {
              blocCouponState = state;

              // เริ่ม Loading
              if (state is CouponLoading) {
                setState(() {
                  isLoading = true;
                });
              }

              if (state is CouponsLoaded) {
                setState(() {
                  isLoading = false; // ปิด Loading
                  if (state.coupons.isNotEmpty) {
                    listData.addAll(state.coupons);
                  }
                });
              }
              if (state is CouponOperationSuccess) {
                setState(() {
                  isLoading = false; // ปิด Loading
                  global.showSnackBar(
                    context,
                    Icon(Icons.save, color: global.theme.onPrimaryColor),
                    state.message,
                    global.theme.infoHighlightTextColor,
                  );
                  clearEditData();
                  listData.clear();
                  loadDataList(searchText);
                });
              }
              if (state is CouponError) {
                setState(() {
                  isLoading = false; // ปิด Loading แม้เกิด error
                  global.showSnackBar(
                    context,
                    Icon(Icons.error, color: global.theme.onPrimaryColor),
                    state.message,
                    global.theme.negativeHighlightTextColor,
                  );
                });
              }
              if (state is CouponLoaded) {
                setState(() {
                  isLoading = false; // ปิด Loading
                  getDataToEditScreen(state.coupon);
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

              // Import validation completed
              if (state is CouponImportValidated) {
                setState(() {
                  isLoading = false;
                });
                showValidationResultDialog(state.validationResult);
              }

              // Import validation/upload in progress
              if (state is CouponImportValidating ||
                  state is CouponImportUploading) {
                setState(() {
                  isLoading = true;
                });
              }

              // Import status checking (show progress)
              if (state is CouponImportStatusChecking) {
                final statusData = state.statusData;
                currentBatchId = statusData.batchId;

                // Show or update progress dialog
                if (!isImportDialogOpen) {
                  showImportProgressDialog(statusData);
                }

                // Continue polling if not completed
                if (!statusData.isCompleted && !statusData.isFailed) {
                  Future.delayed(const Duration(seconds: 2), () {
                    if (mounted && currentBatchId == statusData.batchId) {
                      context.read<CouponBloc>().add(
                        CheckImportStatus(batchId: statusData.batchId),
                      );
                    }
                  });
                }
              }

              // Import completed successfully
              if (state is CouponImportCompleted) {
                setState(() {
                  isLoading = false;
                  currentBatchId = null;
                });
                showImportCompletionDialog(state.finalStatus);
              }

              // Import failed
              if (state is CouponImportFailed) {
                setState(() {
                  isLoading = false;
                  currentBatchId = null;
                });

                // Close progress dialog if open
                if (isImportDialogOpen) {
                  Navigator.of(context).pop();
                  isImportDialogOpen = false;
                }

                // Show error dialog
                showDialog(
                  context: context,
                  builder: (BuildContext dialogContext) {
                    return AlertDialog(
                      title: Row(
                        children: [
                          Icon(Icons.error, color: global.theme.negativeHighlightTextColor, size: 30),
                          SizedBox(width: 10),
                          Text(global.language("import_failed")),
                        ],
                      ),
                      content: SingleChildScrollView(
                        child: Column(
                          mainAxisSize: MainAxisSize.min,
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            Text(
                              state.message,
                              style: TextStyle(
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                            if (state.errors != null &&
                                state.errors!.isNotEmpty) ...[
                              const SizedBox(height: 16),
                              const Divider(),
                              SizedBox(height: 8),
                              Text(
                                global.language("error_details"),
                                style: TextStyle(
                                  fontWeight: FontWeight.bold,
                                  fontSize: 14,
                                ),
                              ),
                              SizedBox(height: 8),
                              ...state.errors!.take(10).map((error) {
                                return Padding(
                                  padding: EdgeInsets.only(bottom: 8),
                                  child: Row(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
                                      Icon(
                                        Icons.error_outline,
                                        size: 16,
                                        color: global.theme.negativeHighlightTextColor,
                                      ),
                                      SizedBox(width: 8),
                                      Expanded(
                                        child: Text(
                                          '${global.language("row")} ${error.row ?? "-"}${error.couponCode != null ? " (${error.couponCode})" : ""}: ${error.errorMessage}',
                                          style: TextStyle(fontSize: 13),
                                        ),
                                      ),
                                    ],
                                  ),
                                );
                              }),
                              if (state.errors!.length > 10)
                                Text(
                                  '... ${global.language("and")} ${state.errors!.length - 10} ${global.language("more_errors")}',
                                  style: TextStyle(
                                    fontSize: 12,
                                    fontStyle: FontStyle.italic,
                                    color: global.theme.iconSecondaryColor,
                                  ),
                                ),
                            ],
                          ],
                        ),
                      ),
                      actions: [
                        ElevatedButton(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: global.theme.buttonColor,
                            foregroundColor: global.theme.onPrimaryColor,
                          ),
                          onPressed: () => Navigator.of(dialogContext).pop(),
                          child: Text(global.language("ok")),
                        ),
                      ],
                    );
                  },
                );
              }
            },
            child: Stack(
              children: [
                // Main content
                (constraints.maxWidth > 800)
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

                // Loading Overlay
                if (isLoading)
                  LoadingOverlay(
                    message: global.language('loading_coupon_data'),
                    subtitle: global.language('please_wait'),
                  ),
              ],
            ),
          );
        },
      ),
    );
  }
}
