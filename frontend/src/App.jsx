import { useEffect, useState } from 'react'

const emptyForm = { name: '', email: '', phone: '', password: '' }

async function api(path, options = {}) {
  const response = await fetch(path, {
    credentials: 'include',
    headers: { 'Content-Type': 'application/json', ...(options.headers || {}) },
    ...options,
  })
  const body = await response.json().catch(() => ({}))
  if (!response.ok) throw new Error(body.error || '请求失败，请稍后再试')
  return body
}

function App() {
  const [view, setView] = useState('login')
  const [form, setForm] = useState(emptyForm)
  const [message, setMessage] = useState('')
  const [error, setError] = useState('')
  const [user, setUser] = useState(null)
  const [cards, setCards] = useState([])
  const [notifications, setNotifications] = useState([])
  const [rooms, setRooms] = useState([])
  const [equipment, setEquipment] = useState([])
  const [loading, setLoading] = useState(false)
  const params = new URLSearchParams(window.location.search)

  useEffect(() => {
    if (window.location.pathname === '/verify-email') {
      const token = params.get('token')
      if (!token) { setError('验证链接缺少 token'); return }
      api(`/api/auth/verify?token=${encodeURIComponent(token)}`)
        .then((data) => { setMessage(data.message); setView('login') })
        .catch((err) => setError(err.message))
    } else if (window.location.pathname === '/reset-password') {
      setView('reset')
      setForm((old) => ({ ...old, token: params.get('token') || '' }))
    }
    api('/api/auth/me').then((data) => setUser(data.user)).catch(() => {})
  }, [])

  useEffect(() => {
    if (user?.role === 'customer') api('/api/customer/membership/cards').then((data) => setCards(data.cards || [])).catch(() => setCards([]))
    else setCards([])
  }, [user])

  useEffect(() => {
    if (user) api('/api/notifications').then((data) => setNotifications(data.notifications || [])).catch(() => setNotifications([]))
    else setNotifications([])
  }, [user])

  useEffect(() => {
    Promise.all([api('/api/rooms'), api('/api/equipment')]).then(([roomData, equipmentData]) => {
      setRooms(roomData.rooms || []); setEquipment(equipmentData.equipment || [])
    }).catch(() => {})
  }, [])

  function updateField(event) {
    setForm((old) => ({ ...old, [event.target.name]: event.target.value }))
    setError('')
  }

  async function submit(event) {
    event.preventDefault()
    setLoading(true); setError(''); setMessage('')
    try {
      if (view === 'login') {
        const data = await api('/api/auth/login', { method: 'POST', body: JSON.stringify({ email: form.email, password: form.password }) })
        setUser(data.user); setMessage(`欢迎回来，${data.user.name}`)
      } else if (view === 'register') {
        const data = await api('/api/auth/register', { method: 'POST', body: JSON.stringify(form) })
        setMessage(data.message); setView('login'); setForm(emptyForm)
      } else if (view === 'forgot') {
        const data = await api('/api/auth/password-reset/request', { method: 'POST', body: JSON.stringify({ email: form.email }) })
        setMessage(data.message)
      } else {
        const data = await api('/api/auth/password-reset/confirm', { method: 'POST', body: JSON.stringify({ token: form.token, password: form.password }) })
        setMessage(data.message); setView('login'); setForm(emptyForm)
      }
    } catch (err) { setError(err.message) } finally { setLoading(false) }
  }

  async function logout() {
    await api('/api/auth/logout', { method: 'POST' }).catch(() => {})
    setUser(null); setCards([]); setNotifications([]); setMessage('已退出登录')
  }

  const titles = { login: '登录 BandRoom', register: '创建顾客账号', forgot: '找回密码', reset: '设置新密码' }
  const isReset = view === 'reset'

  return (
    <main className="shell">
      <div className="ambient ambient-one" /><div className="ambient ambient-two" />
      <section className="brand-panel">
        <p className="eyebrow">BANDROOM · MUSIC SPACE</p>
        <h1>让每一次排练，<span>准时开始。</span></h1>
        <p className="lead">多房间预约、会员卡与排练设备，一处管理你的乐队时光。</p>
        <div className="feature-list"><span>01　灵活预约</span><span>02　设备借用</span><span>03　会员专属</span></div>
        <Catalog rooms={rooms} equipment={equipment} />
      </section>
      <section className="auth-card">
        <div className="card-top"><span className="record-dot" /><span>REHEARSAL ROOM ACCESS</span></div>
        {user ? <LoggedIn user={user} cards={cards} notifications={notifications} setNotifications={setNotifications} rooms={rooms} equipment={equipment} onLogout={logout} /> : <>
          <div className="card-heading"><p className="eyebrow">ACCOUNT</p><h2>{titles[view]}</h2><p>{view === 'register' ? '注册后即可查看空闲房间并提交预约。' : '登录后管理你的排练预约与会员权益。'}</p></div>
          {message && <div className="notice success">{message}</div>}
          {error && <div className="notice error">{error}</div>}
          <form onSubmit={submit}>
            {view === 'register' && <><Field label="联系人姓名" name="name" value={form.name} onChange={updateField} required /><Field label="手机号" name="phone" value={form.phone} onChange={updateField} required /></>}
            {isReset && <Field label="重置令牌" name="token" value={form.token || ''} onChange={updateField} required />}
            {view !== 'reset' && <Field label="邮箱" name="email" type="email" value={form.email} onChange={updateField} required />}
            {view !== 'forgot' && <Field label={isReset ? '新密码' : '密码'} name="password" type="password" value={form.password} onChange={updateField} required />}
            {view === 'register' && <p className="hint">密码至少 8 位；注册后请从演示邮件中打开验证链接。</p>}
            <button className="primary-button" disabled={loading}>{loading ? '处理中…' : view === 'login' ? '登录' : view === 'register' ? '注册并验证邮箱' : view === 'forgot' ? '发送重置邮件' : '确认新密码'}</button>
          </form>
          <AuthLinks view={view} onChange={(next) => { setView(next); setError(''); setMessage(''); setForm(emptyForm) }} />
        </>}
      </section>
    </main>
  )
}

function Field({ label, name, type = 'text', value, onChange, required }) {
  return <label className="field"><span>{label}</span><input name={name} type={type} value={value} onChange={onChange} required={required} autoComplete={type === 'password' ? 'new-password' : name} /></label>
}

function AuthLinks({ view, onChange }) {
  if (view === 'reset') return <p className="switch-line"><button onClick={() => onChange('login')}>返回登录</button></p>
  return <div className="switch-line">{view === 'login' ? <><span>还没有账号？</span><button onClick={() => onChange('register')}>立即注册</button><button onClick={() => onChange('forgot')}>忘记密码</button></> : <button onClick={() => onChange('login')}>返回登录</button>}</div>
}

function LoggedIn({ user, cards, notifications, setNotifications, rooms, equipment, onLogout }) {
  return <div className="logged-in"><div className="avatar">{user.name.slice(0, 1)}</div><h2>你好，{user.name}</h2><p>{user.email}</p><div className="member-badge">{user.role === 'owner' ? '老板账户' : user.emailVerifiedAt ? '邮箱已验证' : '待验证'}</div><Notifications items={notifications} setItems={setNotifications} />{user.role === 'owner' ? <OwnerDashboard rooms={rooms} equipment={equipment} /> : <><MembershipCards cards={cards} /><BookingCenter user={user} rooms={rooms} equipment={equipment} /></>}<button className="text-button" onClick={onLogout}>退出登录</button></div>
}

function MembershipCards({ cards }) {
  if (!cards.length) return <div className="membership-empty">暂时没有会员卡，请联系老板线下开通。</div>
  return <div className="membership-list">{cards.map((card) => <div className="membership-card" key={card.id}><div><strong>{card.planName}</strong><span>{card.planType === 'monthly' ? '月卡' : '次数卡'}</span></div><b>{card.planType === 'monthly' ? `${card.startsAt.slice(0, 10)} — ${card.endsAt.slice(0, 10)}` : `剩余 ${card.remainingUses ?? 0} 次`}</b><small>{card.status === 'active' ? '当前可用' : card.status === 'scheduled' ? '待生效' : card.status === 'depleted' ? '已用完' : card.status}</small></div>)}</div>
}

function Notifications({ items, setItems }) {
  async function markRead(id) { await api(`/api/notifications/${id}/read`, { method: 'PATCH' }).catch(() => {}); setItems((old) => old.map((item) => item.id === id ? { ...item, readAt: new Date().toISOString() } : item)) }
  if (!items.length) return <div className="notification-empty">暂无站内通知</div>
  return <div className="notifications"><div className="booking-heading"><span>通知</span><small>{items.filter((item) => !item.readAt).length} 条未读</small></div>{items.slice(0, 3).map((item) => <button type="button" className={item.readAt ? 'notification read' : 'notification'} key={item.id} onClick={() => markRead(item.id)}><b>{item.title}</b><span>{item.body}</span></button>)}</div>
}

function OwnerDashboard({ rooms, equipment }) {
  const [bookings, setBookings] = useState([]); const [plans, setPlans] = useState([]); const [issues, setIssues] = useState([]); const [stats, setStats] = useState({}); const [audits, setAudits] = useState([]); const [filters, setFilters] = useState({ date: '', status: '', q: '' }); const [error, setError] = useState('')
  const refresh = () => { const query = new URLSearchParams(filters).toString(); return Promise.all([api(`/api/owner/bookings?${query}`), api('/api/owner/membership/plans'), api('/api/owner/issues'), api('/api/owner/stats'), api('/api/owner/audits')]).then(([bookingData, planData, issueData, statsData, auditData]) => { setBookings(bookingData.bookings || []); setPlans(planData.plans || []); setIssues(issueData.issues || []); setStats(statsData.stats || {}); setAudits(auditData.audits || []) }).catch((err) => setError(err.message)) }
  useEffect(() => { refresh() }, [filters.date, filters.status, filters.q])
  async function bookingAction(id, action) { await api(`/api/owner/bookings/${id}/${action}`, { method: 'POST', body: JSON.stringify({ reason: '老板后台操作' }) }).catch((err) => setError(err.message)); refresh() }
  async function issueAction(issue) { await api(`/api/owner/issues/${issue.id}`, { method: 'PATCH', body: JSON.stringify({ status: issue.status === 'open' ? 'in_progress' : 'resolved', equipmentStatus: issue.equipmentId ? 'maintenance' : '' }) }).catch((err) => setError(err.message)); refresh() }
  async function togglePlan(plan) { await api(`/api/owner/membership/plans/${plan.id}/status`, { method: 'PATCH', body: JSON.stringify({ isActive: !plan.isActive }) }).catch((err) => setError(err.message)); refresh() }
  return <div className="owner-dashboard"><div className="owner-stats"><div><b>{stats.totalBookings ?? bookings.length}</b><small>预约总数</small></div><div><b>{stats.confirmedBookings ?? 0}</b><small>进行中预约</small></div><div><b>{stats.activeCards ?? 0}</b><small>有效会员卡</small></div><div><b>{stats.maintenanceEquipment ?? 0}</b><small>维护设备</small></div></div>{error && <div className="notice error">{error}</div>}<div className="admin-filters"><input type="date" value={filters.date} onChange={(event) => setFilters({ ...filters, date: event.target.value })} /><select value={filters.status} onChange={(event) => setFilters({ ...filters, status: event.target.value })}><option value="">全部状态</option><option value="confirmed">已确认</option><option value="completed">已完成</option><option value="cancelled">已取消</option><option value="no_show">未到场</option></select><input placeholder="搜索乐队/手机号" value={filters.q} onChange={(event) => setFilters({ ...filters, q: event.target.value })} /><a href={`/api/owner/bookings.csv?${new URLSearchParams(filters).toString()}`}>导出预约 CSV</a></div><section className="admin-section"><div className="booking-heading"><span>预约日历 · 全部房间</span><small>按筛选结果排序</small></div>{bookings.slice(0, 8).map((booking) => <div className="admin-row" key={booking.id}><span><b>{booking.startsAt.slice(0, 16).replace('T', ' ')}</b><small>{booking.bandName} · 房间 {booking.roomId} · 会员卡 #{booking.membershipCardId} · 设备 {booking.equipment?.length || 0} 项 · {booking.notes || '无备注'}</small></span><em>{booking.status}</em>{booking.status === 'confirmed' && <span className="admin-actions"><button onClick={() => bookingAction(booking.id, 'no-show')}>未到场</button><button onClick={() => bookingAction(booking.id, 'cancel')}>取消</button></span>}</div>)}{!bookings.length && <small className="muted">暂无预约</small>}</section><section className="admin-section"><div className="booking-heading"><span>会员方案</span><small>{plans.length} 个方案</small></div>{plans.map((plan) => <div className="admin-row" key={plan.id}><span><b>{plan.name}</b><small>{plan.planType === 'monthly' ? '月卡' : `次数卡 · ${plan.includedUses} 次`}</small></span><em>{plan.isActive ? '启用' : '停用'}</em><button onClick={() => togglePlan(plan)}>{plan.isActive ? '停用' : '启用'}</button></div>)}</section><section className="admin-section"><div className="booking-heading"><span>报修处理</span><small>{issues.length} 条记录</small></div>{issues.slice(0, 4).map((issue) => <div className="admin-row" key={issue.id}><span><b>{issue.description}</b><small>报修编号 #{issue.id}</small></span><em>{issue.status}</em>{issue.status !== 'resolved' && <button onClick={() => issueAction(issue)}>处理</button>}</div>)}</section><section className="admin-section"><div className="booking-heading"><span>最近审计日志</span><small>{audits.length} 条</small></div>{audits.slice(0, 4).map((audit) => <div className="admin-row" key={audit.id}><span><b>{audit.action}</b><small>操作者 #{audit.actorUserId} · {audit.targetType} #{audit.targetId || '-'}</small></span><em>{audit.createdAt.slice(0, 16).replace('T', ' ')}</em></div>)}</section><OwnerManagement rooms={rooms} equipment={equipment} plans={plans} onRefresh={refresh} onError={setError} /></div>
}

function OwnerManagement({ rooms, equipment, plans, onRefresh, onError }) {
  const [room, setRoom] = useState({ name: '', description: '', capacity: 6 }); const [item, setItem] = useState({ name: '', description: '', quantity: 1 }); const [plan, setPlan] = useState({ planType: 'monthly', name: '', priceCents: 0, includedUses: 10 }); const [card, setCard] = useState({ userId: 1, planId: plans[0]?.id || '', paidAmountCents: 0, paymentMethod: '线下支付', notes: '' }); const [schedule, setSchedule] = useState({ weekday: 1, opensAt: '08:00', closesAt: '23:00', enabled: true }); const [closure, setClosure] = useState({ startsAt: '', endsAt: '', reason: '' }); const [content, setContent] = useState({ key: 'description', value: '面向乐队的多房间会员制排练空间' }); const [message, setMessage] = useState('')
  async function send(path, options, success) { try { await api(path, options); setMessage(success); onRefresh() } catch (err) { onError(err.message) } }
  return <section className="admin-section management"><div className="booking-heading"><span>管理操作</span><small>房间 · 设备 · 会员 · 排程 · 场馆信息</small></div><div className="management-grid"><AdminForm title="新增排练房" onSubmit={(event) => { event.preventDefault(); send('/api/owner/rooms', { method: 'POST', body: JSON.stringify({ ...room, capacity: Number(room.capacity) }) }, '房间已新增') }}><input placeholder="房间名称" value={room.name} onChange={(event) => setRoom({ ...room, name: event.target.value })} required /><input placeholder="描述" value={room.description} onChange={(event) => setRoom({ ...room, description: event.target.value })} /><input type="number" placeholder="容量" value={room.capacity} onChange={(event) => setRoom({ ...room, capacity: event.target.value })} /><button>新增房间</button></AdminForm><AdminForm title="新增公共设备" onSubmit={(event) => { event.preventDefault(); send('/api/owner/equipment', { method: 'POST', body: JSON.stringify({ ...item, quantity: Number(item.quantity), hourlyPriceCents: 0 }) }, '设备已新增') }}><input placeholder="设备名称" value={item.name} onChange={(event) => setItem({ ...item, name: event.target.value })} required /><input placeholder="描述" value={item.description} onChange={(event) => setItem({ ...item, description: event.target.value })} /><input type="number" placeholder="库存" value={item.quantity} onChange={(event) => setItem({ ...item, quantity: event.target.value })} /><button>新增设备</button></AdminForm><AdminForm title="新增会员方案" onSubmit={(event) => { event.preventDefault(); send('/api/owner/membership/plans', { method: 'POST', body: JSON.stringify({ ...plan, priceCents: Number(plan.priceCents), includedUses: plan.planType === 'count' ? Number(plan.includedUses) : null }) }, '会员方案已新增') }}><select value={plan.planType} onChange={(event) => setPlan({ ...plan, planType: event.target.value })}><option value="monthly">月卡</option><option value="count">次数卡</option></select><input placeholder="方案名称" value={plan.name} onChange={(event) => setPlan({ ...plan, name: event.target.value })} required /><input type="number" placeholder="价格（分）" value={plan.priceCents} onChange={(event) => setPlan({ ...plan, priceCents: event.target.value })} />{plan.planType === 'count' && <input type="number" placeholder="次数" value={plan.includedUses} onChange={(event) => setPlan({ ...plan, includedUses: event.target.value })} />}<button>新增方案</button></AdminForm><AdminForm title="线下开通会员卡" onSubmit={(event) => { event.preventDefault(); send('/api/owner/membership/cards', { method: 'POST', body: JSON.stringify({ ...card, userId: Number(card.userId), planId: Number(card.planId), paidAmountCents: Number(card.paidAmountCents) }) }, '会员卡已开通') }}><input type="number" placeholder="顾客 ID" value={card.userId} onChange={(event) => setCard({ ...card, userId: event.target.value })} /><select value={card.planId} onChange={(event) => setCard({ ...card, planId: event.target.value })}><option value="">选择方案</option>{plans.map((item) => <option value={item.id} key={item.id}>{item.name}</option>)}</select><input type="number" placeholder="实收金额（分）" value={card.paidAmountCents} onChange={(event) => setCard({ ...card, paidAmountCents: event.target.value })} /><input placeholder="支付方式/备注" value={card.paymentMethod} onChange={(event) => setCard({ ...card, paymentMethod: event.target.value })} /><button>开通会员卡</button></AdminForm><AdminForm title="修改营业时间" onSubmit={(event) => { event.preventDefault(); send(`/api/owner/schedule/${schedule.weekday}`, { method: 'PUT', body: JSON.stringify(schedule) }, '营业时间已更新') }}><input type="number" min="0" max="6" value={schedule.weekday} onChange={(event) => setSchedule({ ...schedule, weekday: event.target.value })} /><input value={schedule.opensAt} onChange={(event) => setSchedule({ ...schedule, opensAt: event.target.value })} /><input value={schedule.closesAt} onChange={(event) => setSchedule({ ...schedule, closesAt: event.target.value })} /><button>保存营业时间</button></AdminForm><AdminForm title="添加闭店时段" onSubmit={(event) => { event.preventDefault(); send('/api/owner/closures', { method: 'POST', body: JSON.stringify(closure) }, '闭店安排已添加') }}><input placeholder="开始时间 RFC3339" value={closure.startsAt} onChange={(event) => setClosure({ ...closure, startsAt: event.target.value })} /><input placeholder="结束时间 RFC3339" value={closure.endsAt} onChange={(event) => setClosure({ ...closure, endsAt: event.target.value })} /><input placeholder="原因" value={closure.reason} onChange={(event) => setClosure({ ...closure, reason: event.target.value })} /><button>添加闭店</button></AdminForm><AdminForm title="编辑场馆信息" onSubmit={(event) => { event.preventDefault(); send(`/api/owner/venue/${content.key}`, { method: 'PUT', body: JSON.stringify({ value: content.value }) }, '场馆信息已保存') }}><select value={content.key} onChange={(event) => setContent({ ...content, key: event.target.value })}><option value="name">场馆名称</option><option value="description">场馆介绍</option><option value="phone">联系电话</option></select><input placeholder="内容" value={content.value} onChange={(event) => setContent({ ...content, value: event.target.value })} /><button>保存场馆信息</button></AdminForm></div>{message && <div className="notice success">{message}</div>}</section>
}

function AdminForm({ title, onSubmit, children }) { return <form className="admin-form" onSubmit={onSubmit}><b>{title}</b>{children}</form> }

function Catalog({ rooms, equipment }) {
  return <div className="catalog"><div className="catalog-heading"><span>空间与设备</span><small>开放浏览 · 登录后预约</small></div><div className="catalog-grid">{rooms.slice(0, 3).map((room) => <div className="catalog-item" key={room.id}><strong>{room.name}</strong><span>{room.status === 'available' ? '可预约' : '暂不可用'} · {room.capacity ? `${room.capacity} 人` : '标准房'}</span><small>{room.fixedEquipment?.map((item) => item.name).join(' · ') || '基础排练设备'}</small></div>)}{equipment.slice(0, 3).map((item) => <div className="catalog-item equipment-item" key={`equipment-${item.id}`}><strong>{item.name}</strong><span>{item.status === 'available' ? `库存 ${item.quantity}` : '维护中'}</span><small>公共借用设备</small></div>)}</div></div>
}

function BookingCenter({ user, rooms, equipment }) {
  const [date, setDate] = useState(tomorrow())
  const [roomID, setRoomID] = useState('')
  const [slots, setSlots] = useState([])
  const [selectedSlot, setSelectedSlot] = useState(null)
  const [details, setDetails] = useState({ bandName: '', phone: user.phone || '', notes: '' })
  const [quantities, setQuantities] = useState({})
  const [bookings, setBookings] = useState([])
  const [bookingMessage, setBookingMessage] = useState('')
  const [bookingError, setBookingError] = useState('')
  const availableRooms = rooms.filter((room) => room.status === 'available')

  useEffect(() => { if (!roomID && availableRooms[0]) setRoomID(String(availableRooms[0].id)) }, [rooms])
  useEffect(() => {
    if (!roomID || !date) return
    setSelectedSlot(null); setBookingError('')
    api(`/api/availability?roomId=${roomID}&date=${date}`).then((data) => setSlots(data.slots || [])).catch((err) => setBookingError(err.message))
  }, [roomID, date])
  useEffect(() => { api('/api/customer/bookings').then((data) => setBookings(data.bookings || [])).catch(() => {}) }, [])

  async function submitBooking(event) {
    event.preventDefault(); setBookingMessage(''); setBookingError('')
    if (!selectedSlot) { setBookingError('请先选择一个可用时段'); return }
    try {
      const selectedEquipment = Object.entries(quantities).filter(([, quantity]) => Number(quantity) > 0).map(([equipmentId, quantity]) => ({ equipmentId: Number(equipmentId), quantity: Number(quantity) }))
      const data = await api('/api/customer/bookings', { method: 'POST', body: JSON.stringify({ roomId: Number(roomID), ...details, startsAt: selectedSlot.startsAt, endsAt: selectedSlot.endsAt, equipment: selectedEquipment }) })
      setBookingMessage(`预约成功：${data.booking.startsAt.slice(0, 16).replace('T', ' ')} 开始`); setBookings((old) => [data.booking, ...old]); setSelectedSlot(null)
    } catch (err) { setBookingError(err.message) }
  }

  async function cancelBooking(bookingID) {
    try {
      await api(`/api/customer/bookings/${bookingID}/cancel`, { method: 'POST', body: JSON.stringify({ reason: '顾客主动取消' }) })
      setBookings((old) => old.map((booking) => booking.id === bookingID ? { ...booking, status: 'cancelled' } : booking))
      setBookingMessage('预约已取消，房间和次数已释放。')
    } catch (err) { setBookingError(err.message) }
  }

  return <div className="booking-center"><div className="booking-heading"><span>预约排练房</span><small>选择日期 · 房间 · 时段</small></div>{availableRooms.length === 0 ? <div className="membership-empty">当前没有可预约的房间。</div> : <form onSubmit={submitBooking}><div className="booking-fields"><label className="field"><span>预约日期</span><input type="date" value={date} min={today()} onChange={(event) => setDate(event.target.value)} /></label><label className="field"><span>排练房</span><select value={roomID} onChange={(event) => setRoomID(event.target.value)}>{availableRooms.map((room) => <option value={room.id} key={room.id}>{room.name}</option>)}</select></label></div><div className="slot-picker"><span className="field-label">可用时段（含 30 分钟清场）</span><div className="slot-grid">{slots.map((slot) => <button type="button" className={selectedSlot?.startsAt === slot.startsAt && selectedSlot?.endsAt === slot.endsAt ? 'slot selected' : 'slot'} key={`${slot.startsAt}-${slot.endsAt}`} onClick={() => setSelectedSlot(slot)}>{slot.startsAt.slice(11, 16)}–{slot.endsAt.slice(11, 16)}<small>{slot.durationMinutes / 60} 小时</small></button>)}{!slots.length && <small className="muted">当天没有符合条件的时段</small>}</div></div><div className="booking-fields"><Field label="乐队名称" name="bandName" value={details.bandName} onChange={(event) => setDetails({ ...details, bandName: event.target.value })} required /><Field label="到店手机号" name="phone" value={details.phone} onChange={(event) => setDetails({ ...details, phone: event.target.value })} required /></div><label className="field"><span>备注</span><input value={details.notes} onChange={(event) => setDetails({ ...details, notes: event.target.value })} placeholder="可选" /></label><div className="equipment-picker"><span className="field-label">额外借用设备（不额外收费）</span>{equipment.filter((item) => item.status === 'available' && item.quantity > 0).map((item) => <label key={item.id}><span>{item.name} · 库存 {item.quantity}</span><select value={quantities[item.id] || 0} onChange={(event) => setQuantities({ ...quantities, [item.id]: event.target.value })}>{Array.from({ length: Math.min(item.quantity, 5) + 1 }, (_, index) => <option value={index} key={index}>{index === 0 ? '不借用' : `${index} 件`}</option>)}</select></label>)}</div>{bookingMessage && <div className="notice success">{bookingMessage}</div>}{bookingError && <div className="notice error">{bookingError}</div>}<button className="primary-button" type="submit">立即预约</button></form>}<div className="booking-history"><span className="field-label">我的预约</span>{bookings.slice(0, 3).map((booking) => <div className="history-item" key={booking.id}><b>{booking.startsAt.slice(0, 16).replace('T', ' ')}</b><small>{booking.bandName} · {booking.status === 'confirmed' ? '已确认' : booking.status}</small>{booking.status === 'confirmed' && <button type="button" className="cancel-button" onClick={() => cancelBooking(booking.id)}>取消</button>}</div>)}</div></div>
}

function today() { return new Date().toISOString().slice(0, 10) }
function tomorrow() { const date = new Date(); date.setDate(date.getDate() + 1); return date.toISOString().slice(0, 10) }

export default App
