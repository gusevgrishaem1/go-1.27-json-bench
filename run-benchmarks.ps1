param(
    [string]$Benchtime = '2s',
    [int]$Count = 5
)

$ErrorActionPreference = 'Stop'
if ($Count -lt 1) { throw 'Count must be positive' }
$savedExperiment = $env:GOEXPERIMENT
$savedCache = $env:GOCACHE
$results = Join-Path $PSScriptRoot 'benchmark-results'
New-Item -ItemType Directory -Force -Path $results | Out-Null
$env:GOCACHE = Join-Path $PSScriptRoot '.go-build-cache'
try {
    $metadata = @(
        (go version),
        "OS=$([System.Environment]::OSVersion) processors=$([System.Environment]::ProcessorCount)",
        "benchtime=$Benchtime count=$Count",
        'Small=1 item; Large=1000 items; Encode=Response; Decode=fresh Request'
    )
    $metadata | Set-Content (Join-Path $results 'environment.txt') -Encoding utf8
    foreach ($variant in @(
        @{ Name = 'go127-v1'; Module = 'go127-v1'; Experiment = 'jsonv2' },
        @{ Name = 'go127-v1-nojsonv2'; Module = 'go127-v1'; Experiment = 'nojsonv2' },
        @{ Name = 'go127-v2'; Module = 'go127-v2'; Experiment = 'jsonv2' }
    )) {
        $env:GOEXPERIMENT = $variant.Experiment
        Push-Location (Join-Path $PSScriptRoot $variant.Module)
        try {
            Write-Host "$($variant.Name): GOEXPERIMENT=$env:GOEXPERIMENT"
            go test ./... -run '^$' -bench '^BenchmarkJSON$' -benchmem "-benchtime=$Benchtime" "-count=$Count" 2>&1 |
                Tee-Object -FilePath (Join-Path $results "$($variant.Name).txt")
            if ($LASTEXITCODE -ne 0) { throw "Benchmark failed: $($variant.Name)" }
        } finally { Pop-Location }
    }
} finally {
    $env:GOEXPERIMENT = $savedExperiment
    $env:GOCACHE = $savedCache
}
