$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot

Push-Location "$projectRoot/backend"
try {
    $unformatted = @(gofmt -l .)
    if ($LASTEXITCODE -ne 0) { throw 'Go format check failed.' }
    if ($unformatted.Count -gt 0) { throw "Run gofmt on: $($unformatted -join ', ')" }
    go vet ./...
    if ($LASTEXITCODE -ne 0) { throw 'Go vet failed.' }
    go test -count=1 ./...
    if ($LASTEXITCODE -ne 0) { throw 'Go tests failed.' }
} finally {
    Pop-Location
}

Push-Location "$projectRoot/frontend"
try {
    npm test
    if ($LASTEXITCODE -ne 0) { throw 'Frontend tests failed.' }
    npm run build
    if ($LASTEXITCODE -ne 0) { throw 'Frontend build failed.' }
} finally {
    Pop-Location
}
