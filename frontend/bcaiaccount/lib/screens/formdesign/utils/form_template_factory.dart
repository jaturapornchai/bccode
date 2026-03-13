import 'dart:convert';
import 'dart:io';
import 'package:flutter/services.dart' show rootBundle;
import '../models/form_style.dart';
import '../models/form_template.dart';

/// Factory สำหรับโหลด Form Template จาก JSON assets
/// Templates ทั้งหมดเก็บเป็น JSON 100% ใน assets/form_templates/
class FormTemplateFactory {
  // Cache loaded templates to avoid repeated asset reads
  static final Map<String, FormTemplate> _cache = {};

  // ─── Public API ────────────────────────────────────────────────────

  /// รายชื่อ template ทั้งหมดที่รองรับ
  static List<String> get availableTypes => [
        'tax_invoice',
        'abbreviated_tax_invoice',
        'receipt',
        'cash_bill',
        'quotation',
        'purchase_order',
        'delivery_note',
        'invoice',
        'credit_note',
        'debit_note',
        'booking_form',
        'payment_voucher',
      ];

  /// สร้าง template ตาม document type identifier (async — โหลดจาก JSON asset)
  /// [styleId] — ถ้าระบุจะ apply FormStyle ทับสีและเส้นขอบ
  static Future<FormTemplate> createTemplate(String documentType,
      {String? styleId}) async {
    // Return cached copy (deep clone via JSON round-trip)
    if (_cache.containsKey(documentType)) {
      return _applyStyle(_cloneTemplate(_cache[documentType]!), styleId);
    }

    try {
      final jsonString = await rootBundle
          .loadString('assets/form_templates/$documentType.json');
      final jsonData = jsonDecode(jsonString) as Map<String, dynamic>;
      final template = FormTemplate.fromJson(jsonData);
      _cache[documentType] = template;
      return _applyStyle(_cloneTemplate(template), styleId);
    } catch (e) {
      // Fallback: return blank template if JSON not found
      return _createBlankFallback(documentType);
    }
  }

  /// Apply style to template if styleId is provided
  static FormTemplate _applyStyle(FormTemplate template, String? styleId) {
    if (styleId == null) return template;
    final style = findFormStyle(styleId);
    if (style == null) return template;
    return style.apply(template);
  }

  /// Synchronous version — uses pre-loaded cache (call init() first)
  static FormTemplate createTemplateSync(String documentType) {
    if (_cache.containsKey(documentType)) {
      return _cloneTemplate(_cache[documentType]!);
    }
    // Return blank if available
    if (_cache.containsKey('blank')) {
      return _cloneTemplate(_cache['blank']!);
    }
    // Absolute fallback — create minimal blank template
    return _createBlankFallback(documentType);
  }

  /// Pre-load all templates into cache (call once at app startup)
  static Future<void> init() async {
    for (final type in availableTypes) {
      await createTemplate(type);
    }
    // Also preload blank
    await createTemplate('blank');
  }

  /// Clear cache
  static void clearCache() => _cache.clear();

  // ─── Save / Load (user files) ──────────────────────────────────────

  /// บันทึก Template เป็นไฟล์ JSON
  static Future<void> saveTemplate(
    FormTemplate template,
    String filePath,
  ) async {
    final jsonString =
        const JsonEncoder.withIndent('  ').convert(template.toJson());
    final file = File(filePath);
    await file.writeAsString(jsonString);
  }

  /// โหลด Template จากไฟล์ JSON
  static Future<FormTemplate> loadTemplate(String filePath) async {
    final file = File(filePath);
    final jsonString = await file.readAsString();
    final jsonData = jsonDecode(jsonString) as Map<String, dynamic>;
    return FormTemplate.fromJson(jsonData);
  }

  // ─── Private helpers ───────────────────────────────────────────────

  /// Deep clone template via JSON round-trip to prevent mutation of cache
  static FormTemplate _cloneTemplate(FormTemplate template) {
    final json = template.toJson();
    return FormTemplate.fromJson(json);
  }

  /// Fallback blank template when JSON asset not found
  static FormTemplate _createBlankFallback(String documentType) {
    return FormTemplate(
      name: documentType,
      description: 'Template: $documentType',
      documentType: documentType,
    );
  }
}
