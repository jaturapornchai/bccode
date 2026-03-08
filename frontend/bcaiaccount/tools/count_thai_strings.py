#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Script to count Thai strings in Dart files
Excludes comments and only counts UI strings
"""

import re
import sys
from pathlib import Path

def count_thai_strings(file_path):
    """Count Thai strings in a Dart file, excluding comments"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # Remove single-line comments
        content = re.sub(r'//.*?$', '', content, flags=re.MULTILINE)

        # Remove multi-line comments
        content = re.sub(r'/\*.*?\*/', '', content, flags=re.DOTALL)

        # Find all Thai strings (between quotes)
        # Match Thai Unicode range: \u0E00-\u0E7F
        thai_pattern = r'''['"]([^'"]*[\u0E00-\u0E7F]+[^'"]*)['"]'''
        matches = re.findall(thai_pattern, content)

        # Filter out empty strings and whitespace-only strings
        thai_strings = [s.strip() for s in matches if s.strip()]

        return len(thai_strings), thai_strings
    except Exception as e:
        return 0, []

def main():
    screens_dir = Path('/c/git/bcaiclouderp/frontend/lib/screens')

    # Files to skip (already converted)
    skip_files = {
        'document_product_list_widget.dart',
        'option_list.dart', 'product_list.dart', 'purchase_list.dart',
        'line_notify_screen.dart', 'reminder_screen.dart',
        'desktop_pdf_saver.dart', 'web_pdf_saver.dart',
        'select_language_screen.dart', 'login_screen.dart', 'login_with.dart', 'select_branch_screen.dart',
        'company_screen.dart', 'config_screen.dart', 'customer_screen.dart',
        'doc_format_screen.dart', 'employee_screen.dart', 'kitchen_screen.dart', 'otp_screen.dart',
        'add_product_to_branch_screen.dart', 'add_product_to_department_screen.dart',
        'add_product_to_kitchen_screen.dart', 'table_order_screen.dart',
        'qrcode_order_screen.dart', 'report_product_balance.dart', 'database_info.dart',
    }

    # Directories to skip
    skip_dirs = {'components', 'widgets'}

    # Find all .dart files
    dart_files = []
    for dart_file in screens_dir.rglob('*.dart'):
        # Check if file is in skip list
        if dart_file.name in skip_files:
            continue

        # Check if file is in a skipped directory
        parts = dart_file.relative_to(screens_dir).parts
        if any(skip_dir in parts for skip_dir in skip_dirs):
            continue

        # Check if it's a .bak file
        if dart_file.suffix == '.bak' or '.bak' in dart_file.name:
            continue

        count, strings = count_thai_strings(dart_file)
        if count > 0:
            rel_path = dart_file.relative_to(Path('/c/git/bcaiclouderp/frontend'))
            dart_files.append((count, str(rel_path), strings))

    # Sort by count (ascending)
    dart_files.sort(key=lambda x: x[0])

    # Print results
    print(f"Found {len(dart_files)} files with Thai strings in lib/screens/\n")
    print("Files sorted by number of Thai strings:\n")
    print(f"{'Count':<8} {'File Path'}")
    print("=" * 80)

    for count, filepath, strings in dart_files[:30]:  # Show first 30
        print(f"{count:<8} {filepath}")

    print("\n" + "=" * 80)
    print(f"\nTotal files with Thai strings: {len(dart_files)}")

    # Show details of first 10 files
    print("\n\nDetailed view of first 10 files:")
    print("=" * 80)
    for count, filepath, strings in dart_files[:10]:
        print(f"\n{filepath} ({count} strings):")
        for i, s in enumerate(strings[:5], 1):  # Show first 5 strings
            print(f"  {i}. {s}")
        if len(strings) > 5:
            print(f"  ... and {len(strings) - 5} more")

if __name__ == '__main__':
    main()
