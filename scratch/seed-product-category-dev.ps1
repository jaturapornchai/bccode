#Requires -Version 7.0

param(
  [string]$HoldingCode = "xxx",
  [string]$Username = "jaturapornchai@gmail.com",
  [string]$BaseUrl = "http://localhost:8888",
  [int]$CategoryCount = 80,
  [int]$ProductLimit = 160
)

$ErrorActionPreference = "Stop"

function Get-SelectedToken {
  param([string]$HoldingCode, [string]$Username)
  $keys = docker exec redis redis-cli --scan --pattern "auth-*"
  foreach ($key in $keys) {
    if ([string]::IsNullOrWhiteSpace($key)) { continue }
    $values = docker exec redis redis-cli HMGET $key username holding_code
    if ($values[0] -eq $Username -and $values[1] -eq $HoldingCode) {
      return $key.Substring(5)
    }
  }
  throw "No selected auth token found for $Username / $HoldingCode"
}

function Invoke-JsonApi {
  param([string]$Method, [string]$Path, [object]$Payload, [string]$Token)
  $bodyFile = [IO.Path]::Combine($env:TEMP, "bc-category-seed-body-$([guid]::NewGuid()).json")
  $outFile = [IO.Path]::Combine($env:TEMP, "bc-category-seed-out-$([guid]::NewGuid()).json")
  try {
    $args = @(
      "--max-time", "60",
      "-sS",
      "-o", $outFile,
      "-w", "%{http_code}",
      "-X", $Method,
      "-H", "Authorization: Bearer $Token",
      "-H", "Content-Type: application/json"
    )
    if ($null -ne $Payload) {
      $Payload | ConvertTo-Json -Depth 80 | Set-Content -LiteralPath $bodyFile -Encoding UTF8
      $args += @("--data-binary", "@$bodyFile")
    }
    $args += "$BaseUrl$Path"
    $status = curl.exe @args
    $raw = if (Test-Path $outFile) { Get-Content -LiteralPath $outFile -Raw } else { "" }
    if ([int]$status -lt 200 -or [int]$status -ge 300) {
      throw "HTTP $status $Path $raw"
    }
    if ([string]::IsNullOrWhiteSpace($raw)) { return $null }
    return $raw | ConvertFrom-Json
  }
  finally {
    Remove-Item -LiteralPath $bodyFile -ErrorAction SilentlyContinue
    Remove-Item -LiteralPath $outFile -ErrorAction SilentlyContinue
  }
}

function New-Names {
  param([string]$Thai, [string]$English)
  return @(
    @{ code = "th"; name = $Thai; isauto = $false; isdelete = $false },
    @{ code = "en"; name = $English; isauto = $false; isdelete = $false }
  )
}

function Pick-Name {
  param([object]$Row, [string]$Fallback)
  foreach ($name in @($Row.names)) {
    if ($name.code -eq "th" -and $name.name) { return [string]$name.name }
  }
  foreach ($name in @($Row.names)) {
    if ($name.name) { return [string]$name.name }
  }
  return $Fallback
}

function New-CodeListItem {
  param([object]$BarcodeRow, [int]$Order)
  $barcode = [string]($BarcodeRow.barcode ?? "")
  $code = [string]($BarcodeRow.itemcode ?? $BarcodeRow.code ?? "")
  return @{
    code = $code
    xorder = $Order
    barcode = $barcode
    unitcode = [string]($BarcodeRow.item_unit_code ?? $BarcodeRow.itemunitcode ?? "")
    unitnames = @($BarcodeRow.itemunitnames ?? $BarcodeRow.item_unit_names ?? @())
    names = @($BarcodeRow.names ?? @())
    manufacturerguid = [string]($BarcodeRow.manufacturerguid ?? $BarcodeRow.manufacturer_guid ?? "")
  }
}

$token = Get-SelectedToken -HoldingCode $HoldingCode -Username $Username

$existingCategories = Invoke-JsonApi -Method "GET" -Path "/product/category/list?limit=2000&offset=0" -Payload $null -Token $token
$existingBySeedName = @{}
foreach ($category in @($existingCategories.data)) {
  $name = Pick-Name $category ""
  if ($name -like "ตัวอย่างหมวดสินค้า *") {
    $existingBySeedName[$name] = $true
  }
}

$barcodesResponse = Invoke-JsonApi -Method "GET" -Path "/product/barcode?limit=$ProductLimit&offset=0" -Payload $null -Token $token
$barcodes = @($barcodesResponse.data) | Where-Object {
  ([string]($_.barcode ?? "")).Trim() -ne "" -and ([string]($_.itemcode ?? $_.code ?? "")).Trim() -ne ""
}

$thaiGroups = @(
  "อาหารจานเดียว", "เครื่องดื่มเย็น", "ของทานเล่น", "ของหวาน", "วัตถุดิบสด",
  "เครื่องปรุง", "อุปกรณ์ครัว", "สินค้าแพ็ก", "สินค้าขายดี", "สินค้าแนะนำ",
  "อาหารเช้า", "อาหารกลางวัน", "อาหารเย็น", "กาแฟและชา", "น้ำผลไม้",
  "เบเกอรี่", "ขนมไทย", "สินค้าตามฤดูกาล", "สินค้าโปรโมชั่น", "สินค้าออนไลน์"
)
$englishGroups = @(
  "Single Dish", "Cold Drinks", "Snacks", "Desserts", "Fresh Ingredients",
  "Condiments", "Kitchen Supplies", "Packaged Goods", "Best Sellers", "Recommended",
  "Breakfast", "Lunch", "Dinner", "Coffee and Tea", "Juice",
  "Bakery", "Thai Desserts", "Seasonal", "Promotion", "Online"
)
$colors = @("#0EA5E9", "#10B981", "#F59E0B", "#EF4444", "#8B5CF6", "#14B8A6", "#F97316", "#64748B")

$payload = @()
for ($i = 1; $i -le $CategoryCount; $i++) {
  $index = ($i - 1) % $thaiGroups.Count
  $round = [math]::Floor(($i - 1) / $thaiGroups.Count)
  $thaiName = "ตัวอย่างหมวดสินค้า {0:D3} - {1}{2}" -f $i, $thaiGroups[$index], ($(if ($round -gt 0) { " ชุด $round" } else { "" }))
  if ($existingBySeedName.ContainsKey($thaiName)) { continue }
  $englishName = "Sample Category {0:D3} - {1}{2}" -f $i, $englishGroups[$index], ($(if ($round -gt 0) { " Set $round" } else { "" }))
  $start = (($i - 1) * 3) % [Math]::Max(1, $barcodes.Count)
  $list = @()
  if ($barcodes.Count -gt 0) {
    for ($j = 0; $j -lt [Math]::Min(12, $barcodes.Count); $j++) {
      $list += New-CodeListItem -BarcodeRow $barcodes[($start + $j) % $barcodes.Count] -Order $j
    }
  }
  $payload += @{
    names = @(New-Names $thaiName $englishName)
    group_number = (($i - 1) % 20) + 1
    parent_guid = ""
    parentguidall = ""
    childcount = 0
    xsorts = @()
    codelist = @($list)
    useimageorcolor = $true
    colorselect = ""
    colorselecthex = $colors[($i - 1) % $colors.Count]
    isdisabled = $false
    imageuri = ""
    coveruri = ""
    timeforsales = @()
  }
}

if ($payload.Count -gt 0) {
  [void](Invoke-JsonApi -Method "POST" -Path "/product/category/bulk" -Payload $payload -Token $token)
}

$after = Invoke-JsonApi -Method "GET" -Path "/product/category/list?limit=2000&offset=0" -Payload $null -Token $token
$sampleAfter = @($after.data) | Where-Object { (Pick-Name $_ "") -like "ตัวอย่างหมวดสินค้า *" }

[pscustomobject]@{
  holding_code = $HoldingCode
  username = $Username
  categories_requested = $CategoryCount
  categories_created_this_run = $payload.Count
  sample_categories_total = @($sampleAfter).Count
  category_total = $after.total
  source_barcodes_used = $barcodes.Count
  first_sample = if (@($sampleAfter).Count) { Pick-Name (@($sampleAfter) | Select-Object -First 1) "" } else { "" }
} | ConvertTo-Json -Depth 6
