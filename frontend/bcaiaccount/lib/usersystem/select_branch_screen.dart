import 'package:auto_size_text/auto_size_text.dart';
import 'package:flutter/foundation.dart';
import 'package:smlaicloud/bloc/company_branch/company_branch_bloc.dart';
import 'package:smlaicloud/global.dart';
import 'package:smlaicloud/menu_screen.dart';
import 'package:smlaicloud/model/company_branch_model.dart';
import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:flutter_bloc/flutter_bloc.dart';

class SelectBranchScreen extends StatefulWidget {
  const SelectBranchScreen({super.key});

  @override
  SelectBranchScreenState createState() => SelectBranchScreenState();
}

class SelectBranchScreenState extends State<SelectBranchScreen> {
  List<CompanyBranchModel> listData = [];

  @override
  void initState() {
    if (kDebugMode) {
      AppLogger.debug("[SelectBranchScreen] initState");
    }
    loadDataList();
    super.initState();
  }

  void loadDataList() {
    context
        .read<CompanyBranchBloc>()
        .add(const CompanyBranchLoadList(offset: 0, limit: 100, search: ""));
  }

  Widget menuWidget(
      {required String guidfixed,
      required String label,
      Color color = Colors.white,
      icon,
      required Function callback}) {
    Widget textWidget = Center(
        child: AutoSizeText(label,
            maxLines: 3,
            textAlign: TextAlign.center,
            style: const TextStyle(fontSize: 20, fontWeight: FontWeight.bold)));
    return ElevatedButton(
      style: ElevatedButton.styleFrom(
        padding: const EdgeInsets.all(4),
        foregroundColor: Colors.black,
        backgroundColor: color,
        shape: const RoundedRectangleBorder(
          borderRadius: BorderRadius.all(Radius.circular(4)),
        ),
      ),
      onPressed: (guidfixed != global.getBranchGuidFixed())
          ? () {
              callback();
            }
          : null,
      child: (icon == null)
          ? textWidget
          : Stack(
              children: [
                textWidget,
                Positioned(
                    right: 4, top: 8, child: Icon(icon as IconData, size: 25)),
              ],
            ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return MultiBlocListener(
      listeners: [
        BlocListener<CompanyBranchBloc, CompanyBranchState>(
          listener: (context, state) {
            if (state is CompanyBranchLoadSuccess) {
              setState(() {
                listData = state.companyBranch;
              });
            }
            if (state is CompanyBranchLoadFailed) {
              setState(() {
                global.showSnackBar(
                  context,
                  const Icon(
                    Icons.save,
                    color: Colors.white,
                  ),
                  "${global.language('load_data_failed')} : ${state.message}",
                  Colors.red,
                );
              });
            }
          },
        ),
      ],
      child: Scaffold(
        resizeToAvoidBottomInset: true,
        appBar: AppBar(
          title: Text(global.language("select_branch")),
          backgroundColor: global.theme.appBarColor,
        ),
        body: SingleChildScrollView(
            child: Padding(
          padding: const EdgeInsets.all(10),
          child: GridView.builder(
              padding: const EdgeInsets.all(0),
              shrinkWrap: true,
              physics: const NeverScrollableScrollPhysics(),
              gridDelegate: const SliverGridDelegateWithMaxCrossAxisExtent(
                  maxCrossAxisExtent: 150,
                  crossAxisSpacing: 10,
                  mainAxisSpacing: 10),
              itemCount: listData.length,
              itemBuilder: (BuildContext ctx, index) {
                final branch = listData[index];
                return menuWidget(
                    guidfixed: branch.guidfixed,
                    label: global.packName(branch.names),
                    color: Colors.blue.shade100,
                    icon: Icons.home_work_rounded,
                    callback: () {
                      if (branch.guidfixed != global.getBranchGuidFixed()) {
                        appConfig.setString(
                            "branch_guidfixed", branch.guidfixed);
                        appConfig.setInt("branch_total", listData.length);
                        global.companyBranchSelectData = branch;

                        Navigator.of(context).pushAndRemoveUntil(
                            MaterialPageRoute(
                                builder: (_) => const MenuScreen()),
                            (route) => false);
                      }
                    });
              }),
        )),
      ),
    );
  }
}
