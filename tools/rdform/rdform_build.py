"""Build RD tax-form specs for the backend renderer (backend/internal/rdform).

A source spec (tools/rdform/specs/<code>.json, human-authored) names every fillable box of an official
RD PDF (mydocs/sample/*.pdf) by widget index — see tools/rdform/README.md. This script resolves each
widget index to its page/rect, detects the printed cell dividers inside the box (digit boxes, baht|satang
split) from a high-resolution render, copies the PDF template and writes the generated spec that the Go
renderer embeds.

usage:
  py tools/rdform/rdform_build.py --check [code ...]   validate only (default: all specs)
  py tools/rdform/rdform_build.py [code ...]           validate + write backend/internal/rdform/{assets,specs}
"""
import json, os, re, shutil, sys, statistics
import pypdfium2 as pdfium

sys.path.insert(0, os.path.dirname(__file__))
from rdform_widgets import widgets  # noqa: E402

ROOT = os.path.abspath(os.path.join(os.path.dirname(__file__), '..', '..'))
SRC_DIR = os.path.join(ROOT, 'tools', 'rdform', 'specs')
SAMPLE_DIR = os.path.join(ROOT, 'mydocs', 'sample')
OUT_DIR = os.path.join(ROOT, 'backend', 'internal', 'rdform')
TYPES = {'text', 'digits', 'taxid', 'money', 'int', 'check', 'choice'}
KEY_RE = re.compile(r'^[a-z][a-z0-9_]*$')
SCALE = 4.0


def fail(errors, msg):
    errors.append(msg)


def check_spec(spec, ws):
    errors, used = [], {}
    code = spec.get('code', '')
    if not KEY_RE.match(code):
        fail(errors, f'bad code {code!r}')
    pages = spec.get('pages') or []
    npages = max(w['page'] for w in ws) if ws else 1
    if not pages or any(p < 1 or p > npages for p in pages):
        fail(errors, f'pages {pages} outside 1..{npages}')

    def use(w, where):
        if not isinstance(w, int) or w < 0 or w >= len(ws):
            fail(errors, f'{where}: widget {w!r} does not exist')
            return
        if w in used:
            fail(errors, f'{where}: widget {w} already used by {used[w]}')
        used[w] = where
        if ws[w]['page'] not in pages:
            fail(errors, f'{where}: widget {w} is on page {ws[w]["page"]} which is not in pages {pages}')

    keys = set()
    for f in spec.get('fields', []):
        k = f.get('key', '')
        if not KEY_RE.match(k) or k in keys:
            fail(errors, f'field key {k!r} invalid or duplicated')
        keys.add(k)
        t = f.get('type')
        if t not in TYPES:
            fail(errors, f'{k}: type {t!r} not in {sorted(TYPES)}')
        if not f.get('label'):
            fail(errors, f'{k}: label missing')
        if t == 'choice':
            vals = set()
            for o in f.get('options', []):
                if not o.get('label') or str(o.get('value', '')) in vals or str(o.get('value', '')) == '':
                    fail(errors, f'{k}: option {o} needs unique value + label')
                vals.add(str(o.get('value')))
                use(o.get('w'), f'{k}={o.get("value")}')
            if not vals:
                fail(errors, f'{k}: choice without options')
        else:
            use(f.get('w'), k)
    table = spec.get('table')
    if table:
        tk = table.get('key', '')
        if not KEY_RE.match(tk) or tk in keys:
            fail(errors, f'table key {tk!r} invalid or clashes with a field')
        cols = {}
        for c in table.get('columns', []):
            ck = c.get('key', '')
            if not KEY_RE.match(ck) or ck in cols or c.get('type') not in TYPES - {'choice'} or not c.get('label'):
                fail(errors, f'table column {c} invalid')
            cols[ck] = c
        rows = table.get('rows') or []
        if not rows:
            fail(errors, 'table without rows')
        for ri, row in enumerate(rows):
            for ck, w in row.items():
                if ck not in cols:
                    fail(errors, f'row {ri + 1}: unknown column {ck!r}')
                use(w, f'{tk}[{ri + 1}].{ck}')
    unused = [w for w in ws if w['i'] not in used and w['page'] in pages and w['kind'] != 'button']
    return errors, unused


def dividers(gray, page_h, rect):
    """printed cell boundaries (pt) inside the widget box: returns (divider xs, cells).

    Vertical lines (solid box borders or dotted inner separators) split the box; a segment whose
    vertical middle is crossed by a horizontal line is the connector between two box groups (the
    "-" between tax-ID digit groups), not a cell, and is dropped."""
    x0, y0, x1, y1 = rect
    h = y1 - y0
    top = int((page_h - y1 + h * 0.22) * SCALE)
    bottom = int((page_h - y0 - h * 0.22) * SCALE)
    left, right = int((x0 - 6) * SCALE), int((x1 + 6) * SCALE)  # borders may sit just outside the widget
    if bottom - top < 3 or right - left < 3:
        return [], []
    px = gray.load()
    # ink = darker than the box background by 25 levels: pale green/pink rules on white and blue rules on a
    # light-blue filled box both count, the fill itself does not
    band = sorted(px[x, y] for x in range(left, right, 2) for y in range(top, bottom, 2))
    ink = min(230, band[len(band) // 2] - 25)
    lines, run = [], []
    for x in range(left, right):
        dark = sum(1 for y in range(top, bottom) if px[x, y] < ink)
        if dark >= 0.3 * (bottom - top):
            run.append(x)
        elif run:
            lines.append(run)
            run = []
    if run:
        lines.append(run)
    xs = [(r[0], r[-1]) for r in lines]
    mid = int((page_h - (y0 + y1) / 2) * SCALE)
    segs = []
    for (a0, a1), (b0, b1) in zip(xs, xs[1:]):
        inner = range(a1 + 2, b0 - 1)
        if len(inner) < 2:
            continue
        across = max(sum(1 for x in inner if px[x, y] < ink) for y in range(mid - 3, mid + 4))
        if across >= 0.6 * len(inner):
            continue  # connector between groups
        segs.append(((a0 + a1) / 2 / SCALE, (b0 + b1) / 2 / SCALE))
    divs = [round((r[0] + r[-1]) / 2 / SCALE, 2) for r in lines]
    return divs, [[round(a, 2), round(b, 2)] for a, b in segs]


def inside_cells(rect, segs):
    """edge slivers (page column rules just outside the widget, 3-5pt strips at the box border) are not
    cells: drop a cell lying mostly outside the widget or narrower than half the median cell."""
    x0, x1 = rect[0], rect[2]
    inside = [s for s in segs if min(s[1], x1) - max(s[0], x0) > 0.5 * (s[1] - s[0])]
    if inside:
        med = statistics.median(s[1] - s[0] for s in inside)
        inside = [s for s in inside if s[1] - s[0] >= 0.5 * med]
    return inside


def cells_from(rect, divs, want, segs):
    """segs that match the expected cell count (after dropping edge slivers)."""
    if not want:
        return None
    for cand in (inside_cells(rect, segs), segs):
        if len(cand) == want:
            return cand
    return None


def resolve_box(w, kind, gray, page_h, warnings, where):
    box = {'page': w['page'], 'rect': w['rect'], 'align': w['align']}
    if kind in ('digits', 'taxid', 'money', 'int', 'text'):
        divs, segs = dividers(gray[w['page']], page_h[w['page']], w['rect'])
        if kind == 'taxid':
            want = 13
            cells = cells_from(w['rect'], divs, want, segs)
            if cells is None:
                if w['maxlen'] in (13, 17, 18):
                    box['comb'] = w['maxlen']
                else:
                    warnings.append(f'{where}: taxid box has {len(divs)} dividers, no comb — text fallback')
            else:
                box['cells'] = cells
        elif kind == 'digits':
            cells = cells_from(w['rect'], divs, w['maxlen'], segs)
            if cells and len(cells) > 1:
                box['cells'] = cells
            elif w['maxlen']:
                box['comb'] = w['maxlen']
        elif kind == 'money':
            x0, x1 = w['rect'][0], w['rect'][2]
            cells = inside_cells(w['rect'], segs)
            runs = []  # touching cells; a connector "-" between boxes leaves a gap
            for c in cells:
                if runs and c[0] - runs[-1][-1][1] < 1.5:
                    runs[-1].append(c)
                else:
                    runs.append([c])
            width = x1 - x0
            narrow = lambda c: c[1] - c[0] < 0.2 * width
            if len(runs) >= 2 and len(runs[-1]) == 2 and len(cells) - 2 >= 3 and all(narrow(c) for c in cells):
                box['cells'] = cells  # one digit per cell: baht cells, "-", 2 satang cells
                box['satang'] = 2
            elif len(cells) >= 3 and narrow(cells[-1]) and narrow(cells[-2]):
                # satang box split into 2 digit cells by a dotted line: split before the pair
                box['split'] = cells[-2][0]
                if cells[-2][0] - cells[-3][1] > 1.5:  # "-" connector between the baht and satang boxes
                    box['baht_end'] = cells[-3][1]
            else:
                inner = [d for d in divs if x0 + 2 < d < x1 - 2]
                edge = [d for d in divs if d <= x1 + 1]  # ignore label glyphs right of the widget
                if edge and edge[-1] > x1 - 6:  # the printed right border sits inside the widget
                    inner = [d for d in inner if d < edge[-1] - 1]
                if inner and x1 - inner[-1] < 0.4 * width:
                    box['split'] = inner[-1]
    return box


def build(code, write):
    src = json.load(open(os.path.join(SRC_DIR, code + '.json'), encoding='utf-8'))
    pdf = os.path.join(SAMPLE_DIR, src['source'])
    ws = widgets(pdf)
    errors, unused = check_spec(src, ws)
    warnings = []
    if unused:
        warnings.append('unused widgets: ' + ', '.join(f"{w['i']}({w['kind']})" for w in unused))
    if errors or not write:
        return errors, warnings
    doc = pdfium.PdfDocument(pdf)
    gray, page_h, sizes = {}, {}, []
    for p in src['pages']:
        page = doc[p - 1]
        page_h[p] = page.get_height()
        gray[p] = page.render(scale=SCALE).to_pil().convert('L')
    for p in range(len(doc)):
        sizes.append([round(doc[p].get_width(), 2), round(doc[p].get_height(), 2)])
    out = {'code': src['code'], 'title': src['title'], 'template': src['code'] + '.pdf',
           'pages': src['pages'], 'sizes': [sizes[p - 1] for p in src['pages']], 'fields': []}
    for f in src.get('fields', []):
        g = {k: f[k] for k in ('key', 'type', 'label', 'group', 'note') if k in f}
        if f['type'] == 'choice':
            g['options'] = [{'value': str(o['value']), 'label': o['label'],
                             **resolve_box(ws[o['w']], 'check', gray, page_h, warnings, f['key'])}
                            for o in f['options']]
        else:
            g.update(resolve_box(ws[f['w']], f['type'], gray, page_h, warnings, f['key']))
        out['fields'].append(g)
    if src.get('table'):
        t = src['table']
        types = {c['key']: c['type'] for c in t['columns']}
        out['table'] = {'key': t['key'], 'label': t.get('label', ''), 'columns': t['columns'],
                        'rows': [{ck: resolve_box(ws[w], types[ck], gray, page_h, warnings, f'{t["key"]}[{ri + 1}].{ck}')
                                  for ck, w in row.items()} for ri, row in enumerate(t['rows'])]}
    os.makedirs(os.path.join(OUT_DIR, 'assets'), exist_ok=True)
    os.makedirs(os.path.join(OUT_DIR, 'specs'), exist_ok=True)
    shutil.copyfile(pdf, os.path.join(OUT_DIR, 'assets', src['code'] + '.pdf'))
    with open(os.path.join(OUT_DIR, 'specs', src['code'] + '.json'), 'w', encoding='utf-8', newline='\n') as fh:
        json.dump(out, fh, ensure_ascii=False, indent=1)
        fh.write('\n')
    return errors, warnings


def main():
    args = [a for a in sys.argv[1:] if not a.startswith('--')]
    write = '--check' not in sys.argv
    codes = args or sorted(f[:-5] for f in os.listdir(SRC_DIR) if f.endswith('.json'))
    bad = 0
    for code in codes:
        errors, warnings = build(code, write)
        status = 'FAIL' if errors else 'ok'
        print(f'{code}: {status}')
        for e in errors:
            print('  error:', e)
        for w in warnings:
            print('  warn:', w)
        bad += bool(errors)
    sys.exit(1 if bad else 0)


if __name__ == '__main__':
    main()
