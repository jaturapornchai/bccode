// ignore_for_file: prefer_const_constructors, prefer_const_literals_to_create_immutables, non_constant_identifier_names

import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/bloc/product_category/product_category_bloc.dart';
import 'package:smlaicloud/bloc/option/option_bloc.dart';
import 'package:smlaicloud/model/product_category_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/option_model.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:loader_overlay/loader_overlay.dart';
import 'package:smlaicloud/screens/product/product_image.dart';
import 'package:smlaicloud/screens/product/product_screen.dart';
import 'package:smlaicloud/bloc/inventory/inventory_bloc.dart';
import 'package:smlaicloud/components/appbar.dart';
import 'package:smlaicloud/components/background_main.dart';
import 'package:smlaicloud/components/textfield_input.dart';
import 'package:smlaicloud/model/inventory_model.dart';
import 'package:smlaicloud/utils/util.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

class ProductAdd extends StatefulWidget {
  final String? guidfixed;

  const ProductAdd({super.key, this.guidfixed = ""});

  @override
  State<ProductAdd> createState() => _ProductAddState();
}

class _ProductAddState extends State<ProductAdd> {
  // ignore: unused_field
  final List<String> _values = [];
  final List<OptionModel> _option = [];
  List<OptionModel> _optionDetail = [];
  // ignore: unused_field
  final TextEditingController _textEditingController = TextEditingController();

  List<ImageUpload> _images = [];

  // ignore: unused_field
  bool _reloadImage = false;

  String? inputTag;
  String? shopid;

  String? groupcode;

  ProductCategoryModel selectCate = ProductCategoryModel(
    guidfixed: "",
    parentguid: "",
    parentguidall: "",
    imageuri: "",
    useimageorcolor: true,
    isdisabled: false,
    colorselect: "",
    colorselecthex: "",
    childcount: 0,
    names: [],
    xsorts: [],
    codelist: [],
  );

  final _barcode = TextEditingController();
  final _name1 = TextEditingController();
  final _name2 = TextEditingController();
  final _name3 = TextEditingController();
  final _name4 = TextEditingController();
  final _name5 = TextEditingController();

  final _description1 = TextEditingController();
  final _description2 = TextEditingController();
  final _description3 = TextEditingController();
  final _description4 = TextEditingController();
  final _description5 = TextEditingController();

  final _itemcode = TextEditingController();
  final _itemunitcode = TextEditingController();
  final _itemunitstd = TextEditingController();
  final _itemunitdiv = TextEditingController();

  final _unitname1 = TextEditingController();
  final _unitname2 = TextEditingController();
  final _unitname3 = TextEditingController();
  final _unitname4 = TextEditingController();
  final _unitname5 = TextEditingController();

  final _price = TextEditingController();

  bool _displayNewTextFieldName2 = false;
  bool _displayNewTextFieldName3 = false;
  bool _displayNewTextFieldName4 = false;
  bool _displayNewTextFieldName5 = false;

  bool _displayNewTextFielddescription2 = false;
  bool _displayNewTextFielddescription3 = false;
  bool _displayNewTextFielddescription4 = false;
  bool _displayNewTextFielddescription5 = false;

  bool _displayNewUnitCode2 = false;
  bool _displayNewUnitCode3 = false;
  bool _displayNewUnitCode4 = false;
  bool _displayNewUnitCode5 = false;

  bool createMode = true;

  // ignore: unused_field
  final int _perPage = 100;
  // ignore: unused_field
  final String _search = "";

  // _onDelete(index) {
  //   setState(() {
  //     _values.removeAt(index);
  //   });
  // }

  @override
  void initState() {
    super.initState();

    widget.guidfixed.toString() != "" ? createMode = false : createMode = true;

    if (!createMode) {
      context.read<InventoryBloc>().add(
        ListInventoryById(id: widget.guidfixed.toString()),
      );
    }
    //load  category
    //context.read<CategoryBloc>().add(ListCategoryLoadDropDown());

    //load  Option
    context.read<OptionBloc>().add(ListOptionLoadSelect());

    shopid = global.getShopId();
  }

  @override
  Widget build(BuildContext context) {
    // Size size = MediaQuery.of(context).size;
    return MultiBlocListener(
      listeners: [
        BlocListener<InventoryBloc, InventoryState>(
          listener: (context, state) {
            if (state is InventorySearchLoadSuccess) {
              setState(() {
                groupcode = state.inventory.groupcode;
                _barcode.text = state.inventory.barcode;
                _name1.text = state.inventory.name1;
                _name2.text = state.inventory.name2;
                _name3.text = state.inventory.name3;
                _name4.text = state.inventory.name4;
                _name5.text = state.inventory.name5;
                _description1.text = state.inventory.description1;
                _description2.text = state.inventory.description2;
                _description3.text = state.inventory.description3;
                _description4.text = state.inventory.description4;
                _description5.text = state.inventory.description5;
                _unitname1.text = state.inventory.unitname1;
                _unitname2.text = state.inventory.unitname2;
                _unitname3.text = state.inventory.unitname3;
                _unitname4.text = state.inventory.unitname4;
                _unitname5.text = state.inventory.unitname5;
                _itemcode.text = state.inventory.itemcode;
                _itemunitcode.text = state.inventory.itemunitcode;
                _itemunitstd.text = state.inventory.itemunitstd.toString();
                _itemunitdiv.text = state.inventory.itemunitdiv.toString();
                _price.text = state.inventory.price.toString();

                // Debug log for options (ป้องกัน null)
                try {
                  AppLogger.debug(
                    'Options from state: ${state.inventory.options}',
                  );
                  AppLogger.debug(
                    'Options type: ${state.inventory.options.runtimeType}',
                  );
                  AppLogger.debug(
                    'Options length: ${state.inventory.options.length}',
                  );
                } catch (e) {
                  AppLogger.error('Error in options debug log: $e');
                }

                // จัดการกรณีที่ options เป็น null หรือ empty จาก JSON
                try {
                  final inventoryOptions = state.inventory.options;
                  AppLogger.debug(
                    'ตรวจสอบ options type: ${inventoryOptions.runtimeType}',
                  );
                  AppLogger.debug('ตรวจสอบ options value: $inventoryOptions');

                  if (inventoryOptions.isEmpty) {
                    AppLogger.warning('WARNING: Options list is empty');
                  }
                  // สร้าง list ใหม่และจัดการกับข้อมูล
                  _optionDetail = inventoryOptions.isNotEmpty
                      ? List<OptionModel>.from(inventoryOptions)
                      : <OptionModel>[];

                  AppLogger.debug(
                    'สำเร็จ: ได้ options จำนวน ${_optionDetail.length} รายการ',
                  );
                } catch (e, stackTrace) {
                  AppLogger.error('❌ ERROR ตอนแปลง options: $e');
                  AppLogger.error('Stack trace: $stackTrace');
                  AppLogger.error('State inventory: ${state.inventory}');
                  _optionDetail = <OptionModel>[];
                }

                _displayNewTextFieldName2 = (state.inventory.name2.isNotEmpty);
                _displayNewTextFieldName3 = (state.inventory.name3.isNotEmpty);
                _displayNewTextFieldName4 = (state.inventory.name4.isNotEmpty);
                _displayNewTextFieldName5 = (state.inventory.name5.isNotEmpty);

                _displayNewTextFielddescription2 =
                    (state.inventory.description2.isNotEmpty);
                _displayNewTextFielddescription3 =
                    (state.inventory.description3.isNotEmpty);
                _displayNewTextFielddescription4 =
                    (state.inventory.description4.isNotEmpty);
                _displayNewTextFielddescription5 =
                    (state.inventory.description5.isNotEmpty);

                _displayNewUnitCode2 = (state.inventory.unitname2.isNotEmpty);
                _displayNewUnitCode3 = (state.inventory.unitname3.isNotEmpty);
                _displayNewUnitCode4 = (state.inventory.unitname4.isNotEmpty);
                _displayNewUnitCode5 = (state.inventory.unitname5.isNotEmpty);

                // Debug log และจัดการ images
                try {
                  final inventoryImages = state.inventory.images;
                  AppLogger.debug(
                    'ตรวจสอบ images type: ${inventoryImages.runtimeType}',
                  );
                  AppLogger.debug('ตรวจสอบ images value: $inventoryImages');
                  AppLogger.debug(
                    'ตรวจสอบ images length: ${inventoryImages.length}',
                  );

                  // จัดการกรณีที่ images เป็น null หรือ empty จาก JSON
                  // สร้าง list ใหม่และจัดการกับข้อมูล
                  _images = inventoryImages.isNotEmpty
                      ? List<ImageUpload>.from(inventoryImages)
                      : <ImageUpload>[];

                  AppLogger.debug(
                    'สำเร็จ: ได้ images จำนวน ${_images.length} รายการ',
                  );
                } catch (e, stackTrace) {
                  AppLogger.error('❌ ERROR ตอนแปลง images: $e');
                  AppLogger.error('Stack trace: $stackTrace');
                  AppLogger.error('State inventory: ${state.inventory}');
                  _images = <ImageUpload>[];
                }
              });
            } else if (state is InventoryDeleteSuccess) {
              Navigator.of(context).pushReplacement(
                MaterialPageRoute(builder: (_) => const ProductScreen()),
              );
            } else if (state is InventoryFormSaveSuccess) {
              showDeleteDialog('save');
            } else if (state is InventoryUpdateSuccess) {
              showDeleteDialog('update');
            }
          },
        ),
        BlocListener<ProductCategoryBloc, ProductCategoryState>(
          listener: (context, state) {
            /*if (state is CategoryLoadDropDownSuccess) {
              List<CategoryModel> _cat = state.category;

              CategoryModel result = _cat.firstWhere(
                  (element) => element.guidfixed == categoryguid,
                  orElse: () => selectCate);
              setState(() {
                selectCate = CategoryModel(
                  guidfixed: result.guidfixed,
                  parentGuid: "",
                  name1: result.name1,
                );
              });
            }*/
          },
        ),
        BlocListener<OptionBloc, OptionState>(
          listener: (context, state) {
            if (state is OptionLoadSelectSuccess) {
              if (state.option.isNotEmpty) {
                //_option = state.option;
              }
            }
          },
        ),
      ],
      child: Scaffold(
        appBar: BaseAppBar(
          title: (createMode)
              ? Text(global.language("product_add"))
              : Text(global.language("product_edit")),
          leading: IconButton(
            color: Colors.black,
            icon: const Icon(Icons.arrow_back),
            onPressed: () {
              Navigator.of(context).pushReplacement(
                MaterialPageRoute(builder: (_) => ProductScreen()),
              );
            },
          ),
          widgets: <Widget>[],
          appBar: AppBar(),
        ),
        body: LoaderOverlay(
          overlayColor: Colors.black,
          child: BackgroundMain(
            child: (Util.isLandscape(context))
                ? Row(
                    children: [
                      Expanded(
                        child: SingleChildScrollView(
                          child: Padding(
                            padding: const EdgeInsets.all(8.0),
                            child: selectImage(),
                          ),
                        ),
                      ),
                      Expanded(
                        child: SingleChildScrollView(
                          child: Padding(
                            padding: const EdgeInsets.all(8.0),
                            child: inputData(),
                          ),
                        ),
                      ),
                      (!createMode)
                          ? Container(
                              width: double.infinity,
                              margin: EdgeInsets.all(8.0),
                              child: ElevatedButton.icon(
                                icon: const Icon(
                                  Icons.delete,
                                  color: Colors.white,
                                  size: 24.0,
                                ),
                                label: Text(
                                  global.language('product_delete_data'),
                                ),
                                style: ElevatedButton.styleFrom(
                                  backgroundColor: Colors.red,
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 50,
                                    vertical: 20,
                                  ),
                                ),
                                onPressed: () {
                                  showDeleteDialog(widget.guidfixed.toString());
                                },
                              ),
                            )
                          : Container(),
                    ],
                  )
                : Column(
                    children: [
                      Expanded(
                        child: SingleChildScrollView(
                          child: Padding(
                            padding: EdgeInsets.all(8.0),
                            child: Column(
                              children: [
                                selectOption(),
                                inputData(),
                                selectImage(),
                                (!createMode)
                                    ? Container(
                                        width: double.infinity,
                                        margin: EdgeInsets.all(8.0),
                                        child: ElevatedButton.icon(
                                          icon: const Icon(
                                            Icons.delete,
                                            color: Colors.white,
                                            size: 24.0,
                                          ),
                                          label: Text(
                                            global.language('product_delete'),
                                          ),
                                          style: ElevatedButton.styleFrom(
                                            backgroundColor: Colors.red,
                                            padding: const EdgeInsets.symmetric(
                                              horizontal: 50,
                                              vertical: 20,
                                            ),
                                          ),
                                          onPressed: () {
                                            showDeleteDialog(
                                              widget.guidfixed.toString(),
                                            );
                                          },
                                        ),
                                      )
                                    : Container(),
                              ],
                            ),
                          ),
                        ),
                      ),
                      BlocBuilder<InventoryBloc, InventoryState>(
                        builder: (context, state) {
                          if (state is InventoryFormSaveInProgress) {
                            context.loaderOverlay.show(
                              widgetBuilder: (progress) {
                                return const ReconnectingOverlay();
                              },
                            );
                          } else {
                            context.loaderOverlay.hide();
                          }
                          return (state is InventoryFormSaveFailure)
                              ? Text(
                                  state.message,
                                  style: const TextStyle(color: Colors.red),
                                )
                              : Container();
                        },
                      ),
                    ],
                  ),
          ),
        ),
        floatingActionButton: FloatingActionButton(
          onPressed: () {
            if (_price.text == '') {
              _price.text = '0';
            }
            if (_itemunitstd.text == '') {
              _itemunitstd.text = '0';
            }
            if (_itemunitdiv.text == '') {
              _itemunitdiv.text = '0';
            }

            InventoryModel inventory = InventoryModel(
              groupcode: selectCate.guidfixed,
              itemguid: shopid.toString() + _barcode.text,
              guidfixed: widget.guidfixed.toString(),
              barcode: _barcode.text,
              name1: _name1.text,
              name2: _name2.text,
              name3: _name2.text,
              name4: _name2.text,
              name5: _name2.text,
              description1: _description1.text,
              description2: _description2.text,
              description3: _description3.text,
              description4: _description4.text,
              description5: _description5.text,
              itemcode: _itemcode.text,
              itemunitcode: _itemunitcode.text,
              itemunitstd: double.parse(_itemunitstd.text),
              itemunitdiv: double.parse(_itemunitdiv.text),
              unitname1: _unitname1.text,
              unitname2: _unitname2.text,
              unitname3: _unitname3.text,
              unitname4: _unitname4.text,
              unitname5: _unitname5.text,
              price: double.parse(_price.text),
              options: _optionDetail,
              images: _images,
            );

            if (createMode) {
              context.read<InventoryBloc>().add(
                InventorySaved(inventory: inventory),
              );
            } else {
              context.read<InventoryBloc>().add(
                InventoryUpdate(inventory: inventory),
              );
            }
          },
          backgroundColor: Colors.green,
          child: const Icon(Icons.save),
        ),
      ),
    );
  }

  void showDeleteDialog(String id) {
    // set up the buttons
    Widget cancelButton = TextButton(
      child: Text(global.language("cancel")),
      onPressed: () {
        Navigator.of(context).pop();
      },
    );
    Widget continueButton = TextButton(
      child: Text(global.language("confirm")),
      onPressed: () {
        if (id == 'save') {
          Navigator.of(context).pop();
          Navigator.of(context).pushReplacement(
            MaterialPageRoute(builder: (_) => const ProductScreen()),
          );
        } else if (id == 'update') {
          Navigator.of(context).pop();
          Navigator.of(context).pushReplacement(
            MaterialPageRoute(builder: (_) => const ProductScreen()),
          );
        } else {
          context.read<InventoryBloc>().add(InventoryDelete(id: id));
          Navigator.of(context).pop();
        }
      },
    );
    // set up the AlertDialog
    AlertDialog alert = AlertDialog(
      title: Text(global.language("system_notification")),
      content: (id == 'save')
          ? Text(global.language("product_save_success"))
          : (id == 'update')
          ? Text(global.language("product_update_success"))
          : Text(global.language("product_delete_confirm")),
      actions: [cancelButton, continueButton],
    );
    // show the dialog
    showDialog(
      context: context,
      builder: (BuildContext context) {
        return alert;
      },
    );
  }

  Widget inputData() {
    return Card(
      child: Padding(
        padding: EdgeInsets.all(10.0),
        child: Column(
          children: [
            //Category
            Container(
              padding: EdgeInsets.only(left: 8.0, right: 8.0, top: 15.0),
              child: Row(
                children: [
                  (Util.isLandscape(context))
                      ? Expanded(
                          flex: 1,
                          child: Text(global.language('category')),
                        )
                      : Container(),
                  /*Expanded(
                    flex: 3,
                    child: BlocBuilder<CategoryBloc, CategoryState>(
                      builder: (context, state) {
                        return DropdownSearch<Category>(
                          dropdownSearchDecoration: InputDecoration(
                            contentPadding: const EdgeInsets.only(left: 10),
                            border: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(10.0),
                            ),
                          ),
                          mode: Mode.DIALOG,
                          showSearchBox: true,
                          itemAsString: (Category? u) => u!.toShowLabel(),
                          items: state is CategoryLoadDropDownSuccess
                              ? state.category
                              : [],
                          selectedItem: selectCate,
                          onChanged: (cate) {
                            setState(() {
                              selectCate = cate!;
                            });
                          },
                        );
                      },
                    ),
                  ),*/
                ],
              ),
            ),

            //barcode
            Container(
              padding: EdgeInsets.only(left: 8.0, right: 8.0, top: 15.0),
              child: Row(
                children: [
                  (Util.isLandscape(context))
                      ? Expanded(
                          flex: 1,
                          child: Text(global.language('barcode')),
                        )
                      : Container(),
                  Expanded(
                    flex: 3,
                    child: TextFieldInput(
                      keyboardType: TextInputType.text,
                      controller: _barcode,
                      labelText: 'barcode',
                    ),
                  ),
                ],
              ),
            ),

            // ชื่อสินค้า
            Container(
              padding: EdgeInsets.only(left: 8.0, right: 8.0, top: 15.0),
              child: Row(
                children: [
                  (Util.isLandscape(context))
                      ? Expanded(
                          flex: 1,
                          child: Text(global.language('product_name')),
                        )
                      : Container(),
                  Expanded(
                    flex: 2,
                    child: TextFieldInput(
                      keyboardType: TextInputType.text,
                      controller: _name1,
                      labelText: global.language('product_name'),
                    ),
                  ),
                  Expanded(
                    flex: 1,
                    child: SizedBox(
                      height: 30.0,
                      child: IconButton(
                        padding: EdgeInsets.zero,
                        iconSize: 30.0,
                        icon: Icon(Icons.add_circle_outline),
                        color: _displayNewTextFieldName2 == true
                            ? Colors.grey
                            : Colors.blue,
                        onPressed: () {
                          _displayNewTextFieldName2 != true
                              ? setState(() {
                                  _displayNewTextFieldName2 = true;
                                })
                              : null;
                        },
                      ),
                    ),
                  ),
                ],
              ),
            ),
            Visibility(
              visible: _displayNewTextFieldName2,
              child: Container(
                padding: const EdgeInsets.only(
                  left: 8.0,
                  right: 8.0,
                  top: 15.0,
                ),
                child: Row(
                  children: [
                    (Util.isLandscape(context))
                        ? Expanded(
                            flex: 1,
                            child: Text("${global.language('product_name')} 2"),
                          )
                        : Container(),
                    Expanded(
                      flex: 2,
                      child: TextFieldInput(
                        controller: _name2,
                        keyboardType: TextInputType.text,
                        labelText: '${global.language('product_name')} 2',
                      ),
                    ),
                    Expanded(
                      flex: 1,
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          SizedBox(
                            height: 30.0,
                            child: IconButton(
                              padding: EdgeInsets.zero,
                              iconSize: 30.0,
                              icon: Icon(Icons.add_circle_outline),
                              color: _displayNewTextFieldName3 == true
                                  ? Colors.grey
                                  : Colors.blue,
                              onPressed: () {
                                _displayNewTextFieldName3 != true
                                    ? setState(() {
                                        _displayNewTextFieldName3 = true;
                                      })
                                    : null;
                              },
                            ),
                          ),
                          SizedBox(
                            height: 30.0,
                            child: IconButton(
                              padding: EdgeInsets.zero,
                              iconSize: 30.0,
                              icon: Icon(Icons.remove_circle_outline),
                              color: Colors.red,
                              onPressed: () {
                                setState(() {
                                  _name2.clear();
                                  _displayNewTextFieldName2 = false;
                                });
                              },
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
            Visibility(
              visible: _displayNewTextFieldName3,
              child: Container(
                padding: const EdgeInsets.only(
                  left: 8.0,
                  right: 8.0,
                  top: 15.0,
                ),
                child: Row(
                  children: [
                    (Util.isLandscape(context))
                        ? Expanded(
                            flex: 1,
                            child: Text(global.language('product_name_3')),
                          )
                        : Container(),
                    Expanded(
                      flex: 2,
                      child: TextFieldInput(
                        controller: _name3,
                        keyboardType: TextInputType.text,
                        labelText: global.language('product_name_3'),
                      ),
                    ),
                    Expanded(
                      flex: 1,
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          SizedBox(
                            height: 30.0,
                            child: IconButton(
                              padding: EdgeInsets.zero,
                              iconSize: 30.0,
                              icon: Icon(Icons.add_circle_outline),
                              color: _displayNewTextFieldName4 == true
                                  ? Colors.grey
                                  : Colors.blue,
                              onPressed: () {
                                _displayNewTextFieldName4 != true
                                    ? setState(() {
                                        _displayNewTextFieldName4 = true;
                                      })
                                    : null;
                              },
                            ),
                          ),
                          SizedBox(
                            height: 30.0,
                            child: IconButton(
                              padding: EdgeInsets.zero,
                              iconSize: 30.0,
                              icon: Icon(Icons.remove_circle_outline),
                              color: Colors.red,
                              onPressed: () {
                                setState(() {
                                  _name3.clear();
                                  _displayNewTextFieldName3 = false;
                                });
                              },
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
            Visibility(
              visible: _displayNewTextFieldName4,
              child: Container(
                padding: const EdgeInsets.only(
                  left: 8.0,
                  right: 8.0,
                  top: 15.0,
                ),
                child: Row(
                  children: [
                    (Util.isLandscape(context))
                        ? Expanded(
                            flex: 1,
                            child: Text(global.language('product_name_4')),
                          )
                        : Container(),
                    Expanded(
                      flex: 2,
                      child: TextFieldInput(
                        controller: _name4,
                        keyboardType: TextInputType.text,
                        labelText: global.language('product_name_4'),
                      ),
                    ),
                    Expanded(
                      flex: 1,
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          SizedBox(
                            height: 30.0,
                            child: IconButton(
                              padding: EdgeInsets.zero,
                              iconSize: 30.0,
                              icon: Icon(Icons.add_circle_outline),
                              color: _displayNewTextFieldName5 == true
                                  ? Colors.grey
                                  : Colors.blue,
                              onPressed: () {
                                _displayNewTextFieldName5 != true
                                    ? setState(() {
                                        _displayNewTextFieldName5 = true;
                                      })
                                    : null;
                              },
                            ),
                          ),
                          SizedBox(
                            height: 30.0,
                            child: IconButton(
                              padding: EdgeInsets.zero,
                              iconSize: 30.0,
                              icon: Icon(Icons.remove_circle_outline),
                              color: Colors.red,
                              onPressed: () {
                                setState(() {
                                  _name4.clear();
                                  _displayNewTextFieldName4 = false;
                                });
                              },
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
            Visibility(
              visible: _displayNewTextFieldName5,
              child: Container(
                padding: const EdgeInsets.only(
                  left: 8.0,
                  right: 8.0,
                  top: 15.0,
                ),
                child: Row(
                  children: [
                    (Util.isLandscape(context))
                        ? Expanded(
                            flex: 1,
                            child: Text(global.language('product_name_5')),
                          )
                        : Container(),
                    Expanded(
                      flex: 2,
                      child: TextFieldInput(
                        controller: _name5,
                        keyboardType: TextInputType.text,
                        labelText: global.language('product_name_5'),
                      ),
                    ),
                    Expanded(
                      flex: 1,
                      child: SizedBox(
                        height: 30.0,
                        child: IconButton(
                          padding: EdgeInsets.zero,
                          iconSize: 30.0,
                          icon: Icon(Icons.remove_circle_outline),
                          color: Colors.red,
                          onPressed: () {
                            setState(() {
                              _name5.clear();
                              _displayNewTextFieldName5 = false;
                            });
                          },
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
            // รายละเอียด
            Container(
              padding: EdgeInsets.only(left: 8.0, right: 8.0, top: 15.0),
              child: Row(
                children: [
                  (Util.isLandscape(context))
                      ? Expanded(
                          flex: 1,
                          child: Text(global.language('description_1')),
                        )
                      : Container(),
                  Expanded(
                    flex: 2,
                    child: TextFieldInput(
                      keyboardType: TextInputType.text,
                      controller: _description1,
                      labelText: global.language('description_1'),
                    ),
                  ),
                  Expanded(
                    flex: 1,
                    child: SizedBox(
                      height: 30.0,
                      child: IconButton(
                        padding: EdgeInsets.zero,
                        iconSize: 30.0,
                        icon: Icon(Icons.add_circle_outline),
                        color: _displayNewTextFielddescription2 == true
                            ? Colors.grey
                            : Colors.blue,
                        onPressed: () {
                          _displayNewTextFielddescription2 != true
                              ? setState(() {
                                  _displayNewTextFielddescription2 = true;
                                })
                              : null;
                        },
                      ),
                    ),
                  ),
                ],
              ),
            ),
            Visibility(
              visible: _displayNewTextFielddescription2,
              child: Container(
                padding: const EdgeInsets.only(
                  left: 8.0,
                  right: 8.0,
                  top: 15.0,
                ),
                child: Row(
                  children: [
                    (Util.isLandscape(context))
                        ? Expanded(
                            flex: 1,
                            child: Text(global.language('description_2')),
                          )
                        : Container(),
                    Expanded(
                      flex: 2,
                      child: TextFieldInput(
                        controller: _description2,
                        keyboardType: TextInputType.text,
                        labelText: global.language('description_2'),
                      ),
                    ),
                    Expanded(
                      flex: 1,
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          SizedBox(
                            height: 30.0,
                            child: IconButton(
                              padding: EdgeInsets.zero,
                              iconSize: 30.0,
                              icon: Icon(Icons.add_circle_outline),
                              color: _displayNewTextFielddescription3 == true
                                  ? Colors.grey
                                  : Colors.blue,
                              onPressed: () {
                                _displayNewTextFielddescription3 != true
                                    ? setState(() {
                                        _displayNewTextFielddescription3 = true;
                                      })
                                    : null;
                              },
                            ),
                          ),
                          SizedBox(
                            height: 30.0,
                            child: IconButton(
                              padding: EdgeInsets.zero,
                              iconSize: 30.0,
                              icon: Icon(Icons.remove_circle_outline),
                              color: Colors.red,
                              onPressed: () {
                                setState(() {
                                  _description2.clear();
                                  _displayNewTextFielddescription2 = false;
                                });
                              },
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
            Visibility(
              visible: _displayNewTextFielddescription3,
              child: Container(
                padding: const EdgeInsets.only(
                  left: 8.0,
                  right: 8.0,
                  top: 15.0,
                ),
                child: Row(
                  children: [
                    (Util.isLandscape(context))
                        ? Expanded(
                            flex: 1,
                            child: Text(global.language('description_3')),
                          )
                        : Container(),
                    Expanded(
                      flex: 2,
                      child: TextFieldInput(
                        controller: _description3,
                        keyboardType: TextInputType.text,
                        labelText: global.language('description_3'),
                      ),
                    ),
                    Expanded(
                      flex: 1,
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          SizedBox(
                            height: 30.0,
                            child: IconButton(
                              padding: EdgeInsets.zero,
                              iconSize: 30.0,
                              icon: Icon(Icons.add_circle_outline),
                              color: _displayNewTextFielddescription4 == true
                                  ? Colors.grey
                                  : Colors.blue,
                              onPressed: () {
                                _displayNewTextFielddescription4 != true
                                    ? setState(() {
                                        _displayNewTextFielddescription4 = true;
                                      })
                                    : null;
                              },
                            ),
                          ),
                          SizedBox(
                            height: 30.0,
                            child: IconButton(
                              padding: EdgeInsets.zero,
                              iconSize: 30.0,
                              icon: Icon(Icons.remove_circle_outline),
                              color: Colors.red,
                              onPressed: () {
                                setState(() {
                                  _description3.clear();
                                  _displayNewTextFielddescription3 = false;
                                });
                              },
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
            Visibility(
              visible: _displayNewTextFielddescription4,
              child: Container(
                padding: const EdgeInsets.only(
                  left: 8.0,
                  right: 8.0,
                  top: 15.0,
                ),
                child: Row(
                  children: [
                    (Util.isLandscape(context))
                        ? Expanded(
                            flex: 1,
                            child: Text(global.language('description_4')),
                          )
                        : Container(),
                    Expanded(
                      flex: 2,
                      child: TextFieldInput(
                        controller: _description4,
                        keyboardType: TextInputType.text,
                        labelText: global.language('description_4'),
                      ),
                    ),
                    Expanded(
                      flex: 1,
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          SizedBox(
                            height: 30.0,
                            child: IconButton(
                              padding: EdgeInsets.zero,
                              iconSize: 30.0,
                              icon: Icon(Icons.add_circle_outline),
                              color: _displayNewTextFielddescription5 == true
                                  ? Colors.grey
                                  : Colors.blue,
                              onPressed: () {
                                _displayNewTextFielddescription5 != true
                                    ? setState(() {
                                        _displayNewTextFielddescription5 = true;
                                      })
                                    : null;
                              },
                            ),
                          ),
                          SizedBox(
                            height: 30.0,
                            child: IconButton(
                              padding: EdgeInsets.zero,
                              iconSize: 30.0,
                              icon: Icon(Icons.remove_circle_outline),
                              color: Colors.red,
                              onPressed: () {
                                setState(() {
                                  _description4.clear();
                                  _displayNewTextFielddescription4 = false;
                                });
                              },
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
            Visibility(
              visible: _displayNewTextFielddescription5,
              child: Container(
                padding: const EdgeInsets.only(
                  left: 8.0,
                  right: 8.0,
                  top: 15.0,
                ),
                child: Row(
                  children: [
                    (Util.isLandscape(context))
                        ? Expanded(
                            flex: 1,
                            child: Text(global.language('description_5')),
                          )
                        : Container(),
                    Expanded(
                      flex: 2,
                      child: TextFieldInput(
                        controller: _description5,
                        keyboardType: TextInputType.text,
                        labelText: global.language('description_5'),
                      ),
                    ),
                    Expanded(
                      flex: 1,
                      child: SizedBox(
                        height: 30.0,
                        child: IconButton(
                          padding: EdgeInsets.zero,
                          iconSize: 30.0,
                          icon: Icon(Icons.remove_circle_outline),
                          color: Colors.red,
                          onPressed: () {
                            setState(() {
                              _description5.clear();
                              _displayNewTextFielddescription5 = false;
                            });
                          },
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),

            //รหัสสินค้าอ้างอิง
            Container(
              padding: EdgeInsets.only(left: 8.0, right: 8.0, top: 15.0),
              child: Row(
                children: [
                  (Util.isLandscape(context))
                      ? Expanded(
                          flex: 1,
                          child: Text(global.language('item_reference_code')),
                        )
                      : Container(),
                  Expanded(
                    flex: 3,
                    child: TextFieldInput(
                      keyboardType: TextInputType.text,
                      controller: _itemcode,
                      labelText: global.language('item_reference_code'),
                    ),
                  ),
                ],
              ),
            ),
            //รหัสหน่วยนับอ้างอิง
            Container(
              padding: EdgeInsets.only(left: 8.0, right: 8.0, top: 15.0),
              child: Row(
                children: [
                  (Util.isLandscape(context))
                      ? Expanded(
                          flex: 1,
                          child: Text(global.language('unit_reference_code')),
                        )
                      : Container(),
                  Expanded(
                    flex: 3,
                    child: TextFieldInput(
                      keyboardType: TextInputType.text,
                      controller: _itemunitcode,
                      labelText: global.language('unit_reference_code'),
                    ),
                  ),
                ],
              ),
            ),
            //หน่วย (ตัวตั้ง)
            Container(
              padding: EdgeInsets.only(left: 8.0, right: 8.0, top: 15.0),
              child: Row(
                children: [
                  (Util.isLandscape(context))
                      ? Expanded(
                          flex: 1,
                          child: Text(global.language('unit_numerator')),
                        )
                      : Container(),
                  Expanded(
                    flex: 3,
                    child: TextFieldInput(
                      keyboardType: TextInputType.number,
                      controller: _itemunitstd,
                      labelText: global.language('unit_numerator'),
                    ),
                  ),
                ],
              ),
            ),
            //หน่วย (ตัวหาร)
            Container(
              padding: EdgeInsets.only(left: 8.0, right: 8.0, top: 15.0),
              child: Row(
                children: [
                  (Util.isLandscape(context))
                      ? Expanded(
                          flex: 1,
                          child: Text(global.language('unit_denominator')),
                        )
                      : Container(),
                  Expanded(
                    flex: 3,
                    child: TextFieldInput(
                      keyboardType: TextInputType.number,
                      controller: _itemunitdiv,
                      labelText: global.language('unit_denominator'),
                    ),
                  ),
                ],
              ),
            ),

            // ชื่อหน่วยนับ
            Container(
              padding: EdgeInsets.only(left: 8.0, right: 8.0, top: 15.0),
              child: Row(
                children: [
                  (Util.isLandscape(context))
                      ? Expanded(
                          flex: 1,
                          child: Text(global.language('unit_name')),
                        )
                      : Container(),
                  Expanded(
                    flex: 2,
                    child: TextFieldInput(
                      keyboardType: TextInputType.text,
                      controller: _unitname1,
                      labelText: global.language('unit_name'),
                    ),
                  ),
                  Expanded(
                    flex: 1,
                    child: SizedBox(
                      height: 30.0,
                      child: IconButton(
                        padding: EdgeInsets.zero,
                        iconSize: 30.0,
                        icon: Icon(Icons.add_circle_outline),
                        color: _displayNewUnitCode2 == true
                            ? Colors.grey
                            : Colors.blue,
                        onPressed: () {
                          _displayNewUnitCode2 != true
                              ? setState(() {
                                  _displayNewUnitCode2 = true;
                                })
                              : null;
                        },
                      ),
                    ),
                  ),
                ],
              ),
            ),
            Visibility(
              visible: _displayNewUnitCode2,
              child: Container(
                padding: const EdgeInsets.only(
                  left: 8.0,
                  right: 8.0,
                  top: 15.0,
                ),
                child: Row(
                  children: [
                    (Util.isLandscape(context))
                        ? Expanded(
                            flex: 1,
                            child: Text(global.language('unit_name_2')),
                          )
                        : Container(),
                    Expanded(
                      flex: 2,
                      child: TextFieldInput(
                        controller: _unitname2,
                        keyboardType: TextInputType.text,
                        labelText: global.language('unit_name_2'),
                      ),
                    ),
                    Expanded(
                      flex: 1,
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          SizedBox(
                            height: 30.0,
                            child: IconButton(
                              padding: EdgeInsets.zero,
                              iconSize: 30.0,
                              icon: Icon(Icons.add_circle_outline),
                              color: _displayNewUnitCode3 == true
                                  ? Colors.grey
                                  : Colors.blue,
                              onPressed: () {
                                _displayNewUnitCode3 != true
                                    ? setState(() {
                                        _displayNewUnitCode3 = true;
                                      })
                                    : null;
                              },
                            ),
                          ),
                          SizedBox(
                            height: 30.0,
                            child: IconButton(
                              padding: EdgeInsets.zero,
                              iconSize: 30.0,
                              icon: Icon(Icons.remove_circle_outline),
                              color: Colors.red,
                              onPressed: () {
                                setState(() {
                                  _unitname2.clear();
                                  _displayNewUnitCode2 = false;
                                });
                              },
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
            Visibility(
              visible: _displayNewUnitCode3,
              child: Container(
                padding: const EdgeInsets.only(
                  left: 8.0,
                  right: 8.0,
                  top: 15.0,
                ),
                child: Row(
                  children: [
                    (Util.isLandscape(context))
                        ? Expanded(
                            flex: 1,
                            child: Text(global.language('unit_name_3')),
                          )
                        : Container(),
                    Expanded(
                      flex: 2,
                      child: TextFieldInput(
                        controller: _unitname3,
                        keyboardType: TextInputType.text,
                        labelText: global.language('unit_name_3'),
                      ),
                    ),
                    Expanded(
                      flex: 1,
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          SizedBox(
                            height: 30.0,
                            child: IconButton(
                              padding: EdgeInsets.zero,
                              iconSize: 30.0,
                              icon: Icon(Icons.add_circle_outline),
                              color: _displayNewUnitCode4 == true
                                  ? Colors.grey
                                  : Colors.blue,
                              onPressed: () {
                                _displayNewUnitCode4 != true
                                    ? setState(() {
                                        _displayNewUnitCode4 = true;
                                      })
                                    : null;
                              },
                            ),
                          ),
                          SizedBox(
                            height: 30.0,
                            child: IconButton(
                              padding: EdgeInsets.zero,
                              iconSize: 30.0,
                              icon: Icon(Icons.remove_circle_outline),
                              color: Colors.red,
                              onPressed: () {
                                setState(() {
                                  _unitname3.clear();
                                  _displayNewUnitCode3 = false;
                                });
                              },
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
            Visibility(
              visible: _displayNewUnitCode4,
              child: Container(
                padding: const EdgeInsets.only(
                  left: 8.0,
                  right: 8.0,
                  top: 15.0,
                ),
                child: Row(
                  children: [
                    (Util.isLandscape(context))
                        ? Expanded(
                            flex: 1,
                            child: Text(global.language('unit_name_4')),
                          )
                        : Container(),
                    Expanded(
                      flex: 2,
                      child: TextFieldInput(
                        controller: _unitname4,
                        keyboardType: TextInputType.text,
                        labelText: global.language('unit_name_4'),
                      ),
                    ),
                    Expanded(
                      flex: 1,
                      child: Row(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          SizedBox(
                            height: 30.0,
                            child: IconButton(
                              padding: EdgeInsets.zero,
                              iconSize: 30.0,
                              icon: Icon(Icons.add_circle_outline),
                              color: _displayNewUnitCode5 == true
                                  ? Colors.grey
                                  : Colors.blue,
                              onPressed: () {
                                _displayNewUnitCode5 != true
                                    ? setState(() {
                                        _displayNewUnitCode5 = true;
                                      })
                                    : null;
                              },
                            ),
                          ),
                          SizedBox(
                            height: 30.0,
                            child: IconButton(
                              padding: EdgeInsets.zero,
                              iconSize: 30.0,
                              icon: Icon(Icons.remove_circle_outline),
                              color: Colors.red,
                              onPressed: () {
                                setState(() {
                                  _unitname4.clear();
                                  _displayNewUnitCode4 = false;
                                });
                              },
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
            ),
            Visibility(
              visible: _displayNewUnitCode5,
              child: Container(
                padding: const EdgeInsets.only(
                  left: 8.0,
                  right: 8.0,
                  top: 15.0,
                ),
                child: Row(
                  children: [
                    (Util.isLandscape(context))
                        ? Expanded(
                            flex: 1,
                            child: Text(global.language('unit_name_5')),
                          )
                        : Container(),
                    Expanded(
                      flex: 2,
                      child: TextFieldInput(
                        controller: _unitname5,
                        keyboardType: TextInputType.text,
                        labelText: global.language('unit_name_5'),
                      ),
                    ),
                    Expanded(
                      flex: 1,
                      child: SizedBox(
                        height: 30.0,
                        child: IconButton(
                          padding: EdgeInsets.zero,
                          iconSize: 30.0,
                          icon: Icon(Icons.remove_circle_outline),
                          color: Colors.red,
                          onPressed: () {
                            setState(() {
                              _unitname5.clear();
                              _displayNewUnitCode5 = false;
                            });
                          },
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
            Container(
              padding: EdgeInsets.only(left: 8.0, right: 8.0, top: 15.0),
              child: Row(
                children: [
                  (Util.isLandscape(context))
                      ? Expanded(flex: 1, child: Text(global.language('price')))
                      : Container(),
                  Expanded(
                    flex: 3,
                    child: TextFieldInput(
                      controller: _price,
                      keyboardType: TextInputType.number,
                      labelText: global.language('price'),
                    ),
                  ),
                ],
              ),
            ),

            //แท็กสินค้า
            /*
            Container(
              padding: const EdgeInsets.all(8.0),
              child: Row(
                children: [
                  (Util.isLandscape(context))
                      ? Expanded(
                          flex: 1,
                          child: Text(global.language('tags')),
                        )
                      : Container(),
                  Expanded(
                    flex: 3,
                    child: TagEditor(
                      length: _values.length,
                      controller: _textEditingController,
                      delimiters: [',', ' '],
                      hasAddButton: true,
                      resetTextOnSubmitted: true,
                      // This is set to grey just to illustrate the `textStyle` prop
                      textStyle: const TextStyle(color: Colors.black),
                      onSubmitted: (outstandingValue) {
                        setState(() {
                          _values.add(outstandingValue);
                        });
                      },
                      inputDecoration: const InputDecoration(
                        border: InputBorder.none,
                        hintText: 'แท็กสินค้า ...',
                      ),
                      onTagChanged: (newValue) {
                        setState(() {
                          _values.add(newValue);
                        });
                      },
                      tagBuilder: (context, index) => _Chip(
                        index: index,
                        label: _values[index],
                        onDeleted: _onDelete,
                      ),
                      // InputFormatters example, this disallow \ and /
                      inputFormatters: [
                        FilteringTextInputFormatter.deny(RegExp(r'[/\\]'))
                      ],
                    ),
                  ),
                ],
              ),
            ),
            Divider(),
            Row(
              children: [
                ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    primary: Color(0xFFF88975),
                  ),
                  onPressed: () {
                    var _text = 'สินค้าแนะนำ';
                    _textEditingController.text = _text;
                    _textEditingController.value =
                        _textEditingController.value.copyWith(
                      text: _text,
                      selection: TextSelection(
                        baseOffset: _text.length,
                        extentOffset: _text.length,
                      ),
                    );
                  },
                  child: const Text('# สินค้าแนะนำ'),
                ),
                SizedBox(width: 5.0),
                ElevatedButton(
                  onPressed: () {
                    var _text = global.language('best_selling_products');
                    _textEditingController.text = _text;
                    _textEditingController.value =
                        _textEditingController.value.copyWith(
                      text: _text,
                      selection: TextSelection(
                        baseOffset: _text.length,
                        extentOffset: _text.length,
                      ),
                    );
                  },
                  child: const Text('# สินค้าขายดี'),
                ),
              ],
            ),
            */
          ],
        ),
      ),
    );
  }

  Widget selectOption() {
    return Card(
      child: SizedBox(
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.only(
                left: 30.0,
                right: 30.0,
                top: 30.0,
              ),
              child: Align(
                alignment: Alignment.topLeft,
                child: ElevatedButton(
                  style: TextButton.styleFrom(
                    textStyle: const TextStyle(fontSize: 20),
                  ),
                  onPressed: () {
                    showDialog(
                      context: context,
                      builder: (BuildContext context) {
                        return AlertDialog(
                          title: Text(global.language('select_addon_option')),
                          content: setupAlertDialoadOption(),
                        );
                      },
                    );
                  },
                  child: Text(global.language('add_addon_option')),
                ),
              ),
            ),
            Padding(
              padding: EdgeInsets.all(10.0),
              child: ListView.builder(
                shrinkWrap: true,
                physics: const NeverScrollableScrollPhysics(),
                itemCount: _optionDetail.length,
                itemBuilder: (context, index) {
                  return Card(
                    child: ListTile(
                      title: Align(
                        alignment: Alignment.center,
                        child: Text(_optionDetail[index].name1),
                      ),
                      subtitle: Column(
                        children: [
                          Text(
                            "${global.language('sub_options')}${_optionDetail[index].choices.length}",
                          ),
                          Padding(
                            padding: EdgeInsets.all(8.0),
                            child: ElevatedButton.icon(
                              icon: const Icon(
                                Icons.delete,
                                color: Colors.white,
                                size: 24.0,
                              ),
                              label: Text(
                                global.language('delete_addon_option'),
                              ),
                              style: ElevatedButton.styleFrom(
                                backgroundColor: Colors.red,
                                padding: const EdgeInsets.symmetric(
                                  vertical: 5,
                                  horizontal: 15,
                                ),
                              ),
                              onPressed: () {
                                _optionDetail.removeAt(index);
                                setState(() {
                                  _optionDetail = _optionDetail;
                                });
                              },
                            ),
                          ),
                        ],
                      ),
                    ),
                  );
                },
              ),
            ),
          ],
        ),
      ),
    );
  }

  bool checkEnable(OptionModel opt) {
    for (var detail in _optionDetail) {
      if (opt.name1 == detail.name1) {
        return false;
      }
    }
    return true;
  }

  Widget setupAlertDialoadOption() {
    return SizedBox(
      width: double.maxFinite, // Change as per your requirement
      child: BlocBuilder<OptionBloc, OptionState>(
        builder: (context, state) {
          return (state is OptionLoadSelectInProgress)
              ? Container(
                  color: Colors.transparent,
                  alignment: Alignment.center,
                  child: const CircularProgressIndicator(),
                )
              : (state is OptionLoadSelectSuccess)
              ? ListView.builder(
                  itemCount: _option.length,
                  itemBuilder: (context, index) {
                    return Card(
                      elevation: 1,
                      child: ListTile(
                        enabled: checkEnable(state.option[index]),
                        onTap: () {
                          OptionModel result = OptionModel(
                            guidfixed: state.option[index].guidfixed,
                            code: state.option[index].code,
                            name1: state.option[index].name1,
                            name2: state.option[index].name2,
                            name3: state.option[index].name3,
                            name4: state.option[index].name4,
                            name5: state.option[index].name5,
                            choicetype: state.option[index].choicetype,
                            maxselect: state.option[index].maxselect,
                            isrequired: state.option[index].isrequired,
                            choices: state.option[index].choices,
                          );

                          setState(() {
                            _optionDetail.add(result);
                            // // print(_result.toJson());
                          });

                          Navigator.pop(context, result);
                        },
                        title: Text(_option[index].name1.toString()),
                        subtitle: Text(
                          '${global.language('sub_options')}${_option[index].choices.length}',
                          style: const TextStyle(color: Colors.black),
                        ),
                        trailing: const Icon(Icons.keyboard_arrow_right_sharp),
                      ),
                    );
                  },
                )
              : (state is OptionLoadSelectFailure)
              ? Text(
                  global.language('no_data_found'),
                  style: const TextStyle(fontSize: 24),
                )
              : Container();
        },
      ),
    );
  }

  Widget selectImage() {
    return Card(
      child: SizedBox(
        child: Column(
          children: [
            Container(
              padding: const EdgeInsets.only(
                left: 30.0,
                right: 30.0,
                top: 30.0,
              ),
              child: Align(
                alignment: Alignment.topLeft,
                child: Column(
                  children: [Text(global.language('select_pircture'))],
                ),
              ),
            ),
            Padding(
              padding: const EdgeInsets.all(30.0),
              child: _imageProduct(),
            ),
          ],
        ),
      ),
    );
  }

  Widget _imageProduct() {
    return Center(
      child: Column(
        children: [
          (_images.isNotEmpty)
              ? ListView.builder(
                  shrinkWrap: true,
                  physics: const NeverScrollableScrollPhysics(),
                  itemCount: _images.length,
                  itemBuilder: (context, index) {
                    return ListTile(
                      title: Align(
                        alignment: Alignment.center,
                        child: Stack(
                          children: [
                            Padding(
                              padding: const EdgeInsets.only(
                                top: 10.0,
                                right: 10.0,
                              ),
                              child: Image.network(_images[index].uri),
                            ),
                            Positioned(
                              right: 0.0,
                              child: GestureDetector(
                                onTap: () {
                                  _images.removeAt(index);
                                  setState(() {
                                    _images = _images;
                                  });
                                },
                                child: Align(
                                  alignment: Alignment.topRight,
                                  child: CircleAvatar(
                                    radius: 18.0,
                                    backgroundColor: Colors.red,
                                    child: Icon(
                                      Icons.close,
                                      color: Colors.white,
                                    ),
                                  ),
                                ),
                              ),
                            ),
                          ],
                        ),
                      ),
                    );
                  },
                )
              : IconButton(
                  padding: const EdgeInsets.all(2.0),
                  icon: Icon(
                    Icons.add_photo_alternate_outlined,
                    color: Colors.teal,
                  ),
                  iconSize: MediaQuery.of(context).size.height * 0.15,
                  onPressed: () {
                    showModalBottomSheet(
                      context: context,
                      builder: ((builder) => _bottomSheet()),
                    );
                  },
                ),
          SizedBox(height: MediaQuery.of(context).size.height * 0.05),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Expanded(
                child: _addPhoto(
                  title: global.language('take_photo'),
                  type: true,
                  icon: Icons.add_a_photo,
                ),
              ),
              Expanded(
                child: _addPhoto(
                  title: global.language('select_image'),
                  type: false,
                  icon: Icons.photo_library,
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _addPhoto({
    required bool type,
    required IconData icon,
    required String title,
  }) {
    return TextButton(
      onPressed: () async {
        final ImageUpload? result = await Navigator.push(
          context,
          MaterialPageRoute(builder: (context) => ProductEditImageScreen(type)),
        );
        _reloadImage = true;
        setState(() {
          if (result != null) {
            _images.add(result);
          }
        });
      },
      child: Column(
        children: <Widget>[
          Icon(icon, size: 50.0, color: Colors.blue[300]),
          Text(title),
        ],
      ),
    );
  }

  Widget _bottomSheet() {
    return Container(
      height: 100.0,
      width: MediaQuery.of(context).size.width,
      margin: EdgeInsets.symmetric(horizontal: 20, vertical: 20),
      child: Column(
        children: [
          Text(
            global.language("select_image"),
            style: const TextStyle(fontSize: 20.0),
          ),
          SizedBox(height: 20),
          Row(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              TextButton.icon(
                onPressed: () async {
                  Navigator.pop(context);
                  final ImageUpload? result = await Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => const ProductEditImageScreen(true),
                    ),
                  );
                  _reloadImage = true;
                  setState(() {
                    if (result != null) {
                      _images.add(result);
                    }
                  });
                },
                icon: Icon(Icons.add_a_photo),
                label: Text(global.language("take_photo")),
              ),
              TextButton.icon(
                onPressed: () async {
                  Navigator.pop(context);
                  final ImageUpload? result = await Navigator.push(
                    context,
                    MaterialPageRoute(
                      builder: (context) => const ProductEditImageScreen(false),
                    ),
                  );
                  _reloadImage = true;
                  setState(() {
                    if (result != null) {
                      _images.add(result);
                    }
                  });
                },
                icon: Icon(Icons.photo_library),
                label: Text(global.language("photo_gallery")),
              ),
            ],
          ),
        ],
      ),
    );
  }
}

class CChip extends StatelessWidget {
  const CChip({
    super.key,
    required this.label,
    required this.onDeleted,
    required this.index,
  });

  final String label;
  final ValueChanged<int> onDeleted;
  final int index;

  @override
  Widget build(BuildContext context) {
    return Chip(
      labelPadding: const EdgeInsets.only(left: 8.0),
      label: Text(label, style: TextStyle(color: Colors.black)),
      elevation: 1.0,
      // backgroundColor: Color(0xFFF88975),
      deleteIcon: const Icon(Icons.cancel, size: 18, color: Colors.black),
      onDeleted: () {
        onDeleted(index);
      },
    );
  }
}
