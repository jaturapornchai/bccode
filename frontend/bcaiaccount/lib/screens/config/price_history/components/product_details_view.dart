import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/model/price_history_model.dart';
import 'package:smlaicloud/model/global_model.dart';
import 'package:smlaicloud/bloc/product_barcode/product_barcode_bloc.dart';
import 'package:smlaicloud/global.dart' as global;

class ProductDetailsView extends StatefulWidget {
  final ProductBarcodeModel? productData;
  final TabController? tabController;
  final bool isMobile;

  const ProductDetailsView({
    super.key,
    required this.productData,
    this.tabController,
    this.isMobile = false,
  });

  @override
  State<ProductDetailsView> createState() => _ProductDetailsViewState();
}

class _ProductDetailsViewState extends State<ProductDetailsView> {
  ProductBarcodeModel? detailProductData;

  @override
  void initState() {
    super.initState();
    if (widget.productData?.barcode != null) {
      context.read<ProductBarcodeBloc>().add(
            ProductBarcodeGetPriceHistory(
                barcode: widget.productData!.barcode!),
          );
    }
  }

  @override
  void didUpdateWidget(ProductDetailsView oldWidget) {
    super.didUpdateWidget(oldWidget);
    if (widget.productData?.barcode != oldWidget.productData?.barcode &&
        widget.productData?.barcode != null) {
      context.read<ProductBarcodeBloc>().add(
            ProductBarcodeGetPriceHistory(
                barcode: widget.productData!.barcode!),
          );
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        backgroundColor: global.theme.appBarColor,
        automaticallyImplyLeading: false,
        title: Text(global.language('product_details_price_history')),
        leading: widget.isMobile && widget.tabController != null
            ? IconButton(
                focusNode: FocusNode(skipTraversal: true),
                icon: const Icon(Icons.arrow_back),
                onPressed: () {
                  widget.tabController!.animateTo(0);
                },
              )
            : null,
      ),
      body: Container(
        padding: const EdgeInsets.all(16),
        child: widget.productData != null
            ? Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  // Product info header
                  Container(
                    width: double.infinity,
                    padding: const EdgeInsets.all(12),
                    decoration: BoxDecoration(
                      color: Colors.blue[50],
                      borderRadius: BorderRadius.circular(8),
                      border: Border.all(color: Colors.blue[200]!),
                    ),
                    child: Column(
                      crossAxisAlignment: CrossAxisAlignment.start,
                      children: [
                        Text(
                          'Barcode: ${widget.productData!.barcode}',
                          style: const TextStyle(
                            fontSize: 16,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                        const SizedBox(height: 4),
                        Text(
                          'Product: ${global.activeLangName(widget.productData!.names!)}',
                          style: const TextStyle(fontSize: 14),
                        ),
                      ],
                    ),
                  ),
                  const SizedBox(height: 16),
                  // Price history content
                  Expanded(
                    child:
                        BlocListener<ProductBarcodeBloc, ProductBarcodeState>(
                      listener: (context, state) {
                        if (state is ProductBarcodeGetSuccess) {
                          setState(() {
                            detailProductData = state.productBarcode;
                          });
                        }
                      },
                      child:
                          BlocBuilder<ProductBarcodeBloc, ProductBarcodeState>(
                        builder: (context, state) {
                          if (state
                              is ProductBarcodeGetPriceHistoryInProgress) {
                            return const Center(
                              child: CircularProgressIndicator(),
                            );
                          } else if (state
                              is ProductBarcodeGetPriceHistorySuccess) {
                            return _buildPriceHistoryContent(
                                state.priceHistory);
                          } else if (state
                              is ProductBarcodeGetPriceHistoryFailed) {
                            return Center(
                              child: Text(
                                'Error: ${state.message}',
                                style: const TextStyle(color: Colors.red),
                              ),
                            );
                          }
                          return const Center(
                            child: Text('No price history data'),
                          );
                        },
                      ),
                    ),
                  ),
                ],
              )
            : Center(
                child: Text(
                  global.language('product_details_select_product'),
                  style: TextStyle(
                    fontSize: 16,
                    color: Colors.grey[600],
                  ),
                ),
              ),
      ),
    );
  }

  Widget _buildPriceHistoryContent(List<PriceHistoryModel> priceHistory) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Current Prices Section
        // _buildCurrentPricesSection(),
        // const SizedBox(height: 16),

        // Price History Section Header
        Row(
          children: [
            Icon(
              Icons.history,
              size: 20,
              color: global.theme.inputTextBoxForceColor,
            ),
            SizedBox(width: 6),
            Text(
              global.language('product_details_price_history'),
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: global.theme.inputTextBoxForceColor,
              ),
            ),
            if (priceHistory.isNotEmpty) ...[
              const Spacer(),
              Container(
                padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 2),
                decoration: BoxDecoration(
                  color: Colors.blue[100],
                  borderRadius: BorderRadius.circular(10),
                ),
                child: Text(
                  '${priceHistory.length}',
                  style: TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w600,
                    color: Colors.blue[800],
                  ),
                ),
              ),
            ],
          ],
        ),
        const SizedBox(height: 12),

        // Price History Table
        if (priceHistory.isEmpty)
          _buildEmptyState()
        else
          Expanded(
            child: _buildCompactTable(priceHistory),
          ),
      ],
    );
  }

  Widget _buildEmptyState() {
    return Center(
      child: Padding(
        padding: EdgeInsets.all(32),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(
              Icons.inbox_outlined,
              size: 48,
              color: Colors.grey[400],
            ),
            SizedBox(height: 12),
            Text(
              global.language('product_details_no_history'),
              style: TextStyle(
                fontSize: 16,
                color: Colors.grey[600],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildCompactTable(List<PriceHistoryModel> priceHistory) {
    return Container(
      decoration: BoxDecoration(
        border: Border.all(color: Colors.grey[300]!),
        borderRadius: BorderRadius.circular(6),
      ),
      child: Column(
        children: [
          // Compact Header
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            decoration: BoxDecoration(
              color: Colors.grey[100],
              borderRadius: const BorderRadius.only(
                topLeft: Radius.circular(6),
                topRight: Radius.circular(6),
              ),
            ),
            child: Row(
              children: [
                const SizedBox(
                    width: 30,
                    child: Text('#',
                        style: TextStyle(
                            fontSize: 14, fontWeight: FontWeight.w600))),
                Expanded(
                    flex: 1,
                    child: Text(global.language('product_details_date_time'),
                        style: const TextStyle(
                            fontSize: 14, fontWeight: FontWeight.w600))),
                Expanded(
                    flex: 1,
                    child: Text(global.language('product_details_level'),
                        style: const TextStyle(
                            fontSize: 14, fontWeight: FontWeight.w600))),
                Expanded(
                    flex: 1,
                    child: Text(global.language('product_details_old'),
                        style: const TextStyle(
                            fontSize: 14, fontWeight: FontWeight.w600),
                        textAlign: TextAlign.right)),
                Expanded(
                    flex: 1,
                    child: Text(global.language('product_details_new'),
                        style: const TextStyle(
                            fontSize: 14, fontWeight: FontWeight.w600),
                        textAlign: TextAlign.right)),
                Expanded(
                    flex: 1,
                    child: Text(global.language('product_details_change'),
                        style: const TextStyle(
                            fontSize: 14, fontWeight: FontWeight.w600),
                        textAlign: TextAlign.center)),
                Expanded(
                    flex: 1,
                    child: Text(global.language('product_details_editor'),
                        style: const TextStyle(
                            fontSize: 14, fontWeight: FontWeight.w600))),
              ],
            ),
          ),

          // Compact Rows
          Expanded(
            child: ListView.builder(
              itemCount: priceHistory.length,
              itemBuilder: (context, index) {
                final history = priceHistory[index];
                final isEven = index % 2 == 0;
                final isLatest = index == 0;
                return _buildCompactRow(history, isEven, isLatest, index + 1);
              },
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildCompactRow(
      PriceHistoryModel history, bool isEven, bool isLatest, int rowNumber) {
    final isDecrease = (history.pricedifference ?? 0) < 0;
    final isIncrease = (history.pricedifference ?? 0) > 0;

    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
      decoration: BoxDecoration(
        color: isLatest
            ? Colors.amber[50]
            : (isEven ? Colors.grey[25] : Colors.white),
        border: Border(
          bottom: BorderSide(color: Colors.grey[200]!, width: 0.5),
          left: isLatest
              ? BorderSide(color: Colors.amber[400]!, width: 3)
              : BorderSide.none,
        ),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.center,
        children: [
          // ลำดับ
          SizedBox(
            width: 30,
            child: Row(
              children: [
                Text(
                  '$rowNumber',
                  style: TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.w600,
                    color: isLatest ? Colors.amber[800] : Colors.grey[700],
                  ),
                ),
              ],
            ),
          ),

          // วันที่และเวลา
          Expanded(
            flex: 1,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  _formatCompactDate(history.createdat),
                  style: const TextStyle(
                    fontSize: 13,
                    fontWeight: FontWeight.w600,
                  ),
                ),
                Text(
                  _formatTimeOnly(history.createdat),
                  style: TextStyle(
                    fontSize: 11,
                    color: Colors.grey[600],
                  ),
                ),
              ],
            ),
          ),

          // ระดับราคา
          Expanded(
            flex: 1,
            child: Text(
              _getShortPriceLevelName(history.keynumber ?? 0),
              style: TextStyle(
                fontSize: 12,
                color: Colors.blue[700],
                fontWeight: FontWeight.w500,
              ),
              maxLines: 1,
              overflow: TextOverflow.ellipsis,
            ),
          ),

          // ราคาเก่า
          Expanded(
            flex: 1,
            child: Text(
              _formatCompactPrice(history.oldprice),
              style: TextStyle(
                fontSize: 13,
                color: Colors.grey[600],
              ),
              textAlign: TextAlign.right,
            ),
          ),

          // ราคาใหม่
          Expanded(
            flex: 1,
            child: Text(
              _formatCompactPrice(history.newprice),
              style: const TextStyle(
                fontSize: 14,
                fontWeight: FontWeight.bold,
                color: Colors.black87,
              ),
              textAlign: TextAlign.right,
            ),
          ),

          // การเปลี่ยนแปลง
          Expanded(
            flex: 1,
            child: Row(
              mainAxisAlignment: MainAxisAlignment.center,
              children: [
                Icon(
                  isIncrease
                      ? Icons.arrow_upward
                      : isDecrease
                          ? Icons.arrow_downward
                          : Icons.remove,
                  color: isIncrease
                      ? Colors.green[600]
                      : isDecrease
                          ? Colors.red[600]
                          : Colors.grey[600],
                  size: 14,
                ),
                const SizedBox(width: 2),
                Text(
                  _formatCompactPrice(history.pricedifference?.abs()),
                  style: TextStyle(
                    fontSize: 11,
                    fontWeight: FontWeight.bold,
                    color: isIncrease
                        ? Colors.green[600]
                        : isDecrease
                            ? Colors.red[600]
                            : Colors.grey[600],
                  ),
                ),
              ],
            ),
          ),

          // ผู้แก้ไข
          Expanded(
            flex: 1,
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              mainAxisSize: MainAxisSize.min,
              children: [
                Text(
                  _getShortUserName(history.createdby),
                  style: const TextStyle(
                    fontSize: 12,
                    fontWeight: FontWeight.w500,
                  ),
                  maxLines: 1,
                  overflow: TextOverflow.ellipsis,
                ),
                if (history.remark?.isNotEmpty == true)
                  Row(
                    children: [
                      Icon(Icons.note, size: 10, color: Colors.orange[600]),
                      const SizedBox(width: 2),
                      Expanded(
                        child: Text(
                          history.remark!,
                          style: TextStyle(
                            fontSize: 10,
                            color: Colors.orange[700],
                            fontStyle: FontStyle.italic,
                          ),
                          maxLines: 1,
                          overflow: TextOverflow.ellipsis,
                        ),
                      ),
                    ],
                  ),
              ],
            ),
          ),
        ],
      ),
    );
  }

  String _formatCompactDate(String? dateStr) {
    if (dateStr == null) return '';
    try {
      final date = DateTime.parse(dateStr);
      return '${date.day}/${date.month}/${date.year + 543}';
    } catch (e) {
      return dateStr;
    }
  }

  String _formatCompactPrice(double? price) {
    if (price == null) return '0';
    if (price == price.toInt()) {
      return price.toInt().toString();
    }
    return price.toStringAsFixed(2);
  }

  String _getShortPriceLevelName(int keyNumber) {
    try {
      final priceLevel = global.config.prices.firstWhere(
        (price) => price.keyNumber == keyNumber,
      );
      final fullName = global.activeLangName(priceLevel.names
          .map((lang) => LanguageDataModel(code: lang.code!, name: lang.name!))
          .toList());

      // แสดงเฉพาะ 2-3 คำแรก
      final words = fullName.split(' ');
      if (words.length <= 2) return fullName;
      return '${words[0]} ${words[1]}';
    } catch (e) {
      return 'N/A';
    }
  }

  String _getShortUserName(String? username) {
    if (username == null || username.isEmpty) return global.language('product_details_not_specified');

    return username;
  }

  String _formatTimeOnly(String? dateStr) {
    if (dateStr == null) return '';
    try {
      final date = DateTime.parse(dateStr);
      return '${date.hour.toString().padLeft(2, '0')}:${date.minute.toString().padLeft(2, '0')} ${global.language("time_suffix")}';
    } catch (e) {
      return dateStr;
    }
  }

  Widget _buildCurrentPricesSection() {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        Row(
          children: [
            Icon(
              Icons.attach_money,
              size: 20,
              color: global.theme.inputTextBoxForceColor,
            ),
            SizedBox(width: 6),
            Text(
              global.language('product_details_current_price'),
              style: TextStyle(
                fontSize: 18,
                fontWeight: FontWeight.bold,
                color: global.theme.inputTextBoxForceColor,
              ),
            ),
          ],
        ),
        SizedBox(height: 12),
        Container(
          decoration: BoxDecoration(
            border: Border.all(color: Colors.green[300]!),
            borderRadius: BorderRadius.circular(6),
          ),
          child: Column(
            children: [
              // Header
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
                decoration: BoxDecoration(
                  color: Colors.green[50],
                  borderRadius: const BorderRadius.only(
                    topLeft: Radius.circular(6),
                    topRight: Radius.circular(6),
                  ),
                ),
                child: Row(
                  children: [
                    Expanded(
                      flex: 2,
                      child: Text(
                        global.language('product_details_price_level'),
                        style: const TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w600,
                        ),
                      ),
                    ),
                    Expanded(
                      flex: 1,
                      child: Text(
                        global.language('product_details_price_baht'),
                        style: const TextStyle(
                          fontSize: 14,
                          fontWeight: FontWeight.w600,
                        ),
                        textAlign: TextAlign.right,
                      ),
                    ),
                  ],
                ),
              ),

              // Price rows
              ..._buildCompactPriceRows(),
            ],
          ),
        ),
      ],
    );
  }

  List<Widget> _buildCompactPriceRows() {
    List<Widget> priceRows = [];
    int index = 0;

    // Loop through all available price levels from global config
    for (var priceConfig in global.config.prices) {
      if (priceConfig.isUse) {
        // Find matching price from product data
        double priceValue = 0.0;

        if (detailProductData?.prices != null) {
          try {
            final matchingPrice = detailProductData!.prices!.firstWhere(
              (price) => price.keynumber == priceConfig.keyNumber,
            );
            priceValue = matchingPrice.price;
          } catch (e) {
            // No matching price found, keep default 0.0
            priceValue = 0.0;
          }
        }

        final isEven = index % 2 == 0;
        priceRows.add(
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 8),
            decoration: BoxDecoration(
              color: isEven ? Colors.green[25] : Colors.white,
              border: Border(
                bottom: BorderSide(color: Colors.green[100]!, width: 0.5),
              ),
            ),
            child: Row(
              children: [
                Expanded(
                  flex: 2,
                  child: Row(
                    children: [
                      Container(
                        width: 6,
                        height: 6,
                        decoration: BoxDecoration(
                          color: priceValue > 0
                              ? Colors.green[500]
                              : Colors.grey[400],
                          shape: BoxShape.circle,
                        ),
                      ),
                      const SizedBox(width: 8),
                      Expanded(
                        child: Text(
                          _getShortPriceLevelName(priceConfig.keyNumber),
                          style: const TextStyle(
                            fontSize: 13,
                            fontWeight: FontWeight.w500,
                            color: Colors.black87,
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
                Expanded(
                  flex: 1,
                  child: Text(
                    '฿${_formatCompactPrice(priceValue)}',
                    style: TextStyle(
                      fontSize: 14,
                      fontWeight: FontWeight.bold,
                      color:
                          priceValue > 0 ? Colors.green[700] : Colors.grey[500],
                    ),
                    textAlign: TextAlign.right,
                  ),
                ),
              ],
            ),
          ),
        );
        index++;
      }
    }

    return priceRows;
  }
}
