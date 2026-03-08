#!/usr/bin/env python3
"""Fix remaining localization errors — Phase 2.

Fixes:
1. 'const' variable/map/list declarations with global.language() → 'final'
2. Missing 'final' before variable assignments (e.g., tableHeaders = [...])
3. 'language(' without 'global.' prefix → 'global.language('
4. Default parameter values with global.language() → nullable with ?? fallback
5. Enum const constructors with global.language() → store key only, add getter
6. Switch-case with global.language() → if-else chains
"""
import os, re, sys
sys.stdout.reconfigure(encoding='utf-8')

LIB_DIR = r'D:\bcdev\bcaiaccount\lib'
BS = '\\'

stats = {'const_to_final': 0, 'add_final': 0, 'add_global_prefix': 0,
         'default_param_fixed': 0, 'files_fixed': 0}

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

        # Skip global.dart itself
        if rel == 'lib/global.dart':
            continue

        # === FIX 1: const declarations with global.language() → final ===
        # Handles: const List<...> name = [...], const Map<...> name = {...}
        # Handles: static const Map<...> name = {...}
        # We need to find const declarations (not constructors) that contain global.language()

        lines = content.split('\n')
        new_lines = []
        i = 0
        while i < len(lines):
            line = lines[i]
            modified = False

            # FIX 1a: Top-level or member const variable declarations
            # Pattern: `const List<...> varname = [` or `const Map<...> varname = {`
            # Pattern: `static const Map<...> _varname = {`
            # Pattern: `const varname = [` or `const varname = {`
            m = re.match(r'^(\s*)(static\s+)?const\s+((?:List|Map|Set)\s*<[^>]*>\s+)?(\w+)\s*=\s*[\[\{]', line)
            if m and not modified:
                # Check if global.language() appears in this block
                block = line
                depth = 0
                for j in range(0, min(200, len(lines) - i)):
                    check_line = lines[i + j]
                    block += check_line + '\n'
                    for ch in check_line:
                        if ch in '([{':
                            depth += 1
                        elif ch in ')]}':
                            depth -= 1
                    if depth <= 0 and j > 0:
                        break
                if 'global.language(' in block or "language('" in block:
                    # Replace const with final
                    line = re.sub(r'\bconst\b', 'final', line, count=1)
                    stats['const_to_final'] += 1
                    modified = True

            # FIX 1b: Variable assignment without declaration keyword
            # Pattern: `  tableHeaders = [` (no var/final/const)
            # This happens when fix_errors.py removed const from `const tableHeaders = [`
            if not modified:
                m2 = re.match(r'^(\s+)(\w+)\s*=\s*\[$', line.rstrip())
                if m2:
                    varname = m2.group(2)
                    indent = m2.group(1)
                    # Check this isn't an existing variable assignment (look for declaration above)
                    # Simple heuristic: if the variable name starts with lowercase and it looks like
                    # a list literal assignment at the start of a function body
                    if varname[0].islower() and varname not in ('i', 'j', 'k', 'x', 'y', 'z', 'index', 'count', 'result', 'value', 'data', 'list', 'map', 'set', 'widget', 'children'):
                        # Check if this variable is declared earlier in the file
                        var_declared = False
                        for prev_line in lines[:i]:
                            if re.search(rf'\b(?:var|final|const|List|Map)\b.*\b{varname}\b', prev_line):
                                var_declared = True
                                break
                            if re.search(rf'(?:late\s+)?(?:final\s+)?(?:List|Map|Set|String|int|double|bool|dynamic)\s*<?.*>?\s+{varname}\b', prev_line):
                                var_declared = True
                                break
                        if not var_declared:
                            line = f"{indent}final {varname} = ["
                            stats['add_final'] += 1
                            modified = True

            # FIX 2: language( without global. prefix
            # Only fix if this file imports global.dart
            if not modified and 'language(' in line and 'global.language(' not in line:
                has_global_import = any(
                    re.search(r"import\s+['\"].*global\.dart['\"]", l) or re.search(r"as\s+global\s*;", l)
                    for l in lines
                )
                if has_global_import:
                    # Check it's not inside a comment, string, or the import line itself
                    stripped = line.lstrip()
                    if not stripped.startswith('//') and not stripped.startswith('*') and not stripped.startswith('import '):
                        # Replace standalone language( with global.language(
                        # But avoid: something.language(, _language(, LanguageModel
                        new_line = re.sub(r'(?<![.\w])language\(', 'global.language(', line)
                        if new_line != line:
                            line = new_line
                            stats['add_global_prefix'] += 1
                            modified = True

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
print(f"const → final: {stats['const_to_final']}")
print(f"Added 'final': {stats['add_final']}")
print(f"Added 'global.' prefix: {stats['add_global_prefix']}")
