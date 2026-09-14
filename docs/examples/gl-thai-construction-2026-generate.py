"""Reproducible fictional GL scenario; all accounting arithmetic uses Decimal."""
import calendar
import json
from collections import defaultdict
from decimal import Decimal, ROUND_HALF_UP, getcontext
from pathlib import Path

getcontext().prec = 50
D = Decimal
ZERO = D("0")
CENT = D("0.01")
VAT = D("0.07")
ROOT = Path(__file__).resolve().parent
PREFIX = "DEMO-BM69-"
FY = "DEMO2569"
BRANCH = "00000"


def money(value):
    value = D(value)
    rounded = value.quantize(CENT, rounding=ROUND_HALF_UP)
    assert value == rounded, f"unexpected rounding required: {value}"
    return format(rounded, ".2f")


def code(value):
    return "BM69-" + value


groups = [
    {"code": "BM69-" + key, "name": "ข้อมูลตัวอย่าง - " + name, "isactive": True, "amount": "0.00"}
    for key, name in [("A", "สินทรัพย์"), ("L", "หนี้สิน"), ("E", "ส่วนทุน"), ("R", "รายได้"), ("X", "ค่าใช้จ่าย")]
]
type_info = {
    "asset": ("A", "1000", "debit"), "liability": ("L", "2000", "credit"),
    "equity": ("E", "3000", "credit"), "income": ("R", "4000", "credit"),
    "expense": ("X", "5000", "debit"),
}
specs = [
    ("1000", "สินทรัพย์", "asset", False, False, "debit"),
    ("2000", "หนี้สิน", "liability", False, False, "credit"),
    ("3000", "ส่วนทุน", "equity", False, False, "credit"),
    ("4000", "รายได้", "income", False, False, "credit"),
    ("5000", "ค่าใช้จ่าย", "expense", False, False, "debit"),
]
for account, name, kind, cash, normal in [
    ("110101", "เงินสดหน้าร้าน", "asset", True, "debit"),
    ("110102", "เงินฝากธนาคาร", "asset", True, "debit"),
    ("110103", "เงินสดย่อย", "asset", True, "debit"),
    ("110201", "ลูกหนี้การค้า", "asset", False, "debit"),
    ("110301", "สินค้าคงเหลือ - ปูนซีเมนต์", "asset", False, "debit"),
    ("110302", "สินค้าคงเหลือ - เหล็กเส้น", "asset", False, "debit"),
    ("110303", "สินค้าคงเหลือ - สีทาอาคาร", "asset", False, "debit"),
    ("110401", "ภาษีซื้อ", "asset", False, "debit"),
    ("110501", "เบี้ยประกันภัยจ่ายล่วงหน้า", "asset", False, "debit"),
    ("120101", "อุปกรณ์ร้านและคลังสินค้า", "asset", False, "debit"),
    ("120102", "รถขนส่งสินค้า", "asset", False, "debit"),
    ("129101", "ค่าเสื่อมราคาสะสม - อุปกรณ์", "asset", False, "credit"),
    ("129102", "ค่าเสื่อมราคาสะสม - รถขนส่ง", "asset", False, "credit"),
    ("210101", "เจ้าหนี้การค้า", "liability", False, "credit"),
    ("210201", "ภาษีขาย", "liability", False, "credit"),
    ("210301", "ค่าใช้จ่ายค้างจ่าย", "liability", False, "credit"),
    ("220101", "เงินกู้ระยะยาว", "liability", False, "credit"),
    ("310101", "ทุนเจ้าของ", "equity", False, "credit"),
    ("310201", "กำไรสะสม", "equity", False, "credit"),
    ("310301", "กำไรขาดทุนรอจัดสรร", "equity", False, "credit"),
    ("410101", "รายได้ขายปูนซีเมนต์", "income", False, "credit"),
    ("410102", "รายได้ขายเหล็กเส้น", "income", False, "credit"),
    ("410103", "รายได้ขายสีทาอาคาร", "income", False, "credit"),
    ("419101", "รับคืนสินค้า", "income", False, "debit"),
    ("510101", "ต้นทุนขายปูนซีเมนต์", "expense", False, "debit"),
    ("510102", "ต้นทุนขายเหล็กเส้น", "expense", False, "debit"),
    ("510103", "ต้นทุนขายสีทาอาคาร", "expense", False, "debit"),
    ("520101", "ค่าไฟฟ้า", "expense", False, "debit"),
    ("520102", "ค่าน้ำมันรถขนส่ง", "expense", False, "debit"),
    ("520103", "ค่าเสื่อมราคา", "expense", False, "debit"),
    ("520104", "ค่าเบี้ยประกันภัย", "expense", False, "debit"),
    ("520105", "ดอกเบี้ยจ่าย", "expense", False, "debit"),
    ("520106", "สินค้าชำรุดเสียหาย", "expense", False, "debit"),
]:
    specs.append((account, name, kind, True, cash, normal))
accounts = []
for account, name, kind, posting, cash, normal in specs:
    group, parent, _ = type_info[kind]
    accounts.append({"accountcode": code(account), "names": [{"code": "th", "name": "ข้อมูลตัวอย่าง - " + name}],
                     "accounttype": kind, "parentaccountcode": code(parent) if posting else "", "normalbalance": normal,
                     "allowposting": posting, "isactive": True, "accountgroup": "BM69-" + group, "iscash": cash})
account_map = {a["accountcode"]: a for a in accounts}
assert len(accounts) == 38
year = {"code": FY, "startdate": "2026-01-01", "enddate": "2026-12-31", "currency": "THB", "scale": 2,
        "isactive": True, "closed": False, "profitlossaccount": code("310301"), "retainedearningsaccount": code("310201")}
periods = [{"code": f"BM69-P{month:02d}", "name": f"ข้อมูลตัวอย่าง - งวด {month:02d}/2569", "fiscalyear": FY,
            "startdate": f"2026-{month:02d}-01", "enddate": f"2026-{month:02d}-{calendar.monthrange(2026, month)[1]:02d}",
            "isactive": True, "locked": False, "amount": "0.00"} for month in range(1, 13)]

items = {
    "CEMENT": {"name": "ปูนซีเมนต์สมมติ 50 กก.", "unit": "ถุง", "cost": D("100"), "price": D("150"), "opening": D("1000"), "inventory": "110301", "sales": "410101", "cogs": "510101"},
    "STEEL": {"name": "เหล็กเส้นสมมติ", "unit": "เส้น", "cost": D("500"), "price": D("650"), "opening": D("200"), "inventory": "110302", "sales": "410102", "cogs": "510102"},
    "PAINT": {"name": "สีทาอาคารสมมติ", "unit": "ถัง", "cost": D("400"), "price": D("600"), "opening": D("100"), "inventory": "110303", "sales": "410103", "cogs": "510103"},
}
stock_quantity = {key: item["opening"] for key, item in items.items()}
stock_moves = []
journals = []


def line(account, side, amount, department="BM69-ADMIN", project="", cashflow=""):
    amount = D(amount)
    assert amount > ZERO
    return {"accountcode": code(account), "description": "ข้อมูลตัวอย่าง", "debit": money(amount if side == "debit" else ZERO),
            "credit": money(amount if side == "credit" else ZERO), "departmentcode": department, "projectcode": project, "cashflow": cashflow}


def dr(account, amount, **kw):
    return line(account, "debit", amount, **kw)


def cr(account, amount, **kw):
    return line(account, "credit", amount, **kw)


def add(doc, date, book, description, lines, *, opening=False, post=True, facts=None):
    debit = sum((D(row["debit"]) for row in lines), ZERO)
    credit = sum((D(row["credit"]) for row in lines), ZERO)
    assert debit == credit, (doc, debit, credit)
    journals.append({"post": post, "journal": {"docno": PREFIX + doc, "date": date, "bookcode": book, "fiscalyear": FY,
                     "description": "ข้อมูลตัวอย่าง - " + description, "reference": PREFIX + "SCENARIO", "branchcode": BRANCH,
                     "currency": "THB", "kind": "opening" if opening else "manual", "lines": lines},
                     "businessfacts": facts or {}, "expecteddebit": money(debit), "expectedcredit": money(credit)})


def stock_move(doc, sku, quantity):
    quantity = D(quantity)
    stock_quantity[sku] += quantity
    assert stock_quantity[sku] >= ZERO
    stock_moves.append({"docno": PREFIX + doc, "sku": "BM69-" + sku, "quantitychange": money(quantity),
                        "unitcost": money(items[sku]["cost"]), "valuechange": money(quantity * items[sku]["cost"])})


opening_debits = [("110101", "20000"), ("110102", "300000"), ("110103", "5000"), ("110201", "50000"),
                  ("110301", "100000"), ("110302", "100000"), ("110303", "40000"), ("110501", "12000"),
                  ("120101", "60000"), ("120102", "400000")]
opening_credits = [("129101", "12000"), ("129102", "80000"), ("210101", "85000"), ("210301", "10000"),
                   ("220101", "300000"), ("310101", "500000"), ("310201", "100000")]
add("OPEN-001", "2026-01-01", "JV", "ยอดยกมาร้านวัสดุก่อสร้างสมมติ", [dr(a, v) for a, v in opening_debits] + [cr(a, v) for a, v in opening_credits], opening=True)
# BC book labels use SV for purchases and UV for sales; retain that source contract.
for doc, date, sku, qty, paid in [("SV-001", "2026-09-01", "CEMENT", "500", False), ("SV-002", "2026-09-01", "STEEL", "100", True), ("SV-003", "2026-09-02", "PAINT", "50", False)]:
    item, quantity = items[sku], D(qty)
    net = quantity * item["cost"]
    tax, total = net * VAT, net * (D("1") + VAT)
    add(doc, date, "SV", f"ซื้อ{item['name']} {qty} {item['unit']} " + ("จ่ายผ่านธนาคาร" if paid else "เป็นเงินเชื่อ"),
        [dr(item["inventory"], net), dr("110401", tax), cr("110102" if paid else "210101", total, cashflow="operating" if paid else "")],
        facts={"net": money(net), "vat": money(tax), "gross": money(total), "vatfraction": "0.07", "fictionalsupplier": "ผู้จำหน่ายตัวอย่าง " + sku})
    stock_move(doc, sku, quantity)

for doc, date, quantities, cash, project in [
    ("UV-001", "2026-09-03", [("CEMENT", "600"), ("STEEL", "80")], False, "BM69-SITE-A"),
    ("RV-001", "2026-09-03", [("PAINT", "40")], True, "BM69-RETAIL"),
    ("UV-002", "2026-09-04", [("STEEL", "70"), ("PAINT", "30")], False, "BM69-SITE-B"),
]:
    net = sum((D(qty) * items[sku]["price"] for sku, qty in quantities), ZERO)
    tax, total = net * VAT, net * (D("1") + VAT)
    dimension = {"department": "BM69-SALES", "project": project}
    lines = [dr("110101" if cash else "110201", total, cashflow="operating" if cash else "", **dimension), cr("210201", tax, **dimension)]
    for sku, qty in quantities:
        item, quantity = items[sku], D(qty)
        lines += [cr(item["sales"], quantity * item["price"], **dimension), dr(item["cogs"], quantity * item["cost"], **dimension), cr(item["inventory"], quantity * item["cost"], **dimension)]
        stock_move(doc, sku, -quantity)
    add(doc, date, "RV" if cash else "UV", "ขายวัสดุก่อสร้าง " + ("รับเงินสดหน้าร้าน" if cash else "ให้ลูกค้าโครงการสมมติ " + project), lines,
        facts={"net": money(net), "vat": money(tax), "gross": money(total), "vatfraction": "0.07", "fictionalcustomer": project})

add("RV-002", "2026-09-05", "RV", "เก็บลูกหนี้ยกมา 30,000 และลูกหนี้ UV-001 อีก 100,000", [dr("110102", "130000", cashflow="operating"), cr("110201", "130000")])
add("PV-001", "2026-09-05", "PV", "จ่ายเจ้าหนี้ยกมา 45,000 และเจ้าหนี้ SV-001 ครบ 53,500", [dr("210101", "98500"), cr("110102", "98500", cashflow="operating")])
add("PV-002", "2026-09-06", "PV", "ชำระค่าไฟฟ้า 3,000 และน้ำมันรถ 2,500 พร้อมภาษีซื้อสมมติ", [dr("520101", "3000"), dr("520102", "2500"), dr("110401", "385"), cr("110102", "5885", cashflow="operating")], facts={"net": "5500.00", "vat": "385.00", "gross": "5885.00", "vatfraction": "0.07"})
dim = {"department": "BM69-SALES", "project": "BM69-SITE-A"}
add("UV-003", "2026-09-07", "UV", "รับคืนปูน 20 ถุงจาก UV-001 คืนเข้าสต็อกและลดลูกหนี้", [dr("419101", "3000", **dim), dr("210201", "210", **dim), cr("110201", "3210", **dim), dr("110301", "2000", **dim), cr("510101", "2000", **dim)], facts={"originaldoc": PREFIX + "UV-001", "net": "3000.00", "vat": "210.00", "gross": "3210.00", "vatfraction": "0.07", "return": True})
stock_move("UV-003", "CEMENT", "20")
add("SV-004", "2026-09-07", "SV", "คืนสี 5 ถังจาก SV-003 ลดเจ้าหนี้และภาษีซื้อ", [dr("210101", "2140"), cr("110303", "2000"), cr("110401", "140")], facts={"originaldoc": PREFIX + "SV-003", "net": "2000.00", "vat": "140.00", "gross": "2140.00", "vatfraction": "0.07", "return": True})
stock_move("SV-004", "PAINT", "-5")
add("JV-001", "2026-09-08", "JV", "เจ้าของสมมตินำเงินทุนเพิ่มเข้าธนาคาร", [dr("110102", "50000", cashflow="financing"), cr("310101", "50000")])
add("PV-003", "2026-09-09", "PV", "ผ่อนเงินกู้ เงินต้น 8,000 และดอกเบี้ย 2,000 แยกกิจกรรมกระแสเงินสด", [dr("220101", "8000"), dr("520105", "2000"), cr("110102", "8000", cashflow="financing"), cr("110102", "2000", cashflow="operating")])
add("JV-002", "2026-09-10", "JV", "ค่าเสื่อมสินทรัพย์เดิมเดือนกันยายน อุปกรณ์ 1,000 และรถ 4,000 ตามตารางสมมติ", [dr("520103", "5000"), cr("129101", "1000"), cr("129102", "4000")])
add("JV-003", "2026-09-10", "JV", "ตัดเบี้ยประกันจ่ายล่วงหน้าเดือนกันยายนตามตารางสมมติ", [dr("520104", "1000"), cr("110501", "1000")])
add("PV-004", "2026-09-10", "PV", "ซื้ออุปกรณ์คลังสมมติ 20,000 พร้อมภาษีซื้อ บันทึกสินทรัพย์ใหม่ ยังไม่คิดค่าเสื่อมในชุดนี้", [dr("120101", "20000"), dr("110401", "1400"), cr("110102", "21400", cashflow="investing")], facts={"net": "20000.00", "vat": "1400.00", "gross": "21400.00", "vatfraction": "0.07"})
add("JV-DRAFT-001", "2026-09-11", "JV", "ร่างตรวจสอบปูนเสียหาย 3 ถุง ยังไม่ผ่านบัญชีและยังไม่ลดสต็อก", [dr("520106", "300"), cr("110301", "300")], post=False, facts={"excludedfromexpectedactuals": True})

budgets = []
for account, amount in [("410101", "100000"), ("410102", "100000"), ("410103", "50000"), ("510101", "70000"), ("510102", "80000"), ("510103", "30000"), ("520101", "4000"), ("520102", "3000"), ("520103", "5000"), ("520104", "1000"), ("520105", "2500")]:
    budgets.append({"code": PREFIX + "BUD-" + account, "name": "ข้อมูลตัวอย่าง - งบกันยายน " + account_map[code(account)]["names"][0]["name"], "isactive": True, "accountcode": code(account), "fiscalyear": FY, "startdate": "2026-09-01", "enddate": "2026-09-30", "amount": money(amount), "branchcode": BRANCH})
forecast = [{"code": PREFIX + f"CF-{n:03d}", "name": "ข้อมูลตัวอย่าง - " + name, "isactive": True, "accountcode": code("110102"), "fiscalyear": FY,
             "startdate": date, "enddate": date, "direction": direction, "amount": money(amount), "branchcode": BRANCH}
            for n, (date, direction, amount, name) in enumerate([
                ("2026-09-15", "in", "80000", "คาดเก็บลูกหนี้"), ("2026-09-17", "out", "40000", "คาดชำระเจ้าหนี้"),
                ("2026-09-20", "in", "42800", "คาดรับขายหน้าร้านรวมภาษี"), ("2026-09-24", "out", "32100", "คาดซื้อสินค้าเพิ่มรวมภาษี"),
                ("2026-09-28", "out", "6500", "คาดจ่ายค่าใช้จ่ายร้าน")], 1)]
product_groups = [{"code": "BM69-" + sku, "name": "ข้อมูลตัวอย่าง - " + item["name"], "isactive": True, "amount": "0.00", "itemaccount": code(item["inventory"]), "costaccount": code(item["cogs"]), "revenueaccount": code(item["sales"])} for sku, item in items.items()]

balances, opening, debits, credits = (defaultdict(lambda: ZERO) for _ in range(4))
cash_categories = defaultdict(lambda: {"receipts": ZERO, "payments": ZERO})
journal_checks = []
for item in journals:
    journal = item["journal"]
    if not item["post"]:
        continue
    journal_checks.append({"docno": journal["docno"], "debit": item["expecteddebit"], "credit": item["expectedcredit"]})
    for row in journal["lines"]:
        account, debit, credit = row["accountcode"], D(row["debit"]), D(row["credit"])
        balances[account] += debit - credit
        if journal["kind"] == "opening":
            opening[account] += debit - credit
        else:
            debits[account] += debit
            credits[account] += credit
            if account_map[account]["iscash"]:
                assert row["cashflow"] in ("operating", "investing", "financing")
                cash_categories[row["cashflow"]]["receipts"] += debit
                cash_categories[row["cashflow"]]["payments"] += credit
revenue = -sum((v for a, v in balances.items() if account_map[a]["accounttype"] == "income"), ZERO)
expense = sum((v for a, v in balances.items() if account_map[a]["accounttype"] == "expense"), ZERO)
profit = revenue - expense
assets = sum((v for a, v in balances.items() if account_map[a]["accounttype"] == "asset"), ZERO)
liabilities = -sum((v for a, v in balances.items() if account_map[a]["accounttype"] == "liability"), ZERO)
equity = -sum((v for a, v in balances.items() if account_map[a]["accounttype"] == "equity"), ZERO) + profit
cash = sum((v for a, v in balances.items() if account_map[a]["iscash"]), ZERO)
stock = sum((quantity * items[sku]["cost"] for sku, quantity in stock_quantity.items()), ZERO)
assert sum(balances.values(), ZERO) == ZERO and assets == liabilities + equity
assert profit == D("52000") and stock == D("197000") and cash == D("341395")
assert stock == sum((balances[code(items[sku]["inventory"])] for sku in items), ZERO)
trial_rows = [{"accountcode": a, "accounttype": account_map[a]["accounttype"], "openingdebit": money(max(opening[a], ZERO)), "openingcredit": money(max(-opening[a], ZERO)), "debit": money(debits[a]), "credit": money(credits[a]), "endingdebit": money(max(v, ZERO)), "endingcredit": money(max(-v, ZERO)), "balance": money(v)} for a, v in sorted(balances.items())]
trial_totals = {key: money(sum((D(row[key]) for row in trial_rows), ZERO)) for key in ("openingdebit", "openingcredit", "debit", "credit", "endingdebit", "endingcredit", "balance")}
trial_totals["difference"] = "0.00"
cash_opening = sum((v for a, v in opening.items() if account_map[a]["iscash"]), ZERO)
plan_in = sum((D(p["amount"]) for p in forecast if p["direction"] == "in"), ZERO)
plan_out = sum((D(p["amount"]) for p in forecast if p["direction"] == "out"), ZERO)
expected = {"reportquery": {"fiscalyear": FY, "from": "2026-09-01", "to": "2026-09-11", "branchcode": BRANCH},
            "counts": {"accountgroups": len(groups), "accounts": len(accounts), "periods": len(periods), "productaccountgroups": len(product_groups), "budgets": len(budgets), "forecast": len(forecast), "postedjournals": len(journal_checks), "draftjournals": len(journals) - len(journal_checks)},
            "pnl": {"revenue": money(revenue), "expense": money(expense), "profit": money(profit), "cogs": money(sum((balances[code(a)] for a in ("510101", "510102", "510103")), ZERO)), "grossprofit": money(revenue - sum((balances[code(a)] for a in ("510101", "510102", "510103")), ZERO))},
            "balancesheet": {"assets": money(assets), "liabilities": money(liabilities), "equity": money(equity), "currentearnings": money(profit), "difference": "0.00"},
            "trialbalance": {"totals": trial_totals, "rows": trial_rows},
            "controlbalances": {"inventory": money(stock), "ar": money(balances[code("110201")]), "ap": money(-balances[code("210101")]), "cashandbanks": money(cash), "cash": money(balances[code("110101")]), "bank": money(balances[code("110102")]), "pettycash": money(balances[code("110103")]), "inputvat": money(balances[code("110401")]), "outputvat": money(-balances[code("210201")]), "netvatinformationalonly": money(-balances[code("210201")] - balances[code("110401")])},
            "cashflow": {"opening": money(cash_opening), "closing": money(cash), "net": money(cash-cash_opening), "unclassifiedlines": "0", "categories": {key: {"receipts": money(v["receipts"]), "payments": money(v["payments"]), "net": money(v["receipts"]-v["payments"])} for key, v in cash_categories.items()}},
            "cashflowforecast": {"query": {"fiscalyear": FY, "from": "2026-09-15", "to": "2026-09-30", "branchcode": BRANCH}, "opening": money(cash), "receipts": money(plan_in), "payments": money(plan_out), "net": money(plan_in-plan_out), "closing": money(cash+plan_in-plan_out)},
            "inventoryschedule": [{"sku": "BM69-" + sku, "name": item["name"], "unit": item["unit"], "openingquantity": money(item["opening"]), "closingquantity": money(stock_quantity[sku]), "unitcost": money(item["cost"]), "openingvalue": money(item["opening"]*item["cost"]), "closingvalue": money(stock_quantity[sku]*item["cost"])} for sku, item in items.items()],
            "arallocations": [{"source": PREFIX+"OPEN-001", "remaining": "20000.00"}, {"source": PREFIX+"UV-001", "remaining": "48730.00"}, {"source": PREFIX+"UV-002", "remaining": "67945.00"}],
            "apallocations": [{"source": PREFIX+"OPEN-001", "remaining": "40000.00"}, {"source": PREFIX+"SV-001", "remaining": "0.00"}, {"source": PREFIX+"SV-003", "remaining": "19260.00"}],
            "journalchecks": journal_checks}
assert sum((D(r["remaining"]) for r in expected["arallocations"]), ZERO) == balances[code("110201")]
assert sum((D(r["remaining"]) for r in expected["apallocations"]), ZERO) == -balances[code("210101")]
fixture = {"scenario": {"name": "ข้อมูลตัวอย่างร้านวัสดุก่อสร้างในไทย", "documentprefix": PREFIX, "currency": "THB", "scale": 2,
                       "vatfraction": "0.07", "vatrounding": "ROUND_HALF_UP at document VAT; all chosen VAT results are exact to two decimals, so no rounding occurs",
                       "vatsource": "https://www.rd.go.th/fileadmin/user_upload/news/2568thai/news34_2568.pdf", "synthetic": True,
                       "scopebinding": "No holding/company/actor/record identity is included. The authenticated importer supplies its permitted scope.",
                       "stockpolicy": "Perpetual inventory. Constant unit costs per fictional SKU; purchases, sales costs and both returns reconcile exactly. The draft damage journal is excluded.",
                       "limitations": ["เป็นข้อมูลสมมติ ไม่มีเลขประจำตัวผู้เสียภาษีหรือบุคคลจริง", "ทะเบียนสินค้าและลูกหนี้เจ้าหนี้ในไฟล์นี้เป็นตารางประกอบตรวจยอด GL ไม่ใช่การสร้างเอกสารในโมดูลซื้อขายเดิม", "ไม่สร้างเงินเดือน ไม่ปิดปี ไม่โอนกำไร และไม่บันทึกชำระภาษี", "ค่าเสื่อมและเบี้ยประกันใช้ตารางสมมติสำหรับเดือนกันยายน ไม่ใช่คำแนะนำคำนวณภาษี", "แผน cash forecast ไม่ผ่านบัญชีและไม่รวมในยอดจริง"]},
           "accountgroups": groups, "accounts": accounts, "fiscalyear": year, "periods": periods, "productaccountgroups": product_groups,
           "budgets": budgets, "forecast": forecast, "journals": journals, "stockmovements": stock_moves}
for name, data in [("gl-thai-construction-2026.json", fixture), ("gl-thai-construction-2026-expected.json", expected)]:
    (ROOT / name).write_text(json.dumps(data, ensure_ascii=False, indent=2) + "\n", encoding="utf-8")
print(json.dumps({"counts": expected["counts"], "pnl": expected["pnl"], "balancesheet": expected["balancesheet"], "controlbalances": expected["controlbalances"], "cashflowforecast": expected["cashflowforecast"]}, ensure_ascii=False, indent=2))
