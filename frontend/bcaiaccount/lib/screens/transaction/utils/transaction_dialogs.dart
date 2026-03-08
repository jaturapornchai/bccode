// ignore_for_file: deprecated_member_use

import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/model/location_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/model/warehouse_model.dart';
import 'package:smlaicloud/imports_bloc.dart';
import 'package:smlaicloud/global.dart' as global;

/// Helper class สำหรับ dialog functions
/// รองรับทั้ง TransactionEditScreenState และ PurchaseOrderEditScreenState
class TransactionDialogs {
  final dynamic state;

  TransactionDialogs(this.state);

  /// แสดง Dialog สำหรับกรอก Barcode
  void showBarcodeDialog(BuildContext context) async {
    TextEditingController barcode = TextEditingController();
    FocusNode barcodefocus = FocusNode();
    barcode.text = "";
    return showDialog(
      context: context,
      barrierDismissible: true, // Allows tapping outside the dialog to close it
      builder: (BuildContext context) {
        return AlertDialog(
          title: Text(global.language("barcode")),
          content: SizedBox(
            width: (global.isMobileScreen(context)) ? 350 : 400,
            child: TextField(
              autofocus: true,
              controller: barcode,
              focusNode: barcodefocus,
              decoration: InputDecoration(
                labelText: '',
                border: const OutlineInputBorder(),
                suffixIcon: IconButton(
                  icon: const Icon(Icons.clear),
                  onPressed: () {
                    barcode.text = '';
                  },
                ),
              ),
              onSubmitted: (value) {
                if (value.trim().isNotEmpty) {
                  double inputQty = 1;
                  String barcodeValue = "";
                  if (value.trim().contains("*")) {
                    String numberValue = value.trim().split("*")[0];
                    if (numberValue != "" && numberValue != "0") {
                      inputQty = double.parse(value.trim().split("*")[0]);
                    }
                    barcodeValue = value.trim().split("*")[1];
                  } else {
                    barcodeValue = value.trim();
                  }
                  state.productBarcodeRepository
                      .getProductBarcodeDetail(barcodeValue)
                      .then((value) {
                        if (value.success && value.data != null) {
                          ProductBarcodeModel result =
                              ProductBarcodeModel.fromJson(value.data);

                          if (state.widget.type ==
                                  global.TransactionTypeEnum.sale ||
                              state.widget.type ==
                                  global.TransactionTypeEnum.saleorder ||
                              state.widget.type ==
                                  global.TransactionTypeEnum.salereturn) {
                            if (result.itemtype == 3) {
                              return;
                            }
                          }

                          state.screenData.details!.add(
                            TransactionDetailModel(
                              docdatetime: DateTime.now()
                                  .toLocal()
                                  .toIso8601String(),
                              itemguid: result.guidfixed,
                              barcode: result.barcode!,
                              itemcode: result.itemcode ?? "",
                              itemnames: result.names,
                              unitcode: result.itemunitcode,
                              qty: inputQty,
                              price: state.getPrice(result.prices),
                              discount: '',
                              sumofcost: 0,
                              sumamount: 0,
                              remark: '',
                              linenumber: 0,
                              whcode: state.defaultWarehouse,
                              whnames: state.defaultWarehouseNames,
                              shelfcode: '',
                              locationcode: state.defaultLocation,
                              locationnames: state.defaultLocationNames,
                              totalvaluevat: 0,
                              totalqty: 0,
                              standvalue: result.standvalue!,
                              dividevalue: result.dividevalue!,
                              multiunit: true,
                              unitnames: result.itemunitnames,
                              calcflag: state.calcflag,
                              vattype: 0,
                              averagecost: 0,
                              sumamountexcludevat: 0,
                              discountamount: 0,
                              ispos: 0,
                              laststatus: 0,
                              itemtype: 0,
                              inquirytype: 0,
                              priceexcludevat: 0,
                              taxtype: result.taxtype!,
                              towhcode: state.defaultToWarehouse,
                              towhnames: state.defaultToWarehouseNames,
                              tolocationcode: state.defaultToLocation,
                              tolocationnames: state.defaultToLocationNames,
                              vatcal: result.vatcal,
                              refbarcodes: result.refbarcodes,
                              manufacturerguid: result.manufacturerguid,
                            ),
                          );

                          state.transactionCalculator.calTotalValue();
                          barcode.text = '';
                          Future.delayed(const Duration(milliseconds: 200), () {
                            FocusScope.of(context).requestFocus(barcodefocus);
                          });
                        }
                        state.setState(() {});
                      })
                      .onError((error, stackTrace) {
                        barcode.text = '';
                        Future.delayed(const Duration(milliseconds: 200), () {
                          FocusScope.of(context).requestFocus(barcodefocus);
                        });
                      });
                } else {}
              },
            ),
          ),
          actions: <Widget>[
            TextButton(
              child: Text(global.language("close")),
              onPressed: () {
                Navigator.pop(context);
              },
            ),
          ],
        );
      },
    );
  }
}

/// แสดง Dialog สำหรับแก้ไขส่วนลด
Future<String?> showDiscountDialog(
  BuildContext context,
  String currentdiscount,
  String itemnames,
) async {
  TextEditingController discount = TextEditingController();
  discount.text = currentdiscount.toString();
  return showDialog<String?>(
    context: context,
    barrierDismissible: true, // Allows tapping outside the dialog to close it
    builder: (BuildContext context) {
      // เลือกข้อความทั้งหมดหลังจาก widget สร้างเสร็จ
      WidgetsBinding.instance.addPostFrameCallback((_) {
        if (discount.text.isNotEmpty) {
          discount.selection = TextSelection(
            baseOffset: 0,
            extentOffset: discount.text.length,
          );
        }
      });

      return AlertDialog(
        title: Text('$itemnames ${global.language("discount")}'),
        content: SizedBox(
          width: (global.isMobileScreen(context)) ? 350 : 400,
          child: TextField(
            autofocus: true,
            controller: discount,
            decoration: InputDecoration(
              labelText: global.language("enter_discount"),
              border: const OutlineInputBorder(),
              suffixIcon: IconButton(
                icon: const Icon(Icons.clear),
                onPressed: () {
                  discount.text = '';
                },
              ),
            ),
            onSubmitted: (value) {
              Navigator.pop(context, discount.text);
            },
          ),
        ),
        actions: <Widget>[
          TextButton(
            child: Text(global.language("close")),
            onPressed: () {
              Navigator.pop(context, currentdiscount.toString());
            },
          ),
          TextButton(
            child: Text(global.language("update")),
            onPressed: () {
              Navigator.pop(context, discount.text);
            },
          ),
        ],
      );
    },
  );
}

/// แสดง Dialog สำหรับแก้ไขราคา
Future<String?> showPriceDialog(
  BuildContext context,
  double currentprice,
  String itemnames,
) async {
  String initialValue = currentprice > 0 ? currentprice.toString() : '';

  return await global.showNumericInputDialog(
    context,
    title: '$itemnames ${global.language("price")}',
    initialValue: initialValue,
    maxLength: 12,
    decimalPlaces: 2, // Allow 2 decimal places for price
    allowNegative: false,
  );
}

/// แสดง Dialog สำหรับแก้ไขจำนวน
Future<String?> showQtyDialog(
  BuildContext context,
  double currentqty,
  String itemnames,
) async {
  String initialValue = currentqty > 0 ? currentqty.toString() : '';

  return await global.showNumericInputDialog(
    context,
    title: '$itemnames ${global.language("qty")}',
    initialValue: initialValue,
    maxLength: 10,
    decimalPlaces: 3, // Allow 3 decimal places for quantity
    allowNegative: false,
  );
}

/// แสดง Dialog สำหรับแก้ไขหมายเหตุ
Future<String?> showDescriptionDialog(
  BuildContext context,
  String currentDescription,
  String itemnames,
) async {
  TextEditingController description = TextEditingController();
  description.text = currentDescription;

  return showDialog<String?>(
    context: context,
    barrierDismissible: true,
    builder: (BuildContext context) {
      return AlertDialog(
        title: Text('${global.language("note_label")}: $itemnames'),
        content: SizedBox(
          width: (global.isMobileScreen(context)) ? 350 : 400,
          child: TextField(
            autofocus: true,
            controller: description,
            maxLines: 4,
            decoration: InputDecoration(
              labelText: global.language('note_label'),
              border: const OutlineInputBorder(),
              suffixIcon: IconButton(
                icon: const Icon(Icons.clear),
                onPressed: () {
                  description.clear();
                },
              ),
            ),
            onSubmitted: (value) {
              Navigator.pop(context, description.text);
            },
          ),
        ),
        actions: <Widget>[
          TextButton(
            child: Text(global.language("close")),
            onPressed: () {
              Navigator.pop(context, currentDescription);
            },
          ),
          TextButton(
            child: Text(global.language("update")),
            onPressed: () {
              Navigator.pop(context, description.text);
            },
          ),
        ],
      );
    },
  );
}

/// แสดง Dialog สำหรับเลือกที่เก็บจากคลัง
Future<LocationModel?> showWareHouseLocationDialog(
  BuildContext context,
  String itemguid,
  String whcode,
) async {
  context.read<WarehouseBloc>().add(WarehouseGetByCode(code: whcode));
  late List<LocationModel> location = [];
  return showDialog<LocationModel?>(
    context: context,
    barrierDismissible: true,
    builder: (BuildContext context) {
      return BlocBuilder<WarehouseBloc, WarehouseState>(
        builder: (context, state) {
          if (state is WarehouseGetSuccess) {
            if (state.warehouse.location.isNotEmpty) {
              location = state.warehouse.location;
            }
          }
          return Dialog(
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
            child: Container(
              width: (global.isMobileScreen(context)) ? 350 : 420,
              constraints: BoxConstraints(maxHeight: 500),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  // Header
                  Container(
                    padding: const EdgeInsets.all(16),
                    decoration: const BoxDecoration(
                      color: Color(0xFF2A6F97),
                      borderRadius: BorderRadius.only(
                        topLeft: Radius.circular(16),
                        topRight: Radius.circular(16),
                      ),
                    ),
                    child: Row(
                      children: [
                        const Icon(Icons.location_on, color: Colors.white, size: 24),
                        SizedBox(width: 12),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                global.language("select_location_from_warehouse"),
                                style: const TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.bold),
                              ),
                              Text(
                                whcode,
                                style: const TextStyle(color: Colors.white70, fontSize: 13),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
                  // Content
                  if (state is WarehouseGetSuccess)
                    Flexible(
                      child: ListView.separated(
                        shrinkWrap: true,
                        padding: const EdgeInsets.symmetric(vertical: 8),
                        itemCount: location.length,
                        separatorBuilder: (context, index) => const Divider(height: 1),
                        itemBuilder: (BuildContext context, int index) {
                          final loc = location[index];
                          final locName = loc.names.firstWhere((ele) => ele.code == global.userLanguage, orElse: () => loc.names.first).name;
                          return Material(
                            color: Colors.transparent,
                            child: InkWell(
                              onTap: () => Navigator.pop(context, loc),
                              child: Padding(
                                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                                child: Row(
                                  children: [
                                    Container(
                                      padding: const EdgeInsets.all(8),
                                      decoration: BoxDecoration(
                                        color: const Color(0xFF2A6F97).withOpacity(0.1),
                                        borderRadius: BorderRadius.circular(8),
                                      ),
                                      child: const Icon(Icons.inventory_2_outlined, color: Color(0xFF2A6F97), size: 20),
                                    ),
                                    const SizedBox(width: 12),
                                    Expanded(
                                      child: Column(
                                        crossAxisAlignment: CrossAxisAlignment.start,
                                        children: [
                                          Text(
                                            loc.code,
                                            style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: Color(0xFF2A6F97)),
                                          ),
                                          const SizedBox(height: 2),
                                          Text(
                                            locName,
                                            style: TextStyle(fontSize: 13, color: Colors.grey[600]),
                                          ),
                                        ],
                                      ),
                                    ),
                                    const Icon(Icons.chevron_right, color: Colors.grey),
                                  ],
                                ),
                              ),
                            ),
                          );
                        },
                      ),
                    )
                  else
                    const Padding(
                      padding: EdgeInsets.all(24),
                      child: CircularProgressIndicator(),
                    ),
                  // Footer
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      border: Border(top: BorderSide(color: Colors.grey[300]!)),
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.end,
                      children: [
                        TextButton(
                          onPressed: () => Navigator.pop(context, LocationModel(code: '')),
                          style: TextButton.styleFrom(
                            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
                          ),
                          child: Text(global.language("close"), style: TextStyle(color: Color(0xFF2A6F97))),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          );
        },
      );
    },
  );
}

/// แสดง Dialog สำหรับเลือกที่เก็บเริ่มต้นจากคลัง
Future<LocationModel?> showWareHouseLocationDefualtDialog(
  BuildContext context,
  String whcode,
) async {
  context.read<WarehouseBloc>().add(WarehouseGetByCode(code: whcode));
  late List<LocationModel> location = [];
  return showDialog<LocationModel?>(
    context: context,
    barrierDismissible: true,
    builder: (BuildContext context) {
      return BlocBuilder<WarehouseBloc, WarehouseState>(
        builder: (context, state) {
          if (state is WarehouseGetSuccess) {
            if (state.warehouse.location.isNotEmpty) {
              location = state.warehouse.location;
            }
          }
          return Dialog(
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
            child: Container(
              width: (global.isMobileScreen(context)) ? 350 : 420,
              constraints: BoxConstraints(maxHeight: 500),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  // Header
                  Container(
                    padding: const EdgeInsets.all(16),
                    decoration: const BoxDecoration(
                      color: Color(0xFF2A6F97),
                      borderRadius: BorderRadius.only(
                        topLeft: Radius.circular(16),
                        topRight: Radius.circular(16),
                      ),
                    ),
                    child: Row(
                      children: [
                        const Icon(Icons.location_on, color: Colors.white, size: 24),
                        SizedBox(width: 12),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                global.language("select_location_from_warehouse"),
                                style: const TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.bold),
                              ),
                              Text(
                                whcode,
                                style: const TextStyle(color: Colors.white70, fontSize: 13),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),
                  // Content
                  if (state is WarehouseGetSuccess)
                    Flexible(
                      child: ListView.separated(
                        shrinkWrap: true,
                        padding: const EdgeInsets.symmetric(vertical: 8),
                        itemCount: location.length,
                        separatorBuilder: (context, index) => const Divider(height: 1),
                        itemBuilder: (BuildContext context, int index) {
                          final loc = location[index];
                          final locName = loc.names.firstWhere((ele) => ele.code == global.userLanguage, orElse: () => loc.names.first).name;
                          return Material(
                            color: Colors.transparent,
                            child: InkWell(
                              onTap: () => Navigator.pop(context, loc),
                              child: Padding(
                                padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                                child: Row(
                                  children: [
                                    Container(
                                      padding: const EdgeInsets.all(8),
                                      decoration: BoxDecoration(
                                        color: const Color(0xFF2A6F97).withOpacity(0.1),
                                        borderRadius: BorderRadius.circular(8),
                                      ),
                                      child: const Icon(Icons.inventory_2_outlined, color: Color(0xFF2A6F97), size: 20),
                                    ),
                                    const SizedBox(width: 12),
                                    Expanded(
                                      child: Column(
                                        crossAxisAlignment: CrossAxisAlignment.start,
                                        children: [
                                          Text(
                                            loc.code,
                                            style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: Color(0xFF2A6F97)),
                                          ),
                                          const SizedBox(height: 2),
                                          Text(
                                            locName,
                                            style: TextStyle(fontSize: 13, color: Colors.grey[600]),
                                          ),
                                        ],
                                      ),
                                    ),
                                    const Icon(Icons.chevron_right, color: Colors.grey),
                                  ],
                                ),
                              ),
                            ),
                          );
                        },
                      ),
                    )
                  else
                    const Padding(
                      padding: EdgeInsets.all(24),
                      child: CircularProgressIndicator(),
                    ),
                  // Footer
                  Container(
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      border: Border(top: BorderSide(color: Colors.grey[300]!)),
                    ),
                    child: Row(
                      mainAxisAlignment: MainAxisAlignment.end,
                      children: [
                        TextButton(
                          onPressed: () => Navigator.pop(context, LocationModel(code: '')),
                          style: TextButton.styleFrom(
                            padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
                          ),
                          child: Text(global.language("close"), style: TextStyle(color: Color(0xFF2A6F97))),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          );
        },
      );
    },
  );
}

/// แสดง Dialog สำหรับเลือกคลังสินค้า
Future<WarehouseModel?> showWareHouseDialog(
  BuildContext context,
  String itemguid,
  List<WarehouseModel> warehouseList,
) async {
  return showDialog<WarehouseModel?>(
    context: context,
    barrierDismissible: true,
    builder: (BuildContext context) {
      return Dialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        child: Container(
          width: (global.isMobileScreen(context)) ? 350 : 420,
          constraints: BoxConstraints(maxHeight: 500),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // Header
              Container(
                padding: const EdgeInsets.all(16),
                decoration: const BoxDecoration(
                  color: Color(0xFF2A6F97),
                  borderRadius: BorderRadius.only(
                    topLeft: Radius.circular(16),
                    topRight: Radius.circular(16),
                  ),
                ),
                child: Row(
                  children: [
                    const Icon(Icons.warehouse_outlined, color: Colors.white, size: 24),
                    SizedBox(width: 12),
                    Expanded(
                      child: Text(
                        global.language("select_warehouse"),
                        style: const TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.bold),
                      ),
                    ),
                  ],
                ),
              ),
              // Content
              Flexible(
                child: ListView.separated(
                  shrinkWrap: true,
                  padding: const EdgeInsets.symmetric(vertical: 8),
                  itemCount: warehouseList.length,
                  separatorBuilder: (context, index) => const Divider(height: 1),
                  itemBuilder: (BuildContext context, int index) {
                    final wh = warehouseList[index];
                    final whName = wh.names.firstWhere((ele) => ele.code == global.userLanguage, orElse: () => wh.names.first).name;
                    return Material(
                      color: Colors.transparent,
                      child: InkWell(
                        onTap: () => Navigator.pop(context, wh),
                        child: Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                          child: Row(
                            children: [
                              Container(
                                padding: const EdgeInsets.all(8),
                                decoration: BoxDecoration(
                                  color: const Color(0xFF2A6F97).withOpacity(0.1),
                                  borderRadius: BorderRadius.circular(8),
                                ),
                                child: const Icon(Icons.warehouse, color: Color(0xFF2A6F97), size: 20),
                              ),
                              const SizedBox(width: 12),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      wh.code,
                                      style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: Color(0xFF2A6F97)),
                                    ),
                                    const SizedBox(height: 2),
                                    Text(
                                      whName,
                                      style: TextStyle(fontSize: 13, color: Colors.grey[600]),
                                    ),
                                  ],
                                ),
                              ),
                              const Icon(Icons.chevron_right, color: Colors.grey),
                            ],
                          ),
                        ),
                      ),
                    );
                  },
                ),
              ),
              // Footer
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  border: Border(top: BorderSide(color: Colors.grey[300]!)),
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    TextButton(
                      onPressed: () => Navigator.pop(context, WarehouseModel(guidfixed: '', code: '')),
                      style: TextButton.styleFrom(
                        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
                      ),
                      child: Text(global.language("close"), style: TextStyle(color: Color(0xFF2A6F97))),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      );
    },
  );
}

/// แสดง Dialog สำหรับเลือกคลังเริ่มต้น
Future<WarehouseModel?> showWareHouseDefualtDialog(
  BuildContext context,
  List<WarehouseModel> warehouseList,
) async {
  return showDialog<WarehouseModel?>(
    context: context,
    barrierDismissible: true,
    builder: (BuildContext context) {
      return Dialog(
        shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
        child: Container(
          width: (global.isMobileScreen(context)) ? 350 : 420,
          constraints: BoxConstraints(maxHeight: 500),
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // Header
              Container(
                padding: const EdgeInsets.all(16),
                decoration: const BoxDecoration(
                  color: Color(0xFF2A6F97),
                  borderRadius: BorderRadius.only(
                    topLeft: Radius.circular(16),
                    topRight: Radius.circular(16),
                  ),
                ),
                child: Row(
                  children: [
                    const Icon(Icons.warehouse_outlined, color: Colors.white, size: 24),
                    SizedBox(width: 12),
                    Expanded(
                      child: Text(
                        global.language("select_warehouse"),
                        style: const TextStyle(color: Colors.white, fontSize: 16, fontWeight: FontWeight.bold),
                      ),
                    ),
                  ],
                ),
              ),
              // Content
              Flexible(
                child: ListView.separated(
                  shrinkWrap: true,
                  padding: const EdgeInsets.symmetric(vertical: 8),
                  itemCount: warehouseList.length,
                  separatorBuilder: (context, index) => const Divider(height: 1),
                  itemBuilder: (BuildContext context, int index) {
                    final wh = warehouseList[index];
                    final whName = wh.names.firstWhere((ele) => ele.code == global.userLanguage, orElse: () => wh.names.first).name;
                    return Material(
                      color: Colors.transparent,
                      child: InkWell(
                        onTap: () => Navigator.pop(context, wh),
                        child: Padding(
                          padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                          child: Row(
                            children: [
                              Container(
                                padding: const EdgeInsets.all(8),
                                decoration: BoxDecoration(
                                  color: const Color(0xFF2A6F97).withOpacity(0.1),
                                  borderRadius: BorderRadius.circular(8),
                                ),
                                child: const Icon(Icons.warehouse, color: Color(0xFF2A6F97), size: 20),
                              ),
                              const SizedBox(width: 12),
                              Expanded(
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      wh.code,
                                      style: const TextStyle(fontWeight: FontWeight.bold, fontSize: 14, color: Color(0xFF2A6F97)),
                                    ),
                                    const SizedBox(height: 2),
                                    Text(
                                      whName,
                                      style: TextStyle(fontSize: 13, color: Colors.grey[600]),
                                    ),
                                  ],
                                ),
                              ),
                              const Icon(Icons.chevron_right, color: Colors.grey),
                            ],
                          ),
                        ),
                      ),
                    );
                  },
                ),
              ),
              // Footer
              Container(
                padding: const EdgeInsets.all(12),
                decoration: BoxDecoration(
                  border: Border(top: BorderSide(color: Colors.grey[300]!)),
                ),
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.end,
                  children: [
                    TextButton(
                      onPressed: () => Navigator.pop(context, WarehouseModel(guidfixed: '', code: '')),
                      style: TextButton.styleFrom(
                        padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 10),
                      ),
                      child: Text(global.language("close"), style: TextStyle(color: Color(0xFF2A6F97))),
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      );
    },
  );
}

/// แสดง Dialog สำหรับเลือกหน่วยนับ
Future<ProductBarcodeModel?> showUnitsDialog(
  BuildContext context,
  String itemguid,
) async {
  context.read<ProductBarcodeBloc>().add(ProductBarcodeGet(guid: itemguid));
  late List<ProductBarcodeModel> productList = [];
  return showDialog<ProductBarcodeModel?>(
    context: context,
    barrierDismissible: true, // Allows tapping outside the dialog to close it
    builder: (BuildContext context) {
      return BlocBuilder<ProductBarcodeBloc, ProductBarcodeState>(
        builder: (context, state) {
          if (state is ProductBarcodeGetSuccess) {
            if (state.productBarcode.refbarcodes!.isNotEmpty) {
              for (var item in state.productBarcode.refbarcodes!) {
                productList.add(
                  ProductBarcodeModel(
                    barcode: item.barcode,
                    guidfixed: item.guidfixed!,
                    names: item.names,
                    itemunitnames: item.itemunitnames,
                    itemunitcode: item.itemunitcode,
                    dividevalue: item.dividevalue,
                    standvalue: item.standvalue,
                  ),
                );
              }
              context.read<ProductBarcodeBloc>().add(
                ProductBarcodeGetRef(
                  guid: state.productBarcode.refbarcodes![0].barcode,
                ),
              );
            } else {
              productList.add(state.productBarcode);
              context.read<ProductBarcodeBloc>().add(
                ProductBarcodeGetRef(guid: state.productBarcode.barcode!),
              );
            }
          }
          if (state is ProductBarcodeGetRefSuccess) {
            productList.addAll(state.productBarcodes);
          }
          return AlertDialog(
            title: Text(global.language("select_unit")),
            content: (state is ProductBarcodeGetRefSuccess)
                ? SizedBox(
                    width: (global.isMobileScreen(context)) ? 350 : 400,
                    child: ListView.builder(
                      shrinkWrap:
                          true, // Required for the ListView to be displayed properly
                      itemCount: productList.length,
                      itemBuilder: (BuildContext context, int index) {
                        return ListTile(
                          title: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                '${global.language("barcode")}: ${productList[index].barcode}',
                              ),
                              Text(
                                '${global.language("product_name")}: ${global.packName(productList[index].names!)}',
                              ),
                              Text(
                                '${global.language("product_unit_name")}: ${global.packName(productList[index].itemunitnames!)}',
                              ),
                            ],
                          ),
                          onTap: () {
                            Navigator.pop(context, productList[index]);
                          },
                        );
                      },
                    ),
                  )
                : Text(global.language('processing')),
            actions: <Widget>[
              TextButton(
                child: Text(global.language("close")),
                onPressed: () {
                  Navigator.pop(
                    context,
                    ProductBarcodeModel(barcode: '', guidfixed: ''),
                  );
                },
              ),
            ],
          );
        },
      );
    },
  );
}

/// แสดง Dialog ยืนยันการออกใบกำกับภาษีเต็มรูปแบบ
Future<bool?> showAlertConfirmFullInvoiceDialog(
  BuildContext context,
  String docno,
) async {
  return showDialog<bool?>(
    context: context,
    builder: (BuildContext context) {
      // Return the AlertDialog.
      return AlertDialog(
        title: Text(global.language("alert")),
        content: Text('${global.language("confirm_full_tax_invoice")} $docno?'),
        actions: [
          TextButton(
            child: Text(global.language("cancel")),
            onPressed: () {
              // Do something here.
              Navigator.pop(context, false);
            },
          ),
          TextButton(
            child: Text(global.language("confirm")),
            onPressed: () {
              Navigator.pop(context, true);
            },
          ),
        ],
      );
    },
  );
}

/// แสดง Dialog ยืนยันการลบ
Future<bool?> showAlertConfirmDeleteDialog(
  BuildContext context,
  String docno,
) async {
  return showDialog<bool?>(
    context: context,
    builder: (BuildContext context) {
      // Return the AlertDialog.
      return AlertDialog(
        title: Text(global.language("alert")),
        content: Text(global.language("confirm_delete")),
        actions: [
          TextButton(
            child: Text(global.language("cancel")),
            onPressed: () {
              // Do something here.
              Navigator.pop(context, false);
            },
          ),
          TextButton(
            child: Text(global.language("delete")),
            onPressed: () {
              Navigator.pop(context, true);
            },
          ),
        ],
      );
    },
  );
}

/// Model สำหรับผลลัพธ์จาก Dialog ยืนยันการบันทึก
class SaveConfirmResult {
  final bool confirmed;
  final bool printAfterSave;

  SaveConfirmResult({required this.confirmed, required this.printAfterSave});
}

/// แสดง Dialog ยืนยันการบันทึก (แบบเดิม - สำหรับ backward compatibility)
Future<bool?> showAlertConfirmSaveDialog(
  BuildContext context,
  String guid,
) async {
  return showDialog<bool?>(
    context: context,
    builder: (BuildContext context) {
      // Return the AlertDialog.
      return AlertDialog(
        title: Text(global.language("alert")),
        content: (guid == '')
            ? Text(global.language("confirm_save"))
            : Text(global.language("confirm_update")),
        actions: [
          TextButton(
            child: Text(global.language("cancel")),
            onPressed: () {
              // Do something here.
              Navigator.pop(context, false);
            },
          ),
          TextButton(
            child: Text(global.language("save")),
            onPressed: () {
              Navigator.pop(context, true);
            },
          ),
        ],
      );
    },
  );
}

/// Key สำหรับ SharedPreferences
const String _printAfterSaveKey = 'print_after_save_preference';

/// แสดง Dialog ยืนยันการบันทึกแบบสวยงาม พร้อม checkbox เลือกพิมพ์
Future<SaveConfirmResult?> showEnhancedSaveConfirmDialog(
  BuildContext context,
  String guid, {
  bool showPrintOption = true,
}) async {
  // ดึงค่า preference จาก SharedPreferences
  bool printAfterSave = global.prefs.getBool(_printAfterSaveKey) ?? true;

  return showDialog<SaveConfirmResult?>(
    context: context,
    barrierDismissible: true,
    builder: (BuildContext context) {
      return StatefulBuilder(
        builder: (context, setState) {
          final bool isNewDocument = guid.isEmpty;
          final String title = isNewDocument
              ? global.language("confirm_save")
              : global.language("confirm_update");
          final IconData icon = isNewDocument ? Icons.save_outlined : Icons.edit_outlined;
          final Color primaryColor = const Color(0xFF2A6F97);

          return Dialog(
            shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
            child: Container(
              width: (global.isMobileScreen(context)) ? 320 : 380,
              padding: EdgeInsets.zero,
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: [
                  // Header
                  Container(
                    padding: const EdgeInsets.all(20),
                    decoration: BoxDecoration(
                      gradient: LinearGradient(
                        colors: [primaryColor, primaryColor.withOpacity(0.8)],
                        begin: Alignment.topLeft,
                        end: Alignment.bottomRight,
                      ),
                      borderRadius: const BorderRadius.only(
                        topLeft: Radius.circular(16),
                        topRight: Radius.circular(16),
                      ),
                    ),
                    child: Row(
                      children: [
                        Container(
                          padding: const EdgeInsets.all(10),
                          decoration: BoxDecoration(
                            color: Colors.white.withOpacity(0.2),
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: Icon(icon, color: Colors.white, size: 28),
                        ),
                        SizedBox(width: 16),
                        Expanded(
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Text(
                                title,
                                style: const TextStyle(
                                  color: Colors.white,
                                  fontSize: 18,
                                  fontWeight: FontWeight.bold,
                                ),
                              ),
                              SizedBox(height: 4),
                              Text(
                                isNewDocument
                                    ? global.language("new_document")
                                    : global.language("edit_document"),
                                style: TextStyle(
                                  color: Colors.white.withOpacity(0.85),
                                  fontSize: 13,
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],
                    ),
                  ),

                  // Content
                  Padding(
                    padding: EdgeInsets.all(20),
                    child: Column(
                      children: [
                        // Message
                        Container(
                          padding: const EdgeInsets.all(16),
                          decoration: BoxDecoration(
                            color: Colors.grey[50],
                            borderRadius: BorderRadius.circular(12),
                            border: Border.all(color: Colors.grey[200]!),
                          ),
                          child: Row(
                            children: [
                              Icon(
                                Icons.info_outline,
                                color: primaryColor,
                                size: 24,
                              ),
                              SizedBox(width: 12),
                              Expanded(
                                child: Text(
                                  isNewDocument
                                      ? global.language("are_you_sure_save")
                                      : global.language("are_you_sure_update"),
                                  style: TextStyle(
                                    fontSize: 14,
                                    color: Colors.grey[700],
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ),

                        // Print Option Checkbox
                        if (showPrintOption) ...[
                          const SizedBox(height: 16),
                          InkWell(
                            onTap: () {
                              setState(() {
                                printAfterSave = !printAfterSave;
                              });
                            },
                            borderRadius: BorderRadius.circular(12),
                            child: Container(
                              padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 10),
                              decoration: BoxDecoration(
                                color: printAfterSave ? primaryColor.withOpacity(0.1) : Colors.grey[100],
                                borderRadius: BorderRadius.circular(12),
                                border: Border.all(
                                  color: printAfterSave ? primaryColor : Colors.grey[300]!,
                                  width: 1.5,
                                ),
                              ),
                              child: Row(
                                children: [
                                  Checkbox(
                                    value: printAfterSave,
                                    onChanged: (value) {
                                      setState(() {
                                        printAfterSave = value ?? false;
                                      });
                                    },
                                    activeColor: primaryColor,
                                    shape: RoundedRectangleBorder(
                                      borderRadius: BorderRadius.circular(4),
                                    ),
                                  ),
                                  const SizedBox(width: 8),
                                  Icon(
                                    Icons.print_outlined,
                                    color: printAfterSave ? primaryColor : Colors.grey[600],
                                    size: 22,
                                  ),
                                  SizedBox(width: 10),
                                  Expanded(
                                    child: Column(
                                      crossAxisAlignment: CrossAxisAlignment.start,
                                      children: [
                                        Text(
                                          global.language("print_after_save"),
                                          style: TextStyle(
                                            fontSize: 14,
                                            fontWeight: FontWeight.w600,
                                            color: printAfterSave ? primaryColor : Colors.grey[700],
                                          ),
                                        ),
                                        SizedBox(height: 2),
                                        Text(
                                          global.language("print_form_description"),
                                          style: TextStyle(
                                            fontSize: 11,
                                            color: Colors.grey[600],
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                ],
                              ),
                            ),
                          ),
                        ],
                      ],
                    ),
                  ),

                  // Footer Buttons
                  Container(
                    padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                    decoration: BoxDecoration(
                      color: Colors.grey[50],
                      border: Border(top: BorderSide(color: Colors.grey[200]!)),
                      borderRadius: const BorderRadius.only(
                        bottomLeft: Radius.circular(16),
                        bottomRight: Radius.circular(16),
                      ),
                    ),
                    child: Row(
                      children: [
                        Expanded(
                          child: OutlinedButton(
                            onPressed: () => Navigator.pop(context, null),
                            style: OutlinedButton.styleFrom(
                              padding: const EdgeInsets.symmetric(vertical: 12),
                              side: BorderSide(color: Colors.grey[400]!),
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(10),
                              ),
                            ),
                            child: Text(
                              global.language("cancel"),
                              style: TextStyle(
                                color: Colors.grey[700],
                                fontWeight: FontWeight.w600,
                              ),
                            ),
                          ),
                        ),
                        const SizedBox(width: 12),
                        Expanded(
                          child: ElevatedButton(
                            onPressed: () async {
                              // Save preference
                              await global.prefs.setBool(_printAfterSaveKey, printAfterSave);
                              if (context.mounted) {
                                Navigator.pop(
                                  context,
                                  SaveConfirmResult(
                                    confirmed: true,
                                    printAfterSave: printAfterSave,
                                  ),
                                );
                              }
                            },
                            style: ElevatedButton.styleFrom(
                              backgroundColor: primaryColor,
                              padding: const EdgeInsets.symmetric(vertical: 12),
                              elevation: 0,
                              shape: RoundedRectangleBorder(
                                borderRadius: BorderRadius.circular(10),
                              ),
                            ),
                            child: Row(
                              mainAxisAlignment: MainAxisAlignment.center,
                              children: [
                                Icon(
                                  isNewDocument ? Icons.save : Icons.check,
                                  color: Colors.white,
                                  size: 18,
                                ),
                                SizedBox(width: 8),
                                Text(
                                  global.language("save"),
                                  style: const TextStyle(
                                    color: Colors.white,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ),
                      ],
                    ),
                  ),
                ],
              ),
            ),
          );
        },
      );
    },
  );
}

/// แสดง Dialog ยืนยันการยกเลิกเอกสาร พร้อมเลือกเหตุผล
Future<String?> showAlertConfirmCancelDialog(BuildContext context) async {
  TextEditingController controller = TextEditingController();
  final List<String> reasons = <String>[
    global.language('wrong_name_address'),
    global.language('wrong_date'),
    global.language('wrong_quantity_value'),
    global.language('wrong_product'),
    global.language('other'),
  ];
  final GlobalKey<FormState> formKey = GlobalKey<FormState>();
  bool showOther = false;

  controller.text = reasons[0];

  String? result = await showDialog<String?>(
    context: context,
    barrierDismissible: true, // User must tap a button to close the dialog.
    builder: (BuildContext dialogContext) {
      return StatefulBuilder(
        builder: (BuildContext context, StateSetter setState) {
          return AlertDialog(
            title: Text(global.language("alert")),
            content: Form(
              key: formKey,
              child: SizedBox(
                width: 400,
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Padding(
                      padding: EdgeInsets.only(bottom: 10),
                      child: DropdownButtonFormField<String>(
                        decoration: InputDecoration(
                          border: OutlineInputBorder(),
                          labelText: global.language("reason"),
                        ),
                        value: reasons[0],
                        onChanged: (String? newValue) {
                          controller.text = newValue!;
                          setState(() {
                            if (newValue == global.language('other')) {
                              showOther = true;
                              controller.text = "";
                            } else {
                              showOther = false;
                            }
                          });
                        },
                        items: reasons.map<DropdownMenuItem<String>>((
                          String value,
                        ) {
                          return DropdownMenuItem<String>(
                            value: value,
                            child: Text(value),
                          );
                        }).toList(),
                        validator: (value) {
                          if (value == null || value.isEmpty) {
                            return 'Please enter some text';
                          }
                          return null;
                        },
                      ),
                    ),
                    if (showOther)
                      TextFormField(
                        controller: controller,
                        maxLines: 4,
                        decoration: InputDecoration(
                          floatingLabelBehavior:
                              FloatingLabelBehavior.always,
                          border: OutlineInputBorder(),
                          labelText: global.language('description'),
                        ),
                        validator: (value) {
                          if (value == null || value.isEmpty) {
                            return 'Please enter some text';
                          }
                          return null;
                        },
                      ),
                  ],
                ),
              ),
            ),
            actions: [
              TextButton(
                child: Text(global.language("cancel")),
                onPressed: () {
                  Navigator.pop(context);
                },
              ),
              TextButton(
                style: TextButton.styleFrom(foregroundColor: Colors.red),
                child: Text(global.language("save")),
                onPressed: () {
                  if (formKey.currentState!.validate()) {
                    Navigator.pop(
                      context,
                      controller.text,
                    ); // Return the text from the controller
                  }
                },
              ),
            ],
          );
        },
      );
    },
  );
  return result;
}
