"use client";

import { Input } from "@/components/ui/input";
import { type Product } from "@/lib/product-barcode/types";

export function TabProductLogistics({
  value,
  onChange,
}: {
  value: Product;
  onChange: (next: Product) => void;
}) {
  return (
    <div className="space-y-4">
      <div className="border border-border bg-card rounded-lg p-4 space-y-4">
        <h4 className="font-bold text-base text-foreground border-b border-border pb-2">ข้อมูลขนส่งและขนาดพัสดุ (Logistics & Shipping)</h4>
        <div className="grid gap-4 sm:grid-cols-2">
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">น้ำหนักพัสดุรวมกล่อง (kg)</label>
            <Input
              type="number"
              min={0}
              step="any"
              value={value.packageweight ?? 0}
              onChange={(e) => onChange({ ...value, packageweight: Math.max(0, Number(e.target.value) || 0) })}
            />
          </div>
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">น้ำหนักเชิงปริมาตรประเมิน (kg)</label>
            <Input
              type="text"
              readOnly
              className="bg-muted/40 font-mono"
              value={`${(((value.packagewidth ?? 0) * (value.packagelength ?? 0) * (value.packageheight ?? 0)) / 5000).toFixed(3)} kg`}
            />
            <p className="text-[10px] text-muted-foreground mt-1">คำนวณจาก (กว้าง x ยาว x สูง) / 5000</p>
          </div>
        </div>

        <div className="grid gap-4 sm:grid-cols-3">
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">ความกว้างกล่อง (cm)</label>
            <Input
              type="number"
              min={0}
              value={value.packagewidth ?? 0}
              onChange={(e) => onChange({ ...value, packagewidth: Math.max(0, Number(e.target.value) || 0) })}
            />
          </div>
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">ความยาวกล่อง (cm)</label>
            <Input
              type="number"
              min={0}
              value={value.packagelength ?? 0}
              onChange={(e) => onChange({ ...value, packagelength: Math.max(0, Number(e.target.value) || 0) })}
            />
          </div>
          <div className="space-y-1">
            <label className="text-xs text-muted-foreground">ความสูงกล่อง (cm)</label>
            <Input
              type="number"
              min={0}
              value={value.packageheight ?? 0}
              onChange={(e) => onChange({ ...value, packageheight: Math.max(0, Number(e.target.value) || 0) })}
            />
          </div>
        </div>

        <div className="p-3 bg-muted/20 border border-border rounded-lg space-y-3">
          <p className="text-xs font-semibold text-muted-foreground">คุณลักษณะการจัดส่งและพิมพ์ฉลาก (Shipping Badges):</p>
          <div className="grid grid-cols-2 gap-3 text-xs">
            <label className="flex items-center gap-2 cursor-pointer">
              <input
                type="checkbox"
                checked={value.isalert ?? false}
                onChange={(e) => onChange({ ...value, isalert: e.target.checked })}
                className="rounded accent-primary size-4"
              />
              <div>
                <span className="font-semibold block">สินค้าแตกหักง่าย / ระวังแตก (Fragile)</span>
                <span className="text-[10px] text-muted-foreground">ติดป้ายเตือนและพิมพ์สติ๊กเกอร์เตือนพิเศษ</span>
              </div>
            </label>
          </div>
          {value.isalert && (
            <div className="space-y-1">
              <label className="text-xs text-muted-foreground">คำเตือนสำหรับสติ๊กเกอร์จัดส่ง</label>
              <Input
                placeholder="ระบุข้อความ เช่น ห้ามโยน ระวังของแตกหักง่าย"
                value={value.alertdescription || ""}
                onChange={(e) => onChange({ ...value, alertdescription: e.target.value })}
              />
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
