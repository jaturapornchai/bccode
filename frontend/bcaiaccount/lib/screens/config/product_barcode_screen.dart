import 'dart:async';
import 'package:smlaicloud/widgets/manual_button.dart';
import 'package:smlaicloud/widgets/edit_font_size_control.dart';
import 'package:smlaicloud/widgets/list_font_size_control.dart';
import 'dart:convert';
import 'dart:io';
import 'package:smlaicloud/bloc/business_type/business_type_bloc.dart';
import 'package:smlaicloud/bloc/company_branch/company_branch_bloc.dart';
import 'package:smlaicloud/bloc/export_csv/export_csv_bloc.dart';
import 'package:smlaicloud/bloc/image/image_upload_bloc.dart';
import 'package:smlaicloud/bloc/product_dimension/product_dimension_bloc.dart';
import 'package:smlaicloud/bloc/sale_channel/sale_channel_bloc.dart';
import 'package:smlaicloud/bloc/unit/unit_bloc.dart';
import 'package:smlaicloud/global.dart';
import 'package:smlaicloud/model/business_type_model.dart';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:smlaicloud/model/dimension_model.dart';
import 'package:smlaicloud/model/order_type_model.dart';
import 'package:smlaicloud/model/product_bom_model.dart';
import 'package:smlaicloud/model/product_category_model.dart';
import 'package:smlaicloud/model/product_group_model.dart';
import 'package:smlaicloud/model/product_type_model.dart';
import 'package:smlaicloud/model/sale_channel_model.dart';
import 'package:smlaicloud/repositories/product_group_repository.dart';
import 'package:smlaicloud/repositories/unit_repository.dart';
import 'package:smlaicloud/screen_search/barcode_search_screen.dart';
import 'package:smlaicloud/screen_search/dimension_search_screen.dart';
import 'package:smlaicloud/screen_search/order_type_search_screen.dart';
import 'package:smlaicloud/screen_search/product_group_search_screen.dart';
import 'package:smlaicloud/screen_search/product_search_screen.dart';
import 'package:smlaicloud/screen_search/product_type_search_screen.dart';
import 'package:smlaicloud/screen_search/supplier_search_screen.dart';
import 'package:smlaicloud/screens/config/product_bom_widget.dart';
import 'package:smlaicloud/widgets/unit_flow_widget.dart';
import 'package:smlaicloud/screens/components/product_preview_screen.dart';
import 'package:smlaicloud/utils/date_picker.dart';
import 'package:smlaicloud/utils/image_tooltip.dart';
import 'package:smlaicloud/utils/util.dart';
import 'package:dropdown_search/dropdown_search.dart';
import 'package:flex_color_picker/flex_color_picker.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter_dropzone/flutter_dropzone.dart';
import 'package:image_picker/image_picker.dart';
import 'package:smlaicloud/bloc/product_barcode/product_barcode_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/price_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:loader_overlay/loader_overlay.dart';
import 'package:loading_animation_widget/loading_animation_widget.dart';
import 'package:split_view/split_view.dart';
import 'package:smlaicloud/screen_search/unit_search_screen.dart';
import 'package:smlaicloud/screens/master/group_selection_screens.dart';
import 'package:smlaicloud/model/master_group_model.dart';
import 'package:smlaicloud/model/master_group_sub1_model.dart';
import 'package:smlaicloud/model/master_group_sub2_model.dart';
import 'package:smlaicloud/model/master_brand_model.dart';
import 'package:smlaicloud/model/master_category_model.dart';
import 'package:smlaicloud/model/master_class_model.dart';
import 'package:smlaicloud/model/master_design_model.dart';
import 'package:smlaicloud/model/master_grade_model.dart';
import 'package:smlaicloud/model/master_model_model.dart';
import 'package:smlaicloud/model/master_pattern_model.dart';
import 'package:translator/translator.dart';
import 'package:uuid/uuid.dart';

import 'package:intl/intl.dart';
import 'package:smlaicloud/utils/focus_utils.dart';

class ProductBarcodeScreen extends StatefulWidget {
  const ProductBarcodeScreen({super.key});

  @override
  State<ProductBarcodeScreen> createState() => ProductBarcodeScreenState();
}

class ProductBarcodeScreenState extends State<ProductBarcodeScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  final translator = GoogleTranslator();
  List<GlobalKey<ImageTooltipState>> tooltipKeys = [];
  late TabController tabController;
  ScrollController editScrollController = ScrollController();
  bool refreshFocus = false;
  TextEditingController searchController = TextEditingController();
  TextEditingController groupController = TextEditingController();
  int focusNodeMax = 0;
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  List<LanguageModel> languageList = <LanguageModel>[];
  List<PriceModel> priceList = <PriceModel>[];
  List<global.FieldFocusModel> fieldFocusNodes = [];
  int focusNodeIndex = 0;
  List<ProductBarcodeModel> listData = [];
  List<String> guidListChecked = [];
  int currentBarcodeNode = -1;
  int bomCurrentBarcodeNode = -1;
  String textDescription = "";
  List<LanguageDataModel> names = [];
  ScrollController listScrollController = ScrollController();
  List<GlobalKey> listKeys = [];
  String searchText = "";
  String selectGuid = "";
  String selectBarcode = "";
  bool isDataChange = false;

  // PG list filter state
  bool _showFilterPanel = false;
  String _filterGroupCode = '';
  String _filterGroupName = '';
  String _filterBrandCode = '';
  String _filterBrandName = '';
  String _filterCategoryCode = '';
  String _filterCategoryName = '';
  double? _filterPriceMin;
  double? _filterPriceMax;
  final TextEditingController _priceMinController = TextEditingController();
  final TextEditingController _priceMaxController = TextEditingController();
  int _totalItems = 0;
  int _hoverIndex = -1;
  final String _sortField = 'barcode';
  final String _sortOrder = 'asc';
  bool isSaveAllow = false;
  late ProductBarcodeState blocCurrentState;
  String headerEdit = "";
  late MediaQueryData queryData;
  int currentListIndex = -1;
  GlobalKey headerKey = GlobalKey();
  bool isKeyUp = false;
  bool isKeyDown = false;
  bool showCheckBox = false;
  bool isEditMode = false;

  TextEditingController alertDescriptionTextEditController =
      TextEditingController();
  TextEditingController descriptionTextEditController = TextEditingController();
  late ProductBarcodeModel screenData;
  File imageFile = File('');
  Uint8List? imageWeb;
  final ImagePicker imagePicker = ImagePicker();
  late DropzoneViewController dropZoneController;
  Color colorSelected = Colors.white;
  final _debouncer = global.Debouncer(300);
  List<UnitModel> unitListData = [];
  List<SaleChannelModel> saleChannelList = [];
  // late Timer screenTimer;
  bool loadingData = false;
  bool showImage = false;
  bool useReferenceBarcode = false;
  bool isPreview = true;
  bool isPreviewBom = false;

  TextEditingController barcodeController = TextEditingController();

  final _formKey = GlobalKey<FormState>();

  List<File> imageFileChoice = [File('')];
  List<Uint8List> imageWebChoice = [Uint8List(0)];
  int uploadImageOptionIndex = -1;
  int uploadImageChoiceIndex = -1;

  List<DimensionModel> dimensionSeleted = [];

  LanguageModel languangeExport = global.config.languages[0];
  bool isLoadTranslation = false;
  final _popupBuilderBranchKey = GlobalKey<DropdownSearchState<String>>();
  final _popupBuilderBusinessKey = GlobalKey<DropdownSearchState<String>>();
  List<BusinessTypeModel> listDataBusinessType = [];
  List<BranchModel> listDataBranchBusinessType = [];
  FiltterBarcodeModel filterBarcode = FiltterBarcodeModel(branch: false);
  List<CompanyBranchModel> listDataBranchAll = [];
  List<BranchModel>? ignorebranches = [];

  ProductBomModel productBom = ProductBomModel(
    guidfixed: '',
    names: [],
    itemunitcode: '',
    itemunitnames: [],
    barcode: '',
    condition: false,
    dividevalue: 0,
    standvalue: 0,
    qty: 0,
    imageuri: '',
    bom: [],
  );

  /// time for sale
  List<TextEditingController> mediaFromDateController = [];
  List<TextEditingController> mediaToDateController = [];
  List<TextEditingController> mediaFromTimeController = [];
  List<TextEditingController> mediaToTimeController = [];
  List<List<DayOfWeekModel>> dayOfWeekSeleted = [];
  List<TimeOfDay> _selectedTime = [];
  List<DayOfWeekModel> dayOfWeekList = [
    DayOfWeekModel(code: '1', name: global.language('monday')),
    DayOfWeekModel(code: '2', name: global.language('tuesday')),
    DayOfWeekModel(code: '3', name: global.language('wendesday')),
    DayOfWeekModel(code: '4', name: global.language('thursday')),
    DayOfWeekModel(code: '5', name: global.language('friday')),
    DayOfWeekModel(code: '6', name: global.language('saturday')),
    DayOfWeekModel(code: '7', name: global.language('sunday')),
  ];
  late DateTime dateNow = DateTime.now();
  List<TimeForSaleModel> timeForSales = [];
  bool isShowTimeForSale = false;

  /// ต้นทุนมาตรฐาน
  List<TextEditingController> asDateController = [];
  List<TextEditingController> fiexdCostAmountController = [];
  List<FiexdCostModel> fiexdCostList = [];

  void setSystemLanguageList() async {
    await global.setSystemLanguage(context);

    for (int i = 0; i < global.config.languages.length; i++) {
      if (global.config.languages[i].isuse!) {
        languageList.add(global.config.languages[i]);
      }
    }

    for (int i = 0; i < global.config.prices.length; i++) {
      if (global.config.prices[i].isUse) {
        priceList.add(global.config.prices[i]);
      }
    }

    clearEditData();
    listData = [];
    // Using Future.wait to wait for multiple asynchronous operations
    await Future.wait<void>([
      loadDataList("", filterBarcode),
      loadUnit(),
      loadBranchAll(),
      loadSaleChannelList(),
    ]);
  }

  @override
  void initState() {
    AppLogger.info(
      '🏁 PRODUCT BARCODE SCREEN: เข้าสู่หน้าจอจัดการบาร์โค้ดสินค้า',
    );

    loadDataBusinessTypesList();
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
    // screenTimer.cancel();
    removeTooltip();
    for (int i = 0; i < tooltipKeys.length; i++) {
      tooltipKeys[i].currentState?.dispose();
    }
    super.dispose();
  }

  void removeTooltip() {
    for (int i = 0; i < tooltipKeys.length; i++) {
      tooltipKeys[i].currentState?.removeTooltip();
    }
  }

  Future<void> loadDataList(
    String search,
    FiltterBarcodeModel filterBarcode, {
    bool resetList = false,
  }) async {
    AppLogger.info('📋 LOAD DATA: กำลังโหลดรายการบาร์โค้ดจาก PG (search: "$search")');

    // ถ้า search เปลี่ยน → reset list เริ่ม offset 0
    final bool shouldReset = resetList || search != searchText;
    if (shouldReset) {
      listData = [];
    }

    setState(() {
      loadingData = true;
    });
    searchText = search;

    context.read<ProductBarcodeBloc>().add(
      ProductBarcodeLoadListPg(
        offset: (listData.isEmpty) ? 0 : listData.length,
        limit: global.loadDataPerPage,
        keyword: search,
        groupCode: _filterGroupCode,
        brandCode: _filterBrandCode,
        categoryCode: _filterCategoryCode,
        priceMin: _filterPriceMin,
        priceMax: _filterPriceMax,
        sortField: search.isNotEmpty ? 'relevance' : _sortField,
        sortOrder: search.isNotEmpty ? 'desc' : _sortOrder,
      ),
    );
  }

  Future<void> loadUnit() async {
    context.read<UnitBloc>().add(
      const UnitLoadList(offset: 0, limit: 2000, search: ""),
    );
  }

  Future<void> loadBranchAll() async {
    listDataBranchAll = [];
    context.read<CompanyBranchBloc>().add(
      const CompanyBranchLoadList(offset: 0, limit: 100, search: ""),
    );
  }

  Future<void> loadSaleChannelList() async {
    context.read<SaleChannelBloc>().add(
      const SaleChannelLoadList(offset: 0, limit: 2000, search: ""),
    );
  }

  void onScrollList() {
    if (listScrollController.offset >=
            listScrollController.position.maxScrollExtent &&
        !listScrollController.position.outOfRange) {
      /// set delay 500 ms
      _debouncer.run(() {
        loadDataList(searchText, filterBarcode);
      });
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

  List<LanguageDataModel> getUnitName(String data) {
    if (unitListData.isNotEmpty && data != '') {
      List<LanguageDataModel> res = [];
      for (var ele in unitListData) {
        if (ele.unitcode == data) {
          res = ele.names;
        }
      }
      return res;
    } else {
      return [];
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
    List<PriceDataModel> prices = [];
    for (int i = 0; i < priceList.length; i++) {
      prices.add(PriceDataModel(keynumber: priceList[i].keyNumber, price: 0));
    }

    barcodeController.text = "";
    alertDescriptionTextEditController.text = "";
    descriptionTextEditController.text = "";
    screenData = ProductBarcodeModel(
      barcode: "",
      guidfixed: "",
      groupcode: "",
      groupnames: [],
      names: names,
      itemcode: "",
      itemunitcode: "",
      itemunitnames: itemunitnames,
      prices: prices,
      imageuri: "",
      useimageorcolor: true,
      colorselect: "",
      colorselecthex: "",
      options: [],
      dividevalue: 1,
      standvalue: 1,
      refbarcode: null,
      isalacarte: true,
      isdiscountpointofpurchase: true,
      restaurant: ProductBarcodeRestaurantModel(),
      issplitunitprint: true,
      isonlystaff: false,
      ordertypes: [],
      producttype: ProductTypeModel(guidfixed: "", code: "", names: []),
      businesstypes: [
        BusinessTypeModel(
          guidfixed: global.companyBranchSelectData.businesstype!.guidfixed,
          code: global.companyBranchSelectData.businesstype!.code!,
          names: global.companyBranchSelectData.businesstype!.names!,
        ),
      ],
      ignorebranches: [
        // BranchModel(
        //   guidfixed: global.companyBranchSelectData.guidfixed,
        //   code: global.companyBranchSelectData.code,
        //   names: global.companyBranchSelectData.names,
        // )
      ],
      foodtype: 0,
      isstockforrestaurant: false,
      discount: "",
      manufacturerguid: "",
      manufacturercode: "",
      manufacturernames: [],
      isalert: false,
      alertdescription: "",
      description: "",
      // Initialize new master data fields
      brandcode: "",
      brandguid: "",
      brandnames: [],
      categorycode: "",
      categoryguid: "",
      categorynames: [],
      classcode: "",
      classguid: "",
      classnames: [],
      designcode: "",
      designguid: "",
      designnames: [],
      gradecode: "",
      gradeguid: "",
      gradenames: [],
      modelcode: "",
      modelguid: "",
      modelnames: [],
      groupsubonecode: "",
      groupsuboneguid: "",
      groupsubonenames: [],
      groupsubtwocode: "",
      groupsubtwoguid: "",
      groupsubtwonames: [],
      patterncode: "",
      patternguid: "",
      patternnames: [],
    );

    isDataChange = false;
    focusNodeIndex = 0;
    refreshFocus = true;

    uploadImageOptionIndex = -1;
    uploadImageChoiceIndex = -1;
    imageFileChoice = [File('')];
    imageWebChoice = [Uint8List(0)];

    setState(() {
      imageFile = File('');
      imageWeb = null;
    });

    _selectedTime = [TimeOfDay.now(), TimeOfDay.now()];
    dayOfWeekSeleted = [];
    mediaFromDateController = [];
    mediaToDateController = [];
    mediaFromTimeController = [];
    mediaToTimeController = [];
    timeForSales = [];

    asDateController = [];
    fiexdCostAmountController = [];
    fiexdCostList = [];

    isShowTimeForSale = false;

    loadDataBranchList(screenData.businesstypes!);
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
              style: ElevatedButton.styleFrom(backgroundColor: global.theme.buttonDangerColor),
              onPressed: () => Navigator.pop(context),
              child: Text(global.language('no')),
            ),
            ElevatedButton(
              style: ElevatedButton.styleFrom(backgroundColor: global.theme.primaryColor),
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

  void getData(String barcode) {
    AppLogger.info('🔍 GET DATA: กำลังโหลดข้อมูลสินค้า (barcode: $barcode)');

    headerEdit = global.language("show");
    isEditMode = false;
    imageWeb = null;
    imageFile = File('');

    if (barcode.isNotEmpty) {
      AppLogger.debug('🔍 GET DATA: ส่ง ProductBarcodeGetByBarcode event');
      context.read<ProductBarcodeBloc>().add(ProductBarcodeGetByBarcode(barcode: barcode));
    } else {
      AppLogger.warning('⚠️ GET DATA: barcode ว่าง — ข้ามการโหลด');
    }
  }

  void upLoadImageChoice() {
    if (imageFileChoice.isNotEmpty) {
      if (imageFileChoice[0].path != '') {
        context.read<ImageUploadBloc>().add(
          ImageUploadFileSaved(
            imageFiles: imageFileChoice,
            imageWeb: imageWebChoice,
          ),
        );
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
        title: Text(global.language('barcode')),
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
          const ManualButton(path: 'barcode'),
          Padding(
            padding: const EdgeInsets.only(right: 20.0),
            child: IconButton(
              onPressed: () {
                AppLogger.info('📤 Upload/Import สินค้า');
                // นำทางไปหน้า Import Product From File
                Navigator.pushNamed(context, '/importproductfromfile');
              },
              icon: const Icon(Icons.upload),
              tooltip: 'Import สินค้า',
            ),
          ),
          Padding(
            padding: EdgeInsets.only(right: 20.0),
            child: IconButton(
              onPressed: () {
                AppLogger.info('📥 Export ข้อมูลบาร์โค้ดสินค้า');

                /// dialog select language code
                showDialog<String>(
                  context: context,
                  builder: (BuildContext context) {
                    // Initialize the selected language outside of StatefulBuilder
                    LanguageModel? selectedLanguage =
                        global.config.languages[0];

                    return AlertDialog(
                      title: Text(global.language('select_language')),
                      content: StatefulBuilder(
                        builder: (BuildContext context, StateSetter setState) {
                          return InputDecorator(
                            decoration: const InputDecoration(
                              border: OutlineInputBorder(),
                            ),
                            child: DropdownButtonHideUnderline(
                              child: DropdownButton<LanguageModel>(
                                value: selectedLanguage,
                                icon: const Icon(Icons.arrow_drop_down),
                                style: TextStyle(
                                  color: global.theme.secondaryColor,
                                ),
                                underline: Container(
                                  color: global.theme.secondaryColor,
                                ),
                                onChanged: (LanguageModel? value) {
                                  setState(() {
                                    selectedLanguage = value!;
                                  });
                                },
                                isDense: true,
                                isExpanded: true,
                                items: global.config.languages
                                    .map<DropdownMenuItem<LanguageModel>>((
                                      LanguageModel value,
                                    ) {
                                      return DropdownMenuItem<LanguageModel>(
                                        value: value,
                                        child: Row(
                                          children: <Widget>[
                                            Image.asset(
                                              'assets/flags/${value.code}.png', // Ensure the image path is correct
                                              width: 30,
                                              height: 30,
                                              errorBuilder:
                                                  (context, error, stackTrace) {
                                                    return const Icon(
                                                      Icons.error,
                                                    ); // Error icon if the image fails to load
                                                  },
                                            ),
                                            const SizedBox(
                                              width: 10,
                                            ), // Spacing between the image and text
                                            Text(value.name!),
                                          ],
                                        ),
                                      );
                                    })
                                    .toList(),
                              ),
                            ),
                          );
                        },
                      ),
                      actions: <Widget>[
                        // Text button cancel
                        ElevatedButton(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: global.theme.buttonDangerColor,
                          ),
                          onPressed: () {
                            Navigator.pop(context);
                          },
                          child: Text(global.language('cancel')),
                        ),

                        // Export button
                        ElevatedButton(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: global.theme.primaryColor,
                          ),
                          onPressed: () {
                            context.read<ExportCsvBloc>().add(
                              ProductBarcodeExport(
                                languageCode: selectedLanguage!.code!,
                              ),
                            );
                            Navigator.pop(context);
                          },
                          child: Text(global.language('export')),
                        ),
                      ],
                    );
                  },
                );
              },
              icon: const Icon(Icons.download),
              tooltip: 'Export ข้อมูล', // เพิ่ม tooltip
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
                      if (showCheckBox) {
                        showCheckBox = false;
                        guidListChecked.clear();
                      } else {
                        showCheckBox = true;
                        global.showSnackBar(
                          context,
                          Icon(Icons.delete, color: global.theme.onPrimaryColor),
                          global.language("choose_item_delete"),
                          global.theme.primaryColor,
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
                            backgroundColor: global.theme.buttonDangerColor,
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
                            // guidListChecked เก็บ barcode → map กลับเป็น guidfixed สำหรับ MongoDB
                            final guidsToDelete = <String>[];
                            for (final bc in guidListChecked) {
                              final match = listData.where((e) => (e.barcode ?? '') == bc);
                              if (match.isNotEmpty && match.first.guidfixed.isNotEmpty) {
                                guidsToDelete.add(match.first.guidfixed);
                              }
                            }
                            if (guidsToDelete.isNotEmpty) {
                              context.read<ProductBarcodeBloc>().add(
                                ProductBarcodeDeleteMany(guid: guidsToDelete),
                              );
                            }
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
                      isPreview = false;
                      selectGuid = "";
                      selectBarcode = "";
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
                    (element) => (element.barcode ?? '') == selectBarcode,
                  ),
                );
                if (index > 0) {
                  selectBarcode = listData[index - 1].barcode ?? '';
                  selectGuid = listData[index - 1].guidfixed;
                  currentListIndex = index + 1;
                  isKeyUp = true;

                  getData(listData[index - 1].barcode ?? '');
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                isKeyUp = false;
                int index = listData.indexOf(
                  listData.firstWhere(
                    (element) => (element.barcode ?? '') == selectBarcode,
                  ),
                );
                selectBarcode = listData[index + 1].barcode ?? '';
                selectGuid = listData[index + 1].guidfixed;
                currentListIndex = index + 1;
                isKeyDown = true;
                getData(listData[index + 1].barcode ?? '');
              }
            }
          }
          return KeyEventResult.ignored;
        },
        child: Column(
          children: [
            // Search bar + filter toggle + font size
            Container(
              padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
              decoration: BoxDecoration(
                color: global.theme.cardColor,
                borderRadius: BorderRadius.circular(4),
              ),
              child: Row(
                children: [
                  Expanded(
                    child: TextField(
                      inputFormatters: [_ThaiCodePointBackspaceFormatter()],
                      onSubmitted: (value) {
                        searchFocusNode.requestFocus();
                      },
                      onChanged: (value) {
                        _debouncer.run(() {
                          loadDataList(value, filterBarcode);
                        });
                      },
                      autofocus: false,
                      focusNode: searchFocusNode,
                      controller: searchController,
                      decoration: InputDecoration(
                        isDense: true,
                        contentPadding: const EdgeInsets.symmetric(vertical: 8, horizontal: 4),
                        border: InputBorder.none,
                        hintText: global.language('search'),
                        prefixIcon: const Icon(Icons.search, size: 20),
                        prefixIconConstraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                      ),
                    ),
                  ),
                  // Filter toggle
                  IconButton(
                    focusNode: FocusNode(skipTraversal: true),
                    icon: Icon(
                      _showFilterPanel ? Icons.filter_alt : Icons.filter_alt_outlined,
                      color: (_filterGroupCode.isNotEmpty || _filterBrandCode.isNotEmpty ||
                              _filterCategoryCode.isNotEmpty || _filterPriceMin != null || _filterPriceMax != null)
                          ? global.theme.primaryColor
                          : global.theme.textSecondaryColor,
                    ),
                    tooltip: global.language('filter'),
                    onPressed: () {
                      setState(() {
                        _showFilterPanel = !_showFilterPanel;
                      });
                    },
                  ),
                  // Image toggle
                  IconButton(
                    focusNode: FocusNode(skipTraversal: true),
                    icon: Icon(
                      (showImage) ? Icons.image_not_supported : Icons.image,
                      size: 20,
                    ),
                    onPressed: () {
                      setState(() {
                        showImage = !showImage;
                      });
                    },
                  ),
                  ListFontSizeControl(onChanged: () => setState(() {})),
                  const SizedBox(width: 4),
                  // Total items
                  if (_totalItems > 0)
                    Text(
                      '($_totalItems)',
                      style: TextStyle(fontSize: 11, color: global.theme.textSecondaryColor),
                    ),
                ],
              ),
            ),
            // Advanced filter panel
            if (_showFilterPanel)
              _buildFilterPanel(),
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
                    flex: 6,
                    child: Text(
                      global.language("barcode"),
                      style: TextStyle(
                        color: global.theme.textColor,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 10,
                    child: Text(
                      global.language("product_name"),
                      style: TextStyle(
                        color: global.theme.textColor,
                        fontWeight: FontWeight.bold,
                      ),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  Expanded(
                    flex: 3,
                    child: Text(
                      global.language("unit"),
                      style: TextStyle(
                        color: global.theme.textColor,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 4,
                    child: Text(
                      global.language("item_code"),
                      style: TextStyle(
                        color: global.theme.textColor,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 4,
                    child: Text(
                      global.language("balance"),
                      style: TextStyle(
                        color: global.theme.textColor,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                  Expanded(
                    flex: 4,
                    child: Text(
                      global.language("price_retail"),
                      style: TextStyle(
                        color: global.theme.textColor,
                        fontWeight: FontWeight.bold,
                      ),
                      textAlign: TextAlign.right,
                    ),
                  ),
                  if (showImage)
                    Expanded(
                      flex: 1,
                      child: Icon(Icons.image, color: global.theme.textColor, size: 12),
                    ),
                  if (showCheckBox)
                    Expanded(
                      flex: 1,
                      child: Icon(Icons.check, color: global.theme.textColor, size: 12),
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

  Widget _buildFilterPanel() {
    bool hasFilters = _filterGroupCode.isNotEmpty ||
        _filterBrandCode.isNotEmpty ||
        _filterCategoryCode.isNotEmpty ||
        _filterPriceMin != null ||
        _filterPriceMax != null;

    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: global.theme.surfaceColor,
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: global.theme.dividerBorderColor),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Row 1: Group + Brand
          Row(
            children: [
              Expanded(
                child: _buildFilterChip(
                  icon: Icons.workspaces,
                  label: _filterGroupName.isNotEmpty
                      ? _filterGroupName
                      : global.language('product_group'),
                  isActive: _filterGroupCode.isNotEmpty,
                  onTap: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) => ProductGroupSearchScreen(word: ''),
                      ),
                    ).then((result) {
                      if (result != null && result is ProductGroupModel) {
                        setState(() {
                          _filterGroupCode = result.code;
                          _filterGroupName = result.names.isNotEmpty ? result.names[0].name : result.code;
                          listData = [];
                        });
                        loadDataList(searchText, filterBarcode);
                      }
                    });
                  },
                  onClear: _filterGroupCode.isNotEmpty
                      ? () {
                          setState(() {
                            _filterGroupCode = '';
                            _filterGroupName = '';
                            listData = [];
                          });
                          loadDataList(searchText, filterBarcode);
                        }
                      : null,
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: _buildFilterChip(
                  icon: Icons.branding_watermark,
                  label: _filterBrandName.isNotEmpty
                      ? _filterBrandName
                      : global.language('brand'),
                  isActive: _filterBrandCode.isNotEmpty,
                  onTap: () {
                    GroupSelectionHelper.showMasterBrandSelectionScreen(
                      context: context,
                      onItemSelected: (MasterBrandModel result) {
                        setState(() {
                          _filterBrandCode = result.code;
                          _filterBrandName = result.names.isNotEmpty
                              ? result.names[0].name
                              : result.code;
                          listData = [];
                        });
                        loadDataList(searchText, filterBarcode);
                      },
                    );
                  },
                  onClear: _filterBrandCode.isNotEmpty
                      ? () {
                          setState(() {
                            _filterBrandCode = '';
                            _filterBrandName = '';
                            listData = [];
                          });
                          loadDataList(searchText, filterBarcode);
                        }
                      : null,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          // Row 2: Category + Price range
          Row(
            children: [
              Expanded(
                child: _buildFilterChip(
                  icon: Icons.category,
                  label: _filterCategoryName.isNotEmpty
                      ? _filterCategoryName
                      : global.language('category'),
                  isActive: _filterCategoryCode.isNotEmpty,
                  onTap: () {
                    GroupSelectionHelper.showMasterCategorySelectionScreen(
                      context: context,
                      onItemSelected: (MasterCategoryModel result) {
                        setState(() {
                          _filterCategoryCode = result.code;
                          _filterCategoryName = result.names.isNotEmpty
                              ? result.names[0].name
                              : result.code;
                          listData = [];
                        });
                        loadDataList(searchText, filterBarcode);
                      },
                    );
                  },
                  onClear: _filterCategoryCode.isNotEmpty
                      ? () {
                          setState(() {
                            _filterCategoryCode = '';
                            _filterCategoryName = '';
                            listData = [];
                          });
                          loadDataList(searchText, filterBarcode);
                        }
                      : null,
                ),
              ),
              const SizedBox(width: 8),
              Expanded(
                child: Row(
                  children: [
                    Expanded(
                      child: TextField(
                        controller: _priceMinController,
                        keyboardType: TextInputType.number,
                        decoration: InputDecoration(
                          isDense: true,
                          contentPadding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
                          hintText: global.language('price_min'),
                          hintStyle: const TextStyle(fontSize: 12),
                          border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                          filled: true,
                          fillColor: global.theme.cardColor,
                        ),
                        style: TextStyle(fontSize: 12),
                        onSubmitted: (_) => _applyPriceFilter(),
                      ),
                    ),
                    Padding(
                      padding: const EdgeInsets.symmetric(horizontal: 4),
                      child: Text('→', style: TextStyle(fontSize: 14, color: global.theme.textSecondaryColor)),
                    ),
                    Expanded(
                      child: TextField(
                        controller: _priceMaxController,
                        keyboardType: TextInputType.number,
                        decoration: InputDecoration(
                          isDense: true,
                          contentPadding: const EdgeInsets.symmetric(horizontal: 8, vertical: 8),
                          hintText: global.language('price_max'),
                          hintStyle: const TextStyle(fontSize: 12),
                          border: OutlineInputBorder(borderRadius: BorderRadius.circular(8)),
                          filled: true,
                          fillColor: global.theme.cardColor,
                        ),
                        style: const TextStyle(fontSize: 12),
                        onSubmitted: (_) => _applyPriceFilter(),
                      ),
                    ),
                    if (_filterPriceMin != null || _filterPriceMax != null)
                      InkWell(
                        onTap: () {
                          setState(() {
                            _priceMinController.clear();
                            _priceMaxController.clear();
                            _filterPriceMin = null;
                            _filterPriceMax = null;
                            listData = [];
                          });
                          loadDataList(searchText, filterBarcode);
                        },
                        child: Padding(
                          padding: const EdgeInsets.only(left: 4),
                          child: Icon(Icons.close, size: 18, color: global.theme.negativeHighlightTextColor),
                        ),
                      ),
                  ],
                ),
              ),
            ],
          ),
          // Clear all filters
          if (hasFilters)
            Padding(
              padding: const EdgeInsets.only(top: 8),
              child: InkWell(
                onTap: () {
                  setState(() {
                    _filterGroupCode = '';
                    _filterGroupName = '';
                    _filterBrandCode = '';
                    _filterBrandName = '';
                    _filterCategoryCode = '';
                    _filterCategoryName = '';
                    _filterPriceMin = null;
                    _filterPriceMax = null;
                    _priceMinController.clear();
                    _priceMaxController.clear();
                    listData = [];
                  });
                  loadDataList(searchText, filterBarcode);
                },
                child: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(Icons.clear_all, size: 16, color: global.theme.negativeHighlightTextColor),
                    const SizedBox(width: 4),
                    Text(
                      global.language('clear_filter'),
                      style: TextStyle(fontSize: 12, color: global.theme.negativeHighlightTextColor),
                    ),
                  ],
                ),
              ),
            ),
        ],
      ),
    );
  }

  Widget _buildFilterChip({
    required IconData icon,
    required String label,
    required bool isActive,
    required VoidCallback onTap,
    VoidCallback? onClear,
  }) {
    return InkWell(
      onTap: onTap,
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
        decoration: BoxDecoration(
          color: isActive ? global.theme.infoHighlightColor : global.theme.cardColor,
          borderRadius: BorderRadius.circular(8),
          border: Border.all(
            color: isActive ? global.theme.primaryColor : global.theme.dividerBorderColor,
          ),
        ),
        child: Row(
          children: [
            Icon(icon, size: 16, color: isActive ? global.theme.primaryColor : global.theme.textSecondaryColor),
            const SizedBox(width: 4),
            Expanded(
              child: Text(
                label,
                style: TextStyle(
                  fontSize: 12,
                  color: isActive ? global.theme.primaryColor : global.theme.textSecondaryColor,
                ),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            if (onClear != null)
              InkWell(
                onTap: onClear,
                child: Icon(Icons.close, size: 14, color: global.theme.negativeHighlightTextColor),
              ),
          ],
        ),
      ),
    );
  }

  void _applyPriceFilter() {
    setState(() {
      _filterPriceMin = double.tryParse(_priceMinController.text);
      _filterPriceMax = double.tryParse(_priceMaxController.text);
      listData = [];
    });
    loadDataList(searchText, filterBarcode);
  }

  Future<FiltterBarcodeModel> filterBox(
    FiltterBarcodeModel filterBarcode,
  ) async {
    await showDialog(
      context: context,
      builder: (BuildContext context) {
        return StatefulBuilder(
          builder: (context, setState) {
            return AlertDialog(
              title: Text(global.language("filter_product")),
              content: SizedBox(
                width: 600.0,
                height: 600.0,
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Divider(),
                    RadioListTile(
                      title: Text(global.language("branch_selected")),
                      value: true,
                      groupValue: filterBarcode.branch,
                      onChanged: (value) {
                        setState(() {
                          filterBarcode.branch = value!;
                        });
                      },
                    ),
                    RadioListTile(
                      title: Text(global.language("all")),
                      value: false,
                      groupValue: filterBarcode.branch,
                      onChanged: (value) {
                        setState(() {
                          filterBarcode.branch = value!;
                        });
                      },
                    ),
                  ],
                ),
              ),
              actions: [
                ElevatedButton(
                  onPressed: () {
                    Navigator.pop(context);
                  },
                  child: Text(global.language("confirm")),
                ),
                ElevatedButton(
                  style: ElevatedButton.styleFrom(backgroundColor: global.theme.buttonDangerColor),
                  onPressed: () {
                    Navigator.pop(context);
                  },
                  child: Text(global.language("clear")),
                ),
              ],
            );
          },
        );
      },
    );

    return filterBarcode;
  }

  void switchToEdit(ProductBarcodeModel value) {
    setState(() {
      selectBarcode = value.barcode ?? '';
      selectGuid = value.guidfixed;
      getData(value.barcode ?? '');
      headerEdit = global.language("edit");
      isSaveAllow = true;
      isEditMode = true;
      isPreview = false;
    });
  }

  void optionSearch() {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => const BarcodeSearchScreen(word: "", screen: ""),
      ),
    ).then((value) {
      if (value != null) {
        setState(() {
          ProductBarcodeModel result = value;
          if (result.guidfixed.isNotEmpty) {
            screenData.options!.clear();
            if (result.options!.isNotEmpty) {
              screenData.options = result.options!;
            } else {
              screenData.options!.add(
                ProductOptionModel(
                  guid: const Uuid().v4(),
                  choicetype: 0,
                  minselect: 1,
                  maxselect: 1,
                  names: names,
                  choices: [],
                ),
              );
            }
          }
        });
      }
    });
  }

  void barcodeSearch(String word, int optionIndex, int choiceIndex) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => BarcodeSearchScreen(word: word, screen: ""),
      ),
    ).then((value) {
      if (value != null) {
        setState(() {
          ProductBarcodeModel result = value;
          if (result.guidfixed.isNotEmpty) {
            screenData
                    .options![optionIndex]
                    .choices[choiceIndex]
                    .refproductcode =
                result.itemcode!;
            screenData.options![optionIndex].choices[choiceIndex].refbarcode =
                result.barcode!;
            screenData
                    .options![optionIndex]
                    .choices[choiceIndex]
                    .refbarcodenames =
                result.names!;
            screenData.options![optionIndex].choices[choiceIndex].refunitcode =
                result.itemunitcode;
            screenData.options![optionIndex].choices[choiceIndex].refunitnames =
                result.itemunitnames!;
            screenData.options![optionIndex].choices[choiceIndex].vatcal =
                result.vatcal;
          }
        });
      }
    });
  }

  void orderTypeSearch(int orderTypeIndex) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => const OrderTypeSearchScreen(word: ''),
      ),
    ).then((value) {
      if (value != null) {
        setState(() {
          OrderTypeModel result = value;
          if (result.code.isNotEmpty) {
            screenData.ordertypes![orderTypeIndex].guidfixed =
                result.guidfixed!;
            screenData.ordertypes![orderTypeIndex].code = result.code;
            screenData.ordertypes![orderTypeIndex].names = result.names;
          }
        });
      }
    });
  }

  void dimensionSearch(int dimensionIndex) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => const DimensionSearchScreen(word: ''),
      ),
    ).then((value) {
      if (value != null) {
        setState(() {
          /// print log json result
          if (kDebugMode) {
            AppLogger.debug(jsonEncode(value.toJson()));
          }

          DimensionModel result = value;
          if (result.guidfixed!.isNotEmpty) {
            screenData.dimensions![dimensionIndex].guidfixed =
                result.guidfixed!;
            screenData.dimensions![dimensionIndex].names = result.names;
            screenData.dimensions![dimensionIndex].isdisabled =
                result.isdisabled;

            dimensionSeleted[dimensionIndex] = result;
          }
        });
      }
    });
  }

  Future<ProductBarcodeModel> subbarcodeSearch() async {
    ProductBarcodeModel res = ProductBarcodeModel(guidfixed: '', itemcode: '');
    res = await Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) =>
            const BarcodeSearchScreen(word: '', screen: 'subbarcodeSearch'),
      ),
    );
    return res;
  }

  void productSearch(String word, int optionIndex, int choiceIndex) {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (context) => ProductSearchScreen(word: word)),
    ).then((value) {
      if (value != null) {
        setState(() {
          ProductModel result = value;
          if (result.guidfixed.isNotEmpty) {
            screenData
                    .options![optionIndex]
                    .choices[choiceIndex]
                    .refproductcode =
                result.itemcode;
          }
        });
      }
    });
  }

  void loadDataBusinessTypesList() {
    context.read<BusinessTypeBloc>().add(
      const BusinessTypeLoadList(offset: 0, limit: 1000, search: ""),
    );
  }

  Future<List<BusinessTypeModel>> getDataBusiness(filter) async {
    return listDataBusinessType;
  }

  void loadDataBranchList(List<BusinessTypeModel> businessType) {
    listDataBranchBusinessType = [];

    // Extract the 'code' value from each element and concatenate them
    String businesstypecode = businessType.map((item) => item.code).join(',');

    if (businesstypecode.isNotEmpty) {
      context.read<CompanyBranchBloc>().add(
        CompanyBranchByBusinessTypeLoadList(
          offset: 0,
          limit: 1000,
          search: "",
          businesstypecode: businesstypecode,
        ),
      );
    }
  }


  void _selectMediaFromTime(BuildContext context, int mediaIndex) async {
    final TimeOfDay? pickedTime = await showTimePicker(
      context: context,
      initialTime: _selectedTime[mediaIndex],
      builder: (BuildContext context, Widget? child) {
        return MediaQuery(
          data: MediaQuery.of(context).copyWith(alwaysUse24HourFormat: true),
          child: child!,
        );
      },
    );

    // ignore: unrelated_type_equality_checks
    if (pickedTime != null) {
      setState(() {
        // Format the hour and minute to ensure they have two digits (e.g., 09, 12, 23, etc.)
        String formattedHour = pickedTime.hour.toString().padLeft(2, '0');
        String formattedMinute = pickedTime.minute.toString().padLeft(2, '0');

        // Combine the formatted hour and minute with a ":" separator
        String formattedTime = '$formattedHour:$formattedMinute';

        timeForSales[mediaIndex].fromtime = formattedTime;
        mediaFromTimeController[mediaIndex].text = formattedTime;
      });
    }
  }

  void _selectMediaToTime(BuildContext context, int mediaIndex) async {
    final TimeOfDay? pickedTime = await showTimePicker(
      context: context,
      initialTime: _selectedTime[mediaIndex],
      builder: (BuildContext context, Widget? child) {
        return MediaQuery(
          data: MediaQuery.of(context).copyWith(alwaysUse24HourFormat: true),
          child: child!,
        );
      },
    );

    // ignore: unrelated_type_equality_checks
    if (pickedTime != null) {
      setState(() {
        // Format the hour and minute to ensure they have two digits (e.g., 09, 12, 23, etc.)
        String formattedHour = pickedTime.hour.toString().padLeft(2, '0');
        String formattedMinute = pickedTime.minute.toString().padLeft(2, '0');

        // Combine the formatted hour and minute with a ":" separator
        String formattedTime = '$formattedHour:$formattedMinute';

        // Convert the time string to a TimeOfDay object
        TimeOfDay fromTime = global.getTimeOfDayFromString(
          timeForSales[mediaIndex].fromtime!,
        );
        TimeOfDay toTime = global.getTimeOfDayFromString(formattedTime);

        // Check if toTime is not earlier than fromTime
        if (toTime.hour < fromTime.hour ||
            (toTime.hour == fromTime.hour &&
                toTime.minute <= fromTime.minute)) {
          global.showSnackBar(
            context,
            Icon(Icons.edit, color: global.theme.onPrimaryColor),
            global.language("to_time_must_be_later_than_from_time"),
            global.theme.buttonDangerColor,
          );
          timeForSales[mediaIndex].totime = '';
          mediaToTimeController[mediaIndex].text = '';
          return;
        } else {
          // Update the toTime and the text in the text field
          timeForSales[mediaIndex].totime = formattedTime;
          mediaToTimeController[mediaIndex].text = formattedTime;
        }
      });
    }
  }

  Future<List<DayOfWeekModel>> getDataDayOfWeek(filter) async {
    return dayOfWeekList;
  }

  Future<List<BranchModel>> getDataBranch(filter) async {
    return listDataBranchBusinessType;
  }

  Widget listObject(int index, ProductBarcodeModel value, bool showCheckBox) {
    bool isCheck = false;
    final isSelected = selectBarcode == (value.barcode ?? '');
    TextStyle textStyle = isSelected
        ? TextStyle(
            fontSize: global.deviceConfig.listDataFontSize,
            fontWeight: FontWeight.bold,
            color: global.theme.textColor,
          )
        : TextStyle(
            fontSize: global.deviceConfig.listDataFontSize,
            color: global.theme.textSecondaryColor,
          );
    if (showCheckBox) {
      for (int i = 0; i < guidListChecked.length; i++) {
        if (guidListChecked[i] == (value.barcode ?? '')) {
          isCheck = true;
          break;
        }
      }
    }
    // ลบการ add key ออก - keys ถูกสร้างใน LoadSuccess แล้ว
    // listKeys.add(GlobalKey());  // ❌ ตรงนี้ทำให้เกิด infinite loop!
    return MouseRegion(
      cursor: SystemMouseCursors.click,
      onEnter: (_) => setState(() => _hoverIndex = index),
      onExit: (_) => setState(() => _hoverIndex = -1),
      child: GestureDetector(
      onTap: () {
        if (showCheckBox == true) {
          setState(() {
            selectBarcode = value.barcode ?? '';
            selectGuid = value.guidfixed;
            if (isCheck == true) {
              guidListChecked.remove(value.barcode ?? '');
            } else {
              guidListChecked.add(value.barcode ?? '');
            }
            global.showSnackBar(
              context,
              Icon(Icons.check, color: global.theme.onPrimaryColor),
              "${global.language("chosen")} ${guidListChecked.length} ${global.language("list")}",
              global.theme.primaryColor,
            );
          });
        } else {
          setState(() {
            discardData(
              callBack: () {
                isSaveAllow = false;
                isEditMode = false;
                selectBarcode = value.barcode ?? '';
                selectGuid = value.guidfixed;

                getData(value.barcode ?? '');
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
          isPreview = false;
          switchToEdit(value);
        }
      },
      child: Container(
        key: index < listKeys.length ? listKeys[index] : null,
        color: _getContainerColor(selectBarcode, value.barcode ?? '', index),
        padding: EdgeInsets.only(
          left: 10,
          right: 10,
          top: global.deviceConfig.listDataLineSpace,
          bottom: global.deviceConfig.listDataLineSpace,
        ),
        child: Column(
          children: [
            _buildRow(
              value,
              textStyle,
              showImage,
              tooltipKeys[index],
              showCheckBox,
              isCheck,
            ),
            (filterBarcode.branch == false &&
                    (appConfig.getInt("branch_total") != 1))
                ? SizedBox(
                    width: double.infinity,
                    child: _buildChipWrap(value.branches!),
                  )
                : Container(),
          ],
        ),
      ),
    ),
    );
  }

  Color? _getContainerColor(String selectedBarcode, String itemBarcode, int index) {
    if (selectedBarcode.isNotEmpty && selectedBarcode == itemBarcode) {
      return isEditMode ? global.theme.rowEditColor : global.theme.rowSelectedColor;
    }
    if (_hoverIndex == index) {
      return global.theme.rowHoverColor;
    }
    return (index % 2 == 0)
        ? global.theme.columnAlternateEvenColor
        : global.theme.columnAlternateOddColor;
  }

  /// แสดงหน่วยนับ + icon หลายหน่วย + sub-units
  Widget _buildUnitCell(ProductBarcodeModel value, TextStyle textStyle) {
    final unitName = global.activeLangName(value.itemunitnames!);
    final isMultiUnit = value.unitCount > 1;

    if (!isMultiUnit) {
      // หน่วยนับเดียว
      return Row(
        children: [
          Icon(Icons.inventory_2_outlined, size: 12, color: global.theme.textSecondaryColor),
          const SizedBox(width: 4),
          Expanded(
            child: Text(
              unitName,
              style: textStyle,
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),
        ],
      );
    }

    // หลายหน่วยนับ
    return Tooltip(
      message: value.allUnitNames,
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        mainAxisSize: MainAxisSize.min,
        children: [
          Row(
            children: [
              Icon(Icons.layers, size: 12, color: global.theme.primaryColor),
              const SizedBox(width: 4),
              Expanded(
                child: Text(
                  unitName,
                  style: textStyle.copyWith(color: global.theme.primaryColor),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
                decoration: BoxDecoration(
                  color: global.theme.infoHighlightColor,
                  borderRadius: BorderRadius.circular(8),
                  border: Border.all(color: global.theme.infoHighlightTextColor.withValues(alpha: 0.3)),
                ),
                child: Text(
                  '${value.unitCount}',
                  style: TextStyle(fontSize: 9, color: global.theme.primaryColor, fontWeight: FontWeight.bold),
                ),
              ),
            ],
          ),
          if (value.allUnitNames.isNotEmpty)
            Padding(
              padding: EdgeInsets.only(left: 16, top: 1),
              child: Text(
                value.allUnitNames,
                style: TextStyle(fontSize: 10, color: global.theme.textSecondaryColor),
              ),
            ),
        ],
      ),
    );
  }

  Row _buildRow(
    ProductBarcodeModel value,
    TextStyle textStyle,
    bool showImage,
    GlobalKey<ImageTooltipState> key,
    bool showCheckBox,
    bool isCheck,
  ) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Expanded(
          flex: 6,
          child: Text(
            value.barcode!,
            style: textStyle,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
        ),
        Expanded(
          flex: 10,
          child: Text(
            global.activeLangName(value.names!),
            style: textStyle,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
        ),
        Expanded(
          flex: 3,
          child: _buildUnitCell(value, textStyle),
        ),
        Expanded(
          flex: 4,
          child: Text(
            value.itemcode!,
            style: textStyle,
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
        ),
        Expanded(
          flex: 4,
          child: Text(
            value.balanceFormatted.isNotEmpty
                ? value.balanceFormatted
                : '0',
            style: textStyle.copyWith(
              color: value.balanceQty <= 0 ? global.theme.negativeHighlightTextColor : null,
            ),
          ),
        ),
        Expanded(
          flex: 4,
          child: Text(
            _getRetailPrice(value),
            style: textStyle,
            textAlign: TextAlign.right,
          ),
        ),
        if (showImage) _buildImageWidget(value, key),
        if (showCheckBox)
          Expanded(
            flex: 1,
            child: (isCheck) ? const Icon(Icons.check, size: 12) : Container(),
          ),
      ],
    );
  }

  String _getRetailPrice(ProductBarcodeModel value) {
    if (value.prices == null || value.prices!.isEmpty) return '0';
    for (final p in value.prices!) {
      if (p.keynumber == 1) {
        if (p.price <= 0) return '0';
        return global.moneyFormat.format(p.price);
      }
    }
    return '0';
  }

  Widget _buildImageWidget(
    ProductBarcodeModel value,
    GlobalKey<ImageTooltipState> key,
  ) {
    return Expanded(
      flex: 1,
      child:
          (value.imageuri != null && value.imageuri!.isNotEmpty)
          ? ImageTooltip(
              key: key,
              image: Image.network(global.resolveFileUrl(value.imageuri!)),
              child: _buildImageContainer(global.resolveFileUrl(value.imageuri!)),
            )
          : Container(),
    );
  }

  Container _buildImageContainer(String imageUri) {
    return Container(
      decoration: BoxDecoration(
        color: global.theme.searchBarColor,
        borderRadius: BorderRadius.circular(2),
        border: Border.all(color: global.theme.textSecondaryColor, width: 1),
        boxShadow: [
          BoxShadow(
            color: global.theme.textSecondaryColor.withValues(alpha: 0.5),
            spreadRadius: 1,
            blurRadius: 1,
            offset: const Offset(1, 1),
          ),
        ],
      ),
      padding: const EdgeInsets.all(1),
      child: ClipRRect(
        borderRadius: BorderRadius.circular(2),
        child: Image.network(
          imageUri,
          fit: BoxFit.fitHeight,
          height: 20,
          errorBuilder:
              (BuildContext context, Object exception, StackTrace? stackTrace) {
                return const Icon(Icons.image_not_supported);
              },
        ),
      ),
    );
  }

  Wrap _buildChipWrap(List<CompanyBranchModel> branches) {
    return Wrap(
      spacing: 2.0,
      runSpacing: 2.0,
      children: branches
          .asMap()
          .entries
          .map((entry) => _buildChip(entry.value))
          .toList(),
    );
  }

  Chip _buildChip(CompanyBranchModel branch) {
    return Chip(
      materialTapTargetSize: MaterialTapTargetSize.shrinkWrap,
      padding: const EdgeInsets.all(0),
      label: Text(
        '${branch.code} - ${global.activeLangName(branch.names)}',
        style: const TextStyle(fontSize: 12),
      ),
    );
  }

  void saveOrUpdateData() {
    if (timeForSales.isNotEmpty) {
      for (int i = 0; i < timeForSales.length; i++) {
        if (mediaFromDateController[i].text.isNotEmpty) {
          DateTime fromDateUtc = DateTime.parse(timeForSales[i].fromdate!);
          timeForSales[i].fromdate = fromDateUtc.toUtc().toIso8601String();
        }

        if (mediaToDateController[i].text.isNotEmpty) {
          DateTime toDateUtc = DateTime.parse(timeForSales[i].todate!);
          timeForSales[i].todate = toDateUtc.toUtc().toIso8601String();
        }

        timeForSales[i].daysofweek = [];
        for (var element in dayOfWeekSeleted[i]) {
          timeForSales[i].daysofweek!.add(int.parse(element.code));
        }
      }

      screenData.timeforsales = timeForSales;
    }

    if (fiexdCostList.isNotEmpty) {
      for (int i = 0; i < fiexdCostList.length; i++) {
        if (asDateController[i].text.isNotEmpty) {
          DateTime asDateUtc = DateTime.parse(fiexdCostList[i].effectdate!);
          fiexdCostList[i].effectdate = asDateUtc.toUtc().toIso8601String();
        }

        fiexdCostAmountController[i].text = global.formatNumber(
          fiexdCostList[i].amount!,
        );
      }

      screenData.fixedcost = fiexdCostList;
    }

    if (screenData.itemtype != 2) {
      for (var element in screenData.refbarcodes!) {
        if (!element.condition) {
          element.standvalue = double.parse(element.qty.toString());
          element.dividevalue = 1;
          element.qty = 1;
        } else {
          element.standvalue = 1;
          element.dividevalue = double.parse(element.qty.toString());
          element.qty = 1;
        }
      }
    }

    /// fillter data company branch in listDataBranchBusinessType where screenData.ignorebranches ! = listDataBranchBusinessType
    /// ค้นหา สาขา ที่ไม่ตรงกับ ui ที่เลือกเพื่อส่งค่าไป ignore
    List<BranchModel> listDataBranchBusinessTypeFilter = [];
    for (var element in listDataBranchBusinessType) {
      bool isFound = false;
      for (var ignoreBranch in ignorebranches!) {
        if (element.guidfixed == ignoreBranch.guidfixed) {
          isFound = true;
          break;
        }
      }
      if (!isFound) {
        listDataBranchBusinessTypeFilter.add(element);
      }
    }

    screenData.ignorebranches = listDataBranchBusinessTypeFilter;

    if (kDebugMode) {
      AppLogger.debug(jsonEncode(screenData.toJson()));
    }

    if (descriptionTextEditController.text.isNotEmpty) {
      screenData.description = descriptionTextEditController.text;
    }

    if (alertDescriptionTextEditController.text.isNotEmpty) {
      screenData.alertdescription = alertDescriptionTextEditController.text;
    }

    if (screenData.refbarcodes!.isNotEmpty) {
      screenData.ismainbarcode = false;
    }

    showCheckBox = false;
    // ใช้ screenData.guidfixed (จาก MongoDB) สำหรับ save/update
    final guid = screenData.guidfixed.trim();
    if (guid.isEmpty) {
      if (imageFile.path.isNotEmpty) {
        context.read<ProductBarcodeBloc>().add(
          ProductBarcodeWithImageSave(
            productBarcode: screenData,
            imageFile: imageFile,
            imageWeb: imageWeb,
          ),
        );
      } else {
        context.read<ProductBarcodeBloc>().add(
          ProductBarcodeSave(productBarcode: screenData),
        );
      }
    } else {
      updateData(guid);
    }
  }

  void updateData(String guid) {
    showCheckBox = false;
    if (imageWeb != null) {
      context.read<ProductBarcodeBloc>().add(
        ProductBarcodeWithImageUpdate(
          guid: guid,
          productBarcode: screenData,
          imageFile: imageFile,
          imageWeb: imageWeb!,
        ),
      );
    } else {
      context.read<ProductBarcodeBloc>().add(
        ProductBarcodeUpdate(guid: guid, productBarcode: screenData),
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

  void searchUnit({required String word, required Function callBack}) {
    Navigator.push(
      context,
      MaterialPageRoute(builder: (context) => UnitSearchScreen(word: word)),
    ).then((value) {
      global.SearchCodeNameModel result = value;

      if (result.code.trim().isNotEmpty) {
        callBack(true, result.code, result.names);
      }
      if (result.isCancel == false) findFocusNext(focusNodeIndex);
    });
  }

  void setFocusNode(FocusNode focus) {
    focus.unfocus();
    Future.delayed(const Duration(milliseconds: 500), () {
      focusAndCursorToEnd(focus);
    });
  }

  Future<void> translateNames() async {
    for (var name in screenData.names!) {
      if (name.name.isNotEmpty) {
        var translatedNames = await global.translateNames(
          namesData: screenData.names!,
        );
        setState(() => screenData.names = translatedNames);
        break;
      }
    }
  }

  Future<void> translateOptions() async {
    for (var option in screenData.options!) {
      await translateOptionNames(option);
      await translateChoices(option);
    }
  }

  Future<void> translateOptionNames(ProductOptionModel option) async {
    for (var name in option.names) {
      if (name.name.isEmpty) {
        var translatedNames = await global.translateNames(
          namesData: option.names,
        );
        setState(() => option.names = translatedNames);
        break;
      }
    }
  }

  Future<void> translateChoices(ProductOptionModel option) async {
    for (var choice in option.choices) {
      for (var name in choice.names) {
        if (name.name.isEmpty) {
          var translatedNames = await global.translateNames(
            namesData: choice.names,
          );
          setState(() => choice.names = translatedNames);
          break;
        }
      }
    }
  }

  Widget unitTextEditWidget({
    required String label,
    required String unitCode,
    required List<LanguageDataModel> unitNames,
    required bool isReadOnly,
    required bool enableIcon,
    required TextEditingController unitCodeController,
    required TextEditingController unitNameController,
    required Function callBack,
  }) {
    return Row(
      children: [
        Expanded(
          child: RawKeyboardListener(
            focusNode: FocusNode(),
            onKey: (RawKeyEvent event) {
              if (event is RawKeyDownEvent) {
                if (event.logicalKey == LogicalKeyboardKey.f2) {
                  searchUnit(word: unitCodeController.text, callBack: callBack);
                }
              }
            },
            child: TextFormField(
              onEditingComplete: () {
                if (kIsWeb) {
                  findFocusNext(focusNodeIndex);
                }
              },
              readOnly: isReadOnly,
              onChanged: (code) {
                isDataChange = true;
                if (code.trim().isNotEmpty) {
                  UnitRepository().getUnitManyByCode([code.trim()]).then((
                    value,
                  ) {
                    for (var data in value.data) {
                      if (data != null) {
                        UnitModel unit = UnitModel.fromJson(data);
                        unitNameController.text = global.activeLangName(
                          unit.names,
                        );
                        callBack(false, code, unit.names);
                      } else {
                        unitNameController.text = "";
                        callBack(false, code, null);
                      }
                    }
                  });
                } else {
                  unitNameController.text = "";
                  callBack(false, null, null);
                }
              },
              focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
              textAlign: TextAlign.left,
              controller: unitCodeController,
              decoration: InputDecoration(
                contentPadding: EdgeInsets.all(10.0),
                hintText: global.language("must") + label,
                enabledBorder: OutlineInputBorder(
                  borderSide: BorderSide(color: global.theme.dividerBorderColor, width: 0.0),
                ),
                floatingLabelBehavior: FloatingLabelBehavior.always,
                border: const OutlineInputBorder(),
                labelText: label,
                suffixIcon: Row(
                  mainAxisAlignment:
                      MainAxisAlignment.spaceBetween, // added line
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    if (enableIcon)
                      IconButton(
                        focusNode: FocusNode(skipTraversal: true),
                        icon: const Icon(Icons.search),
                        onPressed: () {
                          searchUnit(
                            word: unitCodeController.text,
                            callBack: callBack,
                          );
                        },
                      ),
                  ],
                ),
              ),
              validator: (value) {
                if (value == null || value.isEmpty) {
                  return 'This field is required';
                }

                return null;
              },
            ),
          ),
        ),
        SizedBox(width: 5),
        Expanded(
          child: TextFormField(
            readOnly: true,
            focusNode: null,
            textAlign: TextAlign.left,
            controller: unitNameController,
            decoration: InputDecoration(
              contentPadding: EdgeInsets.all(10.0),
              enabledBorder: OutlineInputBorder(
                borderSide: BorderSide(color: global.theme.dividerBorderColor, width: 0.0),
              ),
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
      ],
    );
  }

  void productGroupSearch() {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) =>
            ProductGroupSearchScreen(word: screenData.groupcode!),
      ),
    ).then((value) {
      if (value != null) {
        setState(() {
          global.SearchCodeNameModel result = value;
          if (result.isCancel == false) {
            groupController.text = result.code;
            screenData.groupcode = result.code;
            screenData.groupnames = result.names;
          }
        });
      }
    });
  }

  void findProductGroup() {
    ProductGroupRepository()
        .getProductGroupManyByCode([screenData.groupcode!.trim()])
        .then((value) {
          screenData.groupnames = [];
          for (var data in value.data) {
            if (data != null) {
              ProductGroupModel productGroupModel = ProductGroupModel.fromJson(
                data,
              );
              screenData.groupnames = productGroupModel.names;
            }
            setState(() {});
          }
        });
  }

  Future<List<ItemDimension>> getDataItemDimension(
    filter,
    dimensionIndex,
  ) async {
    return dimensionSeleted[dimensionIndex].items!;
  }

  // Helper function to build master data selection widgets
  Widget buildMasterDataSelectionWidget({
    required String label,
    required String code,
    required String displayName,
    required VoidCallback onTap,
    required VoidCallback onClear,
    bool isSelected = false,
  }) {
    return Padding(
      padding: EdgeInsets.only(left: 10, right: 10, bottom: 4),
      child: InkWell(
        onTap: isEditMode ? onTap : null,
        child: Container(
          height: 36,
          decoration: BoxDecoration(
            border: Border.all(color: isSelected ? global.theme.primaryColor : global.theme.textSecondaryColor),
            borderRadius: BorderRadius.circular(4),
            color: isSelected ? global.theme.primaryColor.withValues(alpha: 0.08) : null,
          ),
          child: Row(
            children: [
              const SizedBox(width: 8),
              Text(
                '$label: ',
                style: TextStyle(
                  color: global.theme.textSecondaryColor,
                  fontSize: 12,
                ),
              ),
              Expanded(
                child: Text(
                  isSelected ? '$code ~ $displayName' : '-',
                  style: TextStyle(
                    color: isSelected ? global.theme.textColor : global.theme.textSecondaryColor,
                    fontSize: 13,
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              if (isSelected && isEditMode)
                InkWell(
                  onTap: onClear,
                  child: Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 4),
                    child: Icon(Icons.clear, color: global.theme.negativeHighlightTextColor, size: 16),
                  ),
                ),
              Icon(
                Icons.search,
                color: isEditMode ? global.theme.primaryColor : global.theme.textSecondaryColor,
                size: 16,
              ),
              const SizedBox(width: 6),
            ],
          ),
        ),
      ),
    );
  }

  Widget editScreen({mobileScreen}) {
    List<Widget> formWidgets = [];
    List<Widget> groupMenuWidgets = [];

    Widget productTypeRadio = Row(
      children: [
        Expanded(
          child: Row(
            children: [
              Radio(
                value: 0,
                groupValue: screenData.itemtype,
                onChanged: (value) {
                  if (isEditMode) {
                    setState(() {
                      screenData.itemtype = value;
                      if (screenData.refbarcodes!.length > 1) {
                        screenData.refbarcodes!.removeRange(
                          1,
                          screenData.refbarcodes!.length,
                        );

                        currentBarcodeNode = 0;
                      }
                    });
                  }
                },
              ),
              Expanded(
                child: Text(
                  global.language("product_is_stock"),
                  overflow: TextOverflow.clip,
                ),
              ),
            ],
          ),
        ),
        Expanded(
          child: Row(
            children: [
              Radio(
                value: 1,
                groupValue: screenData.itemtype,
                activeColor: global.theme.negativeHighlightTextColor,
                onChanged: (value) {
                  if (isEditMode) {
                    setState(() {
                      screenData.itemtype = value;
                      screenData.refbarcodes!.clear();
                      screenData.isusesubbarcodes = false;
                    });
                  }
                },
              ),
              Expanded(
                child: Text(
                  global.language("product_is_service"),
                  overflow: TextOverflow.clip,
                ),
              ),
            ],
          ),
        ),
        Expanded(
          child: Row(
            children: [
              Radio(
                value: 2,
                groupValue: screenData.itemtype,
                activeColor: global.theme.warningHighlightTextColor,
                onChanged: (value) {
                  if (isEditMode) {
                    setState(() {
                      screenData.itemtype = value;
                    });
                  }
                },
              ),
              Expanded(
                child: Text(
                  global.language("product_is_set"),
                  overflow: TextOverflow.clip,
                ),
              ),
            ],
          ),
        ),
        if (global.posVersion == global.PosVersionEnum.restaurant)
          Expanded(
            child: Row(
              children: [
                Radio(
                  value: 3,
                  groupValue: screenData.itemtype,
                  activeColor: global.theme.negativeHighlightTextColor,
                  onChanged: (value) {
                    if (isEditMode) {
                      setState(() {
                        screenData.itemtype = value;
                        screenData.refbarcodes!.clear();
                        screenData.isusesubbarcodes = false;
                      });
                    }
                  },
                ),
                Expanded(
                  child: Text(
                    global.language("product_is_not_stock"),
                    overflow: TextOverflow.clip,
                  ),
                ),
              ],
            ),
          ),
      ],
    );

    /// ประเภทวัตถุดิบ
    Widget materialTypeRadio = Row(
      children: [
        Expanded(
          child: Row(
            children: [
              Radio(
                value: 0,
                groupValue: screenData.materialtype,
                activeColor: global.theme.negativeHighlightTextColor,
                onChanged: (value) {
                  if (isEditMode) {
                    setState(() {
                      screenData.materialtype = value;
                    });
                  }
                },
              ),
              Expanded(
                child: Text(
                  global.language("product_general"),
                  overflow: TextOverflow.clip,
                ),
              ),
            ],
          ),
        ),
        Expanded(
          child: Row(
            children: [
              Radio(
                value: 1,
                groupValue: screenData.materialtype,
                activeColor: global.theme.negativeHighlightTextColor,
                onChanged: (value) {
                  if (isEditMode) {
                    setState(() {
                      screenData.materialtype = value;
                    });
                  }
                },
              ),
              Expanded(
                child: Text(
                  global.language("product_material"),
                  overflow: TextOverflow.clip,
                ),
              ),
            ],
          ),
        ),
        Expanded(
          child: Row(
            children: [
              Radio(
                value: 2,
                groupValue: screenData.materialtype,
                activeColor: global.theme.warningHighlightTextColor,
                onChanged: (value) {
                  if (isEditMode) {
                    setState(() {
                      screenData.materialtype = value;
                    });
                  }
                },
              ),
              Expanded(
                child: Text(
                  global.language("product_semi_finished"),
                  overflow: TextOverflow.clip,
                ),
              ),
            ],
          ),
        ),
      ],
    );

    Widget foodTypeRadio = InputDecorator(
      decoration: InputDecoration(
        labelText: global.language("food_type"),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(4.0),
          borderSide: BorderSide(color: global.theme.dividerBorderColor, width: 0.0),
        ),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.start,
        children: [
          Expanded(
            child: Row(
              children: [
                Radio(
                  value: 0,
                  groupValue: screenData.foodtype,
                  onChanged: (value) {
                    if (isEditMode) {
                      setState(() {
                        screenData.foodtype = value;
                      });
                    }
                  },
                ),
                Expanded(
                  child: Text(
                    global.language("food"),
                    overflow: TextOverflow.clip,
                  ),
                ),
              ],
            ),
          ),
          Expanded(
            child: Row(
              children: [
                Radio(
                  value: 1,
                  groupValue: screenData.foodtype,
                  activeColor: global.theme.negativeHighlightTextColor,
                  onChanged: (value) {
                    if (isEditMode) {
                      setState(() {
                        screenData.foodtype = value;
                      });
                    }
                  },
                ),
                Expanded(
                  child: Text(
                    global.language("drink"),
                    overflow: TextOverflow.clip,
                  ),
                ),
              ],
            ),
          ),
          Expanded(
            child: Row(
              children: [
                Radio(
                  value: 2,
                  groupValue: screenData.foodtype,
                  activeColor: global.theme.warningHighlightTextColor,
                  onChanged: (value) {
                    if (isEditMode) {
                      setState(() {
                        screenData.foodtype = value;
                      });
                    }
                  },
                ),
                Expanded(
                  child: Text(
                    global.language("alcohol"),
                    overflow: TextOverflow.clip,
                  ),
                ),
              ],
            ),
          ),
          Expanded(
            child: Row(
              children: [
                Radio(
                  value: 3,
                  groupValue: screenData.foodtype,
                  activeColor: global.theme.negativeHighlightTextColor,
                  onChanged: (value) {
                    if (isEditMode) {
                      setState(() {
                        screenData.foodtype = value;
                      });
                    }
                  },
                ),
                Expanded(
                  child: Text(
                    global.language("other"),
                    overflow: TextOverflow.clip,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );

    Widget vatTypeRadio = InputDecorator(
      decoration: InputDecoration(
        labelText: global.language("vat_type"),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(4.0),
          borderSide: BorderSide(color: global.theme.dividerBorderColor, width: 0.0),
        ),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.start,
        children: [
          Expanded(
            child: Row(
              children: [
                Radio(
                  value: 0,
                  groupValue: screenData.vatcal,
                  onChanged: (value) {
                    if (isEditMode) {
                      setState(() {
                        screenData.vatcal = value;
                      });
                    }
                  },
                ),
                Expanded(
                  child: Text(
                    global.language("product_vat_type_1"),
                    overflow: TextOverflow.clip,
                  ),
                ),
              ],
            ),
          ),
          Expanded(
            child: Row(
              children: [
                Radio(
                  value: 1,
                  groupValue: screenData.vatcal,
                  activeColor: global.theme.negativeHighlightTextColor,
                  onChanged: (value) {
                    if (isEditMode) {
                      setState(() {
                        screenData.vatcal = value;
                      });
                    }
                  },
                ),
                Expanded(
                  child: Text(
                    global.language("product_vat_type_2"),
                    overflow: TextOverflow.clip,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );

    Widget issumPointRadio = InputDecorator(
      decoration: InputDecoration(
        labelText: global.language("issumpoint"),
        enabledBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(4.0),
          borderSide: BorderSide(color: global.theme.dividerBorderColor, width: 0.0),
        ),
      ),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.start,
        children: [
          Expanded(
            child: Row(
              children: [
                Radio(
                  value: true,
                  groupValue: screenData.issumpoint,
                  onChanged: (value) {
                    if (isEditMode) {
                      setState(() {
                        screenData.issumpoint = value;
                      });
                    }
                  },
                ),
                Expanded(
                  child: Text(
                    global.language("product_use_point_1"),
                    overflow: TextOverflow.clip,
                  ),
                ),
              ],
            ),
          ),
          Expanded(
            child: Row(
              children: [
                Radio(
                  value: false,
                  groupValue: screenData.issumpoint,
                  activeColor: global.theme.negativeHighlightTextColor,
                  onChanged: (value) {
                    if (isEditMode) {
                      setState(() {
                        screenData.issumpoint = value;
                      });
                    }
                  },
                ),
                Expanded(
                  child: Text(
                    global.language("product_use_point_2"),
                    overflow: TextOverflow.clip,
                  ),
                ),
              ],
            ),
          ),
        ],
      ),
    );

    global.deviceConfigLoad();
    groupMenuWidgets.add(
      Expanded(
        child: ElevatedButton(
          child: Text(
            global.language("item_display_all"),
            overflow: TextOverflow.clip,
          ),
          onPressed: () {
            setState(() {
              global.deviceConfig.itemDisplaySku = true;
              global.deviceConfig.itemDisplayPrice = true;
              global.deviceConfigSaveJson();
            });
          },
        ),
      ),
    );
    groupMenuWidgets.add( SizedBox(width: 5));
    groupMenuWidgets.add(
      Expanded(
        child: ElevatedButton(
          child: Row(
            children: [
              (global.deviceConfig.itemDisplaySku)
                  ? const Icon(Icons.check)
                  : const Icon(Icons.error),
              SizedBox(width: 5),
              Expanded(
                child: Text(
                  global.language("item_display_sku"),
                  overflow: TextOverflow.clip,
                ),
              ),
            ],
          ),
          onPressed: () {
            setState(() {
              global.deviceConfig.itemDisplaySku =
                  !global.deviceConfig.itemDisplaySku;
              global.deviceConfigSaveJson();
            });
          },
        ),
      ),
    );
    groupMenuWidgets.add( SizedBox(width: 5));
    groupMenuWidgets.add(
      Expanded(
        child: ElevatedButton(
          child: Row(
            children: [
              (global.deviceConfig.itemDisplayPrice)
                  ? const Icon(Icons.check)
                  : const Icon(Icons.error),
              SizedBox(width: 5),
              Expanded(
                child: Text(
                  global.language("item_display_price"),
                  overflow: TextOverflow.clip,
                ),
              ),
            ],
          ),
          onPressed: () {
            setState(() {
              global.deviceConfig.itemDisplayPrice =
                  !global.deviceConfig.itemDisplayPrice;
              global.deviceConfigSaveJson();
            });
          },
        ),
      ),
    );
    // formWidgets.add(Container(
    //     padding: const EdgeInsets.only(left: 10, right: 10, bottom: 10),
    //     child: IntrinsicHeight(child: Row(crossAxisAlignment: CrossAxisAlignment.stretch, children: groupMenuWidgets))));
    focusNodeMax = 0;
    // Barcode
    formWidgets.add(
      Padding(
        padding: const EdgeInsets.only(left: 10, right: 10, bottom: 10),
        child: TextFormField(
          enabled: (screenData.guidfixed.isEmpty && isEditMode),
          onEditingComplete: () {
            if (kIsWeb) {
              findFocusNext(focusNodeIndex);
            }
          },
          focusNode: fieldFocusNodes[focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: barcodeController,
          textCapitalization: TextCapitalization.characters,
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.allow(RegExp('[a-zA-Z0-9-]')),
            FilteringTextInputFormatter.deny(' '), // Exclude space character
          ],
          onChanged: (value) {
            isDataChange = true;
            screenData.barcode = value.toUpperCase();
            barcodeController.value = TextEditingValue(
              text: value.toUpperCase(),
              selection: barcodeController.selection,
            );
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("barcode"),
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
    formWidgets.addAll(listNamesFields(screenData.names!, "product_name"));

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
        child: unitTextEditWidget(
          unitCode: screenData.itemunitcode ?? "",
          unitNames: screenData.itemunitnames ?? [],
          unitCodeController: TextEditingController(
            text: screenData.itemunitcode ?? "",
          ),
          unitNameController: TextEditingController(
            text: global.activeLangName(screenData.itemunitnames!),
          ),
          label: global.language("unit"),
          isReadOnly: !isEditMode,
          enableIcon: isEditMode,
          callBack: (isChange, code, name) {
            if (isChange) {
              setState(() {
                screenData.itemunitcode = code;
                screenData.itemunitnames = name;
              });
            }
          },
        ),
      ),
    );

    String subLenght = (screenData.refbarcodes!.isNotEmpty)
        ? "(${screenData.refbarcodes!.length})"
        : "";
    List<Widget> barcodes = [];
    if (screenData.isusesubbarcodes!) {
      List<TextEditingController> textStandController = [];
      List<TextEditingController> textDivideController = [];
      List<FocusNode> barcodesFocusNode = [];
      List<FocusNode> qtySetFocusNode = [];
      List<TextEditingController> qtyController = [];
      List<ValueNotifier<String>> qtyNotifiers = [];

      for (int i = 0; i < screenData.refbarcodes!.length; i++) {
        var data = screenData.refbarcodes![i];

        barcodesFocusNode.add(FocusNode());
        qtySetFocusNode.add(FocusNode());
        textStandController.add(
          TextEditingController(text: data.standvalue.toString()),
        );
        textDivideController.add(
          TextEditingController(text: data.dividevalue.toString()),
        );
        qtyController.add(TextEditingController(text: data.qty.toString()));
        qtyNotifiers.add(ValueNotifier<String>(data.qty.toString()));
      }

      for (int i = 0; i < screenData.refbarcodes!.length; i++) {
        var data = screenData.refbarcodes![i];
        if (isEditMode && currentBarcodeNode > -1) {
          focusAndCursorToEnd(barcodesFocusNode[currentBarcodeNode]);
          focusAndCursorToEnd(qtySetFocusNode[currentBarcodeNode]);
        }

        /// ไม่ใช่สินค้าชุด
        if (screenData.itemtype != 2) {
          barcodes.add(
            Container(
              margin: const EdgeInsets.only(bottom: 10),
              width: double.infinity,
              height: 40,
              child: ElevatedButton(
                style: ButtonStyle(
                  backgroundColor: WidgetStateProperty.all<Color>(
                    global.theme.positiveHighlightTextColor,
                  ),
                  foregroundColor: WidgetStateProperty.all<Color>(global.theme.onPrimaryColor),
                ),
                onPressed: () {
                  subbarcodeSearch().then((result) {
                    if (!result.isusesubbarcodes!) {
                      if (result.guidfixed.isNotEmpty) {
                        data.guidfixed = result.guidfixed;
                        data.barcode = result.barcode!;
                        data.names = result.names!;
                        data.itemunitcode = result.itemunitcode;
                        data.itemunitnames = result.itemunitnames!;
                        currentBarcodeNode = i;
                        data.qty = 0;
                        data.condition = false;
                      }
                    } else {
                      data.barcode = "";
                      global.showSnackBar(
                        context,
                        Icon(Icons.warning_amber, color: global.theme.onPrimaryColor),
                        global.language("ref_barcode_is_sub_barcode"),
                        global.theme.warningHighlightTextColor,
                      );
                    }
                    setState(() {});
                  });
                },
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    Text(
                      (data.barcode.isEmpty)
                          ? "${global.language("barcode")} / ${global.language("item_name")} / ${global.language("unit_name")}"
                          : "${data.barcode} / ${global.activeLangName(data.names)} / ${global.activeLangName(data.itemunitnames)}",
                    ),
                    const Icon(Icons.search),
                  ],
                ),
              ),
            ),
          );
          if (data.barcode.isNotEmpty) {
            barcodes.add(
              Container(
                margin: const EdgeInsets.only(bottom: 10, top: 10),
                child: Column(
                  children: [
                    Row(
                      children: [
                        Expanded(
                          child: TextField(
                            readOnly: !isEditMode,
                            keyboardType: const TextInputType.numberWithOptions(
                              decimal: true,
                            ),
                            focusNode: qtySetFocusNode[i],
                            inputFormatters: <TextInputFormatter>[
                              FilteringTextInputFormatter.allow(
                                RegExp(r'^\d*\.?\d*'),
                              ),
                            ],
                            onEditingComplete: () {},
                            onChanged: (code) {
                              if (code.isNotEmpty && code != '.') {
                                final parsedValue = double.tryParse(code);
                                if (parsedValue != null) {
                                  data.qty = parsedValue;
                                  currentBarcodeNode = i;
                                  // อัปเดตเฉพาะ ValueNotifier แทน setState
                                  qtyNotifiers[i].value = code;
                                }
                              } else if (code.isEmpty) {
                                qtyNotifiers[i].value = '0';
                              }
                            },
                            textAlign: TextAlign.center,
                            controller: qtyController[i],
                            decoration: InputDecoration(
                              floatingLabelBehavior:
                                  FloatingLabelBehavior.always,
                              border: OutlineInputBorder(),
                              labelText: global.language("qty"),
                            ),
                          ),
                        ),
                        Expanded(
                          child: Row(
                            children: [
                              Expanded(
                                child: Row(
                                  children: [
                                    Radio(
                                      value: false,
                                      groupValue: data.condition,
                                      onChanged: (value) {
                                        if (isEditMode) {
                                          setState(() {
                                            data.condition = false;
                                            data.dividevalue = 1;
                                          });
                                        }
                                      },
                                    ),
                                    Expanded(
                                      child: Text(
                                        global.activeLangName(
                                          data.itemunitnames,
                                        ),
                                        overflow: TextOverflow.clip,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                              Expanded(
                                child: Row(
                                  children: [
                                    Radio(
                                      value: true,
                                      groupValue: data.condition,
                                      onChanged: (value) {
                                        if (isEditMode) {
                                          setState(() {
                                            data.standvalue = 1;
                                            data.condition = true;
                                          });
                                        }
                                      },
                                    ),
                                    Expanded(
                                      child: Text(
                                        global.activeLangName(
                                          screenData.itemunitnames!,
                                        ),
                                        overflow: TextOverflow.clip,
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ],
                          ),
                        ),
                        ValueListenableBuilder<String>(
                          valueListenable: qtyNotifiers[i],
                          builder: (context, qtyValue, child) {
                            final displayValue =
                                qtyValue.isEmpty || qtyValue == '.'
                                ? '0'
                                : qtyValue;
                            return Container(
                              alignment: Alignment.centerRight,
                              child: Text(
                                (!data.condition)
                                    ? "${global.formatNumber(double.tryParse(displayValue) ?? 0)} ${global.activeLangName(data.itemunitnames)} = 1 ${global.activeLangName(screenData.itemunitnames!)}"
                                    : "1 ${global.activeLangName(data.itemunitnames)} = ${global.formatNumber(double.tryParse(displayValue) ?? 0)} ${global.activeLangName(screenData.itemunitnames!)}",
                                style: const TextStyle(
                                  fontSize: 24.0,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            );
                          },
                        ),
                      ],
                    ),
                  ],
                ),
              ),
            );
          }

          /// สินค้าชุด
        } else {
          barcodes.add(
            Container(
              margin: const EdgeInsets.only(bottom: 10),
              width: double.infinity,
              height: 40,
              child: Row(
                children: [
                  Expanded(
                    child: SizedBox(
                      height: 40,
                      child: ElevatedButton(
                        style: ButtonStyle(
                          backgroundColor: WidgetStateProperty.all<Color>(
                            global.theme.positiveHighlightTextColor,
                          ),
                          foregroundColor: WidgetStateProperty.all<Color>(
                            global.theme.onPrimaryColor,
                          ),
                        ),
                        onPressed: () {
                          subbarcodeSearch().then((result) {
                            if (result.guidfixed.isNotEmpty) {
                              if (result.refbarcodes!.length == 1) {
                                data.dividevalue =
                                    result.refbarcodes![0].dividevalue;
                                data.standvalue =
                                    result.refbarcodes![0].standvalue;
                              } else {
                                data.dividevalue = 1;
                                data.standvalue = 1;
                              }
                              data.guidfixed = result.guidfixed;
                              data.barcode = result.barcode!;
                              data.names = result.names!;
                              data.itemunitcode = result.itemunitcode;
                              data.itemunitnames = result.itemunitnames!;
                              currentBarcodeNode = i;
                              data.qty = 0;
                              data.condition = false;
                            }
                            setState(() {});
                          });
                        },
                        child: Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Expanded(
                              child: Text(
                                (data.barcode.isEmpty)
                                    ? "${global.language("barcode")} / ${global.language("item_name")} / ${global.language("unit_name")}"
                                    : "${data.barcode} / ${global.activeLangName(data.names)} / ${global.activeLangName(data.itemunitnames)}",
                                style: const TextStyle(
                                  overflow: TextOverflow.ellipsis,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                            ),
                            const Icon(Icons.search),
                          ],
                        ),
                      ),
                    ),
                  ),
                  IconButton(
                    iconSize: 30.0,
                    onPressed: () {
                      setState(() {
                        currentBarcodeNode = 0;
                        if (screenData.refbarcodes!.length > 1) {
                          screenData.refbarcodes!.removeAt(i);
                        } else {
                          data.barcode = "";
                        }
                      });
                    },
                    icon: Icon(Icons.delete, color: global.theme.negativeHighlightTextColor),
                  ),
                ],
              ),
            ),
          );

          if (data.barcode.isNotEmpty) {
            barcodes.add(
              Container(
                margin: const EdgeInsets.only(bottom: 10, top: 10),
                child: Row(
                  children: [
                    Expanded(
                      child: TextField(
                        readOnly: !isEditMode,
                        keyboardType: const TextInputType.numberWithOptions(
                          decimal: true,
                        ),
                        focusNode: qtySetFocusNode[i],
                        inputFormatters: <TextInputFormatter>[
                          FilteringTextInputFormatter.allow(
                            RegExp(r'^\d*\.?\d*'),
                          ),
                        ],
                        onEditingComplete: () {},
                        onChanged: (code) {
                          if (code.isNotEmpty && code != '.') {
                            final parsedValue = double.tryParse(code);
                            if (parsedValue != null) {
                              data.qty = parsedValue;
                              currentBarcodeNode = i;
                              // อัปเดตเฉพาะ ValueNotifier แทน setState
                              qtyNotifiers[i].value = code;
                            }
                          } else if (code.isEmpty) {
                            qtyNotifiers[i].value = '0';
                          }
                        },
                        textAlign: TextAlign.center,
                        controller: qtyController[i],
                        decoration: InputDecoration(
                          floatingLabelBehavior: FloatingLabelBehavior.always,
                          border: OutlineInputBorder(),
                          labelText: global.language("qty"),
                        ),
                      ),
                    ),
                    ValueListenableBuilder<String>(
                      valueListenable: qtyNotifiers[i],
                      builder: (context, qtyValue, child) {
                        final displayValue = qtyValue.isEmpty || qtyValue == '.'
                            ? '0'
                            : qtyValue;
                        return Expanded(
                          child: Container(
                            alignment: Alignment.centerRight,
                            child: Text(
                              "${global.formatNumber(double.tryParse(displayValue) ?? 0)} ${global.activeLangName(data.itemunitnames)} ",
                              style: const TextStyle(
                                fontSize: 24.0,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ),
                        );
                      },
                    ),
                  ],
                ),
              ),
            );
          }
        }
      }
    }

    ///  ประเภทสินค้า
    formWidgets.add(
      Container(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
        child: InputDecorator(
          decoration: InputDecoration(
            labelText: global.language("product_type"),
            enabledBorder: OutlineInputBorder(
              borderRadius: BorderRadius.circular(4.0),
              borderSide: BorderSide(color: global.theme.dividerBorderColor, width: 0.0),
            ),
          ),
          child: Column(
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [Expanded(child: productTypeRadio)],
              ),
              const SizedBox(height: 10),

              ///  ใช้บาร์โค้ดอ้างอิง
              Row(
                children: [
                  Text(
                    "${global.language("use_reference_barcode")} $subLenght",
                    style: TextStyle(
                      fontSize: 16,
                      color:
                          (screenData.itemtype == 1 ||
                              screenData.itemtype == 3 ||
                              screenData.itemtype == 4 ||
                              screenData.itemtype == 5)
                          ? global.theme.textSecondaryColor
                          : global.theme.textColor,
                    ),
                  ),
                  Switch(
                    value: screenData.isusesubbarcodes!,
                    onChanged:
                        (screenData.itemtype == 1 ||
                            screenData.itemtype == 3 ||
                            screenData.itemtype == 4 ||
                            screenData.itemtype == 5)
                        ? null
                        : (bool value) {
                            setState(() {
                              if (value && screenData.refbarcodes!.isEmpty) {
                                screenData.refbarcodes!.add(
                                  BarCodeSubModel(
                                    guidfixed: '',
                                    barcode: '',
                                    names: [],
                                    itemunitcode: '',
                                    itemunitnames: [],
                                    condition: false,
                                    standvalue: 1,
                                    dividevalue: 1,
                                    qty: 0,
                                  ),
                                );
                              } else {
                                screenData.refbarcodes!.clear();
                              }
                              screenData.isusesubbarcodes = value;
                            });
                          },
                  ),
                  Expanded(
                    child:
                        (screenData.itemtype == 2 &&
                            screenData.isusesubbarcodes!)
                        ? SizedBox(
                            width: double.infinity,
                            child: ElevatedButton.icon(
                              icon: const Icon(Icons.add),
                              focusNode: FocusNode(skipTraversal: true),
                              onPressed: () {
                                setState(() {
                                  screenData.refbarcodes!.add(
                                    BarCodeSubModel(
                                      guidfixed: '',
                                      barcode: '',
                                      names: [],
                                      itemunitcode: '',
                                      itemunitnames: [],
                                      condition: false,
                                      standvalue: 1,
                                      dividevalue: 1,
                                      qty: 0,
                                    ),
                                  );
                                });
                              },
                              label: Text(global.language("add_barcode")),
                            ),
                          )
                        : Container(),
                  ),
                ],
              ),
              (screenData.isusesubbarcodes!)
                  ? Padding(
                      padding: const EdgeInsets.only(left: 10, right: 10),
                      child: Center(
                        child: Card(
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(4),
                            side: BorderSide(
                              color: global.theme.infoHighlightTextColor.withValues(alpha: 0.3),
                              width: 2,
                            ), // Set the border color and width
                          ),
                          child: Padding(
                            padding: const EdgeInsets.all(12),
                            child: Column(children: barcodes),
                          ),
                        ),
                      ),
                    )
                  : Container(),

              ///  บาร์โค้ดทดแทน
            ],
          ),
        ),
      ),
    );

    ///  ประเภทวัตถุดิบ
    if (global.posVersion == global.PosVersionEnum.restaurant) {
      formWidgets.add(
        Container(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
          child: InputDecorator(
            decoration: InputDecoration(
              labelText: global.language("material_type"),
              enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(4.0),
                borderSide: BorderSide(color: global.theme.dividerBorderColor, width: 0.0),
              ),
            ),
            child: Row(
              mainAxisAlignment: MainAxisAlignment.spaceBetween,
              children: [Expanded(child: materialTypeRadio)],
            ),
          ),
        ),
      );
    }

    /// bom สูตรผลิต
    String bomLenght = (screenData.bom!.isNotEmpty)
        ? "(${screenData.bom!.length})"
        : "";
    List<Widget> bomBarcodes = [];
    if (screenData.bom!.isNotEmpty) {
      List<FocusNode> bomBarcodesFocusNode = [];
      List<FocusNode> bomQtySetFocusNode = [];
      List<TextEditingController> bomQtyController = [];

      for (int i = 0; i < screenData.bom!.length; i++) {
        var data = screenData.bom![i];

        bomBarcodesFocusNode.add(FocusNode());
        bomQtySetFocusNode.add(FocusNode());
        bomQtyController.add(TextEditingController(text: data.qty.toString()));
      }

      for (int i = 0; i < screenData.bom!.length; i++) {
        var data = screenData.bom![i];
        if (isEditMode && bomCurrentBarcodeNode > -1) {
          focusAndCursorToEnd(bomBarcodesFocusNode[bomCurrentBarcodeNode]);
          focusAndCursorToEnd(bomQtySetFocusNode[bomCurrentBarcodeNode]);
        }

        bomBarcodes.add(
          Container(
            margin: const EdgeInsets.only(bottom: 10),
            width: double.infinity,
            height: 40,
            child: Row(
              children: [
                Expanded(
                  child: SizedBox(
                    height: 40,
                    child: ElevatedButton(
                      style: ButtonStyle(
                        backgroundColor: WidgetStateProperty.all<Color>(
                          global.theme.warningHighlightTextColor,
                        ),
                        foregroundColor: WidgetStateProperty.all<Color>(
                          global.theme.onPrimaryColor,
                        ),
                      ),
                      onPressed: () {
                        subbarcodeSearch().then((result) {
                          if (result.guidfixed.isNotEmpty) {
                            if (result.refbarcodes!.length == 1) {
                              data.dividevalue =
                                  result.refbarcodes![0].dividevalue;
                              data.standvalue =
                                  result.refbarcodes![0].standvalue;
                            } else {
                              data.dividevalue = 1;
                              data.standvalue = 1;
                            }
                            data.guidfixed = result.guidfixed;
                            data.barcode = result.barcode!;
                            data.names = result.names!;
                            data.itemunitcode = result.itemunitcode;
                            data.itemunitnames = result.itemunitnames!;
                            bomCurrentBarcodeNode = i;
                            data.qty = 0;
                            data.condition = false;
                          }
                          setState(() {});
                        });
                      },
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Expanded(
                            child: Text(
                              (data.barcode.isEmpty)
                                  ? "${global.language("barcode")} / ${global.language("item_name")} / ${global.language("unit_name")}"
                                  : "${data.barcode} / ${global.activeLangName(data.names)} / ${global.activeLangName(data.itemunitnames)}",
                              style: const TextStyle(
                                overflow: TextOverflow.ellipsis,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                          ),
                          const Icon(Icons.search),
                        ],
                      ),
                    ),
                  ),
                ),
                IconButton(
                  iconSize: 30.0,
                  onPressed: () {
                    setState(() {
                      bomCurrentBarcodeNode = 0;
                      if (screenData.bom!.length > 1) {
                        screenData.bom!.removeAt(i);
                      } else {
                        data.barcode = "";
                      }
                    });
                  },
                  icon: Icon(Icons.delete, color: global.theme.negativeHighlightTextColor),
                ),
              ],
            ),
          ),
        );

        if (data.barcode.isNotEmpty) {
          bomBarcodes.add(
            Container(
              margin: const EdgeInsets.only(bottom: 10, top: 10),
              child: Row(
                children: [
                  Expanded(
                    child: TextField(
                      readOnly: !isEditMode,
                      keyboardType: const TextInputType.numberWithOptions(
                        decimal: true,
                      ), // Allow decimal input
                      focusNode: bomQtySetFocusNode[i],
                      inputFormatters: <TextInputFormatter>[
                        // Allow digits and dot, limit to 6 decimal places
                        FilteringTextInputFormatter.allow(
                          RegExp(r'^\d*\.?\d{0,6}'),
                        ),
                      ],
                      onEditingComplete: () {
                        setState(() {});
                      },
                      onChanged: (value) {
                        if (value.isNotEmpty) {
                          data.qty = double.parse(value);
                          bomCurrentBarcodeNode = i;
                          _debouncer.run(() {
                            setState(() {});
                          });
                        }
                      },
                      textAlign: TextAlign.center,
                      controller: bomQtyController[i],
                      decoration: InputDecoration(
                        floatingLabelBehavior: FloatingLabelBehavior.always,
                        border: OutlineInputBorder(),
                        labelText: global.language("qty"),
                      ),
                    ),
                  ),
                  Expanded(
                    child: Container(
                      alignment: Alignment.centerRight,
                      child: Text(
                        "${bomQtyController[i].text} ${global.activeLangName(data.itemunitnames)} ",
                        style: const TextStyle(
                          fontSize: 24.0,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          );
        }
      }
    }

    ///  bom สูตรผลิต
    if (global.posVersion == global.PosVersionEnum.restaurant) {
      formWidgets.add(
        Container(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
          child: Row(
            children: [
              Text(
                "${global.language("use_product_boom")} $bomLenght",
                style: TextStyle(
                  fontSize: 16,
                  color: (screenData.itemtype == 1)
                      ? global.theme.textSecondaryColor
                      : global.theme.textColor,
                ),
              ),
              Switch(
                value: (screenData.bom!.isNotEmpty) ? true : false,
                onChanged: (bool value) {
                  if (value) {
                    setState(() {
                      if (value && screenData.bom!.isEmpty) {
                        screenData.bom!.add(
                          BomModel(
                            guidfixed: '',
                            barcode: '',
                            names: [],
                            itemunitcode: '',
                            itemunitnames: [],
                            condition: false,
                            standvalue: 1,
                            dividevalue: 1,
                            qty: 0,
                          ),
                        );
                      }
                    });
                  } else {
                    setState(() {
                      screenData.bom!.clear();
                    });
                  }
                },
              ),
              Expanded(
                child: (screenData.bom!.isNotEmpty)
                    ? SizedBox(
                        width: double.infinity,
                        child: ElevatedButton.icon(
                          style: ButtonStyle(
                            backgroundColor: WidgetStateProperty.all<Color?>(
                              global.theme.infoHighlightColor,
                            ),
                            foregroundColor: WidgetStateProperty.all<Color>(
                              global.theme.textColor,
                            ),
                          ),
                          icon: const Icon(Icons.add),
                          focusNode: FocusNode(skipTraversal: true),
                          onPressed: () {
                            setState(() {
                              screenData.bom!.add(
                                BomModel(
                                  guidfixed: '',
                                  barcode: '',
                                  names: [],
                                  itemunitcode: '',
                                  itemunitnames: [],
                                  condition: false,
                                  standvalue: 1,
                                  dividevalue: 1,
                                  qty: 0,
                                ),
                              );
                            });
                          },
                          label: Text(global.language("add_barcode")),
                        ),
                      )
                    : Container(),
              ),
            ],
          ),
        ),
      );
    }

    ///  bom สูตรผลิต
    if (screenData.bom!.isNotEmpty) {
      formWidgets.add(
        Container(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
          child: InputDecorator(
            decoration: InputDecoration(
              labelText: global.language("product_bom"),
              enabledBorder: OutlineInputBorder(
                borderRadius: BorderRadius.circular(4.0),
                borderSide: BorderSide(color: global.theme.dividerBorderColor, width: 0.0),
              ),
            ),
            child: Padding(
              padding: const EdgeInsets.only(left: 10, right: 10),
              child: Center(
                child: Card(
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(4),
                    side: BorderSide(
                      color: global.theme.infoHighlightTextColor.withValues(alpha: 0.3),
                      width: 2,
                    ), // Set the border color and width
                  ),
                  child: Padding(
                    padding: const EdgeInsets.all(12),
                    child: Column(children: bomBarcodes),
                  ),
                ),
              ),
            ),
          ),
        ),
      );
    }

    if (global.posVersion == global.PosVersionEnum.restaurant) {
      ///  ประเภทอาหาร
      formWidgets.add(
        Container(
          padding: const EdgeInsets.only(left: 10, right: 10, bottom: 10),
          child: Row(
            mainAxisAlignment: MainAxisAlignment.spaceBetween,
            children: [Expanded(child: foodTypeRadio)],
          ),
        ),
      );
    }

    ///  ประเภทภาษี
    formWidgets.add(
      Container(
        margin: const EdgeInsets.only(bottom: 10),
        padding: const EdgeInsets.only(left: 10, right: 10),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.spaceBetween,
          children: [Expanded(child: vatTypeRadio)],
        ),
      ),
    );

    ///  คะแนนสะสม
    formWidgets.add(
      Container(
        margin: const EdgeInsets.only(bottom: 10),
        padding: const EdgeInsets.only(left: 10, right: 10),
        child: issumPointRadio,
      ),
    );

    // รหัสสินค้า
    focusNodeMax++;
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
        child: TextField(
          readOnly: !isEditMode,
          onSubmitted: (value) {
            if (kIsWeb) {
              findFocusNext(focusNodeIndex);
            }
          },
          focusNode: fieldFocusNodes[focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: TextEditingController(text: screenData.itemcode),
          textCapitalization: TextCapitalization.characters,
          inputFormatters: <TextInputFormatter>[
            FilteringTextInputFormatter.allow(RegExp('[a-z A-Z 0-9 -]')),
          ],
          onChanged: (value) {
            isDataChange = true;
            screenData.itemcode = value.toUpperCase();
            TextEditingController(text: screenData.itemcode).selection =
                TextSelection.fromPosition(TextPosition(offset: value.length));
          },
          decoration: InputDecoration(
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: OutlineInputBorder(),
            labelText: global.language("item_code"),
          ),
        ),
      ),
    );

    // ///  กลุ่มสินค้า
    // formWidgets.add(Container(
    //   margin: const EdgeInsets.only(bottom: 10),
    //   padding: const EdgeInsets.only(left: 10, right: 10),
    //   child: Row(
    //     children: [
    //       Expanded(
    //           child: RawKeyboardListener(
    //               focusNode: FocusNode(),
    //               onKey: (RawKeyEvent event) {
    //                 if (event is RawKeyDownEvent) {
    //                   if (event.logicalKey == LogicalKeyboardKey.f2) {
    //                     productGroupSearch();
    //                   }
    //                 }
    //               },
    //               child: TextField(
    //                   readOnly: false,
    //                   textInputAction: TextInputAction.next,
    //                   focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
    //                   controller: TextEditingController(
    //                       text: screenData.groupcode ?? ""),
    //                   textAlign: TextAlign.left,
    //                   textCapitalization: TextCapitalization.characters,
    //                   onChanged: (code) {
    //                     isDataChange = true;
    //                     screenData.groupcode = code;
    //                     findProductGroup();
    //                   },
    //                   decoration: InputDecoration(
    //                     contentPadding: EdgeInsets.all(10.0),
    //                     enabledBorder: OutlineInputBorder(
    //                       borderSide:
    //                           BorderSide(color: global.theme.dividerBorderColor, width: 0.0),
    //                     ),
    //                     floatingLabelBehavior: FloatingLabelBehavior.always,
    //                     suffixIcon: Row(
    //                       mainAxisAlignment:
    //                           MainAxisAlignment.spaceBetween, // added line
    //                       mainAxisSize: MainAxisSize.min,
    //                       children: [
    //                         (isEditMode)
    //                             ? IconButton(
    //                                 focusNode: FocusNode(skipTraversal: true),
    //                                 icon: const Icon(Icons.search),
    //                                 onPressed: () {
    //                                   productGroupSearch();
    //                                 },
    //                               )
    //                             : Container(),
    //                       ],
    //                     ),
    // border: OutlineInputBorder(),
    //                     labelText: global.language("product_group_code"),
    //                   )))),
    // SizedBox(width: 5),
    //       Expanded(
    //           child: TextField(
    //               readOnly: true,
    //               controller: TextEditingController(
    //                   text: global.activeLangName(screenData.groupnames!)),
    //               textAlign: TextAlign.left,
    //               textCapitalization: TextCapitalization.characters,
    //               decoration: InputDecoration(
    //                 contentPadding: EdgeInsets.all(10.0),
    //                 enabledBorder: OutlineInputBorder(
    //                   borderSide: BorderSide(color: global.theme.dividerBorderColor, width: 0.0),
    //                 ),
    //                 floatingLabelBehavior: FloatingLabelBehavior.always,
    // border: OutlineInputBorder(),
    //                 labelText: global.language("product_group_name"),
    //               )))
    //     ],
    //   ),
    // ));

    /// ประเภทสินค้า
    formWidgets.add(
      Padding(
        padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Row(
          children: [
            Expanded(
              child: SizedBox(
                height: 50,
                child: ElevatedButton(
                  style: ButtonStyle(
                    backgroundColor: WidgetStateProperty.all<Color>(
                      (screenData.producttype!.guidfixed!.isNotEmpty)
                          ? global.theme.primaryColor.withValues(alpha: 0.08)
                          : global.theme.cardColor,
                    ),
                    foregroundColor: WidgetStateProperty.all<Color>(
                      global.theme.textColor,
                    ),
                    elevation: WidgetStateProperty.all(0),
                    shape: WidgetStateProperty.all(
                      RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(8),
                        side: BorderSide(
                          color: (screenData.producttype!.guidfixed!.isNotEmpty)
                              ? global.theme.primaryColor
                              : global.theme.dividerBorderColor,
                        ),
                      ),
                    ),
                  ),
                  onPressed: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) =>
                            const ProductTypeSearchScreen(word: ''),
                      ),
                    ).then((value) {
                      if (value != null) {
                        setState(() {
                          ProductTypeModel result = value;
                          if (result.guidfixed!.isNotEmpty) {
                            screenData.producttype!.guidfixed =
                                result.guidfixed;
                            screenData.producttype!.code = result.code;
                            screenData.producttype!.names = result.names;
                          }
                        });
                      }
                    });
                    setState(() {});
                  },
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Icon(Icons.category, color: global.theme.primaryColor),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          (screenData.producttype!.guidfixed!.isEmpty)
                              ? "${global.language("product_type_code")} ~ ${global.language("product_type_name")} "
                              : "${global.language("product_type_name")} : ${screenData.producttype!.code} ~ ${global.activeLangName(screenData.producttype!.names!)}  ",
                          overflow: TextOverflow.ellipsis,
                          maxLines: 2,
                        ),
                      ),
                      (screenData.producttype!.guidfixed!.isNotEmpty)
                          ? IconButton(
                              onPressed: () {
                                setState(() {
                                  screenData.producttype!.guidfixed = '';
                                  screenData.producttype!.code = '';
                                  screenData.producttype!.names = [];
                                });
                              },
                              icon: Icon(Icons.clear, color: global.theme.negativeHighlightTextColor),
                            )
                          : Container(),
                      Icon(Icons.search, color: global.theme.textSecondaryColor),
                    ],
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );

    /// Master Data Selections
    // Brand Selection and Category Selection
    formWidgets.add(
      Row(
        children: [
          Expanded(
            child: buildMasterDataSelectionWidget(
              label: global.language("brand"),
              code: screenData.brandcode ?? "",
              displayName: global.activeLangName(screenData.brandnames ?? []),
              isSelected: (screenData.brandguid ?? "").isNotEmpty,
              onTap: () {
                GroupSelectionHelper.showMasterDataSelectionScreen<
                  MasterBrandModel
                >(
                  context: context,
                  dataType: MasterDataType.brand,
                  onItemSelected: (selectedItem) {
                    setState(() {
                      screenData.brandcode = selectedItem.code;
                      screenData.brandguid = selectedItem.guidfixed;
                      screenData.brandnames = selectedItem.names;
                      isDataChange = true;
                    });
                  },
                );
              },
              onClear: () {
                setState(() {
                  screenData.brandcode = "";
                  screenData.brandguid = "";
                  screenData.brandnames = [];
                  isDataChange = true;
                });
              },
            ),
          ),
          Expanded(
            child: buildMasterDataSelectionWidget(
              label: global.language("category"),
              code: screenData.categorycode ?? "",
              displayName: global.activeLangName(
                screenData.categorynames ?? [],
              ),
              isSelected: (screenData.categoryguid ?? "").isNotEmpty,
              onTap: () {
                GroupSelectionHelper.showMasterDataSelectionScreen<
                  MasterCategoryModel
                >(
                  context: context,
                  dataType: MasterDataType.category,
                  onItemSelected: (selectedItem) {
                    setState(() {
                      screenData.categorycode = selectedItem.code;
                      screenData.categoryguid = selectedItem.guidfixed;
                      screenData.categorynames = selectedItem.names;
                      isDataChange = true;
                    });
                  },
                );
              },
              onClear: () {
                setState(() {
                  screenData.categorycode = "";
                  screenData.categoryguid = "";
                  screenData.categorynames = [];
                  isDataChange = true;
                });
              },
            ),
          ),
        ],
      ),
    );

    // Class Selection and Design Selection
    formWidgets.add(
      Row(
        children: [
          Expanded(
            child: buildMasterDataSelectionWidget(
              label: global.language("class"),
              code: screenData.classcode ?? "",
              displayName: global.activeLangName(screenData.classnames ?? []),
              isSelected: (screenData.classguid ?? "").isNotEmpty,
              onTap: () {
                GroupSelectionHelper.showMasterDataSelectionScreen<
                  MasterClassModel
                >(
                  context: context,
                  dataType: MasterDataType.masterClass,
                  onItemSelected: (selectedItem) {
                    setState(() {
                      screenData.classcode = selectedItem.code;
                      screenData.classguid = selectedItem.guidfixed;
                      screenData.classnames = selectedItem.names;
                      isDataChange = true;
                    });
                  },
                );
              },
              onClear: () {
                setState(() {
                  screenData.classcode = "";
                  screenData.classguid = "";
                  screenData.classnames = [];
                  isDataChange = true;
                });
              },
            ),
          ),
          Expanded(
            child: buildMasterDataSelectionWidget(
              label: global.language("design"),
              code: screenData.designcode ?? "",
              displayName: global.activeLangName(screenData.designnames ?? []),
              isSelected: (screenData.designguid ?? "").isNotEmpty,
              onTap: () {
                GroupSelectionHelper.showMasterDataSelectionScreen<
                  MasterDesignModel
                >(
                  context: context,
                  dataType: MasterDataType.design,
                  onItemSelected: (selectedItem) {
                    setState(() {
                      screenData.designcode = selectedItem.code;
                      screenData.designguid = selectedItem.guidfixed;
                      screenData.designnames = selectedItem.names;
                      isDataChange = true;
                    });
                  },
                );
              },
              onClear: () {
                setState(() {
                  screenData.designcode = "";
                  screenData.designguid = "";
                  screenData.designnames = [];
                  isDataChange = true;
                });
              },
            ),
          ),
        ],
      ),
    );

    // Grade Selection and Model Selection
    formWidgets.add(
      Row(
        children: [
          Expanded(
            child: buildMasterDataSelectionWidget(
              label: global.language("grade"),
              code: screenData.gradecode ?? "",
              displayName: global.activeLangName(screenData.gradenames ?? []),
              isSelected: (screenData.gradeguid ?? "").isNotEmpty,
              onTap: () {
                GroupSelectionHelper.showMasterDataSelectionScreen<
                  MasterGradeModel
                >(
                  context: context,
                  dataType: MasterDataType.grade,
                  onItemSelected: (selectedItem) {
                    setState(() {
                      screenData.gradecode = selectedItem.code;
                      screenData.gradeguid = selectedItem.guidfixed;
                      screenData.gradenames = selectedItem.names;
                      isDataChange = true;
                    });
                  },
                );
              },
              onClear: () {
                setState(() {
                  screenData.gradecode = "";
                  screenData.gradeguid = "";
                  screenData.gradenames = [];
                  isDataChange = true;
                });
              },
            ),
          ),
          Expanded(
            child: buildMasterDataSelectionWidget(
              label: global.language("model"),
              code: screenData.modelcode ?? "",
              displayName: global.activeLangName(screenData.modelnames ?? []),
              isSelected: (screenData.modelguid ?? "").isNotEmpty,
              onTap: () {
                GroupSelectionHelper.showMasterDataSelectionScreen<
                  MasterModelModel
                >(
                  context: context,
                  dataType: MasterDataType.model,
                  onItemSelected: (selectedItem) {
                    setState(() {
                      screenData.modelcode = selectedItem.code;
                      screenData.modelguid = selectedItem.guidfixed;
                      screenData.modelnames = selectedItem.names;
                      isDataChange = true;
                    });
                  },
                );
              },
              onClear: () {
                setState(() {
                  screenData.modelcode = "";
                  screenData.modelguid = "";
                  screenData.modelnames = [];
                  isDataChange = true;
                });
              },
            ),
          ),
        ],
      ),
    );

    // Group selection and Subgroup selections
    formWidgets.add(
      Row(
        children: [
          Expanded(
            child: buildMasterDataSelectionWidget(
              label: global.language("group_main"),
              code: screenData.groupcode ?? "",
              displayName: global.activeLangName(screenData.groupnames ?? []),
              isSelected: (screenData.groupguid ?? "").isNotEmpty,
              onTap: () {
                GroupSelectionHelper.showMasterDataSelectionScreen<
                  MasterGroupModel
                >(
                  context: context,
                  dataType: MasterDataType.group,
                  onItemSelected: (selectedItem) {
                    setState(() {
                      screenData.groupcode = selectedItem.code;
                      screenData.groupguid = selectedItem.guidfixed;
                      screenData.groupnames = selectedItem.names;
                      isDataChange = true;
                    });
                  },
                );
              },
              onClear: () {
                setState(() {
                  screenData.groupcode = "";
                  screenData.groupguid = "";
                  screenData.groupnames = [];
                  isDataChange = true;
                });
              },
            ),
          ),
          Expanded(
            child: buildMasterDataSelectionWidget(
              label: global.language("group_sub1"),
              code: screenData.groupsubonecode ?? "",
              displayName: global.activeLangName(
                screenData.groupsubonenames ?? [],
              ),
              isSelected: (screenData.groupsuboneguid ?? "").isNotEmpty,
              onTap: () {
                GroupSelectionHelper.showMasterDataSelectionScreen<
                  MasterGroupSub1Model
                >(
                  context: context,
                  dataType: MasterDataType.groupSub1,
                  onItemSelected: (selectedItem) {
                    setState(() {
                      screenData.groupsubonecode = selectedItem.code;
                      screenData.groupsuboneguid = selectedItem.guidfixed;
                      screenData.groupsubonenames = selectedItem.names;
                      isDataChange = true;
                    });
                  },
                );
              },
              onClear: () {
                setState(() {
                  screenData.groupsubonecode = "";
                  screenData.groupsuboneguid = "";
                  screenData.groupsubonenames = [];
                  isDataChange = true;
                });
              },
            ),
          ),
        ],
      ),
    );

    // Group Sub2 Selection and Pattern Selection
    formWidgets.add(
      Row(
        children: [
          Expanded(
            child: buildMasterDataSelectionWidget(
              label: global.language("group_sub2"),
              code: screenData.groupsubtwocode ?? "",
              displayName: global.activeLangName(
                screenData.groupsubtwonames ?? [],
              ),
              isSelected: (screenData.groupsubtwoguid ?? "").isNotEmpty,
              onTap: () {
                GroupSelectionHelper.showMasterDataSelectionScreen<
                  MasterGroupSub2Model
                >(
                  context: context,
                  dataType: MasterDataType.groupSub2,
                  onItemSelected: (selectedItem) {
                    setState(() {
                      screenData.groupsubtwocode = selectedItem.code;
                      screenData.groupsubtwoguid = selectedItem.guidfixed;
                      screenData.groupsubtwonames = selectedItem.names;
                      isDataChange = true;
                    });
                  },
                );
              },
              onClear: () {
                setState(() {
                  screenData.groupsubtwocode = "";
                  screenData.groupsubtwoguid = "";
                  screenData.groupsubtwonames = [];
                  isDataChange = true;
                });
              },
            ),
          ),
          Expanded(
            child: buildMasterDataSelectionWidget(
              label: global.language("pattern"),
              code: screenData.patterncode ?? "",
              displayName: global.activeLangName(screenData.patternnames ?? []),
              isSelected: (screenData.patternguid ?? "").isNotEmpty,
              onTap: () {
                GroupSelectionHelper.showMasterDataSelectionScreen<
                  MasterPatternModel
                >(
                  context: context,
                  dataType: MasterDataType.pattern,
                  onItemSelected: (selectedItem) {
                    setState(() {
                      screenData.patterncode = selectedItem.code;
                      screenData.patternguid = selectedItem.guidfixed;
                      screenData.patternnames = selectedItem.names;
                      isDataChange = true;
                    });
                  },
                );
              },
              onClear: () {
                setState(() {
                  screenData.patterncode = "";
                  screenData.patternguid = "";
                  screenData.patternnames = [];
                  isDataChange = true;
                });
              },
            ),
          ),
        ],
      ),
    );

    /// ผู้ผลิต
    formWidgets.add(
      Padding(
        padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Row(
          children: [
            Expanded(
              child: SizedBox(
                height: 50,
                child: ElevatedButton(
                  style: ButtonStyle(
                    backgroundColor: WidgetStateProperty.all<Color>(
                      (screenData.manufacturerguid!.isNotEmpty)
                          ? global.theme.primaryColor.withValues(alpha: 0.08)
                          : global.theme.cardColor,
                    ),
                    foregroundColor: WidgetStateProperty.all<Color>(
                      global.theme.textColor,
                    ),
                    elevation: WidgetStateProperty.all(0),
                    shape: WidgetStateProperty.all(
                      RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(8),
                        side: BorderSide(
                          color: (screenData.manufacturerguid!.isNotEmpty)
                              ? global.theme.primaryColor
                              : global.theme.dividerBorderColor,
                        ),
                      ),
                    ),
                  ),
                  onPressed: () {
                    Navigator.push(
                      context,
                      MaterialPageRoute(
                        builder: (context) =>
                            const SupplierSearchScreen(word: ''),
                      ),
                    ).then((value) {
                      if (value != null) {
                        setState(() {
                          SearchGuidCodeNameModel result = value;
                          if (result.guid.isNotEmpty) {
                            screenData.manufacturerguid = result.guid;
                            screenData.manufacturercode = result.code;
                            screenData.manufacturernames = result.names;
                          }
                        });
                      }
                    });
                    setState(() {});
                  },
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.spaceBetween,
                    children: [
                      Icon(Icons.person_search_rounded, color: global.theme.primaryColor),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          (screenData.manufacturerguid!.isEmpty)
                              ? "${global.language("manufacturer_code")} ~ ${global.language("manufacturer_name")} "
                              : "${global.language("manufacturer_name")} : ${screenData.manufacturercode} ~ ${global.activeLangName(screenData.manufacturernames!)}  ",
                          overflow: TextOverflow.ellipsis,
                          maxLines: 2,
                        ),
                      ),
                      (screenData.manufacturerguid!.isNotEmpty)
                          ? IconButton(
                              onPressed: () {
                                setState(() {
                                  screenData.manufacturerguid = '';
                                  screenData.manufacturercode = '';
                                  screenData.manufacturernames = [];
                                });
                              },
                              icon: Icon(Icons.clear, color: global.theme.negativeHighlightTextColor),
                            )
                          : Container(),
                      Icon(Icons.search, color: global.theme.textSecondaryColor),
                    ],
                  ),
                ),
              ),
            ),
          ],
        ),
      ),
    );

    /// มิติสินค้า
    formWidgets.add(
      Container(
        height: 50,
        padding: const EdgeInsets.only(left: 10, right: 10, bottom: 10),
        width: double.infinity,
        child: ElevatedButton.icon(
          onPressed: () {
            setState(() {
              screenData.dimensions!.add(
                DimensionProductModel(
                  guidfixed: '',
                  names: names,
                  item: ItemDimension(
                    guidfixed: '',
                    names: [],
                    isdisabled: false,
                  ),
                ),
              );
              dimensionSeleted.add(DimensionModel());
            });
          },
          style: ButtonStyle(
            backgroundColor: WidgetStateProperty.all<Color>(
              global.theme.primaryColor,
            ), // Set the background color here
          ),
          icon: Icon(Icons.add), // Set the icon here
          label: Text(global.language("add_dimension")),
        ),
      ),
    );

    /// มิติสินค้า
    if (screenData.dimensions!.isNotEmpty) {
      for (
        int dimensionsIndex = 0;
        dimensionsIndex < screenData.dimensions!.length;
        dimensionsIndex++
      ) {
        formWidgets.add(
          Padding(
            padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Expanded(
                  child: SizedBox(
                    height: 50,
                    child: ElevatedButton(
                      style: ButtonStyle(
                        backgroundColor: WidgetStateProperty.all<Color>(
                          global.theme.primaryColor,
                        ),
                        foregroundColor: WidgetStateProperty.all<Color>(
                          global.theme.onPrimaryColor,
                        ),
                      ),
                      onPressed: () {
                        dimensionSearch(dimensionsIndex);
                        setState(() {});
                      },
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Expanded(
                            child: Text(
                              (screenData
                                      .dimensions![dimensionsIndex]
                                      .guidfixed!
                                      .isEmpty)
                                  ? "${global.language("level_dimension")} : ${dimensionsIndex + 1} "
                                  : "${global.language("level_dimension")} : ${dimensionsIndex + 1} ~ ${global.activeLangName(screenData.dimensions![dimensionsIndex].names!)}  ",
                              overflow: TextOverflow.ellipsis,
                              maxLines: 2,
                            ),
                          ),
                          const Icon(Icons.search),
                        ],
                      ),
                    ),
                  ),
                ),
                const SizedBox(width: 10),
                Expanded(
                  child: DropdownSearch<ItemDimension>(
                    enabled: isEditMode,
                    items: (String filter, infiniteScrollProps) =>
                        getDataItemDimension(filter, dimensionsIndex),
                    compareFn: (item, selectedItem) =>
                        item.guidfixed == selectedItem.guidfixed,
                    itemAsString: (ItemDimension? item) {
                      if (item == null || item.guidfixed!.isEmpty) return '';
                      return global.activeLangName(item.names!);
                    },
                    decoratorProps: const DropDownDecoratorProps(
                      decoration: InputDecoration(
                        border: OutlineInputBorder(),
                        contentPadding: EdgeInsets.only(
                          left: 10,
                          top: 0,
                          bottom: 0,
                          right: 10,
                        ),
                        floatingLabelBehavior: FloatingLabelBehavior.always,
                        filled: true,
                      ),
                    ),
                    onChanged: (ItemDimension? value) {
                      if (value != null && value.isdisabled!) {
                        setState(() {
                          screenData.dimensions![dimensionsIndex].item = value;
                        });
                      } else {
                        setState(() {
                          screenData.dimensions![dimensionsIndex].item =
                              ItemDimension(
                                guidfixed: '',
                                names: [],
                                isdisabled: false,
                              );
                          global.showSnackBar(
                            context,
                            Icon(
                              Icons.warning_amber,
                              color: global.theme.onPrimaryColor,
                            ),
                            global.language("ref_dimension_is_disabled"),
                            global.theme.warningHighlightTextColor,
                          );
                        });
                      }
                    },
                    validator: (ItemDimension? u) =>
                        u!.guidfixed!.isEmpty ? "required field" : null,
                    popupProps: PopupPropsMultiSelection.dialog(
                      showSearchBox: true,
                      showSelectedItems: true,
                      itemBuilder: (ctx, item, isDisabled, isSelected) {
                        return ListTile(
                          enabled: !isDisabled,
                          title: Text(global.activeLangName(item.names!)),
                          selected: isSelected,
                        );
                      },
                    ),
                    selectedItem: screenData.dimensions![dimensionsIndex].item,
                  ),
                ),
                const SizedBox(width: 10),
                IconButton(
                  onPressed: () {
                    setState(() {
                      dimensionSeleted.removeAt(dimensionsIndex);
                      screenData.dimensions!.removeAt(dimensionsIndex);
                    });
                  },
                  icon: const Icon(Icons.delete),
                  focusNode: FocusNode(skipTraversal: true),
                  color: global.theme.negativeHighlightTextColor,
                  iconSize: 20,
                ),
              ],
            ),
          ),
        );
      }
    }

    ///  ราคาขาย (2-column compact layout)
    {
      List<Widget> priceFields = [];
      for (int priceIndex = 0;
          priceIndex < priceList.length;
          priceIndex++) {
        priceFields.add(
          Expanded(
            child: Padding(
              padding: EdgeInsets.only(
                left: priceIndex.isEven ? 10 : 4,
                right: priceIndex.isEven ? 4 : 10,
                bottom: 6,
              ),
              child: TextField(
                readOnly: !isEditMode,
                keyboardType:
                    const TextInputType.numberWithOptions(decimal: true),
                inputFormatters: [global.NumberInputFormatter()],
                style: const TextStyle(fontSize: 13),
                onChanged: (value) {
                  isDataChange = true;
                  if (value != '') {
                    screenData.prices![priceIndex].price = double.parse(
                      value.replaceAll(',', ''),
                    );
                  } else {
                    screenData.prices![priceIndex].price = 0;
                  }
                },
                onSubmitted: (value) {
                  findFocusNext(focusNodeIndex);
                },
                focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
                textAlign: TextAlign.right,
                controller: TextEditingController(
                  text: (screenData.prices![priceIndex].price == 0.0)
                      ? ""
                      : global.formatNumber(
                          screenData.prices!
                              .firstWhere(
                                (element) =>
                                    element.keynumber ==
                                    priceList[priceIndex].keyNumber,
                              )
                              .price,
                        ),
                ),
                decoration: InputDecoration(
                  isDense: true,
                  contentPadding: const EdgeInsets.symmetric(
                      horizontal: 8, vertical: 10),
                  floatingLabelBehavior: FloatingLabelBehavior.always,
                  border: const OutlineInputBorder(),
                  labelText: priceList[priceIndex].names[0].name,
                  suffixIcon: isEditMode
                      ? IconButton(
                          icon: const Icon(Icons.calculate_outlined, size: 18),
                          padding: EdgeInsets.zero,
                          constraints: const BoxConstraints(
                              minWidth: 30, minHeight: 30),
                          onPressed: () async {
                            final result = await showNumPadDialog(
                              context,
                              title: priceList[priceIndex].names[0].name ?? '',
                              initialValue: screenData
                                          .prices![priceIndex].price ==
                                      0.0
                                  ? ''
                                  : screenData.prices![priceIndex].price
                                      .toString(),
                            );
                            if (result != null && result.isNotEmpty) {
                              setState(() {
                                isDataChange = true;
                                screenData.prices![priceIndex].price =
                                    double.parse(result.replaceAll(',', ''));
                              });
                            }
                          },
                        )
                      : null,
                ),
              ),
            ),
          ),
        );
      }
      // Group into rows of 2
      for (int i = 0; i < priceFields.length; i += 2) {
        formWidgets.add(
          Row(
            children: [
              priceFields[i],
              if (i + 1 < priceFields.length)
                priceFields[i + 1]
              else
                const Expanded(child: SizedBox()),
            ],
          ),
        );
      }
    }
    // ส่วนลด
    focusNodeMax++;
    formWidgets.add(
      Padding(
        padding: const EdgeInsets.only(left: 10, right: 10, bottom: 6),
        child: TextField(
          readOnly: !isEditMode,
          style: const TextStyle(fontSize: 13),
          onSubmitted: (value) {
            if (kIsWeb) {
              findFocusNext(++focusNodeIndex);
            }
          },
          focusNode: fieldFocusNodes[focusNodeMax].focusNode,
          textAlign: TextAlign.left,
          controller: TextEditingController(text: screenData.discount),
          textCapitalization: TextCapitalization.characters,
          onChanged: (value) {
            isDataChange = true;
            screenData.discount = value;
            TextEditingController(text: screenData.discount).selection =
                TextSelection.fromPosition(TextPosition(offset: value.length));
          },
          decoration: InputDecoration(
            isDense: true,
            contentPadding:
                const EdgeInsets.symmetric(horizontal: 8, vertical: 10),
            floatingLabelBehavior: FloatingLabelBehavior.always,
            border: const OutlineInputBorder(),
            labelText: global.language("discount"),
          ),
        ),
      ),
    );
    if (global.posVersion == global.PosVersionEnum.restaurant) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: Row(
            children: [
              Switch(
                value: screenData.isalacarte!,
                onChanged: (bool? value) {
                  setState(() {
                    screenData.isalacarte = value ?? false;
                  });
                },
              ),
              Text(global.language("alacarte")),
            ],
          ),
        ),
      );
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: Row(
            children: [
              Switch(
                value: screenData.isstockforrestaurant!,
                onChanged: (bool? value) {
                  setState(() {
                    screenData.isstockforrestaurant = value ?? false;
                  });
                },
              ),
              Text(global.language("is_stock_for_restaurant")),
            ],
          ),
        ),
      );
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: Row(
            children: [
              Switch(
                value: screenData.issplitunitprint!,
                onChanged: (bool? value) {
                  setState(() {
                    screenData.issplitunitprint = value ?? false;
                  });
                },
              ),
              Text(global.language("is_split_unit_print")),
            ],
          ),
        ),
      );
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: Row(
            children: [
              Switch(
                value: screenData.isonlystaff!,
                onChanged: (bool? value) {
                  setState(() {
                    screenData.isonlystaff = value ?? false;
                  });
                },
              ),
              Text(global.language("is_only_employee")),
            ],
          ),
        ),
      );
    }

    // ลดราคา ณ จุดขาย + แสดงเวลาการขาย (combined row)
    formWidgets.add(
      Padding(
        padding: const EdgeInsets.only(left: 10, right: 10, bottom: 4),
        child: Row(
          children: [
            Expanded(
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Switch(
                    value: screenData.isdiscountpointofpurchase!,
                    onChanged: (bool? value) {
                      setState(() {
                        screenData.isdiscountpointofpurchase = value ?? false;
                      });
                    },
                  ),
                  Flexible(
                    child: Text(
                      global.language("is_discount_point_of_purchase"),
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                ],
              ),
            ),
            Expanded(
              child: Row(
                mainAxisSize: MainAxisSize.min,
                children: [
                  Switch(
                    value: isShowTimeForSale,
                    onChanged: (value) {
                      setState(() {
                        isShowTimeForSale = value;
                        if (!isShowTimeForSale) {
                          timeForSales.clear();
                          mediaFromDateController.clear();
                          mediaToDateController.clear();
                          mediaFromTimeController.clear();
                          mediaToTimeController.clear();
                          dayOfWeekSeleted.clear();
                        }
                      });
                    },
                  ),
                  Flexible(
                    child: Text(
                      "${global.language("show_time_for_sale")} ${(timeForSales.isNotEmpty) ? "(${timeForSales.length})" : ""}",
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
    if (global.posVersion == global.PosVersionEnum.restaurant) {
      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: Row(
            children: [
              Switch(
                value: screenData.restaurant!.isforrestaurant!,
                onChanged: (bool? value) {
                  setState(() {
                    screenData.restaurant!.isforrestaurant = value ?? false;
                  });
                },
              ),
              Text(global.language("is_for_restaurant")),
            ],
          ),
        ),
      );

      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: Row(
            children: [
              Switch(
                value: screenData.restaurant!.isfortakeaway!,
                onChanged: (bool? value) {
                  setState(() {
                    screenData.restaurant!.isfortakeaway = value ?? false;
                  });
                },
              ),
              Text(global.language("is_for_takeaway")),
            ],
          ),
        ),
      );

      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: Row(
            children: [
              Switch(
                value: screenData.restaurant!.isfordelivery!,
                onChanged: (bool? value) {
                  setState(() {
                    screenData.restaurant!.isfordelivery = value ?? false;
                  });
                },
              ),
              Text(global.language("is_for_delivery")),
            ],
          ),
        ),
      );

      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: Row(
            children: [
              Switch(
                value: screenData.restaurant!.isforcustomer!,
                onChanged: (bool? value) {
                  setState(() {
                    screenData.restaurant!.isforcustomer = value ?? false;
                  });
                },
              ),
              Text(global.language("is_for_customer")),
            ],
          ),
        ),
      );

      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: Row(
            children: [
              Switch(
                value: screenData.restaurant!.isforcustomerpreorder!,
                onChanged: (bool? value) {
                  setState(() {
                    screenData.restaurant!.isforcustomerpreorder =
                        value ?? false;
                  });
                },
              ),
              Text(global.language("is_for_customer_preorder")),
            ],
          ),
        ),
      );

      formWidgets.add(
        Padding(
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
          child: Column(
            children: [
              Row(
                children: [
                  Switch(
                    value: screenData.isalert!,
                    onChanged: (bool value) {
                      setState(() {
                        screenData.isalert = value;
                      });
                    },
                  ),
                  Text(global.language("is_use_alert")),
                ],
              ),
              (screenData.isalert!)
                  ? Padding(
                      padding: EdgeInsets.all(8.0),
                      child: Row(
                        children: [
                          Expanded(
                            child: TextField(
                              maxLines: 4,
                              decoration: InputDecoration(
                                floatingLabelBehavior:
                                    FloatingLabelBehavior.always,
                                border: OutlineInputBorder(),
                                labelText: global.language('disciption'),
                              ),
                              controller: alertDescriptionTextEditController,
                            ),
                          ),
                        ],
                      ),
                    )
                  : Container(),
            ],
          ),
        ),
      );
    }

    /// เวลาการขาย (switch อยู่ใน combined row ด้านบนแล้ว)
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Column(
          children: [
            (isShowTimeForSale)
                ? Container(
                    width: double.infinity,
                    padding: const EdgeInsets.only(bottom: 10, top: 10),
                    child: ElevatedButton.icon(
                      style: ElevatedButton.styleFrom(
                        backgroundColor: global.theme.positiveHighlightTextColor,
                        foregroundColor: global.theme.textColor,
                      ),
                      focusNode: FocusNode(skipTraversal: true),
                      onPressed: () {
                        timeForSales.add(
                          TimeForSaleModel(
                            fromdate: '',
                            todate: '',
                            fromtime: '',
                            totime: '',
                            daysofweek: [],
                          ),
                        );

                        mediaFromDateController.add(TextEditingController());
                        mediaToDateController.add(TextEditingController());
                        mediaFromTimeController.add(TextEditingController());
                        mediaToTimeController.add(TextEditingController());
                        _selectedTime.add(TimeOfDay.now());

                        dayOfWeekSeleted.add([
                          dayOfWeekList[0],
                          dayOfWeekList[1],
                          dayOfWeekList[2],
                          dayOfWeekList[3],
                          dayOfWeekList[4],
                          dayOfWeekList[5],
                          dayOfWeekList[6],
                        ]);

                        setState(() {});
                      },
                      icon: Icon(Icons.timer),
                      label: Text(global.language("add_time_for_sale")),
                    ),
                  )
                : Container(),
            for (
              int mediaIndex = 0;
              mediaIndex < timeForSales.length;
              mediaIndex++
            )
              /// time for sale
              (isShowTimeForSale)
                  ? Column(
                      children: [
                        Row(
                          children: [
                            Text(
                              "${global.language("time_for_sale")} ${mediaIndex + 1}",
                              style: TextStyle(
                                fontSize: 16,
                                color: global.theme.textColor,
                              ),
                            ),
                            IconButton(
                              focusNode: FocusNode(skipTraversal: true),
                              onPressed: () async {
                                timeForSales.removeAt(mediaIndex);

                                mediaFromDateController.removeAt(mediaIndex);
                                mediaToDateController.removeAt(mediaIndex);
                                mediaFromTimeController.removeAt(mediaIndex);
                                mediaToTimeController.removeAt(mediaIndex);
                                dayOfWeekSeleted.removeAt(mediaIndex);

                                setState(() {});
                              },
                              icon: Icon(Icons.delete, color: global.theme.negativeHighlightTextColor),
                            ),
                          ],
                        ),

                        const SizedBox(height: 10),

                        /// fromdate todate
                        Row(
                          children: [
                            Expanded(
                              child: CustomDatePicker(
                                labelText: global.language("from_date"),
                                useIconSelectDate: true,
                                initialDate: DateTime.tryParse(
                                  (timeForSales[mediaIndex].fromdate?.isNotEmpty == true)
                                      ? timeForSales[mediaIndex].fromdate.toString()
                                      : dateNow.toIso8601String(),
                                ),
                                firstDate: DateTime(2000),
                                lastDate: DateTime(2100),
                                onDateSelected: (date) {
                                  if (date != null) {
                                    setState(() {
                                      final adjustedDate = date.add(const Duration(hours: 7));
                                      timeForSales[mediaIndex].fromdate =
                                          adjustedDate.toIso8601String();
                                      mediaFromDateController[mediaIndex].text =
                                          DateFormat('dd/MM/yyyy').format(adjustedDate);
                                    });
                                  }
                                },
                                decoration: const InputDecoration(),
                              ),
                            ),
                            SizedBox(width: 10),
                            Expanded(
                              child: CustomDatePicker(
                                labelText: global.language("to_date"),
                                useIconSelectDate: true,
                                initialDate: DateTime.tryParse(
                                  (timeForSales[mediaIndex].todate?.isNotEmpty == true)
                                      ? timeForSales[mediaIndex].todate.toString()
                                      : dateNow.toIso8601String(),
                                ),
                                firstDate: DateTime(2000),
                                lastDate: DateTime(2100),
                                onDateSelected: (date) {
                                  if (date != null) {
                                    setState(() {
                                      final adjustedDate = date.add(const Duration(hours: 7));
                                      timeForSales[mediaIndex].todate =
                                          adjustedDate.toIso8601String();
                                      mediaToDateController[mediaIndex].text =
                                          DateFormat('dd/MM/yyyy').format(adjustedDate);
                                    });
                                  }
                                },
                                decoration: const InputDecoration(),
                              ),
                            ),
                          ],
                        ),

                        const SizedBox(height: 10),

                        /// fromtime totime
                        Row(
                          children: [
                            Expanded(
                              child: TextFormField(
                                controller: mediaFromTimeController[mediaIndex],
                                onTap: () async {
                                  _selectMediaFromTime(context, mediaIndex);
                                },
                                readOnly: true,
                                decoration: InputDecoration(
                                  labelText: global.language("from_time"),
                                  border: const OutlineInputBorder(),
                                  floatingLabelBehavior:
                                      FloatingLabelBehavior.always,
                                  suffixIcon: Row(
                                    mainAxisAlignment: MainAxisAlignment
                                        .spaceBetween, // added line
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      IconButton(
                                        focusNode: FocusNode(
                                          skipTraversal: true,
                                        ),
                                        icon: const Icon(Icons.timer_sharp),
                                        onPressed: () {
                                          _selectMediaFromTime(
                                            context,
                                            mediaIndex,
                                          );
                                        },
                                      ),
                                    ],
                                  ),
                                ),
                                inputFormatters: [global.TimeInputFormatter()],
                              ),
                            ),
                            SizedBox(width: 10),
                            Expanded(
                              child: TextFormField(
                                controller: mediaToTimeController[mediaIndex],
                                onTap: () async {
                                  _selectMediaToTime(context, mediaIndex);
                                },
                                readOnly: true,
                                decoration: InputDecoration(
                                  labelText: global.language("from_to"),
                                  border: const OutlineInputBorder(),
                                  floatingLabelBehavior:
                                      FloatingLabelBehavior.always,
                                  suffixIcon: Row(
                                    mainAxisAlignment: MainAxisAlignment
                                        .spaceBetween, // added line
                                    mainAxisSize: MainAxisSize.min,
                                    children: [
                                      IconButton(
                                        focusNode: FocusNode(
                                          skipTraversal: true,
                                        ),
                                        icon: const Icon(Icons.timer_sharp),
                                        onPressed: () {
                                          _selectMediaToTime(
                                            context,
                                            mediaIndex,
                                          );
                                        },
                                      ),
                                    ],
                                  ),
                                ),
                                inputFormatters: [global.TimeInputFormatter()],
                              ),
                            ),
                          ],
                        ),

                        const SizedBox(height: 10),

                        /// dayofweek checkbox
                        DropdownSearch<DayOfWeekModel>.multiSelection(
                          enabled: isEditMode,
                          items: (String filter, infiniteScrollProps) =>
                              getDataDayOfWeek(filter),
                          compareFn: (item, selectedItem) =>
                              item.code == selectedItem.code,
                          itemAsString: (DayOfWeekModel? group) {
                            if (group == null) return '';
                            return '${group.code} - ${group.name}';
                          },
                          decoratorProps: DropDownDecoratorProps(
                            decoration: InputDecoration(
                              border: const OutlineInputBorder(),
                              contentPadding: const EdgeInsets.only(
                                left: 10,
                                top: 15,
                                bottom: 10,
                                right: 10,
                              ),
                              floatingLabelBehavior:
                                  FloatingLabelBehavior.always,
                              labelText: global.language("day_of_week"),
                            ),
                          ),
                          onChanged: (List<DayOfWeekModel> value) {
                            setState(() {
                              dayOfWeekSeleted[mediaIndex] = value;
                            });
                          },
                          popupProps: const PopupPropsMultiSelection.dialog(
                            showSearchBox: false,
                            showSelectedItems: true,
                          ),
                          selectedItems: dayOfWeekSeleted[mediaIndex],
                        ),
                      ],
                    )
                  : Container(),
          ],
        ),
      ),
    );

    /// ประเภทธูรกิจ
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: DropdownSearch<BusinessTypeModel>.multiSelection(
          enabled: isEditMode,
          key: _popupBuilderBusinessKey,
          items: (String filter, infiniteScrollProps) =>
              getDataBusiness(filter),
          compareFn: (item, selectedItem) =>
              item.guidfixed == selectedItem.guidfixed,
          itemAsString: (BusinessTypeModel? businessTypeModel) {
            if (businessTypeModel == null) return '';
            return '${businessTypeModel.code} - ${global.activeLangName(businessTypeModel.names!)}';
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
              labelText: global.language("business_type"),
            ),
          ),
          onChanged: (List<BusinessTypeModel> value) {
            setState(() {
              screenData.businesstypes = value;
              loadDataBranchList(screenData.businesstypes!);
            });
          },
          popupProps: PopupPropsMultiSelection.dialog(
            showSearchBox: false,
            showSelectedItems: true,
            title: Padding(
              padding: EdgeInsets.all(8.0),
              child: Text(global.language("select_business_type")),
            ),
          ),
          selectedItems: screenData.businesstypes!,
        ),
      ),
    );

    /// สาขา
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: DropdownSearch<BranchModel>.multiSelection(
          enabled: isEditMode,
          key: _popupBuilderBranchKey,
          items: (String filter, infiniteScrollProps) => getDataBranch(filter),
          compareFn: (item, selectedItem) =>
              item.guidfixed == selectedItem.guidfixed,
          itemAsString: (BranchModel? branchModel) {
            if (branchModel == null) return '';
            return '${branchModel.code} - ${global.activeLangName(branchModel.names!)}';
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
              labelText: global.language("branch"),
            ),
          ),
          onChanged: (List<BranchModel> value) {
            setState(() {
              ignorebranches = value;
            });
          },
          popupProps: PopupPropsMultiSelection.dialog(
            showSearchBox: false,
            showSelectedItems: true,
            emptyBuilder: (context, searchEntry) => Center(
              child: (screenData.businesstypes!.isEmpty)
                  ? Text(global.language("please_select_business_type"))
                  : Text(global.language("business_type_not_match_branch")),
            ),
            title: Padding(
              padding: EdgeInsets.all(8.0),
              child: Text(global.language("select_branch")),
            ),
          ),
          selectedItems: ignorebranches!,
        ),
      ),
    );

    if (screenData.ordertypes!.isNotEmpty) {
      for (
        int optionTypeIndex = 0;
        optionTypeIndex < screenData.ordertypes!.length;
        optionTypeIndex++
      ) {
        formWidgets.add(
          Padding(
            padding: const EdgeInsets.only(left: 10, right: 10, bottom: 15),
            child: Row(
              children: [
                Expanded(
                  child: SizedBox(
                    height: 50,
                    child: ElevatedButton(
                      style: ButtonStyle(
                        backgroundColor: WidgetStateProperty.all<Color>(
                          (screenData
                                  .ordertypes![optionTypeIndex]
                                  .code
                                  .isNotEmpty)
                              ? global.theme.primaryLightColor
                              : global.theme.surfaceColor,
                        ),
                        foregroundColor: WidgetStateProperty.all<Color>(
                          global.theme.textColor,
                        ),
                      ),
                      onPressed: () {
                        orderTypeSearch(optionTypeIndex);
                      },
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Expanded(
                            child: Text(
                              (screenData
                                      .ordertypes![optionTypeIndex]
                                      .code
                                      .isEmpty)
                                  ? "${global.language("order_type_code")} ~ ${global.language("order_type_name")} "
                                  : "${screenData.ordertypes![optionTypeIndex].code} ~ ${global.activeLangName(screenData.ordertypes![optionTypeIndex].names)}  ",
                              overflow: TextOverflow.ellipsis,
                              maxLines: 2,
                            ),
                          ),
                          const Icon(Icons.search),
                        ],
                      ),
                    ),
                  ),
                ),
                SizedBox(width: 10),
                Expanded(
                  child: TextField(
                    enabled: (screenData
                        .ordertypes![optionTypeIndex]
                        .code
                        .isNotEmpty),
                    controller: TextEditingController(
                      text: screenData.ordertypes![optionTypeIndex].price
                          .toString(),
                    ),
                    keyboardType: TextInputType.number,
                    inputFormatters: <TextInputFormatter>[
                      FilteringTextInputFormatter.digitsOnly,
                    ],
                    textAlign: TextAlign.center,
                    decoration: InputDecoration(
                      floatingLabelBehavior: FloatingLabelBehavior.always,
                      border: OutlineInputBorder(),
                      labelText: global.language("charge_price"),
                    ),
                    onChanged: (value) {
                      if (value.isNotEmpty) {
                        screenData.ordertypes![optionTypeIndex].price =
                            double.parse(value.replaceAll(',', ''));
                      } else {
                        screenData.ordertypes![optionTypeIndex].price = 0;
                      }
                    },
                  ),
                ),
                const SizedBox(width: 10),
                IconButton(
                  onPressed: () {
                    setState(() {
                      screenData.ordertypes!.removeAt(optionTypeIndex);
                    });
                  },
                  icon: const Icon(Icons.delete),
                  focusNode: FocusNode(skipTraversal: true),
                  color: global.theme.negativeHighlightTextColor,
                  iconSize: 20,
                ),
              ],
            ),
          ),
        );
      }
    }

    /// ประเภทการสั่งอาหาร
    if (global.posVersion == global.PosVersionEnum.restaurant) {
      formWidgets.add(
        Container(
          height: 50,
          padding: const EdgeInsets.only(left: 10, right: 10, bottom: 10),
          width: double.infinity,
          child: ElevatedButton.icon(
            onPressed: () {
              setState(() {
                List<LanguageDataModel> names = [];
                for (int k = 0; k < languageList.length; k++) {
                  names.add(
                    LanguageDataModel(code: languageList[k].code!, name: ""),
                  );
                }
                screenData.ordertypes!.add(
                  OrderTypeProductBarcode(
                    guidfixed: '',
                    code: '',
                    names: names,
                    price: 0,
                  ),
                );
              });
            },
            style: ButtonStyle(
              backgroundColor: WidgetStateProperty.all<Color>(
                global.theme.primaryColor,
              ), // Set the background color here
            ),
            icon: Icon(Icons.add), // Set the icon here
            label: Text(global.language("add_order_type")),
          ),
        ),
      );
    }

    if (screenData.options!.isNotEmpty) {
      formWidgets.add(
        Container(
          padding: const EdgeInsets.only(left: 10, right: 10, bottom: 10),
          width: double.infinity,
          height: 50,
          child: ElevatedButton(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () {
              optionSearch();
            },
            child: Text(global.language("clone_option")),
          ),
        ),
      );
      for (
        int optionIndex = 0;
        optionIndex < screenData.options!.length;
        optionIndex++
      ) {
        List<Widget> optionList = [];
        optionList.add(
          Container(
            padding: const EdgeInsets.only(top: 5, bottom: 10),
            width: double.infinity,
            child: Row(
              children: [
                Expanded(
                  child: Padding(
                    padding: EdgeInsets.only(left: 10),
                    child: Text(
                      "${global.language("option")} ${optionIndex + 1}",
                      style: const TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                ),
                IconButton(
                  onPressed: () {
                    setState(() {
                      screenData.options!.removeAt(optionIndex);
                    });
                  },
                  icon: const Icon(Icons.delete),
                  focusNode: FocusNode(skipTraversal: true),
                  color: global.theme.negativeHighlightTextColor,
                  iconSize: 20,
                ),
                if (optionIndex > 0)
                  IconButton(
                    onPressed: () {
                      setState(() {
                        screenData.options!.insert(
                          optionIndex - 1,
                          screenData.options!.removeAt(optionIndex),
                        );
                      });
                    },
                    icon: Icon(Icons.move_up),
                    focusNode: FocusNode(skipTraversal: true),
                    color: global.theme.primaryColor,
                    iconSize: 20,
                  ),
                if (optionIndex < screenData.options!.length - 1)
                  IconButton(
                    onPressed: () {
                      setState(() {
                        screenData.options!.insert(
                          optionIndex + 1,
                          screenData.options!.removeAt(optionIndex),
                        );
                      });
                    },
                    icon: const Icon(Icons.move_down),
                    color: global.theme.negativeHighlightTextColor,
                    focusNode: FocusNode(skipTraversal: true),
                    iconSize: 20,
                  ),
              ],
            ),
          ),
        );
        // for (int languageIndex = 0; languageIndex < languageList.length; languageIndex++) {
        //   LanguageDataModel optionName =
        //       screenData.options![optionIndex].names.firstWhere((element) => element.code == languageList[languageIndex].code, orElse: () => LanguageDataModel(code: '', name: ''));
        //   if (optionName.code == '') {
        //     screenData.options![optionIndex].names.add(LanguageDataModel(code: languageList[languageIndex].code!, name: ''));
        //   }

        //   optionList.add(Padding(
        // padding: EdgeInsets.only(left: 5, right: 5, bottom: 10),
        //     child: TextField(
        //       readOnly: !isEditMode,
        //       onChanged: (value) {
        //         isDataChange = true;
        //         screenData.options![optionIndex].names[languageIndex].name = value;
        //       },
        //       onSubmitted: (value) {
        //         findFocusNext(focusNodeIndex);
        //       },
        //       focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
        //       textAlign: TextAlign.left,
        //       controller: TextEditingController(text: screenData.options![optionIndex].names[languageIndex].name),
        //       decoration: InputDecoration(
        //         floatingLabelBehavior: FloatingLabelBehavior.always,
        // border: OutlineInputBorder(),
        //         labelText: "${global.language("option_name")} (${getLangName(screenData.options![optionIndex].names[languageIndex].code)})",
        //       ),
        //     ),
        //   ));
        // }
        optionList.addAll(
          listNamesFields(
            screenData.options![optionIndex].names,
            "option_name",
          ),
        );
        optionList.add(
          Row(
            children: [
              Expanded(
                child: ListTile(
                  title: Text(
                    global.language("product_option_choice_type_multi"),
                  ),
                  leading: Radio(
                    value: 0,
                    groupValue: screenData.options![optionIndex].choicetype,
                    onChanged: (value) {
                      setState(() {
                        screenData.options![optionIndex].choicetype = value!;
                      });
                    },
                  ),
                ),
              ),
              Expanded(
                child: ListTile(
                  title: Text(
                    global.language("product_option_choice_type_single"),
                  ),
                  leading: Radio(
                    value: 1,
                    groupValue: screenData.options![optionIndex].choicetype,
                    onChanged: (value) {
                      setState(() {
                        screenData.options![optionIndex].choicetype = value!;
                      });
                    },
                  ),
                ),
              ),
            ],
          ),
        );
        List<Widget> choiceList = [];
        for (
          int choiceIndex = 0;
          choiceIndex < screenData.options![optionIndex].choices.length;
          choiceIndex++
        ) {
          List<Widget> choiceRow = [];
          if (choiceList.isNotEmpty) {
            choiceRow.add(Divider(color: global.theme.textColor));
          }
          choiceRow.add(
            Row(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                Expanded(
                  child: Padding(
                    padding: EdgeInsets.all(10),
                    child: Text(
                      "${global.language("choice")} ${choiceIndex + 1}",
                      style: const TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                  ),
                ),
                IconButton(
                  onPressed: () {
                    setState(() {
                      screenData.options![optionIndex].choices.removeAt(
                        choiceIndex,
                      );
                    });
                  },
                  icon: const Icon(Icons.delete),
                  focusNode: FocusNode(skipTraversal: true),
                  color: global.theme.negativeHighlightTextColor,
                  iconSize: 20,
                ),
                if (choiceIndex > 0)
                  IconButton(
                    onPressed: () {
                      setState(() {
                        screenData.options![optionIndex].choices.insert(
                          choiceIndex - 1,
                          screenData.options![optionIndex].choices.removeAt(
                            choiceIndex,
                          ),
                        );
                      });
                    },
                    icon: Icon(Icons.move_up),
                    focusNode: FocusNode(skipTraversal: true),
                    color: global.theme.primaryColor,
                    iconSize: 20,
                  ),
                if (choiceIndex <
                    screenData.options![optionIndex].choices.length - 1)
                  IconButton(
                    onPressed: () {
                      setState(() {
                        screenData.options![optionIndex].choices.insert(
                          choiceIndex + 1,
                          screenData.options![optionIndex].choices.removeAt(
                            choiceIndex,
                          ),
                        );
                      });
                    },
                    icon: const Icon(Icons.move_down),
                    color: global.theme.positiveHighlightTextColor,
                    focusNode: FocusNode(skipTraversal: true),
                    iconSize: 20,
                  ),
              ],
            ),
          );

          choiceRow.add(
            Padding(
              padding: EdgeInsets.only(left: 10, right: 10),
              child: Column(
                children: [
                  Align(
                    alignment: Alignment.centerLeft,
                    child: Text(
                      global.language("image"),
                      style: Theme.of(context).textTheme.titleMedium,
                    ),
                  ),
                  SizedBox(height: 10),
                  Row(
                    children: [
                      Expanded(
                        child: ElevatedButton.icon(
                          focusNode: FocusNode(skipTraversal: true),
                          onPressed: () async {
                            setState(() {
                              screenData
                                      .options![optionIndex]
                                      .choices[choiceIndex]
                                      .imageuri =
                                  "";
                            });
                          },
                          icon: Icon(Icons.delete),
                          label: Text(global.language('delete_picture')),
                        ),
                      ),
                      SizedBox(width: 5),
                      Expanded(
                        child: ElevatedButton.icon(
                          focusNode: FocusNode(skipTraversal: true),
                          onPressed: () async {
                            XFile? image = await imagePicker.pickImage(
                              source: ImageSource.gallery,
                              maxHeight: 480,
                              maxWidth: 640,
                              imageQuality: 60,
                            );
                            if (image != null) {
                              var f = await image.readAsBytes();
                              setState(() {
                                imageWebChoice[0] = f;
                                imageFileChoice[0] = File(image.path);
                                uploadImageOptionIndex = optionIndex;
                                uploadImageChoiceIndex = choiceIndex;
                                upLoadImageChoice();
                              });
                            }
                          },
                          icon: Icon(Icons.folder),
                          label: Text(global.language("select_picture")),
                        ),
                      ),
                      const SizedBox(width: 5),
                    ],
                  ),
                  SizedBox(
                    width: 300,
                    height: 300,
                    child: Stack(
                      children: [
                        Center(
                          child: DecoratedBox(
                            decoration: BoxDecoration(
                              color: global.theme.cardColor,
                              border: Border.all(color: global.theme.textColor),
                              borderRadius: BorderRadius.circular(5),
                              image:
                                  (screenData
                                          .options![optionIndex]
                                          .choices[choiceIndex]
                                          .imageuri !=
                                      '')
                                  ? DecorationImage(
                                      image: NetworkImage(
                                        global.resolveFileUrl(screenData
                                            .options![optionIndex]
                                            .choices[choiceIndex]
                                            .imageuri!),
                                      ),
                                      fit: BoxFit.fill,
                                    )
                                  : const DecorationImage(
                                      image: AssetImage(
                                        'assets/img/noimage.png',
                                      ),
                                    ),
                            ),
                            child: const SizedBox(
                              width: double.infinity,
                              height: 200,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          );

          choiceRow.addAll(
            listNamesFields(
              screenData.options![optionIndex].choices[choiceIndex].names,
              "choice_name",
            ),
          );
          choiceRow.add(
            Padding(
              padding: const EdgeInsets.only(left: 10, right: 10, bottom: 10),
              child: TextField(
                readOnly: !isEditMode,
                onChanged: (value) {
                  isDataChange = true;
                  screenData.options![optionIndex].choices[choiceIndex].price =
                      value;
                },
                onSubmitted: (value) {
                  findFocusNext(focusNodeIndex);
                },
                textAlign: TextAlign.right,
                keyboardType: const TextInputType.numberWithOptions(
                  decimal: true,
                ),
                inputFormatters: [global.NumberInputFormatter()],
                focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
                controller: TextEditingController(
                  text: screenData
                      .options![optionIndex]
                      .choices[choiceIndex]
                      .price
                      .toString(),
                ),
                decoration: InputDecoration(
                  floatingLabelBehavior: FloatingLabelBehavior.always,
                  border: OutlineInputBorder(),
                  labelText: global.language("choice_add_price_percent"),
                ),
              ),
            ),
          );
          choiceRow.add(
            Padding(
              padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
              child: SwitchListTile(
                contentPadding: EdgeInsets.zero,
                focusNode: FocusNode(skipTraversal: true),
                title: Text(global.language("choice_is_stock")),
                value: screenData
                    .options![optionIndex]
                    .choices[choiceIndex]
                    .isstock,
                onChanged: ((value) {
                  setState(() {
                    screenData
                        .options![optionIndex]
                        .choices[choiceIndex]
                        .isstock = !screenData
                        .options![optionIndex]
                        .choices[choiceIndex]
                        .isstock;
                  });
                }),
              ),
            ),
          );
          choiceRow.add(
            Padding(
              padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
              child: SwitchListTile(
                contentPadding: EdgeInsets.zero,
                focusNode: FocusNode(skipTraversal: true),
                title: Text(global.language("choice_is_select")),
                value: screenData
                    .options![optionIndex]
                    .choices[choiceIndex]
                    .isdefault!,
                onChanged: ((value) {
                  setState(() {
                    screenData
                        .options![optionIndex]
                        .choices[choiceIndex]
                        .isdefault = !screenData
                        .options![optionIndex]
                        .choices[choiceIndex]
                        .isdefault!;
                  });
                }),
              ),
            ),
          );
          if (screenData.options![optionIndex].choices[choiceIndex].isstock) {
            choiceRow.add(
              Padding(
                padding: const EdgeInsets.only(left: 10, right: 10, bottom: 10),
                child: TextField(
                  readOnly: true,
                  onSubmitted: (value) {
                    findFocusNext(focusNodeIndex);
                  },
                  focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
                  textAlign: TextAlign.left,
                  controller: TextEditingController(
                    text:
                        "${screenData.options![optionIndex].choices[choiceIndex].refbarcode} ~ ${global.activeLangName(screenData.options![optionIndex].choices[choiceIndex].refbarcodenames!)}",
                  ),
                  textCapitalization: TextCapitalization.characters,
                  inputFormatters: <TextInputFormatter>[
                    FilteringTextInputFormatter.allow(
                      RegExp('[a-z A-Z 0-9 -]'),
                    ),
                  ],
                  onChanged: (value) {
                    isDataChange = true;
                    screenData
                        .options![optionIndex]
                        .choices[choiceIndex]
                        .refbarcode = value
                        .toUpperCase();
                  },
                  decoration: InputDecoration(
                    floatingLabelBehavior: FloatingLabelBehavior.always,
                    suffixIcon: IconButton(
                      focusNode: FocusNode(skipTraversal: true),
                      icon: const Icon(Icons.search),
                      onPressed: () {
                        String word = screenData
                            .options![optionIndex]
                            .choices[choiceIndex]
                            .refbarcode;
                        barcodeSearch(word, optionIndex, choiceIndex);
                      },
                    ),
                    border: OutlineInputBorder(),
                    labelText: global.language("barcode"),
                  ),
                ),
              ),
            );

            choiceRow.add(
              Padding(
                padding: const EdgeInsets.only(left: 10, right: 10, bottom: 10),
                child: TextField(
                  readOnly: !isEditMode,
                  onChanged: (value) {
                    isDataChange = true;
                    screenData.options![optionIndex].choices[choiceIndex].qty =
                        double.tryParse(value)!;
                  },
                  onSubmitted: (value) {
                    findFocusNext(focusNodeIndex);
                  },
                  textAlign: TextAlign.right,
                  keyboardType: const TextInputType.numberWithOptions(
                    decimal: true,
                  ),
                  inputFormatters: [global.NumberInputFormatter()],
                  focusNode: fieldFocusNodes[++focusNodeMax].focusNode,
                  controller: TextEditingController(
                    text:
                        (screenData
                                .options![optionIndex]
                                .choices[choiceIndex]
                                .qty ==
                            0)
                        ? ""
                        : screenData
                              .options![optionIndex]
                              .choices[choiceIndex]
                              .qty
                              .toString(),
                  ),
                  decoration: InputDecoration(
                    floatingLabelBehavior: FloatingLabelBehavior.always,
                    border: OutlineInputBorder(),
                    labelText: global.language("choice_stock_qty"),
                  ),
                ),
              ),
            );
          }
          choiceList.add(Column(children: choiceRow));
        }
        if (choiceList.isNotEmpty) {
          optionList.add(
            Container(
              padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
              width: double.infinity,
              child: Container(
                decoration: BoxDecoration(
                  color: global.theme.surfaceColor,
                  border: Border.all(color: global.theme.textSecondaryColor),
                  borderRadius: BorderRadius.circular(5),
                ),
                width: double.infinity,
                child: Column(children: choiceList),
              ),
            ),
          );
        }
        if (isEditMode) {
          optionList.add(
            Container(
              padding: const EdgeInsets.only(left: 10, right: 10, bottom: 10),
              width: double.infinity,
              child: ElevatedButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () {
                  setState(() {
                    List<LanguageDataModel> names = [];
                    for (int k = 0; k < languageList.length; k++) {
                      names.add(
                        LanguageDataModel(
                          code: languageList[k].code!,
                          name: "",
                        ),
                      );
                    }
                    screenData.options![optionIndex].choices.add(
                      ProductChoiceModel(
                        guid: const Uuid().v4(),
                        refbarcode: "",
                        refbarcodenames: [],
                        isstock: false,
                        isdefault: false,
                        refproductcode: "",
                        refunitcode: "",
                        names: names,
                        qty: 0,
                        price: "",
                      ),
                    );
                  });
                },
                child: Text(global.language("add_choice")),
              ),
            ),
          );
        }
        formWidgets.add(
          Container(
            padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
            width: double.infinity,
            child: Container(
              decoration: BoxDecoration(
                color: global.theme.surfaceColor,
                border: Border.all(color: global.theme.textSecondaryColor),
                borderRadius: BorderRadius.circular(5),
              ),
              width: double.infinity,
              child: Column(children: optionList),
            ),
          ),
        );
        if (screenData.options![optionIndex].choicetype == 0) {
          optionList.add(
            Padding(
              padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
              child: Row(
                children: [
                  Expanded(
                    child: TextField(
                      keyboardType: TextInputType.number,
                      inputFormatters: <TextInputFormatter>[
                        FilteringTextInputFormatter.digitsOnly,
                      ],
                      readOnly: true,
                      focusNode: FocusNode(skipTraversal: true),
                      textAlign: TextAlign.center,
                      controller: TextEditingController(
                        text: screenData.options![optionIndex].minselect
                            .toString(),
                      ),
                      decoration: InputDecoration(
                        floatingLabelBehavior: FloatingLabelBehavior.always,
                        border: OutlineInputBorder(),
                        labelText: global.language("option_min_select_choice"),
                        prefixIcon: IconButton(
                          focusNode: FocusNode(skipTraversal: true),
                          icon: const Icon(Icons.arrow_downward),
                          onPressed: () {
                            setState(() {
                              if (screenData.options![optionIndex].minselect >
                                  0) {
                                screenData.options![optionIndex].minselect--;
                              }
                            });
                          },
                        ),
                        suffixIcon: IconButton(
                          focusNode: FocusNode(skipTraversal: true),
                          icon: const Icon(Icons.arrow_upward),
                          onPressed: () {
                            setState(() {
                              if (screenData.options![optionIndex].minselect <=
                                  screenData
                                      .options![optionIndex]
                                      .choices
                                      .length) {
                                screenData.options![optionIndex].minselect++;
                              }
                            });
                          },
                        ),
                      ),
                    ),
                  ),
                  SizedBox(width: 5),
                  Expanded(
                    child: TextField(
                      keyboardType: TextInputType.number,
                      inputFormatters: <TextInputFormatter>[
                        FilteringTextInputFormatter.digitsOnly,
                      ],
                      readOnly: true,
                      focusNode: FocusNode(skipTraversal: true),
                      textAlign: TextAlign.center,
                      controller: TextEditingController(
                        text: screenData.options![optionIndex].maxselect
                            .toString(),
                      ),
                      decoration: InputDecoration(
                        floatingLabelBehavior: FloatingLabelBehavior.always,
                        border: OutlineInputBorder(),
                        labelText: global.language("option_max_select_choice"),
                        prefixIcon: IconButton(
                          focusNode: FocusNode(skipTraversal: true),
                          icon: const Icon(Icons.arrow_downward),
                          onPressed: () {
                            setState(() {
                              if (screenData.options![optionIndex].maxselect >
                                  1) {
                                screenData.options![optionIndex].maxselect--;
                              }
                            });
                          },
                        ),
                        suffixIcon: IconButton(
                          focusNode: FocusNode(skipTraversal: true),
                          icon: const Icon(Icons.arrow_upward),
                          onPressed: () {
                            setState(() {
                              if (screenData.options![optionIndex].maxselect <=
                                  screenData
                                      .options![optionIndex]
                                      .choices
                                      .length) {
                                screenData.options![optionIndex].maxselect++;
                              }
                            });
                          },
                        ),
                      ),
                    ),
                  ),
                ],
              ),
            ),
          );
        }
      }
    }

    if (global.posVersion == global.PosVersionEnum.restaurant) {
      /// ตัวเลือก
      formWidgets.add(
        Container(
          padding: const EdgeInsets.only(left: 10, right: 10, bottom: 10),
          width: double.infinity,
          height: 50,
          child: ElevatedButton(
            focusNode: FocusNode(skipTraversal: true),
            onPressed: () {
              setState(() {
                List<LanguageDataModel> names = [];
                for (int k = 0; k < languageList.length; k++) {
                  names.add(
                    LanguageDataModel(code: languageList[k].code!, name: ""),
                  );
                }
                screenData.options!.add(
                  ProductOptionModel(
                    guid: const Uuid().v4(),
                    choicetype: 0,
                    minselect: 1,
                    maxselect: 1,
                    names: names,
                    choices: [],
                  ),
                );
              });
            },
            child: Text(global.language("add_option")),
          ),
        ),
      );
    }

    /// ต้นทุนการขาย
    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Column(
          children: [
            Container(
              width: double.infinity,
              height: 60,
              padding: EdgeInsets.only(bottom: 10, top: 10),
              child: ElevatedButton.icon(
                style: ElevatedButton.styleFrom(
                  backgroundColor: global.theme.positiveHighlightTextColor,
                  foregroundColor: global.theme.onPrimaryColor,
                ),
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () {
                  fiexdCostList.add(FiexdCostModel(effectdate: '', amount: 0));

                  asDateController.add(TextEditingController());
                  fiexdCostAmountController.add(TextEditingController());

                  setState(() {});
                },
                icon: Icon(Icons.add),
                label: Text(
                  "${global.language("add")} ${global.language("cost_standard")}",
                ),
              ),
            ),
            if (fiexdCostList.isNotEmpty)
              for (
                int fiexdCostIndex = 0;
                fiexdCostIndex < fiexdCostList.length;
                fiexdCostIndex++
              )
                Column(
                  children: [
                    Row(
                      children: [
                        Text(
                          "${global.language("cost_standard")} ${fiexdCostIndex + 1}",
                          style: TextStyle(
                            fontSize: 16,
                            color: global.theme.textColor,
                          ),
                        ),
                        IconButton(
                          focusNode: FocusNode(skipTraversal: true),
                          onPressed: () async {
                            fiexdCostList.removeAt(fiexdCostIndex);

                            asDateController.removeAt(fiexdCostIndex);
                            fiexdCostAmountController.removeAt(fiexdCostIndex);

                            setState(() {});
                          },
                          icon: Icon(Icons.delete, color: global.theme.negativeHighlightTextColor),
                        ),
                      ],
                    ),
                    SizedBox(height: 10),
                    CustomDatePicker(
                      labelText: global.language("as_date"),
                      useIconSelectDate: true,
                      initialDate: DateTime.tryParse(
                        (fiexdCostList[fiexdCostIndex].effectdate?.isNotEmpty == true)
                            ? fiexdCostList[fiexdCostIndex].effectdate.toString()
                            : dateNow.toIso8601String(),
                      ),
                      firstDate: DateTime(2000),
                      lastDate: DateTime(2100),
                      onDateSelected: (date) {
                        if (date != null) {
                          setState(() {
                            final adjustedDate = date.add(const Duration(hours: 7));
                            fiexdCostList[fiexdCostIndex].effectdate =
                                adjustedDate.toIso8601String();
                            asDateController[fiexdCostIndex].text =
                                DateFormat('dd/MM/yyyy').format(adjustedDate);
                          });
                        }
                      },
                      decoration: const InputDecoration(),
                    ),
                    const SizedBox(height: 10),

                    /// amount
                    TextField(
                      readOnly: !isEditMode,
                      keyboardType: const TextInputType.numberWithOptions(
                        decimal: true,
                      ),
                      inputFormatters: [global.NumberInputFormatter()],
                      onChanged: (value) {
                        if (value != '') {
                          fiexdCostList[fiexdCostIndex].amount = double.parse(
                            value.replaceAll(',', ''),
                          );
                        } else {
                          fiexdCostList[fiexdCostIndex].amount = 0;
                        }
                      },
                      textAlign: TextAlign.right,
                      controller: fiexdCostAmountController[fiexdCostIndex],
                      decoration: InputDecoration(
                        floatingLabelBehavior: FloatingLabelBehavior.always,
                        border: OutlineInputBorder(),
                        labelText: "${global.language("amount")} ",
                      ),
                    ),

                    const SizedBox(height: 10),
                  ],
                ),
          ],
        ),
      ),
    );

    formWidgets.add(
      Container(
        width: double.infinity,
        padding: EdgeInsets.only(bottom: 10),
        child: Row(
          children: [
            Radio(
              focusNode: FocusNode(skipTraversal: true),
              value: true,
              groupValue: screenData.useimageorcolor,
              onChanged: (value) {
                if (isEditMode) {
                  setState(() {
                    screenData.useimageorcolor = true;
                  });
                }
              },
            ),
            Text(global.language("use_image")),
            const SizedBox(width: 10),
            Radio(
              focusNode: FocusNode(skipTraversal: true),
              value: false,
              groupValue: screenData.useimageorcolor,
              onChanged: (value) {
                if (isEditMode) {
                  setState(() {
                    screenData.useimageorcolor = false;
                  });
                }
              },
            ),
            Text(global.language("use_color")),
          ],
        ),
      ),
    );

    if (isEditMode) {
      if (screenData.useimageorcolor!) {
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
                        screenData.imageuri = "";
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
                        : () async {
                            FocusScope.of(context).unfocus();
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
    }
    if (screenData.useimageorcolor!) {
      formWidgets.add(
        Container(
          padding: const EdgeInsets.only(left: 10, right: 10, bottom: 10),
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
                    color: global.theme.cardColor,
                    border: Border.all(color: global.theme.textColor),
                    borderRadius: BorderRadius.circular(5),
                    image: (imageWeb != null)
                        ? DecorationImage(
                            image: MemoryImage(imageWeb!),
                            fit: BoxFit.fill,
                          )
                        : (screenData.imageuri != null && screenData.imageuri!.isNotEmpty)
                        ? DecorationImage(
                            image: NetworkImage(global.resolveFileUrl(screenData.imageuri!)),
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
    } else {
      formWidgets.add(
        Container(
          margin: EdgeInsets.all(8),
          width: double.infinity,
          decoration: BoxDecoration(
            color: colorSelected,
            border: Border.all(color: global.theme.textSecondaryColor),
            borderRadius: BorderRadius.circular(5),
          ),
          child: Container(
            padding: EdgeInsets.all(10),
            child: ColorPicker(
              color: colorSelected,
              padding: EdgeInsets.all(0),
              heading: Container(
                width: double.infinity,
                decoration: BoxDecoration(
                  color: global.theme.cardColor,
                  border: Border.all(color: global.theme.textSecondaryColor),
                  borderRadius: BorderRadius.circular(5),
                ),
                child: Row(
                  children: [
                    Container(
                      padding: EdgeInsets.only(left: 10),
                      child: Text(
                        global.language('select_color_preview'),
                        style: Theme.of(context).textTheme.titleMedium,
                      ),
                    ),
                    const Spacer(),
                    IconButton(
                      icon: const Icon(Icons.clear),
                      onPressed: () {
                        setState(() {
                          colorSelected = Colors.white;
                        });
                      },
                    ),
                  ],
                ),
              ),
              showColorName: true,
              showColorCode: true,
              onColorChanged: (Color colorValue) {
                setState(() {
                  colorSelected = colorValue;
                  screenData.colorselect = colorValue.value.toString();
                  screenData.colorselecthex = colorValue.value
                      .toRadixString(16)
                      .toString();
                });
              },
              pickersEnabled: const <ColorPickerType, bool>{
                ColorPickerType.both: true,
                ColorPickerType.primary: true,
                ColorPickerType.accent: true,
                ColorPickerType.bw: false,
                ColorPickerType.custom: true,
                ColorPickerType.wheel: true,
              },
            ),
          ),
        ),
      );
    }

    formWidgets.add(
      Padding(
        padding: EdgeInsets.only(left: 10, right: 10, bottom: 15),
        child: Container(
          margin: EdgeInsets.only(bottom: 10),
          child: Row(
            children: [
              Expanded(
                child: TextField(
                  maxLines: 4,
                  decoration: InputDecoration(
                    floatingLabelBehavior: FloatingLabelBehavior.always,
                    border: OutlineInputBorder(),
                    labelText: global.language('disciption'),
                  ),
                  controller: descriptionTextEditController,
                ),
              ),
            ],
          ),
        ),
      ),
    );
    // แผนผังการแปลงหน่วยสินค้า — รวบรวม barcode ทั้งหมดจาก itemcode เดียวกัน
    if (screenData.itemcode != null && screenData.itemcode!.isNotEmpty) {
      final allBarcodesForProduct = listData
          .where((item) => item.itemcode == screenData.itemcode)
          .toList();
      formWidgets.add(
        Padding(
          padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
          child: UnitFlowWidget(
            productBarcodes: allBarcodesForProduct,
            currentBarcode: selectBarcode,
          ),
        ),
      );
    }

    if (isSaveAllow) {
      formWidgets.add(
        Container(
          width: double.infinity,
          padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
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
      backgroundColor: global.theme.cardColor,
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
        title: Text(headerEdit + global.language("barcode")),
        actions: <Widget>[
          EditFontSizeControl(onChanged: () => setState(() {})),
          const SizedBox(width: 8),
          if (selectGuid.isNotEmpty)
            Tooltip(
              message: global.language("show_product_bom"),
              child: Padding(
                padding: const EdgeInsets.only(right: 20.0),
                child: IconButton(
                  focusNode: FocusNode(skipTraversal: true),
                  onPressed: () {
                    isPreviewBom = !isPreviewBom;
                    setState(() {});
                  },
                  icon: Icon(
                    (!isPreviewBom)
                        ? Icons.account_tree_outlined
                        : Icons.account_tree_rounded,
                  ),
                ),
              ),
            ),
          (isPreview)
              ? Padding(
                  padding: const EdgeInsets.only(right: 20.0),
                  child: IconButton(
                    focusNode: FocusNode(skipTraversal: true),
                    onPressed: () {
                      isPreview = false;
                      isPreviewBom = false;
                      setState(() {});
                    },
                    icon: const Icon(Icons.text_fields),
                  ),
                )
              : Padding(
                  padding: const EdgeInsets.only(right: 20.0),
                  child: IconButton(
                    focusNode: FocusNode(skipTraversal: true),
                    onPressed: () {
                      isPreview = true;
                      isPreviewBom = false;
                      setState(() {});
                    },
                    icon: const Icon(Icons.preview),
                  ),
                ),
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
                      title: Text(global.language('coppy_confirm')),
                      actions: <Widget>[
                        ElevatedButton(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: global.theme.negativeHighlightTextColor,
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

                            discardData(
                              callBack: () {
                                setState(() {
                                  isEditMode = true;
                                  isPreview = false;
                                  selectGuid = "";
                                  selectBarcode = "";
                                  showCheckBox = false;
                                  isDataChange = false;
                                  // clearEditData();
                                  screenData.guidfixed = "";
                                  screenData.barcode = "";
                                  headerEdit = global.language("append");
                                  isSaveAllow = true;
                                  if (mobileScreen) {
                                    WidgetsBinding.instance
                                        .addPostFrameCallback((timeStamp) {
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
                          child: Text(global.language('confirm')),
                        ),
                      ],
                    ),
                  );
                  setState(() {});
                },
                icon: const Icon(Icons.copy),
              ),
            ),
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
                            backgroundColor: global.theme.negativeHighlightTextColor,
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
                            context.read<ProductBarcodeBloc>().add(
                              ProductBarcodeDelete(guid: screenData.guidfixed),
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
                    context.loaderOverlay.show(
                      widgetBuilder: (progress) {
                        return const ReconnectingOverlay();
                      },
                    );
                    isLoadTranslation = true;
                  });
                  try {
                    await translateNames();
                    await translateOptions();
                    setState(() {
                      isLoadTranslation = false;
                      context.loaderOverlay.hide();
                    });
                  } catch (e) {
                    if (kDebugMode) {
                      AppLogger.debug(e);
                    }
                  }
                },
                icon: const Icon(Icons.translate),
              ),
            ),
          if (isSaveAllow == false && selectBarcode.trim().isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () {
                  showCheckBox = false;
                  switchToEdit(
                    listData.firstWhere(
                      (element) => (element.barcode ?? '') == selectBarcode,
                    ),
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
      body: LoaderOverlay(
        overlayColor: Colors.black.withValues(alpha: 0.8),
        child: Focus(
          skipTraversal: true,
          onKeyEvent: (node, event) {
            if (event is KeyUpEvent) return KeyEventResult.ignored;
            if (event.logicalKey == LogicalKeyboardKey.f10) {
              if (event is KeyDownEvent) {
                if (_formKey.currentState!.validate()) {
                  saveOrUpdateData();
                }
              }
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
              // เพิ่ม spacing ระหว่าง formWidgets ตาม scale
              List<Widget> scaledFormWidgets = formWidgets;
              if (scale > 1.0) {
                final extraGap = 8.0 * (scale - 1);
                scaledFormWidgets = [];
                for (int i = 0; i < formWidgets.length; i++) {
                  scaledFormWidgets.add(formWidgets[i]);
                  if (i < formWidgets.length - 1) {
                    scaledFormWidgets.add(SizedBox(height: extraGap));
                  }
                }
              }
              return MediaQuery(
                data: MediaQuery.of(context).copyWith(
                  textScaler: TextScaler.linear(scale),
                ),
                child: Theme(
                  data: Theme.of(context).copyWith(
                    inputDecorationTheme: Theme.of(context).inputDecorationTheme.copyWith(
                      contentPadding: EdgeInsets.fromLTRB(
                        12 * scale, 20 * scale, 12 * scale, 12 * scale,
                      ),
                    ),
                  ),
                  child: IconTheme(
                    data: IconTheme.of(context).copyWith(
                      size: 24.0 * scale,
                    ),
                    child: SingleChildScrollView(
                      controller: editScrollController,
                      scrollDirection: Axis.vertical,
                      child: Padding(
                        padding: EdgeInsets.only(
                          top: 10 * scale, bottom: 10 * scale,
                        ),
                        child: (!isPreviewBom)
                            ? Form(
                                key: _formKey,
                                child: Column(
                                  children: (isPreview)
                                      ? [
                                          ProductPreviewScreen(
                                            screenData: screenData,
                                            priceList: priceList,
                                            imageWeb: imageWeb,
                                            allProductBarcodes: listData
                                                .where((item) =>
                                                    item.itemcode ==
                                                    screenData.itemcode)
                                                .toList(),
                                          ),
                                        ]
                                      : scaledFormWidgets,
                                ),
                              )
                            : ProductBomWidget(productBom: productBom),
                      ),
                    ),
                  ),
                ),
              );
            },
          ),
          ),
  ),
),
    );
  }

  String productTypeName(int type) {
    String productTypeName = "";
    if (type == 0) {
      productTypeName = global.language("product_is_stock");
    } else if (type == 1) {
      productTypeName = global.language("product_is_service");
    } else if (type == 2) {
      productTypeName = global.language("product_is_set");
    } else if (type == 3) {
      productTypeName = global.language("product_is_material");
    }
    return productTypeName;
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
              BlocListener<ProductBarcodeBloc, ProductBarcodeState>(
                listener: (context, state) async {
                  blocCurrentState = state;
                  // Load
                  if (state is ProductBarcodeLoadSuccess) {
                    AppLogger.info(
                      '✅ LOAD SUCCESS: โหลดข้อมูลสำเร็จ (จำนวน: ${state.productBarcodes.length} รายการ)',
                    );
                    AppLogger.debug(
                      '✅ LOAD SUCCESS: Current listData.length = ${listData.length}',
                    );

                    removeTooltip();
                    setState(() {
                      loadingData = false;
                    });

                    if (state.productBarcodes.isNotEmpty) {
                      AppLogger.debug('✅ LOAD SUCCESS: กำลังประมวลผลข้อมูล...');
                      // Create a list of futures
                      List<Future<void>> futures = [];

                      for (int i = 0; i < state.productBarcodes.length; i++) {
                        futures.add(
                          Future(() async {
                            // Debug log
                            AppLogger.debug(
                              '🔍 Processing item $i: ${state.productBarcodes[i].barcode}',
                            );
                            AppLogger.debug(
                              '🔍 businesstypes: ${state.productBarcodes[i].businesstypes}',
                            );
                            AppLogger.debug(
                              '🔍 ignorebranches: ${state.productBarcodes[i].ignorebranches}',
                            );

                            List<CompanyBranchModel>
                            listDataBranchInBusinessType = [];

                            /// find branch in listDataBranchAll where state.productBarcodes[i].businesstypes = listDataBranchAll[j].businesstype
                            final businesstypes =
                                state.productBarcodes[i].businesstypes ?? [];
                            for (int j = 0; j < businesstypes.length; j++) {
                              listDataBranchInBusinessType.addAll(
                                listDataBranchAll.where(
                                  (element) =>
                                      element.businesstype!.guidfixed ==
                                      businesstypes[j].guidfixed,
                                ),
                              );
                            }

                            final ignorebranches =
                                state.productBarcodes[i].ignorebranches ?? [];
                            state.productBarcodes[i].branches =
                                listDataBranchInBusinessType
                                    .where(
                                      (element) => !ignorebranches.any(
                                        (ignoreBranch) =>
                                            element.guidfixed ==
                                            ignoreBranch.guidfixed,
                                      ),
                                    )
                                    .toList();
                          }),
                        );
                      }

                      // Wait for all futures to complete
                      await Future.wait(futures);

                      setState(() {
                        listData.addAll(state.productBarcodes);
                      });

                      AppLogger.info(
                        '✅ LOAD SUCCESS: เพิ่มข้อมูลเข้า listData แล้ว (Total: ${listData.length} รายการ)',
                      );
                    } else {
                      AppLogger.warning('⚠️ LOAD SUCCESS: ไม่มีข้อมูลจาก API');
                    }

                    tooltipKeys.clear();
                    for (int i = 0; i < listData.length; i++) {
                      tooltipKeys.add(GlobalKey<ImageTooltipState>());
                    }

                    AppLogger.debug(
                      '✅ LOAD SUCCESS: สร้าง tooltipKeys เรียบร้อย (${tooltipKeys.length} รายการ)',
                    );
                  }

                  // Load from PG
                  if (state is ProductBarcodeLoadPgSuccess) {
                    removeTooltip();
                    setState(() {
                      loadingData = false;
                      _totalItems = state.total;
                    });

                    if (state.productBarcodes.isNotEmpty) {
                      setState(() {
                        listData.addAll(state.productBarcodes);
                      });
                    }

                    tooltipKeys.clear();
                    for (int i = 0; i < listData.length; i++) {
                      tooltipKeys.add(GlobalKey<ImageTooltipState>());
                    }
                  }

                  if (state is ProductBarcodeLoadPgFailed) {
                    setState(() {
                      loadingData = false;
                      global.showSnackBar(
                        context,
                        Icon(Icons.error_outline, color: global.theme.onPrimaryColor),
                        state.message,
                        global.theme.negativeHighlightTextColor,
                      );
                    });
                  }

                  if (state is ProductBarcodeLoadFailed) {
                    AppLogger.error(
                      '❌ LOAD FAILED: โหลดข้อมูลล้มเหลว - ${state.message}',
                    );

                    setState(() {
                      loadingData = false;
                      global.showSnackBar(
                        context,
                        Icon(Icons.error_outline, color: global.theme.onPrimaryColor),
                        state.message,
                        global.theme.negativeHighlightTextColor,
                      );
                    });
                  }
                  // Save
                  if (state is ProductBarcodeSaveSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.save, color: global.theme.onPrimaryColor),
                        global.language("save_success"),
                        global.theme.primaryColor,
                      );
                      clearEditData();
                      listData = [];
                    });
                    // รอ Kafka sync จาก MongoDB → PostgreSQL ก่อน refresh
                    Future.delayed(const Duration(milliseconds: 1200), () {
                      if (mounted) {
                        loadDataList(searchText, filterBarcode);
                      }
                    });
                  }
                  if (state is ProductBarcodeSaveFailed) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.save, color: global.theme.onPrimaryColor),
                        "//${global.language("not_success_save")} : ${state.message}",
                        global.theme.negativeHighlightTextColor,
                      );
                    });
                  }
                  // Update
                  if (state is ProductBarcodeUpdateSuccess) {
                    global.showSnackBar(
                      context,
                      Icon(Icons.edit, color: global.theme.onPrimaryColor),
                      global.language("edit_success"),
                      global.theme.primaryColor,
                    );
                    WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                      tabController.animateTo(0);
                    });
                    isSaveAllow = false;
                    clearEditData();
                    listData = [];
                    // รอ Kafka sync จาก MongoDB → PostgreSQL ก่อน refresh
                    Future.delayed(const Duration(milliseconds: 1200), () {
                      if (mounted) {
                        loadDataList(searchText, filterBarcode);
                      }
                    });
                  }
                  if (state is ProductBarcodeUpdateFailed) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.edit, color: global.theme.onPrimaryColor),
                        "${global.language("not_edit_success")} : ${state.message}",
                        global.theme.negativeHighlightTextColor,
                      );
                    });
                  }
                  // Delete
                  if (state is ProductBarcodeDeleteSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.delete, color: global.theme.onPrimaryColor),
                        global.language("delete_success"),
                        global.theme.primaryColor,
                      );
                      listData = [];
                      clearEditData();
                      WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                        tabController.animateTo(0);
                      });
                    });
                    // รอ Kafka sync จาก MongoDB → PostgreSQL ก่อน refresh
                    Future.delayed(const Duration(milliseconds: 1200), () {
                      if (mounted) {
                        loadDataList(searchText, filterBarcode);
                      }
                    });
                  } else if (state is ProductBarcodeDeleteFailed) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.delete, color: global.theme.onPrimaryColor),
                        "${global.language("not_delete_success")} : ${state.message}",
                        global.theme.negativeHighlightTextColor,
                      );
                    });
                  }
                  // Delete Many
                  if (state is ProductBarcodeDeleteManySuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.delete, color: global.theme.onPrimaryColor),
                        "${global.language("not_delete_success")} : ${state.message}",
                        global.theme.primaryColor,
                      );
                      listData = [];
                      clearEditData();
                      showCheckBox = false;
                    });
                    // รอ Kafka sync จาก MongoDB → PostgreSQL ก่อน refresh
                    Future.delayed(const Duration(milliseconds: 1200), () {
                      if (mounted) {
                        loadDataList(searchText, filterBarcode);
                      }
                    });
                  } else if (state is ProductBarcodeDeleteManyFailed) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.delete, color: global.theme.onPrimaryColor),
                        "${global.language("not_delete_success")} : ${state.message}",
                        global.theme.negativeHighlightTextColor,
                      );
                    });
                  }

                  // Get - In Progress
                  if (state is ProductBarcodeGetInProgress) {
                    AppLogger.debug(
                      '⏳ GET IN PROGRESS: กำลังโหลดข้อมูลสินค้า...',
                    );
                  }

                  // Get - Failed
                  if (state is ProductBarcodeGetFailed) {
                    AppLogger.error(
                      '❌ GET FAILED UI: โหลดข้อมูลสินค้าล้มเหลว - ${state.message}',
                    );
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.error_outline, color: global.theme.onPrimaryColor),
                        '❌ โหลดข้อมูลสินค้าล้มเหลว: ${state.message}',
                        global.theme.negativeHighlightTextColor,
                      );
                    });
                  }

                  // Get - Success
                  if (state is ProductBarcodeGetSuccess) {
                    AppLogger.info('✅ GET SUCCESS: โหลดข้อมูลสินค้าสำเร็จ');
                    AppLogger.debug(
                      '✅ GET SUCCESS: barcode=${state.productBarcode.barcode}',
                    );
                    AppLogger.debug(
                      '✅ GET SUCCESS: guidfixed=${state.productBarcode.guidfixed}',
                    );

                    timeForSales = [];
                    fiexdCostList = [];
                    _selectedTime = [];
                    mediaFromDateController = [];
                    mediaToDateController = [];
                    mediaFromTimeController = [];
                    mediaToTimeController = [];
                    dayOfWeekSeleted = [];
                    isShowTimeForSale = false;

                    removeTooltip();

                    setState(() {
                      isDataChange = false;

                      screenData = state.productBarcode;

                      // ป้องกัน null error
                      try {
                        barcodeController.text = screenData.barcode ?? '';
                        AppLogger.debug('✅ SET barcode: ${screenData.barcode}');
                      } catch (e) {
                        AppLogger.error('❌ ERROR setting barcode: $e');
                        barcodeController.text = '';
                      }

                      // ตรวจสอบ dimensions
                      try {
                        if (screenData.dimensions != null &&
                            screenData.dimensions!.isNotEmpty) {
                          AppLogger.debug(
                            '✅ Loading ${screenData.dimensions!.length} dimensions',
                          );
                          for (
                            int i = 0;
                            i < screenData.dimensions!.length;
                            i++
                          ) {
                            context.read<ProductDimensionBloc>().add(
                              ProductDimensionGet(
                                guid: screenData.dimensions![i].guidfixed!,
                              ),
                            );
                          }
                        }
                      } catch (e) {
                        AppLogger.error('❌ ERROR loading dimensions: $e');
                      }

                      ///ไม่ใช่ชุด
                      try {
                        if (screenData.itemtype != 2) {
                          if (screenData.refbarcodes != null &&
                              screenData.refbarcodes!.isNotEmpty) {
                            AppLogger.debug(
                              '✅ Processing refbarcodes (itemtype: ${screenData.itemtype})',
                            );
                            if (!screenData.refbarcodes![0].condition) {
                              screenData.refbarcodes![0].qty =
                                  screenData.refbarcodes![0].standvalue;
                            } else {
                              screenData.refbarcodes![0].qty =
                                  screenData.refbarcodes![0].dividevalue;
                            }
                          }
                        }
                      } catch (e) {
                        AppLogger.error('❌ ERROR processing refbarcodes: $e');
                      }

                      // ตรวจสอบ description
                      try {
                        if (screenData.description != null &&
                            screenData.description!.isNotEmpty) {
                          descriptionTextEditController.text =
                              screenData.description!;
                          AppLogger.debug('✅ SET description');
                        }
                      } catch (e) {
                        AppLogger.error('❌ ERROR setting description: $e');
                      }

                      // ตรวจสอบ alertdescription
                      try {
                        if (screenData.alertdescription != null &&
                            screenData.alertdescription!.isNotEmpty) {
                          alertDescriptionTextEditController.text =
                              screenData.alertdescription!;
                          AppLogger.debug('✅ SET alertdescription');
                        }
                      } catch (e) {
                        AppLogger.error('❌ ERROR setting alertdescription: $e');
                      }

                      /// remove screenData.prices where keynumber not in priceList
                      try {
                        if (screenData.prices != null) {
                          AppLogger.debug(
                            '✅ Processing prices (count: ${screenData.prices!.length})',
                          );
                          screenData.prices!.removeWhere(
                            (element) => !priceList.any(
                              (price) => price.keyNumber == element.keynumber,
                            ),
                          );

                          for (
                            int priceIndex = 0;
                            priceIndex < priceList.length;
                            priceIndex++
                          ) {
                            PriceDataModel priceData = screenData.prices!
                                .firstWhere(
                                  (element) =>
                                      element.keynumber ==
                                      priceList[priceIndex].keyNumber,
                                  orElse: () =>
                                      PriceDataModel(keynumber: -1, price: 0),
                                );
                            if (priceData.keynumber < 0) {
                              screenData.prices!.add(
                                PriceDataModel(
                                  keynumber: priceList[priceIndex].keyNumber,
                                  price: 0,
                                ),
                              );
                            }
                          }
                          AppLogger.debug('✅ Prices processed successfully');
                        } else {
                          AppLogger.warning('⚠️ prices is null');
                        }
                      } catch (e) {
                        AppLogger.error('❌ ERROR processing prices: $e');
                      }

                      // ตรวจสอบ color
                      try {
                        if (screenData.colorselecthex != null &&
                            screenData.colorselecthex!.isNotEmpty) {
                          colorSelected = global.colorFromHex(
                            screenData.colorselecthex!.replaceAll("#", ""),
                          );
                          AppLogger.debug(
                            '✅ SET color: ${screenData.colorselecthex}',
                          );
                        }
                      } catch (e) {
                        AppLogger.error('❌ ERROR setting color: $e');
                      }

                      /// load สาขาตามประเภทธุรกิจ
                      try {
                        if (screenData.businesstypes != null) {
                          AppLogger.debug(
                            '✅ Loading branch list (businesstypes: ${screenData.businesstypes!.length})',
                          );
                          loadDataBranchList(screenData.businesstypes!);
                        } else {
                          AppLogger.warning('⚠️ businesstypes is null');
                        }
                      } catch (e) {
                        AppLogger.error('❌ ERROR loading branch list: $e');
                      }

                      // ตรวจสอบ bom
                      try {
                        if (screenData.bom != null &&
                            screenData.bom!.isNotEmpty) {
                          AppLogger.debug('✅ Loading BOM data');
                          context.read<ProductBarcodeBloc>().add(
                            ProductBarcodeGetBom(
                              barcode: screenData.barcode ?? '',
                            ),
                          );
                        } else {
                          AppLogger.debug('✅ No BOM data');
                          productBom = ProductBomModel(
                            guidfixed: '',
                            names: [],
                            itemunitcode: '',
                            itemunitnames: [],
                            barcode: '',
                            condition: false,
                            dividevalue: 0,
                            standvalue: 0,
                            qty: 0,
                            imageuri: '',
                            bom: [],
                          );
                        }
                      } catch (e) {
                        AppLogger.error('❌ ERROR loading BOM: $e');
                      }

                      if (screenData.timeforsales!.isNotEmpty) {
                        isShowTimeForSale = true;
                        for (
                          int i = 0;
                          i < screenData.timeforsales!.length;
                          i++
                        ) {
                          timeForSales.add(screenData.timeforsales![i]);
                          _selectedTime.add(TimeOfDay.now());
                          mediaFromDateController.add(TextEditingController());
                          mediaToDateController.add(TextEditingController());
                          mediaFromTimeController.add(TextEditingController());
                          mediaToTimeController.add(TextEditingController());

                          mediaFromDateController[i].text =
                              (screenData.timeforsales![i].fromdate!.isNotEmpty)
                              ? DateFormat('dd/MM/yyyy').format(
                                  DateTime.parse(
                                    screenData.timeforsales![i].fromdate!,
                                  ),
                                )
                              : '';
                          mediaToDateController[i].text =
                              (screenData.timeforsales![i].todate!.isNotEmpty)
                              ? DateFormat('dd/MM/yyyy').format(
                                  DateTime.parse(
                                    screenData.timeforsales![i].todate!,
                                  ),
                                )
                              : '';
                          mediaFromTimeController[i].text =
                              screenData.timeforsales![i].fromtime!;
                          mediaToTimeController[i].text =
                              screenData.timeforsales![i].totime!;
                          dayOfWeekSeleted.add(
                            screenData.timeforsales![i].daysofweek!
                                .map(
                                  (e) => dayOfWeekList.firstWhere(
                                    (element) => element.code == e.toString(),
                                  ),
                                )
                                .toList(),
                          );
                        }
                      }

                      if (screenData.fixedcost!.isNotEmpty) {
                        for (int i = 0; i < screenData.fixedcost!.length; i++) {
                          fiexdCostList.add(screenData.fixedcost![i]);
                          asDateController.add(TextEditingController());
                          fiexdCostAmountController.add(
                            TextEditingController(),
                          );

                          asDateController[i].text =
                              (screenData.fixedcost![i].effectdate!.isNotEmpty)
                              ? DateFormat('dd/MM/yyyy').format(
                                  DateTime.parse(
                                    screenData.fixedcost![i].effectdate!,
                                  ),
                                )
                              : '';
                          fiexdCostAmountController[i].text = global
                              .formatNumber(screenData.fixedcost![i].amount!);
                        }
                      }

                      if (isEditMode) {
                        WidgetsBinding.instance.addPostFrameCallback((
                          timeStamp,
                        ) {
                          tabController.animateTo(1);
                        });
                        WidgetsBinding.instance.addPostFrameCallback((_) {
                          focusFirstTextField(context);
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

                  // Get Bom
                  if (state is ProductBarcodeGetBomInProgress) {
                    context.loaderOverlay.show(
                      widgetBuilder: (progress) {
                        return const ReconnectingOverlay();
                      },
                    );
                  } else if (state is ProductBarcodeGetBomSuccess) {
                    context.loaderOverlay.hide();
                    setState(() {
                      productBom = state.productBom;
                      productBom.qty = 1;
                    });
                  } else if (state is ProductBarcodeGetBomFailed) {
                    context.loaderOverlay.hide();
                    global.showSnackBar(
                      context,
                      Icon(Icons.error_outline, color: global.theme.onPrimaryColor),
                      state.message,
                      global.theme.negativeHighlightTextColor,
                    );
                  }
                },
              ),
              BlocListener<CompanyBranchBloc, CompanyBranchState>(
                listener: (context, state) {
                  if (state is CompanyBranchLoadSuccess) {
                    setState(() {
                      listDataBranchAll = state.companyBranch;
                    });
                  }
                  if (state is CompanyBranchByBusinessTypeLoadSuccess) {
                    // Add elements to listDataBranchBusinessType only if guidfixed is not empty
                    listDataBranchBusinessType.addAll(
                      state.companyBranch
                          .where((element) => element.guidfixed.isNotEmpty)
                          .map(
                            (element) => BranchModel(
                              guidfixed: element.guidfixed,
                              code: element.code,
                              names: element.names,
                              isignore: false,
                            ),
                          ),
                    );

                    setState(() {
                      ignorebranches = [];
                      if (screenData.ignorebranches!.isNotEmpty) {
                        for (var element in listDataBranchBusinessType) {
                          if (screenData.ignorebranches!.any(
                            (ignoreBranch) => ignoreBranch.code != element.code,
                          )) {
                            ignorebranches!.add(element);
                          }
                        }
                      } else {
                        ignorebranches!.addAll(listDataBranchBusinessType);
                      }
                    });
                  }
                },
              ),
              BlocListener<ExportCsvBloc, ExportCsvState>(
                listener: (context, state) {
                  /// download csv
                  if (state is ProductBarcodeExportInProgress) {
                    // Show a loading indicator if desired
                  } else if (state is ProductBarcodeExportSuccess) {
                    global.showSnackBar(
                      context,
                      Icon(Icons.save, color: global.theme.onPrimaryColor),
                      global.language("export_success"),
                      global.theme.primaryColor,
                    );
                  } else if (state is ProductBarcodeExportFailed) {
                    // Show an error message
                    global.showSnackBar(
                      context,
                      Icon(Icons.save, color: global.theme.onPrimaryColor),
                      "${global.language("not_export_success")} : ${state.message}",
                      global.theme.negativeHighlightTextColor,
                    );
                  }
                },
              ),
              BlocListener<UnitBloc, UnitState>(
                listener: (context, state) {
                  if (state is UnitLoadSuccess) {
                    setState(() {
                      if (state.units.isNotEmpty) {
                        unitListData.addAll(state.units);
                      }
                    });
                  }
                },
              ),

              BlocListener<SaleChannelBloc, SaleChannelState>(
                listener: (context, state) {
                  if (state is SaleChannelLoadSuccess) {
                    setState(() {
                      if (state.salechannel.isNotEmpty) {
                        saleChannelList.addAll(state.salechannel);

                        /// ดึงชื่อจาก master ที่กำหนดใช้ราคา ไหน ตาม keyNumber
                        for (int i = 0; i < saleChannelList.length; i++) {
                          var priceElement = priceList.firstWhere(
                            (element) =>
                                element.keyNumber == saleChannelList[i].price,
                            orElse: () => PriceModel(
                              keyNumber: -1,
                              names: [],
                              isUse: false,
                            ),
                          );
                          priceElement.names
                                  .firstWhere(
                                    (element) => element.code == 'th',
                                    orElse: () =>
                                        LanguageModel(code: 'th', name: ''),
                                  )
                                  .name =
                              saleChannelList[i].name!;
                        }
                      }
                    });
                  }
                },
              ),

              BlocListener<ImageUploadBloc, ImageUploadState>(
                listener: (context, state) {
                  if (state is ImageUploadSaveSuccess) {
                    setState(() {
                      screenData
                              .options![uploadImageOptionIndex]
                              .choices[uploadImageChoiceIndex]
                              .imageuri =
                          state.imageUpload.uri;
                    });
                  }
                },
              ),

              BlocListener<ProductDimensionBloc, ProductDimensionState>(
                listener: (context, state) {
                  if (state is ProductDimensionGetSuccess) {
                    setState(() {
                      for (int i = 0; i < screenData.dimensions!.length; i++) {
                        dimensionSeleted.add(
                          DimensionModel(
                            guidfixed: screenData.dimensions![i].guidfixed,
                            names: screenData.dimensions![i].names,
                            isdisabled: screenData.dimensions![i].isdisabled,
                            items: state.productDimension.items,
                          ),
                        );
                      }
                    });
                  }
                },
              ),

              /// BusinessTypeBloc
              BlocListener<BusinessTypeBloc, BusinessTypeState>(
                listener: (context, state) {
                  if (state is BusinessTypeLoadSuccess) {
                    setState(() {
                      if (state.businessType.isNotEmpty) {
                        listDataBusinessType.addAll(state.businessType);
                      }
                    });
                  }
                },
              ),
            ],
            child: (constraints.maxWidth > 800)
                ? SplitView(
                    gripSize: 8,
                    gripColor: global.theme.dividerBorderColor,
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

  List<Widget> listNamesFields(
    List<LanguageDataModel> names,
    String fieldname,
  ) {
    List<Widget> forms = [];
    for (
      int languageIndex = 0;
      languageIndex < languageList.length;
      languageIndex++
    ) {
      LanguageDataModel nameObj = names.firstWhere(
        (element) => element.code == languageList[languageIndex].code,
        orElse: () => LanguageDataModel(code: '', name: ''),
      );
      if (nameObj.code == '') {
        names.add(
          LanguageDataModel(code: languageList[languageIndex].code!, name: ''),
        );
      }
    }
    for (
      int languageIndex = 0;
      languageIndex < languageList.length;
      languageIndex++
    ) {
      LanguageDataModel nameObj = names.firstWhere(
        (element) => element.code == languageList[languageIndex].code,
        orElse: () => LanguageDataModel(code: '', name: ''),
      );
      if (nameObj.code != '') {
        forms.add(
          Padding(
            padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
            child: TextFormField(
              readOnly: !isEditMode,
              onChanged: (value) {
                isDataChange = true;
                nameObj.name = value;
              },
              textAlign: TextAlign.left,
              controller: TextEditingController(text: nameObj.name),
              decoration: InputDecoration(
                floatingLabelBehavior: FloatingLabelBehavior.always,
                border: const OutlineInputBorder(),
                labelText:
                    "${global.language(fieldname)} (${getLangName(nameObj.code)})",
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
                    return 'This field is required';
                  }
                }

                return null;
              },
            ),
          ),
        );
      }
    }

    return forms;
  }
}

/// Formatter ที่แก้ปัญหา Flutter Web ลบ Thai grapheme cluster ทั้งก้อน
/// เช่น "สี" กด backspace → ลบเฉพาะ สระอี (ี) เหลือ "ส"
class _ThaiCodePointBackspaceFormatter extends TextInputFormatter {
  @override
  TextEditingValue formatEditUpdate(
    TextEditingValue oldValue,
    TextEditingValue newValue,
  ) {
    // ถ้าไม่ใช่ deletion → ปล่อยผ่าน
    if (newValue.text.length >= oldValue.text.length) return newValue;

    // ถ้า selection ไม่ใช่ collapsed (ลบ range) → ปล่อยผ่าน
    if (!oldValue.selection.isCollapsed || !newValue.selection.isCollapsed) {
      return newValue;
    }

    final oldCursor = oldValue.selection.baseOffset;
    final newCursor = newValue.selection.baseOffset;
    final deletedCount = oldCursor - newCursor;

    // ถ้าลบแค่ 1 code unit → ปกติ
    if (deletedCount <= 1) return newValue;

    // ลบหลาย code units (grapheme cluster) → ลบเฉพาะ 1 ตัวสุดท้าย
    final correctedText =
        oldValue.text.substring(0, oldCursor - 1) +
        oldValue.text.substring(oldCursor);
    return TextEditingValue(
      text: correctedText,
      selection: TextSelection.collapsed(offset: oldCursor - 1),
    );
  }
}
