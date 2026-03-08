// ignore_for_file: deprecated_member_use

import 'package:flutter/material.dart';
import 'package:smlaicloud/model/transaction_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/warehouse_model.dart';
import 'package:smlaicloud/model/location_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/price_model.dart';
import 'package:smlaicloud/screens/transaction/components/advance_payment_widget.dart';
import 'package:smlaicloud/screens/transaction/components/bottom_bar_widget.dart';
import 'package:smlaicloud/screens/transaction/components/table_header_widget.dart';
import 'package:smlaicloud/screens/transaction/components/document_total_widget.dart';
import 'package:smlaicloud/screen_search/barcode_search_screen.dart';
import 'package:smlaicloud/global.dart' as global;

class DocumentProductListWidget extends StatelessWidget {
  final TransactionModel screenData;
  final global.TransactionTypeEnum transactionType;
  final List<global.DataTableHeader> headers;
  final List<WarehouseModel> warehouseList;
  final String defaultWarehouse;
  final List<LanguageDataModel> defaultWarehouseNames;
  final String defaultLocation;
  final List<LanguageDataModel> defaultLocationNames;
  final String defaultToWarehouse;
  final List<LanguageDataModel> defaultToWarehouseNames;
  final String defaultToLocation;
  final List<LanguageDataModel> defaultToLocationNames;
  final int calcflag;
  final List<String> cartList;
  final List<dynamic> docrefs;
  final Function(void Function()) setState;
  final BuildContext context;

  // Callback functions
  final Future<WarehouseModel?> Function(BuildContext, List<WarehouseModel>) showWareHouseDefualtDialog;
  final Future<LocationModel?> Function(BuildContext, String) showWareHouseLocationDefualtDialog;
  final Function(String cmd, int index, TransactionDetailModel details) showDialogCommand;
  final Function(int index) deleteItemDetail;
  final Function() calTotalValue;
  final Function(List<PriceDataModel>?) getPrice;
  final Function() showBarcodeDialog;
  final Function(String, List<LanguageDataModel>, String, List<LanguageDataModel>, String, List<LanguageDataModel>, String, List<LanguageDataModel>) onWarehouseChanged;

  const DocumentProductListWidget({
    super.key,
    required this.screenData,
    required this.transactionType,
    required this.headers,
    required this.warehouseList,
    required this.defaultWarehouse,
    required this.defaultWarehouseNames,
    required this.defaultLocation,
    required this.defaultLocationNames,
    required this.defaultToWarehouse,
    required this.defaultToWarehouseNames,
    required this.defaultToLocation,
    required this.defaultToLocationNames,
    required this.calcflag,
    required this.cartList,
    required this.docrefs,
    required this.setState,
    required this.context,
    required this.showWareHouseDefualtDialog,
    required this.showWareHouseLocationDefualtDialog,
    required this.showDialogCommand,
    required this.deleteItemDetail,
    required this.calTotalValue,
    required this.getPrice,
    required this.showBarcodeDialog,
    required this.onWarehouseChanged,
  });

  @override
  Widget build(BuildContext context) {
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.only(top: 5),
      child: LayoutBuilder(
        builder: (context, constraints) {
          return Column(
            children: [
              (transactionType == global.TransactionTypeEnum.paidAdvance ||
                      transactionType == global.TransactionTypeEnum.paidAdvanceRefund ||
                      transactionType == global.TransactionTypeEnum.receiveDeposit ||
                      transactionType == global.TransactionTypeEnum.receiveDepositRefund)
                  ? Container()
                  : _buildWarehouseSelectionWidget(),
              BottomBarWidget(
                transactionType: transactionType,
                screenData: screenData,
                cartList: cartList,
                onBarcodePressed: showBarcodeDialog,
                onProductAdded: _addProductFromBarcode,
                onEmptyProductAdded: _addEmptyProduct,
                setState: setState,
                context: context,
              ),
              TableHeaderWidget(transactionType: transactionType, headers: headers, constraints: constraints),
              Expanded(
                child:
                    (transactionType == global.TransactionTypeEnum.advancePayment ||
                        transactionType == global.TransactionTypeEnum.advancePaymentRefund ||
                        transactionType == global.TransactionTypeEnum.deposit ||
                        transactionType == global.TransactionTypeEnum.depositRefund ||
                        transactionType == global.TransactionTypeEnum.paidAdvance ||
                        transactionType == global.TransactionTypeEnum.paidAdvanceRefund ||
                        transactionType == global.TransactionTypeEnum.receiveDeposit ||
                        transactionType == global.TransactionTypeEnum.receiveDepositRefund)
                    ? AdvancePaymentWidget(screenData: screenData, deleteItemDetail: deleteItemDetail, calTotalValue: calTotalValue, setState: setState)
                    : SingleChildScrollView(child: _buildProductListDetail(constraints.maxWidth)),
              ),
              BottomBarWidget(
                transactionType: transactionType,
                screenData: screenData,
                cartList: cartList,
                onBarcodePressed: showBarcodeDialog,
                onProductAdded: _addProductFromBarcode,
                onEmptyProductAdded: _addEmptyProduct,
                setState: setState,
                context: context,
              ),
            ],
          );
        },
      ),
    );
  }

  Widget _buildWarehouseSelectionWidget() {
    if (transactionType != global.TransactionTypeEnum.stocktransfer) {
      return Container(
        padding: EdgeInsets.symmetric(horizontal: 4, vertical: 4),
        child: Row(
          children: [
            // Warehouse Section
            Expanded(
              flex: 2,
              child: _buildWarehouseSelector(
                label: global.language("warehouse"),
                value: global.activeLangName(defaultWarehouseNames),
                icon: Icons.warehouse,
                onTap: () async {
                  WarehouseModel? result = await showWareHouseDefualtDialog(context, warehouseList) ?? WarehouseModel(guidfixed: '', code: '');
                  if (result.code.isNotEmpty) {
                    String newDefualtwarehouse = result.code;
                    List<LanguageDataModel> newDefualtwarehousenames = result.names;
                    String newDefualtlocation = "";
                    List<LanguageDataModel> newDefualtlocationnames = [];

                    if (result.location.isNotEmpty) {
                      newDefualtlocation = result.location[0].code;
                      newDefualtlocationnames = result.location[0].names;
                    }

                    onWarehouseChanged(
                      newDefualtwarehouse,
                      newDefualtwarehousenames,
                      newDefualtlocation,
                      newDefualtlocationnames,
                      defaultToWarehouse,
                      defaultToWarehouseNames,
                      defaultToLocation,
                      defaultToLocationNames,
                    );
                  }
                },
              ),
            ),
            const SizedBox(width: 5),
            // Location Section
            Expanded(
              flex: 2,
              child: _buildWarehouseSelector(
                label: global.language("location"),
                value: global.activeLangName(defaultLocationNames),
                icon: Icons.location_on,
                onTap: () async {
                  if (defaultWarehouse.isNotEmpty) {
                    LocationModel? result = await showWareHouseLocationDefualtDialog(context, defaultWarehouse) ?? LocationModel(code: '');
                    if (result.code.isNotEmpty) {
                      onWarehouseChanged(defaultWarehouse, defaultWarehouseNames, result.code, result.names, defaultToWarehouse, defaultToWarehouseNames, defaultToLocation, defaultToLocationNames);
                    }
                  }
                },
                enabled: defaultWarehouse.isNotEmpty,
              ),
            ),
          ],
        ),
      );
    } else {
      return Container(
        margin: const EdgeInsets.only(bottom: 4),
        padding: const EdgeInsets.all(8),
        decoration: BoxDecoration(
          color: const Color(0xFF2A6F97).withValues(alpha: 0.1),
          border: Border.all(color: const Color(0xFF2A6F97).withValues(alpha: 0.3)),
          borderRadius: BorderRadius.circular(10),
        ),
        child: Column(
          children: [
            // Compact Transfer Row
            Row(
              children: [
                // FROM Section - Compact
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        global.language("from"),
                        style: const TextStyle(fontSize: 11, color: Color(0xFF2A6F97), fontWeight: FontWeight.w500),
                      ),
                      SizedBox(height: 2),
                      Row(
                        children: [
                          Expanded(
                            child: _buildWarehouseSelector(
                              label: global.language("warehouse"),
                              value: global.activeLangName(defaultWarehouseNames),
                              icon: Icons.warehouse,
                              compact: true,
                              onTap: () async {
                                WarehouseModel? result = await showWareHouseDefualtDialog(context, warehouseList) ?? WarehouseModel(guidfixed: '', code: '');
                                if (result.code.isNotEmpty) {
                                  String newDefualtwarehouse = result.code;
                                  List<LanguageDataModel> newDefualtwarehousenames = result.names;
                                  String newDefualtlocation = "";
                                  List<LanguageDataModel> newDefualtlocationnames = [];

                                  if (result.location.isNotEmpty) {
                                    newDefualtlocation = result.location[0].code;
                                    newDefualtlocationnames = result.location[0].names;
                                  }

                                  onWarehouseChanged(
                                    newDefualtwarehouse,
                                    newDefualtwarehousenames,
                                    newDefualtlocation,
                                    newDefualtlocationnames,
                                    defaultToWarehouse,
                                    defaultToWarehouseNames,
                                    defaultToLocation,
                                    defaultToLocationNames,
                                  );
                                }
                              },
                            ),
                          ),
                          SizedBox(width: 4),
                          Expanded(
                            child: _buildWarehouseSelector(
                              label: global.language("position"),
                              value: global.activeLangName(defaultLocationNames),
                              icon: Icons.location_on,
                              compact: true,
                              onTap: () async {
                                if (defaultWarehouse.isNotEmpty) {
                                  LocationModel? result = await showWareHouseLocationDefualtDialog(context, defaultWarehouse) ?? LocationModel(code: '');
                                  if (result.code.isNotEmpty) {
                                    onWarehouseChanged(
                                      defaultWarehouse,
                                      defaultWarehouseNames,
                                      result.code,
                                      result.names,
                                      defaultToWarehouse,
                                      defaultToWarehouseNames,
                                      defaultToLocation,
                                      defaultToLocationNames,
                                    );
                                  }
                                }
                              },
                              enabled: defaultWarehouse.isNotEmpty,
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),

                // Compact Arrow
                Padding(
                  padding: const EdgeInsets.symmetric(horizontal: 4),
                  child: Icon(Icons.arrow_forward, color: const Color(0xFF2A6F97), size: 14),
                ),

                // TO Section - Compact
                Expanded(
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        global.language("to"),
                        style: TextStyle(fontSize: 11, color: const Color(0xFF2A6F97).withValues(alpha: 0.8), fontWeight: FontWeight.w500),
                      ),
                      SizedBox(height: 2),
                      Row(
                        children: [
                          Expanded(
                            child: _buildWarehouseSelector(
                              label: global.language("warehouse"),
                              value: global.activeLangName(defaultToWarehouseNames),
                              icon: Icons.warehouse,
                              compact: true,
                              onTap: () async {
                                WarehouseModel? result = await showWareHouseDefualtDialog(context, warehouseList) ?? WarehouseModel(guidfixed: '', code: '');
                                if (result.code.isNotEmpty) {
                                  String newDefualttowarehouse = result.code;
                                  List<LanguageDataModel> newDefualttowarehousenames = result.names;
                                  String newDefualttolocation = "";
                                  List<LanguageDataModel> newDefualttolocationnames = [];

                                  if (result.location.isNotEmpty) {
                                    newDefualttolocation = result.location[0].code;
                                    newDefualttolocationnames = result.location[0].names;
                                  }

                                  onWarehouseChanged(
                                    defaultWarehouse,
                                    defaultWarehouseNames,
                                    defaultLocation,
                                    defaultLocationNames,
                                    newDefualttowarehouse,
                                    newDefualttowarehousenames,
                                    newDefualttolocation,
                                    newDefualttolocationnames,
                                  );
                                }
                              },
                            ),
                          ),
                          SizedBox(width: 4),
                          Expanded(
                            child: _buildWarehouseSelector(
                              label: global.language("position"),
                              value: global.activeLangName(defaultToLocationNames),
                              icon: Icons.location_on,
                              compact: true,
                              onTap: () async {
                                if (defaultToWarehouse.isNotEmpty) {
                                  LocationModel? result = await showWareHouseLocationDefualtDialog(context, defaultToWarehouse) ?? LocationModel(code: '');
                                  if (result.code.isNotEmpty) {
                                    onWarehouseChanged(
                                      defaultWarehouse,
                                      defaultWarehouseNames,
                                      defaultLocation,
                                      defaultLocationNames,
                                      defaultToWarehouse,
                                      defaultToWarehouseNames,
                                      result.code,
                                      result.names,
                                    );
                                  }
                                }
                              },
                              enabled: defaultToWarehouse.isNotEmpty,
                            ),
                          ),
                        ],
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ],
        ),
      );
    }
  }

  Widget _buildWarehouseSelector({required String label, required String value, required IconData icon, required VoidCallback onTap, bool enabled = true, bool compact = false}) {
    return GestureDetector(
      onTap: enabled ? onTap : null,
      child: Container(
        padding: EdgeInsets.symmetric(horizontal: compact ? 6 : 8, vertical: compact ? 4 : 6),
        decoration: BoxDecoration(
          color: enabled ? Colors.white : Colors.grey[100],
          border: Border.all(color: enabled ? Colors.grey[300]! : Colors.grey[200]!),
          borderRadius: BorderRadius.circular(10),
        ),
        child: Row(
          children: [
            Icon(icon, size: compact ? 14 : 16, color: enabled ? Colors.grey[600] : Colors.grey[400]),
            SizedBox(width: 6),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  if (!compact)
                    Text(
                      label,
                      style: TextStyle(fontSize: 10, color: Colors.grey[500], fontWeight: FontWeight.w500),
                    ),
                  Text(
                    value.isEmpty ? '${global.language("select_label")}$label' : value,
                    style: TextStyle(
                      fontSize: compact ? 11 : 12,
                      color: enabled ? (value.isEmpty ? Colors.grey[500] : Colors.black87) : Colors.grey[400],
                      fontWeight: value.isEmpty ? FontWeight.normal : FontWeight.w500,
                    ),
                    overflow: TextOverflow.ellipsis,
                  ),
                ],
              ),
            ),
            Icon(Icons.keyboard_arrow_down, size: compact ? 14 : 16, color: enabled ? Colors.grey[500] : Colors.grey[300]),
          ],
        ),
      ),
    );
  }

  Widget _buildProductListDetail(double maxWidth) {
    if (MediaQuery.of(context).size.width <= 799) {
      // Mobile view - use enhanced card with edit buttons
      return _buildMobileProductList();
    }

    // Web view - use optimized table layout with ListView.builder for performance
    final itemCount = screenData.details?.length ?? 0;

    if (itemCount == 0) {
      return Container(
        margin: const EdgeInsets.symmetric(horizontal: 4),
        padding: const EdgeInsets.all(16),
        decoration: BoxDecoration(
          border: Border.all(color: Colors.grey[300]!),
          borderRadius: BorderRadius.circular(10),
        ),
        child: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(Icons.inventory_2_outlined, size: 36, color: Colors.grey[400]),
              SizedBox(height: 8),
              Text(
                global.language("no_product_items"),
                style: TextStyle(fontSize: 14, color: Colors.grey[600], fontWeight: FontWeight.w500),
              ),
              SizedBox(height: 4),
              Text(global.language("press_add_product_button"), style: TextStyle(fontSize: 12, color: Colors.grey[500])),
            ],
          ),
        ),
      );
    }

    // ตารางรายการสินค้า
    return Container(
      margin: const EdgeInsets.symmetric(horizontal: 4),
      decoration: BoxDecoration(
        border: Border.all(color: Colors.grey[300]!),
        borderRadius: BorderRadius.circular(10),
      ),
      clipBehavior: Clip.antiAlias,
      child: ListView.builder(
        itemCount: itemCount,
        shrinkWrap: true,
        physics: const NeverScrollableScrollPhysics(),
        itemBuilder: (context, index) {
          return _buildOptimizedTableRow(index);
        },
      ),
    );
  }

  Widget _buildOptimizedTableRow(int index) {
    // Use RepaintBoundary for better performance with enhanced animations
    return RepaintBoundary(
      key: ValueKey('table_row_$index'),
      child: AnimatedSlide(
        duration: Duration(milliseconds: 300 + (index * 50).clamp(0, 200)),
        offset: Offset.zero,
        curve: Curves.easeOutBack,
        child: AnimatedOpacity(
          duration: Duration(milliseconds: 200 + (index * 30).clamp(0, 150)),
          opacity: 1.0,
          child: _buildSimpleTableRow(index),
        ),
      ),
    );
  }

  Widget _buildSimpleTableRow(int index) {
    final isEven = index % 2 == 0;

    return AnimatedContainer(
      duration: const Duration(milliseconds: 300),
      curve: Curves.easeInOut,
      decoration: BoxDecoration(
        color: isEven ? Colors.white : Colors.grey[50],
        border: Border(bottom: BorderSide(color: Colors.grey[200]!, width: 0.5)),
        boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.02), blurRadius: 1, offset: const Offset(0, 0.5))],
      ),
      child: TweenAnimationBuilder<double>(
        duration: const Duration(milliseconds: 400),
        tween: Tween(begin: 0.0, end: 1.0),
        curve: Curves.easeOutCubic,
        builder: (context, scale, child) {
          return Transform.scale(
            scale: 0.98 + (scale * 0.02),
            child: Table(
              columnWidths: {for (int i = 0; i < headers.length; i++) i: headers[i].code == 'line_number' ? const FixedColumnWidth(50.0) : FlexColumnWidth(headers[i].width)},
              defaultVerticalAlignment: TableCellVerticalAlignment.middle,
              children: [
                TableRow(
                  children: headers.map((header) => Container(padding: const EdgeInsets.symmetric(horizontal: 0, vertical: 0), child: _buildTableCellContent(header, index))).toList(),
                ),
              ],
            ),
          );
        },
      ),
    );
  }

  Widget _buildTableCellContent(global.DataTableHeader header, int index) {
    String dataText = _getDataText(header.code, index);
    final detail = screenData.details![index];

    if (header.code == 'delete') {
      return Semantics(
        label: '${global.language("delete_item_at")} ${index + 1}',
        button: true,
        onTapHint: global.language("press_to_delete_item"),
        child: Center(
          child: IconButton(
            padding: EdgeInsets.zero,
            constraints: const BoxConstraints(),
            onPressed: () {
              deleteItemDetail(index);
            },
            icon: Icon(Icons.delete_outline, color: Colors.red[600], size: 16),
            tooltip: global.language("delete_item"),
          ),
        ),
      );
    }

    return Semantics(
      label: _buildSemanticLabel(header, dataText, index),
      button: true,
      onTapHint: _buildSemanticHint(header.code),
      child: Focus(
        onKeyEvent: (node, event) {
          if (event.logicalKey.keyLabel == 'Enter' || event.logicalKey.keyLabel == ' ') {
            showDialogCommand(header.code, index, screenData.details![index]);
            return KeyEventResult.handled;
          }
          return KeyEventResult.ignored;
        },
        child: Material(
          color: Colors.transparent,
          child: InkWell(
            onTap: () {
              showDialogCommand(header.code, index, screenData.details![index]);
            },
            borderRadius: BorderRadius.circular(6),
            hoverColor: const Color(0xFF2A6F97).withValues(alpha: 0.1),
            splashColor: const Color(0xFF2A6F97).withValues(alpha: 0.2),
            focusColor: const Color(0xFF2A6F97).withValues(alpha: 0.15),
            child: TweenAnimationBuilder<double>(
              duration: const Duration(milliseconds: 300),
              tween: Tween(begin: 0.0, end: 1.0),
              curve: Curves.easeOutQuart,
              builder: (context, animation, child) {
                return AnimatedContainer(
                  duration: const Duration(milliseconds: 200),
                  curve: Curves.easeInOut,
                  padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 0),
                  transform: Matrix4.identity()..scale(0.995 + (animation * 0.005)),
                  decoration: BoxDecoration(
                    borderRadius: BorderRadius.circular(6),
                    border: Border.all(color: Colors.transparent, width: 1),
                  ),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: _shouldAlignRight(header.code) ? CrossAxisAlignment.end : CrossAxisAlignment.start,
                    children: [
                      // Main content - ราคาในสกุลเงินเอกสาร (Document Currency)
                      Row(
                        mainAxisSize: MainAxisSize.min,
                        mainAxisAlignment: _shouldAlignRight(header.code) ? MainAxisAlignment.end : MainAxisAlignment.start,
                        children: [
                          // แสดงสัญลักษณ์สกุลเงินเอกสาร (ถ้าเป็น multi-currency และเป็น column ราคา)
                          if (_isMultiCurrency() && _isPriceColumn(header.code) && dataText.isNotEmpty)
                            Padding(
                              padding: const EdgeInsets.only(right: 2),
                              child: Text(
                                _getDocCurrencySymbol(),
                                style: TextStyle(
                                  fontSize: _getSimpleFontSize(header.code) - 1,
                                  color: Colors.grey[600],
                                  fontWeight: FontWeight.normal,
                                ),
                              ),
                            ),
                          Flexible(
                            child: Text(
                              dataText.isEmpty ? '-' : dataText,
                              textAlign: _shouldAlignRight(header.code) ? TextAlign.right : header.textAlign,
                              style: TextStyle(
                                fontSize: _getSimpleFontSize(header.code),
                                color: dataText.isEmpty ? Colors.grey[500] : _getSimpleTextColor(header.code),
                                fontWeight: _getSimpleFontWeight(header.code),
                                fontFamily: _shouldUseMonospace(header.code) ? 'monospace' : null,
                              ),
                              maxLines: 2,
                              overflow: TextOverflow.ellipsis,
                            ),
                          ),
                        ],
                      ),

                      // แสดงราคาในสกุลเงินหลัก (Base Currency) - เฉพาะ Multi-Currency และ column ราคา
                      if (_isMultiCurrency() && _isPriceColumn(header.code) && dataText.isNotEmpty)
                        Padding(
                          padding: const EdgeInsets.only(top: 2),
                          child: Row(
                            mainAxisSize: MainAxisSize.min,
                            mainAxisAlignment: MainAxisAlignment.end,
                            children: [
                              Text(
                                '${_getBaseCurrencySymbol()} ${global.formatNumber(_getBaseCurrencyValue(header.code, index))}',
                                style: TextStyle(
                                  fontSize: _getSimpleFontSize(header.code) - 2,
                                  color: Colors.grey[500],
                                  fontStyle: FontStyle.italic,
                                ),
                              ),
                            ],
                          ),
                        ),

                      // Show Description and Options under product name only
                      if (header.code == 'product_name') ...[
                        // Description
                        if (detail.description != null && detail.description!.trim().isNotEmpty) ...[
                          SizedBox(height: 4),
                          Semantics(
                            label: '${global.language("note")}: ${detail.description}',
                            button: true,
                            onTapHint: global.language("press_to_edit_note"),
                            child: InkWell(
                              onTap: () {
                                showDialogCommand('description', index, detail);
                              },
                              borderRadius: BorderRadius.circular(4),
                              hoverColor: Colors.amber[100]?.withValues(alpha: 0.5),
                              child: AnimatedContainer(
                                duration: const Duration(milliseconds: 150),
                                padding: const EdgeInsets.all(6),
                                decoration: BoxDecoration(
                                  color: _getAccessibleColor(Colors.amber[50]!),
                                  borderRadius: BorderRadius.circular(4),
                                  border: Border.all(color: Colors.amber[200]!.withValues(alpha: 0.6), width: 0.8),
                                  boxShadow: [BoxShadow(color: Colors.amber.withValues(alpha: 0.08), blurRadius: 2, offset: const Offset(0, 1))],
                                ),
                                child: Text(
                                  '${global.language("note")}: ${detail.description}',
                                  style: TextStyle(fontSize: 9, color: Colors.amber[800], fontStyle: FontStyle.italic),
                                  maxLines: 2,
                                  overflow: TextOverflow.ellipsis,
                                ),
                              ),
                            ),
                          ),
                        ],

                        // Options
                        if (detail.extrajsonlist != null && detail.extrajsonlist!.isNotEmpty) ...[
                          SizedBox(height: 4),
                          Semantics(
                            label: '${global.language("additional_options")} ${detail.extrajsonlist!.length} ${global.language("items")}',
                            button: true,
                            onTapHint: global.language("press_to_edit_options"),
                            child: InkWell(
                              onTap: () {
                                showDialogCommand('options', index, detail);
                              },
                              borderRadius: BorderRadius.circular(4),
                              hoverColor: Colors.indigo[100]?.withValues(alpha: 0.5),
                              child: AnimatedContainer(
                                duration: const Duration(milliseconds: 150),
                                padding: const EdgeInsets.all(6),
                                decoration: BoxDecoration(
                                  color: _getAccessibleColor(Colors.indigo[50]!),
                                  borderRadius: BorderRadius.circular(4),
                                  border: Border.all(color: Colors.indigo[200]!.withValues(alpha: 0.6), width: 0.8),
                                  boxShadow: [BoxShadow(color: Colors.indigo.withValues(alpha: 0.08), blurRadius: 2, offset: const Offset(0, 1))],
                                ),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    Text(
                                      '${global.language("options")} (${detail.extrajsonlist!.length})',
                                      style: TextStyle(fontSize: 9, fontWeight: FontWeight.w600, color: Colors.indigo[700]),
                                    ),
                                    const SizedBox(height: 2),
                                    Wrap(
                                      spacing: 4,
                                      runSpacing: 2,
                                      children:
                                          detail.extrajsonlist!
                                              .take(2)
                                              .map(
                                                (option) => Container(
                                                  padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
                                                  decoration: BoxDecoration(
                                                    color: Colors.white,
                                                    borderRadius: BorderRadius.circular(6),
                                                    border: Border.all(color: Colors.indigo[300]!),
                                                  ),
                                                  child: Text(
                                                    '${global.activeLangName(option.itemnames!)} (${global.formatNumber(option.price!)})',
                                                    style: TextStyle(fontSize: 8, color: Colors.indigo[700], fontWeight: FontWeight.w500),
                                                  ),
                                                ),
                                              )
                                              .toList()
                                            ..addAll(
                                              detail.extrajsonlist!.length > 2
                                                  ? [
                                                      Container(
                                                        padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
                                                        decoration: BoxDecoration(color: Colors.indigo[100], borderRadius: BorderRadius.circular(6)),
                                                        child: Text(
                                                          '+${detail.extrajsonlist!.length - 2}',
                                                          style: TextStyle(fontSize: 8, color: Colors.indigo[700], fontWeight: FontWeight.w600),
                                                        ),
                                                      ),
                                                    ]
                                                  : [],
                                            ),
                                    ),
                                  ],
                                ),
                              ),
                            ),
                          ),
                        ],
                      ],
                    ],
                  ),
                );
              },
            ),
          ),
        ),
      ),
    );
  }

  Widget _buildMobileProductList() {
    final itemCount = screenData.details?.length ?? 0;

    if (itemCount == 0) {
      return Container(
        padding: EdgeInsets.all(16),
        child: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              Icon(Icons.shopping_cart_outlined, size: 36, color: Colors.grey[400]),
              SizedBox(height: 8),
              Text(
                global.language("no_product_items"),
                style: TextStyle(fontSize: 14, color: Colors.grey[600], fontWeight: FontWeight.w500),
              ),
              SizedBox(height: 4),
              Text(global.language("tap_add_product_button"), style: TextStyle(fontSize: 12, color: Colors.grey[500])),
            ],
          ),
        ),
      );
    }

    return Column(
      children: [
        // รายการสินค้า (Card)
        ListView.builder(
          itemCount: itemCount,
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          itemBuilder: (context, index) {
            return Column(
              children: [RepaintBoundary(key: ValueKey('mobile_card_$index'), child: _buildMobileCard(index))],
            );
          },
        ),
        // แสดงยอดรวมด้านล่าง (เฉพาะ transaction type ที่แสดงราคา)
        if (_shouldShowPriceAmount())
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 4),
            child: DocumentTotalWidget(screenData: screenData),
          ),
      ],
    );
  }

  Widget _buildMobileCard(int index) {
    final detail = screenData.details![index];
    final productName = global.activeLangName(detail.itemnames!);
    final qty = global.formatQuantity(detail.qty);
    final price = _shouldShowPriceAmount() ? global.formatUnitPrice(detail.price) : '';
    final amount = _shouldShowPriceAmount() ? global.formatNumber(detail.sumamount) : '';

    String cardLabel = '${global.language("item_at")} ${index + 1}: $productName';
    if (qty.isNotEmpty) cardLabel += ', ${global.language("quantity")} $qty';
    if (price.isNotEmpty) cardLabel += ', ${global.language("price")} $price';
    if (amount.isNotEmpty) cardLabel += ', ${global.language("amount")} $amount';

    return TweenAnimationBuilder<double>(
      duration: Duration(milliseconds: 400 + (index * 50)),
      tween: Tween(begin: 0.0, end: 1.0),
      curve: Curves.easeOutCubic,
      builder: (context, animation, child) {
        return AnimatedSlide(
          duration: Duration(milliseconds: 350 + (index * 30)),
          offset: Offset(0, (1 - animation) * 0.2),
          curve: Curves.easeOutBack,
          child: AnimatedOpacity(
            duration: Duration(milliseconds: 300 + (index * 20)),
            opacity: animation,
            child: Transform.scale(
              scale: 0.96 + (animation * 0.04),
              child: Semantics(
                label: cardLabel,
                button: true,
                onTapHint: global.language("press_to_edit_product"),
                child: Container(
                  margin: const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
                  decoration: BoxDecoration(
                    color: Colors.white,
                    border: Border.all(color: Colors.grey[300]!, width: 0.5),
                    borderRadius: BorderRadius.circular(10),
                    boxShadow: [BoxShadow(color: Colors.black.withValues(alpha: 0.04), blurRadius: 4, offset: const Offset(0, 2))],
                  ),
                  child: Focus(
                    onKeyEvent: (node, event) {
                      if (event.logicalKey.keyLabel == 'Enter' || event.logicalKey.keyLabel == ' ') {
                        showDialogCommand('product_name', index, detail);
                        return KeyEventResult.handled;
                      }
                      return KeyEventResult.ignored;
                    },
                    child: Material(
                      color: Colors.transparent,
                      borderRadius: BorderRadius.circular(8),
                      child: InkWell(
                        onTap: () {
                          showDialogCommand('product_name', index, detail);
                        },
                        borderRadius: BorderRadius.circular(10),
                        splashColor: const Color(0xFF2A6F97).withValues(alpha: 0.1),
                        focusColor: const Color(0xFF2A6F97).withValues(alpha: 0.15),
                        child: Padding(
                          padding: const EdgeInsets.all(8),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              // Header row
                              Row(
                                children: [
                                  Container(
                                    padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                                    decoration: BoxDecoration(color: const Color(0xFF2A6F97).withValues(alpha: 0.1), borderRadius: BorderRadius.circular(3)),
                                    child: Text(
                                      '#${index + 1}',
                                      style: TextStyle(fontSize: 10, fontWeight: FontWeight.bold, color: const Color(0xFF2A6F97)),
                                    ),
                                  ),
                                  const Spacer(),
                                  IconButton(
                                    padding: EdgeInsets.zero,
                                    constraints: const BoxConstraints(minWidth: 32, minHeight: 32),
                                    onPressed: () {
                                      deleteItemDetail(index);
                                    },
                                    icon: Icon(Icons.delete_outline, color: Colors.red[600], size: 18),
                                  ),
                                ],
                              ),

                              const SizedBox(height: 4),

                              // Product name - clickable button
                              InkWell(
                                onTap: () {
                                  showDialogCommand('product_name', index, detail);
                                },
                                borderRadius: BorderRadius.circular(6),
                                child: Container(
                                  width: double.infinity,
                                  padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
                                  decoration: BoxDecoration(
                                    color: Colors.blue[50],
                                    borderRadius: BorderRadius.circular(6),
                                    border: Border.all(color: Colors.blue[200]!),
                                  ),
                                  child: Row(
                                    children: [
                                      Icon(Icons.edit, size: 14, color: Colors.blue[600]),
                                      const SizedBox(width: 6),
                                      Expanded(
                                        child: Text(
                                          global.activeLangName(detail.itemnames!),
                                          style: TextStyle(fontSize: 13, fontWeight: FontWeight.w600, color: Colors.blue[800]),
                                          maxLines: 2,
                                          overflow: TextOverflow.ellipsis,
                                        ),
                                      ),
                                    ],
                                  ),
                                ),
                              ),

                              if (detail.barcode.isNotEmpty || detail.itemcode.isNotEmpty) ...[
                                const SizedBox(height: 6),
                                Row(
                                  children: [
                                    if (detail.barcode.isNotEmpty) ...[
                                      Expanded(
                                        child: InkWell(
                                          onTap: () {
                                            showDialogCommand('barcode', index, detail);
                                          },
                                          borderRadius: BorderRadius.circular(4),
                                          child: Container(
                                            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
                                            decoration: BoxDecoration(
                                              color: Colors.teal[50],
                                              borderRadius: BorderRadius.circular(4),
                                              border: Border.all(color: Colors.teal[200]!),
                                            ),
                                            child: Row(
                                              mainAxisSize: MainAxisSize.min,
                                              children: [
                                                Icon(Icons.qr_code, size: 12, color: Colors.teal[600]),
                                                const SizedBox(width: 4),
                                                Flexible(
                                                  child: Text(
                                                    detail.barcode,
                                                    style: TextStyle(fontSize: 10, color: Colors.teal[700], fontFamily: 'monospace', fontWeight: FontWeight.w500),
                                                    overflow: TextOverflow.ellipsis,
                                                  ),
                                                ),
                                              ],
                                            ),
                                          ),
                                        ),
                                      ),
                                    ],
                                    if (detail.itemcode.isNotEmpty) ...[
                                      if (detail.barcode.isNotEmpty) const SizedBox(width: 8),
                                      Expanded(
                                        child: InkWell(
                                          onTap: () {
                                            showDialogCommand('item_code', index, detail);
                                          },
                                          borderRadius: BorderRadius.circular(4),
                                          child: Container(
                                            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
                                            decoration: BoxDecoration(
                                              color: Colors.cyan[50],
                                              borderRadius: BorderRadius.circular(4),
                                              border: Border.all(color: Colors.cyan[200]!),
                                            ),
                                            child: Row(
                                              mainAxisSize: MainAxisSize.min,
                                              children: [
                                                Icon(Icons.tag, size: 12, color: Colors.cyan[600]),
                                                const SizedBox(width: 4),
                                                Flexible(
                                                  child: Text(
                                                    detail.itemcode,
                                                    style: TextStyle(fontSize: 10, color: Colors.cyan[700], fontWeight: FontWeight.w500),
                                                    overflow: TextOverflow.ellipsis,
                                                  ),
                                                ),
                                              ],
                                            ),
                                          ),
                                        ),
                                      ),
                                    ],
                                  ],
                                ),
                              ],

                              const SizedBox(height: 8),

                              // Info grid - clickable buttons
                              Row(
                                children: [
                                  Expanded(
                                    child: _buildMobileClickableInfoItem(global.language("quantity"), global.formatQuantity(detail.qty), Colors.orange, () => showDialogCommand('qty', index, detail)),
                                  ),
                                  if (_shouldShowPriceAmount()) ...[
                                    SizedBox(width: 8),
                                    Expanded(
                                      child: _buildMobileClickableInfoItem(global.language("price"), global.formatUnitPrice(detail.price), Colors.blue, () => showDialogCommand('price', index, detail)),
                                    ),
                                    SizedBox(width: 8),
                                    Expanded(
                                      child: _buildMobileClickableInfoItem(
                                        global.language("amount"),
                                        global.formatNumber(detail.sumamount),
                                        Colors.green,
                                        () => showDialogCommand('product_amount', index, detail),
                                      ),
                                    ),
                                  ],
                                ],
                              ),

                              // Unit info - clickable button
                              if (global.activeLangName(detail.unitnames!).isNotEmpty) ...[
                                SizedBox(height: 8),
                                InkWell(
                                  onTap: () {
                                    showDialogCommand('product_unit', index, detail);
                                  },
                                  borderRadius: BorderRadius.circular(6),
                                  child: Container(
                                    padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                                    decoration: BoxDecoration(
                                      color: Colors.pink[50],
                                      borderRadius: BorderRadius.circular(6),
                                      border: Border.all(color: Colors.pink[200]!),
                                    ),
                                    child: Row(
                                      mainAxisSize: MainAxisSize.min,
                                      children: [
                                        Icon(Icons.straighten, size: 14, color: Colors.pink[600]),
                                        SizedBox(width: 6),
                                        Text(
                                          '${global.language("unit_label")}: ${global.activeLangName(detail.unitnames!)}',
                                          style: TextStyle(fontSize: 11, color: Colors.pink[700], fontWeight: FontWeight.w500),
                                        ),
                                      ],
                                    ),
                                  ),
                                ),
                              ],

                              // Warehouse info - clickable buttons
                              if (global.activeLangName(detail.whnames!).isNotEmpty) ...[
                                const SizedBox(height: 8),
                                Row(
                                  children: [
                                    // Warehouse button
                                    Expanded(
                                      child: InkWell(
                                        onTap: () {
                                          showDialogCommand('warehouse', index, detail);
                                        },
                                        borderRadius: BorderRadius.circular(6),
                                        child: Container(
                                          padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                                          decoration: BoxDecoration(
                                            color: Colors.purple[50],
                                            borderRadius: BorderRadius.circular(6),
                                            border: Border.all(color: Colors.purple[200]!),
                                          ),
                                          child: Row(
                                            mainAxisSize: MainAxisSize.min,
                                            children: [
                                              Icon(Icons.warehouse, size: 14, color: Colors.purple[600]),
                                              const SizedBox(width: 4),
                                              Flexible(
                                                child: Text(
                                                  global.activeLangName(detail.whnames!),
                                                  style: TextStyle(fontSize: 10, color: Colors.purple[700], fontWeight: FontWeight.w500),
                                                  overflow: TextOverflow.ellipsis,
                                                ),
                                              ),
                                            ],
                                          ),
                                        ),
                                      ),
                                    ),
                                    if (global.activeLangName(detail.locationnames!).isNotEmpty) ...[
                                      const SizedBox(width: 8),
                                      // Location button
                                      Expanded(
                                        child: InkWell(
                                          onTap: () {
                                            showDialogCommand('location', index, detail);
                                          },
                                          borderRadius: BorderRadius.circular(6),
                                          child: Container(
                                            padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
                                            decoration: BoxDecoration(
                                              color: Colors.indigo[50],
                                              borderRadius: BorderRadius.circular(6),
                                              border: Border.all(color: Colors.indigo[200]!),
                                            ),
                                            child: Row(
                                              mainAxisSize: MainAxisSize.min,
                                              children: [
                                                Icon(Icons.location_on, size: 14, color: Colors.indigo[600]),
                                                const SizedBox(width: 4),
                                                Flexible(
                                                  child: Text(
                                                    global.activeLangName(detail.locationnames!),
                                                    style: TextStyle(fontSize: 10, color: Colors.indigo[700], fontWeight: FontWeight.w500),
                                                    overflow: TextOverflow.ellipsis,
                                                  ),
                                                ),
                                              ],
                                            ),
                                          ),
                                        ),
                                      ),
                                    ],
                                  ],
                                ),
                              ],

                              // Description
                              if (detail.description != null && detail.description!.trim().isNotEmpty) ...[
                                SizedBox(height: 8),
                                InkWell(
                                  onTap: () {
                                    showDialogCommand('description', index, detail);
                                  },
                                  borderRadius: BorderRadius.circular(3),
                                  child: Container(
                                    padding: const EdgeInsets.all(6),
                                    decoration: BoxDecoration(
                                      color: Colors.amber[50],
                                      borderRadius: BorderRadius.circular(3),
                                      border: Border.all(color: Colors.amber[200]!.withValues(alpha: 0.5), width: 1),
                                    ),
                                    child: Text(
                                      '${global.language("note")}: ${detail.description}',
                                      style: TextStyle(fontSize: 10, color: Colors.amber[800], fontStyle: FontStyle.italic),
                                    ),
                                  ),
                                ),
                              ],

                              // Options
                              if (detail.extrajsonlist != null && detail.extrajsonlist!.isNotEmpty) ...[
                                SizedBox(height: 8),
                                InkWell(
                                  onTap: () {
                                    showDialogCommand('options', index, detail);
                                  },
                                  borderRadius: BorderRadius.circular(3),
                                  child: Container(
                                    padding: const EdgeInsets.all(6),
                                    decoration: BoxDecoration(
                                      color: Colors.indigo[50],
                                      borderRadius: BorderRadius.circular(3),
                                      border: Border.all(color: Colors.indigo[200]!.withValues(alpha: 0.5), width: 1),
                                    ),
                                    child: Column(
                                      crossAxisAlignment: CrossAxisAlignment.start,
                                      children: [
                                        Text(
                                          '${global.language("additional_options")} (${detail.extrajsonlist!.length} ${global.language("items")})',
                                          style: TextStyle(fontSize: 10, fontWeight: FontWeight.w600, color: Colors.indigo[700]),
                                        ),
                                        const SizedBox(height: 4),
                                        Wrap(
                                          spacing: 4,
                                          runSpacing: 2,
                                          children:
                                              detail.extrajsonlist!
                                                  .take(2)
                                                  .map(
                                                    (option) => Container(
                                                      padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
                                                      decoration: BoxDecoration(
                                                        color: Colors.white,
                                                        borderRadius: BorderRadius.circular(6),
                                                        border: Border.all(color: Colors.indigo[300]!),
                                                      ),
                                                      child: Text(
                                                        '${global.activeLangName(option.itemnames!)} (${global.formatNumber(option.price!)})',
                                                        style: TextStyle(fontSize: 8, color: Colors.indigo[700], fontWeight: FontWeight.w500),
                                                      ),
                                                    ),
                                                  )
                                                  .toList()
                                                ..addAll(
                                                  detail.extrajsonlist!.length > 2
                                                      ? [
                                                          Container(
                                                            padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 1),
                                                            decoration: BoxDecoration(color: Colors.indigo[100], borderRadius: BorderRadius.circular(6)),
                                                            child: Text(
                                                              '+${detail.extrajsonlist!.length - 2}',
                                                              style: TextStyle(fontSize: 8, color: Colors.indigo[700], fontWeight: FontWeight.w600),
                                                            ),
                                                          ),
                                                        ]
                                                      : [],
                                                ),
                                        ),
                                      ],
                                    ),
                                  ),
                                ),
                              ], // Closing the options if condition
                            ], // Closing the main children list
                          ),
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
    );
  }

  Widget _buildMobileClickableInfoItem(String label, String value, MaterialColor color, VoidCallback onTap) {
    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(6),
      child: Container(
        padding: const EdgeInsets.all(8),
        decoration: BoxDecoration(
          color: color[50],
          borderRadius: BorderRadius.circular(6),
          border: Border.all(color: color[200]!),
          boxShadow: [BoxShadow(color: color[100]!.withValues(alpha: 0.3), blurRadius: 2, offset: const Offset(0, 1))],
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  label,
                  style: TextStyle(fontSize: 9, color: color[600], fontWeight: FontWeight.w500),
                ),
                const SizedBox(width: 4),
                Icon(Icons.edit, size: 10, color: color[500]),
              ],
            ),
            const SizedBox(height: 4),
            Text(
              value,
              style: TextStyle(fontSize: 12, color: color[800], fontWeight: FontWeight.bold),
              textAlign: TextAlign.center,
            ),
          ],
        ),
      ),
    );
  }

  // Simplified styling methods
  double _getSimpleFontSize(String code) {
    switch (code) {
      case 'product_name':
        return 12;
      case 'line_number':
        return 11;
      default:
        return 11;
    }
  }

  Color _getSimpleTextColor(String code) {
    switch (code) {
      case 'product_amount':
        return const Color(0xFF2A6F97);
      case 'product_price':
        return const Color(0xFF2A6F97).withValues(alpha: 0.8);
      case 'product_qty':
        return Colors.orange[700]!;
      case 'line_number':
        return Colors.grey[600]!;
      default:
        return Colors.black87;
    }
  }

  FontWeight _getSimpleFontWeight(String code) {
    // ไม่ใช้ bold สำหรับทุก column ใน row
    return FontWeight.normal;
  }

  bool _shouldUseMonospace(String code) {
    return code == 'barcode' || code == 'item_code';
  }

  bool _shouldAlignRight(String code) {
    return code == 'product_qty' || code == 'product_price' || code == 'product_price_adjust' || code == 'product_amount' || code == 'eventqty' || code == 'line_number';
  }

  // Accessibility helper methods
  String _buildSemanticLabel(global.DataTableHeader header, String dataText, int index) {
    final fieldName = _getFieldDisplayName(header.code);
    final value = dataText.isEmpty ? global.language("no_data") : dataText;
    return '${global.language("item_at")} ${index + 1} $fieldName: $value';
  }

  String _buildSemanticHint(String code) {
    switch (code) {
      case 'product_name':
        return global.language("press_to_select_product");
      case 'product_qty':
        return global.language("press_to_edit_quantity");
      case 'product_price':
        return global.language("press_to_edit_price");
      case 'barcode':
        return global.language("press_to_edit_barcode");
      case 'product_ware_house':
        return global.language("press_to_select_warehouse");
      case 'product_location':
        return global.language("press_to_select_location");
      default:
        return global.language("press_to_edit_data");
    }
  }

  String _getFieldDisplayName(String code) {
    switch (code) {
      case 'line_number':
        return global.language("sequence_number");
      case 'product_name':
        return global.language("product_name");
      case 'product_qty':
        return global.language("quantity");
      case 'product_price':
        return global.language("price");
      case 'product_amount':
        return global.language("amount");
      case 'barcode':
        return global.language("barcode");
      case 'product_ware_house':
        return global.language("warehouse");
      case 'product_location':
        return global.language("position");
      default:
        return global.language("data_field");
    }
  }

  // High contrast support
  Color _getAccessibleColor(Color baseColor) {
    // For high contrast, use slightly darker shades
    if (baseColor == Colors.amber[50]) return Colors.amber[100]!;
    if (baseColor == Colors.indigo[50]) return Colors.indigo[100]!;
    return baseColor;
  }

  bool _shouldShowPriceAmount() {
    return !(transactionType == global.TransactionTypeEnum.stocktransfer ||
        transactionType == global.TransactionTypeEnum.stockpickupproduct ||
        transactionType == global.TransactionTypeEnum.stockreturnproduct ||
        transactionType == global.TransactionTypeEnum.adjust);
  }

  /// ตรวจสอบว่าเป็นเอกสาร Multi-Currency หรือไม่
  /// (docCurrency != currency และ exchangerate != 1)
  bool _isMultiCurrency() {
    return screenData.docCurrency != null &&
        screenData.docCurrency!.isNotEmpty &&
        screenData.currency != null &&
        screenData.currency!.isNotEmpty &&
        screenData.docCurrency != screenData.currency &&
        (screenData.exchangerate ?? 1.0) != 1.0;
  }

  /// ดึงสัญลักษณ์สกุลเงินเอกสาร
  String _getDocCurrencySymbol() {
    return screenData.docCurrencySymbol ?? screenData.docCurrency ?? '';
  }

  /// ดึงสัญลักษณ์สกุลเงินหลัก
  String _getBaseCurrencySymbol() {
    return screenData.currencysymbol ?? screenData.currency ?? '฿';
  }

  /// ตรวจสอบว่า column นี้ต้องแสดงข้อมูลราคาหรือไม่
  bool _isPriceColumn(String code) {
    return code == 'product_price' ||
        code == 'product_price_adjust' ||
        code == 'product_discount' ||
        code == 'product_amount';
  }

  /// ดึงค่า base currency จาก column (สำหรับแสดง sub-line ราคาในสกุลเงินหลัก)
  double _getBaseCurrencyValue(String code, int index) {
    if (index < 0 || index >= (screenData.details?.length ?? 0)) return 0;
    final detail = screenData.details![index];
    switch (code) {
      case 'product_price':
      case 'product_price_adjust':
        return detail.price;
      case 'product_discount':
        final discountStr = detail.discount;
        if (discountStr.isEmpty) return 0;
        final cleanDiscount = discountStr.replaceAll('%', '').trim();
        return double.tryParse(cleanDiscount) ?? 0;
      case 'product_amount':
        return detail.sumamount;
      default:
        return 0;
    }
  }

  String _getDataText(String code, int index) {
    final detail = screenData.details![index];
    switch (code) {
      case "delete":
        return "";
      case "line_number":
        return (index + 1).toString();
      case "barcode":
        return detail.barcode;
      case "item_code":
        return detail.itemcode;
      case "product_name":
        return global.activeLangName(detail.itemnames!);
      case "docref":
        return detail.docref ?? "";
      case "docrefdatetime":
        if (detail.docrefdatetime != null && detail.docrefdatetime!.isNotEmpty) {
          try {
            DateTime date = DateTime.parse(detail.docrefdatetime!);
            return "${date.day.toString().padLeft(2, '0')}/${date.month.toString().padLeft(2, '0')}/${date.year}";
          } catch (e) {
            return detail.docrefdatetime ?? "";
          }
        }
        return "";
      case "eventqty":
        return global.formatNumber(detail.eventqty ?? 0);
      case "product_ware_house":
        return global.activeLangName(detail.whnames!);
      case "product_location":
        return global.activeLangName(detail.locationnames!);
      case "product_to_ware_house":
        return global.activeLangName(detail.towhnames!);
      case "product_to_location":
        return global.activeLangName(detail.tolocationnames!);
      case "product_unit":
        return global.activeLangName(detail.unitnames!);
      case "product_qty":
        return global.formatQuantity(detail.qty);
      case "product_price_adjust":
      case "product_price":
        // ถ้า multi-currency + มี priceDoc → แสดงราคาในสกุลเงินเอกสาร
        if (_isMultiCurrency() && detail.priceDoc != null) {
          return global.formatUnitPrice(detail.priceDoc!);
        }
        return global.formatUnitPrice(detail.price);
      case "product_discount":
        return detail.discount;
      case "product_amount":
        // ถ้า multi-currency + มี sumAmountDoc → แสดงยอดรวมในสกุลเงินเอกสาร
        if (_isMultiCurrency() && detail.sumAmountDoc != null) {
          return global.formatNumber(detail.sumAmountDoc!);
        }
        return global.formatNumber(detail.sumamount);
      default:
        return "";
    }
  }

  void _addProductFromBarcode(ProductBarcodeModel result) {
    String whcode = defaultWarehouse;
    List<LanguageDataModel> whnames = defaultWarehouseNames;
    String locationcode = defaultLocation;
    List<LanguageDataModel> locationnames = defaultLocationNames;

    screenData.details!.add(
      TransactionDetailModel(
        docdatetime: DateTime.now().toUtc().toIso8601String(), // แปลงเป็น UTC ตามกฎ CLAUDE.md
        itemguid: result.guidfixed,
        barcode: result.barcode!,
        itemcode: result.itemcode ?? "",
        itemnames: result.names,
        unitcode: result.itemunitcode,
        qty: (transactionType == global.TransactionTypeEnum.adjust) ? 0 : 1,
        price: (transactionType == global.TransactionTypeEnum.adjust) ? 0 : getPrice(result.prices),
        discount: '',
        sumofcost: 0,
        sumamount: 0,
        remark: '',
        linenumber: 0,
        whcode: whcode,
        whnames: whnames,
        shelfcode: '',
        locationcode: locationcode,
        locationnames: locationnames,
        totalvaluevat: 0,
        totalqty: 0,
        standvalue: result.standvalue!,
        dividevalue: result.dividevalue!,
        multiunit: true,
        unitnames: result.itemunitnames,
        calcflag: calcflag,
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
        vatcal: result.vatcal,
        towhcode: defaultToWarehouse,
        towhnames: defaultToWarehouseNames,
        tolocationcode: defaultToLocation,
        tolocationnames: defaultToLocationNames,
        refbarcodes: result.refbarcodes,
        manufacturerguid: result.manufacturerguid,
        description: "", // Initialize description field
      ),
    );

    // Multi-currency: คำนวณ priceDoc จาก base price
    if (_isMultiCurrency()) {
      final lastDetail = screenData.details!.last;
      final rate = screenData.exchangerate ?? 1.0;
      if (rate > 0) {
        lastDetail.priceDoc = lastDetail.price / rate;
      }
    }

    calTotalValue();
    setState(() {});
  }

  void _addEmptyProduct() {
    // เปิดหน้าค้นหาสินค้าแทนการเพิ่มแถวว่าง
    Navigator.push(
      context,
      MaterialPageRoute(
        builder: (context) => BarcodeSearchScreen(
          word: '',
          screen: (transactionType == global.TransactionTypeEnum.sale || transactionType == global.TransactionTypeEnum.saleorder || transactionType == global.TransactionTypeEnum.salereturn)
              ? 'not_material'
              : 'material',
        ),
      ),
    ).then((value) {
      if (value != null) {
        ProductBarcodeModel result = value;
        if (result.barcode != null && result.barcode!.trim().isNotEmpty) {
          _addProductFromBarcode(result);
        }
      }
    });
  }
}
