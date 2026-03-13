import 'package:smlaicloud/model/product_category_list_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/product_category/product_category_bloc.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/product_category_model.dart';
import 'package:smlaicloud/model/global_model.dart';

class ProductCategorySearchScreen extends StatefulWidget {
  final int groupnumber;
  const ProductCategorySearchScreen({super.key, required this.groupnumber});

  @override
  State<ProductCategorySearchScreen> createState() =>
      ProductCategorySearchScreenState();
}

class ProductCategorySearchScreenState
    extends State<ProductCategorySearchScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  TextEditingController searchController = TextEditingController();
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  ScrollController listScrollController = ScrollController();

  String searchText = "";
  String selectParentGuid = "";
  List<ProductCategoryModel> categoryListDatas = [];
  List<ProductCategoryListModel> rootCategorys = [];
  List<Widget> listColumns = [];
  bool isKeyUp = false;
  bool isKeyDown = false;
  String selectGuid = "";
  int _hoverIndex = -1;
  int currentListIndex = 0;
  List<String> xorderUpdateList = [];
  void setSystemLanguageList() async {
    await global.setSystemLanguage(context);

    loadDataList("");
  }

  @override
  void initState() {
    setSystemLanguageList();
    listScrollController.addListener(onScrollList);

    super.initState();
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

  // void loadDataList(String search) {
  //   context.read<CategoryBloc>().add(CategoryLoadList(
  //       offset: (categoryListDatas.isEmpty) ? 0 : categoryListDatas.length,
  //       limit: global.loadDataPerPage,
  //       search: search));
  // }

  void onScrollList() {
    if (listScrollController.offset >=
            listScrollController.position.maxScrollExtent &&
        !listScrollController.position.outOfRange) {
      //loadDataList(searchText);
    }
  }

  @override
  void dispose() {
    listScrollController.dispose();
    searchController.dispose();
    super.dispose();
  }

  Widget listScreen({bool mobileScreen = false}) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('product_group')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: Icon(Icons.arrow_back),
          onPressed: () {
            Navigator.pop(context);
          },
        ),
      ),
      body: Focus(
        focusNode: FocusNode(skipTraversal: true, canRequestFocus: true),
        onKey: (node, event) {
          if (kIsWeb) {
            if (event is RawKeyDownEvent) {
              if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
                isKeyDown = false;
                int index = categoryListDatas.indexOf(
                  categoryListDatas.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                if (index > 0) {
                  selectGuid = categoryListDatas[index - 1].guidfixed;
                  currentListIndex = index + 1;
                  isKeyUp = true;
                }
              }
              if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
                isKeyUp = false;
                int index = categoryListDatas.indexOf(
                  categoryListDatas.firstWhere(
                    (element) => element.guidfixed == selectGuid,
                  ),
                );
                selectGuid = categoryListDatas[index + 1].guidfixed;
                currentListIndex = index + 1;
                isKeyDown = true;
              }
            }
          }
          return KeyEventResult.ignored;
        },
        child: Column(
          children: [
            Expanded(
              child: SingleChildScrollView(
                controller: listScrollController,
                child: Column(children: listColumns),
              ),
            ),
          ],
        ),
      ),
    );
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

  Widget listObject(ProductCategoryModel value, int index) {
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
        setState(() {
          Navigator.pop(context, [value.guidfixed, value.names!]);
        });
      },
      child: Container(
        decoration: BoxDecoration(
          color: _getContainerColor(value.guidfixed ?? "", index),
          border: Border(
                bottom: BorderSide(width: 1.0, color: global.theme.dividerBorderColor),
          ),
        ),
        padding: const EdgeInsets.only(left: 10, right: 10, top: 5, bottom: 5),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Expanded(
              flex: 5,
              child: Text(
                value.guidfixed,
                style: textStyle,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
            Expanded(
              flex: 10,
              child: Text(
                global.packName(value.names!),
                style: textStyle,
                maxLines: 2,
                overflow: TextOverflow.ellipsis,
              ),
            ),
          ],
        ),
      ),
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

  void updateXorderAll(List<ProductCategoryListModel> categorys) {
    // Update Xorder ทุกตัว
    for (var index = 0; index < categorys.length; index++) {
      xorderUpdateList.add(categorys[index].detail.guidfixed);
    }
  }

  Widget categoryDetail(
    int level,
    ProductCategoryListModel category,
    int index,
    int count, {
    required Function moveUpCallBack,
    required Function moveDownCallBack,
  }) {
    int codeListCount = category.detail.codelist?.length ?? 0;
    return Container(
      padding: const EdgeInsets.all(4),
      width: double.infinity,
      child: Row(
        children: <Widget>[
          SizedBox(width: level * 20),
          Expanded(
            child: Text(
              "${category.detail.xsorts![0].xorder} ${global.packName(category.detail.names!)} ($codeListCount)",
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
                  category.isExpand = !category.isExpand;
                  buildColumnWidget();
                });
              },
            ),
        ],
      ),
    );
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
            selectGuid = categorys[index].detail.guidfixed;
            Navigator.pop(context, [
              selectGuid,
              categorys[index].detail.names!,
            ]);
          },
          child: detail,
        ),
      );
      if (categorys[index].childCategories.isNotEmpty &&
          categorys[index].isExpand) {
        categoryList(level + 1, categorys[index].childCategories);
      }
    }

    return Column(children: listColumns);
  }

  void buildColumnWidget() {
    listColumns.clear();
    categoryList(0, rootCategorys);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: global.theme.backgroundColor,
      resizeToAvoidBottomInset: true,
      body: LayoutBuilder(
        builder: (context, constraints) {
          return BlocListener<ProductCategoryBloc, ProductCategoryState>(
            listener: (context, state) {
              // Load
              if (state is ProductCategoryLoadSuccess) {
                rootCategorys.clear();
                for (var item in state.productCategorys) {
                  rootCategorys.add(
                    ProductCategoryListModel(detail: item, childCategories: []),
                  );
                }
                for (int loop = 0; loop < rootCategorys.length; loop++) {
                  if (rootCategorys[loop].detail.xsorts == null) {
                    rootCategorys[loop].detail.xsorts = [
                      SortDataModel(code: "X", xorder: 0),
                    ];
                    //  // print(rootCategorys[loop].detail.names![0].name);
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
                      findCategory.childCategories.add(rootCategorys[index]);
                      rootCategorys.removeAt(index);
                    } else {
                      index++;
                    }
                  } else {
                    index++;
                  }
                }

                buildColumnWidget();
                setState(() {});
              }
            },
            child: (constraints.maxWidth > 800)
                ? listScreen(mobileScreen: false)
                : listScreen(mobileScreen: true),
          );
        },
      ),
    );
  }
}
