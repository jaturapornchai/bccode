#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Language Manager Tool
จัดการ language.json และแปลภาษาอัตโนมัติ
"""

import json
import sys
from typing import Dict, List, Any

class LanguageManager:
    def __init__(self, json_file_path: str):
        self.json_file_path = json_file_path
        self.data: List[Dict[str, Any]] = []
        self.load()

    def load(self):
        """Load data from JSON file"""
        try:
            with open(self.json_file_path, 'r', encoding='utf-8') as f:
                self.data = json.load(f)
            print(f"Loaded successfully: {len(self.data)} entries")
        except Exception as e:
            print(f"Error loading file: {e}")
            sys.exit(1)

    def save(self):
        """Save data to JSON file"""
        try:
            with open(self.json_file_path, 'w', encoding='utf-8') as f:
                json.dump(self.data, f, ensure_ascii=False, indent=4)
            print(f"Saved successfully: {len(self.data)} entries")
        except Exception as e:
            print(f"Error saving file: {e}")

    def find_by_code(self, code: str) -> Dict[str, Any] | None:
        """ค้นหารายการตาม code"""
        for item in self.data:
            if item['code'] == code:
                return item
        return None

    def find_by_text(self, text: str, lang: str = 'th') -> Dict[str, Any] | None:
        """ค้นหารายการตาม text ในภาษาที่กำหนด"""
        for item in self.data:
            for lang_item in item['langs']:
                if lang_item['code'] == lang and lang_item['text'] == text:
                    return item
        return None

    def get_all_languages(self) -> List[str]:
        """ดึงรายการภาษาทั้งหมดที่มีในไฟล์"""
        if not self.data:
            return []

        languages = set()
        for item in self.data:
            for lang in item['langs']:
                languages.add(lang['code'])
        return sorted(list(languages))

    def add_language_to_all(self, lang_code: str, default_text: str = ""):
        """เพิ่มภาษาใหม่ให้กับทุกรายการ"""
        count = 0
        for item in self.data:
            # เช็คว่ามีภาษานี้อยู่แล้วหรือยัง
            has_lang = any(l['code'] == lang_code for l in item['langs'])
            if not has_lang:
                item['langs'].append({
                    'code': lang_code,
                    'text': default_text or f"{item['code']}_{lang_code}"
                })
                count += 1
        print(f"เพิ่มภาษา '{lang_code}' ให้กับ {count} รายการ")

    def add_new_entry(self, code: str, translations: Dict[str, str]):
        """เพิ่มรายการใหม่"""
        # เช็คว่ามี code นี้อยู่แล้วหรือยัง
        if self.find_by_code(code):
            print(f"มี code '{code}' อยู่แล้ว")
            return False

        new_entry = {
            'code': code,
            'langs': []
        }

        for lang_code, text in translations.items():
            new_entry['langs'].append({
                'code': lang_code,
                'text': text
            })

        self.data.append(new_entry)
        print(f"เพิ่มรายการใหม่: {code}")
        return True

    def update_translation(self, code: str, lang_code: str, text: str):
        """อัพเดทการแปลของรายการที่มีอยู่"""
        item = self.find_by_code(code)
        if not item:
            print(f"ไม่พบ code '{code}'")
            return False

        # หา lang ที่ต้องการอัพเดท
        for lang in item['langs']:
            if lang['code'] == lang_code:
                lang['text'] = text
                print(f"อัพเดท {code} [{lang_code}]: {text}")
                return True

        # ถ้ายังไม่มีภาษานี้ ให้เพิ่มใหม่
        item['langs'].append({
            'code': lang_code,
            'text': text
        })
        print(f"เพิ่มภาษาใหม่ {code} [{lang_code}]: {text}")
        return True

    def list_missing_translations(self, required_langs: List[str]):
        """แสดงรายการที่ยังแปลไม่ครบ"""
        missing = []
        for item in self.data:
            item_langs = {l['code'] for l in item['langs']}
            missing_langs = set(required_langs) - item_langs
            if missing_langs:
                missing.append({
                    'code': item['code'],
                    'missing': list(missing_langs)
                })
        return missing


if __name__ == '__main__':
    import os

    # ตัวอย่างการใช้งาน
    script_dir = os.path.dirname(os.path.abspath(__file__))
    json_path = os.path.join(script_dir, '..', 'assets', 'language.json')

    manager = LanguageManager(json_path)

    # แสดงภาษาทั้งหมดที่มี
    print("\nLanguages in file:")
    print(manager.get_all_languages())

    print(f"\nTotal entries: {len(manager.data)}")

    # เพิ่มภาษาใหม่
    #new_languages = ['ja', 'ko', 'my', 'km', 'vi']
    #print("\nAdding new languages:")
    #for lang in new_languages:
    #    manager.add_language_to_all(lang, "")

    # บันทึกไฟล์
    # manager.save()
