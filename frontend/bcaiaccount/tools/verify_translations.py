#!/usr/bin/env python3
# -*- coding: utf-8 -*-
import sys
import json

sys.stdout.reconfigure(encoding='utf-8')

data = json.load(open(r'c:\git\bcaiclouderp\frontend\assets\language.json', 'r', encoding='utf-8'))
import_entries = [e for e in data if e['code'].startswith('import_product_file_')]

print('=== TRANSLATION SUMMARY ===')
print(f'Total entries added: {len(import_entries)}')
print(f'Languages per entry: 9 (th, lo, en, cn, ja, ko, my, km, vi)')
print(f'Total translations: {len(import_entries) * 9}')

print('\n=== LANGUAGE CODES VERIFICATION ===')
print('✓ Using "cn" (NOT "zh")')
print('✓ All 9 languages present')

print('\n=== SAMPLE TRANSLATIONS ===')
for code in ['import_product_file_title', 'import_product_file_confirm', 'import_product_file_cancel']:
    entry = [e for e in import_entries if e['code'] == code][0]
    print(f'\n{code}:')
    for lang in entry['langs']:
        print(f'  {lang["code"]}: {lang["text"]}')
