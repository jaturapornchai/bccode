import 'dart:convert';
import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:smlaicloud/global.dart' as global;
import 'package:smlaicloud/repositories/form_template_repository.dart';
import 'models/form_template.dart';
import 'models/form_element.dart';
import 'models/form_section.dart';
import 'models/paper_size.dart';
import 'models/editor_command.dart';
import 'widgets/form_preview.dart';
import 'widgets/full_form_editor.dart';
import 'widgets/toolbar/form_toolbar.dart';
import 'widgets/panels/element_palette_panel.dart';
import 'widgets/panels/layers_panel.dart';
import 'widgets/panels/properties_inspector.dart';
import 'widgets/status_bar.dart';
import 'widgets/template_gallery_dialog.dart';
import 'utils/form_template_factory.dart';

class FormDesignScreen extends StatefulWidget {
  const FormDesignScreen({super.key});

  @override
  State<FormDesignScreen> createState() => _FormDesignScreenState();
}

// docType options สำหรับเลือกประเภทเอกสาร
const List<Map<String, String>> _docTypeOptions = [
  {'value': 'tax_invoice', 'label': 'ใบกำกับภาษี'},
  {'value': 'receipt', 'label': 'ใบเสร็จรับเงิน'},
  {'value': 'cash_bill', 'label': 'บิลเงินสด'},
  {'value': 'quotation', 'label': 'ใบเสนอราคา'},
  {'value': 'purchase_order', 'label': 'ใบสั่งซื้อ'},
  {'value': 'delivery_note', 'label': 'ใบส่งของ'},
  {'value': 'invoice', 'label': 'ใบแจ้งหนี้'},
  {'value': 'credit_note', 'label': 'ใบลดหนี้'},
  {'value': 'debit_note', 'label': 'ใบเพิ่มหนี้'},
  {'value': 'payment_voucher', 'label': 'ใบสำคัญจ่าย'},
];

class _FormDesignScreenState extends State<FormDesignScreen>
    with global.ThemeRefreshMixin {
  late FormTemplate currentTemplate;
  final FormTemplateRepository _formTemplateRepo = FormTemplateRepository();
  String? _savedGuidFixed; // guidfixed จาก backend (null = ยังไม่เคย save)
  bool _isPreviewMode = false;
  double _zoomLevel = 1.0;
  bool _showRulers = true;
  bool _showGrid = true;
  bool _snapToGrid = true;
  final double _gridSize = 10.0;
  bool _showLeftPanel = true;
  bool _showRightPanel = true;

  // Page type (multi-page editing)
  PageType _currentPageType = PageType.first;

  // Selection
  FormElement? _selectedElement;
  SectionType? _selectedSection;
  final Set<String> _selectedElementIds = {};

  // Undo/Redo (max 200 steps)
  static const int _maxUndoSteps = 200;
  final List<EditorCommand> _undoStack = [];
  final List<EditorCommand> _redoStack = [];

  // Mouse position for status bar
  Offset? _mousePosition;

  // JSON dialog
  late TextEditingController _jsonController;

  @override
  void initState() {
    super.initState();
    currentTemplate = FormTemplate(
      name: global.language('form_design_new_form'),
      description: global.language('form_design_created_with_designer'),
    );
    _jsonController = TextEditingController(
      text: _formatJson(currentTemplate.toJson()),
    );
    // Preload all JSON templates into cache
    FormTemplateFactory.init();
  }

  @override
  void dispose() {
    _jsonController.dispose();
    super.dispose();
  }

  /// สร้าง template "view" ที่สลับ header/footer ตาม PageType ที่เลือก
  /// detail = auto fill พื้นที่เหลือ (paper height - header - footer - margins)
  FormTemplate get _viewTemplate {
    final header = currentTemplate.headerForPage(_currentPageType);
    final footer = currentTemplate.footerForPage(_currentPageType);
    final paperW = currentTemplate.effectiveWidth;
    final paperH = currentTemplate.effectiveHeight;

    // คำนวณ height จริงของ header/footer (resolve rows ถ้ามี)
    final headerH = header.hasRows ? header.resolveHeight(paperW) : header.height;
    final footerH = footer.hasRows ? footer.resolveHeight(paperW) : footer.height;
    final margins = currentTemplate.marginTop + currentTemplate.marginBottom;

    // detail ได้พื้นที่ที่เหลือ
    final detailH = (paperH - headerH - footerH - margins).clamp(50.0, double.infinity);
    final detail = currentTemplate.detail.copyWith(height: detailH);

    return currentTemplate.copyWith(
      header: header,
      detail: detail,
      footer: footer,
    );
  }

  /// เมื่อ editor คืน template ที่แก้แล้ว → เขียนกลับไป section ที่ถูกต้อง
  void _applyViewTemplateChange(FormTemplate viewChanged) {
    final Map<String, dynamic> updates = {};

    // header ที่ editor แก้ → ใส่กลับ section ตาม page type
    switch (_currentPageType) {
      case PageType.first:
        updates['header'] = viewChanged.header;
        break;
      case PageType.middle:
        updates['headerMiddle'] = viewChanged.header;
        break;
      case PageType.last:
        updates['headerLast'] = viewChanged.header;
        break;
    }

    // footer ที่ editor แก้ → ใส่กลับ section ตาม page type
    switch (_currentPageType) {
      case PageType.first:
        updates['footerFirst'] = viewChanged.footer;
        break;
      case PageType.middle:
        updates['footerMiddle'] = viewChanged.footer;
        break;
      case PageType.last:
        updates['footer'] = viewChanged.footer;
        break;
    }

    var result = currentTemplate.copyWith(detail: viewChanged.detail);

    if (updates.containsKey('header')) {
      result = result.copyWith(header: updates['header'] as FormSection);
    }
    if (updates.containsKey('headerMiddle')) {
      result =
          result.copyWith(headerMiddle: updates['headerMiddle'] as FormSection);
    }
    if (updates.containsKey('headerLast')) {
      result =
          result.copyWith(headerLast: updates['headerLast'] as FormSection);
    }
    if (updates.containsKey('footer')) {
      result = result.copyWith(footer: updates['footer'] as FormSection);
    }
    if (updates.containsKey('footerFirst')) {
      result =
          result.copyWith(footerFirst: updates['footerFirst'] as FormSection);
    }
    if (updates.containsKey('footerMiddle')) {
      result =
          result.copyWith(footerMiddle: updates['footerMiddle'] as FormSection);
    }

    setState(() => currentTemplate = result);
  }

  /// สร้าง section ว่างสำหรับ page type ที่ยังไม่มี
  void _initPageSections(PageType type) {
    if (type == PageType.first) return; // first page ใช้ header/footer เสมอ

    bool needUpdate = false;
    var t = currentTemplate;

    if (type == PageType.middle || type == PageType.last) {
      if (type == PageType.middle && t.headerMiddle == null) {
        t = t.copyWith(
          headerMiddle: FormSection(
            id: 'headerMiddle',
            type: SectionType.header,
            height: 60,
          ),
        );
        needUpdate = true;
      }
      if (type == PageType.last && t.headerLast == null) {
        t = t.copyWith(
          headerLast: FormSection(
            id: 'headerLast',
            type: SectionType.header,
            height: 60,
          ),
        );
        needUpdate = true;
      }
    }

    if (type == PageType.first && t.footerFirst == null) {
      t = t.copyWith(
        footerFirst: FormSection(
          id: 'footerFirst',
          type: SectionType.footer,
          height: 40,
        ),
      );
      needUpdate = true;
    }
    if (type == PageType.middle && t.footerMiddle == null) {
      t = t.copyWith(
        footerMiddle: FormSection(
          id: 'footerMiddle',
          type: SectionType.footer,
          height: 40,
        ),
      );
      needUpdate = true;
    }

    if (needUpdate) {
      _updateTemplate(t, description: 'Initialize ${type.label} sections');
    }
  }

  // ── Undo/Redo ──

  void _executeCommand(EditorCommand command) {
    setState(() {
      currentTemplate = command.execute(currentTemplate);
      currentTemplate =
          currentTemplate.copyWith(updatedAt: DateTime.now());
      _undoStack.add(command);
      _redoStack.clear();
      // Trim undo stack to prevent memory bloat
      if (_undoStack.length > _maxUndoSteps) {
        _undoStack.removeRange(0, _undoStack.length - _maxUndoSteps);
      }
    });
  }

  void _undo() {
    if (_undoStack.isEmpty) return;
    setState(() {
      final command = _undoStack.removeLast();
      currentTemplate = command.undo(currentTemplate);
      _redoStack.add(command);
    });
  }

  void _redo() {
    if (_redoStack.isEmpty) return;
    setState(() {
      final command = _redoStack.removeLast();
      currentTemplate = command.execute(currentTemplate);
      _undoStack.add(command);
    });
  }

  // ── Template Operations ──

  String _formatJson(Map<String, dynamic> json) {
    const encoder = JsonEncoder.withIndent('  ');
    return encoder.convert(json);
  }

  void _updateTemplate(FormTemplate template, {String description = 'Update template'}) {
    _executeCommand(ReplaceTemplateCommand(
      oldTemplate: currentTemplate,
      newTemplate: template,
      description: description,
    ));
  }

  void _onPaperSizeChanged(PaperSize newSize) {
    final oldSize = currentTemplate.paperSize;
    if (oldSize == newSize) return;

    final hasElements = currentTemplate.header.elements.isNotEmpty ||
        currentTemplate.detail.elements.isNotEmpty ||
        currentTemplate.footer.elements.isNotEmpty;

    if (!hasElements) {
      _updateTemplate(
        currentTemplate.copyWith(paperSize: newSize),
        description: 'Change paper size to ${newSize.name}',
      );
      return;
    }

    // ถามว่าจะ auto-fit หรือไม่
    showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: global.theme.dialogColor,
        title: Text('เปลี่ยนขนาดกระดาษ',
            style: TextStyle(color: global.theme.textColor)),
        content: Text(
          'ต้องการจัดขนาด/ตำแหน่ง elements ให้พอดีกับกระดาษ ${newSize.name} อัตโนมัติหรือไม่?',
          style: TextStyle(color: global.theme.textSecondaryColor),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('เปลี่ยนอย่างเดียว'),
          ),
          FilledButton.icon(
            onPressed: () => Navigator.pop(ctx, true),
            icon: const Icon(Icons.auto_fix_high, size: 18),
            label: const Text('AUTO จัดใหม่'),
          ),
        ],
      ),
    ).then((autoFit) {
      if (autoFit == null) return; // dismissed
      if (autoFit) {
        _autoFitToPaperSize(oldSize, newSize);
      } else {
        _updateTemplate(
          currentTemplate.copyWith(paperSize: newSize),
          description: 'Change paper size to ${newSize.name}',
        );
      }
    });
  }

  void _autoFitToPaperSize(PaperSize oldSize, PaperSize newSize) {
    final oldW = currentTemplate.effectiveWidth;
    final oldH = currentTemplate.effectiveHeight > 0
        ? currentTemplate.effectiveHeight
        : oldW * 1.5;
    final tmp = currentTemplate.copyWith(paperSize: newSize);
    final newW = tmp.effectiveWidth;
    final newH = tmp.effectiveHeight > 0 ? tmp.effectiveHeight : newW * 1.5;
    _autoFitWithScale(oldW, oldH, newW, newH,
        currentTemplate.copyWith(paperSize: newSize),
        'Auto-fit to ${newSize.name}');
  }

  void _onOrientationChanged(PageOrientation newOrientation) {
    final old = currentTemplate.orientation;
    if (old == newOrientation) return;

    final hasElements = currentTemplate.header.elements.isNotEmpty ||
        currentTemplate.detail.elements.isNotEmpty ||
        currentTemplate.footer.elements.isNotEmpty;

    if (!hasElements) {
      _updateTemplate(
        currentTemplate.copyWith(orientation: newOrientation),
        description: 'Change to ${newOrientation.label}',
      );
      return;
    }

    showDialog<bool>(
      context: context,
      builder: (ctx) => AlertDialog(
        backgroundColor: global.theme.dialogColor,
        title: Text('เปลี่ยนแนวกระดาษ',
            style: TextStyle(color: global.theme.textColor)),
        content: Text(
          'เปลี่ยนเป็น${newOrientation.label} ต้องการจัด elements ใหม่ให้พอดีหรือไม่?',
          style: TextStyle(color: global.theme.textSecondaryColor),
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.pop(ctx, false),
            child: const Text('เปลี่ยนอย่างเดียว'),
          ),
          FilledButton.icon(
            onPressed: () => Navigator.pop(ctx, true),
            icon: const Icon(Icons.auto_fix_high, size: 18),
            label: const Text('AUTO จัดใหม่'),
          ),
        ],
      ),
    ).then((autoFit) {
      if (autoFit == null) return;
      if (autoFit) {
        // Auto-fit: old effective → new effective
        final oldW = currentTemplate.effectiveWidth;
        final oldH = currentTemplate.effectiveHeight;
        final tmp = currentTemplate.copyWith(orientation: newOrientation);
        final newW = tmp.effectiveWidth;
        final newH = tmp.effectiveHeight;
        _autoFitWithScale(oldW, oldH, newW, newH,
            currentTemplate.copyWith(orientation: newOrientation),
            'Auto-fit ${newOrientation.label}');
      } else {
        _updateTemplate(
          currentTemplate.copyWith(orientation: newOrientation),
          description: 'Change to ${newOrientation.label}',
        );
      }
    });
  }

  /// จัดกลุ่ม elements เป็น "แถว" ตาม vertical overlap
  /// elements ที่ซ้อนกันในแนวตั้ง (Y range ตัดกัน) ถือว่าแถวเดียวกัน
  List<List<FormElement>> _groupIntoVisualRows(List<FormElement> elements) {
    if (elements.isEmpty) return [];
    final sorted = List<FormElement>.from(elements)
      ..sort((a, b) => a.y.compareTo(b.y));

    final rows = <List<FormElement>>[];
    final used = List.filled(sorted.length, false);

    for (int i = 0; i < sorted.length; i++) {
      if (used[i]) continue;

      final row = [sorted[i]];
      used[i] = true;
      double rowTop = sorted[i].y;
      double rowBottom = sorted[i].y + sorted[i].height;

      // หา elements ที่ซ้อนกัน (transitive)
      bool found = true;
      while (found) {
        found = false;
        for (int j = i + 1; j < sorted.length; j++) {
          if (used[j]) continue;
          final el = sorted[j];
          // ซ้อนกัน = Y range ตัดกัน
          if (el.y < rowBottom && (el.y + el.height) > rowTop) {
            row.add(el);
            used[j] = true;
            if (el.y + el.height > rowBottom) rowBottom = el.y + el.height;
            if (el.y < rowTop) rowTop = el.y;
            found = true;
          }
        }
      }
      rows.add(row);
    }
    return rows;
  }

  /// คำนวณความสูงที่ text element ต้องการจริงๆ ที่ width ใหม่
  /// ใช้ approximate: charWidth ≈ fontSize * factor → คำนวณจำนวนบรรทัด
  double _estimateTextHeight(FormElement el, double newWidth, double fontSize) {
    final text = el.text ?? el.dataBindingKey ?? '';
    if (text.isEmpty) return fontSize * 1.5 + 8;

    // Thai chars กว้างกว่า Latin เล็กน้อย — ใช้ค่าเฉลี่ย
    final charWidth = fontSize * 0.52;
    final charsPerLine = (newWidth / charWidth).floor().clamp(1, 9999);
    final numLines = (text.length / charsPerLine).ceil().clamp(1, 999);
    final lineHeight = fontSize * 1.45;
    return numLines * lineHeight + 8; // 8 = padding top+bottom
  }

  /// คำนวณ height ที่เหมาะสมสำหรับ element หลัง reflow
  double _computeElementHeight(
      FormElement el, double newWidth, double scaleX) {
    final fontSize = el.fontSize ?? 14.0;
    final isTextLike = el.type == ElementType.text ||
        el.type == ElementType.dataField ||
        el.type == ElementType.pageNumber ||
        el.type == ElementType.dateTime;

    if (isTextLike) {
      // คำนวณจากเนื้อหาจริง
      final estimated = _estimateTextHeight(el, newWidth, fontSize);
      // ใช้ค่าที่มากกว่าระหว่าง estimated กับ original
      // (ป้องกันกรณี element ถูก resize ให้ใหญ่ตั้งใจ)
      return estimated.clamp(fontSize * 1.5, double.infinity);
    }

    if (el.type == ElementType.table) {
      // table: scale height ตาม rows
      final rowCount = el.rows ?? 3;
      final rowH = fontSize * 1.6;
      final headerH = fontSize * 1.8;
      return headerH + rowCount * rowH + 8;
    }

    // Non-text elements: ใช้ height เดิม (image, line, rectangle, etc.)
    return el.height;
  }

  /// Re-layout elements: จัดเป็นแถว → scale X → คำนวณ height จากเนื้อหา → ไล่ Y จากบนลงล่าง
  ({List<FormElement> elements, double height}) _reflowSection(
      FormSection section, double scaleX, double newPaperW) {
    if (section.elements.isEmpty) {
      return (elements: <FormElement>[], height: 40.0);
    }

    final rows = _groupIntoVisualRows(section.elements);
    final List<FormElement> result = [];
    double cursorY = 4.0;
    const rowGap = 4.0;

    for (final row in rows) {
      // Pass 1: คำนวณ width + height ใหม่ของแต่ละ element
      final computed = <({FormElement el, double newW, double newX, double newH, double? newFontSize})>[];
      double maxH = 0;

      for (final el in row) {
        final newW = (el.width * scaleX).clamp(4.0, newPaperW - 4);
        final newX = (el.x * scaleX).clamp(0.0, newPaperW - newW);
        final newH = _computeElementHeight(el, newW, scaleX);

        // fontSize: ไม่ scale — เก็บค่าเดิมเสมอ (ผู้ใช้ตั้งเอง)
        double? newFontSize = el.fontSize;

        computed.add((el: el, newW: newW, newX: newX, newH: newH, newFontSize: newFontSize));
        if (newH > maxH) maxH = newH;
      }

      // Pass 2: วาง element ลง row
      for (final c in computed) {
        result.add(c.el.copyWith(
          x: c.newX,
          y: cursorY,
          width: c.newW,
          height: c.newH,
          fontSize: c.newFontSize,
          columnWidths: c.el.columnWidths?.map((w) => w * scaleX).toList(),
        ));
      }
      cursorY += maxH + rowGap;
    }

    return (elements: result, height: cursorY + 4.0);
  }

  /// Reflow section — ใช้ height จริงจากเนื้อหา
  FormSection _reflowToSection(
      FormSection section, double scaleX, double newW) {
    final r = _reflowSection(section, scaleX, newW);
    return section.copyWith(
      height: r.height.clamp(20.0, double.infinity),
      elements: r.elements,
    );
  }

  /// Reflow section สำหรับ fixed-layout เท่านั้น
  /// Row-based sections ไม่ต้อง reflow — resolve ใหม่อัตโนมัติตาม width ใหม่
  FormSection _smartReflow(FormSection section, double scaleX, double newW) {
    if (section.hasRows) return section; // flex layout = auto-adapt
    return _reflowToSection(section, scaleX, newW);
  }

  /// Auto-fit แบบ smart: Header(re-flow) → Footer(re-flow) → Detail(ที่เหลือ)
  /// Row-based sections: ไม่ต้อง reflow (flex layout ปรับ auto)
  /// Fixed sections: reflow elements ตามสัดส่วน
  void _autoFitWithScale(double oldW, double oldH, double newW, double newH,
      FormTemplate baseTemplate, String description) {
    final scaleX = newW / oldW;
    final totalH = newH > 0 ? newH : newW * 1.5;

    // ── Header ──
    final newHeader = _smartReflow(currentTemplate.header, scaleX, newW);
    final headerH = newHeader.hasRows
        ? newHeader.resolveHeight(newW)
        : newHeader.height;

    // ── Footer ──
    final newFooter = _smartReflow(currentTemplate.footer, scaleX, newW);
    final footerH = newFooter.hasRows
        ? newFooter.resolveHeight(newW)
        : newFooter.height;

    // ── Detail: พื้นที่ที่เหลือ ──
    final remainingH = (totalH - headerH - footerH).clamp(50.0, double.infinity);
    FormSection newDetail;
    if (currentTemplate.detail.hasRows) {
      newDetail = currentTemplate.detail; // flex → auto
    } else {
      final detailR = _reflowSection(currentTemplate.detail, scaleX, newW);
      final actualDetailH =
          detailR.height > remainingH ? detailR.height : remainingH;
      newDetail = currentTemplate.detail.copyWith(
        height: actualDetailH,
        elements: detailR.elements,
      );
    }

    // ── Multi-page sections ──
    FormSection? newHeaderMiddle;
    if (currentTemplate.headerMiddle != null) {
      newHeaderMiddle =
          _smartReflow(currentTemplate.headerMiddle!, scaleX, newW);
    }
    FormSection? newHeaderLast;
    if (currentTemplate.headerLast != null) {
      newHeaderLast =
          _smartReflow(currentTemplate.headerLast!, scaleX, newW);
    }
    FormSection? newFooterFirst;
    if (currentTemplate.footerFirst != null) {
      newFooterFirst =
          _smartReflow(currentTemplate.footerFirst!, scaleX, newW);
    }
    FormSection? newFooterMiddle;
    if (currentTemplate.footerMiddle != null) {
      newFooterMiddle =
          _smartReflow(currentTemplate.footerMiddle!, scaleX, newW);
    }

    final mScaleX = scaleX.clamp(0.1, 10.0);
    final newTemplate = baseTemplate.copyWith(
      header: newHeader,
      detail: newDetail,
      footer: newFooter,
      headerMiddle: newHeaderMiddle,
      headerLast: newHeaderLast,
      footerFirst: newFooterFirst,
      footerMiddle: newFooterMiddle,
      marginTop: currentTemplate.marginTop * mScaleX,
      marginBottom: currentTemplate.marginBottom * mScaleX,
      marginLeft: currentTemplate.marginLeft * mScaleX,
      marginRight: currentTemplate.marginRight * mScaleX,
    );

    _updateTemplate(newTemplate, description: description);
  }

  void _onZoomChanged(double zoom) {
    if (zoom == -1.0 || zoom == -2.0) {
      // Fit page or Fit width — need layout constraints
      // Use post-frame callback to get actual canvas size
      WidgetsBinding.instance.addPostFrameCallback((_) {
        final renderBox = context.findRenderObject() as RenderBox?;
        if (renderBox == null) return;
        final canvasWidth = renderBox.size.width - 300; // approx panels
        final canvasHeight = renderBox.size.height - 120; // toolbar+status
        final paperW = currentTemplate.effectiveWidth;
        final paperH = currentTemplate.effectiveHeight > 0
            ? currentTemplate.totalHeight
            : paperW * 1.5;

        double newZoom;
        if (zoom == -1.0) {
          // Fit page: fit both width and height
          final zx = (canvasWidth - 60) / paperW;
          final zy = (canvasHeight - 60) / paperH;
          newZoom = zx < zy ? zx : zy;
        } else {
          // Fit width
          newZoom = (canvasWidth - 60) / paperW;
        }
        setState(() => _zoomLevel = newZoom.clamp(0.1, 5.0));
      });
    } else {
      setState(() => _zoomLevel = zoom);
    }
  }

  Future<void> _saveTemplate() async {
    await _showSaveDialog();
  }

  Future<void> _loadTemplate() async {
    await _showLoadDialog();
  }

  // ── Save to Backend ──

  Future<void> _showSaveDialog() async {
    final codeController = TextEditingController(
      text: currentTemplate.documentType ?? '',
    );
    final nameController = TextEditingController(text: currentTemplate.name);
    String selectedDocType = currentTemplate.documentType ?? 'tax_invoice';

    final confirmed = await showDialog<bool>(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) => AlertDialog(
          backgroundColor: global.theme.dialogColor,
          title: Text('บันทึกฟอร์มไป Server',
              style: TextStyle(color: global.theme.textColor)),
          content: SizedBox(
            width: 400,
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                // ชื่อฟอร์ม
                TextField(
                  controller: nameController,
                  style: TextStyle(color: global.theme.textColor),
                  decoration: InputDecoration(
                    labelText: 'ชื่อฟอร์ม',
                    labelStyle:
                        TextStyle(color: global.theme.textSecondaryColor),
                    border: const OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 12),
                // รหัสฟอร์ม (code)
                TextField(
                  controller: codeController,
                  style: TextStyle(color: global.theme.textColor),
                  decoration: InputDecoration(
                    labelText: 'รหัสฟอร์ม (code)',
                    hintText: 'เช่น po_standard, inv_a4',
                    labelStyle:
                        TextStyle(color: global.theme.textSecondaryColor),
                    hintStyle:
                        TextStyle(color: global.theme.textSecondaryColor.withValues(alpha: 0.5)),
                    border: const OutlineInputBorder(),
                  ),
                ),
                const SizedBox(height: 12),
                // ประเภทเอกสาร
                DropdownButtonFormField<String>(
                  initialValue: selectedDocType,
                  dropdownColor: global.theme.dialogColor,
                  style: TextStyle(color: global.theme.textColor),
                  decoration: InputDecoration(
                    labelText: 'ประเภทเอกสาร',
                    labelStyle:
                        TextStyle(color: global.theme.textSecondaryColor),
                    border: const OutlineInputBorder(),
                  ),
                  items: _docTypeOptions
                      .map((opt) => DropdownMenuItem(
                            value: opt['value'],
                            child: Text(opt['label']!),
                          ))
                      .toList(),
                  onChanged: (v) {
                    if (v != null) {
                      setDialogState(() => selectedDocType = v);
                      if (codeController.text.isEmpty) {
                        codeController.text = v;
                      }
                    }
                  },
                ),
              ],
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(ctx, false),
              child: Text('ยกเลิก',
                  style: TextStyle(color: global.theme.textSecondaryColor)),
            ),
            FilledButton.icon(
              onPressed: () => Navigator.pop(ctx, true),
              icon: const Icon(Icons.cloud_upload),
              label: const Text('บันทึก'),
            ),
          ],
        ),
      ),
    );

    if (confirmed != true || !mounted) return;

    final code = codeController.text.trim();
    final name = nameController.text.trim();
    if (code.isEmpty || name.isEmpty) {
      global.showErrorSnackBar(context, 'กรุณากรอก ชื่อ และ รหัสฟอร์ม');
      return;
    }

    // Update template metadata
    setState(() {
      currentTemplate = currentTemplate.copyWith(
        name: name,
        documentType: selectedDocType,
      );
    });

    // Build API payload
    final payload = {
      'code': code,
      'doctype': selectedDocType,
      'names': [
        {'code': 'th', 'name': name},
      ],
      'isdefault': false,
      'templatedata': currentTemplate.toJson(),
    };

    try {
      final response = await _formTemplateRepo.saveFormTemplate(payload);
      if (response.success) {
        _savedGuidFixed = response.id;
        if (mounted) {
          global.showSuccessSnackBar(context, 'บันทึกฟอร์มสำเร็จ');
        }
      } else {
        if (mounted) {
          global.showErrorSnackBar(context, 'บันทึกล้มเหลว: ${response.message}');
        }
      }
    } catch (e) {
      if (mounted) {
        global.showErrorSnackBar(context, 'บันทึกล้มเหลว: $e');
      }
    }
  }

  // ── Load from Backend ──

  Future<void> _showLoadDialog() async {
    // โหลดรายการ template จาก backend
    List<dynamic> templates = [];
    bool isLoading = true;
    String filterDocType = '';

    await showDialog(
      context: context,
      builder: (ctx) => StatefulBuilder(
        builder: (ctx, setDialogState) {
          // โหลดข้อมูลครั้งแรก
          if (isLoading) {
            _formTemplateRepo
                .getFormTemplateList(limit: 100, doctype: filterDocType)
                .then((response) {
              if (response.success && response.data != null) {
                setDialogState(() {
                  templates = response.data is List ? response.data : [];
                  isLoading = false;
                });
              } else {
                setDialogState(() {
                  templates = [];
                  isLoading = false;
                });
              }
            }).catchError((e) {
              setDialogState(() {
                templates = [];
                isLoading = false;
              });
            });
          }

          return AlertDialog(
            backgroundColor: global.theme.dialogColor,
            title: Row(
              children: [
                Icon(Icons.cloud_download, color: global.theme.textColor),
                const SizedBox(width: 8),
                Text('โหลดฟอร์มจาก Server',
                    style: TextStyle(color: global.theme.textColor)),
              ],
            ),
            content: SizedBox(
              width: 500,
              height: 400,
              child: Column(
                children: [
                  // Filter by doctype
                  DropdownButtonFormField<String>(
                    initialValue: filterDocType.isEmpty ? null : filterDocType,
                    dropdownColor: global.theme.dialogColor,
                    style: TextStyle(color: global.theme.textColor),
                    decoration: InputDecoration(
                      labelText: 'กรองตามประเภทเอกสาร',
                      labelStyle:
                          TextStyle(color: global.theme.textSecondaryColor),
                      border: const OutlineInputBorder(),
                      isDense: true,
                    ),
                    items: [
                      DropdownMenuItem<String>(
                        value: '',
                        child: Text('ทั้งหมด',
                            style:
                                TextStyle(color: global.theme.textColor)),
                      ),
                      ..._docTypeOptions.map((opt) => DropdownMenuItem(
                            value: opt['value'],
                            child: Text(opt['label']!),
                          )),
                    ],
                    onChanged: (v) {
                      setDialogState(() {
                        filterDocType = v ?? '';
                        isLoading = true;
                      });
                    },
                  ),
                  const SizedBox(height: 12),
                  // Template list
                  Expanded(
                    child: isLoading
                        ? const Center(child: CircularProgressIndicator())
                        : templates.isEmpty
                            ? Center(
                                child: Text('ไม่พบฟอร์ม',
                                    style: TextStyle(
                                        color: global
                                            .theme.textSecondaryColor)))
                            : ListView.builder(
                                itemCount: templates.length,
                                itemBuilder: (ctx, index) {
                                  final t = templates[index];
                                  final names = t['names'] as List? ?? [];
                                  final displayName = names.isNotEmpty
                                      ? (names[0]['name'] ?? 'ไม่มีชื่อ')
                                      : 'ไม่มีชื่อ';
                                  final code = t['code'] ?? '';
                                  final docType = t['doctype'] ?? '';
                                  final docTypeLabel = _docTypeOptions
                                      .firstWhere(
                                        (o) => o['value'] == docType,
                                        orElse: () => {'label': docType},
                                      )['label']!;

                                  return Card(
                                    color: global.theme.cardColor,
                                    child: ListTile(
                                      leading: Icon(Icons.description,
                                          color: global.theme.primaryColor),
                                      title: Text(displayName,
                                          style: TextStyle(
                                              color:
                                                  global.theme.textColor)),
                                      subtitle: Text(
                                          '$code • $docTypeLabel',
                                          style: TextStyle(
                                              color: global.theme
                                                  .textSecondaryColor,
                                              fontSize: 12)),
                                      trailing: Icon(
                                          Icons.arrow_forward_ios,
                                          size: 16,
                                          color: global
                                              .theme.textSecondaryColor),
                                      onTap: () =>
                                          Navigator.pop(ctx, t),
                                    ),
                                  );
                                },
                              ),
                  ),
                ],
              ),
            ),
            actions: [
              TextButton(
                onPressed: () => Navigator.pop(ctx, null),
                child: Text('ยกเลิก',
                    style:
                        TextStyle(color: global.theme.textSecondaryColor)),
              ),
            ],
          );
        },
      ),
    ).then((selected) async {
      if (selected == null || selected is! Map) return;

      try {
        final guidFixed = selected['guidfixed'] as String?;
        if (guidFixed == null) return;

        // โหลด template เต็มจาก API
        final response = await _formTemplateRepo.getFormTemplate(guidFixed);
        if (!response.success || response.data == null) {
          if (mounted) {
            global.showErrorSnackBar(context, 'โหลดล้มเหลว: ${response.message}');
          }
          return;
        }

        final data = response.data as Map<String, dynamic>;
        final templateData = data['templatedata'] as Map<String, dynamic>?;
        if (templateData == null) {
          if (mounted) {
            global.showErrorSnackBar(context, 'ฟอร์มไม่มีข้อมูล templatedata');
          }
          return;
        }

        final template = FormTemplate.fromJson(templateData);
        setState(() {
          currentTemplate = template;
          _savedGuidFixed = guidFixed;
          _undoStack.clear();
          _redoStack.clear();
          _selectedElement = null;
          _selectedSection = null;
          _selectedElementIds.clear();
        });
        if (mounted) {
          global.showSuccessSnackBar(context, 'โหลดฟอร์มสำเร็จ');
        }
      } catch (e) {
        if (mounted) {
          global.showErrorSnackBar(context, 'โหลดล้มเหลว: $e');
        }
      }
    });
  }

  // ── Element Operations ──

  void _addElement(ElementType type, SectionType section, {Offset? position}) {
    final element = FormElement(
      id: DateTime.now().millisecondsSinceEpoch.toString(),
      type: type,
      x: position?.dx ?? 50,
      y: position?.dy ?? 50,
      width: _defaultWidth(type),
      height: _defaultHeight(type),
      text: _defaultText(type),
      fontSize: 14,
      fontFamily: 'Sarabun',
      textColor: global.theme.textColor,
      borderColor: type == ElementType.rectangle ||
              type == ElementType.table ||
              type == ElementType.signature
          ? global.theme.dividerBorderColor
          : null,
      borderWidth: type == ElementType.rectangle ||
              type == ElementType.table ||
              type == ElementType.signature
          ? 1
          : null,
      rows: type == ElementType.table ? 5 : null,
      columns: type == ElementType.table ? 5 : null,
      tableData: type == ElementType.table
          ? List.generate(5, (r) => List.generate(5, (c) => r == 0 ? 'Header ${c + 1}' : ''))
          : null,
      tableHasHeader: type == ElementType.table,
      lineStyle: type == ElementType.separator || type == ElementType.line
          ? 'solid'
          : null,
      dataBindingKey: type == ElementType.dataField ? 'doc_number' : null,
    );

    _executeCommand(AddElementCommand(sectionType: section, element: element));
    setState(() {
      _selectedElement = element;
      _selectedSection = section;
    });
  }

  double _defaultWidth(ElementType type) {
    switch (type) {
      case ElementType.text:
      case ElementType.dataField:
        return 200;
      case ElementType.image:
        return 150;
      case ElementType.line:
      case ElementType.separator:
        return 400;
      case ElementType.rectangle:
        return 200;
      case ElementType.table:
        return 500;
      case ElementType.barcode:
        return 200;
      case ElementType.qrCode:
        return 100;
      case ElementType.pageNumber:
      case ElementType.dateTime:
        return 120;
      case ElementType.signature:
        return 200;
    }
  }

  double _defaultHeight(ElementType type) {
    switch (type) {
      case ElementType.text:
      case ElementType.dataField:
      case ElementType.pageNumber:
      case ElementType.dateTime:
        return 30;
      case ElementType.image:
        return 100;
      case ElementType.line:
      case ElementType.separator:
        return 2;
      case ElementType.rectangle:
        return 100;
      case ElementType.table:
        return 200;
      case ElementType.barcode:
        return 60;
      case ElementType.qrCode:
        return 100;
      case ElementType.signature:
        return 60;
    }
  }

  String? _defaultText(ElementType type) {
    switch (type) {
      case ElementType.text:
        return 'Text';
      case ElementType.dataField:
        return null;
      case ElementType.pageNumber:
        return 'Page {page_number}/{page_count}';
      case ElementType.dateTime:
        return '{doc_date}';
      case ElementType.signature:
        return '..................................';
      default:
        return null;
    }
  }

  void _deleteSelectedElement() {
    if (_selectedSection == null) return;
    // Delete all selected elements as a batch
    if (_selectedElementIds.length > 1) {
      final section = _getSection(_selectedSection!);
      final commands = <EditorCommand>[];
      for (final el in section.elements) {
        if (_selectedElementIds.contains(el.id)) {
          commands.add(DeleteElementCommand(sectionType: _selectedSection!, element: el));
        }
      }
      if (commands.isNotEmpty) {
        _executeCommand(BatchCommand(commands: commands, description: 'Delete ${commands.length} elements'));
      }
    } else if (_selectedElement != null) {
      _executeCommand(DeleteElementCommand(
        sectionType: _selectedSection!,
        element: _selectedElement!,
      ));
    }
    setState(() {
      _selectedElement = null;
      _selectedSection = null;
      _selectedElementIds.clear();
    });
  }

  void _updateSelectedElement(FormElement updated) {
    if (_selectedElement == null || _selectedSection == null) return;
    _executeCommand(UpdatePropertyCommand(
      sectionType: _selectedSection!,
      oldElement: _selectedElement!,
      newElement: updated,
    ));
    setState(() {
      _selectedElement = updated;
    });
  }

  // ── Helpers for multi-select ──

  SectionType? _findElementSection(String elementId) {
    for (final type in SectionType.values) {
      final section = _getSection(type);
      if (section.elements.any((e) => e.id == elementId)) return type;
    }
    return null;
  }

  FormElement? _findElement(String elementId, SectionType section) {
    final elements = _getSection(section).elements;
    for (final e in elements) {
      if (e.id == elementId) return e;
    }
    return null;
  }

  /// Get all selected elements with their sections (only from _selectedSection for alignment)
  List<FormElement> _getSelectedElements() {
    if (_selectedSection == null) return [];
    final section = _getSection(_selectedSection!);
    return section.elements
        .where((e) => _selectedElementIds.contains(e.id))
        .toList();
  }

  // ── Alignment ──

  void _alignLeft() {
    final elements = _getSelectedElements();
    if (elements.length < 2 || _selectedSection == null) return;
    final minX = elements.map((e) => e.x).reduce((a, b) => a < b ? a : b);
    final commands = <EditorCommand>[];
    for (final e in elements) {
      if (e.x != minX) {
        commands.add(MoveElementCommand(
          sectionType: _selectedSection!, elementId: e.id,
          oldX: e.x, oldY: e.y, newX: minX, newY: e.y,
        ));
      }
    }
    if (commands.isNotEmpty) {
      _executeCommand(BatchCommand(commands: commands, description: 'Align left'));
    }
  }

  void _alignCenter() {
    final elements = _getSelectedElements();
    if (elements.length < 2 || _selectedSection == null) return;
    final minX = elements.map((e) => e.x).reduce((a, b) => a < b ? a : b);
    final maxRight = elements.map((e) => e.x + e.width).reduce((a, b) => a > b ? a : b);
    final centerX = (minX + maxRight) / 2;
    final commands = <EditorCommand>[];
    for (final e in elements) {
      final newX = centerX - e.width / 2;
      if ((newX - e.x).abs() > 0.5) {
        commands.add(MoveElementCommand(
          sectionType: _selectedSection!, elementId: e.id,
          oldX: e.x, oldY: e.y, newX: newX, newY: e.y,
        ));
      }
    }
    if (commands.isNotEmpty) {
      _executeCommand(BatchCommand(commands: commands, description: 'Align center'));
    }
  }

  void _alignRight() {
    final elements = _getSelectedElements();
    if (elements.length < 2 || _selectedSection == null) return;
    final maxRight = elements.map((e) => e.x + e.width).reduce((a, b) => a > b ? a : b);
    final commands = <EditorCommand>[];
    for (final e in elements) {
      final newX = maxRight - e.width;
      if ((newX - e.x).abs() > 0.5) {
        commands.add(MoveElementCommand(
          sectionType: _selectedSection!, elementId: e.id,
          oldX: e.x, oldY: e.y, newX: newX, newY: e.y,
        ));
      }
    }
    if (commands.isNotEmpty) {
      _executeCommand(BatchCommand(commands: commands, description: 'Align right'));
    }
  }

  void _alignTop() {
    final elements = _getSelectedElements();
    if (elements.length < 2 || _selectedSection == null) return;
    final minY = elements.map((e) => e.y).reduce((a, b) => a < b ? a : b);
    final commands = <EditorCommand>[];
    for (final e in elements) {
      if (e.y != minY) {
        commands.add(MoveElementCommand(
          sectionType: _selectedSection!, elementId: e.id,
          oldX: e.x, oldY: e.y, newX: e.x, newY: minY,
        ));
      }
    }
    if (commands.isNotEmpty) {
      _executeCommand(BatchCommand(commands: commands, description: 'Align top'));
    }
  }

  void _alignMiddle() {
    final elements = _getSelectedElements();
    if (elements.length < 2 || _selectedSection == null) return;
    final minY = elements.map((e) => e.y).reduce((a, b) => a < b ? a : b);
    final maxBottom = elements.map((e) => e.y + e.height).reduce((a, b) => a > b ? a : b);
    final centerY = (minY + maxBottom) / 2;
    final commands = <EditorCommand>[];
    for (final e in elements) {
      final newY = centerY - e.height / 2;
      if ((newY - e.y).abs() > 0.5) {
        commands.add(MoveElementCommand(
          sectionType: _selectedSection!, elementId: e.id,
          oldX: e.x, oldY: e.y, newX: e.x, newY: newY,
        ));
      }
    }
    if (commands.isNotEmpty) {
      _executeCommand(BatchCommand(commands: commands, description: 'Align middle'));
    }
  }

  void _alignBottom() {
    final elements = _getSelectedElements();
    if (elements.length < 2 || _selectedSection == null) return;
    final maxBottom = elements.map((e) => e.y + e.height).reduce((a, b) => a > b ? a : b);
    final commands = <EditorCommand>[];
    for (final e in elements) {
      final newY = maxBottom - e.height;
      if ((newY - e.y).abs() > 0.5) {
        commands.add(MoveElementCommand(
          sectionType: _selectedSection!, elementId: e.id,
          oldX: e.x, oldY: e.y, newX: e.x, newY: newY,
        ));
      }
    }
    if (commands.isNotEmpty) {
      _executeCommand(BatchCommand(commands: commands, description: 'Align bottom'));
    }
  }

  void _distributeH() {
    final elements = _getSelectedElements();
    if (elements.length < 3 || _selectedSection == null) return;
    final sorted = [...elements]..sort((a, b) => a.x.compareTo(b.x));
    final first = sorted.first;
    final last = sorted.last;
    final totalSpace = (last.x + last.width) - first.x;
    final totalWidths = sorted.fold<double>(0, (sum, e) => sum + e.width);
    final gap = (totalSpace - totalWidths) / (sorted.length - 1);
    final commands = <EditorCommand>[];
    double currentX = first.x;
    for (final e in sorted) {
      if ((currentX - e.x).abs() > 0.5) {
        commands.add(MoveElementCommand(
          sectionType: _selectedSection!, elementId: e.id,
          oldX: e.x, oldY: e.y, newX: currentX, newY: e.y,
        ));
      }
      currentX += e.width + gap;
    }
    if (commands.isNotEmpty) {
      _executeCommand(BatchCommand(commands: commands, description: 'Distribute horizontal'));
    }
  }

  void _distributeV() {
    final elements = _getSelectedElements();
    if (elements.length < 3 || _selectedSection == null) return;
    final sorted = [...elements]..sort((a, b) => a.y.compareTo(b.y));
    final first = sorted.first;
    final last = sorted.last;
    final totalSpace = (last.y + last.height) - first.y;
    final totalHeights = sorted.fold<double>(0, (sum, e) => sum + e.height);
    final gap = (totalSpace - totalHeights) / (sorted.length - 1);
    final commands = <EditorCommand>[];
    double currentY = first.y;
    for (final e in sorted) {
      if ((currentY - e.y).abs() > 0.5) {
        commands.add(MoveElementCommand(
          sectionType: _selectedSection!, elementId: e.id,
          oldX: e.x, oldY: e.y, newX: e.x, newY: currentY,
        ));
      }
      currentY += e.height + gap;
    }
    if (commands.isNotEmpty) {
      _executeCommand(BatchCommand(commands: commands, description: 'Distribute vertical'));
    }
  }

  void _bringToFront() {
    if (_selectedElement == null || _selectedSection == null) return;
    final section = _getSection(_selectedSection!);
    final idx = section.elements.indexWhere((e) => e.id == _selectedElement!.id);
    if (idx < 0 || idx >= section.elements.length - 1) return;
    _executeCommand(ReorderElementCommand(
      sectionType: _selectedSection!,
      elementId: _selectedElement!.id,
      oldIndex: idx,
      newIndex: section.elements.length - 1,
    ));
  }

  void _sendToBack() {
    if (_selectedElement == null || _selectedSection == null) return;
    final section = _getSection(_selectedSection!);
    final idx = section.elements.indexWhere((e) => e.id == _selectedElement!.id);
    if (idx <= 0) return;
    _executeCommand(ReorderElementCommand(
      sectionType: _selectedSection!,
      elementId: _selectedElement!.id,
      oldIndex: idx,
      newIndex: 0,
    ));
  }

  FormSection _getSection(SectionType type) {
    switch (type) {
      case SectionType.header:
        return currentTemplate.header;
      case SectionType.detail:
        return currentTemplate.detail;
      case SectionType.footer:
        return currentTemplate.footer;
    }
  }

  // ── Keyboard Shortcuts ──

  KeyEventResult _handleKeyEvent(FocusNode node, KeyEvent event) {
    if (event is! KeyDownEvent) return KeyEventResult.ignored;
    final isCtrl = HardwareKeyboard.instance.isControlPressed ||
        HardwareKeyboard.instance.isMetaPressed;

    if (event.logicalKey == LogicalKeyboardKey.delete ||
        event.logicalKey == LogicalKeyboardKey.backspace) {
      if (_selectedElementIds.isNotEmpty) {
        _deleteSelectedElement();
        return KeyEventResult.handled;
      }
    }
    if (isCtrl && event.logicalKey == LogicalKeyboardKey.keyZ) {
      _undo();
      return KeyEventResult.handled;
    }
    if (isCtrl && event.logicalKey == LogicalKeyboardKey.keyY) {
      _redo();
      return KeyEventResult.handled;
    }
    if (isCtrl && event.logicalKey == LogicalKeyboardKey.keyD) {
      if (_selectedElement != null && _selectedSection != null) {
        final clone = _selectedElement!.copyWith(
          id: DateTime.now().millisecondsSinceEpoch.toString(),
          x: _selectedElement!.x + 20,
          y: _selectedElement!.y + 20,
        );
        _executeCommand(
            AddElementCommand(sectionType: _selectedSection!, element: clone));
        setState(() {
          _selectedElement = clone;
        });
        return KeyEventResult.handled;
      }
    }
    // Arrow keys: nudge
    if (_selectedElement != null && _selectedSection != null) {
      final step = isCtrl ? _gridSize : 1.0;
      double dx = 0, dy = 0;
      if (event.logicalKey == LogicalKeyboardKey.arrowLeft) dx = -step;
      if (event.logicalKey == LogicalKeyboardKey.arrowRight) dx = step;
      if (event.logicalKey == LogicalKeyboardKey.arrowUp) dy = -step;
      if (event.logicalKey == LogicalKeyboardKey.arrowDown) dy = step;
      if (dx != 0 || dy != 0) {
        final updated = _selectedElement!.copyWith(
          x: (_selectedElement!.x + dx).clamp(0, double.infinity),
          y: (_selectedElement!.y + dy).clamp(0, double.infinity),
        );
        _executeCommand(UpdatePropertyCommand(
          sectionType: _selectedSection!,
          oldElement: _selectedElement!,
          newElement: updated,
        ));
        setState(() {
          _selectedElement = updated;
        });
        return KeyEventResult.handled;
      }
    }
    return KeyEventResult.ignored;
  }

  // ── Template Gallery ──

  void _showTemplateGallery() {
    showDialog(
      context: context,
      builder: (context) => TemplateGalleryDialog(
        onTemplateSelected: (template) {
          setState(() {
            currentTemplate = template;
            _undoStack.clear();
            _redoStack.clear();
            _selectedElement = null;
            _selectedSection = null;
            _selectedElementIds.clear();
          });
        },
      ),
    );
  }



  // ── Dialogs ──

  void _showFormInfoDialog() {
    final nameController = TextEditingController(text: currentTemplate.name);
    final descController =
        TextEditingController(text: currentTemplate.description);
    PaperSize selectedSize = currentTemplate.paperSize;
    PrintMode selectedPrintMode = currentTemplate.printMode;
    bool multiPage = currentTemplate.multiPageEnabled;
    int rowsPerPage = currentTemplate.detailRowsPerPage;

    showDialog(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setDialogState) => AlertDialog(
          backgroundColor: global.theme.dialogColor,
          title: Text(global.language('form_design_form_info'),
              style: TextStyle(color: global.theme.textColor)),
          content: SingleChildScrollView(
            child: Column(
              mainAxisSize: MainAxisSize.min,
              children: [
                TextField(
                  controller: nameController,
                  style: TextStyle(color: global.theme.textColor),
                  decoration: InputDecoration(
                    labelText: global.language('form_design_form_name'),
                    labelStyle:
                        TextStyle(color: global.theme.textSecondaryColor),
                  ),
                ),
                const SizedBox(height: 12),
                TextField(
                  controller: descController,
                  style: TextStyle(color: global.theme.textColor),
                  decoration: InputDecoration(
                    labelText: global.language('form_design_description'),
                    labelStyle:
                        TextStyle(color: global.theme.textSecondaryColor),
                  ),
                  maxLines: 2,
                ),
                const SizedBox(height: 12),
                DropdownButtonFormField<PaperSize>(
                  initialValue: selectedSize,
                  dropdownColor: global.theme.dialogColor,
                  style: TextStyle(color: global.theme.textColor),
                  decoration: InputDecoration(
                    labelText: global.language('form_design_paper_size'),
                    labelStyle:
                        TextStyle(color: global.theme.textSecondaryColor),
                  ),
                  items: PaperSize.values.map((size) {
                    return DropdownMenuItem(
                      value: size,
                      child: Text('${size.name} (${size.description})'),
                    );
                  }).toList(),
                  onChanged: (value) {
                    if (value != null) {
                      setDialogState(() => selectedSize = value);
                    }
                  },
                ),
                const SizedBox(height: 12),
                DropdownButtonFormField<PrintMode>(
                  initialValue: selectedPrintMode,
                  dropdownColor: global.theme.dialogColor,
                  style: TextStyle(color: global.theme.textColor),
                  decoration: InputDecoration(
                    labelText: global.language('fd_print_mode'),
                    labelStyle:
                        TextStyle(color: global.theme.textSecondaryColor),
                  ),
                  items: PrintMode.values.map((mode) {
                    return DropdownMenuItem(
                      value: mode,
                      child: Text(mode.label),
                    );
                  }).toList(),
                  onChanged: (value) {
                    if (value != null) {
                      setDialogState(() => selectedPrintMode = value);
                    }
                  },
                ),
                const SizedBox(height: 12),
                SwitchListTile(
                  title: Text(global.language('fd_multi_page'),
                      style: TextStyle(color: global.theme.textColor)),
                  subtitle: Text(global.language('fd_multi_page_desc'),
                      style:
                          TextStyle(color: global.theme.textSecondaryColor)),
                  value: multiPage,
                  onChanged: (v) => setDialogState(() => multiPage = v),
                ),
                if (multiPage)
                  TextField(
                    controller:
                        TextEditingController(text: rowsPerPage.toString()),
                    style: TextStyle(color: global.theme.textColor),
                    keyboardType: TextInputType.number,
                    decoration: InputDecoration(
                      labelText: global.language('fd_rows_per_page'),
                      labelStyle:
                          TextStyle(color: global.theme.textSecondaryColor),
                    ),
                    onChanged: (v) {
                      final n = int.tryParse(v);
                      if (n != null && n > 0) rowsPerPage = n;
                    },
                  ),
              ],
            ),
          ),
          actions: [
            TextButton(
              onPressed: () => Navigator.pop(context),
              child: Text(global.language('cancel')),
            ),
            ElevatedButton(
              onPressed: () {
                _updateTemplate(currentTemplate.copyWith(
                  name: nameController.text,
                  description: descController.text,
                  paperSize: selectedSize,
                  printMode: selectedPrintMode,
                  multiPageEnabled: multiPage,
                  detailRowsPerPage: rowsPerPage,
                ));
                Navigator.pop(context);
              },
              style: ElevatedButton.styleFrom(
                backgroundColor: global.theme.primaryColor,
                foregroundColor: global.theme.onPrimaryColor,
              ),
              child: Text(global.language('save')),
            ),
          ],
        ),
      ),
    );
  }

  void _showJsonDialog() {
    _jsonController.text = _formatJson(currentTemplate.toJson());
    String jsonError = '';

    showDialog(
      context: context,
      builder: (context) => StatefulBuilder(
        builder: (context, setDialogState) => Dialog(
          backgroundColor: global.theme.dialogColor,
          child: SizedBox(
            width: 600,
            height: 500,
            child: Column(
              children: [
                Container(
                  padding: const EdgeInsets.all(12),
                  color: global.theme.surfaceColor,
                  child: Row(
                    children: [
                      Icon(Icons.code,
                          size: 20, color: global.theme.iconColor),
                      const SizedBox(width: 8),
                      Text(global.language('fd_json_editor'),
                          style: TextStyle(
                              fontWeight: FontWeight.bold,
                              color: global.theme.textColor)),
                      const Spacer(),
                      IconButton(
                        icon: Icon(Icons.close,
                            size: 20, color: global.theme.iconColor),
                        onPressed: () => Navigator.pop(context),
                      ),
                    ],
                  ),
                ),
                if (jsonError.isNotEmpty)
                  Container(
                    padding: const EdgeInsets.all(8),
                    color: global.theme.negativeHighlightColor,
                    child: Text(jsonError,
                        style: TextStyle(
                            color: global.theme.negativeHighlightTextColor,
                            fontSize: 12)),
                  ),
                Expanded(
                  child: Padding(
                    padding: const EdgeInsets.all(8),
                    child: TextField(
                      controller: _jsonController,
                      maxLines: null,
                      expands: true,
                      style: TextStyle(
                          fontFamily: 'Courier',
                          fontSize: 12,
                          color: global.theme.textColor),
                      decoration: InputDecoration(
                        border: OutlineInputBorder(
                          borderSide: BorderSide(
                              color: global.theme.dividerBorderColor),
                        ),
                        fillColor: global.theme.cardColor,
                        filled: true,
                      ),
                    ),
                  ),
                ),
                Container(
                  padding: const EdgeInsets.all(8),
                  color: global.theme.surfaceColor,
                  child: Row(
                    mainAxisAlignment: MainAxisAlignment.end,
                    children: [
                      ElevatedButton.icon(
                        onPressed: () {
                          try {
                            final jsonData =
                                jsonDecode(_jsonController.text);
                            final template =
                                FormTemplate.fromJson(jsonData);
                            _updateTemplate(template);
                            Navigator.pop(context);
                          } catch (e) {
                            setDialogState(
                                () => jsonError = 'JSON Error: $e');
                          }
                        },
                        icon: Icon(Icons.check, size: 18),
                        label: Text(global.language('fd_apply')),
                        style: ElevatedButton.styleFrom(
                          backgroundColor:
                              global.theme.positiveHighlightTextColor,
                          foregroundColor: global.theme.onPrimaryColor,
                        ),
                      ),
                      const SizedBox(width: 8),
                      TextButton(
                        onPressed: () => Navigator.pop(context),
                        child: Text(global.language('cancel')),
                      ),
                    ],
                  ),
                ),
              ],
            ),
          ),
        ),
      ),
    );
  }

  // ── Build ──

  @override
  Widget build(BuildContext context) {
    final focusNode = FocusNode();

    return Focus(
      focusNode: focusNode,
      onKeyEvent: _handleKeyEvent,
      autofocus: true,
      child: GestureDetector(
        onTap: () => focusNode.requestFocus(),
        child: Scaffold(
          backgroundColor: global.theme.backgroundColor,
          body: _isPreviewMode ? _buildPreviewMode() : _buildEditMode(),
        ),
      ),
    );
  }

  Widget _buildPageTypeTabs() {
    return Container(
      height: 34,
      decoration: BoxDecoration(
        color: global.theme.surfaceColor,
        border: Border(
          bottom: BorderSide(color: global.theme.dividerBorderColor),
        ),
      ),
      child: Row(
        children: [
          const SizedBox(width: 12),
          Text('หน้า:', style: TextStyle(
            fontSize: 12, color: global.theme.textSecondaryColor)),
          const SizedBox(width: 8),
          for (final type in PageType.values)
            Padding(
              padding: const EdgeInsets.only(right: 2),
              child: _pageTypeTab(type),
            ),
          const Spacer(),
          // ปุ่มจัดการ: ถ้าเป็น middle/last สามารถ reset ได้
          if (_currentPageType != PageType.first)
            TextButton.icon(
              onPressed: () => _resetPageSection(_currentPageType),
              icon: Icon(Icons.restart_alt, size: 14,
                  color: global.theme.textSecondaryColor),
              label: Text('ใช้ค่าเดียวกับหน้าแรก',
                  style: TextStyle(fontSize: 11,
                      color: global.theme.textSecondaryColor)),
              style: TextButton.styleFrom(
                padding: const EdgeInsets.symmetric(horizontal: 8),
                minimumSize: Size.zero,
                tapTargetSize: MaterialTapTargetSize.shrinkWrap,
              ),
            ),
          const SizedBox(width: 8),
        ],
      ),
    );
  }

  Widget _pageTypeTab(PageType type) {
    final isSelected = _currentPageType == type;
    return InkWell(
      onTap: () {
        if (_currentPageType == type) return;
        // สร้าง section ถ้ายังไม่มี
        _initPageSections(type);
        setState(() {
          _currentPageType = type;
          // clear selection เมื่อสลับหน้า
          _selectedElement = null;
          _selectedSection = null;
          _selectedElementIds.clear();
        });
      },
      borderRadius: BorderRadius.circular(4),
      child: Container(
        padding: const EdgeInsets.symmetric(horizontal: 12, vertical: 4),
        decoration: BoxDecoration(
          color: isSelected
              ? global.theme.primaryColor.withValues(alpha: 0.15)
              : Colors.transparent,
          borderRadius: BorderRadius.circular(4),
          border: isSelected
              ? Border.all(color: global.theme.primaryColor, width: 1)
              : null,
        ),
        child: Row(
          mainAxisSize: MainAxisSize.min,
          children: [
            Icon(type.icon, size: 14,
                color: isSelected
                    ? global.theme.primaryColor
                    : global.theme.textSecondaryColor),
            const SizedBox(width: 4),
            Text(type.label,
                style: TextStyle(
                  fontSize: 11,
                  fontWeight: isSelected ? FontWeight.bold : FontWeight.normal,
                  color: isSelected
                      ? global.theme.primaryColor
                      : global.theme.textSecondaryColor,
                )),
          ],
        ),
      ),
    );
  }

  void _resetPageSection(PageType type) {
    switch (type) {
      case PageType.middle:
        _updateTemplate(
          currentTemplate.copyWith(
              clearHeaderMiddle: true, clearFooterMiddle: true),
          description: 'Reset middle page to default',
        );
        break;
      case PageType.last:
        _updateTemplate(
          currentTemplate.copyWith(clearHeaderLast: true),
          description: 'Reset last page header to default',
        );
        break;
      case PageType.first:
        break;
    }
  }

  Widget _buildEditMode() {
    return Column(
      children: [
        // ── Toolbar ──
        FormToolbar(
          canUndo: _undoStack.isNotEmpty,
          canRedo: _redoStack.isNotEmpty,
          onUndo: _undo,
          onRedo: _redo,
          zoomLevel: _zoomLevel,
          onZoomChanged: (z) => setState(() => _zoomLevel = z),
          showGrid: _showGrid,
          snapToGrid: _snapToGrid,
          showRulers: _showRulers,
          onToggleGrid: () => setState(() => _showGrid = !_showGrid),
          onToggleSnap: () => setState(() => _snapToGrid = !_snapToGrid),
          onToggleRulers: () => setState(() => _showRulers = !_showRulers),
          onAlignLeft: _alignLeft,
          onAlignCenter: _alignCenter,
          onAlignRight: _alignRight,
          onAlignTop: _alignTop,
          onAlignMiddle: _alignMiddle,
          onAlignBottom: _alignBottom,
          onDistributeH: _distributeH,
          onDistributeV: _distributeV,
          onBringToFront: _bringToFront,
          onSendToBack: _sendToBack,
          hasSelection: _selectedElementIds.isNotEmpty,
          hasMultiSelection: _selectedElementIds.length >= 2,
          onPreviewToggle: () =>
              setState(() => _isPreviewMode = !_isPreviewMode),
          isPreviewMode: _isPreviewMode,
          onSave: _saveTemplate,
          onLoad: _loadTemplate,
          onShowTemplateGallery: _showTemplateGallery,
          onShowFormInfo: _showFormInfoDialog,
          template: currentTemplate,
          onToggleLeftPanel: () =>
              setState(() => _showLeftPanel = !_showLeftPanel),
          onToggleRightPanel: () =>
              setState(() => _showRightPanel = !_showRightPanel),
          showLeftPanel: _showLeftPanel,
          showRightPanel: _showRightPanel,
          onShowJsonEditor: _showJsonDialog,
        ),

        // ── Page Type Tabs ──
        _buildPageTypeTabs(),

        // ── Main 3-Panel Area ──
        Expanded(
          child: Row(
            children: [
              // Left Panel: Element Palette + Layers
              if (_showLeftPanel)
                SizedBox(
                  width: 240,
                  child: Container(
                    decoration: BoxDecoration(
                      color: global.theme.cardColor,
                      border: Border(
                        right: BorderSide(
                            color: global.theme.dividerBorderColor),
                      ),
                    ),
                    child: Column(
                      children: [
                        // Element Palette
                        Expanded(
                          flex: 3,
                          child: ElementPalettePanel(),
                        ),
                        Divider(
                            height: 1,
                            color: global.theme.dividerBorderColor),
                        // Layers Panel
                        Expanded(
                          flex: 2,
                          child: LayersPanel(
                            template: _viewTemplate,
                            selectedElement: _selectedElement,
                            onElementSelected: (element, section) {
                              setState(() {
                                _selectedElement = element;
                                _selectedSection = section;
                              });
                            },
                            onElementVisibilityToggled: (element, section) {
                              final updated = element.copyWith(
                                  isVisible: !element.isVisible);
                              _executeCommand(UpdatePropertyCommand(
                                sectionType: section,
                                oldElement: element,
                                newElement: updated,
                              ));
                              if (_selectedElement?.id == element.id) {
                                setState(
                                    () => _selectedElement = updated);
                              }
                            },
                            onElementLockToggled: (element, section) {
                              final updated = element.copyWith(
                                  isLocked: !element.isLocked);
                              _executeCommand(UpdatePropertyCommand(
                                sectionType: section,
                                oldElement: element,
                                newElement: updated,
                              ));
                              if (_selectedElement?.id == element.id) {
                                setState(
                                    () => _selectedElement = updated);
                              }
                            },
                            onReorder: (template) {
                              _executeCommand(ReplaceTemplateCommand(
                                oldTemplate: currentTemplate,
                                newTemplate: template,
                                description: 'Reorder layers',
                              ));
                            },
                          ),
                        ),
                      ],
                    ),
                  ),
                ),

              // Center: Canvas
              Expanded(
                child: FullFormEditor(
                  template: _viewTemplate,
                  onTemplateChanged: (updatedTemplate) {
                    _applyViewTemplateChange(updatedTemplate);
                  },
                  zoomLevel: _zoomLevel,
                  showRulers: _showRulers,
                  showGrid: _showGrid,
                  snapToGrid: _snapToGrid,
                  gridSize: _gridSize,
                  selectedElement: _selectedElement,
                  selectedElementIds: _selectedElementIds,
                  onElementSelected: (element, section, {bool isShift = false}) {
                    setState(() {
                      if (isShift && element != null) {
                        // Multi-select: toggle element in set
                        if (_selectedElementIds.contains(element.id)) {
                          _selectedElementIds.remove(element.id);
                          if (_selectedElement?.id == element.id) {
                            // Switch primary to another selected, or null
                            if (_selectedElementIds.isNotEmpty) {
                              final nextId = _selectedElementIds.last;
                              final nextSection = _findElementSection(nextId);
                              if (nextSection != null) {
                                _selectedElement = _findElement(nextId, nextSection);
                                _selectedSection = nextSection;
                              }
                            } else {
                              _selectedElement = null;
                              _selectedSection = null;
                            }
                          }
                        } else {
                          _selectedElementIds.add(element.id);
                          // Also add current primary if not yet in set
                          if (_selectedElement != null) {
                            _selectedElementIds.add(_selectedElement!.id);
                          }
                          _selectedElement = element;
                          _selectedSection = section;
                        }
                      } else {
                        // Single select: clear multi-select
                        _selectedElementIds.clear();
                        _selectedElement = element;
                        _selectedSection = section;
                        if (element != null) {
                          _selectedElementIds.add(element.id);
                        }
                      }
                    });
                  },
                  onElementDeselected: () {
                    setState(() {
                      _selectedElement = null;
                      _selectedSection = null;
                      _selectedElementIds.clear();
                    });
                  },
                  onMousePositionChanged: (pos) {
                    setState(() => _mousePosition = pos);
                  },
                  onElementAdd: (type, section, position) {
                    _addElement(type, section, position: position);
                  },
                  onElementMoved: (section, elementId, oldX, oldY, newX, newY) {
                    setState(() {
                      _undoStack.add(MoveElementCommand(
                        sectionType: section,
                        elementId: elementId,
                        oldX: oldX,
                        oldY: oldY,
                        newX: newX,
                        newY: newY,
                      ));
                      _redoStack.clear();
                      if (_undoStack.length > _maxUndoSteps) {
                        _undoStack.removeRange(0, _undoStack.length - _maxUndoSteps);
                      }
                    });
                  },
                  onElementResized: (section, elementId, oldX, oldY, oldW, oldH, newX, newY, newW, newH) {
                    setState(() {
                      _undoStack.add(ResizeElementCommand(
                        sectionType: section,
                        elementId: elementId,
                        oldX: oldX,
                        oldY: oldY,
                        oldW: oldW,
                        oldH: oldH,
                        newX: newX,
                        newY: newY,
                        newW: newW,
                        newH: newH,
                      ));
                      _redoStack.clear();
                      if (_undoStack.length > _maxUndoSteps) {
                        _undoStack.removeRange(0, _undoStack.length - _maxUndoSteps);
                      }
                    });
                  },
                ),
              ),

              // Right Panel: Properties Inspector
              if (_showRightPanel)
                SizedBox(
                  width: 280,
                  child: Container(
                    decoration: BoxDecoration(
                      color: global.theme.cardColor,
                      border: Border(
                        left: BorderSide(
                            color: global.theme.dividerBorderColor),
                      ),
                    ),
                    child: PropertiesInspector(
                      selectedElement: _selectedElement,
                      selectedSection: _selectedSection,
                      onElementChanged: _updateSelectedElement,
                      onDeleteElement: _deleteSelectedElement,
                    ),
                  ),
                ),
            ],
          ),
        ),

        // ── Status Bar ──
        FormStatusBar(
          mousePosition: _mousePosition,
          selectedElement: _selectedElement,
          zoomLevel: _zoomLevel,
          template: currentTemplate,
          showGrid: _showGrid,
          gridSize: _gridSize,
          onPaperSizeChanged: (size) {
            _onPaperSizeChanged(size);
          },
          onOrientationChanged: (orientation) {
            _onOrientationChanged(orientation);
          },
          onPrintModeChanged: (mode) {
            _updateTemplate(
              currentTemplate.copyWith(printMode: mode),
              description: 'Change print mode to ${mode.label}',
            );
          },
          onZoomChanged: (zoom) => _onZoomChanged(zoom),
        ),
      ],
    );
  }

  Widget _buildPreviewMode() {
    return Column(
      children: [
        // Toolbar in preview mode
        FormToolbar(
          canUndo: false,
          canRedo: false,
          onUndo: () {},
          onRedo: () {},
          zoomLevel: _zoomLevel,
          onZoomChanged: (z) => setState(() => _zoomLevel = z),
          showGrid: false,
          snapToGrid: false,
          showRulers: false,
          onToggleGrid: () {},
          onToggleSnap: () {},
          onToggleRulers: () {},
          onAlignLeft: () {},
          onAlignCenter: () {},
          onAlignRight: () {},
          onAlignTop: () {},
          onAlignMiddle: () {},
          onAlignBottom: () {},
          onDistributeH: () {},
          onDistributeV: () {},
          onBringToFront: () {},
          onSendToBack: () {},
          hasSelection: false,
          onPreviewToggle: () =>
              setState(() => _isPreviewMode = !_isPreviewMode),
          isPreviewMode: true,
          onSave: _saveTemplate,
          onLoad: _loadTemplate,
          onShowTemplateGallery: _showTemplateGallery,
          onShowFormInfo: _showFormInfoDialog,
          template: currentTemplate,
          onToggleLeftPanel: () {},
          onToggleRightPanel: () {},
          showLeftPanel: false,
          showRightPanel: false,
          onShowJsonEditor: _showJsonDialog,
        ),
        Expanded(
          child: Center(
            child: SingleChildScrollView(
              padding: const EdgeInsets.all(16),
              child: FormPreview(
                  template: currentTemplate, scale: _zoomLevel > 0 ? _zoomLevel * 0.8 : 0.8),
            ),
          ),
        ),
      ],
    );
  }
}
