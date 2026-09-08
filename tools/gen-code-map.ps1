# gen-code-map.ps1 - auto-generate docs/reference/CODE-MAP.md index of large source files.
# Indexes function, component, and section-comment names (type/const-export are parsed and mostly
# filtered out - names matching the handle/on/load/create... prefix filter still pass) so
# any agent can jump straight to the right line range instead of grep+re-read.
#
# Usage (regenerate):  pwsh -NoProfile -File tools/gen-code-map.ps1
# Usage (verify only): pwsh -NoProfile -File tools/gen-code-map.ps1 -Check
#   -Check regenerates in memory and compares with the committed file, ignoring the volatile
#   "> Generated:" line. Exit 0 = in sync, exit 1 = drifted (CI and the pre-commit hook use this).
# Output: docs/reference/CODE-MAP.md (repo root is derived from this script's location,
#   so it works on any machine and on Linux CI runners - do not hard-code a path here again).
[CmdletBinding()]
param(
    [switch]$Check
)
$ErrorActionPreference = "Stop"

$repoRoot = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path.TrimEnd([System.IO.Path]::DirectorySeparatorChar)
$frontendSrc = Join-Path $repoRoot (Join-Path "frontend" "src")
$backendRoot = Join-Path $repoRoot "backend"
$outFile = Join-Path $repoRoot (Join-Path "docs" (Join-Path "reference" "CODE-MAP.md"))
$lineThreshold = 1000  # only index files at or above this size

# Patterns: line starts with one of these anchors. Capture leading indent + name.
$funcPattern = '^\s*(export\s+)?(default\s+)?(async\s+)?function\s+([A-Za-z0-9_]+)'
# Go: "func Name(" and "func (r *Recv) Name(" - TS "function foo" never matches ('func' + 't').
$goFuncPattern = '^func\s+(\([^)]*\)\s*)?([A-Za-z0-9_]+)'
$arrowPattern = '^\s*(export\s+)?const\s+([A-Za-z0-9_]+)\s*=\s*(async\s*)?\('
$arrowGenPattern = '^\s*(export\s+)?const\s+([A-Za-z0-9_]+)\s*=\s*(async\s*)?function'
$typePattern = '^\s*(export\s+)?(type|interface)\s+([A-Za-z0-9_]+)'
$constExportPattern = '^\s*export\s+const\s+([A-Za-z0-9_]+)\s*='
$commentSectionPattern = '^\s*(//|/\*|\{\/\*)\s*[-=]{2,}\s*(SECTION|===|----|BLOCK|PANEL|BEGIN|END|PART|STEP)?\s*([A-Za-z0-9 _\-]+?)\s*[-=]{0,}\s*(\*/|\}*\/\})?\s*$'

function ConvertTo-RelPath {
    # Repo-relative path with forward slashes, on every OS.
    param([string]$FullPath)
    return $FullPath.Substring($repoRoot.Length).TrimStart('\', '/').Replace('\', '/')
}

function Test-ExcludedPath {
    param([string]$FullPath)
    # Normalize to forward slashes first: Windows gives '\', Linux CI gives '/'.
    $p = $FullPath.Replace('\', '/')
    return (
        $p -like '*/node_modules/*' -or
        $p -like '*/.next/*' -or
        $p -like '*.gen.*' -or
        $p -like '*/vendor/*' -or
        $p -like '*/dist/*' -or
        $p -like '*/__pycache__/*' -or
        $p -like '*/docs/docs.go' -or
        $p -like '*/mock/*' -or
        $p -like '*/mockdata/*'
    )
}

function Get-SourceFiles {
    param([string[]]$Roots, [int]$MinLines)
    $results = @()
    foreach ($root in $Roots) {
        if (-not (Test-Path $root)) { continue }
        $files = Get-ChildItem -Path $root -Recurse -File -ErrorAction SilentlyContinue |
            Where-Object {
                ($_.Extension -eq '.tsx' -or $_.Extension -eq '.ts' -or $_.Extension -eq '.go') -and
                -not (Test-ExcludedPath -FullPath $_.FullName)
            }
        foreach ($f in $files) {
            # Count ALL lines (Measure-Object -Line skips blank lines -> wrong "(N lines)" + wrong threshold).
            $lineCount = @(Get-Content -LiteralPath $f.FullName -Encoding UTF8 -ErrorAction SilentlyContinue).Count
            if ($lineCount -ge $MinLines) {
                $results += [pscustomobject]@{
                    path  = $f.FullName
                    rel   = (ConvertTo-RelPath -FullPath $f.FullName)
                    lines = $lineCount
                }
            }
        }
    }
    return $results
}

function Index-File {
    param([string]$Path)
    $entries = @()
    $lineNum = 0
    foreach ($line in Get-Content -LiteralPath $Path -Encoding UTF8 -ErrorAction SilentlyContinue) {
        $lineNum++
        $kind = ""
        $name = ""

        if ($line -match $funcPattern) {
            $kind = "function"
            $name = $Matches[4]
        }
        elseif ($line -match $goFuncPattern) {
            $kind = "function"
            $name = $Matches[2]
        }
        elseif ($line -match $arrowGenPattern) {
            $kind = "fn-const"
            $name = $Matches[2]
        }
        elseif ($line -match $arrowPattern) {
            $kind = "const-arrow"
            $name = $Matches[2]
        }
        elseif ($line -match $typePattern) {
            $kind = $Matches[2]
            $name = $Matches[3]
        }
        elseif ($line -match $constExportPattern) {
            $kind = "export-const"
            $name = $Matches[1]
        }
        elseif ($line -match $commentSectionPattern -and $Matches[3]) {
            $kind = "section"
            $name = $Matches[3]
        }

        # PowerShell treats " " as truthy, so a divider comment like /* ------ */ used to emit
        # a junk row with a blank name. Require at least one alphanumeric char.
        # Identifiers may run to 80 chars (Go test names do; at 60 ALL 11 funcs of
        # appurchasereceive_transaction_phaser_test.go were dropped). Comment-derived
        # "section" names stay capped at 60 - past that they are wrapped prose, not headings.
        $maxName = if ($kind -eq "section") { 60 } else { 80 }
        if ($name -and $name -match '[A-Za-z0-9]' -and $name.Trim().Length -le $maxName) {
            $entries += [pscustomobject]@{ line = $lineNum; kind = $kind; name = $name.Trim() }
        }
    }
    return $entries
}

# Deterministic order: biggest file first, ties broken by repo-relative path.
# Without the secondary key the order follows OS directory enumeration, so Windows and
# Linux CI produce different files and -Check fails for no real reason.
# Sort-Object compares strings with the CURRENT CULTURE, so a Thai and an invariant machine can
# order two files with equal line counts differently and -Check fails for no real reason.
# CompareOrdinal is byte order: same answer everywhere.
$fileList = [System.Collections.Generic.List[object]]::new()
foreach ($item in (Get-SourceFiles -Roots @($frontendSrc, $backendRoot) -MinLines $lineThreshold)) {
    [void]$fileList.Add($item)
}
$fileList.Sort([System.Comparison[object]] {
    param($x, $y)
    $c = $y.lines.CompareTo($x.lines)
    if ($c -ne 0) { return $c }
    return [string]::CompareOrdinal($x.rel, $y.rel)
})
$files = $fileList

# The date stamp must be Gregorian regardless of machine locale: on a Thai culture
# 'Get-Date -Format yyyy-MM-dd' stamps the Buddhist year 2569, not 2026 - hence InvariantCulture below.
$genCommit = "unknown"
try {
    $rev = (& git -C $repoRoot rev-parse --short HEAD 2>$null)
    if ($rev) { $genCommit = ($rev | Select-Object -First 1).Trim() }
} catch { $genCommit = "unknown" }

$sb = [System.Text.StringBuilder]::new()
[void]$sb.AppendLine("# CODE-MAP - large-file navigation index")
[void]$sb.AppendLine("")
[void]$sb.AppendLine("> Auto-generated by tools/gen-code-map.ps1 - index เฉพาะไฟล์ TS/TSX/Go ขนาด >= $lineThreshold บรรทัด ที่อยู่ใต้ frontend/src/ และ backend/ เท่านั้น (ข้าม node_modules, .next, vendor, dist, mock/mockdata, *.gen.*, backend/docs/docs.go และไม่รวม frontend/e2e/) - ไฟล์ที่ไม่อยู่ในแผนที่ไม่ได้แปลว่าเล็กกว่า $lineThreshold บรรทัดเสมอไป")
[void]$sb.AppendLine("> Read this FIRST before editing a big file so you can jump straight to the right line range")
[void]$sb.AppendLine("> instead of grep + re-read. Regenerate after major refactors.")
[void]$sb.AppendLine("> ตัวกันดริฟต์: git hook ``.githooks/pre-commit`` รัน ``-Check`` ทุกครั้งที่ commit แตะไฟล์ใหญ่ - ต้องติดตั้งเองครั้งเดียวต่อ clone ด้วย ``npm run hooks:install`` (CI job ``code-map-check`` เขียนไว้แล้วแต่ GitHub Actions ของ repo นี้ยังรันไม่ได้ - บัญชีถูกล็อกเรื่อง billing ตั้งแต่ 2026-09-02)")
[void]$sb.AppendLine("> Generated: $([datetime]::Now.ToString('yyyy-MM-dd', [System.Globalization.CultureInfo]::InvariantCulture)) @ commit $genCommit - ถ้า commit ปัจจุบันไม่ใช่อันนี้ ให้ถือว่าเลขบรรทัดอาจเลื่อนแล้ว ตรวจด้วย grep ก่อนใช้ หรือรัน tools/gen-code-map.ps1 ใหม่")
[void]$sb.AppendLine("")
[void]$sb.AppendLine("Files indexed: $($files.Count)")
[void]$sb.AppendLine("")

foreach ($f in $files) {
    [void]$sb.AppendLine("## $($f.rel) ($($f.lines) lines)")
    [void]$sb.AppendLine("")
    [void]$sb.AppendLine("| Line | Kind | Name |")
    [void]$sb.AppendLine("|---:|---|---|")

    $entries = Index-File -Path $f.path
    # Actionable navigation: functions, named components, section comments, and handlers.
    # Drop type/interface/export-const noise (easy to grep when needed) to keep the map concise.
    $filtered = $entries | Where-Object {
        $_.kind -in @('function','fn-const','section') -or
        ($_.kind -eq 'const-arrow' -and $_.name -cmatch '^[A-Z]') -or
        ($_.name -match '^(handle|on|render|load|save|fetch|open|close|toggle|run|reset|delete|create|update)')
    }

    foreach ($e in $filtered) {
        [void]$sb.AppendLine("| $($e.line) | $($e.kind) | ``$($e.name)`` |")
    }
    [void]$sb.AppendLine("")
}

# AppendLine uses the host newline (CRLF on Windows, LF on Linux). Normalize so the file is
# byte-identical everywhere and -Check does not fail purely on line endings.
$content = $sb.ToString().Replace("`r`n", "`n")

function Get-ComparableLines {
    # Drop the volatile "> Generated:" stamp so -Check compares content, not the clock.
    param([string]$Text)
    return @($Text.Replace("`r`n", "`n").Split("`n") | Where-Object { $_ -notlike '> Generated:*' })
}

if ($Check) {
    if (-not (Test-Path -LiteralPath $outFile)) {
        Write-Host "FAIL: $outFile is missing. Run: pwsh -NoProfile -File tools/gen-code-map.ps1"
        exit 1
    }
    $existing = [System.IO.File]::ReadAllText($outFile)
    $a = Get-ComparableLines -Text $existing
    $b = Get-ComparableLines -Text $content
    if ($a.Count -eq $b.Count) {
        $same = $true
        for ($i = 0; $i -lt $a.Count; $i++) {
            if ($a[$i] -cne $b[$i]) { $same = $false; break }
        }
        if ($same) {
            Write-Host ("CODE-MAP is in sync ({0} files indexed)." -f $files.Count)
            exit 0
        }
    }

    # Report the first few real differences so the failure is actionable, not just "files differ".
    Write-Host "FAIL: docs/reference/CODE-MAP.md is out of date - large source files moved but the map was not regenerated."
    Write-Host ""
    $max = [Math]::Max($a.Count, $b.Count)
    $shown = 0
    $section = "(header)"
    for ($i = 0; $i -lt $max -and $shown -lt 15; $i++) {
        $av = if ($i -lt $a.Count) { $a[$i] } else { "<missing>" }
        $bv = if ($i -lt $b.Count) { $b[$i] } else { "<missing>" }
        if ($bv -like '## *') { $section = $bv }
        if ($av -cne $bv) {
            Write-Host ("  {0}`n    committed: {1}`n    actual   : {2}" -f $section, $av, $bv)
            $shown++
        }
    }
    if ($shown -ge 15) { Write-Host "  ... (more differences suppressed)" }
    Write-Host ""
    Write-Host "Fix: pwsh -NoProfile -File tools/gen-code-map.ps1   then commit docs/reference/CODE-MAP.md"
    exit 1
}

[System.IO.File]::WriteAllText($outFile, $content, [System.Text.UTF8Encoding]::new($false))
Write-Host ("wrote {0} files -> {1} ({2} bytes)" -f $files.Count, $outFile, $content.Length)
