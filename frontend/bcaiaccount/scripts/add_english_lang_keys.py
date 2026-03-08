#!/usr/bin/env python3
"""Add new language keys for hardcoded English text localization."""
import json
import sys

LANG_FILE = r"D:\bcdev\backend\assets\language\languages.json"

NEW_KEYS = {
    # Menu & Navigation
    "mcp_token": {"th": "MCP Token", "en": "MCP Token"},
    "bc_alert_agent": {"th": "BC Alert Agent", "en": "BC Alert Agent"},
    "ai_analysis": {"th": "วิเคราะห์ AI", "en": "AI Analysis"},
    "ai_chat_assistant": {"th": "ผู้ช่วย AI Chat", "en": "AI Chat Assistant"},

    # Time Picker
    "midnight": {"th": "เที่ยงคืน", "en": "Midnight"},
    "morning": {"th": "เช้า", "en": "Morning"},
    "noon": {"th": "เที่ยง", "en": "Noon"},
    "evening": {"th": "เย็น", "en": "Evening"},
    "quick_select": {"th": "เลือกด่วน", "en": "Quick Select"},
    "minutes": {"th": "นาที", "en": "Minutes"},

    # Bill Design - Alignment
    "align_left": {"th": "ซ้าย", "en": "Left"},
    "align_center": {"th": "กลาง", "en": "Center"},
    "align_right": {"th": "ขวา", "en": "Right"},

    # Bill Design - Text Style
    "text_bold": {"th": "ตัวหนา", "en": "Bold"},
    "text_italic": {"th": "ตัวเอียง", "en": "Italic"},
    "text_underline": {"th": "ขีดเส้นใต้", "en": "Underline"},

    # Bill Design - Sections
    "item_details": {"th": "รายละเอียดสินค้า", "en": "Item Details"},
    "extra_details": {"th": "รายละเอียดเพิ่มเติม", "en": "Extra Details"},
    "bill_command": {"th": "คำสั่ง", "en": "Command"},
    "bill_width": {"th": "ความกว้าง", "en": "Width"},
    "bill_size": {"th": "ขนาด", "en": "Size"},

    # Tooltips & Actions
    "export_pdf": {"th": "ส่งออก PDF", "en": "Export PDF"},
    "copy_to_clipboard": {"th": "คัดลอกไปคลิปบอร์ด", "en": "Copy to Clipboard"},
    "scanning_in_progress": {"th": "กำลังสแกน...", "en": "Scanning..."},
    "remove_filter": {"th": "ลบตัวกรอง", "en": "Remove Filter"},
    "add_filter": {"th": "เพิ่มตัวกรอง", "en": "Add Filter"},

    # LINE OA Config
    "channel_id": {"th": "Channel ID", "en": "Channel ID"},
    "channel_secret": {"th": "Channel Secret", "en": "Channel Secret"},
    "channel_access_token": {"th": "Channel Access Token", "en": "Channel Access Token"},

    # Form Labels
    "host": {"th": "โฮสต์", "en": "HOST"},
    "app_id_label": {"th": "APP ID", "en": "APP ID"},
    "pos_id": {"th": "POS ID", "en": "POS ID"},
    "backend_url": {"th": "Backend URL", "en": "Backend URL"},

    # Audit Screen
    "pgsql_count": {"th": "จำนวน PgSql", "en": "PgSql Count"},
    "clickhouse_count": {"th": "จำนวน ClickHouse", "en": "ClickHouse Count"},
    "pgsql_qty": {"th": "จำนวน PgSql", "en": "PgSql Qty"},
    "clickhouse_qty": {"th": "จำนวน ClickHouse", "en": "ClickHouse Qty"},
    "pgsql_amount": {"th": "ยอดเงิน PgSql", "en": "PgSql Amount"},
    "clickhouse_amount": {"th": "ยอดเงิน ClickHouse", "en": "ClickHouse Amount"},

    # Dashboard / Database
    "no_data": {"th": "ไม่มีข้อมูล", "en": "No data"},

    # MCP API Key
    "example_claude_desktop_key": {"th": "เช่น Claude Desktop Key", "en": "e.g., Claude Desktop Key"},
    "role_readonly": {"th": "อ่านอย่างเดียว", "en": "Readonly"},
    "role_developer": {"th": "นักพัฒนา", "en": "Developer"},
    "role_custom": {"th": "กำหนดเอง", "en": "Custom"},

    # Validation
    "this_field_is_required": {"th": "กรุณากรอกข้อมูล", "en": "This field is required"},

    # Export
    "export_excel": {"th": "ส่งออก Excel", "en": "Export Excel"},

    # Currency
    "currency_code_example": {"th": "USD, EUR, JPY, CNY เป็นต้น", "en": "USD, EUR, JPY, CNY, etc."},
    "currency_name_example": {"th": "ดอลลาร์สหรัฐ, ยูโร, เยน เป็นต้น", "en": "US Dollar, Euro, Japanese Yen, etc."},

    # PDF Watermarks
    "watermark_premium": {"th": "พรีเมียม", "en": "PREMIUM"},
    "watermark_confidential": {"th": "ลับ", "en": "CONFIDENTIAL"},

    # Product
    "crop_image": {"th": "ครอปรูปภาพ", "en": "Crop Image"},
    "json_code_hint": {"th": "JSON Code...", "en": "JSON Code..."},

    # Error
    "error_loading_pdf": {"th": "เกิดข้อผิดพลาดในการโหลด PDF", "en": "Error loading PDF"},
    "error_infinite_loop": {"th": "ตรวจพบลูปซ้ำไม่สิ้นสุด", "en": "Error: Infinite Loop Detected"},

    # Copy UAT
    "collection_name": {"th": "คอลเลกชัน", "en": "Collection"},

    # LINE OA QR
    "line_oa_qrcode": {"th": "Line OA Qrcode", "en": "Line OA QR Code"},

    # Line number column
    "line_number": {"th": "ลำดับ", "en": "Line"},

    # QR Payment
    "qr_payment": {"th": "QR ชำระเงิน", "en": "QR Payment"},
    "line_notify_menu": {"th": "Line Notify", "en": "Line Notify"},
    "line_oa_menu": {"th": "Line OA", "en": "Line OA"},
}

def main():
    with open(LANG_FILE, "r", encoding="utf-8") as f:
        data = json.load(f)

    added = 0
    skipped = 0
    all_langs = ["th", "en", "cn", "ja", "km", "ko", "lo", "my", "vi"]

    for key, translations in NEW_KEYS.items():
        if key in data:
            skipped += 1
            continue

        entry = {}
        for lang in all_langs:
            if lang in translations:
                entry[lang] = translations[lang]
            else:
                entry[lang] = translations.get("en", "")
        data[key] = entry
        added += 1

    sorted_data = dict(sorted(data.items()))

    with open(LANG_FILE, "w", encoding="utf-8") as f:
        json.dump(sorted_data, f, ensure_ascii=False, indent=2)
        f.write("\n")

    print(f"Added: {added}, Skipped (already exist): {skipped}")
    print(f"Total keys: {len(sorted_data)}")

if __name__ == "__main__":
    main()
