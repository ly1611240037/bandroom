## Purpose

改进 BandRoom 前端的预约选择、历史记录与通知反馈，以具体房间和设备的演示场景帮助用户理解门店资源，并维持现有接口、业务约束以及桌面和移动端一致的操作方式。

## ADDED Requirements

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
