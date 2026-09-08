## Purpose

为 BandRoom 提供清晰分离的认证页、顾客端和老板后台，使不同角色能够在桌面和移动设备上使用易于理解的导航与页面布局。

## ADDED Requirements

### Requirement: Role-based application areas
系统 SHALL 根据已登录用户的角色，将顾客和老板分别带入对应的应用区域，并阻止用户访问不属于其角色的页面。

#### Scenario: Customer enters customer area
- **WHEN** 已登录用户的角色为顾客
- **THEN** 系统将用户带到顾客端默认页面“预约排练”

#### Scenario: Owner enters owner area
- **WHEN** 已登录用户的角色为老板
- **THEN** 系统将用户带到老板后台默认页面“数据概览”

#### Scenario: Unauthenticated user opens protected page
- **WHEN** 未登录用户访问顾客端或老板后台页面
- **THEN** 系统将用户带到登录页，并保留可返回的目标地址

### Requirement: Separate navigation structures
系统 SHALL 为认证页、顾客端和老板后台提供互不混杂的页面布局与导航。

#### Scenario: Customer navigation
- **WHEN** 顾客查看顾客端页面
- **THEN** 侧边栏提供场馆首页、预约排练、我的预约、会员卡和通知入口

#### Scenario: Owner navigation
- **WHEN** 老板查看后台页面
- **THEN** 侧边栏提供数据概览、预约管理、排练房、设备管理、会员卡管理和场馆设置入口

#### Scenario: Navigation selection
- **WHEN** 用户进入某个功能页面
- **THEN** 对应导航项显示当前选中状态，且内容区只展示该页面职责范围内的信息

### Requirement: Responsive navigation
系统 SHALL 优先适配桌面浏览器，并在窄屏设备上收起侧边栏，允许用户通过菜单按钮打开或关闭导航。

#### Scenario: Desktop layout
- **WHEN** 页面在桌面宽度下打开
- **THEN** 侧边栏固定显示，内容区保持可读宽度并避免横向滚动

#### Scenario: Narrow layout
- **WHEN** 页面在手机或窄屏宽度下打开
- **THEN** 侧边栏默认收起，菜单按钮可打开导航，选择页面后导航可以再次收起

### Requirement: Readable visual hierarchy
系统 SHALL 使用足够大的文字、清晰的层级和充足的间距呈现页面内容，并保持深色音乐空间的品牌视觉。

#### Scenario: Form readability
- **WHEN** 用户查看登录、注册或预约表单
- **THEN** 标签、输入内容、按钮和错误提示均清晰可读，且相邻控件不会拥挤重叠

#### Scenario: Feature navigation readability
- **WHEN** 用户查看侧边栏、统计卡片或管理列表
- **THEN** 文字与线性功能图标共同表达功能含义，当前状态具有明显视觉区分

