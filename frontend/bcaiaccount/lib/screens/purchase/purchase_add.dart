// ignore_for_file: prefer_const_constructors

import 'package:flutter/material.dart';
import 'package:group_radio_button/group_radio_button.dart';
import 'package:smlaicloud/components/appbar.dart';
import 'package:smlaicloud/components/background_main.dart';
import 'package:smlaicloud/components/textfield_input.dart';
import 'package:smlaicloud/global.dart' as global;

class PurchaseAdd extends StatefulWidget {
  const PurchaseAdd({super.key});

  @override
  State<PurchaseAdd> createState() => _PurchaseAddState();
}

class _PurchaseAddState extends State<PurchaseAdd> {
  final memberId = TextEditingController();
  final fristName = TextEditingController();
  final lastName = TextEditingController();
  final phone = TextEditingController();
  String _purchaseType = "";
  String _taxType = "";
  final creditDay = TextEditingController();

  @override
  void initState() {
    super.initState();
    _purchaseType = global.language('cash_purchase');
    _taxType = global.language('tax_included');
  }

  final houseNo = TextEditingController();
  final villageNo = TextEditingController();
  final village = TextEditingController();
  final tambon = TextEditingController();
  final amper = TextEditingController();
  final province = TextEditingController();
  final zipCode = TextEditingController();

  @override
  Widget build(BuildContext context) {
    Size size = MediaQuery.of(context).size;

    return Scaffold(
      appBar: BaseAppBar(
        title: Text(global.language('create_purchase_document')),
        appBar: AppBar(),
        widgets: const <Widget>[],
      ),
      body: BackgroundMain(
        child: SingleChildScrollView(
          child: Column(
            children: [
              Container(
                margin: const EdgeInsets.only(left: 8.0, right: 8.0),
                height: size.height * 0.34,
                child: Column(
                  children: [
                    Expanded(
                      child: Card(
                        margin: EdgeInsets.all(8.0),
                        child: Container(
                          margin: EdgeInsets.all(10.0),
                          child: Column(
                            children: [
                              Container(
                                padding: const EdgeInsets.only(
                                  left: 8.0,
                                  right: 8.0,
                                  top: 15.0,
                                ),
                                child: Row(
                                  children: [
                                    Expanded(
                                      flex: 1,
                                      child: Align(
                                        alignment: Alignment.topCenter,
                                        child: Text(
                                          global.language('document_number'),
                                        ),
                                      ),
                                    ),
                                    Expanded(
                                      flex: 2,
                                      child: TextFieldInput(
                                        keyboardType: TextInputType.text,
                                        controller: houseNo,
                                        labelText: global.language(
                                          'document_number',
                                        ),
                                      ),
                                    ),
                                    Expanded(
                                      flex: 1,
                                      child: Align(
                                        alignment: Alignment.topCenter,
                                        child: Text(
                                          global.language('document_date'),
                                        ),
                                      ),
                                    ),
                                    Expanded(
                                      flex: 2,
                                      child: TextFieldInput(
                                        keyboardType: TextInputType.text,
                                        controller: villageNo,
                                        labelText: global.language(
                                          'document_date',
                                        ),
                                      ),
                                    ),
                                    Expanded(
                                      flex: 1,
                                      child: Align(
                                        alignment: Alignment.topCenter,
                                        child: Text(
                                          global.language('reference_document'),
                                        ),
                                      ),
                                    ),
                                    Expanded(
                                      flex: 2,
                                      child: TextFieldInput(
                                        keyboardType: TextInputType.text,
                                        controller: village,
                                        labelText: global.language(
                                          'reference_document',
                                        ),
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                              Container(
                                padding: const EdgeInsets.only(
                                  left: 8.0,
                                  right: 8.0,
                                  top: 15.0,
                                ),
                                child: Row(
                                  children: [
                                    Expanded(
                                      flex: 1,
                                      child: Align(
                                        alignment: Alignment.topCenter,
                                        child: Text(
                                          global.language('creditor'),
                                        ),
                                      ),
                                    ),
                                    Expanded(
                                      flex: 2,
                                      child: TextFieldInput(
                                        keyboardType: TextInputType.text,
                                        controller: houseNo,
                                        labelText: global.language('creditor'),
                                      ),
                                    ),
                                    Expanded(
                                      flex: 1,
                                      child: Align(
                                        alignment: Alignment.topCenter,
                                        child: Text(
                                          global.language('tax_invoice_number'),
                                        ),
                                      ),
                                    ),
                                    Expanded(
                                      flex: 2,
                                      child: TextFieldInput(
                                        keyboardType: TextInputType.text,
                                        controller: villageNo,
                                        labelText: global.language(
                                          'tax_invoice_number',
                                        ),
                                      ),
                                    ),
                                    Expanded(
                                      flex: 1,
                                      child: Align(
                                        alignment: Alignment.topCenter,
                                        child: Text(
                                          global.language('tax_invoice_date'),
                                        ),
                                      ),
                                    ),
                                    Expanded(
                                      flex: 2,
                                      child: TextFieldInput(
                                        keyboardType: TextInputType.text,
                                        controller: village,
                                        labelText: global.language(
                                          'tax_invoice_date',
                                        ),
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                              Container(
                                padding: const EdgeInsets.only(
                                  left: 8.0,
                                  right: 8.0,
                                  top: 0.0,
                                ),
                                child: Row(
                                  children: [
                                    Expanded(
                                      flex: 1,
                                      child: Align(
                                        alignment: Alignment.topCenter,
                                        child: Text(
                                          global.language('purchase_type'),
                                        ),
                                      ),
                                    ),
                                    Expanded(
                                      flex: 2,
                                      child: Row(
                                        children: [
                                          RadioButton(
                                            description: global.language(
                                              'cash_purchase',
                                            ),
                                            value: global.language(
                                              'cash_purchase',
                                            ),
                                            groupValue: _purchaseType,
                                            onChanged: (value) => setState(
                                              () => _purchaseType = value
                                                  .toString(),
                                            ),
                                          ),
                                          RadioButton(
                                            description: global.language(
                                              'credit_purchase',
                                            ),
                                            value: global.language(
                                              'credit_purchase',
                                            ),
                                            groupValue: _purchaseType,
                                            activeColor: Colors.red,
                                            onChanged: (value) => setState(
                                              () => _purchaseType = value
                                                  .toString(),
                                            ),
                                          ),
                                        ],
                                      ),
                                    ),
                                    Expanded(
                                      flex: 1,
                                      child: Align(
                                        alignment: Alignment.topCenter,
                                        child: Text(
                                          global.language('tax_type'),
                                        ),
                                      ),
                                    ),
                                    Expanded(
                                      flex: 3,
                                      child: Row(
                                        children: [
                                          RadioButton(
                                            description: global.language(
                                              'tax_included',
                                            ),
                                            value: global.language(
                                              'tax_included',
                                            ),
                                            groupValue: _taxType,
                                            onChanged: (value) => setState(
                                              () => _taxType = value.toString(),
                                            ),
                                          ),
                                          RadioButton(
                                            description: global.language(
                                              'tax_excluded',
                                            ),
                                            value: global.language(
                                              'tax_excluded',
                                            ),
                                            groupValue: _taxType,
                                            activeColor: Colors.red,
                                            onChanged: (value) => setState(
                                              () => _taxType = value.toString(),
                                            ),
                                          ),
                                          RadioButton(
                                            description: global.language(
                                              'zero_rated_tax',
                                            ),
                                            value: global.language(
                                              'zero_rated_tax',
                                            ),
                                            groupValue: _taxType,
                                            activeColor: Colors.orange,
                                            onChanged: (value) => setState(
                                              () => _taxType = value.toString(),
                                            ),
                                          ),
                                        ],
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ],
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
              Container(
                margin: const EdgeInsets.only(left: 8.0, right: 8.0),
                height: size.height * 0.34,
                child: Column(
                  children: [
                    Expanded(
                      child: Card(
                        margin: EdgeInsets.all(8.0),
                        child: Container(
                          margin: EdgeInsets.all(10.0),
                          child: Column(
                            children: [
                              Container(
                                padding: const EdgeInsets.only(
                                  left: 8.0,
                                  right: 8.0,
                                  top: 15.0,
                                ),
                                child: Row(
                                  children: [
                                    Expanded(
                                      child: Align(
                                        alignment: Alignment.topCenter,
                                        child: Text(
                                          global.language('document_number'),
                                        ),
                                      ),
                                    ),
                                  ],
                                ),
                              ),
                            ],
                          ),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
      floatingActionButton: FloatingActionButton(
        onPressed: () {},
        backgroundColor: Colors.green,
        child: const Icon(Icons.save),
      ),
    );
  }
}
