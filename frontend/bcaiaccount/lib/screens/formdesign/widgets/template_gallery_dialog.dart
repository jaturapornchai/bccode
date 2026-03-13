import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import '../models/form_style.dart';
import '../models/form_template.dart';
import '../utils/form_template_factory.dart';

class _TemplateInfo {
  final String id;
  final String name;
  final String nameEn;
  final String description;
  final String category;
  final IconData icon;
  final String documentType;

  const _TemplateInfo({
    required this.id,
    required this.name,
    required this.nameEn,
    required this.description,
    required this.category,
    required this.icon,
    required this.documentType,
  });
}

const List<_TemplateInfo> _templates = [
  _TemplateInfo(
    id: 'tax_invoice',
    name: 'ใบกำกับภาษี',
    nameEn: 'Tax Invoice',
    description: 'ใบกำกับภาษีเต็มรูปแบบ สำหรับการขายสินค้าและบริการ',
    category: 'sales',
    icon: Icons.receipt_long,
    documentType: 'tax_invoice',
  ),
  _TemplateInfo(
    id: 'abbreviated_tax_invoice',
    name: 'ใบกำกับภาษีอย่างย่อ',
    nameEn: 'Abbreviated Tax Invoice',
    description: 'ใบกำกับภาษีอย่างย่อ สำหรับการขายปลีก',
    category: 'sales',
    icon: Icons.receipt,
    documentType: 'abbreviated_tax_invoice',
  ),
  _TemplateInfo(
    id: 'receipt',
    name: 'ใบเสร็จรับเงิน',
    nameEn: 'Receipt',
    description: 'ใบเสร็จรับเงิน สำหรับยืนยันการรับชำระเงิน',
    category: 'sales',
    icon: Icons.payments,
    documentType: 'receipt',
  ),
  _TemplateInfo(
    id: 'cash_bill',
    name: 'บิลเงินสด',
    nameEn: 'Cash Bill',
    description: 'บิลเงินสด สำหรับการขายเงินสดทั่วไป',
    category: 'sales',
    icon: Icons.point_of_sale,
    documentType: 'cash_bill',
  ),
  _TemplateInfo(
    id: 'quotation',
    name: 'ใบเสนอราคา',
    nameEn: 'Quotation',
    description: 'ใบเสนอราคา สำหรับเสนอราคาสินค้าและบริการ',
    category: 'sales',
    icon: Icons.request_quote,
    documentType: 'quotation',
  ),
  _TemplateInfo(
    id: 'purchase_order',
    name: 'ใบสั่งซื้อ',
    nameEn: 'Purchase Order',
    description: 'ใบสั่งซื้อ สำหรับสั่งซื้อสินค้าจากผู้จำหน่าย',
    category: 'purchase',
    icon: Icons.shopping_cart,
    documentType: 'purchase_order',
  ),
  _TemplateInfo(
    id: 'delivery_note',
    name: 'ใบส่งของ',
    nameEn: 'Delivery Note',
    description: 'ใบส่งของ สำหรับการจัดส่งสินค้า',
    category: 'inventory',
    icon: Icons.local_shipping,
    documentType: 'delivery_note',
  ),
  _TemplateInfo(
    id: 'invoice',
    name: 'ใบแจ้งหนี้',
    nameEn: 'Invoice',
    description: 'ใบแจ้งหนี้ สำหรับแจ้งยอดค้างชำระ',
    category: 'financial',
    icon: Icons.description,
    documentType: 'invoice',
  ),
  _TemplateInfo(
    id: 'credit_note',
    name: 'ใบลดหนี้',
    nameEn: 'Credit Note',
    description: 'ใบลดหนี้ สำหรับปรับลดยอดหนี้',
    category: 'financial',
    icon: Icons.remove_circle_outline,
    documentType: 'credit_note',
  ),
  _TemplateInfo(
    id: 'debit_note',
    name: 'ใบเพิ่มหนี้',
    nameEn: 'Debit Note',
    description: 'ใบเพิ่มหนี้ สำหรับปรับเพิ่มยอดหนี้',
    category: 'financial',
    icon: Icons.add_circle_outline,
    documentType: 'debit_note',
  ),
  _TemplateInfo(
    id: 'booking_form',
    name: 'ใบรับจอง',
    nameEn: 'Booking Form',
    description: 'ใบรับจอง สำหรับรับจองสินค้าหรือบริการ',
    category: 'sales',
    icon: Icons.bookmark,
    documentType: 'booking_form',
  ),
  _TemplateInfo(
    id: 'payment_voucher',
    name: 'ใบวางบิล',
    nameEn: 'Payment Voucher',
    description: 'ใบวางบิล สำหรับวางบิลเรียกเก็บเงิน',
    category: 'financial',
    icon: Icons.account_balance_wallet,
    documentType: 'payment_voucher',
  ),
];

class TemplateGalleryDialog extends StatefulWidget {
  final Function(FormTemplate template) onTemplateSelected;

  const TemplateGalleryDialog({super.key, required this.onTemplateSelected});

  @override
  State<TemplateGalleryDialog> createState() => _TemplateGalleryDialogState();
}

class _TemplateGalleryDialogState extends State<TemplateGalleryDialog>
    with global.ThemeRefreshMixin {
  String _selectedCategory = 'all';
  String _searchQuery = '';
  String? _hoveredTemplateId;
  String? _selectedStyleId;
  String? _hoveredStyleId;
  final TextEditingController _searchController = TextEditingController();

  static const Map<String, String> _categoryLabels = {
    'all': 'ทั้งหมด',
    'sales': 'ขาย',
    'purchase': 'จัดซื้อ',
    'inventory': 'คลังสินค้า',
    'financial': 'การเงิน',
  };

  static const Map<String, IconData> _categoryIcons = {
    'all': Icons.grid_view_rounded,
    'sales': Icons.storefront,
    'purchase': Icons.shopping_bag_outlined,
    'inventory': Icons.inventory_2_outlined,
    'financial': Icons.account_balance_outlined,
  };

  List<_TemplateInfo> get _filteredTemplates {
    return _templates.where((t) {
      final matchesCategory =
          _selectedCategory == 'all' || t.category == _selectedCategory;
      final query = _searchQuery.toLowerCase();
      final matchesSearch = query.isEmpty ||
          t.name.toLowerCase().contains(query) ||
          t.nameEn.toLowerCase().contains(query) ||
          t.description.toLowerCase().contains(query);
      return matchesCategory && matchesSearch;
    }).toList();
  }

  @override
  void dispose() {
    _searchController.dispose();
    super.dispose();
  }

  Future<FormTemplate> _createTemplateFromInfo(_TemplateInfo info) {
    return FormTemplateFactory.createTemplate(info.documentType,
        styleId: _selectedStyleId);
  }

  @override
  Widget build(BuildContext context) {
    return Dialog(
      backgroundColor: global.theme.dialogColor,
      shape: RoundedRectangleBorder(borderRadius: BorderRadius.circular(16)),
      child: SizedBox(
        width: 800,
        height: 600,
        child: Column(
          children: [
            _buildTitleBar(),
            _buildStylePicker(),
            _buildCategoryTabs(),
            _buildSearchField(),
            Expanded(child: _buildTemplateGrid()),
          ],
        ),
      ),
    );
  }

  Widget _buildTitleBar() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 20, vertical: 14),
      decoration: BoxDecoration(
        color: global.theme.surfaceColor,
        borderRadius: const BorderRadius.vertical(top: Radius.circular(16)),
        border: Border(
          bottom: BorderSide(color: global.theme.dividerBorderColor, width: 1),
        ),
      ),
      child: Row(
        children: [
          Icon(Icons.dashboard_customize_outlined,
              color: global.theme.primaryColor, size: 24),
          const SizedBox(width: 10),
          Text(
            'Template Gallery',
            style: TextStyle(
              fontSize: 18,
              fontWeight: FontWeight.w600,
              color: global.theme.textColor,
            ),
          ),
          const Spacer(),
          Material(
            color: Colors.transparent,
            shape: const CircleBorder(),
            clipBehavior: Clip.antiAlias,
            child: IconButton(
              onPressed: () => Navigator.of(context).pop(),
              icon: Icon(Icons.close, color: global.theme.iconColor, size: 22),
              splashRadius: 20,
              tooltip: 'ปิด',
            ),
          ),
        ],
      ),
    );
  }

  Widget _buildStylePicker() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
      decoration: BoxDecoration(
        border: Border(
          bottom:
              BorderSide(color: global.theme.dividerBorderColor, width: 0.5),
        ),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Row(
            children: [
              Icon(Icons.palette_outlined,
                  size: 16, color: global.theme.iconSecondaryColor),
              const SizedBox(width: 6),
              Text(
                'สไตล์เอกสาร',
                style: TextStyle(
                  fontSize: 12,
                  fontWeight: FontWeight.w600,
                  color: global.theme.textSecondaryColor,
                ),
              ),
              const Spacer(),
              if (_selectedStyleId != null)
                GestureDetector(
                  onTap: () => setState(() => _selectedStyleId = null),
                  child: Text(
                    'ล้าง',
                    style: TextStyle(
                      fontSize: 11,
                      color: global.theme.primaryColor,
                    ),
                  ),
                ),
            ],
          ),
          const SizedBox(height: 8),
          SizedBox(
            height: 160,
            child: Row(
              children: formStyles.map((style) {
                final isSelected = _selectedStyleId == style.id;
                final isHovered = _hoveredStyleId == style.id;
                return Expanded(
                  child: Padding(
                    padding: const EdgeInsets.symmetric(horizontal: 4),
                    child: MouseRegion(
                      onEnter: (_) =>
                          setState(() => _hoveredStyleId = style.id),
                      onExit: (_) => setState(() => _hoveredStyleId = null),
                      cursor: SystemMouseCursors.click,
                      child: GestureDetector(
                        onTap: () => setState(() => _selectedStyleId =
                            isSelected ? null : style.id),
                        child: AnimatedContainer(
                          duration: const Duration(milliseconds: 150),
                          padding: const EdgeInsets.all(6),
                          decoration: BoxDecoration(
                            color: isSelected
                                ? style.titleColor.withAlpha(12)
                                : isHovered
                                    ? global.theme.rowHoverColor
                                    : global.theme.cardColor,
                            borderRadius: BorderRadius.circular(10),
                            border: Border.all(
                              color: isSelected
                                  ? style.titleColor
                                  : isHovered
                                      ? global.theme.primaryColor
                                          .withAlpha(80)
                                      : global.theme.dividerBorderColor,
                              width: isSelected ? 2 : 1,
                            ),
                            boxShadow: isSelected
                                ? [
                                    BoxShadow(
                                      color: style.titleColor.withAlpha(30),
                                      blurRadius: 8,
                                      offset: const Offset(0, 2),
                                    ),
                                  ]
                                : [],
                          ),
                          child: Column(
                            children: [
                              // Mini document preview
                              Expanded(
                                child: _buildMiniDocPreview(style),
                              ),
                              const SizedBox(height: 4),
                              // Style name
                              Text(
                                style.name,
                                style: TextStyle(
                                  fontSize: 11,
                                  fontWeight: isSelected
                                      ? FontWeight.w700
                                      : FontWeight.w500,
                                  color: isSelected
                                      ? style.titleColor
                                      : global.theme.textColor,
                                ),
                              ),
                              Text(
                                style.nameEn,
                                style: TextStyle(
                                  fontSize: 9,
                                  color: global.theme.textSecondaryColor,
                                ),
                              ),
                            ],
                          ),
                        ),
                      ),
                    ),
                  ),
                );
              }).toList(),
            ),
          ),
        ],
      ),
    );
  }

  /// Mini document preview — จำลองเอกสารย่อส่วน
  Widget _buildMiniDocPreview(FormStyle style) {
    final bw = style.borderWidth;
    final sepStyle = style.separatorStyle;

    return Container(
      decoration: BoxDecoration(
        color: Colors.white,
        borderRadius: BorderRadius.circular(4),
        border: Border.all(color: style.borderColor, width: 0.5),
      ),
      clipBehavior: Clip.antiAlias,
      padding: const EdgeInsets.all(5),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.stretch,
        children: [
          // ── Header: company + title ──
          Row(
            children: [
              // Logo placeholder
              Container(
                width: 14,
                height: 14,
                decoration: BoxDecoration(
                  color: style.borderColor.withAlpha(40),
                  borderRadius: BorderRadius.circular(2),
                ),
                child: Icon(Icons.business, size: 8, color: style.labelColor),
              ),
              const SizedBox(width: 4),
              Expanded(
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Container(
                      height: 3,
                      width: 36,
                      decoration: BoxDecoration(
                        color: style.textColor,
                        borderRadius: BorderRadius.circular(1),
                      ),
                    ),
                    const SizedBox(height: 2),
                    Container(
                      height: 2,
                      width: 28,
                      decoration: BoxDecoration(
                        color: style.labelColor.withAlpha(120),
                        borderRadius: BorderRadius.circular(1),
                      ),
                    ),
                  ],
                ),
              ),
              // Title
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
                decoration: BoxDecoration(
                  border: Border.all(
                      color: style.titleColor.withAlpha(60), width: 0.5),
                  borderRadius: BorderRadius.circular(2),
                ),
                child: Text(
                  'TAX',
                  style: TextStyle(
                    fontSize: 6,
                    fontWeight: FontWeight.w800,
                    color: style.titleColor,
                    height: 1,
                  ),
                ),
              ),
            ],
          ),

          const SizedBox(height: 3),

          // ── Separator ──
          _buildMiniSeparator(style.borderColor, bw, sepStyle),

          const SizedBox(height: 3),

          // ── Info rows ──
          for (var i = 0; i < 2; i++) ...[
            Row(
              children: [
                Container(
                  height: 2,
                  width: 16,
                  color: style.labelColor.withAlpha(150),
                ),
                const SizedBox(width: 3),
                Expanded(
                  child: Container(
                    height: 2,
                    decoration: BoxDecoration(
                      color: style.textColor.withAlpha(80),
                      borderRadius: BorderRadius.circular(0.5),
                    ),
                  ),
                ),
                const SizedBox(width: 6),
                Container(
                  height: 2,
                  width: 12,
                  color: style.labelColor.withAlpha(150),
                ),
                const SizedBox(width: 3),
                Container(
                  height: 2,
                  width: 18,
                  decoration: BoxDecoration(
                    color: style.textColor.withAlpha(80),
                    borderRadius: BorderRadius.circular(0.5),
                  ),
                ),
              ],
            ),
            const SizedBox(height: 2),
          ],

          const SizedBox(height: 2),

          // ── Table ──
          Expanded(
            child: Container(
              decoration: BoxDecoration(
                border: Border.all(
                    color: style.borderColor, width: bw.clamp(0.3, 1.0)),
              ),
              child: Column(
                children: [
                  // Table header
                  Container(
                    height: 7,
                    decoration: BoxDecoration(
                      color: style.borderColor.withAlpha(30),
                      border: Border(
                        bottom: BorderSide(
                            color: style.borderColor,
                            width: bw.clamp(0.3, 1.0)),
                      ),
                    ),
                    child: Row(
                      children: List.generate(
                        5,
                        (i) => Expanded(
                          flex: i == 2 ? 3 : 1,
                          child: Container(
                            decoration: i < 4
                                ? BoxDecoration(
                                    border: Border(
                                      right: BorderSide(
                                          color: style.borderColor,
                                          width: bw.clamp(0.2, 0.5)),
                                    ),
                                  )
                                : null,
                            alignment: Alignment.center,
                            child: Container(
                              height: 1.5,
                              width: i == 2 ? 14 : 6,
                              color: style.textColor.withAlpha(130),
                            ),
                          ),
                        ),
                      ),
                    ),
                  ),
                  // Table rows
                  for (var r = 0; r < 3; r++)
                    Expanded(
                      child: Container(
                        decoration: BoxDecoration(
                          border: r < 2
                              ? Border(
                                  bottom: BorderSide(
                                      color: style.borderColor.withAlpha(80),
                                      width: 0.3),
                                )
                              : null,
                        ),
                        child: Row(
                          children: List.generate(
                            5,
                            (i) => Expanded(
                              flex: i == 2 ? 3 : 1,
                              child: Container(
                                decoration: i < 4
                                    ? BoxDecoration(
                                        border: Border(
                                          right: BorderSide(
                                              color: style.borderColor
                                                  .withAlpha(60),
                                              width: 0.3),
                                        ),
                                      )
                                    : null,
                                alignment: Alignment.center,
                                child: Container(
                                  height: 1.2,
                                  width: i == 2 ? 12 : 5,
                                  color: style.textColor.withAlpha(60),
                                ),
                              ),
                            ),
                          ),
                        ),
                      ),
                    ),
                ],
              ),
            ),
          ),

          const SizedBox(height: 3),

          // ── Footer: total + signatures ──
          Row(
            mainAxisAlignment: MainAxisAlignment.end,
            children: [
              Container(
                height: 2,
                width: 16,
                color: style.labelColor.withAlpha(120),
              ),
              const SizedBox(width: 4),
              Container(
                padding:
                    const EdgeInsets.symmetric(horizontal: 3, vertical: 1),
                decoration: BoxDecoration(
                  border: Border.all(
                      color: style.borderColor, width: bw.clamp(0.3, 0.8)),
                  borderRadius: BorderRadius.circular(1),
                ),
                child: Text(
                  '14,850.00',
                  style: TextStyle(
                    fontSize: 5,
                    fontWeight: FontWeight.w700,
                    color: style.titleColor,
                    height: 1,
                  ),
                ),
              ),
            ],
          ),

          const SizedBox(height: 3),

          _buildMiniSeparator(style.borderColor, bw, sepStyle),

          const SizedBox(height: 2),

          // ── Signatures ──
          Row(
            mainAxisAlignment: MainAxisAlignment.spaceAround,
            children: [
              for (var i = 0; i < 3; i++)
                Column(
                  children: [
                    Container(
                      width: 24,
                      height: 0.5,
                      color: style.borderColor,
                    ),
                    const SizedBox(height: 1),
                    Container(
                      height: 1.5,
                      width: 14,
                      color: style.labelColor.withAlpha(100),
                    ),
                  ],
                ),
            ],
          ),
        ],
      ),
    );
  }

  /// Mini separator line with style
  Widget _buildMiniSeparator(Color color, double width, String style) {
    final w = width.clamp(0.3, 1.0);
    if (style == 'dotted') {
      return Row(
        children: List.generate(
          30,
          (i) => Expanded(
            child: Container(
              height: w,
              margin: const EdgeInsets.symmetric(horizontal: 0.5),
              color: i.isEven ? color : Colors.transparent,
            ),
          ),
        ),
      );
    }
    if (style == 'dashed') {
      return Row(
        children: List.generate(
          15,
          (i) => Expanded(
            child: Container(
              height: w,
              margin: const EdgeInsets.symmetric(horizontal: 0.8),
              color: i.isEven ? color : Colors.transparent,
            ),
          ),
        ),
      );
    }
    // solid
    return Container(height: w, color: color);
  }

  Widget _buildCategoryTabs() {
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
      child: SingleChildScrollView(
        scrollDirection: Axis.horizontal,
        child: Row(
          children: _categoryLabels.entries.map((entry) {
            final isSelected = _selectedCategory == entry.key;
            return Padding(
              padding: const EdgeInsets.only(right: 8),
              child: FilterChip(
                selected: isSelected,
                label: Row(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(
                      _categoryIcons[entry.key],
                      size: 16,
                      color: isSelected
                          ? global.theme.primaryColor
                          : global.theme.iconSecondaryColor,
                    ),
                    const SizedBox(width: 6),
                    Text(entry.value),
                  ],
                ),
                labelStyle: TextStyle(
                  fontSize: 13,
                  fontWeight: isSelected ? FontWeight.w600 : FontWeight.normal,
                  color: isSelected
                      ? global.theme.primaryColor
                      : global.theme.textSecondaryColor,
                ),
                backgroundColor: global.theme.cardColor,
                selectedColor: global.theme.primaryColor.withAlpha(30),
                side: BorderSide(
                  color: isSelected
                      ? global.theme.primaryColor
                      : global.theme.dividerBorderColor,
                ),
                shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(20)),
                showCheckmark: false,
                onSelected: (_) {
                  setState(() => _selectedCategory = entry.key);
                },
              ),
            );
          }).toList(),
        ),
      ),
    );
  }

  Widget _buildSearchField() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 16, vertical: 4),
      child: TextField(
        controller: _searchController,
        onChanged: (value) => setState(() => _searchQuery = value),
        style: TextStyle(fontSize: 14, color: global.theme.formTextColor),
        decoration: InputDecoration(
          hintText: 'ค้นหาเทมเพลต...',
          hintStyle:
              TextStyle(fontSize: 14, color: global.theme.formHintColor),
          prefixIcon:
              Icon(Icons.search, color: global.theme.iconSecondaryColor),
          suffixIcon: _searchQuery.isNotEmpty
              ? IconButton(
                  icon: Icon(Icons.clear,
                      size: 18, color: global.theme.iconSecondaryColor),
                  onPressed: () {
                    _searchController.clear();
                    setState(() => _searchQuery = '');
                  },
                )
              : null,
          filled: true,
          fillColor: global.theme.inputFillColor,
          contentPadding:
              const EdgeInsets.symmetric(horizontal: 16, vertical: 10),
          border: OutlineInputBorder(
            borderRadius: BorderRadius.circular(10),
            borderSide: BorderSide(color: global.theme.formBorderColor),
          ),
          enabledBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(10),
            borderSide: BorderSide(color: global.theme.formBorderColor),
          ),
          focusedBorder: OutlineInputBorder(
            borderRadius: BorderRadius.circular(10),
            borderSide:
                BorderSide(color: global.theme.formFocusBorderColor, width: 2),
          ),
        ),
      ),
    );
  }

  Widget _buildTemplateGrid() {
    final filtered = _filteredTemplates;

    if (filtered.isEmpty) {
      return Center(
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(Icons.search_off_rounded,
                size: 48, color: global.theme.iconSecondaryColor),
            const SizedBox(height: 12),
            Text(
              'ไม่พบเทมเพลตที่ค้นหา',
              style: TextStyle(
                  fontSize: 15, color: global.theme.textSecondaryColor),
            ),
          ],
        ),
      );
    }

    return Padding(
      padding: const EdgeInsets.all(16),
      child: GridView.builder(
        gridDelegate: const SliverGridDelegateWithFixedCrossAxisCount(
          crossAxisCount: 3,
          mainAxisSpacing: 12,
          crossAxisSpacing: 12,
          childAspectRatio: 1.05,
        ),
        itemCount: filtered.length,
        itemBuilder: (context, index) => _buildTemplateCard(filtered[index]),
      ),
    );
  }

  Widget _buildTemplateCard(_TemplateInfo info) {
    final isHovered = _hoveredTemplateId == info.id;

    return MouseRegion(
      onEnter: (_) => setState(() => _hoveredTemplateId = info.id),
      onExit: (_) => setState(() => _hoveredTemplateId = null),
      cursor: SystemMouseCursors.click,
      child: GestureDetector(
        onTap: () async {
          final template = await _createTemplateFromInfo(info);
          if (!mounted) return;
          widget.onTemplateSelected(template);
          Navigator.of(context).pop();
        },
        child: AnimatedContainer(
          duration: const Duration(milliseconds: 180),
          curve: Curves.easeOut,
          decoration: BoxDecoration(
            color: isHovered
                ? global.theme.rowHoverColor
                : global.theme.cardColor,
            borderRadius: BorderRadius.circular(12),
            border: Border.all(
              color: isHovered
                  ? global.theme.primaryColor
                  : global.theme.dividerBorderColor,
              width: isHovered ? 1.5 : 1,
            ),
            boxShadow: isHovered
                ? [
                    BoxShadow(
                      color: global.theme.primaryColor.withAlpha(25),
                      blurRadius: 8,
                      offset: const Offset(0, 2),
                    ),
                  ]
                : [],
          ),
          padding: const EdgeInsets.all(14),
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            children: [
              Row(
                children: [
                  Container(
                    width: 40,
                    height: 40,
                    decoration: BoxDecoration(
                      color: global.theme.primaryColor.withAlpha(30),
                      borderRadius: BorderRadius.circular(10),
                    ),
                    child: Icon(info.icon,
                        size: 22, color: global.theme.primaryColor),
                  ),
                  const Spacer(),
                  _buildCategoryBadge(info.category),
                ],
              ),
              const SizedBox(height: 10),
              Text(
                info.name,
                style: TextStyle(
                  fontSize: 14,
                  fontWeight: FontWeight.w600,
                  color: global.theme.textColor,
                ),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
              const SizedBox(height: 2),
              Text(
                info.nameEn,
                style: TextStyle(
                  fontSize: 12,
                  color: global.theme.textSecondaryColor,
                ),
                maxLines: 1,
                overflow: TextOverflow.ellipsis,
              ),
              const SizedBox(height: 6),
              Expanded(
                child: Text(
                  info.description,
                  style: TextStyle(
                    fontSize: 11,
                    color: global.theme.textSecondaryColor,
                    height: 1.4,
                  ),
                  maxLines: 2,
                  overflow: TextOverflow.ellipsis,
                ),
              ),
            ],
          ),
        ),
      ),
    );
  }

  Widget _buildCategoryBadge(String category) {
    final label = _categoryLabels[category] ?? category;
    return Container(
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 3),
      decoration: BoxDecoration(
        color: global.theme.primaryColor.withAlpha(30),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Text(
        label,
        style: TextStyle(
          fontSize: 10,
          fontWeight: FontWeight.w500,
          color: global.theme.primaryColor,
        ),
      ),
    );
  }
}
