import os
import re
from pathlib import Path

# Thai Unicode range
thai_pattern = re.compile(r'["\'][\u0E00-\u0E7F][^"\']*["\']')

# Directories and files to exclude
exclude_dirs = {'components', 'widgets', 'tools', '.git', 'build', 'android', 'ios', 'windows', 'linux', 'macos'}
exclude_files = {
    'lib/screens/option/option_list.dart',
    'lib/screens/product/product_list.dart',
    'lib/screens/purchase/purchase_list.dart',
    'lib/screens/config/line_notify_screen.dart',
    'lib/screens/config/reminder_screen.dart',
    'lib/screens/report/web_pdf_saver.dart'
}

def should_skip_file(file_path):
    """Check if file should be skipped"""
    rel_path = str(file_path).replace('\\', '/')

    # Check if already converted
    if any(rel_path.endswith(excluded) for excluded in exclude_files):
        return True

    # Check if in excluded directories
    parts = Path(rel_path).parts
    if any(part in exclude_dirs for part in parts):
        return True

    # Check if filename contains 'widget'
    if 'widget' in file_path.name.lower():
        return True

    return False

def count_thai_strings(file_path):
    """Count Thai strings in a file"""
    try:
        with open(file_path, 'r', encoding='utf-8') as f:
            content = f.read()
            matches = thai_pattern.findall(content)
            return len(matches), matches
    except Exception as e:
        return 0, []

def find_single_thai_files(root_dir, limit=10):
    """Find .dart files with exactly 1 Thai string"""
    root_path = Path(root_dir)
    results = []

    for dart_file in root_path.rglob('*.dart'):
        if should_skip_file(dart_file):
            continue

        count, matches = count_thai_strings(dart_file)

        if count == 1:
            rel_path = dart_file.relative_to(root_path)
            results.append({
                'file': str(rel_path).replace('\\', '/'),
                'thai_string': matches[0],
                'full_path': str(dart_file)
            })

            if len(results) >= limit:
                break

    return results

if __name__ == '__main__':
    import sys
    sys.stdout.reconfigure(encoding='utf-8')

    root = r'c:\git\bcaiclouderp\frontend'
    files = find_single_thai_files(root, limit=10)

    print(f"Found {len(files)} files with exactly 1 Thai string:\n")
    for i, file_info in enumerate(files, 1):
        print(f"{i}. {file_info['file']}")
        print(f"   Thai string: {file_info['thai_string']}")
        print(f"   Full path: {file_info['full_path']}\n")
