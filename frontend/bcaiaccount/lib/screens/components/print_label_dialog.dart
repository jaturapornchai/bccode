import 'package:flutter/foundation.dart';
import 'package:flutter/material.dart';
import 'package:pdf/pdf.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/screens/components/selected_product_grid_item.dart';
import 'package:smlaicloud/components/product_label_print_a4_shelf.dart';
import 'package:smlaicloud/components/product_label_print_a4_shelf_medium.dart';
import 'package:smlaicloud/components/product_label_print_a4_shelf_large.dart';
import 'package:smlaicloud/components/product_label_print_a4_shelf_xlarge.dart';
import 'package:smlaicloud/model/product_model.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';

enum LabelSize {
  regular, // ปกติ A4
  small, // 6.5x3.5 ซม.
  medium, // 10x3.5 ซม.
  large, // 15x10 ซม.
  xlarge, // 21x30 ซม. แนวนอน
}

extension LabelSizeExtension on LabelSize {
  String get displayName {
    switch (this) {
      case LabelSize.regular:
        return global.language('print_on_a4');
      case LabelSize.small:
        return global.language('print_shelf_label_a4');
      case LabelSize.medium:
        return global.language('print_shelf_label_a4');
      case LabelSize.large:
        return global.language('print_shelf_label_a4');
      case LabelSize.xlarge:
        return global.language('print_shelf_label_a4');
    }
  }

  String get description {
    switch (this) {
      case LabelSize.regular:
        return global.language('label_format_regular_desc');
      case LabelSize.small:
        return 'A4 แนวตั้ง 3 ฉลาก/แถว (6.5x3.5 ซม.)';
      case LabelSize.medium:
        return 'A4 แนวตั้ง 2 ฉลาก/แถว (10x3.5 ซม.)';
      case LabelSize.large:
        return 'A4 แนวตั้ง 1 ฉลาก/แถว (15x10 ซม.)';
      case LabelSize.xlarge:
        return 'A4 แนวตั้ง 1 ฉลาก/แถว (21x30 ซม.) แนวนอน';
    }
  }

  IconData get icon {
    switch (this) {
      case LabelSize.regular:
        return Icons.print;
      case LabelSize.small:
        return Icons.grid_view;
      case LabelSize.medium:
        return Icons.view_column;
      case LabelSize.large:
        return Icons.view_agenda;
      case LabelSize.xlarge:
        return Icons.crop_landscape;
    }
  }

  MaterialColor get color {
    switch (this) {
      case LabelSize.regular:
        return Colors.blue;
      case LabelSize.small:
        return Colors.orange;
      case LabelSize.medium:
        return Colors.teal;
      case LabelSize.large:
        return Colors.green;
      case LabelSize.xlarge:
        return Colors.purple;
    }
  }

  String get source {
    switch (this) {
      case LabelSize.regular:
        return '';
      case LabelSize.small:
        return 'a4_shelf_label';
      case LabelSize.medium:
        return 'a4_shelf_label_medium';
      case LabelSize.large:
        return 'a4_shelf_label_large';
      case LabelSize.xlarge:
        return 'a4_shelf_label_xlarge';
    }
  }

  bool get needsColorSelection {
    return this != LabelSize.regular;
  }
}

class PrintLabelDialog {
  static void show({
    required BuildContext context,
    required List<ProductWithCopies> products,
    VoidCallback? onRegularLabelSelected,
    Function(PdfColor priceTextColor, bool showBorder)? onA4ShelfSmallSelected,
    Function(PdfColor priceTextColor, bool showBorder)? onA4ShelfMediumSelected,
    Function(PdfColor priceTextColor, bool showBorder)? onA4ShelfLargeSelected,
    Function(PdfColor priceTextColor, bool showBorder)? onA4ShelfXLargeSelected,
  }) {
    if (kDebugMode) {
      AppLogger.debug(
        'PrintLabelDialog.show called with ${products.length} products',
      ); // Debug log
    }
    if (products.isEmpty) {
      if (kDebugMode) {
        AppLogger.debug('PrintLabelDialog: No products provided'); // Debug log
      }
      ScaffoldMessenger.of(
        context,
      ).showSnackBar(SnackBar(content: Text(global.language('please_select_product_first'))));
      return;
    }

    if (kDebugMode) {
      AppLogger.debug('PrintLabelDialog: Showing dialog'); // Debug log
    }
    showDialog(
      context: context,
      builder: (dialogContext) => _PrintLabelSelectionDialog(
        products: products,
        onRegularLabelSelected: onRegularLabelSelected,
        onA4ShelfSmallSelected: onA4ShelfSmallSelected,
        onA4ShelfMediumSelected: onA4ShelfMediumSelected,
        onA4ShelfLargeSelected: onA4ShelfLargeSelected,
        onA4ShelfXLargeSelected: onA4ShelfXLargeSelected,
      ),
    );
  }
}

class _PrintLabelSelectionDialog extends StatelessWidget {
  final List<ProductWithCopies> products;
  final VoidCallback? onRegularLabelSelected;
  final Function(PdfColor priceTextColor, bool showBorder)?
  onA4ShelfSmallSelected;
  final Function(PdfColor priceTextColor, bool showBorder)?
  onA4ShelfMediumSelected;
  final Function(PdfColor priceTextColor, bool showBorder)?
  onA4ShelfLargeSelected;
  final Function(PdfColor priceTextColor, bool showBorder)?
  onA4ShelfXLargeSelected;

  const _PrintLabelSelectionDialog({
    required this.products,
    this.onRegularLabelSelected,
    this.onA4ShelfSmallSelected,
    this.onA4ShelfMediumSelected,
    this.onA4ShelfLargeSelected,
    this.onA4ShelfXLargeSelected,
  });

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(global.language('select_print_format')),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(global.language('select_label_print_format')),
          const SizedBox(height: 20),

          // สร้างรายการตัวเลือกการพิมพ์
          ...LabelSize.values.map(
            (labelSize) => _buildPrintOption(context, labelSize),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text(global.language('cancel')),
        ),
      ],
    );
  }

  Widget _buildPrintOption(BuildContext context, LabelSize labelSize) {
    return InkWell(
      onTap: () {
        if (kDebugMode) {
          AppLogger.debug('Print option selected: ${labelSize.displayName}'); // Debug log
        }
        Navigator.pop(context);
        if (labelSize.needsColorSelection) {
          if (kDebugMode) {
            AppLogger.debug('Needs color selection, showing color dialog'); // Debug log
          }
          _showColorSelectionDialog(
            context,
            labelSize,
            products,
            _getCallbackForLabelSize(labelSize),
          );
        } else {
          if (kDebugMode) {
            AppLogger.debug(
              'Regular label selected, calling callback function',
            ); // Debug log
          }
          // เรียก callback function สำหรับ regular label แทนการ call ProductBloc โดยตรง
          if (onRegularLabelSelected != null) {
            onRegularLabelSelected!();
          } else {
            if (kDebugMode) {
              AppLogger.debug('onRegularLabelSelected callback is null');
            }
          }
        }
      },
      child: Container(
        padding: const EdgeInsets.all(12),
        margin: const EdgeInsets.only(bottom: 10),
        decoration: BoxDecoration(
          color: labelSize.color.shade50,
          border: Border.all(color: labelSize.color.shade200),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Row(
          children: [
            Icon(labelSize.icon, color: labelSize.color.shade700),
            const SizedBox(width: 12),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    labelSize.displayName,
                    style: TextStyle(
                      fontWeight: FontWeight.bold,
                      fontSize: 16,
                      color: labelSize.color.shade700,
                    ),
                  ),
                  const SizedBox(height: 4),
                  Text(
                    labelSize.description,
                    style: TextStyle(
                      fontSize: 12,
                      color: labelSize.color.shade900,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  Function(PdfColor, bool)? _getCallbackForLabelSize(LabelSize labelSize) {
    switch (labelSize) {
      case LabelSize.small:
        return onA4ShelfSmallSelected;
      case LabelSize.medium:
        return onA4ShelfMediumSelected;
      case LabelSize.large:
        return onA4ShelfLargeSelected;
      case LabelSize.xlarge:
        return onA4ShelfXLargeSelected;
      case LabelSize.regular:
        return null; // Regular doesn't need color selection
    }
  }

  static void _showColorSelectionDialog(
    BuildContext context,
    LabelSize labelSize,
    List<ProductWithCopies> products,
    Function(PdfColor, bool)? onA4FormSelected,
  ) {
    showDialog(
      context: context,
      builder: (dialogContext) => _ColorSelectionDialog(
        labelSize: labelSize,
        products: products,
        onA4FormSelected: onA4FormSelected,
      ),
    );
  }
}

class _ColorSelectionDialog extends StatefulWidget {
  final LabelSize labelSize;
  final List<ProductWithCopies> products;
  final Function(PdfColor priceTextColor, bool showBorder)? onA4FormSelected;

  const _ColorSelectionDialog({
    required this.labelSize,
    required this.products,
    this.onA4FormSelected,
  });

  @override
  State<_ColorSelectionDialog> createState() => _ColorSelectionDialogState();
}

class _ColorSelectionDialogState extends State<_ColorSelectionDialog> {
  PdfColor _priceTextColor = PdfColors.black;
  bool _showBorder = true;

  @override
  Widget build(BuildContext context) {
    return AlertDialog(
      title: Text(global.language('select_price_text_color')),
      content: Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Text(
            '${global.language("select_color_and_border")} (${widget.labelSize.description})',
          ),
          const SizedBox(height: 20),

          // ตัวเลือกแสดงกรอบ
          _buildBorderOptions(),

          const SizedBox(height: 15),

          // ตัวเลือกสีดำ
          _buildColorOption(
            color: PdfColors.black,
            displayColor: Colors.grey,
            backgroundColor: Colors.grey.shade50,
            borderColor: Colors.grey.shade300,
            title: global.language('color_black'),
            description: global.language('label_general_description'),
          ),

          const SizedBox(height: 10),

          // ตัวเลือกสีแดง
          _buildColorOption(
            color: PdfColors.red,
            displayColor: Colors.red,
            backgroundColor: Colors.grey.shade50,
            borderColor: Colors.grey.shade300,
            title: global.language('color_red'),
            description: global.language('label_special_price_description'),
          ),
        ],
      ),
      actions: [
        TextButton(
          onPressed: () => Navigator.pop(context),
          child: Text(global.language('cancel')),
        ),
      ],
    );
  }

  Widget _buildBorderOptions() {
    return Container(
      padding: const EdgeInsets.all(12),
      margin: const EdgeInsets.only(bottom: 15),
      decoration: BoxDecoration(
        color: widget.labelSize.color.shade50,
        border: Border.all(color: widget.labelSize.color.shade200),
        borderRadius: BorderRadius.circular(8),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Text(
            global.language('border_display'),
            style: TextStyle(
              fontSize: 16,
              fontWeight: FontWeight.bold,
              color: widget.labelSize.color.shade700,
            ),
          ),
          const SizedBox(height: 8),
          Row(
            children: [
              Expanded(
                child: _buildBorderOption(
                  showBorder: true,
                  icon: Icons.border_outer,
                  label: global.language('show_border'),
                ),
              ),
              const SizedBox(width: 10),
              Expanded(
                child: _buildBorderOption(
                  showBorder: false,
                  icon: Icons.border_clear,
                  label: global.language('hide_border'),
                ),
              ),
            ],
          ),
        ],
      ),
    );
  }

  Widget _buildBorderOption({
    required bool showBorder,
    required IconData icon,
    required String label,
  }) {
    final bool isSelected = _showBorder == showBorder;

    return InkWell(
      onTap: () {
        setState(() {
          _showBorder = showBorder;
        });
      },
      child: Container(
        padding: const EdgeInsets.symmetric(vertical: 8, horizontal: 12),
        decoration: BoxDecoration(
          color: isSelected ? widget.labelSize.color.shade100 : Colors.white,
          border: Border.all(
            color: isSelected ? widget.labelSize.color : Colors.grey.shade300,
            width: 2,
          ),
          borderRadius: BorderRadius.circular(6),
        ),
        child: Row(
          mainAxisAlignment: MainAxisAlignment.center,
          children: [
            Icon(
              icon,
              color: isSelected ? widget.labelSize.color.shade700 : Colors.grey,
              size: 18,
            ),
            const SizedBox(width: 6),
            Text(
              label,
              style: TextStyle(
                color: isSelected
                    ? widget.labelSize.color.shade700
                    : Colors.grey.shade600,
                fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
              ),
            ),
          ],
        ),
      ),
    );
  }

  Widget _buildColorOption({
    required PdfColor color,
    required Color displayColor,
    required Color backgroundColor,
    required Color borderColor,
    required String title,
    required String description,
  }) {
    return InkWell(
      onTap: () {
        // เซ็ตค่าสีก่อนปิด dialog
        _priceTextColor = color;
        Navigator.pop(context);
        _printLabels();
      },
      child: Container(
        padding: const EdgeInsets.all(12),
        decoration: BoxDecoration(
          color: backgroundColor,
          border: Border.all(color: borderColor, width: 2),
          borderRadius: BorderRadius.circular(8),
        ),
        child: Row(
          children: [
            Container(
              width: 40,
              height: 40,
              decoration: BoxDecoration(
                color: title == global.language('color_black') ? Colors.black : displayColor,
                border: Border.all(color: Colors.grey),
                borderRadius: BorderRadius.circular(20),
              ),
            ),
            SizedBox(width: 16),
            Expanded(
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    title,
                    style: TextStyle(
                      fontSize: 16,
                      fontWeight: FontWeight.bold,
                      color: title == global.language('color_black')
                          ? Colors.grey.shade800
                          : Colors.red.shade700,
                    ),
                  ),
                  Text(
                    description,
                    style: TextStyle(
                      fontSize: 12,
                      color: title == global.language('color_black')
                          ? Colors.grey.shade600
                          : Colors.red.shade600,
                    ),
                  ),
                ],
              ),
            ),
          ],
        ),
      ),
    );
  }

  void _printLabels() {
    if (kDebugMode) {
      AppLogger.debug(
        'PrintLabelDialog: A4 form selected, calling callback function',
      ); // Debug log
      AppLogger.debug(
        'PrintLabelDialog: Products count: ${widget.products.length}',
      ); // Debug log
      AppLogger.debug(
        'PrintLabelDialog: Selected label size: ${widget.labelSize}',
      ); // Debug log
      AppLogger.debug(
        'PrintLabelDialog: Selected color: ${_priceTextColor == PdfColors.red ? "Red" : "Black"}',
      ); // Debug log
      AppLogger.debug('PrintLabelDialog: Show border: $_showBorder'); // Debug log
    }

    // เรียก callback function สำหรับ A4 ฟอร์มต่างๆ พร้อมส่งค่าสีและการแสดงกรอบ
    if (widget.onA4FormSelected != null) {
      widget.onA4FormSelected!(_priceTextColor, _showBorder);
    } else {
      if (kDebugMode) {
        AppLogger.debug('onA4FormSelected callback is null, falling back to direct call');
      }
      // Fallback: เรียกโดยตรงถ้าไม่มี callback (เก็บไว้เป็น backup)
      _directPrintLabels();
    }
  }

  void _directPrintLabels() {
    // แปลง ProductWithCopies เป็น ProductBarcodeModel
    List<ProductBarcodeModel> products = [];
    List<int> copies = [];

    for (var productWithCopies in widget.products) {
      // สร้าง ProductBarcodeModel จากข้อมูลที่มี รวมข้อมูล shelf
      final product = ProductBarcodeModel(
        guidfixed: DateTime.now().millisecondsSinceEpoch.toString(),
        barcode: productWithCopies.product.barcode,
        itemcode: productWithCopies.product.code, // ใช้ code แทน itemcode
        names: productWithCopies.product.name,
        itemunitcode: productWithCopies.product.unitcode,
        itemunitnames: productWithCopies.product.unitname,
        shelfCode: productWithCopies.product.shelfCode, // ข้อมูล shelf
        shelfName: productWithCopies.product.shelfName, // ข้อมูล shelf
        prices: [
          // ใช้ราคาเริ่มต้น 0.0 เนื่องจาก SearchCodeAndNameAndUnitModel ไม่มีข้อมูลราคา
          // ProductLabelPrintA4Shelf จะแสดงข้อความเตือนแทน
        ],
      );

      products.add(product);
      copies.add(productWithCopies.copies);
    }

    // เลือกใช้ class ที่ถูกต้องตาม labelSize
    switch (widget.labelSize) {
      case LabelSize.xlarge:
        ProductLabelPrintA4ShelfXLarge.showPdfPreview(
          context,
          products,
          copies,
          priceTextColor: _priceTextColor,
          showBorder: _showBorder,
        );
        break;
      case LabelSize.large:
        ProductLabelPrintA4ShelfLarge.showPdfPreview(
          context,
          products,
          copies,
          priceTextColor: _priceTextColor,
          showBorder: _showBorder,
        );
        break;
      case LabelSize.medium:
        ProductLabelPrintA4ShelfMedium.showPdfPreview(
          context,
          products,
          copies,
          priceTextColor: _priceTextColor,
          showBorder: _showBorder,
        );
        break;
      case LabelSize.small:
      default:
        ProductLabelPrintA4Shelf.showPdfPreview(
          context,
          products,
          copies,
          showImages: false,
          priceTextColor: _priceTextColor,
          showBorder: _showBorder,
        );
        break;
    }
  }
}
