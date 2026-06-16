[CmdletBinding()]
param(
    [Parameter(Mandatory = $true, ParameterSetName = "Inline")]
    [string]$Prompt,
    [Parameter(ParameterSetName = "File")]
    [string]$PromptFile,
    [string]$ContextFile,
    [string]$Model = "gpt-5.5",
    [ValidateSet("low", "medium", "high")]
    [string]$Reasoning = "high",
    [ValidateSet("plan", "code", "review", "weakness", "all")]
    [string]$Mode = "all",
    [int]$TimeoutSec = 600,
    [switch]$Raw
)

# Codex Advisor helper - wraps `codex exec` (gpt-5.5 think) so GLM 5.2 / Claude
# can ask Codex a focused advisory question and feed Codex's answer back into
# GLM's synthesis step. Mirrors glm52-think-planner.ps1 (reverse direction).
#
# Subscription-only policy: ABORT if OPENAI_API_KEY is set (no per-token billing).

$ErrorActionPreference = "Stop"

function Read-OptionalFile {
    param([string]$Path)
    if ([string]::IsNullOrWhiteSpace($Path)) { return "" }
    if (-not (Test-Path -LiteralPath $Path)) {
        throw "File not found: $Path"
    }
    return Get-Content -LiteralPath $Path -Raw
}

function Get-ModeInstruction {
    param([string]$ModeName)
    switch ($ModeName) {
        "plan" {
            return @"
Focus on planning:
- propose a safe sequence of work
- identify source files / runtime evidence to inspect (path:line when possible)
- call out risks, edge cases, and verification commands
"@
        }
        "code" {
            return @"
Focus on implementation assistance:
- suggest the smallest safe code change shape
- prefer existing local project patterns and backwards compatibility
- point out exact files/functions to inspect or patch
- do not invent APIs, schemas, paths, or runtime behavior
"@
        }
        "review" {
            return @"
Focus on code review:
- prioritize bugs, regressions, security, data correctness, multi-tenant scope, accounting-decimal safety
- cite file/function/line references only when present in the input context
- separate confirmed issues from questions or assumptions
- skip style-only comments unless they affect correctness
"@
        }
        "weakness" {
            return @"
Focus on GLM/Claude blind spots:
- identify assumptions the caller may be making
- look for missing source evidence, missing runtime checks, unsafe shortcuts, untested edge cases
- call out accounting decimal, tenant scope (holdingcode), secret handling, migration, concurrency, performance, and rollback risks when relevant
- propose concrete checks that close each gap
"@
        }
        "all" {
            return @"
Cover all helper roles:
- PLAN: safe sequence and evidence to inspect
- CODE: smallest safe implementation suggestions
- REVIEW: likely bugs, regressions, security, multi-tenant, missing tests
- WEAKNESS: blind spots, assumptions, and verification gaps
"@
        }
    }
}

# Subscription-only guard - same policy as dispatch-codex.sh.
$openaiKey = [Environment]::GetEnvironmentVariable("OPENAI_API_KEY", "Process")
if ([string]::IsNullOrWhiteSpace($openaiKey)) {
    $openaiKey = [Environment]::GetEnvironmentVariable("OPENAI_API_KEY", "User")
}
if (-not [string]::IsNullOrWhiteSpace($openaiKey)) {
    throw "OPENAI_API_KEY is set - policy is subscription-only (codex login). Unset it before running this helper."
}

# Resolve prompt text.
$promptText = $Prompt
if (-not [string]::IsNullOrWhiteSpace($PromptFile)) {
    $promptText = Read-OptionalFile -Path $PromptFile
}
if ([string]::IsNullOrWhiteSpace($promptText)) {
    throw "No prompt provided. Pass -Prompt or -PromptFile."
}

$contextText = Read-OptionalFile -Path $ContextFile

# Build the advisory system prompt so Codex stays in "advisor" mode (read-only,
# analysis only) even though codex exec defaults to workspace-write. We tell it
# explicitly to answer only, not edit files, because the caller (GLM/Claude)
# will synthesize and apply patches itself.
$systemPrompt = @"
You are Codex (gpt-5.5) acting as an ADVISORY helper for another AI (GLM 5.2 Think or Claude) in the BC Ai Account project (D:\bccode).

CRITICAL RULES:
- You are an ADVISOR in this call. Answer in Markdown. DO NOT edit files in this turn even though the sandbox allows it - the caller will synthesize your advice and apply patches itself.
- Read the active project source, docs, tests, runtime output, and database/API evidence first. Local evidence always wins.
- Do not invent APIs, schemas, paths, business rules, or runtime facts.
- Return concise, actionable advice.
- Honor BC Account rules: MongoDB-first operational CRUD, lowercase-no-underscore DB naming, Decimal128/numeric for accounting, holdingcode tenant scope, immutable guidfixed identity, no float for money.

Mode: $Mode
$(Get-ModeInstruction -ModeName $Mode)

Output sections (keep each short, skip a section if nothing useful):
PLAN
CODE
REVIEW
WEAKNESS
RISKS
VERIFY
QUESTIONS
"@

$finalUserMessage = if ([string]::IsNullOrWhiteSpace($contextText)) {
    $promptText
} else {
    @"
Task:
$promptText

Context (source evidence provided by caller):
$contextText
"@
}

# Write the system + user prompt to a temp prompt file that codex exec consumes
# as a single combined instruction. codex exec takes the prompt as the last
# positional arg; we compose system+user into one prompt to preserve role intent.
$combinedPrompt = @"
$systemPrompt

---

ADVISORY QUESTION FROM CALLER:

$finalUserMessage
"@

# codex exec invocation mirrors dispatch-codex.sh:
#   --skip-git-repo-check --ephemeral  : no session files, no git guard
#   -m gpt-5.5                          : subscription model
#   -s workspace-write                  : kept for parity, but advisor must NOT write
#   -c model_reasoning_effort=<level>   : reasoning knob (low/medium/high)
#   -c 'mcp_servers={}'                 : no MCP load (faster, no auth noise)
$codexArgs = @(
    "exec",
    "--skip-git-repo-check",
    "--ephemeral",
    "-m", $Model,
    "-s", "workspace-write",
    "-c", "model_reasoning_effort=$Reasoning",
    "-c", "mcp_servers={}",
    $combinedPrompt
)

Write-Host "[codex-advisor] model=$Model reasoning=$Reasoning mode=$Mode mcp=off" -ForegroundColor DarkGray

# Capture stdout. codex exec streams reasoning + final answer to stdout; we
# forward everything to the caller. The caller (GLM/Claude) parses the tail as
# the advisory answer.
$stdout = & codex @codexArgs 2>&1

if ($LASTEXITCODE -ne 0) {
    throw "codex exec failed with exit code $LASTEXITCODE. Output:`n$stdout"
}

if ($Raw) {
    $summary = [ordered]@{
        success  = $true
        model    = $Model
        reasoning = $Reasoning
        mode     = $Mode
        outputchars = ($stdout -join "`n").Length
    }
    $summary | ConvertTo-Json -Depth 10
    Write-Output ""
    Write-Output ($stdout -join "`n")
    return
}

Write-Output ($stdout -join "`n")
Write-Output ""
Write-Output ("[codex-advisor] model={0} reasoning={1} mode={2} outputchars={3}" -f $Model, $Reasoning, $Mode, ($stdout -join "`n").Length)
