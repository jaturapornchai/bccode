import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/company_branch_model.dart';

class VatConfigWidget extends StatelessWidget {
  final PosModel? posData;
  final ValueChanged<bool> onDataChanged;

  const VatConfigWidget({
    super.key,
    required this.posData,
    required this.onDataChanged,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.only(left: 10, right: 10),
      child: Container(
        margin: const EdgeInsets.only(top: 10, bottom: 10),
        padding: EdgeInsets.all(10),
        decoration: BoxDecoration(
          border: Border.all(color: global.theme.primaryColor),
          borderRadius: BorderRadius.circular(5),
          color: global.theme.cardColor,
        ),
        child: Column(
          children: [
            // VAT Rate
            Text(
              global.language("vat_rate"),
              style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
            ),
            SizedBox(height: 10),
            TextFormField(
              decoration: InputDecoration(
                labelText: global.language("vat_rate"),
                border: const OutlineInputBorder(),
              ),
              controller: TextEditingController(
                text: posData!.vatrate.toString(),
              ),
              keyboardType: const TextInputType.numberWithOptions(
                decimal: true,
              ),
              inputFormatters: [global.NumberInputFormatter()],
              onChanged: (value) {
                onDataChanged(true);
                posData!.vatrate = double.parse(value);
              },
            ),
            const SizedBox(height: 10),
            const Divider(),

            // BOM Switch
            Row(
              mainAxisAlignment: MainAxisAlignment.start,
              children: [
                Expanded(
                  child: Row(
                    children: [
                      Switch(
                        focusNode: FocusNode(skipTraversal: true),
                        value: posData!.isbom!,
                        onChanged: (value) {
                          onDataChanged(true);
                          posData!.isbom = value;
                        },
                      ),
                      Expanded(
                        child: Text(
                          global.language("cut_stock_by_bom"),
                          overflow: TextOverflow.clip,
                        ),
                      ),
                    ],
                  ),
                ),
              ],
            ),
            const Divider(),

            // VAT Type Purchase
            _buildVatTypeSection(
              title: global.language("vattype_purchase"),
              groupValue: posData!.vattypepurchase,
              onChanged: (value) {
                onDataChanged(true);
                posData!.vattypepurchase = value;
              },
            ),
            const Divider(),

            // Inquiry Type Purchase
            _buildInquiryTypeSection(
              title: global.language("inquirytype_purchase"),
              groupValue: posData!.inquirytypepurchase,
              onChanged: (value) {
                onDataChanged(true);
                posData!.inquirytypepurchase = value;
              },
            ),
            const Divider(),

            // VAT Type Sale
            _buildVatTypeSection(
              title: global.language("vattype_sale"),
              groupValue: posData!.vattypesale,
              onChanged: (value) {
                onDataChanged(true);
                posData!.vattypesale = value;
              },
            ),
            const Divider(),

            // Inquiry Type Sale
            _buildInquiryTypeSection(
              title: global.language("inquirytype_sale"),
              groupValue: posData!.inquirytypesale,
              onChanged: (value) {
                onDataChanged(true);
                posData!.inquirytypesale = value;
              },
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildVatTypeSection({
    required String title,
    required int? groupValue,
    required ValueChanged<int?> onChanged,
  }) {
    return Column(
      children: [
        Text(
          title,
          style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
        ),
        SizedBox(height: 10),
        Row(
          mainAxisAlignment: MainAxisAlignment.start,
          children: [
            Expanded(
              child: Row(
                children: [
                  Radio(focusNode: FocusNode(skipTraversal: true), value: 0, groupValue: groupValue, onChanged: onChanged),
                  Expanded(
                    child: Text(
                      global.language("vat_exclude"),
                      overflow: TextOverflow.clip,
                    ),
                  ),
                ],
              ),
            ),
            Expanded(
              child: Row(
                children: [
                  Radio(focusNode: FocusNode(skipTraversal: true), value: 1, groupValue: groupValue, onChanged: onChanged),
                  Expanded(
                    child: Text(
                      global.language("vat_include"),
                      overflow: TextOverflow.clip,
                    ),
                  ),
                ],
              ),
            ),
            Expanded(
              child: Row(
                children: [
                  Radio(
                    focusNode: FocusNode(skipTraversal: true),
                    value: 2,
                    groupValue: groupValue,
                    activeColor: global.theme.infoHighlightTextColor,
                    onChanged: onChanged,
                  ),
                  Expanded(
                    child: Text(
                      global.language("vat_zero"),
                      overflow: TextOverflow.clip,
                    ),
                  ),
                ],
              ),
            ),
            Expanded(
              child: Row(
                children: [
                  Radio(
                    focusNode: FocusNode(skipTraversal: true),
                    value: 3,
                    groupValue: groupValue,
                    activeColor: global.theme.infoHighlightTextColor,
                    onChanged: onChanged,
                  ),
                  Expanded(
                    child: Text(
                      global.language("vat_none"),
                      overflow: TextOverflow.clip,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ],
    );
  }

  Widget _buildInquiryTypeSection({
    required String title,
    required int? groupValue,
    required ValueChanged<int?> onChanged,
  }) {
    return Column(
      children: [
        Text(
          title,
          style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
        ),
        SizedBox(height: 10),
        Row(
          mainAxisAlignment: MainAxisAlignment.start,
          children: [
            Expanded(
              child: Row(
                children: [
                  Radio(focusNode: FocusNode(skipTraversal: true), value: 0, groupValue: groupValue, onChanged: onChanged),
                  Expanded(
                    child: Text(
                      global.language("credit"),
                      overflow: TextOverflow.clip,
                    ),
                  ),
                ],
              ),
            ),
            Expanded(
              child: Row(
                children: [
                  Radio(focusNode: FocusNode(skipTraversal: true), value: 1, groupValue: groupValue, onChanged: onChanged),
                  Expanded(
                    child: Text(
                      global.language("cash"),
                      overflow: TextOverflow.clip,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ],
    );
  }
}
