#Requires -Version 7.0

param(
  [string]$HoldingCode = "xxx",
  [string]$Username = "jaturapornchai@gmail.com",
  [string]$BaseUrl = "http://localhost:8888"
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
  throw "No selected auth token found for $Username / $HoldingCode"
}

function Invoke-JsonApi {
  param([string]$Method, [string]$Path, [object]$Payload, [string]$Token)
  $bodyFile = [IO.Path]::Combine($env:TEMP, "bc-variant-seed-body-$([guid]::NewGuid()).json")
  $outFile = [IO.Path]::Combine($env:TEMP, "bc-variant-seed-out-$([guid]::NewGuid()).json")
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
      $Payload | ConvertTo-Json -Depth 100 | Set-Content -LiteralPath $bodyFile -Encoding UTF8
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

function Invoke-AtlasUpsert {
  param([string]$Collection, [hashtable]$Data, [string]$Token)
  $payload = @{
    collection = $Collection
    holdingcode = $HoldingCode
    guidfixed = $Data.guidfixed
    upsert = $true
    data = $Data
  }
  return Invoke-JsonApi -Method "POST" -Path "/goapi/atlas/update" -Payload $payload -Token $Token
}

function Invoke-AtlasGetCount {
  param([string]$Collection, [string]$Token)
  $payload = @{
    collection = $Collection
    holdingcode = $HoldingCode
    limit = 2000
    skip = 0
  }
  $response = Invoke-JsonApi -Method "POST" -Path "/goapi/atlas/get" -Payload $payload -Token $Token
  return [int]($response.count ?? @($response.data).Count)
}

function New-Names {
  param([string]$Thai, [string]$English)
  return @(
    @{ code = "th"; name = $Thai; isauto = $false; isdelete = $false },
    @{ code = "en"; name = $English; isauto = $false; isdelete = $false }
  )
}

function New-MediaAsset {
  param(
    [string]$Kind,
    [string]$Uri,
    [int]$SortOrder,
    [string]$OptionCode = "",
    [string]$OptionValue = "",
    [string]$UseCase = "MAIN_IMAGE"
  )
  return @{
    kind = $Kind
    uri = $Uri
    external_id = ""
    external_url = ""
    optioncode = $OptionCode
    optionvalue = $OptionValue
    sortorder = $SortOrder
    alt_text = "$Kind $OptionValue".Trim()
    usecase = $UseCase
    mime_type = "image/webp"
    width = 1200
    height = 1200
    last_imported_at = ""
  }
}

function New-AttributeValue {
  param([string]$Code, [string]$Text, [string]$UnitCode = "")
  return @{
    value_id = ""
    valuecode = $Code
    valuetext = $Text
    display_text = $Text
    unitcode = $UnitCode
    sortorder = 0
    is_custom_value = $false
  }
}

function New-SpecAttribute {
  param(
    [string]$Code,
    [string]$Name,
    [string]$InputType,
    [string]$Scope,
    [object[]]$Values,
    [bool]$Required = $false,
    [bool]$SaleProp = $false
  )
  return @{
    attribute_id = ""
    attributecode = $Code
    attributename = $Name
    inputtype = $InputType
    scope = $Scope
    is_required = $Required
    is_sale_prop = $SaleProp
    is_custom = $false
    values = @($Values)
  }
}

function New-SpecGroup {
  param([string]$Code, [string]$Name, [object[]]$Attributes)
  return @{
    groupcode = $Code
    groupname = $Name
    attributes = @($Attributes)
  }
}

function New-PayloadExamples {
  param([string]$Family)
  return @(
    @{
      direction = "import"
      usecase = "$Family product detail"
      payload = @{
        title = "{{title}}"
        images = @("{{main_image_url}}", "{{gallery_image_url}}")
        attributes = @(@{ name = "Material"; value = "{{material}}" })
        skus = @(@{ sellersku = "{{sellersku}}"; price = "{{saleprice}}"; stock = "{{stock}}"; images = @("{{sku_image_url}}") })
      }
    },
    @{
      direction = "export"
      usecase = "$Family create/update"
      payload = @{
        category_id = "{{external_category_id}}"
        media_assets = @(@{ kind = "main"; uri = "{{main_image_uri}}" }, @{ kind = "sku"; optioncode = "COLOR"; optionvalue = "{{color}}" })
        specification_groups = @(@{ groupcode = "GENERAL"; attributes = @(@{ attributecode = "MATERIAL"; values = @("{{material}}") }) })
        sku_combinations = @(@{ sellersku = "{{sellersku}}"; optionvalues = @("{{option_1}}", "{{option_2}}"); price = "{{saleprice}}"; stock = "{{stock}}" })
      }
    }
  )
}

function Get-CompanyCodes {
  param([string]$Token)
  try {
    $result = Invoke-JsonApi -Method "GET" -Path "/organization/company" -Payload $null -Token $Token
    return @($result.data) |
      ForEach-Object { [string]($_.code ?? $_.businesscode ?? "") } |
      Where-Object { $_.Trim() -ne "" } |
      ForEach-Object { $_.Trim().ToUpperInvariant() } |
      Select-Object -Unique
  }
  catch {
    return @()
  }
}

$token = Get-SelectedToken -HoldingCode $HoldingCode -Username $Username
$companyCodes = @(Get-CompanyCodes -Token $token)
$allCompanies = @()
$companyA = if ($companyCodes.Count -ge 1) { @($companyCodes[0]) } else { @() }
$companyB = if ($companyCodes.Count -ge 2) { @($companyCodes[1]) } else { @() }
$companyC = if ($companyCodes.Count -ge 3) { @($companyCodes[2]) } else { @() }

$colors = @(
  @{ code = "BLACK"; th = "ดำ"; en = "Black"; hex = "#111827"; family = "neutral"; aliases = @("ดำ", "สีดำ", "black", "BK"); businesscodes = $allCompanies },
  @{ code = "WHITE"; th = "ขาว"; en = "White"; hex = "#F9FAFB"; family = "neutral"; aliases = @("ขาว", "สีขาว", "white", "WH"); businesscodes = $allCompanies },
  @{ code = "GREY"; th = "เทา"; en = "Grey"; hex = "#6B7280"; family = "neutral"; aliases = @("เทา", "สีเทา", "gray", "grey"); businesscodes = $allCompanies },
  @{ code = "NAVY"; th = "น้ำเงิน"; en = "Navy"; hex = "#1E3A8A"; family = "blue"; aliases = @("กรม", "น้ำเงิน", "navy", "dark blue"); businesscodes = $companyA },
  @{ code = "KHAKI"; th = "กากี"; en = "Khaki"; hex = "#A16207"; family = "brown"; aliases = @("กากี", "khaki", "tan"); businesscodes = $companyA },
  @{ code = "GREEN"; th = "เขียว"; en = "Green"; hex = "#15803D"; family = "green"; aliases = @("เขียว", "สีเขียว", "green"); businesscodes = $companyB },
  @{ code = "CREAM"; th = "ครีม"; en = "Cream"; hex = "#F5E8C7"; family = "yellow"; aliases = @("ครีม", "cream", "ivory"); businesscodes = $companyB },
  @{ code = "RED"; th = "แดง"; en = "Red"; hex = "#DC2626"; family = "red"; aliases = @("แดง", "red", "scarlet"); businesscodes = $companyC }
)

$sizes = @(
  @{ code = "XS"; th = "XS"; en = "XS"; system = "intl"; type = "regular"; order = 10; aliases = @("XS", "Extra Small"); businesscodes = $allCompanies },
  @{ code = "S"; th = "S (29-32)"; en = "S"; system = "intl"; type = "regular"; order = 20; aliases = @("S", "S(29-32)", "S 29-32"); businesscodes = $allCompanies },
  @{ code = "M"; th = "M (32-34)"; en = "M"; system = "intl"; type = "regular"; order = 30; aliases = @("M", "M(40-50kg)", "M 32-34"); businesscodes = $allCompanies },
  @{ code = "L"; th = "L (34-36)"; en = "L"; system = "intl"; type = "regular"; order = 40; aliases = @("L", "L(50-60kg)", "L 34-36"); businesscodes = $allCompanies },
  @{ code = "XL"; th = "XL (36-37)"; en = "XL"; system = "intl"; type = "regular"; order = 50; aliases = @("XL", "XL(60-70kg)", "XL 36-37"); businesscodes = $companyA },
  @{ code = "2XL"; th = "2XL (37-38)"; en = "2XL"; system = "intl"; type = "plus"; order = 60; aliases = @("XXL", "2XL", "2XL(70-80kg)"); businesscodes = $companyA },
  @{ code = "128GB"; th = "128GB"; en = "128GB"; system = "intl"; type = "regular"; order = 100; aliases = @("128G", "128 GB", "128GB"); businesscodes = $companyB },
  @{ code = "256GB"; th = "256GB"; en = "256GB"; system = "intl"; type = "regular"; order = 110; aliases = @("256G", "256 GB", "256GB"); businesscodes = $companyB },
  @{ code = "512GB"; th = "512GB"; en = "512GB"; system = "intl"; type = "regular"; order = 120; aliases = @("512G", "512 GB", "512GB"); businesscodes = $companyB },
  @{ code = "ESIM"; th = "eSIM"; en = "eSIM"; system = "intl"; type = "regular"; order = 130; aliases = @("ESIM", "e-sim", "embedded sim"); businesscodes = $companyC },
  @{ code = "NANO_SIM"; th = "Nano SIM"; en = "Nano SIM"; system = "intl"; type = "regular"; order = 140; aliases = @("nano", "nano sim", "ซิมนาโน"); businesscodes = $companyC }
)

function New-ExternalIntegrationProfiles {
  return @(
    @{
      channel = "shopee"
      payload_shape = @{
        product_keys = @("item_id", "category_id", "brand", "image", "description_info", "attribute_list", "tier_variation", "model")
        media_keys = @("image_id_list", "video_upload_id", "description_images", "tier_variation.option_list.image_id")
        specification_keys = @("attribute_list", "tier_variation", "model.tier_index")
        sku_keys = @("model_id", "model_sku", "price", "stock", "tier_index", "gtin_code")
      }
      stock_price_fields = @{
        sellersku = "model_sku"
        saleprice = "price"
        openingstock = "stock"
      }
    },
    @{
      channel = "lazada"
      payload_shape = @{
        product_keys = @("PrimaryCategory", "Attributes", "Skus", "Images", "description", "short_description")
        media_keys = @("Images.Image", "Skus.Sku.Images.Image", "description", "video_url")
        specification_keys = @("Attributes", "SaleProp", "GetCategoryAttributes")
        sku_keys = @("SellerSku", "ShopSku", "SaleProp", "price", "quantity", "packageweight", "packagelength", "packagewidth", "packageheight")
      }
      stock_price_fields = @{
        sellersku = "SellerSku"
        saleprice = "price"
        openingstock = "quantity"
      }
    },
    @{
      channel = "tiktok"
      payload_shape = @{
        product_keys = @("category_id", "title", "description", "main_images", "product_attributes", "sales_attributes", "skus")
        media_keys = @("main_images", "skus.sales_attributes.sku_img", "size_chart.image", "certifications.files")
        specification_keys = @("product_attributes", "sales_attributes", "custom_attributes")
        sku_keys = @("sellersku", "price", "inventory", "sales_attributes", "identifier_code", "packageweight", "package_dimensions")
      }
      stock_price_fields = @{
        sellersku = "sellersku"
        saleprice = "price"
        openingstock = "inventory"
      }
    },
    @{
      channel = "aliexpress"
      payload_shape = @{
        solution_api = @{
          product_keys = @("subject_list", "category_id", "brand_name", "main_image_urls_list", "attribute_list", "sku_info_list", "product_unit", "currency_code", "product_price", "detail_source_list")
          media_keys = @("main_image_urls_list", "detail_source_list.mobile_detail", "sku_info_list.sku_attributes_list.sku_image_url")
          specification_keys = @("attribute_list", "supporting_common_attribute_list", "supporting_sku_attribute_list")
          attribute_keys = @("aliexpress_attributename_id", "aliexpress_attribute_value_id", "attributename", "attribute_value")
          sku_keys = @("sku_code", "sku_attributes_list", "inventory", "price", "discount_price", "ean_code", "bar_code", "inventory_deduction_strategy", "weight", "packagelength", "packageheight", "packagewidth", "multi_country_price_configuration")
          sku_attribute_keys = @("sku_attributename_id", "sku_attribute_value_id", "sku_attribute_value", "sku_image_url")
        }
        redefining_api = @{
          product_keys = @("subject_list", "category_id", "image_u_r_ls", "aeop_ae_product_propertys", "aeop_ae_product_s_k_us", "product_unit", "is_pack_sell", "product_price", "currency_code", "detail_source_list")
          media_keys = @("image_u_r_ls", "detail_source_list.mobile_detail", "aeop_s_k_u_property.sku_image", "image_url_list")
          specification_keys = @("aeop_ae_product_propertys", "aeop_s_k_u_property")
          property_keys = @("attr_name_id", "attr_value_id", "attr_name", "attr_value")
          sku_keys = @("id", "aeop_s_k_u_property", "currency_code", "ipm_sku_stock", "sku_price", "sku_stock", "sku_code", "barcode", "gross_weight", "packageheight", "packagewidth", "packagelength", "ean_code", "aeop_s_k_u_national_discount_price_list")
          sku_property_keys = @("sku_property_id", "property_value_id", "property_value_definition_name", "sku_image", "image_url_list")
        }
      }
      stock_price_fields = @{
        sellersku = "sku_code"
        saleprice = "sku_price"
        openingstock = "ipm_sku_stock"
        gtin = "ean_code"
        barcode = "bar_code"
      }
      sample_payload = @{
        subject_list = @(@{ locale = "en_US"; value = "{{name_en}}" })
        category_id = "{{external_category_id}}"
        brand_name = "{{brand_name}}"
        main_image_urls_list = @("{{main_image_url}}")
        attribute_list = @(
          @{ aliexpress_attributename_id = 14; aliexpress_attribute_value_id = "{{attribute_value_id}}"; attribute_value = "{{attribute_value}}" }
        )
        sku_info_list = @(
          @{
            sku_code = "{{sellersku}}"
            sku_attributes_list = @(
              @{ sku_attributename_id = 14; sku_attribute_value_id = "{{color_value_id}}"; sku_attribute_value = "{{color}}"; sku_image_url = "{{sku_image_url}}" },
              @{ sku_attributename_id = 5; sku_attribute_value_id = "{{size_value_id}}"; sku_attribute_value = "{{size}}" }
            )
            inventory = "{{openingstock}}"
            price = "{{saleprice}}"
            discount_price = "{{discount_price}}"
            ean_code = "{{gtin}}"
            bar_code = "{{barcode}}"
            inventory_deduction_strategy = "payment_success_deduct"
            weight = "{{packageweight}}"
            packagelength = "{{packagelength}}"
            packageheight = "{{packageheight}}"
            packagewidth = "{{packagewidth}}"
            multi_country_price_configuration = @{
              price_type = "absolute"
              country_price_list = @(
                @{ ship_to_country = "TH"; sku_price_by_country_list = @(@{ sku_code = "{{sellersku}}"; price = "{{saleprice_th}}" }) }
              )
            }
          }
        )
        product_unit = 100000015
        currency_code = "USD"
        product_price = "{{min_saleprice}}"
        detail_source_list = @(
          @{ locale = "en_US"; mobile_detail = '{"version":"2.0.0","moduleList":[]}' }
        )
      }
    }
  )
}

$matrices = @(
  @{
    code = "APPAREL_COLOR_SIZE"
    names = @(New-Names "เสื้อผ้า สี x ไซซ์" "Apparel color x size")
    matrix_type = "apparel"
    serial_tracking_mode = "none"
    businesscodes = $companyA
    media_assets = @(
      New-MediaAsset -Kind "main" -Uri "images/products/apparel-premium-tshirt-main.webp" -SortOrder 1
      New-MediaAsset -Kind "gallery" -Uri "images/products/apparel-premium-tshirt-detail.webp" -SortOrder 2
      New-MediaAsset -Kind "size_chart" -Uri "images/products/apparel-size-chart.webp" -SortOrder 3 -UseCase "SIZE_CHART"
      New-MediaAsset -Kind "sku" -Uri "images/products/apparel-tshirt-black.webp" -SortOrder 10 -OptionCode "COLOR" -OptionValue "BLACK" -UseCase "ATTRIBUTE_IMAGE"
      New-MediaAsset -Kind "sku" -Uri "images/products/apparel-tshirt-khaki.webp" -SortOrder 11 -OptionCode "COLOR" -OptionValue "KHAKI" -UseCase "ATTRIBUTE_IMAGE"
    )
    specification_groups = @(
      New-SpecGroup -Code "GENERAL" -Name "ข้อมูลทั่วไป" -Attributes @(
        New-SpecAttribute -Code "MATERIAL" -Name "วัสดุ" -InputType "multiselect" -Scope "product" -Required $true -Values @(
          New-AttributeValue -Code "COTTON" -Text "Cotton 100%"
        )
        New-SpecAttribute -Code "GENDER" -Name "เพศ" -InputType "single_select" -Scope "product" -Values @(
          New-AttributeValue -Code "UNISEX" -Text "Unisex"
        )
        New-SpecAttribute -Code "SEASON" -Name "ฤดูกาล" -InputType "multiselect" -Scope "product" -Values @(
          New-AttributeValue -Code "ALL_SEASON" -Text "All season"
        )
      )
      New-SpecGroup -Code "SALE_PROPS" -Name "ตัวเลือกขาย" -Attributes @(
        New-SpecAttribute -Code "COLOR" -Name "สี" -InputType "single_select" -Scope "sku" -SaleProp $true -Values @(
          New-AttributeValue -Code "BLACK" -Text "ดำ"
          New-AttributeValue -Code "NAVY" -Text "น้ำเงิน"
          New-AttributeValue -Code "KHAKI" -Text "กากี"
        )
        New-SpecAttribute -Code "SIZE" -Name "ไซซ์" -InputType "single_select" -Scope "sku" -SaleProp $true -Values @(
          New-AttributeValue -Code "M" -Text "M"
          New-AttributeValue -Code "L" -Text "L"
          New-AttributeValue -Code "XL" -Text "XL"
        )
      )
    )
    option_tiers = @(
      @{ tierno = 1; optioncode = "COLOR"; name = "สี/Color"; values = @("BLACK", "NAVY", "KHAKI", "GREEN", "CREAM") },
      @{ tierno = 2; optioncode = "SIZE"; name = "ขนาด/Size"; values = @("S", "M", "L", "XL", "2XL") }
    )
    sku_combinations = @(
      @{ sellersku = "APP-TSHIRT-BLACK-M"; barcode = "885100000001"; gtin = "885100000001"; optionvalues = @("BLACK", "M"); saleprice = 390; cost = 185; openingstock = 24; unitcode = "PCS"; packageweight = 0.25; packagelength = 24; packagewidth = 18; packageheight = 3 },
      @{ sellersku = "APP-PANTS-KHAKI-L"; barcode = "885100000002"; gtin = "885100000002"; optionvalues = @("KHAKI", "L"); saleprice = 790; cost = 420; openingstock = 12; unitcode = "PCS"; packageweight = 0.6; packagelength = 30; packagewidth = 22; packageheight = 5 },
      @{ sellersku = "APP-JACKET-NAVY-XL"; barcode = "885100000003"; gtin = "885100000003"; optionvalues = @("NAVY", "XL"); saleprice = 1290; cost = 680; openingstock = 8; unitcode = "PCS"; packageweight = 0.9; packagelength = 36; packagewidth = 28; packageheight = 8 }
    )
    import_attribute_maps = @(
      @{ sourcename = "สี/Color"; targetoptioncode = "COLOR" },
      @{ sourcename = "ขนาด/Size"; targetoptioncode = "SIZE" }
    )
    integration_profiles = @(New-ExternalIntegrationProfiles)
    payload_examples = @(New-PayloadExamples -Family "apparel")
  },
  @{
    code = "MOBILE_COLOR_STORAGE"
    names = @(New-Names "มือถือ สี x ความจุ" "Mobile color x storage")
    matrix_type = "mobile_phone"
    serial_tracking_mode = "imei"
    businesscodes = $companyB
    media_assets = @(
      New-MediaAsset -Kind "main" -Uri "images/products/mobile-pro-main.webp" -SortOrder 1
      New-MediaAsset -Kind "gallery" -Uri "images/products/mobile-pro-camera.webp" -SortOrder 2
      New-MediaAsset -Kind "video" -Uri "videos/products/mobile-pro-demo.mp4" -SortOrder 3 -UseCase "VIDEO"
      New-MediaAsset -Kind "sku" -Uri "images/products/mobile-pro-black.webp" -SortOrder 10 -OptionCode "COLOR" -OptionValue "BLACK" -UseCase "ATTRIBUTE_IMAGE"
      New-MediaAsset -Kind "sku" -Uri "images/products/mobile-pro-white.webp" -SortOrder 11 -OptionCode "COLOR" -OptionValue "WHITE" -UseCase "ATTRIBUTE_IMAGE"
    )
    specification_groups = @(
      New-SpecGroup -Code "DEVICE" -Name "ข้อมูลเครื่อง" -Attributes @(
        New-SpecAttribute -Code "BRAND" -Name "ยี่ห้อ" -InputType "single_select" -Scope "product" -Required $true -Values @(
          New-AttributeValue -Code "SOLAO_MOBILE" -Text "Solao Mobile"
        )
        New-SpecAttribute -Code "OS" -Name "ระบบปฏิบัติการ" -InputType "single_select" -Scope "product" -Values @(
          New-AttributeValue -Code "ANDROID" -Text "Android"
        )
        New-SpecAttribute -Code "WARRANTY_MONTHS" -Name "ประกัน" -InputType "number" -Scope "product" -Values @(
          New-AttributeValue -Code "12" -Text "12" -UnitCode "MONTH"
        )
      )
      New-SpecGroup -Code "SALE_PROPS" -Name "ตัวเลือกขาย" -Attributes @(
        New-SpecAttribute -Code "COLOR" -Name "สี" -InputType "single_select" -Scope "sku" -SaleProp $true -Values @(
          New-AttributeValue -Code "BLACK" -Text "ดำ"
          New-AttributeValue -Code "WHITE" -Text "ขาว"
          New-AttributeValue -Code "NAVY" -Text "น้ำเงิน"
        )
        New-SpecAttribute -Code "STORAGE" -Name "ความจุ" -InputType "single_select" -Scope "sku" -SaleProp $true -Values @(
          New-AttributeValue -Code "128GB" -Text "128GB"
          New-AttributeValue -Code "256GB" -Text "256GB"
          New-AttributeValue -Code "512GB" -Text "512GB"
        )
      )
      New-SpecGroup -Code "COMPLIANCE" -Name "เลขเครื่อง" -Attributes @(
        New-SpecAttribute -Code "IMEI_REQUIRED" -Name "ต้องบันทึก IMEI" -InputType "boolean" -Scope "compliance" -Required $true -Values @(
          New-AttributeValue -Code "YES" -Text "Yes"
        )
      )
    )
    option_tiers = @(
      @{ tierno = 1; optioncode = "COLOR"; name = "สี/Color"; values = @("BLACK", "WHITE", "NAVY") },
      @{ tierno = 2; optioncode = "STORAGE"; name = "ความจุ/Storage"; values = @("128GB", "256GB", "512GB") }
    )
    sku_combinations = @(
      @{ sellersku = "MB-PRO-BLACK-128"; barcode = "885200000001"; gtin = "885200000001"; optionvalues = @("BLACK", "128GB"); saleprice = 18900; cost = 16200; openingstock = 6; unitcode = "PCS"; packageweight = 0.45; packagelength = 18; packagewidth = 10; packageheight = 6; serialidentifiers = @("IMEI") },
      @{ sellersku = "MB-PRO-WHITE-256"; barcode = "885200000002"; gtin = "885200000002"; optionvalues = @("WHITE", "256GB"); saleprice = 21900; cost = 18800; openingstock = 4; unitcode = "PCS"; packageweight = 0.45; packagelength = 18; packagewidth = 10; packageheight = 6; serialidentifiers = @("IMEI") },
      @{ sellersku = "MB-MAX-NAVY-512"; barcode = "885200000003"; gtin = "885200000003"; optionvalues = @("NAVY", "512GB"); saleprice = 29900; cost = 25700; openingstock = 3; unitcode = "PCS"; packageweight = 0.5; packagelength = 18; packagewidth = 10; packageheight = 6; serialidentifiers = @("IMEI") }
    )
    import_attribute_maps = @(
      @{ sourcename = "Color"; targetoptioncode = "COLOR" },
      @{ sourcename = "Storage"; targetoptioncode = "STORAGE" },
      @{ sourcename = "Memory"; targetoptioncode = "STORAGE" }
    )
    integration_profiles = @(New-ExternalIntegrationProfiles)
    payload_examples = @(New-PayloadExamples -Family "mobile")
  },
  @{
    code = "SIM_NETWORK_PLAN"
    names = @(New-Names "ซิม เครือข่าย x แพ็กเกจ" "SIM network x plan")
    matrix_type = "sim"
    serial_tracking_mode = "iccid"
    businesscodes = $companyC
    media_assets = @(
      New-MediaAsset -Kind "main" -Uri "images/products/sim-starter-main.webp" -SortOrder 1
      New-MediaAsset -Kind "gallery" -Uri "images/products/sim-package-back.webp" -SortOrder 2
      New-MediaAsset -Kind "description" -Uri "images/products/sim-coverage-map.webp" -SortOrder 3 -UseCase "DESCRIPTION_IMAGE"
    )
    specification_groups = @(
      New-SpecGroup -Code "SIM_INFO" -Name "ข้อมูลซิม" -Attributes @(
        New-SpecAttribute -Code "NETWORK" -Name "เครือข่าย" -InputType "single_select" -Scope "product" -Required $true -Values @(
          New-AttributeValue -Code "SOLAO_TEL" -Text "Solao Tel"
        )
        New-SpecAttribute -Code "REGISTRATION_REQUIRED" -Name "ต้องลงทะเบียนซิม" -InputType "boolean" -Scope "compliance" -Required $true -Values @(
          New-AttributeValue -Code "YES" -Text "Yes"
        )
      )
      New-SpecGroup -Code "SALE_PROPS" -Name "ตัวเลือกขาย" -Attributes @(
        New-SpecAttribute -Code "SIM_TYPE" -Name "ชนิดซิม" -InputType "single_select" -Scope "sku" -SaleProp $true -Values @(
          New-AttributeValue -Code "NANO_SIM" -Text "Nano SIM"
          New-AttributeValue -Code "ESIM" -Text "eSIM"
        )
        New-SpecAttribute -Code "PLAN" -Name "แพ็กเกจ" -InputType "single_select" -Scope "sku" -SaleProp $true -Values @(
          New-AttributeValue -Code "PREPAID" -Text "Prepaid"
          New-AttributeValue -Code "DATA_10GB" -Text "Data 10GB"
          New-AttributeValue -Code "DATA_UNLIMITED" -Text "Unlimited"
        )
      )
      New-SpecGroup -Code "COMPLIANCE" -Name "เลขซิม" -Attributes @(
        New-SpecAttribute -Code "ICCID_REQUIRED" -Name "ต้องบันทึก ICCID" -InputType "boolean" -Scope "compliance" -Required $true -Values @(
          New-AttributeValue -Code "YES" -Text "Yes"
        )
      )
    )
    option_tiers = @(
      @{ tierno = 1; optioncode = "SIM_TYPE"; name = "ชนิดซิม/SIM type"; values = @("NANO_SIM", "ESIM") },
      @{ tierno = 2; optioncode = "PLAN"; name = "แพ็กเกจ/Plan"; values = @("PREPAID", "DATA_10GB", "DATA_UNLIMITED") }
    )
    sku_combinations = @(
      @{ sellersku = "SIM-NANO-PREPAID"; barcode = "885300000001"; gtin = "885300000001"; optionvalues = @("NANO_SIM", "PREPAID"); saleprice = 49; cost = 18; openingstock = 100; unitcode = "PCS"; serialidentifiers = @("ICCID") },
      @{ sellersku = "SIM-ESIM-10GB"; barcode = "885300000002"; gtin = "885300000002"; optionvalues = @("ESIM", "DATA_10GB"); saleprice = 199; cost = 88; openingstock = 50; unitcode = "PCS"; serialidentifiers = @("ICCID") },
      @{ sellersku = "SIM-NANO-UNLIMITED"; barcode = "885300000003"; gtin = "885300000003"; optionvalues = @("NANO_SIM", "DATA_UNLIMITED"); saleprice = 399; cost = 210; openingstock = 40; unitcode = "PCS"; serialidentifiers = @("ICCID") }
    )
    import_attribute_maps = @(
      @{ sourcename = "SIM Type"; targetoptioncode = "SIM_TYPE" },
      @{ sourcename = "Package"; targetoptioncode = "PLAN" },
      @{ sourcename = "Plan"; targetoptioncode = "PLAN" }
    )
    integration_profiles = @(New-ExternalIntegrationProfiles)
    payload_examples = @(New-PayloadExamples -Family "sim")
  },
  @{
    code = "COMPUTER_CPU_RAM_STORAGE"
    names = @(New-Names "คอมพิวเตอร์ CPU x RAM x Storage" "Computer CPU x RAM x storage")
    matrix_type = "computer"
    serial_tracking_mode = "serial_no"
    businesscodes = $allCompanies
    media_assets = @(
      New-MediaAsset -Kind "main" -Uri "images/products/notebook-workstation-main.webp" -SortOrder 1
      New-MediaAsset -Kind "gallery" -Uri "images/products/notebook-ports.webp" -SortOrder 2
      New-MediaAsset -Kind "gallery" -Uri "images/products/notebook-keyboard.webp" -SortOrder 3
      New-MediaAsset -Kind "description" -Uri "images/products/notebook-spec-table.webp" -SortOrder 4 -UseCase "DESCRIPTION_IMAGE"
    )
    specification_groups = @(
      New-SpecGroup -Code "COMPUTER_SPEC" -Name "สเปกคอมพิวเตอร์" -Attributes @(
        New-SpecAttribute -Code "CPU" -Name "CPU" -InputType "single_select" -Scope "sku" -Required $true -SaleProp $true -Values @(
          New-AttributeValue -Code "I5" -Text "Intel Core i5"
          New-AttributeValue -Code "I7" -Text "Intel Core i7"
          New-AttributeValue -Code "R7" -Text "Ryzen 7"
        )
        New-SpecAttribute -Code "RAM" -Name "RAM" -InputType "single_select" -Scope "sku" -Required $true -SaleProp $true -Values @(
          New-AttributeValue -Code "16GB" -Text "16GB"
          New-AttributeValue -Code "32GB" -Text "32GB"
        )
        New-SpecAttribute -Code "STORAGE" -Name "Storage" -InputType "single_select" -Scope "sku" -Required $true -SaleProp $true -Values @(
          New-AttributeValue -Code "512GB" -Text "512GB SSD"
          New-AttributeValue -Code "1TB" -Text "1TB SSD"
        )
        New-SpecAttribute -Code "WARRANTY_MONTHS" -Name "ประกัน" -InputType "number" -Scope "product" -Values @(
          New-AttributeValue -Code "36" -Text "36" -UnitCode "MONTH"
        )
      )
      New-SpecGroup -Code "COMPLIANCE" -Name "เลขเครื่อง" -Attributes @(
        New-SpecAttribute -Code "SERIAL_NO_REQUIRED" -Name "ต้องบันทึก Serial No." -InputType "boolean" -Scope "compliance" -Required $true -Values @(
          New-AttributeValue -Code "YES" -Text "Yes"
        )
        New-SpecAttribute -Code "MAC_ADDRESS_OPTIONAL" -Name "MAC address" -InputType "text" -Scope "compliance" -Values @()
      )
    )
    option_tiers = @(
      @{ tierno = 1; optioncode = "CPU"; name = "CPU"; values = @("I5", "I7", "R5", "R7") },
      @{ tierno = 2; optioncode = "RAM"; name = "RAM"; values = @("8GB", "16GB", "32GB") },
      @{ tierno = 3; optioncode = "STORAGE"; name = "Storage"; values = @("512GB", "1TB") }
    )
    sku_combinations = @(
      @{ sellersku = "NB-I5-16-512"; barcode = "885400000001"; gtin = "885400000001"; optionvalues = @("I5", "16GB", "512GB"); saleprice = 24900; cost = 21100; openingstock = 5; unitcode = "PCS"; packageweight = 2.4; packagelength = 42; packagewidth = 32; packageheight = 8; serialidentifiers = @("SERIAL_NO", "MAC_ADDRESS") },
      @{ sellersku = "NB-I7-32-1TB"; barcode = "885400000002"; gtin = "885400000002"; optionvalues = @("I7", "32GB", "1TB"); saleprice = 45900; cost = 39800; openingstock = 2; unitcode = "PCS"; packageweight = 2.6; packagelength = 42; packagewidth = 32; packageheight = 8; serialidentifiers = @("SERIAL_NO", "MAC_ADDRESS") },
      @{ sellersku = "PC-R7-32-1TB"; barcode = "885400000003"; gtin = "885400000003"; optionvalues = @("R7", "32GB", "1TB"); saleprice = 38900; cost = 33100; openingstock = 3; unitcode = "PCS"; packageweight = 8.5; packagelength = 55; packagewidth = 28; packageheight = 50; serialidentifiers = @("SERIAL_NO") }
    )
    import_attribute_maps = @(
      @{ sourcename = "Processor"; targetoptioncode = "CPU" },
      @{ sourcename = "CPU"; targetoptioncode = "CPU" },
      @{ sourcename = "RAM"; targetoptioncode = "RAM" },
      @{ sourcename = "Storage"; targetoptioncode = "STORAGE" }
    )
    integration_profiles = @(New-ExternalIntegrationProfiles)
    payload_examples = @(New-PayloadExamples -Family "computer")
  }
)

$created = @{
  colors = 0
  sizes = 0
  matrices = 0
}

foreach ($item in $colors) {
  $guid = "seed-product-color-$($item.code.ToLowerInvariant())"
  $data = @{
    guidfixed = $guid
    holdingcode = $HoldingCode
    code = $item.code
    names = @(New-Names $item.th $item.en)
    hex_color = $item.hex
    color_family = $item.family
    aliases = @($item.aliases)
    businesscodes = @($item.businesscodes)
    isdisabled = $false
    seed_group = "product_variant_master"
  }
  [void](Invoke-AtlasUpsert -Collection "product_colors" -Data $data -Token $token)
  $created.colors++
}

foreach ($item in $sizes) {
  $guid = "seed-product-size-$($item.code.ToLowerInvariant().Replace('_', '-'))"
  $data = @{
    guidfixed = $guid
    holdingcode = $HoldingCode
    code = $item.code
    names = @(New-Names $item.th $item.en)
    size_system = $item.system
    size_type = $item.type
    sortorder = $item.order
    aliases = @($item.aliases)
    businesscodes = @($item.businesscodes)
    isdisabled = $false
    seed_group = "product_variant_master"
  }
  [void](Invoke-AtlasUpsert -Collection "product_sizes" -Data $data -Token $token)
  $created.sizes++
}

foreach ($item in $matrices) {
  $guid = "seed-product-variant-matrix-$($item.code.ToLowerInvariant().Replace('_', '-'))"
  $data = @{
    guidfixed = $guid
    holdingcode = $HoldingCode
    code = $item.code
    names = @($item.names)
    matrix_type = $item.matrix_type
    serial_tracking_mode = $item.serial_tracking_mode
    option_tiers = @($item.option_tiers)
    sku_combinations = @($item.sku_combinations)
    media_assets = @($item.media_assets)
    specification_groups = @($item.specification_groups)
    import_attribute_maps = @($item.import_attribute_maps)
    integration_profiles = @($item.integration_profiles)
    payload_examples = @($item.payload_examples)
    businesscodes = @($item.businesscodes)
    isdisabled = $false
    seed_group = "product_variant_master"
  }
  [void](Invoke-AtlasUpsert -Collection "product_variant_matrices" -Data $data -Token $token)
  $created.matrices++
}

[pscustomobject]@{
  holdingcode = $HoldingCode
  username = $Username
  company_codes_found = $companyCodes
  upserted_this_run = $created
  totals = @{
    product_colors = Invoke-AtlasGetCount -Collection "product_colors" -Token $token
    product_sizes = Invoke-AtlasGetCount -Collection "product_sizes" -Token $token
    product_variant_matrices = Invoke-AtlasGetCount -Collection "product_variant_matrices" -Token $token
  }
} | ConvertTo-Json -Depth 20
