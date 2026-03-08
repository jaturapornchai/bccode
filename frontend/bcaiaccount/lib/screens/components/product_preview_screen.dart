import 'dart:typed_data';
import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/model/price_model.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/widgets/unit_flow_widget.dart';

class ProductPreviewScreen extends StatelessWidget {
  final ProductBarcodeModel screenData;
  final List<PriceModel> priceList;
  final Uint8List? imageWeb;

  /// รายการ barcode ทั้งหมดของสินค้านี้ (สำหรับแสดงแผนผังหน่วยนับ)
  final List<ProductBarcodeModel> allProductBarcodes;

  const ProductPreviewScreen({
    super.key,
    required this.screenData,
    required this.priceList,
    this.imageWeb,
    this.allProductBarcodes = const [],
  });

  String _productTypeName(int productType) {
    switch (productType) {
      case 0:
        return global.language("product_is_stock");
      case 1:
        return global.language("product_is_service");
      case 2:
        return global.language("product_is_set");
      case 3:
        return global.language("product_is_material");
      default:
        return "-";
    }
  }

  String _vatTypeName(int? vatcal) {
    return (vatcal == 0)
        ? global.language("product_vat_type_1")
        : global.language("product_vat_type_2");
  }

  String _pointName(bool? issumpoint) {
    return (issumpoint == true)
        ? global.language("product_use_point_1")
        : global.language("product_use_point_2");
  }

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 4),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // === Header: Image + Basic Info ===
          _buildHeaderSection(),
          _divider(),
          // === All Language Names ===
          _buildNamesSection(),
          _divider(),
          // === Prices ===
          if (priceList.isNotEmpty) ...[
            _buildPriceSection(),
            _divider(),
          ],
          // === Product Type Details ===
          _buildProductTypeDetailSection(),
          _divider(),
          // === Master Data ===
          _buildMasterDataSection(),
          _divider(),
          // === Settings ===
          _buildSettingsSection(),
          // === Reference Barcodes ===
          if (screenData.isusesubbarcodes == true &&
              screenData.refbarcodes!.isNotEmpty) ...[
            _divider(),
            _buildRefBarcodeSection(),
          ],
          // === แผนผังหน่วยนับ ===
          if (allProductBarcodes.length > 1) ...[
            _divider(),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 8),
              child: UnitFlowWidget(
                productBarcodes: allProductBarcodes,
                currentBarcode: screenData.barcode ?? '',
              ),
            ),
          ],
          // === BOM ===
          if (screenData.bom?.isNotEmpty == true) ...[
            _divider(),
            _buildBomSection(),
          ],
          // === Dimensions ===
          if (screenData.dimensions?.isNotEmpty == true) ...[
            _divider(),
            _buildDimensionsSection(),
          ],
          // === Business Type & Branch ===
          if (screenData.businesstypes?.isNotEmpty == true ||
              screenData.ignorebranches?.isNotEmpty == true) ...[
            _divider(),
            _buildBusinessSection(),
          ],
          // === Fixed Cost ===
          if (screenData.fixedcost?.isNotEmpty == true) ...[
            _divider(),
            _buildFixedCostSection(),
          ],
          // === Time for Sale ===
          if (screenData.timeforsales?.isNotEmpty == true) ...[
            _divider(),
            _buildTimeForSaleSection(),
          ],
          // === Order Types ===
          if (screenData.ordertypes!.isNotEmpty) ...[
            _divider(),
            _buildOrderTypesSection(),
          ],
          // === Options ===
          if (screenData.options!.isNotEmpty) ...[
            _divider(),
            _buildOptionsSection(),
          ],
          // === Description ===
          if (screenData.description?.isNotEmpty == true) ...[
            _divider(),
            _buildDescriptionSection(),
          ],
          const SizedBox(height: 8),
        ],
      ),
    );
  }

  Widget _divider() {
    return Divider(height: 1, color: Colors.grey[300]);
  }

  // === HEADER: Image + Basic Info ===
  Widget _buildHeaderSection() {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Thumbnail
          Container(
            width: 80,
            height: 80,
            decoration: BoxDecoration(
              color: (screenData.useimageorcolor ?? true) == false
                  ? _parseColor()
                  : Colors.grey[100],
              borderRadius: BorderRadius.circular(6),
              border: Border.all(color: Colors.grey[300]!),
              image: (screenData.useimageorcolor ?? true) != false
                  ? DecorationImage(
                      image: (imageWeb != null)
                          ? MemoryImage(imageWeb!) as ImageProvider
                          : (screenData.imageuri != null &&
                                  screenData.imageuri!.isNotEmpty)
                              ? NetworkImage(global.resolveFileUrl(screenData.imageuri!))
                                  as ImageProvider
                              : const AssetImage('assets/img/noimage.png'),
                      fit: BoxFit.contain,
                    )
                  : null,
            ),
          ),
          const SizedBox(width: 10),
          // Basic Info
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  global.activeLangName(screenData.names!),
                  style: const TextStyle(
                    fontSize: 16,
                    fontWeight: FontWeight.bold,
                  ),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
                const SizedBox(height: 4),
                _row2col(
                  global.language("barcode"),
                  screenData.barcode ?? "-",
                  global.language("item_code"),
                  screenData.itemcode ?? "-",
                ),
                _row2col(
                  global.language("unit"),
                  "${screenData.itemunitcode} | ${global.activeLangName(screenData.itemunitnames!)}",
                  global.language("product_type"),
                  _productTypeName(screenData.itemtype ?? 0),
                ),
                _row2col(
                  global.language("vat_type"),
                  _vatTypeName(screenData.vatcal),
                  global.language("issumpoint"),
                  _pointName(screenData.issumpoint),
                ),
                _row2col(
                  global.language("discount"),
                  (screenData.discount?.isNotEmpty == true)
                      ? screenData.discount!
                      : "-",
                  global.language("product_group_code"),
                  (screenData.groupcode?.isNotEmpty == true)
                      ? "${screenData.groupcode} | ${global.activeLangName(screenData.groupnames!)}"
                      : "-",
                ),
                // === Unit Relation (all units of same itemcode) ===
                if (screenData.allUnitNames.isNotEmpty)
                  Container(
                    margin: const EdgeInsets.only(top: 4),
                    padding:
                        const EdgeInsets.symmetric(horizontal: 6, vertical: 4),
                    decoration: BoxDecoration(
                      color: Colors.blue[50],
                      borderRadius: BorderRadius.circular(4),
                      border: Border.all(color: Colors.blue[200]!),
                    ),
                    child: Row(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Icon(Icons.link, size: 13, color: Colors.blue[700]),
                        const SizedBox(width: 4),
                        Expanded(
                          child: Text.rich(
                            TextSpan(
                              children: [
                                TextSpan(
                                  text:
                                      "${global.language("unit")} (${screenData.unitCount}): ",
                                  style: TextStyle(
                                    fontSize: 11,
                                    fontWeight: FontWeight.bold,
                                    color: Colors.blue[700],
                                  ),
                                ),
                                TextSpan(
                                  text: screenData.allUnitNames,
                                  style: TextStyle(
                                    fontSize: 11,
                                    color: Colors.blue[800],
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
        ],
      ),
    );
  }

  // === ALL LANGUAGE NAMES ===
  Widget _buildNamesSection() {
    final names = screenData.names ?? [];
    if (names.isEmpty) return const SizedBox.shrink();
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _sectionHeader(Icons.translate, global.language("product_name"),
              Colors.brown[700]!),
          const SizedBox(height: 2),
          for (int i = 0; i < names.length; i += 2)
            Padding(
              padding: const EdgeInsets.only(bottom: 1),
              child: Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Expanded(
                    child: _labelValue(
                      (names[i].code ?? '').toUpperCase(),
                      names[i].name.isNotEmpty ? names[i].name : "-",
                    ),
                  ),
                  const SizedBox(width: 8),
                  if (i + 1 < names.length)
                    Expanded(
                      child: _labelValue(
                        (names[i + 1].code ?? '').toUpperCase(),
                        names[i + 1].name.isNotEmpty
                            ? names[i + 1].name
                            : "-",
                      ),
                    )
                  else
                    const Expanded(child: SizedBox()),
                ],
              ),
            ),
        ],
      ),
    );
  }

  // === PRODUCT TYPE DETAIL (item type, food type, material type) ===
  Widget _buildProductTypeDetailSection() {
    String foodTypeName = "-";
    if (screenData.foodtype != null) {
      switch (screenData.foodtype!) {
        case 0:
          foodTypeName = global.language("food");
          break;
        case 1:
          foodTypeName = global.language("drink");
          break;
        case 2:
          foodTypeName = global.language("alcohol");
          break;
        case 3:
          foodTypeName = global.language("other");
          break;
      }
    }

    String materialTypeName = "-";
    if (screenData.materialtype != null) {
      switch (screenData.materialtype!) {
        case 0:
          materialTypeName = global.language("product_general");
          break;
        case 1:
          materialTypeName = global.language("product_material");
          break;
        case 2:
          materialTypeName = global.language("product_semi_finished");
          break;
      }
    }

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _sectionHeader(Icons.inventory_2,
              global.language("product_type"), Colors.teal[700]!),
          const SizedBox(height: 2),
          _row2col(
            global.language("product_type"),
            _productTypeName(screenData.itemtype ?? 0),
            global.language("vat_type"),
            _vatTypeName(screenData.vatcal),
          ),
          _row2col(
            global.language("issumpoint"),
            _pointName(screenData.issumpoint),
            global.language("discount"),
            (screenData.discount?.isNotEmpty == true)
                ? screenData.discount!
                : "-",
          ),
          if (global.posVersion == global.PosVersionEnum.restaurant) ...[
            _row2col(
              global.language("food_type"),
              foodTypeName,
              global.language("material_type"),
              materialTypeName,
            ),
          ],
          _row2col(
            global.language("product_type_code"),
            (screenData.producttype?.code?.isNotEmpty == true)
                ? "${screenData.producttype!.code} | ${global.activeLangName(screenData.producttype!.names ?? [])}"
                : "-",
            global.language("manufacturer"),
            (screenData.manufacturercode?.isNotEmpty == true)
                ? "${screenData.manufacturercode} | ${global.activeLangName(screenData.manufacturernames ?? [])}"
                : "-",
          ),
        ],
      ),
    );
  }

  // === PRICES ===
  Widget _buildPriceSection() {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _sectionHeader(Icons.attach_money, global.language("price"),
              Colors.green[700]!),
          const SizedBox(height: 2),
          // Compact price grid: 2 columns
          for (int i = 0; i < priceList.length; i += 2)
            Padding(
              padding: const EdgeInsets.only(bottom: 1),
              child: Row(
                children: [
                  Expanded(
                    child: _priceItem(
                      priceList[i].names[0].name ?? "",
                      screenData.prices![i].price,
                    ),
                  ),
                  if (i + 1 < priceList.length)
                    Expanded(
                      child: _priceItem(
                        priceList[i + 1].names[0].name ?? "",
                        screenData.prices![i + 1].price,
                      ),
                    )
                  else
                    const Expanded(child: SizedBox()),
                ],
              ),
            ),
        ],
      ),
    );
  }

  Widget _priceItem(String label, double price) {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
      margin: const EdgeInsets.only(right: 4),
      decoration: BoxDecoration(
        color: price > 0 ? Colors.green[50] : Colors.grey[50],
        borderRadius: BorderRadius.circular(4),
      ),
      child: Row(
        children: [
          Expanded(
            child: Text(
              label,
              style: TextStyle(fontSize: 11, color: Colors.grey[700]),
              overflow: TextOverflow.ellipsis,
            ),
          ),
          Text(
            price > 0 ? global.formatNumber(price) : "0.00",
            style: TextStyle(
              fontSize: 12,
              fontWeight: FontWeight.bold,
              color: price > 0 ? Colors.green[700] : Colors.grey[500],
            ),
          ),
        ],
      ),
    );
  }

  // === MASTER DATA ===
  Widget _buildMasterDataSection() {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _sectionHeader(Icons.category_outlined,
              global.language("master_data"), Colors.blue[700]!),
          const SizedBox(height: 2),
          _row2col(
            global.language("brand"),
            _masterVal(screenData.brandcode, screenData.brandnames),
            global.language("category"),
            _masterVal(screenData.categorycode, screenData.categorynames),
          ),
          _row2col(
            global.language("class"),
            _masterVal(screenData.classcode, screenData.classnames),
            global.language("design"),
            _masterVal(screenData.designcode, screenData.designnames),
          ),
          _row2col(
            global.language("grade"),
            _masterVal(screenData.gradecode, screenData.gradenames),
            global.language("model"),
            _masterVal(screenData.modelcode, screenData.modelnames),
          ),
          _row2col(
            global.language("group_main"),
            _masterVal(screenData.groupcode, screenData.groupnames),
            global.language("group_sub1"),
            _masterVal(screenData.groupsubonecode, screenData.groupsubonenames),
          ),
          _row2col(
            global.language("group_sub2"),
            _masterVal(screenData.groupsubtwocode, screenData.groupsubtwonames),
            global.language("pattern"),
            _masterVal(screenData.patterncode, screenData.patternnames),
          ),
        ],
      ),
    );
  }

  String _masterVal(String? code, List<LanguageDataModel>? names) {
    if (code == null || code.isEmpty) return "-";
    return "$code | ${global.activeLangName(names ?? [])}";
  }

  // === SETTINGS (Switches/Flags) ===
  Widget _buildSettingsSection() {
    final List<_FlagItem> flags = [];

    flags.add(_FlagItem(
        global.language("is_discount_point_of_purchase"),
        screenData.isdiscountpointofpurchase == true));

    if (global.posVersion == global.PosVersionEnum.restaurant) {
      flags.add(_FlagItem(
          global.language("alacarte"), screenData.isalacarte == true));
      flags.add(_FlagItem(global.language("is_stock_for_restaurant"),
          screenData.isstockforrestaurant == true));
      flags.add(_FlagItem(global.language("is_split_unit_print"),
          screenData.issplitunitprint == true));
      flags.add(_FlagItem(global.language("is_only_employee"),
          screenData.isonlystaff == true));
      flags.add(_FlagItem(global.language("is_for_restaurant"),
          screenData.restaurant?.isforrestaurant == true));
      flags.add(_FlagItem(global.language("is_for_takeaway"),
          screenData.restaurant?.isfortakeaway == true));
      flags.add(_FlagItem(global.language("is_for_delivery"),
          screenData.restaurant?.isfordelivery == true));
      flags.add(_FlagItem(global.language("is_for_customer"),
          screenData.restaurant?.isforcustomer == true));
      flags.add(_FlagItem(global.language("is_for_customer_preorder"),
          screenData.restaurant?.isforcustomerpreorder == true));
      flags.add(_FlagItem(
          global.language("is_use_alert"), screenData.isalert == true));
    }

    // Only show if there are active flags
    final activeFlags = flags.where((f) => f.isActive).toList();
    if (activeFlags.isEmpty) return const SizedBox.shrink();

    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _sectionHeader(
              Icons.toggle_on, global.language("settings"), Colors.grey[700]!),
          const SizedBox(height: 2),
          Wrap(
            spacing: 4,
            runSpacing: 4,
            children: activeFlags.map((f) {
              return Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
                decoration: BoxDecoration(
                  color: Colors.blue[50],
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: Colors.blue[200]!),
                ),
                child: Text(
                  f.label,
                  style: TextStyle(fontSize: 11, color: Colors.blue[700]),
                ),
              );
            }).toList(),
          ),
          if (screenData.isalert == true &&
              screenData.alertdescription?.isNotEmpty == true)
            Padding(
              padding: const EdgeInsets.only(top: 2),
              child: Text(
                "${global.language("is_use_alert")}: ${screenData.alertdescription}",
                style: TextStyle(fontSize: 11, color: Colors.orange[700]),
              ),
            ),
        ],
      ),
    );
  }

  // === REFERENCE BARCODES ===
  Widget _buildRefBarcodeSection() {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _sectionHeader(Icons.qr_code, global.language("reference_barcode"),
              Colors.indigo[700]!),
          const SizedBox(height: 2),
          ...screenData.refbarcodes!.map((data) {
            final ratioText = (!data.condition)
                ? "${data.qty} ${global.activeLangName(data.itemunitnames)} = 1 ${global.activeLangName(screenData.itemunitnames!)}"
                : "1 ${global.activeLangName(data.itemunitnames)} = ${data.qty} ${global.activeLangName(screenData.itemunitnames!)}";
            return Container(
              margin: const EdgeInsets.only(bottom: 2),
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
              decoration: BoxDecoration(
                color: Colors.indigo[50],
                borderRadius: BorderRadius.circular(4),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Expanded(
                        child: Text(
                          "${data.barcode} | ${global.activeLangName(data.names)} | ${global.activeLangName(data.itemunitnames)}",
                          style: const TextStyle(fontSize: 12),
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                    ],
                  ),
                  Row(
                    children: [
                      Text(
                        "stand: ${data.standvalue}  divide: ${data.dividevalue}",
                        style: TextStyle(fontSize: 10, color: Colors.grey[600]),
                      ),
                      const Spacer(),
                      Text(
                        ratioText,
                        style: TextStyle(
                          fontSize: 11,
                          fontWeight: FontWeight.bold,
                          color: Colors.indigo[700],
                        ),
                      ),
                    ],
                  ),
                ],
              ),
            );
          }),
        ],
      ),
    );
  }

  // === BOM ===
  Widget _buildBomSection() {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _sectionHeader(Icons.account_tree, global.language("product_bom"),
              Colors.orange[700]!),
          const SizedBox(height: 2),
          ...screenData.bom!.map((data) {
            return Container(
              margin: const EdgeInsets.only(bottom: 2),
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
              decoration: BoxDecoration(
                color: Colors.orange[50],
                borderRadius: BorderRadius.circular(4),
              ),
              child: Row(
                children: [
                  Expanded(
                    child: Text(
                      "${data.barcode} | ${global.activeLangName(data.names)} | ${global.activeLangName(data.itemunitnames)}",
                      style: const TextStyle(fontSize: 12),
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  Text(
                    "${data.qty} ${global.activeLangName(data.itemunitnames)}",
                    style: TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.bold,
                      color: Colors.orange[700],
                    ),
                  ),
                ],
              ),
            );
          }),
        ],
      ),
    );
  }

  // === DIMENSIONS ===
  Widget _buildDimensionsSection() {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _sectionHeader(Icons.straighten, global.language("dimensions"),
              Colors.teal[700]!),
          const SizedBox(height: 2),
          ...screenData.dimensions!.asMap().entries.map((entry) {
            final i = entry.key;
            final dim = entry.value;
            return _row1col(
              "${global.language("level_dimension")} ${i + 1}",
              "${global.activeLangName(dim.names ?? [])} ${dim.item?.guidfixed?.isNotEmpty == true ? '| ${global.activeLangName(dim.item!.names!)}' : ''}",
            );
          }),
        ],
      ),
    );
  }

  // === BUSINESS TYPE & BRANCH ===
  Widget _buildBusinessSection() {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _sectionHeader(Icons.business, global.language("business_type"),
              Colors.purple[700]!),
          const SizedBox(height: 2),
          if (screenData.businesstypes?.isNotEmpty == true)
            Wrap(
              spacing: 4,
              runSpacing: 2,
              children: screenData.businesstypes!.map((bt) {
                return Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                  decoration: BoxDecoration(
                    color: Colors.purple[50],
                    borderRadius: BorderRadius.circular(4),
                    border: Border.all(color: Colors.purple[200]!),
                  ),
                  child: Text(
                    "${bt.code} | ${global.activeLangName(bt.names!)}",
                    style: TextStyle(fontSize: 11, color: Colors.purple[700]),
                  ),
                );
              }).toList(),
            ),
          if (screenData.ignorebranches?.isNotEmpty == true) ...[
            const SizedBox(height: 4),
            Text(
              "${global.language("branch")}:",
              style: TextStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.w600,
                  color: Colors.grey[600]),
            ),
            Wrap(
              spacing: 4,
              runSpacing: 2,
              children: screenData.ignorebranches!.map((br) {
                return Container(
                  padding:
                      const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                  decoration: BoxDecoration(
                    color: Colors.red[50],
                    borderRadius: BorderRadius.circular(4),
                    border: Border.all(color: Colors.red[200]!),
                  ),
                  child: Text(
                    "${br.code} | ${global.activeLangName(br.names!)}",
                    style: TextStyle(fontSize: 11, color: Colors.red[700]),
                  ),
                );
              }).toList(),
            ),
          ],
        ],
      ),
    );
  }

  // === FIXED COST ===
  Widget _buildFixedCostSection() {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _sectionHeader(Icons.monetization_on,
              global.language("cost_standard"), Colors.green[700]!),
          const SizedBox(height: 2),
          ...screenData.fixedcost!.map((cost) {
            return _row2col(
              global.language("as_date"),
              cost.effectdate ?? "-",
              global.language("amount"),
              global.formatNumber(cost.amount ?? 0),
            );
          }),
        ],
      ),
    );
  }

  // === TIME FOR SALE ===
  Widget _buildTimeForSaleSection() {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _sectionHeader(Icons.access_time,
              global.language("show_time_for_sale"), Colors.cyan[700]!),
          const SizedBox(height: 2),
          ...screenData.timeforsales!.asMap().entries.map((entry) {
            final i = entry.key;
            final tfs = entry.value;
            final daysText = tfs.daysofweek?.map((day) {
                  switch (day) {
                    case 1:
                      return global.language('monday');
                    case 2:
                      return global.language('tuesday');
                    case 3:
                      return global.language('wendesday');
                    case 4:
                      return global.language('thursday');
                    case 5:
                      return global.language('friday');
                    case 6:
                      return global.language('saturday');
                    case 7:
                      return global.language('sunday');
                    default:
                      return '';
                  }
                }).join(', ') ??
                '';
            return Container(
              margin: const EdgeInsets.only(bottom: 2),
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
              decoration: BoxDecoration(
                color: Colors.cyan[50],
                borderRadius: BorderRadius.circular(4),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    "${global.language("time_for_sale")} ${i + 1}: ${tfs.fromdate ?? ''} ~ ${tfs.todate ?? ''} | ${tfs.fromtime ?? ''} ~ ${tfs.totime ?? ''}",
                    style: const TextStyle(fontSize: 11),
                  ),
                  if (daysText.isNotEmpty)
                    Text(
                      daysText,
                      style: TextStyle(fontSize: 10, color: Colors.cyan[700]),
                    ),
                ],
              ),
            );
          }),
        ],
      ),
    );
  }

  // === ORDER TYPES ===
  Widget _buildOrderTypesSection() {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _sectionHeader(Icons.shopping_cart, global.language("order_type"),
              Colors.orange[700]!),
          const SizedBox(height: 2),
          ...screenData.ordertypes!.map((data) {
            return Container(
              margin: const EdgeInsets.only(bottom: 2),
              padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
              decoration: BoxDecoration(
                color: Colors.orange[50],
                borderRadius: BorderRadius.circular(4),
              ),
              child: Row(
                children: [
                  Expanded(
                    child: Text(
                      (data.code.isEmpty)
                          ? "-"
                          : "${data.code} | ${global.activeLangName(data.names)}",
                      style: const TextStyle(fontSize: 12),
                      overflow: TextOverflow.ellipsis,
                    ),
                  ),
                  Text(
                    "${global.language("charge_price")}: ${data.price}",
                    style: TextStyle(
                      fontSize: 11,
                      fontWeight: FontWeight.bold,
                      color: Colors.orange[700],
                    ),
                  ),
                ],
              ),
            );
          }),
        ],
      ),
    );
  }

  // === OPTIONS ===
  Widget _buildOptionsSection() {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _sectionHeader(Icons.tune, global.language("options"),
              Colors.deepPurple[700]!),
          const SizedBox(height: 2),
          ...screenData.options!.asMap().entries.map((entry) {
            final optIdx = entry.key;
            final opt = entry.value;
            final typeLabel = (opt.choicetype == 0)
                ? global.language("product_option_choice_type_multi")
                : global.language("product_option_choice_type_single");
            return Container(
              margin: const EdgeInsets.only(bottom: 4),
              padding: const EdgeInsets.all(6),
              decoration: BoxDecoration(
                color: Colors.deepPurple[50],
                borderRadius: BorderRadius.circular(4),
                border: Border.all(color: Colors.deepPurple[100]!),
              ),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Row(
                    children: [
                      Text(
                        "${global.language("option")} ${optIdx + 1}: ${global.activeLangName(opt.names)}",
                        style: const TextStyle(
                            fontSize: 12, fontWeight: FontWeight.w600),
                      ),
                      const SizedBox(width: 6),
                      Container(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 6, vertical: 1),
                        decoration: BoxDecoration(
                          color: (opt.choicetype == 0)
                              ? Colors.blue[100]
                              : Colors.green[100],
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Text(
                          typeLabel,
                          style: TextStyle(
                            fontSize: 10,
                            color: (opt.choicetype == 0)
                                ? Colors.blue[700]
                                : Colors.green[700],
                          ),
                        ),
                      ),
                      if (opt.choicetype == 0) ...[
                        const SizedBox(width: 6),
                        Text(
                          "min:${opt.minselect} max:${opt.maxselect}",
                          style:
                              TextStyle(fontSize: 10, color: Colors.grey[600]),
                        ),
                      ],
                    ],
                  ),
                  if (opt.choices.isNotEmpty)
                    ...opt.choices.asMap().entries.map((cEntry) {
                      final cIdx = cEntry.key;
                      final choice = cEntry.value;
                      return Padding(
                        padding: const EdgeInsets.only(left: 12, top: 2),
                        child: Row(
                          children: [
                            Text(
                              "${cIdx + 1}. ${global.activeLangName(choice.names)}",
                              style: const TextStyle(fontSize: 11),
                            ),
                            const SizedBox(width: 4),
                            Text(
                              "${global.language("price")}: ${choice.price}",
                              style: TextStyle(
                                fontSize: 10,
                                fontWeight: FontWeight.bold,
                                color: Colors.green[700],
                              ),
                            ),
                            if (choice.isstock) ...[
                              const SizedBox(width: 4),
                              Text(
                                "[${choice.refbarcode} x${choice.qty}]",
                                style: TextStyle(
                                    fontSize: 10, color: Colors.grey[600]),
                              ),
                            ],
                          ],
                        ),
                      );
                    }),
                ],
              ),
            );
          }),
        ],
      ),
    );
  }

  // === DESCRIPTION ===
  Widget _buildDescriptionSection() {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          _sectionHeader(Icons.description, global.language("disciption"),
              Colors.grey[700]!),
          const SizedBox(height: 2),
          Container(
            width: double.infinity,
            padding: const EdgeInsets.all(6),
            decoration: BoxDecoration(
              color: Colors.grey[50],
              borderRadius: BorderRadius.circular(4),
            ),
            child: Text(
              screenData.description!,
              style: const TextStyle(fontSize: 12),
            ),
          ),
        ],
      ),
    );
  }

  // ===== HELPER WIDGETS =====

  /// Section header with icon
  Widget _sectionHeader(IconData icon, String title, Color color) {
    return Padding(
      padding: const EdgeInsets.only(top: 4, bottom: 2),
      child: Row(
        children: [
          Icon(icon, size: 14, color: color),
          const SizedBox(width: 4),
          Text(
            title,
            style: TextStyle(
              fontSize: 13,
              fontWeight: FontWeight.bold,
              color: color,
            ),
          ),
        ],
      ),
    );
  }

  /// Two-column label:value row (compact)
  Widget _row2col(
      String label1, String value1, String label2, String value2) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 2),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Expanded(
            child: _labelValue(label1, value1),
          ),
          const SizedBox(width: 8),
          Expanded(
            child: _labelValue(label2, value2),
          ),
        ],
      ),
    );
  }

  /// Single-column label:value row
  Widget _row1col(String label, String value) {
    return Padding(
      padding: const EdgeInsets.only(bottom: 2),
      child: _labelValue(label, value),
    );
  }

  /// Label: Value inline
  Widget _labelValue(String label, String value) {
    return RichText(
      overflow: TextOverflow.ellipsis,
      maxLines: 2,
      text: TextSpan(
        style: const TextStyle(fontSize: 11, color: Colors.black),
        children: [
          TextSpan(
            text: "$label: ",
            style: TextStyle(
              fontWeight: FontWeight.w600,
              color: Colors.grey[600],
              fontSize: 11,
            ),
          ),
          TextSpan(
            text: value,
            style: const TextStyle(fontSize: 12),
          ),
        ],
      ),
    );
  }

  /// Parse color from screenData
  Color _parseColor() {
    if (screenData.colorselect != null &&
        screenData.colorselect!.isNotEmpty) {
      try {
        return Color(int.parse(screenData.colorselect!));
      } catch (_) {}
    }
    return Colors.white;
  }
}

class _FlagItem {
  final String label;
  final bool isActive;
  const _FlagItem(this.label, this.isActive);
}
