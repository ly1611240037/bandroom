# BandRoom 开发规则

开始工作前阅读 README.md、docs/architecture.md 和相关 OpenSpec 规格。
所有代码编写、审查和重构均应用 `.agents/skills/karpathy-guidelines/SKILL.md`：
先理解需求、保持简单、精确修改、验证目标。这是开发约定，不是运行时依赖。

## 业务边界

- React + Go + SQLite，保留现有 HTTP/service/db 分层与界面风格。
- 单门店、单老板；没有明确需求不引入支付、多门店、员工权限或新框架。
- 门店时间 UTC+8；预约半小时对齐、0.5–3 小时，另有 30 分钟清场。
- 创建预约在事务中校验会员、未来预约、房间和公共设备库存，并扣次。
- 保留规格允许的老板例外调整权限；改变业务规则前先更新规格。
- 已同步主规格在 openspec/specs/；尚未同步的历史规格在 openspec/changes/archive/，当前变更在 openspec/changes/。

## 修改与验证

- 每处修改对应需求或已确认问题；不做无关格式化、抽象或依赖升级。
- 小的实现选择自行决定；影响业务的歧义需澄清。保留用户已有改动。
- 前端处理异步失败与过期响应；SQL 参数绑定，约束覆盖每条连接。
- 不提交凭据、数据库、备份和构建输出；测试不得依赖固定日期仍在未来。
- 根目录运行 `powershell -ExecutionPolicy Bypass -File ./scripts/check.ps1`。
- 有规格变更另运行 `openspec validate <change> --type change --strict --no-interactive`。
- 页面交互修改需验证受影响流程；交付说明实际结果与验证限制。
