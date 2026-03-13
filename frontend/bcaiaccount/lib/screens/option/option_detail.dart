import 'package:smlaicloud/components/appbar.dart';
import 'package:smlaicloud/components/textfield_input.dart';
import 'package:smlaicloud/model/choice_model.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/widgets/edit_font_size_control.dart';

import 'package:flutter/material.dart';

class OptionDetailScreen extends StatefulWidget {
  final ChoiceModel choices;

  const OptionDetailScreen({super.key, required this.choices});

  @override
  _OptionDetailScreenState createState() => _OptionDetailScreenState();
}

class _OptionDetailScreenState extends State<OptionDetailScreen> with global.ThemeRefreshMixin {
  bool createMode = true;
  // from
  final _formKey = GlobalKey<FormState>();

  final _barcode = TextEditingController();
  final _name1 = TextEditingController();
  final _name2 = TextEditingController();
  final _name3 = TextEditingController();
  final _name4 = TextEditingController();
  final _name5 = TextEditingController();
  final _priceController = TextEditingController();
  final _qty = TextEditingController();
  final _qtymax = TextEditingController();
  final _suggestcode = TextEditingController();
  final _itemunit = TextEditingController();
  bool _selected = false;
  bool _default = true;

  bool _displayNewTextFieldName2 = false;
  bool _displayNewTextFieldName3 = false;
  bool _displayNewTextFieldName4 = false;
  bool _displayNewTextFieldName5 = false;

  @override
  void initState() {
    _barcode.text = widget.choices.barcode;
    _name1.text = widget.choices.name1;
    if (widget.choices.name2 != '') {
      _name2.text = widget.choices.name2;
      _displayNewTextFieldName2 = true;
    }
    if (widget.choices.name3 != '') {
      _name3.text = widget.choices.name3;
      _displayNewTextFieldName3 = true;
    }
    if (widget.choices.name4 != '') {
      _name4.text = widget.choices.name4;
      _displayNewTextFieldName4 = true;
    }
    if (widget.choices.name5 != '') {
      _name5.text = widget.choices.name5;
      _displayNewTextFieldName5 = true;
    }
    if (widget.choices.price.toString() != '0.0') {
      _priceController.text = widget.choices.price.toString();
    }
    if (widget.choices.qty.toString() != '0.0') {
      _qty.text = widget.choices.qty.toString();
    }
    if (widget.choices.qtymax.toString() != '0.0') {
      _qtymax.text = widget.choices.qtymax.toString();
    }

    _suggestcode.text = widget.choices.suggestcode;
    _itemunit.text = widget.choices.itemunit;
    if (widget.choices.selected == true) {
      setState(() {
        _selected = true;
      });
    }
    if (widget.choices.isdefault == true) {
      setState(() {
        _default = true;
      });
    }

    createMode = false;
    super.initState();
  }

  @override
  void dispose() {
    super.dispose();
  }

  Future _saveData() async {
    double price = double.parse(_priceController.text.toString());

    if (_qty.text == '') {
      _qty.text = '0';
    }
    if (_qtymax.text == '') {
      _qtymax.text = '0';
    }

    ChoiceModel result = ChoiceModel(
      barcode: _barcode.text,
      name1: _name1.text,
      name2: _name2.text,
      name3: _name3.text,
      name4: _name4.text,
      name5: _name5.text,
      price: price,
      qty: double.parse(_qty.text),
      qtymax: double.parse(_qtymax.text),
      suggestcode: _suggestcode.text,
      itemunit: _itemunit.text,
      selected: _selected,
      isdefault: _default,
    );

    // // print(_result.toJson());

    Navigator.pop(context, result);
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      backgroundColor: global.theme.infoHighlightColor,
      appBar: BaseAppBar(
        title: (createMode)
            ? Text(global.language("option_detail.add_sub_option"))
            : Text(global.language("option_detail.edit_sub_option")),
        appBar: AppBar(),
        widgets: <Widget>[
          EditFontSizeControl(onChanged: () => setState(() {})),
        ],
      ),
      body: Builder(
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
              child: Container(
        margin: EdgeInsets.all(10.0),
        child: SingleChildScrollView(
          scrollDirection: Axis.vertical,
          child: Form(
            key: _formKey,
            child: Padding(
              padding: EdgeInsets.only(top: 8.0),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                crossAxisAlignment: CrossAxisAlignment.stretch,
                children: <Widget>[
                  TextFormField(
                    controller: _barcode,
                    decoration: InputDecoration(
                      border: OutlineInputBorder(
                        borderRadius: BorderRadius.circular(10.0),
                      ),
                      labelText: global.language('barcode'),
                      isDense: true,
                      contentPadding: const EdgeInsets.all(10),
                    ),
                    validator: (value) {
                      if (value == null || value.isEmpty) {
                        return global.language(
                          'option_detail.please_specify_barcode',
                        );
                      }
                      return null;
                    },
                  ),
                  Container(
                    padding: EdgeInsets.only(top: 8.0),
                    child: Row(
                      children: [
                        Expanded(
                          flex: 3,
                          child: TextFormField(
                            controller: _name1,
                            decoration: InputDecoration(
                              border: OutlineInputBorder(
                                borderRadius: BorderRadius.circular(10.0),
                              ),
                              labelText: global.language(
                                'option_detail.option_group_name_1',
                              ),
                              isDense: true,
                              contentPadding: const EdgeInsets.all(10),
                            ),
                            validator: (value) {
                              if (value == null || value.isEmpty) {
                                return global.language(
                                  'option_detail.please_specify_option_group_name',
                                );
                              }
                              return null;
                            },
                          ),
                        ),
                        Expanded(
                          flex: 1,
                          child: SizedBox(
                            height: 30.0,
                            width: 30.0,
                            child: IconButton(
                              padding: EdgeInsets.zero,
                              iconSize: 30.0,
                              icon: Icon(Icons.add_circle_outline),
                              color: _displayNewTextFieldName2 == true
                                  ? global.theme.iconSecondaryColor
                                  : global.theme.infoHighlightTextColor,
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
                      padding: EdgeInsets.only(top: 8.0),
                      child: Row(
                        children: [
                          Expanded(
                            flex: 3,
                            child: TextFieldInput(
                              controller: _name2,
                              keyboardType: TextInputType.text,
                              labelText: global.language(
                                'option_detail.option_group_name_2',
                              ),
                            ),
                          ),
                          Expanded(
                            flex: 1,
                            child: Row(
                              mainAxisAlignment: MainAxisAlignment.spaceAround,
                              children: [
                                SizedBox(
                                  height: 30.0,
                                  width: 30.0,
                                  child: IconButton(
                                    padding: EdgeInsets.zero,
                                    iconSize: 30.0,
                                    icon: Icon(Icons.add_circle_outline),
                                    color: _displayNewTextFieldName3 == true
                                        ? global.theme.iconSecondaryColor
                                        : global.theme.infoHighlightTextColor,
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
                                  width: 30.0,
                                  child: IconButton(
                                    padding: EdgeInsets.zero,
                                    iconSize: 30.0,
                                    icon: Icon(
                                      Icons.remove_circle_outline,
                                    ),
                                    color: global.theme.negativeHighlightTextColor,
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
                      padding: EdgeInsets.only(top: 8.0),
                      child: Row(
                        children: [
                          Expanded(
                            flex: 3,
                            child: TextFieldInput(
                              controller: _name3,
                              keyboardType: TextInputType.text,
                              labelText: global.language(
                                'option_detail.option_group_name_3',
                              ),
                            ),
                          ),
                          Expanded(
                            flex: 1,
                            child: Row(
                              mainAxisAlignment: MainAxisAlignment.spaceAround,
                              children: [
                                SizedBox(
                                  height: 30.0,
                                  width: 30.0,
                                  child: IconButton(
                                    padding: EdgeInsets.zero,
                                    iconSize: 30.0,
                                    icon: Icon(Icons.add_circle_outline),
                                    color: _displayNewTextFieldName4 == true
                                        ? global.theme.iconSecondaryColor
                                        : global.theme.infoHighlightTextColor,
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
                                  width: 30.0,
                                  child: IconButton(
                                    padding: EdgeInsets.zero,
                                    iconSize: 30.0,
                                    icon: Icon(
                                      Icons.remove_circle_outline,
                                    ),
                                    color: global.theme.negativeHighlightTextColor,
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
                      padding: EdgeInsets.only(top: 8.0),
                      child: Row(
                        children: [
                          Expanded(
                            flex: 3,
                            child: TextFieldInput(
                              controller: _name4,
                              keyboardType: TextInputType.text,
                              labelText: global.language(
                                'option_detail.option_group_name_4',
                              ),
                            ),
                          ),
                          Expanded(
                            flex: 1,
                            child: Row(
                              mainAxisAlignment: MainAxisAlignment.spaceAround,
                              children: [
                                SizedBox(
                                  height: 30.0,
                                  width: 30.0,
                                  child: IconButton(
                                    padding: EdgeInsets.zero,
                                    iconSize: 30.0,
                                    icon: Icon(Icons.add_circle_outline),
                                    color: _displayNewTextFieldName5 == true
                                        ? global.theme.iconSecondaryColor
                                        : global.theme.infoHighlightTextColor,
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
                                  width: 30.0,
                                  child: IconButton(
                                    padding: EdgeInsets.zero,
                                    iconSize: 30.0,
                                    icon: Icon(
                                      Icons.remove_circle_outline,
                                    ),
                                    color: global.theme.negativeHighlightTextColor,
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
                      padding: EdgeInsets.only(top: 8.0),
                      child: Row(
                        children: [
                          Expanded(
                            flex: 3,
                            child: TextFieldInput(
                              controller: _name5,
                              keyboardType: TextInputType.text,
                              labelText: global.language(
                                'option_detail.option_group_name_5',
                              ),
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
                                color: global.theme.negativeHighlightTextColor,
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
                  Container(
                    padding: EdgeInsets.only(top: 8.0),
                    child: TextFormField(
                      controller: _priceController,
                      keyboardType: TextInputType.number,
                      decoration: InputDecoration(
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(10.0),
                        ),
                        labelText: global.language('option_detail.price'),
                        isDense: true,
                        contentPadding: const EdgeInsets.all(10),
                      ),
                      validator: (value) {
                        if (value == null || value.isEmpty) {
                          return global.language(
                            'option_detail.please_specify_price',
                          );
                        }
                        return null;
                      },
                    ),
                  ),
                  Container(
                    padding: EdgeInsets.only(top: 8.0),
                    child: TextFormField(
                      controller: _qty,
                      keyboardType: TextInputType.number,
                      decoration: InputDecoration(
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(10.0),
                        ),
                        labelText: global.language('option_detail.quantity'),
                        isDense: true,
                        contentPadding: const EdgeInsets.all(10),
                      ),
                    ),
                  ),
                  Container(
                    padding: EdgeInsets.only(top: 8.0),
                    child: TextFormField(
                      controller: _qtymax,
                      keyboardType: TextInputType.number,
                      decoration: InputDecoration(
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(10.0),
                        ),
                        labelText: global.language(
                          'option_detail.maximum_quantity',
                        ),
                        isDense: true,
                        contentPadding: const EdgeInsets.all(10),
                      ),
                    ),
                  ),
                  Container(
                    padding: EdgeInsets.only(top: 8.0),
                    child: TextFormField(
                      controller: _suggestcode,
                      keyboardType: TextInputType.text,
                      decoration: InputDecoration(
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(10.0),
                        ),
                        labelText: global.language(
                          'option_detail.suggest_code',
                        ),
                        isDense: true,
                        contentPadding: const EdgeInsets.all(10),
                      ),
                    ),
                  ),
                  Container(
                    padding: EdgeInsets.only(top: 8.0),
                    child: TextFormField(
                      controller: _itemunit,
                      keyboardType: TextInputType.text,
                      decoration: InputDecoration(
                        border: OutlineInputBorder(
                          borderRadius: BorderRadius.circular(10.0),
                        ),
                        labelText: global.language('option_detail.unit'),
                        isDense: true,
                        contentPadding: const EdgeInsets.all(10),
                      ),
                    ),
                  ),
                  Container(
                    padding: EdgeInsets.only(top: 8.0),
                    child: Row(
                      children: [
                        Checkbox(
                          shape: const RoundedRectangleBorder(
                            borderRadius: BorderRadius.all(
                              Radius.circular(5.0),
                            ),
                          ),
                          checkColor: global.theme.onPrimaryColor,
                          value: _selected,
                          onChanged: (bool? value) {
                            setState(() {
                              _selected = value!;
                            });
                          },
                        ),
                        Text(
                          global.language('option_detail.select_immediately'),
                        ),
                      ],
                    ),
                  ),
                  Row(
                    children: [
                      Checkbox(
                        shape: const RoundedRectangleBorder(
                          borderRadius: BorderRadius.all(Radius.circular(5.0)),
                        ),
                        checkColor: global.theme.onPrimaryColor,
                        value: _default,
                        onChanged: (bool? value) {
                          setState(() {
                            _default = value!;
                          });
                        },
                      ),
                      Text(global.language('option_detail.default_value')), //
                    ],
                  ),
                ],
              ),
            ),
          ),
        ),
            ),
          ),
        );
        },
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () {
          if (_formKey.currentState!.validate()) {
            _saveData();
          }
        },
        backgroundColor: global.theme.positiveHighlightTextColor,
        child: Icon(Icons.save),
      ),
    );
  }
}
