import 'dart:async';
import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import 'package:smlaicloud/services/product_cache_api_service.dart';
import '../global.dart' as global;

class SelectProductBarcodeWidget extends StatefulWidget {
  const SelectProductBarcodeWidget({super.key, required this.selectedItems});

  final List<SelectedProductItemStruct> selectedItems;

  @override
  SelectProductBarcodeWidgetState createState() =>
      SelectProductBarcodeWidgetState();
}

class SelectProductBarcodeWidgetState
    extends State<SelectProductBarcodeWidget> {
  List<SelectedProductItemStruct> _products = [];
  String? _error;
  late TextEditingController _searchController;
  int _timeCount = 0;
  late Timer _timer;
  String _lastSearchText = "";
  late FocusNode _searchFocusNode;
  final List<SelectedProductItemStruct> _itemCodeSelected = [];
  bool _isLoading = false;

  // เพิ่มตัวแปรสำหรับควบคุมขนาดตัวอักษร
  double _fontSizeScale = 1.0; // ค่าเริ่มต้น 1.0

  // คำนวณขนาดอักษรตามสเกล
  double _fontSize(double baseSize) => baseSize * _fontSizeScale;

  @override
  void initState() {
    super.initState();

    // เอาค่าจาก selectedItems ที่ส่งเข้ามา
    _itemCodeSelected.addAll(widget.selectedItems);

    _searchController = TextEditingController();
    _searchFocusNode = FocusNode();
    _searchController.addListener(() {
      _timeCount = 0;
    });
    _timer = Timer.periodic(const Duration(milliseconds: 200), (timer) {
      _timeCount++;
      if (_timeCount > 2) {
        _timeCount = 0;
        String searchText = _searchController.text.trim();
        if (_lastSearchText != searchText) {
          _fetchProducts();
          if (searchText != _lastSearchText) {
            _lastSearchText = searchText;
            setState(() {});
          }
        }
      }
    });
    _fetchProducts();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      _searchFocusNode.requestFocus();
    });
  }

  @override
  void dispose() {
    _searchController.dispose();
    _searchFocusNode.dispose();
    _timer.cancel();
    super.dispose();
  }

  Future<void> _fetchProducts() async {
    setState(() {
      _isLoading = true;
    });
    try {
      final searchText = _searchController.text.trim();

      if (kDebugMode) {
        AppLogger.debug("Fetching products with search: $searchText");
      }

      // ใช้ ProductCacheApiService แทน raw SQL query (ป้องกัน SQL injection)
      final response = await ProductCacheApiService.searchProducts(
        search: searchText,
        limit: 200,
        offset: 0,
      );

      if (!response.isSuccess) {
        setState(() {
          _products = [];
          _isLoading = false;
          _error = response.errorMessage;
        });
        return;
      }

      // แปลงข้อมูลเป็น SelectedProductItemStruct
      List<SelectedProductItemStruct> products = [];
      for (var item in response.products) {
        products.add(
          SelectedProductItemStruct(
            itemCode: item.itemCode,
            itemName: item.itemName,
            barcode: item.barcode,
          ),
        );
      }

      setState(() {
        _products = products;
        _isLoading = false;
        _error = null;
      });
    } catch (e) {
      setState(() {
        _error = e.toString();
        _isLoading = false;
      });
    }
  }

  // ฟังก์ชันเพิ่มขนาดตัวอักษร
  void _increaseFontSize() {
    setState(() {
      _fontSizeScale += 0.1;
      if (_fontSizeScale > 1.5) _fontSizeScale = 1.5; // ขนาดสูงสุด
    });
  }

  // ฟังก์ชันลดขนาดตัวอักษร
  void _decreaseFontSize() {
    setState(() {
      _fontSizeScale -= 0.1;
      if (_fontSizeScale < 0.8) _fontSizeScale = 0.8; // ขนาดต่ำสุด
    });
  }

  @override
  Widget build(BuildContext context) {
    double padding = 4.0;

    if (_error != null) {
      return Center(child: Text('Error: $_error'));
    }

    return Scaffold(
      appBar: AppBar(
        title: Text(global.language("select_product_barcode")),
        backgroundColor: Colors.blue.shade700,
        elevation: 2,
        actions: [
          // เพิ่มปุ่มปรับขนาดตัวอักษร
          IconButton(
            icon: Icon(Icons.text_decrease),
            tooltip: global.language("decrease_font_size"),
            onPressed: _decreaseFontSize,
          ),
          IconButton(
            icon: Icon(Icons.text_increase),
            tooltip: global.language("increase_font_size"),
            onPressed: _increaseFontSize,
          ),
        ],
      ),
      body: Container(
        padding: EdgeInsets.all(padding),
        color: Colors.grey.shade100,
        child: Column(
          children: [
            Card(
              elevation: 2,
              margin: const EdgeInsets.symmetric(
                horizontal: 4.0,
                vertical: 4.0,
              ),
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(10),
              ),
              child: Padding(
                padding: EdgeInsets.all(4.0),
                child: TextField(
                  controller: _searchController,
                  focusNode: _searchFocusNode,
                  autofocus: true,
                  style: TextStyle(
                    fontSize: _fontSize(16),
                  ), // ปรับขนาดตัวอักษรในช่องค้นหา
                  decoration: InputDecoration(
                    labelText: global.language('filter_product'),
                    labelStyle: TextStyle(
                      fontSize: _fontSize(16),
                    ), // ปรับขนาดป้ายกำกับ
                    hintText: global.language("search_barcode_or_product"),
                    hintStyle: TextStyle(
                      fontSize: _fontSize(14),
                    ), // ปรับขนาดคำแนะนำ
                    prefixIcon: const Icon(Icons.search, color: Colors.blue),
                    contentPadding: const EdgeInsets.symmetric(
                      horizontal: 8.0,
                      vertical: 8.0,
                    ),
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(8),
                      borderSide: BorderSide(color: Colors.blue.shade200),
                    ),
                    enabledBorder: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(8),
                      borderSide: BorderSide(color: Colors.blue.shade200),
                    ),
                    focusedBorder: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(8),
                      borderSide: BorderSide(
                        color: Colors.blue.shade500,
                        width: 2,
                      ),
                    ),
                    suffixIcon: IconButton(
                      icon: const Icon(Icons.clear, color: Colors.red),
                      onPressed: () {
                        _searchController.clear();
                        setState(() {
                          _products = [];
                        });
                      },
                    ),
                  ),
                  onChanged: (value) {
                    setState(() {});
                  },
                ),
              ),
            ),

            // แสดงรายการสินค้าที่เลือกแล้ว
            if (_itemCodeSelected.isNotEmpty)
              Card(
                elevation: 2,
                margin: const EdgeInsets.symmetric(
                  horizontal: 4.0,
                  vertical: 4.0,
                ),
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Padding(
                  padding: const EdgeInsets.all(8.0),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        children: [
                          const Icon(
                            Icons.shopping_cart,
                            color: Colors.blue,
                            size: 20,
                          ),
                          const SizedBox(width: 4),
                          Text(
                            "สินค้าที่เลือก (${_itemCodeSelected.length})",
                            style: TextStyle(
                              fontSize: _fontSize(16), // ปรับขนาดตัวอักษร
                              fontWeight: FontWeight.bold,
                              color: Colors.blue,
                            ),
                          ),
                          Spacer(),
                          TextButton.icon(
                            icon: const Icon(
                              Icons.delete_sweep,
                              color: Colors.red,
                              size: 18,
                            ),
                            label: Text(
                              global.language("clear_all"),
                              style: TextStyle(
                                color: Colors.red,
                                fontSize: _fontSize(14),
                              ),
                            ),
                            style: TextButton.styleFrom(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 8,
                                vertical: 0,
                              ),
                            ),
                            onPressed: () {
                              setState(() {
                                _itemCodeSelected.clear();
                              });
                            },
                          ),
                        ],
                      ),
                      const Divider(height: 8),
                      Wrap(
                        spacing: 4,
                        runSpacing: 4,
                        children: _itemCodeSelected.map((item) {
                          return Chip(
                            materialTapTargetSize:
                                MaterialTapTargetSize.shrinkWrap,
                            padding: const EdgeInsets.all(2),
                            labelPadding: const EdgeInsets.symmetric(
                              horizontal: 6,
                            ),
                            backgroundColor: Colors.blue.shade100,
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(12),
                              side: BorderSide(color: Colors.blue.shade300),
                            ),
                            label: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                // รหัสสินค้า (แสดงแค่เมื่อมี)
                                if (item.itemCode.isNotEmpty) ...[
                                  Text(
                                    item.itemCode,
                                    style: TextStyle(
                                      fontSize: _fontSize(13),
                                      color: Colors.red.shade700,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                  const SizedBox(height: 2),
                                ],

                                // Barcode
                                Text(
                                  item.barcode,
                                  style: TextStyle(
                                    fontSize: _fontSize(12),
                                    color: Colors.orange.shade700,
                                    fontWeight: FontWeight.w600,
                                  ),
                                  softWrap: true,
                                ),

                                // ชื่อสินค้า (แสดงแค่เมื่อมี)
                                if (item.itemName.isNotEmpty) ...[
                                  const SizedBox(height: 2),
                                  Text(
                                    item.itemName,
                                    style: TextStyle(
                                      fontSize: _fontSize(10),
                                      color: Colors.grey.shade600,
                                    ),
                                    softWrap: true,
                                  ),
                                ],
                              ],
                            ),
                            deleteIcon: const Icon(
                              Icons.cancel,
                              size: 18,
                              color: Colors.red,
                            ),
                            onDeleted: () {
                              setState(() {
                                _itemCodeSelected.remove(item);
                              });
                            },
                          );
                        }).toList(),
                      ),
                    ],
                  ),
                ),
              ),

            // สถานะการโหลด
            if (_isLoading)
              const Padding(
                padding: EdgeInsets.all(8.0),
                child: Center(
                  child: SizedBox(
                    width: 30,
                    height: 30,
                    child: CircularProgressIndicator(strokeWidth: 3),
                  ),
                ),
              ),

            // แสดงรายการสินค้าที่ค้นหา
            Expanded(
              child: _products.isEmpty
                  ? Center(
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(
                            _searchController.text.isEmpty
                                ? Icons.search
                                : Icons.inventory_2_outlined,
                            size: 48,
                            color: Colors.grey.shade400,
                          ),
                          const SizedBox(height: 8),
                          Text(
                            _searchController.text.isEmpty
                                ? "กรุณากรอกข้อมูลที่ต้องการค้นหา"
                                : "ไม่พบข้อมูลสินค้าที่ค้นหา",
                            style: TextStyle(
                              fontSize: _fontSize(16), // ปรับขนาดตัวอักษร
                              color: Colors.grey.shade600,
                            ),
                          ),
                        ],
                      ),
                    )
                  : SizedBox(
                      width: double.infinity,
                      child: Wrap(
                        spacing: 4,
                        runSpacing: 4,
                        alignment: WrapAlignment.start,
                        crossAxisAlignment: WrapCrossAlignment.start,
                        children: _products.map((product) {
                          final isSelected = _itemCodeSelected.contains(
                            product,
                          );

                          return Container(
                            width: 150,
                            padding: const EdgeInsets.all(6),
                            decoration: BoxDecoration(
                              color: isSelected
                                  ? Colors.blue.shade50
                                  : Colors.white,
                              borderRadius: BorderRadius.circular(8),
                              border: Border.all(
                                color: isSelected
                                    ? Colors.blue
                                    : Colors.grey.shade300,
                              ),
                            ),
                            margin: EdgeInsets.zero,
                            child: InkWell(
                              onTap: () {
                                setState(() {
                                  if (isSelected) {
                                    _itemCodeSelected.remove(product);
                                  } else {
                                    _itemCodeSelected.add(product);
                                  }
                                });
                              },
                              borderRadius: BorderRadius.circular(8),
                              child: Column(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  Row(
                                    children: [
                                      Icon(
                                        isSelected
                                            ? Icons.check_circle
                                            : Icons.circle_outlined,
                                        color: isSelected
                                            ? Colors.blue
                                            : Colors.grey,
                                        size: 18,
                                      ),
                                      const Spacer(),
                                      if (isSelected)
                                        Container(
                                          padding: const EdgeInsets.symmetric(
                                            horizontal: 6,
                                            vertical: 2,
                                          ),
                                          decoration: BoxDecoration(
                                            color: Colors.blue,
                                            borderRadius: BorderRadius.circular(
                                              8,
                                            ),
                                          ),
                                          child: Text(
                                            global.language("chosen"),
                                            style: TextStyle(
                                              color: Colors.white,
                                              fontSize: _fontSize(
                                                10,
                                              ), // ปรับขนาดตัวอักษร
                                            ),
                                          ),
                                        ),
                                    ],
                                  ),
                                  const SizedBox(height: 4),
                                  Text(
                                    product.itemCode,
                                    style: TextStyle(
                                      fontSize: _fontSize(14),
                                      color: Colors.red.shade700,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                  const SizedBox(height: 4),
                                  Text(
                                    product.barcode,
                                    style: TextStyle(
                                      fontSize: _fontSize(12),
                                      color: Colors.blue.shade800,
                                      fontWeight: FontWeight.w500,
                                    ),
                                    softWrap: true,
                                  ),
                                  const SizedBox(height: 4),
                                  Text(
                                    product.itemName,
                                    style: TextStyle(
                                      fontSize: _fontSize(12),
                                      color: Colors.black,
                                      fontWeight: FontWeight.w500,
                                    ),
                                    softWrap: true,
                                  ),
                                ],
                              ),
                            ),
                          );
                        }).toList(),
                      ),
                    ),
            ),
          ],
        ),
      ),
      floatingActionButton: _itemCodeSelected.isNotEmpty
          ? FloatingActionButton.extended(
              onPressed: () {
                Navigator.pop(context, _itemCodeSelected);
              },
              backgroundColor: Colors.green,
              icon: const Icon(Icons.check, size: 20),
              label: Text(
                "ยืนยัน (${_itemCodeSelected.length})",
                style: TextStyle(fontSize: _fontSize(14)), // ปรับขนาดตัวอักษร
              ),
            )
          : null,
    );
  }
}
