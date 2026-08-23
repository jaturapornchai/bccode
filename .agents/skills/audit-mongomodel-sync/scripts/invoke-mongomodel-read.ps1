[CmdletBinding()]
param(
    [Parameter(Mandatory = $true)]
    [ValidateSet(
        "list_projects",
        "get_project_context",
        "get_diagram",
        "get_workflow",
        "get_shared_brain",
        "list_revisions",
        "check_descriptions",
        "lint_model",
        "lint_workflows",
        "generate_code",
        "wait_for_project_change"
    )]
    [string]$Tool,

    [string]$ArgumentsJson = "{}",

    [ValidateRange(1, 60)]
    [int]$TimeoutSeconds = 15,

    [string]$Url = "http://localhost:3100/mcp"
)

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$uri = [Uri]$Url
if ($uri.Scheme -ne "http" -or $uri.Host -notin @("localhost", "127.0.0.1", "[::1]")) {
    throw "Refusing non-loopback MongoModel MCP URL: $Url"
}

try {
    $arguments = $ArgumentsJson | ConvertFrom-Json -AsHashtable
} catch {
    throw "ArgumentsJson must be valid JSON: $($_.Exception.Message)"
}

$payload = @{
    jsonrpc = "2.0"
    id = 1
    method = "tools/call"
    params = @{
        name = $Tool
        arguments = $arguments
    }
} | ConvertTo-Json -Depth 50 -Compress

$raw = & curl.exe `
    --max-time $TimeoutSeconds `
    --silent `
    --show-error `
    --no-buffer `
    --request POST `
    $Url `
    --header "Content-Type: application/json" `
    --header "Accept: application/json, text/event-stream" `
    --data-binary $payload

if ($LASTEXITCODE -ne 0) {
    throw "MongoModel MCP request failed with curl exit code $LASTEXITCODE"
}

$dataLines = @($raw | ForEach-Object {
    if ($_ -match '^data:\s*(.+)$') {
        $Matches[1]
    }
})
$responseText = if ($dataLines.Count -gt 0) {
    $dataLines[-1]
} else {
    ($raw -join [Environment]::NewLine).Trim()
}

if ([string]::IsNullOrWhiteSpace($responseText)) {
    throw "MongoModel MCP returned an empty response"
}

try {
    $response = $responseText | ConvertFrom-Json
} catch {
    throw "MongoModel MCP returned invalid JSON: $($_.Exception.Message)"
}

if ($response.PSObject.Properties.Name -contains "error" -and $null -ne $response.error) {
    throw ($response.error | ConvertTo-Json -Depth 20 -Compress)
}

$textBlocks = @($response.result.content | Where-Object { $_.type -eq "text" })
if ($textBlocks.Count -gt 0) {
    $textBlocks | ForEach-Object { $_.text }
} else {
    $response.result | ConvertTo-Json -Depth 50
}
