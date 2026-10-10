# ADR 2026-10-10 — ปรับเมนูดรอปดาวน์เป็นแบบกระชับ (Compact Density)

**สถานะ:** ทำแล้ว
**ผู้ตัดสิน:** ลุงจืด (คำสั่ง 2026-10-10: *"เมนู ให้ ประหยัดบรรทัด จะได้แสดงได้เยอะๆ"*)
**ต่อยอดจาก:** [2026-10-07-champ-style-menu-bar.md](2026-10-07-champ-style-menu-bar.md), Skill [`.agents/skills/compact-menu-density/SKILL.md`](../../../.agents/skills/compact-menu-density/SKILL.md)

---

## บริบทและปัญหา

หลังการเปลี่ยนเมนูหลักเป็นแบบ 13 หัว Champ ([ADR 2026-10-07](2026-10-07-champ-style-menu-bar.md)) ดรอปดาวน์ถูกกำหนดค่าเริ่มต้นตามมาตรฐานปุ่มทั่วไป (`min-h-[44px]`, `py-1.5`, ไอคอนในกล่อง `h-7 w-7`) ส่งผลให้:
1. แต่ละแถวกินความสูงมาก เมนูที่มีรายการ 15–20 แถว (เช่น หมวดสินค้าคงคลัง, จัดซื้อ, จัดขาย) กินความสูงเกิน 660–700px ทำให้ล้นจอ และผู้ใช้ต้อง scroll ดูรายการ
2. กล่องไอคอนสี่เหลี่ยมสีเทาทึบ (`bg-muted rounded-md h-7 w-7`) ดันความสูงแถวและดึงดูดสายตาเกินจำเป็น
3. แถวที่เป็นเมนูย่อย (Submenu Trigger เช่น "สินค้าชุด") ไม่มีไอคอน ทำให้ข้อความเยื้องฟันปลาไปทางซ้ายไม่ตรงกับแถวอื่น

---

## การตัดสินใจ

1. **ลดความสูงแถว (ประหยัดพื้นที่แนวตั้ง ~35%):**
   - ปรับ `ROW_CLASS` ใน `frontend/src/app/menu/champ-menu-bar.tsx` เป็น `min-h-[30px] gap-2 px-2 py-0.5 text-[0.88rem] leading-tight`
   - แผงดรอปดาวน์ขยายเป็น `max-h-[min(88vh,var(--radix-dropdown-menu-content-available-height))] p-1`
   - ส่งผลให้ในหน้าจอความละเอียดปกติ (1080p / 900p) สามารถแสดงเมนูได้ 15–20 รายการครบถ้วนในหน้าเดียวโดยไม่ต้อง scroll
2. **ไอคอนกะทัดรัด สะอาดตา (Clean Inline Icon):**
   - นำกรอบพื้นหลัง `bg-muted rounded-md h-7 w-7` ออก เปลี่ยนเป็นไอคอน SVG ขนาด 14px ภายในกรอบ `h-5 w-5 text-muted-foreground`
3. **จัดระเบียบแนวข้อความให้เสมอกัน (Uniform Alignment):**
   - แถวที่เป็น Submenu Trigger ใส่ placeholder `<span className="flex h-5 w-5 shrink-0" aria-hidden="true" />` นำหน้า เพื่อให้ข้อความทุกแถวตรงแนวเดียวกันเป๊ะ
4. **ตัดทอนข้อความบรรทัดเดียว (Single-line Truncate):**
   - เปลี่ยนจาก `line-clamp-2` เป็น `truncate` เพื่อป้องกันข้อความยาวดันแถวให้หนาขึ้น
   - เส้นคั่น (`DropdownMenuSeparator`) ใช้ `className="my-0.5"`
5. **สร้าง Skill ประจำโปรเจกต์:**
   - สร้างข้อกำหนดลงใน `.agents/skills/compact-menu-density/SKILL.md` เพื่อให้ AI ทุกตัวยึดถือเป็นมาตรฐานเดียวกันในการออกแบบ Navigation และ Dropdown
