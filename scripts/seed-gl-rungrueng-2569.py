"""Create 144 synthetic draft journals through the normal GL API.

Dry run: py scripts/seed-gl-rungrueng-2569.py
Apply:   py scripts/seed-gl-rungrueng-2569.py --apply
Fixed scope: rungrueng / 01 / 00000 / 2569. No posting or deletion.
All money uses integer satang; VAT rounds half-up once per invoice.
"""
import argparse
import calendar
import json
import random
import urllib.error
import urllib.request
import uuid
from collections import Counter
from decimal import Decimal
from pathlib import Path

BASE = "https://account.bcaicloud.com"
BATCH = "gl-rungrueng-01-2569-20260920-v1"
OUT = Path(__file__).resolve().parents[1] / "docs" / "examples" / (BATCH + ".json")
HEADERS = {"Content-Type": "application/json", "Origin": BASE,
           "X-BC-Backend-URL": BASE + "/backend/goapi"}


def call(path, payload=None):
    request = urllib.request.Request(BASE + path, headers=HEADERS,
        data=None if payload is None else json.dumps(payload).encode())
    try:
        with urllib.request.urlopen(request, timeout=40) as response:
            body = json.load(response)
    except urllib.error.HTTPError as error:
        raise RuntimeError(f"{path}: HTTP {error.code}") from None
    if body.get("success") is False:
        raise RuntimeError(f"{path}: {body.get('message', 'request rejected')}")
    return body.get("data", body)


def listing(resource):
    result, page = [], 1
    while True:
        data = call(f"/api/gl/{resource}?limit=1000&page={page}")
        result.extend(data["items"])
        if len(result) >= data["total"]:
            return result
        if not data["items"]:
            raise RuntimeError("Incomplete pagination")
        page += 1


def money(satang):
    assert isinstance(satang, int) and satang >= 0
    return f"{satang // 100}.{satang % 100:02d}"


def pair(debit, credit, amount):
    return [(debit, amount, 0), (credit, 0, amount)]


def invoice(net, purchase=False, bank=False):
    vat = (net * 7 + 50) // 100
    gross = net + vat
    if purchase:
        return [("1141", net, 0), ("1151", vat, 0), ("1121" if bank else "2111", 0, gross)]
    return [("1121" if bank else "1131", gross, 0), ("4111", 0, net), ("2131", 0, vat)]


def generate(accounts):
    rng = random.Random(20260920)
    journals = []
    materials = ["ปูนซีเมนต์และอิฐมวลเบา", "เหล็กเส้นและลวดผูกเหล็ก", "กระเบื้องและวัสดุมุงหลังคา", "อุปกรณ์ประปาและสุขภัณฑ์"]
    for month in range(1, 13):
        dates = sorted(rng.sample(range(1, calendar.monthrange(2026, month)[1] + 1), 12))
        purchase, sale = rng.randint(6000000, 10000000), rng.randint(9000000, 16000000)
        purchase_gross = purchase + (purchase * 7 + 50) // 100
        sale_gross = sale + (sale * 7 + 50) // 100
        material = rng.choice(materials)
        specs = [
            ("RV", "รับเงินกู้ยืมระยะสั้นจากกรรมการเข้าบัญชีธนาคาร", pair("1121", "2173", rng.randint(8000000, 14000000)), "LN"),
            ("SV", f"ซื้อ{material}เข้าคลังสินค้า เครดิต 30 วัน", invoice(purchase, purchase=True), "PI"),
            ("UV", f"ขาย{material}ให้ผู้รับเหมา เครดิต 30 วัน", invoice(sale), "INV"),
            ("RV", "รับชำระลูกหนี้การค้าบางส่วนผ่านธนาคาร", pair("1121", "1131", sale_gross * rng.randint(55, 85) // 100), "INV"),
            ("PV", "ชำระเจ้าหนี้ค่าวัสดุก่อสร้างบางส่วนผ่านธนาคาร", pair("2111", "1121", purchase_gross * rng.randint(50, 80) // 100), "PI"),
            ("UV", "ขายวัสดุก่อสร้าง รับชำระโดยโอนเข้าธนาคาร", invoice(rng.randint(3000000, 7000000), bank=True), "INV"),
            ("SV", "ซื้อวัสดุก่อสร้างเติมคลัง ชำระโดยโอนธนาคาร", invoice(rng.randint(700000, 1500000), purchase=True, bank=True), "PI"),
            ("PV", "ชำระค่าเครื่องเขียนและวัสดุสำนักงานตามใบเสร็จ", pair("5331", "1121", rng.randint(150000, 450000)), "RC"),
            ("PV", "ชำระค่าบริการระบบคลาวด์ประจำเดือนตามใบเสร็จ", pair("5325", "1121", rng.randint(120000, 300000)), "RC"),
            ("PV", "บันทึกค่าธรรมเนียมธนาคารประจำเดือน", pair("5421", "1121", rng.randint(10000, 45000)), "BNK"),
            ("JV", "ถอนเงินจากธนาคารสำรองเป็นเงินสดในมือ", pair("1111", "1121", rng.randint(500000, 1000000)), "TR"),
            ("PV", "คืนเงินกู้ยืมระยะสั้นแก่กรรมการบางส่วน", pair("2173", "1121", rng.randint(1500000, 3000000)), "LN"),
        ]
        for index, (book, description, lines, reference) in enumerate(specs, 1):
            ref_number = {4: 3, 5: 2, 12: 1}.get(index, index)
            doc = {"docno": f"{book}69{month:02d}-{index:04d}", "date": f"2026-{month:02d}-{dates[index-1]:02d}",
                   "bookcode": book, "fiscalyear": "2569", "description": description,
                   "reference": f"{reference}69{month:02d}-{ref_number:04d}", "branchcode": "00000", "kind": "manual", "lines": []}
            for code, debit, credit in lines:
                account = accounts[code]
                assert account["isactive"] and account["allowposting"] and not account.get("isdeleted")
                name = next(n["name"] for n in account["names"] if n["code"] == "th")
                doc["lines"].append({"accountcode": code, "accountname": name, "description": description,
                                     "debit": money(debit), "credit": money(credit), "departmentcode": "", "projectcode": "", "cashflow": ""})
            assert sum(Decimal(x["debit"]) for x in doc["lines"]) == sum(Decimal(x["credit"]) for x in doc["lines"])
            journals.append(doc)
    assert len(journals) == len({j["docno"] for j in journals}) == 144
    return journals


def same_document(actual, expected):
    for key, value in expected.items():
        if key != "lines" and actual.get(key) != value:
            return False
    if len(actual.get("lines", [])) != len(expected["lines"]):
        return False
    for actual_line, expected_line in zip(actual["lines"], expected["lines"]):
        for key in ("accountcode", "description", "debit", "credit"):
            if key in ("debit", "credit"):
                if Decimal(actual_line[key]) != Decimal(expected_line[key]):
                    return False
            elif actual_line.get(key) != expected_line[key]:
                return False
    return actual.get("status") == "draft" and not actual.get("isdeleted")


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--apply", action="store_true")
    args = parser.parse_args()
    login = call("/api/auth/demo-login", {})
    HEADERS["Authorization"] = "Bearer " + login["token"]
    call("/api/workspace/select-holding", {"holdingcode": "rungrueng", "businesscode": "01", "branchuid": "00000"})
    years = listing("fiscal-years")
    year = next(y for y in years if y["code"] == "2569")
    assert year["isactive"] and not year["closed"] and year["scale"] == 2
    assert year["startdate"] == "2026-01-01" and year["enddate"] == "2026-12-31"
    accounts = {a["accountcode"]: a for a in listing("accounts")}
    journals = generate(accounts)
    for period in listing("periods"):
        if period.get("locked") and any(period["startdate"] <= j["date"] <= period["enddate"] for j in journals):
            raise RuntimeError("A target period is locked")
    existing = {j["docno"]: j for j in listing("journals")}
    for journal in journals:
        if journal["docno"] in existing and not same_document(existing[journal["docno"]], journal):
            raise RuntimeError("Existing document collision: " + journal["docno"])
    manifest = {"batch": BATCH, "synthetic": True, "seed": 20260920, "holding": "rungrueng", "company": "01",
                "fiscalyear": "2569", "status": "draft", "currency": "THB", "rounding": "integer satang; VAT 7/100 half-up once per invoice",
                "vatSource": "https://rd.go.th/59.html", "journals": journals}
    if OUT.exists():
        prior = json.loads(OUT.read_text(encoding="utf-8"))
        assert prior["batch"] == BATCH
        for field in ("supersededDrafts", "persistedIds", "verifiedMonths"):
            if field in prior:
                manifest[field] = prior[field]
    OUT.write_text(json.dumps(manifest, ensure_ascii=False, indent=2), encoding="utf-8")
    print(json.dumps({"planned": len(journals), "existingIdentical": sum(j["docno"] in existing for j in journals), "apply": args.apply}))
    if not args.apply:
        return
    for index, journal in enumerate(journals, 1):
        if journal["docno"] not in existing:
            command = {"resource": "journals", "action": "create", "requestid": str(uuid.uuid5(uuid.NAMESPACE_URL, BATCH + "/" + journal["docno"])), "journal": journal}
            result = call("/api/gl/command", command)
            saved = call("/api/gl/journals/" + result["id"])
            assert same_document(saved, journal), "Persisted document mismatch"
        if index % 12 == 0:
            print(f"Verified month {index // 12:02d}: 12 journals", flush=True)
    persisted = {j["docno"]: j for j in listing("journals")}
    for j in journals:
        assert same_document(persisted[j["docno"]], j)
    monthly = Counter(j["date"][:7] for j in journals)
    assert len(monthly) == 12 and all(n == 12 for n in monthly.values())
    ids = {j["docno"]: persisted[j["docno"]]["id"] for j in journals}
    manifest["persistedIds"] = ids
    manifest["verifiedMonths"] = dict(sorted(monthly.items()))
    OUT.write_text(json.dumps(manifest, ensure_ascii=False, indent=2), encoding="utf-8")
    print(json.dumps({"verified": 144, "months": dict(sorted(monthly.items())), "draft": True}))


if __name__ == "__main__":
    main()
