#!/bin/bash
# Analyze results from kung_test.sh
OUT=/tmp/kung-test
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

printf "%-3s %-45s %-10s %-8s %-8s %-12s\n" "#" "Question" "HTTP" "time" "tools" "html_ok"
printf "%s\n" "------------------------------------------------------------------------------------------"
for i in "${!QS[@]}"; do
  meta="$OUT/meta-$i.txt"
  resp="$OUT/resp-$i.json"
  http=$(grep -o 'HTTP=[0-9]*' "$meta" | cut -d= -f2)
  time=$(grep -o 'TIME=[0-9.]*' "$meta" | cut -d= -f2)
  tools=$(jq -r '.data.tools_used | length // 0' "$resp" 2>/dev/null)
  ans=$(jq -r '.data.answer // ""' "$resp" 2>/dev/null)
  # Check if answer looks like HTML (contains <h2 or <p> or <table)
  if echo "$ans" | grep -qE '<h[1-6]|<p>|<table|<ul>' ; then
    html_ok="YES"
  else
    html_ok="NO"
  fi
  # Check if markdown leaked
  if echo "$ans" | grep -qE '^(##|###|- \*\*|\| )' ; then
    html_ok="$html_ok-MD-LEAK"
  fi
  printf "%-3s %-45s %-10s %-8s %-8s %-12s\n" "$i" "${QS[$i]:0:43}" "$http" "$time" "$tools" "$html_ok"
done

echo ""
echo "=== Tools used per query ==="
for i in "${!QS[@]}"; do
  resp="$OUT/resp-$i.json"
  tools=$(jq -r '.data.tools_used[]?.name // empty' "$resp" 2>/dev/null | sort | uniq -c | awk '{printf "%s:%s ", $2, $1}')
  printf "[%s] %-45s -> %s\n" "$i" "${QS[$i]:0:43}" "$tools"
done
