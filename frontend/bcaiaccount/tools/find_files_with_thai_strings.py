#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
ค้นหาไฟล์ .dart ที่มี Thai strings 2-3 strings
"""

import os
import re
import sys
import json
import io

# Set UTF-8 encoding for Windows console
if sys.platform == 'win32':
    sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

def count_thai_strings_in_file(file_path):
    """นับจำนวน Thai strings ในไฟล์"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()

        # หา Thai strings ที่อยู่ใน quotes (single หรือ double)
        # ต้องมีอักษรไทยอย่างน้อย 1 ตัว และไม่ใช่แค่ comment
        thai_pattern = r'["\']([^"\']*[\u0E00-\u0E7F]+[^"\']*)["\']'
        matches = re.findall(thai_pattern, content)

        # กรองเฉพาะ strings ที่มีอักษรไทยจริงๆ และไม่ใช่แค่ตัวแปร
        thai_strings = []
        for match in matches:
            if re.search(r'[\u0E00-\u0E7F]', match):
                # ข้าม strings ที่เป็น variable interpolation หรือ empty
                if match.strip() and not match.startswith('$'):
                    thai_strings.append(match)

        # ตรวจสอบว่ามีการใช้ global.language() อยู่แล้วหรือไม่
        has_translation = 'global.language(' in content or 'language(' in content

        return len(thai_strings), thai_strings, has_translation
    except Exception as e:
        return 0, [], False

def should_skip_file(file_path):
    """ตรวจสอบว่าควรข้ามไฟล์นี้หรือไม่"""
    # ข้ามไฟล์ใน components/, widgets/
    if '/components/' in file_path or '\\components\\' in file_path:
        return True
    if '/widgets/' in file_path or '\\widgets\\' in file_path:
        return True

    # ข้ามไฟล์ที่มี "widget" ในชื่อ
    if 'widget' in os.path.basename(file_path).lower():
        return True

    return False

def main():
    # เริ่มต้นจากโฟลเดอร์ lib
    lib_dir = os.path.join(os.path.dirname(__file__), '..', 'lib')

    results = []

    # วนลูปหาไฟล์ .dart ทั้งหมด
    for root, dirs, files in os.walk(lib_dir):
        for file in files:
            if not file.endswith('.dart'):
                continue

            file_path = os.path.join(root, file)

            # ข้ามไฟล์ที่ควรข้าม
            if should_skip_file(file_path):
                continue

            count, strings, has_translation = count_thai_strings_in_file(file_path)

            # เฉพาะไฟล์ที่มี 2-3 Thai strings
            if 2 <= count <= 3:
                # ข้ามไฟล์ที่แปลงแล้ว (มีการใช้ global.language)
                if has_translation:
                    continue

                results.append({
                    'file': file_path.replace('\\', '/'),
                    'count': count,
                    'strings': strings
                })

    # เรียงลำดับตามจำนวน strings
    results.sort(key=lambda x: x['count'])

    # จำกัดแค่ 10 ไฟล์แรก
    results = results[:10]

    # แสดงผล
    print(f"Found {len(results)} files with 2-3 Thai strings (not yet translated):\n")
    for i, result in enumerate(results, 1):
        rel_path = result['file'].replace(lib_dir.replace('\\', '/'), 'lib')
        print(f"{i}. {rel_path}")
        print(f"   Count: {result['count']}")
        print(f"   Strings:")
        for s in result['strings']:
            print(f"     - {s}")
        print()

    # บันทึกผลลัพธ์เป็น JSON
    output_file = os.path.join(os.path.dirname(__file__), 'files_to_translate.json')
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(results, f, ensure_ascii=False, indent=2)

    print(f"Results saved to: {output_file}")

if __name__ == '__main__':
    main()
