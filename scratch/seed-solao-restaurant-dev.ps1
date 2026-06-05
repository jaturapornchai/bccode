#Requires -Version 7.0

param(
  [Parameter(Mandatory = $true)][string]$HoldingCode,
  [Parameter(Mandatory = $true)][string]$Username,
  [string]$BaseUrl = "http://localhost:8888",
  [int]$ProductCount = 220
)

$ErrorActionPreference = "Stop"

function Get-SelectedToken {
  param([string]$HoldingCode, [string]$Username)
  $keys = docker exec redis redis-cli --scan --pattern "auth-*"
  foreach ($key in $keys) {
    if ([string]::IsNullOrWhiteSpace($key)) { continue }
    $values = docker exec redis redis-cli HMGET $key username holdingcode
    if ($values[0] -eq $Username -and $values[1] -eq $HoldingCode) {
      return $key.Substring(5)
    }
  }
  throw "No selected auth token found for holdingcode $HoldingCode"
}

function ConvertTo-QueryText {
  param([string]$Value)
  return [System.Uri]::EscapeDataString($Value)
}

function Invoke-JsonApi {
  param(
    [string]$Method,
    [string]$Path,
    [object]$Payload,
    [string]$Token
  )

  $outFile = [IO.Path]::Combine($env:TEMP, "bc-solao-out-$([guid]::NewGuid()).json")
  $bodyFile = $null
  try {
    $args = @(
      "--max-time", "45",
      "-sS",
      "-o", $outFile,
      "-w", "%{http_code}",
      "-X", $Method,
      "-H", "Authorization: Bearer $Token",
      "-H", "Content-Type: application/json"
    )
    if ($null -ne $Payload) {
      $bodyFile = [IO.Path]::Combine($env:TEMP, "bc-solao-body-$([guid]::NewGuid()).json")
      $Payload | ConvertTo-Json -Depth 80 | Set-Content -LiteralPath $bodyFile -Encoding UTF8
      $args += @("--data-binary", "@$bodyFile")
    }
    $args += "$BaseUrl$Path"
    $status = & curl.exe @args
    $raw = if (Test-Path $outFile) { Get-Content -LiteralPath $outFile -Raw } else { "" }
    if ([int]$status -lt 200 -or [int]$status -ge 300) {
      throw "HTTP $status $Path $raw"
    }
    if ([string]::IsNullOrWhiteSpace($raw)) { return $null }
    return $raw | ConvertFrom-Json
  }
  finally {
    if ($bodyFile) { Remove-Item -LiteralPath $bodyFile -ErrorAction SilentlyContinue }
    Remove-Item -LiteralPath $outFile -ErrorAction SilentlyContinue
  }
}

function Invoke-GetJson {
  param([string]$Path, [string]$Token)
  return Invoke-JsonApi -Method "GET" -Path $Path -Payload $null -Token $Token
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

function Get-DocGuid {
  param([object]$Doc)
  foreach ($name in @("guidfixed", "guidfixed", "GuidFixed", "guid")) {
    if ($Doc.PSObject.Properties.Name -contains $name) {
      $value = [string]$Doc.$name
      if (-not [string]::IsNullOrWhiteSpace($value)) { return $value }
    }
  }
  return ""
}

function Get-DocCode {
  param([object]$Doc)
  foreach ($name in @("code", "itemcode", "barcode")) {
    if ($Doc.PSObject.Properties.Name -contains $name) {
      $value = [string]$Doc.$name
      if (-not [string]::IsNullOrWhiteSpace($value)) { return $value }
    }
  }
  return ""
}

function Get-DocThaiName {
  param([object]$Doc)
  if ($Doc.PSObject.Properties.Name -notcontains "names" -or $null -eq $Doc.names) { return "" }
  foreach ($name in @($Doc.names)) {
    if ($name.code -eq "th" -and -not [string]::IsNullOrWhiteSpace([string]$name.name)) {
      return [string]$name.name
    }
  }
  $first = @($Doc.names) | Select-Object -First 1
  if ($first) { return [string]$first.name }
  return ""
}

function Get-CodeListCount {
  param([object]$Doc)
  if ($null -eq $Doc -or $Doc.PSObject.Properties.Name -notcontains "codelist" -or $null -eq $Doc.codelist) { return 0 }
  return @($Doc.codelist).Count
}

function Get-UniqueRows {
  param([array]$Rows)
  $seen = @{}
  $result = @()
  foreach ($row in @($Rows)) {
    $guid = Get-DocGuid $row
    $key = if ([string]::IsNullOrWhiteSpace($guid)) { Get-DocCode $row } else { $guid }
    if ([string]::IsNullOrWhiteSpace($key) -or $seen.ContainsKey($key)) { continue }
    $seen[$key] = $true
    $result += $row
  }
  return @($result)
}

function Search-Products {
  param([string]$Token, [string]$Query, [int]$Limit = 500)
  $q = ConvertTo-QueryText $Query
  $res = Invoke-GetJson -Path "/product?page=1&limit=$Limit&q=$q" -Token $Token
  return @($res.data)
}

function Search-Barcodes {
  param([string]$Token, [string]$Query, [int]$Limit = 1000)
  $q = ConvertTo-QueryText $Query
  $res = Invoke-GetJson -Path "/product/barcode?page=1&limit=$Limit&q=$q" -Token $Token
  return @($res.data)
}

function Search-Categories {
  param([string]$Token, [string]$Query, [int]$Limit = 1000)
  $q = ConvertTo-QueryText $Query
  $res = Invoke-GetJson -Path "/product/category/list?offset=0&limit=$Limit&q=$q&lang=th" -Token $Token
  return @($res.data)
}

function Remove-GuidBatch {
  param([string]$Token, [string]$Path, [array]$Guids)
  $deleted = 0
  $errors = @()
  $unique = @($Guids | Where-Object { -not [string]::IsNullOrWhiteSpace([string]$_) } | Select-Object -Unique)
  for ($offset = 0; $offset -lt $unique.Count; $offset += 80) {
    $end = [math]::Min($offset + 79, $unique.Count - 1)
    $chunk = @($unique[$offset..$end])
    try {
      [void](Invoke-JsonApi -Method "DELETE" -Path $Path -Payload $chunk -Token $Token)
      $deleted += $chunk.Count
    }
    catch {
      $errors += $_.Exception.Message
    }
  }
  return [pscustomobject]@{ deleted = $deleted; errors = @($errors) }
}

function Remove-ProductRows {
  param([string]$Token, [array]$Rows)
  $deleted = 0
  $errors = @()
  foreach ($row in @(Get-UniqueRows $Rows)) {
    $guid = Get-DocGuid $row
    if ([string]::IsNullOrWhiteSpace($guid)) { continue }
    try {
      [void](Invoke-JsonApi -Method "DELETE" -Path "/product/$guid" -Payload $null -Token $Token)
      $deleted++
    }
    catch {
      $errors += "$(Get-DocCode $row): $($_.Exception.Message)"
    }
  }
  return [pscustomobject]@{ deleted = $deleted; errors = @($errors) }
}

function Resolve-Unit {
  param([string]$ThaiName)
  if ($ThaiName -match "น้ำเปล่า|โค้ก|สไปรท์|โซดา") { return @{ code = "BOTTLE"; th = "ขวด"; en = "Bottle" } }
  if ($ThaiName -match "ชา|กาแฟ|น้ำ|โกโก้|อัญชัน|ลำไย|กระเจี๊ยบ") { return @{ code = "GLASS"; th = "แก้ว"; en = "Glass" } }
  if ($ThaiName -match "ต้ม|แกง|ข้าวซอย|ขนมจีน|ซุป|โจ๊ก|ก๋วยเตี๋ยว") { return @{ code = "BOWL"; th = "ชาม"; en = "Bowl" } }
  if ($ThaiName -match "ชุด|ถาด|สำรับ") { return @{ code = "SET"; th = "ชุด"; en = "Set" } }
  return @{ code = "DISH"; th = "จาน"; en = "Dish" }
}

function New-ProductPayload {
  param([object]$Item)
  return @{
    code = $Item.code
    names = @(New-Names $Item.th $Item.en)
    group_code = "SOLAO_RESTAURANT"
    group_names = @(New-Names "ร้านอาหารโซลาว" "Solao Restaurant")
    item_type = [int]$Item.item_type
    materialtype = [int]$Item.materialtype
    vat_type = 0
    tax_type = 0
    useimageorcolor = $false
    isstockforrestaurant = $false
    restaurant = New-RestaurantFlags
    condition = $false
    dividevalue = 1
    standvalue = 1
    qty = 0
    barcodes = @(
      @{
        barcode = $Item.barcode
        item_unit_code = $Item.unit.code
        itemunitnames = @(New-Names $Item.unit.th $Item.unit.en)
        prices = @(@{ key_number = 1; price = [double]$Item.price })
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
  param([object]$Item)
  return @{
    itemcode = $Item.code
    barcode = $Item.barcode
    group_code = "SOLAO_RESTAURANT"
    group_names = @(New-Names "ร้านอาหารโซลาว" "Solao Restaurant")
    names = @(New-Names $Item.th $Item.en)
    xsorts = @()
    item_unit_code = $Item.unit.code
    itemunitnames = @(New-Names $Item.unit.th $Item.unit.en)
    prices = @(@{ key_number = 1; price = [double]$Item.price })
    useimageorcolor = $false
    condition = $false
    dividevalue = 1
    standvalue = 1
    isusesubbarcodes = $false
    is_main_barcode = $true
    item_type = [int]$Item.item_type
    materialtype = [int]$Item.materialtype
    tax_type = 0
    vat_type = 0
    qty = 0
    restaurant = New-RestaurantFlags
    refbarcodes = @()
    bom = @($Item.bom)
    ignorebranches = @()
    businesstypes = @()
  }
}

function New-CodeListFromItems {
  param(
    [array]$Items,
    [int]$Limit = 0
  )
  $codeList = @()
  $xorder = 1
  $source = if ($Limit -gt 0) { @($Items | Select-Object -First $Limit) } else { @($Items) }
  foreach ($item in $source) {
    $codeList += @{
      code = $item.code
      xorder = $xorder
      barcode = $item.barcode
      unitcode = $item.unit.code
      unitnames = @(New-Names $item.unit.th $item.unit.en)
      names = @(New-Names $item.th $item.en)
      manufacturerguid = ""
    }
    $xorder++
  }
  return @($codeList)
}

function New-CategoryPayload {
  param(
    [int]$GroupNumber,
    [string]$Thai,
    [string]$English,
    [string]$ParentGuid,
    [string]$ParentGuidAll,
    [int]$XOrder,
    [string]$Color,
    [array]$Items = @()
  )
  return @{
    childcount = 0
    parent_guid = $ParentGuid
    parentguidall = $ParentGuidAll
    imageuri = ""
    names = @(New-Names $Thai $English)
    xsorts = @(@{ code = "X"; xorder = $XOrder })
    codelist = @(New-CodeListFromItems -Items $Items)
    useimageorcolor = $true
    colorselecthex = $Color
    isdisabled = $false
    coveruri = ""
    group_number = $GroupNumber
    timeforsales = @()
  }
}

function Get-ResponseGuid {
  param([object]$Response)
  if ($null -eq $Response) { return "" }
  if ($Response -is [string]) { return [string]$Response }
  foreach ($name in @("id", "ID", "guidfixed", "guidfixed", "guid")) {
    if ($Response.PSObject.Properties.Name -contains $name) {
      $value = [string]$Response.$name
      if (-not [string]::IsNullOrWhiteSpace($value)) { return $value }
    }
  }
  if ($Response.PSObject.Properties.Name -contains "data") {
    return Get-ResponseGuid $Response.data
  }
  return ""
}

function New-CategoryNode {
  param(
    [string]$Token,
    [int]$GroupNumber,
    [string]$Thai,
    [string]$English,
    [string]$ParentGuid = "",
    [string]$ParentGuidAll = "",
    [int]$XOrder,
    [string]$Color,
    [array]$Items = @()
  )
  $payload = New-CategoryPayload `
    -GroupNumber $GroupNumber `
    -Thai $Thai `
    -English $English `
    -ParentGuid $ParentGuid `
    -ParentGuidAll $ParentGuidAll `
    -XOrder $XOrder `
    -Color $Color `
    -Items $Items
  $response = Invoke-JsonApi -Method "POST" -Path "/product/category" -Payload $payload -Token $Token
  $script:categoriesSubmitted++
  $guid = Get-ResponseGuid $response
  if ([string]::IsNullOrWhiteSpace($guid)) {
    throw "Category create did not return guid for $Thai"
  }
  return $guid
}

function Join-ParentGuidAll {
  param([string]$ParentGuidAll, [string]$ParentGuid)
  if ([string]::IsNullOrWhiteSpace($ParentGuid)) { return "" }
  if ([string]::IsNullOrWhiteSpace($ParentGuidAll)) { return $ParentGuid }
  return "$ParentGuidAll,$ParentGuid"
}

$sections = @(
  @{ key = "SALAD"; th = "ส้มตำและตำ"; en = "Papaya salads"; color = "#f97316"; items = @(
      @("ตำโซลาว","Solao Papaya Salad",89), @("ตำไทย","Thai Papaya Salad",79), @("ตำลาว","Lao Papaya Salad",79), @("ตำปูปลาร้า","Papaya Salad with Crab and Fermented Fish",85),
      @("ตำปูม้าดอง","Papaya Salad with Pickled Blue Crab",169), @("ตำปูม้าล้วน","Blue Crab Salad",219), @("ตำหอยดอง","Papaya Salad with Pickled Mussel",95), @("ตำแตง","Cucumber Salad",75),
      @("ตำข้าวโพดไข่เค็ม","Corn Salad with Salted Egg",95), @("ตำถาดไทย","Thai Papaya Salad Tray",199), @("ตำถาดลาว","Lao Papaya Salad Tray",199), @("ตำมั่วโซลาว","Solao Mixed Papaya Salad",109)
    ) },
  @{ key = "LARB"; th = "ลาบ น้ำตก และก้อย"; en = "Larb and spicy meat salads"; color = "#dc2626"; items = @(
      @("ลาบหมู","Pork Larb",99), @("ลาบไก่","Chicken Larb",95), @("ลาบเป็ด","Duck Larb",129), @("ลาบคั่วเหนือ","Northern Spiced Larb",129),
      @("น้ำตกหมู","Pork Nam Tok",109), @("น้ำตกเนื้อ","Beef Nam Tok",139), @("ก้อยเนื้อ","Raw Spicy Beef Salad",139), @("ตับหวาน","Spicy Pork Liver Salad",99),
      @("ซุปหน่อไม้","Bamboo Shoot Salad",79), @("ลาบปลา","Fish Larb",129)
    ) },
  @{ key = "YUM"; th = "ยำแซ่บ"; en = "Spicy yum salads"; color = "#e11d48"; items = @(
      @("ยำวุ้นเส้นหมูสับ","Glass Noodle Salad with Pork",99), @("ยำมาม่าทะเล","Instant Noodle Seafood Salad",129), @("ยำทะเลรวม","Mixed Seafood Salad",159), @("ยำหมูยอ","Vietnamese Sausage Salad",99),
      @("ยำหอยแครง","Cockle Salad",169), @("ยำกุ้งสด","Raw Shrimp Salad",159), @("ยำปูม้า","Blue Crab Salad",219), @("ยำข้าวโพดไข่เค็ม","Corn Salted Egg Salad",109),
      @("กุ้งแช่น้ำปลา","Shrimp in Fish Sauce",159), @("พล่ากุ้ง","Spicy Lemongrass Shrimp Salad",159)
    ) },
  @{ key = "GRILL"; th = "ย่างและทอด"; en = "Grilled and fried"; color = "#a16207"; items = @(
      @("คอหมูย่าง","Grilled Pork Neck",129), @("ไก่ย่างโซลาว","Solao Grilled Chicken",129), @("เสือร้องไห้","Grilled Beef Brisket",169), @("ปีกไก่ทอดน้ำปลา","Fish Sauce Fried Chicken Wings",119),
      @("หมูแดดเดียว","Sun-Dried Pork",119), @("เนื้อแดดเดียว","Sun-Dried Beef",149), @("แหนมซี่โครงทอด","Fried Fermented Pork Ribs",139), @("ไส้กรอกอีสาน","Isan Sausage",109),
      @("ปลาทอดสมุนไพร","Herb Fried Fish",189), @("ปลากระพงทอดน้ำปลา","Fried Sea Bass with Fish Sauce",359)
    ) },
  @{ key = "SOUP"; th = "ต้ม แกง และอ่อม"; en = "Soups and curries"; color = "#16a34a"; items = @(
      @("ต้มแซ่บกระดูกอ่อน","Spicy Pork Rib Soup",129), @("ต้มยำกุ้งน้ำข้น","Creamy Tom Yum Shrimp",159), @("แกงอ่อมหมู","Isan Pork Herb Soup",119), @("แกงอ่อมไก่","Isan Chicken Herb Soup",109),
      @("ต้มข่าไก่","Chicken Coconut Soup",129), @("แกงเห็ดรวม","Mixed Mushroom Soup",99), @("ต้มยำปลาคัง","Tom Yum Fish",179), @("ซุปเปอร์ตีนไก่","Spicy Chicken Feet Soup",129),
      @("แกงหน่อไม้","Bamboo Shoot Curry",99), @("ต้มโคล้งปลาย่าง","Sour Smoked Fish Soup",149)
    ) },
  @{ key = "NORTHERN"; th = "อาหารเมืองเหนือ"; en = "Northern Thai dishes"; color = "#7c3aed"; items = @(
      @("ข้าวซอยไก่","Chicken Khao Soi",89), @("ขนมจีนน้ำเงี้ยว","Northern Tomato Pork Noodle",85), @("แกงฮังเล","Northern Pork Curry",139), @("น้ำพริกหนุ่มชุดผัก","Green Chili Dip Set",119),
      @("น้ำพริกอ่องชุดผัก","Tomato Pork Chili Dip Set",119), @("แคบหมู","Pork Crackling",59), @("ไส้อั่วสมุนไพร","Northern Herb Sausage",129), @("จิ๊นทอด","Northern Fried Pork",129),
      @("ผักเชียงดาผัดไข่","Chiang Da Vegetable with Egg",99), @("แกงโฮะ","Northern Mixed Curry",119)
    ) },
  @{ key = "RICE"; th = "ข้าวและเส้น"; en = "Rice and noodles"; color = "#0891b2"; items = @(
      @("ข้าวเหนียว","Sticky Rice",20), @("ข้าวสวย","Steamed Rice",20), @("ข้าวผัดหมู","Pork Fried Rice",79), @("ข้าวผัดกุ้ง","Shrimp Fried Rice",99),
      @("ผัดไทยกุ้งสด","Pad Thai Shrimp",109), @("หมี่โคราช","Korat Stir-Fried Noodles",95), @("ผัดซีอิ๊วหมู","Pork Soy Sauce Noodles",79), @("ราดหน้าหมู","Pork Gravy Noodles",79),
      @("ก๋วยเตี๋ยวคั่วไก่","Chicken Fried Noodles",89), @("ข้าวกะเพราหมู","Basil Pork Rice",79)
    ) },
  @{ key = "SEAFOOD"; th = "ทะเล"; en = "Seafood"; color = "#2563eb"; items = @(
      @("หอยแครงลวก","Blanched Cockles",169), @("หอยแมลงภู่อบสมุนไพร","Steamed Mussels with Herbs",149), @("กุ้งเผา","Grilled Prawns",259), @("ปลาหมึกย่าง","Grilled Squid",189),
      @("ปูม้าดองน้ำปลา","Pickled Blue Crab",229), @("กุ้งอบวุ้นเส้น","Baked Prawns with Glass Noodles",189), @("ทะเลลวกจิ้ม","Blanched Seafood Platter",219), @("ปลานึ่งมะนาว","Steamed Fish with Lime",299),
      @("ปลากระพงนึ่งซีอิ๊ว","Steamed Sea Bass with Soy Sauce",329), @("กุ้งทอดกระเทียม","Garlic Fried Shrimp",189)
    ) },
  @{ key = "THAI"; th = "กับข้าวไทย"; en = "Thai main dishes"; color = "#0f766e"; items = @(
      @("ไข่เจียวหมูสับ","Pork Omelet",69), @("คะน้าหมูกรอบ","Chinese Kale with Crispy Pork",119), @("ผัดผักรวม","Stir-Fried Mixed Vegetables",89), @("ผัดผักบุ้งไฟแดง","Stir-Fried Morning Glory",79),
      @("หมูกรอบผัดพริกเกลือ","Crispy Pork with Chili Salt",139), @("กะเพราเนื้อ","Basil Beef",139), @("พะแนงหมู","Pork Panang Curry",129), @("แกงเขียวหวานไก่","Green Curry Chicken",119),
      @("ผัดฉ่าทะเล","Spicy Seafood Stir-Fry",159), @("ปลาหมึกผัดไข่เค็ม","Squid with Salted Egg",159)
    ) },
  @{ key = "SNACK"; th = "ของกินเล่น"; en = "Snacks"; color = "#9333ea"; items = @(
      @("เฟรนช์ฟรายส์","French Fries",69), @("นักเก็ตไก่","Chicken Nuggets",79), @("เอ็นไก่ทอด","Fried Chicken Tendons",99), @("หมูยอทอด","Fried Vietnamese Sausage",89),
      @("แหนมทอด","Fried Fermented Pork",89), @("ปอเปี๊ยะทอด","Fried Spring Rolls",79), @("เกี๊ยวซ่า","Gyoza",89), @("ถั่วทอดสมุนไพร","Herb Fried Peanuts",59),
      @("ข้าวเกรียบ","Prawn Crackers",59), @("ไข่เจียวสมุนไพร","Herb Omelet",79)
    ) },
  @{ key = "DRINK"; th = "เครื่องดื่ม"; en = "Drinks"; color = "#0284c7"; items = @(
      @("น้ำเปล่า","Water",15), @("โค้ก","Coke",25), @("สไปรท์","Sprite",25), @("น้ำแข็งถังเล็ก","Small Ice Bucket",20),
      @("ชาไทยเย็น","Thai Iced Tea",55), @("ชาเขียวเย็น","Iced Green Tea",55), @("กาแฟเย็น","Iced Coffee",55), @("น้ำมะนาว","Lime Juice",55),
      @("น้ำผึ้งมะนาว","Honey Lime Juice",65), @("น้ำอัญชันมะนาว","Butterfly Pea Lime",60), @("น้ำลำไย","Longan Drink",45), @("น้ำกระเจี๊ยบ","Roselle Drink",45)
    ) },
  @{ key = "DESSERT"; th = "ของหวาน"; en = "Desserts"; color = "#be185d"; items = @(
      @("เฉาก๊วย","Grass Jelly",45), @("ไอศกรีมกะทิ","Coconut Ice Cream",55), @("บัวลอย","Bua Loy",55), @("ข้าวเหนียวมะม่วง","Mango Sticky Rice",89),
      @("ลอดช่อง","Lod Chong",45), @("รวมมิตร","Thai Ruam Mit Dessert",55), @("ทับทิมกรอบ","Tub Tim Grob",55), @("กล้วยบวชชี","Banana in Coconut Milk",55)
    ) }
)

$token = Get-SelectedToken -HoldingCode $HoldingCode -Username $Username

$categoryRows = @()
$categoryRows += Search-Categories -Token $token -Query "ตัวอย่างหมวดสินค้า"
$categoryRows += Search-Categories -Token $token -Query "โซลาวตัวอย่าง"
$categoryRows = Get-UniqueRows $categoryRows | Where-Object {
  $n = Get-DocThaiName $_
  $n.StartsWith("ตัวอย่างหมวดสินค้า") -or $n.StartsWith("โซลาวตัวอย่าง")
}
$categoryDelete = Remove-GuidBatch -Token $token -Path "/product/category" -Guids @($categoryRows | ForEach-Object { Get-DocGuid $_ })

$barcodeRows = @()
foreach ($query in @("SOLAO-", "885660604", "MENU-", "885260529", "SET-000")) {
  $barcodeRows += Search-Barcodes -Token $token -Query $query
}
$barcodeRows = Get-UniqueRows $barcodeRows | Where-Object {
  $itemcode = [string]$_.itemcode
  $barcode = [string]$_.barcode
  $itemcode.StartsWith("SOLAO-") -or
  $itemcode.StartsWith("MENU-") -or
  ($itemcode -in @("SET-0001","SET-0002","SET-0003","SET-0004","SET-0005")) -or
  $barcode.StartsWith("885660604") -or
  $barcode.StartsWith("885260529")
}
$setBarcodeRows = @($barcodeRows | Where-Object { [int]$_.item_type -eq 2 -or [string]$_.itemcode -like "SOLAO-SET-*" -or [string]$_.itemcode -in @("SET-0001","SET-0002","SET-0003","SET-0004","SET-0005") })
$normalBarcodeRows = @($barcodeRows | Where-Object {
  $guid = Get-DocGuid $_
  -not (@($setBarcodeRows | ForEach-Object { Get-DocGuid $_ }) -contains $guid)
})
$setBarcodeDelete = Remove-GuidBatch -Token $token -Path "/product/barcode" -Guids @($setBarcodeRows | ForEach-Object { Get-DocGuid $_ })
$normalBarcodeDelete = Remove-GuidBatch -Token $token -Path "/product/barcode" -Guids @($normalBarcodeRows | ForEach-Object { Get-DocGuid $_ })

$productRows = @()
foreach ($query in @("SOLAO-", "MENU-", "SET-000")) {
  $productRows += Search-Products -Token $token -Query $query
}
$productRows = Get-UniqueRows $productRows | Where-Object {
  $code = [string]$_.code
  $code.StartsWith("SOLAO-") -or
  $code.StartsWith("MENU-") -or
  ($code -in @("SET-0001","SET-0002","SET-0003","SET-0004","SET-0005"))
}
$productDelete = Remove-ProductRows -Token $token -Rows $productRows

$menuItems = @()
$sequence = 1
$round = 0
while ($menuItems.Count -lt $ProductCount) {
  foreach ($section in $sections) {
    foreach ($raw in $section.items) {
      if ($menuItems.Count -ge $ProductCount) { break }
      $variantTh = ""
      $variantEn = ""
      $priceAdd = 0
      if ($round -eq 1) {
        $variantTh = " จานใหญ่"
        $variantEn = " Large"
        $priceAdd = 40
      }
      elseif ($round -ge 2) {
        $variantTh = " ชุดแชร์"
        $variantEn = " Sharing Set"
        $priceAdd = 90
      }
      $th = [string]$raw[0] + $variantTh
      $en = [string]$raw[1] + $variantEn
      $unit = Resolve-Unit $th
      $menuItems += [pscustomobject]@{
        section = $section.key
        section_th = $section.th
        section_en = $section.en
        color = $section.color
        code = "SOLAO-{0:D4}" -f $sequence
        barcode = "885660604{0:D4}" -f $sequence
        th = $th
        en = $en
        price = [decimal]$raw[2] + $priceAdd
        unit = $unit
        item_type = 0
        materialtype = 0
        bom = @()
      }
      $sequence++
    }
    if ($menuItems.Count -ge $ProductCount) { break }
  }
  $round++
}

$setDefs = @(
  @{ th = "ชุดตำโซลาวไก่ย่าง"; en = "Solao Papaya Salad and Grilled Chicken Set"; price = 299; refs = @("8856606040001","8856606040042","8856606040069") },
  @{ th = "ชุดปูม้าดองแซ่บ"; en = "Pickled Blue Crab Spicy Set"; price = 399; refs = @("8856606040005","8856606040027","8856606040085") },
  @{ th = "ชุดลาบน้ำตกข้าวเหนียว"; en = "Larb Nam Tok Sticky Rice Set"; price = 259; refs = @("8856606040013","8856606040017","8856606040067") },
  @{ th = "ชุดต้มแซ่บครอบครัว"; en = "Family Spicy Soup Set"; price = 349; refs = @("8856606040041","8856606040042","8856606040071") },
  @{ th = "ชุดเมืองเหนือ"; en = "Northern Thai Set"; price = 329; refs = @("8856606040051","8856606040053","8856606040054") },
  @{ th = "ชุดทะเลรวมโซลาว"; en = "Solao Mixed Seafood Set"; price = 459; refs = @("8856606040071","8856606040073","8856606040077") },
  @{ th = "ชุดยำแซ่บเพื่อนเยอะ"; en = "Spicy Yum Party Set"; price = 379; refs = @("8856606040023","8856606040026","8856606040028") },
  @{ th = "ชุดของหวานและเครื่องดื่ม"; en = "Dessert and Drink Set"; price = 199; refs = @("8856606040115","8856606040121","8856606040124") }
)

$setItems = @()
$setSeq = 1
foreach ($set in $setDefs) {
  $bom = @()
  foreach ($ref in $set.refs) {
    $bom += @{ barcode = $ref; condition = $false; dividevalue = 1; standvalue = 1; qty = 1 }
  }
  $setItems += [pscustomobject]@{
    section = "SET"
    section_th = "ชุดเมนูแนะนำ"
    section_en = "Recommended sets"
    color = "#0e7490"
    code = "SOLAO-SET-{0:D3}" -f $setSeq
    barcode = "8856606049{0:D3}" -f $setSeq
    th = $set.th
    en = $set.en
    price = [decimal]$set.price
    unit = @{ code = "SET"; th = "ชุด"; en = "Set" }
    item_type = 2
    materialtype = 3
    bom = @($bom)
  }
  $setSeq++
}

$allItems = @($menuItems + $setItems)

$existingProductRows = Search-Products -Token $token -Query "SOLAO-" -Limit 1000
$existingProductCodes = @{}
foreach ($row in @($existingProductRows)) { if ($row.code) { $existingProductCodes[[string]$row.code] = $true } }

$productsCreated = 0
$productsSkipped = 0
foreach ($item in $allItems) {
  if ($existingProductCodes.ContainsKey($item.code)) {
    $productsSkipped++
    continue
  }
  [void](Invoke-JsonApi -Method "POST" -Path "/product" -Payload (New-ProductPayload $item) -Token $token)
  $productsCreated++
}

$existingBarcodeRows = @()
$existingBarcodeRows += Search-Barcodes -Token $token -Query "SOLAO-" -Limit 1000
$existingBarcodeRows += Search-Barcodes -Token $token -Query "885660604" -Limit 1000
$existingBarcodes = @{}
foreach ($row in @(Get-UniqueRows $existingBarcodeRows)) { if ($row.barcode) { $existingBarcodes[[string]$row.barcode] = $true } }

$barcodePayloads = @()
foreach ($item in $allItems) {
  if (-not $existingBarcodes.ContainsKey($item.barcode)) {
    $barcodePayloads += New-BarcodePayload $item
  }
}

$barcodesSubmitted = 0
for ($offset = 0; $offset -lt $barcodePayloads.Count; $offset += 40) {
  $end = [math]::Min($offset + 39, $barcodePayloads.Count - 1)
  $chunk = @($barcodePayloads[$offset..$end])
  [void](Invoke-JsonApi -Method "POST" -Path "/product/barcode/bulk" -Payload $chunk -Token $token)
  $barcodesSubmitted += $chunk.Count
}

Start-Sleep -Seconds 2

function Select-MenuItems {
  param(
    [string[]]$SectionKeys = @(),
    [int]$Limit = 0,
    [scriptblock]$Filter = $null
  )
  $rows = @($menuItems)
  if ($SectionKeys.Count -gt 0) {
    $rows = @($rows | Where-Object { $SectionKeys -contains $_.section })
  }
  if ($null -ne $Filter) {
    $rows = @($rows | Where-Object $Filter)
  }
  if ($Limit -gt 0) {
    return @($rows | Select-Object -First $Limit)
  }
  return @($rows)
}

$script:categoriesSubmitted = 0

$groupTrees = @(
  @{
    number = 1
    th = "โซลาวตัวอย่าง - สั่งอาหารหน้าร้าน"
    en = "Solao sample - Dine-in ordering"
    color = "#047857"
    branches = @(
      @{ th = "เมนูแนะนำ"; en = "Recommended"; color = "#0369a1"; leaves = @(
          @{ th = "ขายดีหน้าร้าน"; en = "Best sellers"; color = "#0369a1"; items = @(Select-MenuItems -Limit 18) },
          @{ th = "ชุดแชร์"; en = "Sharing sets"; color = "#0e7490"; items = @($setItems) }
        ) },
      @{ th = "อาหารอีสาน"; en = "Isan dishes"; color = "#dc2626"; leaves = @(
          @{ th = "ส้มตำและตำ"; en = "Papaya salads"; color = "#f97316"; items = @(Select-MenuItems -SectionKeys @("SALAD") -Limit 24) },
          @{ th = "ลาบ น้ำตก และก้อย"; en = "Larb and spicy meat salads"; color = "#dc2626"; items = @(Select-MenuItems -SectionKeys @("LARB") -Limit 24) },
          @{ th = "ยำแซ่บ"; en = "Spicy yum salads"; color = "#e11d48"; items = @(Select-MenuItems -SectionKeys @("YUM") -Limit 24) }
        ) },
      @{ th = "ครัวร้อน"; en = "Hot kitchen"; color = "#16a34a"; leaves = @(
          @{ th = "ต้ม แกง และอ่อม"; en = "Soups and curries"; color = "#16a34a"; items = @(Select-MenuItems -SectionKeys @("SOUP") -Limit 24) },
          @{ th = "ย่างและทอด"; en = "Grilled and fried"; color = "#a16207"; items = @(Select-MenuItems -SectionKeys @("GRILL") -Limit 24) },
          @{ th = "อาหารเมืองเหนือ"; en = "Northern Thai dishes"; color = "#7c3aed"; items = @(Select-MenuItems -SectionKeys @("NORTHERN") -Limit 24) }
        ) },
      @{ th = "เครื่องดื่มและของหวาน"; en = "Drinks and desserts"; color = "#0284c7"; leaves = @(
          @{ th = "เครื่องดื่ม"; en = "Drinks"; color = "#0284c7"; items = @(Select-MenuItems -SectionKeys @("DRINK") -Limit 24) },
          @{ th = "ของหวาน"; en = "Desserts"; color = "#be185d"; items = @(Select-MenuItems -SectionKeys @("DESSERT") -Limit 24) }
        ) }
    )
  },
  @{
    number = 2
    th = "โซลาวตัวอย่าง - แคชเชียร์ / POS"
    en = "Solao sample - Cashier / POS"
    color = "#0f766e"
    branches = @(
      @{ th = "ขายเร็ว"; en = "Quick sale"; color = "#0f766e"; leaves = @(
          @{ th = "เมนูขายดี"; en = "Fast best sellers"; color = "#0f766e"; items = @(Select-MenuItems -Limit 22) },
          @{ th = "เครื่องดื่มขายเร็ว"; en = "Quick drinks"; color = "#0284c7"; items = @(Select-MenuItems -SectionKeys @("DRINK") -Limit 18) },
          @{ th = "ของทานเล่น"; en = "Quick snacks"; color = "#9333ea"; items = @(Select-MenuItems -SectionKeys @("SNACK") -Limit 18) }
        ) },
      @{ th = "ชำระเงินและแพ็กกลับบ้าน"; en = "Checkout and takeaway"; color = "#ca8a04"; leaves = @(
          @{ th = "ชุดกลับบ้าน"; en = "Takeaway sets"; color = "#0e7490"; items = @($setItems) },
          @{ th = "รายการราคาไม่สูง"; en = "Low-price items"; color = "#ca8a04"; items = @(Select-MenuItems -Limit 24 -Filter { $_.price -le 89 }) }
        ) }
    )
  },
  @{
    number = 3
    th = "โซลาวตัวอย่าง - ครัวและ KDS"
    en = "Solao sample - Kitchen and KDS"
    color = "#1d4ed8"
    branches = @(
      @{ th = "สถานีส้มตำและยำ"; en = "Papaya salad and yum station"; color = "#e11d48"; leaves = @(
          @{ th = "ส้มตำ"; en = "Papaya salad"; color = "#f97316"; items = @(Select-MenuItems -SectionKeys @("SALAD") -Limit 30) },
          @{ th = "ยำ"; en = "Yum salads"; color = "#e11d48"; items = @(Select-MenuItems -SectionKeys @("YUM") -Limit 24) }
        ) },
      @{ th = "สถานีครัวร้อน"; en = "Hot kitchen station"; color = "#16a34a"; leaves = @(
          @{ th = "ต้มและแกง"; en = "Soups and curries"; color = "#16a34a"; items = @(Select-MenuItems -SectionKeys @("SOUP") -Limit 24) },
          @{ th = "กับข้าวไทย"; en = "Thai main dishes"; color = "#0f766e"; items = @(Select-MenuItems -SectionKeys @("THAI") -Limit 24) },
          @{ th = "อาหารเมืองเหนือ"; en = "Northern Thai dishes"; color = "#7c3aed"; items = @(Select-MenuItems -SectionKeys @("NORTHERN") -Limit 24) }
        ) },
      @{ th = "สถานีย่างทอด"; en = "Grill and fry station"; color = "#a16207"; leaves = @(
          @{ th = "ย่างทอด"; en = "Grill and fry"; color = "#a16207"; items = @(Select-MenuItems -SectionKeys @("GRILL") -Limit 24) },
          @{ th = "ของกินเล่น"; en = "Snacks"; color = "#9333ea"; items = @(Select-MenuItems -SectionKeys @("SNACK") -Limit 24) }
        ) },
      @{ th = "สถานีเครื่องดื่ม"; en = "Beverage station"; color = "#0284c7"; leaves = @(
          @{ th = "เครื่องดื่ม"; en = "Drinks"; color = "#0284c7"; items = @(Select-MenuItems -SectionKeys @("DRINK") -Limit 24) },
          @{ th = "ของหวาน"; en = "Desserts"; color = "#be185d"; items = @(Select-MenuItems -SectionKeys @("DESSERT") -Limit 18) }
        ) }
    )
  },
  @{
    number = 4
    th = "โซลาวตัวอย่าง - เดลิเวอรี่"
    en = "Solao sample - Delivery"
    color = "#7c2d12"
    branches = @(
      @{ th = "เมนูพร้อมส่ง"; en = "Delivery-ready menu"; color = "#7c2d12"; leaves = @(
          @{ th = "กล่องข้าวและเส้น"; en = "Rice and noodle boxes"; color = "#0891b2"; items = @(Select-MenuItems -SectionKeys @("RICE") -Limit 24) },
          @{ th = "ย่างทอดพร้อมส่ง"; en = "Grilled and fried delivery"; color = "#a16207"; items = @(Select-MenuItems -SectionKeys @("GRILL","SNACK") -Limit 28) }
        ) },
      @{ th = "ชุดเดลิเวอรี่"; en = "Delivery sets"; color = "#0e7490"; leaves = @(
          @{ th = "ชุดแนะนำ"; en = "Recommended sets"; color = "#0e7490"; items = @($setItems) }
        ) },
      @{ th = "เครื่องดื่มเดลิเวอรี่"; en = "Delivery drinks"; color = "#0284c7"; leaves = @(
          @{ th = "เครื่องดื่ม"; en = "Drinks"; color = "#0284c7"; items = @(Select-MenuItems -SectionKeys @("DRINK") -Limit 24) }
        ) }
    )
  },
  @{
    number = 5
    th = "โซลาวตัวอย่าง - งานเลี้ยงและห้องจัดเลี้ยง"
    en = "Solao sample - Party and private room"
    color = "#6d28d9"
    branches = @(
      @{ th = "ชุดโต๊ะใหญ่"; en = "Large table sets"; color = "#6d28d9"; leaves = @(
          @{ th = "ชุดแชร์"; en = "Sharing sets"; color = "#0e7490"; items = @($setItems) },
          @{ th = "เมนูครอบครัว"; en = "Family dishes"; color = "#15803d"; items = @(Select-MenuItems -Limit 28 -Filter { $_.price -ge 119 }) }
        ) },
      @{ th = "ทะเลและกับข้าว"; en = "Seafood and main dishes"; color = "#2563eb"; leaves = @(
          @{ th = "ทะเล"; en = "Seafood"; color = "#2563eb"; items = @(Select-MenuItems -SectionKeys @("SEAFOOD") -Limit 24) },
          @{ th = "กับข้าวไทย"; en = "Thai main dishes"; color = "#0f766e"; items = @(Select-MenuItems -SectionKeys @("THAI") -Limit 24) }
        ) },
      @{ th = "ของหวานปิดท้าย"; en = "Dessert closing"; color = "#be185d"; leaves = @(
          @{ th = "ของหวาน"; en = "Desserts"; color = "#be185d"; items = @(Select-MenuItems -SectionKeys @("DESSERT") -Limit 18) }
        ) }
    )
  },
  @{
    number = 6
    th = "โซลาวตัวอย่าง - ผู้จัดการร้าน"
    en = "Solao sample - Store manager"
    color = "#334155"
    branches = @(
      @{ th = "วิเคราะห์การขาย"; en = "Sales analysis"; color = "#334155"; leaves = @(
          @{ th = "เมนูกำไรดี"; en = "High margin menu"; color = "#15803d"; items = @(Select-MenuItems -Limit 26 -Filter { $_.price -ge 129 }) },
          @{ th = "เมนูต้องติดตาม"; en = "Watch list menu"; color = "#ca8a04"; items = @(Select-MenuItems -SectionKeys @("SEAFOOD","THAI") -Limit 24) },
          @{ th = "ควบคุมต้นทุน"; en = "Cost control"; color = "#b91c1c"; items = @(Select-MenuItems -SectionKeys @("GRILL","SEAFOOD") -Limit 24) }
        ) }
    )
  },
  @{
    number = 7
    th = "โซลาวตัวอย่าง - เมนูตามเวลา"
    en = "Solao sample - Time-based menu"
    color = "#475569"
    branches = @(
      @{ th = "ช่วงเวลาให้บริการ"; en = "Service time"; color = "#475569"; leaves = @(
          @{ th = "กลางวัน"; en = "Lunch"; color = "#ca8a04"; items = @(Select-MenuItems -SectionKeys @("RICE","THAI","DRINK") -Limit 28) },
          @{ th = "เย็น"; en = "Dinner"; color = "#7c2d12"; items = @(Select-MenuItems -SectionKeys @("SEAFOOD","GRILL","SOUP","SET") -Limit 28) },
          @{ th = "ทั้งวัน"; en = "All day"; color = "#047857"; items = @(Select-MenuItems -Limit 30) }
        ) }
    )
  },
  @{
    number = 8
    th = "โซลาวตัวอย่าง - โปรโมชั่น"
    en = "Solao sample - Promotions"
    color = "#b91c1c"
    branches = @(
      @{ th = "แคมเปญขาย"; en = "Sales campaigns"; color = "#b91c1c"; leaves = @(
          @{ th = "ขายดี"; en = "Best sellers"; color = "#0369a1"; items = @(Select-MenuItems -Limit 20) },
          @{ th = "ครอบครัว"; en = "Family"; color = "#15803d"; items = @(Select-MenuItems -Limit 24 -Filter { $_.price -ge 119 }) },
          @{ th = "เผ็ดนัว"; en = "Spicy signatures"; color = "#b91c1c"; items = @(Select-MenuItems -Limit 24 -Filter { $_.th -match "ตำ|ยำ|ลาบ|น้ำตก|แซ่บ|ก้อย" }) },
          @{ th = "เด็กและไม่เผ็ด"; en = "Mild and kids menu"; color = "#ca8a04"; items = @(Select-MenuItems -Limit 24 -Filter { $_.th -match "ข้าว|ไข่|ทอด|น้ำ|ของหวาน|ไอศกรีม" }) }
        ) }
    )
  }
)

foreach ($group in $groupTrees) {
  $rootGuid = New-CategoryNode `
    -Token $token `
    -GroupNumber ([int]$group.number) `
    -Thai $group.th `
    -English $group.en `
    -XOrder ([int]$group.number) `
    -Color $group.color
  $branchXOrder = 1
  foreach ($branch in @($group.branches)) {
    $branchGuid = New-CategoryNode `
      -Token $token `
      -GroupNumber ([int]$group.number) `
      -Thai $branch.th `
      -English $branch.en `
      -ParentGuid $rootGuid `
      -ParentGuidAll $rootGuid `
      -XOrder $branchXOrder `
      -Color $branch.color
    $leafParentGuidAll = Join-ParentGuidAll -ParentGuidAll $rootGuid -ParentGuid $branchGuid
    $leafXOrder = 1
    foreach ($leaf in @($branch.leaves)) {
      [void](New-CategoryNode `
        -Token $token `
        -GroupNumber ([int]$group.number) `
        -Thai $leaf.th `
        -English $leaf.en `
        -ParentGuid $branchGuid `
        -ParentGuidAll $leafParentGuidAll `
        -XOrder $leafXOrder `
        -Color $leaf.color `
        -Items @($leaf.items))
      $leafXOrder++
    }
    $branchXOrder++
  }
}

$categoriesSubmitted = $script:categoriesSubmitted

Start-Sleep -Seconds 2

$verifyProducts = Search-Products -Token $token -Query "SOLAO-" -Limit 1000
$verifyBarcodes = Search-Barcodes -Token $token -Query "SOLAO-" -Limit 1000
$verifyCategoryRows = @()
for ($gn = 1; $gn -le 8; $gn++) {
  $groupResponse = Invoke-GetJson -Path "/product/category/list?offset=0&limit=1000&group-number=$gn&lang=th" -Token $token
  $verifyCategoryRows += @($groupResponse.data)
}
$verifyGroup1 = Invoke-GetJson -Path "/product/category/list?offset=0&limit=1000&group-number=1&lang=th" -Token $token
$verifyGroup1Rows = @($verifyGroup1.data)
$verifyGroup1Root = @($verifyGroup1Rows | Where-Object { [string]::IsNullOrWhiteSpace([string]$_.parent_guid) -and [string]::IsNullOrWhiteSpace([string]$_.parentguid) })
$verifyGroup1Leaves = @($verifyGroup1Rows | Where-Object { (Get-CodeListCount $_) -gt 0 })
$verifyMenuCategory = @($verifyGroup1Rows | Where-Object { (Get-DocThaiName $_) -eq "ส้มตำและตำ" } | Select-Object -First 1)
$verifyGroup1MaxDepth = 0
foreach ($row in $verifyGroup1Rows) {
  $parentPath = [string]$row.parentguidall
  $depth = if ([string]::IsNullOrWhiteSpace($parentPath)) { 1 } else { ($parentPath -split ",").Count + 1 }
  if ($depth -gt $verifyGroup1MaxDepth) { $verifyGroup1MaxDepth = $depth }
}

[pscustomobject]@{
  holdingcode = $HoldingCode
  cleanup = @{
    categories_deleted = $categoryDelete.deleted
    category_delete_errors = @($categoryDelete.errors)
    set_barcodes_deleted = $setBarcodeDelete.deleted
    set_barcode_delete_errors = @($setBarcodeDelete.errors)
    normal_barcodes_deleted = $normalBarcodeDelete.deleted
    normal_barcode_delete_errors = @($normalBarcodeDelete.errors)
    products_deleted = $productDelete.deleted
    product_delete_errors = @($productDelete.errors)
  }
  seed = @{
    products_requested = $allItems.Count
    products_created = $productsCreated
    products_skipped_existing = $productsSkipped
    barcodes_submitted = $barcodesSubmitted
    categories_submitted = $categoriesSubmitted
  }
  verify = @{
    solao_product_rows = @($verifyProducts).Count
    solao_barcode_rows = @($verifyBarcodes).Count
    solao_category_rows = @($verifyCategoryRows).Count
    solao_category_root_rows = @($verifyCategoryRows | Where-Object { [string]::IsNullOrWhiteSpace([string]$_.parent_guid) -and [string]::IsNullOrWhiteSpace([string]$_.parentguid) }).Count
    group1_root_count = @($verifyGroup1Root).Count
    group1_node_count = @($verifyGroup1Rows).Count
    group1_sellable_leaf_count = @($verifyGroup1Leaves).Count
    group1_max_depth = $verifyGroup1MaxDepth
    sample_category = if ($verifyMenuCategory) { Get-DocThaiName $verifyMenuCategory[0] } else { "" }
    sample_category_item_count = if ($verifyMenuCategory) { Get-CodeListCount $verifyMenuCategory[0] } else { 0 }
  }
} | ConvertTo-Json -Depth 12
