"""Render an RD fillable PDF with every widget boxed + numbered (for mapping fields to meaning).

usage: py tools/rdform/rdform_widgets.py <form.pdf> <out-prefix> [scale]
writes <out-prefix>-p<N>.png and <out-prefix>.widgets.json (index, page, name, kind, rect, maxlen, align, on)
"""
import json, sys
import pypdf, pypdfium2 as pdfium
from PIL import ImageDraw, ImageFont

def full_name(a):
    parts, node = [], a
    while node is not None:
        t = node.get('/T')
        if t is not None:
            parts.append(str(t))
        p = node.get('/Parent')
        node = p.get_object() if p is not None else None
    return '.'.join(reversed(parts))

def inherited(a, key):
    node = a
    while node is not None:
        if key in node:
            return node[key]
        p = node.get('/Parent')
        node = p.get_object() if p is not None else None
    return None

def widgets(path):
    r = pypdf.PdfReader(path)
    out = []
    for pi, page in enumerate(r.pages):
        for a in page.get('/Annots') or []:
            a = a.get_object()
            if a.get('/Subtype') != '/Widget':
                continue
            ft = inherited(a, '/FT')
            ff = int(inherited(a, '/Ff') or 0)
            kind = 'text'
            on = None
            if ft == '/Btn':
                kind = 'radio' if ff & (1 << 15) else ('button' if ff & (1 << 16) else 'check')
                states = [k for k in (a.get('/AP') or {}).get('/N', {}).keys() if k != '/Off']
                on = states[0][1:] if states else None
            elif ft == '/Tx' and ff & (1 << 24):
                kind = 'comb'
            ml = inherited(a, '/MaxLen')
            q = inherited(a, '/Q')
            out.append({'i': len(out), 'page': pi + 1, 'name': full_name(a), 'kind': kind,
                        'rect': [round(float(x), 1) for x in a['/Rect']],
                        'maxlen': int(ml) if ml is not None else None,
                        'align': int(q) if q is not None else 0, 'on': on})
    return out

def main():
    src, prefix = sys.argv[1], sys.argv[2]
    scale = float(sys.argv[3]) if len(sys.argv) > 3 else 2.0
    ws = widgets(src)
    json.dump(ws, open(prefix + '.widgets.json', 'w', encoding='utf-8'), ensure_ascii=False, indent=0)
    doc = pdfium.PdfDocument(src)
    font = ImageFont.load_default(size=int(9 * scale / 2 + 6))
    for pi in range(len(doc)):
        page = doc[pi]
        h = page.get_height()
        img = page.render(scale=scale).to_pil().convert('RGB')
        d = ImageDraw.Draw(img)
        for w in ws:
            if w['page'] != pi + 1:
                continue
            x0, y0, x1, y1 = w['rect']
            box = [x0 * scale, (h - y1) * scale, x1 * scale, (h - y0) * scale]
            color = (220, 0, 0) if w['kind'] in ('text', 'comb') else (0, 90, 220)
            d.rectangle(box, outline=color, width=1)
            d.text((box[0] + 1, box[1]), str(w['i']), fill=color, font=font)
        img.save(f"{prefix}-p{pi + 1}.png")
    print(f"{src}: {len(ws)} widgets, {len(doc)} pages")

if __name__ == '__main__':
    main()
