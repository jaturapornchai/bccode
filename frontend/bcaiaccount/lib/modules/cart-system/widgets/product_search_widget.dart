import 'dart:async';
import 'dart:math' as math;
import 'package:flutter/material.dart';
import 'package:flutter_bloc/flutter_bloc.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import '../bloc/product_search_cubit.dart';
import '../bloc/product_search_state.dart';
import '../models/product_search_model.dart';
import '../models/product_card_data.dart';
import 'product_detail_dialog.dart';
import '../../../global.dart' as global;

/// โหมดการแสดงผลสินค้า
enum ProductViewMode {
  grid, // แสดงแบบตาราง (Wrap)
  list, // แสดงแบบรายการ (Column)
}

/// Widget สำหรับค้นหาสินค้าด้วย Backend Unified Search API
///
/// 🔍 **Features:**
/// - ✅ Cross-lingual Search: พิมพ์ไทย→เจออังกฤษ, พิมพ์อังกฤษ→เจอไทย
/// - ✅ Brand Recognition: มากิต้า→MAKITA, โบช→BOSCH อัตโนมัติ
/// - ✅ Hybrid Search: BM25(75%) + Vector(25%)
/// - ✅ Multi-field: Name, Barcode, Code, Model Number
/// - ✅ Smart Transliteration: แยกและแปลงแบรนด์อัตโนมัติ
///
/// 📋 **Search Types:**
/// - ชื่อสินค้า: "สว่าน", "drill"
/// - แบรนด์: "มากิต้า", "MAKITA"
/// - Barcode: "84032676"
/// - Product Code: "10317824"
/// - Model Number: "M6001B"
/// - คำผสม: "สว่านมากิต้า", "ค้อนโบช"
class ProductSearchWidget extends StatefulWidget {
  final String shopId;
  final Function(ProductSearchModel)? onProductSelected;
  final bool showPopularProducts;
  final int cardZoomLevel;
  final Function(int)? onZoomChanged;
  final VoidCallback? onCartUpdated; // Callback เมื่อ cart มีการเปลี่ยนแปลง
  final String? activeCartId; // Cart ID ที่กำลังใช้งาน

  const ProductSearchWidget({
    super.key,
    required this.shopId,
    this.onProductSelected,
    this.showPopularProducts = true,
    this.cardZoomLevel = 2,
    this.onZoomChanged,
    this.onCartUpdated,
    this.activeCartId,
  });

  @override
  State<ProductSearchWidget> createState() => _ProductSearchWidgetState();
}

class _ProductSearchWidgetState extends State<ProductSearchWidget> with AutomaticKeepAliveClientMixin {
  late TextEditingController _searchController;
  late FocusNode _searchFocusNode;
  late ValueNotifier<bool> _hasText;
  late ScrollController _scrollController; // สำหรับ infinite scroll

  // โหมดการแสดงผล (grid หรือ list)
  ProductViewMode _viewMode = ProductViewMode.grid;

  // Debounce timer สำหรับยกเลิก pending search
  Timer? _debounceTimer;

  @override
  bool get wantKeepAlive => true; // เก็บ state ไว้เมื่อสลับ tab

  // เก็บ widget ล่าสุดที่แสดงผลสำเร็จ เพื่อไม่ให้กระพริบ
  Widget? _lastSuccessfulWidget;

  // เก็บ units ทั้งหมดของแต่ละ itemcode (จาก Backend Unified API)
  // key = itemcode, value = List<ProductSearchModel> (ทุก barcode/unit)
  final Map<String, List<ProductSearchModel>> _itemCodeUnitsMap = {};

  // เก็บยอดคงเหลือของแต่ละ itemcode (จาก Backend Unified API)
  // key = itemcode, value = ProductBalanceModel
  final Map<String, ProductBalanceModel> _balancesMap = {};

  // เก็บ formatted balance string จาก Backend (เช่น "1 กล่อง x 2 โหล x 3 ชิ้น")
  // key = itemcode, value = formatted string
  final Map<String, String> _balanceFormattedMap = {};

  // หมายเหตุ: ข้อมูลทั้งหมดมาจาก Backend Unified API ในครั้งเดียว (1 API call)
  // ไม่มีการเรียก API เพิ่มเติมจาก PgSQLProductService

  // แสดงข้อความเมื่อพิมพ์แต่ละ whitespace อย่างเดียว
  bool _showEmptySearchMessage = false;

  // Config ขนาด Card แต่ละระดับ Zoom
  static const Map<int, double> _cardWidthConfig = {
    1: 220.0, // เล็กสุด
    2: 280.0, // ปกติ (default)
    3: 340.0, // กลาง
    4: 400.0, // ใหญ่
    5: 480.0, // ใหญ่สุด
  };

  @override
  void initState() {
    super.initState();
    _searchController = TextEditingController();
    _searchFocusNode = FocusNode();
    _hasText = ValueNotifier<bool>(false);
    _scrollController = ScrollController();

    // เพิ่ม listener สำหรับ infinite scroll
    _scrollController.addListener(_onScroll);

    // แสดงข้อความ "กรุณาระบุข้อความที่ต้องการค้นหา" ทันทีเมื่อเปิดหน้าจอ
    _showEmptySearchMessage = true;
  }

  @override
  void dispose() {
    _debounceTimer?.cancel();
    _searchController.dispose();
    _searchFocusNode.dispose();
    _hasText.dispose();
    _scrollController.removeListener(_onScroll);
    _scrollController.dispose();
    super.dispose();
  }

  /// ตรวจจับเมื่อเลื่อนใกล้ถึงล่างสุด แล้วโหลดเพิ่ม
  void _onScroll() {
    if (!_scrollController.hasClients) return;

    final maxScroll = _scrollController.position.maxScrollExtent;
    final currentScroll = _scrollController.position.pixels;
    final threshold = 200.0; // โหลดเพิ่มเมื่อเหลืออีก 200 pixels จะถึงล่างสุด

    if (currentScroll >= maxScroll - threshold) {
      // เรียก loadMore
      context.read<ProductSearchCubit>().loadMore(shopId: widget.shopId);
    }
  }

  void _onSearch(String keyword) {
    // trim หัวท้าย
    final trimmedKeyword = keyword.trim();

    if (trimmedKeyword.isEmpty) {
      // ถ้า trim แล้วว่าง → แสดงข้อความแจ้งเตือน "กรุณาระบุข้อความที่ต้องการค้นหา"
      setState(() {
        _showEmptySearchMessage = true;
      });
      return;
    } else {
      // ปิดข้อความแจ้งเตือน
      setState(() {
        _showEmptySearchMessage = false;
      });

      // เรียก search (debounce จัดการใน TextField.onChanged แล้ว)
      AppLogger.debug('🔍 Searching for: $trimmedKeyword');
      context.read<ProductSearchCubit>().universalSearch(shopId: widget.shopId, keyword: trimmedKeyword);
    }
  }

  void _onProductTap(ProductSearchModel product) async {
    // แสดง Product Detail Dialog และรอผลลัพธ์
    final result = await ProductDetailDialog.show(
      context: context,
      product: product,
      shopId: widget.shopId,
      activeCartId: widget.activeCartId, // ส่ง active cart ID
    );

    // ถ้าเพิ่มสินค้าสำเร็จ ให้ refresh cart icon
    if (result == true && mounted) {
      AppLogger.info('🛒 Product added to cart, refreshing cart icon');
      widget.onCartUpdated?.call();
    }

    // ยังคงเรียก callback ถ้ามี (สำหรับ use case อื่นๆ)
    if (widget.onProductSelected != null) {
      widget.onProductSelected!(product);
    }
  }

  /// Footer สำหรับ infinite scroll (loading indicator / scroll for more)
  Widget _buildScrollFooter(bool isLoadingMore, bool hasMore, int loadedCount, int displayTotal) {
    if (isLoadingMore) {
      return const Padding(
        padding: EdgeInsets.all(16),
        child: Center(child: CircularProgressIndicator(color: Color(0xFFEE4D2D))),
      );
    }
    if (hasMore) {
      return Padding(
        padding: const EdgeInsets.all(16),
        child: Center(
          child: Column(
            children: [
              Text(
                '${global.language("showing")} $loadedCount ${global.language("of")} $displayTotal ${global.language("items")}',
                style: TextStyle(color: Colors.grey[600], fontSize: 12, fontWeight: FontWeight.w500),
              ),
              const SizedBox(height: 4),
              Text(
                '${global.language("scroll_down_to_load_more")}...',
                style: TextStyle(color: Colors.grey[400], fontSize: 11),
              ),
            ],
          ),
        ),
      );
    }
    return const SizedBox.shrink();
  }

  @override
  Widget build(BuildContext context) {
    super.build(context); // ต้องเรียก super.build เพื่อให้ AutomaticKeepAliveClientMixin ทำงาน

    // ใช้ BlocBuilder เพื่อดึงข้อมูล total/count สำหรับแสดงใน header
    return BlocBuilder<ProductSearchCubit, ProductSearchState>(
      buildWhen: (previous, current) {
        // Rebuild เฉพาะเมื่อ state เปลี่ยนประเภท หรือ total/count เปลี่ยน
        if (previous.runtimeType != current.runtimeType) return true;
        if (previous is ProductSearchLoaded && current is ProductSearchLoaded) {
          return previous.total != current.total || previous.products.length != current.products.length;
        }
        if (previous is ProductSearchLoadingMore && current is ProductSearchLoadingMore) {
          return previous.total != current.total || previous.currentProducts.length != current.currentProducts.length;
        }
        return true;
      },
      builder: (context, headerState) {
        // ดึงข้อมูลสำหรับแสดงใน header
        int headerTotal = 0;
        int headerLoaded = 0;
        bool showStats = false;

        if (headerState is ProductSearchLoaded) {
          headerTotal = headerState.total > 0 ? headerState.total : headerState.products.length;
          headerLoaded = headerState.products.length;
          showStats = headerLoaded > 0;
        } else if (headerState is ProductSearchLoadingMore) {
          headerTotal = headerState.total > 0 ? headerState.total : headerState.currentProducts.length;
          headerLoaded = headerState.currentProducts.length;
          showStats = headerLoaded > 0;
        }

        return Column(
          children: [
            // Shopee-style Search bar with gradient background
            Container(
              decoration: const BoxDecoration(
                gradient: LinearGradient(colors: [Color(0xFFEE4D2D), Color(0xFFFF6347)], begin: Alignment.topLeft, end: Alignment.bottomRight),
              ),
              child: SafeArea(
                bottom: false,
                child: Padding(
                  padding: EdgeInsets.all(10),
                  child: Row(
                    children: [
                      // Search TextField (ขยายเต็มพื้นที่ที่เหลือ)
                      Expanded(
                        child: Container(
                          decoration: BoxDecoration(
                            color: Colors.white,
                            borderRadius: BorderRadius.circular(8),
                            boxShadow: [BoxShadow(color: Colors.black.withOpacity(0.1), blurRadius: 4, offset: const Offset(0, 2))],
                          ),
                          child: TextField(
                            controller: _searchController,
                            focusNode: _searchFocusNode,
                            decoration: InputDecoration(
                              hintText: global.language("search_product_hint"),
                              hintStyle: TextStyle(color: Colors.grey[400], fontSize: 14),
                              prefixIcon: Icon(Icons.search, color: Color(0xFFEE4D2D)),
                              suffixIcon: Row(
                                mainAxisSize: MainAxisSize.min,
                                children: [
                                  // Toggle View Mode Button
                                  IconButton(
                                    icon: Icon(_viewMode == ProductViewMode.grid ? Icons.list : Icons.apps, color: const Color(0xFFEE4D2D), size: 20),
                                    onPressed: () {
                                      setState(() {
                                        _viewMode = _viewMode == ProductViewMode.grid ? ProductViewMode.list : ProductViewMode.grid;
                                      });
                                    },
                                    tooltip: _viewMode == ProductViewMode.grid ? global.language("show_as_list") : global.language("show_as_grid"),
                                  ),
                                  // Clear/Scanner Button
                                  ValueListenableBuilder<bool>(
                                    valueListenable: _hasText,
                                    builder: (context, hasText, child) {
                                      return hasText
                                          ? IconButton(
                                              icon: Icon(Icons.clear, color: Colors.grey[400]),
                                              onPressed: () {
                                                _searchController.clear();
                                                _hasText.value = false;
                                                _onSearch('');
                                              },
                                            )
                                          : Icon(Icons.qr_code_scanner, color: Colors.grey[400]);
                                    },
                                  ),
                                ],
                              ),
                              border: InputBorder.none,
                              contentPadding: const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                            ),
                            onChanged: (value) {
                              _hasText.value = value.isNotEmpty;

                              // ✅ ยกเลิก timer เดิมก่อนสร้างใหม่
                              _debounceTimer?.cancel();

                              // สร้าง timer ใหม่ (300ms — เร็วพอให้รู้สึกตอบสนอง)
                              _debounceTimer = Timer(const Duration(milliseconds: 300), () {
                                if (_searchController.text == value && mounted) {
                                  _onSearch(value);
                                }
                              });
                            },
                            onSubmitted: (value) {
                              // ยกเลิก debounce timer และค้นหาทันที
                              _debounceTimer?.cancel();
                              _onSearch(value);
                            },
                          ),
                        ),
                      ),
                      // แสดงสถานะจำนวนรายการ (ด้านขวาของ textbox)
                      if (showStats)
                        Padding(
                          padding: const EdgeInsets.only(left: 10),
                          child: Container(
                            padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                            decoration: BoxDecoration(
                              color: Colors.white.withOpacity(0.2),
                              borderRadius: BorderRadius.circular(16),
                            ),
                            child: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                const Icon(Icons.inventory_2_outlined, size: 14, color: Colors.white),
                                const SizedBox(width: 4),
                                Text(
                                  '$headerTotal',
                                  style: const TextStyle(color: Colors.white, fontSize: 13, fontWeight: FontWeight.bold),
                                ),
                              ],
                            ),
                          ),
                        ),
                    ],
                  ),
                ),
              ),
            ),

            // Search results
            Expanded(
              child: BlocConsumer<ProductSearchCubit, ProductSearchState>(
            listener: (context, state) {
              // Debug: Log ทุกครั้งที่ state เปลี่ยน
              AppLogger.info('🔄 BlocListener received state: ${state.runtimeType}');

              // Auto-select if found by barcode
              if (state is ProductFoundByBarcode) {
                _onProductTap(state.product);
              }
              // ใช้ข้อมูลจาก Unified API โดยตรง (1 API call)
              else if (state is ProductSearchLoaded) {
                AppLogger.info('📦 ProductSearchLoaded: ${state.products.length} products, hasMore=${state.hasMore}');
                // ใช้ข้อมูลจาก Backend Unified API โดยตรง (ไม่ต้องเรียก API เพิ่มเติม)
                if (state.hasUnifiedData && mounted) {
                  setState(() {
                    _itemCodeUnitsMap.addAll(state.unitsMap!);
                    _balancesMap.addAll(state.balancesMap!);
                    if (state.balanceFormattedMap != null) {
                      _balanceFormattedMap.addAll(state.balanceFormattedMap!);
                    }
                  });
                  AppLogger.info('✅ Using Unified API data: ${state.unitsMap!.length} items (1 API call)');
                }
              }
              // กำลังโหลดเพิ่ม (infinite scroll) - อัพเดท maps ถ้ามี
              else if (state is ProductSearchLoadingMore) {
                AppLogger.info('⏳ ProductSearchLoadingMore: ${state.currentProducts.length} products');
                if (state.unitsMap != null && mounted) {
                  setState(() {
                    _itemCodeUnitsMap.addAll(state.unitsMap!);
                    if (state.balancesMap != null) {
                      _balancesMap.addAll(state.balancesMap!);
                    }
                    if (state.balanceFormattedMap != null) {
                      _balanceFormattedMap.addAll(state.balanceFormattedMap!);
                    }
                  });
                }
              }
            },
            builder: (context, state) {
              // แสดงเฉพาะผลลัพธ์ที่ได้จริงๆ ไม่แสดง loading, error หรือ initial state
              // เพื่อไม่ให้กระพริบ

              Widget? currentWidget;

              // แสดงข้อความเมื่อพิมพ์แต่ whitespace อย่างเดียว
              if (_showEmptySearchMessage) {
                return Center(
                  child: Padding(
                    padding: EdgeInsets.all(32),
                    child: Column(
                      mainAxisAlignment: MainAxisAlignment.center,
                      children: [
                        Icon(Icons.search_off, size: 64, color: Colors.grey[400]),
                        SizedBox(height: 16),
                        Text(
                          global.language("please_enter_search_text"),
                          style: TextStyle(fontSize: 16, color: Colors.grey[600], fontWeight: FontWeight.w500),
                          textAlign: TextAlign.center,
                        ),
                      ],
                    ),
                  ),
                );
              }

              // Auto-selected product - แสดงเฉพาะสินค้าที่เจอ
              if (state is ProductFoundByBarcode) {
                final product = state.product;
                AppLogger.info('🎯 ProductFoundByBarcode: ${product.itemCode} | Barcode: ${product.barcode}');

                // ดึง units จาก map (จะถูกโหลดใน listener แล้ว)
                final units = _itemCodeUnitsMap[product.itemCode] ?? [];
                // ดึงยอดคงเหลือจาก map
                final balance = _balancesMap[product.itemCode];

                currentWidget = SingleChildScrollView(
                  physics: const BouncingScrollPhysics(),
                  child: Align(
                    alignment: Alignment.topCenter,
                    child: Padding(
                      padding: const EdgeInsets.all(8),
                      child: LayoutBuilder(
                        builder: (context, constraints) {
                          // สำหรับ barcode เดียว ให้ความกว้างพอดีแต่ไม่เกิน 400
                          final cardWidth = constraints.maxWidth.clamp(280.0, 400.0);

                          return SizedBox(
                            width: cardWidth,
                            child: _ShopeeProductCard(
                              cardData: ProductCardData.fromData(
                                product: product,
                                allUnits: units,
                                balance: balance,
                                balanceFormatted: _balanceFormattedMap[product.itemCode],
                              ),
                              onTap: () => _onProductTap(product),
                              shopId: widget.shopId,
                              onProductSelected: _onProductTap,
                            ),
                          );
                        },
                      ),
                    ),
                  ),
                );
              }
              // แสดงผลลัพธ์การค้นหา (รองรับทั้ง ProductSearchLoaded และ ProductSearchLoadingMore)
              else if (state is ProductSearchLoaded || state is ProductSearchLoadingMore) {
                // ดึง products จาก state ที่เหมาะสม
                final List<ProductSearchModel> products;
                final bool isLoadingMore;
                final bool hasMore;
                final int total;

                if (state is ProductSearchLoadingMore) {
                  products = state.currentProducts;
                  isLoadingMore = true;
                  hasMore = true; // กำลังโหลดอยู่ แสดงว่ายังมีอีก
                  total = state.total;
                } else {
                  products = (state as ProductSearchLoaded).products;
                  isLoadingMore = false;
                  hasMore = state.hasMore;
                  total = state.total;
                }

                AppLogger.info('🎨 [Widget] Rendering ${products.length} product cards (total: $total, isLoadingMore: $isLoadingMore, hasMore: $hasMore)');

                if (products.isEmpty) {
                  currentWidget = Center(
                    child: Padding(
                      padding: EdgeInsets.all(32),
                      child: Column(
                        mainAxisAlignment: MainAxisAlignment.center,
                        children: [
                          Icon(Icons.inventory_2_outlined, size: 80, color: Colors.grey[300]),
                          SizedBox(height: 16),
                          Text(global.language("no_products_yet"), style: TextStyle(fontSize: 18, color: Colors.grey[600])),
                        ],
                      ),
                    ),
                  );
                } else {
                  // สร้าง ProductCardData - 1 card ต่อ 1 itemcode
                  final seenItemCodes = <String>{};
                  final productDataList = <ProductCardData>[];

                  for (final product in products) {
                    if (!seenItemCodes.contains(product.itemCode)) {
                      seenItemCodes.add(product.itemCode);
                      productDataList.add(
                        ProductCardData.fromData(
                          product: product,
                          allUnits: _itemCodeUnitsMap[product.itemCode] ?? [],
                          balance: _balancesMap[product.itemCode],
                          balanceFormatted: _balanceFormattedMap[product.itemCode],
                        ),
                      );
                    }
                  }

                  final displayTotal = total > 0 ? total : productDataList.length;

                  // ใช้ LayoutBuilder เพื่อคำนวณ responsive columns
                  currentWidget = LayoutBuilder(
                    builder: (context, constraints) {
                      // หัก padding ของ ListView (8 ซ้าย + 8 ขวา = 16)
                      final availableWidth = constraints.maxWidth - 16;
                      final minCardWidth = _cardWidthConfig[widget.cardZoomLevel] ?? 280.0;
                      final maxCardWidth = minCardWidth + 100;
                      const spacing = 1.0;

                      int columns = (availableWidth / (minCardWidth + spacing)).floor();
                      if (columns < 1) columns = 1;

                      double cardWidth = (availableWidth - (spacing * (columns - 1))) / columns;
                      if (cardWidth > maxCardWidth) cardWidth = maxCardWidth;

                      if (_viewMode == ProductViewMode.list) {
                        // List view — virtualized ListView.builder
                        final int itemCount = productDataList.length + (isLoadingMore || hasMore ? 1 : 0);
                        return ListView.builder(
                          controller: _scrollController,
                          physics: const AlwaysScrollableScrollPhysics(),
                          padding: const EdgeInsets.all(8),
                          itemCount: itemCount,
                          itemBuilder: (context, index) {
                            if (index >= productDataList.length) {
                              return _buildScrollFooter(isLoadingMore, hasMore, productDataList.length, displayTotal);
                            }
                            final cardData = productDataList[index];
                            return Padding(
                              padding: const EdgeInsets.only(bottom: 8),
                              child: RepaintBoundary(
                                child: _AnimatedCardEntry(
                                  index: index,
                                  child: _ShopeeProductListCard(
                                    key: ValueKey('${cardData.itemCode}_${cardData.barcode}'),
                                    cardData: cardData,
                                    onTap: () => _onProductTap(cardData.product),
                                    shopId: widget.shopId,
                                    onProductSelected: _onProductTap,
                                  ),
                                ),
                              ),
                            );
                          },
                        );
                      } else {
                        // Grid view — virtualized row-based ListView.builder
                        final int rowCount = (productDataList.length / columns).ceil();
                        final int totalRows = rowCount + (isLoadingMore || hasMore ? 1 : 0);
                        return ListView.builder(
                          controller: _scrollController,
                          physics: const AlwaysScrollableScrollPhysics(),
                          padding: const EdgeInsets.all(8),
                          itemCount: totalRows,
                          itemBuilder: (context, rowIndex) {
                            if (rowIndex >= rowCount) {
                              return _buildScrollFooter(isLoadingMore, hasMore, productDataList.length, displayTotal);
                            }
                            final startIdx = rowIndex * columns;
                            final endIdx = math.min(startIdx + columns, productDataList.length);
                            final rowItems = productDataList.sublist(startIdx, endIdx);

                            return Padding(
                              padding: EdgeInsets.only(bottom: spacing),
                              child: Row(
                                crossAxisAlignment: CrossAxisAlignment.start,
                                children: [
                                  for (int i = 0; i < rowItems.length; i++) ...[
                                    if (i > 0) SizedBox(width: spacing),
                                    SizedBox(
                                      width: cardWidth,
                                      child: RepaintBoundary(
                                        child: _AnimatedCardEntry(
                                          index: startIdx + i,
                                          child: _ShopeeProductCard(
                                            key: ValueKey('${rowItems[i].itemCode}_${rowItems[i].barcode}'),
                                            cardData: rowItems[i],
                                            onTap: () => _onProductTap(rowItems[i].product),
                                            shopId: widget.shopId,
                                            onProductSelected: _onProductTap,
                                          ),
                                        ),
                                      ),
                                    ),
                                  ],
                                ],
                              ),
                            );
                          },
                        );
                      }
                    },
                  );
                }
              }

              // ถ้ามี widget ใหม่ ให้เก็บไว้
              if (currentWidget != null) {
                _lastSuccessfulWidget = currentWidget;
              }

              // Loading state: แสดง shimmer skeleton + progress indicator
              if (state is ProductSearchLoading) {
                if (_lastSuccessfulWidget != null) {
                  // มีผลลัพธ์เดิม → แสดง faded + progress bar ด้านบน
                  return Column(
                    children: [
                      const LinearProgressIndicator(
                        color: Color(0xFFEE4D2D),
                        backgroundColor: Color(0xFFFFE0D6),
                      ),
                      Expanded(
                        child: AnimatedOpacity(
                          opacity: 0.4,
                          duration: const Duration(milliseconds: 200),
                          child: IgnorePointer(child: _lastSuccessfulWidget!),
                        ),
                      ),
                    ],
                  );
                } else {
                  // ค้นหาครั้งแรก → แสดง shimmer skeleton
                  return const _ShimmerSkeletonGrid();
                }
              }

              // แสดง widget ล่าสุด หรือ widget ใหม่
              return _lastSuccessfulWidget ?? const SizedBox.shrink();
            },
          ),
        ),
        ],
      );
      },
    );
  }
}

/// Shopee-style Product Card with Image (Grid view - แนวตั้ง)
/// ข้อมูล warehouse/location มาจาก cardData.warehouses (Backend API) โดยตรง
class _ShopeeProductCard extends StatefulWidget {
  final ProductCardData cardData;
  final VoidCallback onTap;
  final String shopId;
  final Function(ProductSearchModel)? onProductSelected;

  const _ShopeeProductCard({
    super.key,
    required this.cardData,
    required this.onTap,
    required this.shopId,
    this.onProductSelected,
  });

  @override
  State<_ShopeeProductCard> createState() => _ShopeeProductCardState();
}

class _ShopeeProductCardState extends State<_ShopeeProductCard> {
  /// Helper function: เรียงลำดับ units ตามอัตราส่วน (มาก→น้อย) แล้วตาม barcode (A→Z)
  List<ProductSearchModel> _sortUnitsByRatioAndBarcode(List<ProductSearchModel> units) {
    final sortedUnits = List<ProductSearchModel>.from(units);
    sortedUnits.sort((a, b) {
      // 1. เรียงตามอัตราส่วน (มาก→น้อย)
      final ratioA = a.unitStand / a.unitDive;
      final ratioB = b.unitStand / b.unitDive;
      final ratioCompare = ratioB.compareTo(ratioA);
      if (ratioCompare != 0) return ratioCompare;

      // 2. ถ้าอัตราส่วนเท่ากัน เรียงตาม barcode (A→Z)
      return a.barcode.compareTo(b.barcode);
    });
    return sortedUnits;
  }

  /// Helper function: แสดงตัวเลขแบบ smart - ถ้าเป็นจำนวนเต็มไม่แสดงทศนิยม
  String _formatNumber(double value, {int decimals = 2}) {
    // ตรวจสอบว่าเป็นจำนวนเต็มหรือใกล้เคียงจำนวนเต็ม
    final floorValue = value.floor().toDouble();
    if ((value - floorValue).abs() < 0.0001) {
      // เป็นจำนวนเต็ม → แสดงเป็นเลขเต็ม
      return floorValue.toInt().toString();
    } else {
      // มีทศนิยม → แสดงทศนิยม แล้วตัด trailing zeros
      return value.toStringAsFixed(decimals).replaceAll(RegExp(r'0+$'), '').replaceAll(RegExp(r'\.$'), '');
    }
  }

  // Generate color based on itemCode
  // สร้างสีที่แตกต่างกันสำหรับแต่ละ itemCode เพื่อให้เห็นเป็นกลุ่มชัดเจน
  Color _getItemCodeColor(String itemCode) {
    if (itemCode.isEmpty) return Colors.blue;

    // Hash the itemCode to get consistent color
    int hash = 0;
    for (int i = 0; i < itemCode.length; i++) {
      hash = itemCode.codeUnitAt(i) + ((hash << 5) - hash);
    }

    // ชุดสีที่เด่นชัดและแตกต่างกันมาก (24 สี)
    final colorPalette = [
      Colors.blue[400]!, // น้ำเงิน
      Colors.green[400]!, // เขียว
      Colors.orange[400]!, // ส้ม
      Colors.purple[400]!, // ม่วง
      Colors.pink[400]!, // ชมพู
      Colors.teal[400]!, // เขียวอมฟ้า
      Colors.indigo[400]!, // คราม
      Colors.cyan[400]!, // ฟ้าอมเขียว
      Colors.red[400]!, // แดง
      Colors.amber[600]!, // เหลืองทอง
      Colors.deepOrange[400]!, // ส้มเข้ม
      Colors.lightGreen[500]!, // เขียวอ่อน
      Colors.deepPurple[400]!, // ม่วงเข้ม
      Colors.lightBlue[400]!, // ฟ้าอ่อน
      Colors.lime[700]!, // เหลืองมะนาว
      Colors.brown[400]!, // น้ำตาล
      Colors.blueGrey[400]!, // เทาน้ำเงิน
      Colors.pinkAccent[200]!, // ชมพูสด
      Colors.tealAccent[400]!, // เขียวมิ้นท์
      Colors.purpleAccent[200]!, // ม่วงสด
      Colors.orangeAccent[400]!, // ส้มสด
      Colors.greenAccent[400]!, // เขียวสด
      Colors.blueAccent[400]!, // น้ำเงินสด
      Colors.redAccent[400]!, // แดงสด
    ];

    // เลือกสีจาก palette ตาม hash
    final colorIndex = hash.abs() % colorPalette.length;
    return colorPalette[colorIndex];
  }

  /// สร้าง Widget แสดงยอดคงเหลือแยกตาม WH/Location
  /// ข้อมูลมาจาก Backend API โดยตรง (widget.cardData.warehouses)
  Widget _buildWarehouseLocationBalances() {
    // ถ้าไม่มียอดคงเหลือ หรือไม่มี warehouse data ไม่ต้องแสดง
    if (!widget.cardData.hasBalance || !widget.cardData.hasWarehouses) {
      return const SizedBox.shrink();
    }

    // หาหน่วย 1:1 สำหรับแสดงชื่อหน่วย
    final baseUnit = widget.cardData.allUnits.firstWhere(
      (u) => u.unitStand == 1 && u.unitDive == 1,
      orElse: () => widget.cardData.product,
    );
    final unitName = baseUnit.unitName.isNotEmpty ? baseUnit.unitName : baseUnit.unitCode;

    return Container(
      margin: const EdgeInsets.only(top: 4),
      padding: const EdgeInsets.all(6),
      decoration: BoxDecoration(
        color: Colors.indigo[50],
        borderRadius: BorderRadius.circular(4),
        border: Border.all(color: Colors.indigo[200]!, width: 0.5),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          // Header
          Row(
            children: [
              Icon(Icons.warehouse, size: 12, color: Colors.indigo[700]),
              SizedBox(width: 4),
              Text(
                global.language("warehouse_location"),
                style: TextStyle(fontSize: 9, fontWeight: FontWeight.bold, color: Colors.indigo[900]),
              ),
            ],
          ),
          const SizedBox(height: 4),

          // รายการ WH/Location จาก Backend API
          ...widget.cardData.warehouses.map((warehouse) {
            return Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                // Warehouse row
                Container(
                  margin: const EdgeInsets.only(bottom: 2),
                  padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
                  decoration: BoxDecoration(
                    color: Colors.blue[100],
                    borderRadius: BorderRadius.circular(3),
                  ),
                  child: Row(
                    children: [
                      Icon(Icons.store, size: 10, color: Colors.blue[800]),
                      SizedBox(width: 4),
                      Expanded(
                        child: Text(
                          warehouse.warehouseCode.isNotEmpty ? warehouse.warehouseCode : global.language("default_warehouse"),
                          style: TextStyle(fontSize: 9, fontWeight: FontWeight.w600, color: Colors.blue[900]),
                        ),
                      ),
                      // แสดงยอดคงเหลือ 2 แบบ: 1:1 unit และ word format
                      Column(
                        crossAxisAlignment: CrossAxisAlignment.end,
                        children: [
                          // 1. ยอดคงเหลือตามหน่วย 1:1
                          Text(
                            '${_formatNumber(warehouse.balanceQty)} $unitName',
                            style: TextStyle(fontSize: 9, fontWeight: FontWeight.bold, color: Colors.blue[900]),
                          ),
                          // 2. ยอดคงเหลือแบบ word (ถ้ามี)
                          if (warehouse.balanceWord.isNotEmpty)
                            Text(
                              warehouse.balanceWord,
                              style: TextStyle(fontSize: 8, color: Colors.blue[700]),
                            ),
                        ],
                      ),
                    ],
                  ),
                ),

                // Location rows (ถ้ามี)
                if (warehouse.hasLocations)
                  ...warehouse.locations.where((loc) => loc.balanceQty > 0).map((location) {
                    return Container(
                      margin: const EdgeInsets.only(left: 12, bottom: 2),
                      padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
                      decoration: BoxDecoration(
                        color: Colors.green[50],
                        borderRadius: BorderRadius.circular(3),
                      ),
                      child: Row(
                        children: [
                          Icon(Icons.location_on, size: 10, color: Colors.green[700]),
                          SizedBox(width: 4),
                          Expanded(
                            child: Text(
                              location.locationCode.isNotEmpty ? location.locationCode : global.language("general_location"),
                              style: TextStyle(fontSize: 8, color: Colors.green[800]),
                            ),
                          ),
                          // แสดงยอดคงเหลือ 2 แบบ: 1:1 unit และ word format
                          Column(
                            crossAxisAlignment: CrossAxisAlignment.end,
                            children: [
                              // 1. ยอดคงเหลือตามหน่วย 1:1
                              Text(
                                '${_formatNumber(location.balanceQty)} $unitName',
                                style: TextStyle(fontSize: 8, fontWeight: FontWeight.w600, color: Colors.green[900]),
                              ),
                              // 2. ยอดคงเหลือแบบ word (ถ้ามี)
                              if (location.balanceWord.isNotEmpty)
                                Text(
                                  location.balanceWord,
                                  style: TextStyle(fontSize: 7, color: Colors.green[700]),
                                ),
                            ],
                          ),
                        ],
                      ),
                    );
                  }),
              ],
            );
          }),
        ],
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final itemColor = _getItemCodeColor(widget.cardData.product.itemCode);
    final itemColorDark = HSLColor.fromColor(itemColor).withLightness(0.4).toColor(); // ปรับเป็น 0.4 เพื่อให้เข้มขึ้น

    return Card(
      elevation: 2,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(8),
        side: BorderSide(color: itemColorDark, width: 3), // เพิ่มความหนาเป็น 3
      ),
      child: InkWell(
        onTap: widget.onTap,
        borderRadius: BorderRadius.circular(8),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            // Product Image - แสดงเฉพาะเมื่อมีรูปภาพ (รูปเต็ม 100%)
            if (widget.cardData.product.imageuri != null && widget.cardData.product.imageuri!.isNotEmpty)
              AspectRatio(
                aspectRatio: 1.0, // สี่เหลี่ยมจัตุรัส ให้รูปแสดงเต็ม
                child: Container(
                  width: double.infinity,
                  decoration: BoxDecoration(
                    color: Colors.grey[100],
                    borderRadius: const BorderRadius.vertical(top: Radius.circular(8)),
                  ),
                  child: Stack(
                    children: [
                      // รูปภาพจริงจาก imageuri - ไม่กำหนด height ให้แสดงเต็ม
                      ClipRRect(
                        borderRadius: const BorderRadius.vertical(top: Radius.circular(8)),
                        child: Image.network(
                          widget.cardData.product.imageuri!,
                          width: double.infinity,
                          fit: BoxFit.fitHeight,
                          loadingBuilder: (context, child, loadingProgress) {
                            if (loadingProgress == null) return child;
                            // แสดง placeholder ระหว่างโหลด
                            return Container(
                              decoration: BoxDecoration(
                                gradient: LinearGradient(begin: Alignment.topLeft, end: Alignment.bottomRight, colors: [itemColor.withOpacity(0.3), itemColor.withOpacity(0.6)]),
                              ),
                              child: Center(
                                child: CircularProgressIndicator(
                                  value: loadingProgress.expectedTotalBytes != null ? loadingProgress.cumulativeBytesLoaded / loadingProgress.expectedTotalBytes! : null,
                                  strokeWidth: 2,
                                  color: itemColorDark,
                                ),
                              ),
                            );
                          },
                          errorBuilder: (context, error, stackTrace) {
                            // โหลดรูปไม่สำเร็จ ไม่แสดงอะไร
                            return const SizedBox.shrink();
                          },
                        ),
                      ),
                      // Item Code Badge (top right)
                      if (widget.cardData.product.itemCode.isNotEmpty)
                        Positioned(
                          top: 8,
                          right: 8,
                          child: Container(
                            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                            decoration: BoxDecoration(color: itemColorDark, borderRadius: BorderRadius.circular(4)),
                            child: Text(
                              widget.cardData.product.itemCode,
                              style: const TextStyle(color: Colors.white, fontSize: 9, fontWeight: FontWeight.bold),
                            ),
                          ),
                        ),
                      // Badge for special offer (if unitStand > 1)
                      if (widget.cardData.product.unitStand > 1)
                        Positioned(
                          top: 8,
                          left: 8,
                          child: Container(
                            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 2),
                            decoration: BoxDecoration(color: Colors.red, borderRadius: BorderRadius.circular(4)),
                            child: Text(
                              'Pack ${widget.cardData.product.unitStand}',
                              style: const TextStyle(color: Colors.white, fontSize: 10, fontWeight: FontWeight.bold),
                            ),
                          ),
                        ),
                    ],
                  ),
                ),
              ),

            // Product Info
            Padding(
              padding: const EdgeInsets.all(8),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                mainAxisSize: MainAxisSize.min,
                children: [
                  // Product Name
                  Text(
                    widget.cardData.product.name0,
                    style: const TextStyle(fontSize: 13, height: 1.2, fontWeight: FontWeight.bold),
                    maxLines: 2,
                    overflow: TextOverflow.ellipsis,
                  ),
                  const SizedBox(height: 4),

                  // Item Code & Barcode & Ratio
                  if (widget.cardData.product.itemCode.isNotEmpty)
                    Row(
                      children: [
                        Container(
                          width: 3,
                          height: 12,
                          decoration: BoxDecoration(color: itemColorDark, borderRadius: BorderRadius.circular(1.5)),
                        ),
                        SizedBox(width: 4),
                        Expanded(
                          child: Text(
                            '${global.language("code")}: ${widget.cardData.product.itemCode}',
                            style: TextStyle(fontSize: 9, color: itemColorDark, fontWeight: FontWeight.w600),
                            maxLines: 1,
                            overflow: TextOverflow.ellipsis,
                          ),
                        ),
                      ],
                    ),

                  // ยอดคงเหลือ (แสดงจำนวนรวมของ itemcode)
                  if (widget.cardData.hasBalance)
                    Container(
                      margin: const EdgeInsets.only(top: 4),
                      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 4),
                      decoration: BoxDecoration(
                        color: Colors.green[50],
                        borderRadius: BorderRadius.circular(4),
                        border: Border.all(color: Colors.green[300]!, width: 1),
                      ),
                      child: Row(
                        children: [
                          Icon(Icons.inventory_2, size: 14, color: Colors.green[700]),
                          const SizedBox(width: 4),
                          Expanded(
                            child: Builder(
                              builder: (context) {
                                // หาหน่วย 1:1 (unitStand = 1, unitDivide = 1)
                                final baseUnit = widget.cardData.allUnits.firstWhere((u) => u.unitStand == 1 && u.unitDive == 1, orElse: () => widget.cardData.product);

                                final baseUnitName = baseUnit.unitName.isNotEmpty ? baseUnit.unitName : baseUnit.unitCode;

                                // ใช้ formatted string จาก Backend (ถ้ามี) หรือคำนวณเอง
                                String multiLevelPacking;
                                if (widget.cardData.hasFormattedBalance) {
                                  // ใช้ค่าที่ Backend คำนวณมาแล้ว
                                  multiLevelPacking = widget.cardData.balanceFormatted!;
                                } else {
                                  // Fallback: คำนวณเองจาก units
                                  final balanceModel = ProductBalanceModel(
                                    itemCode: widget.cardData.itemCode,
                                    totalBalance: widget.cardData.totalBalance,
                                    balanceByShop: widget.cardData.balanceByShop,
                                  );
                                  multiLevelPacking = balanceModel.formatWithMultiLevelPacking(allUnits: widget.cardData.allUnits);
                                }

                                return Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    // บรรทัดแรก: ยอดคงเหลือแบบ 1:1 (ปกติ)
                                    Text(
                                      '${global.language("balance_remaining")}: ${_formatNumber(widget.cardData.totalBalance)} $baseUnitName',
                                      style: TextStyle(fontSize: 10, color: Colors.green[900], fontWeight: FontWeight.bold),
                                    ),

                                    // บรรทัดที่สอง: ยอดคงเหลือแบบ Multi-Level Packing
                                    // แสดงเฉพาะถ้ามี formatted balance หรือมีหลายหน่วย
                                    if (widget.cardData.hasFormattedBalance || widget.cardData.allUnits.length > 1) ...[
                                      const SizedBox(height: 2),
                                      Text(
                                        multiLevelPacking,
                                        style: TextStyle(fontSize: 9, color: Colors.green[700], fontStyle: FontStyle.italic),
                                      ),
                                    ],
                                  ],
                                );
                              },
                            ),
                          ),
                        ],
                      ),
                    ),

                  // ยอดค้างรับ / ค้างส่ง
                  if (widget.cardData.hasPending)
                    Container(
                      margin: const EdgeInsets.only(top: 4),
                      padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 4),
                      decoration: BoxDecoration(
                        color: Colors.orange[50],
                        borderRadius: BorderRadius.circular(4),
                        border: Border.all(color: Colors.orange[300]!, width: 1),
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          if (widget.cardData.hasPendingRecv)
                            Row(
                              children: [
                                Icon(Icons.call_received, size: 12, color: Colors.blue[700]),
                                const SizedBox(width: 4),
                                Expanded(
                                  child: Text(
                                    '${global.language("pending_receive")}: ${widget.cardData.pendingRecvWord.isNotEmpty ? widget.cardData.pendingRecvWord : _formatNumber(widget.cardData.pendingRecvQty)}',
                                    style: TextStyle(fontSize: 10, color: Colors.blue[900], fontWeight: FontWeight.w600),
                                  ),
                                ),
                              ],
                            ),
                          if (widget.cardData.hasPendingRecv && widget.cardData.hasPendingSend)
                            const SizedBox(height: 2),
                          if (widget.cardData.hasPendingSend)
                            Row(
                              children: [
                                Icon(Icons.call_made, size: 12, color: Colors.red[700]),
                                const SizedBox(width: 4),
                                Expanded(
                                  child: Text(
                                    '${global.language("pending_delivery")}: ${widget.cardData.pendingSendWord.isNotEmpty ? widget.cardData.pendingSendWord : _formatNumber(widget.cardData.pendingSendQty)}',
                                    style: TextStyle(fontSize: 10, color: Colors.red[900], fontWeight: FontWeight.w600),
                                  ),
                                ),
                              ],
                            ),
                        ],
                      ),
                    ),

                  // ยอดคงเหลือแยกตาม WH/Location
                  _buildWarehouseLocationBalances(),

                  // หน่วยนับทั้งหมด (เรียงตามอัตราส่วนมากไปน้อย)
                  if (widget.cardData.allUnits.isNotEmpty)
                    Container(
                      margin: const EdgeInsets.only(top: 4),
                      padding: const EdgeInsets.all(4),
                      decoration: BoxDecoration(
                        color: Colors.blue[50],
                        borderRadius: BorderRadius.circular(4),
                        border: Border.all(color: Colors.blue[200]!, width: 0.5),
                      ),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          // รายการหน่วยนับพร้อม barcode (เรียงตามอัตราส่วน + barcode)
                          ...(_sortUnitsByRatioAndBarcode(widget.cardData.allUnits).map((unit) {
                            final isCurrentUnit = unit.barcode == widget.cardData.product.barcode;
                            return _BarcodeRowWithHover(
                              unit: unit,
                              isCurrentUnit: isCurrentUnit,
                              onTap: () {
                                // เรียก callback เพื่อเพิ่มในตระกร้า
                                if (widget.onProductSelected != null) {
                                  widget.onProductSelected!(unit);
                                }
                              },
                            );
                          }).toList()),
                        ],
                      ),
                    ),

                  // Unit info (ซ่อนเพราะมีใน หน่วยนับทั้งหมดแล้ว)
                  /*if (widget.cardData.product.unitCode.isNotEmpty ||
                        widget.cardData.product.unitName.isNotEmpty)
                      Row(
                        children: [
                          if (widget.cardData.product.unitCode.isNotEmpty)
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 4,
                                vertical: 2,
                              ),
                              decoration: BoxDecoration(
                                color: Colors.blue[50],
                                borderRadius: BorderRadius.circular(3),
                              ),
                              child: Text(
                                widget.cardData.product.unitCode,
                                style: TextStyle(
                                  fontSize: 8,
                                  color: Colors.blue[700],
                                  fontWeight: FontWeight.w500,
                                ),
                              ),
                            ),
                          if (widget.cardData.product.unitCode.isNotEmpty &&
                              widget.cardData.product.unitStand > 1)
                            const SizedBox(width: 4),
                          if (widget.cardData.product.unitStand > 1)
                            Container(
                              padding: const EdgeInsets.symmetric(
                                horizontal: 4,
                                vertical: 2,
                              ),
                              decoration: BoxDecoration(
                                color: Colors.orange[50],
                                borderRadius: BorderRadius.circular(3),
                              ),
                              child: Text(
                                '${widget.cardData.product.unitStand}:${widget.cardData.product.unitDive}',
                                style: TextStyle(
                                  fontSize: 8,
                                  color: Colors.orange[700],
                                  fontWeight: FontWeight.w500,
                                ),
                              ),
                            ),
                        ],
                      ),*/
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }
}

/// Widget สำหรับแสดงแถว barcode พร้อม hover effect
class _BarcodeRowWithHover extends StatefulWidget {
  final ProductSearchModel unit;
  final bool isCurrentUnit;
  final VoidCallback onTap;

  const _BarcodeRowWithHover({required this.unit, required this.isCurrentUnit, required this.onTap});

  @override
  State<_BarcodeRowWithHover> createState() => _BarcodeRowWithHoverState();
}

class _BarcodeRowWithHoverState extends State<_BarcodeRowWithHover> {
  bool _isHovered = false;

  @override
  Widget build(BuildContext context) {
    return MouseRegion(
      onEnter: (_) => setState(() => _isHovered = true),
      onExit: (_) => setState(() => _isHovered = false),
      cursor: SystemMouseCursors.click,
      child: GestureDetector(
        onTap: widget.onTap,
        child: Container(
          margin: const EdgeInsets.only(bottom: 2),
          padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 3),
          decoration: BoxDecoration(color: _isHovered ? Colors.blue[100] : Colors.transparent, borderRadius: BorderRadius.circular(3)),
          child: Row(
            children: [
              // อัตราส่วนและหน่วย
              Text(
                '${global.formatNumberRemoveRightZero(widget.unit.unitStand)}:${global.formatNumberRemoveRightZero(widget.unit.unitDive)} ${widget.unit.unitName}',
                style: TextStyle(fontSize: 9, color: Colors.blue[900], fontWeight: widget.isCurrentUnit ? FontWeight.bold : FontWeight.w500),
              ),
              const SizedBox(width: 6),
              // Barcode (ไม่มี background)
              Text(
                widget.unit.barcode,
                style: TextStyle(fontSize: 8, color: Colors.grey[700], fontFamily: 'monospace'),
              ),
              // ใช้ Spacer เพื่อผลักราคาไปทางขวา
              const Spacer(),
              // ราคา (ใช้ราคาขายปลีก ถ้าไม่มีใช้ price1)
              Text(
                global.formatPrice(widget.unit.priceRetail > 0 ? widget.unit.priceRetail : widget.unit.price1),
                style: TextStyle(fontSize: 9, color: Colors.green[700], fontWeight: FontWeight.bold),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

/// Shopee-style Product Card สำหรับ List view (แนวนอน)
class _ShopeeProductListCard extends StatefulWidget {
  final ProductCardData cardData;
  final VoidCallback onTap;
  final String shopId;
  final Function(ProductSearchModel)? onProductSelected;

  const _ShopeeProductListCard({super.key, required this.cardData, required this.onTap, required this.shopId, this.onProductSelected});

  @override
  State<_ShopeeProductListCard> createState() => _ShopeeProductListCardState();
}

class _ShopeeProductListCardState extends State<_ShopeeProductListCard> {
  /// Helper function: เรียงลำดับ units ตามอัตราส่วน (มาก→น้อย) แล้วตาม barcode (A→Z)
  List<ProductSearchModel> _sortUnitsByRatioAndBarcode(List<ProductSearchModel> units) {
    final sortedUnits = List<ProductSearchModel>.from(units);
    sortedUnits.sort((a, b) {
      // 1. เรียงตามอัตราส่วน (มาก→น้อย)
      final ratioA = a.unitStand / a.unitDive;
      final ratioB = b.unitStand / b.unitDive;
      final ratioCompare = ratioB.compareTo(ratioA);
      if (ratioCompare != 0) return ratioCompare;

      // 2. ถ้าอัตราส่วนเท่ากัน เรียงตาม barcode (A→Z)
      return a.barcode.compareTo(b.barcode);
    });
    return sortedUnits;
  }

  /// Helper function: แสดงตัวเลขแบบ smart - ถ้าเป็นจำนวนเต็มไม่แสดงทศนิยม
  String _formatNumber(double value, {int decimals = 2}) {
    final floorValue = value.floor().toDouble();
    if ((value - floorValue).abs() < 0.0001) {
      return floorValue.toInt().toString();
    } else {
      return value.toStringAsFixed(decimals).replaceAll(RegExp(r'0+$'), '').replaceAll(RegExp(r'\.$'), '');
    }
  }

  // Generate color based on itemCode
  Color _getItemCodeColor(String itemCode) {
    if (itemCode.isEmpty) return Colors.blue;

    int hash = 0;
    for (int i = 0; i < itemCode.length; i++) {
      hash = itemCode.codeUnitAt(i) + ((hash << 5) - hash);
    }

    // ชุดสีที่เด่นชัดและแตกต่างกันมาก (24 สี)
    final colorPalette = [
      Colors.blue[400]!, // น้ำเงิน
      Colors.green[400]!, // เขียว
      Colors.orange[400]!, // ส้ม
      Colors.purple[400]!, // ม่วง
      Colors.pink[400]!, // ชมพู
      Colors.teal[400]!, // เขียวอมฟ้า
      Colors.indigo[400]!, // คราม
      Colors.cyan[400]!, // ฟ้าอมเขียว
      Colors.red[400]!, // แดง
      Colors.amber[600]!, // เหลืองทอง
      Colors.deepOrange[400]!, // ส้มเข้ม
      Colors.lightGreen[500]!, // เขียวอ่อน
      Colors.deepPurple[400]!, // ม่วงเข้ม
      Colors.lightBlue[400]!, // ฟ้าอ่อน
      Colors.lime[700]!, // เหลืองมะนาว
      Colors.brown[400]!, // น้ำตาล
      Colors.blueGrey[400]!, // เทาน้ำเงิน
      Colors.pinkAccent[200]!, // ชมพูสด
      Colors.tealAccent[400]!, // เขียวมิ้นท์
      Colors.purpleAccent[200]!, // ม่วงสด
      Colors.orangeAccent[400]!, // ส้มสด
      Colors.greenAccent[400]!, // เขียวสด
      Colors.blueAccent[400]!, // น้ำเงินสด
      Colors.redAccent[400]!, // แดงสด
    ];

    final colorIndex = hash.abs() % colorPalette.length;
    return colorPalette[colorIndex];
  }

  @override
  Widget build(BuildContext context) {
    final itemColor = _getItemCodeColor(widget.cardData.product.itemCode);
    final itemColorDark = HSLColor.fromColor(itemColor).withLightness(0.4).toColor(); // ปรับเป็น 0.4 เพื่อให้เข้มขึ้น

    return Card(
      elevation: 2,
      shape: RoundedRectangleBorder(
        borderRadius: BorderRadius.circular(8),
        side: BorderSide(color: itemColorDark, width: 3), // เพิ่มความหนาเป็น 3
      ),
      child: InkWell(
        onTap: widget.onTap,
        borderRadius: BorderRadius.circular(8),
        child: Padding(
          padding: const EdgeInsets.all(8),
          child: Row(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              // รูปภาพสินค้า (ซ้ายสุด)
              if (widget.cardData.product.imageuri != null && widget.cardData.product.imageuri!.isNotEmpty)
                Container(
                  width: 100,
                  height: 100,
                  decoration: BoxDecoration(color: Colors.grey[100], borderRadius: BorderRadius.circular(8)),
                  child: Stack(
                    children: [
                      ClipRRect(
                        borderRadius: BorderRadius.circular(8),
                        child: Image.network(
                          widget.cardData.product.imageuri!,
                          width: 100,
                          height: 100,
                          fit: BoxFit.cover,
                          loadingBuilder: (context, child, loadingProgress) {
                            if (loadingProgress == null) return child;
                            return Container(
                              decoration: BoxDecoration(gradient: LinearGradient(colors: [itemColor.withOpacity(0.3), itemColor.withOpacity(0.6)])),
                              child: Center(
                                child: CircularProgressIndicator(
                                  value: loadingProgress.expectedTotalBytes != null ? loadingProgress.cumulativeBytesLoaded / loadingProgress.expectedTotalBytes! : null,
                                  strokeWidth: 2,
                                  color: itemColorDark,
                                ),
                              ),
                            );
                          },
                          errorBuilder: (context, error, stackTrace) {
                            return const SizedBox.shrink();
                          },
                        ),
                      ),
                      // Item Code Badge
                      if (widget.cardData.product.itemCode.isNotEmpty)
                        Positioned(
                          top: 4,
                          right: 4,
                          child: Container(
                            padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
                            decoration: BoxDecoration(color: itemColorDark, borderRadius: BorderRadius.circular(3)),
                            child: Text(
                              widget.cardData.product.itemCode,
                              style: const TextStyle(color: Colors.white, fontSize: 8, fontWeight: FontWeight.bold),
                            ),
                          ),
                        ),
                    ],
                  ),
                ),

              const SizedBox(width: 12),

              // ข้อมูลสินค้า (ขวา)
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    // ชื่อสินค้า
                    Text(
                      widget.cardData.product.name0,
                      style: const TextStyle(fontSize: 14, fontWeight: FontWeight.bold, height: 1.2),
                      maxLines: 2,
                      overflow: TextOverflow.ellipsis,
                    ),
                    const SizedBox(height: 6),

                    // รหัสสินค้าและบาร์โค้ด
                    if (widget.cardData.product.itemCode.isNotEmpty)
                      Row(
                        children: [
                          Container(
                            width: 3,
                            height: 12,
                            decoration: BoxDecoration(color: itemColorDark, borderRadius: BorderRadius.circular(1.5)),
                          ),
                          SizedBox(width: 4),
                          Expanded(
                            child: Text(
                              '${global.language("code")}: ${widget.cardData.product.itemCode}',
                              style: TextStyle(fontSize: 10, color: itemColorDark, fontWeight: FontWeight.w600),
                            ),
                          ),
                        ],
                      ),

                    const SizedBox(height: 6),

                    // ยอดคงเหลือ
                    if (widget.cardData.hasBalance)
                      Container(
                        padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 4),
                        decoration: BoxDecoration(
                          color: Colors.green[50],
                          borderRadius: BorderRadius.circular(4),
                          border: Border.all(color: Colors.green[300]!, width: 1),
                        ),
                        child: Row(
                          children: [
                            Icon(Icons.inventory_2, size: 12, color: Colors.green[700]),
                            SizedBox(width: 4),
                            Expanded(
                              child: Builder(
                                builder: (context) {
                                  final baseUnit = widget.cardData.allUnits.firstWhere((u) => u.unitStand == 1 && u.unitDive == 1, orElse: () => widget.cardData.product);

                                  final baseUnitName = baseUnit.unitName.isNotEmpty ? baseUnit.unitName : baseUnit.unitCode;

                                  final balanceModel = ProductBalanceModel(
                                    itemCode: widget.cardData.itemCode,
                                    totalBalance: widget.cardData.totalBalance,
                                    balanceByShop: widget.cardData.balanceByShop,
                                  );

                                  final multiLevelPacking = balanceModel.formatWithMultiLevelPacking(allUnits: widget.cardData.allUnits);

                                  return Column(
                                    crossAxisAlignment: CrossAxisAlignment.start,
                                    children: [
                                      Text(
                                        '${global.language("balance_remaining")}: ${_formatNumber(widget.cardData.totalBalance)} $baseUnitName',
                                        style: TextStyle(fontSize: 10, color: Colors.green[900], fontWeight: FontWeight.bold),
                                      ),
                                      if (widget.cardData.allUnits.length > 1) ...[
                                        const SizedBox(height: 2),
                                        Text(
                                          multiLevelPacking,
                                          style: TextStyle(fontSize: 9, color: Colors.green[700], fontStyle: FontStyle.italic),
                                        ),
                                      ],
                                    ],
                                  );
                                },
                              ),
                            ),
                          ],
                        ),
                      ),

                    // ยอดค้างรับ / ค้างส่ง
                    if (widget.cardData.hasPending)
                      Padding(
                        padding: const EdgeInsets.only(top: 6),
                        child: Container(
                          padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 4),
                          decoration: BoxDecoration(
                            color: Colors.orange[50],
                            borderRadius: BorderRadius.circular(4),
                            border: Border.all(color: Colors.orange[300]!, width: 1),
                          ),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              if (widget.cardData.hasPendingRecv)
                                Row(
                                  children: [
                                    Icon(Icons.call_received, size: 12, color: Colors.blue[700]),
                                    const SizedBox(width: 4),
                                    Expanded(
                                      child: Text(
                                        '${global.language("pending_receive")}: ${widget.cardData.pendingRecvWord.isNotEmpty ? widget.cardData.pendingRecvWord : _formatNumber(widget.cardData.pendingRecvQty)}',
                                        style: TextStyle(fontSize: 10, color: Colors.blue[900], fontWeight: FontWeight.w600),
                                      ),
                                    ),
                                  ],
                                ),
                              if (widget.cardData.hasPendingRecv && widget.cardData.hasPendingSend)
                                const SizedBox(height: 2),
                              if (widget.cardData.hasPendingSend)
                                Row(
                                  children: [
                                    Icon(Icons.call_made, size: 12, color: Colors.red[700]),
                                    const SizedBox(width: 4),
                                    Expanded(
                                      child: Text(
                                        '${global.language("pending_delivery")}: ${widget.cardData.pendingSendWord.isNotEmpty ? widget.cardData.pendingSendWord : _formatNumber(widget.cardData.pendingSendQty)}',
                                        style: TextStyle(fontSize: 10, color: Colors.red[900], fontWeight: FontWeight.w600),
                                      ),
                                    ),
                                  ],
                                ),
                            ],
                          ),
                        ),
                      ),

                    const SizedBox(height: 6),

                    // หน่วยนับทั้งหมด
                    if (widget.cardData.allUnits.isNotEmpty)
                      Container(
                        padding: const EdgeInsets.all(6),
                        decoration: BoxDecoration(
                          color: Colors.blue[50],
                          borderRadius: BorderRadius.circular(4),
                          border: Border.all(color: Colors.blue[200]!, width: 0.5),
                        ),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: _sortUnitsByRatioAndBarcode(widget.cardData.allUnits).map((unit) {
                            final isCurrentUnit = unit.barcode == widget.cardData.product.barcode;
                            return _BarcodeRowWithHover(
                              unit: unit,
                              isCurrentUnit: isCurrentUnit,
                              onTap: () {
                                // เรียก callback เพื่อเพิ่มในตระกร้า
                                if (widget.onProductSelected != null) {
                                  widget.onProductSelected!(unit);
                                }
                              },
                            );
                          }).toList(),
                        ),
                      ),
                  ],
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }
}

/// Card entry animation — fade + slide up เมื่อ card ปรากฏครั้งแรก
class _AnimatedCardEntry extends StatefulWidget {
  final int index;
  final Widget child;

  const _AnimatedCardEntry({required this.index, required this.child});

  @override
  State<_AnimatedCardEntry> createState() => _AnimatedCardEntryState();
}

class _AnimatedCardEntryState extends State<_AnimatedCardEntry>
    with SingleTickerProviderStateMixin {
  late final AnimationController _controller;
  late final Animation<double> _fadeAnimation;
  late final Animation<Offset> _slideAnimation;

  @override
  void initState() {
    super.initState();
    // stagger delay: แต่ละ card หน่วงเล็กน้อย (max 300ms สำหรับ card ที่ 10+)
    final delay = math.min(widget.index * 30, 300);
    _controller = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 250),
    );
    _fadeAnimation = CurvedAnimation(parent: _controller, curve: Curves.easeOut);
    _slideAnimation = Tween<Offset>(
      begin: const Offset(0, 0.08),
      end: Offset.zero,
    ).animate(CurvedAnimation(parent: _controller, curve: Curves.easeOut));

    Future.delayed(Duration(milliseconds: delay), () {
      if (mounted) _controller.forward();
    });
  }

  @override
  void dispose() {
    _controller.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return FadeTransition(
      opacity: _fadeAnimation,
      child: SlideTransition(
        position: _slideAnimation,
        child: widget.child,
      ),
    );
  }
}

/// Shimmer skeleton grid — แสดงระหว่างรอผลค้นหาครั้งแรก
class _ShimmerSkeletonGrid extends StatefulWidget {
  const _ShimmerSkeletonGrid();

  @override
  State<_ShimmerSkeletonGrid> createState() => _ShimmerSkeletonGridState();
}

class _ShimmerSkeletonGridState extends State<_ShimmerSkeletonGrid>
    with SingleTickerProviderStateMixin {
  late final AnimationController _shimmerController;

  @override
  void initState() {
    super.initState();
    _shimmerController = AnimationController(
      vsync: this,
      duration: const Duration(milliseconds: 1200),
    )..repeat();
  }

  @override
  void dispose() {
    _shimmerController.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return AnimatedBuilder(
      animation: _shimmerController,
      builder: (context, child) {
        return LayoutBuilder(
          builder: (context, constraints) {
            const cardWidth = 280.0;
            const spacing = 4.0;
            int columns = (constraints.maxWidth / (cardWidth + spacing)).floor();
            if (columns < 1) columns = 1;
            final actualWidth = (constraints.maxWidth - (spacing * (columns - 1))) / columns;

            return Padding(
              padding: const EdgeInsets.all(8),
              child: Wrap(
                spacing: spacing,
                runSpacing: spacing,
                children: List.generate(columns * 3, (index) {
                  return SizedBox(
                    width: actualWidth,
                    child: _buildSkeletonCard(_shimmerController.value),
                  );
                }),
              ),
            );
          },
        );
      },
    );
  }

  Widget _buildSkeletonCard(double shimmerValue) {
    final baseColor = Colors.grey[200]!;
    final highlightColor = Colors.grey[100]!;
    final opacity = 0.5 + 0.5 * math.sin(shimmerValue * math.pi * 2);
    final color = Color.lerp(baseColor, highlightColor, opacity)!;

    return Card(
      elevation: 1,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(8)),
      child: Padding(
        padding: const EdgeInsets.all(12),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            Container(
              height: 16, width: double.infinity,
              decoration: BoxDecoration(color: color, borderRadius: BorderRadius.circular(4)),
            ),
            const SizedBox(height: 8),
            Container(
              height: 14, width: 120,
              decoration: BoxDecoration(color: color, borderRadius: BorderRadius.circular(4)),
            ),
            const SizedBox(height: 12),
            Container(
              height: 12, width: 80,
              decoration: BoxDecoration(color: color, borderRadius: BorderRadius.circular(4)),
            ),
            const SizedBox(height: 10),
            Container(
              height: 24, width: double.infinity,
              decoration: BoxDecoration(color: color, borderRadius: BorderRadius.circular(4)),
            ),
            const SizedBox(height: 6),
            Container(
              height: 24, width: 180,
              decoration: BoxDecoration(color: color, borderRadius: BorderRadius.circular(4)),
            ),
          ],
        ),
      ),
    );
  }
}
