#!/bin/bash
# Random test runner for น้องกุ้ง agent
# Runs 10 diverse questions and collects timing + tools used + format detection
SHOP="3AEz8tu22GHPpAZ0XhwPFM4fjY9"
URL="http://localhost:8888/goapi/api/v1/chatbot/chat-agent-v2-sync"
OUT=/tmp/kung-test
mkdir -p $OUT

declare -a QS=(
  "ROF002 ราคาเท่าไหร่"
  "หาสินค้า TOA"
  "กระเบื้องลอนคู่ มียี่ห้อไหนบ้าง"
  "ลูกค้าชื่อกระเบื้องทอง มีไหม"
  "ซัพพลายเออร์ปูนซีเมนต์มีใครบ้าง"
  "วิธีลางานพนักงาน"
  "กระเบื้องระเบิด สาเหตุและวิธีป้องกัน"
  "สินค้าราคาแพงที่สุด 5 อันดับ"
  "ปูนซีเมนต์ตราอินทรี ราคาตลาดเท่าไหร่"
  "มีลูกหนี้ค้างชำระกี่ราย"
)

run_one() {
  local idx=$1
  local q=$2
  local sid="rand-$(date +%s)-$idx"
  local start=$(date +%s.%N)
  curl -s -X POST "$URL" \
    -H "Content-Type: application/json" \
    -d "{\"holdingcode\":\"$SHOP\",\"question\":\"$q\",\"sessionid\":\"$sid\",\"outputformat\":\"html\"}" \
    -o "$OUT/resp-$idx.json" \
    -w "HTTP=%{http_code} TIME=%{time_total}\n" > "$OUT/meta-$idx.txt"
  local end=$(date +%s.%N)
  echo "[$idx] done: $q"
}

export -f run_one
export SHOP URL OUT

# Run all in parallel
for i in "${!QS[@]}"; do
  run_one "$i" "${QS[$i]}" &
done
wait

echo "=== All done ==="
