## Purpose

改进 BandRoom 前端的预约选择、历史记录与通知反馈，以具体房间和设备的演示场景帮助用户理解门店资源，并维持现有接口、业务约束以及桌面和移动端一致的操作方式。

## ADDED Requirements

### Requirement: Find a suitable rehearsal room
系统 SHALL 在首页提供按人数及房间、固定设备关键词组合筛选，展示匹配数量、清除条件及无结果提示，完整显示公共设备。

#### Scenario: Filter by capacity and equipment
- **WHEN** 用户输入人数与设备关键词
- **THEN** 仅显示容量足够且名称或可用固定设备匹配的房间，维护设备不作为设备匹配；清除条件恢复全部房间

### Requirement: Reliable JSON responses
系统 SHALL 将无效的 JSON API 成功响应展示为异常而非空数据，保留 HTTP 错误状态及取消信号。

#### Scenario: Server returns an HTML fallback
- **WHEN** JSON API 返回 HTML 或无效 JSON
- **THEN** 请求明确失败，页面不得将其解释为没有记录或操作成功

### Requirement: Refresh and batch read notifications
系统 SHALL 支持通知刷新、读取失败重试、北京时间展示和当前已加载通知批量已读。

#### Scenario: Batch read with concurrent arrivals
- **WHEN** 用户批量标记已加载通知且之后出现新消息
- **THEN** 服务端只更新当前用户且 ID 不超过已加载最大 ID 的未读通知，其他用户、新消息及已有阅读时间不受影响，重复操作可安全执行

#### Scenario: Refresh fails
- **WHEN** 刷新通知失败
- **THEN** 保留已有记录，显示错误与重试入口；加载期间禁止读取操作，过期响应不得覆盖新状态

### Requirement: Compact slot selection
系统 SHALL 按时长和开始时间所属的上午、下午、晚上分组展示可用时段，保留所有合法候选。

#### Scenario: Change time range
- **WHEN** 用户改变时长或时间范围
- **THEN** 页面只显示对应候选，清除旧选择，摘要与提交按钮保持一致

### Requirement: Accurate booking presentation
系统 SHALL 按北京时间显示记录，以中文展示接口状态，区分当前预约与历史记录，并防止状态徽标拉伸。

#### Scenario: Completed booking
- **WHEN** 接口返回已完成预约
- **THEN** 预约出现在历史分组，无取消按钮，显示实际排练时间和紧凑的已完成徽标

### Requirement: Notification empty groups
系统 SHALL 对未读与已读通知分别显示数量及明确空态，加载与失败状态不得混同为空记录。

#### Scenario: No unread notifications
- **WHEN** 有已读通知但没有未读通知
- **THEN** 未读分组显示没有未读通知的提示，已读记录仍可查看

### Requirement: Concrete rehearsal room demo
系统 SHALL 提供三种用途的房间演示资料、固定设备及布局示意，并连接现有预约流程。

#### Scenario: Inspect a room
- **WHEN** 用户查看场馆或选择房间
- **THEN** 能了解容量、用途、固定设备和状态，区分随房设备与公共借用设备；示意明确标注非实景

#### Scenario: Repeat demo initialization
- **WHEN** 重复执行演示初始化
- **THEN** 不重复添加相同固定设备，不覆盖用户已定制的房间描述

#### Scenario: Expand spatial layout
- **WHEN** 用户展开房间布局详情
- **THEN** 看到含地板、墙面与立体设备的等距示意，设备编号与清单对应，手机宽度下不横向溢出

#### Scenario: Drum kit inventory and legacy demo records
- **WHEN** 展示演示房间的鼓组，或迁移旧版重复初始化的设备数据
- **THEN** 鼓组按套计数并说明底鼓、军鼓、通鼓及镲片；只合并 B 房相同旧默认描述、数量一且可用的重复记录，保留定制及不同数量状态的设备
