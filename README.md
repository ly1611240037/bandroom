# BandRoom 音乐排练空间

乐队排练室会员预约与设备管理系统。

## 项目简介

BandRoom 面向单门店、多排练房场景，支持顾客使用月卡或次数卡预约排练房，并支持房间固定设备、门店公共设备、预约冲突、会员管理和老板后台管理。

本项目是大学生演示及软件著作权申请项目，采用 OpenSpec 管理需求、设计和实现任务。

## 核心功能

- 顾客注册、邮箱验证、登录和密码重置
- 月卡与次数卡会员管理
- 多排练房预约和 30 分钟清场缓冲
- 固定设备和公共设备库存管理
- 预约取消、爽约、自动完成和邮件通知
- 老板后台、预约日历、统计、CSV 导出和操作日志
- 门店首页、房间照片、公告和会员办理说明

## 技术规划

- 前端：React
- 后端：Go
- 数据库：SQLite
- 开发方式：前后端分离
- 演示方式：Go 统一提供前端页面和后端接口

## 项目文档

OpenSpec 规划文件位于：

```text
openspec/changes/build-bandroom-booking-system/
```

其中包括项目提案、功能规格、技术设计和开发任务清单。

配套材料位于 `docs/`：

- `architecture.md`：系统架构说明
- `database.md`：数据库表和预约占用规则
- `user-manual.md`：顾客与老板操作手册
- `test-report.md`：测试范围与执行命令
- `demo-script.md`：导师演示脚本和截图清单

## 当前范围

v1.0 仅面向单个门店和一个老板管理员，不包含在线支付、独立 App、多门店和员工权限。

## 本地启动

环境要求：Go 1.22+、Node.js 18+ 和 npm。以下命令以 PowerShell 为例。

### 开发模式（前后端分开）

先启动后端：

```powershell
cd backend
$env:BANDROOM_DATABASE = "./data/bandroom.db"
$env:BANDROOM_APP_URL = "http://localhost:5173"
go run ./cmd/seed
go run ./cmd/server
```

另开一个终端启动前端：

```powershell
cd frontend
npm install
npm run dev
```

打开 <http://localhost:5173>。Vite 会把 `/api` 请求代理到 Go 的 8080 端口。

### 演示模式（Go 提供一个地址）

先构建 React，再由 Go 同时提供页面和 API：

```powershell
cd frontend
npm install
npm run build

cd ../backend
$env:BANDROOM_DATABASE = "./data/bandroom.db"
$env:BANDROOM_APP_URL = "http://localhost:8080"
$env:BANDROOM_FRONTEND_DIST = "../frontend/dist"
go run ./cmd/seed
go run ./cmd/server
```

打开 <http://localhost:8080>。`BANDROOM_FRONTEND_DIST` 也可以指向其他 React 构建目录。

### 演示账号和测试邮件

运行 seed 后可使用：

- 老板：`owner@bandroom.test` / `demo123456`
- 顾客：`customer@bandroom.test` / `demo123456`

开发环境邮件默认使用日志邮件适配器，不会发送真实邮件。注册验证、密码重置、预约通知等链接会打印在 Go 后端终端中；正式邮件服务可通过后续邮件适配器配置接入。

### 数据备份与恢复

停止 Go 服务后，复制 `BANDROOM_DATABASE` 指向的 SQLite 文件即可备份。例如：

```powershell
Copy-Item .\data\bandroom.db .\data\bandroom.db.bak
Copy-Item .\data\bandroom.db.bak .\data\bandroom.db
```

演示场景建议依次展示：游客浏览房间 → 顾客登录查看会员卡 → 选择日期和时段预约 → 借用公共设备 → 顾客取消预约 → 老板查看预约、统计和审计日志。

## 许可证

当前未指定开源许可证。
