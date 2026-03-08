#!/usr/bin/env python3
"""Full list of remaining Thai strings (not just top 30)."""
import os, re, sys
from collections import Counter
sys.stdout.reconfigure(encoding='utf-8')

lib = r'D:\bcdev\bcaiaccount\lib'
bs = '\\'
q = chr(39)
thai_re = re.compile('[\u0e00-\u0e7f]+')
skip_line_pats = [r'^\s*(//|/\*|\*)', r'AppLogger', r'debugPrint', r'print\(', r'LanguageDataModel\(', r'PublicNameModel\(', r'LanguageModel\(', r'NumberToWord', r'static const String _', r'languageCode:', r'codeTranslator:', r'Exception\(', r'throw ']

SKIP_STRINGS = set(['฿','ไทย'] + list('มกราคม กุมภาพันธ์ มีนาคม เมษายน พฤษภาคม มิถุนายน กรกฎาคม สิงหาคม กันยายน ตุลาคม พฤศจิกายน ธันวาคม'.split()) + 'ม.ค. ก.พ. มี.ค. เม.ย. พ.ค. มิ.ย. ก.ค. ส.ค. ก.ย. ต.ค. พ.ย. ธ.ค.'.split() + 'จันทร์ อังคาร พุธ พฤหัสบดี ศุกร์ เสาร์ อาทิตย์'.split() + list('จ อ พ ศ ส'.split()) + ['พฤ','อา'] + 'บาท สตางค์ ศูนย์ เอ็ด ร้อย พัน หมื่น แสน ล้าน พันล้าน'.split() + 'หนึ่ง สอง สาม สี่ ห้า หก เจ็ด แปด เก้า สิบ'.split() + 'สิบเอ็ด สิบสอง สิบสาม สิบสี่ สิบห้า สิบหก สิบเจ็ด สิบแปด สิบเก้า ยี่สิบ'.split() + 'สามสิบ สี่สิบ ห้าสิบ หกสิบ เจ็ดสิบ แปดสิบ เก้าสิบ ถ้วน'.split())

LOGIN_FILES = {'login_password_screen.dart','login_screen.dart','registration.dart','select_language_screen.dart','login_line_screen.dart','select_shop.dart','server_config_screen.dart'}

plain_unique = Counter()
interp_unique = Counter()
plain_files = {}  # string -> set of files

for root, dirs, files in os.walk(lib):
    for fn in files:
        if not fn.endswith('.dart'): continue
        if fn in LOGIN_FILES: continue
        fp = os.path.join(root, fn)
        rel = fp.replace(lib + bs, '').replace(bs, '/')
        try: lines = open(fp, encoding='utf-8').readlines()
        except: continue
        for i, line in enumerate(lines, 1):
            if not thai_re.search(line): continue
            s = line.strip()
            if any(re.search(p, line) for p in skip_line_pats): continue
            code = line.split('//')[0] if '//' in line else line
            hits = re.findall('"([^"\n]*[\u0e00-\u0e7f][^"\n]*)"', code)
            hits += re.findall(q+'([^'+q+'\n]*[\u0e00-\u0e7f][^'+q+'\n]*)'+q, code)
            for h in hits:
                if h in SKIP_STRINGS or len(h) <= 1: continue
                if '${' in h or '$' in h:
                    interp_unique[h] += 1
                else:
                    plain_unique[h] += 1
                    if h not in plain_files:
                        plain_files[h] = set()
                    plain_files[h].add(rel)

print("=== ALL PLAIN UNMAPPED STRINGS (sorted by count) ===")
for t, c in plain_unique.most_common():
    files = ', '.join(sorted(plain_files[t])[:3])
    if len(plain_files[t]) > 3:
        files += f' +{len(plain_files[t])-3} more'
    print(f'  [{c:3d}x] "{t}" | {files}')

print(f"\n=== TOTAL: {sum(plain_unique.values())} plain, {sum(interp_unique.values())} interpolated ===")
print(f"\n=== TOP INTERPOLATED STRINGS ===")
for t, c in interp_unique.most_common(50):
    print(f'  [{c:3d}x] "{t}"')
