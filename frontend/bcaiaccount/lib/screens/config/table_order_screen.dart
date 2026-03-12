import 'package:smlaicloud/bloc/table/table_bloc.dart';
import 'package:smlaicloud/bloc/zone/zone_bloc.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/table_model.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:split_view/split_view.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:flutter_draggable_gridview/flutter_draggable_gridview.dart';

class TableOrderScreen extends StatefulWidget {
  final int groupnumber;
  const TableOrderScreen({super.key, required this.groupnumber});

  @override
  State<TableOrderScreen> createState() => TableOrderScreenState();
}

class TableOrderScreenState extends State<TableOrderScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  TextEditingController searchController = TextEditingController();
  List<DraggableGridItem> draggableGridItemList = [];
  List<ProductBarcodeModel> listData = [];
  bool loadingData = false;
  bool isDataChange = false;
  FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  List<TableModel> tableList = [];
  String searchText = "";
  String selectGuid = "";
  String tableGuid = "";
  String tableName = "";
  TableModel? table;
  late SplitViewController splitViewController;
  @override
  void initState() {
    loadDataZoneList();
    loadDataList("");

    super.initState();
  }

  @override
  void dispose() {
    searchController.dispose();
    searchFocusNode.dispose();
    super.dispose();
  }

  void rebuildGrid() {
    draggableGridItemList = [];
    for (int i = 0; i < tableList.length; i++) {
      draggableGridItemList.add(
        DraggableGridItem(
          isDraggable: true,
          child: Padding(
            padding: const EdgeInsets.all(8.0),
            child: objectBox(tableList[i]),
          ),
        ),
      );
    }
  }

  Widget objectBox(TableModel data) {
    return Container(
      width: 150,
      height: 150,
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        border: Border.all(color: global.theme.dividerBorderColor),
      ),
      child: Column(
        children: [
          Expanded(
            child: Padding(
              padding: const EdgeInsets.only(top: 10, left: 10, right: 10),
              child: Column(
                children: [
                  Text(
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                    data.number,
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  Expanded(
                    child: Text(
                      global.packName(data.names),
                      maxLines: 3,
                      overflow: TextOverflow.ellipsis,
                      style: TextStyle(fontSize: 12),
                    ),
                  ),
                  Expanded(
                    child: Text(
                      "xorder :${data.xorder}",
                      maxLines: 3,
                      overflow: TextOverflow.ellipsis,
                      style: TextStyle(fontSize: 12),
                    ),
                  ),
                  Expanded(
                    child: Text(
                      "zone :${data.zone}",
                      maxLines: 3,
                      overflow: TextOverflow.ellipsis,
                      style: TextStyle(fontSize: 12),
                    ),
                  ),
                ],
              ),
            ),
          ),
        ],
      ),
    );
  }

  void discardData({required Function callBack}) {
    if (isDataChange) {
      showDialog<String>(
        context: context,
        builder: (BuildContext context) => AlertDialog(
          title: Text(global.language('data_editing')),
          content: Text(global.language('leave_this_screen')),
          actions: <Widget>[
            ElevatedButton(
              style: ElevatedButton.styleFrom(backgroundColor: global.theme.negativeHighlightTextColor),
              onPressed: () => Navigator.pop(context),
              child: Text(global.language('no')),
            ),
            ElevatedButton(
              style: ElevatedButton.styleFrom(backgroundColor: global.theme.infoHighlightTextColor),
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

  void loadDataZoneList() {
    // context
    //     .read<ZoneBloc>()
    //     .add(const ZoneLoadList(offset: 0, limit: 1000, search: ""));
  }

  void loadDataList(String search) {
    setState(() {
      loadingData = true;
    });
    searchText = search;
    context.read<TableBloc>().add(
      TableLoadList(
        offset: (listData.isEmpty) ? 0 : listData.length,
        limit: global.loadDataPerPage,
        search: search,
        groupNumber: widget.groupnumber,
      ),
    );
  }

  Widget gridView2() {
    return LayoutBuilder(
      builder: (context, constraints) {
        return Container(
          padding: EdgeInsets.all(5),
          child: Column(
            children: [
              Expanded(
                child: (tableList.isEmpty)
                    ? Center(child: Text(global.language('no_setting_table')))
                    : DraggableGridViewBuilder(
                        gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                          crossAxisCount:
                              constraints.widthConstraints().maxWidth ~/ 150,
                        ),
                        children: draggableGridItemList,
                        isOnlyLongPress: true,
                        dragCompletion:
                            (
                              List<DraggableGridItem> list,
                              int beforeIndex,
                              int afterIndex,
                            ) {
                              if (beforeIndex != afterIndex) {
                                setState(() {
                                  isDataChange = true;
                                  tableList.insert(
                                    afterIndex,
                                    tableList.removeAt(beforeIndex),
                                  );
                                  for (int i = 0; i < tableList.length; i++) {
                                    tableList[i].xorder = i + 1;
                                  }
                                  rebuildGrid();
                                });
                              }
                            },
                        dragFeedback:
                            (List<DraggableGridItem> list, int index) {
                              return objectBox(tableList[index]);
                            },
                        dragPlaceHolder:
                            (List<DraggableGridItem> list, int index) {
                              return PlaceHolderWidget(
                                child: Container(color: global.theme.cardColor),
                              );
                            },
                      ),
              ),
            ],
          ),
        );
      },
    );
  }

  Widget gridView() {
    return LayoutBuilder(
      builder: (context, constraints) {
        return Container(
          padding: EdgeInsets.all(5),
          child: Column(
            children: [
              Expanded(
                child: (tableList.isEmpty)
                    ? Center(child: Text(global.language('no_setting_table')))
                    : DraggableGridViewBuilder(
                        gridDelegate: SliverGridDelegateWithFixedCrossAxisCount(
                          crossAxisCount:
                              constraints.widthConstraints().maxWidth ~/ 150,
                        ),
                        children: draggableGridItemList,
                        isOnlyLongPress: true,
                        dragCompletion:
                            (
                              List<DraggableGridItem> list,
                              int beforeIndex,
                              int afterIndex,
                            ) {
                              if (beforeIndex != afterIndex) {
                                setState(() {
                                  isDataChange = true;
                                  tableList.insert(
                                    afterIndex,
                                    tableList.removeAt(beforeIndex),
                                  );
                                  for (int i = 0; i < tableList.length; i++) {
                                    tableList[i].xorder = i + 1;
                                  }
                                  rebuildGrid();
                                });
                              }
                            },
                        dragFeedback:
                            (List<DraggableGridItem> list, int index) {
                              return objectBox(tableList[index]);
                            },
                        dragPlaceHolder:
                            (List<DraggableGridItem> list, int index) {
                              return PlaceHolderWidget(
                                child: Container(color: global.theme.cardColor),
                              );
                            },
                      ),
              ),
            ],
          ),
        );
      },
    );
  }

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        return MultiBlocListener(
          listeners: [
            BlocListener<TableBloc, TableState>(
              listener: (context, state) {
                // Load
                if (state is TableLoadSuccess) {
                  tableList = [];
                  for (int i = 0; i < state.tables.length; i++) {
                    tableList.add(
                      TableModel(
                        guidfixed: state.tables[i].guidfixed,
                        number: state.tables[i].number,
                        names: state.tables[i].names,
                        zone: state.tables[i].zone,
                        xorder: state.tables[i].xorder,
                      ),
                    );
                  }

                  /// sort by xorder
                  tableList.sort((a, b) => a.xorder!.compareTo(b.xorder!));

                  rebuildGrid();
                  setState(() {});
                }
                if (state is TableUpdateSuccess) {
                  setState(() {
                    global.showSnackBar(
                      context,
                      Icon(Icons.edit, color: global.theme.onPrimaryColor),
                      global.language('edit_success'),
                      global.theme.infoHighlightTextColor,
                    );
                    tableList = [];
                    isDataChange = false;
                    selectGuid = "";
                    tableGuid = "";
                    tableName = "";
                    loadDataList(searchText);
                    setState(() {});
                  });
                }
                if (state is TableUpdateFailed) {
                  setState(() {
                    global.showSnackBar(
                      context,
                      Icon(Icons.edit, color: global.theme.onPrimaryColor),
                      "${global.language('edit_failed')} : ${state.message}",
                      global.theme.negativeHighlightTextColor,
                    );
                  });
                }
              },
            ),
            BlocListener<ZoneBloc, ZoneState>(
              listener: (context, state) {
                if (state is ZoneLoadSuccess) {
                  if (state.zones.isNotEmpty) {
                    setState(() {
                      selectGuid = state.zones[0].guidfixed;
                      tableGuid = state.zones[0].guidfixed;
                      tableName = state.zones[0].names.toString();
                      loadDataList("");
                    });
                  }
                }
              },
            ),
          ],
          child: Scaffold(
            resizeToAvoidBottomInset: true,
            appBar: AppBar(
              backgroundColor: global.theme.appBarColor,
              automaticallyImplyLeading: false,
              title: Text(global.language('table_order')),
              leading: IconButton(
                focusNode: FocusNode(skipTraversal: true),
                icon: Icon(Icons.arrow_back),
                onPressed: () {
                  discardData(
                    callBack: () {
                      Navigator.pop(context);
                    },
                  );
                },
              ),
              actions: [
                /// save button
                IconButton(
                  focusNode: FocusNode(skipTraversal: true),
                  icon: Icon(Icons.save),
                  onPressed: () {
                    List<TableXorderModel> dataSave = [];
                    for (int i = 0; i < tableList.length; i++) {
                      dataSave.add(
                        TableXorderModel(
                          guidfixed: tableList[i].guidfixed,
                          xorder: tableList[i].xorder!,
                        ),
                      );
                    }
                    context.read<TableBloc>().add(
                      TableUpdateXorder(tableModel: dataSave),
                    );
                  },
                ),
              ],
            ),
            body: Container(
              child: Column(
                children: [
                  // Expanded(child: gridView2()),
                  Expanded(child: gridView()),
                ],
              ),
            ),
          ),
        );
      },
    );
  }
}
