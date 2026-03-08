import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:smlaicloud/global.dart' as global;

// Generic configuration class for different types of selections
class SelectionConfig<T> {
  final String title;
  final String searchHint;
  final List<String> columnHeaders;
  final List<int> columnFlexes;
  final Function(T) onItemSelected;
  final String Function(T) getCode;
  final String Function(T) getName;
  final String Function(T)? getSecondaryInfo;
  final String Function(T)? getTertiaryInfo;
  final void Function(int offset, int limit, String search) loadData;
  final List<T> Function() getCurrentData;
  final bool Function() isLoading;
  final bool Function() hasError;
  final String Function() getErrorMessage;
  final void Function()? onRefresh;
  final String Function(T) getGuidFixed; // Add this for keyboard navigation

  SelectionConfig({
    required this.title,
    required this.searchHint,
    required this.columnHeaders,
    required this.columnFlexes,
    required this.onItemSelected,
    required this.getCode,
    required this.getName,
    this.getSecondaryInfo,
    this.getTertiaryInfo,
    required this.loadData,
    required this.getCurrentData,
    required this.isLoading,
    required this.hasError,
    required this.getErrorMessage,
    this.onRefresh,
    required this.getGuidFixed, // Add this for keyboard navigation
  });
}

class GenericSelectionDialog<T> extends StatefulWidget {
  final SelectionConfig<T> config;

  const GenericSelectionDialog({super.key, required this.config});

  @override
  State<GenericSelectionDialog<T>> createState() =>
      _GenericSelectionDialogState<T>();
}

class _GenericSelectionDialogState<T> extends State<GenericSelectionDialog<T>> {
  @override
  Widget build(BuildContext context) {
    return Dialog(
      insetPadding: const EdgeInsets.all(20),
      child: SizedBox(
        width: MediaQuery.of(context).size.width * 0.8,
        height: MediaQuery.of(context).size.height * 0.8,
        child: GenericSelectionContent<T>(
          config: widget.config,
          listData: widget.config.getCurrentData(),
        ),
      ),
    );
  }
}

// Generic Selection Content for full screen usage
class GenericSelectionContent<T> extends StatefulWidget {
  final SelectionConfig<T> config;
  final List<T> listData;

  const GenericSelectionContent({
    super.key,
    required this.config,
    required this.listData,
  });

  @override
  State<GenericSelectionContent<T>> createState() =>
      _GenericSelectionContentState<T>();
}

class _GenericSelectionContentState<T>
    extends State<GenericSelectionContent<T>> {
  final TextEditingController searchController = TextEditingController();
  final FocusNode searchFocusNode = FocusNode(skipTraversal: true);
  final ScrollController listScrollController = ScrollController();
  String searchText = "";
  String selectGuid = "";
  int currentListIndex = 0;
  bool isKeyUp = false;
  bool isKeyDown = false;
  final debouncer = global.Debouncer(1000);
  @override
  void initState() {
    super.initState();
    loadDataList("");
    listScrollController.addListener(onScrollList);
  }

  @override
  void dispose() {
    searchController.dispose();
    searchFocusNode.dispose();
    listScrollController.dispose();
    super.dispose();
  }

  void loadDataList(String search) {
    widget.config.loadData(
      (widget.listData.isEmpty) ? 0 : widget.listData.length,
      global.loadDataPerPage,
      search,
    );
  }

  void onScrollList() {
    if (listScrollController.offset >=
            listScrollController.position.maxScrollExtent &&
        !listScrollController.position.outOfRange) {
      loadDataList(searchText);
    }
  }

  @override
  Widget build(BuildContext context) {
    // Update selected item for keyboard navigation
    if (widget.listData.isNotEmpty && selectGuid.isEmpty) {
      selectGuid = widget.config.getGuidFixed(widget.listData[0]);
    }

    for (int i = 0; i < widget.listData.length; i++) {
      if (widget.config.getGuidFixed(widget.listData[i]) == selectGuid) {
        currentListIndex = i;
        break;
      }
    }

    return Focus(
      focusNode: FocusNode(skipTraversal: true, canRequestFocus: true),
      onKey: (node, event) {
        if (kIsWeb ||
            Platform.isWindows ||
            Platform.isLinux ||
            Platform.isMacOS) {
          if (event is RawKeyDownEvent) {
            if (event.logicalKey == LogicalKeyboardKey.escape) {
              Navigator.pop(context);
              return KeyEventResult.handled;
            }
            if (event.logicalKey == LogicalKeyboardKey.tab) {
              if (selectGuid.isNotEmpty && widget.listData.isNotEmpty) {
                final selectedItem = widget.listData.firstWhere(
                  (element) =>
                      widget.config.getGuidFixed(element) == selectGuid,
                  orElse: () => widget.listData[0],
                );
                widget.config.onItemSelected(selectedItem);
                return KeyEventResult.handled;
              }
            }
            if (event.logicalKey == LogicalKeyboardKey.arrowUp) {
              isKeyDown = false;
              if (widget.listData.isNotEmpty) {
                int index = widget.listData.indexWhere(
                  (element) =>
                      widget.config.getGuidFixed(element) == selectGuid,
                );
                if (index > 0) {
                  selectGuid = widget.config.getGuidFixed(
                    widget.listData[index - 1],
                  );
                  isKeyUp = true;
                }
                setState(() {});
              }
              return KeyEventResult.handled;
            }
            if (event.logicalKey == LogicalKeyboardKey.arrowDown) {
              isKeyUp = false;
              if (widget.listData.isNotEmpty) {
                int index = widget.listData.indexWhere(
                  (element) =>
                      widget.config.getGuidFixed(element) == selectGuid,
                );
                if (index < widget.listData.length - 1) {
                  selectGuid = widget.config.getGuidFixed(
                    widget.listData[index + 1],
                  );
                }
                isKeyDown = true;
                setState(() {});
              }
              return KeyEventResult.handled;
            }
            if (event.logicalKey == LogicalKeyboardKey.enter) {
              if (selectGuid.isNotEmpty && widget.listData.isNotEmpty) {
                final selectedItem = widget.listData.firstWhere(
                  (element) =>
                      widget.config.getGuidFixed(element) == selectGuid,
                  orElse: () => widget.listData[0],
                );
                widget.config.onItemSelected(selectedItem);
                return KeyEventResult.handled;
              }
            }
          }
        }
        return KeyEventResult.ignored;
      },
      child: Column(
        children: [
          // Search Bar (matching ProductSearchScreen style)
          Container(
            padding: const EdgeInsets.all(5),
            color: global.theme.appBarColor,
            child: Container(
              padding: const EdgeInsets.all(5),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(2),
                boxShadow: [
                  BoxShadow(
                    color: Colors.grey.withOpacity(0.5),
                    spreadRadius: 5,
                    blurRadius: 7,
                    offset: const Offset(0, 2),
                  ),
                ],
              ),
              child: Padding(
                padding: const EdgeInsets.only(left: 10, right: 10),
                child: TextFormField(
                  onFieldSubmitted: (value) {
                    searchFocusNode.requestFocus();
                  },
                  onChanged: (value) {
                    debouncer.run(() {
                      try {
                        setState(() {
                          searchText = value;
                          widget.listData.clear();
                          selectGuid = "";
                        });
                        loadDataList(value);
                      } catch (_) {}
                    });
                  },
                  autofocus: true,
                  focusNode: searchFocusNode,
                  controller: searchController,
                  decoration: InputDecoration(
                    isDense: true,
                    contentPadding: const EdgeInsets.only(top: 10, bottom: 10),
                    border: InputBorder.none,
                    hintText: widget.config.searchHint,
                  ),
                ),
              ),
            ),
          ),

          // Column Headers (matching ProductSearchScreen style)
          Container(
            padding: const EdgeInsets.only(
              left: 10,
              right: 10,
              top: 5,
              bottom: 5,
            ),
            decoration: BoxDecoration(
              color: global.theme.columnHeaderColor,
              border: const Border(
                bottom: BorderSide(width: 1.0, color: Colors.grey),
              ),
            ),
            child: Row(
              children: widget.config.columnHeaders.asMap().entries.map((
                entry,
              ) {
                int index = entry.key;
                String header = entry.value;
                return Expanded(
                  flex: widget.config.columnFlexes[index],
                  child: Text(
                    header,
                    style: const TextStyle(
                      color: Colors.black,
                      fontWeight: FontWeight.bold,
                    ),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                );
              }).toList(),
            ),
          ),

          // List (matching ProductSearchScreen style)
          Expanded(
            child: widget.config.isLoading() && widget.listData.isEmpty
                ? const Center(child: CircularProgressIndicator())
                : widget.listData.isEmpty && !widget.config.isLoading()
                ? Center(
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(
                          Icons.search_off,
                          size: 64,
                          color: Colors.grey[400],
                        ),
                        SizedBox(height: 16),
                        Text(
                          global.language("no_data_found"),
                          style: TextStyle(
                            fontSize: 16,
                            color: Colors.grey[600],
                          ),
                        ),
                      ],
                    ),
                  )
                : SingleChildScrollView(
                    controller: listScrollController,
                    child: Column(
                      children: widget.listData
                          .map((item) => _buildListItem(item))
                          .toList(),
                    ),
                  ),
          ),
        ],
      ),
    );
  }

  Widget _buildListItem(T item) {
    return GestureDetector(
      onTap: () {
        widget.config.onItemSelected(item);
      },
      child: Container(
        decoration: BoxDecoration(
          color: (selectGuid == widget.config.getGuidFixed(item))
              ? Colors.cyan[100]
              : Colors.white,
          border: const Border(
            bottom: BorderSide(width: 1.0, color: Colors.grey),
          ),
        ),
        padding: const EdgeInsets.only(left: 10, right: 10, top: 5, bottom: 5),
        child: Row(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: _buildRowContent(item),
        ),
      ),
    );
  }

  List<Widget> _buildRowContent(T item) {
    List<Widget> widgets = [];

    // Add code column
    widgets.add(
      Expanded(
        flex: widget.config.columnFlexes[0],
        child: Text(
          widget.config.getCode(item),
          maxLines: 2,
          overflow: TextOverflow.ellipsis,
        ),
      ),
    );

    // Add name column
    widgets.add(
      Expanded(
        flex: widget.config.columnFlexes[1],
        child: Text(
          widget.config.getName(item),
          maxLines: 2,
          overflow: TextOverflow.ellipsis,
        ),
      ),
    );

    // Add secondary info column if available
    if (widget.config.getSecondaryInfo != null &&
        widget.config.columnFlexes.length > 2) {
      widgets.add(
        Expanded(
          flex: widget.config.columnFlexes[2],
          child: Text(
            widget.config.getSecondaryInfo!(item),
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
        ),
      );
    }

    // Add tertiary info column if available
    if (widget.config.getTertiaryInfo != null &&
        widget.config.columnFlexes.length > 3) {
      widgets.add(
        Expanded(
          flex: widget.config.columnFlexes[3],
          child: Text(
            widget.config.getTertiaryInfo!(item),
            maxLines: 2,
            overflow: TextOverflow.ellipsis,
          ),
        ),
      );
    }

    return widgets;
  }
}
