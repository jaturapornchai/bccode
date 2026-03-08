import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/bloc/master_group/master_group_bloc.dart';
import 'package:smlaicloud/bloc/master_group_sub1/master_group_sub1_bloc.dart';
import 'package:smlaicloud/bloc/master_group_sub2/master_group_sub2_bloc.dart';
import 'package:smlaicloud/bloc/master_brand/master_brand_bloc.dart';
import 'package:smlaicloud/bloc/master_category/master_category_bloc.dart';
import 'package:smlaicloud/bloc/master_class/master_class_bloc.dart';
import 'package:smlaicloud/bloc/master_design/master_design_bloc.dart';
import 'package:smlaicloud/bloc/master_grade/master_grade_bloc.dart';
import 'package:smlaicloud/bloc/master_model/master_model_bloc.dart';
import 'package:smlaicloud/bloc/master_pattern/master_pattern_bloc.dart';
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
import 'package:smlaicloud/screens/master/generic_selection_components.dart';
import 'package:smlaicloud/global.dart' as global;

enum MasterDataType {
  group,
  groupSub1,
  groupSub2,
  brand,
  category,
  masterClass,
  design,
  grade,
  model,
  pattern,
}

class GroupSelectionHelper {
  // Generic method to show master data selection screen
  static void showMasterDataSelectionScreen<T>({
    required BuildContext context,
    required MasterDataType dataType,
    required Function(T) onItemSelected,
  }) {
    switch (dataType) {
      case MasterDataType.group:
        showMasterGroupSelectionScreen(
          context: context,
          onGroupSelected: onItemSelected as Function(MasterGroupModel),
        );
        break;
      case MasterDataType.groupSub1:
        showMasterGroupSub1SelectionScreen(
          context: context,
          onGroupSelected: onItemSelected as Function(MasterGroupSub1Model),
        );
        break;
      case MasterDataType.groupSub2:
        showMasterGroupSub2SelectionScreen(
          context: context,
          onGroupSelected: onItemSelected as Function(MasterGroupSub2Model),
        );
        break;
      case MasterDataType.brand:
        showMasterBrandSelectionScreen(
          context: context,
          onItemSelected: onItemSelected as Function(MasterBrandModel),
        );
        break;
      case MasterDataType.category:
        showMasterCategorySelectionScreen(
          context: context,
          onItemSelected: onItemSelected as Function(MasterCategoryModel),
        );
        break;
      case MasterDataType.masterClass:
        showMasterClassSelectionScreen(
          context: context,
          onItemSelected: onItemSelected as Function(MasterClassModel),
        );
        break;
      case MasterDataType.design:
        showMasterDesignSelectionScreen(
          context: context,
          onItemSelected: onItemSelected as Function(MasterDesignModel),
        );
        break;
      case MasterDataType.grade:
        showMasterGradeSelectionScreen(
          context: context,
          onItemSelected: onItemSelected as Function(MasterGradeModel),
        );
        break;
      case MasterDataType.model:
        showMasterModelSelectionScreen(
          context: context,
          onItemSelected: onItemSelected as Function(MasterModelModel),
        );
        break;
      case MasterDataType.pattern:
        showMasterPatternSelectionScreen(
          context: context,
          onItemSelected: onItemSelected as Function(MasterPatternModel),
        );
        break;
    }
  }

  // Navigate to Master Group Selection Screen
  static void showMasterGroupSelectionScreen({
    required BuildContext context,
    required Function(MasterGroupModel) onGroupSelected,
  }) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => BlocProvider.value(
          value: context.read<MasterGroupBloc>(),
          child: _MasterGroupSelectionScreen(onGroupSelected: onGroupSelected),
        ),
      ),
    );
  }

  // Navigate to Master Group Sub1 Selection Screen
  static void showMasterGroupSub1SelectionScreen({
    required BuildContext context,
    required Function(MasterGroupSub1Model) onGroupSelected,
  }) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => BlocProvider.value(
          value: context.read<MasterGroupSub1Bloc>(),
          child:
              _MasterGroupSub1SelectionScreen(onGroupSelected: onGroupSelected),
        ),
      ),
    );
  }

  // Navigate to Master Group Sub2 Selection Screen
  static void showMasterGroupSub2SelectionScreen({
    required BuildContext context,
    required Function(MasterGroupSub2Model) onGroupSelected,
  }) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => BlocProvider.value(
          value: context.read<MasterGroupSub2Bloc>(),
          child:
              _MasterGroupSub2SelectionScreen(onGroupSelected: onGroupSelected),
        ),
      ),
    );
  }

  // Navigate to Master Brand Selection Screen
  static void showMasterBrandSelectionScreen({
    required BuildContext context,
    required Function(MasterBrandModel) onItemSelected,
  }) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => BlocProvider.value(
          value: context.read<MasterBrandBloc>(),
          child: _MasterBrandSelectionScreen(onItemSelected: onItemSelected),
        ),
      ),
    );
  }

  // Navigate to Master Category Selection Screen
  static void showMasterCategorySelectionScreen({
    required BuildContext context,
    required Function(MasterCategoryModel) onItemSelected,
  }) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => BlocProvider.value(
          value: context.read<MasterCategoryBloc>(),
          child: _MasterCategorySelectionScreen(onItemSelected: onItemSelected),
        ),
      ),
    );
  }

  // Navigate to Master Class Selection Screen
  static void showMasterClassSelectionScreen({
    required BuildContext context,
    required Function(MasterClassModel) onItemSelected,
  }) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => BlocProvider.value(
          value: context.read<MasterClassBloc>(),
          child: _MasterClassSelectionScreen(onItemSelected: onItemSelected),
        ),
      ),
    );
  }

  // Navigate to Master Design Selection Screen
  static void showMasterDesignSelectionScreen({
    required BuildContext context,
    required Function(MasterDesignModel) onItemSelected,
  }) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => BlocProvider.value(
          value: context.read<MasterDesignBloc>(),
          child: _MasterDesignSelectionScreen(onItemSelected: onItemSelected),
        ),
      ),
    );
  }

  // Navigate to Master Grade Selection Screen
  static void showMasterGradeSelectionScreen({
    required BuildContext context,
    required Function(MasterGradeModel) onItemSelected,
  }) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => BlocProvider.value(
          value: context.read<MasterGradeBloc>(),
          child: _MasterGradeSelectionScreen(onItemSelected: onItemSelected),
        ),
      ),
    );
  }

  // Navigate to Master Model Selection Screen
  static void showMasterModelSelectionScreen({
    required BuildContext context,
    required Function(MasterModelModel) onItemSelected,
  }) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => BlocProvider.value(
          value: context.read<MasterModelBloc>(),
          child: _MasterModelSelectionScreen(onItemSelected: onItemSelected),
        ),
      ),
    );
  }

  // Navigate to Master Pattern Selection Screen
  static void showMasterPatternSelectionScreen({
    required BuildContext context,
    required Function(MasterPatternModel) onItemSelected,
  }) {
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => BlocProvider.value(
          value: context.read<MasterPatternBloc>(),
          child: _MasterPatternSelectionScreen(onItemSelected: onItemSelected),
        ),
      ),
    );
  }
}

class _MasterGroupSelectionScreen extends StatefulWidget {
  final Function(MasterGroupModel) onGroupSelected;

  const _MasterGroupSelectionScreen({required this.onGroupSelected});

  @override
  State<_MasterGroupSelectionScreen> createState() =>
      _MasterGroupSelectionScreenState();
}

class _MasterGroupSelectionScreenState
    extends State<_MasterGroupSelectionScreen> {
  List<MasterGroupModel> listData = [];
  @override
  Widget build(BuildContext context) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('select_group_main')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () => Navigator.of(context).pop(),
        ),
      ),
      body: BlocListener<MasterGroupBloc, MasterGroupState>(
        listener: (context, state) {
          if (state is MasterGroupLoadSuccess) {
            setState(() {
              if (state.groups.isNotEmpty) {
                // Add only new items to avoid duplicates
                final newItems = state.groups
                    .where((item) => !listData.any(
                        (existing) => existing.guidfixed == item.guidfixed))
                    .toList();
                listData.addAll(newItems);
              }
            });
          } else if (state is MasterGroupLoadFailed) {
            global.showSnackBar(
                context,
                const Icon(Icons.error, color: Colors.white),
                state.message,
                Colors.red);
          }
        },
        child: GenericSelectionContent<MasterGroupModel>(
          listData: listData,
          config: SelectionConfig<MasterGroupModel>(
            title: global.language('select_group_main'),
            searchHint: global.language('search'),
            columnHeaders: [
              global.language("group_main_code"),
              global.language("group_main_name"),
            ],
            columnFlexes: [3, 4],
            onItemSelected: (item) {
              widget.onGroupSelected(item);
              Navigator.of(context).pop();
            },
            getCode: (item) => item.code,
            getName: (item) => global.packName(item.names),
            getGuidFixed: (item) => item.guidfixed,
            loadData: (offset, limit, search) {
              context.read<MasterGroupBloc>().add(MasterGroupLoadList(
                    offset: offset,
                    limit: limit,
                    search: search,
                  ));
            },
            getCurrentData: () => listData,
            isLoading: () {
              final state = context.read<MasterGroupBloc>().state;
              return state is MasterGroupInProgress;
            },
            hasError: () {
              final state = context.read<MasterGroupBloc>().state;
              return state is MasterGroupLoadFailed;
            },
            getErrorMessage: () {
              final state = context.read<MasterGroupBloc>().state;
              return state is MasterGroupLoadFailed ? state.message : '';
            },
            onRefresh: () {
              listData.clear();
            },
          ),
        ),
      ),
    );
  }
}

class _MasterGroupSub1SelectionScreen extends StatefulWidget {
  final Function(MasterGroupSub1Model) onGroupSelected;

  const _MasterGroupSub1SelectionScreen({required this.onGroupSelected});

  @override
  State<_MasterGroupSub1SelectionScreen> createState() =>
      _MasterGroupSub1SelectionScreenState();
}

class _MasterGroupSub1SelectionScreenState
    extends State<_MasterGroupSub1SelectionScreen> {
  List<MasterGroupSub1Model> listData = [];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('select_group_sub1')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () => Navigator.of(context).pop(),
        ),
      ),
      body: BlocListener<MasterGroupSub1Bloc, MasterGroupSub1State>(
        listener: (context, state) {
          if (state is MasterGroupSub1LoadSuccess) {
            setState(() {
              if (state.groupSub1s.isNotEmpty) {
                // Add only new items to avoid duplicates
                final newItems = state.groupSub1s
                    .where((item) => !listData.any(
                        (existing) => existing.guidfixed == item.guidfixed))
                    .toList();
                listData.addAll(newItems);
              }
            });
          } else if (state is MasterGroupSub1LoadFailed) {
            global.showSnackBar(
                context,
                const Icon(Icons.error, color: Colors.white),
                state.message,
                Colors.red);
          }
        },
        child: GenericSelectionContent<MasterGroupSub1Model>(
          listData: listData,
          config: SelectionConfig<MasterGroupSub1Model>(
            title: global.language('select_group_sub1'),
            searchHint: global.language('search'),
            columnHeaders: [
              global.language("group_sub1_code"),
              global.language("group_sub1_name"),
              global.language("group_main_name"),
            ],
            columnFlexes: [3, 4, 3],
            onItemSelected: (item) {
              widget.onGroupSelected(item);
              Navigator.of(context).pop();
            },
            getCode: (item) => item.code,
            getName: (item) => global.packName(item.names),
            getGuidFixed: (item) => item.guidfixed,
            getSecondaryInfo: (item) => global.packName(item.groupMainNames),
            loadData: (offset, limit, search) {
              context.read<MasterGroupSub1Bloc>().add(MasterGroupSub1LoadList(
                    offset: offset,
                    limit: limit,
                    search: search,
                  ));
            },
            getCurrentData: () => listData,
            isLoading: () {
              final state = context.read<MasterGroupSub1Bloc>().state;
              return state is MasterGroupSub1InProgress;
            },
            hasError: () {
              final state = context.read<MasterGroupSub1Bloc>().state;
              return state is MasterGroupSub1LoadFailed;
            },
            getErrorMessage: () {
              final state = context.read<MasterGroupSub1Bloc>().state;
              return state is MasterGroupSub1LoadFailed ? state.message : '';
            },
            onRefresh: () {
              listData.clear();
            },
          ),
        ),
      ),
    );
  }
}

class _MasterGroupSub2SelectionScreen extends StatefulWidget {
  final Function(MasterGroupSub2Model) onGroupSelected;

  const _MasterGroupSub2SelectionScreen({required this.onGroupSelected});

  @override
  State<_MasterGroupSub2SelectionScreen> createState() =>
      _MasterGroupSub2SelectionScreenState();
}

class _MasterGroupSub2SelectionScreenState
    extends State<_MasterGroupSub2SelectionScreen> {
  List<MasterGroupSub2Model> listData = [];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('select_group_sub2')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () => Navigator.of(context).pop(),
        ),
      ),
      body: BlocListener<MasterGroupSub2Bloc, MasterGroupSub2State>(
        listener: (context, state) {
          if (state is MasterGroupSub2LoadSuccess) {
            setState(() {
              if (state.groupSub2s.isNotEmpty) {
                // Add only new items to avoid duplicates
                final newItems = state.groupSub2s
                    .where((item) => !listData.any(
                        (existing) => existing.guidfixed == item.guidfixed))
                    .toList();
                listData.addAll(newItems);
              }
            });
          } else if (state is MasterGroupSub2LoadFailed) {
            global.showSnackBar(
                context,
                const Icon(Icons.error, color: Colors.white),
                state.message,
                Colors.red);
          }
        },
        child: GenericSelectionContent<MasterGroupSub2Model>(
          listData: listData,
          config: SelectionConfig<MasterGroupSub2Model>(
            title: global.language('select_group_sub2'),
            searchHint: global.language('search'),
            columnHeaders: [
              global.language("group_sub2_code"),
              global.language("group_sub2_name"),
              global.language("group_sub1_name"),
              global.language("group_main_name"),
            ],
            columnFlexes: [3, 4, 3, 3],
            onItemSelected: (item) {
              widget.onGroupSelected(item);
              Navigator.of(context).pop();
            },
            getCode: (item) => item.code,
            getName: (item) => global.packName(item.names),
            getGuidFixed: (item) => item.guidfixed,
            getSecondaryInfo: (item) => global.packName(item.groupSubNames),
            getTertiaryInfo: (item) => global.packName(item.groupMainNames),
            loadData: (offset, limit, search) {
              context.read<MasterGroupSub2Bloc>().add(MasterGroupSub2LoadList(
                    offset: offset,
                    limit: limit,
                    search: search,
                  ));
            },
            getCurrentData: () => listData,
            isLoading: () {
              final state = context.read<MasterGroupSub2Bloc>().state;
              return state is MasterGroupSub2InProgress;
            },
            hasError: () {
              final state = context.read<MasterGroupSub2Bloc>().state;
              return state is MasterGroupSub2LoadFailed;
            },
            getErrorMessage: () {
              final state = context.read<MasterGroupSub2Bloc>().state;
              return state is MasterGroupSub2LoadFailed ? state.message : '';
            },
            onRefresh: () {
              listData.clear();
            },
          ),
        ),
      ),
    );
  }
}

// Brand Selection Screen
class _MasterBrandSelectionScreen extends StatefulWidget {
  final Function(MasterBrandModel) onItemSelected;

  const _MasterBrandSelectionScreen({required this.onItemSelected});

  @override
  State<_MasterBrandSelectionScreen> createState() =>
      _MasterBrandSelectionScreenState();
}

class _MasterBrandSelectionScreenState
    extends State<_MasterBrandSelectionScreen> {
  List<MasterBrandModel> listData = [];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('select_brand')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () => Navigator.of(context).pop(),
        ),
      ),
      body: BlocListener<MasterBrandBloc, MasterBrandState>(
        listener: (context, state) {
          if (state is MasterBrandLoadSuccess) {
            setState(() {
              if (state.brands.isNotEmpty) {
                final newItems = state.brands
                    .where((item) => !listData.any(
                        (existing) => existing.guidfixed == item.guidfixed))
                    .toList();
                listData.addAll(newItems);
              }
            });
          } else if (state is MasterBrandLoadFailed) {
            global.showSnackBar(
                context,
                const Icon(Icons.error, color: Colors.white),
                state.message,
                Colors.red);
          }
        },
        child: GenericSelectionContent<MasterBrandModel>(
          listData: listData,
          config: SelectionConfig<MasterBrandModel>(
            title: global.language('select_brand'),
            searchHint: global.language('search'),
            columnHeaders: [
              global.language("brand_code"),
              global.language("brand_name"),
            ],
            columnFlexes: [3, 4],
            onItemSelected: (item) {
              widget.onItemSelected(item);
              Navigator.of(context).pop();
            },
            getCode: (item) => item.code,
            getName: (item) => global.packName(item.names),
            getGuidFixed: (item) => item.guidfixed,
            loadData: (offset, limit, search) {
              context.read<MasterBrandBloc>().add(MasterBrandLoadList(
                    offset: offset,
                    limit: limit,
                    search: search,
                  ));
            },
            getCurrentData: () => listData,
            isLoading: () {
              final state = context.read<MasterBrandBloc>().state;
              return state is MasterBrandInProgress;
            },
            hasError: () {
              final state = context.read<MasterBrandBloc>().state;
              return state is MasterBrandLoadFailed;
            },
            getErrorMessage: () {
              final state = context.read<MasterBrandBloc>().state;
              return state is MasterBrandLoadFailed ? state.message : '';
            },
            onRefresh: () {
              listData.clear();
            },
          ),
        ),
      ),
    );
  }
}

// Category Selection Screen
class _MasterCategorySelectionScreen extends StatefulWidget {
  final Function(MasterCategoryModel) onItemSelected;

  const _MasterCategorySelectionScreen({required this.onItemSelected});

  @override
  State<_MasterCategorySelectionScreen> createState() =>
      _MasterCategorySelectionScreenState();
}

class _MasterCategorySelectionScreenState
    extends State<_MasterCategorySelectionScreen> {
  List<MasterCategoryModel> listData = [];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('select_category')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () => Navigator.of(context).pop(),
        ),
      ),
      body: BlocListener<MasterCategoryBloc, MasterCategoryState>(
        listener: (context, state) {
          if (state is MasterCategoryLoadSuccess) {
            setState(() {
              if (state.categories.isNotEmpty) {
                final newItems = state.categories
                    .where((item) => !listData.any(
                        (existing) => existing.guidfixed == item.guidfixed))
                    .toList();
                listData.addAll(newItems);
              }
            });
          } else if (state is MasterCategoryLoadFailed) {
            global.showSnackBar(
                context,
                const Icon(Icons.error, color: Colors.white),
                state.message,
                Colors.red);
          }
        },
        child: GenericSelectionContent<MasterCategoryModel>(
          listData: listData,
          config: SelectionConfig<MasterCategoryModel>(
            title: global.language('select_category'),
            searchHint: global.language('search'),
            columnHeaders: [
              global.language("category_code"),
              global.language("category_name"),
            ],
            columnFlexes: [3, 4],
            onItemSelected: (item) {
              widget.onItemSelected(item);
              Navigator.of(context).pop();
            },
            getCode: (item) => item.code,
            getName: (item) => global.packName(item.names),
            getGuidFixed: (item) => item.guidfixed,
            loadData: (offset, limit, search) {
              context.read<MasterCategoryBloc>().add(MasterCategoryLoadList(
                    offset: offset,
                    limit: limit,
                    search: search,
                  ));
            },
            getCurrentData: () => listData,
            isLoading: () {
              final state = context.read<MasterCategoryBloc>().state;
              return state is MasterCategoryInProgress;
            },
            hasError: () {
              final state = context.read<MasterCategoryBloc>().state;
              return state is MasterCategoryLoadFailed;
            },
            getErrorMessage: () {
              final state = context.read<MasterCategoryBloc>().state;
              return state is MasterCategoryLoadFailed ? state.message : '';
            },
            onRefresh: () {
              listData.clear();
            },
          ),
        ),
      ),
    );
  }
}

// Class Selection Screen
class _MasterClassSelectionScreen extends StatefulWidget {
  final Function(MasterClassModel) onItemSelected;

  const _MasterClassSelectionScreen({required this.onItemSelected});

  @override
  State<_MasterClassSelectionScreen> createState() =>
      _MasterClassSelectionScreenState();
}

class _MasterClassSelectionScreenState
    extends State<_MasterClassSelectionScreen> {
  List<MasterClassModel> listData = [];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('select_class')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () => Navigator.of(context).pop(),
        ),
      ),
      body: BlocListener<MasterClassBloc, MasterClassState>(
        listener: (context, state) {
          if (state is MasterClassLoadSuccess) {
            setState(() {
              if (state.classes.isNotEmpty) {
                final newItems = state.classes
                    .where((item) => !listData.any(
                        (existing) => existing.guidfixed == item.guidfixed))
                    .toList();
                listData.addAll(newItems);
              }
            });
          } else if (state is MasterClassLoadFailed) {
            global.showSnackBar(
                context,
                const Icon(Icons.error, color: Colors.white),
                state.message,
                Colors.red);
          }
        },
        child: GenericSelectionContent<MasterClassModel>(
          listData: listData,
          config: SelectionConfig<MasterClassModel>(
            title: global.language('select_class'),
            searchHint: global.language('search'),
            columnHeaders: [
              global.language("class_code"),
              global.language("class_name"),
            ],
            columnFlexes: [3, 4],
            onItemSelected: (item) {
              widget.onItemSelected(item);
              Navigator.of(context).pop();
            },
            getCode: (item) => item.code,
            getName: (item) => global.packName(item.names),
            getGuidFixed: (item) => item.guidfixed,
            loadData: (offset, limit, search) {
              context.read<MasterClassBloc>().add(MasterClassLoadList(
                    offset: offset,
                    limit: limit,
                    search: search,
                  ));
            },
            getCurrentData: () => listData,
            isLoading: () {
              final state = context.read<MasterClassBloc>().state;
              return state is MasterClassInProgress;
            },
            hasError: () {
              final state = context.read<MasterClassBloc>().state;
              return state is MasterClassLoadFailed;
            },
            getErrorMessage: () {
              final state = context.read<MasterClassBloc>().state;
              return state is MasterClassLoadFailed ? state.message : '';
            },
            onRefresh: () {
              listData.clear();
            },
          ),
        ),
      ),
    );
  }
}

// Design Selection Screen
class _MasterDesignSelectionScreen extends StatefulWidget {
  final Function(MasterDesignModel) onItemSelected;

  const _MasterDesignSelectionScreen({required this.onItemSelected});

  @override
  State<_MasterDesignSelectionScreen> createState() =>
      _MasterDesignSelectionScreenState();
}

class _MasterDesignSelectionScreenState
    extends State<_MasterDesignSelectionScreen> {
  List<MasterDesignModel> listData = [];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('select_design')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () => Navigator.of(context).pop(),
        ),
      ),
      body: BlocListener<MasterDesignBloc, MasterDesignState>(
        listener: (context, state) {
          if (state is MasterDesignLoadSuccess) {
            setState(() {
              if (state.designs.isNotEmpty) {
                final newItems = state.designs
                    .where((item) => !listData.any(
                        (existing) => existing.guidfixed == item.guidfixed))
                    .toList();
                listData.addAll(newItems);
              }
            });
          } else if (state is MasterDesignLoadFailed) {
            global.showSnackBar(
                context,
                const Icon(Icons.error, color: Colors.white),
                state.message,
                Colors.red);
          }
        },
        child: GenericSelectionContent<MasterDesignModel>(
          listData: listData,
          config: SelectionConfig<MasterDesignModel>(
            title: global.language('select_design'),
            searchHint: global.language('search'),
            columnHeaders: [
              global.language("design_code"),
              global.language("design_name"),
            ],
            columnFlexes: [3, 4],
            onItemSelected: (item) {
              widget.onItemSelected(item);
              Navigator.of(context).pop();
            },
            getCode: (item) => item.code,
            getName: (item) => global.packName(item.names),
            getGuidFixed: (item) => item.guidfixed,
            loadData: (offset, limit, search) {
              context.read<MasterDesignBloc>().add(MasterDesignLoadList(
                    offset: offset,
                    limit: limit,
                    search: search,
                  ));
            },
            getCurrentData: () => listData,
            isLoading: () {
              final state = context.read<MasterDesignBloc>().state;
              return state is MasterDesignInProgress;
            },
            hasError: () {
              final state = context.read<MasterDesignBloc>().state;
              return state is MasterDesignLoadFailed;
            },
            getErrorMessage: () {
              final state = context.read<MasterDesignBloc>().state;
              return state is MasterDesignLoadFailed ? state.message : '';
            },
            onRefresh: () {
              listData.clear();
            },
          ),
        ),
      ),
    );
  }
}

// Grade Selection Screen
class _MasterGradeSelectionScreen extends StatefulWidget {
  final Function(MasterGradeModel) onItemSelected;

  const _MasterGradeSelectionScreen({required this.onItemSelected});

  @override
  State<_MasterGradeSelectionScreen> createState() =>
      _MasterGradeSelectionScreenState();
}

class _MasterGradeSelectionScreenState
    extends State<_MasterGradeSelectionScreen> {
  List<MasterGradeModel> listData = [];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('select_grade')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () => Navigator.of(context).pop(),
        ),
      ),
      body: BlocListener<MasterGradeBloc, MasterGradeState>(
        listener: (context, state) {
          if (state is MasterGradeLoadSuccess) {
            setState(() {
              if (state.grades.isNotEmpty) {
                final newItems = state.grades
                    .where((item) => !listData.any(
                        (existing) => existing.guidfixed == item.guidfixed))
                    .toList();
                listData.addAll(newItems);
              }
            });
          } else if (state is MasterGradeLoadFailed) {
            global.showSnackBar(
                context,
                const Icon(Icons.error, color: Colors.white),
                state.message,
                Colors.red);
          }
        },
        child: GenericSelectionContent<MasterGradeModel>(
          listData: listData,
          config: SelectionConfig<MasterGradeModel>(
            title: global.language('select_grade'),
            searchHint: global.language('search'),
            columnHeaders: [
              global.language("grade_code"),
              global.language("grade_name"),
            ],
            columnFlexes: [3, 4],
            onItemSelected: (item) {
              widget.onItemSelected(item);
              Navigator.of(context).pop();
            },
            getCode: (item) => item.code,
            getName: (item) => global.packName(item.names),
            getGuidFixed: (item) => item.guidfixed,
            loadData: (offset, limit, search) {
              context.read<MasterGradeBloc>().add(MasterGradeLoadList(
                    offset: offset,
                    limit: limit,
                    search: search,
                  ));
            },
            getCurrentData: () => listData,
            isLoading: () {
              final state = context.read<MasterGradeBloc>().state;
              return state is MasterGradeInProgress;
            },
            hasError: () {
              final state = context.read<MasterGradeBloc>().state;
              return state is MasterGradeLoadFailed;
            },
            getErrorMessage: () {
              final state = context.read<MasterGradeBloc>().state;
              return state is MasterGradeLoadFailed ? state.message : '';
            },
            onRefresh: () {
              listData.clear();
            },
          ),
        ),
      ),
    );
  }
}

// Model Selection Screen
class _MasterModelSelectionScreen extends StatefulWidget {
  final Function(MasterModelModel) onItemSelected;

  const _MasterModelSelectionScreen({required this.onItemSelected});

  @override
  State<_MasterModelSelectionScreen> createState() =>
      _MasterModelSelectionScreenState();
}

class _MasterModelSelectionScreenState
    extends State<_MasterModelSelectionScreen> {
  List<MasterModelModel> listData = [];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('select_model')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () => Navigator.of(context).pop(),
        ),
      ),
      body: BlocListener<MasterModelBloc, MasterModelState>(
        listener: (context, state) {
          if (state is MasterModelLoadSuccess) {
            setState(() {
              if (state.models.isNotEmpty) {
                final newItems = state.models
                    .where((item) => !listData.any(
                        (existing) => existing.guidfixed == item.guidfixed))
                    .toList();
                listData.addAll(newItems);
              }
            });
          } else if (state is MasterModelLoadFailed) {
            global.showSnackBar(
                context,
                const Icon(Icons.error, color: Colors.white),
                state.message,
                Colors.red);
          }
        },
        child: GenericSelectionContent<MasterModelModel>(
          listData: listData,
          config: SelectionConfig<MasterModelModel>(
            title: global.language('select_model'),
            searchHint: global.language('search'),
            columnHeaders: [
              global.language("model_code"),
              global.language("model_name"),
            ],
            columnFlexes: [3, 4],
            onItemSelected: (item) {
              widget.onItemSelected(item);
              Navigator.of(context).pop();
            },
            getCode: (item) => item.code,
            getName: (item) => global.packName(item.names),
            getGuidFixed: (item) => item.guidfixed,
            loadData: (offset, limit, search) {
              context.read<MasterModelBloc>().add(MasterModelLoadList(
                    offset: offset,
                    limit: limit,
                    search: search,
                  ));
            },
            getCurrentData: () => listData,
            isLoading: () {
              final state = context.read<MasterModelBloc>().state;
              return state is MasterModelInProgress;
            },
            hasError: () {
              final state = context.read<MasterModelBloc>().state;
              return state is MasterModelLoadFailed;
            },
            getErrorMessage: () {
              final state = context.read<MasterModelBloc>().state;
              return state is MasterModelLoadFailed ? state.message : '';
            },
            onRefresh: () {
              listData.clear();
            },
          ),
        ),
      ),
    );
  }
}

// Pattern Selection Screen
class _MasterPatternSelectionScreen extends StatefulWidget {
  final Function(MasterPatternModel) onItemSelected;

  const _MasterPatternSelectionScreen({required this.onItemSelected});

  @override
  State<_MasterPatternSelectionScreen> createState() =>
      _MasterPatternSelectionScreenState();
}

class _MasterPatternSelectionScreenState
    extends State<_MasterPatternSelectionScreen> {
  List<MasterPatternModel> listData = [];

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('select_pattern')),
        leading: IconButton(
          focusNode: FocusNode(skipTraversal: true),
          icon: const Icon(Icons.arrow_back),
          onPressed: () => Navigator.of(context).pop(),
        ),
      ),
      body: BlocListener<MasterPatternBloc, MasterPatternState>(
        listener: (context, state) {
          if (state is MasterPatternLoadSuccess) {
            setState(() {
              if (state.patterns.isNotEmpty) {
                final newItems = state.patterns
                    .where((item) => !listData.any(
                        (existing) => existing.guidfixed == item.guidfixed))
                    .toList();
                listData.addAll(newItems);
              }
            });
          } else if (state is MasterPatternLoadFailed) {
            global.showSnackBar(
                context,
                const Icon(Icons.error, color: Colors.white),
                state.message,
                Colors.red);
          }
        },
        child: GenericSelectionContent<MasterPatternModel>(
          listData: listData,
          config: SelectionConfig<MasterPatternModel>(
            title: global.language('select_pattern'),
            searchHint: global.language('search'),
            columnHeaders: [
              global.language("pattern_code"),
              global.language("pattern_name"),
            ],
            columnFlexes: [3, 4],
            onItemSelected: (item) {
              widget.onItemSelected(item);
              Navigator.of(context).pop();
            },
            getCode: (item) => item.code,
            getName: (item) => global.packName(item.names),
            getGuidFixed: (item) => item.guidfixed,
            loadData: (offset, limit, search) {
              context.read<MasterPatternBloc>().add(MasterPatternLoadList(
                    offset: offset,
                    limit: limit,
                    search: search,
                  ));
            },
            getCurrentData: () => listData,
            isLoading: () {
              final state = context.read<MasterPatternBloc>().state;
              return state is MasterPatternInProgress;
            },
            hasError: () {
              final state = context.read<MasterPatternBloc>().state;
              return state is MasterPatternLoadFailed;
            },
            getErrorMessage: () {
              final state = context.read<MasterPatternBloc>().state;
              return state is MasterPatternLoadFailed ? state.message : '';
            },
            onRefresh: () {
              listData.clear();
            },
          ),
        ),
      ),
    );
  }
}
