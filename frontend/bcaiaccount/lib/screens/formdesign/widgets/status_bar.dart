import 'package:flutter/material.dart';
import 'package:smlaicloud/global.dart' as global;
import '../models/form_element.dart';
import '../models/paper_size.dart';
import '../models/form_template.dart';

class FormStatusBar extends StatelessWidget {
  final Offset? mousePosition;
  final FormElement? selectedElement;
  final double zoomLevel;
  final FormTemplate template;
  final bool showGrid;
  final double gridSize;
  final ValueChanged<PaperSize>? onPaperSizeChanged;
  final ValueChanged<PrintMode>? onPrintModeChanged;
  final ValueChanged<double>? onZoomChanged;
  final ValueChanged<PageOrientation>? onOrientationChanged;

  const FormStatusBar({
    super.key,
    this.mousePosition,
    this.selectedElement,
    required this.zoomLevel,
    required this.template,
    required this.showGrid,
    required this.gridSize,
    this.onPaperSizeChanged,
    this.onPrintModeChanged,
    this.onZoomChanged,
    this.onOrientationChanged,
  });

  @override
  Widget build(BuildContext context) {
    final textStyle = TextStyle(
      fontSize: 11,
      color: global.theme.textSecondaryColor,
    );

    return Container(
      height: 28,
      decoration: BoxDecoration(
        color: global.theme.surfaceColor,
        border: Border(
          top: BorderSide(color: global.theme.dividerBorderColor),
        ),
      ),
      padding: const EdgeInsets.symmetric(horizontal: 8),
      child: Row(
        children: [
          // Mouse position
          SizedBox(
            width: 100,
            child: Text(
              mousePosition != null
                  ? 'X: ${mousePosition!.dx.round()}  Y: ${mousePosition!.dy.round()}'
                  : 'X: -  Y: -',
              style: textStyle,
            ),
          ),
          _buildDivider(),
          // Selected element size
          SizedBox(
            width: 100,
            child: Text(
              selectedElement != null
                  ? 'W: ${selectedElement!.width.round()}  H: ${selectedElement!.height.round()}'
                  : 'W: -  H: -',
              style: textStyle,
            ),
          ),
          _buildDivider(),
          // Zoom — clickable
          _buildClickableItem(
            context,
            '${global.language('fd_zoom')}: ${(zoomLevel * 100).round()}%',
            onZoomChanged != null ? () => _showZoomMenu(context) : null,
          ),
          _buildDivider(),
          // Paper size — clickable
          _buildClickableItem(
            context,
            'Paper: ${template.paperSize.name} '
            '(${template.effectiveWidth.round()}x${template.effectiveHeight.round()})',
            onPaperSizeChanged != null
                ? () => _showPaperSizeMenu(context)
                : null,
          ),
          _buildDivider(),
          // Orientation — clickable toggle
          _buildClickableItem(
            context,
            template.orientation.label,
            onOrientationChanged != null
                ? () => _toggleOrientation()
                : null,
          ),
          _buildDivider(),
          // Print mode — clickable
          _buildClickableItem(
            context,
            '${global.language('fd_print_mode')}: ${template.printMode.label}',
            onPrintModeChanged != null
                ? () => _showPrintModeMenu(context)
                : null,
          ),
          _buildDivider(),
          // Grid
          Text(
            showGrid
                ? '${global.language('fd_grid')}: ${gridSize.round()}px'
                : global.language('fd_grid_off'),
            style: textStyle,
          ),
          const Spacer(),
        ],
      ),
    );
  }

  Widget _buildClickableItem(
      BuildContext context, String text, VoidCallback? onTap) {
    final style = TextStyle(
      fontSize: 11,
      color: onTap != null
          ? global.theme.primaryColor
          : global.theme.textSecondaryColor,
    );
    if (onTap == null) return Text(text, style: style);

    return InkWell(
      onTap: onTap,
      borderRadius: BorderRadius.circular(4),
      child: Padding(
        padding: const EdgeInsets.symmetric(horizontal: 4, vertical: 2),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Text(text, style: style),
            const SizedBox(width: 2),
            Icon(Icons.arrow_drop_down,
                size: 14, color: global.theme.primaryColor),
          ],
        ),
      ),
    );
  }

  void _toggleOrientation() {
    final newOrientation = template.orientation == PageOrientation.portrait
        ? PageOrientation.landscape
        : PageOrientation.portrait;
    onOrientationChanged?.call(newOrientation);
  }

  void _showPaperSizeMenu(BuildContext context) {
    final RenderBox box = context.findRenderObject() as RenderBox;
    final pos = box.localToGlobal(Offset.zero);

    showMenu<PaperSize>(
      context: context,
      position: RelativeRect.fromLTRB(
          pos.dx + 300, pos.dy - PaperSize.values.length * 40, pos.dx + 500, pos.dy),
      color: global.theme.dialogColor,
      items: PaperSize.values.map((size) {
        final selected = size == template.paperSize;
        return PopupMenuItem<PaperSize>(
          value: size,
          child: Row(
            children: [
              SizedBox(
                width: 20,
                child: selected
                    ? Icon(Icons.check, size: 16, color: global.theme.primaryColor)
                    : null,
              ),
              const SizedBox(width: 4),
              Text('${size.name}  ${size.description}',
                  style: TextStyle(
                    color: global.theme.textColor,
                    fontWeight: selected ? FontWeight.bold : FontWeight.normal,
                  )),
            ],
          ),
        );
      }).toList(),
    ).then((v) {
      if (v != null) onPaperSizeChanged?.call(v);
    });
  }

  void _showPrintModeMenu(BuildContext context) {
    final RenderBox box = context.findRenderObject() as RenderBox;
    final pos = box.localToGlobal(Offset.zero);

    showMenu<PrintMode>(
      context: context,
      position: RelativeRect.fromLTRB(
          pos.dx + 450, pos.dy - PrintMode.values.length * 40, pos.dx + 600, pos.dy),
      color: global.theme.dialogColor,
      items: PrintMode.values.map((mode) {
        final selected = mode == template.printMode;
        return PopupMenuItem<PrintMode>(
          value: mode,
          child: Row(
            children: [
              SizedBox(
                width: 20,
                child: selected
                    ? Icon(Icons.check, size: 16, color: global.theme.primaryColor)
                    : null,
              ),
              const SizedBox(width: 4),
              Text(mode.label,
                  style: TextStyle(
                    color: global.theme.textColor,
                    fontWeight: selected ? FontWeight.bold : FontWeight.normal,
                  )),
            ],
          ),
        );
      }).toList(),
    ).then((v) {
      if (v != null) onPrintModeChanged?.call(v);
    });
  }

  void _showZoomMenu(BuildContext context) {
    final levels = [0.25, 0.5, 0.75, 1.0, 1.25, 1.5, 2.0, 3.0];
    final RenderBox box = context.findRenderObject() as RenderBox;
    final pos = box.localToGlobal(Offset.zero);

    final items = <PopupMenuEntry<double>>[
      // Fit options
      PopupMenuItem<double>(
        value: -1.0,
        child: Row(children: [
          const SizedBox(width: 20),
          const SizedBox(width: 4),
          Text('พอดีหน้าจอ (Fit)',
              style: TextStyle(color: global.theme.textColor)),
        ]),
      ),
      PopupMenuItem<double>(
        value: -2.0,
        child: Row(children: [
          const SizedBox(width: 20),
          const SizedBox(width: 4),
          Text('พอดีแนวกว้าง (Fit Width)',
              style: TextStyle(color: global.theme.textColor)),
        ]),
      ),
      const PopupMenuDivider(),
      // Fixed zoom levels
      ...levels.map((z) {
        final selected = (z - zoomLevel).abs() < 0.01;
        return PopupMenuItem<double>(
          value: z,
          child: Row(
            children: [
              SizedBox(
                width: 20,
                child: selected
                    ? Icon(Icons.check, size: 16, color: global.theme.primaryColor)
                    : null,
              ),
              const SizedBox(width: 4),
              Text('${(z * 100).round()}%',
                  style: TextStyle(
                    color: global.theme.textColor,
                    fontWeight: selected ? FontWeight.bold : FontWeight.normal,
                  )),
            ],
          ),
        );
      }),
    ];

    showMenu<double>(
      context: context,
      position: RelativeRect.fromLTRB(
          pos.dx + 200, pos.dy - items.length * 36, pos.dx + 380, pos.dy),
      color: global.theme.dialogColor,
      items: items,
    ).then((v) {
      if (v != null) onZoomChanged?.call(v);
    });
  }

  Widget _buildDivider() {
    return Padding(
      padding: const EdgeInsets.symmetric(horizontal: 8),
      child: Container(
        width: 1,
        height: 16,
        color: global.theme.dividerBorderColor,
      ),
    );
  }
}
