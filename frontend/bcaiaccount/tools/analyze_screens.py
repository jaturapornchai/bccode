#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Analyze Thai strings in lib/screens/ directory
"""

import re
import os
from pathlib import Path
import json

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
        thai_pattern = r'''['"]([^'"]*[\u0E00-\u0E7F]+[^'"]*)['"]'''
        matches = re.findall(thai_pattern, content)

        # Filter out empty strings and whitespace-only strings
        thai_strings = [s.strip() for s in matches if s.strip()]

        return len(thai_strings), thai_strings
    except Exception as e:
        return 0, []

def main():
    base_dir = Path(r'c:\git\bcaiclouderp\frontend')
    screens_dir = base_dir / 'lib' / 'screens'

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
    skip_dirs = {'components', 'widgets', 'cubit', 'utils', 'templates'}

    # Find all .dart files
    dart_files = []

    for root, dirs, files in os.walk(screens_dir):
        # Skip directories
        root_path = Path(root)
        parts = root_path.relative_to(screens_dir).parts if root_path != screens_dir else []
        if any(skip_dir in parts for skip_dir in skip_dirs):
            continue

        for file in files:
            if not file.endswith('.dart'):
                continue
            if file in skip_files:
                continue
            if '.bak' in file:
                continue

            file_path = root_path / file
            count, strings = count_thai_strings(file_path)

            if count > 0:
                rel_path = file_path.relative_to(base_dir)
                dart_files.append({
                    'count': count,
                    'path': str(rel_path).replace('\\', '/'),
                    'strings': strings
                })

    # Sort by count (ascending)
    dart_files.sort(key=lambda x: x['count'])

    # Print results
    print(f"Found {len(dart_files)} files with Thai strings in lib/screens/\n")
    print("Files sorted by number of Thai strings:\n")
    print(f"{'Count':<8} {'File Path'}")
    print("=" * 100)

    for item in dart_files[:40]:  # Show first 40
        print(f"{item['count']:<8} {item['path']}")

    print("\n" + "=" * 100)
    print(f"\nTotal files with Thai strings: {len(dart_files)}")

    # Show details of first 15 files
    print("\n\nDetailed view of first 15 files (candidates for translation):")
    print("=" * 100)
    try:
        for item in dart_files[:15]:
            print(f"\n{item['path']} ({item['count']} strings):")
            for i, s in enumerate(item['strings'][:10], 1):  # Show first 10 strings
                display_str = s[:80] + '...' if len(s) > 80 else s
                try:
                    print(f"  {i}. {display_str}")
                except UnicodeEncodeError:
                    print(f"  {i}. [Thai string - encoding issue]")
            if len(item['strings']) > 10:
                print(f"  ... and {len(item['strings']) - 10} more")
    except Exception as e:
        print(f"Error printing details: {e}")

    # Save to JSON
    output_file = base_dir / 'tools' / 'screens_analysis.json'
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(dart_files[:20], f, ensure_ascii=False, indent=2)

    print(f"\n\nTop 20 files saved to: {output_file}")

if __name__ == '__main__':
    main()
