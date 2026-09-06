$ErrorActionPreference = 'Stop'

$root = Split-Path -Parent $PSScriptRoot
$frontend = Join-Path $root 'frontend'
$backend = Join-Path $root 'backend'

Write-Host '==> 检查前端依赖'
Push-Location $frontend
try {
    if (-not (Test-Path (Join-Path $frontend 'node_modules'))) {
        npm ci
    }
    Write-Host '==> 构建 React 前端'
    npm run build
}
finally {
    Pop-Location
}

Write-Host '==> 初始化演示数据库并启动 BandRoom'
Push-Location $backend
try {
    $env:BANDROOM_DATABASE = './data/bandroom.db'
    $env:BANDROOM_APP_URL = 'http://localhost:8080'
    $env:BANDROOM_FRONTEND_DIST = '../frontend/dist'
    go run ./cmd/seed
    go run ./cmd/server
}
finally {
    Pop-Location
}
