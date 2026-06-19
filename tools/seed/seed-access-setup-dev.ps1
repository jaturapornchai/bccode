#Requires -Version 7.0

param(
  [Parameter(Mandatory = $true)][string]$HoldingCode,
  [Parameter(Mandatory = $true)][string]$Username,
  [string]$FrontendUrl = "http://localhost:3000",
  [string]$BackendUrl = "http://localhost:3000/backend/goapi",
  [string]$MainApiUrl = "http://localhost:8888",
  [int]$CompanyCount = 8,
  [int]$BranchesPerCompany = 4,
  [int]$PermissionCount = 30,
  [int]$GroupCount = 8,
  [int]$UserCount = 12,
  [int]$ApprovalCount = 10
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

function Invoke-SystemApi {
  param(
    [string]$Method,
    [string]$Path,
    [object]$Payload,
    [string]$Token
  )
  $bodyFile = [IO.Path]::Combine($env:TEMP, "bc-access-seed-body-$([guid]::NewGuid()).json")
  $outFile = [IO.Path]::Combine($env:TEMP, "bc-access-seed-out-$([guid]::NewGuid()).json")
  try {
    $args = @(
      "--max-time", "60",
      "-sS",
      "-o", $outFile,
      "-w", "%{http_code}",
      "-X", $Method,
      "-H", "Authorization: Bearer $Token",
      "-H", "Content-Type: application/json",
      "-H", "x-bc-backend-url: $BackendUrl"
    )
    if ($null -ne $Payload) {
      $Payload | ConvertTo-Json -Depth 60 | Set-Content -LiteralPath $bodyFile -Encoding UTF8
      $args += @("--data-binary", "@$bodyFile")
    }
    $args += "$FrontendUrl$Path"
    $status = curl.exe @args
    $raw = if (Test-Path $outFile) { Get-Content -LiteralPath $outFile -Raw } else { "" }
    if ([int]$status -lt 200 -or [int]$status -ge 300) {
      throw "HTTP $status $Method $Path $raw"
    }
    if ([string]::IsNullOrWhiteSpace($raw)) { return $null }
    return $raw | ConvertFrom-Json
  }
  finally {
    Remove-Item -LiteralPath $bodyFile -ErrorAction SilentlyContinue
    Remove-Item -LiteralPath $outFile -ErrorAction SilentlyContinue
  }
}

function Invoke-MainApi {
  param(
    [string]$Method,
    [string]$Path,
    [object]$Payload,
    [string]$Token
  )
  $bodyFile = [IO.Path]::Combine($env:TEMP, "bc-access-main-body-$([guid]::NewGuid()).json")
  $outFile = [IO.Path]::Combine($env:TEMP, "bc-access-main-out-$([guid]::NewGuid()).json")
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
      $Payload | ConvertTo-Json -Depth 60 | Set-Content -LiteralPath $bodyFile -Encoding UTF8
      $args += @("--data-binary", "@$bodyFile")
    }
    $args += "$MainApiUrl$Path"
    $status = curl.exe @args
    $raw = if (Test-Path $outFile) { Get-Content -LiteralPath $outFile -Raw } else { "" }
    if ([int]$status -lt 200 -or [int]$status -ge 300) {
      throw "HTTP $status $Method $Path $raw"
    }
    if ([string]::IsNullOrWhiteSpace($raw)) { return $null }
    return $raw | ConvertFrom-Json
  }
  finally {
    Remove-Item -LiteralPath $bodyFile -ErrorAction SilentlyContinue
    Remove-Item -LiteralPath $outFile -ErrorAction SilentlyContinue
  }
}

function New-GuidFixed {
  param([string]$Kind, [int]$Index)
  return "bcseed-$Kind-$('{0:D3}' -f $Index)"
}

function New-ScopeRules {
  param([int]$Index)
  if ($Index % 5 -eq 0) {
    return @(@{ scopetype = "holding"; allbranches = $false })
  }
  $companyCode = "BCS{0:D3}" -f ((($Index - 1) % 8) + 1)
  if ($Index % 2 -eq 0) {
    return @(@{ scopetype = "company"; businesscode = $companyCode; allbranches = $true })
  }
  return @(@{
    scopetype = "branch"
    businesscode = $companyCode
    branchcode = "{0:D5}" -f ((($Index - 1) % 4))
    allbranches = $false
  })
}

function New-Names {
  param([string]$Thai, [string]$English)
  return @(
    @{ code = "th"; name = $Thai; isauto = $false; isdelete = $false },
    @{ code = "en"; name = $English; isauto = $false; isdelete = $false },
    @{ code = "lo"; name = ""; isauto = $false; isdelete = $false },
    @{ code = "zh"; name = ""; isauto = $false; isdelete = $false }
  )
}

function Get-Records {
  param([string]$Slug, [string]$Token)
  $payload = Invoke-SystemApi -Method "GET" -Path "/api/system-settings/$Slug`?limit=1000&offset=0&page=1&holdingcode=$HoldingCode" -Payload $null -Token $Token
  if ($payload.data -is [array]) { return @($payload.data) }
  if ($payload.data) { return @($payload.data) }
  return @()
}

$token = Get-SelectedToken -HoldingCode $HoldingCode -Username $Username
$now = (Get-Date).ToUniversalTime().ToString("s") + "Z"

$companiesBefore = Invoke-MainApi -Method "GET" -Path "/organization/company" -Payload $null -Token $token
$companyByCode = @{}
foreach ($company in @($companiesBefore.data)) {
  if ($company.code) { $companyByCode[[string]$company.code] = $company }
}
$companyNamePool = @("อรุณพาณิชย์", "นทีซัพพลาย", "ภูมิใจค้าปลีก", "สยามคลังสินค้า", "ล้านนาบริการ", "บางกอกมาร์ท", "อีสานเทรดดิ้ง", "อันดามันฟู้ดส์")
$createdCompanies = 0
for ($i = 1; $i -le $CompanyCount; $i++) {
  $code = "BCS{0:D3}" -f $i
  if (-not $companyByCode.ContainsKey($code)) {
    $name = $companyNamePool[($i - 1) % $companyNamePool.Count]
    $payload = @{
      code = $code
      names = @(New-Names -Thai "บริษัท $name ตัวอย่าง จำกัด" -English "BC Seed Company $i Ltd.")
      taxid = "09999000{0:D5}" -f $i
      isactive = $true
    }
    $created = Invoke-MainApi -Method "POST" -Path "/organization/company" -Payload $payload -Token $token
    $companyByCode[$code] = [pscustomobject]@{ guidfixed = [string]$created.id; code = $code }
    $createdCompanies++
  }
}

$branchesBefore = Invoke-MainApi -Method "GET" -Path "/organization/branch" -Payload $null -Token $token
$branchKeys = @{}
foreach ($branch in @($branchesBefore.data)) {
  $branchKeys["$($branch.companyguid)|$($branch.code)"] = $true
}
$branchNamePool = @("สำนักงานใหญ่", "สาขารังสิต", "สาขาบางนา", "สาขาเชียงใหม่", "สาขาขอนแก่น", "สาขาหาดใหญ่")
$createdBranches = 0
for ($i = 1; $i -le $CompanyCount; $i++) {
  $company = $companyByCode["BCS{0:D3}" -f $i]
  $companyGuid = [string]$company.guidfixed
  if ([string]::IsNullOrWhiteSpace($companyGuid)) { continue }
  for ($b = 0; $b -lt $BranchesPerCompany; $b++) {
    $branchCode = "{0:D5}" -f $b
    $key = "$companyGuid|$branchCode"
    if ($branchKeys.ContainsKey($key)) { continue }
    $branchThai = if ($b -eq 0) { "สำนักงานใหญ่" } else { $branchNamePool[(($i + $b - 1) % ($branchNamePool.Count - 1)) + 1] }
    $payload = @{
      companyguid = $companyGuid
      code = $branchCode
      names = @(New-Names -Thai $branchThai -English $(if ($b -eq 0) { "Head Office" } else { "Branch $b" }))
      isactive = $true
    }
    [void](Invoke-MainApi -Method "POST" -Path "/organization/branch" -Payload $payload -Token $token)
    $branchKeys[$key] = $true
    $createdBranches++
  }
}

$companyProfilePayload = Invoke-SystemApi -Method "GET" -Path "/api/system-settings/company?holdingcode=$HoldingCode" -Payload $null -Token $token
$companyProfile = if ($companyProfilePayload.data) { $companyProfilePayload.data } else { $companyProfilePayload }
$companyProfile.holdingcode = $HoldingCode
if (-not $companyProfile.settings) {
  $companyProfile | Add-Member -NotePropertyName settings -NotePropertyValue ([pscustomobject]@{}) -Force
}
$companyProfile.settings.languageconfigs = @(
  @{ code = "th"; name = "ภาษาไทย"; isuse = $true; isdefault = $true; xsort = 1 },
  @{ code = "en"; name = "English"; isuse = $true; isdefault = $false; xsort = 2 },
  @{ code = "lo"; name = "ພາສາລາວ"; isuse = $true; isdefault = $false; xsort = 3 },
  @{ code = "zh"; name = "中文"; isuse = $true; isdefault = $false; xsort = 4 }
)
[void](Invoke-SystemApi -Method "PUT" -Path "/api/system-settings/activelanguages?holdingcode=$HoldingCode" -Payload $companyProfile -Token $token)

$permissionCodes = @()
for ($i = 1; $i -le $PermissionCount; $i++) {
  $code = "BCSEED-PERM-{0:D3}" -f $i
  $permissionCodes += $code
  $payload = @{
    holdingcode = $HoldingCode
    guidfixed = New-GuidFixed -Kind "perm" -Index $i
    permissioncode = $code
    permissionname = "สิทธิ์ตัวอย่าง $i"
    description = "ข้อมูลตัวอย่างสำหรับทดสอบหน้าจอและเมนู $i"
    isactive = $true
    scoperules = @(New-ScopeRules -Index $i)
    accessrules = @(
      @{ route = "/dashboard"; actions = @("view") },
      @{ route = "/product"; actions = @("view", "create", "edit") },
      @{ route = "/sale"; actions = @("view", "approve") }
    )
    updatedat = $now
  }
  [void](Invoke-SystemApi -Method "POST" -Path "/api/system-settings/permissiondefinition?holdingcode=$HoldingCode" -Payload $payload -Token $token)
}

$groupCodes = @()
for ($i = 1; $i -le $GroupCount; $i++) {
  $code = "BCSEED-GROUP-{0:D2}" -f $i
  $groupCodes += $code
  $start = (($i - 1) * 3) % $permissionCodes.Count
  $codes = @($permissionCodes[$start], $permissionCodes[($start + 1) % $permissionCodes.Count], $permissionCodes[($start + 2) % $permissionCodes.Count], $permissionCodes[($start + 3) % $permissionCodes.Count])
  $payload = @{
    holdingcode = $HoldingCode
    guidfixed = New-GuidFixed -Kind "group" -Index $i
    groupcode = $code
    groupname = "กลุ่มสิทธิ์ตัวอย่าง $i"
    description = "ชุดสิทธิ์ตัวอย่างสำหรับทีมงาน $i"
    isactive = $true
    scoperules = @(New-ScopeRules -Index $i)
    permissioncodes = $codes
    updatedat = $now
  }
  [void](Invoke-SystemApi -Method "POST" -Path "/api/system-settings/permissiongroup?holdingcode=$HoldingCode" -Payload $payload -Token $token)
}

$approvalCodes = @()
for ($i = 1; $i -le $ApprovalCount; $i++) {
  $code = "BCSEED-APR-{0:D2}" -f $i
  $approvalCodes += $code
  $payload = @{
    holdingcode = $HoldingCode
    guidfixed = New-GuidFixed -Kind "approval" -Index $i
    approvalcode = $code
    approvalname = "สิทธิ์อนุมัติตัวอย่าง $i"
    description = "วงเงินและเอกสารอนุมัติตัวอย่าง $i"
    isactive = $true
    approvalrules = @(New-ScopeRules -Index $i)
    approvals = @(
      @{ doctype = "PO"; limitamount = "{0}.00" -f (10000 * $i); currency = "THB"; levels = @(@{ level = 1; groupcode = $groupCodes[($i - 1) % $groupCodes.Count] }) },
      @{ doctype = "SO"; limitamount = "{0}.00" -f (8000 * $i); currency = "THB"; levels = @(@{ level = 1; groupcode = $groupCodes[$i % $groupCodes.Count] }) }
    )
    updatedat = $now
  }
  [void](Invoke-SystemApi -Method "POST" -Path "/api/system-settings/approvalsetting?holdingcode=$HoldingCode" -Payload $payload -Token $token)
}

$usernames = @()
for ($i = 1; $i -le $UserCount; $i++) {
  $username = "bcseed.user{0:D2}@example.com" -f $i
  $usernames += $username
  $payload = @{
    holdingcode = $HoldingCode
    username = $username
    email = $username
    userprofilename = "ผู้ใช้ตัวอย่าง $i"
    name = "ผู้ใช้ตัวอย่าง $i"
    role = if ($i -le 2) { 1 } else { 0 }
    isaccessdisabled = $false
    position = if ($i % 3 -eq 0) { "ผู้จัดการสาขา" } elseif ($i % 3 -eq 1) { "พนักงานขาย" } else { "บัญชี" }
    department = if ($i % 2 -eq 0) { "ฝ่ายขาย" } else { "สำนักงานใหญ่" }
    accessscopes = @(New-ScopeRules -Index $i)
    approvalcodes = @($approvalCodes[($i - 1) % $approvalCodes.Count])
    updatedat = $now
  }
  [void](Invoke-SystemApi -Method "POST" -Path "/api/system-settings/user?holdingcode=$HoldingCode" -Payload $payload -Token $token)
}

for ($i = 1; $i -le $UserCount; $i++) {
  $username = $usernames[$i - 1]
  $groupCode = $groupCodes[($i - 1) % $groupCodes.Count]
  $permStart = (($i - 1) * 2) % $permissionCodes.Count
  $payload = @{
    holdingcode = $HoldingCode
    guidfixed = New-GuidFixed -Kind "plink" -Index $i
    employeecode = $username
    employeename = "ผู้ใช้ตัวอย่าง $i"
    useruid = $username
    groupcode = $groupCode
    scoperules = @(New-ScopeRules -Index $i)
    permissioncodes = @($permissionCodes[$permStart], $permissionCodes[($permStart + 1) % $permissionCodes.Count], $permissionCodes[($permStart + 2) % $permissionCodes.Count])
    approvalcodes = @($approvalCodes[($i - 1) % $approvalCodes.Count], $approvalCodes[$i % $approvalCodes.Count])
    updatedat = $now
  }
  [void](Invoke-SystemApi -Method "POST" -Path "/api/system-settings/permissionlink?holdingcode=$HoldingCode" -Payload $payload -Token $token)
}

$summary = [ordered]@{
  holdingcode = $HoldingCode
  active_languages = 4
  companies_created = $createdCompanies
  branches_created = $createdBranches
  permissiondefinitions = (@(Get-Records -Slug "permissiondefinition" -Token $token) | Where-Object { $_.permissioncode -like "BCSEED-PERM-*" }).Count
  permissiongroups = (@(Get-Records -Slug "permissiongroup" -Token $token) | Where-Object { $_.groupcode -like "BCSEED-GROUP-*" }).Count
  users = (@(Get-Records -Slug "user" -Token $token) | Where-Object { $_.username -like "bcseed.user*" }).Count
  permissionlinks = (@(Get-Records -Slug "permissionlink" -Token $token) | Where-Object { $_.employeecode -like "bcseed.user*" }).Count
  approvalsettings = (@(Get-Records -Slug "approvalsetting" -Token $token) | Where-Object { $_.approvalcode -like "BCSEED-APR-*" }).Count
}

$summary | ConvertTo-Json -Depth 8
