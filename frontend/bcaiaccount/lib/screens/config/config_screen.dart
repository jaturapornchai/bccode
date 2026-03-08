import 'package:smlaicloud/bloc/shop/shop_bloc.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/shop_model.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/widgets/manual_button.dart';

class ConfigScreen extends StatefulWidget {
  const ConfigScreen({super.key});

  @override
  State<ConfigScreen> createState() => ConfigScreenState();
}

class ConfigScreenState extends State<ConfigScreen>
    with SingleTickerProviderStateMixin {
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

  late ShopModel screenData;
  TextEditingController vatrateController = TextEditingController();

  @override
  void initState() {
    screenData = ShopModel();
    loadDataList();
    super.initState();
  }

  void loadDataList() {
    context.read<ShopBloc>().add(GetShopInfo(shopid: global.getShopId()));
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      resizeToAvoidBottomInset: true,
      body: BlocListener<ShopBloc, ShopState>(
        listener: (context, state) {
          if (state is GetShopInfoSuccess) {
            setState(() {
              if (state.shop.guidfixed!.isNotEmpty) {
                screenData = state.shop;
                vatrateController.text = screenData.settings!.vatrate
                    .toString();
              }
              // global.showSnackBar(
              //   context,
              //   const Icon(
              //     Icons.save,
              //     color: Colors.white,
              //   ),
              //   global.language("load_success"),
              //   Colors.green,
              // );
            });
          }
          if (state is ShopUpdateSuccess) {
            global.showSnackBar(
              context,
              Icon(Icons.save, color: Colors.white),
              global.language("save_success"),
              Colors.green,
            );

            global.gotoMainMenu(context);
          }

          if (state is ShopUpdateFailed) {
            setState(() {
              global.showSnackBar(
                context,
                Icon(Icons.save, color: Colors.white),
                "${global.language("save_failed")} ${state.message}",
                Colors.red,
              );
            });
          }
        },
        child: LayoutBuilder(
          builder: (context, constraints) {
            return Scaffold(
              appBar: AppBar(
                backgroundColor: global.theme.appBarColor,
                automaticallyImplyLeading: false,
                title: Text(global.language('system_config')),
                leading: IconButton(
                  focusNode: FocusNode(skipTraversal: true),
                  icon: const Icon(Icons.arrow_back),
                  onPressed: () {
                    global.gotoMainMenu(context);
                  },
                ),
                actions: <Widget>[
                  const ManualButton(),
                  Padding(
                    padding: const EdgeInsets.only(right: 20.0),
                    child: IconButton(
                      focusNode: FocusNode(skipTraversal: true),
                      onPressed: () {
                        saveOrUpdateData();
                      },
                      icon: const Icon(Icons.save, size: 26.0),
                    ),
                  ),
                ],
              ),
              body: SingleChildScrollView(
                child: Center(
                  child: Column(
                    children: [
                      Container(
                        margin: const EdgeInsets.only(top: 10, bottom: 10),
                        padding: const EdgeInsets.all(10),
                        width: 600,
                        decoration: BoxDecoration(
                          border: Border.all(color: Colors.blue),
                          borderRadius: BorderRadius.circular(5),
                          color: Colors.white,
                          boxShadow: [
                            BoxShadow(
                              color: Colors.grey.withValues(alpha: 0.5),
                              spreadRadius: 1,
                              blurRadius: 5,
                              offset: const Offset(
                                0,
                                3,
                              ), // changes position of shadow
                            ),
                          ],
                        ),
                        child: Column(
                          children: [
                            Text(
                              global.language('select_data_language'),
                              style: const TextStyle(
                                fontSize: 20,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                            const SizedBox(height: 10),
                            for (
                              var i = 0;
                              i < screenData.settings!.languageconfigs!.length;
                              i++
                            )
                              Container(
                                margin: const EdgeInsets.only(bottom: 10),
                                child: Row(
                                  children: [
                                    Expanded(
                                      child: Row(
                                        children: [
                                          SizedBox(
                                            width: 30,
                                            child: Text((i + 1).toString()),
                                          ),
                                          Expanded(
                                            child: ElevatedButton(
                                              onPressed: () {
                                                List<LanguageModel>
                                                languagesSelectList = [];
                                                languagesSelectList.addAll(
                                                  defaultlanguageList,
                                                );
                                                for (var selected
                                                    in screenData
                                                        .settings!
                                                        .languageconfigs!) {
                                                  languagesSelectList
                                                      .removeWhere(
                                                        (element) =>
                                                            element.code ==
                                                            selected.code,
                                                      );
                                                }

                                                showDialog(
                                                  context: context,
                                                  builder: (BuildContext context) {
                                                    return AlertDialog(
                                                      title: Text(
                                                        global.language(
                                                          'select_language',
                                                        ),
                                                      ),
                                                      content: SizedBox(
                                                        width: 300,
                                                        height: 400,
                                                        child: ListView.builder(
                                                          itemCount:
                                                              languagesSelectList
                                                                  .length,
                                                          itemBuilder: (context, index) {
                                                            return ListTile(
                                                              title: Row(
                                                                children: [
                                                                  Text(
                                                                    languagesSelectList[index]
                                                                        .name!,
                                                                  ),
                                                                  const Spacer(),
                                                                  Image.asset(
                                                                    'assets/flags/${languagesSelectList[index].code}.png',
                                                                    width: 30,
                                                                    height: 30,
                                                                  ),
                                                                ],
                                                              ),
                                                              onTap: () {
                                                                setState(() {
                                                                  Navigator.of(
                                                                    context,
                                                                  ).pop(
                                                                    languagesSelectList[index]
                                                                        .code,
                                                                  );
                                                                });
                                                              },
                                                            );
                                                          },
                                                        ),
                                                      ),
                                                    );
                                                  },
                                                ).then((value) {
                                                  setState(() {
                                                    if (value != null) {
                                                      screenData
                                                          .settings!
                                                          .languageconfigs![i]
                                                          .code = value
                                                          .toString();
                                                      screenData
                                                          .settings!
                                                          .languageconfigs![i]
                                                          .codeTranslator = value
                                                          .toString();

                                                      screenData
                                                          .settings!
                                                          .languageconfigs![i]
                                                          .name = defaultlanguageList
                                                          .firstWhere(
                                                            (element) =>
                                                                element.code ==
                                                                value
                                                                    .toString(),
                                                          )
                                                          .name;
                                                      screenData
                                                              .settings!
                                                              .languageconfigs![i]
                                                              .isuse =
                                                          true;
                                                    }
                                                  });
                                                });
                                              },
                                              child: Row(
                                                children: [
                                                  (screenData
                                                          .settings!
                                                          .languageconfigs![i]
                                                          .code!
                                                          .isNotEmpty)
                                                      ? Image.asset(
                                                          'assets/flags/${screenData.settings!.languageconfigs![i].code}.png',
                                                          width: 30,
                                                          height: 30,
                                                        )
                                                      : Container(),
                                                  const SizedBox(width: 10),
                                                  (screenData
                                                          .settings!
                                                          .languageconfigs![i]
                                                          .code!
                                                          .isNotEmpty)
                                                      ? Text(
                                                          defaultlanguageList
                                                              .firstWhere(
                                                                (element) =>
                                                                    element
                                                                        .code ==
                                                                    screenData
                                                                        .settings!
                                                                        .languageconfigs![i]
                                                                        .code,
                                                              )
                                                              .name!,
                                                        )
                                                      : Text(
                                                          global.language(
                                                            'select_language',
                                                          ),
                                                        ),
                                                ],
                                              ),
                                            ),
                                          ),
                                        ],
                                      ),
                                    ),
                                    (i == 0)
                                        ? const SizedBox(width: 50)
                                        : SizedBox(
                                            width: 50,
                                            child: IconButton(
                                              onPressed: () {
                                                setState(() {
                                                  screenData
                                                      .settings!
                                                      .languageconfigs!
                                                      .removeAt(i);
                                                });
                                              },
                                              color: Colors.red,
                                              icon: const Icon(Icons.delete),
                                            ),
                                          ),
                                  ],
                                ),
                              ),
                            SizedBox(height: 10),
                            SizedBox(
                              width: double.infinity,
                              child: ElevatedButton(
                                onPressed: () {
                                  setState(() {
                                    screenData.settings!.languageconfigs!.add(
                                      LanguageModel(
                                        code: "",
                                        codeTranslator: "",
                                        name: "",
                                        isuse: false,
                                      ),
                                    );
                                  });
                                },
                                child: Text(global.language('add_language')),
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
          },
        ),
      ),
    );
  }

  void saveOrUpdateData() {
    screenData.settings!.languageconfigs![0].isdefault = true;
    context.read<ShopBloc>().add(
      ShopUpdate(shopid: screenData.guidfixed!, shopdata: screenData),
    );
  }
}
