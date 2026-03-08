#!/usr/bin/env python3
"""Check remaining hardcoded Thai strings — aligned with localize_all.py logic."""
import os, re, sys
from collections import Counter
sys.stdout.reconfigure(encoding='utf-8')

lib = r'D:\bcdev\bcaiaccount\lib'
bs = '\\'
q = chr(39)
thai_re = re.compile('[\u0e00-\u0e7f]+')

# Same skip patterns as localize_all.py
skip_line_pats = [
    r'^\s*(//|/\*|\*)', r'AppLogger', r'debugPrint', r'print\(',
    r'LanguageDataModel\(', r'PublicNameModel\(', r'LanguageModel\(',
    r'NumberToWord', r'static const String _', r'languageCode:', r'codeTranslator:',
    r'code:\s*"th"', r'code:\s*"en"',
    r'Exception\(', r'throw ',
]

# Same sample patterns as localize_all.py
SAMPLE_PATTERNS = [
    r'สิธิพัช', r'เบอรรี่', r'วงเหลา', r'สมใจ', r'เนื่องแป้น', r'คำปลา', r'ดูดี', r'สามวง',
    r'บะหมี่กึ่งสำเร็จรูป', r'ซอสถั่วเหลือง', r'น้ำปลา', r'ข้าวสาร', r'น้ำตาลทราย', r'ผงซักฟอก',
    r'700บาท', r'600บาท', r'500บาท', r'1,800\.00 บาท', r'900\.00 บาท',
    r'บริษัท บ้านเชียง', r'ร้านโซลาว',
    r'displayNameThai',
]

SKIP_STRINGS = set(['฿','ไทย'] + list('มกราคม กุมภาพันธ์ มีนาคม เมษายน พฤษภาคม มิถุนายน กรกฎาคม สิงหาคม กันยายน ตุลาคม พฤศจิกายน ธันวาคม'.split()) + 'ม.ค. ก.พ. มี.ค. เม.ย. พ.ค. มิ.ย. ก.ค. ส.ค. ก.ย. ต.ค. พ.ย. ธ.ค.'.split() + 'จันทร์ อังคาร พุธ พฤหัสบดี ศุกร์ เสาร์ อาทิตย์'.split() + list('จ อ พ ศ ส'.split()) + ['พฤ','อา'] + 'บาท สตางค์ ศูนย์ เอ็ด ร้อย พัน หมื่น แสน ล้าน พันล้าน'.split() + 'หนึ่ง สอง สาม สี่ ห้า หก เจ็ด แปด เก้า สิบ'.split() + 'สิบเอ็ด สิบสอง สิบสาม สิบสี่ สิบห้า สิบหก สิบเจ็ด สิบแปด สิบเก้า ยี่สิบ'.split() + 'สามสิบ สี่สิบ ห้าสิบ หกสิบ เจ็ดสิบ แปดสิบ เก้าสิบ ถ้วน'.split())

LOGIN_FILES = {'login_password_screen.dart','login_screen.dart','registration.dart','select_language_screen.dart','login_line_screen.dart','select_shop.dart','server_config_screen.dart'}

interp = 0
plain = 0
skipd = 0
sample = 0
plain_unique = Counter()

for root, dirs, files in os.walk(lib):
    for fn in files:
        if not fn.endswith('.dart'): continue
        if fn in LOGIN_FILES: continue
        fp = os.path.join(root, fn)
        try: lines = open(fp, encoding='utf-8').readlines()
        except: continue
        for i, line in enumerate(lines, 1):
            if not thai_re.search(line): continue
            if any(re.search(p, line) for p in skip_line_pats): continue
            if any(re.search(p, line) for p in SAMPLE_PATTERNS):
                sample += 1
                continue
            code = line.split('//')[0] if '//' in line else line
            hits = re.findall('"([^"\n]*[\u0e00-\u0e7f][^"\n]*)"', code)
            hits += re.findall(q+'([^'+q+'\n]*[\u0e00-\u0e7f][^'+q+'\n]*)'+q, code)
            for h in hits:
                if h in SKIP_STRINGS or len(h) <= 1:
                    skipd += 1
                    continue
                if '${' in h or '$' in h:
                    interp += 1
                else:
                    plain += 1
                    plain_unique[h] += 1

print(f'Interpolated (${{...}} or $var): {interp}')
print(f'Plain (unmapped): {plain}')
print(f'Skipped (month/number): {skipd}')
print(f'Skipped (sample/mock data lines): {sample}')
print(f'Total remaining actionable: {interp + plain}')
print(f'\nTop 30 unmapped plain strings:')
for t, c in plain_unique.most_common(30):
    print(f'  [{c}x] "{t}"')
