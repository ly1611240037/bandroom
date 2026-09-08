## Purpose

为顾客预约和老板经营管理提供分区清晰、状态明确的界面体验，同时保留现有预约、会员卡、设备和通知业务行为。

## ADDED Requirements

### Requirement: Venue home overview
顾客端场馆首页 SHALL 展示场馆介绍、公告、营业时间、排练房概览、可借用设备概览和会员卡状态，并提供进入预约页面的明确入口。

#### Scenario: Customer views venue home
- **WHEN** 顾客打开场馆首页
- **THEN** 页面展示场馆信息和可用资源概览，不直接堆叠完整预约表单

#### Scenario: Customer starts booking
- **WHEN** 顾客点击“立即预约”
- **THEN** 系统进入独立的预约排练页面

### Requirement: Sectioned booking experience
预约排练页面 SHALL 将日期、房间、时段、联系人、设备和预约提交组织为清晰的分区，并提供实时预约摘要。

#### Scenario: Customer selects booking details
- **WHEN** 顾客选择日期、排练房、可用时段和额外设备
- **THEN** 页面更新预约摘要，展示当前选择的完整内容

#### Scenario: Customer submits valid booking
- **WHEN** 顾客填写有效信息并提交预约
- **THEN** 系统调用现有预约接口，显示成功反馈并保留现有业务校验结果

#### Scenario: Customer sees booking error
- **WHEN** 预约因冲突、会员卡或其他业务规则失败
- **THEN** 页面显示清晰的失败原因，并保留用户已填写的内容

### Requirement: Grouped customer booking history
我的预约页面 SHALL 将预约按即将到来和历史预约分组，并以卡片形式展示时间、房间、乐队、设备和预约状态。

#### Scenario: Customer views upcoming booking
- **WHEN** 顾客打开我的预约页面
- **THEN** 即将到来的预约显示在独立分组中，并为可取消预约提供取消入口

#### Scenario: Customer views historical bookings
- **WHEN** 顾客查看历史预约
- **THEN** 已完成、已取消和未到场预约按状态清晰标识

### Requirement: Owner overview dashboard
老板数据概览页面 SHALL 展示今日预约数、有效会员卡、可用排练房、待处理报修、今日预约时间线、最近预约和快捷操作。

#### Scenario: Owner views overview
- **WHEN** 老板进入数据概览
- **THEN** 页面优先展示经营摘要和今日安排，而不是所有管理表单

#### Scenario: Owner uses quick action
- **WHEN** 老板点击新增排练房、新增设备或开通会员卡等快捷操作
- **THEN** 系统进入对应管理功能，并保留现有接口行为

### Requirement: Independent owner management pages
老板后台 SHALL 将预约管理、排练房管理、设备管理、会员卡管理和场馆设置分别组织为独立页面。

#### Scenario: Owner opens management feature
- **WHEN** 老板从侧边栏选择一个管理模块
- **THEN** 内容区只展示该模块的筛选、列表、表单和操作

#### Scenario: Owner confirms destructive action
- **WHEN** 老板执行取消预约、停用设备或停用会员方案等重要操作
- **THEN** 系统先显示二次确认，确认后才调用现有操作接口

### Requirement: Consistent feedback and notifications
系统 SHALL 统一显示加载、成功和失败状态，并在通知页面按未读和已读分组，支持将通知标记为已读。

#### Scenario: Operation feedback
- **WHEN** 用户提交表单或执行操作
- **THEN** 操作按钮显示处理中状态，完成后显示成功或可理解的失败提示，并避免重复提交

#### Scenario: Notification grouping
- **WHEN** 用户打开通知页面
- **THEN** 页面分别展示未读和已读通知，并提供标记已读操作

