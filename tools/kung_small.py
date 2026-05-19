#!/usr/bin/env python3
# -*- coding: utf-8 -*-
"""Smaller focused test — 5 representative questions × 2 rounds, parallelism=1.
Used when upstream providers are flaky and we want a cleaner signal."""
import sys, os
sys.path.insert(0, os.path.dirname(os.path.abspath(__file__)))
if hasattr(sys.stdout, "reconfigure"):
    sys.stdout.reconfigure(encoding="utf-8", errors="replace")  # type: ignore[attr-defined]

import kung_test
kung_test.QUESTIONS = [
    "ROF002 ราคาเท่าไหร่",
    "TIL001 มีในสต็อกไหม",
    "หาสินค้า TOA",
    "ลูกค้าชื่อกระเบื้องทอง มีไหม",
    "ซัพพลายเออร์ปูนซีเมนต์",
    "วิธีลางานพนักงาน",
    "สินค้าราคาแพงที่สุด 5 อันดับ",
    "ยอดขายเดือนนี้",
    "มีลูกหนี้ค้างชำระกี่ราย",
    "สวัสดีครับ",
]

import kung_loop  # type: ignore
sys.argv = [sys.argv[0], "--rounds", "2", "--parallel", "2", "--tag", "iter6"]
kung_loop.main()
