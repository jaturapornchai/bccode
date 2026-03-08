#!/usr/bin/env python3
"""Batch localization script - Replace hardcoded Thai text with global.language() calls"""

import re
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
            print(f"  WARNING: Not found in {os.path.basename(filepath)}: {old[:60]}...")

    if content != original:
        with open(filepath, 'w', encoding='utf-8') as f:
            f.write(content)
        print(f"  Updated {os.path.basename(filepath)} ({count} replacements)")
    return count

# ============================================================
# 1. permission_definition_screen.dart
# ============================================================
f1 = os.path.join(BASE, "screens", "config", "permission_definition_screen.dart")
replace_in_file(f1, [
    # Line 48 - error snackbar
    ("global.showErrorSnackBar(context, 'เกิดข้อผิดพลาด: $e');",
     "global.showErrorSnackBar(context, '${global.language(\"error_occurred\")}: $e');"),
    # Line 95 - search hint
    ("hintText: 'ค้นหารหัสสิทธิ์หรือชื่อสิทธิ์...',",
     "hintText: global.language('search_permission_hint'),"),
    # Line 117 - empty list text
    ("'ยังไม่มีรายการสิทธิ์',",
     "global.language('no_permission_list'),"),
    # Line 183 - branch/screen count
    ("'$branchCount สาขา, $screenCount หน้าจอ',",
     "'$branchCount ${global.language(\"branch\")} $screenCount ${global.language(\"screen\")}',"),
    # Line 211 - copy tooltip
    ("tooltip: 'คัดลอกสิทธิ์',",
     "tooltip: global.language('copy_permission'),"),
    # Line 261 - add success
    ("global.showSuccessSnackBar(context, isNew ? 'เพิ่มสิทธิ์สำเร็จ' : 'แก้ไขสิทธิ์สำเร็จ');",
     "global.showSuccessSnackBar(context, isNew ? global.language('add_permission_success') : global.language('edit_permission_success'));"),
    # Line 277 - confirm delete text
    ("content: Text('ต้องการลบสิทธิ์ \"${permission.permissionCode}\" หรือไม่?'),",
     "content: Text('${global.language(\"confirm_delete_permission\")} \"${permission.permissionCode}\"?'),"),
    # Line 303 - delete success
    ("global.showSuccessSnackBar(context, 'ลบสิทธิ์สำเร็จ');",
     "global.showSuccessSnackBar(context, global.language('delete_permission_success'));"),
    # Line 308 - delete error
    ("global.showErrorSnackBar(context, 'เกิดข้อผิดพลาดในการลบ');",
     "global.showErrorSnackBar(context, global.language('error_deleting'));"),
    # Line 327 - copy dialog title
    ("Text('คัดลอกสิทธิ์จาก \"${source.permissionCode}\"'),",
     "Text('${global.language(\"copy_permission_from\")} \"${source.permissionCode}\"'),"),
    # Line 338 - new permission code label
    ("labelText: 'รหัสสิทธิ์ใหม่ *',",
     "labelText: '${global.language(\"new_permission_code\")} *',"),
    # Line 347 - code exists
    ("return 'รหัสสิทธิ์นี้มีอยู่แล้ว';",
     "return global.language('permission_code_exists');"),
    # Line 356 - new permission name label
    ("labelText: 'ชื่อสิทธิ์ใหม่ *',",
     "labelText: '${global.language(\"new_permission_name\")} *',"),
    # Line 363 - enter permission name
    ("return 'กรุณากรอกชื่อสิทธิ์';",
     "return global.language('please_enter_permission_name');"),
    # Line 419 - copy success
    ("global.showSuccessSnackBar(context, 'คัดลอกสิทธิ์ \"${source.permissionCode}\" → \"$newCode\" สำเร็จ');",
     "global.showSuccessSnackBar(context, '${global.language(\"copy_permission_success\")} \"${source.permissionCode}\" → \"$newCode\"');"),
    # Line 424 - copy error
    ("global.showErrorSnackBar(context, 'เกิดข้อผิดพลาดในการคัดลอก');",
     "global.showErrorSnackBar(context, global.language('error_copying'));"),
    # Line 430 - error with $e (in _copyPermission)
    ("global.showErrorSnackBar(context, 'เกิดข้อผิดพลาด: $e');",
     "global.showErrorSnackBar(context, '${global.language(\"error_occurred\")}: $e');"),
    # Line 483 - code exists validator
    ("_codeError = 'รหัสนี้มีอยู่แล้ว';",
     "_codeError = global.language('permission_code_exists');"),
    # Line 653 - permission code label
    ("labelText: 'รหัสสิทธิ์ *',",
     "labelText: '${global.language(\"permission_code_required\")} *',"),
    # Line 694 - permission name label
    ("labelText: 'ชื่อสิทธิ์',",
     "labelText: global.language('permission_name_label'),"),
    # Line 716 - active status
    ("title: const Text('สถานะใช้งาน'),",
     "title: Text(global.language('active_status')),"),
    # Line 723-724 - set permissions by branch
    ("'กำหนดสิทธิ์ตามสาขา',",
     "global.language('set_permissions_by_branch'),"),
    # Line 806 - code label
    ("subtitle: Text('รหัส: ${branch.code}'),",
     "subtitle: Text('${global.language(\"code_label\")}: ${branch.code}'),"),
    # Line 878 - add all
    ("label: 'เพิ่มทั้งหมด',",
     "label: global.language('add_all'),"),
    # Line 885 - edit all
    ("label: 'แก้ไขทั้งหมด',",
     "label: global.language('edit_all'),"),
    # Line 899 - print all
    ("label: 'พิมพ์ทั้งหมด',",
     "label: global.language('print_all'),"),
    # Line 906 - all permissions
    ("label: 'ทุกสิทธิ์',",
     "label: global.language('all_permissions'),"),
    # Line 668 - auto code tooltip
    ("message: 'สร้างรหัสอัตโนมัติ',",
     "message: global.language('auto_generate_code'),"),
])

# ============================================================
# 2. permission_link_screen.dart
# ============================================================
f2 = os.path.join(BASE, "screens", "config", "permission_link_screen.dart")
replace_in_file(f2, [
    # Line 54 - error snackbar
    ("global.showErrorSnackBar(context, 'เกิดข้อผิดพลาด: $e');",
     "global.showErrorSnackBar(context, '${global.language(\"error_occurred\")}: $e');"),
    # Line 105 - search hint
    ("hintText: 'ค้นหาผู้ใช้งาน...',",
     "hintText: global.language('search_user_hint'),"),
    # Line 131 - no permissions warning
    ("'ยังไม่มีสิทธิ์ที่ใช้งานได้ กรุณาไปที่หน้ากำหนดสิทธิ์เพื่อสร้างสิทธิ์ก่อน',",
     "global.language('no_available_permissions_hint'),"),
    # Line 152 - no users found
    ("'ไม่พบผู้ใช้งาน',",
     "global.language('no_users_found'),"),
    # Line 230 - permission count
    ("'$permCount สิทธิ์',",
     "'$permCount ${global.language(\"permission\")}',"),
    # Line 246 - check permission tooltip
    ("tooltip: 'ตรวจสอบสิทธิ์',",
     "tooltip: global.language('check_permission'),"),
    # Line 264 - user role general
    ("return 'ผู้ใช้ทั่วไป';",
     "return global.language('role_general_user');"),
    # Line 266 - admin
    ("return 'แอดมิน';",
     "return global.language('role_admin');"),
    # Line 268 - super admin
    ("return 'ซุปเปอร์แอดมิน';",
     "return global.language('role_super_admin');"),
    # Line 324 - save success
    ("global.showSuccessSnackBar(context, 'บันทึกสิทธิ์สำเร็จ');",
     "global.showSuccessSnackBar(context, global.language('save_permission_success'));"),
    # Line 441 - check combined permission header
    ("'ตรวจสอบสิทธิ์รวม',",
     "global.language('check_combined_permission'),"),
    # Line 445 - user label
    ("'ผู้ใช้: $userName',",
     "'${global.language(\"user_label\")}: $userName',"),
    # Line 475 - linked permissions label
    ("const Text('สิทธิ์ที่เชื่อม: ', style: TextStyle(fontSize: 12)),",
     "Text('${global.language(\"linked_permissions\")}: ', style: TextStyle(fontSize: 12)),"),
    # Line 565 - branch label
    ("'สาขา: $branchCode',",
     "'${global.language(\"branch\")}: $branchCode',"),
    # Line 577 - screen count
    ("'${screens.length} หน้าจอ',",
     "'${screens.length} ${global.language(\"screen\")}',"),
    # Line 785 - user label in dialog
    ("'ผู้ใช้: ${widget.userName}',",
     "'${global.language(\"user_label\")}: ${widget.userName}',"),
    # Line 812 - select permission info
    ("'เลือกสิทธิ์ที่ต้องการให้ผู้ใช้นี้มี (สามารถเลือกได้หลายสิทธิ์ - จะรวมสิทธิ์ทั้งหมดเข้าด้วยกัน)',",
     "global.language('select_permission_info'),"),
    # Line 830 - no permissions available
    ("const Text('ไม่มีสิทธิ์ที่ใช้งานได้'),",
     "Text(global.language('no_available_permissions')),"),
    # Line 832 - go to permission definition
    ("'กรุณาไปที่หน้ากำหนดสิทธิ์เพื่อสร้างสิทธิ์ก่อน',",
     "global.language('go_to_permission_definition'),"),
    # Line 889 - save button text
    ("label: Text('บันทึก (${_selectedCodes.length} สิทธิ์)'),",
     "label: Text('${global.language(\"save\")} (${_selectedCodes.length} ${global.language(\"permission\")})'),"),
    # Line 931 - branch/screen count
    ("'$branchCount สาขา, $screenCount หน้าจอ',",
     "'$branchCount ${global.language(\"branch\")}, $screenCount ${global.language(\"screen\")}',"),
])

# ============================================================
# 3. mcp_apikey_screen.dart
# ============================================================
f3 = os.path.join(BASE, "screens", "config", "mcp_apikey_screen.dart")
replace_in_file(f3, [
    # Line 790 - permission level
    ("const Text('ระดับสิทธิ์', style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),",
     "Text(global.language('permission_level'), style: TextStyle(fontWeight: FontWeight.bold, fontSize: 14)),"),
    # Line 797 - readonly subtitle
    ("subtitle: 'ดูข้อมูลอย่างเดียว (${MCPTools.readonlyTools.length} tools) — เหมาะกับ Claude Desktop, frontend dev',",
     "subtitle: '${global.language(\"mcp_readonly_desc\")} (${MCPTools.readonlyTools.length} tools)',"),
    # Line 804 - developer subtitle
    ("subtitle: 'อ่าน+เขียน ทุก tool (${MCPTools.allTools.length} tools) — เหมาะกับ backend dev',",
     "subtitle: '${global.language(\"mcp_readwrite_desc\")} (${MCPTools.allTools.length} tools)',"),
    # Line 811 - custom subtitle
    ("subtitle: 'เลือก tools เองทีละตัว',",
     "subtitle: global.language('mcp_custom_desc'),"),
    # Line 821 - select tools
    ("'เลือก Tools (${_selectedTools.length}/${MCPTools.allTools.length})',",
     "'${global.language(\"select_tools\")} (${_selectedTools.length}/${MCPTools.allTools.length})',"),
    # Line 838 - clear button
    ("child: const Text('ล้าง', style: TextStyle(fontSize: 11, color: Colors.red)),",
     "child: Text(global.language('clear'), style: TextStyle(fontSize: 11, color: Colors.red)),"),
])

# ============================================================
# 4. lineoa_config_screen.dart
# ============================================================
f4 = os.path.join(BASE, "screens", "lineoa", "lineoa_config_screen.dart")
replace_in_file(f4, [
    # Line 57 - load error
    ("global.showErrorSnackBar(context, 'โหลดข้อมูลไม่สำเร็จ: ${result.errorMessage}');",
     "global.showErrorSnackBar(context, '${global.language(\"load_data_failed\")}: ${result.errorMessage}');"),
    # Line 66 - title
    ("title: const Text('ตั้งค่า Line OA'),",
     "title: Text(global.language('lineoa_config_title')),"),
    # Line 105 - connect line oa
    ("'เชื่อมต่อ Line OA',",
     "global.language('connect_lineoa'),"),
    # Line 169 - not activated
    (": 'ยังไม่เปิดใช้งาน')",
     ": global.language('not_activated'))"),
    # Line 213 - select type
    ("'เลือกประเภท Line OA ที่ต้องการตั้งค่า',",
     "global.language('select_lineoa_type'),"),
    # Line 310 - fill fields warning
    ("global.showWarningSnackBar(context, 'กรุณากรอก Channel ID, Channel Secret และ Access Token');",
     "global.showWarningSnackBar(context, global.language('please_fill_all_fields'));"),
    # Line 363 - save error
    ("global.showErrorSnackBar(context, 'บันทึกไม่สำเร็จ: ${result.errorMessage}');",
     "global.showErrorSnackBar(context, '${global.language(\"save_failed\")}: ${result.errorMessage}');"),
    # Line 393 - config header
    ("'ตั้งค่า ${widget.type.displayName}',",
     "'${global.language(\"settings\")} ${widget.type.displayName}',"),
    # Line 570 - copied
    ("global.showInfoSnackBar(context, 'คัดลอกแล้ว');",
     "global.showInfoSnackBar(context, global.language('copied'));"),
    # Line 579 - required field
    ("return 'กรุณากรอก $label';",
     "return '${global.language(\"please_fill\")} $label';"),
    # Line 612 - link success
    ("'เชื่อมต่อสำเร็จ'",
     "global.language('link_success')"),
    # Line 652 - connected employees
    ("'พนักงานที่เชื่อมต่อ',",
     "global.language('connected_employees'),"),
    # Line 756 - no employees
    ("'ยังไม่มีพนักงานเชื่อมต่อ',",
     "global.language('no_connected_employees'),"),
    # Line 785 - employee code
    ("Text('รหัส: ${employee.employeeCode}'),",
     "Text('${global.language(\"code_label\")}: ${employee.employeeCode}'),"),
    # Line 797 - create connect link tooltip
    ("tooltip: 'สร้างลิงก์เชื่อมต่อ',",
     "tooltip: global.language('create_connect_link'),"),
    # Line 831 - line connect success
    ("global.showSuccessSnackBar(context, 'เชื่อมต่อ LINE สำเร็จ: ${result.displayName}');",
     "global.showSuccessSnackBar(context, '${global.language(\"link_success\")}: ${result.displayName}');"),
    # Line 842 - confirm remove
    ("content: Text('ต้องการลบ ${employee.employeeName} ออกจาก Line OA?'),",
     "content: Text('${global.language(\"confirm_delete\")} ${employee.employeeName}?'),"),
    # Line 899 - fill data warning
    ("global.showWarningSnackBar(context, 'กรุณากรอกข้อมูลให้ครบ');",
     "global.showWarningSnackBar(context, global.language('please_fill_all_fields'));"),
    # Line 918 - add error
    ("global.showErrorSnackBar(context, 'เพิ่มไม่สำเร็จ: ${result.errorMessage}');",
     "global.showErrorSnackBar(context, '${global.language(\"add_failed\")}: ${result.errorMessage}');"),
])

# ============================================================
# 5. shelf_product_selector.dart
# ============================================================
f5 = os.path.join(BASE, "screens", "components", "shelf_product_selector.dart")
replace_in_file(f5, [
    # Line 56 - header
    ("'เลือกสินค้าจากชั้นวาง',",
     "global.language('select_product_from_shelf'),"),
    # Line 141 - loading warehouse
    ("const Text('กำลังโหลดคลังสินค้า...'),",
     "Text(global.language('loading_warehouse')),"),
    # Line 214 - please select warehouse
    ("'โปรดเลือกคลังสินค้าก่อน',",
     "global.language('please_select_warehouse'),"),
    # Line 234 - loading locations
    ("const Text('กำลังโหลดตำแหน่งที่เก็บ...'),",
     "Text(global.language('loading_locations')),"),
    # Line 311 - please select location
    ("'โปรดเลือกตำแหน่งที่เก็บก่อน',",
     "global.language('please_select_location'),"),
    # Line 398 - please select shelf
    ("'โปรดเลือกชั้นวางเพื่อดูรายการสินค้า',",
     "global.language('please_select_shelf'),"),
    # Line 425 - loading products
    ("const Text('กำลังโหลดรายการสินค้า...'),",
     "Text(global.language('loading_products')),"),
    # Line 351 - items count
    ("'$productCount รายการ',",
     "'$productCount ${global.language(\"items\")}',"),
    # Line 482 - products in shelf header
    ("'สินค้าในชั้นวาง (${state.shelfProducts.where((p) => !deletedProducts.contains(p.barcode)).length} รายการ)',",
     "'${global.language(\"products_in_shelf\")} (${state.shelfProducts.where((p) => !deletedProducts.contains(p.barcode)).length} ${global.language(\"items\")})',"),
    # Line 521 - barcode label
    ("'บาร์โค้ด: ${product.barcode}',",
     "'${global.language(\"barcode_label\")}: ${product.barcode}',"),
    # Line 531 - delete tooltip
    ("tooltip: 'ลบสินค้านี้',",
     "tooltip: global.language('delete_this_product'),"),
    # Line 572 - add all products button
    ("? 'เพิ่มสินค้าทั้งหมด (${availableProducts.length})'",
     "? '${global.language(\"add_all_products\")} (${availableProducts.length})'"),
])

# ============================================================
# 6. print_label_dialog.dart
# ============================================================
f6 = os.path.join(BASE, "screens", "components", "print_label_dialog.dart")
replace_in_file(f6, [
    # Line 173 - select label format
    ("const Text('กรุณาเลือกรูปแบบการพิมพ์ฉลากสินค้า'),",
     "Text(global.language('select_label_print_format')),"),
    # Line 324 - color description
    ("'เลือกสีสำหรับข้อความราคาและตั้งค่าการแสดงกรอบ (${widget.labelSize.description})',",
     "'${global.language(\"select_color_and_border\")} (${widget.labelSize.description})',"),
    # Line 351 - red color
    ("title: 'สีแดง',",
     "title: global.language('color_red'),"),
    # Line 352 - red description
    ("description: 'เหมาะสำหรับการเน้นราคาพิเศษ',",
     "description: global.language('label_special_price_description'),"),
    # Line 341 - black description
    ("description: 'เหมาะสำหรับการพิมพ์ทั่วไป',",
     "description: global.language('label_general_description'),"),
    # Line 378 - border display header
    ("'การแสดงกรอบ',",
     "global.language('border_display'),"),
    # Line 393 - show border
    ("label: 'แสดงกรอบ',",
     "label: global.language('show_border'),"),
    # Line 400 - hide border
    ("label: 'ไม่แสดงกรอบ',",
     "label: global.language('hide_border'),"),
    # LabelSize enum descriptions - these are in const-like context, skip enum
    # Line 25 - regular displayName (this is in extension)
    ("return 'พิมพ์บนกระดาษ A4';",
     "return global.language('print_on_a4');"),
    # Description texts in extension
    ("return 'รูปแบบดั้งเดิม สำหรับพิมพ์ติดสินค้า';",
     "return global.language('label_format_regular_desc');"),
])

# ============================================================
# 7. purchase_type_screen.dart
# ============================================================
f7 = os.path.join(BASE, "screens", "config", "approval", "purchase_type_screen.dart")
replace_in_file(f7, [
    # Line 55 - load error
    ("_showError('โหลดข้อมูลไม่สำเร็จ: ${result.errorMessage}');",
     "_showError('${global.language(\"load_data_failed\")}: ${result.errorMessage}');"),
    # Line 106 - save error
    ("_showError('บันทึกไม่สำเร็จ: ${result.errorMessage}');",
     "_showError('${global.language(\"save_failed\")}: ${result.errorMessage}');"),
    # Line 122 - confirm delete
    ("content: Text('ต้องการลบประเภท \"${item.getName(global.systemLanguage)}\" หรือไม่?'),",
     "content: Text('${global.language(\"confirm_delete\")} \"${item.getName(global.systemLanguage)}\"?'),"),
    # Line 239 - list header
    ("'รายการประเภทการจัดซื้อ',",
     "global.language('purchase_type_list'),"),
    # Line 255 - items count
    ("'${_items.length} รายการ',",
     "'${_items.length} ${global.language(\"items\")}',"),
    # Line 329 - empty state
    ("'ไม่มีข้อมูลประเภทการจัดซื้อ',",
     "global.language('no_purchase_types'),"),
    # Line 338 - add hint
    ("'กดปุ่มด้านล่างเพื่อเพิ่มประเภทการจัดซื้อใหม่',",
     "global.language('press_below_to_add_purchase_type'),"),
    # Line 521 - edit header
    ("'แก้ไขประเภทการจัดซื้อ',",
     "global.language('edit_purchase_type'),"),
    # Line 531 - edit subtitle
    ("'แก้ไขข้อมูลประเภท: ${_selectedItem!.code}',",
     "'${global.language(\"edit_type_data\")}: ${_selectedItem!.code}',"),
])

# ============================================================
# 8. quotation_type_screen.dart
# ============================================================
f8 = os.path.join(BASE, "screens", "config", "approval", "quotation_type_screen.dart")
replace_in_file(f8, [
    # Line 55 - load error
    ("_showError('โหลดข้อมูลไม่สำเร็จ: ${result.errorMessage}');",
     "_showError('${global.language(\"load_data_failed\")}: ${result.errorMessage}');"),
    # Line 106 - save error
    ("_showError('บันทึกไม่สำเร็จ: ${result.errorMessage}');",
     "_showError('${global.language(\"save_failed\")}: ${result.errorMessage}');"),
    # Line 122 - confirm delete
    ("content: Text('ต้องการลบประเภท \"${item.getName(global.systemLanguage)}\" หรือไม่?'),",
     "content: Text('${global.language(\"confirm_delete\")} \"${item.getName(global.systemLanguage)}\"?'),"),
    # Line 239 - list header
    ("'รายการประเภทใบเสนอราคา',",
     "global.language('quotation_type_list'),"),
    # Line 255 - items count
    ("'${_items.length} รายการ',",
     "'${_items.length} ${global.language(\"items\")}',"),
    # Line 329 - empty state
    ("'ไม่มีข้อมูลประเภทใบเสนอราคา',",
     "global.language('no_quotation_types'),"),
    # Line 338 - add hint
    ("'กดปุ่มด้านล่างเพื่อเพิ่มประเภทใบเสนอราคาใหม่',",
     "global.language('press_below_to_add_quotation_type'),"),
    # Line 521 - edit header
    ("'แก้ไขประเภทใบเสนอราคา',",
     "global.language('edit_quotation_type'),"),
    # Line 531 - edit subtitle
    ("'แก้ไขข้อมูลประเภท: ${_selectedItem!.code}',",
     "'${global.language(\"edit_type_data\")}: ${_selectedItem!.code}',"),
])

# ============================================================
# 9. sale_order_type_screen.dart
# ============================================================
f9 = os.path.join(BASE, "screens", "config", "approval", "sale_order_type_screen.dart")
replace_in_file(f9, [
    # Line 55 - load error
    ("_showError('โหลดข้อมูลไม่สำเร็จ: ${result.errorMessage}');",
     "_showError('${global.language(\"load_data_failed\")}: ${result.errorMessage}');"),
    # Line 106 - save error
    ("_showError('บันทึกไม่สำเร็จ: ${result.errorMessage}');",
     "_showError('${global.language(\"save_failed\")}: ${result.errorMessage}');"),
    # Line 122 - confirm delete
    ("content: Text('ต้องการลบประเภท \"${item.getName(global.systemLanguage)}\" หรือไม่?'),",
     "content: Text('${global.language(\"confirm_delete\")} \"${item.getName(global.systemLanguage)}\"?'),"),
    # Line 239 - list header
    ("'รายการประเภทใบสั่งขาย',",
     "global.language('sale_order_type_list'),"),
    # Line 255 - items count
    ("'${_items.length} รายการ',",
     "'${_items.length} ${global.language(\"items\")}',"),
    # Line 329 - empty state
    ("'ไม่มีข้อมูลประเภทใบสั่งขาย',",
     "global.language('no_sale_order_types'),"),
    # Line 338 - add hint
    ("'กดปุ่มด้านล่างเพื่อเพิ่มประเภทใบสั่งขายใหม่',",
     "global.language('press_below_to_add_sale_order_type'),"),
    # Line 521 - edit header
    ("'แก้ไขประเภทใบสั่งขาย',",
     "global.language('edit_sale_order_type'),"),
    # Line 531 - edit subtitle
    ("'แก้ไขข้อมูลประเภท: ${_selectedItem!.code}',",
     "'${global.language(\"edit_type_data\")}: ${_selectedItem!.code}',"),
])

# ============================================================
# 10. po_approval_setting_screen.dart
# ============================================================
f10 = os.path.join(BASE, "screens", "config", "approval", "po_approval_setting_screen.dart")
replace_in_file(f10, [
    # Line 106 - cannot create default
    ("_showError('ไม่สามารถสร้างประเภทการจัดซื้อเริ่มต้นได้');",
     "_showError(global.language('cannot_create_default_purchase_type'));"),
    # Line 174 - rule validation error
    ("_showError('วงเงินขั้นต่ำของกฎที่ ${i + 1} ต้องมากกว่าหรือเท่ากับวงเงินสูงสุดของกฎก่อนหน้า');",
     "_showError('${global.language(\"rule_min_amount_error\")} ${i + 1}');"),
    # Line 199 - save error
    ("_showError('บันทึกไม่สำเร็จ: ${result.errorMessage}');",
     "_showError('${global.language(\"save_failed\")}: ${result.errorMessage}');"),
])

# ============================================================
# 11. qt_approval_setting_screen.dart
# ============================================================
f11 = os.path.join(BASE, "screens", "config", "approval", "qt_approval_setting_screen.dart")
replace_in_file(f11, [
    # Line 106 - cannot create default
    ("_showError('ไม่สามารถสร้างประเภทใบเสนอราคาเริ่มต้นได้');",
     "_showError(global.language('cannot_create_default_quotation_type'));"),
    # Line 174 - rule validation error
    ("_showError('วงเงินขั้นต่ำของกฎที่ ${i + 1} ต้องมากกว่าหรือเท่ากับวงเงินสูงสุดของกฎก่อนหน้า');",
     "_showError('${global.language(\"rule_min_amount_error\")} ${i + 1}');"),
    # Line 199 - save error
    ("_showError('บันทึกไม่สำเร็จ: ${result.errorMessage}');",
     "_showError('${global.language(\"save_failed\")}: ${result.errorMessage}');"),
])

# ============================================================
# 12. so_approval_setting_screen.dart
# ============================================================
f12 = os.path.join(BASE, "screens", "config", "approval", "so_approval_setting_screen.dart")
replace_in_file(f12, [
    # Line 106 - cannot create default
    ("_showError('ไม่สามารถสร้างประเภทใบสั่งขายเริ่มต้นได้');",
     "_showError(global.language('cannot_create_default_sale_order_type'));"),
    # Line 174 - rule validation error
    ("_showError('วงเงินขั้นต่ำของกฎที่ ${i + 1} ต้องมากกว่าหรือเท่ากับวงเงินสูงสุดของกฎก่อนหน้า');",
     "_showError('${global.language(\"rule_min_amount_error\")} ${i + 1}');"),
    # Line 199 - save error
    ("_showError('บันทึกไม่สำเร็จ: ${result.errorMessage}');",
     "_showError('${global.language(\"save_failed\")}: ${result.errorMessage}');"),
])

print("\n=== Done! ===")
