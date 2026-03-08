import 'dart:async';
import 'dart:convert';
import 'dart:io';

import 'package:barcode_widget/barcode_widget.dart';
import 'package:smlaicloud/model/form_design_struct.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:flutter/services.dart';
import 'package:flutter_dropzone/flutter_dropzone.dart';
import 'package:image_picker/image_picker.dart';
import 'package:qr_flutter/qr_flutter.dart';

import 'package:split_view/split_view.dart';

class BillDesignScreen extends StatefulWidget {
  const BillDesignScreen({super.key});
  @override
  _BillDesignScreenState createState() => _BillDesignScreenState();
}

class _BillDesignScreenState extends State<BillDesignScreen>
    with TickerProviderStateMixin {
  late Timer findTerminalTimer;
  late SplitViewController splitViewController;
  late TabController tabController;
  late TabController editTabController;
  late FormDesignStruct formDesignObj;
  List<TextEditingController> headerController = [];
  List<TextEditingController> footerController = [];
  List<TextEditingController> footerValueController = [];
  List<TextEditingController> headerValueController = [];
  List<LanguageModel> languageList = [];
  final ImagePicker imagePicker = ImagePicker();
  late DropzoneViewController dropZoneController;
  List<File> imageFile = [];
  List<Uint8List> imageWeb = [];
  List<File> imageHeaderFile = [];
  List<Uint8List> imageHeaderWeb = [];
  List<FormDesignHeaderModel> header = [
    FormDesignHeaderModel(
      description: [
        LanguageDataModel(code: "th", name: "สวัสดีปีใหม่ 2564"),
        LanguageDataModel(code: "en", name: "Happy New Year 2021"),
      ],
    ),
  ];
  List<FormDesignColumnModel> detailColumn = [
    FormDesignColumnModel(
      command:
          "&item_qty& &item_name&/&item_unit_name& &item_price_and_symbol& &item_discount&",
      header_names: [
        LanguageDataModel(code: "th", name: "รายละเอียด"),
        LanguageDataModel(code: "en", name: "Description"),
      ],
      width: 5,
    ),
    FormDesignColumnModel(
      command: "&item_total_amount&",
      font_size: 12,
      font_weight_bold: false,
      font_style_italic: false,
      decoration_underline: false,
      text_align: global.PrintColumnAlign.right,
      header_names: [
        LanguageDataModel(code: "th", name: "รวม"),
        LanguageDataModel(code: "en", name: "Amount"),
      ],
      width: 2,
    ),
  ];
  List<FormDesignColumnModel> detailExtraColumn = [
    FormDesignColumnModel(
      command: " + &item_extra_name& &item_extra_qty& &item_extra_unit_name&",
      width: 5,
    ),
    FormDesignColumnModel(
      command: "&item_extra_price&",
      text_align: global.PrintColumnAlign.right,
      width: 1,
    ),
    FormDesignColumnModel(
      command: "&item_extra_total_amount&",
      text_align: global.PrintColumnAlign.right,
      width: 2,
    ),
  ];
  List<List<FormDesignColumnModel>> detailTotalColumn = [
    [
      FormDesignColumnModel(command: "&item_name&", width: 5),
      FormDesignColumnModel(
        command: "&total_piece&",
        text_align: global.PrintColumnAlign.right,
        width: 1,
      ),
      FormDesignColumnModel(
        command: "&total_amount&",
        text_align: global.PrintColumnAlign.right,
        width: 2,
      ),
    ],
  ];
  List<FormDesignFooterModel> footer = [
    FormDesignFooterModel(
      description: [
        LanguageDataModel(code: "th", name: "ขอบคุณที่ใช้บริการ"),
        LanguageDataModel(code: "en", name: "Thank you for using the service."),
      ],
      print_qr_doc_no: true,
    ),
  ];

  @override
  void initState() {
    splitViewController = SplitViewController(
      limits: [null, WeightLimit(min: 0.1, max: 0.9)],
    );
    splitViewController.weights = [0.25, 0.75];
    formDesignObj = FormDesignStruct(
      type: 0,
      guid_fixed: "",
      code: "",
      names_json: jsonEncode(<LanguageDataModel>[
        LanguageDataModel(code: "th", name: "ใบเสร็จรับเงิน"),
        LanguageDataModel(code: "en", name: "Receipt"),
      ]),
      header_json: jsonEncode(header),
      detail_json: jsonEncode(detailColumn),
      detail_total_json: jsonEncode(detailTotalColumn),
      detail_extra_json: jsonEncode(detailExtraColumn),
      detail_footer_json: "{}",
      footer_json: jsonEncode(footer),
    );

    setSystemLanguageList();
    super.initState();
    tabController = TabController(length: 2, vsync: this);
    editTabController = TabController(length: 4, vsync: this);
  }

  void setSystemLanguageList() async {
    await global.setSystemLanguage(context);

    for (int i = 0; i < global.config.languages.length; i++) {
      if (global.config.languages[i].isuse!) {
        languageList.add(global.config.languages[i]);
      }
    }

    setState(() {});
  }

  @override
  void dispose() {
    splitViewController.dispose();
    tabController.dispose();
    super.dispose();
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
        headerController.add(TextEditingController());
        headerController[languageIndex].text = nameObj.name;
        final textLength = headerController[languageIndex].text.length;
        headerController[languageIndex].selection = TextSelection(
          baseOffset: textLength,
          extentOffset: textLength,
        );
        forms.add(
          Expanded(
            child: Padding(
              padding: EdgeInsets.only(left: 10, right: 10, bottom: 10),
              child: TextFormField(
                onChanged: (value) {
                  nameObj.name = value;
                  setState(() {});
                },
                textAlign: TextAlign.left,
                controller: headerController[languageIndex],
                decoration: InputDecoration(
                  floatingLabelBehavior: FloatingLabelBehavior.always,
                  border: const OutlineInputBorder(),
                  labelText:
                      "${global.language(fieldname)} (${getLangName(nameObj.code)})",
                ),
                validator: (value) {
                  if (languageIndex == 0) {
                    if (value == null || value.isEmpty) {
                      return global.language('this_field_is_required');
                    }
                  }

                  return null;
                },
              ),
            ),
          ),
        );
      }
    }

    return forms;
  }

  String getLangName(String? code) {
    LanguageModel name = languageList.firstWhere(
      (element) => element.code == (code ?? ''),
      orElse: () =>
          LanguageModel(code: '', codeTranslator: '', name: '', isuse: false),
    );
    return name.name!;
  }

  @override
  Widget build(BuildContext context) {
    return LayoutBuilder(
      builder: (context, constraints) {
        return SafeArea(
          child: Scaffold(
            appBar: AppBar(
              backgroundColor: global.theme.appBarColor,
              leading: BackButton(
                onPressed: () {
                  Navigator.pop(context);
                },
              ),
              title: Text(global.language('slip_design')),
            ),
            body: Container(
              child: (constraints.maxWidth > 700)
                  ? SplitView(
                      controller: splitViewController,
                      gripSize: 8,
                      gripColor: global.theme.appBarColor,
                      gripColorActive: Colors.blueAccent.shade700,
                      viewMode: SplitViewMode.Horizontal,
                      indicator: const SplitIndicator(
                        viewMode: SplitViewMode.Horizontal,
                      ),
                      activeIndicator: const SplitIndicator(
                        viewMode: SplitViewMode.Horizontal,
                        isActive: true,
                      ),
                      children: [previewWidget(), editWidget()],
                    )
                  : TabBarView(
                      physics: const NeverScrollableScrollPhysics(),
                      controller: tabController,
                      children: [editWidget(), previewWidget()],
                    ),
            ),
          ),
        );
      },
    );
  }

  Widget editWidget() {
    List<Widget> formWidget = [];

    formWidget.add(headerTab());
    formWidget.add(detailTab());
    formWidget.add(totalTab());
    formWidget.add(footerTab());
    List<Widget> tabx = [
      Tab(text: global.language("bill_header")),
      Tab(text: global.language("bill_detail")),
      Tab(text: global.language("bill_total")),
      Tab(text: global.language("bill_footer")),
    ];

    return Scaffold(
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: TabBar(controller: editTabController, tabs: tabx),
      ),
      body: RawKeyboardListener(
        focusNode: FocusNode(skipTraversal: true),
        onKey: (event) async {
          if (kIsWeb ||
              Platform.isWindows ||
              Platform.isLinux ||
              Platform.isMacOS) {
            if (event is RawKeyUpEvent) {
              if (event.logicalKey == LogicalKeyboardKey.f10) {}
            }
          }
        },
        child: TabBarView(controller: editTabController, children: formWidget),
      ),
    );
  }

  Widget footerTab() {
    List<Widget> formFooter = [];
    footerController = [];

    formFooter.add(
      Container(
        alignment: Alignment.centerLeft,
        child: ElevatedButton(
          onPressed: () {
            List<LanguageDataModel> names = [];
            for (var data in languageList) {
              names.add(LanguageDataModel(code: data.code!, name: ''));
            }
            footer.add(
              FormDesignFooterModel(description: names, print_qr_doc_no: false),
            );
            setState(() {});
          },
          child: Text(global.language("bill_add_column")),
        ),
      ),
    );
    formFooter.add(const SizedBox(height: 10));

    int idx = 0;
    List<int> selectFooterOptions = [];
    footerValueController = [];
    imageWeb = [];
    imageFile = [];
    for (int i = 0; i < footer.length; i++) {
      var main = footer[i];
      selectFooterOptions.add(main.footerType);
      footerValueController.add(TextEditingController());
      footerValueController[i].text = main.value;
      final detailWidthLength = footerValueController[i].text.length;
      footerValueController[i].selection = TextSelection(
        baseOffset: detailWidthLength,
        extentOffset: detailWidthLength,
      );
      List<Widget> formFooterDetail = [];
      imageWeb.add(Uint8List(0));
      imageFile.add(File(''));

      formFooter.add(
        Row(
          children: <Widget>[
            Expanded(
              flex: 1,
              child: RadioListTile<int>(
                value: 0,
                groupValue: selectFooterOptions[i],
                onChanged: (int? value) {
                  setState(() {
                    main.footerType = value!;
                  });
                },
                title: Text(global.language('bill_text')),
              ),
            ),
            Expanded(
              flex: 1,
              child: RadioListTile<int>(
                value: 1,
                groupValue: selectFooterOptions[i],
                onChanged: (int? value) {
                  setState(() {
                    main.footerType = value!;
                  });
                },
                title: Text(global.language('bill_image')),
              ),
            ),
            Expanded(
              flex: 1,
              child: RadioListTile<int>(
                value: 2,
                groupValue: selectFooterOptions[i],
                onChanged: (int? value) {
                  setState(() {
                    main.footerType = value!;
                  });
                },
                title: const Text('QR Code'),
              ),
            ),
            Expanded(
              flex: 1,
              child: RadioListTile<int>(
                value: 3,
                groupValue: selectFooterOptions[i],
                onChanged: (int? value) {
                  setState(() {
                    main.footerType = value!;
                  });
                },
                title: const Text('Barcode'),
              ),
            ),
            Expanded(
              flex: 2,
              child: IconButton(
                onPressed: () {
                  footer.removeAt(i);
                  setState(() {});
                },
                icon: const Icon(Icons.delete),
                color: Colors.red,
              ),
            ),
          ],
        ),
      );
      for (
        int languageIndex = 0;
        languageIndex < languageList.length;
        languageIndex++
      ) {
        LanguageDataModel nameObj = main.description.firstWhere(
          (element) => element.code == languageList[languageIndex].code,
          orElse: () => LanguageDataModel(code: '', name: ''),
        );
        if (nameObj.code == '') {
          main.description.add(
            LanguageDataModel(
              code: languageList[languageIndex].code!,
              name: '',
            ),
          );
        }
      }
      if (main.footerType == 0) {
        for (
          int languageIndex = 0;
          languageIndex < languageList.length;
          languageIndex++
        ) {
          LanguageDataModel nameObj = main.description.firstWhere(
            (element) => element.code == languageList[languageIndex].code,
            orElse: () => LanguageDataModel(code: '', name: ''),
          );
          if (nameObj.code != '') {
            footerController.add(TextEditingController());
            footerController[idx].text = nameObj.name;
            final textLength = footerController[idx].text.length;
            footerController[idx].selection = TextSelection(
              baseOffset: textLength,
              extentOffset: textLength,
            );

            formFooterDetail.add(
              Expanded(
                child: Padding(
                  padding: EdgeInsets.all(8.0),
                  child: TextFormField(
                    decoration: InputDecoration(
                      floatingLabelBehavior: FloatingLabelBehavior.always,
                      border: const OutlineInputBorder(),
                      labelText:
                          '${global.language("bill_design_detail")} ${nameObj.code}',
                    ),
                    onChanged: (value) {
                      setState(() {
                        nameObj.name = value;
                      });
                    },
                    controller: footerController[idx],
                  ),
                ),
              ),
            );
          }
          idx++;
        }
      }
      formFooter.add(Row(children: formFooterDetail));
      if (main.footerType == 1) {
        formFooter.add(
          Container(
            width: 300,
            padding: const EdgeInsets.only(
              left: 5,
              right: 5,
              bottom: 5,
              top: 5,
            ),
            decoration: BoxDecoration(
              border: Border.all(color: Colors.grey),
              borderRadius: const BorderRadius.all(Radius.circular(5.0)),
            ),
            child: Column(
              children: [
                Row(
                  children: [
                    Expanded(
                      child: IconButton(
                        focusNode: FocusNode(skipTraversal: true),
                        onPressed: () async {
                          imageWeb[i] = Uint8List(0);
                          imageFile[i] = File('');

                          setState(() {});
                        },
                        icon: const Icon(Icons.delete),
                      ),
                    ),
                    const SizedBox(width: 5),
                    Expanded(
                      child: IconButton(
                        focusNode: FocusNode(skipTraversal: true),
                        onPressed: () async {
                          final XFile? image = await imagePicker.pickImage(
                            source: ImageSource.gallery,
                            maxHeight: 480,
                            maxWidth: 640,
                          );
                          if (image != null) {
                            var f = await image.readAsBytes();
                            imageWeb[i] = f;
                            imageFile[i] = File(image.path);
                            setState(() {});
                          }
                        },
                        icon: const Icon(Icons.folder),
                      ),
                    ),
                    const SizedBox(width: 5),
                    Expanded(
                      child: IconButton(
                        focusNode: FocusNode(skipTraversal: true),
                        onPressed: () async {
                          final XFile? photo = await imagePicker.pickImage(
                            source: ImageSource.camera,
                            maxHeight: 480,
                            maxWidth: 640,
                            imageQuality: 60,
                          );
                          if (photo != null) {
                            var f = await photo.readAsBytes();
                            imageWeb[i] = f;
                            imageFile[i] = File(photo.path);
                            setState(() {});
                          }
                        },
                        icon: const Icon(Icons.camera_alt),
                      ),
                    ),
                  ],
                ),
                SizedBox(
                  width: 300,
                  height: 300,
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
                            imageWeb[i] = bytes;
                          });
                        },
                        onDropMultiple: (ev) async {},
                      ),
                      Center(
                        child: DecoratedBox(
                          decoration: BoxDecoration(
                            color: Colors.white,
                            border: Border.all(color: Colors.black),
                            borderRadius: BorderRadius.circular(5),
                            boxShadow: const [
                              BoxShadow(
                                offset: Offset(0, 4),
                                color: Colors.cyan, //edited
                                spreadRadius: 4,
                                blurRadius: 10, //edited
                              ),
                            ],
                            image: (imageWeb[i].isNotEmpty)
                                ? DecorationImage(
                                    image: MemoryImage(imageWeb[i]),
                                    fit: BoxFit.fill,
                                  )
                                : (main.imgUri != '')
                                ? DecorationImage(
                                    image: NetworkImage(main.imgUri),
                                    fit: BoxFit.fill,
                                  )
                                : const DecorationImage(
                                    image: AssetImage('assets/img/noimage.png'),
                                  ),
                          ),
                          child: const SizedBox(width: 500, height: 500),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        );
      } else if (main.footerType == 2) {
        formFooter.add(
          Padding(
            padding: EdgeInsets.all(8.0),
            child: TextFormField(
              decoration: InputDecoration(
                floatingLabelBehavior: FloatingLabelBehavior.always,
                border: OutlineInputBorder(),
                labelText: global.language('bill_text'),
              ),
              onChanged: (value) {
                main.value = value;
                setState(() {});
              },
              controller: footerValueController[i],
            ),
          ),
        );
        if (main.value.isNotEmpty) {
          formFooter.add(
            Center(
              child: QrImageView(
                data: main.value,
                version: QrVersions.auto,
                size: 300,
                gapless: false,
              ),
            ),
          );
        }
      } else if (main.footerType == 3) {
        formFooter.add(
          Padding(
            padding: EdgeInsets.all(8.0),
            child: TextFormField(
              decoration: InputDecoration(
                floatingLabelBehavior: FloatingLabelBehavior.always,
                border: OutlineInputBorder(),
                labelText: global.language('bill_text'),
              ),
              onChanged: (value) {
                main.value = value;
                setState(() {});
              },
              controller: footerValueController[i],
            ),
          ),
        );
        if (main.value.isNotEmpty) {
          formFooter.add(
            Center(
              child: BarcodeWidget(
                width: 200,
                height: 100,
                barcode: Barcode.code128(),
                data: main.value,
                errorBuilder: (context, error) => Center(child: Text(error)),
              ),
            ),
          );
        }
      }
    }

    formFooter.add(const SizedBox(height: 20));

    return Container(
      margin: const EdgeInsets.all(5),
      padding: const EdgeInsets.all(10),
      height: 100,
      width: double.infinity,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        boxShadow: [
          BoxShadow(
            color: Colors.grey.withOpacity(0.5),
            spreadRadius: 5,
            blurRadius: 7,
            offset: const Offset(0, 3),
          ),
        ],
      ),
      child: SingleChildScrollView(child: Column(children: formFooter)),
    );
  }

  Widget totalTab() {
    List<Widget> totalForm = [];
    totalForm.add(
      Padding(
        padding: const EdgeInsets.only(top: 10, bottom: 10),
        child: Row(
          children: [
            SizedBox(
              child: Text(
                global.language('item_details'),
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
              ),
            ),
            Container(
              margin: EdgeInsets.only(left: 10),
              child: ElevatedButton(
                onPressed: () {
                  List<LanguageDataModel> names = [];
                  for (var data in languageList) {
                    names.add(LanguageDataModel(code: data.code!, name: ''));
                  }
                  detailTotalColumn.add([
                    FormDesignColumnModel(
                      command: "column",
                      header_names: names,
                      width: 1,
                    ),
                  ]);
                  setState(() {});
                },
                child: Text(global.language("bill_add_column")),
              ),
            ),
          ],
        ),
      ),
    );

    totalForm.add( SizedBox(height: 10));
    totalForm.add(
      SizedBox(
        width: double.infinity,
        child: Padding(
          padding: EdgeInsets.only(left: 8, right: 8),
          child: Text(
            "${global.language("variables")} : &item_qty&=${global.language("quantity")} , &item_name&=${global.language("product_name")} , &item_unit_name&=${global.language("unit_name")} , &item_price_and_symbol&=${global.language("price")} , &item_discount&=${global.language("discount")} , &item_total_amount&=${global.language("total_amount")}",
            maxLines: 3,
            overflow: TextOverflow.ellipsis,
            style: const TextStyle(fontSize: 15),
            textAlign: TextAlign.left,
          ),
        ),
      ),
    );
    totalForm.add(const SizedBox(height: 10));
    List<Widget> columnDetail = [];

    for (int z = 0; z < detailTotalColumn.length; z++) {
      var main = detailTotalColumn[z];
      List<Widget> rowDetails = [];
      rowDetails.add(
        Padding(
          padding: const EdgeInsets.only(top: 10, bottom: 10),
          child: Row(
            children: [
              SizedBox(
                child: Text(
                  "Column ${z + 1}",
                  style: const TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                  ),
                ),
              ),
              Container(
                margin: EdgeInsets.only(left: 10),
                child: ElevatedButton(
                  onPressed: () {
                    List<LanguageDataModel> names = [];
                    for (var data in languageList) {
                      names.add(LanguageDataModel(code: data.code!, name: ''));
                    }
                    main.add(
                      FormDesignColumnModel(
                        command: "row",
                        header_names: names,
                        width: 1,
                      ),
                    );
                    setState(() {});
                  },
                  child: Text(global.language("bill_add_row")),
                ),
              ),
              Container(
                margin: EdgeInsets.only(left: 10),
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(backgroundColor: Colors.red),
                  onPressed: () {
                    detailTotalColumn.removeAt(z);
                    setState(() {});
                  },
                  child: Text(global.language("bill_delete_column")),
                ),
              ),
            ],
          ),
        ),
      );
      List<TextEditingController> detailTotalCommand = [];
      List<TextEditingController> detailTotalWidth = [];
      List<TextEditingController> detailTotalSize = [];
      List<global.PrintColumnAlign> selectTotalOptions = [];

      for (int i = 0; i < main.length; i++) {
        rowDetails.add(
          Row(
            children: [
              Padding(
                padding: const EdgeInsets.symmetric(horizontal: 8.0),
                child: Text(
                  'Row ${i + 1}',
                  style: TextStyle(color: Colors.grey.shade600),
                ),
              ),
              Expanded(
                child: Padding(
                  padding: const EdgeInsets.only(right: 8),
                  child: Divider(color: Colors.grey.shade600),
                ),
              ),
              IconButton(
                onPressed: () {
                  main.removeAt(i);
                  setState(() {});
                },
                icon: const Icon(Icons.delete),
                color: Colors.red,
              ),
            ],
          ),
        );
        rowDetails.add(const SizedBox(height: 10));

        detailTotalCommand.add(TextEditingController());
        detailTotalWidth.add(TextEditingController());
        detailTotalSize.add(TextEditingController());
        detailTotalCommand[i].text = main[i].command;
        final detailtextLength = detailTotalCommand[i].text.length;
        detailTotalCommand[i].selection = TextSelection(
          baseOffset: detailtextLength,
          extentOffset: detailtextLength,
        );
        detailTotalWidth[i].text = main[i].width.toString();
        final detailWidthLength = detailTotalWidth[i].text.length;
        detailTotalWidth[i].selection = TextSelection(
          baseOffset: detailWidthLength,
          extentOffset: detailWidthLength,
        );
        detailTotalSize[i].text = main[i].font_size.toString();
        final detailSizeLength = detailTotalSize[i].text.length;
        detailTotalSize[i].selection = TextSelection(
          baseOffset: detailSizeLength,
          extentOffset: detailSizeLength,
        );
        selectTotalOptions.add(global.PrintColumnAlign.left);
        selectTotalOptions[i] = main[i].text_align;
        rowDetails.add(
          Row(
            children: [
              Expanded(
                flex: 1,
                child: Padding(
                  padding: const EdgeInsets.all(8.0),
                  child: TextFormField(
                    decoration: InputDecoration(
                      floatingLabelBehavior: FloatingLabelBehavior.always,
                      border: const OutlineInputBorder(),
                      labelText: global.language('bill_command'),
                    ),
                    onChanged: (value) {
                      setState(() {
                        main[i].command = value;
                      });
                    },
                    controller: detailTotalCommand[i],
                  ),
                ),
              ),
            ],
          ),
        );

        rowDetails.add(
          Row(
            children: [
              Expanded(
                flex: 1,
                child: Padding(
                  padding: const EdgeInsets.all(8.0),
                  child: TextFormField(
                    decoration: InputDecoration(
                      floatingLabelBehavior: FloatingLabelBehavior.always,
                      border: const OutlineInputBorder(),
                      labelText: global.language('bill_width'),
                    ),
                    onChanged: (value) {
                      double widthVal = 0;
                      if (value.isNotEmpty) {
                        widthVal = double.parse(value);
                      }

                      setState(() {
                        main[i].width = widthVal;
                      });
                    },
                    keyboardType: TextInputType.number,
                    inputFormatters: [
                      FilteringTextInputFormatter.allow(RegExp(r'[0-9]')),
                    ],
                    controller: detailTotalWidth[i],
                  ),
                ),
              ),
              Expanded(
                flex: 1,
                child: Padding(
                  padding: const EdgeInsets.all(8.0),
                  child: TextFormField(
                    decoration: InputDecoration(
                      floatingLabelBehavior: FloatingLabelBehavior.always,
                      border: const OutlineInputBorder(),
                      labelText: global.language('bill_size'),
                    ),
                    onChanged: (value) {
                      double widthVal = 0;
                      if (value.isNotEmpty) {
                        widthVal = double.parse(value);
                      }

                      setState(() {
                        main[i].font_size = widthVal;
                      });
                    },
                    keyboardType: TextInputType.number,
                    inputFormatters: [
                      FilteringTextInputFormatter.allow(RegExp(r'[0-9]')),
                    ],
                    controller: detailTotalSize[i],
                  ),
                ),
              ),
              Expanded(
                flex: 1,
                child: RadioListTile(
                  title: Text(global.language('align_left')),
                  value: global.PrintColumnAlign.left,
                  groupValue: selectTotalOptions[i],
                  onChanged: (value) {
                    setState(() {
                      main[i].text_align = global.PrintColumnAlign.left;
                    });
                  },
                ),
              ),
              Expanded(
                flex: 1,
                child: RadioListTile(
                  title: Text(global.language('align_center')),
                  value: global.PrintColumnAlign.center,
                  groupValue: selectTotalOptions[i],
                  onChanged: (value) {
                    setState(() {
                      main[i].text_align = global.PrintColumnAlign.center;
                    });
                  },
                ),
              ),
              Expanded(
                flex: 1,
                child: RadioListTile(
                  title: Text(global.language('align_right')),
                  value: global.PrintColumnAlign.right,
                  groupValue: selectTotalOptions[i],
                  onChanged: (value) {
                    setState(() {
                      main[i].text_align = global.PrintColumnAlign.right;
                    });
                  },
                ),
              ),
            ],
          ),
        );

        rowDetails.add(
          Row(
            children: [
              Expanded(
                flex: 1,
                child: Padding(
                  padding: const EdgeInsets.all(8.0),
                  child: Row(
                    children: [
                      Checkbox(
                        value: main[i].font_weight_bold,
                        onChanged: (value) {
                          if (main[i].font_weight_bold == true) {
                            main[i].font_weight_bold = false;
                          } else {
                            main[i].font_weight_bold = true;
                          }
                          setState(() {});
                        },
                      ),
                      Text(global.language('text_bold')),
                      const SizedBox(width: 30),
                      Checkbox(
                        value: main[i].font_style_italic,
                        onChanged: (value) {
                          if (main[i].font_style_italic == true) {
                            main[i].font_style_italic = false;
                          } else {
                            main[i].font_style_italic = true;
                          }
                          setState(() {});
                        },
                      ),
                      Text(global.language('text_italic')),
                      const SizedBox(width: 30),
                      Checkbox(
                        value: main[i].decoration_underline,
                        onChanged: (value) {
                          if (main[i].decoration_underline == true) {
                            main[i].decoration_underline = false;
                          } else {
                            main[i].decoration_underline = true;
                          }
                          setState(() {});
                        },
                      ),
                      Text(global.language('text_underline')),
                    ],
                  ),
                ),
              ),
            ],
          ),
        );
      }

      columnDetail.addAll(rowDetails);
    }
    totalForm.addAll(columnDetail);
    return Container(
      margin: const EdgeInsets.all(5),
      padding: const EdgeInsets.all(10),
      height: 100,
      width: double.infinity,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        boxShadow: [
          BoxShadow(
            color: Colors.grey.withOpacity(0.5),
            spreadRadius: 5,
            blurRadius: 7,
            offset: const Offset(0, 3),
          ),
        ],
      ),
      child: SingleChildScrollView(child: Column(children: totalForm)),
    );
  }

  Widget detailTab() {
    List<Widget> formDetail = [];
    List<Widget> formExtra = [];

    List<TextEditingController> detailCommand = [];
    List<TextEditingController> detailWidth = [];
    List<TextEditingController> detailSize = [];
    List<global.PrintColumnAlign> selectOptions = [];

    List<TextEditingController> detailExtraCommand = [];
    List<TextEditingController> detailExtraWidth = [];
    List<TextEditingController> detailExtraSize = [];
    List<global.PrintColumnAlign> detailExtraselectOptions = [];
    formDetail.add(
      Padding(
        padding: const EdgeInsets.only(top: 10, bottom: 10),
        child: Row(
          children: [
            SizedBox(
              child: Text(
                global.language('item_details'),
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
              ),
            ),
            Container(
              margin: EdgeInsets.only(left: 10),
              child: ElevatedButton(
                onPressed: () {
                  List<LanguageDataModel> names = [];
                  for (var data in languageList) {
                    names.add(LanguageDataModel(code: data.code!, name: ''));
                  }
                  detailColumn.add(
                    FormDesignColumnModel(
                      command: "",
                      header_names: names,
                      width: 1,
                    ),
                  );
                  setState(() {});
                },
                child: Text(global.language("bill_add_detail")),
              ),
            ),
          ],
        ),
      ),
    );

    formDetail.add( SizedBox(height: 10));
    formDetail.add(
      SizedBox(
        width: double.infinity,
        child: Padding(
          padding: EdgeInsets.only(left: 8, right: 8),
          child: Text(
            "${global.language("variables")} : &item_qty&=${global.language("quantity")} , &item_name&=${global.language("product_name")} , &item_unit_name&=${global.language("unit_name")} , &item_price_and_symbol&=${global.language("price")} , &item_discount&=${global.language("discount")} , &item_total_amount&=${global.language("total_amount")}",
            maxLines: 3,
            overflow: TextOverflow.ellipsis,
            style: const TextStyle(fontSize: 15),
            textAlign: TextAlign.left,
          ),
        ),
      ),
    );
    formDetail.add(const SizedBox(height: 10));

    for (int i = 0; i < detailColumn.length; i++) {
      formDetail.add(
        Row(
          children: [
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 8.0),
              child: Text(
                'Row ${i + 1}',
                style: TextStyle(color: Colors.grey.shade600),
              ),
            ),
            Expanded(
              child: Padding(
                padding: const EdgeInsets.only(right: 8),
                child: Divider(color: Colors.grey.shade600),
              ),
            ),
            IconButton(
              onPressed: () {
                detailColumn.removeAt(i);
                setState(() {});
              },
              icon: const Icon(Icons.delete),
              color: Colors.red,
            ),
          ],
        ),
      );
      formDetail.add(const SizedBox(height: 10));

      detailCommand.add(TextEditingController());
      detailWidth.add(TextEditingController());
      detailSize.add(TextEditingController());
      detailCommand[i].text = detailColumn[i].command;
      final detailtextLength = detailCommand[i].text.length;
      detailCommand[i].selection = TextSelection(
        baseOffset: detailtextLength,
        extentOffset: detailtextLength,
      );
      detailWidth[i].text = detailColumn[i].width.toString();
      final detailWidthLength = detailWidth[i].text.length;
      detailWidth[i].selection = TextSelection(
        baseOffset: detailWidthLength,
        extentOffset: detailWidthLength,
      );
      detailSize[i].text = detailColumn[i].font_size.toString();
      final detailSizeLength = detailSize[i].text.length;
      detailSize[i].selection = TextSelection(
        baseOffset: detailSizeLength,
        extentOffset: detailSizeLength,
      );
      selectOptions.add(global.PrintColumnAlign.left);
      selectOptions[i] = detailColumn[i].text_align;
      formDetail.add(
        Row(
          children: [
            Expanded(
              flex: 1,
              child: Padding(
                padding: const EdgeInsets.all(8.0),
                child: TextFormField(
                  decoration: InputDecoration(
                    floatingLabelBehavior: FloatingLabelBehavior.always,
                    border: const OutlineInputBorder(),
                    labelText: global.language('bill_command'),
                  ),
                  onChanged: (value) {
                    setState(() {
                      detailColumn[i].command = value;
                    });
                  },
                  controller: detailCommand[i],
                ),
              ),
            ),
          ],
        ),
      );

      formDetail.add(
        Row(
          children: [
            Expanded(
              flex: 1,
              child: Padding(
                padding: const EdgeInsets.all(8.0),
                child: TextFormField(
                  decoration: InputDecoration(
                    floatingLabelBehavior: FloatingLabelBehavior.always,
                    border: const OutlineInputBorder(),
                    labelText: global.language('bill_width'),
                  ),
                  onChanged: (value) {
                    double widthVal = 0;
                    if (value.isNotEmpty) {
                      widthVal = double.parse(value);
                    }

                    setState(() {
                      detailColumn[i].width = widthVal;
                    });
                  },
                  keyboardType: TextInputType.number,
                  inputFormatters: [
                    FilteringTextInputFormatter.allow(RegExp(r'[0-9]')),
                  ],
                  controller: detailWidth[i],
                ),
              ),
            ),
            Expanded(
              flex: 1,
              child: Padding(
                padding: const EdgeInsets.all(8.0),
                child: TextFormField(
                  decoration: InputDecoration(
                    floatingLabelBehavior: FloatingLabelBehavior.always,
                    border: const OutlineInputBorder(),
                    labelText: global.language('bill_size'),
                  ),
                  onChanged: (value) {
                    double widthVal = 0;
                    if (value.isNotEmpty) {
                      widthVal = double.parse(value);
                    }

                    setState(() {
                      detailColumn[i].font_size = widthVal;
                    });
                  },
                  keyboardType: TextInputType.number,
                  inputFormatters: [
                    FilteringTextInputFormatter.allow(RegExp(r'[0-9]')),
                  ],
                  controller: detailSize[i],
                ),
              ),
            ),
            Expanded(
              flex: 1,
              child: RadioListTile(
                title: Text(global.language('align_left')),
                value: global.PrintColumnAlign.left,
                groupValue: selectOptions[i],
                onChanged: (value) {
                  setState(() {
                    detailColumn[i].text_align = global.PrintColumnAlign.left;
                  });
                },
              ),
            ),
            Expanded(
              flex: 1,
              child: RadioListTile(
                title: Text(global.language('align_center')),
                value: global.PrintColumnAlign.center,
                groupValue: selectOptions[i],
                onChanged: (value) {
                  setState(() {
                    detailColumn[i].text_align = global.PrintColumnAlign.center;
                  });
                },
              ),
            ),
            Expanded(
              flex: 1,
              child: RadioListTile(
                title: Text(global.language('align_right')),
                value: global.PrintColumnAlign.right,
                groupValue: selectOptions[i],
                onChanged: (value) {
                  setState(() {
                    detailColumn[i].text_align = global.PrintColumnAlign.right;
                  });
                },
              ),
            ),
          ],
        ),
      );

      formDetail.add(
        Row(
          children: [
            Expanded(
              flex: 1,
              child: Padding(
                padding: const EdgeInsets.all(8.0),
                child: Row(
                  children: [
                    Checkbox(
                      value: detailColumn[i].font_weight_bold,
                      onChanged: (value) {
                        if (detailColumn[i].font_weight_bold == true) {
                          detailColumn[i].font_weight_bold = false;
                        } else {
                          detailColumn[i].font_weight_bold = true;
                        }
                        setState(() {});
                      },
                    ),
                    Text(global.language('text_bold')),
                    const SizedBox(width: 30),
                    Checkbox(
                      value: detailColumn[i].font_style_italic,
                      onChanged: (value) {
                        if (detailColumn[i].font_style_italic == true) {
                          detailColumn[i].font_style_italic = false;
                        } else {
                          detailColumn[i].font_style_italic = true;
                        }
                        setState(() {});
                      },
                    ),
                    Text(global.language('text_italic')),
                    const SizedBox(width: 30),
                    Checkbox(
                      value: detailColumn[i].decoration_underline,
                      onChanged: (value) {
                        if (detailColumn[i].decoration_underline == true) {
                          detailColumn[i].decoration_underline = false;
                        } else {
                          detailColumn[i].decoration_underline = true;
                        }
                        setState(() {});
                      },
                    ),
                    Text(global.language('text_underline')),
                  ],
                ),
              ),
            ),
          ],
        ),
      );
    }
    formExtra.add(
      const Padding(
        padding: EdgeInsets.only(top: 10),
        child: Divider(height: 3),
      ),
    );
    formExtra.add(
      Padding(
        padding: const EdgeInsets.only(top: 10, bottom: 10),
        child: Row(
          children: [
            SizedBox(
              child: Text(
                global.language('extra_details'),
                style: TextStyle(fontSize: 16, fontWeight: FontWeight.bold),
              ),
            ),
            Container(
              margin: EdgeInsets.only(left: 10),
              child: ElevatedButton(
                onPressed: () {
                  List<LanguageDataModel> names = [];
                  for (var data in languageList) {
                    names.add(LanguageDataModel(code: data.code!, name: ''));
                  }
                  detailExtraColumn.add(
                    FormDesignColumnModel(
                      command: "",
                      header_names: names,
                      width: 1,
                    ),
                  );
                  setState(() {});
                },
                child: Text(global.language("bill_add_detail")),
              ),
            ),
          ],
        ),
      ),
    );

    formExtra.add( SizedBox(height: 10));
    formExtra.add(
      SizedBox(
        width: double.infinity,
        child: Padding(
          padding: EdgeInsets.only(left: 8, right: 8),
          child: Text(
            "${global.language("variables")} : &item_qty&=${global.language("quantity")} , &item_name&=${global.language("product_name")} , &item_unit_name&=${global.language("unit_name")} , &item_price_and_symbol&=${global.language("price")} , &item_discount&=${global.language("discount")} , &item_total_amount&=${global.language("total_amount")}",
            maxLines: 3,
            overflow: TextOverflow.ellipsis,
            style: const TextStyle(fontSize: 15),
            textAlign: TextAlign.left,
          ),
        ),
      ),
    );

    formExtra.add(const SizedBox(height: 20));
    for (int i = 0; i < detailExtraColumn.length; i++) {
      formExtra.add(
        Row(
          children: [
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 8.0),
              child: Text(
                'Row ${i + 1}',
                style: TextStyle(color: Colors.grey.shade600),
              ),
            ),
            Expanded(
              child: Padding(
                padding: const EdgeInsets.only(right: 8),
                child: Divider(color: Colors.grey.shade600),
              ),
            ),
            IconButton(
              onPressed: () {
                detailExtraColumn.removeAt(i);
                setState(() {});
              },
              icon: const Icon(Icons.delete),
              color: Colors.red,
            ),
          ],
        ),
      );
      formExtra.add(const SizedBox(height: 10));

      detailExtraCommand.add(TextEditingController());
      detailExtraWidth.add(TextEditingController());
      detailExtraSize.add(TextEditingController());
      detailExtraCommand[i].text = detailExtraColumn[i].command;
      final detailExtratextLength = detailExtraCommand[i].text.length;
      detailExtraCommand[i].selection = TextSelection(
        baseOffset: detailExtratextLength,
        extentOffset: detailExtratextLength,
      );
      detailExtraWidth[i].text = detailExtraColumn[i].width.toString();
      final detailExtraWidthLength = detailExtraWidth[i].text.length;
      detailExtraWidth[i].selection = TextSelection(
        baseOffset: detailExtraWidthLength,
        extentOffset: detailExtraWidthLength,
      );
      detailExtraSize[i].text = detailExtraColumn[i].font_size.toString();
      final detailExtraSizeLength = detailExtraSize[i].text.length;
      detailExtraSize[i].selection = TextSelection(
        baseOffset: detailExtraSizeLength,
        extentOffset: detailExtraSizeLength,
      );
      detailExtraselectOptions.add(global.PrintColumnAlign.left);
      detailExtraselectOptions[i] = detailExtraColumn[i].text_align;

      formExtra.add(
        Row(
          children: [
            Expanded(
              flex: 1,
              child: Padding(
                padding: const EdgeInsets.all(8.0),
                child: TextFormField(
                  decoration: InputDecoration(
                    floatingLabelBehavior: FloatingLabelBehavior.always,
                    border: const OutlineInputBorder(),
                    labelText: global.language('bill_command'),
                  ),
                  onChanged: (value) {
                    setState(() {
                      detailExtraColumn[i].command = value;
                    });
                  },
                  controller: detailExtraCommand[i],
                ),
              ),
            ),
          ],
        ),
      );

      formExtra.add(
        Row(
          children: [
            Expanded(
              flex: 1,
              child: Padding(
                padding: const EdgeInsets.all(8.0),
                child: TextFormField(
                  decoration: InputDecoration(
                    floatingLabelBehavior: FloatingLabelBehavior.always,
                    border: const OutlineInputBorder(),
                    labelText: global.language('bill_width'),
                  ),
                  onChanged: (value) {
                    double widthVal = 0;
                    if (value.isNotEmpty) {
                      widthVal = double.parse(value);
                    }

                    setState(() {
                      detailExtraColumn[i].width = widthVal;
                    });
                  },
                  keyboardType: TextInputType.number,
                  inputFormatters: [
                    FilteringTextInputFormatter.allow(RegExp(r'[0-9]')),
                  ],
                  controller: detailExtraWidth[i],
                ),
              ),
            ),
            Expanded(
              flex: 1,
              child: Padding(
                padding: const EdgeInsets.all(8.0),
                child: TextFormField(
                  decoration: InputDecoration(
                    floatingLabelBehavior: FloatingLabelBehavior.always,
                    border: const OutlineInputBorder(),
                    labelText: global.language('bill_size'),
                  ),
                  onChanged: (value) {
                    double widthVal = 0;
                    if (value.isNotEmpty) {
                      widthVal = double.parse(value);
                    }

                    setState(() {
                      detailExtraColumn[i].font_size = widthVal;
                    });
                  },
                  keyboardType: TextInputType.number,
                  inputFormatters: [
                    FilteringTextInputFormatter.allow(RegExp(r'[0-9]')),
                  ],
                  controller: detailExtraSize[i],
                ),
              ),
            ),
            Expanded(
              flex: 1,
              child: RadioListTile(
                title: Text(global.language('align_left')),
                value: global.PrintColumnAlign.left,
                groupValue: detailExtraselectOptions[i],
                onChanged: (value) {
                  setState(() {
                    detailExtraColumn[i].text_align =
                        global.PrintColumnAlign.left;
                  });
                },
              ),
            ),
            Expanded(
              flex: 1,
              child: RadioListTile(
                title: Text(global.language('align_center')),
                value: global.PrintColumnAlign.center,
                groupValue: detailExtraselectOptions[i],
                onChanged: (value) {
                  setState(() {
                    detailExtraColumn[i].text_align =
                        global.PrintColumnAlign.center;
                  });
                },
              ),
            ),
            Expanded(
              flex: 1,
              child: RadioListTile(
                title: Text(global.language('align_right')),
                value: global.PrintColumnAlign.right,
                groupValue: detailExtraselectOptions[i],
                onChanged: (value) {
                  setState(() {
                    detailExtraColumn[i].text_align =
                        global.PrintColumnAlign.right;
                  });
                },
              ),
            ),
          ],
        ),
      );

      formExtra.add(
        Row(
          children: [
            Expanded(
              flex: 1,
              child: Padding(
                padding: const EdgeInsets.all(8.0),
                child: Row(
                  children: [
                    Checkbox(
                      value: detailExtraColumn[i].font_weight_bold,
                      onChanged: (value) {
                        if (detailExtraColumn[i].font_weight_bold == true) {
                          detailExtraColumn[i].font_weight_bold = false;
                        } else {
                          detailExtraColumn[i].font_weight_bold = true;
                        }

                        setState(() {});
                      },
                    ),
                    Text(global.language('text_bold')),
                    const SizedBox(width: 30),
                    Checkbox(
                      value: detailExtraColumn[i].font_style_italic,
                      onChanged: (value) {
                        if (detailExtraColumn[i].font_style_italic == true) {
                          detailExtraColumn[i].font_style_italic = false;
                        } else {
                          detailExtraColumn[i].font_style_italic = true;
                        }
                        setState(() {});
                      },
                    ),
                    Text(global.language('text_italic')),
                    const SizedBox(width: 30),
                    Checkbox(
                      value: detailExtraColumn[i].decoration_underline,
                      onChanged: (value) {
                        if (detailExtraColumn[i].decoration_underline == true) {
                          detailExtraColumn[i].decoration_underline = false;
                        } else {
                          detailExtraColumn[i].decoration_underline = true;
                        }
                        setState(() {});
                      },
                    ),
                    Text(global.language('text_underline')),
                  ],
                ),
              ),
            ),
          ],
        ),
      );
    }
    List<Widget> detailContainer = [];
    detailContainer.addAll(formDetail);
    detailContainer.addAll(formExtra);
    return Container(
      margin: const EdgeInsets.all(5),
      padding: const EdgeInsets.all(10),
      height: 100,
      width: double.infinity,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        boxShadow: [
          BoxShadow(
            color: Colors.grey.withOpacity(0.5),
            spreadRadius: 5,
            blurRadius: 7,
            offset: const Offset(0, 3),
          ),
        ],
      ),
      child: SingleChildScrollView(child: Column(children: detailContainer)),
    );
  }

  Widget headerTab() {
    List<Widget> formHeader = [];
    headerController = [];
    int idx = 0;
    List<int> selectHeaderOptions = [];
    headerValueController = [];
    imageHeaderWeb = [];
    imageHeaderFile = [];
    formHeader.add(
      SizedBox(
        width: double.infinity,
        child: Row(
          children: [
            Expanded(
              child: Padding(
                padding: EdgeInsets.all(8.0),
                child: TextFormField(
                  decoration: InputDecoration(
                    floatingLabelBehavior: FloatingLabelBehavior.always,
                    border: OutlineInputBorder(),
                    labelText: global.language('bill_design_code'),
                  ),
                  onChanged: (value) {
                    setState(() {});
                  },
                ),
              ),
            ),
            Expanded(
              child: Padding(
                padding: EdgeInsets.all(8.0),
                child: TextFormField(
                  decoration: InputDecoration(
                    floatingLabelBehavior: FloatingLabelBehavior.always,
                    border: OutlineInputBorder(),
                    labelText: global.language('bill_design_name'),
                  ),
                  onChanged: (value) {
                    setState(() {});
                  },
                ),
              ),
            ),
          ],
        ),
      ),
    );
    formHeader.add(
      Container(
        alignment: Alignment.centerLeft,
        child: ElevatedButton(
          onPressed: () {
            List<LanguageDataModel> names = [];
            for (var data in languageList) {
              names.add(LanguageDataModel(code: data.code!, name: ''));
            }
            header.add(FormDesignHeaderModel(description: names));
            setState(() {});
          },
          child: Text(global.language("bill_add_column")),
        ),
      ),
    );
    formHeader.add(const SizedBox(height: 10));

    for (int i = 0; i < header.length; i++) {
      var main = header[i];
      selectHeaderOptions.add(main.headerType);
      headerValueController.add(TextEditingController());
      headerValueController[i].text = main.value;
      final detailWidthLength = headerValueController[i].text.length;
      headerValueController[i].selection = TextSelection(
        baseOffset: detailWidthLength,
        extentOffset: detailWidthLength,
      );
      List<Widget> formHeaderDetail = [];
      imageHeaderWeb.add(Uint8List(0));
      imageHeaderFile.add(File(''));

      formHeader.add(
        Row(
          children: <Widget>[
            Expanded(
              flex: 1,
              child: RadioListTile<int>(
                value: 0,
                groupValue: selectHeaderOptions[i],
                onChanged: (int? value) {
                  setState(() {
                    main.headerType = value!;
                  });
                },
                title: Text(global.language('bill_text')),
              ),
            ),
            Expanded(
              flex: 1,
              child: RadioListTile<int>(
                value: 1,
                groupValue: selectHeaderOptions[i],
                onChanged: (int? value) {
                  setState(() {
                    main.headerType = value!;
                  });
                },
                title: Text(global.language('bill_image')),
              ),
            ),
            Expanded(
              flex: 1,
              child: RadioListTile<int>(
                value: 2,
                groupValue: selectHeaderOptions[i],
                onChanged: (int? value) {
                  setState(() {
                    main.headerType = value!;
                  });
                },
                title: const Text('QR Code'),
              ),
            ),
            Expanded(
              flex: 1,
              child: RadioListTile<int>(
                value: 3,
                groupValue: selectHeaderOptions[i],
                onChanged: (int? value) {
                  setState(() {
                    main.headerType = value!;
                  });
                },
                title: const Text('Barcode'),
              ),
            ),
            Expanded(
              flex: 2,
              child: IconButton(
                onPressed: () {
                  header.removeAt(i);
                  setState(() {});
                },
                icon: const Icon(Icons.delete),
                color: Colors.red,
              ),
            ),
          ],
        ),
      );
      for (
        int languageIndex = 0;
        languageIndex < languageList.length;
        languageIndex++
      ) {
        LanguageDataModel nameObj = main.description.firstWhere(
          (element) => element.code == languageList[languageIndex].code,
          orElse: () => LanguageDataModel(code: '', name: ''),
        );
        if (nameObj.code == '') {
          main.description.add(
            LanguageDataModel(
              code: languageList[languageIndex].code!,
              name: '',
            ),
          );
        }
      }
      if (main.headerType == 0) {
        for (
          int languageIndex = 0;
          languageIndex < languageList.length;
          languageIndex++
        ) {
          LanguageDataModel nameObj = main.description.firstWhere(
            (element) => element.code == languageList[languageIndex].code,
            orElse: () => LanguageDataModel(code: '', name: ''),
          );
          if (nameObj.code != '') {
            headerController.add(TextEditingController());
            headerController[idx].text = nameObj.name;
            final textLength = headerController[idx].text.length;
            headerController[idx].selection = TextSelection(
              baseOffset: textLength,
              extentOffset: textLength,
            );

            formHeaderDetail.add(
              Expanded(
                child: Padding(
                  padding: EdgeInsets.all(8.0),
                  child: TextFormField(
                    decoration: InputDecoration(
                      floatingLabelBehavior: FloatingLabelBehavior.always,
                      border: const OutlineInputBorder(),
                      labelText:
                          '${global.language("bill_design_detail")} ${nameObj.code}',
                    ),
                    onChanged: (value) {
                      setState(() {
                        nameObj.name = value;
                      });
                    },
                    controller: headerController[idx],
                  ),
                ),
              ),
            );
          }
          idx++;
        }
      }
      formHeader.add(Row(children: formHeaderDetail));
      if (main.headerType == 1) {
        formHeader.add(
          Container(
            width: 300,
            padding: const EdgeInsets.only(
              left: 5,
              right: 5,
              bottom: 5,
              top: 5,
            ),
            decoration: BoxDecoration(
              border: Border.all(color: Colors.grey),
              borderRadius: const BorderRadius.all(Radius.circular(5.0)),
            ),
            child: Column(
              children: [
                Row(
                  children: [
                    Expanded(
                      child: IconButton(
                        focusNode: FocusNode(skipTraversal: true),
                        onPressed: () async {
                          imageHeaderWeb[i] = Uint8List(0);
                          imageHeaderFile[i] = File('');

                          setState(() {});
                        },
                        icon: const Icon(Icons.delete),
                      ),
                    ),
                    const SizedBox(width: 5),
                    Expanded(
                      child: IconButton(
                        focusNode: FocusNode(skipTraversal: true),
                        onPressed: () async {
                          final XFile? image = await imagePicker.pickImage(
                            source: ImageSource.gallery,
                            maxHeight: 480,
                            maxWidth: 640,
                          );
                          if (image != null) {
                            var f = await image.readAsBytes();
                            imageHeaderWeb[i] = f;
                            imageHeaderFile[i] = File(image.path);
                            setState(() {});
                          }
                        },
                        icon: const Icon(Icons.folder),
                      ),
                    ),
                    const SizedBox(width: 5),
                    Expanded(
                      child: IconButton(
                        focusNode: FocusNode(skipTraversal: true),
                        onPressed: () async {
                          final XFile? photo = await imagePicker.pickImage(
                            source: ImageSource.camera,
                            maxHeight: 480,
                            maxWidth: 640,
                            imageQuality: 60,
                          );
                          if (photo != null) {
                            var f = await photo.readAsBytes();
                            imageHeaderWeb[i] = f;
                            imageHeaderFile[i] = File(photo.path);
                            setState(() {});
                          }
                        },
                        icon: const Icon(Icons.camera_alt),
                      ),
                    ),
                  ],
                ),
                SizedBox(
                  width: 300,
                  height: 300,
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
                            imageHeaderWeb[i] = bytes;
                          });
                        },
                        onDropMultiple: (ev) async {},
                      ),
                      Center(
                        child: DecoratedBox(
                          decoration: BoxDecoration(
                            color: Colors.white,
                            border: Border.all(color: Colors.black),
                            borderRadius: BorderRadius.circular(5),
                            boxShadow: const [
                              BoxShadow(
                                offset: Offset(0, 4),
                                color: Colors.cyan, //edited
                                spreadRadius: 4,
                                blurRadius: 10, //edited
                              ),
                            ],
                            image: (imageHeaderWeb[i].isNotEmpty)
                                ? DecorationImage(
                                    image: MemoryImage(imageHeaderWeb[i]),
                                    fit: BoxFit.fill,
                                  )
                                : (main.imgUri != '')
                                ? DecorationImage(
                                    image: NetworkImage(main.imgUri),
                                    fit: BoxFit.fill,
                                  )
                                : const DecorationImage(
                                    image: AssetImage('assets/img/noimage.png'),
                                  ),
                          ),
                          child: const SizedBox(width: 500, height: 500),
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        );
      } else if (main.headerType == 2) {
        formHeader.add(
          Padding(
            padding: EdgeInsets.all(8.0),
            child: TextFormField(
              decoration: InputDecoration(
                floatingLabelBehavior: FloatingLabelBehavior.always,
                border: OutlineInputBorder(),
                labelText: global.language('bill_text'),
              ),
              onChanged: (value) {
                main.value = value;
                setState(() {});
              },
              controller: headerValueController[i],
            ),
          ),
        );
        if (main.value.isNotEmpty) {
          formHeader.add(
            Center(
              child: QrImageView(
                data: main.value,
                version: QrVersions.auto,
                size: 300,
                gapless: false,
              ),
            ),
          );
        }
      } else if (main.headerType == 3) {
        formHeader.add(
          Padding(
            padding: EdgeInsets.all(8.0),
            child: TextFormField(
              decoration: InputDecoration(
                floatingLabelBehavior: FloatingLabelBehavior.always,
                border: OutlineInputBorder(),
                labelText: global.language('bill_text'),
              ),
              onChanged: (value) {
                main.value = value;
                setState(() {});
              },
              controller: headerValueController[i],
            ),
          ),
        );
        if (main.value.isNotEmpty) {
          formHeader.add(
            Center(
              child: BarcodeWidget(
                width: 200,
                height: 100,
                barcode: Barcode.code128(),
                data: main.value,
                errorBuilder: (context, error) => Center(child: Text(error)),
              ),
            ),
          );
        }
      }
    }

    formHeader.add(const SizedBox(height: 20));

    return Container(
      margin: const EdgeInsets.all(5),
      padding: const EdgeInsets.all(10),
      height: 100,
      width: double.infinity,
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        boxShadow: [
          BoxShadow(
            color: Colors.grey.withOpacity(0.5),
            spreadRadius: 5,
            blurRadius: 7,
            offset: const Offset(0, 3),
          ),
        ],
      ),
      child: SingleChildScrollView(child: Column(children: formHeader)),
    );
  }

  Widget previewWidget() {
    List<Widget> headerWidget = [];
    List<Widget> detailWidget = [];
    List<Widget> totalWidget = [];
    for (int x = 0; x < header.length; x++) {
      var data = header[x];

      if (data.headerType == 0) {
        headerWidget.add(
          SizedBox(
            width: double.infinity,
            child: Text(
              global.activeLangName(data.description),
              textAlign: TextAlign.center,
            ),
          ),
        );
      } else if (data.headerType == 1) {
        if (imageHeaderWeb[x].isNotEmpty || data.imgUri.isNotEmpty) {
          headerWidget.add(
            SizedBox(
              width: double.infinity,
              child: Center(
                child: DecoratedBox(
                  decoration: BoxDecoration(
                    image: (imageHeaderWeb[x].isNotEmpty)
                        ? DecorationImage(
                            image: MemoryImage(imageHeaderWeb[x]),
                            fit: BoxFit.fill,
                          )
                        : (data.imgUri != '')
                        ? DecorationImage(
                            image: NetworkImage(data.imgUri),
                            fit: BoxFit.fill,
                          )
                        : null,
                  ),
                  child: const SizedBox(width: 80, height: 80),
                ),
              ),
            ),
          );
        }
      } else if (data.headerType == 2) {
        if (data.value.isNotEmpty) {
          headerWidget.add(
            Center(
              child: QrImageView(
                data: data.value,
                version: QrVersions.auto,
                size: 100,
                gapless: true,
              ),
            ),
          );
          headerWidget.add(
            Container(
              margin: const EdgeInsets.only(bottom: 5),
              child: Center(
                child: Text(data.value, style: const TextStyle(fontSize: 12)),
              ),
            ),
          );
        }
      } else if (data.headerType == 3) {
        if (data.value.isNotEmpty) {
          headerWidget.add(
            Center(
              child: BarcodeWidget(
                width: 180,
                height: 80,
                barcode: Barcode.code128(),
                data: data.value,
                errorBuilder: (context, error) => Center(child: Text(error)),
              ),
            ),
          );
        }
      }
    }
    List<Widget> detailHeader = [];
    List<Widget> detailBody = [];
    for (var data in detailColumn) {
      detailHeader.add(
        Expanded(
          flex: data.width.toInt(),
          child: Text(
            global.activeLangName(data.header_names),
            textAlign: (data.text_align == global.PrintColumnAlign.right)
                ? TextAlign.right
                : (data.text_align == global.PrintColumnAlign.center)
                ? TextAlign.center
                : TextAlign.left,
          ),
        ),
      );
      //&item_qty& &item_name&/&item_unit_name& &item_price_and_symbol& &item_discount&

      String showDetail = data.command;
      showDetail = showDetail.replaceAll("&item_qty&", global.language("quantity"));
      showDetail = showDetail.replaceAll("&item_name&", global.language("product_name"));
      showDetail = showDetail.replaceAll("&item_unit_name&", global.language("unit_name"));
      showDetail = showDetail.replaceAll("&item_price_and_symbol&", global.language("price"));
      showDetail = showDetail.replaceAll("&item_discount&", global.language("discount"));
      showDetail = showDetail.replaceAll("&item_total_amount&", global.language("total_amount"));
      detailBody.add(
        Expanded(
          flex: data.width.toInt(),
          child: Text(
            showDetail,
            style: TextStyle(
              decoration: (data.decoration_underline)
                  ? TextDecoration.underline
                  : TextDecoration.none,
              fontSize: data.font_size,
              fontWeight: (data.font_weight_bold)
                  ? FontWeight.bold
                  : FontWeight.normal,
              fontStyle: (data.font_style_italic)
                  ? FontStyle.italic
                  : FontStyle.normal,
            ),
            textAlign: (data.text_align == global.PrintColumnAlign.right)
                ? TextAlign.right
                : (data.text_align == global.PrintColumnAlign.center)
                ? TextAlign.center
                : TextAlign.left,
          ),
        ),
      );
    }
    List<Widget> detailExtra = [];
    for (var data in detailExtraColumn) {
      String showDetail = data.command;
      showDetail = showDetail.replaceAll("&item_extra_name&", global.language("extra_name"));
      showDetail = showDetail.replaceAll("&item_extra_qty&", global.language("quantity"));
      showDetail = showDetail.replaceAll("&item_extra_unit_name&", global.language("unit_name"));
      showDetail = showDetail.replaceAll("&item_extra_price&", global.language("price"));
      showDetail = showDetail.replaceAll("&item_extra_total_amount&", global.language("total_amount"));
      detailExtra.add(
        Expanded(
          flex: data.width.toInt(),
          child: Text(
            showDetail,
            style: TextStyle(
              decoration: (data.decoration_underline)
                  ? TextDecoration.underline
                  : TextDecoration.none,
              fontSize: data.font_size,
              fontWeight: (data.font_weight_bold)
                  ? FontWeight.bold
                  : FontWeight.normal,
              fontStyle: (data.font_style_italic)
                  ? FontStyle.italic
                  : FontStyle.normal,
            ),
            textAlign: (data.text_align == global.PrintColumnAlign.right)
                ? TextAlign.right
                : (data.text_align == global.PrintColumnAlign.center)
                ? TextAlign.center
                : TextAlign.left,
          ),
        ),
      );
    }

    for (var main in detailTotalColumn) {
      List<Widget> detailRowTotal = [];
      for (var data in main) {
        String showDetail = data.command;
        showDetail = showDetail.replaceAll("&item_name&", global.language("total"));
        showDetail = showDetail.replaceAll("&total_piece&", global.language("price"));
        showDetail = showDetail.replaceAll("&total_amount&", global.language("total_amount"));
        detailRowTotal.add(
          Expanded(
            flex: data.width.toInt(),
            child: Text(
              showDetail,
              textAlign: (data.text_align == global.PrintColumnAlign.right)
                  ? TextAlign.right
                  : (data.text_align == global.PrintColumnAlign.center)
                  ? TextAlign.center
                  : TextAlign.left,
            ),
          ),
        );
      }
      totalWidget.add(Row(children: detailRowTotal));
    }

    List<Widget> footerWidget = [];
    for (int x = 0; x < footer.length; x++) {
      var data = footer[x];
      if (data.footerType == 0) {
        footerWidget.add(
          SizedBox(
            width: double.infinity,
            child: Text(
              global.activeLangName(data.description),
              textAlign: TextAlign.center,
            ),
          ),
        );
      } else if (data.footerType == 1) {
        if (imageWeb[x].isNotEmpty || data.imgUri.isNotEmpty) {
          footerWidget.add(
            SizedBox(
              width: double.infinity,
              child: Center(
                child: DecoratedBox(
                  decoration: BoxDecoration(
                    image: (imageWeb[x].isNotEmpty)
                        ? DecorationImage(
                            image: MemoryImage(imageWeb[x]),
                            fit: BoxFit.fill,
                          )
                        : (data.imgUri != '')
                        ? DecorationImage(
                            image: NetworkImage(data.imgUri),
                            fit: BoxFit.fill,
                          )
                        : null,
                  ),
                  child: const SizedBox(width: 80, height: 80),
                ),
              ),
            ),
          );
        }
      } else if (data.footerType == 2) {
        if (data.value.isNotEmpty) {
          footerWidget.add(
            Center(
              child: QrImageView(
                data: data.value,
                version: QrVersions.auto,
                size: 100,
                gapless: true,
              ),
            ),
          );
          footerWidget.add(
            Container(
              margin: const EdgeInsets.only(bottom: 5),
              child: Center(
                child: Text(data.value, style: const TextStyle(fontSize: 12)),
              ),
            ),
          );
        }
      } else if (data.footerType == 3) {
        if (data.value.isNotEmpty) {
          footerWidget.add(
            Center(
              child: BarcodeWidget(
                width: 180,
                height: 80,
                barcode: Barcode.code128(),
                data: data.value,
                errorBuilder: (context, error) => Center(child: Text(error)),
              ),
            ),
          );
        }
      }
    }
    detailWidget.add(
      const Padding(
        padding: EdgeInsets.only(top: 5, bottom: 1),
        child: Divider(color: Colors.black, height: 2),
      ),
    );
    detailWidget.add(Row(children: detailHeader));
    detailWidget.add(
      const Padding(
        padding: EdgeInsets.only(top: 5, bottom: 3),
        child: Divider(color: Colors.black, height: 2),
      ),
    );
    detailWidget.add(Row(children: detailBody));
    detailWidget.add(Row(children: detailExtra));

    List<Widget> preViewDetail = [];
    preViewDetail.addAll(headerWidget);
    preViewDetail.add(Expanded(child: Column(children: detailWidget)));
    preViewDetail.add(
      const Padding(
        padding: EdgeInsets.only(top: 5, bottom: 1),
        child: Divider(color: Colors.black, height: 2),
      ),
    );
    preViewDetail.addAll(totalWidget);
    preViewDetail.add(
      const Padding(
        padding: EdgeInsets.only(top: 5, bottom: 3),
        child: Divider(color: Colors.black, height: 2),
      ),
    );
    preViewDetail.addAll(footerWidget);

    Widget containerPreview = Container(
      padding: const EdgeInsets.all(10),
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(10),
        boxShadow: [
          BoxShadow(
            color: Colors.grey.withOpacity(0.5),
            spreadRadius: 5,
            blurRadius: 7,
            offset: const Offset(0, 3), // changes position of shadow
          ),
        ],
      ),
      child: Column(children: preViewDetail),
    );

    return Container(
      margin: const EdgeInsets.all(5),
      width: double.infinity,
      child: Column(
        children: [
          Row(
            children: [
              Container(
                margin: EdgeInsets.all(2),
                child: ElevatedButton(
                  onPressed: () {
                    const double paperWidthMM = 58.0;
                    const double mmToPixel = 3.7795275591;
                    double paperWidthPixels = (paperWidthMM * mmToPixel);
                    showDialog(
                      context: context,
                      builder: (BuildContext context) {
                        return AlertDialog(
                          content: SizedBox(
                            width: paperWidthPixels.toDouble(),
                            child: containerPreview,
                          ),
                          actions: [
                            ElevatedButton(
                              child: Text(global.language('bill_close')),
                              onPressed: () {
                                Navigator.of(context).pop();
                              },
                            ),
                          ],
                        );
                      },
                    );
                  },
                  child: const Text("58mm"),
                ),
              ),
              Container(
                margin: EdgeInsets.all(2),
                child: ElevatedButton(
                  onPressed: () {
                    const double paperWidthMM = 80.0;
                    const double mmToPixel = 3.7795275591;
                    double paperWidthPixels = (paperWidthMM * mmToPixel);
                    showDialog(
                      context: context,
                      builder: (BuildContext context) {
                        return AlertDialog(
                          content: SizedBox(
                            width: paperWidthPixels.toDouble(),
                            child: containerPreview,
                          ),
                          actions: [
                            ElevatedButton(
                              child: Text(global.language('bill_close')),
                              onPressed: () {
                                Navigator.of(context).pop();
                              },
                            ),
                          ],
                        );
                      },
                    );
                  },
                  child: const Text("80mm"),
                ),
              ),
            ],
          ),
          Expanded(child: containerPreview),
        ],
      ),
    );
  }
}
