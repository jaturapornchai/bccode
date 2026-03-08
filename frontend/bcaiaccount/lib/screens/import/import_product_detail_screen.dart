import 'package:flutter/material.dart';
import 'package:smlaicloud/model/import/product_import_response.dart';
import 'package:smlaicloud/global.dart' as global;

class ImportProductDetailScreen extends StatefulWidget {
  final ProductImportResponse importResult;

  const ImportProductDetailScreen({super.key, required this.importResult});

  @override
  State<ImportProductDetailScreen> createState() =>
      _ImportProductDetailScreenState();
}

class _ImportProductDetailScreenState extends State<ImportProductDetailScreen> {
  // Filter and search
  String searchQuery = '';
  int filterAction = -1; // -1=ทั้งหมด, 0=ไม่เปลี่ยนแปลง, 1=Insert, 2=Update
  final TextEditingController _searchController = TextEditingController();
  List<ProductComparison> _filteredProducts = [];

  @override
  void initState() {
    super.initState();
    _updateFilteredProducts();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  // Filter products based on search and action
  void _updateFilteredProducts() {
    var products = List<ProductComparison>.from(
      widget.importResult.comparison?.products ?? [],
    );

    // Sort by barcode
    products.sort((a, b) => a.barcode.compareTo(b.barcode));

    // Filter by action
    if (filterAction != -1) {
      products = products.where((p) => p.action == filterAction).toList();
    }

    // Filter by search query
    if (searchQuery.isNotEmpty) {
      final query = searchQuery.toLowerCase();
      products = products.where((p) {
        final barcodeMatch = p.barcode.toLowerCase().contains(query);
        final mongoCodeMatch =
            p.mongo?.code.toLowerCase().contains(query) ?? false;
        final mongoNameMatch =
            p.mongo?.name.toLowerCase().contains(query) ?? false;
        final excelCodeMatch =
            p.excel?.code.toLowerCase().contains(query) ?? false;
        final excelNameMatch =
            p.excel?.name.toLowerCase().contains(query) ?? false;

        return barcodeMatch ||
            mongoCodeMatch ||
            mongoNameMatch ||
            excelCodeMatch ||
            excelNameMatch;
      }).toList();
    }

    setState(() {
      _filteredProducts = products;
    });
  }

  Widget _buildDataField(String label, String? value) {
    return Padding(
      padding: EdgeInsets.only(bottom: 6),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 80,
            child: Text(
              '$label:',
              style: TextStyle(fontSize: 13, color: Colors.grey[700]),
            ),
          ),
          Expanded(
            child: Text(
              value ?? global.language('import_product_detail.no_data'),
              style: TextStyle(
                fontSize: 13,
                fontWeight: value != null ? FontWeight.w500 : FontWeight.normal,
                color: value != null ? Colors.black : Colors.grey,
                fontStyle: value != null ? FontStyle.normal : FontStyle.italic,
              ),
            ),
          ),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(
          global.language('import_product_detail.product_import_details'),
        ),
        backgroundColor: const Color(0xFF0A3880),
      ),
      body: Column(
        children: [
          // Search and Filter Section
          Container(
            padding: const EdgeInsets.all(16),
            color: Colors.grey[100],
            child: Column(
              children: [
                // Header with count
                Row(
                  children: [
                    Text(
                      '${global.language('import_product_detail.total')} ${widget.importResult.comparison?.products.length ?? 0} ${global.language('import_product_detail.items')}',
                      style: const TextStyle(
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    Spacer(),
                    Text(
                      '${global.language('import_product_detail.showing')} ${_filteredProducts.length} ${global.language('import_product_detail.items')}',
                      style: TextStyle(fontSize: 14, color: Colors.grey[600]),
                    ),
                  ],
                ),
                const SizedBox(height: 12),

                // Search Box
                TextField(
                  controller: _searchController,
                  decoration: InputDecoration(
                    labelText: global.language('import_product_detail.search'),
                    hintText: global.language(
                      'import_product_detail.search_hint',
                    ),
                    prefixIcon: const Icon(Icons.search),
                    suffixIcon: searchQuery.isNotEmpty
                        ? IconButton(
                            icon: const Icon(Icons.clear),
                            onPressed: () {
                              _searchController.clear();
                              searchQuery = '';
                              _updateFilteredProducts();
                            },
                          )
                        : null,
                    border: OutlineInputBorder(
                      borderRadius: BorderRadius.circular(8),
                    ),
                    filled: true,
                    fillColor: Colors.white,
                    contentPadding: const EdgeInsets.symmetric(
                      horizontal: 16,
                      vertical: 12,
                    ),
                  ),
                  onChanged: (value) {
                    searchQuery = value;
                    _updateFilteredProducts();
                  },
                ),
                const SizedBox(height: 12),

                // Filter Buttons
                Wrap(
                  spacing: 8,
                  runSpacing: 8,
                  children: [
                    FilterChip(
                      label: Text(
                        '${global.language('import_product_detail.all')} (${widget.importResult.comparison?.products.length ?? 0})',
                      ),
                      selected: filterAction == -1,
                      onSelected: (selected) {
                        filterAction = -1;
                        _updateFilteredProducts();
                      },
                      selectedColor: Colors.blue[100],
                    ),
                    FilterChip(
                      label: Text(
                        'Insert (${widget.importResult.comparison?.newProductCount ?? 0})',
                      ),
                      selected: filterAction == 1,
                      onSelected: (selected) {
                        filterAction = 1;
                        _updateFilteredProducts();
                      },
                      selectedColor: Colors.green[100],
                    ),
                    FilterChip(
                      label: Text(
                        'Update (${widget.importResult.comparison?.updatedProductCount ?? 0})',
                      ),
                      selected: filterAction == 2,
                      onSelected: (selected) {
                        filterAction = 2;
                        _updateFilteredProducts();
                      },
                      selectedColor: Colors.orange[100],
                    ),
                    FilterChip(
                      label: Text(
                        '${global.language('import_product_detail.unchanged')} (${widget.importResult.comparison?.unchangedCount ?? 0})',
                      ),
                      selected: filterAction == 0,
                      onSelected: (selected) {
                        filterAction = 0;
                        _updateFilteredProducts();
                      },
                      selectedColor: Colors.grey[300],
                    ),
                  ],
                ),
              ],
            ),
          ),

          // Products List with ListView.builder (Lazy Loading)
          Expanded(
            child: _filteredProducts.isEmpty
                ? Center(
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(
                          Icons.search_off,
                          size: 64,
                          color: Colors.grey[400],
                        ),
                        SizedBox(height: 16),
                        Text(
                          global.language(
                            'import_product_detail.no_data_found',
                          ),
                          style: TextStyle(
                            fontSize: 16,
                            color: Colors.grey[600],
                          ),
                        ),
                      ],
                    ),
                  )
                : ListView.builder(
                    itemCount: _filteredProducts.length,
                    padding: const EdgeInsets.all(16),
                    itemBuilder: (context, index) {
                      final product = _filteredProducts[index];
                      return Container(
                        margin: const EdgeInsets.only(bottom: 12),
                        padding: const EdgeInsets.all(16),
                        decoration: BoxDecoration(
                          color: product.actionColor.withValues(alpha: 0.1),
                          border: Border.all(
                            color: product.actionColor.withValues(alpha: 0.3),
                            width: 2,
                          ),
                          borderRadius: BorderRadius.circular(8),
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            // Status Badge and Barcode
                            Row(
                              children: [
                                Container(
                                  padding: const EdgeInsets.symmetric(
                                    horizontal: 12,
                                    vertical: 6,
                                  ),
                                  decoration: BoxDecoration(
                                    color: product.actionColor,
                                    borderRadius: BorderRadius.circular(6),
                                  ),
                                  child: Text(
                                    product.actionText,
                                    style: const TextStyle(
                                      color: Colors.white,
                                      fontSize: 13,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                ),
                                const SizedBox(width: 12),
                                Icon(
                                  Icons.qr_code,
                                  color: Colors.grey[700],
                                  size: 20,
                                ),
                                const SizedBox(width: 6),
                                Expanded(
                                  child: Text(
                                    product.barcode,
                                    style: const TextStyle(
                                      fontSize: 16,
                                      fontWeight: FontWeight.bold,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 16),

                            // Data Comparison
                            Row(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                // Mongo Data
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
                                      Row(
                                        children: [
                                          Icon(
                                            Icons.storage,
                                            color: Colors.blue[700],
                                            size: 18,
                                          ),
                                          const SizedBox(width: 6),
                                          Text(
                                            'MongoDB',
                                            style: TextStyle(
                                              fontSize: 14,
                                              fontWeight: FontWeight.bold,
                                              color: Colors.blue[700],
                                            ),
                                          ),
                                        ],
                                      ),
                                      SizedBox(height: 8),
                                      _buildDataField(
                                        global.language(
                                          'import_product_detail.product_code',
                                        ),
                                        product.mongo?.code,
                                      ),
                                      _buildDataField(
                                        global.language(
                                          'import_product_detail.product_name',
                                        ),
                                        product.mongo?.name,
                                      ),
                                      _buildDataField(
                                        global.language(
                                          'import_product_detail.unit',
                                        ),
                                        product.mongo?.unitCode,
                                      ),
                                      _buildDataField(
                                        'Divide',
                                        product.mongo?.divideValue.toString(),
                                      ),
                                      _buildDataField(
                                        'Stand',
                                        product.mongo?.standValue.toString(),
                                      ),
                                    ],
                                  ),
                                ),
                                const SizedBox(width: 20),
                                // Excel Data
                                Expanded(
                                  child: Column(
                                    crossAxisAlignment:
                                        CrossAxisAlignment.start,
                                    children: [
                                      Row(
                                        children: [
                                          Icon(
                                            Icons.table_chart,
                                            color: Colors.green[700],
                                            size: 18,
                                          ),
                                          const SizedBox(width: 6),
                                          Text(
                                            'Excel',
                                            style: TextStyle(
                                              fontSize: 14,
                                              fontWeight: FontWeight.bold,
                                              color: Colors.green[700],
                                            ),
                                          ),
                                        ],
                                      ),
                                      SizedBox(height: 8),
                                      _buildDataField(
                                        global.language(
                                          'import_product_detail.product_code',
                                        ),
                                        product.excel?.code,
                                      ),
                                      _buildDataField(
                                        global.language(
                                          'import_product_detail.product_name',
                                        ),
                                        product.excel?.name,
                                      ),
                                      _buildDataField(
                                        global.language(
                                          'import_product_detail.unit',
                                        ),
                                        product.excel?.unitCode,
                                      ),
                                      _buildDataField(
                                        'Divide',
                                        product.excel?.divideValue.toString(),
                                      ),
                                      _buildDataField(
                                        'Stand',
                                        product.excel?.standValue.toString(),
                                      ),
                                    ],
                                  ),
                                ),
                              ],
                            ),
                          ],
                        ),
                      );
                    },
                  ),
          ),
        ],
      ),
    );
  }
}
