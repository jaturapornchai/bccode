#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""
Script to split language.json into separate language files
From: assets/language.json (2 MB, 9 languages)
To: assets/language/{lang}.json (220 KB per language)
"""
import json
import os
import sys

# Set UTF-8 encoding for Windows console
if sys.platform == 'win32':
    import io
    sys.stdout = io.TextIOWrapper(sys.stdout.buffer, encoding='utf-8', errors='replace')

# Languages to split
LANGUAGES = ['th', 'lo', 'en', 'cn', 'ja', 'ko', 'my', 'km', 'vi']

def split_language_files():
    """Split language.json into separate files per language"""

    # Read source file
    input_file = r'c:\git\bcaiclouderp\frontend\assets\language.json'
    output_dir = r'c:\git\bcaiclouderp\frontend\assets\language'

    print(f"Reading file {input_file}...")

    with open(input_file, 'r', encoding='utf-8') as f:
        data = json.load(f)

    print(f"Found {len(data)} entries")

    # Create dictionary for each language
    language_data = {lang: [] for lang in LANGUAGES}

    # Loop through and split data by language
    for entry in data:
        code = entry.get('code', '')
        langs = entry.get('langs', [])

        # Split each language
        for lang_entry in langs:
            lang_code = lang_entry.get('code', '')
            lang_text = lang_entry.get('text', '')

            if lang_code in LANGUAGES:
                language_data[lang_code].append({
                    'code': code,
                    'text': lang_text
                })

    # Save files by language
    for lang in LANGUAGES:
        output_file = os.path.join(output_dir, f'{lang}.json')

        with open(output_file, 'w', encoding='utf-8') as f:
            json.dump(language_data[lang], f, ensure_ascii=False, indent=2)

        file_size = os.path.getsize(output_file) / 1024  # KB
        print(f"  Created {lang}.json ({len(language_data[lang])} entries, {file_size:.1f} KB)")

    # Show summary
    print("\n" + "="*60)
    print("Summary:")
    print("="*60)

    original_size = os.path.getsize(input_file) / 1024 / 1024  # MB
    print(f"Original file: {original_size:.2f} MB")

    for lang in LANGUAGES:
        output_file = os.path.join(output_dir, f'{lang}.json')
        file_size = os.path.getsize(output_file) / 1024  # KB
        reduction = (1 - (file_size / 1024) / original_size) * 100
        print(f"  {lang}.json: {file_size:.1f} KB (reduced {reduction:.1f}%)")

    print(f"\nMemory saved: ~{((len(LANGUAGES) - 1) / len(LANGUAGES)) * 100:.1f}% when loading 1 language at a time")
    print("="*60)

if __name__ == '__main__':
    split_language_files()
