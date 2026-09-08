# Project Reliability Specification

## Purpose

为 BandRoom 建立可持续的开发约定与可靠性验收基线，确保预约选择、门店时间、数据库连接约束和认证导航符合既有业务规则，并使后续开发能够通过可重复的检查及时发现回归问题。

## Requirements

### Requirement: Persistent development guidelines
项目 SHALL 保存固定版本 skill，通过根规则与 OpenSpec 接入四项原则、业务约束与验证命令。

#### Scenario: Contributor starts work
- **WHEN** 开发者读取项目规则
- **THEN** 能找到 skill 原文、来源版本、项目限制和验证入口

### Requirement: Reliable booking selection
系统 SHALL 区分不同结束时间的时段，清除已失效选择并忽略过期响应。

#### Scenario: Customer changes filters
- **WHEN** 顾客切换日期或房间
- **THEN** 旧选择被清除，只展示当前条件的结果

### Requirement: Availability and database integrity
系统 SHALL 按 UTC+8 校验营业时间与日期筛选，以固定次数查询计算候选，并在每条数据库连接开启外键。

#### Scenario: Unavailable room
- **WHEN** 房间维护或停用
- **THEN** 不返回可预约时段

#### Scenario: Multiple connections
- **WHEN** 连接池创建额外连接
- **THEN** 新连接同样拒绝外键违规写入

### Requirement: Authentication navigation
系统 SHALL 根据当前路由展示认证表单，读取重置链接令牌并保留成功提示。

#### Scenario: Reset link or browser navigation
- **WHEN** 用户打开重置链接或使用前进后退
- **THEN** 表单匹配当前路由，令牌自动填入，成功后可以登录

### Requirement: Repeatable verification
项目 SHALL 提供无缓存后端测试、前端测试和构建入口，失败时终止后续步骤。

#### Scenario: Time advances
- **WHEN** 将来运行预约测试
- **THEN** 用例不会因固定预约日期过期而失败
