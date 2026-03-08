// ignore_for_file: depend_on_referenced_packages

import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;

/// Helper class สำหรับ Mobile Product List Widgets
/// ใช้ class นี้เพื่อแยก mobile product list widget logic ออกจาก main screen
/// รองรับทั้ง TransactionEditScreenState และ PurchaseOrderEditScreenState
class MobileProductListWidgets {
  final dynamic state;

  MobileProductListWidgets(this.state);

  /// Widget สำหรับแสดงรายการสินค้าแบบ Receive (รับสินค้า) บน Mobile
  Widget buildReceiveDetailMobileWidget() {
    List<Widget> dataWidgets = [];
    for (var index = 0; index < state.screenData.details!.length; index++) {
      dataWidgets.add(
        Card(
          color: Colors.blue.shade100,
          child: Padding(
            padding: const EdgeInsets.all(8.0),
            child: Stack(
              children: <Widget>[
                Column(
                  children: <Widget>[
                    // Row แรก: Barcode และปุ่มลบ
                    _buildBarcodeRow(index),
                    // Row ที่สอง: Unit, Warehouse, Location
                    _buildUnitWarehouseLocationRow(index),
                    // Row ที่สาม: Qty และ Price
                    _buildQtyPriceRow(index),
                    // Row สุดท้าย: Total Value
                    _buildTotalValueRow(index),
                  ],
                ),
              ],
            ),
          ),
        ),
      );
    }

    return Column(children: dataWidgets);
  }

  /// Widget สำหรับแสดงรายการสินค้าแบบ PickUp (หยิบสินค้า) บน Mobile
  Widget buildPickUpDetailMobileWidget() {
    List<Widget> dataWidgets = [];
    for (var index = 0; index < state.screenData.details!.length; index++) {
      dataWidgets.add(
        Card(
          color: Colors.blue.shade100,
          child: Padding(
            padding: const EdgeInsets.all(8.0),
            child: Stack(
              children: <Widget>[
                Column(
                  children: <Widget>[
                    // Row แรก: Barcode และปุ่มลบ
                    _buildBarcodeRow(index),
                    // Row ที่สอง: Unit, Warehouse, Location
                    _buildUnitWarehouseLocationRow(index),
                    // Row ที่สาม: Qty เท่านั้น
                    _buildQtyOnlyRow(index),
                  ],
                ),
              ],
            ),
          ),
        ),
      );
    }

    return Column(children: dataWidgets);
  }

  /// Widget สำหรับแสดงรายการสินค้าทั่วไปบน Mobile
  Widget buildDetailMobileWidget() {
    List<Widget> dataWidgets = [];
    for (var index = 0; index < state.screenData.details!.length; index++) {
      dataWidgets.add(
        Card(
          color: Colors.blue.shade100,
          child: Padding(
            padding: const EdgeInsets.all(8.0),
            child: Stack(
              children: <Widget>[
                Column(
                  children: <Widget>[
                    // Row แรก: Barcode และปุ่มลบ
                    _buildBarcodeRow(index),
                    // Row ที่สอง: Unit, Warehouse, Location
                    _buildUnitWarehouseLocationRow(index),
                    // Row ที่สาม: Qty เท่านั้น
                    _buildQtyOnlyRow(index),
                  ],
                ),
              ],
            ),
          ),
        ),
      );
    }

    return Column(children: dataWidgets);
  }

  /// Widget สำหรับแสดงรายการสินค้าแบบ Transfer (โอนสินค้า) บน Mobile
  Widget buildTransferDetailMobileWidget() {
    List<Widget> dataWidgets = [];
    for (var index = 0; index < state.screenData.details!.length; index++) {
      dataWidgets.add(
        Card(
          color: Colors.blue.shade100,
          child: Padding(
            padding: const EdgeInsets.all(8.0),
            child: Stack(
              children: <Widget>[
                Column(
                  children: <Widget>[
                    // Row แรก: Barcode และปุ่มลบ
                    _buildBarcodeRow(index),
                    // Row ที่สอง: Unit, Warehouse, Location (ต้นทาง)
                    _buildUnitWarehouseLocationRow(index),
                    // Row ที่สาม: To Warehouse, To Location, Qty
                    _buildTransferDestinationRow(index),
                  ],
                ),
              ],
            ),
          ),
        ),
      );
    }

    return Column(children: dataWidgets);
  }

  // ===== Private Helper Methods =====

  /// สร้าง Row สำหรับแสดง Barcode และปุ่มลบ
  Widget _buildBarcodeRow(int index) {
    return Container(
      margin: const EdgeInsets.only(bottom: 5),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.spaceBetween,
        children: [
          ElevatedButton(
            style: _buttonStyle(),
            onPressed: () {
              state.showDialogCommand(
                'barcode',
                index,
                state.screenData.details![index],
              );
            },
            child: Text(
              '${state.screenData.details![index].barcode}~${global.activeLangName(state.screenData.details![index].itemnames!)}',
              style: const TextStyle(fontSize: 12),
            ),
          ),
          IconButton(
            onPressed: () {
              state.deleteItemDetail(index);
            },
            icon: const Icon(Icons.delete, color: Colors.red),
          ),
        ],
      ),
    );
  }

  /// สร้าง Row สำหรับแสดง Unit, Warehouse, Location
  Widget _buildUnitWarehouseLocationRow(int index) {
    return Container(
      margin: EdgeInsets.only(bottom: 5),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.start,
        children: [
          Text(global.language("product_unit")),
          ElevatedButton(
            style: _buttonStyle(),
            onPressed: () {
              state.showDialogCommand(
                'product_unit',
                index,
                state.screenData.details![index],
              );
            },
            child: Text(
              global.activeLangName(
                state.screenData.details![index].unitnames!,
              ),
              style: const TextStyle(fontSize: 12),
            ),
          ),
          SizedBox(width: 5),
          Text(global.language("warehouse")),
          ElevatedButton(
            style: _buttonStyle(),
            onPressed: () {
              state.showDialogCommand(
                'product_ware_house',
                index,
                state.screenData.details![index],
              );
            },
            child: Text(
              global.activeLangName(
                state.screenData.details![index].whnames!,
              ),
              style: const TextStyle(fontSize: 12),
            ),
          ),
          SizedBox(width: 5),
          Text(global.language("location")),
          ElevatedButton(
            style: _buttonStyle(),
            onPressed: () {
              state.showDialogCommand(
                'product_location',
                index,
                state.screenData.details![index],
              );
            },
            child: Text(
              global.activeLangName(
                state.screenData.details![index].locationnames!,
              ),
              style: const TextStyle(fontSize: 12),
            ),
          ),
        ],
      ),
    );
  }

  /// สร้าง Row สำหรับแสดง Qty และ Price
  Widget _buildQtyPriceRow(int index) {
    return Container(
      margin: EdgeInsets.only(bottom: 5),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.start,
        children: [
          Text(global.language("product_qty")),
          ElevatedButton(
            style: _buttonStyle(),
            onPressed: () {
              state.showDialogCommand(
                'product_qty',
                index,
                state.screenData.details![index],
              );
            },
            child: Text(
              global.formatNumber(
                state.screenData.details![index].qty,
              ),
              style: const TextStyle(fontSize: 12),
            ),
          ),
          SizedBox(width: 5),
          Text(global.language("product_cost")),
          ElevatedButton(
            style: _buttonStyle(),
            onPressed: () {
              if (state.screenData.transflag == 966) {
                state.showDialogCommand(
                  'product_price_adjust',
                  index,
                  state.screenData.details![index],
                );
              } else {
                state.showDialogCommand(
                  'product_price',
                  index,
                  state.screenData.details![index],
                );
              }
            },
            child: Text(
              global.formatNumber(
                state.screenData.details![index].price,
              ),
              style: const TextStyle(fontSize: 12),
            ),
          ),
        ],
      ),
    );
  }

  /// สร้าง Row สำหรับแสดง Qty เท่านั้น
  Widget _buildQtyOnlyRow(int index) {
    return Container(
      margin: EdgeInsets.only(bottom: 5),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.start,
        children: [
          Text(global.language("product_qty")),
          ElevatedButton(
            style: _buttonStyle(),
            onPressed: () {
              state.showDialogCommand(
                'product_qty',
                index,
                state.screenData.details![index],
              );
            },
            child: Text(
              global.formatNumber(
                state.screenData.details![index].qty,
              ),
              style: const TextStyle(fontSize: 12),
            ),
          ),
        ],
      ),
    );
  }

  /// สร้าง Row สำหรับแสดง Total Value
  Widget _buildTotalValueRow(int index) {
    return Row(
      mainAxisAlignment: MainAxisAlignment.end,
      children: [
        Text(
          '${global.language("total_value")}: ${global.formatNumber(state.screenData.details![index].sumamount)}',
          style: const TextStyle(fontWeight: FontWeight.bold),
        ),
      ],
    );
  }

  /// สร้าง Row สำหรับแสดง Transfer Destination (To Warehouse, To Location, Qty)
  Widget _buildTransferDestinationRow(int index) {
    return Container(
      margin: EdgeInsets.only(bottom: 5),
      child: Row(
        mainAxisAlignment: MainAxisAlignment.start,
        children: [
          Text(global.language("to_warehouse")),
          ElevatedButton(
            style: _buttonStyle(),
            onPressed: () {
              state.showDialogCommand(
                'product_to_ware_house',
                index,
                state.screenData.details![index],
              );
            },
            child: Text(
              global.activeLangName(
                state.screenData.details![index].towhnames!,
              ),
              style: const TextStyle(fontSize: 12),
            ),
          ),
          SizedBox(width: 5),
          Text(global.language("location")),
          ElevatedButton(
            style: _buttonStyle(),
            onPressed: () {
              state.showDialogCommand(
                'product_to_location',
                index,
                state.screenData.details![index],
              );
            },
            child: Text(
              global.activeLangName(
                state.screenData.details![index].tolocationnames!,
              ),
              style: const TextStyle(fontSize: 12),
            ),
          ),
          SizedBox(width: 5),
          Text(global.language("product_qty")),
          ElevatedButton(
            style: _buttonStyle(),
            onPressed: () {
              state.showDialogCommand(
                'product_qty',
                index,
                state.screenData.details![index],
              );
            },
            child: Text(
              global.formatNumber(
                state.screenData.details![index].qty,
              ),
              style: const TextStyle(fontSize: 12),
            ),
          ),
        ],
      ),
    );
  }

  /// สร้าง ButtonStyle สำหรับปุ่มทั่วไป
  ButtonStyle _buttonStyle() {
    return ElevatedButton.styleFrom(
      padding: const EdgeInsets.all(4),
      foregroundColor: Colors.black,
      backgroundColor: Colors.white,
      shape: const RoundedRectangleBorder(
        borderRadius: BorderRadius.all(
          Radius.circular(0),
        ),
      ),
    );
  }
}
