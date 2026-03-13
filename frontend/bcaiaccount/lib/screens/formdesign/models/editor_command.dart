import 'package:flutter/material.dart';
import 'form_element.dart';
import 'form_layout.dart';
import 'form_section.dart';
import 'form_template.dart';

/// Command pattern สำหรับ undo/redo
abstract class EditorCommand {
  String get description;
  FormTemplate execute(FormTemplate template);
  FormTemplate undo(FormTemplate template);
}

/// Batch multiple commands into one undo/redo step
class BatchCommand extends EditorCommand {
  final List<EditorCommand> commands;
  final String _description;

  BatchCommand({required this.commands, String description = 'Batch'})
      : _description = description;

  @override
  String get description => _description;

  @override
  FormTemplate execute(FormTemplate template) {
    var t = template;
    for (final cmd in commands) {
      t = cmd.execute(t);
    }
    return t;
  }

  @override
  FormTemplate undo(FormTemplate template) {
    var t = template;
    for (final cmd in commands.reversed) {
      t = cmd.undo(t);
    }
    return t;
  }
}

/// Add element to a section (supports both fixed + row-based)
class AddElementCommand extends EditorCommand {
  final SectionType sectionType;
  final FormElement element;

  AddElementCommand({required this.sectionType, required this.element});

  @override
  String get description => 'Add ${element.type.label}';

  @override
  FormTemplate execute(FormTemplate template) {
    final section = _getSection(template, sectionType);
    if (section.hasRows) {
      // Row-based: add new row with single cell at end
      final cell = _elementToCell(element);
      final newRow = FormLayoutRow(
        id: 'row_${element.id}',
        cells: [cell],
      );
      final updatedRows = [...section.rows, newRow];
      return _updateSection(
          template, sectionType, section.copyWith(rows: updatedRows));
    }
    final updatedElements = [...section.elements, element];
    return _updateSection(
        template, sectionType, section.copyWith(elements: updatedElements));
  }

  @override
  FormTemplate undo(FormTemplate template) {
    final section = _getSection(template, sectionType);
    if (section.hasRows) {
      final updatedRows = section.rows
          .where((r) => !r.cells.any((c) => c.id == element.id))
          .toList();
      return _updateSection(
          template, sectionType, section.copyWith(rows: updatedRows));
    }
    final updatedElements =
        section.elements.where((e) => e.id != element.id).toList();
    return _updateSection(
        template, sectionType, section.copyWith(elements: updatedElements));
  }
}

/// Delete element from a section (supports both fixed + row-based)
class DeleteElementCommand extends EditorCommand {
  final SectionType sectionType;
  final FormElement element;

  DeleteElementCommand({required this.sectionType, required this.element});

  @override
  String get description => 'Delete ${element.displayName}';

  @override
  FormTemplate execute(FormTemplate template) {
    final section = _getSection(template, sectionType);
    if (section.hasRows) {
      return _removeCellFromRows(template, sectionType, section, element.id);
    }
    final updatedElements =
        section.elements.where((e) => e.id != element.id).toList();
    return _updateSection(
        template, sectionType, section.copyWith(elements: updatedElements));
  }

  @override
  FormTemplate undo(FormTemplate template) {
    final section = _getSection(template, sectionType);
    if (section.hasRows) {
      final cell = _elementToCell(element);
      final newRow = FormLayoutRow(
        id: 'row_${element.id}',
        cells: [cell],
      );
      final updatedRows = [...section.rows, newRow];
      return _updateSection(
          template, sectionType, section.copyWith(rows: updatedRows));
    }
    final updatedElements = [...section.elements, element];
    return _updateSection(
        template, sectionType, section.copyWith(elements: updatedElements));
  }
}

/// Move element to new position
class MoveElementCommand extends EditorCommand {
  final SectionType sectionType;
  final String elementId;
  final double oldX, oldY;
  final double newX, newY;

  MoveElementCommand({
    required this.sectionType,
    required this.elementId,
    required this.oldX,
    required this.oldY,
    required this.newX,
    required this.newY,
  });

  @override
  String get description => 'Move element';

  @override
  FormTemplate execute(FormTemplate template) {
    return _updateElement(template, sectionType, elementId,
        (e) => e.copyWith(x: newX, y: newY));
  }

  @override
  FormTemplate undo(FormTemplate template) {
    return _updateElement(template, sectionType, elementId,
        (e) => e.copyWith(x: oldX, y: oldY));
  }
}

/// Resize element
class ResizeElementCommand extends EditorCommand {
  final SectionType sectionType;
  final String elementId;
  final double oldX, oldY, oldW, oldH;
  final double newX, newY, newW, newH;

  ResizeElementCommand({
    required this.sectionType,
    required this.elementId,
    required this.oldX,
    required this.oldY,
    required this.oldW,
    required this.oldH,
    required this.newX,
    required this.newY,
    required this.newW,
    required this.newH,
  });

  @override
  String get description => 'Resize element';

  @override
  FormTemplate execute(FormTemplate template) {
    return _updateElement(template, sectionType, elementId,
        (e) => e.copyWith(x: newX, y: newY, width: newW, height: newH));
  }

  @override
  FormTemplate undo(FormTemplate template) {
    return _updateElement(template, sectionType, elementId,
        (e) => e.copyWith(x: oldX, y: oldY, width: oldW, height: oldH));
  }
}

/// Update element property
class UpdatePropertyCommand extends EditorCommand {
  final SectionType sectionType;
  final FormElement oldElement;
  final FormElement newElement;

  UpdatePropertyCommand({
    required this.sectionType,
    required this.oldElement,
    required this.newElement,
  });

  @override
  String get description => 'Update ${newElement.type.label}';

  @override
  FormTemplate execute(FormTemplate template) {
    return _replaceElement(template, sectionType, oldElement.id, newElement);
  }

  @override
  FormTemplate undo(FormTemplate template) {
    return _replaceElement(template, sectionType, newElement.id, oldElement);
  }
}

/// Reorder element z-index
class ReorderElementCommand extends EditorCommand {
  final SectionType sectionType;
  final String elementId;
  final int oldIndex;
  final int newIndex;

  ReorderElementCommand({
    required this.sectionType,
    required this.elementId,
    required this.oldIndex,
    required this.newIndex,
  });

  @override
  String get description => 'Reorder element';

  @override
  FormTemplate execute(FormTemplate template) {
    return _reorderElement(template, sectionType, oldIndex, newIndex);
  }

  @override
  FormTemplate undo(FormTemplate template) {
    return _reorderElement(template, sectionType, newIndex, oldIndex);
  }
}

/// Resize section height
class ResizeSectionCommand extends EditorCommand {
  final SectionType sectionType;
  final double oldHeight;
  final double newHeight;

  ResizeSectionCommand({
    required this.sectionType,
    required this.oldHeight,
    required this.newHeight,
  });

  @override
  String get description => 'Resize ${sectionType.name}';

  @override
  FormTemplate execute(FormTemplate template) {
    final section = _getSection(template, sectionType);
    return _updateSection(
        template, sectionType, section.copyWith(height: newHeight));
  }

  @override
  FormTemplate undo(FormTemplate template) {
    final section = _getSection(template, sectionType);
    return _updateSection(
        template, sectionType, section.copyWith(height: oldHeight));
  }
}

/// Replace entire template (for form info, JSON editor, etc.)
class ReplaceTemplateCommand extends EditorCommand {
  final FormTemplate oldTemplate;
  final FormTemplate newTemplate;
  final String _description;

  ReplaceTemplateCommand({
    required this.oldTemplate,
    required this.newTemplate,
    String description = 'Update template',
  }) : _description = description;

  @override
  String get description => _description;

  @override
  FormTemplate execute(FormTemplate template) => newTemplate;

  @override
  FormTemplate undo(FormTemplate template) => oldTemplate;
}

// ── Helpers ──

FormSection _getSection(FormTemplate template, SectionType type) {
  switch (type) {
    case SectionType.header:
      return template.header;
    case SectionType.detail:
      return template.detail;
    case SectionType.footer:
      return template.footer;
  }
}

FormTemplate _updateSection(
    FormTemplate template, SectionType type, FormSection section) {
  switch (type) {
    case SectionType.header:
      return template.copyWith(header: section);
    case SectionType.detail:
      return template.copyWith(detail: section);
    case SectionType.footer:
      return template.copyWith(footer: section);
  }
}

FormTemplate _updateElement(FormTemplate template, SectionType type,
    String elementId, FormElement Function(FormElement) updater) {
  final section = _getSection(template, type);
  if (section.hasRows) {
    // Row-based: find cell, apply updater to resolved element, write back to cell
    final updatedRows = section.rows.map((row) {
      final updatedCells = row.cells.map((cell) {
        if (cell.id != elementId) return cell;
        // Create a dummy element from cell, apply updater, write back
        final resolved = FormLayoutResolver.resolve([FormLayoutRow(cells: [cell])], 1000);
        if (resolved.elements.isEmpty) return cell;
        final updated = updater(resolved.elements.first);
        return _elementToCell(updated);
      }).toList();
      return row.copyWith(cells: updatedCells);
    }).toList();
    return _updateSection(template, type, section.copyWith(rows: updatedRows));
  }
  final updatedElements = section.elements.map((e) {
    if (e.id == elementId) return updater(e);
    return e;
  }).toList();
  return _updateSection(
      template, type, section.copyWith(elements: updatedElements));
}

FormTemplate _replaceElement(FormTemplate template, SectionType type,
    String elementId, FormElement newElement) {
  final section = _getSection(template, type);
  if (section.hasRows) {
    final updatedRows = section.rows.map((row) {
      final updatedCells = row.cells.map((cell) {
        if (cell.id != elementId) return cell;
        return _elementToCell(newElement);
      }).toList();
      return row.copyWith(cells: updatedCells);
    }).toList();
    return _updateSection(template, type, section.copyWith(rows: updatedRows));
  }
  final updatedElements = section.elements.map((e) {
    if (e.id == elementId) return newElement;
    return e;
  }).toList();
  return _updateSection(
      template, type, section.copyWith(elements: updatedElements));
}

/// Convert FormElement → FormLayoutCell (for adding to row-based sections)
FormLayoutCell _elementToCell(FormElement el) {
  return FormLayoutCell(
    id: el.id,
    type: el.type,
    flex: 1.0,
    text: el.text,
    dataBindingKey: el.dataBindingKey,
    dataBindingFormat: el.dataBindingFormat,
    imagePath: el.imagePath,
    fontFamily: el.fontFamily ?? 'Sarabun',
    fontSize: el.fontSize ?? 7.0,
    fontBold: el.fontBold,
    fontItalic: el.fontItalic,
    fontUnderline: el.fontUnderline,
    textAlign: el.textAlign ?? TextAlign.left,
    textColor: el.textColor,
    backgroundColor: el.backgroundColor,
    borderColor: el.borderColor,
    borderWidth: el.borderWidth,
    borderRadius: el.borderRadius,
    opacity: el.opacity,
    barcodeFormat: el.barcodeFormat,
    barcodeValue: el.barcodeValue,
    lineStyle: el.lineStyle,
    rows: el.rows,
    columns: el.columns,
    tableData: el.tableData,
    tableHasHeader: el.tableHasHeader,
  );
}

/// Remove cell from rows by ID; remove empty rows
FormTemplate _removeCellFromRows(FormTemplate template, SectionType type,
    FormSection section, String cellId) {
  final updatedRows = <FormLayoutRow>[];
  for (final row in section.rows) {
    final filteredCells = row.cells.where((c) => c.id != cellId).toList();
    if (filteredCells.isNotEmpty) {
      updatedRows.add(row.copyWith(cells: filteredCells));
    }
  }
  return _updateSection(template, type, section.copyWith(rows: updatedRows));
}

FormTemplate _reorderElement(
    FormTemplate template, SectionType type, int oldIndex, int newIndex) {
  final section = _getSection(template, type);
  final elements = [...section.elements];
  if (oldIndex < 0 ||
      oldIndex >= elements.length ||
      newIndex < 0 ||
      newIndex >= elements.length) {
    return template;
  }
  final element = elements.removeAt(oldIndex);
  elements.insert(newIndex, element);
  return _updateSection(
      template, type, section.copyWith(elements: elements));
}
