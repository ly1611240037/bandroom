# BandRoom 音乐排练空间

乐队排练室会员预约与设备管理系统。

## 项目简介

BandRoom 面向单门店、多排练房场景，支持顾客使用月卡或次数卡预约排练房，并支持房间固定设备、门店公共设备、预约冲突、会员管理和老板后台管理。

本项目是大学生演示及软件著作权申请项目，采用 OpenSpec 管理需求、设计和实现任务。

## 代码仓库

- GitHub：<https://github.com/ly1611240037/bandroom>（远端 `origin`）
- Gitee：<https://gitee.com/ly1611240037/bandroom>（远端 `gitee`）

## 核心功能

- 顾客注册、邮箱验证、登录和密码重置
- 月卡与次数卡会员管理
- 多排练房预约和 30 分钟清场缓冲
- 固定设备和公共设备库存管理
- 预约取消、爽约、自动完成和邮件通知
- 老板后台、预约日历、统计、CSV 导出和操作日志
- 门店首页、房间照片、公告和会员办理说明

## 技术栈

- 前端：React
- 后端：Go
- 数据库：SQLite
- 开发方式：前后端分离
- 演示方式：Go 统一提供前端页面和后端接口

## 项目文档

OpenSpec 规划文件位于：

```text
openspec/changes/archive/2026-09-06-build-bandroom-booking-system/
```

其中包括项目提案、功能规格、技术设计和开发任务清单。后续历史变更见 `openspec/changes/archive/`，当前变更见 `openspec/changes/`。

开发规则入口为 [AGENTS.md](AGENTS.md)，其中接入了固定版本的 [Karpathy Guidelines](.agents/skills/karpathy-guidelines/SKILL.md)，来源版本见同目录 `SOURCE.md`。

配套材料位于 `docs/`：

- `architecture.md`：系统架构说明
- `database.md`：数据库表和预约占用规则
- `user-manual.md`：顾客与老板操作手册
- `test-report.md`：测试范围与执行命令
- `demo-script.md`：导师演示脚本和截图清单

## 当前范围

v1.0 仅面向单个门店和一个老板管理员，不包含在线支付、独立 App、多门店和员工权限。

## 本地启动

环境要求：Go 1.26+（与 `backend/go.mod` 一致）、Node.js 22.12+ 和 npm。以下命令以 PowerShell 为例。

### 一条命令启动演示版

在项目根目录执行：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\start-demo.ps1
```

首次使用请先在 `frontend` 目录执行一次 `npm install`。之后脚本会构建 React、初始化演示数据，并启动统一的 Go 服务。打开 <http://localhost:8080> 即可。按 `Ctrl+C` 停止服务。

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

`seed` 用于本地演示，会创建测试账号、房间和会员数据。测试账号定义见 `backend/internal/seed/seed.go`，仅用于隔离的开发环境；公开部署应使用独立账号和密码。重复执行 `seed` 会重置测试账号密码。

开发环境邮件默认使用日志邮件适配器，不会发送真实邮件。注册验证、密码重置、预约通知等链接会打印在 Go 后端终端中；正式邮件服务可通过后续邮件适配器配置接入。

### 数据备份与恢复

停止 Go 服务后，复制 `BANDROOM_DATABASE` 指向的 SQLite 文件即可备份。例如：

```powershell
Copy-Item .\data\bandroom.db .\data\bandroom.db.bak
Copy-Item .\data\bandroom.db.bak .\data\bandroom.db
```

演示场景建议依次展示：游客浏览房间 → 顾客登录查看会员卡 → 选择日期和时段预约 → 借用公共设备 → 顾客取消预约 → 老板查看预约、统计和审计日志。

## 开发验证

在根目录执行：

```powershell
powershell -ExecutionPolicy Bypass -File .\scripts\check.ps1
```

依次检查 Go 格式、静态分析、无缓存测试、前端测试和生产构建；任一步失败即停止。

## 部署说明

前端执行 `npm ci` 和 `npm run build` 后，将构建产物交给 Go 服务提供；后端在 `backend` 目录执行 `go build -o bin/server ./cmd/server` 构建。

运行配置通过环境变量传入，字段示例见 `.env.example`。程序不会自动加载该文件，需要由 shell 或进程管理器注入。

`deploy/` 提供通用的 systemd、Nginx 和环境变量模板。模板中的 `example.com`、端口和安装路径需按自己的环境调整，并保持 Nginx 上游端口与 `BANDROOM_HTTP_ADDR` 一致。

公开部署时配置 HTTPS，并将 `BANDROOM_APP_URL` 设为实际访问地址。真实环境配置、账号凭据、数据库及备份应保存在仓库之外。

## 许可证

当前未指定开源许可证。
