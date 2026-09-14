---
date: 2026-09-11
status: accepted
tags: [bc-account, deployment, frontend, champ]
---

# Deploy เมนู Champ และเมนูบนเป็นค่าเริ่มต้น

## วัตถุประสงค์

ตามคำสั่งลุงจืด deploy ไป https://account.bcaicloud.com/ นำ frontend จาก working tree ล่าสุดขึ้น production: 223 เมนู, ชื่อหมวดไทยล้วน และเมนูบนเป็นค่าเริ่มต้น ยังจำตัวเลือกเมนูซ้ายของผู้ใช้เดิม

หลักฐาน implementation: `frontend/src/lib/menu-data.ts:73`, `frontend/src/app/menu/main-menu-screen.tsx:414` และ `frontend/e2e/menu-consistency.spec.ts:17`

## Release ที่ใช้งานจริง

- Host: DigitalOcean `159.223.43.229`; compose project `bcai-account`
- Frontend ใหม่: `bcai-account-frontend:r20260911-menu-top-1`
- Image ID: `sha256:1ec6ae43ee1323f60e975e7d8ac3007206f1eafec33564dff78ce2c0e638a02e`
- Frontend ก่อนหน้า: `bcai-account-frontend:r20260909-1`
- Source base: `a04bfb57` + working tree ที่มีงานค้างอยู่ก่อนเซสชัน; ไม่ได้สร้าง commit หรือ push ในงาน deploy นี้
- Runtime: `bcai-account-frontend-1` เป็น `healthy`; HTTPS สาธารณะตอบ 200; mainapi/worker ใช้ `r20260902-11` ตามเดิมและไม่ได้ recreate

## Workflow, config และ dependency

ใช้ Dockerfile เดิม `frontend/Dockerfile`: build บนเครื่อง Windows ด้วย Google public client ID เดิมจาก runtime production และ `BCAI_LOCAL_BACKEND_URL=http://mainapi:8888` ไม่คัดลอก `.env` เข้า image

1. ตรวจ code map, frontend, backend, outbox และ projection
2. Build Docker image แล้ว `docker save` ส่ง archive ไป `/opt/bcai-account/releases/r20260911-menu-top-1/frontend.tar`
3. SHA-256 ของ archive ทั้งสองเครื่องตรงกัน: `2971130b5e2c32ed7287a538e5a284d21c05ba140d70fafc4c51895eb85dbc2d`
4. เปิด preview container ด้วย env เดิมและ network `bcai-account_edge` ที่ loopback port 3201; ตรวจ HTTP 200 และ Playwright ผ่าน SSH tunnel ก่อนสลับ
5. สำรอง `/etc/bcai-account/release.env` ไว้ที่ `/opt/bcai-account/releases/r20260911-menu-top-1/release.env.before` บนเซิร์ฟเวอร์ (mode 600); เปลี่ยนเฉพาะ `FRONTEND_IMAGE` แบบ atomic
6. ใช้ compose เดิมที่ `/opt/bcai-account/deploy` สั่ง `up -d --no-deps frontend`; script `/opt/bcai-account/releases/r20260911-menu-top-1/deploy.sh` มี rollback เมื่อ start/health/HTTP checks ไม่ผ่าน
7. ตรวจผ่าน HTTPS จริง แล้วลบ preview container และปิด SSH tunnel ที่สร้างสำหรับงานนี้

ไม่ได้แก้ backend env, bootstrap, database schema, DNS, Caddy หรือข้อมูลบัญชี และไม่ได้อัปเดต image/backend language TSV ของ mainapi; ชื่อเมนูไทยใหม่ใช้ catalog ใน frontend และข้อความที่ไม่มี dictionary ใช้ fallback ตามโค้ด

## ผลตรวจ

- `npm run verify:all`: code map, backend, outbox, projection ผ่าน; frontend/frontend-build ของ wrapper เรียก npm ไม่ได้เพราะ Git Bash บน Windows จึง **ไม่อ้างว่า wrapper ทั้งคำสั่งผ่าน**
- รันคำสั่ง frontend เดิมผ่าน PowerShell แทน: lint 0 errors / 219 warnings, TypeScript ผ่าน, Vitest **49 files / 357 tests ผ่าน**
- Production Docker build ผ่าน Next.js build และ TypeScript; image นี้ทดสอบ preview ก่อนใช้งานจริง
- Preview: `menu-consistency.spec.ts` ผ่าน รวม Light/Dark × 1600/1280/1024/768 portrait
- Production HTTPS: `menu-champ-upgrade`, `menu-consistency`, `menu-tree-master` ผ่านครบ 3 tests; ตรวจเมนู/ค้นชื่อ Champ/pending/ค่าเริ่มต้น top/สลับ left แล้ว reload/เปิดหน้าที่เชื่อมแล้ว
- รอบแรกของ tree test มีข้อความ CSP แบบ report-only จาก iframe Google ระหว่าง login; ปรับ listener ให้เริ่มตรวจเมื่อเข้าเมนูเหมือนอีกสองเทสต์ แล้วรัน tree test ผ่าน ไม่ได้ลด CSP หรือกรอง error ของแอปทิ้ง
- ไม่ทำ CRUD กับข้อมูลจริงระหว่าง smoke test; ใช้บัญชี Demo เพื่ออ่านและเปิดเมนู

## ตัวอย่างตรวจและ rollback

บนเซิร์ฟเวอร์:

```sh
docker inspect bcai-account-frontend-1 --format '{{.Config.Image}} {{.State.Health.Status}}'
curl --fail --silent --output /dev/null --write-out '%{http_code}\n' https://account.bcaicloud.com/
```

หากต้องย้อน release นี้ ให้ยืนยันว่าไม่มี release ใหม่กว่าก่อน แล้วรัน:

```sh
cp -p /opt/bcai-account/releases/r20260911-menu-top-1/release.env.before /etc/bcai-account/release.env
cd /opt/bcai-account/deploy
docker compose -p bcai-account --env-file /etc/bcai-account/release.env -f compose.yml -f compose.8gb.yml up -d --no-deps frontend
```

## ข้อจำกัด

เป็นการ deploy frontend ไม่ใช่การรับรองว่าระบบบัญชีหรือรายงานทุกแบบพัฒนาเสร็จแล้ว; เมนูที่ยังไม่พร้อมยังขึ้นรอพัฒนา รายการ warning ของ lint เดิมยังอยู่ และ backend tests มี quarantine ตาม `backend/.ci/test-quarantine.txt` ซึ่ง compile แต่ไม่รัน business tests ของ package เหล่านั้น
