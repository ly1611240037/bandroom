$ErrorActionPreference = 'Stop'
$projectRoot = Split-Path -Parent $PSScriptRoot
$previousEnv = @{}
foreach ($name in @('BANDROOM_DATABASE', 'BANDROOM_APP_URL', 'BANDROOM_FRONTEND_DIST')) {
    $previousEnv[$name] = [Environment]::GetEnvironmentVariable($name, 'Process')
}
Push-Location $projectRoot
try {
    Set-Location "$projectRoot/frontend"
    npm run build
    if ($LASTEXITCODE -ne 0) { throw 'Frontend build failed; demo was not started.' }

    Set-Location "$projectRoot/backend"
    $env:BANDROOM_DATABASE = './data/bandroom.db'
    $env:BANDROOM_APP_URL = 'http://localhost:8080'
    $env:BANDROOM_FRONTEND_DIST = '../frontend/dist'
    go run ./cmd/seed
    if ($LASTEXITCODE -ne 0) { throw 'Demo seed failed; server was not started.' }
    go run ./cmd/server
    if ($LASTEXITCODE -ne 0) { throw 'Demo server failed.' }
} finally {
    foreach ($name in $previousEnv.Keys) {
        if ($null -eq $previousEnv[$name]) {
            Remove-Item -LiteralPath "Env:$name" -ErrorAction SilentlyContinue
        } else {
            [Environment]::SetEnvironmentVariable($name, $previousEnv[$name], 'Process')
        }
    }
    Pop-Location
}
