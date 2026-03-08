#!/usr/bin/env python3
"""Batch localization script - Part 3b: Remaining files (user_screen, chatbot_overlay, debtor_screen, map_get_location_screen)"""

import os

BASE = r"D:\bcdev\bcaiaccount\lib"

def replace_in_file(filepath, replacements):
    """Replace exact strings in a file. Returns count of replacements made."""
    with open(filepath, 'r', encoding='utf-8') as f:
        content = f.read()

    original = content
    count = 0
    for old, new in replacements:
        if old in content:
            content = content.replace(old, new, 1)
            count += 1
        else:
            print(f"  WARNING: Not found in {os.path.basename(filepath)}: {old[:80]}...")

    if content != original:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(content)
        print(f"  Updated {os.path.basename(filepath)} ({count} replacements)")
    return count

# ============================================================
# 1. map_get_location_screen.dart
# ============================================================
f1 = os.path.join(BASE, "screens", "config", "map_get_location_screen.dart")
replace_in_file(f1, [
    # Line 84 - location service warning
    ("global.showWarningSnackBar(context, 'กรุณาเปิด Location Service ในการตั้งค่าเครื่อง');",
     "global.showWarningSnackBar(context, global.language('please_enable_location_service'));"),
    # Line 95 - location permission denied
    ("global.showWarningSnackBar(context, 'สิทธิ์การเข้าถึงตำแหน่งถูกปฏิเสธ');",
     "global.showWarningSnackBar(context, global.language('location_permission_denied'));"),
    # Line 103 - location permission denied forever
    ("global.showErrorSnackBar(context, 'สิทธิ์การเข้าถึงตำแหน่งถูกปฏิเสธถาวร — กรุณาเปิดในการตั้งค่าเครื่อง');",
     "global.showErrorSnackBar(context, global.language('location_permission_denied_forever'));"),
    # Line 121 - cannot find location
    ("global.showErrorSnackBar(context, 'ไม่สามารถหาตำแหน่งปัจจุบันได้: $e');",
     "global.showErrorSnackBar(context, '${global.language(\"cannot_get_current_location\")}: $e');"),
])

# ============================================================
# 2. user_screen.dart
# ============================================================
f2 = os.path.join(BASE, "screens", "config", "user_screen.dart")
replace_in_file(f2, [
    # Line 95 - department: accounting (in commonDepartments list)
    ("    'บัญชี',",
     "    global.language('accounting'),"),
    # Line 97 - department: purchasing
    ("    'จัดซื้อ',",
     "    global.language('purchasing'),"),
    # Line 99 - department: marketing
    ("    'การตลาด',",
     "    global.language('marketing'),"),
    # Line 102 - department: production
    ("    'ผลิต',",
     "    global.language('production'),"),
    # Line 103 - department: HR
    ("    'บุคคล',",
     "    global.language('human_resources'),"),
    # Line 299 - cleanup users tooltip
    ("tooltip: 'ลบข้อมูลผู้ใช้ที่ไม่ถูกต้อง',",
     "tooltip: global.language('cleanup_invalid_users'),"),
    # Line 645 - test message tooltip
    ("tooltip: 'ส่งข้อความทดสอบ',",
     "tooltip: global.language('send_test_message'),"),
    # Line 653 - reconnect LINE tooltip
    ("tooltip: 'เชื่อมต่อ LINE ใหม่',",
     "tooltip: global.language('reconnect_line'),"),
    # Line 809 - cleanup dialog title
    ("title: const Text('ล้างข้อมูลผู้ใช้ที่ไม่ถูกต้อง'),",
     "title: Text(global.language('cleanup_invalid_users')),"),
    # Line 810-814 - cleanup dialog content
    ("content: const Text(\n"
     "          'ต้องการลบข้อมูลต่อไปนี้หรือไม่?\\n\\n'\n"
     "          '1. ผู้ใช้ที่ชื่อผู้ใช้ว่าง\\n'\n"
     "          '2. ข้อมูล LINE ที่ไม่มีผู้ใช้แล้ว\\n\\n'\n"
     "          'การดำเนินการนี้ไม่สามารถยกเลิกได้',\n"
     "        ),",
     "content: Text(\n"
     "          global.language('cleanup_users_confirm_message'),\n"
     "        ),"),
    # Line 1149 - approval level description label
    ("'คำอธิบายระดับ:',",
     "global.language('level_description'),"),
    # Line 1183 - max approval amount label
    ("labelText: 'วงเงินอนุมัติสูงสุด (บาท)',",
     "labelText: global.language('max_approval_amount_label'),"),
    # Line 1184 - hint: 0 = unlimited
    ("hintText: '0 = ไม่จำกัด',",
     "hintText: global.language('zero_means_unlimited'),"),
    # Line 1187 - suffix: baht
    ("suffixText: 'บาท',",
     "suffixText: global.language('baht'),"),
    # Line 1188 - helper: must select role first
    ("helperText: currentRole == 0 ? 'ต้องเลือกบทบาทการอนุมัติก่อน' : null,",
     "helperText: currentRole == 0 ? global.language('must_select_approval_role_first') : null,"),
    # Line 1481 - position label
    ("labelText: 'ตำแหน่งงาน',",
     "labelText: global.language('job_position'),"),
    # Line 1482 - position hint
    ("hintText: 'เช่น หัวหน้าแผนก, ผู้จัดการฝ่าย',",
     "hintText: global.language('job_position_hint'),"),
    # Line 1525 - department hint
    ("hintText: 'เช่น IT, บัญชี, จัดซื้อ',",
     "hintText: global.language('department_hint'),"),
    # Line 1594 - delete user confirm
    ("content: Text('ลบผู้ใช้ \"$username\" ?'),",
     "content: Text('${global.language(\"confirm_delete\")} \"$username\"?'),"),
])

# ============================================================
# 3. chatbot_overlay.dart
# ============================================================
f3 = os.path.join(BASE, "screens", "chatbot", "chatbot_overlay.dart")
replace_in_file(f3, [
    # Lines 79-85 - suggested questions
    ("'ยอดขายวันนี้เท่าไหร่',",
     "global.language('chatbot_q_today_sales'),"),
    ("'สินค้าขายดี 10 อันดับเดือนนี้',",
     "global.language('chatbot_q_top_selling'),"),
    ("'สินค้าไหนใกล้หมดสต็อก',",
     "global.language('chatbot_q_low_stock'),"),
    ("'เปรียบเทียบยอดขายเดือนนี้กับเดือนที่แล้ว',",
     "global.language('chatbot_q_compare_sales'),"),
    ("'ลูกค้ารายใหญ่ 5 อันดับแรก',",
     "global.language('chatbot_q_top_customers'),"),
    ("'กำไรเดือนนี้เท่าไหร่',",
     "global.language('chatbot_q_profit'),"),
    ("'สุขภาพธุรกิจเป็นอย่างไร',",
     "global.language('chatbot_q_business_health'),"),
    # Line 202 - retry message
    ("text: '$errorText\\n\\n⏳ ลองใหม่อีกครั้งใน $retryDelay วินาที... (ครั้งที่ ${attempt + 1}/$maxAttempts)',",
     "text: '$errorText\\n\\n⏳ ${global.language(\"chatbot_retry_message\")} $retryDelay ${global.language(\"seconds\")}... (${global.language(\"attempt\")} ${attempt + 1}/$maxAttempts)',"),
    # Line 1475 - tool label: sales by date range
    ("'get_sales_by_date_range': 'ยอดขายตามช่วงเวลา',",
     "'get_sales_by_date_range': global.language('sales_by_date_range'),"),
    # Line 1477 - tool label: sales by seller
    ("'get_sales_by_seller': 'ยอดขายตามพนักงาน',",
     "'get_sales_by_seller': global.language('sales_by_seller'),"),
    # Line 1478 - tool label: monthly summary
    ("'get_monthly_summary': 'สรุปรายเดือน',",
     "'get_monthly_summary': global.language('monthly_summary'),"),
    # Line 1480 - tool label: business health
    ("'get_business_health': 'สุขภาพธุรกิจ',",
     "'get_business_health': global.language('business_health'),"),
    # Line 1481 - tool label: profit analysis
    ("'get_profit_analysis': 'วิเคราะห์กำไร',",
     "'get_profit_analysis': global.language('profit_analysis'),"),
    # Line 1486 - tool label: low stock
    ("'get_low_stock_alerts': 'สินค้าใกล้หมด',",
     "'get_low_stock_alerts': global.language('low_stock_alerts'),"),
    # Line 1487 - tool label: dead stock
    ("'get_dead_stock': 'สินค้าค้างสต็อก',",
     "'get_dead_stock': global.language('dead_stock'),"),
    # Line 1489 - tool label: top customers
    ("'get_top_customers': 'ลูกค้ารายใหญ่',",
     "'get_top_customers': global.language('top_customers'),"),
    # Line 1490 - tool label: customer growth
    ("'get_customer_growth': 'การเติบโตลูกค้า',",
     "'get_customer_growth': global.language('customer_growth'),"),
    # Line 1492 - tool label: yoy comparison
    ("'get_yoy_comparison': 'เปรียบเทียบปีต่อปี',",
     "'get_yoy_comparison': global.language('yoy_comparison'),"),
    # Line 1493 - tool label: mom comparison
    ("'get_mom_comparison': 'เปรียบเทียบเดือนต่อเดือน',",
     "'get_mom_comparison': global.language('mom_comparison'),"),
    # Line 1494 - tool label: list units
    ("'list_units': 'หน่วยนับสินค้า',",
     "'list_units': global.language('product_unit'),"),
    # Line 1513 - AI data source label
    ("'AI ใช้ข้อมูลจาก:',",
     "global.language('ai_data_source'),"),
    # Line 1588 - AI thinking
    ("'AI กำลังคิด...',",
     "global.language('ai_thinking'),"),
    # Line 1595 - fetching data
    ("'กำลังดึงข้อมูลจากระบบ อาจใช้เวลาสักครู่',",
     "global.language('fetching_data_from_system'),"),
])

# ============================================================
# 4. debtor_screen.dart
# ============================================================
f4 = os.path.join(BASE, "screens", "config", "debtor_screen.dart")
replace_in_file(f4, [
    # Line 314 - select import type
    ("'เลือกประเภทข้อมูลที่ต้องการนำเข้า',",
     "global.language('select_import_type'),"),
    # Line 322-323 - import debtor data
    ("title: const Text(\n"
     "                  'นำเข้าข้อมูลลูกหนี้',",
     "title: Text(\n"
     "                  global.language('import_debtor_data'),"),
    # Line 326 - import debtor subtitle
    ("subtitle: const Text('นำเข้ารายชื่อและข้อมูลลูกหนี้'),",
     "subtitle: Text(global.language('import_debtor_list_data')),"),
    # Line 346-347 - import debtor points
    ("title: const Text(\n"
     "                  'นำเข้าแต้มลูกหนี้',",
     "title: Text(\n"
     "                  global.language('import_debtor_points'),"),
    # Line 350 - import points subtitle
    ("subtitle: const Text('นำเข้าข้อมูลแต้มสะสมของลูกหนี้'),",
     "subtitle: Text(global.language('import_debtor_points_data')),"),
    # Line 390 - select template to download
    ("'เลือก Template ที่ต้องการดาวน์โหลด',",
     "global.language('select_template_to_download'),"),
    # Line 398-399 - debtor template title
    ("title: const Text(\n"
     "                  'Template ข้อมูลลูกหนี้',",
     "title: Text(\n"
     "                  global.language('debtor_data_template'),"),
    # Line 402 - debtor template subtitle
    ("subtitle: const Text('ไฟล์ตัวอย่างสำหรับนำเข้าข้อมูลลูกหนี้'),",
     "subtitle: Text(global.language('sample_file_import_debtor')),"),
    # Line 422-423 - points template title
    ("title: const Text(\n"
     "                  'Template แต้มลูกหนี้',",
     "title: Text(\n"
     "                  global.language('debtor_points_template'),"),
    # Line 426 - points template subtitle
    ("subtitle: const Text('ไฟล์ตัวอย่างสำหรับนำเข้าแต้มลูกหนี้'),",
     "subtitle: Text(global.language('sample_file_import_points')),"),
    # Line 456 - loading title
    ("String title = importType == 'debtor' ? 'ข้อมูลลูกหนี้' : 'แต้มลูกหนี้';",
     "String title = importType == 'debtor' ? global.language('debtor_data') : global.language('debtor_points');"),
    # Line 505 - importing title
    ("'กำลังนำเข้า$title...',",
     "'${global.language(\"importing\")} $title...',"),
    # Line 514 - processing items
    ("'กำลังประมวลผล $itemCount รายการ',",
     "'${global.language(\"processing\")} $itemCount ${global.language(\"items\")}',"),
    # Line 842 - no valid points warning
    ("global.showWarningSnackBar(context, 'ไม่พบข้อมูลแต้มที่ถูกต้องในไฟล์ Excel');",
     "global.showWarningSnackBar(context, global.language('no_valid_points_in_file'));"),
    # Line 846 - import points error
    ("global.showErrorSnackBar(context, 'เกิดข้อผิดพลาดในการ Import แต้ม: $e');",
     "global.showErrorSnackBar(context, '${global.language(\"import_points_error\")}: $e');"),
    # Line 929 - download success
    ("\"ดาวน์โหลดไฟล์ template สำเร็จ\",",
     "global.language('download_template_success'),"),
    # Line 939 - download failed
    ("\"ดาวน์โหลดไฟล์ไม่สำเร็จ\",",
     "global.language('download_template_failed'),"),
    # Line 2241 - points balance label
    ("labelText: 'แต้มคงเหลือ',",
     "labelText: global.language('points_balance'),"),
    # Line 3405 - import points success
    ("'นำเข้าแต้มสำเร็จ ${state.successCount} รายการ';",
     "'${global.language(\"import_points_success\")} ${state.successCount} ${global.language(\"items\")}';"),
    # Line 3408 - failed count
    ("message += ' (ล้มเหลว ${state.failedCount} รายการ)';",
     "message += ' (${global.language(\"failed\")} ${state.failedCount} ${global.language(\"items\")})';"),
    # Line 3421 - and more
    ("'• และอื่นๆ อีก ${state.failedItems!.length - 3} รายการ';",
     "'• ${global.language(\"and_more\")} ${state.failedItems!.length - 3} ${global.language(\"items\")}';"),
    # Line 3454 - import points failed
    ("'นำเข้าแต้มไม่สำเร็จ: ${state.message}',",
     "'${global.language(\"import_points_failed\")}: ${state.message}',"),
])

print("\n=== Done! Part 3b complete ===")
