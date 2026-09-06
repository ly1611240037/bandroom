# BandRoom 数据库说明

数据库使用 SQLite，默认文件为 `backend/data/bandroom.db`（以启动时的当前目录和 `BANDROOM_DATABASE` 为准）。迁移文件位于 `backend/internal/db/migrations/`。

主要数据表：

| 表 | 用途 |
| --- | --- |
| `users` | 老板和顾客账号、手机号、邮箱验证状态 |
| `sessions` | HttpOnly 登录会话 |
| `membership_plans` | 月卡和次数卡方案 |
| `membership_cards` | 顾客线下开通的会员卡、有效期和剩余次数 |
| `rooms` / `room_photos` | 排练房资料、状态和照片 |
| `fixed_equipment` | 房间内固定设备 |
| `public_equipment` | 可额外借用的公共设备和库存 |
| `weekly_schedules` / `closures` | 营业时间和临时闭店 |
| `bookings` / `booking_equipment` | 预约主记录、清场占用区间和设备数量 |
| `notifications` | 站内通知及邮件发送状态 |
| `issue_reports` | 顾客报修和老板处理状态 |
| `venue_content` | 场馆名称、介绍、公告和会员说明 |
| `audit_logs` | 老板操作的对象、动作和前后数据 |

预约的 `occupied_until` 比实际排练结束时间晚 30 分钟，房间和公共设备冲突均使用这个占用区间判断。数据库迁移可重复执行，演示数据由 `go run ./cmd/seed` 幂等初始化。
