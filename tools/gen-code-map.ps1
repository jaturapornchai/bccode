# gen-code-map.ps1 - auto-generate docs/reference/CODE-MAP.md index of large source files.
# Indexes function, component, and section-comment names (type/const-export are parsed and mostly
# filtered out - names matching the handle/on/load/create... prefix filter still pass) so
# any agent can jump straight to the right line range instead of grep+re-read.
#
# Usage: powershell -NoProfile -ExecutionPolicy Bypass -File D:\bccode\tools\gen-code-map.ps1
# Output: D:\bccode\docs\reference\CODE-MAP.md
$ErrorActionPreference = "Stop"

$repoRoot = "D:\bccode"
$frontendSrc = Join-Path $repoRoot "frontend\src"
$backendRoot = Join-Path $repoRoot "backend"
$outFile = Join-Path $repoRoot "docs/reference/CODE-MAP.md"
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

function Test-ExcludedPath {
    param([string]$FullPath)
    # Use -like wildcards (no regex escaping headaches).
    return (
        $FullPath -like '*\node_modules\*' -or
        $FullPath -like '*\.next\*' -or
        $FullPath -like '*.gen.*' -or
        $FullPath -like '*\vendor\*' -or
        $FullPath -like '*\dist\*' -or
        $FullPath -like '*\__pycache__\*' -or
        $FullPath -like '*\docs\docs.go' -or
        $FullPath -like '*\mock\*' -or
        $FullPath -like '*\mockdata\*'
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
            $lineCount = @(Get-Content $f.FullName -ErrorAction SilentlyContinue).Count
            if ($lineCount -ge $MinLines) {
                $results += [pscustomobject]@{ path = $f.FullName; lines = $lineCount }
            }
        }
    }
    return $results
}

function Index-File {
    param([string]$Path)
    $entries = @()
    $lineNum = 0
    foreach ($line in Get-Content $Path -ErrorAction SilentlyContinue) {
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

$files = Get-SourceFiles -Roots @($frontendSrc, $backendRoot) -MinLines $lineThreshold | Sort-Object lines -Descending

$genCommit = "unknown"
try {
    $rev = (& git -C $repoRoot rev-parse --short HEAD 2>$null)
    if ($rev) { $genCommit = ($rev | Select-Object -First 1).Trim() }
} catch { $genCommit = "unknown" }

$sb = [System.Text.StringBuilder]::new()
[void]$sb.AppendLine("# CODE-MAP - large-file navigation index")
[void]$sb.AppendLine("")
[void]$sb.AppendLine("> Auto-generated by tools/gen-code-map.ps1 — index เฉพาะไฟล์ TS/TSX/Go ขนาด >= $lineThreshold บรรทัด ที่อยู่ใต้ frontend/src/ และ backend/ เท่านั้น (ข้าม node_modules, .next, vendor, dist, mock/mockdata, *.gen.*, backend/docs/docs.go และไม่รวม frontend/e2e/) — ไฟล์ที่ไม่อยู่ในแผนที่ไม่ได้แปลว่าเล็กกว่า $lineThreshold บรรทัดเสมอไป")
[void]$sb.AppendLine("> Read this FIRST before editing a big file so you can jump straight to the right line range")
[void]$sb.AppendLine("> instead of grep + re-read. Regenerate after major refactors.")
[void]$sb.AppendLine("> Generated: $(Get-Date -Format yyyy-MM-dd) @ commit $genCommit — ถ้า commit ปัจจุบันไม่ใช่อันนี้ ให้ถือว่าเลขบรรทัดอาจเลื่อนแล้ว ตรวจด้วย grep ก่อนใช้ หรือรัน tools/gen-code-map.ps1 ใหม่")
[void]$sb.AppendLine("")
[void]$sb.AppendLine("Files indexed: $($files.Count)")
[void]$sb.AppendLine("")

foreach ($f in $files) {
    $rel = $f.path.Replace($repoRoot + '\', '').Replace('\','/')
    [void]$sb.AppendLine("## $rel ($($f.lines) lines)")
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

[System.IO.File]::WriteAllText($outFile, $sb.ToString(), [System.Text.UTF8Encoding]::new($false))
Write-Host ("wrote {0} files -> {1} ({2} bytes)" -f $files.Count, $outFile, $sb.Length)
