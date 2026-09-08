# BandRoom 前端页面路由

登录后系统会根据角色自动进入对应区域：

- `/login`：登录
- `/register`：注册
- `/customer/home`：顾客场馆首页
- `/customer/bookings`：顾客预约排练
- `/customer/history`：我的预约
- `/customer/membership`：我的会员卡
- `/customer/notifications`：通知
- `/owner`：老板数据概览
- `/owner/bookings`：预约管理
- `/owner/rooms`：排练房管理
- `/owner/equipment`：设备管理
- `/owner/membership`：会员卡管理
- `/owner/settings`：场馆设置

开发时可以使用 Vite 单独运行前端；正式演示或部署时，执行 `npm run build` 后由 Go 服务统一提供 `frontend/dist` 和 `/api` 接口。
