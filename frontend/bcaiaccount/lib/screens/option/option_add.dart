import 'package:smlaicloud/screens/Option/option_detail.dart';
import 'package:smlaicloud/screens/Option/option_list.dart';
import 'package:smlaicloud/bloc/option/option_bloc.dart';
import 'package:smlaicloud/model/choice_model.dart';
import 'package:smlaicloud/model/option_model.dart';
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:group_radio_button/group_radio_button.dart';
import 'package:loader_overlay/loader_overlay.dart';
import 'package:smlaicloud/components/appbar.dart';
import 'package:smlaicloud/components/background_main.dart';
import 'package:smlaicloud/components/textfield_input.dart';
import 'package:smlaicloud/utils/util.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/widgets/edit_font_size_control.dart';

class OptionAdd extends StatefulWidget {
  final String? guidfixed;
  const OptionAdd({super.key, this.guidfixed = ""});

  @override
  State<OptionAdd> createState() => _OptionAddState();
}

class _OptionAddState extends State<OptionAdd> with global.ThemeRefreshMixin {
  bool createMode = true;

  // from
  final _formKey = GlobalKey<FormState>();
  String? _uuid;
  final _name1 = TextEditingController();
  final _name2 = TextEditingController();
  final _name3 = TextEditingController();
  final _name4 = TextEditingController();
  final _name5 = TextEditingController();
  final _maxselected = TextEditingController();

  bool _displayNewTextFieldName2 = false;
  bool _displayNewTextFieldName3 = false;
  bool _displayNewTextFieldName4 = false;
  bool _displayNewTextFieldName5 = false;
  bool _isRequired = true;

  bool _ismaxselected = false;

  List<ChoiceModel> _optionChoices = <ChoiceModel>[];

  int _choiceType = 0;

  @override
  void initState() {
    super.initState();
    widget.guidfixed.toString() != "" ? createMode = false : createMode = true;

    if (!createMode) {
      context.read<OptionBloc>().add(
        ListOptionLoadById(id: widget.guidfixed.toString()),
      );
    }
  }

  @override
  void dispose() {
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    Size size = MediaQuery.of(context).size;

    return BlocListener<OptionBloc, OptionState>(
      listener: (context, state) {
        if (state is OptionLoadByIdInProgress) {
          context.loaderOverlay.show(
            widgetBuilder: (progress) {
              return const ReconnectingOverlay();
            },
          );
        } else if (state is OptionLoadByIdLoadSuccess) {
          context.loaderOverlay.hide();
          setState(() {
            _uuid = state.option.code;
            _name1.text = state.option.name1;
            if (state.option.name2 != '') {
              _name2.text = state.option.name2;
              _displayNewTextFieldName2 = true;
            }
            if (state.option.name3 != '') {
              _name3.text = state.option.name3;
              _displayNewTextFieldName3 = true;
            }
            if (state.option.name4 != '') {
              _name4.text = state.option.name4;
              _displayNewTextFieldName4 = true;
            }
            if (state.option.name5 != '') {
              _name5.text = state.option.name5;
              _displayNewTextFieldName5 = true;
            }

            if (state.option.isrequired == true) {
              _isRequired = true;
            } else {
              _isRequired = false;
            }

            if (state.option.choicetype == 0) {
              _choiceType = 0;
            } else {
              _choiceType = 1;
              _ismaxselected = true;
              _maxselected.text = state.option.maxselect.toString();
            }

            if (state.option.choices.isNotEmpty) {
              _optionChoices = state.option.choices;
            }
          });
        }

        if (state is OptionSaveInProgress) {
          context.loaderOverlay.show(
            widgetBuilder: (progress) {
              return const ReconnectingOverlay();
            },
          );
        } else if (state is OptionSaveSuccess) {
          context.loaderOverlay.hide();
          showDeleteDialog('save');
        }

        if (state is OptionUpdateInProgress) {
          context.loaderOverlay.show(
            widgetBuilder: (progress) {
              return const ReconnectingOverlay();
            },
          );
        } else if (state is OptionUpdateSuccess) {
          context.loaderOverlay.hide();
          showDeleteDialog('update');
        }

        if (state is OptionDeleteSuccess) {
          Navigator.of(context).pushReplacement(
            MaterialPageRoute(builder: (_) => const OptionScreen()),
          );
        }
      },
      child: Scaffold(
        appBar: BaseAppBar(
          title: (createMode)
              ? Text(global.language('add_supplement_option'))
              : Text(global.language('edit_supplement_option')),
          leading: IconButton(
            color: global.theme.textColor,
            icon: Icon(Icons.arrow_back),
            onPressed: () {
              Navigator.of(context).pushReplacement(
                MaterialPageRoute(builder: (_) => const OptionScreen()),
              );
            },
          ),
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
                child: LoaderOverlay(
          overlayColor: global.theme.textColor,
          child: BackgroundMain(
            child: Container(
              margin: EdgeInsets.all(10.0),
              child: SingleChildScrollView(
                scrollDirection: Axis.vertical,
                child: Form(
                  key: _formKey,
                  child: Padding(
                    padding: EdgeInsets.only(top: 15.0),
                    child: Column(
                      mainAxisSize: MainAxisSize.min,
                      crossAxisAlignment: CrossAxisAlignment.stretch,
                      children: <Widget>[
                        Row(
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
                                    'option_group_name_1',
                                  ),
                                  isDense: true,
                                  contentPadding: const EdgeInsets.all(10),
                                ),
                                validator: (value) {
                                  if (value == null || value.isEmpty) {
                                    return global.language(
                                      'please_specify_option_group_name',
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

                        Visibility(
                          visible: _displayNewTextFieldName2,
                          child: Container(
                            padding: EdgeInsets.only(top: 15.0),
                            child: Row(
                              children: [
                                Expanded(
                                  flex: 3,
                                  child: TextFieldInput(
                                    controller: _name2,
                                    keyboardType: TextInputType.text,
                                    labelText: global.language(
                                      'option_group_name_2',
                                    ),
                                  ),
                                ),
                                Expanded(
                                  flex: 1,
                                  child: Row(
                                    mainAxisAlignment:
                                        MainAxisAlignment.spaceAround,
                                    children: [
                                      SizedBox(
                                        height: 30.0,
                                        width: 30.0,
                                        child: IconButton(
                                          padding: EdgeInsets.zero,
                                          iconSize: 30.0,
                                          icon: Icon(
                                            Icons.add_circle_outline,
                                          ),
                                          color:
                                              _displayNewTextFieldName3 == true
                                              ? global.theme.iconSecondaryColor
                                              : global.theme.infoHighlightTextColor,
                                          onPressed: () {
                                            _displayNewTextFieldName3 != true
                                                ? setState(() {
                                                    _displayNewTextFieldName3 =
                                                        true;
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
                            padding: EdgeInsets.only(top: 15.0),
                            child: Row(
                              children: [
                                Expanded(
                                  flex: 3,
                                  child: TextFieldInput(
                                    controller: _name3,
                                    keyboardType: TextInputType.text,
                                    labelText: global.language(
                                      'option_group_name_3',
                                    ),
                                  ),
                                ),
                                Expanded(
                                  flex: 1,
                                  child: Row(
                                    mainAxisAlignment:
                                        MainAxisAlignment.spaceAround,
                                    children: [
                                      SizedBox(
                                        height: 30.0,
                                        width: 30.0,
                                        child: IconButton(
                                          padding: EdgeInsets.zero,
                                          iconSize: 30.0,
                                          icon: Icon(
                                            Icons.add_circle_outline,
                                          ),
                                          color:
                                              _displayNewTextFieldName4 == true
                                              ? global.theme.iconSecondaryColor
                                              : global.theme.infoHighlightTextColor,
                                          onPressed: () {
                                            _displayNewTextFieldName4 != true
                                                ? setState(() {
                                                    _displayNewTextFieldName4 =
                                                        true;
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
                            padding: EdgeInsets.only(top: 15.0),
                            child: Row(
                              children: [
                                Expanded(
                                  flex: 3,
                                  child: TextFieldInput(
                                    controller: _name4,
                                    keyboardType: TextInputType.text,
                                    labelText: global.language(
                                      'option_group_name_4',
                                    ),
                                  ),
                                ),
                                Expanded(
                                  flex: 1,
                                  child: Row(
                                    mainAxisAlignment:
                                        MainAxisAlignment.spaceAround,
                                    children: [
                                      SizedBox(
                                        height: 30.0,
                                        width: 30.0,
                                        child: IconButton(
                                          padding: EdgeInsets.zero,
                                          iconSize: 30.0,
                                          icon: Icon(
                                            Icons.add_circle_outline,
                                          ),
                                          color:
                                              _displayNewTextFieldName5 == true
                                              ? global.theme.iconSecondaryColor
                                              : global.theme.infoHighlightTextColor,
                                          onPressed: () {
                                            _displayNewTextFieldName5 != true
                                                ? setState(() {
                                                    _displayNewTextFieldName5 =
                                                        true;
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
                            padding: EdgeInsets.only(top: 15.0),
                            child: Row(
                              children: [
                                Expanded(
                                  flex: 3,
                                  child: TextFieldInput(
                                    controller: _name5,
                                    keyboardType: TextInputType.text,
                                    labelText: global.language(
                                      'option_group_name_5',
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
                                      icon: Icon(
                                        Icons.remove_circle_outline,
                                      ),
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

                        SizedBox(height: size.height * 0.03),
                        // Widget
                        TitleLabel(
                          title: global.language('option_choices'),
                          subtitle: global.language('option_example_text'),
                        ),
                        SizedBox(height: size.height * 0.03),
                        Align(
                          alignment: Alignment.centerLeft,
                          child: TextButton(
                            style: TextButton.styleFrom(
                              textStyle: TextStyle(fontSize: 20),
                            ),
                            onPressed: () async {
                              var result = await Navigator.push(
                                context,
                                MaterialPageRoute(
                                  builder: (context) => OptionDetailScreen(
                                    choices: ChoiceModel(
                                      isdefault: true,
                                      selected: false,
                                    ),
                                  ),
                                ),
                              );

                              if (result != null) {
                                _optionChoices.add(result);
                              }
                            },
                            child: Text(global.language('add_sub_option')),
                          ),
                        ),

                        ListView.builder(
                          shrinkWrap: true,
                          physics: const NeverScrollableScrollPhysics(),
                          itemCount: _optionChoices.length,
                          itemBuilder: (context, index) {
                            return Card(
                              child: ListTile(
                                title: Align(
                                  alignment: Alignment.center,
                                  child: Text(_optionChoices[index].barcode),
                                ),
                                subtitle: Column(
                                  children: [
                                    Text(
                                      "${global.language('name_1_prefix')}${_optionChoices[index].name1} ",
                                    ),
                                    // Text("ชื่อ 2 :  ${_optionChoices[index].name2} "),
                                    // Text("ชื่อ 3 :  ${_optionChoices[index].name3} "),
                                    // Text("ชื่อ 4 :  ${_optionChoices[index].name4} "),
                                    // Text("ชื่อ 5 :  ${_optionChoices[index].name5} "),
                                    Text(
                                      "${global.language('price_prefix')}${_optionChoices[index].price}${global.language('baht_suffix')}",
                                    ),
                                    // Text("จำนวน :  ${_optionChoices[index].qty} "),
                                    // Text(
                                    //     "จำนวนสูงสุด :  ${_optionChoices[index].qtymax} "),
                                    // Text(
                                    //     "รหัสแนะนำ :  ${_optionChoices[index].suggestcode} "),
                                    // Text(
                                    //     "หน่วยนับ :  ${_optionChoices[index].itemunit} "),
                                    // Text(
                                    //     "เลือกให้เลย :  ${_optionChoices[index].selected} "),
                                    // Text(
                                    //     "ค่าเริ่มต้น :  ${_optionChoices[index].isdefault} "),
                                    Padding(
                                      padding: EdgeInsets.all(8.0),
                                      child: ElevatedButton.icon(
                                        icon: Icon(
                                          Icons.delete,
                                          color: global.theme.onPrimaryColor,
                                          size: 24.0,
                                        ),
                                        label: Text(
                                          global.language('delete_sub_option'),
                                        ),
                                        style: ElevatedButton.styleFrom(
                                          backgroundColor: global.theme.negativeHighlightTextColor,
                                          padding: const EdgeInsets.symmetric(
                                            vertical: 5,
                                            horizontal: 15,
                                          ),
                                        ),
                                        onPressed: () {
                                          _optionChoices.removeAt(index);
                                          setState(() {
                                            _optionChoices = _optionChoices;
                                          });
                                        },
                                      ),
                                    ),
                                  ],
                                ),
                                trailing: Icon(
                                  Icons.keyboard_arrow_right_sharp,
                                ),
                                onTap: () async {
                                  var result = await Navigator.push(
                                    context,
                                    MaterialPageRoute(
                                      builder: (context) => OptionDetailScreen(
                                        choices: _optionChoices[index],
                                      ),
                                    ),
                                  );

                                  if (result != null) {
                                    _optionChoices[index] = result;
                                  }
                                },
                              ),
                            );
                          },
                        ),

                        SizedBox(height: size.height * 0.03),
                        // Widget
                        TitleLabel(
                          title: global.language('option_details'),
                          subtitle: global.language('option_details_example'),
                        ),
                        SizedBox(height: size.height * 0.03),
                        Align(
                          alignment: Alignment.centerLeft,
                          child: Container(
                            padding: EdgeInsets.all(2.0),
                            child: Text(
                              global.language(
                                'customer_must_select_option_question',
                              ),
                              style: TextStyle(fontSize: 16),
                            ),
                          ),
                        ),
                        Column(
                          children: [
                            RadioButton(
                              description: global.language(
                                'required_to_select',
                              ),
                              value: true,
                              groupValue: _isRequired,
                              onChanged: (value) {
                                setState(() {
                                  _isRequired = true;
                                });
                              },
                            ),
                            RadioButton(
                              description: global.language('not_required'),
                              value: false,
                              groupValue: _isRequired,
                              activeColor: global.theme.negativeHighlightTextColor,
                              onChanged: (value) {
                                setState(() {
                                  _isRequired = false;
                                });
                              },
                            ),
                          ],
                        ),

                        Align(
                          alignment: Alignment.centerLeft,
                          child: Container(
                            padding: EdgeInsets.all(2.0),
                            child: Text(
                              global.language(
                                'customer_can_select_how_many_sub_options_question',
                              ),
                              style: TextStyle(fontSize: 16),
                            ),
                          ),
                        ),

                        Column(
                          children: [
                            RadioButton(
                              description: global.language('one_item_only'),
                              value: 0,
                              groupValue: _choiceType,
                              onChanged: (value) {
                                setState(() {
                                  _choiceType = 0;
                                  _ismaxselected = false;
                                  _maxselected.text = "0";
                                });
                              },
                            ),
                            RadioButton(
                              description: global.language('multiple_items'),
                              value: 1,
                              groupValue: _choiceType,
                              onChanged: (value) {
                                setState(() {
                                  _choiceType = 1;
                                  _ismaxselected = true;
                                });
                              },
                            ),
                            Visibility(
                              visible: _ismaxselected,
                              child: Container(
                                padding: EdgeInsets.only(top: 15.0),
                                child: TextFormField(
                                  controller: _maxselected,
                                  keyboardType: TextInputType.number,
                                  decoration: InputDecoration(
                                    border: OutlineInputBorder(
                                      borderRadius: BorderRadius.circular(10.0),
                                    ),
                                    labelText: global.language(
                                      'maximum_selectable',
                                    ),
                                    isDense: true,
                                    contentPadding: const EdgeInsets.all(10),
                                  ),
                                ),
                              ),
                            ),
                          ],
                        ),
                        (!createMode)
                            ? Container(
                                width: double.infinity,
                                margin: EdgeInsets.all(8.0),
                                child: ElevatedButton.icon(
                                  icon: Icon(
                                    Icons.delete,
                                    color: global.theme.onPrimaryColor,
                                    size: 24.0,
                                  ),
                                  label: Text(
                                    global.language(
                                      'delete_supplement_option_data',
                                    ),
                                  ),
                                  style: ElevatedButton.styleFrom(
                                    backgroundColor: global.theme.negativeHighlightTextColor,
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
            ),
          ),
              ),
            ),
          );
          },
        ),
        floatingActionButton: FloatingActionButton(
          onPressed: () {
            if (_maxselected.text == '') {
              _maxselected.text = "0";
            }

            OptionModel option = OptionModel(
              guidfixed: widget.guidfixed.toString(),
              code: _uuid.toString(),
              name1: _name1.text,
              name2: _name2.text,
              name3: _name3.text,
              name4: _name4.text,
              name5: _name5.text,
              choicetype: _choiceType,
              maxselect: int.parse(_maxselected.text),
              isrequired: _isRequired,
              choices: _optionChoices,
            );

            if (createMode) {
              context.read<OptionBloc>().add(OptionSaved(option: option));
            } else {
              context.read<OptionBloc>().add(OptionUpdate(option: option));
            }
          },
          backgroundColor: global.theme.positiveHighlightTextColor,
          child: Icon(Icons.save),
        ),
      ),
    );
  }

  void showDeleteDialog(String id) {
    // set up the buttons
    Widget continueButton = TextButton(
      child: Text(global.language("confirm")),
      onPressed: () {
        if (id == 'save' || id == 'update') {
          Navigator.of(context).pop();
          Navigator.of(context).pushReplacement(
            MaterialPageRoute(builder: (_) => const OptionScreen()),
          );
        } else {
          context.read<OptionBloc>().add(OptionDelete(id: id));
          Navigator.of(context).pop();
        }
      },
    );
    // set up the AlertDialog
    AlertDialog alert = AlertDialog(
      title: Text(global.language("system_notification")),
      content: (id == 'save')
          ? Text(global.language("save_supplement_option_success"))
          : (id == 'update')
          ? Text(global.language("edit_supplement_option_success"))
          : Text(global.language("confirm_delete_supplement_option_question")),
      actions: [
        // cancelButton,
        continueButton,
      ],
    );
    // show the dialog
    showDialog(
      context: context,
      builder: (BuildContext context) {
        return alert;
      },
    );
  }
}

class TitleLabel extends StatelessWidget {
  final String title;
  final String subtitle;
  const TitleLabel({super.key, required this.title, this.subtitle = ""});

  @override
  Widget build(BuildContext context) {
    return Container(
      color: global.theme.surfaceColor,
      child: Padding(
        padding: const EdgeInsets.all(8.0),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: <Widget>[
            Text(title, style: Theme.of(context).textTheme.titleMedium),
            Text(subtitle, style: Theme.of(context).textTheme.bodySmall),
          ],
        ),
      ),
    );
  }
}
