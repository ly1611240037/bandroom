$root = Split-Path -Parent $PSScriptRoot

Set-Location "$root/frontend"
npm run build

Set-Location "$root/backend"
$env:BANDROOM_DATABASE = './data/bandroom.db'
$env:BANDROOM_APP_URL = 'http://localhost:8080'
$env:BANDROOM_FRONTEND_DIST = '../frontend/dist'
go run ./cmd/seed
go run ./cmd/server
