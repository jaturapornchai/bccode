import 'package:flutter/material.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/location_model.dart';
import 'package:smlaicloud/model/price_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/model/warehouse_model.dart';
import 'package:smlaicloud/screens/quotation/utils/qt_product_search_handler.dart';
import 'package:smlaicloud/components/dual_currency_price_dialog.dart';
import 'package:smlaicloud/screens/transaction/utils/transaction_dialogs.dart';
import 'package:smlaicloud/global.dart' as global;

/// Handler สำหรับจัดการคำสั่งต่างๆ ในตารางรายการสินค้าของใบเสนอราคา
///
/// รับผิดชอบ:
/// - ค้นหาสินค้าด้วย Barcode/ItemCode
/// - แก้ไขชื่อสินค้า/หมายเหตุ
/// - เลือกคลัง/ที่เก็บสินค้า
/// - เปลี่ยนหน่วยสินค้า
/// - แก้ไขจำนวน/ราคา/ส่วนลด
/// - จัดการตัวเลือกเพิ่มเติม
class QTDetailCommandHandler {
  final BuildContext context;
  final TransactionModel screenData;
  final List<WarehouseModel> warehouseList;
  final QTWarehouseInfo warehouseInfo;
  final double Function(List<PriceDataModel>?) getPrice;
  final VoidCallback onStateChanged;
  final VoidCallback onRecalculate;
  final int calcflag;
  final String defaultToWarehouse;
  final List<LanguageDataModel> defaultToWarehouseNames;
  final String defaultToLocation;
  final List<LanguageDataModel> defaultToLocationNames;

  QTDetailCommandHandler({
    required this.context,
    required this.screenData,
    required this.warehouseList,
    required this.warehouseInfo,
    required this.getPrice,
    required this.onStateChanged,
    required this.onRecalculate,
    required this.calcflag,
    required this.defaultToWarehouse,
    required this.defaultToWarehouseNames,
    required this.defaultToLocation,
    required this.defaultToLocationNames,
  });

  /// จัดการคำสั่งที่ส่งมาจากตารางรายการสินค้า
  Future<void> handleCommand(String cmd, int index, TransactionDetailModel details) async {
    switch (cmd) {
      case 'barcode':
        await _handleBarcodeSearch(index, details);
        break;
      case 'item_code':
        await _handleItemCodeSearch(index, details);
        break;
      case 'product_name':
        await _handleEditProductName(index, details);
        break;
      case 'product_ware_house':
      case 'warehouse':
        await _handleWarehouseSelection(index, details);
        break;
      case 'product_location':
      case 'location':
        await _handleLocationSelection(index, details);
        break;
      case 'product_to_ware_house':
        await _handleToWarehouseSelection(index, details);
        break;
      case 'product_to_location':
        await _handleToLocationSelection(index, details);
        break;
      case 'product_unit':
        await _handleUnitSelection(index, details);
        break;
      case 'product_qty':
      case 'qty':
        await _handleQtyEdit(index, details);
        break;
      case 'product_price':
      case 'product_price_adjust':
      case 'price':
        await _handlePriceEdit(index, details);
        break;
      case 'product_discount':
        await _handleDiscountEdit(index, details);
        break;
      case 'product_amount':
        await _handleAmountEdit(index, details);
        break;
      case 'description':
        await _handleDescriptionEdit(index, details);
        break;
      case 'options':
        await _handleOptionsView(index, details);
        break;
    }
  }

  Future<void> _handleBarcodeSearch(int index, TransactionDetailModel details) async {
    final result = await QTProductSearchHandler.searchByBarcode(
      context,
      initialWord: details.barcode,
    );
    if (result != null) {
      QTProductSearchHandler.updateDetailFromProduct(
        detail: screenData.details![index],
        product: result,
        warehouseInfo: warehouseInfo,
        getPrice: getPrice,
        exchangeRate: _isMultiCurrency() ? screenData.exchangerate : null,
      );
      onStateChanged();
      onRecalculate();
    }
  }

  Future<void> _handleItemCodeSearch(int index, TransactionDetailModel details) async {
    final result = await QTProductSearchHandler.searchByItemCode(
      context,
      initialWord: details.itemcode,
    );
    if (result != null) {
      QTProductSearchHandler.updateDetailFromProduct(
        detail: screenData.details![index],
        product: result,
        warehouseInfo: warehouseInfo,
        getPrice: getPrice,
        exchangeRate: _isMultiCurrency() ? screenData.exchangerate : null,
      );
      onStateChanged();
      onRecalculate();
    }
  }

  Future<void> _handleEditProductName(int index, TransactionDetailModel details) async {
    String currentProductName = global.activeLangName(details.itemnames!);
    String currentDescription = details.description ?? "";

    TextEditingController productNameController = TextEditingController(text: currentProductName);
    TextEditingController descriptionController = TextEditingController(text: currentDescription);

    await showDialog(
      context: context,
      builder: (BuildContext dialogContext) {
        return Dialog(
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
          child: SizedBox(
            width: (global.isMobileScreen(context)) ? 350 : 420,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                Container(
                  padding: const EdgeInsets.all(16),
                  decoration: BoxDecoration(
                    color: global.theme.primaryColor,
                    borderRadius: const BorderRadius.only(topLeft: Radius.circular(16), topRight: Radius.circular(16)),
                  ),
                  child: Row(
                    children: [
                      Icon(Icons.edit_note, color: global.theme.onPrimaryColor, size: 24),
                      SizedBox(width: 12),
                      Expanded(
                        child: Text(
                          global.language("edit_product_info"),
                          style: TextStyle(color: global.theme.onPrimaryColor, fontSize: 16, fontWeight: FontWeight.bold),
                        ),
                      ),
                    ],
                  ),
                ),
                Padding(
                  padding: EdgeInsets.all(16),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      _buildFieldHeader(Icons.inventory_2_outlined, global.language("product_name")),
                      SizedBox(height: 8),
                      _buildTextField(productNameController, global.language("enter_product_name"), maxLines: 2),
                      SizedBox(height: 16),
                      _buildFieldHeader(Icons.description_outlined, global.language("description")),
                      SizedBox(height: 8),
                      _buildTextField(descriptionController, global.language("enter_description"), maxLines: 3),
                    ],
                  ),
                ),
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    border: Border(top: BorderSide(color: global.theme.dividerBorderColor)),
                  ),
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.end,
                    children: [
                      TextButton(
                        onPressed: () => Navigator.of(dialogContext).pop(),
                        style: TextButton.styleFrom(padding: EdgeInsets.symmetric(horizontal: 20, vertical: 10)),
                        child: Text(global.language("cancel"), style: TextStyle(color: global.theme.iconSecondaryColor)),
                      ),
                      SizedBox(width: 8),
                      ElevatedButton(
                        onPressed: () {
                          for (int i = 0; i < screenData.details![index].itemnames!.length; i++) {
                            if (screenData.details![index].itemnames![i].code == global.userLanguage) {
                              screenData.details![index].itemnames![i].name = productNameController.text;
                              break;
                            }
                          }
                          screenData.details![index].description = descriptionController.text;
                          Navigator.of(dialogContext).pop();
                          onStateChanged();
                        },
                        style: ElevatedButton.styleFrom(
                          backgroundColor: global.theme.primaryColor,
                          foregroundColor: global.theme.onPrimaryColor,
                          padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 10),
                          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
                        ),
                        child: Text(global.language("confirm")),
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

    productNameController.dispose();
    descriptionController.dispose();
  }

  Widget _buildFieldHeader(IconData icon, String label) {
    return Row(
      children: [
        Container(
          padding: const EdgeInsets.all(6),
          decoration: BoxDecoration(color: global.theme.primaryColor.withValues(alpha: 0.1), borderRadius: BorderRadius.circular(6)),
          child: Icon(icon, color: global.theme.primaryColor, size: 18),
        ),
        const SizedBox(width: 8),
        Text(
          label,
          style: TextStyle(fontSize: 14, fontWeight: FontWeight.w600, color: global.theme.primaryColor),
        ),
      ],
    );
  }

  Widget _buildTextField(TextEditingController controller, String hintText, {int maxLines = 1}) {
    return TextField(
      controller: controller,
      decoration: InputDecoration(
        hintText: hintText,
        hintStyle: TextStyle(color: global.theme.formHintColor),
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(10)),
        focusedBorder: OutlineInputBorder(
          borderRadius: BorderRadius.circular(10),
          borderSide: BorderSide(color: global.theme.primaryColor, width: 2),
        ),
        contentPadding: const EdgeInsets.symmetric(horizontal: 12, vertical: 12),
      ),
      maxLines: maxLines,
    );
  }

  Future<void> _handleWarehouseSelection(int index, TransactionDetailModel details) async {
    if (details.itemguid.trim().isEmpty) return;

    WarehouseModel? result = await showWareHouseDialog(context, details.itemguid, warehouseList) ?? WarehouseModel(guidfixed: '', code: '');
    if (result.code.isNotEmpty) {
      screenData.details![index].whcode = result.code;
      screenData.details![index].whnames = result.names;

      if (result.location.isNotEmpty) {
        screenData.details![index].locationcode = result.location[0].code;
        screenData.details![index].locationnames = result.location[0].names;
      } else {
        screenData.details![index].locationcode = "";
        screenData.details![index].locationnames = [];
      }
      onStateChanged();
    }
  }

  Future<void> _handleLocationSelection(int index, TransactionDetailModel details) async {
    if (details.itemguid.trim().isEmpty || details.whcode.trim().isEmpty) return;

    LocationModel? result = await showWareHouseLocationDialog(context, details.itemguid, details.whcode) ?? LocationModel(code: '');
    if (result.code.isNotEmpty) {
      screenData.details![index].locationcode = result.code;
      screenData.details![index].locationnames = result.names;
      onStateChanged();
    }
  }

  Future<void> _handleToWarehouseSelection(int index, TransactionDetailModel details) async {
    if (details.itemguid.trim().isEmpty) return;

    WarehouseModel? result = await showWareHouseDialog(context, details.itemguid, warehouseList) ?? WarehouseModel(guidfixed: '', code: '');
    if (result.code.isNotEmpty) {
      screenData.details![index].towhcode = result.code;
      screenData.details![index].towhnames = result.names;

      if (result.location.isNotEmpty) {
        screenData.details![index].tolocationcode = result.location[0].code;
        screenData.details![index].tolocationnames = result.location[0].names;
      } else {
        screenData.details![index].tolocationcode = "";
        screenData.details![index].tolocationnames = [];
      }
      onStateChanged();
    }
  }

  Future<void> _handleToLocationSelection(int index, TransactionDetailModel details) async {
    if (details.itemguid.trim().isEmpty || (details.towhcode?.trim().isEmpty ?? true)) return;

    LocationModel? result = await showWareHouseLocationDialog(context, details.itemguid, details.towhcode!) ?? LocationModel(code: '');
    if (result.code.isNotEmpty) {
      screenData.details![index].tolocationcode = result.code;
      screenData.details![index].tolocationnames = result.names;
      onStateChanged();
    }
  }

  Future<void> _handleUnitSelection(int index, TransactionDetailModel details) async {
    if (details.itemguid.trim().isEmpty || details.multiunit != true) return;

    ProductBarcodeModel? result = await showUnitsDialog(context, details.itemguid) ?? ProductBarcodeModel(guidfixed: '');

    if (result.guidfixed.isNotEmpty) {
      final detail = screenData.details![index];
      detail.docdatetime = DateTime.now().toUtc().toIso8601String();
      detail.itemguid = result.guidfixed;
      detail.barcode = result.barcode!;
      detail.itemcode = result.itemcode ?? "";
      detail.itemnames = result.names;
      detail.unitcode = result.itemunitcode;
      detail.qty = 1;
      detail.price = getPrice(result.prices);
      detail.discount = '';
      detail.sumofcost = 0;
      detail.sumamount = 0;
      detail.remark = '';
      detail.linenumber = 0;
      detail.shelfcode = '';
      detail.totalvaluevat = 0;
      detail.totalqty = 0;
      detail.standvalue = result.standvalue!;
      detail.dividevalue = result.dividevalue!;
      detail.multiunit = true;
      detail.unitnames = result.itemunitnames;
      detail.calcflag = calcflag;
      detail.vattype = 0;
      detail.averagecost = 0;
      detail.sumamountexcludevat = 0;
      detail.discountamount = 0;
      detail.ispos = 0;
      detail.laststatus = 0;
      detail.itemtype = 0;
      detail.inquirytype = 0;
      detail.priceexcludevat = 0;
      detail.taxtype = result.taxtype!;
      detail.vatcal = result.vatcal;
      detail.towhcode = defaultToWarehouse;
      detail.towhnames = defaultToWarehouseNames;
      detail.tolocationcode = defaultToLocation;
      detail.tolocationnames = defaultToLocationNames;

      if (_isMultiCurrency()) {
        final rate = screenData.exchangerate ?? 1.0;
        detail.priceDoc = (rate > 0) ? detail.price / rate : detail.price;
      } else {
        detail.priceDoc = null;
      }

      onRecalculate();
      onStateChanged();
    }
  }

  Future<void> _handleQtyEdit(int index, TransactionDetailModel details) async {
    String? result = await showQtyDialog(context, details.qty, global.activeLangName(details.itemnames!)) ?? screenData.details![index].qty.toString();
    final numbercheck = result.replaceAll(',', '');
    screenData.details![index].qty = double.parse(numbercheck);
    onRecalculate();
    onStateChanged();
  }

  bool _isMultiCurrency() {
    return screenData.docCurrency != null &&
        screenData.docCurrency!.isNotEmpty &&
        screenData.currency != null &&
        screenData.currency!.isNotEmpty &&
        screenData.docCurrency != screenData.currency &&
        (screenData.exchangerate ?? 1.0) != 1.0;
  }

  Future<void> _handlePriceEdit(int index, TransactionDetailModel details) async {
    final detail = screenData.details![index];
    final itemName = global.activeLangName(details.itemnames!);

    if (_isMultiCurrency()) {
      final rate = screenData.exchangerate ?? 1.0;
      final currentDocPrice = detail.priceDoc ?? (rate > 0 ? detail.price / rate : detail.price);
      final currentBasePrice = detail.price;

      final result = await showDualCurrencyPriceDialog(
        context,
        title: '$itemName ${global.language("price")}',
        initialDocPrice: currentDocPrice,
        initialBasePrice: currentBasePrice,
        exchangeRate: rate,
        docCurrencyCode: screenData.docCurrency ?? '',
        docCurrencySymbol: screenData.docCurrencySymbol ?? '',
        baseCurrencyCode: screenData.currency ?? '',
        baseCurrencySymbol: screenData.currencysymbol ?? '',
      );

      if (result != null) {
        detail.priceDoc = result.docPrice;
        detail.price = result.basePrice;
        onRecalculate();
        onStateChanged();
      }
    } else {
      String? result = await showPriceDialog(context, details.price, itemName) ?? detail.price.toString();
      final numbercheck = result.replaceAll(',', '');
      detail.price = double.parse(numbercheck);
      detail.priceDoc = null;
      onRecalculate();
      onStateChanged();
    }
  }

  Future<void> _handleDiscountEdit(int index, TransactionDetailModel details) async {
    String? result = await showDiscountDialog(context, details.discount, global.activeLangName(details.itemnames!)) ?? screenData.details![index].discount.toString();
    String numbercheck = result.replaceAll(',', '');
    screenData.details![index].discount = numbercheck;
    onRecalculate();
    onStateChanged();
  }

  Future<void> _handleAmountEdit(int index, TransactionDetailModel details) async {
    String? result = await global.showNumericInputDialog(
      context,
      title: '${global.activeLangName(details.itemnames!)} ${global.language("amount")}',
      initialValue: details.sumamount.toString(),
      decimalPlaces: 2,
    );
    final numbercheck = (result ?? screenData.details![index].sumamount.toString()).replaceAll(',', '');
    screenData.details![index].sumamount = double.parse(numbercheck);
    onStateChanged();
  }

  Future<void> _handleDescriptionEdit(int index, TransactionDetailModel details) async {
    String? result = await showDescriptionDialog(context, details.description ?? '', global.activeLangName(details.itemnames!)) ?? screenData.details![index].description;
    screenData.details![index].description = result;
    onStateChanged();
  }

  Future<void> _handleOptionsView(int index, TransactionDetailModel details) async {
    await showDialog(
      context: context,
      builder: (BuildContext dialogContext) {
        return AlertDialog(
          title: Text('${global.language("additional_options")}: ${global.activeLangName(details.itemnames!)}'),
          content: SizedBox(
            width: (global.isMobileScreen(context)) ? 350 : 400,
            height: 300,
            child: details.extrajsonlist != null && details.extrajsonlist!.isNotEmpty
                ? Column(
                    children: [
                      Text('${global.language("options_found")} ${details.extrajsonlist!.length} ${global.language("items_text")}'),
                      SizedBox(height: 10),
                      Expanded(
                        child: ListView.builder(
                          itemCount: details.extrajsonlist!.length,
                          itemBuilder: (context, optionIndex) {
                            final option = details.extrajsonlist![optionIndex];
                            return Card(
                              child: ListTile(
                                title: Text(global.activeLangName(option.itemnames!)),
                                subtitle: Text('${global.language("price_colon")}: ${global.formatUnitPrice(option.price!)} | ${global.language("amount")}: ${global.formatQuantity(option.qty!)}'),
                                trailing: IconButton(
                                  icon: Icon(Icons.delete, color: global.theme.negativeHighlightTextColor),
                                  onPressed: () {
                                    screenData.details![index].extrajsonlist!.removeAt(optionIndex);
                                    onStateChanged();
                                    Navigator.pop(dialogContext);
                                  },
                                ),
                              ),
                            );
                          },
                        ),
                      ),
                    ],
                  )
                : Center(child: Text(global.language('no_additional_options'))),
          ),
          actions: <Widget>[
            TextButton(
              child: Text(global.language("close")),
              onPressed: () => Navigator.pop(dialogContext),
            ),
          ],
        );
      },
    );
  }
}
