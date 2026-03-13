import 'dart:io';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import '../models/form_element.dart';
import '../models/form_section.dart';
import '../models/form_template.dart';
import '../models/data_binding.dart';
import 'ruler_widget.dart';
import 'package:smlaicloud/global.dart' as global;

class FullFormEditor extends StatefulWidget {
  final FormTemplate template;
  final Function(FormTemplate) onTemplateChanged;
  final double zoomLevel;
  final bool showRulers;
  final bool showGrid;
  final bool snapToGrid;
  final double gridSize;
  final FormElement? selectedElement;
  final Set<String> selectedElementIds;
  final Function(FormElement?, SectionType?, {bool isShift})? onElementSelected;
  final VoidCallback? onElementDeselected;
  final Function(Offset?)? onMousePositionChanged;
  final Function(ElementType type, SectionType section, Offset position)?
      onElementAdd;
  final Function(SectionType section, String elementId, double oldX,
      double oldY, double newX, double newY)? onElementMoved;
  final Function(SectionType section, String elementId, double oldX,
      double oldY, double oldW, double oldH, double newX, double newY,
      double newW, double newH)? onElementResized;

  const FullFormEditor({
    super.key,
    required this.template,
    required this.onTemplateChanged,
    this.zoomLevel = 1.0,
    this.showRulers = true,
    this.showGrid = false,
    this.snapToGrid = false,
    this.gridSize = 10,
    this.selectedElement,
    this.selectedElementIds = const {},
    this.onElementSelected,
    this.onElementDeselected,
    this.onMousePositionChanged,
    this.onElementAdd,
    this.onElementMoved,
    this.onElementResized,
  });

  @override
  State<FullFormEditor> createState() => _FullFormEditorState();
}

class _FullFormEditorState extends State<FullFormEditor>
    with global.ThemeRefreshMixin {
  // ── Pointer tracking ──
  int? _activePointer;
  String? _activeAction; // 'drag' or 'resize'

  // ── Drag state ──
  String? _draggingId;
  SectionType? _draggingSectionType;
  double _dragRawX = 0;
  double _dragRawY = 0;
  double _dragCurrentX = 0;
  double _dragCurrentY = 0;
  double _dragStartX = 0;
  double _dragStartY = 0;
  double _dragElementW = 0;
  double _dragElementH = 0;

  // ── Resize state ──
  String? _resizingId;
  SectionType? _resizingSectionType;
  Alignment? _resizeHandle;
  double _resizeStartX = 0;
  double _resizeStartY = 0;
  double _resizeStartW = 0;
  double _resizeStartH = 0;
  double _resizeCurrentX = 0;
  double _resizeCurrentY = 0;
  double _resizeCurrentW = 0;
  double _resizeCurrentH = 0;
  double _resizeSectionH = 0;

  // Canvas key for coordinate conversion
  final GlobalKey _canvasKey = GlobalKey();

  double _snap(double value) {
    if (!widget.snapToGrid) return value;
    return (value / widget.gridSize).round() * widget.gridSize;
  }

  FormSection _getSection(SectionType type) {
    switch (type) {
      case SectionType.header:
        return widget.template.header;
      case SectionType.detail:
        return widget.template.detail;
      case SectionType.footer:
        return widget.template.footer;
    }
  }

  double _sectionOffsetY(SectionType type) {
    final headerH = _resolveSectionHeight(widget.template.header);
    final paperH = widget.template.effectiveHeight;
    final footerH = _resolveSectionHeight(widget.template.footer);
    switch (type) {
      case SectionType.header:
        return 0;
      case SectionType.detail:
        return headerH;
      case SectionType.footer:
        // Footer anchored at bottom of page
        return paperH - footerH;
    }
  }

  /// Detail height fills remaining space between header and footer
  double _detailDisplayHeight() {
    final paperH = widget.template.effectiveHeight;
    final headerH = _resolveSectionHeight(widget.template.header);
    final footerH = _resolveSectionHeight(widget.template.footer);
    return (paperH - headerH - footerH).clamp(20.0, double.infinity);
  }

  FormTemplate _updateSection(SectionType type, FormSection section) {
    switch (type) {
      case SectionType.header:
        return widget.template.copyWith(header: section);
      case SectionType.detail:
        return widget.template.copyWith(detail: section);
      case SectionType.footer:
        return widget.template.copyWith(footer: section);
    }
  }

  void _commitElement(SectionType sectionType, FormElement element) {
    final section = _getSection(sectionType);

    if (section.hasRows) {
      // Row-based: find cell by id and update its properties
      final updatedRows = section.rows.map((row) {
        final updatedCells = row.cells.map((cell) {
          if (cell.id != element.id) return cell;
          return cell.copyWith(
            text: element.text,
            fontSize: element.fontSize,
            fontFamily: element.fontFamily,
            fontBold: element.fontBold,
            fontItalic: element.fontItalic,
            fontUnderline: element.fontUnderline,
            textColor: element.textColor,
            textAlign: element.textAlign,
            backgroundColor: element.backgroundColor,
            borderColor: element.borderColor,
            borderWidth: element.borderWidth,
            borderRadius: element.borderRadius,
            opacity: element.opacity,
            dataBindingKey: element.dataBindingKey,
            dataBindingFormat: element.dataBindingFormat,
            imagePath: element.imagePath,
            barcodeFormat: element.barcodeFormat,
            barcodeValue: element.barcodeValue,
            lineStyle: element.lineStyle,
            rows: element.rows,
            columns: element.columns,
            tableData: element.tableData,
            tableHasHeader: element.tableHasHeader,
          );
        }).toList();
        return row.copyWith(cells: updatedCells);
      }).toList();
      final updatedSection = section.copyWith(rows: updatedRows);
      widget.onTemplateChanged(_updateSection(sectionType, updatedSection));
    } else {
      // Fixed: update element directly
      final updatedElements = section.elements.map((e) {
        return e.id == element.id ? element : e;
      }).toList();
      final updatedSection = section.copyWith(elements: updatedElements);
      widget.onTemplateChanged(_updateSection(sectionType, updatedSection));
    }
  }

  // ── Hit testing ──

  Offset? _canvasLocalPosition(Offset globalPos) {
    final renderBox =
        _canvasKey.currentContext?.findRenderObject() as RenderBox?;
    if (renderBox == null) return null;
    return renderBox.globalToLocal(globalPos);
  }

  /// Hit test resize handles of the selected element. Returns handle alignment or null.
  Alignment? _hitTestResizeHandle(double canvasX, double canvasY) {
    final sel = widget.selectedElement;
    if (sel == null || sel.isLocked) return null;

    // Find which section the selected element is in
    SectionType? selSection;
    for (final st in SectionType.values) {
      if (_resolveElements(_getSection(st)).any((e) => e.id == sel.id)) {
        selSection = st;
        break;
      }
    }
    if (selSection == null) return null;

    final offsetY = _sectionOffsetY(selSection);
    final ex = sel.x;
    final ey = offsetY + sel.y;
    final ew = sel.width;
    final eh = sel.height;
    const r = 8.0; // handle hit radius in canvas coords

    final handles = <Alignment, Offset>{
      Alignment.topLeft: Offset(ex, ey),
      Alignment.topCenter: Offset(ex + ew / 2, ey),
      Alignment.topRight: Offset(ex + ew, ey),
      Alignment.centerLeft: Offset(ex, ey + eh / 2),
      Alignment.centerRight: Offset(ex + ew, ey + eh / 2),
      Alignment.bottomLeft: Offset(ex, ey + eh),
      Alignment.bottomCenter: Offset(ex + ew / 2, ey + eh),
      Alignment.bottomRight: Offset(ex + ew, ey + eh),
    };

    for (final entry in handles.entries) {
      if ((canvasX - entry.value.dx).abs() < r &&
          (canvasY - entry.value.dy).abs() < r) {
        return entry.key;
      }
    }
    return null;
  }

  /// Hit test elements. Returns the topmost visible element at position.
  ({FormElement element, SectionType sectionType})? _hitTestElement(
      double canvasX, double canvasY) {
    // Check sections in reverse render order (footer on top of detail, etc.)
    for (final st in [SectionType.footer, SectionType.detail, SectionType.header]) {
      final section = _getSection(st);
      final offsetY = _sectionOffsetY(st);
      final localY = canvasY - offsetY;

      final sectionH = st == SectionType.detail
          ? _detailDisplayHeight()
          : _resolveSectionHeight(section);
      if (localY < 0 || localY > sectionH) continue;

      // Check elements in reverse order (last = topmost in Stack)
      for (final element in _resolveElements(section).reversed) {
        if (!element.isVisible) continue;
        if (canvasX >= element.x &&
            canvasX <= element.x + element.width &&
            localY >= element.y &&
            localY <= element.y + element.height) {
          return (element: element, sectionType: st);
        }
      }
    }
    return null;
  }

  // ── Pointer handlers (canvas-level, immediate response) ──

  void _onCanvasPointerDown(PointerDownEvent event, double zoom) {
    if (_activePointer != null) return;
    final localPos = _canvasLocalPosition(event.position);
    if (localPos == null) return;

    final canvasX = localPos.dx / zoom;
    final canvasY = localPos.dy / zoom;
    final isShift = HardwareKeyboard.instance.isShiftPressed;

    // 1. Check resize handles first
    final handle = _hitTestResizeHandle(canvasX, canvasY);
    if (handle != null) {
      _activePointer = event.pointer;
      _activeAction = 'resize';
      _startResize(widget.selectedElement!, handle);
      return;
    }

    // 2. Hit test elements
    final hit = _hitTestElement(canvasX, canvasY);
    if (hit != null) {
      _activePointer = event.pointer;
      _activeAction = 'drag';
      _startDrag(hit.element, hit.sectionType);
      widget.onElementSelected?.call(hit.element, hit.sectionType, isShift: isShift);
      return;
    }

    // 3. Empty space → deselect
    widget.onElementDeselected?.call();
  }

  void _onCanvasPointerMove(PointerMoveEvent event, double zoom) {
    // Update mouse position (always)
    final localPos = _canvasLocalPosition(event.position);
    if (localPos != null) {
      widget.onMousePositionChanged?.call(Offset(
        localPos.dx / zoom,
        localPos.dy / zoom,
      ));
    }

    if (_activePointer != event.pointer) return;

    final dx = event.delta.dx / zoom;
    final dy = event.delta.dy / zoom;

    if (_activeAction == 'drag' && _draggingId != null) {
      setState(() {
        _dragRawX += dx;
        _dragRawY += dy;
        _dragCurrentX = _snap(_dragRawX)
            .clamp(0.0, widget.template.effectiveWidth - _dragElementW);
        final dragSectionH = _draggingSectionType == SectionType.detail
            ? _detailDisplayHeight()
            : _resolveSectionHeight(_getSection(_draggingSectionType!));
        _dragCurrentY = _snap(_dragRawY).clamp(0.0, dragSectionH - _dragElementH);
      });
    } else if (_activeAction == 'resize' && _resizingId != null) {
      _updateResize(dx, dy);
    }
  }

  void _onCanvasPointerUp(PointerUpEvent event, double zoom) {
    if (_activePointer != event.pointer) return;

    if (_activeAction == 'drag') {
      _finishDrag();
    } else if (_activeAction == 'resize') {
      _finishResize();
    }

    _activePointer = null;
    _activeAction = null;
  }

  // ── Drag ──

  void _startDrag(FormElement element, SectionType sectionType) {
    if (element.isLocked) return;
    setState(() {
      _draggingId = element.id;
      _draggingSectionType = sectionType;
      _dragStartX = element.x;
      _dragStartY = element.y;
      _dragRawX = element.x;
      _dragRawY = element.y;
      _dragCurrentX = element.x;
      _dragCurrentY = element.y;
      _dragElementW = element.width;
      _dragElementH = element.height;
    });
  }

  void _finishDrag() {
    if (_draggingId == null) return;
    final sectionType = _draggingSectionType!;
    final section = _getSection(sectionType);

    final newX = _snap(_dragRawX)
        .clamp(0.0, widget.template.effectiveWidth - _dragElementW);
    final sectionH = sectionType == SectionType.detail
        ? _detailDisplayHeight()
        : _resolveSectionHeight(section);
    final newY = _snap(_dragRawY).clamp(0.0, sectionH - _dragElementH);
    final oldX = _dragStartX;
    final oldY = _dragStartY;

    final element =
        _resolveElements(section).where((e) => e.id == _draggingId).firstOrNull;

    setState(() {
      _draggingId = null;
      _draggingSectionType = null;
    });

    if (element == null) return;

    if ((newX - oldX).abs() > 0.5 || (newY - oldY).abs() > 0.5) {
      final updated = element.copyWith(x: newX, y: newY);
      _commitElement(sectionType, updated);
      widget.onElementSelected?.call(updated, sectionType, isShift: false);
      widget.onElementMoved
          ?.call(sectionType, element.id, oldX, oldY, newX, newY);
    }
  }

  // ── Resize ──

  void _startResize(FormElement element, Alignment handle) {
    if (element.isLocked) return;
    SectionType? sType;
    for (final st in SectionType.values) {
      if (_resolveElements(_getSection(st)).any((e) => e.id == element.id)) {
        sType = st;
        break;
      }
    }
    if (sType == null) return;

    setState(() {
      _resizingId = element.id;
      _resizingSectionType = sType;
      _resizeHandle = handle;
      _resizeStartX = element.x;
      _resizeStartY = element.y;
      _resizeStartW = element.width;
      _resizeStartH = element.height;
      _resizeCurrentX = element.x;
      _resizeCurrentY = element.y;
      _resizeCurrentW = element.width;
      _resizeCurrentH = element.height;
      _resizeSectionH = sType == SectionType.detail
          ? _detailDisplayHeight()
          : _resolveSectionHeight(_getSection(sType!));
    });
  }

  void _updateResize(double dx, double dy) {
    if (_resizingId == null || _resizeHandle == null) return;
    const minSize = 10.0;

    setState(() {
      final handle = _resizeHandle!;
      double x = _resizeCurrentX;
      double y = _resizeCurrentY;
      double w = _resizeCurrentW;
      double h = _resizeCurrentH;

      // Horizontal
      if (handle == Alignment.topLeft ||
          handle == Alignment.centerLeft ||
          handle == Alignment.bottomLeft) {
        final newX = _snap(x + dx);
        final maxX = x + w - minSize;
        x = newX.clamp(0.0, maxX);
        w = (_resizeCurrentX + _resizeCurrentW) - x;
      } else if (handle == Alignment.topRight ||
          handle == Alignment.centerRight ||
          handle == Alignment.bottomRight) {
        w = _snap(w + dx)
            .clamp(minSize, widget.template.effectiveWidth - x);
      }

      // Vertical
      if (handle == Alignment.topLeft ||
          handle == Alignment.topCenter ||
          handle == Alignment.topRight) {
        final newY = _snap(y + dy);
        final maxY = y + h - minSize;
        y = newY.clamp(0.0, maxY);
        h = (_resizeCurrentY + _resizeCurrentH) - y;
      } else if (handle == Alignment.bottomLeft ||
          handle == Alignment.bottomCenter ||
          handle == Alignment.bottomRight) {
        h = _snap(h + dy).clamp(minSize, _resizeSectionH - y);
      }

      _resizeCurrentX = x;
      _resizeCurrentY = y;
      _resizeCurrentW = w;
      _resizeCurrentH = h;
    });
  }

  void _finishResize() {
    if (_resizingId == null || _resizingSectionType == null) return;
    final sectionType = _resizingSectionType!;
    final section = _getSection(sectionType);
    final element =
        _resolveElements(section).where((e) => e.id == _resizingId).firstOrNull;

    final newX = _resizeCurrentX;
    final newY = _resizeCurrentY;
    final newW = _resizeCurrentW;
    final newH = _resizeCurrentH;

    setState(() {
      _resizingId = null;
      _resizingSectionType = null;
      _resizeHandle = null;
    });

    if (element == null) return;

    if ((newX - _resizeStartX).abs() > 0.5 ||
        (newY - _resizeStartY).abs() > 0.5 ||
        (newW - _resizeStartW).abs() > 0.5 ||
        (newH - _resizeStartH).abs() > 0.5) {
      final updated =
          element.copyWith(x: newX, y: newY, width: newW, height: newH);
      _commitElement(sectionType, updated);
      widget.onElementSelected?.call(updated, sectionType, isShift: false);
      widget.onElementResized?.call(
        sectionType, element.id,
        _resizeStartX, _resizeStartY, _resizeStartW, _resizeStartH,
        newX, newY, newW, newH,
      );
    }
  }

  // ── Print Mode Filter ──

  Widget _wrapPrintMode(Widget child) {
    switch (widget.template.printMode) {
      case PrintMode.grayscale:
        return ColorFiltered(
          colorFilter: const ColorFilter.matrix(<double>[
            0.2126, 0.7152, 0.0722, 0, 0,
            0.2126, 0.7152, 0.0722, 0, 0,
            0.2126, 0.7152, 0.0722, 0, 0,
            0,      0,      0,      1, 0,
          ]),
          child: child,
        );
      case PrintMode.blackAndWhite:
        // Grayscale + high contrast threshold
        return ColorFiltered(
          colorFilter: const ColorFilter.matrix(<double>[
            0.2126 * 3, 0.7152 * 3, 0.0722 * 3, 0, -128,
            0.2126 * 3, 0.7152 * 3, 0.0722 * 3, 0, -128,
            0.2126 * 3, 0.7152 * 3, 0.0722 * 3, 0, -128,
            0,          0,          0,           1,    0,
          ]),
          child: child,
        );
      case PrintMode.color:
        return child;
    }
  }

  // ── Build ──

  @override
  Widget build(BuildContext context) {
    final double actualZoom = widget.zoomLevel <= 0 ? 1.0 : widget.zoomLevel;

    return Container(
      color: global.theme.backgroundColor,
      child: widget.zoomLevel == -1
          ? _buildFullScreenCanvas()
          : SingleChildScrollView(
              scrollDirection: Axis.horizontal,
              child: SingleChildScrollView(
                scrollDirection: Axis.vertical,
                child: Padding(
                  padding: const EdgeInsets.all(20),
                  child: _buildCanvasWithRulers(actualZoom),
                ),
              ),
            ),
    );
  }

  Widget _buildFullScreenCanvas() {
    return LayoutBuilder(
      builder: (context, constraints) {
        final scaleX =
            (constraints.maxWidth - 60) / widget.template.effectiveWidth;
        final scaleY =
            (constraints.maxHeight - 60) / widget.template.effectiveHeight;
        final scale = scaleX < scaleY ? scaleX : scaleY;
        return Center(
          child: Padding(
            padding: const EdgeInsets.all(20),
            child: _buildCanvasWithRulers(scale),
          ),
        );
      },
    );
  }

  Widget _buildCanvasWithRulers(double zoom) {
    return Row(
      crossAxisAlignment: CrossAxisAlignment.start,
      mainAxisSize: MainAxisSize.min,
      children: [
        if (widget.showRulers)
          Column(
            children: [
              SizedBox(width: 25, height: 25),
              VerticalRuler(
                  height: widget.template.effectiveHeight, zoom: zoom),
            ],
          ),
        Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            if (widget.showRulers)
              HorizontalRuler(
                  width: widget.template.effectiveWidth, zoom: zoom),
            _buildCanvas(zoom),
          ],
        ),
      ],
    );
  }

  Widget _buildCanvas(double zoom) {
    final paperWidth = widget.template.effectiveWidth * zoom;
    final paperHeight = widget.template.effectiveHeight * zoom;

    return MouseRegion(
      onHover: (event) {
        final localPos = _canvasLocalPosition(event.position);
        if (localPos != null) {
          widget.onMousePositionChanged?.call(Offset(
            localPos.dx / zoom,
            localPos.dy / zoom,
          ));
        }
      },
      onExit: (_) => widget.onMousePositionChanged?.call(null),
      child: Listener(
        behavior: HitTestBehavior.translucent,
        onPointerDown: (e) => _onCanvasPointerDown(e, zoom),
        onPointerMove: (e) => _onCanvasPointerMove(e, zoom),
        onPointerUp: (e) => _onCanvasPointerUp(e, zoom),
        child: DragTarget<Map<String, dynamic>>(
          onAcceptWithDetails: (details) {
            final data = details.data;
            final type = data['type'] as ElementType;
            final localPos = _canvasLocalPosition(details.offset);
            if (localPos == null) return;
            final canvasX = (localPos.dx / zoom)
                .clamp(0.0, widget.template.effectiveWidth - 50);
            final canvasY = localPos.dy / zoom;

            final hdrH = _resolveSectionHeight(widget.template.header);
            final footerOffsetY = _sectionOffsetY(SectionType.footer);
            SectionType section;
            double adjustedY;
            if (canvasY < hdrH) {
              section = SectionType.header;
              adjustedY = canvasY;
            } else if (canvasY < footerOffsetY) {
              section = SectionType.detail;
              adjustedY = canvasY - hdrH;
            } else {
              section = SectionType.footer;
              adjustedY = canvasY - footerOffsetY;
            }

            widget.onElementAdd?.call(
              type,
              section,
              Offset(_snap(canvasX),
                  _snap(adjustedY.clamp(0, double.infinity))),
            );
          },
          builder: (context, candidateData, rejectedData) {
            final isDragOver = candidateData.isNotEmpty;
            return AnimatedContainer(
              key: _canvasKey,
              duration: const Duration(milliseconds: 150),
              width: paperWidth,
              height: paperHeight,
              decoration: BoxDecoration(
                color: global.theme.cardColor,
                border: Border.all(
                  color: isDragOver
                      ? global.theme.primaryColor
                      : global.theme.dividerBorderColor,
                  width: isDragOver ? 2 : 1,
                ),
                boxShadow: [
                  BoxShadow(
                    color: Colors.black
                        .withValues(alpha: isDragOver ? 0.25 : 0.15),
                    blurRadius: isDragOver ? 16 : 8,
                    offset: const Offset(2, 2),
                  ),
                ],
                image: widget.template.globalBackgroundImagePath != null
                    ? DecorationImage(
                        image: FileImage(File(
                            widget.template.globalBackgroundImagePath!)),
                        fit: BoxFit.cover,
                        opacity: 0.1,
                      )
                    : null,
              ),
              child: _wrapPrintMode(
                Stack(
                  clipBehavior: Clip.none,
                  children: [
                    // Grid overlay
                    if (widget.showGrid)
                      Positioned.fill(
                        child: CustomPaint(
                          painter: _GridPainter(
                            gridSize: widget.gridSize,
                            zoom: zoom,
                            color: global.theme.dividerBorderColor
                                .withValues(alpha: 0.3),
                          ),
                        ),
                      ),
                    // Section dividers
                    _buildSectionDivider(
                        _sectionOffsetY(SectionType.detail) * zoom,
                        global.language('fd_header'),
                        zoom),
                    _buildSectionDivider(
                        _sectionOffsetY(SectionType.footer) * zoom,
                        global.language('fd_footer'),
                        zoom),
                    // Elements
                    ..._buildAllElements(zoom),
                  ],
                ),
              ),
            );
          },
        ),
      ),
    );
  }

  Widget _buildSectionDivider(double y, String label, double zoom) {
    return Positioned(
      left: 0,
      top: y,
      right: 0,
      child: Row(
        children: [
          Expanded(
            child: Container(
              height: 1,
              color: global.theme.primaryColor.withValues(alpha: 0.4),
            ),
          ),
          Container(
            padding: const EdgeInsets.symmetric(horizontal: 6, vertical: 1),
            decoration: BoxDecoration(
              color: global.theme.primaryColor.withValues(alpha: 0.15),
              borderRadius: BorderRadius.circular(3),
            ),
            child: Text(
              label,
              style: TextStyle(
                fontSize: 9,
                color: global.theme.primaryColor,
                fontWeight: FontWeight.w500,
              ),
            ),
          ),
          Expanded(
            child: Container(
              height: 1,
              color: global.theme.primaryColor.withValues(alpha: 0.4),
            ),
          ),
        ],
      ),
    );
  }

  /// Resolve elements from section — ใช้ rows (flex) ถ้ามี, fallback to elements (fixed)
  List<FormElement> _resolveElements(FormSection section) {
    if (section.hasRows) {
      return section.resolveElements(widget.template.effectiveWidth);
    }
    return section.elements;
  }

  /// Resolve section height — auto จาก rows ถ้ามี
  double _resolveSectionHeight(FormSection section) {
    if (section.hasRows) {
      return section.resolveHeight(widget.template.effectiveWidth);
    }
    return section.height;
  }

  List<Widget> _buildAllElements(double zoom) {
    final widgets = <Widget>[];

    void addSection(
        FormSection section, double offsetY, SectionType sectionType) {
      for (final element
          in _resolveElements(section).where((e) => e.isVisible)) {
        widgets.add(_buildPositionedElement(
            element, offsetY, sectionType, section, zoom));
      }
    }

    addSection(widget.template.header, _sectionOffsetY(SectionType.header),
        SectionType.header);
    addSection(widget.template.detail, _sectionOffsetY(SectionType.detail),
        SectionType.detail);
    addSection(widget.template.footer, _sectionOffsetY(SectionType.footer),
        SectionType.footer);

    return widgets;
  }

  Widget _buildPositionedElement(FormElement element, double offsetY,
      SectionType sectionType, FormSection section, double zoom) {
    final isSelected = widget.selectedElement?.id == element.id;
    final isMultiSelected = widget.selectedElementIds.contains(element.id);
    final isDragging = _draggingId == element.id;
    final isResizing = _resizingId == element.id;

    double displayX, displayY, displayW, displayH;
    if (isDragging) {
      displayX = _dragCurrentX;
      displayY = _dragCurrentY;
      displayW = element.width;
      displayH = element.height;
    } else if (isResizing) {
      displayX = _resizeCurrentX;
      displayY = _resizeCurrentY;
      displayW = _resizeCurrentW;
      displayH = _resizeCurrentH;
    } else {
      displayX = element.x;
      displayY = element.y;
      displayW = element.width;
      displayH = element.height;
    }

    final left = displayX * zoom;
    final top = (offsetY + displayY) * zoom;

    Widget child = _buildElementWidget(
      element, zoom, isSelected, isMultiSelected, isDragging, isResizing,
      displayW, displayH,
    );

    // Drag shadow
    if (isDragging) {
      child = DecoratedBox(
        decoration: BoxDecoration(
          boxShadow: [
            BoxShadow(
              color: global.theme.primaryColor.withValues(alpha: 0.3),
              blurRadius: 14,
              offset: const Offset(0, 4),
            ),
          ],
        ),
        child: child,
      );
    }

    // Smooth animation for non-interactive elements
    if (isDragging || isResizing) {
      return Positioned(left: left, top: top, child: child);
    }

    return AnimatedPositioned(
      duration: const Duration(milliseconds: 120),
      curve: Curves.easeOutCubic,
      left: left,
      top: top,
      child: child,
    );
  }

  Widget _buildElementWidget(
    FormElement element,
    double zoom,
    bool isSelected,
    bool isMultiSelected,
    bool isDragging,
    bool isResizing,
    double displayW,
    double displayH,
  ) {
    Widget content = _buildElementContent(element);

    if (element.opacity < 1.0) {
      content = Opacity(opacity: element.opacity, child: content);
    }
    if (element.rotation != 0) {
      content = Transform.rotate(
        angle: element.rotation * 3.14159265 / 180,
        child: content,
      );
    }

    final dragOpacity = isDragging ? 0.85 : 1.0;

    // Border style: selected > multi-selected > normal
    Border? border;
    if (isSelected) {
      border = Border.all(color: global.theme.primaryColor, width: 2);
    } else if (isMultiSelected) {
      border = Border.all(
          color: global.theme.primaryColor.withValues(alpha: 0.6), width: 1.5);
    } else if (element.borderColor != null && element.borderWidth != null) {
      border = Border.all(
        color: element.borderColor!,
        width: (element.borderWidth ?? 1) * zoom,
      );
    }

    return Opacity(
      opacity: dragOpacity,
      child: AnimatedContainer(
        duration:
            (isDragging || isResizing) ? Duration.zero : const Duration(milliseconds: 120),
        width: displayW * zoom,
        height: displayH * zoom,
        decoration: BoxDecoration(
          color: element.backgroundColor,
          borderRadius: element.borderRadius != null
              ? BorderRadius.circular(element.borderRadius! * zoom)
              : null,
          border: border,
        ),
        child: Stack(
          clipBehavior: Clip.none,
          children: [
            Positioned.fill(
              child: ClipRRect(
                borderRadius: element.borderRadius != null
                    ? BorderRadius.circular(element.borderRadius! * zoom)
                    : BorderRadius.zero,
                child: Transform.scale(
                  scale: zoom,
                  alignment: Alignment.topLeft,
                  child: SizedBox(
                    width: element.width,
                    height: element.height,
                    child: content,
                  ),
                ),
              ),
            ),
            // Resize handles (visual only — hit testing is canvas-level)
            if (isSelected && !element.isLocked)
              ..._buildResizeHandles(element, zoom, isResizing),
            // Lock indicator
            if (element.isLocked && isSelected)
              Positioned(
                right: 2,
                top: 2,
                child: Icon(Icons.lock,
                    size: 12, color: global.theme.primaryColor),
              ),
            // Multi-select indicator
            if (isMultiSelected && !isSelected)
              Positioned(
                left: 2,
                top: 2,
                child: Container(
                  width: 14,
                  height: 14,
                  decoration: BoxDecoration(
                    color: global.theme.primaryColor,
                    borderRadius: BorderRadius.circular(3),
                  ),
                  child: Icon(Icons.check,
                      size: 10, color: global.theme.onPrimaryColor),
                ),
              ),
          ],
        ),
      ),
    );
  }

  List<Widget> _buildResizeHandles(
      FormElement element, double zoom, bool isResizing) {
    const handleSize = 10.0;
    final positions = [
      Alignment.topLeft,
      Alignment.topCenter,
      Alignment.topRight,
      Alignment.centerLeft,
      Alignment.centerRight,
      Alignment.bottomLeft,
      Alignment.bottomCenter,
      Alignment.bottomRight,
    ];

    final w = isResizing ? _resizeCurrentW * zoom : element.width * zoom;
    final h = isResizing ? _resizeCurrentH * zoom : element.height * zoom;

    return positions.map((alignment) {
      double left = 0, top = 0;

      if (alignment == Alignment.topLeft) {
        left = -handleSize / 2;
        top = -handleSize / 2;
      } else if (alignment == Alignment.topCenter) {
        left = w / 2 - handleSize / 2;
        top = -handleSize / 2;
      } else if (alignment == Alignment.topRight) {
        left = w - handleSize / 2;
        top = -handleSize / 2;
      } else if (alignment == Alignment.centerLeft) {
        left = -handleSize / 2;
        top = h / 2 - handleSize / 2;
      } else if (alignment == Alignment.centerRight) {
        left = w - handleSize / 2;
        top = h / 2 - handleSize / 2;
      } else if (alignment == Alignment.bottomLeft) {
        left = -handleSize / 2;
        top = h - handleSize / 2;
      } else if (alignment == Alignment.bottomCenter) {
        left = w / 2 - handleSize / 2;
        top = h - handleSize / 2;
      } else if (alignment == Alignment.bottomRight) {
        left = w - handleSize / 2;
        top = h - handleSize / 2;
      }

      final isActive = _resizingId == element.id && _resizeHandle == alignment;

      return Positioned(
        left: left,
        top: top,
        child: Container(
          width: handleSize,
          height: handleSize,
          decoration: BoxDecoration(
            color: isActive
                ? global.theme.primaryColor
                : global.theme.cardColor,
            border:
                Border.all(color: global.theme.primaryColor, width: 1.5),
            borderRadius: BorderRadius.circular(2),
          ),
        ),
      );
    }).toList();
  }

  Widget _buildElementContent(FormElement element) {
    switch (element.type) {
      case ElementType.text:
        return Padding(
          padding: const EdgeInsets.all(2),
          child: Text(
            element.text ?? '',
            style: element.textStyle,
            textAlign: element.textAlign,
            overflow: TextOverflow.clip,
          ),
        );
      case ElementType.dataField:
        final field =
            DataBindingRegistry.findField(element.dataBindingKey ?? '');
        final displayText = field != null
            ? '{${field.label}}'
            : '{${element.dataBindingKey ?? "field"}}';
        return Container(
          padding: const EdgeInsets.all(2),
          decoration: BoxDecoration(
            border: Border.all(
              color: global.theme.primaryColor.withValues(alpha: 0.3),
              width: 1,
            ),
            color: global.theme.primaryColor.withValues(alpha: 0.05),
          ),
          child: Text(
            displayText,
            style: element.textStyle.copyWith(
              color: global.theme.primaryColor,
            ),
            textAlign: element.textAlign,
          ),
        );
      case ElementType.image:
        if (element.imagePath != null) {
          return Image.file(File(element.imagePath!), fit: BoxFit.contain);
        }
        return Center(
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.image,
                  size: 24, color: global.theme.iconSecondaryColor),
              Text('Image',
                  style: TextStyle(
                      fontSize: 10, color: global.theme.textSecondaryColor)),
            ],
          ),
        );
      case ElementType.table:
        return _buildTable(element);
      case ElementType.rectangle:
        return const SizedBox.shrink();
      case ElementType.line:
      case ElementType.separator:
        return Center(
          child: Container(
            height: 1,
            color: element.borderColor ?? global.theme.textColor,
          ),
        );
      case ElementType.barcode:
        return Container(
          padding: const EdgeInsets.all(4),
          child: Column(
            mainAxisAlignment: MainAxisAlignment.center,
            children: [
              Icon(Icons.view_week, size: 20, color: global.theme.textColor),
              Text(
                element.barcodeValue ?? element.dataBindingKey ?? 'BARCODE',
                style: TextStyle(
                    fontSize: 8, color: global.theme.textSecondaryColor),
              ),
            ],
          ),
        );
      case ElementType.qrCode:
        return Center(
          child:
              Icon(Icons.qr_code, size: 40, color: global.theme.textColor),
        );
      case ElementType.pageNumber:
        return Text(
          element.text ?? 'Page 1/1',
          style: element.textStyle,
          textAlign: element.textAlign ?? TextAlign.center,
        );
      case ElementType.dateTime:
        return Text(
          element.text ?? '13/03/2569',
          style: element.textStyle,
          textAlign: element.textAlign ?? TextAlign.left,
        );
      case ElementType.signature:
        return Column(
          mainAxisAlignment: MainAxisAlignment.end,
          children: [
            Container(
              height: 1,
              color: element.borderColor ?? global.theme.textColor,
              margin: const EdgeInsets.symmetric(horizontal: 10),
            ),
            const SizedBox(height: 4),
            Text(
              element.text ?? '...................................',
              style: TextStyle(
                  fontSize: 10, color: global.theme.textSecondaryColor),
              textAlign: TextAlign.center,
            ),
          ],
        );
    }
  }

  Widget _buildTable(FormElement element) {
    if (element.tableData == null) return const SizedBox.shrink();
    final aligns = element.columnAligns;

    return Table(
      border: TableBorder.all(
        color: element.borderColor ?? global.theme.textColor,
        width: element.borderWidth ?? 1,
      ),
      children: element.tableData!.asMap().entries.map((entry) {
        final isHeader = element.tableHasHeader && entry.key == 0;
        return TableRow(
          decoration: isHeader
              ? BoxDecoration(color: global.theme.surfaceColor)
              : null,
          children: entry.value.asMap().entries.map((cellEntry) {
            final colIdx = cellEntry.key;
            final cell = cellEntry.value;
            final align = aligns != null && colIdx < aligns.length
                ? TextAlign.values[aligns[colIdx].clamp(0, 5)]
                : (isHeader ? TextAlign.center : TextAlign.left);
            return Padding(
              padding: const EdgeInsets.all(4),
              child: Text(
                cell,
                textAlign: align,
                style: TextStyle(
                  fontSize: 11,
                  fontWeight: isHeader ? FontWeight.bold : FontWeight.normal,
                  color: global.theme.textColor,
                ),
              ),
            );
          }).toList(),
        );
      }).toList(),
    );
  }
}

/// Grid overlay painter
class _GridPainter extends CustomPainter {
  final double gridSize;
  final double zoom;
  final Color color;

  _GridPainter({
    required this.gridSize,
    required this.zoom,
    required this.color,
  });

  @override
  void paint(Canvas canvas, Size size) {
    final paint = Paint()
      ..color = color
      ..strokeWidth = 0.5;

    final step = gridSize * zoom;
    if (step < 5) return;

    for (double x = 0; x < size.width; x += step) {
      canvas.drawLine(Offset(x, 0), Offset(x, size.height), paint);
    }
    for (double y = 0; y < size.height; y += step) {
      canvas.drawLine(Offset(0, y), Offset(size.width, y), paint);
    }
  }

  @override
  bool shouldRepaint(covariant _GridPainter oldDelegate) {
    return gridSize != oldDelegate.gridSize ||
        zoom != oldDelegate.zoom ||
        color != oldDelegate.color;
  }
}
