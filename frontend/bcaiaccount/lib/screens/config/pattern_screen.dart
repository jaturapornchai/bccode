import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/public_name_model.dart';

class PatternDetailModel {
  String code = "";
  List<PublicNameModel> names = [];
  List<TextEditingController> editController = [];
  List<FocusNode> focus = [];
}

class PatternScreen extends StatefulWidget {
  const PatternScreen({super.key});

  @override
  State<PatternScreen> createState() => PatternScreenState();
}

class PatternScreenState extends State<PatternScreen>
    with SingleTickerProviderStateMixin, global.ThemeRefreshMixin {
  late TabController tabController;
  ScrollController editScrollController = ScrollController();
  TextEditingController codeController = TextEditingController();
  List<TextEditingController> nameController = [];
  List<String> fieldName = [];
  List<FocusNode> nameFocusNode = [];
  List<PatternDetailModel> details = [];

  void setSystemLanguageList() async {
    await global.setSystemLanguage(context);

    for (int i = 0; i < global.config.languages.length; i++) {
      nameController.add(TextEditingController());
      nameFocusNode.add(FocusNode());
      fieldName.add("name${i + 1}");
    }
  }

  @override
  void initState() {
    setSystemLanguageList();
    tabController = TabController(vsync: this, length: 2);

    super.initState();
  }

  @override
  void dispose() {
    tabController.dispose();
    editScrollController.dispose();
    codeController.dispose();
    for (int i = 0; i < nameController.length; i++) {
      nameController[i].dispose();
      nameFocusNode[i].dispose();
    }
    super.dispose();
  }

  Widget listScreen() {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        title: Text(global.language("product_type")),
        actions: <Widget>[
          IconButton(
            onPressed: () {
              setState(() {});
            },
            icon: Icon(Icons.delete),
          ),
          IconButton(
            onPressed: () {
              tabController.index = 1;
              nameFocusNode[0].requestFocus();
            },
            icon: Icon(Icons.add),
          ),
        ],
      ),
      body: Container(),
    );
  }

  void saveData() {
    global.showInfoSnackBar(context, global.language("save_success"));
  }

  void addDetail() {
    setState(() {
      PatternDetailModel detail = PatternDetailModel();
      for (int i = 0; i < global.config.languages.length; i++) {
        detail.editController.add(TextEditingController());
        detail.names.add(PublicNameModel());
        detail.focus.add(FocusNode());
      }
      details.add(detail);
      detail.focus[0].requestFocus();
    });
  }

  Widget editScreen({showBackButton = true}) {
    return Scaffold(
      appBar: AppBar(
        leading: showBackButton
            ? IconButton(
                icon: Icon(Icons.arrow_back),
                onPressed: () {
                  tabController.index = 0;
                },
              )
            : null,
        backgroundColor: global.theme.appBarColor,
        title: Text(global.language("product_type")),
        actions: <Widget>[
          IconButton(
            onPressed: () {
              addDetail();
            },
            icon: Icon(Icons.add),
          ),
          IconButton(
            onPressed: () {
              saveData();
            },
            icon: Icon(Icons.save),
          ),
        ],
      ),
      body: Column(
        children: [
          Container(
            width: double.infinity,
            padding: EdgeInsets.all(10),
            child: Column(
              children: [
                for (int i = 0; i < global.config.languages.length; i++)
                  if (global.config.languages[i].isuse!)
                    Padding(
                      padding: EdgeInsets.only(bottom: 10),
                      child: TextField(
                        focusNode: nameFocusNode[i],
                        textAlign: TextAlign.left,
                        controller: nameController[i],
                        decoration: InputDecoration(
                          border: const OutlineInputBorder(),
                          labelText:
                              "${global.language("namepdtype")} (${global.config.languages[i].name})",
                        ),
                      ),
                    ),
                Wrap(
                  spacing: 4.0,
                  runSpacing: 4.0,
                  children: [
                    for (var i = 0; i < details.length; i++)
                      ElevatedButton(
                        onPressed: () {
                          details[i].focus[0].requestFocus();
                        },
                        child: Text(details[i].names[0].name),
                      ),
                  ],
                ),
                const SizedBox(height: 10),
              ],
            ),
          ),
          Expanded(
            child: SingleChildScrollView(
              controller: editScrollController,
              child: Column(
                children: [
                  for (var i = 0; i < details.length; i++)
                    Container(
                      padding: EdgeInsets.all(10),
                      decoration: BoxDecoration(
                        color: global.theme.cardColor,
                        border: Border.all(color: global.theme.dividerBorderColor),
                        borderRadius: BorderRadius.circular(4),
                      ),
                      child: Column(
                        children: [
                          for (
                            int j = 0;
                            j < global.config.languages.length;
                            j++
                          )
                            if (global.config.languages[j].isuse!)
                              Padding(
                                padding: EdgeInsets.only(bottom: 10),
                                child: TextField(
                                  focusNode: details[i].focus[j],
                                  onChanged: (value) {
                                    setState(() {
                                      details[i].names[j].name = value;
                                    });
                                  },
                                  textAlign: TextAlign.left,
                                  controller: details[i].editController[j],
                                  decoration: InputDecoration(
                                    border: const OutlineInputBorder(),
                                    labelText:
                                        "${global.language("sub_pattern_name")} (${global.config.languages[j].name})",
                                  ),
                                ),
                              ),
                        ],
                      ),
                    ),
                  SizedBox(height: 10),
                  SizedBox(
                    width: double.infinity,
                    child: Row(
                      children: [
                        ElevatedButton.icon(
                          onPressed: () {
                            addDetail();
                          },
                          icon: Icon(Icons.add),
                          label: Text(global.language("add_sub_item")),
                        ),
                        ElevatedButton.icon(
                          onPressed: () {
                            saveData();
                          },
                          icon: Icon(Icons.save),
                          label: Text(global.language("save")),
                        ),
                      ],
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

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        return Scaffold(
          body: (constraints.maxWidth > 800)
              ? Row(
                  children: [
                    Expanded(child: listScreen()),
                    Container(width: 1.5, color: global.theme.textColor),
                    Expanded(child: editScreen(showBackButton: false)),
                  ],
                )
              : TabBarView(
                  physics: const NeverScrollableScrollPhysics(),
                  controller: tabController,
                  children: [listScreen(), editScreen()],
                ),
        );
      },
    );
  }
}
