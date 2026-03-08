#!/usr/bin/env python3
"""
Automated localization script for BC AI Cloud Flutter App.
Replaces hardcoded Thai strings with language() calls.

Usage:
  python scripts/localize_all.py --dry-run    # Preview changes
  python scripts/localize_all.py --apply      # Apply changes
"""

import os, re, json, sys, hashlib

# === CONFIG ===
LIB_DIR = r'D:\bcdev\bcaiaccount\lib'
LANG_JSON = r'D:\bcdev\backend\assets\language\languages.json'
BS = '\\'

# Login flow files — ยังไม่ load language.json (SKIP)
SKIP_FILES = {
    'lib/usersystem/login_password_screen.dart',
    'lib/usersystem/login_screen.dart',
    'lib/usersystem/registration.dart',
    'lib/select_language_screen.dart',
    'lib/usersystem/login_line_screen.dart',
    'lib/usersystem/select_shop.dart',
    # Config/URL setup screens (ก่อน connect backend)
    'lib/usersystem/server_config_screen.dart',
}

# Files ที่มี sample/mock data — skip Thai strings ที่เป็น test data
SAMPLE_DATA_FILES = {
    'lib/screens/receive/receivelist.dart',  # 177 items ส่วนใหญ่ mock data
}

# Patterns ที่เป็น sample/test data — skip
SAMPLE_PATTERNS = [
    # Mock customer names
    r'สิธิพัช', r'เบอรรี่', r'วงเหลา', r'สมใจ', r'เนื่องแป้น', r'คำปลา', r'ดูดี', r'สามวง',
    # Mock product names in receivelist
    r'บะหมี่กึ่งสำเร็จรูป', r'ซอสถั่วเหลือง', r'น้ำปลา', r'ข้าวสาร', r'น้ำตาลทราย', r'ผงซักฟอก',
    # Mock prices
    r'^\s*["\x27]\d[\d,]*\.?\d*\s*บาท["\x27]',  # "700บาท", "1,800.00 บาท"
    r'700บาท', r'600บาท', r'500บาท', r'1,800\.00 บาท', r'900\.00 บาท',
    # Mock company names
    r'บริษัท บ้านเชียง', r'ร้านโซลาว',
    # Cart system displayNameThai (enum values, not user-facing)
    r'displayNameThai',
]

# Strings ที่ OK ไม่ต้องแก้
SKIP_LINE_PATTERNS = [
    r'^\s*(//|/\*|\*)',          # comments
    r'AppLogger', r'debugPrint', r'print\(',  # logging
    r'LanguageDataModel\(', r'PublicNameModel\(', r'LanguageModel\(',  # translation data
    r'NumberToWord',             # number-to-word constants
    r'static const String _',   # constants
    r'languageCode:', r'codeTranslator:',
    r'code:\s*"th"', r'code:\s*"en"',
]

# Month/day names arrays — keep as is
SKIP_STRINGS = {
    '฿', 'มกราคม', 'กุมภาพันธ์', 'มีนาคม', 'เมษายน', 'พฤษภาคม',
    'มิถุนายน', 'กรกฎาคม', 'สิงหาคม', 'กันยายน', 'ตุลาคม',
    'พฤศจิกายน', 'ธันวาคม', 'ม.ค.', 'ก.พ.', 'มี.ค.', 'เม.ย.',
    'พ.ค.', 'มิ.ย.', 'ก.ค.', 'ส.ค.', 'ก.ย.', 'ต.ค.', 'พ.ย.', 'ธ.ค.',
    'จันทร์', 'อังคาร', 'พุธ', 'พฤหัสบดี', 'ศุกร์', 'เสาร์', 'อาทิตย์',
    'จ', 'อ', 'พ', 'พฤ', 'ศ', 'ส', 'อา',
    # NumberToWord
    'บาท', 'สตางค์', 'ศูนย์', 'เอ็ด', 'ร้อย', 'พัน', 'หมื่น', 'แสน',
    'ล้าน', 'พันล้าน', 'หนึ่ง', 'สอง', 'สาม', 'สี่', 'ห้า', 'หก',
    'เจ็ด', 'แปด', 'เก้า', 'สิบ', 'สิบเอ็ด', 'สิบสอง', 'สิบสาม',
    'สิบสี่', 'สิบห้า', 'สิบหก', 'สิบเจ็ด', 'สิบแปด', 'สิบเก้า',
    'ยี่สิบ', 'สามสิบ', 'สี่สิบ', 'ห้าสิบ', 'หกสิบ', 'เจ็ดสิบ',
    'แปดสิบ', 'เก้าสิบ', 'ถ้วน',
    # Language names
    'ไทย',
}

# === MANUAL KEY MAPPINGS ===
# สำหรับ strings ที่รู้แน่ว่าต้อง map เป็น key อะไร
MANUAL_MAP = {
    # PDF common labels
    'ผู้จ่ายเงิน.......................................................': 'pdf_payer_sign',
    'ผู้รับเงิน.......................................................': 'pdf_receiver_sign',
    'วันที่.......................................................': 'pdf_date_sign',
    'วันที่.......................................': 'pdf_date_sign_short',
    'หน้า': 'pdf_page',
    'เจ้าหนี้': 'creditor',
    'ลูกค้า': 'customer',
    'ที่อยู่': 'address',
    'เลขที่เอกสาร': 'docno',
    'วันที่เอกสาร': 'document_date',
    'เลขที่ใบกำกับ': 'vat_num',
    'วันที่ใบกำกับ': 'pdf_vat_date',
    'เลขประจำตัวผู้เสียภาษีอากร': 'pdf_tax_id_number',
    'รวมมูลค่า': 'doc_total_value',
    'หัก ส่วนลด': 'pdf_discount',
    'รวมมูลค่าก่อนภาษี': 'pdf_total_before_vat',
    'สินค้ายกเว้นภาษี': 'pdf_tax_exempt_goods',
    'สินค้ามีภาษี': 'pdf_taxable_goods',
    'รวมเงินทั้งสิ้น': 'pdf_grand_total',
    'มูลค่าสินค้า': 'pdf_goods_value',
    'ส่วนลด': 'discount',
    'ราคา/หน่วย': 'pdf_price_per_unit',
    'ราคารวม': 'pdf_total_price',
    'ลำดับ': 'pdf_sequence',
    'รหัสสินค้า': 'product_code',
    'รายการ': 'pdf_item',
    'จำนวน': 'enter_qty',
    'หน่วย': 'chatbot_product_unit',
    'ราคา': 'price',
    'บาร์โค้ด': 'barcode',
    'สินค้า': 'product',
    'สาขา': 'branch',
    'ลูกหนี้': 'debtor',
    'แผนก': 'department',
    'คลังสินค้า': 'warehouse',
    'รายการสินค้า': 'doc_details',
    'จำนวนเงิน': 'cash_amount',
    'ทั้งหมด': 'all',
    'บันทึก': 'form_design_save',
    'ยกเลิก': 'cancel',
    'ลบ': 'delete',
    'แก้ไข': 'edit',
    'เพิ่ม': 'add',
    'ค้นหา': 'search',
    'ปิด': 'close',
    'ตกลง': 'ok',
    'ยืนยัน': 'confirm',
    'เลือก': 'select',
    'กรุณารอสักครู่...': 'please_wait',
    'กำลังโหลด...': 'loading',
    'ไม่พบข้อมูล': 'no_data',
    'หมายเหตุ': 'remark',
    'วันที่': 'date',
    'รหัส': 'code',
    'ชื่อ': 'name',
    'สถานะ': 'status',
    'ประเภท': 'type',
    'ลำดับที่': 'group_number',
    'ผู้ขาย': 'seller',
    'ผู้ซื้อ': 'buyer',
    'เบอร์โทร': 'phone',
    'อีเมล': 'email',
    'พนักงาน': 'employee',
    'พนักงานขาย': 'sale_person',
    'หมวดหมู่': 'product_category',
    'กลุ่มสินค้า': 'product_group',
    'หน่วยนับ': 'unit',
    'คงเหลือ': 'balance',
    'ต้นทุน': 'cost',
    'กำไร': 'profit',
    'ขาย': 'transaction_sale',
    'ซื้อ': 'transaction_purchase',
    'รวม': 'total',
    'สุทธิ': 'net',
    'ภาษี': 'vat',
    'ดูรายละเอียด': 'view_detail',
    'กรุณาเลือก': 'please_select',
    'ไม่มีข้อมูล': 'no_data_available',
    'โหลดข้อมูลสำเร็จ': 'load_data_success',
    'บันทึกสำเร็จ': 'save_success',
    'บันทึกไม่สำเร็จ': 'save_failed',
    'เกิดข้อผิดพลาด': 'error_occurred',
    'ลบสำเร็จ': 'delete_success',
    'กรุณากรอกข้อมูล': 'please_fill_data',
    'ข้อมูลไม่ถูกต้อง': 'invalid_data',
    'กำลังบันทึก...': 'saving',
    'กำลังลบ...': 'deleting',
    'เลือกวันที่': 'select_date',
    'เริ่มต้น': 'start',
    'สิ้นสุด': 'end',
    'ล้างทั้งหมด': 'clear_all',
    'เลือกทั้งหมด': 'select_all',
    'ไม่มีรายการ': 'no_items',
    'กำลังโหลดรูปภาพ...': 'loading_image',
    'กำลังสร้าง PDF...': 'generating_pdf',
    'ล้มเหลว': 'failed',
    'สำเร็จ': 'success',
    'เปิด': 'open',
    'ปิดการใช้งาน': 'deactivate',
    'เปิดการใช้งาน': 'activate',
    'คลัง': 'warehouse',
    'ตำแหน่ง': 'location',
    'เลือกคลังสินค้า': 'select_warehouse',
    'เลือกตำแหน่ง': 'select_location',
    'เลือกสาขา': 'select_branch',
    'เลือกแผนก': 'select_department',
    'ตั้งค่า': 'setting',
    'ข้อมูลหลัก': 'menu_master',
    'รายงาน': 'report',
    'รายการธุรกรรม': 'transaction_list',
    'โปรโมชั่น': 'promotion',
    'ผู้ใช้งาน': 'user',
    'สิทธิ์': 'permission',
    'อนุมัติ': 'approve',
    'ไม่อนุมัติ': 'reject',
    'รออนุมัติ': 'pending_approval',
    'ใบสั่งซื้อ': 'transaction_purchase_order',
    'ใบเสนอราคา': 'transaction_quotation',
    'ใบสั่งขาย': 'transaction_sale_order',
    'ชื่อลูกค้า': 'customer_name',
    'ชื่อเจ้าหนี้': 'creditor_name',
    'ราคาขายปลีก': 'price_retail',
    'ราคาขายสมาชิก': 'price_member',
    # PDF document titles
    'ตั้งหนี้จากการทยอยรับ': 'transaction_accrual_receive',
    'ใบสั่งซื้อสินค้า': 'pdf_purchase_order_title',
    'ใบสำคัญรับ': 'pdf_receipt_voucher',
    'ใบสำคัญจ่าย': 'pdf_payment_voucher',
    'ใบกำกับภาษี': 'pdf_tax_invoice',
    'ใบกำกับภาษีอย่างย่อ': 'pdf_short_tax_invoice',
    'ใบส่งสินค้า/ใบกำกับภาษี': 'pdf_delivery_tax_invoice',
    'ใบรับคืนสินค้า/ใบลดหนี้': 'pdf_return_credit_note',
    'ใบส่งสินค้า': 'pdf_delivery_note',
    'ซื้อสินค้า': 'transaction_purchase',
    'ส่งคืนสินค้า': 'pdf_return_goods',
    'ทยอยรับสินค้า': 'pdf_partial_receive',
    'จ่ายเงินล่วงหน้า': 'transaction_advance_payment',
    # GL/Accounting
    'ประเภทเอกสาร': 'document_type',
    'ประเภทภาษี': 'vat_type',
    'ประเภทรายการ': 'transaction_type',
    'ไม่ได้เลือกประเภทเอกสาร': 'gl_no_doc_type_selected',
    'ไม่ได้กำหนดผังบัญชี': 'gl_no_chart_of_account',
    # Report
    'จากทั้งหมด': 'report_of_total',
    'วันที่เริ่มต้น': 'start_date',
    'วันที่สิ้นสุด': 'end_date',
    # Setup/Config
    'ตั้งค่าการแสดงผล': 'display_settings',
    'หน้างาน': 'workspace',
    'เมนูหลัก': 'main_menu',
    'เข้าสู่ระบบ': 'login',
    'เลือกร้าน': 'select_shop',
    'บริษัท': 'company',
    'บาร์โค้ดสินค้า': 'product_barcode',
    'หน่วยสินค้า': 'product_unit',
    'หมวดหมู่สินค้า': 'product_category_list',
    'กลุ่มเจ้าหนี้': 'creditor_group',
    'กลุ่มลูกหนี้': 'debtor_group',
    'ธนาคาร': 'bank',
    'ออกแบบฟอร์ม': 'form_design',
    'ออกแบบบิล': 'bill_design',
    'รูปแบบเอกสาร': 'doc_format',
    'ตั้งค่าระบบ': 'system_setting',
    'ข้อมูลมิติสินค้า': 'product_dimension',
    'BOM สินค้า': 'product_bom',
    'บาร์โค้ดชั้นวาง': 'barcode_shelf',
    'ประวัติราคา': 'price_history',
    'นำเข้าสินค้า': 'import_product',
    'นำเข้าจากไฟล์': 'import_from_file',
    'นำเข้ารูปสินค้า': 'import_product_image',
    'เลือกกลุ่มหมวดหมู่': 'select_category_group',
    'เลือกโซน': 'select_zone',
    'เลือกโต๊ะ': 'select_table',
    'แผนผังโต๊ะ': 'table_map',
    'เลือกครัว': 'select_kitchen',
    'สร้าง QR สั่งอาหาร': 'create_qr_order',
    'ยอดคงเหลือ': 'stock_balance',
    'ซื้อสินค้า': 'transaction_purchase',
    'คืนสินค้าซื้อ': 'transaction_purchasereturn',
    'เงินมัดจำซื้อ': 'transaction_advance_payment',
    'คืนเงินมัดจำซื้อ': 'transaction_advance_payment_refund',
    'เงินประกันซื้อ': 'transaction_deposit',
    'คืนเงินประกันซื้อ': 'transaction_deposit_refund',
    'ซื้อบางส่วน': 'transaction_purchase_partial',
    'รับสินค้าล่วงหน้า': 'transaction_accrual_receive',
    'ขายสินค้า': 'transaction_sale',
    'คืนสินค้าขาย': 'transaction_salereturn',
    'รับเงินล่วงหน้า': 'transaction_paid_advance',
    'คืนเงินล่วงหน้า': 'transaction_paid_advance_refund',
    'รับเงินประกัน': 'transaction_receive_deposit',
    'คืนเงินประกัน': 'transaction_receive_deposit_refund',
    'โอนสินค้า': 'transaction_stock_transfer',
    'รับสินค้า': 'transaction_stock_receive_product',
    'เบิกสินค้า': 'transaction_stock_pick_up_product',
    'คืนสินค้า': 'transaction_stock_return_product',
    'ปรับสต๊อก': 'transaction_adjust',
    'จ่ายเงิน': 'transaction_paid',
    'รับเงิน': 'transaction_pay',
    'รายงานความเคลื่อนไหว': 'report_movement',
    'รายงาน PDF': 'report_pdf',
    'คงเหลือสินค้า': 'report_stock_balance',
    'เคลื่อนไหวสต๊อก': 'report_stock_movement',
    'ประมวลผลบัญชี': 'gl_process',
    'ข้อมูลรายวัน': 'daily_info',
    'การแจ้งเตือน': 'notification',
    'เงินสด': 'cash',
    'ตั้งค่าแต้ม': 'point_setting',
    'คูปอง': 'coupon',
    'ประเภทการจัดซื้อ': 'purchase_type',
    'อนุมัติใบสั่งซื้อ': 'po_approval_setting',
    'ยี่ห้อ': 'brand',
    'ชั้น': 'class',
    'ดีไซน์': 'design',
    'เกรด': 'grade',
    'รุ่น': 'model',
    'ลาย': 'pattern',
    'กลุ่ม': 'group',
    'กลุ่มย่อย 1': 'sub_group_1',
    'กลุ่มย่อย 2': 'sub_group_2',
    'วันทำงาน': 'work_day',
    'วันหยุด': 'holiday',
    'สี': 'color',
    'ประเภทธุรกิจ': 'business_type',
    'บัญชีธนาคาร': 'book_bank',
    'ช่องทางขาย': 'sale_channel',
    'ช่องทางขนส่ง': 'transport_channel',
    'ตำแหน่งสินค้า': 'product_location',
    'ประเภทคำสั่ง': 'order_type',
    'ประเภทสินค้า': 'product_type',
    'เพิ่มสินค้าสาขา': 'add_product_to_branch',
    'เพิ่มสินค้าแผนก': 'add_product_to_department',
    'เพิ่มสินค้าครัว': 'add_product_to_kitchen',
    # Print labels
    'พิมพ์รายการสินค้า': 'print_product_list',
    'ตั้งค่าการพิมพ์': 'print_settings',
    'กำลังพิมพ์...': 'printing',
    'แสดงรูปภาพ': 'show_image',
    'แสดงลำดับรายการ': 'show_item_sequence',
    'พิมพ์ Label สินค้า': 'print_product_label',
    'พิมพ์ป้ายชั้นวางสินค้า': 'print_shelf_label',
    'เลือกรูปแบบการพิมพ์': 'select_print_format',
    'เลือกสีข้อความราคาและกรอบ': 'select_price_text_color',
    'กรุณาเลือกสินค้าก่อน': 'please_select_product_first',
    # Report labels
    'ยอดขายรายวัน': 'report_daily_sales',
    'รับจ่ายรายวัน': 'report_daily_payment',
    'กำไรขั้นต้นตามเอกสาร': 'report_gross_profit_by_doc',
    'กำไรขั้นต้นตามสินค้า': 'report_gross_profit_by_product',
    'รายงานภาษีขาย': 'report_vat_sale',
    'รายงานภาษีซื้อ': 'report_vat_buy',
    'คงเหลือตามคลัง': 'report_stock_by_warehouse',
    'คงเหลือคลัง': 'report_warehouse_balance',
    'คงเหลือตามตำแหน่ง': 'report_stock_by_location',
    'ต้นทุนความเคลื่อนไหว': 'report_stock_movement_cost',
    'ตรวจนับสินค้า': 'audit_product',
    # Chatbot/AI
    'กรุณากรอกข้อความ': 'please_enter_text',
    # Slip
    'SLIP เงินเข้า': 'slip_money_in',
    'SLIP เงินออก': 'slip_money_out',
    # Common UI
    'รหัสผ่านและรหัสยืนยันไม่ตรงกัน': 'password_mismatch',
    'ทดสอบประมวลผลใหม่': 'config_rebuild_stock_test',
    'ตั้งค่า POS': 'pos_setting',
    'ตั้งค่าคำสั่งซื้อ': 'order_setting',
    'เทมเพลตคำสั่งซื้อ': 'order_template_setting',
    'สื่อ POS': 'pos_media',
    # === ROUND 2: Remaining actionable strings ===
    # Display/Report options
    'ตัวเลือกการแสดงผล': 'display_options',
    'ไม่ได้กำหนด': 'not_specified',
    'ค้นหาข้อความ': 'search_text',
    'พิมพ์เอกสารเสร็จสมบูรณ์': 'print_document_complete',
    'ไม่ต้องอนุมัติ': 'no_approval_required',
    'เพิ่มกฎ': 'add_rule',
    'แสดงเอกสารยกเลิก': 'show_cancelled_documents',
    'ไม่ระบุชื่อสินค้า': 'product_name_not_specified',
    'จำนวนรวม': 'total_qty',
    'อ้างอิงแล้ว': 'referenced',
    'POS เท่านั้น': 'pos_only',
    'ต้นทุนรวม': 'total_cost',
    'แสดงผล': 'display',
    'จำนวนคงเหลือ(คำอ่าน)': 'balance_qty_readable',
    'ประมวลผลรายงาน': 'process_report',
    'เตรียมข้อมูล...': 'preparing_data',
    # PDF labels round 2
    'มูลค่ารวมภาษีมูลค่าเพิ่ม': 'pdf_total_incl_vat',
    'รวมมูลค่าหลังหักส่วนลด': 'pdf_total_after_discount',
    'รวมมูลค่ายกเว้นภาษี': 'pdf_total_tax_exempt',
    'มูลค่าหลังหักส่วนลด': 'pdf_value_after_discount',
    'หัก ส่วนลดท้ายบิล': 'pdf_less_bill_discount',
    'อ้างอิงใบกำกับภาษีเดิม': 'pdf_ref_original_tax_invoice',
    'วันครบกำหนด': 'pdf_due_date',
    'มูลค่าใบกำกับภาษีเดิม': 'pdf_original_tax_invoice_value',
    'อ้างอิง': 'reference',
    'ผู้จำหน่าย': 'supplier',
    'ใบเสร็จรับเงิน': 'pdf_receipt',
    # PDF sign labels
    'ผู้อนุมัติ.......................................................': 'pdf_approver_sign',
    'ผู้อนุมัติ................................................................': 'pdf_approver_sign_long',
    'ผู้ส่งสินค้า.......................................................': 'pdf_sender_sign',
    'ผู้รับสินค้า.......................................................': 'pdf_receiver_goods_sign',
    'วันที่........................................': 'pdf_date_sign_medium',
    'วันที่...................................': 'pdf_date_sign_medium2',
    'วันที่ ........................': 'pdf_date_sign_short2',
    # Print label dialog
    'พิมพ์ฉลากติดชั้นวาง (A4)': 'print_shelf_label_a4',
    'สีดำ': 'color_black',
    'รหัสสินค้าไม่ระบุ': 'product_code_not_specified',
    # Approval screens
    'หัวหน้าแผนก': 'dept_head',
    'กรรมการผู้จัดการ': 'managing_director',
    'ผู้จัดการ': 'manager',
    'ผู้บริหาร': 'executive',
    'ระดับ 1': 'level_1',
    'ระดับ 2': 'level_2',
    'ระดับ 3': 'level_3',
    'ชื่อประเภท': 'type_name',
    'คำอธิบาย (ไม่บังคับ)': 'description_optional',
    'กรุณากรอกรหัสและชื่อประเภท': 'please_enter_code_and_type_name',
    'ลบประเภทนี้': 'delete_this_type',
    'กำหนดเงื่อนไขการอนุมัติตามวงเงิน': 'set_approval_conditions_by_amount',
    'คำอธิบายระดับผู้อนุมัติ:': 'approval_level_description',
    'ยังไม่มีกฎการอนุมัติ': 'no_approval_rules_yet',
    'ลบกฎนี้': 'delete_this_rule',
    'วงเงินขั้นต่ำ (บาท)': 'min_amount_baht',
    'วงเงินสูงสุด (0=ไม่จำกัด)': 'max_amount_unlimited',
    'ระดับผู้อนุมัติ': 'approval_level',
    'ผู้มีสิทธิ์อนุมัติ (ใครก็ได้ในรายการ)': 'approvers_anyone_in_list',
    'ยังไม่ได้เลือกผู้อนุมัติ (จะใช้ระดับสิทธิ์ตรวจสอบแทน)': 'no_approver_selected_use_permission',
    'ไม่มี email/LINE': 'no_email_line',
    'เลือกผู้มีสิทธิ์อนุมัติ': 'select_approvers',
    'ค้นหา ชื่อ, อีเมล, LINE...': 'search_name_email_line',
    'ไม่พบผู้ใช้': 'user_not_found',
    'ไม่พบผู้ใช้ที่ค้นหา': 'user_search_not_found',
    'กรอกข้อมูลด้านล่างเพื่อเพิ่มประเภทใหม่': 'fill_data_to_add_new_type',
    'กรุณากรอกรหัส': 'please_enter_code',
    'คำอธิบายเพิ่มเติม (ไม่บังคับ)': 'additional_description_optional',
    'กรุณากรอกชื่อ': 'please_enter_name',
    '• ระดับ 1: อนุมัติอัตโนมัติทันทีเมื่อบันทึกเอกสาร': 'approval_level_1_desc',
    '• ระดับ 2: รออนุมัติจาก ระดับ 1 ก่อน แล้วจึงอนุมัติได้': 'approval_level_2_desc',
    '• ระดับ 3: รออนุมัติจาก ระดับ 1,2 ก่อน แล้วจึงอนุมัติได้': 'approval_level_3_desc',
    '• สูงสุด: ผู้มีอำนาจอนุมัติสูงสุด (เช่น กรรมการผู้จัดการ)': 'approval_level_max_desc',
    'เช่น GENERAL, SPECIAL, VIP': 'eg_general_special_vip',
    'เพิ่มประเภทการจัดซื้อ': 'add_purchase_type',
    'เพิ่มประเภทการจัดซื้อใหม่': 'add_new_purchase_type',
    'เพิ่มประเภทใบเสนอราคา': 'add_quotation_type',
    'เพิ่มประเภทใบเสนอราคาใหม่': 'add_new_quotation_type',
    'เพิ่มประเภทใบสั่งขาย': 'add_sale_order_type',
    'เพิ่มประเภทใบสั่งขายใหม่': 'add_new_sale_order_type',
    'เช่น เสนอราคาทั่วไป, เสนอราคาพิเศษ': 'eg_general_special_quotation',
    'เช่น สั่งขายทั่วไป, สั่งขายพิเศษ': 'eg_general_special_sale_order',
    # Document status
    'ร่าง': 'draft',
    'รับครบแล้ว': 'fully_received',
    'เปิดอ่าน': 'open_read',
    'ยังไม่ส่งขออนุมัติ': 'not_yet_submitted',
    'ยังรับไม่ครบ': 'partially_received',
    'ลบแล้ว': 'deleted',
    'ตัวกรองขั้นสูง': 'advanced_filter',
    'เปิดเอกสาร': 'open_document',
    'ปิดเอกสาร': 'close_document',
    'ส่งขออนุมัติ': 'submit_for_approval',
    'ถอนการอนุมัติ': 'withdraw_approval',
    'ไม่พบข้อมูลที่ตรงกับเงื่อนไข': 'no_data_matching_criteria',
    # Cash/Shift
    'เปิดกะ': 'open_shift',
    'ปิดกะ': 'close_shift',
    'ถอนเงิน': 'withdraw_cash',
    'เพิ่มเงินในลิ้นชัก': 'add_cash_to_drawer',
    # Permission screens
    'เพิ่มสิทธิ์ใหม่': 'add_new_permission',
    'การเงิน': 'finance',
    'เชื่อมสิทธิ์': 'link_permission',
    'กรุณากรอกรหัสสิทธิ์': 'please_enter_permission_code',
    'แก้ไขสิทธิ์': 'edit_permission',
    'ผู้ใช้และพนักงาน': 'users_and_employees',
    'ลูกค้าและผู้ขาย': 'customers_and_sellers',
    'ตั้งค่าขาย': 'sale_settings',
    'ธุรกรรมซื้อ': 'purchase_transactions',
    'ธุรกรรมขาย': 'sale_transactions',
    'สต๊อก': 'stock',
    'นำเข้า/ส่งออก': 'import_export',
    'AI & ระบบอัจฉริยะ': 'ai_smart_system',
    'ยกเลิกทั้งหมด': 'cancel_all',
    'หน้าจอ': 'screen',
    'แก้ตัวเอง': 'edit_own',
    'แก้ทั้งหมด': 'edit_all',
    'ลบตัวเอง': 'delete_own',
    # Report screens
    'กำไรขั้นต้น': 'gross_profit',
    'ยอดขายรวม': 'total_sales',
    'ไม่แสดงเอกสารยกเลิก': 'hide_cancelled_documents',
    'ไม่พบข้อมูลรายงาน': 'no_report_data',
    'กรุณาปรับเปลี่ยนเงื่อนไขการค้นหา': 'please_adjust_search_criteria',
    'วันที่รายงาน': 'report_date',
    'ประมวลผลสำเร็จ': 'process_success',
    'ซ่อนคำแนะนำ': 'hide_tips',
    'การใช้งาน PDF': 'pdf_usage',
    'บันทึกเป็นไฟล์ PDF': 'save_as_pdf',
    'พิมพ์รายงาน': 'print_report',
    'ยืนยันการออกจากรายงาน': 'confirm_exit_report',
    'คุณต้องการออกจากรายงานนี้หรือไม่?': 'confirm_exit_report_question',
    'กำลังประมวลผล...': 'processing',
    'ปิดหน้ารายงาน': 'close_report_page',
    'แสดง/ซ่อน คำอธิบายการใช้งาน': 'show_hide_usage_guide',
    'กำลังโหลด PDF กรุณารอสักครู่...': 'loading_pdf_please_wait',
    'พิมพ์คำที่ต้องการค้นหาในเอกสาร': 'type_search_text_in_document',
    'เลือกสินค้าที่ต้องการแสดง หรือไม่เลือกเพื่อแสดงทั้งหมด': 'select_products_or_show_all',
    'เลือกแสดงเฉพาะสินค้าที่มียอดคงเหลือหรือแสดงทั้งหมด': 'show_only_with_balance_or_all',
    'ยอดคงเหลือ ณ. วันที่': 'balance_as_of_date',
    'เลือกวันที่ที่ต้องการแสดงยอด': 'select_date_for_balance',
    'กำหนดรูปแบบการแสดงผลรายงานตามต้องการ': 'customize_report_display',
    'แสดงเฉพาะสินค้าที่มียอดคงเหลือ': 'show_only_with_balance',
    'เลือกสินค้าที่ต้องการแสดงในรายงาน หากไม่เลือก ระบบจะแสดงสินค้าทั้งหมด': 'select_products_for_report_or_show_all',
    'ไม่มีสินค้าที่เลือก ระบบจะแสดงสินค้าทั้งหมด': 'no_products_selected_show_all',
    'เลือกคลังสินค้าที่ต้องการแสดงในรายงาน หากไม่เลือก ระบบจะแสดงทุกคลังสินค้า': 'select_warehouse_for_report_or_show_all',
    '* ใช้ปุ่ม + หรือ - เพื่อปรับขนาดตัวอักษร และเลื่อนลงเพื่อดูข้อมูลเพิ่มเติม': 'report_zoom_tip',
    'ในแท็บนี้ คุณสามารถค้นหาข้อความในเอกสาร บันทึก PDF ไว้ในอุปกรณ์ หรือสั่งพิมพ์รายงานได้': 'report_pdf_tab_desc',
    # MCP/Chatbot labels
    'ลูกหนี้การค้า': 'accounts_receivable',
    'เจ้าหนี้การค้า': 'accounts_payable',
    'กระแสเงินสด': 'cash_flow',
    'มูลค่าสินค้าคงเหลือ': 'inventory_value',
    'อัตราหมุนเวียนสินค้า': 'inventory_turnover',
    'สินค้าขายดี': 'best_selling_products',
    'บ่อยเกินไป': 'too_frequent',
    # Other common labels
    'จ่ายล่วงหน้า': 'pay_advance',
    'รับล่วงหน้า': 'receive_advance',
    'รับมัดจำ': 'receive_deposit',
    'ยอดคงเหลือตามสินค้า': 'balance_by_product',
    'ยอดคงเหลือตามคลัง': 'balance_by_warehouse',
    'รับสินค้าบางส่วน': 'partial_receive',
    'เลือกตำแหน่งที่เก็บ': 'select_storage_location',
    'เลือกชั้นวาง': 'select_shelf',
    'เกิดข้อผิดพลาดในการบันทึก': 'error_saving',
    'เกิดข้อผิดพลาดในการดาวน์โหลดไฟล์': 'error_downloading_file',
    'เกิดข้อผิดพลาดในการดึงข้อมูลรายงาน': 'error_fetching_report',
    'เกิดข้อผิดพลาดในการดึงข้อมูลสรุปรายงาน': 'error_fetching_report_summary',
    'เพิ่มแต้มสำเร็จ': 'add_point_success',
    'กรุณากรอกชื่อร้านย่อย': 'please_enter_sub_shop_name',
    'กำลังลบรายการ...': 'deleting_items',
    'ใบรับจอง': 'booking_receipt',
    'เชื่อมต่อแล้ว': 'connected',
    'กำลังทดสอบ...': 'testing',
    'ทดสอบการเชื่อมต่อ': 'test_connection',
    'ลองตัวนี้ก่อน': 'try_this_first',
    'พิมพ์ รหัส/ชื่อเจ้าหนี้ เพื่อค้นหา': 'type_creditor_code_name_to_search',
    # LINE Login QR
    'เข้าสู่ระบบสำเร็จ!': 'login_success',
    'กดยืนยันการเข้าสู่ระบบ': 'confirm_login',
    'ระบบจะเข้าสู่ระบบอัตโนมัติ': 'auto_login',
    'รอการยืนยัน...': 'waiting_confirmation',
    'รหัสหมดอายุใน 5 นาที': 'code_expires_in_5_min',
    'ไม่พบข้อมูลการอนุมัติสำหรับเอกสารนี้': 'no_approval_data_for_document',
    # Form design/config
    'สำนักงานใหญ่': 'head_office',
    # Misc validation/error
    'รหัสผู้ใช้ไม่สามารถเป็นค่าว่างได้': 'user_code_cannot_be_empty',
    'POS ID ไม่สามารถเป็นค่าว่างได้': 'pos_id_cannot_be_empty',
    'เลขที่เอกสารไม่สามารถเป็นค่าว่างได้': 'doc_no_cannot_be_empty',
    # Report hints — REMOVED fragment entries that break strings with \"
    # POS/Config labels
    'พิมพ์ Label สินค้า': 'print_product_label',
    'เลือกรูปแบบการพิมพ์': 'select_print_format',
    'เลือกสีข้อความราคาและกรอบ': 'select_price_text_color',
    'กรุณาเลือกสินค้าก่อน': 'please_select_product_first',
    'แสดงลำดับรายการ': 'show_item_sequence',
    # Report instructions with escaped quotes (full strings)
    'ยังไม่ได้ประมวลผลรายงาน กรุณากลับไปที่แท็บ \\"เงื่อนไข\\" และกดปุ่ม \\"ประมวลผล\\"': 'report_not_processed_hint',
    'กลับไปที่แท็บ \\"เงื่อนไข\\" และกดปุ่ม \\"ประมวลผลรายงาน\\"': 'report_go_to_condition_tab',
    # === ROUND 3: Remaining 2x+ strings ===
    # PO/QT/SO document messages
    'กรุณาเลือกประเภทการจัดซื้อ': 'please_select_purchase_type',
    'เอกสารอยู่ระหว่างรออนุมัติ ไม่สามารถแก้ไขได้': 'doc_pending_approval_cannot_edit',
    'กรุณาเพิ่มรายการสินค้า': 'please_add_product_items',
    'ยอดรวมต้องมากกว่า 0': 'total_must_be_greater_than_zero',
    'กรุณาบันทึกเอกสารก่อนส่งอนุมัติ': 'please_save_before_submit',
    'เอกสารถูกอนุมัติบางส่วนแล้ว ไม่สามารถส่งอนุมัติใหม่ได้': 'doc_partially_approved_cannot_resubmit',
    'เอกสารอนุมัติแล้ว': 'document_approved',
    'เอกสารถูกยกเลิกแล้ว': 'document_cancelled',
    'เอกสารถูกอ้างอิงแล้ว ไม่สามารถยกเลิกได้': 'doc_referenced_cannot_cancel',
    'เอกสารอยู่ระหว่างรออนุมัติ กรุณาถอนการส่งอนุมัติก่อน': 'doc_pending_withdraw_first',
    'สร้าง PDF': 'create_pdf',
    'รายละเอียดสินค้า': 'product_detail',
    'สินค้าทั้งหมด': 'all_products',
    'มูลค่าขาย': 'sale_value',
    'ต้นทุนขาย': 'cost_of_sales',
    'กำไร(ขาดทุน)': 'profit_loss',
    '%กำไรขั้นต้น': 'gross_profit_pct',
    'เลือกพนักงานขาย': 'select_salesperson',
    'เข้าสู่ระบบด้วย LINE': 'login_with_line',
    'กำลังเข้าสู่ระบบ...': 'logging_in',
    'พิมพ์ รหัส/ชื่อเจ้าหนี้ เพื่อค้นหา': 'type_creditor_code_name_to_search',
    # 1x but actionable strings
    # Round 3b - more 2x strings
    'ตัวกรองข้อมูล': 'data_filter',
    'หลังบ้าน': 'back_office',
    'ไม่มีข้อมูลยอดรวม': 'no_total_data',
    'กำไรขั้นต้นรวม': 'total_gross_profit',
    'เปอร์เซ็นต์กำไรเฉลี่ย': 'avg_profit_pct',
    'ฐานภาษีรวม': 'total_tax_base',
    'ยกเว้นภาษีรวม': 'total_tax_exempt',
    'เลขประจำตัวผู้เสียภาษี': 'tax_id_number',
    'วิธีการใช้งานรายงานสินค้าคงเหลือ': 'stock_report_usage_guide',
    'เลือกประเภทรายงาน': 'select_report_type',
    'กรุณาลองประมวลผลรายงานใหม่อีกครั้ง': 'please_try_process_report_again',
    'รหัสคลังสินค้า': 'warehouse_code',
    'ไม่พบข้อมูลตามเงื่อนไขที่เลือก': 'no_data_for_selected_criteria',
    'ยังไม่ได้ประมวลผลรายงาน': 'report_not_processed',
    'ไปยังหน้าเงื่อนไข': 'go_to_condition_page',
    'เลือกคลังสินค้าที่ต้องการแสดง หรือไม่เลือกเพื่อแสดงทั้งหมด': 'select_warehouse_or_show_all',
    'ดูผลลัพธ์': 'view_results',
    # Round 3c
    'รายงาน Process Stock Cost': 'report_process_stock_cost',
    'รายงานเคลื่อนไหวสินค้า/ต้นทุน': 'report_stock_movement_cost_title',
    'แตะเพื่อเลือกเอกสาร': 'tap_to_select_document',
    'เพิ่มหมายเหตุ...': 'add_remark',
    'ไม่สามารถระบุผู้สร้างเอกสารได้': 'cannot_identify_doc_creator',
    'เอกสารอื่น': 'other_document',
    'สร้างรหัสใหม่': 'create_new_code',
    'คัดลอกลิงก์แล้ว': 'link_copied',
    'คัดลอกลิงก์': 'copy_link',
    'รองกรรมการผู้จัดการ': 'vice_managing_director',
    'ผู้จัดการฝ่าย': 'department_manager',
    'ผู้อำนวยการ': 'director',
    # Price channels
    'ราคาตามช่องทาง 1': 'price_channel_1',
    'ราคาตามช่องทาง 2': 'price_channel_2',
    'ราคาตามช่องทาง 3': 'price_channel_3',
    'ราคาตามช่องทาง 4': 'price_channel_4',
    'ราคาตามช่องทาง 5': 'price_channel_5',
    'ปิดเมนู': 'close_menu',
    'เปิดเมนู': 'open_menu',
    'ยืนยันการลบ': 'confirm_delete',
    'คุณต้องการลบรายการนี้?': 'confirm_delete_question',
    'ต้องการลบรายการนี้?': 'confirm_delete_item_question',
    'เพิ่มรายการ': 'add_item',
    'เลือกเงื่อนไข': 'select_condition',
    'เงื่อนไข': 'condition',
    'ดาวน์โหลด': 'download',
    'กรุณาเลือกประเภทใบเสนอราคา': 'please_select_quotation_type',
    'กรุณาเลือกเจ้าหนี้': 'please_select_creditor',
    'กรุณาเลือกลูกค้า': 'please_select_customer',
    'ต้องการยกเลิกเอกสารนี้?': 'confirm_cancel_document',
    'ยกเลิกเอกสาร': 'cancel_document',
    'รูปแบบ': 'format',
    'กรุณาเลือกคลังสินค้า': 'please_select_warehouse',
    'กรุณาเลือกเอกสาร': 'please_select_document',
    'เลือกรูปแบบ': 'select_format',
    'สร้างเอกสาร': 'create_document',
    'บันทึกร่าง': 'save_draft',
    'กรุณาเลือกประเภทใบสั่งขาย': 'please_select_sale_order_type',
    'เลือกลูกค้า': 'select_customer',
    'เลือกเจ้าหนี้': 'select_creditor',
    'ข้อมูลลูกค้า': 'customer_info',
    'ข้อมูลเจ้าหนี้': 'creditor_info',
    'จำนวนสินค้า': 'product_qty',
    'ยอดรวม': 'total_amount',
    'มูลค่ารวม': 'total_value',
    'ส่วนลดรวม': 'total_discount',
    'ภาษีมูลค่าเพิ่ม': 'vat_amount',
    'ยอดสุทธิ': 'net_amount',
    'คำอธิบาย': 'description',
    'เลขที่อ้างอิง': 'reference_no',
    'ผลรวม': 'sum_total',
    'จ่ายชำระ': 'make_payment',
    'รับชำระ': 'receive_payment',
    'เลือกประเภท': 'select_type',
    'กดเพื่อเลือก': 'tap_to_select',
    'กดเพิ่ม': 'tap_to_add',
    'ไม่มีสิทธิ์': 'no_permission',
    'ไม่มีสิทธิ์เข้าถึง': 'no_access_permission',
    'พิมพ์เอกสาร': 'print_document',
    'แชร์เอกสาร': 'share_document',
    'ดูตัวอย่าง': 'preview',
    'ส่งอีเมล': 'send_email',
    'คัดลอก': 'copy',
    'วาง': 'paste',
    'แนบไฟล์': 'attach_file',
    'เลือกไฟล์': 'select_file',
    'อัพโหลด': 'upload',
    'ดาวน์โหลดไฟล์': 'download_file',
    'เลือกรูปภาพ': 'select_image',
    'ถ่ายรูป': 'take_photo',
    'แสกนบาร์โค้ด': 'scan_barcode',
    'ดูทั้งหมด': 'view_all',
    'ไม่ระบุ': 'not_specified',
    'อื่นๆ': 'others',
    'ดำเนินการ': 'proceed',
    'เริ่มใหม่': 'start_over',
    'ย้อนกลับ': 'go_back',
    'ถัดไป': 'next',
    'ก่อนหน้า': 'previous',
    'เสร็จสิ้น': 'finished',
}

# === English translations for new keys ===
NEW_KEY_TRANSLATIONS = {
    'pdf_payer_sign': 'Payer',
    'pdf_receiver_sign': 'Receiver',
    'pdf_date_sign': 'Date',
    'pdf_date_sign_short': 'Date',
    'pdf_page': 'Page',
    'pdf_vat_date': 'Tax Invoice Date',
    'pdf_tax_id_number': 'Tax ID Number',
    'pdf_discount': 'Less: Discount',
    'pdf_total_before_vat': 'Total Before VAT',
    'pdf_tax_exempt_goods': 'Tax Exempt Goods',
    'pdf_taxable_goods': 'Taxable Goods',
    'pdf_grand_total': 'Grand Total',
    'pdf_goods_value': 'Goods Value',
    'pdf_price_per_unit': 'Price/Unit',
    'pdf_total_price': 'Total Price',
    'pdf_sequence': 'No.',
    'pdf_item': 'Item',
    'pdf_purchase_order_title': 'Purchase Order',
    'pdf_receipt_voucher': 'Receipt Voucher',
    'pdf_payment_voucher': 'Payment Voucher',
    'pdf_tax_invoice': 'Tax Invoice',
    'pdf_short_tax_invoice': 'Abbreviated Tax Invoice',
    'pdf_delivery_tax_invoice': 'Delivery Note/Tax Invoice',
    'pdf_return_credit_note': 'Return/Credit Note',
    'pdf_delivery_note': 'Delivery Note',
    'pdf_return_goods': 'Return Goods',
    'pdf_partial_receive': 'Partial Receive',
    'gl_no_doc_type_selected': 'No document type selected',
    'gl_no_chart_of_account': 'Chart of account not set',
    'report_of_total': 'of total',
    'start_date': 'Start Date',
    'end_date': 'End Date',
    'display_settings': 'Display Settings',
    'workspace': 'Workspace',
    'main_menu': 'Main Menu',
    'login': 'Login',
    'select_shop': 'Select Shop',
    'product_barcode': 'Product Barcode',
    'product_unit': 'Product Unit',
    'product_category_list': 'Product Category',
    'creditor_group': 'Creditor Group',
    'debtor_group': 'Debtor Group',
    'form_design': 'Form Design',
    'bill_design': 'Bill Design',
    'doc_format': 'Document Format',
    'system_setting': 'System Setting',
    'product_dimension': 'Product Dimension',
    'product_bom': 'Product BOM',
    'barcode_shelf': 'Shelf Barcode',
    'price_history': 'Price History',
    'import_product': 'Import Product',
    'import_from_file': 'Import from File',
    'import_product_image': 'Import Product Image',
    'select_category_group': 'Select Category Group',
    'select_zone': 'Select Zone',
    'select_table': 'Select Table',
    'table_map': 'Table Map',
    'select_kitchen': 'Select Kitchen',
    'create_qr_order': 'Create QR Order',
    'stock_balance': 'Stock Balance',
    'report_movement': 'Movement Report',
    'report_pdf': 'PDF Report',
    'report_stock_balance': 'Stock Balance Report',
    'report_stock_movement': 'Stock Movement',
    'gl_process': 'GL Process',
    'daily_info': 'Daily Info',
    'notification': 'Notification',
    'point_setting': 'Point Setting',
    'purchase_type': 'Purchase Type',
    'po_approval_setting': 'PO Approval Setting',
    'sub_group_1': 'Sub Group 1',
    'sub_group_2': 'Sub Group 2',
    'work_day': 'Work Day',
    'holiday': 'Holiday',
    'business_type': 'Business Type',
    'book_bank': 'Bank Account',
    'sale_channel': 'Sale Channel',
    'transport_channel': 'Transport Channel',
    'product_location': 'Product Location',
    'order_type': 'Order Type',
    'product_type': 'Product Type',
    'add_product_to_branch': 'Add Product to Branch',
    'add_product_to_department': 'Add Product to Department',
    'add_product_to_kitchen': 'Add Product to Kitchen',
    'print_product_list': 'Print Product List',
    'print_settings': 'Print Settings',
    'printing': 'Printing...',
    'show_image': 'Show Image',
    'show_item_sequence': 'Show Item Sequence',
    'print_product_label': 'Print Product Label',
    'print_shelf_label': 'Print Shelf Label',
    'select_print_format': 'Select Print Format',
    'select_price_text_color': 'Select Price Text Color',
    'please_select_product_first': 'Please select product first',
    'report_daily_sales': 'Daily Sales Report',
    'report_daily_payment': 'Daily Payment Report',
    'report_gross_profit_by_doc': 'Gross Profit by Document',
    'report_gross_profit_by_product': 'Gross Profit by Product',
    'report_vat_sale': 'Sales VAT Report',
    'report_vat_buy': 'Purchase VAT Report',
    'report_stock_by_warehouse': 'Stock by Warehouse',
    'report_warehouse_balance': 'Warehouse Balance',
    'report_stock_by_location': 'Stock by Location',
    'report_stock_movement_cost': 'Stock Movement Cost',
    'audit_product': 'Product Audit',
    'please_enter_text': 'Please enter text',
    'slip_money_in': 'Money In Slip',
    'slip_money_out': 'Money Out Slip',
    'setting': 'Settings',
    'reject': 'Reject',
    'buyer': 'Buyer',
    'cost': 'Cost',
    'sale_person': 'Salesperson',
    'price_member': 'Member Price',
    'net': 'Net',
    'cash': 'Cash',
    'open': 'Open',
    'activate': 'Activate',
    'deactivate': 'Deactivate',
    'no_data': 'No Data',
    'failed': 'Failed',
    'success': 'Success',
    'password_mismatch': 'Password and confirm password do not match',
    'config_rebuild_stock_test': 'Test Reprocess',
    'pos_setting': 'POS Setting',
    'order_setting': 'Order Setting',
    'order_template_setting': 'Order Template Setting',
    'pos_media': 'POS Media',
    # === ROUND 2 translations ===
    'display_options': 'Display Options',
    'not_specified': 'Not Specified',
    'search_text': 'Search Text',
    'print_document_complete': 'Print Document Complete',
    'no_approval_required': 'No Approval Required',
    'add_rule': 'Add Rule',
    'show_cancelled_documents': 'Show Cancelled Documents',
    'product_name_not_specified': 'Product Name Not Specified',
    'total_qty': 'Total Quantity',
    'referenced': 'Referenced',
    'pos_only': 'POS Only',
    'total_cost': 'Total Cost',
    'display': 'Display',
    'balance_qty_readable': 'Balance Quantity (Readable)',
    'process_report': 'Process Report',
    'preparing_data': 'Preparing data...',
    'pdf_total_incl_vat': 'Total Including VAT',
    'pdf_total_after_discount': 'Total After Discount',
    'pdf_total_tax_exempt': 'Total Tax Exempt',
    'pdf_value_after_discount': 'Value After Discount',
    'pdf_less_bill_discount': 'Less: Bill Discount',
    'pdf_ref_original_tax_invoice': 'Ref. Original Tax Invoice',
    'pdf_due_date': 'Due Date',
    'pdf_original_tax_invoice_value': 'Original Tax Invoice Value',
    'reference': 'Reference',
    'supplier': 'Supplier',
    'pdf_receipt': 'Receipt',
    'pdf_approver_sign': 'Approver',
    'pdf_approver_sign_long': 'Approver',
    'pdf_sender_sign': 'Sender',
    'pdf_receiver_goods_sign': 'Goods Receiver',
    'pdf_date_sign_medium': 'Date',
    'pdf_date_sign_medium2': 'Date',
    'pdf_date_sign_short2': 'Date',
    'print_shelf_label_a4': 'Print Shelf Label (A4)',
    'color_black': 'Black',
    'product_code_not_specified': 'Product Code Not Specified',
    'dept_head': 'Department Head',
    'managing_director': 'Managing Director',
    'manager': 'Manager',
    'executive': 'Executive',
    'level_1': 'Level 1',
    'level_2': 'Level 2',
    'level_3': 'Level 3',
    'type_name': 'Type Name',
    'description_optional': 'Description (Optional)',
    'please_enter_code_and_type_name': 'Please enter code and type name',
    'delete_this_type': 'Delete This Type',
    'set_approval_conditions_by_amount': 'Set approval conditions by amount',
    'approval_level_description': 'Approval Level Description:',
    'no_approval_rules_yet': 'No approval rules yet',
    'delete_this_rule': 'Delete This Rule',
    'min_amount_baht': 'Minimum Amount (Baht)',
    'max_amount_unlimited': 'Maximum Amount (0=Unlimited)',
    'approval_level': 'Approval Level',
    'approvers_anyone_in_list': 'Approvers (Anyone in List)',
    'no_approver_selected_use_permission': 'No approver selected (will use permission level instead)',
    'no_email_line': 'No email/LINE',
    'select_approvers': 'Select Approvers',
    'search_name_email_line': 'Search name, email, LINE...',
    'user_not_found': 'User Not Found',
    'user_search_not_found': 'No users found matching search',
    'fill_data_to_add_new_type': 'Fill data below to add new type',
    'please_enter_code': 'Please enter code',
    'additional_description_optional': 'Additional description (Optional)',
    'please_enter_name': 'Please enter name',
    'approval_level_1_desc': 'Level 1: Auto-approve immediately upon document save',
    'approval_level_2_desc': 'Level 2: Wait for Level 1 approval first',
    'approval_level_3_desc': 'Level 3: Wait for Level 1, 2 approval first',
    'approval_level_max_desc': 'Highest: Highest authority (e.g., Managing Director)',
    'eg_general_special_vip': 'e.g., GENERAL, SPECIAL, VIP',
    'add_purchase_type': 'Add Purchase Type',
    'add_new_purchase_type': 'Add New Purchase Type',
    'add_quotation_type': 'Add Quotation Type',
    'add_new_quotation_type': 'Add New Quotation Type',
    'add_sale_order_type': 'Add Sale Order Type',
    'add_new_sale_order_type': 'Add New Sale Order Type',
    'eg_general_special_quotation': 'e.g., General Quotation, Special Quotation',
    'eg_general_special_sale_order': 'e.g., General Sale Order, Special Sale Order',
    'draft': 'Draft',
    'fully_received': 'Fully Received',
    'open_read': 'Open to Read',
    'not_yet_submitted': 'Not Yet Submitted',
    'partially_received': 'Partially Received',
    'deleted': 'Deleted',
    'advanced_filter': 'Advanced Filter',
    'open_document': 'Open Document',
    'close_document': 'Close Document',
    'submit_for_approval': 'Submit for Approval',
    'withdraw_approval': 'Withdraw Approval',
    'no_data_matching_criteria': 'No data matching criteria',
    'open_shift': 'Open Shift',
    'close_shift': 'Close Shift',
    'withdraw_cash': 'Withdraw Cash',
    'add_cash_to_drawer': 'Add Cash to Drawer',
    'add_new_permission': 'Add New Permission',
    'finance': 'Finance',
    'link_permission': 'Link Permission',
    'please_enter_permission_code': 'Please enter permission code',
    'edit_permission': 'Edit Permission',
    'users_and_employees': 'Users & Employees',
    'customers_and_sellers': 'Customers & Sellers',
    'sale_settings': 'Sale Settings',
    'purchase_transactions': 'Purchase Transactions',
    'sale_transactions': 'Sale Transactions',
    'stock': 'Stock',
    'import_export': 'Import/Export',
    'ai_smart_system': 'AI & Smart System',
    'cancel_all': 'Cancel All',
    'screen': 'Screen',
    'edit_own': 'Edit Own',
    'edit_all': 'Edit All',
    'delete_own': 'Delete Own',
    'gross_profit': 'Gross Profit',
    'total_sales': 'Total Sales',
    'hide_cancelled_documents': 'Hide Cancelled Documents',
    'no_report_data': 'No report data',
    'please_adjust_search_criteria': 'Please adjust search criteria',
    'report_date': 'Report Date',
    'process_success': 'Process Success',
    'hide_tips': 'Hide Tips',
    'pdf_usage': 'PDF Usage',
    'save_as_pdf': 'Save as PDF',
    'print_report': 'Print Report',
    'confirm_exit_report': 'Confirm Exit Report',
    'confirm_exit_report_question': 'Do you want to exit this report?',
    'processing': 'Processing...',
    'close_report_page': 'Close Report Page',
    'show_hide_usage_guide': 'Show/Hide Usage Guide',
    'loading_pdf_please_wait': 'Loading PDF, please wait...',
    'type_search_text_in_document': 'Type text to search in document',
    'select_products_or_show_all': 'Select products to display, or leave empty to show all',
    'show_only_with_balance_or_all': 'Show only products with balance or show all',
    'balance_as_of_date': 'Balance as of date',
    'select_date_for_balance': 'Select date for balance display',
    'customize_report_display': 'Customize report display settings',
    'show_only_with_balance': 'Show only products with balance',
    'select_products_for_report_or_show_all': 'Select products for report. If none selected, all products will be shown',
    'no_products_selected_show_all': 'No products selected, all products will be shown',
    'select_warehouse_for_report_or_show_all': 'Select warehouse for report. If none selected, all warehouses will be shown',
    'report_zoom_tip': 'Use + or - to adjust font size, scroll down for more data',
    'report_pdf_tab_desc': 'In this tab, you can search text in document, save PDF, or print report',
    'accounts_receivable': 'Accounts Receivable',
    'accounts_payable': 'Accounts Payable',
    'cash_flow': 'Cash Flow',
    'inventory_value': 'Inventory Value',
    'inventory_turnover': 'Inventory Turnover',
    'best_selling_products': 'Best Selling Products',
    'too_frequent': 'Too Frequent',
    'pay_advance': 'Pay Advance',
    'receive_advance': 'Receive Advance',
    'receive_deposit': 'Receive Deposit',
    'balance_by_product': 'Balance by Product',
    'balance_by_warehouse': 'Balance by Warehouse',
    'partial_receive': 'Partial Receive',
    'select_storage_location': 'Select Storage Location',
    'select_shelf': 'Select Shelf',
    'error_saving': 'Error saving',
    'error_downloading_file': 'Error downloading file',
    'error_fetching_report': 'Error fetching report data',
    'error_fetching_report_summary': 'Error fetching report summary',
    'add_point_success': 'Points added successfully',
    'please_enter_sub_shop_name': 'Please enter sub-shop name',
    'deleting_items': 'Deleting items...',
    'booking_receipt': 'Booking Receipt',
    'connected': 'Connected',
    'testing': 'Testing...',
    'test_connection': 'Test Connection',
    'try_this_first': 'Try this first',
    'type_creditor_code_name_to_search': 'Type creditor code/name to search',
    'login_success': 'Login Successful!',
    'confirm_login': 'Confirm login',
    'auto_login': 'System will auto-login',
    'waiting_confirmation': 'Waiting for confirmation...',
    'code_expires_in_5_min': 'Code expires in 5 minutes',
    'no_approval_data_for_document': 'No approval data for this document',
    'head_office': 'Head Office',
    'user_code_cannot_be_empty': 'User code cannot be empty',
    'pos_id_cannot_be_empty': 'POS ID cannot be empty',
    'doc_no_cannot_be_empty': 'Document number cannot be empty',
    'to_create_report': ' to create report',
    'to_see_results': ' to see results',
    'for_viewing_data_or': ' for viewing data, or',
    'for_print_and_save_report': ' for printing and saving report',
    'report_not_processed_hint': 'Report not processed yet. Please go to the "Condition" tab and press "Process"',
    'report_go_to_condition_tab': 'Go to the "Condition" tab and press "Process Report"',
    # Round 3b translations
    'data_filter': 'Data Filter',
    'back_office': 'Back Office',
    'no_total_data': 'No total data available',
    'total_gross_profit': 'Total Gross Profit',
    'avg_profit_pct': 'Average Profit Percentage',
    'total_tax_base': 'Total Tax Base',
    'total_tax_exempt': 'Total Tax Exempt',
    'tax_id_number': 'Tax ID Number',
    'stock_report_usage_guide': 'Stock Report Usage Guide',
    'select_report_type': 'Select Report Type',
    'please_try_process_report_again': 'Please try processing the report again',
    'warehouse_code': 'Warehouse Code',
    'no_data_for_selected_criteria': 'No data for selected criteria',
    'report_not_processed': 'Report not yet processed',
    'go_to_condition_page': 'Go to Condition Page',
    'select_warehouse_or_show_all': 'Select warehouse to display, or leave empty to show all',
    'view_results': 'View Results',
    # Round 3c translations
    'report_process_stock_cost': 'Stock Cost Process Report',
    'report_stock_movement_cost_title': 'Stock Movement/Cost Report',
    'tap_to_select_document': 'Tap to select document',
    'add_remark': 'Add remark...',
    'cannot_identify_doc_creator': 'Cannot identify document creator',
    'other_document': 'Other Document',
    'create_new_code': 'Create New Code',
    'link_copied': 'Link Copied',
    'copy_link': 'Copy Link',
    'vice_managing_director': 'Vice Managing Director',
    'department_manager': 'Department Manager',
    'director': 'Director',
    'price_channel_1': 'Price Channel 1',
    'price_channel_2': 'Price Channel 2',
    'price_channel_3': 'Price Channel 3',
    'price_channel_4': 'Price Channel 4',
    'price_channel_5': 'Price Channel 5',
    # === ROUND 3 translations ===
    'please_select_purchase_type': 'Please select purchase type',
    'doc_pending_approval_cannot_edit': 'Document is pending approval and cannot be edited',
    'please_add_product_items': 'Please add product items',
    'total_must_be_greater_than_zero': 'Total must be greater than 0',
    'please_save_before_submit': 'Please save document before submitting for approval',
    'doc_partially_approved_cannot_resubmit': 'Document partially approved, cannot resubmit',
    'document_approved': 'Document Approved',
    'document_cancelled': 'Document Cancelled',
    'doc_referenced_cannot_cancel': 'Document is referenced and cannot be cancelled',
    'doc_pending_withdraw_first': 'Document is pending approval, please withdraw first',
    'create_pdf': 'Create PDF',
    'product_detail': 'Product Detail',
    'all_products': 'All Products',
    'sale_value': 'Sale Value',
    'cost_of_sales': 'Cost of Sales',
    'profit_loss': 'Profit (Loss)',
    'gross_profit_pct': '% Gross Profit',
    'select_salesperson': 'Select Salesperson',
    'login_with_line': 'Login with LINE',
    'logging_in': 'Logging in...',
    'close_menu': 'Close Menu',
    'open_menu': 'Open Menu',
    'confirm_delete': 'Confirm Delete',
    'confirm_delete_question': 'Do you want to delete this item?',
    'confirm_delete_item_question': 'Do you want to delete this item?',
    'add_item': 'Add Item',
    'select_condition': 'Select Condition',
    'condition': 'Condition',
    'download': 'Download',
    'please_select_quotation_type': 'Please select quotation type',
    'please_select_creditor': 'Please select creditor',
    'please_select_customer': 'Please select customer',
    'confirm_cancel_document': 'Do you want to cancel this document?',
    'cancel_document': 'Cancel Document',
    'format': 'Format',
    'please_select_warehouse': 'Please select warehouse',
    'please_select_document': 'Please select document',
    'select_format': 'Select Format',
    'create_document': 'Create Document',
    'save_draft': 'Save Draft',
    'please_select_sale_order_type': 'Please select sale order type',
    'select_customer': 'Select Customer',
    'select_creditor': 'Select Creditor',
    'customer_info': 'Customer Info',
    'creditor_info': 'Creditor Info',
    'product_qty': 'Product Quantity',
    'total_amount': 'Total Amount',
    'total_value': 'Total Value',
    'total_discount': 'Total Discount',
    'vat_amount': 'VAT Amount',
    'net_amount': 'Net Amount',
    'description': 'Description',
    'reference_no': 'Reference No.',
    'sum_total': 'Sum Total',
    'make_payment': 'Make Payment',
    'receive_payment': 'Receive Payment',
    'select_type': 'Select Type',
    'tap_to_select': 'Tap to Select',
    'tap_to_add': 'Tap to Add',
    'no_permission': 'No Permission',
    'no_access_permission': 'No Access Permission',
    'print_document': 'Print Document',
    'share_document': 'Share Document',
    'preview': 'Preview',
    'send_email': 'Send Email',
    'copy': 'Copy',
    'paste': 'Paste',
    'attach_file': 'Attach File',
    'select_file': 'Select File',
    'upload': 'Upload',
    'download_file': 'Download File',
    'select_image': 'Select Image',
    'take_photo': 'Take Photo',
    'scan_barcode': 'Scan Barcode',
    'view_all': 'View All',
    'not_specified': 'Not Specified',
    'others': 'Others',
    'proceed': 'Proceed',
    'start_over': 'Start Over',
    'go_back': 'Go Back',
    'next': 'Next',
    'previous': 'Previous',
    'finished': 'Finished',
    # Interpolation pattern keys
    'error_connecting': 'Connection error',
    'unexpected_error': 'Unexpected error',
    'server_unreachable': 'Cannot connect to server',
    'upload_failed': 'Upload failed',
    'verify_failed': 'Verification failed',
    'cannot_load_data': 'Cannot load data',
    'customer_name': 'Customer Name',
    'creditor_name': 'Creditor Name',
    'price_retail': 'Retail Price',
    'price_member': 'Member Price',
    'document_type': 'Document Type',
    'vat_type': 'VAT Type',
    'transaction_type': 'Transaction Type',
    'select_warehouse': 'Select Warehouse',
    'select_location': 'Select Location',
    'select_branch': 'Select Branch',
    'select_department': 'Select Department',
    'loading_image': 'Loading image...',
    'generating_pdf': 'Generating PDF...',
    'clear_all': 'Clear All',
    'select_all': 'Select All',
    'no_items': 'No items',
    'transaction_list': 'Transaction List',
    'please_wait': 'Please wait...',
    'loading': 'Loading...',
    'no_data_available': 'No data available',
    'save_failed': 'Save failed',
    'error_occurred': 'An error occurred',
    'please_fill_data': 'Please fill in data',
    'invalid_data': 'Invalid data',
    'saving': 'Saving...',
    'deleting': 'Deleting...',
    'select_date': 'Select Date',
    'view_detail': 'View Detail',
    'please_select': 'Please select',
}


# === INTERPOLATION PATTERNS ===
# (regex_pattern, language_key, replacement_template)
# For strings like "เกิดข้อผิดพลาด: $e" -> "${language("error_occurred")}: $e"
INTERPOLATION_MAP = [
    (r'^เกิดข้อผิดพลาด: \$e$', 'error_occurred', '"${global.language(\"error_occurred\")}: $e"'),
    (r'^เกิดข้อผิดพลาดในการเชื่อมต่อ: \$e$', 'error_connecting', '"${global.language(\"error_connecting\")}: $e"'),
    (r'^เกิดข้อผิดพลาดไม่คาดคิด: ', 'unexpected_error', None),  # skip - complex ${e.toString()}
    (r'^ติดต่อ Server ไม่ได้ : \$exception$', 'server_unreachable', '"${global.language(\"server_unreachable\")}: $exception"'),
    (r'^ติดต่อ Server ไม่ได้ : \$e$', 'server_unreachable', '"${global.language(\"server_unreachable\")}: $e"'),
    (r'^อัพโหลดล้มเหลว: \$e$', 'upload_failed', '"${global.language(\"upload_failed\")}: $e"'),
    (r'^ตรวจสอบล้มเหลว: \$e$', 'verify_failed', '"${global.language(\"verify_failed\")}: $e"'),
    (r'^ไม่สามารถโหลดข้อมูลได้: \$e$', 'cannot_load_data', '"${global.language(\"cannot_load_data\")}: $e"'),
]


def main():
    mode = sys.argv[1] if len(sys.argv) > 1 else '--dry-run'

    # Load existing language data
    lang_data = json.load(open(LANG_JSON, encoding='utf-8'))

    # Build reverse map: Thai text -> existing key
    th_to_key = {}
    for key, translations in lang_data.items():
        if isinstance(translations, dict) and 'th' in translations:
            th_text = translations['th'].strip()
            if th_text and th_text not in th_to_key:
                th_to_key[th_text] = key

    # Merge manual map (manual takes priority)
    full_map = {}
    for thai, key in MANUAL_MAP.items():
        full_map[thai] = key
    for thai, key in th_to_key.items():
        if thai not in full_map:
            full_map[thai] = key

    thai_re = re.compile('[\u0e00-\u0e7f]+')

    print(f"Loaded {len(full_map)} Thai->key mappings")
    print(f"Mode: {mode}")

    # Collect new keys that need to be added to languages.json
    new_keys = {}

    # Track all changes
    total_changes = 0
    changed_files = 0

    for root, dirs, files in os.walk(LIB_DIR):
        for fn in files:
            if not fn.endswith('.dart'):
                continue
            fp = os.path.join(root, fn)
            rel = fp.replace(LIB_DIR + BS, 'lib/').replace(BS, '/')

            # Skip login flow files
            if rel in SKIP_FILES:
                continue

            try:
                content = open(fp, encoding='utf-8').read()
                original = content
            except:
                continue

            lines = content.split('\n')
            new_lines = []
            file_changes = 0

            for line_num, line in enumerate(lines, 1):
                s = line.strip()

                # Skip comments, logging, data structures
                if any(re.search(p, line) for p in SKIP_LINE_PATTERNS):
                    new_lines.append(line)
                    continue

                # Skip if no Thai chars
                if not thai_re.search(line):
                    new_lines.append(line)
                    continue

                # Skip sample data patterns
                if any(re.search(p, line) for p in SAMPLE_PATTERNS):
                    new_lines.append(line)
                    continue

                new_line = line

                # Find Thai strings in this line (handles escaped quotes \")
                for match in re.finditer(r'"((?:[^"\\\n]|\\.)*[\u0e00-\u0e7f](?:[^"\\\n]|\\.)*)"', line):
                    thai_str = match.group(1)
                    full_match = match.group(0)  # includes quotes

                    # Skip if in SKIP_STRINGS
                    if thai_str in SKIP_STRINGS:
                        continue

                    # Skip strings with complex interpolation (${...}) — too complex for auto
                    if '${' in thai_str:
                        continue

                    # Handle simple $variable interpolation (e.g., "เกิดข้อผิดพลาด: $e")
                    if '$' in thai_str:
                        # Only handle known patterns
                        interp_handled = False
                        for interp_pattern, interp_key, interp_template in INTERPOLATION_MAP:
                            if re.match(interp_pattern, thai_str):
                                if interp_template is None:
                                    # Skip complex patterns
                                    interp_handled = True
                                    break
                                replacement = interp_template
                                new_line = new_line.replace(f'"{thai_str}"', replacement, 1)
                                file_changes += 1
                                interp_handled = True
                                if interp_key not in lang_data:
                                    th_text = re.sub(r'\$\w+', '', thai_str).strip().rstrip(':').strip()
                                    new_keys[interp_key] = {
                                        'th': th_text,
                                        'en': NEW_KEY_TRANSLATIONS.get(interp_key, interp_key),
                                    }
                                break
                        if interp_handled:
                            continue
                        # Skip unknown interpolation patterns
                        continue

                    # Check if this specific string is inside language() call
                    before = line[:match.start()]
                    if before.rstrip().endswith('language('):
                        # Already in language() - just need to check if key exists
                        if thai_str in full_map:
                            key = full_map[thai_str]
                            new_line = new_line.replace(f'language("{thai_str}")', f'language("{key}")')
                            file_changes += 1
                        continue

                    # Try to map
                    if thai_str in full_map:
                        key = full_map[thai_str]

                        # Determine replacement based on context
                        if 'pdfgen' in rel:
                            # PDF: use global.language()
                            replacement = f'global.language("{key}")'
                        elif 'global.' in before or 'import' in before:
                            replacement = f'language("{key}")'
                        else:
                            replacement = f'global.language("{key}")'

                        new_line = new_line.replace(f'"{thai_str}"', replacement, 1)
                        file_changes += 1

                        # Track new keys
                        if key not in lang_data:
                            en_text = NEW_KEY_TRANSLATIONS.get(key, thai_str)
                            new_keys[key] = {
                                'th': thai_str,
                                'en': en_text,
                            }

                # Also check single-quoted strings
                for match in re.finditer(r"'([^'\n]*[\u0e00-\u0e7f][^'\n]*)'", line):
                    thai_str = match.group(1)

                    if thai_str in SKIP_STRINGS:
                        continue
                    if '${' in thai_str:
                        continue

                    if thai_str in full_map:
                        key = full_map[thai_str]

                        # Check if this specific string is inside language() call
                        before = line[:match.start()]
                        if before.rstrip().endswith('language('):
                            new_line = new_line.replace(f"language('{thai_str}')", f"language('{key}')")
                            file_changes += 1
                            continue

                        if 'pdfgen' in rel:
                            replacement = f"global.language('{key}')"
                        elif 'global.' in before:
                            replacement = f"language('{key}')"
                        else:
                            replacement = f"global.language('{key}')"

                        new_line = new_line.replace(f"'{thai_str}'", replacement, 1)
                        file_changes += 1

                        if key not in lang_data:
                            en_text = NEW_KEY_TRANSLATIONS.get(key, thai_str)
                            new_keys[key] = {'th': thai_str, 'en': en_text}

                new_lines.append(new_line)

            if file_changes > 0:
                new_content = '\n'.join(new_lines)
                total_changes += file_changes
                changed_files += 1

                if mode == '--apply':
                    with open(fp, 'w', encoding='utf-8') as f:
                        f.write(new_content)
                    print(f"  [{file_changes:3d}] {rel}")
                else:
                    print(f"  [{file_changes:3d}] {rel} (dry-run)")

    print(f"\n=== Summary ===")
    print(f"Files changed: {changed_files}")
    print(f"Total replacements: {total_changes}")
    print(f"New language keys needed: {len(new_keys)}")

    # Save new keys
    outpath = r'D:\bcdev\bcaiaccount\new_lang_keys.json'
    with open(outpath, 'w', encoding='utf-8') as f:
        json.dump(new_keys, f, ensure_ascii=False, indent=2)
    print(f"New keys saved to: {outpath}")

    if mode == '--apply' and new_keys:
        # Add new keys to languages.json
        for key, vals in new_keys.items():
            if key not in lang_data:
                lang_data[key] = {
                    'cn': vals['en'],
                    'en': vals['en'],
                    'ja': vals['en'],
                    'km': vals['en'],
                    'ko': vals['en'],
                    'lo': vals['en'],
                    'my': vals['en'],
                    'th': vals['th'],
                    'vi': vals['en'],
                }

        # Sort and save
        sorted_data = dict(sorted(lang_data.items()))
        with open(LANG_JSON, 'w', encoding='utf-8') as f:
            json.dump(sorted_data, f, ensure_ascii=False, indent=2)
        print(f"Updated languages.json with {len(new_keys)} new keys")


if __name__ == '__main__':
    main()
