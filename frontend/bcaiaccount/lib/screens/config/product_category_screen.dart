import 'dart:async';
import 'dart:io';
import 'package:smlaicloud/utils/focus_utils.dart';
import 'package:smlaicloud/bloc/image/image_upload_bloc.dart';
import 'package:smlaicloud/widgets/manual_button.dart';
import 'package:smlaicloud/widgets/list_font_size_control.dart';
import 'package:smlaicloud/widgets/edit_font_size_control.dart';
import 'package:smlaicloud/bloc/product_category/product_category_bloc.dart';
import 'package:smlaicloud/model/product_category_model.dart';
import 'package:flex_color_picker/flex_color_picker.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:split_view/split_view.dart';
import 'package:smlaicloud/model/product_category_list_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:translator/translator.dart';
import 'package:flutter/services.dart';
import 'package:image_picker/image_picker.dart';
import 'package:flutter_dropzone/flutter_dropzone.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:intl/intl.dart';
import 'package:smlaicloud/utils/date_picker.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class ProductCategoryScreen extends StatefulWidget {
  final int groupnumber;
  const ProductCategoryScreen({super.key, required this.groupnumber});

  @override
  State<ProductCategoryScreen> createState() => ProductCategoryScreenState();
}

class ProductCategoryScreenState extends State<ProductCategoryScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  final translator = GoogleTranslator();
  late DropzoneViewController dropZoneController;
  final ImagePicker _picker = ImagePicker();
  File imageFile = File('');
  List<ProductCategoryListModel> rootCategorys = [];
  List<Widget> listColumns = [];
  TextEditingController searchController = TextEditingController();
  ScrollController listScrollController = ScrollController();
  ScrollController editScrollController = ScrollController();
  bool needScroll = false;
  String selectGuid = "";
  String selectParentGuid = "";
  String selectParentName = "";
  String selectDragTargetGuid = "";
  late TabController tabController;
  List<String> fieldName = [];
  Uint8List? imageWeb;
  List<LanguageModel> languageList = <LanguageModel>[];
  List<TextEditingController> fieldTextController = [];
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  List<FocusNode> fieldFocusNodes = [];
  int focusNodeIndex = 0;
  late ProductCategoryState blocCategoryState;
  bool isEditMode = false;
  bool isChange = false;
  bool isSaveAllow = false;
  bool isDeleteAllow = false;
  String headerEdit = "";
  String selectImageUri = "";
  int xorder = 0;
  List<String> xorderUpdateList = [];
  late Timer timerForUpdateXOrder;
  Color colorSelected = Colors.white;
  String colorSelectedHex = "";
  bool useImageOrColor = false;
  bool isdisabled = false;
  final debouncer = global.Debouncer(1000);
  late SplitViewController splitViewController;
  List<ProductCategoryCodeListModel> codeList = [];
  ProductCategoryModel dataTemp = ProductCategoryModel(
    guidfixed: '',
    parentguid: '',
    parentguidall: '',
    names: [],
    imageuri: '',
    childcount: 0,
    xsorts: [],
    useimageorcolor: false,
    isdisabled: false,
    codelist: [],
    colorselect: '',
    colorselecthex: '',
    timeforsales: [],
  );

  List<File> imageFileCover = [File('')];
  List<Uint8List> imageWebCover = [Uint8List(0)];
  String selectImageUriCover = "";
  TextEditingController groupNumber = TextEditingController();
  bool isLoadTranslation = false;

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

  void updateXOrder() {
    List<XSortModel> updateList = [];
    List<String> distinctXOrder = xorderUpdateList.toSet().toList();
    for (int index = 0; index < distinctXOrder.length; index++) {
      String guid = xorderUpdateList[index];
      ProductCategoryListModel? getCategory = findByGuid(rootCategorys, guid);
      if (getCategory != null) {
        updateList.add(
          XSortModel(
            guidfixed: getCategory.detail.guidfixed,
            xorder: getCategory.detail.xsorts![0].xorder,
            code: "X",
          ),
        );
      }
    }
    if (updateList.isNotEmpty) {
      context.read<ProductCategoryBloc>().add(
        ProductCategoryUpdateXOrder(orderLists: updateList),
      );
    }
    xorderUpdateList.clear();
  }

  String packName(List<LanguageDataModel> names) {
    String result = "";
    for (int i = 0; i < names.length; i++) {
      result += names[i].name;
      if (i < names.length - 1) {
        result += ",";
      }
    }
    return result;
  }

  List<LanguageDataModel> packLanguage() {
    List<LanguageDataModel> names = [];
    for (int i = 0; i < languageList.length; i++) {
      if (languageList[i].code!.trim().isNotEmpty) {
        names.add(
          LanguageDataModel(
            code: languageList[i].code!,
            name: fieldTextController[i].text,
          ),
        );
      }
    }

    for (var defaultValueLang in dataTemp.names!) {
      LanguageDataModel result = names.firstWhere(
        (data) => data.code == defaultValueLang.code,
        orElse: () => LanguageDataModel(code: '', name: ''),
      );
      if (result.code == '') {
        names.add(defaultValueLang);
      }
    }

    return names;
  }

  void loadDataList(String search) {
    context.read<ProductCategoryBloc>().add(
      ProductCategoryLoadList(
        offset: 0,
        limit: 100000,
        search: search,
        groupNumber: widget.groupnumber,
      ),
    );
  }

  ProductCategoryListModel? findByGuid(
    List<ProductCategoryListModel> categorys,
    String guid,
  ) {
    for (int findIndex = 0; findIndex < categorys.length; findIndex++) {
      if (guid == categorys[findIndex].detail.guidfixed) {
        return categorys[findIndex];
      }
      if (categorys[findIndex].childCategories.isNotEmpty) {
        ProductCategoryListModel? find = findByGuid(
          categorys[findIndex].childCategories,
          guid,
        );
        if (find != null) {
          return find;
        }
      }
    }
    return null;
  }

  String categoryFullName(
    List<ProductCategoryListModel> categorys,
    String guid, {
    String result = "",
  }) {
    for (int findIndex = 0; findIndex < categorys.length; findIndex++) {
      if (guid == categorys[findIndex].detail.guidfixed) {
        return result + categorys[findIndex].detail.names![0].name;
      }
      if (categorys[findIndex].childCategories.isNotEmpty) {
        String find = categoryFullName(
          categorys[findIndex].childCategories,
          guid,
          result: result,
        );
        if (find.isNotEmpty) {
          return "$result${categorys[findIndex].detail.names![0].name},$find";
        }
      }
    }
    return result;
  }

  String selectParentGuidAll(
    List<ProductCategoryListModel> categorys,
    String guid,
  ) {
    String result = "";
    for (int findIndex = 0; findIndex < categorys.length; findIndex++) {
      if (guid == categorys[findIndex].detail.guidfixed) {
        return result + categorys[findIndex].detail.guidfixed;
      }
      if (categorys[findIndex].childCategories.isNotEmpty) {
        String find = selectParentGuidAll(
          categorys[findIndex].childCategories,
          guid,
        );
        if (find.isNotEmpty) {
          return "$result${categorys[findIndex].detail.guidfixed},$find";
        }
      }
    }
    return result;
  }

  void saveOrUpdateData() {
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

    if (selectGuid.trim().isEmpty) {
      ProductCategoryModel categoryData = ProductCategoryModel(
        guidfixed: "",
        parentguid: selectParentGuid,
        parentguidall: selectParentGuidAll(rootCategorys, selectParentGuid),
        useimageorcolor: useImageOrColor,
        isdisabled: isdisabled,
        colorselect: colorSelected.value.toString(),
        colorselecthex: colorSelectedHex,
        imageuri: "",
        childcount: 0,
        names: packLanguage(),
        xsorts: [SortDataModel(code: "X", xorder: xorder)],
        codelist: [],
        coveruri: selectImageUriCover,
        groupnumber: widget.groupnumber,
        timeforsales: timeForSales,
      );
      if (imageFile.path.isNotEmpty) {
        context.read<ProductCategoryBloc>().add(
          ProductCategoryWithImageSave(
            category: categoryData,
            imageFile: imageFile,
            imageWeb: imageWeb,
          ),
        );
      } else {
        context.read<ProductCategoryBloc>().add(
          ProductCategorySave(category: categoryData),
        );
      }
    } else {
      updateData(selectGuid);
    }
  }

  void updateData(String guid) {
    ProductCategoryModel categoryModel = ProductCategoryModel(
      guidfixed: guid,
      parentguid: selectParentGuid,
      parentguidall: selectParentGuidAll(rootCategorys, selectParentGuid),
      imageuri: selectImageUri,
      childcount: 0,
      useimageorcolor: useImageOrColor,
      isdisabled: isdisabled,
      colorselect: colorSelected.value.toString(),
      colorselecthex: colorSelectedHex,
      names: packLanguage(),
      xsorts: [SortDataModel(code: "X", xorder: xorder)],
      codelist: codeList,
      coveruri: selectImageUriCover,
      timeforsales: timeForSales,
    );
    if (imageWeb != null) {
      context.read<ProductCategoryBloc>().add(
        ProductCategoryWithImageUpdate(
          guid: guid,
          category: categoryModel,
          imageFile: imageFile,
          imageWeb: imageWeb!,
        ),
      );
    } else {
      context.read<ProductCategoryBloc>().add(
        ProductCategoryUpdate(guid: guid, category: categoryModel),
      );
    }
  }

  void buildColumnWidget() {
    listColumns.clear();
    categoryList(0, rootCategorys);
  }

  void categoryListScrollToEnd() async {
    listScrollController.animateTo(
      listScrollController.position.maxScrollExtent,
      duration: const Duration(milliseconds: 300),
      curve: Curves.easeOut,
    );
  }

  Widget categoryDetail(
    int level,
    ProductCategoryListModel category,
    int index,
    int count, {
    required Function moveUpCallBack,
    required Function moveDownCallBack,
  }) {
    return DragTarget(
      onWillAcceptWithDetails: (data) {
        setState(() {
          selectDragTargetGuid = category.detail.guidfixed;
          buildColumnWidget();
        });
        return !category.detail.parentguidall.contains(selectGuid);
        //return (activeSelectedGuid != category.detail.parentGuid);
      },
      onAcceptWithDetails: (data) {
        setState(() {
          if (selectGuid != selectDragTargetGuid) {
            var findCategory = findByGuid(rootCategorys, selectGuid);
            if (findCategory != null) {
              selectParentGuid = selectDragTargetGuid;
              updateData(selectGuid);
            }
          }
        });
      },
      onLeave: (data) {
        setState(() {
          //selectDragTargetGuid = "";
          //buildColumnWidget();
        });
      },
      builder: (context, candidateData, rejectedData) {
        Color color = global.theme.cardColor;
        if (selectDragTargetGuid == category.detail.guidfixed) {
          color = global.theme.positiveHighlightColor;
        }
        if (selectGuid == category.detail.guidfixed) {
          color = global.theme.infoHighlightColor;
        }
        if (selectDragTargetGuid.isNotEmpty &&
            category.detail.parentguidall.contains(selectGuid)) {
          color = global.theme.negativeHighlightColor;
        }

        if (category.detail.xsorts!.isEmpty) {
          category.detail.xsorts!.add(
            SortDataModel(code: "X", xorder: index + 1),
          );
        }
        if (category.detail.xsorts![0].xorder != index + 1) {
          category.detail.xsorts![0].xorder = index + 1;
          xorderUpdateList.add(category.detail.guidfixed);
        }

        return Container(
          color: color,
          padding: const EdgeInsets.all(4),
          width: double.infinity,
          child: Row(
            children: <Widget>[
              SizedBox(width: level * 20),
              Expanded(
                child: Text(
                  "${category.detail.xsorts![0].xorder} ${global.packName(category.detail.names!)}",
                  style: TextStyle(fontSize: 18, color: global.theme.textColor),
                ),
              ),
              if (category.childCategories.isNotEmpty)
                IconButton(
                  padding: EdgeInsets.zero,
                  color: global.theme.positiveHighlightTextColor,
                  focusNode: FocusNode(skipTraversal: true),
                  icon: Icon(
                    (category.isExpand) ? Icons.expand_less : Icons.expand_more,
                  ),
                  onPressed: () {
                    setState(() {
                      selectDragTargetGuid = "";
                      category.isExpand = !category.isExpand;
                      buildColumnWidget();
                    });
                  },
                ),
              if (index > 0)
                IconButton(
                  padding: EdgeInsets.zero,
                  color: (selectGuid == category.detail.guidfixed)
                      ? global.theme.iconColor
                      : global.theme.primaryColor,
                  focusNode: FocusNode(skipTraversal: true),
                  icon: const Icon(Icons.move_up),
                  onPressed: () {
                    setState(() {
                      moveUpCallBack();
                      buildColumnWidget();
                    });
                  },
                ),
              if (index >= 0 && index < count - 1)
                IconButton(
                  padding: EdgeInsets.zero,
                  color: (selectGuid == category.detail.guidfixed)
                      ? global.theme.iconColor
                      : global.theme.negativeHighlightTextColor,
                  focusNode: FocusNode(skipTraversal: true),
                  icon: const Icon(Icons.move_down),
                  onPressed: () {
                    setState(() {
                      moveDownCallBack();
                      buildColumnWidget();
                    });
                  },
                ),
            ],
          ),
        );
      },
    );
  }

  void updateXorderAll(List<ProductCategoryListModel> categorys) {
    // Update Xorder ทุกตัว
    for (var index = 0; index < categorys.length; index++) {
      xorderUpdateList.add(categorys[index].detail.guidfixed);
    }
  }

  Widget categoryList(int level, List<ProductCategoryListModel> categorys) {
    for (var index = 0; index < categorys.length; index++) {
      Widget detail = categoryDetail(
        level,
        categorys[index],
        index,
        categorys.length,
        moveUpCallBack: () {
          categorys.insert(index, categorys.removeAt(index - 1));
          updateXorderAll(categorys);
          buildColumnWidget();
        },
        moveDownCallBack: () {
          categorys.insert(index, categorys.removeAt(index + 1));
          updateXorderAll(categorys);
          buildColumnWidget();
        },
      );
      listColumns.add(
        GestureDetector(
          onTap: () {
            discardData(
              callBack: () {
                setState(() {
                  AppLogger.debug("onTab");
                  isSaveAllow = false;
                  isEditMode = false;
                  selectDragTargetGuid = "";
                  selectGuid = categorys[index].detail.guidfixed;
                  selectParentName = categorys[index].detail.names![0].name;
                  selectImageUri = categorys[index].detail.imageuri;
                  headerEdit = global.language("show");
                  isDeleteAllow = categorys[index].childCategories.isEmpty;
                  context.read<ProductCategoryBloc>().add(
                    ProductCategoryGet(guid: selectGuid),
                  );
                  buildColumnWidget();
                });
              },
            );
          },
          onDoubleTap: () {
            discardData(
              callBack: () {
                setState(() {
                  isSaveAllow = true;
                  isEditMode = true;
                  selectDragTargetGuid = "";
                  selectGuid = categorys[index].detail.guidfixed;
                  selectParentName = categorys[index].detail.names![0].name;
                  headerEdit = global.language("edit");
                  isDeleteAllow = categorys[index].childCategories.isEmpty;
                  switchToEdit(selectGuid);
                  buildColumnWidget();
                  WidgetsBinding.instance.addPostFrameCallback((_) {
                    focusFirstTextField(context);
                  });
                });
              },
            );
          },
          child: LongPressDraggable<Widget>(
            data: detail,
            onDragStarted: () {
              setState(() {
                selectGuid = categorys[index].detail.guidfixed;
                context.read<ProductCategoryBloc>().add(
                  ProductCategoryGet(guid: selectGuid),
                );
                buildColumnWidget();
              });
            },
            dragAnchorStrategy: pointerDragAnchorStrategy,
            axis: Axis.vertical,
            feedback: SizedBox(
              width: 100,
              height: 100,
              child: Text(
                global.packName(categorys[index].detail.names!),
                style: const TextStyle(fontSize: 18),
              ),
            ),
            child: detail,
          ),
        ),
      );
      if (categorys[index].childCategories.isNotEmpty &&
          categorys[index].isExpand) {
        categoryList(level + 1, categorys[index].childCategories);
      }
    }

    return Column(children: listColumns);
  }

  void setSystemLanguageList() async {
    await global.setSystemLanguage(context);

    for (int i = 0; i < global.config.languages.length; i++) {
      if (global.config.languages[i].isuse!) {
        languageList.add(global.config.languages[i]);
      }
    }
    for (int i = 0; i < languageList.length; i++) {
      fieldTextController.add(TextEditingController());
      FocusNode focusNode = FocusNode();
      focusNode.addListener(() {
        focusNodeIndex = i;
      });
      fieldFocusNodes.add(focusNode);
    }
    loadDataList("");
  }

  @override
  void initState() {
    tabController = TabController(vsync: this, length: 2);

    // เรียงลำดับ Focus
    fieldTextController.add(TextEditingController());

    setSystemLanguageList();
    //listScrollController.addListener(onScrollList);

    timerForUpdateXOrder = Timer.periodic(
      const Duration(seconds: 1),
      (Timer t) => updateXOrder(),
    );

    searchFocusNode.requestFocus();
    splitViewController = SplitViewController(
      limits: [null, WeightLimit(min: 0.2, max: 0.8)],
    );
    super.initState();
  }

  @override
  void dispose() {
    listScrollController.dispose();
    tabController.dispose();
    editScrollController.dispose();
    for (int i = 0; i < fieldTextController.length; i++) {
      fieldTextController[i].dispose();
    }
    for (int i = 0; i < fieldFocusNodes.length; i++) {
      fieldFocusNodes[i].dispose();
    }
    timerForUpdateXOrder.cancel();
    groupNumber.dispose();
    super.dispose();
  }

  void deleteDialog() {
    showDialog<String>(
      context: context,
      builder: (BuildContext context) => AlertDialog(
        title: Text(global.language('delete_confirm')),
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
              context.read<ProductCategoryBloc>().add(
                ProductCategoryDelete(guid: selectGuid),
              );
            },
            child: Text(global.language('confirm')),
          ),
        ],
      ),
    );
  }

  Widget listScreen({required bool mobileScreen}) {
    if (needScroll) {
      categoryListScrollToEnd();
      needScroll = false;
    }
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('product_category')),
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
          const ManualButton(path: 'product-category'),
          Padding(
            padding: EdgeInsets.only(right: 20.0),
            child: IconButton(
              focusNode: FocusNode(skipTraversal: true),
              onPressed: () {
                discardData(
                  callBack: () {
                    setState(() {
                      isEditMode = true;
                      isChange = false;
                      isSaveAllow = true;
                      selectGuid = "";
                      headerEdit = global.language('append');
                      clearEditData();
                      if (mobileScreen && selectDragTargetGuid.isEmpty) {
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
              icon: Icon(Icons.add, size: 26.0),
            ),
          ),
        ],
      ),
      body: Column(
        children: [
          Container(
            padding: EdgeInsets.symmetric(horizontal: 8, vertical: 4),
            color: global.theme.searchBarColor,
            child: Row(
              children: [
                Expanded(
                  child: TextField(
                    onSubmitted: (value) {
                      searchFocusNode.requestFocus();
                    },
                    onChanged: (value) {
                      debouncer.run(() {
                        loadDataList(value);
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
                      prefixIcon: Icon(Icons.search, size: 20),
                      prefixIconConstraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                    ),
                  ),
                ),
                ListFontSizeControl(onChanged: () => setState(() {})),
                const SizedBox(width: 4),
                if (rootCategorys.isNotEmpty)
                  Text('(${rootCategorys.length})', style: TextStyle(fontSize: 11, color: global.theme.textSecondaryColor)),
              ],
            ),
          ),
          Container(color: global.theme.appBarColor, height: 2),
          if (selectDragTargetGuid.isNotEmpty)
            DragTarget(
              builder: (context, candidateData, rejectedData) {
                return Container(
                  width: double.infinity,
                  height: 40,
                  color: global.theme.positiveHighlightColor,
                  child: Icon(
                    Icons.home,
                    color: global.theme.positiveHighlightTextColor,
                    size: 26.0,
                  ),
                );
              },
              onWillAcceptWithDetails: (data) {
                return true;
              },
              onAcceptWithDetails: (data) {
                setState(() {
                  selectParentGuid = "";
                  updateData(selectGuid);
                });
              },
            ),
          Expanded(
            child: SingleChildScrollView(
              controller: listScrollController,
              child: Column(children: listColumns),
            ),
          ),
        ],
      ),
    );
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

  void switchToEdit(String guid) {
    setState(() {
      selectGuid = guid;
      context.read<ProductCategoryBloc>().add(
        ProductCategoryGet(guid: selectGuid),
      );
      headerEdit = global.language("edit");
      isSaveAllow = true;
      isEditMode = true;
    });
  }

  void upLoadImage() {
    if (imageFileCover.isNotEmpty) {
      if (imageFileCover[0].path != '') {
        context.read<ImageUploadBloc>().add(
          ImageUploadFileSaved(
            imageFiles: imageFileCover,
            imageWeb: imageWebCover,
          ),
        );
      }
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
            Icon(Icons.edit, color: Colors.white),
            global.language("to_time_must_be_later_than_from_time"),
            Colors.red,
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

  Widget editScreen({mobileScreen = true}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      appBar: AppBar(
        automaticallyImplyLeading: false,
        title: Text(headerEdit + global.language('product_category')),
        backgroundColor: (isEditMode)
            ? global.theme.toolBarEditModeColor
            : global.theme.appBarColor,
        leading: mobileScreen
            ? IconButton(
                focusNode: FocusNode(skipTraversal: true),
                icon: const Icon(Icons.arrow_back),
                onPressed: () {
                  tabController.index = 0;
                },
              )
            : null,
        actions: <Widget>[
          EditFontSizeControl(onChanged: () => setState(() {})),
          const SizedBox(width: 8),
          if (isDeleteAllow && selectGuid.isNotEmpty)
            Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () {
                  deleteDialog();
                  setState(() {});
                },
                icon: const Icon(Icons.delete),
              ),
            ),
          if (global.systemLanguage.length > 1 && isSaveAllow)
            Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () async {
                  setState(() {
                    isLoadTranslation = true;
                  });
                  for (int i = 0; i < languageList.length; i++) {
                    try {
                      var translation = await translator.translate(
                        fieldTextController[0].text,
                        to: languageList[i].codeTranslator!,
                      );
                      if (fieldTextController[i].text.isEmpty) {
                        fieldTextController[i].text = translation.text;
                      }
                    } catch (e) {
                      AppLogger.debug(e.toString());
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
                  switchToEdit(selectGuid);
                },
                icon: const Icon(Icons.edit),
              ),
            ),
          if (isSaveAllow)
            Padding(
              padding: const EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () => saveOrUpdateData(),
                icon: const Icon(Icons.save, size: 26.0),
              ),
            ),
          if (selectGuid.isNotEmpty)
            Padding(
              padding: EdgeInsets.only(right: 20.0),
              child: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                onPressed: () {
                  discardData(
                    callBack: () {
                      setState(() {
                        selectParentGuid = selectGuid;
                        selectGuid = "";
                        isSaveAllow = true;
                        isEditMode = true;
                        isChange = false;
                        headerEdit = global.language('append');
                        tabController.index = 1;
                        clearEditData();
                        WidgetsBinding.instance.addPostFrameCallback((_) {
                          focusFirstTextField(context);
                        });
                      });
                    },
                  );
                },
                icon: const Icon(Icons.add_link, size: 26.0),
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
        child: Builder(
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
            width: double.infinity,
            padding: const EdgeInsets.all(10),
            child: Column(
              children: [
                if (selectParentGuid.isNotEmpty)
                  SizedBox(
                    width: double.infinity,
                    child: Container(
                      margin: EdgeInsets.only(bottom: 5),
                      padding: EdgeInsets.all(10),
                      decoration: BoxDecoration(
                        color: global.theme.cardColor,
                        border: Border.all(color: global.theme.textSecondaryColor),
                        borderRadius: BorderRadius.circular(5),
                        boxShadow: [
                          BoxShadow(
                            color: global.theme.textSecondaryColor.withValues(alpha: 0.5),
                            spreadRadius: 2,
                            blurRadius: 2,
                            offset: const Offset(0, 1),
                          ),
                        ],
                      ),
                      child: Row(
                        crossAxisAlignment: CrossAxisAlignment.end,
                        children: [
                          Text(global.language("select_category")),
                          const SizedBox(width: 5),
                          Text(
                            categoryFullName(rootCategorys, selectParentGuid),
                          ),
                        ],
                      ),
                    ),
                  ),
                SizedBox(height: 10),
                TextFormField(
                  enabled: false,
                  controller: groupNumber,
                  decoration: InputDecoration(
                    floatingLabelBehavior: FloatingLabelBehavior.always,
                    border: OutlineInputBorder(),
                    labelText: global.language("category_group_number"),
                  ),
                ),
                const SizedBox(height: 10),
                for (int i = 0; i < languageList.length; i++)
                  Padding(
                    padding: EdgeInsets.only(bottom: 10),
                    child: TextFormField(
                      onChanged: (value) {
                        isChange = true;
                      },
                      onFieldSubmitted: (value) {
                        findFocusNext(focusNodeIndex);
                      },
                      textInputAction: TextInputAction.next,
                      focusNode: fieldFocusNodes[i],
                      textAlign: TextAlign.left,
                      controller: fieldTextController[i],
                      decoration: InputDecoration(
                        floatingLabelBehavior: FloatingLabelBehavior.always,
                        border: const OutlineInputBorder(),
                        labelText:
                            "${global.language("product_category_name")} (${languageList[i].name})",
                        suffixIcon: isLoadTranslation
                            ? const Padding(
                                padding: EdgeInsets.all(8.0),
                                child: SizedBox(
                                  width: 20,
                                  height: 20,
                                  child: CircularProgressIndicator(
                                    strokeWidth: 2,
                                  ),
                                ),
                              )
                            : null,
                      ),
                    ),
                  ),

                Row(
                  crossAxisAlignment: CrossAxisAlignment.center,
                  children: [
                    Text(
                      "${global.language("show_time_for_sale")} ${(timeForSales.isNotEmpty) ? "(${timeForSales.length})" : ""}",
                      style: TextStyle(
                        fontSize: 16,
                        color: (!isShowTimeForSale)
                            ? global.theme.textSecondaryColor
                            : global.theme.textColor,
                      ),
                    ),
                    Switch(
                      value: isShowTimeForSale,
                      onChanged: (value) {
                        setState(() {
                          isShowTimeForSale = value;
                        });
                      },
                    ),
                  ],
                ),

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

                                    mediaFromDateController.removeAt(
                                      mediaIndex,
                                    );
                                    mediaToDateController.removeAt(mediaIndex);
                                    mediaFromTimeController.removeAt(
                                      mediaIndex,
                                    );
                                    mediaToTimeController.removeAt(mediaIndex);
                                    dayOfWeekSeleted.removeAt(mediaIndex);

                                    setState(() {});
                                  },
                                  icon: const Icon(
                                    Icons.delete,
                                    color: Colors.red,
                                  ),
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
                                    initialDate: (timeForSales[mediaIndex].fromdate != null && timeForSales[mediaIndex].fromdate!.isNotEmpty)
                                        ? DateTime.tryParse(timeForSales[mediaIndex].fromdate!)
                                        : dateNow,
                                    firstDate: DateTime(2000),
                                    lastDate: DateTime(2100),
                                    onDateSelected: (date) {
                                      if (date != null) {
                                        setState(() {
                                          timeForSales[mediaIndex].fromdate = date.toLocal().toIso8601String();
                                          mediaFromDateController[mediaIndex].text = DateFormat('dd/MM/yyyy').format(date);
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
                                    initialDate: (timeForSales[mediaIndex].todate != null && timeForSales[mediaIndex].todate!.isNotEmpty)
                                        ? DateTime.tryParse(timeForSales[mediaIndex].todate!)
                                        : dateNow,
                                    firstDate: DateTime(2000),
                                    lastDate: DateTime(2100),
                                    onDateSelected: (date) {
                                      if (date != null) {
                                        setState(() {
                                          timeForSales[mediaIndex].todate = date.toLocal().toIso8601String();
                                          mediaToDateController[mediaIndex].text = DateFormat('dd/MM/yyyy').format(date);
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
                                    controller:
                                        mediaFromTimeController[mediaIndex],
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
                                    inputFormatters: [
                                      global.TimeInputFormatter(),
                                    ],
                                  ),
                                ),
                                SizedBox(width: 10),
                                Expanded(
                                  child: TextFormField(
                                    controller:
                                        mediaToTimeController[mediaIndex],
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
                                    inputFormatters: [
                                      global.TimeInputFormatter(),
                                    ],
                                  ),
                                ),
                              ],
                            ),

                            const SizedBox(height: 10),

                            /// dayofweek checkbox
                            /* XXXX
                              DropdownSearch<DayOfWeekModel>.multiSelection(
                                enabled: isEditMode,
                                asyncItems: (String filter) => getDataDayOfWeek(filter),
                                compareFn: (item, selectedItem) => item.code == selectedItem.code,
                                itemAsString: (DayOfWeekModel? group) {
                                  if (group == null) return '';
                                  return '${group.code} - ${group.name}';
                                },
                                dropdownDecoratorProps: DropDownDecoratorProps(
                                  dropdownSearchDecoration: InputDecoration(
                                    border: const OutlineInputBorder(),
                                    contentPadding: const EdgeInsets.only(left: 10, top: 15, bottom: 10, right: 10),
                                    floatingLabelBehavior: FloatingLabelBehavior.always,
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
                              ),*/
                          ],
                        )
                      : Container(),
                (isShowTimeForSale)
                    ? Container(
                        width: double.infinity,
                        padding: const EdgeInsets.only(bottom: 10, top: 10),
                        child: ElevatedButton.icon(
                          style: ElevatedButton.styleFrom(
                            backgroundColor: Colors.green.shade300,
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

                            mediaFromDateController.add(
                              TextEditingController(),
                            );
                            mediaToDateController.add(TextEditingController());
                            mediaFromTimeController.add(
                              TextEditingController(),
                            );
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
                          icon: Icon(Icons.add),
                          label: Text(global.language("add_time_for_sale")),
                        ),
                      )
                    : Container(),
                Container(
                  width: double.infinity,
                  padding: EdgeInsets.only(bottom: 10),
                  child: Row(
                    children: [
                      Radio(
                        focusNode: FocusNode(skipTraversal: true),
                        value: false,
                        groupValue: useImageOrColor,
                        onChanged: (value) {
                          setState(() {
                            useImageOrColor = false;
                          });
                        },
                      ),
                      Text(global.language("use_image")),
                      const SizedBox(width: 10),
                      Radio(
                        focusNode: FocusNode(skipTraversal: true),
                        value: true,
                        groupValue: useImageOrColor,
                        onChanged: (value) {
                          setState(() {
                            useImageOrColor = true;
                          });
                        },
                      ),
                      Text(global.language("use_color")),
                    ],
                  ),
                ),
                if (!useImageOrColor)
                  Row(
                    children: [
                      Expanded(
                        child: ElevatedButton.icon(
                          focusNode: FocusNode(skipTraversal: true),
                          onPressed: () async {
                            setState(() {
                              imageWeb = null;
                              selectImageUri = "";
                              imageFile = File('');
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
                                  XFile? image = await _picker.pickImage(
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
                                  final XFile? photo = await _picker.pickImage(
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
                              final XFile? photo = await _picker.pickImage(
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
                const SizedBox(height: 10),
                if (!useImageOrColor)
                  SizedBox(
                    width: 500,
                    height: 500,
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
                            final bytes = await dropZoneController.getFileData(
                              ev,
                            );
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
                                  : (selectImageUri != '')
                                  ? DecorationImage(
                                      image: NetworkImage(selectImageUri),
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
                              height: 400,
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                const SizedBox(height: 10),
                (useImageOrColor)
                    ? Container(
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
                                      global.language("choose_color"),
                                      style: Theme.of(
                                        context,
                                      ).textTheme.titleMedium,
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
                                colorSelectedHex = colorValue.value
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
                      )
                    : Container(),
                const SizedBox(height: 10),

                /// text title รูปหน้าปก
                Align(
                  alignment: Alignment.centerLeft,
                  child: Text(
                    global.language("image_cover"),
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
                            selectImageUriCover = "";
                          });
                        },
                        icon: Icon(Icons.delete),
                        label: Text(global.language('delete_picture_cover')),
                      ),
                    ),
                    SizedBox(width: 5),
                    Expanded(
                      child: ElevatedButton.icon(
                        focusNode: FocusNode(skipTraversal: true),
                        onPressed: () async {
                          XFile? image = await _picker.pickImage(
                            source: ImageSource.gallery,
                            maxHeight: 480,
                            maxWidth: 640,
                            imageQuality: 60,
                          );
                          if (image != null) {
                            var f = await image.readAsBytes();
                            setState(() {
                              imageWebCover[0] = f;
                              imageFileCover[0] = File(image.path);
                              upLoadImage();
                            });
                          }
                        },
                        icon: Icon(Icons.folder),
                        label: Text(global.language("select_picture_cover")),
                      ),
                    ),
                    const SizedBox(width: 5),
                  ],
                ),
                const SizedBox(height: 10),
                SizedBox(
                  width: 500,
                  height: 500,
                  child: Stack(
                    children: [
                      Center(
                        child: DecoratedBox(
                          decoration: BoxDecoration(
                            color: global.theme.cardColor,
                            border: Border.all(color: global.theme.textColor),
                            borderRadius: BorderRadius.circular(5),
                            image: (selectImageUriCover != '')
                                ? DecorationImage(
                                    image: NetworkImage(selectImageUriCover),
                                    fit: BoxFit.fill,
                                  )
                                : const DecorationImage(
                                    image: AssetImage('assets/img/noimage.png'),
                                  ),
                          ),
                          child: const SizedBox(
                            width: double.infinity,
                            height: 400,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
                const SizedBox(height: 10),
                if (isEditMode)
                  SizedBox(
                    width: double.infinity,
                    child: ElevatedButton.icon(
                      focusNode: FocusNode(skipTraversal: true),
                      onPressed: () {
                        saveOrUpdateData();
                      },
                      icon: Icon(Icons.save),
                      label: Text(
                        global.language("save") + ((kIsWeb) ? " (F10)" : ""),
                      ),
                    ),
                  ),
              ],
            ),
          ),
        ),
        ),
      );
    },
  ),
      ),
    );
  }

  void clearEditData() {
    for (int i = 0; i < fieldTextController.length; i++) {
      fieldTextController[i].clear();
    }
    isChange = false;
    focusNodeIndex = 0;
    groupNumber.text = widget.groupnumber.toString();
    selectGuid = "";

    _selectedTime = [TimeOfDay.now(), TimeOfDay.now()];
    dayOfWeekSeleted = [];
    mediaFromDateController = [];
    mediaToDateController = [];
    mediaFromTimeController = [];
    mediaToTimeController = [];
    timeForSales = [];

    isShowTimeForSale = false;

    setState(() {
      imageFile = File('');
      imageWeb = null;
      selectImageUri = "";
      selectImageUriCover = "";
    });
  }

  void getDataToEditScreen(ProductCategoryModel category) {
    isChange = false;
    dataTemp = category;
    selectGuid = category.guidfixed;
    selectParentGuid = category.parentguid;
    selectImageUri = category.imageuri;
    selectImageUriCover = category.coveruri!;
    codeList = category.codelist!;
    imageWeb = null;
    imageFile = File('');
    xorder = category.xsorts![0].xorder;
    useImageOrColor = category.useimageorcolor;
    isdisabled = category.isdisabled;
    colorSelected = global.colorFromHex(
      category.colorselecthex.replaceAll("#", ""),
    );
    colorSelectedHex = category.colorselecthex;
    for (int i = 0; i < languageList.length; i++) {
      fieldTextController[i].text = "";
    }
    for (int i = 0; i < languageList.length; i++) {
      for (int j = 0; j < category.names!.length; j++) {
        if (languageList[i].code == category.names![j].code) {
          fieldTextController[i].text = category.names![j].name;
        }
      }
    }

    groupNumber.text = widget.groupnumber.toString();

    if (category.timeforsales!.isNotEmpty) {
      isShowTimeForSale = true;
      for (int i = 0; i < category.timeforsales!.length; i++) {
        timeForSales.add(category.timeforsales![i]);
        _selectedTime.add(TimeOfDay.now());
        mediaFromDateController.add(TextEditingController());
        mediaToDateController.add(TextEditingController());
        mediaFromTimeController.add(TextEditingController());
        mediaToTimeController.add(TextEditingController());

        mediaFromDateController[i].text =
            (category.timeforsales![i].fromdate!.isNotEmpty)
            ? DateFormat(
                'dd/MM/yyyy',
              ).format(DateTime.parse(category.timeforsales![i].fromdate!))
            : '';
        mediaToDateController[i].text =
            (category.timeforsales![i].todate!.isNotEmpty)
            ? DateFormat(
                'dd/MM/yyyy',
              ).format(DateTime.parse(category.timeforsales![i].todate!))
            : '';
        mediaFromTimeController[i].text = category.timeforsales![i].fromtime!;
        mediaToTimeController[i].text = category.timeforsales![i].totime!;
        dayOfWeekSeleted.add(
          category.timeforsales![i].daysofweek!
              .map(
                (e) => dayOfWeekList.firstWhere(
                  (element) => element.code == e.toString(),
                ),
              )
              .toList(),
        );
      }
    }

    setState(() {});
  }

  void findFocusNext(int index) {
    focusNodeIndex = index + 1;
    if (focusNodeIndex > fieldFocusNodes.length - 1) {
      focusNodeIndex = 0;
    }
    focusAndCursorToEnd(fieldFocusNodes[focusNodeIndex]);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          bool mobileScreen = (constraints.maxWidth < 800.0);
          return MultiBlocListener(
            listeners: [
              BlocListener<ProductCategoryBloc, ProductCategoryState>(
                listener: (context, state) {
                  blocCategoryState = state;
                  // Load
                  if (state is ProductCategoryLoadSuccess) {
                    rootCategorys.clear();
                    for (var item in state.productCategorys) {
                      rootCategorys.add(
                        ProductCategoryListModel(
                          detail: item,
                          childCategories: [],
                        ),
                      );
                    }
                    for (int loop = 0; loop < rootCategorys.length; loop++) {
                      if (rootCategorys[loop].detail.xsorts == null) {
                        rootCategorys[loop].detail.xsorts = [
                          SortDataModel(code: "X", xorder: 0),
                        ];
                        // print(rootCategorys[loop].detail.names![0].name);
                      }
                    }
                    // Sort
                    rootCategorys.sort(
                      (a, b) => a.detail.xsorts![0].xorder.compareTo(
                        b.detail.xsorts![0].xorder,
                      ),
                    );
                    int index = 0;
                    while (index < rootCategorys.length) {
                      if (rootCategorys[index].detail.parentguid.isNotEmpty) {
                        ProductCategoryListModel? findCategory = findByGuid(
                          rootCategorys,
                          rootCategorys[index].detail.parentguid,
                        );
                        if (findCategory != null) {
                          findCategory.childCategories.add(
                            rootCategorys[index],
                          );
                          rootCategorys.removeAt(index);
                        } else {
                          index++;
                        }
                      } else {
                        index++;
                      }
                    }
                    // Build
                    buildColumnWidget();
                    setState(() {});
                  }
                },
              ),
              BlocListener<ProductCategoryBloc, ProductCategoryState>(
                listener: (context, state) {
                  // Save
                  if (state is ProductCategorySaveSuccess) {
                    global.showSnackBar(
                      context,
                      Icon(Icons.save, color: Colors.white),
                      global.language("save_success"),
                      Colors.blue,
                    );
                    clearEditData();
                    loadDataList("");
                    setState(() {});
                  }
                  if (state is ProductCategorySaveFailed) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.save, color: Colors.white),
                        "${global.language("not_success_save")} : ${state.message}",
                        Colors.red,
                      );
                    });
                  }
                },
              ),
              BlocListener<ProductCategoryBloc, ProductCategoryState>(
                listener: (context, state) {
                  // Update
                  if (state is ProductCategoryUpdateSuccess) {
                    clearEditData();

                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.edit, color: Colors.white),
                        global.language("edit_success"),
                        Colors.blue,
                      );
                      WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                        tabController.animateTo(0);
                      });
                      loadDataList("");
                      isSaveAllow = false;
                      isEditMode = false;
                      selectDragTargetGuid = "";
                    });
                  }
                  if (state is ProductCategoryUpdateFailed) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.edit, color: Colors.white),
                        "${global.language("not_edit_success")} : ${state.message}",
                        Colors.red,
                      );
                    });
                  }
                },
              ),
              BlocListener<ProductCategoryBloc, ProductCategoryState>(
                listener: (context, state) {
                  // Delete
                  if (state is ProductCategoryDeleteSuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.delete, color: Colors.white),
                        global.language("delete_success"),
                        Colors.blue,
                      );
                      selectGuid = "";
                      isSaveAllow = false;
                      isEditMode = false;
                      clearEditData();
                      WidgetsBinding.instance.addPostFrameCallback((timeStamp) {
                        tabController.animateTo(0);
                      });
                      loadDataList("");
                    });
                  }
                },
              ),
              BlocListener<ProductCategoryBloc, ProductCategoryState>(
                listener: (context, state) {
                  // Delete Many
                  if (state is ProductCategoryDeleteManySuccess) {
                    setState(() {
                      global.showSnackBar(
                        context,
                        Icon(Icons.delete, color: Colors.white),
                        global.language("delete_success"),
                        Colors.blue,
                      );
                    });
                  }
                },
              ),
              BlocListener<ProductCategoryBloc, ProductCategoryState>(
                listener: (context, state) {
                  // Get
                  if (state is ProductCategoryGetSuccess) {
                    setState(() {
                      clearEditData();
                      getDataToEditScreen(state.category);
                      if (mobileScreen && selectDragTargetGuid.isEmpty) {
                        WidgetsBinding.instance.addPostFrameCallback((
                          timeStamp,
                        ) {
                          tabController.animateTo(1);
                        });
                      }
                    });
                  }
                },
              ),
              BlocListener<ImageUploadBloc, ImageUploadState>(
                listener: (context, state) {
                  if (state is ImageUploadSaveSuccess) {
                    setState(() {
                      selectImageUriCover = state.imageUpload.uri;
                    });
                  }
                },
              ),
            ],
            child: (mobileScreen == false)
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
