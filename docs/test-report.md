# BandRoom 测试报告

本轮问题、验证结果与边界见 [optimization-report.md](optimization-report.md)。

## 执行命令

```powershell
cd backend
go vet ./...
go test -count=1 ./...

cd ../frontend
npm test
npm run build

cd ..
openspec validate --specs --strict --no-interactive
```

也可在根目录执行 `powershell -ExecutionPolicy Bypass -File ./scripts/check.ps1`。

## 覆盖内容

- `backend/internal/integration/e2e_test.go`：临时 SQLite 和真实 HTTP 链路。
- 顾客注册、邮箱验证、登录和 Cookie 会话。
- 老板创建会员方案、房间、公共设备及线下开卡。
- 有效预约、房间冲突、设备库存冲突。
- 顾客取消、次数卡恢复和自动完成。
- 老板统计接口和场馆内容管理接口。
- 各业务包的单元测试、数据库迁移测试和种子数据测试。

## 演示前检查

启动前端构建和 Go 测试均通过后，再使用 README 中的统一启动方式访问 `http://localhost:8080`。
