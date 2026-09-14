"use client";

import { useState } from "react";
import { Download } from "lucide-react";
import { Button } from "@/components/ui/button";
import { GL_RESOURCES, type GLPage, type GLRecord } from "@/lib/general-ledger";
import { glAllRecords, glRequest } from "@/lib/general-ledger-api";
import { Notice, actionClass, downloadText, panel } from "./gl-common";
import { GLReports } from "./gl-reports";

export function GLExport() {
  const [busy, setBusy] = useState(false), [message, setMessage] = useState(""), [error, setError] = useState("");
  async function exportData() {
    if (busy) return;
    setBusy(true); setMessage(""); setError("");
    try {
      const first = await glRequest<GLPage<GLRecord>>("accounts?page=1&limit=1");
      const data: Record<string, GLRecord[]> = {};
      for (const resource of GL_RESOURCES) data[resource] = await glAllRecords(resource, "", 100000, first.sequence);
      const count = Object.values(data).reduce((total, records) => total + records.length, 0);
      downloadText(`บัญชีแยกประเภท-${new Date().toISOString().slice(0, 10)}.json`, JSON.stringify({ format: "bc-general-ledger-v2", exportedat: new Date().toISOString(), sequence: first.sequence, scope: "current-company", data }, null, 2), "application/json;charset=utf-8");
      setMessage(`ส่งออกข้อมูลบัญชีแยกประเภทครบ ${count.toLocaleString("th-TH")} รายการแล้ว`);
    } catch (e) { setError((e as Error).message); } finally { setBusy(false); }
  }
  return <section className={`${panel} grid gap-3`}><h2 className="text-lg font-semibold">ส่งออกข้อมูลบัญชีของบริษัทปัจจุบัน</h2><p className="text-[0.95rem] leading-relaxed">ไฟล์ประกอบด้วยผังบัญชี ปีบัญชี กลุ่มบัญชี รูปแบบเชื่อมโยง งบประมาณ งวด ประมาณการ และรายวันทุกหน้า เก็บจำนวนเงินเป็นข้อความทศนิยมสำหรับตรวจสอบต่อได้</p><Notice text="การส่งออกนี้ครอบคลุมเฉพาะข้อมูลบัญชีแยกประเภท ไม่รวมไฟล์แนบ ข้อมูลระบบอื่น ประวัติเหตุการณ์ หรือชุดกู้คืนฐานข้อมูลทั้งระบบ จำกัดทรัพยากรละ 100,000 รายการ และจะหยุดหากข้อมูลเปลี่ยนระหว่างส่งออก" /><Notice error text={error} /><Notice text={message} /><div><Button className={actionClass} disabled={busy} onClick={() => void exportData()}><Download />{busy ? "กำลังรวบรวมข้อมูล…" : "ส่งออกข้อมูลบัญชีทั้งหมด"}</Button></div></section>;
}
export function GLXbrlPreparation() {
  return <div className="grid gap-3"><section className={`${panel} grid gap-2`}><h2 className="text-lg font-semibold">เตรียมข้อมูลสำหรับยื่นงบการเงิน</h2><Notice text="ยังไม่เปิดสร้างไฟล์สำหรับยื่น DBD ต้องใช้แบบนำส่งและผังรายการของบริษัทที่ได้รับจาก DBD ก่อน จึงจะจับคู่และตรวจความถูกต้องของไฟล์ได้" /><p className="text-[0.95rem] leading-relaxed">ตรวจปีบัญชี ยอดบัญชี และกระดาษทำการด้านล่าง พร้อมส่งออกตารางเพื่อกระทบยอดกับแบบงบที่ใช้จริง ไฟล์ตารางนี้ใช้ตรวจสอบประกอบ และไม่ใช่ไฟล์ที่ผ่านการตรวจรับของ DBD</p></section><GLReports name="workingpaper" heading="กระดาษทำการสำหรับจัดทำงบ" /></div>;
}
