#Requires -Version 7.0

param(
  [Parameter(Mandatory = $true)][string]$HoldingCode,
  [Parameter(Mandatory = $true)][string]$Username,
  [string]$BaseUrl = "http://localhost:8888",
  [int]$CompanyCount = 5,
  [int]$MinExtraBranches = 1,
  [int]$MaxExtraBranches = 3
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
  $bodyFile = [IO.Path]::Combine($env:TEMP, "bc-org-body-$([guid]::NewGuid()).json")
  $outFile = [IO.Path]::Combine($env:TEMP, "bc-org-out-$([guid]::NewGuid()).json")
  try {
    $args = @(
      "--max-time", "30",
      "-sS",
      "-o", $outFile,
      "-w", "%{http_code}",
      "-X", $Method,
      "-H", "Authorization: Bearer $Token",
      "-H", "Content-Type: application/json"
    )
    if ($null -ne $Payload) {
      $Payload | ConvertTo-Json -Depth 30 | Set-Content -LiteralPath $bodyFile -Encoding UTF8
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

$token = Get-SelectedToken -HoldingCode $HoldingCode -Username $Username
$existing = Invoke-GetJson -Path "/organization/company" -Token $token
$existingCodes = @{}
foreach ($row in @($existing.data)) {
  if ($row.code) { $existingCodes[[string]$row.code] = $true }
}

$stamp = (Get-Date).ToUniversalTime().ToString("yyMMddHHmm")
$companyNames = @("อรุณพาณิชย์", "นทีซัพพลาย", "ภูมิใจค้าปลีก", "สยามคลังสินค้า", "ล้านนาบริการ")
$branchNamePool = @("สาขารังสิต", "สาขาบางนา", "สาขาเชียงใหม่", "สาขาขอนแก่น", "สาขาหาดใหญ่", "สาขานครปฐม", "สาขาศรีราชา", "สาขาโคราช")
$created = @()

for ($i = 1; $i -le $CompanyCount; $i++) {
  $nameIndex = ($i - 1) % $companyNames.Count
  $code = "T$stamp$i".ToUpperInvariant()
  while ($existingCodes.ContainsKey($code)) {
    $code = "T$stamp$i$((Get-Random -Minimum 10 -Maximum 99))".ToUpperInvariant()
  }

  $companyThai = "บริษัท $($companyNames[$nameIndex]) ทดสอบ จำกัด"
  $companyEn = "Random Test Company $i Ltd."
  $companyPayload = @{
    code = $code
    names = @(New-Names -Thai $companyThai -English $companyEn)
    tax_id = ("0999999999{0:D3}" -f $i)
    isactive = $true
  }
  $companyResult = Invoke-JsonApi -Method "POST" -Path "/organization/company" -Payload $companyPayload -Token $token
  $companyGuid = [string]$companyResult.id
  $extraBranchCount = Get-Random -Minimum $MinExtraBranches -Maximum ($MaxExtraBranches + 1)
  $branchCodes = @("00000")

  for ($b = 1; $b -le $extraBranchCount; $b++) {
    $branchCode = "{0:D5}" -f $b
    $branchThai = $branchNamePool[(($i + $b - 2) % $branchNamePool.Count)]
    $branchEn = "Branch $b"
    $branchPayload = @{
      companyguid = $companyGuid
      code = $branchCode
      names = @(New-Names -Thai $branchThai -English $branchEn)
      isactive = $true
    }
    [void](Invoke-JsonApi -Method "POST" -Path "/organization/branch" -Payload $branchPayload -Token $token)
    $branchCodes += $branchCode
  }

  $created += [pscustomobject]@{
    code = $code
    name_th = $companyThai
    companyguid = $companyGuid
    branches = ($branchCodes -join ",")
  }
  $existingCodes[$code] = $true
}

$companiesAfter = Invoke-GetJson -Path "/organization/company" -Token $token
$branchesAfter = Invoke-GetJson -Path "/organization/branch" -Token $token
[pscustomobject]@{
  holdingcode = $HoldingCode
  companies_created = $created.Count
  total_companies = @($companiesAfter.data).Count
  total_branches = @($branchesAfter.data).Count
  created = $created
} | ConvertTo-Json -Depth 10
