[CmdletBinding()]
param(
    [string]$Prompt,
    [string]$PromptFile,
    [string]$ContextFile,
    [string]$Model = "glm-5.2",
    [string]$BaseUrl = "https://api.z.ai/api/coding/paas/v4",
    [string]$ApiKeyEnv = "ZAI_API_KEY",
    [ValidateSet("plan", "code", "review", "weakness", "all")]
    [string]$Mode = "plan",
    [int]$MaxTokens = 4096,
    [switch]$RawJson
)

$ErrorActionPreference = "Stop"

function Get-SecretFromEnv {
    param([string]$Name)
    $value = [Environment]::GetEnvironmentVariable($Name, "Process")
    if ([string]::IsNullOrWhiteSpace($value)) {
        $value = [Environment]::GetEnvironmentVariable($Name, "User")
    }
    if ([string]::IsNullOrWhiteSpace($value)) {
        $value = [Environment]::GetEnvironmentVariable($Name, "Machine")
    }
    return $value
}

function Read-OptionalFile {
    param([string]$Path)
    if ([string]::IsNullOrWhiteSpace($Path)) {
        return ""
    }
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
- identify source files/runtime evidence to inspect
- call out risks, edge cases, and verification commands
"@
        }
        "code" {
            return @"
Focus on implementation assistance:
- suggest the smallest safe code changes
- prefer local project patterns and backwards compatibility
- point out exact files/functions to inspect or patch
- do not output full rewritten files unless explicitly necessary
- do not invent APIs, schemas, or behavior
"@
        }
        "review" {
            return @"
Focus on code review:
- prioritize bugs, regressions, security, data correctness, and missing tests
- cite file/function/line references only when provided in the input context
- separate confirmed issues from questions or assumptions
- avoid style-only comments unless they affect maintainability or behavior
"@
        }
        "weakness" {
            return @"
Focus on Codex blind spots and weakness audit:
- identify assumptions Codex may be making
- look for missing source evidence, missing runtime checks, unsafe shortcuts, and untested edge cases
- call out accounting decimal, tenant scope, secret handling, migration, concurrency, performance, and rollback risks when relevant
- propose concrete checks that close each gap
"@
        }
        "all" {
            return @"
Cover all helper roles:
- PLAN: safe sequence and evidence to inspect
- CODE: smallest safe implementation suggestions
- REVIEW: likely bugs, regressions, security, and missing tests
- WEAKNESS: blind spots, assumptions, and verification gaps
"@
        }
    }
}

$apiKey = Get-SecretFromEnv -Name $ApiKeyEnv
if ([string]::IsNullOrWhiteSpace($apiKey)) {
    throw "Missing API key. Set `$env:$ApiKeyEnv before running this helper. Do not commit the key to repo files."
}

$promptText = $Prompt
if (-not [string]::IsNullOrWhiteSpace($PromptFile)) {
    $promptText = Read-OptionalFile -Path $PromptFile
}
if ([string]::IsNullOrWhiteSpace($promptText)) {
    throw "No prompt provided. Pass -Prompt or -PromptFile."
}

$contextText = Read-OptionalFile -Path $ContextFile
if (-not [string]::IsNullOrWhiteSpace($contextText)) {
    $promptText = @"
Task:
$promptText

Context:
$contextText
"@
}

$systemPrompt = @"
You are GLM 5.2 Think acting as a secondary AI helper for Codex in the BC Ai Account project.

Rules:
- Use thinking internally, but never expose hidden reasoning or chain-of-thought.
- Codex remains the source-of-truth analyst; local source code, docs, tests, runtime output, and database/API evidence win.
- Do not invent APIs, schemas, paths, business rules, or runtime facts.
- Return concise, actionable advice in Markdown.
- Help Codex close blind spots; do not replace Codex verification.

Mode: $Mode
$(Get-ModeInstruction -ModeName $Mode)

Output sections:
PLAN
CODE
REVIEW
WEAKNESS
RISKS
VERIFY
QUESTIONS
"@

$messages = @(
    @{
        role = "system"
        content = $systemPrompt
    },
    @{
        role = "user"
        content = $promptText
    }
)

$payload = @{
    model = $Model
    messages = $messages
    stream = $false
    max_tokens = $MaxTokens
    thinking = @{
        type = "enabled"
        clear_thinking = $true
    }
}

$jsonBody = $payload | ConvertTo-Json -Depth 20
$chatUrl = $BaseUrl.TrimEnd("/") + "/chat/completions"
$headers = @{
    Authorization = "Bearer $apiKey"
    "Content-Type" = "application/json"
}

try {
    $response = Invoke-RestMethod -Method Post -Uri $chatUrl -Headers $headers -Body $jsonBody -TimeoutSec 180
    $choice = $response.choices[0]
    $message = $choice.message
    $content = [string]$message.content
    $reasoning = [string]$message.reasoning_content

    $result = [ordered]@{
        success = $true
        model = $response.model
        mode = $Mode
        finishreason = $choice.finish_reason
        content = $content
        reasoningchars = $reasoning.Length
        usage = $response.usage
    }

    if ($RawJson) {
        $result | ConvertTo-Json -Depth 20
        return
    }

    Write-Output $content
    Write-Output ""
    Write-Output ("[glm52helper] model={0} mode={1} finish={2} reasoningchars={3} totaltokens={4}" -f $result.model, $result.mode, $result.finishreason, $result.reasoningchars, $response.usage.total_tokens)
} catch {
    $message = $_.Exception.Message
    if ($_.ErrorDetails -and $_.ErrorDetails.Message) {
        $message = $_.ErrorDetails.Message
    }

    $statusCode = $null
    $reason = $null
    if ($_.Exception.Response) {
        try {
            $statusCode = [int]$_.Exception.Response.StatusCode
            $reason = [string]$_.Exception.Response.ReasonPhrase
            if ([string]::IsNullOrWhiteSpace($reason)) {
                $reason = [string]$_.Exception.Response.StatusDescription
            }
        } catch {
            $statusCode = $null
        }
    }

    if ($statusCode -eq 429 -or $message -match "429|Too Many Requests") {
        Write-Output "[glm52helper] success=false model=$Model mode=$Mode error=rate_limited_or_quota_exhausted status=429 message=""GLM 5.2 quota/rate limit is currently exhausted; continue with Codex-only fallback until quota resets."""
        exit 2
    }

    $safeMessage = $message -replace "`r|`n", " "
    if ($statusCode) {
        Write-Output ("[glm52helper] success=false model={0} mode={1} status={2} reason=""{3}"" message=""{4}""" -f $Model, $Mode, $statusCode, $reason, $safeMessage)
    } else {
        Write-Output ("[glm52helper] success=false model={0} mode={1} message=""{2}""" -f $Model, $Mode, $safeMessage)
    }
    exit 1
}
