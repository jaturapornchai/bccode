---
name: compact-menu-density
description: มาตรฐานการออกแบบเมนูดรอปดาวน์และ Navigation ให้ประหยัดพื้นที่แนวตั้ง (High Information Density / Compact Menu) ตามแบบฉบับ Desktop ERP (Champ Parity) ของ BC Ai Account เพื่อให้แสดงรายการได้ 15-20 แถวในหน้าจอเดียวโดยไม่ต้องเลื่อน scroll
---

# Compact Menu Density System — BC Ai Account

> บันทึกมาตรฐานเมื่อ 2026-10-10 ตามคำสั่งลุงจืด: *"เมนู ให้ ประหยัดบรรทัด จะได้แสดงได้เยอะๆ"*
> อ้างอิงคอมโพเนนต์หลัก: `frontend/src/app/menu/champ-menu-bar.tsx`

---

## 1. วัตถุประสงค์และที่มา (Core Principle & Rationale)

1. **ผู้ใช้ ERP และนักบัญชี (Champ Parity)**:
   - ผู้ใช้โปรแกรมบัญชีต้องการสแกนรายการเมนูทั้งหมดได้อย่างรวดเร็วในหน้าจอเดียว ไม่ต้องการให้เมนูยืดลงมาล้นขอบจอ หรือต้องเลื่อนเมาส์ scroll เพื่อหาเมนูที่ต้องการ
   - โปรแกรมบัญชีระดับเดสก์ท็อป (เช่น Champ 5.5, Express, Winspeed) มีเมนูที่กระชับมาก (ความสูงแถว 24–30px) ทำให้แสดงเมนูย่อย 15–20 รายการได้พร้อมกันครบถ้วน

2. **ปัญหาเดิม (Anti-pattern ที่แก้)**:
   - การบังคับ `min-h-[44px]` และ padding แนวตั้ง `py-1.5` ตามกฎปุ่มสัมผัสมือถือทั่วไป ทำให้เมนูที่มี 15 แถว กินความสูงเกิน 660px–700px ซึ่งยาวเกินความสูง viewport ของโน้ตบุ๊ก (768px–900px) และต้อง scroll ทันที
   - การใส่กล่องไอคอนสี่เหลี่ยมทึบ (`bg-muted rounded-md h-7 w-7` ขนาด 28px) ทำให้ความสูงขั้นต่ำของแถวถูกดันให้หนาขึ้น
   - การใช้ `line-clamp-2` หรือไม่ล็อกบรรทัด ทำให้บางเมนูตัดคำขึ้น 2 บรรทัดจนแถวบวมไม่สม่ำเสมอ
   - เมนูย่อย (Submenu Trigger) ที่ไม่มีไอคอน เกิดอาการข้อความเยื้องไปทางซ้ายกว่าแถวอื่น (ฟันปลา ไม่ตรงแนว)

---

## 2. กฎเหล็กความกระชับของเมนู (Compact Menu Rules)

### 2.1 ความสูงแถวและช่องไฟ (Row Height & Padding)
- **ความสูงแถว**: กำหนด `min-h-[30px]` และ padding `px-2 py-0.5`
- **ตัวหนังสือ**: ใช้ `text-[0.88rem]` พร้อม `leading-tight` (สระบน/ล่าง และวรรณยุกต์ไทยไม่ชนกัน และอ่านชัดเจนตามเกณฑ์คนไทย 40+)
- **CSS Class มาตรฐานของแถว (`ROW_CLASS`)**:
  ```tsx
  const ROW_CLASS = "min-h-[30px] gap-2 px-2 py-0.5 text-[0.88rem] leading-tight data-[disabled]:cursor-not-allowed";
  ```

### 2.2 โครงสร้างไอคอน (Icon Box & Alignment)
- **ไม่ใช้กรอบพื้นหลังทึบ**: ห้ามใส่ `bg-muted` หรือ `rounded-md` สี่เหลี่ยมล้อมรอบไอคอนในแถวเมนู
- **ขนาดไอคอนและคอนเทนเนอร์**:
  - คอนเทนเนอร์: `flex h-5 w-5 shrink-0 items-center justify-center text-muted-foreground`
  - ขนาดไอคอน SVG: `size={14}` (หรือ `h-3.5 w-3.5`)
- **การจัดแนวข้อความ (Indentation & Alignment)**:
  - เมนูย่อย (Submenu) ที่ไม่มีไอคอน **ต้องมี placeholder `<span className="flex h-5 w-5 shrink-0 items-center justify-center text-muted-foreground" aria-hidden="true" />` เสมอ** เพื่อให้ข้อความทุกแถวตรงแนวกัน ไม่เยื้องฟันปลา

### 2.3 ข้อความเมนูบรรทัดเดียว (Single-line Truncate)
- ป้ายชื่อเมนูต้องใช้ `truncate font-medium` บรรทัดเดียว ห้ามใช้ `line-clamp-2` หรือตัดคำขึ้นหลายบรรทัด
  ```tsx
  <span className="min-w-0 flex-1 truncate font-medium">
    {menuText(item.label, language, backendLanguage)}
  </span>
  ```

### 2.4 เส้นคั่นและการกำหนดขนาดแผง (Dividers & Panel Sizing)
- **เส้นคั่น (DropdownMenuSeparator)**: ใช้ `className="my-0.5"` เพื่อประหยัดช่องไฟระหว่างกลุ่มเมนู
- **ขนาดแผงเมนู (`PANEL_CLASS`)**:
  ```tsx
  const PANEL_CLASS =
    "max-h-[min(88vh,var(--radix-dropdown-menu-content-available-height))] max-w-[min(28rem,calc(100vw-1rem))] overflow-y-auto! p-1";
  ```
  - `p-1` (4px) รอบแผงเมนู (แทน `p-1.5`)
  - `max-h-[min(88vh,...)]` ใช้พื้นที่แนวตั้งสูงสุดเท่าที่จอมีได้โดยไม่หลุดขอบ

---

## 3. ตัวอย่างโค้ดอ้างอิง (Reference Implementation)

ดูโค้ดจริงที่: [`frontend/src/app/menu/champ-menu-bar.tsx`](file:///d:/bccode/frontend/src/app/menu/champ-menu-bar.tsx)

```tsx
// 1. กำหนดสไตล์ความหนาแน่นสูง
const ROW_CLASS = "min-h-[30px] gap-2 px-2 py-0.5 text-[0.88rem] leading-tight data-[disabled]:cursor-not-allowed";
const PANEL_CLASS =
  "max-h-[min(88vh,var(--radix-dropdown-menu-content-available-height))] max-w-[min(28rem,calc(100vw-1rem))] overflow-y-auto! p-1";

// 2. การเรนเดอร์ Divider
if (node.kind === "divider") {
  return <DropdownMenuSeparator key={`${keyPrefix}-divider-${index}`} className="my-0.5" />;
}

// 3. การเรนเดอร์ Submenu (มี placeholder ไอคอนเพื่อจัดแนว)
if (node.kind === "submenu") {
  return (
    <DropdownMenuSub key={node.id}>
      <DropdownMenuSubTrigger className={ROW_CLASS}>
        <span className="flex h-5 w-5 shrink-0 items-center justify-center text-muted-foreground" aria-hidden="true" />
        <span className="min-w-0 flex-1 truncate font-medium">{menuText(node.label, language, backendLanguage)}</span>
      </DropdownMenuSubTrigger>
      <DropdownMenuPortal>
        <DropdownMenuSubContent className={cn("min-w-[17rem]", PANEL_CLASS)} collisionPadding={8}>
          {renderNodes(node.nodes, node.id)}
        </DropdownMenuSubContent>
      </DropdownMenuPortal>
    </DropdownMenuSub>
  );
}

// 4. การเรนเดอร์ Action Item
if (node.kind === "action") {
  const Icon = ACTION_ICONS[node.id];
  return (
    <DropdownMenuItem className={ROW_CLASS} key={node.id} onSelect={() => onAction(node.id)}>
      <span className="flex h-5 w-5 shrink-0 items-center justify-center text-muted-foreground">
        <Icon aria-hidden="true" className="h-3.5 w-3.5" />
      </span>
      <span className="min-w-0 flex-1 truncate">{menuText(node.label, language, backendLanguage)}</span>
    </DropdownMenuItem>
  );
}

// 5. การเรนเดอร์ Menu Item ปกติ
return (
  <DropdownMenuItem
    className={ROW_CLASS}
    disabled={locked}
    key={item.id}
    onClick={(event) => {
      newTabRequested.current = event.ctrlKey || event.metaKey || event.shiftKey;
    }}
    onSelect={() => {
      const newTab = newTabRequested.current;
      newTabRequested.current = false;
      if (newTab) onOpenNewItem(item);
      else onOpenItem(item);
    }}
  >
    <span className="flex h-5 w-5 shrink-0 items-center justify-center text-muted-foreground">
      <MenuRouteIcon item={item} size={14} />
    </span>
    <span className="min-w-0 flex-1 truncate font-medium">{menuText(item.label, language, backendLanguage)}</span>
    <MenuPendingBadge route={item.route} language={language} backendLanguage={backendLanguage} />
    {locked ? <Lock aria-hidden="true" className="h-3.5 w-3.5 shrink-0 text-muted-foreground" /> : null}
  </DropdownMenuItem>
);
```

---

## 4. กับดักที่ห้ามทำซ้ำ (Anti-patterns to Avoid)

1. **ห้ามใช้ `min-h-[44px]` ใน Dropdown Menu**:
   - กฎ 44px+ มีไว้สำหรับปุ่ม CTA หลักบนหน้าจอทั่วไป แต่สำหรับ Dropdown/Context Menu/Navigation Menu ของระบบ ERP บนเดสก์ท็อป การใช้ 44px จะทำให้เมนูบวมจนแสดงได้เพียง 10 รายการก็ล้นจอ
2. **ห้ามใช้กล่องไอคอน `h-7 w-7` หรือพื้นหลังสีเทาเข้ม**:
   - กรอบสี่เหลี่ยมสีเทาดึงดูดสายตาเกินความจำเป็นและดันความสูงแถว ให้ใช้ไอคอนสะอาดตาขนาด 14px ลอยบนพื้นโปร่ง
3. **ห้ามลืม Placeholder ใน Submenu**:
   - ถ้าแถวหนึ่งมีไอคอน แต่อีกแถวไม่มีไอคอน ข้อความจะเหลื่อมกันดูไม่เป็นมืออาชีพ ให้ใช้ `<span className="h-5 w-5" />` รักษาตำแหน่งข้อความเสมอ
4. **ห้ามละเลยการอัปเดต SKILL**:
   - ทุกครั้งที่ปรับปรุง UI ระดับระบบ ต้องซิงก์ `.agents/skills/ui-scale-polish/SKILL.md` เสมอตามกฎเหล็กข้อ 0

---

## 5. เช็คลิสต์การตรวจรับ (Verification Checklist)

- [ ] เมนูที่มี 15–20 รายการ (เช่น สินค้าคงคลัง) แสดงผลได้ครบโดยไม่ต้อง scroll บนหน้าจอ 1080p และ 900p
- [ ] ความสูงแต่ละแถวไม่เกิน 32px (`min-h-[30px]`)
- [ ] ข้อความชื่อเมนูทุกแถว (ทั้งเมนูปกติ และเมนูย่อย Submenu) มี alignment ชิดซ้ายตรงกันเป๊ะ
- [ ] ตัวหนังสือภาษาไทยอ่านชัดเจน สระบนล่างไม่ทับกัน (`leading-tight`)
- [ ] คีย์ลัด คีย์บอร์ด (ลูกศรขึ้น/ลง/ซ้าย/ขวา/Enter/Esc) และ hover สลับหัวข้อทำงานลื่นไหลปกติ
- [ ] รัน `npm test -- src/lib/menu-bars.test.ts` และ `npx tsc --noEmit` ผ่าน 100%
