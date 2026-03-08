#!/usr/bin/env python3
"""Fix const context errors from batch3 localization — remove const from expressions containing global.language()"""

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
        print(f"  Fixed {os.path.basename(filepath)} ({count} const removals)")
    return count

# ============================================================
# 1. chatbot_overlay.dart — suggestedQuestions: const [...]
# ============================================================
f1 = os.path.join(BASE, "screens", "chatbot", "chatbot_overlay.dart")
replace_in_file(f1, [
    ("suggestedQuestions: const [",
     "suggestedQuestions: ["),
])

# ============================================================
# 2. shelf_product_selector.dart — const Align/Text with global.language()
# ============================================================
f2 = os.path.join(BASE, "screens", "components", "shelf_product_selector.dart")
replace_in_file(f2, [
    # Line 211 - const Align wrapping Text with global.language
    ("child: const Align(\n"
     "                    alignment: Alignment.centerLeft,\n"
     "                    child: Text(\n"
     "                      global.language('please_select_warehouse'),",
     "child: Align(\n"
     "                    alignment: Alignment.centerLeft,\n"
     "                    child: Text(\n"
     "                      global.language('please_select_warehouse'),"),
    # Line 308 - const Align wrapping Text with global.language
    ("child: const Align(\n"
     "                    alignment: Alignment.centerLeft,\n"
     "                    child: Text(\n"
     "                      global.language('please_select_location'),",
     "child: Align(\n"
     "                    alignment: Alignment.centerLeft,\n"
     "                    child: Text(\n"
     "                      global.language('please_select_location'),"),
])

# Line 398 - this one doesn't have const, check but the error says line 398
# Actually line 398 is in a non-const context already, let me check the actual error

# ============================================================
# 3. debtor_screen.dart
# ============================================================
f3 = os.path.join(BASE, "screens", "config", "debtor_screen.dart")
replace_in_file(f3, [
    # Line 313 - const Text with global.language
    ("const Text(\n"
     "                global.language('select_import_type'),",
     "Text(\n"
     "                global.language('select_import_type'),"),
    # Line 389 - const Text with global.language
    ("const Text(\n"
     "                global.language('select_template_to_download'),",
     "Text(\n"
     "                global.language('select_template_to_download'),"),
])

# ============================================================
# 4. permission_definition_screen.dart
# ============================================================
f4 = os.path.join(BASE, "screens", "config", "permission_definition_screen.dart")
replace_in_file(f4, [
    # Line 337 - const InputDecoration with global.language
    ("decoration: const InputDecoration(\n"
     "                  labelText: '${global.language(\"new_permission_code\")} *',",
     "decoration: InputDecoration(\n"
     "                  labelText: '${global.language(\"new_permission_code\")} *',"),
    # Line 355 - const InputDecoration with global.language
    ("decoration: const InputDecoration(\n"
     "                  labelText: '${global.language(\"new_permission_name\")} *',",
     "decoration: InputDecoration(\n"
     "                  labelText: '${global.language(\"new_permission_name\")} *',"),
    # Line 693 - const InputDecoration with global.language
    ("decoration: const InputDecoration(\n"
     "                        labelText: global.language('permission_name_label'),",
     "decoration: InputDecoration(\n"
     "                        labelText: global.language('permission_name_label'),"),
    # Line 723 - const Text with global.language
    ("const Text(\n"
     "                      global.language('set_permissions_by_branch'),",
     "Text(\n"
     "                      global.language('set_permissions_by_branch'),"),
])

# ============================================================
# 5. permission_link_screen.dart
# ============================================================
f5 = os.path.join(BASE, "screens", "config", "permission_link_screen.dart")
replace_in_file(f5, [
    # Line 129 - const Expanded with global.language
    ("const Expanded(\n"
     "                      child: Text(\n"
     "                        global.language('no_available_permissions_hint'),",
     "Expanded(\n"
     "                      child: Text(\n"
     "                        global.language('no_available_permissions_hint'),"),
    # Line 440 - const Text with global.language
    ("const Text(\n"
     "                        global.language('check_combined_permission'),",
     "Text(\n"
     "                        global.language('check_combined_permission'),"),
    # Line 810 - const Expanded with global.language
    ("const Expanded(\n"
     "                    child: Text(\n"
     "                      global.language('select_permission_info'),",
     "Expanded(\n"
     "                    child: Text(\n"
     "                      global.language('select_permission_info'),"),
    # Line 831 - const Text with global.language
    ("const Text(\n"
     "                            global.language('go_to_permission_definition'),",
     "Text(\n"
     "                            global.language('go_to_permission_definition'),"),
])

# ============================================================
# 6. user_screen.dart
# ============================================================
f6 = os.path.join(BASE, "screens", "config", "user_screen.dart")
replace_in_file(f6, [
    # Line 1477 - const InputDecoration with global.language
    ("decoration: const InputDecoration(\n"
     "                    labelText: global.language('job_position'),\n"
     "                    hintText: global.language('job_position_hint'),",
     "decoration: InputDecoration(\n"
     "                    labelText: global.language('job_position'),\n"
     "                    hintText: global.language('job_position_hint'),"),
])

# ============================================================
# 7. lineoa_config_screen.dart
# ============================================================
f7 = os.path.join(BASE, "screens", "lineoa", "lineoa_config_screen.dart")
replace_in_file(f7, [
    # Line 212 - Not a const issue, this is actually:
    # const Text(global.language('select_lineoa_type'))
    # Wait - let me check the exact error. Line 213 is:
    # global.language('select_lineoa_type')
    # The Text() here is not const, but the error is on line 213
    # Let me look more carefully at the context

    # Line 652 - const Text with global.language
    ("const Text(\n"
     "              global.language('connected_employees'),",
     "Text(\n"
     "              global.language('connected_employees'),"),
    # Line 756 - the Text at this line is not wrapped in const
    # It says error at line 756 but I see it's not const
    # Let me look at nearby const
])

print("\n=== Done! Const fixes applied ===")
