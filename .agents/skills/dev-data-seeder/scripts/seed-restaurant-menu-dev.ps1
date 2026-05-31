#Requires -Version 7.0

param(
  [Parameter(Mandatory = $true)][string]$ShopId,
  [Parameter(Mandatory = $true)][string]$Username,
  [string]$BaseUrl = "http://localhost:8888",
  [int]$Count = 200
)

$ErrorActionPreference = "Stop"

function Get-SelectedToken {
  param([string]$ShopId, [string]$Username)
  $keys = docker exec redis redis-cli --scan --pattern "auth-*"
  foreach ($key in $keys) {
    if ([string]::IsNullOrWhiteSpace($key)) { continue }
    $values = docker exec redis redis-cli HMGET $key username shopid
    if ($values[0] -eq $Username -and $values[1] -eq $ShopId) {
      return $key.Substring(5)
    }
  }
  throw "No selected auth token found for shop $ShopId"
}

function Invoke-JsonApi {
  param([string]$Method, [string]$Path, [object]$Payload, [string]$Token)
  $bodyFile = [IO.Path]::Combine($env:TEMP, "bc-seed-body-$([guid]::NewGuid()).json")
  $outFile = [IO.Path]::Combine($env:TEMP, "bc-seed-out-$([guid]::NewGuid()).json")
  try {
    $Payload | ConvertTo-Json -Depth 50 | Set-Content -LiteralPath $bodyFile -Encoding UTF8
    $status = curl.exe --max-time 30 -sS -o $outFile -w "%{http_code}" -X $Method `
      -H "Authorization: Bearer $Token" `
      -H "Content-Type: application/json" `
      --data-binary "@$bodyFile" "$BaseUrl$Path"
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

function Invoke-GetJson {
  param([string]$Path, [string]$Token)
  $outFile = [IO.Path]::Combine($env:TEMP, "bc-seed-get-$([guid]::NewGuid()).json")
  try {
    $status = curl.exe --max-time 30 -sS -o $outFile -w "%{http_code}" `
      -H "Authorization: Bearer $Token" "$BaseUrl$Path"
    $raw = if (Test-Path $outFile) { Get-Content -LiteralPath $outFile -Raw } else { "" }
    if ([int]$status -lt 200 -or [int]$status -ge 300) {
      throw "HTTP $status $Path $raw"
    }
    if ([string]::IsNullOrWhiteSpace($raw)) { return $null }
    return $raw | ConvertFrom-Json
  }
  finally {
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

function New-RestaurantFlags {
  return @{
    isforrestaurant = $true
    isfortakeaway = $true
    isfordelivery = $true
    isforcustomer = $true
    isforcustomerpreorder = $false
  }
}

function New-ProductPayload {
  param(
    [string]$Code,
    [string]$Thai,
    [string]$English,
    [int]$ItemType,
    [int]$MaterialType,
    [hashtable]$Unit,
    [decimal]$Price
  )
  return @{
    code = $Code
    names = @(New-Names $Thai $English)
    group_code = "RESTAURANT"
    group_names = @(New-Names "ร้านอาหาร" "Restaurant")
    item_type = $ItemType
    materialtype = $MaterialType
    vat_type = 0
    tax_type = 0
    useimageorcolor = $false
    restaurant = New-RestaurantFlags
    condition = $false
    dividevalue = 1
    standvalue = 1
    qty = 0
    barcodes = @(
      @{
        barcode = ""
        item_unit_code = $Unit.code
        itemunitnames = @(New-Names $Unit.th $Unit.en)
        prices = @(@{ key_number = 1; price = [double]$Price })
        condition = $false
        dividevalue = 1
        standvalue = 1
        qty = 1
        is_main_barcode = $true
      }
    )
  }
}

function New-BarcodePayload {
  param(
    [string]$Code,
    [string]$Barcode,
    [string]$Thai,
    [string]$English,
    [int]$ItemType,
    [int]$MaterialType,
    [hashtable]$Unit,
    [decimal]$Price,
    [array]$Bom = @()
  )
  return @{
    itemcode = $Code
    barcode = $Barcode
    group_code = "RESTAURANT"
    group_names = @(New-Names "ร้านอาหาร" "Restaurant")
    names = @(New-Names $Thai $English)
    xsorts = @()
    item_unit_code = $Unit.code
    itemunitnames = @(New-Names $Unit.th $Unit.en)
    prices = @(@{ key_number = 1; price = [double]$Price })
    useimageorcolor = $false
    condition = $false
    dividevalue = 1
    standvalue = 1
    isusesubbarcodes = $false
    is_main_barcode = $true
    item_type = $ItemType
    materialtype = $MaterialType
    tax_type = 0
    vat_type = 0
    qty = 0
    restaurant = New-RestaurantFlags
    refbarcodes = @()
    bom = @($Bom)
    ignorebranches = @()
    businesstypes = @()
  }
}

$thaiBases = @(
  "ข้าวผัดหมู","ข้าวผัดไก่","ข้าวผัดกุ้ง","ข้าวผัดปู","กะเพราหมูสับราดข้าว","กะเพราไก่ราดข้าว","กะเพราทะเลราดข้าว","กะเพราเนื้อราดข้าว","พริกแกงหมูราดข้าว","พริกแกงไก่ราดข้าว",
  "พริกแกงทะเลราดข้าว","พริกแกงหมูกรอบราดข้าว","หมูกระเทียมราดข้าว","ไก่กระเทียมราดข้าว","กุ้งกระเทียมราดข้าว","เนื้อกระเทียมราดข้าว","ผัดผักรวมราดข้าว","คะน้าหมูกรอบราดข้าว","ผัดบรอกโคลีไก่ราดข้าว","ผัดผักบุ้งไฟแดงราดข้าว",
  "ข้าวต้มหมู","ข้าวต้มปลา","โจ๊กหมู","โจ๊กไก่","ก๋วยเตี๋ยวหมูน้ำใส","ก๋วยเตี๋ยวต้มยำ","ก๋วยเตี๋ยวเนื้อตุ๋น","เย็นตาโฟ","ผัดไทยกุ้งสด","ผัดไทยหมู",
  "สุกี้น้ำ","สุกี้แห้ง","ราดหน้าหมู","ราดหน้าไก่","ผัดซีอิ๊วหมู","ผัดซีอิ๊วไก่","มาม่าผัดขี้เมา","เส้นใหญ่ผัดขี้เมา","ต้มยำกุ้ง","ต้มยำไก่",
  "ต้มข่าไก่","แกงเขียวหวานไก่","แกงส้มชะอมกุ้ง","พะแนงหมู","มัสมั่นไก่","แกงจืดเต้าหู้หมูสับ","ไข่เจียวหมูสับ","ไข่ดาว","ไข่ต้ม","ไข่พะโล้",
  "ส้มตำไทย","ส้มตำปูปลาร้า","ตำแตง","ตำข้าวโพด","ลาบหมู","น้ำตกหมู","คอหมูย่าง","ไก่ย่าง","หมูแดดเดียว","ปีกไก่ทอด",
  "เฟรนช์ฟรายส์","นักเก็ตไก่","เกี๊ยวซ่า","ปอเปี๊ยะทอด","ลูกชิ้นทอด","ไส้กรอกทอด","ยำวุ้นเส้น","ยำมาม่า","ยำทะเล","ยำหมูยอ",
  "ชาไทยเย็น","ชาเขียวเย็น","กาแฟเย็น","โอเลี้ยง","โกโก้เย็น","นมชมพู","น้ำเปล่า","น้ำอัดลม","น้ำมะนาว","น้ำผึ้งมะนาว",
  "น้ำสมุนไพร","น้ำเก๊กฮวย","น้ำกระเจี๊ยบ","น้ำลำไย","น้ำส้มคั้น","น้ำแตงโมปั่น","น้ำมะพร้าว","ไอศกรีมกะทิ","เฉาก๊วย","บัวลอย"
)

$englishBases = @(
  "Pork Fried Rice","Chicken Fried Rice","Shrimp Fried Rice","Crab Fried Rice","Basil Pork Rice","Basil Chicken Rice","Basil Seafood Rice","Basil Beef Rice","Red Curry Pork Rice","Red Curry Chicken Rice",
  "Red Curry Seafood Rice","Red Curry Crispy Pork Rice","Garlic Pork Rice","Garlic Chicken Rice","Garlic Shrimp Rice","Garlic Beef Rice","Mixed Vegetable Rice","Chinese Kale Crispy Pork Rice","Broccoli Chicken Rice","Morning Glory Rice",
  "Pork Rice Soup","Fish Rice Soup","Pork Congee","Chicken Congee","Clear Pork Noodles","Tom Yum Noodles","Beef Stew Noodles","Yentafo Noodles","Pad Thai Shrimp","Pad Thai Pork",
  "Suki Soup","Dry Suki","Pork Gravy Noodles","Chicken Gravy Noodles","Pork Soy Sauce Noodles","Chicken Soy Sauce Noodles","Spicy Instant Noodles","Spicy Wide Noodles","Tom Yum Shrimp","Tom Yum Chicken",
  "Chicken Coconut Soup","Green Curry Chicken","Sour Curry Shrimp Omelet","Pork Panang Curry","Chicken Massaman Curry","Tofu Pork Soup","Pork Omelet","Fried Egg","Boiled Egg","Stewed Egg",
  "Thai Papaya Salad","Papaya Salad with Fermented Fish","Cucumber Salad","Corn Salad","Pork Larb","Pork Nam Tok","Grilled Pork Neck","Grilled Chicken","Sun-Dried Pork","Fried Chicken Wings",
  "French Fries","Chicken Nuggets","Gyoza","Spring Rolls","Fried Meatballs","Fried Sausage","Glass Noodle Salad","Instant Noodle Salad","Seafood Salad","Pork Sausage Salad",
  "Thai Iced Tea","Iced Green Tea","Iced Coffee","Thai Black Coffee","Iced Cocoa","Pink Milk","Water","Soft Drink","Lime Juice","Honey Lime Juice",
  "Herbal Drink","Chrysanthemum Drink","Roselle Drink","Longan Drink","Orange Juice","Watermelon Smoothie","Coconut Water","Coconut Ice Cream","Grass Jelly","Bua Loy"
)

$units = @{
  plate = @{ code = "PLATE"; th = "จาน"; en = "Plate" }
  bowl = @{ code = "BOWL"; th = "ชาม"; en = "Bowl" }
  glass = @{ code = "GLASS"; th = "แก้ว"; en = "Glass" }
  set = @{ code = "SET"; th = "ชุด"; en = "Set" }
}

function Resolve-Unit {
  param([string]$Name)
  if ($Name -match "ชา|กาแฟ|น้ำ|โกโก้|โอเลี้ยง|นม") { return $units.glass }
  if ($Name -match "ก๋วยเตี๋ยว|โจ๊ก|ข้าวต้ม|สุกี้|แกง|ต้ม") { return $units.bowl }
  return $units.plate
}

$token = Get-SelectedToken -ShopId $ShopId -Username $Username

$existing = Invoke-GetJson -Path "/product?limit=500&q=MENU-" -Token $token
$existingCodes = @{}
foreach ($row in @($existing.data)) {
  if ($row.code) { $existingCodes[$row.code] = $true }
}

$items = @()
$productsCreated = 0
$productsSkipped = 0
for ($i = 1; $i -le $Count; $i++) {
  $idx = ($i - 1) % $thaiBases.Count
  $round = [math]::Floor(($i - 1) / $thaiBases.Count)
  $suffixTh = if ($round -gt 0) { " พิเศษ $round" } else { "" }
  $suffixEn = if ($round -gt 0) { " Special $round" } else { "" }
  $code = "MENU-{0:D4}" -f $i
  $barcode = "885260529{0:D4}" -f $i
  $thai = $thaiBases[$idx] + $suffixTh
  $english = $englishBases[$idx] + $suffixEn
  $unit = Resolve-Unit $thai
  $price = [decimal](59 + (($i - 1) % 5) * 10)
  $items += [pscustomobject]@{ code = $code; barcode = $barcode; thai = $thai; english = $english; unit = $unit; price = $price }
  if ($existingCodes.ContainsKey($code)) {
    $productsSkipped++
    continue
  }
  $payload = New-ProductPayload -Code $code -Thai $thai -English $english -ItemType 0 -MaterialType 0 -Unit $unit -Price $price
  [void](Invoke-JsonApi -Method "POST" -Path "/product" -Payload $payload -Token $token)
  $productsCreated++
}

$sets = @(
  @{ code = "SET-0001"; barcode = "8852605299001"; thai = "ชุดข้าวกะเพราพร้อมเครื่องดื่ม"; english = "Basil Rice Drink Set"; refs = @("8852605290005","8852605290006","8852605290071") },
  @{ code = "SET-0002"; barcode = "8852605299002"; thai = "ชุดผัดไทยกุ้งพร้อมของทานเล่น"; english = "Pad Thai Snack Set"; refs = @("8852605290029","8852605290061","8852605290072") },
  @{ code = "SET-0003"; barcode = "8852605299003"; thai = "ชุดต้มยำครอบครัว"; english = "Family Tom Yum Set"; refs = @("8852605290039","8852605290046","8852605290064") },
  @{ code = "SET-0004"; barcode = "8852605299004"; thai = "ชุดส้มตำไก่ย่าง"; english = "Papaya Salad Grilled Chicken Set"; refs = @("8852605290051","8852605290058","8852605290060") },
  @{ code = "SET-0005"; barcode = "8852605299005"; thai = "ชุดของหวานไทย"; english = "Thai Dessert Set"; refs = @("8852605290088","8852605290089","8852605290090") }
)

$existingSets = Invoke-GetJson -Path "/product?limit=100&q=SET-" -Token $token
foreach ($row in @($existingSets.data)) {
  if ($row.code) { $existingCodes[$row.code] = $true }
}

$setsCreated = 0
$setsSkipped = 0
foreach ($set in $sets) {
  if ($existingCodes.ContainsKey($set.code)) {
    $setsSkipped++
    continue
  }
  $payload = New-ProductPayload -Code $set.code -Thai $set.thai -English $set.english -ItemType 2 -MaterialType 3 -Unit $units.set -Price 159
  [void](Invoke-JsonApi -Method "POST" -Path "/product" -Payload $payload -Token $token)
  $setsCreated++
}

$barcodePayloads = @()
foreach ($item in $items) {
  $barcodePayloads += New-BarcodePayload -Code $item.code -Barcode $item.barcode -Thai $item.thai -English $item.english -ItemType 0 -MaterialType 0 -Unit $item.unit -Price $item.price
}
foreach ($set in $sets) {
  $bom = @()
  foreach ($ref in $set.refs) {
    $bom += @{ barcode = $ref; condition = $false; dividevalue = 1; standvalue = 1; qty = 1 }
  }
  $barcodePayloads += New-BarcodePayload -Code $set.code -Barcode $set.barcode -Thai $set.thai -English $set.english -ItemType 2 -MaterialType 3 -Unit $units.set -Price 159 -Bom $bom
}

$barcodeSubmitted = 0
for ($offset = 0; $offset -lt $barcodePayloads.Count; $offset += 40) {
  $end = [math]::Min($offset + 39, $barcodePayloads.Count - 1)
  $chunk = @($barcodePayloads[$offset..$end])
  [void](Invoke-JsonApi -Method "POST" -Path "/product/barcode/bulk" -Payload $chunk -Token $token)
  $barcodeSubmitted += $chunk.Count
}

$setBarcodeUpdated = 0
foreach ($set in $sets) {
  $bom = @()
  foreach ($ref in $set.refs) {
    $bom += @{ barcode = $ref; condition = $false; dividevalue = 1; standvalue = 1; qty = 1 }
  }
  $payload = New-BarcodePayload -Code $set.code -Barcode $set.barcode -Thai $set.thai -English $set.english -ItemType 2 -MaterialType 3 -Unit $units.set -Price 159 -Bom $bom
  $current = Invoke-GetJson -Path "/product/barcode/pk/$($set.barcode)" -Token $token
  $guid = [string]($current.data.guid_fixed ?? $current.data.guidfixed)
  if ([string]::IsNullOrWhiteSpace($guid)) {
    throw "Set barcode $($set.barcode) has no guidfixed after bulk seed"
  }
  [void](Invoke-JsonApi -Method "PUT" -Path "/product/barcode/$guid" -Payload $payload -Token $token)
  $setBarcodeUpdated++
}

Start-Sleep -Seconds 2
$productCheck = Invoke-GetJson -Path "/product?limit=500&q=MENU-" -Token $token
$setCheck = Invoke-GetJson -Path "/product?limit=100&q=SET-" -Token $token
$barcodeCheck = Invoke-GetJson -Path "/product/barcode?limit=500&q=MENU-" -Token $token
$setBarcodeCheck = Invoke-GetJson -Path "/product/barcode/pk/8852605299001" -Token $token

[pscustomobject]@{
  shopid = $ShopId
  products_created = $productsCreated
  products_skipped = $productsSkipped
  sets_created = $setsCreated
  sets_skipped = $setsSkipped
  barcode_submitted = $barcodeSubmitted
  set_barcode_updated = $setBarcodeUpdated
  product_api_count = @($productCheck.data).Count
  set_api_count = @($setCheck.data).Count
  barcode_api_count = @($barcodeCheck.data).Count
  set_barcode_item_type = $setBarcodeCheck.data.item_type
  set_barcode_materialtype = $setBarcodeCheck.data.materialtype
  set_bom_count = @($setBarcodeCheck.data.bom).Count
} | ConvertTo-Json -Depth 6
