import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/company_branch_model.dart';

class ReceiptConfigWidget extends StatelessWidget {
  final PosModel? posData;
  final ValueChanged<bool> onDataChanged;

  const ReceiptConfigWidget({
    super.key,
    required this.posData,
    required this.onDataChanged,
  });

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: EdgeInsets.all(10.0),
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
            // Header Receipt
            Text(
              global.language("header_receipt_pos"),
              style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
            ),
            SizedBox(height: 10),
            TextFormField(
              decoration: InputDecoration(
                hintText: global.language("header_receipt_pos"),
                border: const OutlineInputBorder(),
              ),
              controller: TextEditingController(
                text: posData!.headerreceiptpos,
              ),
              maxLines: 4,
              keyboardType: TextInputType.multiline,
              onChanged: (value) {
                onDataChanged(true);
                posData!.headerreceiptpos = value;
              },
            ),
            const SizedBox(height: 10),
            const Divider(),

            // Footer Receipt
            Text(
              global.language("footer_receipt_pos"),
              style: TextStyle(fontSize: 20, fontWeight: FontWeight.bold),
            ),
            SizedBox(height: 10),
            TextFormField(
              decoration: InputDecoration(
                hintText: global.language("footer_receipt_pos"),
                border: const OutlineInputBorder(),
              ),
              controller: TextEditingController(
                text: posData!.footerreceiptpos,
              ),
              maxLines: 4,
              keyboardType: TextInputType.multiline,
              onChanged: (value) {
                onDataChanged(true);
                posData!.footerreceiptpos = value;
              },
            ),
          ],
        ),
      ),
    );
  }
}
