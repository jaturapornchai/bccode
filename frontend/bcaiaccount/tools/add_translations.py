#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
เครื่องมือเพิ่มคำแปลใหม่ลง language.json
"""

import json
import os
import sys

def add_translations_to_language_json(ui_translations_file, language_json_file):
    """
    เพิ่มคำแปลจากไฟล์ ui_translations ลง language.json
    """
    # อ่านไฟล์คำแปล UI/UX
    with open(ui_translations_file, 'r', encoding='utf-8') as f:
        ui_data = json.load(f)

    # อ่านไฟล์ language.json
    with open(language_json_file, 'r', encoding='utf-8') as f:
        lang_data = json.load(f)

    # สร้าง dict สำหรับเช็ค code ที่มีอยู่แล้ว
    existing_codes = {item['code'] for item in lang_data}

    added_count = 0
    skipped_count = 0

    # วนลูปเพิ่มคำแปลใหม่
    for translation in ui_data['translations']:
        code = translation['code']

        # ถ้ามี code นี้อยู่แล้ว ให้ข้าม
        if code in existing_codes:
            print(f"Skip: {code} (already exists)")
            skipped_count += 1
            continue

        # สร้างรายการภาษา
        langs = []
        for lang_code in ['th', 'lo', 'en', 'cn', 'ja', 'ko', 'my', 'km', 'vi']:
            if lang_code in translation:
                langs.append({
                    'code': lang_code,
                    'text': translation[lang_code]
                })

        # เพิ่มรายการใหม่
        new_entry = {
            'code': code,
            'langs': langs
        }
        lang_data.append(new_entry)
        print(f"Added: {code} ({len(langs)} languages)")
        added_count += 1

    # บันทึกไฟล์
    with open(language_json_file, 'w', encoding='utf-8') as f:
        json.dump(lang_data, f, ensure_ascii=False, indent=4)

    print(f"\nCompleted:")
    print(f"   - Added: {added_count} entries")
    print(f"   - Skipped: {skipped_count} entries")
    print(f"   - Total: {len(lang_data)} entries")

if __name__ == '__main__':
    script_dir = os.path.dirname(os.path.abspath(__file__))
    ui_trans_file = os.path.join(script_dir, 'ui_translations_to_add.json')
    lang_file = os.path.join(script_dir, '..', 'assets', 'language.json')

    if not os.path.exists(ui_trans_file):
        print(f"Error: File not found: {ui_trans_file}")
        sys.exit(1)

    if not os.path.exists(lang_file):
        print(f"Error: File not found: {lang_file}")
        sys.exit(1)

    print("Starting to add translations to language.json...")
    add_translations_to_language_json(ui_trans_file, lang_file)
