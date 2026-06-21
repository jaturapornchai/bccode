# BC Ai Account — On-Premise Server Specification

> รายงานสเปกเซิร์ฟเวอร์พร้อมข้อมูลการเข้าถึง (credentials) — สำหรับการดูแลระบบภายในองค์กร
>
> **วันที่ติดตั้ง**: 2026-06-19
> **ผู้ติดตั้ง**: Jead (ลุงจืด)
> **หมายเหตุ**: เอกสารนี้มีข้อมูลละเอียดอ่อน (credentials) เก็บในที่ปลอดภัยเท่านั้น

---

## 1. ข้อมูลเซิร์ฟเวอร์ (Server Info)

| รายการ | ค่า |
|---|---|
| **ชื่อเครื่อง** | smlsoft-System-Product-Name |
| **IP Address (LAN)** | `192.168.2.202` |
| **IP Address (Docker)** | `172.17.0.1`, `172.18.0.1` |
| **ระบบปฏิบัติการ** | Ubuntu 26.04 LTS (Resolute Raccoon) |
| **Kernel** | 7.0.0-22-generic (x86_64) |
| **CPU** | Intel Core i5-12400 (12th Gen) — 6 cores / 12 threads |
| **RAM** | 30 GB (ใช้ ~4.4 GB, ว่าง ~26 GB) |
| **Disk** | 457 GB NVMe (ใช้ 28 GB, ว่าง 406 GB) — แบ่งพาร์ทิชั่นเดียว |
| **Uptime ขณะรายงาน** | 5 ชม. 50 นาที |

---

## 2. ข้อมูลเข้าถึงเซิร์ฟเวอร์ (Access)

### SSH
| รายการ | ค่า |
|---|---|
| **Host** | `192.168.2.202` |
| **Port** | `22` |
| **Username** | `smlsoft` |
| **Password** | `smlsoft` |
| **Sudo** | `echo smlsoft \| sudo -S <command>` |

```bash
ssh smlsoft@192.168.2.202
```

---

## 3. บริการที่ติดตั้ง (Services)

### 3.1 PostgreSQL 18.4 (Native — systemd)

| รายการ | ค่า |
|---|---|
| **รุ่น** | PostgreSQL 18.4 (Ubuntu 18.4-0ubuntu0.26.04.1) |
| **วิธีติดตั้ง** | `apt install postgresql-18` (Ubuntu default repo) |
| **การรัน** | Native systemd service |
| **Config** | `/etc/postgresql/18/main/postgresql.conf` |
| **HBA** | `/etc/postgresql/18/main/pg_hba.conf` |
| **Listen** | `0.0.0.0:5432` |
| **Service คำสั่ง** | `systemctl restart postgresql` |

**Connection:**
```
host: 192.168.2.202
port: 5432
database: postgres (admin schema), appdb (business)
user: smlsoft
password: smlsoft
```

**Database ที่สร้างไว้:**
- `postgres` — admin database (มี business tables: organizationcompanies, organizationbranches, journals, queues ฯลฯ 9 tables)
- `appdb` — business database

**Tables ใน postgres db:**
- `deadletterqueue`, `distributedlocks`, `queues` — queue infra
- `journals`, `journalsdetail`, `journalvatsdetails`, `journaltaxesdetails` — GL/VAT
- `organizationcompanies`, `organizationbranches` — org master

### 3.2 ClickHouse 26.6.1 (Native — daemon)

| รายการ | ค่า |
|---|---|
| **รุ่น** | ClickHouse 26.6.1.1015 (official build) |
| **วิธีติดตั้ง** | Native binary installer |
| **การรัน** | Daemon (ผ่าน `clickhouse start/stop`) |
| **Config** | `/etc/clickhouse-server/config.xml` |
| **Users override** | `/etc/clickhouse-server/users.d/default-password.xml` |
| **Listen** | `0.0.0.0:9000` (native), `0.0.0.0:8123` (HTTP) |
| **Data dir** | `/var/lib/clickhouse/` |
| **Service คำสั่ง** | `clickhouse start` / `clickhouse stop` |

**Connection:**
```
host: 192.168.2.202
native port: 9000
HTTP port: 8123
user: default
password: smlsoft
database: appdb
```

**Test:**
```bash
curl -u "default:smlsoft" "http://192.168.2.202:8123/?query=SELECT+version()"
# → 26.6.1.1015
```

### 3.3 MongoDB 7.0 (Docker container)

| รายการ | ค่า |
|---|---|
| **รุ่น** | MongoDB 7.0 (official Docker image `mongo:7.0`) |
| **เหตุผลที่ใช้ 7.0** | Mongo 8.0 มี known incompatibility กับ Linux kernel 6.19+ (SERVER-121912) → kernel 7.0 บังคับใช้ 7.0 |
| **ชื่อ container** | `mongodb` |
| **Listen** | `0.0.0.0:27017` |
| **Data volume** | `/data/mongodb` (host) → `/data/db` (container) |
| **Auto-restart** | `unless-stopped` |
| **Service คำสั่ง** | `docker restart mongodb` |

**Connection:**
```
URI: mongodb://smlsoft:smlsoft@192.168.2.202:27017/appdb?authSource=admin
user: smlsoft
password: smlsoft
auth database: admin
app database: appdb
```

**Data ในระบบ (seeded จาก MongoDB Atlas):**
- `users` — 13 records (owner + 12 seed users)
- `organizationcompanies` — 8 companies
- `organizationbranches` — 38 branches
- `shops` — 7 shops
- 18 collections รวม (permissiondefinitions, permissiongroups, units, warehouse, ฯลฯ)
- **รวม 4,604 documents**

### 3.4 MinIO 2025-09-07 (Docker container — S3-compatible image store)

| รายการ | ค่า |
|---|---|
| **รุ่น** | MinIO RELEASE.2025-09-07T16-13-09Z |
| **ชื่อ container** | `minio` |
| **S3 API** | `0.0.0.0:9100` |
| **Web Console** | `0.0.0.0:9101` |
| **Data volume** | `/data/minio` (host) → `/data` (container) |
| **Auto-restart** | `unless-stopped` |
| **Service คำสั่ง** | `docker restart minio` |

**Credentials:**
| บัญชี | ใช้สำหรับ |
|---|---|
| **Root user**: `smlsoft` / `smlsoft123` | Web Console + admin |
| **Service account (backend)**: `bc-backend` / `bc-backend-secret-2026` | ใช้ใน backend (R2_ACCESS_KEY_ID) |

**Bucket:**
- `app-images` — เก็บรูปภาพทั้งหมด (logos, products, attachments)

**Web Console:**
```
URL: http://192.168.2.202:9101
User: smlsoft
Password: smlsoft123
```

### 3.5 Backend MainAPI (Docker compose)

| รายการ | ค่า |
|---|---|
| **ชื่อ container** | `mainapi` |
| **Listen** | `0.0.0.0:8888` |
| **Health check** | `GET /healthz` → 200 "ok" |
| **Source code** | `/home/smlsoft/bc-backend/` (host) |
| **Build** | `docker compose up -d --build` (ใน `bc-backend/`) |
| **Logs** | `docker logs mainapi -f` |

**Config files:**
- `/home/smlsoft/bc-backend/bootstrap.json` — DB connections, integrations
- `/home/smlsoft/bc-backend/.env` — R2 (MinIO) credentials
- `/home/smlsoft/bc-backend/docker-compose.yml` — service definitions

### 3.6 Redis + Kafka (Docker compose)

| Service | Container | Port | Notes |
|---|---|---|---|
| **Redis** | `redis` | `127.0.0.1:6379` | Cache + session |
| **Kafka** | `kafka` | `127.0.0.1:9092` | Event bus (KRaft mode, no Zookeeper) |

---

## 4. สรุป Port Mapping

| Port | Service | Protocol | Bind |
|---|---|---|---|
| `22` | SSH | TCP | 0.0.0.0 |
| `5432` | PostgreSQL | TCP | 0.0.0.0 |
| `8123` | ClickHouse HTTP | HTTP | 0.0.0.0 |
| `8888` | Backend MainAPI | HTTP | 0.0.0.0 |
| `9000` | ClickHouse native | TCP | 0.0.0.0 |
| `9100` | MinIO S3 API | HTTP | 0.0.0.0 |
| `9101` | MinIO Console | HTTP | 0.0.0.0 |
| `27017` | MongoDB | TCP | 0.0.0.0 |
| `6379` | Redis | TCP | 127.0.0.1 |
| `9092` | Kafka | TCP | 127.0.0.1 |

---

## 5. การเข้าใช้งานระบบ (Application Access)

### Frontend (Next.js)
```
URL: http://localhost:3000  (รันบนเครื่อง dev)
URL ของ backend: http://192.168.2.202:8888  (ผ่าน .env.local ของ frontend)
```

### บัญชีเข้าระบบ (Default Owner)
```
URL: http://localhost:3000
Username: jaturapornchai@gmail.com
Password: smlsoft
Role: Owner (เจ้าของร้าน)
UID: 3EjpEvCK44WNvRi7Z3zA7VOUh7z
```

### API Direct Access
```bash
# Login
curl -X POST http://192.168.2.202:8888/login \
  -H "Content-Type: application/json" \
  -d '{"username":"jaturapornchai@gmail.com","password":"smlsoft"}'

# Profile (with token)
curl -H "Authorization: Bearer <token>" http://192.168.2.202:8888/profile

# Holdings
curl -H "Authorization: Bearer <token>" http://192.168.2.202:8888/list-holding

# Health
curl http://192.168.2.202:8888/healthz
```

---

## 6. การดูแลระบบ (Operations)

### 6.1 คำสั่งดูสถานะ
```bash
ssh smlsoft@192.168.2.202

# Docker containers
docker ps

# Service status
systemctl status postgresql
systemctl status clickhouse-server  # แสดง failed ปกติ เพราะ CH รันเป็น daemon ไม่ใช่ systemd
pgrep -fa clickhouse-server          # ใช้อันนี้ตรวจ CH แทน

# Logs
docker logs mainapi -f --tail 50
docker logs mongodb --tail 20
docker logs minio --tail 20
```

### 6.2 Restart services
```bash
# Backend
cd /home/smlsoft/bc-backend && docker compose restart mainapi

# PostgreSQL
sudo systemctl restart postgresql

# ClickHouse
sudo clickhouse stop && sudo clickhouse start

# MongoDB / MinIO
docker restart mongodb
docker restart minio

# Redis / Kafka (ระวัง: Kafka restart กระทบ event flow)
docker compose restart redis kafka
```

### 6.3 Deploy backend ใหม่ (หลังแก้โค้ด)
```bash
# จากเครื่อง dev
cd D:\bccode\backend
tar --exclude='.git' --exclude='vendor' --exclude='node_modules' --exclude='dist' \
  -czf - . | ssh smlsoft@192.168.2.202 "cat > /tmp/bc-backend.tar.gz"

ssh smlsoft@192.168.2.202
cd /home/smlsoft/bc-backend
tar xzf /tmp/bc-backend.tar.gz
docker compose up -d --build mainapi
curl http://localhost:8888/healthz
```

---

## 7. การสำรองข้อมูล (Backup)

### PostgreSQL
```bash
ssh smlsoft@192.168.2.202
sudo -u postgres pg_dumpall > /tmp/pg-backup-$(date +%F).sql
# หรือ database เดียว
sudo -u postgres pg_dump appdb > /tmp/appdb-backup-$(date +%F).sql
```

### MongoDB
```bash
docker exec mongodb mongodump "mongodb://smlsoft:smlsoft@localhost:27017/appdb?authSource=admin" \
  --out /tmp/mongo-backup-$(date +%F)
```

### ClickHouse
```bash
clickhouse-backup create dev-backup-$(date +%F)
```

### MinIO (S3)
- Sync bucket ผ่าน `mc mirror` ไป remote S3 หรือ disk อื่น

---

## 8. ความปลอดภัย (Security Notes)

⚠️ **สำคัญ**: ข้อมูลนี้เป็น default dev credentials สำหรับวง LAN เท่านั้น

### สิ่งที่ต้องทำก่อน production:
1. **เปลี่ยน password ทุกตัว** (SSH, PG, CH, Mongo, MinIO, owner user)
2. **ปิด port ที่ไม่จำเป็น** จาก 0.0.0.0 → 127.0.0.1 หรือ internal LAN เท่านั้น
3. **เปิด firewall** (ufw) อนุญาตเฉพาะ SSH + 8888 + 3000
4. **Setup TLS/HTTPS** ผ่าน reverse proxy (Caddy/Nginx)
5. **Setup automatic backup** (cron + offsite)
6. **Setup monitoring** (logs, alerts, disk space)
7. **Hardening SSH** (key-based auth, disable password)
8. **MongoDB auth** ปิด anonymous access
9. **MinIO** เปลี่ยน root password จาก default

### Credentials Summary (เก็บลับ)
| Service | User | Password |
|---|---|---|
| SSH | smlsoft | smlsoft |
| PostgreSQL | smlsoft | smlsoft |
| ClickHouse | default | smlsoft |
| MongoDB | smlsoft | smlsoft |
| MinIO Root | smlsoft | smlsoft123 |
| MinIO Backend | bc-backend | bc-backend-secret-2026 |
| App Owner | jaturapornchai@gmail.com | smlsoft |

---

## 9. ปัญหาที่ทราบ (Known Issues)

| ปัญหา | สถานะ | วิธีแก้ |
|---|---|---|
| Mongo 8.0 ใช้กับ kernel 7.0 ไม่ได้ | ✅ แก้แล้ว — ใช้ Mongo 7.0 | รอ Mongo release 8.1+ ที่ fix SERVER-121912 |
| ClickHouse systemd แสดง failed | ⚠️ ปกติ — CH รันเป็น daemon | ใช้ `pgrep clickhouse-server` ตรวจแทน |
| Kafka connection refused ตอน boot | ✅ recover เองได้ | race condition ปกติของ docker compose |
| Frontend next dev compile ช้าครั้งแรก | ⚠️ normal | รอ 2-3 นาที, build cache ช่วยครั้งต่อไป |

---

## 10. แหล่งข้อมูลเพิ่มเติม

- **Project repo**: `D:\bccode` (local dev machine)
- **Backend deploy path**: `/home/smlsoft/bc-backend/`
- **Docker images ที่ใช้**:
  - `bc-backend-mainapi:latest` (build เอง)
  - `mongo:7.0`
  - `minio/minio:latest`
  - `redis:7-alpine`
  - `confluentinc/confluent-local:latest`
- **Native services**: PostgreSQL 18, ClickHouse 26.6
- **OS**: Ubuntu 26.04 LTS
