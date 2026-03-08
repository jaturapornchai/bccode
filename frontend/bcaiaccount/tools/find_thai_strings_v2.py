#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
ค้นหาไฟล์ .dart ที่มี Thai strings 2-3 strings
Version 2: ปรับปรุงการค้นหา Thai strings
"""

import os
import re
import sys
import json
import io

# Set UTF-8 encoding for Windows console
if sys.platform == 'win32':
    sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8')

def extract_thai_strings(file_path):
    """ดึง Thai strings จากไฟล์"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()

        thai_strings = []

        # Pattern สำหรับ strings ในเครื่องหมาย quotes
        # จับ strings ที่มีอักษรไทย และไม่ใช่แค่ comment
        patterns = [
            r'"([^"]*[\u0E00-\u0E7F]+[^"]*)"',  # Double quotes
            r"'([^']*[\u0E00-\u0E7F]+[^']*)'",  # Single quotes
        ]

        for pattern in patterns:
            matches = re.finditer(pattern, content)
            for match in matches:
                text = match.group(1)
                # กรองเฉพาะที่มีอักษรไทยจริงๆ
                if re.search(r'[\u0E00-\u0E7F]', text):
                    # ข้าม empty strings และ variable interpolation
                    if text.strip() and not text.startswith('$') and len(text) > 0:
                        # ข้าม strings ที่ยาวเกินไป (อาจเป็น code)
                        if len(text) < 200:
                            thai_strings.append(text.strip())

        # ลบ duplicates แต่เก็บลำดับ
        seen = set()
        unique_strings = []
        for s in thai_strings:
            if s not in seen:
                seen.add(s)
                unique_strings.append(s)

        # ตรวจสอบว่ามีการใช้ global.language() อยู่แล้วหรือไม่
        has_translation = 'global.language(' in content or 'language(' in content

        return unique_strings, has_translation
    except Exception as e:
        print(f"Error reading {file_path}: {e}")
        return [], False

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
    lib_dir = os.path.abspath(lib_dir)

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

            strings, has_translation = extract_thai_strings(file_path)

            # เฉพาะไฟล์ที่มี 2-3 Thai strings
            if 2 <= len(strings) <= 3:
                # ข้ามไฟล์ที่แปลงแล้ว
                if has_translation:
                    continue

                results.append({
                    'file': file_path.replace('\\', '/'),
                    'count': len(strings),
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
            # จำกัดความยาวของ string ที่แสดง
            display_str = s if len(s) < 80 else s[:77] + '...'
            print(f"     - \"{display_str}\"")
        print()

    # บันทึกผลลัพธ์เป็น JSON
    output_file = os.path.join(os.path.dirname(__file__), 'files_to_translate.json')
    with open(output_file, 'w', encoding='utf-8') as f:
        json.dump(results, f, ensure_ascii=False, indent=2)

    print(f"Results saved to: {output_file}")

if __name__ == '__main__':
    main()
