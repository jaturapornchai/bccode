#!/usr/bin/env python3
"""Fix all localization-related errors automatically.

Fixes:
1. Missing 'import ...global.dart as global;' in files using global.language()
2. Removes 'const' keywords that create const contexts containing global.language()
3. Handles both same-line and multi-line const patterns
"""
import os, re, sys
sys.stdout.reconfigure(encoding='utf-8')

LIB_DIR = r'D:\bcdev\bcaiaccount\lib'
BS = '\\'

stats = {'import_added': 0, 'const_removed': 0, 'files_fixed': 0}

for root, dirs, files in os.walk(LIB_DIR):
    for fn in files:
        if not fn.endswith('.dart'):
            continue
        fp = os.path.join(root, fn)
        rel = fp.replace(LIB_DIR + BS, 'lib/').replace(BS, '/')

        try:
            content = open(fp, encoding='utf-8').read()
            original = content
        except:
            continue

        # Skip if no global.language in file
        if 'global.language(' not in content:
            continue

        # Skip global.dart itself (already fixed)
        if rel == 'lib/global.dart':
            continue

        lines = content.split('\n')

        # === FIX 1: Check if 'global' is imported ===
        has_global_import = False
        for line in lines:
            if re.search(r"import\s+['\"].*global\.dart['\"]", line):
                has_global_import = True
                break
            if re.search(r"as\s+global\s*;", line):
                has_global_import = True
                break

        if not has_global_import:
            parts = rel.replace('lib/', '').split('/')
            depth = len(parts) - 1
            prefix = '../' * depth
            import_line = f"import '{prefix}global.dart' as global;"

            # Find last import line
            last_import_idx = -1
            for i, line in enumerate(lines):
                stripped = line.strip()
                if stripped.startswith('import '):
                    last_import_idx = i

            if last_import_idx >= 0:
                lines.insert(last_import_idx + 1, import_line)
                stats['import_added'] += 1

        # === FIX 2: Remove 'const' in const-contexts containing global.language() ===
        # Strategy:
        # A) Lines containing global.language() — remove ALL const keywords on that line
        # B) Lines with 'const ' but no global.language() — look ahead up to 20 lines;
        #    if global.language() appears while we're still inside the constructor, remove const

        new_lines = []
        i = 0
        while i < len(lines):
            line = lines[i]

            if 'global.language(' in line and 'const ' in line:
                # Case A: Same line has both const and global.language()
                # Remove all 'const' keywords on this line
                new_line = re.sub(r'\bconst\b', '', line)
                # Clean up double spaces left behind
                new_line = re.sub(r'  +', ' ', new_line)
                # Fix leading whitespace (don't collapse indentation to single space)
                orig_indent = len(line) - len(line.lstrip())
                new_indent = len(new_line) - len(new_line.lstrip())
                if new_indent < orig_indent:
                    new_line = ' ' * orig_indent + new_line.lstrip()
                if new_line != line:
                    stats['const_removed'] += 1
                new_lines.append(new_line)

            elif 'const ' in line and 'global.language(' not in line:
                # Case B: Line has const but no global.language()
                # Look ahead to see if global.language() appears in the constructor body
                block_text = ''
                found_language = False
                depth = 0
                # Count brackets from this line onwards
                for j in range(0, min(25, len(lines) - i)):
                    check_line = lines[i + j]
                    block_text += check_line + '\n'
                    for ch in check_line:
                        if ch == '(':
                            depth += 1
                        elif ch == ')':
                            depth -= 1
                    if 'global.language(' in check_line and j > 0:
                        found_language = True
                        break
                    # If brackets are balanced and we've opened at least one, stop
                    if depth <= 0 and j > 0:
                        break

                if found_language:
                    # Remove const keywords on this line
                    new_line = re.sub(r'\bconst\b', '', line)
                    new_line = re.sub(r'  +', ' ', new_line)
                    orig_indent = len(line) - len(line.lstrip())
                    new_indent = len(new_line) - len(new_line.lstrip())
                    if new_indent < orig_indent:
                        new_line = ' ' * orig_indent + new_line.lstrip()
                    if new_line != line:
                        stats['const_removed'] += 1
                    new_lines.append(new_line)
                else:
                    new_lines.append(line)
            else:
                new_lines.append(line)
            i += 1

        content = '\n'.join(new_lines)

        if content != original:
            with open(fp, 'w', encoding='utf-8') as f:
                f.write(content)
            stats['files_fixed'] += 1
            print(f"  [fixed] {rel}")

print(f"\n=== Summary ===")
print(f"Files fixed: {stats['files_fixed']}")
print(f"Imports added: {stats['import_added']}")
print(f"Const keywords removed: {stats['const_removed']}")
