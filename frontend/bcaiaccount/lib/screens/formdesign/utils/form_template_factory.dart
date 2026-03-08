import 'package:flutter/foundation.dart';
import 'dart:convert';
import 'dart:io';
import 'package:flutter/material.dart';
import '../models/form_template.dart';
import '../models/form_section.dart';
import '../models/form_element.dart';
import '../models/paper_size.dart';
import 'package:smlaicloud/utils/logger/app_logger.dart';
import '../../../global.dart' as global;

/// ตัวอย่างการสร้าง Form Template ด้วยโค้ด
/// ใช้สำหรับกรณีที่ต้องการสร้างฟอร์มแบบ programmatic
class FormTemplateFactory {
  /// สร้างใบสั่งซื้อพื้นฐาน
  static FormTemplate createPurchaseOrder() {
    return FormTemplate(
      name: global.language('transaction_purchase_order'),
      description: global.language('form_purchase_order_desc'),
      paperSize: PaperSize.a4,
      header: FormSection(
        id: 'header',
        type: SectionType.header,
        height: 150,
        elements: [
          FormElement(
            id: 'title',
            type: ElementType.text,
            x: 50,
            y: 20,
            width: 400,
            height: 40,
            text: global.language('transaction_purchase_order'),
            textStyle: const TextStyle(
              fontSize: 28,
              fontWeight: FontWeight.bold,
              color: Colors.black,
            ),
            textAlign: TextAlign.center,
          ),
          FormElement(
            id: 'header_box',
            type: ElementType.rectangle,
            x: 40,
            y: 10,
            width: 515,
            height: 130,
            borderColor: Colors.black,
            borderWidth: 2,
          ),
        ],
      ),
      detail: FormSection(
        id: 'detail',
        type: SectionType.detail,
        height: 500,
        elements: [
          FormElement(
            id: 'items_table',
            type: ElementType.table,
            x: 50,
            y: 70,
            width: 495,
            height: 300,
            rows: 6,
            columns: 5,
            tableData: [
              [global.language('pdf_sequence'), global.language('pdf_item'), global.language('enter_qty'), global.language('pdf_price_per_unit'), global.language('cash_amount')],
              ['1', '', '', '', ''],
              ['2', '', '', '', ''],
              ['3', '', '', '', ''],
              ['4', '', '', '', ''],
              [global.language('product_amount'), '', '', '', ''],
            ],
            borderColor: Colors.black,
            borderWidth: 1,
          ),
        ],
      ),
      footer: FormSection(
        id: 'footer',
        type: SectionType.footer,
        height: 150,
        elements: [
          FormElement(
            id: 'signature_box',
            type: ElementType.rectangle,
            x: 50,
            y: 20,
            width: 230,
            height: 100,
            borderColor: Colors.black,
            borderWidth: 1,
          ),
          FormElement(
            id: 'signature_label',
            type: ElementType.text,
            x: 60,
            y: 30,
            width: 210,
            height: 80,
            text:
                '${global.language("pdf_orderer")}\n\n_____________________\n(                                    )',
            textStyle: const TextStyle(fontSize: 12, color: Colors.black),
            textAlign: TextAlign.center,
          ),
        ],
      ),
    );
  }

  /// สร้างใบรับจองพื้นฐาน
  static FormTemplate createBookingForm() {
    return FormTemplate(
      name: global.language('booking_receipt'),
      description: global.language('form_booking_desc'),
      paperSize: PaperSize.a4,
      header: FormSection(
        id: 'header',
        type: SectionType.header,
        height: 180,
        elements: [
          FormElement(
            id: 'title',
            type: ElementType.text,
            x: 200,
            y: 30,
            width: 200,
            height: 50,
            text: global.language('booking_receipt'),
            textStyle: const TextStyle(
              fontSize: 32,
              fontWeight: FontWeight.bold,
              color: Colors.blue,
            ),
            textAlign: TextAlign.center,
          ),
        ],
      ),
      detail: FormSection(
        id: 'detail',
        type: SectionType.detail,
        height: 450,
        elements: [
          FormElement(
            id: 'booking_table',
            type: ElementType.table,
            x: 50,
            y: 200,
            width: 495,
            height: 200,
            rows: 5,
            columns: 4,
            tableData: [
              [global.language('pdf_item'), global.language('bill_design_detail'), global.language('date_time_section'), global.language('cash_amount')],
              ['', '', '', ''],
              ['', '', '', ''],
              ['', '', '', ''],
              [global.language('amount'), '', '', ''],
            ],
            borderColor: Colors.black,
            borderWidth: 1,
          ),
        ],
      ),
      footer: FormSection(
        id: 'footer',
        type: SectionType.footer,
        height: 180,
      ),
    );
  }

  /// สร้างใบเสนอราคาพื้นฐาน
  static FormTemplate createQuotation() {
    return FormTemplate(
      name: global.language('transaction_quotation'),
      description: global.language('form_quotation_desc'),
      paperSize: PaperSize.a4,
      header: FormSection(
        id: 'header',
        type: SectionType.header,
        height: 150,
        elements: [
          FormElement(
            id: 'title',
            type: ElementType.text,
            x: 50,
            y: 20,
            width: 495,
            height: 40,
            text: 'ใบเสนอราคา / QUOTATION',
            textStyle: const TextStyle(
              fontSize: 26,
              fontWeight: FontWeight.bold,
              color: Colors.black,
            ),
            textAlign: TextAlign.center,
          ),
        ],
      ),
      detail: FormSection(
        id: 'detail',
        type: SectionType.detail,
        height: 500,
        elements: [
          FormElement(
            id: 'quotation_table',
            type: ElementType.table,
            x: 50,
            y: 100,
            width: 495,
            height: 350,
            rows: 8,
            columns: 5,
            tableData: [
              [global.language('pdf_sequence'), global.language('pdf_item'), global.language('enter_qty'), global.language('pdf_price_per_unit'), global.language('cash_amount')],
              ['1', '', '', '', ''],
              ['2', '', '', '', ''],
              ['3', '', '', '', ''],
              ['4', '', '', '', ''],
              ['5', '', '', '', ''],
              ['', '', '', 'รวมเงิน', ''],
              ['', '', '', 'ภาษี 7%', ''],
            ],
            borderColor: Colors.black,
            borderWidth: 1,
          ),
        ],
      ),
      footer: FormSection(
        id: 'footer',
        type: SectionType.footer,
        height: 150,
        elements: [
          FormElement(
            id: 'terms',
            type: ElementType.text,
            x: 50,
            y: 10,
            width: 495,
            height: 40,
            text: global.language('quotation_validity_note'),
            textStyle: const TextStyle(fontSize: 11, color: Colors.red),
          ),
        ],
      ),
    );
  }

  /// สร้างฟอร์มแบบกำหนดเอง
  static FormTemplate createCustomForm({
    required String name,
    required String description,
    PaperSize paperSize = PaperSize.a4,
  }) {
    return FormTemplate(
      name: name,
      description: description,
      paperSize: paperSize,
    );
  }

  /// บันทึก Template เป็นไฟล์ JSON
  static Future<void> saveTemplate(
    FormTemplate template,
    String filePath,
  ) async {
    final jsonString = jsonEncode(template.toJson());
    final file = File(filePath);
    await file.writeAsString(jsonString);
  }

  /// โหลด Template จากไฟล์ JSON
  static Future<FormTemplate> loadTemplate(String filePath) async {
    final file = File(filePath);
    final jsonString = await file.readAsString();
    final jsonData = jsonDecode(jsonString);
    return FormTemplate.fromJson(jsonData);
  }
}

/// ตัวอย่างการใช้งาน
void main() async {
  // สร้างใบสั่งซื้อ
  final purchaseOrder = FormTemplateFactory.createPurchaseOrder();
  if (kDebugMode) {
    AppLogger.debug('Created: ${purchaseOrder.name}');
  }

  // สร้างใบรับจอง
  final booking = FormTemplateFactory.createBookingForm();
  if (kDebugMode) {
    AppLogger.debug('Created: ${booking.name}');
  }

  // สร้างใบเสนอราคา
  final quotation = FormTemplateFactory.createQuotation();
  if (kDebugMode) {
    AppLogger.debug('Created: ${quotation.name}');
  }

  // สร้างฟอร์มกำหนดเอง
  final custom = FormTemplateFactory.createCustomForm(
    name: global.language('my_form'),
    description: global.language('custom_form'),
    paperSize: PaperSize.b5,
  );
  if (kDebugMode) {
    AppLogger.debug('Created: ${custom.name}');
  }

  // บันทึกเป็นไฟล์
  // await FormTemplateFactory.saveTemplate(
  //   purchaseOrder,
  //   'my_purchase_order.json',
  // );

  // โหลดจากไฟล์
  // final loaded = await FormTemplateFactory.loadTemplate('my_purchase_order.json');
  // print('Loaded: ${loaded.name}');
}
