// ignore_for_file: depend_on_referenced_packages

import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/transaction_model.dart';

/// Helper class สำหรับ Desktop Product Widgets
/// ใช้ class นี้เพื่อแยก desktop product widget logic ออกจาก main screen
/// รองรับทั้ง TransactionEditScreenState และ PurchaseOrderEditScreenState
class DesktopProductWidgets {
  final dynamic state;

  DesktopProductWidgets(this.state);

  /// Widget สำหรับแสดงคำอธิบายสินค้า (description row)
  Widget buildProductDescriptionWidget(
    double maxWidth,
    double sumWidth,
    String description,
  ) {
    List<Widget> widgets = [];
    if (description.isNotEmpty) {
      List<Widget> descriptions = [];

      for (var loop = 0; loop < state.headers.length; loop++) {
        String dataText = '';

        switch (state.headers[loop].code) {
          case "delete":
            dataText = "";
            break;
          case "line_number":
            dataText = "";
            break;
          case "barcode":
            dataText = "";
            break;
          case "product_name":
            dataText = "${global.language("note_label")} : $description";
            break;
          case "product_ware_house":
            dataText = "";
            break;
          case "product_location":
            dataText = "";
            break;
          case "product_to_ware_house":
            dataText = "";
            break;
          case "product_to_location":
            dataText = "";
            break;
          case "product_unit":
            dataText = "";
            break;
          case "product_qty":
            dataText = "";
            break;
          case "product_price_adjust":
            dataText = "";
            break;
          case "product_price":
            dataText = "";
            break;
          case "product_discount":
            dataText = "";
            break;
          case "product_amount":
            dataText = "";

            break;
        }

        if (state.headers[loop].code == 'delete') {
          descriptions.add(
            SizedBox(width: maxWidth * state.headers[loop].width / sumWidth),
          );
        } else {
          descriptions.add(
            Container(
              padding: (loop == 0)
                  ? const EdgeInsets.only(left: 0)
                  : (loop == state.headers.length - 1)
                  ? const EdgeInsets.only(right: 4, left: 4)
                  : const EdgeInsets.only(left: 4),
              width: maxWidth * state.headers[loop].width / sumWidth,
              child: Text(
                dataText,
                textAlign: state.headers[loop].textAlign,
                style: TextStyle(fontSize: 12),
              ),
            ),
          );
        }
      }
      widgets.add(
        Padding(
          padding: const EdgeInsets.only(bottom: 0),
          child: Row(children: descriptions),
        ),
      );
    }

    return Column(children: widgets);
  }

  /// Widget สำหรับแสดงรายการ options ของสินค้า
  Widget buildProductOptionsListDetailWidget(
    double maxWidth,
    double sumWidth,
    List<ExtraJsonListModel> opstions,
  ) {
    List<Widget> widgets = [];
    if (opstions.isNotEmpty) {
      for (var index = 0; index < opstions.length; index++) {
        List<Widget> dataOptions = [];

        for (var loop = 0; loop < state.headers.length; loop++) {
          String dataText = '';

          switch (state.headers[loop].code) {
            case "delete":
              dataText = "";
              break;
            case "line_number":
              dataText = "";
              break;
            case "barcode":
              dataText = opstions[index].barcode!;
              break;
            case "product_name":
              dataText =
                  "- ${global.activeLangName(opstions[index].itemnames!)}";
              break;
            case "product_ware_house":
              dataText = "";
              break;
            case "product_location":
              dataText = "";
              break;
            case "product_to_ware_house":
              dataText = "";
              break;
            case "product_to_location":
              dataText = "";
              break;
            case "product_unit":
              dataText = opstions[index].unit_name!;
              break;
            case "product_qty":
              dataText = global.formatNumber(
                (opstions[index].qty == 0) ? 1 : opstions[index].qty!,
              );
              break;
            case "product_price_adjust":
              dataText = "";
              break;
            case "product_price":
              dataText = global.formatNumber(opstions[index].price!);
              break;
            case "product_discount":
              dataText = "";
              break;
            case "product_amount":
              dataText = global.formatNumber(
                opstions[index].price! *
                    ((opstions[index].qty == 0) ? 1 : opstions[index].qty!),
              );

              break;
          }

          if (state.headers[loop].code == 'delete') {
            dataOptions.add(
              SizedBox(width: maxWidth * state.headers[loop].width / sumWidth),
            );
          } else {
            dataOptions.add(
              Container(
                padding: (loop == 0)
                    ? const EdgeInsets.only(left: 0)
                    : (loop == state.headers.length - 1)
                    ? const EdgeInsets.only(right: 4, left: 4)
                    : const EdgeInsets.only(left: 4),
                width: maxWidth * state.headers[loop].width / sumWidth,
                child: ElevatedButton(
                  style: ElevatedButton.styleFrom(
                    padding: const EdgeInsets.all(2),
                    alignment: state.headers[loop].alignment,
                    foregroundColor: global.theme.textColor,
                    backgroundColor: global.theme.cardColor,
                    shape: const RoundedRectangleBorder(
                      borderRadius: BorderRadius.all(Radius.circular(0)),
                    ),
                  ),
                  onPressed: null,
                  child: Text(
                    dataText,
                    textAlign: state.headers[loop].textAlign,
                    style: TextStyle(fontSize: 12),
                  ),
                ),
              ),
            );
          }
        }
        widgets.add(
          Padding(
            padding: const EdgeInsets.only(bottom: 0),
            child: Row(children: dataOptions),
          ),
        );
      }
    }

    return Column(children: widgets);
  }
}
