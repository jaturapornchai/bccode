import 'package:flutter/material.dart';
import 'package:smlaicloud/model/report_condition_model.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/services/stock_report_lookup_service.dart';

class ReportStockConditionClass {
  final VoidCallback? onStateUpdate;

  List<ConditionWareHouseSelectModel> conditionWareHouseCodeList = [];
  List<SelectedProductItemStruct> conditionItemCodeList = [];
  bool showOnlyBalance = true;
  bool showOnlyWithMovement = true;

  ReportStockConditionClass({this.onStateUpdate}) {
    conditionWareHouseCodeList = [];
    conditionItemCodeList = [];
    if (onStateUpdate != null) {
      onStateUpdate!();
    }
  }

  Future<void> reloadWareHouseCode() async {
    conditionWareHouseCodeList.clear();

    // 1. ดึง barcodes จาก item codes ผ่าน Backend API (ปลอดภัยจาก SQL injection)
    List<String> barcodeList = [];
    if (conditionItemCodeList.isNotEmpty) {
      final itemCodes = conditionItemCodeList.map((e) => e.itemCode).toList();
      barcodeList = await StockReportLookupService.getBarcodesByItemCodes(itemCodes);
      // เพิ่ม item codes เข้าไปด้วย (ใช้เป็น fallback)
      barcodeList.addAll(itemCodes);
    }

    // 2. ดึง warehouses + locations ผ่าน Backend API
    final warehouses = await StockReportLookupService.getWarehousesAndLocations(barcodeList);

    for (final wh in warehouses) {
      final whModel = ConditionWareHouseSelectModel(wh.whcode, wh.whcode);
      for (final loc in wh.locations) {
        whModel.locations.add(ConditionLocationSelectModel(loc, loc));
      }
      conditionWareHouseCodeList.add(whModel);
    }

    if (onStateUpdate != null) {
      onStateUpdate!();
    }
  }

  // เลือกคลังสินค้า
  Widget selectWareHouseWidget(
      {required Color primaryColor, required Color secondaryColor}) {
    return Container(
      width: double.infinity,
      margin: const EdgeInsets.only(top: 8.0),
      padding: const EdgeInsets.all(8.0),
      decoration: BoxDecoration(
        color: global.theme.cardColor,
        border: Border.all(color: global.theme.textSecondaryColor),
        borderRadius: BorderRadius.circular(4.0),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Wrap(
            spacing: 4.0,
            runSpacing: 4.0,
            children: [
              ElevatedButton.icon(
                onPressed: () {
                  for (var item in conditionWareHouseCodeList) {
                    item.isSelected = true;
                  }

                  if (onStateUpdate != null) {
                    onStateUpdate!();
                  }
                },
                icon:
                    Icon(Icons.check_box, size: 18), // เพิ่มไอคอน global.language("add")
                label: Text(global.language("select_all_warehouses")),
                style: ElevatedButton.styleFrom(
                  backgroundColor: global.theme.infoHighlightTextColor, // สีปุ่มน้ำเงินเข้ม
                  foregroundColor: global.theme.onPrimaryColor, // สีตัวอักษรและไอคอน
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(4.0),
                  ),
                ),
              ),
              SizedBox(width: 8),
              ElevatedButton.icon(
                onPressed: (conditionWareHouseCodeList.isNotEmpty)
                    ? () {
                        for (var item in conditionWareHouseCodeList) {
                          item.isSelected = false;
                        }

                        if (onStateUpdate != null) {
                          onStateUpdate!();
                        }
                      }
                    : null,
                icon: Icon(Icons.delete, size: 18),
                label: Text(global.language("deselect_all_warehouses")),
                style: ElevatedButton.styleFrom(
                  backgroundColor: global.theme.warningHighlightTextColor, // สีปุ่มส้ม
                  foregroundColor: global.theme.onPrimaryColor,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(4.0),
                  ),
                ),
              ),
              const SizedBox(width: 12),
              Text(
                global.language("select_warehouse_for_report"),
                style: TextStyle(
                  color: global.theme.textColor, // สีเทาเข้มเพื่อความนุ่มนวล
                  fontSize: 14,
                ),
              ),
            ],
          ),
          const SizedBox(height: 8),
          Wrap(
            alignment: WrapAlignment.start,
            crossAxisAlignment: WrapCrossAlignment.center,
            spacing: 4.0,
            runSpacing: 4.0,
            children: conditionWareHouseCodeList.map((item) {
              return ElevatedButton(
                onPressed: () {
                  item.isSelected = !item.isSelected;
                  if (onStateUpdate != null) {
                    onStateUpdate!();
                  }
                },
                style: ElevatedButton.styleFrom(
                  backgroundColor: item.isSelected
                      ? global.theme.infoHighlightTextColor
                      : global.theme.infoHighlightTextColor,
                  foregroundColor: global.theme.onPrimaryColor,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(4.0),
                  ),
                ),
                child: Row(mainAxisSize: MainAxisSize.min, children: [
                  Row(
                    children: [
                      Icon(
                        (item.isSelected) ? Icons.check_circle : Icons.circle,
                        size: 16,
                        color: global.theme.cardColor,
                      ),
                      const SizedBox(width: 4),
                    ],
                  ),
                  Text(item.code),
                ]),
              );
            }).toList(),
          ),
        ],
      ),
    );
  }

  // เลือกที่เก็บสินค้า
  Widget selectLocationWidget(
      {required Color primaryColor, required Color secondaryColor}) {
    List<Widget> locationWidgetList = [];
    for (var wareHouse in conditionWareHouseCodeList) {
      if (wareHouse.isSelected) {
        locationWidgetList.add(Container(
          width: double.infinity,
          margin: const EdgeInsets.only(top: 8.0),
          padding: const EdgeInsets.all(8.0),
          decoration: BoxDecoration(
            color: global.theme.cardColor,
            border: Border.all(color: global.theme.textSecondaryColor),
            borderRadius: BorderRadius.circular(4.0),
          ),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                mainAxisAlignment: MainAxisAlignment.spaceBetween,
                children: [
                  Text(
                    "คลังสินค้า ${wareHouse.code}",
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  IconButton(
                    icon: Icon(Icons.close),
                    onPressed: () {
                      wareHouse.isSelected = false;
                      if (onStateUpdate != null) {
                        onStateUpdate!();
                      }
                    },
                  ),
                ],
              ),
              Wrap(
                alignment: WrapAlignment.start,
                crossAxisAlignment: WrapCrossAlignment.center,
                spacing: 4.0,
                runSpacing: 4.0,
                children: [
                  ElevatedButton.icon(
                    onPressed: () {
                      for (var item in wareHouse.locations) {
                        item.isSelected = true;
                      }
                      if (onStateUpdate != null) {
                        onStateUpdate!();
                      }
                    },
                    icon: Icon(Icons.check_box,
                        size: 18), // เพิ่มไอคอน global.language("add")
                    label: Text(global.language("select_all_locations")),
                    style: ElevatedButton.styleFrom(
                      backgroundColor:
                          global.theme.infoHighlightTextColor, // สีปุ่มน้ำเงินเข้ม
                      foregroundColor: global.theme.onPrimaryColor, // สีตัวอักษรและไอคอน
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(4.0),
                      ),
                    ),
                  ),
                  SizedBox(width: 8),
                  ElevatedButton.icon(
                    onPressed: (wareHouse.locations.isNotEmpty)
                        ? () {
                            for (var item in wareHouse.locations) {
                              item.isSelected = false;
                            }
                            if (onStateUpdate != null) {
                              onStateUpdate!();
                            }
                          }
                        : null,
                    icon: Icon(Icons.delete,
                        size: 18), // เพิ่มไอคอน global.language("database_master_info.refresh")
                    label: Text(global.language("deselect_all_locations")),
                    style: ElevatedButton.styleFrom(
                      backgroundColor: global.theme.warningHighlightTextColor, // สีปุ่มส้ม
                      foregroundColor: global.theme.onPrimaryColor,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(4.0),
                      ),
                    ),
                  ),
                  const SizedBox(width: 12),
                  Text(
                    global.language("select_location_for_report"),
                    style: TextStyle(
                      color: global.theme.textColor,
                      fontSize: 14,
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 8),
              Wrap(
                alignment: WrapAlignment.start,
                crossAxisAlignment: WrapCrossAlignment.center,
                spacing: 4.0,
                runSpacing: 4.0,
                children: wareHouse.locations.map((item) {
                  return ElevatedButton(
                    onPressed: () {
                      item.isSelected = !item.isSelected;
                      if (onStateUpdate != null) {
                        onStateUpdate!();
                      }
                    },
                    style: ElevatedButton.styleFrom(
                      backgroundColor: item.isSelected
                          ? global.theme.infoHighlightTextColor
                          : global.theme.infoHighlightTextColor,
                      foregroundColor: global.theme.onPrimaryColor,
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(4.0),
                      ),
                    ),
                    child: Row(mainAxisSize: MainAxisSize.min, children: [
                      Row(
                        children: [
                          Icon(
                            (item.isSelected)
                                ? Icons.check_circle
                                : Icons.circle,
                            size: 16,
                            color: global.theme.cardColor,
                          ),
                          const SizedBox(width: 4),
                        ],
                      ),
                      Text(item.code),
                    ]),
                  );
                }).toList(),
              ),
            ],
          ),
        ));
      }
    }
    return Column(children: locationWidgetList);
  }
}
