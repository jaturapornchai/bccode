import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../models/product_search_model.dart';
import '../models/product_transaction_model.dart';
import '../models/cart_item_model.dart';
import '../models/cart_model.dart';
import '../models/cart_system_type.dart';
import '../services/pgsql_product_service.dart';
import '../services/mongodb_cart_service.dart';
import 'warehouse_selector_dialog.dart';
import '../../../utils/logger/app_logger.dart';
import '../../../global.dart' as global;

/// Dialog แสดงรายละเอียดสินค้าแบบละเอียด
/// - รายการซื้อล่าสุด 5 รายการ
/// - รายการขายล่าสุด 5 รายการ
/// - ยอดคงเหลือตาม warehouse แบบละเอียด
/// - ยอดคงเหลือตาม location แบบละเอียด
class ProductDetailDialog extends StatefulWidget {
  final ProductSearchModel product;
  final String shopId;
  final String? activeCartId; // Cart ID ที่กำลังใช้งาน

  const ProductDetailDialog({
    super.key,
    required this.product,
    required this.shopId,
    this.activeCartId,
  });

  /// แสดง Dialog ขนาดที่เหมาะสม
  static Future<bool?> show({
    required BuildContext context,
    required ProductSearchModel product,
    required String shopId,
    String? activeCartId,
  }) {
    return showDialog<bool>(
      context: context,
      builder: (context) {
        final screenWidth = MediaQuery.of(context).size.width;
        final screenHeight = MediaQuery.of(context).size.height;

        // คำนวณขนาดที่เหมาะสม
        final dialogWidth = screenWidth > 1200
            ? 1100.0 // จอใหญ่ ให้ max 1100px
            : screenWidth * 0.85; // จอเล็ก ให้ 85%

        return Dialog(
          insetPadding: EdgeInsets.symmetric(
            horizontal: (screenWidth - dialogWidth) / 2,
            vertical: 40, // ความสูงแบบ auto พร้อม margin บนล่าง
          ),
          child: ConstrainedBox(
            constraints: BoxConstraints(
              maxHeight: screenHeight - 80, // ไม่เกินความสูงจอ - margin
            ),
            child: ProductDetailDialog(
              product: product,
              shopId: shopId,
              activeCartId: activeCartId,
            ),
          ),
        );
      },
    );
  }

  @override
  State<ProductDetailDialog> createState() => _ProductDetailDialogState();
}

class _ProductDetailDialogState extends State<ProductDetailDialog> {
  final _service = PgSQLProductService();
  final _cartService = MongoDBCartService();

  // Form controllers
  final _quantityController = TextEditingController(text: '1');
  final _priceController = TextEditingController();
  final _discountController = TextEditingController(text: '0');
  final _remarkController = TextEditingController();
  final _formKey = GlobalKey<FormState>();

  bool _isLoading = true;
  bool _isSavingToCart = false;
  List<ProductTransactionModel> _recentPurchases = [];
  List<ProductTransactionModel> _recentSales = [];
  Map<String, ProductWarehouseBalanceModel> _detailedBalances = {};
  List<ProductSearchModel> _allUnits = []; // หน่วยนับทั้งหมดสำหรับ auto packing
  double _totalBalance = 0.0; // ยอดรวมทั้งหมด

  // Selected unit for cart
  ProductSearchModel? _selectedUnit;

  // ข้อมูลคลังและที่เก็บที่เลือกสำหรับสินค้านี้
  String? _selectedWarehouseId;
  String? _selectedWarehouseName;
  String? _selectedLocationId;
  String? _selectedLocationName;

  @override
  void initState() {
    super.initState();
    _loadData();
  }

  @override
  void dispose() {
    _quantityController.dispose();
    _priceController.dispose();
    _discountController.dispose();
    _remarkController.dispose();
    super.dispose();
  }

  Future<void> _loadData() async {
    setState(() => _isLoading = true);

    try {
      // โหลดข้อมูลพร้อมกัน (parallel)
      final results = await Future.wait([
        _service.getRecentPurchases(
          itemCode: widget.product.itemCode,
          shopId: widget.shopId,
          limit: 5,
        ),
        _service.getRecentSales(
          itemCode: widget.product.itemCode,
          shopId: widget.shopId,
          limit: 5,
        ),
        _service.getDetailedBalances(
          itemCode: widget.product.itemCode,
          shopId: widget.shopId,
        ),
        _service.getAllUnitsForProduct(widget.product.itemCode, widget.shopId),
      ]);

      // คำนวณยอดรวมทั้งหมด
      final detailedBalances =
          results[2] as Map<String, ProductWarehouseBalanceModel>;
      double total = 0.0;
      for (final warehouse in detailedBalances.values) {
        total += warehouse.totalBalance;
      }

      setState(() {
        _recentPurchases = results[0] as List<ProductTransactionModel>;
        _recentSales = results[1] as List<ProductTransactionModel>;
        _detailedBalances = detailedBalances;
        _allUnits = results[3] as List<ProductSearchModel>;
        _totalBalance = total;
        _isLoading = false;

        // ตั้งค่าเริ่มต้นสำหรับ form
        if (_allUnits.isNotEmpty) {
          // หา unit ที่ตรงกับ product จาก _allUnits (ใช้ barcode เป็น key)
          _selectedUnit = _allUnits.firstWhere(
            (u) => u.barcode == widget.product.barcode,
            orElse: () => _allUnits.first,
          );
          _priceController.text = global.formatNumberRemoveRightZero(
            _selectedUnit?.price1 ?? widget.product.price1,
          );
        }
      });
    } catch (e) {
      AppLogger.error('Error loading product detail: $e');
      setState(() => _isLoading = false);
    }
  }

  @override
  Widget build(BuildContext context) {
    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(16),
      ),
      child: Column(
        children: [
          // Header
          _buildHeader(),

          // Content
          Expanded(
            child: _isLoading
                ? const Center(child: CircularProgressIndicator())
                : _buildContent(),
          ),
        ],
      ),
    );
  }

  Widget _buildHeader() {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: const BoxDecoration(
        gradient: LinearGradient(
          colors: [Color(0xFFEE4D2D), Color(0xFFFF6347)],
          begin: Alignment.topLeft,
          end: Alignment.bottomRight,
        ),
        borderRadius: BorderRadius.vertical(top: Radius.circular(16)),
      ),
      child: Row(
        children: [
          // Product Image or Icon - แสดงรูปถ้ามี ไม่มีให้แสดง icon
          if (widget.product.imageuri != null &&
              widget.product.imageuri!.isNotEmpty)
            Container(
              width: 64,
              height: 64,
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(8),
                border: Border.all(color: Colors.white, width: 2),
              ),
              child: ClipRRect(
                borderRadius: BorderRadius.circular(6),
                child: Image.network(
                  widget.product.imageuri!,
                  fit: BoxFit.cover,
                  loadingBuilder: (context, child, loadingProgress) {
                    if (loadingProgress == null) return child;
                    return Center(
                      child: CircularProgressIndicator(
                        value: loadingProgress.expectedTotalBytes != null
                            ? loadingProgress.cumulativeBytesLoaded /
                                loadingProgress.expectedTotalBytes!
                            : null,
                        strokeWidth: 2,
                        color: const Color(0xFFEE4D2D),
                      ),
                    );
                  },
                  errorBuilder: (context, error, stackTrace) {
                    return Container(
                      padding: const EdgeInsets.all(12),
                      color: Colors.white,
                      child: const Icon(
                        Icons.inventory_2,
                        color: Color(0xFFEE4D2D),
                        size: 32,
                      ),
                    );
                  },
                ),
              ),
            )
          else
            Container(
              padding: const EdgeInsets.all(12),
              decoration: BoxDecoration(
                color: Colors.white,
                borderRadius: BorderRadius.circular(8),
              ),
              child: const Icon(
                Icons.inventory_2,
                color: Color(0xFFEE4D2D),
                size: 32,
              ),
            ),
          const SizedBox(width: 16),

          // Product Info
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(
                  widget.product.name0,
                  style: const TextStyle(
                    fontSize: 18,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                  ),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
                SizedBox(height: 4),
                Text(
                  '${global.language("code")}: ${widget.product.itemCode}',
                  style: const TextStyle(fontSize: 14, color: Colors.white70),
                ),
                Text(
                  '${global.language("barcode")}: ${widget.product.barcode}',
                  style: const TextStyle(fontSize: 12, color: Colors.white60),
                ),
              ],
            ),
          ),

          // Close Button
          IconButton(
            onPressed: () => Navigator.of(context).pop(),
            icon: Icon(Icons.close, color: Colors.white),
            tooltip: global.language("close"),
          ),
        ],
      ),
    );
  }

  Widget _buildContent() {
    return SingleChildScrollView(
      padding: const EdgeInsets.all(12),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // แถวที่ 1: ข้อมูลพื้นฐาน และ Barcode
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // ข้อมูลพื้นฐาน
              SizedBox(width: 200, child: _buildBasicInfo()),
              const SizedBox(width: 12),

              // รายการ Barcode ทั้งหมด
              Expanded(child: _buildAllBarcodes()),
            ],
          ),
          const SizedBox(height: 12),

          // ยอดคงเหลือแบบละเอียด
          _buildDetailedBalances(),
          const SizedBox(height: 12),

          // เพิ่มสินค้าเข้าตะกร้า
          _buildCartForm(),
          const SizedBox(height: 12),

          // รายการซื้อ/ขายล่าสุด
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // รายการซื้อล่าสุด
              Expanded(child: _buildRecentPurchases()),
              const SizedBox(width: 12),

              // รายการขายล่าสุด
              Expanded(child: _buildRecentSales()),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildBasicInfo() {
    return Card(
      child: Padding(
        padding: EdgeInsets.all(10),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.info_outline, size: 16, color: Colors.blue[700]),
                SizedBox(width: 6),
                Text(
                  global.language("basic_info"),
                  style: const TextStyle(fontSize: 13, fontWeight: FontWeight.bold),
                ),
              ],
            ),
            Divider(height: 12),
            _buildInfoRow(global.language("item_code"), widget.product.itemCode),
            _buildInfoRow(global.language("unit"), widget.product.unitName),
            _buildInfoRow(
              global.language("ratio"),
              '${widget.product.unitStand}:${widget.product.unitDive}',
            ),
            _buildInfoRow(
              global.language("selling_price"),
              global.formatPrice(widget.product.price1),
            ),
          ],
        ),
      ),
    );
  }

  /// แสดงรายการ Barcode ทั้งหมดของสินค้า
  Widget _buildAllBarcodes() {
    return Card(
      child: Theme(
        data: Theme.of(context).copyWith(dividerColor: Colors.transparent),
        child: ExpansionTile(
          initiallyExpanded: true,
          tilePadding: const EdgeInsets.symmetric(horizontal: 10, vertical: 2),
          childrenPadding: const EdgeInsets.only(
            left: 10,
            right: 10,
            bottom: 10,
          ),
          leading: Icon(Icons.qr_code, color: Colors.orange[700], size: 18),
          title: Text(
            'Barcode (${_allUnits.length})',
            style: const TextStyle(fontSize: 13, fontWeight: FontWeight.bold),
          ),
          children: [
            if (_allUnits.isEmpty)
              Padding(
                padding: EdgeInsets.all(10),
                child: Center(
                  child: Text(
                    global.language("no_barcode_data"),
                    style: TextStyle(color: Colors.grey[600], fontSize: 12),
                  ),
                ),
              )
            else
              ..._allUnits.map((unit) {
                final ratio = unit.unitStand / unit.unitDive;
                final isBasicUnit = ratio == 1.0;

                return Container(
                  margin: const EdgeInsets.only(bottom: 6),
                  padding: const EdgeInsets.all(8),
                  decoration: BoxDecoration(
                    color: isBasicUnit ? Colors.orange[50] : Colors.grey[50],
                    borderRadius: BorderRadius.circular(6),
                    border: Border.all(
                      color: isBasicUnit
                          ? Colors.orange[200]!
                          : Colors.grey[300]!,
                    ),
                  ),
                  child: Row(
                    children: [
                      // Barcode + Badge
                      Expanded(
                        flex: 3,
                        child: Row(
                          children: [
                            Icon(
                              Icons.qr_code_2,
                              size: 14,
                              color: isBasicUnit
                                  ? Colors.orange[700]
                                  : Colors.grey[700],
                            ),
                            const SizedBox(width: 6),
                            Expanded(
                              child: Text(
                                unit.barcode,
                                style: TextStyle(
                                  fontSize: 12,
                                  fontWeight: FontWeight.bold,
                                  color: isBasicUnit
                                      ? Colors.orange[900]
                                      : Colors.grey[900],
                                ),
                                maxLines: 1,
                                overflow: TextOverflow.ellipsis,
                              ),
                            ),
                            if (isBasicUnit) ...[
                              SizedBox(width: 4),
                              Container(
                                padding: const EdgeInsets.symmetric(
                                  horizontal: 6,
                                  vertical: 2,
                                ),
                                decoration: BoxDecoration(
                                  color: Colors.orange[700],
                                  borderRadius: BorderRadius.circular(10),
                                ),
                                child: Text(
                                  global.language("base_unit"),
                                  style: const TextStyle(
                                    fontSize: 9,
                                    color: Colors.white,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ),
                            ],
                          ],
                        ),
                      ),
                      const SizedBox(width: 8),

                      // หน่วยนับ
                      Expanded(
                        flex: 2,
                        child: Text(
                          unit.unitName.isNotEmpty
                              ? unit.unitName
                              : unit.unitCode,
                          style: const TextStyle(
                            fontSize: 11,
                            fontWeight: FontWeight.w600,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),

                      // อัตราส่วน
                      Expanded(
                        flex: 1,
                        child: Text(
                          '${global.formatNumberRemoveRightZero(unit.unitStand)}:${global.formatNumberRemoveRightZero(unit.unitDive)}',
                          style: TextStyle(
                            fontSize: 11,
                            fontWeight: FontWeight.w600,
                            color: isBasicUnit
                                ? Colors.orange[800]
                                : Colors.grey[800],
                          ),
                          textAlign: TextAlign.center,
                        ),
                      ),

                      // ราคา
                      Expanded(
                        flex: 2,
                        child: Text(
                          global.formatPrice(unit.price1),
                          style: TextStyle(
                            fontSize: 11,
                            fontWeight: FontWeight.bold,
                            color: Colors.green[700],
                          ),
                          textAlign: TextAlign.right,
                        ),
                      ),
                    ],
                  ),
                );
              }),
          ],
        ),
      ),
    );
  }

  Widget _buildInfoRow(String label, String value) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 3),
      child: Row(
        children: [
          SizedBox(
            width: 80,
            child: Text(
              label,
              style: TextStyle(fontSize: 11, color: Colors.grey[600]),
            ),
          ),
          Expanded(
            child: Text(
              value,
              style: const TextStyle(fontSize: 11, fontWeight: FontWeight.w500),
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildDetailedBalances() {
    return Card(
      child: Padding(
        padding: EdgeInsets.all(10),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.warehouse, color: Colors.blue[700], size: 18),
                SizedBox(width: 6),
                Text(
                  global.language("detailed_balance"),
                  style: const TextStyle(fontSize: 13, fontWeight: FontWeight.bold),
                ),
              ],
            ),
            const Divider(height: 12),

            // แสดงยอดรวมทั้งหมด
            if (_totalBalance > 0) ...[
              Container(
                padding: const EdgeInsets.all(8),
                decoration: BoxDecoration(
                  color: Colors.purple[50],
                  borderRadius: BorderRadius.circular(6),
                  border: Border.all(color: Colors.purple[200]!),
                ),
                child: Row(
                  children: [
                    Icon(Icons.inventory, color: Colors.purple[700], size: 16),
                    SizedBox(width: 6),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Text(
                            global.language("total_balance"),
                            style: TextStyle(
                              fontSize: 11,
                              fontWeight: FontWeight.bold,
                              color: Colors.purple[900],
                            ),
                          ),
                          const SizedBox(height: 2),
                          Text(
                            _formatBalance(_totalBalance),
                            style: TextStyle(
                              fontSize: 10,
                              color: Colors.purple[800],
                              height: 1.3,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 8),
            ],

            if (_detailedBalances.isEmpty)
              Padding(
                padding: EdgeInsets.all(16),
                child: Center(
                  child: Text(
                    global.language("no_balance_data"),
                    style: TextStyle(color: Colors.grey[600]),
                  ),
                ),
              )
            else
              ..._detailedBalances.entries.map((entry) {
                final warehouse = entry.value;
                return _buildWarehouseCard(warehouse);
              }),
          ],
        ),
      ),
    );
  }

  Widget _buildWarehouseCard(ProductWarehouseBalanceModel warehouse) {
    return Container(
      margin: const EdgeInsets.only(bottom: 8),
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: Colors.blue[50],
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: Colors.blue[200]!),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Warehouse Header
          Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Icon(Icons.store, size: 16, color: Colors.blue[700]),
              const SizedBox(width: 6),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // ชื่อคลัง + รหัสคลัง
                    Text(
                      '${warehouse.warehouseName}${warehouse.warehouseId.isNotEmpty ? " (${warehouse.warehouseId})" : ""}',
                      style: TextStyle(
                        fontSize: 12,
                        fontWeight: FontWeight.bold,
                        color: Colors.blue[900],
                      ),
                    ),
                    const SizedBox(height: 2),
                    // ยอดคงเหลือ
                    Text(
                      _formatBalance(warehouse.totalBalance),
                      style: TextStyle(
                        fontSize: 10,
                        color: Colors.blue[800],
                        fontWeight: FontWeight.w600,
                        height: 1.3,
                      ),
                    ),
                  ],
                ),
              ),
            ],
          ),

          // Locations
          if (warehouse.locationBalances.isNotEmpty) ...[
            const SizedBox(height: 6),
            const Divider(height: 8),
            ...warehouse.locations.map((location) {
              return Container(
                margin: const EdgeInsets.only(bottom: 4, top: 4),
                padding: const EdgeInsets.all(6),
                decoration: BoxDecoration(
                  color: Colors.green[50],
                  borderRadius: BorderRadius.circular(4),
                  border: Border.all(color: Colors.green[200]!),
                ),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Icon(Icons.location_on, size: 14, color: Colors.green[700]),
                    const SizedBox(width: 6),
                    Expanded(
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          // ชื่อ location + รหัส
                          Text(
                            '${location.locationName}${location.locationId.isNotEmpty ? " (${location.locationId})" : ""}',
                            style: TextStyle(
                              fontSize: 11,
                              fontWeight: FontWeight.w600,
                              color: Colors.grey[800],
                            ),
                          ),
                          const SizedBox(height: 2),
                          // ยอดคงเหลือ
                          Text(
                            _formatBalance(location.balance),
                            style: TextStyle(
                              fontSize: 10,
                              color: Colors.green[900],
                              fontWeight: FontWeight.w500,
                              height: 1.3,
                            ),
                          ),
                        ],
                      ),
                    ),
                  ],
                ),
              );
            }),
          ],
        ],
      ),
    );
  }

  Widget _buildRecentPurchases() {
    return Card(
      child: Padding(
        padding: EdgeInsets.all(10),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.shopping_cart, color: Colors.green[700], size: 16),
                SizedBox(width: 6),
                Text(
                  global.language("recent_purchases"),
                  style: const TextStyle(fontSize: 13, fontWeight: FontWeight.bold),
                ),
              ],
            ),
            const Divider(height: 12),

            if (_recentPurchases.isEmpty)
              Padding(
                padding: EdgeInsets.all(10),
                child: Center(
                  child: Text(
                    global.language("no_purchase_records"),
                    style: TextStyle(color: Colors.grey[600], fontSize: 11),
                  ),
                ),
              )
            else
              ..._recentPurchases.map((transaction) {
                return _buildTransactionCard(transaction, isPurchase: true);
              }),
          ],
        ),
      ),
    );
  }

  Widget _buildRecentSales() {
    return Card(
      child: Padding(
        padding: EdgeInsets.all(10),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Row(
              children: [
                Icon(Icons.sell, color: Colors.orange[700], size: 16),
                SizedBox(width: 6),
                Text(
                  global.language("recent_sales"),
                  style: const TextStyle(fontSize: 13, fontWeight: FontWeight.bold),
                ),
              ],
            ),
            const Divider(height: 12),

            if (_recentSales.isEmpty)
              Padding(
                padding: EdgeInsets.all(10),
                child: Center(
                  child: Text(
                    global.language("no_sales_records"),
                    style: TextStyle(color: Colors.grey[600], fontSize: 11),
                  ),
                ),
              )
            else
              ..._recentSales.map((transaction) {
                return _buildTransactionCard(transaction, isPurchase: false);
              }),
          ],
        ),
      ),
    );
  }

  Widget _buildTransactionCard(
    ProductTransactionModel transaction, {
    required bool isPurchase,
  }) {
    final color = isPurchase ? Colors.green : Colors.orange;

    return Container(
      margin: const EdgeInsets.only(bottom: 6),
      padding: const EdgeInsets.all(8),
      decoration: BoxDecoration(
        color: color[50],
        borderRadius: BorderRadius.circular(6),
        border: Border.all(color: color[200]!),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Doc Code & Date
          Row(
            children: [
              Expanded(
                child: Text(
                  transaction.docCode,
                  style: TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.bold,
                    color: color[900],
                  ),
                ),
              ),
              Text(
                global.formatThaiDateTime(
                  dateTime: transaction.docDate,
                  showTime: true,
                ),
                style: TextStyle(fontSize: 10, color: Colors.grey[600]),
              ),
            ],
          ),
          const SizedBox(height: 4),

          // Qty & Unit
          Row(
            children: [
              Icon(Icons.inventory_2_outlined, size: 12, color: color[700]),
              SizedBox(width: 4),
              Text(
                '${global.language("qty")}: ${global.formatNumberRemoveRightZero(transaction.qty)} ${transaction.unitName}',
                style: TextStyle(fontSize: 10, color: Colors.grey[800]),
              ),
            ],
          ),
          const SizedBox(height: 2),

          // Price per unit
          Row(
            children: [
              Icon(Icons.attach_money, size: 12, color: color[700]),
              SizedBox(width: 4),
              Text(
                '${global.language("price_per_unit")}: ${global.formatPrice(transaction.price)}',
                style: TextStyle(fontSize: 10, color: Colors.grey[700]),
              ),
            ],
          ),
          const SizedBox(height: 2),

          // Total Amount
          Row(
            children: [
              Icon(Icons.payments_outlined, size: 12, color: color[700]),
              SizedBox(width: 4),
              Expanded(
                child: Text(
                  '${global.language("grand_total")}: ',
                  style: TextStyle(fontSize: 10, color: Colors.grey[700]),
                ),
              ),
              Text(
                global.formatPrice(transaction.amount),
                style: TextStyle(
                  fontSize: 11,
                  fontWeight: FontWeight.bold,
                  color: color[900],
                ),
              ),
            ],
          ),

          // Document Type (if available)
          if (transaction.docTypeName.isNotEmpty) ...[
            const SizedBox(height: 2),
            Row(
              children: [
                Icon(
                  Icons.description_outlined,
                  size: 12,
                  color: Colors.grey[600],
                ),
                const SizedBox(width: 4),
                Text(
                  transaction.docTypeName,
                  style: TextStyle(fontSize: 10, color: Colors.grey[600]),
                ),
              ],
            ),
          ],

          // Customer Name (if available)
          if (transaction.customerName != null &&
              transaction.customerName!.isNotEmpty) ...[
            const SizedBox(height: 2),
            Row(
              children: [
                Icon(Icons.person_outline, size: 12, color: Colors.grey[600]),
                const SizedBox(width: 4),
                Expanded(
                  child: Text(
                    transaction.customerName!,
                    style: TextStyle(fontSize: 10, color: Colors.grey[600]),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ],
            ),
          ],

          // Remark (if available)
          if (transaction.remark != null && transaction.remark!.isNotEmpty) ...[
            const SizedBox(height: 2),
            Row(
              children: [
                Icon(Icons.note_outlined, size: 12, color: Colors.grey[600]),
                const SizedBox(width: 4),
                Expanded(
                  child: Text(
                    transaction.remark!,
                    style: TextStyle(fontSize: 10, color: Colors.grey[600]),
                    maxLines: 1,
                    overflow: TextOverflow.ellipsis,
                  ),
                ),
              ],
            ),
          ],
        ],
      ),
    );
  }

  /// Format ยอดคงเหลือแบบ 1:1 และ auto packing
  /// แสดง 2 บรรทัด: บรรทัดแรก 1:1, บรรทัดสอง auto packing
  String _formatBalance(double balance) {
    if (balance <= 0) return '0';

    // หาหน่วยฐาน (1:1)
    final baseUnit = _allUnits.firstWhere(
      (u) => u.unitStand == 1 && u.unitDive == 1,
      orElse: () => widget.product,
    );

    final baseUnitName = baseUnit.unitName.isNotEmpty
        ? baseUnit.unitName
        : baseUnit.unitCode;

    // บรรทัดที่ 1: แสดงยอดแบบ 1:1
    final line1 =
        '${global.formatNumberRemoveRightZero(balance)} $baseUnitName';

    // บรรทัดที่ 2: แสดงยอดแบบ auto packing
    final line2 = _formatAutoPacking(balance);

    if (line1 == line2) {
      return line1; // ถ้าเหมือนกัน แสดงบรรทัดเดียว
    }

    return '$line1\n($line2)';
  }

  /// Format ยอดคงเหลือแบบ auto packing
  String _formatAutoPacking(double balance) {
    if (balance <= 0) return '0';
    if (_allUnits.isEmpty) {
      return global.formatNumberRemoveRightZero(balance);
    }

    // หาหน่วยฐาน (1:1)
    final baseUnit = _allUnits.firstWhere(
      (u) => u.unitStand == 1 && u.unitDive == 1,
      orElse: () => _allUnits.first,
    );
    final baseUnitName = baseUnit.unitName.isNotEmpty
        ? baseUnit.unitName
        : baseUnit.unitCode;

    // กรอง units ที่มี ratio > 1 และเรียงจากมาก → น้อย
    final packingUnits =
        _allUnits.where((u) {
          final ratio = u.unitStand / u.unitDive;
          return ratio > 1.0;
        }).toList()..sort((a, b) {
          final ratioA = a.unitStand / a.unitDive;
          final ratioB = b.unitStand / b.unitDive;
          return ratioB.compareTo(ratioA); // มาก → น้อย
        });

    // ถ้าไม่มีหน่วย packing ให้แสดงแบบปกติ
    if (packingUnits.isEmpty) {
      return '${global.formatNumberRemoveRightZero(balance)} $baseUnitName';
    }

    // คำนวณแบบ Greedy
    double remainingBalance = balance;
    final List<String> parts = [];
    final Set<String> usedUnitNames = {}; // ตรวจสอบชื่อหน่วยซ้ำ

    for (final unit in packingUnits) {
      final ratio = unit.unitStand / unit.unitDive;
      final count = (remainingBalance / ratio).floor();

      if (count > 0) {
        // ตรวจสอบชื่อหน่วยซ้ำ
        String displayUnitName = unit.unitName.isNotEmpty
            ? unit.unitName
            : unit.unitCode;

        // ถ้าชื่อซ้ำกับที่ใช้ไปแล้ว หรือซ้ำกับ base unit → ใช้ "แพ็ค"
        if (usedUnitNames.contains(displayUnitName) ||
            displayUnitName == baseUnitName) {
          displayUnitName = global.language("pack");
        }

        usedUnitNames.add(displayUnitName);
        parts.add('$count $displayUnitName');
        remainingBalance -= count * ratio;
      }
    }

    // เศษที่เหลือ (ถ้ามี)
    if (remainingBalance > 0.001) {
      String remainderDisplay;

      // ตรวจสอบว่าเป็นจำนวนเต็มหรือมีทศนิยม
      final floorValue = remainingBalance.floor().toDouble();
      if ((remainingBalance - floorValue).abs() < 0.0001) {
        remainderDisplay = floorValue.toInt().toString();
      } else {
        remainderDisplay = remainingBalance
            .toStringAsFixed(6)
            .replaceAll(RegExp(r'0+$'), '')
            .replaceAll(RegExp(r'\.$'), '');
      }

      // ถ้าหน่วยเศษซ้ำกับหน่วยที่ใช้ไปแล้ว → ใช้ "อัน-ชิ้น"
      final remainderUnitName = usedUnitNames.contains(baseUnitName)
          ? global.language("piece")
          : baseUnitName;

      parts.add('$remainderDisplay $remainderUnitName');
    }

    // สร้าง output string
    if (parts.isEmpty) {
      return '${global.formatNumberRemoveRightZero(balance)} $baseUnitName';
    }

    return parts.join(' x ');
  }

  /// เปิด Dialog เลือกคลัง/ที่เก็บ
  Future<void> _selectWarehouse() async {
    final result = await showDialog<Map<String, dynamic>>(
      context: context,
      builder: (context) => WarehouseSelectorDialog(
        shopId: widget.shopId,
        currentWarehouseId: _selectedWarehouseId,
        currentLocationId: _selectedLocationId,
      ),
    );

    if (result != null && mounted) {
      final warehouse = result['warehouse'];
      final location = result['location'];

      setState(() {
        _selectedWarehouseId = warehouse.guidFixed;
        _selectedWarehouseName = warehouse.displayName;
        _selectedLocationId = location?.guidFixed;
        _selectedLocationName = location?.displayName;
      });

      AppLogger.info(
        '✅ เลือกคลัง: $_selectedWarehouseName${location != null ? ', ที่เก็บ: ${location.displayName}' : ''}',
      );
    }
  }

  /// Form สำหรับเพิ่มสินค้าเข้าตะกร้า
  Widget _buildCartForm() {
    return Card(
      child: Padding(
        padding: EdgeInsets.all(10),
        child: Form(
          key: _formKey,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Icon(
                    Icons.add_shopping_cart,
                    color: Colors.blue[700],
                    size: 18,
                  ),
                  SizedBox(width: 6),
                  Text(
                    global.language("add_product_to_cart"),
                    style: const TextStyle(fontSize: 13, fontWeight: FontWeight.bold),
                  ),
                ],
              ),
              const Divider(height: 12),

              // ปุ่มเลือกคลัง/ที่เก็บ (แบบกระทัดรัด)
              InkWell(
                onTap: _selectWarehouse,
                borderRadius: BorderRadius.circular(8),
                child: Container(
                  padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
                  decoration: BoxDecoration(
                    color: Colors.purple[50],
                    border: Border.all(color: Colors.purple[300]!),
                    borderRadius: BorderRadius.circular(8),
                  ),
                  child: Row(
                    children: [
                      Icon(Icons.warehouse, size: 16, color: Colors.purple[700]),
                      SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          _selectedWarehouseName != null
                              ? (_selectedLocationName != null
                                  ? '${global.language("warehouse")}: $_selectedWarehouseName / $_selectedLocationName'
                                  : '${global.language("warehouse")}: $_selectedWarehouseName')
                              : global.language("select_warehouse_optional"),
                          style: TextStyle(
                            fontSize: 12,
                            color: _selectedWarehouseName != null
                                ? Colors.purple[900]
                                : Colors.grey[600],
                          ),
                        ),
                      ),
                      Icon(Icons.arrow_forward_ios, size: 12, color: Colors.purple[700]),
                    ],
                  ),
                ),
              ),
              const SizedBox(height: 10),

              // Grid layout สำหรับ form fields
              Row(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // คอลัมน์ซ้าย: หน่วย + จำนวน
                  Expanded(
                    flex: 2,
                    child: Column(
                      children: [
                        // Dropdown เลือกหน่วย
                        _buildUnitDropdown(),
                        const SizedBox(height: 8),
                        // TextField จำนวน
                        _buildQuantityField(),
                      ],
                    ),
                  ),
                  const SizedBox(width: 10),

                  // คอลัมน์กลาง: ราคา + ส่วนลด
                  Expanded(
                    flex: 2,
                    child: Column(
                      children: [
                        // TextField ราคา
                        _buildPriceField(),
                        const SizedBox(height: 8),
                        // TextField ส่วนลด
                        _buildDiscountField(),
                      ],
                    ),
                  ),
                  const SizedBox(width: 10),

                  // คอลัมน์ขวา: หมายเหตุ + ปุ่ม
                  Expanded(
                    flex: 3,
                    child: Column(
                      children: [
                        // TextField หมายเหตุ
                        _buildRemarkField(),
                        const SizedBox(height: 8),
                        // ปุ่มเพิ่มเข้าตะกร้า
                        _buildAddToCartButton(),
                      ],
                    ),
                  ),
                ],
              ),
            ],
          ),
        ),
      ),
    );
  }

  /// Dropdown เลือกหน่วยนับ
  Widget _buildUnitDropdown() {
    return DropdownButtonFormField<ProductSearchModel>(
      initialValue: _selectedUnit,
      decoration: InputDecoration(
        labelText: global.language("unit"),
        labelStyle: const TextStyle(fontSize: 12),
        contentPadding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(6)),
        isDense: true,
      ),
      style: const TextStyle(fontSize: 12, color: Colors.black87),
      items: _allUnits.map((unit) {
        final ratio = unit.unitStand / unit.unitDive;
        final isBasicUnit = ratio == 1.0;
        return DropdownMenuItem<ProductSearchModel>(
          value: unit,
          child: Row(
            mainAxisSize: MainAxisSize.min,
            children: [
              Flexible(
                child: Text(
                  unit.unitName.isNotEmpty ? unit.unitName : unit.unitCode,
                  style: TextStyle(
                    fontWeight: isBasicUnit
                        ? FontWeight.bold
                        : FontWeight.normal,
                    color: isBasicUnit ? Colors.orange[900] : Colors.black87,
                  ),
                  overflow: TextOverflow.ellipsis,
                ),
              ),
              Text(
                ' (${global.formatNumberRemoveRightZero(unit.unitStand)}:${global.formatNumberRemoveRightZero(unit.unitDive)})',
                style: TextStyle(fontSize: 11, color: Colors.grey[600]),
              ),
            ],
          ),
        );
      }).toList(),
      onChanged: (ProductSearchModel? newUnit) {
        if (newUnit != null) {
          setState(() {
            _selectedUnit = newUnit;
            // อัพเดทราคาตามหน่วยที่เลือก
            _priceController.text = global.formatNumberRemoveRightZero(
              newUnit.price1,
            );
          });
        }
      },
      validator: (value) {
        if (value == null) {
          return global.language("please_select_unit");
        }
        return null;
      },
    );
  }

  /// TextField สำหรับจำนวน
  Widget _buildQuantityField() {
    return TextFormField(
      controller: _quantityController,
      decoration: InputDecoration(
        labelText: global.language("qty"),
        labelStyle: const TextStyle(fontSize: 12),
        contentPadding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(6)),
        isDense: true,
        suffixIcon: Icon(Icons.straighten, size: 16, color: Colors.grey[600]),
      ),
      style: const TextStyle(fontSize: 12),
      keyboardType: const TextInputType.numberWithOptions(decimal: true),
      inputFormatters: [
        FilteringTextInputFormatter.allow(RegExp(r'^\d+\.?\d{0,6}')),
      ],
      validator: (value) {
        if (value == null || value.isEmpty) {
          return global.language("please_enter_quantity");
        }
        final qty = double.tryParse(value);
        if (qty == null || qty <= 0) {
          return global.language("quantity_must_be_greater_than_zero");
        }
        return null;
      },
    );
  }

  /// TextField สำหรับราคา
  Widget _buildPriceField() {
    return TextFormField(
      controller: _priceController,
      decoration: InputDecoration(
        labelText: global.language("price_per_unit"),
        labelStyle: const TextStyle(fontSize: 12),
        contentPadding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(6)),
        isDense: true,
        prefixText: global.getBaseCurrency() == 'THB' ? null : '฿',
        suffixIcon: Icon(Icons.attach_money, size: 16, color: Colors.grey[600]),
      ),
      style: const TextStyle(fontSize: 12),
      keyboardType: const TextInputType.numberWithOptions(decimal: true),
      inputFormatters: [
        FilteringTextInputFormatter.allow(RegExp(r'^\d+\.?\d{0,2}')),
      ],
      validator: (value) {
        if (value == null || value.isEmpty) {
          return global.language("please_enter_price");
        }
        final price = double.tryParse(value);
        if (price == null || price < 0) {
          return global.language("price_must_not_be_negative");
        }
        return null;
      },
    );
  }

  /// TextField สำหรับส่วนลด (%)
  Widget _buildDiscountField() {
    return TextFormField(
      controller: _discountController,
      decoration: InputDecoration(
        labelText: '${global.language("discount")} (%)',
        labelStyle: const TextStyle(fontSize: 12),
        contentPadding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(6)),
        isDense: true,
        suffixIcon: Icon(Icons.percent, size: 16, color: Colors.grey[600]),
      ),
      style: const TextStyle(fontSize: 12),
      keyboardType: const TextInputType.numberWithOptions(decimal: true),
      inputFormatters: [
        FilteringTextInputFormatter.allow(RegExp(r'^\d+\.?\d{0,2}')),
      ],
      validator: (value) {
        if (value != null && value.isNotEmpty) {
          final discount = double.tryParse(value);
          if (discount == null || discount < 0 || discount > 100) {
            return global.language("discount_must_be_between_0_100");
          }
        }
        return null;
      },
    );
  }

  /// TextField สำหรับหมายเหตุ
  Widget _buildRemarkField() {
    return TextFormField(
      controller: _remarkController,
      decoration: InputDecoration(
        labelText: global.language("remark_optional"),
        labelStyle: const TextStyle(fontSize: 12),
        contentPadding: const EdgeInsets.symmetric(horizontal: 10, vertical: 8),
        border: OutlineInputBorder(borderRadius: BorderRadius.circular(6)),
        isDense: true,
        suffixIcon: Icon(Icons.note_alt, size: 16, color: Colors.grey[600]),
      ),
      style: const TextStyle(fontSize: 12),
      maxLines: 1,
    );
  }

  /// ปุ่มเพิ่มสินค้าเข้าตะกร้า
  Widget _buildAddToCartButton() {
    return SizedBox(
      width: double.infinity,
      child: ElevatedButton.icon(
        onPressed: _isSavingToCart ? null : _addToCart,
        icon: _isSavingToCart
            ? const SizedBox(
                width: 16,
                height: 16,
                child: CircularProgressIndicator(
                  strokeWidth: 2,
                  valueColor: AlwaysStoppedAnimation<Color>(Colors.white),
                ),
              )
            : Icon(Icons.add_shopping_cart, size: 18),
        label: Text(
          _isSavingToCart ? global.language("saving") : global.language("add_to_cart"),
          style: const TextStyle(fontSize: 12, fontWeight: FontWeight.bold),
        ),
        style: ElevatedButton.styleFrom(
          backgroundColor: Colors.orange[700],
          foregroundColor: Colors.white,
          padding: const EdgeInsets.symmetric(vertical: 12),
          shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(6)),
        ),
      ),
    );
  }

  /// บันทึกสินค้าเข้าตะกร้า
  Future<void> _addToCart() async {
    // Validate form
    if (!_formKey.currentState!.validate()) {
      AppLogger.warning('🛒 Form validation failed');
      return;
    }

    if (_selectedUnit == null) {
      AppLogger.warning('🛒 No unit selected');
      _showMessage(global.language("please_select_unit"), isError: true);
      return;
    }

    AppLogger.info('🛒 Starting add to cart process...');
    AppLogger.info(
      '  - Product: ${widget.product.itemCode} - ${widget.product.name0}',
    );
    AppLogger.info(
      '  - Selected Unit: ${_selectedUnit!.unitCode} (${_selectedUnit!.unitName})',
    );
    AppLogger.info('  - Barcode: ${_selectedUnit!.barcode}');

    setState(() => _isSavingToCart = true);

    try {
      // Parse ค่าจาก form
      final quantity = double.parse(_quantityController.text);
      final price = double.parse(_priceController.text);
      final discount = _discountController.text.isNotEmpty
          ? double.parse(_discountController.text)
          : 0.0;
      final remark = _remarkController.text.isNotEmpty
          ? _remarkController.text
          : null;

      AppLogger.info(
        '  - Quantity: $quantity, Price: $price, Discount: $discount',
      );

      // ดึงตะกร้าที่ active หรือสร้างใหม่
      // ถ้าไม่มี email ให้ใช้ username แทน
      var email = global.userLoginData.email;
      if (email.isEmpty) {
        email = global.apiUserName;
        AppLogger.info('🛒 Email is empty, using username instead: $email');
      }

      AppLogger.info('🛒 Login email: $email');
      AppLogger.info('🛒 User name: ${global.userLoginData.name}');

      if (email.isEmpty) {
        AppLogger.error('🛒 Both email and username are empty!');
        _showMessage(global.language("user_not_found_please_login"), isError: true);
        setState(() => _isSavingToCart = false);
        return;
      }

      AppLogger.info(
        '🛒 Fetching carts for shopId: ${widget.shopId}, email: $email',
      );

      // ดึงตะกร้าทั้งหมดของ user
      final carts = await _cartService.getCarts(widget.shopId, email);
      AppLogger.info('🛒 Found ${carts.length} existing cart(s)');

      // หาตะกร้าที่จะใช้งาน
      CartModel cart;

      // ถ้ามี activeCartId ให้ใช้ตะกร้านั้น
      if (widget.activeCartId != null) {
        AppLogger.info('🛒 Using active cart ID: ${widget.activeCartId}');
        final activeCart = carts
            .where((c) => c.cartId == widget.activeCartId)
            .firstOrNull;

        if (activeCart != null) {
          cart = activeCart;
          AppLogger.info('🛒 Found active cart: ${cart.cartName}');
        } else {
          // ถ้าไม่เจอตะกร้าที่ระบุ ให้สร้างใหม่
          AppLogger.warning('🛒 Active cart not found, creating new cart...');
          cart = CartModel(
            shopId: widget.shopId,
            email: email,
            cartName:
                '${global.language("cart")} ${global.formatThaiDateTime(dateTime: DateTime.now(), showTime: false)}',
            systemType: CartSystemType.inventory, // Default เป็นระบบคงคลัง
            warehouseId: widget.shopId, // ใช้ shopId เป็น default warehouse
            warehouseName: global.language("default_warehouse"), // Default warehouse name
          );
        }
      } else if (carts.isEmpty || carts.every((c) => c.status != 'active')) {
        // ไม่มี activeCartId และไม่มีตะกร้า active ให้สร้างใหม่
        AppLogger.info('🛒 No active cart found, creating new cart...');
        cart = CartModel(
          shopId: widget.shopId,
          email: email,
          cartName:
              '${global.language("cart")} ${global.formatThaiDateTime(dateTime: DateTime.now(), showTime: false)}',
          systemType: CartSystemType.inventory, // Default เป็นระบบคงคลัง
          warehouseId: widget.shopId, // ใช้ shopId เป็น default warehouse
          warehouseName: global.language("default_warehouse"), // Default warehouse name
        );
        AppLogger.info('🛒 New cart created: ${cart.cartId}');
      } else {
        // ใช้ตะกร้า active ที่มีอยู่
        cart = carts.firstWhere((c) => c.status == 'active');
        AppLogger.info('🛒 Using existing active cart: ${cart.cartId}');
      }

      AppLogger.info('🛒 Cart before adding item: ${cart.items.length} items');

      // สร้าง CartItem พร้อมข้อมูลคลัง/ที่เก็บ
      // ใช้คลังที่เลือกในหน้านี้ก่อน ถ้าไม่มีค่อยใช้จากตะกร้า
      final warehouseId = _selectedWarehouseId ?? cart.warehouseId;
      final warehouseName = _selectedWarehouseName ?? cart.warehouseName;
      final locationId = _selectedLocationId ?? cart.locationId;
      final locationName = _selectedLocationName ?? cart.locationName;

      final cartItem = CartItemModel(
        itemCode: widget.product.itemCode,
        barcode: _selectedUnit!.barcode,
        productName: widget.product.name0,
        unitCode: _selectedUnit!.unitCode,
        unitName: _selectedUnit!.unitName,
        quantity: quantity,
        price: price,
        discount: discount,
        remark: remark,
        imageuri: _selectedUnit!.imageuri ?? widget.product.imageuri,
        unitStand: _selectedUnit!.unitStand.toInt(),
        unitDive: _selectedUnit!.unitDive.toInt(),
        // บันทึกข้อมูลคลัง/ที่เก็บ (ใช้ที่เลือกในหน้านี้ หรือจากตะกร้า)
        warehouseId: warehouseId,
        warehouseName: warehouseName,
        locationId: locationId,
        locationName: locationName,
      );

      AppLogger.info('🛒 CartItem created: ${cartItem.toString()}');
      AppLogger.info('🛒 Warehouse: $warehouseName, Location: $locationName');

      // เพิ่มสินค้าเข้าตะกร้า
      final updatedCart = cart.addItem(cartItem);

      AppLogger.info(
        '🛒 Cart after adding item: ${updatedCart.items.length} items',
      );
      AppLogger.info('🛒 Saving cart to MongoDB...');

      // บันทึกลง MongoDB
      if (carts.isEmpty || carts.every((c) => c.status != 'active')) {
        AppLogger.info('🛒 Calling createCart()...');
        await _cartService.createCart(updatedCart);
        AppLogger.info('✅ Cart created successfully');
      } else {
        AppLogger.info('🛒 Calling updateCart()...');
        await _cartService.updateCart(updatedCart);
        AppLogger.info('✅ Cart updated successfully');
      }

      // แสดงข้อความสำเร็จ
      _showMessage(
        '${global.language("add_product_to_cart_success")}\n${cartItem.productName} x ${global.formatNumberRemoveRightZero(quantity)} ${cartItem.unitName}',
        isError: false,
      );

      // ปิด dialog และส่งสัญญาณให้ refresh
      if (mounted) {
        Navigator.of(
          context,
        ).pop(true); // ส่ง true เพื่อบอกว่าเพิ่มสินค้าสำเร็จ
      }
    } catch (e, stackTrace) {
      AppLogger.error(
        '🛒 Error adding to cart',
        error: e,
        stackTrace: stackTrace,
      );
      _showMessage('${global.language("error_adding_product")}: $e', isError: true);
    } finally {
      setState(() => _isSavingToCart = false);
    }
  }

  /// แสดงข้อความแจ้งเตือน
  void _showMessage(String message, {required bool isError}) {
    if (!mounted) return;

    if (isError) {
      global.showErrorSnackBar(context, message);
    } else {
      global.showSuccessSnackBar(context, message);
    }
  }
}
